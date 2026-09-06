package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/freel/backend/internal/carrier/domain"
	"github.com/jmoiron/sqlx"
)

var (
	ErrNotFound       = errors.New("carrier integration record not found")
	ErrDuplicateEntry = errors.New("carrier integration already exists for this organization and environment")
)

type CarrierRepository interface {
	// Providers
	ListProviders(ctx context.Context) ([]domain.CarrierProvider, error)
	GetProviderByCode(ctx context.Context, code string) (*domain.CarrierProvider, error)
	GetProviderBySCAC(ctx context.Context, scac string) (*domain.CarrierProvider, error)

	// Tenant Integrations
	ListIntegrations(ctx context.Context, orgID int64) ([]domain.CarrierIntegration, error)
	GetIntegrationByID(ctx context.Context, orgID int64, id int64) (*domain.CarrierIntegration, error)
	GetIntegrationBySCAC(ctx context.Context, orgID int64, scac string, env domain.Environment) (*domain.CarrierIntegration, error)
	CreateIntegration(ctx context.Context, ci *domain.CarrierIntegration) error
	UpdateIntegration(ctx context.Context, ci *domain.CarrierIntegration) error
	DeleteIntegration(ctx context.Context, orgID int64, id int64) error
	UpdateStatus(ctx context.Context, orgID int64, id int64, status domain.ConnectionStatus, lastError *string) error
	UpdateSyncMetadata(ctx context.Context, orgID int64, id int64, syncStatus string, success bool, lastErr *string) error

	// Audit Logging
	RecordAuditLog(ctx context.Context, orgID int64, userID *int64, action string, resourceID int64, details map[string]interface{}) error

	// Sync Jobs & History (Task 6)
	CreateSyncJob(ctx context.Context, job *domain.IntegrationSyncJob) error
	UpdateSyncJob(ctx context.Context, job *domain.IntegrationSyncJob) error
	GetSyncJobByID(ctx context.Context, orgID int64, jobID int64) (*domain.IntegrationSyncJob, error)
	GetRunningSyncJob(ctx context.Context, orgID int64, integrationID int64) (*domain.IntegrationSyncJob, error)
	ListSyncJobs(ctx context.Context, orgID int64, filter domain.SyncHistoryFilter) ([]domain.IntegrationSyncJob, int, error)

	// Webhook Ingestion & Idempotency (Task 6)
	CreateWebhookEvent(ctx context.Context, evt *domain.CarrierWebhookEvent) error
	GetWebhookEventByFingerprint(ctx context.Context, orgID int64, fingerprint string) (*domain.CarrierWebhookEvent, error)
	UpdateWebhookEventStatus(ctx context.Context, orgID int64, eventID int64, status domain.WebhookEventStatus, errMsg *string) error
	ListWebhookEvents(ctx context.Context, orgID int64, integrationID *int64, limit int) ([]domain.CarrierWebhookEvent, error)
}

type mysqlRepository struct {
	db *sqlx.DB
}

func NewCarrierRepository(db *sqlx.DB) CarrierRepository {
	return &mysqlRepository{db: db}
}

func (r *mysqlRepository) ListProviders(ctx context.Context) ([]domain.CarrierProvider, error) {
	query := `
		SELECT 
			id, code, name, scac, modes, adapter_key, is_active, 
			supported_capabilities, description, documentation_url, logo_url, 
			created_at, updated_at
		FROM carrier_providers
		WHERE is_active = TRUE
		ORDER BY name ASC
	`
	var providers []domain.CarrierProvider
	if err := r.db.SelectContext(ctx, &providers, query); err != nil {
		return nil, fmt.Errorf("failed to list carrier providers: %w", err)
	}

	for i := range providers {
		providers[i].UnmarshalDBFields()
	}
	return providers, nil
}

func (r *mysqlRepository) GetProviderByCode(ctx context.Context, code string) (*domain.CarrierProvider, error) {
	query := `
		SELECT 
			id, code, name, scac, modes, adapter_key, is_active, 
			supported_capabilities, description, documentation_url, logo_url, 
			created_at, updated_at
		FROM carrier_providers
		WHERE code = ? LIMIT 1
	`
	var provider domain.CarrierProvider
	if err := r.db.GetContext(ctx, &provider, query, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get carrier provider by code %s: %w", code, err)
	}
	provider.UnmarshalDBFields()
	return &provider, nil
}

