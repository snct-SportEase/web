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
	// class_scores is a VIEW, so it initializes dynamically when a class is created.
	return nil
}
