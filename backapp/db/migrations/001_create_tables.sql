-- 文字コードと照合順序を設定
SET NAMES utf8mb4;
SET time_zone = '+09:00'; -- 必要に応じて日本のタイムゾーンに設定

-- ログインを許可するメールアドレスのホワイトリスト
CREATE TABLE whitelisted_emails (
    email VARCHAR(255) PRIMARY KEY,
    role ENUM('root', 'admin', 'student') NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ロール（役割）マスタテーブル
CREATE TABLE roles (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) UNIQUE NOT NULL -- 'root', 'admin', 'student' など
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- クラス情報テーブル
CREATE TABLE classes (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) UNIQUE NOT NULL,
    student_count INT NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ユーザーテーブル
CREATE TABLE users (
    id CHAR(36) PRIMARY KEY, -- アプリケーション側でUUIDを生成
    email VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    class_id INT, -- FK
    is_profile_complete BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ユーザーとロールの中間テーブル
CREATE TABLE user_roles (
    user_id CHAR(36) NOT NULL, -- FK to users.id
    role_id INT NOT NULL,      -- FK to roles.id
    PRIMARY KEY (user_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 大会イベントテーブル
CREATE TABLE events (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    `year` INT NOT NULL,
    season ENUM('spring', 'autumn') NOT NULL,
    start_date DATE,
    end_date DATE,
    UNIQUE(`year`, season)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 競技（スポーツ）テーブル
CREATE TABLE sports (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    rules TEXT,
    location ENUM('gym1', 'gym2', 'ground', 'noon_game', 'other') NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- チームテーブル
CREATE TABLE teams (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    class_id INT NOT NULL, -- FK
    sport_id INT NOT NULL, -- FK
    event_id INT NOT NULL, -- FK
    UNIQUE(class_id, sport_id, event_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- チームメンバーの中間テーブル
CREATE TABLE team_members (
    team_id INT NOT NULL, -- FK
    user_id CHAR(36) NOT NULL, -- FK
    PRIMARY KEY (team_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- トーナメントテーブル
CREATE TABLE tournaments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    event_id INT NOT NULL, -- FK
    sport_id INT NOT NULL -- FK
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 試合テーブル
CREATE TABLE matches (
    id INT PRIMARY KEY AUTO_INCREMENT,
    tournament_id INT, -- FK
    round INT,
    match_number_in_round INT,
    team1_id INT, -- FK
    team2_id INT, -- FK
    team1_score INT,
    team2_score INT,
    winner_team_id INT, -- FK
    next_match_id INT, -- FK
    status VARCHAR(50),
    start_time TIMESTAMP NULL,
    court_number VARCHAR(255)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- クラスごとの集計得点テーブル
CREATE TABLE class_scores (
    id INT PRIMARY KEY AUTO_INCREMENT,
    event_id INT NOT NULL, -- FK
    class_id INT NOT NULL, -- FK
    initial_points INT DEFAULT 0,
    survey_points INT DEFAULT 0,
    attendance_points INT DEFAULT 0,
    gym1_win1_points INT DEFAULT 0,
    gym1_win2_points INT DEFAULT 0,
    gym1_win3_points INT DEFAULT 0,
    gym1_champion_points INT DEFAULT 0,
    gym2_win1_points INT DEFAULT 0,
    gym2_win2_points INT DEFAULT 0,
    gym2_win3_points INT DEFAULT 0,
    gym2_champion_points INT DEFAULT 0,
    ground_win1_points INT DEFAULT 0,
    ground_win2_points INT DEFAULT 0,
    ground_win3_points INT DEFAULT 0,
    ground_champion_points INT DEFAULT 0,
    noon_game_points INT DEFAULT 0,
    mvp_points INT DEFAULT 0,
    total_points_current_event INT,
    rank_current_event INT,
    total_points_overall INT,
    rank_overall INT,
    UNIQUE(event_id, class_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 得点履歴テーブル
CREATE TABLE score_logs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    event_id INT NOT NULL, -- FK
    class_id INT NOT NULL, -- FK
    points INT NOT NULL,
    reason TEXT NOT NULL,
    source_match_id INT, -- FK
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 出席チェックインテーブル
CREATE TABLE check_ins (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id CHAR(36) NOT NULL, -- FK
    event_id INT NOT NULL, -- FK
    purpose ENUM('opening_ceremony', 'event_participation') NOT NULL,
    checked_in_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- MVP投票テーブル
CREATE TABLE mvp_votes (
    id INT PRIMARY KEY AUTO_INCREMENT,
    event_id INT NOT NULL, -- FK
    voter_user_id CHAR(36) NOT NULL, -- FK
    voted_for_class_id INT NOT NULL, -- FK
    reason TEXT NOT NULL,
    points INT NOT NULL DEFAULT 3,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_id, voter_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 通知テーブル
CREATE TABLE notifications (
    id INT PRIMARY KEY AUTO_INCREMENT,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    created_by CHAR(36), -- FK
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 通知の宛先テーブル
CREATE TABLE notification_recipients (
    notification_id INT NOT NULL, -- FK
    user_id CHAR(36), -- FK
    class_id INT, -- FK
    PRIMARY KEY (notification_id, user_id, class_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- --- 外部キー制約 (Foreign Key Constraints) ---

-- ユーザーとロールの中間テーブル
ALTER TABLE user_roles ADD CONSTRAINT fk_user_roles_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE user_roles ADD CONSTRAINT fk_user_roles_role_id FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;

-- users テーブル
ALTER TABLE users ADD CONSTRAINT fk_users_class_id FOREIGN KEY (class_id) REFERENCES classes(id);

-- teams テーブル
ALTER TABLE teams ADD CONSTRAINT fk_teams_class_id FOREIGN KEY (class_id) REFERENCES classes(id);
ALTER TABLE teams ADD CONSTRAINT fk_teams_sport_id FOREIGN KEY (sport_id) REFERENCES sports(id);
ALTER TABLE teams ADD CONSTRAINT fk_teams_event_id FOREIGN KEY (event_id) REFERENCES events(id);

-- team_members テーブル
ALTER TABLE team_members ADD CONSTRAINT fk_team_members_team_id FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;
ALTER TABLE team_members ADD CONSTRAINT fk_team_members_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- tournaments テーブル
ALTER TABLE tournaments ADD CONSTRAINT fk_tournaments_event_id FOREIGN KEY (event_id) REFERENCES events(id);
ALTER TABLE tournaments ADD CONSTRAINT fk_tournaments_sport_id FOREIGN KEY (sport_id) REFERENCES sports(id);

-- matches テーブル
ALTER TABLE matches ADD CONSTRAINT fk_matches_tournament_id FOREIGN KEY (tournament_id) REFERENCES tournaments(id);
ALTER TABLE matches ADD CONSTRAINT fk_matches_team1_id FOREIGN KEY (team1_id) REFERENCES teams(id);
ALTER TABLE matches ADD CONSTRAINT fk_matches_team2_id FOREIGN KEY (team2_id) REFERENCES teams(id);
ALTER TABLE matches ADD CONSTRAINT fk_matches_winner_team_id FOREIGN KEY (winner_team_id) REFERENCES teams(id);
ALTER TABLE matches ADD CONSTRAINT fk_matches_next_match_id FOREIGN KEY (next_match_id) REFERENCES matches(id);

-- class_scores テーブル
ALTER TABLE class_scores ADD CONSTRAINT fk_class_scores_event_id FOREIGN KEY (event_id) REFERENCES events(id);
ALTER TABLE class_scores ADD CONSTRAINT fk_class_scores_class_id FOREIGN KEY (class_id) REFERENCES classes(id);

-- score_logs テーブル
ALTER TABLE score_logs ADD CONSTRAINT fk_score_logs_event_id FOREIGN KEY (event_id) REFERENCES events(id);
ALTER TABLE score_logs ADD CONSTRAINT fk_score_logs_class_id FOREIGN KEY (class_id) REFERENCES classes(id);
ALTER TABLE score_logs ADD CONSTRAINT fk_score_logs_source_match_id FOREIGN KEY (source_match_id) REFERENCES matches(id);

-- check_ins テーブル
ALTER TABLE check_ins ADD CONSTRAINT fk_check_ins_user_id FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE check_ins ADD CONSTRAINT fk_check_ins_event_id FOREIGN KEY (event_id) REFERENCES events(id);

-- mvp_votes テーブル
ALTER TABLE mvp_votes ADD CONSTRAINT fk_mvp_votes_event_id FOREIGN KEY (event_id) REFERENCES events(id);
ALTER TABLE mvp_votes ADD CONSTRAINT fk_mvp_votes_voter_user_id FOREIGN KEY (voter_user_id) REFERENCES users(id);
ALTER TABLE mvp_votes ADD CONSTRAINT fk_mvp_votes_voted_for_class_id FOREIGN KEY (voted_for_class_id) REFERENCES classes(id);

-- notifications テーブル
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_created_by FOREIGN KEY (created_by) REFERENCES users(id);

-- notification_recipients テーブル
ALTER TABLE notification_recipients ADD CONSTRAINT fk_recipients_notification_id FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE;
ALTER TABLE notification_recipients ADD CONSTRAINT fk_recipients_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE notification_recipients ADD CONSTRAINT fk_recipients_class_id FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE;