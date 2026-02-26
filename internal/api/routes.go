package api

import (
	"kjt_go_dana/internal/sse"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up all API routes using Official SDK implementation
func SetupRoutes(r *gin.Engine, h *SDKAPIHandler, sseBroker *sse.Broker) {
	// Root demo page (optional, but good for testing)
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "demo.html", nil)
	})

	// API endpoints - all using Official SDK
	api := r.Group("/api/v1")
	{
		// Health check
		api.GET("/health", h.HealthCheck)

		// Shop Management endpoints (Official SDK Implementation)
		shop := api.Group("/shop")
		{
			shop.POST("/create", h.CreateShop)
			shop.GET("/query", h.QueryShop)
			shop.POST("/update", h.UpdateShop)
		}

		// Division Management endpoints
		division := api.Group("/division")
		{
			division.POST("/create", h.CreateDivision)
			division.GET("/query", h.QueryDivision)
			division.POST("/update", h.UpdateDivision)
		}

		// Disbursement endpoints
		disbursement := api.Group("/disbursement")
		{
			disbursement.POST("/transfer-to-dana", h.TransferToDana)
			disbursement.POST("/transfer-to-dana/status", h.TransferToDanaInquiryStatus)
		}
	}

	// Webhook endpoint
	r.POST("/webhook/dana", func(c *gin.Context) {
		// Webhook handler - placeholder for future implementation
		c.JSON(200, gin.H{"success": true, "message": "webhook received"})
	})

	// SSE endpoint (for real-time updates)
	r.GET("/sse/payment", func(c *gin.Context) {
		sseBroker.ServeHTTP(c.Writer, c.Request)
	})
}
