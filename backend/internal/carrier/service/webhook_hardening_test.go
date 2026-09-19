package service_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freel/backend/internal/carrier/domain"
	carrierRepo "github.com/freel/backend/internal/carrier/repository"
	carrierService "github.com/freel/backend/internal/carrier/service"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupWebhookTestContext(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, carrierRepo.CarrierRepository, carrierService.CarrierService) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := carrierRepo.NewCarrierRepository(db)
	svc := carrierService.NewCarrierService(repo, "TestEncryptionKey_32CharactersLong!")
	svc.SetDB(db)

	return db, mock, repo, svc
}

func TestProcessWebhook_TenantResolutionByShipmentReference(t *testing.T) {
	db, mock, _, svc := setupWebhookTestContext(t)
	defer db.Close()

	ctx := context.Background()
	rawBody := []byte(`{"eventId":"MAEU-001","containerNumber":"MSKU9988776","eventType":"GATE_IN"}`)

	// Mock GetProviderByCode
	mock.ExpectQuery("SELECT (.+) FROM carrier_providers WHERE code = \\?").
		WithArgs("MAERSK").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "code", "name", "scac", "modes", "adapter_key", "is_active", "supported_capabilities", "created_at", "updated_at",
		}).AddRow(
			1, "MAERSK", "A.P. Moller – Maersk", "MAEU", `["OCEAN"]`, "MAERSK_ADAPTER", true, `["TRACKING"]`, time.Now(), time.Now(),
		))

	// Mock SELECT * FROM carrier_integrations WHERE carrier_scac = ? AND is_active = 1 (returns 2 tenants: Org 10 and Org 20)
	mock.ExpectQuery("SELECT (.+) FROM carrier_integrations WHERE carrier_scac = \\? AND is_active = 1 AND connection_status = 'CONNECTED'").
		WithArgs("MAEU").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "carrier_scac", "carrier_name", "connection_method",
			"environment", "connection_status", "is_active", "capabilities", "config_options", "failed_attempts", "created_at", "updated_at",
		}).AddRow(
			101, int64(10), "MAEU", "Maersk", "API",
			"PRODUCTION", "CONNECTED", true, `["TRACKING"]`, nil, 0, time.Now(), time.Now(),
		).AddRow(
			202, int64(20), "MAEU", "Maersk", "API",
			"PRODUCTION", "CONNECTED", true, `["TRACKING"]`, nil, 0, time.Now(), time.Now(),
		))

	// Mock matchingOrgID query from shipments table for container MSKU9988776
	mock.ExpectQuery("SELECT s.org_id FROM shipments s INNER JOIN carrier_integrations ci ON ci.org_id = s.org_id").
		WithArgs("MAEU", "MSKU9988776", "MSKU9988776", "MSKU9988776").
		WillReturnRows(sqlmock.NewRows([]string{"org_id"}).AddRow(int64(20)))

	// Mock GetWebhookEventByFingerprint for Org 20 (not duplicate)
	mock.ExpectQuery("SELECT (.+) FROM carrier_webhook_events WHERE org_id = \\? AND event_fingerprint = \\?").
		WithArgs(int64(20), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// Mock CreateWebhookEvent for Org 20
	mock.ExpectExec("INSERT INTO carrier_webhook_events").
		WillReturnResult(sqlmock.NewResult(888, 1))

	// Mock async shipment ID lookup
	mock.ExpectQuery("SELECT id FROM shipments WHERE org_id = \\? AND \\(container_number = \\? OR booking_number = \\? OR mbl_number = \\?\\)").
		WithArgs(int64(20), "MSKU9988776", "MSKU9988776", "MSKU9988776").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(999))

	// Mock sync update status
	mock.ExpectExec("UPDATE carrier_webhook_events SET status = \\?, processed_at = NOW\\(\\) WHERE id = \\? AND org_id = \\?").
		WithArgs(domain.WebhookStatusProcessed, sqlmock.AnyArg(), int64(20)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	var syncedShipmentID int64
	svc.SetTrackingSyncer(func(ctx context.Context, oID int64, sID int64) (int, error) {
		assert.Equal(t, int64(20), oID)
		syncedShipmentID = sID
		return 1, nil
	})

	evt, err := svc.ProcessWebhook(ctx, "MAERSK", rawBody, map[string]string{})
	require.NoError(t, err)
	require.NotNil(t, evt)
	assert.Equal(t, int64(20), evt.OrgID, "Webhook must be securely resolved to Org 20 based on container reference")

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(999), syncedShipmentID, "Shipment #999 must be immediately synchronized upon webhook receipt")
}

