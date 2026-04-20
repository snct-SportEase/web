package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"backapp/internal/repository"
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

func TestSportHandler_GetAllSportsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		repository.GlobalCache.Flush()
		mockSportRepo := new(MockSportRepository)
		h := handler.NewSportHandler(mockSportRepo, nil, nil, nil, nil)

		sports := []*models.Sport{{ID: 1, Name: "Soccer"}, {ID: 2, Name: "Basketball"}}
		for _, name := range []string{"学年対抗リレー", "コース対抗リレー", "綱引き"} {
			mockSportRepo.On("GetSportByName", name).Return(nil, nil).Once()
			mockSportRepo.On("CreateSport", mock.MatchedBy(func(sport *models.Sport) bool {
				return sport != nil && sport.Name == name
			})).Return(int64(10), nil).Once()
		}
		mockSportRepo.On("GetAllSports").Return(sports, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		h.GetAllSportsHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []*models.Sport
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 2)
		assert.Equal(t, "Soccer", response[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		repository.GlobalCache.Flush()
		mockSportRepo := new(MockSportRepository)
		h := handler.NewSportHandler(mockSportRepo, nil, nil, nil, nil)

		for _, name := range []string{"学年対抗リレー", "コース対抗リレー", "綱引き"} {
			mockSportRepo.On("GetSportByName", name).Return(nil, nil).Once()
			mockSportRepo.On("CreateSport", mock.MatchedBy(func(sport *models.Sport) bool {
				return sport != nil && sport.Name == name
			})).Return(int64(10), nil).Once()
		}
		mockSportRepo.On("GetAllSports").Return(nil, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		h.GetAllSportsHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "[]", w.Body.String())
	})

	t.Run("Error", func(t *testing.T) {
		repository.GlobalCache.Flush()
		mockSportRepo := new(MockSportRepository)
		h := handler.NewSportHandler(mockSportRepo, nil, nil, nil, nil)

		for _, name := range []string{"学年対抗リレー", "コース対抗リレー", "綱引き"} {
			mockSportRepo.On("GetSportByName", name).Return(nil, nil).Once()
			mockSportRepo.On("CreateSport", mock.MatchedBy(func(sport *models.Sport) bool {
				return sport != nil && sport.Name == name
			})).Return(int64(10), nil).Once()
		}
		mockSportRepo.On("GetAllSports").Return(nil, assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		h.GetAllSportsHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSportHandler_CreateSportHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		repository.GlobalCache.Flush()
		mockSportRepo := new(MockSportRepository)
		h := handler.NewSportHandler(mockSportRepo, nil, nil, nil, nil)

		sport := models.Sport{Name: "Tennis"}
		mockSportRepo.On("CreateSport", mock.Anything).Return(int64(3), nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		jsonBody, _ := json.Marshal(sport)
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))

		h.CreateSportHandler(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response models.Sport
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, 3, response.ID)
		assert.Equal(t, "Tennis", response.Name)
	})

	t.Run("Invalid Body", func(t *testing.T) {
		h := handler.NewSportHandler(nil, nil, nil, nil, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("invalid"))
		h.CreateSportHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Empty Name", func(t *testing.T) {
		h := handler.NewSportHandler(nil, nil, nil, nil, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		jsonBody, _ := json.Marshal(models.Sport{Name: ""})
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))
		h.CreateSportHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestSportHandler_GetSportsByEventHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		repository.GlobalCache.Flush()
		mockSportRepo := new(MockSportRepository)
		h := handler.NewSportHandler(mockSportRepo, nil, nil, nil, nil)

		eventSports := []*models.EventSport{{EventID: 1, SportID: 1, Location: "Gym"}}
		mockSportRepo.On("GetSportsByEventID", 1).Return(eventSports, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.GetSportsByEventHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []*models.EventSport
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 1)
		assert.Equal(t, "Gym", response[0].Location)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		h := handler.NewSportHandler(nil, nil, nil, nil, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "abc"}}
		h.GetSportsByEventHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestSportHandler_AssignSportToEventHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		repository.GlobalCache.Flush()
		mockSportRepo := new(MockSportRepository)
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		h := handler.NewSportHandler(mockSportRepo, mockClassRepo, mockTeamRepo, nil, nil)

		eventSport := models.EventSport{SportID: 1}
		classes := []*models.Class{{ID: 1, Name: "1-1"}}
		sport := models.Sport{ID: 1, Name: "Soccer"}

		mockSportRepo.On("AssignSportToEvent", mock.Anything).Return(nil).Once()
		mockSportRepo.On("GetSportByID", 1).Return(&sport, nil).Once()
		mockClassRepo.On("GetAllClasses", 1).Return(classes, nil).Once()
		mockTeamRepo.On("CreateTeamsBulk", mock.Anything).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "event_id", Value: "1"}}
		jsonBody, _ := json.Marshal(eventSport)
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))

		h.AssignSportToEventHandler(c)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Noon Game Error", func(t *testing.T) {
		h := handler.NewSportHandler(nil, nil, nil, nil, nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "event_id", Value: "1"}}
		jsonBody, _ := json.Marshal(models.EventSport{Location: "noon_game"})
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))

		h.AssignSportToEventHandler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestSportHandler_DeleteSportFromEventHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		repository.GlobalCache.Flush()
		mockSportRepo := new(MockSportRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockTournRepo := new(MockTournamentRepository)
		h := handler.NewSportHandler(mockSportRepo, nil, mockTeamRepo, nil, mockTournRepo)

		mockTournRepo.On("DeleteTournamentsByEventAndSportID", 1, 2).Return(nil).Once()
		mockTeamRepo.On("DeleteTeamsByEventAndSportID", 1, 2).Return(nil).Once()
		mockSportRepo.On("DeleteSportFromEvent", 1, 2).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "event_id", Value: "1"}, {Key: "sport_id", Value: "2"}}
		h.DeleteSportFromEventHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSportHandler_GetTeamsBySportHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		repository.GlobalCache.Flush()
		mockSportRepo := new(MockSportRepository)
		h := handler.NewSportHandler(mockSportRepo, nil, nil, nil, nil)

		teams := []*models.Team{{ID: 1, Name: "Team 1"}}
		mockSportRepo.On("GetTeamsBySportID", 1).Return(teams, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		h.GetTeamsBySportHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []*models.Team
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 1)
	})
}

func TestSportHandler_GetSportDetailsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		repository.GlobalCache.Flush()
		mockSportRepo := new(MockSportRepository)
		h := handler.NewSportHandler(mockSportRepo, nil, nil, nil, nil)

		details := &models.EventSport{EventID: 1, SportID: 2, Location: "Gym"}
		mockSportRepo.On("GetSportDetails", 1, 2).Return(details, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "event_id", Value: "1"}, {Key: "sport_id", Value: "2"}}
		h.GetSportDetailsHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.EventSport
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Gym", response.Location)
	})
}

func TestSportHandler_UpdateSportDetailsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		repository.GlobalCache.Flush()
		mockSportRepo := new(MockSportRepository)
		h := handler.NewSportHandler(mockSportRepo, nil, nil, nil, nil)

		details := models.EventSport{Description: stringPtr("New Desc")}
		mockSportRepo.On("UpdateSportDetails", 1, 2, mock.Anything).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "event_id", Value: "1"}, {Key: "sport_id", Value: "2"}}
		jsonBody, _ := json.Marshal(details)
		c.Request, _ = http.NewRequest("PUT", "/", bytes.NewBuffer(jsonBody))

		h.UpdateSportDetailsHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
