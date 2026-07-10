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

-- 完全な旧スキーマが確認できる場合だけ、失敗した初回適用のdirty状態を修復する。
DELETE FROM schema_migrations
WHERE dirty = TRUE AND @base_schema_complete = 3;

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
