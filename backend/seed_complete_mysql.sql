-- ==============================================================================
-- LOGISTICSHQ MASTER SEED DATA (Organizations, Users, Customers, RFQs, Quotes, Shipments)
-- ==============================================================================

SET FOREIGN_KEY_CHECKS = 0;

-- 1. Organizations
INSERT INTO organizations (id, name, created_at, updated_at)
VALUES (1, 'Freel Global Logistics Pvt Ltd', NOW(), NOW())
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- 2. Users
INSERT INTO users (id, cognito_sub, email, first_name, last_name, created_at, updated_at) VALUES
(1, 'local-ceo', 'ceo@freel-demo.local', 'Varun', 'Kanade', NOW(), NOW()),
(2, 'local-sales', 'sales@freel-demo.local', 'Priya', 'Sharma', NOW(), NOW()),
(3, 'local-pricing', 'pricing@freel-demo.local', 'Aditya', 'Kumar', NOW(), NOW()),
(4, 'local-customer', 'customer@tata-exports.local', 'Ravi', 'Mehta', NOW(), NOW()),
(5, 'user-varun-direct', 'varunkanade3456@gmail.com', 'Varun', 'Kanade', NOW(), NOW())
ON DUPLICATE KEY UPDATE first_name = VALUES(first_name), last_name = VALUES(last_name);

-- 3. Roles
INSERT INTO roles (id, org_id, name, description, created_at, updated_at) VALUES
(1, 1, 'CEO', 'Chief Executive Officer / Super Admin', NOW(), NOW()),
(2, 1, 'ADMIN', 'Administrator', NOW(), NOW()),
(3, 1, 'SALES', 'Sales Representative', NOW(), NOW()),
(4, 1, 'PRICING', 'Pricing & Procurement Specialist', NOW(), NOW()),
(5, 1, 'OPERATIONS', 'Operations & Freight Forwarding Manager', NOW(), NOW()),
(6, 1, 'CUSTOMER_CONTACT', 'External Client Contact', NOW(), NOW())
ON DUPLICATE KEY UPDATE description = VALUES(description);

-- 4. Org Members
INSERT INTO org_members (id, org_id, user_id, role_id, status, created_at, updated_at) VALUES
(1, 1, 1, 1, 'ACTIVE', NOW(), NOW()),
(2, 1, 2, 3, 'ACTIVE', NOW(), NOW()),
(3, 1, 3, 4, 'ACTIVE', NOW(), NOW()),
(4, 1, 4, 6, 'ACTIVE', NOW(), NOW()),
(5, 1, 5, 2, 'ACTIVE', NOW(), NOW())
ON DUPLICATE KEY UPDATE status = 'ACTIVE';

-- 5. Addresses & Companies
INSERT INTO addresses (id, org_id, address_line_1, city, state, postal_code, country_code) VALUES
(1, 1, 'Plot 42, MIDC Industrial Area', 'Mumbai', 'Maharashtra', '400093', 'IN'),
(2, 1, 'GIDC Estate, Phase II', 'Vadodara', 'Gujarat', '390010', 'IN'),
(3, 1, 'Sector 18, Electronic City', 'Gurgaon', 'Haryana', '122015', 'IN')
ON DUPLICATE KEY UPDATE city = VALUES(city);

INSERT INTO companies (id, org_id, name, domain, industry, address_id) VALUES
(1, 1, 'Tata Exports Ltd', 'tataexports.com', 'Automotive & Heavy Manufacturing', 1),
(2, 1, 'Sun Pharma International', 'sunpharma.com', 'Pharmaceuticals & Life Sciences', 2),
(3, 1, 'Reliance Petrochem Global', 'reliancepetro.com', 'Chemicals & Energy', 3),
(4, 1, 'Mahindra Auto Logistics', 'mahindra.com', 'Automotive', 1),
(5, 1, 'Adani Ports & SEZ', 'adani.com', 'Port Infrastructure & Trade', 1)
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- 6. Contacts
INSERT INTO contacts (id, org_id, company_id, first_name, last_name, email, phone, job_title) VALUES
(1, 1, 1, 'Rajesh', 'Verma', 'r.verma@tataexports.com', '+91 98200 12345', 'VP Supply Chain'),
(2, 1, 2, 'Sneha', 'Patel', 's.patel@sunpharma.com', '+91 98330 54321', 'Head of Logistics'),
(3, 1, 3, 'Amit', 'Deshmukh', 'a.deshmukh@reliancepetro.com', '+91 98111 99887', 'Procurement Director')
ON DUPLICATE KEY UPDATE first_name = VALUES(first_name);

-- 7. Customers
INSERT INTO customers (id, org_id, company_id, status) VALUES
(1, 1, 1, 'ACTIVE'),
(2, 1, 2, 'ACTIVE'),
(3, 1, 3, 'ACTIVE'),
(4, 1, 4, 'ACTIVE'),
(5, 1, 5, 'ACTIVE')
ON DUPLICATE KEY UPDATE status = VALUES(status);

