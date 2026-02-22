package dana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"kjt_go_dana/internal/config"
)

// Client adalah DANA API client
type Client struct {
	cfg        *config.DANAConfig
	httpClient *http.Client
	signer     *Signer
}

// NewClient membuat DANA client baru
func NewClient(cfg *config.DANAConfig) (*Client, error) {
	signer, err := NewSigner(cfg.ClientID, cfg.PrivateKey, cfg.PublicKey)
	if err != nil {
		log.Printf("[DANA] Warning: signer init: %v (pastikan key sudah dikonfigurasi untuk production)", err)
		signer = &Signer{clientID: cfg.ClientID}
	}

	return &Client{
		cfg:    cfg,
		signer: signer,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// ============================================================
// QR PAYMENT APIs
// ============================================================

// CreateQR membuat QR Code pembayaran baru
func (c *Client) CreateQR(ctx context.Context, req *CreateQRRequest) (*CreateQRResponse, error) {
	path := "/v1.0/qr/qr-mpm-generate.htm"

	resp, err := c.doRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("create QR request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DANA API error %d: %s", resp.StatusCode, string(body))
	}

	var result CreateQRResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	log.Printf("[DANA] CreateQR success: partnerRef=%s, ref=%s, status=%s",
		result.PartnerReferenceNo, result.ReferenceNo, result.ResponseCode)

	return &result, nil
}

// QueryQR mengecek status pembayaran QR
func (c *Client) QueryQR(ctx context.Context, req *QueryQRRequest) (*QueryQRResponse, error) {
	path := "/v1.0/qr/dynamic-qris/query"

	resp, err := c.doRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("query QR request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DANA API error %d: %s", resp.StatusCode, string(body))
	}

	var result QueryQRResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	log.Printf("[DANA] QueryQR: partnerRef=%s, status=%s",
		result.PartnerReferenceNo, result.TransactionStatus)

	return &result, nil
}

// ============================================================
// SHOP MANAGEMENT APIs
// ============================================================

// CreateShop membuat shop baru
func (c *Client) CreateShop(ctx context.Context, req *CreateShopRequest) (*CreateShopResponse, error) {
	path := "/v1.0/merchant-management/shop/create"

	log.Printf("[DANA] CreateShop request: shopParentType=%s, shopParentId=%s, shopName=%s",
		req.ShopParentType, req.ShopParentId, req.ShopName)

	resp, err := c.doRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("create shop request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[DANA] CreateShop failed: HTTP %d, Response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("DANA API error %d: %s", resp.StatusCode, string(body))
	}

	var result CreateShopResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	log.Printf("[DANA] CreateShop success: shopId=%s, shopName=%s, status=%s",
		result.ShopID, result.ShopName, result.ResponseCode)

	return &result, nil
}

// UpdateShop mengupdate informasi shop yang sudah ada
func (c *Client) UpdateShop(ctx context.Context, req *UpdateShopRequest) (*UpdateShopResponse, error) {
	path := "/v1.0/merchant-management/shop/update"

	resp, err := c.doRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("update shop request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[DANA] UpdateShop failed: HTTP %d, Response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("DANA API error %d: %s", resp.StatusCode, string(body))
	}

	var result UpdateShopResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	log.Printf("[DANA] UpdateShop success: shopId=%s, shopName=%s, status=%s",
		result.ShopID, result.ShopName, result.ResponseCode)

	return &result, nil
}

// QueryShop mendapatkan informasi shop berdasarkan filter
func (c *Client) QueryShop(ctx context.Context, req *QueryShopRequest) (*QueryShopResponse, error) {
	path := "/v1.0/merchant-management/shop/query"

	log.Printf("[DANA] QueryShop request: shopId=%s, shopParentType=%s, shopParentId=%s, pageNo=%d, pageSize=%d",
		req.ShopID, req.ShopParentType, req.ShopParentId, req.PageNo, req.PageSize)

	resp, err := c.doRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("query shop request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[DANA] QueryShop failed: HTTP %d, Response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("DANA API error %d: %s", resp.StatusCode, string(body))
	}

	var result QueryShopResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	log.Printf("[DANA] QueryShop success: found=%d shops, status=%s",
		len(result.ShopDetailInfoList), result.ResponseCode)

	return &result, nil
}

// ============================================================
// HTTP HELPER
// ============================================================

func (c *Client) doRequest(ctx context.Context, method, path string, payload interface{}) (*http.Response, error) {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	url := c.cfg.BaseURL() + path
	log.Printf("[DANA] Request URL: %s", url)

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers standar
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-PARTNER-ID", c.cfg.PartnerID)
	req.Header.Set("X-EXTERNAL-ID", fmt.Sprintf("%d", time.Now().UnixNano()))
	req.Header.Set("CHANNEL-ID", "PC_WEB")

	// Set timestamp
	timestamp := time.Now().Format("2006-01-02T15:04:05+07:00")
	req.Header.Set("X-TIMESTAMP", timestamp)

	// Force Bearer for testing
	log.Printf("[DANA] Testing with Bearer Authentication (ClientSecret)")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cfg.ClientSecret))

	/*
			// Buat signature jika private key tersedia
			if c.signer != nil && c.signer.privateKey != nil {
		...
			}
	*/

	log.Printf("[DANA] %s %s", method, url)

	return c.httpClient.Do(req)
}
