package repository_test

import (
	"database/sql"
	"testing"

	"backapp/internal/models"
	"backapp/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestMICRepository_VoteMIC(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := repository.NewMICRepository(db)

	t.Run("Success - Spring", func(t *testing.T) {
		userID := "test-user"
		classID := 1
		eventID := 1
		reason := "test reason"

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COUNT(.+) FROM mic_votes").WithArgs(userID, eventID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec("INSERT INTO mic_votes").WithArgs(userID, classID, eventID, reason, 3).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO score_logs").WithArgs(eventID, classID, 3, "mic_points").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := r.VoteMIC(userID, classID, eventID, reason)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Autumn", func(t *testing.T) {
		userID := "test-user-2"
		classID := 2
		eventID := 2
		reason := ""

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COUNT(.+) FROM mic_votes").WithArgs(userID, eventID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec("INSERT INTO mic_votes").WithArgs(userID, classID, eventID, reason, 3).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO score_logs").WithArgs(eventID, classID, 3, "mic_points").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := r.VoteMIC(userID, classID, eventID, reason)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Already Voted", func(t *testing.T) {
		userID := "test-user-3"
		classID := 3
		eventID := 3
		reason := ""

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COUNT(.+) FROM mic_votes").WithArgs(userID, eventID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectRollback()

		err := r.VoteMIC(userID, classID, eventID, reason)
		assert.Error(t, err)
		assert.Equal(t, "user has already voted", err.Error())
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMICRepository_GetMICClass(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := repository.NewMICRepository(db)

	t.Run("Success - Spring", func(t *testing.T) {
		eventID := 1
		expectedResult := &models.MICResult{
			ClassName:   "Test Class 1",
			TotalPoints: 100,
			Season:      "spring",
			EventID:     eventID,
		}

		mock.ExpectQuery("SELECT season FROM events").WithArgs(eventID).WillReturnRows(sqlmock.NewRows([]string{"season"}).AddRow("spring"))
		mock.ExpectQuery(`SELECT c.name, \(cs.total_points_overall \+ cs.mic_points\) AS total_points`).WithArgs(eventID).WillReturnRows(sqlmock.NewRows([]string{"name", "total_points"}).AddRow(expectedResult.ClassName, expectedResult.TotalPoints))

		result, err := r.GetMICClass(eventID)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Autumn", func(t *testing.T) {
		eventID := 2
		expectedResult := &models.MICResult{
			ClassName:   "Test Class 2",
			TotalPoints: 250,
			Season:      "autumn",
			EventID:     eventID,
		}

		mock.ExpectQuery("SELECT season FROM events").WithArgs(eventID).WillReturnRows(sqlmock.NewRows([]string{"season"}).AddRow("autumn"))
		mock.ExpectQuery(`SELECT c.name, \(cs.total_points_overall \+ cs.mic_points\) AS total_points`).WithArgs(eventID).WillReturnRows(sqlmock.NewRows([]string{"name", "total_points"}).AddRow(expectedResult.ClassName, expectedResult.TotalPoints))

		result, err := r.GetMICClass(eventID)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Event Not Found", func(t *testing.T) {
		eventID := 99

		mock.ExpectQuery("SELECT season FROM events").WithArgs(eventID).WillReturnError(sql.ErrNoRows)

		result, err := r.GetMICClass(eventID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "event not found", err.Error())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No MIC Class Found", func(t *testing.T) {
		eventID := 3

		mock.ExpectQuery("SELECT season FROM events").WithArgs(eventID).WillReturnRows(sqlmock.NewRows([]string{"season"}).AddRow("spring"))
		mock.ExpectQuery(`SELECT c.name, \(cs.total_points_overall \+ cs.mic_points\) AS total_points`).WithArgs(eventID).WillReturnError(sql.ErrNoRows)

		result, err := r.GetMICClass(eventID)
		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
