package models

import "time"

const (
	EventStatusPreparing = "preparing"
	EventStatusUpcoming  = "upcoming"
	EventStatusActive    = "active"
	EventStatusArchived  = "archived"
)

func IsValidEventStatus(status string) bool {
	switch status {
	case EventStatusPreparing, EventStatusUpcoming, EventStatusActive, EventStatusArchived:
		return true
	default:
		return false
	}
}

type Event struct {
	ID                             int        `json:"id"`
	Name                           string     `json:"name"`
	Year                           int        `json:"year"`
	Season                         string     `json:"season"`
	Start_date                     *time.Time `json:"start_date"`
	End_date                       *time.Time `json:"end_date"`
	IsRainyMode                    bool       `json:"is_rainy_mode"`
	CompetitionGuidelinesPdfUrl    *string    `json:"competition_guidelines_pdf_url,omitempty"`
	SurveyUrl                      *string    `json:"survey_url,omitempty"`
	IsSurveyPublished              bool       `json:"is_survey_published"`
	Status                         string     `json:"status"`
	HideScores                     bool       `json:"hide_scores"`
	DuplicateRegistrationThreshold int        `json:"duplicate_registration_threshold"`
}

type SetActiveEventRequest struct {
	// Use pointer to allow null (clearing active event)
	EventID *int `json:"event_id"`
}
