package repository

import (
	"regexp"
	"testing"

	"backapp/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestValidateBoardGameRankingOrder(t *testing.T) {
	query := regexp.QuoteMeta(`SELECT team1_id,team2_id,COALESCE(winner_team_id,CASE WHEN team1_score>team2_score THEN team1_id WHEN team2_score>team1_score THEN team2_id END)
			FROM matches WHERE tournament_id=? AND is_bronze_match=? ORDER BY round DESC,match_number_in_round LIMIT 1`)
	tests := []struct {
		name      string
		inputs    []models.BoardGameRankingInput
		wantError bool
	}{
		{
			name: "決勝と3位決定戦どおりの順位を受理する",
			inputs: []models.BoardGameRankingInput{
				{EntryID: 101, Rank: 1},
				{EntryID: 102, Rank: 2},
				{EntryID: 104, Rank: 3},
				{EntryID: 103, Rank: 4},
			},
		},
		{
			name: "試合結果と異なる順位を拒否する",
			inputs: []models.BoardGameRankingInput{
				{EntryID: 102, Rank: 1},
				{EntryID: 101, Rank: 2},
				{EntryID: 104, Rank: 3},
				{EntryID: 103, Rank: 4},
			},
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectBegin()
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			mock.ExpectQuery(query).WithArgs(7, false).
				WillReturnRows(sqlmock.NewRows([]string{"team1_id", "team2_id", "winner_team_id"}).AddRow(10, 20, 10))
			mock.ExpectQuery(query).WithArgs(7, true).
				WillReturnRows(sqlmock.NewRows([]string{"team1_id", "team2_id", "winner_team_id"}).AddRow(30, 40, 40))
			mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id,id FROM board_game_entries WHERE run_id=? AND tournament_id=?")).
				WithArgs(5, 7).
				WillReturnRows(sqlmock.NewRows([]string{"team_id", "id"}).
					AddRow(10, 101).AddRow(20, 102).AddRow(30, 103).AddRow(40, 104))

			err = validateBoardGameRankingOrder(tx, 5, 7, 4, tt.inputs)
			if (err != nil) != tt.wantError {
				t.Fatalf("unexpected error: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestDeleteBoardGameRunDataDeletesOnlyOwnedTournaments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id FROM board_game_entries WHERE run_id=?")).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"team_id"}).AddRow(10).AddRow(20))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM score_logs WHERE board_game_run_id=?")).
		WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT tournament_id FROM board_game_entries WHERE run_id=?")).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"tournament_id"}).AddRow(100).AddRow(200))
	for _, tournamentID := range []int{100, 200} {
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM matches WHERE tournament_id=?")).
			WithArgs(tournamentID).WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM tournaments WHERE id=? AND event_id=? AND sport_id=?")).
			WithArgs(tournamentID, 7, 9).WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM board_game_runs WHERE id=?")).
		WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 1))
	for _, teamID := range []int{10, 20} {
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM teams WHERE id=?")).
			WithArgs(teamID).WillReturnResult(sqlmock.NewResult(0, 1))
	}

	if err := deleteBoardGameRunData(tx, 5, 7, 9); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
