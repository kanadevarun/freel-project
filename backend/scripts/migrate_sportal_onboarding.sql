-- SPortal Task S4: Freight Forwarder Customer Onboarding & KYC Schema

CREATE TABLE IF NOT EXISTS organization_onboardings (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'IN_PROGRESS',
    current_step VARCHAR(50) NOT NULL DEFAULT 'COMPANY_INFO',
    progress_percentage INT NOT NULL DEFAULT 20,
    assigned_owner VARCHAR(255) NULL,
    
    -- Billing profile
    billing_company_name VARCHAR(255) NULL,
    billing_contact_name VARCHAR(255) NULL,
    billing_email VARCHAR(255) NULL,
    billing_phone VARCHAR(50) NULL,
    billing_address TEXT NULL,
    billing_city VARCHAR(100) NULL,
    billing_country VARCHAR(100) NULL,
    billing_currency VARCHAR(10) DEFAULT 'INR',
    payment_terms VARCHAR(50) DEFAULT 'NET_30',
    
    -- Financial KYC (Sensitive: masked in UI, audited access)
    bank_name VARCHAR(255) NULL,
    account_holder_name VARCHAR(255) NULL,
    account_number_encrypted VARCHAR(255) NULL,
    account_type VARCHAR(50) DEFAULT 'CURRENT',
    routing_code VARCHAR(50) NULL,
    bank_branch VARCHAR(255) NULL,
    financial_verified TINYINT(1) DEFAULT 0,
    
    -- Agreement & Compliance
    agreement_status VARCHAR(50) DEFAULT 'PENDING',
    agreement_reference VARCHAR(100) NULL,
    agreement_signed_at DATETIME NULL,
    agreement_signed_by VARCHAR(255) NULL,
    
    -- Initial Customer Super Admin Handoff
    admin_name VARCHAR(255) NULL,
    admin_email VARCHAR(255) NULL,
    admin_phone VARCHAR(50) NULL,
    invitation_status VARCHAR(50) DEFAULT 'NOT_SENT',
    invitation_sent_at DATETIME NULL,
    invitation_token VARCHAR(255) NULL,
    
    -- Review and Notes
    verification_notes TEXT NULL,
    completed_by VARCHAR(255) NULL,
    completed_at DATETIME NULL,
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_onboarding_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS organization_onboarding_documents (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    onboarding_id BIGINT NOT NULL,
    org_id BIGINT NOT NULL,
    document_type VARCHAR(100) NOT NULL,
    document_name VARCHAR(255) NOT NULL,
    file_url VARCHAR(512) NOT NULL,
    file_size_bytes BIGINT DEFAULT 0,
    verification_status VARCHAR(50) NOT NULL DEFAULT 'UPLOADED',
    rejection_reason TEXT NULL,
    verified_by VARCHAR(255) NULL,
    verified_at DATETIME NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_onb_doc_org (org_id),
    INDEX idx_onb_doc_onboarding (onboarding_id)
);
