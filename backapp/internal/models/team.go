package models

import "time"

// Team represents the teams table
type Team struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ClassID     int    `json:"class_id"`
	SportID     int    `json:"sport_id"`
	EventID     int    `json:"event_id"`
	MinCapacity *int   `json:"min_capacity"`
	MaxCapacity *int   `json:"max_capacity"`
}

// TeamWithSport represents a team with sport name
type TeamWithSport struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ClassID   int    `json:"class_id"`
	SportID   int    `json:"sport_id"`
	EventID   int    `json:"event_id"`
	SportName string `json:"sport_name"`
}

// BarcodeCheckInRequest represents a request to check in a student to the selected match round with a MyID barcode.
type BarcodeCheckInRequest struct {
	BarcodeData string `json:"barcode_data"`
	EventID     int    `json:"event_id"`
	SportID     int    `json:"sport_id"`
	MatchID     int    `json:"match_id"`
	MatchIDs    []int  `json:"match_ids"`
}

// MatchCheckInMember represents a student checked in for a selected match.
type MatchCheckInMember struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName *string   `json:"display_name"`
	ClassID     int       `json:"class_id"`
	ClassName   string    `json:"class_name"`
	TeamID      int       `json:"team_id"`
	TeamName    string    `json:"team_name"`
	EventID     int       `json:"event_id"`
	SportID     int       `json:"sport_id"`
	MatchID     int       `json:"match_id"`
	Round       int       `json:"round"`
	CheckedInAt time.Time `json:"checked_in_at"`
}