-- 8. Leads
INSERT INTO leads (id, org_id, company_name, contact_name, email, phone, source, status, ai_score, ai_research_report) VALUES
(1, 1, 'Bharat Forge High-Tech', 'Karan Singhal', 'karan@bharatforge.local', '+91 99001 11223', 'INBOUND_EMAIL', 'QUALIFIED', 88, 'Strong export volume to EU and North America. High intent for monthly 40HQ reefer containers.'),
(2, 1, 'Lupin Generics Export', 'Ananya Roy', 'ananya.roy@lupin.local', '+91 99002 22334', 'WEBSITE', 'IN_CONVERSATION', 92, 'Requires temperature-controlled air freight capacity to Frankfurt (FRA) weekly.'),
(3, 1, 'Havells Industrial Cable', 'Vikram Seth', 'v.seth@havells.local', '+91 99003 33445', 'REFERRAL', 'NEW', 65, 'RFQ requested for Nhava Sheva to Jebel Ali, ocean FCL 2x40GP.')
ON DUPLICATE KEY UPDATE status = VALUES(status), ai_score = VALUES(ai_score);

-- 9. Carriers
INSERT INTO carriers (scac, name, api_endpoint, is_active) VALUES
('MAEU', 'Maersk Line', 'https://api.maersk.com', 1),
('MSCU', 'MSC Mediterranean Shipping Co', 'https://api.msc.com', 1),
('CMDU', 'CMA CGM Group', 'https://api.cma-cgm.com', 1),
('HLCU', 'Hapag-Lloyd AG', 'https://api.hapag-lloyd.com', 1),
('COSU', 'COSCO Shipping', 'https://api.coscoshipping.com', 1),
('ONEU', 'Ocean Network Express (ONE)', 'https://api.one-line.com', 1)
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- 10. RFQs
INSERT INTO rfqs (id, org_id, rfq_number, customer_id, stage, status, agent_status, origin, destination, incoterms, target_date) VALUES
(1, 1, 'RFQ-2026-1001', 1, 'WON', 'CONFIRMED', 'IDLE', 'INNSA (Nhava Sheva)', 'DEHAM (Hamburg)', 'FOB', DATE_ADD(NOW(), INTERVAL 30 DAY)),
(2, 1, 'RFQ-2026-1002', 2, 'QUOTE_SENT', 'ACTIVE', 'IDLE', 'INBOM (Mumbai Air)', 'FRA (Frankfurt Air)', 'CIP', DATE_ADD(NOW(), INTERVAL 14 DAY)),
(3, 1, 'RFQ-2026-1003', 3, 'QUOTE_GENERATED', 'DRAFT', 'COMPLETED', 'INNSA (Nhava Sheva)', 'USNYC (New York)', 'CIF', DATE_ADD(NOW(), INTERVAL 45 DAY)),
(4, 1, 'RFQ-2026-1004', 1, 'PRICING_ASSIGNED', 'IN_REVIEW', 'RUNNING', 'INCCU (Kolkata)', 'SGSIN (Singapore)', 'FOB', DATE_ADD(NOW(), INTERVAL 25 DAY)),
(5, 1, 'RFQ-2026-1005', 4, 'RFQ_CREATED', 'NEW', 'QUEUED', 'INBLR (Bangalore)', 'AEAUH (Abu Dhabi)', 'EXW', DATE_ADD(NOW(), INTERVAL 20 DAY)),
(6, 1, 'RFQ-2026-1006', 5, 'LOST', 'CLOSED', 'IDLE', 'INMUN (Mundra)', 'CNSHA (Shanghai)', 'FOB', DATE_SUB(NOW(), INTERVAL 10 DAY))
ON DUPLICATE KEY UPDATE stage = VALUES(stage), status = VALUES(status);

-- 11. RFQ Items
INSERT INTO rfq_items (id, rfq_id, description, quantity, weight_kg, volume_cbm) VALUES
(1, 1, 'Precision CNC Automotive Components in Wooden Crates', 10, 18500.00, 32.00),
(2, 2, 'Temperature-Controlled Active Pharmaceutical Ingredients', 4, 3200.00, 12.50),
(3, 3, 'Specialty Polymers & Synthetic Resin (20ft Dry)', 2, 24000.00, 58.00),
(4, 4, 'Industrial Steel Fasteners & Raw Castings', 6, 21000.00, 28.00),
(5, 5, 'Heavy Electrical Switchgear Panels', 3, 8500.00, 19.20),
(6, 6, 'Ceramic Floor Tiles & Porcellanato Slabs', 5, 27500.00, 44.00)
ON DUPLICATE KEY UPDATE description = VALUES(description);

