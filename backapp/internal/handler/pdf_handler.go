package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PdfHandler struct{}

func NewPdfHandler() *PdfHandler {
	return &PdfHandler{}
}

func (h *PdfHandler) UploadPdfHandler(c *gin.Context) {
	file, err := c.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PDF upload failed"})
		return
	}

	// Generate a unique filename
	ext := filepath.Ext(file.Filename)
	if ext != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is not a PDF"})
		return
	}
	newFilename := uuid.New().String() + ext
	
	// The directory to save uploaded files, this path is inside the container
	uploadDir := "/app/uploads/pdfs"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Save the file
	dst := filepath.Join(uploadDir, newFilename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save PDF"})
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
	url := fmt.Sprintf("%s://%s/uploads/pdfs/%s", scheme, host, newFilename)

	c.JSON(http.StatusOK, gin.H{"url": url})
}
