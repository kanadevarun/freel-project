package rates

import (
	"time"
)

// CalculateRateContractStatus deterministically computes contract status based on date validity
func CalculateRateContractStatus(effectiveDate, expiryDate time.Time, currentStatus string) string {
	if currentStatus == ContractStatusArchived {
		return ContractStatusArchived
	}
	if currentStatus == ContractStatusDraft {
		return ContractStatusDraft
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	eff := time.Date(effectiveDate.Year(), effectiveDate.Month(), effectiveDate.Day(), 0, 0, 0, 0, time.UTC)
	exp := time.Date(expiryDate.Year(), expiryDate.Month(), expiryDate.Day(), 0, 0, 0, 0, time.UTC)

	// If contract is yet to become effective
	if today.Before(eff) {
		return ContractStatusDraft
	}

	// If contract has passed expiry date
	if today.After(exp) {
		return ContractStatusExpired
	}

	// If within 30 days of expiry
	daysLeft := int(exp.Sub(today).Hours() / 24)
	if daysLeft <= 30 {
		return ContractStatusExpiringSoon
	}

	return ContractStatusActive
}

// CalculateContractRenewalStatus determines if contract renewal action is required
func CalculateContractRenewalStatus(expiryDate time.Time, currentRenewalStatus string) string {
	if currentRenewalStatus == RenewalStatusRenewed || currentRenewalStatus == RenewalStatusNotRenewing {
		return currentRenewalStatus
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	exp := time.Date(expiryDate.Year(), expiryDate.Month(), expiryDate.Day(), 0, 0, 0, 0, time.UTC)

	daysLeft := int(exp.Sub(today).Hours() / 24)

	// If expired or within 30 days and renewal has not started
	if daysLeft <= 30 && currentRenewalStatus == "" {
		return RenewalStatusNotStarted
	}
	if currentRenewalStatus == "" {
		return RenewalStatusNotStarted
	}
	return currentRenewalStatus
}

// CanModifyRateDirectly determines whether a rate can be updated in-place without versioning
func CanModifyRateDirectly(isReferenced bool, status, versionStatus string) bool {
	// If rate has been referenced by downstream quotations or bookings, NEVER modify directly
	if isReferenced {
		return false
	}
	// If rate is superseded or archived, cannot modify directly
	if versionStatus == VersionStatusSuperseded || status == RateStatusArchived {
		return false
	}
	// Draft or unreferenced rates can be modified directly
	return true
}
