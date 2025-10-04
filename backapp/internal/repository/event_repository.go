package repository

import (
	"backapp/internal/models"
	"database/sql"
)

type EventRepository interface {
	CreateEvent(event *models.Event) (int64, error)
	GetAllEvents() ([]*models.Event, error)
	UpdateEvent(event *models.Event) error
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
