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
}

type eventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) CreateEvent(event *models.Event) (int64, error) {
	query := "INSERT INTO events (name, `year`, season, start_date, end_date) VALUES (?, ?, ?, ?, ?)"
	result, err := r.db.Exec(query, event.Name, event.Year, event.Season, event.Start_date, event.End_date)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *eventRepository) GetAllEvents() ([]*models.Event, error) {
	query := "SELECT id, name, `year`, season, start_date, end_date FROM events ORDER BY `year` DESC, FIELD(season, 'autumn', 'spring')"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		if err := rows.Scan(&event.ID, &event.Name, &event.Year, &event.Season, &event.Start_date, &event.End_date); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (r *eventRepository) UpdateEvent(event *models.Event) error {
	query := "UPDATE events SET name = ?, `year` = ?, season = ?, start_date = ?, end_date = ? WHERE id = ?"
	_, err := r.db.Exec(query, event.Name, event.Year, event.Season, event.Start_date, event.End_date, event.ID)
	return err
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
	query := "INSERT INTO active_event (id, event_id) VALUES (1, ?) ON DUPLICATE KEY UPDATE event_id = VALUES(event_id)"
	if event_id == nil {
		_, err := r.db.Exec(query, nil)
		return err
	}
	_, err := r.db.Exec(query, *event_id)
	return err
}

func (r *eventRepository) GetEventByYearAndSeason(year int, season string) (*models.Event, error) {
	query := "SELECT id, name, `year`, season, start_date, end_date FROM events WHERE `year` = ? AND season = ?"
	event := &models.Event{}
	err := r.db.QueryRow(query, year, season).Scan(&event.ID, &event.Name, &event.Year, &event.Season, &event.Start_date, &event.End_date)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return event, nil
}

func (r *eventRepository) CopyClassScores(fromEventID int, toEventID int) error {
	query := `
		INSERT INTO class_scores (event_id, class_id, initial_points)
		SELECT ?, class_id, total_points_current_event
		FROM class_scores
		WHERE event_id = ?
		ON DUPLICATE KEY UPDATE initial_points = VALUES(initial_points)
	`
	_, err := r.db.Exec(query, toEventID, fromEventID)
	return err
}
