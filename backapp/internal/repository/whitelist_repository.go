package repository

import (
	"database/sql"
)

type WhitelistEntry struct {
	Email   string `json:"email"`
	Role    string `json:"role"`
	EventID *int   `json:"event_id"`
}

type WhitelistRepository interface {
	IsEmailWhitelisted(email string) (bool, error)
	AddWhitelistedEmail(email, role string, eventID *int) error
	GetAllWhitelistedEmails() ([]WhitelistEntry, error)
	AddWhitelistedEmails(entries []WhitelistEntry) error
	UpdateNullEventIDs(eventID int) error
}

type whitelistRepository struct {
	db *sql.DB
}

func NewWhitelistRepository(db *sql.DB) WhitelistRepository {
	return &whitelistRepository{db: db}
}

func (r *whitelistRepository) IsEmailWhitelisted(email string) (bool, error) {
	var count int
	// This query checks if the email exists for ANY event.
	err := r.db.QueryRow("SELECT COUNT(*) FROM whitelisted_emails WHERE email = ?", email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *whitelistRepository) AddWhitelistedEmail(email, role string, eventID *int) error {
	// The unique key is on (email, event_id). If that combination exists, update the role.
	query := `
		INSERT INTO whitelisted_emails (email, role, event_id)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE role = VALUES(role)
	`
	_, err := r.db.Exec(query, email, role, eventID)
	return err
}

func (r *whitelistRepository) GetAllWhitelistedEmails() ([]WhitelistEntry, error) {
	rows, err := r.db.Query("SELECT email, role, event_id FROM whitelisted_emails ORDER BY email")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []WhitelistEntry
	for rows.Next() {
		var entry WhitelistEntry
		if err := rows.Scan(&entry.Email, &entry.Role, &entry.EventID); err != nil {
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

	// The unique key is on (email, event_id). If that combination exists, update the role.
	stmt, err := tx.Prepare(`
		INSERT INTO whitelisted_emails (email, role, event_id)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE role = VALUES(role)
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, entry := range entries {
		_, err := stmt.Exec(entry.Email, entry.Role, entry.EventID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *whitelistRepository) UpdateNullEventIDs(eventID int) error {
	query := `UPDATE whitelisted_emails SET event_id = ? WHERE event_id IS NULL`
	_, err := r.db.Exec(query, eventID)
	return err
}
