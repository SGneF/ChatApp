package websocket

import (
	"net/http"
	"strings"

	pkgjwt "chatapp-backend/pkg/jwt"

	"github.com/gin-gonic/gin"
	gorillawebsocket "github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type Handler struct {
	hub      *Hub
	db       *gorm.DB
	upgrader gorillawebsocket.Upgrader
}

func NewHandler(hub *Hub, db *gorm.DB) *Handler {
	return &Handler{
		hub: hub,
		db:  db,
		upgrader: gorillawebsocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// 开发阶段先允许所有来源
				// 正式环境可以限制 origin
				return true
			},
		},
	}
}

func (h *Handler) ServeWS(c *gin.Context) {
	token := getTokenFromRequest(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 0,
			"msg":  "token 不能为空",
		})
		return
	}

	claims, err := pkgjwt.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 0,
			"msg":  "token 无效",
		})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		UserID:   claims.UserID,
		Username: claims.Username,
		Hub:      h.hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		DB:       h.db,
	}

	h.hub.Register(client)

	client.sendJSON(OutgoingMessage{
		Type: EventConnected,
		Data: gin.H{
			"user_id": claims.UserID,
			"message": "WebSocket 连接成功",
		},
	})

	go client.WritePump()
	go client.ReadPump()
}

func getTokenFromRequest(c *gin.Context) string {
	token := c.Query("token")
	if token != "" {
		return token
	}

	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}
