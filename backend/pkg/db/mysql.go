package db

import (
	"os"
	"time"

	"chatapp-backend/internal/conversation"
	"chatapp-backend/internal/friend"
	"chatapp-backend/internal/message"
	"chatapp-backend/internal/user"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitMySQL() (*gorm.DB, error) {
	dsn := os.Getenv("LIGHTCHAT_MYSQL_DSN")

	if dsn == "" {
		dsn = "root:123456@tcp(127.0.0.1:3307)/lightchat?charset=utf8mb4&parseTime=True&loc=Local"
	}

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := AutoMigrate(database); err != nil {
		return nil, err
	}

	return database, nil
}

func AutoMigrate(database *gorm.DB) error {
	return database.AutoMigrate(
		&user.User{},
		&friend.FriendRequest{},
		&friend.Friend{},
		&conversation.Conversation{},
		&message.Message{},
	)
}
