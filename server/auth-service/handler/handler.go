package handler

import (
	"net/http"
	"strings"

	"talking/server/auth-service/repository"
	"talking/server/auth-service/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	db           *repository.Database
	authProvider service.AuthProvider
}

func NewHandler(db *repository.Database, authProvider service.AuthProvider) *Handler {
	return &Handler{
		db:           db,
		authProvider: authProvider,
	}
}

// AuthMiddleware는 요청의 Authorization 헤더 내 Bearer JWT 토큰을 검증합니다.
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be Bearer token"})
			c.Abort()
			return
		}

		payload, err := h.authProvider.VerifyToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("userID", payload.UserID)
		c.Set("username", payload.Username)
		c.Set("nickname", payload.Nickname)
		c.Next()
	}
}
