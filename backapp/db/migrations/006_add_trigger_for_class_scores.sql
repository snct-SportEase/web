-- 既存のトリガーとプロシージャを安全に削除
DROP TRIGGER IF EXISTS update_total_points_before_insert;
DROP TRIGGER IF EXISTS update_total_points_before_update;
DROP TRIGGER IF EXISTS after_class_scores_insert;
DROP TRIGGER IF EXISTS after_class_scores_update;
DROP PROCEDURE IF EXISTS update_class_ranks;

-- 合計得点を更新するBEFOREトリガー
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

-- ランキングを更新するためのストアドプロシージャ
DELIMITER //
CREATE PROCEDURE update_class_ranks(p_event_id INT)
BEGIN
    UPDATE class_scores AS cs
    JOIN (
        SELECT
            class_id,
            RANK() OVER (ORDER BY total_points_current_event DESC) AS new_rank
        FROM class_scores
        WHERE event_id = p_event_id
    ) AS ranked_data ON cs.class_id = ranked_data.class_id
    SET cs.rank_current_event = ranked_data.new_rank
    WHERE cs.event_id = p_event_id;
END //
DELIMITER ;


-- 全体のランキングを更新するBEFOREトリガー
DELIMITER //
CREATE PROCEDURE update_class_overall_ranks(p_event_id INT)
BEGIN
    UPDATE class_scores AS cs
    JOIN (
        SELECT
            class_id,
            RANK() OVER (ORDER BY total_points_overall DESC) AS new_rank
        FROM class_scores
        WHERE event_id = p_event_id
    ) AS ranked_data ON cs.class_id = ranked_data.class_id
    SET cs.rank_overall = ranked_data.new_rank
    WHERE cs.event_id = p_event_id;
END //
DELIMITER ;