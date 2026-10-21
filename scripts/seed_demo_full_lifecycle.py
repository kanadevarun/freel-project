import pymysql
import json
from datetime import datetime, timedelta

conn = pymysql.connect(host='127.0.0.1', port=3306, user='root', password='', db='freel_mysql', autocommit=True)
cur = conn.cursor(pymysql.cursors.DictCursor)

print("=================================================================")
print("1. CLEANING UP & RENAMING CUSTOMERS TO AUTHENTIC EXPORT SHIPPERS")
print("=================================================================")

# Map awkward / test names to real manufacturing & export shippers
exporter_names = {
    1: ('Nordic Precision Engineering Exporters AB', 'SE', 'Gothenburg, Sweden'),
    2: ('Solaria CleanTech Solar Exporters Ltd', 'CN', 'Ningbo, China'),
    3: ('Bavaria Chemical & Specialty Resins SE', 'DE', 'Hamburg, Germany'),
    4: ('Indo-Pacific Agro Commodities Exporters Ltd', 'IN', 'Mumbai, India'),
    5: ('Apex Global Precision Tooling Exporters Corp', 'US', 'Chicago, United States'),
    101: ('Apex Global Precision Tooling Exporters Corp', 'US', 'Chicago, United States'),
    102: ('Nordic Precision Engineering Exporters AB', 'SE', 'Gothenburg, Sweden'),
    103: ('Bharat Industrial Machinery Exporters Pvt Ltd', 'IN', 'Chennai, India'),
    104: ('Hindustan Polymer Resins Exporters Ltd', 'IN', 'Mundra, India'),
    105: ('Euro-Asia Electronics Manufacturing Exporters Ltd', 'NL', 'Rotterdam, Netherlands'),
    9395: ('ScanSteel Heavy Fabrication Exporters AB', 'SE', 'Gothenburg, Sweden'),
    9396: ('Rhine-Ruhr Industrial Pump Exporters GmbH', 'DE', 'Duisburg, Germany'),
    9397: ('Baltic Timber & Modular Housing Exporters AS', 'EE', 'Tallinn, Estonia'),
    9398: ('Alpine Precision Hydraulics Exporters AG', 'CH', 'Basel, Switzerland'),
    9399: ('American Heavy Industrial Machinery Exporters Inc', 'US', 'Houston, United States'),
    9400: ('Nordic Sustainable Wood & Cellulose Exporters AB', 'SE', 'Gothenburg, Sweden')
}

for cid, (cname, ccountry, ccity) in exporter_names.items():
    cur.execute("""
        UPDATE customers 
        SET name = %s, country = %s, city = %s 
        WHERE id = %s
    """, (cname, ccountry, ccity, cid))

# Also ensure any customer containing "Audit Shipper" or "Test Shipper" is renamed
cur.execute("""
    UPDATE customers 
    SET name = 'Nordic Maritime Component Exporters AB', country = 'SE' 
    WHERE name LIKE '%Audit Shipper%' OR name LIKE '%Test Shipper%'
""")

print("Customers cleaned successfully.")

print("\n=================================================================")
print("2. CONNECTING CMA CGM CARRIER INTEGRATION IN CARRIER SETTINGS")
print("=================================================================")

cur.execute("""
    UPDATE carrier_integrations 
    SET connection_status = 'CONNECTED',
        is_active = 1,
        sync_status = 'Completed',
        last_synced_at = NOW(),
        last_success_at = NOW(),
        credential_mask = '{"client_id": "cma-cgm-api-freel-prod-9812", "status": "Active & Certified"}',
        capabilities = '["TRACKING", "RATES", "CONTRACT_RATES", "SPOT_RATES", "BOOKING", "DOCUMENTS"]'
    WHERE carrier_scac = 'CMDU' AND org_id = 1
""")
if cur.rowcount == 0:
    cur.execute("""
        INSERT INTO carrier_integrations (
            org_id, carrier_provider_id, carrier_scac, carrier_name, connection_method,
            environment, connection_status, is_active, credential_mask, capabilities,
            sync_status, last_synced_at, last_success_at, created_at, updated_at
        ) VALUES (
            1, 4, 'CMDU', 'CMA CGM Group', 'API',
            'PRODUCTION', 'CONNECTED', 1, '{"client_id": "cma-cgm-api-freel-prod-9812", "status": "Active & Certified"}',
            '["TRACKING", "RATES", "CONTRACT_RATES", "SPOT_RATES", "BOOKING", "DOCUMENTS"]',
            'Completed', NOW(), NOW(), NOW(), NOW()
        )
    """)
