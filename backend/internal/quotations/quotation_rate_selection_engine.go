package quotations

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ValidateQuotationForRateSelection verifies that a quotation allows commercial editing.
func ValidateQuotationForRateSelection(q *Quotation) error {
	if q == nil {
		return fmt.Errorf("quotation is nil")
	}
	if !EditableStatuses[q.Status] {
		return fmt.Errorf("quotation in status '%s' cannot be modified; rate selection is locked", q.Status)
	}
	return nil
}

// EvaluateCandidateRecommendations flags Cheapest, Fastest, and Best Value candidates.
func EvaluateCandidateRecommendations(candidates []QuotationRateCandidate, targetCurrency string) []QuotationRateCandidate {
	if len(candidates) == 0 {
		return candidates
	}

	var minPrice float64 = -1
	var minPriceIdx int = -1

	var minTransit int = -1
	var minTransitIdx int = -1

	var bestScore float64 = -1
	var bestScoreIdx int = -1

	for i, c := range candidates {
		// Evaluate cheapest (filtering to matching currency if targetCurrency is set)
		if targetCurrency == "" || strings.EqualFold(c.Currency, targetCurrency) {
			if minPrice < 0 || c.CommercialTotal < minPrice {
				minPrice = c.CommercialTotal
				minPriceIdx = i
			}
		}

		// Evaluate fastest
		if c.TransitDays > 0 {
			if minTransit < 0 || c.TransitDays < minTransit {
				minTransit = c.TransitDays
				minTransitIdx = i
			}
		}

		// Calculate composite best value score (price + transit)
		score := 100000.0 - c.CommercialTotal - float64(c.TransitDays*50)
		if c.FreeDaysDestination > 0 {
			score += float64(c.FreeDaysDestination * 10)
		}

		if bestScoreIdx < 0 || score > bestScore {
			bestScore = score
			bestScoreIdx = i
		}
	}

	res := make([]QuotationRateCandidate, len(candidates))
	copy(res, candidates)

	if minPriceIdx >= 0 {
		res[minPriceIdx].RecommendationTags = append(res[minPriceIdx].RecommendationTags, "CHEAPEST")
	}
	if minTransitIdx >= 0 {
		res[minTransitIdx].RecommendationTags = append(res[minTransitIdx].RecommendationTags, "FASTEST")
	}
	if bestScoreIdx >= 0 && (len(candidates) > 1 || minPriceIdx < 0) {
		res[bestScoreIdx].RecommendationTags = append(res[bestScoreIdx].RecommendationTags, "BEST_VALUE")
	}

	return res
}

// DetectCandidateRiskWarnings generates warnings for upcoming expiries or superseded rates.
func DetectCandidateRiskWarnings(c *QuotationRateCandidate, now time.Time) []string {
	var warnings []string

	if c.ValidUntil != nil {
		if c.ValidUntil.Before(now) {
			warnings = append(warnings, "RATE_EXPIRED")
		} else if c.ValidUntil.Sub(now).Hours() < 120 { // within 5 days
			warnings = append(warnings, fmt.Sprintf("Expires soon (%d days)", int(c.ValidUntil.Sub(now).Hours()/24)))
		}
	}

	if strings.EqualFold(c.Status, "SUPERSEDED") {
		warnings = append(warnings, "Rate has been superseded by a newer version")
	} else if strings.EqualFold(c.Status, "EXPIRED") {
		warnings = append(warnings, "Rate contract is expired")
	}

	return warnings
}

// BuildQuotationRateSnapshot constructs an immutable commercial snapshot from a candidate and charge lines.
func BuildQuotationRateSnapshot(
	orgID int64,
	quotationID int64,
	selectionID int64,
	candidate *QuotationRateCandidate,
	user string,
) (*QuotationRateSnapshot, error) {
	if candidate == nil {
		return nil, fmt.Errorf("candidate is nil")
	}

	pricingJSON, err := json.Marshal(map[string]interface{}{
		"source_type":       candidate.SourceType,
		"carrier_name":      candidate.CarrierName,
		"carrier_code":      candidate.CarrierCode,
		"rate_type":         candidate.RateType,
		"rate_version":      candidate.RateVersion,
		"contract_id":       candidate.ContractID,
		"contract_code":     candidate.ContractCode,
		"origin":            candidate.Origin,
		"destination":       candidate.Destination,
		"equipment_type":    candidate.EquipmentType,
		"currency":          candidate.Currency,
		"base_rate":         candidate.BaseRate,
		"total_charges":     candidate.TotalCharges,
		"commercial_total":  candidate.CommercialTotal,
		"transit_days":      candidate.TransitDays,
		"free_days_orig":    candidate.FreeDaysOrigin,
		"free_days_dest":    candidate.FreeDaysDestination,
		"itemized_charges": candidate.Charges,
		"snapshot_at":       time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pricing snapshot: %w", err)
	}

	snap := &QuotationRateSnapshot{
		OrgID:                    orgID,
		QuotationID:              quotationID,
		QuotationRateSelectionID: selectionID,
		SourceRateID:             candidate.RateID,

		CarrierName:       candidate.CarrierName,
		CarrierReference:  candidate.CarrierCode,
		TransportMode:     candidate.TransportMode,
		ServiceType:       candidate.ServiceType,
		EquipmentType:     candidate.EquipmentType,
		Origin:            candidate.Origin,
		Destination:       candidate.Destination,
		Currency:          candidate.Currency,
		BaseRate:          candidate.BaseRate,
		AdditionalCharges: candidate.TotalCharges,
		CommercialTotal:   candidate.CommercialTotal,
		PricingSnapshotJSON: string(pricingJSON),
		ValidFrom:         candidate.ValidFrom,
		ValidUntil:        candidate.ValidUntil,
		SnapshotCreatedAt: time.Now(),
		CreatedBy:         user,
	}

	if candidate.RateVersion > 0 {
		v := candidate.RateVersion
		snap.SourceRateVersion = &v
	}
	if candidate.ContractID != nil {
		snap.SourceContractID = candidate.ContractID
	}
	if candidate.SpotRateRequestID != nil {
		snap.SourceSpotRateRequestID = candidate.SpotRateRequestID
	}
	if candidate.SpotRateResponseID != nil {
		snap.SourceSpotRateResponseID = candidate.SpotRateResponseID
	}

	return snap, nil
}
