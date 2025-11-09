package repository

import (
	"backapp/internal/models"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateClasses(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewClassRepository(db)

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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetAllClasses(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewClassRepository(db)

	eventID := 1
	rows := sqlmock.NewRows([]string{"id", "event_id", "name", "student_count", "attend_count"}).
		AddRow(1, eventID, "Class A", 30, 25).
		AddRow(2, eventID, "Class B", 32, 30)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, event_id, name, student_count, attend_count FROM classes WHERE event_id = ? ORDER BY name")).
		WithArgs(eventID).
		WillReturnRows(rows)

	classes, err := repo.GetAllClasses(eventID)
	assert.NoError(t, err)
	assert.NotNil(t, classes)
	assert.Len(t, classes, 2)
	assert.Equal(t, "Class A", classes[0].Name)
	assert.Equal(t, "Class B", classes[1].Name)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetClassByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewClassRepository(db)

	classID := 1
	rows := sqlmock.NewRows([]string{"id", "event_id", "name", "student_count", "attend_count"}).
		AddRow(classID, 1, "Class A", 30, 25)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, event_id, name, student_count, attend_count FROM classes WHERE id = ?")).
		WithArgs(classID).
		WillReturnRows(rows)

	class, err := repo.GetClassByID(classID)
	assert.NoError(t, err)
	assert.NotNil(t, class)
	assert.Equal(t, classID, class.ID)

	// Test not found
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, event_id, name, student_count, attend_count FROM classes WHERE id = ?")).
		WithArgs(2).
		WillReturnError(sql.ErrNoRows)

	class, err = repo.GetClassByID(2)
	assert.NoError(t, err)
	assert.Nil(t, class)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetClassMembers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewClassRepository(db)
	classID := 1

	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "email", "display_name", "class_id", "is_profile_complete", "created_at", "updated_at"}).
		AddRow("user1", "user1@example.com", "User One", sql.NullInt32{Int32: int32(classID), Valid: true}, true, now, now).
		AddRow("user2", "user2@example.com", "User Two", sql.NullInt32{Int32: int32(classID), Valid: true}, true, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, display_name, class_id, is_profile_complete, created_at, updated_at FROM users WHERE class_id = ? ORDER BY display_name, email")).
		WithArgs(classID).
		WillReturnRows(rows)

	users, err := repo.GetClassMembers(classID)

	assert.NoError(t, err)
	assert.NotNil(t, users)
	assert.Len(t, users, 2)
	assert.Equal(t, "user1", users[0].ID)
	assert.Equal(t, "User Two", *users[1].DisplayName)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetClassByRepRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := NewClassRepository(db)

	userID := "rep_user"
	eventID := 1
	expectedClass := &models.Class{
		ID:           1,
		Name:         "1-1",
		StudentCount: 30,
		AttendCount:  25,
	}
	eventIDPtr := &eventID

	rows := sqlmock.NewRows([]string{"id", "event_id", "name", "student_count", "attend_count"}).
		AddRow(expectedClass.ID, eventIDPtr, expectedClass.Name, expectedClass.StudentCount, 25)

	query := `
		SELECT c.id, c.event_id, c.name, c.student_count, c.attend_count
		FROM classes c
		INNER JOIN user_roles ur ON ur.user_id = ?
		INNER JOIN roles ro ON ur.role_id = ro.id
		WHERE ro.name = CONCAT(c.name, '_rep') 
		AND (ur.event_id = ? OR ur.event_id IS NULL)
		AND c.event_id = ?
		LIMIT 1
	`

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(userID, eventID, eventID).
		WillReturnRows(rows)

	class, err := repo.GetClassByRepRole(userID, eventID)

	assert.NoError(t, err)
	assert.NotNil(t, class)
	assert.Equal(t, expectedClass.ID, class.ID)
	assert.Equal(t, expectedClass.Name, class.Name)
	assert.Equal(t, eventID, *class.EventID)
	assert.Equal(t, expectedClass.AttendCount, class.AttendCount)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
