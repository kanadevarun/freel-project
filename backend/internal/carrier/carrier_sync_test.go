package carrier_test

import (
	"context"
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

func setupSyncTestContext(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, carrierRepo.CarrierRepository, carrierService.CarrierService) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := carrierRepo.NewCarrierRepository(db)
	svc := carrierService.NewCarrierService(repo, "TestEncryptionKey_32CharactersLong!")
	svc.SetDB(db)

	return db, mock, repo, svc
}

func TestCarrierSyncEngine_SyncNow_Success(t *testing.T) {
	db, mock, _, svc := setupSyncTestContext(t)
	defer db.Close()

	ctx := context.Background()
	orgID := int64(18133)
	integrationID := int64(42)

	// Mock GetIntegrationByID
	capsJSON := `["TRACKING","BOOKING"]`
	mock.ExpectQuery("SELECT (.+) FROM carrier_integrations ci LEFT JOIN carrier_providers cp (.+) WHERE ci.org_id = \\? AND ci.id = \\?").
		WithArgs(orgID, integrationID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "carrier_provider_id", "carrier_scac", "carrier_name", "connection_method",
			"environment", "connection_status", "is_active", "credentials_json", "encrypted_credentials",
			"credential_mask", "capabilities", "config_options", "sync_status", "last_synced_at",
			"last_success_at", "last_failure_at", "last_error", "failed_attempts", "error_details",
			"next_retry_time", "created_at", "updated_at",
		}).AddRow(
			integrationID, orgID, nil, "MAEU", "A.P. Moller – Maersk", "API",
			"PRODUCTION", "CONNECTED", true, nil, nil,
			nil, capsJSON, nil, nil, nil,
			nil, nil, nil, 0, nil,
			nil, time.Now(), time.Now(),
		))

	// Mock GetRunningSyncJob (returns none)
	mock.ExpectQuery("SELECT (.+) FROM carrier_sync_jobs WHERE org_id = \\? AND carrier_integration_id = \\? AND status = 'RUNNING'").
		WithArgs(orgID, integrationID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// Mock CreateSyncJob
	mock.ExpectExec("INSERT INTO carrier_sync_jobs").
		WillReturnResult(sqlmock.NewResult(101, 1))

	// Mock UpdateSyncMetadata ("Syncing")
	mock.ExpectExec("UPDATE carrier_integrations SET sync_status = \\?, last_synced_at = NOW\\(\\)").
		WithArgs("Syncing", integrationID, orgID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock SyncActiveShipments DB query
	mock.ExpectQuery("SELECT s.id FROM shipments s LEFT JOIN bookings b ON s.booking_id = b.id WHERE s.org_id = \\? AND \\(s.carrier_scac = \\? OR b.carrier_name LIKE \\?\\)").
		WithArgs(orgID, "MAEU", "%MAEU%").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))

	// Mock SyncActiveBookings DB query
	mock.ExpectQuery("SELECT id FROM bookings WHERE org_id = \\? AND \\(carrier_scac = \\? OR carrier_name LIKE \\?\\)").
		WithArgs(orgID, "MAEU", "%MAEU%").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	// Mock UpdateSyncJob (SUCCESS)
	mock.ExpectExec("UPDATE carrier_sync_jobs SET status = \\?, completed_at = \\?").
		WillReturnResult(sqlmock.NewResult(101, 1))

	// Mock UpdateSyncMetadata ("Completed")
	mock.ExpectExec("UPDATE carrier_integrations SET sync_status = \\?, last_synced_at = NOW\\(\\), last_success_at = NOW\\(\\), failed_attempts = 0").
		WithArgs("Completed", integrationID, orgID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock RecordAuditLog
	mock.ExpectExec("INSERT INTO audit_logs").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Register test tracking and booking sync handlers
	svc.SetTrackingSyncer(func(ctx context.Context, oID int64, sID int64) (int, error) {
		assert.Equal(t, orgID, oID)
		return 1, nil
	})
	svc.SetBookingSyncer(func(ctx context.Context, oID int64, bID int64) error {
		assert.Equal(t, orgID, oID)
		return nil
	})

	jobView, err := svc.SyncNow(ctx, orgID, integrationID, domain.SyncNowRequest{
		Operation: "FULL_SYNC",
	})
	require.NoError(t, err)
	require.NotNil(t, jobView)

	assert.Equal(t, int64(101), jobView.ID)
	assert.Equal(t, orgID, jobView.OrgID)
	assert.Equal(t, integrationID, jobView.CarrierIntegrationID)
	assert.Equal(t, domain.SyncStatusSuccess, jobView.Status)
	assert.Equal(t, 3, jobView.RecordsProcessed) // 2 shipments + 1 booking
	assert.Equal(t, 0, jobView.RecordsFailed)
	assert.NotEmpty(t, jobView.CorrelationID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCarrierSyncEngine_SyncNow_PreventDuplicateSync(t *testing.T) {
	db, mock, _, svc := setupSyncTestContext(t)
	defer db.Close()

	ctx := context.Background()
	orgID := int64(18133)
	integrationID := int64(42)

	// Mock GetIntegrationByID
	mock.ExpectQuery("SELECT (.+) FROM carrier_integrations ci LEFT JOIN carrier_providers cp (.+) WHERE ci.org_id = \\? AND ci.id = \\?").
		WithArgs(orgID, integrationID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "carrier_provider_id", "carrier_scac", "carrier_name", "connection_method",
			"environment", "connection_status", "is_active", "credentials_json", "encrypted_credentials",
			"credential_mask", "capabilities", "config_options", "sync_status", "last_synced_at",
			"last_success_at", "last_failure_at", "last_error", "failed_attempts", "error_details",
			"next_retry_time", "created_at", "updated_at",
		}).AddRow(
			integrationID, orgID, nil, "MAEU", "Maersk", "API",
			"PRODUCTION", "CONNECTED", true, nil, nil,
			nil, `["TRACKING"]`, nil, nil, nil,
			nil, nil, nil, 0, nil,
			nil, time.Now(), time.Now(),
		))

	// Mock GetRunningSyncJob (returns an actively running job started 1 minute ago)
	mock.ExpectQuery("SELECT (.+) FROM carrier_sync_jobs WHERE org_id = \\? AND carrier_integration_id = \\? AND status = 'RUNNING'").
		WithArgs(orgID, integrationID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "carrier_integration_id", "operation", "status", "started_at", "correlation_id", "created_at",
		}).AddRow(
			88, orgID, integrationID, "TRACKING", "RUNNING", time.Now().Add(-1*time.Minute), "sync-corr-1", time.Now(),
		))

	// Trigger sync should be rejected with ErrSyncInProgress
	jobView, err := svc.SyncNow(ctx, orgID, integrationID, domain.SyncNowRequest{Operation: "TRACKING"})
	require.Error(t, err)
	assert.Nil(t, jobView)
	assert.Equal(t, carrierService.ErrSyncInProgress, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCarrierSyncEngine_HealthComputation(t *testing.T) {
	db, mock, _, svc := setupSyncTestContext(t)
	defer db.Close()

	ctx := context.Background()
	orgID := int64(18133)
	integrationID := int64(42)

	cols := []string{
		"id", "org_id", "carrier_provider_id", "carrier_scac", "carrier_name", "connection_method",
		"environment", "connection_status", "is_active", "credentials_json", "encrypted_credentials",
		"credential_mask", "capabilities", "config_options", "sync_status", "last_synced_at",
		"last_success_at", "last_failure_at", "last_error", "failed_attempts", "error_details",
		"next_retry_time", "created_at", "updated_at",
	}

	// Case 1: Healthy integration (0 failed attempts)
	mock.ExpectQuery("SELECT (.+) FROM carrier_integrations ci LEFT JOIN carrier_providers cp (.+) WHERE ci.org_id = \\? AND ci.id = \\?").
		WithArgs(orgID, integrationID).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(
			integrationID, orgID, nil, "MAEU", "Maersk", "API",
			"PRODUCTION", "CONNECTED", true, nil, nil,
			nil, `["TRACKING"]`, nil, nil, nil,
			nil, nil, nil, 0, nil,
			nil, time.Now(), time.Now(),
		))

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM carrier_sync_jobs WHERE org_id = \\? AND carrier_integration_id = \\?").
		WithArgs(orgID, integrationID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT (.+) FROM carrier_sync_jobs WHERE org_id = \\? AND carrier_integration_id = \\? ORDER BY started_at DESC LIMIT \\? OFFSET \\?").
		WithArgs(orgID, integrationID, 5, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery("SELECT (.+) FROM carrier_sync_jobs WHERE org_id = \\? AND carrier_integration_id = \\? AND status = 'RUNNING'").
		WithArgs(orgID, integrationID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	health, err := svc.GetIntegrationHealth(ctx, orgID, integrationID)
	require.NoError(t, err)
	require.NotNil(t, health)
	assert.Equal(t, domain.HealthHealthy, health.HealthState)
	assert.Equal(t, 0, health.ConsecutiveFailures)

	// Case 2: Attention state (2 failed attempts)
	mock.ExpectQuery("SELECT (.+) FROM carrier_integrations ci LEFT JOIN carrier_providers cp (.+) WHERE ci.org_id = \\? AND ci.id = \\?").
		WithArgs(orgID, integrationID).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(
			integrationID, orgID, nil, "MAEU", "Maersk", "API",
			"PRODUCTION", "CONNECTED", true, nil, nil,
			nil, `["TRACKING"]`, nil, nil, nil,
			nil, nil, nil, 2, nil,
			nil, time.Now(), time.Now(),
		))

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM carrier_sync_jobs WHERE org_id = \\? AND carrier_integration_id = \\?").
		WithArgs(orgID, integrationID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT (.+) FROM carrier_sync_jobs WHERE org_id = \\? AND carrier_integration_id = \\? ORDER BY started_at DESC LIMIT \\? OFFSET \\?").
		WithArgs(orgID, integrationID, 5, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery("SELECT (.+) FROM carrier_sync_jobs WHERE org_id = \\? AND carrier_integration_id = \\? AND status = 'RUNNING'").
		WithArgs(orgID, integrationID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	health2, err := svc.GetIntegrationHealth(ctx, orgID, integrationID)
	require.NoError(t, err)
	assert.Equal(t, domain.HealthAttention, health2.HealthState)
	assert.Equal(t, 2, health2.ConsecutiveFailures)

	// Case 3: Error / Action Required state (5 failed attempts)
	mock.ExpectQuery("SELECT (.+) FROM carrier_integrations ci LEFT JOIN carrier_providers cp (.+) WHERE ci.org_id = \\? AND ci.id = \\?").
		WithArgs(orgID, integrationID).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(
			integrationID, orgID, nil, "MAEU", "Maersk", "API",
			"PRODUCTION", "CONNECTED", true, nil, nil,
			nil, `["TRACKING"]`, nil, nil, nil,
			nil, nil, nil, 5, nil,
			nil, time.Now(), time.Now(),
		))

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM carrier_sync_jobs WHERE org_id = \\? AND carrier_integration_id = \\?").
		WithArgs(orgID, integrationID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT (.+) FROM carrier_sync_jobs WHERE org_id = \\? AND carrier_integration_id = \\? ORDER BY started_at DESC LIMIT \\? OFFSET \\?").
		WithArgs(orgID, integrationID, 5, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery("SELECT (.+) FROM carrier_sync_jobs WHERE org_id = \\? AND carrier_integration_id = \\? AND status = 'RUNNING'").
		WithArgs(orgID, integrationID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	health3, err := svc.GetIntegrationHealth(ctx, orgID, integrationID)
	require.NoError(t, err)
	assert.Equal(t, domain.HealthError, health3.HealthState)
	assert.Equal(t, 5, health3.ConsecutiveFailures)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCarrierSyncEngine_ProcessWebhook_Idempotency(t *testing.T) {
	db, mock, _, svc := setupSyncTestContext(t)
	defer db.Close()

	ctx := context.Background()
	orgID := int64(18133)
	integrationID := int64(42)

	rawBody := []byte(`{"eventId":"evt-123456","eventType":"EQUIPMENT_GATE_IN","containerNumber":"MSKU9012345"}`)

	// Mock GetProviderByCode
	mock.ExpectQuery("SELECT (.+) FROM carrier_providers WHERE code = \\?").
		WithArgs("MAERSK").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "code", "name", "scac", "modes", "adapter_key", "is_active", "supported_capabilities", "created_at", "updated_at",
		}).AddRow(
			1, "MAERSK", "A.P. Moller – Maersk", "MAEU", `["OCEAN"]`, "MAERSK_ADAPTER", true, `["TRACKING"]`, time.Now(), time.Now(),
		))

	// Mock Get active integration for provider
	mock.ExpectQuery("SELECT (.+) FROM carrier_integrations WHERE carrier_scac = \\? AND is_active = 1 AND connection_status = 'CONNECTED'").
		WithArgs("MAEU").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "carrier_scac", "carrier_name", "connection_method",
			"environment", "connection_status", "is_active", "capabilities", "failed_attempts", "created_at", "updated_at",
		}).AddRow(
			integrationID, orgID, "MAEU", "Maersk", "API",
			"PRODUCTION", "CONNECTED", true, `["TRACKING"]`, 0, time.Now(), time.Now(),
		))

	// Mock GetWebhookEventByFingerprint (returns an existing processed event for deduplication)
	mock.ExpectQuery("SELECT (.+) FROM carrier_webhook_events WHERE org_id = \\? AND event_fingerprint = \\?").
		WithArgs(orgID, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_id", "carrier_integration_id", "carrier_scac", "provider_event_id",
			"event_type", "event_fingerprint", "received_at", "processed_at", "status", "correlation_id", "created_at",
		}).AddRow(
			55, orgID, integrationID, "MAEU", "evt-123456",
			"EQUIPMENT_GATE_IN", "mock-fingerprint", time.Now().Add(-10*time.Minute), time.Now().Add(-9*time.Minute), "PROCESSED", "wh-corr-55", time.Now(),
		))

	evt, err := svc.ProcessWebhook(ctx, "MAERSK", rawBody, map[string]string{})
	require.NoError(t, err)
	require.NotNil(t, evt)
	assert.Equal(t, int64(55), evt.ID)
	assert.Equal(t, domain.WebhookStatusDuplicate, evt.Status)

	require.NoError(t, mock.ExpectationsWereMet())
}
