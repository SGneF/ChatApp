package conversation

type CreateSingleConversationRequest struct {
	TargetID uint64 `json:"target_id" binding:"required"`
}

type SetTopRequest struct {
	IsTop bool `json:"is_top"`
}
