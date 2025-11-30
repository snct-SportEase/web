package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRainyModeHandler_GetRainyModeSettingsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Get settings", func(t *testing.T) {
		mockRainyModeRepo := new(MockRainyModeRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewRainyModeHandler(mockRainyModeRepo, mockEventRepo)

		eventID := 1
		settings := []*models.RainyModeSetting{
			{
				ID:          1,
				EventID:     eventID,
				SportID:     1,
				ClassID:     1,
				MinCapacity: intPtr(5),
				MaxCapacity: intPtr(10),
			},
		}

		mockRainyModeRepo.On("GetSettingsByEventID", eventID).Return(settings, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		h.GetRainyModeSettingsHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []*models.RainyModeSetting
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Len(t, result, 1)

		mockRainyModeRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid event ID", func(t *testing.T) {
		mockRainyModeRepo := new(MockRainyModeRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewRainyModeHandler(mockRainyModeRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		h.GetRainyModeSettingsHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockRainyModeRepo.AssertNotCalled(t, "GetSettingsByEventID", mock.Anything)
	})
}

func TestRainyModeHandler_UpsertRainyModeSettingHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Create new setting", func(t *testing.T) {
		mockRainyModeRepo := new(MockRainyModeRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewRainyModeHandler(mockRainyModeRepo, mockEventRepo)

		eventID := 1
		reqBody := models.RainyModeSettingRequest{
			SportID:        1,
			ClassID:        1,
			MinCapacity:    intPtr(5),
			MaxCapacity:    intPtr(10),
			MatchStartTime: stringPtr("09:00"),
		}

		mockRainyModeRepo.On("UpsertSetting", mock.MatchedBy(func(setting *models.RainyModeSetting) bool {
			return setting.EventID == eventID &&
				setting.SportID == reqBody.SportID &&
				setting.ClassID == reqBody.ClassID &&
				setting.MinCapacity != nil && *setting.MinCapacity == 5 &&
				setting.MaxCapacity != nil && *setting.MaxCapacity == 10 &&
				setting.MatchStartTime != nil && *setting.MatchStartTime == "09:00"
		})).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/events/1/rainy-mode/settings", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpsertRainyModeSettingHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRainyModeRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid event ID", func(t *testing.T) {
		mockRainyModeRepo := new(MockRainyModeRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewRainyModeHandler(mockRainyModeRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		h.UpsertRainyModeSettingHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockRainyModeRepo.AssertNotCalled(t, "UpsertSetting", mock.Anything)
	})

	t.Run("Error - Invalid request body", func(t *testing.T) {
		mockRainyModeRepo := new(MockRainyModeRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewRainyModeHandler(mockRainyModeRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		jsonBody := []byte("invalid json")
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/root/events/1/rainy-mode/settings", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpsertRainyModeSettingHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockRainyModeRepo.AssertNotCalled(t, "UpsertSetting", mock.Anything)
	})
}

func TestRainyModeHandler_DeleteRainyModeSettingHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Delete setting", func(t *testing.T) {
		mockRainyModeRepo := new(MockRainyModeRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewRainyModeHandler(mockRainyModeRepo, mockEventRepo)

		eventID := 1
		sportID := 1
		classID := 1

		mockRainyModeRepo.On("DeleteSetting", eventID, sportID, classID).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "sport_id", Value: "1"},
			{Key: "class_id", Value: "1"},
		}

		h.DeleteRainyModeSettingHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRainyModeRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid event ID", func(t *testing.T) {
		mockRainyModeRepo := new(MockRainyModeRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewRainyModeHandler(mockRainyModeRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "event_id", Value: "invalid"},
			{Key: "sport_id", Value: "1"},
			{Key: "class_id", Value: "1"},
		}

		h.DeleteRainyModeSettingHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockRainyModeRepo.AssertNotCalled(t, "DeleteSetting", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Error - Invalid sport ID", func(t *testing.T) {
		mockRainyModeRepo := new(MockRainyModeRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewRainyModeHandler(mockRainyModeRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "event_id", Value: "1"},
			{Key: "sport_id", Value: "invalid"},
			{Key: "class_id", Value: "1"},
		}

		h.DeleteRainyModeSettingHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockRainyModeRepo.AssertNotCalled(t, "DeleteSetting", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Error - Invalid class ID", func(t *testing.T) {
		mockRainyModeRepo := new(MockRainyModeRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewRainyModeHandler(mockRainyModeRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "event_id", Value: "1"},
			{Key: "sport_id", Value: "1"},
			{Key: "class_id", Value: "invalid"},
		}

		h.DeleteRainyModeSettingHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockRainyModeRepo.AssertNotCalled(t, "DeleteSetting", mock.Anything, mock.Anything, mock.Anything)
	})
}
