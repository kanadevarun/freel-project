package quotations

import (
	"fmt"
	"time"
)

// Allowed status transitions mapping:
// currentStatus -> map[targetStatus]bool
var allowedTransitions = map[string]map[string]bool{
	QuotationStatusDraft: {
		QuotationStatusReadyForReview:   true,
		QuotationStatusSent:             true, // When approval is not required
		QuotationStatusCancelled:        true,
		QuotationStatusExpired:          true,
	},
	QuotationStatusReadyForReview: {
		QuotationStatusApproved:         true,
		QuotationStatusChangesRequested: true,
		QuotationStatusCancelled:        true,
		QuotationStatusExpired:          true,
	},
	QuotationStatusChangesRequested: {
		QuotationStatusReadyForReview:   true,
		QuotationStatusDraft:            true,
		QuotationStatusCancelled:        true,
		QuotationStatusExpired:          true,
	},
	QuotationStatusApproved: {
		QuotationStatusSent:             true,
		QuotationStatusDraft:            true, // Allow pulling back if not sent yet
		QuotationStatusCancelled:        true,
		QuotationStatusExpired:          true,
	},
	QuotationStatusSent: {
		QuotationStatusViewed:           true,
		QuotationStatusAccepted:         true,
		QuotationStatusDeclined:         true,
		QuotationStatusRejected:         true, // Alias for Declined
		QuotationStatusExpired:          true,
		QuotationStatusCancelled:        true,
	},
	QuotationStatusViewed: {
		QuotationStatusAccepted:         true,
		QuotationStatusDeclined:         true,
		QuotationStatusRejected:         true, // Alias for Declined
		QuotationStatusExpired:          true,
		QuotationStatusCancelled:        true,
	},
	QuotationStatusAccepted: {
		// Terminal state (except exceptional administrative cancellation)
	},
	QuotationStatusDeclined: {
		// Terminal state
	},
	QuotationStatusRejected: {
		// Terminal state
	},
	QuotationStatusExpired: {
		// Terminal state
	},
	QuotationStatusCancelled: {
		// Terminal state
	},
}

// CanTransitionQuotationStatus verifies whether moving from currentStatus to targetStatus is valid
// according to the quotation lifecycle state machine and organization approval policies.
func CanTransitionQuotationStatus(currentStatus, targetStatus string, approvalRequired bool) error {
	if currentStatus == targetStatus {
		return nil // No-op transition
	}

	targets, ok := allowedTransitions[currentStatus]
	if !ok || !targets[targetStatus] {
		return fmt.Errorf("invalid quotation status transition from %s to %s", currentStatus, targetStatus)
	}

	// If transitioning from DRAFT directly to SENT, verify approval requirement
	if currentStatus == QuotationStatusDraft && targetStatus == QuotationStatusSent {
		if approvalRequired {
			return fmt.Errorf("approval is required before sending quotation; submit for review first")
		}
	}

	return nil
}

// IsQuotationCommerciallyEditable returns true if commercial pricing, charges, templates,
// or commercial terms are allowed to be modified for this quotation.
func IsQuotationCommerciallyEditable(status string) bool {
	return EditableStatuses[status]
}

// CalculateQuotationValidityStatus determines the current commercial validity badge
// based on the validUntil date.
func CalculateQuotationValidityStatus(validUntil *time.Time) string {
	if validUntil == nil {
		return QuotationValidityNotSet
	}

	now := time.Now().UTC()
	// Compare dates at midnight boundary
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	validDate := time.Date(validUntil.Year(), validUntil.Month(), validUntil.Day(), 23, 59, 59, 0, time.UTC)

	if today.After(validDate) {
		return QuotationValidityExpired
	}

	daysRemaining := validDate.Sub(now).Hours() / 24
	if daysRemaining <= 7 {
		return QuotationValidityExpiringSoon
	}

	return QuotationValidityActive
}