print("CMA CGM Group integration set to CONNECTED.")

print("\n=================================================================")
print("3. SEEDING GOLDEN PATH: LEAD (INBOUND EMAIL)")
print("=================================================================")

cur.execute("DELETE FROM leads WHERE email = 'astrid.lindqvist@nordic-precision-exports.se'")
cur.execute("""
    INSERT INTO leads (
        org_id, company_name, contact_name, email, phone, source,
        status, ai_score, notes, location, created_at, updated_at
    ) VALUES (
        1, 'Nordic Precision Engineering Exporters AB', 'Astrid Lindqvist',
        'astrid.lindqvist@nordic-precision-exports.se', '+46 31 712 4490',
        'EMAIL_INBOUND', 'CONVERTED', 95,
        'Export Request: 2x 40ft High Cube containers of industrial CNC machinery & robotics from Gothenburg (SEGOT) to Singapore (SGSIN). Preferred ocean carrier: CMA CGM Group.',
        'Gothenburg, Sweden',
        NOW() - INTERVAL 12 DAY, NOW() - INTERVAL 10 DAY
    )
""")
lead_id = cur.lastrowid
print(f"Created Lead ID {lead_id}: Nordic Precision Engineering Exporters AB (Contact: Astrid Lindqvist)")

print("\n=================================================================")
print("4. SEEDING GOLDEN PATH: RFQ (RFQ-2026-9810)")
print("=================================================================")

cur.execute("DELETE FROM rfq_items WHERE rfq_id IN (SELECT id FROM rfqs WHERE rfq_number = 'RFQ-2026-9810')")
cur.execute("DELETE FROM rfqs WHERE rfq_number = 'RFQ-2026-9810'")

cur.execute("""
    INSERT INTO rfqs (
        org_id, rfq_number, customer_id, lead_id, origin, destination,
        incoterms, stage, status, health_score, target_date, created_at, updated_at
    ) VALUES (
        1, 'RFQ-2026-9810', 1, %s, 'SEGOT - Gothenburg Port, Sweden',
        'SGSIN - Port of Singapore', 'FOB', 'STAGE_QUOTE_GENERATED', 'ACTIVE', 98,
        NOW() + INTERVAL 15 DAY, NOW() - INTERVAL 10 DAY, NOW() - INTERVAL 8 DAY
    )
""", (lead_id,))
rfq_id = cur.lastrowid

cur.execute("""
    INSERT INTO rfq_items (
        rfq_id, description, quantity, weight_kg, volume_cbm, created_at, updated_at
    ) VALUES (
        %s, '2x 40ft High Cube Containers - Automated CNC Milling Centers & Control Spares',
        2, 38400.00, 152.00, NOW() - INTERVAL 10 DAY, NOW() - INTERVAL 10 DAY
    )
""", (rfq_id,))
print(f"Created RFQ ID {rfq_id}: RFQ-2026-9810 for Nordic Precision Engineering Exporters AB")

print("\n=================================================================")
print("5. SEEDING GOLDEN PATH: QUOTATION (QT-2026-9810) WITH CMA CGM")
print("=================================================================")

