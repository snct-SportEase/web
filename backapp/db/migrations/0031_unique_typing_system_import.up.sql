-- Drop extra history rows left by the import race so the unique key can be added.
-- Prefer an active row over an inactive one; otherwise keep the oldest id.
-- Point rows are left untouched; doubled noon_game_points still need a manual repair.
DELETE drop_row
FROM noon_game_typing_system_imports AS drop_row
INNER JOIN noon_game_typing_system_imports AS keep_row
    ON keep_row.session_id = drop_row.session_id
    AND keep_row.export_id = drop_row.export_id
    AND keep_row.sha256 = drop_row.sha256
    AND keep_row.id <> drop_row.id
    AND (
        (drop_row.is_active = FALSE AND keep_row.is_active = TRUE)
        OR (drop_row.is_active = keep_row.is_active AND drop_row.id > keep_row.id)
    );

ALTER TABLE noon_game_typing_system_imports
    ADD UNIQUE KEY uq_noon_game_typing_import_session_export_sha (session_id, export_id, sha256);
