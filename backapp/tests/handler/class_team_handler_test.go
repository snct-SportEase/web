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

	t.Run("Success - User With Class Role Assigns Team Members", func(t *testing.T) {
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

		// Mock current user with the class role
		currentUser := &models.User{
			ID:      "rep-user-id",
			ClassID: classTeamIntPtr(10),
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
		mockEventRepo.On("GetEventByID", activeEventID).Return(&models.Event{ID: activeEventID, DuplicateRegistrationThreshold: 31}, nil).Once()
		managedClass := &models.Class{ID: 10, Name: "1A"}
		mockClassRepo.On("GetClassByID", managedClass.ID).Return(managedClass, nil).Once()
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
		mockTeamRepo.On("GetTeamsByUserID", "user1").Return([]*models.TeamWithSport{}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user2").Return([]*models.TeamWithSport{}, nil).Once()

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
			ID:      "rep-user-id",
			ClassID: classTeamIntPtr(10),
		}

		reqBody := gin.H{
			"sport_id": 1,
			"user_ids": []string{"user3"},
		}
		jsonBody, _ := json.Marshal(reqBody)

		activeEventID := 1
		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		mockEventRepo.On("GetEventByID", activeEventID).Return(&models.Event{ID: activeEventID, DuplicateRegistrationThreshold: 31}, nil).Once()
		managedClass := &models.Class{ID: 10, Name: "1A"}
		mockClassRepo.On("GetClassByID", managedClass.ID).Return(managedClass, nil).Once()
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

	t.Run("Error - Duplicate Registration For Regular Class", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockSportRepo := new(MockSportRepository)
		h := handler.NewClassTeamHandler(mockClassRepo, mockTeamRepo, mockUserRepo, mockEventRepo, mockSportRepo)

		currentUser := &models.User{ID: "rep-user-id", ClassID: classTeamIntPtr(10)}
		body, _ := json.Marshal(gin.H{"sport_id": 2, "user_ids": []string{"user1"}})
		activeEventID := 1
		classID := 10
		managedClass := &models.Class{ID: classID, Name: "1A", StudentCount: 40}

		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		mockEventRepo.On("GetEventByID", activeEventID).Return(&models.Event{ID: activeEventID, DuplicateRegistrationThreshold: 31}, nil).Once()
		mockClassRepo.On("GetClassByID", managedClass.ID).Return(managedClass, nil).Once()
		mockSportRepo.On("GetSportByID", 2).Return(&models.Sport{ID: 2, Name: "Volleyball"}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", classID, 2, activeEventID).Return(&models.Team{ID: 200}, nil).Once()
		mockSportRepo.On("GetSportDetails", activeEventID, 2).Return(&models.EventSport{}, nil).Once()
		mockUserRepo.On("GetUserWithRoles", "user1").Return(&models.User{ID: "user1", ClassID: &classID}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user1").Return([]*models.TeamWithSport{{SportID: 1, EventID: activeEventID}}, nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/assign", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", currentUser)
		h.AssignTeamMembersHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "1競技")
		mockTeamRepo.AssertNotCalled(t, "AddTeamMember", mock.Anything, mock.Anything)
	})

	t.Run("Success - Small Class Can Register For Two Sports", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockSportRepo := new(MockSportRepository)
		h := handler.NewClassTeamHandler(mockClassRepo, mockTeamRepo, mockUserRepo, mockEventRepo, mockSportRepo)

		currentUser := &models.User{ID: "rep-user-id", ClassID: classTeamIntPtr(10)}
		body, _ := json.Marshal(gin.H{"sport_id": 2, "user_ids": []string{"user1"}})
		activeEventID := 1
		classID := 10
		managedClass := &models.Class{ID: classID, Name: "1A", StudentCount: 25}

		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		mockEventRepo.On("GetEventByID", activeEventID).Return(&models.Event{ID: activeEventID, DuplicateRegistrationThreshold: 25}, nil).Once()
		mockClassRepo.On("GetClassByID", managedClass.ID).Return(managedClass, nil).Once()
		mockSportRepo.On("GetSportByID", 2).Return(&models.Sport{ID: 2, Name: "Volleyball"}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", classID, 2, activeEventID).Return(&models.Team{ID: 200}, nil).Once()
		mockSportRepo.On("GetSportDetails", activeEventID, 2).Return(&models.EventSport{}, nil).Once()
		mockUserRepo.On("GetUserWithRoles", "user1").Return(&models.User{ID: "user1", ClassID: &classID}, nil).Once()
		mockTeamRepo.On("GetTeamsByUserID", "user1").Return([]*models.TeamWithSport{{SportID: 1, EventID: activeEventID}}, nil).Once()
		mockTeamRepo.On("AddTeamMember", 200, "user1").Return(nil).Once()
		mockUserRepo.On("UpdateUserRole", "user1", "1A_Volleyball", &activeEventID).Return(nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/assign", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", currentUser)
		h.AssignTeamMembersHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Success - Special Class Has No Duplicate Limit", func(t *testing.T) {
		mockClassRepo := new(MockClassRepository)
		mockTeamRepo := new(MockTeamRepository)
		mockUserRepo := new(MockUserRepository)
		mockEventRepo := new(MockEventRepository)
		mockSportRepo := new(MockSportRepository)
		h := handler.NewClassTeamHandler(mockClassRepo, mockTeamRepo, mockUserRepo, mockEventRepo, mockSportRepo)

		currentUser := &models.User{ID: "rep-user-id", ClassID: classTeamIntPtr(20)}
		body, _ := json.Marshal(gin.H{"sport_id": 3, "user_ids": []string{"teacher1"}})
		activeEventID := 1
		classID := 20
		managedClass := &models.Class{ID: classID, Name: "専教", StudentCount: 40}

		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		mockEventRepo.On("GetEventByID", activeEventID).Return(&models.Event{ID: activeEventID, DuplicateRegistrationThreshold: 31}, nil).Once()
		mockClassRepo.On("GetClassByID", managedClass.ID).Return(managedClass, nil).Once()
		mockSportRepo.On("GetSportByID", 3).Return(&models.Sport{ID: 3, Name: "Tennis"}, nil).Once()
		mockTeamRepo.On("GetTeamByClassAndSport", classID, 3, activeEventID).Return(&models.Team{ID: 300}, nil).Once()
		mockSportRepo.On("GetSportDetails", activeEventID, 3).Return(&models.EventSport{}, nil).Once()
		mockUserRepo.On("GetUserWithRoles", "teacher1").Return(&models.User{ID: "teacher1", ClassID: &classID}, nil).Once()
		mockTeamRepo.On("AddTeamMember", 300, "teacher1").Return(nil).Once()
		mockUserRepo.On("UpdateUserRole", "teacher1", "専教_Tennis", &activeEventID).Return(nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/assign", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user", currentUser)
		h.AssignTeamMembersHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockTeamRepo.AssertNotCalled(t, "GetTeamsByUserID", mock.Anything)
	})
}

func TestClassTeamHandler_DuplicateRegistrationRules(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		studentCount  int
		threshold     int
		existingTeams []*models.TeamWithSport
		wantStatus    int
		wantLimitText string
	}{
		{
			name:         "class at configured threshold can register second sport",
			studentCount: 20, threshold: 20,
			existingTeams: []*models.TeamWithSport{{SportID: 1, EventID: 10}},
			wantStatus:    http.StatusOK,
		},
		{
			name:         "class above configured threshold cannot register second sport",
			studentCount: 21, threshold: 20,
			existingTeams: []*models.TeamWithSport{{SportID: 1, EventID: 10}},
			wantStatus:    http.StatusBadRequest, wantLimitText: "1競技",
		},
		{
			name:         "small class cannot register third sport",
			studentCount: 20, threshold: 20,
			existingTeams: []*models.TeamWithSport{{SportID: 1, EventID: 10}, {SportID: 2, EventID: 10}},
			wantStatus:    http.StatusBadRequest, wantLimitText: "2競技",
		},
		{
			name:         "registrations from another event are ignored",
			studentCount: 40, threshold: 20,
			existingTeams: []*models.TeamWithSport{{SportID: 1, EventID: 9}, {SportID: 2, EventID: 9}},
			wantStatus:    http.StatusOK,
		},
		{
			name:         "assigning the same sport again does not consume another slot",
			studentCount: 40, threshold: 20,
			existingTeams: []*models.TeamWithSport{{SportID: 3, EventID: 10}},
			wantStatus:    http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classRepo := new(MockClassRepository)
			teamRepo := new(MockTeamRepository)
			userRepo := new(MockUserRepository)
			eventRepo := new(MockEventRepository)
			sportRepo := new(MockSportRepository)
			h := handler.NewClassTeamHandler(classRepo, teamRepo, userRepo, eventRepo, sportRepo)

			const eventID = 10
			const classID = 100
			const sportID = 3
			currentUser := &models.User{ID: "rep", ClassID: classTeamIntPtr(classID)}
			managedClass := &models.Class{ID: classID, Name: "1A", StudentCount: tt.studentCount}

			eventRepo.On("GetActiveEvent").Return(eventID, nil).Once()
			eventRepo.On("GetEventByID", eventID).Return(&models.Event{ID: eventID, DuplicateRegistrationThreshold: tt.threshold}, nil).Once()
			classRepo.On("GetClassByID", classID).Return(managedClass, nil).Once()
			sportRepo.On("GetSportByID", sportID).Return(&models.Sport{ID: sportID, Name: "Tennis"}, nil).Once()
			teamRepo.On("GetTeamByClassAndSport", classID, sportID, eventID).Return(&models.Team{ID: 300}, nil).Once()
			sportRepo.On("GetSportDetails", eventID, sportID).Return(&models.EventSport{}, nil).Once()
			userRepo.On("GetUserWithRoles", "user1").Return(&models.User{ID: "user1", ClassID: classTeamIntPtr(classID)}, nil).Once()
			teamRepo.On("GetTeamsByUserID", "user1").Return(tt.existingTeams, nil).Once()

			if tt.wantStatus == http.StatusOK {
				teamRepo.On("AddTeamMember", 300, "user1").Return(nil).Once()
				userRepo.On("UpdateUserRole", "user1", "1A_Tennis", mock.Anything).Return(nil).Once()
			}

			body, _ := json.Marshal(gin.H{"sport_id": sportID, "user_ids": []string{"user1"}})
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/assign", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set("user", currentUser)

			h.AssignTeamMembersHandler(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantLimitText != "" {
				assert.Contains(t, w.Body.String(), tt.wantLimitText)
				teamRepo.AssertNotCalled(t, "AddTeamMember", mock.Anything, mock.Anything)
			}
			eventRepo.AssertExpectations(t)
			teamRepo.AssertExpectations(t)
		})
	}
}

func classTeamIntPtr(value int) *int {
	return &value
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
			ID:      "rep-user-id",
			ClassID: classTeamIntPtr(10),
		}

		reqBody := gin.H{
			"sport_id": 1,
			"user_id":  "user1",
		}
		jsonBody, _ := json.Marshal(reqBody)

		activeEventID := 1
		mockEventRepo.On("GetActiveEvent").Return(activeEventID, nil).Once()
		managedClass := &models.Class{ID: 10, Name: "1A"}
		mockClassRepo.On("GetClassByID", managedClass.ID).Return(managedClass, nil).Once()
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
