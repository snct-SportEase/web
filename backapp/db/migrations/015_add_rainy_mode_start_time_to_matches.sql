-- matchesテーブルに雨天時モード用の開始時間カラムを追加
ALTER TABLE matches
ADD COLUMN rainy_mode_start_time VARCHAR(255) NULL;

