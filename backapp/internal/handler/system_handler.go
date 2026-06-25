package handler

import (
	"archive/zip"
	"backapp/internal/config"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	cfg *config.Config
}

func NewSystemHandler(cfg *config.Config) *SystemHandler {
	return &SystemHandler{
		cfg: cfg,
	}
}

func (h *SystemHandler) ExportDBDump(c *gin.Context) {
	// Construct mysqldump command
	// Note: It's generally safer to use a configuration file for the password or pass it via MYSQL_PWD env var
	// but using PWD env var is supported by mysqldump.
	cmd := exec.Command(
		"mysqldump",
		"-h", h.cfg.DBHost,
		"-P", h.cfg.DBPort,
		"-u", h.cfg.DBUser,
		"--default-character-set=utf8mb4",
		"--single-transaction",
		"--routines",
		"--triggers",
		h.cfg.DBName,
	)

	// Pass password via environment variable
	cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", h.cfg.DBPassword))

	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=database_dump_%s.sql", time.Now().Format("20060102_150405")))
	c.Writer.Header().Set("Content-Type", "application/sql")

	cmd.Stdout = c.Writer

	// Print stderr if any error occurs
	if err := cmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate database dump"})
		return
	}
}

func (h *SystemHandler) ExportUploadsDump(c *gin.Context) {
	uploadsDir := "./uploads"

	stat, err := os.Stat(uploadsDir)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Uploads directory not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access uploads directory"})
		return
	}
	if !stat.IsDir() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Uploads path is not a directory"})
		return
	}

	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=uploads_dump_%s.zip", time.Now().Format("20060102_150405")))
	c.Writer.Header().Set("Content-Type", "application/zip")

	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()

	if err := filepath.WalkDir(uploadsDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		relativePath, err := filepath.Rel(uploadsDir, path)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relativePath)
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		if _, err = io.Copy(writer, file); err != nil {
			file.Close()
			return err
		}
		return file.Close()
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate uploads dump"})
		return
	}
}
