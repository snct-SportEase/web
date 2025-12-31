SET NAMES utf8mb4;
SET time_zone = '+09:00';

-- 昼競技テンプレートの「実行単位(run)」を管理する
CREATE TABLE noon_game_template_runs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    session_id INT NOT NULL,
    template_key VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_by CHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_noon_template_runs_session (session_id),
    INDEX idx_noon_template_runs_template (template_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- run に紐づく試合（Aブロック/Bブロック/総合ボーナス等）を紐づける
CREATE TABLE noon_game_template_run_matches (
    id INT PRIMARY KEY AUTO_INCREMENT,
    run_id INT NOT NULL,
    match_id INT NOT NULL,
    match_key VARCHAR(50) NOT NULL,
    UNIQUE KEY uq_noon_run_match_key (run_id, match_key),
    UNIQUE KEY uq_noon_run_match_match (match_id),
    INDEX idx_noon_run_match_run (run_id),
    INDEX idx_noon_run_match_match_key (match_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


