package repository_test

import (
	"database/sql"
	"regexp"
	"testing"

	"backapp/internal/models"
	"backapp/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestRainyModeRepository_GetSettingsByEventID(t *testing.T) {
	t.Run("Success - Get settings", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		r := repository.NewRainyModeRepository(db)

		eventID := 1

		rows := sqlmock.NewRows([]string{"id", "event_id", "sport_id", "class_id", "min_capacity", "max_capacity", "match_start_time"}).
			AddRow(1, eventID, 1, 1, 5, 10, "09:00").
			AddRow(2, eventID, 1, 2, nil, 8, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, event_id, sport_id, class_id, min_capacity, max_capacity, match_start_time
		FROM rainy_mode_settings
		WHERE event_id = ?
		ORDER BY sport_id, class_id
		`)).
			WithArgs(eventID).
			WillReturnRows(rows)

		settings, err := r.GetSettingsByEventID(eventID)

		assert.NoError(t, err)
		assert.Len(t, settings, 2)
		assert.Equal(t, 1, settings[0].ID)
		assert.Equal(t, eventID, settings[0].EventID)
		assert.Equal(t, 1, settings[0].SportID)
		assert.Equal(t, 1, settings[0].ClassID)
		assert.NotNil(t, settings[0].MinCapacity)
		assert.Equal(t, 5, *settings[0].MinCapacity)
		assert.NotNil(t, settings[0].MaxCapacity)
		assert.Equal(t, 10, *settings[0].MaxCapacity)
		assert.NotNil(t, settings[0].MatchStartTime)
		assert.Equal(t, "09:00", *settings[0].MatchStartTime)

		assert.Equal(t, 2, settings[1].ID)
		assert.Nil(t, settings[1].MinCapacity)
		assert.NotNil(t, settings[1].MaxCapacity)
		assert.Equal(t, 8, *settings[1].MaxCapacity)
		assert.Nil(t, settings[1].MatchStartTime)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - No settings found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		r := repository.NewRainyModeRepository(db)

		eventID := 999

		rows := sqlmock.NewRows([]string{"id", "event_id", "sport_id", "class_id", "min_capacity", "max_capacity", "match_start_time"})

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, event_id, sport_id, class_id, min_capacity, max_capacity, match_start_time
		FROM rainy_mode_settings
		WHERE event_id = ?
		ORDER BY sport_id, class_id
		`)).
			WithArgs(eventID).
			WillReturnRows(rows)

		settings, err := r.GetSettingsByEventID(eventID)

		assert.NoError(t, err)
		assert.Len(t, settings, 0)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRainyModeRepository_GetSetting(t *testing.T) {
	t.Run("Success - Get setting", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		r := repository.NewRainyModeRepository(db)

		eventID := 1
		sportID := 1
		classID := 1

		rows := sqlmock.NewRows([]string{"id", "event_id", "sport_id", "class_id", "min_capacity", "max_capacity", "match_start_time"}).
			AddRow(1, eventID, sportID, classID, 5, 10, "09:00")

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, event_id, sport_id, class_id, min_capacity, max_capacity, match_start_time
		FROM rainy_mode_settings
		WHERE event_id = ? AND sport_id = ? AND class_id = ?
		`)).
			WithArgs(eventID, sportID, classID).
			WillReturnRows(rows)

		setting, err := r.GetSetting(eventID, sportID, classID)

		assert.NoError(t, err)
		assert.NotNil(t, setting)
		assert.Equal(t, 1, setting.ID)
		assert.Equal(t, eventID, setting.EventID)
		assert.Equal(t, sportID, setting.SportID)
		assert.Equal(t, classID, setting.ClassID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Setting not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		r := repository.NewRainyModeRepository(db)

		eventID := 1
		sportID := 1
		classID := 999

		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, event_id, sport_id, class_id, min_capacity, max_capacity, match_start_time
		FROM rainy_mode_settings
		WHERE event_id = ? AND sport_id = ? AND class_id = ?
		`)).
			WithArgs(eventID, sportID, classID).
			WillReturnError(sql.ErrNoRows)

		setting, err := r.GetSetting(eventID, sportID, classID)

		assert.NoError(t, err)
		assert.Nil(t, setting)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRainyModeRepository_UpsertSetting(t *testing.T) {
	t.Run("Success - Insert new setting", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		r := repository.NewRainyModeRepository(db)

		eventID := 1
		sportID := 1
		classID := 1
		minCapacity := 5
		maxCapacity := 10
		matchStartTime := "09:00"

		setting := &models.RainyModeSetting{
			EventID:        eventID,
			SportID:        sportID,
			ClassID:        classID,
			MinCapacity:    &minCapacity,
			MaxCapacity:    &maxCapacity,
			MatchStartTime: &matchStartTime,
		}

		mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO rainy_mode_settings (event_id, sport_id, class_id, min_capacity, max_capacity, match_start_time)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			min_capacity = VALUES(min_capacity),
			max_capacity = VALUES(max_capacity),
			match_start_time = VALUES(match_start_time)
		`)).
			WithArgs(eventID, sportID, classID, minCapacity, maxCapacity, matchStartTime).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = r.UpsertSetting(setting)

		assert.NoError(t, err)
		assert.Equal(t, 1, setting.ID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Update existing setting", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		r := repository.NewRainyModeRepository(db)

		eventID := 1
		sportID := 1
		classID := 1
		maxCapacity := 15

		setting := &models.RainyModeSetting{
			ID:          1,
			EventID:     eventID,
			SportID:     sportID,
			ClassID:     classID,
			MaxCapacity: &maxCapacity,
		}

		mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO rainy_mode_settings (event_id, sport_id, class_id, min_capacity, max_capacity, match_start_time)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			min_capacity = VALUES(min_capacity),
			max_capacity = VALUES(max_capacity),
			match_start_time = VALUES(match_start_time)
		`)).
			WithArgs(eventID, sportID, classID, nil, maxCapacity, nil).
			WillReturnResult(sqlmock.NewResult(0, 1)) // Updated, not inserted

		err = r.UpsertSetting(setting)

		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRainyModeRepository_DeleteSetting(t *testing.T) {
	t.Run("Success - Delete setting", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		r := repository.NewRainyModeRepository(db)

		eventID := 1
		sportID := 1
		classID := 1

		mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM rainy_mode_settings
		WHERE event_id = ? AND sport_id = ? AND class_id = ?
		`)).
			WithArgs(eventID, sportID, classID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = r.DeleteSetting(eventID, sportID, classID)

		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
