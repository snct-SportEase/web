package handler_test

import (
	"archive/zip"
	"backapp/internal/config"
	"backapp/internal/handler"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestExportDBDump(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		DBHost:     "127.0.0.1",
		DBPort:     "3307",
		DBUser:     "dump_user",
		DBPassword: "test_password",
		DBName:     "sportease",
	}

	systemHandler := handler.NewSystemHandlerWithDBDumpExporter(cfg, func(_ context.Context, gotCfg *config.Config, writer io.Writer) error {
		assert.Equal(t, cfg, gotCfg)
		_, err := io.WriteString(writer, "CREATE TABLE users (id int);\nINSERT INTO users VALUES (1);\n")
		return err
	})

	router := gin.Default()
	router.GET("/api/root/db/export", systemHandler.ExportDBDump)

	req, _ := http.NewRequest(http.MethodGet, "/api/root/db/export", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	contentDisposition := w.Header().Get("Content-Disposition")
	assert.Contains(t, contentDisposition, "attachment; filename=database_dump_")
	assert.Contains(t, contentDisposition, ".sql")

	contentType := w.Header().Get("Content-Type")
	assert.Equal(t, "application/sql", contentType)

	assert.Contains(t, w.Body.String(), "CREATE TABLE users")
	assert.Contains(t, w.Body.String(), "INSERT INTO users VALUES")
}

func TestExportDBDumpCommandFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	systemHandler := handler.NewSystemHandlerWithDBDumpExporter(&config.Config{
		DBHost:     "127.0.0.1",
		DBPort:     "3307",
		DBUser:     "dump_user",
		DBPassword: "test_password",
		DBName:     "sportease",
	}, func(context.Context, *config.Config, io.Writer) error {
		return errors.New("dump failed")
	})

	router := gin.Default()
	router.GET("/api/root/db/export", systemHandler.ExportDBDump)

	req, _ := http.NewRequest(http.MethodGet, "/api/root/db/export", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to generate database dump")
}

func TestExportUploadsDump(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Chdir(t.TempDir())

	assert.NoError(t, os.MkdirAll(filepath.Join("uploads", "pdfs"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join("uploads", "pdfs", "guide.pdf"), []byte("pdf-data"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join("uploads", "image.png"), []byte("image-data"), 0o644))

	systemHandler := handler.NewSystemHandler(&config.Config{})

	router := gin.Default()
	router.GET("/api/root/uploads/export", systemHandler.ExportUploadsDump)

	req, _ := http.NewRequest(http.MethodGet, "/api/root/uploads/export", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment; filename=uploads_dump_")
	assert.Contains(t, w.Header().Get("Content-Disposition"), ".zip")
	assert.Equal(t, "application/zip", w.Header().Get("Content-Type"))

	reader, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	assert.NoError(t, err)

	entries := make(map[string]string)
	for _, file := range reader.File {
		opened, err := file.Open()
		assert.NoError(t, err)

		body := new(bytes.Buffer)
		_, err = body.ReadFrom(opened)
		assert.NoError(t, err)
		assert.NoError(t, opened.Close())

		entries[file.Name] = body.String()
	}

	assert.Equal(t, map[string]string{
		"image.png":      "image-data",
		"pdfs/guide.pdf": "pdf-data",
	}, entries)
}

func TestExportUploadsDumpMissingDirectory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Chdir(t.TempDir())

	systemHandler := handler.NewSystemHandler(&config.Config{})

	router := gin.Default()
	router.GET("/api/root/uploads/export", systemHandler.ExportUploadsDump)

	req, _ := http.NewRequest(http.MethodGet, "/api/root/uploads/export", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Uploads directory not found")
}
