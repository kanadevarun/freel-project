-- Migration 109: AI Governance and Production Controls for LogisticsHQ
-- Authoritative schema for AI policies, kill switches, feature flags, and safety violation audits.

CREATE TABLE IF NOT EXISTS ai_governance_policies (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL COMMENT 'Organization ID, 0 for system global template',
    allowed_workflows JSON NOT NULL COMMENT 'Array of allowed AI workflow keys',
    allowed_models JSON NOT NULL COMMENT 'Array of allowed model identifiers',
    allowed_providers JSON NOT NULL COMMENT 'Array of allowed AI providers (gemini, openai, anthropic)',
    allowed_actions JSON NOT NULL COMMENT 'Array of allowed action names',
    max_input_chars INT NOT NULL DEFAULT 100000 COMMENT 'Hard limit on prompt/input character length',
    max_output_chars INT NOT NULL DEFAULT 50000 COMMENT 'Hard limit on generated output character length',
    timeout_seconds INT NOT NULL DEFAULT 30 COMMENT 'Execution timeout per workflow',
    max_retries INT NOT NULL DEFAULT 3 COMMENT 'Maximum automated retries',
    rate_limit_rpm INT NOT NULL DEFAULT 120 COMMENT 'Requests per minute ceiling per tenant',
    enforce_prompt_injection_check TINYINT(1) NOT NULL DEFAULT 1,
    enforce_sensitive_data_redaction TINYINT(1) NOT NULL DEFAULT 1,
    enforce_hitl_approvals TINYINT(1) NOT NULL DEFAULT 1,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_gov_policy_org (org_id),
    INDEX idx_gov_policy_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ai_governance_kill_switches (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL DEFAULT 0 COMMENT '0 for global system kill switch, or specific org ID',
    scope VARCHAR(50) NOT NULL COMMENT 'GLOBAL, WORKFLOW, MODEL, PROVIDER, ACTION',
    target_identifier VARCHAR(150) NOT NULL COMMENT 'GLOBAL_AI, WORKFLOW:pricing, MODEL:gpt-4o, ACTION:PAYMENT_DISPATCH, etc.',
    is_killed TINYINT(1) NOT NULL DEFAULT 1 COMMENT '1 = blocked/disabled, 0 = active',
    reason VARCHAR(255) NOT NULL COMMENT 'Operational justification or incident ticket',
    killed_by_user_id BIGINT NULL COMMENT 'Admin user who toggled the switch',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_gov_kill_target_org (org_id, scope, target_identifier),
    INDEX idx_gov_kill_status (is_killed),
    INDEX idx_gov_kill_scope (scope)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ai_safety_violations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL,
    user_id BIGINT NULL,
    workflow_name VARCHAR(100) NOT NULL,
    violation_type VARCHAR(100) NOT NULL COMMENT 'PROMPT_INJECTION, SENSITIVE_DATA_EXPOSURE, UNSUPPORTED_CLAIM, KILL_SWITCH_ACTIVE, POLICY_DISALLOWED_WORKFLOW, RATE_LIMIT_EXCEEDED, UNSAFE_ACTION_ATTEMPT',
    severity VARCHAR(20) NOT NULL DEFAULT 'HIGH' COMMENT 'LOW, MEDIUM, HIGH, CRITICAL',
    action_taken VARCHAR(50) NOT NULL DEFAULT 'REJECTED' COMMENT 'REJECTED, SANITIZED, FLAGGED, GATED',
    details TEXT NOT NULL COMMENT 'Inspection explanation and diagnostic details',
    input_snippet_redacted TEXT NULL COMMENT 'Sanitized snippet of problematic input (no PII/secrets)',
    correlation_id VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_asv_org_created (org_id, created_at),
    INDEX idx_asv_type (violation_type),
    INDEX idx_asv_severity (severity),
    INDEX idx_asv_correlation (correlation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default system-wide global governance policy template (org_id = 0) if not exists
INSERT INTO ai_governance_policies (
    org_id,
    allowed_workflows,
    allowed_models,
    allowed_providers,
    allowed_actions,
    max_input_chars,
    max_output_chars,
    timeout_seconds,
    max_retries,
    rate_limit_rpm,
    enforce_prompt_injection_check,
    enforce_sensitive_data_redaction,
    enforce_hitl_approvals,
    is_active
) VALUES (
    0,
    '["pricing", "sales", "operations", "contracts", "compliance", "finance", "reporting", "copilot", "notifications", "outreach"]',
    '["gemini-1.5-pro", "gemini-1.5-flash", "gpt-4o", "gpt-4o-mini", "claude-3-5-sonnet", "deterministic_test"]',
    '["gemini", "openai", "anthropic", "deterministic_test"]',
    '["CREATE_RECOMMENDATION", "CREATE_INTERNAL_TASK", "REQUEST_HUMAN_APPROVAL", "CALCULATE_FORECAST", "GENERATE_DRAFT", "NOTIFY_OPERATOR"]',
    100000,
    50000,
    30,
    3,
    120,
    1,
    1,
    1,
    1
) ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP;
