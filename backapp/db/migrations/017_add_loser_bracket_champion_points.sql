-- 敗者戦ブロック優勝の得点項目を追加

-- class_scoresテーブルにgym2_loser_bracket_champion_pointsカラムを追加
ALTER TABLE class_scores
ADD COLUMN gym2_loser_bracket_champion_points INT DEFAULT 0 AFTER gym2_champion_points;

