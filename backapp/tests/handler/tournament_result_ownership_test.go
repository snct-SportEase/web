package handler_test

import (
	"backapp/internal/handler"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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
