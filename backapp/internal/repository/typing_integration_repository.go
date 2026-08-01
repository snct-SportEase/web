package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"backapp/internal/models"
)

var ErrTypingCompetitionNotFound = errors.New("typing competition not found")
var ErrActiveEventNotFound = errors.New("active event not found")
var ErrTypingRosterMismatch = errors.New("typing roster does not match confirmed team members")

type TypingIntegrationRepository interface {
	GetActiveEvent() (*models.TypingEvent, []*models.TypingSport, error)
	GetCompetitionSnapshot(eventID, sportID int) (*models.TypingCompetitionSnapshot, error)
	SetTeamEntryOrder(teamID int, playerIDs []string) error
}

func (r *typingIntegrationRepository) SetTeamEntryOrder(teamID int, playerIDs []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`
		SELECT user_id, is_confirmed
		FROM team_members
		WHERE team_id = ?
		ORDER BY entry_order
		FOR UPDATE
	`, teamID)
	if err != nil {
		return err
	}

	confirmed := make(map[string]struct{})
	unconfirmed := make([]string, 0)
	for rows.Next() {
		var playerID string
		var isConfirmed bool
		if err := rows.Scan(&playerID, &isConfirmed); err != nil {
			rows.Close()
			return err
		}
		if isConfirmed {
			confirmed[playerID] = struct{}{}
		} else {
			unconfirmed = append(unconfirmed, playerID)
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if len(confirmed) != len(playerIDs) {
		return ErrTypingRosterMismatch
	}
	seen := make(map[string]struct{}, len(playerIDs))
	for _, playerID := range playerIDs {
		if _, exists := confirmed[playerID]; !exists {
			return ErrTypingRosterMismatch
		}
		if _, duplicate := seen[playerID]; duplicate {
			return ErrTypingRosterMismatch
		}
		seen[playerID] = struct{}{}
	}

	// Move every position out of the target range before swapping values. The
	// unique (team_id, entry_order) constraint would otherwise reject a swap.
	if _, err := tx.Exec(`
		UPDATE team_members
		SET entry_order = entry_order + 10000
		WHERE team_id = ?
	`, teamID); err != nil {
		return err
	}

	orderedPlayerIDs := append(append(make([]string, 0, len(playerIDs)+len(unconfirmed)), playerIDs...), unconfirmed...)
	for index, playerID := range orderedPlayerIDs {
		result, err := tx.Exec(`
			UPDATE team_members
			SET entry_order = ?
			WHERE team_id = ? AND user_id = ?
		`, index+1, teamID, playerID)
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected != 1 {
			return ErrTypingRosterMismatch
		}
	}

	return tx.Commit()
}

type typingIntegrationRepository struct {
	db *sql.DB
}

func NewTypingIntegrationRepository(db *sql.DB) TypingIntegrationRepository {
	return &typingIntegrationRepository{db: db}
}

func (r *typingIntegrationRepository) GetActiveEvent() (*models.TypingEvent, []*models.TypingSport, error) {
	event := &models.TypingEvent{}
	err := r.db.QueryRow(`
		SELECT e.id, e.name, e.year, e.season, e.status, e.start_date, e.end_date
		FROM active_event ae
		JOIN events e ON e.id = ae.event_id
		ORDER BY ae.id
		LIMIT 1
	`).Scan(
		&event.ID,
		&event.Name,
		&event.Year,
		&event.Season,
		&event.Status,
		&event.StartDate,
		&event.EndDate,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrActiveEventNotFound
	}
	if err != nil {
		return nil, nil, err
	}

	rows, err := r.db.Query(`
		SELECT s.id, s.name
		FROM event_sports es
		JOIN sports s ON s.id = es.sport_id
		WHERE es.event_id = ?
		ORDER BY s.id
	`, event.ID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	sports := make([]*models.TypingSport, 0)
	for rows.Next() {
		sport := &models.TypingSport{}
		if err := rows.Scan(&sport.ID, &sport.Name); err != nil {
			return nil, nil, err
		}
		sports = append(sports, sport)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return event, sports, nil
}

func (r *typingIntegrationRepository) GetCompetitionSnapshot(eventID, sportID int) (*models.TypingCompetitionSnapshot, error) {
	snapshot := &models.TypingCompetitionSnapshot{
		Event:       &models.TypingEvent{},
		Sport:       &models.TypingSport{},
		Tournaments: make([]*models.TypingTournament, 0),
		Teams:       make([]*models.TypingTeam, 0),
		Warnings:    make([]string, 0),
	}

	err := r.db.QueryRow(`
		SELECT
			e.id, e.name, e.year, e.season, e.status, e.start_date, e.end_date,
			s.id, s.name
		FROM event_sports es
		JOIN events e ON e.id = es.event_id
		JOIN sports s ON s.id = es.sport_id
		WHERE es.event_id = ? AND es.sport_id = ?
	`, eventID, sportID).Scan(
		&snapshot.Event.ID,
		&snapshot.Event.Name,
		&snapshot.Event.Year,
		&snapshot.Event.Season,
		&snapshot.Event.Status,
		&snapshot.Event.StartDate,
		&snapshot.Event.EndDate,
		&snapshot.Sport.ID,
		&snapshot.Sport.Name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTypingCompetitionNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := r.loadTypingTeams(snapshot, eventID, sportID); err != nil {
		return nil, err
	}
	if err := r.loadTypingTournaments(snapshot, eventID, sportID); err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (r *typingIntegrationRepository) loadTypingTeams(snapshot *models.TypingCompetitionSnapshot, eventID, sportID int) error {
	rows, err := r.db.Query(`
		SELECT
			t.id, t.name, c.id, c.name,
			u.id, u.display_name, tm.entry_order
		FROM teams t
		JOIN classes c ON c.id = t.class_id
		LEFT JOIN team_members tm
			ON tm.team_id = t.id AND tm.is_confirmed = TRUE
		LEFT JOIN users u ON u.id = tm.user_id
		WHERE c.event_id = ? AND t.sport_id = ?
		ORDER BY c.name, t.id, tm.entry_order
	`, eventID, sportID)
	if err != nil {
		return err
	}
	defer rows.Close()

	teamByID := make(map[int]*models.TypingTeam)
	for rows.Next() {
		var teamID, classID int
		var teamName, className string
		var playerID, playerName sql.NullString
		var entryOrder sql.NullInt64
		if err := rows.Scan(
			&teamID,
			&teamName,
			&classID,
			&className,
			&playerID,
			&playerName,
			&entryOrder,
		); err != nil {
			return err
		}

		team, exists := teamByID[teamID]
		if !exists {
			team = &models.TypingTeam{
				ID:        teamID,
				Name:      teamName,
				ClassID:   classID,
				ClassName: className,
				Players:   make([]*models.TypingPlayer, 0),
			}
			teamByID[teamID] = team
			snapshot.Teams = append(snapshot.Teams, team)
		}

		if !playerID.Valid {
			continue
		}

		player := &models.TypingPlayer{
			ID:         playerID.String,
			EntryOrder: int(entryOrder.Int64),
		}
		if playerName.Valid {
			name := playerName.String
			player.Name = &name
		}
		team.Players = append(team.Players, player)
	}

	return rows.Err()
}

func (r *typingIntegrationRepository) loadTypingTournaments(snapshot *models.TypingCompetitionSnapshot, eventID, sportID int) error {
	rows, err := r.db.Query(`
		SELECT
			t.id, t.name,
			m.id, m.round, m.match_number_in_round, m.status,
			m.match_start_time, m.match_end_time, m.team1_id, m.team2_id
		FROM tournaments t
		LEFT JOIN matches m ON m.tournament_id = t.id
		WHERE t.event_id = ? AND t.sport_id = ?
		ORDER BY t.id, m.round, m.match_number_in_round
	`, eventID, sportID)
	if err != nil {
		return err
	}
	defer rows.Close()

	tournamentByID := make(map[int]*models.TypingTournament)
	for rows.Next() {
		var tournamentID int
		var tournamentName string
		var matchID, round, matchNumber sql.NullInt64
		var status sql.NullString
		var startTime, endTime sql.NullTime
		var team1ID, team2ID sql.NullInt64
		if err := rows.Scan(
			&tournamentID,
			&tournamentName,
			&matchID,
			&round,
			&matchNumber,
			&status,
			&startTime,
			&endTime,
			&team1ID,
			&team2ID,
		); err != nil {
			return err
		}

		tournament, exists := tournamentByID[tournamentID]
		if !exists {
			tournament = &models.TypingTournament{
				ID:      tournamentID,
				Name:    tournamentName,
				Matches: make([]*models.TypingMatch, 0),
			}
			tournamentByID[tournamentID] = tournament
			snapshot.Tournaments = append(snapshot.Tournaments, tournament)
		}

		if !matchID.Valid {
			continue
		}

		roundNumber := int(round.Int64) + 1
		matchNumberValue := int(matchNumber.Int64) + 1
		match := &models.TypingMatch{
			ID:          int(matchID.Int64),
			Name:        fmt.Sprintf("第%dラウンド 第%d試合", roundNumber, matchNumberValue),
			RoundNumber: roundNumber,
			MatchNumber: matchNumberValue,
			Status:      status.String,
			TeamIDs:     make([]int, 0, 2),
		}
		if startTime.Valid {
			value := startTime.Time
			match.StartTime = &value
		}
		if endTime.Valid {
			value := endTime.Time
			match.EndTime = &value
		}
		if team1ID.Valid {
			match.TeamIDs = append(match.TeamIDs, int(team1ID.Int64))
		}
		if team2ID.Valid {
			match.TeamIDs = append(match.TeamIDs, int(team2ID.Int64))
		}
		tournament.Matches = append(tournament.Matches, match)
	}

	return rows.Err()
}
