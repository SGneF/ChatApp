package friend

import (
	"context"
	"errors"

	"chatapp-backend/internal/user"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound       = errors.New("用户不存在")
	ErrCannotAddSelf      = errors.New("不能添加自己为好友")
	ErrAlreadyFriend      = errors.New("已经是好友")
	ErrRequestExists      = errors.New("好友申请已存在")
	ErrRequestNotFound    = errors.New("好友申请不存在")
	ErrNoPermission       = errors.New("无权处理该好友申请")
	ErrInvalidRequest     = errors.New("无效的好友申请")
	ErrFriendRelationship = errors.New("好友关系不存在")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Apply(ctx context.Context, fromUserID uint64, req ApplyFriendRequest) error {
	if fromUserID == req.ToUserID {
		return ErrCannotAddSelf
	}

	var toUser user.User
	if err := s.db.WithContext(ctx).Where("id = ?", req.ToUserID).First(&toUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	isFriend, err := s.isFriend(ctx, fromUserID, req.ToUserID)
	if err != nil {
		return err
	}
	if isFriend {
		return ErrAlreadyFriend
	}

	var count int64
	err = s.db.WithContext(ctx).
		Model(&FriendRequest{}).
		Where("from_user_id = ? AND to_user_id = ? AND status = ?", fromUserID, req.ToUserID, RequestStatusPending).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {

		return ErrRequestExists
	}

	apply := FriendRequest{
		FromUserID: fromUserID,
		ToUserID:   req.ToUserID,
		Remark:     req.Remark,
		Status:     RequestStatusPending,
	}

	return s.db.WithContext(ctx).Create(&apply).Error
}

func (s *Service) ListRequests(ctx context.Context, currentUserID uint64) ([]FriendRequestResponse, error) {
	var requests []FriendRequest

	err := s.db.WithContext(ctx).
		Where("to_user_id = ? AND status = ?", currentUserID, RequestStatusPending).
		Order("create_time DESC").
		Find(&requests).Error
	if err != nil {
		return nil, err
	}

	result := make([]FriendRequestResponse, 0, len(requests))

	for _, r := range requests {
		var fromUser user.User
		if err := s.db.WithContext(ctx).Where("id = ?", r.FromUserID).First(&fromUser).Error; err != nil {
			continue
		}

		result = append(result, FriendRequestResponse{
			ID:           r.ID,
			FromUserID:   r.FromUserID,
			FromUsername: fromUser.Username,
			FromNickname: fromUser.Nickname,
			FromAvatar:   fromUser.Avatar,
			Remark:       r.Remark,
			Status:       r.Status,
			CreateTime:   r.CreateTime,
		})
	}

	return result, nil
}

func (s *Service) Accept(ctx context.Context, currentUserID uint64, req HandleFriendRequest) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var friendReq FriendRequest

		err := tx.Where("id = ?", req.RequestID).First(&friendReq).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRequestNotFound
		}
		if err != nil {
			return err
		}

		if friendReq.ToUserID != currentUserID {
			return ErrNoPermission
		}

		if friendReq.Status != RequestStatusPending {
			return ErrInvalidRequest
		}

		friendReq.Status = RequestStatusAccepted
		if err := tx.Save(&friendReq).Error; err != nil {
			return err
		}

		f1 := Friend{
			UserID:   friendReq.FromUserID,
			FriendID: friendReq.ToUserID,
		}

		f2 := Friend{
			UserID:   friendReq.ToUserID,
			FriendID: friendReq.FromUserID,
		}

		if err := tx.FirstOrCreate(&f1, Friend{
			UserID:   friendReq.FromUserID,
			FriendID: friendReq.ToUserID,
		}).Error; err != nil {
			return err
		}

		if err := tx.FirstOrCreate(&f2, Friend{
			UserID:   friendReq.ToUserID,
			FriendID: friendReq.FromUserID,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) Reject(ctx context.Context, currentUserID uint64, req HandleFriendRequest) error {
	var friendReq FriendRequest

	err := s.db.WithContext(ctx).Where("id = ?", req.RequestID).First(&friendReq).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRequestNotFound
	}
	if err != nil {
		return err
	}

	if friendReq.ToUserID != currentUserID {
		return ErrNoPermission
	}

	if friendReq.Status != RequestStatusPending {
		return ErrInvalidRequest
	}

	friendReq.Status = RequestStatusRejected
	return s.db.WithContext(ctx).Save(&friendReq).Error
}

func (s *Service) ListFriends(ctx context.Context, currentUserID uint64) ([]FriendResponse, error) {
	var friends []Friend

	err := s.db.WithContext(ctx).
		Where("user_id = ?", currentUserID).
		Order("create_time DESC").
		Find(&friends).Error
	if err != nil {
		return nil, err
	}

	result := make([]FriendResponse, 0, len(friends))

	for _, f := range friends {
		var u user.User
		if err := s.db.WithContext(ctx).Where("id = ?", f.FriendID).First(&u).Error; err != nil {
			continue
		}

		result = append(result, FriendResponse{
			ID:        u.ID,
			Username:  u.Username,
			Nickname:  u.Nickname,
			Avatar:    u.Avatar,
			Signature: u.Signature,
			Remark:    f.Remark,
		})
	}

	return result, nil
}

func (s *Service) DeleteFriend(ctx context.Context, currentUserID uint64, friendID uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result1 := tx.Where("user_id = ? AND friend_id = ?", currentUserID, friendID).Delete(&Friend{})
		if result1.Error != nil {
			return result1.Error
		}

		result2 := tx.Where("user_id = ? AND friend_id = ?", friendID, currentUserID).Delete(&Friend{})
		if result2.Error != nil {
			return result2.Error
		}

		if result1.RowsAffected == 0 && result2.RowsAffected == 0 {
			return ErrFriendRelationship
		}

		return nil
	})
}

func (s *Service) isFriend(ctx context.Context, userID uint64, friendID uint64) (bool, error) {
	var count int64

	err := s.db.WithContext(ctx).
		Model(&Friend{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Count(&count).Error

	return count > 0, err
}
