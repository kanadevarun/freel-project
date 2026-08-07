-- +migrate Up
ALTER TABLE rfq_quotes
ADD COLUMN IF NOT EXISTS free_days INT DEFAULT 0;

-- +migrate Down
ALTER TABLE rfq_quotes
DROP COLUMN IF EXISTS free_days;
