-- Migration 096: Human-in-the-Loop Approval Expansion
-- Standardizes approval lifecycle, decision history, execution tracking, action preview, and audit compliance.

CREATE TABLE IF NOT EXISTS approval_decisions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    approval_id BIGINT NOT NULL,
    action_name VARCHAR(100) NULL,
    decision VARCHAR(50) NOT NULL, -- 'APPROVE', 'REJECT', 'RETURN_FOR_CHANGES', 'CANCEL', 'EXPIRE'
    actor_id BIGINT NULL,
    actor_name VARCHAR(100) NOT NULL,
    reason TEXT NULL,
    notes TEXT NULL,
    correlation_id VARCHAR(100) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_appdec_org_app (org_id, approval_id),
    INDEX idx_appdec_corr (correlation_id),
    INDEX idx_appdec_decision (decision)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add execution status, source record tracking, preview metadata, and return-for-changes support to approval_requests
ALTER TABLE approval_requests
    ADD COLUMN IF NOT EXISTS execution_status VARCHAR(50) NOT NULL DEFAULT 'NOT_STARTED',
    ADD COLUMN IF NOT EXISTS execution_result LONGTEXT NULL,
    ADD COLUMN IF NOT EXISTS execution_error TEXT NULL,
    ADD COLUMN IF NOT EXISTS execution_retries INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS source_module VARCHAR(50) NULL,
    ADD COLUMN IF NOT EXISTS source_record_type VARCHAR(50) NULL,
    ADD COLUMN IF NOT EXISTS source_record_id VARCHAR(100) NULL,
    ADD COLUMN IF NOT EXISTS source_record_snapshot LONGTEXT NULL,
    ADD COLUMN IF NOT EXISTS evidence TEXT NULL,
    ADD COLUMN IF NOT EXISTS impact_summary TEXT NULL,
    ADD COLUMN IF NOT EXISTS is_reversible TINYINT(1) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS external_communication TINYINT(1) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS required_approval_level VARCHAR(50) NOT NULL DEFAULT 'MANAGER',
    ADD COLUMN IF NOT EXISTS returned_by VARCHAR(100) NULL,
    ADD COLUMN IF NOT EXISTS returned_at DATETIME NULL,
    ADD COLUMN IF NOT EXISTS returned_reason TEXT NULL;
