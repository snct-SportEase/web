package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBarcodeHandler_GetUserTeamsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Get user teams", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository))

		userID := "test-user-id"
		displayName := "Test User"
		user := &models.User{
			ID:          userID,
			DisplayName: &displayName,
		}

		expectedTeams := []*models.TeamWithSport{
			{
				ID:        1,
				Name:      "Team A",
				ClassID:   1,
				SportID:   1,
				EventID:   1,
				SportName: "Basketball",
			},
			{
				ID:        2,
				Name:      "Team B",
				ClassID:   2,
				SportID:   2,
				EventID:   1,
				SportName: "Soccer",
			},
		}

		mockTeamRepo.On("GetTeamsByUserID", userID).Return(expectedTeams, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/barcode/teams", nil)

		h.GetUserTeamsHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var actualTeams []*models.TeamWithSport
		err := json.Unmarshal(w.Body.Bytes(), &actualTeams)
		assert.NoError(t, err)
		assert.Equal(t, expectedTeams, actualTeams)

		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Failure - User not in context", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/barcode/teams", nil)

		h.GetUserTeamsHandler(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User not found in context", response["error"])

		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - Repository error", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository))

		userID := "test-user-id"
		displayName := "Test User"
		user := &models.User{
			ID:          userID,
			DisplayName: &displayName,
		}

		mockTeamRepo.On("GetTeamsByUserID", userID).Return(nil, assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/barcode/teams", nil)

		h.GetUserTeamsHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to retrieve teams", response["error"])

		mockTeamRepo.AssertExpectations(t)
	})
}

