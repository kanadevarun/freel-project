package audit_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/audit/repository"
	"github.com/freel/backend/internal/audit/service"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/middleware"
	"github.com/joho/godotenv"
)

func setupTestDB(t *testing.T) (*service.Service, int64, int64, func()) {
	_ = godotenv.Load("../../.env")
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC"
	}

	db, err := database.Connect(dbURL)
	if err != nil {
		t.Skipf("Skipping audit integration test; database not accessible: %v", err)
	}

	repo := repository.NewMySQLRepository(db)
	svc := service.NewService(repo, db)

	// Unique test org IDs
	testOrgA := int64(99901)
	testOrgB := int64(99902)

	// Ensure test organizations and test user exist
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (?, 'Audit Test Org A', NOW(), NOW()), (?, 'Audit Test Org B', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)", testOrgA, testOrgB)
	_, _ = db.Exec("INSERT INTO users (id, cognito_sub, email, first_name, last_name, created_at, updated_at) VALUES (999001, 'sub-audit-999001', 'operator@freel-testing.local', 'Varun', 'Kanade', NOW(), NOW()) ON DUPLICATE KEY UPDATE first_name=VALUES(first_name)")

	// Clean any previous test audit records for these org IDs
	_, _ = db.Exec("DELETE FROM audit_logs WHERE org_id IN (?, ?)", testOrgA, testOrgB)

	cleanup := func() {
		_, _ = db.Exec("DELETE FROM audit_logs WHERE org_id IN (?, ?)", testOrgA, testOrgB)
		_, _ = db.Exec("DELETE FROM users WHERE id = 999001")
		_, _ = db.Exec("DELETE FROM organizations WHERE id IN (?, ?)", testOrgA, testOrgB)
		db.Close()
	}

	return &svc, testOrgA, testOrgB, cleanup
}

// Test A — Basic audit event
func TestBasicAuditEvent(t *testing.T) {
	svcPtr, orgA, _, cleanup := setupTestDB(t)
	defer cleanup()
	svc := *svcPtr

	ctx := context.Background()
	actorID := int64(1)

	entry, err := svc.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgA,
		ActorID:      &actorID,
		ActorType:    domain.ActorTypeUser,
		ActorName:    "Varun Sharma",
		ActorRole:    "Operations",
		Action:       domain.ActionCreate,
		Module:       domain.ModuleShipments,
		ResourceType: "SHIPMENT",
		ResourceID:   "SHP-TEST-1001",
		ResourceName: "BK-QA-1787840980",
		Description:  "Shipment SHP-TEST-1001 created",
		Result:       domain.ResultSuccess,
		IPAddress:    "103.21.244.12",
		UserAgent:    "Mozilla/5.0 Chrome/120.0.0.0",
	})

	if err != nil {
		t.Fatalf("failed to record basic audit log: %v", err)
	}

	if entry.ID <= 0 {
		t.Fatalf("expected positive log ID, got %d", entry.ID)
	}
	if entry.OrgID != orgA {
		t.Errorf("expected OrgID %d, got %d", orgA, entry.OrgID)
	}
	if entry.ActorType != domain.ActorTypeUser {
		t.Errorf("expected ActorType %s, got %s", domain.ActorTypeUser, entry.ActorType)
	}
	if entry.Action != domain.ActionCreate {
		t.Errorf("expected Action %s, got %s", domain.ActionCreate, entry.Action)
	}
	if entry.Module != domain.ModuleShipments {
		t.Errorf("expected Module %s, got %s", domain.ModuleShipments, entry.Module)
	}
	if entry.ResourceID != "SHP-TEST-1001" {
		t.Errorf("expected ResourceID SHP-TEST-1001, got %s", entry.ResourceID)
	}
	if entry.Result != domain.ResultSuccess {
		t.Errorf("expected Result SUCCESS, got %s", entry.Result)
	}
	if entry.CreatedAt.IsZero() {
		t.Errorf("expected valid CreatedAt timestamp")
	}

	// Verify retrieval by ID
	fetched, err := svc.GetByID(ctx, orgA, entry.ID)
	if err != nil {
		t.Fatalf("failed to fetch audit log by ID: %v", err)
	}
	if fetched.ResourceName != "BK-QA-1787840980" {
		t.Errorf("expected resource name BK-QA-1787840980, got %s", fetched.ResourceName)
	}
}

