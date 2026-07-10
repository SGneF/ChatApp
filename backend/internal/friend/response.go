package friend

import "time"

type FriendRequestResponse struct {
	ID           uint64    `json:"id"`
	FromUserID   uint64    `json:"from_user_id"`
	FromUsername string    `json:"from_username"`
	FromNickname string    `json:"from_nickname"`
	FromAvatar   string    `json:"from_avatar"`
	Remark       string    `json:"remark"`
	Status       string    `json:"status"`
	CreateTime   time.Time `json:"create_time"`
}

type FriendResponse struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
	Remark    string `json:"remark"`
}
