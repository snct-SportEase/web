ALTER TABLE team_members
    ADD COLUMN entry_order SMALLINT UNSIGNED NULL AFTER is_confirmed;

UPDATE team_members tm
JOIN (
    SELECT
        team_id,
        user_id,
        ROW_NUMBER() OVER (PARTITION BY team_id ORDER BY user_id) AS generated_entry_order
    FROM team_members
) ranked
    ON ranked.team_id = tm.team_id
    AND ranked.user_id = tm.user_id
SET tm.entry_order = ranked.generated_entry_order;

ALTER TABLE team_members
    MODIFY COLUMN entry_order SMALLINT UNSIGNED NOT NULL,
    ADD CONSTRAINT uq_team_members_entry_order UNIQUE (team_id, entry_order);
