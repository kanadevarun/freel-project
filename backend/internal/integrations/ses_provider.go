package integrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sestypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
)

const (
	EnvAWSRegion = "AWS_REGION"
)

// Email regex pattern conforming to RFC 5322 standard.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+$`)

// ValidateEmailAddress validates recipient email format strictly.
func ValidateEmailAddress(address string) (string, error) {
	trimmed := strings.TrimSpace(address)
	if trimmed == "" {
		return "", NewInvalidRequestError("AWS_SES", "recipient email address cannot be empty")
	}

	// First parse with net/mail standard library
	parsed, err := mail.ParseAddress(trimmed)
	if err != nil {
		return "", NewInvalidRequestError("AWS_SES", fmt.Sprintf("invalid recipient email format: %q", trimmed))
	}

	// Strictly verify with regex to ensure complete domain with TLD
	if !emailRegex.MatchString(parsed.Address) {
		return "", NewInvalidRequestError("AWS_SES", fmt.Sprintf("invalid recipient email syntax: %q", trimmed))
	}

	return strings.ToLower(parsed.Address), nil
}

// ValidateEmailContent checks non-emptiness, maximum length, and credential hygiene.
func ValidateEmailContent(subject, body string) error {
	trimmedSub := strings.TrimSpace(subject)
	if trimmedSub == "" {
		return NewInvalidRequestError("AWS_SES", "email subject cannot be empty")
	}
	if len(trimmedSub) > 998 {
		return NewInvalidRequestError("AWS_SES", "email subject exceeds maximum allowed length of 998 characters")
	}

	trimmedBody := strings.TrimSpace(body)
	if trimmedBody == "" {
		return NewInvalidRequestError("AWS_SES", "email body content cannot be empty")
	}
	if len(trimmedBody) > 1000000 { // 1MB text limit
		return NewInvalidRequestError("AWS_SES", "email body exceeds maximum allowed size of 1MB")
	}

	// Pre-dispatch secret leak prevention
	bodyLower := strings.ToLower(trimmedBody)
	subLower := strings.ToLower(trimmedSub)
	prohibitedPatterns := []string{
		"bearer ",
		"auth_token",
		"secret_key",
		"aws_secret_access_key",
		"password=",
		"private_key",
	}

	for _, p := range prohibitedPatterns {
		if strings.Contains(bodyLower, p) || strings.Contains(subLower, p) {
			return NewInvalidRequestError("AWS_SES", fmt.Sprintf("outgoing email content violates security policy: contains prohibited sensitive pattern %q", p))
		}
	}

	return nil
}

// SESResolvedConfig holds tenant-resolved or environment-resolved AWS SES credentials.
type SESResolvedConfig struct {
	Region         string `json:"region"`
	FromEmail      string `json:"from_email"`
	AccessKeyID    string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	IsEnabled      bool   `json:"is_enabled"`
}

// SESNotificationProvider implements NotificationProvider for AWS SES.
type SESNotificationProvider struct {
	db         *sqlx.DB
	configRepo ConfigRepository
	httpClient *ResilientHTTPClient
}

// NewSESNotificationProvider constructs an AWS SES provider adapter.
func NewSESNotificationProvider(db *sqlx.DB, configRepo ConfigRepository, httpClient *ResilientHTTPClient) *SESNotificationProvider {
	return &SESNotificationProvider{
		db:         db,
		configRepo: configRepo,
		httpClient: httpClient,
	}
}

func (p *SESNotificationProvider) ProviderName() ProviderName {
	return ProviderAWSSES
}

// ResolveConfig resolves AWS SES configuration with strict tenant isolation.
func (p *SESNotificationProvider) ResolveConfig(ctx context.Context, orgID int64) (*SESResolvedConfig, error) {
	cfg := &SESResolvedConfig{
		Region: "ap-south-1",
	}

	// 1. Try tenant database configuration if present
	if p.configRepo != nil && orgID > 0 {
		tenantCfg, err := p.configRepo.GetConfig(ctx, orgID, TypeEmail, ProviderAWSSES)
		if err == nil && tenantCfg != nil {
			var nonSec struct {
				Region    string `json:"region"`
				FromEmail string `json:"from_address"`
			}
			var sec struct {
				AccessKeyID     string `json:"access_key_id"`
				SecretAccessKey string `json:"secret_value"`
			}
			_ = json.Unmarshal([]byte(tenantCfg.ConfigJSON), &nonSec)
			_ = json.Unmarshal([]byte(tenantCfg.EncryptedSecret), &sec)

			if nonSec.Region != "" {
				cfg.Region = nonSec.Region
			}
			if nonSec.FromEmail != "" {
				cfg.FromEmail = nonSec.FromEmail
			}
			if sec.AccessKeyID != "" && sec.AccessKeyID != "••••••••" {
				cfg.AccessKeyID = sec.AccessKeyID
			}
			if sec.SecretAccessKey != "" && sec.SecretAccessKey != "••••••••" {
				cfg.SecretAccessKey = sec.SecretAccessKey
			}
			cfg.IsEnabled = tenantCfg.IsEnabled
		}
	}

	// 2. Fall back to environment configuration
	if cfg.Region == "" {
		cfg.Region = os.Getenv(EnvAWSRegion)
		if cfg.Region == "" {
			cfg.Region = "ap-south-1"
		}
	}
	if cfg.FromEmail == "" {
		cfg.FromEmail = os.Getenv(EnvSESFromEmail)
	}
	if cfg.AccessKeyID == "" {
		cfg.AccessKeyID = os.Getenv(EnvAWSAccessKeyID)
	}
	if cfg.SecretAccessKey == "" {
		cfg.SecretAccessKey = os.Getenv(EnvAWSSecretKey)
	}

	// Environment feature flag
	envEnabled := os.Getenv(EnvSESEnabled)
	if strings.EqualFold(envEnabled, "true") || envEnabled == "1" {
		cfg.IsEnabled = true
	}

	return cfg, nil
}

