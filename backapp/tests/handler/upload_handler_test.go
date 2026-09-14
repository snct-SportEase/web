package handler_test

import (
	"backapp/internal/handler"
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func multipartUploadContext(t *testing.T, field, filename string, content []byte) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(field, filename)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/upload", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	return c, w
}

func useTemporaryUploadDirectory(t *testing.T) string {
	t.Helper()
	originalDirectory, err := os.Getwd()
	require.NoError(t, err)
	temporaryDirectory := t.TempDir()
	require.NoError(t, os.Chdir(temporaryDirectory))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDirectory))
	})
	return temporaryDirectory
}

func TestImageHandler_UploadValidationAndSave(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("rejects content that does not match its extension", func(t *testing.T) {
		c, w := multipartUploadContext(t, "image", "result.png", []byte("plain text"))
		handler.NewImageHandler().UploadImageHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Unsupported image format")
	})

	t.Run("stores a detected PNG with a generated filename", func(t *testing.T) {
		temporaryDirectory := useTemporaryUploadDirectory(t)
		png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d}
		c, w := multipartUploadContext(t, "image", "result.png", png)
		c.Request.Host = "example.test"
		handler.NewImageHandler().UploadImageHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "/uploads/images/")
		storedFiles, err := filepath.Glob(filepath.Join(temporaryDirectory, "uploads", "images", "*.png"))
		require.NoError(t, err)
		assert.Len(t, storedFiles, 1)
	})
}

func TestPdfHandler_UploadValidationAndSave(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("rejects a PDF signature with a non-PDF extension", func(t *testing.T) {
		c, w := multipartUploadContext(t, "pdf", "rules.txt", []byte("%PDF-1.4\n%%EOF"))
		handler.NewPdfHandler().UploadPdfHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "File extension does not match PDF content")
	})

	t.Run("stores a validated PDF with a generated filename", func(t *testing.T) {
		temporaryDirectory := useTemporaryUploadDirectory(t)
		c, w := multipartUploadContext(t, "pdf", "rules.pdf", []byte("%PDF-1.4\n%%EOF"))
		c.Request.Host = "example.test"
		handler.NewPdfHandler().UploadPdfHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "/uploads/pdfs/")
		storedFiles, err := filepath.Glob(filepath.Join(temporaryDirectory, "uploads", "pdfs", "*.pdf"))
		require.NoError(t, err)
		assert.Len(t, storedFiles, 1)
	})
}
