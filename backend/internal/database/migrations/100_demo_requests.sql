-- Migration 100: Demo Requests Table
-- Stores inbound demo requests from the public website's "Request Demo" form.
-- These are viewed and managed by internal staff via SPortal.

CREATE TABLE IF NOT EXISTS demo_requests (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    full_name       VARCHAR(255) NOT NULL,
    email           VARCHAR(255) NOT NULL,
    company_name    VARCHAR(255) NOT NULL,
    phone           VARCHAR(64)  DEFAULT NULL,
    country         VARCHAR(128) DEFAULT NULL,
    company_size    VARCHAR(64)  DEFAULT NULL COMMENT 'e.g. 1-10, 11-50, 51-200, 201-500, 500+',
    message         TEXT         DEFAULT NULL COMMENT 'Optional message from the prospect',
    shipment_volume VARCHAR(64)  DEFAULT NULL COMMENT 'Monthly shipment volume range',
    services        TEXT         DEFAULT NULL COMMENT 'Comma-separated list of interested services',
    status          VARCHAR(32)  NOT NULL DEFAULT 'NEW' COMMENT 'NEW, CONTACTED, QUALIFIED, CONVERTED, DISMISSED',
    notes           TEXT         DEFAULT NULL COMMENT 'Internal staff notes',
    assigned_to     VARCHAR(255) DEFAULT NULL COMMENT 'Staff member handling the request',
    source          VARCHAR(64)  DEFAULT 'WEBSITE' COMMENT 'Where the request originated',
    ip_address      VARCHAR(64)  DEFAULT NULL,
    user_agent      TEXT         DEFAULT NULL,
    contacted_at    DATETIME     DEFAULT NULL,
    converted_at    DATETIME     DEFAULT NULL,
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_demo_requests_status (status),
    INDEX idx_demo_requests_email (email),
    INDEX idx_demo_requests_created (created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
