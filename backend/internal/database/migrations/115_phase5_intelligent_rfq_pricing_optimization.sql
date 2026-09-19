-- Migration 115: Phase 5 Task 5.5 - Intelligent RFQ and Pricing Optimization
-- Purpose: Schema support for AI-driven RFQ understanding, candidate pricing strategies, margin risk intelligence, quotation versioning, and Action System boundary.

-- 1. RFQ Pricing Optimizations Table
CREATE TABLE IF NOT EXISTS rfq_pricing_optimizations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    rfq_id BIGINT NOT NULL,
    quotation_id BIGINT NULL,
    plan_id VARCHAR(64) NULL,
    current_version INT NOT NULL DEFAULT 1,
    status VARCHAR(32) NOT NULL DEFAULT 'ANALYZED',
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    base_cost DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    predicted_cost DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    actual_facts JSON NULL,
    predictions JSON NULL,
    assumptions JSON NULL,
    candidate_strategies JSON NOT NULL,
    recommended_strategy_id VARCHAR(64) NOT NULL,
    recommended_price DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    recommended_margin_pct DECIMAL(6,2) NOT NULL DEFAULT 0.00,
    target_margin_pct DECIMAL(6,2) NOT NULL DEFAULT 15.00,
    min_margin_pct DECIMAL(6,2) NOT NULL DEFAULT 8.00,
    margin_risk_level VARCHAR(32) NOT NULL DEFAULT 'LOW',
    operational_risk_level VARCHAR(32) NOT NULL DEFAULT 'LOW',
    confidence_score DECIMAL(5,2) NOT NULL DEFAULT 0.85,
    data_sufficiency VARCHAR(32) NOT NULL DEFAULT 'COMPLETE',
    rate_freshness_status VARCHAR(32) NOT NULL DEFAULT 'FRESH',
    rate_source VARCHAR(64) NOT NULL DEFAULT 'RATE_SHEET',
    requires_approval TINYINT(1) NOT NULL DEFAULT 0,
    approval_reason VARCHAR(255) NULL,
    approval_status VARCHAR(32) NOT NULL DEFAULT 'NOT_REQUIRED',
    approval_id VARCHAR(64) NULL,
    idempotency_key VARCHAR(128) NOT NULL,
    reasoning_summary TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_org_rfq_opt_idemp (org_id, idempotency_key),
    INDEX idx_rfq_opt_org_rfq (org_id, rfq_id),
    INDEX idx_rfq_opt_quote (org_id, quotation_id),
    INDEX idx_rfq_opt_status (org_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. RFQ Pricing Versions Table (Audit & Re-evaluation History)
CREATE TABLE IF NOT EXISTS rfq_pricing_versions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    optimization_id BIGINT NOT NULL,
    rfq_id BIGINT NOT NULL,
    quotation_id BIGINT NULL,
    version INT NOT NULL,
    strategy_name VARCHAR(128) NOT NULL,
    price DECIMAL(12,2) NOT NULL,
    cost DECIMAL(12,2) NOT NULL,
    margin_pct DECIMAL(6,2) NOT NULL,
    change_reason VARCHAR(255) NOT NULL,
    approval_status VARCHAR(32) NOT NULL DEFAULT 'APPROVED',
    created_by VARCHAR(64) NOT NULL DEFAULT 'AI_PRICING_OPTIMIZER',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_rfq_pv_org_opt (org_id, optimization_id, version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Seed default autonomy policies for pricing module if not present
INSERT INTO autonomy_policies (org_id, module, autonomy_level, allowed_action_types, prohibited_action_types, requires_approval, max_monetary_threshold, customer_impact_threshold, shipment_impact_threshold, compliance_sensitivity, min_confidence_threshold, require_data_sufficiency, max_plan_steps, max_execution_attempts, cooldown_seconds, emergency_stop, is_active, policy_version)
SELECT 1, 'pricing', 'LEVEL_2_PREPARE', '["ANALYZE_RFQ","GENERATE_PRICING_STRATEGY","PREPARE_QUOTATION"]', '["DIRECT_SEND_UNAPPROVED","EXECUTE_BELOW_MIN_MARGIN"]', 1, 25000.00, 0.70, 0.60, 0.80, 0.75, 1, 10, 3, 30, 0, 1, '1.0'
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM autonomy_policies WHERE org_id = 1 AND module = 'pricing');

INSERT INTO autonomy_policies (org_id, module, autonomy_level, allowed_action_types, prohibited_action_types, requires_approval, max_monetary_threshold, customer_impact_threshold, shipment_impact_threshold, compliance_sensitivity, min_confidence_threshold, require_data_sufficiency, max_plan_steps, max_execution_attempts, cooldown_seconds, emergency_stop, is_active, policy_version)
SELECT 2, 'pricing', 'LEVEL_2_PREPARE', '["ANALYZE_RFQ","GENERATE_PRICING_STRATEGY","PREPARE_QUOTATION"]', '["DIRECT_SEND_UNAPPROVED","EXECUTE_BELOW_MIN_MARGIN"]', 1, 25000.00, 0.70, 0.60, 0.80, 0.75, 1, 10, 3, 30, 0, 1, '1.0'
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM autonomy_policies WHERE org_id = 2 AND module = 'pricing');
