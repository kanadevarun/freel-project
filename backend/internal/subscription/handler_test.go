package subscription

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/freel/backend/internal/middleware"
)

type mockService struct {
	workspace *SubscriptionWorkspaceResponse
	plans     []SubscriptionPlan
	err       error
}

func (m *mockService) GetAvailablePlans(ctx context.Context) ([]SubscriptionPlan, error) {
	return m.plans, m.err
}

func (m *mockService) GetWorkspace(ctx context.Context, orgID int64) (*SubscriptionWorkspaceResponse, error) {
	return m.workspace, m.err
}

func (m *mockService) ChangePlan(ctx context.Context, orgID int64, req ChangePlanRequest) error {
	return m.err
}

func (m *mockService) CancelSubscription(ctx context.Context, orgID int64) error {
	return m.err
}

func (m *mockService) CreateCheckoutSession(ctx context.Context, orgID int64, req ChangePlanRequest) (string, error) {
	return "https://checkout.stripe.com/test", nil
}

func (m *mockService) ProcessStripeWebhook(ctx context.Context, payload []byte, signatureHeader string) error {
	return nil
}

func (m *mockService) ReactivateSubscription(ctx context.Context, orgID int64) error {
	return m.err
}

func (m *mockService) CreateCustomerPortalSession(ctx context.Context, orgID int64) (string, error) {
	return "https://billing.stripe.com/p/session/test", nil
}

func (m *mockService) PreviewPlanChange(ctx context.Context, orgID int64, req ChangePlanRequest) (*PlanChangePreviewResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &PlanChangePreviewResponse{
		CurrentPlanName: "Starter",
		NewPlanName:     "Growth",
		BillingCycle:    "monthly",
		NewPrice:        299.00,
		ProratedCharge:  10.50,
		EffectiveDate:   "Jan 01, 2024",
	}, nil
}

func (m *mockService) GetAddonConfigs(ctx context.Context) ([]AddonConfig, error) {
	return []AddonConfig{{ID: 1, Name: "Test Addon", UnitPrice: 10.0}}, m.err
}

func (m *mockService) UpdateAddons(ctx context.Context, orgID int64, req UpdateAddonsRequest) error {
	return m.err
}

func TestGetWorkspace(t *testing.T) {
	mockSvc := &mockService{
		workspace: &SubscriptionWorkspaceResponse{
			CurrentPlan: &SubscriptionPlan{
				Name: "Starter",
			},
		},
	}

	handler := NewHandler(mockSvc)

	req, err := http.NewRequest("GET", "/api/v1/subscription", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Inject User Context
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, &middleware.UserContext{
		OrgID: 1,
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetWorkspace(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp SubscriptionWorkspaceResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if resp.CurrentPlan == nil || resp.CurrentPlan.Name != "Starter" {
		t.Errorf("handler returned unexpected body: got %v", resp)
	}
}

func TestGetWorkspace_Unauthorized(t *testing.T) {
	mockSvc := &mockService{}
	handler := NewHandler(mockSvc)

	req, err := http.NewRequest("GET", "/api/v1/subscription", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Missing User Context

	rr := httptest.NewRecorder()
	handler.GetWorkspace(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestUpdateAddons(t *testing.T) {
	mockSvc := &mockService{}
	handler := NewHandler(mockSvc)

	req, err := http.NewRequest("POST", "/api/v1/subscription/addons", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	// Inject User Context
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, &middleware.UserContext{
		OrgID: 1,
	})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdateAddons(rr, req)

	// Since `{}` parses to zero-values, and mock returns nil, it returns 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
