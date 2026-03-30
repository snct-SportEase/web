package repository

import (
	"database/sql"
	"errors"

	"backapp/internal/models"
)

type MICRepository interface {
	GetEligibleClasses(eventID int) ([]models.Class, error)
	VoteMIC(userID string, votedForClassID int, eventID int, reason string) error
	GetMICVotes(eventID int) ([]models.MICVote, error)
	GetVoteByUserID(userID string, eventID int) (*models.MICVote, error)
	GetMICClass(eventID int) (*models.MICResult, error)
}

type micRepository struct {
	db *sql.DB
}

func NewMICRepository(db *sql.DB) MICRepository {
	return &micRepository{db: db}
}

func (r *micRepository) GetEligibleClasses(eventID int) ([]models.Class, error) {
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

func (r *micRepository) VoteMIC(userID string, votedForClassID int, eventID int, reason string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback on any error

	// 投票対象クラスがそのイベントの有効なMIC対象クラスか確認
	var eligibleCount int
	err = tx.QueryRow(
		"SELECT COUNT(*) FROM classes WHERE id = ? AND event_id = ? AND name IN ('1-1', '1-2', '1-3', 'IS2', 'IT2', 'IE2')",
		votedForClassID, eventID,
	).Scan(&eligibleCount)
	if err != nil {
		return err
	}
	if eligibleCount == 0 {
		return errors.New("voted class is not eligible for MIC")
	}

	// Check if the user has already voted
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM mic_votes WHERE voter_user_id = ? AND event_id = ?", userID, eventID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("user has already voted")
	}

	// Insert the vote
	micPoints := 3
	_, err = tx.Exec("INSERT INTO mic_votes (voter_user_id, voted_for_class_id, event_id, reason, points) VALUES (?, ?, ?, ?, ?)", userID, votedForClassID, eventID, reason, micPoints)
	if err != nil {
		return err
	}

	// Insert into score_logs
	_, err = tx.Exec("INSERT INTO score_logs (event_id, class_id, points, reason) VALUES (?, ?, ?, ?)", eventID, votedForClassID, micPoints, "mic_points")
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *micRepository) GetMICVotes(eventID int) ([]models.MICVote, error) {
	rows, err := r.db.Query("SELECT voted_for_class_id, COUNT(*) as points FROM mic_votes WHERE event_id = ? GROUP BY voted_for_class_id", eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []models.MICVote
	for rows.Next() {
		var vote models.MICVote
		if err := rows.Scan(&vote.VotedForClassID, &vote.Points); err != nil {
			return nil, err
		}
		votes = append(votes, vote)
	}

	return votes, nil
}

func (r *micRepository) GetVoteByUserID(userID string, eventID int) (*models.MICVote, error) {
	var vote models.MICVote
	err := r.db.QueryRow("SELECT id, event_id, voter_user_id, voted_for_class_id, reason, points, created_at FROM mic_votes WHERE voter_user_id = ? AND event_id = ?", userID, eventID).Scan(&vote.ID, &vote.EventID, &vote.VoterUserID, &vote.VotedForClassID, &vote.Reason, &vote.Points, &vote.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No vote found, not an error
		}
		return nil, err
	}
	return &vote, nil
}

func (r *micRepository) GetMICClass(eventID int) (*models.MICResult, error) {
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
		SELECT c.name, (cs.total_points_overall + cs.mic_points) AS total_points
		FROM class_scores cs
		JOIN classes c ON cs.class_id = c.id
		WHERE cs.event_id = ?
			AND c.name IN ('1-1', '1-2', '1-3', 'IS2', 'IT2', 'IE2')
		ORDER BY total_points DESC
		LIMIT 1
	`

	var result models.MICResult
	err = r.db.QueryRow(query, eventID).Scan(&result.ClassName, &result.TotalPoints)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No MIC class found yet
		}
		return nil, err
	}

	result.Season = season
	result.EventID = eventID

	return &result, nil
}
