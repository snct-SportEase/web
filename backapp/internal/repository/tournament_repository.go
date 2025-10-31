package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"backapp/internal/models"
)

type TournamentRepository interface {
	SaveTournament(eventID int, sportID int, sportName string, tournamentData *models.TournamentData, teams []*models.Team) error
	DeleteTournamentsByEventID(eventID int) error
	DeleteTournamentsByEventAndSportID(eventID int, sportID int) error
	GetTournamentsByEventID(eventID int) ([]*models.Tournament, error)
	UpdateMatchStartTime(matchID int, startTime string) error
	UpdateMatchStatus(matchID int, status string) error
}

type tournamentRepository struct {
	db *sql.DB
}

func NewTournamentRepository(db *sql.DB) TournamentRepository {
	return &tournamentRepository{db: db}
}

func (r *tournamentRepository) GetTournamentsByEventID(eventID int) ([]*models.Tournament, error) {
	rows, err := r.db.Query("SELECT id, name, sport_id FROM tournaments WHERE event_id = ?", eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tournaments := make([]*models.Tournament, 0)
	for rows.Next() {
		var t models.Tournament
		if err := rows.Scan(&t.ID, &t.Name, &t.SportID); err != nil {
			return nil, err
		}

		matches, err := r.getMatchesByTournamentID(int64(t.ID))
		if err != nil {
			return nil, err
		}

		var bracketryMatches []models.Match
		contestants := make(map[string]models.Contestant)
		teamMap := make(map[int]*models.Team)
		contestantCounter := 0

		for _, m := range matches {
			var sides []models.Side

			if m.Team1ID.Valid {
				side, team, err := r.getSide(m.Team1ID.Int64, &contestantCounter, teamMap, contestants)
				if err != nil {
					return nil, err
				}
				sides = append(sides, side)
				if team != nil {
					teamMap[int(m.Team1ID.Int64)] = team
				}
			}

			if m.Team2ID.Valid {
				side, team, err := r.getSide(m.Team2ID.Int64, &contestantCounter, teamMap, contestants)
				if err != nil {
					return nil, err
				}
				sides = append(sides, side)
				if team != nil {
					teamMap[int(m.Team2ID.Int64)] = team
				}
			}

			bracketryMatches = append(bracketryMatches, models.Match{
				ID:          m.ID,
				RoundIndex:  m.Round,
				Order:       m.MatchNumberInRound,
				Sides:       sides,
				MatchStatus: m.StartTime.String,
				StartTime: func() string {
					if m.StartTime.Valid {
						return m.StartTime.String
					}
					return ""
				}(),
			})
		}

		numRounds := 0
		for _, m := range bracketryMatches {
			if m.RoundIndex+1 > numRounds {
				numRounds = m.RoundIndex + 1
			}
		}
		rounds := make([]models.Round, numRounds)
		for i := 0; i < numRounds; i++ {
			rounds[i] = models.Round{Name: fmt.Sprintf("Round %d", i+1)}
		}

		t.Data, err = r.marshal(models.TournamentData{
			Rounds:      rounds,
			Matches:     bracketryMatches,
			Contestants: contestants,
		})
		if err != nil {
			return nil, err
		}
		tournaments = append(tournaments, &t)
	}

	return tournaments, nil
}

func (r *tournamentRepository) getSide(teamID int64, contestantCounter *int, teamMap map[int]*models.Team, contestants map[string]models.Contestant) (models.Side, *models.Team, error) {
	var team *models.Team
	var ok bool

	if team, ok = teamMap[int(teamID)]; !ok {
		var err error
		team, err = r.getTeamByID(teamID)
		if err != nil {
			if err == sql.ErrNoRows {
				return models.Side{}, nil, nil // Team not found is not a fatal error here
			}
			return models.Side{}, nil, err
		}
	}

	if team == nil {
		return models.Side{}, nil, nil
	}

	contestantID := ""
	for id, c := range contestants {
		if len(c.Players) > 0 && c.Players[0].Title == team.Name {
			contestantID = id
			break
		}
	}

	if contestantID == "" {
		contestantID = "c" + strconv.Itoa(*contestantCounter)
		contestants[contestantID] = models.Contestant{
			Players: []models.Player{{Title: team.Name}},
		}
		(*contestantCounter)++
	}

	return models.Side{ContestantID: contestantID}, team, nil
}

func (r *tournamentRepository) getMatchesByTournamentID(tournamentID int64) ([]*models.MatchDB, error) {
	rows, err := r.db.Query("SELECT id, round, match_number_in_round, team1_id, team2_id, winner_team_id, status, next_match_id, start_time FROM matches WHERE tournament_id = ? ORDER BY round, match_number_in_round", tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []*models.MatchDB
	for rows.Next() {
		var m models.MatchDB
		if err := rows.Scan(&m.ID, &m.Round, &m.MatchNumberInRound, &m.Team1ID, &m.Team2ID, &m.WinnerID, &m.Status, &m.NextMatchID, &m.StartTime); err != nil {
			return nil, err
		}
		matches = append(matches, &m)
	}
	return matches, nil
}

func (r *tournamentRepository) getTeamByID(teamID int64) (*models.Team, error) {
	var team models.Team
	row := r.db.QueryRow("SELECT id, name, class_id FROM teams WHERE id = ?", teamID)
	if err := row.Scan(&team.ID, &team.Name, &team.ClassID); err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *tournamentRepository) marshal(data interface{}) (json.RawMessage, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (r *tournamentRepository) SaveTournament(eventID int, sportID int, sportName string, tournamentData *models.TournamentData, teams []*models.Team) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	tournamentName := fmt.Sprintf("%s Tournament", sportName)
	res, err := tx.Exec("INSERT INTO tournaments (name, event_id, sport_id) VALUES (?, ?, ?)", tournamentName, eventID, sportID)
	if err != nil {
		tx.Rollback()
		return err
	}
	tournamentID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	roundMatchIDs := make([][]int64, len(tournamentData.Rounds))
	matchMetas := map[int64]models.Match{}

	for _, match := range tournamentData.Matches {
		var team1ID, team2ID sql.NullInt64

		if len(match.Sides) > 0 && match.Sides[0].ContestantID != "" {
			contestantIndex, _ := strconv.Atoi(strings.TrimPrefix(match.Sides[0].ContestantID, "c"))
			if contestantIndex < len(teams) {
				team1ID = sql.NullInt64{Int64: int64(teams[contestantIndex].ID), Valid: true}
			}
		}
		if len(match.Sides) > 1 && match.Sides[1].ContestantID != "" {
			contestantIndex, _ := strconv.Atoi(strings.TrimPrefix(match.Sides[1].ContestantID, "c"))
			if contestantIndex < len(teams) {
				team2ID = sql.NullInt64{Int64: int64(teams[contestantIndex].ID), Valid: true}
			}
		}

		res, err := tx.Exec(
			"INSERT INTO matches (tournament_id, round, match_number_in_round, team1_id, team2_id, status) VALUES (?, ?, ?, ?, ?, ?)",
			tournamentID,
			match.RoundIndex,
			match.Order,
			team1ID,
			team2ID,
			"pending",
		)
		if err != nil {
			tx.Rollback()
			return err
		}
		matchID, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return err
		}
		if match.RoundIndex < len(roundMatchIDs) {
			roundMatchIDs[match.RoundIndex] = append(roundMatchIDs[match.RoundIndex], matchID)
			matchMetas[matchID] = match
		}
	}

	for i := 0; i < len(roundMatchIDs)-1; i++ {
		for j, matchID := range roundMatchIDs[i] {
			if match, ok := matchMetas[matchID]; ok && match.IsBronzeMatch {
				continue
			}
			nextMatchID := roundMatchIDs[i+1][j/2]
			_, err := tx.Exec("UPDATE matches SET next_match_id = ? WHERE id = ?", nextMatchID, matchID)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *tournamentRepository) DeleteTournamentsByEventID(eventID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	rows, err := tx.Query("SELECT id FROM tournaments WHERE event_id = ?", eventID)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()

	var tournamentIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			tx.Rollback()
			return err
		}
		tournamentIDs = append(tournamentIDs, id)
	}

	if len(tournamentIDs) > 0 {
		args := make([]interface{}, len(tournamentIDs))
		for i, v := range tournamentIDs {
			args[i] = v
		}

		qMarks := strings.Repeat("?,", len(args)-1) + "?"

		query := fmt.Sprintf("DELETE FROM matches WHERE tournament_id IN (%s)", qMarks)
		_, err = tx.Exec(query, args...)
		if err != nil {
			tx.Rollback()
			return err
		}

		query = fmt.Sprintf("DELETE FROM tournaments WHERE id IN (%s)", qMarks)
		_, err = tx.Exec(query, args...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *tournamentRepository) DeleteTournamentsByEventAndSportID(eventID int, sportID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	rows, err := tx.Query("SELECT id FROM tournaments WHERE event_id = ? AND sport_id = ?", eventID, sportID)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()

	var tournamentIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			tx.Rollback()
			return err
		}
		tournamentIDs = append(tournamentIDs, id)
	}

	if len(tournamentIDs) > 0 {
		args := make([]interface{}, len(tournamentIDs))
		for i, v := range tournamentIDs {
			args[i] = v
		}

		qMarks := strings.Repeat("?,", len(args)-1) + "?"

		query := fmt.Sprintf("DELETE FROM matches WHERE tournament_id IN (%s)", qMarks)
		_, err = tx.Exec(query, args...)
		if err != nil {
			tx.Rollback()
			return err
		}

		query = fmt.Sprintf("DELETE FROM tournaments WHERE id IN (%s)", qMarks)
		_, err = tx.Exec(query, args...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *tournamentRepository) UpdateMatchStartTime(matchID int, startTime string) error {
	if startTime != "" {
		// Only change status to 'scheduled' if it is currently 'pending'.
		_, err := r.db.Exec("UPDATE matches SET start_time = ?, status = CASE WHEN status = 'pending' THEN 'scheduled' ELSE status END WHERE id = ?", startTime, matchID)
		return err
	}
	// If startTime is empty, just update the time.
	_, err := r.db.Exec("UPDATE matches SET start_time = ? WHERE id = ?", startTime, matchID)
	return err
}

func (r *tournamentRepository) UpdateMatchStatus(matchID int, status string) error {
	_, err := r.db.Exec("UPDATE matches SET status = ? WHERE id = ?", status, matchID)
	return err
}
