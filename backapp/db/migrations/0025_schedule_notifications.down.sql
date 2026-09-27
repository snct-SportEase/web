ALTER TABLE notifications
    DROP INDEX idx_notifications_due,
    DROP COLUMN sent_at,
    DROP COLUMN scheduled_at;