cur.execute("DELETE FROM quotations WHERE quotation_number = 'QT-2026-9810'")
cur.execute("""
    INSERT INTO quotations (
        org_id, quotation_number, customer_id, customer_name, rfq_id, rfq_number,
        status, origin, origin_code, destination, destination_code, service_type,
        transport_mode, currency, payment_terms, subtotal, surcharges, taxes,
        total_cost, total_amount, gross_profit, gross_margin_pct,
        valid_from, valid_until, sent_at, sent_by, viewed_at, accepted_at,
        commercial_terms, customer_notes, internal_notes, created_by, created_at, updated_at
    ) VALUES (
        1, 'QT-2026-9810', 1, 'Nordic Precision Engineering Exporters AB',
        %s, 'RFQ-2026-9810', 'ACCEPTED',
        'Gothenburg Port, Sweden', 'SEGOT', 'Port of Singapore', 'SGSIN',
        'Direct Ocean FCL (2x 40ft HC)', 'OCEAN_FCL', 'USD', 'Net 30 Days',
        7720.00, 880.00, 0.00,
        7720.00, 8600.00, 880.00, 10.23,
        CURDATE() - INTERVAL 7 DAY, CURDATE() + INTERVAL 21 DAY,
        NOW() - INTERVAL 6 DAY, 'Sales Desk (sales@freellogistics.com)',
        NOW() - INTERVAL 5 DAY, NOW() - INTERVAL 4 DAY,
        'Carrier Contract: CMA CGM Falcon Express Direct. Includes BAF, Origin THC, and Export Customs Clearance. 14 Days Free Time at Destination.',
        'High-value robotics machinery requiring dry cargo container inspection and under-deck stowage.',
        'CMA CGM allocation locked in under Q4 Freight Forwarder Master Service Agreement.',
        'Sarah Jenkins (Senior Freight Executive)',
        NOW() - INTERVAL 8 DAY, NOW() - INTERVAL 4 DAY
    )
""", (rfq_id,))
quotation_id = cur.lastrowid
print(f"Created Quotation ID {quotation_id}: QT-2026-9810 ($8,600.00 USD - CMA CGM Group)")

print("\n=================================================================")
print("6. SEEDING GOLDEN PATH: BOOKING (BKG-2026-9810) WITH CMA CGM")
print("=================================================================")

cur.execute("DELETE FROM bookings WHERE booking_number = 'BKG-2026-9810'")
cur.execute("""
    INSERT INTO bookings (
        org_id, rfq_id, quote_id, booking_number, carrier_id, carrier_name, carrier_scac,
        carrier_booking_reference, carrier_booking_status, carrier_confirmation_reference,
        carrier_booked_at, status, origin_port, destination_port, vessel_name, voyage_number,
        etd, eta, cargo_summary, special_instructions, created_by, created_at, updated_at,
        source_quotation_id, source_quote_number, commercial_handover_status
    ) VALUES (
        1, %s, %s, 'BKG-2026-9810', '4', 'CMA CGM Group', 'CMDU',
        'CMA-9810442-GOT', 'CONFIRMED', 'CMA-CONF-88219',
        NOW() - INTERVAL 4 DAY, 'CONFIRMED',
        'Gothenburg (SEGOT)', 'Singapore (SGSIN)',
        'CMA CGM Antoine de Saint Exupery', '082WFE',
        NOW() + INTERVAL 1 DAY, NOW() + INTERVAL 24 DAY,
        '2x 40ft High Cube Containers - Automated CNC Milling Centers & Control Spares (38,400 kg)',
        'Carrier instructed for under-deck stowage. Seal verification required upon terminal gate-in.',
        'Operations Desk (ops@freellogistics.com)',
        NOW() - INTERVAL 4 DAY, NOW() - INTERVAL 4 DAY,
        %s, 'QT-2026-9810', 'COMPLETED'
    )
""", (rfq_id, quotation_id, quotation_id))
booking_id = cur.lastrowid
print(f"Created Booking ID {booking_id}: BKG-2026-9810 (CMA CGM Antoine de Saint Exupery, Ref CMA-9810442-GOT)")

print("\n=================================================================")
print("7. SEEDING GOLDEN PATH: SHIPMENT (SH-2026-9810) & MILESTONES")
print("=================================================================")

cur.execute("DELETE FROM shipment_milestones WHERE shipment_id IN (SELECT id FROM shipments WHERE booking_number = 'BKG-2026-9810')")
cur.execute("DELETE FROM shipments WHERE booking_number = 'BKG-2026-9810'")

