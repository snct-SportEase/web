ALTER TABLE matches
    DROP FOREIGN KEY fk_matches_winner_team,
    DROP INDEX idx_matches_winner_team,
    DROP COLUMN winner_team_id;
