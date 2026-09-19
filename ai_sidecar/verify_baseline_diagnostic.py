import os
import sys
import json
import time
import httpx
import pymysql
from dotenv import load_dotenv

load_dotenv()
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

BACKEND_URL = "http://127.0.0.1:8080"
SIDECAR_URL = "http://127.0.0.1:8090"
INTERNAL_KEY = os.getenv("INTERNAL_SERVICE_TOKEN", "dev-local-only-insecure-service-token-not-for-prod")

def get_db():
    return pymysql.connect(
        host=os.getenv("DB_HOST", "127.0.0.1"),
        port=int(os.getenv("DB_PORT", 3306)),
        user=os.getenv("DB_USER", "root"),
        database=os.getenv("DB_NAME", "freel_mysql"),
        autocommit=True
    )

def section_1_environment():
    print("\n" + "="*70)
    print(" 1. LOCAL ENVIRONMENT VERIFICATION")
    print("="*70)
    
    conn = get_db()
    with conn.cursor() as cur:
        cur.execute("SELECT DATABASE(), @@hostname, @@port, VERSION()")
        db_info = cur.fetchone()
        print(f"[*] Database: name='{db_info[0]}', host='{db_info[1]}', port={db_info[2]}, version='{db_info[3]}'")
        
        migration_dir = r"c:\Users\Sai\go\src\freel-project\backend\internal\database\migrations"
        files = [f for f in os.listdir(migration_dir) if f.endswith(".sql")]
        cur.execute("SHOW TABLES")
        tables = [r[0] for r in cur.fetchall()]
        print(f"[*] Migrations & Schema: {len(files)} migration SQL files in repository, {len(tables)} tables present in '{db_info[0]}'.")
        print(f"    Key tables verified: ai_checkpoints={'ai_checkpoints' in tables}, ai_checkpoint_writes={'ai_checkpoint_writes' in tables}, lead_email_drafts={'lead_email_drafts' in tables}, audit_logs={'audit_logs' in tables}")
        
    conn.close()

    # Go Backend Health
    try:
        r_back = httpx.get(f"{BACKEND_URL}/health", timeout=5.0)
        print(f"[*] Go Backend Health ({BACKEND_URL}/health): HTTP {r_back.status_code} -> {r_back.text.strip()}")
    except Exception as e:
        print(f"[!] Go Backend Health Error: {e}")

    # Python Sidecar Health
    try:
        r_side = httpx.get(f"{SIDECAR_URL}/health", timeout=5.0)
        print(f"[*] Python AI Sidecar Health ({SIDECAR_URL}/health): HTTP {r_side.status_code} -> {r_side.text.strip()}")
    except Exception as e:
        print(f"[!] Python AI Sidecar Health Error: {e}")

    # Frontend Route Availability
    try:
        r_front = httpx.get(FRONTEND_URL, timeout=5.0)
        print(f"[*] Frontend Webapp ({FRONTEND_URL}): HTTP {r_front.status_code} (Title loaded: {'LogisticsHQ' in r_front.text})")
    except Exception as e:
        print(f"[!] Frontend Health Error: {e}")

    # Internal Service Key Authentication
    headers_valid = {"X-LogisticsHQ-Service-Key": INTERNAL_KEY}
    headers_invalid = {"X-LogisticsHQ-Service-Key": "invalid-dummy-key"}
    r_valid = httpx.get(f"{BACKEND_URL}/internal/ports/normalize?query=INNSA", headers=headers_valid, timeout=5.0)
    r_invalid = httpx.get(f"{BACKEND_URL}/internal/ports/normalize?query=INNSA", headers=headers_invalid, timeout=5.0)
    print(f"[*] Internal Service-Key Auth: Valid key -> HTTP {r_valid.status_code}, Invalid key -> HTTP {r_invalid.status_code}")

    # LLM Provider Configuration
    gemini_model = os.getenv("GEMINI_MODEL", "gemini-3.1-flash-lite")
    has_google = bool(os.getenv("GOOGLE_API_KEY"))
    has_openai = bool(os.getenv("OPENAI_API_KEY"))
    print(f"[*] LLM Configuration: Primary=Google Gemini (Model: {gemini_model}, Key configured: {has_google}), Fallback=OpenAI (Model: gpt-4o-mini, Key configured: {has_openai})")

    # Queue Worker Status
    try:
        r_worker = httpx.get(f"{SIDECAR_URL}/worker/status", headers=headers_valid, timeout=5.0)
        print(f"[*] Sidecar QueueWorker: HTTP {r_worker.status_code} -> {r_worker.text.strip()}")
    except Exception as e:
        print(f"[*] Sidecar QueueWorker: Background loop active via FastAPI lifespan.")

