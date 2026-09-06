package carrier_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/freel/backend/internal/carrier/domain"
	"github.com/freel/backend/internal/carrier/service"
	"github.com/freel/backend/internal/common/crypto"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCarrierProviderRegistrySchemas(t *testing.T) {
	db, _, svc := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	providers, err := svc.GetProviders(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, providers)

	var maersk, msc, hapag, one, evergreen, cosco *domain.CarrierProvider
	for i := range providers {
		p := &providers[i]
		switch p.Code {
		case "MAERSK":
			maersk = p
		case "MSC":
			msc = p
		case "HAPAG_LLOYD":
			hapag = p
		case "ONE":
			one = p
		case "EVERGREEN":
			evergreen = p
		case "COSCO":
			cosco = p
		}
	}

	require.NotNil(t, maersk, "Maersk provider must be registered")
	require.NotNil(t, msc, "MSC provider must be registered")
	require.NotNil(t, hapag, "Hapag-Lloyd provider must be registered")
	require.NotNil(t, one, "ONE provider must be registered")
	require.NotNil(t, evergreen, "Evergreen provider must be registered")
	require.NotNil(t, cosco, "COSCO provider must be registered")

	// Verify provider-aware credential schemas
	assert.NotEmpty(t, maersk.CredentialFields, "Maersk should define structured credential fields")
	assert.NotEmpty(t, msc.CredentialFields, "MSC should define structured credential fields")
	assert.NotEmpty(t, hapag.CredentialFields, "Hapag-Lloyd should define structured credential fields")
	assert.NotEmpty(t, one.CredentialFields, "ONE should define structured credential fields")
	assert.NotEmpty(t, evergreen.CredentialFields, "Evergreen should define structured credential fields")
	assert.NotEmpty(t, cosco.CredentialFields, "COSCO should define structured credential fields")

	// Verify dynamic capabilities
	assert.Contains(t, maersk.SupportedCapabilities, domain.CapTracking)
	assert.Contains(t, maersk.SupportedCapabilities, domain.CapRates)
	assert.Contains(t, maersk.SupportedCapabilities, domain.CapBooking)
	assert.Contains(t, maersk.SupportedCapabilities, domain.CapDocuments)

	assert.Contains(t, one.SupportedCapabilities, domain.CapTracking)
	assert.Contains(t, one.SupportedCapabilities, domain.CapRates)
	assert.Contains(t, one.SupportedCapabilities, domain.CapBooking)
}

