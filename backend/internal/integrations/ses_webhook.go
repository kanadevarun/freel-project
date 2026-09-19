package integrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
)

// SESWebhookHandler handles incoming AWS SNS / SES event callbacks for deliveries, bounces, and complaints.
type SESWebhookHandler struct {
	db             *sqlx.DB
	configRepo     ConfigRepository
	webhookGateway WebhookGateway
}

// NewSESWebhookHandler constructs an SES webhook ingress handler.
func NewSESWebhookHandler(db *sqlx.DB, configRepo ConfigRepository, gateway WebhookGateway) *SESWebhookHandler {
	return &SESWebhookHandler{
		db:             db,
		configRepo:     configRepo,
		webhookGateway: gateway,
	}
}

// SNSMessageEnvelope represents the standard outer payload sent by AWS SNS HTTP subscriptions.
type SNSMessageEnvelope struct {
	Type             string `json:"Type"`
	MessageID        string `json:"MessageId"`
	TopicArn         string `json:"TopicArn"`
	Subject          string `json:"Subject,omitempty"`
	Message          string `json:"Message"`
	Timestamp        string `json:"Timestamp"`
	SignatureVersion string `json:"SignatureVersion"`
	Signature        string `json:"Signature"`
	SigningCertURL   string `json:"SigningCertURL"`
	SubscribeURL     string `json:"SubscribeURL,omitempty"`
	Token            string `json:"Token,omitempty"`
}

