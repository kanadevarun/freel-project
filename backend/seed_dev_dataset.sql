-- ==============================================================================
-- LOGISTICSHQ CONTROLLED REALISTIC DEVELOPMENT SEED DATA (ORG 2)
-- Organization: Freight Forwarder Workspace (Org ID: 2)
-- User: kanadevarun123@gmail.com (User ID: 6)
-- ==============================================================================

SET FOREIGN_KEY_CHECKS = 0;

-- 1. Ensure Organization 2 profile is realistic and complete
UPDATE organizations 
SET 
    name = 'LogisticsHQ Dev Org - Varun Logistics',
    legal_name = 'Varun Freight & Logistics Solutions Pvt Ltd',
    registration_number = 'U63090MH2026PTC398124',
    tax_number = '27AAACV1234F1Z8',
    website = 'https://varunlogistics.dev',
    primary_email = 'kanadevarun123@gmail.com',
    phone_number = '+91 98200 98765',
    address = 'Unit 401, Trade Center, Bandra Kurla Complex',
    city = 'Mumbai',
    state = 'Maharashtra',
    country = 'India',
    postal_code = '400051',
    industry = 'Freight Forwarding & 3PL Logistics',
    company_type = 'Freight Forwarder',
    default_currency = 'USD',
    default_timezone = 'UTC'
WHERE id = 2;

-- 2. Development Customers
INSERT INTO customers (
    id, org_id, customer_code, name, trading_name, domain, industry, customer_type, 
    tax_id, pan_number, eori_number, currency, payment_terms, credit_limit, health_score, 
    account_owner_id, country, city, contact_name, contact_email, contact_phone, status, created_at, updated_at
) VALUES
(101, 2, 'DEV-CUST-001', 'Apex Global Logistics Corp', 'Apex Global', 'apexlogistics.com', 'Manufacturing & Retail', 'SHIPPER',
 'US-EIN-987654321', 'AABCA1234B', 'GB123456789000', 'USD', 'NET30', 50000.00, 92,
 6, 'United States', 'Chicago', 'Vikram Malhotra', 'v.malhotra@apexlogistics.com', '+1 312 555 0199', 'ACTIVE', NOW(), NOW()),

(102, 2, 'DEV-CUST-002', 'Nordic Freight Dynamics AB', 'Nordic Dynamics', 'nordicfreight.se', 'Machinery & Heavy Equip', 'SHIPPER',
 'SE-VAT-556677889901', 'BBBCB2345C', 'SE556677889900', 'USD', 'NET15', 35000.00, 78,
 6, 'Sweden', 'Gothenburg', 'Astrid Lindgren', 'astrid@nordicfreight.se', '+46 31 123 4567', 'ACTIVE', NOW(), NOW()),

(103, 2, 'DEV-CUST-003', 'Bharat Tech Exports Pvt Ltd', 'Bharat Tech', 'bharatexports.in', 'Electronics & High-Tech', 'SHIPPER',
 '27AABCB9988D1Z5', 'CCCDC3456D', 'IN27AABCB9988D1', 'USD', 'NET45', 20000.00, 65,
 6, 'India', 'Mumbai', 'Karan Johar', 'karan@bharatexports.in', '+91 98111 22334', 'ACTIVE', NOW(), NOW())
ON DUPLICATE KEY UPDATE name = VALUES(name), status = 'ACTIVE';

-- 3. Customer Contacts
INSERT INTO contacts (
    id, org_id, customer_id, first_name, last_name, email, phone, job_title, department, is_primary
) VALUES
(101, 2, 101, 'Vikram', 'Malhotra', 'v.malhotra@apexlogistics.com', '+1 312 555 0199', 'Supply Chain Director', 'Procurement', TRUE),
(102, 2, 102, 'Astrid', 'Lindgren', 'astrid@nordicfreight.se', '+46 31 123 4567', 'Logistics Operations Lead', 'Logistics', TRUE),
(103, 2, 103, 'Karan', 'Johar', 'karan@bharatexports.in', '+91 98111 22334', 'Managing Director', 'Executive', TRUE)
ON DUPLICATE KEY UPDATE first_name = VALUES(first_name);

