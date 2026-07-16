package websocket

import (
	"context"
	"log"
	"sort"

	"chatapp-backend/internal/conversation"
	"chatapp-backend/internal/message"

	"gorm.io/gorm"
)

const offlineSyncMessageLimit = 200

//离线消息同步
//用户断线期间，其他人发来的消息已经保存进 MySQL。

// 用户重新建立 WebSocket 后执行：sendOfflineSync()，把未读消息主动推给前端
func (c *Client) sendOfflineSync() {
	data, err := BuildOfflineSyncData(context.Background(), c.DB, c.UserID, offlineSyncMessageLimit)
	if err != nil {
		log.Println("build offline sync payload failed:", err)
		return
	}

	c.sendJSON(OutgoingMessage{
		Type: EventOfflineSync,
		Data: data,
	})
}

func BuildOfflineSyncData(ctx context.Context, db *gorm.DB, userID uint64, maxMessages int) (*OfflineSyncData, error) {
	if maxMessages <= 0 {
		maxMessages = offlineSyncMessageLimit
	}

	var conversations []conversation.Conversation
	//查询未读的消息会话
	if err := db.WithContext(ctx).
		Where("user_id = ? AND unread_count > 0", userID).
		Order("update_time DESC").
		Find(&conversations).Error; err != nil {
		return nil, err
	}

	data := &OfflineSyncData{
		Conversations: make([]OfflineSyncConversation, 0, len(conversations)),
	}
	//全局消息数量限制
	remaining := maxMessages

	for _, item := range conversations {
		data.UnreadTotal += item.UnreadCount

		syncConversation := OfflineSyncConversation{
			ConversationID: item.ID,
			TargetID:       item.TargetID,
			UnreadCount:    item.UnreadCount,
			LastMessageID:  item.LastMessageID,
			LastMessage:    item.LastMessage,
			UpdateTime:     item.UpdateTime,
			Messages:       []message.MessageResponse{},
		}

		if remaining > 0 && item.UnreadCount > 0 {
			limit := item.UnreadCount
			if limit > remaining {
				limit = remaining
			}

			messages, err := loadUnreadMessages(ctx, db, item.TargetID, userID, limit)
			if err != nil {
				return nil, err
			}

			syncConversation.Messages = messages
			remaining -= len(messages)
		}

		data.Conversations = append(data.Conversations, syncConversation)
	}

	data.HasUnread = data.UnreadTotal > 0
	return data, nil
}

func loadUnreadMessages(ctx context.Context, db *gorm.DB, senderID uint64, receiverID uint64, limit int) ([]message.MessageResponse, error) {
	if limit <= 0 {
		return []message.MessageResponse{}, nil
	}

	var records []message.Message
	if err := db.WithContext(ctx).
		Where("sender_id = ? AND receiver_id = ?", senderID, receiverID).
		Order("create_time DESC").
		Order("id DESC").
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, err
	}
	//把消息重新按时间正序排列。
	sort.SliceStable(records, func(i, j int) bool {
		if !records[i].CreateTime.Equal(records[j].CreateTime) {
			return records[i].CreateTime.Before(records[j].CreateTime)
		}
		return records[i].ID < records[j].ID
	})

	result := make([]message.MessageResponse, 0, len(records))
	for _, record := range records {
		result = append(result, message.ToMessageResponse(record))
	}

	return result, nil
}
