package repository_test

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"backapp/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypingIntegrationRepositoryGetCompetitionSnapshot(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewTypingIntegrationRepository(db)
	startsAt := time.Date(2026, time.September, 1, 9, 0, 0, 0, time.Local)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT
			e.id, e.name, e.year, e.season, e.status, e.start_date, e.end_date,
			s.id, s.name
		FROM event_sports es
		JOIN events e ON e.id = es.event_id
		JOIN sports s ON s.id = es.sport_id
		WHERE es.event_id = ? AND es.sport_id = ?
	`)).WithArgs(1, 2).WillReturnRows(sqlmock.NewRows([]string{
		"event_id", "event_name", "year", "season", "status", "start_date", "end_date", "sport_id", "sport_name",
	}).AddRow(1, "秋季スポーツ大会", 2026, "autumn", "active", startsAt, nil, 2, "タイピング"))

	mock.ExpectQuery("SELECT[[:space:]]+t.id, t.name, c.id, c.name").
		WithArgs(1, 2).
		WillReturnRows(sqlmock.NewRows([]string{
			"team_id", "team_name", "class_id", "class_name", "player_id", "display_name", "entry_order",
		}).
			AddRow(10, "1-A", 100, "1-A", "player-1", "選手1", 1).
			AddRow(10, "1-A", 100, "1-A", "player-2", nil, 2))

	mock.ExpectQuery("SELECT[[:space:]]+t.id, t.name,[[:space:]]+m.id").
		WithArgs(1, 2).
		WillReturnRows(sqlmock.NewRows([]string{
			"tournament_id", "tournament_name", "match_id", "round", "match_number", "status", "start_time", "end_time", "team1_id", "team2_id",
		}).AddRow(20, "タイピング", 30, 0, 0, "scheduled", startsAt, nil, 10, 11))

	snapshot, err := repo.GetCompetitionSnapshot(1, 2)
	require.NoError(t, err)
	require.NotNil(t, snapshot)
	assert.Equal(t, "秋季スポーツ大会", snapshot.Event.Name)
	assert.Equal(t, "タイピング", snapshot.Sport.Name)
	require.Len(t, snapshot.Teams, 1)
	assert.Equal(t, 2, snapshot.Teams[0].Players[1].EntryOrder)
	assert.Nil(t, snapshot.Teams[0].Players[1].Name)
	require.Len(t, snapshot.Tournaments, 1)
	require.Len(t, snapshot.Tournaments[0].Matches, 1)
	assert.Equal(t, "第1ラウンド 第1試合", snapshot.Tournaments[0].Matches[0].Name)
	assert.Equal(t, []int{10, 11}, snapshot.Tournaments[0].Matches[0].TeamIDs)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTypingIntegrationRepositoryReturnsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("FROM event_sports es").
		WithArgs(999, 2).
		WillReturnError(sql.ErrNoRows)

	repo := repository.NewTypingIntegrationRepository(db)
	snapshot, err := repo.GetCompetitionSnapshot(999, 2)
	assert.Nil(t, snapshot)
	assert.ErrorIs(t, err, repository.ErrTypingCompetitionNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTypingIntegrationRepositoryGetActiveEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("FROM active_event ae").WillReturnRows(sqlmock.NewRows([]string{
		"id", "name", "year", "season", "status", "start_date", "end_date",
	}).AddRow(1, "秋季スポーツ大会", 2026, "autumn", "active", nil, nil))
	mock.ExpectQuery("FROM event_sports es").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{
		"id", "name",
	}).AddRow(2, "タイピング"))

	repo := repository.NewTypingIntegrationRepository(db)
	event, sports, err := repo.GetActiveEvent()
	require.NoError(t, err)
	assert.Equal(t, 1, event.ID)
	require.Len(t, sports, 1)
	assert.Equal(t, 2, sports[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTypingIntegrationRepositorySetTeamEntryOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT user_id, is_confirmed").WithArgs(10).WillReturnRows(
		sqlmock.NewRows([]string{"user_id", "is_confirmed"}).
			AddRow("player-1", true).
			AddRow("player-2", true).
			AddRow("player-3", true).
			AddRow("reserve", false),
	)
	mock.ExpectExec("SET entry_order = entry_order \\+ 10000").WithArgs(10).
		WillReturnResult(sqlmock.NewResult(0, 4))
	for index, playerID := range []string{"player-3", "player-1", "player-2", "reserve"} {
		mock.ExpectExec("SET entry_order = \\?").WithArgs(index+1, 10, playerID).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()

	repo := repository.NewTypingIntegrationRepository(db)
	err = repo.SetTeamEntryOrder(10, []string{"player-3", "player-1", "player-2"})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
