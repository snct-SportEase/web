package repository

import (
	"backapp/internal/models"
	"database/sql"
)

type EventRepository interface {
	CreateEvent(event *models.Event) (int64, error)
	GetAllEvents() ([]*models.Event, error)
	UpdateEvent(event *models.Event) error
	GetActiveEvent() (event_id int, err error)
	SetActiveEvent(event_id *int) error
	GetEventByYearAndSeason(year int, season string) (*models.Event, error)
	CopyClassScores(fromEventID int, toEventID int) error
	GetEventByID(id int) (*models.Event, error)
	SetRainyMode(eventID int, isRainyMode bool) error
	PublishSurvey(eventID int) error
}

type eventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) GetEventByID(id int) (*models.Event, error) {
	query := "SELECT id, name, `year`, season, start_date, end_date, is_rainy_mode, competition_guidelines_pdf_url, survey_url, is_survey_published, status, hide_scores FROM events WHERE id = ?"
	event := &models.Event{}
	var competitionGuidelinesPdfUrl sql.NullString
	var surveyUrl sql.NullString
	err := r.db.QueryRow(query, id).Scan(&event.ID, &event.Name, &event.Year, &event.Season, &event.Start_date, &event.End_date, &event.IsRainyMode, &competitionGuidelinesPdfUrl, &surveyUrl, &event.IsSurveyPublished, &event.Status, &event.HideScores)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	if competitionGuidelinesPdfUrl.Valid {
		event.CompetitionGuidelinesPdfUrl = &competitionGuidelinesPdfUrl.String
	}
	if surveyUrl.Valid {
		event.SurveyUrl = &surveyUrl.String
	}
	return event, nil
}

func (r *eventRepository) CreateEvent(event *models.Event) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	query := "INSERT INTO events (name, `year`, season, start_date, end_date, is_rainy_mode, competition_guidelines_pdf_url, survey_url, is_survey_published, status, hide_scores) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	result, err := tx.Exec(query, event.Name, event.Year, event.Season, event.Start_date, event.End_date, event.IsRainyMode, event.CompetitionGuidelinesPdfUrl, event.SurveyUrl, event.IsSurveyPublished, event.Status, event.HideScores)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if event.Status == "active" {
		archiveQuery := "UPDATE events SET status = 'archived' WHERE id != ? AND status = 'active'"
		_, err = tx.Exec(archiveQuery, id)
		if err != nil {
			tx.Rollback()
			return 0, err
		}

		activeQuery := "INSERT INTO active_event (id, event_id) VALUES (1, ?) ON DUPLICATE KEY UPDATE event_id = VALUES(event_id)"
		_, err = tx.Exec(activeQuery, id)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *eventRepository) GetAllEvents() ([]*models.Event, error) {
	query := "SELECT id, name, `year`, season, start_date, end_date, is_rainy_mode, competition_guidelines_pdf_url, survey_url, is_survey_published, status, hide_scores FROM events ORDER BY `year` DESC, FIELD(season, 'autumn', 'spring')"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		var competitionGuidelinesPdfUrl sql.NullString
		var surveyUrl sql.NullString
		if err := rows.Scan(&event.ID, &event.Name, &event.Year, &event.Season, &event.Start_date, &event.End_date, &event.IsRainyMode, &competitionGuidelinesPdfUrl, &surveyUrl, &event.IsSurveyPublished, &event.Status, &event.HideScores); err != nil {
			return nil, err
		}
		if competitionGuidelinesPdfUrl.Valid {
			event.CompetitionGuidelinesPdfUrl = &competitionGuidelinesPdfUrl.String
		}
		if surveyUrl.Valid {
			event.SurveyUrl = &surveyUrl.String
		}
		events = append(events, event)
	}
	return events, nil
}

