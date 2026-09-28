package repository_test

import (
	"backapp/internal/repository"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func setupUser(t *testing.T) (repository.UserRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return repository.NewUserRepository(db), mock, func() { db.Close() }
}

func TestFindUsersIncludesNotificationAndPWAStatus(t *testing.T) {
	repo, mock, cleanup := setupUser(t)
	defer cleanup()

	now := time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "email", "display_name", "class_id", "notification_filters",
		"is_profile_complete", "created_at", "updated_at", "has_push_subscription", "pwa_last_seen_at",
	}).
		AddRow("user-1", "one@example.com", "One", nil, `["general"]`, true, now, now, true, now).
		AddRow("user-2", "two@example.com", nil, nil, `["general"]`, false, now, now, false, nil)
	mock.ExpectQuery(`(?s)SELECT u.id.*EXISTS\(SELECT 1 FROM push_subscriptions ps WHERE ps.user_id = u.id\).*u.pwa_last_seen_at.*FROM users u`).
		WillReturnRows(rows)
	mock.ExpectQuery(`(?s)SELECT DISTINCT ur.user_id.*FROM roles r.*WHERE ur.user_id IN \(\?,\?\)`).
		WithArgs("user-1", "user-2").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "id", "name"}))

	users, err := repo.FindUsers("", "")
	require.NoError(t, err)
	require.Len(t, users, 2)
	require.True(t, users[0].HasPushSubscription)
	require.Equal(t, now, *users[0].PWALastSeenAt)
	require.False(t, users[1].HasPushSubscription)
	require.Nil(t, users[1].PWALastSeenAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordPWAVisitPreservesProfileUpdatedAt(t *testing.T) {
	repo, mock, cleanup := setupUser(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET pwa_last_seen_at = CURRENT_TIMESTAMP, updated_at = updated_at WHERE id = ?")).
		WithArgs("user-1").WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.RecordPWAVisit("user-1"))
	require.NoError(t, mock.ExpectationsWereMet())
}
