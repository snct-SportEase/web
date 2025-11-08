-- Remove AFTER triggers that cause error 1442
-- These triggers try to update class_scores table from within a trigger on the same table,
-- which is not allowed in MySQL

DROP TRIGGER IF EXISTS after_class_scores_insert;
DROP TRIGGER IF EXISTS after_class_scores_update;

-- Remove stored procedures that are no longer needed
-- (Ranking will be updated by the application code)
DROP PROCEDURE IF EXISTS update_class_ranks;
DROP PROCEDURE IF EXISTS update_class_overall_ranks;


