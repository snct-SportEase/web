package repository

import (
	"backapp/internal/models"
	"database/sql"
)

type RainyModeRepository interface {
	GetSettingsByEventID(eventID int) ([]*models.RainyModeSetting, error)
	GetSetting(eventID int, sportID int, classID int) (*models.RainyModeSetting, error)
	UpsertSetting(setting *models.RainyModeSetting) error
	DeleteSetting(eventID int, sportID int, classID int) error
}

type rainyModeRepository struct {
	db *sql.DB
}

func NewRainyModeRepository(db *sql.DB) RainyModeRepository {
	return &rainyModeRepository{db: db}
}

func (r *rainyModeRepository) GetSettingsByEventID(eventID int) ([]*models.RainyModeSetting, error) {
	query := `
		SELECT id, event_id, sport_id, class_id, min_capacity, max_capacity, match_start_time
		FROM rainy_mode_settings
		WHERE event_id = ?
		ORDER BY sport_id, class_id
	`
	rows, err := r.db.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*models.RainyModeSetting
	for rows.Next() {
		var setting models.RainyModeSetting
		var minCapacity, maxCapacity sql.NullInt64
		var matchStartTime sql.NullString

		err := rows.Scan(
			&setting.ID,
			&setting.EventID,
			&setting.SportID,
			&setting.ClassID,
			&minCapacity,
			&maxCapacity,
			&matchStartTime,
		)
		if err != nil {
			return nil, err
		}

		if minCapacity.Valid {
			capacity := int(minCapacity.Int64)
			setting.MinCapacity = &capacity
		}
		if maxCapacity.Valid {
			capacity := int(maxCapacity.Int64)
			setting.MaxCapacity = &capacity
		}
		if matchStartTime.Valid {
			setting.MatchStartTime = &matchStartTime.String
		}

		settings = append(settings, &setting)
	}

	return settings, nil
}

func (r *rainyModeRepository) GetSetting(eventID int, sportID int, classID int) (*models.RainyModeSetting, error) {
	query := `
		SELECT id, event_id, sport_id, class_id, min_capacity, max_capacity, match_start_time
		FROM rainy_mode_settings
		WHERE event_id = ? AND sport_id = ? AND class_id = ?
	`
	var setting models.RainyModeSetting
	var minCapacity, maxCapacity sql.NullInt64
	var matchStartTime sql.NullString

	err := r.db.QueryRow(query, eventID, sportID, classID).Scan(
		&setting.ID,
		&setting.EventID,
		&setting.SportID,
		&setting.ClassID,
		&minCapacity,
		&maxCapacity,
		&matchStartTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if minCapacity.Valid {
		capacity := int(minCapacity.Int64)
		setting.MinCapacity = &capacity
	}
	if maxCapacity.Valid {
		capacity := int(maxCapacity.Int64)
		setting.MaxCapacity = &capacity
	}
	if matchStartTime.Valid {
		setting.MatchStartTime = &matchStartTime.String
	}

	return &setting, nil
}

func (r *rainyModeRepository) UpsertSetting(setting *models.RainyModeSetting) error {
	query := `
		INSERT INTO rainy_mode_settings (event_id, sport_id, class_id, min_capacity, max_capacity, match_start_time)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			min_capacity = VALUES(min_capacity),
			max_capacity = VALUES(max_capacity),
			match_start_time = VALUES(match_start_time)
	`

	var minCapacity, maxCapacity interface{}
	var matchStartTime interface{}

	if setting.MinCapacity != nil {
		minCapacity = *setting.MinCapacity
	}
	if setting.MaxCapacity != nil {
		maxCapacity = *setting.MaxCapacity
	}
	if setting.MatchStartTime != nil {
		matchStartTime = *setting.MatchStartTime
	}

	result, err := r.db.Exec(
		query,
		setting.EventID,
		setting.SportID,
		setting.ClassID,
		minCapacity,
		maxCapacity,
		matchStartTime,
	)
	if err != nil {
		return err
	}

	// Get the ID if it's a new insert
	if setting.ID == 0 {
		id, err := result.LastInsertId()
		if err == nil {
			setting.ID = int(id)
		}
	}

	return nil
}

func (r *rainyModeRepository) DeleteSetting(eventID int, sportID int, classID int) error {
	query := `
		DELETE FROM rainy_mode_settings
		WHERE event_id = ? AND sport_id = ? AND class_id = ?
	`
	_, err := r.db.Exec(query, eventID, sportID, classID)
	return err
}
