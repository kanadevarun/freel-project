-- Migration 090: Customer Follow-Up Assistant
-- Extends ai_recommendations with customer-specific follow-up fields, suggested owners,
-- editable message drafts, and links to idempotent internal follow-up tasks.

ALTER TABLE ai_recommendations 
    ADD COLUMN IF NOT EXISTS customer_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS customer_name VARCHAR(255) NULL,
    ADD COLUMN IF NOT EXISTS followup_type VARCHAR(64) NULL,
    ADD COLUMN IF NOT EXISTS suggested_owner_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS suggested_owner_name VARCHAR(128) NULL,
    ADD COLUMN IF NOT EXISTS draft_subject VARCHAR(255) NULL,
    ADD COLUMN IF NOT EXISTS draft_body TEXT NULL,
    ADD COLUMN IF NOT EXISTS draft_status VARCHAR(32) NOT NULL DEFAULT 'NONE',
    ADD COLUMN IF NOT EXISTS draft_generated_at DATETIME NULL,
    ADD COLUMN IF NOT EXISTS followup_task_id BIGINT NULL;

-- Idempotent internal follow-up task records
CREATE TABLE IF NOT EXISTS customer_followup_tasks (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    recommendation_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    source_type VARCHAR(64) NOT NULL,
    source_id BIGINT NOT NULL,
    source_reference VARCHAR(128) NOT NULL,
    followup_type VARCHAR(64) NOT NULL,
    title VARCHAR(255) NOT NULL,
    reason TEXT NOT NULL,
    suggested_action TEXT NOT NULL,
    priority VARCHAR(32) NOT NULL DEFAULT 'medium',
    assignee_id BIGINT NULL,
    assignee_name VARCHAR(128) NULL,
    due_date DATETIME NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    completed_at DATETIME NULL,
    created_by_id BIGINT NULL,
    created_by_name VARCHAR(128) NOT NULL DEFAULT 'SYSTEM',
    correlation_id VARCHAR(128) NOT NULL,
    notes TEXT NULL,
    UNIQUE KEY uq_rec_task (org_id, recommendation_id),
    INDEX idx_cft_org_customer (org_id, customer_id),
    INDEX idx_cft_org_status (org_id, status),
    INDEX idx_cft_org_assignee (org_id, assignee_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
