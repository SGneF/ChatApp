package conversation

import "time"

type TargetUserResponse struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
}

type ConversationResponse struct {
	ID            uint64             `json:"id"`
	UserID        uint64             `json:"user_id"`
	TargetID      uint64             `json:"target_id"`
	Type          string             `json:"type"`
	TargetUser    TargetUserResponse `json:"target_user"`
	LastMessageID uint64             `json:"last_message_id"`
	LastMessage   string             `json:"last_message"`
	UnreadCount   int                `json:"unread_count"`
	IsTop         bool               `json:"is_top"`
	CreateTime    time.Time          `json:"create_time"`
	UpdateTime    time.Time          `json:"update_time"`
}
