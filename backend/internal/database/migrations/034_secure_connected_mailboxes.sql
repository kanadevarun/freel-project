-- +migrate Up
ALTER TABLE org_connected_mailboxes
ADD COLUMN provider VARCHAR(50) NOT NULL DEFAULT 'IMAP',
ADD COLUMN access_token_encrypted TEXT NULL,
ADD COLUMN refresh_token_encrypted TEXT NULL,
ADD COLUMN token_expiry TIMESTAMP NULL,
ADD COLUMN oauth_scopes TEXT NULL,
ADD COLUMN sync_cursor VARCHAR(255) NULL,
ADD COLUMN last_sync_started_at TIMESTAMP NULL,
ADD COLUMN last_sync_success_at TIMESTAMP NULL,
ADD COLUMN last_sync_error TEXT NULL,
ADD COLUMN last_processed_message_id VARCHAR(255) NULL;

-- +migrate Down
ALTER TABLE org_connected_mailboxes
DROP COLUMN provider,
DROP COLUMN access_token_encrypted,
DROP COLUMN refresh_token_encrypted,
DROP COLUMN token_expiry,
DROP COLUMN oauth_scopes,
DROP COLUMN sync_cursor,
DROP COLUMN last_sync_started_at,
DROP COLUMN last_sync_success_at,
DROP COLUMN last_sync_error,
DROP COLUMN last_processed_message_id;
