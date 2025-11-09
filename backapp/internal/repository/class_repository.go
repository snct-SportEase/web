package repository

import (
	"backapp/internal/models"
	"database/sql"
	"fmt"
)

type ClassRepository interface {
	GetAllClasses(eventID int) ([]*models.Class, error)
	GetClassByID(id int) (*models.Class, error)
	GetClassDetails(classID int, eventID int) (*models.ClassDetails, error)
	UpdateAttendance(classID int, eventID int, attendanceCount int) (int, error)
	UpdateStudentCounts(eventID int, counts map[int]int) error
	CreateClasses(eventID int, classNames []string) error
	GetClassScoresByEvent(eventID int) ([]*models.ClassScore, error)
	UpdateClassRanks(eventID int) error
	GetClassByRepRole(userID string, eventID int) (*models.Class, error)
	GetClassMembers(classID int) ([]*models.User, error)
	SetNoonGamePoints(eventID int, points map[int]int) error
}

type classRepository struct {
	db *sql.DB
}

func NewClassRepository(db *sql.DB) ClassRepository {
	return &classRepository{db: db}
}

func (r *classRepository) CreateClasses(eventID int, classNames []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO classes (event_id, name) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, name := range classNames {
		_, err := stmt.Exec(eventID, name)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *classRepository) GetAllClasses(eventID int) ([]*models.Class, error) {
	var rows *sql.Rows
	var err error

	rows, err = r.db.Query("SELECT id, event_id, name, student_count, attend_count FROM classes WHERE event_id = ? ORDER BY name", eventID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []*models.Class
	for rows.Next() {
		class := &models.Class{}
		if err := rows.Scan(&class.ID, &class.EventID, &class.Name, &class.StudentCount, &class.AttendCount); err != nil {
			return nil, err
		}
		classes = append(classes, class)
	}

	return classes, nil
}

func (r *classRepository) GetClassByID(id int) (*models.Class, error) {
	row := r.db.QueryRow("SELECT id, event_id, name, student_count, attend_count FROM classes WHERE id = ?", id)

	class := &models.Class{}
	err := row.Scan(&class.ID, &class.EventID, &class.Name, &class.StudentCount, &class.AttendCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Class not found
		}
		return nil, err
	}
	return class, nil
}

func (r *classRepository) GetClassDetails(classID int, eventID int) (*models.ClassDetails, error) {
	query := `
        SELECT c.id, c.name, c.student_count, COALESCE(cs.attendance_points, 0)
        FROM classes c
        LEFT JOIN class_scores cs ON c.id = cs.class_id AND cs.event_id = ?
        WHERE c.id = ? AND c.event_id = ?
    `

	row := r.db.QueryRow(query, eventID, classID, eventID)

	details := &models.ClassDetails{}
	err := row.Scan(&details.ID, &details.Name, &details.StudentCount, &details.AttendancePoints)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Class not found
		}
		return nil, err
	}

	return details, nil
}

