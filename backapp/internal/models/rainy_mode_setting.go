package models

// RainyModeSetting represents a rainy mode setting for a specific event, sport, and class
type RainyModeSetting struct {
	ID             int     `json:"id"`
	EventID        int     `json:"event_id"`
	SportID        int     `json:"sport_id"`
	ClassID        int     `json:"class_id"`
	MinCapacity    *int    `json:"min_capacity,omitempty"`
	MaxCapacity    *int    `json:"max_capacity,omitempty"`
	MatchStartTime *string `json:"match_start_time,omitempty"`
}

// RainyModeSettingRequest represents the request body for creating/updating a rainy mode setting
type RainyModeSettingRequest struct {
	SportID        int     `json:"sport_id" binding:"required"`
	ClassID        int     `json:"class_id" binding:"required"`
	MinCapacity    *int    `json:"min_capacity,omitempty"`
	MaxCapacity    *int    `json:"max_capacity,omitempty"`
	MatchStartTime *string `json:"match_start_time,omitempty"`
}
