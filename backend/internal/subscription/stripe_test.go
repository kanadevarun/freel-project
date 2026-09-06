package subscription

import (
	"context"
	"testing"
	"github.com/stripe/stripe-go/v78"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) GetAvailablePlans(ctx context.Context) ([]SubscriptionPlan, error) { return nil, nil }
func (m *mockRepo) GetPlanByID(ctx context.Context, id int64) (*SubscriptionPlan, error) { return nil, nil }
func (m *mockRepo) GetSubscriptionByOrgID(ctx context.Context, id int64) (*OrganizationSubscription, error) { return nil, nil }
func (m *mockRepo) GetCustomerByOrgID(ctx context.Context, id int64) (*BillingCustomer, error) { return nil, nil }
func (m *mockRepo) GetPaymentMethodByOrgID(ctx context.Context, id int64) (*BillingPaymentMethod, error) { return nil, nil }
func (m *mockRepo) GetUsageByOrgID(ctx context.Context, id int64) ([]SubscriptionUsage, error) { return nil, nil }
func (m *mockRepo) GetUsageByMetric(ctx context.Context, id int64, metric string) (*SubscriptionUsage, error) { return nil, nil }
func (m *mockRepo) GetAddonsByOrgID(ctx context.Context, id int64) ([]SubscriptionAddon, error) { return nil, nil }
func (m *mockRepo) GetInvoicesByOrgID(ctx context.Context, id int64) ([]Invoice, error) { return nil, nil }
func (m *mockRepo) UpsertSubscription(ctx context.Context, sub *OrganizationSubscription) error { return nil }
func (m *mockRepo) CancelSubscription(ctx context.Context, id int64) error { return nil }
func (m *mockRepo) UpsertCustomer(ctx context.Context, cust *BillingCustomer) error { return nil }
func (m *mockRepo) IncrementUsage(ctx context.Context, id int64, metric string, amt int) error { return nil }
func (m *mockRepo) UpsertUsage(ctx context.Context, usage *SubscriptionUsage) error { return nil }
func (m *mockRepo) UpsertInvoice(ctx context.Context, inv *Invoice) error { return nil }
func (m *mockRepo) GetAddonConfigs(ctx context.Context) ([]AddonConfig, error) { return nil, nil }
func (m *mockRepo) GetAddonConfigByID(ctx context.Context, id int64) (*AddonConfig, error) { return nil, nil }
func (m *mockRepo) UpsertSubscriptionAddon(ctx context.Context, addon *SubscriptionAddon) error { return nil }
func (m *mockRepo) DeleteSubscriptionAddon(ctx context.Context, orgID int64, configID int64) error { return nil }
func (m *mockRepo) UpsertPaymentMethod(ctx context.Context, pm *BillingPaymentMethod) error {
	args := m.Called(ctx, pm)
	return args.Error(0)
}

func TestHandleCustomerUpdated(t *testing.T) {
	repo := new(mockRepo)
	client := &StripeClient{}

	ctx := context.Background()
	cust := &stripe.Customer{
		Metadata: map[string]string{
			"org_id": "1",
		},
		InvoiceSettings: &stripe.CustomerInvoiceSettings{
			DefaultPaymentMethod: &stripe.PaymentMethod{
				ID: "pm_12345",
			},
		},
	}

	repo.On("UpsertPaymentMethod", ctx, mock.MatchedBy(func(pm *BillingPaymentMethod) bool {
		return pm.OrgID == 1 && pm.ProviderPaymentMethodID == "pm_12345" && pm.IsDefault == true
	})).Return(nil).Once()

	err := client.handleCustomerUpdated(ctx, cust, repo)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
