package handler

import (
	"log"
	"net/http"
	"strings"

	"talking/server/auth-service/service"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Nickname = strings.TrimSpace(req.Nickname)

	// 1. 유저네임 중복 체크
	_, err := h.db.GetUserByUsername(c.Request.Context(), req.Username)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username already exists"})
		return
	}

	// 2. 유저 생성
	userID, err := h.db.CreateUser(c.Request.Context(), req.Username, req.Password, req.Nickname)
	if err != nil {
		log.Printf("[Register] Failed to create user '%s': %v", req.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	log.Printf("[Register] Successfully registered user '%s' (ID: %s)", req.Username, userID)
	c.JSON(http.StatusCreated, gin.H{
		"message": "registration successful",
		"userId":  userID,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Username = strings.TrimSpace(req.Username)

	user, err := h.db.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		log.Printf("[Login] User not found '%s': %v", req.Username, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	if !h.db.CheckPassword(user.PasswordHash, req.Password) {
		log.Printf("[Login] Password mismatch for user '%s'", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// 3. 토큰 발급
	token, err := h.authProvider.GenerateToken(service.TokenPayload{
		UserID:   user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"userId":   user.ID,
		"nickname": user.Nickname,
	})
}
