ALTER TABLE events
    MODIFY COLUMN status ENUM('preparing', 'upcoming', 'active', 'archived') NOT NULL DEFAULT 'upcoming';