func TestBarcodeHandler_GenerateBarcodeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Generate barcode", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		displayName := "Test User"
		user := &models.User{
			ID:          userID,
			DisplayName: &displayName,
		}

		reqBody := models.BarcodeRequest{
			EventID: 1,
			SportID: 1,
		}

		teams := []*models.TeamWithSport{
			{
				ID:        1,
				Name:      "Team A",
				ClassID:   1,
				SportID:   1,
				EventID:   1,
				SportName: "Basketball",
			},
		}

		mockTeamRepo.On("GetTeamsByUserID", userID).Return(teams, nil).Once()
		mockTokenStore.On("SaveActiveToken", userID, reqBody.EventID, reqBody.SportID, mock.AnythingOfType("string"), 10*time.Second).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/generate", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.GenerateBarcodeHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.BarcodeResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.BarcodeData)
		assert.Greater(t, response.ExpiresAt, time.Now().Unix())

		// Verify barcode data can be decoded
		decoded, err := base64.URLEncoding.DecodeString(response.BarcodeData)
		assert.NoError(t, err)

		var combinedData map[string]interface{}
		err = json.Unmarshal(decoded, &combinedData)
		assert.NoError(t, err)
		assert.Contains(t, combinedData, "token")
		assert.Contains(t, combinedData, "data")

		// Verify barcode data content
		dataBytes, _ := json.Marshal(combinedData["data"])
		var barcodeTokenData models.BarcodeData
		err = json.Unmarshal(dataBytes, &barcodeTokenData)
		assert.NoError(t, err)
		assert.Equal(t, reqBody.EventID, barcodeTokenData.EventID)
		assert.Equal(t, reqBody.SportID, barcodeTokenData.SportID)
		assert.Equal(t, "Basketball", barcodeTokenData.SportName)
		assert.Equal(t, userID, barcodeTokenData.UserID)
		assert.Equal(t, displayName, barcodeTokenData.DisplayName)
		assert.Equal(t, response.ExpiresAt, barcodeTokenData.ExpiresAt)
		assert.Equal(t, barcodeTokenData.Timestamp+10, barcodeTokenData.ExpiresAt)

		mockTeamRepo.AssertExpectations(t)
		mockTokenStore.AssertExpectations(t)
	})

	t.Run("Failure - User not in context", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		reqBody := models.BarcodeRequest{
			EventID: 1,
			SportID: 1,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/generate", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.GenerateBarcodeHandler(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User not found in context", response["error"])

		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
		mockTokenStore.AssertNotCalled(t, "SaveActiveToken")
	})

	t.Run("Failure - Invalid request body", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		displayName := "Test User"
		user := &models.User{
			ID:          userID,
			DisplayName: &displayName,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)

		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/generate", bytes.NewBufferString("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		h.GenerateBarcodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request body", response["error"])

		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
		mockTokenStore.AssertNotCalled(t, "SaveActiveToken")
	})

	t.Run("Failure - User not assigned to sport", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		displayName := "Test User"
		user := &models.User{
			ID:          userID,
			DisplayName: &displayName,
		}

		reqBody := models.BarcodeRequest{
			EventID: 1,
			SportID: 99, // User is not assigned to this sport
		}

		teams := []*models.TeamWithSport{
			{
				ID:        1,
				Name:      "Team A",
				ClassID:   1,
				SportID:   1,
				EventID:   1,
				SportName: "Basketball",
			},
		}

		mockTeamRepo.On("GetTeamsByUserID", userID).Return(teams, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/generate", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.GenerateBarcodeHandler(c)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "この競技に参加登録されていません", response["error"])

		mockTeamRepo.AssertExpectations(t)
		mockTokenStore.AssertNotCalled(t, "SaveActiveToken")
	})

	t.Run("Failure - Token store error", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		displayName := "Test User"
		user := &models.User{
			ID:          userID,
			DisplayName: &displayName,
		}

		reqBody := models.BarcodeRequest{
			EventID: 1,
			SportID: 1,
		}

		teams := []*models.TeamWithSport{
			{
				ID:        1,
				Name:      "Team A",
				ClassID:   1,
				SportID:   1,
				EventID:   1,
				SportName: "Basketball",
			},
		}

		mockTeamRepo.On("GetTeamsByUserID", userID).Return(teams, nil).Once()
		mockTokenStore.On("SaveActiveToken", userID, reqBody.EventID, reqBody.SportID, mock.AnythingOfType("string"), 10*time.Second).Return(assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/generate", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.GenerateBarcodeHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to store barcode token", response["error"])

		mockTeamRepo.AssertExpectations(t)
		mockTokenStore.AssertExpectations(t)
	})

	t.Run("Failure - Repository error", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		displayName := "Test User"
		user := &models.User{
			ID:          userID,
			DisplayName: &displayName,
		}

		reqBody := models.BarcodeRequest{
			EventID: 1,
			SportID: 1,
		}

		mockTeamRepo.On("GetTeamsByUserID", userID).Return(nil, assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/generate", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.GenerateBarcodeHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to retrieve teams", response["error"])

		mockTeamRepo.AssertExpectations(t)
		mockTokenStore.AssertNotCalled(t, "SaveActiveToken")
	})

	t.Run("Success - User without display name", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		user := &models.User{
			ID:          userID,
			DisplayName: nil,
		}

		reqBody := models.BarcodeRequest{
			EventID: 1,
			SportID: 1,
		}

		teams := []*models.TeamWithSport{
			{
				ID:        1,
				Name:      "Team A",
				ClassID:   1,
				SportID:   1,
				EventID:   1,
				SportName: "Basketball",
			},
		}

		mockTeamRepo.On("GetTeamsByUserID", userID).Return(teams, nil).Once()
		mockTokenStore.On("SaveActiveToken", userID, reqBody.EventID, reqBody.SportID, mock.AnythingOfType("string"), 10*time.Second).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/generate", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.GenerateBarcodeHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.BarcodeResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Verify display name is empty string
		decoded, _ := base64.URLEncoding.DecodeString(response.BarcodeData)
		var combinedData map[string]interface{}
		json.Unmarshal(decoded, &combinedData)
		dataBytes, _ := json.Marshal(combinedData["data"])
		var barcodeTokenData models.BarcodeData
		json.Unmarshal(dataBytes, &barcodeTokenData)
		assert.Equal(t, "", barcodeTokenData.DisplayName)

		mockTeamRepo.AssertExpectations(t)
		mockTokenStore.AssertExpectations(t)
	})
}

