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
		user := &models.User{ID: "user-1", Email: "s2301059@example.com", DisplayName: &displayName}
		team := &models.TeamWithSport{ID: 10, Name: "1A", ClassID: 1, EventID: 1, SportID: 2, SportName: "バスケットボール"}
		teamDetails := &models.Team{ID: 10, ClassID: 1, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", 1, 2, 1).Return(teamDetails, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", 10, "user-1").Return(nil).Once()
		mockTeamRepo.On("CheckInRound", 10, "user-1", 1, 2, 3).Return(nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     2,
			Round:       3,
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, true, response["valid"])
		assert.Equal(t, true, response["confirmed"])
		assert.Equal(t, true, response["checked_in"])
		assert.Equal(t, "2301059", response["student_number"])
		assert.Equal(t, "H102301059", response["barcode_data"])
		assert.Equal(t, "バスケットボール", response["sport_name"])
		assert.Equal(t, float64(3), response["round"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("Success - Trim barcode and match exact s-prefixed email local part", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		otherUser := &models.User{ID: "user-2", Email: "xs2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, EventID: 1, SportID: 2, SportName: "バスケットボール"}
		teamDetails := &models.Team{ID: 10, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{otherUser, user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", 0, 2, 1).Return(teamDetails, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", 10, "user-1").Return(nil).Once()
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
		tests := []struct {
			name        string
			barcodeData string
		}{
			{name: "missing H10 prefix", barcodeData: "2301059"},
			{name: "student number is too short", barcodeData: "H10230105"},
			{name: "student number is too long", barcodeData: "H1023010599"},
			{name: "student number contains non-digit", barcodeData: "H10230105A"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				h, mockTeamRepo, mockUserRepo := newHandler()

				w, response := request(h, models.BarcodeCheckInRequest{
					BarcodeData: tt.barcodeData,
					EventID:     1,
					SportID:     2,
					Round:       1,
				})

				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Equal(t, "バーコード形式が不正です", response["error"])
				mockUserRepo.AssertNotCalled(t, "FindUsers")
				mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
			})
		}
	})

	t.Run("Failure - User not found by exact s-prefixed email local part", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo := newHandler()

		mockUserRepo.On("FindUsers", "s2301059", "email").
			Return([]*models.User{{ID: "user-2", Email: "xs2301059@example.com"}}, nil).Once()

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

	t.Run("Failure - Email local part without s prefix is not accepted", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo := newHandler()

		mockUserRepo.On("FindUsers", "s2301059", "email").
			Return([]*models.User{{ID: "user-2", Email: "2301059@example.com"}}, nil).Once()

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

		mockUserRepo.On("FindUsers", "s2301059", "email").Return(nil, assert.AnError).Once()

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

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
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
		mockTeamRepo.AssertNotCalled(t, "ConfirmTeamMember")
	})

	t.Run("Failure - Confirm team member error", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, EventID: 1, SportID: 2}
		teamDetails := &models.Team{ID: 10, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", 0, 2, 1).Return(teamDetails, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", 10, "user-1").Return(assert.AnError).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     2,
			Round:       1,
		})

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Failed to confirm team member", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "CheckInRound")
	})

	t.Run("Failure - Check-in repository error", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, EventID: 1, SportID: 2}
		teamDetails := &models.Team{ID: 10, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", 0, 2, 1).Return(teamDetails, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", 10, "user-1").Return(nil).Once()
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

	t.Run("Success - Returns capacity warning after confirming team member", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, mockClassRepo)

		minCapacity := 3
		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, Name: "1A", ClassID: 1, EventID: 1, SportID: 2, SportName: "バスケットボール"}
		teamDetails := &models.Team{ID: 10, ClassID: 1, EventID: 1, SportID: 2, MinCapacity: &minCapacity}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", 1, 2, 1).Return(teamDetails, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", 10, "user-1").Return(nil).Once()
		mockTeamRepo.On("CheckInRound", 10, "user-1", 1, 2, 1).Return(nil).Once()
		mockTeamRepo.On("GetConfirmedTeamMembersCount", 10).Return(2, nil).Once()
		mockClassRepo.On("GetClassByID", 1).Return(&models.Class{Name: "1A"}, nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H102301059",
			EventID:     1,
			SportID:     2,
			Round:       1,
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, true, response["confirmed"])
		assert.Equal(t, "1Aクラスのバスケットボールの参加本登録済みメンバー数（2人）が最低人数（3人）に達していません", response["capacity_warning"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
	})
}
