package user

import "time"

type User struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Password  string `gorm:"type:varchar(255);not null" json:"-"`
	Nickname  string `gorm:"type:varchar(64);not null" json:"nickname"`
	Avatar    string `gorm:"type:varchar(255)" json:"avatar"`
	Signature string `gorm:"type:varchar(255)" json:"signature"`

	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
}

func (User) TableName() string {
	return "users"
}
