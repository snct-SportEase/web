package repository

import (
	"backapp/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrClassNotFound = errors.New("class not found in event")

type ClassRepository interface {
	GetAllClasses(eventID int) ([]*models.Class, error)
	GetClassByID(id int) (*models.Class, error)
	GetClassDetails(classID int, eventID int) (*models.ClassDetails, error)
	UpdateAttendance(classID int, eventID int, attendanceCount int) (int, error)
	UpdateStudentCounts(eventID int, counts map[int]int) error
	CreateClasses(eventID int, classNames []string) error
	GetClassScoresByEvent(eventID int) ([]*models.ClassScore, error)
	GetClassScoresByEvents(eventIDs []int) (map[int][]*models.ClassScore, error)
	UpdateClassRanks(eventID int) error
	GetClassMembers(classID int) ([]*models.User, error)
	CountClassMembers(classID int) (int, error)
	SetNoonGamePoints(eventID int, points map[int]int) error
	SetSurveyPoints(eventID int, points map[int]int) error
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

	classes := make([]*models.Class, 0)
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

	// Delete old attendance points
	_, err = tx.Exec("DELETE FROM score_logs WHERE event_id = ? AND class_id = ? AND reason = 'attendance_points'", eventID, classID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Insert new attendance points into score_logs
	logQuery := `
        INSERT INTO score_logs (event_id, class_id, points, reason)
        VALUES (?, ?, ?, ?)
    `
	reason := "attendance_points"
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

	stmt, err := tx.Prepare("UPDATE classes SET student_count = ? WHERE id = ? AND event_id = ? AND name <> '専教'")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for classID, count := range counts {
		result, err := stmt.Exec(count, classID, eventID)
		if err != nil {
			return fmt.Errorf("failed to update student_count for class %d: %w", classID, err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to inspect student_count update for class %d: %w", classID, err)
		}
		if affected == 0 {
			return fmt.Errorf("%w: %d", ErrClassNotFound, classID)
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
			cs.mic_points,
			cs.gym1_win1_points,
			cs.gym1_win2_points,
			cs.gym1_win3_points,
			cs.gym1_champion_points,
		cs.gym2_win1_points,
		cs.gym2_win2_points,
		cs.gym2_win3_points,
		cs.gym2_champion_points,
		cs.gym2_loser_bracket_champion_points,
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
			&score.MicPoints,
			&score.Gym1Win1Points,
			&score.Gym1Win2Points,
			&score.Gym1Win3Points,
			&score.Gym1ChampionPoints,
			&score.Gym2Win1Points,
			&score.Gym2Win2Points,
			&score.Gym2Win3Points,
			&score.Gym2ChampionPoints,
			&score.Gym2LoserBracketChampionPoints,
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

	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	if err := r.attachSportScores([]int{eventID}, map[int][]*models.ClassScore{eventID: scores}); err != nil {
		return nil, err
	}
	return scores, nil
}

func (r *classRepository) GetClassScoresByEvents(eventIDs []int) (map[int][]*models.ClassScore, error) {
	result := make(map[int][]*models.ClassScore, len(eventIDs))
	if len(eventIDs) == 0 {
		return result, nil
	}

	// eventIDs は呼び出し側で取得した数値IDのみなので、プレースホルダでIN句を組み立てる。
	placeholders := strings.TrimRight(strings.Repeat("?,", len(eventIDs)), ",")
	args := make([]interface{}, 0, len(eventIDs))
	for _, id := range eventIDs {
		args = append(args, id)
		result[id] = []*models.ClassScore{}
	}

	sportMaps := make(map[int]map[string]string, len(eventIDs))
	sportRows, err := r.db.Query(fmt.Sprintf(`
		SELECT es.event_id, es.location, s.name
		FROM event_sports es
		JOIN sports s ON es.sport_id = s.id
		WHERE es.event_id IN (%s)
	`, placeholders), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get sport mappings: %w", err)
	}
	defer sportRows.Close()

	for sportRows.Next() {
		var eventID int
		var location, sportName string
		if err := sportRows.Scan(&eventID, &location, &sportName); err != nil {
			return nil, fmt.Errorf("failed to scan sport mapping: %w", err)
		}
		if sportMaps[eventID] == nil {
			sportMaps[eventID] = make(map[string]string)
		}
		sportMaps[eventID][location] = sportName
	}
	if err = sportRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sport mappings: %w", err)
	}

	// #nosec G201 -- placeholders is generated internally from an integer ID slice; values are bound below.
	query := fmt.Sprintf(`
		SELECT
			cs.id,
			cs.event_id,
			cs.class_id,
			c.name as class_name,
			e.season,
			cs.initial_points,
			cs.survey_points,
			cs.attendance_points,
			cs.mic_points,
			cs.gym1_win1_points,
			cs.gym1_win2_points,
			cs.gym1_win3_points,
			cs.gym1_champion_points,
			cs.gym2_win1_points,
			cs.gym2_win2_points,
			cs.gym2_win3_points,
			cs.gym2_champion_points,
			cs.gym2_loser_bracket_champion_points,
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
		WHERE cs.event_id IN (%s)
		ORDER BY cs.event_id, cs.rank_overall, cs.rank_current_event
	`, placeholders)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
			&score.MicPoints,
			&score.Gym1Win1Points,
			&score.Gym1Win2Points,
			&score.Gym1Win3Points,
			&score.Gym1ChampionPoints,
			&score.Gym2Win1Points,
			&score.Gym2Win2Points,
			&score.Gym2Win3Points,
			&score.Gym2ChampionPoints,
			&score.Gym2LoserBracketChampionPoints,
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
		score.SportNames = sportMaps[score.EventID]
		result[score.EventID] = append(result[score.EventID], score)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	rows.Close()
	if err := r.attachSportScores(eventIDs, result); err != nil {
		return nil, err
	}
	return result, nil
}

// attachSportScores includes assigned competitions even before their first result.
// Ordinary sports are separated by sport ID. Board games are additionally
// separated by tournament so shogi's A and B blocks are displayed independently.
func (r *classRepository) attachSportScores(eventIDs []int, scores map[int][]*models.ClassScore) error {
	placeholders := strings.TrimRight(strings.Repeat("?,", len(eventIDs)), ",")
	args := make([]interface{}, len(eventIDs))
	for i, id := range eventIDs {
		args[i] = id
	}
	// #nosec G201 -- only internally generated placeholders are interpolated; IDs are bound.
	rows, err := r.db.Query(fmt.Sprintf(`
		SELECT es.event_id, c.id, s.id, s.name,
			0 AS tournament_id, '' AS tournament_name, '' AS slot_key, FALSE AS is_board_game,
			COALESCE(sl.reason, '') AS reason, COALESCE(SUM(sl.points), 0) AS points
		FROM event_sports es
		JOIN sports s ON s.id=es.sport_id
		JOIN classes c ON c.event_id=es.event_id
		LEFT JOIN score_logs sl ON sl.event_id=es.event_id AND sl.class_id=c.id AND sl.sport_id=es.sport_id
		WHERE es.event_id IN (%s) AND es.location <> 'noon_game'
			AND NOT EXISTS (
				SELECT 1
				FROM board_game_runs bg
				JOIN sports board_sport ON board_sport.id=bg.sport_id
				WHERE bg.event_id=es.event_id
					AND (bg.sport_id=es.sport_id OR TRIM(board_sport.name)=TRIM(s.name))
			)
		GROUP BY es.event_id, es.sport_id, c.id, s.id, s.name, sl.reason

		UNION ALL

		SELECT bg.event_id, c.id, s.id, s.name,
			block.tournament_id, block.tournament_name, block.slot_key, TRUE AS is_board_game,
			COALESCE(sl.reason, '') AS reason, COALESCE(SUM(sl.points), 0) AS points
		FROM board_game_runs bg
		JOIN sports s ON s.id=bg.sport_id
		JOIN classes c ON c.event_id=bg.event_id
		JOIN (
			SELECT DISTINCT entry.run_id, entry.tournament_id, tournament.name AS tournament_name, entry.slot_key
			FROM board_game_entries entry
			JOIN tournaments tournament ON tournament.id=entry.tournament_id
		) block ON block.run_id=bg.id
		LEFT JOIN score_logs sl ON sl.event_id=bg.event_id
			AND sl.class_id=c.id
			AND sl.sport_id=bg.sport_id
			AND sl.board_game_run_id=bg.id
			AND EXISTS (
				SELECT 1 FROM matches source_match
				WHERE source_match.id=sl.source_match_id
					AND source_match.tournament_id=block.tournament_id
			)
		WHERE bg.event_id IN (%s)
		GROUP BY bg.event_id, c.id, s.id, s.name,
			block.tournament_id, block.tournament_name, block.slot_key, sl.reason

		ORDER BY 1, 2, 3, 5, 9
	`, placeholders, placeholders), append(args, args...)...)
	if err != nil {
		return fmt.Errorf("failed to get sport scores: %w", err)
	}
	defer rows.Close()
	byClass := make(map[int]map[int]*models.ClassScore)
	for eventID, eventScores := range scores {
		byClass[eventID] = make(map[int]*models.ClassScore)
		for _, score := range eventScores {
			score.SportScores = []models.SportScore{}
			byClass[eventID][score.ClassID] = score
		}
	}
	for rows.Next() {
		var eventID, classID, sportID, tournamentID, points int
		var name, tournamentName, slotKey, reason string
		var boardGame bool
		if err := rows.Scan(&eventID, &classID, &sportID, &name, &tournamentID, &tournamentName, &slotKey, &boardGame, &reason, &points); err != nil {
			return err
		}
		score := byClass[eventID][classID]
		if score == nil {
			continue
		}
		if len(score.SportScores) == 0 ||
			score.SportScores[len(score.SportScores)-1].SportID != sportID ||
			score.SportScores[len(score.SportScores)-1].TournamentID != tournamentID {
			score.SportScores = append(score.SportScores, models.SportScore{
				SportID:        sportID,
				SportName:      name,
				TournamentID:   tournamentID,
				TournamentName: tournamentName,
				SlotKey:        slotKey,
				IsBoardGame:    boardGame,
			})
		}
		sport := &score.SportScores[len(score.SportScores)-1]
		switch {
		case reason == "board_game_win_points":
			sport.WinPoints += points
		case reason == "board_game_rank_points":
			sport.RankPoints += points
		case strings.HasSuffix(reason, "_loser_bracket_champion_points"):
			sport.LoserBracketChampionPoints += points
		case strings.HasSuffix(reason, "_win1_points"):
			sport.Win1Points += points
		case strings.HasSuffix(reason, "_win2_points"):
			sport.Win2Points += points
		case strings.HasSuffix(reason, "_win3_points"):
			sport.Win3Points += points
		case strings.HasSuffix(reason, "_champion_points"):
			sport.ChampionPoints += points
		}
	}
	return rows.Err()
}

func (r *classRepository) UpdateClassRanks(eventID int) error {
	// class_scores is a VIEW, ranking is dynamic
	return nil
}

// CountClassMembers avoids loading and sorting personal data for summary views.
func (r *classRepository) CountClassMembers(classID int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE class_id = ?", classID).Scan(&count)
	return count, err
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

	users := make([]*models.User, 0)
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

	return users, rows.Err()
}

func (r *classRepository) SetNoonGamePoints(eventID int, points map[int]int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM score_logs WHERE event_id = ? AND reason = 'noon_game_points'", eventID); err != nil {
		return fmt.Errorf("failed to reset noon_game_points: %w", err)
	}

	if len(points) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO score_logs (event_id, class_id, points, reason)
			VALUES (?, ?, ?, 'noon_game_points')
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit noon_game_points transaction: %w", err)
	}

	return nil
}

func (r *classRepository) SetSurveyPoints(eventID int, points map[int]int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear old points specifically for the event
	if _, err := tx.Exec("DELETE FROM score_logs WHERE event_id = ? AND reason = 'survey_points'", eventID); err != nil {
		return fmt.Errorf("failed to reset survey_points: %w", err)
	}

	if len(points) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO score_logs (event_id, class_id, points, reason)
			VALUES (?, ?, ?, 'survey_points')
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare survey_points statement: %w", err)
		}
		defer stmt.Close()

		for classID, value := range points {
			if _, err := stmt.Exec(eventID, classID, value); err != nil {
				return fmt.Errorf("failed to update survey_points for class %d: %w", classID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit survey_points transaction: %w", err)
	}

	return nil
}
