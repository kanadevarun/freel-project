package subscription

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockEntitlementRepo struct {
	mock.Mock
}

func (m *mockEntitlementRepo) GetAvailablePlans(ctx context.Context) ([]SubscriptionPlan, error) {
	args := m.Called(ctx)
	return args.Get(0).([]SubscriptionPlan), args.Error(1)
}

func (m *mockEntitlementRepo) GetPlanByID(ctx context.Context, planID int64) (*SubscriptionPlan, error) {
	args := m.Called(ctx, planID)
	if val := args.Get(0); val != nil {
		return val.(*SubscriptionPlan), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEntitlementRepo) GetSubscriptionByOrgID(ctx context.Context, orgID int64) (*OrganizationSubscription, error) {
	args := m.Called(ctx, orgID)
	if val := args.Get(0); val != nil {
		return val.(*OrganizationSubscription), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEntitlementRepo) GetCustomerByOrgID(ctx context.Context, orgID int64) (*BillingCustomer, error) {
	args := m.Called(ctx, orgID)
	if val := args.Get(0); val != nil {
		return val.(*BillingCustomer), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEntitlementRepo) GetPaymentMethodByOrgID(ctx context.Context, orgID int64) (*BillingPaymentMethod, error) {
	args := m.Called(ctx, orgID)
	if val := args.Get(0); val != nil {
		return val.(*BillingPaymentMethod), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEntitlementRepo) GetUsageByOrgID(ctx context.Context, orgID int64) ([]SubscriptionUsage, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]SubscriptionUsage), args.Error(1)
}

func (m *mockEntitlementRepo) GetUsageByMetric(ctx context.Context, orgID int64, metricName string) (*SubscriptionUsage, error) {
	args := m.Called(ctx, orgID, metricName)
	if val := args.Get(0); val != nil {
		return val.(*SubscriptionUsage), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEntitlementRepo) GetAddonsByOrgID(ctx context.Context, orgID int64) ([]SubscriptionAddon, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]SubscriptionAddon), args.Error(1)
}

func (m *mockEntitlementRepo) GetInvoicesByOrgID(ctx context.Context, orgID int64) ([]Invoice, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]Invoice), args.Error(1)
}

func (m *mockEntitlementRepo) UpsertSubscription(ctx context.Context, sub *OrganizationSubscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *mockEntitlementRepo) CancelSubscription(ctx context.Context, orgID int64) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}

func (m *mockEntitlementRepo) UpsertCustomer(ctx context.Context, customer *BillingCustomer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *mockEntitlementRepo) IncrementUsage(ctx context.Context, orgID int64, metricName string, amount int) error {
	args := m.Called(ctx, orgID, metricName, amount)
	return args.Error(0)
}

func (m *mockEntitlementRepo) UpsertUsage(ctx context.Context, usage *SubscriptionUsage) error {
	args := m.Called(ctx, usage)
	return args.Error(0)
}

func (m *mockEntitlementRepo) UpsertInvoice(ctx context.Context, invoice *Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *mockEntitlementRepo) GetAddonConfigs(ctx context.Context) ([]AddonConfig, error) {
	args := m.Called(ctx)
	return args.Get(0).([]AddonConfig), args.Error(1)
}

