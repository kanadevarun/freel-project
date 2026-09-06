package rates

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/freel/backend/internal/rates/spec"
)

// CalculateSpotRateRequestStatus evaluates the authoritative lifecycle status for a spot request
func CalculateSpotRateRequestStatus(req *spec.SpotRateRequest, responses []*spec.SpotRateResponse, now time.Time) string {
	if req.Status == SpotRequestCancelled {
		return SpotRequestCancelled
	}

	// Check if any response is marked preferred / selected
	for _, r := range responses {
		if r.IsPreferred || r.Status == SpotResponseSelected {
			return SpotRequestSelected
		}
	}

	reqDate, err := time.Parse("2006-01-02", req.RequiredByDate)
	if err == nil && now.After(reqDate.Add(24*time.Hour)) {
		return SpotRequestExpired
	}

	validResponseCount := 0
	for _, r := range responses {
		if r.Status == SpotResponseReceived || r.Status == SpotResponseSelected {
			validResponseCount++
		}
	}

	if validResponseCount == 0 {
		if req.Status == SpotRequestDraft {
			return SpotRequestDraft
		}
		return SpotRequestSent
	}

	if validResponseCount == 1 {
		return SpotRequestPartiallyResponded
	}

	return SpotRequestResponded
}

// CalculateSpotRateResponseStatus evaluates status and validity for a carrier response
func CalculateSpotRateResponseStatus(resp *spec.SpotRateResponse, now time.Time) string {
	if resp.Status == SpotResponseDeclined {
		return SpotResponseDeclined
	}
	if resp.IsPreferred || resp.Status == SpotResponseSelected {
		return SpotResponseSelected
	}

	validUntil, err := time.Parse("2006-01-02", resp.ValidUntil)
	if err == nil && now.After(validUntil.Add(24*time.Hour)) {
		return SpotResponseExpired
	}

	return SpotResponseReceived
}

