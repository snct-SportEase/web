package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClassTeamHandler_AssignTeamMembersHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Class Rep Assigns Team Members", func(t *testing.T) {
		// Setup mocks
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockSportRepo := new(MockSportRepository)

		// Create handler
		h := handler.NewClassTeamHandler(mockClassRepo, mockTeamRepo, mockUserRepo, mockEventRepo, mockSportRepo)

		// Setup router
		r := gin.Default()
		r.POST("/assign", h.AssignTeamMembersHandler)

		// Mock current user (class rep)
		currentUser := &models.User{
			ID: "rep-user-id",
			Roles: []models.Role{
				{Name: "1A_rep"},
			},
		}

		// Mock request body
		reqBody := gin.H{
			"sport_id": 1,
			"user_ids": []string{"user1", "user2"},
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock repository calls
		activeEventID := 1
		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		managedClass := &models.Class{ID: 10, Name: "1A"}
		mockClassRepo.On("GetClassByRepRole", currentUser.ID, activeEventID).Return(managedClass, nil).Once()
		sport := &models.Sport{ID: 1, Name: "Basketball"}
		mockSportRepo.On("GetSportByID", 1).Return(sport, nil).Once()
		
		// Team lookup
		mockTeamRepo.On("GetTeamByClassAndSport", managedClass.ID, 1, activeEventID).Return(nil, nil).Once()
		
		// Team creation
		newTeamID := int64(100)
		mockTeamRepo.On("CreateTeam", mock.AnythingOfType("*models.Team")).Return(newTeamID, nil).Once()

		// Capacity check
		// Since team is new, MaxCapacity is nil. It falls back to SportDetails
		mockSportRepo.On("GetSportDetails", activeEventID, 1).Return(&models.EventSport{MaxCapacity: nil}, nil).Once()

		user1ClassID := 10
		user2ClassID := 10
		mockUser1 := &models.User{ID: "user1", ClassID: &user1ClassID}
		mockUser2 := &models.User{ID: "user2", ClassID: &user2ClassID}
		mockUserRepo.On("GetUserWithRoles", "user1").Return(mockUser1, nil).Once()
		mockUserRepo.On("GetUserWithRoles", "user2").Return(mockUser2, nil).Once()

		mockTeamRepo.On("AddTeamMember", int(newTeamID), "user1").Return(nil).Once()
		mockTeamRepo.On("AddTeamMember", int(newTeamID), "user2").Return(nil).Once()

		mockUserRepo.On("UpdateUserRole", "user1", "1A_Basketball", &activeEventID).Return(nil).Once()
		mockUserRepo.On("UpdateUserRole", "user2", "1A_Basketball", &activeEventID).Return(nil).Once()

		// Create request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/assign", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Set user in context
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", currentUser)

		// Execute handler
		h.AssignTeamMembersHandler(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		var response gin.H
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Team members assigned successfully. 2 users assigned.", response["message"])

		// Verify mocks
		mockClassRepo.AssertExpectations(t)
		mockTeamRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
		mockSportRepo.AssertExpectations(t)
	})

	t.Run("Error - Capacity Exceeded", func(t *testing.T) {
		// Setup mocks
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockSportRepo := new(MockSportRepository)

		h := handler.NewClassTeamHandler(mockClassRepo, mockTeamRepo, mockUserRepo, mockEventRepo, mockSportRepo)
		r := gin.Default()
		r.POST("/assign", h.AssignTeamMembersHandler)

		currentUser := &models.User{
			ID: "rep-user-id",
			Roles: []models.Role{{Name: "1A_rep"}},
		}

		reqBody := gin.H{
			"sport_id": 1,
			"user_ids": []string{"user3"},
		}
		jsonBody, _ := json.Marshal(reqBody)

		activeEventID := 1
		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		managedClass := &models.Class{ID: 10, Name: "1A"}
		mockClassRepo.On("GetClassByRepRole", currentUser.ID, activeEventID).Return(managedClass, nil).Once()
		sport := &models.Sport{ID: 1, Name: "Basketball"}
		mockSportRepo.On("GetSportByID", 1).Return(sport, nil).Once()

		// Team exists and has capacity limit
		limit := 2
		existingTeam := &models.Team{ID: 100, ClassID: 10, SportID: 1, EventID: activeEventID, MaxCapacity: &limit}
		mockTeamRepo.On("GetTeamByClassAndSport", managedClass.ID, 1, activeEventID).Return(existingTeam, nil).Once()

		// Get current members to check capacity
		currentMembers := []*models.User{{ID: "user1"}, {ID: "user2"}} // Already 2 members
		mockTeamRepo.On("GetTeamMembers", existingTeam.ID).Return(currentMembers, nil).Once()

		// Request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/assign", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", currentUser)

		h.AssignTeamMembersHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response gin.H
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "定員オーバーです")

		mockTeamRepo.AssertExpectations(t)
	})
}

func TestClassTeamHandler_RemoveTeamMemberHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Remove Team Member", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockSportRepo := new(MockSportRepository)

		h := handler.NewClassTeamHandler(mockClassRepo, mockTeamRepo, mockUserRepo, mockEventRepo, mockSportRepo)

		currentUser := &models.User{
			ID: "rep-user-id",
			Roles: []models.Role{{Name: "1A_rep"}},
		}

		reqBody := gin.H{
			"sport_id": 1,
			"user_id":  "user1",
		}
		jsonBody, _ := json.Marshal(reqBody)

		activeEventID := 1
		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		managedClass := &models.Class{ID: 10, Name: "1A"}
		mockClassRepo.On("GetClassByRepRole", currentUser.ID, activeEventID).Return(managedClass, nil).Once()
		sport := &models.Sport{ID: 1, Name: "Basketball"}
		mockSportRepo.On("GetSportByID", 1).Return(sport, nil).Once()

		existingTeam := &models.Team{ID: 100}
		mockTeamRepo.On("GetTeamByClassAndSport", managedClass.ID, 1, activeEventID).Return(existingTeam, nil).Once()

		// Remove member
		mockTeamRepo.On("RemoveTeamMember", existingTeam.ID, "user1").Return(nil).Once()
		
		// Remove role
		roleName := fmt.Sprintf("%s_%s", managedClass.Name, sport.Name)
		mockUserRepo.On("DeleteUserRole", "user1", roleName).Return(nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/remove-member", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", currentUser)

		h.RemoveTeamMemberHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockTeamRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}
