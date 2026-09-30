DELETE FROM score_logs
WHERE reason IN (
    'tournament_win1_points',
    'tournament_win2_points',
    'tournament_win3_points',
    'tournament_champion_points'
);

ALTER TABLE score_logs
    DROP CHECK chk_score_logs_reason,
    DROP FOREIGN KEY fk_score_logs_sport,
    DROP INDEX idx_score_logs_event_sport_class,
    DROP COLUMN sport_id,
    ADD CONSTRAINT chk_score_logs_reason CHECK (reason IN (
        'attendance_points', 'initial_points', 'survey_points', 'mic_points',
        'gym1_win1_points', 'gym1_win2_points', 'gym1_win3_points', 'gym1_champion_points',
        'gym2_win1_points', 'gym2_win2_points', 'gym2_win3_points', 'gym2_champion_points',
        'gym2_loser_bracket_champion_points',
        'ground_win1_points', 'ground_win2_points', 'ground_win3_points', 'ground_champion_points',
        'noon_game_points', 'board_game_win_points', 'board_game_rank_points'
    ));
