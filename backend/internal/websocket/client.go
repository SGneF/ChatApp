package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"chatapp-backend/internal/message"

	gorillawebsocket "github.com/gorilla/websocket"
	"gorm.io/gorm"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 50 * time.Second
	maxMsgSize = 5120
)

type Client struct {
	UserID   uint64
	Username string

	Hub  *Hub
	Conn *gorillawebsocket.Conn
	Send chan []byte

	DB *gorm.DB
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		_ = c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMsgSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, rawMsg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var incoming IncomingMessage
		if err := json.Unmarshal(rawMsg, &incoming); err != nil {
			c.sendError("消息格式错误")
			continue
		}

		switch incoming.Type {
		case EventChatMessage:
			c.handleChatMessage(incoming.Data)
		default:
			c.sendError("未知消息类型")
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				_ = c.Conn.WriteMessage(gorillawebsocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(gorillawebsocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.Conn.WriteMessage(gorillawebsocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleChatMessage(raw json.RawMessage) {
	var data ChatMessageData
	if err := json.Unmarshal(raw, &data); err != nil {
		c.sendError("聊天消息格式错误")
		return
	}

	if data.Content == "" {
		c.sendError("消息内容不能为空")
		return
	}

	msgService := message.NewService(c.DB)

	msgResp, err := msgService.Send(context.Background(), c.UserID, message.SendMessageRequest{
		ConversationID: data.ConversationID,
		Type:           data.Type,
		Content:        data.Content,
	})

	if err != nil {
		log.Println("发送消息失败：", err)
		c.sendError("发送消息失败")
		return
	}

	// 1. 给发送者返回 ack
	c.sendJSON(OutgoingMessage{
		Type: EventChatAck,
		Data: msgResp,
	})

	// 2. 如果接收者在线，推送给接收者
	c.Hub.SendToUser(msgResp.ReceiverID, OutgoingMessage{
		Type: EventChatMessage,
		Data: msgResp,
	})
}

func (c *Client) sendError(msg string) {
	c.sendJSON(OutgoingMessage{
		Type: EventChatError,
		Data: ErrorMessage{
			Message: msg,
		},
	})
}

func (c *Client) sendJSON(msg OutgoingMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case c.Send <- data:
	default:
	}
}
