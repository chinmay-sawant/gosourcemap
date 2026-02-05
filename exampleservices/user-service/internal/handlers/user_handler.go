package handlers

import (
	"net/http"

	"user-service/internal/models"
	"user-service/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for users.
type UserHandler struct {
	svc service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GetAll returns all users.
func (h *UserHandler) GetAll(c *gin.Context) {
	users := h.svc.GetAll()
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetByID returns a single user by ID.
func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	user, _ := h.svc.GetByID(id)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Create creates a new user.
func (h *UserHandler) Create(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := h.svc.Create(req)
	c.JSON(http.StatusCreated, user)
}

// GetDashboard returns aggregated data from Order and Inventory services.
func (h *UserHandler) GetDashboard(c *gin.Context) {
	id := c.Param("id")
	dashboard, err := h.svc.GetDashboard(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if dashboard == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, dashboard)
}
