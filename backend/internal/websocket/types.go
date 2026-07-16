package websocket

import (
	"encoding/json"
	"time"

	"chatapp-backend/internal/message"
)

const (
	EventConnected   = "connected"
	EventChatMessage = "chat_message"
	EventChatAck     = "chat_ack"
	EventChatError   = "chat_error"
	EventOfflineSync = "offline_sync"
)

type IncomingMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type ChatMessageData struct {
	ConversationID uint64 `json:"conversation_id"`
	Type           string `json:"type"`
	Content        string `json:"content"`
}

type OutgoingMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type ErrorMessage struct {
	Message string `json:"message"`
}

type OfflineSyncData struct {
	HasUnread     bool                      `json:"has_unread"`
	UnreadTotal   int                       `json:"unread_total"`
	Conversations []OfflineSyncConversation `json:"conversations"`
}

type OfflineSyncConversation struct {
	ConversationID uint64                    `json:"conversation_id"`
	TargetID       uint64                    `json:"target_id"`
	UnreadCount    int                       `json:"unread_count"`
	LastMessageID  uint64                    `json:"last_message_id"`
	LastMessage    string                    `json:"last_message"`
	UpdateTime     time.Time                 `json:"update_time"`
	Messages       []message.MessageResponse `json:"messages"`
}
