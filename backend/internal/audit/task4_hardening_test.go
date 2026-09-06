package audit_test

import (
	"context"
	"testing"

	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/middleware"
)

// TestTask4_ProductionHardeningAndSecurity performs the final validation and security review
// of the universal audit logs system.
func TestTask4_ProductionHardeningAndSecurity(t *testing.T) {
	svcPtr, orgA, orgB, cleanup := setupTestDB(t)
	defer cleanup()
	svc := *svcPtr
	audit.SetDefaultService(svc)

	ctx := context.Background()
	adminUserID := int64(999001)
	userCtxA := middleware.UserContext{
		UserID: adminUserID,
		OrgID:  orgA,
		Role:   "SUPER_ADMIN",
	}
	authCtxA := context.WithValue(ctx, middleware.UserContextKey, userCtxA)

	// ── 1. Tenant Isolation & ID Spoofing Protection ──────────────────────────
	t.Run("TenantIsolation_StrictEnforcement", func(t *testing.T) {
		// Create an event for Org A
		entryA, err := audit.Record(authCtxA, domain.CreateAuditLogParams{
			OrgID:        orgA,
			ActorID:      &adminUserID,
			Action:       domain.ActionCreate,
			Module:       domain.ModuleQuotations,
			ResourceType: "QUOTATION",
			ResourceID:   "QT-SECRET-9988",
			ResourceName: "Confidential Pricing Quotation",
			Description:  "Created high value confidential quotation",
			Result:       domain.ResultSuccess,
			Metadata: map[string]interface{}{
				"client_name": "Org A VIP Client",
				"total_value": 250000,
			},
		})
		if err != nil {
			t.Fatalf("Failed to create Org A audit log: %v", err)
		}

		// Verify Org A can fetch it
		fetchedA, err := svc.GetByID(ctx, orgA, entryA.ID)
		if err != nil || fetchedA == nil {
			t.Fatalf("Org A failed to retrieve its own audit log: %v", err)
		}

		// Security Check: Org B MUST NOT be able to access Org A's audit log by ID
		fetchedB, err := svc.GetByID(ctx, orgB, entryA.ID)
		if err == nil && fetchedB != nil {
			t.Fatalf("CRITICAL SECURITY VULNERABILITY: Org B was able to read Org A's audit log ID %d!", entryA.ID)
		}

		// Security Check: Org B searching Org A's confidential record ID or client name must return 0 results
		resB, err := svc.List(ctx, domain.AuditLogFilter{
			OrgID:  orgB,
			Search: "QT-SECRET-9988",
			Page:   1,
			Limit:  20,
		})
		if err != nil {
			t.Fatalf("Org B search query failed: %v", err)
		}
		if resB.Total != 0 || len(resB.Items) != 0 {
			t.Fatalf("CRITICAL SECURITY VULNERABILITY: Org B search leaked Org A's record! Total=%d", resB.Total)
		}
	})

	// ── 2. Comprehensive Secret Sanitization ──────────────────────────────────
	t.Run("SecretSanitization_AllSecretTypes", func(t *testing.T) {
		rawSecrets := map[string]interface{}{
			"password":            "SuperSecretPassword123!",
			"password_hash":       "$2a$12$e8uqkjsdfh234789sdjhfksdj",
			"api_key":             "live_carrier_api_key_xyz",
			"api_secret":          "carrier_secret_signature_abc",
			"access_token":        "bearer_token_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			"refresh_token":       "refresh_token_abc987654321",
			"webhook_secret":      "whsec_3498573498579384759384",
			"card_number":         "4111222233334444",
			"cvv":                 "123",
			"bank_account_number": "987654321000",
			"routing_number":      "122000496",
			"client_name":         "Acme Corporation", // safe field
			"account_number":      "CARRIER-ACC-9900", // safe carrier account identifier
		}

		entry, err := audit.Record(authCtxA, domain.CreateAuditLogParams{
			OrgID:        orgA,
			ActorID:      &adminUserID,
			Action:       domain.ActionConnect,
			Module:       domain.ModuleCarrierIntegrations,
			ResourceType: "CARRIER",
			ResourceID:   "MAEU",
			Description:  "Connected Maersk Carrier Integration",
			Result:       domain.ResultSuccess,
			Before:       rawSecrets,
			After:        rawSecrets,
			Metadata:     rawSecrets,
		})
		if err != nil {
			t.Fatalf("Failed to create audit log with sensitive payload: %v", err)
		}

		fetched, err := svc.GetByID(ctx, orgA, entry.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve audit log: %v", err)
		}

		// Verify BeforeData, AfterData, and Metadata redactions
		for _, targetMap := range []map[string]interface{}{fetched.BeforeData, fetched.AfterData, fetched.Metadata} {
			if targetMap["password"] != "[REDACTED]" {
				t.Errorf("expected password to be [REDACTED], got %v", targetMap["password"])
			}
			if targetMap["password_hash"] != "[REDACTED]" {
				t.Errorf("expected password_hash to be [REDACTED], got %v", targetMap["password_hash"])
			}
			if targetMap["api_key"] != "[REDACTED]" {
				t.Errorf("expected api_key to be [REDACTED], got %v", targetMap["api_key"])
			}
			if targetMap["api_secret"] != "[REDACTED]" {
				t.Errorf("expected api_secret to be [REDACTED], got %v", targetMap["api_secret"])
			}
			if targetMap["access_token"] != "[REDACTED]" {
				t.Errorf("expected access_token to be [REDACTED], got %v", targetMap["access_token"])
			}
			if targetMap["webhook_secret"] != "[REDACTED]" {
				t.Errorf("expected webhook_secret to be [REDACTED], got %v", targetMap["webhook_secret"])
			}
			if targetMap["card_number"] != "[REDACTED]" {
				t.Errorf("expected card_number to be [REDACTED], got %v", targetMap["card_number"])
			}
			if targetMap["bank_account_number"] != "[REDACTED]" {
				t.Errorf("expected bank_account_number to be [REDACTED], got %v", targetMap["bank_account_number"])
			}
			// Verify safe business identifiers are preserved
			if targetMap["client_name"] != "Acme Corporation" {
				t.Errorf("expected client_name preserved, got %v", targetMap["client_name"])
			}
			if targetMap["account_number"] != "CARRIER-ACC-9900" {
				t.Errorf("expected account_number preserved, got %v", targetMap["account_number"])
			}
		}
	})

	// ── 3. Background / System Actor Attribution ──────────────────────────────
	t.Run("SystemActor_BackgroundAttribution", func(t *testing.T) {
		entry, err := audit.Record(ctx, domain.CreateAuditLogParams{
			OrgID:        orgA,
			ActorType:    domain.ActorTypeSystem,
			ActorName:    "Carrier Tracking Engine",
			Action:       domain.ActionSync,
			Module:       domain.ModuleTracking,
			ResourceType: "TRACKING",
			ResourceID:   "MSKU-9988112",
			Description:  "Automated tracking sync completed with 2 milestone updates",
			Result:       domain.ResultSuccess,
			Metadata: map[string]interface{}{
				"container_no": "MSKU-9988112",
				"milestones":   2,
			},
		})
		if err != nil {
			t.Fatalf("Failed to create system audit log: %v", err)
		}

		if entry.ActorType != domain.ActorTypeSystem {
			t.Errorf("expected ActorType SYSTEM, got %s", entry.ActorType)
		}
		if entry.ActorID != nil {
			t.Errorf("expected nil ActorID for system event, got %v", entry.ActorID)
		}
	})

	// ── 4. Failed Business Operation Outcome Recording ────────────────────────
	t.Run("FailedOperation_ResultAndErrorCaptured", func(t *testing.T) {
		failMsg := "Credit limit exceeded ($150,000 > $100,000)"
		entry, err := audit.Record(authCtxA, domain.CreateAuditLogParams{
			OrgID:        orgA,
			ActorID:      &adminUserID,
			Action:       domain.ActionCreate,
			Module:       domain.ModuleBookings,
			ResourceType: "BOOKING",
			ResourceID:   "BKG-REJECTED-01",
			Description:  "Booking creation rejected: credit limit exceeded",
			Result:       domain.ResultFailed,
			ErrorMessage: failMsg,
		})
		if err != nil {
			t.Fatalf("Failed to record failed operation: %v", err)
		}

		if entry.Result != domain.ResultFailed {
			t.Errorf("expected result FAILED, got %s", entry.Result)
		}
		if entry.ErrorMessage != failMsg {
			t.Errorf("expected error message '%s', got '%s'", failMsg, entry.ErrorMessage)
		}
	})

	// ── 5. Field-Level Diffs (Before vs After) ─────────────────────────────────
	t.Run("FieldLevelDiffs_AccurateCalculation", func(t *testing.T) {
		before := map[string]interface{}{
			"status": "DRAFT",
			"eta":    "2026-09-10",
			"notes":  "Initial draft",
		}
		after := map[string]interface{}{
			"status": "CONFIRMED",
			"eta":    "2026-09-08",
			"notes":  "Initial draft", // unchanged
		}

		entry, err := audit.Record(authCtxA, domain.CreateAuditLogParams{
			OrgID:        orgA,
			ActorID:      &adminUserID,
			Action:       domain.ActionUpdate,
			Module:       domain.ModuleShipments,
			ResourceType: "SHIPMENT",
			ResourceID:   "SHP-9001",
			Description:  "Updated shipment status and ETA",
			Result:       domain.ResultSuccess,
			Before:       before,
			After:        after,
		})
		if err != nil {
			t.Fatalf("Failed to record update with diff: %v", err)
		}

		if len(entry.Changes) != 2 {
			t.Fatalf("expected exactly 2 changed fields, got %d", len(entry.Changes))
		}

		changedFields := make(map[string]domain.FieldChange)
		for _, ch := range entry.Changes {
			changedFields[ch.Field] = ch
		}

		if statusCh, ok := changedFields["status"]; !ok {
			t.Errorf("missing status change")
		} else {
			if statusCh.Before != "DRAFT" || statusCh.After != "CONFIRMED" {
				t.Errorf("unexpected status diff: %+v", statusCh)
			}
		}

		if etaCh, ok := changedFields["eta"]; !ok {
			t.Errorf("missing eta change")
		} else {
			if etaCh.Before != "2026-09-10" || etaCh.After != "2026-09-08" {
				t.Errorf("unexpected eta diff: %+v", etaCh)
			}
		}
	})

	// ── 6. Pagination, Date Filters, and Result Filtering ─────────────────────
	t.Run("Query_FiltersAndPagination", func(t *testing.T) {
		// Query with Result = FAILED
		res, err := svc.List(ctx, domain.AuditLogFilter{
			OrgID:  orgA,
			Result: domain.ResultFailed,
			Page:   1,
			Limit:  10,
		})
		if err != nil {
			t.Fatalf("Failed to list failed audit logs: %v", err)
		}
		if res.Total < 1 {
			t.Errorf("expected at least 1 failed audit log, got %d", res.Total)
		}
		for _, item := range res.Items {
			if item.Result != domain.ResultFailed {
				t.Errorf("expected only FAILED items, got %s", item.Result)
			}
		}

		// Query with Module = SHIPMENTS
		resMod, err := svc.List(ctx, domain.AuditLogFilter{
			OrgID:  orgA,
			Module: domain.ModuleShipments,
			Page:   1,
			Limit:  10,
		})
		if err != nil {
			t.Fatalf("Failed to list shipment audit logs: %v", err)
		}
		if resMod.Total < 1 {
			t.Errorf("expected at least 1 shipment audit log, got %d", resMod.Total)
		}
		for _, item := range resMod.Items {
			if item.Module != domain.ModuleShipments {
				t.Errorf("expected only SHIPMENTS items, got %s", item.Module)
			}
		}
	})
}
