-- 118_phase5_autonomous_exception_resolution.sql
-- LogisticsHQ Phase 5 Task 5.8: Autonomous Exception Resolution Schema

CREATE TABLE IF NOT EXISTS exception_resolution_plans (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    exception_id BIGINT NOT NULL,
    shipment_id BIGINT NOT NULL,
    exception_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    lifecycle_status VARCHAR(50) NOT NULL DEFAULT 'PLAN_READY',
    waiting_state VARCHAR(50) NULL,
    likely_root_cause TEXT NULL,
    symptom TEXT NULL,
    contributing_factors LONGTEXT NULL,
    evidence LONGTEXT NULL,
    confidence DECIMAL(5,2) NOT NULL DEFAULT 0.85,
    impact_assessment LONGTEXT NULL,
    constraints LONGTEXT NULL,
    candidate_strategies LONGTEXT NULL,
    selected_strategy_id VARCHAR(100) NOT NULL,
    recovery_plan_steps LONGTEXT NULL,
    verification_criteria LONGTEXT NULL,
    requires_approval TINYINT(1) NOT NULL DEFAULT 0,
    approval_reason VARCHAR(500) NULL,
    autonomy_level VARCHAR(50) NOT NULL DEFAULT 'LEVEL_2_PREPARE',
    stop_reason VARCHAR(500) NULL,
    escalation_reason VARCHAR(500) NULL,
    resolution_notes TEXT NULL,
    version INT NOT NULL DEFAULT 1,
    idempotency_key VARCHAR(128) NOT NULL,
    correlation_id VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_erp_plan_org_idemp (org_id, idempotency_key),
    INDEX idx_erp_plan_org_exception (org_id, exception_id),
    INDEX idx_erp_plan_org_status (org_id, lifecycle_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS exception_resolution_versions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    plan_id BIGINT NOT NULL,
    org_id BIGINT NOT NULL,
    exception_id BIGINT NOT NULL,
    version_number INT NOT NULL,
    trigger_event VARCHAR(100) NOT NULL,
    lifecycle_status VARCHAR(50) NOT NULL,
    selected_strategy_id VARCHAR(100) NOT NULL,
    change_reason VARCHAR(500) NULL,
    snapshot LONGTEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_erp_ver_plan (org_id, plan_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed default autonomy policies for module 'exceptions' if not existing
INSERT INTO autonomy_policies (
    org_id, module, autonomy_level, allowed_action_types, prohibited_action_types,
    requires_approval, max_monetary_threshold, customer_impact_threshold,
    shipment_impact_threshold, compliance_sensitivity, min_confidence_threshold,
    require_data_sufficiency, max_plan_steps, max_execution_attempts,
    cooldown_seconds, emergency_stop, is_active, policy_version, created_at, updated_at
)
SELECT 1, 'exceptions', 'LEVEL_2_PREPARE',
       '["carrier_inquiry", "customer_advisory", "customs_broker_notification", "audit_log"]',
       '["autonomous_cargo_reroute", "autonomous_liability_waiver", "autonomous_financial_settlement"]',
       1, 0.00, 'HIGH', 'HIGH', 'HIGH', 0.8500, 1, 7, 3, 300, 0, 1, 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM autonomy_policies WHERE org_id = 1 AND module = 'exceptions');

INSERT INTO autonomy_policies (
    org_id, module, autonomy_level, allowed_action_types, prohibited_action_types,
    requires_approval, max_monetary_threshold, customer_impact_threshold,
    shipment_impact_threshold, compliance_sensitivity, min_confidence_threshold,
    require_data_sufficiency, max_plan_steps, max_execution_attempts,
    cooldown_seconds, emergency_stop, is_active, policy_version, created_at, updated_at
)
SELECT 2, 'exceptions', 'LEVEL_2_PREPARE',
       '["carrier_inquiry", "customer_advisory", "customs_broker_notification", "audit_log"]',
       '["autonomous_cargo_reroute", "autonomous_liability_waiver", "autonomous_financial_settlement"]',
       1, 0.00, 'HIGH', 'HIGH', 'HIGH', 0.8500, 1, 7, 3, 300, 0, 1, 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM autonomy_policies WHERE org_id = 2 AND module = 'exceptions');

-- Ensure Org 1 has a test shipment and exception for cross-tenant isolation testing
INSERT INTO shipments (
    id, org_id, booking_number, carrier_scac, origin_port, destination_port, status, etd, eta, customer_commitment_date, current_risk_level, adaptive_status, created_at, updated_at
)
SELECT 1, 1, 'BK-2026-ORG1-001', 'MAEU', 'INNSA', 'USLAX', 'IN_TRANSIT', '2026-03-01', '2026-03-25', '2026-03-26', 'MEDIUM', 'MONITORING', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM shipments WHERE org_id = 1 AND id = 1);

INSERT INTO shipment_exceptions (
    id, org_id, shipment_id, exception_type, severity, title, description, resolved, status, created_at, updated_at
)
SELECT 1, 1, 1, 'ETA_DELAY', 'MEDIUM', 'Port Delay at Origin', 'Vessel delayed by 2 days due to tidal restrictions.', 0, 'OPEN', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM shipment_exceptions WHERE org_id = 1 AND id = 1);
