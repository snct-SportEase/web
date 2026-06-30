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

// BarcodeRequest represents a request to generate a barcode
type BarcodeRequest struct {
	EventID int `json:"event_id"`
	SportID int `json:"sport_id"`
}

// BarcodeResponse represents a barcode generation response
type BarcodeResponse struct {
	BarcodeData string `json:"barcode_data"`
	ExpiresAt   int64  `json:"expires_at"`
}

// BarcodeData represents the data encoded in the barcode
type BarcodeData struct {
	EventID     int    `json:"event_id"`
	SportID     int    `json:"sport_id"`
	SportName   string `json:"sport_name"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Timestamp   int64  `json:"timestamp"`
	ExpiresAt   int64  `json:"expires_at"`
}

// BarcodeVerifyRequest represents a request to verify a barcode
type BarcodeVerifyRequest struct {
	BarcodeData string `json:"barcode_data"`
	EventID     int    `json:"event_id"`
	SportID     int    `json:"sport_id"`
}
