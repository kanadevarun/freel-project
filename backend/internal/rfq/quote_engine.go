package rfq

import (
	"fmt"
	"math"
	"time"

	"github.com/freel/backend/internal/rfq/spec"
)

// CalculateQuoteMargin safely computes the absolute margin and margin percentage.
// It avoids division by zero or NaN if SellPrice <= 0.
func CalculateQuoteMargin(buyPrice, sellPrice float64) (float64, float64) {
	marginAmount := sellPrice - buyPrice
	if sellPrice <= 0 {
		return marginAmount, 0.0
	}
	marginPercentage := (marginAmount / sellPrice) * 100.0
	// Round to 1 decimal place
	marginPercentage = math.Round(marginPercentage*10) / 10
	marginAmount = math.Round(marginAmount*100) / 100
	return marginAmount, marginPercentage
}

// BuildQuotesResponse executes deterministic multi-carrier comparison, margin analysis,
// validity calculation, and recommendation synthesis across persistent quote records.
func BuildQuotesResponse(
	rfq *spec.RFQ,
	quotes []spec.RFQQuote,
	requirements *spec.GetRequirementsResponse,
	now time.Time,
) *spec.GetQuotesResponse {
	if rfq == nil {
		rfq = &spec.RFQ{}
	}

	resp := &spec.GetQuotesResponse{
		Quotes:     make([]spec.RFQQuote, 0),
		Comparison: make([]spec.QuoteComparison, 0),
		Summary: spec.QuoteSummary{
			PrimaryCurrency: "USD",
		},
	}

	// Attach Live RFQ Requirements Readiness if available
	if requirements != nil {
		resp.RFQReadiness = requirements.OperationalReadiness
	} else {
		resp.RFQReadiness = spec.OperationalReadiness{
			OverallStatus:  spec.ReadinessReadyForQuotation,
			ReadinessScore: 100,
			NextBestAction: "Proceed to commercial review and quote selection.",
		}
	}

	if len(quotes) == 0 {
		return resp
	}

	// 1. Analyze Currencies & Validity for each quote
	currencyCounts := make(map[string]int)
	for i := range quotes {
		q := &quotes[i]
		q.UnmarshalCharges()

		if q.Currency == "" {
			q.Currency = "USD"
		}
		currencyCounts[q.Currency]++

		// Calculate deterministic margins
		q.MarginAmount, q.MarginPercentage = CalculateQuoteMargin(q.BuyPrice, q.SellPrice)

		// Calculate deterministic validity status
		q.ValidityStatus, q.DaysUntilExpiry = EvaluateQuoteValidity(q.ValidUntil, now)

		// Expiry status reflection: if date has passed and quote is not approved/selected,
		// note validity status without destructive mutation
		if q.ValidityStatus == spec.ValidityExpired && q.Status != spec.QuoteStatusApproved && q.Status != spec.QuoteStatusSelectedForCustomer {
			// runtime awareness
		}
	}

	// Determine primary dominant currency
	maxCount := 0
	primaryCurrency := "USD"
	for curr, count := range currencyCounts {
		if count > maxCount {
			maxCount = count
			primaryCurrency = curr
		}
	}
	resp.Summary.PrimaryCurrency = primaryCurrency
	resp.Summary.HasMixedCurrencies = len(currencyCounts) > 1

	// 2. Identify Comparison Benchmarks across valid quotes (same currency)
	var (
		lowestBuyQuoteID       int64 = 0
		highestMarginQuoteID   int64 = 0
		fastestTransitQuoteID  int64 = 0
		minBuyPrice                  = math.MaxFloat64
		maxMarginAmount              = -math.MaxFloat64
		minTransitDays               = math.MaxInt32
	)

	for _, q := range quotes {
		// Only consider active, non-rejected, non-withdrawn, non-expired quotes with same currency for comparison
		isValidForComparison := q.Status != spec.QuoteStatusRejected &&
			q.Status != spec.QuoteStatusWithdrawn &&
			q.ValidityStatus != spec.ValidityExpired &&
			q.Currency == primaryCurrency &&
			q.BuyPrice > 0

		if isValidForComparison {
			if q.BuyPrice < minBuyPrice {
				minBuyPrice = q.BuyPrice
				lowestBuyQuoteID = q.ID
			}

			if q.MarginAmount > maxMarginAmount {
				maxMarginAmount = q.MarginAmount
				highestMarginQuoteID = q.ID
			}

			if q.TransitTimeDays != nil && *q.TransitTimeDays > 0 && *q.TransitTimeDays < minTransitDays {
				minTransitDays = *q.TransitTimeDays
				fastestTransitQuoteID = q.ID
			}
		}
	}

	if minBuyPrice < math.MaxFloat64 {
		b := minBuyPrice
		resp.Summary.LowestBuyAmount = &b
	}
	if maxMarginAmount > -math.MaxFloat64 {
		m := maxMarginAmount
		resp.Summary.HighestMarginAmount = &m
	}
	if minTransitDays < math.MaxInt32 {
		t := minTransitDays
		resp.Summary.FastestTransitDays = &t
	}

	// 3. Build Summary Counters & Comparison Items
	for i := range quotes {
		q := &quotes[i]

		// Increment summary buckets
		resp.Summary.TotalQuotes++
		switch q.Status {
		case spec.QuoteStatusDraft:
			resp.Summary.DraftQuotes++
		case spec.QuoteStatusRequested:
			resp.Summary.RequestedQuotes++
		case spec.QuoteStatusReceived:
			resp.Summary.ReceivedQuotes++
		case spec.QuoteStatusUnderReview:
			resp.Summary.UnderReviewQuotes++
		case spec.QuoteStatusRecommended:
			resp.Summary.RecommendedQuotes++
		case spec.QuoteStatusApproved:
			resp.Summary.ApprovedQuotes++
		case spec.QuoteStatusSelectedForCustomer:
			resp.Summary.SelectedQuotes++
		}

		if q.ValidityStatus == spec.ValidityExpired {
			resp.Summary.ExpiredQuotes++
		} else if q.ValidityStatus == spec.ValidityExpiringSoon {
			resp.Summary.QuotesExpiringSoon++
		}

		// Track explicit flags from DB
		if q.IsRecommended || q.Status == spec.QuoteStatusRecommended {
			quoteCopy := *q
			resp.RecommendedQuote = &quoteCopy
			resp.Summary.RecommendedQuoteID = &q.ID
		}
		if q.Status == spec.QuoteStatusApproved || q.Status == spec.QuoteStatusSelectedForCustomer {
			quoteCopy := *q
			resp.ApprovedQuote = &quoteCopy
			resp.Summary.ApprovedQuoteID = &q.ID
		}

		// Calculate comparison flags
		isLowestCost := q.ID == lowestBuyQuoteID && lowestBuyQuoteID > 0
		isHighestMargin := q.ID == highestMarginQuoteID && highestMarginQuoteID > 0
		isFastest := q.ID == fastestTransitQuoteID && fastestTransitQuoteID > 0
		isApproved := q.Status == spec.QuoteStatusApproved || q.Status == spec.QuoteStatusSelectedForCustomer
		isRecommended := q.IsRecommended || q.Status == spec.QuoteStatusRecommended

		// Generate explainable recommendation reason
		recommendationReason := generateDeterministicReason(q, isLowestCost, isHighestMargin, isFastest)

		// Base Score (0-100 deterministic)
		score := calculateDeterministicScore(q, isLowestCost, isHighestMargin, isFastest)

		comparisonItem := spec.QuoteComparison{
			QuoteID:              q.ID,
			CarrierName:          q.CarrierName,
			QuoteReference:       q.QuoteReference,
			Status:               q.Status,
			Currency:             q.Currency,
			BuyPrice:             q.BuyPrice,
			SellPrice:            q.SellPrice,
			MarginAmount:         q.MarginAmount,
			MarginPercentage:     q.MarginPercentage,
			TransitTimeDays:      q.TransitTimeDays,
			ValidUntil:           q.ValidUntil,
			ValidityStatus:       q.ValidityStatus,
			IsLowestCost:         isLowestCost,
			IsHighestMargin:      isHighestMargin,
			IsFastest:            isFastest,
			IsRecommended:        isRecommended,
			IsApproved:           isApproved,
			Score:                score,
			RecommendationReason: recommendationReason,
		}

		resp.Comparison = append(resp.Comparison, comparisonItem)
	}

	// 4. If no explicit recommendation exists in DB, identify best deterministic candidate
	if resp.RecommendedQuote == nil && len(quotes) > 0 {
		for i := range quotes {
			q := &quotes[i]
			if CanRecommendQuote(q, now) == nil {
				if q.ID == highestMarginQuoteID || (highestMarginQuoteID == 0 && q.ID == lowestBuyQuoteID) {
					// Found best candidate
					quoteCopy := *q
					resp.RecommendedQuote = &quoteCopy
					resp.Summary.RecommendedQuoteID = &q.ID
					break
				}
			}
		}
	}

	resp.Quotes = quotes
	return resp
}

