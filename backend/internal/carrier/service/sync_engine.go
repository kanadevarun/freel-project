package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/freel/backend/internal/carrier/domain"
	"github.com/freel/backend/internal/carrier/repository"
	audit "github.com/freel/backend/internal/audit"
	auditDomain "github.com/freel/backend/internal/audit/domain"
	"github.com/jmoiron/sqlx"
)

var (
	ErrSyncInProgress = errors.New("a synchronization job is already running for this carrier integration")
)

// TrackingSyncHandler defines the signature for syncing a shipment's tracking data.
type TrackingSyncHandler func(ctx context.Context, orgID int64, shipmentID int64) (newEvents int, err error)

// BookingSyncHandler defines the signature for syncing a booking's status and space allocation.
type BookingSyncHandler func(ctx context.Context, orgID int64, bookingID int64) error

// CarrierSyncEngine orchestrates multi-tenant sync scheduling, manual triggers, idempotency,
// webhook verification, and health telemetry.
type CarrierSyncEngine struct {
	db             *sqlx.DB
	repo           repository.CarrierRepository
	carrierSvc     CarrierService
	trackingSyncer TrackingSyncHandler
	bookingSyncer  BookingSyncHandler
}

// NewCarrierSyncEngine initializes the central CarrierSyncEngine.
func NewCarrierSyncEngine(
	db *sqlx.DB,
	repo repository.CarrierRepository,
	carrierSvc CarrierService,
) *CarrierSyncEngine {
	return &CarrierSyncEngine{
		db:         db,
		repo:       repo,
		carrierSvc: carrierSvc,
	}
}

// SetTrackingSyncer sets the shipment tracking refresh handler.
func (e *CarrierSyncEngine) SetTrackingSyncer(handler TrackingSyncHandler) {
	e.trackingSyncer = handler
}

// SetBookingSyncer sets the booking synchronization handler.
func (e *CarrierSyncEngine) SetBookingSyncer(handler BookingSyncHandler) {
	e.bookingSyncer = handler
}

