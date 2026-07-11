ALTER TABLE events
ADD COLUMN duplicate_registration_threshold INT NOT NULL DEFAULT 31
COMMENT '2競技まで重複登録できるクラス人数の上限'
AFTER hide_scores;
