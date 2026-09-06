-- Migration 048: Real-Time Shipment Tracking Positions & Telemetry
-- Authoritative database schema for storing geographic positions, coordinates, speed, heading, and data freshness.

CREATE TABLE IF NOT EXISTS shipment_tracking_positions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    shipment_id BIGINT NOT NULL,
    vessel_name VARCHAR(255) NULL,
    latitude DECIMAL(10, 6) NOT NULL,
    longitude DECIMAL(10, 6) NOT NULL,
    speed_knots DECIMAL(6, 2) NOT NULL DEFAULT 0.00,
    heading_degrees DECIMAL(6, 2) NOT NULL DEFAULT 0.00,
    location_name VARCHAR(255) NULL,
    tracking_source VARCHAR(100) NOT NULL DEFAULT 'CARRIER_API',
    data_freshness VARCHAR(50) NOT NULL DEFAULT 'RECENT',
    recorded_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_stp_org_shipment (org_id, shipment_id, recorded_at DESC),
    INDEX idx_stp_recorded (recorded_at DESC),
    CONSTRAINT fk_stp_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_stp_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
