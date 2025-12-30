SET NAMES utf8mb4;
SET time_zone = '+09:00';

-- テンプレートランに点数設定を追加（コース対抗リレーと綱引き用）
ALTER TABLE noon_game_template_runs
ADD COLUMN points_by_rank JSON DEFAULT NULL COMMENT '順位ごとの点数設定（例: {"1": 40, "2": 30, "3": 20, "4": 10}）';

