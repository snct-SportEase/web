package repository

import (
	"backapp/internal/models"
	"database/sql"
	"errors"
)

type NotificationRequestRepository interface {
	CreateRequest(req *models.NotificationRequest) (int64, error)
	UpdateRequestStatus(id int, status models.NotificationRequestStatus, resolverID *string) error
	GetRequestByID(id int) (*models.NotificationRequest, error)
	GetRequestsByRequester(requesterID string) ([]*models.NotificationRequest, error)
	GetAllRequests() ([]*models.NotificationRequest, error)
	AddMessage(requestID int, senderID, message string) (int64, error)
	GetMessages(requestID int) ([]*models.NotificationRequestMessage, error)
	GetParticipants(requestID int) (string, *string, error)
}

type notificationRequestRepository struct {
	db *sql.DB
}

func NewNotificationRequestRepository(db *sql.DB) NotificationRequestRepository {
	return &notificationRequestRepository{db: db}
}

func (r *notificationRequestRepository) CreateRequest(req *models.NotificationRequest) (int64, error) {
	query := `
		INSERT INTO notification_requests (title, body, target_text, status, requester_id)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, req.Title, req.Body, req.TargetText, req.Status, req.RequesterID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *notificationRequestRepository) UpdateRequestStatus(id int, status models.NotificationRequestStatus, resolverID *string) error {
	query := `
		UPDATE notification_requests
		SET status = ?, resolved_by = ?, resolved_at = CASE WHEN ? IS NOT NULL THEN NOW() ELSE resolved_at END
		WHERE id = ?
	`
	var resolver interface{}
	if resolverID != nil {
		resolver = *resolverID
	} else {
		resolver = nil
	}
	resolvedParam := resolver
	_, err := r.db.Exec(query, status, resolver, resolvedParam, id)
	return err
}

func (r *notificationRequestRepository) GetRequestByID(id int) (*models.NotificationRequest, error) {
	query := `
		SELECT
			nr.id,
			nr.title,
			nr.body,
			nr.target_text,
			nr.status,
			nr.requester_id,
			nr.resolved_by,
			nr.resolved_at,
			nr.created_at,
			nr.updated_at,
			req.id,
			req.email,
			req.display_name,
			res.id,
			res.email,
			res.display_name
		FROM notification_requests nr
		JOIN users req ON nr.requester_id = req.id
		LEFT JOIN users res ON nr.resolved_by = res.id
		WHERE nr.id = ?
	`
	row := r.db.QueryRow(query, id)
	req := &models.NotificationRequest{}
	var resolvedBy sql.NullString
	var resolvedAt sql.NullTime
	requester := &models.User{}
	var requesterEmail string
	var requesterDisplay sql.NullString
	var resolverID sql.NullString
	var resolverEmail sql.NullString
	var resolverDisplay sql.NullString

	if err := row.Scan(
		&req.ID,
		&req.Title,
		&req.Body,
		&req.TargetText,
		&req.Status,
		&req.RequesterID,
		&resolvedBy,
		&resolvedAt,
		&req.CreatedAt,
		&req.UpdatedAt,
		&requester.ID,
		&requesterEmail,
		&requesterDisplay,
		&resolverID,
		&resolverEmail,
		&resolverDisplay,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if resolvedBy.Valid {
		req.ResolvedBy = &resolvedBy.String
	}
	if resolvedAt.Valid {
		value := resolvedAt.Time
		req.ResolvedAt = &value
	}
	requester.Email = requesterEmail
	if requesterDisplay.Valid {
		name := requesterDisplay.String
		requester.DisplayName = &name
	}
	req.Requester = requester
	if resolverID.Valid {
		resolverUser := &models.User{
			ID:    resolverID.String,
			Email: resolverEmail.String,
		}
		if resolverDisplay.Valid {
			name := resolverDisplay.String
			resolverUser.DisplayName = &name
		}
		req.Resolver = resolverUser
	}
	return req, nil
}

func (r *notificationRequestRepository) GetRequestsByRequester(requesterID string) ([]*models.NotificationRequest, error) {
	query := `
		SELECT
			nr.id,
			nr.title,
			nr.body,
			nr.target_text,
			nr.status,
			nr.requester_id,
			nr.resolved_by,
			nr.resolved_at,
			nr.created_at,
			nr.updated_at
		FROM notification_requests nr
		WHERE nr.requester_id = ?
		ORDER BY nr.created_at DESC
	`
	rows, err := r.db.Query(query, requesterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.NotificationRequest
	for rows.Next() {
		req := &models.NotificationRequest{}
		var resolvedBy sql.NullString
		var resolvedAt sql.NullTime
		if err := rows.Scan(
			&req.ID,
			&req.Title,
			&req.Body,
			&req.TargetText,
			&req.Status,
			&req.RequesterID,
			&resolvedBy,
			&resolvedAt,
			&req.CreatedAt,
			&req.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if resolvedBy.Valid {
			req.ResolvedBy = &resolvedBy.String
		}
		if resolvedAt.Valid {
			value := resolvedAt.Time
			req.ResolvedAt = &value
		}
		results = append(results, req)
	}
	return results, nil
}

func (r *notificationRequestRepository) GetAllRequests() ([]*models.NotificationRequest, error) {
	query := `
		SELECT
			nr.id,
			nr.title,
			nr.body,
			nr.target_text,
			nr.status,
			nr.requester_id,
			nr.resolved_by,
			nr.resolved_at,
			nr.created_at,
			nr.updated_at,
			req.id,
			req.email,
			req.display_name
		FROM notification_requests nr
		JOIN users req ON nr.requester_id = req.id
		ORDER BY nr.status = 'pending' DESC, nr.created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.NotificationRequest
	for rows.Next() {
		req := &models.NotificationRequest{}
		var resolvedBy sql.NullString
		var resolvedAt sql.NullTime
		var requesterEmail string
		var requesterDisplay sql.NullString
		requester := &models.User{}
		if err := rows.Scan(
			&req.ID,
			&req.Title,
			&req.Body,
			&req.TargetText,
			&req.Status,
			&req.RequesterID,
			&resolvedBy,
			&resolvedAt,
			&req.CreatedAt,
			&req.UpdatedAt,
			&requester.ID,
			&requesterEmail,
			&requesterDisplay,
		); err != nil {
			return nil, err
		}
		if resolvedBy.Valid {
			req.ResolvedBy = &resolvedBy.String
		}
		if resolvedAt.Valid {
			value := resolvedAt.Time
			req.ResolvedAt = &value
		}
		requester.Email = requesterEmail
		if requesterDisplay.Valid {
			name := requesterDisplay.String
			requester.DisplayName = &name
		}
		req.Requester = requester
		results = append(results, req)
	}
	return results, nil
}

