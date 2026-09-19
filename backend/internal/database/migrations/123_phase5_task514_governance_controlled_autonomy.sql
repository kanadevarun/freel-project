-- 123_phase5_task514_governance_controlled_autonomy.sql
-- LogisticsHQ Phase 5, Task 5.14: Governance for Controlled Autonomy
-- Production-grade governance layer, action allowlist, feature flags, tenant limits, and policy audit logs.

-- 1. Action Allowlist Catalog
CREATE TABLE IF NOT EXISTS ai_governance_action_allowlist (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    action_type VARCHAR(128) NOT NULL,
    action_name VARCHAR(255) NOT NULL,
    module VARCHAR(64) NOT NULL,
    risk_class VARCHAR(32) NOT NULL DEFAULT 'MEDIUM', -- LOW, MEDIUM, HIGH
    allowed_autonomy_levels JSON NOT NULL,            -- e.g. [2, 3, 4]
    approval_requirement VARCHAR(32) NOT NULL DEFAULT 'CONDITIONAL_RISK_THRESHOLD', -- ALWAYS, CONDITIONAL_RISK_THRESHOLD, NEVER
    required_permission VARCHAR(128) NOT NULL DEFAULT 'operations:execute',
    reversibility VARCHAR(32) NOT NULL DEFAULT 'REVERSIBLE', -- REVERSIBLE, PARTIALLY_REVERSIBLE, IRREVERSIBLE
    max_financial_limit DECIMAL(12, 2) NOT NULL DEFAULT 0.00,
    is_enabled TINYINT(1) NOT NULL DEFAULT 1,
    description TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_org_action (org_id, action_type),
    INDEX idx_org_module (org_id, module),
    INDEX idx_risk_class (risk_class)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Capability Feature Flags
CREATE TABLE IF NOT EXISTS ai_governance_feature_flags (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    flag_key VARCHAR(64) NOT NULL,
    flag_name VARCHAR(128) NOT NULL,
    is_enabled TINYINT(1) NOT NULL DEFAULT 1,
    max_autonomy_level INT NOT NULL DEFAULT 3,
    requires_approval TINYINT(1) NOT NULL DEFAULT 1,
    description TEXT NULL,
    updated_by_id BIGINT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_org_flag (org_id, flag_key),
    INDEX idx_org_enabled (org_id, is_enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Tenant Governance Limits & Emergency Stop
CREATE TABLE IF NOT EXISTS ai_governance_tenant_limits (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    max_tenant_autonomy INT NOT NULL DEFAULT 3,
    kill_switch_active TINYINT(1) NOT NULL DEFAULT 0,
    kill_switch_reason TEXT NULL,
    kill_switch_by_id BIGINT NULL,
    kill_switch_at DATETIME NULL,
    max_actions_per_hour INT NOT NULL DEFAULT 100,
    max_financial_exposure_per_workflow DECIMAL(12, 2) NOT NULL DEFAULT 5000.00,
    max_retries_per_step INT NOT NULL DEFAULT 3,
    max_replans_per_plan INT NOT NULL DEFAULT 5,
    enforce_four_eyes TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_tenant_org (org_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. Policy Decision Evaluations Audit Log (Every governance check)
CREATE TABLE IF NOT EXISTS ai_governance_policy_evaluations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    user_id BIGINT NULL,
    entity_type VARCHAR(64) NOT NULL,
    entity_id VARCHAR(64) NOT NULL,
    action_type VARCHAR(128) NOT NULL,
    module VARCHAR(64) NOT NULL,
    requested_autonomy INT NOT NULL,
    effective_autonomy INT NOT NULL,
    decision VARCHAR(32) NOT NULL, -- ALLOW, BLOCK, REQUIRE_REVIEW, GOVERNANCE_UNAVAILABLE
    reasons JSON NOT NULL,
    risk_level VARCHAR(32) NOT NULL, -- LOW, MEDIUM, HIGH, CRITICAL
    data_sufficiency VARCHAR(32) NOT NULL DEFAULT 'SUFFICIENT', -- SUFFICIENT, PARTIALLY_SUFFICIENT, INSUFFICIENT
    confidence VARCHAR(32) NOT NULL DEFAULT 'HIGH', -- HIGH, MEDIUM, LOW
    approval_required TINYINT(1) NOT NULL DEFAULT 0,
    four_eyes_required TINYINT(1) NOT NULL DEFAULT 0,
    policy_version INT NOT NULL DEFAULT 1,
    correlation_id VARCHAR(64) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_org_created (org_id, created_at),
    INDEX idx_decision (decision),
    INDEX idx_action_type (action_type),
    INDEX idx_correlation (correlation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. Policy Configuration Audit Trail (Changes made by administrators)
CREATE TABLE IF NOT EXISTS ai_governance_policy_audit_log (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    change_type VARCHAR(64) NOT NULL, -- POLICY_UPDATE, KILL_SWITCH_TOGGLE, FLAG_UPDATE, ALLOWLIST_UPDATE
    target_type VARCHAR(64) NOT NULL,
    target_id VARCHAR(128) NOT NULL,
    old_value JSON NULL,
    new_value JSON NOT NULL,
    reason TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_org_change (org_id, change_type),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
