-- ═══════════════════════════════════════════════════════════════════════════
-- Migration 060: Carrier Rate Contracts, Rate Versions & Commercial Validity
-- Task 19.3 — Enterprise rate contracts, historical revisions, and renewal tracking
-- ═══════════════════════════════════════════════════════════════════════════

-- 1. Create rate_contracts table
CREATE TABLE IF NOT EXISTS rate_contracts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    contract_reference VARCHAR(100) NOT NULL,
    carrier_name VARCHAR(100) NOT NULL,
    carrier_code VARCHAR(20) NULL,
    contract_name VARCHAR(255) NOT NULL,
    contract_type VARCHAR(50) NOT NULL DEFAULT 'ANNUAL_SERVICE',
    transport_mode VARCHAR(50) NULL DEFAULT 'Ocean FCL',
    currency VARCHAR(10) NULL DEFAULT 'USD',
    effective_date DATE NOT NULL,
    expiry_date DATE NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
    renewal_status VARCHAR(30) NOT NULL DEFAULT 'NOT_STARTED',
    renewal_owner VARCHAR(255) NULL,
    notes TEXT NULL,
    created_by VARCHAR(255) NULL,
    updated_by VARCHAR(255) NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_rc_org_carrier (org_id, carrier_name),
    INDEX idx_rc_org_status (org_id, status),
    INDEX idx_rc_org_expiry (org_id, expiry_date),
    INDEX idx_rc_org_renewal (org_id, renewal_status),
    INDEX idx_rc_org_ref (org_id, contract_reference)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Extend rates table with contract linkage and versioning columns
ALTER TABLE rates
    ADD COLUMN IF NOT EXISTS contract_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS version_number INT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS version_status VARCHAR(30) NOT NULL DEFAULT 'CURRENT',
    ADD COLUMN IF NOT EXISTS supersedes_rate_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS superseded_by_rate_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS version_created_at DATETIME NULL;

-- Indexing versioning & contract lookups on rates
ALTER TABLE rates
    ADD INDEX IF NOT EXISTS idx_rates_org_contract (org_id, contract_id),
    ADD INDEX IF NOT EXISTS idx_rates_org_contract_version (org_id, contract_id, version_number),
    ADD INDEX IF NOT EXISTS idx_rates_org_version_status (org_id, version_status),
    ADD INDEX IF NOT EXISTS idx_rates_org_supersedes (org_id, supersedes_rate_id);

-- 3. Create rate_version_history table (Audit trail)
CREATE TABLE IF NOT EXISTS rate_version_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    rate_id BIGINT NOT NULL,
    version_number INT NOT NULL DEFAULT 1,
    action VARCHAR(50) NOT NULL DEFAULT 'RATE_VERSION_CREATED',
    previous_rate_id BIGINT NULL,
    new_rate_id BIGINT NULL,
    description TEXT NOT NULL,
    performed_by VARCHAR(255) NULL,
    metadata JSON NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_rvh_org_rate (org_id, rate_id),
    INDEX idx_rvh_org_action (org_id, action),
    INDEX idx_rvh_org_created (org_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
