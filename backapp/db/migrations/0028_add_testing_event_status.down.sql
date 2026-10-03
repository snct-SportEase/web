UPDATE events SET status = 'preparing' WHERE status = 'testing';

ALTER TABLE events
    MODIFY COLUMN status ENUM('preparing', 'upcoming', 'active', 'archived') NOT NULL DEFAULT 'upcoming';
