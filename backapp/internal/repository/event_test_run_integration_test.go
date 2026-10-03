package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func eventTestRunDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("SPORTEASE_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("MySQL統合テストには SPORTEASE_TEST_MYSQL_DSN が必要です")
	}
	cfg, err := mysql.ParseDSN(dsn)
	require.NoError(t, err)
	cfg.DBName = ""
	cfg.MultiStatements = true
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	require.NoError(t, err)
	t.Cleanup(func() { admin.Close() })

	dbName := fmt.Sprintf("sportease_test_run_test_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE DATABASE " + dbName)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = admin.Exec("DROP DATABASE " + dbName) })

	cfg.DBName = dbName
	db, err := sql.Open("mysql", cfg.FormatDSN())
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	_, err = db.Exec(`
		CREATE TABLE events (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(100) NOT NULL,
			status VARCHAR(20) NOT NULL
		);
		CREATE TABLE users (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(100) NOT NULL,
			normalized_name VARCHAR(100) GENERATED ALWAYS AS (LOWER(name)) VIRTUAL
		);
		CREATE TABLE schema_migrations (version BIGINT NOT NULL, dirty BOOLEAN NOT NULL);
		CREATE TABLE event_test_runs (
			id TINYINT PRIMARY KEY,
			event_id INT NOT NULL,
			started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE event_test_run_tables (
			table_name VARCHAR(64) PRIMARY KEY,
			auto_increment_value BIGINT UNSIGNED NULL
		);
		INSERT INTO events (name, status) VALUES ('original event', 'preparing');
		INSERT INTO users (name) VALUES ('original user');
		INSERT INTO schema_migrations VALUES (29, FALSE);
	`)
	require.NoError(t, err)
	return db
}

func TestEventTestRunSnapshotRestoreMySQL(t *testing.T) {
	db := eventTestRunDB(t)
	repo := NewEventTestRunRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.Begin(ctx, 1))
	_, err := db.Exec(`
		UPDATE events SET name = 'test event', status = 'testing' WHERE id = 1;
		INSERT INTO users (name) VALUES ('test-only user');
	`)
	require.NoError(t, err)

	require.NoError(t, repo.Restore(ctx, 1))

	var eventName, status string
	require.NoError(t, db.QueryRow("SELECT name, status FROM events WHERE id = 1").Scan(&eventName, &status))
	require.Equal(t, "original event", eventName)
	require.Equal(t, "preparing", status)
	var userCount int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount))
	require.Equal(t, 1, userCount)
	var normalizedName string
	require.NoError(t, db.QueryRow("SELECT normalized_name FROM users WHERE id = 1").Scan(&normalizedName))
	require.Equal(t, "original user", normalizedName)
	var testRunCount int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM event_test_runs").Scan(&testRunCount))
	require.Zero(t, testRunCount)
	var snapshotCount int
	require.NoError(t, db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE() AND LEFT(TABLE_NAME, ?) = ?
	`, len(testRunSnapshotPrefix), testRunSnapshotPrefix).Scan(&snapshotCount))
	require.Zero(t, snapshotCount)
}
