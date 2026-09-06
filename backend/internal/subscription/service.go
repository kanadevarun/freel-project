package subscription

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrPlanNotFound         = errors.New("plan not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type Service interface {
	GetAvailablePlans(ctx context.Context) ([]SubscriptionPlan, error)
	GetWorkspace(ctx context.Context, orgID int64) (*SubscriptionWorkspaceResponse, error)
	ChangePlan(ctx context.Context, orgID int64, req ChangePlanRequest) error
	CancelSubscription(ctx context.Context, orgID int64) error
	ReactivateSubscription(ctx context.Context, orgID int64) error
	CreateCheckoutSession(ctx context.Context, orgID int64, req ChangePlanRequest) (string, error)
	CreateCustomerPortalSession(ctx context.Context, orgID int64) (string, error)
	ProcessStripeWebhook(ctx context.Context, payload []byte, signatureHeader string) error
	PreviewPlanChange(ctx context.Context, orgID int64, req ChangePlanRequest) (*PlanChangePreviewResponse, error)
	GetAddonConfigs(ctx context.Context) ([]AddonConfig, error)
	UpdateAddons(ctx context.Context, orgID int64, req UpdateAddonsRequest) error
}

type service struct {
	repo         Repository
	stripeClient *StripeClient
}

func NewService(repo Repository, stripeClient *StripeClient) Service {
	return &service{
		repo:         repo,
		stripeClient: stripeClient,
	}
}

func (s *service) GetAvailablePlans(ctx context.Context) ([]SubscriptionPlan, error) {
	return s.repo.GetAvailablePlans(ctx)
}

func (s *service) GetWorkspace(ctx context.Context, orgID int64) (*SubscriptionWorkspaceResponse, error) {
	sub, err := s.repo.GetSubscriptionByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Default empty workspace if no sub
	if sub == nil {
		return &SubscriptionWorkspaceResponse{
			Usage:    []SubscriptionUsage{},
			Addons:   []SubscriptionAddon{},
			Invoices: []Invoice{},
		}, nil
	}

	plan, err := s.repo.GetPlanByID(ctx, sub.PlanID)
	if err != nil {
		return nil, err
	}

	customer, err := s.repo.GetCustomerByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	pm, err := s.repo.GetPaymentMethodByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	usage, err := s.repo.GetUsageByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	addons, err := s.repo.GetAddonsByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	invoices, err := s.repo.GetInvoicesByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Compute enriched usage based on plan limits
	var planLimits map[string]int
	if plan != nil && plan.Limits != nil {
		_ = json.Unmarshal(plan.Limits, &planLimits)
	}

	usageMap := make(map[string]*SubscriptionUsage)
	for i := range usage {
		usageMap[usage[i].MetricName] = &usage[i]
	}

	var enrichedUsage []SubscriptionUsage
	for metric, limit := range planLimits {
		u, exists := usageMap[metric]
		if !exists {
			u = &SubscriptionUsage{
				OrgID:        orgID,
				MetricName:   metric,
				CurrentUsage: 0,
			}
		}

		if limit == -1 {
			u.Unlimited = true
			u.Percentage = 0
			u.Remaining = -1
		} else {
			u.Unlimited = false
			u.Remaining = limit - u.CurrentUsage
			if u.Remaining < 0 {
				u.Remaining = 0
			}
			if limit > 0 {
				pct := (float64(u.CurrentUsage) / float64(limit)) * 100
				if pct > 100 {
					pct = 100
				}
				u.Percentage = int(pct)
			} else {
				u.Percentage = 100
			}
		}
		l := limit
		u.LimitAmount = &l
		enrichedUsage = append(enrichedUsage, *u)
	}

	return &SubscriptionWorkspaceResponse{
		Subscription:  sub,
		CurrentPlan:   plan,
		Customer:      customer,
		PaymentMethod: pm,
		Usage:         enrichedUsage,
		Addons:        addons,
		Invoices:      invoices,
	}, nil
}

