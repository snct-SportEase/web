ALTER TABLE events
ADD COLUMN is_mic_voting_enabled BOOLEAN NOT NULL DEFAULT TRUE
COMMENT '行事委員会賞投票を利用可能にするか'
AFTER is_survey_published;
