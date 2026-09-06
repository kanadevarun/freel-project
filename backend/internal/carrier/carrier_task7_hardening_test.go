package carrier_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/freel/backend/internal/carrier/domain"
	"github.com/freel/backend/internal/carrier/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask7_ProductionHardeningAndTenantIsolation(t *testing.T) {
	db, _, svc := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Mock Gateway Server for testing connection and adapter operations
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "access_token": "mock_test_token_123", "expires_in": 3600}`))
	}))
	defer testServer.Close()

	orgA := int64(88801)
	orgB := int64(88802)
	userA := int64(9901)
	userB := int64(9902)

	// Clean up previous test runs for test orgs
	_, _ = db.Exec("DELETE FROM carrier_sync_jobs WHERE org_id IN (?, ?)", orgA, orgB)
	_, _ = db.Exec("DELETE FROM audit_logs WHERE org_id IN (?, ?)", orgA, orgB)
	_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id IN (?, ?)", orgA, orgB)
	_, _ = db.Exec("DELETE FROM organizations WHERE id IN (?, ?)", orgA, orgB)

	_, err := db.Exec("INSERT INTO organizations (id, name) VALUES (?, 'Test Hardening Org A')", orgA)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO organizations (id, name) VALUES (?, 'Test Hardening Org B')", orgB)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM carrier_sync_jobs WHERE org_id IN (?, ?)", orgA, orgB)
		_, _ = db.Exec("DELETE FROM audit_logs WHERE org_id IN (?, ?)", orgA, orgB)
		_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id IN (?, ?)", orgA, orgB)
		_, _ = db.Exec("DELETE FROM organizations WHERE id IN (?, ?)", orgA, orgB)
	}()

	// 1. Verify Connect Carrier with Encrypted Credentials for Org A
	connectReqA := domain.ConnectCarrierRequest{
		CarrierSCAC:      "MAEU",
		Environment:      "SANDBOX",
		ConnectionMethod: "API",
		Credentials: map[string]interface{}{
			"api_key":       "msk_sandbox_client_9988",
			"client_secret": "super_secret_password_do_not_leak",
			"base_url":      testServer.URL,
		},
		Capabilities: []string{"TRACKING", "RATES", "BOOKING"},
	}

	createdA, err := svc.ConnectCarrier(ctx, orgA, userA, connectReqA)
	require.NoError(t, err)
	require.NotNil(t, createdA)
	assert.Equal(t, orgA, createdA.OrgID)
	assert.Equal(t, "MAEU", createdA.CarrierSCAC)
	assert.True(t, createdA.HasCredentials)

	// 2. Production Hardening: Verify Decrypted Credentials NEVER appear in View
	assert.NotEmpty(t, createdA.CredentialMask)
	for key, maskedVal := range createdA.CredentialMask {
		assert.Contains(t, maskedVal, "••••••••", "Credential key %s must be masked in API View", key)
		assert.NotContains(t, maskedVal, "super_secret_password_do_not_leak", "Raw secret must never appear in API View")
	}

	// 3. Multi-Tenant Isolation Audit: Org B MUST NOT be able to view Org A's integration
	viewOrgB, err := svc.GetIntegration(ctx, orgB, createdA.ID)
	assert.Error(t, err, "Org B should receive NOT FOUND when trying to access Org A's integration")
	assert.Nil(t, viewOrgB)

	integrationsOrgB, err := svc.GetIntegrations(ctx, orgB)
	require.NoError(t, err)
	for _, integ := range integrationsOrgB {
		assert.NotEqual(t, orgA, integ.OrgID, "Org B should not see Org A's integrations")
	}

	// 4. Duplicate Connection Prevention: Org A cannot connect same carrier/env twice
	_, errDup := svc.ConnectCarrier(ctx, orgA, userA, connectReqA)
	assert.Error(t, errDup, "Duplicate active connection for same carrier & env must fail")
	assert.Equal(t, service.ErrDuplicateIntegration, errDup)

	// 5. Tenant Isolation: Org B CAN connect the same carrier independently
	createdB, err := svc.ConnectCarrier(ctx, orgB, userB, connectReqA)
	require.NoError(t, err, "Org B should be able to create its own independent carrier connection")
	assert.Equal(t, orgB, createdB.OrgID)
	assert.NotEqual(t, createdA.ID, createdB.ID)

	// 6. Security Audit: Database persistence check (encrypted_credentials must be encrypted)
	var encryptedBlob string
	err = db.GetContext(ctx, &encryptedBlob, "SELECT encrypted_credentials FROM carrier_integrations WHERE id = ?", createdA.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, encryptedBlob)
	assert.NotContains(t, encryptedBlob, "super_secret_password_do_not_leak", "Database record must be encrypted at rest with AES-GCM")

	// 7. Adapter Resolution & Normalization
	adapter := adapters.GetDefaultRegistry().GetAdapterBySCAC("MAEU")
	require.NotNil(t, adapter)
	assert.Equal(t, "MAERSK", adapter.ProviderCode())

	// 8. Health and Status Verification
	health, err := svc.GetIntegrationHealth(ctx, orgA, createdA.ID)
	require.NoError(t, err)
	require.NotNil(t, health)
	assert.Equal(t, createdA.ID, health.IntegrationID)
	assert.Equal(t, domain.HealthHealthy, health.HealthState)

	// 9. Sync Engine Idempotency and Execution
	jobView, err := svc.SyncNow(ctx, orgA, createdA.ID, domain.SyncNowRequest{ForceOverride: true})
	require.NoError(t, err)
	require.NotNil(t, jobView)
	assert.Equal(t, orgA, jobView.OrgID)

	// 10. Audit Trail Verification without Secret Leakage
	var auditAction string
	var auditDetails string
	err = db.QueryRowContext(ctx, "SELECT action, details FROM audit_logs WHERE org_id = ? ORDER BY id DESC LIMIT 1", orgA).Scan(&auditAction, &auditDetails)
	require.NoError(t, err)
	assert.NotEmpty(t, auditAction)
	assert.NotContains(t, auditDetails, "super_secret_password_do_not_leak", "Audit logs must never contain raw credentials")

	// 11. Disconnect and Cleanup
	err = svc.DisconnectCarrier(ctx, orgA, userA, createdA.ID)
	require.NoError(t, err)

	viewAfterDisconnect, err := svc.GetIntegration(ctx, orgA, createdA.ID)
	assert.Error(t, err)
	assert.Equal(t, service.ErrIntegrationNotFound, err)
	assert.Nil(t, viewAfterDisconnect)

	t.Log("✅ Task 7 Carrier Integration Verification & Production Hardening passed 100%!")
}