-- 4. Development Leads (including Lead eligible for scoring)
INSERT INTO leads (
    id, org_id, company_name, contact_name, email, phone, status, source, ai_score, notes, assigned_to, created_at, updated_at
) VALUES
(101, 2, 'Apex Global Logistics Corp', 'Vikram Malhotra', 'v.malhotra@apexlogistics.com', '+1 312 555 0199', 'CONVERTED', 'INBOUND_EMAIL', 92, 'Converted to customer DEV-CUST-001 with recurring 40GP shipments', 6, NOW(), NOW()),
(102, 2, 'Hindustan Petrochem Distributors', 'Rajiv Singhania', 'r.singhania@hindpetro.in', '+91 98222 33445', 'NEW', 'INBOUND_WEB', NULL, 'New inbound prospect requesting chemical bulk container ocean freight. Eligible for AI scoring.', 6, NOW(), NOW()),
(103, 2, 'Euro-Asia Retailers Ltd', 'Elena Rostova', 'elena@euroasia-retail.de', '+49 30 9876 5432', 'QUALIFIED', 'PARTNER_REFERRAL', 85, 'Qualified lead for European retail distribution lane', 6, NOW(), NOW())
ON DUPLICATE KEY UPDATE company_name = VALUES(company_name);

-- 5. Inbound Email / Lead Interaction (for Sales parser agent testing)
INSERT INTO lead_interactions (
    id, lead_id, org_id, interaction_type, summary, full_body, direction, sender, recipients, status, created_at, updated_at
) VALUES
(101, 102, 2, 'EMAIL', 'Inbound RFQ request for chemicals shipment',
 'Dear LogisticsHQ Team,\n\nWe would like to request an ocean freight quotation for 1x40GP container of industrial specialty polymers.\nOrigin: Nhava Sheva (INNSA)\nDestination: Rotterdam (NLRTM)\nIncoterms: FOB\nCargo ready date: 2026-10-15\nGross Weight: 21,500 kg\nVolume: 35 CBM\n\nPlease let us know your best available carrier rates and transit times.\n\nBest regards,\nRajiv Singhania\nHindustan Petrochem Distributors',
 'INBOUND', 'r.singhania@hindpetro.in', 'sales@logisticshq.in', 'RECEIVED', NOW(), NOW())
ON DUPLICATE KEY UPDATE summary = VALUES(summary);

-- 6. Carrier Rates in rate_entries (Contract Rates)
INSERT INTO rate_entries (
    id, org_id, source, source_ref, origin_port, destination_port, carrier_scac, carrier_name, 
    equipment_type, ocean_freight, origin_charges, destination_charges, total_buy_price, 
    currency_original, valid_from, valid_until, free_days_origin, free_days_destination, 
    transit_days, incoterms, extraction_status, created_at, updated_at
) VALUES
('rate-dev-001', 2, 'CONTRACT', 'SC-MAEU-2026-01', 'INNSA', 'NLRTM', 'MAEU', 'Maersk Line',
 '40GP', 1850.00, 150.00, 150.00, 2150.00, 'USD', '2026-01-01', '2026-12-31', 7, 14, 22, 'FOB', 'CONFIRMED', NOW(), NOW()),

('rate-dev-002', 2, 'CONTRACT', 'SC-MSCU-2026-02', 'INNSA', 'DEHAM', 'MSCU', 'MSC Mediterranean Shipping Co',
 '40GP', 1950.00, 160.00, 170.00, 2280.00, 'USD', '2026-01-01', '2026-12-31', 7, 14, 24, 'FOB', 'CONFIRMED', NOW(), NOW()),

('rate-dev-003', 2, 'CONTRACT', 'SC-CMDU-2026-03', 'INNSA', 'USNYC', 'CMDU', 'CMA CGM Group',
 '40GP', 2400.00, 180.00, 170.00, 2750.00, 'USD', '2026-01-01', '2026-12-31', 7, 14, 28, 'FOB', 'CONFIRMED', NOW(), NOW())
ON DUPLICATE KEY UPDATE carrier_name = VALUES(carrier_name);

