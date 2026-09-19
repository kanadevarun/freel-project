-- Migration 110: Phase 4 Predictive Intelligence & Decision Support Foundation
-- Creates centralized tables for cross-module prediction records, lifecycle tracking,
-- source grounding, and outcome evaluation across LogisticsHQ.

CREATE TABLE IF NOT EXISTS predictions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id INT NOT NULL,
    user_id INT NULL,
    prediction_id VARCHAR(100) NOT NULL UNIQUE,
    idempotency_key VARCHAR(191) NOT NULL,
    module VARCHAR(50) NOT NULL COMMENT 'shipments, rfq, pricing, finance, contracts, customers, cross_module',
    prediction_type VARCHAR(80) NOT NULL COMMENT 'SHIPMENT_ETA_DELAY, INVOICE_PAYMENT_DEFAULT, CUSTOMER_CHURN_RISK, etc.',
    status VARCHAR(40) NOT NULL DEFAULT 'PUBLISHED' COMMENT 'GENERATED, VALIDATED, PUBLISHED, ACKNOWLEDGED, IN_REVIEW, ACCEPTED, DISMISSED, ACTION_REQUESTED, AWAITING_APPROVAL, ACTION_EXECUTED, EXPIRED, SUPERSEDED, FAILED',
    severity VARCHAR(20) NOT NULL DEFAULT 'MEDIUM' COMMENT 'LOW, MEDIUM, HIGH, CRITICAL',
    confidence_score DECIMAL(5, 4) NOT NULL DEFAULT 0.0000 COMMENT '0.0000 to 1.0000',
    confidence_band VARCHAR(20) NOT NULL DEFAULT 'MEDIUM' COMMENT 'LOW, MEDIUM, HIGH',
    related_record_type VARCHAR(50) NOT NULL COMMENT 'SHIPMENT, INVOICE, LEAD, RFQ, CONTRACT, CUSTOMER',
    related_record_id VARCHAR(100) NOT NULL,
    prediction_statement TEXT NOT NULL,
    predicted_value VARCHAR(255) NULL,
    time_horizon VARCHAR(100) NULL,
    target_date DATETIME NULL,
    explanation TEXT NOT NULL,
    supporting_signals JSON NULL,
    source_references JSON NOT NULL,
    source_timestamp DATETIME NOT NULL,
    recommended_action TEXT NULL,
    action_type VARCHAR(80) NULL,
    is_action_required BOOLEAN NOT NULL DEFAULT FALSE,
    requires_approval BOOLEAN NOT NULL DEFAULT FALSE,
    action_proposal_id VARCHAR(100) NULL,
    review_status VARCHAR(40) NOT NULL DEFAULT 'UNREVIEWED' COMMENT 'UNREVIEWED, ACKNOWLEDGED, UNDER_REVIEW, DISMISSED, ACTION_TAKEN',
    reviewed_by INT NULL,
    reviewed_at DATETIME NULL,
    review_notes TEXT NULL,
    actual_outcome_status VARCHAR(40) NOT NULL DEFAULT 'PENDING' COMMENT 'PENDING, CORRECT, INCORRECT, INCONCLUSIVE',
    actual_outcome_value VARCHAR(255) NULL,
    feedback_notes TEXT NULL,
    model_version VARCHAR(50) NOT NULL DEFAULT 'gemini-1.5-pro',
    expires_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_predictions_org_module (org_id, module),
    INDEX idx_predictions_org_status (org_id, status),
    INDEX idx_predictions_org_severity (org_id, severity),
    INDEX idx_predictions_related (org_id, related_record_type, related_record_id),
    INDEX idx_predictions_idempotency (org_id, idempotency_key),
    INDEX idx_predictions_type_status (org_id, prediction_type, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS prediction_audit_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    prediction_id VARCHAR(100) NOT NULL,
    org_id INT NOT NULL,
    user_id INT NULL,
    previous_status VARCHAR(40) NULL,
    new_status VARCHAR(40) NOT NULL,
    action VARCHAR(50) NOT NULL,
    notes TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_pred_audit (org_id, prediction_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
