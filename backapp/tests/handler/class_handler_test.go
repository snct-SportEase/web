package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClassHandler_GetClassScores(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
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

	t.Run("with event_id query", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewClassHandler(mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo)

		expectedScores := []*models.ClassScore{
			{ID: 1, ClassName: "Class A", Season: "spring", TotalPointsCurrentEvent: 100, RankCurrentEvent: 1},
		}

		mockClassRepo.On("GetClassScoresByEvent", 2).Return(expectedScores, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/scores/class?event_id=2", nil)

		h.GetClassScores(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var actualScores []*models.ClassScore
		err := json.Unmarshal(w.Body.Bytes(), &actualScores)
		assert.NoError(t, err)
		assert.Equal(t, expectedScores, actualScores)

		mockClassRepo.AssertExpectations(t)
		mockEventRepo.AssertNotCalled(t, "GetActiveEvent")
	})

	t.Run("student cannot view active event scores when hidden", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewClassHandler(mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo)

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Twice()
		mockEventRepo.On("GetEventByID", 1).Return(&models.Event{ID: 1, HideScores: true}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/scores/class", nil)
		c.Set("user", &models.User{ID: "student-1", Roles: []models.Role{{Name: "student"}}})

		h.GetClassScores(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Scores are hidden for this event", response["message"])

		mockClassRepo.AssertNotCalled(t, "GetClassScoresByEvent", mock.Anything)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("root can view active event scores when hidden", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewClassHandler(mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo)

		expectedScores := []*models.ClassScore{
			{ID: 1, ClassName: "Class A", Season: "spring", TotalPointsCurrentEvent: 100, RankCurrentEvent: 1},
		}

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassScoresByEvent", 1).Return(expectedScores, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/scores/class", nil)
		c.Set("user", &models.User{ID: "root-1", Roles: []models.Role{{Name: "root"}}})

		h.GetClassScores(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var actualScores []*models.ClassScore
		err := json.Unmarshal(w.Body.Bytes(), &actualScores)
		assert.NoError(t, err)
		assert.Equal(t, expectedScores, actualScores)

		mockClassRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("no active event", func(t *testing.T) {
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

	t.Run("GetActiveEvent returns error", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewClassHandler(mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo)

		mockEventRepo.On("GetActiveEvent").Return(0, errors.New("db error")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/scores/class", nil)

		h.GetClassScores(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertNotCalled(t, "GetClassScoresByEvent", mock.Anything)
	})

	t.Run("repository returns error", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockTournamentRepo := new(MockTournamentRepository)

		h := handler.NewClassHandler(mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo)

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassScoresByEvent", 1).Return(nil, errors.New("db error")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/scores/class", nil)

		h.GetClassScores(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockClassRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
	})
}

func TestClassHandler_GetClassProgress(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newHandler := func() (*handler.ClassHandler, *MockClassRepository, *MockEventRepository, *MockTeamRepository, *MockTournamentRepository) {
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockTournamentRepo := new(MockTournamentRepository)
		h := handler.NewClassHandler(mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo)
		return h, mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo
	}

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

	t.Run("success", func(t *testing.T) {
		h, mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo := newHandler()

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassByRepRole", user.ID, 1).Return(class, nil).Once()
		mockClassRepo.On("GetClassMembers", class.ID).Return(members, nil).Once()
		mockTeamRepo.On("GetTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{team}, nil).Once()
		mockTeamRepo.On("GetNoonGameTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{}, nil).Once()
		mockTeamRepo.On("GetTeamMembersByTeamIDs", []int{team.ID}).Return(map[int][]*models.User{team.ID: members}, nil).Once()
		mockTeamRepo.On("GetTeamMembersByTeamIDs", []int{}).Return(map[int][]*models.User{}, nil).Once()
		mockTournamentRepo.On("GetMatchesForTeams", 1, []int{team.ID}).Return(map[int][]*models.MatchDetail{team.ID: matchDetails}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/student/class-progress", nil)
		c.Set("user", user)

		h.GetClassProgress(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(class.ID), response["class_id"])
		assert.Equal(t, class.Name, response["class_name"])

		memberList, ok := response["members"].([]any)
		assert.True(t, ok)
		assert.Len(t, memberList, 2)
		firstMember := memberList[0].(map[string]any)
		assignments, ok := firstMember["assignments"].([]any)
		assert.True(t, ok)
		assert.Len(t, assignments, 1)

		progressData, ok := response["progress"].([]any)
		assert.True(t, ok)
		assert.Len(t, progressData, 1)

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})

	t.Run("includes noon game relay assignments in member list", func(t *testing.T) {
		h, mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo := newHandler()

		noonGameTeam := &models.TeamWithSport{ID: 200, Name: "IS3リレー", SportID: 99, SportName: "学年対抗リレー"}

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassByRepRole", user.ID, 1).Return(class, nil).Once()
		mockClassRepo.On("GetClassMembers", class.ID).Return(members, nil).Once()
		mockTeamRepo.On("GetTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{}, nil).Once()
		mockTeamRepo.On("GetNoonGameTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{noonGameTeam}, nil).Once()
		mockTournamentRepo.On("GetMatchesForTeams", 1, []int{}).Return(map[int][]*models.MatchDetail{}, nil).Once()
		mockTeamRepo.On("GetTeamMembersByTeamIDs", []int{}).Return(map[int][]*models.User{}, nil).Once()
		mockTeamRepo.On("GetTeamMembersByTeamIDs", []int{noonGameTeam.ID}).Return(map[int][]*models.User{
			noonGameTeam.ID: {members[0]},
		}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/student/class-progress", nil)
		c.Set("user", user)

		h.GetClassProgress(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		memberList, ok := response["members"].([]any)
		assert.True(t, ok)
		assert.Len(t, memberList, 2)

		var relayAssignments []any
		for _, rawMember := range memberList {
			member := rawMember.(map[string]any)
			if member["id"] == members[0].ID {
				assignments, ok := member["assignments"].([]any)
				assert.True(t, ok)
				relayAssignments = assignments
				break
			}
		}

		assert.Len(t, relayAssignments, 1)
		firstAssignment := relayAssignments[0].(map[string]any)
		assert.Equal(t, noonGameTeam.SportName, firstAssignment["sport_name"])
		assert.Equal(t, noonGameTeam.Name, firstAssignment["team_name"])

		assert.Nil(t, response["progress"])

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})

	t.Run("does not show advancement after losing in same tournament", func(t *testing.T) {
		h, mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo := newHandler()

		losingMatches := []*models.MatchDetail{
			{
				MatchID:        10,
				TournamentID:   9,
				TournamentName: "卓球 Tournament",
				SportName:      "卓球",
				MaxRound:       2,
				Round:          0,
				MatchNumber:    0,
				Team1ID:        sql.NullInt64{Int64: 100, Valid: true},
				Team2ID:        sql.NullInt64{Int64: 200, Valid: true},
				Team1Score:     sql.NullInt32{Int32: 0, Valid: true},
				Team2Score:     sql.NullInt32{Int32: 5, Valid: true},
				WinnerTeamID:   sql.NullInt64{Int64: 200, Valid: true},
				Status:         "finished",
				NextMatchID:    sql.NullInt64{Int64: 11, Valid: true},
				Team1Name:      sql.NullString{String: "IS4", Valid: true},
				Team2Name:      sql.NullString{String: "IT3", Valid: true},
			},
			{
				MatchID:        11,
				TournamentID:   9,
				TournamentName: "卓球 Tournament",
				SportName:      "卓球",
				MaxRound:       2,
				Round:          0,
				MatchNumber:    1,
				Team1ID:        sql.NullInt64{Int64: 100, Valid: true},
				Team2ID:        sql.NullInt64{Int64: 300, Valid: true},
				WinnerTeamID:   sql.NullInt64{Valid: false},
				Status:         "pending",
				NextMatchID:    sql.NullInt64{Int64: 12, Valid: true},
				Team1Name:      sql.NullString{String: "IS4", Valid: true},
				Team2Name:      sql.NullString{String: "IE4", Valid: true},
			},
		}

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassByRepRole", user.ID, 1).Return(class, nil).Once()
		mockClassRepo.On("GetClassMembers", class.ID).Return(members, nil).Once()
		mockTeamRepo.On("GetTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{team}, nil).Once()
		mockTeamRepo.On("GetNoonGameTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{}, nil).Once()
		mockTeamRepo.On("GetTeamMembersByTeamIDs", []int{team.ID}).Return(map[int][]*models.User{team.ID: members}, nil).Once()
		mockTeamRepo.On("GetTeamMembersByTeamIDs", []int{}).Return(map[int][]*models.User{}, nil).Once()
		mockTournamentRepo.On("GetMatchesForTeams", 1, []int{team.ID}).Return(map[int][]*models.MatchDetail{team.ID: losingMatches}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/student/class-progress", nil)
		c.Set("user", user)

		h.GetClassProgress(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		progressData, ok := response["progress"].([]any)
		assert.True(t, ok)
		assert.Len(t, progressData, 1)

		progress := progressData[0].(map[string]any)
		assert.Equal(t, "準々決勝敗退", progress["status"])
		assert.Nil(t, progress["next_match"])

		lastMatch, ok := progress["last_match"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "敗戦", lastMatch["result"])

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})

	t.Run("hidden scores remove score text from class progress for students", func(t *testing.T) {
		h, mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo := newHandler()

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Twice()
		mockEventRepo.On("GetEventByID", 1).Return(&models.Event{ID: 1, HideScores: true}, nil).Once()
		mockClassRepo.On("GetClassByRepRole", user.ID, 1).Return(class, nil).Once()
		mockClassRepo.On("GetClassMembers", class.ID).Return(members, nil).Once()
		mockTeamRepo.On("GetTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{team}, nil).Once()
		mockTeamRepo.On("GetNoonGameTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{}, nil).Once()
		mockTeamRepo.On("GetTeamMembersByTeamIDs", []int{team.ID}).Return(map[int][]*models.User{team.ID: members}, nil).Once()
		mockTeamRepo.On("GetTeamMembersByTeamIDs", []int{}).Return(map[int][]*models.User{}, nil).Once()
		mockTournamentRepo.On("GetMatchesForTeams", 1, []int{team.ID}).Return(map[int][]*models.MatchDetail{team.ID: matchDetails}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/student/class-progress", nil)
		c.Set("user", &models.User{ID: user.ID, Roles: []models.Role{{Name: "student"}}})

		h.GetClassProgress(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		progressData, ok := response["progress"].([]any)
		assert.True(t, ok)
		assert.Len(t, progressData, 1)

		progress := progressData[0].(map[string]any)
		lastMatch, ok := progress["last_match"].(map[string]any)
		assert.True(t, ok)
		assert.Nil(t, lastMatch["score"])

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})

	t.Run("unauthorized - no user in context", func(t *testing.T) {
		h, mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo := newHandler()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/student/class-progress", nil)

		h.GetClassProgress(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		mockEventRepo.AssertNotCalled(t, "GetActiveEvent")
		mockClassRepo.AssertNotCalled(t, "GetClassByRepRole", mock.Anything, mock.Anything)
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByClassID", mock.Anything, mock.Anything)
		mockTournamentRepo.AssertNotCalled(t, "GetMatchesForTeams", mock.Anything, mock.Anything)
	})

	t.Run("GetActiveEvent returns error", func(t *testing.T) {
		h, mockClassRepo, mockEventRepo, _, _ := newHandler()

		mockEventRepo.On("GetActiveEvent").Return(0, errors.New("db error")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/student/class-progress", nil)
		c.Set("user", user)

		h.GetClassProgress(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertNotCalled(t, "GetClassByRepRole", mock.Anything, mock.Anything)
	})

	t.Run("no active event", func(t *testing.T) {
		h, mockClassRepo, mockEventRepo, _, _ := newHandler()

		mockEventRepo.On("GetActiveEvent").Return(0, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/student/class-progress", nil)
		c.Set("user", user)

		h.GetClassProgress(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertNotCalled(t, "GetClassByRepRole", mock.Anything, mock.Anything)
	})

	t.Run("GetClassByRepRole returns error", func(t *testing.T) {
		h, mockClassRepo, mockEventRepo, mockTeamRepo, _ := newHandler()

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassByRepRole", user.ID, 1).Return(nil, errors.New("db error")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/student/class-progress", nil)
		c.Set("user", user)

		h.GetClassProgress(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByClassID", mock.Anything, mock.Anything)
	})

	t.Run("user is not a class rep", func(t *testing.T) {
		h, mockClassRepo, mockEventRepo, mockTeamRepo, _ := newHandler()

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassByRepRole", user.ID, 1).Return(nil, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/student/class-progress", nil)
		c.Set("user", user)

		h.GetClassProgress(c)

		assert.Equal(t, http.StatusForbidden, w.Code)

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByClassID", mock.Anything, mock.Anything)
	})

	t.Run("GetTeamsByClassID returns error", func(t *testing.T) {
		h, mockClassRepo, mockEventRepo, mockTeamRepo, _ := newHandler()

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassByRepRole", user.ID, 1).Return(class, nil).Once()
		mockTeamRepo.On("GetTeamsByClassID", class.ID, 1).Return(nil, errors.New("db error")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/student/class-progress", nil)
		c.Set("user", user)

		h.GetClassProgress(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("GetClassMembers returns error", func(t *testing.T) {
		h, mockClassRepo, mockEventRepo, mockTeamRepo, _ := newHandler()

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassByRepRole", user.ID, 1).Return(class, nil).Once()
		mockTeamRepo.On("GetTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{team}, nil).Once()
		mockTeamRepo.On("GetNoonGameTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{}, nil).Once()
		mockClassRepo.On("GetClassMembers", class.ID).Return(nil, errors.New("db error")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/student/class-progress", nil)
		c.Set("user", user)

		h.GetClassProgress(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
	})

	t.Run("GetMatchesForTeams returns error", func(t *testing.T) {
		h, mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo := newHandler()

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassByRepRole", user.ID, 1).Return(class, nil).Once()
		mockTeamRepo.On("GetTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{team}, nil).Once()
		mockTeamRepo.On("GetNoonGameTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{}, nil).Once()
		mockClassRepo.On("GetClassMembers", class.ID).Return(members, nil).Once()
		mockTournamentRepo.On("GetMatchesForTeams", 1, []int{team.ID}).Return(nil, errors.New("db error")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/student/class-progress", nil)
		c.Set("user", user)

		h.GetClassProgress(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})

	t.Run("GetTeamMembersByTeamIDs returns error", func(t *testing.T) {
		h, mockClassRepo, mockEventRepo, mockTeamRepo, mockTournamentRepo := newHandler()

		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		mockClassRepo.On("GetClassByRepRole", user.ID, 1).Return(class, nil).Once()
		mockTeamRepo.On("GetTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{team}, nil).Once()
		mockTeamRepo.On("GetNoonGameTeamsByClassID", class.ID, 1).Return([]*models.TeamWithSport{}, nil).Once()
		mockClassRepo.On("GetClassMembers", class.ID).Return(members, nil).Once()
		mockTournamentRepo.On("GetMatchesForTeams", 1, []int{team.ID}).Return(map[int][]*models.MatchDetail{team.ID: matchDetails}, nil).Once()
		mockTeamRepo.On("GetTeamMembersByTeamIDs", []int{team.ID}).Return(nil, errors.New("db error")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/student/class-progress", nil)
		c.Set("user", user)

		h.GetClassProgress(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockEventRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockTournamentRepo.AssertExpectations(t)
	})
}
