package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"backapp/internal/models"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	raceUserID  = "22222222-2222-4222-8222-222222222222"
	raceSHA     = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	raceSHAAlt  = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	raceTrials  = 8
	raceWorkers = 12
)

func typingImportRaceDB(t *testing.T, withUnique bool) *sql.DB {
	t.Helper()
	dsn := os.Getenv("SPORTEASE_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("MySQL統合テストには SPORTEASE_TEST_MYSQL_DSN が必要です")
	}
	cfg, err := mysql.ParseDSN(dsn)
	require.NoError(t, err)
	cfg.DBName = ""
	cfg.MultiStatements = true
	cfg.ParseTime = true
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	require.NoError(t, err)
	t.Cleanup(func() { admin.Close() })
	require.NoError(t, admin.Ping())

	dbName := fmt.Sprintf("sportease_typing_race_%d", time.Now().UnixNano())
	_, err = admin.Exec("CREATE DATABASE `" + dbName + "`")
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = admin.Exec("DROP DATABASE `" + dbName + "`") })

	cfg.DBName = dbName
	db, err := sql.Open("mysql", cfg.FormatDSN())
	require.NoError(t, err)
	db.SetMaxOpenConns(raceWorkers * 2)
	db.SetMaxIdleConns(raceWorkers * 2)
	t.Cleanup(func() { db.Close() })

	unique := ""
	if withUnique {
		unique = ",\n        UNIQUE KEY uq_noon_game_typing_import_session_export_sha (session_id, export_id, sha256)"
	}
	_, err = db.Exec(`
		CREATE TABLE noon_game_sessions (
			id INT PRIMARY KEY AUTO_INCREMENT
		) ENGINE=InnoDB;
		CREATE TABLE noon_game_points (
			id INT PRIMARY KEY AUTO_INCREMENT,
			session_id INT NOT NULL,
			match_id INT NULL,
			class_id INT NOT NULL,
			points INT NOT NULL,
			reason VARCHAR(255) NULL,
			source ENUM('result', 'manual', 'typing_system') NOT NULL,
			created_by CHAR(36) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB;
		CREATE TABLE noon_game_typing_system_imports (
			id INT PRIMARY KEY AUTO_INCREMENT,
			session_id INT NOT NULL,
			export_id CHAR(36) NOT NULL,
			sha256 CHAR(64) NOT NULL,
			status ENUM('success', 'failed') NOT NULL,
			action ENUM('import', 'replace') NOT NULL,
			replaced_export_id CHAR(36) NULL,
			requested_by CHAR(36) NOT NULL,
			requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			filename VARCHAR(255) NULL,
			payload_size INT NOT NULL,
			results JSON NULL,
			message VARCHAR(255) NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_noon_game_typing_import_session FOREIGN KEY (session_id) REFERENCES noon_game_sessions(id) ON DELETE CASCADE` + unique + `
		) ENGINE=InnoDB;
	`)
	require.NoError(t, err)
	return db
}

func insertRaceSession(t *testing.T, db *sql.DB) int {
	t.Helper()
	res, err := db.Exec(`INSERT INTO noon_game_sessions () VALUES ()`)
	require.NoError(t, err)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return int(id)
}

func raceHistory(sessionID int, exportID, sha string) *models.NoonGameTypingSystemImportRecord {
	filename := "typing.json"
	return &models.NoonGameTypingSystemImportRecord{
		SessionID:   sessionID,
		ExportID:    exportID,
		SHA256:      sha,
		Status:      "success",
		Action:      "import",
		RequestedBy: raceUserID,
		Filename:    &filename,
		PayloadSize: 128,
		IsActive:    true,
	}
}

func racePoints(sessionID int, values []int) []*models.NoonGamePoint {
	points := make([]*models.NoonGamePoint, len(values))
	for i, value := range values {
		points[i] = &models.NoonGamePoint{
			SessionID: sessionID,
			ClassID:   i + 1,
			Points:    value,
			Source:    "typing_system",
			CreatedBy: raceUserID,
		}
	}
	return points
}

func raceApply(repo NoonGameRepository, n int, call func(i int) error) []error {
	errs := make([]error, n)
	ready := make(chan struct{}, n)
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			ready <- struct{}{}
			<-start
			errs[i] = call(i)
		}(i)
	}
	for i := 0; i < n; i++ {
		<-ready
	}
	close(start)
	wg.Wait()
	return errs
}

func countQuery(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var n int
	require.NoError(t, db.QueryRow(query, args...).Scan(&n))
	return n
}

func assertOneImport(t *testing.T, db *sql.DB, sessionID, wantSum int) {
	t.Helper()
	require.Equal(t, 6, countQuery(t, db, `SELECT COUNT(*) FROM noon_game_points WHERE session_id = ?`, sessionID))
	require.Equal(t, wantSum, countQuery(t, db, `SELECT CAST(COALESCE(SUM(points), 0) AS SIGNED) FROM noon_game_points WHERE session_id = ?`, sessionID))
	require.Equal(t, 1, countQuery(t, db, `SELECT COUNT(*) FROM noon_game_typing_system_imports WHERE session_id = ? AND status = 'success'`, sessionID))
	require.Equal(t, 1, countQuery(t, db, `SELECT COUNT(*) FROM noon_game_typing_system_imports WHERE session_id = ? AND is_active = TRUE`, sessionID))
}

