-- 開発・検証環境専用のデモデータ。
-- docker compose --profile demo run --rm demo-data で投入する。
-- 再実行時は、デモ大会の進行状況をこのファイルの状態に戻す。

SET NAMES utf8mb4;
SET time_zone = '+09:00';

START TRANSACTION;

-- 大会
INSERT INTO events (
    name, `year`, season, start_date, end_date,
    competition_guidelines_pdf_url, status, hide_scores
) VALUES (
    'デモ体育大会', 2037, 'spring', '2037-05-20', '2037-05-21',
    '/uploads/demo/root.pdf', 'active', FALSE
) ON DUPLICATE KEY UPDATE
    name = VALUES(name),
    start_date = VALUES(start_date),
    end_date = VALUES(end_date),
    competition_guidelines_pdf_url = VALUES(competition_guidelines_pdf_url),
    status = VALUES(status),
    hide_scores = VALUES(hide_scores);

SET @demo_event_id = (
    SELECT id FROM events WHERE `year` = 2037 AND season = 'spring'
);

UPDATE events SET status = 'archived'
WHERE id <> @demo_event_id AND status = 'active';
INSERT INTO active_event (id, event_id) VALUES (1, @demo_event_id)
ON DUPLICATE KEY UPDATE event_id = VALUES(event_id);

-- クラス
INSERT INTO classes (event_id, name, student_count, attend_count) VALUES
(@demo_event_id, '1-1', 40, 38),
(@demo_event_id, '1-2', 40, 37),
(@demo_event_id, '1-3', 40, 39),
(@demo_event_id, 'IS2', 40, 36),
(@demo_event_id, 'IT2', 40, 35),
(@demo_event_id, 'IE2', 40, 38),
(@demo_event_id, 'IS3', 40, 37),
(@demo_event_id, 'IT3', 40, 39),
(@demo_event_id, 'IE3', 40, 36),
(@demo_event_id, 'IS4', 40, 38),
(@demo_event_id, 'IT4', 40, 37),
(@demo_event_id, 'IE4', 40, 39),
(@demo_event_id, 'IS5', 40, 36),
(@demo_event_id, 'IT5', 40, 35),
(@demo_event_id, 'IE5', 40, 38),
(@demo_event_id, '専教', 40, 10)
ON DUPLICATE KEY UPDATE
    student_count = VALUES(student_count),
    attend_count = VALUES(attend_count);

-- クラス名とメールアドレス用slugの対応表。
CREATE TEMPORARY TABLE demo_class_map (
    class_name VARCHAR(255) PRIMARY KEY,
    slug VARCHAR(32) NOT NULL
);
INSERT INTO demo_class_map (class_name, slug) VALUES
('1-1', '1-1'), ('1-2', '1-2'), ('1-3', '1-3'),
('IS2', 'is2'), ('IT2', 'it2'), ('IE2', 'ie2'),
('IS3', 'is3'), ('IT3', 'it3'), ('IE3', 'ie3'),
('IS4', 'is4'), ('IT4', 'it4'), ('IE4', 'ie4'),
('IS5', 'is5'), ('IT5', 'it5'), ('IE5', 'ie5'),
('専教', 'advanced');

-- デモユーザー。ログインは引き続き Google OAuth を使用する。
INSERT INTO users (id, email, display_name, class_id, is_profile_complete) VALUES
('00000000-0000-4000-8000-000000000000', 'demo-root@example.com', 'デモシステム管理者', NULL, TRUE),
('00000000-0000-4000-8000-000000000001', 'demo-admin@example.com', 'デモ管理者', NULL, TRUE),
('00000000-0000-4000-8000-000000000101', 'demo-1-1-a@example.com', '1-1 デモ生徒A', (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = '1-1'), TRUE),
('00000000-0000-4000-8000-000000000102', 'demo-1-2-a@example.com', '1-2 デモ生徒A', (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = '1-2'), TRUE),
('00000000-0000-4000-8000-000000000103', 'demo-1-3-a@example.com', '1-3 デモ生徒A', (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = '1-3'), TRUE),
('00000000-0000-4000-8000-000000000104', 'demo-is2-a@example.com', 'IS2 デモ生徒A', (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = 'IS2'), TRUE),
('00000000-0000-4000-8000-000000000105', 'demo-it2-a@example.com', 'IT2 デモ生徒A', (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = 'IT2'), TRUE),
('00000000-0000-4000-8000-000000000106', 'demo-ie2-a@example.com', 'IE2 デモ生徒A', (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = 'IE2'), TRUE),
('00000000-0000-4000-8000-000000000107', 'demo-is3-a@example.com', 'IS3 デモ生徒A', (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = 'IS3'), TRUE),
('00000000-0000-4000-8000-000000000108', 'demo-it3-a@example.com', 'IT3 デモ生徒A', (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = 'IT3'), TRUE)
ON DUPLICATE KEY UPDATE
    display_name = VALUES(display_name),
    class_id = VALUES(class_id),
    is_profile_complete = VALUES(is_profile_complete);

