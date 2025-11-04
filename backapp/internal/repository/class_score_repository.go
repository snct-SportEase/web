package repository

import (
	"database/sql"
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

	return tx.Commit()
}
