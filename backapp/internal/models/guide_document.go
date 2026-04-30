package models

import "time"

type GuideDocument struct {
	ID          int       `json:"id"`
	EventID     int       `json:"event_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	PdfURL      string    `json:"pdf_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
