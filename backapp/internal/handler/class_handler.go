package handler

import (
	"backapp/internal/repository"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ClassHandler struct {
	classRepo repository.ClassRepository
	eventRepo repository.EventRepository
}

func NewClassHandler(classRepo repository.ClassRepository, eventRepo repository.EventRepository) *ClassHandler {
	return &ClassHandler{
		classRepo: classRepo,
		eventRepo: eventRepo,
	}
}

func (h *ClassHandler) GetAllClasses(c *gin.Context) {
	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}
	fmt.Printf("activeEventID: %d\n", activeEventID)
	if activeEventID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active event found"})
		return
	}

	classes, err := h.classRepo.GetAllClasses(activeEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, classes)
}

type UpdateStudentCountRequest struct {
	ClassID      int `json:"class_id"`
	StudentCount int `json:"student_count"`
}

func (h *ClassHandler) UpdateStudentCountsHandler(c *gin.Context) {
	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}
	if activeEventID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active event found"})
		return
	}

	var req []UpdateStudentCountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	counts := make(map[int]int)
	for _, item := range req {
		counts[item.ClassID] = item.StudentCount
	}

	if err := h.classRepo.UpdateStudentCounts(activeEventID, counts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update student counts: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student counts updated successfully"})
}

func (h *ClassHandler) UpdateStudentCountsFromCSVHandler(c *gin.Context) {
	activeEventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}
	if activeEventID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active event found"})
		return
	}

	file, _, err := c.Request.FormFile("csv")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file not provided"})
		return
	}
	defer file.Close()

	// Get all classes for the active event to map names to IDs
	allClasses, err := h.classRepo.GetAllClasses(activeEventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get classes"})
		return
	}
	classNameToID := make(map[string]int)
	for _, class := range allClasses {
		classNameToID[class.Name] = class.ID
	}

	reader := csv.NewReader(file)
	counts := make(map[int]int)

	// Skip header row
	if _, err := reader.Read(); err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read CSV header: " + err.Error()})
		return
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read CSV record: " + err.Error()})
			return
		}

		if len(record) < 2 {
			continue // Skip empty or invalid rows
		}

		className := record[0]
		studentCountStr := record[1]

		classID, ok := classNameToID[className]
		if !ok {
			// If class name is not found, you might want to log this or handle it
			continue
		}

		studentCount, err := strconv.Atoi(studentCountStr)
		if err != nil {
			// Handle error for invalid number format
			continue
		}

		counts[classID] = studentCount
	}

	if len(counts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid data found in CSV file"})
		return
	}

	if err := h.classRepo.UpdateStudentCounts(activeEventID, counts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update student counts from CSV: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student counts updated successfully from CSV"})
}
