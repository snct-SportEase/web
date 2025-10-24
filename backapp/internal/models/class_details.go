package models

type ClassDetails struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	StudentCount     int    `json:"student_count"`
	AttendancePoints int    `json:"attendance_points"`
}
