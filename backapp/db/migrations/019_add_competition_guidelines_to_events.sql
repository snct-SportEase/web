-- 競技要項PDF URLをeventsテーブルに追加
ALTER TABLE events ADD COLUMN competition_guidelines_pdf_url VARCHAR(500) NULL;

