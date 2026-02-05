package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/chinmay-sawant/gosourcemapper/internal/service"
	"github.com/gin-gonic/gin"
)

type ScanHandler struct {
	service service.ScanService
}

func NewScanHandler(service service.ScanService) *ScanHandler {
	return &ScanHandler{service: service}
}

type ScanRequest struct {
	FilePath string `json:"file_path" binding:"required"`
	Content  string `json:"content" binding:"required"` // Base64 encoded content
}

func (h *ScanHandler) ScanFile(c *gin.Context) {
	var req ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 content"})
		return
	}

	nodes, err := h.service.ScanFile(req.FilePath, decoded)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nodes": nodes, "count": len(nodes)})
}

func (h *ScanHandler) UploadZip(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Persist in .temp in root
	destRoot := ".temp"
	nodes, err := h.service.ProcessZipUpload(file, destRoot)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nodes": nodes, "count": len(nodes)})
}

func (h *ScanHandler) ScanDirectory(c *gin.Context) {
	var req struct {
		DirPath string `json:"dir_path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nodes, err := h.service.ScanDirectory(req.DirPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nodes": nodes, "count": len(nodes)})
}

func (h *ScanHandler) GetAllNodes(c *gin.Context) {
	nodes := h.service.GetAllNodes()
	c.JSON(http.StatusOK, gin.H{"nodes": nodes, "count": len(nodes)})
}
