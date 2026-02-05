package router

import (
	"inventory-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

// Setup configures the Gin router with all routes.
func Setup(inventoryHandler *handlers.InventoryHandler) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/v1")
	{
		v1.GET("/inventory", inventoryHandler.GetAll)
		v1.GET("/inventory/:id", inventoryHandler.GetByID)
		v1.PUT("/inventory/:id", inventoryHandler.Update)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "inventory-service"})
	})

	return r
}
