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
	writeWait  = 10 * time.Second //每次写消息最多允许花费 10 秒。
	pongWait   = 60 * time.Second //服务端最多等待客户端 60 秒的 Pong。
	pingPeriod = 50 * time.Second //服务端每 50 秒发送一次 Ping。
	maxMsgSize = 5120
)

type Client struct {
	UserID   uint64
	Username string

	Hub  *Hub
	Conn *gorillawebsocket.Conn
	Send chan []byte

	DB     *gorm.DB
	Online *OnlineService
}

// 读取客户端消息
func (c *Client) ReadPump() {
	defer func() {
		//连接退出处理
		noLocalClients := c.Hub.Unregister(c)

		if noLocalClients {
			_ = c.Online.SetOffline(context.Background(), c.UserID)
		}
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

		case EventMessageRead:
			c.handleMessageRead(incoming.Data)

		case EventMessageRevoke:
			c.handleMessageRevoke(incoming.Data)

		default:
			c.sendError("未知消息类型")
		}
	}
}

// 这个协程负责两件事：1.从 Send 队列取出业务消息并发送；2.定时发送 Ping 心跳
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod) //一个定时触发器（Ticker），，每隔指定时间 pingPeriod 自动往自身的 C 通道发送当前时间，循环持续触发，直到手动停止。
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait)) //设置写超时

			if !ok {
				_ = c.Conn.WriteMessage(gorillawebsocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(gorillawebsocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Online.RefreshOnline(context.Background(), c.UserID)

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

	// 给发送者返回 ack
	c.sendJSON(OutgoingMessage{
		Type: EventChatAck,
		Data: msgResp,
	})

	// 推送给接收者
	c.Hub.SendToUser(msgResp.ReceiverID, OutgoingMessage{
		Type: EventChatMessage,
		Data: msgResp,
	})
}

func (c *Client) handleMessageRead(raw json.RawMessage) {
	var data MessageReadData
	if err := json.Unmarshal(raw, &data); err != nil {
		c.sendError("已读消息格式错误")
		return
	}

	if data.ConversationID == 0 {
		c.sendError("会话ID不能为空")
		return
	}

	msgService := message.NewService(c.DB)

	readResp, err := msgService.MarkConversationRead(context.Background(), c.UserID, data.ConversationID)
	if err != nil {
		c.sendError("标记已读失败")
		return
	}

	// 1. 给当前用户返回 ack
	c.sendJSON(OutgoingMessage{
		Type: EventMessageReadAck,
		Data: readResp,
	})

	// 2. 通知对方：我已读了
	c.Hub.SendToUser(readResp.TargetID, OutgoingMessage{
		Type: EventMessageRead,
		Data: readResp,
	})
}

func (c *Client) handleMessageRevoke(raw json.RawMessage) {
	var data MessageRevokeData
	if err := json.Unmarshal(raw, &data); err != nil {
		c.sendError("撤回消息格式错误")
		return
	}

	if data.MessageID == 0 {
		c.sendError("消息ID不能为空")
		return
	}

	msgService := message.NewService(c.DB)

	revokeResp, err := msgService.Revoke(context.Background(), c.UserID, data.MessageID)
	if err != nil {
		c.sendError("撤回消息失败")
		return
	}

	// 1. 给发送者返回撤回成功 ack
	c.sendJSON(OutgoingMessage{
		Type: EventMessageRevokeAck,
		Data: revokeResp,
	})

	// 2. 通知接收者：这条消息已撤回
	c.Hub.SendToUser(revokeResp.ReceiverID, OutgoingMessage{
		Type: EventMessageRevoke,
		Data: revokeResp,
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

	select { //这里是非阻塞写入。如果 Send 队列已经满了，就执行 default，直接丢弃消息。
	case c.Send <- data:
	default:
		log.Println("消息发送队列已满，丢弃消息", c.UserID)
	}
}
