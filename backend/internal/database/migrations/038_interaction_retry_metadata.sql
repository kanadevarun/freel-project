ALTER TABLE lead_interactions ADD COLUMN retry_count INT NOT NULL DEFAULT 0;
ALTER TABLE lead_interactions ADD COLUMN last_retry_at DATETIME NULL;
ALTER TABLE lead_interactions ADD COLUMN last_error TEXT NULL;
