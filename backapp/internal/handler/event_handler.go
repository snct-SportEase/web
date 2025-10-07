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
	eventRepo     repository.EventRepository
	whitelistRepo repository.WhitelistRepository
}

func NewEventHandler(eventRepo repository.EventRepository, whitelistRepo repository.WhitelistRepository) *EventHandler {
	return &EventHandler{eventRepo: eventRepo, whitelistRepo: whitelistRepo}
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
		c.JSON(http.StatusOK, gin.H{"event_id": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"event_id": event_id})
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
