package repository

import (
	"backapp/internal/models"
	"database/sql"
	"fmt"
	"strings"
)

type NotificationRepository interface {
	CreateNotification(title, body, createdBy string, eventID *int) (int64, error)
	AddNotificationTargets(notificationID int64, roles []string) error
	GetNotificationsForAccess(roleNames []string, authorID string, includeAuthored bool, limit int) ([]models.Notification, error)
	GetUserIDsByRoles(roleNames []string) ([]string, error)
	GetPushSubscriptionsByUserIDs(userIDs []string) ([]models.PushSubscription, error)
	GetPushSubscriptionsByUserID(userID string) ([]models.PushSubscription, error)
	UpsertPushSubscription(userID, endpoint, authKey, p256dhKey string) error
	DeletePushSubscription(userID, endpoint string) error
}

type notificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) CreateNotification(title, body, createdBy string, eventID *int) (int64, error) {
	query := "INSERT INTO notifications (title, body, created_by, event_id) VALUES (?, ?, ?, ?)"
	var result sql.Result
	var err error

	if eventID != nil {
		result, err = r.db.Exec(query, title, body, createdBy, *eventID)
	} else {
		result, err = r.db.Exec(query, title, body, createdBy, nil)
	}

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *notificationRepository) AddNotificationTargets(notificationID int64, roles []string) error {
	if len(roles) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT IGNORE INTO notification_targets (notification_id, role_name) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, role := range roles {
		if _, err := stmt.Exec(notificationID, role); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *notificationRepository) GetNotificationsForAccess(roleNames []string, authorID string, includeAuthored bool, limit int) ([]models.Notification, error) {
	var args []interface{}
	var filters []string

	if len(roleNames) > 0 {
		placeholders := strings.Repeat(",?", len(roleNames)-1)
		filters = append(filters, fmt.Sprintf("nt.role_name IN (?%s)", placeholders))
		for _, role := range roleNames {
			args = append(args, role)
		}
	}

	if includeAuthored && authorID != "" {
		filters = append(filters, "n.created_by = ?")
		args = append(args, authorID)
	}

	query := `
		SELECT
			n.id,
			n.title,
			n.body,
			n.created_by,
			n.event_id,
			n.created_at,
			GROUP_CONCAT(DISTINCT nt.role_name ORDER BY nt.role_name SEPARATOR ',') AS target_roles
		FROM notifications n
		LEFT JOIN notification_targets nt ON n.id = nt.notification_id
	`

	if len(filters) > 0 {
		query += " WHERE (" + strings.Join(filters, " OR ") + ")"
	} else {
		// どのフィルターも無い場合は空を返す
		return []models.Notification{}, nil
	}

	query += " GROUP BY n.id ORDER BY n.created_at DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification

	for rows.Next() {
		var notif models.Notification
		var createdBy sql.NullString
		var eventID sql.NullInt64
		var targetRoles sql.NullString

		if err := rows.Scan(
			&notif.ID,
			&notif.Title,
			&notif.Body,
			&createdBy,
			&eventID,
			&notif.CreatedAt,
			&targetRoles,
		); err != nil {
			return nil, err
		}

		if createdBy.Valid {
			value := createdBy.String
			notif.CreatedBy = &value
		}
		if eventID.Valid {
			value := int(eventID.Int64)
			notif.EventID = &value
		}
		if targetRoles.Valid && targetRoles.String != "" {
			notif.TargetRoles = strings.Split(targetRoles.String, ",")
		} else {
			notif.TargetRoles = []string{}
		}

		notifications = append(notifications, notif)
	}

	return notifications, nil
}

func (r *notificationRepository) GetUserIDsByRoles(roleNames []string) ([]string, error) {
	if len(roleNames) == 0 {
		return []string{}, nil
	}

	placeholders := strings.Repeat(",?", len(roleNames)-1)
	query := `
		SELECT DISTINCT u.id
		FROM users u
		INNER JOIN user_roles ur ON u.id = ur.user_id
		INNER JOIN roles r ON ur.role_id = r.id
		WHERE r.name IN (?` + placeholders + `)
	`

	args := make([]interface{}, len(roleNames))
	for i, role := range roleNames {
		args[i] = role
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}

	return userIDs, nil
}

func (r *notificationRepository) GetPushSubscriptionsByUserIDs(userIDs []string) ([]models.PushSubscription, error) {
	if len(userIDs) == 0 {
		return []models.PushSubscription{}, nil
	}

	placeholders := strings.Repeat(",?", len(userIDs)-1)
	query := `
		SELECT id, user_id, endpoint, auth_key, p256dh_key, created_at
		FROM push_subscriptions
		WHERE user_id IN (?` + placeholders + `)
	`

	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		args[i] = id
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.PushSubscription
	for rows.Next() {
		var sub models.PushSubscription
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.Endpoint, &sub.AuthKey, &sub.P256dhKey, &sub.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}

	return subs, nil
}

func (r *notificationRepository) GetPushSubscriptionsByUserID(userID string) ([]models.PushSubscription, error) {
	query := `
		SELECT id, user_id, endpoint, auth_key, p256dh_key, created_at
		FROM push_subscriptions
		WHERE user_id = ?
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.PushSubscription
	for rows.Next() {
		var sub models.PushSubscription
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.Endpoint, &sub.AuthKey, &sub.P256dhKey, &sub.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}

	return subs, nil
}

func (r *notificationRepository) UpsertPushSubscription(userID, endpoint, authKey, p256dhKey string) error {
	query := `
		INSERT INTO push_subscriptions (user_id, endpoint, auth_key, p256dh_key)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			auth_key = VALUES(auth_key),
			p256dh_key = VALUES(p256dh_key)
	`
	_, err := r.db.Exec(query, userID, endpoint, authKey, p256dhKey)
	return err
}

func (r *notificationRepository) DeletePushSubscription(userID, endpoint string) error {
	_, err := r.db.Exec("DELETE FROM push_subscriptions WHERE user_id = ? AND endpoint = ?", userID, endpoint)
	return err
}
