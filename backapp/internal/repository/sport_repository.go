package repository

import (
	"backapp/internal/models"
	"database/sql"
	"errors"
	"fmt"
)

// SportRepository defines the interface for sport and event_sport related database operations.
type SportRepository interface {
	GetAllSports() ([]*models.Sport, error)
	GetSportByID(sportID int) (*models.Sport, error)
	CreateSport(sport *models.Sport) (int64, error)
	GetSportsByEventID(eventID int) ([]*models.EventSport, error)
	AssignSportToEvent(eventSport *models.EventSport) error
	DeleteSportFromEvent(eventID int, sportID int) error
	GetTeamsBySportID(sportID int) ([]*models.Team, error)
	GetSportDetails(eventID int, sportID int) (*models.EventSport, error)
	UpdateSportDetails(eventID int, sportID int, details models.EventSport) error
}

type sportRepository struct {
	db *sql.DB
}

// NewSportRepository creates a new instance of SportRepository.
func NewSportRepository(db *sql.DB) SportRepository {
	return &sportRepository{db: db}
}

// GetAllSports retrieves all sports from the database.
func (r *sportRepository) GetAllSports() ([]*models.Sport, error) {
	query := "SELECT id, name FROM sports ORDER BY id"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sports []*models.Sport
	for rows.Next() {
		sport := &models.Sport{}
		if err := rows.Scan(&sport.ID, &sport.Name); err != nil {
			return nil, err
		}
		sports = append(sports, sport)
	}
	return sports, nil
}

// CreateSport adds a new sport to the database.
func (r *sportRepository) CreateSport(sport *models.Sport) (int64, error) {
	query := "INSERT INTO sports (name) VALUES (?)"
	result, err := r.db.Exec(query, sport.Name)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetSportByID retrieves a sport by its ID.
func (r *sportRepository) GetSportByID(sportID int) (*models.Sport, error) {
	query := "SELECT id, name FROM sports WHERE id = ?"
	sport := &models.Sport{}
	err := r.db.QueryRow(query, sportID).Scan(&sport.ID, &sport.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("sport not found")
		}
		return nil, err
	}
	return sport, nil
}

// GetSportsByEventID retrieves all sports assigned to a specific event.
func (r *sportRepository) GetSportsByEventID(eventID int) ([]*models.EventSport, error) {
	query := "SELECT event_id, sport_id, description, rules, rules_type, rules_pdf_url, location, min_capacity, max_capacity FROM event_sports WHERE event_id = ?"
	rows, err := r.db.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eventSports []*models.EventSport
	for rows.Next() {
		eventSport := &models.EventSport{}
		if err := rows.Scan(&eventSport.EventID, &eventSport.SportID, &eventSport.Description, &eventSport.Rules, &eventSport.RulesType, &eventSport.RulesPdfURL, &eventSport.Location, &eventSport.MinCapacity, &eventSport.MaxCapacity); err != nil {
			return nil, err
		}
		eventSports = append(eventSports, eventSport)
	}
	return eventSports, nil
}

// AssignSportToEvent assigns a sport to an event in the database.
func (r *sportRepository) AssignSportToEvent(eventSport *models.EventSport) error {
	// Check if the sport is already assigned to the event
	var count int
	query := "SELECT COUNT(*) FROM event_sports WHERE event_id = ? AND sport_id = ?"
	err := r.db.QueryRow(query, eventSport.EventID, eventSport.SportID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("この競技はすでにこの大会に割り当てられています。")
	}

	// Prevent duplicate locations, except for 'other'
	if eventSport.Location != "other" {
		query := "SELECT COUNT(*) FROM event_sports WHERE event_id = ? AND location = ?"
		err := r.db.QueryRow(query, eventSport.EventID, eventSport.Location).Scan(&count)
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.New("この場所は、この大会で既に使用されています。")
		}
	}

	query = "INSERT INTO event_sports (event_id, sport_id, description, rules, rules_type, rules_pdf_url, location, min_capacity, max_capacity) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err = r.db.Exec(query, eventSport.EventID, eventSport.SportID, eventSport.Description, eventSport.Rules, eventSport.RulesType, eventSport.RulesPdfURL, eventSport.Location, eventSport.MinCapacity, eventSport.MaxCapacity)
	if err != nil {
		fmt.Printf("Error inserting EventSport: %v\n", err)
	}
	return err
}

// DeleteSportFromEvent removes the assignment of a sport from an event.
func (r *sportRepository) DeleteSportFromEvent(eventID int, sportID int) error {
	query := "DELETE FROM event_sports WHERE event_id = ? AND sport_id = ?"
	_, err := r.db.Exec(query, eventID, sportID)
	return err
}

// GetTeamsBySportID retrieves all teams for a given sport ID from the database.
func (r *sportRepository) GetTeamsBySportID(sportID int) ([]*models.Team, error) {
	query := `
		SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id, t.min_capacity, t.max_capacity 
		FROM teams t 
		JOIN classes c ON t.class_id = c.id 
		WHERE t.sport_id = ?
	`
	rows, err := r.db.Query(query, sportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		team := &models.Team{}
		var minCap, maxCap sql.NullInt64
		if err := rows.Scan(&team.ID, &team.Name, &team.ClassID, &team.SportID, &team.EventID, &minCap, &maxCap); err != nil {
			return nil, err
		}

		if minCap.Valid {
			val := int(minCap.Int64)
			team.MinCapacity = &val
		}
		if maxCap.Valid {
			val := int(maxCap.Int64)
			team.MaxCapacity = &val
		}

		teams = append(teams, team)
	}
	return teams, nil
}

func (r *sportRepository) GetSportDetails(eventID int, sportID int) (*models.EventSport, error) {
	query := "SELECT event_id, sport_id, description, rules, rules_type, rules_pdf_url, location, min_capacity, max_capacity FROM event_sports WHERE event_id = ? AND sport_id = ?"
	eventSport := &models.EventSport{}
	err := r.db.QueryRow(query, eventID, sportID).Scan(&eventSport.EventID, &eventSport.SportID, &eventSport.Description, &eventSport.Rules, &eventSport.RulesType, &eventSport.RulesPdfURL, &eventSport.Location, &eventSport.MinCapacity, &eventSport.MaxCapacity)
	if err != nil {
		if err == sql.ErrNoRows {
			return &models.EventSport{
				EventID:   eventID,
				SportID:   sportID,
				RulesType: "markdown",
			}, nil
		}
		return nil, err
	}
	return eventSport, nil
}

func (r *sportRepository) UpdateSportDetails(eventID int, sportID int, details models.EventSport) error {
	query := "UPDATE event_sports SET description = ?, rules = ?, rules_type = ?, rules_pdf_url = ?, min_capacity = ?, max_capacity = ? WHERE event_id = ? AND sport_id = ?"
	_, err := r.db.Exec(query, details.Description, details.Rules, details.RulesType, details.RulesPdfURL, details.MinCapacity, details.MaxCapacity, eventID, sportID)
	return err
}
