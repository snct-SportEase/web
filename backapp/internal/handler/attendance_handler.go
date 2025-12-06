package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AttendanceHandler struct {
	classRepo repository.ClassRepository
	eventRepo repository.EventRepository
}

func NewAttendanceHandler(classRepo repository.ClassRepository, eventRepo repository.EventRepository) *AttendanceHandler {
	return &AttendanceHandler{
		classRepo: classRepo,
		eventRepo: eventRepo,
	}
}

func (h *AttendanceHandler) GetClassDetailsHandler(c *gin.Context) {
	classIDStr := c.Param("classID")
	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID format"})
		return
	}

	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}
	if activeEventID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active event found"})
		return
	}

	// Get user from context
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	user := userCtx.(*models.User)

	// Check if user is admin or root
	isAdmin := false
	for _, role := range user.Roles {
		if role.Name == "admin" || role.Name == "root" {
			isAdmin = true
			break
		}
	}

	// If user is admin and has a class_name_rep role, restrict to that class only
	if isAdmin {
		managedClass, err := h.classRepo.GetClassByRepRole(user.ID, activeEventID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check managed class"})
			return
		}
		// If managedClass exists, user can only get details for that class
		if managedClass != nil {
			if managedClass.ID != classID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You are only authorized to view details for your managed class"})
				return
			}
		}
	}

	classDetails, err := h.classRepo.GetClassDetails(classID, activeEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get class details"})
		return
	}
	if classDetails == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Class details not found"})
		return
	}

	c.JSON(http.StatusOK, classDetails)
}

type RegisterAttendanceRequest struct {
	ClassID         int `json:"class_id"`
	AttendanceCount int `json:"attendance_count"`
}

func (h *AttendanceHandler) RegisterAttendanceHandler(c *gin.Context) {
	var req RegisterAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}
	if activeEventID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active event found"})
		return
	}

	// Get user from context
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	user := userCtx.(*models.User)

	// Check if user is admin or root
	isAdmin := false
	for _, role := range user.Roles {
		if role.Name == "admin" || role.Name == "root" {
			isAdmin = true
			break
		}
	}

	// If user is admin and has a class_name_rep role, restrict to that class only
	if isAdmin {
		managedClass, err := h.classRepo.GetClassByRepRole(user.ID, activeEventID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check managed class"})
			return
		}
		// If managedClass exists, user can only register attendance for that class
		if managedClass != nil {
			if managedClass.ID != req.ClassID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You are only authorized to register attendance for your managed class"})
				return
			}
		}
	}

	class, err := h.classRepo.GetClassByID(req.ClassID)
	if err != nil || class == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get class info"})
		return
	}

	if class.EventID != nil && *class.EventID != activeEventID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Class does not belong to the active event"})
		return
	}

	if req.AttendanceCount > class.StudentCount {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Attendance count (%d) cannot exceed student count (%d)", req.AttendanceCount, class.StudentCount)})
		return
	}

	points, err := h.classRepo.UpdateAttendance(req.ClassID, activeEventID, req.AttendanceCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to register attendance and calculate points: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Successfully registered attendance for class %s. Points awarded: %d", class.Name, points)})
}