func (s *service) ChangePlan(ctx context.Context, orgID int64, req ChangePlanRequest) error {
	plan, err := s.repo.GetPlanByID(ctx, req.PlanID)
	if err != nil {
		return err
	}
	if plan == nil || !plan.IsActive {
		return ErrPlanNotFound
	}

	sub, err := s.repo.GetSubscriptionByOrgID(ctx, orgID)
	if err != nil {
		return err
	}

	// If no provider subscription exists yet, we should use checkout instead of direct upgrade
	if sub == nil || sub.ProviderSubscriptionID == nil || *sub.ProviderSubscriptionID == "" {
		return errors.New("no active stripe subscription found. use checkout")
	}

	var priceID string
	if req.BillingCycle == "annual" && plan.ProviderPriceIDAnnual != nil {
		priceID = *plan.ProviderPriceIDAnnual
	} else if plan.ProviderPriceIDMonthly != nil {
		priceID = *plan.ProviderPriceIDMonthly
	}

	if priceID == "" {
		return errors.New("plan does not have a configured stripe price")
	}

	err = s.stripeClient.UpdateSubscription(ctx, *sub.ProviderSubscriptionID, priceID)
	if err != nil {
		return err
	}

	// We can update the DB immediately, or rely on webhooks.
	// Optimistic update for better UX:
	sub.PlanID = plan.ID
	sub.BillingCycle = req.BillingCycle
	sub.CancelAtPeriodEnd = false

	return s.repo.UpsertSubscription(ctx, sub)
}

func (s *service) CancelSubscription(ctx context.Context, orgID int64) error {
	sub, err := s.repo.GetSubscriptionByOrgID(ctx, orgID)
	if err != nil {
		return err
	}
	if sub == nil || sub.ProviderSubscriptionID == nil {
		return ErrSubscriptionNotFound
	}

	err = s.stripeClient.CancelSubscription(ctx, *sub.ProviderSubscriptionID)
	if err != nil {
		return err
	}

	return s.repo.CancelSubscription(ctx, orgID)
}

func (s *service) ReactivateSubscription(ctx context.Context, orgID int64) error {
	sub, err := s.repo.GetSubscriptionByOrgID(ctx, orgID)
	if err != nil {
		return err
	}
	if sub == nil || sub.ProviderSubscriptionID == nil {
		return ErrSubscriptionNotFound
	}

	err = s.stripeClient.ReactivateSubscription(ctx, *sub.ProviderSubscriptionID)
	if err != nil {
		return err
	}

	sub.CancelAtPeriodEnd = false
	return s.repo.UpsertSubscription(ctx, sub)
}

func (s *service) CreateCustomerPortalSession(ctx context.Context, orgID int64) (string, error) {
	customer, err := s.repo.GetCustomerByOrgID(ctx, orgID)
	if err != nil {
		return "", err
	}
	if customer == nil {
		return "", errors.New("customer not found for this organization")
	}

	return s.stripeClient.CreateCustomerPortalSession(ctx, customer.ProviderCustomerID)
}

func (s *service) CreateCheckoutSession(ctx context.Context, orgID int64, req ChangePlanRequest) (string, error) {
	plan, err := s.repo.GetPlanByID(ctx, req.PlanID)
	if err != nil {
		return "", err
	}
	if plan == nil || !plan.IsActive {
		return "", ErrPlanNotFound
	}

	var priceID string
	if req.BillingCycle == "annual" && plan.ProviderPriceIDAnnual != nil {
		priceID = *plan.ProviderPriceIDAnnual
	} else if plan.ProviderPriceIDMonthly != nil {
		priceID = *plan.ProviderPriceIDMonthly
	}

	if priceID == "" {
		return "", errors.New("plan does not have a configured stripe price")
	}

	// Get or Create Customer
	customer, err := s.repo.GetCustomerByOrgID(ctx, orgID)
	if err != nil {
		return "", err
	}

	var stripeCustomerID string
	if customer == nil {
		// Create a customer in Stripe
		// We'd ideally pass the company name/email, but orgID is the absolute minimum requirement.
		stripeCustomerID, err = s.stripeClient.CreateCustomer(ctx, "LogisticsHQ Customer", "", orgID)
		if err != nil {
			return "", err
		}

		// Save customer mapping
		now := time.Now()
		err = s.repo.UpsertCustomer(ctx, &BillingCustomer{
			OrgID:              orgID,
			ProviderCustomerID: stripeCustomerID,
			CreatedAt:          now,
			UpdatedAt:          now,
		})
		if err != nil {
			return "", err
		}
	} else {
		stripeCustomerID = customer.ProviderCustomerID
	}

	return s.stripeClient.CreateCheckoutSession(ctx, stripeCustomerID, priceID, orgID)
}

