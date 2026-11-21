-- golang-migrate導入前から存在する開発DBを、安全にベースライン登録する。
-- 新規DBでは空のschema_migrationsだけを作り、demo-migrateがversion 1から適用する。

SET @base_schema_complete = (
    SELECT COUNT(*)
    FROM information_schema.tables
    WHERE table_schema = DATABASE()
      AND table_name IN ('roles', 'noon_game_template_default_groups', 'class_scores')
);

SET @has_guide_documents = (
    SELECT COUNT(*)
    FROM information_schema.tables
    WHERE table_schema = DATABASE() AND table_name = 'guide_documents'
);

SET @has_round_check_ins = (
    SELECT COUNT(*)
    FROM information_schema.tables
    WHERE table_schema = DATABASE() AND table_name = 'round_check_ins'
);

CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT NOT NULL PRIMARY KEY,
    dirty BOOLEAN NOT NULL
);

-- dirty を無条件に消して古いバージョンへ戻すと、適用済みの変更を再実行してしまう。
-- 25 は MySQL が ALTER TABLE をコミットした後に失敗しうるため、
-- 必要な列とインデックスが揃っている場合に限り残りのデータ更新を完了する。
SET @recover_25 = (
    SELECT COUNT(*) = 1 FROM schema_migrations WHERE version = 25 AND dirty = TRUE
) AND (
    SELECT COUNT(*) = 2 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'notifications'
      AND column_name IN ('scheduled_at', 'sent_at') AND column_type = 'datetime(6)'
) AND (
    SELECT COUNT(*) = 2 AND
        SUM(seq_in_index = 1 AND column_name = 'sent_at') = 1 AND
        SUM(seq_in_index = 2 AND column_name = 'scheduled_at') = 1
    FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = 'notifications'
      AND index_name = 'idx_notifications_due'
);

SET @recovery_sql = IF(@recover_25,
    'UPDATE notifications SET sent_at = created_at WHERE sent_at IS NULL AND scheduled_at IS NULL',
    'SELECT 1');
PREPARE recovery_stmt FROM @recovery_sql;
EXECUTE recovery_stmt;
DEALLOCATE PREPARE recovery_stmt;

UPDATE schema_migrations SET dirty = FALSE WHERE version = 25 AND dirty = TRUE AND @recover_25;

INSERT INTO schema_migrations (version, dirty)
SELECT
    CASE
        WHEN @has_guide_documents = 1 AND @has_round_check_ins = 1 THEN 3
        WHEN @has_guide_documents = 1 THEN 2
        ELSE 1
    END,
    FALSE
WHERE @base_schema_complete = 3
  AND NOT EXISTS (SELECT 1 FROM schema_migrations);

SELECT
    CASE
        WHEN @base_schema_complete = 3 THEN 'Existing schema baseline is ready'
        ELSE 'Fresh database is ready for migrations'
    END AS result;
