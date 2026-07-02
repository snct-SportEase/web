DROP VIEW IF EXISTS class_scores;

SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS noon_game_template_default_groups;
DROP TABLE IF EXISTS noon_game_template_run_matches;
DROP TABLE IF EXISTS noon_game_template_runs;
DROP TABLE IF EXISTS noon_game_result_details;
DROP TABLE IF EXISTS noon_game_match_entries;
DROP TABLE IF EXISTS noon_game_points;
DROP TABLE IF EXISTS noon_game_results;
DROP TABLE IF EXISTS noon_game_matches;
DROP TABLE IF EXISTS noon_game_group_members;
DROP TABLE IF EXISTS noon_game_groups;
DROP TABLE IF EXISTS noon_game_sessions;
DROP TABLE IF EXISTS notification_request_messages;
DROP TABLE IF EXISTS notification_requests;
DROP TABLE IF EXISTS push_subscriptions;
DROP TABLE IF EXISTS notification_targets;
DROP TABLE IF EXISTS notification_recipients;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS mic_votes;
DROP TABLE IF EXISTS check_ins;
DROP TABLE IF EXISTS score_logs;
DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS tournaments;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS rainy_mode_settings;
DROP TABLE IF EXISTS event_sports;
DROP TABLE IF EXISTS sports;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS classes;
DROP TABLE IF EXISTS active_event;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS roles;

SET FOREIGN_KEY_CHECKS = 1;
