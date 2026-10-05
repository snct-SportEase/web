package repository

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/patrickmn/go-cache"
)

func TestUserReadCacheQueriesAndWrites(t *testing.T) {
	previous := GlobalCache
	GlobalCache = cache.New(time.Minute, time.Minute)
	defer func() { GlobalCache = previous }()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	users := NewUserRepository(db)
	expectActive := func(id int) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT event_id FROM active_event WHERE id = 1")).WillReturnRows(sqlmock.NewRows([]string{"event_id"}).AddRow(id))
	}
	expectUser := func(role string) {
		mock.ExpectQuery("SELECT id, email, display_name").WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"id", "email", "display_name", "class_id", "filters", "complete", "created", "updated"}).AddRow("alice", "alice@example.com", "Alice", 1, "[]", true, time.Now(), time.Now()))
		mock.ExpectQuery("SELECT DISTINCT r.id, r.name").WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, role))
	}
	expectActive(1)
	expectUser("student")
	for i := 0; i < 8; i++ {
		user, err := users.GetUserWithRoles("alice")
		if err != nil || user.Roles[0].Name != "student" {
			t.Fatalf("read: %v", err)
		}
	}
	mock.ExpectExec("UPDATE users SET display_name").WithArgs("new name", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := users.UpdateUserDisplayName("alice", "new name"); err != nil {
		t.Fatal(err)
	}
	expectUser("student")
	if _, err := users.GetUserWithRoles("alice"); err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM roles").WithArgs("student").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec("DELETE FROM user_roles").WithArgs("alice", 1, nil).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := users.DeleteUserRole("alice", "student", nil); err != nil {
		t.Fatal(err)
	}
	expectUser("admin")
	if _, err := users.GetUserWithRoles("alice"); err != nil {
		t.Fatal(err)
	}
	// Event transitions invalidate migrated profiles as well as scoped roles.
	invalidateReads(db, "events")
	expectActive(2)
	expectUser("student")
	if _, err := users.GetUserWithRoles("alice"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
