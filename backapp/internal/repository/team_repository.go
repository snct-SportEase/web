package repository

import (
	"backapp/internal/models"
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

var ErrRoundAlreadyCheckedIn = errors.New("round already checked in")

type TeamRepository interface {
	CreateTeam(team *models.Team) (int64, error)
	DeleteTeamsByEventAndSportID(eventID int, sportID int) error
	GetTeamsByUserID(userID string) ([]*models.TeamWithSport, error)
	GetTeamsByClassID(classID int, eventID int) ([]*models.TeamWithSport, error)
	GetNoonGameTeamsByClassID(classID int, eventID int) ([]*models.TeamWithSport, error)
	GetTeamByClassAndSport(classID int, sportID int, eventID int) (*models.Team, error)
	AddTeamMember(teamID int, userID string) error
	GetTeamMembers(teamID int) ([]*models.User, error)
	GetTeamMembersByTeamIDs(teamIDs []int) (map[int][]*models.User, error)
	RemoveTeamMember(teamID int, userID string) error
	UpdateTeamCapacity(eventID int, sportID int, classID int, minCapacity *int, maxCapacity *int) error
	GetTeamCapacity(eventID int, sportID int, classID int) (*models.Team, error)
	ConfirmTeamMember(teamID int, userID string) error
	GetConfirmedTeamMembers(teamID int) ([]*models.User, error)
	GetConfirmedTeamMembersCount(teamID int) (int, error)
	CheckInRound(teamID int, userID string, eventID int, sportID int, matchID int, round int) error
	GetMatchCheckIns(eventID int, sportID int, matchID int) ([]*models.MatchCheckInMember, error)
	CreateTeamsBulk(teams []*models.Team) error
}

type teamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) CreateTeam(team *models.Team) (int64, error) {
	query := "INSERT INTO teams (name, class_id, sport_id) VALUES (?, ?, ?)"
	result, err := r.db.Exec(query, team.Name, team.ClassID, team.SportID)
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
	query := `
		DELETE t FROM teams t 
		JOIN classes c ON t.class_id = c.id 
		WHERE c.event_id = ? AND t.sport_id = ?
	`
	_, err := r.db.Exec(query, eventID, sportID)
	return err
}

func (r *teamRepository) GetTeamsByUserID(userID string) ([]*models.TeamWithSport, error) {
	query := `
		SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id, s.name as sport_name
		FROM teams t
		INNER JOIN team_members tm ON t.id = tm.team_id
		INNER JOIN sports s ON t.sport_id = s.id
		INNER JOIN classes c ON t.class_id = c.id
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
		SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id, s.name as sport_name
		FROM teams t
		INNER JOIN sports s ON t.sport_id = s.id
		INNER JOIN classes c ON t.class_id = c.id
		INNER JOIN event_sports es ON es.event_id = c.event_id AND es.sport_id = t.sport_id
		WHERE t.class_id = ? AND c.event_id = ? AND es.location <> 'noon_game'
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

func (r *teamRepository) GetNoonGameTeamsByClassID(classID int, eventID int) ([]*models.TeamWithSport, error) {
	query := `
		SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id, s.name as sport_name
		FROM teams t
		INNER JOIN sports s ON t.sport_id = s.id
		INNER JOIN classes c ON t.class_id = c.id
		INNER JOIN event_sports es ON es.event_id = c.event_id AND es.sport_id = t.sport_id
		WHERE t.class_id = ? AND c.event_id = ? AND es.location = 'noon_game'
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
	query := "SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id, t.min_capacity, t.max_capacity FROM teams t JOIN classes c ON t.class_id = c.id WHERE t.class_id = ? AND t.sport_id = ? AND c.event_id = ?"
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

func (r *teamRepository) GetTeamMembersByTeamIDs(teamIDs []int) (map[int][]*models.User, error) {
	result := make(map[int][]*models.User)
	if len(teamIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(teamIDs))
	args := make([]interface{}, len(teamIDs))
	for i, teamID := range teamIDs {
		placeholders[i] = "?"
		args[i] = teamID
	}

	query := `
		SELECT tm.team_id, u.id, u.email, u.display_name, u.class_id, u.is_profile_complete, u.created_at, u.updated_at
		FROM users u
		INNER JOIN team_members tm ON u.id = tm.user_id
		WHERE tm.team_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY tm.team_id, u.display_name, u.email
	`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var teamID int
		user := &models.User{}
		var tempClassID sql.NullInt32
		var tempDisplayName sql.NullString

		if err := rows.Scan(&teamID, &user.ID, &user.Email, &tempDisplayName, &tempClassID, &user.IsProfileComplete, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}

		if tempDisplayName.Valid {
			user.DisplayName = &tempDisplayName.String
		}
		if tempClassID.Valid {
			val := int(tempClassID.Int32)
			user.ClassID = &val
		}

		result[teamID] = append(result[teamID], user)
	}

	return result, rows.Err()
}

