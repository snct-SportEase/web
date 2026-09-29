package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrTournamentSeedUnsupported = errors.New("tournament seed editing is unsupported")
	ErrTournamentSeedLocked      = errors.New("tournament matches have started")
	ErrTournamentSeedChanged     = errors.New("tournament seed order has changed")
	ErrInvalidTournamentSeed     = errors.New("invalid tournament seed order")
)

type TournamentSeedTeam struct {
	TeamID   int    `json:"team_id"`
	TeamName string `json:"team_name"`
}

type seedMatch struct {
	id    int
	team1 sql.NullInt64
	team2 sql.NullInt64
}

// GetTournamentSeedOrder returns the occupied first-round slots in bracket order.
func (r *tournamentRepository) GetTournamentSeedOrder(eventID, tournamentID int) ([]TournamentSeedTeam, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	seeds, _, err := readTournamentSeeds(tx, eventID, tournamentID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return seeds, nil
}

// UpdateTournamentSeedOrder changes only first-round team IDs. Match IDs, times,
// links, and all other tournaments remain intact.
func (r *tournamentRepository) UpdateTournamentSeedOrder(eventID, tournamentID int, expectedIDs, teamIDs []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	seeds, matches, err := readTournamentSeeds(tx, eventID, tournamentID)
	if err != nil {
		return err
	}
	if len(expectedIDs) != len(seeds) {
		return ErrTournamentSeedChanged
	}
	if len(teamIDs) != len(seeds) {
		return ErrInvalidTournamentSeed
	}
	available := make(map[int]bool, len(seeds))
	for i, seed := range seeds {
		if expectedIDs[i] != seed.TeamID {
			return ErrTournamentSeedChanged
		}
		available[seed.TeamID] = true
	}
	for _, id := range teamIDs {
		if id <= 0 || !available[id] {
			return ErrInvalidTournamentSeed
		}
		delete(available, id)
	}

	next := 0
	for _, match := range matches {
		team1, team2 := match.team1, match.team2
		if team1.Valid {
			team1.Int64 = int64(teamIDs[next])
			next++
		}
		if team2.Valid {
			team2.Int64 = int64(teamIDs[next])
			next++
		}
		if _, err := tx.Exec("UPDATE matches SET team1_id = ?, team2_id = ? WHERE id = ? AND tournament_id = ?", nullableTeamID(team1), nullableTeamID(team2), match.id, tournamentID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func nullableTeamID(id sql.NullInt64) any {
	if id.Valid {
		return id.Int64
	}
	return nil
}

func readTournamentSeeds(tx *sql.Tx, eventID, tournamentID int) ([]TournamentSeedTeam, []seedMatch, error) {
	var templateKey sql.NullString
	var sportName, location string
	err := tx.QueryRow(`SELECT es.template_key, s.name, es.location
		FROM tournaments t
		JOIN event_sports es ON es.event_id = t.event_id AND es.sport_id = t.sport_id
		JOIN sports s ON s.id = t.sport_id
		WHERE t.event_id = ? AND t.id = ? FOR UPDATE`, eventID, tournamentID).Scan(&templateKey, &sportName, &location)
	if err != nil {
		return nil, nil, err
	}
	if templateKey.String == "board_game_tournament" || location == "noon_game" || strings.TrimSpace(sportName) == "将棋" || strings.TrimSpace(sportName) == "オセロ" {
		return nil, nil, ErrTournamentSeedUnsupported
	}

	rows, err := tx.Query(`SELECT m.id, m.round, m.team1_id, m.team2_id,
		m.team1_score, m.team2_score, m.winner_team_id, m.status,
		m.is_loser_bracket_match, team1.name, team2.name
		FROM matches m
		LEFT JOIN teams team1 ON team1.id = m.team1_id
		LEFT JOIN teams team2 ON team2.id = m.team2_id
		WHERE m.tournament_id = ?
		ORDER BY m.round, m.match_number_in_round, m.id FOR UPDATE`, tournamentID)
	if err != nil {
		return nil, nil, err
	}
	seeds := make([]TournamentSeedTeam, 0)
	matches := make([]seedMatch, 0)
	seen := map[int64]bool{}
	locked, unsupported := false, false
	for rows.Next() {
		var id, round int
		var team1, team2, score1, score2, winner sql.NullInt64
		var name1, name2, status sql.NullString
		var loser bool
		if err := rows.Scan(&id, &round, &team1, &team2, &score1, &score2, &winner, &status, &loser, &name1, &name2); err != nil {
			rows.Close()
			return nil, nil, err
		}
		if !status.Valid || status.String != "pending" || score1.Valid || score2.Valid || winner.Valid || (round > 0 && (team1.Valid || team2.Valid)) {
			locked = true
		}
		if loser {
			unsupported = true
		}
		if round != 0 {
			continue
		}
		matches = append(matches, seedMatch{id: id, team1: team1, team2: team2})
		for _, slot := range []struct {
			team sql.NullInt64
			name sql.NullString
		}{{team1, name1}, {team2, name2}} {
			if !slot.team.Valid {
				continue
			}
			if slot.team.Int64 <= 0 || !slot.name.Valid || seen[slot.team.Int64] {
				unsupported = true
				continue
			}
			seen[slot.team.Int64] = true
			seeds = append(seeds, TournamentSeedTeam{TeamID: int(slot.team.Int64), TeamName: slot.name.String})
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, nil, err
	}
	rows.Close()
	if unsupported || len(seeds) < 2 {
		return nil, nil, ErrTournamentSeedUnsupported
	}
	var scoreLogCount int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM score_logs WHERE source_match_id IN
		(SELECT id FROM matches WHERE tournament_id = ?)`, tournamentID).Scan(&scoreLogCount); err != nil {
		return nil, nil, fmt.Errorf("check tournament score logs: %w", err)
	}
	if locked || scoreLogCount > 0 {
		return nil, nil, ErrTournamentSeedLocked
	}
	return seeds, matches, nil
}
