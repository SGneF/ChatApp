package websocket

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	hub := NewHub()
	handler := NewHandler(hub, db)

	r.GET("/ws", handler.ServeWS)
}
