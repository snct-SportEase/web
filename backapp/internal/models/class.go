package models

type Class struct {
	ID           int    `json:"id"`
	EventID      *int   `json:"event_id,omitempty"`
	Name         string `json:"name"`
	StudentCount int    `json:"student_count"`
	AttendCount  int    `json:"attend_count"`
}
