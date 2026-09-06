-- +migrate Up
ALTER TABLE lead_interactions
ADD COLUMN rfc_message_id VARCHAR(255) NULL,
ADD COLUMN in_reply_to VARCHAR(255) NULL,
ADD COLUMN references_header TEXT NULL,
ADD COLUMN sender VARCHAR(255) NULL,
ADD COLUMN recipients TEXT NULL,
ADD COLUMN cc_recipients TEXT NULL;

-- +migrate Down
ALTER TABLE lead_interactions
DROP COLUMN rfc_message_id,
DROP COLUMN in_reply_to,
DROP COLUMN references_header,
DROP COLUMN sender,
DROP COLUMN recipients,
DROP COLUMN cc_recipients;
