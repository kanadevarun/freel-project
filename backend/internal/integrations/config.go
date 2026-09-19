package integrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// Standard environment feature flags for external integrations.
const (
	EnvSMSEnabled              = "SMS_ENABLED"
	EnvSESEnabled              = "SES_ENABLED"
	EnvCarrierTrackingEnabled  = "CARRIER_TRACKING_ENABLED"
	EnvTrackingWebhooksEnabled = "TRACKING_WEBHOOKS_ENABLED"
	EnvS3Enabled               = "S3_ENABLED"
	EnvTextractEnabled         = "TEXTRACT_ENABLED"

	// Provider environment variable references
	EnvTwilioAccountSID = "TWILIO_ACCOUNT_SID"
	EnvTwilioAuthToken  = "TWILIO_AUTH_TOKEN"
	EnvTwilioFromNumber = "TWILIO_FROM_NUMBER"

	EnvSESRegion       = "SES_REGION"
	EnvSESFromEmail    = "SES_FROM_EMAIL"
	EnvAWSAccessKeyID  = "AWS_ACCESS_KEY_ID"
	EnvAWSSecretKey    = "AWS_SECRET_ACCESS_KEY"

	EnvWebhookSecretDefault = "WEBHOOK_SIGNING_SECRET"
)

// NonSecretConfig represents safe parameters that can be safely logged and exposed via API.
type NonSecretConfig struct {
	EndpointURL        string `json:"endpoint_url,omitempty"`
	Region             string `json:"region,omitempty"`
	FromIdentifier     string `json:"from_identifier,omitempty"`
	TimeoutSeconds     int    `json:"timeout_seconds"`
	MaxRetries         int    `json:"max_retries"`
	TimestampTolerance int    `json:"timestamp_tolerance_seconds,omitempty"` // for webhooks
}

// SecretConfig represents sensitive credentials that must never be exposed or logged.
type SecretConfig struct {
	APIKey         string `json:"-"`
	AuthToken      string `json:"-"`
	SecretKey      string `json:"-"`
	WebhookSecret  string `json:"-"`
	SMTPPassword   string `json:"-"`
}

// MaskSecret masks sensitive credentials for safe display and logging.
// Example: "AC1234567890abcdef" -> "AC1••••••••def" or "••••••••" if short.
func MaskSecret(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "••••••••"
	}
	prefix := s[:3]
	suffix := s[len(s)-3:]
	return prefix + "••••••••" + suffix
}

// GlobalFeatureFlags captures the active state of all integration feature flags.
type GlobalFeatureFlags struct {
	SMSEnabled              bool `json:"sms_enabled"`
	SESEnabled              bool `json:"ses_enabled"`
	CarrierTrackingEnabled  bool `json:"carrier_tracking_enabled"`
	TrackingWebhooksEnabled bool `json:"tracking_webhooks_enabled"`
	S3Enabled               bool `json:"s3_enabled"`
	TextractEnabled         bool `json:"textract_enabled"`
}

// LoadFeatureFlags reads feature flags from the environment, defaulting safely to disabled (false).
func LoadFeatureFlags() GlobalFeatureFlags {
	smsEnabled := parseBoolEnv(EnvSMSEnabled, false)
	if !smsEnabled {
		smsEnabled = parseBoolEnv("TWILIO_ENABLED", false)
	}

	return GlobalFeatureFlags{
		SMSEnabled:              smsEnabled,
		SESEnabled:              parseBoolEnv(EnvSESEnabled, false),
		CarrierTrackingEnabled:  parseBoolEnv(EnvCarrierTrackingEnabled, false),
		TrackingWebhooksEnabled: parseBoolEnv(EnvTrackingWebhooksEnabled, false),
		S3Enabled:               parseBoolEnv(EnvS3Enabled, false),
		TextractEnabled:         parseBoolEnv(EnvTextractEnabled, false),
	}
}

