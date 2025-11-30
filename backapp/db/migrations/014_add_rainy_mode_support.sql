-- 雨天時モード対応のマイグレーション

-- eventsテーブルにis_rainy_modeカラムを追加
ALTER TABLE events
ADD COLUMN is_rainy_mode BOOLEAN NOT NULL DEFAULT FALSE;

-- 雨天時設定テーブル（試合時間設定、登録制限など）
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

-- matchesテーブルにis_loser_bracket_matchカラムを追加（敗者戦用）
ALTER TABLE matches
ADD COLUMN is_loser_bracket_match BOOLEAN NOT NULL DEFAULT FALSE;

-- matchesテーブルにloser_bracket_roundカラムを追加（敗者戦のラウンド番号）
ALTER TABLE matches
ADD COLUMN loser_bracket_round INT NULL DEFAULT NULL;

