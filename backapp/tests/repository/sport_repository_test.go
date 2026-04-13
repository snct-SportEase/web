package repository_test

import (
	"errors"
	"regexp"
	"testing"

	"backapp/internal/models"
	"backapp/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSport(t *testing.T) (repository.SportRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return repository.NewSportRepository(db), mock, func() { db.Close() }
}

var eventSportCols = []string{
	"event_id", "sport_id", "description", "rules", "rules_type",
	"rules_pdf_url", "location", "min_capacity", "max_capacity",
}

// ─── GetAllSports ──────────────────────────────────────────────────────────

func TestSportRepository_GetAllSports(t *testing.T) {
	const q = "SELECT id, name FROM sports ORDER BY id"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "バスケットボール").
				AddRow(2, "サッカー"))

		sports, err := repo.GetAllSports()
		require.NoError(t, err)
		assert.Len(t, sports, 2)
		assert.Equal(t, "バスケットボール", sports[0].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

		sports, err := repo.GetAllSports()
		assert.NoError(t, err)
		assert.Nil(t, sports)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		sports, err := repo.GetAllSports()
		assert.Error(t, err)
		assert.Nil(t, sports)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── GetSportByID ──────────────────────────────────────────────────────────

func TestSportRepository_GetSportByID(t *testing.T) {
	const q = "SELECT id, name FROM sports WHERE id = ?"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "バスケットボール"))

		sport, err := repo.GetSportByID(1)
		require.NoError(t, err)
		require.NotNil(t, sport)
		assert.Equal(t, "バスケットボール", sport.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found returns error", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		// GetSportByID returns errors.New("sport not found") on ErrNoRows
		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(999).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

		sport, err := repo.GetSportByID(999)
		assert.Error(t, err)
		assert.Equal(t, "sport not found", err.Error())
		assert.Nil(t, sport)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		sport, err := repo.GetSportByID(1)
		assert.Error(t, err)
		assert.Nil(t, sport)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── CreateSport ───────────────────────────────────────────────────────────

func TestSportRepository_CreateSport(t *testing.T) {
	const q = "INSERT INTO sports (name) VALUES (?)"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WithArgs("バレーボール").
			WillReturnResult(sqlmock.NewResult(3, 1))

		id, err := repo.CreateSport(&models.Sport{Name: "バレーボール"})
		assert.NoError(t, err)
		assert.Equal(t, int64(3), id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		id, err := repo.CreateSport(&models.Sport{Name: "X"})
		assert.Error(t, err)
		assert.Equal(t, int64(0), id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

// ─── GetSportsByEventID ────────────────────────────────────────────────────

func TestSportRepository_GetSportsByEventID(t *testing.T) {
	const q = "SELECT event_id, sport_id, description, rules, rules_type, rules_pdf_url, location, min_capacity, max_capacity FROM event_sports WHERE event_id = ?"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(1).
			WillReturnRows(sqlmock.NewRows(eventSportCols).
				AddRow(1, 1, "desc", "rules", "markdown", nil, "gym1", 3, 8).
				AddRow(1, 2, "desc2", "rules2", "markdown", nil, "ground", 5, 10))

		sports, err := repo.GetSportsByEventID(1)
		require.NoError(t, err)
		assert.Len(t, sports, 2)
		assert.Equal(t, "gym1", sports[0].Location)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(999).
			WillReturnRows(sqlmock.NewRows(eventSportCols))

		sports, err := repo.GetSportsByEventID(999)
		assert.NoError(t, err)
		assert.Nil(t, sports)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		sports, err := repo.GetSportsByEventID(1)
		assert.Error(t, err)
		assert.Nil(t, sports)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── AssignSportToEvent ────────────────────────────────────────────────────

func TestSportRepository_AssignSportToEvent(t *testing.T) {
	const checkDupSportQ = "SELECT COUNT(*) FROM event_sports WHERE event_id = ? AND sport_id = ?"
	const checkDupLocQ = "SELECT COUNT(*) FROM event_sports WHERE event_id = ? AND location = ?"
	const insertQ = "INSERT INTO event_sports (event_id, sport_id, description, rules, rules_type, rules_pdf_url, location, min_capacity, max_capacity) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"

	es := &models.EventSport{EventID: 1, SportID: 3, Description: stringPtr("desc"), Rules: stringPtr("rules"), RulesType: "markdown", Location: "gym1", MinCapacity: intPtr(3), MaxCapacity: intPtr(8)}

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(checkDupSportQ)).WithArgs(1, 3).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery(regexp.QuoteMeta(checkDupLocQ)).WithArgs(1, "gym1").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec(regexp.QuoteMeta(insertQ)).
			WithArgs(es.EventID, es.SportID, es.Description, es.Rules, es.RulesType, es.RulesPdfURL, es.Location, es.MinCapacity, es.MaxCapacity).
			WillReturnResult(sqlmock.NewResult(1, 1))

		assert.NoError(t, repo.AssignSportToEvent(es))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("sport already assigned to event", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(checkDupSportQ)).WithArgs(1, 3).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		err := repo.AssignSportToEvent(es)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "すでにこの大会に割り当てられています")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("location already used in event", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(checkDupSportQ)).WithArgs(1, 3).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery(regexp.QuoteMeta(checkDupLocQ)).WithArgs(1, "gym1").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		err := repo.AssignSportToEvent(es)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "この場所は")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("location=other allows duplicate location", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		esOther := *es
		esOther.Location = "other"

		mock.ExpectQuery(regexp.QuoteMeta(checkDupSportQ)).WithArgs(1, 3).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		// No location check for "other"
		mock.ExpectExec(regexp.QuoteMeta(insertQ)).
			WithArgs(esOther.EventID, esOther.SportID, esOther.Description, esOther.Rules, esOther.RulesType, esOther.RulesPdfURL, esOther.Location, esOther.MinCapacity, esOther.MaxCapacity).
			WillReturnResult(sqlmock.NewResult(1, 1))

		assert.NoError(t, repo.AssignSportToEvent(&esOther))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── DeleteSportFromEvent ──────────────────────────────────────────────────

func TestSportRepository_DeleteSportFromEvent(t *testing.T) {
	const q = "DELETE FROM event_sports WHERE event_id = ? AND sport_id = ?"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WithArgs(1, 2).
			WillReturnResult(sqlmock.NewResult(0, 1))

		assert.NoError(t, repo.DeleteSportFromEvent(1, 2))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		assert.Error(t, repo.DeleteSportFromEvent(1, 2))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── GetTeamsBySportID ─────────────────────────────────────────────────────

func TestSportRepository_GetTeamsBySportID(t *testing.T) {
	const q = `
			SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id, t.min_capacity, t.max_capacity
			FROM teams t
			JOIN classes c ON t.class_id = c.id
			WHERE t.sport_id = ?
		`
	cols := []string{"id", "name", "class_id", "sport_id", "event_id", "min_capacity", "max_capacity"}

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(1).
			WillReturnRows(sqlmock.NewRows(cols).
				AddRow(10, "IS3-A", 5, 1, 1, 3, 8).
				AddRow(11, "IS3-B", 6, 1, 1, nil, nil))

		teams, err := repo.GetTeamsBySportID(1)
		require.NoError(t, err)
		assert.Len(t, teams, 2)
		assert.Equal(t, 3, *teams[0].MinCapacity)
		assert.Nil(t, teams[1].MinCapacity)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(99).
			WillReturnRows(sqlmock.NewRows(cols))

		teams, err := repo.GetTeamsBySportID(99)
		assert.NoError(t, err)
		assert.Nil(t, teams)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		teams, err := repo.GetTeamsBySportID(1)
		assert.Error(t, err)
		assert.Nil(t, teams)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── GetSportDetails ───────────────────────────────────────────────────────

func TestSportRepository_GetSportDetails(t *testing.T) {
	const q = "SELECT event_id, sport_id, description, rules, rules_type, rules_pdf_url, location, min_capacity, max_capacity FROM event_sports WHERE event_id = ? AND sport_id = ?"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows(eventSportCols).
				AddRow(1, 1, "desc", "rules", "markdown", nil, "gym1", 3, 8))

		es, err := repo.GetSportDetails(1, 1)
		require.NoError(t, err)
		require.NotNil(t, es)
		assert.Equal(t, "gym1", es.Location)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found returns default EventSport with markdown type", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(1, 99).
			WillReturnRows(sqlmock.NewRows(eventSportCols))

		es, err := repo.GetSportDetails(1, 99)
		require.NoError(t, err)
		require.NotNil(t, es)
		assert.Equal(t, 1, es.EventID)
		assert.Equal(t, 99, es.SportID)
		assert.Equal(t, "markdown", es.RulesType)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		es, err := repo.GetSportDetails(1, 1)
		assert.Error(t, err)
		assert.Nil(t, es)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── UpdateSportDetails ────────────────────────────────────────────────────

func TestSportRepository_UpdateSportDetails(t *testing.T) {
	const q = "UPDATE event_sports SET description = ?, rules = ?, rules_type = ?, rules_pdf_url = ?, min_capacity = ?, max_capacity = ? WHERE event_id = ? AND sport_id = ?"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		details := models.EventSport{Description: stringPtr("updated"), Rules: stringPtr("new rules"), RulesType: "markdown", MinCapacity: intPtr(5), MaxCapacity: intPtr(10)}
		mock.ExpectExec(regexp.QuoteMeta(q)).
			WithArgs(details.Description, details.Rules, details.RulesType, details.RulesPdfURL, details.MinCapacity, details.MaxCapacity, 1, 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		assert.NoError(t, repo.UpdateSportDetails(1, 1, details))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupSport(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		assert.Error(t, repo.UpdateSportDetails(1, 1, models.EventSport{}))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
