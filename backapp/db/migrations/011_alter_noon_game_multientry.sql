SET NAMES utf8mb4;
SET time_zone = '+09:00';

CREATE TABLE noon_game_match_entries (
    id INT PRIMARY KEY AUTO_INCREMENT,
    match_id INT NOT NULL,
    entry_index INT NOT NULL,
    side_type ENUM('class', 'group') NOT NULL,
    class_id INT NULL,
    group_id INT NULL,
    display_name VARCHAR(255),
    UNIQUE KEY uq_noon_match_entries (match_id, entry_index),
    INDEX idx_noon_match_entries_match (match_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE noon_game_result_details (
    id INT PRIMARY KEY AUTO_INCREMENT,
    result_id INT NOT NULL,
    entry_id INT NOT NULL,
    placement_rank INT NULL,
    points INT NOT NULL DEFAULT 0,
    note VARCHAR(255),
    UNIQUE KEY uq_noon_result_details (result_id, entry_id),
    INDEX idx_noon_result_details_result (result_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

