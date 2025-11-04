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

	// 5. Update total_points_current_event
	updateTotalQuery := `
        UPDATE class_scores SET
        total_points_current_event = COALESCE(initial_points, 0) + COALESCE(survey_points, 0) + COALESCE(attendance_points, 0) +
                                   COALESCE(gym1_win1_points, 0) + COALESCE(gym1_win2_points, 0) + COALESCE(gym1_win3_points, 0) + COALESCE(gym1_champion_points, 0) +
                                   COALESCE(gym2_win1_points, 0) + COALESCE(gym2_win2_points, 0) + COALESCE(gym2_win3_points, 0) + COALESCE(gym2_champion_points, 0) +
                                   COALESCE(ground_win1_points, 0) + COALESCE(ground_win2_points, 0) + COALESCE(ground_win3_points, 0) + COALESCE(ground_champion_points, 0) +
                                   COALESCE(noon_game_points, 0) + COALESCE(mvp_points, 0)
        WHERE event_id = ? AND class_id = ?
    `
	_, err = tx.Exec(updateTotalQuery, eventID, classID)
	if err != nil {
		tx.Rollback()
		return 0, err
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
		scores = append(scores, score)
	}

	return scores, nil
}