func (r *mysqlRepository) GetProviderBySCAC(ctx context.Context, scac string) (*domain.CarrierProvider, error) {
	query := `
		SELECT 
			id, code, name, scac, modes, adapter_key, is_active, 
			supported_capabilities, description, documentation_url, logo_url, 
			created_at, updated_at
		FROM carrier_providers
		WHERE scac = ? LIMIT 1
	`
	var provider domain.CarrierProvider
	if err := r.db.GetContext(ctx, &provider, query, scac); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get carrier provider by scac %s: %w", scac, err)
	}
	provider.UnmarshalDBFields()
	return &provider, nil
}

func (r *mysqlRepository) ListIntegrations(ctx context.Context, orgID int64) ([]domain.CarrierIntegration, error) {
	query := `
		SELECT 
			ci.id, ci.org_id, ci.carrier_provider_id, ci.carrier_scac, 
			COALESCE(ci.carrier_name, cp.name, ci.carrier_scac) AS carrier_name,
			COALESCE(ci.connection_method, 'API') AS connection_method,
			COALESCE(ci.environment, 'PRODUCTION') AS environment,
			COALESCE(ci.connection_status, 'DISCONNECTED') AS connection_status,
			COALESCE(ci.is_active, 1) AS is_active,
			ci.credentials_json, ci.encrypted_credentials, ci.credential_mask,
			ci.capabilities, ci.config_options, ci.sync_status,
			ci.last_synced_at, ci.last_success_at, ci.last_failure_at, ci.last_error,
			COALESCE(ci.failed_attempts, 0) AS failed_attempts,
			ci.error_details, ci.next_retry_time,
			ci.created_at, ci.updated_at
		FROM carrier_integrations ci
		LEFT JOIN carrier_providers cp ON ci.carrier_provider_id = cp.id OR ci.carrier_scac = cp.scac
		WHERE ci.org_id = ?
		ORDER BY ci.created_at DESC
	`
	var list []domain.CarrierIntegration
	if err := r.db.SelectContext(ctx, &list, query, orgID); err != nil {
		return nil, fmt.Errorf("failed to list carrier integrations for org %d: %w", orgID, err)
	}

	for i := range list {
		list[i].UnmarshalRuntimeFields()
	}
	return list, nil
}

func (r *mysqlRepository) GetIntegrationByID(ctx context.Context, orgID int64, id int64) (*domain.CarrierIntegration, error) {
	query := `
		SELECT 
			ci.id, ci.org_id, ci.carrier_provider_id, ci.carrier_scac, 
			COALESCE(ci.carrier_name, cp.name, ci.carrier_scac) AS carrier_name,
			COALESCE(ci.connection_method, 'API') AS connection_method,
			COALESCE(ci.environment, 'PRODUCTION') AS environment,
			COALESCE(ci.connection_status, 'DISCONNECTED') AS connection_status,
			COALESCE(ci.is_active, 1) AS is_active,
			ci.credentials_json, ci.encrypted_credentials, ci.credential_mask,
			ci.capabilities, ci.config_options, ci.sync_status,
			ci.last_synced_at, ci.last_success_at, ci.last_failure_at, ci.last_error,
			COALESCE(ci.failed_attempts, 0) AS failed_attempts,
			ci.error_details, ci.next_retry_time,
			ci.created_at, ci.updated_at
		FROM carrier_integrations ci
		LEFT JOIN carrier_providers cp ON ci.carrier_provider_id = cp.id OR ci.carrier_scac = cp.scac
		WHERE ci.org_id = ? AND ci.id = ? LIMIT 1
	`
	var ci domain.CarrierIntegration
	if err := r.db.GetContext(ctx, &ci, query, orgID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get carrier integration %d: %w", id, err)
	}
	ci.UnmarshalRuntimeFields()
	return &ci, nil
}

