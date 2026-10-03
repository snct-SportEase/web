package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestEventTestRunRepositoryBegin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewEventTestRunRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT GET_LOCK(?, 10)")).
		WithArgs(testRunLockName).
		WillReturnRows(sqlmock.NewRows([]string{"lock"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT event_id FROM event_test_runs WHERE id = 1")).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("(?s)SELECT TABLE_NAME.*LEFT\\(TABLE_NAME").
		WithArgs(len(testRunSnapshotPrefix), testRunSnapshotPrefix).
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_NAME"}))
	mock.ExpectQuery("(?s)SELECT TABLE_NAME, AUTO_INCREMENT.*TABLE_NAME NOT IN").
		WithArgs(len(testRunSnapshotPrefix), testRunSnapshotPrefix).
		WillReturnRows(sqlmock.NewRows([]string{"TABLE_NAME", "AUTO_INCREMENT"}).
			AddRow("events", 9).
			AddRow("users", nil))
	mock.ExpectQuery("(?s)SELECT COLUMN_NAME.*EXTRA NOT LIKE").
		WithArgs("events").
		WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME"}).AddRow("id").AddRow("name"))
	mock.ExpectQuery("(?s)SELECT COLUMN_NAME.*EXTRA NOT LIKE").
		WithArgs("users").
		WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME"}).AddRow("id").AddRow("name"))
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE `_event_test_snapshot_events` LIKE `events`")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE `_event_test_snapshot_users` LIKE `users`")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `_event_test_snapshot_events` (`id`, `name`) SELECT `id`, `name` FROM `events`")).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `_event_test_snapshot_users` (`id`, `name`) SELECT `id`, `name` FROM `users`")).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO event_test_runs (id, event_id) VALUES (1, ?)")).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO event_test_run_tables (table_name, auto_increment_value) VALUES (?, ?)")).
		WithArgs("events", int64(9)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO event_test_run_tables (table_name, auto_increment_value) VALUES (?, ?)")).
		WithArgs("users", nil).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectExec(regexp.QuoteMeta("SELECT RELEASE_LOCK(?)")).
		WithArgs(testRunLockName).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Begin(context.Background(), 7); err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEventTestRunRepositoryRestore(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewEventTestRunRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT GET_LOCK(?, 10)")).
		WithArgs(testRunLockName).
		WillReturnRows(sqlmock.NewRows([]string{"lock"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT event_id FROM event_test_runs WHERE id = 1")).
		WillReturnRows(sqlmock.NewRows([]string{"event_id"}).AddRow(7))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT table_name, auto_increment_value FROM event_test_run_tables ORDER BY table_name")).
		WillReturnRows(sqlmock.NewRows([]string{"table_name", "auto_increment_value"}).
			AddRow("events", 9).
			AddRow("users", nil))
	mock.ExpectQuery("(?s)SELECT COLUMN_NAME.*EXTRA NOT LIKE").
		WithArgs("events").
		WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME"}).AddRow("id").AddRow("name"))
	mock.ExpectQuery("(?s)SELECT COLUMN_NAME.*EXTRA NOT LIKE").
		WithArgs("users").
		WillReturnRows(sqlmock.NewRows([]string{"COLUMN_NAME"}).AddRow("id").AddRow("name"))
	mock.ExpectExec(regexp.QuoteMeta("SET FOREIGN_KEY_CHECKS = 0")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `events`")).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `users`")).WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `events` (`id`, `name`) SELECT `id`, `name` FROM `_event_test_snapshot_events`")).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users` (`id`, `name`) SELECT `id`, `name` FROM `_event_test_snapshot_users`")).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()
	mock.ExpectExec(regexp.QuoteMeta("ALTER TABLE `events` AUTO_INCREMENT = 9")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM event_test_run_tables")).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM event_test_runs WHERE id = 1")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("DROP TABLE IF EXISTS `_event_test_snapshot_events`, `_event_test_snapshot_users`")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("SET FOREIGN_KEY_CHECKS = 1")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("SELECT RELEASE_LOCK(?)")).
		WithArgs(testRunLockName).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Restore(context.Background(), 7); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSnapshotTableNameRejectsOverlongNames(t *testing.T) {
	if _, err := snapshotTableName("this_table_name_is_far_too_long_to_fit_after_the_snapshot_prefix_is_added"); err == nil {
		t.Fatal("expected an overlong snapshot table name to be rejected")
	}
}
