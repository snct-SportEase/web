package repository

import (
	"backapp/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	testRunLockName       = "sportease_event_test_run"
	testRunSnapshotPrefix = "_event_test_snapshot_"
)

var (
	ErrTestRunAlreadyActive = errors.New("an event test run is already active")
	ErrTestRunNotFound      = errors.New("event test run snapshot not found")
	ErrTestRunEventMismatch = errors.New("event test run belongs to another event")
	sqlIdentifierPattern    = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
)

type EventTestRunRepository interface {
	Begin(ctx context.Context, eventID int) error
	Restore(ctx context.Context, eventID int) error
	Discard(ctx context.Context, eventID int) error
	GetStatus(ctx context.Context) (*models.EventTestRunStatus, error)
	ResolveNotificationDelivery(ctx context.Context, policy string) error
}

type eventTestRunRepository struct {
	db                 *sql.DB
	uploadSnapshotRoot string
	uploadDirectories  []string
}

type testRunTable struct {
	name          string
	autoIncrement sql.NullInt64
	columns       []string
}

func NewEventTestRunRepository(db *sql.DB) EventTestRunRepository {
	return &eventTestRunRepository{db: db}
}

func NewEventTestRunRepositoryWithUploads(db *sql.DB, snapshotRoot string, uploadDirectories ...string) EventTestRunRepository {
	return &eventTestRunRepository{
		db:                 db,
		uploadSnapshotRoot: snapshotRoot,
		uploadDirectories:  uploadDirectories,
	}
}

func (r *eventTestRunRepository) Begin(ctx context.Context, eventID int) error {
	return r.withLock(ctx, func(conn *sql.Conn) error {
		var activeEventID int
		err := conn.QueryRowContext(ctx, "SELECT event_id FROM event_test_runs WHERE id = 1").Scan(&activeEventID)
		if err == nil {
			return ErrTestRunAlreadyActive
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if err := r.dropOrphanedSnapshots(ctx, conn); err != nil {
			return err
		}
		if _, err := conn.ExecContext(ctx, "DELETE FROM event_test_run_tables"); err != nil {
			return err
		}
		if _, err := conn.ExecContext(ctx, "INSERT INTO event_test_runs (id, event_id, state) VALUES (1, ?, ?)", eventID, models.EventTestRunStateStarting); err != nil {
			return err
		}

		uploadSnapshotCreated := false
		created := []string{}
		snapshotCommitted := false
		defer func() {
			if snapshotCommitted {
				return
			}
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_ = r.dropSnapshotTables(cleanupCtx, conn, created)
			_, _ = conn.ExecContext(cleanupCtx, "DELETE FROM event_test_run_tables")
			_, _ = conn.ExecContext(cleanupCtx, "DELETE FROM event_test_runs WHERE id = 1")
			if uploadSnapshotCreated {
				_ = r.removeUploadSnapshot()
			}
		}()

		if r.hasUploadSnapshot() {
			if err := r.createUploadSnapshot(); err != nil {
				return fmt.Errorf("snapshot uploaded files: %w", err)
			}
			uploadSnapshotCreated = true
		}

		tables, err := r.listApplicationTables(ctx, conn)
		if err != nil {
			return err
		}
		if len(tables) == 0 {
			return errors.New("no application tables found to snapshot")
		}
		if err := r.populateWritableColumns(ctx, conn, tables); err != nil {
			return err
		}

		created = make([]string, 0, len(tables))
		for _, table := range tables {
			snapshotName, err := snapshotTableName(table.name)
			if err != nil {
				return err
			}
			query := fmt.Sprintf("CREATE TABLE %s LIKE %s", quoteIdentifier(snapshotName), quoteIdentifier(table.name)) // #nosec G201 -- both identifiers come from database metadata and pass strict ASCII validation
			// codeql[go/sql-injection]
			if _, err := conn.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("create snapshot for %s: %w", table.name, err)
			}
			created = append(created, snapshotName)
		}

		tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			return err
		}
		defer tx.Rollback()

		for _, table := range tables {
			snapshotName, _ := snapshotTableName(table.name)
			columns := quoteIdentifiers(table.columns)
			query := fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s", quoteIdentifier(snapshotName), columns, columns, quoteIdentifier(table.name)) // #nosec G201 -- every identifier is validated against sqlIdentifierPattern after loading database metadata
			// codeql[go/sql-injection]
			if _, err := tx.ExecContext(ctx, query); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("copy snapshot for %s: %w", table.name, err)
			}
		}
		for _, table := range tables {
			var value any
			if table.autoIncrement.Valid {
				value = table.autoIncrement.Int64
			}
			if _, err := tx.ExecContext(ctx, "INSERT INTO event_test_run_tables (table_name, auto_increment_value) VALUES (?, ?)", table.name, value); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		if _, err := conn.ExecContext(ctx, "UPDATE event_test_runs SET state = ?, last_error = NULL WHERE id = 1", models.EventTestRunStateTesting); err != nil {
			return err
		}
		snapshotCommitted = true
		return nil
	})
}