cur.execute("""
    INSERT INTO shipments (
        org_id, rfq_id, quote_id, booking_id, booking_number, mbl_number, hbl_number,
        carrier_scac, vessel_name, voyage_number, origin_port, destination_port,
        container_numbers, status, etd, eta, customer_commitment_date,
        current_risk_level, adaptive_status, created_at, updated_at,
        source_quotation_id, source_booking_id, closure_status
    ) VALUES (
        1, %s, %s, %s, 'BKG-2026-9810',
        'CMDU981044200', 'FRLL-GOT-SIN-9810', 'CMDU',
        'CMA CGM Antoine de Saint Exupery', '082WFE',
        'SEGOT', 'SGSIN',
        '[\"CMAU8219401\", \"CMAU8219402\"]',
        'IN_TRANSIT',
        NOW() - INTERVAL 1 DAY, NOW() + INTERVAL 22 DAY, NOW() + INTERVAL 24 DAY,
        'LOW', 'NORMAL',
        NOW() - INTERVAL 3 DAY, NOW() - INTERVAL 1 HOUR,
        %s, %s, 'ACTIVE'
    )
""", (rfq_id, quotation_id, booking_id, quotation_id, booking_id))
shipment_id = cur.lastrowid

now = datetime.now()
milestones = [
    ('GATE_IN', 'Containers gated-in at APM Terminals Gothenburg', now - timedelta(days=3), now - timedelta(days=3), 'COMPLETED', 'Gothenburg APM Terminal'),
    ('CUSTOMS_CLEAR', 'Export customs clearance validated by Swedish Tullverket', now - timedelta(days=2), now - timedelta(days=2), 'COMPLETED', 'Gothenburg Tullverket'),
    ('LOADED', 'Loaded aboard CMA CGM Antoine de Saint Exupery (Bay 24/Tier 04)', now - timedelta(days=1), now - timedelta(days=1), 'COMPLETED', 'Gothenburg Quay Berth 6'),
    ('DEPARTED', 'Vessel unberthed and departed Gothenburg Port - Sea transit commenced', now - timedelta(hours=18), now - timedelta(hours=18), 'COMPLETED', 'Kattegat Channel'),
    ('IN_TRANSIT', 'Sea voyage en route to Singapore via English Channel and Suez', now + timedelta(days=5), None, 'PLANNED', 'English Channel (49.32° N, 3.84° W)'),
    ('ARRIVAL', 'Scheduled arrival PSA Singapore Container Terminal', now + timedelta(days=22), None, 'PLANNED', 'PSA Singapore Container Terminal'),
    ('DELIVERED', 'Consignee final delivery in Singapore industrial free zone', now + timedelta(days=24), None, 'PLANNED', 'Jurong Logistics Hub, Singapore')
]

for code, desc, planned, actual, stat, loc in milestones:
    cur.execute("""
        INSERT INTO shipment_milestones (
            shipment_id, milestone_code, description, planned_date, actual_date, status, location, notes, updated_at
        ) VALUES (
            %s, %s, %s, %s, %s, %s, %s, 'Validated via CMA CGM EDI-315 automated status feed', NOW()
        )
    """, (shipment_id, code, desc, planned, actual, stat, loc))

print(f"Created Shipment ID {shipment_id}: SH-2026-9810 (MBL: CMDU981044200, Containers: CMAU8219401, CMAU8219402)")

print("\n=================================================================")
print("8. SEEDING GOLDEN PATH: COMMERCIAL INVOICE (INV-2026-9810)")
print("=================================================================")

cur.execute("DELETE FROM customer_invoice_documents WHERE invoice_id IN (SELECT id FROM customer_invoices WHERE invoice_number = 'INV-2026-9810')")
cur.execute("DELETE FROM customer_invoice_history WHERE invoice_id IN (SELECT id FROM customer_invoices WHERE invoice_number = 'INV-2026-9810')")
cur.execute("DELETE FROM customer_invoice_payments WHERE invoice_id IN (SELECT id FROM customer_invoices WHERE invoice_number = 'INV-2026-9810')")
cur.execute("DELETE FROM customer_invoice_items WHERE invoice_id IN (SELECT id FROM customer_invoices WHERE invoice_number = 'INV-2026-9810')")
cur.execute("DELETE FROM customer_invoices WHERE invoice_number = 'INV-2026-9810'")

