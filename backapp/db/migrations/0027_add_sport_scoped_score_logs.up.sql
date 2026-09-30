ALTER TABLE score_logs
    DROP CHECK chk_score_logs_reason,
    ADD COLUMN sport_id INT NULL AFTER class_id,
    ADD INDEX idx_score_logs_event_sport_class (event_id, sport_id, class_id),
    ADD CONSTRAINT fk_score_logs_sport
        FOREIGN KEY (sport_id) REFERENCES sports(id) ON DELETE SET NULL,
    ADD CONSTRAINT chk_score_logs_reason CHECK (reason IN (
        'attendance_points', 'initial_points', 'survey_points', 'mic_points',
        'gym1_win1_points', 'gym1_win2_points', 'gym1_win3_points', 'gym1_champion_points',
        'gym2_win1_points', 'gym2_win2_points', 'gym2_win3_points', 'gym2_champion_points',
        'gym2_loser_bracket_champion_points',
        'ground_win1_points', 'ground_win2_points', 'ground_win3_points', 'ground_champion_points',
        'tournament_win1_points', 'tournament_win2_points', 'tournament_win3_points',
        'tournament_champion_points',
        'noon_game_points', 'board_game_win_points', 'board_game_rank_points'
    ));

-- Attribute existing tournament scores to their sport. General class points remain
-- NULL because they do not belong to a single sport.
UPDATE score_logs score_log
JOIN matches match_result ON match_result.id = score_log.source_match_id
JOIN tournaments tournament ON tournament.id = match_result.tournament_id
SET score_log.sport_id = tournament.sport_id
WHERE score_log.sport_id IS NULL;

-- Ranking points are not always tied to a match, but the board-game run always
-- carries the owning sport.
UPDATE score_logs score_log
JOIN board_game_runs board_game ON board_game.id = score_log.board_game_run_id
SET score_log.sport_id = board_game.sport_id
WHERE score_log.sport_id IS NULL;
