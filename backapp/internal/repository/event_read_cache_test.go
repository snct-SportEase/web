package repository

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/patrickmn/go-cache"
)

func TestEventReadCacheInvalidatesSettings(t *testing.T) {
	previous := GlobalCache
	GlobalCache = cache.New(time.Minute, time.Minute)
	defer func() { GlobalCache = previous }()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewEventRepository(db)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT event_id FROM active_event WHERE id = 1")).WillReturnRows(sqlmock.NewRows([]string{"event_id"}).AddRow(1))
	expectEvent := func(rainy bool) {
		mock.ExpectQuery("SELECT id, name").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "year", "season", "start", "end", "rainy", "pdf", "survey", "published", "mic", "status", "hidden", "threshold"}).AddRow(1, "Event", 2026, "spring", nil, nil, rainy, nil, nil, false, false, "active", false, 2))
	}
	expectEvent(false)
	for i := 0; i < 8; i++ {
		id, err := repo.GetActiveEvent()
		if err != nil || id != 1 {
			t.Fatalf("active event: %v", err)
		}
		event, err := repo.GetEventByID(id)
		if err != nil || event.IsRainyMode {
			t.Fatalf("event: %v", err)
		}
	}
	mock.ExpectExec("UPDATE events SET is_rainy_mode").WithArgs(true, 1).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.SetRainyMode(1, true); err != nil {
		t.Fatal(err)
	}
	expectEvent(true)
	event, err := repo.GetEventByID(1)
	if err != nil || !event.IsRainyMode {
		t.Fatal("rainy mode remained cached")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
