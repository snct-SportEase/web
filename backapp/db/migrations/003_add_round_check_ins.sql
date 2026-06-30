CREATE TABLE round_check_ins (
    id INT PRIMARY KEY AUTO_INCREMENT,
    event_id INT NOT NULL,
    sport_id INT NOT NULL,
    round INT NOT NULL,
    user_id CHAR(36) NOT NULL,
    team_id INT NOT NULL,
    checked_in_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_round_check_ins (event_id, sport_id, round, user_id),
    INDEX idx_round_check_ins_event_sport_round (event_id, sport_id, round),
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    FOREIGN KEY (sport_id) REFERENCES sports(id) ON DELETE RESTRICT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
