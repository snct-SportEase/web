DROP TABLE IF EXISTS noon_game_typing_system_imports;

ALTER TABLE noon_game_points
    MODIFY COLUMN source ENUM('result', 'manual') NOT NULL;
