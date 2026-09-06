-- Ensure contract_parties exist for org_id = 1
INSERT INTO contract_parties (id, org_id, party_name, party_type, contact_name, contact_email) VALUES
(1, 1, 'Maersk Line', 'CARRIER', 'Commercial Desk', 'commercial@maersk.com'),
(2, 1, 'Acme Corp Industries', 'CUSTOMER', 'John Doe', 'john@acme.com'),
(3, 1, 'Apex Drayage & Intermodal', 'VENDOR', 'Dispatcher Rick', 'ops@apexdray.com'),
(4, 1, 'Cargolux Airlines', 'CARRIER', 'Charter Desk', 'charter@cargolux.com')
ON DUPLICATE KEY UPDATE party_name = VALUES(party_name);

-- Contract 1: Draft Carrier Contract
INSERT INTO contracts (
    id, org_id, contract_name, contract_reference, contract_type, 
    party_id, party_name, status, transport_mode, 
    contract_value, currency, effective_date, expiry_date, 
    description, notes, owner, created_by
) VALUES (
    101, 1, 'Trans-Pacific Master Agreement 2026', 'CTR-TP-2026-01', 'CARRIER_AGREEMENT',
    1, 'Maersk Line', 'DRAFT', 'OCEAN',
    750000.00, 'USD', '2026-10-01', '2027-09-30',
    'Annual carrier volume commitment for Trans-Pacific Eastbound lanes covering 4,000 TEU target.',
    'Negotiation on bunker adjustment clause in final review with procurement team.', 'Varun Kanade', 'varunkanade3456@gmail.com'
) ON DUPLICATE KEY UPDATE contract_name = VALUES(contract_name);

-- Contract 2: Active Healthy Customer SLA
INSERT INTO contracts (
    id, org_id, contract_name, contract_reference, contract_type, 
    party_id, party_name, status, transport_mode, 
    contract_value, currency, effective_date, expiry_date, 
    description, notes, owner, created_by
) VALUES (
    102, 1, 'Acme Global Supply Chain Master SLA', 'SLA-ACME-2025', 'CUSTOMER_SLA',
    2, 'Acme Corp Industries', 'ACTIVE', 'AIR',
    1200000.00, 'USD', '2026-01-01', '2027-03-31',
    'Tier-1 Customer Master Service Level Agreement with guaranteed 48-hour transit KPI and customized pricing matrix.',
    'Quarterly rebate schedule applies if volume exceeds 500 tons per month.', 'Varun Kanade', 'varunkanade3456@gmail.com'
) ON DUPLICATE KEY UPDATE contract_name = VALUES(contract_name);

-- Contract 3: Active Expiring Soon Vendor Logistics Agreement (Expires in ~18 days)
INSERT INTO contracts (
    id, org_id, contract_name, contract_reference, contract_type, 
    party_id, party_name, status, transport_mode, 
    contract_value, currency, effective_date, expiry_date, 
    description, notes, owner, created_by
) VALUES (
    103, 1, 'Tri-State Drayage & Intermodal Contract', 'VEND-DRAY-2025-Q3', 'VENDOR_AGREEMENT',
    3, 'Apex Drayage & Intermodal', 'ACTIVE', 'ROAD',
    280000.00, 'USD', '2025-09-15', DATE_ADD(CURDATE(), INTERVAL 18 DAY),
    'Dedicated port-to-rail drayage services for Los Angeles / Long Beach and Oakland ramps.',
    'Renewal reminder sent. Negotiating fuel surcharge index for 2026-2027 cycle.', 'Varun Kanade', 'varunkanade3456@gmail.com'
) ON DUPLICATE KEY UPDATE contract_name = VALUES(contract_name);