-- 12. RFQ Quotes
INSERT INTO rfq_quotes (id, rfq_id, carrier_id, carrier_name, quote_reference, currency, buy_price, sell_price, ocean_freight, origin_charges, destination_charges, total_buy_price, free_days, transit_time_days, is_recommended, reliability_score, historical_success_rate, ai_reasoning, status) VALUES
(1, 1, 'MAEU', 'Maersk Line', 'MSK-QT-8891', 'USD', 2150.00, 2580.00, 1900.00, 150.00, 100.00, 2150.00, 14, 24, 1, 96.50, 98.20, 'Maersk offers guaranteed equipment release at Nhava Sheva with direct service to Hamburg in 24 days. Optimal margin profile.', 'APPROVED'),
(2, 1, 'MSCU', 'MSC Mediterranean Shipping Co', 'MSC-QT-4412', 'USD', 1980.00, 2376.00, 1750.00, 130.00, 100.00, 1980.00, 7, 28, 0, 89.00, 91.50, 'Lower spot rate, but transit time is 4 days longer with transshipment at Colombo.', 'DRAFT'),
(3, 2, 'HLCU', 'Hapag-Lloyd AG', 'HL-QT-9920', 'USD', 3400.00, 4080.00, 3100.00, 180.00, 120.00, 3400.00, 10, 7, 1, 94.00, 96.00, 'Validated temperature-controlled pharma charter lane. Meets GDP compliance criteria.', 'APPROVED'),
(4, 3, 'CMDU', 'CMA CGM Group', 'CMA-QT-1104', 'USD', 2850.00, 3420.00, 2500.00, 200.00, 150.00, 2850.00, 14, 26, 1, 92.00, 94.50, 'AI Auto-Quote generated: Recommended based on 14 free days at NY/NJ and 26 day direct transit.', 'DRAFT')
ON DUPLICATE KEY UPDATE buy_price = VALUES(buy_price), sell_price = VALUES(sell_price), status = VALUES(status);

-- 13. Shipments
INSERT INTO shipments (id, org_id, rfq_id, quote_id, booking_number, mbl_number, hbl_number, carrier_scac, vessel_name, voyage_number, origin_port, destination_port, status, etd, eta) VALUES
(101, 1, 1, 1, 'BKG-MAEU-2026-001', 'MAEU988776655', 'HBL-INNSA-0441', 'MAEU', 'MAERSK MC-KINNEY MOLLER', '2604W', 'INNSA', 'DEHAM', 'IN_TRANSIT', DATE_SUB(NOW(), INTERVAL 5 DAY), DATE_ADD(NOW(), INTERVAL 19 DAY)),
(102, 1, 2, 3, 'BKG-HLCU-2026-089', 'HLCU112233445', 'HBL-INBOM-0912', 'HLCU', 'AL DAHNA EXPRESS', '092E', 'INBOM', 'FRA', 'BOOKED', DATE_ADD(NOW(), INTERVAL 3 DAY), DATE_ADD(NOW(), INTERVAL 10 DAY)),
(103, 1, 1, 1, 'BKG-MSCU-2026-302', 'MSCU556677889', 'HBL-INNSA-0102', 'MSCU', 'MSC OSCAR', '2611N', 'INNSA', 'USNYC', 'DELIVERED', DATE_SUB(NOW(), INTERVAL 35 DAY), DATE_SUB(NOW(), INTERVAL 3 DAY))
ON DUPLICATE KEY UPDATE status = VALUES(status), vessel_name = VALUES(vessel_name);

-- 14. Rates & Rate Entries
INSERT INTO rates (id, org_id, rate_reference, carrier_name, carrier_code, rate_type, transport_mode, service_type, equipment_type, origin_port, destination_port, currency, base_amount, effective_date, expiry_date, status) VALUES
(1, 1, 'RATE-2026-MAEU-EU', 'Maersk Line', 'MAEU', 'CONTRACT', 'Ocean FCL', 'FCL', '40GP', 'INNSA', 'DEHAM', 'USD', 2100.00, '2026-01-01', '2026-12-31', 'ACTIVE'),
(2, 1, 'RATE-2026-MSCU-US', 'MSC Mediterranean Shipping Co', 'MSCU', 'CONTRACT', 'Ocean FCL', 'FCL', '40HC', 'INNSA', 'USNYC', 'USD', 2850.00, '2026-01-01', '2026-12-31', 'ACTIVE'),
(3, 1, 'RATE-2026-CMDU-ME', 'CMA CGM Group', 'CMDU', 'SPOT', 'Ocean FCL', 'FCL', '20GP', 'INNSA', 'AEJEA', 'USD', 850.00, '2026-08-01', '2026-10-31', 'ACTIVE')
ON DUPLICATE KEY UPDATE base_amount = VALUES(base_amount), status = VALUES(status);

SET FOREIGN_KEY_CHECKS = 1;
