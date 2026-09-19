-- Migration 103: Phase 3 Task 3.6 - Finance and Collections Automation and Intelligent Receivables Follow-Up
-- Establishes durable persistence for AI-analyzed receivables risk profiles and communication drafts.

CREATE TABLE IF NOT EXISTS ai_finance_receivables_analyses (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    invoice_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    risk_level VARCHAR(50) NOT NULL DEFAULT 'LOW',
    risk_score DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    days_overdue INT NOT NULL DEFAULT 0,
    aging_bucket VARCHAR(50) NOT NULL DEFAULT 'CURRENT',
    outstanding_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    receivables_summary TEXT NOT NULL,
    deterministic_signals JSON NOT NULL,
    key_risks JSON NOT NULL,
    recommended_next_steps JSON NOT NULL,
    evidence JSON NOT NULL,
    confidence_score DECIMAL(4,3) NOT NULL DEFAULT 0.850,
    correlation_id VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ai_rec_org_inv (org_id, invoice_id),
    INDEX idx_ai_rec_org_cust (org_id, customer_id),
    INDEX idx_ai_rec_risk (org_id, risk_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ai_finance_collection_drafts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    invoice_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    draft_type VARCHAR(50) NOT NULL, -- FIRST_REMINDER, OVERDUE_NOTICE, FINAL_DEMAND, PAYMENT_PLAN_OFFER, INTERNAL_ESCALATION
    subject VARCHAR(255) NOT NULL,
    message_body TEXT NOT NULL,
    internal_notes TEXT NULL,
    recipient_name VARCHAR(255) NOT NULL,
    recipient_email VARCHAR(255) NOT NULL,
    outstanding_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT', -- DRAFT, PENDING_APPROVAL, APPROVED, REJECTED, SENT
    requires_approval TINYINT(1) NOT NULL DEFAULT 1,
    approval_id BIGINT NULL,
    action_proposal_id VARCHAR(100) NULL,
    created_by_user_id BIGINT NULL,
    correlation_id VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ai_col_org_inv (org_id, invoice_id),
    INDEX idx_ai_col_org_cust (org_id, customer_id),
    INDEX idx_ai_col_status (org_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