// SyncNow executes an on-demand, tenant-scoped synchronization for a carrier integration.
func (e *CarrierSyncEngine) SyncNow(ctx context.Context, orgID int64, integrationID int64, req domain.SyncNowRequest) (*domain.IntegrationSyncJobView, error) {
	ci, err := e.repo.GetIntegrationByID(ctx, orgID, integrationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrIntegrationNotFound
		}
		return nil, err
	}

	if !ci.IsEnabled {
		return nil, errors.New("cannot sync a disabled carrier integration; enable the connection first")
	}
	if ci.ConnectionStatus == domain.StatusDisconnected {
		return nil, errors.New("carrier credentials are not configured or connection is disconnected")
	}

	// 1. Prevent duplicate concurrent sync jobs
	if !req.ForceOverride {
		runningJob, err := e.repo.GetRunningSyncJob(ctx, orgID, integrationID)
		if err == nil && runningJob != nil {
			// If started less than 5 minutes ago, reject to prevent hammering carrier APIs
			if time.Since(runningJob.StartedAt) < 5*time.Minute {
				return nil, ErrSyncInProgress
			}
			// Stale job cleanup
			runningJob.Status = domain.SyncStatusCancelled
			now := time.Now().UTC()
			runningJob.CompletedAt = &now
			msg := "Cancelled due to timeout override"
			runningJob.ErrorMessage = &msg
			_ = e.repo.UpdateSyncJob(ctx, runningJob)
		}
	}

	// 2. Determine operation scope
	op := domain.SyncOpFullSync
	if req.Operation != "" {
		if parsedOp, ok := domain.ParseSyncOperation(req.Operation); ok {
			op = parsedOp
		}
	}

	// 3. Create persistent sync job record
	correlationID := fmt.Sprintf("sync-%s-%d-%04d", ci.CarrierSCAC, time.Now().Unix(), rand.Intn(10000))
	job := &domain.IntegrationSyncJob{
		OrgID:                orgID,
		CarrierIntegrationID: integrationID,
		Operation:            op,
		Status:               domain.SyncStatusRunning,
		StartedAt:            time.Now().UTC(),
		CorrelationID:        correlationID,
	}

	if err := e.repo.CreateSyncJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to initialize carrier sync job: %w", err)
	}

	// Update integration status to currently syncing
	_ = e.repo.UpdateSyncMetadata(ctx, orgID, integrationID, "Syncing", true, nil)

	// 4. Execute synchronization pipeline
	processed, created, updated, failed, syncErr := e.executeSync(ctx, ci, op)

	// 5. Finalize Sync Job record
	now := time.Now().UTC()
	job.CompletedAt = &now
	job.RecordsProcessed = processed
	job.RecordsCreated = created
	job.RecordsUpdated = updated
	job.RecordsFailed = failed

	if syncErr != nil {
		job.Status = domain.SyncStatusFailed
		errStr := domain.SanitizeError(syncErr)
		job.ErrorMessage = &errStr
		code := "CARRIER_SYNC_ERROR"
		job.ErrorCode = &code
		_ = e.repo.UpdateSyncJob(ctx, job)
		_ = e.repo.UpdateSyncMetadata(ctx, orgID, integrationID, "Failed", false, &errStr)
	} else if failed > 0 {
		if created > 0 || updated > 0 || (processed > failed) {
			job.Status = domain.SyncStatusPartialSuccess
			msg := fmt.Sprintf("Completed with %d partial failures", failed)
			job.ErrorMessage = &msg
		} else {
			job.Status = domain.SyncStatusFailed
			msg := fmt.Sprintf("All %d records failed to synchronize", failed)
			job.ErrorMessage = &msg
		}
		_ = e.repo.UpdateSyncJob(ctx, job)
		_ = e.repo.UpdateSyncMetadata(ctx, orgID, integrationID, string(job.Status), true, job.ErrorMessage)
	} else {
		job.Status = domain.SyncStatusSuccess
		_ = e.repo.UpdateSyncJob(ctx, job)
		_ = e.repo.UpdateSyncMetadata(ctx, orgID, integrationID, "Completed", true, nil)
	}

	// Record audit activity
	_ = e.repo.RecordAuditLog(ctx, orgID, nil, "CARRIER_SYNC_COMPLETED", integrationID, map[string]interface{}{
		"carrier_scac":      ci.CarrierSCAC,
		"operation":         string(op),
		"status":            string(job.Status),
		"records_processed": processed,
		"records_failed":    failed,
		"correlation_id":    correlationID,
	})

	// Record universal system audit activity
	resStatus := auditDomain.ResultSuccess
	if job.Status == domain.SyncStatusFailed {
		resStatus = auditDomain.ResultFailed
	}
	_, _ = audit.Record(ctx, auditDomain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorType:    auditDomain.ActorTypeSystem,
		ActorName:    "Carrier Sync Engine",
		Action:       auditDomain.ActionSync,
		Module:       auditDomain.ModuleCarrierIntegrations,
		ResourceType: "CARRIER",
		ResourceID:   fmt.Sprintf("%d", integrationID),
		ResourceName: ci.CarrierSCAC,
		Description:  fmt.Sprintf("Automated carrier synchronization completed for %s (%d processed, %d created, %d updated)", ci.CarrierSCAC, processed, created, updated),
		Result:       resStatus,
		Metadata: map[string]interface{}{
			"carrier_scac":      ci.CarrierSCAC,
			"operation":         string(op),
			"records_processed": processed,
			"records_created":   created,
			"records_updated":   updated,
			"records_failed":    failed,
			"correlation_id":    correlationID,
		},
	})

	view := e.toJobView(job, ci.CarrierSCAC, ci.CarrierName)
	return &view, nil
}

// executeSync performs the actual carrier API calls based on integration capabilities.
func (e *CarrierSyncEngine) executeSync(ctx context.Context, ci *domain.CarrierIntegration, op domain.SyncOperation) (processed, created, updated, failed int, err error) {
	hasCap := func(cap domain.Capability) bool {
		for _, c := range ci.Capabilities {
			if c == cap {
				return true
			}
		}
		return false
	}

	// 1. Tracking Sync (Active shipments)
	if (op == domain.SyncOpFullSync || op == domain.SyncOpTracking) && hasCap(domain.CapTracking) {
		tProcessed, tCreated, tUpdated, tFailed, tErr := e.SyncActiveShipments(ctx, ci.OrgID, ci.CarrierSCAC)
		processed += tProcessed
		created += tCreated
		updated += tUpdated
		failed += tFailed
		if tErr != nil {
			log.Printf("[CarrierSyncEngine] Tracking sync encountered error for org %d (%s): %v", ci.OrgID, ci.CarrierSCAC, tErr)
			if err == nil {
				err = tErr
			}
		}
	}

	// 2. Booking Sync (Active / Open bookings)
	if (op == domain.SyncOpFullSync || op == domain.SyncOpBookings) && hasCap(domain.CapBooking) {
		bProcessed, bCreated, bUpdated, bFailed, bErr := e.SyncActiveBookings(ctx, ci.OrgID, ci.CarrierSCAC)
		processed += bProcessed
		created += bCreated
		updated += bUpdated
		failed += bFailed
		if bErr != nil {
			log.Printf("[CarrierSyncEngine] Booking sync encountered error for org %d (%s): %v", ci.OrgID, ci.CarrierSCAC, bErr)
			if err == nil {
				err = bErr
			}
		}
	}

	return processed, created, updated, failed, err
}

