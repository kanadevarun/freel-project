-- ==============================================================================
-- LogisticsHQ Phase 3 Task 3.4: RFQ-to-Quotation Automation & Intelligent Pricing Workflow
-- Tables for Extracted Requirements Persistence and AI Quotation Drafts
-- ==============================================================================

CREATE TABLE IF NOT EXISTS ai_rfq_requirements_extractions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    rfq_id BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'COMPLETE',
    extracted_data JSON NOT NULL,
    missing_fields JSON NULL,
    clarification_needed BOOLEAN NOT NULL DEFAULT FALSE,
    confidence_score DECIMAL(4,3) NOT NULL DEFAULT 1.000,
    correlation_id VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_rfq_req_org_rfq (org_id, rfq_id),
    INDEX idx_rfq_req_correlation (correlation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ai_quotation_drafts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    rfq_id BIGINT NOT NULL,
    quotation_id BIGINT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT',
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    base_cost DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    total_cost DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    base_sell DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    surcharges DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    discounts DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    tax_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    total_selling_price DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    gross_margin_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    gross_margin_pct DECIMAL(6,2) NOT NULL DEFAULT 0.00,
    margin_health VARCHAR(20) NOT NULL DEFAULT 'HEALTHY',
    rate_references JSON NULL,
    cost_components JSON NULL,
    selling_components JSON NULL,
    terms_and_conditions TEXT NULL,
    internal_summary TEXT NULL,
    customer_wording TEXT NULL,
    pricing_explanation TEXT NULL,
    recipient_email VARCHAR(255) NOT NULL DEFAULT '',
    recipient_name VARCHAR(255) NOT NULL DEFAULT '',
    validity_start DATE NULL,
    validity_end DATE NULL,
    requires_approval BOOLEAN NOT NULL DEFAULT FALSE,
    approval_id BIGINT NULL,
    action_proposal_id VARCHAR(128) NULL,
    execution_id BIGINT NULL,
    created_by_user_id BIGINT NULL,
    correlation_id VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_quote_draft_org_rfq (org_id, rfq_id),
    INDEX idx_quote_draft_status (org_id, status),
    INDEX idx_quote_draft_correlation (correlation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
