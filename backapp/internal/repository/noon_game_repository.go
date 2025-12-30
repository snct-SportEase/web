package repository

import (
	"backapp/internal/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type NoonGameRepository interface {
	GetSessionByID(sessionID int) (*models.NoonGameSession, error)
	GetSessionByEvent(eventID int) (*models.NoonGameSession, error)
	UpsertSession(session *models.NoonGameSession) (*models.NoonGameSession, error)

	GetGroupsWithMembers(sessionID int) ([]*models.NoonGameGroupWithMembers, error)
	GetGroupWithMembers(sessionID int, groupID int) (*models.NoonGameGroupWithMembers, error)
	SaveGroup(group *models.NoonGameGroup, memberClassIDs []int) (*models.NoonGameGroupWithMembers, error)
	DeleteGroup(sessionID int, groupID int) error

	GetMatchesWithResults(sessionID int) ([]*models.NoonGameMatchWithResult, error)
	GetMatchByID(matchID int) (*models.NoonGameMatchWithResult, error)
	SaveMatch(match *models.NoonGameMatch) (*models.NoonGameMatch, error)
	DeleteMatch(sessionID int, matchID int) error

	SaveResult(result *models.NoonGameResult) (*models.NoonGameResult, error)
	GetResultByMatchID(matchID int) (*models.NoonGameResult, error)

	ClearPointsForMatch(matchID int) error
	InsertPoints(points []*models.NoonGamePoint) error
	InsertPoint(point *models.NoonGamePoint) (*models.NoonGamePoint, error)
	SumPointsByClass(sessionID int) (map[int]int, error)

	GetGroupMembers(groupID int) ([]*models.NoonGameGroupMember, error)

	// --- Template runs ---
	CreateTemplateRun(run *models.NoonGameTemplateRun) (*models.NoonGameTemplateRun, error)
	CreateTemplateRunWithPointsByRankJSON(sessionID int, templateKey, name, createdBy string, pointsByRankJSON interface{}) (*models.NoonGameTemplateRun, error)
	GetTemplateRunByID(runID int) (*models.NoonGameTemplateRun, error)
	ListTemplateRunsBySession(sessionID int) ([]*models.NoonGameTemplateRun, error)
	ListTemplateRunMatches(runID int) ([]*models.NoonGameTemplateRunMatch, error)
	LinkTemplateRunMatch(runID int, matchID int, matchKey string) (*models.NoonGameTemplateRunMatch, error)
	GetTemplateRunMatchByKey(runID int, matchKey string) (*models.NoonGameTemplateRunMatch, error)
	GetTemplateRunMatchByMatchID(matchID int) (*models.NoonGameTemplateRunMatch, error)
	DeleteTemplateRunAndRelatedData(sessionID int) error
}

type noonGameRepository struct {
	db *sql.DB
}

func NewNoonGameRepository(db *sql.DB) NoonGameRepository {
	return &noonGameRepository{db: db}
}