cur.execute("""
    INSERT INTO customer_invoices (
        org_id, invoice_number, customer_id, customer_name, customer_country,
        shipment_id, shipment_number, booking_id, booking_number, quotation_id, quote_number,
        route, origin, destination, invoice_date, due_date, days_left,
        currency, subtotal, tax_amount, discount_amount, total_amount, paid_amount, balance_due,
        status, type, bookmarked, is_my_invoice, creator_name, created_at, updated_at
    ) VALUES (
        1, 'INV-2026-9810', 1, 'Nordic Precision Engineering Exporters AB', 'SE',
        %s, 'SH-2026-9810', %s, 'BKG-2026-9810', %s, 'QT-2026-9810',
        'Gothenburg (SEGOT) ➔ Singapore (SGSIN)', 'Gothenburg Port (SEGOT)', 'Singapore Port (SGSIN)',
        CURDATE() - INTERVAL 2 DAY, CURDATE() + INTERVAL 28 DAY, '28 days left',
        'USD', 8600.00, 0.00, 0.00, 8600.00, 0.00, 8600.00,
        'Issued', 'CUSTOMER_AR', 1, 1, 'By Billing Desk', NOW() - INTERVAL 2 DAY, NOW()
    )
""", (shipment_id, booking_id, quotation_id))
invoice_id = cur.lastrowid

# Line items
inv_items = [
    ('CMA CGM Ocean Freight (2x 40ft High Cube Dry Containers)', 'Ocean Freight', 2, 3200.00, 6400.00, 1),
    ('CMA CGM Low Sulphur Fuel & BAF Adjustment Factor', 'Bunker / Fuel', 2, 350.00, 700.00, 2),
    ('Origin Terminal Handling Charges (THC) - Gothenburg APM', 'Terminal Handling', 2, 220.00, 440.00, 3),
    ('Swedish Customs Export Regulatory Filing & Compliance', 'Customs Clearance', 1, 180.00, 180.00, 4),
    ('Freel Global Forwarding Documentation & Sea Waybill Issue', 'Freight Forwarding', 1, 880.00, 880.00, 5)
]
for desc, cat, qty, rate, amt, order in inv_items:
    cur.execute("""
        INSERT INTO customer_invoice_items (
            org_id, invoice_id, description, service_category, quantity, unit_price, total_amount, display_order, created_at
        ) VALUES (
            1, %s, %s, %s, %s, %s, %s, %s, NOW()
        )
    """, (invoice_id, desc, cat, qty, rate, amt, order))

# Documents
docs = [
    ('CMA_CGM_Master_Bill_of_Lading_CMDU981044200.pdf', '248 KB', 'application/pdf'),
    ('Nordic_Precision_Commercial_Invoice_Packing_List.pdf', '412 KB', 'application/pdf'),
    ('Swedish_Customs_Export_Declaration_Release.pdf', '185 KB', 'application/pdf')
]
for dname, dsize, dtype in docs:
    cur.execute("""
        INSERT INTO customer_invoice_documents (
            org_id, invoice_id, document_name, file_size, file_type, s3_key, uploaded_at
        ) VALUES (
            1, %s, %s, %s, %s, 'documents/invoices/9810', NOW() - INTERVAL 2 DAY
        )
    """, (invoice_id, dname, dsize, dtype))

# History
hist = [
    ('Invoice Generated', 'Commercial freight invoice automatically generated from accepted quotation QT-2026-9810', 'System Automation'),
    ('Carrier Booking Reconciled', 'CMA CGM booking ref CMA-9810442-GOT and MBL verified by operations', 'Marcus Vance (Senior Ops)'),
    ('Approved & Issued', 'Invoice approved by finance director and issued to Nordic Precision accounts', 'David Kim (Finance Controller)')
]
for title, desc, user in hist:
    cur.execute("""
        INSERT INTO customer_invoice_history (
            org_id, invoice_id, title, description, user_name, created_at
        ) VALUES (
            1, %s, %s, %s, %s, NOW() - INTERVAL 2 DAY
        )
    """, (invoice_id, title, desc, user))

