package integrations

import (
	"time"
)

// IntegrationType defines the category of external integration.
type IntegrationType string

const (
	TypeSMS             IntegrationType = "SMS"
	TypeEmail           IntegrationType = "EMAIL"
	TypeCarrierTracking IntegrationType = "CARRIER_TRACKING"
	TypeStorage         IntegrationType = "STORAGE"
	TypeTextract        IntegrationType = "TEXTRACT"
	TypeWebhook         IntegrationType = "WEBHOOK"
)

// ProviderName defines known external vendor names.
type ProviderName string

const (
	ProviderTwilio       ProviderName = "TWILIO"
	ProviderAWSSES       ProviderName = "AWS_SES"
	ProviderSMTP         ProviderName = "SMTP"
	ProviderMaerskAPI    ProviderName = "MAERSK_API"
	ProviderMSCApi       ProviderName = "MSC_API"
	ProviderAWSS3        ProviderName = "AWS_S3"
	ProviderAWSTextract  ProviderName = "AWS_TEXTRACT"
	ProviderGeneric      ProviderName = "GENERIC"
	ProviderUnconfigured ProviderName = "NONE"
)

// IntegrationStatus represents the authoritative operational state of an integration.
type IntegrationStatus string

const (
	StatusEnabled          IntegrationStatus = "ENABLED"
	StatusDisabled         IntegrationStatus = "DISABLED"
	StatusNotConfigured    IntegrationStatus = "NOT_CONFIGURED"
	StatusConfigInvalid    IntegrationStatus = "CONFIG_INVALID"
	StatusHealthy          IntegrationStatus = "HEALTHY"
	StatusDegraded         IntegrationStatus = "DEGRADED"
	StatusUnavailable      IntegrationStatus = "UNAVAILABLE"
)

// ProviderHealth captures runtime health metrics and status.
type ProviderHealth struct {
	Provider       ProviderName      `json:"provider"`
	Type           IntegrationType   `json:"type"`
	Status         IntegrationStatus `json:"status"`
	Message        string            `json:"message"`
	LatencyMs      int64             `json:"latency_ms,omitempty"`
	LastCheckedAt  time.Time         `json:"last_checked_at"`
	ProductionSafe bool              `json:"production_safe"`
}

