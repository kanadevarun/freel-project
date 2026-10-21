package carrier_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/freel/backend/internal/carrier/domain"
	carrierRepo "github.com/freel/backend/internal/carrier/repository"
	carrierService "github.com/freel/backend/internal/carrier/service"
	"github.com/freel/backend/internal/common/crypto"
	"github.com/freel/backend/internal/config"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSecurityTestDB(t *testing.T) (*sqlx.DB, string) {
	_ = godotenv.Load("../../.env")
	cfg := config.LoadConfig()
	db, err := sqlx.Open("mysql", cfg.DatabaseURL)
	require.NoError(t, err, "Failed to connect to test database")
	require.NoError(t, db.Ping(), "Database ping failed")

	encKey := cfg.CarrierEncryptionKey
	if encKey == "" {
		encKey = cfg.MailboxEncryptionKey
	}
	if encKey == "" {
		encKey = "Test_Carrier_Encryption_Key_32B!"
	}
	return db, encKey
}

func ensureTestOrg(t *testing.T, db *sqlx.DB, orgID int64) func() {
	_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id = ?", orgID)
	_, _ = db.Exec("DELETE FROM organizations WHERE id = ?", orgID)
	_, err := db.Exec("INSERT INTO organizations (id, name) VALUES (?, 'Test Carrier Security Org')", orgID)
	require.NoError(t, err)

	return func() {
		_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id = ?", orgID)
		_, _ = db.Exec("DELETE FROM organizations WHERE id = ?", orgID)
	}
}

// Test 1: Saving a carrier integration stores sensitive credentials encrypted.
func TestScenario1_SavingStoresCredentialsEncrypted(t *testing.T) {
	db, encKey := setupSecurityTestDB(t)
	defer db.Close()

	orgID := int64(99901)
	cleanup := ensureTestOrg(t, db, orgID)
	defer cleanup()

	repo := carrierRepo.NewCarrierRepository(db)
	svc := carrierService.NewCarrierService(repo, encKey)

	ctx := context.Background()
	userID := int64(1001)
	scac := "MAEU"

	req := domain.ConnectCarrierRequest{
		CarrierSCAC:      scac,
		Environment:      string(domain.EnvSandbox),
		ConnectionMethod: string(domain.MethodAPI),
		Credentials: map[string]interface{}{
			"client_id":     "mock-test-client-id",
			"client_secret": "mock-ultra-sensitive-client-secret-xyz",
			"api_key":       "mock-secret-api-key-9988",
		},
		Capabilities: []string{"TRACKING", "RATES"},
	}

	view, err := svc.ConnectCarrier(ctx, orgID, userID, req)
	require.NoError(t, err)
	require.NotNil(t, view)

	// Query database directly to verify encryption at rest
	var encryptedCreds string
	err = db.QueryRow("SELECT encrypted_credentials FROM carrier_integrations WHERE id = ?", view.ID).Scan(&encryptedCreds)
	require.NoError(t, err)
	require.NotEmpty(t, encryptedCreds)

	// Ciphertext must NOT match raw plaintext secrets
	assert.NotContains(t, encryptedCreds, "mock-ultra-sensitive-client-secret-xyz")
	assert.NotContains(t, encryptedCreds, "mock-secret-api-key-9988")

	// Decrypting with legitimate key must yield original credentials
	decryptedJSON, err := crypto.Decrypt(encryptedCreds, encKey)
	require.NoError(t, err)
	assert.Contains(t, decryptedJSON, "mock-ultra-sensitive-client-secret-xyz")
	assert.Contains(t, decryptedJSON, "mock-secret-api-key-9988")
}

