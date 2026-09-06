-- Migration 080: Carrier Sync Jobs, Webhook Events & Integration Health Monitoring
-- Establishes structured sync history, webhook idempotency storage, and operational telemetry.

-- 1. Carrier Integration Sync Jobs
CREATE TABLE IF NOT EXISTS carrier_sync_jobs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    carrier_integration_id BIGINT NOT NULL,
    operation VARCHAR(50) NOT NULL DEFAULT 'TRACKING',
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME NULL,
    records_processed INT NOT NULL DEFAULT 0,
    records_created INT NOT NULL DEFAULT 0,
    records_updated INT NOT NULL DEFAULT 0,
    records_failed INT NOT NULL DEFAULT 0,
    error_code VARCHAR(100) NULL,
    error_message TEXT NULL,
    correlation_id VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_csj_org_int (org_id, carrier_integration_id),
    INDEX idx_csj_int_started (carrier_integration_id, started_at DESC),
    INDEX idx_csj_status (status),
    INDEX idx_csj_corr (correlation_id),
    CONSTRAINT fk_csj_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_csj_integration FOREIGN KEY (carrier_integration_id) REFERENCES carrier_integrations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Carrier Webhook Events
CREATE TABLE IF NOT EXISTS carrier_webhook_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    carrier_integration_id BIGINT NOT NULL,
    carrier_scac VARCHAR(20) NOT NULL,
    provider_event_id VARCHAR(150) NULL,
    event_type VARCHAR(100) NOT NULL,
    event_fingerprint VARCHAR(128) NOT NULL,
    received_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at DATETIME NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    error_message TEXT NULL,
    correlation_id VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_cwe_fingerprint (org_id, event_fingerprint),
    INDEX idx_cwe_org_int (org_id, carrier_integration_id),
    INDEX idx_cwe_status (status),
    INDEX idx_cwe_provider_evt (provider_event_id),
    CONSTRAINT fk_cwe_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_cwe_integration FOREIGN KEY (carrier_integration_id) REFERENCES carrier_integrations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
