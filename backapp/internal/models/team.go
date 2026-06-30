package models

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

// BarcodeCheckInRequest represents a request to check in a student to a round with a MyID barcode.
type BarcodeCheckInRequest struct {
	BarcodeData string `json:"barcode_data"`
	EventID     int    `json:"event_id"`
	SportID     int    `json:"sport_id"`
	Round       int    `json:"round"`
}