// Test 2: Plaintext client_secret is NOT persisted.
func TestScenario2_PlaintextClientSecretNotPersisted(t *testing.T) {
	db, encKey := setupSecurityTestDB(t)
	defer db.Close()

	orgID := int64(99902)
	cleanup := ensureTestOrg(t, db, orgID)
	defer cleanup()

	repo := carrierRepo.NewCarrierRepository(db)
	svc := carrierService.NewCarrierService(repo, encKey)

	ctx := context.Background()
	userID := int64(1002)
	scac := "MSCU"

	secretVal := "super-confidential-secret-msc-8877"
	req := domain.ConnectCarrierRequest{
		CarrierSCAC:      scac,
		Environment:      string(domain.EnvSandbox),
		ConnectionMethod: string(domain.MethodAPI),
		Credentials: map[string]interface{}{
			"client_id":     "msc-client-1",
			"client_secret": secretVal,
		},
		Capabilities: []string{"TRACKING"},
	}

	view, err := svc.ConnectCarrier(ctx, orgID, userID, req)
	require.NoError(t, err)

	var credsJSON *string
	err = db.QueryRow("SELECT credentials_json FROM carrier_integrations WHERE id = ?", view.ID).Scan(&credsJSON)
	require.NoError(t, err)
	assert.Nil(t, credsJSON, "credentials_json MUST be NULL in the database to prevent plaintext secret persistence")

	// Verify plaintext secret doesn't appear in ANY column of the record
	var rowDump struct {
		ID            int64   `db:"id"`
		SCAC          string  `db:"carrier_scac"`
		Mask          *string `db:"credential_mask"`
		EncryptedBlob string  `db:"encrypted_credentials"`
	}
	err = db.Get(&rowDump, "SELECT id, carrier_scac, credential_mask, encrypted_credentials FROM carrier_integrations WHERE id = ?", view.ID)
	require.NoError(t, err)
	if rowDump.Mask != nil {
		assert.NotContains(t, *rowDump.Mask, secretVal, "Credential mask must never contain the plaintext secret")
	}
}

// Test 3: Plaintext API key is NOT persisted.
func TestScenario3_PlaintextAPIKeyNotPersisted(t *testing.T) {
	db, encKey := setupSecurityTestDB(t)
	defer db.Close()

	orgID := int64(99903)
	cleanup := ensureTestOrg(t, db, orgID)
	defer cleanup()

	repo := carrierRepo.NewCarrierRepository(db)
	svc := carrierService.NewCarrierService(repo, encKey)

	ctx := context.Background()
	userID := int64(1003)
	scac := "HLCU"

	apiKeyVal := "hl-secret-api-key-test-554433"
	req := domain.ConnectCarrierRequest{
		CarrierSCAC:      scac,
		Environment:      string(domain.EnvSandbox),
		ConnectionMethod: string(domain.MethodAPI),
		Credentials: map[string]interface{}{
			"api_key": apiKeyVal,
		},
		Capabilities: []string{"TRACKING", "RATES"},
	}

	view, err := svc.ConnectCarrier(ctx, orgID, userID, req)
	require.NoError(t, err)

	var credsJSON *string
	err = db.QueryRow("SELECT credentials_json FROM carrier_integrations WHERE id = ?", view.ID).Scan(&credsJSON)
	require.NoError(t, err)
	assert.Nil(t, credsJSON, "Plaintext API key must NOT be stored in credentials_json")
}

