package websocket

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, online *OnlineService) {
	hub := NewHub()
	handler := NewHandler(hub, db, online)

	r.GET("/ws", handler.ServeWS)
}