func (s *service) ProcessStripeWebhook(ctx context.Context, payload []byte, signatureHeader string) error {
	return s.stripeClient.ProcessWebhook(ctx, payload, signatureHeader, s.repo)
}

func (s *service) PreviewPlanChange(ctx context.Context, orgID int64, req ChangePlanRequest) (*PlanChangePreviewResponse, error) {
	plan, err := s.repo.GetPlanByID(ctx, req.PlanID)
	if err != nil {
		return nil, err
	}
	if plan == nil || !plan.IsActive {
		return nil, ErrPlanNotFound
	}

	sub, err := s.repo.GetSubscriptionByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	var currentPlanName = "None"
	if sub != nil {
		currPlan, err := s.repo.GetPlanByID(ctx, sub.PlanID)
		if err == nil && currPlan != nil {
			currentPlanName = currPlan.Name
		}
	}

	newPrice := plan.PriceMonthly
	if req.BillingCycle == "annual" {
		newPrice = plan.PriceAnnual
	}

	var proration float64 = 0
	effectiveDate := time.Now().Format("Jan 02, 2006")

	// If there's a Stripe sub, we can simulate an upcoming invoice
	if sub != nil && sub.ProviderSubscriptionID != nil && *sub.ProviderSubscriptionID != "" {
		customer, _ := s.repo.GetCustomerByOrgID(ctx, orgID)
		if customer != nil {
			var priceID string
			if req.BillingCycle == "annual" && plan.ProviderPriceIDAnnual != nil {
				priceID = *plan.ProviderPriceIDAnnual
			} else if plan.ProviderPriceIDMonthly != nil {
				priceID = *plan.ProviderPriceIDMonthly
			}
			if priceID != "" {
				proration, _ = s.stripeClient.GetUpcomingInvoice(ctx, customer.ProviderCustomerID, *sub.ProviderSubscriptionID, priceID)
			}
		}
	}

	return &PlanChangePreviewResponse{
		CurrentPlanName: currentPlanName,
		NewPlanName:     plan.Name,
		BillingCycle:    req.BillingCycle,
		NewPrice:        newPrice,
		ProratedCharge:  proration,
		EffectiveDate:   effectiveDate,
	}, nil
}

func (s *service) GetAddonConfigs(ctx context.Context) ([]AddonConfig, error) {
	return s.repo.GetAddonConfigs(ctx)
}

func (s *service) UpdateAddons(ctx context.Context, orgID int64, req UpdateAddonsRequest) error {
	addonConfig, err := s.repo.GetAddonConfigByID(ctx, req.AddonConfigID)
	if err != nil {
		return err
	}
	if addonConfig == nil {
		return errors.New("addon config not found")
	}

	// Update stripe subscription if applicable
	// We're omitting full stripe integration for add-ons here based on complexity constraints
	
	if req.Quantity > 0 {
		addon := &SubscriptionAddon{
			OrgID:         orgID,
			AddonConfigID: &req.AddonConfigID,
			AddonName:     addonConfig.Name,
			Quantity:      req.Quantity,
			PricePerUnit:  addonConfig.UnitPrice,
		}
		return s.repo.UpsertSubscriptionAddon(ctx, addon)
	} else {
		return s.repo.DeleteSubscriptionAddon(ctx, orgID, req.AddonConfigID)
	}
}
