UPDATE event_sports
SET location = 'other'
WHERE location NOT IN ('gym1', 'gym2', 'ground', 'pool', 'noon_game', 'other');

ALTER TABLE event_sports
    MODIFY COLUMN location ENUM('gym1', 'gym2', 'ground', 'pool', 'noon_game', 'other') NOT NULL;