func (r *eventRepository) UpdateEvent(event *models.Event) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	query := "UPDATE events SET name = ?, `year` = ?, season = ?, start_date = ?, end_date = ?, is_rainy_mode = ?, competition_guidelines_pdf_url = ?, survey_url = ?, is_survey_published = ?, status = ?, hide_scores = ? WHERE id = ?"
	_, err = tx.Exec(query, event.Name, event.Year, event.Season, event.Start_date, event.End_date, event.IsRainyMode, event.CompetitionGuidelinesPdfUrl, event.SurveyUrl, event.IsSurveyPublished, event.Status, event.HideScores, event.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if event.Status == "active" {
		archiveQuery := "UPDATE events SET status = 'archived' WHERE id != ? AND status = 'active'"
		_, err = tx.Exec(archiveQuery, event.ID)
		if err != nil {
			tx.Rollback()
			return err
		}

		activeQuery := "INSERT INTO active_event (id, event_id) VALUES (1, ?) ON DUPLICATE KEY UPDATE event_id = VALUES(event_id)"
		_, err = tx.Exec(activeQuery, event.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		clearActiveQuery := "UPDATE active_event SET event_id = NULL WHERE id = 1 AND event_id = ?"
		_, err = tx.Exec(clearActiveQuery, event.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *eventRepository) GetActiveEvent() (event_id int, err error) {
	query := "SELECT event_id FROM active_event WHERE id = 1"
	var nullableEventId sql.NullInt64
	err = r.db.QueryRow(query).Scan(&nullableEventId)
	if err != nil {
		if err == sql.ErrNoRows {
			// レコードが存在しない場合は、0とnilを返す
			return 0, nil
		}
		return 0, err
	}

	if nullableEventId.Valid {
		return int(nullableEventId.Int64), nil
	}
	// event_id が NULL の場合
	return 0, nil
}

func (r *eventRepository) SetActiveEvent(event_id *int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	query := "INSERT INTO active_event (id, event_id) VALUES (1, ?) ON DUPLICATE KEY UPDATE event_id = VALUES(event_id)"
	var execErr error
	if event_id == nil {
		_, execErr = tx.Exec(query, nil)
	} else {
		_, execErr = tx.Exec(query, *event_id)
	}
	if execErr != nil {
		tx.Rollback()
		return execErr
	}

	if event_id != nil {
		_, err = tx.Exec("UPDATE events SET status = 'archived' WHERE id != ? AND status = 'active'", *event_id)
		if err != nil {
			tx.Rollback()
			return err
		}
		_, err = tx.Exec("UPDATE events SET status = 'active' WHERE id = ?", *event_id)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		_, err = tx.Exec("UPDATE events SET status = 'archived' WHERE status = 'active'")
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *eventRepository) GetEventByYearAndSeason(year int, season string) (*models.Event, error) {
	query := "SELECT id, name, `year`, season, start_date, end_date, is_rainy_mode, competition_guidelines_pdf_url, survey_url, is_survey_published, status FROM events WHERE `year` = ? AND season = ?"
	event := &models.Event{}
	var competitionGuidelinesPdfUrl sql.NullString
	var surveyUrl sql.NullString
	err := r.db.QueryRow(query, year, season).Scan(&event.ID, &event.Name, &event.Year, &event.Season, &event.Start_date, &event.End_date, &event.IsRainyMode, &competitionGuidelinesPdfUrl, &surveyUrl, &event.IsSurveyPublished, &event.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	if competitionGuidelinesPdfUrl.Valid {
		event.CompetitionGuidelinesPdfUrl = &competitionGuidelinesPdfUrl.String
	}
	if surveyUrl.Valid {
		event.SurveyUrl = &surveyUrl.String
	}
	return event, nil
}

func (r *eventRepository) CopyClassScores(fromEventID int, toEventID int) error {
	deleteQuery := "DELETE FROM score_logs WHERE event_id = ? AND reason = 'initial_points'"
	_, err := r.db.Exec(deleteQuery, toEventID)
	if err != nil {
		return err
	}

	insertQuery := `
		INSERT INTO score_logs (event_id, class_id, points, reason)
		SELECT ?, class_id, total_points_current_event, 'initial_points'
		FROM class_scores
		WHERE event_id = ? AND total_points_current_event > 0
	`
	_, err = r.db.Exec(insertQuery, toEventID, fromEventID)
	return err
}

func (r *eventRepository) SetRainyMode(eventID int, isRainyMode bool) error {
	query := "UPDATE events SET is_rainy_mode = ? WHERE id = ?"
	_, err := r.db.Exec(query, isRainyMode, eventID)
	return err
}

func (r *eventRepository) PublishSurvey(eventID int) error {
	query := "UPDATE events SET is_survey_published = TRUE WHERE id = ?"
	_, err := r.db.Exec(query, eventID)
	return err
}
