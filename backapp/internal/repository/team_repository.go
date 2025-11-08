package repository

import (
	"backapp/internal/models"
	"database/sql"
)

type TeamRepository interface {
	CreateTeam(team *models.Team) (int64, error)
	DeleteTeamsByEventAndSportID(eventID int, sportID int) error
	GetTeamsByUserID(userID string) ([]*models.TeamWithSport, error)
}

type teamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) CreateTeam(team *models.Team) (int64, error) {
	query := "INSERT INTO teams (name, class_id, sport_id, event_id) VALUES (?, ?, ?, ?)"
	result, err := r.db.Exec(query, team.Name, team.ClassID, team.SportID, team.EventID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *teamRepository) DeleteTeamsByEventAndSportID(eventID int, sportID int) error {
	query := "DELETE FROM teams WHERE event_id = ? AND sport_id = ?"
	_, err := r.db.Exec(query, eventID, sportID)
	return err
}

func (r *teamRepository) GetTeamsByUserID(userID string) ([]*models.TeamWithSport, error) {
	query := `
		SELECT t.id, t.name, t.class_id, t.sport_id, t.event_id, s.name as sport_name
		FROM teams t
		INNER JOIN team_members tm ON t.id = tm.team_id
		INNER JOIN sports s ON t.sport_id = s.id
		WHERE tm.user_id = ?
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*models.TeamWithSport
	for rows.Next() {
		team := &models.TeamWithSport{}
		if err := rows.Scan(&team.ID, &team.Name, &team.ClassID, &team.SportID, &team.EventID, &team.SportName); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, nil
}
