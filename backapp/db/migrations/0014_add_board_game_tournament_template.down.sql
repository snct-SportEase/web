DELETE FROM score_logs WHERE board_game_run_id IS NOT NULL OR reason IN ('board_game_win_points', 'board_game_rank_points');

ALTER TABLE score_logs
    DROP CHECK chk_score_logs_reason,
    DROP FOREIGN KEY fk_score_logs_board_game_run,
    DROP INDEX idx_score_logs_board_game_run,
    DROP COLUMN board_game_run_id,
    ADD CONSTRAINT chk_score_logs_reason CHECK (reason IN (
        'attendance_points', 'initial_points', 'survey_points', 'mic_points',
        'gym1_win1_points', 'gym1_win2_points', 'gym1_win3_points', 'gym1_champion_points',
        'gym2_win1_points', 'gym2_win2_points', 'gym2_win3_points', 'gym2_champion_points',
        'gym2_loser_bracket_champion_points',
        'ground_win1_points', 'ground_win2_points', 'ground_win3_points', 'ground_champion_points',
        'noon_game_points'
    ));

DROP TABLE IF EXISTS board_game_rankings;
DROP TABLE IF EXISTS board_game_entry_members;
DROP TABLE IF EXISTS board_game_entries;
DROP TABLE IF EXISTS board_game_runs;

ALTER TABLE teams
    DROP INDEX uq_teams_class_sport_entry,
    DROP COLUMN entry_key,
    ADD CONSTRAINT uq_teams_class_sport UNIQUE (class_id, sport_id);

ALTER TABLE event_sports
    DROP INDEX idx_event_sports_template_key,
    DROP COLUMN template_key;
