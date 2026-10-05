package repository

import (
	"backapp/internal/models"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/patrickmn/go-cache"
)

func TestSportDefaultsCacheSkipsAllWarmQueries(t *testing.T) {
	previous := GlobalCache
	GlobalCache = cache.New(time.Minute, time.Minute)
	defer func() { GlobalCache = previous }()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewSportRepository(db)
	names := []string{"Relay", "Tug"}
	rows := func() *sqlmock.Rows {
		return sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Relay").AddRow(2, "Tug")
	}
	mock.ExpectQuery("SELECT id, name FROM sports").WillReturnRows(rows())
	for i := 0; i < 8; i++ {
		sports, err := repo.GetAllSportsWithDefaults(names)
		if err != nil || len(sports) != 2 {
			t.Fatalf("read: %v", err)
		}
	}
	mock.ExpectExec("INSERT INTO sports").WithArgs("New").WillReturnResult(sqlmock.NewResult(3, 1))
	if _, err := repo.CreateSport(&models.Sport{Name: "New"}); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT id, name FROM sports").WillReturnRows(rows().AddRow(3, "New"))
	sports, err := repo.GetAllSportsWithDefaults(names)
	if err != nil || len(sports) != 3 {
		t.Fatal("new sport did not invalidate cache")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSportDefaultsCreatesOnlyMissingNames(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewSportRepository(db)
	mock.ExpectQuery("SELECT id, name FROM sports").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Relay"))
	mock.ExpectExec("INSERT INTO sports").WithArgs("Tug").WillReturnResult(sqlmock.NewResult(2, 1))
	sports, err := repo.GetAllSportsWithDefaults([]string{"Relay", "Tug", "Tug", " "})
	if err != nil || len(sports) != 2 {
		t.Fatalf("defaults: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
