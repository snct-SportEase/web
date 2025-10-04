package repository

import (
	"database/sql"
)

type WhitelistEntry struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type WhitelistRepository interface {
	IsEmailWhitelisted(email string) (bool, error)
	AddWhitelistedEmail(email, role string) error
	GetAllWhitelistedEmails() ([]WhitelistEntry, error)
	AddWhitelistedEmails(entries []WhitelistEntry) error
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
	_, err := r.db.Exec("INSERT INTO whitelisted_emails (email, role) VALUES (?, ?) ON DUPLICATE KEY UPDATE role = VALUES(role)", email, role)
	return err
}

func (r *whitelistRepository) GetAllWhitelistedEmails() ([]WhitelistEntry, error) {
	rows, err := r.db.Query("SELECT email, role FROM whitelisted_emails ORDER BY email")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []WhitelistEntry
	for rows.Next() {
		var entry WhitelistEntry
		if err := rows.Scan(&entry.Email, &entry.Role); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (r *whitelistRepository) AddWhitelistedEmails(entries []WhitelistEntry) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO whitelisted_emails (email, role) VALUES (?, ?) ON DUPLICATE KEY UPDATE role = VALUES(role)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, entry := range entries {
		_, err := stmt.Exec(entry.Email, entry.Role)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
