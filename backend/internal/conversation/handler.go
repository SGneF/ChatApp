package conversation

import (
	"errors"
	"net/http"
	"strconv"

	"chatapp-backend/pkg/middleware"
	"chatapp-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateOrGetSingle(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	var req CreateSingleConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	conversationResp, err := h.service.CreateOrGetSingle(c.Request.Context(), currentUserID, req.TargetID)
	switch {
	case errors.Is(err, ErrCannotChatWithSelf):
		response.Fail(c, http.StatusBadRequest, "不能和自己创建会话")
	case errors.Is(err, ErrTargetNotFound):
		response.Fail(c, http.StatusNotFound, "目标用户不存在")
	case errors.Is(err, ErrNotFriend):
		response.Fail(c, http.StatusForbidden, "不是好友，不能创建会话")
	case err != nil:
		response.Fail(c, http.StatusInternalServerError, "创建会话失败")
	default:
		response.Success(c, conversationResp)
	}
}

func (h *Handler) List(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	list, err := h.service.List(c.Request.Context(), currentUserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "获取会话列表失败")
		return
	}

	response.Success(c, list)
}

func (h *Handler) Detail(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	conversationID, err := strconv.ParseUint(c.Param("conversation_id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "会话 ID 错误")
		return
	}

	conversationResp, err := h.service.GetByID(c.Request.Context(), currentUserID, conversationID)
	switch {
	case errors.Is(err, ErrConversationNotFound):
		response.Fail(c, http.StatusNotFound, "会话不存在")
	case err != nil:
		response.Fail(c, http.StatusInternalServerError, "获取会话详情失败")
	default:
		response.Success(c, conversationResp)
	}
}

func (h *Handler) Delete(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	conversationID, err := strconv.ParseUint(c.Param("conversation_id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "会话 ID 错误")
		return
	}

	err = h.service.Delete(c.Request.Context(), currentUserID, conversationID)
	switch {
	case errors.Is(err, ErrConversationNotFound):
		response.Fail(c, http.StatusNotFound, "会话不存在")
	case err != nil:
		response.Fail(c, http.StatusInternalServerError, "删除会话失败")
	default:
		response.Success(c, gin.H{"message": "会话已删除"})
	}
}

func (h *Handler) MarkRead(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	conversationID, err := strconv.ParseUint(c.Param("conversation_id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "会话 ID 错误")
		return
	}

	err = h.service.MarkRead(c.Request.Context(), currentUserID, conversationID)
	switch {
	case errors.Is(err, ErrConversationNotFound):
		response.Fail(c, http.StatusNotFound, "会话不存在")
	case err != nil:
		response.Fail(c, http.StatusInternalServerError, "设置已读失败")
	default:
		response.Success(c, gin.H{"message": "已读成功"})
	}
}

func (h *Handler) SetTop(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	conversationID, err := strconv.ParseUint(c.Param("conversation_id"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "会话 ID 错误")
		return
	}

	var req SetTopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	err = h.service.SetTop(c.Request.Context(), currentUserID, conversationID, req.IsTop)
	switch {
	case errors.Is(err, ErrConversationNotFound):
		response.Fail(c, http.StatusNotFound, "会话不存在")
	case err != nil:
		response.Fail(c, http.StatusInternalServerError, "设置置顶失败")
	default:
		response.Success(c, gin.H{"message": "设置成功"})
	}
}
