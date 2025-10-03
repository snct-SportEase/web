package handler

import (
	"backapp/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ClassHandler struct {
	classRepo repository.ClassRepository
}

func NewClassHandler(classRepo repository.ClassRepository) *ClassHandler {
	return &ClassHandler{classRepo: classRepo}
}

func (h *ClassHandler) GetAllClasses(c *gin.Context) {
	classes, err := h.classRepo.GetAllClasses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, classes)
}