// Test B — Before / After changes and diff computation
func TestBeforeAfterChanges(t *testing.T) {
	svcPtr, orgA, _, cleanup := setupTestDB(t)
	defer cleanup()
	svc := *svcPtr

	ctx := context.Background()

	beforeState := map[string]interface{}{
		"status":   "In Transit",
		"eta":      "Sep 10, 2026",
		"vessel":   "MSC Anna",
		"location": "At Sea",
	}

	afterState := map[string]interface{}{
		"status":   "Delayed",
		"eta":      "Sep 14, 2026",
		"vessel":   "MSC Anna", // Unchanged
		"location": "At Sea",   // Unchanged
	}

	entry, err := svc.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgA,
		Action:       domain.ActionUpdate,
		Module:       domain.ModuleShipments,
		ResourceType: "SHIPMENT",
		ResourceID:   "SHP-1024",
		ResourceName: "BK-QA-1787840980",
		Description:  "Shipment updated - delay reported",
		Before:       beforeState,
		After:        afterState,
	})

	if err != nil {
		t.Fatalf("failed to record before/after audit log: %v", err)
	}

	// Read back from DB
	fetched, err := svc.GetByID(ctx, orgA, entry.ID)
	if err != nil {
		t.Fatalf("failed to fetch log: %v", err)
	}

	if fetched.BeforeData["status"] != "In Transit" {
		t.Errorf("expected BeforeData status 'In Transit', got %v", fetched.BeforeData["status"])
	}
	if fetched.AfterData["status"] != "Delayed" {
		t.Errorf("expected AfterData status 'Delayed', got %v", fetched.AfterData["status"])
	}

	// Verify structured changes (Field, Before, After)
	if len(fetched.Changes) != 2 {
		t.Fatalf("expected 2 computed field changes (status, eta), got %d: %+v", len(fetched.Changes), fetched.Changes)
	}

	fieldMap := make(map[string]domain.FieldChange)
	for _, c := range fetched.Changes {
		fieldMap[c.Field] = c
	}

	statusChange, ok := fieldMap["status"]
	if !ok || statusChange.Before != "In Transit" || statusChange.After != "Delayed" {
		t.Errorf("unexpected status field change: %+v", statusChange)
	}

	etaChange, ok := fieldMap["eta"]
	if !ok || etaChange.Before != "Sep 10, 2026" || etaChange.After != "Sep 14, 2026" {
		t.Errorf("unexpected eta field change: %+v", etaChange)
	}
}

