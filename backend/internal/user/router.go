package user

import (
	"chatapp-backend/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	service := NewService(db)
	handler := NewHandler(service)

	userGroup := r.Group("/user")
	{
		userGroup.POST("/register", handler.Register)
		userGroup.POST("/login", handler.Login)

		authGroup := userGroup.Group("")
		authGroup.Use(middleware.AuthMiddleware())
		{
			authGroup.GET("/info", handler.Info)
			authGroup.POST("/profile", handler.UpdateProfile)
			authGroup.GET("/search", handler.Search)
		}
	}
}
