-- ==============================================================================
-- 124_external_integration_foundation.sql
-- Canonical Migration: External Integration Foundation & Webhook Security Gateway
-- LogisticsHQ Task 2.1
-- ==============================================================================

CREATE TABLE IF NOT EXISTS external_integration_configs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    integration_type VARCHAR(50) NOT NULL COMMENT 'SMS, EMAIL, CARRIER_TRACKING, STORAGE, TEXTRACT, WEBHOOK',
    provider_name VARCHAR(50) NOT NULL COMMENT 'TWILIO, AWS_SES, MAERSK_API, AWS_S3, AWS_TEXTRACT, GENERIC',
    is_enabled TINYINT(1) NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'NOT_CONFIGURED' COMMENT 'ENABLED, DISABLED, NOT_CONFIGURED, CONFIG_INVALID, HEALTHY, DEGRADED, UNAVAILABLE',
    config_json LONGTEXT NULL COMMENT 'Non-secret configuration: regions, base URLs, timeouts, retry counts, from names',
    encrypted_secrets LONGTEXT NULL COMMENT 'Secure encrypted secrets or masked credential indicators',
    last_health_check DATETIME NULL,
    health_message TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_org_integration_provider (org_id, integration_type, provider_name),
    INDEX idx_org_integration (org_id, integration_type),
    INDEX idx_integration_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS external_webhook_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    provider_name VARCHAR(50) NOT NULL,
    provider_event_id VARCHAR(150) NULL,
    event_fingerprint VARCHAR(128) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    signature VARCHAR(255) NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'RECEIVED' COMMENT 'RECEIVED, VERIFIED, PROCESSED, DUPLICATE_IGNORED, REJECTED, DEAD_LETTER',
    rejection_reason VARCHAR(255) NULL,
    payload_preview TEXT NULL COMMENT 'Sanitized non-secret preview for audit/debug',
    correlation_id VARCHAR(100) NOT NULL,
    received_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at DATETIME NULL,
    UNIQUE KEY uq_provider_event_dedup (org_id, provider_name, provider_event_id),
    INDEX idx_webhook_fingerprint (org_id, event_fingerprint),
    INDEX idx_webhook_status (status, received_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS external_webhook_dead_letter (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    provider_name VARCHAR(50) NOT NULL,
    correlation_id VARCHAR(100) NOT NULL,
    error_code VARCHAR(50) NOT NULL,
    error_message TEXT NOT NULL,
    raw_headers TEXT NULL,
    payload_preview TEXT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    resolved TINYINT(1) NOT NULL DEFAULT 0,
    resolved_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_dead_letter_org (org_id, resolved),
    INDEX idx_dead_letter_corr (correlation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
