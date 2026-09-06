-- +migrate Up
-- 058_rate_management_foundation.sql
-- Rate Management Foundation & Core Rates Table

CREATE TABLE IF NOT EXISTS rates (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    rate_reference VARCHAR(100) NOT NULL,
    carrier_name VARCHAR(100) NOT NULL,
    carrier_code VARCHAR(20) NULL,
    service_provider VARCHAR(100) NULL,
    rate_type VARCHAR(30) NOT NULL DEFAULT 'SPOT',
    transport_mode VARCHAR(50) NOT NULL DEFAULT 'Ocean FCL',
    service_type VARCHAR(50) NOT NULL DEFAULT 'FCL',
    equipment_type VARCHAR(50) NULL DEFAULT '40GP',
    origin_port VARCHAR(255) NOT NULL,
    origin_code VARCHAR(20) NULL,
    destination_port VARCHAR(255) NOT NULL,
    destination_code VARCHAR(20) NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    base_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    effective_date DATE NOT NULL,
    expiry_date DATE NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
    carrier_reference VARCHAR(100) NULL,
    contract_reference VARCHAR(100) NULL,
    notes TEXT NULL,
    created_by VARCHAR(255) NULL,
    updated_by VARCHAR(255) NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_rates_org_status (org_id, status),
    INDEX idx_rates_org_carrier (org_id, carrier_name),
    INDEX idx_rates_org_lane (org_id, origin_port, destination_port),
    INDEX idx_rates_org_dates (org_id, effective_date, expiry_date),
    INDEX idx_rates_org_ref (org_id, rate_reference)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE IF EXISTS rates;
