package router

import (
	"user-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

// Setup configures the Gin router with all routes.
func Setup(userHandler *handlers.UserHandler) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/v1")
	{
		v1.GET("/users", userHandler.GetAll)
		v1.GET("/users/:id", userHandler.GetByID)
		v1.POST("/users", userHandler.Create)
		v1.GET("/users/:id/dashboard", userHandler.GetDashboard)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "user-service"})
	})

	return r
}
