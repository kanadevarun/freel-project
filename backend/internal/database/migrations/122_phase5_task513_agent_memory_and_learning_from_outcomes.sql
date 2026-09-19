-- ==============================================================================
-- Migration 122: Agent Memory and Learning from Outcomes (Phase 5 Task 5.13)
-- Safe, persistent, tenant-isolated AI memory and outcome learning for LogisticsHQ.
-- ==============================================================================

-- 1. Extend ai_memory_items with Phase 5 Autonomy Learning columns if not present
ALTER TABLE ai_memory_items
    ADD COLUMN IF NOT EXISTS category VARCHAR(64) NOT NULL DEFAULT 'OPERATIONAL' AFTER memory_type,
    ADD COLUMN IF NOT EXISTS outcome_id VARCHAR(64) NULL AFTER structured_value,
    ADD COLUMN IF NOT EXISTS entity_type VARCHAR(64) NULL AFTER outcome_id,
    ADD COLUMN IF NOT EXISTS entity_id VARCHAR(64) NULL AFTER entity_type,
    ADD COLUMN IF NOT EXISTS recency_weight DECIMAL(5,2) NOT NULL DEFAULT 1.00 AFTER confidence,
    ADD COLUMN IF NOT EXISTS times_observed INT NOT NULL DEFAULT 1 AFTER recency_weight,
    ADD COLUMN IF NOT EXISTS times_used INT NOT NULL DEFAULT 0 AFTER times_observed,
    ADD COLUMN IF NOT EXISTS success_count INT NOT NULL DEFAULT 1 AFTER times_used,
    ADD COLUMN IF NOT EXISTS failure_count INT NOT NULL DEFAULT 0 AFTER success_count,
    ADD COLUMN IF NOT EXISTS is_stale TINYINT(1) NOT NULL DEFAULT 0 AFTER failure_count,
    ADD COLUMN IF NOT EXISTS invalidated_at DATETIME NULL AFTER is_stale,
    ADD COLUMN IF NOT EXISTS invalidation_reason TEXT NULL AFTER invalidated_at,
    ADD COLUMN IF NOT EXISTS invalidated_by_id BIGINT NULL AFTER invalidation_reason,
    ADD COLUMN IF NOT EXISTS conflict_status VARCHAR(32) NOT NULL DEFAULT 'NONE' AFTER invalidated_by_id,
    ADD COLUMN IF NOT EXISTS superseded_by_id BIGINT NULL AFTER conflict_status,
    ADD COLUMN IF NOT EXISTS provenance_type VARCHAR(64) NOT NULL DEFAULT 'SYSTEM_DERIVED' AFTER superseded_by_id,
    ADD COLUMN IF NOT EXISTS original_content TEXT NULL AFTER content;