func (m *mockEntitlementRepo) GetAddonConfigByID(ctx context.Context, id int64) (*AddonConfig, error) {
	args := m.Called(ctx, id)
	if val := args.Get(0); val != nil {
		return val.(*AddonConfig), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEntitlementRepo) UpsertSubscriptionAddon(ctx context.Context, addon *SubscriptionAddon) error {
	args := m.Called(ctx, addon)
	return args.Error(0)
}

func (m *mockEntitlementRepo) DeleteSubscriptionAddon(ctx context.Context, orgID int64, addonConfigID int64) error {
	args := m.Called(ctx, orgID, addonConfigID)
	return args.Error(0)
}

func (m *mockEntitlementRepo) UpsertPaymentMethod(ctx context.Context, pm *BillingPaymentMethod) error {
	args := m.Called(ctx, pm)
	return args.Error(0)
}

func TestCheckEntitlement(t *testing.T) {
	repo := new(mockEntitlementRepo)
	svc := NewEntitlementService(repo)
	ctx := context.Background()

	orgID := int64(1)
	planID := int64(10)
	
	limits := map[string]int{
		MetricRFQs:        100,
		MetricTeamMembers: -1, // unlimited
	}
	limitsJSON, _ := json.Marshal(limits)

	sub := &OrganizationSubscription{
		OrgID:  orgID,
		PlanID: planID,
		Status: "active",
	}

	plan := &SubscriptionPlan{
		ID:     planID,
		Limits: limitsJSON,
	}

	t.Run("Below limit", func(t *testing.T) {
		repo.On("GetSubscriptionByOrgID", ctx, orgID).Return(sub, nil).Once()
		repo.On("GetPlanByID", ctx, planID).Return(plan, nil).Once()
		repo.On("GetUsageByMetric", ctx, orgID, MetricRFQs).Return(&SubscriptionUsage{
			MetricName:   MetricRFQs,
			CurrentUsage: 50,
		}, nil).Once()

		err := svc.CheckEntitlement(ctx, orgID, MetricRFQs)
		assert.NoError(t, err)
	})

	t.Run("At limit", func(t *testing.T) {
		repo.On("GetSubscriptionByOrgID", ctx, orgID).Return(sub, nil).Once()
		repo.On("GetPlanByID", ctx, planID).Return(plan, nil).Once()
		repo.On("GetUsageByMetric", ctx, orgID, MetricRFQs).Return(&SubscriptionUsage{
			MetricName:   MetricRFQs,
			CurrentUsage: 100,
		}, nil).Once()

		err := svc.CheckEntitlement(ctx, orgID, MetricRFQs)
		assert.Equal(t, ErrLimitReached, err)
	})

	t.Run("Over limit", func(t *testing.T) {
		repo.On("GetSubscriptionByOrgID", ctx, orgID).Return(sub, nil).Once()
		repo.On("GetPlanByID", ctx, planID).Return(plan, nil).Once()
		repo.On("GetUsageByMetric", ctx, orgID, MetricRFQs).Return(&SubscriptionUsage{
			MetricName:   MetricRFQs,
			CurrentUsage: 101,
		}, nil).Once()

		err := svc.CheckEntitlement(ctx, orgID, MetricRFQs)
		assert.Equal(t, ErrLimitReached, err)
	})

	t.Run("Unlimited", func(t *testing.T) {
		repo.On("GetSubscriptionByOrgID", ctx, orgID).Return(sub, nil).Once()
		repo.On("GetPlanByID", ctx, planID).Return(plan, nil).Once()

		err := svc.CheckEntitlement(ctx, orgID, MetricTeamMembers)
		assert.NoError(t, err)
	})

	t.Run("Period reset", func(t *testing.T) {
		pastTime := time.Now().Add(-24 * time.Hour)
		
		repo.On("GetSubscriptionByOrgID", ctx, orgID).Return(sub, nil).Once()
		repo.On("GetPlanByID", ctx, planID).Return(plan, nil).Once()
		repo.On("GetUsageByMetric", ctx, orgID, MetricRFQs).Return(&SubscriptionUsage{
			MetricName:   MetricRFQs,
			CurrentUsage: 100,
			PeriodEnd:    &pastTime, // expired
		}, nil).Once()
		
		repo.On("UpsertUsage", ctx, mock.AnythingOfType("*subscription.SubscriptionUsage")).Return(nil).Once()

		err := svc.CheckEntitlement(ctx, orgID, MetricRFQs)
		assert.NoError(t, err) // usage was reset to 0 internally, so < 100
	})

	t.Run("Past Due limits access", func(t *testing.T) {
		pastDueSub := &OrganizationSubscription{
			OrgID:  orgID,
			PlanID: planID,
			Status: "past_due",
		}
		repo.On("GetSubscriptionByOrgID", ctx, orgID).Return(pastDueSub, nil).Once()

		err := svc.CheckEntitlement(ctx, orgID, MetricRFQs)
		assert.Equal(t, ErrLimitReached, err)
	})
}