func (r *classRepository) UpdateAttendance(classID int, eventID int, attendanceCount int) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	// 1. Get student_count and name from classes table
	var studentCount int
	var className string
	row := tx.QueryRow("SELECT student_count, name FROM classes WHERE id = ? AND event_id = ?", classID, eventID)
	if err := row.Scan(&studentCount, &className); err != nil {
		tx.Rollback()
		return 0, err
	}

	// 2. Update attend_count in classes table
	_, err = tx.Exec("UPDATE classes SET attend_count = ? WHERE id = ? AND event_id = ?", attendanceCount, classID, eventID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// 3. Calculate points
	var points int
	var attendanceRate float64

	if className == "専教" {
		points = 0
	} else {
		if studentCount == 0 {
			tx.Rollback()
			return 0, fmt.Errorf("class '%s' (ID %d) has zero students, cannot calculate attendance points", className, classID)
		}
		attendanceRate = float64(attendanceCount) / float64(studentCount)
		switch {
		case attendanceRate >= 0.9:
			points = 10
		case attendanceRate >= 0.8:
			points = 9
		case attendanceRate >= 0.7:
			points = 8
		case attendanceRate >= 0.6:
			points = 7
		case attendanceRate >= 0.5:
			points = 6
		default:
			points = 5
		}
	}

	// 4. Update attendance_points in class_scores table
	// Note: BEFORE UPDATE trigger will automatically calculate total_points_current_event and total_points_overall
	scoreQuery := `
        INSERT INTO class_scores (event_id, class_id, attendance_points)
        VALUES (?, ?, ?)
        ON DUPLICATE KEY UPDATE attendance_points = VALUES(attendance_points)
    `
	_, err = tx.Exec(scoreQuery, eventID, classID, points)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// 5. Update ranks for all classes in the event
	if err := r.updateClassRanksInTransaction(tx, eventID); err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to update ranks: %w", err)
	}

	// 6. Log the score update
	logQuery := `
        INSERT INTO score_logs (event_id, class_id, points, reason)
        VALUES (?, ?, ?, ?)
    `
	var reason string
	if className == "専教" {
		reason = "Attendance points for faculty team are fixed to 0."
	} else {
		reason = fmt.Sprintf("Attendance points updated based on attendance rate (%.2f%%)", attendanceRate*100)
	}
	_, err = tx.Exec(logQuery, eventID, classID, points, reason)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return points, tx.Commit()
}

func (r *classRepository) UpdateStudentCounts(eventID int, counts map[int]int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback on error

	stmt, err := tx.Prepare("UPDATE classes SET student_count = ? WHERE id = ? AND event_id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for classID, count := range counts {
		_, err := stmt.Exec(count, classID, eventID)
		if err != nil {
			return fmt.Errorf("failed to update student_count for class %d: %w", classID, err)
		}
	}

	return tx.Commit()
}

func (r *classRepository) GetClassScoresByEvent(eventID int) ([]*models.ClassScore, error) {
	// Fetch sport mappings for the event.
	sportMap := make(map[string]string) // map[location]sportName
	sportRows, err := r.db.Query(`
		SELECT es.location, s.name
		FROM event_sports es
		JOIN sports s ON es.sport_id = s.id
		WHERE es.event_id = ?
	`, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sport mappings: %w", err)
	}
	defer sportRows.Close()

	for sportRows.Next() {
		var location, sportName string
		if err := sportRows.Scan(&location, &sportName); err != nil {
			return nil, fmt.Errorf("failed to scan sport mapping: %w", err)
		}
		sportMap[location] = sportName
	}
	if err = sportRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sport mappings: %w", err)
	}

	// Fetch class scores
	query := `
		SELECT
			cs.id,
			cs.event_id,
			cs.class_id,
			c.name as class_name,
			e.season,
			cs.initial_points,
			cs.survey_points,
			cs.attendance_points,
			cs.gym1_win1_points,
			cs.gym1_win2_points,
			cs.gym1_win3_points,
			cs.gym1_champion_points,
			cs.gym2_win1_points,
			cs.gym2_win2_points,
			cs.gym2_win3_points,
			cs.gym2_champion_points,
			cs.ground_win1_points,
			cs.ground_win2_points,
			cs.ground_win3_points,
			cs.ground_champion_points,
			cs.noon_game_points,
			cs.total_points_current_event,
			cs.rank_current_event,
			cs.total_points_overall,
			cs.rank_overall
		FROM class_scores cs
		JOIN classes c ON cs.class_id = c.id
		JOIN events e ON cs.event_id = e.id
		WHERE cs.event_id = ?
		ORDER BY cs.rank_overall, cs.rank_current_event
	`

	rows, err := r.db.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []*models.ClassScore
	for rows.Next() {
		score := &models.ClassScore{}
		if err := rows.Scan(
			&score.ID,
			&score.EventID,
			&score.ClassID,
			&score.ClassName,
			&score.Season,
			&score.InitialPoints,
			&score.SurveyPoints,
			&score.AttendancePoints,
			&score.Gym1Win1Points,
			&score.Gym1Win2Points,
			&score.Gym1Win3Points,
			&score.Gym1ChampionPoints,
			&score.Gym2Win1Points,
			&score.Gym2Win2Points,
			&score.Gym2Win3Points,
			&score.Gym2ChampionPoints,
			&score.GroundWin1Points,
			&score.GroundWin2Points,
			&score.GroundWin3Points,
			&score.GroundChampionPoints,
			&score.NoonGamePoints,
			&score.TotalPointsCurrentEvent,
			&score.RankCurrentEvent,
			&score.TotalPointsOverall,
			&score.RankOverall,
		); err != nil {
			return nil, err
		}
		score.SportNames = sportMap // Assign the fetched sport map
		scores = append(scores, score)
	}

	return scores, nil
}

