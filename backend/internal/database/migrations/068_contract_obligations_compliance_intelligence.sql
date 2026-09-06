-- ═══════════════════════════════════════════════════════════════════════════
-- Migration 068: Contract Obligations, Terms, Compliance & Operational Intelligence
-- Tasks 20.5 & 20.6
-- ═══════════════════════════════════════════════════════════════════════════

-- 1. Create contract_terms table
CREATE TABLE IF NOT EXISTS contract_terms (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    contract_id BIGINT NOT NULL,
    contract_version_id BIGINT NULL,
    term_category VARCHAR(64) NOT NULL DEFAULT 'COMMERCIAL', -- COMMERCIAL, PAYMENT, PRICING, LIABILITY, SERVICE_LEVEL, TERMINATION, RENEWAL, COMPLIANCE, OPERATIONAL, OTHER
    term_key VARCHAR(128) NOT NULL,
    term_title VARCHAR(255) NOT NULL,
    term_value TEXT NOT NULL,
    value_type VARCHAR(32) NOT NULL DEFAULT 'STRING', -- STRING, NUMBER, CURRENCY, PERCENTAGE, DATE, BOOLEAN
    currency VARCHAR(10) NULL,
    effective_date DATE NULL,
    expiry_date DATE NULL,
    display_order INT NOT NULL DEFAULT 0,
    is_critical BOOLEAN NOT NULL DEFAULT FALSE,
    created_by BIGINT NULL,
    updated_by BIGINT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_ct_org_contract (org_id, contract_id),
    INDEX idx_ct_category (org_id, term_category),
    INDEX idx_ct_critical (org_id, is_critical),
    INDEX idx_ct_version (org_id, contract_version_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Create contract_obligations table
CREATE TABLE IF NOT EXISTS contract_obligations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    contract_id BIGINT NOT NULL,
    contract_version_id BIGINT NULL,
    obligation_reference VARCHAR(64) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    obligation_type VARCHAR(64) NOT NULL DEFAULT 'OPERATIONAL', -- VOLUME_COMMITMENT, TRANSIT_TIME, SLA, DOCUMENTATION, INSURANCE, PAYMENT, RENEWAL_NOTICE, PRICING_REVIEW, FREE_DAYS, BOOKING_COMMITMENT, CUSTOMS, COMPLIANCE, OTHER
    category VARCHAR(64) NOT NULL DEFAULT 'GENERAL', -- SHIPMENT, CARRIER, CUSTOMER, RATE, COMPLIANCE, FINANCIAL
    responsible_party VARCHAR(64) NOT NULL DEFAULT 'CARRIER', -- CARRIER, CUSTOMER, SHIPPER, FORWARDER, VENDOR, INTERNAL
    owner VARCHAR(255) NULL,
    priority VARCHAR(32) NOT NULL DEFAULT 'MEDIUM', -- CRITICAL, HIGH, MEDIUM, LOW
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, DUE_SOON, AT_RISK, BREACHED, FULFILLED, WAIVED, CANCELLED
    effective_date DATE NULL,
    due_date DATE NULL,
    completion_date DATETIME NULL,
    is_recurring BOOLEAN NOT NULL DEFAULT FALSE,
    recurrence_type VARCHAR(32) NULL DEFAULT 'NONE', -- NONE, MONTHLY, QUARTERLY, ANNUALLY, PER_SHIPMENT
    target_value DECIMAL(15, 2) NULL,
    target_unit VARCHAR(64) NULL, -- TEU, DAYS, PERCENT, USD, DOCUMENTS
    current_value DECIMAL(15, 2) NULL DEFAULT 0.00,
    warning_threshold DECIMAL(15, 2) NULL,
    critical_threshold DECIMAL(15, 2) NULL,
    source_document_id VARCHAR(64) NULL,
    source_term_id BIGINT NULL,
    notes TEXT NULL,
    created_by BIGINT NULL,
    fulfilled_by BIGINT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_co_org_contract (org_id, contract_id),
    INDEX idx_co_status (org_id, status),
    INDEX idx_co_due (org_id, due_date),
    INDEX idx_co_priority (org_id, priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Create contract_compliance_events table
CREATE TABLE IF NOT EXISTS contract_compliance_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    contract_id BIGINT NOT NULL,
    contract_obligation_id BIGINT NULL,
    related_entity_type VARCHAR(64) NULL, -- SHIPMENT, BOOKING, RATE, QUOTATION, DOCUMENT, CERTIFICATE, PARTY
    related_entity_id BIGINT NULL,
    event_type VARCHAR(64) NOT NULL, -- SLA_BREACHED, REQUIRED_DOC_MISSING, VOLUME_MISSED, PERFORMANCE_DEFICIT, EXPIRY_WARNING, INSURANCE_EXPIRED, COMPLIANCE_FULFILLED, AUDIT_FLAG
    severity VARCHAR(32) NOT NULL DEFAULT 'ATTENTION', -- INFO, ATTENTION, WARNING, CRITICAL
    status VARCHAR(32) NOT NULL DEFAULT 'OPEN', -- OPEN, ACKNOWLEDGED, IN_PROGRESS, RESOLVED, WAIVED
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    detected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME NULL,
    resolved_by BIGINT NULL,
    resolution_notes TEXT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_cce_org_contract (org_id, contract_id),
    INDEX idx_cce_status (org_id, status),
    INDEX idx_cce_severity (org_id, severity),
    INDEX idx_cce_obligation (org_id, contract_obligation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. Create contract_compliance_requirements table
CREATE TABLE IF NOT EXISTS contract_compliance_requirements (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    contract_id BIGINT NOT NULL,
    requirement_type VARCHAR(64) NOT NULL DEFAULT 'INSURANCE', -- INSURANCE, CARRIER_CERTIFICATION, VENDOR_CERTIFICATION, REGULATORY, SAFETY, CUSTOMS, SLA, FINANCIAL
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    responsible_party VARCHAR(64) NOT NULL DEFAULT 'CARRIER',
    valid_from DATE NULL,
    valid_until DATE NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING', -- PENDING, COMPLIANT, NON_COMPLIANT, EXPIRED, WAIVED
    evidence_document_id VARCHAR(64) NULL,
    verification_date DATETIME NULL,
    verified_by BIGINT NULL,
    risk_severity VARCHAR(32) NOT NULL DEFAULT 'MEDIUM', -- CRITICAL, HIGH, MEDIUM, LOW
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_ccr_org_contract (org_id, contract_id),
    INDEX idx_ccr_status (org_id, status),
    INDEX idx_ccr_validity (org_id, valid_until)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
