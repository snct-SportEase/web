package models

import "time"

type User struct {
	ID                string    `json:"id"`
	Email             string    `json:"email"`
	DisplayName       *string   `json:"display_name"`
	ClassID           *int      `json:"class_id"`
	IsProfileComplete bool      `json:"is_profile_complete"`
	Roles             []Role    `json:"roles,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
	ClassID     int    `json:"class_id"`
}
