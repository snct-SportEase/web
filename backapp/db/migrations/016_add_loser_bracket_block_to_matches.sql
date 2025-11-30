-- matchesテーブルに敗者戦ブロック識別カラムを追加
ALTER TABLE matches
ADD COLUMN loser_bracket_block VARCHAR(1) NULL DEFAULT NULL COMMENT '敗者戦ブロック識別: A or B';