// CheckSuppressed verifies if the destination email is in the organization's suppression list.
func (p *SESNotificationProvider) CheckSuppressed(ctx context.Context, orgID int64, email string) error {
	if p.db == nil || orgID <= 0 {
		return nil
	}

	var reason string
	query := `SELECT reason FROM email_suppressions WHERE org_id = ? AND email = ? LIMIT 1`
	err := p.db.GetContext(ctx, &reason, query, orgID, email)
	if err == nil && reason != "" {
		return NewInvalidRequestError("AWS_SES", fmt.Sprintf("recipient %q is suppressed for this organization (reason: %s); delivery cancelled", email, reason))
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		// Log warning, don't fail hard on transient read error
		return nil
	}
	return nil
}

// SendEmail validates parameters, resolves credentials, and dispatches via AWS SES.
func (p *SESNotificationProvider) SendEmail(ctx context.Context, orgID int64, req EmailRequest) (*EmailResponse, error) {
	if orgID <= 0 {
		return nil, NewInvalidRequestError("AWS_SES", "valid org_id is required")
	}

	// 1. Validation
	toEmail, err := ValidateEmailAddress(req.GetRecipient())
	if err != nil {
		return nil, err
	}

	if err := ValidateEmailContent(req.Subject, req.GetBody()); err != nil {
		return nil, err
	}

	// 2. Suppression check
	if err := p.CheckSuppressed(ctx, orgID, toEmail); err != nil {
		return nil, err
	}

	// 3. Resolve configuration
	cfg, err := p.ResolveConfig(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if cfg.FromEmail == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, NewProviderNotConfiguredError("AWS_SES", "Email (missing AWS credentials or from_email)")
	}

	if !cfg.IsEnabled {
		return nil, NewProviderDisabledError("AWS_SES")
	}

	correlationID := req.CorrelationID
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	idempotencyKey := req.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey = fmt.Sprintf("email-%d-%s-%d", orgID, toEmail, time.Now().UnixNano())
	}

	// 4. Initialize AWS SES Client with static credentials
	credProvider := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")
	awsSdkCfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(cfg.Region),
		awsConfig.WithCredentialsProvider(credProvider),
	)
	if err != nil {
		return nil, NewAuthenticationFailedError("AWS_SES", fmt.Sprintf("failed to initialize AWS credentials: %v", err))
	}

	sesClient := ses.NewFromConfig(awsSdkCfg)

	// 5. Construct SES input
	fromEmail := cfg.FromEmail
	if req.FromEmail != "" {
		fromEmail = req.FromEmail
	}

	bodyContent := &sestypes.Body{}
	if req.BodyText != "" {
		bodyContent.Text = &sestypes.Content{Data: aws.String(req.BodyText)}
	}
	if req.BodyHTML != "" {
		bodyContent.Html = &sestypes.Content{Data: aws.String(req.BodyHTML)}
	}
	if bodyContent.Text == nil && bodyContent.Html == nil {
		bodyContent.Text = &sestypes.Content{Data: aws.String(req.GetBody())}
	}

	input := &ses.SendEmailInput{
		Source: aws.String(fromEmail),
		Destination: &sestypes.Destination{
			ToAddresses: []string{toEmail},
		},
		Message: &sestypes.Message{
			Subject: &sestypes.Content{
				Data: aws.String(req.Subject),
			},
			Body: bodyContent,
		},
	}

	// 6. Dispatch via AWS SES with timeout
	sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	output, err := sesClient.SendEmail(sendCtx, input)
	if err != nil {
		// Normalize AWS SES errors
		errStr := err.Error()
		normalizedErr := normalizeSESError(err)

		// Record failed attempt in email_messages
		bodyPreview := req.Subject
		if len(bodyPreview) > 250 {
			bodyPreview = bodyPreview[:250]
		}
		p.persistEmailMessage(ctx, orgID, "", toEmail, fromEmail, req.Subject, bodyPreview, "FAILED", normalizedErr.Code(), errStr, idempotencyKey, correlationID)

		return nil, normalizedErr
	}

	messageID := ""
	if output != nil && output.MessageId != nil {
		messageID = *output.MessageId
	}

	bodyPreview := req.Subject
	if len(bodyPreview) > 250 {
		bodyPreview = bodyPreview[:250]
	}

	// 7. Persist successful submission record (Status: ACCEPTED)
	p.persistEmailMessage(ctx, orgID, messageID, toEmail, fromEmail, req.Subject, bodyPreview, "ACCEPTED", "", "", idempotencyKey, correlationID)

	// 8. Audit event
	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorType:    "SYSTEM",
		ActorName:    "SESNotificationProvider",
		Action:       "SEND_EMAIL",
		Module:       "INTEGRATIONS",
		ResourceType: "AWS_SES",
		ResourceID:   messageID,
		Description:  fmt.Sprintf("Outbound email dispatched to %s with subject %q (Status: ACCEPTED)", toEmail, req.Subject),
		Result:       domain.ResultSuccess,
	})

	return &EmailResponse{
		MessageID:     messageID,
		Provider:      "AWS_SES",
		Status:        "ACCEPTED",
		SentAt:        time.Now().UTC(),
		CorrelationID: correlationID,
	}, nil
}

