-- Migration 047: Shipment Financial Operations & Cost Management
-- Authoritative database schema for tracking operational costs, line items, and profitability adjustments.

CREATE TABLE IF NOT EXISTS shipment_financial_charges (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    shipment_id BIGINT NOT NULL,
    booking_id BIGINT NULL,
    rfq_id BIGINT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'OTHER',
    charge_type VARCHAR(20) NOT NULL DEFAULT 'COST',
    description VARCHAR(255) NOT NULL,
    vendor_name VARCHAR(255) NULL,
    estimated_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    actual_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    reference_number VARCHAR(100) NULL,
    charge_date DATETIME NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ESTIMATED',
    notes TEXT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_sfc_org_shipment (org_id, shipment_id),
    INDEX idx_sfc_org_cat (org_id, category),
    INDEX idx_sfc_status (org_id, status),
    CONSTRAINT fk_sfc_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_sfc_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
