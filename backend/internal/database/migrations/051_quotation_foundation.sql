-- Migration: 051_quotation_foundation.sql
-- Description: Create quotations table as a first-class commercial entity with full lifecycle management

CREATE TABLE IF NOT EXISTS quotations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,

    quotation_number VARCHAR(50) NOT NULL,

    customer_id BIGINT NULL,
    customer_name VARCHAR(255) NULL,
    rfq_id BIGINT NULL,
    rfq_number VARCHAR(100) NULL,

    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT',

    origin VARCHAR(255) NULL,
    origin_code VARCHAR(20) NULL,
    destination VARCHAR(255) NULL,
    destination_code VARCHAR(20) NULL,

    service_type VARCHAR(100) NULL,
    transport_mode VARCHAR(100) NULL,

    currency VARCHAR(10) NOT NULL DEFAULT 'USD',

    subtotal DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    surcharges DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    taxes DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    valid_from DATE NULL,
    valid_until DATE NULL,

    sent_at DATETIME NULL,
    viewed_at DATETIME NULL,
    accepted_at DATETIME NULL,
    rejected_at DATETIME NULL,
    expired_at DATETIME NULL,

    notes TEXT NULL,

    created_by VARCHAR(255) NULL,
    updated_by VARCHAR(255) NULL,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uq_quotation_org_number (org_id, quotation_number),
    INDEX idx_quotation_org_status (org_id, status),
    INDEX idx_quotation_org_customer (org_id, customer_id),
    INDEX idx_quotation_org_valid_until (org_id, valid_until),
    INDEX idx_quotation_org_created (org_id, created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Quotation activity log for timeline
CREATE TABLE IF NOT EXISTS quotation_activity (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    activity_type VARCHAR(100) NOT NULL,
    description TEXT NULL,
    actor VARCHAR(255) NULL,
    metadata JSON NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_qact_org_quote_created (org_id, quotation_id, created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
