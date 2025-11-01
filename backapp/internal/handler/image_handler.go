package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ImageHandler struct{}

func NewImageHandler() *ImageHandler {
	return &ImageHandler{}
}

func (h *ImageHandler) UploadImageHandler(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image upload failed"})
		return
	}

	// Generate a unique filename
	ext := filepath.Ext(file.Filename)
	newFilename := uuid.New().String() + ext
	
	// The directory to save uploaded files, this path is inside the container
	uploadDir := "/app/uploads/images"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Save the file
	dst := filepath.Join(uploadDir, newFilename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	// Determine scheme
	scheme := "http"
	if proto := c.GetHeader("X-Forwarded-Proto"); proto == "https" {
		scheme = "https"
	} else if c.Request.TLS != nil {
		scheme = "https"
	}

	// Construct the full URL
	host := c.Request.Host
	url := fmt.Sprintf("%s://%s/uploads/images/%s", scheme, host, newFilename)

	c.JSON(http.StatusOK, gin.H{"url": url})
}