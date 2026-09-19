package integrations

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/freel/backend/internal/carrier/domain"
	carrierRepo "github.com/freel/backend/internal/carrier/repository"
)

// mockCarrierRepo implements carrierRepo.CarrierRepository for unit testing.
type mockCarrierRepo struct {
	carrierRepo.CarrierRepository
	integrations []domain.CarrierIntegration
	getErr       error
}

func (m *mockCarrierRepo) GetIntegrationBySCAC(ctx context.Context, orgID int64, scac string, env domain.Environment) (*domain.CarrierIntegration, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, ci := range m.integrations {
		if ci.OrgID == orgID && ci.CarrierSCAC == scac && ci.Environment == env {
			return &ci, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockCarrierRepo) ListIntegrations(ctx context.Context, orgID int64) ([]domain.CarrierIntegration, error) {
	var result []domain.CarrierIntegration
	for _, ci := range m.integrations {
		if ci.OrgID == orgID {
			result = append(result, ci)
		}
	}
	return result, nil
}

func TestCarrierGatewayTrackingProvider_Validation(t *testing.T) {
	repo := &mockCarrierRepo{}
	p := NewCarrierGatewayTrackingProvider(nil, repo)

	ctx := context.Background()

	// 1. Invalid Org ID
	_, err := p.GetTracking(ctx, 0, "MAEU", "MSKU1234567")
	if err == nil {
		t.Fatal("expected error for org_id <= 0")
	}

	// 2. Invalid SCAC format
	_, err = p.GetTracking(ctx, 1, "TOOLONGSCAC", "MSKU1234567")
	if err == nil {
		t.Fatal("expected error for invalid SCAC format")
	}
	var intErr *IntegrationError
	if !errors.As(err, &intErr) || intErr.HTTPStatus() != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400 for invalid SCAC, got %v", err)
	}

	// 3. Invalid tracking reference format
	_, err = p.GetTracking(ctx, 1, "MAEU", "ab")
	if err == nil {
		t.Fatal("expected error for short tracking reference")
	}
	if !errors.As(err, &intErr) || intErr.HTTPStatus() != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400 for invalid tracking number, got %v", err)
	}
}

func TestCarrierGatewayTrackingProvider_Unconfigured(t *testing.T) {
	repo := &mockCarrierRepo{
		integrations: []domain.CarrierIntegration{}, // Empty
	}
	p := NewCarrierGatewayTrackingProvider(nil, repo)

	ctx := context.Background()
	_, err := p.GetTracking(ctx, 1, "MAEU", "MSKU1234567")
	if err == nil {
		t.Fatal("expected error for unconfigured carrier")
	}

	var intErr *IntegrationError
	if !errors.As(err, &intErr) {
		t.Fatalf("expected IntegrationError, got %T", err)
	}

	if intErr.Code() != ErrCodeProviderNotConfigured {
		t.Errorf("expected ErrCodeProviderNotConfigured, got %s", intErr.Code())
	}
	if intErr.HTTPStatus() != http.StatusServiceUnavailable {
		t.Errorf("expected HTTP 503, got %d", intErr.HTTPStatus())
	}
}

func TestCarrierGatewayTrackingProvider_TenantIsolation(t *testing.T) {
	// Configured for Org 2 only
	repo := &mockCarrierRepo{
		integrations: []domain.CarrierIntegration{
			{
				ID:               1,
				OrgID:            2,
				CarrierSCAC:      "MSCU",
				Environment:      domain.EnvProduction,
				IsEnabled:        true,
				ConnectionStatus: domain.StatusConnected,
				Capabilities:     []domain.Capability{domain.CapTracking},
			},
		},
	}
	p := NewCarrierGatewayTrackingProvider(nil, repo)
	ctx := context.Background()

	// Org 1 querying MSCU must receive provider_not_configured
	_, err := p.GetTracking(ctx, 1, "MSCU", "MEDU1234567")
	if err == nil {
		t.Fatal("expected Org 1 to not be able to use Org 2 carrier integration")
	}
	var intErr *IntegrationError
	if !errors.As(err, &intErr) || intErr.Code() != ErrCodeProviderNotConfigured {
		t.Fatalf("expected provider_not_configured for cross-tenant query, got %v", err)
	}
}

func TestCarrierGatewayTrackingProvider_StatusHealth(t *testing.T) {
	repo := &mockCarrierRepo{
		integrations: []domain.CarrierIntegration{
			{
				ID:               1,
				OrgID:            1,
				CarrierSCAC:      "MAEU",
				Environment:      domain.EnvProduction,
				IsEnabled:        true,
				ConnectionStatus: domain.StatusConnected,
			},
		},
	}
	p := NewCarrierGatewayTrackingProvider(nil, repo)
	ctx := context.Background()

	// Org 1 has 1 connected integration
	h1, err := p.Status(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error checking status: %v", err)
	}
	if h1.Status != StatusHealthy {
		t.Errorf("expected StatusHealthy for Org 1, got %s", h1.Status)
	}

	// Org 2 has 0 connected integrations
	h2, err := p.Status(ctx, 2)
	if err != nil {
		t.Fatalf("unexpected error checking status: %v", err)
	}
	if h2.Status != StatusNotConfigured {
		t.Errorf("expected StatusNotConfigured for Org 2, got %s", h2.Status)
	}
}
