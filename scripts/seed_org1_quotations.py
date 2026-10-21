import pymysql
from datetime import datetime, date, timedelta

conn = pymysql.connect(host='127.0.0.1', port=3306, user='root', password='', db='freel_mysql', autocommit=True)
cur = conn.cursor(pymysql.cursors.DictCursor)

print("1. Updating Org 1 customer names...")
customer_names = {
    1: 'Tata Exports Ltd',
    2: 'Sun Pharma International',
    3: 'Reliance Petrochem Global',
    4: 'Mahindra Auto Logistics',
    5: 'Adani Ports & SEZ'
}
for cid, cname in customer_names.items():
    cur.execute("UPDATE customers SET name = %s WHERE id = %s AND org_id = 1", (cname, cid))

print("2. Ensuring user 5 has access to Org 2 as well...")
cur.execute("""
    INSERT INTO org_members (org_id, user_id, role_id, status, created_at, updated_at)
    VALUES (2, 5, 1, 'ACTIVE', NOW(), NOW())
    ON DUPLICATE KEY UPDATE status = 'ACTIVE'
""")

print("3. Seeding Quotation Templates for Org 1...")
cur.execute("""
    INSERT INTO quotation_templates (
        id, org_id, name, description, shipment_mode, transport_mode, 
        currency, validity_days, payment_terms, commercial_terms, 
        customer_notes, internal_notes, is_active, created_by, created_at, updated_at
    ) VALUES
    (11, 1, 'Standard Ocean FCL Export', 'Standard complete charge template for full container export shipments', 'FCL', 'Ocean Freight', 'USD', 30, 'PREPAID', 'Rates are subject to standard carrier GRI, Bunker Adjustment Factor fluctuations, and space availability at the time of booking. Free time at destination: 7 calendar days.', 'Thank you for choosing LogisticsHQ. Please confirm acceptance within validity period to secure equipment and space.', 'Standard 15% commercial markup applied. Check carrier direct contract allocation before confirmation.', 1, 'System Default', NOW(), NOW()),
    (12, 1, 'Standard Air Freight Express', 'Turnkey air freight template including airline handling and security', 'Air Freight', 'Air Freight', 'USD', 14, 'PREPAID', 'Rates are based on chargeable weight (1 CBM = 167 kg). Subject to airline fuel and security surcharges applicable on flight departure date.', 'Fast-track airport-to-airport transit. Estimated transit time: 3-5 days.', 'Verify airline flight schedule and cut-off times before issuing booking confirmation.', 1, 'System Default', NOW(), NOW())
    ON DUPLICATE KEY UPDATE name = VALUES(name)
""")