-- 各クラス8名。既存のaユーザーは更新し、残りを追加する。
INSERT INTO users (
    id, email, display_name, class_id, notification_filters, is_profile_complete
)
SELECT
    UUID(),
    CONCAT('demo-', dcm.slug, '-', member.code, '@example.com'),
    CONCAT(c.name, ' デモ生徒', UPPER(member.code)),
    c.id,
    JSON_ARRAY('general', 'match_my_class', 'finals', 'all_matches'),
    TRUE
FROM classes c
JOIN demo_class_map dcm ON dcm.class_name = c.name
CROSS JOIN (
    SELECT 'a' AS code UNION ALL SELECT 'b' UNION ALL SELECT 'c' UNION ALL SELECT 'd'
    UNION ALL SELECT 'e' UNION ALL SELECT 'f' UNION ALL SELECT 'g' UNION ALL SELECT 'h'
) member
WHERE c.event_id = @demo_event_id
ON DUPLICATE KEY UPDATE
    display_name = VALUES(display_name),
    class_id = VALUES(class_id),
    notification_filters = VALUES(notification_filters),
    is_profile_complete = VALUES(is_profile_complete);

INSERT INTO user_roles (user_id, role_id, event_id)
SELECT '00000000-0000-4000-8000-000000000000', id, @demo_event_id
FROM roles WHERE name = 'root'
ON DUPLICATE KEY UPDATE event_id = VALUES(event_id);

INSERT INTO user_roles (user_id, role_id, event_id)
SELECT '00000000-0000-4000-8000-000000000001', id, @demo_event_id
FROM roles WHERE name = 'admin'
ON DUPLICATE KEY UPDATE event_id = VALUES(event_id);

-- a〜hの全デモ生徒へstudent権限を付与する。
INSERT INTO user_roles (user_id, role_id, event_id)
SELECT u.id, r.id, @demo_event_id
FROM users u
JOIN roles r ON r.name = 'student'
JOIN classes c ON c.id = u.class_id AND c.event_id = @demo_event_id
WHERE u.email LIKE 'demo-%@example.com'
ON DUPLICATE KEY UPDATE event_id = VALUES(event_id);

-- 競技
INSERT INTO sports (name)
SELECT 'デモバスケットボール'
WHERE NOT EXISTS (SELECT 1 FROM sports WHERE name = 'デモバスケットボール');
INSERT INTO sports (name)
SELECT 'デモバレーボール'
WHERE NOT EXISTS (SELECT 1 FROM sports WHERE name = 'デモバレーボール');
INSERT INTO sports (name)
SELECT 'デモサッカー'
WHERE NOT EXISTS (SELECT 1 FROM sports WHERE name = 'デモサッカー');
INSERT INTO sports (name)
SELECT '学年対抗リレー'
WHERE NOT EXISTS (SELECT 1 FROM sports WHERE name = '学年対抗リレー');

SET @basketball_id = (SELECT MIN(id) FROM sports WHERE name = 'デモバスケットボール');
SET @volleyball_id = (SELECT MIN(id) FROM sports WHERE name = 'デモバレーボール');
SET @soccer_id = (SELECT MIN(id) FROM sports WHERE name = 'デモサッカー');
SET @noon_sport_id = (SELECT MIN(id) FROM sports WHERE name = '学年対抗リレー');

INSERT INTO event_sports (
    event_id, sport_id, description, location, min_capacity, max_capacity
) VALUES
(@demo_event_id, @basketball_id, '進行中の試合を含む8チーム制デモ競技', 'gym1', 5, 10),
(@demo_event_id, @volleyball_id, '開始前の4チーム制デモ競技', 'gym2', 6, 12),
(@demo_event_id, @soccer_id, 'チーム編成・参加者管理確認用の屋外競技', 'ground', 8, 15),
(@demo_event_id, @noon_sport_id, '昼競技のグループ・結果登録確認用', 'noon_game', 1, 40)
ON DUPLICATE KEY UPDATE
    description = VALUES(description),
    location = VALUES(location),
    min_capacity = VALUES(min_capacity),
    max_capacity = VALUES(max_capacity);

