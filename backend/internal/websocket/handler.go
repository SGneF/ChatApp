package websocket

import (
	"context"
	"log"
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
	online   *OnlineService
	upgrader gorillawebsocket.Upgrader //把普通 HTTP 请求升级为 WebSocket 连接。
}

//WebSocket的连接入口

func NewHandler(hub *Hub, db *gorm.DB, online *OnlineService) *Handler {
	return &Handler{
		hub:    hub,
		db:     db,
		online: online,
		upgrader: gorillawebsocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
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
			"msg":  "token cannot be empty",
		})
		return
	}

	claims, err := pkgjwt.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 0,
			"msg":  "invalid token",
		})
		return
	}

	//升级连接，这一步成功后：HTTP 请求结束，WebSocket 长连接建立，后续不能继续使用C.JSON(...)来回复数据，而是要用conn.WriteMessage(...)
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	//创建client，一个client代表一个websocket连接
	client := &Client{
		UserID:   claims.UserID,
		Username: claims.Username,
		Hub:      h.hub,
		Conn:     conn,
		Send:     make(chan []byte, 256), //消息缓冲队列，业务代码不会直接调用Conn.writeMessage（），而是先把消息放到client.Send里，然后由writepump统一发送，这样可以保证一个连接只有一个写协程
		DB:       h.db,
		Online:   h.online,
	}

	//注册连接，把连接放入hub
	h.hub.Register(client)
	//设置redis在线状态
	if err := h.online.SetOnline(context.Background(), claims.UserID); err != nil {
		log.Println("set websocket online state failed:", err)
	}

	//发送连接成功消息
	client.sendJSON(OutgoingMessage{
		Type: EventConnected,
		Data: gin.H{
			"user_id": claims.UserID,
			"message": "WebSocket connected",
		},
	})

	//启动三个协程
	go client.WritePump()       //向客户端发送消息
	go client.ReadPump()        //读取客户端发来的消息
	go client.sendOfflineSync() //查询并同步离线消息
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
