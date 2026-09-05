ALTER TABLE matches
    ADD COLUMN winner_team_id INT NULL AFTER team2_score,
    ADD INDEX idx_matches_winner_team (winner_team_id),
    ADD CONSTRAINT fk_matches_winner_team FOREIGN KEY (winner_team_id)
        REFERENCES teams(id) ON DELETE SET NULL;

UPDATE matches
SET winner_team_id = CASE
    WHEN team1_score > team2_score THEN team1_id
    WHEN team2_score > team1_score THEN team2_id
    ELSE NULL
END
WHERE status = 'finished';

UPDATE matches current_match
JOIN matches next_match ON next_match.id = current_match.next_match_id
SET current_match.winner_team_id = CASE
    WHEN current_match.team1_id IN (next_match.team1_id, next_match.team2_id) THEN current_match.team1_id
    WHEN current_match.team2_id IN (next_match.team1_id, next_match.team2_id) THEN current_match.team2_id
    ELSE NULL
END
WHERE current_match.status = 'finished'
  AND current_match.winner_team_id IS NULL;

UPDATE matches m
JOIN score_logs sl
  ON sl.source_match_id = m.id
 AND sl.reason = 'board_game_win_points'
JOIN teams winner_team
  ON winner_team.class_id = sl.class_id
 AND winner_team.id IN (m.team1_id, m.team2_id)
SET m.winner_team_id = winner_team.id
WHERE m.status = 'finished'
  AND m.winner_team_id IS NULL;