func (r *mysqlRepository) GetIntegrationBySCAC(ctx context.Context, orgID int64, scac string, env domain.Environment) (*domain.CarrierIntegration, error) {
	query := `
		SELECT 
			ci.id, ci.org_id, ci.carrier_provider_id, ci.carrier_scac, 
			COALESCE(ci.carrier_name, cp.name, ci.carrier_scac) AS carrier_name,
			COALESCE(ci.connection_method, 'API') AS connection_method,
			COALESCE(ci.environment, 'PRODUCTION') AS environment,
			COALESCE(ci.connection_status, 'DISCONNECTED') AS connection_status,
			COALESCE(ci.is_active, 1) AS is_active,
			ci.credentials_json, ci.encrypted_credentials, ci.credential_mask,
			ci.capabilities, ci.config_options, ci.sync_status,
			ci.last_synced_at, ci.last_success_at, ci.last_failure_at, ci.last_error,
			COALESCE(ci.failed_attempts, 0) AS failed_attempts,
			ci.error_details, ci.next_retry_time,
			ci.created_at, ci.updated_at
		FROM carrier_integrations ci
		LEFT JOIN carrier_providers cp ON ci.carrier_provider_id = cp.id OR ci.carrier_scac = cp.scac
		WHERE ci.org_id = ? AND ci.carrier_scac = ? AND ci.environment = ? LIMIT 1
	`
	var ci domain.CarrierIntegration
	if err := r.db.GetContext(ctx, &ci, query, orgID, scac, env); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get carrier integration for scac %s: %w", scac, err)
	}
	ci.UnmarshalRuntimeFields()
	return &ci, nil
}

