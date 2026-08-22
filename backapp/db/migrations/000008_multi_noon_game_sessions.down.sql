-- A rollback is safe only while every event still has at most one session.
ALTER TABLE score_logs
    DROP FOREIGN KEY fk_score_logs_noon_game_match,
    DROP FOREIGN KEY fk_score_logs_noon_game_session,
    DROP INDEX idx_score_logs_noon_game_session,
    DROP COLUMN noon_game_match_id,
    DROP COLUMN noon_game_session_id;

ALTER TABLE noon_game_sessions
    DROP INDEX idx_noon_game_sessions_event_status,
    DROP INDEX idx_noon_game_sessions_event,
    DROP COLUMN status,
    DROP COLUMN location,
    DROP COLUMN scheduled_at,
    DROP COLUMN template_key,
    ADD UNIQUE KEY uq_noon_session_event (event_id);
