package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventRepo      repository.EventRepository
	whitelistRepo  repository.WhitelistRepository
	tournamentRepo repository.TournamentRepository
}

func NewEventHandler(eventRepo repository.EventRepository, whitelistRepo repository.WhitelistRepository, tournamentRepo repository.TournamentRepository) *EventHandler {
	return &EventHandler{eventRepo: eventRepo, whitelistRepo: whitelistRepo, tournamentRepo: tournamentRepo}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req struct {
		Name      string `json:"name"`
		Year      int    `json:"year"`
		Season    string `json:"season"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Date parsing
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD."})
			return
		}
		startDate = &t
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD."})
			return
		}
		endDate = &t
	}

	event := &models.Event{
		Name:       req.Name,
		Year:       req.Year,
		Season:     req.Season,
		Start_date: startDate,
		End_date:   endDate,
	}

	id, err := h.eventRepo.CreateEvent(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	event.ID = int(id)

	if event.Season == "autumn" {
		springEvent, err := h.eventRepo.GetEventByYearAndSeason(event.Year, "spring")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if springEvent != nil {
			err := h.eventRepo.CopyClassScores(springEvent.ID, event.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusCreated, event)
}

func (h *EventHandler) GetAllEvents(c *gin.Context) {
	events, err := h.eventRepo.GetAllEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

func (h *EventHandler) UpdateEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req struct {
		Name      string `json:"name"`
		Year      int    `json:"year"`
		Season    string `json:"season"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Date parsing
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD."})
			return
		}
		startDate = &t
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD."})
			return
		}
		endDate = &t
	}

	event := &models.Event{
		ID:         id,
		Name:       req.Name,
		Year:       req.Year,
		Season:     req.Season,
		Start_date: startDate,
		End_date:   endDate,
	}

	err = h.eventRepo.UpdateEvent(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

func (h *EventHandler) GetActiveEvent(c *gin.Context) {
	event_id, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if event_id == 0 {
		c.JSON(http.StatusOK, gin.H{"event_id": nil, "event_name": nil, "competition_guidelines_pdf_url": nil})
		return
	}

	event, err := h.eventRepo.GetEventByID(event_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if event == nil {
		c.JSON(http.StatusOK, gin.H{"event_id": nil, "event_name": nil, "competition_guidelines_pdf_url": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event_id":                       event.ID,
		"event_name":                     event.Name,
		"competition_guidelines_pdf_url": event.CompetitionGuidelinesPdfUrl,
	})
}

func (h *EventHandler) SetActiveEvent(c *gin.Context) {
	// 1. リクエストボディの構造体を定義
	req := models.SetActiveEventRequest{}

	// 2. リクエストボディをパースし、バリデーション
	if err := c.ShouldBindJSON(&req); err != nil {
		// バリデーションエラーやJSONパースエラーの場合、400 Bad Requestを返す
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body or missing event_id", "details": err.Error()})
		return
	}

	// Pass pointer to allow nil -> clearing active event
	err := h.eventRepo.SetActiveEvent(req.EventID)
	if err != nil {
		// DB更新に失敗した場合、500 Internal Server Errorを返す
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set active event", "details": err.Error()})
		return
	}

	// If a new event is being activated (not cleared), update the whitelist
	if req.EventID != nil {
		if err := h.whitelistRepo.UpdateNullEventIDs(*req.EventID); err != nil {
			// このエラーはクリティカルではないので、ログには残すがクライアントには成功を返す
			// c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update whitelist event IDs", "details": err.Error()})
			// return
			// TODO: Add proper logging here
		}
	}

	// 4. 成功レスポンスを返す
	c.JSON(http.StatusOK, gin.H{
		"message":  "Active event set successfully",
		"event_id": req.EventID,
	})
}

func (h *EventHandler) SetRainyMode(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req struct {
		IsRainyMode bool `json:"is_rainy_mode"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.eventRepo.SetRainyMode(eventID, req.IsRainyMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If rainy mode is being enabled, apply rainy mode start times to tournaments
	if req.IsRainyMode {
		err = h.tournamentRepo.ApplyRainyModeStartTimes(eventID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply rainy mode start times", "details": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rainy mode updated successfully", "is_rainy_mode": req.IsRainyMode})
}

func (h *EventHandler) UpdateCompetitionGuidelines(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req struct {
		PdfUrl *string `json:"pdf_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := h.eventRepo.GetEventByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	event.CompetitionGuidelinesPdfUrl = req.PdfUrl
	err = h.eventRepo.UpdateEvent(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}
