package file

import (
	"chatapp-backend/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, minioClient *minio.Client) {
	service := NewService(db, minioClient)
	handler := NewHandler(service)

	fileGroup := r.Group("/file")
	fileGroup.Use(middleware.AuthMiddleware())
	{
		fileGroup.POST("/upload", handler.Upload)
		fileGroup.GET("/url", handler.GetURL)
	}
}
