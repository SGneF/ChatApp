package main

import (
	"log"
	"net/http"

	"chatapp-backend/internal/conversation"
	"chatapp-backend/internal/friend"
	"chatapp-backend/internal/message"
	"chatapp-backend/internal/user"
	ws "chatapp-backend/internal/websocket"
	"chatapp-backend/pkg/db"
	pkgredis "chatapp-backend/pkg/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	database, err := db.InitMySQL()
	if err != nil {
		log.Fatal("MySQL init failed: ", err)
	}

	var onlineService *ws.OnlineService
	redisClient, err := pkgredis.InitRedis()
	if err != nil {
		log.Println("Redis init failed, websocket online state will use local memory only:", err)
	} else {
		defer redisClient.Close()
		onlineService = ws.NewOnlineService(redisClient)
		log.Println("Redis connected, websocket online state enabled")
	}

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

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

	ws.RegisterRoutes(r, database, onlineService)

	if err := r.Run(":8080"); err != nil {
		log.Fatal("server start failed: ", err)
	}
}
