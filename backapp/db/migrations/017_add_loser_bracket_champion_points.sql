-- 敗者戦ブロック優勝の得点項目を追加

-- class_scoresテーブルにgym2_loser_bracket_champion_pointsカラムを追加
ALTER TABLE class_scores
ADD COLUMN gym2_loser_bracket_champion_points INT DEFAULT 0 AFTER gym2_champion_points;

-- トリガーを更新してgym2_loser_bracket_champion_pointsを含める
DROP TRIGGER IF EXISTS update_total_points_before_insert;
DROP TRIGGER IF EXISTS update_total_points_before_update;

DELIMITER //
CREATE TRIGGER update_total_points_before_insert
BEFORE INSERT ON class_scores
FOR EACH ROW
BEGIN
    -- initial_points を除く現在のイベントの合計得点を計算
    SET NEW.total_points_current_event =
        IFNULL(NEW.survey_points, 0) +
        IFNULL(NEW.attendance_points, 0) +
        IFNULL(NEW.gym1_win1_points, 0) +
        IFNULL(NEW.gym1_win2_points, 0) +
        IFNULL(NEW.gym1_win3_points, 0) +
        IFNULL(NEW.gym1_champion_points, 0) +
        IFNULL(NEW.gym2_win1_points, 0) +
        IFNULL(NEW.gym2_win2_points, 0) +
        IFNULL(NEW.gym2_win3_points, 0) +
        IFNULL(NEW.gym2_champion_points, 0) +
        IFNULL(NEW.gym2_loser_bracket_champion_points, 0) +
        IFNULL(NEW.ground_win1_points, 0) +
        IFNULL(NEW.ground_win2_points, 0) +
        IFNULL(NEW.ground_win3_points, 0) +
        IFNULL(NEW.ground_champion_points, 0) +
        IFNULL(NEW.noon_game_points, 0) +
        IFNULL(NEW.mvp_points, 0);

    -- initial_points を含む総合計得点を計算
    SET NEW.total_points_overall = NEW.total_points_current_event + IFNULL(NEW.initial_points, 0);
END;
//
DELIMITER ;

DELIMITER //
CREATE TRIGGER update_total_points_before_update
BEFORE UPDATE ON class_scores
FOR EACH ROW
BEGIN
    -- initial_points を除く現在のイベントの合計得点を計算
    SET NEW.total_points_current_event =
        IFNULL(NEW.survey_points, 0) +
        IFNULL(NEW.attendance_points, 0) +
        IFNULL(NEW.gym1_win1_points, 0) +
        IFNULL(NEW.gym1_win2_points, 0) +
        IFNULL(NEW.gym1_win3_points, 0) +
        IFNULL(NEW.gym1_champion_points, 0) +
        IFNULL(NEW.gym2_win1_points, 0) +
        IFNULL(NEW.gym2_win2_points, 0) +
        IFNULL(NEW.gym2_win3_points, 0) +
        IFNULL(NEW.gym2_champion_points, 0) +
        IFNULL(NEW.gym2_loser_bracket_champion_points, 0) +
        IFNULL(NEW.ground_win1_points, 0) +
        IFNULL(NEW.ground_win2_points, 0) +
        IFNULL(NEW.ground_win3_points, 0) +
        IFNULL(NEW.ground_champion_points, 0) +
        IFNULL(NEW.noon_game_points, 0) +
        IFNULL(NEW.mvp_points, 0);

    -- initial_points を含む総合計得点を計算
    SET NEW.total_points_overall = NEW.total_points_current_event + IFNULL(NEW.initial_points, 0);
END;
//
DELIMITER ;

