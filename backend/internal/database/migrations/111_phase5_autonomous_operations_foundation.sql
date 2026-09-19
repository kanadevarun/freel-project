-- Migration 111: Phase 5 Controlled Autonomous Operations & Adaptive Planning Foundation
-- Establishes authoritative persistence for autonomy policies, operational plans,
-- multi-step execution tracking, verification states, durable replan lineages, and operational memory.

CREATE TABLE IF NOT EXISTS autonomy_policies (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL COMMENT '0 for system baseline default, or specific tenant org_id',
    module VARCHAR(50) NOT NULL COMMENT 'shipments, rfq, pricing, finance, contracts, customers, general',
    autonomy_level VARCHAR(40) NOT NULL DEFAULT 'LEVEL_1_RECOMMEND' COMMENT 'LEVEL_0_OBSERVE, LEVEL_1_RECOMMEND, LEVEL_2_PREPARE, LEVEL_3_CONTROLLED_EXECUTION, LEVEL_4_CONTROLLED_MULTI_STEP',
    allowed_action_types JSON NULL COMMENT 'List of Action System action_type strings permitted',
    prohibited_action_types JSON NULL COMMENT 'List of explicitly forbidden action_type strings',
    requires_approval BOOLEAN NOT NULL DEFAULT TRUE,
    max_monetary_threshold DECIMAL(12, 2) NOT NULL DEFAULT 1000.00,
    customer_impact_threshold VARCHAR(20) NOT NULL DEFAULT 'LOW' COMMENT 'LOW, MEDIUM, HIGH, CRITICAL',
    shipment_impact_threshold VARCHAR(20) NOT NULL DEFAULT 'LOW' COMMENT 'LOW, MEDIUM, HIGH, CRITICAL',
    compliance_sensitivity VARCHAR(20) NOT NULL DEFAULT 'STANDARD' COMMENT 'STANDARD, HIGH, REGULATED',
    min_confidence_threshold DECIMAL(5, 4) NOT NULL DEFAULT 0.7500,
    require_data_sufficiency BOOLEAN NOT NULL DEFAULT TRUE,
    max_plan_steps INT NOT NULL DEFAULT 5,
    max_execution_attempts INT NOT NULL DEFAULT 3,
    cooldown_seconds INT NOT NULL DEFAULT 60,
    emergency_stop BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    policy_version INT NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uq_autonomy_policy (org_id, module),
    INDEX idx_autonomy_policy_active (org_id, is_active),
    INDEX idx_autonomy_policy_level (org_id, autonomy_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS autonomous_plans (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL,
    user_id INT NULL,
    plan_id VARCHAR(100) NOT NULL UNIQUE,
    version INT NOT NULL DEFAULT 1,
    parent_plan_id VARCHAR(100) NULL COMMENT 'Previous version plan_id if replanned',
    correlation_id VARCHAR(100) NOT NULL,
    goal TEXT NOT NULL,
    module VARCHAR(50) NOT NULL COMMENT 'shipments, rfq, pricing, finance, contracts, customers, cross_module',
    related_entity_type VARCHAR(50) NOT NULL COMMENT 'SHIPMENT, INVOICE, LEAD, RFQ, CONTRACT, CUSTOMER',
    related_entity_id VARCHAR(100) NOT NULL,
    current_state_summary TEXT NOT NULL,
    constraints JSON NULL,
    assumptions JSON NULL,
    risks JSON NULL,
    confidence_score DECIMAL(5, 4) NOT NULL DEFAULT 0.0000,
    data_sufficiency BOOLEAN NOT NULL DEFAULT TRUE,
    estimated_impact TEXT NULL,
    risk_level VARCHAR(20) NOT NULL DEFAULT 'MEDIUM' COMMENT 'LOW, MEDIUM, HIGH, CRITICAL',
    autonomy_level VARCHAR(40) NOT NULL DEFAULT 'LEVEL_1_RECOMMEND',
    policy_decision VARCHAR(40) NOT NULL DEFAULT 'PERMITTED' COMMENT 'PERMITTED, REQUIRES_APPROVAL, BLOCKED_POLICY, BLOCKED_EMERGENCY_STOP',
    policy_reason TEXT NULL,
    status VARCHAR(40) NOT NULL DEFAULT 'DRAFT' COMMENT 'DRAFT, GENERATED, VALIDATING, REQUIRES_APPROVAL, APPROVED, REJECTED, EXECUTING, PAUSED, WAITING, COMPLETED, PARTIALLY_COMPLETED, FAILED, CANCELLED, EXPIRED, REPLANNING',
    replan_status VARCHAR(40) NOT NULL DEFAULT 'NONE' COMMENT 'NONE, REQUESTED, IN_PROGRESS, COMPLETED',
    replan_reason TEXT NULL,
    triggering_event VARCHAR(100) NULL,
    execution_status VARCHAR(40) NOT NULL DEFAULT 'NOT_STARTED' COMMENT 'NOT_STARTED, IN_PROGRESS, SUCCEEDED, FAILED, PAUSED',
    verification_status VARCHAR(40) NOT NULL DEFAULT 'PENDING' COMMENT 'PENDING, VERIFIED_SUCCESS, VERIFIED_DISCREPANCY, VERIFIED_FAILURE',
    expires_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_plans_org_module (org_id, module),
    INDEX idx_plans_org_status (org_id, status),
    INDEX idx_plans_entity (org_id, related_entity_type, related_entity_id),
    INDEX idx_plans_correlation (org_id, correlation_id),
    INDEX idx_plans_parent (org_id, parent_plan_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS autonomous_plan_steps (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    plan_id VARCHAR(100) NOT NULL,
    org_id INT NOT NULL,
    step_number INT NOT NULL,
    step_id VARCHAR(100) NOT NULL,
    action_type VARCHAR(80) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    parameters JSON NOT NULL,
    dependencies JSON NULL COMMENT 'Array of prior step_ids required before execution',
    expected_outcome TEXT NOT NULL,
    risk_level VARCHAR(20) NOT NULL DEFAULT 'LOW' COMMENT 'LOW, MEDIUM, HIGH, CRITICAL',
    requires_approval BOOLEAN NOT NULL DEFAULT FALSE,
    action_system_action_id VARCHAR(100) NULL,
    idempotency_key VARCHAR(191) NOT NULL,
    status VARCHAR(40) NOT NULL DEFAULT 'PENDING' COMMENT 'PENDING, AWAITING_APPROVAL, APPROVED, REJECTED, EXECUTING, COMPLETED, FAILED, SKIPPED',
    execution_attempt INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    executed_at DATETIME NULL,
    execution_result JSON NULL,
    error_message TEXT NULL,
    verification_status VARCHAR(40) NOT NULL DEFAULT 'PENDING' COMMENT 'PENDING, VERIFIED_SUCCESS, VERIFIED_FAILURE',
    verification_details JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uq_plan_step (plan_id, step_number),
    INDEX idx_plan_steps_org_plan (org_id, plan_id),
    INDEX idx_plan_steps_idempotency (org_id, idempotency_key),
    INDEX idx_plan_steps_status (org_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS autonomous_plan_audit_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    plan_id VARCHAR(100) NOT NULL,
    org_id INT NOT NULL,
    user_id INT NULL,
    event_type VARCHAR(50) NOT NULL,
    previous_status VARCHAR(40) NULL,
    new_status VARCHAR(40) NOT NULL,
    step_id VARCHAR(100) NULL,
    details JSON NULL,
    notes TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_plan_audit (org_id, plan_id),
    INDEX idx_plan_audit_event (org_id, event_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS autonomous_operational_memory (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL,
    memory_type VARCHAR(50) NOT NULL COMMENT 'PLAN_OUTCOME, RECURRING_PATTERN, POLICY_DECISION, STRATEGY_PERFORMANCE',
    entity_type VARCHAR(50) NOT NULL COMMENT 'SHIPMENT, INVOICE, LEAD, RFQ, CONTRACT, CARRIER, CUSTOMER',
    entity_id VARCHAR(100) NOT NULL,
    summary TEXT NOT NULL,
    structured_payload JSON NOT NULL,
    success_rating DECIMAL(3, 2) NOT NULL DEFAULT 1.00 COMMENT '0.00 to 1.00',
    usage_count INT NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_autonomy_memory_org (org_id, memory_type, entity_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
