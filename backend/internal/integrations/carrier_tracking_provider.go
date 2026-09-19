package integrations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	carrierAdapters "github.com/freel/backend/internal/carrier/adapters"
	carrierDomain "github.com/freel/backend/internal/carrier/domain"
	carrierRepo "github.com/freel/backend/internal/carrier/repository"
	carrierService "github.com/freel/backend/internal/carrier/service"
)

var (
	scacRegex        = regexp.MustCompile(`^[A-Z0-9]{2,6}$`)
	trackingNumRegex = regexp.MustCompile(`^[A-Za-z0-9\-_]{4,64}$`)
)

// CarrierGatewayTrackingProvider bridges Task 2.1 TrackingIntegrationProvider to the multi-carrier adapter architecture.
// It enforces organization scoping, credential protection, error normalization, and authentic status reporting.
type CarrierGatewayTrackingProvider struct {
	carrierSvc  carrierService.CarrierService
	carrierRepo carrierRepo.CarrierRepository
	registry    *carrierAdapters.AdapterRegistry
}

// NewCarrierGatewayTrackingProvider creates a new production-hardened carrier tracking integration provider.
func NewCarrierGatewayTrackingProvider(
	svc carrierService.CarrierService,
	repo carrierRepo.CarrierRepository,
) TrackingIntegrationProvider {
	return &CarrierGatewayTrackingProvider{
		carrierSvc:  svc,
		carrierRepo: repo,
		registry:    carrierAdapters.GetDefaultRegistry(),
	}
}

func (p *CarrierGatewayTrackingProvider) ProviderName() ProviderName {
	return ProviderName("CARRIER_GATEWAY")
}

// GetTracking queries the external carrier API or EDI telemetry through the normalized provider adapter.
func (p *CarrierGatewayTrackingProvider) GetTracking(ctx context.Context, orgID int64, carrierSCAC, trackingNumber string) (*TrackingResponse, error) {
	if orgID <= 0 {
		return nil, NewInvalidRequestError("carrier_tracking", "valid org_id is required")
	}

	cleanSCAC := strings.ToUpper(strings.TrimSpace(carrierSCAC))
	cleanTracking := strings.TrimSpace(trackingNumber)

	if !scacRegex.MatchString(cleanSCAC) {
		return nil, NewInvalidRequestError(cleanSCAC, fmt.Sprintf("invalid carrier SCAC %q (expected 2-6 alphanumeric characters)", cleanSCAC))
	}
	if !trackingNumRegex.MatchString(cleanTracking) {
		return nil, NewInvalidRequestError(cleanSCAC, fmt.Sprintf("invalid tracking reference format %q (expected 4-64 alphanumeric characters)", cleanTracking))
	}

	if p.carrierRepo == nil {
		return nil, NewProviderUnavailableError(cleanSCAC, "carrier integration repository is unavailable")
	}

	// 1. Resolve tenant's carrier integration (check Production first, fallback to Sandbox)
	ci, err := p.carrierRepo.GetIntegrationBySCAC(ctx, orgID, cleanSCAC, carrierDomain.EnvProduction)
	if err != nil || ci == nil {
		ci, err = p.carrierRepo.GetIntegrationBySCAC(ctx, orgID, cleanSCAC, carrierDomain.EnvSandbox)
	}
	if err != nil || ci == nil {
		return nil, NewProviderNotConfiguredError(cleanSCAC, "Carrier Tracking")
	}

	if p.carrierSvc == nil {
		return nil, NewProviderUnavailableError(cleanSCAC, "carrier integration service is unavailable")
	}

	// 2. Tenant isolation & enablement verification
	if ci.OrgID != orgID {
		return nil, NewInvalidRequestError(cleanSCAC, "organization boundary violation")
	}
	if !ci.IsEnabled || ci.ConnectionStatus == carrierDomain.StatusDisabled {
		return nil, NewProviderUnavailableError(cleanSCAC, fmt.Sprintf("carrier integration for %s is disabled", cleanSCAC))
	}

	// 3. Capability enforcement
	hasTracking := false
	for _, cap := range ci.Capabilities {
		if cap == carrierDomain.CapTracking {
			hasTracking = true
			break
		}
	}
	if !hasTracking {
		return nil, NewProviderUnavailableError(cleanSCAC, fmt.Sprintf("tracking capability is disabled for carrier connection %s", cleanSCAC))
	}

	// 4. Resolve adapter and decrypted credentials
	adapter, creds, err := p.carrierSvc.GetAdapterForIntegration(ctx, orgID, cleanSCAC, ci.Environment)
	if err != nil {
		if errors.Is(err, carrierService.ErrIntegrationNotFound) {
			return nil, NewProviderNotConfiguredError(cleanSCAC, "Carrier Tracking")
		}
		if errors.Is(err, carrierService.ErrDecryptionFailed) {
			return nil, NewAuthenticationFailedError(cleanSCAC, "failed to decrypt carrier credentials")
		}
		return nil, NewProviderUnavailableError(cleanSCAC, fmt.Sprintf("failed to initialize carrier adapter: %v", err))
	}

	// 5. Query the external carrier API
	req := carrierDomain.TrackingRequest{
		CarrierSCAC:     cleanSCAC,
		ContainerNumber: cleanTracking,
		BookingNumber:   cleanTracking,
	}

	result, err := adapter.GetTracking(ctx, creds, ci.Environment, req)
	if err != nil {
		return nil, p.normalizeCarrierError(cleanSCAC, err)
	}
	if result == nil {
		return nil, NewProviderUnavailableError(cleanSCAC, "carrier returned empty tracking telemetry")
	}

	now := time.Now().UTC()
	corrID := fmt.Sprintf("track-%s-%d", cleanSCAC, now.UnixNano())

	return &TrackingResponse{
		CarrierSCAC:    cleanSCAC,
		TrackingNumber: cleanTracking,
		Status:         result.CurrentStatus,
		CurrentPort:    result.LatestLocation,
		ETD:            result.ActualDeparture,
		ETA:            result.EstimatedArrival,
		CorrelationID:  corrID,
		LastUpdated:    now,
	}, nil
}

