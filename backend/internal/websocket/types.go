package websocket

import "encoding/json"

const (
	EventConnected   = "connected"
	EventChatMessage = "chat_message"
	EventChatAck     = "chat_ack"
	EventChatError   = "chat_error"
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
