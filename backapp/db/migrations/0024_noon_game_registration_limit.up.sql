ALTER TABLE noon_game_sessions
    ADD COLUMN exclude_registration_limit BOOLEAN NOT NULL DEFAULT FALSE AFTER allow_manual_points;
