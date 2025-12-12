-- team_membersテーブルに参加本登録を示すis_confirmedカラムを追加
ALTER TABLE team_members
ADD COLUMN is_confirmed BOOLEAN NOT NULL DEFAULT false;

-- 既存データはすべて未確認（仮登録）として扱う
UPDATE team_members SET is_confirmed = false;

