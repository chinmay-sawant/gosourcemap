package handlers

import (
	"net/http"

	"inventory-service/internal/models"
	"inventory-service/internal/service"

	"github.com/gin-gonic/gin"
)

// InventoryHandler handles HTTP requests for inventory.
type InventoryHandler struct {
	svc service.InventoryService
}

// NewInventoryHandler creates a new InventoryHandler.
func NewInventoryHandler(svc service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

// GetAll returns all inventory items.
func (h *InventoryHandler) GetAll(c *gin.Context) {
	items := h.svc.GetAll()
	c.JSON(http.StatusOK, gin.H{"inventory": items})
}

// GetByID returns a single inventory item by ID.
func (h *InventoryHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, _ := h.svc.GetByID(id)
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Update updates an inventory item.
func (h *InventoryHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, _ := h.svc.Update(id, req)
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}