func (r *teamRepository) RemoveTeamMember(teamID int, userID string) error {
	query := "DELETE FROM team_members WHERE team_id = ? AND user_id = ?"
	_, err := r.db.Exec(query, teamID, userID)
	return err
}

func (r *teamRepository) UpdateTeamCapacity(eventID int, sportID int, classID int, minCapacity *int, maxCapacity *int) error {
	query := "UPDATE teams SET min_capacity = ?, max_capacity = ? WHERE sport_id = ? AND class_id = ?"
	_, err := r.db.Exec(query, minCapacity, maxCapacity, sportID, classID)
	return err
}

func (r *teamRepository) GetTeamCapacity(eventID int, sportID int, classID int) (*models.Team, error) {
	query := `
		SELECT t.id, t.name, t.class_id, t.sport_id, c.event_id, t.min_capacity, t.max_capacity 
		FROM teams t 
		JOIN classes c ON t.class_id = c.id 
		WHERE c.event_id = ? AND t.sport_id = ? AND t.class_id = ?
	`
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

// ConfirmTeamMember marks a team member as confirmed (参加本登録)
func (r *teamRepository) ConfirmTeamMember(teamID int, userID string) error {
	query := "UPDATE team_members SET is_confirmed = true WHERE team_id = ? AND user_id = ?"
	result, err := r.db.Exec(query, teamID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		var exists int
		err := r.db.QueryRow("SELECT COUNT(*) FROM team_members WHERE team_id = ? AND user_id = ?", teamID, userID).Scan(&exists)
		if err != nil {
			return err
		}
		if exists == 0 {
			return sql.ErrNoRows // Team member not found
		}
	}
	return nil
}

// GetConfirmedTeamMembers returns all confirmed team members
func (r *teamRepository) GetConfirmedTeamMembers(teamID int) ([]*models.User, error) {
	query := `
		SELECT u.id, u.email, u.display_name, u.class_id, u.is_profile_complete, u.created_at, u.updated_at
		FROM users u
		INNER JOIN team_members tm ON u.id = tm.user_id
		WHERE tm.team_id = ? AND tm.is_confirmed = true
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

// GetConfirmedTeamMembersCount returns the count of confirmed team members
func (r *teamRepository) GetConfirmedTeamMembersCount(teamID int) (int, error) {
	query := "SELECT COUNT(*) FROM team_members WHERE team_id = ? AND is_confirmed = true"
	var count int
	err := r.db.QueryRow(query, teamID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CheckInRound records that a pre-entered student checked in for a specific event/sport match round.
func (r *teamRepository) CheckInRound(teamID int, userID string, eventID int, sportID int, matchID int, round int) error {
	query := `
		INSERT INTO round_check_ins (event_id, sport_id, match_id, round, user_id, team_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, eventID, sportID, matchID, round, userID, teamID)
	if err != nil {
		if isMySQLDuplicateEntryError(err) {
			return ErrRoundAlreadyCheckedIn
		}
		return err
	}
	return nil
}

func isMySQLDuplicateEntryError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

// GetMatchCheckIns returns students checked in for a selected match.
func (r *teamRepository) GetMatchCheckIns(eventID int, sportID int, matchID int) ([]*models.MatchCheckInMember, error) {
	query := `
		SELECT
			rci.user_id,
			u.email,
			u.display_name,
			t.class_id,
			c.name AS class_name,
			rci.team_id,
			t.name AS team_name,
			rci.event_id,
			rci.sport_id,
			rci.match_id,
			rci.round,
			rci.checked_in_at
		FROM round_check_ins rci
		JOIN users u ON u.id = rci.user_id
		JOIN teams t ON t.id = rci.team_id
		JOIN classes c ON c.id = t.class_id
		WHERE rci.event_id = ? AND rci.sport_id = ? AND rci.match_id = ?
		ORDER BY rci.checked_in_at DESC, u.display_name ASC, u.email ASC
	`
	rows, err := r.db.Query(query, eventID, sportID, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.MatchCheckInMember
	for rows.Next() {
		member := &models.MatchCheckInMember{}
		var displayName sql.NullString
		if err := rows.Scan(
			&member.UserID,
			&member.Email,
			&displayName,
			&member.ClassID,
			&member.ClassName,
			&member.TeamID,
			&member.TeamName,
			&member.EventID,
			&member.SportID,
			&member.MatchID,
			&member.Round,
			&member.CheckedInAt,
		); err != nil {
			return nil, err
		}
		if displayName.Valid {
			member.DisplayName = &displayName.String
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (r *teamRepository) CreateTeamsBulk(teams []*models.Team) error {
	if len(teams) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("INSERT INTO teams (name, class_id, sport_id) VALUES ")

	args := make([]interface{}, 0, len(teams)*3)
	for i, team := range teams {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("(?, ?, ?)")
		args = append(args, team.Name, team.ClassID, team.SportID)
	}

	_, err := r.db.Exec(sb.String(), args...)
	return err
}