// SyncActiveShipments synchronizes active and recent shipments for the given organization and carrier.
func (e *CarrierSyncEngine) SyncActiveShipments(ctx context.Context, orgID int64, carrierSCAC string) (processed, created, updated, failed int, err error) {
	if e.db == nil {
		return 0, 0, 0, 0, nil
	}

	// Query active shipments matching this carrier SCAC or name
	var shipmentIDs []int64
	query := `
		SELECT s.id FROM shipments s
		LEFT JOIN bookings b ON s.booking_id = b.id
		WHERE s.org_id = ?
		  AND (s.carrier_scac = ? OR b.carrier_name LIKE ?)
		  AND s.status IN ('BOOKED', 'IN_TRANSIT', 'PENDING_CONFIRMATION', 'BOOKING_PENDING', 'ARRIVED')
		ORDER BY s.id DESC
		LIMIT 50
	`
	wildcard := "%" + carrierSCAC + "%"
	if dbErr := e.db.SelectContext(ctx, &shipmentIDs, query, orgID, carrierSCAC, wildcard); dbErr != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to query active shipments: %w", dbErr)
	}

	if len(shipmentIDs) == 0 {
		return 0, 0, 0, 0, nil
	}

	for _, sid := range shipmentIDs {
		processed++
		if e.trackingSyncer != nil {
			newEvents, sErr := e.trackingSyncer(ctx, orgID, sid)
			if sErr != nil {
				log.Printf("[CarrierSyncEngine] Failed tracking sync for shipment #%d: %v", sid, sErr)
				failed++
			} else {
				if newEvents > 0 {
					created += newEvents
				} else {
					updated++
				}
			}
		} else {
			updated++
		}
	}

	return processed, created, updated, failed, nil
}

// SyncActiveBookings synchronizes active bookings with carrier references.
func (e *CarrierSyncEngine) SyncActiveBookings(ctx context.Context, orgID int64, carrierSCAC string) (processed, created, updated, failed int, err error) {
	if e.db == nil {
		return 0, 0, 0, 0, nil
	}

	var bookingIDs []int64
	query := `
		SELECT id FROM bookings
		WHERE org_id = ?
		  AND (carrier_scac = ? OR carrier_name LIKE ?)
		  AND carrier_booking_reference IS NOT NULL
		  AND status NOT IN ('CANCELLED', 'COMPLETED')
		ORDER BY id DESC
		LIMIT 50
	`
	wildcard := "%" + carrierSCAC + "%"
	if dbErr := e.db.SelectContext(ctx, &bookingIDs, query, orgID, carrierSCAC, wildcard); dbErr != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to query active bookings: %w", dbErr)
	}

	if len(bookingIDs) == 0 {
		return 0, 0, 0, 0, nil
	}

	for _, bid := range bookingIDs {
		processed++
		if e.bookingSyncer != nil {
			if bErr := e.bookingSyncer(ctx, orgID, bid); bErr != nil {
				log.Printf("[CarrierSyncEngine] Failed booking sync for booking #%d: %v", bid, bErr)
				failed++
			} else {
				updated++
			}
		} else {
			updated++
		}
	}

	return processed, created, updated, failed, nil
}

// GetSyncHistory retrieves paginated sync history for a carrier integration.
func (e *CarrierSyncEngine) GetSyncHistory(ctx context.Context, orgID int64, integrationID int64, limit, offset int) ([]domain.IntegrationSyncJobView, int, error) {
	ci, err := e.repo.GetIntegrationByID(ctx, orgID, integrationID)
	if err != nil {
		return nil, 0, ErrIntegrationNotFound
	}

	filter := domain.SyncHistoryFilter{
		IntegrationID: &integrationID,
		Limit:         limit,
		Offset:        offset,
	}

	jobs, total, err := e.repo.ListSyncJobs(ctx, orgID, filter)
	if err != nil {
		return nil, 0, err
	}

	views := make([]domain.IntegrationSyncJobView, len(jobs))
	for i, j := range jobs {
		views[i] = e.toJobView(&j, ci.CarrierSCAC, ci.CarrierName)
	}

	return views, total, nil
}

