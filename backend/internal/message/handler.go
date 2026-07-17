package message

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
	return &Handler{
		service: service,
	}
}

func (h *Handler) Send(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误："+err.Error())
		return
	}

	msgResp, err := h.service.Send(c.Request.Context(), currentUserID, req)

	switch {
	case errors.Is(err, ErrConversationNotFound):
		response.Fail(c, http.StatusNotFound, "会话不存在")
	case err != nil:
		response.Fail(c, http.StatusInternalServerError, "发送消息失败")
	default:
		response.Success(c, msgResp)
	}
}

func (h *Handler) History(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	conversationID, err := strconv.ParseUint(c.Query("conversation_id"), 10, 64)
	if err != nil || conversationID == 0 {
		response.Fail(c, http.StatusBadRequest, "会话ID错误")
		return
	}

	page := parseIntWithDefault(c.Query("page"), 1)
	pageSize := parseIntWithDefault(c.Query("page_size"), 20)

	historyResp, err := h.service.History(c.Request.Context(), currentUserID, conversationID, page, pageSize)

	switch {
	case errors.Is(err, ErrConversationNotFound):
		response.Fail(c, http.StatusNotFound, "会话不存在")
	case err != nil:
		response.Fail(c, http.StatusInternalServerError, "获取历史消息失败")
	default:
		response.Success(c, historyResp)
	}
}

func (h *Handler) Revoke(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	messageID, err := strconv.ParseUint(c.Param("message_id"), 10, 64)
	if err != nil || messageID == 0 {
		response.Fail(c, http.StatusBadRequest, "消息ID错误")
		return
	}

	revokeResp, err := h.service.Revoke(c.Request.Context(), currentUserID, messageID)

	switch {
	case errors.Is(err, ErrMessageNotFound):
		response.Fail(c, http.StatusNotFound, "消息不存在")
	case errors.Is(err, ErrNoPermission):
		response.Fail(c, http.StatusForbidden, "无权限撤回该消息")
	case errors.Is(err, ErrRevokeTimeout):
		response.Fail(c, http.StatusBadRequest, "消息已超过可撤回时间")
	case errors.Is(err, ErrMessageAlreadyRevoked):
		response.Fail(c, http.StatusBadRequest, "消息已撤回")
	case err != nil:
		response.Fail(c, http.StatusInternalServerError, "撤回消息失败")
	default:
		response.Success(c, revokeResp)
	}
}

func parseIntWithDefault(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	num, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	if num <= 0 {
		return defaultValue
	}

	return num
}
