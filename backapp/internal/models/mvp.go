package models

import "time"

// MVPVote represents a vote for the MVP.
// @Description MVPVote represents a vote for the MVP.
// @Description MVPVote model
type MVPVote struct {
	ID              int       `json:"id"`
	EventID         int       `json:"event_id"`
	VoterUserID     string    `json:"voter_user_id"`
	VotedForClassID int       `json:"voted_for_class_id"`
	Reason          string    `json:"reason"`
	Points          int       `json:"points"`
	CreatedAt       time.Time `json:"created_at"`
}
