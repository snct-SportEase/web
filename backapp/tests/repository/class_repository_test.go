package repository_test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"backapp/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestClassRepository_CreateClasses(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		defer db.Close()

		repo := repository.NewClassRepository(db)
		eventID := 1
		classNames := []string{"Class A", "Class B"}

		mock.ExpectBegin()
		stmt := mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO classes (event_id, name) VALUES (?, ?)"))
		for _, name := range classNames {
			stmt.ExpectExec().WithArgs(eventID, name).WillReturnResult(sqlmock.NewResult(1, 1))
		}
		mock.ExpectCommit()

		err = repo.CreateClasses(eventID, classNames)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("begin transaction error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		defer db.Close()

		repo := repository.NewClassRepository(db)
		dbErr := errors.New("connection refused")
		mock.ExpectBegin().WillReturnError(dbErr)

		err = repo.CreateClasses(1, []string{"Class A"})
		assert.ErrorIs(t, err, dbErr)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exec error rolls back transaction", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		defer db.Close()

		repo := repository.NewClassRepository(db)
		dbErr := errors.New("duplicate entry")

		mock.ExpectBegin()
		stmt := mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO classes (event_id, name) VALUES (?, ?)"))
		stmt.ExpectExec().WithArgs(1, "Class A").WillReturnError(dbErr)
		mock.ExpectRollback()

		err = repo.CreateClasses(1, []string{"Class A"})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestClassRepository_GetAllClasses(t *testing.T) {
	const query = "SELECT id, event_id, name, student_count, attend_count FROM classes WHERE event_id = ? ORDER BY name"

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		defer db.Close()

		repo := repository.NewClassRepository(db)
		eventID := 1
		rows := sqlmock.NewRows([]string{"id", "event_id", "name", "student_count", "attend_count"}).
			AddRow(1, eventID, "Class A", 30, 25).
			AddRow(2, eventID, "Class B", 32, 30)

		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(eventID).WillReturnRows(rows)

		classes, err := repo.GetAllClasses(eventID)
		assert.NoError(t, err)
		assert.Len(t, classes, 2)
		assert.Equal(t, "Class A", classes[0].Name)
		assert.Equal(t, "Class B", classes[1].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		defer db.Close()

		repo := repository.NewClassRepository(db)
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(999).
			WillReturnRows(sqlmock.NewRows([]string{"id", "event_id", "name", "student_count", "attend_count"}))

		classes, err := repo.GetAllClasses(999)
		assert.NoError(t, err)
		assert.Nil(t, classes)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		defer db.Close()

		repo := repository.NewClassRepository(db)
		dbErr := errors.New("db error")
		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(1).WillReturnError(dbErr)

		classes, err := repo.GetAllClasses(1)
		assert.Error(t, err)
		assert.Nil(t, classes)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestClassRepository_GetClassByID(t *testing.T) {
	const query = "SELECT id, event_id, name, student_count, attend_count FROM classes WHERE id = ?"

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		defer db.Close()

		repo := repository.NewClassRepository(db)
		classID := 1
		rows := sqlmock.NewRows([]string{"id", "event_id", "name", "student_count", "attend_count"}).
			AddRow(classID, 1, "Class A", 30, 25)

		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(classID).WillReturnRows(rows)

		class, err := repo.GetClassByID(classID)
		assert.NoError(t, err)
		assert.NotNil(t, class)
		assert.Equal(t, classID, class.ID)
		assert.Equal(t, "Class A", class.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		defer db.Close()

		repo := repository.NewClassRepository(db)
		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(999).WillReturnError(sql.ErrNoRows)

		class, err := repo.GetClassByID(999)
		assert.NoError(t, err)
		assert.Nil(t, class)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		defer db.Close()

		repo := repository.NewClassRepository(db)
		dbErr := errors.New("connection timeout")
		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(1).WillReturnError(dbErr)

		class, err := repo.GetClassByID(1)
		assert.Error(t, err)
		assert.Nil(t, class)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestClassRepository_GetClassMembers(t *testing.T) {
	const query = "SELECT id, email, display_name, class_id, is_profile_complete, created_at, updated_at FROM users WHERE class_id = ? ORDER BY display_name, email"

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		defer db.Close()

		repo := repository.NewClassRepository(db)
		classID := 1
		now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		rows := sqlmock.NewRows([]string{"id", "email", "display_name", "class_id", "is_profile_complete", "created_at", "updated_at"}).
			AddRow("user1", "user1@example.com", "User One", sql.NullInt32{Int32: int32(classID), Valid: true}, true, now, now).
			AddRow("user2", "user2@example.com", "User Two", sql.NullInt32{Int32: int32(classID), Valid: true}, true, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(classID).WillReturnRows(rows)

		users, err := repo.GetClassMembers(classID)
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "user1", users[0].ID)
		assert.Equal(t, "User Two", *users[1].DisplayName)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		defer db.Close()

		repo := repository.NewClassRepository(db)
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(999).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "display_name", "class_id", "is_profile_complete", "created_at", "updated_at"}))

		users, err := repo.GetClassMembers(999)
		assert.NoError(t, err)
		assert.Nil(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to open stub db: %s", err)
		}
		defer db.Close()

		repo := repository.NewClassRepository(db)
		dbErr := errors.New("db error")
		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(1).WillReturnError(dbErr)

		users, err := repo.GetClassMembers(1)
		assert.Error(t, err)
		assert.Nil(t, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