-- 7. Pricing Rules
INSERT INTO pricing_rules (
    id, org_id, rule_name, rule_type, conditions, markup_type, markup_value, markup_pct, markup_flat, min_margin_pct, priority, is_active, created_at, updated_at
) VALUES
(101, 2, 'Default Development Markup Policy', 'DEFAULT', '{}', 'PERCENTAGE', 15.00, 15.00, 0.00, 5.00, 1, 1, NOW(), NOW()),
(102, 2, 'INNSA to DEHAM Lane Promotion', 'LANE', '{"origin": "INNSA", "destination": "DEHAM"}', 'PERCENTAGE', 12.00, 12.00, 0.00, 5.00, 10, 1, NOW(), NOW()),
(103, 2, 'High Volume Tier Discount', 'CUSTOMER_TIER', '{"tier": "ENTERPRISE"}', 'PERCENTAGE', 10.00, 10.00, 0.00, 4.00, 5, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE rule_name = VALUES(rule_name);

-- 8. RFQs (including Complete RFQ ready for pricing & Incomplete RFQ)
INSERT INTO rfqs (
    id, org_id, rfq_number, customer_id, stage, status, agent_status, origin, destination, incoterms, 
    target_date, sales_assignee_id, pricing_assignee_id, health_score, lead_id, created_at, updated_at
) VALUES
-- RFQ 101: Complete converted workflow RFQ
(101, 2, 'RFQ-2026-DEV-001', 101, 'WON', 'WON', 'COMPLETED', 'Nhava Sheva (INNSA)', 'Rotterdam (NLRTM)', 'FOB', 
 '2026-09-15 00:00:00', 6, 6, 95, 101, NOW(), NOW()),

-- RFQ 102: Complete RFQ READY FOR PRICING
(102, 2, 'RFQ-2026-DEV-002', 102, 'DRAFT', 'SUBMITTED', 'IDLE', 'INNSA', 'DEHAM', 'FOB', 
 '2026-10-20 00:00:00', 6, 6, 85, 103, NOW(), NOW()),

-- RFQ 103: Incomplete RFQ (missing destination and target date)
(103, 2, 'RFQ-2026-DEV-003', 103, 'DRAFT', 'DRAFT', 'INCOMPLETE', 'Nhava Sheva (INNSA)', NULL, NULL, 
 NULL, 6, 6, 40, 102, NOW(), NOW())
ON DUPLICATE KEY UPDATE rfq_number = VALUES(rfq_number);

-- 9. RFQ Items
INSERT INTO rfq_items (id, rfq_id, description, quantity, weight_kg, volume_cbm, created_at, updated_at) VALUES
(101, 101, 'Automotive Engine Components in steel crates', 1, 16000.00, 28.00, NOW(), NOW()),
(102, 102, 'Industrial Valves & High-Pressure Pumps', 1, 18500.00, 32.00, NOW(), NOW()),
(103, 103, 'Sample Textile Goods (Incomplete specs)', 1, NULL, NULL, NOW(), NOW())
ON DUPLICATE KEY UPDATE description = VALUES(description);

-- 10. Quotations (including Quotation awaiting review)
INSERT INTO quotations (
    id, org_id, quotation_number, customer_id, customer_name, rfq_id, rfq_number, status, 
    origin, origin_code, destination, destination_code, service_type, transport_mode, 
    currency, payment_terms, subtotal, surcharges, taxes, total_amount, total_cost, 
    gross_profit, gross_margin_pct, valid_from, valid_until, conversion_status, 
    converted_booking_id, created_by, created_at, updated_at
) VALUES
-- Quote 101: Accepted & converted quote
(101, 2, 'QT-2026-DEV-001', 101, 'Apex Global Logistics Corp', 101, 'RFQ-2026-DEV-001', 'ACCEPTED',
 'Nhava Sheva', 'INNSA', 'Rotterdam', 'NLRTM', 'PORT_TO_PORT', 'OCEAN_FCL',
 'USD', 'NET30', 2550.00, 300.00, 0.00, 2850.00, 2150.00,
 700.00, 24.56, '2026-08-01', '2026-09-30', 'CONVERTED',
 101, 'Operations Lead', NOW(), NOW()),

-- Quote 102: QUOTATION AWAITING REVIEW
(102, 2, 'QT-2026-DEV-002', 102, 'Nordic Freight Dynamics AB', 102, 'RFQ-2026-DEV-002', 'PENDING_APPROVAL',
 'Nhava Sheva', 'INNSA', 'Hamburg', 'DEHAM', 'PORT_TO_PORT', 'OCEAN_FCL',
 'USD', 'NET15', 2800.00, 350.00, 0.00, 3150.00, 2280.00,
 870.00, 27.62, '2026-09-01', '2026-10-31', 'NOT_CONVERTED',
 NULL, 'Operations Lead', NOW(), NOW())
ON DUPLICATE KEY UPDATE quotation_number = VALUES(quotation_number);

-- 11. Bookings
INSERT INTO bookings (
    id, org_id, rfq_id, quote_id, booking_number, carrier_name, carrier_scac, 
    carrier_booking_reference, carrier_booking_status, status, origin_port, 
    destination_port, vessel_name, voyage_number, etd, eta, cargo_summary, 
    created_by, created_at, updated_at
) VALUES
(101, 2, 101, 101, 'BK-2026-DEV-001', 'Maersk Line', 'MAEU',
 'MSK-BKG-889911', 'CONFIRMED', 'CONFIRMED', 'INNSA',
 'NLRTM', 'MAERSK MC-KINNEY MOLLER', '2601W', '2026-08-20 10:00:00', '2026-09-15 18:00:00', '1x40GP Automotive Engine Components',
 'Operations Desk', NOW(), NOW())
ON DUPLICATE KEY UPDATE booking_number = VALUES(booking_number);

-- 12. Shipments (including delayed ETA and critical exception)
INSERT INTO shipments (
    id, org_id, rfq_id, quote_id, booking_id, booking_number, mbl_number, hbl_number, 
    carrier_scac, vessel_name, voyage_number, origin_port, destination_port, 
    container_numbers, status, etd, eta, closure_status, created_at, updated_at
) VALUES
-- Shipment 101: In-transit normal workflow
(101, 2, 101, 101, 101, 'BK-2026-DEV-001', 'MAEU123456789', 'LH-HBL-2601',
 'MAEU', 'MAERSK MC-KINNEY MOLLER', '2601W', 'INNSA', 'NLRTM',
 '["MSKU7891234"]', 'IN_TRANSIT', '2026-08-20 10:00:00', '2026-09-15 18:00:00', 'ACTIVE', NOW(), NOW()),

-- Shipment 102: SHIPMENT WITH DELAYED ETA (planned: 2026-09-02, revised: 2026-09-12)
(102, 2, 102, 102, 101, 'BK-2026-DEV-002', 'MSCU987654321', 'LH-HBL-2602',
 'MSCU', 'MSC OSCAR', '2602W', 'INNSA', 'DEHAM',
 '["MSCU9876543"]', 'IN_TRANSIT', '2026-08-15 08:00:00', '2026-09-12 14:00:00', 'ACTIVE', NOW(), NOW()),

-- Shipment 103: SHIPMENT WITH CRITICAL EXCEPTION (Customs hold)
(103, 2, 101, 101, 101, 'BK-2026-DEV-003', 'CMDU543216789', 'LH-HBL-2603',
 'CMDU', 'CMA CGM ANTOINE', '2603W', 'INNSA', 'USNYC',
 '["CMAU5432109"]', 'CUSTOMS_HOLD', '2026-08-18 12:00:00', '2026-09-20 10:00:00', 'ACTIVE', NOW(), NOW())
ON DUPLICATE KEY UPDATE vessel_name = VALUES(vessel_name);

-- 13. Tracking Milestones for Shipments
INSERT INTO shipment_milestones (
    id, shipment_id, milestone_code, description, planned_date, actual_date, status, location, notes
) VALUES
(101, 101, 'GATE_IN', 'Container gated in at terminal', '2026-08-18 08:00:00', '2026-08-18 09:30:00', 'COMPLETED', 'Nhava Sheva Port, Terminal 3', 'Clean gate-in inspection passed'),
(102, 101, 'LOADED', 'Container loaded on vessel', '2026-08-19 14:00:00', '2026-08-19 15:20:00', 'COMPLETED', 'Nhava Sheva Port, Berth 2', 'Stowed under deck bay 14'),
(103, 101, 'DEPARTED', 'Vessel departed origin port', '2026-08-20 10:00:00', '2026-08-20 11:15:00', 'COMPLETED', 'Nhava Sheva Port', 'Vessel on schedule'),
(104, 101, 'ARRIVAL', 'Vessel scheduled arrival', '2026-09-15 18:00:00', NULL, 'PLANNED', 'Rotterdam Port (NLRTM)', 'Estimated arrival on track'),

(105, 102, 'GATE_IN', 'Container gated in at terminal', '2026-08-14 10:00:00', '2026-08-14 11:00:00', 'COMPLETED', 'Nhava Sheva Port', 'Initial gate in completed'),
(106, 102, 'DELAY_NOTICE', 'Vessel departure delayed due to typhoon', '2026-08-15 08:00:00', '2026-08-18 20:00:00', 'COMPLETED', 'Arabian Sea / Colombo Transshipment', 'Carrier bulletin #26-089: 10-day weather delay. Revised ETA: 2026-09-12')
ON DUPLICATE KEY UPDATE description = VALUES(description);

-- 14. Shipment Exceptions (including Critical Exception)
INSERT INTO shipment_exceptions (
    id, shipment_id, org_id, exception_type, severity, title, description, status, resolved, created_at, updated_at
) VALUES
(101, 103, 2, 'CUSTOMS_HOLD', 'CRITICAL', 
 'Customs Hold: Discrepancy in HS Code declarations', 
 'Container CMAU5432109 detained at transshipment inspection point. Commercial invoice states HS 8481.80 while bill of lading lists HS 8483.40. Immediate commercial resolution required to release container.',
 'OPEN', 0, NOW(), NOW()),

(102, 102, 2, 'ETA_DELAY', 'HIGH',
 'Vessel ETA Delayed by 10 Days',
 'Carrier MSC reported severe weather and transshipment port congestion. Estimated arrival pushed from 2026-09-02 to 2026-09-12.',
 'OPEN', 0, NOW(), NOW())
ON DUPLICATE KEY UPDATE title = VALUES(title);

-- 15. Contract Documents (including Contract requiring extraction & Anomaly document)
INSERT INTO contract_documents (
    id, org_id, carrier_scac, carrier_name, file_name, s3_key, file_type, file_size_bytes, 
    page_count, status, ai_document_summary, extracted_rate_count, confirmed_rate_count, 
    pending_review_count, failed_rate_count, created_by, created_at, updated_at
) VALUES
-- Doc 1: ONE CONTRACT REQUIRING EXTRACTION
('doc-dev-contract-001', 2, 'MAEU', 'Maersk Line', 'maersk_service_contract_2026.pdf', 
 'contracts/maersk_service_contract_2026.pdf', 'PDF', 184500, 6, 'PENDING_EXTRACTION', 
 NULL, 0, 0, 0, 0, 6, NOW(), NOW()),

-- Doc 2: ONE DOCUMENT WITH INTENTIONAL ANOMALY (e.g. rate spike / anomaly flag)
('doc-dev-contract-002', 2, 'MSCU', 'MSC Mediterranean Shipping Co', 'msc_rate_sheet_anomalous.pdf', 
 'contracts/msc_rate_sheet_anomalous.pdf', 'PDF', 92400, 3, 'PENDING_EXTRACTION', 
 NULL, 0, 0, 0, 0, 6, NOW(), NOW())
ON DUPLICATE KEY UPDATE file_name = VALUES(file_name);

-- 16. Shipment Documents
INSERT INTO shipment_documents (
    id, org_id, shipment_id, customer_id, booking_id, doc_type, document_name, 
    category, file_name, file_type, status, created_at, updated_at
) VALUES
(101, 2, 101, 101, 101, 'COMMERCIAL_INVOICE', 'Commercial Invoice - APEX-2601', 'COMMERCIAL', 'commercial_invoice_101.pdf', 'PDF', 'VERIFIED', NOW(), NOW()),
(102, 2, 101, 101, 101, 'BILL_OF_LADING', 'Master Bill of Lading MAEU123456789', 'TRANSPORT', 'bill_of_lading_101.pdf', 'PDF', 'VERIFIED', NOW(), NOW()),
(103, 2, 103, 101, 101, 'BILL_OF_LADING', 'Master Bill of Lading CMDU543216789', 'TRANSPORT', 'bill_of_lading_103.pdf', 'PDF', 'PENDING_REVIEW', NOW(), NOW())
ON DUPLICATE KEY UPDATE document_name = VALUES(document_name);

-- 17. Invoices (including Invoice requiring reconciliation & Overdue invoice)
INSERT INTO customer_invoices (
    id, org_id, invoice_number, customer_id, customer_name, customer_country, 
    shipment_id, shipment_number, booking_id, booking_number, quotation_id, quote_number, 
    route, origin, destination, invoice_date, due_date, days_left, currency, 
    subtotal, tax_amount, discount_amount, total_amount, paid_amount, balance_due, 
    status, type, bookmarked, is_my_invoice, creator_name, created_by_id, created_at, updated_at
) VALUES
-- Invoice 1: Closed paid invoice for Shipment 101
(101, 2, 'INV-2026-DEV-001', 101, 'Apex Global Logistics Corp', 'United States',
 101, 'SH-2026-DEV-001', 101, 'BK-2026-DEV-001', 101, 'QT-2026-DEV-001',
 'Nhava Sheva ➔ Rotterdam', 'INNSA', 'NLRTM', '2026-08-22', '2026-09-22', '16 days left', 'USD',
 3200.00, 0.00, 0.00, 3200.00, 3200.00, 0.00,
 'Paid', 'CUSTOMER_AR', 0, 1, 'Billing Operations', 6, NOW(), NOW()),

-- Invoice 2: INVOICE REQUIRING RECONCILIATION ($2,450 billed vs $2,150 contracted — $300 variance)
(102, 2, 'INV-2026-DEV-002', 102, 'Nordic Freight Dynamics AB', 'Sweden',
 102, 'SH-2026-DEV-002', 101, 'BK-2026-DEV-002', 102, 'QT-2026-DEV-002',
 'Nhava Sheva ➔ Hamburg', 'INNSA', 'DEHAM', '2026-08-25', '2026-09-25', '19 days left', 'USD',
 2450.00, 0.00, 0.00, 2450.00, 0.00, 2450.00,
 'Pending Approval', 'CUSTOMER_AR', 1, 1, 'Billing Operations', 6, NOW(), NOW()),

-- Invoice 3: OVERDUE INVOICE (Due 2026-08-15, 22 days overdue)
(103, 2, 'INV-2026-DEV-003', 102, 'Nordic Freight Dynamics AB', 'Sweden',
 102, 'SH-2026-DEV-002', 101, 'BK-2026-DEV-002', 102, 'QT-2026-DEV-002',
 'Nhava Sheva ➔ Hamburg', 'INNSA', 'DEHAM', '2026-07-15', '2026-08-15', '22 days overdue', 'USD',
 4500.00, 0.00, 0.00, 4500.00, 0.00, 4500.00,
 'Overdue', 'CUSTOMER_AR', 1, 1, 'Billing Operations', 6, NOW(), NOW())
ON DUPLICATE KEY UPDATE invoice_number = VALUES(invoice_number);

-- 18. Invoice Payment History
INSERT INTO customer_invoice_payments (
    id, org_id, invoice_id, payment_ref, amount, payment_method, status, payment_date, notes, created_at
) VALUES
(101, 2, 101, 'PAY-2026-DEV-001', 3200.00, 'Wire Transfer', 'Completed', '2026-08-24', 'Full settlement received via Swift MT103 from Citibank NY', NOW())
ON DUPLICATE KEY UPDATE payment_ref = VALUES(payment_ref);

-- 19. Approvals (including Human Approval required and High-Risk rule block)
INSERT INTO approval_requests (
    id, org_id, request_code, title, category, type, status, priority, 
    related_entity_type, related_entity_id, related_ref, customer_name, customer_id, 
    shipment_id, requested_by_id, requested_by_name, department, due_date, comments, created_at, updated_at
) VALUES
-- Record 1: ONE RECORD THAT REQUIRES HUMAN APPROVAL (Invoice price discrepancy resolution)
(101, 2, 'AP-2026-DEV-001', 'Vendor Bill Price Variance Approval ($300 Discrepancy)', 'FINANCE', 'Invoice Approval',
 'Pending', 'HIGH', 'INVOICE', 102, 'INV-2026-DEV-002', 'Nordic Freight Dynamics AB', 102,
 102, 6, 'Operations Lead', 'Finance & Billing', '2026-09-10 18:00:00',
 'Billed ocean freight ($2,450) exceeds agreed contract rate ($2,150) by $300. Human verification needed.', NOW(), NOW()),

-- Record 2: ONE RECORD THAT SHOULD BE REJECTED OR BLOCKED BY PERMISSION RULES
(102, 2, 'AP-2026-DEV-002', 'High-Risk Unsecured Credit Limit Override Request ($50,000)', 'CREDIT_LIMIT', 'Credit Limit Increase',
 'Pending', 'CRITICAL', 'CUSTOMER', 103, 'CUST-DEV-003', 'Bharat Tech Exports Pvt Ltd', 103,
 NULL, 6, 'Operations Lead', 'Risk & Compliance', '2026-09-08 12:00:00',
 'High-risk action: Exceeds default authorization matrix ($20,000 max). Policy prohibits approval without audited financials.', NOW(), NOW())
ON DUPLICATE KEY UPDATE title = VALUES(title);

SET FOREIGN_KEY_CHECKS = 1;
