import httpx
import json

def verify_all_modules():
    print("=== [PHASE 8] VERIFYING ALL FRONTEND & BACKEND MODULE ENDPOINTS ===")
    
    login_resp = httpx.post("http://localhost:8080/auth/login", json={
        "email": "kanadevarun123@gmail.com",
        "password": "Varun@123"
    }, timeout=5.0)
    
    assert login_resp.status_code == 200, f"Login failed: {login_resp.status_code}"
    token = login_resp.json()["data"]["access_token"]
    headers = {"Authorization": f"Bearer {token}"}

    endpoints = [
        ("Dashboard Mission Control", "GET", "/api/v1/dashboard/mission-control?preset=LAST_7D"),
        ("Leads List", "GET", "/api/v1/leads"),
        ("RFQs List", "GET", "/api/v1/rfqs"),
        ("Quotations List", "GET", "/api/v1/quotations"),
        ("Bookings List", "GET", "/api/v1/bookings"),
        ("Shipments List", "GET", "/api/v1/shipments"),
        ("Tracking Alerts", "GET", "/api/v1/shipments/101/tracking/alerts"),
        ("Shipment Documents", "GET", "/api/v1/shipments/101/documents"),
        ("Contract Documents", "GET", "/api/v1/contracts"),
        ("Contract Review Queue", "GET", "/api/v1/contracts/review"),
        ("Invoices List", "GET", "/api/v1/invoices"),
        ("Approvals List", "GET", "/api/v1/approvals"),
        ("Approvals Stats", "GET", "/api/v1/approvals/stats"),
        ("Audit Logs", "GET", "/api/v1/audit-logs?page=1&limit=10"),
    ]

    results = []
    for name, method, path in endpoints:
        url = f"http://localhost:8080{path}"
        resp = httpx.request(method, url, headers=headers, timeout=5.0)
        data = resp.json() if resp.status_code == 200 else {}
        count = None
        if isinstance(data.get("data"), list):
            count = len(data["data"])
        elif isinstance(data.get("data"), dict):
            count = len(data["data"].keys())
        
        status_str = "PASS" if resp.status_code == 200 else f"FAIL ({resp.status_code})"
        print(f"[{status_str}] {name:25} -> {path:55} | count/keys: {count}")
        results.append((name, path, resp.status_code, status_str, count))

    return results

if __name__ == '__main__':
    verify_all_modules()
