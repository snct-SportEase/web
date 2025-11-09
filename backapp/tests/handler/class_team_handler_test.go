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
		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		managedClass := &models.Class{ID: 10, Name: "1A"}
		mockClassRepo.On("GetClassByRepRole", currentUser.ID, 1).Return(managedClass, nil).Once()
		sport := &models.Sport{ID: 1, Name: "Basketball"}
		mockSportRepo.On("GetSportByID", 1).Return(sport, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", managedClass.ID, 1, 1).Return(nil, nil).Once() // Team doesn't exist
		newTeamID := int64(100)
		mockTeamRepo.On("CreateTeam", mock.AnythingOfType("*models.Team")).Return(newTeamID, nil).Once()

		user1ClassID := 10
		user2ClassID := 10
		mockUser1 := &models.User{ID: "user1", ClassID: &user1ClassID}
		mockUser2 := &models.User{ID: "user2", ClassID: &user2ClassID}
		mockUserRepo.On("GetUserWithRoles", "user1").Return(mockUser1, nil).Once()
		mockUserRepo.On("GetUserWithRoles", "user2").Return(mockUser2, nil).Once()

		mockTeamRepo.On("AddTeamMember", int(newTeamID), "user1").Return(nil).Once()
		mockTeamRepo.On("AddTeamMember", int(newTeamID), "user2").Return(nil).Once()

		activeEventID := 1
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

	t.Run("Success - Admin Assigns Team Members", func(t *testing.T) {
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

		// Mock current user (admin)
		currentUser := &models.User{
			ID: "admin-user-id",
			Roles: []models.Role{
				{Name: "admin"},
			},
		}

		// Mock request body
		reqBody := gin.H{
			"class_id": 10,
			"sport_id": 1,
			"user_ids": []string{"user1", "user2"},
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Mock repository calls
		mockEventRepo.On("GetActiveEvent").Return(1, nil).Once()
		managedClass := &models.Class{ID: 10, Name: "1A"}
		mockClassRepo.On("GetClassByID", 10).Return(managedClass, nil).Once()
		sport := &models.Sport{ID: 1, Name: "Basketball"}
		mockSportRepo.On("GetSportByID", 1).Return(sport, nil).Once()
		existingTeam := &models.Team{ID: 100}
		mockTeamRepo.On("GetTeamByClassAndSport", managedClass.ID, 1, 1).Return(existingTeam, nil).Once() // Team exists

		user1ClassID := 10
		user2ClassID := 10
		mockUser1 := &models.User{ID: "user1", ClassID: &user1ClassID}
		mockUser2 := &models.User{ID: "user2", ClassID: &user2ClassID}
		mockUserRepo.On("GetUserWithRoles", "user1").Return(mockUser1, nil).Once()
		mockUserRepo.On("GetUserWithRoles", "user2").Return(mockUser2, nil).Once()

		mockTeamRepo.On("AddTeamMember", existingTeam.ID, "user1").Return(nil).Once()
		mockTeamRepo.On("AddTeamMember", existingTeam.ID, "user2").Return(nil).Once()

		activeEventID := 1
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
}
