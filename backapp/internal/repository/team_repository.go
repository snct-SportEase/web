package repository

import (
	"backapp/internal/models"
	"database/sql"
)

type TeamRepository interface {
	CreateTeam(team *models.Team) (int64, error)
	DeleteTeamsByEventAndSportID(eventID int, sportID int) error
	GetTeamsByUserID(userID string) ([]*models.TeamWithSport, error)
	GetTeamsByClassID(classID int, eventID int) ([]*models.TeamWithSport, error)
	GetTeamByClassAndSport(classID int, sportID int, eventID int) (*models.Team, error)
	AddTeamMember(teamID int, userID string) error
	GetTeamMembers(teamID int) ([]*models.User, error)
	RemoveTeamMember(teamID int, userID string) error
	UpdateTeamCapacity(eventID int, sportID int, classID int, minCapacity *int, maxCapacity *int) error
	GetTeamCapacity(eventID int, sportID int, classID int) (*models.Team, error)
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

func (r *teamRepository) GetTeamsByClassID(classID int, eventID int) ([]*models.TeamWithSport, error) {
	query := `
		SELECT t.id, t.name, t.class_id, t.sport_id, t.event_id, s.name as sport_name
		FROM teams t
		INNER JOIN sports s ON t.sport_id = s.id
		WHERE t.class_id = ? AND t.event_id = ?
	`
	rows, err := r.db.Query(query, classID, eventID)
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

func (r *teamRepository) GetTeamByClassAndSport(classID int, sportID int, eventID int) (*models.Team, error) {
	query := "SELECT id, name, class_id, sport_id, event_id, min_capacity, max_capacity FROM teams WHERE class_id = ? AND sport_id = ? AND event_id = ?"
	row := r.db.QueryRow(query, classID, sportID, eventID)

	team := &models.Team{}
	var minCap, maxCap sql.NullInt64
	err := row.Scan(&team.ID, &team.Name, &team.ClassID, &team.SportID, &team.EventID, &minCap, &maxCap)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Team not found
		}
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

	return team, nil
}

func (r *teamRepository) AddTeamMember(teamID int, userID string) error {
	query := "INSERT INTO team_members (team_id, user_id) VALUES (?, ?)"
	_, err := r.db.Exec(query, teamID, userID)
	return err
}

func (r *teamRepository) GetTeamMembers(teamID int) ([]*models.User, error) {
	query := `
		SELECT u.id, u.email, u.display_name, u.class_id, u.is_profile_complete, u.created_at, u.updated_at
		FROM users u
		INNER JOIN team_members tm ON u.id = tm.user_id
		WHERE tm.team_id = ?
		ORDER BY u.display_name, u.email
	`
	rows, err := r.db.Query(query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		var tempClassID sql.NullInt32
		var tempDisplayName sql.NullString

		err := rows.Scan(&user.ID, &user.Email, &tempDisplayName, &tempClassID, &user.IsProfileComplete, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if tempDisplayName.Valid {
			user.DisplayName = &tempDisplayName.String
		}
		if tempClassID.Valid {
			val := int(tempClassID.Int32)
			user.ClassID = &val
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *teamRepository) RemoveTeamMember(teamID int, userID string) error {
	query := "DELETE FROM team_members WHERE team_id = ? AND user_id = ?"
	_, err := r.db.Exec(query, teamID, userID)
	return err
}

func (r *teamRepository) UpdateTeamCapacity(eventID int, sportID int, classID int, minCapacity *int, maxCapacity *int) error {
	query := "UPDATE teams SET min_capacity = ?, max_capacity = ? WHERE event_id = ? AND sport_id = ? AND class_id = ?"
	_, err := r.db.Exec(query, minCapacity, maxCapacity, eventID, sportID, classID)
	return err
}

func (r *teamRepository) GetTeamCapacity(eventID int, sportID int, classID int) (*models.Team, error) {
	query := "SELECT id, name, class_id, sport_id, event_id, min_capacity, max_capacity FROM teams WHERE event_id = ? AND sport_id = ? AND class_id = ?"
	row := r.db.QueryRow(query, eventID, sportID, classID)

	team := &models.Team{}
	var minCap, maxCap sql.NullInt64
	err := row.Scan(&team.ID, &team.Name, &team.ClassID, &team.SportID, &team.EventID, &minCap, &maxCap)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Team not found
		}
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

	return team, nil
}