-- 全クラスに通常競技3種のチームを作る。
INSERT INTO teams (name, class_id, sport_id, min_capacity, max_capacity)
SELECT CONCAT(c.name, ' バスケ'), c.id, @basketball_id, 5, 10
FROM classes c WHERE c.event_id = @demo_event_id
ON DUPLICATE KEY UPDATE
    name = VALUES(name), min_capacity = VALUES(min_capacity), max_capacity = VALUES(max_capacity);

INSERT INTO teams (name, class_id, sport_id, min_capacity, max_capacity)
SELECT CONCAT(c.name, ' バレー'), c.id, @volleyball_id, 6, 12
FROM classes c
WHERE c.event_id = @demo_event_id
ON DUPLICATE KEY UPDATE
    name = VALUES(name), min_capacity = VALUES(min_capacity), max_capacity = VALUES(max_capacity);

INSERT INTO teams (name, class_id, sport_id, min_capacity, max_capacity)
SELECT CONCAT(c.name, ' サッカー'), c.id, @soccer_id, 8, 15
FROM classes c WHERE c.event_id = @demo_event_id
ON DUPLICATE KEY UPDATE
    name = VALUES(name), min_capacity = VALUES(min_capacity), max_capacity = VALUES(max_capacity);

INSERT INTO team_members (team_id, user_id, is_confirmed)
SELECT t.id, u.id, TRUE
FROM teams t
JOIN classes c ON c.id = t.class_id
JOIN users u ON u.class_id = c.id AND u.email LIKE 'demo-%@example.com'
WHERE c.event_id = @demo_event_id
ON DUPLICATE KEY UPDATE is_confirmed = VALUES(is_confirmed);

-- 雨天時の人数・開始時刻設定。画面で通常値との差分を確認できる。
INSERT INTO rainy_mode_settings (
    event_id, sport_id, class_id, min_capacity, max_capacity, match_start_time
)
SELECT @demo_event_id, @basketball_id, c.id, 4, 8,
       CONCAT('2037-05-20 ', LPAD(9 + MOD(c.id, 6), 2, '0'), ':00')
FROM classes c WHERE c.event_id = @demo_event_id
ON DUPLICATE KEY UPDATE
    min_capacity = VALUES(min_capacity),
    max_capacity = VALUES(max_capacity),
    match_start_time = VALUES(match_start_time);

-- 再実行時はデモ用対戦表だけを初期状態へ戻す。
UPDATE matches m
JOIN tournaments t ON t.id = m.tournament_id
SET m.next_match_id = NULL
WHERE t.event_id = @demo_event_id
  AND t.name IN ('デモバスケットボール Tournament', 'デモバレーボール Tournament');

DELETE sl FROM score_logs sl
JOIN matches m ON m.id = sl.source_match_id
JOIN tournaments t ON t.id = m.tournament_id
WHERE t.event_id = @demo_event_id
  AND t.name IN ('デモバスケットボール Tournament', 'デモバレーボール Tournament');

DELETE m FROM matches m
JOIN tournaments t ON t.id = m.tournament_id
WHERE t.event_id = @demo_event_id
  AND t.name IN ('デモバスケットボール Tournament', 'デモバレーボール Tournament');

DELETE FROM tournaments
WHERE event_id = @demo_event_id
  AND name IN ('デモバスケットボール Tournament', 'デモバレーボール Tournament');

INSERT INTO tournaments (name, event_id, sport_id) VALUES
('デモバスケットボール Tournament', @demo_event_id, @basketball_id),
('デモバレーボール Tournament', @demo_event_id, @volleyball_id);

SET @basketball_tournament_id = (
    SELECT id FROM tournaments
    WHERE event_id = @demo_event_id AND name = 'デモバスケットボール Tournament'
);
SET @volleyball_tournament_id = (
    SELECT id FROM tournaments
    WHERE event_id = @demo_event_id AND name = 'デモバレーボール Tournament'
);

-- バスケットボール: 準々決勝の一部が終了し、準決勝へ進んだ状態。
INSERT INTO matches (
    tournament_id, round, match_number_in_round, match_start_time,
    team1_id, team2_id, team1_score, team2_score, status, court_number
) VALUES
(@basketball_tournament_id, 0, 0, '2037-05-20 09:00:00',
 (SELECT t.id FROM teams t JOIN classes c ON c.id = t.class_id WHERE c.event_id = @demo_event_id AND c.name = '1-1' AND t.sport_id = @basketball_id),
 (SELECT t.id FROM teams t JOIN classes c ON c.id = t.class_id WHERE c.event_id = @demo_event_id AND c.name = '1-2' AND t.sport_id = @basketball_id), 12, 8, 'completed', 'A'),
