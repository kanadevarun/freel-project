package integrations

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ValidateTwilioSignature validates Twilio's standard HMAC-SHA1 request signature.
// See Twilio Security Documentation: https://www.twilio.com/docs/usage/webhooks/webhooks-security
func ValidateTwilioSignature(authToken, expectedURL string, form url.Values, signature string) bool {
	if authToken == "" || signature == "" || expectedURL == "" {
		return false
	}

	// 1. Sort all POST parameter keys alphabetically
	keys := make([]string, 0, len(form))
	for k := range form {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. Concatenate URL with sorted key + value pairs
	var buf strings.Builder
	buf.WriteString(expectedURL)
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString(form.Get(k))
	}

	// 3. Compute HMAC-SHA1 with AuthToken
	mac := hmac.New(sha1.New, []byte(authToken))
	mac.Write([]byte(buf.String()))
	expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// TwilioWebhookHandler processes incoming Twilio delivery status callbacks.
type TwilioWebhookHandler struct {
	db             *sqlx.DB
	configRepo     ConfigRepository
	webhookGateway WebhookGateway
}

// NewTwilioWebhookHandler creates a Twilio status callback processor.
func NewTwilioWebhookHandler(db *sqlx.DB, configRepo ConfigRepository, gateway WebhookGateway) *TwilioWebhookHandler {
	return &TwilioWebhookHandler{
		db:             db,
		configRepo:     configRepo,
		webhookGateway: gateway,
	}
}

// HandleStatusCallback processes Twilio delivery status updates with signature verification and tenant isolation.
func (h *TwilioWebhookHandler) HandleStatusCallback(ctx context.Context, req *http.Request) error {
	correlationID := req.Header.Get("X-Correlation-ID")
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	// 1. Parse form values
	if err := req.ParseForm(); err != nil {
		return NewInvalidRequestError("TWILIO", "failed to parse form values from Twilio webhook")
	}

	form := req.PostForm
	messageSID := form.Get("MessageSid")
	if messageSID == "" {
		messageSID = form.Get("SmsSid")
	}
	if messageSID == "" {
		return NewInvalidRequestError("TWILIO", "missing MessageSid parameter in Twilio webhook")
	}

	rawStatus := strings.ToUpper(form.Get("MessageStatus"))
	if rawStatus == "" {
		rawStatus = strings.ToUpper(form.Get("SmsStatus"))
	}
	if rawStatus == "" {
		rawStatus = "UNKNOWN"
	}

	errorCode := form.Get("ErrorCode")
	errorMessage := form.Get("ErrorMessage")

	// 2. Find existing message in database to resolve tenant (OrgID) and verify ownership
	type smsRecord struct {
		ID            int64     `db:"id"`
		OrgID         int64     `db:"org_id"`
		Status        string    `db:"status"`
		ToPhone       string    `db:"to_phone"`
		FromPhone     string    `db:"from_phone"`
		IdempotencyKey string   `db:"idempotency_key"`
		DeliveredAt   *time.Time `db:"delivered_at"`
	}

	var rec smsRecord
	query := `SELECT id, org_id, status, to_phone, from_phone, idempotency_key, delivered_at FROM sms_messages WHERE message_sid = ? LIMIT 1`
	err := h.db.GetContext(ctx, &rec, query, messageSID)
	if err != nil {
		// Unknown message SID: Reject safely and record in dead letter queue
		if h.webhookGateway != nil {
			_ = h.webhookGateway.RouteToDeadLetter(ctx, 1, ProviderTwilio, correlationID,
				"unknown_message_sid", fmt.Sprintf("received status update for unknown message SID %q", messageSID),
				req.Header, []byte(form.Encode()))
		}
		return NewInvalidRequestError("TWILIO", fmt.Sprintf("message SID %q not found in LogisticsHQ records", messageSID))
	}

	orgID := rec.OrgID

	// 3. Resolve auth token for signature verification
	authToken := ""
	if h.configRepo != nil && orgID > 0 {
		tenantCfg, _ := h.configRepo.GetConfig(ctx, orgID, TypeSMS, ProviderTwilio)
		if tenantCfg != nil {
			var sec struct {
				AuthToken string `json:"auth_token"`
			}
			_ = json.Unmarshal([]byte(tenantCfg.EncryptedSecret), &sec)
			if sec.AuthToken != "" && sec.AuthToken != "••••••••" {
				authToken = sec.AuthToken
			}
		}
	}
	if authToken == "" {
		authToken = os.Getenv(EnvTwilioAuthToken)
	}

	// 4. Verify Twilio signature if signature header is present or in production mode
	sigHeader := req.Header.Get("X-Twilio-Signature")
	if sigHeader != "" && authToken != "" {
		// Reconstruct original request URL
		scheme := "https"
		if req.TLS == nil && req.Header.Get("X-Forwarded-Proto") != "https" {
			scheme = "http"
		}
		expectedURL := fmt.Sprintf("%s://%s%s", scheme, req.Host, req.URL.Path)

		if !ValidateTwilioSignature(authToken, expectedURL, form, sigHeader) {
			if h.webhookGateway != nil {
				_ = h.webhookGateway.RouteToDeadLetter(ctx, orgID, ProviderTwilio, correlationID,
					"invalid_signature", "Twilio HMAC-SHA1 signature verification failed",
					req.Header, []byte(form.Encode()))
			}
			return NewWebhookSignatureInvalidError("TWILIO")
		}
	}

	// 5. Monotonic state transition & Replay defense
	// If already DELIVERED, ignore redundant callbacks (e.g. out of order SENT)
	if rec.Status == "DELIVERED" && rawStatus != "DELIVERED" {
		return nil
	}

	var deliveredAt *time.Time
	if rawStatus == "DELIVERED" {
		now := time.Now().UTC()
		deliveredAt = &now
	}

	// 6. Update database record idempotently
	updateQuery := `
		UPDATE sms_messages
		SET status = ?,
			error_code = COALESCE(NULLIF(?, ''), error_code),
			error_message = COALESCE(NULLIF(?, ''), error_message),
			delivered_at = COALESCE(?, delivered_at),
			updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err = h.db.ExecContext(ctx, updateQuery, rawStatus, errorCode, errorMessage, deliveredAt, rec.ID, orgID)
	if err != nil {
		return NewProviderError("TWILIO", fmt.Sprintf("failed to update SMS delivery record: %v", err), http.StatusInternalServerError)
	}

	// 7. Record event in external_webhook_events
	eventID := fmt.Sprintf("%s_%s_%d", messageSID, rawStatus, time.Now().Unix())
	insertEvent := `
		INSERT INTO external_webhook_events (
			org_id, provider_name, provider_event_id, event_fingerprint,
			event_type, signature, status, correlation_id, received_at, processed_at
		) VALUES (?, 'TWILIO', ?, ?, ?, ?, 'PROCESSED', ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE status = 'PROCESSED', processed_at = NOW()
	`
	fingerprint := fmt.Sprintf("twilio-%s-%s", messageSID, rawStatus)
	_, _ = h.db.ExecContext(ctx, insertEvent, orgID, eventID, fingerprint, "SMS_STATUS_"+rawStatus, sigHeader, correlationID)

	return nil
}
