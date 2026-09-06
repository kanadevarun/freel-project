package rates

import (
	"fmt"
	"time"
)

// CalculateRateValidity deterministically calculates the RateStatus and days until expiry.
func CalculateRateValidity(effectiveDate, expiryDate time.Time, currentStatus string, now time.Time) (string, int) {
	// Normalize now to start of day for clean date comparisons
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	effDate := time.Date(effectiveDate.Year(), effectiveDate.Month(), effectiveDate.Day(), 0, 0, 0, 0, effectiveDate.Location())
	expDate := time.Date(expiryDate.Year(), expiryDate.Month(), expiryDate.Day(), 0, 0, 0, 0, expiryDate.Location())

	daysUntilExpiry := int(expDate.Sub(nowDate).Hours() / 24)

	// If explicit user state is ARCHIVED, retain ARCHIVED
	if currentStatus == RateStatusArchived {
		return RateStatusArchived, daysUntilExpiry
	}

	// If explicit user state is DRAFT and effective date is in the future
	if currentStatus == RateStatusDraft && effDate.After(nowDate) {
		return RateStatusDraft, daysUntilExpiry
	}

	// Expired check
	if expDate.Before(nowDate) {
		return RateStatusExpired, daysUntilExpiry
	}

	// Not yet effective check
	if effDate.After(nowDate) {
		return RateStatusDraft, daysUntilExpiry
	}

	// Expiring soon: within 30 days
	if daysUntilExpiry >= 0 && daysUntilExpiry <= 30 {
		return RateStatusExpiringSoon, daysUntilExpiry
	}

	return RateStatusActive, daysUntilExpiry
}

// FormatValidityText produces human-friendly validity messages.
func FormatValidityText(status string, daysUntilExpiry int, expDate time.Time) string {
	switch status {
	case RateStatusExpired:
		return fmt.Sprintf("Expired (%s)", expDate.Format("02 Jan 2006"))
	case RateStatusExpiringSoon:
		if daysUntilExpiry == 0 {
			return "Expires today"
		} else if daysUntilExpiry == 1 {
			return "Expires tomorrow"
		}
		return fmt.Sprintf("Expires in %d days (%s)", daysUntilExpiry, expDate.Format("02 Jan 2006"))
	case RateStatusDraft:
		return "Draft / Upcoming"
	case RateStatusArchived:
		return "Archived"
	default:
		return fmt.Sprintf("Valid until %s", expDate.Format("02 Jan 2006"))
	}
}
