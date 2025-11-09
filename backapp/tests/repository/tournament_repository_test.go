package repository_test

import (
	"regexp"
	"testing"

	"backapp/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestTournamentRepository_UpdateMatchResult(t *testing.T) {
	t.Run("Success - Update match and advance winner", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		r := repository.NewTournamentRepository(db)

		matchID := 1
		team1ID := int64(1)
		team2ID := int64(2)
		nextMatchID := int64(2)
		tournamentID := 1
		team1Score := 2
		team2Score := 1
		winnerID := team1ID

		mock.ExpectBegin()

		// Mock getMatchByID for the current match
		rows := sqlmock.NewRows([]string{"id", "tournament_id", "round", "match_number_in_round", "team1_id", "team2_id", "winner_team_id", "status", "next_match_id", "start_time", "is_bronze_match"}).
			AddRow(matchID, tournamentID, 1, 1, team1ID, team2ID, nil, "inprogress", nextMatchID, "", false)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tournament_id, round, match_number_in_round, team1_id, team2_id, winner_team_id, status, next_match_id, start_time, is_bronze_match FROM matches WHERE id = ?")).
			WithArgs(matchID).WillReturnRows(rows)

		// Mock update current match
		mock.ExpectExec(regexp.QuoteMeta("UPDATE matches SET team1_score = ?, team2_score = ?, winner_team_id = ?, status = 'finished' WHERE id = ?")).
			WithArgs(team1Score, team2Score, winnerID, matchID).WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock getMatchByID for the next match
		nextMatchRows := sqlmock.NewRows([]string{"id", "tournament_id", "round", "match_number_in_round", "team1_id", "team2_id", "winner_team_id", "status", "next_match_id", "start_time", "is_bronze_match"}).
			AddRow(nextMatchID, tournamentID, 2, 1, nil, nil, nil, "pending", nil, "", false)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tournament_id, round, match_number_in_round, team1_id, team2_id, winner_team_id, status, next_match_id, start_time, is_bronze_match FROM matches WHERE id = ?")).
			WithArgs(nextMatchID).WillReturnRows(nextMatchRows)

		// Mock update next match
		mock.ExpectExec(regexp.QuoteMeta("UPDATE matches SET team1_id = ? WHERE id = ?")).
			WithArgs(winnerID, nextMatchID).WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock for bronze match logic (not a semi-final)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT MAX(round) FROM matches WHERE tournament_id = ?")).WithArgs(tournamentID).WillReturnRows(sqlmock.NewRows([]string{"MAX(round)"}).AddRow(3))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT t.event_id, t.sport_id, es.location FROM tournaments t LEFT JOIN event_sports es ON es.event_id = t.event_id AND es.sport_id = t.sport_id WHERE t.id = ?")).
			WithArgs(tournamentID).
			WillReturnRows(sqlmock.NewRows([]string{"event_id", "sport_id", "location"}).AddRow(1, 1, "gym1"))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, class_id, sport_id, event_id FROM teams WHERE id = ?")).
			WithArgs(winnerID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "class_id", "sport_id", "event_id"}).AddRow(winnerID, "Winner Team", 101, 1, 1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, class_id, sport_id, event_id FROM teams WHERE id = ?")).
			WithArgs(team2ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "class_id", "sport_id", "event_id"}).AddRow(team2ID, "Loser Team", 102, 1, 1))

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO class_scores (event_id, class_id, gym1_win2_points) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE gym1_win2_points = gym1_win2_points + VALUES(gym1_win2_points)")).
			WithArgs(1, 101, 10).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT season FROM events WHERE id = ?")).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"season"}).AddRow("spring"))

		mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE class_scores cs
		JOIN (
			SELECT
				class_id,
				RANK() OVER (ORDER BY total_points_current_event DESC) AS new_rank
			FROM class_scores
			WHERE event_id = ?
		) ranked_data ON cs.class_id = ranked_data.class_id
		SET cs.rank_current_event = ranked_data.new_rank
		WHERE cs.event_id = ?
		`)).
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err = r.UpdateMatchResult(matchID, team1Score, team2Score, int(winnerID))
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Semi-final, advance winner and loser to bronze match", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		r := repository.NewTournamentRepository(db)

		matchID := 3
		team1ID := int64(3)
		team2ID := int64(4)
		nextMatchID := int64(5) // Final match
		bronzeMatchID := int64(6)
		tournamentID := 1
		team1Score := 1
		team2Score := 2
		winnerID := team2ID
		loserID := team1ID

		mock.ExpectBegin()

		// Mock getMatchByID for the current match (semi-final)
		rows := sqlmock.NewRows([]string{"id", "tournament_id", "round", "match_number_in_round", "team1_id", "team2_id", "winner_team_id", "status", "next_match_id", "start_time", "is_bronze_match"}).
			AddRow(matchID, tournamentID, 2, 1, team1ID, team2ID, nil, "inprogress", nextMatchID, "", false)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tournament_id, round, match_number_in_round, team1_id, team2_id, winner_team_id, status, next_match_id, start_time, is_bronze_match FROM matches WHERE id = ?")).
			WithArgs(matchID).WillReturnRows(rows)

		// Mock update current match
		mock.ExpectExec(regexp.QuoteMeta("UPDATE matches SET team1_score = ?, team2_score = ?, winner_team_id = ?, status = 'finished' WHERE id = ?")).
			WithArgs(team1Score, team2Score, winnerID, matchID).WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock getMatchByID for the next match (final)
		nextMatchRows := sqlmock.NewRows([]string{"id", "tournament_id", "round", "match_number_in_round", "team1_id", "team2_id", "winner_team_id", "status", "next_match_id", "start_time", "is_bronze_match"}).
			AddRow(nextMatchID, tournamentID, 3, 1, nil, nil, nil, "pending", nil, "", false)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tournament_id, round, match_number_in_round, team1_id, team2_id, winner_team_id, status, next_match_id, start_time, is_bronze_match FROM matches WHERE id = ?")).
			WithArgs(nextMatchID).WillReturnRows(nextMatchRows)

		// Mock update next match
		mock.ExpectExec(regexp.QuoteMeta("UPDATE matches SET team1_id = ? WHERE id = ?")).
			WithArgs(winnerID, nextMatchID).WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock for bronze match logic (is a semi-final)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT MAX(round) FROM matches WHERE tournament_id = ?")).WithArgs(tournamentID).WillReturnRows(sqlmock.NewRows([]string{"MAX(round)"}).AddRow(3))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM matches WHERE tournament_id = ? AND is_bronze_match = TRUE")).WithArgs(tournamentID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(bronzeMatchID))

		// Mock getMatchByID for the bronze match
		bronzeMatchRows := sqlmock.NewRows([]string{"id", "tournament_id", "round", "match_number_in_round", "team1_id", "team2_id", "winner_team_id", "status", "next_match_id", "start_time", "is_bronze_match"}).
			AddRow(bronzeMatchID, tournamentID, 3, 2, nil, nil, nil, "pending", nil, "", true)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tournament_id, round, match_number_in_round, team1_id, team2_id, winner_team_id, status, next_match_id, start_time, is_bronze_match FROM matches WHERE id = ?")).
			WithArgs(bronzeMatchID).WillReturnRows(bronzeMatchRows)

		// Mock update bronze match
		mock.ExpectExec(regexp.QuoteMeta("UPDATE matches SET team1_id = ? WHERE id = ?")).
			WithArgs(loserID, bronzeMatchID).WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT t.event_id, t.sport_id, es.location FROM tournaments t LEFT JOIN event_sports es ON es.event_id = t.event_id AND es.sport_id = t.sport_id WHERE t.id = ?")).
			WithArgs(tournamentID).
			WillReturnRows(sqlmock.NewRows([]string{"event_id", "sport_id", "location"}).AddRow(1, 1, "gym1"))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, class_id, sport_id, event_id FROM teams WHERE id = ?")).
			WithArgs(winnerID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "class_id", "sport_id", "event_id"}).AddRow(winnerID, "Winner Team", 202, 1, 1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, class_id, sport_id, event_id FROM teams WHERE id = ?")).
			WithArgs(loserID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "class_id", "sport_id", "event_id"}).AddRow(loserID, "Loser Team", 201, 1, 1))

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO class_scores (event_id, class_id, gym1_win3_points) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE gym1_win3_points = gym1_win3_points + VALUES(gym1_win3_points)")).
			WithArgs(1, 202, 10).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT season FROM events WHERE id = ?")).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"season"}).AddRow("spring"))

		mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE class_scores cs
		JOIN (
			SELECT
				class_id,
				RANK() OVER (ORDER BY total_points_current_event DESC) AS new_rank
			FROM class_scores
			WHERE event_id = ?
		) ranked_data ON cs.class_id = ranked_data.class_id
		SET cs.rank_current_event = ranked_data.new_rank
		WHERE cs.event_id = ?
		`)).
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err = r.UpdateMatchResult(matchID, team1Score, team2Score, int(winnerID))
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
