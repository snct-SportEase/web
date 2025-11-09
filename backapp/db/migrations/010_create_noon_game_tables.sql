SET NAMES utf8mb4;
SET time_zone = '+09:00';

CREATE TABLE noon_game_sessions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    event_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    mode ENUM('class', 'group', 'mixed') NOT NULL DEFAULT 'mixed',
    win_points INT NOT NULL DEFAULT 0,
    loss_points INT NOT NULL DEFAULT 0,
    draw_points INT NOT NULL DEFAULT 0,
    participation_points INT NOT NULL DEFAULT 0,
    allow_manual_points BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_noon_session_event (event_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE noon_game_groups (
    id INT PRIMARY KEY AUTO_INCREMENT,
    session_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_noon_groups_session (session_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE noon_game_group_members (
    id INT PRIMARY KEY AUTO_INCREMENT,
    group_id INT NOT NULL,
    class_id INT NOT NULL,
    weight DECIMAL(6,2) NOT NULL DEFAULT 1.00,
    UNIQUE KEY uq_noon_group_member (group_id, class_id),
    INDEX idx_noon_group_members_group (group_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE noon_game_matches (
    id INT PRIMARY KEY AUTO_INCREMENT,
    session_id INT NOT NULL,
    title VARCHAR(255),
    scheduled_at DATETIME NULL,
    location VARCHAR(255),
    format VARCHAR(100),
    status ENUM('scheduled', 'in_progress', 'completed', 'cancelled') NOT NULL DEFAULT 'scheduled',
    memo TEXT,
    home_side_type ENUM('class', 'group') NOT NULL,
    home_class_id INT NULL,
    home_group_id INT NULL,
    away_side_type ENUM('class', 'group') NOT NULL,
    away_class_id INT NULL,
    away_group_id INT NULL,
    allow_draw BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_noon_matches_session (session_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE noon_game_results (
    id INT PRIMARY KEY AUTO_INCREMENT,
    match_id INT NOT NULL,
    winner ENUM('home', 'away', 'draw') NOT NULL,
    home_score INT,
    away_score INT,
    recorded_by CHAR(36) NOT NULL,
    recorded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    note TEXT,
    UNIQUE KEY uq_noon_match_result (match_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE noon_game_points (
    id INT PRIMARY KEY AUTO_INCREMENT,
    session_id INT NOT NULL,
    match_id INT NULL,
    class_id INT NOT NULL,
    points INT NOT NULL,
    reason VARCHAR(255),
    source ENUM('result', 'manual') NOT NULL,
    created_by CHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_noon_points_session_class (session_id, class_id),
    INDEX idx_noon_points_match (match_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

