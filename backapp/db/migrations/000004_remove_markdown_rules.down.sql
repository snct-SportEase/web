ALTER TABLE event_sports
    ADD COLUMN rules TEXT NULL AFTER description,
    ADD COLUMN rules_type ENUM('markdown', 'pdf') NOT NULL DEFAULT 'markdown' AFTER location;
