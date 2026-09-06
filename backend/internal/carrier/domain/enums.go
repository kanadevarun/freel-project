package domain

import (
	"fmt"
	"strings"
)

// Environment represents the target carrier runtime environment.
type Environment string

const (
	EnvProduction Environment = "PRODUCTION"
	EnvSandbox    Environment = "SANDBOX"
)

func ParseEnvironment(s string) (Environment, error) {
	upper := strings.ToUpper(strings.TrimSpace(s))
	switch upper {
	case "PRODUCTION", "PROD":
		return EnvProduction, nil
	case "SANDBOX", "TEST", "DEV", "STAGING":
		return EnvSandbox, nil
	default:
		return "", fmt.Errorf("invalid environment: '%s', allowed: PRODUCTION, SANDBOX", s)
	}
}

// ConnectionMethod represents how LogisticsHQ interfaces with the carrier.
type ConnectionMethod string

const (
	MethodAPI     ConnectionMethod = "API"
	MethodWebhook ConnectionMethod = "WEBHOOK"
	MethodEDI     ConnectionMethod = "EDI"
	MethodSFTP    ConnectionMethod = "SFTP"
	MethodManual  ConnectionMethod = "MANUAL"
)

func ParseConnectionMethod(s string) (ConnectionMethod, error) {
	upper := strings.ToUpper(strings.TrimSpace(s))
	switch upper {
	case "API", "API_INTEGRATION":
		return MethodAPI, nil
	case "WEBHOOK":
		return MethodWebhook, nil
	case "EDI":
		return MethodEDI, nil
	case "SFTP":
		return MethodSFTP, nil
	case "MANUAL":
		return MethodManual, nil
	default:
		return "", fmt.Errorf("invalid connection method: '%s', allowed: API, WEBHOOK, EDI, SFTP, MANUAL", s)
	}
}

// ConnectionStatus represents the health and connectivity state of an integration.
type ConnectionStatus string

const (
	StatusDisconnected ConnectionStatus = "DISCONNECTED"
	StatusConnecting   ConnectionStatus = "CONNECTING"
	StatusConnected    ConnectionStatus = "CONNECTED"
	StatusError        ConnectionStatus = "ERROR"
	StatusDisabled     ConnectionStatus = "DISABLED"
)

func ParseConnectionStatus(s string) ConnectionStatus {
	upper := strings.ToUpper(strings.TrimSpace(s))
	switch upper {
	case "CONNECTED":
		return StatusConnected
	case "CONNECTING":
		return StatusConnecting
	case "ERROR", "NEEDS_ATTENTION", "FAILED":
		return StatusError
	case "DISABLED":
		return StatusDisabled
	default:
		return StatusDisconnected
	}
}

// Capability represents an operational feature offered by a carrier.
type Capability string

const (
	CapTracking      Capability = "TRACKING"
	CapRates         Capability = "RATES"
	CapContractRates Capability = "CONTRACT_RATES"
	CapSpotRates     Capability = "SPOT_RATES"
	CapBooking       Capability = "BOOKING"
	CapDocuments     Capability = "DOCUMENTS"
)

func ParseCapability(s string) (Capability, bool) {
	upper := strings.ToUpper(strings.TrimSpace(s))
	switch upper {
	case "TRACKING":
		return CapTracking, true
	case "RATES":
		return CapRates, true
	case "CONTRACT_RATES", "CONTRACTRATES":
		return CapContractRates, true
	case "SPOT_RATES", "SPOTRATES":
		return CapSpotRates, true
	case "BOOKING":
		return CapBooking, true
	case "DOCUMENTS":
		return CapDocuments, true
	default:
		return "", false
	}
}

// SyncOperation represents the domain scope of a carrier synchronization job.
type SyncOperation string

const (
	SyncOpTracking  SyncOperation = "TRACKING"
	SyncOpRates     SyncOperation = "RATES"
	SyncOpBookings  SyncOperation = "BOOKINGS"
	SyncOpDocuments SyncOperation = "DOCUMENTS"
	SyncOpFullSync  SyncOperation = "FULL_SYNC"
)

func ParseSyncOperation(s string) (SyncOperation, bool) {
	upper := strings.ToUpper(strings.TrimSpace(s))
	switch upper {
	case "TRACKING":
		return SyncOpTracking, true
	case "RATES":
		return SyncOpRates, true
	case "BOOKINGS", "BOOKING":
		return SyncOpBookings, true
	case "DOCUMENTS", "DOCUMENT":
		return SyncOpDocuments, true
	case "FULL_SYNC", "FULL", "ALL":
		return SyncOpFullSync, true
	default:
		return "", false
	}
}

// SyncStatus represents the execution state of an integration sync job.
type SyncStatus string

const (
	SyncStatusPending        SyncStatus = "PENDING"
	SyncStatusRunning        SyncStatus = "RUNNING"
	SyncStatusSuccess        SyncStatus = "SUCCESS"
	SyncStatusPartialSuccess SyncStatus = "PARTIAL_SUCCESS"
	SyncStatusFailed         SyncStatus = "FAILED"
	SyncStatusCancelled      SyncStatus = "CANCELLED"
)

func ParseSyncStatus(s string) SyncStatus {
	upper := strings.ToUpper(strings.TrimSpace(s))
	switch upper {
	case "RUNNING", "IN_PROGRESS":
		return SyncStatusRunning
	case "SUCCESS", "COMPLETED", "OK":
		return SyncStatusSuccess
	case "PARTIAL_SUCCESS", "PARTIAL":
		return SyncStatusPartialSuccess
	case "FAILED", "ERROR":
		return SyncStatusFailed
	case "CANCELLED", "ABORTED":
		return SyncStatusCancelled
	default:
		return SyncStatusPending
	}
}

// IntegrationHealthState represents the real-time operational health of a carrier integration.
type IntegrationHealthState string

const (
	HealthHealthy      IntegrationHealthState = "HEALTHY"
	HealthAttention    IntegrationHealthState = "ATTENTION"
	HealthError        IntegrationHealthState = "ERROR"
	HealthDisabled     IntegrationHealthState = "DISABLED"
	HealthDisconnected IntegrationHealthState = "DISCONNECTED"
)

// WebhookEventStatus represents the processing outcome of an inbound carrier webhook.
type WebhookEventStatus string

const (
	WebhookStatusPending   WebhookEventStatus = "PENDING"
	WebhookStatusProcessed WebhookEventStatus = "PROCESSED"
	WebhookStatusFailed    WebhookEventStatus = "FAILED"
	WebhookStatusDuplicate WebhookEventStatus = "DUPLICATE"
	WebhookStatusIgnored   WebhookEventStatus = "IGNORED"
)
