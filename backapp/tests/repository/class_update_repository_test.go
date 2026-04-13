package repository_test

import (
	"errors"
	"regexp"
	"testing"

	"backapp/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// ─── UpdateAttendance ──────────────────────────────────────────────────────

func TestClassRepository_UpdateAttendance(t *testing.T) {
	const (
		selectClass  = "SELECT student_count, name FROM classes WHERE id = ? AND event_id = ?"
		updateAttend = "UPDATE classes SET attend_count = ? WHERE id = ? AND event_id = ?"
		deleteLog    = "DELETE FROM score_logs WHERE event_id = ? AND class_id = ? AND reason = 'attendance_points'"
		insertLog    = "INSERT INTO score_logs (event_id, class_id, points, reason) VALUES (?, ?, ?, ?)"
	)

	setup := func(t *testing.T) (repository.ClassRepository, sqlmock.Sqlmock, func()) {
		t.Helper()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		repo := repository.NewClassRepository(db)
		return repo, mock, func() { db.Close() }
	}

	t.Run("success - attendance >= 0.9 returns 10 points", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		// 27/30 = 0.9 → 10 points
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(selectClass)).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"student_count", "name"}).AddRow(30, "1-1"))
		mock.ExpectExec(regexp.QuoteMeta(updateAttend)).
			WithArgs(27, 1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(deleteLog)).
			WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(regexp.QuoteMeta(insertLog)).
			WithArgs(1, 1, 10, "attendance_points").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		points, err := repo.UpdateAttendance(1, 1, 27)
		assert.NoError(t, err)
		assert.Equal(t, 10, points)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - attendance >= 0.8 returns 9 points", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		// 24/30 = 0.8 → 9 points
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(selectClass)).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"student_count", "name"}).AddRow(30, "1-1"))
		mock.ExpectExec(regexp.QuoteMeta(updateAttend)).
			WithArgs(24, 1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(deleteLog)).
			WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(regexp.QuoteMeta(insertLog)).
			WithArgs(1, 1, 9, "attendance_points").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		points, err := repo.UpdateAttendance(1, 1, 24)
		assert.NoError(t, err)
		assert.Equal(t, 9, points)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - attendance < 0.5 returns 5 points (minimum)", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		// 10/30 = 0.33 → 5 points
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(selectClass)).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"student_count", "name"}).AddRow(30, "1-1"))
		mock.ExpectExec(regexp.QuoteMeta(updateAttend)).
			WithArgs(10, 1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(deleteLog)).
			WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(regexp.QuoteMeta(insertLog)).
			WithArgs(1, 1, 5, "attendance_points").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		points, err := repo.UpdateAttendance(1, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, 5, points)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - 専教 class always returns 0 points", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(selectClass)).
			WithArgs(2, 1).
			WillReturnRows(sqlmock.NewRows([]string{"student_count", "name"}).AddRow(20, "専教"))
		mock.ExpectExec(regexp.QuoteMeta(updateAttend)).
			WithArgs(15, 2, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(deleteLog)).
			WithArgs(1, 2).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(regexp.QuoteMeta(insertLog)).
			WithArgs(1, 2, 0, "attendance_points").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		points, err := repo.UpdateAttendance(2, 1, 15)
		assert.NoError(t, err)
		assert.Equal(t, 0, points)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - begin transaction error", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		dbErr := errors.New("connection refused")
		mock.ExpectBegin().WillReturnError(dbErr)

		points, err := repo.UpdateAttendance(1, 1, 27)
		assert.Error(t, err)
		assert.Equal(t, 0, points)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - class not found", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(selectClass)).
			WithArgs(999, 1).
			WillReturnRows(sqlmock.NewRows([]string{"student_count", "name"})) // no row → Scan returns ErrNoRows
		mock.ExpectRollback()

		points, err := repo.UpdateAttendance(999, 1, 27)
		assert.Error(t, err)
		assert.Equal(t, 0, points)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - student_count is zero", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(selectClass)).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"student_count", "name"}).AddRow(0, "1-1"))
		mock.ExpectExec(regexp.QuoteMeta(updateAttend)).
			WithArgs(5, 1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectRollback()

		points, err := repo.UpdateAttendance(1, 1, 5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "zero students")
		assert.Equal(t, 0, points)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - UPDATE attend_count fails", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		dbErr := errors.New("db error")
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(selectClass)).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"student_count", "name"}).AddRow(30, "1-1"))
		mock.ExpectExec(regexp.QuoteMeta(updateAttend)).
			WithArgs(27, 1, 1).WillReturnError(dbErr)
		mock.ExpectRollback()

		points, err := repo.UpdateAttendance(1, 1, 27)
		assert.Error(t, err)
		assert.Equal(t, 0, points)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - DELETE score_logs fails", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		dbErr := errors.New("db error")
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(selectClass)).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"student_count", "name"}).AddRow(30, "1-1"))
		mock.ExpectExec(regexp.QuoteMeta(updateAttend)).
			WithArgs(27, 1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(deleteLog)).
			WithArgs(1, 1).WillReturnError(dbErr)
		mock.ExpectRollback()

		points, err := repo.UpdateAttendance(1, 1, 27)
		assert.Error(t, err)
		assert.Equal(t, 0, points)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - INSERT score_logs fails", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		dbErr := errors.New("db error")
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(selectClass)).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"student_count", "name"}).AddRow(30, "1-1"))
		mock.ExpectExec(regexp.QuoteMeta(updateAttend)).
			WithArgs(27, 1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(deleteLog)).
			WithArgs(1, 1).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(regexp.QuoteMeta(insertLog)).
			WithArgs(1, 1, 10, "attendance_points").WillReturnError(dbErr)
		mock.ExpectRollback()

		points, err := repo.UpdateAttendance(1, 1, 27)
		assert.Error(t, err)
		assert.Equal(t, 0, points)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── UpdateStudentCounts ───────────────────────────────────────────────────