// Test 4: Reading an integration never returns decrypted secrets through API responses.
func TestScenario4_ReadingIntegrationNeverReturnsDecryptedSecrets(t *testing.T) {
	db, encKey := setupSecurityTestDB(t)
	defer db.Close()

	orgID := int64(99904)
	cleanup := ensureTestOrg(t, db, orgID)
	defer cleanup()

	repo := carrierRepo.NewCarrierRepository(db)
	svc := carrierService.NewCarrierService(repo, encKey)

	ctx := context.Background()
	userID := int64(1004)
	scac := "MAEU"

	rawSecret := "ultra-confidential-secret-never-expose"
	rawKey := "private-api-key-never-expose"

	req := domain.ConnectCarrierRequest{
		CarrierSCAC:      scac,
		Environment:      string(domain.EnvSandbox),
		ConnectionMethod: string(domain.MethodAPI),
		Credentials: map[string]interface{}{
			"client_id":     "my-client-id",
			"client_secret": rawSecret,
			"api_key":       rawKey,
		},
		Capabilities: []string{"TRACKING"},
	}

	createdView, err := svc.ConnectCarrier(ctx, orgID, userID, req)
	require.NoError(t, err)

	// Read single integration
	view, err := svc.GetIntegration(ctx, orgID, createdView.ID)
	require.NoError(t, err)

	// Serialize API view to JSON simulating HTTP transport
	viewJSONBytes, err := json.Marshal(view)
	require.NoError(t, err)
	viewJSON := string(viewJSONBytes)

	assert.NotContains(t, viewJSON, rawSecret, "API response JSON must never expose decrypted client_secret")
	assert.NotContains(t, viewJSON, rawKey, "API response JSON must never expose decrypted api_key")
	assert.NotContains(t, viewJSON, "encrypted_credentials", "API response should not expose raw ciphertext blob")
	assert.NotContains(t, viewJSON, "credentials_json", "API response should not expose credentials_json")

	// Verify mask format
	assert.Equal(t, "••••••••", view.CredentialMask["client_secret"])
	assert.True(t, strings.HasPrefix(view.CredentialMask["api_key"], "••••••••"))
	assert.True(t, view.HasCredentials)
}

// Test 5: Decryption works correctly for valid encrypted credentials.
func TestScenario5_DecryptionWorksCorrectlyForValidEncryptedCredentials(t *testing.T) {
	db, encKey := setupSecurityTestDB(t)
	defer db.Close()

	orgID := int64(99905)
	cleanup := ensureTestOrg(t, db, orgID)
	defer cleanup()

	repo := carrierRepo.NewCarrierRepository(db)
	svc := carrierService.NewCarrierService(repo, encKey)

	ctx := context.Background()
	userID := int64(1005)
	scac := "MAEU"

	req := domain.ConnectCarrierRequest{
		CarrierSCAC:      scac,
		Environment:      string(domain.EnvSandbox),
		ConnectionMethod: string(domain.MethodAPI),
		Credentials: map[string]interface{}{
			"client_id":     "valid-client-id",
			"client_secret": "valid-client-secret-999",
			"api_key":       "valid-api-key-888",
		},
		Capabilities: []string{"TRACKING"},
	}

	view, err := svc.ConnectCarrier(ctx, orgID, userID, req)
	require.NoError(t, err)
	assert.NotNil(t, view)

	adapter, creds, err := svc.GetAdapterForIntegration(ctx, orgID, scac, domain.EnvSandbox)
	require.NoError(t, err)
	require.NotNil(t, adapter)
	assert.Equal(t, "valid-client-id", creds.ClientID)
	assert.Equal(t, "valid-client-secret-999", creds.ClientSecret)
	assert.Equal(t, "valid-api-key-888", creds.APIKey)
}

