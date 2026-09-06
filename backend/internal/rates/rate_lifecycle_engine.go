package rates

import (
	"fmt"
	"math"
	"time"
)

// EvaluateRateLifecycleStatus computes the current lifecycle status and days remaining for a carrier rate.
func EvaluateRateLifecycleStatus(status string, validFrom, validUntil *time.Time, now time.Time) (string, int) {
	if status == RateStatusArchived {
		return RateStatusArchived, 0
	}
	if status == "SUPERSEDED" {
		return "SUPERSEDED", 0
	}

	if validUntil == nil {
		if status == "" {
			return RateStatusActive, 9999
		}
		return status, 9999
	}

	diff := validUntil.Sub(now)
	daysRemaining := int(math.Ceil(diff.Hours() / 24.0))

	if daysRemaining < 0 {
		return RateStatusExpired, daysRemaining
	}
	if daysRemaining <= 30 {
		return RateStatusExpiringSoon, daysRemaining
	}

	if validFrom != nil && now.Before(*validFrom) {
		return RateStatusDraft, daysRemaining
	}

	return RateStatusActive, daysRemaining
}

// EvaluateContractLifecycleStatus computes lifecycle status and days remaining for a carrier contract.
func EvaluateContractLifecycleStatus(status, renewalStatus string, endDate *time.Time, now time.Time) (string, string, int) {
	if status == ContractStatusArchived {
		return ContractStatusArchived, renewalStatus, 0
	}

	if endDate == nil {
		if status == "" {
			return ContractStatusActive, renewalStatus, 9999
		}
		return status, renewalStatus, 9999
	}

	diff := endDate.Sub(now)
	daysRemaining := int(math.Ceil(diff.Hours() / 24.0))

	if daysRemaining < 0 {
		return ContractStatusExpired, RenewalStatusInProgress, daysRemaining
	}
	if daysRemaining <= 30 {
		if renewalStatus == RenewalStatusNotStarted || renewalStatus == "" {
			renewalStatus = RenewalStatusInProgress
		}
		return ContractStatusExpiringSoon, renewalStatus, daysRemaining
	}

	return ContractStatusActive, renewalStatus, daysRemaining
}

// RateRiskEvaluationInput holds inputs for evaluating quotation risks.
type RateRiskEvaluationInput struct {
	QuotationID              int64
	QuotationNumber          string
	QuotationStatus          string
	SnapshotCarrierName      string
	SnapshotCurrency         string
	SnapshotCommercialTotal  float64
	SnapshotValidUntil       *time.Time
	SourceRateID             *int64
	SourceRateStatus         string
	SourceRateValidUntil     *time.Time
	SourceRateVersion        int
	LatestRateVersion        int
	SourceContractID         *int64
	SourceContractCode       string
	SourceContractStatus     string
	SourceContractEndDate    *time.Time
	SourceSpotRateResponseID *int64
	SourceSpotValidUntil     *time.Time
	SourceSpotStatus         string
}

// QuotationRiskDetected represents a detected risk event to be persisted.
type QuotationRiskDetected struct {
	RiskType          string
	Severity          string
	Headline          string
	Description       string
	RecommendedAction string
}

