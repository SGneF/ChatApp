package conversation

import (
	"context"
	"errors"

	"chatapp-backend/internal/friend"
	"chatapp-backend/internal/user"

	"gorm.io/gorm"
)

var (
	ErrCannotChatWithSelf   = errors.New("不能和自己创建会话")
	ErrTargetNotFound       = errors.New("目标用户不存在")
	ErrNotFriend            = errors.New("不是好友，不能创建会话")
	ErrConversationNotFound = errors.New("会话不存在")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateOrGetSingle(ctx context.Context, currentUserID uint64, targetID uint64) (*ConversationResponse, error) {
	if currentUserID == targetID {
		return nil, ErrCannotChatWithSelf
	}

	var target user.User
	if err := s.db.WithContext(ctx).Where("id = ?", targetID).First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTargetNotFound
		}
		return nil, err
	}

	var friendCount int64
	if err := s.db.WithContext(ctx).
		Model(&friend.Friend{}).
		Where("user_id = ? AND friend_id = ?", currentUserID, targetID).
		Count(&friendCount).Error; err != nil {
		return nil, err
	}
	if friendCount == 0 {
		return nil, ErrNotFriend
	}

	conv := Conversation{
		UserID:   currentUserID,
		TargetID: targetID,
		Type:     ConversationTypeSingle,
	}

	if err := s.db.WithContext(ctx).FirstOrCreate(&conv, Conversation{
		UserID:   currentUserID,
		TargetID: targetID,
		Type:     ConversationTypeSingle,
	}).Error; err != nil {
		return nil, err
	}

	resp := toConversationResponse(conv, target)
	return &resp, nil
}

func (s *Service) List(ctx context.Context, currentUserID uint64) ([]ConversationResponse, error) {
	var conversations []Conversation
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", currentUserID).
		Order("is_top DESC").
		Order("update_time DESC").
		Find(&conversations).Error; err != nil {
		return nil, err
	}

	result := make([]ConversationResponse, 0, len(conversations))
	for _, conv := range conversations {
		resp, err := s.buildResponse(ctx, conv)
		if err != nil {
			continue
		}
		result = append(result, resp)
	}

	return result, nil
}

func (s *Service) GetByID(ctx context.Context, currentUserID uint64, conversationID uint64) (*ConversationResponse, error) {
	conv, err := s.findOwnedConversation(ctx, currentUserID, conversationID)
	if err != nil {
		return nil, err
	}

	resp, err := s.buildResponse(ctx, conv)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, currentUserID uint64, conversationID uint64) error {
	result := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", conversationID, currentUserID).
		Delete(&Conversation{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrConversationNotFound
	}
	return nil
}

func (s *Service) MarkRead(ctx context.Context, currentUserID uint64, conversationID uint64) error {
	if _, err := s.findOwnedConversation(ctx, currentUserID, conversationID); err != nil {
		return err
	}

	return s.db.WithContext(ctx).
		Model(&Conversation{}).
		Where("id = ? AND user_id = ?", conversationID, currentUserID).
		Update("unread_count", 0).Error
}

func (s *Service) SetTop(ctx context.Context, currentUserID uint64, conversationID uint64, isTop bool) error {
	if _, err := s.findOwnedConversation(ctx, currentUserID, conversationID); err != nil {
		return err
	}

	return s.db.WithContext(ctx).
		Model(&Conversation{}).
		Where("id = ? AND user_id = ?", conversationID, currentUserID).
		Update("is_top", isTop).Error
}

func (s *Service) findOwnedConversation(ctx context.Context, currentUserID uint64, conversationID uint64) (Conversation, error) {
	var conv Conversation
	err := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", conversationID, currentUserID).
		First(&conv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Conversation{}, ErrConversationNotFound
	}
	return conv, err
}

func (s *Service) buildResponse(ctx context.Context, conv Conversation) (ConversationResponse, error) {
	var target user.User
	if conv.Type == ConversationTypeSingle {
		if err := s.db.WithContext(ctx).Where("id = ?", conv.TargetID).First(&target).Error; err != nil {
			return ConversationResponse{}, err
		}
	}
	return toConversationResponse(conv, target), nil
}

func toConversationResponse(conv Conversation, target user.User) ConversationResponse {
	return ConversationResponse{
		ID:            conv.ID,
		UserID:        conv.UserID,
		TargetID:      conv.TargetID,
		Type:          conv.Type,
		TargetUser:    toTargetUserResponse(target),
		LastMessageID: conv.LastMessageID,
		LastMessage:   conv.LastMessage,
		UnreadCount:   conv.UnreadCount,
		IsTop:         conv.IsTop,
		CreateTime:    conv.CreateTime,
		UpdateTime:    conv.UpdateTime,
	}
}

func toTargetUserResponse(u user.User) TargetUserResponse {
	return TargetUserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Signature: u.Signature,
	}
}
