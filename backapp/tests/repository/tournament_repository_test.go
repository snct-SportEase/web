package repository_test

import (
	"regexp"
	"testing"

	"backapp/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestTournamentRepository_GetTournamentsByEventID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := repository.NewTournamentRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT is_rainy_mode FROM events WHERE id = ?")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"is_rainy_mode"}).AddRow(false))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, sport_id FROM tournaments WHERE event_id = ?")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "sport_id"}).
			AddRow(10, "Basketball Tournament", 1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT
			m.id,
			m.tournament_id,
			m.round,
			m.match_number_in_round,
			m.team1_id,
			m.team2_id,
			m.team1_score,
			m.team2_score,
			CASE
				WHEN m.team1_score > m.team2_score THEN m.team1_id
				WHEN m.team2_score > m.team1_score THEN m.team2_id
				ELSE NULL
			END AS winner_team_id,
			m.status,
			m.next_match_id,
			m.match_start_time,
			m.is_bronze_match,
			m.is_loser_bracket_match,
			m.loser_bracket_round,
			m.loser_bracket_block,
			m.rainy_mode_start_time
		FROM matches m
		JOIN tournaments t ON m.tournament_id = t.id
		WHERE t.event_id = ?
		ORDER BY m.tournament_id, m.round, m.match_number_in_round
	`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"tournament_id",
			"round",
			"match_number_in_round",
			"team1_id",
			"team2_id",
			"team1_score",
			"team2_score",
			"winner_team_id",
			"status",
			"next_match_id",
			"match_start_time",
			"is_bronze_match",
			"is_loser_bracket_match",
			"loser_bracket_round",
			"loser_bracket_block",
			"rainy_mode_start_time",
		}).AddRow(100, 10, 0, 0, 1, 2, nil, nil, nil, "pending", nil, nil, false, false, nil, nil, nil))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id
		FROM teams t
		JOIN classes c ON t.class_id = c.id
		WHERE c.event_id = ?
	`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "class_id", "sport_id", "event_id"}).
			AddRow(1, "Team A", 101, 1, 1).
			AddRow(2, "Team B", 102, 1, 1))

	tournaments, err := r.GetTournamentsByEventID(1)
	assert.NoError(t, err)
	assert.Len(t, tournaments, 1)
	assert.Contains(t, string(tournaments[0].Data), "Team A")
	assert.Contains(t, string(tournaments[0].Data), "Team B")
	assert.NoError(t, mock.ExpectationsWereMet())
}

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
		rows := sqlmock.NewRows([]string{"id", "tournament_id", "round", "match_number_in_round", "team1_id", "team2_id", "winner_team_id", "status", "next_match_id", "start_time", "is_bronze_match", "is_loser_bracket_match", "loser_bracket_round", "loser_bracket_block", "rainy_mode_start_time"}).
			AddRow(matchID, tournamentID, 1, 1, team1ID, team2ID, nil, "inprogress", nextMatchID, "", false, false, nil, nil, nil)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tournament_id, round, match_number_in_round, team1_id, team2_id, CASE WHEN team1_score > team2_score THEN team1_id WHEN team2_score > team1_score THEN team2_id ELSE NULL END AS winner_team_id, status, next_match_id, match_start_time, is_bronze_match, is_loser_bracket_match, loser_bracket_round, loser_bracket_block, rainy_mode_start_time FROM matches WHERE id = ?")).
			WithArgs(matchID).WillReturnRows(rows)

		// Mock for rainy mode check (happens right after getMatchByID)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT t.event_id, t.sport_id, es.location FROM tournaments t LEFT JOIN event_sports es ON es.event_id = t.event_id AND es.sport_id = t.sport_id WHERE t.id = ?")).
			WithArgs(tournamentID).
			WillReturnRows(sqlmock.NewRows([]string{"event_id", "sport_id", "location"}).AddRow(1, 1, "gym1"))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT is_rainy_mode FROM events WHERE id = ?")).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"is_rainy_mode"}).AddRow(false))

		// Mock update current match
		mock.ExpectExec(regexp.QuoteMeta("UPDATE matches SET team1_score = ?, team2_score = ?, status = 'finished' WHERE id = ?")).
			WithArgs(team1Score, team2Score, matchID).WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock getMatchByID for the next match
		nextMatchRows := sqlmock.NewRows([]string{"id", "tournament_id", "round", "match_number_in_round", "team1_id", "team2_id", "winner_team_id", "status", "next_match_id", "start_time", "is_bronze_match", "is_loser_bracket_match", "loser_bracket_round", "loser_bracket_block", "rainy_mode_start_time"}).
			AddRow(nextMatchID, tournamentID, 2, 1, nil, nil, nil, "pending", nil, "", false, false, nil, nil, nil)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tournament_id, round, match_number_in_round, team1_id, team2_id, CASE WHEN team1_score > team2_score THEN team1_id WHEN team2_score > team1_score THEN team2_id ELSE NULL END AS winner_team_id, status, next_match_id, match_start_time, is_bronze_match, is_loser_bracket_match, loser_bracket_round, loser_bracket_block, rainy_mode_start_time FROM matches WHERE id = ?")).
			WithArgs(nextMatchID).WillReturnRows(nextMatchRows)

		// Mock update next match
		mock.ExpectExec(regexp.QuoteMeta("UPDATE matches SET team1_id = ? WHERE id = ?")).
			WithArgs(winnerID, nextMatchID).WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock for bronze match logic (not a semi-final)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT MAX(round) FROM matches WHERE tournament_id = ?")).WithArgs(tournamentID).WillReturnRows(sqlmock.NewRows([]string{"MAX(round)"}).AddRow(3))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT t.event_id, t.sport_id, es.location FROM tournaments t LEFT JOIN event_sports es ON es.event_id = t.event_id AND es.sport_id = t.sport_id WHERE t.id = ?")).
			WithArgs(tournamentID).
			WillReturnRows(sqlmock.NewRows([]string{"event_id", "sport_id", "location"}).AddRow(1, 1, "gym1"))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id FROM teams t JOIN classes c ON t.class_id = c.id WHERE t.id = ?")).
			WithArgs(winnerID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "class_id", "sport_id", "event_id"}).AddRow(winnerID, "Winner Team", 101, 1, 1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id FROM teams t JOIN classes c ON t.class_id = c.id WHERE t.id = ?")).
			WithArgs(team2ID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "class_id", "sport_id", "event_id"}).AddRow(team2ID, "Loser Team", 102, 1, 1))

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO score_logs (event_id, class_id, points, reason) VALUES (?, ?, ?, ?)")).
			WithArgs(1, 101, 10, "gym1_win2_points").
			WillReturnResult(sqlmock.NewResult(1, 1))

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
		rows := sqlmock.NewRows([]string{"id", "tournament_id", "round", "match_number_in_round", "team1_id", "team2_id", "winner_team_id", "status", "next_match_id", "start_time", "is_bronze_match", "is_loser_bracket_match", "loser_bracket_round", "loser_bracket_block", "rainy_mode_start_time"}).
			AddRow(matchID, tournamentID, 2, 1, team1ID, team2ID, nil, "inprogress", nextMatchID, "", false, false, nil, nil, nil)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tournament_id, round, match_number_in_round, team1_id, team2_id, CASE WHEN team1_score > team2_score THEN team1_id WHEN team2_score > team1_score THEN team2_id ELSE NULL END AS winner_team_id, status, next_match_id, match_start_time, is_bronze_match, is_loser_bracket_match, loser_bracket_round, loser_bracket_block, rainy_mode_start_time FROM matches WHERE id = ?")).
			WithArgs(matchID).WillReturnRows(rows)

		// Mock for rainy mode check (happens right after getMatchByID)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT t.event_id, t.sport_id, es.location FROM tournaments t LEFT JOIN event_sports es ON es.event_id = t.event_id AND es.sport_id = t.sport_id WHERE t.id = ?")).
			WithArgs(tournamentID).
			WillReturnRows(sqlmock.NewRows([]string{"event_id", "sport_id", "location"}).AddRow(1, 1, "gym1"))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT is_rainy_mode FROM events WHERE id = ?")).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"is_rainy_mode"}).AddRow(false))

		// Mock update current match
		mock.ExpectExec(regexp.QuoteMeta("UPDATE matches SET team1_score = ?, team2_score = ?, status = 'finished' WHERE id = ?")).
			WithArgs(team1Score, team2Score, matchID).WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock getMatchByID for the next match (final)
		nextMatchRows := sqlmock.NewRows([]string{"id", "tournament_id", "round", "match_number_in_round", "team1_id", "team2_id", "winner_team_id", "status", "next_match_id", "start_time", "is_bronze_match", "is_loser_bracket_match", "loser_bracket_round", "loser_bracket_block", "rainy_mode_start_time"}).
			AddRow(nextMatchID, tournamentID, 3, 1, nil, nil, nil, "pending", nil, "", false, false, nil, nil, nil)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tournament_id, round, match_number_in_round, team1_id, team2_id, CASE WHEN team1_score > team2_score THEN team1_id WHEN team2_score > team1_score THEN team2_id ELSE NULL END AS winner_team_id, status, next_match_id, match_start_time, is_bronze_match, is_loser_bracket_match, loser_bracket_round, loser_bracket_block, rainy_mode_start_time FROM matches WHERE id = ?")).
			WithArgs(nextMatchID).WillReturnRows(nextMatchRows)

		// Mock update next match
		mock.ExpectExec(regexp.QuoteMeta("UPDATE matches SET team1_id = ? WHERE id = ?")).
			WithArgs(winnerID, nextMatchID).WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock for bronze match logic (is a semi-final)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT MAX(round) FROM matches WHERE tournament_id = ?")).WithArgs(tournamentID).WillReturnRows(sqlmock.NewRows([]string{"MAX(round)"}).AddRow(3))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM matches WHERE tournament_id = ? AND is_bronze_match = TRUE")).WithArgs(tournamentID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(bronzeMatchID))

		// Mock getMatchByID for the bronze match
		bronzeMatchRows := sqlmock.NewRows([]string{"id", "tournament_id", "round", "match_number_in_round", "team1_id", "team2_id", "winner_team_id", "status", "next_match_id", "start_time", "is_bronze_match", "is_loser_bracket_match", "loser_bracket_round", "loser_bracket_block", "rainy_mode_start_time"}).
			AddRow(bronzeMatchID, tournamentID, 3, 2, nil, nil, nil, "pending", nil, "", true, false, nil, nil, nil)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tournament_id, round, match_number_in_round, team1_id, team2_id, CASE WHEN team1_score > team2_score THEN team1_id WHEN team2_score > team1_score THEN team2_id ELSE NULL END AS winner_team_id, status, next_match_id, match_start_time, is_bronze_match, is_loser_bracket_match, loser_bracket_round, loser_bracket_block, rainy_mode_start_time FROM matches WHERE id = ?")).
			WithArgs(bronzeMatchID).WillReturnRows(bronzeMatchRows)

		// Mock update bronze match
		mock.ExpectExec(regexp.QuoteMeta("UPDATE matches SET team1_id = ? WHERE id = ?")).
			WithArgs(loserID, bronzeMatchID).WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT t.event_id, t.sport_id, es.location FROM tournaments t LEFT JOIN event_sports es ON es.event_id = t.event_id AND es.sport_id = t.sport_id WHERE t.id = ?")).
			WithArgs(tournamentID).
			WillReturnRows(sqlmock.NewRows([]string{"event_id", "sport_id", "location"}).AddRow(1, 1, "gym1"))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id FROM teams t JOIN classes c ON t.class_id = c.id WHERE t.id = ?")).
			WithArgs(winnerID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "class_id", "sport_id", "event_id"}).AddRow(winnerID, "Winner Team", 202, 1, 1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id FROM teams t JOIN classes c ON t.class_id = c.id WHERE t.id = ?")).
			WithArgs(loserID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "class_id", "sport_id", "event_id"}).AddRow(loserID, "Loser Team", 201, 1, 1))

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO score_logs (event_id, class_id, points, reason) VALUES (?, ?, ?, ?)")).
			WithArgs(1, 202, 10, "gym1_win3_points").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err = r.UpdateMatchResult(matchID, team1Score, team2Score, int(winnerID))
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - Loser bracket round 2 in gym2 awards 10 points to gym2_loser_bracket_champion_points", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		r := repository.NewTournamentRepository(db)

		matchID := 10
		team1ID := int64(10)
		team2ID := int64(11)
		tournamentID := 2
		eventID := 1
		team1Score := 2
		team2Score := 1
		winnerID := team1ID
		loserID := team2ID
		loserBracketRound := int64(2)
		loserBracketBlock := "A"
		classID := 301

		mock.ExpectBegin()

		// Mock getMatchByID for the loser bracket round 2 match
		rows := sqlmock.NewRows([]string{"id", "tournament_id", "round", "match_number_in_round", "team1_id", "team2_id", "winner_team_id", "status", "next_match_id", "start_time", "is_bronze_match", "is_loser_bracket_match", "loser_bracket_round", "loser_bracket_block", "rainy_mode_start_time"}).
			AddRow(matchID, tournamentID, 1, 0, team1ID, team2ID, nil, "inprogress", nil, "", false, true, loserBracketRound, loserBracketBlock, nil)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, tournament_id, round, match_number_in_round, team1_id, team2_id, CASE WHEN team1_score > team2_score THEN team1_id WHEN team2_score > team1_score THEN team2_id ELSE NULL END AS winner_team_id, status, next_match_id, match_start_time, is_bronze_match, is_loser_bracket_match, loser_bracket_round, loser_bracket_block, rainy_mode_start_time FROM matches WHERE id = ?")).
			WithArgs(matchID).WillReturnRows(rows)

		// Mock for rainy mode check
		mock.ExpectQuery(regexp.QuoteMeta("SELECT t.event_id, t.sport_id, es.location FROM tournaments t LEFT JOIN event_sports es ON es.event_id = t.event_id AND es.sport_id = t.sport_id WHERE t.id = ?")).
			WithArgs(tournamentID).
			WillReturnRows(sqlmock.NewRows([]string{"event_id", "sport_id", "location"}).AddRow(eventID, 2, "gym2"))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT is_rainy_mode FROM events WHERE id = ?")).
			WithArgs(eventID).
			WillReturnRows(sqlmock.NewRows([]string{"is_rainy_mode"}).AddRow(false))

		// Mock update current match
		mock.ExpectExec(regexp.QuoteMeta("UPDATE matches SET team1_score = ?, team2_score = ?, status = 'finished' WHERE id = ?")).
			WithArgs(team1Score, team2Score, matchID).WillReturnResult(sqlmock.NewResult(1, 1))

		// NextMatchID is nil for loser bracket round 2, so no next match update

		// Mock MAX(round) query (needed before applyScoring)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT MAX(round) FROM matches WHERE tournament_id = ?")).
			WithArgs(tournamentID).
			WillReturnRows(sqlmock.NewRows([]string{"MAX(round)"}).AddRow(1))

		// Mock getTournamentMetadata for applyScoring
		mock.ExpectQuery(regexp.QuoteMeta("SELECT t.event_id, t.sport_id, es.location FROM tournaments t LEFT JOIN event_sports es ON es.event_id = t.event_id AND es.sport_id = t.sport_id WHERE t.id = ?")).
			WithArgs(tournamentID).
			WillReturnRows(sqlmock.NewRows([]string{"event_id", "sport_id", "location"}).AddRow(eventID, 2, "gym2"))

		// Mock getTeamByID for winner team
		mock.ExpectQuery(regexp.QuoteMeta("SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id FROM teams t JOIN classes c ON t.class_id = c.id WHERE t.id = ?")).
			WithArgs(winnerID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "class_id", "sport_id", "event_id"}).AddRow(winnerID, "Winner Team", classID, 2, eventID))

		// Mock getTeamByID for loser team (needed for validation)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id FROM teams t JOIN classes c ON t.class_id = c.id WHERE t.id = ?")).
			WithArgs(loserID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "class_id", "sport_id", "event_id"}).AddRow(loserID, "Loser Team", 302, 2, eventID))

		// Mock adding points to gym2_loser_bracket_champion_points
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO score_logs (event_id, class_id, points, reason) VALUES (?, ?, ?, ?)")).
			WithArgs(eventID, classID, 10, "gym2_loser_bracket_champion_points").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err = r.UpdateMatchResult(matchID, team1Score, team2Score, int(winnerID))
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

}
