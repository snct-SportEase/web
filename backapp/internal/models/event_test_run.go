package models

import "time"

const (
	EventTestRunStateStarting                   = "starting"
	EventTestRunStateTesting                    = "testing"
	EventTestRunStateRestoring                  = "restoring"
	EventTestRunStateAwaitingNotificationResume = "awaiting_notification_resume"
	EventTestRunStateFailed                     = "failed"

	NotificationResumePolicyResume        = "resume"
	NotificationResumePolicyShift         = "shift"
	NotificationResumePolicyCancelOverdue = "cancel_overdue"
)

type EventTestRunStatus struct {
	EventID                  int       `json:"event_id"`
	State                    string    `json:"state"`
	StartedAt                time.Time `json:"started_at"`
	UpdatedAt                time.Time `json:"updated_at"`
	LastError                *string   `json:"last_error,omitempty"`
	OverdueNotificationCount int       `json:"overdue_notification_count"`
}

func (s *EventTestRunStatus) IsIsolatingUsers() bool {
	if s == nil {
		return false
	}
	return s.State != EventTestRunStateAwaitingNotificationResume
}