func TestTypingImportRaceSameFileDoesNotDoublePoints(t *testing.T) {
	db := typingImportRaceDB(t, true)
	repo := NewNoonGameRepository(db)
	values := []int{40, 30, 25, 20, 15, 10}

	for trial := 0; trial < raceTrials; trial++ {
		sessionID := insertRaceSession(t, db)
		exportID := uuid.NewString()
		errs := raceApply(repo, raceWorkers, func(i int) error {
			return repo.ApplyTypingSystemResultImport(sessionID, racePoints(sessionID, values), false, raceHistory(sessionID, exportID, raceSHA))
		})
		successes := 0
		for _, err := range errs {
			switch {
			case err == nil:
				successes++
			case errors.Is(err, ErrTypingSystemAlreadyImported):
			default:
				t.Fatalf("trial %d: unexpected error: %v", trial, err)
			}
		}
		require.Equal(t, 1, successes, "trial %d", trial)
		assertOneImport(t, db, sessionID, 140)

		err := repo.ApplyTypingSystemResultImport(sessionID, racePoints(sessionID, values), true, raceHistory(sessionID, exportID, raceSHA))
		require.ErrorIs(t, err, ErrTypingSystemAlreadyImported)
		assertOneImport(t, db, sessionID, 140)
	}
}

func TestTypingImportRaceDifferentExportsDoNotStack(t *testing.T) {
	db := typingImportRaceDB(t, true)
	repo := NewNoonGameRepository(db)
	first := []int{40, 30, 25, 20, 15, 10}
	second := []int{5, 5, 5, 5, 5, 5}

	for trial := 0; trial < raceTrials; trial++ {
		sessionID := insertRaceSession(t, db)
		exportA := uuid.NewString()
		exportB := uuid.NewString()
		errs := raceApply(repo, raceWorkers, func(i int) error {
			if i%2 == 0 {
				return repo.ApplyTypingSystemResultImport(sessionID, racePoints(sessionID, first), false, raceHistory(sessionID, exportA, raceSHA))
			}
			return repo.ApplyTypingSystemResultImport(sessionID, racePoints(sessionID, second), false, raceHistory(sessionID, exportB, raceSHAAlt))
		})
		successes := 0
		for _, err := range errs {
			switch {
			case err == nil:
				successes++
			case errors.Is(err, ErrTypingSystemAlreadyImported), errors.Is(err, ErrTypingSystemActiveImportExists):
			default:
				t.Fatalf("trial %d: unexpected error: %v", trial, err)
			}
		}
		require.Equal(t, 1, successes, "trial %d", trial)
		require.Equal(t, 6, countQuery(t, db, `SELECT COUNT(*) FROM noon_game_points WHERE session_id = ?`, sessionID))
		sum := countQuery(t, db, `SELECT CAST(COALESCE(SUM(points), 0) AS SIGNED) FROM noon_game_points WHERE session_id = ?`, sessionID)
		require.Contains(t, []int{140, 30}, sum, "trial %d stacked points", trial)
		require.Equal(t, 1, countQuery(t, db, `SELECT COUNT(*) FROM noon_game_typing_system_imports WHERE session_id = ?`, sessionID))
	}
}

func TestTypingImportRaceDifferentContentDoesNotStack(t *testing.T) {
	db := typingImportRaceDB(t, true)
	repo := NewNoonGameRepository(db)
	values := []int{40, 30, 25, 20, 15, 10}

	for trial := 0; trial < raceTrials; trial++ {
		sessionID := insertRaceSession(t, db)
		exportID := uuid.NewString()
		errs := raceApply(repo, raceWorkers, func(i int) error {
			sha := raceSHA
			if i%2 == 1 {
				sha = raceSHAAlt
			}
			return repo.ApplyTypingSystemResultImport(sessionID, racePoints(sessionID, values), false, raceHistory(sessionID, exportID, sha))
		})
		successes := 0
		for _, err := range errs {
			switch {
			case err == nil:
				successes++
			case errors.Is(err, ErrTypingSystemAlreadyImported), errors.Is(err, ErrTypingSystemExportConflict):
			default:
				t.Fatalf("trial %d: unexpected error: %v", trial, err)
			}
		}
		require.Equal(t, 1, successes, "trial %d", trial)
		assertOneImport(t, db, sessionID, 140)
	}
}

