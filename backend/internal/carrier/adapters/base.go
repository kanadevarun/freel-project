package adapters

import (
	"context"
	"fmt"

	"github.com/freel/backend/internal/carrier/domain"
)

// BaseAdapter provides default scaffolding and capability checking for carrier adapters.
type BaseAdapter struct {
	Code         string
	CarrierSCAC  string
	Capabilities []domain.Capability
}

func (b *BaseAdapter) ProviderCode() string {
	return b.Code
}

func (b *BaseAdapter) SCAC() string {
	return b.CarrierSCAC
}

func (b *BaseAdapter) SupportedCapabilities() []domain.Capability {
	return b.Capabilities
}

func (b *BaseAdapter) HasCapability(cap domain.Capability) bool {
	for _, c := range b.Capabilities {
		if c == cap {
			return true
		}
	}
	return false
}

func (b *BaseAdapter) CheckCapability(cap domain.Capability) error {
	if !b.HasCapability(cap) {
		return NewCapabilityNotImplementedError(b.Code, string(cap))
	}
	return nil
}

func (b *BaseAdapter) SupportsCapability(cap domain.Capability) bool {
	return b.HasCapability(cap)
}

// TestConnection default implementation returns an explicit not-implemented result.
// Adapters for carriers with real external API support must override this method.
func (b *BaseAdapter) TestConnection(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment) (*domain.TestConnectionResult, error) {
	carrierName := b.Code
	if carrierName == "" {
		carrierName = "Carrier"
	}
	return &domain.TestConnectionResult{
		Success:            false,
		Message:            fmt.Sprintf("Carrier %s API integration not yet implemented", carrierName),
		ErrorCode:          domain.ErrCodeUnsupported,
		TestedCapabilities: b.Capabilities,
		TestedEnvironment:  env,
	}, nil
}

func (b *BaseAdapter) GetTracking(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.TrackingRequest) (*domain.NormalizedTrackingResult, error) {
	if !b.HasCapability(domain.CapTracking) {
		return nil, NewCapabilityNotImplementedError(b.Code, string(domain.CapTracking))
	}
	return nil, NewCapabilityNotImplementedError(b.Code, "GetTracking")
}

func (b *BaseAdapter) GetRates(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.RateRequest) ([]domain.NormalizedCarrierRate, error) {
	if !b.HasCapability(domain.CapRates) {
		return nil, NewCapabilityNotImplementedError(b.Code, string(domain.CapRates))
	}
	return nil, NewCapabilityNotImplementedError(b.Code, "GetRates")
}

func (b *BaseAdapter) GetContractRates(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.ContractRateRequest) ([]domain.NormalizedCarrierRate, error) {
	if !b.HasCapability(domain.CapContractRates) {
		return nil, NewCapabilityNotImplementedError(b.Code, string(domain.CapContractRates))
	}
	return nil, NewCapabilityNotImplementedError(b.Code, "GetContractRates")
}

func (b *BaseAdapter) GetSpotRates(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.SpotRateRequest) ([]domain.NormalizedCarrierRate, error) {
	if !b.HasCapability(domain.CapSpotRates) {
		return nil, NewCapabilityNotImplementedError(b.Code, string(domain.CapSpotRates))
	}
	return nil, NewCapabilityNotImplementedError(b.Code, "GetSpotRates")
}

func (b *BaseAdapter) CreateBooking(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.BookingRequest) (*domain.NormalizedBookingResult, error) {
	if !b.HasCapability(domain.CapBooking) {
		return nil, NewCapabilityNotImplementedError(b.Code, string(domain.CapBooking))
	}
	return nil, NewCapabilityNotImplementedError(b.Code, "CreateBooking")
}

func (b *BaseAdapter) GetBooking(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, bookingNumber string) (*domain.NormalizedBookingResult, error) {
	if !b.HasCapability(domain.CapBooking) {
		return nil, NewCapabilityNotImplementedError(b.Code, string(domain.CapBooking))
	}
	return nil, NewCapabilityNotImplementedError(b.Code, "GetBooking")
}

func (b *BaseAdapter) GetDocuments(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.DocumentRequest) ([]domain.NormalizedDocumentResult, error) {
	if !b.HasCapability(domain.CapDocuments) {
		return nil, NewCapabilityNotImplementedError(b.Code, string(domain.CapDocuments))
	}
	return nil, NewCapabilityNotImplementedError(b.Code, "GetDocuments")
}