func (r *eventTestRunRepository) Restore(ctx context.Context, eventID int) error {
	return r.withLock(ctx, func(conn *sql.Conn) (restoreErr error) {
		var activeEventID int
		if err := conn.QueryRowContext(ctx, "SELECT event_id FROM event_test_runs WHERE id = 1").Scan(&activeEventID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrTestRunNotFound
			}
			return err
		}
		if activeEventID != eventID {
			return ErrTestRunEventMismatch
		}
		if _, err := conn.ExecContext(ctx, "UPDATE event_test_runs SET state = ?, last_error = NULL WHERE id = 1", models.EventTestRunStateRestoring); err != nil {
			return err
		}
		defer func() {
			if restoreErr != nil {
				_, _ = conn.ExecContext(context.Background(), "UPDATE event_test_runs SET state = ?, last_error = ? WHERE id = 1", models.EventTestRunStateFailed, restoreErr.Error())
			}
		}()

		tables, err := r.loadSnapshotTables(ctx, conn, eventID)
		if err != nil {
			return err
		}
		if err := r.populateWritableColumns(ctx, conn, tables); err != nil {
			return err
		}
		if r.hasUploadSnapshot() {
			if err := r.restoreUploadSnapshot(); err != nil {
				return fmt.Errorf("restore uploaded files: %w", err)
			}
		}

		if _, err := conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
			return err
		}
		defer conn.ExecContext(context.Background(), "SET FOREIGN_KEY_CHECKS = 1")

		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		for _, table := range tables {
			if _, err := tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", quoteIdentifier(table.name))); err != nil {
				return fmt.Errorf("clear %s before restore: %w", table.name, err)
			}
		}
		for _, table := range tables {
			snapshotName, _ := snapshotTableName(table.name)
			columns := quoteIdentifiers(table.columns)
			query := fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s", quoteIdentifier(table.name), columns, columns, quoteIdentifier(snapshotName)) // #nosec G201 -- every identifier is validated against sqlIdentifierPattern after loading database metadata
			// codeql[go/sql-injection]
			if _, err := tx.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("restore %s: %w", table.name, err)
			}
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		for _, table := range tables {
			if !table.autoIncrement.Valid {
				continue
			}
			query := fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = %d", quoteIdentifier(table.name), table.autoIncrement.Int64)
			if _, err := conn.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("restore auto increment for %s: %w", table.name, err)
			}
		}
		if err := r.clearSnapshotArtifacts(ctx, conn, tables); err != nil {
			return err
		}
		_ = r.removeUploadSnapshot()
		if _, err := conn.ExecContext(ctx, "UPDATE event_test_runs SET state = ?, last_error = NULL WHERE id = 1", models.EventTestRunStateAwaitingNotificationResume); err != nil {
			return err
		}
		return nil
	})
}

func (r *eventTestRunRepository) Discard(ctx context.Context, eventID int) error {
	return r.withLock(ctx, func(conn *sql.Conn) error {
		var activeEventID int
		if err := conn.QueryRowContext(ctx, "SELECT event_id FROM event_test_runs WHERE id = 1").Scan(&activeEventID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrTestRunNotFound
			}
			return err
		}
		if activeEventID != eventID {
			return ErrTestRunEventMismatch
		}
		if err := r.dropOrphanedSnapshots(ctx, conn); err != nil {
			return err
		}
		if _, err := conn.ExecContext(ctx, "DELETE FROM event_test_run_tables"); err != nil {
			return err
		}
		if _, err := conn.ExecContext(ctx, "DELETE FROM event_test_runs WHERE id = 1"); err != nil {
			return err
		}
		_ = r.removeUploadSnapshot()
		return nil
	})
}

