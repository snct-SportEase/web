package repository

import (
	"backapp/internal/models"
	"database/sql"
	"fmt"
)

type ClassRepository interface {
	GetAllClasses() ([]*models.Class, error)
	GetClassByID(id int) (*models.Class, error)
	GetClassDetails(classID int, eventID int) (*models.ClassDetails, error)
	UpdateAttendance(classID int, eventID int, attendanceCount int) (int, error)
}

type classRepository struct {
	db *sql.DB
}

func NewClassRepository(db *sql.DB) ClassRepository {
	return &classRepository{db: db}
}

func (r *classRepository) GetAllClasses() ([]*models.Class, error) {
	rows, err := r.db.Query("SELECT id, name, student_count, attend_count FROM classes ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []*models.Class
	for rows.Next() {
		class := &models.Class{}
		if err := rows.Scan(&class.ID, &class.Name, &class.StudentCount, &class.AttendCount); err != nil {
			return nil, err
		}
		classes = append(classes, class)
	}

	return classes, nil
}

func (r *classRepository) GetClassByID(id int) (*models.Class, error) {
	row := r.db.QueryRow("SELECT id, name, student_count, attend_count FROM classes WHERE id = ?", id)

	class := &models.Class{}
	err := row.Scan(&class.ID, &class.Name, &class.StudentCount, &class.AttendCount)
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
        WHERE c.id = ?
    `

	row := r.db.QueryRow(query, eventID, classID)

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

	// 1. Get student_count from classes table
	var studentCount int
	row := tx.QueryRow("SELECT student_count FROM classes WHERE id = ?", classID)
	if err := row.Scan(&studentCount); err != nil {
		tx.Rollback()
		return 0, err
	}

	if studentCount == 0 {
		tx.Rollback()
		return 0, fmt.Errorf("class with ID %d has zero students", classID)
	}

	// 2. Update attend_count in classes table
	_, err = tx.Exec("UPDATE classes SET attend_count = ? WHERE id = ?", attendanceCount, classID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// 3. Calculate points based on attendance rate
	attendanceRate := float64(attendanceCount) / float64(studentCount)
	points := 0
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
	reason := fmt.Sprintf("Attendance points updated based on attendance rate (%.2f%%)", attendanceRate*100)
	_, err = tx.Exec(logQuery, eventID, classID, points, reason)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	return points, tx.Commit()
}