// DetectQuotationRateRisks evaluates whether a quotation's commercial snapshot faces external lifecycle risks.
func DetectQuotationRateRisks(input RateRiskEvaluationInput, now time.Time) []QuotationRiskDetected {
	var risks []QuotationRiskDetected

	// 1. Source Managed Rate Expiry / Expiring Soon
	if input.SourceRateID != nil {
		if input.SourceRateValidUntil != nil {
			diff := input.SourceRateValidUntil.Sub(now)
			days := int(math.Ceil(diff.Hours() / 24.0))

			if days < 0 || input.SourceRateStatus == RateStatusExpired {
				severity := SeverityWarning
				if input.QuotationStatus == "DRAFT" || input.QuotationStatus == "CHANGES_REQUESTED" {
					severity = SeverityCritical
				}
				risks = append(risks, QuotationRiskDetected{
					RiskType:          "RATE_EXPIRED",
					Severity:          severity,
					Headline:          fmt.Sprintf("Carrier rate from %s has expired", input.SnapshotCarrierName),
					Description:       fmt.Sprintf("The underlying carrier rate expired on %s. The quotation snapshot retains its pricing, but rate replacement is recommended before quote submission.", input.SourceRateValidUntil.Format("2006-01-02")),
					RecommendedAction: "Review replacement rates and update quotation before sending to client.",
				})
			} else if days <= 30 || input.SourceRateStatus == RateStatusExpiringSoon {
				risks = append(risks, QuotationRiskDetected{
					RiskType:          "RATE_EXPIRING_SOON",
					Severity:          SeverityWarning,
					Headline:          fmt.Sprintf("Carrier rate from %s expires in %d days", input.SnapshotCarrierName, days),
					Description:       fmt.Sprintf("The underlying carrier rate will expire on %s (%d days remaining).", input.SourceRateValidUntil.Format("2006-01-02"), days),
					RecommendedAction: "Confirm carrier validity window or check for newly published contract rate revisions.",
				})
			}
		}

		// Check for supersession / newer version
		if input.LatestRateVersion > input.SourceRateVersion || input.SourceRateStatus == "SUPERSEDED" {
			risks = append(risks, QuotationRiskDetected{
				RiskType:          "RATE_SUPERSEDED",
				Severity:          SeverityInfo,
				Headline:          fmt.Sprintf("Newer rate revision v%d available for %s", input.LatestRateVersion, input.SnapshotCarrierName),
				Description:       fmt.Sprintf("This quotation references rate version v%d, but version v%d is now published.", input.SourceRateVersion, input.LatestRateVersion),
				RecommendedAction: "Compare new version pricing against the current snapshot to evaluate commercial impact.",
			})
		}
	}

	// 2. Source Contract Expiry
	if input.SourceContractID != nil && input.SourceContractEndDate != nil {
		diff := input.SourceContractEndDate.Sub(now)
		days := int(math.Ceil(diff.Hours() / 24.0))

		if days < 0 || input.SourceContractStatus == ContractStatusExpired {
			risks = append(risks, QuotationRiskDetected{
				RiskType:          "CONTRACT_EXPIRED",
				Severity:          SeverityWarning,
				Headline:          fmt.Sprintf("Carrier contract %s has expired", input.SourceContractCode),
				Description:       fmt.Sprintf("Carrier contract %s expired on %s.", input.SourceContractCode, input.SourceContractEndDate.Format("2006-01-02")),
				RecommendedAction: "Verify carrier contract renewal status or attach an active spot quote.",
			})
		} else if days <= 30 || input.SourceContractStatus == ContractStatusExpiringSoon {
			risks = append(risks, QuotationRiskDetected{
				RiskType:          "CONTRACT_EXPIRING_SOON",
				Severity:          SeverityInfo,
				Headline:          fmt.Sprintf("Carrier contract %s expires in %d days", input.SourceContractCode, days),
				Description:       fmt.Sprintf("Carrier contract %s expires on %s (%d days remaining).", input.SourceContractCode, input.SourceContractEndDate.Format("2006-01-02"), days),
				RecommendedAction: "Track contract renewal progress.",
			})
		}
	}

	// 3. Source Spot Rate Response Expiry
	if input.SourceSpotRateResponseID != nil && input.SourceSpotValidUntil != nil {
		diff := input.SourceSpotValidUntil.Sub(now)
		days := int(math.Ceil(diff.Hours() / 24.0))

		if days < 0 || input.SourceSpotStatus == "EXPIRED" {
			severity := SeverityWarning
			if input.QuotationStatus == "DRAFT" || input.QuotationStatus == "CHANGES_REQUESTED" {
				severity = SeverityCritical
			}
			risks = append(risks, QuotationRiskDetected{
				RiskType:          "SPOT_RATE_EXPIRED",
				Severity:          severity,
				Headline:          fmt.Sprintf("Spot rate quote from %s has expired", input.SnapshotCarrierName),
				Description:       fmt.Sprintf("The carrier spot response expired on %s.", input.SourceSpotValidUntil.Format("2006-01-02")),
				RecommendedAction: "Request a refreshed spot quote or select a current contract rate.",
			})
		} else if days <= 7 {
			risks = append(risks, QuotationRiskDetected{
				RiskType:          "SPOT_RATE_EXPIRING",
				Severity:          SeverityWarning,
				Headline:          fmt.Sprintf("Spot rate quote from %s expires in %d days", input.SnapshotCarrierName, days),
				Description:       fmt.Sprintf("Spot response validity ends on %s (%d days remaining).", input.SourceSpotValidUntil.Format("2006-01-02"), days),
				RecommendedAction: "Expedite customer confirmation before spot pricing expires.",
			})
		}
	}

	return risks
}