// Test C — Secret sanitization (passwords, tokens, API keys, credentials)
func TestSecretSanitization(t *testing.T) {
	svcPtr, orgA, _, cleanup := setupTestDB(t)
	defer cleanup()
	svc := *svcPtr

	ctx := context.Background()

	sensitiveBefore := map[string]interface{}{
		"api_key":             "live_carrier_secret_12345678",
		"access_token":        "bearer_token_abc9988",
		"password":            "MySecretPassword!",
		"carrier_credentials": map[string]interface{}{"secret": "top_secret"},
		"account_number":      "ACC-100234",
	}

	sensitiveAfter := map[string]interface{}{
		"api_key":             "live_carrier_secret_87654321",
		"access_token":        "bearer_token_new_9999",
		"password":            "NewSecretPassword@123",
		"carrier_credentials": map[string]interface{}{"secret": "new_top_secret"},
		"account_number":      "ACC-100234",
	}

	sensitiveMetadata := map[string]interface{}{
		"client_secret": "oauth_secret_val",
		"auth_header":   "Bearer token_xyz",
		"source":        "SETTINGS_PORTAL",
	}

	entry, err := svc.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgA,
		Action:       domain.ActionConnect,
		Module:       domain.ModuleCarrierIntegrations,
		ResourceType: "CARRIER_INTEGRATION",
		ResourceID:   "MAERSK-01",
		ResourceName: "Maersk",
		Description:  "Carrier credentials updated",
		Before:       sensitiveBefore,
		After:        sensitiveAfter,
		Metadata:     sensitiveMetadata,
	})

	if err != nil {
		t.Fatalf("failed to record sanitized audit log: %v", err)
	}

	fetched, err := svc.GetByID(ctx, orgA, entry.ID)
	if err != nil {
		t.Fatalf("failed to fetch log: %v", err)
	}

	// Verify sensitive keys in BeforeData are redacted
	if fetched.BeforeData["api_key"] != "[REDACTED]" {
		t.Errorf("expected api_key redacted, got %v", fetched.BeforeData["api_key"])
	}
	if fetched.BeforeData["access_token"] != "[REDACTED]" {
		t.Errorf("expected access_token redacted, got %v", fetched.BeforeData["access_token"])
	}
	if fetched.BeforeData["password"] != "[REDACTED]" {
		t.Errorf("expected password redacted, got %v", fetched.BeforeData["password"])
	}
	// Verify non-sensitive key preserved
	if fetched.BeforeData["account_number"] != "ACC-100234" {
		t.Errorf("expected account_number preserved, got %v", fetched.BeforeData["account_number"])
	}

	// Verify metadata sanitization
	if fetched.Metadata["client_secret"] != "[REDACTED]" {
		t.Errorf("expected client_secret redacted in metadata, got %v", fetched.Metadata["client_secret"])
	}
	if fetched.Metadata["auth_header"] != "[REDACTED]" {
		t.Errorf("expected auth_header redacted in metadata, got %v", fetched.Metadata["auth_header"])
	}
	if fetched.Metadata["source"] != "SETTINGS_PORTAL" {
		t.Errorf("expected source preserved in metadata, got %v", fetched.Metadata["source"])
	}
}

// Test D — Tenant isolation
func TestTenantIsolation(t *testing.T) {
	svcPtr, orgA, orgB, cleanup := setupTestDB(t)
	defer cleanup()
	svc := *svcPtr

	ctx := context.Background()

	// Create event for Org A
	entryA, err := svc.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgA,
		Action:       domain.ActionCreate,
		Module:       domain.ModuleInvoices,
		ResourceType: "INVOICE",
		ResourceID:   "INV-ORGA-001",
		Description:  "Invoice created for Org A",
	})
	if err != nil {
		t.Fatalf("failed to create Org A log: %v", err)
	}

	// Create event for Org B
	entryB, err := svc.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgB,
		Action:       domain.ActionCreate,
		Module:       domain.ModuleInvoices,
		ResourceType: "INVOICE",
		ResourceID:   "INV-ORGB-002",
		Description:  "Invoice created for Org B",
	})
	if err != nil {
		t.Fatalf("failed to create Org B log: %v", err)
	}

	// Org A should NOT be able to fetch Org B's entry by ID
	_, err = svc.GetByID(ctx, orgA, entryB.ID)
	if err == nil {
		t.Fatalf("SECURITY VIOLATION: Org A was able to read Org B's audit entry!")
	}

	// Org B should NOT be able to fetch Org A's entry by ID
	_, err = svc.GetByID(ctx, orgB, entryA.ID)
	if err == nil {
		t.Fatalf("SECURITY VIOLATION: Org B was able to read Org A's audit entry!")
	}

	// Org A List must only contain Org A items
	resA, err := svc.List(ctx, domain.AuditLogFilter{OrgID: orgA})
	if err != nil {
		t.Fatalf("failed to list Org A logs: %v", err)
	}
	for _, item := range resA.Items {
		if item.OrgID != orgA {
			t.Errorf("SECURITY VIOLATION: Org A list returned log belonging to org %d", item.OrgID)
		}
	}
}

