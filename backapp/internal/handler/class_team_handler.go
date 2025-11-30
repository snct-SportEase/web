package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ClassTeamHandler handles class and team management API requests
type ClassTeamHandler struct {
	classRepo repository.ClassRepository
	teamRepo  repository.TeamRepository
	userRepo  repository.UserRepository
	eventRepo repository.EventRepository
	sportRepo repository.SportRepository
}

// NewClassTeamHandler creates a new instance of ClassTeamHandler
func NewClassTeamHandler(classRepo repository.ClassRepository, teamRepo repository.TeamRepository, userRepo repository.UserRepository, eventRepo repository.EventRepository, sportRepo repository.SportRepository) *ClassTeamHandler {
	return &ClassTeamHandler{
		classRepo: classRepo,
		teamRepo:  teamRepo,
		userRepo:  userRepo,
		eventRepo: eventRepo,
		sportRepo: sportRepo,
	}
}

// GetManagedClassHandler returns the class that the current user can manage based on class_name_rep role
func (h *ClassTeamHandler) GetManagedClassHandler(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	user := userCtx.(*models.User)

	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}

	if activeEventID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active event found"})
		return
	}

	// Check if user is admin or root - if so, return all classes
	isAdmin := false
	for _, role := range user.Roles {
		if role.Name == "admin" || role.Name == "root" {
			isAdmin = true
			break
		}
	}

	if isAdmin {
		// For admin users, return all classes
		classes, err := h.classRepo.GetAllClasses(activeEventID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get classes"})
			return
		}
		c.JSON(http.StatusOK, classes)
		return
	}

	// For class rep users, return only their managed class
	class, err := h.classRepo.GetClassByRepRole(user.ID, activeEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get managed class"})
		return
	}

	if class == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No class found for this user"})
		return
	}

	c.JSON(http.StatusOK, []*models.Class{class})
}

// GetClassMembersHandler returns all members of a class
func (h *ClassTeamHandler) GetClassMembersHandler(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	user := userCtx.(*models.User)

	classIDStr := c.Param("class_id")
	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
		return
	}

	// Verify that the user can manage this class
	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}

	// Check if user is admin or root
	isAdmin := false
	for _, role := range user.Roles {
		if role.Name == "admin" || role.Name == "root" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		// For class rep users, verify they can manage this class
		managedClass, err := h.classRepo.GetClassByRepRole(user.ID, activeEventID)
		if err != nil || managedClass == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to manage this class"})
			return
		}

		if managedClass.ID != classID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to manage this class"})
			return
		}
	}

	members, err := h.classRepo.GetClassMembers(classID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get class members"})
		return
	}

	c.JSON(http.StatusOK, members)
}