// GetSyncJob retrieves a specific sync job by ID with full error diagnostics.
func (e *CarrierSyncEngine) GetSyncJob(ctx context.Context, orgID int64, jobID int64) (*domain.IntegrationSyncJobView, error) {
	job, err := e.repo.GetSyncJobByID(ctx, orgID, jobID)
	if err != nil || job == nil {
		return nil, errors.New("sync job record not found")
	}

	ci, _ := e.repo.GetIntegrationByID(ctx, orgID, job.CarrierIntegrationID)
	scac := ""
	var name *string
	if ci != nil {
		scac = ci.CarrierSCAC
		name = ci.CarrierName
	}

	view := e.toJobView(job, scac, name)
	return &view, nil
}

// GetIntegrationHealth computes real-time health telemetry for an integration.
func (e *CarrierSyncEngine) GetIntegrationHealth(ctx context.Context, orgID int64, integrationID int64) (*domain.IntegrationHealthDetail, error) {
	ci, err := e.repo.GetIntegrationByID(ctx, orgID, integrationID)
	if err != nil {
		return nil, ErrIntegrationNotFound
	}

	healthState, healthReason := e.computeHealth(ci)

	// Fetch recent 5 sync jobs
	recentJobs, _, _ := e.repo.ListSyncJobs(ctx, orgID, domain.SyncHistoryFilter{
		IntegrationID: &integrationID,
		Limit:         5,
	})

	jobViews := make([]domain.IntegrationSyncJobView, len(recentJobs))
	for i, j := range recentJobs {
		jobViews[i] = e.toJobView(&j, ci.CarrierSCAC, ci.CarrierName)
	}

	runningJob, _ := e.repo.GetRunningSyncJob(ctx, orgID, integrationID)
	isSyncing := runningJob != nil

	carrierName := ci.CarrierSCAC
	if ci.CarrierName != nil && *ci.CarrierName != "" {
		carrierName = *ci.CarrierName
	}

	return &domain.IntegrationHealthDetail{
		IntegrationID:       ci.ID,
		CarrierSCAC:         ci.CarrierSCAC,
		CarrierName:         carrierName,
		Environment:         ci.Environment,
		ConnectionStatus:    ci.ConnectionStatus,
		HealthState:         healthState,
		HealthReason:        healthReason,
		ConsecutiveFailures: ci.FailedAttempts,
		LastSuccessAt:       ci.LastSuccessAt,
		LastFailureAt:       ci.LastFailureAt,
		LastError:           ci.LastError,
		NextRetryTime:       ci.NextRetryTime,
		IsSyncing:           isSyncing,
		RecentSyncs:         jobViews,
	}, nil
}

