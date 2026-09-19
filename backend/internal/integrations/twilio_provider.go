package integrations

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var e164Regex = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

// ValidateE164Phone checks if a phone number strictly matches standard E.164 format.
func ValidateE164Phone(phone string) (string, error) {
	cleaned := strings.TrimSpace(phone)
	if cleaned == "" {
		return "", NewInvalidRequestError("TWILIO", "destination phone number is required")
	}
	if !e164Regex.MatchString(cleaned) {
		return "", NewInvalidRequestError("TWILIO", fmt.Sprintf("invalid destination phone number %q: must be in E.164 format (e.g. +1234567890)", cleaned))
	}
	return cleaned, nil
}

// ValidateSMSBody validates outgoing text message constraints and prevents credential leakage.
func ValidateSMSBody(body string) error {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return NewInvalidRequestError("TWILIO", "SMS body text cannot be empty")
	}
	if len(trimmed) > 1600 {
		return NewInvalidRequestError("TWILIO", fmt.Sprintf("SMS body exceeds maximum allowable length of 1600 characters (got %d)", len(trimmed)))
	}

	// Safety check: ensure no internal credentials or authorization tokens leak into SMS
	lower := strings.ToLower(trimmed)
	sensitivePatterns := []string{"bearer ", "auth_token", "secret_key", "password=", "jwt=", "aws_secret"}
	for _, pat := range sensitivePatterns {
		if strings.Contains(lower, pat) {
			return NewInvalidRequestError("TWILIO", "outgoing SMS message contains prohibited sensitive credential patterns")
		}
	}

	return nil
}

// TwilioResolvedConfig holds resolved credentials and endpoint configuration for a tenant.
type TwilioResolvedConfig struct {
	AccountSID        string
	AuthToken         string
	FromNumber        string
	BaseURL           string
	StatusCallbackURL string
	IsEnabled         bool
}

