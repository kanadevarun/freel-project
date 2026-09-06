-- 062_rate_quotation_integration.sql
-- Task 19.5: Rate-to-Quotation Integration & Commercial Rate Selection

CREATE TABLE IF NOT EXISTS quotation_rate_selections (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    rate_id BIGINT NULL,
    spot_rate_request_id BIGINT NULL,
    spot_rate_response_id BIGINT NULL,
    rate_source_type VARCHAR(32) NOT NULL DEFAULT 'MANAGED_RATE', -- MANAGED_RATE, SPOT_RATE, CUSTOM_RATE
    selected_by VARCHAR(128) NULL,
    selected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    notes TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_qrs_org_quotation (org_id, quotation_id),
    INDEX idx_qrs_org_rate (org_id, rate_id),
    INDEX idx_qrs_org_spot_resp (org_id, spot_rate_response_id),
    INDEX idx_qrs_org_active (org_id, quotation_id, is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS quotation_rate_snapshots (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    quotation_rate_selection_id BIGINT NOT NULL,
    source_rate_id BIGINT NULL,
    source_rate_version INT NULL,
    source_contract_id BIGINT NULL,
    source_spot_rate_request_id BIGINT NULL,
    source_spot_rate_response_id BIGINT NULL,
    carrier_name VARCHAR(128) NOT NULL,
    carrier_reference VARCHAR(128) NULL,
    transport_mode VARCHAR(64) NOT NULL,
    service_type VARCHAR(64) NULL,
    equipment_type VARCHAR(64) NULL,
    origin VARCHAR(128) NOT NULL,
    destination VARCHAR(128) NOT NULL,
    currency VARCHAR(8) NOT NULL DEFAULT 'USD',
    base_rate DECIMAL(14, 2) NOT NULL DEFAULT 0.00,
    additional_charges DECIMAL(14, 2) NOT NULL DEFAULT 0.00,
    commercial_total DECIMAL(14, 2) NOT NULL DEFAULT 0.00,
    pricing_snapshot JSON NOT NULL,
    valid_from DATE NULL,
    valid_until DATE NULL,
    snapshot_created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(128) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_qrsnap_org_quotation (org_id, quotation_id),
    INDEX idx_qrsnap_org_selection (org_id, quotation_rate_selection_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS quotation_rate_selection_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    event_type VARCHAR(64) NOT NULL, -- RATE_SELECTED, RATE_REPLACED, RATE_SNAPSHOT_CREATED, RATE_EXPIRED, SPOT_RATE_SELECTED, RATE_REMOVED
    previous_selection_id BIGINT NULL,
    new_selection_id BIGINT NULL,
    description TEXT NOT NULL,
    metadata JSON NULL,
    performed_by VARCHAR(128) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_qrsh_org_quotation (org_id, quotation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
