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

func TestSportHandler_UpdateCapacityHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Update both min and max capacity", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		eventID := 1
		sportID := 2
		minCapacity := 10
		maxCapacity := 20

		// Current sport details
		currentDetails := &models.EventSport{
			EventID:     eventID,
			SportID:     sportID,
			Description: stringPtr("Test sport"),
			Rules:       stringPtr("Test rules"),
			RulesType:   "markdown",
			Location:    "gym1",
		}

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: &minCapacity,
			MaxCapacity: &maxCapacity,
		}

		mockSportRepo.On("GetSportDetails", eventID, sportID).Return(currentDetails, nil).Once()
		mockSportRepo.On("UpdateSportDetails", eventID, sportID, mock.MatchedBy(func(details models.EventSport) bool {
			return details.MinCapacity != nil && *details.MinCapacity == minCapacity &&
				details.MaxCapacity != nil && *details.MaxCapacity == maxCapacity
		})).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCapacityHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Capacity updated successfully", response["message"])
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("Success - Update only min capacity", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		eventID := 1
		sportID := 2
		minCapacity := 10

		currentDetails := &models.EventSport{
			EventID:   eventID,
			SportID:   sportID,
			RulesType: "markdown",
			Location:  "gym1",
		}

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: &minCapacity,
			MaxCapacity: nil,
		}

		mockSportRepo.On("GetSportDetails", eventID, sportID).Return(currentDetails, nil).Once()
		mockSportRepo.On("UpdateSportDetails", eventID, sportID, mock.MatchedBy(func(details models.EventSport) bool {
			return details.MinCapacity != nil && *details.MinCapacity == minCapacity &&
				details.MaxCapacity == nil
		})).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCapacityHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("Success - Update only max capacity", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		eventID := 1
		sportID := 2
		maxCapacity := 20

		currentDetails := &models.EventSport{
			EventID:   eventID,
			SportID:   sportID,
			RulesType: "markdown",
			Location:  "gym1",
		}

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: nil,
			MaxCapacity: &maxCapacity,
		}

		mockSportRepo.On("GetSportDetails", eventID, sportID).Return(currentDetails, nil).Once()
		mockSportRepo.On("UpdateSportDetails", eventID, sportID, mock.MatchedBy(func(details models.EventSport) bool {
			return details.MinCapacity == nil &&
				details.MaxCapacity != nil && *details.MaxCapacity == maxCapacity
		})).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCapacityHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("Success - Clear both capacities", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		eventID := 1
		sportID := 2

		currentDetails := &models.EventSport{
			EventID:     eventID,
			SportID:     sportID,
			MinCapacity: intPtr(10),
			MaxCapacity: intPtr(20),
			RulesType:   "markdown",
			Location:    "gym1",
		}

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: nil,
			MaxCapacity: nil,
		}

		mockSportRepo.On("GetSportDetails", eventID, sportID).Return(currentDetails, nil).Once()
		mockSportRepo.On("UpdateSportDetails", eventID, sportID, mock.MatchedBy(func(details models.EventSport) bool {
			return details.MinCapacity == nil && details.MaxCapacity == nil
		})).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCapacityHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("Error - Min capacity greater than max capacity", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		minCapacity := 20
		maxCapacity := 10

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: &minCapacity,
			MaxCapacity: &maxCapacity,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCapacityHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "最低定員は最高定員以下である必要があります")
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid event ID", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: intPtr(10),
			MaxCapacity: intPtr(20),
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "invalid"},
			gin.Param{Key: "sport_id", Value: "2"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/invalid/sports/2/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCapacityHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid event ID", response["error"])
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid sport ID", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: intPtr(10),
			MaxCapacity: intPtr(20),
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "invalid"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/invalid/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCapacityHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid sport ID", response["error"])
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid request body", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
		}

		invalidJSON := `{"min_capacity": "invalid"}`
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/capacity", bytes.NewBufferString(invalidJSON))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCapacityHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request body", response["error"])
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("Error - GetSportDetails fails", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		eventID := 1
		sportID := 2
		minCapacity := 10
		maxCapacity := 20

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: &minCapacity,
			MaxCapacity: &maxCapacity,
		}

		mockSportRepo.On("GetSportDetails", eventID, sportID).Return(nil, assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCapacityHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to retrieve sport details", response["error"])
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("Error - UpdateSportDetails fails", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		eventID := 1
		sportID := 2
		minCapacity := 10
		maxCapacity := 20

		currentDetails := &models.EventSport{
			EventID:   eventID,
			SportID:   sportID,
			RulesType: "markdown",
			Location:  "gym1",
		}

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: &minCapacity,
			MaxCapacity: &maxCapacity,
		}

		mockSportRepo.On("GetSportDetails", eventID, sportID).Return(currentDetails, nil).Once()
		mockSportRepo.On("UpdateSportDetails", eventID, sportID, mock.Anything).Return(assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateCapacityHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to update capacity", response["error"])
		mockSportRepo.AssertExpectations(t)
	})
}