// SESEventPayload represents the inner SES notification format.
type SESEventPayload struct {
	EventType        string `json:"eventType,omitempty"`
	NotificationType string `json:"notificationType,omitempty"`
	Mail             struct {
		MessageID   string   `json:"messageId"`
		Source      string   `json:"source"`
		Destination []string `json:"destination"`
		Timestamp   string   `json:"timestamp"`
	} `json:"mail"`
	Delivery *struct {
		Timestamp            string   `json:"timestamp"`
		ProcessingTimeMillis int64    `json:"processingTimeMillis"`
		Recipients           []string `json:"recipients"`
		SMTPResponse         string   `json:"smtpResponse"`
	} `json:"delivery,omitempty"`
	Bounce *struct {
		BounceType        string `json:"bounceType"`
		BounceSubType     string `json:"bounceSubType"`
		BouncedRecipients []struct {
			EmailAddress   string `json:"emailAddress"`
			Action         string `json:"action"`
			Status         string `json:"status"`
			DiagnosticCode string `json:"diagnosticCode"`
		} `json:"bouncedRecipients"`
		Timestamp string `json:"timestamp"`
	} `json:"bounce,omitempty"`
	Complaint *struct {
		ComplaintFeedbackType string `json:"complaintFeedbackType"`
		ComplainedRecipients  []struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"complainedRecipients"`
		Timestamp string `json:"timestamp"`
	} `json:"complaint,omitempty"`
}

// HandleSESEvent processes incoming SNS/SES webhook requests.
func (h *SESWebhookHandler) HandleSESEvent(ctx context.Context, req *http.Request) error {
	correlationID := req.Header.Get("X-Correlation-ID")
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		return NewInvalidRequestError("AWS_SES", "failed to read webhook body")
	}
	defer req.Body.Close()

	if len(rawBody) == 0 {
		return NewInvalidRequestError("AWS_SES", "empty webhook payload")
	}

	// 1. Detect if payload is wrapped in an SNS envelope
	var snsMsg SNSMessageEnvelope
	var sesPayload SESEventPayload

	if err := json.Unmarshal(rawBody, &snsMsg); err == nil && snsMsg.Type != "" {
		// Handle SNS Subscription Confirmation
		if snsMsg.Type == "SubscriptionConfirmation" {
			return h.handleSubscriptionConfirmation(snsMsg.SubscribeURL)
		}

		// Parse inner Message JSON string
		if snsMsg.Message != "" {
			if err := json.Unmarshal([]byte(snsMsg.Message), &sesPayload); err != nil {
				return NewInvalidRequestError("AWS_SES", fmt.Sprintf("failed to parse inner SES message: %v", err))
			}
		}
	} else {
		// Attempt direct unmarshal as SES event
		if err := json.Unmarshal(rawBody, &sesPayload); err != nil {
			return NewInvalidRequestError("AWS_SES", fmt.Sprintf("invalid SES event payload: %v", err))
		}
	}

	messageID := sesPayload.Mail.MessageID
	if messageID == "" {
		if h.webhookGateway != nil {
			_ = h.webhookGateway.RouteToDeadLetter(ctx, 1, ProviderAWSSES, correlationID,
				"missing_message_id", "SES event missing mail.messageId", req.Header, rawBody)
		}
		return NewInvalidRequestError("AWS_SES", "missing mail.messageId in event")
	}

	// 2. Query email_messages to resolve tenant context
	type emailRecord struct {
		ID        int64     `db:"id"`
		OrgID     int64     `db:"org_id"`
		Status    string    `db:"status"`
		ToEmail   string    `db:"to_email"`
		Subject   string    `db:"subject"`
		CreatedAt time.Time `db:"created_at"`
	}

	var rec emailRecord
	query := `SELECT id, org_id, status, to_email, subject, created_at FROM email_messages WHERE message_id = ? LIMIT 1`
	err = h.db.GetContext(ctx, &rec, query, messageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Unknown message ID: route to dead letter without mutating arbitrary records
			if h.webhookGateway != nil {
				_ = h.webhookGateway.RouteToDeadLetter(ctx, 1, ProviderAWSSES, correlationID,
					"unknown_message_id", fmt.Sprintf("received event for unknown SES messageId %q", messageID),
					req.Header, rawBody)
			}
			return NewInvalidRequestError("AWS_SES", fmt.Sprintf("SES message ID %q not found in LogisticsHQ records", messageID))
		}
		return NewProviderError("AWS_SES", fmt.Sprintf("database query failed: %v", err), 500)
	}

	eventType := strings.ToUpper(sesPayload.EventType)
	if eventType == "" {
		eventType = strings.ToUpper(sesPayload.NotificationType)
	}
	if eventType == "" {
		eventType = "UNKNOWN"
	}

	// 3. Monotonic delivery status progression & suppression handling
	var newStatus string
	var bounceType, bounceSubType, complaintType *string
	var errCode, errMsg *string

	switch eventType {
	case "DELIVERY":
		log.Printf("[SES Webhook] Processing DELIVERY for msg %s (current status: %s)", messageID, rec.Status)
		// Only advance status if not already bounced/complained
		if rec.Status != "BOUNCED" && rec.Status != "COMPLAINED" {
			newStatus = "DELIVERED"
		} else {
			newStatus = rec.Status
		}
		updateQuery := `UPDATE email_messages SET status = ?, delivered_at = NOW(), updated_at = NOW() WHERE id = ?`
		res, err := h.db.ExecContext(ctx, updateQuery, newStatus, rec.ID)
		if err != nil {
			log.Printf("[SES Webhook] Delivery update error: %v", err)
		} else {
			rows, _ := res.RowsAffected()
			log.Printf("[SES Webhook] Delivery update success: rows affected=%d, newStatus=%s", rows, newStatus)
		}

	case "BOUNCE":
		newStatus = "BOUNCED"
		bType := "Permanent"
		bSubType := "General"
		if sesPayload.Bounce != nil {
			if sesPayload.Bounce.BounceType != "" {
				bType = sesPayload.Bounce.BounceType
			}
			if sesPayload.Bounce.BounceSubType != "" {
				bSubType = sesPayload.Bounce.BounceSubType
			}
		}
		bounceType = &bType
		bounceSubType = &bSubType

		code := "BOUNCE_" + strings.ToUpper(bType)
		msg := fmt.Sprintf("Email bounced (%s/%s)", bType, bSubType)
		errCode = &code
		errMsg = &msg

		updateQuery := `
			UPDATE email_messages 
			SET status = ?, bounce_type = ?, bounce_sub_type = ?, error_code = ?, error_message = ?, bounced_at = NOW(), updated_at = NOW() 
			WHERE id = ?
		`
		_, _ = h.db.ExecContext(ctx, updateQuery, newStatus, bounceType, bounceSubType, errCode, errMsg, rec.ID)

		// Hard bounce protection: Add recipient to suppression list
		if strings.EqualFold(bType, "Permanent") {
			h.addSuppression(ctx, rec.OrgID, rec.ToEmail, "HARD_BOUNCE", messageID)
		}

	case "COMPLAINT":
		newStatus = "COMPLAINED"
		cType := "abuse"
		if sesPayload.Complaint != nil && sesPayload.Complaint.ComplaintFeedbackType != "" {
			cType = sesPayload.Complaint.ComplaintFeedbackType
		}
		complaintType = &cType

		code := "COMPLAINT"
		msg := fmt.Sprintf("Recipient filed complaint (%s)", cType)
		errCode = &code
		errMsg = &msg

		updateQuery := `
			UPDATE email_messages 
			SET status = ?, complaint_feedback_type = ?, error_code = ?, error_message = ?, complained_at = NOW(), updated_at = NOW() 
			WHERE id = ?
		`
		_, _ = h.db.ExecContext(ctx, updateQuery, newStatus, complaintType, errCode, errMsg, rec.ID)

		// Complaint protection: Add recipient to suppression list
		h.addSuppression(ctx, rec.OrgID, rec.ToEmail, "COMPLAINT", messageID)

	case "REJECT":
		newStatus = "REJECTED"
		code := "REJECTED"
		msg := "SES rejected the message before transmission"
		errCode = &code
		errMsg = &msg
		updateQuery := `UPDATE email_messages SET status = ?, error_code = ?, error_message = ?, updated_at = NOW() WHERE id = ?`
		_, _ = h.db.ExecContext(ctx, updateQuery, newStatus, errCode, errMsg, rec.ID)

	case "SEND":
		if rec.Status == "QUEUED" || rec.Status == "ACCEPTED" {
			newStatus = "SENT"
			updateQuery := `UPDATE email_messages SET status = ?, updated_at = NOW() WHERE id = ?`
			_, _ = h.db.ExecContext(ctx, updateQuery, newStatus, rec.ID)
		}
	}

	// 4. Log event in external_webhook_events
	preview := string(rawBody)
	if len(preview) > 500 {
		preview = preview[:500]
	}

	insertEventQuery := `
		INSERT INTO external_webhook_events (
			org_id, provider_name, provider_event_id, event_fingerprint, event_type,
			status, payload_preview, correlation_id, received_at, processed_at
		) VALUES (?, 'AWS_SES', ?, ?, ?, 'PROCESSED', ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE status = 'PROCESSED', processed_at = NOW()
	`
	_, _ = h.db.ExecContext(ctx, insertEventQuery, rec.OrgID, messageID+"-"+eventType, correlationID, eventType, preview, correlationID)

	// 5. Audit record
	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        rec.OrgID,
		ActorType:    "WEBHOOK",
		ActorName:    "AWS_SES_SNS",
		Action:       "EMAIL_STATUS_CALLBACK",
		Module:       "INTEGRATIONS",
		ResourceType: "EMAIL_MESSAGE",
		ResourceID:   messageID,
		Description:  fmt.Sprintf("Email %s delivery status updated to %s for recipient %s", messageID, newStatus, rec.ToEmail),
		Result:       domain.ResultSuccess,
	})

	return nil
}

func (h *SESWebhookHandler) handleSubscriptionConfirmation(subscribeURL string) error {
	if subscribeURL == "" {
		return NewInvalidRequestError("AWS_SES", "missing SubscribeURL in SubscriptionConfirmation")
	}

	parsedURL, err := url.Parse(subscribeURL)
	if err != nil {
		return NewInvalidRequestError("AWS_SES", "invalid SubscribeURL")
	}

	// SSRF Defense: Ensure domain strictly matches *.amazonaws.com
	host := strings.ToLower(parsedURL.Host)
	if !strings.HasSuffix(host, ".amazonaws.com") && host != "amazonaws.com" {
		return NewInvalidRequestError("AWS_SES", fmt.Sprintf("invalid subscription confirmation host %q: not an Amazon domain", host))
	}

	log.Printf("AWS SNS SubscriptionConfirmation received for SES. Confirming via: %s", subscribeURL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(subscribeURL)
	if err != nil {
		return fmt.Errorf("failed to confirm SNS subscription: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("subscription confirmation returned HTTP status %d", resp.StatusCode)
	}

	log.Printf("AWS SNS Subscription successfully confirmed!")
	return nil
}

func (h *SESWebhookHandler) addSuppression(ctx context.Context, orgID int64, email, reason, messageID string) {
	if h.db == nil || orgID <= 0 || email == "" {
		return
	}

	query := `
		INSERT INTO email_suppressions (org_id, email, reason, source_message_id, created_at)
		VALUES (?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE reason = VALUES(reason), source_message_id = VALUES(source_message_id)
	`
	_, _ = h.db.ExecContext(ctx, query, orgID, strings.ToLower(strings.TrimSpace(email)), reason, messageID)
}
