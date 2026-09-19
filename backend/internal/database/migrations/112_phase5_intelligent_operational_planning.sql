-- Migration 112: Phase 5 Task 5.2 Intelligent Operational Planning
-- Expands autonomous operations with structured planning goals, candidate plan alternatives,
-- multi-attribute evaluations, hard/soft constraints, step conditions, and staleness tracking.

CREATE TABLE IF NOT EXISTS planning_goals (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL,
    user_id INT NULL,
    goal_id VARCHAR(100) NOT NULL UNIQUE,
    correlation_id VARCHAR(100) NOT NULL,
    source VARCHAR(50) NOT NULL DEFAULT 'USER' COMMENT 'USER, PROACTIVE_ALERT, EXCEPTION_EVENT, DELAY_PREDICTION',
    module VARCHAR(50) NOT NULL COMMENT 'shipments, rfq, pricing, finance, contracts, customers',
    related_entity_type VARCHAR(50) NOT NULL COMMENT 'SHIPMENT, INVOICE, LEAD, RFQ, CONTRACT, CUSTOMER',
    related_entity_id VARCHAR(100) NOT NULL,
    objective TEXT NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'MEDIUM' COMMENT 'LOW, MEDIUM, HIGH, CRITICAL',
    deadline DATETIME NULL,
    hard_constraints JSON NULL,
    soft_constraints JSON NULL,
    success_criteria TEXT NULL,
    risk_tolerance VARCHAR(20) NOT NULL DEFAULT 'BALANCED' COMMENT 'CONSERVATIVE, BALANCED, AGGRESSIVE',
    autonomy_level VARCHAR(40) NOT NULL DEFAULT 'LEVEL_2_PREPARE',
    required_permissions JSON NULL,
    status VARCHAR(40) NOT NULL DEFAULT 'ACTIVE' COMMENT 'ACTIVE, FULFILLED, SUPERSEDED, CANCELLED',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_goals_org_module (org_id, module),
    INDEX idx_goals_entity (org_id, related_entity_type, related_entity_id),
    INDEX idx_goals_correlation (org_id, correlation_id),
    INDEX idx_goals_status (org_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Enhance autonomous_plans to store candidate plan alternatives and evaluations
ALTER TABLE autonomous_plans
    ADD COLUMN IF NOT EXISTS goal_id VARCHAR(100) NULL AFTER correlation_id,
    ADD COLUMN IF NOT EXISTS candidate_plans JSON NULL AFTER risks,
    ADD COLUMN IF NOT EXISTS selected_candidate_id VARCHAR(100) NULL AFTER candidate_plans,
    ADD COLUMN IF NOT EXISTS evaluation_summary JSON NULL AFTER selected_candidate_id,
    ADD COLUMN IF NOT EXISTS hard_constraints JSON NULL AFTER constraints,
    ADD COLUMN IF NOT EXISTS soft_constraints JSON NULL AFTER hard_constraints,
    ADD COLUMN IF NOT EXISTS staleness_status VARCHAR(40) NOT NULL DEFAULT 'FRESH' AFTER verification_status;

-- Enhance autonomous_plan_steps for conditional logic, verification criteria, and fallbacks
ALTER TABLE autonomous_plan_steps
    ADD COLUMN IF NOT EXISTS condition_predicate JSON NULL AFTER dependencies,
    ADD COLUMN IF NOT EXISTS verification_criteria JSON NULL AFTER expected_outcome,
    ADD COLUMN IF NOT EXISTS reversibility VARCHAR(30) NOT NULL DEFAULT 'REVERSIBLE' AFTER risk_level,
    ADD COLUMN IF NOT EXISTS fallback_action JSON NULL AFTER reversibility,
    ADD COLUMN IF NOT EXISTS timeout_seconds INT NOT NULL DEFAULT 300 AFTER fallback_action;
