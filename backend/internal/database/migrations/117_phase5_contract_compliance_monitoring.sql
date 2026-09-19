-- 117_phase5_contract_compliance_monitoring.sql
-- LogisticsHQ Phase 5 Task 5.7: Contract and Compliance Monitoring Schema

CREATE TABLE IF NOT EXISTS contract_compliance_monitoring_plans (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    org_id BIGINT NOT NULL,
    contract_id BIGINT NOT NULL,
    contract_reference VARCHAR(100) NOT NULL,
    contract_name VARCHAR(255) NOT NULL,
    party_name VARCHAR(255) NOT NULL,
    contract_type VARCHAR(50) NOT NULL DEFAULT 'CUSTOMER_SLA',
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    effective_date DATE NULL,
    expiry_date DATE NULL,
    days_until_expiration INT NOT NULL DEFAULT 0,
    expiration_status VARCHAR(50) NOT NULL DEFAULT 'CURRENT',
    compliance_status VARCHAR(50) NOT NULL DEFAULT 'COMPLIANT',
    hard_requirement_count INT NOT NULL DEFAULT 0,
    soft_requirement_count INT NOT NULL DEFAULT 0,
    hard_violations_count INT NOT NULL DEFAULT 0,
    soft_deviations_count INT NOT NULL DEFAULT 0,
    missing_documents_count INT NOT NULL DEFAULT 0,
    expired_documents_count INT NOT NULL DEFAULT 0,
    risk_level VARCHAR(50) NOT NULL DEFAULT 'LOW',
    risk_score DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    recommended_remediation_strategy_id VARCHAR(100) NOT NULL,
    selected_remediation_strategy_id VARCHAR(100) NOT NULL,
    candidate_strategies LONGTEXT NULL,
    authoritative_facts LONGTEXT NULL,
    extracted_terms LONGTEXT NULL,
    predictions LONGTEXT NULL,
    assumptions LONGTEXT NULL,
    remediation_plan_steps LONGTEXT NULL,
    deviations LONGTEXT NULL,
    requires_approval TINYINT(1) NOT NULL DEFAULT 0,
    approval_reason VARCHAR(500) NULL,
    autonomy_level VARCHAR(50) NOT NULL DEFAULT 'LEVEL_2_PREPARE',
    stop_reason VARCHAR(500) NULL,
    execution_status VARCHAR(50) NOT NULL DEFAULT 'GENERATED',
    version INT NOT NULL DEFAULT 1,
    idempotency_key VARCHAR(128) NOT NULL,
    correlation_id VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_ccm_plan_org_idemp (org_id, idempotency_key),
    INDEX idx_ccm_plan_org_contract (org_id, contract_id),
    INDEX idx_ccm_plan_org_status (org_id, compliance_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS contract_compliance_monitoring_versions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    plan_id BIGINT NOT NULL,
    org_id BIGINT NOT NULL,
    version_number INT NOT NULL,
    trigger_event VARCHAR(100) NOT NULL,
    compliance_status VARCHAR(50) NOT NULL,
    strategy_id VARCHAR(100) NOT NULL,
    change_reason VARCHAR(500) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ccm_ver_plan (org_id, plan_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed default autonomy policies for module 'contract_compliance' if not existing
INSERT INTO autonomy_policies (
    org_id, module, autonomy_level, allowed_action_types, prohibited_action_types,
    requires_approval, max_monetary_threshold, customer_impact_threshold,
    shipment_impact_threshold, compliance_sensitivity, min_confidence_threshold,
    require_data_sufficiency, max_plan_steps, max_execution_attempts,
    cooldown_seconds, emergency_stop, is_active, policy_version, created_at, updated_at
)
SELECT 1, 'contract_compliance', 'LEVEL_2_PREPARE',
       '["request_document_renewal", "notify_compliance_officer", "log_deviation_audit"]',
       '["autonomous_contract_modification", "compliance_waiver", "autonomous_contract_signing"]',
       1, 0.00, 'HIGH', 'HIGH', 'HIGH', 0.8500, 1, 7, 3, 300, 0, 1, 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM autonomy_policies WHERE org_id = 1 AND module = 'contract_compliance');

INSERT INTO autonomy_policies (
    org_id, module, autonomy_level, allowed_action_types, prohibited_action_types,
    requires_approval, max_monetary_threshold, customer_impact_threshold,
    shipment_impact_threshold, compliance_sensitivity, min_confidence_threshold,
    require_data_sufficiency, max_plan_steps, max_execution_attempts,
    cooldown_seconds, emergency_stop, is_active, policy_version, created_at, updated_at
)
SELECT 2, 'contract_compliance', 'LEVEL_2_PREPARE',
       '["request_document_renewal", "notify_compliance_officer", "log_deviation_audit"]',
       '["autonomous_contract_modification", "compliance_waiver", "autonomous_contract_signing"]',
       1, 0.00, 'HIGH', 'HIGH', 'HIGH', 0.8500, 1, 7, 3, 300, 0, 1, 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM autonomy_policies WHERE org_id = 2 AND module = 'contract_compliance');

-- Ensure Org 2 has realistic Contracts for demonstration and testing
INSERT INTO contracts (
    id, org_id, contract_reference, contract_name, contract_type, party_id, party_name,
    transport_mode, status, currency, contract_value, effective_date, expiry_date,
    owner, description, notes, created_by, created_at, updated_at
)
SELECT 201, 2, 'CTR-TP-2026-ORG2', 'Trans-Pacific Carrier Master Service Agreement', 'CARRIER_AGREEMENT', 101, 'Hapag-Lloyd AG',
       'OCEAN', 'ACTIVE', 'USD', 650000.00, '2026-01-01', '2027-06-30',
       'Senior Procurement Director', 'Annual volume commitment for Far East to US West Coast corridor.',
       'Standard FMC-filed agreement with low sulfur bunker adjustment formula.', 'admin@freel.com', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM contracts WHERE org_id = 2 AND id = 201);

