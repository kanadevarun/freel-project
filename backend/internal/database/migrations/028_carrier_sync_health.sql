ALTER TABLE carrier_integrations
ADD COLUMN failed_attempts INT NOT NULL DEFAULT 0,
ADD COLUMN error_details TEXT,
ADD COLUMN next_retry_time DATETIME;
