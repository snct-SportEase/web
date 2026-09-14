package handler_test

import (
	"backapp/internal/handler"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateMatchResultRejectsMatchFromInactiveEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tournamentRepo := new(MockTournamentRepository)
	h := handler.NewTournamentHandler(tournamentRepo, nil, nil, nil, nil, nil)

	tournamentRepo.On("GetEventIDByMatchID", 42).Return(7, nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "match_id", Value: "42"}}
	c.Set("active_event_id", 8)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/matches/42/result", bytes.NewBufferString(`{"team1_score":2,"team2_score":1,"winner_id":10}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateMatchResultHandler(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	tournamentRepo.AssertNotCalled(t, "UpdateMatchResult")
	tournamentRepo.AssertNotCalled(t, "UpdateMatchResultForCorrection")
	tournamentRepo.AssertExpectations(t)
}

func TestUpdateMatchResultRejectsNegativeScores(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tournamentRepo := new(MockTournamentRepository)
	h := handler.NewTournamentHandler(tournamentRepo, nil, nil, nil, nil, nil)

	tournamentRepo.On("GetEventIDByMatchID", 42).Return(8, nil).Once()
	tournamentRepo.On("IsMatchResultAlreadyEntered", 42).Return(false, nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "match_id", Value: "42"}}
	c.Set("active_event_id", 8)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/matches/42/result", bytes.NewBufferString(`{"team1_score":-1,"team2_score":0}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateMatchResultHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	tournamentRepo.AssertNotCalled(t, "UpdateMatchResult")
}

func TestTournamentHandler_UpdateMatchStartTimes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("updates the regular start time", func(t *testing.T) {
		tournamentRepo := new(MockTournamentRepository)
		h := handler.NewTournamentHandler(tournamentRepo, nil, nil, nil, nil, nil)
		tournamentRepo.On("UpdateMatchStartTime", 42, "2025-04-01 09:30:00").Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "match_id", Value: "42"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/matches/42/start-time", bytes.NewBufferString(`{"start_time":"2025-04-01 09:30:00"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		h.UpdateMatchStartTimeHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		tournamentRepo.AssertExpectations(t)
	})

	t.Run("updates the rainy-mode start time", func(t *testing.T) {
		tournamentRepo := new(MockTournamentRepository)
		h := handler.NewTournamentHandler(tournamentRepo, nil, nil, nil, nil, nil)
		tournamentRepo.On("UpdateMatchRainyModeStartTime", 42, "2025-04-01 11:00:00").Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "match_id", Value: "42"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/matches/42/rainy-mode-start-time", bytes.NewBufferString(`{"rainy_mode_start_time":"2025-04-01 11:00:00"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		h.UpdateMatchRainyModeStartTimeHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		tournamentRepo.AssertExpectations(t)
	})

	t.Run("rejects an invalid match id", func(t *testing.T) {
		tournamentRepo := new(MockTournamentRepository)
		h := handler.NewTournamentHandler(tournamentRepo, nil, nil, nil, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "match_id", Value: "invalid"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/matches/invalid/start-time", bytes.NewBufferString(`{"start_time":"2025-04-01 09:30:00"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		h.UpdateMatchStartTimeHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		tournamentRepo.AssertNotCalled(t, "UpdateMatchStartTime", mock.Anything, mock.Anything)
	})
}
