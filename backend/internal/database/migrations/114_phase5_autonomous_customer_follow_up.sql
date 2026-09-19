-- Migration 114: Phase 5 Task 5.4 - Autonomous Customer Follow-Up
-- Purpose: Schema support for AI-driven customer follow-up decisions, communication preferences, message versioning, and response tracking.

-- 1. Customer Communication Preferences Table
CREATE TABLE IF NOT EXISTS customer_communication_preferences (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    preferred_channel VARCHAR(32) NOT NULL DEFAULT 'EMAIL',
    opt_out TINYINT(1) NOT NULL DEFAULT 0,
    opt_out_reason TEXT NULL,
    contact_restrictions VARCHAR(255) NOT NULL DEFAULT 'NONE',
    business_hours_only TINYINT(1) NOT NULL DEFAULT 1,
    designated_contact_id BIGINT NULL,
    max_followups_per_incident INT NOT NULL DEFAULT 3,
    min_followup_interval_hours INT NOT NULL DEFAULT 24,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_org_customer_pref (org_id, customer_id),
    INDEX idx_cust_pref_optout (org_id, opt_out)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Customer Follow-Up Records Table (Message Versioning, Approvals, Responses)
CREATE TABLE IF NOT EXISTS customer_followup_records (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    contact_id BIGINT NULL,
    plan_id VARCHAR(64) NULL,
    step_id VARCHAR(64) NULL,
    event_type VARCHAR(64) NOT NULL,
    channel VARCHAR(32) NOT NULL DEFAULT 'EMAIL',
    recipient_email VARCHAR(255) NOT NULL,
    recipient_name VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    actual_facts JSON NULL,
    predictions JSON NULL,
    recommendations JSON NULL,
    full_body TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT',
    approval_id VARCHAR(64) NULL,
    approval_status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    idempotency_key VARCHAR(128) NOT NULL,
    sent_at DATETIME NULL,
    customer_response TEXT NULL,
    response_received_at DATETIME NULL,
    response_classification VARCHAR(64) NULL,
    stop_reason VARCHAR(128) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_org_followup_idemp (org_id, idempotency_key),
    INDEX idx_followup_org_cust (org_id, customer_id, created_at),
    INDEX idx_followup_plan (org_id, plan_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Enhance customers table with follow-up workflow state
ALTER TABLE customers
    ADD COLUMN IF NOT EXISTS followup_status VARCHAR(32) NOT NULL DEFAULT 'HEALTHY',
    ADD COLUMN IF NOT EXISTS last_followup_at DATETIME NULL,
    ADD COLUMN IF NOT EXISTS active_followup_plan_id VARCHAR(64) NULL;
