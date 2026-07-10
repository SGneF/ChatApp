package user

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
	return &Handler{
		service: service,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	userResp, err := h.service.Register(c.Request.Context(), req)
	if errors.Is(err, ErrUsernameExists) {
		response.Fail(c, http.StatusBadRequest, "用户名已存在")
		return
	}

	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "注册失败")
		return
	}

	response.Success(c, userResp)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	loginResp, err := h.service.Login(c.Request.Context(), req)
	if errors.Is(err, ErrInvalidAccount) {
		response.Fail(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "登录失败")
		return
	}

	response.Success(c, loginResp)
}

func (h *Handler) Info(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	userResp, err := h.service.GetUserInfo(c.Request.Context(), userID)
	if errors.Is(err, ErrUserNotFound) {
		response.Fail(c, http.StatusNotFound, "用户不存在")
		return
	}

	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	response.Success(c, userResp)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	userResp, err := h.service.UpdateProfile(c.Request.Context(), userID, req)
	if errors.Is(err, ErrUserNotFound) {
		response.Fail(c, http.StatusNotFound, "用户不存在")
		return
	}

	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "更新用户信息失败")
		return
	}

	response.Success(c, userResp)
}

func (h *Handler) Search(c *gin.Context) {
	currentUserID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	keyword := c.Query("keyword")

	list, err := h.service.SearchUsers(c.Request.Context(), currentUserID, keyword)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "搜索用户失败")
		return
	}

	response.Success(c, list)
}
