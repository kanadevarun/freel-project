package integrations

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// WebhookSecurityConfig holds signature and replay rules for webhook verification.
type WebhookSecurityConfig struct {
	SigningSecret      string
	TimestampTolerance time.Duration
	SignatureHeader    string
	TimestampHeader    string
}

// DefaultWebhookSecurityConfig supplies standard defaults.
func DefaultWebhookSecurityConfig(secret string) WebhookSecurityConfig {
	return WebhookSecurityConfig{
		SigningSecret:      secret,
		TimestampTolerance: 5 * time.Minute,
		SignatureHeader:    "X-Signature-SHA256",
		TimestampHeader:    "X-Timestamp",
	}
}

// WebhookGateway manages secure ingress of external webhook events.
type WebhookGateway interface {
	VerifyAndRecord(ctx context.Context, provider ProviderName, orgID int64, headers http.Header, rawBody []byte) (*WebhookVerificationResult, error)
	RouteToDeadLetter(ctx context.Context, orgID int64, provider ProviderName, correlationID string, errCode, errMsg string, headers http.Header, rawBody []byte) error
}

type sqlWebhookGateway struct {
	db     *sqlx.DB
	config WebhookSecurityConfig
}

// NewWebhookGateway creates a database-backed webhook security gateway.
func NewWebhookGateway(db *sqlx.DB, config WebhookSecurityConfig) WebhookGateway {
	if config.TimestampTolerance <= 0 {
		config.TimestampTolerance = 5 * time.Minute
	}
	if config.SignatureHeader == "" {
		config.SignatureHeader = "X-Signature-SHA256"
	}
	return &sqlWebhookGateway{
		db:     db,
		config: config,
	}
}