(@basketball_tournament_id, 0, 1, '2037-05-20 09:30:00',
 (SELECT t.id FROM teams t JOIN classes c ON c.id = t.class_id WHERE c.event_id = @demo_event_id AND c.name = '1-3' AND t.sport_id = @basketball_id),
 (SELECT t.id FROM teams t JOIN classes c ON c.id = t.class_id WHERE c.event_id = @demo_event_id AND c.name = 'IS2' AND t.sport_id = @basketball_id), 5, 9, 'completed', 'A'),
(@basketball_tournament_id, 0, 2, '2037-05-20 10:00:00',
 (SELECT t.id FROM teams t JOIN classes c ON c.id = t.class_id WHERE c.event_id = @demo_event_id AND c.name = 'IT2' AND t.sport_id = @basketball_id),
 (SELECT t.id FROM teams t JOIN classes c ON c.id = t.class_id WHERE c.event_id = @demo_event_id AND c.name = 'IE2' AND t.sport_id = @basketball_id), NULL, NULL, 'scheduled', 'A'),
(@basketball_tournament_id, 0, 3, '2037-05-20 10:30:00',
 (SELECT t.id FROM teams t JOIN classes c ON c.id = t.class_id WHERE c.event_id = @demo_event_id AND c.name = 'IS3' AND t.sport_id = @basketball_id),
 (SELECT t.id FROM teams t JOIN classes c ON c.id = t.class_id WHERE c.event_id = @demo_event_id AND c.name = 'IT3' AND t.sport_id = @basketball_id), NULL, NULL, 'scheduled', 'A'),
(@basketball_tournament_id, 1, 0, '2037-05-20 13:00:00',
 (SELECT t.id FROM teams t JOIN classes c ON c.id = t.class_id WHERE c.event_id = @demo_event_id AND c.name = '1-1' AND t.sport_id = @basketball_id),
 (SELECT t.id FROM teams t JOIN classes c ON c.id = t.class_id WHERE c.event_id = @demo_event_id AND c.name = 'IS2' AND t.sport_id = @basketball_id), NULL, NULL, 'scheduled', 'A'),
(@basketball_tournament_id, 1, 1, '2037-05-20 13:30:00', NULL, NULL, NULL, NULL, 'pending', 'A'),
(@basketball_tournament_id, 2, 0, '2037-05-20 15:30:00', NULL, NULL, NULL, NULL, 'pending', 'A');

SET @basket_qf1 = (SELECT id FROM matches WHERE tournament_id = @basketball_tournament_id AND round = 0 AND match_number_in_round = 0);
SET @basket_qf2 = (SELECT id FROM matches WHERE tournament_id = @basketball_tournament_id AND round = 0 AND match_number_in_round = 1);
SET @basket_qf3 = (SELECT id FROM matches WHERE tournament_id = @basketball_tournament_id AND round = 0 AND match_number_in_round = 2);
SET @basket_qf4 = (SELECT id FROM matches WHERE tournament_id = @basketball_tournament_id AND round = 0 AND match_number_in_round = 3);
SET @basket_sf1 = (SELECT id FROM matches WHERE tournament_id = @basketball_tournament_id AND round = 1 AND match_number_in_round = 0);
SET @basket_sf2 = (SELECT id FROM matches WHERE tournament_id = @basketball_tournament_id AND round = 1 AND match_number_in_round = 1);
SET @basket_final = (SELECT id FROM matches WHERE tournament_id = @basketball_tournament_id AND round = 2 AND match_number_in_round = 0);

UPDATE matches SET next_match_id = @basket_sf1 WHERE id IN (@basket_qf1, @basket_qf2);
UPDATE matches SET next_match_id = @basket_sf2 WHERE id IN (@basket_qf3, @basket_qf4);
UPDATE matches SET next_match_id = @basket_final WHERE id IN (@basket_sf1, @basket_sf2);

-- バレーボール: 試合開始前の状態。
INSERT INTO matches (
    tournament_id, round, match_number_in_round, match_start_time,
    team1_id, team2_id, status, court_number
) VALUES
(@volleyball_tournament_id, 0, 0, '2037-05-20 09:00:00',
 (SELECT t.id FROM teams t JOIN classes c ON c.id = t.class_id WHERE c.event_id = @demo_event_id AND c.name = '1-1' AND t.sport_id = @volleyball_id),
 (SELECT t.id FROM teams t JOIN classes c ON c.id = t.class_id WHERE c.event_id = @demo_event_id AND c.name = '1-2' AND t.sport_id = @volleyball_id), 'scheduled', 'B'),
