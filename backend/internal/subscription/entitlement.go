package subscription

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrLimitReached = errors.New("limit reached")
)

type EntitlementService interface {
	CheckEntitlement(ctx context.Context, orgID int64, metricName string) error
	IncrementUsage(ctx context.Context, orgID int64, metricName string, amount int) error
}

type entitlementService struct {
	repo Repository
}

func NewEntitlementService(repo Repository) EntitlementService {
	return &entitlementService{repo: repo}
}

// isPeriodBased returns whether a metric resets every billing cycle.
// Things like "team_members", "storage_gb", "carrier_connections" are absolute.
// Things like "rfqs", "shipments", "ai_email_processing" are period-based.
func isPeriodBased(metricName string) bool {
	switch metricName {
	case MetricRFQs, MetricShipments, MetricAIEmails:
		return true
	default:
		return false
	}
}

func (s *entitlementService) CheckEntitlement(ctx context.Context, orgID int64, metricName string) error {
	// 1. Get the current active subscription
	sub, err := s.repo.GetSubscriptionByOrgID(ctx, orgID)
	if err != nil {
		return err
	}
	if sub == nil || (sub.Status != "active" && sub.Status != "trialing") {
		// If no active or trialing subscription, assume limits are 0.
		// This strictly blocks past_due, canceled, unpaid, and incomplete states.
		return ErrLimitReached
	}

	// 2. Get the plan to find the limit
	plan, err := s.repo.GetPlanByID(ctx, sub.PlanID)
	if err != nil {
		return err
	}
	if plan == nil {
		return ErrLimitReached
	}

	// Parse limits from plan
	var limits map[string]int
	if err := json.Unmarshal(plan.Limits, &limits); err != nil {
		return err
	}

	limit, exists := limits[metricName]
	if !exists {
		// If limit is not defined, we'll assume it's 0 (not allowed)
		return ErrLimitReached
	}

	if limit == -1 {
		return nil // Unlimited
	}

	// 3. Get current usage
	usage, err := s.repo.GetUsageByMetric(ctx, orgID, metricName)
	if err != nil {
		return err
	}

	// 4. Period reset check
	if usage != nil && isPeriodBased(metricName) && usage.PeriodEnd != nil {
		if time.Now().After(*usage.PeriodEnd) {
			// Period expired, reset usage
			usage.CurrentUsage = 0
			usage.PeriodStart = sub.CurrentPeriodStart
			usage.PeriodEnd = sub.CurrentPeriodEnd
			_ = s.repo.UpsertUsage(ctx, usage)
		}
	}

	currentUsage := 0
	if usage != nil {
		currentUsage = usage.CurrentUsage
	}

	// 5. Check limit
	if currentUsage >= limit {
		return ErrLimitReached
	}

	return nil
}

func (s *entitlementService) IncrementUsage(ctx context.Context, orgID int64, metricName string, amount int) error {
	usage, err := s.repo.GetUsageByMetric(ctx, orgID, metricName)
	if err != nil {
		return err
	}

	// If it doesn't exist, create it
	if usage == nil {
		sub, _ := s.repo.GetSubscriptionByOrgID(ctx, orgID)
		var pStart, pEnd *time.Time
		if sub != nil {
			pStart = sub.CurrentPeriodStart
			pEnd = sub.CurrentPeriodEnd
		}
		
		usage = &SubscriptionUsage{
			OrgID:        orgID,
			MetricName:   metricName,
			CurrentUsage: amount,
			PeriodStart:  pStart,
			PeriodEnd:    pEnd,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		return s.repo.UpsertUsage(ctx, usage)
	}

	// Otherwise, just increment
	return s.repo.IncrementUsage(ctx, orgID, metricName, amount)
}
