package models

import "time"

type Event struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Year       int        `json:"year"`
	Season     string     `json:"season"`
	Start_date *time.Time `json:"start_date"`
	End_date   *time.Time `json:"end_date"`
}

type SetActiveEventRequest struct {
	// Use pointer to allow null (clearing active event)
	EventID *int `json:"event_id"`
}