print(f"Created Customer Invoice ID {invoice_id}: INV-2026-9810 ($8,600.00 USD, 5 itemized lines)")

print("\n=================================================================")
print("9. SEEDING GOLDEN PATH: DEBIT NOTE (DN-20261020-0001) - CMA CGM")
print("=================================================================")

cur.execute("DELETE FROM debit_note_items WHERE debit_note_id IN (SELECT id FROM debit_notes WHERE debit_note_number = 'DN-20261020-0001')")
cur.execute("DELETE FROM debit_notes WHERE debit_note_number = 'DN-20261020-0001'")

cur.execute("""
    INSERT INTO debit_notes (
        org_id, debit_note_number, customer_id, customer_name, customer_country,
        shipment_id, shipment_number, invoice_id, invoice_number,
        reason, currency, subtotal, tax_amount, total_amount, status,
        issue_date, due_date, notes, created_by_id, created_at, updated_at
    ) VALUES (
        1, 'DN-20261020-0001', 1, 'Nordic Precision Engineering Exporters AB', 'SE',
        %s, 'SH-2026-9810', %s, 'INV-2026-9810',
        'CMA CGM extended terminal quay storage & pre-carriage detention at Gothenburg terminal due to specialized heavy-machinery export inspection',
        'USD', 650.00, 0.00, 650.00, 'ISSUED',
        CURDATE() - INTERVAL 1 DAY, CURDATE() + INTERVAL 14 DAY,
        'CMA CGM equipment detention advice ref: CMA-DET-GOT-9810. Approved by customer logistics representative.',
        1, NOW() - INTERVAL 1 DAY, NOW()
    )
""", (shipment_id, invoice_id))
debit_note_id = cur.lastrowid

dn_items = [
    ('CMA CGM Container Terminal Quay Storage (2 Days beyond tariff free-time)', 'Demurrage / Storage', 2, 225.00, 450.00, 1),
    ('Gothenburg Specialized Heavy-Lift Terminal Crane Ancillary Surcharge', 'Terminal Ancillary', 1, 200.00, 200.00, 2)
]
for desc, cat, qty, rate, amt, order in dn_items:
    cur.execute("""
        INSERT INTO debit_note_items (
            org_id, debit_note_id, description, service_category, quantity, unit_price, total_amount, display_order, created_at
        ) VALUES (
            1, %s, %s, %s, %s, %s, %s, %s, NOW()
        )
    """, (debit_note_id, desc, cat, qty, rate, amt, order))

print(f"Created Debit Note ID {debit_note_id}: DN-20261020-0001 ($650.00 USD - CMA CGM Storage & Ancillaries)")

print("\n=================================================================")
print("10. ALSO REFINING ALL EXISTING INVOICES & DEBIT NOTES NAMES")
print("=================================================================")

# Update existing invoices and debit notes to clean shipper names
cur.execute("""
    UPDATE customer_invoices 
    SET customer_name = 'Nordic Precision Engineering Exporters AB', customer_country = 'SE'
    WHERE customer_name LIKE '%Sunrise%' OR customer_name LIKE '%East Coast%' OR customer_name LIKE '%Alpha%'
""")

cur.execute("""
    UPDATE customer_invoices 
    SET customer_name = 'Bavaria Chemical & Specialty Resins SE', customer_country = 'DE'
    WHERE customer_name LIKE '%Global Traders%' OR customer_name LIKE '%Southern Retail%'
""")

cur.execute("""
    UPDATE debit_notes 
    SET customer_name = 'Euro-Asia Electronics Manufacturing Exporters Ltd', customer_country = 'NL'
    WHERE id = 2
""")

cur.execute("""
    UPDATE debit_notes 
    SET customer_name = 'Apex Global Precision Tooling Exporters Corp', customer_country = 'US'
    WHERE id = 1
""")

print("All existing records cleaned with professional exporter names.")
conn.close()
print("\n>>> ALL DEMO DATA SEEDED & LINKED END-TO-END! <<<")
