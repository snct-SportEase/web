package models

// Team represents the teams table
type Team struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	ClassID int    `json:"class_id"`
	SportID int    `json:"sport_id"`
	EventID int    `json:"event_id"`
}