func TestSportHandler_UpdateClassCapacityHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Update both min and max capacity for class", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		eventID := 1
		sportID := 2
		classID := 3
		minCapacity := 10
		maxCapacity := 20

		// Existing team
		existingTeam := &models.Team{
			ID:      1,
			Name:    "Test Team",
			ClassID: classID,
			SportID: sportID,
			EventID: eventID,
		}

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: &minCapacity,
			MaxCapacity: &maxCapacity,
		}

		mockTeamRepo.On("GetTeamByClassAndSport", classID, sportID, eventID).Return(existingTeam, nil).Once()
		mockTeamRepo.On("UpdateTeamCapacity", eventID, sportID, classID, &minCapacity, &maxCapacity).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
			gin.Param{Key: "class_id", Value: "3"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/classes/3/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateClassCapacityHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Class capacity updated successfully", response["message"])
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Success - Update only min capacity for class", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		eventID := 1
		sportID := 2
		classID := 3
		minCapacity := 10

		existingTeam := &models.Team{
			ID:      1,
			Name:    "Test Team",
			ClassID: classID,
			SportID: sportID,
			EventID: eventID,
		}

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: &minCapacity,
			MaxCapacity: nil,
		}

		mockTeamRepo.On("GetTeamByClassAndSport", classID, sportID, eventID).Return(existingTeam, nil).Once()
		mockTeamRepo.On("UpdateTeamCapacity", eventID, sportID, classID, &minCapacity, (*int)(nil)).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
			gin.Param{Key: "class_id", Value: "3"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/classes/3/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateClassCapacityHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Error - Min capacity greater than max capacity", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		minCapacity := 20
		maxCapacity := 10

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: &minCapacity,
			MaxCapacity: &maxCapacity,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
			gin.Param{Key: "class_id", Value: "3"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/classes/3/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateClassCapacityHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "最低定員は最高定員以下である必要があります")
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid event ID", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: intPtr(10),
			MaxCapacity: intPtr(20),
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "invalid"},
			gin.Param{Key: "sport_id", Value: "2"},
			gin.Param{Key: "class_id", Value: "3"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/invalid/sports/2/classes/3/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateClassCapacityHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid event ID", response["error"])
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid sport ID", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: intPtr(10),
			MaxCapacity: intPtr(20),
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "invalid"},
			gin.Param{Key: "class_id", Value: "3"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/invalid/classes/3/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateClassCapacityHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid sport ID", response["error"])
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid class ID", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: intPtr(10),
			MaxCapacity: intPtr(20),
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
			gin.Param{Key: "class_id", Value: "invalid"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/classes/invalid/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateClassCapacityHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid class ID", response["error"])
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Error - Team not found", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		eventID := 1
		sportID := 2
		classID := 3
		minCapacity := 10
		maxCapacity := 20

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: &minCapacity,
			MaxCapacity: &maxCapacity,
		}

		mockTeamRepo.On("GetTeamByClassAndSport", classID, sportID, eventID).Return(nil, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
			gin.Param{Key: "class_id", Value: "3"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/classes/3/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateClassCapacityHandler(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Team not found")
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Error - GetTeamByClassAndSport fails", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		eventID := 1
		sportID := 2
		classID := 3
		minCapacity := 10
		maxCapacity := 20

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: &minCapacity,
			MaxCapacity: &maxCapacity,
		}

		mockTeamRepo.On("GetTeamByClassAndSport", classID, sportID, eventID).Return(nil, assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
			gin.Param{Key: "class_id", Value: "3"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/classes/3/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateClassCapacityHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to retrieve team", response["error"])
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Error - UpdateTeamCapacity fails", func(t *testing.T) {
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, mockEventRepo, mockTournamentRepo)

		eventID := 1
		sportID := 2
		classID := 3
		minCapacity := 10
		maxCapacity := 20

		existingTeam := &models.Team{
			ID:      1,
			Name:    "Test Team",
			ClassID: classID,
			SportID: sportID,
			EventID: eventID,
		}

		reqBody := struct {
			MinCapacity *int `json:"min_capacity"`
			MaxCapacity *int `json:"max_capacity"`
		}{
			MinCapacity: &minCapacity,
			MaxCapacity: &maxCapacity,
		}

		mockTeamRepo.On("GetTeamByClassAndSport", classID, sportID, eventID).Return(existingTeam, nil).Once()
		mockTeamRepo.On("UpdateTeamCapacity", eventID, sportID, classID, &minCapacity, &maxCapacity).Return(assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			gin.Param{Key: "event_id", Value: "1"},
			gin.Param{Key: "sport_id", Value: "2"},
			gin.Param{Key: "class_id", Value: "3"},
		}

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPut, "/api/admin/events/1/sports/2/classes/3/capacity", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateClassCapacityHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to update capacity", response["error"])
		mockTeamRepo.AssertExpectations(t)
	})
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
