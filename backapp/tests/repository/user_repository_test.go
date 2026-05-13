package repository_test

import (
	"backapp/internal/repository"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func setupUser(t *testing.T) (repository.UserRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return repository.NewUserRepository(db), mock, func() { db.Close() }
}

func TestUserRepository_ReplaceClassRepRole(t *testing.T) {
	repo, mock, cleanup := setupUser(t)
	defer cleanup()

	eventID := 3
	const userID = "user-1"
	const roleName = "2A_rep"
	const classID = 12

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET class_id = ? WHERE id = ?")).
		WithArgs(classID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM roles WHERE name = ?")).
		WithArgs(roleName).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))
	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE ur
		FROM user_roles ur
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = ? AND RIGHT(r.name, 4) = '_rep'
	`)).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO user_roles (user_id, role_id, event_id) VALUES (?, ?, ?)")).
		WithArgs(userID, int64(9), eventID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.ReplaceClassRepRole(userID, roleName, classID, &eventID))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_ReplaceClassRepRoleCreatesRole(t *testing.T) {
	repo, mock, cleanup := setupUser(t)
	defer cleanup()

	const userID = "user-1"
	const roleName = "2A_rep"
	const classID = 12

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET class_id = ? WHERE id = ?")).
		WithArgs(classID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM roles WHERE name = ?")).
		WithArgs(roleName).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO roles (name) VALUES (?)")).
		WithArgs(roleName).
		WillReturnResult(sqlmock.NewResult(9, 1))
	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE ur
		FROM user_roles ur
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = ? AND RIGHT(r.name, 4) = '_rep'
	`)).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO user_roles (user_id, role_id, event_id) VALUES (?, ?, NULL)")).
		WithArgs(userID, int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.ReplaceClassRepRole(userID, roleName, classID, nil))
	require.NoError(t, mock.ExpectationsWereMet())
}