func (g *sqlWebhookGateway) VerifyAndRecord(ctx context.Context, provider ProviderName, orgID int64, headers http.Header, rawBody []byte) (*WebhookVerificationResult, error) {
	correlationID := uuid.New().String()
	now := time.Now().UTC()

	result := &WebhookVerificationResult{
		Valid:         false,
		Provider:      provider,
		OrgID:         orgID,
		Timestamp:     now,
		CorrelationID: correlationID,
	}

	if provider == "" || provider == ProviderUnconfigured {
		result.RejectionReason = "unknown or unsupported provider"
		_ = g.RouteToDeadLetter(ctx, orgID, provider, correlationID, ErrCodeInvalidRequest, result.RejectionReason, headers, rawBody)
		return result, NewInvalidRequestError(string(provider), result.RejectionReason)
	}

	if len(rawBody) == 0 {
		result.RejectionReason = "empty webhook body"
		_ = g.RouteToDeadLetter(ctx, orgID, provider, correlationID, ErrCodeInvalidRequest, result.RejectionReason, headers, rawBody)
		return result, NewInvalidRequestError(string(provider), result.RejectionReason)
	}

	// 1. Signature Verification (if signing secret configured)
	if g.config.SigningSecret != "" {
		sig := headers.Get(g.config.SignatureHeader)
		if sig == "" {
			sig = headers.Get("X-Hub-Signature-256")
		}
		if sig == "" {
			sig = headers.Get("X-Webhook-Signature")
		}

		if sig == "" {
			result.RejectionReason = "missing signature header"
			_ = g.RouteToDeadLetter(ctx, orgID, provider, correlationID, ErrCodeWebhookSignatureInvalid, result.RejectionReason, headers, rawBody)
			return result, NewWebhookSignatureInvalidError(string(provider))
		}

		// Compute expected HMAC-SHA256
		if !VerifyHMACSHA256(rawBody, sig, g.config.SigningSecret) {
			result.RejectionReason = "invalid webhook signature"
			_ = g.RouteToDeadLetter(ctx, orgID, provider, correlationID, ErrCodeWebhookSignatureInvalid, result.RejectionReason, headers, rawBody)
			return result, NewWebhookSignatureInvalidError(string(provider))
		}
	}

	// 2. Timestamp Tolerance Check (Replay Window Prevention)
	tsHeader := headers.Get(g.config.TimestampHeader)
	if tsHeader != "" {
		if tsInt, err := strconv.ParseInt(tsHeader, 10, 64); err == nil {
			var eventTime time.Time
			if tsInt > 1e11 {
				eventTime = time.UnixMilli(tsInt)
			} else {
				eventTime = time.Unix(tsInt, 0)
			}
			result.Timestamp = eventTime

			drift := math.Abs(now.Sub(eventTime).Seconds())
			if drift > g.config.TimestampTolerance.Seconds() {
				result.RejectionReason = fmt.Sprintf("timestamp drift exceeded tolerance window (drift: %.0fs, tolerance: %.0fs)", drift, g.config.TimestampTolerance.Seconds())
				_ = g.RouteToDeadLetter(ctx, orgID, provider, correlationID, ErrCodeWebhookReplayDetected, result.RejectionReason, headers, rawBody)
				return result, NewWebhookReplayDetectedError(string(provider), "timestamp_expired")
			}
		}
	}

	// 3. Event Deduplication and Replay Attack Protection
	eventID := headers.Get("X-Event-ID")
	if eventID == "" {
		eventID = headers.Get("X-Message-ID")
	}
	result.EventID = eventID

	// Compute payload fingerprint
	fingerprint := ComputePayloadFingerprint(rawBody)

	// Check if this event was already received
	var existingCount int
	if eventID != "" {
		_ = g.db.GetContext(ctx, &existingCount, `
			SELECT COUNT(*) FROM external_webhook_events
			WHERE org_id = ? AND provider_name = ? AND provider_event_id = ?
		`, orgID, string(provider), eventID)
	} else {
		_ = g.db.GetContext(ctx, &existingCount, `
			SELECT COUNT(*) FROM external_webhook_events
			WHERE org_id = ? AND event_fingerprint = ?
		`, orgID, fingerprint)
	}

	if existingCount > 0 {
		result.IsReplay = true
		result.RejectionReason = "duplicate event detected"
		return result, NewWebhookReplayDetectedError(string(provider), eventID)
	}

	// 4. Sanitize preview for safe audit storage (max 256 chars, no secrets)
	preview := string(rawBody)
	if len(preview) > 256 {
		preview = preview[:256] + "..."
	}

	// 5. Persist to external_webhook_events
	query := `
		INSERT INTO external_webhook_events (
			org_id, provider_name, provider_event_id, event_fingerprint, event_type, status, payload_preview, correlation_id, received_at
		) VALUES (
			?, ?, ?, ?, 'INGRESS_EVENT', 'VERIFIED', ?, ?, NOW()
		)
	`
	_, err := g.db.ExecContext(ctx, query, orgID, string(provider), eventID, fingerprint, preview, correlationID)
	if err != nil {
		return result, err
	}

	result.Valid = true
	return result, nil
}

func (g *sqlWebhookGateway) RouteToDeadLetter(ctx context.Context, orgID int64, provider ProviderName, correlationID string, errCode, errMsg string, headers http.Header, rawBody []byte) error {
	preview := string(rawBody)
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}

	query := `
		INSERT INTO external_webhook_dead_letter (
			org_id, provider_name, correlation_id, error_code, error_message, payload_preview, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?, NOW()
		)
	`
	_, err := g.db.ExecContext(ctx, query, orgID, string(provider), correlationID, errCode, errMsg, preview)
	return err
}

// VerifyHMACSHA256 checks if candidate signature matches HMAC-SHA256 of body using the secret.
func VerifyHMACSHA256(body []byte, candidateSig, secret string) bool {
	candidateSig = strings.TrimSpace(candidateSig)
	if strings.HasPrefix(candidateSig, "sha256=") {
		candidateSig = strings.TrimPrefix(candidateSig, "sha256=")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedHex := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(strings.ToLower(candidateSig)), []byte(strings.ToLower(expectedHex)))
}

// ComputePayloadFingerprint returns SHA256 hex digest of the raw byte stream.
func ComputePayloadFingerprint(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
