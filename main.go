package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"kjt_go_dana/internal/api"
	"kjt_go_dana/internal/config"
	"kjt_go_dana/internal/dana"
	"kjt_go_dana/internal/sse"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Configuration
	cfg := config.DefaultConfig()

	// Validate Payment Gateway configuration (fail-fast)
	if err := cfg.ValidatePaymentGateway(); err != nil {
		log.Fatalf("❌ Payment Gateway configuration error: %v", err)
	}
	log.Printf("✅ Payment Gateway configuration validated")

	// Initialize Official DANA SDK client
	sdkClient, err := dana.NewSDKClient(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize DANA SDK client: %v", err)
	}

	// Initialize SSE Broker
	sseBroker := sse.NewBroker()

	// Initialize SDK API Handler
	sdkHandler := api.NewSDKAPIHandler(sdkClient, sseBroker)

	r := gin.Default()

	// Fix: You trusted all proxies warning
	r.SetTrustedProxies(nil)

	// Load templates
	r.LoadHTMLGlob("templates/*")

	// Setup Routes - only SDK API
	api.SetupRoutes(r, sdkHandler, sseBroker)

	// Run server
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("🚀 DANA Payment Gateway Server started on http://localhost:%d", cfg.ServerPort)
	log.Printf("💳 Payment Gateway API: /api/v1/payment/*")
	log.Printf("🔔 Webhook: POST /webhook/dana")
	log.Printf("📦 Shop API: /api/v1/shop/*")
	log.Printf("🏥 Health Check: /api/v1/health")
	log.Printf("📡 SSE Events: GET /sse/payment")

	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
