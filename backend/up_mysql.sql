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
    carrier_name VARCHAR(255),
    buy_price DECIMAL(15,2) DEFAULT 0,
    sell_price DECIMAL(15,2) DEFAULT 0,
    ocean_freight DECIMAL(15,2) DEFAULT 0,
    origin_charges DECIMAL(15,2) DEFAULT 0,
    destination_charges DECIMAL(15,2) DEFAULT 0,
    total_buy_price DECIMAL(15,2) DEFAULT 0,
    free_days INT DEFAULT 0,
    transit_time_days INT DEFAULT 0,
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

CREATE TABLE IF NOT EXISTS carrier_integrations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    carrier_scac VARCHAR(10) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_ci_org FOREIGN KEY (org_id) REFERENCES organizations(id),
    CONSTRAINT fk_ci_carrier FOREIGN KEY (carrier_scac) REFERENCES carriers(scac)
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

-- 7. Operations & Shipments
CREATE TABLE IF NOT EXISTS shipments (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    rfq_id BIGINT,
    quote_id BIGINT,
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
    doc_type VARCHAR(50) NOT NULL,
    s3_key VARCHAR(512) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'UPLOADED',
    extracted_data JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_ship_doc (shipment_id, doc_type),
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

SET FOREIGN_KEY_CHECKS = 1;
