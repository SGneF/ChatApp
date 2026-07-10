package friend

import (
	"errors"
	"net/http"

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

func (h *Handler) Apply(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	var req ApplyFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误："+err.Error())
		return
	}

	err := h.service.Apply(c.Request.Context(), currentUserID, req)

	switch {
	case errors.Is(err, ErrCannotAddSelf):
		response.Fail(c, http.StatusBadRequest, "不能添加自己为好友")
	case errors.Is(err, ErrUserNotFound):
		response.Fail(c, http.StatusNotFound, "用户不存在")
	case errors.Is(err, ErrAlreadyFriend):
		response.Fail(c, http.StatusBadRequest, "已经是好友")
	case errors.Is(err, ErrRequestExists):
		response.Fail(c, http.StatusBadRequest, "好友申请已存在")
	case err != nil:
		response.Fail(c, http.StatusInternalServerError, "发送好友申请失败")
	default:
		response.Success(c, gin.H{"message": "好友申请已发送"})
	}
}

func (h *Handler) ListRequests(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	list, err := h.service.ListRequests(c.Request.Context(), currentUserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "获取好友申请失败")
		return
	}

	response.Success(c, list)
}

func (h *Handler) Accept(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	var req HandleFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误："+err.Error())
		return
	}

	err := h.service.Accept(c.Request.Context(), currentUserID, req)

	switch {
	case errors.Is(err, ErrRequestNotFound):
		response.Fail(c, http.StatusNotFound, "好友申请不存在")
	case errors.Is(err, ErrNoPermission):
		response.Fail(c, http.StatusForbidden, "无权处理该好友申请")
	case errors.Is(err, ErrInvalidRequest):
		response.Fail(c, http.StatusBadRequest, "该好友申请已处理")
	case err != nil:
		response.Fail(c, http.StatusInternalServerError, "同意好友申请失败")
	default:
		response.Success(c, gin.H{"message": "已同意好友申请"})
	}
}

func (h *Handler) Reject(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	var req HandleFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误："+err.Error())
		return
	}

	err := h.service.Reject(c.Request.Context(), currentUserID, req)

	switch {
	case errors.Is(err, ErrRequestNotFound):
		response.Fail(c, http.StatusNotFound, "好友申请不存在")
	case errors.Is(err, ErrNoPermission):
		response.Fail(c, http.StatusForbidden, "无权处理该好友申请")
	case errors.Is(err, ErrInvalidRequest):
		response.Fail(c, http.StatusBadRequest, "该好友申请已处理")
	case err != nil:
		response.Fail(c, http.StatusInternalServerError, "拒绝好友申请失败")
	default:
		response.Success(c, gin.H{"message": "已拒绝好友申请"})
	}
}

func (h *Handler) ListFriends(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	list, err := h.service.ListFriends(c.Request.Context(), currentUserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "获取好友列表失败")
		return
	}

	response.Success(c, list)
}

func (h *Handler) DeleteFriend(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	var req DeleteFriendRequest
	if err := c.ShouldBindUri(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误："+err.Error())
		return
	}

	err := h.service.DeleteFriend(c.Request.Context(), currentUserID, req.FriendID)

	switch {
	case errors.Is(err, ErrFriendRelationship):
		response.Fail(c, http.StatusNotFound, "好友关系不存在")
	case err != nil:
		response.Fail(c, http.StatusInternalServerError, "删除好友失败")
	default:
		response.Success(c, gin.H{"message": "好友已删除"})
	}
}
