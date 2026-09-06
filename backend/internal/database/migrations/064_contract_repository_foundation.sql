-- ═══════════════════════════════════════════════════════════════════════════
-- Migration 064: Contract Repository & Core Agreement Management Foundation
-- Task 20.1
-- ═══════════════════════════════════════════════════════════════════════════

-- 1. Create contract_parties table
CREATE TABLE IF NOT EXISTS contract_parties (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    party_name VARCHAR(255) NOT NULL,
    party_type VARCHAR(50) NOT NULL, -- e.g. 'CUSTOMER', 'CARRIER', 'VENDOR', 'OTHER'
    customer_id BIGINT NULL,         -- Soft reference to customers.id
    carrier_id BIGINT NULL,          -- Soft reference to carriers.id
    vendor_id BIGINT NULL,
    contact_name VARCHAR(255) NULL,
    contact_email VARCHAR(255) NULL,
    contact_phone VARCHAR(50) NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_cp_org_type (org_id, party_type),
    INDEX idx_cp_org_customer (org_id, customer_id),
    INDEX idx_cp_org_carrier (org_id, carrier_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Create contracts table
CREATE TABLE IF NOT EXISTS contracts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    contract_reference VARCHAR(100) NOT NULL,
    contract_name VARCHAR(255) NOT NULL,
    contract_type VARCHAR(50) NOT NULL,    -- e.g. 'CARRIER_AGREEMENT', 'CUSTOMER_SLA'
    party_id BIGINT NOT NULL,              -- Reference to contract_parties.id
    party_name VARCHAR(255) NOT NULL,      -- Denormalized for quick reads
    transport_mode VARCHAR(50) NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT',
    currency VARCHAR(10) NULL DEFAULT 'USD',
    contract_value DECIMAL(15, 2) NULL DEFAULT 0.00,
    effective_date DATE NULL,
    expiry_date DATE NULL,
    owner VARCHAR(255) NULL,
    description TEXT NULL,
    notes TEXT NULL,
    created_by VARCHAR(255) NULL,
    updated_by VARCHAR(255) NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    archived_at DATETIME NULL,

    INDEX idx_c_org_party (org_id, party_id),
    INDEX idx_c_org_status (org_id, status),
    INDEX idx_c_org_ref (org_id, contract_reference),
    INDEX idx_c_org_expiry (org_id, expiry_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Create contract_lifecycle_events table
CREATE TABLE IF NOT EXISTS contract_lifecycle_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    contract_id BIGINT NOT NULL,
    previous_status VARCHAR(50) NULL,
    new_status VARCHAR(50) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    description TEXT NULL,
    performed_by VARCHAR(255) NULL,
    metadata JSON NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_cle_org_contract (org_id, contract_id),
    INDEX idx_cle_org_event (org_id, event_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