// updateClassRanksInTransaction updates class ranks within a transaction
func (r *classRepository) updateClassRanksInTransaction(tx *sql.Tx, eventID int) error {
	// Update rank_current_event
	updateCurrentRankQuery := `
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
	`
	_, err := tx.Exec(updateCurrentRankQuery, eventID, eventID)
	if err != nil {
		return fmt.Errorf("failed to update current event ranks: %w", err)
	}

	// Update rank_overall
	updateOverallRankQuery := `
		UPDATE class_scores cs
		JOIN (
			SELECT
				class_id,
				RANK() OVER (ORDER BY total_points_overall DESC) AS new_rank
			FROM class_scores
			WHERE event_id = ?
		) ranked_data ON cs.class_id = ranked_data.class_id
		SET cs.rank_overall = ranked_data.new_rank
		WHERE cs.event_id = ?
	`
	_, err = tx.Exec(updateOverallRankQuery, eventID, eventID)
	if err != nil {
		return fmt.Errorf("failed to update overall ranks: %w", err)
	}

	return nil
}

// UpdateClassRanks updates class ranks for the given event
func (r *classRepository) UpdateClassRanks(eventID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := r.updateClassRanksInTransaction(tx, eventID); err != nil {
		return err
	}

	return tx.Commit()
}

// GetClassByRepRole gets the class that a user with class_name_rep role can manage
func (r *classRepository) GetClassByRepRole(userID string, eventID int) (*models.Class, error) {
	query := `
		SELECT c.id, c.event_id, c.name, c.student_count, c.attend_count
		FROM classes c
		INNER JOIN user_roles ur ON ur.user_id = ?
		INNER JOIN roles ro ON ur.role_id = ro.id
		WHERE ro.name = CONCAT(c.name, '_rep') 
		AND (ur.event_id = ? OR ur.event_id IS NULL)
		AND c.event_id = ?
		LIMIT 1
	`
	row := r.db.QueryRow(query, userID, eventID, eventID)

	class := &models.Class{}
	var eventIDPtr *int
	err := row.Scan(&class.ID, &eventIDPtr, &class.Name, &class.StudentCount, &class.AttendCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Class not found
		}
		return nil, err
	}
	class.EventID = eventIDPtr
	return class, nil
}

// GetClassMembers gets all users in a class
func (r *classRepository) GetClassMembers(classID int) ([]*models.User, error) {
	query := `
		SELECT id, email, display_name, class_id, is_profile_complete, created_at, updated_at
		FROM users
		WHERE class_id = ?
		ORDER BY display_name, email
	`
	rows, err := r.db.Query(query, classID)
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

func (r *classRepository) SetNoonGamePoints(eventID int, points map[int]int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE class_scores SET noon_game_points = 0 WHERE event_id = ?", eventID); err != nil {
		return fmt.Errorf("failed to reset noon_game_points: %w", err)
	}

	if len(points) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO class_scores (event_id, class_id, noon_game_points)
			VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE noon_game_points = VALUES(noon_game_points)
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare noon_game_points statement: %w", err)
		}
		defer stmt.Close()

		for classID, value := range points {
			if _, err := stmt.Exec(eventID, classID, value); err != nil {
				return fmt.Errorf("failed to update noon_game_points for class %d: %w", classID, err)
			}
		}
	}

	if err := r.updateClassRanksInTransaction(tx, eventID); err != nil {
		return fmt.Errorf("failed to update class ranks: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit noon_game_points transaction: %w", err)
	}

	return nil
}
