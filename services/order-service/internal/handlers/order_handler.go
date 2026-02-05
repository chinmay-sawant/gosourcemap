package handlers

import (
	"net/http"

	"order-service/internal/models"
	"order-service/internal/service"

	"github.com/gin-gonic/gin"
)

// OrderHandler handles HTTP requests for orders.
type OrderHandler struct {
	svc service.OrderService
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// GetAll returns all orders.
func (h *OrderHandler) GetAll(c *gin.Context) {
	userID := c.Query("user_id")
	if userID != "" {
		orders := h.svc.GetByUserID(userID)
		c.JSON(http.StatusOK, gin.H{"orders": orders})
		return
	}
	orders := h.svc.GetAll()
	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// GetByID returns a single order by ID.
func (h *OrderHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	order, _ := h.svc.GetByID(id)
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// Create creates a new order.
func (h *OrderHandler) Create(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order := h.svc.Create(req)
	c.JSON(http.StatusCreated, order)
}