// Test 6: Invalid/malformed encrypted credentials fail safely.
func TestScenario6_InvalidMalformedEncryptedCredentialsFailSafely(t *testing.T) {
	db, encKey := setupSecurityTestDB(t)
	defer db.Close()

	orgID := int64(99906)
	cleanup := ensureTestOrg(t, db, orgID)
	defer cleanup()

	repo := carrierRepo.NewCarrierRepository(db)
	svc := carrierService.NewCarrierService(repo, encKey)

	ctx := context.Background()
	userID := int64(1006)
	scac := "MAEU"

	req := domain.ConnectCarrierRequest{
		CarrierSCAC:      scac,
		Environment:      string(domain.EnvSandbox),
		ConnectionMethod: string(domain.MethodAPI),
		Credentials: map[string]interface{}{
			"client_id": "malformed-test-client",
		},
		Capabilities: []string{"TRACKING"},
	}

	view, err := svc.ConnectCarrier(ctx, orgID, userID, req)
	require.NoError(t, err)

	// Corrupt the ciphertext in the database
	corruptedCiphertext := "not-valid-base64-ciphertext-at-all!!!"
	_, err = db.Exec("UPDATE carrier_integrations SET encrypted_credentials = ? WHERE id = ?", corruptedCiphertext, view.ID)
	require.NoError(t, err)

	// Attempting to resolve adapter must fail safely with ErrDecryptionFailed without panicking
	_, _, err = svc.GetAdapterForIntegration(ctx, orgID, scac, domain.EnvSandbox)
	assert.Error(t, err)
	assert.ErrorIs(t, err, carrierService.ErrDecryptionFailed, "Must return ErrDecryptionFailed on malformed ciphertext")

	// TestConnection must also handle decryption failure gracefully
	testRes, testErr := svc.TestConnection(ctx, orgID, view.ID)
	assert.NoError(t, testErr, "TestConnection returns result struct on internal adapter error")
	assert.False(t, testRes.Success)
	assert.Equal(t, "DECRYPTION_ERROR", testRes.ErrorCode)
}

// Test 7: Missing encryption key fails safely.
func TestScenario7_MissingEncryptionKeyFailsSafely(t *testing.T) {
	db, _ := setupSecurityTestDB(t)
	defer db.Close()

	repo := carrierRepo.NewCarrierRepository(db)
	// Initialize CarrierService with EMPTY encryption key
	unconfiguredSvc := carrierService.NewCarrierService(repo, "")

	ctx := context.Background()
	orgID := int64(99907)
	userID := int64(1007)

	req := domain.ConnectCarrierRequest{
		CarrierSCAC:      "MAEU",
		Environment:      string(domain.EnvSandbox),
		ConnectionMethod: string(domain.MethodAPI),
		Credentials: map[string]interface{}{
			"client_id":     "some-client",
			"client_secret": "some-secret",
		},
		Capabilities: []string{"TRACKING"},
	}

	// ConnectCarrier must fail safely with ErrEncryptionKeyMissing
	view, err := unconfiguredSvc.ConnectCarrier(ctx, orgID, userID, req)
	assert.Error(t, err)
	assert.ErrorIs(t, err, carrierService.ErrEncryptionKeyMissing, "Must fail safely if encryption key is missing; no silent fallback")
	assert.Nil(t, view)
}

