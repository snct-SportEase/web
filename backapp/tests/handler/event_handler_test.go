package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Tests ---

func TestEventHandler_CreateEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Create Spring Event", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockWhitelistRepo := new(MockWhitelistRepository)

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo)

		startDate := time.Now()
		endDate := startDate.Add(24 * time.Hour)
		reqBody := struct {
			Name      string `json:"name"`
			Year      int    `json:"year"`
			Season    string `json:"season"`
			StartDate string `json:"start_date"`
			EndDate   string `json:"end_date"`
		}{
			Name:       "2025春季スポーツ大会",
			Year:       2025,
			Season:     "spring",
			StartDate: startDate.Format("2006-01-02"),
			EndDate:   endDate.Format("2006-01-02"),
		}

		mockEventRepo.On("CreateEvent", mock.AnythingOfType("*models.Event")).Return(int64(1), nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/events", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.CreateEvent(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("Success - Create Autumn Event with Spring Event existing", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockWhitelistRepo := new(MockWhitelistRepository)

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo)

		springEvent := &models.Event{ID: 1, Year: 2025, Season: "spring"}
		autumnEventReq := struct {
			Name      string `json:"name"`
			Year      int    `json:"year"`
			Season    string `json:"season"`
			StartDate string `json:"start_date"`
			EndDate   string `json:"end_date"`
		}{
			Name:       "2025秋季スポーツ大会",
			Year:       2025,
			Season:     "autumn",
			StartDate:  time.Now().Format("2006-01-02"),
			EndDate:    time.Now().Add(24 * time.Hour).Format("2006-01-02"),
		}

		mockEventRepo.On("CreateEvent", mock.AnythingOfType("*models.Event")).Return(int64(2), nil).Once()
		mockEventRepo.On("GetEventByYearAndSeason", 2025, "spring").Return(springEvent, nil).Once()
		mockEventRepo.On("CopyClassScores", springEvent.ID, 2).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(autumnEventReq)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/events", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.CreateEvent(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("Success - Create Autumn Event without Spring Event", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockWhitelistRepo := new(MockWhitelistRepository)

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo)

		autumnEventReq := struct {
			Name      string `json:"name"`
			Year      int    `json:"year"`
			Season    string `json:"season"`
			StartDate string `json:"start_date"`
			EndDate   string `json:"end_date"`
		}{
			Name:       "2026秋季スポーツ大会",
			Year:       2026,
			Season:     "autumn",
			StartDate:  time.Now().Format("2006-01-02"),
			EndDate:    time.Now().Add(24 * time.Hour).Format("2006-01-02"),
		}

		mockEventRepo.On("CreateEvent", mock.AnythingOfType("*models.Event")).Return(int64(3), nil).Once()
		mockEventRepo.On("GetEventByYearAndSeason", 2026, "spring").Return((*models.Event)(nil), nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(autumnEventReq)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/events", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.CreateEvent(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockEventRepo.AssertNotCalled(t, "CopyClassScores", mock.Anything, mock.Anything)
		mockEventRepo.AssertExpectations(t)
	})
}
