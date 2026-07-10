package conversation

import "time"

const (
	ConversationTypeSingle = "single"
	ConversationTypeGroup  = "group"
)

type Conversation struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`

	// 当前用户 ID
	UserID uint64 `gorm:"not null;index;uniqueIndex:idx_user_target_type" json:"user_id"`

	// 单聊时是好友 ID，群聊时可以是 group_id
	TargetID uint64 `gorm:"not null;index;uniqueIndex:idx_user_target_type" json:"target_id"`

	// single / group
	Type string `gorm:"type:varchar(20);not null;uniqueIndex:idx_user_target_type" json:"type"`

	LastMessageID uint64 `gorm:"default:0" json:"last_message_id"`

	LastMessage string `gorm:"type:varchar(500)" json:"last_message"`

	UnreadCount int `gorm:"default:0" json:"unread_count"`

	IsTop bool `gorm:"default:false" json:"is_top"`

	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
}

func (Conversation) TableName() string {
	return "conversations"
}