print("4. Seeding Quotations for Org 1...")
today = date.today()
quotes_data = [
    {
        "id": 1001,
        "org_id": 1,
        "quotation_number": "QT-2026-1001",
        "customer_id": 1,
        "customer_name": "Tata Exports Ltd",
        "rfq_id": 1,
        "rfq_number": "RFQ-2026-1001",
        "status": "ACCEPTED",
        "origin": "Nhava Sheva",
        "origin_code": "INNSA",
        "destination": "Hamburg",
        "destination_code": "DEHAM",
        "service_type": "PORT_TO_PORT",
        "transport_mode": "Ocean FCL",
        "currency": "USD",
        "payment_terms": "NET_30",
        "subtotal": 2550.00,
        "surcharges": 300.00,
        "taxes": 0.00,
        "total_amount": 2850.00,
        "total_cost": 2150.00,
        "gross_profit": 700.00,
        "gross_margin_pct": 24.56,
        "valid_from": today - timedelta(days=20),
        "valid_until": today + timedelta(days=25),
        "sent_at": datetime.now() - timedelta(days=18),
        "sent_by": "Varun Kanade",
        "accepted_at": datetime.now() - timedelta(days=5),
        "conversion_status": "CONVERTED",
        "converted_booking_id": 101,
        "converted_at": datetime.now() - timedelta(days=5),
        "commercial_terms": "Standard FCL terms. Detention free time: 14 days at port of discharge.",
        "customer_notes": "All-in ocean freight inclusive of BAF, CAF, and ISPS.",
        "created_by": "Varun Kanade"
    },
    {
        "id": 1002,
        "org_id": 1,
        "quotation_number": "QT-2026-1002",
        "customer_id": 2,
        "customer_name": "Sun Pharma International",
        "rfq_id": 2,
        "rfq_number": "RFQ-2026-1002",
        "status": "SENT",
        "origin": "Mumbai Air",
        "origin_code": "INBOM",
        "destination": "Frankfurt Air",
        "destination_code": "FRA",
        "service_type": "AIRPORT_TO_AIRPORT",
        "transport_mode": "Air Freight",
        "currency": "USD",
        "payment_terms": "PREPAID",
        "subtotal": 3800.00,
        "surcharges": 280.00,
        "taxes": 0.00,
        "total_amount": 4080.00,
        "total_cost": 3400.00,
        "gross_profit": 680.00,
        "gross_margin_pct": 16.67,
        "valid_from": today - timedelta(days=3),
        "valid_until": today + timedelta(days=11),
        "sent_at": datetime.now() - timedelta(days=3),
        "sent_by": "Varun Kanade",
        "conversion_status": "NOT_CONVERTED",
        "commercial_terms": "Pharma temperature-controlled packaging compliant (+15C to +25C).",
        "customer_notes": "Priority scheduled flight booking confirmed on Lufthansa Cargo.",
        "created_by": "Varun Kanade"
    },
    {
        "id": 1003,
        "org_id": 1,
        "quotation_number": "QT-2026-1003",
        "customer_id": 3,
        "customer_name": "Reliance Petrochem Global",
        "rfq_id": 3,
        "rfq_number": "RFQ-2026-1003",
        "status": "READY_FOR_REVIEW",
        "origin": "Nhava Sheva",
        "origin_code": "INNSA",
        "destination": "New York",
        "destination_code": "USNYC",
        "service_type": "PORT_TO_PORT",
        "transport_mode": "Ocean FCL",
        "currency": "USD",
        "payment_terms": "NET_30",
        "subtotal": 3100.00,
        "surcharges": 320.00,
        "taxes": 0.00,
        "total_amount": 3420.00,
        "total_cost": 2850.00,
        "gross_profit": 570.00,
        "gross_margin_pct": 16.67,
        "valid_from": today - timedelta(days=1),
        "valid_until": today + timedelta(days=29),
        "submitted_for_review_at": datetime.now() - timedelta(hours=4),
        "submitted_for_review_by": "Aditya Kumar",
        "conversion_status": "NOT_CONVERTED",
        "commercial_terms": "Direct routing via Suez/Atlantic. Free time at destination: 10 days.",
        "created_by": "Aditya Kumar"
    },
    {
        "id": 1004,
        "org_id": 1,
        "quotation_number": "QT-2026-1004",
        "customer_id": 1,
        "customer_name": "Tata Exports Ltd",
        "rfq_id": 4,
        "rfq_number": "RFQ-2026-1004",
        "status": "APPROVED",
        "origin": "Kolkata",
        "origin_code": "INCCU",
        "destination": "Singapore",
        "destination_code": "SGSIN",
        "service_type": "PORT_TO_PORT",
        "transport_mode": "Ocean FCL",
        "currency": "USD",
        "payment_terms": "NET_15",
        "subtotal": 1450.00,
        "surcharges": 150.00,
        "taxes": 0.00,
        "total_amount": 1600.00,
        "total_cost": 1250.00,
        "gross_profit": 350.00,
        "gross_margin_pct": 21.88,
        "valid_from": today - timedelta(days=2),
        "valid_until": today + timedelta(days=28),
        "approved_at": datetime.now() - timedelta(hours=12),
        "approved_by": "Varun Kanade",
        "approval_notes": "Approved standard industrial export quote.",
        "conversion_status": "NOT_CONVERTED",
        "created_by": "Priya Sharma"
    },
    {
        "id": 1005,
        "org_id": 1,
        "quotation_number": "QT-2026-1005",
        "customer_id": 4,
        "customer_name": "Mahindra Auto Logistics",
        "rfq_id": 5,
        "rfq_number": "RFQ-2026-1005",
        "status": "DRAFT",
        "origin": "Bangalore",
        "origin_code": "INBLR",
        "destination": "Abu Dhabi",
        "destination_code": "AEAUH",
        "service_type": "DOOR_TO_PORT",
        "transport_mode": "Multimodal",
        "currency": "USD",
        "payment_terms": "PREPAID",
        "subtotal": 2100.00,
        "surcharges": 200.00,
        "taxes": 0.00,
        "total_amount": 2300.00,
        "total_cost": 1800.00,
        "gross_profit": 500.00,
        "gross_margin_pct": 21.74,
        "valid_from": today,
        "valid_until": today + timedelta(days=30),
        "conversion_status": "NOT_CONVERTED",
        "created_by": "Priya Sharma"
    },
    {
        "id": 1006,
        "org_id": 1,
        "quotation_number": "QT-2026-1006",
        "customer_id": 5,
        "customer_name": "Adani Ports & SEZ",
        "rfq_id": 6,
        "rfq_number": "RFQ-2026-1006",
        "status": "EXPIRED",
        "origin": "Mundra",
        "origin_code": "INMUN",
        "destination": "Shanghai",
        "destination_code": "CNSHA",
        "service_type": "PORT_TO_PORT",
        "transport_mode": "Ocean FCL",
        "currency": "USD",
        "payment_terms": "PREPAID",
        "subtotal": 1200.00,
        "surcharges": 150.00,
        "taxes": 0.00,
        "total_amount": 1350.00,
        "total_cost": 1100.00,
        "gross_profit": 250.00,
        "gross_margin_pct": 18.52,
        "valid_from": today - timedelta(days=45),
        "valid_until": today - timedelta(days=15),
        "sent_at": datetime.now() - timedelta(days=40),
        "expired_at": datetime.now() - timedelta(days=15),
        "conversion_status": "NOT_CONVERTED",
        "created_by": "Priya Sharma"
    },
    {
        "id": 1007,
        "org_id": 1,
        "quotation_number": "QT-2026-1007",
        "customer_id": 2,
        "customer_name": "Sun Pharma International",
        "rfq_id": 2,
        "rfq_number": "RFQ-2026-1002",
        "status": "VIEWED",
        "origin": "Mumbai Air",
        "origin_code": "INBOM",
        "destination": "Chicago O'Hare",
        "destination_code": "ORD",
        "service_type": "AIRPORT_TO_AIRPORT",
        "transport_mode": "Air Freight",
        "currency": "USD",
        "payment_terms": "NET_30",
        "subtotal": 5200.00,
        "surcharges": 450.00,
        "taxes": 0.00,
        "total_amount": 5650.00,
        "total_cost": 4600.00,
        "gross_profit": 1050.00,
        "gross_margin_pct": 18.58,
        "valid_from": today - timedelta(days=4),
        "valid_until": today + timedelta(days=10),
        "sent_at": datetime.now() - timedelta(days=4),
        "viewed_at": datetime.now() - timedelta(days=1),
        "view_count": 3,
        "conversion_status": "NOT_CONVERTED",
        "created_by": "Varun Kanade"
    },
    {
        "id": 1008,
        "org_id": 1,
        "quotation_number": "QT-2026-1008",
        "customer_id": 1,
        "customer_name": "Tata Exports Ltd",
        "rfq_id": 1,
        "rfq_number": "RFQ-2026-1001",
        "status": "CHANGES_REQUESTED",
        "origin": "Nhava Sheva",
        "origin_code": "INNSA",
        "destination": "Rotterdam",
        "destination_code": "NLRTM",
        "service_type": "PORT_TO_PORT",
        "transport_mode": "Ocean FCL",
        "currency": "USD",
        "payment_terms": "NET_30",
        "subtotal": 2900.00,
        "surcharges": 350.00,
        "taxes": 0.00,
        "total_amount": 3250.00,
        "total_cost": 2700.00,
        "gross_profit": 550.00,
        "gross_margin_pct": 16.92,
        "valid_from": today - timedelta(days=5),
        "valid_until": today + timedelta(days=25),
        "sent_at": datetime.now() - timedelta(days=5),
        "changes_requested_at": datetime.now() - timedelta(days=2),
        "changes_requested_by": "Ravi Mehta (Client)",
        "changes_requested_reason": "Requested 7 extra detention days at destination due to customs clearance queue.",
        "conversion_status": "NOT_CONVERTED",
        "created_by": "Varun Kanade"
    },
    {
        "id": 1009,
        "org_id": 1,
        "quotation_number": "QT-2026-1009",
        "customer_id": 3,
        "customer_name": "Reliance Petrochem Global",
        "rfq_id": 3,
        "rfq_number": "RFQ-2026-1003",
        "status": "ACCEPTED",
        "origin": "Mundra",
        "origin_code": "INMUN",
        "destination": "Jebel Ali",
        "destination_code": "AEJEA",
        "service_type": "PORT_TO_PORT",
        "transport_mode": "Ocean FCL",
        "currency": "USD",
        "payment_terms": "NET_30",
        "subtotal": 1850.00,
        "surcharges": 200.00,
        "taxes": 0.00,
        "total_amount": 2050.00,
        "total_cost": 1600.00,
        "gross_profit": 450.00,
        "gross_margin_pct": 21.95,
        "valid_from": today - timedelta(days=10),
        "valid_until": today + timedelta(days=20),
        "sent_at": datetime.now() - timedelta(days=10),
        "accepted_at": datetime.now() - timedelta(days=2),
        "conversion_status": "CONVERTED",
        "created_by": "Aditya Kumar"
    },
    {
        "id": 1010,
        "org_id": 1,
        "quotation_number": "QT-2026-1010",
        "customer_id": 4,
        "customer_name": "Mahindra Auto Logistics",
        "rfq_id": 5,
        "rfq_number": "RFQ-2026-1005",
        "status": "DECLINED",
        "origin": "Chennai",
        "origin_code": "INMAA",
        "destination": "Felixstowe",
        "destination_code": "GBFXT",
        "service_type": "PORT_TO_PORT",
        "transport_mode": "Ocean FCL",
        "currency": "USD",
        "payment_terms": "PREPAID",
        "subtotal": 2400.00,
        "surcharges": 250.00,
        "taxes": 0.00,
        "total_amount": 2650.00,
        "total_cost": 2100.00,
        "gross_profit": 550.00,
        "gross_margin_pct": 20.75,
        "valid_from": today - timedelta(days=20),
        "valid_until": today - timedelta(days=5),
        "sent_at": datetime.now() - timedelta(days=19),
        "declined_at": datetime.now() - timedelta(days=7),
        "declined_reason": "Customer chose alternate carrier due to transit time urgency.",
        "conversion_status": "NOT_CONVERTED",
        "created_by": "Priya Sharma"
    }
]

