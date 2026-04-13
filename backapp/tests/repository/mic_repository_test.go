package repository_test

import (
	"database/sql"
	"testing"
	"time"

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
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM classes WHERE id = \\? AND event_id = \\? AND name IN \\('1-1', '1-2', '1-3', 'IS2', 'IT2', 'IE2'\\)").
			WithArgs(classID, eventID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
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
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM classes WHERE id = \\? AND event_id = \\? AND name IN \\('1-1', '1-2', '1-3', 'IS2', 'IT2', 'IE2'\\)").
			WithArgs(classID, eventID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
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
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM classes WHERE id = \\? AND event_id = \\? AND name IN \\('1-1', '1-2', '1-3', 'IS2', 'IT2', 'IE2'\\)").
			WithArgs(classID, eventID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery("SELECT COUNT(.+) FROM mic_votes").WithArgs(userID, eventID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectRollback()

		err := r.VoteMIC(userID, classID, eventID, reason)
		assert.Error(t, err)
		assert.Equal(t, "user has already voted", err.Error())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("class not eligible for MIC", func(t *testing.T) {
		userID := "test-user-4"
		classID := 4
		eventID := 4
		reason := ""

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM classes WHERE id = \\? AND event_id = \\? AND name IN \\('1-1', '1-2', '1-3', 'IS2', 'IT2', 'IE2'\\)").
			WithArgs(classID, eventID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectRollback()

		err := r.VoteMIC(userID, classID, eventID, reason)
		assert.Error(t, err)
		assert.Equal(t, "voted class is not eligible for MIC", err.Error())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rollback when score_log insert fails", func(t *testing.T) {
		userID := "test-user-5"
		classID := 5
		eventID := 5
		reason := "test reason"
		dbErr := sql.ErrConnDone

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM classes WHERE id = \\? AND event_id = \\? AND name IN \\('1-1', '1-2', '1-3', 'IS2', 'IT2', 'IE2'\\)").
			WithArgs(classID, eventID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery("SELECT COUNT(.+) FROM mic_votes").WithArgs(userID, eventID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec("INSERT INTO mic_votes").WithArgs(userID, classID, eventID, reason, 3).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO score_logs").WithArgs(eventID, classID, 3, "mic_points").WillReturnError(dbErr)
		mock.ExpectRollback()

		err := r.VoteMIC(userID, classID, eventID, reason)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMICRepository_GetEligibleClasses(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := repository.NewMICRepository(db)

	t.Run("Success", func(t *testing.T) {
		eventID := 1
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "1-1").
			AddRow(2, "IS2")

		mock.ExpectQuery("SELECT id, name FROM classes WHERE event_id = \\? AND name IN \\('1-1', '1-2', '1-3', 'IS2', 'IT2', 'IE2'\\)").
			WithArgs(eventID).
			WillReturnRows(rows)

		classes, err := r.GetEligibleClasses(eventID)
		assert.NoError(t, err)
		assert.Len(t, classes, 2)
		assert.Equal(t, "1-1", classes[0].Name)
		assert.Equal(t, "IS2", classes[1].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		eventID := 1
		mock.ExpectQuery("SELECT id, name FROM classes").WithArgs(eventID).WillReturnError(sql.ErrConnDone)

		classes, err := r.GetEligibleClasses(eventID)
		assert.Error(t, err)
		assert.Nil(t, classes)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMICRepository_GetMICVotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := repository.NewMICRepository(db)

	t.Run("Success", func(t *testing.T) {
		eventID := 1
		rows := sqlmock.NewRows([]string{"voted_for_class_id", "points"}).
			AddRow(1, 10).
			AddRow(2, 5)

		mock.ExpectQuery("SELECT voted_for_class_id, COUNT\\(\\*\\) as points FROM mic_votes WHERE event_id = \\? GROUP BY voted_for_class_id").
			WithArgs(eventID).
			WillReturnRows(rows)

		votes, err := r.GetMICVotes(eventID)
		assert.NoError(t, err)
		assert.Len(t, votes, 2)
		assert.Equal(t, 1, votes[0].VotedForClassID)
		assert.Equal(t, 10, votes[0].Points)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		eventID := 1
		mock.ExpectQuery("SELECT voted_for_class_id").WithArgs(eventID).WillReturnError(sql.ErrConnDone)

		votes, err := r.GetMICVotes(eventID)
		assert.Error(t, err)
		assert.Nil(t, votes)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMICRepository_GetVoteByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := repository.NewMICRepository(db)

	t.Run("Success", func(t *testing.T) {
		userID := "user1"
		eventID := 1
		now := time.Now()
		rows := sqlmock.NewRows([]string{"id", "event_id", "voter_user_id", "voted_for_class_id", "reason", "points", "created_at"}).
			AddRow(1, eventID, userID, 2, "nice", 3, now)

		mock.ExpectQuery("SELECT id, event_id, voter_user_id, voted_for_class_id, reason, points, created_at FROM mic_votes WHERE voter_user_id = \\? AND event_id = \\?").
			WithArgs(userID, eventID).
			WillReturnRows(rows)

		vote, err := r.GetVoteByUserID(userID, eventID)
		assert.NoError(t, err)
		assert.NotNil(t, vote)
		assert.Equal(t, 2, vote.VotedForClassID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No Vote Found", func(t *testing.T) {
		userID := "user2"
		eventID := 1
		mock.ExpectQuery("SELECT id").WithArgs(userID, eventID).WillReturnError(sql.ErrNoRows)

		vote, err := r.GetVoteByUserID(userID, eventID)
		assert.NoError(t, err)
		assert.Nil(t, vote)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		userID := "user3"
		eventID := 1
		mock.ExpectQuery("SELECT id").WithArgs(userID, eventID).WillReturnError(sql.ErrConnDone)

		vote, err := r.GetVoteByUserID(userID, eventID)
		assert.Error(t, err)
		assert.Nil(t, vote)
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
