-- 管理画面から作成され、クラスが1件も登録されていない大会を復旧する。
INSERT INTO classes (event_id, name)
SELECT e.id, default_classes.name
FROM events e
CROSS JOIN (
    SELECT '1-1' AS name
    UNION ALL SELECT '1-2'
    UNION ALL SELECT '1-3'
    UNION ALL SELECT 'IS2'
    UNION ALL SELECT 'IS3'
    UNION ALL SELECT 'IS4'
    UNION ALL SELECT 'IS5'
    UNION ALL SELECT 'IT2'
    UNION ALL SELECT 'IT3'
    UNION ALL SELECT 'IT4'
    UNION ALL SELECT 'IT5'
    UNION ALL SELECT 'IE2'
    UNION ALL SELECT 'IE3'
    UNION ALL SELECT 'IE4'
    UNION ALL SELECT 'IE5'
    UNION ALL SELECT '専教'
) AS default_classes
WHERE NOT EXISTS (
    SELECT 1
    FROM classes existing_classes
    WHERE existing_classes.event_id = e.id
);
