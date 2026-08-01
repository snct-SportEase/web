ALTER TABLE team_members
    DROP INDEX uq_team_members_entry_order,
    DROP COLUMN entry_order;
