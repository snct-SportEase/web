ALTER TABLE event_sports
ADD COLUMN rules_type ENUM('markdown', 'pdf') NOT NULL DEFAULT 'markdown',
ADD COLUMN rules_pdf_url VARCHAR(255) NULL;
