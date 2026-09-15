package repository_test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

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

func TestAddNotificationRecipients(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	repo := repository.NewNotificationRepository(db)

	mock.ExpectBegin()
	prepared := mock.ExpectPrepare(regexp.QuoteMeta("INSERT IGNORE INTO notification_recipients (notification_id, user_id) VALUES (?, ?)"))
	prepared.ExpectExec().WithArgs(int64(7), "user-1").WillReturnResult(sqlmock.NewResult(1, 1))
	prepared.ExpectExec().WithArgs(int64(7), "user-2").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	if err := repo.AddNotificationRecipients(7, []string{"user-1", "user-2"}); err != nil {
		t.Fatalf("add notification recipients: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetNotificationsForAccessIncludesIndividualRecipient(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	repo := repository.NewNotificationRepository(db)

	createdAt := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`(?s)SELECT.*LEFT JOIN notification_recipients nr.*WHERE \(nt.role_name IN \(\?\) OR nr.user_id = \?\).*GROUP BY n.id`).
		WithArgs("student", "user-1", 50).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "body", "type", "created_by", "event_id", "created_at", "target_roles", "target_user_count",
		}).AddRow(9, "個人連絡", "本文", "general", "root-1", nil, createdAt, nil, 1))

	notifications, err := repo.GetNotificationsForAccess([]string{"student"}, "user-1", false, 50)
	if err != nil {
		t.Fatalf("get notifications: %v", err)
	}
	if len(notifications) != 1 || notifications[0].TargetUserCount != 1 {
		t.Fatalf("unexpected notifications: %#v", notifications)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetPushSubscriptionStatsByTargetsCombinesRolesAndUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	repo := repository.NewNotificationRepository(db)

	mock.ExpectQuery(`(?s)SELECT.*COUNT\(DISTINCT u.id\).*WHERE \(r.name IN \(\?,\?\) OR u.id IN \(\?\)\)`).
		WithArgs("admin", "student", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"target_user_count", "subscribed_user_count", "subscription_endpoint_count"}).AddRow(11, 7, 9))

	stats, err := repo.GetPushSubscriptionStatsByTargets([]string{"admin", "student"}, []string{"user-1"})
	if err != nil {
		t.Fatalf("get push subscription stats: %v", err)
	}
	if stats.TargetUserCount != 11 || stats.SubscribedUserCount != 7 || stats.SubscriptionEndpointCount != 9 {
		t.Fatalf("unexpected stats: %#v", stats)
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
