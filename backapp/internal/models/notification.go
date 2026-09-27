package models

import "time"

type Notification struct {
	ID              int        `json:"id"`
	Title           string     `json:"title"`
	Body            string     `json:"body"`
	Type            string     `json:"type"`
	CreatedAt       time.Time  `json:"created_at"`
	CreatedBy       *string    `json:"created_by,omitempty"`
	EventID         *int       `json:"event_id,omitempty"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	SentAt          *time.Time `json:"sent_at,omitempty"`
	TargetRoles     []string   `json:"target_roles"`
	TargetUserCount int        `json:"target_user_count"`
}

// ScheduledNotification contains the persisted delivery targets needed by the
// background worker after a scheduled notification becomes due.
type ScheduledNotification struct {
	ID            int
	Title         string
	Body          string
	Type          string
	EventID       *int
	ScheduledAt   time.Time
	TargetRoles   []string
	TargetUserIDs []string
}

type PushSubscription struct {
	ID        int       `json:"id"`
	UserID    string    `json:"user_id"`
	Endpoint  string    `json:"endpoint"`
	AuthKey   string    `json:"auth_key"`
	P256dhKey string    `json:"p256dh_key"`
	CreatedAt time.Time `json:"created_at"`
}

type PushSubscriptionStats struct {
	TargetUserCount           int `json:"target_user_count"`
	SubscribedUserCount       int `json:"subscribed_user_count"`
	SubscriptionEndpointCount int `json:"subscription_endpoint_count"`
}