func (r *eventTestRunRepository) GetStatus(ctx context.Context) (*models.EventTestRunStatus, error) {
	var status models.EventTestRunStatus
	var lastError sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT event_id, state, started_at, updated_at, last_error,
			(SELECT COUNT(*) FROM notifications
			 WHERE scheduled_at IS NOT NULL AND sent_at IS NULL AND scheduled_at <= UTC_TIMESTAMP(6))
		FROM event_test_runs
		WHERE id = 1`,
	).Scan(&status.EventID, &status.State, &status.StartedAt, &status.UpdatedAt, &lastError, &status.OverdueNotificationCount)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if lastError.Valid {
		value := lastError.String
		status.LastError = &value
	}
	return &status, nil
}

func (r *eventTestRunRepository) ResolveNotificationDelivery(ctx context.Context, policy string) error {
	return r.withLock(ctx, func(conn *sql.Conn) error {
		var state string
		if err := conn.QueryRowContext(ctx, "SELECT state FROM event_test_runs WHERE id = 1").Scan(&state); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrTestRunNotFound
			}
			return err
		}
		if state != models.EventTestRunStateAwaitingNotificationResume {
			return fmt.Errorf("notification delivery cannot be resumed while test run state is %s", state)
		}

		switch policy {
		case models.NotificationResumePolicyResume:
			// The next worker tick sends reservations that became due while paused.
		case models.NotificationResumePolicyShift:
			if _, err := conn.ExecContext(ctx, `
				UPDATE notifications n
				INNER JOIN event_test_runs tr ON tr.id = 1
				SET n.scheduled_at = DATE_ADD(n.scheduled_at, INTERVAL TIMESTAMPDIFF(SECOND, tr.started_at, tr.updated_at) SECOND)
				WHERE n.scheduled_at IS NOT NULL AND n.sent_at IS NULL`); err != nil {
				return err
			}
		case models.NotificationResumePolicyCancelOverdue:
			if _, err := conn.ExecContext(ctx, "DELETE FROM notifications WHERE scheduled_at IS NOT NULL AND sent_at IS NULL AND scheduled_at <= UTC_TIMESTAMP(6)"); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid notification resume policy: %s", policy)
		}

		_, err := conn.ExecContext(ctx, "DELETE FROM event_test_runs WHERE id = 1")
		return err
	})
}

func (r *eventTestRunRepository) withLock(ctx context.Context, fn func(*sql.Conn) error) error {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	var acquired int
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, 10)", testRunLockName).Scan(&acquired); err != nil {
		return err
	}
	if acquired != 1 {
		return errors.New("timed out waiting for the event test run lock")
	}
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = conn.ExecContext(releaseCtx, "SELECT RELEASE_LOCK(?)", testRunLockName)
	}()

	return fn(conn)
}

func (r *eventTestRunRepository) listApplicationTables(ctx context.Context, conn *sql.Conn) ([]testRunTable, error) {
	rows, err := conn.QueryContext(ctx, `
		SELECT TABLE_NAME, AUTO_INCREMENT
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_TYPE = 'BASE TABLE'
			AND TABLE_NAME NOT IN ('schema_migrations', 'event_test_runs', 'event_test_run_tables')
			AND LEFT(TABLE_NAME, ?) <> ?
		ORDER BY TABLE_NAME`, len(testRunSnapshotPrefix), testRunSnapshotPrefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []testRunTable
	for rows.Next() {
		var table testRunTable
		if err := rows.Scan(&table.name, &table.autoIncrement); err != nil {
			return nil, err
		}
		if err := validateSQLIdentifier(table.name); err != nil {
			return nil, fmt.Errorf("invalid application table name: %w", err)
		}
		tables = append(tables, table)
	}
	return tables, rows.Err()
}

func (r *eventTestRunRepository) loadSnapshotTables(ctx context.Context, conn *sql.Conn, eventID int) ([]testRunTable, error) {
	var activeEventID int
	err := conn.QueryRowContext(ctx, "SELECT event_id FROM event_test_runs WHERE id = 1").Scan(&activeEventID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTestRunNotFound
	}
	if err != nil {
		return nil, err
	}
	if activeEventID != eventID {
		return nil, ErrTestRunEventMismatch
	}

	rows, err := conn.QueryContext(ctx, "SELECT table_name, auto_increment_value FROM event_test_run_tables ORDER BY table_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []testRunTable
	for rows.Next() {
		var table testRunTable
		if err := rows.Scan(&table.name, &table.autoIncrement); err != nil {
			return nil, err
		}
		if err := validateSQLIdentifier(table.name); err != nil {
			return nil, fmt.Errorf("invalid snapshot table metadata: %w", err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(tables) == 0 {
		return nil, ErrTestRunNotFound
	}
	return tables, nil
}

func (r *eventTestRunRepository) populateWritableColumns(ctx context.Context, conn *sql.Conn, tables []testRunTable) error {
	for index := range tables {
		rows, err := conn.QueryContext(ctx, `
			SELECT COLUMN_NAME
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE()
				AND TABLE_NAME = ?
				AND EXTRA NOT LIKE '%GENERATED%'
			ORDER BY ORDINAL_POSITION`, tables[index].name)
		if err != nil {
			return err
		}
		for rows.Next() {
			var column string
			if err := rows.Scan(&column); err != nil {
				rows.Close()
				return err
			}
			if err := validateSQLIdentifier(column); err != nil {
				rows.Close()
				return fmt.Errorf("invalid column name for table %s: %w", tables[index].name, err)
			}
			tables[index].columns = append(tables[index].columns, column)
		}
		if err := rows.Close(); err != nil {
			return err
		}
		if len(tables[index].columns) == 0 {
			return fmt.Errorf("table %s has no writable columns", tables[index].name)
		}
	}
	return nil
}

func (r *eventTestRunRepository) clearSnapshotArtifacts(ctx context.Context, conn *sql.Conn, tables []testRunTable) error {
	names := make([]string, 0, len(tables))
	for _, table := range tables {
		snapshotName, err := snapshotTableName(table.name)
		if err != nil {
			return err
		}
		names = append(names, snapshotName)
	}
	if err := r.dropSnapshotTables(ctx, conn, names); err != nil {
		return err
	}
	_, err := conn.ExecContext(ctx, "DELETE FROM event_test_run_tables")
	return err
}

func (r *eventTestRunRepository) dropOrphanedSnapshots(ctx context.Context, conn *sql.Conn) error {
	rows, err := conn.QueryContext(ctx, `
		SELECT TABLE_NAME
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_TYPE = 'BASE TABLE'
			AND LEFT(TABLE_NAME, ?) = ?
		ORDER BY TABLE_NAME`, len(testRunSnapshotPrefix), testRunSnapshotPrefix)
	if err != nil {
		return err
	}
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return err
		}
		if err := validateSQLIdentifier(name); err != nil {
			rows.Close()
			return fmt.Errorf("invalid orphaned snapshot table name: %w", err)
		}
		names = append(names, name)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	return r.dropSnapshotTables(ctx, conn, names)
}

func (r *eventTestRunRepository) dropSnapshotTables(ctx context.Context, conn *sql.Conn, names []string) error {
	if len(names) == 0 {
		return nil
	}
	sort.Strings(names)
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		if !strings.HasPrefix(name, testRunSnapshotPrefix) {
			return fmt.Errorf("refusing to drop non-snapshot table %q", name)
		}
		if err := validateSQLIdentifier(name); err != nil {
			return err
		}
		quoted = append(quoted, quoteIdentifier(name))
	}
	query := "DROP TABLE IF EXISTS " + strings.Join(quoted, ", ") // #nosec G202 -- names require the snapshot prefix and strict ASCII SQL identifier validation
	// codeql[go/sql-injection]
	_, err := conn.ExecContext(ctx, query)
	return err
}

func snapshotTableName(tableName string) (string, error) {
	if err := validateSQLIdentifier(tableName); err != nil {
		return "", err
	}
	name := testRunSnapshotPrefix + tableName
	if len(name) > 64 {
		return "", fmt.Errorf("snapshot table name is too long for %q", tableName)
	}
	return name, nil
}

func validateSQLIdentifier(identifier string) error {
	if identifier == "" || len(identifier) > 64 || !sqlIdentifierPattern.MatchString(identifier) {
		return fmt.Errorf("unsafe SQL identifier %q", identifier)
	}
	return nil
}

func quoteIdentifier(identifier string) string {
	return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
}

func quoteIdentifiers(identifiers []string) string {
	quoted := make([]string, 0, len(identifiers))
	for _, identifier := range identifiers {
		quoted = append(quoted, quoteIdentifier(identifier))
	}
	return strings.Join(quoted, ", ")
}
