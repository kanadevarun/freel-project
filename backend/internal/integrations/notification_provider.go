package integrations

import (
	"context"
	"time"
)

// NotificationProvider defines the vendor-agnostic contract for SMS and Email communication.
type NotificationProvider interface {
	ProviderName() ProviderName
	SendSMS(ctx context.Context, orgID int64, req SMSRequest) (*SMSResponse, error)
	SendEmail(ctx context.Context, orgID int64, req EmailRequest) (*EmailResponse, error)
	Status(ctx context.Context, orgID int64) (*ProviderHealth, error)
}

// UnconfiguredNotificationProvider provides an honest, fail-safe implementation when no SMS/Email provider is wired.
// It strictly returns ErrCodeProviderNotConfigured and NEVER generates fake successful dispatch records.
type UnconfiguredNotificationProvider struct {
	Name ProviderName
}

func NewUnconfiguredNotificationProvider(name ProviderName) NotificationProvider {
	if name == "" {
		name = ProviderUnconfigured
	}
	return &UnconfiguredNotificationProvider{Name: name}
}

func (p *UnconfiguredNotificationProvider) ProviderName() ProviderName {
	return p.Name
}

func (p *UnconfiguredNotificationProvider) SendSMS(ctx context.Context, orgID int64, req SMSRequest) (*SMSResponse, error) {
	return nil, NewProviderNotConfiguredError(string(p.Name), "SMS")
}

func (p *UnconfiguredNotificationProvider) SendEmail(ctx context.Context, orgID int64, req EmailRequest) (*EmailResponse, error) {
	return nil, NewProviderNotConfiguredError(string(p.Name), "Email")
}

func (p *UnconfiguredNotificationProvider) Status(ctx context.Context, orgID int64) (*ProviderHealth, error) {
	return &ProviderHealth{
		Provider:       p.Name,
		Type:           TypeSMS,
		Status:         StatusNotConfigured,
		Message:        "Notification provider is not configured. Live outbound messaging is disabled.",
		LastCheckedAt:  time.Now().UTC(),
		ProductionSafe: true,
	}, nil
}
