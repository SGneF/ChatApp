package user

import (
	"context"
	"errors"
	"strconv"
	"strings"

	pkgjwt "chatapp-backend/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUsernameExists  = errors.New("鐢ㄦ埛鍚嶅凡瀛樺湪")
	ErrInvalidAccount  = errors.New("鐢ㄦ埛鍚嶆垨瀵嗙爜閿欒")
	ErrUserNotFound    = errors.New("用户不存在")
	ErrInvalidPassword = errors.New("瀵嗙爜閿欒")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*UserResponse, error) {
	var count int64

	err := s.db.WithContext(ctx).
		Model(&User{}).
		Where("username = ?", req.Username).
		Count(&count).Error

	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, ErrUsernameExists
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}

	u := User{
		Username:  req.Username,
		Password:  string(hashPassword),
		Nickname:  nickname,
		Avatar:    "",
		Signature: "",
	}

	if err := s.db.WithContext(ctx).Create(&u).Error; err != nil {
		return nil, err
	}

	resp := ToUserResponse(u)
	return &resp, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	var u User

	err := s.db.WithContext(ctx).
		Where("username = ?", req.Username).
		First(&u).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvalidAccount
	}

	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password))
	if err != nil {
		return nil, ErrInvalidAccount
	}

	token, err := pkgjwt.GenerateToken(u.ID, u.Username)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: token,
		User:  ToUserResponse(u),
	}, nil
}

func (s *Service) GetUserInfo(ctx context.Context, userID uint64) (*UserResponse, error) {
	var u User

	err := s.db.WithContext(ctx).
		Where("id = ?", userID).
		First(&u).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	resp := ToUserResponse(u)
	return &resp, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uint64, req UpdateProfileRequest) (*UserResponse, error) {
	var u User

	err := s.db.WithContext(ctx).
		Where("id = ?", userID).
		First(&u).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	if req.Nickname != "" {
		u.Nickname = req.Nickname
	}

	if req.Avatar != "" {
		u.Avatar = req.Avatar
	}

	if req.Signature != "" {
		u.Signature = req.Signature
	}

	if err := s.db.WithContext(ctx).Save(&u).Error; err != nil {
		return nil, err
	}

	resp := ToUserResponse(u)
	return &resp, nil
}

func (s *Service) SearchUsers(ctx context.Context, currentUserID uint64, keyword string) ([]SearchUserResponse, error) {
	var users []User

	keyword = strings.TrimSpace(keyword)
	query := s.db.WithContext(ctx).
		Where("id <> ?", currentUserID)

	if keyword != "" {
		like := "%" + keyword + "%"
		if id, err := strconv.ParseUint(keyword, 10, 64); err == nil {
			query = query.Where("username LIKE ? OR nickname LIKE ? OR id = ?", like, like, id)
		} else {
			query = query.Where("username LIKE ? OR nickname LIKE ?", like, like)
		}
	}

	err := query.Limit(20).Find(&users).Error
	if err != nil {
		return nil, err
	}

	result := make([]SearchUserResponse, 0, len(users))
	for _, u := range users {
		result = append(result, SearchUserResponse{
			ID:        u.ID,
			Username:  u.Username,
			Nickname:  u.Nickname,
			Avatar:    u.Avatar,
			Signature: u.Signature,
		})
	}

	return result, nil
}
