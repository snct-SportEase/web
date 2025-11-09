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
)

func TestQRCodeHandler_GetUserTeamsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Get user teams", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

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
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/qrcode/teams", nil)

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

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/qrcode/teams", nil)

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

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

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
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/qrcode/teams", nil)

		h.GetUserTeamsHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to retrieve teams", response["error"])

		mockTeamRepo.AssertExpectations(t)
	})
}

func TestQRCodeHandler_GenerateQRCodeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Generate QR code", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

		userID := "test-user-id"
		displayName := "Test User"
		user := &models.User{
			ID:          userID,
			DisplayName: &displayName,
		}

		reqBody := models.QRCodeRequest{
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

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/qrcode/generate", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.GenerateQRCodeHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.QRCodeResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.QRCodeData)
		assert.Greater(t, response.ExpiresAt, time.Now().Unix())

		// Verify QR code data can be decoded
		decoded, err := base64.URLEncoding.DecodeString(response.QRCodeData)
		assert.NoError(t, err)

		var combinedData map[string]interface{}
		err = json.Unmarshal(decoded, &combinedData)
		assert.NoError(t, err)
		assert.Contains(t, combinedData, "token")
		assert.Contains(t, combinedData, "data")

		// Verify QR code data content
		dataBytes, _ := json.Marshal(combinedData["data"])
		var qrData models.QRCodeData
		err = json.Unmarshal(dataBytes, &qrData)
		assert.NoError(t, err)
		assert.Equal(t, reqBody.EventID, qrData.EventID)
		assert.Equal(t, reqBody.SportID, qrData.SportID)
		assert.Equal(t, "Basketball", qrData.SportName)
		assert.Equal(t, userID, qrData.UserID)
		assert.Equal(t, displayName, qrData.DisplayName)
		assert.Equal(t, response.ExpiresAt, qrData.ExpiresAt)
		assert.Equal(t, qrData.Timestamp+300, qrData.ExpiresAt) // 5 minutes

		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Failure - User not in context", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

		reqBody := models.QRCodeRequest{
			EventID: 1,
			SportID: 1,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/qrcode/generate", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.GenerateQRCodeHandler(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User not found in context", response["error"])

		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - Invalid request body", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

		userID := "test-user-id"
		displayName := "Test User"
		user := &models.User{
			ID:          userID,
			DisplayName: &displayName,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)

		c.Request, _ = http.NewRequest(http.MethodPost, "/api/qrcode/generate", bytes.NewBufferString("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		h.GenerateQRCodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request body", response["error"])

		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - User not assigned to sport", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

		userID := "test-user-id"
		displayName := "Test User"
		user := &models.User{
			ID:          userID,
			DisplayName: &displayName,
		}

		reqBody := models.QRCodeRequest{
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
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/qrcode/generate", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.GenerateQRCodeHandler(c)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "この競技に参加登録されていません", response["error"])

		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Failure - Repository error", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

		userID := "test-user-id"
		displayName := "Test User"
		user := &models.User{
			ID:          userID,
			DisplayName: &displayName,
		}

		reqBody := models.QRCodeRequest{
			EventID: 1,
			SportID: 1,
		}

		mockTeamRepo.On("GetTeamsByUserID", userID).Return(nil, assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/qrcode/generate", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.GenerateQRCodeHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to retrieve teams", response["error"])

		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Success - User without display name", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

		userID := "test-user-id"
		user := &models.User{
			ID:          userID,
			DisplayName: nil,
		}

		reqBody := models.QRCodeRequest{
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

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", user)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/qrcode/generate", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.GenerateQRCodeHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.QRCodeResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Verify display name is empty string
		decoded, _ := base64.URLEncoding.DecodeString(response.QRCodeData)
		var combinedData map[string]interface{}
		json.Unmarshal(decoded, &combinedData)
		dataBytes, _ := json.Marshal(combinedData["data"])
		var qrData models.QRCodeData
		json.Unmarshal(dataBytes, &qrData)
		assert.Equal(t, "", qrData.DisplayName)

		mockTeamRepo.AssertExpectations(t)
	})
}

func TestQRCodeHandler_VerifyQRCodeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Verify valid QR code", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

		userID := "test-user-id"
		displayName := "Test User"

		// Create a valid QR code data
		timestamp := time.Now().Unix()
		expiresAt := timestamp + 300 // 5 minutes from now
		qrData := models.QRCodeData{
			EventID:     1,
			SportID:     1,
			SportName:   "Basketball",
			UserID:      userID,
			DisplayName: displayName,
			Timestamp:   timestamp,
			ExpiresAt:   expiresAt,
		}

		// Encode QR code
		combinedData := map[string]interface{}{
			"token": "test-token",
			"data":  qrData,
		}
		combinedJSON, _ := json.Marshal(combinedData)
		qrCodeData := base64.URLEncoding.EncodeToString(combinedJSON)

		reqBody := models.QRCodeVerifyRequest{
			QRCodeData: qrCodeData,
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

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/qrcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyQRCodeHandler(c)

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
	})

	t.Run("Failure - Invalid request body", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodPost, "/api/qrcode/verify", bytes.NewBufferString("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyQRCodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request body", response["error"])

		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - Invalid QR code format", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

		reqBody := models.QRCodeVerifyRequest{
			QRCodeData: "invalid-base64-!@#$",
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/qrcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyQRCodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid QR code format", response["error"])

		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - Expired QR code", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

		userID := "test-user-id"
		displayName := "Test User"

		// Create an expired QR code data
		timestamp := time.Now().Unix() - 600 // 10 minutes ago
		expiresAt := timestamp + 300         // expired 5 minutes ago
		qrData := models.QRCodeData{
			EventID:     1,
			SportID:     1,
			SportName:   "Basketball",
			UserID:      userID,
			DisplayName: displayName,
			Timestamp:   timestamp,
			ExpiresAt:   expiresAt,
		}

		// Encode QR code
		combinedData := map[string]interface{}{
			"token": "test-token",
			"data":  qrData,
		}
		combinedJSON, _ := json.Marshal(combinedData)
		qrCodeData := base64.URLEncoding.EncodeToString(combinedJSON)

		reqBody := models.QRCodeVerifyRequest{
			QRCodeData: qrCodeData,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/qrcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyQRCodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "QRコードの有効期限が切れています", response["error"])
		assert.Equal(t, true, response["expired"])

		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - User not assigned to sport", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

		userID := "test-user-id"
		displayName := "Test User"

		// Create a valid QR code data
		timestamp := time.Now().Unix()
		expiresAt := timestamp + 300 // 5 minutes from now
		qrData := models.QRCodeData{
			EventID:     1,
			SportID:     99, // User is not assigned to this sport
			SportName:   "Unknown Sport",
			UserID:      userID,
			DisplayName: displayName,
			Timestamp:   timestamp,
			ExpiresAt:   expiresAt,
		}

		// Encode QR code
		combinedData := map[string]interface{}{
			"token": "test-token",
			"data":  qrData,
		}
		combinedJSON, _ := json.Marshal(combinedData)
		qrCodeData := base64.URLEncoding.EncodeToString(combinedJSON)

		reqBody := models.QRCodeVerifyRequest{
			QRCodeData: qrCodeData,
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

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/qrcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyQRCodeHandler(c)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "このユーザーはこの競技に参加登録されていません", response["error"])

		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Failure - Repository error", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

		userID := "test-user-id"
		displayName := "Test User"

		// Create a valid QR code data
		timestamp := time.Now().Unix()
		expiresAt := timestamp + 300 // 5 minutes from now
		qrData := models.QRCodeData{
			EventID:     1,
			SportID:     1,
			SportName:   "Basketball",
			UserID:      userID,
			DisplayName: displayName,
			Timestamp:   timestamp,
			ExpiresAt:   expiresAt,
		}

		// Encode QR code
		combinedData := map[string]interface{}{
			"token": "test-token",
			"data":  qrData,
		}
		combinedJSON, _ := json.Marshal(combinedData)
		qrCodeData := base64.URLEncoding.EncodeToString(combinedJSON)

		reqBody := models.QRCodeVerifyRequest{
			QRCodeData: qrCodeData,
		}

		mockTeamRepo.On("GetTeamsByUserID", userID).Return(nil, assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/qrcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyQRCodeHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to verify user assignment", response["error"])

		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Failure - Invalid QR code data structure", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewQRCodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo)

		// Create invalid QR code data (malformed data field)
		invalidData := map[string]interface{}{
			"token": "test-token",
			"data":  "invalid-string-data", // Should be an object, not a string
		}
		invalidJSON, _ := json.Marshal(invalidData)
		qrCodeData := base64.URLEncoding.EncodeToString(invalidJSON)

		reqBody := models.QRCodeVerifyRequest{
			QRCodeData: qrCodeData,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/qrcode/verify", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.VerifyQRCodeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		errorMsg, ok := response["error"].(string)
		assert.True(t, ok)
		assert.True(t, errorMsg == "Invalid QR code data structure" || errorMsg == "Failed to parse QR code data", "Expected error message about invalid QR code data, got: %s", errorMsg)

		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})
}
