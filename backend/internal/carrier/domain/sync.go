package domain

import (
	"time"
)

// IntegrationSyncJob represents a persistent synchronization attempt for an integration.
type IntegrationSyncJob struct {
	ID                   int64         `json:"id" db:"id"`
	OrgID                int64         `json:"org_id" db:"org_id"`
	CarrierIntegrationID int64         `json:"carrier_integration_id" db:"carrier_integration_id"`
	Operation            SyncOperation `json:"operation" db:"operation"`
	Status               SyncStatus    `json:"status" db:"status"`
	StartedAt            time.Time     `json:"started_at" db:"started_at"`
	CompletedAt          *time.Time    `json:"completed_at,omitempty" db:"completed_at"`
	RecordsProcessed     int           `json:"records_processed" db:"records_processed"`
	RecordsCreated       int           `json:"records_created" db:"records_created"`
	RecordsUpdated       int           `json:"records_updated" db:"records_updated"`
	RecordsFailed        int           `json:"records_failed" db:"records_failed"`
	ErrorCode            *string       `json:"error_code,omitempty" db:"error_code"`
	ErrorMessage         *string       `json:"error_message,omitempty" db:"error_message"`
	CorrelationID        string        `json:"correlation_id" db:"correlation_id"`
	CreatedAt            time.Time     `json:"created_at" db:"created_at"`
}

// IntegrationSyncJobView represents a user-facing synchronization log item.
type IntegrationSyncJobView struct {
	ID                   int64         `json:"id"`
	OrgID                int64         `json:"org_id"`
	CarrierIntegrationID int64         `json:"carrier_integration_id"`
	CarrierSCAC          string        `json:"carrier_scac,omitempty"`
	CarrierName          string        `json:"carrier_name,omitempty"`
	Operation            SyncOperation `json:"operation"`
	Status               SyncStatus    `json:"status"`
	StartedAt            time.Time     `json:"started_at"`
	CompletedAt          *time.Time    `json:"completed_at,omitempty"`
	DurationMs           *int64        `json:"duration_ms,omitempty"`
	RecordsProcessed     int           `json:"records_processed"`
	RecordsCreated       int           `json:"records_created"`
	RecordsUpdated       int           `json:"records_updated"`
	RecordsFailed        int           `json:"records_failed"`
	ErrorCode            *string       `json:"error_code,omitempty"`
	ErrorMessage         *string       `json:"error_message,omitempty"`
	CorrelationID        string        `json:"correlation_id"`
	CreatedAt            time.Time     `json:"created_at"`
}

// CarrierWebhookEvent represents a captured inbound carrier webhook for audit and deduplication.
type CarrierWebhookEvent struct {
	ID                   int64              `json:"id" db:"id"`
	OrgID                int64              `json:"org_id" db:"org_id"`
	CarrierIntegrationID int64              `json:"carrier_integration_id" db:"carrier_integration_id"`
	CarrierSCAC          string             `json:"carrier_scac" db:"carrier_scac"`
	ProviderEventID      *string            `json:"provider_event_id,omitempty" db:"provider_event_id"`
	EventType            string             `json:"event_type" db:"event_type"`
	EventFingerprint     string             `json:"event_fingerprint" db:"event_fingerprint"`
	ReceivedAt           time.Time          `json:"received_at" db:"received_at"`
	ProcessedAt          *time.Time         `json:"processed_at,omitempty" db:"processed_at"`
	Status               WebhookEventStatus `json:"status" db:"status"`
	ErrorMessage         *string            `json:"error_message,omitempty" db:"error_message"`
	CorrelationID        string             `json:"correlation_id" db:"correlation_id"`
	CreatedAt            time.Time          `json:"created_at" db:"created_at"`
}

// SyncNowRequest represents the payload to trigger an on-demand synchronization job.
type SyncNowRequest struct {
	Operation     string `json:"operation,omitempty"` // TRACKING, BOOKINGS, FULL_SYNC
	ForceOverride bool   `json:"force,omitempty"`
}

// SyncHistoryFilter specifies filtering parameters for sync history logs.
type SyncHistoryFilter struct {
	IntegrationID *int64
	Operation     *SyncOperation
	Status        *SyncStatus
	Limit         int
	Offset        int
}

// IntegrationHealthDetail provides comprehensive operational health telemetry.
type IntegrationHealthDetail struct {
	IntegrationID       int64                  `json:"integration_id"`
	CarrierSCAC         string                 `json:"carrier_scac"`
	CarrierName         string                 `json:"carrier_name"`
	Environment         Environment            `json:"environment"`
	ConnectionStatus    ConnectionStatus       `json:"connection_status"`
	HealthState         IntegrationHealthState `json:"health_state"` // HEALTHY, ATTENTION, ERROR, DISABLED, DISCONNECTED
	HealthReason        string                 `json:"health_reason"`
	ConsecutiveFailures int                    `json:"consecutive_failures"`
	LastSuccessAt       *time.Time             `json:"last_success_at,omitempty"`
	LastFailureAt       *time.Time             `json:"last_failure_at,omitempty"`
	LastError           *string                `json:"last_error,omitempty"`
	NextRetryTime       *time.Time             `json:"next_retry_time,omitempty"`
	IsSyncing           bool                   `json:"is_syncing"`
	RecentSyncs         []IntegrationSyncJobView `json:"recent_syncs,omitempty"`
}
