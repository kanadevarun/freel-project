package rfq

import (
	"fmt"
	"strings"

	"github.com/freel/backend/internal/rfq/spec"
)

// EvaluateBookingEligibility evaluates whether an RFQ is ready for carrier booking creation.
// All calculations are 100% deterministic and server-side verified.
func EvaluateBookingEligibility(
	rfq *spec.RFQ,
	quotes []spec.RFQQuote,
	reqs *spec.OperationalReadiness,
	docsSummary *spec.DocumentSummary,
) spec.BookingEligibility {
	eligibility := spec.BookingEligibility{
		IsEligible:           true,
		Reasons:              make([]string, 0),
		MissingPrerequisites: make([]string, 0),
		ReadinessScore:       100,
	}

	if rfq == nil {
		eligibility.IsEligible = false
		eligibility.MissingPrerequisites = append(eligibility.MissingPrerequisites, "RFQ Record Not Found")
		eligibility.Reasons = append(eligibility.Reasons, "No active RFQ record exists.")
		eligibility.ReadinessScore = 0
		return eligibility
	}

	// 1. Check for Approved or Customer-Selected Quote
	var approvedQuote *spec.RFQQuote
	for i := range quotes {
		q := &quotes[i]
		if q.Status == spec.QuoteStatusApproved || q.Status == spec.QuoteStatusSelectedForCustomer {
			approvedQuote = q
			break
		}
	}

	if approvedQuote == nil {
		// Fallback: check if any quote is recommended
		for i := range quotes {
			if quotes[i].IsRecommended || quotes[i].Status == spec.QuoteStatusRecommended {
				eligibility.Reasons = append(eligibility.Reasons, fmt.Sprintf("Quote from %s is recommended but must be approved before booking.", quotes[i].CarrierName))
				break
			}
		}
		eligibility.IsEligible = false
		eligibility.MissingPrerequisites = append(eligibility.MissingPrerequisites, "Commercial Quote Approval Required")
		eligibility.Reasons = append(eligibility.Reasons, "An approved or customer-selected carrier quote is required to create a booking.")
		eligibility.ReadinessScore -= 40
	} else {
		eligibility.ApprovedQuoteID = &approvedQuote.ID
		carrierName := approvedQuote.CarrierName
		eligibility.ApprovedCarrier = &carrierName
		status := approvedQuote.Status
		eligibility.QuoteStatus = &status
		eligibility.Reasons = append(eligibility.Reasons, fmt.Sprintf("Approved carrier quote confirmed with %s.", approvedQuote.CarrierName))
	}

	// 2. Check Origin & Destination
	if rfq.Origin == nil || strings.TrimSpace(*rfq.Origin) == "" || rfq.Destination == nil || strings.TrimSpace(*rfq.Destination) == "" {
		eligibility.IsEligible = false
		eligibility.MissingPrerequisites = append(eligibility.MissingPrerequisites, "Origin and Destination Ports Required")
		eligibility.Reasons = append(eligibility.Reasons, "Port pair routing must be defined before booking.")
		eligibility.ReadinessScore -= 30
	}

	// 3. Check Operational Requirements Blockers
	if reqs != nil && reqs.BlockingCount > 0 {
		eligibility.IsEligible = false
		eligibility.MissingPrerequisites = append(eligibility.MissingPrerequisites, fmt.Sprintf("%d Operational Requirement Blockers", reqs.BlockingCount))
		eligibility.Reasons = append(eligibility.Reasons, fmt.Sprintf("There are %d blocking cargo/trade requirements.", reqs.BlockingCount))
		eligibility.ReadinessScore -= 20
	}

	// 4. Check Mandatory Documents
	if docsSummary != nil && docsSummary.MissingDocuments > 0 {
		// Only block if mandatory documents are unsatisfied
		eligibility.Reasons = append(eligibility.Reasons, fmt.Sprintf("%d compliance documents require review.", docsSummary.MissingDocuments))
	}

	if eligibility.ReadinessScore < 0 {
		eligibility.ReadinessScore = 0
	}

	if eligibility.IsEligible {
		eligibility.Reasons = []string{
			"All operational, trade, and commercial prerequisites satisfied.",
			"Eligible for immediate carrier booking execution.",
		}
	}

	return eligibility
}

// ValidateBookingTransition enforces valid state machine lifecycle transitions.
func ValidateBookingTransition(currentStatus, targetStatus string) error {
	if currentStatus == targetStatus {
		return nil
	}

	validTransitions := map[string][]string{
		spec.BookingStatusDraft: {
			spec.BookingStatusRequested,
			spec.BookingStatusCancelled,
		},
		spec.BookingStatusRequested: {
			spec.BookingStatusPendingConfirmation,
			spec.BookingStatusConfirmed,
			spec.BookingStatusCancelled,
		},
		spec.BookingStatusPendingConfirmation: {
			spec.BookingStatusConfirmed,
			spec.BookingStatusCancelled,
		},
		spec.BookingStatusConfirmed: {
			spec.BookingStatusCompleted,
			spec.BookingStatusCancelled,
		},
		spec.BookingStatusCancelled: {},
		spec.BookingStatusCompleted: {},
	}

	allowed, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("unknown current booking status: %s", currentStatus)
	}

	for _, a := range allowed {
		if a == targetStatus {
			return nil
		}
	}

	return fmt.Errorf("illegal booking transition from %s to %s", currentStatus, targetStatus)
}
