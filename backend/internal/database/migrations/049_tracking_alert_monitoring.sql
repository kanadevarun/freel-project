-- Migration 049: Tracking Alert Monitoring & Lifecycle Management (Task 17.5)

CREATE TABLE IF NOT EXISTS shipment_tracking_alerts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    shipment_id BIGINT NOT NULL,
    alert_key VARCHAR(150) NOT NULL,
    alert_type VARCHAR(100) NOT NULL,
    severity VARCHAR(30) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'OPEN',
    first_detected_at DATETIME NOT NULL,
    last_detected_at DATETIME NOT NULL,
    acknowledged_at DATETIME NULL,
    acknowledged_by BIGINT NULL,
    resolved_at DATETIME NULL,
    resolved_by BIGINT NULL,
    suppressed_at DATETIME NULL,
    suppressed_by BIGINT NULL,
    notification_count INT NOT NULL DEFAULT 0,
    last_notified_at DATETIME NULL,
    metadata JSON NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_org_shipment_alert_key (org_id, shipment_id, alert_key),
    INDEX idx_sta_org_shipment_status (org_id, shipment_id, status),
    INDEX idx_sta_org_status (org_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
