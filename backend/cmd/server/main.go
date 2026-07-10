package main

import (
	"log"
	"net/http"

	"chatapp-backend/internal/conversation"
	"chatapp-backend/internal/friend"
	"chatapp-backend/internal/message"
	"chatapp-backend/internal/user"
	"chatapp-backend/pkg/db"

	"github.com/gin-gonic/gin"
)

func main() {
	database, err := db.InitMySQL()
	if err != nil {
		log.Fatal("MySQL 初始化失败：", err)
	}

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "pong",
		})
	})

	api := r.Group("/api")
	{
		user.RegisterRoutes(api, database)
		friend.RegisterRoutes(api, database)
		conversation.RegisterRoutes(api, database)
		message.RegisterRoutes(api, database)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务启动失败：", err)
	}
}
