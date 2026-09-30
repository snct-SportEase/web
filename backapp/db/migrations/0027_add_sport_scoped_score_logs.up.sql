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

-- Recover completed ordinary tournaments that previously received no points
-- because their venue was "other". Existing logs and dedicated board-game
-- tournaments are excluded, so their points are never awarded twice.
INSERT INTO score_logs (event_id, class_id, sport_id, points, reason, source_match_id)
WITH completed AS (
    SELECT m.*, t.event_id, t.sport_id, rounds.max_round,
        COALESCE(m.winner_team_id,
            CASE WHEN m.team1_score > m.team2_score THEN m.team1_id
                 WHEN m.team2_score > m.team1_score THEN m.team2_id END) AS winner_id
    FROM matches m
    JOIN tournaments t ON t.id = m.tournament_id
    JOIN event_sports es ON es.event_id=t.event_id AND es.sport_id=t.sport_id
    JOIN (SELECT tournament_id, MAX(round) AS max_round FROM matches GROUP BY tournament_id) rounds
        ON rounds.tournament_id=m.tournament_id
    WHERE (es.location='other' OR es.location LIKE 'other:%')
        AND m.status='finished' AND m.team1_id IS NOT NULL AND m.team2_id IS NOT NULL
        AND m.is_loser_bracket_match=FALSE
        AND NOT EXISTS (SELECT 1 FROM board_game_entries e WHERE e.tournament_id=t.id)
        AND NOT EXISTS (SELECT 1 FROM score_logs sl WHERE sl.source_match_id=m.id)
), awards AS (
    SELECT id, event_id, sport_id, winner_id AS team_id, 10 AS points,
        CONCAT('tournament_win', round+1, '_points') AS reason
    FROM completed WHERE round BETWEEN 0 AND 2
    UNION ALL
    SELECT id, event_id, sport_id, winner_id,
        CASE WHEN is_bronze_match OR (round=max_round AND match_number_in_round>0) THEN 50 ELSE 80 END,
        'tournament_champion_points'
    FROM completed WHERE max_round>0 AND (is_bronze_match OR round=max_round)
    UNION ALL
    SELECT id, event_id, sport_id,
        CASE WHEN winner_id=team1_id THEN team2_id ELSE team1_id END,
        CASE WHEN is_bronze_match OR (round=max_round AND match_number_in_round>0) THEN 40 ELSE 60 END,
        'tournament_champion_points'
    FROM completed WHERE max_round>0 AND (is_bronze_match OR round=max_round)
        AND winner_id IN (team1_id, team2_id)
)
SELECT a.event_id, team.class_id, a.sport_id, a.points, a.reason, a.id
FROM awards a
JOIN teams team ON team.id=a.team_id
JOIN classes c ON c.id=team.class_id AND c.event_id=a.event_id;
