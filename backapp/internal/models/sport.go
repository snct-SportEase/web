package models

// Sport represents the sports table
type Sport struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// EventSport represents the event_sports table
type EventSport struct {
	EventID     int     `json:"event_id"`
	SportID     int     `json:"sport_id"`
	Description *string `json:"description"`
	Rules       *string `json:"rules"`
	RulesType   string  `json:"rules_type"`
	RulesPdfURL *string `json:"rules_pdf_url"`
	Location    string  `json:"location"`
}
