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
		mockTournamentRepo := new(MockTournamentRepository)
		mockNotificationRepo := new(MockNotificationRepository)
		mockClassRepo := new(MockClassRepository)
		_ = mockClassRepo
		_ = mockClassRepo

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo, mockTournamentRepo, nil, mockNotificationRepo, "", "")

		startDate := time.Now()
		endDate := startDate.Add(24 * time.Hour)
		reqBody := struct {
			Name      string `json:"name"`
			Year      int    `json:"year"`
			Season    string `json:"season"`
			StartDate string `json:"start_date"`
			EndDate   string `json:"end_date"`
		}{
			Name:      "2025春季スポーツ大会",
			Year:      2025,
			Season:    "spring",
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
		mockTournamentRepo := new(MockTournamentRepository)
		mockNotificationRepo := new(MockNotificationRepository)
		mockClassRepo := new(MockClassRepository)
		_ = mockClassRepo

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo, mockTournamentRepo, nil, mockNotificationRepo, "", "")

		springEvent := &models.Event{ID: 1, Year: 2025, Season: "spring"}
		autumnEventReq := struct {
			Name      string `json:"name"`
			Year      int    `json:"year"`
			Season    string `json:"season"`
			StartDate string `json:"start_date"`
			EndDate   string `json:"end_date"`
		}{
			Name:      "2025秋季スポーツ大会",
			Year:      2025,
			Season:    "autumn",
			StartDate: time.Now().Format("2006-01-02"),
			EndDate:   time.Now().Add(24 * time.Hour).Format("2006-01-02"),
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
		mockTournamentRepo := new(MockTournamentRepository)
		mockNotificationRepo := new(MockNotificationRepository)
		mockClassRepo := new(MockClassRepository)
		_ = mockClassRepo

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo, mockTournamentRepo, nil, mockNotificationRepo, "", "")

		autumnEventReq := struct {
			Name      string `json:"name"`
			Year      int    `json:"year"`
			Season    string `json:"season"`
			StartDate string `json:"start_date"`
			EndDate   string `json:"end_date"`
		}{
			Name:      "2026秋季スポーツ大会",
			Year:      2026,
			Season:    "autumn",
			StartDate: time.Now().Format("2006-01-02"),
			EndDate:   time.Now().Add(24 * time.Hour).Format("2006-01-02"),
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

func TestEventHandler_UpdateCompetitionGuidelines(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Update competition guidelines with PDF URL", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockNotificationRepo := new(MockNotificationRepository)
		mockClassRepo := new(MockClassRepository)
		_ = mockClassRepo

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo, mockTournamentRepo, nil, mockNotificationRepo, "", "")

		eventID := 1
		pdfUrl := "https://example.com/uploads/pdfs/guidelines.pdf"
		event := &models.Event{
			ID:                          eventID,
			Name:                        "2025春季スポーツ大会",
			Year:                        2025,
			Season:                      "spring",
			CompetitionGuidelinesPdfUrl: nil,
		}

		reqBody := map[string]interface{}{
			"pdf_url": pdfUrl,
		}

		mockEventRepo.On("GetEventByID", eventID).Return(event, nil).Once()
		mockEventRepo.On("UpdateEvent", mock.MatchedBy(func(e *models.Event) bool {
			return e.ID == eventID && e.CompetitionGuidelinesPdfUrl != nil && *e.CompetitionGuidelinesPdfUrl == pdfUrl
		})).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/root/events/1/competition-guidelines", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCompetitionGuidelines(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("Success - Delete competition guidelines (set to null)", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockNotificationRepo := new(MockNotificationRepository)
		mockClassRepo := new(MockClassRepository)
		_ = mockClassRepo

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo, mockTournamentRepo, nil, mockNotificationRepo, "", "")

		eventID := 1
		existingPdfUrl := "https://example.com/uploads/pdfs/old-guidelines.pdf"
		event := &models.Event{
			ID:                          eventID,
			Name:                        "2025春季スポーツ大会",
			Year:                        2025,
			Season:                      "spring",
			CompetitionGuidelinesPdfUrl: &existingPdfUrl,
		}

		reqBody := map[string]interface{}{
			"pdf_url": nil,
		}

		mockEventRepo.On("GetEventByID", eventID).Return(event, nil).Once()
		mockEventRepo.On("UpdateEvent", mock.MatchedBy(func(e *models.Event) bool {
			return e.ID == eventID && e.CompetitionGuidelinesPdfUrl == nil
		})).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/root/events/1/competition-guidelines", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCompetitionGuidelines(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid event ID", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockNotificationRepo := new(MockNotificationRepository)
		mockClassRepo := new(MockClassRepository)
		_ = mockClassRepo

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo, mockTournamentRepo, nil, mockNotificationRepo, "", "")

		reqBody := map[string]interface{}{
			"pdf_url": "https://example.com/uploads/pdfs/guidelines.pdf",
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/root/events/invalid/competition-guidelines", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCompetitionGuidelines(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response["error"], "Invalid event ID")
		mockEventRepo.AssertNotCalled(t, "GetEventByID", mock.Anything)
		mockEventRepo.AssertNotCalled(t, "UpdateEvent", mock.Anything)
	})

	t.Run("Error - Event not found", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockNotificationRepo := new(MockNotificationRepository)
		mockClassRepo := new(MockClassRepository)
		_ = mockClassRepo

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo, mockTournamentRepo, nil, mockNotificationRepo, "", "")

		eventID := 999
		reqBody := map[string]interface{}{
			"pdf_url": "https://example.com/uploads/pdfs/guidelines.pdf",
		}

		mockEventRepo.On("GetEventByID", eventID).Return((*models.Event)(nil), nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/root/events/999/competition-guidelines", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCompetitionGuidelines(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response["error"], "Event not found")
		mockEventRepo.AssertExpectations(t)
		mockEventRepo.AssertNotCalled(t, "UpdateEvent", mock.Anything)
	})

	t.Run("Error - Invalid JSON", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockNotificationRepo := new(MockNotificationRepository)
		mockClassRepo := new(MockClassRepository)
		_ = mockClassRepo
		_ = mockClassRepo
		_ = mockClassRepo

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo, mockTournamentRepo, nil, mockNotificationRepo, "", "")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		invalidJSON := "{invalid json}"
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/root/events/1/competition-guidelines", bytes.NewBufferString(invalidJSON))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCompetitionGuidelines(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockEventRepo.AssertNotCalled(t, "GetEventByID", mock.Anything)
		mockEventRepo.AssertNotCalled(t, "UpdateEvent", mock.Anything)
	})

	t.Run("Error - GetEventByID returns error", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockNotificationRepo := new(MockNotificationRepository)
		mockClassRepo := new(MockClassRepository)
		_ = mockClassRepo
		_ = mockClassRepo

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo, mockTournamentRepo, nil, mockNotificationRepo, "", "")

		eventID := 1
		reqBody := map[string]interface{}{
			"pdf_url": "https://example.com/uploads/pdfs/guidelines.pdf",
		}

		mockEventRepo.On("GetEventByID", eventID).Return((*models.Event)(nil), assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/root/events/1/competition-guidelines", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCompetitionGuidelines(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockEventRepo.AssertExpectations(t)
		mockEventRepo.AssertNotCalled(t, "UpdateEvent", mock.Anything)
	})

	t.Run("Error - UpdateEvent returns error", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockNotificationRepo := new(MockNotificationRepository)
		mockClassRepo := new(MockClassRepository)
		_ = mockClassRepo
		_ = mockClassRepo

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo, mockTournamentRepo, nil, mockNotificationRepo, "", "")

		eventID := 1
		pdfUrl := "https://example.com/uploads/pdfs/guidelines.pdf"
		event := &models.Event{
			ID:                          eventID,
			Name:                        "2025春季スポーツ大会",
			Year:                        2025,
			Season:                      "spring",
			CompetitionGuidelinesPdfUrl: nil,
		}

		reqBody := map[string]interface{}{
			"pdf_url": pdfUrl,
		}

		mockEventRepo.On("GetEventByID", eventID).Return(event, nil).Once()
		mockEventRepo.On("UpdateEvent", mock.AnythingOfType("*models.Event")).Return(assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/root/events/1/competition-guidelines", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCompetitionGuidelines(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockEventRepo.AssertExpectations(t)
	})
}

