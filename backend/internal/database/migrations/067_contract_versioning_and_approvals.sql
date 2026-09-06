-- ═══════════════════════════════════════════════════════════════════════════
-- Migration 067: Contract Versioning, Amendments & Approval Workflow
-- Task 20.4: Immutable version history, structured amendments, and approval governance
-- ═══════════════════════════════════════════════════════════════════════════

-- 1. Contract Versions
CREATE TABLE IF NOT EXISTS contract_versions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    contract_id BIGINT NOT NULL,
    version_number INT NOT NULL,
    version_label VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT', -- DRAFT, PENDING_APPROVAL, APPROVED, EFFECTIVE, SUPERSEDED, REJECTED
    effective_date DATE NULL,
    expiry_date DATE NULL,
    contract_snapshot JSON NOT NULL,
    change_summary TEXT NULL,
    created_by BIGINT NULL,
    approved_by BIGINT NULL,
    approved_at DATETIME NULL,
    superseded_at DATETIME NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY unique_contract_version_per_org (org_id, contract_id, version_number),
    INDEX idx_cv_org_contract (org_id, contract_id),
    INDEX idx_cv_status (org_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Contract Amendments
CREATE TABLE IF NOT EXISTS contract_amendments (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    contract_id BIGINT NOT NULL,
    base_version_id BIGINT NULL,
    amendment_reference VARCHAR(64) NOT NULL,
    amendment_type VARCHAR(64) NOT NULL DEFAULT 'COMMERCIAL_TERMS', -- RATE_REVISION, SCOPE_EXTENSION, CLAUSE_UPDATE, COMMERCIAL_TERMS, VALIDITY_EXTENSION
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    change_summary TEXT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT', -- DRAFT, SUBMITTED, UNDER_REVIEW, APPROVED, REJECTED, CANCELLED, IMPLEMENTED
    proposed_effective_date DATE NULL,
    created_by BIGINT NULL,
    submitted_at DATETIME NULL,
    approved_at DATETIME NULL,
    rejected_at DATETIME NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY unique_amendment_ref_per_org (org_id, amendment_reference),
    INDEX idx_ca_org_contract (org_id, contract_id),
    INDEX idx_ca_status (org_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Contract Amendment Field Changes
CREATE TABLE IF NOT EXISTS contract_amendment_changes (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    amendment_id BIGINT NOT NULL,
    field_name VARCHAR(128) NOT NULL,
    previous_value TEXT NULL,
    proposed_value TEXT NULL,
    change_type VARCHAR(32) NOT NULL DEFAULT 'MODIFY', -- ADD, MODIFY, REMOVE
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_cac_org_amendment (org_id, amendment_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. Contract Approval Requests
CREATE TABLE IF NOT EXISTS contract_approval_requests (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    contract_id BIGINT NOT NULL,
    version_id BIGINT NULL,
    amendment_id BIGINT NULL,
    approval_type VARCHAR(32) NOT NULL DEFAULT 'AMENDMENT', -- VERSION, AMENDMENT
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING', -- PENDING, APPROVED, REJECTED, CANCELLED
    requested_by BIGINT NULL,
    assigned_to BIGINT NULL,
    decision_by BIGINT NULL,
    decision_comment TEXT NULL,
    requested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    decided_at DATETIME NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_car_org_contract (org_id, contract_id),
    INDEX idx_car_status (org_id, status),
    INDEX idx_car_version (org_id, version_id),
    INDEX idx_car_amendment (org_id, amendment_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. Safe idempotent baseline version generation for existing contracts
INSERT IGNORE INTO contract_versions (
    org_id,
    contract_id,
    version_number,
    version_label,
    status,
    effective_date,
    expiry_date,
    contract_snapshot,
    change_summary,
    created_by,
    created_at,
    updated_at
)
SELECT 
    c.org_id,
    c.id AS contract_id,
    1 AS version_number,
    'v1.0 Baseline' AS version_label,
    CASE 
        WHEN c.status = 'ACTIVE' THEN 'EFFECTIVE'
        WHEN c.status = 'EXPIRED' THEN 'SUPERSEDED'
        ELSE 'DRAFT'
    END AS status,
    c.effective_date,
    c.expiry_date,
    JSON_OBJECT(
        'contract_id', c.id,
        'contract_reference', c.contract_reference,
        'contract_name', c.contract_name,
        'contract_type', c.contract_type,
        'status', c.status,
        'transport_mode', c.transport_mode,
        'effective_date', c.effective_date,
        'expiry_date', c.expiry_date,
        'contract_value', c.contract_value,
        'currency', c.currency,
        'owner', c.owner,
        'description', c.description,
        'baseline_note', 'Auto-initialized v1.0 baseline snapshot'
    ) AS contract_snapshot,
    'Initial contract baseline record' AS change_summary,
    1 AS created_by,
    c.created_at,
    c.updated_at
FROM contracts c
WHERE NOT EXISTS (
    SELECT 1 FROM contract_versions cv 
    WHERE cv.org_id = c.org_id AND cv.contract_id = c.id
);
