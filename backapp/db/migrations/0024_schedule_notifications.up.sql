ALTER TABLE notifications
    ADD COLUMN scheduled_at DATETIME(6) NULL AFTER event_id,
    ADD COLUMN sent_at DATETIME(6) NULL AFTER scheduled_at,
    ADD INDEX idx_notifications_due (sent_at, scheduled_at);

-- Notifications created before this migration were all delivered immediately.
UPDATE notifications
SET sent_at = created_at
WHERE sent_at IS NULL;
