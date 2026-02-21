-- 文字コードと照合順序を設定
SET NAMES utf8mb4;
SET time_zone = '+09:00'; -- 必要に応じて日本のタイムゾーンに設定

-- ログインを許可するメールアドレスのホワイトリスト
CREATE TABLE whitelisted_emails (
    email VARCHAR(255) NOT NULL,
    role ENUM('root', 'admin', 'student') NOT NULL,
    event_id INT NULL,
    UNIQUE KEY uk_email_event (email, event_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ロール（役割）マスタテーブル
CREATE TABLE roles (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) UNIQUE NOT NULL -- 'root', 'admin', 'student' など
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO roles (name) VALUES ('root'), ('admin'), ('student');

-- 大会イベントテーブル
CREATE TABLE events (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    `year` INT NOT NULL,
    season ENUM('spring', 'autumn') NOT NULL,
    start_date DATE,
    end_date DATE,
    competition_guidelines_pdf_url VARCHAR(500) NULL,
    survey_url VARCHAR(500) NULL COMMENT 'アンケートURL',
    is_survey_published BOOLEAN NOT NULL DEFAULT FALSE COMMENT 'アンケートが通知済みかどうか',
    is_rainy_mode BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(`year`, season)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE active_event (
    id INT PRIMARY KEY AUTO_INCREMENT,
    event_id INT UNIQUE,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- クラス情報テーブル (3NF対応済)
CREATE TABLE classes (
    id INT PRIMARY KEY AUTO_INCREMENT,
    event_id INT NOT NULL, 
    name VARCHAR(255) NOT NULL,
    student_count INT NOT NULL DEFAULT 0,
    attend_count INT NOT NULL DEFAULT 0,
    UNIQUE KEY uk_name_event (name, event_id),
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ユーザーテーブル
CREATE TABLE users (
    id CHAR(36) PRIMARY KEY, 
    email VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    class_id INT, -- FK
    is_profile_complete BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (class_id) REFERENCES classes(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ユーザーとロールの中間テーブル
CREATE TABLE user_roles (
    user_id CHAR(36) NOT NULL, 
    role_id INT NOT NULL,      
    event_id INT NULL,         
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 競技（スポーツ）テーブル
CREATE TABLE sports (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- events_sportsテーブル
CREATE TABLE event_sports (
    event_id INT NOT NULL,  
    sport_id INT NOT NULL,  
    description TEXT,
    rules TEXT,
    location ENUM('gym1', 'gym2', 'ground', 'noon_game', 'other') NOT NULL,
    rules_type ENUM('markdown', 'pdf') NOT NULL DEFAULT 'markdown',
    rules_pdf_url VARCHAR(255) NULL,
    min_capacity INT NULL DEFAULT NULL,
    max_capacity INT NULL DEFAULT NULL,
    PRIMARY KEY (event_id, sport_id),
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE RESTRICT,
    FOREIGN KEY (sport_id) REFERENCES sports(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 雨天時設定テーブル
CREATE TABLE rainy_mode_settings (
    id INT PRIMARY KEY AUTO_INCREMENT,
    event_id INT NOT NULL,
    sport_id INT NOT NULL,
    class_id INT NOT NULL,
    min_capacity INT NULL,
    max_capacity INT NULL,
    match_start_time VARCHAR(255) NULL,
    UNIQUE(event_id, sport_id, class_id),
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    FOREIGN KEY (sport_id) REFERENCES sports(id) ON DELETE CASCADE,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- チームテーブル (3NF対応済：event_idを削除し正規化)
CREATE TABLE teams (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    class_id INT NOT NULL, 
    sport_id INT NOT NULL, 
    min_capacity INT NULL DEFAULT NULL,
    max_capacity INT NULL DEFAULT NULL,
    UNIQUE(class_id, sport_id),
    FOREIGN KEY (class_id) REFERENCES classes(id),
    FOREIGN KEY (sport_id) REFERENCES sports(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- チームメンバーの中間テーブル
CREATE TABLE team_members (
    team_id INT NOT NULL, 
    user_id CHAR(36) NOT NULL, 
    is_confirmed BOOLEAN NOT NULL DEFAULT false,
    PRIMARY KEY (team_id, user_id),
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- トーナメントテーブル
CREATE TABLE tournaments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    event_id INT NOT NULL, 
    sport_id INT NOT NULL, 
    FOREIGN KEY (event_id, sport_id) REFERENCES event_sports(event_id, sport_id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 試合テーブル (3NF対応済：winner_team_id, start_timeの重複を排除)
CREATE TABLE matches (
    id INT PRIMARY KEY AUTO_INCREMENT,
    tournament_id INT, 
    round INT,
    match_number_in_round INT,
    match_start_time TIMESTAMP NULL,
    match_end_time TIMESTAMP NULL,
    team1_id INT, 
    team2_id INT, 
    team1_score INT,
    team2_score INT,
    next_match_id INT, 
    status VARCHAR(50),
    court_number VARCHAR(255),
    is_bronze_match BOOLEAN NOT NULL DEFAULT FALSE,
    is_loser_bracket_match BOOLEAN NOT NULL DEFAULT FALSE,
    loser_bracket_round INT NULL DEFAULT NULL,
    rainy_mode_start_time VARCHAR(255) NULL,
    loser_bracket_block VARCHAR(1) NULL DEFAULT NULL COMMENT '敗者戦ブロック識別: A or B',
    FOREIGN KEY (tournament_id) REFERENCES tournaments(id),
    FOREIGN KEY (team1_id) REFERENCES teams(id),
    FOREIGN KEY (team2_id) REFERENCES teams(id),
    FOREIGN KEY (next_match_id) REFERENCES matches(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 得点履歴テーブル (得点の正規化：スコアの真実の源源)
CREATE TABLE score_logs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    event_id INT NOT NULL, 
    class_id INT NOT NULL, 
    points INT NOT NULL,
    reason TEXT NOT NULL,
    source_match_id INT, 
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (event_id) REFERENCES events(id),
    FOREIGN KEY (class_id) REFERENCES classes(id),
    FOREIGN KEY (source_match_id) REFERENCES matches(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 出席チェックインテーブル
CREATE TABLE check_ins (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id CHAR(36) NOT NULL, 
    event_id INT NOT NULL, 
    purpose ENUM('opening_ceremony', 'event_participation') NOT NULL,
    checked_in_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (event_id) REFERENCES events(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- MVP投票テーブル
CREATE TABLE mvp_votes (
    id INT PRIMARY KEY AUTO_INCREMENT,
    event_id INT NOT NULL, 
    voter_user_id CHAR(36) NOT NULL, 
    voted_for_class_id INT NOT NULL, 
    reason TEXT NOT NULL,
    points INT NOT NULL DEFAULT 3,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_id, voter_user_id),
    FOREIGN KEY (event_id) REFERENCES events(id),
    FOREIGN KEY (voter_user_id) REFERENCES users(id),
    FOREIGN KEY (voted_for_class_id) REFERENCES classes(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 通知テーブル
CREATE TABLE notifications (
    id INT PRIMARY KEY AUTO_INCREMENT,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    created_by CHAR(36), 
    event_id INT NULL, 
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 通知の宛先テーブル
CREATE TABLE notification_recipients (
    notification_id INT NOT NULL, 
    user_id CHAR(36), 
    class_id INT, 
    PRIMARY KEY (notification_id, user_id, class_id),
    FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 通知ターゲットテーブル
CREATE TABLE notification_targets (
    notification_id INT NOT NULL,
    role_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (notification_id, role_name),
    FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
    FOREIGN KEY (role_name) REFERENCES roles(name) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE push_subscriptions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id CHAR(36) NOT NULL,
    endpoint VARCHAR(500) NOT NULL,
    auth_key VARCHAR(255) NOT NULL,
    p256dh_key VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_push_subscriptions_endpoint (endpoint),
    KEY idx_push_subscriptions_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE notification_requests (
    id INT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    target_text TEXT NOT NULL,
    status ENUM('pending', 'approved', 'rejected') NOT NULL DEFAULT 'pending',
    requester_id CHAR(36) NOT NULL,
    resolved_by CHAR(36) NULL,
    resolved_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (requester_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (resolved_by) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_notification_requests_status (status),
    INDEX idx_notification_requests_requester (requester_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE notification_request_messages (
    id INT PRIMARY KEY AUTO_INCREMENT,
    request_id INT NOT NULL,
    sender_id CHAR(36) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (request_id) REFERENCES notification_requests(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_notification_request_messages_request (request_id),
    INDEX idx_notification_request_messages_sender (sender_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
    entry_resolved_name VARCHAR(255) DEFAULT NULL COMMENT 'エントリーの解決済み名前',
    UNIQUE KEY uq_noon_result_details (result_id, entry_id),
    INDEX idx_noon_result_details_result (result_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE noon_game_template_runs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    session_id INT NOT NULL,
    template_key VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    points_by_rank JSON DEFAULT NULL COMMENT '順位ごとの点数設定',
    created_by CHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_noon_template_runs_session (session_id),
    INDEX idx_noon_template_runs_template (template_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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

CREATE TABLE noon_game_template_default_groups (
    id INT PRIMARY KEY AUTO_INCREMENT,
    template_key VARCHAR(50) NOT NULL COMMENT 'テンプレートキー',
    group_index INT NOT NULL COMMENT 'グループの順序',
    group_name VARCHAR(255) NOT NULL COMMENT 'グループ名',
    class_names JSON NOT NULL COMMENT 'クラス名のリスト',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_template_group_index (template_key, group_index),
    INDEX idx_template_key (template_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='昼競技テンプレートのデフォルトグループ設定';

-- デフォルトデータを挿入
INSERT INTO noon_game_template_default_groups (template_key, group_index, group_name, class_names) VALUES
('year_relay', 1, '1年生', JSON_ARRAY('1-1', '1-2', '1-3')),
('year_relay', 2, '2年生', JSON_ARRAY('IS2', 'IT2', 'IE2')),
('year_relay', 3, '3年生', JSON_ARRAY('IS3', 'IT3', 'IE3')),
('year_relay', 4, '4年生', JSON_ARRAY('IS4', 'IT4', 'IE4')),
('year_relay', 5, '5年生', JSON_ARRAY('IS5', 'IT5', 'IE5')),
('year_relay', 6, '専教', JSON_ARRAY('専教')),
('course_relay', 1, '1-1 & IEコース', JSON_ARRAY('1-1', 'IE2', 'IE3', 'IE4', 'IE5')),
('course_relay', 2, '1-2 & ISコース', JSON_ARRAY('1-2', 'IS2', 'IS3', 'IS4', 'IS5')),
('course_relay', 3, '1-3 & ITコース', JSON_ARRAY('1-3', 'IT2', 'IT3', 'IT4', 'IT5')),
('course_relay', 4, '専攻科・教員', JSON_ARRAY('専教')),
('tug_of_war', 1, '1-1 & ISコース', JSON_ARRAY('1-1', 'IS2', 'IS3', 'IS4', 'IS5')),
('tug_of_war', 2, '1-2 & ITコース', JSON_ARRAY('1-2', 'IT2', 'IT3', 'IT4', 'IT5')),
('tug_of_war', 3, '1-3 & IEコース', JSON_ARRAY('1-3', 'IE2', 'IE3', 'IE4', 'IE5')),
('tug_of_war', 4, '専攻科・教員', JSON_ARRAY('専教'));

-- 第3正規形に対応した class_scores ビューの作成
CREATE VIEW class_scores AS
WITH aggregated_scores AS (
    SELECT 
        c.id AS class_id,
        c.event_id AS event_id,
        COALESCE(SUM(CASE WHEN sl.reason = 'initial_points' THEN sl.points ELSE 0 END), 0) AS initial_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'survey_points' THEN sl.points ELSE 0 END), 0) AS survey_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'attendance_points' THEN sl.points ELSE 0 END), 0) AS attendance_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'gym1_win1_points' THEN sl.points ELSE 0 END), 0) AS gym1_win1_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'gym1_win2_points' THEN sl.points ELSE 0 END), 0) AS gym1_win2_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'gym1_win3_points' THEN sl.points ELSE 0 END), 0) AS gym1_win3_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'gym1_champion_points' THEN sl.points ELSE 0 END), 0) AS gym1_champion_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'gym2_win1_points' THEN sl.points ELSE 0 END), 0) AS gym2_win1_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'gym2_win2_points' THEN sl.points ELSE 0 END), 0) AS gym2_win2_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'gym2_win3_points' THEN sl.points ELSE 0 END), 0) AS gym2_win3_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'gym2_champion_points' THEN sl.points ELSE 0 END), 0) AS gym2_champion_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'gym2_loser_bracket_champion_points' THEN sl.points ELSE 0 END), 0) AS gym2_loser_bracket_champion_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'ground_win1_points' THEN sl.points ELSE 0 END), 0) AS ground_win1_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'ground_win2_points' THEN sl.points ELSE 0 END), 0) AS ground_win2_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'ground_win3_points' THEN sl.points ELSE 0 END), 0) AS ground_win3_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'ground_champion_points' THEN sl.points ELSE 0 END), 0) AS ground_champion_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'noon_game_points' THEN sl.points ELSE 0 END), 0) AS noon_game_points,
        COALESCE(SUM(CASE WHEN sl.reason = 'mvp_points' THEN sl.points ELSE 0 END), 0) AS mvp_points,
        COALESCE(SUM(CASE WHEN sl.reason != 'initial_points' THEN sl.points ELSE 0 END), 0) AS total_points_current_event,
        COALESCE(SUM(sl.points), 0) AS total_points_overall
    FROM classes c
    LEFT JOIN score_logs sl ON c.id = sl.class_id
    GROUP BY c.id, c.event_id
)
SELECT 
    class_id AS id, -- Goモデル用のエイリアス
    class_id,
    event_id,
    initial_points,
    survey_points,
    attendance_points,
    gym1_win1_points,
    gym1_win2_points,
    gym1_win3_points,
    gym1_champion_points,
    gym2_win1_points,
    gym2_win2_points,
    gym2_win3_points,
    gym2_champion_points,
    gym2_loser_bracket_champion_points,
    ground_win1_points,
    ground_win2_points,
    ground_win3_points,
    ground_champion_points,
    noon_game_points,
    mvp_points,
    total_points_current_event,
    IF(total_points_current_event = 0, 0, RANK() OVER (PARTITION BY event_id ORDER BY total_points_current_event DESC)) AS rank_current_event,
    total_points_overall,
    IF(total_points_overall = 0, 0, RANK() OVER (PARTITION BY event_id ORDER BY total_points_overall DESC)) AS rank_overall
FROM aggregated_scores;
