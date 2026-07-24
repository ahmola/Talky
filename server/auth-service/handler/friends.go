package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AddFriendRequest struct {
	Username string `json:"username" binding:"required"`
}

func (h *Handler) GetFriends(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	friends, err := h.db.GetFriends(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch friends"})
		return
	}

	c.JSON(http.StatusOK, friends)
}

func (h *Handler) AddFriend(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	username := c.MustGet("username").(string)

	var req AddFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Username == username {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot add yourself as a friend"})
		return
	}

	// 1. 대상 유저 정보 가져오기
	friend, err := h.db.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// 2. 이미 친구인지 검증
	isAlready, err := h.db.IsFriend(c.Request.Context(), userID, friend.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check friendship status"})
		return
	}
	if isAlready {
		c.JSON(http.StatusBadRequest, gin.H{"error": "already friends"})
		return
	}

	// 3. 친구 추가 진행
	err = h.db.AddFriend(c.Request.Context(), userID, friend.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add friend"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "friend added successfully"})
}

func (h *Handler) DeleteFriend(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	friendID := c.Param("userId")

	err := h.db.DeleteFriend(c.Request.Context(), userID, friendID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete friend"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "friend deleted successfully"})
}
