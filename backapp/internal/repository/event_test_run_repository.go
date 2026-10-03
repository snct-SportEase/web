package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
)

type EventTestRunRepository interface {
	Begin(ctx context.Context, eventID int) error
	Restore(ctx context.Context, eventID int) error
	Discard(ctx context.Context, eventID int) error
}

type eventTestRunRepository struct {
	db *sql.DB
}

type testRunTable struct {
	name          string
	autoIncrement sql.NullInt64
	columns       []string
}

func NewEventTestRunRepository(db *sql.DB) EventTestRunRepository {
	return &eventTestRunRepository{db: db}
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

		created := make([]string, 0, len(tables))
		for _, table := range tables {
			snapshotName, err := snapshotTableName(table.name)
			if err != nil {
				r.dropSnapshotTables(context.Background(), conn, created)
				return err
			}
			if _, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE TABLE %s LIKE %s", quoteIdentifier(snapshotName), quoteIdentifier(table.name))); err != nil {
				r.dropSnapshotTables(context.Background(), conn, created)
				return fmt.Errorf("create snapshot for %s: %w", table.name, err)
			}
			created = append(created, snapshotName)
		}

		tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			r.dropSnapshotTables(context.Background(), conn, created)
			return err
		}
		defer tx.Rollback()

		for _, table := range tables {
			snapshotName, _ := snapshotTableName(table.name)
			columns := quoteIdentifiers(table.columns)
			query := fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s", quoteIdentifier(snapshotName), columns, columns, quoteIdentifier(table.name))
			if _, err := tx.ExecContext(ctx, query); err != nil {
				_ = tx.Rollback()
				r.dropSnapshotTables(context.Background(), conn, created)
				return fmt.Errorf("copy snapshot for %s: %w", table.name, err)
			}
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO event_test_runs (id, event_id) VALUES (1, ?)", eventID); err != nil {
			_ = tx.Rollback()
			r.dropSnapshotTables(context.Background(), conn, created)
			return err
		}
		for _, table := range tables {
			var value any
			if table.autoIncrement.Valid {
				value = table.autoIncrement.Int64
			}
			if _, err := tx.ExecContext(ctx, "INSERT INTO event_test_run_tables (table_name, auto_increment_value) VALUES (?, ?)", table.name, value); err != nil {
				_ = tx.Rollback()
				r.dropSnapshotTables(context.Background(), conn, created)
				return err
			}
		}
		if err := tx.Commit(); err != nil {
			r.dropSnapshotTables(context.Background(), conn, created)
			return err
		}
		return nil
	})
}

func (r *eventTestRunRepository) Restore(ctx context.Context, eventID int) error {
	return r.withLock(ctx, func(conn *sql.Conn) error {
		tables, err := r.loadSnapshotTables(ctx, conn, eventID)
		if err != nil {
			return err
		}
		if err := r.populateWritableColumns(ctx, conn, tables); err != nil {
			return err
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
			query := fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s", quoteIdentifier(table.name), columns, columns, quoteIdentifier(snapshotName))
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
		return r.clearSnapshot(ctx, conn, tables)
	})
}

func (r *eventTestRunRepository) Discard(ctx context.Context, eventID int) error {
	return r.withLock(ctx, func(conn *sql.Conn) error {
		tables, err := r.loadSnapshotTables(ctx, conn, eventID)
		if err != nil {
			return err
		}
		return r.clearSnapshot(ctx, conn, tables)
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

func (r *eventTestRunRepository) clearSnapshot(ctx context.Context, conn *sql.Conn, tables []testRunTable) error {
	if _, err := conn.ExecContext(ctx, "DELETE FROM event_test_run_tables"); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, "DELETE FROM event_test_runs WHERE id = 1"); err != nil {
		return err
	}

	names := make([]string, 0, len(tables))
	for _, table := range tables {
		snapshotName, err := snapshotTableName(table.name)
		if err != nil {
			return err
		}
		names = append(names, snapshotName)
	}
	return r.dropSnapshotTables(ctx, conn, names)
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
		quoted = append(quoted, quoteIdentifier(name))
	}
	_, err := conn.ExecContext(ctx, "DROP TABLE IF EXISTS "+strings.Join(quoted, ", "))
	return err
}

func snapshotTableName(tableName string) (string, error) {
	name := testRunSnapshotPrefix + tableName
	if len(name) > 64 {
		return "", fmt.Errorf("snapshot table name is too long for %q", tableName)
	}
	return name, nil
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
