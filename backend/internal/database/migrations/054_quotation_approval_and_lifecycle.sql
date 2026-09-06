-- Migration: 054_quotation_approval_and_lifecycle.sql
-- Description: Add approval workflow fields, quotation approval history, and public view tracking

-- Extend quotations table with lifecycle and audit fields
ALTER TABLE quotations
    ADD COLUMN IF NOT EXISTS submitted_for_review_at DATETIME NULL AFTER notes,
    ADD COLUMN IF NOT EXISTS submitted_for_review_by VARCHAR(255) NULL AFTER submitted_for_review_at,
    ADD COLUMN IF NOT EXISTS approved_at DATETIME NULL AFTER submitted_for_review_by,
    ADD COLUMN IF NOT EXISTS approved_by VARCHAR(255) NULL AFTER approved_at,
    ADD COLUMN IF NOT EXISTS approval_notes TEXT NULL AFTER approved_by,
    ADD COLUMN IF NOT EXISTS changes_requested_at DATETIME NULL AFTER approval_notes,
    ADD COLUMN IF NOT EXISTS changes_requested_by VARCHAR(255) NULL AFTER changes_requested_at,
    ADD COLUMN IF NOT EXISTS changes_requested_reason TEXT NULL AFTER changes_requested_by,
    ADD COLUMN IF NOT EXISTS sent_by VARCHAR(255) NULL AFTER sent_at,
    ADD COLUMN IF NOT EXISTS first_viewed_at DATETIME NULL AFTER viewed_at,
    ADD COLUMN IF NOT EXISTS last_viewed_at DATETIME NULL AFTER first_viewed_at,
    ADD COLUMN IF NOT EXISTS view_count INT NOT NULL DEFAULT 0 AFTER last_viewed_at,
    ADD COLUMN IF NOT EXISTS declined_at DATETIME NULL AFTER accepted_at,
    ADD COLUMN IF NOT EXISTS declined_reason TEXT NULL AFTER declined_at,
    ADD COLUMN IF NOT EXISTS cancelled_at DATETIME NULL AFTER expired_at,
    ADD COLUMN IF NOT EXISTS cancelled_by VARCHAR(255) NULL AFTER cancelled_at,
    ADD COLUMN IF NOT EXISTS cancelled_reason TEXT NULL AFTER cancelled_by;

-- Quotation approval history table for audit trails
CREATE TABLE IF NOT EXISTS quotation_approval_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    action VARCHAR(50) NOT NULL,
    previous_status VARCHAR(50) NOT NULL,
    new_status VARCHAR(50) NOT NULL,
    actor_user_id BIGINT NULL,
    actor_name VARCHAR(255) NULL,
    comments TEXT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_qah_org_quote_created (org_id, quotation_id, created_at DESC),
    INDEX idx_qah_org_quote (org_id, quotation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Quotation public views tracking for customer views
CREATE TABLE IF NOT EXISTS quotation_public_views (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    viewer_name VARCHAR(255) NULL,
    viewer_email VARCHAR(255) NULL,
    ip_address VARCHAR(100) NULL,
    user_agent VARCHAR(500) NULL,
    viewed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_qpv_org_quote (org_id, quotation_id, viewed_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
