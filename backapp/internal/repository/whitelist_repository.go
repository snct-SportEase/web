package repository

import (
	"database/sql"
)

type WhitelistRepository interface {
	IsEmailWhitelisted(email string) (bool, error)
	AddWhitelistedEmail(email, role string) error
}

type whitelistRepository struct {
	db *sql.DB
}

func NewWhitelistRepository(db *sql.DB) WhitelistRepository {
	return &whitelistRepository{db: db}
}

func (r *whitelistRepository) IsEmailWhitelisted(email string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM whitelisted_emails WHERE email = ?", email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *whitelistRepository) AddWhitelistedEmail(email, role string) error {
	_, err := r.db.Exec("INSERT INTO whitelisted_emails (email, role) VALUES (?, ?)", email, role)
	return err
}
