package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UpdateMatchStartTimeRequest struct {
	StartTime string `json:"start_time"`
}

func (h *TournamentHandler) UpdateMatchStartTimeHandler(c *gin.Context) {
	matchID, err := strconv.Atoi(c.Param("match_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match ID"})
		return
	}

	var req UpdateMatchStartTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.tournRepo.UpdateMatchStartTime(matchID, req.StartTime); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update match start time"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match start time updated successfully"})
}

type UpdateMatchStatusRequest struct {
	Status string `json:"status"`
}

func (h *TournamentHandler) UpdateMatchStatusHandler(c *gin.Context) {
	matchID, err := strconv.Atoi(c.Param("match_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match ID"})
		return
	}

	var req UpdateMatchStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.tournRepo.UpdateMatchStatus(matchID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update match status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match status updated successfully"})
}

type UpdateMatchResultRequest struct {
	Team1Score int `json:"team1_score"`
	Team2Score int `json:"team2_score"`
	WinnerID   int `json:"winner_id,omitempty"`
}

func (h *TournamentHandler) UpdateMatchResultHandler(c *gin.Context) {
	matchID, err := strconv.Atoi(c.Param("match_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match ID"})
		return
	}

	var req UpdateMatchResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.tournRepo.UpdateMatchResult(matchID, req.Team1Score, req.Team2Score, req.WinnerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update match result", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match result updated successfully"})
}