-- Add indexes on extended ai_memory_items
CREATE INDEX IF NOT EXISTS idx_ai_mem_cat ON ai_memory_items (org_id, category);
CREATE INDEX IF NOT EXISTS idx_ai_mem_entity ON ai_memory_items (org_id, entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_ai_mem_outcome ON ai_memory_items (outcome_id);
CREATE INDEX IF NOT EXISTS idx_ai_mem_stale ON ai_memory_items (org_id, is_stale, status);

-- 2. Structured Agent Outcomes Table
CREATE TABLE IF NOT EXISTS ai_agent_outcomes (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    outcome_id VARCHAR(64) NOT NULL,
    source_entity_type VARCHAR(64) NOT NULL, -- 'SHIPMENT', 'EXCEPTION', 'INVOICE', 'CONTRACT', 'RFQ', 'PLAN', 'DECISION', 'CUSTOMER', 'CARRIER'
    source_entity_id VARCHAR(64) NOT NULL,
    workflow_id VARCHAR(64) NULL,
    plan_id VARCHAR(64) NULL,
    plan_version INT NOT NULL DEFAULT 1,
    step_id VARCHAR(64) NULL,
    action_id VARCHAR(64) NULL,
    action_type VARCHAR(128) NULL,
    outcome_type VARCHAR(64) NOT NULL, -- 'RECOMMENDATION_OUTCOME', 'PLAN_EXECUTION_OUTCOME', 'EXCEPTION_RECOVERY_OUTCOME', 'CUSTOMER_COMMUNICATION_OUTCOME', 'CARRIER_RESPONSE_OUTCOME', 'PRICING_DECISION_OUTCOME', 'COLLECTION_OUTCOME', 'COMPLIANCE_REMEDIATION_OUTCOME', 'HUMAN_DECISION_OUTCOME'
    expected_result TEXT NULL,
    actual_result TEXT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'UNVERIFIED', -- 'SUCCESS', 'PARTIAL_SUCCESS', 'FAILED', 'CANCELLED', 'REJECTED', 'SUPERSEDED', 'ESCALATED', 'UNVERIFIED', 'TIMEOUT'
    is_verified TINYINT(1) NOT NULL DEFAULT 0,
    verified_at DATETIME NULL,
    verification_method VARCHAR(64) NULL, -- 'AUTHORITATIVE_STATE_CHECK', 'MANUAL_HUMAN_VERIFICATION', 'TELEMETRY_CONFIRMED', 'CARRIER_EDI_CONFIRMED', 'PAYMENT_SETTLED'
    time_to_resolution_sec INT NULL,
    failure_category VARCHAR(64) NULL DEFAULT 'NONE', -- 'NONE', 'CARRIER_NON_RESPONSIVE', 'CUSTOMER_REJECTED', 'POLICY_BLOCKED', 'TIMED_OUT', 'EXECUTION_ERROR', 'ASSUMPTION_INVALIDATED'
    reason TEXT NULL,
    human_involvement VARCHAR(32) NOT NULL DEFAULT 'NONE', -- 'NONE', 'APPROVED', 'MODIFIED', 'OVERRIDDEN', 'HALTED', 'ESCALATED'
    decided_by_id BIGINT NULL,
    decided_by_name VARCHAR(255) NULL,
    confidence_score DECIMAL(5,2) NOT NULL DEFAULT 1.00,
    correlation_id VARCHAR(255) NULL,
    metadata JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_outcome_id (outcome_id),
    INDEX idx_outcome_org_entity (org_id, source_entity_type, source_entity_id),
    INDEX idx_outcome_org_status (org_id, status),
    INDEX idx_outcome_org_type (org_id, outcome_type),
    INDEX idx_outcome_org_verified (org_id, is_verified),
    INDEX idx_outcome_plan (plan_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Learned Operational Patterns Table
CREATE TABLE IF NOT EXISTS ai_learned_patterns (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    pattern_id VARCHAR(64) NOT NULL,
    pattern_type VARCHAR(64) NOT NULL, -- 'CARRIER_DISRUPTION_PATTERN', 'EXCEPTION_RECOVERY_STRATEGY', 'CUSTOMER_PREFERENCE_PATTERN', 'PRICING_CONVERSION_PATTERN', 'COLLECTION_FRICTION_PATTERN', 'COMPLIANCE_BOTTLENECK_PATTERN', 'WORKFLOW_FAILURE_PATTERN'
    entity_type VARCHAR(64) NOT NULL,
    entity_identifier VARCHAR(128) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    recommended_strategy TEXT NULL,
    supporting_observations INT NOT NULL DEFAULT 1,
    success_rate DECIMAL(5,2) NOT NULL DEFAULT 1.00,
    confidence VARCHAR(16) NOT NULL DEFAULT 'MEDIUM', -- 'HIGH', 'MEDIUM', 'LOW'
    scope VARCHAR(32) NOT NULL DEFAULT 'TENANT',
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    last_observed_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_pattern_id (pattern_id),
    INDEX idx_pat_org_type (org_id, pattern_type),
    INDEX idx_pat_org_ident (org_id, entity_type, entity_identifier),
    INDEX idx_pat_org_active (org_id, is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
