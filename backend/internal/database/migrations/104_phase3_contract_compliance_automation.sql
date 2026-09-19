-- Migration 104: Phase 3 Task 3.7 - Contract and Compliance Automation and Intelligent Document Review

CREATE TABLE IF NOT EXISTS ai_contract_compliance_reviews (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    contract_id BIGINT NOT NULL,
    document_id VARCHAR(64) NULL,
    risk_level VARCHAR(32) NOT NULL,
    risk_score DECIMAL(5,2) NOT NULL,
    compliance_status VARCHAR(32) NOT NULL,
    executive_summary TEXT NOT NULL,
    deterministic_signals JSON NOT NULL,
    extracted_clauses JSON NOT NULL,
    structured_discrepancies JSON NOT NULL,
    compliance_obligations JSON NOT NULL,
    missing_information JSON NOT NULL,
    recommendations JSON NOT NULL,
    evidence JSON NOT NULL,
    confidence_score DECIMAL(5,4) NOT NULL,
    correlation_id VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL,
    INDEX idx_accr_contract (org_id, contract_id),
    INDEX idx_accr_risk (org_id, risk_level),
    INDEX idx_accr_created (org_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ai_contract_compliance_drafts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    contract_id BIGINT NOT NULL,
    document_id VARCHAR(64) NULL,
    draft_type VARCHAR(64) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    message_body TEXT NOT NULL,
    internal_notes TEXT NULL,
    recipient_name VARCHAR(255) NOT NULL,
    recipient_email VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT',
    requires_approval TINYINT(1) NOT NULL DEFAULT 1,
    approval_id BIGINT NULL,
    action_proposal_id VARCHAR(64) NULL,
    created_by_user_id BIGINT NULL,
    correlation_id VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    INDEX idx_accd_contract (org_id, contract_id),
    INDEX idx_accd_status (org_id, status),
    INDEX idx_accd_approval (org_id, approval_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
