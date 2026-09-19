package integrations

import (
	"context"
	"time"
)

// TrackingIntegrationProvider defines the vendor-agnostic contract for carrier API telemetry queries.
type TrackingIntegrationProvider interface {
	ProviderName() ProviderName
	GetTracking(ctx context.Context, orgID int64, carrierSCAC, trackingNumber string) (*TrackingResponse, error)
	GetShipmentStatus(ctx context.Context, orgID int64, carrierSCAC, trackingNumber string) (*TrackingResponse, error)
	Status(ctx context.Context, orgID int64) (*ProviderHealth, error)
}

// UnconfiguredTrackingProvider provides an honest, fail-safe implementation when no live carrier API is configured.
// It strictly returns ErrCodeProviderNotConfigured and NEVER generates fake tracking milestones or coordinates.
type UnconfiguredTrackingProvider struct {
	Name ProviderName
}

func NewUnconfiguredTrackingProvider(name ProviderName) TrackingIntegrationProvider {
	if name == "" {
		name = ProviderUnconfigured
	}
	return &UnconfiguredTrackingProvider{Name: name}
}

func (p *UnconfiguredTrackingProvider) ProviderName() ProviderName {
	return p.Name
}

func (p *UnconfiguredTrackingProvider) GetTracking(ctx context.Context, orgID int64, carrierSCAC, trackingNumber string) (*TrackingResponse, error) {
	return nil, NewProviderNotConfiguredError(string(p.Name), "Carrier Tracking")
}

func (p *UnconfiguredTrackingProvider) GetShipmentStatus(ctx context.Context, orgID int64, carrierSCAC, trackingNumber string) (*TrackingResponse, error) {
	return nil, NewProviderNotConfiguredError(string(p.Name), "Carrier Tracking")
}

func (p *UnconfiguredTrackingProvider) Status(ctx context.Context, orgID int64) (*ProviderHealth, error) {
	return &ProviderHealth{
		Provider:       p.Name,
		Type:           TypeCarrierTracking,
		Status:         StatusNotConfigured,
		Message:        "Carrier tracking provider is not configured. Live EDI/API telemetry is disabled.",
		LastCheckedAt:  time.Now().UTC(),
		ProductionSafe: true,
	}, nil
}
