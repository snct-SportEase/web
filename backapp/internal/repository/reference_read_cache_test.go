package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/patrickmn/go-cache"
)

func TestGuideCacheInvalidatesAfterDelete(t *testing.T) {
	previous := GlobalCache
	GlobalCache = cache.New(time.Minute, time.Minute)
	defer func() { GlobalCache = previous }()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewGuideDocumentRepository(db)
	columns := []string{"id", "event", "title", "description", "pdf", "created", "updated"}
	mock.ExpectQuery("SELECT id, event_id").WithArgs(1).WillReturnRows(sqlmock.NewRows(columns).AddRow(1, 1, "Guide", nil, "/guide.pdf", time.Now(), time.Now()))
	for i := 0; i < 2; i++ {
		docs, err := repo.ListGuideDocuments(1)
		if err != nil || len(docs) != 1 {
			t.Fatalf("guide read: %v", err)
		}
	}
	mock.ExpectExec("DELETE FROM guide_documents").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.DeleteGuideDocument(1); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT id, event_id").WithArgs(1).WillReturnRows(sqlmock.NewRows(columns))
	docs, err := repo.ListGuideDocuments(1)
	if err != nil || len(docs) != 0 {
		t.Fatal("deleted guide still cached")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSportCacheInvalidatesAfterUnassignment(t *testing.T) {
	previous := GlobalCache
	GlobalCache = cache.New(time.Minute, time.Minute)
	defer func() { GlobalCache = previous }()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewSportRepository(db)
	columns := []string{"event", "sport", "name", "description", "pdf", "location", "template", "min", "max"}
	mock.ExpectQuery("SELECT es.event_id").WithArgs(1).WillReturnRows(sqlmock.NewRows(columns).AddRow(1, 2, "Basketball", nil, nil, "gym1", nil, nil, nil))
	for i := 0; i < 2; i++ {
		sports, err := repo.GetSportsByEventID(1)
		if err != nil || len(sports) != 1 {
			t.Fatalf("sport read: %v", err)
		}
	}
	mock.ExpectExec("DELETE FROM event_sports").WithArgs(1, 2).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.DeleteSportFromEvent(1, 2); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT es.event_id").WithArgs(1).WillReturnRows(sqlmock.NewRows(columns))
	sports, err := repo.GetSportsByEventID(1)
	if err != nil || len(sports) != 0 {
		t.Fatal("unassigned sport still cached")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