func TestProcessWebhook_AmbiguousTenantRejection(t *testing.T) {
	db, mock, _, svc := setupWebhookTestContext(t)
	defer db.Close()

	ctx := context.Background()
	// Unknown container that is not in the system
	rawBody := []byte(`{"eventId":"MAEU-002","containerNumber":"UNKNOWN12345","eventType":"GATE_IN"}`)

	mock.ExpectQuery("SELECT (.+) FROM carrier_providers WHERE code = \\?").
		WithArgs("MAERSK").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "code", "name", "scac", "modes", "adapter_key", "is_active", "supported_capabilities", "created_at", "updated_at",
		}).AddRow(
			1, "MAERSK", "A.P. Moller – Maersk", "MAEU", `["OCEAN"]`, "MAERSK_ADAPTER", true, `["TRACKING"]`, time.Now(), time.Now(),
		))

	// 2 active integrations for MAEU
	mock.ExpectQuery("SELECT (.+) FROM carrier_integrations WHERE carrier_scac = \\? AND is_active = 1 AND connection_status = 'CONNECTED'").
		WithArgs("MAEU").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "carrier_scac", "carrier_name", "connection_method",
			"environment", "connection_status", "is_active", "capabilities", "config_options", "failed_attempts", "created_at", "updated_at",
		}).AddRow(
			101, int64(10), "MAEU", "Maersk", "API",
			"PRODUCTION", "CONNECTED", true, `["TRACKING"]`, nil, 0, time.Now(), time.Now(),
		).AddRow(
			202, int64(20), "MAEU", "Maersk", "API",
			"PRODUCTION", "CONNECTED", true, `["TRACKING"]`, nil, 0, time.Now(), time.Now(),
		))

	// No shipment matches UNKNOWN12345
	mock.ExpectQuery("SELECT s.org_id FROM shipments s INNER JOIN carrier_integrations ci ON ci.org_id = s.org_id").
		WithArgs("MAEU", "UNKNOWN12345", "UNKNOWN12345", "UNKNOWN12345").
		WillReturnRows(sqlmock.NewRows([]string{"org_id"}))

	// ProcessWebhook must fail with ambiguous tenant resolution error to prevent data corruption
	evt, err := svc.ProcessWebhook(ctx, "MAERSK", rawBody, map[string]string{})
	require.Error(t, err)
	assert.Nil(t, evt)
	assert.Contains(t, err.Error(), "ambiguous tenant resolution")
}

