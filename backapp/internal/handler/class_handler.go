package handler

import (
	"backapp/internal/repository"
	"net/http"

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