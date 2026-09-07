package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RainyModeHandler struct {
	rainyModeRepo repository.RainyModeRepository
	eventRepo     repository.EventRepository
}

func eventIDParam(c *gin.Context) string {
	if eventID := c.Param("event_id"); eventID != "" {
		return eventID
	}
	return c.Param("id")
}

func NewRainyModeHandler(rainyModeRepo repository.RainyModeRepository, eventRepo repository.EventRepository) *RainyModeHandler {
	return &RainyModeHandler{
		rainyModeRepo: rainyModeRepo,
		eventRepo:     eventRepo,
	}
}

// GetRainyModeSettingsHandler は指定されたイベントの雨天時設定一覧を取得します
func (h *RainyModeHandler) GetRainyModeSettingsHandler(c *gin.Context) {
	eventID, err := strconv.Atoi(eventIDParam(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	settings, err := h.rainyModeRepo.GetSettingsByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rainy mode settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpsertRainyModeSettingHandler は雨天時設定を作成または更新します
func (h *RainyModeHandler) UpsertRainyModeSettingHandler(c *gin.Context) {
	eventID, err := strconv.Atoi(eventIDParam(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req models.RainyModeSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if req.MinCapacity != nil && *req.MinCapacity < 0 || req.MaxCapacity != nil && *req.MaxCapacity < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Capacities must be non-negative"})
		return
	}
	if req.MinCapacity != nil && req.MaxCapacity != nil && *req.MinCapacity > *req.MaxCapacity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_capacity must not exceed max_capacity"})
		return
	}

	setting := &models.RainyModeSetting{
		EventID:        eventID,
		SportID:        req.SportID,
		ClassID:        req.ClassID,
		MinCapacity:    req.MinCapacity,
		MaxCapacity:    req.MaxCapacity,
		MatchStartTime: req.MatchStartTime,
	}

	err = h.rainyModeRepo.UpsertSetting(setting)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidRainyModeSetting) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save rainy mode setting"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// DeleteRainyModeSettingHandler は雨天時設定を削除します
func (h *RainyModeHandler) DeleteRainyModeSettingHandler(c *gin.Context) {
	eventID, err := strconv.Atoi(eventIDParam(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	sportID, err := strconv.Atoi(c.Param("sport_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sport ID"})
		return
	}

	classID, err := strconv.Atoi(c.Param("class_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
		return
	}

	err = h.rainyModeRepo.DeleteSetting(eventID, sportID, classID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rainy mode setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rainy mode setting deleted successfully"})
}
