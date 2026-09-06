-- 063_rate_lifecycle_intelligence.sql
-- Task 19.6: Rate Lifecycle Intelligence, Expiry Monitoring & Commercial Impact Management

CREATE TABLE IF NOT EXISTS rate_lifecycle_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    rate_id BIGINT NULL,
    contract_id BIGINT NULL,
    event_type VARCHAR(64) NOT NULL,
    previous_status VARCHAR(32) NULL,
    current_status VARCHAR(32) NOT NULL,
    description TEXT NOT NULL,
    metadata JSON NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_rle_org_rate (org_id, rate_id),
    INDEX idx_rle_org_contract (org_id, contract_id),
    INDEX idx_rle_org_event (org_id, event_type),
    INDEX idx_rle_org_created (org_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS quotation_rate_risk_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    quotation_rate_snapshot_id BIGINT NULL,
    source_rate_id BIGINT NULL,
    source_contract_id BIGINT NULL,
    source_spot_rate_response_id BIGINT NULL,
    risk_type VARCHAR(64) NOT NULL,
    severity VARCHAR(16) NOT NULL DEFAULT 'WARNING', -- 'INFO', 'WARNING', 'CRITICAL'
    headline VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    recommended_action TEXT NULL,
    is_resolved BOOLEAN DEFAULT FALSE,
    resolved_by VARCHAR(128) NULL,
    resolved_at TIMESTAMP NULL,
    metadata JSON NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_qrre_org_quote (org_id, quotation_id),
    INDEX idx_qrre_org_resolved (org_id, is_resolved),
    INDEX idx_qrre_org_severity (org_id, severity),
    INDEX idx_qrre_org_rate (org_id, source_rate_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