// CalculateSpotRateComparison evaluates side-by-side comparison metrics from persisted carrier responses
func CalculateSpotRateComparison(req *spec.SpotRateRequest, responses []*spec.SpotRateResponse) spec.SpotRateComparison {
	comp := spec.SpotRateComparison{
		RequestID:           req.ID,
		RequestReference:    req.RequestReference,
		Lane:                fmt.Sprintf("%s → %s", req.OriginPort, req.DestinationPort),
		TransportMode:       req.TransportMode,
		EquipmentType:       "40GP",
		TargetCurrency:      req.TargetCurrency,
		Responses:           make([]spec.SpotRateComparisonItem, 0, len(responses)),
		TotalResponsesCount: len(responses),
	}
	if req.EquipmentType != nil && *req.EquipmentType != "" {
		comp.EquipmentType = *req.EquipmentType
	}

	if len(responses) == 0 {
		comp.ComparisonSummaryNote = "No carrier spot responses have been recorded yet for this lane request."
		return comp
	}

	// 1. Check for multi-currency
	currenciesMap := make(map[string]bool)
	var activeResponses []*spec.SpotRateResponse
	for _, r := range responses {
		if r.Status != SpotResponseDeclined {
			currenciesMap[r.Currency] = true
			activeResponses = append(activeResponses, r)
		}
	}

	if len(currenciesMap) > 1 {
		comp.IsMultiCurrency = true
		for c := range currenciesMap {
			comp.MultiCurrencyCurrencies = append(comp.MultiCurrencyCurrencies, c)
		}
		sort.Strings(comp.MultiCurrencyCurrencies)
	}

	// 2. Identify Cheapest, Fastest, Best Value among comparable active responses
	var cheapestResp *spec.SpotRateResponse
	var fastestResp *spec.SpotRateResponse
	var bestValueResp *spec.SpotRateResponse
	var preferredResp *spec.SpotRateResponse

	lowestPrice := -1.0
	lowestTransit := 999999
	highestValueScore := -1.0

	// If multi-currency, only compare within target currency or strictly compare transit
	for _, r := range activeResponses {
		if r.IsPreferred {
			preferredResp = r
		}

		// Transit comparison is safe across currencies
		if r.TransitDays != nil && *r.TransitDays > 0 {
			if *r.TransitDays < lowestTransit {
				lowestTransit = *r.TransitDays
				fastestResp = r
			}
		}

		// Price comparison is safe only if single currency OR matching target currency
		if !comp.IsMultiCurrency || strings.EqualFold(r.Currency, req.TargetCurrency) {
			if lowestPrice < 0 || r.TotalAmount < lowestPrice {
				lowestPrice = r.TotalAmount
				cheapestResp = r
			}
		}
	}

	// Best value computation (Price score 60% + Transit score 30% + Free Days score 10%)
	for _, r := range activeResponses {
		score := 50.0 // baseline
		if lowestPrice > 0 && r.TotalAmount > 0 && (!comp.IsMultiCurrency || strings.EqualFold(r.Currency, req.TargetCurrency)) {
			priceRatio := lowestPrice / r.TotalAmount
			if priceRatio > 1.0 {
				priceRatio = 1.0
			}
			score += priceRatio * 35.0
		}
		if lowestTransit < 999999 && r.TransitDays != nil && *r.TransitDays > 0 {
			transitRatio := float64(lowestTransit) / float64(*r.TransitDays)
			if transitRatio > 1.0 {
				transitRatio = 1.0
			}
			score += transitRatio * 15.0
		}
		// Destination free days bonus (up to 10 pts)
		if r.FreeDaysDestination > 0 {
			bonus := float64(r.FreeDaysDestination)
			if bonus > 10.0 {
				bonus = 10.0
			}
			score += bonus
		}

		if score > highestValueScore {
			highestValueScore = score
			bestValueResp = r
		}
	}

	// 3. Build Comparison Items
	for _, r := range responses {
		var tags []string
		if r.IsPreferred {
			tags = append(tags, RecommendationPreferred)
		}
		if cheapestResp != nil && r.ID == cheapestResp.ID {
			tags = append(tags, RecommendationCheapest)
		}
		if fastestResp != nil && r.ID == fastestResp.ID {
			tags = append(tags, RecommendationFastest)
		}
		if bestValueResp != nil && r.ID == bestValueResp.ID && (cheapestResp == nil || r.ID != cheapestResp.ID || fastestResp == nil || r.ID != fastestResp.ID) {
			tags = append(tags, RecommendationBestValue)
		}

		item := spec.SpotRateComparisonItem{
			ResponseID:            r.ID,
			CarrierName:           r.CarrierName,
			CarrierCode:           r.CarrierCode,
			SupplierName:          r.SupplierName,
			Currency:              r.Currency,
			BaseAmount:            r.BaseAmount,
			TotalCommercialAmount: r.TotalAmount,
			TransitDays:           r.TransitDays,
			FreeDaysDestination:   r.FreeDaysDestination,
			ValidUntil:            r.ValidUntil,
			Status:                r.Status,
			IsPreferred:           r.IsPreferred,
			ChargeCount:           len(r.Charges),
			RecommendationTags:    tags,
			ValueScore:            0.0,
		}
		comp.Responses = append(comp.Responses, item)
	}

	if cheapestResp != nil {
		comp.CheapestResponseID = &cheapestResp.ID
		comp.CheapestCarrierName = &cheapestResp.CarrierName
		comp.CheapestAmount = &cheapestResp.TotalAmount
	}
	if fastestResp != nil {
		comp.FastestResponseID = &fastestResp.ID
		comp.FastestCarrierName = &fastestResp.CarrierName
		comp.FastestTransitDays = fastestResp.TransitDays
	}
	if bestValueResp != nil {
		comp.BestValueResponseID = &bestValueResp.ID
		comp.BestValueCarrierName = &bestValueResp.CarrierName
	}
	if preferredResp != nil {
		comp.PreferredResponseID = &preferredResp.ID
		comp.PreferredCarrierName = &preferredResp.CarrierName
	}

	if comp.IsMultiCurrency {
		comp.ComparisonSummaryNote = fmt.Sprintf("Notice: Multiple currencies (%s) detected across carrier quotes. Price comparison reflects quotes in target currency %s.", strings.Join(comp.MultiCurrencyCurrencies, ", "), req.TargetCurrency)
	} else if cheapestResp != nil {
		transitStr := ""
		if cheapestResp.TransitDays != nil {
			transitStr = fmt.Sprintf(" (%d days transit)", *cheapestResp.TransitDays)
		}
		comp.ComparisonSummaryNote = fmt.Sprintf("%s offers the lowest rate at %s %.2f%s.", cheapestResp.CarrierName, cheapestResp.Currency, cheapestResp.TotalAmount, transitStr)
	}

	return comp
}