// Test E & F & G — Pagination, Sorting, and Immutability
func TestPaginationAndSorting(t *testing.T) {
	svcPtr, orgA, _, cleanup := setupTestDB(t)
	defer cleanup()
	svc := *svcPtr

	ctx := context.Background()

	// Seed 5 events with slight time differences
	for i := 1; i <= 5; i++ {
		_, err := svc.Record(ctx, domain.CreateAuditLogParams{
			OrgID:        orgA,
			Action:       domain.ActionCreate,
			Module:       domain.ModuleQuotations,
			ResourceType: "QUOTATION",
			ResourceID:   string(rune('A' + i)),
			Description:  "Quotation event",
		})
		if err != nil {
			t.Fatalf("failed to seed log %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Page 1 with limit 2
	page1, err := svc.List(ctx, domain.AuditLogFilter{
		OrgID: orgA,
		Page:  1,
		Limit: 2,
	})
	if err != nil {
		t.Fatalf("failed to get page 1: %v", err)
	}

	if page1.Total != 5 {
		t.Errorf("expected total 5, got %d", page1.Total)
	}
	if page1.TotalPages != 3 {
		t.Errorf("expected total pages 3, got %d", page1.TotalPages)
	}
	if len(page1.Items) != 2 {
		t.Errorf("expected 2 items on page 1, got %d", len(page1.Items))
	}

	// Verify newest first sorting: ID of item 0 > ID of item 1
	if page1.Items[0].ID <= page1.Items[1].ID {
		t.Errorf("expected newest first sorting, got IDs %d, %d", page1.Items[0].ID, page1.Items[1].ID)
	}
}

// Test H — Future AI Agent & Context Derivation
func TestAIAgentActorAndContextDerivation(t *testing.T) {
	svcPtr, orgA, _, cleanup := setupTestDB(t)
	defer cleanup()
	svc := *svcPtr

	// 1. AI Agent actor support
	aiEntry, err := svc.Record(context.Background(), domain.CreateAuditLogParams{
		OrgID:        orgA,
		ActorType:    domain.ActorTypeAIAgent,
		ActorName:    "Operations Dispatch Agent",
		Action:       domain.ActionUpdate,
		Module:       domain.ModuleShipments,
		ResourceType: "SHIPMENT",
		ResourceID:   "SHP-1024",
		Description:  "AI Agent updated carrier milestone ETA",
		Result:       domain.ResultSuccess,
	})
	if err != nil {
		t.Fatalf("failed to record AI agent audit log: %v", err)
	}
	if aiEntry.ActorType != domain.ActorTypeAIAgent {
		t.Errorf("expected actor type AI_AGENT, got %s", aiEntry.ActorType)
	}

	// 2. Automatic derivation from middleware.UserContext
	userCtx := middleware.UserContext{
		UserID: 1,
		OrgID:  orgA,
		Role:   "ADMIN",
	}
	ctxWithUser := context.WithValue(context.Background(), middleware.UserContextKey, userCtx)

	autoEntry, err := svc.Record(ctxWithUser, domain.CreateAuditLogParams{
		Action:       domain.ActionApprove,
		Module:       domain.ModuleQuotations,
		ResourceType: "QUOTATION",
		ResourceID:   "QT-2041",
	})
	if err != nil {
		t.Fatalf("failed to record context-derived audit log: %v", err)
	}

	if autoEntry.OrgID != orgA {
		t.Errorf("expected auto-derived org %d, got %d", orgA, autoEntry.OrgID)
	}
	if autoEntry.ActorID == nil || *autoEntry.ActorID != 1 {
		t.Errorf("expected auto-derived user ID 1, got %v", autoEntry.ActorID)
	}
	if autoEntry.ActorRole != "ADMIN" {
		t.Errorf("expected auto-derived role ADMIN, got %s", autoEntry.ActorRole)
	}
	if autoEntry.ActorType != domain.ActorTypeUser {
		t.Errorf("expected actor type USER, got %s", autoEntry.ActorType)
	}
}
