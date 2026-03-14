package models

import "time"

// MICVote represents a vote for the MIC.
// @Description MICVote represents a vote for the MIC.
// @Description MICVote model
type MICVote struct {
	ID              int       `json:"id"`
	EventID         int       `json:"event_id"`
	VoterUserID     string    `json:"voter_user_id"`
	VotedForClassID int       `json:"voted_for_class_id"`
	Reason          string    `json:"reason"`
	Points          int       `json:"points"`
	CreatedAt       time.Time `json:"created_at"`
}

// MICResult represents the result of the MIC calculation.
type MICResult struct {
	ClassName   string `json:"class_name"`
	TotalPoints int    `json:"total_points"`
	Season      string `json:"season"`
	EventID     int    `json:"event_id"`
}
