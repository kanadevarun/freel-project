-- ==============================================================================
-- 125_sms_delivery_tracking.sql
-- Canonical Migration: Persistent SMS Message Dispatch and Delivery Tracking
-- LogisticsHQ Task 2.2
-- ==============================================================================

CREATE TABLE IF NOT EXISTS sms_messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    provider VARCHAR(50) NOT NULL DEFAULT 'TWILIO' COMMENT 'TWILIO, etc.',
    message_sid VARCHAR(100) NULL COMMENT 'External provider message identifier (e.g. Twilio SM...)',
    to_phone VARCHAR(30) NOT NULL COMMENT 'Recipient E.164 formatted phone number',
    from_phone VARCHAR(30) NOT NULL COMMENT 'Sender identifier / phone number',
    body_preview VARCHAR(255) NOT NULL COMMENT 'Sanitized preview of outgoing message',
    status VARCHAR(50) NOT NULL DEFAULT 'QUEUED' COMMENT 'QUEUED, ACCEPTED, SENT, DELIVERED, FAILED, UNDELIVERED',
    error_code VARCHAR(50) NULL COMMENT 'Normalized provider error code',
    error_message TEXT NULL COMMENT 'Sanitized error detail',
    idempotency_key VARCHAR(128) NOT NULL COMMENT 'Idempotency key to prevent duplicate transmission',
    correlation_id VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    delivered_at DATETIME NULL,
    UNIQUE KEY uq_org_sms_idem (org_id, idempotency_key),
    INDEX idx_sms_message_sid (message_sid),
    INDEX idx_sms_org_created (org_id, created_at),
    INDEX idx_sms_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