// generateDeterministicReason builds a transparent explanation from stored quote facts.
func generateDeterministicReason(q *spec.RFQQuote, isLowest, isHighestMargin, isFastest bool) string {
	if q.ValidityStatus == spec.ValidityExpired {
		return "Quote validity expired; not eligible for commercial recommendation."
	}
	if q.Status == spec.QuoteStatusRejected {
		return "Quote was rejected during operational review."
	}
	if q.Status == spec.QuoteStatusWithdrawn {
		return "Quote was withdrawn by carrier/operations."
	}
	if q.BuyPrice <= 0 || q.SellPrice <= 0 {
		return "Incomplete commercial pricing data."
	}

	if isHighestMargin && isFastest {
		return fmt.Sprintf("Top recommendation: Highest commercial margin (%s %.2f / %.1f%%) and fastest transit window.", q.Currency, q.MarginAmount, q.MarginPercentage)
	}
	if isHighestMargin && isLowest {
		return fmt.Sprintf("Optimal commercial value: Highest margin (%.1f%%) and lowest carrier buy cost (%s %.2f).", q.MarginPercentage, q.Currency, q.BuyPrice)
	}
	if isHighestMargin {
		return fmt.Sprintf("Highest commercial margin (%s %.2f / %.1f%%) among valid carrier quotations.", q.Currency, q.MarginAmount, q.MarginPercentage)
	}
	if isLowest {
		return fmt.Sprintf("Lowest buy cost (%s %.2f) among active carrier options.", q.Currency, q.BuyPrice)
	}
	if isFastest {
		return "Fastest transit time among valid quotes with positive commercial margin."
	}

	if q.MarginPercentage > 15.0 {
		return fmt.Sprintf("Strong commercial margin (%.1f%%) meeting operational trade requirements.", q.MarginPercentage)
	}

	return "Commercially viable carrier option with positive margin."
}

// calculateDeterministicScore computes a 0-100 rating based on margin, reliability, and transit.
func calculateDeterministicScore(q *spec.RFQQuote, isLowest, isHighestMargin, isFastest bool) float64 {
	score := 50.0

	// Margin contribution (up to 30 pts)
	if q.MarginPercentage > 25.0 {
		score += 30.0
	} else if q.MarginPercentage > 15.0 {
		score += 20.0
	} else if q.MarginPercentage > 0.0 {
		score += 10.0
	}

	// Relative advantages (up to 20 pts)
	if isLowest {
		score += 10.0
	}
	if isHighestMargin {
		score += 10.0
	}
	if isFastest {
		score += 10.0
	}

	// Deductions
	if q.ValidityStatus == spec.ValidityExpired || q.Status == spec.QuoteStatusRejected {
		score = 0.0
	} else if q.ValidityStatus == spec.ValidityExpiringSoon {
		score -= 10.0
	}

	if score > 100.0 {
		score = 100.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return math.Round(score*10) / 10
}
