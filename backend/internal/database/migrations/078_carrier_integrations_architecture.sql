-- Migration 078: Carrier Integration Architecture & Provider Registry Foundation
-- Establishes the reusable carrier provider registry, tenant integration schema,
-- encryption at rest support, connection status, sync metadata, and tenant uniqueness rules.

-- 1. Carrier Provider Registry
CREATE TABLE IF NOT EXISTS carrier_providers (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    scac VARCHAR(20) NOT NULL,
    modes JSON NOT NULL,
    adapter_key VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    supported_capabilities JSON NOT NULL,
    description TEXT NULL,
    documentation_url VARCHAR(500) NULL,
    logo_url VARCHAR(500) NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_cp_code (code),
    INDEX idx_cp_scac (scac)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed Global Supported Carrier Providers
INSERT INTO carrier_providers (code, name, scac, modes, adapter_key, is_active, supported_capabilities, description)
VALUES 
('MAERSK', 'A.P. Moller – Maersk', 'MAEU', JSON_ARRAY('OCEAN', 'INTERMODAL'), 'MAERSK_ADAPTER', TRUE, JSON_ARRAY('TRACKING', 'RATES', 'CONTRACT_RATES', 'SPOT_RATES', 'BOOKING', 'DOCUMENTS'), 'Global container logistics integrator offering end-to-end multi-modal transport solutions.'),
('MSC', 'Mediterranean Shipping Company (MSC)', 'MSCU', JSON_ARRAY('OCEAN'), 'MSC_ADAPTER', TRUE, JSON_ARRAY('TRACKING', 'RATES', 'CONTRACT_RATES', 'BOOKING', 'DOCUMENTS'), 'World leader in global container shipping and digital tracking integrations.'),
('HAPAG_LLOYD', 'Hapag-Lloyd', 'HLCU', JSON_ARRAY('OCEAN'), 'HAPAG_LLOYD_ADAPTER', TRUE, JSON_ARRAY('TRACKING', 'RATES', 'SPOT_RATES', 'BOOKING', 'DOCUMENTS'), 'Leading liner shipping company with extensive vessel networks and instant quote APIs.'),
('CMA_CGM', 'CMA CGM Group', 'CMDU', JSON_ARRAY('OCEAN'), 'CMA_CGM_ADAPTER', TRUE, JSON_ARRAY('TRACKING', 'RATES', 'CONTRACT_RATES', 'BOOKING', 'DOCUMENTS'), 'Global player in sea, land, air, and logistics solutions.'),
('ONE', 'Ocean Network Express (ONE)', 'ONEY', JSON_ARRAY('OCEAN'), 'ONE_ADAPTER', TRUE, JSON_ARRAY('TRACKING', 'RATES', 'BOOKING'), 'Major global container shipping carrier serving key trans-Pacific and Asia-Europe lanes.'),
('EVERGREEN', 'Evergreen Marine Corporation', 'EGLV', JSON_ARRAY('OCEAN'), 'EVERGREEN_ADAPTER', TRUE, JSON_ARRAY('TRACKING', 'RATES', 'BOOKING'), 'Global container transport provider with comprehensive worldwide shipping routes.'),
('COSCO', 'COSCO Shipping Lines', 'COSU', JSON_ARRAY('OCEAN'), 'COSCO_ADAPTER', TRUE, JSON_ARRAY('TRACKING', 'RATES', 'DOCUMENTS'), 'International integrated logistics enterprise providing comprehensive container services.')
ON DUPLICATE KEY UPDATE 
    name = VALUES(name),
    scac = VALUES(scac),
    modes = VALUES(modes),
    adapter_key = VALUES(adapter_key),
    supported_capabilities = VALUES(supported_capabilities),
    description = VALUES(description);

-- 2. Upgrade Carrier Integrations Table
-- Ensure carrier_provider_id, carrier_name, encrypted_credentials, credential_mask, last_success_at, last_failure_at, last_error exist

-- Check and add columns safely if they do not exist
SET @dbname = DATABASE();
SET @tablename = 'carrier_integrations';

-- carrier_provider_id
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = 'carrier_provider_id') > 0,
  'SELECT 1',
  'ALTER TABLE carrier_integrations ADD COLUMN carrier_provider_id BIGINT NULL AFTER org_id'
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- carrier_name
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = 'carrier_name') > 0,
  'SELECT 1',
  'ALTER TABLE carrier_integrations ADD COLUMN carrier_name VARCHAR(255) NULL AFTER carrier_scac'
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- encrypted_credentials
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = 'encrypted_credentials') > 0,
  'SELECT 1',
  'ALTER TABLE carrier_integrations ADD COLUMN encrypted_credentials TEXT NULL AFTER credentials_json'
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- credential_mask
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = 'credential_mask') > 0,
  'SELECT 1',
  'ALTER TABLE carrier_integrations ADD COLUMN credential_mask JSON NULL AFTER encrypted_credentials'
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- last_success_at
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = 'last_success_at') > 0,
  'SELECT 1',
  'ALTER TABLE carrier_integrations ADD COLUMN last_success_at DATETIME NULL AFTER last_synced_at'
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- last_failure_at
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = 'last_failure_at') > 0,
  'SELECT 1',
  'ALTER TABLE carrier_integrations ADD COLUMN last_failure_at DATETIME NULL AFTER last_success_at'
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- last_error
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = 'last_error') > 0,
  'SELECT 1',
  'ALTER TABLE carrier_integrations ADD COLUMN last_error TEXT NULL AFTER last_failure_at'
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- config_options
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND COLUMN_NAME = 'config_options') > 0,
  'SELECT 1',
  'ALTER TABLE carrier_integrations ADD COLUMN config_options JSON NULL AFTER capabilities'
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Unique constraint on (org_id, carrier_scac, environment)
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND INDEX_NAME = 'uk_org_carrier_env') > 0,
  'SELECT 1',
  'ALTER TABLE carrier_integrations ADD UNIQUE KEY uk_org_carrier_env (org_id, carrier_scac, environment)'
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Foreign key on carrier_provider_id
SET @preparedStatement = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tablename AND CONSTRAINT_NAME = 'fk_ci_provider') > 0,
  'SELECT 1',
  'ALTER TABLE carrier_integrations ADD CONSTRAINT fk_ci_provider FOREIGN KEY (carrier_provider_id) REFERENCES carrier_providers(id) ON DELETE SET NULL'
));
PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
