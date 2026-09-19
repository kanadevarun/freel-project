package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/rbac"
)

func TestEmailValidation(t *testing.T) {
	validEmails := []string{
		"ops@logisticshq.in",
		"customer.care+alerts@example.com",
		"finance-team@freel.io",
		"user123@subdomain.domain.co.uk",
	}

	for _, email := range validEmails {
		norm, err := ValidateEmailAddress(email)
		if err != nil {
			t.Errorf("Expected valid email for %q, got error: %v", email, err)
		}
		if norm == "" {
			t.Errorf("Expected non-empty normalized email for %q", email)
		}
	}

	invalidEmails := []string{
		"",
		"   ",
		"plainaddress",
		"@missingusername.com",
		"username@.com",
		"user@domain..com",
		"user with spaces@domain.com",
		"user@domain,com",
	}

	for _, email := range invalidEmails {
		_, err := ValidateEmailAddress(email)
		if err == nil {
			t.Errorf("Expected validation error for invalid email %q, got nil", email)
		}
	}
}

func TestEmailContentValidation(t *testing.T) {
	// Valid content
	err := ValidateEmailContent("Shipment SH-101 Milestone Update", "Your container MSKU9021841 has arrived at Port of Nhava Sheva.")
	if err != nil {
		t.Errorf("Expected valid content to pass, got error: %v", err)
	}

	// Empty subject
	err = ValidateEmailContent("   ", "Valid body")
	if err == nil {
		t.Errorf("Expected error for empty subject, got nil")
	}

	// Empty body
	err = ValidateEmailContent("Valid Subject", "   ")
	if err == nil {
		t.Errorf("Expected error for empty body, got nil")
	}

	// Secret leak prevention
	leaks := []string{
		"Here is your temporary token: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		"Your credentials are: auth_token=secret12345",
		"Master AWS key: aws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"Database access: password=SuperSecretPassword!",
	}

	for _, leak := range leaks {
		err = ValidateEmailContent("Security Alert", leak)
		if err == nil {
			t.Errorf("Expected secret leakage to be blocked for text %q, got nil", leak)
		}
	}
}

func TestUnconfiguredSESProviderFailsSafely(t *testing.T) {
	provider := NewSESNotificationProvider(nil, nil, nil)
	ctx := context.Background()

	// Sending via unconfigured provider must return ErrCodeProviderNotConfigured
	_, err := provider.SendEmail(ctx, 2, EmailRequest{
		ToRecipient: "customer@example.com",
		Subject:     "Testing unconfigured dispatch",
		BodyText:    "This should fail safely.",
	})

	if err == nil {
		t.Fatalf("Expected unconfigured provider to fail, got success")
	}

	var intErr *IntegrationError
	if !errors.As(err, &intErr) {
		t.Fatalf("Expected IntegrationError, got %T: %v", err, err)
	}

	if intErr.Code() != ErrCodeProviderNotConfigured {
		t.Errorf("Expected code %s, got: %s", ErrCodeProviderNotConfigured, intErr.Code())
	}
}

func TestCompositeNotificationProvider(t *testing.T) {
	smsProv := NewUnconfiguredNotificationProvider(ProviderTwilio)
	emailProv := NewUnconfiguredNotificationProvider(ProviderAWSSES)
	composite := NewCompositeNotificationProvider(smsProv, emailProv)

	ctx := context.Background()

	// Test SMS route
	_, err := composite.SendSMS(ctx, 1, SMSRequest{
		RecipientPhoneNumber: "+14155552671",
		Body:                 "Test SMS",
	})
	if err == nil {
		t.Errorf("Expected unconfigured SMS dispatch to fail")
	}

	// Test Email route
	_, err = composite.SendEmail(ctx, 1, EmailRequest{
		ToRecipient: "test@example.com",
		Subject:     "Test",
		BodyText:    "Test",
	})
	if err == nil {
		t.Errorf("Expected unconfigured Email dispatch to fail")
	}
}

func TestSendEmailAction(t *testing.T) {
	mockSvc := &mockGatewayServiceForEmail{
		sendEmailFn: func(ctx context.Context, orgID int64, req EmailRequest) (*EmailResponse, error) {
			return &EmailResponse{
				MessageID:     "<test-ses-msg-12345@email.amazonses.com>",
				Provider:      "AWS_SES",
				Status:        "ACCEPTED",
				SentAt:        time.Now().UTC(),
				CorrelationID: req.CorrelationID,
			}, nil
		},
	}

	action := NewSendEmailAction(mockSvc)
	if action.Name() != "notifications.send_email" {
		t.Errorf("Expected name 'notifications.send_email', got %s", action.Name())
	}

	res, act := action.RequiredPermission()
	if res != rbac.ResourceOutreach || act != rbac.ActionCreate {
		t.Errorf("Unexpected required permission: %s:%s", res, act)
	}

	payload, _ := json.Marshal(SendEmailActionInput{
		ToRecipient:   "customer@acme.com",
		Subject:       "Quotation Q-101 Prepared",
		BodyText:      "Please find your quotation details attached.",
		CorrelationID: "corr-test-email-001",
	})

	actionCtx := &actions.ActionContext{
		Context:        context.Background(),
		OrganizationID: 2,
		RequestID:      "req-test-email",
	}

	result, err := action.Execute(actionCtx, payload)
	if err != nil {
		t.Fatalf("Action execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected successful action result, got failure: %+v", result.Error)
	}
}

func TestSESWebhookSSRFDefense(t *testing.T) {
	handler := NewSESWebhookHandler(nil, nil, nil)

	// Malicious / non-Amazon host in SubscribeURL must be rejected
	err := handler.handleSubscriptionConfirmation("http://malicious-attacker.com/steal-creds")
	if err == nil {
		t.Errorf("Expected SSRF attempt on non-Amazon host to be rejected")
	}

	// Invalid URL
	err = handler.handleSubscriptionConfirmation("://invalid-url")
	if err == nil {
		t.Errorf("Expected invalid URL to be rejected")
	}
}

func TestSESWebhookPayloadParsing(t *testing.T) {
	handler := NewSESWebhookHandler(nil, nil, nil)

	// Missing message ID payload
	req := httptest.NewRequest("POST", "/api/v1/integrations/webhooks/ses", bytes.NewReader([]byte(`{"eventType":"Delivery"}`)))
	err := handler.HandleSESEvent(context.Background(), req)
	if err == nil {
		t.Errorf("Expected error for missing mail.messageId")
	}

	// Empty payload
	reqEmpty := httptest.NewRequest("POST", "/api/v1/integrations/webhooks/ses", bytes.NewReader([]byte(``)))
	err = handler.HandleSESEvent(context.Background(), reqEmpty)
	if err == nil {
		t.Errorf("Expected error for empty webhook payload")
	}
}

type mockGatewayServiceForEmail struct {
	GatewayService
	sendEmailFn func(ctx context.Context, orgID int64, req EmailRequest) (*EmailResponse, error)
}

func (m *mockGatewayServiceForEmail) SendEmail(ctx context.Context, orgID int64, req EmailRequest) (*EmailResponse, error) {
	if m.sendEmailFn != nil {
		return m.sendEmailFn(ctx, orgID, req)
	}
	return nil, nil
}
