import pymysql

conn = pymysql.connect(
    host='127.0.0.1',
    port=3306,
    user='root',
    password='',
    database='freel_mysql',
    autocommit=True
)

cursor = conn.cursor()

# 1. Update customers
cursor.execute("""
UPDATE customers SET 
  website = 'https://nordicfreight.se',
  credit_limit = 150000.00,
  payment_terms = 'NET30',
  credit_status = 'GOOD_STANDING',
  commercial_notes = 'Verified tier-1 enterprise exporter out of Scandinavia. All shipments backed by corporate letter of credit with Net 30 payment settlement.',
  contact_name = 'Astrid Lindgren',
  contact_email = 'astrid@nordicfreight.se',
  contact_phone = '+46 31 123 4567',
  country = 'Sweden',
  city = 'Gothenburg'
WHERE id IN (1, 102)
""")

# 2. Insert contacts for customer 102 and 1
contacts_sql = """
INSERT INTO contacts (id, org_id, customer_id, first_name, last_name, email, phone, job_title, department, is_primary, contact_role, created_at, updated_at)
VALUES 
(201, 2, 102, 'Astrid', 'Lindgren', 'astrid@nordicfreight.se', '+46 31 123 4567', 'Logistics Operations Lead', 'Logistics', 1, 'COMMERCIAL', NOW(), NOW()),
(202, 2, 102, 'Magnus', 'Svensson', 'magnus.svensson@nordicfreight.se', '+46 31 789 1234', 'VP of Global Procurement', 'Procurement', 0, 'COMMERCIAL', NOW(), NOW()),
(203, 2, 102, 'Elin', 'Bergman', 'elin.bergman@nordicfreight.se', '+46 31 456 7890', 'Senior Finance & Billing Manager', 'Finance', 0, 'FINANCIAL', NOW(), NOW()),
(204, 1, 1, 'Astrid', 'Lindgren', 'astrid@nordicfreight.se', '+46 31 123 4567', 'Logistics Operations Lead', 'Logistics', 1, 'COMMERCIAL', NOW(), NOW()),
(205, 1, 1, 'Magnus', 'Svensson', 'magnus.svensson@nordicfreight.se', '+46 31 789 1234', 'VP of Global Procurement', 'Procurement', 0, 'COMMERCIAL', NOW(), NOW()),
(206, 1, 1, 'Elin', 'Bergman', 'elin.bergman@nordicfreight.se', '+46 31 456 7890', 'Senior Finance & Billing Manager', 'Finance', 0, 'FINANCIAL', NOW(), NOW())
ON DUPLICATE KEY UPDATE first_name=VALUES(first_name), email=VALUES(email), phone=VALUES(phone), job_title=VALUES(job_title)
"""
cursor.execute(contacts_sql)

# 3. Insert addresses for customer 102 and 1
addresses_sql = """
INSERT INTO customer_addresses (id, org_id, customer_id, address_type, label, address_line_1, address_line_2, city, state, postal_code, country_code, country, is_primary_billing, is_primary_shipping, contact_name, contact_phone, contact_email, created_at, updated_at)
VALUES 
(201, 2, 102, 'REGISTERED', 'Corporate Headquarters', 'Skeppsbron 4, Port Operations Tower', 'Floor 7, Suite 702', 'Gothenburg', 'Västra Götaland', '411 18', 'SE', 'Sweden', 1, 0, 'Astrid Lindgren', '+46 31 123 4567', 'astrid@nordicfreight.se', NOW(), NOW()),
(202, 2, 102, 'OPERATIONAL', 'Gothenburg Harbor Logistics Depot', 'Port of Gothenburg Berth 6', 'Gate 3, Heavy Machinery Logistics Hub', 'Gothenburg', 'Västra Götaland', '418 34', 'SE', 'Sweden', 0, 1, 'Magnus Svensson', '+46 31 789 1234', 'magnus.svensson@nordicfreight.se', NOW(), NOW()),
(203, 1, 1, 'REGISTERED', 'Corporate Headquarters', 'Skeppsbron 4, Port Operations Tower', 'Floor 7, Suite 702', 'Gothenburg', 'Västra Götaland', '411 18', 'SE', 'Sweden', 1, 0, 'Astrid Lindgren', '+46 31 123 4567', 'astrid@nordicfreight.se', NOW(), NOW()),
(204, 1, 1, 'OPERATIONAL', 'Gothenburg Harbor Logistics Depot', 'Port of Gothenburg Berth 6', 'Gate 3, Heavy Machinery Logistics Hub', 'Gothenburg', 'Västra Götaland', '418 34', 'SE', 'Sweden', 0, 1, 'Magnus Svensson', '+46 31 789 1234', 'magnus.svensson@nordicfreight.se', NOW(), NOW())
ON DUPLICATE KEY UPDATE address_line_1=VALUES(address_line_1), city=VALUES(city), country=VALUES(country)
"""
cursor.execute(addresses_sql)

print("Successfully seeded Nordic customer contacts, addresses, and finance in DB!")
cursor.close()
conn.close()
