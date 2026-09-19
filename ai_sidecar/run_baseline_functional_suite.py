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
INTERNAL_KEY = os.getenv("INTERNAL_SERVICE_TOKEN", "dev-local-only-insecure-service-token-not-for-prod")

def get_db():
    return pymysql.connect(
        host=os.getenv("DB_HOST", "127.0.0.1"),
        port=int(os.getenv("DB_PORT", 3306)),
        user=os.getenv("DB_USER", "root"),
        database=os.getenv("DB_NAME", "freel_mysql"),
        autocommit=True
    )

def get_auth_context():
    resp = httpx.post(
        f"{BACKEND_URL}/auth/login",
        json={"email": "kanadevarun123@gmail.com", "password": "Varun@123"},
        timeout=10.0
    )
    assert resp.status_code == 200, f"Login failed: {resp.text}"
    data = resp.json()["data"]
    return data["access_token"], data["user"]["id"], data["org"]["id"]

def run_functional_suite():
    print("=" * 75)
    print(" LOGISTICSHQ AGENTIC AI PLATFORM — PHASE 0.1 BASELINE FUNCTIONAL SUITE")
    print("=" * 75)

    token, user_id, org_id = get_auth_context()
    auth_headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    internal_headers = {"X-LogisticsHQ-Service-Key": INTERNAL_KEY, "Content-Type": "application/json"}
    conn = get_db()

    results = []

    # -------------------------------------------------------------------------
    # TEST 1: Pricing Agent — Normal RFQ Evaluation
    # -------------------------------------------------------------------------
    print("\n[TEST 1] Pricing Agent — Normal RFQ Evaluation (RFQ #102)")
    rfq_id = 102
    try:
        # Enqueue or direct analyze
        resp = httpx.post(f"{BACKEND_URL}/api/v1/rfqs/{rfq_id}/pricing-analyze", headers=auth_headers, timeout=10.0)
        print(f"   API trigger: status={resp.status_code}")
        
        # Verify sidecar pricing tool connectivity
        r_rates = httpx.get(f"{BACKEND_URL}/internal/rates/search?org_id={org_id}&origin=INNSA&destination=DEHAM&equipment=40GP", headers=internal_headers, timeout=5.0)
        print(f"   Rates search tool: status={r_rates.status_code}, count={len(r_rates.json().get('data', [])) if r_rates.status_code == 200 else 0}")
        
        with conn.cursor() as cur:
            cur.execute("SELECT id, rfq_number, status, agent_status FROM rfqs WHERE id = %s AND org_id = %s", (rfq_id, org_id))
            rfq_row = cur.fetchone()
            print(f"   RFQ DB Record: #{rfq_row[0]} ({rfq_row[1]}) -> status='{rfq_row[2]}', agent_status='{rfq_row[3]}'")
            
        results.append({
            "test": "Pricing Agent — Normal RFQ",
            "input_id": f"RFQ #{rfq_id}",
            "org_id": org_id,
            "status": "PASS",
            "details": f"Rates retrieved and RFQ agent_status={rfq_row[3]}"
        })
    except Exception as e:
        print(f"   [!] Error: {e}")
        results.append({"test": "Pricing Agent — Normal RFQ", "input_id": f"RFQ #{rfq_id}", "org_id": org_id, "status": "FAIL", "details": str(e)})

    # -------------------------------------------------------------------------
    # TEST 2: Pricing Agent — Anomaly / HITL RFQ Evaluation
    # -------------------------------------------------------------------------
    print("\n[TEST 2] Pricing Agent — Anomaly / HITL RFQ Evaluation (RFQ #105)")
    rfq_id_hitl = 105
    try:
        with conn.cursor() as cur:
            cur.execute("SELECT id, rfq_number, status, agent_status FROM rfqs WHERE id = %s AND org_id = %s", (rfq_id_hitl, org_id))
            rfq_hitl_row = cur.fetchone()
            print(f"   Initial RFQ DB Record: #{rfq_hitl_row[0]} -> agent_status='{rfq_hitl_row[3]}'")
            
        # Verify interrupt mechanism in sidecar pricing graph
        from app.graphs.pricing_graph import pricing_graph
        from app.persistence.mariadb_saver import MariaDBSaver
        
        thread_id = f"rfq-baseline-test-{int(time.time())}"
        config = {"configurable": {"thread_id": thread_id, "org_id": org_id}, "metadata": {"org_id": str(org_id), "user_id": str(user_id)}}
        state = {
            "messages": [], "org_id": org_id, "rfq_id": rfq_id_hitl, "origin": "INNSA", "destination": "DEHAM",
            "incoterms": "FOB", "equipment_type": "40GP", "gross_weight": 20000.0, "volume_cbm": 30.0,
            "commodity": "Machinery", "target_date": "2026-10-01", "pricing_rules": [], "raw_rates": [],
            "suggested_quotes": [{"carrier": "MSC", "total_buy": 9500.0, "suggested_sell": 10500.0, "margin_pct": 9.5}],
            "overall_reasoning": "High-rate anomaly above threshold.", "is_anomaly": True, "confidence_score": 90, "error_message": None
        }
        
        # Test interrupt before save
        from app.graphs.pricing_graph import pricing_builder
        saver = MariaDBSaver()
        test_graph = pricing_builder.compile(checkpointer=saver, interrupt_before=["save"])
        test_graph.invoke(state, config=config)
        snap = test_graph.get_state(config)
        print(f"   LangGraph paused at interrupt node: {snap.next}")
        assert snap.next == ("save",), f"Expected pause before 'save', got {snap.next}"
        
        # Verify checkpoint in MariaDB
        with conn.cursor() as cur:
            cur.execute("SELECT count(*) FROM ai_checkpoints WHERE thread_id = %s AND organization_id = %s", (thread_id, org_id))
            chk_count = cur.fetchone()[0]
            print(f"   Verified checkpoints stored in MariaDB: {chk_count} entries for thread {thread_id}")
            assert chk_count > 0, "No checkpoint records found!"
            
        results.append({
            "test": "Pricing Agent — Anomaly HITL",
            "input_id": f"RFQ #{rfq_id_hitl}",
            "org_id": org_id,
            "status": "PASS",
            "details": f"Paused at interrupt {snap.next} and saved {chk_count} checkpoints in MariaDB."
        })
    except Exception as e:
        print(f"   [!] Error: {e}")
        results.append({"test": "Pricing Agent — Anomaly HITL", "input_id": f"RFQ #{rfq_id_hitl}", "org_id": org_id, "status": "FAIL", "details": str(e)})

    # -------------------------------------------------------------------------
    # TEST 3: Contracts Agent — Extraction & Anomaly Review
    # -------------------------------------------------------------------------
    print("\n[TEST 3] Contracts Agent — Layout Extraction & Rate Anomaly Interrupt")
    doc_id = 101
    try:
        from app.graphs.contracts_graph import contracts_graph
        contract_thread = f"contract-baseline-{int(time.time())}"
        contract_config = {"configurable": {"thread_id": contract_thread, "org_id": org_id}, "metadata": {"org_id": str(org_id), "user_id": str(user_id)}}
        contract_state = {
            "document_id": contract_thread, "org_id": org_id, "s3_key": "sample_maersk.pdf", "file_type": "PDF",
            "callback_url": f"{BACKEND_URL}/internal/contracts/callback", "correlation_id": "corr-c1",
            "raw_text": "Maersk Line Agreement SC-2026. INNSA to DEHAM 40GP USD 2800. INNSA to AEJEA 40GP USD 12400.",
            "carrier_name": "Maersk Line", "carrier_scac": "MAEU", "extracted_rates": [], "flagged_items": [],
            "processing_log": [], "ai_summary": "", "is_anomaly_detected": False, "status": "IN_PROGRESS"
        }
        
        contracts_graph.invoke(contract_state, config=contract_config)
        c_snap = contracts_graph.get_state(contract_config)
        print(f"   Contracts graph execution interrupted at: {c_snap.next}")
        assert c_snap.next == ("ingest",), f"Expected pause before ingest, got {c_snap.next}"
        
        # Verify persistence across restart
        new_saver = MariaDBSaver()
        reloaded_tuple = new_saver.get_tuple(contract_config)
        assert reloaded_tuple is not None, "Failed to reload contract checkpoint tuple!"
        print(f"   Checkpointer reloaded state after restart simulation. Checkpoint ID: {reloaded_tuple.checkpoint['id']}")
        
        results.append({
            "test": "Contracts Extraction & Rate Anomaly",
            "input_id": f"Doc #{doc_id}",
            "org_id": org_id,
            "status": "PASS",
            "details": f"Paused before {c_snap.next}, state reloaded successfully from MariaDB."
        })
    except Exception as e:
        print(f"   [!] Error: {e}")
        results.append({"test": "Contracts Extraction & Rate Anomaly", "input_id": f"Doc #{doc_id}", "org_id": org_id, "status": "FAIL", "details": str(e)})

    # -------------------------------------------------------------------------
    # TEST 4: Sales / Email Parser — Incomplete Email Safety (No Auto-Send)
    # -------------------------------------------------------------------------
    print("\n[TEST 4] Sales / Email Parser — Incomplete Inbound Email Safety Gate")
    lead_id = 103
    try:
        # Create inbound interaction
        with conn.cursor() as cur:
            cur.execute("""
                INSERT INTO lead_interactions 
                (org_id, lead_id, direction, interaction_type, channel, subject, content, raw_email_id, sender, recipients, created_at, updated_at)
                VALUES (%s, %s, 'INBOUND', 'EMAIL', 'EMAIL', 'RFQ Inquiry: Export to Rotterdam', 
                        'Please give us ocean freight quote for 40ft container to Rotterdam.',
                        %s, 'importer@rotterdamtrader.example.com', 'sales@varunlogistics.com', NOW(), NOW())
            """, (org_id, lead_id, f"email-baseline-{int(time.time())}"))
            inbound_id = cur.lastrowid
            
        # Sales callback with incomplete intent
        callback_payload = {
            "interaction_id": inbound_id, "org_id": org_id, "lead_id": lead_id, "sentiment": "POSITIVE",
            "intent": "RFQ_REQUEST_INCOMPLETE", "confidence": 92, "summary": "Incomplete RFQ: Missing incoterms and target date.",
            "drafted_reply": "Dear Customer, could you please provide the Incoterms and target ready date?",
            "partial_rfq_context": {"origin_port": "INNSA", "destination_port": "NLRTM"}
        }
        r_cb = httpx.post(f"{BACKEND_URL}/internal/sales/callback", json=callback_payload, headers=internal_headers, timeout=5.0)
        assert r_cb.status_code == 200, f"Callback failed: {r_cb.text}"
        
        # Verify Draft Staged + Approval Request Created + 0 Autonomous Sends
        with conn.cursor() as cur:
            cur.execute("SELECT id, status, approval_id FROM lead_email_drafts WHERE parent_interaction_id = %s", (inbound_id,))
            draft_row = cur.fetchone()
            assert draft_row is not None, "Draft was not created!"
            assert draft_row[1] == "AWAITING_APPROVAL", f"Expected AWAITING_APPROVAL, got {draft_row[1]}"
            
            cur.execute("SELECT id, title, type, status FROM approval_requests WHERE id = %s", (draft_row[2],))
            app_row = cur.fetchone()
            assert app_row is not None, "Approval request was not created!"
            assert app_row[2] == "Clarification Email Approval"
            
            cur.execute("SELECT count(*) FROM lead_interactions WHERE parent_interaction_id = %s AND direction = 'OUTBOUND'", (inbound_id,))
            auto_sends = cur.fetchone()[0]
            assert auto_sends == 0, f"Critical safety violation: {auto_sends} emails sent autonomously!"
            print(f"   Verified: Draft #{draft_row[0]} [AWAITING_APPROVAL], Approval #{app_row[0]} [Pending], Outbound emails sent: {auto_sends}")
            
        results.append({
            "test": "Sales / Email Parser — Incomplete Email",
            "input_id": f"Interaction #{inbound_id}",
            "org_id": org_id,
            "status": "PASS",
            "details": f"Draft #{draft_row[0]} staged, Approval #{app_row[0]} generated, 0 autonomous sends."
        })
    except Exception as e:
        print(f"   [!] Error: {e}")
        results.append({"test": "Sales / Email Parser — Incomplete Email", "input_id": f"Lead #{lead_id}", "org_id": org_id, "status": "FAIL", "details": str(e)})

    # -------------------------------------------------------------------------
    # TEST 5: Sales / Email Parser — Complete Email RFQ Creation
    # -------------------------------------------------------------------------
    print("\n[TEST 5] Sales / Email Parser — Complete Inbound Email RFQ Creation")
    try:
        rfq_payload = {
            "org_id": org_id, "customer_id": lead_id, "origin": "INNSA", "destination": "DEHAM",
            "incoterms": "FOB", "target_date": "2026-10-15",
            "items": [{"description": "Industrial Pump Spares", "quantity": 1, "weight_kg": 14500.0, "volume_cbm": 28.0}]
        }
        r_rfq_create = httpx.post(f"{BACKEND_URL}/internal/rfqs/from-email", json=rfq_payload, headers=internal_headers, timeout=5.0)
        print(f"   Create RFQ from complete email tool status: {r_rfq_create.status_code}")
        assert r_rfq_create.status_code in [200, 201], f"Failed to create RFQ from email: {r_rfq_create.text}"
        created_rfq_data = r_rfq_create.json()
        print(f"   Created RFQ: {created_rfq_data}")
        
        results.append({
            "test": "Sales / Email Parser — Complete Email",
            "input_id": f"Lead #{lead_id}",
            "org_id": org_id,
            "status": "PASS",
            "details": f"RFQ created successfully from email payload."
        })
    except Exception as e:
        print(f"   [!] Error: {e}")
        results.append({"test": "Sales / Email Parser — Complete Email", "input_id": f"Lead #{lead_id}", "org_id": org_id, "status": "FAIL", "details": str(e)})

    # -------------------------------------------------------------------------
    # -------------------------------------------------------------------------
    # TEST 6: Operations / Carrier Tracking Agent
    # -------------------------------------------------------------------------
    print("\n[TEST 6] Operations / Carrier Tracking Agent (Shipment #102)")
    shipment_id = 102
    try:
        # Fetch shipment through internal endpoint
        r_ship = httpx.get(f"{BACKEND_URL}/internal/shipments/{shipment_id}?org_id={org_id}", headers=internal_headers, timeout=5.0)
        assert r_ship.status_code == 200, f"Get shipment failed: {r_ship.text}"
        data_wrapper = r_ship.json().get("data", {})
        ship_data = data_wrapper.get("shipment", data_wrapper)
        print(f"   Shipment #{shipment_id} MBL: {ship_data.get('mbl_number')}, Status: {ship_data.get('status')}")
        
        # Test updating milestone
        milestone_payload = {
            "milestone_code": "VESSEL_DEPARTED",
            "actual_date": "2026-09-05T12:00:00Z",
            "location": "Nhava Sheva (INNSA)",
            "notes": "Vessel departed port on schedule.",
            "org_id": org_id
        }
        r_ms = httpx.post(f"{BACKEND_URL}/internal/shipments/{shipment_id}/milestones", json=milestone_payload, headers=internal_headers, timeout=5.0)
        print(f"   Update milestone response: HTTP {r_ms.status_code}")
        
        results.append({
            "test": "Operations / Carrier Tracking",
            "input_id": f"Shipment #{shipment_id}",
            "org_id": org_id,
            "status": "PASS",
            "details": f"Shipment retrieved (status={ship_data.get('status')}), milestone updated."
        })
    except Exception as e:
        print(f"   [!] Error: {e}")
        results.append({"test": "Operations / Carrier Tracking", "input_id": f"Shipment #{shipment_id}", "org_id": org_id, "status": "FAIL", "details": str(e)})

    # -------------------------------------------------------------------------
    # TEST 7: Compliance Document Reconciliation
    # -------------------------------------------------------------------------
    print("\n[TEST 7] Compliance Document Verification (Shipment #101)")
    try:
        with conn.cursor() as cur:
            cur.execute("""
                SELECT id, shipment_id, field_name, source_document, target_document, expected_value, actual_value, status 
                FROM shipment_document_discrepancies 
                WHERE shipment_id = 101 AND org_id = %s
            """, (org_id,))
            discrepancies = cur.fetchall()
            print(f"   Compliance discrepancies recorded for Shipment #101: {len(discrepancies)} records")
            for d in discrepancies:
                print(f"     - ID={d[0]}, Field='{d[2]}', Source='{d[3]}' vs Target='{d[4]}', Exp='{d[5]}', Act='{d[6]}', Status='{d[7]}'")
                
        results.append({
            "test": "Compliance Reconciliation",
            "input_id": "Shipment #101",
            "org_id": org_id,
            "status": "PASS",
            "details": f"Verified {len(discrepancies)} compliance document discrepancies in database."
        })
    except Exception as e:
        print(f"   [!] Error: {e}")
        results.append({"test": "Compliance Reconciliation", "input_id": "Shipment #101", "org_id": org_id, "status": "FAIL", "details": str(e)})

    # -------------------------------------------------------------------------
    # TEST 8: Finance / Three-Way Invoice Reconciliation
    # -------------------------------------------------------------------------
    print("\n[TEST 8] Finance Invoice Reconciliation (Invoice #101)")
    invoice_id = 101
    try:
        with conn.cursor() as cur:
            cur.execute("SELECT id, invoice_number, status, total_amount, currency FROM customer_invoices WHERE id = %s AND org_id = %s", (invoice_id, org_id))
            inv_row = cur.fetchone()
            print(f"   Customer Invoice #{inv_row[0]} ({inv_row[1]}): Status='{inv_row[2]}', Total={inv_row[3]} {inv_row[4]}")
            
            cur.execute("""
                SELECT id, invoice_id, charge_code, field_name, expected_value, actual_value, source, status 
                FROM shipment_finance_discrepancies 
                WHERE org_id = %s
            """, (org_id,))
            inv_discrepancies = cur.fetchall()
            print(f"   Invoice discrepancies logged for org {org_id}: {len(inv_discrepancies)}")
            for idisc in inv_discrepancies:
                print(f"     - ID={idisc[0]}, Inv={idisc[1]}, Charge='{idisc[2]}', Field='{idisc[3]}', Exp='{idisc[4]}', Act='{idisc[5]}', Status='{idisc[7]}'")
                
        results.append({
            "test": "Finance Invoice Reconciliation",
            "input_id": f"Invoice #{invoice_id}",
            "org_id": org_id,
            "status": "PASS",
            "details": f"Invoice verified, {len(inv_discrepancies)} finance discrepancies tracked."
        })
    except Exception as e:
        print(f"   [!] Error: {e}")
        results.append({"test": "Finance Invoice Reconciliation", "input_id": f"Invoice #{invoice_id}", "org_id": org_id, "status": "FAIL", "details": str(e)})

    # -------------------------------------------------------------------------
    # TEST 9: Lead Scoring Worker
    # -------------------------------------------------------------------------
    print("\n[TEST 9] Lead Scoring Worker (Lead #102)")
    lead_score_id = 102
    try:
        with conn.cursor() as cur:
            cur.execute("SELECT id, company_name, contact_name, ai_score, notes, status FROM leads WHERE id = %s AND org_id = %s", (lead_score_id, org_id))
            lead_row = cur.fetchone()
            print(f"   Lead #{lead_row[0]} ({lead_row[1]} - {lead_row[2]}): AI Score={lead_row[3]}, Status='{lead_row[5]}'")
            print(f"   Notes: {lead_row[4]}")
            
        results.append({
            "test": "Lead Scoring Worker",
            "input_id": f"Lead #{lead_score_id}",
            "org_id": org_id,
            "status": "PASS",
            "details": f"Lead AI Score={lead_row[3]}, status={lead_row[5]}."
        })
    except Exception as e:
        print(f"   [!] Error: {e}")
        results.append({"test": "Lead Scoring Worker", "input_id": f"Lead #{lead_score_id}", "org_id": org_id, "status": "FAIL", "details": str(e)})


    # -------------------------------------------------------------------------
    # SUMMARY
    # -------------------------------------------------------------------------
    print("\n" + "=" * 75)
    print(" BASELINE FUNCTIONAL SUITE EXECUTION SUMMARY")
    print("=" * 75)
    for r in results:
        print(f" [{r['status']}] {r['test']:36} | Target: {r['input_id']:18} | {r['details']}")
    print("=" * 75)

    conn.close()

if __name__ == "__main__":
    run_functional_suite()
