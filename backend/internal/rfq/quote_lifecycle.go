package rfq

import (
	"fmt"
	"time"

	"github.com/freel/backend/internal/rfq/spec"
	"github.com/freel/backend/internal/svcerror"
)

// ValidateQuoteTransition ensures that a status change on an RFQQuote
// follows the deterministic commercial state machine.
func ValidateQuoteTransition(currentStatus, targetStatus string) error {
	if currentStatus == "" {
		currentStatus = spec.QuoteStatusDraft
	}
	if targetStatus == "" {
		return svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	// Idempotent: setting same status is permitted
	if currentStatus == targetStatus {
		return nil
	}

	allowedTransitions := map[string][]string{
		spec.QuoteStatusDraft: {
			spec.QuoteStatusRequested,
			spec.QuoteStatusReceived,
			spec.QuoteStatusUnderReview,
			spec.QuoteStatusRejected,
			spec.QuoteStatusWithdrawn,
		},
		spec.QuoteStatusRequested: {
			spec.QuoteStatusReceived,
			spec.QuoteStatusWithdrawn,
			spec.QuoteStatusRejected,
			spec.QuoteStatusExpired,
		},
		spec.QuoteStatusReceived: {
			spec.QuoteStatusUnderReview,
			spec.QuoteStatusRecommended,
			spec.QuoteStatusRejected,
			spec.QuoteStatusExpired,
			spec.QuoteStatusWithdrawn,
		},
		spec.QuoteStatusUnderReview: {
			spec.QuoteStatusRecommended,
			spec.QuoteStatusApproved,
			spec.QuoteStatusRejected,
			spec.QuoteStatusExpired,
			spec.QuoteStatusWithdrawn,
		},
		spec.QuoteStatusRecommended: {
			spec.QuoteStatusApproved,
			spec.QuoteStatusSelectedForCustomer,
			spec.QuoteStatusUnderReview,
			spec.QuoteStatusRejected,
			spec.QuoteStatusExpired,
			spec.QuoteStatusWithdrawn,
		},
		spec.QuoteStatusApproved: {
			spec.QuoteStatusSelectedForCustomer,
			spec.QuoteStatusUnderReview,
			spec.QuoteStatusExpired,
			spec.QuoteStatusWithdrawn,
		},
		spec.QuoteStatusSelectedForCustomer: {
			spec.QuoteStatusUnderReview,
			spec.QuoteStatusExpired,
			spec.QuoteStatusWithdrawn,
		},
		spec.QuoteStatusRejected: {
			spec.QuoteStatusUnderReview,
			spec.QuoteStatusWithdrawn,
		},
		spec.QuoteStatusExpired: {
			spec.QuoteStatusDraft,
			spec.QuoteStatusUnderReview,
		},
		spec.QuoteStatusWithdrawn: {
			spec.QuoteStatusDraft,
		},
	}

	targets, ok := allowedTransitions[currentStatus]
	if !ok {
		return fmt.Errorf("invalid current quote status %q", currentStatus)
	}

	for _, t := range targets {
		if t == targetStatus {
			return nil
		}
	}

	return fmt.Errorf("cannot transition quote from %s to %s", currentStatus, targetStatus)
}

// EvaluateQuoteValidity computes deterministic validity status and days until expiry.
func EvaluateQuoteValidity(validUntil *time.Time, now time.Time) (string, *int) {
	if validUntil == nil {
		return spec.ValidityValid, nil
	}

	if validUntil.Before(now) {
		days := int(validUntil.Sub(now).Hours() / 24)
		return spec.ValidityExpired, &days
	}

	diff := validUntil.Sub(now)
	days := int(diff.Hours() / 24)
	if days <= 7 {
		return spec.ValidityExpiringSoon, &days
	}

	return spec.ValidityValid, &days
}

// CanRecommendQuote verifies if a quote is eligible to become the operational recommendation.
func CanRecommendQuote(q *spec.RFQQuote, now time.Time) error {
	if q == nil {
		return fmt.Errorf("quote is nil")
	}

	if q.Status == spec.QuoteStatusRejected || q.Status == spec.QuoteStatusWithdrawn || q.Status == spec.QuoteStatusExpired {
		return fmt.Errorf("cannot recommend a %s quote", q.Status)
	}

	if q.ValidUntil != nil && q.ValidUntil.Before(now) {
		return fmt.Errorf("cannot recommend an expired quote")
	}

	if q.BuyPrice <= 0 || q.SellPrice <= 0 {
		return fmt.Errorf("quote missing valid commercial pricing (buy and sell price must be > 0)")
	}

	return nil
}

// CanApproveQuote verifies if a quote is eligible for operational approval.
func CanApproveQuote(q *spec.RFQQuote, now time.Time) error {
	if q == nil {
		return fmt.Errorf("quote is nil")
	}

	if q.Status == spec.QuoteStatusRejected || q.Status == spec.QuoteStatusWithdrawn || q.Status == spec.QuoteStatusExpired {
		return fmt.Errorf("cannot approve a %s quote", q.Status)
	}

	if q.ValidUntil != nil && q.ValidUntil.Before(now) {
		return fmt.Errorf("cannot approve an expired quote (validity expired on %s)", q.ValidUntil.Format("2006-01-02"))
	}

	if q.BuyPrice <= 0 || q.SellPrice <= 0 {
		return fmt.Errorf("quote must have valid commercial buy and sell prices before approval")
	}

	return nil
}

