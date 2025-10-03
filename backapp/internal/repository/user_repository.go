package repository

import (
	"backapp/internal/models"
	"database/sql"
)

type UserRepository interface {
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)
	UpdateUser(user *models.User) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	row := r.db.QueryRow("SELECT id, email, display_name, class_id, is_profile_complete, created_at, updated_at FROM users WHERE email = ?", email)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Email, &user.DisplayName, &user.ClassID, &user.IsProfileComplete, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err
	}
	return user, nil
}

func (r *userRepository) CreateUser(user *models.User) error {
	_, err := r.db.Exec("INSERT INTO users (id, email, display_name, class_id, is_profile_complete) VALUES (?, ?, ?, ?, ?)",
		user.ID, user.Email, user.DisplayName, user.ClassID, user.IsProfileComplete)
	return err
}

func (r *userRepository) GetUserByID(id string) (*models.User, error) {
	row := r.db.QueryRow("SELECT id, email, display_name, class_id, is_profile_complete, created_at, updated_at FROM users WHERE id = ?", id)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Email, &user.DisplayName, &user.ClassID, &user.IsProfileComplete, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) UpdateUser(user *models.User) error {
	_, err := r.db.Exec("UPDATE users SET display_name = ?, class_id = ?, is_profile_complete = ? WHERE id = ?",
		user.DisplayName, user.ClassID, user.IsProfileComplete, user.ID)
	return err
}
