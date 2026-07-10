package friend

type ApplyFriendRequest struct {
	ToUserID uint64 `json:"to_user_id" binding:"required"`
	Remark   string `json:"remark" binding:"omitempty,max=255"`
}

type HandleFriendRequest struct {
	RequestID uint64 `json:"request_id" binding:"required"`
}

type DeleteFriendRequest struct {
	FriendID uint64 `uri:"friend_id" binding:"required"`
}
