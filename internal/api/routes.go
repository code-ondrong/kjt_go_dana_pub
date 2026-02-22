package api

import (
	"kjt_go_dana/internal/sse"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, h *APIHandler, sseBroker *sse.Broker) {
	// Root demo page (optional, but good for testing)
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "demo.html", nil)
	})

	// API endpoints
	api := r.Group("/api")
	{
		// QRIS endpoints
		api.POST("/qris/create", h.CreateQR)
		api.GET("/qris/status/:partnerReferenceNo", h.QueryQR)
		api.GET("/health", h.HealthCheck)

		// Shop Management endpoints
		shop := api.Group("/shop")
		{
			shop.POST("/create", h.CreateShop)
			shop.POST("/update", h.UpdateShop)
			shop.POST("/query", h.QueryShop)
		}
	}

	// Webhook endpoint
	r.POST("/webhook/dana", h.HandleWebhook)

	// SSE endpoint
	r.GET("/sse/payment", func(c *gin.Context) {
		sseBroker.ServeHTTP(c.Writer, c.Request)
	})
}
