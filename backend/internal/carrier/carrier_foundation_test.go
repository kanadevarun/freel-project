package carrier_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/freel/backend/internal/carrier/domain"
	"github.com/freel/backend/internal/carrier/repository"
	"github.com/freel/backend/internal/carrier/service"
	"github.com/freel/backend/internal/common/crypto"
	"github.com/freel/backend/internal/config"
	"github.com/freel/backend/internal/database"
	"github.com/jmoiron/sqlx"
)

func setupTestDB(t *testing.T) (*sqlx.DB, repository.CarrierRepository, service.CarrierService) {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test db: %v", err)
	}

	repo := repository.NewCarrierRepository(db)
	svc := service.NewCarrierService(repo, "Test_Carrier_Encryption_Key_32B!")

	return db, repo, svc
}

// ─────────────────────────────────────────────────────────────────────────────
// 1. Provider Registry Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestCarrierProviderRegistry(t *testing.T) {
	_, _, svc := setupTestDB(t)
	ctx := context.Background()

	providers, err := svc.GetProviders(ctx)
	if err != nil {
		t.Fatalf("GetProviders failed: %v", err)
	}

	if len(providers) < 4 {
		t.Fatalf("Expected at least 4 seeded providers, got %d", len(providers))
	}

	foundMaersk := false
	foundMSC := false
	foundHapag := false
	foundCMA := false

	for _, p := range providers {
		switch p.Code {
		case "MAERSK":
			foundMaersk = true
			if p.SCAC != "MAEU" {
				t.Errorf("Expected Maersk SCAC to be MAEU, got %s", p.SCAC)
			}
			if len(p.SupportedCapabilities) == 0 {
				t.Errorf("Expected Maersk to declare supported capabilities")
			}
		case "MSC":
			foundMSC = true
			if p.SCAC != "MSCU" {
				t.Errorf("Expected MSC SCAC to be MSCU, got %s", p.SCAC)
			}
		case "HAPAG_LLOYD":
			foundHapag = true
			if p.SCAC != "HLCU" {
				t.Errorf("Expected Hapag SCAC to be HLCU, got %s", p.SCAC)
			}
		case "CMA_CGM":
			foundCMA = true
		}
	}

	if !foundMaersk || !foundMSC || !foundHapag || !foundCMA {
		t.Errorf("Registry missing core carriers: Maersk=%v, MSC=%v, Hapag=%v, CMA=%v",
			foundMaersk, foundMSC, foundHapag, foundCMA)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 2. Adapter Registry & Resolution Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestAdapterRegistryResolution(t *testing.T) {
	reg := adapters.GetDefaultRegistry()

	// Direct code lookup
	maerskAdapter, ok := reg.GetAdapter("MAERSK")
	if !ok || maerskAdapter.ProviderCode() != "MAERSK" {
		t.Errorf("Failed to resolve MAERSK adapter by code")
	}

	mscAdapter, ok := reg.GetAdapter("MSC")
	if !ok || mscAdapter.ProviderCode() != "MSC" {
		t.Errorf("Failed to resolve MSC adapter by code")
	}

	// SCAC lookup (including alias SCACs)
	maeuAdapter := reg.GetAdapterBySCAC("MAEU")
	if maeuAdapter.ProviderCode() != "MAERSK" {
		t.Errorf("Expected MAEU to resolve to MAERSK, got %s", maeuAdapter.ProviderCode())
	}

	mskAdapter := reg.GetAdapterBySCAC("MSK")
	if mskAdapter.ProviderCode() != "MAERSK" {
		t.Errorf("Expected MSK alias to resolve to MAERSK, got %s", mskAdapter.ProviderCode())
	}

	hlcuAdapter := reg.GetAdapterBySCAC("HLCU")
	if hlcuAdapter.ProviderCode() != "HAPAG_LLOYD" {
		t.Errorf("Expected HLCU to resolve to HAPAG_LLOYD, got %s", hlcuAdapter.ProviderCode())
	}

	// Unregistered / Custom SCAC resolves to GenericAdapter
	customAdapter := reg.GetAdapterBySCAC("XYZU")
	if customAdapter.SCAC() != "XYZU" {
		t.Errorf("Expected generic adapter with SCAC XYZU, got %s", customAdapter.SCAC())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 3. Unsupported Adapter Operations Fail Safely
// ─────────────────────────────────────────────────────────────────────────────

func TestUnsupportedAdapterOperationsFailSafely(t *testing.T) {
	reg := adapters.GetDefaultRegistry()
	adapter := reg.GetAdapterBySCAC("HLCU")
	ctx := context.Background()
	creds := domain.DecryptedCredentials{APIKey: "test-key"}

	// Hapag-Lloyd adapter does not implement real live booking yet in Task 1 scaffolding
	_, err := adapter.CreateBooking(ctx, creds, domain.EnvSandbox, domain.BookingRequest{
		OriginPort:      "INNSA",
		DestinationPort: "DEHAM",
	})
	if err == nil {
		t.Errorf("Expected CreateBooking to return unsupported error, got nil")
	}

	// Error message should be safe and structured
	if !strings.Contains(err.Error(), "HAPAG_LLOYD") {
		t.Errorf("Expected error to contain provider code, got %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 4. Tenant Scoping, Credential Encryption & Zero-Data Isolation Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestCarrierIntegrationTenantScopingAndSecurity(t *testing.T) {
	db, _, svc := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	const orgA = int64(99901)
	const orgB = int64(99902)
	const testUserID = int64(1001)

	// Clean up previous test artifacts
	_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id IN (?, ?)", orgA, orgB)
	_, _ = db.Exec("DELETE FROM audit_logs WHERE org_id IN (?, ?)", orgA, orgB)
	_, _ = db.Exec("DELETE FROM users WHERE id = ?", testUserID)
	_, _ = db.Exec("DELETE FROM organizations WHERE id IN (?, ?)", orgA, orgB)

	// Insert test organizations to satisfy foreign key constraint
	_, err := db.Exec("INSERT INTO organizations (id, name) VALUES (?, 'Test Org A'), (?, 'Test Org B')", orgA, orgB)
	if err != nil {
		t.Fatalf("Failed to insert test organizations: %v", err)
	}
	_, err = db.Exec("INSERT INTO users (id, cognito_sub, email, first_name, last_name, created_at, updated_at) VALUES (?, 'sub-carrier-test', 'testuser@logistics.com', 'Test', 'User', NOW(), NOW()) ON DUPLICATE KEY UPDATE first_name=VALUES(first_name)", testUserID)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}
	defer func() {
		_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id IN (?, ?)", orgA, orgB)
		_, _ = db.Exec("DELETE FROM audit_logs WHERE org_id IN (?, ?)", orgA, orgB)
		_, _ = db.Exec("DELETE FROM users WHERE id = ?", testUserID)
		_, _ = db.Exec("DELETE FROM organizations WHERE id IN (?, ?)", orgA, orgB)
	}()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer testServer.Close()

	// 1. Connect Maersk for Org A with secret API keys
	rawSecretKey := "secret-api-key-xyz-987654321"
	rawSecretToken := "super-confidential-token-555"

	reqA := domain.ConnectCarrierRequest{
		CarrierSCAC:      "MAEU",
		Environment:      "PRODUCTION",
		ConnectionMethod: "API",
		Credentials: map[string]interface{}{
			"api_key":      rawSecretKey,
			"bearer_token": rawSecretToken,
			"base_url":     testServer.URL,
		},
		Capabilities: []string{"TRACKING", "RATES"},
	}

	createdA, err := svc.ConnectCarrier(ctx, orgA, testUserID, reqA)
	if err != nil {
		t.Fatalf("Org A ConnectCarrier failed: %v", err)
	}

	if createdA.ID <= 0 {
		t.Errorf("Expected positive integration ID, got %d", createdA.ID)
	}
	if createdA.CarrierSCAC != "MAEU" {
		t.Errorf("Expected SCAC MAEU, got %s", createdA.CarrierSCAC)
	}
	if !createdA.HasCredentials {
		t.Errorf("Expected HasCredentials to be true")
	}

	// 2. Verify Credential Masking in API response: Decrypted secrets MUST NEVER leak!
	if createdA.CredentialMask["api_key"] == rawSecretKey {
		t.Fatalf("CRITICAL SECURITY VIOLATION: Plaintext api_key returned in API view!")
	}
	if !strings.Contains(createdA.CredentialMask["api_key"], "••••••••") {
		t.Errorf("Expected masked api_key with bullets, got: %s", createdA.CredentialMask["api_key"])
	}

	// 3. Verify Database Encryption at Rest
	var dbEncryptedCreds string
	err = db.Get(&dbEncryptedCreds, "SELECT encrypted_credentials FROM carrier_integrations WHERE id = ?", createdA.ID)
	if err != nil {
		t.Fatalf("Failed to query encrypted credentials from DB: %v", err)
	}
	if dbEncryptedCreds == "" || strings.Contains(dbEncryptedCreds, rawSecretKey) {
		t.Fatalf("CRITICAL SECURITY VIOLATION: Database contains unencrypted plaintext secrets!")
	}

	// Decrypt using test encryption key to verify integrity
	decryptedStr, err := crypto.Decrypt(dbEncryptedCreds, "Test_Carrier_Encryption_Key_32B!")
	if err != nil {
		t.Fatalf("Failed to decrypt credentials from DB: %v", err)
	}
	if !strings.Contains(decryptedStr, rawSecretKey) {
		t.Errorf("Decrypted credentials mismatch: got %s", decryptedStr)
	}

	// 4. Verify Tenant Isolation: Org B MUST NOT see Org A's integration
	listB, err := svc.GetIntegrations(ctx, orgB)
	if err != nil {
		t.Fatalf("Org B GetIntegrations failed: %v", err)
	}
	if len(listB) != 0 {
		t.Errorf("Tenant isolation breached: Org B saw %d integrations from Org A!", len(listB))
	}

	// Org B cannot access Org A's integration by ID
	_, err = svc.GetIntegration(ctx, orgB, createdA.ID)
	if err == nil {
		t.Fatalf("Tenant isolation breached: Org B was able to fetch Org A's integration %d!", createdA.ID)
	}

	// Org B cannot test Org A's integration
	_, err = svc.TestConnection(ctx, orgB, createdA.ID)
	if err == nil {
		t.Fatalf("Tenant isolation breached: Org B was able to test Org A's integration %d!", createdA.ID)
	}

	// 5. Test Duplicate Integration Prevention
	_, err = svc.ConnectCarrier(ctx, orgA, testUserID, reqA)
	if err == nil {
		t.Errorf("Expected duplicate connection for exact same org + carrier + env to fail, got nil")
	}

	// 6. Test Connection Execution for Org A
	testRes, err := svc.TestConnection(ctx, orgA, createdA.ID)
	if err != nil {
		t.Fatalf("TestConnection failed: %v", err)
	}
	if !testRes.Success {
		t.Errorf("Expected connection test to succeed for configured carrier, got: %s", testRes.Message)
	}

	// 7. Verify Audit Log was recorded without secrets
	var auditCount int
	var auditDetails string
	err = db.Get(&auditCount, "SELECT COUNT(*) FROM audit_logs WHERE org_id = ? AND resource_id = ?", orgA, createdA.ID)
	if err != nil || auditCount == 0 {
		t.Errorf("Expected audit logs to be recorded for carrier actions, got count=%d", auditCount)
	}
	err = db.Get(&auditDetails, "SELECT details FROM audit_logs WHERE org_id = ? AND action = 'CARRIER_CONNECTED' LIMIT 1", orgA)
	if err == nil {
		if strings.Contains(auditDetails, rawSecretKey) || strings.Contains(auditDetails, rawSecretToken) {
			t.Fatalf("CRITICAL SECURITY VIOLATION: Audit logs contain raw secret keys!")
		}
	}

	// 8. Clean up after test
	err = svc.DisconnectCarrier(ctx, orgA, testUserID, createdA.ID)
	if err != nil {
		t.Fatalf("DisconnectCarrier failed: %v", err)
	}

	listAAfter, _ := svc.GetIntegrations(ctx, orgA)
	if len(listAAfter) != 0 {
		t.Errorf("Expected 0 integrations after disconnect, got %d", len(listAAfter))
	}
}