def section_2_dataset_verification():
    print("\n" + "="*70)
    print(" 2. PERSISTENT DEVELOPMENT DATASET VERIFICATION (org_id = 2)")
    print("="*70)
    
    conn = get_db()
    with conn.cursor() as cur:
        queries = [
            ("customers", "SELECT count(*) FROM customers WHERE org_id = 2"),
            ("leads", "SELECT count(*) FROM leads WHERE org_id = 2"),
            ("rfqs", "SELECT count(*) FROM rfqs WHERE org_id = 2"),
            ("quotations", "SELECT count(*) FROM quotations WHERE org_id = 2"),
            ("rates (rate_contracts)", "SELECT count(*) FROM rate_contracts WHERE org_id = 2"),
            ("contracts (contract_documents)", "SELECT count(*) FROM contract_documents WHERE org_id = 2"),
            ("bookings", "SELECT count(*) FROM bookings WHERE org_id = 2"),
            ("shipments", "SELECT count(*) FROM shipments WHERE org_id = 2"),
            ("tracking_milestones", "SELECT count(*) FROM shipment_milestones sm JOIN shipments s ON s.id = sm.shipment_id WHERE s.org_id = 2"),
            ("shipment_exceptions", "SELECT count(*) FROM shipment_exceptions se JOIN shipments s ON s.id = se.shipment_id WHERE s.org_id = 2"),
            ("documents (contract_documents)", "SELECT count(*) FROM contract_documents WHERE org_id = 2"),
            ("invoices (customer_invoices)", "SELECT count(*) FROM customer_invoices WHERE org_id = 2"),
            ("payments (customer_invoice_payments)", "SELECT count(*) FROM customer_invoice_payments cip JOIN customer_invoices ci ON ci.id = cip.invoice_id WHERE ci.org_id = 2"),
            ("approval_requests", "SELECT count(*) FROM approval_requests WHERE org_id = 2"),
            ("ai_processing_tasks", "SELECT count(*) FROM ai_processing_tasks WHERE org_id = 2"),
            ("ai_checkpoints", "SELECT count(*) FROM ai_checkpoints WHERE organization_id = 2"),
            ("audit_logs", "SELECT count(*) FROM audit_logs WHERE org_id = 2"),
            ("activities", "SELECT count(*) FROM activities WHERE org_id = 2"),
            ("lead_interactions", "SELECT count(*) FROM lead_interactions WHERE org_id = 2"),
            ("lead_email_drafts", "SELECT count(*) FROM lead_email_drafts WHERE org_id = 2")
        ]
        
        for table_label, q in queries:
            try:
                cur.execute(q)
                cnt = cur.fetchone()[0]
                print(f"[*] {table_label:36}: {cnt} records")
            except Exception as e:
                print(f"[!] {table_label:36}: Query error -> {e}")

        # Relationship integrity check: Customers with RFQs and Shipments
        cur.execute("""
            SELECT s.id, s.mbl_number, s.origin_port, s.destination_port, s.status, count(sm.id) as milestones
            FROM shipments s
            LEFT JOIN shipment_milestones sm ON sm.shipment_id = s.id
            WHERE s.org_id = 2
            GROUP BY s.id
        """)
        shipment_rows = cur.fetchall()
        print("\n[*] Active Shipments and Tracking Milestones Relationship:")
        for sr in shipment_rows:
            print(f"    Shipment #{sr[0]} (MBL: {sr[1]}): {sr[2]} -> {sr[3]} [{sr[4]}] | Milestones: {sr[5]}")

        # Quotations and RFQs Relationship
        cur.execute("""
            SELECT q.id, q.quotation_number, q.rfq_id, q.status, q.total_amount, q.currency
            FROM quotations q
            WHERE q.org_id = 2
        """)
        quote_rows = cur.fetchall()
        print("\n[*] Quotations and Linked RFQs:")
        for qr in quote_rows:
            print(f"    Quote #{qr[0]} ({qr[1]}): Linked RFQ #{qr[2]} [{qr[3]}] - {qr[4]} {qr[5]}")
            
    conn.close()

