package customers

import (
	"fmt"
	"time"
)

// Customer Health Status Constants
const (
	HealthStatusHealthy          = "HEALTHY"
	HealthStatusWatch            = "WATCH"
	HealthStatusAtRisk           = "AT_RISK"
	HealthStatusCritical         = "CRITICAL"
	HealthStatusInsufficientData = "INSUFFICIENT_DATA"
)

// Risk Severity Constants
const (
	SeverityInfo      = "INFO"
	SeverityAttention = "ATTENTION"
	SeverityWarning   = "WARNING"
	SeverityCritical  = "CRITICAL"
)

// EvaluateCustomerHealth computes a deterministic health status, score (0-100), and contributing factors
func EvaluateCustomerHealth(kpis Customer360KPIs, fin CustomerFinancialProfile, rfqCount int, quoteCount int, bookingCount int, shipmentCount int, contractCount int) (string, int, []string) {
	totalActivity := rfqCount + quoteCount + bookingCount + shipmentCount + contractCount
	if totalActivity == 0 {
		return HealthStatusInsufficientData, 50, []string{"Insufficient historical operational or commercial activity recorded."}
	}

	score := 50
	factors := []string{}

	if rfqCount > 0 {
		score += 10
		factors = append(factors, fmt.Sprintf("Active inquiry pipeline (%d total RFQs)", rfqCount))
	} else {
		score -= 10
		factors = append(factors, "No active RFQ inquiries recorded")
	}

	if bookingCount > 0 || shipmentCount > 0 {
		score += 15
		factors = append(factors, fmt.Sprintf("Active freight operations (%d bookings, %d shipments)", bookingCount, shipmentCount))
	} else {
		score -= 10
		factors = append(factors, "No current active freight shipments or bookings")
	}

	if contractCount > 0 {
		score += 15
		factors = append(factors, fmt.Sprintf("%d active commercial contract agreement(s)", contractCount))
	}

	switch fin.CreditStatus {
	case CreditStatusGoodStanding:
		score += 10
		factors = append(factors, "Financial credit standing is Good")
	case CreditStatusReviewRequired:
		score -= 15
		factors = append(factors, "Credit status requires review")
	case CreditStatusOnHold:
		score -= 30
		factors = append(factors, "Commercial credit is currently On Hold")
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	var status string
	if score >= 80 {
		status = HealthStatusHealthy
	} else if score >= 60 {
		status = HealthStatusWatch
	} else if score >= 30 {
		status = HealthStatusAtRisk
	} else {
		status = HealthStatusCritical
	}

	return status, score, factors
}

// DetectCustomerRisks evaluates data signals to identify actionable customer risk events
func DetectCustomerRisks(cust Customer, kpis Customer360KPIs, fin CustomerFinancialProfile, rfqCount int, bookingCount int, shipmentCount int, contractCount int, hasPrimaryContact bool, hasOwner bool) []CustomerRiskEvent {
	var risks []CustomerRiskEvent
	now := time.Now()

	if fin.CreditStatus == CreditStatusOnHold {
		risks = append(risks, CustomerRiskEvent{
			OrgID:       cust.OrgID,
			CustomerID:  cust.ID,
			RiskType:    "CUSTOMER_ON_HOLD",
			Severity:    SeverityCritical,
			Title:       "Commercial Account On Credit Hold",
			Description: "Customer credit status is set to On Hold. New bookings require credit committee approval.",
			DetectedAt:  now,
			IsResolved:  false,
		})
	}

	if fin.CreditStatus == CreditStatusReviewRequired {
		risks = append(risks, CustomerRiskEvent{
			OrgID:       cust.OrgID,
			CustomerID:  cust.ID,
			RiskType:    "CREDIT_REVIEW_REQUIRED",
			Severity:    SeverityAttention,
			Title:       "Credit Limit Review Required",
			Description: "Account credit standing requires commercial evaluation.",
			DetectedAt:  now,
			IsResolved:  false,
		})
	}

	if !hasOwner {
		risks = append(risks, CustomerRiskEvent{
			OrgID:       cust.OrgID,
			CustomerID:  cust.ID,
			RiskType:    "NO_ACCOUNT_OWNER",
			Severity:    SeverityAttention,
			Title:       "Unassigned Account Owner",
			Description: "No primary commercial owner assigned to manage this customer account.",
			DetectedAt:  now,
			IsResolved:  false,
		})
	}

	if !hasPrimaryContact {
		risks = append(risks, CustomerRiskEvent{
			OrgID:       cust.OrgID,
			CustomerID:  cust.ID,
			RiskType:    "NO_PRIMARY_CONTACT",
			Severity:    SeverityAttention,
			Title:       "Missing Primary Contact",
			Description: "No key primary contact designated under relationship directory.",
			DetectedAt:  now,
			IsResolved:  false,
		})
	}

	if rfqCount == 0 && bookingCount == 0 && shipmentCount == 0 {
		risks = append(risks, CustomerRiskEvent{
			OrgID:       cust.OrgID,
			CustomerID:  cust.ID,
			RiskType:    "CUSTOMER_INACTIVE",
			Severity:    SeverityWarning,
			Title:       "Inactive Customer Account",
			Description: "No recent RFQ inquiries or active shipments recorded.",
			DetectedAt:  now,
			IsResolved:  false,
		})
	}

	return risks
}

// DetectCommercialOpportunities identifies proactive revenue & relationship growth opportunities
func DetectCommercialOpportunities(cust Customer, kpis Customer360KPIs, rfqCount int, openQuoteCount int, activeContractCount int) []CustomerOpportunityEvent {
	var opps []CustomerOpportunityEvent
	now := time.Now()

	if activeContractCount > 0 {
		opps = append(opps, CustomerOpportunityEvent{
			OrgID:           cust.OrgID,
			CustomerID:      cust.ID,
			OpportunityType: "CONTRACT_RENEWAL_OPPORTUNITY",
			Priority:        "HIGH",
			Title:           "Contract Renewal & Volume Expansion",
			Reason:          "Active master agreement in place with steady volume demand.",
			SuggestedAction: "Initiate commercial review for rate contract extension.",
			DetectedAt:      now,
		})
	}

	if openQuoteCount > 0 {
		opps = append(opps, CustomerOpportunityEvent{
			OrgID:           cust.OrgID,
			CustomerID:      cust.ID,
			OpportunityType: "QUOTATION_FOLLOWUP_OPPORTUNITY",
			Priority:        "MEDIUM",
			Title:           "Open Quotation Follow-up",
			Reason:          "Commercial proposal submitted and awaiting customer decision.",
			SuggestedAction: "Follow up with primary commercial contact regarding rate proposal.",
			DetectedAt:      now,
		})
	}

	if rfqCount == 0 {
		opps = append(opps, CustomerOpportunityEvent{
			OrgID:           cust.OrgID,
			CustomerID:      cust.ID,
			OpportunityType: "REENGAGEMENT_OPPORTUNITY",
			Priority:        "MEDIUM",
			Title:           "Account Re-engagement Opportunity",
			Reason:          "Customer has had no recent freight inquiries.",
			SuggestedAction: "Schedule a commercial check-in call with logistics manager.",
			DetectedAt:      now,
		})
	}

	return opps
}

// EvaluateCustomerActivityTrend determines historical operational activity direction
func EvaluateCustomerActivityTrend(rfqCount int, bookingCount int, shipmentCount int) string {
	totalOps := bookingCount + shipmentCount
	if totalOps >= 3 {
		return "INCREASING"
	}
	if totalOps >= 1 {
		return "STABLE"
	}
	if rfqCount > 0 {
		return "DECLINING"
	}
	if rfqCount == 0 && totalOps == 0 {
		return "INACTIVE"
	}
	return "INSUFFICIENT_DATA"
}
