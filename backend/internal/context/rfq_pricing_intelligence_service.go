package bcontext

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
)

// GetRFQ360PricingIntelligence generates complete, organization-scoped, read-only RFQ and pricing intelligence.
func (s *defaultService) GetRFQ360PricingIntelligence(ctx context.Context, orgID int64, rfqID int64, correlationID string, userID int64) (*RFQ360PricingIntelligence, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("organization ID is required")
	}
	if rfqID <= 0 {
		return nil, fmt.Errorf("invalid RFQ ID: %d", rfqID)
	}
	if correlationID == "" {
		correlationID = fmt.Sprintf("rfq-intel-%d-%d", rfqID, time.Now().UnixNano())
	}

	// 1. Fetch RFQ with Customer and Assignee metadata
	queryRFQ := `
		SELECT r.id, r.org_id, r.rfq_number, r.customer_id, r.stage,
		       COALESCE(r.origin, '') AS origin,
		       COALESCE(r.destination, '') AS destination,
		       COALESCE(r.incoterms, '') AS incoterms,
		       r.target_date, r.sales_assignee_id, r.pricing_assignee_id, r.lead_id,
		       r.created_at, r.updated_at,
		       COALESCE(c.name, '') AS customer_name,
		       COALESCE(c.customer_code, '') AS customer_code,
		       TRIM(CONCAT(COALESCE(u1.first_name, ''), ' ', COALESCE(u1.last_name, ''))) AS sales_assignee_name,
		       TRIM(CONCAT(COALESCE(u2.first_name, ''), ' ', COALESCE(u2.last_name, ''))) AS pricing_assignee_name
		FROM rfqs r
		LEFT JOIN customers c ON r.customer_id = c.id
		LEFT JOIN users u1 ON r.sales_assignee_id = u1.id
		LEFT JOIN users u2 ON r.pricing_assignee_id = u2.id
		WHERE r.org_id = ? AND r.id = ?
		LIMIT 1
	`
	var rfqRow struct {
		ID                  int64      `db:"id"`
		OrgID               int64      `db:"org_id"`
		RFQNumber           string     `db:"rfq_number"`
		CustomerID          int64      `db:"customer_id"`
		Stage               string     `db:"stage"`
		Origin              string     `db:"origin"`
		Destination         string     `db:"destination"`
		Incoterms           string     `db:"incoterms"`
		TargetDate          *time.Time `db:"target_date"`
		SalesAssigneeID     *int64     `db:"sales_assignee_id"`
		PricingAssigneeID   *int64     `db:"pricing_assignee_id"`
		LeadID              *int64     `db:"lead_id"`
		CreatedAt           time.Time  `db:"created_at"`
		UpdatedAt           time.Time  `db:"updated_at"`
		CustomerName        string     `db:"customer_name"`
		CustomerCode        string     `db:"customer_code"`
		SalesAssigneeName   string     `db:"sales_assignee_name"`
		PricingAssigneeName string     `db:"pricing_assignee_name"`
	}

	if s.db == nil {
		return nil, fmt.Errorf("database handle not configured")
	}

	err := s.db.GetContext(ctx, &rfqRow, queryRFQ, orgID, rfqID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rfq %d not found in organization", rfqID)
		}
		return nil, fmt.Errorf("failed to load RFQ %d: %w", rfqID, err)
	}

	// 2. Fetch Cargo Items from rfq_items
	queryItems := `
		SELECT id, COALESCE(description, '') AS description, quantity,
		       COALESCE(weight_kg, 0.0) AS weight_kg,
		       COALESCE(volume_cbm, 0.0) AS volume_cbm
		FROM rfq_items
		WHERE rfq_id = ?
		ORDER BY id ASC
	`
	var itemRows []struct {
		ID          int64   `db:"id"`
		Description string  `db:"description"`
		Quantity    int     `db:"quantity"`
		WeightKg    float64 `db:"weight_kg"`
		VolumeCbm   float64 `db:"volume_cbm"`
	}
	_ = s.db.SelectContext(ctx, &itemRows, queryItems, rfqID)

	var totalWeight, totalVolume float64
	var cargoDescs []string
	for _, it := range itemRows {
		totalWeight += it.WeightKg
		totalVolume += it.VolumeCbm
		if it.Description != "" {
			cargoDescs = append(cargoDescs, fmt.Sprintf("%dx %s", it.Quantity, it.Description))
		}
	}

	// Calculate Completeness Score
	missingFields := []string{}
	completenessPoints := 0
	totalChecks := 6

	if rfqRow.Origin != "" {
		completenessPoints++
	} else {
		missingFields = append(missingFields, "origin")
	}

	if rfqRow.Destination != "" {
		completenessPoints++
	} else {
		missingFields = append(missingFields, "destination")
	}

	if rfqRow.TargetDate != nil {
		completenessPoints++
	} else {
		missingFields = append(missingFields, "target_date")
	}

	if rfqRow.Incoterms != "" {
		completenessPoints++
	} else {
		missingFields = append(missingFields, "incoterms")
	}

	if len(itemRows) > 0 {
		completenessPoints++
	} else {
		missingFields = append(missingFields, "cargo_items")
	}

	if totalWeight > 0 || totalVolume > 0 {
		completenessPoints++
	} else {
		missingFields = append(missingFields, "cargo_weight_or_volume")
	}

	completenessScore := int((float64(completenessPoints) / float64(totalChecks)) * 100)

	identity := RFQIdentitySummary{
		ID:                    rfqRow.ID,
		OrgID:                 rfqRow.OrgID,
		RFQNumber:             rfqRow.RFQNumber,
		CustomerID:            rfqRow.CustomerID,
		CustomerName:          rfqRow.CustomerName,
		CustomerCode:          rfqRow.CustomerCode,
		Stage:                 rfqRow.Stage,
		Origin:                rfqRow.Origin,
		Destination:           rfqRow.Destination,
		Incoterms:             rfqRow.Incoterms,
		TargetDate:            rfqRow.TargetDate,
		SalesAssigneeID:       rfqRow.SalesAssigneeID,
		SalesAssigneeName:     rfqRow.SalesAssigneeName,
		PricingAssigneeID:     rfqRow.PricingAssigneeID,
		PricingAssigneeName:   rfqRow.PricingAssigneeName,
		LeadID:                rfqRow.LeadID,
		TotalItemsCount:       len(itemRows),
		TotalWeightKg:         totalWeight,
		TotalVolumeCbm:        totalVolume,
		CargoDescription:      strings.Join(cargoDescs, ", "),
		Currency:              "USD",
		CompletenessScore:     completenessScore,
		IsComplete:            len(missingFields) == 0,
		MissingRequiredFields: missingFields,
		CreatedAt:             rfqRow.CreatedAt,
		UpdatedAt:             rfqRow.UpdatedAt,
	}

	// 3. Fetch Quotations from `quotations` table
	queryCommercialQuotes := `
		SELECT id, quotation_number, status, COALESCE(currency, 'USD') AS currency,
		       COALESCE(total_amount, 0.0) AS total_amount,
		       COALESCE(total_cost, 0.0) AS total_cost,
		       COALESCE(gross_profit, 0.0) AS gross_profit,
		       COALESCE(gross_margin_pct, 0.0) AS gross_margin_pct,
		       valid_from, valid_until, created_at
		FROM quotations
		WHERE org_id = ? AND rfq_id = ?
		ORDER BY total_amount ASC
	`
	var commQuoteRows []struct {
		ID              int64      `db:"id"`
		QuotationNumber string     `db:"quotation_number"`
		Status          string     `db:"status"`
		Currency        string     `db:"currency"`
		TotalAmount     float64    `db:"total_amount"`
		TotalCost       float64    `db:"total_cost"`
		GrossProfit     float64    `db:"gross_profit"`
		GrossMarginPct  float64    `db:"gross_margin_pct"`
		ValidFrom       *time.Time `db:"valid_from"`
		ValidUntil      *time.Time `db:"valid_until"`
		CreatedAt       time.Time  `db:"created_at"`
	}
	_ = s.db.SelectContext(ctx, &commQuoteRows, queryCommercialQuotes, orgID, rfqID)

	// 4. Fetch Carrier Quotes from `rfq_quotes` table
	queryCarrierQuotes := `
		SELECT id, carrier_name, COALESCE(transit_time_days, 0) AS transit_time_days,
		       COALESCE(buy_price, 0.0) AS buy_price,
		       COALESCE(sell_price, 0.0) AS sell_price,
		       COALESCE(is_recommended, false) AS is_recommended,
		       COALESCE(status, 'DRAFT') AS status,
		       created_at
		FROM rfq_quotes
		WHERE rfq_id = ?
		ORDER BY sell_price ASC
	`
	var carrierQuoteRows []struct {
		ID              int64     `db:"id"`
		CarrierName     string    `db:"carrier_name"`
		TransitTimeDays int       `db:"transit_time_days"`
		BuyPrice        float64   `db:"buy_price"`
		SellPrice       float64   `db:"sell_price"`
		IsRecommended   bool      `db:"is_recommended"`
		Status          string    `db:"status"`
		CreatedAt       time.Time `db:"created_at"`
	}
	_ = s.db.SelectContext(ctx, &carrierQuoteRows, queryCarrierQuotes, rfqID)

	// 5. Build Quotation Comparison
	var quotationItems []RFQQuotationItemSummary
	var carrierQuoteItems []RFQCarrierQuoteItemSummary
	var validPrices []float64
	now := time.Now()

	expiredCount := 0
	rejectedCount := 0

	var selectedQuoteID *int64
	var selectedQuoteNumber string
	var selectedPrice *float64
	var selectedCarrier string
	var firstQuoteTime *time.Time
	var finalSelectTime *time.Time

	for _, q := range commQuoteRows {
		isExpired := q.ValidUntil != nil && q.ValidUntil.Before(now)
		if isExpired {
			expiredCount++
		}
		if q.Status == "REJECTED" {
			rejectedCount++
		}

		if firstQuoteTime == nil || q.CreatedAt.Before(*firstQuoteTime) {
			t := q.CreatedAt
			firstQuoteTime = &t
		}

		if q.Status == "ACCEPTED" || q.Status == "APPROVED" {
			id := q.ID
			selectedQuoteID = &id
			selectedQuoteNumber = q.QuotationNumber
			p := q.TotalAmount
			selectedPrice = &p
			t := q.CreatedAt
			finalSelectTime = &t
		}

		if !isExpired && q.Status != "REJECTED" && q.Status != "CANCELLED" && q.TotalAmount > 0 {
			validPrices = append(validPrices, q.TotalAmount)
		}

		quotationItems = append(quotationItems, RFQQuotationItemSummary{
			ID:              q.ID,
			QuotationNumber: q.QuotationNumber,
			Status:          q.Status,
			TotalAmount:     q.TotalAmount,
			TotalCost:       q.TotalCost,
			GrossProfit:     q.GrossProfit,
			GrossMarginPct:  q.GrossMarginPct,
			Currency:        q.Currency,
			ValidFrom:       q.ValidFrom,
			ValidUntil:      q.ValidUntil,
			IsExpired:       isExpired,
			CreatedAt:       q.CreatedAt,
		})
	}

	for _, cq := range carrierQuoteRows {
		marginPct := 0.0
		if cq.SellPrice > 0 {
			marginPct = ((cq.SellPrice - cq.BuyPrice) / cq.SellPrice) * 100.0
		}

		if firstQuoteTime == nil || cq.CreatedAt.Before(*firstQuoteTime) {
			t := cq.CreatedAt
			firstQuoteTime = &t
		}

		if cq.Status == "ACCEPTED" || cq.Status == "APPROVED" {
			id := cq.ID
			selectedQuoteID = &id
			p := cq.SellPrice
			selectedPrice = &p
			selectedCarrier = cq.CarrierName
			t := cq.CreatedAt
			finalSelectTime = &t
		} else if cq.IsRecommended && selectedQuoteID == nil {
			id := cq.ID
			selectedQuoteID = &id
			p := cq.SellPrice
			selectedPrice = &p
			selectedCarrier = cq.CarrierName
		}

		if cq.Status != "REJECTED" && cq.SellPrice > 0 {
			validPrices = append(validPrices, cq.SellPrice)
		}

		carrierQuoteItems = append(carrierQuoteItems, RFQCarrierQuoteItemSummary{
			ID:              cq.ID,
			CarrierName:     cq.CarrierName,
			TransitTimeDays: cq.TransitTimeDays,
			BuyPrice:        cq.BuyPrice,
			SellPrice:       cq.SellPrice,
			GrossMarginPct:  marginPct,
			IsRecommended:   cq.IsRecommended,
			Status:          cq.Status,
			CreatedAt:       cq.CreatedAt,
		})
	}

	// Calculate pricing statistics on valid prices
	var lowestValid, highestValid, averageValid, medianValid, priceSpread, priceSpreadPct *float64
	if len(validPrices) > 0 {
		sort.Float64s(validPrices)
		low := validPrices[0]
		high := validPrices[len(validPrices)-1]
		lowestValid = &low
		highestValid = &high

		sum := 0.0
		for _, p := range validPrices {
			sum += p
		}
		avg := sum / float64(len(validPrices))
		averageValid = &avg

		// Median
		mid := len(validPrices) / 2
		if len(validPrices)%2 == 1 {
			med := validPrices[mid]
			medianValid = &med
		} else {
			med := (validPrices[mid-1] + validPrices[mid]) / 2.0
			medianValid = &med
		}

		spread := high - low
		priceSpread = &spread
		if low > 0 {
			pct := (spread / low) * 100.0
			priceSpreadPct = &pct
		}
	}

	var timeToFirstQuoteHours *float64
	if firstQuoteTime != nil {
		hrs := firstQuoteTime.Sub(rfqRow.CreatedAt).Hours()
		if hrs < 0 {
			hrs = 0
		}
		timeToFirstQuoteHours = &hrs
	}

	var timeToFinalSelectHours *float64
	if finalSelectTime != nil {
		hrs := finalSelectTime.Sub(rfqRow.CreatedAt).Hours()
		if hrs < 0 {
			hrs = 0
		}
		timeToFinalSelectHours = &hrs
	}

	totalReceived := len(commQuoteRows) + len(carrierQuoteRows)
	quoteComparison := RFQQuotationComparison{
		TotalQuotationsRequested: totalReceived,
		TotalQuotationsReceived:  totalReceived,
		ValidQuotationsCount:     len(validPrices),
		ExpiredQuotationsCount:   expiredCount,
		RejectedQuotationsCount:  rejectedCount,
		LowestValidPrice:         lowestValid,
		HighestValidPrice:        highestValid,
		AverageValidPrice:        averageValid,
		MedianValidPrice:         medianValid,
		PriceSpread:              priceSpread,
		PriceSpreadPct:           priceSpreadPct,
		Currency:                 "USD",
		SelectedQuotationID:      selectedQuoteID,
		SelectedQuotationNumber:  selectedQuoteNumber,
		SelectedPrice:            selectedPrice,
		SelectedCarrier:          selectedCarrier,
		TimeToFirstQuoteHours:    timeToFirstQuoteHours,
		TimeToFinalSelectHours:   timeToFinalSelectHours,
		Quotations:               quotationItems,
		CarrierQuotes:            carrierQuoteItems,
	}

	// 6. Linked Bookings and Shipments
	queryBooking := `
		SELECT id, booking_number, status
		FROM bookings
		WHERE org_id = ? AND rfq_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`
	var bookingRow struct {
		ID            int64  `db:"id"`
		BookingNumber string `db:"booking_number"`
		Status        string `db:"status"`
	}
	hasBooking := false
	var linkedBookingID *int64
	if err := s.db.GetContext(ctx, &bookingRow, queryBooking, orgID, rfqID); err == nil && bookingRow.ID > 0 {
		hasBooking = true
		bID := bookingRow.ID
		linkedBookingID = &bID
	}

	queryShipment := `
		SELECT id, shipment_number, status
		FROM shipments
		WHERE org_id = ? AND (rfq_id = ? OR booking_id IN (SELECT id FROM bookings WHERE org_id = ? AND rfq_id = ?))
		ORDER BY created_at DESC
		LIMIT 1
	`
	var shipmentRow struct {
		ID             int64  `db:"id"`
		ShipmentNumber string `db:"shipment_number"`
		Status         string `db:"status"`
	}
	hasShipment := false
	var linkedShipmentID *int64
	if err := s.db.GetContext(ctx, &shipmentRow, queryShipment, orgID, rfqID, orgID, rfqID); err == nil && shipmentRow.ID > 0 {
		hasShipment = true
		sID := shipmentRow.ID
		linkedShipmentID = &sID
	}

	// 7. Lane Benchmarks & Historical Conversion
	queryLaneQuotes := `
		SELECT AVG(sell_price) AS avg_price, COUNT(*) AS quote_count
		FROM rfq_quotes rq
		JOIN rfqs r ON rq.rfq_id = r.id
		WHERE r.org_id = ? AND r.origin = ? AND r.destination = ? AND r.id != ? AND rq.sell_price > 0
	`
	var laneRow struct {
		AvgPrice   *float64 `db:"avg_price"`
		QuoteCount int      `db:"quote_count"`
	}
	_ = s.db.GetContext(ctx, &laneRow, queryLaneQuotes, orgID, rfqRow.Origin, rfqRow.Destination, rfqID)

	queryCustomerWinRate := `
		SELECT COUNT(*) as total_rfqs,
		       SUM(CASE WHEN stage IN ('WON', 'STAGE_BOOKED', 'BOOKED', 'COMPLETED') THEN 1 ELSE 0 END) as won_rfqs
		FROM rfqs
		WHERE org_id = ? AND customer_id = ?
	`
	var custStats struct {
		TotalRFQs int `db:"total_rfqs"`
		WonRFQs   int `db:"won_rfqs"`
	}
	_ = s.db.GetContext(ctx, &custStats, queryCustomerWinRate, orgID, rfqRow.CustomerID)

	var customerWinRate *float64
	if custStats.TotalRFQs > 0 {
		rate := (float64(custStats.WonRFQs) / float64(custStats.TotalRFQs)) * 100.0
		customerWinRate = &rate
	}

	volatilityNote := "Insufficient historical lane data for volatility index."
	if laneRow.QuoteCount >= 3 {
		volatilityNote = fmt.Sprintf("Lane pricing benchmark active: derived from %d historical quotes.", laneRow.QuoteCount)
	}

	commPerformance := RFQCommercialPerformance{
		HasQuotation:              totalReceived > 0,
		HasBooking:                hasBooking,
		HasShipment:               hasShipment,
		LinkedBookingID:           linkedBookingID,
		LinkedShipmentID:          linkedShipmentID,
		LaneAverageQuotedPrice:    laneRow.AvgPrice,
		LaneHistoricalQuotesCount: laneRow.QuoteCount,
		LanePricingVolatilityNote: volatilityNote,
		CustomerHistoricalWinRate: customerWinRate,
		CustomerTotalRFQs:         custStats.TotalRFQs,
	}

	// 8. Margin & Risk Evaluation
	var quotedRevenue, recordedCost, grossMarginAmount, grossMarginPct *float64
	var riskFactors []string
	isLowMargin := false
	isMissingCost := false
	isRateExpired := expiredCount > 0
	isPriceAnomaly := false

	// Evaluate from selected quote or best available quote
	if selectedPrice != nil && *selectedPrice > 0 {
		rev := *selectedPrice
		quotedRevenue = &rev

		var cost float64
		// Check carrier quote buy price
		for _, cq := range carrierQuoteRows {
			if selectedQuoteID != nil && cq.ID == *selectedQuoteID {
				cost = cq.BuyPrice
				break
			}
		}
		if cost == 0 {
			for _, q := range commQuoteRows {
				if selectedQuoteID != nil && q.ID == *selectedQuoteID {
					cost = q.TotalCost
					break
				}
			}
		}

		if cost > 0 {
			recordedCost = &cost
			gm := rev - cost
			grossMarginAmount = &gm
			pct := (gm / rev) * 100.0
			grossMarginPct = &pct

			if pct < 10.0 {
				isLowMargin = true
				riskFactors = append(riskFactors, fmt.Sprintf("Thin profit margin detected: %.1f%% (target >= 15%%)", pct))
			}
		} else {
			isMissingCost = true
			riskFactors = append(riskFactors, "Carrier cost data is missing or unverified.")
		}
	} else if len(carrierQuoteRows) > 0 {
		cq := carrierQuoteRows[0]
		rev := cq.SellPrice
		cost := cq.BuyPrice
		quotedRevenue = &rev
		recordedCost = &cost
		gm := rev - cost
		grossMarginAmount = &gm
		if rev > 0 {
			pct := (gm / rev) * 100.0
			grossMarginPct = &pct
			if pct < 10.0 {
				isLowMargin = true
				riskFactors = append(riskFactors, fmt.Sprintf("Thin profit margin: %.1f%%", pct))
			}
		}
	}

	if laneRow.AvgPrice != nil && selectedPrice != nil && *laneRow.AvgPrice > 0 {
		if *selectedPrice > (*laneRow.AvgPrice * 1.4) {
			isPriceAnomaly = true
			riskFactors = append(riskFactors, fmt.Sprintf("Quoted rate ($%.2f) is >40%% higher than historical lane average ($%.2f).", *selectedPrice, *laneRow.AvgPrice))
		}
	}

	if isRateExpired {
		riskFactors = append(riskFactors, "One or more quotation rates have passed their commercial validity window.")
	}

	marginHealth := "UNKNOWN"
	if grossMarginPct != nil {
		if *grossMarginPct >= 15.0 {
			marginHealth = "HEALTHY"
		} else if *grossMarginPct >= 0.0 {
			marginHealth = "THIN"
		} else {
			marginHealth = "NEGATIVE"
		}
	}

	riskSeverity := "LOW"
	if len(riskFactors) > 0 {
		if isLowMargin || isMissingCost {
			riskSeverity = "MEDIUM"
		}
		if isRateExpired || (grossMarginPct != nil && *grossMarginPct < 0) {
			riskSeverity = "HIGH"
		}
	}

	marginAndRisk := RFQMarginAndRiskIndicators{
		QuotedRevenue:      quotedRevenue,
		RecordedCost:       recordedCost,
		GrossMarginAmount:  grossMarginAmount,
		GrossMarginPct:     grossMarginPct,
		MarginHealth:       marginHealth,
		IsLowMargin:        isLowMargin,
		IsHighPriceAnomaly: isPriceAnomaly,
		IsMissingCostData:  isMissingCost,
		IsRateExpired:      isRateExpired,
		RiskSeverity:       riskSeverity,
		RiskFactors:        riskFactors,
	}

	// 9. Grounded Observations & Evidence Traceability
	var observations []RFQObservation

	obsFinding := fmt.Sprintf("RFQ %s created on lane %s → %s (Incoterms: %s).", rfqRow.RFQNumber, rfqRow.Origin, rfqRow.Destination, rfqRow.Incoterms)
	observations = append(observations, RFQObservation{
		Module:       "RFQ",
		RecordType:   "RFQ",
		RecordID:     rfqRow.RFQNumber,
		Finding:      obsFinding,
		Significance: "NEUTRAL",
		Timestamp:    &rfqRow.CreatedAt,
	})

	for _, cq := range carrierQuoteRows {
		finding := fmt.Sprintf("Carrier quote received from %s: Buy $%0.2f, Sell $%0.2f (Transit: %d days).", cq.CarrierName, cq.BuyPrice, cq.SellPrice, cq.TransitTimeDays)
		sig := "POSITIVE"
		if cq.IsRecommended {
			finding += " [AI Recommended Partner]"
		}
		t := cq.CreatedAt
		observations = append(observations, RFQObservation{
			Module:       "CARRIER",
			RecordType:   "RFQ_QUOTE",
			RecordID:     fmt.Sprintf("CQ-%d", cq.ID),
			Finding:      finding,
			Significance: sig,
			Timestamp:    &t,
		})
	}

	for _, q := range commQuoteRows {
		finding := fmt.Sprintf("Quotation %s status '%s': Total $%0.2f (Gross margin: %.1f%%).", q.QuotationNumber, q.Status, q.TotalAmount, q.GrossMarginPct)
		sig := "NEUTRAL"
		if q.Status == "APPROVED" || q.Status == "ACCEPTED" {
			sig = "POSITIVE"
		}
		t := q.CreatedAt
		observations = append(observations, RFQObservation{
			Module:       "QUOTATION",
			RecordType:   "QUOTATION",
			RecordID:     q.QuotationNumber,
			Finding:      finding,
			Significance: sig,
			Timestamp:    &t,
		})
	}

	if hasBooking {
		observations = append(observations, RFQObservation{
			Module:       "BOOKING",
			RecordType:   "BOOKING",
			RecordID:     bookingRow.BookingNumber,
			Finding:      fmt.Sprintf("Confirmed operational booking %s linked to RFQ.", bookingRow.BookingNumber),
			Significance: "POSITIVE",
		})
	}

	// 10. AI Summary Synthesis
	execSummary := fmt.Sprintf("RFQ %s for customer %s covers %s → %s with %d cargo items (%0.1f kg, %0.1f CBM). Current status is '%s'.",
		rfqRow.RFQNumber, rfqRow.CustomerName, rfqRow.Origin, rfqRow.Destination, len(itemRows), totalWeight, totalVolume, rfqRow.Stage)

	spreadAnalysis := "No commercial quotations recorded yet."
	if len(validPrices) > 1 && lowestValid != nil && highestValid != nil && priceSpread != nil {
		spreadAnalysis = fmt.Sprintf("%d valid quotations analyzed. Rates range from $%0.2f to $%0.2f (spread: $%0.2f, %0.1f%% variation).",
			len(validPrices), *lowestValid, *highestValid, *priceSpread, *priceSpreadPct)
	} else if len(validPrices) == 1 && lowestValid != nil {
		spreadAnalysis = fmt.Sprintf("Single quote received at $%0.2f. No carrier price spread available.", *lowestValid)
	}

	carrierInsight := "Awaiting carrier rate submissions."
	if len(carrierQuoteRows) > 0 {
		carrierInsight = fmt.Sprintf("%d carrier quotation(s) recorded. Shortest transit time: %d days.", len(carrierQuoteRows), carrierQuoteRows[0].TransitTimeDays)
		if selectedCarrier != "" {
			carrierInsight += fmt.Sprintf(" Selected routing via %s.", selectedCarrier)
		}
	}

	commercialRisks := "No critical pricing or margin risks identified."
	if len(riskFactors) > 0 {
		commercialRisks = strings.Join(riskFactors, " ")
	}

	var attentionItems []string
	if completenessScore < 100 {
		attentionItems = append(attentionItems, fmt.Sprintf("Incomplete RFQ: Missing %s.", strings.Join(missingFields, ", ")))
	}
	if len(validPrices) == 0 {
		attentionItems = append(attentionItems, "Zero active quotations. Request carrier spot rates to prevent quote deadline expiration.")
	}
	attentionItems = append(attentionItems, riskFactors...)

	var suggestedQuestions []string
	if isMissingCost {
		suggestedQuestions = append(suggestedQuestions, "Have origin and destination terminal handling charges been included in the buy price?")
	}
	if isLowMargin {
		suggestedQuestions = append(suggestedQuestions, "Can we negotiate carrier container detention terms or increase freight markup to achieve 15% margin?")
	}
	if len(validPrices) == 0 {
		suggestedQuestions = append(suggestedQuestions, "Should we trigger automated carrier rate requests for alternative ocean carriers on this lane?")
	}

	confidence := "HIGH"
	if len(validPrices) == 0 || completenessScore < 80 {
		confidence = "MEDIUM"
	}
	if len(validPrices) == 0 && len(itemRows) == 0 {
		confidence = "LOW"
	}

	aiSummary := RFQAISummary{
		ExecutiveSummary:         execSummary,
		QuotationSpreadAnalysis:  spreadAnalysis,
		CarrierResponseInsight:   carrierInsight,
		CommercialAndMarginRisks: commercialRisks,
		AttentionItems:           attentionItems,
		SuggestedQuestions:       suggestedQuestions,
		SupportingObservations:   observations,
		ConfidenceLevel:          confidence,
		DataFreshness:            fmt.Sprintf("Real-time database query as of %s", time.Now().UTC().Format("2006-01-02 15:04:05 UTC")),
		DataWarnings:             missingFields,
		Classification:           "READ_ONLY_INFORMATIONAL",
	}

	// 11. Audit log the read action for observability
	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorID:      &userID,
		ActorType:    domain.ActorTypeUser,
		Action:       "RFQ_PRICING_INTELLIGENCE_READ",
		Module:       "RFQ",
		ResourceType: "RFQ",
		ResourceID:   fmt.Sprintf("%d", rfqID),
		ResourceName: rfqRow.RFQNumber,
		Description:  fmt.Sprintf("Read-only RFQ and pricing intelligence generated for %s", rfqRow.RFQNumber),
		Result:       domain.ResultSuccess,
	})

	return &RFQ360PricingIntelligence{
		OrgID:                 orgID,
		RFQID:                 rfqID,
		Identity:              identity,
		QuotationComparison:  quoteComparison,
		CommercialPerformance: commPerformance,
		MarginAndRisk:         marginAndRisk,
		AISummary:             aiSummary,
		CorrelationID:         correlationID,
		DataFreshness:         now,
		IsReadOnly:            true,
	}, nil
}
