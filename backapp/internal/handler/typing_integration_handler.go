package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"backapp/internal/models"
	"backapp/internal/repository"

	"github.com/gin-gonic/gin"
)

type TypingIntegrationHandler struct {
	repo repository.TypingIntegrationRepository
}

type updateTypingEntryOrderRequest struct {
	PlayerIDs []string `json:"player_ids"`
}

func NewTypingIntegrationHandler(repo repository.TypingIntegrationRepository) *TypingIntegrationHandler {
	return &TypingIntegrationHandler{repo: repo}
}

func (h *TypingIntegrationHandler) GetActiveEvent(c *gin.Context) {
	event, sports, err := h.repo.GetActiveEvent()
	if errors.Is(err, repository.ErrActiveEventNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Active event not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}

	c.JSON(http.StatusOK, models.TypingActiveEventResponse{
		APIVersion:  models.TypingIntegrationAPIVersion,
		GeneratedAt: time.Now().UTC(),
		Event:       event,
		Sports:      sports,
	})
}

func (h *TypingIntegrationHandler) GetCompetitionSnapshot(c *gin.Context) {
	eventID, err := positivePathID(c, "event_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}
	sportID, err := positivePathID(c, "sport_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sport ID"})
		return
	}

	snapshot, err := h.repo.GetCompetitionSnapshot(eventID, sportID)
	if errors.Is(err, repository.ErrTypingCompetitionNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event sport not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get typing competition"})
		return
	}

	snapshot.APIVersion = models.TypingIntegrationAPIVersion
	snapshot.GeneratedAt = time.Now().UTC()
	snapshot.Warnings = typingSnapshotWarnings(snapshot)
	c.JSON(http.StatusOK, snapshot)
}

func (h *TypingIntegrationHandler) UpdateTeamEntryOrder(c *gin.Context) {
	teamID, err := positivePathID(c, "team_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var request updateTypingEntryOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil || len(request.PlayerIDs) != 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player_ids must contain exactly 3 players"})
		return
	}
	seen := make(map[string]struct{}, len(request.PlayerIDs))
	for _, playerID := range request.PlayerIDs {
		if playerID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "player_ids must not contain an empty ID"})
			return
		}
		if _, duplicate := seen[playerID]; duplicate {
			c.JSON(http.StatusBadRequest, gin.H{"error": "player_ids must be unique"})
			return
		}
		seen[playerID] = struct{}{}
	}

	if err := h.repo.SetTeamEntryOrder(teamID, request.PlayerIDs); errors.Is(err, repository.ErrTypingRosterMismatch) {
		c.JSON(http.StatusConflict, gin.H{"error": "player_ids must exactly match the confirmed team roster"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update entry order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Entry order updated successfully"})
}

func positivePathID(c *gin.Context, parameter string) (int, error) {
	value, err := strconv.Atoi(c.Param(parameter))
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid %s", parameter)
	}
	return value, nil
}

func typingSnapshotWarnings(snapshot *models.TypingCompetitionSnapshot) []string {
	warnings := make([]string, 0)
	if len(snapshot.Teams) != 6 {
		warnings = append(warnings, fmt.Sprintf("expected 6 teams, got %d", len(snapshot.Teams)))
	}

	for _, team := range snapshot.Teams {
		if len(team.Players) != 3 {
			warnings = append(warnings, fmt.Sprintf(
				"team %d (%s) must have 3 confirmed players, got %d",
				team.ID,
				team.Name,
				len(team.Players),
			))
		}

		entryOrders := make(map[int]struct{}, len(team.Players))
		for _, player := range team.Players {
			entryOrders[player.EntryOrder] = struct{}{}
		}
		for expectedOrder := 1; expectedOrder <= len(team.Players); expectedOrder++ {
			if _, exists := entryOrders[expectedOrder]; !exists {
				warnings = append(warnings, fmt.Sprintf(
					"team %d (%s) has a non-contiguous entry order",
					team.ID,
					team.Name,
				))
				break
			}
		}
	}

	matchCount := 0
	for _, tournament := range snapshot.Tournaments {
		matchCount += len(tournament.Matches)
	}
	if matchCount == 0 {
		warnings = append(warnings, "no matches are registered")
	}

	return warnings
}
