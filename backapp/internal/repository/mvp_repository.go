package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"backapp/internal/models"
)

type MVPRepository interface {
	GetEligibleClasses(eventID int) ([]models.Class, error)
	VoteMVP(userID string, votedForClassID int, eventID int, reason string) error
	GetMVPVotes(eventID int) ([]models.MVPVote, error)
	GetVoteByUserID(userID string, eventID int) (*models.MVPVote, error)
	GetMVPClass(eventID int) (*models.MVPResult, error)
}

type mvpRepository struct {
	db *sql.DB
}

func NewMVPRepository(db *sql.DB) MVPRepository {
	return &mvpRepository{db: db}
}

func (r *mvpRepository) GetEligibleClasses(eventID int) ([]models.Class, error) {
	rows, err := r.db.Query("SELECT id, name FROM classes WHERE event_id = ? AND name IN ('1-1', '1-2', '1-3', 'IS2', 'IT2', 'IE2')", eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []models.Class
	for rows.Next() {
		var class models.Class
		if err := rows.Scan(&class.ID, &class.Name); err != nil {
			return nil, err
		}
		classes = append(classes, class)
	}

	return classes, nil
}

func (r *mvpRepository) VoteMVP(userID string, votedForClassID int, eventID int, reason string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback on any error

	// Check if the user has already voted
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM mvp_votes WHERE voter_user_id = ? AND event_id = ?", userID, eventID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("user has already voted")
	}

	// Insert the vote
	mvpPoints := 3
	_, err = tx.Exec("INSERT INTO mvp_votes (voter_user_id, voted_for_class_id, event_id, reason, points) VALUES (?, ?, ?, ?, ?)", userID, votedForClassID, eventID, reason, mvpPoints)
	if err != nil {
		return err
	}

	// Update class_scores
	_, err = tx.Exec(`
    	INSERT INTO class_scores (event_id, class_id, mvp_points)
    	VALUES (?, ?, ?)
    	ON DUPLICATE KEY UPDATE mvp_points = mvp_points + VALUES(mvp_points)
	`, eventID, votedForClassID, mvpPoints)
	if err != nil {
		return err
	}

	// Ensure MVP points do not affect total points used for ranking
	_, err = tx.Exec(`
		UPDATE class_scores
		SET
			total_points_current_event = total_points_current_event - ?,
			total_points_overall = total_points_overall - ?
		WHERE event_id = ? AND class_id = ?
	`, mvpPoints, mvpPoints, eventID, votedForClassID)
	if err != nil {
		return err
	}

	// Insert into score_logs
	logReason := "MVP vote"
	if reason != "" {
		logReason = fmt.Sprintf("MVP vote: %s", reason)
	}
	_, err = tx.Exec("INSERT INTO score_logs (event_id, class_id, points, reason) VALUES (?, ?, ?, ?)", eventID, votedForClassID, mvpPoints, logReason)
	if err != nil {
		return err
	}

	// Get season to determine which rank to update
	var season string
	err = tx.QueryRow("SELECT season FROM events WHERE id = ?", eventID).Scan(&season)
	if err != nil {
		return err
	}

	if err := updateCurrentEventRanks(tx, eventID); err != nil {
		return err
	}

	if season == "autumn" {
		if err := updateOverallRanks(tx, eventID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func updateCurrentEventRanks(tx *sql.Tx, eventID int) error {
	// Check if all classes have 0 points (competition not started)
	var maxPoints int
	err := tx.QueryRow(`
		SELECT COALESCE(MAX(total_points_current_event), 0)
		FROM class_scores
		WHERE event_id = ?
	`, eventID).Scan(&maxPoints)
	if err != nil {
		return fmt.Errorf("failed to check max points: %w", err)
	}

	if maxPoints == 0 {
		// All classes have 0 points, set all ranks to 0
		_, err = tx.Exec(`
			UPDATE class_scores
			SET rank_current_event = 0
			WHERE event_id = ?
		`, eventID)
		if err != nil {
			return fmt.Errorf("failed to reset current event ranks: %w", err)
		}
		return nil
	}

	// Normal ranking
	const query = `
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

	if _, err := tx.Exec(query, eventID, eventID); err != nil {
		return fmt.Errorf("failed to update ranks: %w", err)
	}

	return nil
}

func updateOverallRanks(tx *sql.Tx, eventID int) error {
	// Check if all classes have 0 points (competition not started)
	var maxPoints int
	err := tx.QueryRow(`
		SELECT COALESCE(MAX(total_points_overall), 0)
		FROM class_scores
		WHERE event_id = ?
	`, eventID).Scan(&maxPoints)
	if err != nil {
		return fmt.Errorf("failed to check max points: %w", err)
	}

	if maxPoints == 0 {
		// All classes have 0 points, set all ranks to 0
		_, err = tx.Exec(`
			UPDATE class_scores
			SET rank_overall = 0
			WHERE event_id = ?
		`, eventID)
		if err != nil {
			return fmt.Errorf("failed to reset overall ranks: %w", err)
		}
		return nil
	}

	// Normal ranking
	const query = `
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

	if _, err := tx.Exec(query, eventID, eventID); err != nil {
		return fmt.Errorf("failed to update overall ranks: %w", err)
	}

	return nil
}

func (r *mvpRepository) GetMVPVotes(eventID int) ([]models.MVPVote, error) {
	rows, err := r.db.Query("SELECT voted_for_class_id, COUNT(*) as points FROM mvp_votes WHERE event_id = ? GROUP BY voted_for_class_id", eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []models.MVPVote
	for rows.Next() {
		var vote models.MVPVote
		if err := rows.Scan(&vote.VotedForClassID, &vote.Points); err != nil {
			return nil, err
		}
		votes = append(votes, vote)
	}

	return votes, nil
}

func (r *mvpRepository) GetVoteByUserID(userID string, eventID int) (*models.MVPVote, error) {
	var vote models.MVPVote
	err := r.db.QueryRow("SELECT id, event_id, voter_user_id, voted_for_class_id, reason, points, created_at FROM mvp_votes WHERE voter_user_id = ? AND event_id = ?", userID, eventID).Scan(&vote.ID, &vote.EventID, &vote.VoterUserID, &vote.VotedForClassID, &vote.Reason, &vote.Points, &vote.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No vote found, not an error
		}
		return nil, err
	}
	return &vote, nil
}

func (r *mvpRepository) GetMVPClass(eventID int) (*models.MVPResult, error) {
	var season string
	err := r.db.QueryRow("SELECT season FROM events WHERE id = ?", eventID).Scan(&season)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("event not found")
		}
		return nil, err
	}

	var query string
	query = `
		SELECT c.name, (cs.total_points_overall + cs.mvp_points) AS total_points
		FROM class_scores cs
		JOIN classes c ON cs.class_id = c.id
		WHERE cs.event_id = ?
			AND c.name IN ('1-1', '1-2', '1-3', 'IS2', 'IT2', 'IE2')
		ORDER BY total_points DESC
		LIMIT 1
	`

	var result models.MVPResult
	err = r.db.QueryRow(query, eventID).Scan(&result.ClassName, &result.TotalPoints)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No MVP class found yet
		}
		return nil, err
	}

	result.Season = season
	result.EventID = eventID

	return &result, nil
}
