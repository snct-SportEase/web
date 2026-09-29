package handler

import (
	"backapp/internal/repository"
	"backapp/internal/safelog"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type updateTournamentSeedsRequest struct {
	ExpectedTeamIDs []int `json:"expected_team_ids"`
	TeamIDs         []int `json:"team_ids"`
}

func tournamentSeedIDs(c *gin.Context) (int, int, bool) {
	eventID, eventErr := strconv.Atoi(c.Param("id"))
	tournamentID, tournamentErr := strconv.Atoi(c.Param("tournament_id"))
	if eventErr != nil || tournamentErr != nil || eventID <= 0 || tournamentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "大会またはトーナメントIDが不正です"})
		return 0, 0, false
	}
	return eventID, tournamentID, true
}

func respondTournamentSeedError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		c.JSON(http.StatusNotFound, gin.H{"error": "トーナメントが見つかりません"})
	case errors.Is(err, repository.ErrTournamentSeedUnsupported):
		c.JSON(http.StatusBadRequest, gin.H{"error": "このトーナメントのシードはここでは変更できません"})
	case errors.Is(err, repository.ErrTournamentSeedLocked):
		c.JSON(http.StatusConflict, gin.H{"error": "試合が開始または記録されているため、シードを変更できません"})
	case errors.Is(err, repository.ErrTournamentSeedChanged):
		c.JSON(http.StatusConflict, gin.H{"error": "シード順が他の操作で変更されました。再読み込みしてください"})
	case errors.Is(err, repository.ErrInvalidTournamentSeed):
		c.JSON(http.StatusBadRequest, gin.H{"error": "シード順は現在の参加チームを重複なく指定してください"})
	default:
		log.Printf("[TournamentSeeds] failed: %s", safelog.Value(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "シード順を処理できませんでした"})
	}
}

func (h *TournamentHandler) GetTournamentSeedsHandler(c *gin.Context) {
	eventID, tournamentID, ok := tournamentSeedIDs(c)
	if !ok {
		return
	}
	seeds, err := h.tournRepo.GetTournamentSeedOrder(eventID, tournamentID)
	if err != nil {
		respondTournamentSeedError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"seeds": seeds})
}

func (h *TournamentHandler) UpdateTournamentSeedsHandler(c *gin.Context) {
	eventID, tournamentID, ok := tournamentSeedIDs(c)
	if !ok {
		return
	}
	var req updateTournamentSeedsRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.TeamIDs) < 2 || len(req.ExpectedTeamIDs) != len(req.TeamIDs) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "シード順の指定が不正です"})
		return
	}
	if err := h.tournRepo.UpdateTournamentSeedOrder(eventID, tournamentID, req.ExpectedTeamIDs, req.TeamIDs); err != nil {
		respondTournamentSeedError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "シード順を更新しました"})
}
