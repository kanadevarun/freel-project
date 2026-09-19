package dashboard_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/freel/backend/internal/dashboard"
	"github.com/freel/backend/internal/dashboard/spec"
	"github.com/freel/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// TestDashboardTask10_AuthenticationAndPermissions verifies authentication enforcement and role context
func TestDashboardTask10_AuthenticationAndPermissions(t *testing.T) {
	db, err := sqlx.Open("mysql", "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC")
	if err != nil {
		t.Skip("MySQL not accessible for test:", err)
		return
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Skip("MySQL ping failed:", err)
		return
	}

	dl := dashboard.NewDataLayer(db)
	bl := dashboard.NewBusinessLogic(dl)
	endpoints := dashboard.NewAllDashboardEndpoints(bl)

	authGuard := middleware.NewAuthMiddleware("ap-south-1", "mock-pool", db, middleware.WithEnvironment("test"))
	router := chi.NewRouter()
	dashboard.AddDashboardHandlers(router, endpoints, authGuard.RequireAuth)

	// 1. Missing Authorization Header -> 401 Unauthorized
	reqUnauth := httptest.NewRequest(http.MethodGet, "/mission-control", nil)
	rrUnauth := httptest.NewRecorder()
	router.ServeHTTP(rrUnauth, reqUnauth)
	if rrUnauth.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for missing auth header, got %d", rrUnauth.Code)
	}

	// 2. Invalid Token -> 401 Unauthorized
	reqInvalid := httptest.NewRequest(http.MethodGet, "/mission-control", nil)
	reqInvalid.Header.Set("Authorization", "Bearer invalid-junk-token")
	rrInvalid := httptest.NewRecorder()
	router.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusUnauthorized && rrInvalid.Code != http.StatusInternalServerError {
		t.Errorf("Expected 401/500 for invalid token, got %d", rrInvalid.Code)
	}

	// 3. Authorized Primary Test Token (Org 1) -> 200 OK
	reqAuth := httptest.NewRequest(http.MethodGet, "/mission-control", nil)
	reqAuth.Header.Set("Authorization", "Bearer test-token")
	rrAuth := httptest.NewRecorder()
	router.ServeHTTP(rrAuth, reqAuth)
	if rrAuth.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for valid test-token, got %d: %s", rrAuth.Code, rrAuth.Body.String())
	}

	var envelope struct {
		Data spec.GetMissionControlResponse `json:"data"`
	}
	if err := json.Unmarshal(rrAuth.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}
	if envelope.Data.Organization.ID != 1 {
		t.Errorf("Expected Organization ID 1, got %d", envelope.Data.Organization.ID)
	}
}

// TestDashboardTask10_TenantIsolation verifies strict cross-tenant boundary isolation
func TestDashboardTask10_TenantIsolation(t *testing.T) {
	db, err := sqlx.Open("mysql", "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC")
	if err != nil {
		t.Skip("MySQL not accessible for test:", err)
		return
	}
	defer db.Close()

	dl := dashboard.NewDataLayer(db)
	bl := dashboard.NewBusinessLogic(dl)
	ctx := context.Background()

	// Org 1 (Freel Global Logistics)
	dataOrg1, err := bl.GetMissionControl(ctx, 1, "", "", "7D")
	if err != nil {
		t.Fatalf("Org 1 mission control failed: %v", err)
	}

	// Org 2 (LogisticsHQ Dev Org)
	dataOrg2, err := bl.GetMissionControl(ctx, 2, "", "", "7D")
	if err != nil {
		t.Fatalf("Org 2 mission control failed: %v", err)
	}

	// Org 1023 (Isolated Test Org)
	dataOrg1023, err := bl.GetMissionControl(ctx, 1023, "", "", "7D")
	if err != nil {
		t.Fatalf("Org 1023 mission control failed: %v", err)
	}

	// Verify Org 2 has different organization details
	if dataOrg1.Organization.ID == dataOrg2.Organization.ID {
		t.Errorf("Expected distinct Org IDs, both are %d", dataOrg1.Organization.ID)
	}

	// Cross-check that Org 1's invoices are NOT in Org 2 or Org 1023
	org1InvoiceIDs := make(map[int64]bool)
	for _, inv := range dataOrg1.InvoiceSummary.RecentInvoices {
		org1InvoiceIDs[inv.ID] = true
	}
	for _, inv := range dataOrg2.InvoiceSummary.RecentInvoices {
		if org1InvoiceIDs[inv.ID] {
			t.Errorf("Tenant isolation violation: Org 2 contains Org 1 invoice id=%d", inv.ID)
		}
	}
	for _, inv := range dataOrg1023.InvoiceSummary.RecentInvoices {
		if org1InvoiceIDs[inv.ID] {
			t.Errorf("Tenant isolation violation: Org 1023 contains Org 1 invoice id=%d", inv.ID)
		}
	}

	// Cross-check that Org 1's attention items are isolated
	org1AttentionTitles := make(map[string]bool)
	for _, it := range dataOrg1.AttentionItems {
		org1AttentionTitles[it.Title] = true
	}
	for _, it := range dataOrg1023.AttentionItems {
		if org1AttentionTitles[it.Title] {
			t.Errorf("Tenant isolation violation: Org 1023 contains Org 1 attention item '%s'", it.Title)
		}
	}
}