// GetShipmentStatus returns the current milestone status for a carrier tracking reference.
func (p *CarrierGatewayTrackingProvider) GetShipmentStatus(ctx context.Context, orgID int64, carrierSCAC, trackingNumber string) (*TrackingResponse, error) {
	return p.GetTracking(ctx, orgID, carrierSCAC, trackingNumber)
}

// Status inspects the operational state and health of carrier integrations for the tenant.
func (p *CarrierGatewayTrackingProvider) Status(ctx context.Context, orgID int64) (*ProviderHealth, error) {
	if orgID <= 0 || p.carrierRepo == nil {
		return &ProviderHealth{
			Provider:       p.ProviderName(),
			Type:           TypeCarrierTracking,
			Status:         StatusNotConfigured,
			Message:        "Carrier tracking is not configured for this organization.",
			LastCheckedAt:  time.Now().UTC(),
			ProductionSafe: true,
		}, nil
	}

	integrations, err := p.carrierRepo.ListIntegrations(ctx, orgID)
	if err != nil || len(integrations) == 0 {
		return &ProviderHealth{
			Provider:       p.ProviderName(),
			Type:           TypeCarrierTracking,
			Status:         StatusNotConfigured,
			Message:        "No carrier integrations are configured. Connect ocean carriers in Settings > Carrier Integrations.",
			LastCheckedAt:  time.Now().UTC(),
			ProductionSafe: true,
		}, nil
	}

	connectedCount := 0
	for _, ci := range integrations {
		if ci.IsEnabled && ci.ConnectionStatus == carrierDomain.StatusConnected {
			connectedCount++
		}
	}

	if connectedCount == 0 {
		return &ProviderHealth{
			Provider:       p.ProviderName(),
			Type:           TypeCarrierTracking,
			Status:         StatusDisabled,
			Message:        fmt.Sprintf("%d carrier integrations configured, but none are currently active or connected.", len(integrations)),
			LastCheckedAt:  time.Now().UTC(),
			ProductionSafe: true,
		}, nil
	}

	return &ProviderHealth{
		Provider:       p.ProviderName(),
		Type:           TypeCarrierTracking,
		Status:         StatusHealthy,
		Message:        fmt.Sprintf("%d active carrier integrations connected and ready for tracking telemetry.", connectedCount),
		LastCheckedAt:  time.Now().UTC(),
		ProductionSafe: true,
	}, nil
}

// normalizeCarrierError converts carrier domain or network errors into sanitized IntegrationError.
func (p *CarrierGatewayTrackingProvider) normalizeCarrierError(scac string, err error) *IntegrationError {
	if err == nil {
		return nil
	}

	var intErr *carrierDomain.IntegrationError
	if errors.As(err, &intErr) {
		switch intErr.ErrorCode {
		case carrierDomain.ErrCodeAuthFailed:
			return NewAuthenticationFailedError(scac, intErr.UserMessage)
		case carrierDomain.ErrCodeForbidden:
			return &IntegrationError{
				CodeValue:       ErrCodeAuthorizationFailed,
				MessageValue:    intErr.UserMessage,
				HTTPStatusValue: http.StatusForbidden,
				ProviderValue:   scac,
				RetryableValue:  false,
			}
		case carrierDomain.ErrCodeRateLimited:
			return NewRateLimitedError(scac, 60)
		case carrierDomain.ErrCodeTimeout:
			return NewTimeoutError(scac, "tracking query timed out")
		case carrierDomain.ErrCodeUnavailable:
			return NewProviderUnavailableError(scac, intErr.UserMessage)
		case carrierDomain.ErrCodeNotFound:
			return &IntegrationError{
				CodeValue:       "tracking_reference_not_found",
				MessageValue:    intErr.UserMessage,
				HTTPStatusValue: http.StatusNotFound,
				ProviderValue:   scac,
				RetryableValue:  false,
			}
		case carrierDomain.ErrCodeInvalidRequest:
			return NewInvalidRequestError(scac, intErr.UserMessage)
		default:
			return NewProviderError(scac, intErr.UserMessage, intErr.HTTPStatus)
		}
	}

	errMsg := err.Error()
	if strings.Contains(strings.ToLower(errMsg), "timeout") || strings.Contains(strings.ToLower(errMsg), "deadline") {
		return NewTimeoutError(scac, "carrier API request timed out")
	}
	if strings.Contains(strings.ToLower(errMsg), "connection refused") || strings.Contains(strings.ToLower(errMsg), "no such host") {
		return NewConnectionFailedError(scac, "unable to connect to carrier gateway")
	}

	return NewProviderError(scac, "carrier tracking query failed", http.StatusBadGateway)
}
