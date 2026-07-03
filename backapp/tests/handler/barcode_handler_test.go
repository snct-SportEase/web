package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"bytes"
	"database/sql"
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

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository), new(MockTournamentRepository))

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

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository), new(MockTournamentRepository))

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

		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository), new(MockTournamentRepository))

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

	newHandler := func() (*handler.BarcodeHandler, *MockTeamRepository, *MockUserRepository, *MockTournamentRepository) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository), mockTournamentRepo)
		return h, mockTeamRepo, mockUserRepo, mockTournamentRepo
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
		h, mockTeamRepo, mockUserRepo, mockTournamentRepo := newHandler()

		displayName := "山田 太郎"
		user := &models.User{ID: "user-1", Email: "s2301059@example.com", DisplayName: &displayName}
		team := &models.TeamWithSport{ID: 10, Name: "1A", ClassID: 1, EventID: 1, SportID: 2, SportName: "バスケットボール"}
		teamDetails := &models.Team{ID: 10, ClassID: 1, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTournamentRepo.On("GetMatchForEventSport", 100, 1, 2).Return(&models.MatchDB{ID: 100, Round: 2, Team1ID: sql.NullInt64{Int64: 10, Valid: true}, Team2ID: sql.NullInt64{Int64: 20, Valid: true}}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", 1, 2, 1).Return(teamDetails, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", 10, "user-1").Return(nil).Once()
		mockTeamRepo.On("CheckInRound", 10, "user-1", 1, 2, 100, 3).Return(nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
			MatchID:     100,
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, true, response["valid"])
		assert.Equal(t, true, response["confirmed"])
		assert.Equal(t, true, response["checked_in"])
		assert.Equal(t, "2301059", response["student_number"])
		assert.Equal(t, "H1023010590", response["barcode_data"])
		assert.Equal(t, "バスケットボール", response["sport_name"])
		assert.Equal(t, float64(3), response["round"])
		assert.Equal(t, float64(100), response["match_id"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})

	t.Run("Success - Trim barcode and match exact s-prefixed email local part", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, mockTournamentRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		otherUser := &models.User{ID: "user-2", Email: "xs2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, EventID: 1, SportID: 2, SportName: "バスケットボール"}
		teamDetails := &models.Team{ID: 10, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{otherUser, user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTournamentRepo.On("GetMatchForEventSport", 101, 1, 2).Return(&models.MatchDB{ID: 101, Round: 0, Team2ID: sql.NullInt64{Int64: 10, Valid: true}}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", 0, 2, 1).Return(teamDetails, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", 10, "user-1").Return(nil).Once()
		mockTeamRepo.On("CheckInRound", 10, "user-1", 1, 2, 101, 1).Return(nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "  H1023010590  ",
			EventID:     1,
			SportID:     2,
			MatchID:     101,
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "H1023010590", response["barcode_data"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})

	t.Run("Success - Accept Code 39 start and stop asterisks", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, mockTournamentRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, ClassID: 1, EventID: 1, SportID: 2, SportName: "バスケットボール"}
		teamDetails := &models.Team{ID: 10, ClassID: 1, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTournamentRepo.On("GetMatchForEventSport", 101, 1, 2).Return(&models.MatchDB{ID: 101, Round: 0, Team1ID: sql.NullInt64{Int64: 10, Valid: true}}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", 1, 2, 1).Return(teamDetails, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", 10, "user-1").Return(nil).Once()
		mockTeamRepo.On("CheckInRound", 10, "user-1", 1, 2, 101, 1).Return(nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "*h1023010590*",
			EventID:     1,
			SportID:     2,
			MatchID:     101,
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "2301059", response["student_number"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})

	t.Run("Success - Ignore Code 39 trailing check digit", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, mockTournamentRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, ClassID: 1, EventID: 1, SportID: 2, SportName: "バスケットボール"}
		teamDetails := &models.Team{ID: 10, ClassID: 1, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTournamentRepo.On("GetMatchForEventSport", 101, 1, 2).Return(&models.MatchDB{ID: 101, Round: 0, Team1ID: sql.NullInt64{Int64: 10, Valid: true}}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", 1, 2, 1).Return(teamDetails, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", 10, "user-1").Return(nil).Once()
		mockTeamRepo.On("CheckInRound", 10, "user-1", 1, 2, 101, 1).Return(nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
			MatchID:     101,
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "2301059", response["student_number"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})

	t.Run("Success - Check in matching team from selected time slot", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, mockTournamentRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		team := &models.TeamWithSport{ID: 30, ClassID: 3, EventID: 1, SportID: 2, SportName: "バスケットボール"}
		teamDetails := &models.Team{ID: 30, ClassID: 3, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTournamentRepo.On("GetMatchForEventSport", 101, 1, 2).Return(&models.MatchDB{ID: 101, Round: 0, Team1ID: sql.NullInt64{Int64: 10, Valid: true}, Team2ID: sql.NullInt64{Int64: 20, Valid: true}}, nil).Once()
		mockTournamentRepo.On("GetMatchForEventSport", 102, 1, 2).Return(&models.MatchDB{ID: 102, Round: 0, Team1ID: sql.NullInt64{Int64: 30, Valid: true}, Team2ID: sql.NullInt64{Int64: 40, Valid: true}}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", 3, 2, 1).Return(teamDetails, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", 30, "user-1").Return(nil).Once()
		mockTeamRepo.On("CheckInRound", 30, "user-1", 1, 2, 102, 1).Return(nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
			MatchID:     101,
			MatchIDs:    []int{101, 102},
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, float64(102), response["match_id"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})

	t.Run("Failure - Missing event or sport id", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, _ := newHandler()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			MatchID:     101,
		})

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "イベントIDと競技IDが必要です", response["error"])
		mockUserRepo.AssertNotCalled(t, "FindUsers")
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - Missing match", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, _ := newHandler()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
		})

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "試合を選択してください", response["error"])
		mockUserRepo.AssertNotCalled(t, "FindUsers")
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - Invalid MyID barcode format", func(t *testing.T) {
		tests := []struct {
			name        string
			barcodeData string
		}{
			{name: "missing check digit", barcodeData: "H102301059"},
			{name: "student number only", barcodeData: "2301059"},
			{name: "student number only with check digit", barcodeData: "23010590"},
			{name: "student number is too short", barcodeData: "H10230105"},
			{name: "student number is too long", barcodeData: "H10230105999"},
			{name: "student number contains non-digit", barcodeData: "H10230105A"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				h, mockTeamRepo, mockUserRepo, _ := newHandler()

				w, response := request(h, models.BarcodeCheckInRequest{
					BarcodeData: tt.barcodeData,
					EventID:     1,
					SportID:     2,
					MatchID:     101,
				})

				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Equal(t, "バーコード形式が不正です", response["error"])
				mockUserRepo.AssertNotCalled(t, "FindUsers")
				mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
			})
		}
	})

	t.Run("Failure - User not found by exact s-prefixed email local part", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, _ := newHandler()

		mockUserRepo.On("FindUsers", "s2301059", "email").
			Return([]*models.User{{ID: "user-2", Email: "xs2301059@example.com"}}, nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
			MatchID:     101,
		})

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "該当する学生が見つかりません", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - Email local part without s prefix is not accepted", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, _ := newHandler()

		mockUserRepo.On("FindUsers", "s2301059", "email").
			Return([]*models.User{{ID: "user-2", Email: "2301059@example.com"}}, nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
			MatchID:     101,
		})

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "該当する学生が見つかりません", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - User lookup error", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, _ := newHandler()

		mockUserRepo.On("FindUsers", "s2301059", "email").Return(nil, assert.AnError).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
			MatchID:     101,
		})

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Failed to find user by student number", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID")
	})

	t.Run("Failure - Student is not pre-entered for sport", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, mockTournamentRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").
			Return([]*models.TeamWithSport{{ID: 10, EventID: 1, SportID: 9}}, nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
			MatchID:     101,
		})

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, "このユーザーはこの競技に事前エントリーされていません", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertNotCalled(t, "GetMatchForEventSport")
		mockTeamRepo.AssertNotCalled(t, "CheckInRound")
		mockTeamRepo.AssertNotCalled(t, "ConfirmTeamMember")
	})

	t.Run("Failure - Selected match does not belong to event sport", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, mockTournamentRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTournamentRepo.On("GetMatchForEventSport", 999, 1, 2).Return(nil, nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
			MatchID:     999,
		})

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "選択した試合がこの競技に存在しません", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "ConfirmTeamMember")
		mockTeamRepo.AssertNotCalled(t, "CheckInRound")
	})

	t.Run("Failure - Match verification error", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, mockTournamentRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTournamentRepo.On("GetMatchForEventSport", 101, 1, 2).Return(nil, assert.AnError).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
			MatchID:     101,
		})

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Failed to verify selected match", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "ConfirmTeamMember")
		mockTeamRepo.AssertNotCalled(t, "CheckInRound")
	})

	t.Run("Failure - Student team is not in selected match", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, mockTournamentRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTournamentRepo.On("GetMatchForEventSport", 101, 1, 2).Return(&models.MatchDB{
			ID:      101,
			Round:   0,
			Team1ID: sql.NullInt64{Int64: 20, Valid: true},
			Team2ID: sql.NullInt64{Int64: 30, Valid: true},
		}, nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
			MatchID:     101,
		})

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, "まだあなたのクラスはこの試合にチェックインできません", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "GetTeamByClassAndSport")
		mockTeamRepo.AssertNotCalled(t, "ConfirmTeamMember")
		mockTeamRepo.AssertNotCalled(t, "CheckInRound")
	})

	t.Run("Failure - Selected match has no assigned teams", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, mockTournamentRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTournamentRepo.On("GetMatchForEventSport", 101, 1, 2).Return(&models.MatchDB{ID: 101, Round: 0}, nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
			MatchID:     101,
		})

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, "まだあなたのクラスはこの試合にチェックインできません", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "GetTeamByClassAndSport")
		mockTeamRepo.AssertNotCalled(t, "ConfirmTeamMember")
		mockTeamRepo.AssertNotCalled(t, "CheckInRound")
	})

	t.Run("Failure - Confirm team member error", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, mockTournamentRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, EventID: 1, SportID: 2}
		teamDetails := &models.Team{ID: 10, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTournamentRepo.On("GetMatchForEventSport", 101, 1, 2).Return(&models.MatchDB{ID: 101, Round: 0, Team1ID: sql.NullInt64{Int64: 10, Valid: true}}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", 0, 2, 1).Return(teamDetails, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", 10, "user-1").Return(assert.AnError).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
			MatchID:     101,
		})

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Failed to confirm team member", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "CheckInRound")
	})

	t.Run("Failure - Check-in repository error", func(t *testing.T) {
		h, mockTeamRepo, mockUserRepo, mockTournamentRepo := newHandler()

		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, EventID: 1, SportID: 2}
		teamDetails := &models.Team{ID: 10, EventID: 1, SportID: 2}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTournamentRepo.On("GetMatchForEventSport", 101, 1, 2).Return(&models.MatchDB{ID: 101, Round: 0, Team1ID: sql.NullInt64{Int64: 10, Valid: true}}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", 0, 2, 1).Return(teamDetails, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", 10, "user-1").Return(nil).Once()
		mockTeamRepo.On("CheckInRound", 10, "user-1", 1, 2, 101, 1).Return(assert.AnError).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
			MatchID:     101,
		})

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Failed to check in round", response["error"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})

	t.Run("Success - Returns capacity warning after confirming team member", func(t *testing.T) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockClassRepo := new(MockClassRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, mockClassRepo, mockTournamentRepo)

		minCapacity := 3
		user := &models.User{ID: "user-1", Email: "s2301059@example.com"}
		team := &models.TeamWithSport{ID: 10, Name: "1A", ClassID: 1, EventID: 1, SportID: 2, SportName: "バスケットボール"}
		teamDetails := &models.Team{ID: 10, ClassID: 1, EventID: 1, SportID: 2, MinCapacity: &minCapacity}

		mockUserRepo.On("FindUsers", "s2301059", "email").Return([]*models.User{user}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user-1").Return([]*models.TeamWithSport{team}, nil).Once()
		mockTournamentRepo.On("GetMatchForEventSport", 101, 1, 2).Return(&models.MatchDB{ID: 101, Round: 0, Team1ID: sql.NullInt64{Int64: 10, Valid: true}}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", 1, 2, 1).Return(teamDetails, nil).Once()
		mockTeamRepo.On("ConfirmTeamMember", 10, "user-1").Return(nil).Once()
		mockTeamRepo.On("CheckInRound", 10, "user-1", 1, 2, 101, 1).Return(nil).Once()
		mockTeamRepo.On("GetConfirmedTeamMembersCount", 10).Return(2, nil).Once()
		mockClassRepo.On("GetClassByID", 1).Return(&models.Class{Name: "1A"}, nil).Once()

		w, response := request(h, models.BarcodeCheckInRequest{
			BarcodeData: "H1023010590",
			EventID:     1,
			SportID:     2,
			MatchID:     101,
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, true, response["confirmed"])
		assert.Equal(t, "1Aクラスのバスケットボールの参加本登録済みメンバー数（2人）が最低人数（3人）に達していません", response["capacity_warning"])
		mockUserRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})
}

func TestBarcodeHandler_GetMatchCheckInsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newHandler := func() (*handler.BarcodeHandler, *MockTeamRepository, *MockTournamentRepository) {
		mockTeamRepo := new(MockTeamRepository)
		mockSportRepo := new(MockSportRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		h := handler.NewBarcodeHandler(mockTeamRepo, mockSportRepo, mockUserRepo, mockEventRepo, new(MockClassRepository), mockTournamentRepo)
		return h, mockTeamRepo, mockTournamentRepo
	}

	request := func(h *handler.BarcodeHandler, matchID string, query string) (*httptest.ResponseRecorder, map[string]interface{}) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "match_id", Value: matchID}}
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/barcode/matches/"+matchID+"/check-ins?"+query, nil)

		h.GetMatchCheckInsHandler(c)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		return w, response
	}

	t.Run("Success - Get match check-ins", func(t *testing.T) {
		h, mockTeamRepo, mockTournamentRepo := newHandler()

		displayName := "山田太郎"
		members := []*models.MatchCheckInMember{
			{
				UserID:      "user-1",
				Email:       "s2301059@sendai-nct.jp",
				DisplayName: &displayName,
				ClassID:     1,
				ClassName:   "1A",
				TeamID:      10,
				TeamName:    "1A",
				EventID:     1,
				SportID:     2,
				MatchID:     100,
				Round:       3,
			},
		}

		mockTournamentRepo.On("GetMatchForEventSport", 100, 1, 2).Return(&models.MatchDB{ID: 100, Round: 2}, nil).Once()
		mockTeamRepo.On("GetMatchCheckIns", 1, 2, 100).Return(members, nil).Once()

		w, response := request(h, "100", "event_id=1&sport_id=2")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, float64(1), response["count"])
		responseMembers, ok := response["members"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, responseMembers, 1)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})

	t.Run("Success - Get check-ins for selected time slot", func(t *testing.T) {
		h, mockTeamRepo, mockTournamentRepo := newHandler()

		firstName := "山田太郎"
		secondName := "佐藤花子"
		firstMembers := []*models.MatchCheckInMember{
			{
				UserID:      "user-1",
				DisplayName: &firstName,
				MatchID:     100,
			},
		}
		secondMembers := []*models.MatchCheckInMember{
			{
				UserID:      "user-2",
				DisplayName: &secondName,
				MatchID:     101,
			},
		}

		mockTournamentRepo.On("GetMatchForEventSport", 100, 1, 2).Return(&models.MatchDB{ID: 100, Round: 0}, nil).Once()
		mockTournamentRepo.On("GetMatchForEventSport", 101, 1, 2).Return(&models.MatchDB{ID: 101, Round: 0}, nil).Once()
		mockTeamRepo.On("GetMatchCheckIns", 1, 2, 100).Return(firstMembers, nil).Once()
		mockTeamRepo.On("GetMatchCheckIns", 1, 2, 101).Return(secondMembers, nil).Once()

		w, response := request(h, "100", "event_id=1&sport_id=2&match_ids=100,101")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, float64(2), response["count"])
		responseMembers, ok := response["members"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, responseMembers, 2)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})

	t.Run("Failure - Selected match does not belong to event sport", func(t *testing.T) {
		h, mockTeamRepo, mockTournamentRepo := newHandler()

		mockTournamentRepo.On("GetMatchForEventSport", 100, 1, 2).Return(nil, nil).Once()

		w, response := request(h, "100", "event_id=1&sport_id=2")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "選択した試合がこの競技に存在しません", response["error"])
		mockTeamRepo.AssertNotCalled(t, "GetMatchCheckIns")
		mockTournamentRepo.AssertExpectations(t)
	})

	t.Run("Failure - Repository error", func(t *testing.T) {
		h, mockTeamRepo, mockTournamentRepo := newHandler()

		mockTournamentRepo.On("GetMatchForEventSport", 100, 1, 2).Return(&models.MatchDB{ID: 100, Round: 2}, nil).Once()
		mockTeamRepo.On("GetMatchCheckIns", 1, 2, 100).Return(nil, assert.AnError).Once()

		w, response := request(h, "100", "event_id=1&sport_id=2")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Failed to get match check-ins", response["error"])
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})
}
