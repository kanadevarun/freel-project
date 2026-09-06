-- ==============================================================================
-- LOGISTICSHQ — MASTER MYSQL 8 SCHEMA DEFINITION (40 TABLES)
-- ==============================================================================

SET FOREIGN_KEY_CHECKS = 0;

-- 1. Organizations & RBAC
CREATE TABLE IF NOT EXISTS organizations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    cognito_sub VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS roles (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_org_role (org_id, name),
    CONSTRAINT fk_roles_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS org_members (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_org_user (org_id, user_id),
    CONSTRAINT fk_om_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_om_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_om_role FOREIGN KEY (role_id) REFERENCES roles(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS permissions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_resource_action (resource, action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_rp_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_rp_perm FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. CRM & Companies
CREATE TABLE IF NOT EXISTS addresses (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    address_line_1 VARCHAR(255) NOT NULL,
    address_line_2 VARCHAR(255),
    city VARCHAR(255) NOT NULL,
    state VARCHAR(255) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country_code VARCHAR(10) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_addr_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS companies (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255),
    industry VARCHAR(255),
    address_id BIGINT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_comp_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_comp_addr FOREIGN KEY (address_id) REFERENCES addresses(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS contacts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    company_id BIGINT,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    job_title VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_cont_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_cont_comp FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS customers (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    company_id BIGINT NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_cust_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_cust_comp FOREIGN KEY (company_id) REFERENCES companies(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS leads (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    contact_name VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(50),
    source VARCHAR(50) DEFAULT 'DIRECT',
    status VARCHAR(50) DEFAULT 'NEW',
    ai_score INT DEFAULT 0,
    ai_research_report LONGTEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_leads_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS lead_interactions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    lead_id BIGINT NOT NULL,
    org_id BIGINT NOT NULL,
    interaction_type VARCHAR(50) NOT NULL,
    summary TEXT,
    full_body LONGTEXT,
    direction VARCHAR(20) DEFAULT 'INBOUND',
    ai_response_draft LONGTEXT,
    partial_rfq_context JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_li_lead FOREIGN KEY (lead_id) REFERENCES leads(id) ON DELETE CASCADE,
    CONSTRAINT fk_li_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Outreach & Campaigns
CREATE TABLE IF NOT EXISTS outreach_campaigns (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'DRAFT',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_oc_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS outreach_sequences (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    campaign_id BIGINT NOT NULL,
    step_number INT NOT NULL,
    channel VARCHAR(50) NOT NULL,
    template TEXT NOT NULL,
    delay_days INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_os_camp FOREIGN KEY (campaign_id) REFERENCES outreach_campaigns(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. RFQ & Pricing
CREATE TABLE IF NOT EXISTS rfqs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    rfq_number VARCHAR(100) UNIQUE,
    customer_id BIGINT NOT NULL,
    stage VARCHAR(50) DEFAULT 'DRAFT',
    status VARCHAR(50) DEFAULT 'DRAFT',
    agent_status VARCHAR(50) DEFAULT 'IDLE',
    origin VARCHAR(255),
    destination VARCHAR(255),
    incoterms VARCHAR(50),
    target_date DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_rfq_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_rfq_cust FOREIGN KEY (customer_id) REFERENCES customers(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS rfq_items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    rfq_id BIGINT NOT NULL,
    description TEXT,
    quantity INT NOT NULL,
    weight_kg DECIMAL(10,2),
    volume_cbm DECIMAL(10,2),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_ri_rfq FOREIGN KEY (rfq_id) REFERENCES rfqs(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS rfq_quotes (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    rfq_id BIGINT NOT NULL,
    carrier_id VARCHAR(50),
    carrier_name VARCHAR(255),
    quote_reference VARCHAR(100),
    currency VARCHAR(10) DEFAULT 'USD',
    buy_price DECIMAL(15,2) DEFAULT 0,
    sell_price DECIMAL(15,2) DEFAULT 0,
    ocean_freight DECIMAL(15,2) DEFAULT 0,
    origin_charges DECIMAL(15,2) DEFAULT 0,
    destination_charges DECIMAL(15,2) DEFAULT 0,
    total_buy_price DECIMAL(15,2) DEFAULT 0,
    free_days INT DEFAULT 0,
    transit_time_days INT DEFAULT 0,
    valid_from DATETIME NULL,
    valid_until DATETIME NULL,
    etd DATETIME NULL,
    eta DATETIME NULL,
    notes TEXT NULL,
    approved_by VARCHAR(255) NULL,
    approved_at DATETIME NULL,
    charges JSON NULL,
    is_recommended BOOLEAN DEFAULT FALSE,
    reliability_score DECIMAL(5,2) DEFAULT 0,
    historical_success_rate DECIMAL(5,2) DEFAULT 0,
    ai_reasoning TEXT,
    status VARCHAR(50) DEFAULT 'DRAFT',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_rq_rfq FOREIGN KEY (rfq_id) REFERENCES rfqs(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE IF NOT EXISTS opportunities (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    customer_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    amount DECIMAL(15,2),
    stage VARCHAR(50) DEFAULT 'PROSPECTING',
    expected_close_date DATE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_opp_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_opp_cust FOREIGN KEY (customer_id) REFERENCES customers(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. Carriers & Rates
CREATE TABLE IF NOT EXISTS carriers (
    scac VARCHAR(10) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    api_endpoint VARCHAR(255),
    supported_modes JSON,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS carrier_providers (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    scac VARCHAR(20) NOT NULL,
    modes JSON NOT NULL,
    adapter_key VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    supported_capabilities JSON NOT NULL,
    description TEXT NULL,
    documentation_url VARCHAR(500) NULL,
    logo_url VARCHAR(500) NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_cp_code (code),
    INDEX idx_cp_scac (scac)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS carrier_integrations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    carrier_provider_id BIGINT NULL,
    carrier_scac VARCHAR(10) NOT NULL,
    carrier_name VARCHAR(255) NULL,
    connection_method VARCHAR(50) NOT NULL DEFAULT 'API',
    environment VARCHAR(50) NOT NULL DEFAULT 'PRODUCTION',
    connection_status VARCHAR(50) NOT NULL DEFAULT 'DISCONNECTED',
    is_active BOOLEAN DEFAULT TRUE,
    credentials_json JSON NULL,
    encrypted_credentials TEXT NULL,
    credential_mask JSON NULL,
    capabilities JSON NULL,
    config_options JSON NULL,
    sync_status VARCHAR(255) NULL,
    last_synced_at DATETIME NULL,
    last_success_at DATETIME NULL,
    last_failure_at DATETIME NULL,
    last_error TEXT NULL,
    failed_attempts INT NOT NULL DEFAULT 0,
    error_details TEXT NULL,
    next_retry_time DATETIME NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_org_carrier_env (org_id, carrier_scac, environment),
    CONSTRAINT fk_ci_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_ci_provider FOREIGN KEY (carrier_provider_id) REFERENCES carrier_providers(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS rate_entries (
    id VARCHAR(64) PRIMARY KEY,
    org_id BIGINT NOT NULL,
    source VARCHAR(50) DEFAULT 'CONTRACT',
    source_ref VARCHAR(255),
    contract_doc_id VARCHAR(64),
    origin_port VARCHAR(10) NOT NULL,
    destination_port VARCHAR(10) NOT NULL,
    via_port VARCHAR(10),
    service_code VARCHAR(50),
    carrier_scac VARCHAR(10) NOT NULL,
    carrier_name VARCHAR(255) NOT NULL,
    vessel_name VARCHAR(255),
    equipment_type VARCHAR(20) NOT NULL,
    ocean_freight DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    origin_charges DECIMAL(12,2) DEFAULT 0.00,
    destination_charges DECIMAL(12,2) DEFAULT 0.00,
    surcharges JSON,
    total_buy_price DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    currency_original VARCHAR(10) DEFAULT 'USD',
    exchange_rate_used DECIMAL(10,4) DEFAULT 1.0000,
    included_charges JSON,
    excluded_charges JSON,
    free_days_origin INT DEFAULT 7,
    free_days_destination INT DEFAULT 14,
    transit_days INT,
    incoterms VARCHAR(10),
    commodity_restrictions JSON,
    routing_conditions TEXT,
    valid_from DATE NOT NULL,
    valid_until DATE NOT NULL,
    confidence_score INT DEFAULT 100,
    extraction_status VARCHAR(50) DEFAULT 'CONFIRMED',
    extracted_by VARCHAR(255) DEFAULT 'SYSTEM',
    reviewed_by BIGINT,
    nautical_miles INT DEFAULT 0,
    co2_per_teu DECIMAL(10,2) DEFAULT 0.00,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_re_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_re_carrier FOREIGN KEY (carrier_scac) REFERENCES carriers(scac)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS rate_review_queue (
    id VARCHAR(64) PRIMARY KEY,
    org_id BIGINT NOT NULL,
    contract_doc_id VARCHAR(64) NOT NULL,
    extracted_data JSON,
    confidence_score INT DEFAULT 0,
    review_flags JSON,
    ai_reasoning TEXT,
    source_page INT DEFAULT 1,
    source_text TEXT,
    source_image_url VARCHAR(512),
    status VARCHAR(50) DEFAULT 'PENDING',
    reviewed_by BIGINT,
    reviewed_at DATETIME NULL,
    corrected_data JSON,
    review_notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_rrq_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_rrq_doc FOREIGN KEY (contract_doc_id) REFERENCES contract_documents(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS pricing_rules (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    rule_name VARCHAR(255) NOT NULL,
    customer_id BIGINT,
    origin_port VARCHAR(10),
    destination_port VARCHAR(10),
    equipment_type VARCHAR(20),
    markup_type VARCHAR(20) NOT NULL,
    markup_value DECIMAL(12,2) NOT NULL,
    priority INT DEFAULT 10,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_pr_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 6. Contracts & AI Tasks
CREATE TABLE IF NOT EXISTS contract_documents (
    id VARCHAR(64) PRIMARY KEY,
    org_id BIGINT NOT NULL,
    carrier_scac VARCHAR(10) NULL,
    carrier_name VARCHAR(255) NULL,
    file_name VARCHAR(255) NOT NULL,
    s3_key VARCHAR(512) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    file_size_bytes BIGINT DEFAULT 0,
    page_count INT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'PENDING_EXTRACTION',
    ai_document_summary TEXT,
    extracted_rate_count INT DEFAULT 0,
    confirmed_rate_count INT DEFAULT 0,
    pending_review_count INT DEFAULT 0,
    failed_rate_count INT DEFAULT 0,
    processing_log JSON,
    processing_started_at DATETIME NULL,
    processing_completed_at DATETIME NULL,
    reviewed_by BIGINT NULL,
    reviewed_at DATETIME NULL,
    review_notes TEXT,
    created_by BIGINT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_cd_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS ai_processing_tasks (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    document_id VARCHAR(64) NULL,
    entity_type VARCHAR(50) NULL,
    entity_id VARCHAR(64) NULL,
    task_type VARCHAR(50) NOT NULL,
    resource_id BIGINT NULL,
    payload JSON,
    status VARCHAR(50) DEFAULT 'QUEUED',
    error_message TEXT,
    retry_count INT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_apt_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 7. Operations, Bookings & Shipments
CREATE TABLE IF NOT EXISTS bookings (
    id                   BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id               BIGINT NOT NULL,
    rfq_id               BIGINT NOT NULL,
    quote_id             BIGINT NULL,
    booking_number       VARCHAR(100) NOT NULL,
    carrier_id           VARCHAR(50) NULL,
    carrier_name         VARCHAR(255) NOT NULL,
    carrier_scac         VARCHAR(10) NULL,
    carrier_booking_reference VARCHAR(100) NULL,
    carrier_booking_status VARCHAR(50) NULL,
    carrier_confirmation_reference VARCHAR(100) NULL,
    carrier_booking_error TEXT NULL,
    carrier_booked_at    DATETIME NULL,
    status               VARCHAR(50) NOT NULL DEFAULT 'DRAFT',
    origin_port          VARCHAR(100) NOT NULL,
    destination_port     VARCHAR(100) NOT NULL,
    vessel_name          VARCHAR(255) NULL,
    voyage_number        VARCHAR(100) NULL,
    etd                  DATETIME NULL,
    eta                  DATETIME NULL,
    cargo_summary        TEXT NULL,
    special_instructions TEXT NULL,
    created_by           VARCHAR(255) NULL,
    created_at           DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at           DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_bookings_org (org_id),
    INDEX idx_bookings_rfq (rfq_id),
    INDEX idx_bookings_quote (quote_id),
    INDEX idx_bookings_number (booking_number),
    INDEX idx_bookings_status (status),
    CONSTRAINT fk_bookings_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_bookings_rfq FOREIGN KEY (rfq_id) REFERENCES rfqs(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipments (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    rfq_id BIGINT,
    quote_id BIGINT,
    booking_id BIGINT,
    booking_number VARCHAR(100),
    mbl_number VARCHAR(100),
    hbl_number VARCHAR(100),
    carrier_scac VARCHAR(10) NOT NULL,
    vessel_name VARCHAR(255),
    voyage_number VARCHAR(50),
    origin_port VARCHAR(10) NOT NULL,
    destination_port VARCHAR(10) NOT NULL,
    container_numbers JSON,
    status VARCHAR(50) DEFAULT 'BOOKED',
    etd DATETIME,
    eta DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_ship_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_ship_carrier FOREIGN KEY (carrier_scac) REFERENCES carriers(scac)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_milestones (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    shipment_id BIGINT NOT NULL,
    milestone_name VARCHAR(100) NOT NULL,
    event_timestamp DATETIME NOT NULL,
    location VARCHAR(255),
    is_completed BOOLEAN DEFAULT FALSE,
    source VARCHAR(50) DEFAULT 'CARRIER_API',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_sm_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_exceptions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    shipment_id BIGINT NOT NULL,
    exception_type VARCHAR(100) NOT NULL,
    description TEXT,
    severity VARCHAR(20) DEFAULT 'MEDIUM',
    resolved BOOLEAN DEFAULT FALSE,
    resolved_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_se_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_processed_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    shipment_id BIGINT NOT NULL,
    event_fingerprint VARCHAR(255) UNIQUE NOT NULL,
    processed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_spe_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS carrier_tracking_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    event_id VARCHAR(255) NOT NULL,
    source_type VARCHAR(50) DEFAULT 'WEBHOOK',
    carrier_scac VARCHAR(10) NOT NULL,
    booking_number VARCHAR(100),
    container_number VARCHAR(50),
    mbl_number VARCHAR(100),
    hbl_number VARCHAR(100),
    vessel_name VARCHAR(255),
    voyage_number VARCHAR(50),
    milestone_code VARCHAR(50),
    event_time DATETIME NULL,
    location VARCHAR(255),
    raw_description TEXT,
    raw_payload JSON,
    shipment_id BIGINT NULL,
    matching_status VARCHAR(50) DEFAULT 'UNMATCHED',
    processing_status VARCHAR(50) DEFAULT 'RECEIVED',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_event_org (event_id, org_id),
    CONSTRAINT fk_cte_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 8. Documents & Finance
CREATE TABLE IF NOT EXISTS shipment_documents (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    shipment_id BIGINT NOT NULL,
    doc_type VARCHAR(100) NOT NULL,
    document_name VARCHAR(255) NULL,
    category VARCHAR(100) NOT NULL DEFAULT 'TRANSPORT',
    description TEXT NULL,
    s3_key VARCHAR(1000) NULL,
    file_name VARCHAR(500) NOT NULL,
    file_url VARCHAR(1000) NULL,
    file_size BIGINT NULL,
    mime_type VARCHAR(100) NULL,
    file_type VARCHAR(50) NULL DEFAULT 'PDF',
    status VARCHAR(50) NOT NULL DEFAULT 'MISSING',
    uploaded_by VARCHAR(255) NULL,
    uploaded_at DATETIME NULL,
    reviewed_by VARCHAR(255) NULL,
    reviewed_at DATETIME NULL,
    rejection_reason TEXT NULL,
    expires_at DATETIME NULL,
    document_date DATETIME NULL,
    reference_number VARCHAR(255) NULL,
    source VARCHAR(50) NOT NULL DEFAULT 'SHIPMENT',
    source_id BIGINT NULL,
    extracted_data JSON NULL,
    raw_ocr_text TEXT NULL,
    ai_summary TEXT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_ship_docs_shipment (org_id, shipment_id),
    INDEX idx_ship_docs_cat (org_id, shipment_id, category),
    CONSTRAINT fk_sd_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_sd_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_document_discrepancies (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    shipment_id BIGINT NOT NULL,
    field_name VARCHAR(100) NOT NULL,
    source_document VARCHAR(50) NOT NULL,
    target_document VARCHAR(50) NOT NULL,
    source_value TEXT,
    target_value TEXT,
    severity VARCHAR(20) DEFAULT 'MEDIUM',
    resolved BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_disc_field (shipment_id, field_name, source_document, target_document),
    CONSTRAINT fk_sdd_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_invoices (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    shipment_id BIGINT NOT NULL,
    invoice_number VARCHAR(100) NOT NULL,
    vendor_name VARCHAR(255) NOT NULL,
    vendor_ref VARCHAR(100),
    s3_key VARCHAR(512) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    total_amount DECIMAL(15,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'RECEIVED',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_org_inv (org_id, invoice_number),
    CONSTRAINT fk_si_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_si_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_invoice_items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    invoice_id BIGINT NOT NULL,
    charge_code VARCHAR(50) NOT NULL,
    description TEXT,
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_sii_inv FOREIGN KEY (invoice_id) REFERENCES shipment_invoices(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_finance_discrepancies (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    invoice_id BIGINT NOT NULL,
    charge_code VARCHAR(50) NOT NULL,
    field_name VARCHAR(100) NOT NULL,
    expected_value TEXT,
    actual_value TEXT,
    discrepancy_amount DECIMAL(15,2) DEFAULT 0,
    source VARCHAR(50) DEFAULT 'AUTO_AUDIT',
    status VARCHAR(50) DEFAULT 'OPEN',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_sfd_item (invoice_id, charge_code, field_name, source),
    CONSTRAINT fk_sfd_inv FOREIGN KEY (invoice_id) REFERENCES shipment_invoices(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_financial_charges (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    shipment_id BIGINT NOT NULL,
    booking_id BIGINT NULL,
    rfq_id BIGINT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'OTHER',
    charge_type VARCHAR(20) NOT NULL DEFAULT 'COST',
    description VARCHAR(255) NOT NULL,
    vendor_name VARCHAR(255) NULL,
    estimated_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    actual_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    reference_number VARCHAR(100) NULL,
    charge_date DATETIME NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ESTIMATED',
    notes TEXT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_sfc_org_shipment (org_id, shipment_id),
    INDEX idx_sfc_org_cat (org_id, category),
    INDEX idx_sfc_status (org_id, status),
    CONSTRAINT fk_sfc_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_sfc_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_customer_invoices (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    shipment_id BIGINT NOT NULL,
    invoice_number VARCHAR(100) NOT NULL,
    customer_id BIGINT NOT NULL,
    total_amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(50) DEFAULT 'DRAFT',
    due_date DATE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_sci_num (org_id, invoice_number),
    CONSTRAINT fk_sci_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_sci_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE,
    CONSTRAINT fk_sci_cust FOREIGN KEY (customer_id) REFERENCES customers(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_customer_invoice_items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    customer_invoice_id BIGINT NOT NULL,
    charge_code VARCHAR(50) NOT NULL,
    description TEXT,
    amount DECIMAL(15,2) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_scii_ci FOREIGN KEY (customer_invoice_id) REFERENCES shipment_customer_invoices(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_finance_profitability (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    shipment_id BIGINT UNIQUE NOT NULL,
    org_id BIGINT NOT NULL,
    total_sell_amount DECIMAL(15,2) DEFAULT 0,
    total_buy_amount DECIMAL(15,2) DEFAULT 0,
    net_profit DECIMAL(15,2) DEFAULT 0,
    profit_margin_pct DECIMAL(5,2) DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_sfp_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE,
    CONSTRAINT fk_sfp_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 9. Audit Logs, Activities & Invitations
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    user_id BIGINT,
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(255) NOT NULL,
    resource_id BIGINT NOT NULL,
    details JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_al_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_al_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS activities (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    entity_type VARCHAR(255) NOT NULL,
    entity_id BIGINT NOT NULL,
    action VARCHAR(255) NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_act_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS invitations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_org_inv_email (org_id, email),
    CONSTRAINT fk_inv_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_inv_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_tracking_positions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    shipment_id BIGINT NOT NULL,
    vessel_name VARCHAR(255) NULL,
    latitude DECIMAL(10, 6) NOT NULL,
    longitude DECIMAL(10, 6) NOT NULL,
    speed_knots DECIMAL(6, 2) NOT NULL DEFAULT 0.00,
    heading_degrees DECIMAL(6, 2) NOT NULL DEFAULT 0.00,
    location_name VARCHAR(255) NULL,
    tracking_source VARCHAR(100) NOT NULL DEFAULT 'CARRIER_API',
    data_freshness VARCHAR(50) NOT NULL DEFAULT 'RECENT',
    recorded_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_stp_org_shipment (org_id, shipment_id, recorded_at DESC),
    INDEX idx_stp_recorded (recorded_at DESC),
    CONSTRAINT fk_stp_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_stp_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_tracking_alerts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    shipment_id BIGINT NOT NULL,
    alert_key VARCHAR(150) NOT NULL,
    alert_type VARCHAR(100) NOT NULL,
    severity VARCHAR(30) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'OPEN',
    first_detected_at DATETIME NOT NULL,
    last_detected_at DATETIME NOT NULL,
    acknowledged_at DATETIME NULL,
    acknowledged_by BIGINT NULL,
    resolved_at DATETIME NULL,
    resolved_by BIGINT NULL,
    suppressed_at DATETIME NULL,
    suppressed_by BIGINT NULL,
    notification_count INT NOT NULL DEFAULT 0,
    last_notified_at DATETIME NULL,
    metadata JSON NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_org_shipment_alert_key (org_id, shipment_id, alert_key),
    INDEX idx_sta_org_shipment_status (org_id, shipment_id, status),
    INDEX idx_sta_org_status (org_id, status),
    CONSTRAINT fk_sta_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_sta_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS shipment_tracking_refresh_runs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    shipment_id BIGINT NOT NULL,
    provider_name VARCHAR(100) NULL,
    provider_type VARCHAR(50) NULL,
    trigger_type VARCHAR(50) NOT NULL DEFAULT 'MANUAL',
    status VARCHAR(50) NOT NULL DEFAULT 'STARTED',
    started_at DATETIME NOT NULL,
    completed_at DATETIME NULL,
    new_positions INT NOT NULL DEFAULT 0,
    new_events INT NOT NULL DEFAULT 0,
    data_freshness VARCHAR(50) NULL,
    used_fallback BOOLEAN NOT NULL DEFAULT FALSE,
    error_message TEXT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_trkr_org_shipment_started (org_id, shipment_id, started_at DESC),
    INDEX idx_trkr_org_status_started (org_id, status, started_at DESC),
    CONSTRAINT fk_trkr_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_trkr_ship FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET FOREIGN_KEY_CHECKS = 1;
-- Migration: 051_quotation_foundation.sql
-- Description: Create quotations table as a first-class commercial entity with full lifecycle management

CREATE TABLE IF NOT EXISTS quotations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,

    quotation_number VARCHAR(50) NOT NULL,

    customer_id BIGINT NULL,
    customer_name VARCHAR(255) NULL,
    rfq_id BIGINT NULL,
    rfq_number VARCHAR(100) NULL,

    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT',

    origin VARCHAR(255) NULL,
    origin_code VARCHAR(20) NULL,
    destination VARCHAR(255) NULL,
    destination_code VARCHAR(20) NULL,

    service_type VARCHAR(100) NULL,
    transport_mode VARCHAR(100) NULL,

    currency VARCHAR(10) NOT NULL DEFAULT 'USD',

    subtotal DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    surcharges DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    taxes DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    valid_from DATE NULL,
    valid_until DATE NULL,

    sent_at DATETIME NULL,
    viewed_at DATETIME NULL,
    accepted_at DATETIME NULL,
    rejected_at DATETIME NULL,
    expired_at DATETIME NULL,

    notes TEXT NULL,

    created_by VARCHAR(255) NULL,
    updated_by VARCHAR(255) NULL,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uq_quotation_org_number (org_id, quotation_number),
    INDEX idx_quotation_org_status (org_id, status),
    INDEX idx_quotation_org_customer (org_id, customer_id),
    INDEX idx_quotation_org_valid_until (org_id, valid_until),
    INDEX idx_quotation_org_created (org_id, created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Quotation activity log for timeline
CREATE TABLE IF NOT EXISTS quotation_activity (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    activity_type VARCHAR(100) NOT NULL,
    description TEXT NULL,
    actor VARCHAR(255) NULL,
    metadata JSON NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_qact_org_quote_created (org_id, quotation_id, created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Migration: 052_quotation_pricing.sql
-- Description: Create quotation_charge_items table for line-item pricing and charges, and add margin summary fields to quotations table.
CREATE TABLE IF NOT EXISTS quotation_charge_items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,

    charge_code VARCHAR(50) NOT NULL,
    charge_name VARCHAR(255) NOT NULL,
    charge_category VARCHAR(50) NOT NULL DEFAULT 'OTHER',
    charge_type VARCHAR(20) NOT NULL DEFAULT 'SELL',
    calculation_basis VARCHAR(50) NOT NULL DEFAULT 'FLAT',

    quantity DECIMAL(15,4) NOT NULL DEFAULT 1.0000,
    unit_price DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    cost_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    sell_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    exchange_rate DECIMAL(15,6) NOT NULL DEFAULT 1.000000,

    tax_rate DECIMAL(8,4) NOT NULL DEFAULT 0.0000,
    tax_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    discount_type VARCHAR(20) NOT NULL DEFAULT 'NONE',
    discount_value DECIMAL(15,4) NOT NULL DEFAULT 0.0000,
    discount_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    total_cost DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    total_sell DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    display_order INT NOT NULL DEFAULT 0,
    is_optional BOOLEAN NOT NULL DEFAULT FALSE,
    notes TEXT NULL,

    created_by VARCHAR(255) NULL,
    updated_by VARCHAR(255) NULL,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_qcharges_org_quote (org_id, quotation_id),
    INDEX idx_qcharges_org_quote_order (org_id, quotation_id, display_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Extend quotations table with financial margin summary fields
ALTER TABLE quotations
    ADD COLUMN total_cost DECIMAL(15,2) NOT NULL DEFAULT 0.00 AFTER total_amount,
    ADD COLUMN gross_profit DECIMAL(15,2) NOT NULL DEFAULT 0.00 AFTER total_cost,
    ADD COLUMN gross_margin_pct DECIMAL(8,4) NOT NULL DEFAULT 0.0000 AFTER gross_profit;

-- Migration: 053_quotation_templates_and_terms.sql
-- Description: Create quotation_templates and quotation_template_charge_items tables, and extend quotations with commercial terms & notes.
CREATE TABLE IF NOT EXISTS quotation_templates (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,

    name VARCHAR(255) NOT NULL,
    description TEXT NULL,

    shipment_mode VARCHAR(50) NULL,
    transport_mode VARCHAR(50) NULL,
    origin VARCHAR(255) NULL,
    destination VARCHAR(255) NULL,

    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    validity_days INT NOT NULL DEFAULT 30,

    payment_terms VARCHAR(100) NOT NULL DEFAULT 'PREPAID',
    commercial_terms TEXT NULL,
    customer_notes TEXT NULL,
    internal_notes TEXT NULL,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by VARCHAR(255) NULL,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_qtmpl_org_active (org_id, is_active),
    INDEX idx_qtmpl_org_name (org_id, name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS quotation_template_charge_items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    template_id BIGINT NOT NULL,

    charge_category VARCHAR(50) NOT NULL DEFAULT 'FREIGHT',
    charge_code VARCHAR(50) NOT NULL,
    charge_name VARCHAR(255) NOT NULL,
    calculation_basis VARCHAR(50) NOT NULL DEFAULT 'PER_CONTAINER',

    quantity DECIMAL(15,4) NOT NULL DEFAULT 1.0000,
    unit_price DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    cost_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    discount_type VARCHAR(20) NOT NULL DEFAULT 'NONE',
    discount_value DECIMAL(15,4) NOT NULL DEFAULT 0.0000,
    tax_rate DECIMAL(8,4) NOT NULL DEFAULT 0.0000,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',

    display_order INT NOT NULL DEFAULT 0,
    is_optional BOOLEAN NOT NULL DEFAULT FALSE,
    notes TEXT NULL,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_qtmpl_charges_org_tmpl (org_id, template_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Extend quotations with commercial terms, payment terms, split notes, and template lineage reference
ALTER TABLE quotations
    ADD COLUMN payment_terms VARCHAR(100) NOT NULL DEFAULT 'PREPAID' AFTER currency,
    ADD COLUMN commercial_terms TEXT NULL AFTER notes,
    ADD COLUMN customer_notes TEXT NULL AFTER commercial_terms,
    ADD COLUMN internal_notes TEXT NULL AFTER customer_notes,
    ADD COLUMN template_id BIGINT NULL AFTER internal_notes;

-- Migration: 054_quotation_approval_and_lifecycle.sql
-- Extend quotations table with lifecycle and audit fields
ALTER TABLE quotations
    ADD COLUMN submitted_for_review_at DATETIME NULL AFTER notes,
    ADD COLUMN submitted_for_review_by VARCHAR(255) NULL AFTER submitted_for_review_at,
    ADD COLUMN approved_at DATETIME NULL AFTER submitted_for_review_by,
    ADD COLUMN approved_by VARCHAR(255) NULL AFTER approved_at,
    ADD COLUMN approval_notes TEXT NULL AFTER approved_by,
    ADD COLUMN changes_requested_at DATETIME NULL AFTER approval_notes,
    ADD COLUMN changes_requested_by VARCHAR(255) NULL AFTER changes_requested_at,
    ADD COLUMN changes_requested_reason TEXT NULL AFTER changes_requested_by,
    ADD COLUMN sent_by VARCHAR(255) NULL AFTER sent_at,
    ADD COLUMN first_viewed_at DATETIME NULL AFTER viewed_at,
    ADD COLUMN last_viewed_at DATETIME NULL AFTER first_viewed_at,
    ADD COLUMN view_count INT NOT NULL DEFAULT 0 AFTER last_viewed_at,
    ADD COLUMN declined_at DATETIME NULL AFTER accepted_at,
    ADD COLUMN declined_reason TEXT NULL AFTER declined_at,
    ADD COLUMN cancelled_at DATETIME NULL AFTER expired_at,
    ADD COLUMN cancelled_by VARCHAR(255) NULL AFTER cancelled_at,
    ADD COLUMN cancelled_reason TEXT NULL AFTER cancelled_by;

-- Quotation approval history table for audit trails
CREATE TABLE IF NOT EXISTS quotation_approval_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    action VARCHAR(50) NOT NULL,
    previous_status VARCHAR(50) NOT NULL,
    new_status VARCHAR(50) NOT NULL,
    actor_user_id BIGINT NULL,
    actor_name VARCHAR(255) NULL,
    comments TEXT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_qah_org_quote_created (org_id, quotation_id, created_at DESC),
    INDEX idx_qah_org_quote (org_id, quotation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Quotation public views tracking for customer views
CREATE TABLE IF NOT EXISTS quotation_public_views (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    viewer_name VARCHAR(255) NULL,
    viewer_email VARCHAR(255) NULL,
    ip_address VARCHAR(100) NULL,
    user_agent VARCHAR(500) NULL,
    viewed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_qpv_org_quote (org_id, quotation_id, viewed_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 055_quotation_documents_and_public_sharing.sql
CREATE TABLE IF NOT EXISTS quotation_documents (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    document_type VARCHAR(32) NOT NULL DEFAULT 'PDF',
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    generated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    generated_by BIGINT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_quotation_docs_org_quote (org_id, quotation_id),
    INDEX idx_quotation_docs_type (quotation_id, document_type),
    CONSTRAINT fk_quotation_documents_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_quotation_documents_quote FOREIGN KEY (quotation_id) REFERENCES quotations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS quotation_public_links (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    public_token VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    expires_at DATETIME NULL,
    created_by BIGINT NOT NULL,
    revoked_at DATETIME NULL,
    revoked_by BIGINT NULL,
    revocation_reason VARCHAR(255) NULL,
    last_accessed_at DATETIME NULL,
    access_count INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX uq_quote_public_token (public_token),
    INDEX idx_quote_public_links_org_quote (org_id, quotation_id),
    INDEX idx_quote_public_links_status (quotation_id, status),
    CONSTRAINT fk_quote_public_links_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_quote_public_links_quote FOREIGN KEY (quotation_id) REFERENCES quotations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 056_quotation_booking_conversion.sql
ALTER TABLE quotations
    ADD COLUMN IF NOT EXISTS converted_at DATETIME NULL AFTER cancelled_reason,
    ADD COLUMN IF NOT EXISTS converted_by VARCHAR(255) NULL AFTER converted_at,
    ADD COLUMN IF NOT EXISTS converted_booking_id BIGINT NULL AFTER converted_by,
    ADD COLUMN IF NOT EXISTS converted_shipment_id BIGINT NULL AFTER converted_booking_id,
    ADD COLUMN IF NOT EXISTS conversion_status VARCHAR(50) NOT NULL DEFAULT 'NOT_CONVERTED' AFTER converted_shipment_id,
    ADD COLUMN IF NOT EXISTS conversion_notes TEXT NULL AFTER conversion_status;

ALTER TABLE quotations
    ADD INDEX idx_quotation_org_booking (org_id, converted_booking_id),
    ADD INDEX idx_quotation_org_shipment (org_id, converted_shipment_id),
    ADD INDEX idx_quotation_org_conv_status (org_id, conversion_status);

CREATE TABLE IF NOT EXISTS quotation_conversion_history (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    quotation_id BIGINT NOT NULL,
    booking_id BIGINT NULL,
    shipment_id BIGINT NULL,
    action VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    message TEXT NULL,
    performed_by VARCHAR(255) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_qch_org_quote_created (org_id, quotation_id, created_at DESC),
    INDEX idx_qch_org_booking (org_id, booking_id),
    CONSTRAINT fk_quotation_conversion_history_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_quotation_conversion_history_quote FOREIGN KEY (quotation_id) REFERENCES quotations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS carrier_sync_jobs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    carrier_integration_id BIGINT NOT NULL,
    operation VARCHAR(50) NOT NULL DEFAULT 'TRACKING',
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME NULL,
    records_processed INT NOT NULL DEFAULT 0,
    records_created INT NOT NULL DEFAULT 0,
    records_updated INT NOT NULL DEFAULT 0,
    records_failed INT NOT NULL DEFAULT 0,
    error_code VARCHAR(100) NULL,
    error_message TEXT NULL,
    correlation_id VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_csj_org_int (org_id, carrier_integration_id),
    INDEX idx_csj_int_started (carrier_integration_id, started_at DESC),
    INDEX idx_csj_status (status),
    INDEX idx_csj_corr (correlation_id),
    CONSTRAINT fk_csj_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_csj_integration FOREIGN KEY (carrier_integration_id) REFERENCES carrier_integrations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS carrier_webhook_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    carrier_integration_id BIGINT NOT NULL,
    carrier_scac VARCHAR(20) NOT NULL,
    provider_event_id VARCHAR(150) NULL,
    event_type VARCHAR(100) NOT NULL,
    event_fingerprint VARCHAR(128) NOT NULL,
    received_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at DATETIME NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    error_message TEXT NULL,
    correlation_id VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_cwe_fingerprint (org_id, event_fingerprint),
    INDEX idx_cwe_org_int (org_id, carrier_integration_id),
    INDEX idx_cwe_status (status),
    INDEX idx_cwe_provider_evt (provider_event_id),
    CONSTRAINT fk_cwe_org FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE,
    CONSTRAINT fk_cwe_integration FOREIGN KEY (carrier_integration_id) REFERENCES carrier_integrations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;




