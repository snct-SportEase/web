package repository_test

import (
	"errors"
	"testing"

	"backapp/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func seedRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "round", "team1_id", "team2_id", "team1_score", "team2_score", "winner_team_id", "status", "is_loser_bracket_match", "team1_name", "team2_name"}).
		AddRow(11, 0, 1, 2, nil, nil, nil, "pending", false, "A", "B").
		AddRow(12, 0, 3, 4, nil, nil, nil, "pending", false, "C", "D").
		AddRow(13, 1, nil, nil, nil, nil, nil, "pending", false, nil, nil)
}

func expectSeedRead(mock sqlmock.Sqlmock, matches *sqlmock.Rows) {
	mock.ExpectQuery("SELECT es.template_key, s.name, es.location").WithArgs(2, 22).
		WillReturnRows(sqlmock.NewRows([]string{"template_key", "name", "location"}).AddRow(nil, "バスケットボール", "gym2"))
	mock.ExpectQuery("SELECT m.id, m.round, m.team1_id").WithArgs(22).WillReturnRows(matches)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM score_logs").WithArgs(22).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
}

func TestUpdateTournamentSeedOrderChangesOnlyFirstRoundMatches(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	expectSeedRead(mock, seedRows())
	mock.ExpectExec("UPDATE matches SET team1_id").WithArgs(int64(4), int64(3), 11, 22).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE matches SET team1_id").WithArgs(int64(2), int64(1), 12, 22).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := repository.NewTournamentRepository(db)
	require.NoError(t, repo.UpdateTournamentSeedOrder(2, 22, []int{1, 2, 3, 4}, []int{4, 3, 2, 1}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTournamentSeedOrderRejectsMissingOrRepeatedTeams(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	expectSeedRead(mock, seedRows())
	mock.ExpectRollback()

	repo := repository.NewTournamentRepository(db)
	require.ErrorIs(t, repo.UpdateTournamentSeedOrder(2, 22, []int{1, 2, 3, 4}, []int{1, 2, 2, 4}), repository.ErrInvalidTournamentSeed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTournamentSeedOrderRejectsStartedMatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	matches := sqlmock.NewRows([]string{"id", "round", "team1_id", "team2_id", "team1_score", "team2_score", "winner_team_id", "status", "is_loser_bracket_match", "team1_name", "team2_name"}).
		AddRow(11, 0, 1, 2, 3, 1, 1, "finished", false, "A", "B")
	expectSeedRead(mock, matches)
	mock.ExpectRollback()

	repo := repository.NewTournamentRepository(db)
	require.ErrorIs(t, repo.UpdateTournamentSeedOrder(2, 22, []int{1, 2}, []int{2, 1}), repository.ErrTournamentSeedLocked)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTournamentSeedOrderRejectsStaleView(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	expectSeedRead(mock, seedRows())
	mock.ExpectRollback()

	repo := repository.NewTournamentRepository(db)
	require.ErrorIs(t, repo.UpdateTournamentSeedOrder(2, 22, []int{2, 1, 3, 4}, []int{4, 3, 2, 1}), repository.ErrTournamentSeedChanged)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTournamentSeedOrderRollsBackIfSecondMatchFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	expectSeedRead(mock, seedRows())
	mock.ExpectExec("UPDATE matches SET team1_id").WithArgs(int64(4), int64(3), 11, 22).WillReturnResult(sqlmock.NewResult(0, 1))
	writeErr := errors.New("write failed")
	mock.ExpectExec("UPDATE matches SET team1_id").WithArgs(int64(2), int64(1), 12, 22).WillReturnError(writeErr)
	mock.ExpectRollback()

	repo := repository.NewTournamentRepository(db)
	require.ErrorIs(t, repo.UpdateTournamentSeedOrder(2, 22, []int{1, 2, 3, 4}, []int{4, 3, 2, 1}), writeErr)
	require.NoError(t, mock.ExpectationsWereMet())
}
