package message

import "time"

type MessageResponse struct {
	ID             uint64    `json:"id"`
	ConversationID uint64    `json:"conversation_id"`
	SenderID       uint64    `json:"sender_id"`
	ReceiverID     uint64    `json:"receiver_id"`
	Type           string    `json:"type"`
	Content        string    `json:"content"`
	Status         string    `json:"status"`
	CreateTime     time.Time `json:"create_time"`
	UpdateTime     time.Time `json:"update_time"`
}

type MessageHistoryResponse struct {
	List     []MessageResponse `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

func ToMessageResponse(m Message) MessageResponse {
	return MessageResponse{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		ReceiverID:     m.ReceiverID,
		Type:           m.Type,
		Content:        m.Content,
		Status:         m.Status,
		CreateTime:     m.CreateTime,
		UpdateTime:     m.UpdateTime,
	}
}