func (r *noonGameRepository) CreateTemplateRun(run *models.NoonGameTemplateRun) (*models.NoonGameTemplateRun, error) {
	if run == nil {
		return nil, fmt.Errorf("run is nil")
	}
	if run.SessionID == 0 {
		return nil, fmt.Errorf("session_id is required")
	}
	if strings.TrimSpace(run.TemplateKey) == "" {
		return nil, fmt.Errorf("template_key is required")
	}
	if strings.TrimSpace(run.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	createdBy := strings.TrimSpace(run.CreatedBy)
	if createdBy == "" {
		return nil, fmt.Errorf("created_by is required")
	}
	if len(createdBy) != 36 {
		return nil, fmt.Errorf("created_by must be a valid UUID (36 characters), got %d characters", len(createdBy))
	}

	var pointsByRankJSON interface{}
	if run.PointsByRank != nil {
		jsonBytes, err := json.Marshal(run.PointsByRank)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal points_by_rank: %w", err)
		}
		pointsByRankJSON = string(jsonBytes)
	}

	res, err := r.db.Exec(`
		INSERT INTO noon_game_template_runs (session_id, template_key, name, created_by, points_by_rank)
		VALUES (?, ?, ?, ?, ?)
	`, run.SessionID, strings.TrimSpace(run.TemplateKey), strings.TrimSpace(run.Name), createdBy, pointsByRankJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to insert template run (session_id=%d, template_key=%q, name=%q, created_by=%q): %w", run.SessionID, run.TemplateKey, run.Name, createdBy, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}
	return r.GetTemplateRunByID(int(id))
}

func (r *noonGameRepository) CreateTemplateRunWithPointsByRankJSON(sessionID int, templateKey, name, createdBy string, pointsByRankJSON interface{}) (*models.NoonGameTemplateRun, error) {
	if sessionID == 0 {
		return nil, fmt.Errorf("session_id is required")
	}
	if strings.TrimSpace(templateKey) == "" {
		return nil, fmt.Errorf("template_key is required")
	}
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	createdByTrimmed := strings.TrimSpace(createdBy)
	if createdByTrimmed == "" {
		return nil, fmt.Errorf("created_by is required")
	}
	if len(createdByTrimmed) != 36 {
		return nil, fmt.Errorf("created_by must be a valid UUID (36 characters), got %d characters", len(createdByTrimmed))
	}

	var pointsByRankJSONVal interface{}
	if pointsByRankJSON != nil {
		// JSONとしてマーシャルしてから保存
		jsonBytes, err := json.Marshal(pointsByRankJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal points_by_rank: %w", err)
		}
		var unmarshaled interface{}
		if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
			return nil, fmt.Errorf("failed to validate points_by_rank JSON: %w", err)
		}
		pointsByRankJSONVal = string(jsonBytes)
	}

	res, err := r.db.Exec(`
		INSERT INTO noon_game_template_runs (session_id, template_key, name, created_by, points_by_rank)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, strings.TrimSpace(templateKey), strings.TrimSpace(name), createdByTrimmed, pointsByRankJSONVal)
	if err != nil {
		return nil, fmt.Errorf("failed to insert template run (session_id=%d, template_key=%q, name=%q, created_by=%q): %w", sessionID, templateKey, name, createdByTrimmed, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}
	return r.GetTemplateRunByID(int(id))
}

func (r *noonGameRepository) GetTemplateRunByID(runID int) (*models.NoonGameTemplateRun, error) {
	row := r.db.QueryRow(`
		SELECT id, session_id, template_key, name, created_by, created_at, updated_at, points_by_rank
		FROM noon_game_template_runs
		WHERE id = ?
	`, runID)

	run := &models.NoonGameTemplateRun{}
	var pointsByRankJSON sql.NullString
	if err := row.Scan(
		&run.ID,
		&run.SessionID,
		&run.TemplateKey,
		&run.Name,
		&run.CreatedBy,
		&run.CreatedAt,
		&run.UpdatedAt,
		&pointsByRankJSON,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if pointsByRankJSON.Valid && pointsByRankJSON.String != "" {
		var pointsByRank interface{}
		if err := json.Unmarshal([]byte(pointsByRankJSON.String), &pointsByRank); err != nil {
			return nil, fmt.Errorf("failed to unmarshal points_by_rank: %w", err)
		}
		run.PointsByRank = pointsByRank
	}

	return run, nil
}

func (r *noonGameRepository) ListTemplateRunsBySession(sessionID int) ([]*models.NoonGameTemplateRun, error) {
	rows, err := r.db.Query(`
		SELECT id, session_id, template_key, name, created_by, created_at, updated_at, points_by_rank
		FROM noon_game_template_runs
		WHERE session_id = ?
		ORDER BY created_at
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.NoonGameTemplateRun
	for rows.Next() {
		run := &models.NoonGameTemplateRun{}
		var pointsByRankJSON sql.NullString
		if err := rows.Scan(
			&run.ID,
			&run.SessionID,
			&run.TemplateKey,
			&run.Name,
			&run.CreatedBy,
			&run.CreatedAt,
			&run.UpdatedAt,
			&pointsByRankJSON,
		); err != nil {
			return nil, err
		}

		if pointsByRankJSON.Valid && pointsByRankJSON.String != "" {
			var pointsByRank interface{}
			if err := json.Unmarshal([]byte(pointsByRankJSON.String), &pointsByRank); err != nil {
				return nil, fmt.Errorf("failed to unmarshal points_by_rank: %w", err)
			}
			run.PointsByRank = pointsByRank
		}

		out = append(out, run)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *noonGameRepository) LinkTemplateRunMatch(runID int, matchID int, matchKey string) (*models.NoonGameTemplateRunMatch, error) {
	if runID == 0 || matchID == 0 {
		return nil, fmt.Errorf("run_id and match_id are required")
	}
	key := strings.TrimSpace(matchKey)
	if key == "" {
		return nil, fmt.Errorf("match_key is required")
	}

	res, err := r.db.Exec(`
		INSERT INTO noon_game_template_run_matches (run_id, match_id, match_key)
		VALUES (?, ?, ?)
	`, runID, matchID, key)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRow(`
		SELECT id, run_id, match_id, match_key
		FROM noon_game_template_run_matches
		WHERE id = ?
	`, int(id))
	out := &models.NoonGameTemplateRunMatch{}
	if err := row.Scan(&out.ID, &out.RunID, &out.MatchID, &out.MatchKey); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *noonGameRepository) ListTemplateRunMatches(runID int) ([]*models.NoonGameTemplateRunMatch, error) {
	rows, err := r.db.Query(`
		SELECT id, run_id, match_id, match_key
		FROM noon_game_template_run_matches
		WHERE run_id = ?
		ORDER BY id
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.NoonGameTemplateRunMatch
	for rows.Next() {
		item := &models.NoonGameTemplateRunMatch{}
		if err := rows.Scan(&item.ID, &item.RunID, &item.MatchID, &item.MatchKey); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *noonGameRepository) GetTemplateRunMatchByKey(runID int, matchKey string) (*models.NoonGameTemplateRunMatch, error) {
	key := strings.TrimSpace(matchKey)
	if key == "" {
		return nil, fmt.Errorf("match_key is required")
	}

	row := r.db.QueryRow(`
		SELECT id, run_id, match_id, match_key
		FROM noon_game_template_run_matches
		WHERE run_id = ? AND match_key = ?
	`, runID, key)
	item := &models.NoonGameTemplateRunMatch{}
	if err := row.Scan(&item.ID, &item.RunID, &item.MatchID, &item.MatchKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (r *noonGameRepository) GetTemplateRunMatchByMatchID(matchID int) (*models.NoonGameTemplateRunMatch, error) {
	row := r.db.QueryRow(`
		SELECT id, run_id, match_id, match_key
		FROM noon_game_template_run_matches
		WHERE match_id = ?
	`, matchID)
	item := &models.NoonGameTemplateRunMatch{}
	if err := row.Scan(&item.ID, &item.RunID, &item.MatchID, &item.MatchKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (r *noonGameRepository) DeleteTemplateRunAndRelatedData(sessionID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// セッションに関連するすべてのテンプレートランを取得
	runs, err := r.ListTemplateRunsBySession(sessionID)
	if err != nil {
		return err
	}

	for _, run := range runs {
		// テンプレートランに関連する試合を取得
		runMatches, err := r.ListTemplateRunMatches(run.ID)
		if err != nil {
			return err
		}

		// 試合を削除（関連データも含む）- トランザクション内で直接削除
		for _, runMatch := range runMatches {
			matchID := runMatch.MatchID
			// 試合に関連するポイントを削除
			if _, err := tx.Exec(`DELETE FROM noon_game_points WHERE match_id = ?`, matchID); err != nil {
				return err
			}
			// 試合結果の詳細を削除
			if _, err := tx.Exec(`
				DELETE FROM noon_game_result_details
				WHERE result_id IN (SELECT id FROM noon_game_results WHERE match_id = ?)
			`, matchID); err != nil {
				return err
			}
			// 試合結果を削除
			if _, err := tx.Exec(`DELETE FROM noon_game_results WHERE match_id = ?`, matchID); err != nil {
				return err
			}
			// 試合エントリーを削除
			if _, err := tx.Exec(`DELETE FROM noon_game_match_entries WHERE match_id = ?`, matchID); err != nil {
				return err
			}
			// 試合を削除
			if _, err := tx.Exec(`DELETE FROM noon_game_matches WHERE id = ? AND session_id = ?`, matchID, sessionID); err != nil {
				return err
			}
		}

		// テンプレートラン試合のリンクを削除
		if _, err := tx.Exec(`DELETE FROM noon_game_template_run_matches WHERE run_id = ?`, run.ID); err != nil {
			return err
		}

		// テンプレートランを削除
		if _, err := tx.Exec(`DELETE FROM noon_game_template_runs WHERE id = ?`, run.ID); err != nil {
			return err
		}
	}

	// テンプレートで作成されたグループを削除
	// テンプレート作成時に作成されたグループを削除する
	groups, err := r.GetGroupsWithMembers(sessionID)
	if err != nil {
		return err
	}
	for _, group := range groups {
		// グループメンバーを削除
		if _, err := tx.Exec(`DELETE FROM noon_game_group_members WHERE group_id = ?`, group.ID); err != nil {
			return err
		}
		// グループを削除
		if _, err := tx.Exec(`DELETE FROM noon_game_groups WHERE id = ? AND session_id = ?`, group.ID, sessionID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *noonGameRepository) GetSessionByID(sessionID int) (*models.NoonGameSession, error) {
	row := r.db.QueryRow(`
		SELECT id, event_id, name, description, mode, win_points, loss_points, draw_points,
		       participation_points, allow_manual_points, created_at, updated_at
		FROM noon_game_sessions
		WHERE id = ?
	`, sessionID)

	session := &models.NoonGameSession{}
	var description sql.NullString
	if err := row.Scan(
		&session.ID,
		&session.EventID,
		&session.Name,
		&description,
		&session.Mode,
		&session.WinPoints,
		&session.LossPoints,
		&session.DrawPoints,
		&session.ParticipationPoints,
		&session.AllowManualPoints,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if description.Valid {
		session.Description = &description.String
	}
	return session, nil
}

func (r *noonGameRepository) GetSessionByEvent(eventID int) (*models.NoonGameSession, error) {
	row := r.db.QueryRow(`
		SELECT id, event_id, name, description, mode, win_points, loss_points, draw_points,
		       participation_points, allow_manual_points, created_at, updated_at
		FROM noon_game_sessions
		WHERE event_id = ?
	`, eventID)

	session := &models.NoonGameSession{}
	var description sql.NullString
	if err := row.Scan(
		&session.ID,
		&session.EventID,
		&session.Name,
		&description,
		&session.Mode,
		&session.WinPoints,
		&session.LossPoints,
		&session.DrawPoints,
		&session.ParticipationPoints,
		&session.AllowManualPoints,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if description.Valid {
		session.Description = &description.String
	}
	return session, nil
}

func (r *noonGameRepository) UpsertSession(session *models.NoonGameSession) (*models.NoonGameSession, error) {
	result, err := r.db.Exec(`
		INSERT INTO noon_game_sessions (
			event_id, name, description, mode, win_points, loss_points, draw_points,
			participation_points, allow_manual_points
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			description = VALUES(description),
			mode = VALUES(mode),
			win_points = VALUES(win_points),
			loss_points = VALUES(loss_points),
			draw_points = VALUES(draw_points),
			participation_points = VALUES(participation_points),
			allow_manual_points = VALUES(allow_manual_points),
			updated_at = CURRENT_TIMESTAMP
	`, session.EventID, session.Name, nullableString(session.Description), session.Mode, session.WinPoints, session.LossPoints, session.DrawPoints, session.ParticipationPoints, session.AllowManualPoints)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err == nil && id > 0 {
		session.ID = int(id)
	} else {
		existing, err := r.GetSessionByEvent(session.EventID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			session.ID = existing.ID
		}
	}

	return r.GetSessionByEvent(session.EventID)
}

func (r *noonGameRepository) GetGroupsWithMembers(sessionID int) ([]*models.NoonGameGroupWithMembers, error) {
	rows, err := r.db.Query(`
		SELECT id, session_id, name, description, created_at, updated_at
		FROM noon_game_groups
		WHERE session_id = ?
		ORDER BY id
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*models.NoonGameGroupWithMembers
	for rows.Next() {
		group := &models.NoonGameGroup{}
		var description sql.NullString
		if err := rows.Scan(
			&group.ID,
			&group.SessionID,
			&group.Name,
			&description,
			&group.CreatedAt,
			&group.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if description.Valid {
			group.Description = &description.String
		}

		members, err := r.GetGroupMembers(group.ID)
		if err != nil {
			return nil, err
		}

		groups = append(groups, &models.NoonGameGroupWithMembers{
			NoonGameGroup: group,
			Members:       members,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (r *noonGameRepository) GetGroupMembers(groupID int) ([]*models.NoonGameGroupMember, error) {
	rows, err := r.db.Query(`
		SELECT gm.id, gm.group_id, gm.class_id, gm.weight,
		       c.id, c.event_id, c.name, c.student_count, c.attend_count
		FROM noon_game_group_members gm
		JOIN classes c ON gm.class_id = c.id
		WHERE gm.group_id = ?
		ORDER BY c.name
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.NoonGameGroupMember
	for rows.Next() {
		member := &models.NoonGameGroupMember{
			Class: &models.Class{},
		}
		if err := rows.Scan(
			&member.ID,
			&member.GroupID,
			&member.ClassID,
			&member.Weight,
			&member.Class.ID,
			&member.Class.EventID,
			&member.Class.Name,
			&member.Class.StudentCount,
			&member.Class.AttendCount,
		); err != nil {
			return nil, err
		}

		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

func (r *noonGameRepository) GetGroupWithMembers(sessionID int, groupID int) (*models.NoonGameGroupWithMembers, error) {
	row := r.db.QueryRow(`
		SELECT id, session_id, name, description, created_at, updated_at
		FROM noon_game_groups
		WHERE id = ? AND session_id = ?
	`, groupID, sessionID)

	group := &models.NoonGameGroup{}
	var description sql.NullString
	if err := row.Scan(
		&group.ID,
		&group.SessionID,
		&group.Name,
		&description,
		&group.CreatedAt,
		&group.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if description.Valid {
		group.Description = &description.String
	}

	members, err := r.GetGroupMembers(group.ID)
	if err != nil {
		return nil, err
	}

	return &models.NoonGameGroupWithMembers{
		NoonGameGroup: group,
		Members:       members,
	}, nil
}

func (r *noonGameRepository) SaveGroup(group *models.NoonGameGroup, memberClassIDs []int) (*models.NoonGameGroupWithMembers, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()
	if group.ID == 0 {
		result, err := tx.Exec(`
			INSERT INTO noon_game_groups (session_id, name, description, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)
		`, group.SessionID, group.Name, nullableString(group.Description), now, now)
		if err != nil {
			return nil, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		group.ID = int(id)
		group.CreatedAt = now
		group.UpdatedAt = now
	} else {
		result, err := tx.Exec(`
			UPDATE noon_game_groups
			SET name = ?, description = ?, updated_at = ?
			WHERE id = ? AND session_id = ?
		`, group.Name, nullableString(group.Description), now, group.ID, group.SessionID)
		if err != nil {
			return nil, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}
		if affected == 0 {
			return nil, fmt.Errorf("noon game group not found")
		}
		group.UpdatedAt = now
	}

	if _, err := tx.Exec(`DELETE FROM noon_game_group_members WHERE group_id = ?`, group.ID); err != nil {
		return nil, err
	}

	if len(memberClassIDs) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO noon_game_group_members (group_id, class_id, weight)
			VALUES (?, ?, 1.0)
		`)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		for _, classID := range memberClassIDs {
			if _, err := stmt.Exec(group.ID, classID); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetGroupWithMembers(group.SessionID, group.ID)
}

func (r *noonGameRepository) DeleteGroup(sessionID int, groupID int) error {
	var count int
	if err := r.db.QueryRow(`
		SELECT COUNT(*) FROM noon_game_matches
		WHERE session_id = ? AND (home_group_id = ? OR away_group_id = ?)
	`, sessionID, groupID, groupID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("group is referenced by existing matches and cannot be deleted")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM noon_game_group_members WHERE group_id = ?`, groupID); err != nil {
		return err
	}

	result, err := tx.Exec(`DELETE FROM noon_game_groups WHERE id = ? AND session_id = ?`, groupID, sessionID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("noon game group not found")
	}

	return tx.Commit()
}

func (r *noonGameRepository) GetMatchesWithResults(sessionID int) ([]*models.NoonGameMatchWithResult, error) {
	rows, err := r.db.Query(`
		SELECT
			m.id, m.session_id, m.title, m.scheduled_at, m.location, m.format,
			m.status, m.memo, m.home_side_type, m.home_class_id, m.home_group_id,
			m.away_side_type, m.away_class_id, m.away_group_id, m.allow_draw,
			m.created_at, m.updated_at,
			r.id, r.winner, r.recorded_by, r.recorded_at, r.note
		FROM noon_game_matches m
		LEFT JOIN noon_game_results r ON r.match_id = m.id
		WHERE m.session_id = ?
		ORDER BY
			CASE WHEN m.scheduled_at IS NULL THEN 1 ELSE 0 END,
			m.scheduled_at,
			m.id
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		matches  []*models.NoonGameMatchWithResult
		matchIDs []int
	)

	for rows.Next() {
		match := &models.NoonGameMatch{}
		var (
			title       sql.NullString
			scheduledAt sql.NullTime
			location    sql.NullString
			format      sql.NullString
			memo        sql.NullString
			homeClassID sql.NullInt64
			homeGroupID sql.NullInt64
			awayClassID sql.NullInt64
			awayGroupID sql.NullInt64
		)

		result := &models.NoonGameResult{}
		var (
			resultID   sql.NullInt64
			winner     sql.NullString
			recordedBy sql.NullString
			recordedAt sql.NullTime
			resultNote sql.NullString
		)

		if err := rows.Scan(
			&match.ID,
			&match.SessionID,
			&title,
			&scheduledAt,
			&location,
			&format,
			&match.Status,
			&memo,
			&match.HomeSideType,
			&homeClassID,
			&homeGroupID,
			&match.AwaySideType,
			&awayClassID,
			&awayGroupID,
			&match.AllowDraw,
			&match.CreatedAt,
			&match.UpdatedAt,
			&resultID,
			&winner,
			&recordedBy,
			&recordedAt,
			&resultNote,
		); err != nil {
			return nil, err
		}

		if title.Valid {
			match.Title = &title.String
		}
		if scheduledAt.Valid {
			t := scheduledAt.Time
			match.ScheduledAt = &t
		}
		if location.Valid {
			match.Location = &location.String
		}
		if format.Valid {
			match.Format = &format.String
		}
		if memo.Valid {
			match.Memo = &memo.String
		}
		if homeClassID.Valid {
			v := int(homeClassID.Int64)
			match.HomeClassID = &v
		}
		if homeGroupID.Valid {
			v := int(homeGroupID.Int64)
			match.HomeGroupID = &v
		}
		if awayClassID.Valid {
			v := int(awayClassID.Int64)
			match.AwayClassID = &v
		}
		if awayGroupID.Valid {
			v := int(awayGroupID.Int64)
			match.AwayGroupID = &v
		}

		var matchResult *models.NoonGameResult
		if resultID.Valid {
			matchResult = result
			result.ID = int(resultID.Int64)
			if winner.Valid {
				result.Winner = winner.String
			}
			if recordedBy.Valid {
				result.RecordedBy = recordedBy.String
			}
			if recordedAt.Valid {
				result.RecordedAt = recordedAt.Time
			}
			if resultNote.Valid {
				result.Note = &resultNote.String
			}
		}

		matches = append(matches, &models.NoonGameMatchWithResult{
			NoonGameMatch: match,
			Result:        matchResult,
		})
		matchIDs = append(matchIDs, match.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(matchIDs) == 0 {
		return matches, nil
	}

	entryMap, err := r.fetchMatchEntries(matchIDs)
	if err != nil {
		return nil, err
	}
	detailMap, err := r.fetchResultDetails(matchIDs)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		if entries, ok := entryMap[match.ID]; ok {
			match.Entries = entries
			match.NoonGameMatch.Entries = entries
		} else {
			match.Entries = []*models.NoonGameMatchEntry{}
			match.NoonGameMatch.Entries = []*models.NoonGameMatchEntry{}
		}
		if match.Result != nil {
			if details, ok := detailMap[match.ID]; ok {
				match.Result.Details = details
			} else {
				match.Result.Details = []*models.NoonGameResultDetail{}
			}
		}
	}

	return matches, nil
}

func (r *noonGameRepository) GetMatchByID(matchID int) (*models.NoonGameMatchWithResult, error) {
	row := r.db.QueryRow(`
		SELECT
			m.id, m.session_id, m.title, m.scheduled_at, m.location, m.format,
			m.status, m.memo, m.home_side_type, m.home_class_id, m.home_group_id,
			m.away_side_type, m.away_class_id, m.away_group_id, m.allow_draw,
			m.created_at, m.updated_at,
			r.id, r.winner, r.recorded_by, r.recorded_at, r.note
		FROM noon_game_matches m
		LEFT JOIN noon_game_results r ON r.match_id = m.id
		WHERE m.id = ?
	`, matchID)

	match := &models.NoonGameMatch{}
	var (
		title       sql.NullString
		scheduledAt sql.NullTime
		location    sql.NullString
		format      sql.NullString
		memo        sql.NullString
		homeClassID sql.NullInt64
		homeGroupID sql.NullInt64
		awayClassID sql.NullInt64
		awayGroupID sql.NullInt64
	)
	result := &models.NoonGameResult{}
	var (
		resultID   sql.NullInt64
		winner     sql.NullString
		recordedBy sql.NullString
		recordedAt sql.NullTime
		note       sql.NullString
	)

	if err := row.Scan(
		&match.ID,
		&match.SessionID,
		&title,
		&scheduledAt,
		&location,
		&format,
		&match.Status,
		&memo,
		&match.HomeSideType,
		&homeClassID,
		&homeGroupID,
		&match.AwaySideType,
		&awayClassID,
		&awayGroupID,
		&match.AllowDraw,
		&match.CreatedAt,
		&match.UpdatedAt,
		&resultID,
		&winner,
		&recordedBy,
		&recordedAt,
		&note,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if title.Valid {
		match.Title = &title.String
	}
	if scheduledAt.Valid {
		t := scheduledAt.Time
		match.ScheduledAt = &t
	}
	if location.Valid {
		match.Location = &location.String
	}
	if format.Valid {
		match.Format = &format.String
	}
	if memo.Valid {
		match.Memo = &memo.String
	}
	if homeClassID.Valid {
		v := int(homeClassID.Int64)
		match.HomeClassID = &v
	}
	if homeGroupID.Valid {
		v := int(homeGroupID.Int64)
		match.HomeGroupID = &v
	}
	if awayClassID.Valid {
		v := int(awayClassID.Int64)
		match.AwayClassID = &v
	}
	if awayGroupID.Valid {
		v := int(awayGroupID.Int64)
		match.AwayGroupID = &v
	}

	var matchResult *models.NoonGameResult
	if resultID.Valid {
		matchResult = result
		result.ID = int(resultID.Int64)
		if winner.Valid {
			result.Winner = winner.String
		}
		if recordedBy.Valid {
			result.RecordedBy = recordedBy.String
		}
		if recordedAt.Valid {
			result.RecordedAt = recordedAt.Time
		}
		if note.Valid {
			result.Note = &note.String
		}
	}

	matchWithResult := &models.NoonGameMatchWithResult{
		NoonGameMatch: match,
		Result:        matchResult,
	}

	entryMap, err := r.fetchMatchEntries([]int{match.ID})
	if err != nil {
		return nil, err
	}
	if entries, ok := entryMap[match.ID]; ok {
		matchWithResult.Entries = entries
		matchWithResult.NoonGameMatch.Entries = entries
	} else {
		matchWithResult.Entries = []*models.NoonGameMatchEntry{}
		matchWithResult.NoonGameMatch.Entries = []*models.NoonGameMatchEntry{}
	}

	if matchWithResult.Result != nil {
		detailMap, err := r.fetchResultDetails([]int{match.ID})
		if err != nil {
			return nil, err
		}
		if details, ok := detailMap[match.ID]; ok {
			matchWithResult.Result.Details = details
		} else {
			matchWithResult.Result.Details = []*models.NoonGameResultDetail{}
		}
	}

	return matchWithResult, nil
}

func (r *noonGameRepository) SaveMatch(match *models.NoonGameMatch) (*models.NoonGameMatch, error) {
	if match == nil {
		return nil, fmt.Errorf("match is nil")
	}

	assignPrimarySidesFromEntries(match)

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()

	if match.ID == 0 {
		titleVal := nullableString(match.Title)
		scheduledVal := nullableTime(match.ScheduledAt)
		locationVal := nullableString(match.Location)
		formatVal := nullableString(match.Format)
		memoVal := nullableString(match.Memo)
		homeClassVal := nullableInt(match.HomeClassID)
		homeGroupVal := nullableInt(match.HomeGroupID)
		awayClassVal := nullableInt(match.AwayClassID)
		awayGroupVal := nullableInt(match.AwayGroupID)

		result, err := tx.Exec(`
			INSERT INTO noon_game_matches (
				session_id, title, scheduled_at, location, format, status, memo,
				home_side_type, home_class_id, home_group_id,
				away_side_type, away_class_id, away_group_id,
				allow_draw, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			match.SessionID,
			titleVal,
			scheduledVal,
			locationVal,
			formatVal,
			match.Status,
			memoVal,
			match.HomeSideType,
			homeClassVal,
			homeGroupVal,
			match.AwaySideType,
			awayClassVal,
			awayGroupVal,
			match.AllowDraw,
			now,
			now,
		)
		if err != nil {
			return nil, err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		match.ID = int(id)
		match.CreatedAt = now
		match.UpdatedAt = now
	} else {
		titleVal := nullableString(match.Title)
		scheduledVal := nullableTime(match.ScheduledAt)
		locationVal := nullableString(match.Location)
		formatVal := nullableString(match.Format)
		memoVal := nullableString(match.Memo)
		homeClassVal := nullableInt(match.HomeClassID)
		homeGroupVal := nullableInt(match.HomeGroupID)
		awayClassVal := nullableInt(match.AwayClassID)
		awayGroupVal := nullableInt(match.AwayGroupID)

		result, err := tx.Exec(`
			UPDATE noon_game_matches
			SET title = ?, scheduled_at = ?, location = ?, format = ?, status = ?, memo = ?,
				home_side_type = ?, home_class_id = ?, home_group_id = ?,
				away_side_type = ?, away_class_id = ?, away_group_id = ?,
				allow_draw = ?, updated_at = ?
			WHERE id = ? AND session_id = ?
		`,
			titleVal,
			scheduledVal,
			locationVal,
			formatVal,
			match.Status,
			memoVal,
			match.HomeSideType,
			homeClassVal,
			homeGroupVal,
			match.AwaySideType,
			awayClassVal,
			awayGroupVal,
			match.AllowDraw,
			now,
			match.ID,
			match.SessionID,
		)
		if err != nil {
			return nil, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}
		if affected == 0 {
			return nil, fmt.Errorf("noon game match not found")
		}
		match.UpdatedAt = now
	}

	if match.Entries != nil {
		if err := r.replaceMatchEntriesTx(tx, match.ID, match.Entries); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetMatchEntity(match.ID)
}

func (r *noonGameRepository) GetMatchEntity(matchID int) (*models.NoonGameMatch, error) {
	row := r.db.QueryRow(`
		SELECT
			id, session_id, title, scheduled_at, location, format,
			status, memo, home_side_type, home_class_id, home_group_id,
			away_side_type, away_class_id, away_group_id, allow_draw,
			created_at, updated_at
		FROM noon_game_matches
		WHERE id = ?
	`, matchID)

	match := &models.NoonGameMatch{}
	var (
		title       sql.NullString
		scheduledAt sql.NullTime
		location    sql.NullString
		format      sql.NullString
		memo        sql.NullString
		homeClassID sql.NullInt64
		homeGroupID sql.NullInt64
		awayClassID sql.NullInt64
		awayGroupID sql.NullInt64
	)

	if err := row.Scan(
		&match.ID,
		&match.SessionID,
		&title,
		&scheduledAt,
		&location,
		&format,
		&match.Status,
		&memo,
		&match.HomeSideType,
		&homeClassID,
		&homeGroupID,
		&match.AwaySideType,
		&awayClassID,
		&awayGroupID,
		&match.AllowDraw,
		&match.CreatedAt,
		&match.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if title.Valid {
		match.Title = &title.String
	}
	if scheduledAt.Valid {
		t := scheduledAt.Time
		match.ScheduledAt = &t
	}
	if location.Valid {
		match.Location = &location.String
	}
	if format.Valid {
		match.Format = &format.String
	}
	if memo.Valid {
		match.Memo = &memo.String
	}
	if homeClassID.Valid {
		v := int(homeClassID.Int64)
		match.HomeClassID = &v
	}
	if homeGroupID.Valid {
		v := int(homeGroupID.Int64)
		match.HomeGroupID = &v
	}
	if awayClassID.Valid {
		v := int(awayClassID.Int64)
		match.AwayClassID = &v
	}
	if awayGroupID.Valid {
		v := int(awayGroupID.Int64)
		match.AwayGroupID = &v
	}

	entryMap, err := r.fetchMatchEntries([]int{match.ID})
	if err != nil {
		return nil, err
	}
	if entries, ok := entryMap[match.ID]; ok {
		match.Entries = entries
	} else {
		match.Entries = []*models.NoonGameMatchEntry{}
	}

	return match, nil
}

func (r *noonGameRepository) DeleteMatch(sessionID int, matchID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM noon_game_points WHERE match_id = ?`, matchID); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		DELETE FROM noon_game_result_details
		WHERE result_id IN (SELECT id FROM noon_game_results WHERE match_id = ?)
	`, matchID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM noon_game_results WHERE match_id = ?`, matchID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM noon_game_match_entries WHERE match_id = ?`, matchID); err != nil {
		return err
	}

	result, err := tx.Exec(`DELETE FROM noon_game_matches WHERE id = ? AND session_id = ?`, matchID, sessionID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("noon game match not found")
	}

	return tx.Commit()
}

func (r *noonGameRepository) SaveResult(result *models.NoonGameResult) (*models.NoonGameResult, error) {
	if result == nil {
		return nil, fmt.Errorf("result is nil")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		INSERT INTO noon_game_results (match_id, winner, recorded_by, recorded_at, note)
		VALUES (?, ?, ?, NOW(), ?)
		ON DUPLICATE KEY UPDATE
			winner = VALUES(winner),
			recorded_by = VALUES(recorded_by),
			recorded_at = NOW(),
			note = VALUES(note)
	`,
		result.MatchID,
		result.Winner,
		result.RecordedBy,
		result.Note,
	)
	if err != nil {
		return nil, err
	}

	var resultID int
	if id, err := res.LastInsertId(); err == nil && id > 0 {
		resultID = int(id)
	} else {
		existing, err := r.GetResultByMatchID(result.MatchID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			resultID = existing.ID
		}
	}

	if resultID == 0 {
		return nil, fmt.Errorf("failed to determine result id")
	}
	result.ID = resultID

	if result.Details != nil {
		if err := r.replaceResultDetailsTx(tx, resultID, result.Details); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetResultByMatchID(result.MatchID)
}

func (r *noonGameRepository) GetResultByMatchID(matchID int) (*models.NoonGameResult, error) {
	row := r.db.QueryRow(`
		SELECT id, match_id, winner, recorded_by, recorded_at, note
		FROM noon_game_results
		WHERE match_id = ?
	`, matchID)

	result := &models.NoonGameResult{}
	var note sql.NullString

	if err := row.Scan(
		&result.ID,
		&result.MatchID,
		&result.Winner,
		&result.RecordedBy,
		&result.RecordedAt,
		&note,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if note.Valid {
		result.Note = &note.String
	}

	detailMap, err := r.fetchResultDetails([]int{matchID})
	if err != nil {
		return nil, err
	}
	if details, ok := detailMap[matchID]; ok {
		result.Details = details
	} else {
		result.Details = []*models.NoonGameResultDetail{}
	}

	return result, nil
}

func (r *noonGameRepository) ClearPointsForMatch(matchID int) error {
	_, err := r.db.Exec(`DELETE FROM noon_game_points WHERE match_id = ?`, matchID)
	return err
}

func (r *noonGameRepository) InsertPoints(points []*models.NoonGamePoint) error {
	if len(points) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO noon_game_points (session_id, match_id, class_id, points, reason, source, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, point := range points {
		matchIDVal := nullableInt(point.MatchID)
		reasonVal := nullableString(point.Reason)

		if _, err := stmt.Exec(
			point.SessionID,
			matchIDVal,
			point.ClassID,
			point.Points,
			reasonVal,
			point.Source,
			point.CreatedBy,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *noonGameRepository) InsertPoint(point *models.NoonGamePoint) (*models.NoonGamePoint, error) {
	matchIDVal := nullableInt(point.MatchID)
	reasonVal := nullableString(point.Reason)
	result, err := r.db.Exec(`
		INSERT INTO noon_game_points (session_id, match_id, class_id, points, reason, source, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		point.SessionID,
		matchIDVal,
		point.ClassID,
		point.Points,
		reasonVal,
		point.Source,
		point.CreatedBy,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err == nil {
		point.ID = int(id)
	}
	point.CreatedAt = time.Now()
	return point, nil
}

func (r *noonGameRepository) fetchMatchEntries(matchIDs []int) (map[int][]*models.NoonGameMatchEntry, error) {
	result := make(map[int][]*models.NoonGameMatchEntry)
	if len(matchIDs) == 0 {
		return result, nil
	}

	placeholder, args := buildInClause(matchIDs)
	query := fmt.Sprintf(`
		SELECT id, match_id, entry_index, side_type, class_id, group_id, display_name
		FROM noon_game_match_entries
		WHERE match_id IN (%s)
		ORDER BY match_id, entry_index
	`, placeholder)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		entry := &models.NoonGameMatchEntry{}
		var (
			classID     sql.NullInt64
			groupID     sql.NullInt64
			displayName sql.NullString
		)
		if err := rows.Scan(
			&entry.ID,
			&entry.MatchID,
			&entry.EntryIndex,
			&entry.SideType,
			&classID,
			&groupID,
			&displayName,
		); err != nil {
			return nil, err
		}

		if classID.Valid {
			val := int(classID.Int64)
			entry.ClassID = &val
		}
		if groupID.Valid {
			val := int(groupID.Int64)
			entry.GroupID = &val
		}
		if displayName.Valid {
			str := displayName.String
			entry.DisplayName = &str
		}

		result[entry.MatchID] = append(result[entry.MatchID], entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *noonGameRepository) fetchResultDetails(matchIDs []int) (map[int][]*models.NoonGameResultDetail, error) {
	result := make(map[int][]*models.NoonGameResultDetail)
	if len(matchIDs) == 0 {
		return result, nil
	}

	placeholder, args := buildInClause(matchIDs)
	query := fmt.Sprintf(`
		SELECT rd.id, r.match_id, rd.entry_id, rd.placement_rank, rd.points, rd.note
		FROM noon_game_result_details rd
		JOIN noon_game_results r ON rd.result_id = r.id
		WHERE r.match_id IN (%s)
		ORDER BY r.match_id, rd.placement_rank, rd.id
	`, placeholder)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		detail := &models.NoonGameResultDetail{}
		var (
			matchID sql.NullInt64
			rank    sql.NullInt64
			note    sql.NullString
		)
		if err := rows.Scan(
			&detail.ID,
			&matchID,
			&detail.EntryID,
			&rank,
			&detail.Points,
			&note,
		); err != nil {
			return nil, err
		}

		if !matchID.Valid {
			continue
		}

		if rank.Valid {
			val := int(rank.Int64)
			detail.Rank = &val
		}
		if note.Valid {
			str := note.String
			detail.Note = &str
		}

		result[int(matchID.Int64)] = append(result[int(matchID.Int64)], detail)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func buildInClause(ids []int) (string, []interface{}) {
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	return strings.Join(placeholders, ","), args
}

func (r *noonGameRepository) replaceMatchEntriesTx(tx *sql.Tx, matchID int, entries []*models.NoonGameMatchEntry) error {
	if entries == nil {
		return nil
	}

	if _, err := tx.Exec(`DELETE FROM noon_game_match_entries WHERE match_id = ?`, matchID); err != nil {
		return err
	}

	if len(entries) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(`
		INSERT INTO noon_game_match_entries (match_id, entry_index, side_type, class_id, group_id, display_name)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for index, entry := range entries {
		if entry == nil {
			continue
		}
		sideType := entry.SideType
		if sideType == "" {
			sideType = "class"
		}

		classVal := nullableInt(entry.ClassID)
		groupVal := nullableInt(entry.GroupID)
		displayVal := nullableString(entry.DisplayName)

		if sideType == "class" {
			groupVal = nil
		} else if sideType == "group" {
			classVal = nil
		}

		if _, err := stmt.Exec(matchID, index, sideType, classVal, groupVal, displayVal); err != nil {
			return err
		}

		entry.MatchID = matchID
		entry.EntryIndex = index
		entry.SideType = sideType
		if classVal == nil {
			entry.ClassID = nil
		}
		if groupVal == nil {
			entry.GroupID = nil
		}
	}

	return nil
}

func (r *noonGameRepository) replaceResultDetailsTx(tx *sql.Tx, resultID int, details []*models.NoonGameResultDetail) error {
	if details == nil {
		return nil
	}

	if _, err := tx.Exec(`DELETE FROM noon_game_result_details WHERE result_id = ?`, resultID); err != nil {
		return err
	}

	if len(details) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(`
		INSERT INTO noon_game_result_details (result_id, entry_id, placement_rank, points, note)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, detail := range details {
		if detail == nil {
			continue
		}
		if _, err := stmt.Exec(
			resultID,
			detail.EntryID,
			nullableInt(detail.Rank),
			detail.Points,
			nullableString(detail.Note),
		); err != nil {
			return err
		}
	}

	return nil
}

func assignPrimarySidesFromEntries(match *models.NoonGameMatch) {
	if match == nil || len(match.Entries) == 0 {
		return
	}

	first := match.Entries[0]
	if first != nil && first.SideType != "" {
		match.HomeSideType = first.SideType
		if first.SideType == "class" {
			match.HomeClassID = cloneIntPointer(first.ClassID)
			match.HomeGroupID = nil
		} else {
			match.HomeGroupID = cloneIntPointer(first.GroupID)
			match.HomeClassID = nil
		}
	}

	if len(match.Entries) > 1 {
		second := match.Entries[1]
		if second != nil && second.SideType != "" {
			match.AwaySideType = second.SideType
			if second.SideType == "class" {
				match.AwayClassID = cloneIntPointer(second.ClassID)
				match.AwayGroupID = nil
			} else {
				match.AwayGroupID = cloneIntPointer(second.GroupID)
				match.AwayClassID = nil
			}
		}
	} else if match.AwaySideType == "" {
		match.AwaySideType = match.HomeSideType
		match.AwayClassID = nil
		match.AwayGroupID = nil
	}

	if match.HomeSideType == "" {
		match.HomeSideType = "class"
	}
	if match.AwaySideType == "" {
		match.AwaySideType = match.HomeSideType
	}
}

func cloneIntPointer(src *int) *int {
	if src == nil {
		return nil
	}
	val := *src
	return &val
}

func (r *noonGameRepository) SumPointsByClass(sessionID int) (map[int]int, error) {
	rows, err := r.db.Query(`
		SELECT class_id, COALESCE(SUM(points), 0) AS total_points
		FROM noon_game_points
		WHERE session_id = ?
		GROUP BY class_id
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]int)
	for rows.Next() {
		var classID int
		var total int
		if err := rows.Scan(&classID, &total); err != nil {
			return nil, err
		}
		result[classID] = total
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func nullableString(ptr *string) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}

func nullableTime(ptr *time.Time) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}

func nullableInt(ptr *int) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}
