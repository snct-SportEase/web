SET NAMES utf8mb4;
SET time_zone = '+09:00';

-- 結果詳細にエントリー名を保存するカラムを追加
ALTER TABLE noon_game_result_details
ADD COLUMN entry_resolved_name VARCHAR(255) DEFAULT NULL COMMENT 'エントリーの解決済み名前（結果保存時に設定）';

