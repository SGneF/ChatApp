package middleware

import (
	"net/http"
	"strings"

	pkgjwt "chatapp-backend/pkg/jwt"
	"chatapp-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID   = "user_id"
	ContextUsername = "username"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := getTokenFromRequest(c)

		claims, err := pkgjwt.ParseToken(token)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "登录状态已失效，请重新登录")
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUsername, claims.Username)
		c.Next()
	}
}

func getTokenFromRequest(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")

	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	// 后续 WebSocket 连接时可用: ws?token=xxx
	token := c.Query("token")
	if token != "" {
		return token
	}

	return ""
}

func GetCurrentUserID(c *gin.Context) (uint64, bool) {
	value, exists := c.Get(ContextUserID)
	if !exists {
		return 0, false
	}

	userID, ok := value.(uint64)
	return userID, ok
}

func GetCurrentUsername(c *gin.Context) (string, bool) {
	value, exists := c.Get(ContextUsername)
	if !exists {
		return "", false
	}

	username, ok := value.(string)
	return username, ok
}
