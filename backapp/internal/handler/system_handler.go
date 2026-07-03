package handler

import (
	"archive/zip"
	"backapp/internal/config"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

type SystemHandler struct {
	cfg            *config.Config
	dbDumpExporter DBDumpExporter
}

func NewSystemHandler(cfg *config.Config) *SystemHandler {
	return &SystemHandler{
		cfg:            cfg,
		dbDumpExporter: exportDBDumpWithGoDriver,
	}
}

type DBDumpExporter func(context.Context, *config.Config, io.Writer) error

func NewSystemHandlerWithDBDumpExporter(cfg *config.Config, exporter DBDumpExporter) *SystemHandler {
	handler := NewSystemHandler(cfg)
	handler.dbDumpExporter = exporter
	return handler
}

func (h *SystemHandler) ExportDBDump(c *gin.Context) {
	var dump bytes.Buffer

	if err := h.dbDumpExporter(c.Request.Context(), h.cfg, &dump); err != nil {
		log.Printf("ExportDBDump error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate database dump"})
		return
	}

	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=database_dump_%s.sql", time.Now().Format("20060102_150405")))
	c.Writer.Header().Set("Content-Type", "application/sql")
	if _, err := c.Writer.Write(dump.Bytes()); err != nil {
		log.Printf("ExportDBDump write error: %v", err)
	}
}

func exportDBDumpWithGoDriver(ctx context.Context, cfg *config.Config, writer io.Writer) error {
	mysqlConfig := mysql.NewConfig()
	mysqlConfig.User = cfg.DBUser
	mysqlConfig.Passwd = cfg.DBPassword
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = fmt.Sprintf("%s:%s", cfg.DBHost, cfg.DBPort)
	mysqlConfig.DBName = cfg.DBName
	mysqlConfig.AllowNativePasswords = true
	mysqlConfig.ParseTime = false
	mysqlConfig.Collation = "utf8mb4_unicode_ci"
	mysqlConfig.Params = map[string]string{"charset": "utf8mb4"}

	db, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	return writeSQLDump(ctx, db, cfg.DBName, writer)
}

type dumpObject struct {
	Name string
	Type string
}

func writeSQLDump(ctx context.Context, db *sql.DB, dbName string, writer io.Writer) error {
	objects, err := listDumpObjects(ctx, db, dbName)
	if err != nil {
		return err
	}

	baseTables := make([]dumpObject, 0, len(objects))
	views := make([]dumpObject, 0)
	for _, object := range objects {
		if object.Type == "VIEW" {
			views = append(views, object)
			continue
		}
		baseTables = append(baseTables, object)
	}

	if _, err := fmt.Fprintf(writer, "-- SportEase database dump\n-- Generated at %s\n\n", time.Now().UTC().Format(time.RFC3339)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "SET FOREIGN_KEY_CHECKS=0;"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}

	for _, view := range views {
		if _, err := fmt.Fprintf(writer, "DROP VIEW IF EXISTS %s;\n", quoteIdentifier(view.Name)); err != nil {
			return err
		}
	}
	for _, table := range baseTables {
		if _, err := fmt.Fprintf(writer, "DROP TABLE IF EXISTS %s;\n", quoteIdentifier(table.Name)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}

	for _, table := range baseTables {
		createSQL, err := showCreateObject(ctx, db, table.Name)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(writer, "%s;\n\n", createSQL); err != nil {
			return err
		}
		if err := writeTableRows(ctx, db, table.Name, writer); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(writer); err != nil {
			return err
		}
	}

	for _, view := range views {
		createSQL, err := showCreateObject(ctx, db, view.Name)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(writer, "%s;\n\n", createSQL); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(writer, "SET FOREIGN_KEY_CHECKS=1;")
	return err
}

func listDumpObjects(ctx context.Context, db *sql.DB, dbName string) ([]dumpObject, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT table_name, table_type
		FROM information_schema.tables
		WHERE table_schema = ?
		ORDER BY CASE WHEN table_type = 'VIEW' THEN 1 ELSE 0 END, table_name
	`, dbName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	objects := make([]dumpObject, 0)
	for rows.Next() {
		var object dumpObject
		if err := rows.Scan(&object.Name, &object.Type); err != nil {
			return nil, err
		}
		objects = append(objects, object)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return objects, nil
}

func showCreateObject(ctx context.Context, db *sql.DB, name string) (string, error) {
	rows, err := db.QueryContext(ctx, "SHOW CREATE TABLE "+quoteIdentifier(name))
	if err != nil {
		return "", err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}
	if len(columns) < 2 {
		return "", fmt.Errorf("unexpected SHOW CREATE result for %s", name)
	}
	if !rows.Next() {
		return "", sql.ErrNoRows
	}

	values := make([]sql.RawBytes, len(columns))
	dest := make([]interface{}, len(columns))
	for i := range values {
		dest[i] = &values[i]
	}
	if err := rows.Scan(dest...); err != nil {
		return "", err
	}
	return string(values[1]), rows.Err()
}

func writeTableRows(ctx context.Context, db *sql.DB, tableName string, writer io.Writer) error {
	rows, err := db.QueryContext(ctx, "SELECT * FROM "+quoteIdentifier(tableName))
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	if len(columns) == 0 {
		return nil
	}

	quotedColumns := make([]string, len(columns))
	for i, column := range columns {
		quotedColumns[i] = quoteIdentifier(column)
	}

	values := make([]sql.RawBytes, len(columns))
	dest := make([]interface{}, len(columns))
	for i := range values {
		dest[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			return err
		}

		literals := make([]string, len(values))
		for i, value := range values {
			literals[i] = sqlLiteral(value)
		}

		if _, err := fmt.Fprintf(
			writer,
			"INSERT INTO %s (%s) VALUES (%s);\n",
			quoteIdentifier(tableName),
			strings.Join(quotedColumns, ", "),
			strings.Join(literals, ", "),
		); err != nil {
			return err
		}
	}
	return rows.Err()
}

func quoteIdentifier(identifier string) string {
	return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
}

func sqlLiteral(value []byte) string {
	if value == nil {
		return "NULL"
	}
	escaped := strings.NewReplacer(
		"\\", "\\\\",
		"'", "\\'",
		"\x00", "\\0",
		"\n", "\\n",
		"\r", "\\r",
		"\x1a", "\\Z",
	).Replace(string(value))
	return "'" + escaped + "'"
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
