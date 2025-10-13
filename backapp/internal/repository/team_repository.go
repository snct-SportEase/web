package repository

import (
	"backapp/internal/models"
	"database/sql"
)

type TeamRepository interface {
	CreateTeam(team *models.Team) (int64, error)
	DeleteTeamsByEventAndSportID(eventID int, sportID int) error
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