-- Contract 4: Expired Contract
INSERT INTO contracts (
    id, org_id, contract_name, contract_reference, contract_type, 
    party_id, party_name, status, transport_mode, 
    contract_value, currency, effective_date, expiry_date, 
    description, notes, owner, created_by
) VALUES (
    104, 1, 'Asia-Europe Airfreight Charter Agreement', 'AIR-CH-2024-EUR', 'CARRIER_AGREEMENT',
    4, 'Cargolux Airlines', 'EXPIRED', 'AIR',
    450000.00, 'USD', '2024-06-01', '2025-05-31',
    'Weekly scheduled Boeing 747-8F capacity allocation on PVG-LUX lane.',
    'Agreement completed and archived. Succeeded by 2025 charter schedule.', 'Varun Kanade', 'varunkanade3456@gmail.com'
) ON DUPLICATE KEY UPDATE contract_name = VALUES(contract_name);

-- Lifecycle events for Contract 101 (Draft):
INSERT INTO contract_lifecycle_events (org_id, contract_id, event_type, previous_status, new_status, description, performed_by)
VALUES (1, 101, 'CREATED', NULL, 'DRAFT', 'Contract drafted for upcoming 2026 ocean tender negotiation', 'varunkanade3456@gmail.com');

-- Lifecycle events for Contract 102 (Acme SLA):
INSERT INTO contract_lifecycle_events (org_id, contract_id, event_type, previous_status, new_status, description, performed_by)
VALUES 
(1, 102, 'CREATED', NULL, 'DRAFT', 'Contract initialized via Commercial Contract Hub', 'varunkanade3456@gmail.com'),
(1, 102, 'LINK_ADDED', 'DRAFT', 'DRAFT', 'Linked Quotation QT-2025-8840 as primary commercial rate basis', 'varunkanade3456@gmail.com'),
(1, 102, 'STATUS_CHANGED', 'DRAFT', 'ACTIVE', 'Contract fully executed and signed by both parties', 'Varun Kanade');

-- Lifecycle events for Contract 103 (Expiring Soon):
INSERT INTO contract_lifecycle_events (org_id, contract_id, event_type, previous_status, new_status, description, performed_by)
VALUES 
(1, 103, 'CREATED', NULL, 'DRAFT', 'Draft vendor contract drafted from logistics template', 'varunkanade3456@gmail.com'),
(1, 103, 'STATUS_CHANGED', 'DRAFT', 'ACTIVE', 'Vendor onboarding verified and contract activated', 'Operations Manager'),
(1, 103, 'LINK_ADDED', 'ACTIVE', 'ACTIVE', 'Linked Managed Rate Contract MRC-DRAY-009', 'Operations Manager');

-- Lifecycle events for Contract 104 (Expired):
INSERT INTO contract_lifecycle_events (org_id, contract_id, event_type, previous_status, new_status, description, performed_by)
VALUES 
(1, 104, 'CREATED', NULL, 'DRAFT', 'Draft air charter agreement drafted', 'Charter Desk'),
(1, 104, 'STATUS_CHANGED', 'DRAFT', 'ACTIVE', 'Activated for 2024-2025 season', 'Charter Desk'),
(1, 104, 'STATUS_CHANGED', 'ACTIVE', 'EXPIRED', 'Contract term ended and closed out', 'System Lifecycle Engine');

-- Contract Links for Contract 102 (Acme SLA):
INSERT INTO contract_links (org_id, contract_id, linked_entity_type, linked_entity_id, link_type, is_primary, notes, created_by)
VALUES 
(1, 102, 'QUOTATION', 1, 'QUOTATION', 1, 'Primary commercial quote with contracted lane pricing', 'varunkanade3456@gmail.com'),
(1, 102, 'CARRIER_RATE_CONTRACT', 1, 'RATE_CONTRACT', 0, 'Carrier capacity backing SLA commitments', 'varunkanade3456@gmail.com'),
(1, 102, 'CUSTOMER', 2, 'PARTY', 1, 'Customer legal entity profile link', 'varunkanade3456@gmail.com');

-- Contract Links for Contract 103:
INSERT INTO contract_links (org_id, contract_id, linked_entity_type, linked_entity_id, link_type, is_primary, notes, created_by)
VALUES 
(1, 103, 'CARRIER_RATE_CONTRACT', 1, 'RATE_CONTRACT', 1, 'Standard drayage tariffs table', 'varunkanade3456@gmail.com'),
(1, 103, 'CARRIER', 3, 'PARTY', 1, 'Vendor operating profile', 'varunkanade3456@gmail.com');
