package handler_test

import (
	"archive/zip"
	"backapp/internal/config"
	"backapp/internal/handler"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestExportDBDump(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tempDir := t.TempDir()
	argsFile := filepath.Join(tempDir, "mysqldump.args")
	passwordFile := filepath.Join(tempDir, "mysqldump.password")
	fakeBinDir := filepath.Join(tempDir, "bin")

	assert.NoError(t, os.MkdirAll(fakeBinDir, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(fakeBinDir, "mysqldump"), []byte(`#!/bin/sh
printf '%s\n' "$@" > "$MYSQLDUMP_ARGS_FILE"
printf '%s' "$MYSQL_PWD" > "$MYSQLDUMP_PASSWORD_FILE"
cat <<'SQL'
CREATE TABLE users (id int);
INSERT INTO users VALUES (1);
SQL
`), 0o755))

	t.Setenv("MYSQLDUMP_ARGS_FILE", argsFile)
	t.Setenv("MYSQLDUMP_PASSWORD_FILE", passwordFile)
	t.Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	cfg := &config.Config{
		DBHost:     "127.0.0.1",
		DBPort:     "3307",
		DBUser:     "dump_user",
		DBPassword: "test_password",
		DBName:     "sportease",
	}

	systemHandler := handler.NewSystemHandler(cfg)

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

	argsBytes, err := os.ReadFile(argsFile)
	assert.NoError(t, err)
	args := strings.Fields(string(argsBytes))
	assert.Equal(t, []string{
		"-h", "127.0.0.1",
		"-P", "3307",
		"-u", "dump_user",
		"--default-character-set=utf8mb4",
		"--single-transaction",
		"--routines",
		"--triggers",
		"sportease",
	}, args)
	assert.NotContains(t, args, "--no-data")
	assert.NotContains(t, args, "--no-create-info")

	passwordBytes, err := os.ReadFile(passwordFile)
	assert.NoError(t, err)
	assert.Equal(t, "test_password", string(passwordBytes))
}

func TestExportDBDumpCommandFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tempDir := t.TempDir()
	fakeBinDir := filepath.Join(tempDir, "bin")

	assert.NoError(t, os.MkdirAll(fakeBinDir, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(fakeBinDir, "mysqldump"), []byte(`#!/bin/sh
exit 7
`), 0o755))
	t.Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	systemHandler := handler.NewSystemHandler(&config.Config{
		DBHost:     "127.0.0.1",
		DBPort:     "3307",
		DBUser:     "dump_user",
		DBPassword: "test_password",
		DBName:     "sportease",
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