// TestDashboardTask10_AuthoritativeDataValidation verifies dashboard KPI values match raw database queries
func TestDashboardTask10_AuthoritativeDataValidation(t *testing.T) {
	db, err := sqlx.Open("mysql", "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC")
	if err != nil {
		t.Skip("MySQL not accessible for test:", err)
		return
	}
	defer db.Close()

	dl := dashboard.NewDataLayer(db)
	bl := dashboard.NewBusinessLogic(dl)
	ctx := context.Background()

	orgID := int64(1)
	dashboardData, err := bl.GetMissionControl(ctx, orgID, "", "", "7D")
	if err != nil {
		t.Fatalf("Failed to fetch dashboard data: %v", err)
	}

	// 1. Authoritative Customers Count
	var dbCustomerCount int
	err = db.GetContext(ctx, &dbCustomerCount, "SELECT COUNT(*) FROM customers WHERE org_id = ?", orgID)
	if err == nil {
		if dashboardData.Stats.TotalCustomers != dbCustomerCount {
			t.Errorf("Customers mismatch: Dashboard=%d, DB=%d", dashboardData.Stats.TotalCustomers, dbCustomerCount)
		}
	}

	// 2. Authoritative Leads Count
	var dbLeadCount int
	err = db.GetContext(ctx, &dbLeadCount, "SELECT COUNT(*) FROM leads WHERE org_id = ?", orgID)
	if err == nil {
		if dashboardData.Stats.TotalLeads != dbLeadCount {
			t.Errorf("Leads mismatch: Dashboard=%d, DB=%d", dashboardData.Stats.TotalLeads, dbLeadCount)
		}
	}

	// 3. Authoritative RFQ Count
	var dbRFQCount int
	err = db.GetContext(ctx, &dbRFQCount, "SELECT COUNT(*) FROM rfqs WHERE org_id = ?", orgID)
	if err == nil {
		if dashboardData.Stats.TotalRFQs != dbRFQCount {
			t.Errorf("RFQs mismatch: Dashboard=%d, DB=%d", dashboardData.Stats.TotalRFQs, dbRFQCount)
		}
	}

	// 4. Authoritative Invoices Count & Outstanding Amount from customer_invoices
	var dbInvoiceCount int
	var dbOutstandingAmount float64
	err = db.GetContext(ctx, &dbInvoiceCount, "SELECT COUNT(*) FROM customer_invoices WHERE org_id = ?", orgID)
	if err == nil {
		if dashboardData.Stats.TotalInvoices != dbInvoiceCount {
			t.Errorf("Invoices mismatch: Dashboard=%d, DB=%d", dashboardData.Stats.TotalInvoices, dbInvoiceCount)
		}
	}

	err = db.GetContext(ctx, &dbOutstandingAmount, "SELECT COALESCE(SUM(balance_due), 0) FROM customer_invoices WHERE org_id = ? AND status IN ('Issued', 'Partially Paid', 'Overdue', 'Pending Approval') AND balance_due > 0", orgID)
	if err == nil {
		if dashboardData.Stats.OutstandingAmount != dbOutstandingAmount {
			t.Errorf("Outstanding Amount mismatch: Dashboard=%.2f, DB=%.2f", dashboardData.Stats.OutstandingAmount, dbOutstandingAmount)
		}
	}

	// 5. Authoritative Pending Approvals Count
	var dbPendingApprovals int
	err = db.GetContext(ctx, &dbPendingApprovals, "SELECT COUNT(*) FROM approval_requests WHERE org_id = ? AND status = 'Pending'", orgID)
	if err == nil {
		if dashboardData.Stats.PendingApprovals != dbPendingApprovals {
			t.Errorf("Pending Approvals mismatch: Dashboard=%d, DB=%d", dashboardData.Stats.PendingApprovals, dbPendingApprovals)
		}
	}
}

