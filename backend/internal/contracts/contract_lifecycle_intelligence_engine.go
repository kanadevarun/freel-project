package contracts

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// ExpiryThreshold Constants
const (
	ThresholdExpired0Days = "EXPIRED_0_DAYS"
	ThresholdExpiring7D   = "EXPIRING_7_DAYS"
	ThresholdExpiring30D  = "EXPIRING_30_DAYS"
	ThresholdExpiring60D  = "EXPIRING_60_DAYS"
	ThresholdExpiring90D  = "EXPIRING_90_DAYS"
	ThresholdHealthy90P   = "HEALTHY_90_PLUS"
	ThresholdNoExpiry     = "NO_EXPIRY"
)

// ContractLifecycleIntelligenceEngine handles deterministic lifecycle evaluation.
type ContractLifecycleIntelligenceEngine struct {
	dl DataLayer
}

// NewContractLifecycleIntelligenceEngine creates a new engine instance.
func NewContractLifecycleIntelligenceEngine(dl DataLayer) *ContractLifecycleIntelligenceEngine {
	return &ContractLifecycleIntelligenceEngine{dl: dl}
}

// EvaluationResult contains the computed intelligence for a contract without mutating the contract.
type EvaluationResult struct {
	ContractID           int64
	Condition            LifecycleCondition
	DaysRemaining        int
	ExpiryThreshold      string
	HealthLabel          string
	HealthProgress       int
	GeneratedEvents      []ContractLifecycleIntelligenceEvent
	DetectedRisks        []ContractRiskEvent
	CommercialImpact     ContractCommercialImpactSummary
	RenewalRecommendation string
}