for q in quotes_data:
    cols = list(q.keys())
    vals = list(q.values())
    placeholders = ", ".join(["%s"] * len(vals))
    updates = ", ".join([f"{c} = VALUES({c})" for c in cols if c != "id"])
    sql = f"INSERT INTO quotations ({', '.join(cols)}) VALUES ({placeholders}) ON DUPLICATE KEY UPDATE {updates}"
    cur.execute(sql, vals)

print(f"Successfully seeded {len(quotes_data)} quotations for Org 1.")

print("5. Seeding quotation charge items for Org 1...")
charges = [
    (11, 1, 1001, 'OF-40GP', 'Ocean Freight 40GP', 'FREIGHT', 'STANDARD', 'PER_CONTAINER', 1.0, 2200.00, 1850.00, 2200.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 1850.00, 2200.00, 1, 0, 'Nhava Sheva to Hamburg', 'Varun Kanade', 'Varun Kanade'),
    (12, 1, 1001, 'THC-ORIGIN', 'Origin Terminal Handling', 'ORIGIN', 'STANDARD', 'FLAT', 1.0, 350.00, 300.00, 350.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 300.00, 350.00, 2, 0, 'Port terminal handling', 'Varun Kanade', 'Varun Kanade'),
    (13, 1, 1001, 'DOC-FEE', 'Bill of Lading Documentation Fee', 'DOCUMENTATION', 'STANDARD', 'FLAT', 1.0, 300.00, 0.00, 300.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 0.00, 300.00, 3, 0, 'B/L issuance and filing', 'Varun Kanade', 'Varun Kanade'),
    (14, 1, 1002, 'AF-PERKG', 'Air Freight Main Leg', 'FREIGHT', 'STANDARD', 'PER_WEIGHT', 800.0, 4.50, 4.00, 3600.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 3200.00, 3600.00, 1, 0, 'Chargeable weight 800kg', 'Varun Kanade', 'Varun Kanade'),
    (15, 1, 1002, 'FSC-AIR', 'Air Fuel Surcharge', 'SURCHARGE', 'STANDARD', 'PER_WEIGHT', 800.0, 0.35, 0.25, 280.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 200.00, 280.00, 2, 0, 'Airline fuel indexing', 'Varun Kanade', 'Varun Kanade'),
    (16, 1, 1002, 'SEC-AIR', 'Airport Security & Screening', 'SURCHARGE', 'STANDARD', 'FLAT', 1.0, 200.00, 0.00, 200.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 0.00, 200.00, 3, 0, 'X-ray screening', 'Varun Kanade', 'Varun Kanade'),
    (17, 1, 1003, 'OF-40HC', 'Ocean Freight 40HC', 'FREIGHT', 'STANDARD', 'PER_CONTAINER', 1.0, 1800.00, 1400.00, 1800.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 1400.00, 1800.00, 1, 0, 'Mundra to Felixstowe', 'Varun Kanade', 'Varun Kanade'),
    (18, 1, 1003, 'THC-ORIGIN', 'Origin Terminal Handling', 'ORIGIN', 'STANDARD', 'FLAT', 1.0, 200.00, 150.00, 200.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 150.00, 200.00, 2, 0, 'Terminal handling Mundra', 'Varun Kanade', 'Varun Kanade'),
    (19, 1, 1004, 'OF-20GP', 'Ocean Freight 20GP', 'FREIGHT', 'STANDARD', 'PER_CONTAINER', 1.0, 1400.00, 1150.00, 1400.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 1150.00, 1400.00, 1, 0, 'Kolkata to Singapore Main Leg', 'Varun Kanade', 'Varun Kanade'),
    (20, 1, 1004, 'THC-ORIGIN', 'Origin Terminal Handling (THC)', 'ORIGIN', 'STANDARD', 'FLAT', 1.0, 150.00, 120.00, 150.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 120.00, 150.00, 2, 0, 'Port terminal handling charges', 'Varun Kanade', 'Varun Kanade'),
    (21, 1, 1004, 'DOC-FEE', 'Bill of Lading Documentation Fee', 'DOCUMENTATION', 'STANDARD', 'FLAT', 1.0, 50.00, 30.00, 50.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 30.00, 50.00, 3, 0, 'B/L issuance and filing', 'Varun Kanade', 'Varun Kanade'),
    (22, 1, 1005, 'OF-40GP', 'Ocean Freight 40GP', 'FREIGHT', 'STANDARD', 'PER_CONTAINER', 1.0, 3100.00, 2500.00, 3100.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 2500.00, 3100.00, 1, 0, 'Chennai to Los Angeles', 'Varun Kanade', 'Varun Kanade'),
    (23, 1, 1005, 'BAF-SUR', 'Bunker Adjustment Factor (BAF)', 'SURCHARGE', 'STANDARD', 'FLAT', 1.0, 350.00, 300.00, 350.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 300.00, 350.00, 2, 0, 'Fuel indexing adjustment', 'Varun Kanade', 'Varun Kanade'),
    (24, 1, 1006, 'OF-20GP', 'Ocean Freight 20GP', 'FREIGHT', 'STANDARD', 'PER_CONTAINER', 1.0, 2750.00, 2250.00, 2750.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 2250.00, 2750.00, 1, 0, 'Nhava Sheva to Rotterdam', 'Varun Kanade', 'Varun Kanade'),
    (25, 1, 1006, 'THC-ORIGIN', 'Origin Terminal Handling', 'ORIGIN', 'STANDARD', 'FLAT', 1.0, 300.00, 250.00, 300.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 250.00, 300.00, 2, 0, 'Port terminal handling', 'Varun Kanade', 'Varun Kanade'),
    (26, 1, 1007, 'AF-PERKG', 'Air Freight Heavy Cargo', 'FREIGHT', 'STANDARD', 'PER_WEIGHT', 1200.0, 4.33, 3.83, 5200.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 4600.00, 5200.00, 1, 0, 'Direct express charter leg', 'Varun Kanade', 'Varun Kanade'),
    (27, 1, 1007, 'FSC-AIR', 'Fuel Surcharge', 'SURCHARGE', 'STANDARD', 'FLAT', 1.0, 450.00, 0.00, 450.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 0.00, 450.00, 2, 0, 'Carrier fuel surcharge', 'Varun Kanade', 'Varun Kanade'),
    (28, 1, 1008, 'OF-40GP', 'Ocean Freight 40GP', 'FREIGHT', 'STANDARD', 'PER_CONTAINER', 1.0, 2900.00, 2400.00, 2900.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 2400.00, 2900.00, 1, 0, 'Nhava Sheva to Rotterdam', 'Varun Kanade', 'Varun Kanade'),
    (29, 1, 1008, 'THC-ORIGIN', 'Terminal Handling Charges', 'ORIGIN', 'STANDARD', 'FLAT', 1.0, 350.00, 300.00, 350.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 300.00, 350.00, 2, 0, 'Terminal handling', 'Varun Kanade', 'Varun Kanade'),
    (30, 1, 1009, 'OF-20GP', 'Ocean Freight 20GP', 'FREIGHT', 'STANDARD', 'PER_CONTAINER', 1.0, 1850.00, 1450.00, 1850.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 1450.00, 1850.00, 1, 0, 'Mundra to Jebel Ali', 'Varun Kanade', 'Varun Kanade'),
    (31, 1, 1009, 'DOC-FEE', 'Documentation Fee', 'DOCUMENTATION', 'STANDARD', 'FLAT', 1.0, 200.00, 150.00, 200.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 150.00, 200.00, 2, 0, 'B/L Documentation', 'Varun Kanade', 'Varun Kanade'),
    (32, 1, 1010, 'OF-40GP', 'Ocean Freight 40GP', 'FREIGHT', 'STANDARD', 'PER_CONTAINER', 1.0, 2400.00, 1900.00, 2400.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 1900.00, 2400.00, 1, 0, 'Chennai to Felixstowe', 'Varun Kanade', 'Varun Kanade'),
    (33, 1, 1010, 'THC-ORIGIN', 'Terminal Handling Charges', 'ORIGIN', 'STANDARD', 'FLAT', 1.0, 250.00, 200.00, 250.00, 'USD', 1.0, 0.0, 0.0, 'NONE', 0.0, 0.0, 200.00, 250.00, 2, 0, 'Terminal handling Chennai', 'Varun Kanade', 'Varun Kanade'),
]