(@volleyball_tournament_id, 0, 1, '2037-05-20 10:00:00',
 (SELECT t.id FROM teams t JOIN classes c ON c.id = t.class_id WHERE c.event_id = @demo_event_id AND c.name = '1-3' AND t.sport_id = @volleyball_id),
 (SELECT t.id FROM teams t JOIN classes c ON c.id = t.class_id WHERE c.event_id = @demo_event_id AND c.name = 'IS2' AND t.sport_id = @volleyball_id), 'scheduled', 'B'),
(@volleyball_tournament_id, 1, 0, '2037-05-20 14:00:00', NULL, NULL, 'pending', 'B');

SET @volley_sf1 = (SELECT id FROM matches WHERE tournament_id = @volleyball_tournament_id AND round = 0 AND match_number_in_round = 0);
SET @volley_sf2 = (SELECT id FROM matches WHERE tournament_id = @volleyball_tournament_id AND round = 0 AND match_number_in_round = 1);
SET @volley_final = (SELECT id FROM matches WHERE tournament_id = @volleyball_tournament_id AND round = 1 AND match_number_in_round = 0);
UPDATE matches SET next_match_id = @volley_final WHERE id IN (@volley_sf1, @volley_sf2);

-- 昼競技。再投入時はデモ大会のセッション一式を作り直す。
SET @old_noon_session_id = (
    SELECT id FROM noon_game_sessions WHERE event_id = @demo_event_id
);
DELETE rd FROM noon_game_result_details rd
JOIN noon_game_results nr ON nr.id = rd.result_id
JOIN noon_game_matches nm ON nm.id = nr.match_id
WHERE nm.session_id = @old_noon_session_id;
DELETE trm FROM noon_game_template_run_matches trm
JOIN noon_game_template_runs tr ON tr.id = trm.run_id
WHERE tr.session_id = @old_noon_session_id;
DELETE FROM noon_game_points WHERE session_id = @old_noon_session_id;
DELETE nr FROM noon_game_results nr
JOIN noon_game_matches nm ON nm.id = nr.match_id
WHERE nm.session_id = @old_noon_session_id;
DELETE nme FROM noon_game_match_entries nme
JOIN noon_game_matches nm ON nm.id = nme.match_id
WHERE nm.session_id = @old_noon_session_id;
DELETE FROM noon_game_template_runs WHERE session_id = @old_noon_session_id;
DELETE FROM noon_game_matches WHERE session_id = @old_noon_session_id;
DELETE ngm FROM noon_game_group_members ngm
JOIN noon_game_groups ng ON ng.id = ngm.group_id
WHERE ng.session_id = @old_noon_session_id;
DELETE FROM noon_game_groups WHERE session_id = @old_noon_session_id;
DELETE FROM noon_game_sessions WHERE id = @old_noon_session_id;

INSERT INTO noon_game_sessions (
    event_id, name, description, mode,
    win_points, loss_points, draw_points, participation_points, allow_manual_points
) VALUES (
    @demo_event_id, '学年対抗リレー',
    '完了・進行中・開始前の各状態を確認できるデモ昼競技', 'mixed',
    5, 1, 2, 1, TRUE
);
SET @noon_session_id = LAST_INSERT_ID();

INSERT INTO noon_game_groups (session_id, name, description) VALUES
(@noon_session_id, '赤組', '1組・IS系を中心としたデモグループ'),
(@noon_session_id, '白組', '2組・IT/IE系を中心としたデモグループ');
SET @red_group_id = (
    SELECT id FROM noon_game_groups WHERE session_id = @noon_session_id AND name = '赤組'
);
SET @white_group_id = (
    SELECT id FROM noon_game_groups WHERE session_id = @noon_session_id AND name = '白組'
);

INSERT INTO noon_game_group_members (group_id, class_id, weight)
SELECT
    CASE WHEN c.name IN ('1-1', '1-3', 'IS2', 'IS3', 'IS4', 'IS5', 'IE4', '専教')
         THEN @red_group_id ELSE @white_group_id END,
    c.id,
    CASE WHEN c.name = '専教' THEN 0.50 ELSE 1.00 END
FROM classes c WHERE c.event_id = @demo_event_id;

INSERT INTO noon_game_matches (
    session_id, title, scheduled_at, location, format, status, memo,
    home_side_type, home_class_id, home_group_id,
    away_side_type, away_class_id, away_group_id, allow_draw
) VALUES
(@noon_session_id, '1年生クラス対抗', '2037-05-20 12:10:00', 'グラウンド', '得点制', 'completed', '結果登録済みの確認用',
 'class', (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = '1-1'), NULL,
 'class', (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = '1-2'), NULL, FALSE),