func TestClassRepository_UpdateStudentCounts(t *testing.T) {
	const updateStmt = "UPDATE classes SET student_count = ? WHERE id = ? AND event_id = ?"

	setup := func(t *testing.T) (repository.ClassRepository, sqlmock.Sqlmock, func()) {
		t.Helper()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		return repository.NewClassRepository(db), mock, func() { db.Close() }
	}

	t.Run("success - single class updated", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta(updateStmt)).
			ExpectExec().WithArgs(30, 1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateStudentCounts(1, map[int]int{1: 30})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - empty counts skips all SQL", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta(updateStmt))
		mock.ExpectCommit()

		err := repo.UpdateStudentCounts(1, map[int]int{})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - begin transaction error", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin().WillReturnError(errors.New("connection refused"))

		err := repo.UpdateStudentCounts(1, map[int]int{1: 30})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - prepare error", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta(updateStmt)).WillReturnError(errors.New("prepare error"))
		mock.ExpectRollback()

		err := repo.UpdateStudentCounts(1, map[int]int{1: 30})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - exec error stops subsequent updates", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta(updateStmt)).
			ExpectExec().WithArgs(30, 1, 1).WillReturnError(errors.New("exec error"))
		mock.ExpectRollback()

		err := repo.UpdateStudentCounts(1, map[int]int{1: 30})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── SetNoonGamePoints ─────────────────────────────────────────────────────

func TestClassRepository_SetNoonGamePoints(t *testing.T) {
	const (
		deleteNoon = "DELETE FROM score_logs WHERE event_id = ? AND reason = 'noon_game_points'"
		insertNoon = `
				INSERT INTO score_logs (event_id, class_id, points, reason)
				VALUES (?, ?, ?, 'noon_game_points')
			`
	)

	setup := func(t *testing.T) (repository.ClassRepository, sqlmock.Sqlmock, func()) {
		t.Helper()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		return repository.NewClassRepository(db), mock, func() { db.Close() }
	}

	t.Run("success - replaces existing points", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(deleteNoon)).
			WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectPrepare(regexp.QuoteMeta(insertNoon)).
			ExpectExec().WithArgs(1, 10, 50).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.SetNoonGamePoints(1, map[int]int{10: 50})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - empty points only deletes previous records", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(deleteNoon)).
			WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()

		err := repo.SetNoonGamePoints(1, map[int]int{})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - begin transaction error", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin().WillReturnError(errors.New("connection refused"))

		err := repo.SetNoonGamePoints(1, map[int]int{10: 50})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - DELETE fails", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(deleteNoon)).
			WithArgs(1).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.SetNoonGamePoints(1, map[int]int{10: 50})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to reset noon_game_points")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - prepare INSERT fails", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(deleteNoon)).
			WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectPrepare(regexp.QuoteMeta(insertNoon)).WillReturnError(errors.New("prepare error"))
		mock.ExpectRollback()

		err := repo.SetNoonGamePoints(1, map[int]int{10: 50})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to prepare noon_game_points statement")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - INSERT exec fails", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(deleteNoon)).
			WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectPrepare(regexp.QuoteMeta(insertNoon)).
			ExpectExec().WithArgs(1, 10, 50).WillReturnError(errors.New("insert error"))
		mock.ExpectRollback()

		err := repo.SetNoonGamePoints(1, map[int]int{10: 50})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update noon_game_points for class 10")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── SetSurveyPoints ──────────────────────────────────────────────────────

func TestClassRepository_SetSurveyPoints(t *testing.T) {
	const (
		deleteSurvey = "DELETE FROM score_logs WHERE event_id = ? AND reason = 'survey_points'"
		insertSurvey = `
				INSERT INTO score_logs (event_id, class_id, points, reason)
				VALUES (?, ?, ?, 'survey_points')
			`
	)

	setup := func(t *testing.T) (repository.ClassRepository, sqlmock.Sqlmock, func()) {
		t.Helper()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		return repository.NewClassRepository(db), mock, func() { db.Close() }
	}

	t.Run("success - replaces existing points", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(deleteSurvey)).
			WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectPrepare(regexp.QuoteMeta(insertSurvey)).
			ExpectExec().WithArgs(1, 10, 20).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.SetSurveyPoints(1, map[int]int{10: 20})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - empty points only deletes previous records", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(deleteSurvey)).
			WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()

		err := repo.SetSurveyPoints(1, map[int]int{})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - begin transaction error", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin().WillReturnError(errors.New("connection refused"))

		err := repo.SetSurveyPoints(1, map[int]int{10: 20})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - DELETE fails", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(deleteSurvey)).
			WithArgs(1).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.SetSurveyPoints(1, map[int]int{10: 20})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to reset survey_points")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - prepare INSERT fails", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(deleteSurvey)).
			WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectPrepare(regexp.QuoteMeta(insertSurvey)).WillReturnError(errors.New("prepare error"))
		mock.ExpectRollback()

		err := repo.SetSurveyPoints(1, map[int]int{10: 20})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to prepare survey_points statement")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - INSERT exec fails", func(t *testing.T) {
		repo, mock, close := setup(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(deleteSurvey)).
			WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectPrepare(regexp.QuoteMeta(insertSurvey)).
			ExpectExec().WithArgs(1, 10, 20).WillReturnError(errors.New("insert error"))
		mock.ExpectRollback()

		err := repo.SetSurveyPoints(1, map[int]int{10: 20})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update survey_points for class 10")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
