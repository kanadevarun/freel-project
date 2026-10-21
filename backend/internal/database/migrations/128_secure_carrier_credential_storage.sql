-- Migration 128: Secure Carrier Credential Storage
-- Removes plaintext credential persistence from carrier_integrations.
-- Enforces that credentials are only persisted encrypted in encrypted_credentials.

-- +migrate Up
-- 1. Clear plaintext credentials_json for integrations that have encrypted_credentials
UPDATE carrier_integrations
SET credentials_json = NULL
WHERE encrypted_credentials IS NOT NULL
  AND encrypted_credentials != ''
  AND credentials_json IS NOT NULL;

-- 2. Deprecate credentials_json column: ensure it is NULLable with deprecation comment
ALTER TABLE carrier_integrations
MODIFY COLUMN credentials_json JSON NULL COMMENT 'DEPRECATED: Plaintext credential storage removed. All credentials must be stored encrypted in encrypted_credentials.';

-- +migrate Down
ALTER TABLE carrier_integrations
MODIFY COLUMN credentials_json JSON NULL COMMENT 'Legacy credentials json';
