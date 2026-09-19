-- 116_phase5_adaptive_finance_collections.sql
-- LogisticsHQ Phase 5 Task 5.6: Adaptive Finance and Collections Schema

CREATE TABLE IF NOT EXISTS finance_collection_plans (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    invoice_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    invoice_number VARCHAR(100) NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    balance_due DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    due_date DATE NULL,
    days_overdue INT NOT NULL DEFAULT 0,
    aging_bucket VARCHAR(50) NOT NULL DEFAULT 'CURRENT',
    priority_level VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    priority_score DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    risk_level VARCHAR(50) NOT NULL DEFAULT 'LOW',
    risk_score DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    recommended_strategy_id VARCHAR(100) NOT NULL,
    selected_strategy_id VARCHAR(100) NOT NULL,
    candidate_strategies LONGTEXT NULL,
    actual_facts LONGTEXT NULL,
    predictions LONGTEXT NULL,
    assumptions LONGTEXT NULL,
    draft_message LONGTEXT NULL,
    plan_steps LONGTEXT NULL,
    requires_approval TINYINT(1) NOT NULL DEFAULT 0,
    approval_reason VARCHAR(500) NULL,
    autonomy_level VARCHAR(50) NOT NULL DEFAULT 'LEVEL_2_PREPARE',
    stop_reason VARCHAR(500) NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'GENERATED',
    version INT NOT NULL DEFAULT 1,
    idempotency_key VARCHAR(128) NOT NULL,
    correlation_id VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_fin_plan_org_idemp (org_id, idempotency_key),
    INDEX idx_fin_plan_org_inv (org_id, invoice_id),
    INDEX idx_fin_plan_org_cust (org_id, customer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS finance_collection_versions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    plan_id BIGINT NOT NULL,
    org_id BIGINT NOT NULL,
    version_number INT NOT NULL,
    trigger_event VARCHAR(100) NOT NULL,
    balance_due DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    status VARCHAR(50) NOT NULL,
    strategy_id VARCHAR(100) NOT NULL,
    change_reason VARCHAR(500) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_fin_ver_plan (org_id, plan_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed default autonomy policies for module 'finance_collections' if not existing
INSERT INTO autonomy_policies (
    org_id, module, autonomy_level, allowed_action_types, prohibited_action_types,
    requires_approval, max_monetary_threshold, customer_impact_threshold,
    shipment_impact_threshold, compliance_sensitivity, min_confidence_threshold,
    require_data_sufficiency, max_plan_steps, max_execution_attempts,
    cooldown_seconds, emergency_stop, is_active, policy_version, created_at, updated_at
)
SELECT 1, 'finance_collections', 'LEVEL_2_PREPARE',
       '["send_payment_reminder", "request_status_confirmation", "escalate_to_finance"]',
       '["direct_write_off", "direct_refund", "unauthorized_discount", "direct_accounting_edit"]',
       1, 5000.00, 'MEDIUM', 'LOW', 'MEDIUM', 0.8000, 1, 7, 3, 300, 0, 1, 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM autonomy_policies WHERE org_id = 1 AND module = 'finance_collections');

INSERT INTO autonomy_policies (
    org_id, module, autonomy_level, allowed_action_types, prohibited_action_types,
    requires_approval, max_monetary_threshold, customer_impact_threshold,
    shipment_impact_threshold, compliance_sensitivity, min_confidence_threshold,
    require_data_sufficiency, max_plan_steps, max_execution_attempts,
    cooldown_seconds, emergency_stop, is_active, policy_version, created_at, updated_at
)
SELECT 2, 'finance_collections', 'LEVEL_2_PREPARE',
       '["send_payment_reminder", "request_status_confirmation", "escalate_to_finance"]',
       '["direct_write_off", "direct_refund", "unauthorized_discount", "direct_accounting_edit"]',
       1, 5000.00, 'MEDIUM', 'LOW', 'MEDIUM', 0.8000, 1, 7, 3, 300, 0, 1, 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM autonomy_policies WHERE org_id = 2 AND module = 'finance_collections');
