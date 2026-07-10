package user

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Nickname string `json:"nickname" binding:"omitempty,max=32"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Nickname  string `json:"nickname" binding:"omitempty,max=32"`
	Avatar    string `json:"avatar" binding:"omitempty,max=255"`
	Signature string `json:"signature" binding:"omitempty,max=255"`
}