(@noon_session_id, '赤組対白組 綱引き', '2037-05-20 12:25:00', 'グラウンド', '3本勝負', 'in_progress', '進行中表示の確認用',
 'group', NULL, @red_group_id, 'group', NULL, @white_group_id, FALSE),
(@noon_session_id, '選抜リレー', '2037-05-20 12:40:00', 'トラック', '4チーム順位制', 'completed', '順位詳細とテンプレート実行の確認用',
 'group', NULL, @red_group_id, 'group', NULL, @white_group_id, FALSE),
(@noon_session_id, '午後の部ボーナス競技', '2037-05-20 13:00:00', 'グラウンド', 'ポイント制', 'scheduled', '開始前表示の確認用',
 'group', NULL, @red_group_id, 'group', NULL, @white_group_id, TRUE);

SET @noon_class_match_id = (
    SELECT id FROM noon_game_matches WHERE session_id = @noon_session_id AND title = '1年生クラス対抗'
);
SET @noon_relay_match_id = (
    SELECT id FROM noon_game_matches WHERE session_id = @noon_session_id AND title = '選抜リレー'
);

INSERT INTO noon_game_results (
    match_id, winner, home_score, away_score, recorded_by, recorded_at, note
) VALUES
(@noon_class_match_id, 'home', 12, 8, '00000000-0000-4000-8000-000000000001', '2037-05-20 12:22:00', 'デモ結果'),
(@noon_relay_match_id, 'home', NULL, NULL, '00000000-0000-4000-8000-000000000001', '2037-05-20 12:52:00', '順位入力済み');

INSERT INTO noon_game_match_entries (
    match_id, entry_index, side_type, class_id, display_name
) VALUES
(@noon_relay_match_id, 0, 'class', (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = '1-1'), '1-1選抜'),
(@noon_relay_match_id, 1, 'class', (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = '1-2'), '1-2選抜'),
(@noon_relay_match_id, 2, 'class', (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = 'IS2'), 'IS2選抜'),
(@noon_relay_match_id, 3, 'class', (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = 'IT2'), 'IT2選抜');

SET @noon_relay_result_id = (
    SELECT id FROM noon_game_results WHERE match_id = @noon_relay_match_id
);
INSERT INTO noon_game_result_details (
    result_id, entry_id, placement_rank, points, note, entry_resolved_name
)
SELECT
    @noon_relay_result_id,
    e.id,
    e.entry_index + 1,
    CASE e.entry_index WHEN 0 THEN 8 WHEN 1 THEN 5 WHEN 2 THEN 3 ELSE 1 END,
    'デモ順位',
    e.display_name
FROM noon_game_match_entries e
WHERE e.match_id = @noon_relay_match_id;

INSERT INTO noon_game_template_runs (
    session_id, template_key, name, points_by_rank, created_by
) VALUES (
    @noon_session_id, 'year_relay', '2037年度 学年対抗リレー',
    JSON_ARRAY(8, 5, 3, 1), '00000000-0000-4000-8000-000000000001'
);
SET @noon_run_id = LAST_INSERT_ID();
INSERT INTO noon_game_template_run_matches (run_id, match_id, match_key)
VALUES (@noon_run_id, @noon_relay_match_id, 'final');

INSERT INTO noon_game_points (
    session_id, match_id, class_id, points, reason, source, created_by
) VALUES
(@noon_session_id, @noon_class_match_id, (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = '1-1'), 5, 'クラス対抗勝利', 'result', '00000000-0000-4000-8000-000000000001'),
(@noon_session_id, @noon_class_match_id, (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = '1-2'), 1, 'クラス対抗参加', 'result', '00000000-0000-4000-8000-000000000001'),
(@noon_session_id, @noon_relay_match_id, (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = '1-1'), 8, '選抜リレー1位', 'result', '00000000-0000-4000-8000-000000000001'),
(@noon_session_id, @noon_relay_match_id, (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = '1-2'), 5, '選抜リレー2位', 'result', '00000000-0000-4000-8000-000000000001'),
(@noon_session_id, @noon_relay_match_id, (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = 'IS2'), 3, '選抜リレー3位', 'result', '00000000-0000-4000-8000-000000000001'),
(@noon_session_id, @noon_relay_match_id, (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = 'IT2'), 1, '選抜リレー4位', 'result', '00000000-0000-4000-8000-000000000001'),
(@noon_session_id, NULL, (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = '専教'), 2, '運営協力ボーナス', 'manual', '00000000-0000-4000-8000-000000000001');

