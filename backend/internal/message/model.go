package message

import "time"

const (
	MessageTypeText  = "text"
	MessageTypeImage = "image"
	MessageTypeFile  = "file"
	MessageTypeVoice = "voice"

	MessageStatusNormal  = "normal"
	MessageStatusRevoked = "revoked"
)

type Message struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`

	// 发送方自己的会话 ID
	// 注意：你的 conversation 表是“每个用户一条会话记录”，所以历史消息查询不能只按 conversation_id 查
	ConversationID uint64 `gorm:"not null;index" json:"conversation_id"`

	SenderID   uint64 `gorm:"not null;index" json:"sender_id"`
	ReceiverID uint64 `gorm:"not null;index" json:"receiver_id"`

	Type    string `gorm:"type:varchar(20);not null" json:"type"`
	Content string `gorm:"type:text;not null" json:"content"`

	Status string `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`

	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
}

func (Message) TableName() string {
	return "messages"
}
