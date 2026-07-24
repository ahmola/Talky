package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateRoomRequest struct {
	Name    string   `json:"name" binding:"required"`
	Type    string   `json:"type" binding:"required"` // 'direct' or 'group'
	Members []string `json:"members" binding:"required"`
}

func (h *Handler) GetRooms(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	rooms, err := h.db.GetRoomsForUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch rooms"})
		return
	}

	c.JSON(http.StatusOK, rooms)
}

func (h *Handler) CreateRoom(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 본인을 참여 멤버 목록에 강제 포함 (중복 제거)
	memberSet := make(map[string]bool)
	memberSet[userID] = true
	for _, m := range req.Members {
		memberSet[m] = true
	}

	uniqueMembers := make([]string, 0, len(memberSet))
	for m := range memberSet {
		uniqueMembers = append(uniqueMembers, m)
	}

	roomID, err := h.db.CreateRoom(c.Request.Context(), req.Name, req.Type, uniqueMembers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create room"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"roomId": roomID,
	})
}

func (h *Handler) GetRoomMessages(c *gin.Context) {
	roomID := c.Param("roomId")
	limitStr := c.DefaultQuery("limit", "50")
	beforeStr := c.Query("before")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	var before time.Time
	if beforeStr != "" {
		before, err = time.Parse(time.RFC3339, beforeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid before time format, must be RFC3339"})
			return
		}
	}

	// 현재 사용자가 방 멤버인지 보안 검증
	userID := c.MustGet("userID").(string)
	members, err := h.db.GetRoomMembers(c.Request.Context(), roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify room membership"})
		return
	}

	isMember := false
	for _, m := range members {
		if m == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not authorized to view messages in this room"})
		return
	}

	messages, err := h.db.GetRoomMessages(c.Request.Context(), roomID, limit, before)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch room messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}
