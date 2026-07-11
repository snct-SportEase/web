ALTER TABLE events
    MODIFY COLUMN status ENUM('upcoming', 'active', 'archived') NOT NULL DEFAULT 'upcoming';
