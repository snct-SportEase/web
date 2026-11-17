package repository_test

import (
	"errors"
	"testing"

	"backapp/internal/models"
	"backapp/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupNotificationRequestRepository(t *testing.T) (repository.NotificationRequestRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return repository.NewNotificationRequestRepository(db), mock
}

func TestNotificationRequestRepository_ResolveRequestApprovesAtomically(t *testing.T) {
	repo, mock := setupNotificationRequestRepository(t)
	roles := []string{"student", "admin", "root"}

	mock.ExpectBegin()
	mock.ExpectExec(`(?s)UPDATE notification_requests.*WHERE id = \? AND status = \?`).
		WithArgs(models.NotificationRequestStatusApproved, "root-1", 7, models.NotificationRequestStatusPending).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)INSERT INTO notifications.*SELECT title, body`).
		WithArgs("root-1", 7).
		WillReturnResult(sqlmock.NewResult(23, 1))
	for _, role := range roles {
		mock.ExpectExec(`INSERT INTO notification_targets`).
			WithArgs(int64(23), role).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()

	notificationID, err := repo.ResolveRequest(7, models.NotificationRequestStatusApproved, "root-1", roles)
	require.NoError(t, err)
	assert.Equal(t, int64(23), notificationID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRequestRepository_ResolveRequestRejectsWithoutNotification(t *testing.T) {
	repo, mock := setupNotificationRequestRepository(t)

	mock.ExpectBegin()
	mock.ExpectExec(`(?s)UPDATE notification_requests.*WHERE id = \? AND status = \?`).
		WithArgs(models.NotificationRequestStatusRejected, "root-1", 7, models.NotificationRequestStatusPending).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	notificationID, err := repo.ResolveRequest(7, models.NotificationRequestStatusRejected, "root-1", []string{"student", "admin", "root"})
	require.NoError(t, err)
	assert.Zero(t, notificationID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRequestRepository_ResolveRequestIsIdempotent(t *testing.T) {
	repo, mock := setupNotificationRequestRepository(t)

	mock.ExpectBegin()
	mock.ExpectExec(`(?s)UPDATE notification_requests.*WHERE id = \? AND status = \?`).
		WithArgs(models.NotificationRequestStatusApproved, "root-1", 7, models.NotificationRequestStatusPending).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	_, err := repo.ResolveRequest(7, models.NotificationRequestStatusApproved, "root-1", []string{"student", "admin", "root"})
	assert.True(t, errors.Is(err, repository.ErrNotificationRequestAlreadyResolved))
	require.NoError(t, mock.ExpectationsWereMet())
}
