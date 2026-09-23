package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"backapp/internal/repository"
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These cases cover the boundary between pre-entry, match selection and actual registration.
func TestMyIDEntryBoundaries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	makeHandler := func() (*handler.BarcodeHandler, *MockTeamRepository, *MockUserRepository, *MockTournamentRepository, *MockClassRepository) {
		teams := new(MockTeamRepository)
		users := new(MockUserRepository)
		matches := new(MockTournamentRepository)
		classes := new(MockClassRepository)
		h := handler.NewBarcodeHandler(teams, new(MockSportRepository), users, new(MockEventRepository), classes, matches)
		return h, teams, users, matches, classes
	}
	request := func(h *handler.BarcodeHandler, body any) (*httptest.ResponseRecorder, map[string]any) {
		payload, err := json.Marshal(body)
		require.NoError(t, err)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/barcode/check-in", bytes.NewReader(payload))
		c.Request.Header.Set("Content-Type", "application/json")
		h.CheckInRoundHandler(c)
		var response map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		return w, response
	}
	base := models.BarcodeCheckInRequest{BarcodeData: "H1023010590", EventID: 1, SportID: 2, MatchID: 100}

	t.Run("entry in another event cannot be confirmed", func(t *testing.T) {
		h, teams, users, matches, _ := makeHandler()
		users.On("FindUsers", "s2301059", "email").Return([]*models.User{{ID: "student", Email: "s2301059@example.com"}}, nil).Once()
		teams.On("GetTeamsByUserID", "student").Return([]*models.TeamWithSport{{ID: 10, EventID: 9, SportID: 2}}, nil).Once()
		w, response := request(h, base)
		require.Equal(t, http.StatusForbidden, w.Code)
		require.Equal(t, "このユーザーはこの競技に事前エントリーされていません", response["error"])
		users.AssertExpectations(t)
		teams.AssertExpectations(t)
		matches.AssertNotCalled(t, "GetMatchForEventSport")
		teams.AssertNotCalled(t, "ConfirmTeamMember")
		teams.AssertNotCalled(t, "CheckInRound")
	})

	t.Run("lookup failure cannot confirm", func(t *testing.T) {
		h, teams, users, matches, _ := makeHandler()
		users.On("FindUsers", "s2301059", "email").Return([]*models.User{{ID: "student", Email: "s2301059@example.com"}}, nil).Once()
		teams.On("GetTeamsByUserID", "student").Return(nil, assert.AnError).Once()
		w, _ := request(h, base)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		users.AssertExpectations(t)
		teams.AssertExpectations(t)
		matches.AssertNotCalled(t, "GetMatchForEventSport")
		teams.AssertNotCalled(t, "ConfirmTeamMember")
	})

	t.Run("duplicate selected match is processed once", func(t *testing.T) {
		h, teams, users, matches, _ := makeHandler()
		users.On("FindUsers", "s2301059", "email").Return([]*models.User{{ID: "student", Email: "s2301059@example.com"}}, nil).Once()
		teams.On("GetTeamsByUserID", "student").Return([]*models.TeamWithSport{{ID: 10, ClassID: 1, EventID: 1, SportID: 2}}, nil).Once()
		matches.On("GetMatchForEventSport", 100, 1, 2).Return(&models.MatchDB{ID: 100, Round: 0, Team1ID: sql.NullInt64{Int64: 10, Valid: true}}, nil).Once()
		teams.On("GetTeamByClassAndSport", 1, 2, 1).Return(&models.Team{ID: 10}, nil).Once()
		teams.On("ConfirmTeamMember", 10, "student").Return(nil).Once()
		teams.On("CheckInRound", 10, "student", 1, 2, 100, 1).Return(nil).Once()
		body := base
		body.MatchIDs = []int{100, 0, -1, 100}
		w, response := request(h, body)
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, []any{float64(100)}, response["match_ids"])
		users.AssertExpectations(t)
		teams.AssertExpectations(t)
		matches.AssertExpectations(t)
	})

	t.Run("repeated scan reports conflict and keeps one check-in", func(t *testing.T) {
		h, teams, users, matches, _ := makeHandler()
		users.On("FindUsers", "s2301059", "email").Return([]*models.User{{ID: "student", Email: "s2301059@example.com"}}, nil).Twice()
		teams.On("GetTeamsByUserID", "student").Return([]*models.TeamWithSport{{ID: 10, ClassID: 1, EventID: 1, SportID: 2}}, nil).Twice()
		matches.On("GetMatchForEventSport", 100, 1, 2).Return(&models.MatchDB{ID: 100, Round: 0, Team1ID: sql.NullInt64{Int64: 10, Valid: true}}, nil).Twice()
		teams.On("GetTeamByClassAndSport", 1, 2, 1).Return(&models.Team{ID: 10}, nil).Twice()
		teams.On("ConfirmTeamMember", 10, "student").Return(nil).Twice()
		teams.On("CheckInRound", 10, "student", 1, 2, 100, 1).Return(nil).Once()
		teams.On("CheckInRound", 10, "student", 1, 2, 100, 1).Return(repository.ErrRoundAlreadyCheckedIn).Once()
		first, _ := request(h, base)
		second, response := request(h, base)
		require.Equal(t, http.StatusOK, first.Code)
		require.Equal(t, http.StatusConflict, second.Code)
		require.Equal(t, true, response["already_checked_in"])
		users.AssertExpectations(t)
		teams.AssertExpectations(t)
		matches.AssertExpectations(t)
	})

	for _, tc := range []struct {
		name string
		count int
		warning bool
	}{
		{"below minimum", 2, true},
		{"at minimum", 3, false},
		{"above minimum", 4, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			min := 3
			h, teams, users, matches, classes := makeHandler()
			users.On("FindUsers", "s2301059", "email").Return([]*models.User{{ID: "student", Email: "s2301059@example.com"}}, nil).Once()
			teams.On("GetTeamsByUserID", "student").Return([]*models.TeamWithSport{{ID: 10, ClassID: 1, EventID: 1, SportID: 2, SportName: "バスケットボール"}}, nil).Once()
			matches.On("GetMatchForEventSport", 100, 1, 2).Return(&models.MatchDB{ID: 100, Round: 0, Team1ID: sql.NullInt64{Int64: 10, Valid: true}}, nil).Once()
			teams.On("GetTeamByClassAndSport", 1, 2, 1).Return(&models.Team{ID: 10, MinCapacity: &min}, nil).Once()
			teams.On("ConfirmTeamMember", 10, "student").Return(nil).Once()
			teams.On("CheckInRound", 10, "student", 1, 2, 100, 1).Return(nil).Once()
			teams.On("GetConfirmedTeamMembersCount", 10).Return(tc.count, nil).Once()
			if tc.warning {
				classes.On("GetClassByID", 1).Return(&models.Class{Name: "1A"}, nil).Once()
			}
			w, response := request(h, base)
			require.Equal(t, http.StatusOK, w.Code)
			_, hasWarning := response["capacity_warning"]
			require.Equal(t, tc.warning, hasWarning)
			users.AssertExpectations(t)
			teams.AssertExpectations(t)
			matches.AssertExpectations(t)
			classes.AssertExpectations(t)
		})
	}
}
