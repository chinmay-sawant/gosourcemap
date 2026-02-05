package router

import (
	"github.com/chinmay-sawant/gosourcemapper/internal/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter(scanHandler *handlers.ScanHandler) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/v1")
	{
		v1.POST("/scan", scanHandler.ScanFile)
		v1.POST("/upload", scanHandler.UploadZip)
		v1.POST("/scan/dir", scanHandler.ScanDirectory)
		v1.GET("/nodes", scanHandler.GetAllNodes)
	}

	return r
}
