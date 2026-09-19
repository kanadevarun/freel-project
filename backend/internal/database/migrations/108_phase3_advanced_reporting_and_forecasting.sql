-- ==============================================================================
-- Migration 108: Phase 3 Task 3.11 - Advanced Reporting and Forecasting
-- ==============================================================================

-- 1. Analytics Report Snapshots
-- Stores aggregated authoritative metrics alongside Python AI-generated forecasts and narratives.
CREATE TABLE IF NOT EXISTS analytics_report_snapshots (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    report_type VARCHAR(64) NOT NULL, -- 'OPERATIONAL_VOLUME', 'REVENUE_FINANCE', 'COMMERCIAL_FUNNEL', 'CONTRACT_COMPLIANCE'
    date_range VARCHAR(64) NOT NULL,  -- 'LAST_30D', 'LAST_90D', 'YTD', 'LAST_12M', 'CUSTOM'
    start_date DATETIME NOT NULL,
    end_date DATETIME NOT NULL,
    metrics_payload LONGTEXT NOT NULL,      -- Authoritative deterministic metrics calculated in Go
    forecast_payload LONGTEXT NULL,         -- Generated forecast from Python AI sidecar
    ai_narrative LONGTEXT NULL,             -- AI executive narrative and trend explanations
    confidence_score DECIMAL(5,2) NULL,     -- Forecast confidence score (e.g. 0.85)
    is_forecast_available TINYINT(1) NOT NULL DEFAULT 1,
    correlation_id VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_report_snapshots_org_type (org_id, report_type),
    INDEX idx_report_snapshots_created (org_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Report Export Requests
-- Tracks user requests for data export (CSV, JSON, PDF).
CREATE TABLE IF NOT EXISTS report_export_requests (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    report_type VARCHAR(64) NOT NULL,
    export_format VARCHAR(32) NOT NULL, -- 'CSV', 'JSON'
    status VARCHAR(32) NOT NULL DEFAULT 'COMPLETED', -- 'COMPLETED', 'PENDING', 'FAILED'
    export_content LONGTEXT NULL,
    file_name VARCHAR(255) NULL,
    correlation_id VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_report_exports_org (org_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Report Distribution Requests
-- Tracks distribution of reports to external stakeholders with mandatory HITL approval gating.
CREATE TABLE IF NOT EXISTS report_distribution_requests (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    report_type VARCHAR(64) NOT NULL,
    recipient_emails TEXT NOT NULL,
    distribution_channel VARCHAR(32) NOT NULL DEFAULT 'EMAIL', -- 'EMAIL', 'INTERNAL'
    approval_id BIGINT NULL,                                  -- References approval_requests(id)
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING_APPROVAL',   -- 'PENDING_APPROVAL', 'APPROVED', 'DISTRIBUTED', 'REJECTED'
    notes TEXT NULL,
    correlation_id VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_report_dist_org (org_id, status),
    INDEX idx_report_dist_approval (org_id, approval_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
