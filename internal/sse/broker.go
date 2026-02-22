package sse

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Event adalah pesan yang dikirim melalui SSE
type Event struct {
	ID   string      `json:"id"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
	Time time.Time   `json:"time"`
}

// PaymentEvent adalah event khusus untuk update status pembayaran
type PaymentEvent struct {
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	ReferenceNo        string `json:"referenceNo"`
	Status             string `json:"status"`
	Amount             string `json:"amount,omitempty"`
	Currency           string `json:"currency,omitempty"`
	Message            string `json:"message,omitempty"`
	PaidAt             string `json:"paidAt,omitempty"`
}

// Client merepresentasikan koneksi SSE dari satu browser/client
type Client struct {
	id      string
	channel string // Subscribe ke channel tertentu (misal: partnerReferenceNo)
	send    chan Event
	done    chan struct{}
}

// Broker mengelola semua koneksi SSE dan distribusi event
type Broker struct {
	mu      sync.RWMutex
	clients map[string]*Client // key: client ID

	// Channel subscriptions: channelKey -> []clientID
	subscriptions map[string][]string

	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastMsg
}

type broadcastMsg struct {
	channel string
	event   Event
}

// NewBroker membuat SSE broker baru
func NewBroker() *Broker {
	b := &Broker{
		clients:       make(map[string]*Client),
		subscriptions: make(map[string][]string),
		register:      make(chan *Client, 10),
		unregister:    make(chan *Client, 10),
		broadcast:     make(chan broadcastMsg, 100),
	}
	go b.run()
	return b
}

// run adalah event loop utama broker
func (b *Broker) run() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client.id] = client
			b.subscriptions[client.channel] = append(b.subscriptions[client.channel], client.id)
			b.mu.Unlock()
			log.Printf("[SSE] Client %s terhubung ke channel '%s'", client.id, client.channel)

		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client.id]; ok {
				delete(b.clients, client.id)
				// Hapus dari subscriptions
				subs := b.subscriptions[client.channel]
				for i, id := range subs {
					if id == client.id {
						b.subscriptions[client.channel] = append(subs[:i], subs[i+1:]...)
						break
					}
				}
				close(client.send)
			}
			b.mu.Unlock()
			log.Printf("[SSE] Client %s terputus dari channel '%s'", client.id, client.channel)

		case msg := <-b.broadcast:
			b.mu.RLock()
			subscribers := b.subscriptions[msg.channel]
			for _, clientID := range subscribers {
				if client, ok := b.clients[clientID]; ok {
					select {
					case client.send <- msg.event:
					default:
						log.Printf("[SSE] Buffer penuh untuk client %s, skip event", clientID)
					}
				}
			}
			b.mu.RUnlock()
		}
	}
}

// Publish mengirim event ke semua client yang subscribe ke channel tertentu
func (b *Broker) Publish(channel string, eventType string, data interface{}) {
	event := Event{
		ID:   fmt.Sprintf("%d", time.Now().UnixNano()),
		Type: eventType,
		Data: data,
		Time: time.Now(),
	}
	b.broadcast <- broadcastMsg{channel: channel, event: event}
	log.Printf("[SSE] Publish event '%s' ke channel '%s'", eventType, channel)
}

// PublishPaymentUpdate mengirim update status pembayaran
func (b *Broker) PublishPaymentUpdate(partnerReferenceNo string, event PaymentEvent) {
	b.Publish(partnerReferenceNo, "payment_update", event)
}

// ServeHTTP menangani koneksi SSE baru
// Endpoint: GET /sse/payment?channel=<partnerReferenceNo>
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Pastikan browser support SSE
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE tidak didukung", http.StatusBadRequest)
		return
	}

	// Ambil channel dari query params
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		channel = "global"
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("X-Accel-Buffering", "no") // Untuk nginx

	// Buat client baru
	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
	client := &Client{
		id:      clientID,
		channel: channel,
		send:    make(chan Event, 10),
		done:    make(chan struct{}),
	}

	// Register client
	b.register <- client

	// Kirim event pertama sebagai konfirmasi koneksi
	sendSSEEvent(w, flusher, Event{
		ID:   "0",
		Type: "connected",
		Data: map[string]string{
			"clientId": clientID,
			"channel":  channel,
			"message":  "Terhubung ke DANA Payment SSE",
		},
		Time: time.Now(),
	})

	// Ping setiap 30 detik agar koneksi tetap hidup
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Listen untuk events dan disconnect
	defer func() {
		b.unregister <- client
	}()

	for {
		select {
		case event, ok := <-client.send:
			if !ok {
				return
			}
			sendSSEEvent(w, flusher, event)

		case <-ticker.C:
			// Kirim comment sebagai heartbeat
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()

		case <-r.Context().Done():
			log.Printf("[SSE] Client %s disconnect (context done)", clientID)
			return
		}
	}
}

// sendSSEEvent menulis event ke response writer dalam format SSE
func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event Event) {
	dataBytes, err := json.Marshal(event.Data)
	if err != nil {
		log.Printf("[SSE] Error marshal event data: %v", err)
		return
	}

	fmt.Fprintf(w, "id: %s\n", event.ID)
	fmt.Fprintf(w, "event: %s\n", event.Type)
	fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
	flusher.Flush()
}

// ClientCount mengembalikan jumlah client yang terhubung
func (b *Broker) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}
