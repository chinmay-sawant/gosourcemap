package router

import (
	"order-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

// Setup configures the Gin router with all routes.
func Setup(orderHandler *handlers.OrderHandler) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/v1")
	{
		v1.GET("/orders", orderHandler.GetAll)
		v1.GET("/orders/:id", orderHandler.GetByID)
		v1.POST("/orders", orderHandler.Create)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "order-service"})
	})

	return r
}
