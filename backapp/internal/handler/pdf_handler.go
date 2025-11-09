package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"backapp/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxPdfUploadSize = 10 * 1024 * 1024 // 10MB

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

	if file.Size > maxPdfUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PDF file is too large"})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read PDF"})
		return
	}
	defer openedFile.Close()

	sniff := make([]byte, 512)
	n, err := openedFile.Read(sniff)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to inspect PDF"})
		return
	}

	mimeType := http.DetectContentType(sniff[:n])
	if mimeType != "application/pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is not a PDF"})
		return
	}

	ext := filepath.Ext(file.Filename)
	if strings.ToLower(ext) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File extension does not match PDF content"})
		return
	}

	newFilename := uuid.New().String() + ".pdf"

	uploadDir := "/app/uploads/pdfs"
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Save the file
	dst := filepath.Join(uploadDir, newFilename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save PDF"})
		return
	}

	isSecure := middleware.IsRequestSecure(c.Request)
	scheme := "http"
	if isSecure {
		scheme = "https"
	}

	// Construct the full URL
	host := c.Request.Host
	url := fmt.Sprintf("%s://%s/uploads/pdfs/%s", scheme, host, newFilename)

	c.JSON(http.StatusOK, gin.H{"url": url})
}
