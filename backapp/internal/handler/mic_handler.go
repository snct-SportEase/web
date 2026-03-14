package handler

import (
	"net/http"
	"strconv"

	"backapp/internal/models"
	"backapp/internal/repository"

	"github.com/gin-gonic/gin"
)

type MICHandler struct {
	micRepo repository.MICRepository
}

func NewMICHandler(micRepo repository.MICRepository) *MICHandler {
	return &MICHandler{micRepo: micRepo}
}

func (h *MICHandler) GetEligibleClasses(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Query("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_id"})
		return
	}

	classes, err := h.micRepo.GetEligibleClasses(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, classes)
}

type MICVoteRequest struct {
	VotedForClassID int    `json:"voted_for_class_id"`
	Reason          string `json:"reason"`
	EventID         int    `json:"event_id"`
}

func (h *MICHandler) VoteMIC(c *gin.Context) {
	req := new(MICVoteRequest)
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

	err := h.micRepo.VoteMIC(userID, req.VotedForClassID, req.EventID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vote successful"})
}

func (h *MICHandler) GetMICVotes(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Query("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_id"})
		return
	}

	votes, err := h.micRepo.GetMICVotes(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, votes)
}

func (h *MICHandler) GetUserVote(c *gin.Context) {
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

	vote, err := h.micRepo.GetVoteByUserID(userID, eventID)
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

func (h *MICHandler) GetMICClass(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Query("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event_id"})
		return
	}

	micResult, err := h.micRepo.GetMICClass(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if micResult == nil {
		c.JSON(http.StatusOK, gin.H{"message": "No MIC class found yet"})
		return
	}

	c.JSON(http.StatusOK, micResult)
}
