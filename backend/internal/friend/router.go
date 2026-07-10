package friend

import (
	"chatapp-backend/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	service := NewService(db)
	handler := NewHandler(service)

	friendGroup := r.Group("/friend")
	friendGroup.Use(middleware.AuthMiddleware())
	{
		friendGroup.POST("/apply", handler.Apply)
		friendGroup.GET("/requests", handler.ListRequests)
		friendGroup.POST("/accept", handler.Accept)
		friendGroup.POST("/reject", handler.Reject)
		friendGroup.GET("/list", handler.ListFriends)
		friendGroup.DELETE("/:friend_id", handler.DeleteFriend)

	}
}
