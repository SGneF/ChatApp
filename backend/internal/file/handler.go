package file

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

func (h *Handler) Upload(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, "未登录")
		return
	}

	fileType := c.PostForm("type")
	if fileType == "" {
		fileType = FileTypeFile
	}

	header, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "请选择要上传的文件")
		return
	}

	uploadResp, err := h.service.Upload(c.Request.Context(), userID, fileType, header)

	switch {
	case errors.Is(err, ErrInvalidFileType):
		response.Fail(c, http.StatusBadRequest, "不支持的文件类型")
	case errors.Is(err, ErrFileTooLarge):
		response.Fail(c, http.StatusBadRequest, "文件过大")
	case err != nil:
		response.Fail(c, http.StatusInternalServerError, "上传失败")
	default:
		response.Success(c, uploadResp)
	}
}

func (h *Handler) GetURL(c *gin.Context) {
	objectName := c.Query("object_name")
	if objectName == "" {
		response.Fail(c, http.StatusBadRequest, "object_name 不能为空")
		return
	}

	fileURL, err := h.service.GenerateURL(c.Request.Context(), objectName)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "生成访问链接失败")
		return
	}

	response.Success(c, gin.H{
		"url": fileURL,
	})
}
