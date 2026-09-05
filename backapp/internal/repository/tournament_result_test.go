package repository

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"backapp/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestDetermineMatchOutcome(t *testing.T) {
	match := &models.MatchDB{
		Team1ID: sql.NullInt64{Int64: 10, Valid: true},
		Team2ID: sql.NullInt64{Int64: 20, Valid: true},
	}

	t.Run("得点差から勝者を決定する", func(t *testing.T) {
		winnerID, loserID, err := determineMatchOutcome(match, 3, 1, 0)
		if err != nil || winnerID != 10 || loserID != 20 {
			t.Fatalf("unexpected outcome: winner=%d loser=%d err=%v", winnerID, loserID, err)
		}
	})

	t.Run("同点時の指定勝者を保持する", func(t *testing.T) {
		winnerID, loserID, err := determineMatchOutcome(match, 2, 2, 20)
		if err != nil || winnerID != 20 || loserID != 10 {
			t.Fatalf("unexpected outcome: winner=%d loser=%d err=%v", winnerID, loserID, err)
		}
	})

	t.Run("未確定の対戦は拒否する", func(t *testing.T) {
		undecided := &models.MatchDB{Team1ID: match.Team1ID}
		_, _, err := determineMatchOutcome(undecided, 1, 0, 0)
		if !errors.Is(err, ErrMatchParticipantsUndecided) {
			t.Fatalf("expected undecided error, got %v", err)
		}
	})

	t.Run("対戦外の同点勝者は拒否する", func(t *testing.T) {
		_, _, err := determineMatchOutcome(match, 1, 1, 30)
		if !errors.Is(err, ErrInvalidTieWinner) {
			t.Fatalf("expected invalid winner error, got %v", err)
		}
	})
}

func TestInvalidateBoardGameRankingsForTournament(t *testing.T) {
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

	mock.ExpectQuery(regexp.QuoteMeta("SELECT run_id FROM board_game_entries WHERE tournament_id=? LIMIT 1")).
		WithArgs(7).WillReturnRows(sqlmock.NewRows([]string{"run_id"}).AddRow(5))
	mock.ExpectExec(regexp.QuoteMeta("DELETE sl FROM score_logs sl JOIN matches m ON m.id=sl.source_match_id WHERE sl.board_game_run_id=? AND sl.reason='board_game_rank_points' AND m.tournament_id=?")).
		WithArgs(5, 7).WillReturnResult(sqlmock.NewResult(0, 4))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM board_game_rankings WHERE run_id=? AND tournament_id=?")).
		WithArgs(5, 7).WillReturnResult(sqlmock.NewResult(0, 4))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE board_game_runs SET status='published' WHERE id=? AND status='completed'")).
		WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 1))

	if err := invalidateBoardGameRankingsForTournament(tx, 7); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
