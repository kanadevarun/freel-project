-- Migration 113: Phase 5 Task 5.3 - Adaptive Shipment Management Foundation

-- 1. Shipment Adaptive Events and Deduplication Journal
CREATE TABLE IF NOT EXISTS shipment_adaptive_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    shipment_id BIGINT NOT NULL,
    event_id VARCHAR(100) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    correlation_id VARCHAR(100) NOT NULL,
    deduplication_key VARCHAR(191) NOT NULL UNIQUE,
    severity VARCHAR(20) NOT NULL DEFAULT 'MEDIUM',
    payload JSON NOT NULL,
    decision VARCHAR(40) NOT NULL DEFAULT 'CONTINUE_MONITORING',
    decision_reason TEXT NULL,
    plan_id VARCHAR(100) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_sae_org_ship (org_id, shipment_id),
    INDEX idx_sae_event_type (event_type),
    INDEX idx_sae_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Extend shipments table with customer delivery commitment and adaptive tracking
ALTER TABLE shipments 
    ADD COLUMN IF NOT EXISTS customer_commitment_date DATETIME NULL AFTER eta,
    ADD COLUMN IF NOT EXISTS current_risk_level VARCHAR(20) NOT NULL DEFAULT 'LOW' AFTER customer_commitment_date,
    ADD COLUMN IF NOT EXISTS adaptive_status VARCHAR(50) NOT NULL DEFAULT 'MONITORED' AFTER current_risk_level;

-- 3. Extend autonomous_plans table with waiting state, commitment metrics, and escalation reason
ALTER TABLE autonomous_plans
    ADD COLUMN IF NOT EXISTS waiting_state VARCHAR(50) NULL DEFAULT NULL AFTER execution_status,
    ADD COLUMN IF NOT EXISTS waiting_until DATETIME NULL DEFAULT NULL AFTER waiting_state,
    ADD COLUMN IF NOT EXISTS escalation_reason TEXT NULL DEFAULT NULL AFTER waiting_until,
    ADD COLUMN IF NOT EXISTS customer_commitment_date DATETIME NULL DEFAULT NULL AFTER escalation_reason,
    ADD COLUMN IF NOT EXISTS predicted_eta DATETIME NULL DEFAULT NULL AFTER customer_commitment_date,
    ADD COLUMN IF NOT EXISTS eta_deviation_hours DECIMAL(8,2) NULL DEFAULT 0.00 AFTER predicted_eta,
    ADD COLUMN IF NOT EXISTS commitment_risk_severity VARCHAR(20) NOT NULL DEFAULT 'NONE' AFTER eta_deviation_hours;