func (r *mysqlRepository) CreateIntegration(ctx context.Context, ci *domain.CarrierIntegration) error {
	query := `
		INSERT INTO carrier_integrations (
			org_id, carrier_provider_id, carrier_scac, carrier_name,
			connection_method, environment, connection_status, is_active,
			credentials_json, encrypted_credentials, credential_mask,
			capabilities, config_options, sync_status
		) VALUES (
			:org_id, :carrier_provider_id, :carrier_scac, :carrier_name,
			:connection_method, :environment, :connection_status, :is_active,
			:credentials_json, :encrypted_credentials, :credential_mask,
			:capabilities, :config_options, :sync_status
		)
	`
	res, err := r.db.NamedExecContext(ctx, query, ci)
	if err != nil {
		return fmt.Errorf("failed to insert carrier integration: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		ci.ID = id
	}
	return nil
}

func (r *mysqlRepository) UpdateIntegration(ctx context.Context, ci *domain.CarrierIntegration) error {
	query := `
		UPDATE carrier_integrations SET
			carrier_provider_id = :carrier_provider_id,
			carrier_name = :carrier_name,
			connection_method = :connection_method,
			environment = :environment,
			connection_status = :connection_status,
			is_active = :is_active,
			credentials_json = :credentials_json,
			encrypted_credentials = :encrypted_credentials,
			credential_mask = :credential_mask,
			capabilities = :capabilities,
			config_options = :config_options,
			last_error = :last_error,
			updated_at = NOW()
		WHERE id = :id AND org_id = :org_id
	`
	_, err := r.db.NamedExecContext(ctx, query, ci)
	if err != nil {
		return fmt.Errorf("failed to update carrier integration %d: %w", ci.ID, err)
	}
	return nil
}

func (r *mysqlRepository) DeleteIntegration(ctx context.Context, orgID int64, id int64) error {
	query := `DELETE FROM carrier_integrations WHERE id = ? AND org_id = ?`
	res, err := r.db.ExecContext(ctx, query, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete carrier integration %d: %w", id, err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *mysqlRepository) UpdateStatus(ctx context.Context, orgID int64, id int64, status domain.ConnectionStatus, lastError *string) error {
	query := `
		UPDATE carrier_integrations SET
			connection_status = ?,
			last_error = ?,
			updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, status, lastError, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to update carrier integration status %d: %w", id, err)
	}
	return nil
}

func (r *mysqlRepository) UpdateSyncMetadata(ctx context.Context, orgID int64, id int64, syncStatus string, success bool, lastErr *string) error {
	var query string
	if success {
		query = `
			UPDATE carrier_integrations SET
				sync_status = ?,
				last_synced_at = NOW(),
				last_success_at = NOW(),
				failed_attempts = 0,
				error_details = NULL,
				connection_status = 'CONNECTED',
				last_error = NULL,
				updated_at = NOW()
			WHERE id = ? AND org_id = ?
		`
		_, err := r.db.ExecContext(ctx, query, syncStatus, id, orgID)
		return err
	}

	query = `
		UPDATE carrier_integrations SET
			sync_status = ?,
			last_synced_at = NOW(),
			last_failure_at = NOW(),
			failed_attempts = failed_attempts + 1,
			error_details = ?,
			last_error = ?,
			connection_status = 'ERROR',
			updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, syncStatus, lastErr, lastErr, id, orgID)
	return err
}

func (r *mysqlRepository) RecordAuditLog(ctx context.Context, orgID int64, userID *int64, action string, resourceID int64, details map[string]interface{}) error {
	detailsJSON := "{}"
	if details != nil {
		if b, err := json.Marshal(details); err == nil {
			detailsJSON = string(b)
		}
	}

	desc := fmt.Sprintf("Carrier integration action %s", action)
	if carrierName, ok := details["carrier_name"].(string); ok && carrierName != "" {
		desc = fmt.Sprintf("Carrier integration action %s for %s", action, carrierName)
	} else if carrierSCAC, ok := details["carrier_scac"].(string); ok && carrierSCAC != "" {
		desc = fmt.Sprintf("Carrier integration action %s for %s", action, carrierSCAC)
	}

	query := `
		INSERT INTO audit_logs (org_id, user_id, action, module, resource_type, resource_id, description, result, details, created_at)
		VALUES (?, ?, ?, 'CARRIER_INTEGRATIONS', 'CARRIER_INTEGRATION', ?, ?, 'SUCCESS', ?, NOW())
	`
	_, err := r.db.ExecContext(ctx, query, orgID, userID, action, fmt.Sprintf("%d", resourceID), desc, detailsJSON)
	return err
}

// ── Sync Jobs & History (Task 6) ─────────────────────────────────────────────

func (r *mysqlRepository) CreateSyncJob(ctx context.Context, job *domain.IntegrationSyncJob) error {
	query := `
		INSERT INTO carrier_sync_jobs (
			org_id, carrier_integration_id, operation, status, started_at,
			records_processed, records_created, records_updated, records_failed,
			error_code, error_message, correlation_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`
	res, err := r.db.ExecContext(ctx, query,
		job.OrgID, job.CarrierIntegrationID, job.Operation, job.Status, job.StartedAt,
		job.RecordsProcessed, job.RecordsCreated, job.RecordsUpdated, job.RecordsFailed,
		job.ErrorCode, job.ErrorMessage, job.CorrelationID,
	)
	if err != nil {
		return fmt.Errorf("failed to create carrier sync job: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		job.ID = id
	}
	return nil
}

func (r *mysqlRepository) UpdateSyncJob(ctx context.Context, job *domain.IntegrationSyncJob) error {
	query := `
		UPDATE carrier_sync_jobs SET
			status = ?,
			completed_at = ?,
			records_processed = ?,
			records_created = ?,
			records_updated = ?,
			records_failed = ?,
			error_code = ?,
			error_message = ?
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		job.Status, job.CompletedAt, job.RecordsProcessed, job.RecordsCreated,
		job.RecordsUpdated, job.RecordsFailed, job.ErrorCode, job.ErrorMessage,
		job.ID, job.OrgID,
	)
	if err != nil {
		return fmt.Errorf("failed to update carrier sync job %d: %w", job.ID, err)
	}
	return nil
}

func (r *mysqlRepository) GetSyncJobByID(ctx context.Context, orgID int64, jobID int64) (*domain.IntegrationSyncJob, error) {
	query := `
		SELECT id, org_id, carrier_integration_id, operation, status, started_at, completed_at,
		       records_processed, records_created, records_updated, records_failed,
		       error_code, error_message, correlation_id, created_at
		FROM carrier_sync_jobs
		WHERE id = ? AND org_id = ?
	`
	var job domain.IntegrationSyncJob
	if err := r.db.GetContext(ctx, &job, query, jobID, orgID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get carrier sync job %d: %w", jobID, err)
	}
	return &job, nil
}

func (r *mysqlRepository) GetRunningSyncJob(ctx context.Context, orgID int64, integrationID int64) (*domain.IntegrationSyncJob, error) {
	query := `
		SELECT id, org_id, carrier_integration_id, operation, status, started_at, completed_at,
		       records_processed, records_created, records_updated, records_failed,
		       error_code, error_message, correlation_id, created_at
		FROM carrier_sync_jobs
		WHERE org_id = ? AND carrier_integration_id = ? AND status = 'RUNNING'
		ORDER BY id DESC
		LIMIT 1
	`
	var job domain.IntegrationSyncJob
	if err := r.db.GetContext(ctx, &job, query, orgID, integrationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to check running sync job: %w", err)
	}
	return &job, nil
}

func (r *mysqlRepository) ListSyncJobs(ctx context.Context, orgID int64, filter domain.SyncHistoryFilter) ([]domain.IntegrationSyncJob, int, error) {
	baseQuery := `FROM carrier_sync_jobs WHERE org_id = ?`
	args := []interface{}{orgID}

	if filter.IntegrationID != nil && *filter.IntegrationID > 0 {
		baseQuery += ` AND carrier_integration_id = ?`
		args = append(args, *filter.IntegrationID)
	}
	if filter.Operation != nil && *filter.Operation != "" {
		baseQuery += ` AND operation = ?`
		args = append(args, *filter.Operation)
	}
	if filter.Status != nil && *filter.Status != "" {
		baseQuery += ` AND status = ?`
		args = append(args, *filter.Status)
	}

	countQuery := `SELECT COUNT(*) ` + baseQuery
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count sync jobs: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := `
		SELECT id, org_id, carrier_integration_id, operation, status, started_at, completed_at,
		       records_processed, records_created, records_updated, records_failed,
		       error_code, error_message, correlation_id, created_at
		` + baseQuery + `
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	var jobs []domain.IntegrationSyncJob
	if err := r.db.SelectContext(ctx, &jobs, selectQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to list sync jobs: %w", err)
	}

	return jobs, total, nil
}

// ── Webhook Ingestion & Idempotency (Task 6) ──────────────────────────────────

func (r *mysqlRepository) CreateWebhookEvent(ctx context.Context, evt *domain.CarrierWebhookEvent) error {
	query := `
		INSERT INTO carrier_webhook_events (
			org_id, carrier_integration_id, carrier_scac, provider_event_id,
			event_type, event_fingerprint, received_at, status, correlation_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), ?, ?, NOW())
	`
	res, err := r.db.ExecContext(ctx, query,
		evt.OrgID, evt.CarrierIntegrationID, evt.CarrierSCAC, evt.ProviderEventID,
		evt.EventType, evt.EventFingerprint, evt.Status, evt.CorrelationID,
	)
	if err != nil {
		return fmt.Errorf("failed to create webhook event: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		evt.ID = id
	}
	return nil
}

func (r *mysqlRepository) GetWebhookEventByFingerprint(ctx context.Context, orgID int64, fingerprint string) (*domain.CarrierWebhookEvent, error) {
	query := `
		SELECT id, org_id, carrier_integration_id, carrier_scac, provider_event_id,
		       event_type, event_fingerprint, received_at, processed_at, status,
		       error_message, correlation_id, created_at
		FROM carrier_webhook_events
		WHERE org_id = ? AND event_fingerprint = ?
		LIMIT 1
	`
	var evt domain.CarrierWebhookEvent
	if err := r.db.GetContext(ctx, &evt, query, orgID, fingerprint); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get webhook event by fingerprint: %w", err)
	}
	return &evt, nil
}

func (r *mysqlRepository) UpdateWebhookEventStatus(ctx context.Context, orgID int64, eventID int64, status domain.WebhookEventStatus, errMsg *string) error {
	query := `
		UPDATE carrier_webhook_events SET
			status = ?,
			processed_at = NOW(),
			error_message = ?
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, status, errMsg, eventID, orgID)
	if err != nil {
		return fmt.Errorf("failed to update webhook event status %d: %w", eventID, err)
	}
	return nil
}

func (r *mysqlRepository) ListWebhookEvents(ctx context.Context, orgID int64, integrationID *int64, limit int) ([]domain.CarrierWebhookEvent, error) {
	query := `
		SELECT id, org_id, carrier_integration_id, carrier_scac, provider_event_id,
		       event_type, event_fingerprint, received_at, processed_at, status,
		       error_message, correlation_id, created_at
		FROM carrier_webhook_events
		WHERE org_id = ?
	`
	args := []interface{}{orgID}
	if integrationID != nil && *integrationID > 0 {
		query += ` AND carrier_integration_id = ?`
		args = append(args, *integrationID)
	}
	query += ` ORDER BY received_at DESC LIMIT ?`
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	args = append(args, limit)

	var evts []domain.CarrierWebhookEvent
	if err := r.db.SelectContext(ctx, &evts, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list webhook events: %w", err)
	}
	return evts, nil
}