// EvaluateDaysRemaining calculates days remaining and assigns standard threshold.
func EvaluateDaysRemaining(effectiveDateStr, expiryDateStr *string, status ContractStatus) (int, string, LifecycleCondition, int) {
	if status == ContractStatusDraft {
		return 0, ThresholdNoExpiry, LifecycleConditionDraft, 0
	}
	if status == ContractStatusArchived {
		return 0, ThresholdNoExpiry, LifecycleConditionArchived, 100
	}
	if expiryDateStr == nil || *expiryDateStr == "" {
		return 999, ThresholdNoExpiry, LifecycleConditionActive, 10
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	expiryTime, err := time.Parse("2006-01-02", *expiryDateStr)
	if err != nil {
		// Try ISO8601
		expiryTime, err = time.Parse(time.RFC3339, *expiryDateStr)
		if err != nil {
			return 999, ThresholdNoExpiry, LifecycleConditionActive, 10
		}
	}
	expiryTime = expiryTime.UTC().Truncate(24 * time.Hour)

	var effectiveTime time.Time
	if effectiveDateStr != nil && *effectiveDateStr != "" {
		effectiveTime, _ = time.Parse("2006-01-02", *effectiveDateStr)
	}
	if effectiveTime.IsZero() {
		effectiveTime = today.AddDate(-1, 0, 0)
	}

	diffHours := expiryTime.Sub(today).Hours()
	daysRemaining := int(math.Ceil(diffHours / 24.0))

	totalDuration := math.Max(1, expiryTime.Sub(effectiveTime).Hours())
	elapsed := math.Max(0, today.Sub(effectiveTime).Hours())
	progress := int(math.Min(100, math.Max(0, (elapsed/totalDuration)*100)))

	if status == ContractStatusExpired || daysRemaining <= 0 {
		return daysRemaining, ThresholdExpired0Days, LifecycleConditionExpired, 100
	}
	if daysRemaining <= 7 {
		return daysRemaining, ThresholdExpiring7D, LifecycleConditionExpiringSoon, progress
	}
	if daysRemaining <= 30 {
		return daysRemaining, ThresholdExpiring30D, LifecycleConditionExpiringSoon, progress
	}
	if daysRemaining <= 60 {
		return daysRemaining, ThresholdExpiring60D, LifecycleConditionActive, progress
	}
	if daysRemaining <= 90 {
		return daysRemaining, ThresholdExpiring90D, LifecycleConditionActive, progress
	}

	return daysRemaining, ThresholdHealthy90P, LifecycleConditionActive, progress
}

// EvaluateContract deterministically calculates lifecycle condition, risks, and events for a contract.
func (e *ContractLifecycleIntelligenceEngine) EvaluateContract(
	ctx context.Context,
	orgID int64,
	contract Contract,
	renewal *ContractRenewalTracking,
	impact ContractCommercialImpactSummary,
) EvaluationResult {
	days, threshold, baseCondition, progress := EvaluateDaysRemaining(contract.EffectiveDate, contract.ExpiryDate, contract.Status)

	condition := baseCondition
	healthLabel := "Active & Healthy"

	if renewal != nil && renewal.RenewalStatus == RenewalStatusInProgress {
		condition = LifecycleConditionRenewalInProgress
		healthLabel = "Renewal In Progress"
	} else if renewal != nil && renewal.RenewalStatus == RenewalStatusRenewed {
		condition = LifecycleConditionRenewed
		healthLabel = "Renewed Agreement"
	} else if renewal != nil && renewal.SuccessorContractID != nil && *renewal.SuccessorContractID > 0 {
		condition = LifecycleConditionSuperseded
		healthLabel = "Superseded by Successor"
	} else {
		switch baseCondition {
		case LifecycleConditionExpired:
			healthLabel = "Expired Agreement"
		case LifecycleConditionExpiringSoon:
			if days <= 7 {
				healthLabel = fmt.Sprintf("Critical Expiry (%d days left)", days)
			} else {
				healthLabel = fmt.Sprintf("Expiring Soon (%d days left)", days)
			}
		case LifecycleConditionDraft:
			healthLabel = "Draft Pipeline"
		case LifecycleConditionArchived:
			healthLabel = "Archived Record"
		default:
			healthLabel = "Active & In Effect"
		}
	}

	res := EvaluationResult{
		ContractID:       contract.ID,
		Condition:        condition,
		DaysRemaining:    days,
		ExpiryThreshold:  threshold,
		HealthLabel:      healthLabel,
		HealthProgress:   progress,
		CommercialImpact: impact,
		GeneratedEvents:  []ContractLifecycleIntelligenceEvent{},
		DetectedRisks:    []ContractRiskEvent{},
	}

	// 1. Detect Expiry & Renewal Recommendation Risks
	if (condition == LifecycleConditionExpiringSoon || condition == LifecycleConditionExpired) && 
	   (renewal == nil || renewal.RenewalStatus == RenewalStatusNotStarted) {
		
		var severity RiskSeverity = SeverityWarning
		if days <= 7 || condition == LifecycleConditionExpired {
			severity = SeverityCritical
		}

		res.RenewalRecommendation = "Initiate commercial renewal negotiations with counterparty before operational disruption."

		// Active rates risk
		if impact.ActiveRatesCount > 0 {
			res.DetectedRisks = append(res.DetectedRisks, ContractRiskEvent{
				OrgID:       orgID,
				ContractID:  contract.ID,
				RiskType:    RiskExpiringActiveRates,
				Severity:    severity,
				Description: fmt.Sprintf("Contract %s is %s with %d active commercial rates attached.", contract.ContractReference, string(condition), impact.ActiveRatesCount),
				IsResolved:  false,
			})
		}

		// Draft quotes risk
		if impact.DraftQuotationsCount > 0 && condition == LifecycleConditionExpired {
			res.DetectedRisks = append(res.DetectedRisks, ContractRiskEvent{
				OrgID:       orgID,
				ContractID:  contract.ID,
				RiskType:    RiskExpiredDraftQuotes,
				Severity:    SeverityWarning,
				Description: fmt.Sprintf("Contract %s has expired while %d draft quotations reference it.", contract.ContractReference, impact.DraftQuotationsCount),
				IsResolved:  false,
			})
		}

		// Renewal overdue risk
		if condition == LifecycleConditionExpired {
			res.DetectedRisks = append(res.DetectedRisks, ContractRiskEvent{
				OrgID:       orgID,
				ContractID:  contract.ID,
				RiskType:    RiskRenewalOverdue,
				Severity:    SeverityCritical,
				Description: fmt.Sprintf("Contract %s expired %d days ago with no renewal or replacement recorded.", contract.ContractReference, int(math.Abs(float64(days)))),
				IsResolved:  false,
			})
		}
	}

	// 2. Generate Lifecycle Intelligence Events
	if days <= 7 && days > 0 && contract.Status == ContractStatusActive {
		meta, _ := json.Marshal(map[string]interface{}{"days_remaining": days, "threshold": threshold})
		metaStr := string(meta)
		desc := fmt.Sprintf("Contract %s is in critical expiry window with %d days remaining.", contract.ContractReference, days)
		res.GeneratedEvents = append(res.GeneratedEvents, ContractLifecycleIntelligenceEvent{
			OrgID:       orgID,
			ContractID:  contract.ID,
			EventType:   IntelEventCriticalExpiryRisk,
			Severity:    SeverityCritical,
			Description: &desc,
			Metadata:    &metaStr,
		})
	} else if days <= 30 && days > 7 && contract.Status == ContractStatusActive {
		meta, _ := json.Marshal(map[string]interface{}{"days_remaining": days, "threshold": threshold})
		metaStr := string(meta)
		desc := fmt.Sprintf("Contract %s expiry is approaching in %d days.", contract.ContractReference, days)
		res.GeneratedEvents = append(res.GeneratedEvents, ContractLifecycleIntelligenceEvent{
			OrgID:       orgID,
			ContractID:  contract.ID,
			EventType:   IntelEventExpiryApproaching,
			Severity:    SeverityWarning,
			Description: &desc,
			Metadata:    &metaStr,
		})
	}

	return res
}