// IntegrationConfig holds tenant-level or system-level configuration.
type IntegrationConfig struct {
	ID              int64             `json:"id" db:"id"`
	OrgID           int64             `json:"org_id" db:"org_id"`
	Type            IntegrationType   `json:"type" db:"integration_type"`
	Provider        ProviderName      `json:"provider" db:"provider_name"`
	IsEnabled       bool              `json:"is_enabled" db:"is_enabled"`
	Status          IntegrationStatus `json:"status" db:"status"`
	ConfigJSON      string            `json:"config_json,omitempty" db:"config_json"`
	EncryptedSecret string            `json:"-" db:"encrypted_secrets"`
	LastHealthCheck *time.Time        `json:"last_health_check,omitempty" db:"last_health_check"`
	HealthMessage   string            `json:"health_message,omitempty" db:"health_message"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
}

// MaskedIntegrationConfig is the safe DTO returned via HTTP APIs and rendered in UI.
// Secrets are strictly excluded or masked as "••••••••".
type MaskedIntegrationConfig struct {
	Type          IntegrationType   `json:"type"`
	Provider      ProviderName      `json:"provider"`
	IsEnabled     bool              `json:"is_enabled"`
	Status        IntegrationStatus `json:"status"`
	HasSecret     bool              `json:"has_secret"`
	MaskedSecret  string            `json:"masked_secret,omitempty"`
	EndpointURL   string            `json:"endpoint_url,omitempty"`
	Region        string            `json:"region,omitempty"`
	FromAddress   string            `json:"from_address,omitempty"`
	TimeoutSec    int               `json:"timeout_sec"`
	MaxRetries    int               `json:"max_retries"`
	HealthMessage string            `json:"health_message,omitempty"`
	LastCheckedAt *time.Time        `json:"last_checked_at,omitempty"`
}

// SMSRequest defines payload for sending an outbound SMS.
type SMSRequest struct {
	RecipientPhoneNumber string            `json:"recipient_phone_number"`
	ToPhone              string            `json:"to_phone,omitempty"`
	Body                 string            `json:"body"`
	Message              string            `json:"message,omitempty"`
	SenderID             string            `json:"sender_id,omitempty"`
	CorrelationID        string            `json:"correlation_id"`
	IdempotencyKey       string            `json:"idempotency_key"`
	Metadata             map[string]string `json:"metadata,omitempty"`
}

// GetRecipient returns the normalized recipient phone number.
func (r *SMSRequest) GetRecipient() string {
	if r.RecipientPhoneNumber != "" {
		return r.RecipientPhoneNumber
	}
	return r.ToPhone
}

// GetBody returns the message body text.
func (r *SMSRequest) GetBody() string {
	if r.Body != "" {
		return r.Body
	}
	return r.Message
}

// SMSResponse defines result for SMS dispatch.
type SMSResponse struct {
	MessageID     string    `json:"message_id"`
	Provider      string    `json:"provider"`
	Status        string    `json:"status"`
	SentAt        time.Time `json:"sent_at"`
	CorrelationID string    `json:"correlation_id"`
}

// EmailRequest defines payload for sending an outbound email.
type EmailRequest struct {
	ToRecipient    string            `json:"to_recipient"`
	ToEmail        string            `json:"to_email,omitempty"`
	Subject        string            `json:"subject"`
	BodyText       string            `json:"body_text"`
	BodyHTML       string            `json:"body_html,omitempty"`
	FromEmail      string            `json:"from_email,omitempty"`
	CorrelationID  string            `json:"correlation_id"`
	IdempotencyKey string            `json:"idempotency_key"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// GetRecipient returns the target recipient email address.
func (r *EmailRequest) GetRecipient() string {
	if r.ToRecipient != "" {
		return r.ToRecipient
	}
	return r.ToEmail
}

// GetBody returns the primary body text or HTML.
func (r *EmailRequest) GetBody() string {
	if r.BodyText != "" {
		return r.BodyText
	}
	return r.BodyHTML
}

// EmailResponse defines result for Email dispatch.
type EmailResponse struct {
	MessageID     string    `json:"message_id"`
	Provider      string    `json:"provider"`
	Status        string    `json:"status"`
	SentAt        time.Time `json:"sent_at"`
	CorrelationID string    `json:"correlation_id"`
}

// TrackingResponse holds normalized telemetry from carrier tracking.
type TrackingResponse struct {
	CarrierSCAC    string            `json:"carrier_scac"`
	TrackingNumber string            `json:"tracking_number"`
	Status         string            `json:"status"`
	CurrentPort    string            `json:"current_port,omitempty"`
	DestinationPort string           `json:"destination_port,omitempty"`
	ETD            *time.Time        `json:"etd,omitempty"`
	ETA            *time.Time        `json:"eta,omitempty"`
	CorrelationID  string            `json:"correlation_id"`
	LastUpdated    time.Time         `json:"last_updated"`
}

// UploadResponse holds metadata after uploading to document storage.
type UploadResponse struct {
	Key           string    `json:"key"`
	Location      string    `json:"location"`
	ETag          string    `json:"etag"`
	Size          int64     `json:"size"`
	UploadedAt    time.Time `json:"uploaded_at"`
	CorrelationID string    `json:"correlation_id"`
}

// WebhookVerificationResult encapsulates the outcome of verifying an incoming webhook.
type WebhookVerificationResult struct {
	Valid          bool              `json:"valid"`
	Provider       ProviderName      `json:"provider"`
	OrgID          int64             `json:"org_id"`
	EventID        string            `json:"event_id"`
	EventType      string            `json:"event_type"`
	Timestamp      time.Time         `json:"timestamp"`
	CorrelationID  string            `json:"correlation_id"`
	RejectionReason string           `json:"rejection_reason,omitempty"`
	IsReplay       bool              `json:"is_replay"`
}

// SMSMessageEntry represents a persisted outbound SMS record and delivery state.
type SMSMessageEntry struct {
	ID             int64      `json:"id" db:"id"`
	OrgID          int64      `json:"org_id" db:"org_id"`
	Provider       string     `json:"provider" db:"provider"`
	MessageSID     *string    `json:"message_sid,omitempty" db:"message_sid"`
	ToPhone        string     `json:"to_phone" db:"to_phone"`
	FromPhone      string     `json:"from_phone" db:"from_phone"`
	BodyPreview    string     `json:"body_preview" db:"body_preview"`
	Status         string     `json:"status" db:"status"`
	ErrorCode      *string    `json:"error_code,omitempty" db:"error_code"`
	ErrorMessage   *string    `json:"error_message,omitempty" db:"error_message"`
	IdempotencyKey string     `json:"idempotency_key" db:"idempotency_key"`
	CorrelationID  string     `json:"correlation_id" db:"correlation_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`
}

// EmailMessageEntry represents a persisted outbound email record and delivery state.
type EmailMessageEntry struct {
	ID                    int64      `json:"id" db:"id"`
	OrgID                 int64      `json:"org_id" db:"org_id"`
	Provider              string     `json:"provider" db:"provider"`
	MessageID             *string    `json:"message_id,omitempty" db:"message_id"`
	ToEmail               string     `json:"to_email" db:"to_email"`
	FromEmail             string     `json:"from_email" db:"from_email"`
	Subject               string     `json:"subject" db:"subject"`
	BodyPreview           string     `json:"body_preview" db:"body_preview"`
	Status                string     `json:"status" db:"status"`
	BounceType            *string    `json:"bounce_type,omitempty" db:"bounce_type"`
	BounceSubType         *string    `json:"bounce_sub_type,omitempty" db:"bounce_sub_type"`
	ComplaintFeedbackType *string    `json:"complaint_feedback_type,omitempty" db:"complaint_feedback_type"`
	ErrorCode             *string    `json:"error_code,omitempty" db:"error_code"`
	ErrorMessage          *string    `json:"error_message,omitempty" db:"error_message"`
	IdempotencyKey        string     `json:"idempotency_key" db:"idempotency_key"`
	CorrelationID         string     `json:"correlation_id" db:"correlation_id"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
	DeliveredAt           *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`
	BouncedAt             *time.Time `json:"bounced_at,omitempty" db:"bounced_at"`
	ComplainedAt          *time.Time `json:"complained_at,omitempty" db:"complained_at"`
}
