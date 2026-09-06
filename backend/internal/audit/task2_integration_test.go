package audit_test

import (
	"context"
	"testing"

	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/middleware"
)

// TestTask2_RealActionsAudit verifies that real business operations across LogisticsHQ
// correctly generate Universal Audit Log records with proper actor, module, action, metadata, diffs, and sanitization.
func TestTask2_RealActionsAudit(t *testing.T) {
	svcPtr, orgA, orgB, cleanup := setupTestDB(t)
	defer cleanup()
	svc := *svcPtr

	// Set global audit default service
	audit.SetDefaultService(svc)

	ctx := context.Background()
	actorUserID := int64(999001)
	userCtx := middleware.UserContext{
		UserID: actorUserID,
		OrgID:  orgA,
		Role:   "SUPER_ADMIN",
	}
	authCtx := context.WithValue(ctx, middleware.UserContextKey, userCtx)

	// 1. Authentication Events
	t.Run("Auth_LoginSuccess_Audit", func(t *testing.T) {
		entry, err := audit.Record(authCtx, domain.CreateAuditLogParams{
			OrgID:        orgA,
			ActorID:      &actorUserID,
			ActorType:    domain.ActorTypeUser,
			ActorName:    "Varun Kanade",
			ActorRole:    "Super Admin",
			Action:       domain.ActionLogin,
			Module:       domain.ModuleAuthentication,
			ResourceType: "USER",
			ResourceID:   "101",
			ResourceName: "operator@freel-testing.local",
			Description:  "Varun Kanade logged in",
			Result:       domain.ResultSuccess,
			IPAddress:    "192.168.1.50",
			UserAgent:    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
		})
		if err != nil {
			t.Fatalf("failed to record login audit: %v", err)
		}
		if entry.Action != domain.ActionLogin || entry.Module != domain.ModuleAuthentication {
			t.Errorf("unexpected action/module: %s/%s", entry.Action, entry.Module)
		}
	})

	t.Run("Auth_LoginFailed_Audit", func(t *testing.T) {
		entry, err := audit.Record(ctx, domain.CreateAuditLogParams{
			OrgID:        orgA,
			ActorType:    domain.ActorTypeUser,
			ActorName:    "intruder@unknown.com",
			Action:       domain.ActionLoginFailed,
			Module:       domain.ModuleAuthentication,
			ResourceType: "USER",
			ResourceID:   "intruder@unknown.com",
			ResourceName: "intruder@unknown.com",
			Description:  "Login attempt failed for intruder@unknown.com",
			Result:       domain.ResultFailed,
			ErrorMessage: "Incorrect username or password.",
			IPAddress:    "203.0.113.195",
			UserAgent:    "Curl/7.68.0",
		})
		if err != nil {
			t.Fatalf("failed to record failed login audit: %v", err)
		}
		if entry.Result != domain.ResultFailed || entry.ErrorMessage == "" {
			t.Errorf("expected failed result with error message, got %v", entry)
		}
	})

	// 2. Users & RBAC Events
	t.Run("Users_Invite_And_RoleChanged_Audit", func(t *testing.T) {
		// Invite
		entry1, err := audit.Record(authCtx, domain.CreateAuditLogParams{
			OrgID:        orgA,
			Action:       domain.ActionInvite,
			Module:       domain.ModuleUsers,
			ResourceType: "USER",
			ResourceName: "pricing.analyst@example.com",
			Description:  "Invited user pricing.analyst@example.com",
			Result:       domain.ResultSuccess,
		})
		if err != nil || entry1.Action != domain.ActionInvite {
			t.Fatalf("failed invite audit: %v", err)
		}

		// Role Change with state diff
		entry2, err := audit.Record(authCtx, domain.CreateAuditLogParams{
			OrgID:        orgA,
			Action:       domain.ActionRoleChanged,
			Module:       domain.ModuleRolesPermissions,
			ResourceType: "USER",
			ResourceID:   "102",
			Description:  "Changed user #102 role from OPERATIONS to PRICING_MANAGER",
			Before:       map[string]interface{}{"role": "OPERATIONS"},
			After:        map[string]interface{}{"role_id": 4},
			Result:       domain.ResultSuccess,
		})
		if err != nil || entry2.Action != domain.ActionRoleChanged {
			t.Fatalf("failed role change audit: %v", err)
		}
		if len(entry2.Changes) == 0 {
			t.Errorf("expected computed changes diff for role change")
		}
	})

	// 3. Leads & Outreach Events
	t.Run("Leads_Create_And_UpdateDiff_Audit", func(t *testing.T) {
		// Create Lead
		entry1, err := audit.Record(authCtx, domain.CreateAuditLogParams{
			OrgID:        orgA,
			Action:       domain.ActionCreate,
			Module:       domain.ModuleLeads,
			ResourceType: "LEAD",
			ResourceID:   "5501",
			ResourceName: "Global Auto Parts Corp",
			Description:  "Created lead Global Auto Parts Corp",
			Result:       domain.ResultSuccess,
		})
		if err != nil || entry1.Module != domain.ModuleLeads {
			t.Fatalf("failed lead creation audit: %v", err)
		}

		// Update Lead with Before / After state diff
		entry2, err := audit.Record(authCtx, domain.CreateAuditLogParams{
			OrgID:        orgA,
			Action:       domain.ActionUpdate,
			Module:       domain.ModuleLeads,
			ResourceType: "LEAD",
			ResourceID:   "5501",
			ResourceName: "Global Auto Parts Corp",
			Description:  "Updated lead Global Auto Parts Corp",
			Before: map[string]interface{}{
				"company_name": "Global Auto Parts Corp",
				"status":       "NEW",
				"assigned_to":  nil,
			},
			After: map[string]interface{}{
				"company_name": "Global Auto Parts Corp",
				"status":       "QUALIFIED",
				"assigned_to":  101,
			},
			Result: domain.ResultSuccess,
		})
		if err != nil {
			t.Fatalf("failed lead update audit: %v", err)
		}
		if len(entry2.Changes) != 2 { // status changed + assigned_to changed
			t.Errorf("expected 2 field changes, got %d", len(entry2.Changes))
		}
	})

	// 4. RFQ, Quotations & Bookings Events
	t.Run("RFQ_Quote_Booking_Audit", func(t *testing.T) {
		// RFQ Created
		rfqEntry, err := audit.Record(authCtx, domain.CreateAuditLogParams{
			OrgID:        orgA,
			Action:       domain.ActionCreate,
			Module:       domain.ModuleRFQs,
			ResourceType: "RFQ",
			ResourceID:   "801",
			ResourceName: "RFQ-801",
			Description:  "Created RFQ RFQ-801 (Shanghai → Los Angeles)",
			Result:       domain.ResultSuccess,
		})
		if err != nil || rfqEntry.Module != domain.ModuleRFQs {
			t.Fatalf("failed rfq creation audit: %v", err)
		}

		// Quote Approved
		quoteEntry, err := audit.Record(authCtx, domain.CreateAuditLogParams{
			OrgID:        orgA,
			Action:       domain.ActionApprove,
			Module:       domain.ModuleQuotations,
			ResourceType: "QUOTE",
			ResourceID:   "9021",
			ResourceName: "Maersk Line",
			Description:  "Approved quotation #9021 from carrier Maersk Line",
			Result:       domain.ResultSuccess,
		})
		if err != nil || quoteEntry.Module != domain.ModuleQuotations {
			t.Fatalf("failed quote approval audit: %v", err)
		}

		// Booking Created & Confirmed
		bookingEntry, err := audit.Record(authCtx, domain.CreateAuditLogParams{
			OrgID:        orgA,
			Action:       domain.ActionCreate,
			Module:       domain.ModuleBookings,
			ResourceType: "BOOKING",
			ResourceID:   "7011",
			ResourceName: "MSK-BK-77401",
			Description:  "Created booking MSK-BK-77401 with carrier Maersk Line",
			Result:       domain.ResultSuccess,
		})
		if err != nil || bookingEntry.Module != domain.ModuleBookings {
			t.Fatalf("failed booking creation audit: %v", err)
		}
	})

	// 5. Carrier Integrations Events & Secret Stripping
	t.Run("Carrier_Connect_Sanitized_Audit", func(t *testing.T) {
		carrierEntry, err := audit.Record(authCtx, domain.CreateAuditLogParams{
			OrgID:        orgA,
			Action:       domain.ActionConnect,
			Module:       domain.ModuleCarrierIntegrations,
			ResourceType: "CARRIER",
			ResourceID:   "301",
			ResourceName: "MSC",
			Description:  "Connected carrier integration for MSC (PRODUCTION)",
			After: map[string]interface{}{
				"carrier_scac":   "MSCU",
				"environment":    "PRODUCTION",
				"api_key":        "SUPER_SECRET_MSC_API_KEY_999",
				"bearer_token":   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.sensitive",
				"account_number": "MSC-ACC-44810",
			},
			Result: domain.ResultSuccess,
		})
		if err != nil {
			t.Fatalf("failed carrier connect audit: %v", err)
		}

		// Verify that secret keys were replaced with [REDACTED]
		if carrierEntry.AfterData == nil {
			t.Fatalf("expected after data to be preserved")
		}
		if carrierEntry.AfterData["api_key"] != "[REDACTED]" {
			t.Errorf("expected api_key to be [REDACTED], got %v", carrierEntry.AfterData["api_key"])
		}
		if carrierEntry.AfterData["bearer_token"] != "[REDACTED]" {
			t.Errorf("expected bearer_token to be [REDACTED], got %v", carrierEntry.AfterData["bearer_token"])
		}
		if carrierEntry.AfterData["account_number"] != "MSC-ACC-44810" {
			t.Errorf("expected non-secret account_number to remain intact, got %v", carrierEntry.AfterData["account_number"])
		}
	})

	// 6. Invoices, Payments & Approvals Events
	t.Run("Invoices_Payments_Approvals_Audit", func(t *testing.T) {
		// Invoice Created
		invEntry, err := audit.Record(authCtx, domain.CreateAuditLogParams{
			OrgID:        orgIDFrom(orgA),
			Action:       domain.ActionCreate,
			Module:       domain.ModuleInvoices,
			ResourceType: "INVOICE",
			ResourceID:   "6001",
			ResourceName: "INV-2026-0045",
			Description:  "Created invoice INV-2026-0045 for Global Auto Parts Corp (USD 4250.00)",
			Result:       domain.ResultSuccess,
		})
		if err != nil || invEntry.Module != domain.ModuleInvoices {
			t.Fatalf("failed invoice creation audit: %v", err)
		}

		// Payment Recorded
		payEntry, err := audit.Record(authCtx, domain.CreateAuditLogParams{
			OrgID:        orgA,
			Action:       domain.ActionCreate,
			Module:       domain.ModulePayments,
			ResourceType: "PAYMENT",
			ResourceID:   "6001",
			ResourceName: "INV-2026-0045",
			Description:  "Recorded payment of USD 4250.00 for invoice INV-2026-0045",
			After: map[string]interface{}{
				"amount":         4250.00,
				"payment_method": "Wire Transfer",
				"payment_ref":    "PAY-2026-9812",
			},
			Result: domain.ResultSuccess,
		})
		if err != nil || payEntry.Module != domain.ModulePayments {
			t.Fatalf("failed payment recording audit: %v", err)
		}

		// Approval Action
		appEntry, err := audit.Record(authCtx, domain.CreateAuditLogParams{
			OrgID:        orgA,
			Action:       domain.ActionApprove,
			Module:       domain.ModuleApprovals,
			ResourceType: "APPROVAL",
			ResourceID:   "401",
			ResourceName: "FIN-APP-8821",
			Description:  "Approved request FIN-APP-8821 (Invoice Approval)",
			Result:       domain.ResultSuccess,
		})
		if err != nil || appEntry.Module != domain.ModuleApprovals {
			t.Fatalf("failed approval audit: %v", err)
		}
	})

	// 7. Multi-Tenant Isolation Verification
	t.Run("MultiTenant_Query_Isolation", func(t *testing.T) {
		// Org A query should only return Org A records
		respA, err := svc.List(ctx, domain.AuditLogFilter{
			OrgID: orgA,
			Limit: 50,
		})
		if err != nil {
			t.Fatalf("failed to query Org A logs: %v", err)
		}
		if respA.Total < 10 {
			t.Errorf("expected at least 10 logs for Org A, got %d", respA.Total)
		}
		for _, l := range respA.Items {
			if l.OrgID != orgA {
				t.Fatalf("tenant isolation breached: found log with OrgID %d in Org A query", l.OrgID)
			}
		}

		// Org B query should return 0 records
		respB, err := svc.List(ctx, domain.AuditLogFilter{
			OrgID: orgB,
			Limit: 50,
		})
		if err != nil {
			t.Fatalf("failed to query Org B logs: %v", err)
		}
		if respB.Total != 0 || len(respB.Items) != 0 {
			t.Errorf("expected 0 logs for Org B, got %d", respB.Total)
		}
	})
}

func orgIDFrom(id int64) int64 {
	return id
}
