-- Recovery for the production failure that left migration 16 dirty after
-- MySQL committed the first ALTER TABLE statements. This script is safe to
-- run only when schema_migrations contains (16, true); the deploy workflows
-- enforce that precondition before invoking it.

SET @fk_name = (
    SELECT CONSTRAINT_NAME
    FROM information_schema.KEY_COLUMN_USAGE
    WHERE CONSTRAINT_SCHEMA = DATABASE()
      AND TABLE_NAME = 'check_ins'
      AND COLUMN_NAME = 'user_id'
      AND REFERENCED_TABLE_NAME = 'users'
    LIMIT 1
);
SET @recovery_sql = IF(@fk_name IS NULL, 'SELECT 1', CONCAT('ALTER TABLE check_ins DROP FOREIGN KEY `', REPLACE(@fk_name, '`', '``'), '`'));
PREPARE recovery_stmt FROM @recovery_sql;
EXECUTE recovery_stmt;
DEALLOCATE PREPARE recovery_stmt;
ALTER TABLE check_ins
    ADD CONSTRAINT fk_check_ins_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE;

SET @fk_name = (
    SELECT CONSTRAINT_NAME
    FROM information_schema.KEY_COLUMN_USAGE
    WHERE CONSTRAINT_SCHEMA = DATABASE()
      AND TABLE_NAME = 'mic_votes'
      AND COLUMN_NAME = 'voter_user_id'
      AND REFERENCED_TABLE_NAME = 'users'
    LIMIT 1
);
SET @recovery_sql = IF(@fk_name IS NULL, 'SELECT 1', CONCAT('ALTER TABLE mic_votes DROP FOREIGN KEY `', REPLACE(@fk_name, '`', '``'), '`'));
PREPARE recovery_stmt FROM @recovery_sql;
EXECUTE recovery_stmt;
DEALLOCATE PREPARE recovery_stmt;
ALTER TABLE mic_votes
    ADD CONSTRAINT fk_mic_votes_user FOREIGN KEY (voter_user_id)
        REFERENCES users(id) ON DELETE CASCADE;

SET @fk_name = (
    SELECT CONSTRAINT_NAME
    FROM information_schema.KEY_COLUMN_USAGE
    WHERE CONSTRAINT_SCHEMA = DATABASE()
      AND TABLE_NAME = 'notifications'
      AND COLUMN_NAME = 'created_by'
      AND REFERENCED_TABLE_NAME = 'users'
    LIMIT 1
);
SET @recovery_sql = IF(@fk_name IS NULL, 'SELECT 1', CONCAT('ALTER TABLE notifications DROP FOREIGN KEY `', REPLACE(@fk_name, '`', '``'), '`'));
PREPARE recovery_stmt FROM @recovery_sql;
EXECUTE recovery_stmt;
DEALLOCATE PREPARE recovery_stmt;
ALTER TABLE notifications
    ADD CONSTRAINT fk_notifications_creator FOREIGN KEY (created_by)
        REFERENCES users(id) ON DELETE SET NULL;

SET @fk_name = (
    SELECT CONSTRAINT_NAME
    FROM information_schema.KEY_COLUMN_USAGE
    WHERE CONSTRAINT_SCHEMA = DATABASE()
      AND TABLE_NAME = 'board_game_entry_members'
      AND COLUMN_NAME = 'user_id'
      AND REFERENCED_TABLE_NAME = 'users'
    LIMIT 1
);
SET @recovery_sql = IF(@fk_name IS NULL, 'SELECT 1', CONCAT('ALTER TABLE board_game_entry_members DROP FOREIGN KEY `', REPLACE(@fk_name, '`', '``'), '`'));
PREPARE recovery_stmt FROM @recovery_sql;
EXECUTE recovery_stmt;
DEALLOCATE PREPARE recovery_stmt;
ALTER TABLE board_game_entry_members
    ADD CONSTRAINT fk_board_game_member_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE;

SET @fk_name = (
    SELECT CONSTRAINT_NAME
    FROM information_schema.KEY_COLUMN_USAGE
    WHERE CONSTRAINT_SCHEMA = DATABASE()
      AND TABLE_NAME = 'board_game_runs'
      AND COLUMN_NAME = 'created_by'
      AND REFERENCED_TABLE_NAME = 'users'
    LIMIT 1
);
SET @recovery_sql = IF(@fk_name IS NULL, 'SELECT 1', CONCAT('ALTER TABLE board_game_runs DROP FOREIGN KEY `', REPLACE(@fk_name, '`', '``'), '`'));
PREPARE recovery_stmt FROM @recovery_sql;
EXECUTE recovery_stmt;
DEALLOCATE PREPARE recovery_stmt;
ALTER TABLE board_game_runs
    MODIFY COLUMN created_by CHAR(36) NULL,
    ADD CONSTRAINT fk_board_game_run_creator FOREIGN KEY (created_by)
        REFERENCES users(id) ON DELETE SET NULL;

SET @fk_name = (
    SELECT CONSTRAINT_NAME
    FROM information_schema.KEY_COLUMN_USAGE
    WHERE CONSTRAINT_SCHEMA = DATABASE()
      AND TABLE_NAME = 'board_game_rankings'
      AND COLUMN_NAME = 'recorded_by'
      AND REFERENCED_TABLE_NAME = 'users'
    LIMIT 1
);
SET @recovery_sql = IF(@fk_name IS NULL, 'SELECT 1', CONCAT('ALTER TABLE board_game_rankings DROP FOREIGN KEY `', REPLACE(@fk_name, '`', '``'), '`'));
PREPARE recovery_stmt FROM @recovery_sql;
EXECUTE recovery_stmt;
DEALLOCATE PREPARE recovery_stmt;
ALTER TABLE board_game_rankings
    MODIFY COLUMN recorded_by CHAR(36) NULL,
    ADD CONSTRAINT fk_board_game_ranking_recorder FOREIGN KEY (recorded_by)
        REFERENCES users(id) ON DELETE SET NULL;

UPDATE schema_migrations
SET dirty = FALSE
WHERE version = 16 AND dirty = TRUE;
