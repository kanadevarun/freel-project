-- Migration 099: Phase 3 Business Automation and Operational Intelligence Foundation
-- Extends ai_automations and ai_automation_executions with event-driven triggers,
-- execution step lifecycle, idempotency, approval & action integration,
-- and creates ai_operational_insights for durable deterministic operational intelligence.

-- 1. Extend ai_automations for event triggers, scope, approval policies and retry controls
ALTER TABLE ai_automations 
    ADD COLUMN IF NOT EXISTS trigger_type VARCHAR(64) NOT NULL DEFAULT 'SCHEDULED' COMMENT 'SCHEDULED, MANUAL, RECORD_CREATED, RECORD_UPDATED, STATUS_CHANGED, MILESTONE_MISSED, SHIPMENT_EXCEPTION_DETECTED, INVOICE_OVERDUE, CONTRACT_EXPIRING, RFQ_DEADLINE_APPROACHING, APPROVAL_RETURNED',
    ADD COLUMN IF NOT EXISTS trigger_config JSON NULL COMMENT 'Event filter parameters, status transitions, day thresholds',
    ADD COLUMN IF NOT EXISTS scope VARCHAR(64) NOT NULL DEFAULT 'ORGANIZATION' COMMENT 'ORGANIZATION, REGIONAL, BRANCH, TEAM',
    ADD COLUMN IF NOT EXISTS approval_policy VARCHAR(64) NOT NULL DEFAULT 'ALWAYS_REQUIRE' COMMENT 'ALWAYS_REQUIRE, AUTO_IF_LOW_RISK, MANUAL_ONLY',
    ADD COLUMN IF NOT EXISTS allowed_actions JSON NULL COMMENT 'Whitelisted action names allowed by this automation',
    ADD COLUMN IF NOT EXISTS owner_team VARCHAR(128) NOT NULL DEFAULT 'OPERATIONS' COMMENT 'Operations, Logistics, Finance, Sales, Legal',
    ADD COLUMN IF NOT EXISTS priority VARCHAR(32) NOT NULL DEFAULT 'MEDIUM' COMMENT 'LOW, MEDIUM, HIGH, CRITICAL',
    ADD COLUMN IF NOT EXISTS last_successful_run DATETIME NULL,
    ADD COLUMN IF NOT EXISTS last_failed_run DATETIME NULL,
    ADD COLUMN IF NOT EXISTS retry_policy JSON NULL COMMENT 'Backoff configuration and retry rules',
    ADD COLUMN IF NOT EXISTS max_execution_duration_sec INT NOT NULL DEFAULT 300;

CREATE INDEX IF NOT EXISTS idx_aiauto_trigger_type ON ai_automations (org_id, trigger_type, is_enabled);

-- 2. Extend ai_automation_executions for granular lifecycle, input references, idempotency and action tracking
ALTER TABLE ai_automation_executions
    ADD COLUMN IF NOT EXISTS trigger_event VARCHAR(64) NULL COMMENT 'Specific event name that triggered execution',
    ADD COLUMN IF NOT EXISTS input_record_ref VARCHAR(128) NULL COMMENT 'e.g. shipment:123, invoice:456',
    ADD COLUMN IF NOT EXISTS current_step VARCHAR(64) NOT NULL DEFAULT 'COMPLETED' COMMENT 'PENDING, QUEUED, RUNNING, EVALUATING_RULES, GENERATING_INSIGHTS, WAITING_FOR_APPROVAL, EXECUTING_ACTION, COMPLETED, FAILED, CANCELLED',
    ADD COLUMN IF NOT EXISTS step_results JSON NULL COMMENT 'Intermediate step payloads and execution traces',
    ADD COLUMN IF NOT EXISTS approval_id BIGINT NULL COMMENT 'Reference to approvals table if HITL was required',
    ADD COLUMN IF NOT EXISTS action_id BIGINT NULL COMMENT 'Reference to centralized action audit / idempotency entry',
    ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(128) NULL;

CREATE INDEX IF NOT EXISTS idx_aiauto_exec_idempotency ON ai_automation_executions (org_id, idempotency_key);
CREATE INDEX IF NOT EXISTS idx_aiauto_exec_lifecycle ON ai_automation_executions (org_id, status, current_step);
CREATE INDEX IF NOT EXISTS idx_aiauto_exec_input_ref ON ai_automation_executions (org_id, input_record_ref);

-- 3. Create ai_operational_insights for durable first-class operational intelligence
CREATE TABLE IF NOT EXISTS ai_operational_insights (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    automation_id BIGINT NULL,
    execution_id BIGINT NULL,
    source_module VARCHAR(64) NOT NULL COMMENT 'shipments, invoices, contracts, rfq, customers',
    source_record_id BIGINT NOT NULL,
    source_record_ref VARCHAR(128) NULL,
    insight_type VARCHAR(64) NOT NULL,
    severity VARCHAR(32) NOT NULL DEFAULT 'MEDIUM' COMMENT 'INFO, LOW, MEDIUM, HIGH, CRITICAL',
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    evidence JSON NULL,
    detection_rule VARCHAR(128) NOT NULL,
    confidence DECIMAL(5,4) NOT NULL DEFAULT 1.0000,
    priority VARCHAR(32) NOT NULL DEFAULT 'MEDIUM' COMMENT 'LOW, MEDIUM, HIGH, CRITICAL',
    risk_level VARCHAR(32) NOT NULL DEFAULT 'LOW' COMMENT 'LOW, MEDIUM, HIGH, SEVERE',
    data_freshness VARCHAR(64) NOT NULL DEFAULT 'REAL_TIME' COMMENT 'REAL_TIME, RECENT, HISTORICAL',
    recommended_next_step TEXT NULL,
    is_approval_required TINYINT(1) NOT NULL DEFAULT 1,
    recommended_action_type VARCHAR(64) NULL,
    action_payload JSON NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE' COMMENT 'ACTIVE, ACKNOWLEDGED, ACTIONED, DISMISSED',
    action_id BIGINT NULL,
    approval_id BIGINT NULL,
    correlation_id VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_op_insights_org_status (org_id, status, created_at DESC),
    INDEX idx_op_insights_source (org_id, source_module, source_record_id),
    INDEX idx_op_insights_exec (org_id, execution_id),
    INDEX idx_op_insights_corr (correlation_id),
    INDEX idx_op_insights_rule (org_id, detection_rule, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
