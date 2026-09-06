package adapters

import (
	"context"

	"github.com/freel/backend/internal/carrier/domain"
)

// CarrierAdapter defines the canonical operational interface that all shipping line adapters implement.
// Modules (Tracking, Rates, Bookings, Documents) communicate exclusively through this interface.
type CarrierAdapter interface {
	// ProviderCode returns the unique provider identifier (e.g. "MAERSK", "MSC", "HAPAG_LLOYD").
	ProviderCode() string

	// SCAC returns the standard SCAC code (e.g. "MAEU", "MSCU", "HLCU").
	SCAC() string

	// SupportedCapabilities declares which features the adapter can execute.
	SupportedCapabilities() []domain.Capability

	// HasCapability checks if a specific capability is supported by this adapter.
	HasCapability(cap domain.Capability) bool

	// TestConnection verifies connectivity, credentials validity, and ping latency against carrier servers.
	TestConnection(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment) (*domain.TestConnectionResult, error)

	// GetTracking queries carrier operational milestones for a given container, booking, or B/L.
	GetTracking(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.TrackingRequest) (*domain.NormalizedTrackingResult, error)

	// GetRates queries standard multi-modal container rates between port pairs.
	GetRates(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.RateRequest) ([]domain.NormalizedCarrierRate, error)

	// GetContractRates queries confidential contracted rates using service contract numbers.
	GetContractRates(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.ContractRateRequest) ([]domain.NormalizedCarrierRate, error)

	// GetSpotRates queries dynamic spot rates with guaranteed vessel space.
	GetSpotRates(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.SpotRateRequest) ([]domain.NormalizedCarrierRate, error)

	// CreateBooking places a space allocation request with the carrier.
	CreateBooking(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.BookingRequest) (*domain.NormalizedBookingResult, error)

	// GetBooking queries carrier allocation status and ETD/ETA milestones for a confirmed booking.
	GetBooking(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, bookingNumber string) (*domain.NormalizedBookingResult, error)

	// GetDocuments retrieves verified shipping documents (B/Ls, arrival notices, booking confirmations).
	GetDocuments(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.DocumentRequest) ([]domain.NormalizedDocumentResult, error)
}
