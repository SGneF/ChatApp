package message

type SendMessageRequest struct {
	ConversationID uint64 `json:"conversation_id" binding:"required"`
	Type           string `json:"type" binding:"omitempty,oneof=text image file voice"`
	Content        string `json:"content" binding:"required,max=5000"`
}
