package repository

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRewriteBoardGameRosterMapsShogiSelectionOrder(t *testing.T) {
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

	entries := []boardGameEntryTeam{
		{entryID: 11, teamID: 21, slotKey: "A"},
		{entryID: 12, teamID: 22, slotKey: "B"},
	}
	for index, entry := range entries {
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM board_game_entry_members WHERE entry_id=?")).WithArgs(entry.entryID).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM team_members WHERE team_id=?")).WithArgs(entry.teamID).WillReturnResult(sqlmock.NewResult(0, 0))
		playerID := []string{"player-a", "player-b"}[index]
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO board_game_entry_members (entry_id,user_id,member_order,is_substitute) VALUES (?,?,?,?)")).WithArgs(entry.entryID, playerID, 0, false).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO team_members (team_id,user_id,is_confirmed) VALUES (?,?,TRUE)")).WithArgs(entry.teamID, playerID).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO board_game_entry_members (entry_id,user_id,member_order,is_substitute) VALUES (?,?,?,?)")).WithArgs(entry.entryID, "substitute", 1, true).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO team_members (team_id,user_id,is_confirmed) VALUES (?,?,TRUE)")).WithArgs(entry.teamID, "substitute").WillReturnResult(sqlmock.NewResult(0, 1))
	}

	if err := rewriteBoardGameRoster(tx, "shogi", entries, []string{"player-a", "player-b", "substitute"}); err != nil {
		t.Fatal(err)
	}
	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRewriteBoardGameRosterRejectsMoreThanThreeMembers(t *testing.T) {
	if err := rewriteBoardGameRoster(nil, "othello", nil, []string{"1", "2", "3", "4"}); !errors.Is(err, ErrBoardGameRosterFull) {
		t.Fatalf("expected ErrBoardGameRosterFull, got %v", err)
	}
}