func (p *SESNotificationProvider) persistEmailMessage(ctx context.Context, orgID int64, messageID, toEmail, fromEmail, subject, preview, status, errCode, errMsg, idempotencyKey, correlationID string) {
	if p.db == nil {
		return
	}

	query := `
		INSERT INTO email_messages (
			org_id, provider, message_id, to_email, from_email, subject, body_preview,
			status, error_code, error_message, idempotency_key, correlation_id, created_at, updated_at
		) VALUES (?, 'AWS_SES', NULLIF(?, ''), ?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			status = VALUES(status),
			message_id = COALESCE(VALUES(message_id), message_id),
			error_code = VALUES(error_code),
			error_message = VALUES(error_message),
			updated_at = NOW()
	`
	_, _ = p.db.ExecContext(ctx, query, orgID, messageID, toEmail, fromEmail, subject, preview, status, errCode, errMsg, idempotencyKey, correlationID)
}

func normalizeSESError(err error) *IntegrationError {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "invalidparametervalue") || strings.Contains(errStr, "malformed"):
		return NewInvalidRequestError("AWS_SES", fmt.Sprintf("invalid recipient or email parameter: %v", err))
	case strings.Contains(errStr, "auth") || strings.Contains(errStr, "credentials") || strings.Contains(errStr, "securitytoken"):
		return NewAuthenticationFailedError("AWS_SES", fmt.Sprintf("AWS SES authentication failed: %v", err))
	case strings.Contains(errStr, "limitexceeded") || strings.Contains(errStr, "rate"):
		return NewRateLimitedError("AWS_SES", 60)
	case strings.Contains(errStr, "messagerejected") || strings.Contains(errStr, "not verified"):
		return NewInvalidRequestError("AWS_SES", fmt.Sprintf("SES sender address or domain not verified: %v", err))
	case strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline"):
		return NewTimeoutError("AWS_SES", "dispatch")
	default:
		return NewProviderError("AWS_SES", fmt.Sprintf("failed to communicate with AWS SES: %v", err), 502)
	}
}

// SendSMS is not handled by the SES provider.
func (p *SESNotificationProvider) SendSMS(ctx context.Context, orgID int64, req SMSRequest) (*SMSResponse, error) {
	return nil, NewProviderNotConfiguredError("AWS_SES", "SMS")
}

// Status checks the operational health of the SES gateway.
func (p *SESNotificationProvider) Status(ctx context.Context, orgID int64) (*ProviderHealth, error) {
	cfg, err := p.ResolveConfig(ctx, orgID)
	if err != nil {
		return &ProviderHealth{
			Provider:       ProviderAWSSES,
			Type:           TypeEmail,
			Status:         StatusConfigInvalid,
			Message:        err.Error(),
			LastCheckedAt:  time.Now().UTC(),
			ProductionSafe: false,
		}, nil
	}

	if !cfg.IsEnabled {
		return &ProviderHealth{
			Provider:       ProviderAWSSES,
			Type:           TypeEmail,
			Status:         StatusDisabled,
			Message:        "AWS SES integration is disabled for this tenant.",
			LastCheckedAt:  time.Now().UTC(),
			ProductionSafe: true,
		}, nil
	}

	if cfg.FromEmail == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return &ProviderHealth{
			Provider:       ProviderAWSSES,
			Type:           TypeEmail,
			Status:         StatusNotConfigured,
			Message:        "AWS SES credentials or verified sender email are not configured.",
			LastCheckedAt:  time.Now().UTC(),
			ProductionSafe: true,
		}, nil
	}

	return &ProviderHealth{
		Provider:       ProviderAWSSES,
		Type:           TypeEmail,
		Status:         StatusHealthy,
		Message:        fmt.Sprintf("AWS SES configured in %s with sender %s", cfg.Region, cfg.FromEmail),
		LastCheckedAt:  time.Now().UTC(),
		ProductionSafe: true,
	}, nil
}
