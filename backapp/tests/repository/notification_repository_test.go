package repository_test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"backapp/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUpsertPushSubscriptionInsertsBelowLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	repo := repository.NewNotificationRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id FROM push_subscriptions WHERE endpoint = ? FOR UPDATE")).
		WithArgs("https://push.example/new").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM push_subscriptions WHERE user_id = ? FOR UPDATE")).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2).AddRow(3).AddRow(4))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO push_subscriptions (user_id, endpoint, auth_key, p256dh_key) VALUES (?, ?, ?, ?)")).
		WithArgs("user-1", "https://push.example/new", "auth", "p256dh").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.UpsertPushSubscription("user-1", "https://push.example/new", "auth", "p256dh", 5); err != nil {
		t.Fatalf("upsert subscription: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpsertPushSubscriptionRejectsLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	repo := repository.NewNotificationRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id FROM push_subscriptions WHERE endpoint = ? FOR UPDATE")).
		WithArgs("https://push.example/new").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM push_subscriptions WHERE user_id = ? FOR UPDATE")).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2).AddRow(3).AddRow(4).AddRow(5))
	mock.ExpectRollback()

	err = repo.UpsertPushSubscription("user-1", "https://push.example/new", "auth", "p256dh", 5)
	if !errors.Is(err, repository.ErrPushSubscriptionLimit) {
		t.Fatalf("error = %v, want ErrPushSubscriptionLimit", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpsertPushSubscriptionRejectsEndpointOwnedByAnotherUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	repo := repository.NewNotificationRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id FROM push_subscriptions WHERE endpoint = ? FOR UPDATE")).
		WithArgs("https://push.example/existing").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user-2"))
	mock.ExpectRollback()

	err = repo.UpsertPushSubscription("user-1", "https://push.example/existing", "auth", "p256dh", 5)
	if !errors.Is(err, repository.ErrPushEndpointInUse) {
		t.Fatalf("error = %v, want ErrPushEndpointInUse", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
