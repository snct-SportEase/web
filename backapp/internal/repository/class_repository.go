package repository

import (
	"backapp/internal/models"
	"database/sql"
)

type ClassRepository interface {
	GetAllClasses() ([]*models.Class, error)
}

type classRepository struct {
	db *sql.DB
}

func NewClassRepository(db *sql.DB) ClassRepository {
	return &classRepository{db: db}
}

func (r *classRepository) GetAllClasses() ([]*models.Class, error) {
	rows, err := r.db.Query("SELECT id, name FROM classes ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []*models.Class
	for rows.Next() {
		class := &models.Class{}
		if err := rows.Scan(&class.ID, &class.Name); err != nil {
			return nil, err
		}
		classes = append(classes, class)
	}

	return classes, nil
}
