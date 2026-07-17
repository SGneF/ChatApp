package message

import (
	"context"
	"errors"
	"time"

	"chatapp-backend/internal/conversation"

	"gorm.io/gorm"
)

var (
	ErrConversationNotFound  = errors.New("会话不存在")
	ErrMessageNotFound       = errors.New("消息不存在")
	ErrNoPermission          = errors.New("无权限操作")
	ErrRevokeTimeout         = errors.New("消息超过可撤回时间")
	ErrMessageAlreadyRevoked = errors.New("消息已撤回")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) Send(ctx context.Context, currentUserID uint64, req SendMessageRequest) (*MessageResponse, error) {
	msgType := req.Type
	if msgType == "" {
		msgType = MessageTypeText
	}

	var currentConversation conversation.Conversation

	err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", req.ConversationID, currentUserID).
		First(&currentConversation).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrConversationNotFound
	}

	if err != nil {
		return nil, err
	}

	receiverID := currentConversation.TargetID

	var msg Message

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		msg = Message{
			ConversationID: currentConversation.ID,
			SenderID:       currentUserID,
			ReceiverID:     receiverID,
			Type:           msgType,
			Content:        req.Content,
			Status:         MessageStatusSent,
		}

		if err := tx.Create(&msg).Error; err != nil {
			return err
		}

		now := time.Now()
		lastMessage := buildLastMessagePreview(msg.Type, msg.Content)

		// 更新发送方会话
		if err := tx.Model(&conversation.Conversation{}).
			Where("id = ? AND user_id = ?", currentConversation.ID, currentUserID).
			Updates(map[string]interface{}{
				"last_message_id": msg.ID,
				"last_message":    lastMessage,
				"update_time":     now,
			}).Error; err != nil {
			return err
		}

		// 确保接收方也有一条会话记录
		receiverConversation := conversation.Conversation{
			UserID:   receiverID,
			TargetID: currentUserID,
			Type:     conversation.ConversationTypeSingle,
		}

		if err := tx.
			Where("user_id = ? AND target_id = ? AND type = ?", receiverID, currentUserID, conversation.ConversationTypeSingle).
			FirstOrCreate(&receiverConversation).Error; err != nil {
			return err
		}

		// 更新接收方会话：最后一条消息 + 未读数 + 变更时间
		if err := tx.Model(&conversation.Conversation{}).
			Where("id = ? AND user_id = ?", receiverConversation.ID, receiverID).
			Updates(map[string]interface{}{
				"last_message_id": msg.ID,
				"last_message":    lastMessage,
				"unread_count":    gorm.Expr("unread_count + ?", 1),
				"update_time":     now,
			}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	resp := ToMessageResponse(msg)
	return &resp, nil
}

func (s *Service) History(ctx context.Context, currentUserID uint64, conversationID uint64, page int, pageSize int) (*MessageHistoryResponse, error) {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 20
	}

	if pageSize > 100 {
		pageSize = 100
	}

	var c conversation.Conversation

	err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", conversationID, currentUserID).
		First(&c).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrConversationNotFound
	}

	if err != nil {
		return nil, err
	}

	targetID := c.TargetID

	var total int64

	baseQuery := s.db.WithContext(ctx).
		Model(&Message{}).
		Where(
			"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			currentUserID,
			targetID,
			targetID,
			currentUserID,
		)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	var messages []Message

	offset := (page - 1) * pageSize

	// 先按倒序查最近消息，再反转为正序，方便前端展示
	err = baseQuery.
		Order("create_time DESC").
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	reverseMessages(messages)

	list := make([]MessageResponse, 0, len(messages))
	for _, item := range messages {
		list = append(list, ToMessageResponse(item))
	}

	return &MessageHistoryResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
func (s *Service) Revoke(ctx context.Context, currentUserID uint64, messageID uint64) (*MessageRevokeResponse, error) {
	var msg Message

	err := s.db.WithContext(ctx).
		Where("id = ?", messageID).
		First(&msg).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMessageNotFound
	}

	if err != nil {
		return nil, err
	}

	if msg.SenderID != currentUserID {
		return nil, ErrNoPermission
	}

	if msg.Status == MessageStatusRevoked {
		return nil, ErrMessageAlreadyRevoked
	}

	if time.Since(msg.CreateTime) > 2*time.Minute {
		return nil, ErrRevokeTimeout
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		if err := tx.Model(&Message{}).
			Where("id = ?", msg.ID).
			Updates(map[string]interface{}{
				"status":      MessageStatusRevoked,
				"content":     "",
				"update_time": now,
			}).Error; err != nil {
			return err
		}

		// 如果这条消息是最后一条消息，更新双方会话的 last_message
		if err := tx.Model(&conversation.Conversation{}).
			Where(
				"last_message_id = ? AND ((user_id = ? AND target_id = ?) OR (user_id = ? AND target_id = ?))",
				msg.ID,
				msg.SenderID,
				msg.ReceiverID,
				msg.ReceiverID,
				msg.SenderID,
			).
			Updates(map[string]interface{}{
				"last_message": "撤回了一条消息",
				"update_time":  now,
			}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &MessageRevokeResponse{
		MessageID:  msg.ID,
		SenderID:   msg.SenderID,
		ReceiverID: msg.ReceiverID,
		Status:     MessageStatusRevoked,
	}, nil
}

func buildLastMessagePreview(msgType string, content string) string {
	switch msgType {
	case MessageTypeImage:
		return "[图片]"
	case MessageTypeFile:
		return "[文件]"
	case MessageTypeVoice:
		return "[语音]"
	default:
		if len([]rune(content)) > 50 {
			return string([]rune(content)[:50]) + "..."
		}
		return content
	}
}

func reverseMessages(messages []Message) {
	left := 0
	right := len(messages) - 1

	for left < right {
		messages[left], messages[right] = messages[right], messages[left]
		left++
		right--
	}
}

func (s *Service) MarkConversationRead(ctx context.Context, currentUserID uint64, conversationID uint64) (*MessageReadResponse, error) {
	var c conversation.Conversation

	err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", conversationID, currentUserID).
		First(&c).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrConversationNotFound
	}

	if err != nil {
		return nil, err
	}

	targetID := c.TargetID

	var readCount int64

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 清空当前用户这个会话的未读数
		if err := tx.Model(&conversation.Conversation{}).
			Where("id = ? AND user_id = ?", conversationID, currentUserID).
			Update("unread_count", 0).Error; err != nil {
			return err
		}

		// 2. 把“对方发给我”的消息标记为已读
		result := tx.Model(&Message{}).
			Where(
				"sender_id = ? AND receiver_id = ? AND status = ?",
				targetID,
				currentUserID,
				MessageStatusSent,
			).
			Update("status", MessageStatusRead)

		if result.Error != nil {
			return result.Error
		}

		readCount = result.RowsAffected

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &MessageReadResponse{
		ConversationID: conversationID,
		ReaderID:       currentUserID,
		TargetID:       targetID,
		ReadCount:      readCount,
	}, nil
}
