package dashboard_test

import (
	"context"
	"testing"

	"github.com/freel/backend/internal/dashboard"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func TestDashboardMissionControlAggregation(t *testing.T) {
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
	ctx := context.Background()

	// 1. Test Populated / Mature Org (Org 1)
	respOrg1, err := bl.GetMissionControl(ctx, 1, "", "", "7D")
	if err != nil {
		t.Fatalf("Failed to get mission control for Org 1: %v", err)
	}
	if respOrg1 == nil {
		t.Fatal("Expected non-nil response for Org 1")
	}

	t.Logf("Org 1 Maturity: %s, IsNewUser: %v, IsOperational: %v", respOrg1.Stats.AccountMaturity, respOrg1.Stats.IsNewUser, respOrg1.Stats.IsOperational)
	t.Logf("Org 1 Stats: Customers=%d, Leads=%d, RFQs=%d, Quotations=%d, Bookings=%d, Shipments=%d, Invoices=%d, Revenue=$%.2f",
		respOrg1.Stats.TotalCustomers, respOrg1.Stats.TotalLeads, respOrg1.Stats.TotalRFQs, respOrg1.Stats.TotalQuotations,
		respOrg1.Stats.TotalBookings, respOrg1.Stats.TotalShipments, respOrg1.Stats.TotalInvoices, respOrg1.Stats.TotalRevenue)
	t.Logf("Org 1 Attention Items: %d, Active Shipments: %d, Recent Invoices: %d, Reminders: %d",
		len(respOrg1.AttentionItems), len(respOrg1.ActiveShipments), len(respOrg1.InvoiceSummary.RecentInvoices), len(respOrg1.UpcomingReminders))

	if respOrg1.Stats.IsNewUser {
		t.Errorf("Expected Org 1 to be classified as active/operational user, got IsNewUser=true")
	}
	if respOrg1.Stats.AccountMaturity != "MATURE" && respOrg1.Stats.AccountMaturity != "OPERATIONAL" {
		t.Errorf("Expected Org 1 maturity MATURE/OPERATIONAL, got %s", respOrg1.Stats.AccountMaturity)
	}
	if respOrg1.Stats.TotalCustomers == 0 {
		t.Errorf("Expected Org 1 TotalCustomers > 0, got %d", respOrg1.Stats.TotalCustomers)
	}

	// 2. Test Low-Data Org (Org 8801 - only 1 customer)
	respOrg8801, err := bl.GetMissionControl(ctx, 8801, "", "", "7D")
	if err != nil {
		t.Fatalf("Failed to get mission control for Org 8801: %v", err)
	}
	t.Logf("Org 8801 Maturity: %s, IsNewUser: %v, IsOperational: %v", respOrg8801.Stats.AccountMaturity, respOrg8801.Stats.IsNewUser, respOrg8801.Stats.IsOperational)
	if !respOrg8801.Stats.IsNewUser {
		t.Errorf("Expected Org 8801 (1 customer, no operational records) to be IsNewUser=true, got %v", respOrg8801.Stats.IsNewUser)
	}

	// 3. Test Zero-Data Org (Org 1023)
	respOrg1023, err := bl.GetMissionControl(ctx, 1023, "", "", "7D")
	if err != nil {
		t.Fatalf("Failed to get mission control for Org 1023: %v", err)
	}
	t.Logf("Org 1023 Maturity: %s, IsNewUser: %v, IsOperational: %v", respOrg1023.Stats.AccountMaturity, respOrg1023.Stats.IsNewUser, respOrg1023.Stats.IsOperational)
	if !respOrg1023.Stats.IsNewUser {
		t.Errorf("Expected Org 1023 to be IsNewUser=true, got %v", respOrg1023.Stats.IsNewUser)
	}
	if respOrg1023.Stats.AccountMaturity != "NEW" {
		t.Errorf("Expected Org 1023 maturity NEW, got %s", respOrg1023.Stats.AccountMaturity)
	}
	if respOrg1023.Stats.TotalCustomers != 0 || respOrg1023.Stats.TotalInvoices != 0 {
		t.Errorf("Expected Org 1023 to have 0 counts, got customers=%d invoices=%d", respOrg1023.Stats.TotalCustomers, respOrg1023.Stats.TotalInvoices)
	}

	// 4. Verify Tenant Isolation (Org 1023 must not see any Org 1 data)
	if len(respOrg1023.AttentionItems) != 0 {
		t.Errorf("Tenant isolation breach: Org 1023 got %d attention items", len(respOrg1023.AttentionItems))
	}
	if len(respOrg1023.ActiveShipments) != 0 {
		t.Errorf("Tenant isolation breach: Org 1023 got %d active shipments", len(respOrg1023.ActiveShipments))
	}
	if len(respOrg1023.InvoiceSummary.RecentInvoices) != 0 {
		t.Errorf("Tenant isolation breach: Org 1023 got %d recent invoices", len(respOrg1023.InvoiceSummary.RecentInvoices))
	}
}

func TestPriorityActionsStructureAndUrgency(t *testing.T) {
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

	ctx := context.Background()
	dl := dashboard.NewDataLayer(db)

	items, err := dl.GetAttentionItems(ctx, 1)
	if err != nil {
		t.Fatalf("Failed to get attention items: %v", err)
	}

	for _, it := range items {
		if it.ID == "" {
			t.Errorf("Item missing stable ID: %+v", it)
		}
		if it.Urgency != "CRITICAL" && it.Urgency != "IMPORTANT" && it.Urgency != "INFORMATIONAL" {
			t.Errorf("Invalid urgency '%s' for item: %s", it.Urgency, it.ID)
		}
		if it.Title == "" {
			t.Errorf("Item missing title: %s", it.ID)
		}
		if it.ActionURL == "" {
			t.Errorf("Item missing ActionURL: %s", it.ID)
		}
		if it.Capability != "read_only" && it.Capability != "draft_only" && it.Capability != "approval_gated" {
			t.Errorf("Invalid capability '%s' for item: %s", it.Capability, it.ID)
		}
	}
}

func TestOperationsFinanceActivityOverview(t *testing.T) {
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

	ctx := context.Background()
	dl := dashboard.NewDataLayer(db)

	// 1. Operations: Active Shipments
	shipments, _, err := dl.GetActiveShipments(ctx, 2)
	if err != nil {
		t.Fatalf("Failed to fetch active shipments for Org 2: %v", err)
	}
	t.Logf("Org 2 Active Shipments count: %d", len(shipments))
	for _, s := range shipments {
		if s.ShipmentNo == "" {
			t.Errorf("Shipment missing ShipmentNo: %+v", s)
		}
		if s.Status == "" {
			t.Errorf("Shipment missing Status: %+v", s)
		}
	}

	// 2. Approvals: Pending Approval Queue
	_, approvals, err := dl.GetApprovalQueue(ctx, 2)
	if err != nil {
		t.Fatalf("Failed to fetch approval queue for Org 2: %v", err)
	}
	t.Logf("Org 2 Pending Approvals count: %d", len(approvals))
	for _, a := range approvals {
		if a.Title == "" && a.RequestCode == "" {
			t.Errorf("Approval item missing Title/RequestCode: %+v", a)
		}
	}

	// 3. Activity: Recent Business Activity
	activity, err := dl.GetRecentActivity(ctx, 2)
	if err != nil {
		t.Fatalf("Failed to fetch recent activity for Org 2: %v", err)
	}
	t.Logf("Org 2 Recent Activity items: %d", len(activity))
	for _, act := range activity {
		if act.Title == "" {
			t.Errorf("Activity item missing title: %+v", act)
		}
		if act.Type == "" {
			t.Errorf("Activity item missing type: %+v", act)
		}
	}

	// 4. Reminders: Upcoming Reminders
	reminders, err := dl.GetUpcomingReminders(ctx, 2)
	if err != nil {
		t.Fatalf("Failed to fetch upcoming reminders for Org 2: %v", err)
	}
	t.Logf("Org 2 Upcoming Reminders count: %d", len(reminders))

	// 5. Tenant Isolation Test (Org 1023 has zero records)
	emptyShipments, _, _ := dl.GetActiveShipments(ctx, 1023)
	if len(emptyShipments) != 0 {
		t.Errorf("Tenant isolation failure: Org 1023 has active shipments")
	}
	_, emptyApprovals, _ := dl.GetApprovalQueue(ctx, 1023)
	if len(emptyApprovals) != 0 {
		t.Errorf("Tenant isolation failure: Org 1023 has approvals")
	}
	emptyActivity, _ := dl.GetRecentActivity(ctx, 1023)
	if len(emptyActivity) != 0 {
		t.Errorf("Tenant isolation failure: Org 1023 has activity")
	}
}

func TestDashboardDeduplicationAndDensity(t *testing.T) {
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

	ctx := context.Background()
	dl := dashboard.NewDataLayer(db)
	bl := dashboard.NewBusinessLogic(dl)

	resp, err := bl.GetMissionControl(ctx, 2, "", "", "7D")
	if err != nil {
		t.Fatalf("Failed to fetch mission control: %v", err)
	}

	// 1. Verify that no Upcoming Reminder duplicates an entity featured in Priority Actions
	claimedEntities := make(map[string]bool)
	for _, item := range resp.AttentionItems {
		if item.SourceEntityID > 0 {
			claimedEntities[item.ID] = true
		}
	}

	for _, rem := range resp.UpcomingReminders {
		for _, item := range resp.AttentionItems {
			if item.SourceEntityID > 0 && rem.ActionURL == item.ActionURL && rem.Title == item.Title {
				t.Errorf("Duplicate item found across Priority Actions and Reminders: %s", rem.Title)
			}
		}
	}

	// 2. Verify that Recent Activity items have unique IDs
	seenActivity := make(map[string]bool)
	for _, act := range resp.RecentActivity {
		if seenActivity[act.ID] {
			t.Errorf("Duplicate activity ID detected: %s", act.ID)
		}
		seenActivity[act.ID] = true
	}

	// 3. Verify maximum density limits (reminders <= 3, activity <= 8)
	if len(resp.UpcomingReminders) > 3 {
		t.Errorf("Expected at most 3 deduplicated upcoming reminders, got %d", len(resp.UpcomingReminders))
	}
	if len(resp.RecentActivity) > 8 {
		t.Errorf("Expected at most 8 recent activity events, got %d", len(resp.RecentActivity))
	}
}