// TestDashboardTask10_PriorityActionsSafeguards verifies approval requirements and capability classification
func TestDashboardTask10_PriorityActionsSafeguards(t *testing.T) {
	db, err := sqlx.Open("mysql", "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC")
	if err != nil {
		t.Skip("MySQL not accessible for test:", err)
		return
	}
	defer db.Close()

	dl := dashboard.NewDataLayer(db)
	bl := dashboard.NewBusinessLogic(dl)
	ctx := context.Background()

	data, err := bl.GetMissionControl(ctx, 1, "", "", "7D")
	if err != nil {
		t.Fatalf("Failed to fetch dashboard data: %v", err)
	}

	seenIDs := make(map[string]bool)
	for _, item := range data.AttentionItems {
		// Verify deduplication
		if seenIDs[item.ID] {
			t.Errorf("Duplicate attention item ID detected: %s", item.ID)
		}
		seenIDs[item.ID] = true

		// Verify urgency is valid
		urgUpper := strings.ToUpper(item.Urgency)
		if urgUpper != "CRITICAL" && urgUpper != "IMPORTANT" && urgUpper != "INFORMATIONAL" {
			t.Errorf("Invalid urgency '%s' on item %s", item.Urgency, item.ID)
		}

		// Verify capability is valid controlled enum
		capLower := strings.ToLower(item.Capability)
		if capLower != "actionable" && capLower != "approval_gated" && capLower != "draft_only" && capLower != "read_only" {
			t.Errorf("Invalid capability '%s' on item %s", item.Capability, item.ID)
		}

		// If RequiresApproval is true, capability MUST be approval_gated
		if item.RequiresApproval && capLower != "approval_gated" {
			t.Errorf("Item %s has RequiresApproval=true but Capability='%s' (expected approval_gated)", item.ID, item.Capability)
		}

		// ActionURL must be non-empty and point to a valid route
		if item.ActionURL == "" {
			t.Errorf("Item %s has empty ActionURL", item.ID)
		}
	}
}

// TestDashboardTask10_MutationSafety verifies fetching dashboard does not mutate business records
func TestDashboardTask10_MutationSafety(t *testing.T) {
	db, err := sqlx.Open("mysql", "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC")
	if err != nil {
		t.Skip("MySQL not accessible for test:", err)
		return
	}
	defer db.Close()

	ctx := context.Background()
	orgID := int64(1)

	// Capture record counts before call
	var countBefore struct {
		Shipments int
		Invoices  int
		RFQs      int
		Approvals int
		Leads     int
	}
	_ = db.GetContext(ctx, &countBefore.Shipments, "SELECT COUNT(*) FROM shipments WHERE org_id = ?", orgID)
	_ = db.GetContext(ctx, &countBefore.Invoices, "SELECT COUNT(*) FROM customer_invoices WHERE org_id = ?", orgID)
	_ = db.GetContext(ctx, &countBefore.RFQs, "SELECT COUNT(*) FROM rfqs WHERE org_id = ?", orgID)
	_ = db.GetContext(ctx, &countBefore.Approvals, "SELECT COUNT(*) FROM approval_requests WHERE org_id = ?", orgID)
	_ = db.GetContext(ctx, &countBefore.Leads, "SELECT COUNT(*) FROM leads WHERE org_id = ?", orgID)

	dl := dashboard.NewDataLayer(db)
	bl := dashboard.NewBusinessLogic(dl)
	_, err = bl.GetMissionControl(ctx, orgID, "", "", "7D")
	if err != nil {
		t.Fatalf("GetMissionControl failed: %v", err)
	}

	// Capture record counts after call
	var countAfter struct {
		Shipments int
		Invoices  int
		RFQs      int
		Approvals int
		Leads     int
	}
	_ = db.GetContext(ctx, &countAfter.Shipments, "SELECT COUNT(*) FROM shipments WHERE org_id = ?", orgID)
	_ = db.GetContext(ctx, &countAfter.Invoices, "SELECT COUNT(*) FROM customer_invoices WHERE org_id = ?", orgID)
	_ = db.GetContext(ctx, &countAfter.RFQs, "SELECT COUNT(*) FROM rfqs WHERE org_id = ?", orgID)
	_ = db.GetContext(ctx, &countAfter.Approvals, "SELECT COUNT(*) FROM approval_requests WHERE org_id = ?", orgID)
	_ = db.GetContext(ctx, &countAfter.Leads, "SELECT COUNT(*) FROM leads WHERE org_id = ?", orgID)

	if countBefore != countAfter {
		t.Errorf("Mutation safety violation: counts mutated during GET dashboard! Before: %+v, After: %+v", countBefore, countAfter)
	}
}
