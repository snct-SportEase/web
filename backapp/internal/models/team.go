package models

// Team represents the teams table
type Team struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	ClassID int    `json:"class_id"`
	SportID int    `json:"sport_id"`
	EventID int    `json:"event_id"`
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

// QRCodeRequest represents a request to generate a QR code
type QRCodeRequest struct {
	EventID int `json:"event_id"`
	SportID int `json:"sport_id"`
}

// QRCodeResponse represents a QR code generation response
type QRCodeResponse struct {
	QRCodeData string `json:"qr_code_data"`
	ExpiresAt  int64  `json:"expires_at"`
}

// QRCodeData represents the data encoded in the QR code
type QRCodeData struct {
	EventID     int    `json:"event_id"`
	SportID     int    `json:"sport_id"`
	SportName   string `json:"sport_name"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Timestamp   int64  `json:"timestamp"`
	ExpiresAt   int64  `json:"expires_at"`
}

// QRCodeVerifyRequest represents a request to verify a QR code
type QRCodeVerifyRequest struct {
	QRCodeData string `json:"qr_code_data"`
}
