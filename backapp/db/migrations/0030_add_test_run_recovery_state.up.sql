ALTER TABLE event_test_runs
    ADD COLUMN state ENUM('starting', 'testing', 'restoring', 'awaiting_notification_resume', 'failed') NOT NULL DEFAULT 'testing' AFTER event_id,
    ADD COLUMN last_error TEXT NULL AFTER state,
    ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER started_at;
