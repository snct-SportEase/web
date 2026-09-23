CREATE TABLE noon_game_check_ins (
    id INT PRIMARY KEY AUTO_INCREMENT,
    event_id INT NOT NULL,
    session_id INT NOT NULL,
    match_id INT NOT NULL,
    user_id CHAR(36) NOT NULL,
    checked_in_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_noon_game_check_ins (event_id, session_id, match_id, user_id),
    INDEX idx_noon_game_check_ins_match (event_id, session_id, match_id),
    CONSTRAINT fk_noon_game_check_ins_event FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    CONSTRAINT fk_noon_game_check_ins_session FOREIGN KEY (session_id) REFERENCES noon_game_sessions(id) ON DELETE CASCADE,
    CONSTRAINT fk_noon_game_check_ins_match FOREIGN KEY (match_id) REFERENCES noon_game_matches(id) ON DELETE CASCADE,
    CONSTRAINT fk_noon_game_check_ins_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
