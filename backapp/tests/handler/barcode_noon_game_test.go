package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBarcodeHandler_NoonGameCheckInsOnlyIncludeAssignedMembers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	classID := 1
	assignedName := "割当済み学生"
	assignedMember := &models.User{ID: "assigned-user", Email: "assigned@example.com", DisplayName: &assignedName, ClassID: &classID}
	match := &models.NoonGameMatchWithResult{
		NoonGameMatch: &models.NoonGameMatch{ID: 10, SessionID: 1},
		Entries:       []*models.NoonGameMatchEntry{{MatchID: 10, ClassID: &classID}},
	}

	teamRepo := new(MockTeamRepository)
	classRepo := new(MockClassRepository)
	noonRepo := new(MockNoonGameRepository)
	teamRepo.On("GetTeamByClassAndSport", 1, 2, 1).Return(&models.Team{ID: 20, Name: "1A", ClassID: 1, SportID: 2, EventID: 1}, nil).Once()
	teamRepo.On("GetTeamMembers", 20).Return([]*models.User{assignedMember}, nil).Once()
	classRepo.On("GetClassByID", 1).Return(&models.Class{ID: 1, Name: "1A"}, nil).Once()
	noonRepo.On("GetSessionByID", 1).Return(&models.NoonGameSession{ID: 1, EventID: 1, Name: "昼競技"}, nil).Once()
	noonRepo.On("GetMatchByID", 10).Return(match, nil).Once()
	noonRepo.On("GetMatchCheckIns", 1, 1, 10).Return([]*models.NoonGameCheckIn{}, nil).Once()

	h := handler.NewBarcodeHandler(teamRepo, new(MockSportRepository), new(MockUserRepository), new(MockEventRepository), classRepo, new(MockTournamentRepository)).WithNoonGameRepository(noonRepo)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/barcode/matches/10/check-ins?event_id=1&sport_id=2&noon_game_session_id=1", nil)
	c.Params = gin.Params{{Key: "match_id", Value: "10"}}

	h.GetMatchCheckInsHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response struct {
		CheckedInMembers []*models.MatchCheckInMember `json:"checked_in_members"`
		UncheckedMembers []*models.MatchCheckInMember `json:"unchecked_members"`
	}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Empty(t, response.CheckedInMembers)
	if assert.Len(t, response.UncheckedMembers, 1) {
		assert.Equal(t, "assigned-user", response.UncheckedMembers[0].UserID)
		assert.Equal(t, "割当済み学生", *response.UncheckedMembers[0].DisplayName)
	}
	classRepo.AssertNotCalled(t, "GetClassMembers", 1)
	teamRepo.AssertExpectations(t)
	classRepo.AssertExpectations(t)
	noonRepo.AssertExpectations(t)
}
