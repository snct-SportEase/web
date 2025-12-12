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

func TestClassTeamHandler_GetConfirmedTeamMembersHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Admin gets confirmed team members", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockSportRepo := new(MockSportRepository)

		h := handler.NewClassTeamHandler(mockClassRepo, mockTeamRepo, mockUserRepo, mockEventRepo, mockSportRepo)

		currentUser := &models.User{
			ID: "admin-user-id",
			Roles: []models.Role{
				{Name: "admin"},
			},
		}

		activeEventID := 1
		classID := 1
		sportID := 1
		teamID := 1

		class := &models.Class{
			ID:   classID,
			Name: "1-1",
		}

		minCapacity := 5
		team := &models.Team{
			ID:          teamID,
			Name:        "Team A",
			ClassID:     classID,
			SportID:     sportID,
			EventID:     activeEventID,
			MinCapacity: &minCapacity,
			MaxCapacity: nil,
		}

		displayName1 := "User 1"
		displayName2 := "User 2"
		confirmedMembers := []*models.User{
			{
				ID:          "user1",
				Email:       "user1@example.com",
				DisplayName: &displayName1,
			},
			{
				ID:          "user2",
				Email:       "user2@example.com",
				DisplayName: &displayName2,
			},
		}

		confirmedCount := 2

		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		mockClassRepo.On("GetClassByID", classID).Return(class, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", classID, sportID, activeEventID).Return(team, nil).Once()
		mockTeamRepo.On("GetConfirmedTeamMembers", teamID).Return(confirmedMembers, nil).Once()
		mockTeamRepo.On("GetConfirmedTeamMembersCount", teamID).Return(confirmedCount, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", currentUser)
		c.Params = gin.Params{
			{Key: "sport_id", Value: "1"},
		}
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/admin/class-team/sports/1/confirmed-members?class_id=1", nil)

		h.GetConfirmedTeamMembersHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "members")
		assert.Contains(t, response, "confirmed_count")
		assert.Contains(t, response, "min_capacity")
		assert.Contains(t, response, "capacity_ok")

		members, ok := response["members"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, members, 2)

		assert.Equal(t, float64(confirmedCount), response["confirmed_count"])
		assert.Equal(t, float64(5), response["min_capacity"])
		assert.Equal(t, false, response["capacity_ok"]) // 2 < 5

		mockClassRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("Success - No capacity requirement", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockSportRepo := new(MockSportRepository)

		h := handler.NewClassTeamHandler(mockClassRepo, mockTeamRepo, mockUserRepo, mockEventRepo, mockSportRepo)

		currentUser := &models.User{
			ID: "admin-user-id",
			Roles: []models.Role{
				{Name: "admin"},
			},
		}

		activeEventID := 1
		classID := 1
		sportID := 1
		teamID := 1

		class := &models.Class{
			ID:   classID,
			Name: "1-1",
		}

		team := &models.Team{
			ID:          teamID,
			Name:        "Team A",
			ClassID:     classID,
			SportID:     sportID,
			EventID:     activeEventID,
			MinCapacity: nil, // No capacity requirement
			MaxCapacity: nil,
		}

		confirmedMembers := []*models.User{}
		confirmedCount := 0

		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		mockClassRepo.On("GetClassByID", classID).Return(class, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", classID, sportID, activeEventID).Return(team, nil).Once()
		mockTeamRepo.On("GetConfirmedTeamMembers", teamID).Return(confirmedMembers, nil).Once()
		mockTeamRepo.On("GetConfirmedTeamMembersCount", teamID).Return(confirmedCount, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", currentUser)
		c.Params = gin.Params{
			{Key: "sport_id", Value: "1"},
		}
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/admin/class-team/sports/1/confirmed-members?class_id=1", nil)

		h.GetConfirmedTeamMembersHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Nil(t, response["min_capacity"])
		assert.Equal(t, true, response["capacity_ok"])

		mockClassRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("Failure - Team not found", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockSportRepo := new(MockSportRepository)

		h := handler.NewClassTeamHandler(mockClassRepo, mockTeamRepo, mockUserRepo, mockEventRepo, mockSportRepo)

		currentUser := &models.User{
			ID: "admin-user-id",
			Roles: []models.Role{
				{Name: "admin"},
			},
		}

		activeEventID := 1
		classID := 1
		sportID := 1

		class := &models.Class{
			ID:   classID,
			Name: "1-1",
		}

		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		mockClassRepo.On("GetClassByID", classID).Return(class, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", classID, sportID, activeEventID).Return(nil, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", currentUser)
		c.Params = gin.Params{
			{Key: "sport_id", Value: "1"},
		}
		c.Request, _ = http.NewRequest(http.MethodGet, "/api/admin/class-team/sports/1/confirmed-members?class_id=1", nil)

		h.GetConfirmedTeamMembersHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{}, response["members"])
		assert.Equal(t, float64(0), response["confirmed_count"])
		assert.Nil(t, response["min_capacity"])
		assert.Equal(t, true, response["capacity_ok"])

		mockClassRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
	})
}
