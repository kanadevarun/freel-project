package integrations

import (
	"context"
	"time"
)

// CompositeNotificationProvider routes SMS to Twilio and Email to AWS SES.
type CompositeNotificationProvider struct {
	smsProvider   NotificationProvider
	emailProvider NotificationProvider
}

// NewCompositeNotificationProvider creates a notification provider combining SMS and Email providers.
func NewCompositeNotificationProvider(sms NotificationProvider, email NotificationProvider) NotificationProvider {
	return &CompositeNotificationProvider{
		smsProvider:   sms,
		emailProvider: email,
	}
}

func (c *CompositeNotificationProvider) ProviderName() ProviderName {
	return "COMPOSITE_NOTIFICATION"
}

func (c *CompositeNotificationProvider) SendSMS(ctx context.Context, orgID int64, req SMSRequest) (*SMSResponse, error) {
	if c.smsProvider == nil {
		return nil, NewProviderNotConfiguredError("Twilio", "SMS")
	}
	return c.smsProvider.SendSMS(ctx, orgID, req)
}

func (c *CompositeNotificationProvider) SendEmail(ctx context.Context, orgID int64, req EmailRequest) (*EmailResponse, error) {
	if c.emailProvider == nil {
		return nil, NewProviderNotConfiguredError("AWS_SES", "Email")
	}
	return c.emailProvider.SendEmail(ctx, orgID, req)
}

func (c *CompositeNotificationProvider) Status(ctx context.Context, orgID int64) (*ProviderHealth, error) {
	if c.emailProvider != nil {
		return c.emailProvider.Status(ctx, orgID)
	}
	if c.smsProvider != nil {
		return c.smsProvider.Status(ctx, orgID)
	}
	return &ProviderHealth{
		Provider:       ProviderUnconfigured,
		Type:           TypeEmail,
		Status:         StatusNotConfigured,
		Message:        "No notification providers configured",
		LastCheckedAt:  time.Now().UTC(),
		ProductionSafe: true,
	}, nil
}