-- 出席登録。a〜dは開会式、a〜bは競技参加も登録済みにする。
DELETE ci FROM check_ins ci
JOIN users u ON u.id = ci.user_id
WHERE ci.event_id = @demo_event_id AND u.email LIKE 'demo-%@example.com';
INSERT INTO check_ins (user_id, event_id, purpose, checked_in_at)
SELECT u.id, @demo_event_id, 'opening_ceremony', '2037-05-20 08:25:00'
FROM users u JOIN classes c ON c.id = u.class_id
WHERE c.event_id = @demo_event_id
  AND u.email REGEXP 'demo-.+-[a-d]@example\\.com$';
INSERT INTO check_ins (user_id, event_id, purpose, checked_in_at)
SELECT u.id, @demo_event_id, 'event_participation', '2037-05-20 08:50:00'
FROM users u JOIN classes c ON c.id = u.class_id
WHERE c.event_id = @demo_event_id
  AND u.email REGEXP 'demo-.+-[a-b]@example\\.com$';

-- バスケットボール準々決勝のラウンド受付。
DELETE FROM round_check_ins WHERE event_id = @demo_event_id;
INSERT INTO round_check_ins (
    event_id, sport_id, match_id, round, user_id, team_id, checked_in_at
)
SELECT @demo_event_id, @basketball_id, @basket_qf1, 0, u.id, t.id, '2037-05-20 08:55:00'
FROM users u
JOIN classes c ON c.id = u.class_id
JOIN teams t ON t.class_id = c.id AND t.sport_id = @basketball_id
WHERE c.event_id = @demo_event_id
  AND c.name IN ('1-1', '1-2')
  AND u.email REGEXP 'demo-.+-[a-d]@example\\.com$';

-- MIC投票。各クラスのaユーザーが自クラスへ投票済みの状態。
DELETE FROM mic_votes WHERE event_id = @demo_event_id;
INSERT INTO mic_votes (
    event_id, voter_user_id, voted_for_class_id, reason, points, created_at
)
SELECT @demo_event_id, u.id, c.id, 'クラスの団結力と応援が印象的だったため', 3, '2037-05-20 16:00:00'
FROM classes c
JOIN demo_class_map dcm ON dcm.class_name = c.name
JOIN users u ON u.class_id = c.id
    AND u.email = CONCAT('demo-', dcm.slug, '-a@example.com')
WHERE c.event_id = @demo_event_id;

-- 得点履歴。ランキングと内訳の両方にばらつきが出るようにする。
DELETE FROM score_logs WHERE event_id = @demo_event_id;
INSERT INTO score_logs (event_id, class_id, points, reason)
SELECT @demo_event_id, id, 10 + MOD(id, 5), 'initial_points'
FROM classes WHERE event_id = @demo_event_id;
INSERT INTO score_logs (event_id, class_id, points, reason)
SELECT @demo_event_id, id, MOD(id, 4), 'survey_points'
FROM classes WHERE event_id = @demo_event_id;
INSERT INTO score_logs (event_id, class_id, points, reason)
SELECT @demo_event_id, id, FLOOR(attend_count / 10), 'attendance_points'
FROM classes WHERE event_id = @demo_event_id;
INSERT INTO score_logs (event_id, class_id, points, reason)
SELECT @demo_event_id, id, 3, 'mic_points'
FROM classes WHERE event_id = @demo_event_id;
INSERT INTO score_logs (event_id, class_id, points, reason, source_match_id) VALUES
(@demo_event_id, (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = '1-1'), 5, 'gym1_win1_points', @basket_qf1),
(@demo_event_id, (SELECT id FROM classes WHERE event_id = @demo_event_id AND name = 'IS2'), 5, 'gym1_win1_points', @basket_qf2);
INSERT INTO score_logs (event_id, class_id, points, reason)
SELECT @demo_event_id, class_id, SUM(points), 'noon_game_points'
FROM noon_game_points WHERE session_id = @noon_session_id
GROUP BY class_id;

-- 大会要項・利用ガイド。既存PDFを開発用uploadsへread-onlyマウントしている。
DELETE FROM guide_documents WHERE event_id = @demo_event_id;
INSERT INTO guide_documents (event_id, title, description, pdf_url) VALUES
(@demo_event_id, 'システム管理者向けガイド', '大会・競技・ユーザー管理の操作確認用資料', '/uploads/demo/root.pdf'),
(@demo_event_id, '運営担当者向けガイド', '受付・試合結果登録の操作確認用資料', '/uploads/demo/admin.pdf'),
(@demo_event_id, '参加者向けガイド', 'マイページ・通知・競技情報の操作確認用資料', '/uploads/demo/student.pdf');

