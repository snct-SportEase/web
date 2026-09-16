-- 同じユーザー・ロールを春秋それぞれに保存する。NULLは大会共通の割り当て。
-- 仮想列でNULLを0として一意に扱い、外部キーの削除時CASCADEを維持する。
ALTER TABLE user_roles
    ADD COLUMN id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    ADD COLUMN event_scope_id INT GENERATED ALWAYS AS (COALESCE(event_id, 0)) VIRTUAL,
    ADD KEY idx_user_roles_user (user_id),
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (id),
    ADD UNIQUE KEY uq_user_role_event_scope (user_id, role_id, event_scope_id);

-- 過去のREPLACEで上書きされた割り当てを、各大会のチーム登録から復元する。
INSERT IGNORE INTO user_roles (user_id, role_id, event_id)
SELECT tm.user_id, r.id, c.event_id
FROM team_members tm
JOIN teams t ON t.id = tm.team_id
JOIN classes c ON c.id = t.class_id
JOIN sports s ON s.id = t.sport_id
JOIN roles r ON r.name = CONCAT(c.name, '_', s.name);

-- 競技ロールの大会共通割り当てを解除する。所属先は上記のチーム登録から復元済み。
-- 基本ロールと任意の共通ロールは変更しない。
DELETE ur
FROM user_roles ur
JOIN roles r ON r.id = ur.role_id
WHERE ur.event_id IS NULL
  AND EXISTS (
      SELECT 1 FROM classes c
      JOIN sports s ON r.name = CONCAT(c.name, '_', s.name)
  );