for ch in charges:
    cur.execute("""
        INSERT INTO quotation_charge_items (
            id, org_id, quotation_id, charge_code, charge_name, charge_category, charge_type,
            calculation_basis, quantity, unit_price, cost_amount, sell_amount, currency,
            exchange_rate, tax_rate, tax_amount, discount_type, discount_value, discount_amount,
            total_cost, total_sell, display_order, is_optional, notes, created_by, updated_by, created_at, updated_at
        ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, NOW(), NOW())
        ON DUPLICATE KEY UPDATE charge_name = VALUES(charge_name), sell_amount = VALUES(sell_amount)
    """, ch)

print("6. Seeding sample quotation activity for Org 1...")
activities = [
    (1001, 1, 1001, 'CREATED', 'Quotation QT-2026-1001 created from RFQ-2026-1001', 'Varun Kanade', '{}'),
    (1002, 1, 1001, 'SENT', 'Quotation sent to Rajesh Verma (r.verma@tataexports.com)', 'Varun Kanade', '{}'),
    (1003, 1, 1001, 'ACCEPTED', 'Quotation accepted by customer. Converted to Booking BK-2026-1001', 'Varun Kanade', '{}'),
    (1004, 1, 1002, 'CREATED', 'Air freight quote created for Sun Pharma', 'Varun Kanade', '{}'),
    (1005, 1, 1002, 'SENT', 'Sent to Sneha Patel via secure portal link', 'Varun Kanade', '{}'),
    (1006, 1, 1003, 'SUBMITTED_REVIEW', 'Quotation submitted for management margin sign-off', 'Aditya Kumar', '{}'),
]
for act in activities:
    cur.execute("""
        INSERT INTO quotation_activity (
            id, org_id, quotation_id, activity_type, description, actor, metadata, created_at
        ) VALUES (%s, %s, %s, %s, %s, %s, %s, NOW())
        ON DUPLICATE KEY UPDATE description = VALUES(description)
    """, act)

print("Done! Checking counts:")
cur.execute("SELECT org_id, COUNT(*) as c FROM quotations GROUP BY org_id")
print("Quotations by org:", cur.fetchall())
conn.close()
