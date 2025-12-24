package handler

import (
	"net/http"
	"strconv"

	"backapp/internal/models"

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

type UpdateMatchRainyModeStartTimeRequest struct {
	RainyModeStartTime string `json:"rainy_mode_start_time"`
}

func (h *TournamentHandler) UpdateMatchRainyModeStartTimeHandler(c *gin.Context) {
	matchID, err := strconv.Atoi(c.Param("match_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match ID"})
		return
	}

	var req UpdateMatchRainyModeStartTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.tournRepo.UpdateMatchRainyModeStartTime(matchID, req.RainyModeStartTime); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update match rainy mode start time"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match rainy mode start time updated successfully"})
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

	// 既に入力済みの試合結果かどうかをチェック
	alreadyEntered, err := h.tournRepo.IsMatchResultAlreadyEntered(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check match status", "details": err.Error()})
		return
	}

	// 既に入力済みの場合は、root権限のみ許可
	if alreadyEntered {
		userVal, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			return
		}
		user, ok := userVal.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
			return
		}

		// root権限を持っているかチェック
		hasRootRole := false
		for _, role := range user.Roles {
			if role.Name == "root" {
				hasRootRole = true
				break
			}
		}

		if !hasRootRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "既に入力済みの試合結果の修正はroot権限のみ可能です"})
			return
		}
	}

	var req UpdateMatchResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 既に入力済みの場合は修正用メソッドを使用（次の試合のチームも更新）
	if alreadyEntered {
		if err := h.tournRepo.UpdateMatchResultForCorrection(matchID, req.Team1Score, req.Team2Score, req.WinnerID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to correct match result", "details": err.Error()})
			return
		}
	} else {
		// 未入力の場合は通常の更新メソッドを使用
		if err := h.tournRepo.UpdateMatchResult(matchID, req.Team1Score, req.Team2Score, req.WinnerID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update match result", "details": err.Error()})
			return
		}
	}

	tournamentID, err := h.tournRepo.GetTournamentIDByMatchID(matchID)
	if err == nil && h.hubManager != nil {
		h.hubManager.BroadcastTo("tournament:"+strconv.Itoa(tournamentID), gin.H{"type": "update"})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Match result updated successfully"})
}