-- 通知種別・対象ロールごとの表示確認用データ。
DELETE FROM notifications WHERE event_id = @demo_event_id;
INSERT INTO notifications (title, body, type, created_by, event_id, created_at) VALUES
('デモ大会へようこそ', '競技、チーム、対戦表、得点を確認できます。', 'general', '00000000-0000-4000-8000-000000000001', @demo_event_id, '2037-05-19 17:00:00'),
('集合時刻のお知らせ', '各クラスは8時20分までにグラウンドへ集合してください。', 'general', '00000000-0000-4000-8000-000000000001', @demo_event_id, '2037-05-20 07:30:00'),
('次の試合案内', '対象クラスは試合開始10分前までに受付を済ませてください。', 'match_my_class', '00000000-0000-4000-8000-000000000001', @demo_event_id, '2037-05-20 09:40:00'),
('決勝戦のご案内', '決勝戦は15時30分から第1体育館で行います。', 'finals', '00000000-0000-4000-8000-000000000001', @demo_event_id, '2037-05-20 14:30:00'),
('全試合終了予定', '本日の全試合は16時終了予定です。', 'all_matches', '00000000-0000-4000-8000-000000000001', @demo_event_id, '2037-05-20 15:00:00');

INSERT INTO notification_targets (notification_id, role_name)
SELECT id, 'student' FROM notifications WHERE event_id = @demo_event_id;
INSERT INTO notification_targets (notification_id, role_name)
SELECT id, 'admin' FROM notifications WHERE event_id = @demo_event_id AND type IN ('general', 'all_matches');
INSERT INTO notification_targets (notification_id, role_name)
SELECT id, 'root' FROM notifications WHERE event_id = @demo_event_id AND type = 'all_matches';

-- 通知申請は pending / approved / rejected と会話履歴を用意する。
DELETE nr FROM notification_requests nr
JOIN users u ON u.id = nr.requester_id
WHERE u.email LIKE 'demo-%@example.com';
INSERT INTO notification_requests (
    title, body, target_text, status, requester_id, resolved_by, resolved_at, created_at
) VALUES
('応援場所変更のお知らせ', '雨天のため応援場所を変更します。', '全校生徒', 'pending',
 (SELECT id FROM users WHERE email = 'demo-1-1-a@example.com'), NULL, NULL, '2037-05-20 10:00:00'),
('クラス集合のお知らせ', '1-2は12時に中庭へ集合してください。', '1-2', 'approved',
 (SELECT id FROM users WHERE email = 'demo-1-2-a@example.com'), '00000000-0000-4000-8000-000000000000', '2037-05-20 09:20:00', '2037-05-20 09:00:00'),
('落とし物のお知らせ', '体育館でタオルを拾いました。', '全校生徒', 'rejected',
 (SELECT id FROM users WHERE email = 'demo-is2-a@example.com'), '00000000-0000-4000-8000-000000000000', '2037-05-20 11:10:00', '2037-05-20 11:00:00');

SET @pending_request_id = (SELECT id FROM notification_requests WHERE title = '応援場所変更のお知らせ' AND requester_id = (SELECT id FROM users WHERE email = 'demo-1-1-a@example.com'));
SET @approved_request_id = (SELECT id FROM notification_requests WHERE title = 'クラス集合のお知らせ' AND requester_id = (SELECT id FROM users WHERE email = 'demo-1-2-a@example.com'));
SET @rejected_request_id = (SELECT id FROM notification_requests WHERE title = '落とし物のお知らせ' AND requester_id = (SELECT id FROM users WHERE email = 'demo-is2-a@example.com'));
INSERT INTO notification_request_messages (request_id, sender_id, message, created_at) VALUES
(@pending_request_id, (SELECT id FROM users WHERE email = 'demo-1-1-a@example.com'), 'できるだけ早い配信をお願いします。', '2037-05-20 10:01:00'),
(@pending_request_id, '00000000-0000-4000-8000-000000000000', '対象となる応援場所を確認しています。', '2037-05-20 10:05:00'),
(@approved_request_id, '00000000-0000-4000-8000-000000000000', '内容を確認し、承認しました。', '2037-05-20 09:20:00'),
(@rejected_request_id, '00000000-0000-4000-8000-000000000000', '個別の落とし物連絡として案内してください。', '2037-05-20 11:10:00');

DROP TEMPORARY TABLE demo_class_map;

COMMIT;

SELECT CONCAT('Demo data loaded: event_id=', @demo_event_id) AS result;