func TestTypingImportRaceConcurrentReplaceKeepsSingleScore(t *testing.T) {
	db := typingImportRaceDB(t, true)
	repo := NewNoonGameRepository(db)
	original := []int{1, 1, 1, 1, 1, 1}
	replacementA := []int{40, 30, 25, 20, 15, 10}
	replacementB := []int{8, 8, 8, 8, 8, 8}

	for trial := 0; trial < raceTrials; trial++ {
		sessionID := insertRaceSession(t, db)
		require.NoError(t, repo.ApplyTypingSystemResultImport(sessionID, racePoints(sessionID, original), false, raceHistory(sessionID, uuid.NewString(), raceSHA)))

		exportA := uuid.NewString()
		exportB := uuid.NewString()
		errs := raceApply(repo, raceWorkers, func(i int) error {
			if i%2 == 0 {
				return repo.ApplyTypingSystemResultImport(sessionID, racePoints(sessionID, replacementA), true, raceHistory(sessionID, exportA, raceSHAAlt))
			}
			return repo.ApplyTypingSystemResultImport(sessionID, racePoints(sessionID, replacementB), true, raceHistory(sessionID, exportB, raceSHA))
		})
		successes := 0
		for _, err := range errs {
			switch {
			case err == nil:
				successes++
			case errors.Is(err, ErrTypingSystemAlreadyImported), errors.Is(err, ErrTypingSystemActiveImportExists):
			default:
				t.Fatalf("trial %d: unexpected error: %v", trial, err)
			}
		}
		require.GreaterOrEqual(t, successes, 1, "trial %d", trial)
		require.Equal(t, 6, countQuery(t, db, `SELECT COUNT(*) FROM noon_game_points WHERE session_id = ?`, sessionID), "trial %d", trial)
		sum := countQuery(t, db, `SELECT CAST(COALESCE(SUM(points), 0) AS SIGNED) FROM noon_game_points WHERE session_id = ?`, sessionID)
		require.Contains(t, []int{140, 48}, sum, "trial %d", trial)
		require.Equal(t, 1, countQuery(t, db, `SELECT COUNT(*) FROM noon_game_typing_system_imports WHERE session_id = ? AND is_active = TRUE`, sessionID))
	}
}

func TestTypingImportFailedHistoryDoesNotLeavePartialPoints(t *testing.T) {
	db := typingImportRaceDB(t, true)
	repo := NewNoonGameRepository(db)
	sessionID := insertRaceSession(t, db)
	exportID := uuid.NewString()
	_, err := db.Exec(`
		INSERT INTO noon_game_typing_system_imports
			(session_id, export_id, sha256, status, action, requested_by, payload_size, is_active)
		VALUES (?, ?, ?, 'failed', 'import', ?, 4, TRUE)
	`, sessionID, exportID, raceSHA, raceUserID)
	require.NoError(t, err)

	err = repo.ApplyTypingSystemResultImport(sessionID, racePoints(sessionID, []int{40, 30, 25, 20, 15, 10}), false, raceHistory(sessionID, exportID, raceSHA))
	require.ErrorIs(t, err, ErrTypingSystemAlreadyImported)
	require.Equal(t, 0, countQuery(t, db, `SELECT COUNT(*) FROM noon_game_points WHERE session_id = ?`, sessionID))
	require.Equal(t, 1, countQuery(t, db, `SELECT COUNT(*) FROM noon_game_typing_system_imports WHERE session_id = ?`, sessionID))
}

func TestTypingImportMigrationKeepsActiveDuplicateAndLeavesPoints(t *testing.T) {
	db := typingImportRaceDB(t, false)
	sessionID := insertRaceSession(t, db)
	exportID := uuid.NewString()
	for _, active := range []bool{false, true, false} {
		_, err := db.Exec(`
			INSERT INTO noon_game_typing_system_imports
				(session_id, export_id, sha256, status, action, requested_by, payload_size, is_active)
			VALUES (?, ?, ?, 'success', 'import', ?, 4, ?)
		`, sessionID, exportID, raceSHA, raceUserID, active)
		require.NoError(t, err)
	}
	_, err := db.Exec(`
		INSERT INTO noon_game_points (session_id, class_id, points, source, created_by)
		VALUES (?, 1, 40, 'typing_system', ?), (?, 1, 40, 'typing_system', ?)
	`, sessionID, raceUserID, sessionID, raceUserID)
	require.NoError(t, err)

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	migration, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "../../db/migrations/0031_unique_typing_system_import.up.sql"))
	require.NoError(t, err)
	_, err = db.Exec(string(migration))
	require.NoError(t, err)

	require.Equal(t, 1, countQuery(t, db, `SELECT COUNT(*) FROM noon_game_typing_system_imports WHERE session_id = ?`, sessionID))
	require.Equal(t, 1, countQuery(t, db, `SELECT COUNT(*) FROM noon_game_typing_system_imports WHERE session_id = ? AND is_active = TRUE`, sessionID))
	require.Equal(t, 2, countQuery(t, db, `SELECT COUNT(*) FROM noon_game_points WHERE session_id = ?`, sessionID), "doubled points stay for manual repair")

	_, err = db.Exec(`
		INSERT INTO noon_game_typing_system_imports
			(session_id, export_id, sha256, status, action, requested_by, payload_size, is_active)
		VALUES (?, ?, ?, 'success', 'import', ?, 4, TRUE)
	`, sessionID, exportID, raceSHA, raceUserID)
	var mysqlErr *mysql.MySQLError
	require.ErrorAs(t, err, &mysqlErr)
	require.Equal(t, uint16(1062), mysqlErr.Number)
}
