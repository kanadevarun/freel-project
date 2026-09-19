-- Migration 102: Phase 3 Task 3.5 Shipment Operations Automation and Intelligent Exception Response
-- Persistent tables for durable operations analysis, risk assessment, and communication drafts.

CREATE TABLE IF NOT EXISTS ai_shipment_operations_analyses (
    id                     BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id                 BIGINT NOT NULL,
    shipment_id            BIGINT NOT NULL,
    risk_level             VARCHAR(20) NOT NULL DEFAULT 'LOW', -- LOW, MEDIUM, HIGH, CRITICAL
    risk_score             DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    operational_summary    TEXT NOT NULL,
    deterministic_signals  LONGTEXT NULL,
    key_risks              LONGTEXT NULL,
    recommended_next_steps LONGTEXT NULL,
    evidence               LONGTEXT NULL,
    confidence_score       DECIMAL(5,3) NOT NULL DEFAULT 1.000,
    correlation_id         VARCHAR(128) NOT NULL,
    created_at             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ai_ship_ops_org_ship (org_id, shipment_id),
    INDEX idx_ai_ship_ops_corr (correlation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ai_shipment_communication_drafts (
    id                  BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id              BIGINT NOT NULL,
    shipment_id         BIGINT NOT NULL,
    draft_type          VARCHAR(50) NOT NULL, -- CARRIER_FOLLOWUP, CUSTOMER_UPDATE, INTERNAL_ESCALATION
    subject             VARCHAR(255) NOT NULL,
    customer_wording    TEXT NOT NULL,
    internal_notes      TEXT NULL,
    recipient_name      VARCHAR(255) NOT NULL DEFAULT '',
    recipient_email     VARCHAR(255) NOT NULL DEFAULT '',
    status              VARCHAR(32) NOT NULL DEFAULT 'DRAFT', -- DRAFT, PENDING_APPROVAL, APPROVED, REJECTED, EXECUTED
    requires_approval   TINYINT(1) NOT NULL DEFAULT 1,
    approval_id         BIGINT NULL,
    action_proposal_id  VARCHAR(128) NULL,
    execution_id        BIGINT NULL,
    created_by_user_id  BIGINT NULL,
    correlation_id      VARCHAR(128) NOT NULL,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ai_ship_draft_org_ship (org_id, shipment_id),
    INDEX idx_ai_ship_draft_status (status),
    INDEX idx_ai_ship_draft_corr (correlation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