func TestProcessWebhook_SignatureVerification(t *testing.T) {
	db, mock, _, svc := setupWebhookTestContext(t)
	defer db.Close()

	ctx := context.Background()
	rawBody := []byte(`{"eventId":"MAEU-003","eventType":"DISCHARGE"}`)
	secret := "secret-token-key-12345"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	validSig := hex.EncodeToString(mac.Sum(nil))

	configJSON := `{"webhook_secret":"secret-token-key-12345"}`

	// Provider mock
	mock.ExpectQuery("SELECT (.+) FROM carrier_providers WHERE code = \\?").
		WithArgs("MAERSK").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "code", "name", "scac", "modes", "adapter_key", "is_active", "supported_capabilities", "created_at", "updated_at",
		}).AddRow(
			1, "MAERSK", "A.P. Moller – Maersk", "MAEU", `["OCEAN"]`, "MAERSK_ADAPTER", true, `["TRACKING"]`, time.Now(), time.Now(),
		))

	// Integration mock with secret configured
	mock.ExpectQuery("SELECT (.+) FROM carrier_integrations WHERE carrier_scac = \\? AND is_active = 1 AND connection_status = 'CONNECTED'").
		WithArgs("MAEU").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "carrier_scac", "carrier_name", "connection_method",
			"environment", "connection_status", "is_active", "capabilities", "config_options", "failed_attempts", "created_at", "updated_at",
		}).AddRow(
			101, int64(10), "MAEU", "Maersk", "API",
			"PRODUCTION", "CONNECTED", true, `["TRACKING"]`, configJSON, 0, time.Now(), time.Now(),
		))

	// Test 1: Missing signature header
	_, errMissingSig := svc.ProcessWebhook(ctx, "MAERSK", rawBody, map[string]string{})
	require.Error(t, errMissingSig)
	assert.Contains(t, errMissingSig.Error(), "missing required carrier webhook signature")

	// Test 2: Invalid signature header
	mock.ExpectQuery("SELECT (.+) FROM carrier_providers WHERE code = \\?").
		WithArgs("MAERSK").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "code", "name", "scac", "modes", "adapter_key", "is_active", "supported_capabilities", "created_at", "updated_at",
		}).AddRow(
			1, "MAERSK", "A.P. Moller – Maersk", "MAEU", `["OCEAN"]`, "MAERSK_ADAPTER", true, `["TRACKING"]`, time.Now(), time.Now(),
		))
	mock.ExpectQuery("SELECT (.+) FROM carrier_integrations WHERE carrier_scac = \\? AND is_active = 1 AND connection_status = 'CONNECTED'").
		WithArgs("MAEU").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "carrier_scac", "carrier_name", "connection_method",
			"environment", "connection_status", "is_active", "capabilities", "config_options", "failed_attempts", "created_at", "updated_at",
		}).AddRow(
			101, int64(10), "MAEU", "Maersk", "API",
			"PRODUCTION", "CONNECTED", true, `["TRACKING"]`, configJSON, 0, time.Now(), time.Now(),
		))

	_, errInvalidSig := svc.ProcessWebhook(ctx, "MAERSK", rawBody, map[string]string{"X-Carrier-Signature": "tampered-bad-sig"})
	require.Error(t, errInvalidSig)
	assert.Contains(t, errInvalidSig.Error(), "invalid carrier webhook signature")

	// Test 3: Valid signature header succeeds
	mock.ExpectQuery("SELECT (.+) FROM carrier_providers WHERE code = \\?").
		WithArgs("MAERSK").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "code", "name", "scac", "modes", "adapter_key", "is_active", "supported_capabilities", "created_at", "updated_at",
		}).AddRow(
			1, "MAERSK", "A.P. Moller – Maersk", "MAEU", `["OCEAN"]`, "MAERSK_ADAPTER", true, `["TRACKING"]`, time.Now(), time.Now(),
		))
	mock.ExpectQuery("SELECT (.+) FROM carrier_integrations WHERE carrier_scac = \\? AND is_active = 1 AND connection_status = 'CONNECTED'").
		WithArgs("MAEU").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "carrier_scac", "carrier_name", "connection_method",
			"environment", "connection_status", "is_active", "capabilities", "config_options", "failed_attempts", "created_at", "updated_at",
		}).AddRow(
			101, int64(10), "MAEU", "Maersk", "API",
			"PRODUCTION", "CONNECTED", true, `["TRACKING"]`, configJSON, 0, time.Now(), time.Now(),
		))
	mock.ExpectQuery("SELECT (.+) FROM carrier_webhook_events WHERE org_id = \\? AND event_fingerprint = \\?").
		WithArgs(int64(10), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("INSERT INTO carrier_webhook_events").
		WillReturnResult(sqlmock.NewResult(501, 1))
	mock.ExpectExec("UPDATE carrier_webhook_events SET status = \\?, processed_at = NOW\\(\\) WHERE id = \\? AND org_id = \\?").
		WithArgs(domain.WebhookStatusProcessed, sqlmock.AnyArg(), int64(10)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	validEvt, errValid := svc.ProcessWebhook(ctx, "MAERSK", rawBody, map[string]string{"X-Carrier-Signature": validSig})
	require.NoError(t, errValid)
	require.NotNil(t, validEvt)
	assert.Equal(t, domain.WebhookStatusPending, validEvt.Status)
}
