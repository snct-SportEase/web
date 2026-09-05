package repository

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"backapp/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestGetSessionByEventPrefersPublishedSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &noonGameRepository{db: db}
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, event_id, template_key, name, description, scheduled_at, location, mode, win_points, loss_points, draw_points,
		       participation_points, allow_manual_points, status, created_at, updated_at
		FROM noon_game_sessions
		WHERE event_id = ?
		ORDER BY status = 'published' DESC, scheduled_at IS NULL, scheduled_at, id
		LIMIT 1
	`)).WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{
		"id", "event_id", "template_key", "name", "description", "scheduled_at", "location", "mode",
		"win_points", "loss_points", "draw_points", "participation_points", "allow_manual_points", "status", "created_at", "updated_at",
	}).AddRow(4, 7, "year_relay", "公開リレー", nil, nil, nil, "mixed", 3, 0, 1, 0, false, "published", now, now))

	session, err := repo.GetSessionByEvent(7)

	require.NoError(t, err)
	require.Equal(t, 4, session.ID)
	require.Equal(t, "published", session.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveMatchResultRollsBackWhenPointInsertFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &noonGameRepository{db: db}
	matchID := 12
	point := &models.NoonGamePoint{
		SessionID: 3,
		MatchID:   &matchID,
		ClassID:   4,
		Points:    10,
		Source:    "result",
		CreatedBy: "00000000-0000-0000-0000-000000000001",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM noon_game_points WHERE match_id = ?`)).
		WithArgs(matchID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	prepared := mock.ExpectPrepare(regexp.QuoteMeta(`
			INSERT INTO noon_game_points (session_id, match_id, class_id, points, reason, source, created_by)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`))
	prepared.ExpectExec().
		WithArgs(3, matchID, 4, 10, nil, "result", point.CreatedBy).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	err = repo.SaveMatchResult(&models.NoonGameResult{
		MatchID:    matchID,
		Winner:     "home",
		RecordedBy: point.CreatedBy,
	}, []*models.NoonGamePoint{point})

	require.ErrorContains(t, err, "insert failed")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReplaceMatchEntriesPreservesExistingIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &noonGameRepository{db: db}
	class1, class2 := 101, 102
	entries := []*models.NoonGameMatchEntry{
		{ID: 21, SideType: "class", ClassID: &class1},
		{ID: 22, SideType: "class", ClassID: &class2},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, entry_index, side_type, class_id, group_id FROM noon_game_match_entries WHERE match_id = ?`)).
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"id", "entry_index", "side_type", "class_id", "group_id"}).
			AddRow(21, 0, "class", 101, nil).
			AddRow(22, 1, "class", 102, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM noon_game_results WHERE match_id = ?)`)).
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	for index, entry := range entries {
		mock.ExpectExec(regexp.QuoteMeta(`
				UPDATE noon_game_match_entries
				SET entry_index = ?, side_type = ?, class_id = ?, group_id = ?, display_name = ?
				WHERE id = ? AND match_id = ?
			`)).
			WithArgs(index, "class", *entry.ClassID, nil, nil, entry.ID, 9).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)
	require.NoError(t, repo.replaceMatchEntriesTx(tx, 9, entries))
	require.NoError(t, tx.Commit())
	require.Equal(t, 21, entries[0].ID)
	require.Equal(t, 22, entries[1].ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReplaceMatchEntriesRejectsParticipantChangesAfterResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &noonGameRepository{db: db}
	changedClassID := 999

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, entry_index, side_type, class_id, group_id FROM noon_game_match_entries WHERE match_id = ?`)).
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"id", "entry_index", "side_type", "class_id", "group_id"}).
			AddRow(21, 0, "class", 101, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM noon_game_results WHERE match_id = ?)`)).
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)
	err = repo.replaceMatchEntriesTx(tx, 9, []*models.NoonGameMatchEntry{{ID: 21, SideType: "class", ClassID: &changedClassID}})
	require.ErrorIs(t, err, ErrNoonGameMatchParticipantsLocked)
	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}
