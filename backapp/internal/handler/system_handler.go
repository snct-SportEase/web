package handler

import (
	"backapp/internal/config"
	"fmt"
	"net/http"
	"os/exec"
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
	cmd.Env = append(cmd.Env, fmt.Sprintf("MYSQL_PWD=%s", h.cfg.DBPassword))

	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=database_dump_%s.sql", time.Now().Format("20060102_150405")))
	c.Writer.Header().Set("Content-Type", "application/sql")

	cmd.Stdout = c.Writer

	// Print stderr if any error occurs
	if err := cmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate database dump", "details": err.Error()})
		return
	}
}
