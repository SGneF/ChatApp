package conversation

import (
	"chatapp-backend/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	service := NewService(db)
	handler := NewHandler(service)

	conversationGroup := r.Group("/conversation")
	conversationGroup.Use(middleware.AuthMiddleware())
	{
		conversationGroup.POST("/single", handler.CreateOrGetSingle)
		conversationGroup.GET("/list", handler.List)
		conversationGroup.GET("/:conversation_id", handler.Detail)
		conversationGroup.DELETE("/:conversation_id", handler.Delete)
		conversationGroup.POST("/:conversation_id/read", handler.MarkRead)
		conversationGroup.POST("/:conversation_id/top", handler.SetTop)
	}
}
