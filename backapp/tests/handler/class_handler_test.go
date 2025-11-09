package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClassHandler_GetClassScores(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewClassHandler(mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo)

		expectedScores := []*models.ClassScore{
			{ID: 1, ClassName: "Class A", Season: "spring", TotalPointsCurrentEvent: 100, RankCurrentEvent: 1},
			{ID: 2, ClassName: "Class B", Season: "spring", TotalPointsCurrentEvent: 90, RankCurrentEvent: 2},
		}

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassScoresByEvent", 1).Return(expectedScores, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/scores/class", nil)

		h.GetClassScores(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var actualScores []*models.ClassScore
		err := json.Unmarshal(w.Body.Bytes(), &actualScores)
		assert.NoError(t, err)
		assert.Equal(t, expectedScores, actualScores)

		mockClassRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("No active event", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewClassHandler(mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo)

		mockEventRepo.On("GetActiveEvent").Return(0, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/scores/class", nil)

		h.GetClassScores(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertNotCalled(t, "GetClassScoresByEvent", mock.Anything)
	})
}

func TestClassHandler_GetClassProgress(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockClassRepo := new(MockClassRepository)
	mockEventRepo := new(MockEventRepository)
	mockTeamRepo := new(MockTeamRepository)
	mockTournamentRepo := new(MockTournamentRepository)

	h := handler.NewClassHandler(mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo)

	user := &models.User{ID: "user-1"}
	class := &models.Class{ID: 10, Name: "IS3"}
	team := &models.TeamWithSport{ID: 100, Name: "IS3-A", SportID: 1, SportName: "バスケットボール"}
	displayNameA := "User A"
	displayNameB := "User B"
	members := []*models.User{
		{ID: "user-a", Email: "a@example.com", DisplayName: &displayNameA},
		{ID: "user-b", Email: "b@example.com", DisplayName: &displayNameB},
	}

	matchDetails := []*models.MatchDetail{
		{
			MatchID:        1,
			TournamentID:   5,
			TournamentName: "バスケットボール Tournament",
			SportName:      "バスケットボール",
			MaxRound:       2,
			Round:          0,
			Team1ID:        sql.NullInt64{Int64: 100, Valid: true},
			Team2ID:        sql.NullInt64{Int64: 200, Valid: true},
			Team1Score:     sql.NullInt32{Int32: 20, Valid: true},
			Team2Score:     sql.NullInt32{Int32: 15, Valid: true},
			WinnerTeamID:   sql.NullInt64{Int64: 100, Valid: true},
			Status:         "completed",
			NextMatchID:    sql.NullInt64{Int64: 2, Valid: true},
			Team1Name:      sql.NullString{String: "IS3-A", Valid: true},
			Team2Name:      sql.NullString{String: "B2-A", Valid: true},
		},
		{
			MatchID:        2,
			TournamentID:   5,
			TournamentName: "バスケットボール Tournament",
			SportName:      "バスケットボール",
			MaxRound:       2,
			Round:          1,
			Team1ID:        sql.NullInt64{Int64: 100, Valid: true},
			Team2ID:        sql.NullInt64{Int64: 300, Valid: true},
			Team1Name:      sql.NullString{String: "IS3-A", Valid: true},
			Team2Name:      sql.NullString{String: "C1-A", Valid: true},
			WinnerTeamID:   sql.NullInt64{Valid: false},
			Status:         "scheduled",
			NextMatchID:    sql.NullInt64{Int64: 3, Valid: true},
		},
	}

	mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
	mockClassRepo.On("GetClassByRepRole", user.ID, 1).Return(class, nil).Once()
	mockClassRepo.On("GetClassMembers", class.ID).Return(members, nil).Once()
	mockTeamRepo.On("GetTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{team}, nil).Once()
	mockTeamRepo.On("GetTeamMembers", team.ID).Return(members, nil).Once()
	mockTournamentRepo.On("GetMatchesForTeam", 1, team.ID).Return(matchDetails, nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/student/class-progress", nil)
	c.Set("user", user)

	h.GetClassProgress(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(class.ID), response["class_id"])
	assert.Equal(t, class.Name, response["class_name"])

	memberList, ok := response["members"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, memberList, 2)
	firstMember := memberList[0].(map[string]interface{})
	assignments, ok := firstMember["assignments"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, assignments, 1)

	progressData, ok := response["progress"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, progressData, 1)

	mockEventRepo.AssertExpectations(t)
	mockClassRepo.AssertExpectations(t)
	mockTeamRepo.AssertExpectations(t)
	mockTournamentRepo.AssertExpectations(t)
}