func TestEventHandler_GetActiveEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Get active event with competition guidelines", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockNotificationRepo := new(MockNotificationRepository)
		mockClassRepo := new(MockClassRepository)
		_ = mockClassRepo

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo, mockTournamentRepo, mockClassRepo, mockNotificationRepo, "", "")

		eventID := 1
		pdfUrl := "https://example.com/uploads/pdfs/guidelines.pdf"
		event := &models.Event{
			ID:                          eventID,
			Name:                        "2025春季スポーツ大会",
			Year:                        2025,
			Season:                      "spring",
			CompetitionGuidelinesPdfUrl: &pdfUrl,
		}

		mockEventRepo.On("GetActiveEvent").Return(eventID, nil).Once()
		mockEventRepo.On("GetEventByID", eventID).Return(event, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/events/active", nil)

		h.GetActiveEvent(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(eventID), response["event_id"])
		assert.Equal(t, event.Name, response["event_name"])
		assert.Equal(t, pdfUrl, response["competition_guidelines_pdf_url"])
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("Success - Get active event without competition guidelines", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockNotificationRepo := new(MockNotificationRepository)
		mockClassRepo := new(MockClassRepository)
		_ = mockClassRepo

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo, mockTournamentRepo, mockClassRepo, mockNotificationRepo, "", "")

		eventID := 1
		event := &models.Event{
			ID:                          eventID,
			Name:                        "2025春季スポーツ大会",
			Year:                        2025,
			Season:                      "spring",
			CompetitionGuidelinesPdfUrl: nil,
		}

		mockEventRepo.On("GetActiveEvent").Return(eventID, nil).Once()
		mockEventRepo.On("GetEventByID", eventID).Return(event, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/events/active", nil)

		h.GetActiveEvent(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(eventID), response["event_id"])
		assert.Equal(t, event.Name, response["event_name"])
		assert.Nil(t, response["competition_guidelines_pdf_url"])
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("Success - No active event", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockNotificationRepo := new(MockNotificationRepository)
		mockClassRepo := new(MockClassRepository)

		h := handler.NewEventHandler(mockEventRepo, mockWhitelistRepo, mockTournamentRepo, mockClassRepo, mockNotificationRepo, "", "")

		mockEventRepo.On("GetActiveEvent").Return(0, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/events/active", nil)

		h.GetActiveEvent(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Nil(t, response["event_id"])
		assert.Nil(t, response["event_name"])
		assert.Nil(t, response["competition_guidelines_pdf_url"])
		mockEventRepo.AssertExpectations(t)
		mockEventRepo.AssertNotCalled(t, "GetEventByID", mock.Anything)
	})
}
