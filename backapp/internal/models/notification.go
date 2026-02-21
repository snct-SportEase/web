package models

import "time"

type Notification struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   *string   `json:"created_by,omitempty"`
	EventID     *int      `json:"event_id,omitempty"`
	TargetRoles []string  `json:"target_roles"`
}

type PushSubscription struct {
	ID        int       `json:"id"`
	UserID    string    `json:"user_id"`
	Endpoint  string    `json:"endpoint"`
	AuthKey   string    `json:"auth_key"`
	P256dhKey string    `json:"p256dh_key"`
	CreatedAt time.Time `json:"created_at"`
}