func TestBarcodeHandler_VerifyBarcodeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Verify valid barcode and confirm participation", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, mockClassRepo).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		displayName := "Test User"
		classID := 1
		teamID := 1

		// Create a valid barcode data
		timestamp := time.Now().Unix()
		expiresAt := timestamp + 10
		barcodeTokenData := models.BarcodeData{
			EventID:     1,
			SportID:     1,
			SportName:   "Basketball",
			UserID:      userID,
			DisplayName: displayName,
			Timestamp:   timestamp,
			ExpiresAt:   expiresAt,
		}

		// Encode barcode
		combinedData := map[string]interface{}{
			"token": "test-token",
			"data":  barcodeTokenData,
		}
		combinedJSON, _ := json.Marshal(combinedData)
		barcodeData := base64.URLEncoding.EncodeToString(combinedJSON)

		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: barcodeData,
		}

		teams := []*models.TeamWithSport{
			{
				ID:        teamID,
				Name:      "Team A",
				ClassID:   classID,
				SportID:   1,
				EventID:   1,
				SportName: "Basketball",
			},
		}

		team := &models.Team{
			ID:          teamID,
			Name:        "Team A",
			ClassID:     classID,
			SportID:     1,
			EventID:     1,
			MinCapacity: nil, // No capacity requirement
			MaxCapacity: nil,
		}

		mockTeamRepo.On("GetTeamsByUserID", userID).Return(teams, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", classID, 1, 1).Return(team, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", teamID, userID).Return(nil).Once()
		mockTokenStore.On("GetActiveToken", userID, 1, 1).Return("test-token", nil).Once()
		mockTokenStore.On("ConsumeActiveToken", userID, 1, 1, "test-token").Return(true, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, true, response["valid"])
		assert.Equal(t, float64(1), response["event_id"])
		assert.Equal(t, float64(1), response["sport_id"])
		assert.Equal(t, "Basketball", response["sport_name"])
		assert.Equal(t, userID, response["user_id"])
		assert.Equal(t, displayName, response["display_name"])
		assert.Greater(t, response["remaining_time"], float64(0))

		mockTeamRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockTokenStore.AssertExpectations(t)
	})

	t.Run("Success - Verify MyID barcode and confirm participation", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, mockClassRepo).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		displayName := "Test User"
		classID := 1
		teamID := 1
		studentNumber := "2301059"

		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     1,
		}

		user := &models.User{
			ID:          userID,
			Email:       "2301059@sendai-nct.jp",
			DisplayName: &displayName,
		}

		teams := []*models.TeamWithSport{
			{
				ID:        teamID,
				Name:      "Team A",
				ClassID:   classID,
				SportID:   1,
				EventID:   1,
				SportName: "Basketball",
			},
		}

		team := &models.Team{
			ID:          teamID,
			Name:        "Team A",
			ClassID:     classID,
			SportID:     1,
			EventID:     1,
			MinCapacity: nil,
			MaxCapacity: nil,
		}

		mockUserRepo.On("FindUsers", studentNumber, "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", userID).Return(teams, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", classID, 1, 1).Return(team, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", teamID, userID).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, true, response["valid"])
		assert.Equal(t, float64(1), response["event_id"])
		assert.Equal(t, float64(1), response["sport_id"])
		assert.Equal(t, "Basketball", response["sport_name"])
		assert.Equal(t, userID, response["user_id"])
		assert.Equal(t, displayName, response["display_name"])
		assert.Equal(t, studentNumber, response["student_number"])

		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTokenStore.AssertNotCalled(t, "GetActiveToken")
	})

	t.Run("Success - Verify trimmed MyID barcode and match exact email local part", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, mockClassRepo).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		displayName := "Test User"
		classID := 1
		teamID := 1
		studentNumber := "2301059"

		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: "  H102301059  ",
			EventID:     1,
			SportID:     1,
		}

		partialMatchUser := &models.User{
			ID:    "partial-user-id",
			Email: "x2301059@sendai-nct.jp",
		}
		user := &models.User{
			ID:          userID,
			Email:       "2301059@sendai-nct.jp",
			DisplayName: &displayName,
		}

		teams := []*models.TeamWithSport{
			{
				ID:        teamID,
				Name:      "Team A",
				ClassID:   classID,
				SportID:   1,
				EventID:   1,
				SportName: "Basketball",
			},
		}

		team := &models.Team{
			ID:      teamID,
			Name:    "Team A",
			ClassID: classID,
			SportID: 1,
			EventID: 1,
		}

		mockUserRepo.On("FindUsers", studentNumber, "email").Return([]*models.User{partialMatchUser, user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", userID).Return(teams, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", classID, 1, 1).Return(team, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", teamID, userID).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, true, response["valid"])
		assert.Equal(t, userID, response["user_id"])
		assert.Equal(t, studentNumber, response["student_number"])
		assert.Equal(t, "H102301059", response["barcode_data"])

		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID", "partial-user-id")
		mockTokenStore.AssertNotCalled(t, "GetActiveToken")
	})

	t.Run("Success - Verify MyID barcode with capacity warning", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, mockClassRepo).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		displayName := "Test User"
		classID := 1
		teamID := 1
		minCapacity := 5
		confirmedCount := 3
		studentNumber := "2301059"

		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     1,
		}

		user := &models.User{
			ID:          userID,
			Email:       "2301059@sendai-nct.jp",
			DisplayName: &displayName,
		}

		teams := []*models.TeamWithSport{
			{
				ID:        teamID,
				Name:      "Team A",
				ClassID:   classID,
				SportID:   1,
				EventID:   1,
				SportName: "Basketball",
			},
		}

		team := &models.Team{
			ID:          teamID,
			Name:        "Team A",
			ClassID:     classID,
			SportID:     1,
			EventID:     1,
			MinCapacity: &minCapacity,
		}

		class := &models.Class{
			ID:   classID,
			Name: "1-1",
		}

		mockUserRepo.On("FindUsers", studentNumber, "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", userID).Return(teams, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", classID, 1, 1).Return(team, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", teamID, userID).Return(nil).Once()
		mockTeamRepo.On("GetConfirmedTeamMembersCount", teamID).Return(confirmedCount, nil).Once()
		mockClassRepo.On("GetClassByID", classID).Return(class, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, true, response["valid"])
		assert.Contains(t, response, "capacity_warning")
		assert.Contains(t, response["capacity_warning"], "1-1クラス")
		assert.Contains(t, response["capacity_warning"], "最低人数")

		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockTokenStore.AssertNotCalled(t, "GetActiveToken")
	})

	t.Run("Failure - MyID barcode missing event or sport id", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: "H102301059",
			EventID:     1,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "イベントIDと競技IDが必要です", response["error"])

		mockUserRepo.AssertNotCalled(t, "FindUsers")
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
		mockTokenStore.AssertNotCalled(t, "GetActiveToken")
	})

	t.Run("Failure - Invalid MyID barcode format", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: "2301059",
			EventID:     1,
			SportID:     1,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "バーコード形式が不正です", response["error"])

		mockUserRepo.AssertNotCalled(t, "FindUsers")
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
		mockTokenStore.AssertNotCalled(t, "GetActiveToken")
	})

	t.Run("Failure - MyID barcode user not found by exact email local part", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		studentNumber := "2301059"
		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     1,
		}

		users := []*models.User{
			{
				ID:    "partial-user-id",
				Email: "x2301059@sendai-nct.jp",
			},
			{
				ID:    "domain-user-id",
				Email: "student@example.com",
			},
		}

		mockUserRepo.On("FindUsers", studentNumber, "email").Return(users, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "該当する学生が見つかりません", response["error"])

		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
		mockTokenStore.AssertNotCalled(t, "GetActiveToken")
	})

	t.Run("Failure - MyID barcode user lookup error", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		studentNumber := "2301059"
		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     1,
		}

		mockUserRepo.On("FindUsers", studentNumber, "email").Return(nil, assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to find user by student number", response["error"])

		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
		mockTokenStore.AssertNotCalled(t, "GetActiveToken")
	})

	t.Run("Failure - MyID barcode user not assigned to sport", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		studentNumber := "2301059"
		user := &models.User{
			ID:    userID,
			Email: "2301059@sendai-nct.jp",
		}

		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     99,
		}

		teams := []*models.TeamWithSport{
			{
				ID:        1,
				Name:      "Team A",
				ClassID:   1,
				SportID:   1,
				EventID:   1,
				SportName: "Basketball",
			},
		}

		mockUserRepo.On("FindUsers", studentNumber, "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", userID).Return(teams, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "このユーザーはこの競技に参加登録されていません", response["error"])

		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTokenStore.AssertNotCalled(t, "GetActiveToken")
	})

	t.Run("Failure - MyID barcode confirm member error", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		classID := 1
		teamID := 1
		studentNumber := "2301059"
		user := &models.User{
			ID:    userID,
			Email: "2301059@sendai-nct.jp",
		}

		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     1,
		}

		teams := []*models.TeamWithSport{
			{
				ID:        teamID,
				Name:      "Team A",
				ClassID:   classID,
				SportID:   1,
				EventID:   1,
				SportName: "Basketball",
			},
		}

		team := &models.Team{
			ID:      teamID,
			Name:    "Team A",
			ClassID: classID,
			SportID: 1,
			EventID: 1,
		}

		mockUserRepo.On("FindUsers", studentNumber, "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", userID).Return(teams, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", classID, 1, 1).Return(team, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", teamID, userID).Return(assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to confirm team member", response["error"])

		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTokenStore.AssertNotCalled(t, "GetActiveToken")
	})

	t.Run("Success - Verify barcode with capacity warning", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, mockClassRepo).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		displayName := "Test User"
		classID := 1
		teamID := 1
		minCapacity := 5
		confirmedCount := 3 // Below minimum

		// Create a valid barcode data
		timestamp := time.Now().Unix()
		expiresAt := timestamp + 10
		barcodeTokenData := models.BarcodeData{
			EventID:     1,
			SportID:     1,
			SportName:   "Basketball",
			UserID:      userID,
			DisplayName: displayName,
			Timestamp:   timestamp,
			ExpiresAt:   expiresAt,
		}

		// Encode barcode
		combinedData := map[string]interface{}{
			"token": "test-token",
			"data":  barcodeTokenData,
		}
		combinedJSON, _ := json.Marshal(combinedData)
		barcodeData := base64.URLEncoding.EncodeToString(combinedJSON)

		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: barcodeData,
		}

		teams := []*models.TeamWithSport{
			{
				ID:        teamID,
				Name:      "Team A",
				ClassID:   classID,
				SportID:   1,
				EventID:   1,
				SportName: "Basketball",
			},
		}

		team := &models.Team{
			ID:          teamID,
			Name:        "Team A",
			ClassID:     classID,
			SportID:     1,
			EventID:     1,
			MinCapacity: &minCapacity,
			MaxCapacity: nil,
		}

		class := &models.Class{
			ID:   classID,
			Name: "1-1",
		}

		mockTeamRepo.On("GetTeamsByUserID", userID).Return(teams, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", classID, 1, 1).Return(team, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", teamID, userID).Return(nil).Once()
		mockTeamRepo.On("GetConfirmedTeamMembersCount", teamID).Return(confirmedCount, nil).Once()
		mockClassRepo.On("GetClassByID", classID).Return(class, nil).Once()
		mockTokenStore.On("GetActiveToken", userID, 1, 1).Return("test-token", nil).Once()
		mockTokenStore.On("ConsumeActiveToken", userID, 1, 1, "test-token").Return(true, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, true, response["valid"])
		assert.Contains(t, response, "capacity_warning")
		assert.Contains(t, response["capacity_warning"], "1-1クラス")
		assert.Contains(t, response["capacity_warning"], "最低人数")

		mockTeamRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockTokenStore.AssertExpectations(t)
	})

	t.Run("Failure - Invalid request body", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBufferString("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request body", response["error"])

		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
		mockTokenStore.AssertNotCalled(t, "GetActiveToken")
	})

	t.Run("Failure - Invalid barcode token", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		barcodeTokenData := models.BarcodeData{
			EventID:     1,
			SportID:     1,
			SportName:   "Basketball",
			UserID:      "test-user-id",
			DisplayName: "Test User",
			Timestamp:   time.Now().Unix(),
			ExpiresAt:   time.Now().Unix() + 10,
		}

		combinedJSON, _ := json.Marshal(map[string]interface{}{
			"data": barcodeTokenData,
		})
		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: base64.URLEncoding.EncodeToString(combinedJSON),
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid barcode token", response["error"])

		mockTokenStore.AssertNotCalled(t, "GetActiveToken")
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - Invalid barcode format", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: "invalid-base64-!@#$",
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid barcode format", response["error"])

		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
		mockTokenStore.AssertNotCalled(t, "GetActiveToken")
	})

	t.Run("Failure - Expired barcode", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		displayName := "Test User"

		// Create an expired barcode data
		timestamp := time.Now().Unix() - 600 // 10 minutes ago
		expiresAt := timestamp + 10
		barcodeTokenData := models.BarcodeData{
			EventID:     1,
			SportID:     1,
			SportName:   "Basketball",
			UserID:      userID,
			DisplayName: displayName,
			Timestamp:   timestamp,
			ExpiresAt:   expiresAt,
		}

		// Encode barcode
		combinedData := map[string]interface{}{
			"token": "test-token",
			"data":  barcodeTokenData,
		}
		combinedJSON, _ := json.Marshal(combinedData)
		barcodeData := base64.URLEncoding.EncodeToString(combinedJSON)

		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: barcodeData,
		}

		mockTokenStore.On("GetActiveToken", userID, 1, 1).Return("test-token", nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "バーコードの有効期限が切れています", response["error"])
		assert.Equal(t, true, response["expired"])

		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
		mockTokenStore.AssertExpectations(t)
	})

	t.Run("Failure - Token validation error", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		barcodeTokenData := models.BarcodeData{
			EventID:     1,
			SportID:     1,
			SportName:   "Basketball",
			UserID:      userID,
			DisplayName: "Test User",
			Timestamp:   time.Now().Unix(),
			ExpiresAt:   time.Now().Unix() + 10,
		}

		combinedJSON, _ := json.Marshal(map[string]interface{}{
			"token": "test-token",
			"data":  barcodeTokenData,
		})
		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: base64.URLEncoding.EncodeToString(combinedJSON),
		}

		mockTokenStore.On("GetActiveToken", userID, 1, 1).Return("", assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to validate barcode token", response["error"])

		mockTokenStore.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - barcode token mismatch", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		barcodeTokenData := models.BarcodeData{
			EventID:     1,
			SportID:     1,
			SportName:   "Basketball",
			UserID:      userID,
			DisplayName: "Test User",
			Timestamp:   time.Now().Unix(),
			ExpiresAt:   time.Now().Unix() + 10,
		}

		combinedJSON, _ := json.Marshal(map[string]interface{}{
			"token": "stale-token",
			"data":  barcodeTokenData,
		})
		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: base64.URLEncoding.EncodeToString(combinedJSON),
		}

		mockTokenStore.On("GetActiveToken", userID, 1, 1).Return("latest-token", nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "バーコードは無効または期限切れです", response["error"])

		mockTokenStore.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - barcode token already consumed", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, mockClassRepo).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		displayName := "Test User"
		classID := 1
		teamID := 1
		barcodeTokenData := models.BarcodeData{
			EventID:     1,
			SportID:     1,
			SportName:   "Basketball",
			UserID:      userID,
			DisplayName: displayName,
			Timestamp:   time.Now().Unix(),
			ExpiresAt:   time.Now().Unix() + 10,
		}

		combinedJSON, _ := json.Marshal(map[string]interface{}{
			"token": "test-token",
			"data":  barcodeTokenData,
		})
		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: base64.URLEncoding.EncodeToString(combinedJSON),
		}

		teams := []*models.TeamWithSport{
			{
				ID:        teamID,
				Name:      "Team A",
				ClassID:   classID,
				SportID:   1,
				EventID:   1,
				SportName: "Basketball",
			},
		}
		team := &models.Team{
			ID:      teamID,
			Name:    "Team A",
			ClassID: classID,
			SportID: 1,
			EventID: 1,
		}

		mockTokenStore.On("GetActiveToken", userID, 1, 1).Return("test-token", nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", userID).Return(teams, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", classID, 1, 1).Return(team, nil).Once()
		mockTokenStore.On("ConsumeActiveToken", userID, 1, 1, "test-token").Return(false, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "バーコードは無効または期限切れです", response["error"])

		mockTeamRepo.AssertNotCalled(t, "ConfirmTeamMember")
		mockTeamRepo.AssertExpectations(t)
		mockTokenStore.AssertExpectations(t)
	})

	t.Run("Failure - User not assigned to sport", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		displayName := "Test User"

		// Create a valid barcode data
		timestamp := time.Now().Unix()
		expiresAt := timestamp + 10
		barcodeTokenData := models.BarcodeData{
			EventID:     1,
			SportID:     99, // User is not assigned to this sport
			SportName:   "Unknown Sport",
			UserID:      userID,
			DisplayName: displayName,
			Timestamp:   timestamp,
			ExpiresAt:   expiresAt,
		}

		// Encode barcode
		combinedData := map[string]interface{}{
			"token": "test-token",
			"data":  barcodeTokenData,
		}
		combinedJSON, _ := json.Marshal(combinedData)
		barcodeData := base64.URLEncoding.EncodeToString(combinedJSON)

		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: barcodeData,
		}

		teams := []*models.TeamWithSport{
			{
				ID:        1,
				Name:      "Team A",
				ClassID:   1,
				SportID:   1,
				EventID:   1,
				SportName: "Basketball",
			},
		}

		mockTeamRepo.On("GetTeamsByUserID", userID).Return(teams, nil).Once()
		mockTokenStore.On("GetActiveToken", userID, 1, 99).Return("test-token", nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "このユーザーはこの競技に参加登録されていません", response["error"])

		mockTeamRepo.AssertExpectations(t)
		mockTokenStore.AssertExpectations(t)
	})

	t.Run("Failure - Repository error", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		userID := "test-user-id"
		displayName := "Test User"

		// Create a valid barcode data
		timestamp := time.Now().Unix()
		expiresAt := timestamp + 10
		barcodeTokenData := models.BarcodeData{
			EventID:     1,
			SportID:     1,
			SportName:   "Basketball",
			UserID:      userID,
			DisplayName: displayName,
			Timestamp:   timestamp,
			ExpiresAt:   expiresAt,
		}

		// Encode barcode
		combinedData := map[string]interface{}{
			"token": "test-token",
			"data":  barcodeTokenData,
		}
		combinedJSON, _ := json.Marshal(combinedData)
		barcodeData := base64.URLEncoding.EncodeToString(combinedJSON)

		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: barcodeData,
		}

		mockTeamRepo.On("GetTeamsByUserID", userID).Return(nil, assert.AnError).Once()
		mockTokenStore.On("GetActiveToken", userID, 1, 1).Return("test-token", nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to verify user assignment", response["error"])

		mockTeamRepo.AssertExpectations(t)
		mockTokenStore.AssertExpectations(t)
	})

	t.Run("Failure - Invalid barcode data structure", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTokenStore := new(MockBarcodeTokenStore)

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository)).WithTokenStore(mockTokenStore)

		// Create invalid barcode data (malformed data field)
		invalidData := map[string]interface{}{
			"token": "test-token",
			"data":  "invalid-string-data", // Should be an object, not a string
		}
		invalidJSON, _ := json.Marshal(invalidData)
		barcodeData := base64.URLEncoding.EncodeToString(invalidJSON)

		reqBody := models.BarcodeVerifyRequest{
			BarcodeData: barcodeData,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyBarcodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		errorMsg, ok := response["error"].(string)
		assert.True(t, ok)
		assert.True(t, errorMsg == "Invalid barcode data structure" || errorMsg == "Failed to parse barcode data", "Expected error message about invalid barcode data, got: %s", errorMsg)

		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
		mockTokenStore.AssertNotCalled(t, "GetActiveToken")
	})
}
