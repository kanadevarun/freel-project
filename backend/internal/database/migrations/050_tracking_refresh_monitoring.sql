-- Migration: 050_tracking_refresh_monitoring.sql
-- Description: Create shipment_tracking_refresh_runs table for persistent tracking refresh orchestration

CREATE TABLE IF NOT EXISTS shipment_tracking_refresh_runs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    shipment_id BIGINT NOT NULL,
    provider_name VARCHAR(100) NULL,
    provider_type VARCHAR(50) NULL,
    trigger_type VARCHAR(50) NOT NULL DEFAULT 'MANUAL',
    status VARCHAR(50) NOT NULL DEFAULT 'STARTED',
    started_at DATETIME NOT NULL,
    completed_at DATETIME NULL,
    new_positions INT NOT NULL DEFAULT 0,
    new_events INT NOT NULL DEFAULT 0,
    data_freshness VARCHAR(50) NULL,
    used_fallback BOOLEAN NOT NULL DEFAULT FALSE,
    error_message TEXT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_trkr_org_shipment_started (org_id, shipment_id, started_at DESC),
    INDEX idx_trkr_org_status_started (org_id, status, started_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
