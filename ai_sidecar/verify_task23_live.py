import httpx
import json
import sys

def run_live_verification():
    print("=== LIVE VERIFICATION: TASK 2.3 RFQ & QUOTATION WORKFLOW ASSISTANT ===")
    
    # 1. Login
    login_res = httpx.post("http://localhost:8080/auth/login", json={
        "email": "kanadevarun123@gmail.com",
        "password": "Varun@123"
    }, timeout=5.0)
    
    if login_res.status_code != 200:
        print(f"Login failed: {login_res.status_code} - {login_res.text}")
        sys.exit(1)
        
    token = login_res.json()["data"]["access_token"]
    headers = {"Authorization": f"Bearer {token}"}
    print("[PASS] Successfully authenticated with LogisticsHQ backend.")
    
    # 2. Trigger Deterministic Generation
    gen_res = httpx.post("http://localhost:8080/api/v1/recommendations/generate", headers=headers, timeout=10.0)
    assert gen_res.status_code == 200, f"Generate failed: {gen_res.status_code} {gen_res.text}"
    gen_data = gen_res.json().get("result", {})
    print(f"[PASS] Generation cycle complete. Evaluated {gen_data.get('total_evaluated')} records (created: {gen_data.get('created_count')}, updated: {gen_data.get('updated_count')}).")
    
    # 3. List RFQ recommendations
    rfq_res = httpx.get("http://localhost:8080/api/v1/recommendations?category=rfq", headers=headers, timeout=5.0)
    assert rfq_res.status_code == 200, f"List RFQ recs failed: {rfq_res.status_code}"
    rfq_recs = rfq_res.json().get("recommendations", [])
    print(f"[PASS] Retrieved {len(rfq_recs)} RFQ recommendations.")
    
    # 4. List Quotation recommendations
    quote_res = httpx.get("http://localhost:8080/api/v1/recommendations?category=quotation", headers=headers, timeout=5.0)
    assert quote_res.status_code == 200, f"List Quote recs failed: {quote_res.status_code}"
    quote_recs = quote_res.json().get("recommendations", [])
    print(f"[PASS] Retrieved {len(quote_recs)} Quotation recommendations.")
    
    # 5. List Missing Info filter
    missing_res = httpx.get("http://localhost:8080/api/v1/recommendations?missing_info=true", headers=headers, timeout=5.0)
    assert missing_res.status_code == 200
    missing_recs = missing_res.json().get("recommendations", [])
    print(f"[PASS] Filter missing_info=true returned {len(missing_recs)} recommendations.")
    
    # Verify RFQ recommendation details and evidence
    assert len(rfq_recs) > 0, "Expected at least 1 RFQ recommendation"
    sample_rfq_rec = rfq_recs[0]
    rec_id = sample_rfq_rec["id"]
    print(f"Testing RFQ recommendation #{rec_id} (Ref: {sample_rfq_rec.get('source_reference')}, Rule: {sample_rfq_rec.get('rule_applied')})")
    
    # 6. Action Preview
    preview_res = httpx.get(f"http://localhost:8080/api/v1/recommendations/{rec_id}/action-preview", headers=headers, timeout=5.0)
    assert preview_res.status_code == 200, f"Action preview failed: {preview_res.status_code} {preview_res.text}"
    preview_data = preview_res.json().get("action_preview", {})
    assert preview_data.get("recommendation_id") == rec_id
    assert preview_data.get("proposed_action") != ""
    assert preview_data.get("expected_effect") != ""
    print(f"[PASS] Action Preview verified: Proposed '{preview_data.get('proposed_action')[:40]}...', Effect '{preview_data.get('expected_effect')[:40]}...'")
    
    # 7. Explicit Draft Generation
    draft_res = httpx.post(f"http://localhost:8080/api/v1/recommendations/{rec_id}/draft", headers=headers, timeout=5.0)
    assert draft_res.status_code == 200, f"Draft generation failed: {draft_res.status_code}"
    draft_data = draft_res.json().get("recommendation", {})
    assert draft_data.get("draft_subject") is not None
    assert draft_data.get("draft_body") is not None
    print(f"[PASS] Draft synthesized: Subject '{draft_data.get('draft_subject')}'")
    
    # 8. Save Draft
    save_res = httpx.patch(f"http://localhost:8080/api/v1/recommendations/{rec_id}/draft", headers=headers, json={
        "subject": draft_data["draft_subject"] + " (Operator Reviewed)",
        "body": draft_data["draft_body"] + "\n\nNote: Verified with ops dispatch."
    }, timeout=5.0)
    assert save_res.status_code == 200, f"Save draft failed: {save_res.status_code}"
    print("[PASS] Draft saved with operator edits.")
    
    # 9. Review Recommendation
    rev_res = httpx.post(f"http://localhost:8080/api/v1/recommendations/{rec_id}/review", headers=headers, timeout=5.0)
    assert rev_res.status_code == 200
    assert rev_res.json().get("recommendation", {}).get("status") == "reviewed"
    print(f"[PASS] Recommendation #{rec_id} marked as reviewed.")
    
    # 10. Test HITL Approval Integration on a Quotation recommendation
    if len(quote_recs) > 0:
        quote_rec = quote_recs[0]
        q_id = quote_rec["id"]
        app_res = httpx.post(f"http://localhost:8080/api/v1/recommendations/{q_id}/request-approval", headers=headers, json={
            "notes": "Verified commercial terms. Requesting commercial lead signoff."
        }, timeout=5.0)
        assert app_res.status_code == 200, f"Request approval failed: {app_res.status_code} {app_res.text}"
        app_data = app_res.json().get("recommendation", {})
        assert app_data.get("approval_id") is not None
        print(f"[PASS] Quotation recommendation #{q_id} submitted to HITL approvals. Linked ApprovalRequest #{app_data.get('approval_id')}.")
    
    # 11. Deduplication Check
    gen_res2 = httpx.post("http://localhost:8080/api/v1/recommendations/generate", headers=headers, timeout=10.0)
    assert gen_res2.status_code == 200
    gen_data2 = gen_res2.json().get("result", {})
    assert gen_data2.get("created_count") == 0, f"Deduplication failed! Created {gen_data2.get('created_count')} duplicate records."
    print(f"[PASS] Deduplication confirmed: 0 duplicate records created on second run.")
    
    print("\n=======================================================")
    print("ALL LIVE END-TO-END VERIFICATION CHECKS PASSED (100%)")
    print("=======================================================")

if __name__ == '__main__':
    run_live_verification()