INSERT INTO contracts (
    id, org_id, contract_reference, contract_name, contract_type, party_id, party_name,
    transport_mode, status, currency, contract_value, effective_date, expiry_date,
    owner, description, notes, created_by, created_at, updated_at
)
SELECT 202, 2, 'SLA-NORDIC-2025', 'Nordic Dynamics Global Logistics SLA', 'CUSTOMER_SLA', 102, 'Nordic Freight Dynamics AB',
       'MULTIMODAL', 'ACTIVE', 'USD', 420000.00, '2025-09-25', '2026-09-28',
       'Commercial Account Lead', 'Master customer service level agreement for Northern European shipments.',
       'Approaching expiration window within 16 days. Renewal review pending.', 'admin@freel.com', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM contracts WHERE org_id = 2 AND id = 202);

INSERT INTO contracts (
    id, org_id, contract_reference, contract_name, contract_type, party_id, party_name,
    transport_mode, status, currency, contract_value, effective_date, expiry_date,
    owner, description, notes, created_by, created_at, updated_at
)
SELECT 203, 2, 'CTR-COLD-2026', 'Global Cold-Chain Pharma Agreement', 'CUSTOMER_SLA', 101, 'Apex Global Logistics Corp',
       'AIR', 'ACTIVE', 'USD', 890000.00, '2026-03-01', '2027-03-01',
       'Compliance Specialist', 'Temperature-controlled pharmaceutical transport agreement with strict GDP validation.',
       'Audit flag: Mandatory Good Distribution Practice (GDP) validation certificate missing.', 'admin@freel.com', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM contracts WHERE org_id = 2 AND id = 203);

