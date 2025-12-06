package repository

import (
	"database/sql"
	"fmt"
)

type ClassScoreRepository interface {
	InitializeClassScores(eventID int, classIDs []int) error
}

type classScoreRepository struct {
	db *sql.DB
}

func NewClassScoreRepository(db *sql.DB) ClassScoreRepository {
	return &classScoreRepository{db: db}
}

func (r *classScoreRepository) InitializeClassScores(eventID int, classIDs []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO class_scores (event_id, class_id) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, classID := range classIDs {
		_, err := stmt.Exec(eventID, classID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Update ranks manually within the transaction to avoid trigger conflicts
	if err := r.updateClassRanksInTransaction(tx, eventID); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update ranks: %w", err)
	}

	return tx.Commit()
}

// updateClassRanksInTransaction updates class ranks within a transaction
func (r *classScoreRepository) updateClassRanksInTransaction(tx *sql.Tx, eventID int) error {
	// Check if all classes have 0 points (competition not started)
	var maxCurrentPoints int
	var maxOverallPoints int
	err := tx.QueryRow(`
		SELECT 
			COALESCE(MAX(total_points_current_event), 0),
			COALESCE(MAX(total_points_overall), 0)
		FROM class_scores
		WHERE event_id = ?
	`, eventID).Scan(&maxCurrentPoints, &maxOverallPoints)
	if err != nil {
		return fmt.Errorf("failed to check max points: %w", err)
	}

	// Update rank_current_event
	if maxCurrentPoints == 0 {
		// All classes have 0 points, set all ranks to 0
		_, err = tx.Exec(`
			UPDATE class_scores
			SET rank_current_event = 0
			WHERE event_id = ?
		`, eventID)
		if err != nil {
			return fmt.Errorf("failed to reset current event ranks: %w", err)
		}
	} else {
		// Normal ranking
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
		_, err = tx.Exec(updateCurrentRankQuery, eventID, eventID)
		if err != nil {
			return fmt.Errorf("failed to update current event ranks: %w", err)
		}
	}

	// Update rank_overall
	if maxOverallPoints == 0 {
		// All classes have 0 points, set all ranks to 0
		_, err = tx.Exec(`
			UPDATE class_scores
			SET rank_overall = 0
			WHERE event_id = ?
		`, eventID)
		if err != nil {
			return fmt.Errorf("failed to reset overall ranks: %w", err)
		}
	} else {
		// Normal ranking
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
	}

	return nil
}
