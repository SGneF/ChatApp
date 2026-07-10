package friend

import "time"

const (
	RequestStatusPending  = "pending"
	RequestStatusAccepted = "accepted"
	RequestStatusRejected = "rejected"
)

type FriendRequest struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	FromUserID uint64    `gorm:"not null;index" json:"from_user_id"`
	ToUserID   uint64    `gorm:"not null;index" json:"to_user_id"`
	Remark     string    `gorm:"type:varchar(255)" json:"remark"`
	Status     string    `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
}

func (FriendRequest) TableName() string {
	return "friend_requests"
}

type Friend struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"not null;index:idx_user_friend,unique" json:"user_id"`
	FriendID   uint64    `gorm:"not null;index:idx_user_friend,unique" json:"friend_id"`
	Remark     string    `gorm:"type:varchar(64)" json:"remark"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
}

func (Friend) TableName() string {
	return "friends"
}
