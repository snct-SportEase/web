ALTER TABLE events
    MODIFY COLUMN status ENUM('preparing', 'testing', 'upcoming', 'active', 'archived') NOT NULL DEFAULT 'upcoming';
