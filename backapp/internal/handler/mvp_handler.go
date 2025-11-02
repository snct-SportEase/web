package handler

import (
	"net/http"
	"strconv"

	"backapp/internal/models"
	"backapp/internal/repository"

	"github.com/gin-gonic/gin"
)

type MVPHandler struct {
	mvpRepo repository.MVPRepository
}

func NewMVPHandler(mvpRepo repository.MVPRepository) *MVPHandler {
	return &MVPHandler{mvpRepo: mvpRepo}
}

func (h *MVPHandler) GetEligibleClasses(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Query("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_id"})
		return
	}

	classes, err := h.mvpRepo.GetEligibleClasses(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, classes)
}

type MVPVoteRequest struct {
	VotedForClassID int    `json:"voted_for_class_id"`
	Reason          string `json:"reason"`
	EventID         int    `json:"event_id"`
}

func (h *MVPHandler) VoteMVP(c *gin.Context) {
	req := new(MVPVoteRequest)
	if err := c.BindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	user, ok := userCtx.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type in context"})
		return
	}
	userID := user.ID

	err := h.mvpRepo.VoteMVP(userID, req.VotedForClassID, req.EventID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vote successful"})
}

func (h *MVPHandler) GetMVPVotes(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Query("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_id"})
		return
	}

	votes, err := h.mvpRepo.GetMVPVotes(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, votes)
}

func (h *MVPHandler) GetUserVote(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	user, ok := userCtx.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user type in context"})
		return
	}
	userID := user.ID

	eventID, err := strconv.Atoi(c.Query("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_id"})
		return
	}

	vote, err := h.mvpRepo.GetVoteByUserID(userID, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if vote == nil {
		c.JSON(http.StatusOK, gin.H{"voted": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"voted": true, "vote": vote})
}