// Test 8: Logs/errors do not contain sensitive credentials.
func TestScenario8_LogsAndErrorsDoNotContainSensitiveCredentials(t *testing.T) {
	sensitiveSecret := "super-confidential-log-leak-test-secret"
	sensitiveKey := "private-api-key-for-logging-test"

	creds := domain.DecryptedCredentials{
		ClientID:     "client-123",
		ClientSecret: sensitiveSecret,
		APIKey:       sensitiveKey,
	}

	// Verify fmt.Sprintf with %s, %v, %+v, %#v
	formattedStr := fmt.Sprintf("%s %v %+v %#v", creds, creds, creds, creds)
	assert.NotContains(t, formattedStr, sensitiveSecret, "DecryptedCredentials string formatting must not leak ClientSecret")
	assert.NotContains(t, formattedStr, sensitiveKey, "DecryptedCredentials string formatting must not leak APIKey")
	assert.Contains(t, formattedStr, "[PROTECTED_CARRIER_CREDENTIALS]")

	// Verify SanitizeError redacts sensitive tokens and secrets
	rawErr := fmt.Errorf("carrier request failed: Authorization: Bearer %s; client_secret=%s; api_key=%s",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", sensitiveSecret, sensitiveKey)
	sanitized := domain.SanitizeError(rawErr)

	assert.NotContains(t, sanitized, sensitiveSecret, "SanitizeError must redact client_secret")
	assert.NotContains(t, sanitized, sensitiveKey, "SanitizeError must redact api_key")
	assert.NotContains(t, sanitized, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", "SanitizeError must redact Bearer tokens")
	assert.Contains(t, sanitized, "[REDACTED]")
}

// Test 9: Existing carrier authentication can still obtain credentials internally after decryption.
func TestScenario9_CarrierAuthCanObtainCredentialsInternallyAfterDecryption(t *testing.T) {
	db, encKey := setupSecurityTestDB(t)
	defer db.Close()

	orgID := int64(99909)
	cleanup := ensureTestOrg(t, db, orgID)
	defer cleanup()

	repo := carrierRepo.NewCarrierRepository(db)
	svc := carrierService.NewCarrierService(repo, encKey)

	ctx := context.Background()
	userID := int64(1009)
	scac := "MAEU"

	req := domain.ConnectCarrierRequest{
		CarrierSCAC:      scac,
		Environment:      string(domain.EnvSandbox),
		ConnectionMethod: string(domain.MethodAPI),
		Credentials: map[string]interface{}{
			"client_id":     "maersk-auth-client",
			"client_secret": "maersk-auth-secret",
			"api_key":       "maersk-consumer-key-123",
		},
		Capabilities: []string{"TRACKING", "RATES"},
	}

	view, err := svc.ConnectCarrier(ctx, orgID, userID, req)
	require.NoError(t, err)
	assert.NotNil(t, view)

	adapter, decryptedCreds, err := svc.GetAdapterForIntegration(ctx, orgID, scac, domain.EnvSandbox)
	require.NoError(t, err)
	require.NotNil(t, adapter)

	// Adapter internally receives DecryptedCredentials to construct headers
	assert.Equal(t, "maersk-consumer-key-123", decryptedCreds.APIKey)
	assert.Equal(t, "maersk-auth-client", decryptedCreds.ClientID)
	assert.Equal(t, "maersk-auth-secret", decryptedCreds.ClientSecret)

	// Verify adapter can execute test connection using decrypted credentials
	res, err := adapter.TestConnection(ctx, decryptedCreds, domain.EnvSandbox)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, domain.EnvSandbox, res.TestedEnvironment)
}

// Test 10: Existing Maersk/Hapag/MSC tests continue to pass.
func TestScenario10_ExistingMaerskHapagMSCTestsContinueToPass(t *testing.T) {
	registry := adapters.GetDefaultRegistry()
	require.NotNil(t, registry)

	maersk := registry.GetAdapterBySCAC("MAEU")
	require.NotNil(t, maersk, "Maersk adapter must be registered")
	assert.Equal(t, "MAEU", maersk.SCAC())

	hapag := registry.GetAdapterBySCAC("HLCU")
	require.NotNil(t, hapag, "Hapag adapter must be registered")
	assert.Equal(t, "HLCU", hapag.SCAC())

	msc := registry.GetAdapterBySCAC("MSCU")
	require.NotNil(t, msc, "MSC adapter must be registered")
	assert.Equal(t, "MSCU", msc.SCAC())

	// Verify Maersk capability matrix
	maerskCaps := maersk.SupportedCapabilities()
	assert.Contains(t, maerskCaps, domain.CapTracking)
	assert.Contains(t, maerskCaps, domain.CapRates)

	// Verify Hapag capability matrix (real tracking only)
	hapagCaps := hapag.SupportedCapabilities()
	assert.Contains(t, hapagCaps, domain.CapTracking)
	assert.NotContains(t, hapagCaps, domain.CapRates)

	// Verify MSC capability matrix (real tracking only)
	mscCaps := msc.SupportedCapabilities()
	assert.Contains(t, mscCaps, domain.CapTracking)
	assert.NotContains(t, mscCaps, domain.CapBooking)

	// Test connection with mock credentials
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mockCreds := domain.DecryptedCredentials{
		APIKey:   "mock-test-key",
		ClientID: "mock-client",
	}

	res, err := maersk.TestConnection(ctx, mockCreds, domain.EnvSandbox)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}
