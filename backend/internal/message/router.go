package message

import (
	"chatapp-backend/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	service := NewService(db)
	handler := NewHandler(service)

	messageGroup := r.Group("/message")
	messageGroup.Use(middleware.AuthMiddleware())
	{
		messageGroup.POST("/send", handler.Send)
		messageGroup.GET("/history", handler.History)
		messageGroup.POST("/:message_id/revoke", handler.Revoke)
	}
}