func parseBoolEnv(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

// ConfigRepository persists and retrieves tenant integration configurations from MariaDB.
type ConfigRepository interface {
	GetConfig(ctx context.Context, orgID int64, itype IntegrationType, provider ProviderName) (*IntegrationConfig, error)
	ListConfigsByOrg(ctx context.Context, orgID int64) ([]IntegrationConfig, error)
	SaveConfig(ctx context.Context, cfg *IntegrationConfig) error
	UpdateHealth(ctx context.Context, orgID int64, itype IntegrationType, provider ProviderName, status IntegrationStatus, msg string) error
}

type sqlConfigRepository struct {
	db *sqlx.DB
}

// NewConfigRepository instantiates a database-backed integration configuration repository.
func NewConfigRepository(db *sqlx.DB) ConfigRepository {
	return &sqlConfigRepository{db: db}
}

func (r *sqlConfigRepository) GetConfig(ctx context.Context, orgID int64, itype IntegrationType, provider ProviderName) (*IntegrationConfig, error) {
	var cfg IntegrationConfig
	query := `
		SELECT id, org_id, integration_type, provider_name, is_enabled, status, config_json, encrypted_secrets, last_health_check, health_message, created_at, updated_at
		FROM external_integration_configs
		WHERE org_id = ? AND integration_type = ? AND provider_name = ?
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &cfg, query, orgID, string(itype), string(provider))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *sqlConfigRepository) ListConfigsByOrg(ctx context.Context, orgID int64) ([]IntegrationConfig, error) {
	var list []IntegrationConfig
	query := `
		SELECT id, org_id, integration_type, provider_name, is_enabled, status, config_json, encrypted_secrets, last_health_check, health_message, created_at, updated_at
		FROM external_integration_configs
		WHERE org_id = ?
		ORDER BY integration_type ASC, provider_name ASC
	`
	err := r.db.SelectContext(ctx, &list, query, orgID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *sqlConfigRepository) SaveConfig(ctx context.Context, cfg *IntegrationConfig) error {
	query := `
		INSERT INTO external_integration_configs (
			org_id, integration_type, provider_name, is_enabled, status, config_json, encrypted_secrets, last_health_check, health_message, created_at, updated_at
		) VALUES (
			:org_id, :integration_type, :provider_name, :is_enabled, :status, :config_json, :encrypted_secrets, :last_health_check, :health_message, NOW(), NOW()
		)
		ON DUPLICATE KEY UPDATE
			is_enabled = VALUES(is_enabled),
			status = VALUES(status),
			config_json = VALUES(config_json),
			encrypted_secrets = VALUES(encrypted_secrets),
			last_health_check = VALUES(last_health_check),
			health_message = VALUES(health_message),
			updated_at = NOW()
	`
	_, err := r.db.NamedExecContext(ctx, query, cfg)
	return err
}

func (r *sqlConfigRepository) UpdateHealth(ctx context.Context, orgID int64, itype IntegrationType, provider ProviderName, status IntegrationStatus, msg string) error {
	now := time.Now().UTC()
	query := `
		UPDATE external_integration_configs
		SET status = ?, health_message = ?, last_health_check = ?, updated_at = NOW()
		WHERE org_id = ? AND integration_type = ? AND provider_name = ?
	`
	_, err := r.db.ExecContext(ctx, query, string(status), msg, now, orgID, string(itype), string(provider))
	return err
}

// ToMaskedConfig converts an IntegrationConfig to a safe DTO, never revealing secret content.
func (cfg *IntegrationConfig) ToMaskedConfig() MaskedIntegrationConfig {
	masked := MaskedIntegrationConfig{
		Type:          cfg.Type,
		Provider:      cfg.Provider,
		IsEnabled:     cfg.IsEnabled,
		Status:        cfg.Status,
		HasSecret:     len(cfg.EncryptedSecret) > 0,
		HealthMessage: cfg.HealthMessage,
		LastCheckedAt: cfg.LastHealthCheck,
		TimeoutSec:    15,
		MaxRetries:    3,
	}

	if masked.HasSecret {
		masked.MaskedSecret = "••••••••"
	}

	if cfg.ConfigJSON != "" {
		var nonSec NonSecretConfig
		if err := json.Unmarshal([]byte(cfg.ConfigJSON), &nonSec); err == nil {
			masked.EndpointURL = nonSec.EndpointURL
			masked.Region = nonSec.Region
			masked.FromAddress = nonSec.FromIdentifier
			if nonSec.TimeoutSeconds > 0 {
				masked.TimeoutSec = nonSec.TimeoutSeconds
			}
			if nonSec.MaxRetries >= 0 {
				masked.MaxRetries = nonSec.MaxRetries
			}
		}
	}

	return masked
}