// TwilioErrorPayload represents standard error responses returned by Twilio.
type TwilioErrorPayload struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"more_info"`
	Status   int    `json:"status"`
}

// TwilioMessageResponse represents the JSON payload returned on successful SMS dispatch.
type TwilioMessageResponse struct {
	SID         string `json:"sid"`
	AccountSID  string `json:"account_sid"`
	To          string `json:"to"`
	From        string `json:"from"`
	Body        string `json:"body"`
	Status      string `json:"status"`
	ErrorCode   *int   `json:"error_code"`
	ErrorMessage *string `json:"error_message"`
	DateCreated string `json:"date_created"`
	DateSent    *string `json:"date_sent"`
}

// TwilioNotificationProvider implements NotificationProvider for Twilio SMS.
type TwilioNotificationProvider struct {
	db         *sqlx.DB
	configRepo ConfigRepository
	httpClient *ResilientHTTPClient
}

// NewTwilioNotificationProvider constructs an active Twilio provider.
func NewTwilioNotificationProvider(db *sqlx.DB, configRepo ConfigRepository, httpClient *ResilientHTTPClient) NotificationProvider {
	if httpClient == nil {
		httpClient = NewResilientHTTPClient(DefaultHTTPClientConfig())
	}
	return &TwilioNotificationProvider{
		db:         db,
		configRepo: configRepo,
		httpClient: httpClient,
	}
}

func (p *TwilioNotificationProvider) ProviderName() ProviderName {
	return ProviderTwilio
}

// ResolveConfig resolves Twilio credentials with tenant-isolation:
// 1. Checks tenant-specific external_integration_configs in DB.
// 2. Falls back to environment variables (TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, TWILIO_FROM_NUMBER, TWILIO_ENABLED).
func (p *TwilioNotificationProvider) ResolveConfig(ctx context.Context, orgID int64) (*TwilioResolvedConfig, error) {
	cfg := &TwilioResolvedConfig{
		BaseURL: "https://api.twilio.com",
	}

	// 1. Try tenant database configuration if repository exists
	if p.configRepo != nil && orgID > 0 {
		tenantCfg, err := p.configRepo.GetConfig(ctx, orgID, TypeSMS, ProviderTwilio)
		if err == nil && tenantCfg != nil {
			var nonSec struct {
				AccountSID        string `json:"account_sid"`
				FromNumber        string `json:"from_number"`
				EndpointURL       string `json:"endpoint_url"`
				StatusCallbackURL string `json:"status_callback_url"`
			}
			var sec struct {
				AuthToken string `json:"auth_token"`
			}
			_ = json.Unmarshal([]byte(tenantCfg.ConfigJSON), &nonSec)
			_ = json.Unmarshal([]byte(tenantCfg.EncryptedSecret), &sec)

			if nonSec.AccountSID != "" {
				cfg.AccountSID = nonSec.AccountSID
			}
			if sec.AuthToken != "" && sec.AuthToken != "••••••••" {
				cfg.AuthToken = sec.AuthToken
			}
			if nonSec.FromNumber != "" {
				cfg.FromNumber = nonSec.FromNumber
			}
			if nonSec.EndpointURL != "" {
				cfg.BaseURL = strings.TrimRight(nonSec.EndpointURL, "/")
			}
			if nonSec.StatusCallbackURL != "" {
				cfg.StatusCallbackURL = nonSec.StatusCallbackURL
			}
			cfg.IsEnabled = tenantCfg.IsEnabled
		}
	}

	// 2. Fall back to environment variables if not configured in tenant DB
	if cfg.AccountSID == "" {
		cfg.AccountSID = os.Getenv(EnvTwilioAccountSID)
	}
	if cfg.AuthToken == "" {
		cfg.AuthToken = os.Getenv(EnvTwilioAuthToken)
	}
	if cfg.FromNumber == "" {
		cfg.FromNumber = os.Getenv(EnvTwilioFromNumber)
	}
	if envURL := os.Getenv("TWILIO_BASE_URL"); envURL != "" {
		cfg.BaseURL = strings.TrimRight(envURL, "/")
	}
	if envCallback := os.Getenv("TWILIO_STATUS_CALLBACK"); envCallback != "" {
		cfg.StatusCallbackURL = envCallback
	}
	if !cfg.IsEnabled {
		cfg.IsEnabled = os.Getenv(EnvSMSEnabled) == "true" || os.Getenv("TWILIO_ENABLED") == "true"
	}

	return cfg, nil
}

// SendSMS dispatches a validated, tenant-isolated SMS message through Twilio.
func (p *TwilioNotificationProvider) SendSMS(ctx context.Context, orgID int64, req SMSRequest) (*SMSResponse, error) {
	if orgID <= 0 {
		return nil, NewInvalidRequestError("TWILIO", "valid org_id is required")
	}

	// 1. Validation: Phone number & Message
	toPhone, err := ValidateE164Phone(req.GetRecipient())
	if err != nil {
		return nil, err
	}

	if err := ValidateSMSBody(req.GetBody()); err != nil {
		return nil, err
	}

	// 2. Resolve credentials safely
	cfg, err := p.ResolveConfig(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if !cfg.IsEnabled {
		return nil, NewProviderDisabledError("TWILIO")
	}

	if cfg.AccountSID == "" || cfg.AuthToken == "" || cfg.FromNumber == "" {
		return nil, NewProviderNotConfiguredError("TWILIO", "SMS (missing Account SID, Auth Token, or From Number)")
	}

	correlationID := req.CorrelationID
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	idempotencyKey := req.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = fmt.Sprintf("sms-%d-%s-%d", orgID, toPhone, time.Now().UnixNano())
	}

	// 3. Prepare Twilio API request
	endpoint := fmt.Sprintf("%s/2010-04-01/Accounts/%s/Messages.json", cfg.BaseURL, cfg.AccountSID)

	formData := url.Values{}
	formData.Set("To", toPhone)
	fromNumber := cfg.FromNumber
	if req.SenderID != "" {
		fromNumber = req.SenderID
	}
	formData.Set("From", fromNumber)
	formData.Set("Body", req.GetBody())
	if cfg.StatusCallbackURL != "" {
		formData.Set("StatusCallback", cfg.StatusCallbackURL)
	}
	// Basic Auth credentials: AccountSID : AuthToken
	authString := fmt.Sprintf("%s:%s", cfg.AccountSID, cfg.AuthToken)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(authString))

	opts := RequestOptions{
		Provider:       "TWILIO",
		CorrelationID:  correlationID,
		IdempotencyKey: idempotencyKey,
		AuthHeader:     "Basic " + encodedAuth,
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
	}

	// 4. Execute through ResilientHTTPClient (handles timeouts, retries, 429 backoff)
	httpResp, respBody, err := p.httpClient.Execute(ctx, http.MethodPost, endpoint, []byte(formData.Encode()), opts)
	if err != nil {
		var intErr *IntegrationError
		if errors.As(err, &intErr) {
			if len(respBody) > 0 {
				return nil, p.normalizeTwilioError(intErr.HTTPStatus(), respBody)
			}
			return nil, intErr
		}
		return nil, p.normalizeHTTPError(err)
	}

	// 5. Parse response & Normalize
	if httpResp != nil && (httpResp.StatusCode < 200 || httpResp.StatusCode >= 300) {
		return nil, p.normalizeTwilioError(httpResp.StatusCode, respBody)
	}

	var twilioResp TwilioMessageResponse
	if err := json.Unmarshal(respBody, &twilioResp); err != nil {
		return nil, NewProviderError("TWILIO", "malformed response from Twilio", http.StatusBadGateway)
	}

	// 6. Normalize status: Twilio returns 'queued', 'accepted', 'sent'
	normalizedStatus := strings.ToUpper(twilioResp.Status)
	if normalizedStatus == "" {
		normalizedStatus = "QUEUED"
	}

	// 7. Persist dispatch record in sms_messages
	p.recordDispatch(ctx, orgID, twilioResp.SID, toPhone, fromNumber, req.Body, normalizedStatus, idempotencyKey, correlationID)

	return &SMSResponse{
		MessageID:     twilioResp.SID,
		Provider:      "TWILIO",
		Status:        normalizedStatus,
		SentAt:        time.Now().UTC(),
		CorrelationID: correlationID,
	}, nil
}

func (p *TwilioNotificationProvider) SendEmail(ctx context.Context, orgID int64, req EmailRequest) (*EmailResponse, error) {
	return nil, NewProviderNotConfiguredError("TWILIO", "Email (Twilio provider supports SMS only)")
}

// Status inspects the operational state of Twilio for the given tenant.
func (p *TwilioNotificationProvider) Status(ctx context.Context, orgID int64) (*ProviderHealth, error) {
	cfg, err := p.ResolveConfig(ctx, orgID)
	if err != nil {
		return nil, err
	}

	health := &ProviderHealth{
		Provider:       ProviderTwilio,
		Type:           TypeSMS,
		LastCheckedAt:  time.Now().UTC(),
		ProductionSafe: true,
	}

	if !cfg.IsEnabled {
		health.Status = StatusDisabled
		health.Message = "Twilio SMS integration is disabled by feature flag."
		return health, nil
	}

	if cfg.AccountSID == "" || cfg.AuthToken == "" || cfg.FromNumber == "" {
		health.Status = StatusNotConfigured
		health.Message = "Twilio credentials (Account SID, Auth Token, or From Number) are not configured."
		return health, nil
	}

	health.Status = StatusHealthy
	health.Message = fmt.Sprintf("Twilio SMS is configured with sender %s.", cfg.FromNumber)
	return health, nil
}

// normalizeHTTPError maps low-level network/HTTP errors into normalized integration errors.
func (p *TwilioNotificationProvider) normalizeHTTPError(err error) error {
	msg := err.Error()
	if strings.Contains(msg, "context deadline exceeded") || strings.Contains(msg, "timeout") {
		return NewTimeoutError("TWILIO", "request to Twilio API timed out")
	}
	if strings.Contains(msg, "context canceled") {
		return NewTimeoutError("TWILIO", "request context canceled")
	}
	return NewConnectionFailedError("TWILIO", fmt.Sprintf("network connection to Twilio failed: %v", err))
}

// normalizeTwilioError maps Twilio error codes into normalized domain errors.
func (p *TwilioNotificationProvider) normalizeTwilioError(statusCode int, body []byte) error {
	var twilioErr TwilioErrorPayload
	_ = json.Unmarshal(body, &twilioErr)

	errMsg := twilioErr.Message
	if errMsg == "" {
		errMsg = fmt.Sprintf("Twilio request failed with HTTP %d", statusCode)
	}

	switch twilioErr.Code {
	case 20003:
		return NewAuthenticationFailedError("TWILIO", "Twilio authentication failed: invalid Account SID or Auth Token")
	case 21211, 21614:
		return NewInvalidRequestError("TWILIO", fmt.Sprintf("invalid recipient phone number: %s", errMsg))
	case 21408:
		return NewAuthorizationFailedError("TWILIO", "permission to send SMS to region has not been enabled in Twilio account")
	case 20429:
		return NewRateLimitedError("TWILIO", 60)
	case 30001, 30003:
		return NewConnectionFailedError("TWILIO", fmt.Sprintf("carrier connection failure: %s", errMsg))
	default:
		if statusCode == http.StatusTooManyRequests {
			return NewRateLimitedError("TWILIO", 60)
		}
		if statusCode == http.StatusUnauthorized {
			return NewAuthenticationFailedError("TWILIO", "unauthorized access to Twilio API")
		}
		if statusCode == http.StatusForbidden {
			return NewAuthorizationFailedError("TWILIO", errMsg)
		}
		if statusCode >= 500 {
			return NewProviderUnavailableError("TWILIO", errMsg)
		}
		return NewProviderError("TWILIO", errMsg, statusCode)
	}
}

// recordDispatch asynchronously or synchronously records outgoing SMS dispatch into sms_messages.
func (p *TwilioNotificationProvider) recordDispatch(ctx context.Context, orgID int64, sid, toPhone, fromPhone, body, status, idempotencyKey, correlationID string) {
	if p.db == nil {
		return
	}

	bodyPreview := body
	if len(bodyPreview) > 250 {
		bodyPreview = bodyPreview[:247] + "..."
	}

	query := `
		INSERT INTO sms_messages (
			org_id, provider, message_sid, to_phone, from_phone, body_preview,
			status, idempotency_key, correlation_id, created_at, updated_at
		) VALUES (?, 'TWILIO', ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			status = VALUES(status),
			message_sid = COALESCE(VALUES(message_sid), message_sid),
			updated_at = NOW()
	`
	_, _ = p.db.ExecContext(ctx, query, orgID, sid, toPhone, fromPhone, bodyPreview, status, idempotencyKey, correlationID)
}
