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

func TestBarcodeHandler_CheckInRoundHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newHandler := func() (*handler.BarcodeHandler, *MockTeamRepository, *MockUserRepository) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository))
		return h, mockTeamRepo, mockUserRepo
	}

	request := func(h *handler.BarcodeHandler, body any) (*httptest.ResponseRecorder, map[string]interface{}) {
		jsonBody, err := json.Marshal(body)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/api/barcode/check-in", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.CheckInRoundHandler(c)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		return w, response
	}

	t.Run("Success - Check in round from MyID barcode", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo := newHandler()

		displayName := "山田 太郎"
		user := &models.User{ID: "user-1", Email: "2301059@example.com", DisplayName: &displayName}
		team := &models.TeamWithSport{ID: 10, Name: "1A", ClassID: 1, EventID: 1, SportID: 2, SportName: "バスケットボール"}

		mockUserRepo.On("FindUsers", "2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTeamRepo.On("CheckInRound", 10, "user-1", 1, 2, 3).Return(nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     2,
			Round:       3,
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, true, response["valid"])
		assert.Equal(t, true, response["checked_in"])
		assert.Equal(t, "2301059", response["student_number"])
		assert.Equal(t, "H102301059", response["barcode_data"])
		assert.Equal(t, "バスケットボール", response["sport_name"])
		assert.Equal(t, float64(3), response["round"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Success - Trim barcode and match exact email local part", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "2301059@example.com"}
		otherUser := &models.User{ID: "user-2", Email: "x2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, EventID: 1, SportID: 2, SportName: "バスケットボール"}

		mockUserRepo.On("FindUsers", "2301059", "email").Return([]*models.User{otherUser, user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTeamRepo.On("CheckInRound", 10, "user-1", 1, 2, 1).Return(nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "  H102301059  ",
			EventID:     1,
			SportID:     2,
			Round:       1,
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "H102301059", response["barcode_data"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Failure - Missing event or sport id", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo := newHandler()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H102301059",
			Round:       1,
		})

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "イベントIDと競技IDが必要です", response["error"])
		mockUserRepo.AssertNotCalled(t, "FindUsers")
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - Missing round", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo := newHandler()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     2,
		})

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "ラウンドを指定してください", response["error"])
		mockUserRepo.AssertNotCalled(t, "FindUsers")
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - Invalid MyID barcode format", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo := newHandler()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "2301059",
			EventID:     1,
			SportID:     2,
			Round:       1,
		})

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "バーコード形式が不正です", response["error"])
		mockUserRepo.AssertNotCalled(t, "FindUsers")
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - User not found by exact email local part", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo := newHandler()

		mockUserRepo.On("FindUsers", "2301059", "email").
			Return([]*models.User{{ID: "user-2", Email: "x2301059@example.com"}}, nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     2,
			Round:       1,
		})

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "該当する学生が見つかりません", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - User lookup error", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo := newHandler()

		mockUserRepo.On("FindUsers", "2301059", "email").Return(nil, assert.AnError).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     2,
			Round:       1,
		})

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Failed to find user by student number", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - Student is not pre-entered for sport", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "2301059@example.com"}
		mockUserRepo.On("FindUsers", "2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").
			Return([]*models.TeamWithSport{{ID: 10, EventID: 1, SportID: 9}}, nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     2,
			Round:       1,
		})

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, "このユーザーはこの競技に事前エントリーされていません", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "CheckInRound")
	})

	t.Run("Failure - Check-in repository error", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTeamRepo.On("CheckInRound", 10, "user-1", 1, 2, 1).Return(assert.AnError).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     2,
			Round:       1,
		})

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Failed to check in round", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
	})
}