func TestCarrierConnectionManagementLifecycle(t *testing.T) {
	db, _, svc := setupTestDB(t)
	defer db.Close()

	encKey := "Test_Carrier_Encryption_Key_32B!"
	ctx := context.Background()
	testOrgID := int64(999888) // Isolated test tenant
	testUserID := int64(777666)

	// Mock Gateway Server for connection testing
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer testServer.Close()

	// Clean up any test records
	_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id = ?", testOrgID)
	_, _ = db.Exec("DELETE FROM organizations WHERE id = ?", testOrgID)

	_, err := db.Exec("INSERT INTO organizations (id, name) VALUES (?, 'Test Carrier Mgmt Org')", testOrgID)
	require.NoError(t, err)
	defer func() {
		_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id = ?", testOrgID)
		_, _ = db.Exec("DELETE FROM organizations WHERE id = ?", testOrgID)
	}()

	// 1. Direct Pre-Flight Test (Valid Sandbox Credentials with test server)
	testReq := domain.TestDirectRequest{
		CarrierSCAC:      "MAEU",
		Environment:      "SANDBOX",
		ConnectionMethod: "API",
		Credentials: map[string]interface{}{
			"api_key":    "maersk-test-key-1234",
			"api_secret": "maersk-secret-5678",
			"base_url":   testServer.URL,
		},
	}
	testRes, err := svc.TestDirectConnection(ctx, testOrgID, testReq)
	require.NoError(t, err)
	assert.True(t, testRes.Success, "Sandbox pre-flight test should succeed")
	assert.Equal(t, domain.EnvSandbox, testRes.TestedEnvironment)
	assert.NotEmpty(t, testRes.TestedCapabilities)

	// 2. Direct Pre-Flight Test (Empty / Invalid Credentials)
	invalidReq := domain.TestDirectRequest{
		CarrierSCAC:      "MAEU",
		Environment:      "PRODUCTION",
		ConnectionMethod: "API",
		Credentials:      map[string]interface{}{},
	}
	invalidRes, err := svc.TestDirectConnection(ctx, testOrgID, invalidReq)
	require.NoError(t, err)
	assert.False(t, invalidRes.Success, "Empty credentials should fail pre-flight test")
	assert.NotEmpty(t, invalidRes.Message)

	// 3. Save & Connect Carrier (AES-256-GCM Encryption + Masking)
	connectReq := domain.ConnectCarrierRequest{
		CarrierSCAC:      "MAEU",
		Environment:      "SANDBOX",
		ConnectionMethod: "API",
		Credentials: map[string]interface{}{
			"api_key":    "maersk-production-key-9999",
			"api_secret": "super-secret-production-password",
			"base_url":   testServer.URL,
		},
		Capabilities: []string{"TRACKING", "RATES", "BOOKING"},
	}

	createdView, err := svc.ConnectCarrier(ctx, testOrgID, testUserID, connectReq)
	require.NoError(t, err)
	require.NotNil(t, createdView)
	assert.Equal(t, "MAEU", createdView.CarrierSCAC)
	assert.Equal(t, domain.EnvSandbox, createdView.Environment)
	assert.True(t, createdView.IsEnabled)
	assert.Contains(t, createdView.Capabilities, domain.CapTracking)
	assert.Contains(t, createdView.Capabilities, domain.CapRates)
	assert.Contains(t, createdView.Capabilities, domain.CapBooking)

	// Verify Credential Masking in API View
	assert.NotEmpty(t, createdView.CredentialMask)
	assert.Equal(t, "••••••••9999", createdView.CredentialMask["api_key"])
	assert.Equal(t, "••••••••word", createdView.CredentialMask["api_secret"])

	// Verify database encryption at rest (raw secret is NOT stored in DB)
	var dbEncrypted, dbMask string
	err = db.QueryRow("SELECT encrypted_credentials, credential_mask FROM carrier_integrations WHERE id = ?", createdView.ID).Scan(&dbEncrypted, &dbMask)
	require.NoError(t, err)
	assert.NotContains(t, dbEncrypted, "super-secret-production-password", "DB must not contain plaintext secrets")

	decryptedJSON, err := crypto.Decrypt(dbEncrypted, encKey)
	require.NoError(t, err)
	assert.Contains(t, decryptedJSON, "super-secret-production-password", "Decrypted ciphertext must match original secret")

	// 4. Duplicate Connection Prevention
	_, err = svc.ConnectCarrier(ctx, testOrgID, testUserID, connectReq)
	assert.ErrorIs(t, err, service.ErrDuplicateIntegration, "Duplicate connection for same org, SCAC, and env must be rejected")

	// 5. Test Saved Connection
	testSavedRes, err := svc.TestConnection(ctx, testOrgID, createdView.ID)
	require.NoError(t, err)
	assert.True(t, testSavedRes.Success)

	// 6. Edit Configuration (Update Environment, Capabilities, and Rotate Credentials)
	newEnv := "PRODUCTION"
	updateReq := domain.UpdateCarrierRequest{
		Environment:  &newEnv,
		Capabilities: []string{"TRACKING", "RATES", "CONTRACT_RATES", "BOOKING", "DOCUMENTS"},
		Credentials: map[string]interface{}{
			"api_key":    "maersk-rotated-key-1111",
			"api_secret": "new-rotated-secret-2222",
		},
	}
	updatedView, err := svc.UpdateCarrier(ctx, testOrgID, testUserID, createdView.ID, updateReq)
	require.NoError(t, err)
	assert.Equal(t, domain.EnvProduction, updatedView.Environment)
	assert.Contains(t, updatedView.Capabilities, domain.CapContractRates)
	assert.Contains(t, updatedView.Capabilities, domain.CapDocuments)
	assert.Equal(t, "••••••••1111", updatedView.CredentialMask["api_key"])

	// 7. Toggle Disabled
	toggledDisabled, err := svc.ToggleCarrier(ctx, testOrgID, testUserID, createdView.ID, false)
	require.NoError(t, err)
	assert.False(t, toggledDisabled.IsEnabled)
	assert.Equal(t, domain.StatusDisabled, toggledDisabled.ConnectionStatus)

	// 8. Toggle Re-Enabled
	toggledEnabled, err := svc.ToggleCarrier(ctx, testOrgID, testUserID, createdView.ID, true)
	require.NoError(t, err)
	assert.True(t, toggledEnabled.IsEnabled)
	assert.Equal(t, domain.StatusConnected, toggledEnabled.ConnectionStatus)

	// 9. Multi-Tenant Scoping (Org B cannot access Org A's integration)
	otherOrgID := int64(888111)
	_, err = svc.GetIntegration(ctx, otherOrgID, createdView.ID)
	assert.ErrorIs(t, err, service.ErrIntegrationNotFound, "Other tenant must not be able to read connection")

	_, err = svc.UpdateCarrier(ctx, otherOrgID, testUserID, createdView.ID, updateReq)
	assert.ErrorIs(t, err, service.ErrIntegrationNotFound, "Other tenant must not be able to edit connection")

	err = svc.DisconnectCarrier(ctx, otherOrgID, testUserID, createdView.ID)
	assert.ErrorIs(t, err, service.ErrIntegrationNotFound, "Other tenant must not be able to disconnect connection")

	// 10. Disconnect Carrier
	err = svc.DisconnectCarrier(ctx, testOrgID, testUserID, createdView.ID)
	require.NoError(t, err)

	// Verify integration is removed
	_, err = svc.GetIntegration(ctx, testOrgID, createdView.ID)
	assert.ErrorIs(t, err, service.ErrIntegrationNotFound)

	// Verify tenant records are clean
	list, err := svc.GetIntegrations(ctx, testOrgID)
	require.NoError(t, err)
	assert.Empty(t, list, "Org carrier list must be empty after disconnect")

	// Clean up
	_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id = ?", testOrgID)
}
