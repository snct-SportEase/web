package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
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

func TestEventHandler_NotifySurvey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Notify Survey", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		mockNotificationRepo := new(MockNotificationRepository)
		mockClassRepo := new(MockClassRepository)
		_ = mockClassRepo
		_ = mockWhitelistRepo
		_ = mockTournamentRepo

		// Initialize with dummy VAPID keys to avoid skipping Push logic
		h := handler.NewEventHandler(mockEventRepo, nil, nil, nil, mockNotificationRepo, "dummyPublicKey", "dummyPrivateKey")

		eventID := 1
		surveyUrl := "https://forms.gle/dummy"
		event := &models.Event{
			ID:        eventID,
			Name:      "2025春季スポーツ大会",
			SurveyUrl: &surveyUrl,
		}

		mockEventRepo.On("GetEventByID", eventID).Return(event, nil).Once()
		mockNotificationRepo.On("CreateNotification", "大会アンケートのお願い", "大会に関するアンケート機能が公開されました。「2025春季スポーツ大会」についてダッシュボードの一番上のリンクからアンケートにご協力ください。", "test-user-id", &eventID).Return(int64(10), nil).Once()
		mockNotificationRepo.On("AddNotificationTargets", int64(10), []string{"student", "admin", "root"}).Return(nil).Once()
		mockEventRepo.On("PublishSurvey", eventID).Return(nil).Once()

		// For the push dispatch explicitly matching our targetRoles
		targetRoles := []string{"student", "admin", "root"}
		mockNotificationRepo.On("GetUserIDsByRoles", targetRoles).Return([]string{"user1", "user2"}, nil).Once()

		// Return empty subscriptions to safely pass the WebPush sending phase without making actual requests
		mockNotificationRepo.On("GetPushSubscriptionsByUserIDs", []string{"user1", "user2"}).Return([]models.PushSubscription{}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Set("user_id", "test-user-id")
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/events/1/notify-survey", nil)

		h.NotifySurvey(c)

		// Wait briefly for the goroutine to hit the mock
		time.Sleep(50 * time.Millisecond)

		assert.Equal(t, http.StatusOK, w.Code)
		mockEventRepo.AssertExpectations(t)
		mockNotificationRepo.AssertExpectations(t)
	})

	t.Run("Error - Event has no survey URL configured", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockNotificationRepo := new(MockNotificationRepository)

		h := handler.NewEventHandler(mockEventRepo, nil, nil, nil, mockNotificationRepo, "", "")

		eventID := 2
		event := &models.Event{
			ID:   eventID,
			Name: "2025秋季スポーツ大会", // No survey URL
		}

		mockEventRepo.On("GetEventByID", eventID).Return(event, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "2"}}
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/events/2/notify-survey", nil)

		h.NotifySurvey(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockEventRepo.AssertExpectations(t)
		mockNotificationRepo.AssertNotCalled(t, "CreateNotification")
	})

	t.Run("Error - Event not found", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockNotificationRepo := new(MockNotificationRepository)

		h := handler.NewEventHandler(mockEventRepo, nil, nil, nil, mockNotificationRepo, "", "")

		eventID := 999
		mockEventRepo.On("GetEventByID", eventID).Return((*models.Event)(nil), errors.New("not found")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/events/999/notify-survey", nil)

		h.NotifySurvey(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockEventRepo.AssertExpectations(t)
		mockNotificationRepo.AssertNotCalled(t, "CreateNotification")
	})
}

func TestEventHandler_ImportSurveyScores(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Error - Not a spring event", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)

		h := handler.NewEventHandler(mockEventRepo, nil, nil, nil, nil, "", "")

		eventID := 1
		event := &models.Event{ID: eventID, Season: "autumn"}

		mockEventRepo.On("GetEventByID", eventID).Return(event, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/events/1/import-survey-scores", nil)

		h.ImportSurveyScores(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Survey score import is only allowed for the spring event", response["error"])
	})

	t.Run("Success - Import Survey Scores", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		h := handler.NewEventHandler(mockEventRepo, nil, nil, mockClassRepo, nil, "", "")

		eventID := 1
		event := &models.Event{ID: eventID, Season: "spring"}

		mockEventRepo.On("GetEventByID", eventID).Return(event, nil).Once()

		classes := []*models.Class{
			{ID: 101, Name: "1A", StudentCount: 40},
			{ID: 102, Name: "2B", StudentCount: 38},
		}
		mockClassRepo.On("GetAllClasses", eventID).Return(classes, nil).Once()

		expectedPoints := map[int]int{
			101: 10, // 40/40 = 100% -> 10 points
			102: 6,  // 20/38 = 52.6% -> 6 points
		}
		mockClassRepo.On("SetSurveyPoints", eventID, expectedPoints).Return(nil).Once()

		// Prepare multipart form data
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "survey.csv")
		assert.NoError(t, err)

		csvContent := "タイムスタンプ,クラス名,氏名\n2025/01/01,1A,Student1\n"
		for i := 0; i < 39; i++ {
			csvContent += "2025/01/01,1A,User\n" // Total 40 for 1A
		}
		for i := 0; i < 20; i++ {
			csvContent += "2025/01/01,2B,User\n" // Total 20 for 2B
		}

		part.Write([]byte(csvContent))
		writer.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/events/1/import-survey-scores", body)
		c.Request.Header.Set("Content-Type", writer.FormDataContentType())

		h.ImportSurveyScores(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(2), response["imported_classes_count"])

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
	})

	t.Run("Error - Missing File", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		h := handler.NewEventHandler(mockEventRepo, nil, nil, nil, nil, "", "")

		eventID := 1
		event := &models.Event{ID: eventID, Season: "spring"}

		mockEventRepo.On("GetEventByID", eventID).Return(event, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/events/1/import-survey-scores", nil)
		// Missing multipart body

		h.ImportSurveyScores(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Error - Invalid CSV Format (No Class Column)", func(t *testing.T) {
		mockEventRepo := new(MockEventRepository)
		h := handler.NewEventHandler(mockEventRepo, nil, nil, nil, nil, "", "")

		eventID := 1
		event := &models.Event{ID: eventID, Season: "spring"}

		mockEventRepo.On("GetEventByID", eventID).Return(event, nil).Once()

		// Prepare multipart form data
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "survey.csv")

		csvContent := "タイムスタンプ,氏名\n2025/01/01,Student1\n"
		part.Write([]byte(csvContent))
		writer.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/events/1/import-survey-scores", body)
		c.Request.Header.Set("Content-Type", writer.FormDataContentType())

		h.ImportSurveyScores(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response["error"], "Could not find class name column")
	})
}
