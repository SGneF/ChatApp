package user

import "time"

type UserResponse struct {
	ID         uint64    `json:"id"`
	Username   string    `json:"username"`
	Nickname   string    `json:"nickname"`
	Avatar     string    `json:"avatar"`
	Signature  string    `json:"signature"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

func ToUserResponse(u User) UserResponse {
	return UserResponse{
		ID:         u.ID,
		Username:   u.Username,
		Nickname:   u.Nickname,
		Avatar:     u.Avatar,
		Signature:  u.Signature,
		CreateTime: u.CreateTime,
		UpdateTime: u.UpdateTime,
	}
}

type SearchUserResponse struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
}
