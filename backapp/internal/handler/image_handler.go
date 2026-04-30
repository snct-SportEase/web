package handler

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxImageUploadSize = 5 * 1024 * 1024 // 5MB

var allowedImageMIMEs = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

type ImageHandler struct{}

const imageUploadDir = "uploads/images"

func NewImageHandler() *ImageHandler {
	return &ImageHandler{}
}

func (h *ImageHandler) UploadImageHandler(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image upload failed"})
		return
	}

	if file.Size > maxImageUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is too large"})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read image"})
		return
	}
	defer openedFile.Close()

	sniff := make([]byte, 512)
	n, err := openedFile.Read(sniff)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to inspect image"})
		return
	}

	mimeType := http.DetectContentType(sniff[:n])
	allowedExt, ok := allowedImageMIMEs[mimeType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported image format"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = allowedExt
	} else if ext == ".jpeg" && allowedExt == ".jpg" {
		ext = ".jpg"
	} else if ext != allowedExt {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File extension does not match image content"})
		return
	}

	newFilename := uuid.New().String() + ext

	uploadDir := imageUploadDir
	if err := os.MkdirAll(uploadDir, 0750); err != nil {
		log.Printf("UploadImageHandler mkdir failed: dir=%s err=%v", uploadDir, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Save the file
	dst := filepath.Join(uploadDir, newFilename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		log.Printf("UploadImageHandler save failed: dst=%s err=%v", dst, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	url := buildPublicUploadURL(c.Request, "/uploads/images/"+newFilename)

	c.JSON(http.StatusOK, gin.H{"url": url})
}
