-- +migrate Up
ALTER TABLE rfqs ADD COLUMN lead_id BIGINT DEFAULT NULL REFERENCES leads(id) ON DELETE SET NULL;

-- +migrate Down
ALTER TABLE rfqs DROP COLUMN lead_id;
