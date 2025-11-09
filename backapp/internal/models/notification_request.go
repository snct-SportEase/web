package models

import "time"

type NotificationRequestStatus string

const (
	NotificationRequestStatusPending  NotificationRequestStatus = "pending"
	NotificationRequestStatusApproved NotificationRequestStatus = "approved"
	NotificationRequestStatusRejected NotificationRequestStatus = "rejected"
)

type NotificationRequest struct {
	ID          int                           `json:"id"`
	Title       string                        `json:"title"`
	Body        string                        `json:"body"`
	TargetText  string                        `json:"target_text"`
	Status      NotificationRequestStatus     `json:"status"`
	RequesterID string                        `json:"requester_id"`
	ResolvedBy  *string                       `json:"resolved_by,omitempty"`
	ResolvedAt  *time.Time                    `json:"resolved_at,omitempty"`
	CreatedAt   time.Time                     `json:"created_at"`
	UpdatedAt   time.Time                     `json:"updated_at"`
	Requester   *User                         `json:"requester,omitempty"`
	Resolver    *User                         `json:"resolver,omitempty"`
	Messages    []*NotificationRequestMessage `json:"messages,omitempty"`
}

type NotificationRequestMessage struct {
	ID        int       `json:"id"`
	RequestID int       `json:"request_id"`
	SenderID  string    `json:"sender_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
	Sender    *User     `json:"sender,omitempty"`
	IsSystem  bool      `json:"is_system"`
}