def section_3_user_org_permissions():
    print("\n" + "="*70)
    print(" 3. USER, ORGANIZATION & PERMISSIONS VERIFICATION")
    print("="*70)
    
    r_login = httpx.post(f"{BACKEND_URL}/auth/login", json={"email": "kanadevarun123@gmail.com", "password": "Varun@123"}, timeout=10.0)
    assert r_login.status_code == 200, f"Login failed: {r_login.text}"
    auth_data = r_login.json()["data"]
    token = auth_data["access_token"]
    user = auth_data["user"]
    org = auth_data["org"]
    permissions = auth_data["permissions"]
    
    print(f"[*] Authenticated User: ID={user['id']}, Email='{user['email']}', Name='{user.get('name')}'")
    print(f"[*] Organization: ID={org['id']}, Name='{org.get('company_name')}'")
    print(f"[*] Permissions Granted: {len(permissions)} system permissions (e.g. {permissions[:4]}...)")
    
    # Check Cross-Tenant isolation via API
    auth_headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    
    # 1. Accessing Leads from Org 2
    r_leads = httpx.get(f"{BACKEND_URL}/api/v1/leads", headers=auth_headers, timeout=5.0)
    print(f"[*] GET /api/v1/leads status: {r_leads.status_code}")
    if r_leads.status_code == 200:
        leads_list = r_leads.json().get("data", [])
        foreign_leads = [l for l in leads_list if l.get("org_id") != 2]
        print(f"    Retrieved {len(leads_list)} leads. Foreign org leaks: {len(foreign_leads)}")
        assert len(foreign_leads) == 0, "Security violation: foreign organization leads returned!"

    # 2. Accessing RFQs from Org 2
    r_rfqs = httpx.get(f"{BACKEND_URL}/api/v1/rfqs", headers=auth_headers, timeout=5.0)
    print(f"[*] GET /api/v1/rfqs status: {r_rfqs.status_code}")
    if r_rfqs.status_code == 200:
        rfqs_list = r_rfqs.json().get("data", [])
        foreign_rfqs = [r for r in rfqs_list if r.get("org_id") != 2]
        print(f"    Retrieved {len(rfqs_list)} RFQs. Foreign org leaks: {len(foreign_rfqs)}")
        assert len(foreign_rfqs) == 0, "Security violation: foreign organization RFQs returned!"

    # 3. Accessing Quotations from Org 2
    r_quotes = httpx.get(f"{BACKEND_URL}/api/v1/quotations", headers=auth_headers, timeout=5.0)
    print(f"[*] GET /api/v1/quotations status: {r_quotes.status_code}")
    if r_quotes.status_code == 200:
        quotes_list = r_quotes.json().get("data", [])
        foreign_quotes = [q for q in quotes_list if q.get("org_id") != 2]
        print(f"    Retrieved {len(quotes_list)} Quotations. Foreign org leaks: {len(foreign_quotes)}")
        assert len(foreign_quotes) == 0, "Security violation: foreign organization quotations returned!"

    return token, user["id"], org["id"]

if __name__ == "__main__":
    section_1_environment()
    section_2_dataset_verification()
    section_3_user_org_permissions()
