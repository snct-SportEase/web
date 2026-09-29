package repository_test

import (
	"errors"
	"regexp"
	"testing"

	"backapp/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func expectSportSettingsLock(mock sqlmock.Sqlmock, location string, templateKey any) {
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM events WHERE id = ? FOR UPDATE")).WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectQuery("SELECT es.location, es.template_key, s.name").WithArgs(2, 7).
		WillReturnRows(sqlmock.NewRows([]string{"location", "template_key", "name"}).AddRow(location, templateKey, "バスケットボール"))
}

func TestUpdateSportLocationAndDescriptionPreservesAssignments(t *testing.T) {
	repo, mock, close := setupSport(t)
	defer close()
	expectSportSettingsLock(mock, "gym1", nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM event_sports WHERE event_id = ? AND location = ? AND sport_id <> ?")).WithArgs(2, "gym2", 7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	description := "新しい概要"
	mock.ExpectExec(regexp.QuoteMeta("UPDATE event_sports SET location = ?, description = ? WHERE event_id = ? AND sport_id = ?")).
		WithArgs("gym2", description, 2, 7).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.UpdateSportLocationAndDescription(2, 7, "gym2", &description))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateSportLocationAndDescriptionRejectsOccupiedGym(t *testing.T) {
	repo, mock, close := setupSport(t)
	defer close()
	expectSportSettingsLock(mock, "gym1", nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM event_sports WHERE event_id = ? AND location = ? AND sport_id <> ?")).WithArgs(2, "gym2", 7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectRollback()

	require.ErrorIs(t, repo.UpdateSportLocationAndDescription(2, 7, "gym2", nil), repository.ErrSportLocationInUse)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateSportLocationAndDescriptionAllowsSharedCustomLocation(t *testing.T) {
	repo, mock, close := setupSport(t)
	defer close()
	expectSportSettingsLock(mock, "gym1", nil)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE event_sports SET location = ?, description = ? WHERE event_id = ? AND sport_id = ?")).
		WithArgs("other:中庭", nil, 2, 7).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.UpdateSportLocationAndDescription(2, 7, "other:中庭", nil))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateSportLocationAndDescriptionRejectsManagedTemplate(t *testing.T) {
	repo, mock, close := setupSport(t)
	defer close()
	expectSportSettingsLock(mock, "other", "board_game_tournament")
	mock.ExpectRollback()

	require.ErrorIs(t, repo.UpdateSportLocationAndDescription(2, 7, "gym2", nil), repository.ErrSportManagedInOtherScreen)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateSportLocationAndDescriptionRollsBackOnUpdateError(t *testing.T) {
	repo, mock, close := setupSport(t)
	defer close()
	expectSportSettingsLock(mock, "gym1", nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM event_sports WHERE event_id = ? AND location = ? AND sport_id <> ?")).WithArgs(2, "gym2", 7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	writeErr := errors.New("write failed")
	mock.ExpectExec(regexp.QuoteMeta("UPDATE event_sports SET location = ?, description = ? WHERE event_id = ? AND sport_id = ?")).
		WithArgs("gym2", nil, 2, 7).WillReturnError(writeErr)
	mock.ExpectRollback()

	require.ErrorIs(t, repo.UpdateSportLocationAndDescription(2, 7, "gym2", nil), writeErr)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateSportLocationAndDescriptionRejectsInvalidLocation(t *testing.T) {
	repo, mock, close := setupSport(t)
	defer close()
	require.ErrorIs(t, repo.UpdateSportLocationAndDescription(2, 7, "other: ", nil), repository.ErrInvalidSportLocation)
	require.NoError(t, mock.ExpectationsWereMet())
}