// AssignTeamMembersHandler assigns users to a team and assigns the class_name_sport_name role
func (h *ClassTeamHandler) AssignTeamMembersHandler(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	currentUser := userCtx.(*models.User)

	var req struct {
		ClassID *int     `json:"class_id"` // Optional for admin users
		SportID int      `json:"sport_id"`
		UserIDs []string `json:"user_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get active event
	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}

	if activeEventID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active event found"})
		return
	}

	// Verify that the user can manage a class and get the class
	var managedClass *models.Class
	isAdmin := false
	for _, role := range currentUser.Roles {
		if role.Name == "admin" || role.Name == "root" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		// For class rep users, get their managed class
		var err error
		managedClass, err = h.classRepo.GetClassByRepRole(currentUser.ID, activeEventID)
		if err != nil || managedClass == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to manage a class"})
			return
		}
	} else {
		// For admin users, get the class from class_id in the request
		if req.ClassID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "class_id is required for admin users"})
			return
		}
		var err error
		managedClass, err = h.classRepo.GetClassByID(*req.ClassID)
		if err != nil || managedClass == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
			return
		}
	}

	// Get sport information
	sport, err := h.sportRepo.GetSportByID(req.SportID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sport not found"})
		return
	}

	// Get or create team
	team, err := h.teamRepo.GetTeamByClassAndSport(managedClass.ID, req.SportID, activeEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get team"})
		return
	}

	if team == nil {
		// Create team if it doesn't exist
		newTeam := &models.Team{
			Name:    managedClass.Name,
			ClassID: managedClass.ID,
			SportID: req.SportID,
			EventID: activeEventID,
		}
		teamID, err := h.teamRepo.CreateTeam(newTeam)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team"})
			return
		}
		team = &models.Team{
			ID:      int(teamID),
			Name:    newTeam.Name,
			ClassID: newTeam.ClassID,
			SportID: newTeam.SportID,
			EventID: newTeam.EventID,
		}
	}

	// --- Capacity Check ---
	var maxCapacity *int

	// 1. Check team specific capacity
	if team.MaxCapacity != nil {
		maxCapacity = team.MaxCapacity
	} else {
		// 2. Check event sport default capacity
		eventSport, err := h.sportRepo.GetSportDetails(activeEventID, req.SportID)
		if err == nil && eventSport != nil {
			maxCapacity = eventSport.MaxCapacity
		}
	}

	if maxCapacity != nil {
		// Get current members
		currentMembers, err := h.teamRepo.GetTeamMembers(team.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current team members for capacity check"})
			return
		}
		
		// Check capacity
		currentCount := len(currentMembers)
		addCount := 0
		
		// Filter out users who are already members to avoid double counting
		// Although AddTeamMember ignores duplicates, for capacity check we should be precise
		existingMemberMap := make(map[string]bool)
		for _, m := range currentMembers {
			existingMemberMap[m.ID] = true
		}
		
		for _, uid := range req.UserIDs {
			if !existingMemberMap[uid] {
				addCount++
			}
		}
		
		if currentCount + addCount > *maxCapacity {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("定員オーバーです。現在のメンバー数: %d, 追加人数: %d, 定員: %d", currentCount, addCount, *maxCapacity),
			})
			return
		}
	}
	// --- End Capacity Check ---

	// Create role name: class_name_sport_name
	roleName := fmt.Sprintf("%s_%s", managedClass.Name, sport.Name)

	// Assign team members and roles
	assignedCount := 0
	for _, userID := range req.UserIDs {
		// Verify user belongs to the class
		user, err := h.userRepo.GetUserWithRoles(userID)
		if err != nil || user == nil {
			continue // Skip invalid users
		}

		if user.ClassID == nil || *user.ClassID != managedClass.ID {
			continue // Skip users not in the class
		}

		// Add to team_members (ignore duplicate errors)
		err = h.teamRepo.AddTeamMember(team.ID, userID)
		if err != nil {
			// Continue if user is already a member, but still assign role if needed
		}

		// Assign role
		err = h.userRepo.UpdateUserRole(userID, roleName, &activeEventID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to assign role to user %s: %v", userID, err)})
			return
		}
		assignedCount++
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Team members assigned successfully. %d users assigned.", assignedCount)})
}

// RemoveTeamMemberHandler removes a user from a team and removes the class_name_sport_name role
func (h *ClassTeamHandler) RemoveTeamMemberHandler(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	currentUser := userCtx.(*models.User)

	var req struct {
		ClassID *int   `json:"class_id"` // Optional for admin users
		SportID int    `json:"sport_id"`
		UserID  string `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get active event
	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}

	if activeEventID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active event found"})
		return
	}

	// Verify that the user can manage a class and get the class
	var managedClass *models.Class
	isAdmin := false
	for _, role := range currentUser.Roles {
		if role.Name == "admin" || role.Name == "root" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		// For class rep users, get their managed class
		var err error
		managedClass, err = h.classRepo.GetClassByRepRole(currentUser.ID, activeEventID)
		if err != nil || managedClass == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to manage a class"})
			return
		}
	} else {
		// For admin users, get the class from class_id in the request
		if req.ClassID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "class_id is required for admin users"})
			return
		}
		var err error
		managedClass, err = h.classRepo.GetClassByID(*req.ClassID)
		if err != nil || managedClass == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID"})
			return
		}
	}

	// Get sport information
	sport, err := h.sportRepo.GetSportByID(req.SportID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sport not found"})
		return
	}

	// Get team
	team, err := h.teamRepo.GetTeamByClassAndSport(managedClass.ID, req.SportID, activeEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get team"})
		return
	}

	if team == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Remove from team_members
	err = h.teamRepo.RemoveTeamMember(team.ID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove team member"})
		return
	}

	// Remove role
	roleName := fmt.Sprintf("%s_%s", managedClass.Name, sport.Name)
	err = h.userRepo.DeleteUserRole(req.UserID, roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team member removed successfully"})
}

// GetTeamMembersHandler returns all members of a team
func (h *ClassTeamHandler) GetTeamMembersHandler(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	currentUser := userCtx.(*models.User)

	sportIDStr := c.Param("sport_id")
	sportID, err := strconv.Atoi(sportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sport ID"})
		return
	}

	// Get active event
	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}

	// Verify that the user can manage a class
	var managedClass *models.Class
	isAdmin := false
	for _, role := range currentUser.Roles {
		if role.Name == "admin" || role.Name == "root" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		// For class rep users, get their managed class
		var err error
		managedClass, err = h.classRepo.GetClassByRepRole(currentUser.ID, activeEventID)
		if err != nil || managedClass == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to manage a class"})
			return
		}
	} else {
		// For admin users, we need class_id from query parameter
		classIDStr := c.Query("class_id")
		if classIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "class_id is required for admin users"})
			return
		}
		classID, err := strconv.Atoi(classIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class_id"})
			return
		}
		managedClass, err = h.classRepo.GetClassByID(classID)
		if err != nil || managedClass == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Class not found"})
			return
		}
	}

	// Get team
	team, err := h.teamRepo.GetTeamByClassAndSport(managedClass.ID, sportID, activeEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get team"})
		return
	}

	if team == nil {
		c.JSON(http.StatusOK, []*models.User{}) // Empty team
		return
	}

	// Get team members
	members, err := h.teamRepo.GetTeamMembers(team.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get team members"})
		return
	}

	c.JSON(http.StatusOK, members)
}
