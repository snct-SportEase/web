package repository_test

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"backapp/internal/models"
	"backapp/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// eventCols lists all columns returned by event SELECT queries.
var eventCols = []string{
	"id", "name", "year", "season", "start_date", "end_date",
	"is_rainy_mode", "competition_guidelines_pdf_url", "survey_url",
	"is_survey_published", "status", "hide_scores",
}

func newEvent() *models.Event {
	start := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 4, 2, 0, 0, 0, 0, time.UTC)
	return &models.Event{
		ID: 1, Name: "Spring 2024", Year: 2024, Season: "spring",
		Start_date: &start, End_date: &end,
		IsRainyMode: false, IsSurveyPublished: false,
		Status: "active", HideScores: false,
	}
}

func eventRow(e *models.Event) *sqlmock.Rows {
	return sqlmock.NewRows(eventCols).AddRow(
		e.ID, e.Name, e.Year, e.Season, e.Start_date, e.End_date,
		e.IsRainyMode, nil, nil, e.IsSurveyPublished, e.Status, e.HideScores,
	)
}

func setupEvent(t *testing.T) (repository.EventRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return repository.NewEventRepository(db), mock, func() { db.Close() }
}

// ─── GetEventByID ──────────────────────────────────────────────────────────

func TestEventRepository_GetEventByID(t *testing.T) {
	const q = "SELECT id, name, `year`, season, start_date, end_date, is_rainy_mode, competition_guidelines_pdf_url, survey_url, is_survey_published, status, hide_scores FROM events WHERE id = ?"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		e := newEvent()
		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(1).WillReturnRows(eventRow(e))

		got, err := repo.GetEventByID(1)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, e.ID, got.ID)
		assert.Equal(t, e.Name, got.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(999).WillReturnError(sql.ErrNoRows)

		got, err := repo.GetEventByID(999)
		assert.NoError(t, err)
		assert.Nil(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(1).WillReturnError(errors.New("db error"))

		got, err := repo.GetEventByID(1)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── GetAllEvents ──────────────────────────────────────────────────────────

func TestEventRepository_GetAllEvents(t *testing.T) {
	const q = "SELECT id, name, `year`, season, start_date, end_date, is_rainy_mode, competition_guidelines_pdf_url, survey_url, is_survey_published, status, hide_scores FROM events ORDER BY `year` DESC, FIELD(season, 'autumn', 'spring')"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		e := newEvent()
		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnRows(eventRow(e))

		got, err := repo.GetAllEvents()
		require.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, e.ID, got[0].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnRows(sqlmock.NewRows(eventCols))

		got, err := repo.GetAllEvents()
		assert.NoError(t, err)
		assert.Nil(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		got, err := repo.GetAllEvents()
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── GetActiveEvent ────────────────────────────────────────────────────────

func TestEventRepository_GetActiveEvent(t *testing.T) {
	const q = "SELECT event_id FROM active_event WHERE id = 1"

	t.Run("returns active event id", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WillReturnRows(sqlmock.NewRows([]string{"event_id"}).AddRow(5))

		id, err := repo.GetActiveEvent()
		assert.NoError(t, err)
		assert.Equal(t, 5, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns 0 when event_id is NULL", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).
			WillReturnRows(sqlmock.NewRows([]string{"event_id"}).AddRow(nil))

		id, err := repo.GetActiveEvent()
		assert.NoError(t, err)
		assert.Equal(t, 0, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns 0 when no row exists", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(sql.ErrNoRows)

		id, err := repo.GetActiveEvent()
		assert.NoError(t, err)
		assert.Equal(t, 0, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		id, err := repo.GetActiveEvent()
		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── GetEventByYearAndSeason ───────────────────────────────────────────────

func TestEventRepository_GetEventByYearAndSeason(t *testing.T) {
	const q = "SELECT id, name, `year`, season, start_date, end_date, is_rainy_mode, competition_guidelines_pdf_url, survey_url, is_survey_published, status FROM events WHERE `year` = ? AND season = ?"
	var yasCols = []string{"id", "name", "year", "season", "start_date", "end_date", "is_rainy_mode", "competition_guidelines_pdf_url", "survey_url", "is_survey_published", "status"}

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		now := time.Now()
		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(2024, "spring").
			WillReturnRows(sqlmock.NewRows(yasCols).AddRow(1, "Spring 2024", 2024, "spring", now, now, false, nil, nil, false, "active"))

		got, err := repo.GetEventByYearAndSeason(2024, "spring")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 1, got.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(2000, "spring").WillReturnError(sql.ErrNoRows)

		got, err := repo.GetEventByYearAndSeason(2000, "spring")
		assert.NoError(t, err)
		assert.Nil(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(2024, "spring").WillReturnError(errors.New("db error"))

		got, err := repo.GetEventByYearAndSeason(2024, "spring")
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── CreateEvent ───────────────────────────────────────────────────────────

func TestEventRepository_CreateEvent(t *testing.T) {
	const insertQ = "INSERT INTO events (name, `year`, season, start_date, end_date, is_rainy_mode, competition_guidelines_pdf_url, survey_url, is_survey_published, status, hide_scores) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	const archiveQ = "UPDATE events SET status = 'archived' WHERE id != ? AND status = 'active'"
	const activeQ = "INSERT INTO active_event (id, event_id) VALUES (1, ?) ON DUPLICATE KEY UPDATE event_id = VALUES(event_id)"

	t.Run("success - non-active status", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		e := newEvent()
		e.Status = "upcoming"

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(insertQ)).
			WithArgs(e.Name, e.Year, e.Season, e.Start_date, e.End_date, e.IsRainyMode, nil, nil, e.IsSurveyPublished, e.Status, e.HideScores).
			WillReturnResult(sqlmock.NewResult(10, 1))
		mock.ExpectCommit()

		id, err := repo.CreateEvent(e)
		assert.NoError(t, err)
		assert.Equal(t, int64(10), id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - active status archives others and sets active_event", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		e := newEvent()
		e.Status = "active"

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(insertQ)).
			WithArgs(e.Name, e.Year, e.Season, e.Start_date, e.End_date, e.IsRainyMode, nil, nil, e.IsSurveyPublished, e.Status, e.HideScores).
			WillReturnResult(sqlmock.NewResult(10, 1))
		mock.ExpectExec(regexp.QuoteMeta(archiveQ)).WithArgs(int64(10)).WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectExec(regexp.QuoteMeta(activeQ)).WithArgs(int64(10)).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		id, err := repo.CreateEvent(e)
		assert.NoError(t, err)
		assert.Equal(t, int64(10), id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - insert fails", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		e := newEvent()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(insertQ)).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		id, err := repo.CreateEvent(e)
		assert.Error(t, err)
		assert.Equal(t, int64(0), id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - archive fails", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		e := newEvent()
		e.Status = "active"
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(insertQ)).
			WillReturnResult(sqlmock.NewResult(10, 1))
		mock.ExpectExec(regexp.QuoteMeta(archiveQ)).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		id, err := repo.CreateEvent(e)
		assert.Error(t, err)
		assert.Equal(t, int64(0), id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── UpdateEvent ───────────────────────────────────────────────────────────

func TestEventRepository_UpdateEvent(t *testing.T) {
	const updateQ = "UPDATE events SET name = ?, `year` = ?, season = ?, start_date = ?, end_date = ?, is_rainy_mode = ?, competition_guidelines_pdf_url = ?, survey_url = ?, is_survey_published = ?, status = ?, hide_scores = ? WHERE id = ?"
	const archiveQ = "UPDATE events SET status = 'archived' WHERE id != ? AND status = 'active'"
	const activeQ = "INSERT INTO active_event (id, event_id) VALUES (1, ?) ON DUPLICATE KEY UPDATE event_id = VALUES(event_id)"
	const clearQ = "UPDATE active_event SET event_id = NULL WHERE id = 1 AND event_id = ?"

	t.Run("success - active status archives others", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		e := newEvent()
		e.Status = "active"

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(updateQ)).
			WithArgs(e.Name, e.Year, e.Season, e.Start_date, e.End_date, e.IsRainyMode, nil, nil, e.IsSurveyPublished, e.Status, e.HideScores, e.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(regexp.QuoteMeta(archiveQ)).WithArgs(e.ID).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(regexp.QuoteMeta(activeQ)).WithArgs(e.ID).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateEvent(e)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - non-active status clears active_event", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		e := newEvent()
		e.Status = "upcoming"

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(updateQ)).
			WithArgs(e.Name, e.Year, e.Season, e.Start_date, e.End_date, e.IsRainyMode, nil, nil, e.IsSurveyPublished, e.Status, e.HideScores, e.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(regexp.QuoteMeta(clearQ)).WithArgs(e.ID).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateEvent(e)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - update fails", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		e := newEvent()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(updateQ)).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.UpdateEvent(e)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── SetActiveEvent ────────────────────────────────────────────────────────

func TestEventRepository_SetActiveEvent(t *testing.T) {
	const upsertQ = "INSERT INTO active_event (id, event_id) VALUES (1, ?) ON DUPLICATE KEY UPDATE event_id = VALUES(event_id)"
	const archiveOthersQ = "UPDATE events SET status = 'archived' WHERE id != ? AND status = 'active'"
	const activateQ = "UPDATE events SET status = 'active' WHERE id = ?"
	const archiveAllQ = "UPDATE events SET status = 'archived' WHERE status = 'active'"

	t.Run("success - set specific event", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		eventID := 3
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(upsertQ)).WithArgs(eventID).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(regexp.QuoteMeta(archiveOthersQ)).WithArgs(eventID).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(regexp.QuoteMeta(activateQ)).WithArgs(eventID).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.SetActiveEvent(&eventID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - clear active event (nil)", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(upsertQ)).WithArgs(nil).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(regexp.QuoteMeta(archiveAllQ)).WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		err := repo.SetActiveEvent(nil)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback - upsert fails", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		eventID := 3
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(upsertQ)).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.SetActiveEvent(&eventID)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── CopyClassScores ───────────────────────────────────────────────────────

func TestEventRepository_CopyClassScores(t *testing.T) {
	const deleteQ = "DELETE FROM score_logs WHERE event_id = ? AND reason = 'initial_points'"
	const insertQ = `
			INSERT INTO score_logs (event_id, class_id, points, reason)
			SELECT ?, class_id, total_points_current_event, 'initial_points'
			FROM class_scores
			WHERE event_id = ? AND total_points_current_event > 0
		`

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(deleteQ)).WithArgs(2).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(regexp.QuoteMeta(insertQ)).WithArgs(2, 1).WillReturnResult(sqlmock.NewResult(0, 5))

		err := repo.CopyClassScores(1, 2)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error on delete", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(deleteQ)).WithArgs(2).WillReturnError(errors.New("db error"))

		err := repo.CopyClassScores(1, 2)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error on insert", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(deleteQ)).WithArgs(2).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(regexp.QuoteMeta(insertQ)).WithArgs(2, 1).WillReturnError(errors.New("db error"))

		err := repo.CopyClassScores(1, 2)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── SetRainyMode ──────────────────────────────────────────────────────────

func TestEventRepository_SetRainyMode(t *testing.T) {
	const q = "UPDATE events SET is_rainy_mode = ? WHERE id = ?"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WithArgs(true, 1).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.SetRainyMode(1, true)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		err := repo.SetRainyMode(1, true)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ─── PublishSurvey ─────────────────────────────────────────────────────────

func TestEventRepository_PublishSurvey(t *testing.T) {
	const q = "UPDATE events SET is_survey_published = TRUE WHERE id = ?"

	t.Run("success", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.PublishSurvey(1)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock, close := setupEvent(t)
		defer close()

		mock.ExpectExec(regexp.QuoteMeta(q)).WillReturnError(errors.New("db error"))

		err := repo.PublishSurvey(1)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
