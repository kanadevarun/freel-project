-- ==============================================================================
-- 126_email_delivery_tracking.sql
-- Canonical Migration: Persistent Email Dispatch, Delivery, Bounce, and Suppression Tracking
-- LogisticsHQ Task 2.3
-- ==============================================================================

CREATE TABLE IF NOT EXISTS email_messages (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    provider VARCHAR(50) NOT NULL DEFAULT 'AWS_SES' COMMENT 'AWS_SES, SMTP, etc.',
    message_id VARCHAR(150) NULL COMMENT 'External provider message identifier (e.g. SES Message-ID)',
    to_email VARCHAR(255) NOT NULL COMMENT 'Recipient email address',
    from_email VARCHAR(255) NOT NULL COMMENT 'Sender email address',
    subject VARCHAR(255) NOT NULL COMMENT 'Email subject line',
    body_preview VARCHAR(255) NOT NULL COMMENT 'Sanitized text preview of outgoing message',
    status VARCHAR(50) NOT NULL DEFAULT 'QUEUED' COMMENT 'QUEUED, ACCEPTED, SENT, DELIVERED, BOUNCED, COMPLAINED, REJECTED, FAILED',
    bounce_type VARCHAR(50) NULL COMMENT 'Permanent, Transient, Undetermined',
    bounce_sub_type VARCHAR(50) NULL COMMENT 'General, NoEmail, Suppressed, OnAccountSuppressionList, etc.',
    complaint_feedback_type VARCHAR(50) NULL COMMENT 'abuse, fraud, not-spam, virus, etc.',
    error_code VARCHAR(50) NULL COMMENT 'Normalized provider error code',
    error_message TEXT NULL COMMENT 'Sanitized error detail',
    idempotency_key VARCHAR(128) NOT NULL COMMENT 'Idempotency key to prevent duplicate transmission',
    correlation_id VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    delivered_at DATETIME NULL,
    bounced_at DATETIME NULL,
    complained_at DATETIME NULL,
    UNIQUE KEY uq_org_email_idem (org_id, idempotency_key),
    INDEX idx_email_message_id (message_id),
    INDEX idx_email_org_created (org_id, created_at),
    INDEX idx_email_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS email_suppressions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    email VARCHAR(255) NOT NULL,
    reason VARCHAR(50) NOT NULL COMMENT 'HARD_BOUNCE, COMPLAINT, UNSUBSCRIBED',
    source_message_id VARCHAR(150) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_org_suppressed_email (org_id, email),
    INDEX idx_suppressed_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
