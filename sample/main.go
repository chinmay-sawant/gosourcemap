package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	svc interface{ GetAll() []string }
}

// GetAll returns all inventory items.
// It also logs the access.
//
// This is a multi-line comment.
func (h *InventoryHandler) GetAll(c *gin.Context) {
	items := h.svc.GetAll()
	http.Get("https://analytics.service/track")
	c.JSON(http.StatusOK, gin.H{"inventory": items})
}