// ProcessWebhook receives and processes an inbound carrier webhook with HMAC validation and idempotency.
func (e *CarrierSyncEngine) ProcessWebhook(ctx context.Context, providerCode string, rawBody []byte, headers map[string]string) (*domain.CarrierWebhookEvent, error) {
	normalizedCode := strings.ToUpper(strings.TrimSpace(providerCode))
	provider, err := e.repo.GetProviderByCode(ctx, normalizedCode)
	if err != nil || provider == nil {
		return nil, fmt.Errorf("unknown carrier provider: %s", providerCode)
	}

	// Generate deterministic event fingerprint for idempotency deduplication
	hasher := sha256.New()
	hasher.Write([]byte(normalizedCode))
	hasher.Write(rawBody)
	fingerprint := hex.EncodeToString(hasher.Sum(nil))

	// Find the matching carrier integration
	var ci domain.CarrierIntegration
	query := `
		SELECT * FROM carrier_integrations
		WHERE carrier_scac = ? AND is_active = 1 AND connection_status = 'CONNECTED'
		LIMIT 1
	`
	if dbErr := e.db.GetContext(ctx, &ci, query, provider.SCAC); dbErr != nil {
		return nil, fmt.Errorf("no active carrier integration found for provider %s (%s)", providerCode, provider.SCAC)
	}
	ci.UnmarshalRuntimeFields()

	// 1. Idempotency check: look up existing event
	existingEvt, err := e.repo.GetWebhookEventByFingerprint(ctx, ci.OrgID, fingerprint)
	if err == nil && existingEvt != nil {
		log.Printf("[CarrierSyncEngine] Webhook event %s is DUPLICATE. Returning existing event record.", fingerprint)
		existingEvt.Status = domain.WebhookStatusDuplicate
		return existingEvt, nil
	}

	// 2. Validate webhook signature if secret configured
	if sig, ok := headers["X-Carrier-Signature"]; ok {
		if secret, ok := ci.Config["webhook_secret"].(string); ok && secret != "" {
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write(rawBody)
			expectedSig := hex.EncodeToString(mac.Sum(nil))
			if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
				return nil, errors.New("invalid carrier webhook signature")
			}
		}
	}

	correlationID := fmt.Sprintf("wh-%s-%d-%04d", provider.SCAC, time.Now().Unix(), rand.Intn(10000))
	var eventType string
	var providerEvtID *string

	// Parse generic JSON event fields
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(rawBody, &jsonMap); err == nil {
		if t, ok := jsonMap["eventType"].(string); ok {
			eventType = t
		} else if t, ok := jsonMap["event_type"].(string); ok {
			eventType = t
		}
		if id, ok := jsonMap["eventId"].(string); ok {
			providerEvtID = &id
		} else if id, ok := jsonMap["id"].(string); ok {
			providerEvtID = &id
		}
	}
	if eventType == "" {
		eventType = "CARRIER_UPDATE"
	}

	evt := &domain.CarrierWebhookEvent{
		OrgID:                ci.OrgID,
		CarrierIntegrationID: ci.ID,
		CarrierSCAC:          provider.SCAC,
		ProviderEventID:      providerEvtID,
		EventType:            eventType,
		EventFingerprint:     fingerprint,
		ReceivedAt:           time.Now().UTC(),
		Status:               domain.WebhookStatusPending,
		CorrelationID:        correlationID,
	}

	if err := e.repo.CreateWebhookEvent(ctx, evt); err != nil {
		return nil, fmt.Errorf("failed to record webhook event: %w", err)
	}

	// 3. Process tracking / booking payload asynchronously or in pipeline
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, _, _, _, syncErr := e.executeSync(bgCtx, &ci, domain.SyncOpFullSync)
		if syncErr != nil {
			errStr := syncErr.Error()
			_ = e.repo.UpdateWebhookEventStatus(bgCtx, ci.OrgID, evt.ID, domain.WebhookStatusFailed, &errStr)
		} else {
			_ = e.repo.UpdateWebhookEventStatus(bgCtx, ci.OrgID, evt.ID, domain.WebhookStatusProcessed, nil)
		}
	}()

	return evt, nil
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func (e *CarrierSyncEngine) computeHealth(ci *domain.CarrierIntegration) (domain.IntegrationHealthState, string) {
	if !ci.IsEnabled {
		return domain.HealthDisabled, "Integration is disabled by user"
	}
	if ci.ConnectionStatus == domain.StatusDisconnected {
		return domain.HealthDisconnected, "Carrier credentials not configured"
	}
	if ci.FailedAttempts >= 5 {
		reason := "Persistent sync failures detected (5+ attempts failed)"
		if ci.LastError != nil && *ci.LastError != "" {
			reason = *ci.LastError
		}
		return domain.HealthError, reason
	}
	if ci.FailedAttempts > 0 {
		reason := fmt.Sprintf("Recent sync failure (%d consecutive failures)", ci.FailedAttempts)
		if ci.LastError != nil && *ci.LastError != "" {
			reason = *ci.LastError
		}
		return domain.HealthAttention, reason
	}
	if ci.ConnectionStatus == domain.StatusError {
		reason := "Carrier connection test failed"
		if ci.LastError != nil && *ci.LastError != "" {
			reason = *ci.LastError
		}
		return domain.HealthAttention, reason
	}
	return domain.HealthHealthy, "Connection active and healthy"
}

func (e *CarrierSyncEngine) toJobView(job *domain.IntegrationSyncJob, scac string, name *string) domain.IntegrationSyncJobView {
	carrierName := scac
	if name != nil && *name != "" {
		carrierName = *name
	}

	var durationMs *int64
	if job.CompletedAt != nil {
		diff := job.CompletedAt.Sub(job.StartedAt).Milliseconds()
		durationMs = &diff
	}

	return domain.IntegrationSyncJobView{
		ID:                   job.ID,
		OrgID:                job.OrgID,
		CarrierIntegrationID: job.CarrierIntegrationID,
		CarrierSCAC:          scac,
		CarrierName:          carrierName,
		Operation:            job.Operation,
		Status:               job.Status,
		StartedAt:            job.StartedAt,
		CompletedAt:          job.CompletedAt,
		DurationMs:           durationMs,
		RecordsProcessed:     job.RecordsProcessed,
		RecordsCreated:       job.RecordsCreated,
		RecordsUpdated:       job.RecordsUpdated,
		RecordsFailed:        job.RecordsFailed,
		ErrorCode:            job.ErrorCode,
		ErrorMessage:         job.ErrorMessage,
		CorrelationID:        job.CorrelationID,
		CreatedAt:            job.CreatedAt,
	}
}
