-- Migration 095: Workflow Automation and Scheduled AI Jobs
-- Stores controlled AI automation definitions, execution history, next-run scheduling,
-- and links generated recommendations to automation and execution IDs with tenant isolation.

-- 1. Automations Definition Table
CREATE TABLE IF NOT EXISTS ai_automations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    automation_type VARCHAR(64) NOT NULL COMMENT 'DAILY_OVERDUE_INVOICE_REVIEW, DAILY_CONTRACT_DOCUMENT_EXPIRY_REVIEW, SHIPMENT_EXCEPTION_REVIEW, RFQ_QUOTATION_REVIEW, CUSTOMER_FOLLOWUP_REVIEW, DAILY_OPERATIONAL_SUMMARY',
    description TEXT NULL,
    is_enabled TINYINT(1) NOT NULL DEFAULT 1,
    schedule_type VARCHAR(32) NOT NULL DEFAULT 'DAILY' COMMENT 'DAILY, WEEKLY, HOURLY',
    schedule_time VARCHAR(16) NOT NULL DEFAULT '08:00' COMMENT 'HH:MM in 24h format',
    schedule_days VARCHAR(64) NULL DEFAULT 'MON,TUE,WED,THU,FRI' COMMENT 'Comma separated list of days for weekly',
    timezone VARCHAR(64) NOT NULL DEFAULT 'UTC',
    execution_window_minutes INT NOT NULL DEFAULT 60,
    configuration JSON NULL COMMENT 'Target module parameters and thresholds',
    target_modules JSON NULL COMMENT 'List of modules evaluated',
    last_execution_at DATETIME NULL,
    next_execution_at DATETIME NULL,
    last_execution_status VARCHAR(32) NULL COMMENT 'COMPLETED, FAILED, PARTIALLY_COMPLETED, RUNNING, CANCELLED, SKIPPED',
    last_error TEXT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    correlation_id VARCHAR(128) NULL,
    created_by BIGINT NOT NULL,
    updated_by BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_aiauto_org_enabled (org_id, is_enabled),
    INDEX idx_aiauto_org_type (org_id, automation_type),
    INDEX idx_aiauto_next_exec (is_enabled, next_execution_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Automation Execution History Table
CREATE TABLE IF NOT EXISTS ai_automation_executions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    automation_id BIGINT NOT NULL,
    org_id BIGINT NOT NULL,
    correlation_id VARCHAR(128) NOT NULL,
    trigger_type VARCHAR(32) NOT NULL DEFAULT 'SCHEDULED' COMMENT 'SCHEDULED, MANUAL',
    triggered_by_user_id BIGINT NULL COMMENT 'NULL if scheduled',
    status VARCHAR(32) NOT NULL DEFAULT 'QUEUED' COMMENT 'QUEUED, RUNNING, COMPLETED, PARTIALLY_COMPLETED, FAILED, CANCELLED, SKIPPED',
    queued_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME NULL,
    completed_at DATETIME NULL,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    records_reviewed INT NOT NULL DEFAULT 0,
    recommendations_created INT NOT NULL DEFAULT 0,
    recommendations_updated INT NOT NULL DEFAULT 0,
    summary_text TEXT NULL,
    error_message TEXT NULL,
    details JSON NULL COMMENT 'Sub-job breakdown, warnings, and metrics',
    retry_count INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_aiauto_exec_org (org_id, created_at DESC),
    INDEX idx_aiauto_exec_auto (automation_id, created_at DESC),
    INDEX idx_aiauto_exec_status (org_id, status),
    INDEX idx_aiauto_exec_corr (correlation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Extend ai_recommendations with automation and execution references
ALTER TABLE ai_recommendations ADD COLUMN IF NOT EXISTS automation_id BIGINT NULL;
ALTER TABLE ai_recommendations ADD COLUMN IF NOT EXISTS execution_id BIGINT NULL;
CREATE INDEX IF NOT EXISTS idx_ai_rec_automation ON ai_recommendations (org_id, automation_id);
CREATE INDEX IF NOT EXISTS idx_ai_rec_execution ON ai_recommendations (org_id, execution_id);
