-- +migrate Up
ALTER TABLE leads ADD COLUMN notes TEXT;
ALTER TABLE leads ADD COLUMN location VARCHAR(255);

-- +migrate Down
ALTER TABLE leads DROP COLUMN notes;
ALTER TABLE leads DROP COLUMN location;
