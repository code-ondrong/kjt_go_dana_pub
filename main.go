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

	// Initialize DANA client
	danaClient, err := dana.NewClient(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize DANA client: %v", err)
	}

	// Initialize SSE Broker
	sseBroker := sse.NewBroker()

	// Initialize API Handler
	handler := api.NewAPIHandler(danaClient, sseBroker)

	r := gin.Default()

	// Fix: You trusted all proxies warning
	r.SetTrustedProxies(nil)

	// Load templates
	r.LoadHTMLGlob("templates/*")

	// Setup Routes
	api.SetupRoutes(r, handler, sseBroker)

	// Run server
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	log.Printf("🚀 DANA QRIS Server started on http://localhost:%d", cfg.ServerPort)

	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
