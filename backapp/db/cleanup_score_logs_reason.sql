-- cleanup_score_logs_reason.sql
--
-- This script is intended to be run manually against the database before
-- applying the schema change that constrains score_logs.reason values.
--
-- 1) Run the SELECT to see what values exist.
-- 2) Use the UPDATE statements to normalize / clean up invalid values.
--
-- NOTE: This file is saved in backapp/db/ for manual use only; it is NOT
-- automatically applied by any migration framework.

-- 1) Find invalid "reason" values currently in the table
SELECT DISTINCT reason
FROM score_logs
WHERE reason NOT IN (
    'attendance_points',
    'initial_points',
    'survey_points',
    'mic_points',
    'gym1_win1_points',
    'gym1_win2_points',
    'gym1_win3_points',
    'gym1_champion_points',
    'gym2_win1_points',
    'gym2_win2_points',
    'gym2_win3_points',
    'gym2_champion_points',
    'gym2_loser_bracket_champion_points',
    'ground_win1_points',
    'ground_win2_points',
    'ground_win3_points',
    'ground_champion_points',
    'noon_game_points'
);

-- 2) Normalize known common invalid forms (e.g. "mic_points: ...")
--    Keep only the allowed canonical value.
UPDATE score_logs
SET reason = 'mic_points'
WHERE reason LIKE 'mic_points:%';

-- 3) Optionally normalize any remaining invalid values (uncomment and adjust as needed)
-- UPDATE score_logs
-- SET reason = 'initial_points'
-- WHERE reason NOT IN (
--     'attendance_points',
--     'initial_points',
--     'survey_points',
--     'mic_points',
--     'gym1_win1_points',
--     'gym1_win2_points',
--     'gym1_win3_points',
--     'gym1_champion_points',
--     'gym2_win1_points',
--     'gym2_win2_points',
--     'gym2_win3_points',
--     'gym2_champion_points',
--     'gym2_loser_bracket_champion_points',
--     'ground_win1_points',
--     'ground_win2_points',
--     'ground_win3_points',
--     'ground_champion_points',
--     'noon_game_points'
-- );