func (r *notificationRequestRepository) AddMessage(requestID int, senderID, message string) (int64, error) {
	query := `
		INSERT INTO notification_request_messages (request_id, sender_id, message)
		VALUES (?, ?, ?)
	`
	res, err := r.db.Exec(query, requestID, senderID, message)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *notificationRequestRepository) GetMessages(requestID int) ([]*models.NotificationRequestMessage, error) {
	query := `
		SELECT m.id, m.request_id, m.sender_id, m.message, m.created_at,
		       u.id, u.email, u.display_name
		FROM notification_request_messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.request_id = ?
		ORDER BY m.created_at ASC
	`
	rows, err := r.db.Query(query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.NotificationRequestMessage
	for rows.Next() {
		msg := &models.NotificationRequestMessage{}
		sender := &models.User{}
		var senderEmail string
		var senderDisplay sql.NullString
		if err := rows.Scan(
			&msg.ID,
			&msg.RequestID,
			&msg.SenderID,
			&msg.Message,
			&msg.CreatedAt,
			&sender.ID,
			&senderEmail,
			&senderDisplay,
		); err != nil {
			return nil, err
		}
		sender.Email = senderEmail
		if senderDisplay.Valid {
			name := senderDisplay.String
			sender.DisplayName = &name
		}
		msg.Sender = sender
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *notificationRequestRepository) GetParticipants(requestID int) (string, *string, error) {
	query := `
		SELECT requester_id, resolved_by
		FROM notification_requests
		WHERE id = ?
	`
	row := r.db.QueryRow(query, requestID)
	var requester string
	var resolvedBy sql.NullString
	if err := row.Scan(&requester, &resolvedBy); err != nil {
		return "", nil, err
	}
	var resolverPtr *string
	if resolvedBy.Valid {
		value := resolvedBy.String
		resolverPtr = &value
	}
	return requester, resolverPtr, nil
}
