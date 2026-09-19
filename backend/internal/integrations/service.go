package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/jmoiron/sqlx"
)

// DeadLetterEntry represents a stored dead-letter event for tenant visibility.
type DeadLetterEntry struct {
	ID            int64     `json:"id" db:"id"`
	OrgID         int64     `json:"org_id" db:"org_id"`
	Provider      string    `json:"provider" db:"provider_name"`
	CorrelationID string    `json:"correlation_id" db:"correlation_id"`
	ErrorCode     string    `json:"error_code" db:"error_code"`
	ErrorMessage  string    `json:"error_message" db:"error_message"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// SaveConfigRequest represents a request to update an integration's non-secret or secret configuration.
type SaveConfigRequest struct {
	Type        IntegrationType `json:"type"`
	Provider    ProviderName    `json:"provider"`
	IsEnabled   bool            `json:"is_enabled"`
	EndpointURL string          `json:"endpoint_url,omitempty"`
	Region      string          `json:"region,omitempty"`
	FromAddress string          `json:"from_address,omitempty"`
	TimeoutSec  int             `json:"timeout_sec,omitempty"`
	MaxRetries  int             `json:"max_retries,omitempty"`
	SecretValue string          `json:"secret_value,omitempty"` // will be masked and protected
}

// GatewayService coordinates external integrations, enforces tenant isolation, and gates actions.
type GatewayService interface {
	GetIntegrationStatuses(ctx context.Context, orgID int64) ([]MaskedIntegrationConfig, error)
	GetIntegrationConfig(ctx context.Context, orgID int64, itype IntegrationType, provider ProviderName) (*MaskedIntegrationConfig, error)
	SaveIntegrationConfig(ctx context.Context, orgID int64, req SaveConfigRequest) error
	SendSMS(ctx context.Context, orgID int64, req SMSRequest) (*SMSResponse, error)
	SendEmail(ctx context.Context, orgID int64, req EmailRequest) (*EmailResponse, error)
	GetTracking(ctx context.Context, orgID int64, carrierSCAC, trackingNumber string) (*TrackingResponse, error)
	ProcessWebhook(ctx context.Context, provider string, orgID int64, headers http.Header, rawBody []byte) (*WebhookVerificationResult, error)
	GetDeadLetterEvents(ctx context.Context, orgID int64) ([]DeadLetterEntry, error)
	ListSMSMessages(ctx context.Context, orgID int64, limit int) ([]SMSMessageEntry, error)
	HandleTwilioWebhook(ctx context.Context, req *http.Request) error
	ListEmailMessages(ctx context.Context, orgID int64, limit int) ([]EmailMessageEntry, error)
	HandleSESWebhook(ctx context.Context, req *http.Request) error
	UploadDocument(ctx context.Context, orgID int64, key string, data []byte, contentType string) (*UploadResponse, error)
	DownloadDocument(ctx context.Context, orgID int64, key string) ([]byte, string, error)
	GetDocumentURL(ctx context.Context, orgID int64, key string, durationMinutes int) (string, error)
	DeleteDocument(ctx context.Context, orgID int64, key string) error
	ExtractDocumentText(ctx context.Context, orgID int64, req TextractRequest) (*TextractResponse, error)
}

type defaultGatewayService struct {
	db             *sqlx.DB
	configRepo     ConfigRepository
	notificationPv NotificationProvider
	trackingPv     TrackingIntegrationProvider
	storagePv      StorageProvider
	textractPv     TextractProvider
	webhookGateway WebhookGateway
	idempotencyMgr IdempotencyManager
	httpClient     *ResilientHTTPClient
	actionsSvc     actions.Service
	twilioWebhook  *TwilioWebhookHandler
	sesWebhook     *SESWebhookHandler
	flags          GlobalFeatureFlags
}

// NewGatewayService constructs the authoritative Integration Gateway.
func NewGatewayService(
	db *sqlx.DB,
	configRepo ConfigRepository,
	notificationPv NotificationProvider,
	trackingPv TrackingIntegrationProvider,
	storagePv StorageProvider,
	webhookGateway WebhookGateway,
	idempotencyMgr IdempotencyManager,
	httpClient *ResilientHTTPClient,
	actionsSvc actions.Service,
	textractOpt ...TextractProvider,
) GatewayService {
	if httpClient == nil {
		httpClient = NewResilientHTTPClient(DefaultHTTPClientConfig())
	}
	if notificationPv == nil {
		notificationPv = NewTwilioNotificationProvider(db, configRepo, httpClient)
	}
	if trackingPv == nil {
		trackingPv = NewUnconfiguredTrackingProvider(ProviderMaerskAPI)
	}
	if storagePv == nil {
		storagePv = NewS3StorageProvider(db, configRepo, httpClient)
	}
	var textractPv TextractProvider
	if len(textractOpt) > 0 && textractOpt[0] != nil {
		textractPv = textractOpt[0]
	} else {
		textractPv = NewAWSTextractProvider(db, configRepo, httpClient)
	}

	return &defaultGatewayService{
		db:             db,
		configRepo:     configRepo,
		notificationPv: notificationPv,
		trackingPv:     trackingPv,
		storagePv:      storagePv,
		textractPv:     textractPv,
		webhookGateway: webhookGateway,
		idempotencyMgr: idempotencyMgr,
		httpClient:     httpClient,
		actionsSvc:     actionsSvc,
		twilioWebhook:  NewTwilioWebhookHandler(db, configRepo, webhookGateway),
		sesWebhook:     NewSESWebhookHandler(db, configRepo, webhookGateway),
		flags:          LoadFeatureFlags(),
	}
}

func (s *defaultGatewayService) GetIntegrationStatuses(ctx context.Context, orgID int64) ([]MaskedIntegrationConfig, error) {
	if orgID <= 0 {
		return nil, NewInvalidRequestError("gateway", "valid org_id is required")
	}

	// 1. Fetch persistent configs for org
	tenantConfigs, err := s.configRepo.ListConfigsByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}

	cfgMap := make(map[string]IntegrationConfig)
	for _, c := range tenantConfigs {
		cfgMap[string(c.Type)+":"+string(c.Provider)] = c
	}

	// 2. Build canonical set of integrations
	canonicalList := []struct {
		itype    IntegrationType
		provider ProviderName
		enabled  bool
	}{
		{TypeSMS, ProviderTwilio, s.flags.SMSEnabled},
		{TypeEmail, ProviderAWSSES, s.flags.SESEnabled},
		{TypeCarrierTracking, ProviderMaerskAPI, s.flags.CarrierTrackingEnabled},
		{TypeCarrierTracking, ProviderMSCApi, s.flags.CarrierTrackingEnabled},
		{TypeStorage, ProviderAWSS3, s.flags.S3Enabled},
		{TypeTextract, ProviderAWSTextract, s.flags.TextractEnabled},
		{TypeWebhook, ProviderGeneric, s.flags.TrackingWebhooksEnabled},
	}

	statuses := make([]MaskedIntegrationConfig, 0, len(canonicalList))
	for _, item := range canonicalList {
		key := string(item.itype) + ":" + string(item.provider)
		if existing, ok := cfgMap[key]; ok {
			statuses = append(statuses, existing.ToMaskedConfig())
		} else {
			// Not yet configured in DB: return honest NOT_CONFIGURED state
			status := StatusNotConfigured
			if !item.enabled {
				status = StatusDisabled
			}
			statuses = append(statuses, MaskedIntegrationConfig{
				Type:          item.itype,
				Provider:      item.provider,
				IsEnabled:     item.enabled,
				Status:        status,
				HasSecret:     false,
				TimeoutSec:    15,
				MaxRetries:    3,
				HealthMessage: "Integration is not configured. Live external traffic is disabled.",
			})
		}
	}

	return statuses, nil
}

func (s *defaultGatewayService) GetIntegrationConfig(ctx context.Context, orgID int64, itype IntegrationType, provider ProviderName) (*MaskedIntegrationConfig, error) {
	if orgID <= 0 {
		return nil, NewInvalidRequestError("gateway", "valid org_id is required")
	}

	cfg, err := s.configRepo.GetConfig(ctx, orgID, itype, provider)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		masked := MaskedIntegrationConfig{
			Type:          itype,
			Provider:      provider,
			IsEnabled:     false,
			Status:        StatusNotConfigured,
			HasSecret:     false,
			TimeoutSec:    15,
			MaxRetries:    3,
			HealthMessage: "Integration is not configured.",
		}
		return &masked, nil
	}
	masked := cfg.ToMaskedConfig()
	return &masked, nil
}

func (s *defaultGatewayService) SaveIntegrationConfig(ctx context.Context, orgID int64, req SaveConfigRequest) error {
	if orgID <= 0 {
		return NewInvalidRequestError("gateway", "valid org_id is required")
	}
	if req.Type == "" || req.Provider == "" {
		return NewInvalidRequestError("gateway", "type and provider are required")
	}

	nonSec := NonSecretConfig{
		EndpointURL:    req.EndpointURL,
		Region:         req.Region,
		FromIdentifier: req.FromAddress,
		TimeoutSeconds: req.TimeoutSec,
		MaxRetries:     req.MaxRetries,
	}
	if nonSec.TimeoutSeconds <= 0 {
		nonSec.TimeoutSeconds = 15
	}
	if nonSec.MaxRetries <= 0 {
		nonSec.MaxRetries = 3
	}

	configBytes, _ := json.Marshal(nonSec)

	// Determine status
	status := StatusNotConfigured
	if req.IsEnabled {
		if req.SecretValue != "" || req.EndpointURL != "" {
			status = StatusHealthy
		} else {
			status = StatusConfigInvalid
		}
	} else {
		status = StatusDisabled
	}

	configRecord := &IntegrationConfig{
		OrgID:         orgID,
		Type:          req.Type,
		Provider:      req.Provider,
		IsEnabled:     req.IsEnabled,
		Status:        status,
		ConfigJSON:    string(configBytes),
		HealthMessage: fmt.Sprintf("Configuration updated at %s", time.Now().UTC().Format(time.RFC3339)),
	}

	if req.SecretValue != "" {
		// Store masked placeholder / encrypted token
		configRecord.EncryptedSecret = MaskSecret(req.SecretValue)
	}

	err := s.configRepo.SaveConfig(ctx, configRecord)
	if err != nil {
		return err
	}

	// Record audit event
	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorType:    "USER",
		ActorName:    "Operator",
		Action:       "UPDATE_INTEGRATION_CONFIG",
		Module:       "INTEGRATIONS",
		ResourceType: string(req.Type),
		ResourceID:   string(req.Provider),
		Description:  fmt.Sprintf("Updated %s integration config for provider %s (Enabled: %v, Status: %s)", req.Type, req.Provider, req.IsEnabled, status),
		Result:       domain.ResultSuccess,
	})

	return nil
}

func (s *defaultGatewayService) SendSMS(ctx context.Context, orgID int64, req SMSRequest) (*SMSResponse, error) {
	if orgID <= 0 {
		return nil, NewInvalidRequestError("sms", "valid org_id is required")
	}

	// 1. Validate destination phone number format (E.164)
	if _, err := ValidateE164Phone(req.GetRecipient()); err != nil {
		return nil, NewInvalidRequestError("sms", err.Error())
	}

	// 2. Validate message body hygiene, non-emptiness, and secret leak prevention
	if err := ValidateSMSBody(req.GetBody()); err != nil {
		return nil, NewInvalidRequestError("sms", err.Error())
	}

	flags := LoadFeatureFlags()
	isTenantEnabled := false
	if s.configRepo != nil {
		if tcfg, err := s.configRepo.GetConfig(ctx, orgID, TypeSMS, ProviderTwilio); err == nil && tcfg != nil && tcfg.IsEnabled {
			isTenantEnabled = true
		}
	}

	if !flags.SMSEnabled && !isTenantEnabled {
		return nil, NewProviderNotConfiguredError("Twilio", "SMS")
	}

	// Check idempotency if key provided
	if req.IdempotencyKey != "" && s.idempotencyMgr != nil {
		acquired, prevResult, err := s.idempotencyMgr.Acquire(ctx, orgID, req.IdempotencyKey, "send_sms", 24*time.Hour)
		if err != nil {
			return nil, err
		}
		if !acquired && len(prevResult) > 0 {
			var resp SMSResponse
			if err := json.Unmarshal(prevResult, &resp); err == nil {
				return &resp, nil
			}
		}
	}

	// Execute through notification provider
	resp, err := s.notificationPv.SendSMS(ctx, orgID, req)
	if err != nil {
		return nil, err
	}

	if req.IdempotencyKey != "" && s.idempotencyMgr != nil && resp != nil {
		respBytes, _ := json.Marshal(resp)
		_ = s.idempotencyMgr.Release(ctx, orgID, req.IdempotencyKey, respBytes)
	}

	return resp, nil
}

func (s *defaultGatewayService) SendEmail(ctx context.Context, orgID int64, req EmailRequest) (*EmailResponse, error) {
	if orgID <= 0 {
		return nil, NewInvalidRequestError("email", "valid org_id is required")
	}

	// 1. Validate recipient email format
	toEmail := req.GetRecipient()
	if _, err := ValidateEmailAddress(toEmail); err != nil {
		return nil, NewInvalidRequestError("email", err.Error())
	}

	// 2. Validate email content and secret leak prevention
	if err := ValidateEmailContent(req.Subject, req.GetBody()); err != nil {
		return nil, NewInvalidRequestError("email", err.Error())
	}

	// 3. Suppression check (defense-in-depth)
	if s.db != nil {
		var reason string
		query := `SELECT reason FROM email_suppressions WHERE org_id = ? AND email = ? LIMIT 1`
		if err := s.db.GetContext(ctx, &reason, query, orgID, toEmail); err == nil && reason != "" {
			return nil, NewInvalidRequestError("AWS_SES", fmt.Sprintf("recipient %q is suppressed for this organization (reason: %s); delivery cancelled", toEmail, reason))
		}
	}

	flags := LoadFeatureFlags()
	isTenantEnabled := false
	if s.configRepo != nil {
		if tcfg, err := s.configRepo.GetConfig(ctx, orgID, TypeEmail, ProviderAWSSES); err == nil && tcfg != nil && tcfg.IsEnabled {
			isTenantEnabled = true
		}
	}

	if !flags.SESEnabled && !isTenantEnabled {
		return nil, NewProviderNotConfiguredError("AWS_SES", "Email")
	}

	if req.IdempotencyKey != "" && s.idempotencyMgr != nil {
		acquired, prevResult, err := s.idempotencyMgr.Acquire(ctx, orgID, req.IdempotencyKey, "send_email", 24*time.Hour)
		if err != nil {
			return nil, err
		}
		if !acquired && len(prevResult) > 0 {
			var resp EmailResponse
			if err := json.Unmarshal(prevResult, &resp); err == nil {
				return &resp, nil
			}
		}
	}

	resp, err := s.notificationPv.SendEmail(ctx, orgID, req)
	if err != nil {
		return nil, err
	}

	if req.IdempotencyKey != "" && s.idempotencyMgr != nil && resp != nil {
		respBytes, _ := json.Marshal(resp)
		_ = s.idempotencyMgr.Release(ctx, orgID, req.IdempotencyKey, respBytes)
	}

	return resp, nil
}

func (s *defaultGatewayService) GetTracking(ctx context.Context, orgID int64, carrierSCAC, trackingNumber string) (*TrackingResponse, error) {
	if orgID <= 0 {
		return nil, NewInvalidRequestError("carrier_tracking", "valid org_id is required")
	}

	cleanSCAC := strings.ToUpper(strings.TrimSpace(carrierSCAC))
	cleanTracking := strings.TrimSpace(trackingNumber)
	if cleanSCAC == "" {
		return nil, NewInvalidRequestError("carrier_tracking", "carrier_scac is required")
	}
	if cleanTracking == "" {
		return nil, NewInvalidRequestError("carrier_tracking", "tracking_number is required")
	}

	if s.trackingPv == nil || s.trackingPv.ProviderName() == ProviderUnconfigured {
		return nil, NewProviderNotConfiguredError(cleanSCAC, "Carrier Tracking")
	}

	return s.trackingPv.GetTracking(ctx, orgID, cleanSCAC, cleanTracking)
}

func (s *defaultGatewayService) ProcessWebhook(ctx context.Context, provider string, orgID int64, headers http.Header, rawBody []byte) (*WebhookVerificationResult, error) {
	if !s.flags.TrackingWebhooksEnabled {
		return nil, NewProviderNotConfiguredError(provider, "Tracking Webhooks")
	}

	return s.webhookGateway.VerifyAndRecord(ctx, ProviderName(strings.ToUpper(provider)), orgID, headers, rawBody)
}

func (s *defaultGatewayService) GetDeadLetterEvents(ctx context.Context, orgID int64) ([]DeadLetterEntry, error) {
	if orgID <= 0 {
		return nil, NewInvalidRequestError("dead_letter", "valid org_id is required")
	}

	var entries []DeadLetterEntry
	query := `
		SELECT id, org_id, provider_name, correlation_id, error_code, error_message, created_at
		FROM external_webhook_dead_letter
		WHERE org_id = ?
		ORDER BY id DESC
		LIMIT 50
	`
	err := s.db.SelectContext(ctx, &entries, query, orgID)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *defaultGatewayService) ListSMSMessages(ctx context.Context, orgID int64, limit int) ([]SMSMessageEntry, error) {
	if orgID <= 0 {
		return nil, NewInvalidRequestError("sms", "valid org_id is required")
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	entries := make([]SMSMessageEntry, 0)
	query := `
		SELECT id, org_id, provider, message_sid, to_phone, from_phone, body_preview,
		       status, error_code, error_message, idempotency_key, correlation_id,
		       created_at, updated_at, delivered_at
		FROM sms_messages
		WHERE org_id = ?
		ORDER BY id DESC
		LIMIT ?
	`
	err := s.db.SelectContext(ctx, &entries, query, orgID, limit)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *defaultGatewayService) HandleTwilioWebhook(ctx context.Context, req *http.Request) error {
	if s.twilioWebhook == nil {
		s.twilioWebhook = NewTwilioWebhookHandler(s.db, s.configRepo, s.webhookGateway)
	}
	return s.twilioWebhook.HandleStatusCallback(ctx, req)
}

func (s *defaultGatewayService) ListEmailMessages(ctx context.Context, orgID int64, limit int) ([]EmailMessageEntry, error) {
	if orgID <= 0 {
		return nil, NewInvalidRequestError("email", "valid org_id is required")
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	entries := make([]EmailMessageEntry, 0)
	query := `
		SELECT id, org_id, provider, message_id, to_email, from_email, subject, body_preview,
		       status, bounce_type, bounce_sub_type, complaint_feedback_type, error_code, error_message,
		       idempotency_key, correlation_id, created_at, updated_at, delivered_at, bounced_at, complained_at
		FROM email_messages
		WHERE org_id = ?
		ORDER BY id DESC
		LIMIT ?
	`
	err := s.db.SelectContext(ctx, &entries, query, orgID, limit)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *defaultGatewayService) HandleSESWebhook(ctx context.Context, req *http.Request) error {
	if s.sesWebhook == nil {
		s.sesWebhook = NewSESWebhookHandler(s.db, s.configRepo, s.webhookGateway)
	}
	return s.sesWebhook.HandleSESEvent(ctx, req)
}

func (s *defaultGatewayService) UploadDocument(ctx context.Context, orgID int64, key string, data []byte, contentType string) (*UploadResponse, error) {
	if orgID <= 0 {
		return nil, NewInvalidRequestError("storage", "valid org_id is required")
	}
	if s.storagePv == nil || s.storagePv.ProviderName() == ProviderUnconfigured {
		return nil, NewProviderNotConfiguredError("AWS_S3", "Storage")
	}
	return s.storagePv.Upload(ctx, orgID, key, data, contentType)
}

func (s *defaultGatewayService) DownloadDocument(ctx context.Context, orgID int64, key string) ([]byte, string, error) {
	if orgID <= 0 {
		return nil, "", NewInvalidRequestError("storage", "valid org_id is required")
	}
	if s.storagePv == nil || s.storagePv.ProviderName() == ProviderUnconfigured {
		return nil, "", NewProviderNotConfiguredError("AWS_S3", "Storage")
	}
	return s.storagePv.Download(ctx, orgID, key)
}

func (s *defaultGatewayService) GetDocumentURL(ctx context.Context, orgID int64, key string, durationMinutes int) (string, error) {
	if orgID <= 0 {
		return "", NewInvalidRequestError("storage", "valid org_id is required")
	}
	if s3Pv, ok := s.storagePv.(*S3StorageProvider); ok {
		return s3Pv.GetPresignedURL(ctx, orgID, key, durationMinutes)
	}
	return "", NewProviderNotConfiguredError("AWS_S3", "Storage")
}

func (s *defaultGatewayService) DeleteDocument(ctx context.Context, orgID int64, key string) error {
	if orgID <= 0 {
		return NewInvalidRequestError("storage", "valid org_id is required")
	}
	if s.storagePv == nil || s.storagePv.ProviderName() == ProviderUnconfigured {
		return NewProviderNotConfiguredError("AWS_S3", "Storage")
	}
	return s.storagePv.Delete(ctx, orgID, key)
}

func (s *defaultGatewayService) ExtractDocumentText(ctx context.Context, orgID int64, req TextractRequest) (*TextractResponse, error) {
	if orgID <= 0 {
		return nil, NewInvalidRequestError("textract", "valid org_id is required")
	}
	if s.textractPv == nil || s.textractPv.ProviderName() == ProviderUnconfigured {
		return nil, NewProviderNotConfiguredError("AWS_TEXTRACT", "OCR")
	}
	return s.textractPv.ExtractDocumentText(ctx, orgID, req)
}
