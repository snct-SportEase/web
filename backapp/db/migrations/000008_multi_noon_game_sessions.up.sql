-- Allow an event to have independent noon-game sessions. Existing sessions are
-- retained and deliberately start as drafts so they are not exposed to students.
ALTER TABLE noon_game_sessions
    DROP INDEX uq_noon_session_event,
    ADD COLUMN template_key VARCHAR(64) NOT NULL DEFAULT 'custom' AFTER event_id,
    ADD COLUMN scheduled_at DATETIME NULL AFTER description,
    ADD COLUMN location VARCHAR(255) NULL AFTER scheduled_at,
    ADD COLUMN status ENUM('draft', 'finalized', 'published') NOT NULL DEFAULT 'draft' AFTER allow_manual_points,
    ADD INDEX idx_noon_game_sessions_event (event_id),
    ADD INDEX idx_noon_game_sessions_event_status (event_id, status);

-- Keep noon-game score history attributable to both its session and input match.
ALTER TABLE score_logs
    ADD COLUMN noon_game_session_id INT NULL AFTER source_match_id,
    ADD COLUMN noon_game_match_id INT NULL AFTER noon_game_session_id,
    ADD INDEX idx_score_logs_noon_game_session (noon_game_session_id),
    ADD CONSTRAINT fk_score_logs_noon_game_session
        FOREIGN KEY (noon_game_session_id) REFERENCES noon_game_sessions(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_score_logs_noon_game_match
        FOREIGN KEY (noon_game_match_id) REFERENCES noon_game_matches(id) ON DELETE SET NULL;