INSERT INTO contracts (
    id, org_id, contract_reference, contract_name, contract_type, party_id, party_name,
    transport_mode, status, currency, contract_value, effective_date, expiry_date,
    owner, description, notes, created_by, created_at, updated_at
)
SELECT 204, 2, 'AIR-EXP-2024', 'Trans-Atlantic Express Airfreight Contract', 'CARRIER_AGREEMENT', 103, 'Lufthansa Cargo',
       'AIR', 'EXPIRED', 'EUR', 310000.00, '2024-08-01', '2025-08-01',
       'Operations Manager', 'Legacy airfreight agreement for Frankfurt to Chicago express route.',
       'Expired contract. No active shipments permitted under this agreement.', 'admin@freel.com', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM contracts WHERE org_id = 2 AND id = 204);

-- Seed Contract Terms for Org 2 contracts
INSERT INTO contract_terms (
    org_id, contract_id, term_category, term_key, term_title, term_value,
    value_type, currency, is_critical, created_at, updated_at
)
SELECT 2, 201, 'COMMERCIAL', 'ocean_base_rate_40hc', 'Base Freight 40HC Far East-USWC', '2850.00', 'DECIMAL', 'USD', 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM contract_terms WHERE org_id = 2 AND contract_id = 201 AND term_key = 'ocean_base_rate_40hc');

INSERT INTO contract_terms (
    org_id, contract_id, term_category, term_key, term_title, term_value,
    value_type, currency, is_critical, created_at, updated_at
)
SELECT 2, 202, 'OPERATIONAL', 'sla_on_time_delivery', 'Minimum On-Time Delivery Performance', '98.5%', 'PERCENTAGE', NULL, 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM contract_terms WHERE org_id = 2 AND contract_id = 202 AND term_key = 'sla_on_time_delivery');

INSERT INTO contract_terms (
    org_id, contract_id, term_category, term_key, term_title, term_value,
    value_type, currency, is_critical, created_at, updated_at
)
SELECT 2, 203, 'COMPLIANCE', 'temperature_range', 'Allowed Temperature Range (+2C to +8C)', 'GDP_PHARMA_COLD', 'STRING', NULL, 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM contract_terms WHERE org_id = 2 AND contract_id = 203 AND term_key = 'temperature_range');

-- Seed Compliance Requirements for Org 2 contracts
INSERT INTO contract_compliance_requirements (
    org_id, contract_id, requirement_type, title, description,
    responsible_party, valid_from, valid_until, status, risk_severity, created_at, updated_at
)
SELECT 2, 201, 'REGULATORY', 'FMC Carrier Agreement Filing', 'Mandatory Federal Maritime Commission tariff publication',
       'CARRIER', '2026-01-01', '2027-06-30', 'VERIFIED', 'HIGH', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM contract_compliance_requirements WHERE org_id = 2 AND contract_id = 201 AND requirement_type = 'REGULATORY');

INSERT INTO contract_compliance_requirements (
    org_id, contract_id, requirement_type, title, description,
    responsible_party, valid_from, valid_until, status, risk_severity, created_at, updated_at
)
SELECT 2, 202, 'INSURANCE', 'Comprehensive Cargo Liability Insurance', 'Carrier liability cover minimum $1,000,000 per shipment',
       'CARRIER', '2025-09-25', '2026-09-28', 'EXPIRING', 'MEDIUM', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM contract_compliance_requirements WHERE org_id = 2 AND contract_id = 202 AND requirement_type = 'INSURANCE');

INSERT INTO contract_compliance_requirements (
    org_id, contract_id, requirement_type, title, description,
    responsible_party, valid_from, valid_until, status, risk_severity, created_at, updated_at
)
SELECT 2, 203, 'MANDATORY_CERTIFICATE', 'Good Distribution Practice (GDP) Certificate', 'WHO/FDA compliant pharma GDP certification for refrigerated handling',
       'SHIPPER', '2026-03-01', '2027-03-01', 'MISSING', 'CRITICAL', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM contract_compliance_requirements WHERE org_id = 2 AND contract_id = 203 AND requirement_type = 'MANDATORY_CERTIFICATE');
