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

	// Insert into score_logs
	logReason := "MVP vote"
	if reason != "" {
		logReason = fmt.Sprintf("MVP vote: %s", reason)
	}
	_, err = tx.Exec("INSERT INTO score_logs (event_id, class_id, points, reason) VALUES (?, ?, ?, ?)", eventID, votedForClassID, mvpPoints, logReason)
	if err != nil {
		return err
	}

	return tx.Commit()
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
