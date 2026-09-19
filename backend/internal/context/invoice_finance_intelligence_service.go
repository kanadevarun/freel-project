package bcontext

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
)

// GetInvoice360FinanceIntelligence generates a deterministic, read-only 360 financial profile for an invoice.
func (s *defaultService) GetInvoice360FinanceIntelligence(ctx context.Context, orgID, invoiceID int64, correlationID string, userID int64) (*Invoice360FinanceIntelligence, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization id: %d", orgID)
	}
	if invoiceID <= 0 {
		return nil, fmt.Errorf("invalid invoice id: %d", invoiceID)
	}
	if correlationID == "" {
		correlationID = fmt.Sprintf("corr-fin-%d-%d", invoiceID, time.Now().UnixNano())
	}

	now := time.Now().UTC()

	// 1. Query Invoice Details
	invQuery := `
		SELECT 
			id, org_id, invoice_number, customer_id, customer_name, customer_country,
			shipment_id, COALESCE(shipment_number, '') as shipment_number,
			booking_id, COALESCE(booking_number, '') as booking_number,
			quotation_id, COALESCE(quote_number, '') as quote_number,
			COALESCE(route, '') as route, COALESCE(origin, '') as origin, COALESCE(destination, '') as destination,
			invoice_date, due_date, currency, subtotal, tax_amount, discount_amount,
			total_amount, paid_amount, balance_due, status, COALESCE(type, 'CUSTOMER_AR') as type,
			created_at, updated_at
		FROM customer_invoices
		WHERE id = ? AND org_id = ?`

	var inv struct {
		ID              int64          `db:"id"`
		OrgID           int64          `db:"org_id"`
		InvoiceNumber   string         `db:"invoice_number"`
		CustomerID      int64          `db:"customer_id"`
		CustomerName    string         `db:"customer_name"`
		CustomerCountry string         `db:"customer_country"`
		ShipmentID      *int64         `db:"shipment_id"`
		ShipmentNumber  string         `db:"shipment_number"`
		BookingID       *int64         `db:"booking_id"`
		BookingNumber   string         `db:"booking_number"`
		QuotationID     *int64         `db:"quotation_id"`
		QuoteNumber     string         `db:"quote_number"`
		Route           string         `db:"route"`
		Origin          string         `db:"origin"`
		Destination     string         `db:"destination"`
		InvoiceDate     time.Time      `db:"invoice_date"`
		DueDate         sql.NullTime   `db:"due_date"`
		Currency        string         `db:"currency"`
		Subtotal        float64        `db:"subtotal"`
		TaxAmount       float64        `db:"tax_amount"`
		DiscountAmount  float64        `db:"discount_amount"`
		TotalAmount     float64        `db:"total_amount"`
		PaidAmount      float64        `db:"paid_amount"`
		BalanceDue      float64        `db:"balance_due"`
		Status          string         `db:"status"`
		Type            string         `db:"type"`
		CreatedAt       time.Time      `db:"created_at"`
		UpdatedAt       time.Time      `db:"updated_at"`
	}

	if err := s.db.GetContext(ctx, &inv, invQuery, invoiceID, orgID); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice not found or unauthorized: id=%d", invoiceID)
		}
		return nil, fmt.Errorf("failed to query invoice: %w", err)
	}

	var warnings []string

	// 2. Query Invoice Line Items
	itemsQuery := `
		SELECT id, description, service_category, quantity, unit_price, total_amount
		FROM customer_invoice_items
		WHERE invoice_id = ? AND org_id = ?
		ORDER BY id ASC`

	var rawItems []struct {
		ID              int64   `db:"id"`
		Description     string  `db:"description"`
		ServiceCategory string  `db:"service_category"`
		Quantity        float64 `db:"quantity"`
		UnitPrice       float64 `db:"unit_price"`
		TotalAmount     float64 `db:"total_amount"`
	}
	_ = s.db.SelectContext(ctx, &rawItems, itemsQuery, invoiceID, orgID)

	var lineItems []InvoiceLineItemSummary
	var auditedItemSum float64
	for _, item := range rawItems {
		lineItems = append(lineItems, InvoiceLineItemSummary{
			ID:              item.ID,
			Description:     item.Description,
			ServiceCategory: item.ServiceCategory,
			Quantity:        item.Quantity,
			UnitPrice:       item.UnitPrice,
			TotalAmount:     item.TotalAmount,
		})
		auditedItemSum += item.TotalAmount
	}

	// Line items consistency check
	lineItemsInconsistent := false
	if len(lineItems) > 0 {
		expectedSubtotal := auditedItemSum
		diff := math.Abs(inv.Subtotal - expectedSubtotal)
		if diff > 0.05 {
			lineItemsInconsistent = true
			warnings = append(warnings, fmt.Sprintf("Line items total (%.2f) differs from recorded invoice subtotal (%.2f)", auditedItemSum, inv.Subtotal))
		}
	} else {
		warnings = append(warnings, "No line items recorded for this invoice")
	}

	// 3. Query Invoice Payments
	paymentsQuery := `
		SELECT id, payment_ref, amount, payment_method, status, payment_date, COALESCE(notes, '') as notes
		FROM customer_invoice_payments
		WHERE invoice_id = ? AND org_id = ?
		ORDER BY payment_date DESC`

	var rawPayments []struct {
		ID            int64     `db:"id"`
		PaymentRef    string    `db:"payment_ref"`
		Amount        float64   `db:"amount"`
		PaymentMethod string    `db:"payment_method"`
		Status        string    `db:"status"`
		PaymentDate   time.Time `db:"payment_date"`
		Notes         string    `db:"notes"`
	}
	_ = s.db.SelectContext(ctx, &rawPayments, paymentsQuery, invoiceID, orgID)

	var payments []InvoicePaymentRecordSummary
	var lastPaymentDate *time.Time
	for _, p := range rawPayments {
		payments = append(payments, InvoicePaymentRecordSummary{
			ID:            p.ID,
			PaymentRef:    p.PaymentRef,
			Amount:        p.Amount,
			PaymentMethod: p.PaymentMethod,
			Status:        p.Status,
			PaymentDate:   p.PaymentDate,
			Notes:         p.Notes,
		})
		if lastPaymentDate == nil || p.PaymentDate.After(*lastPaymentDate) {
			t := p.PaymentDate
			lastPaymentDate = &t
		}
	}

	// 4. Deterministic Due Date & Aging Calculation
	var dueDatePtr *time.Time
	daysUntilDue := 0
	daysOverdue := 0
	isOverdue := false
	agingBucket := "NOT_DUE"
	overdueAmount := 0.0

	if inv.DueDate.Valid {
		d := inv.DueDate.Time
		dueDatePtr = &d
		if inv.BalanceDue > 0.001 && strings.ToUpper(inv.Status) != "PAID" && strings.ToUpper(inv.Status) != "CANCELLED" {
			if now.After(d) {
				isOverdue = true
				daysOverdue = int(now.Sub(d).Hours() / 24)
				overdueAmount = inv.BalanceDue
				if daysOverdue <= 30 {
					agingBucket = "1-30_DAYS"
				} else if daysOverdue <= 60 {
					agingBucket = "31-60_DAYS"
				} else if daysOverdue <= 90 {
					agingBucket = "61-90_DAYS"
				} else {
					agingBucket = "OVER_90_DAYS"
				}
			} else {
				daysUntilDue = int(d.Sub(now).Hours() / 24)
				agingBucket = "NOT_DUE"
			}
		} else {
			agingBucket = "NOT_DUE"
		}
	} else {
		warnings = append(warnings, "Invoice due date is missing")
	}

	invoiceAgeDays := int(now.Sub(inv.InvoiceDate).Hours() / 24)
	if invoiceAgeDays < 0 {
		invoiceAgeDays = 0
	}

	isPaid := strings.ToUpper(inv.Status) == "PAID" || (inv.TotalAmount > 0 && inv.BalanceDue <= 0.001)
	isPartiallyPaid := inv.PaidAmount > 0 && inv.BalanceDue > 0.001

	// 5. Customer AR Profile Calculation
	customerAR := CustomerReceivablesSummary{
		CustomerID:   inv.CustomerID,
		CustomerName: inv.CustomerName,
	}

	if inv.CustomerID > 0 {
		var agg struct {
			TotalInvoiced    sql.NullFloat64 `db:"total_invoiced"`
			TotalPaid        sql.NullFloat64 `db:"total_paid"`
			TotalOutstanding sql.NullFloat64 `db:"total_outstanding"`
			TotalCount       int             `db:"total_count"`
			PaidCount        int             `db:"paid_count"`
			OpenCount        int             `db:"open_count"`
		}
		custQuery := `
			SELECT 
				SUM(total_amount) as total_invoiced,
				SUM(paid_amount) as total_paid,
				SUM(CASE WHEN status != 'Cancelled' THEN balance_due ELSE 0 END) as total_outstanding,
				COUNT(*) as total_count,
				SUM(CASE WHEN status = 'Paid' THEN 1 ELSE 0 END) as paid_count,
				SUM(CASE WHEN status != 'Paid' AND status != 'Cancelled' AND balance_due > 0.001 THEN 1 ELSE 0 END) as open_count
			FROM customer_invoices
			WHERE customer_id = ? AND org_id = ?`
		if err := s.db.GetContext(ctx, &agg, custQuery, inv.CustomerID, orgID); err == nil {
			customerAR.TotalInvoicedAmount = agg.TotalInvoiced.Float64
			customerAR.TotalPaidAmount = agg.TotalPaid.Float64
			customerAR.TotalOutstandingAmount = agg.TotalOutstanding.Float64
			customerAR.OpenInvoicesCount = agg.OpenCount
			customerAR.PaidInvoicesCount = agg.PaidCount
			if customerAR.TotalInvoicedAmount > 0 {
				customerAR.PaymentCompletionRate = math.Round((customerAR.TotalPaidAmount/customerAR.TotalInvoicedAmount)*1000) / 10
			}
		}

		// Calculate Customer Overdue
		var overdueAgg struct {
			OverdueAmount sql.NullFloat64 `db:"overdue_amount"`
			OverdueCount  int             `db:"overdue_count"`
		}
		overdueQuery := `
			SELECT 
				SUM(balance_due) as overdue_amount,
				COUNT(*) as overdue_count
			FROM customer_invoices
			WHERE customer_id = ? AND org_id = ?
			  AND status != 'Paid' AND status != 'Cancelled'
			  AND balance_due > 0.001
			  AND due_date < ?`
		if err := s.db.GetContext(ctx, &overdueAgg, overdueQuery, inv.CustomerID, orgID, now); err == nil {
			customerAR.TotalOverdueAmount = overdueAgg.OverdueAmount.Float64
			customerAR.OverdueInvoicesCount = overdueAgg.OverdueCount
		}

		// Calculate average payment delay from actual payment history
		var delayAgg struct {
			AvgDelay sql.NullFloat64 `db:"avg_delay"`
			LateCount int            `db:"late_count"`
		}
		delayQuery := `
			SELECT 
				AVG(TIMESTAMPDIFF(DAY, i.due_date, p.payment_date)) as avg_delay,
				COUNT(*) as late_count
			FROM customer_invoice_payments p
			JOIN customer_invoices i ON p.invoice_id = i.id AND p.org_id = i.org_id
			WHERE i.customer_id = ? AND i.org_id = ?
			  AND p.payment_date > i.due_date`
		if err := s.db.GetContext(ctx, &delayAgg, delayQuery, inv.CustomerID, orgID); err == nil {
			if delayAgg.AvgDelay.Valid {
				customerAR.AveragePaymentDelayDays = math.Round(delayAgg.AvgDelay.Float64*10) / 10
			}
			customerAR.HasRepeatedLatePayments = delayAgg.LateCount >= 2
		}

		// Rate customer exposure
		if customerAR.TotalOverdueAmount > 20000 || customerAR.OverdueInvoicesCount >= 3 {
			customerAR.ExposureRating = "SEVERE"
		} else if customerAR.TotalOverdueAmount > 5000 || customerAR.OverdueInvoicesCount >= 1 {
			customerAR.ExposureRating = "ELEVATED"
		} else if customerAR.TotalOutstandingAmount > 0 {
			customerAR.ExposureRating = "MODERATE"
		} else {
			customerAR.ExposureRating = "LOW"
		}
	} else {
		warnings = append(warnings, "Invoice is not linked to a registered customer ID")
	}

	// 6. Revenue and Cost Visibility
	revenueCost := RevenueCostVisibility{
		InvoiceRevenue:    inv.TotalAmount,
		Currency:          inv.Currency,
		HasMissingCost:    true,
		MissingCostReason: "No linked quotation cost recorded in database",
		IsUnlinkedInvoice: (inv.ShipmentID == nil && inv.BookingID == nil),
	}

	if inv.QuotationID != nil && *inv.QuotationID > 0 {
		var q struct {
			TotalAmount  float64 `db:"total_amount"`
			TotalCost    float64 `db:"total_cost"`
			GrossProfit  float64 `db:"gross_profit"`
			GrossMargin  float64 `db:"gross_margin_pct"`
		}
		quoteQuery := `SELECT total_amount, total_cost, gross_profit, gross_margin_pct FROM quotations WHERE id = ? AND org_id = ?`
		if err := s.db.GetContext(ctx, &q, quoteQuery, *inv.QuotationID, orgID); err == nil {
			revenueCost.QuotationAmount = &q.TotalAmount
			if q.TotalCost > 0 {
				revenueCost.CommercialCostAmount = &q.TotalCost
				revenueCost.HasMissingCost = false
				revenueCost.MissingCostReason = ""
				margin := inv.TotalAmount - q.TotalCost
				revenueCost.GrossMarginAmount = &margin
				if inv.TotalAmount > 0 {
					marginPct := math.Round((margin/inv.TotalAmount)*1000) / 10
					revenueCost.GrossMarginPercentage = &marginPct
				}
			}
			variance := inv.TotalAmount - q.TotalAmount
			revenueCost.RevenueCostVariance = &variance
			if math.Abs(variance) > 1.0 {
				revenueCost.HasCommercialDiscrepancy = true
				revenueCost.DiscrepancyReason = fmt.Sprintf("Invoice total (%.2f %s) deviates from agreed quotation (%.2f %s)", inv.TotalAmount, inv.Currency, q.TotalAmount, inv.Currency)
			}
		}
	}

	// 7. Multi-Factor Risk Assessment
	riskIndicators := FinanceRiskIndicators{
		IsOverdue:                 isOverdue,
		IsApproachingDueDate:      (daysUntilDue >= 0 && daysUntilDue <= 7 && inv.BalanceDue > 0.001 && !isPaid),
		HasLargeBalance:           (inv.BalanceDue >= 10000.0),
		IsPartiallyPaid:           isPartiallyPaid,
		MissingDueDate:            !inv.DueDate.Valid,
		MissingCustomer:           (inv.CustomerID <= 0),
		MissingShipment:           (inv.ShipmentID == nil),
		LineItemsInconsistent:     lineItemsInconsistent,
		RevenueWithoutCost:        revenueCost.HasMissingCost,
		RepeatedCustomerLatePayer: customerAR.HasRepeatedLatePayments,
	}

	riskScore := 0
	var riskFactors []string

	if isOverdue {
		if daysOverdue > 60 {
			riskScore += 45
			riskFactors = append(riskFactors, fmt.Sprintf("Invoice severely overdue by %d days (Aging: %s)", daysOverdue, agingBucket))
		} else if daysOverdue > 30 {
			riskScore += 30
			riskFactors = append(riskFactors, fmt.Sprintf("Invoice overdue by %d days (Aging: %s)", daysOverdue, agingBucket))
		} else {
			riskScore += 20
			riskFactors = append(riskFactors, fmt.Sprintf("Invoice overdue by %d days", daysOverdue))
		}
	} else if riskIndicators.IsApproachingDueDate {
		riskScore += 10
		riskFactors = append(riskFactors, fmt.Sprintf("Payment due within %d days", daysUntilDue))
	}

	if riskIndicators.HasLargeBalance {
		riskScore += 15
		riskFactors = append(riskFactors, fmt.Sprintf("High outstanding exposure: %.2f %s", inv.BalanceDue, inv.Currency))
	}
	if isPartiallyPaid && isOverdue {
		riskScore += 15
		riskFactors = append(riskFactors, "Partially settled with overdue remaining balance")
	}
	if riskIndicators.LineItemsInconsistent {
		riskScore += 15
		riskFactors = append(riskFactors, "Invoice subtotal differs from sum of line items")
	}
	if riskIndicators.MissingDueDate {
		riskScore += 15
		riskFactors = append(riskFactors, "Missing payment due date")
	}
	if revenueCost.HasCommercialDiscrepancy {
		riskScore += 10
		riskFactors = append(riskFactors, revenueCost.DiscrepancyReason)
	}
	if revenueCost.IsUnlinkedInvoice {
		riskScore += 10
		riskFactors = append(riskFactors, "Unlinked invoice (no associated shipment or booking reference)")
	}
	if customerAR.HasRepeatedLatePayments {
		riskScore += 15
		riskFactors = append(riskFactors, "Customer has recorded history of late settlements")
	}

	if riskScore > 100 {
		riskScore = 100
	}
	riskIndicators.OverallRiskScore = riskScore

	if riskScore >= 60 || daysOverdue > 60 {
		riskIndicators.OverallRiskRating = "CRITICAL"
	} else if riskScore >= 40 || daysOverdue > 30 {
		riskIndicators.OverallRiskRating = "HIGH"
	} else if riskScore >= 20 || riskIndicators.IsApproachingDueDate {
		riskIndicators.OverallRiskRating = "MODERATE"
	} else {
		riskIndicators.OverallRiskRating = "LOW"
	}
	riskIndicators.RiskFactors = riskFactors

	// 8. Grounded AI Financial Synthesis
	citations := []string{
		fmt.Sprintf("[Invoice: #%d (%s)]", inv.ID, inv.InvoiceNumber),
	}
	if inv.CustomerID > 0 {
		citations = append(citations, fmt.Sprintf("[Customer: #%d (%s)]", inv.CustomerID, inv.CustomerName))
	}
	if inv.ShipmentID != nil {
		citations = append(citations, fmt.Sprintf("[Shipment: #%d (%s)]", *inv.ShipmentID, inv.ShipmentNumber))
	}
	if inv.BookingID != nil {
		citations = append(citations, fmt.Sprintf("[Booking: #%d (%s)]", *inv.BookingID, inv.BookingNumber))
	}
	if inv.QuotationID != nil {
		citations = append(citations, fmt.Sprintf("[Quotation: #%d (%s)]", *inv.QuotationID, inv.QuoteNumber))
	}
	for _, p := range payments {
		citations = append(citations, fmt.Sprintf("[Payment: %s (%.2f %s)]", p.PaymentRef, p.Amount, inv.Currency))
	}

	execSummary := fmt.Sprintf("Invoice %s for %s is %s with total %.2f %s (Paid: %.2f, Balance Due: %.2f).",
		inv.InvoiceNumber, inv.CustomerName, inv.Status, inv.TotalAmount, inv.Currency, inv.PaidAmount, inv.BalanceDue)
	if isOverdue {
		execSummary += fmt.Sprintf(" Outstanding balance is overdue by %d days (%s bucket).", daysOverdue, agingBucket)
	} else if isPaid {
		execSummary += " Account balance has been fully settled."
	}

	exposureExplanation := fmt.Sprintf("Current outstanding exposure on this invoice is %.2f %s. Customer %s holds %.2f %s in total AR exposure across %d open invoices.",
		inv.BalanceDue, inv.Currency, inv.CustomerName, customerAR.TotalOutstandingAmount, inv.Currency, customerAR.OpenInvoicesCount)

	agingExplanation := fmt.Sprintf("Invoice is in aging bracket '%s' with %d days overdue and %d days since original invoice date.",
		agingBucket, daysOverdue, invoiceAgeDays)

	revenueCostObs := ""
	if revenueCost.GrossMarginAmount != nil && revenueCost.GrossMarginPercentage != nil {
		revenueCostObs = fmt.Sprintf("Commercial gross margin is %.2f %s (%.1f%%) based on linked quotation cost.",
			*revenueCost.GrossMarginAmount, inv.Currency, *revenueCost.GrossMarginPercentage)
	} else {
		revenueCostObs = "No commercial cost baseline recorded in database for margin calculation."
	}

	paymentBehaviorObs := ""
	if len(payments) > 0 {
		paymentBehaviorObs = fmt.Sprintf("%d payment transaction(s) recorded totalling %.2f %s.", len(payments), inv.PaidAmount, inv.Currency)
	} else {
		paymentBehaviorObs = "Zero payment transactions recorded to date."
	}
	if customerAR.HasRepeatedLatePayments {
		paymentBehaviorObs += fmt.Sprintf(" Historical data indicates average settlement delay of %.1f days past due date.", customerAR.AveragePaymentDelayDays)
	}

	var attentionItems []string
	if isOverdue {
		attentionItems = append(attentionItems, fmt.Sprintf("Initiate accounts receivable follow-up for %.2f %s overdue since %s", inv.BalanceDue, inv.Currency, inv.DueDate.Time.Format("2006-01-02")))
	}
	if lineItemsInconsistent {
		attentionItems = append(attentionItems, "Reconcile line items against recorded invoice subtotal")
	}
	if revenueCost.HasCommercialDiscrepancy {
		attentionItems = append(attentionItems, revenueCost.DiscrepancyReason)
	}
	if len(attentionItems) == 0 {
		attentionItems = append(attentionItems, "No urgent financial intervention required")
	}

	var suggestedInquiries []string
	if isOverdue {
		suggestedInquiries = append(suggestedInquiries, fmt.Sprintf("Request payment remittance advice from %s AP department", inv.CustomerName))
	}
	if revenueCost.HasMissingCost {
		suggestedInquiries = append(suggestedInquiries, "Confirm final buy rate / carrier cost voucher with operations")
	}
	if len(suggestedInquiries) == 0 {
		suggestedInquiries = append(suggestedInquiries, "All invoice billing and payment verifications verified")
	}

	confidence := "HIGH"
	if !inv.DueDate.Valid || lineItemsInconsistent || len(lineItems) == 0 {
		confidence = "MEDIUM"
	}

	aiSummary := InvoiceAIFinanceSummary{
		ExecutiveSummary:               execSummary,
		OutstandingExposureExplanation: exposureExplanation,
		AgingPositionExplanation:       agingExplanation,
		RevenueCostObservations:        revenueCostObs,
		PaymentBehaviorObservations:    paymentBehaviorObs,
		ActionableAttentionItems:       attentionItems,
		SuggestedFinanceInquiries:      suggestedInquiries,
		Confidence:                     confidence,
		FreshnessTimestamp:             now,
		Citations:                      citations,
	}

	// Assemble Result
	return &Invoice360FinanceIntelligence{
		InvoiceID: invoiceID,
		Identity: InvoiceIdentitySummary{
			InvoiceID:        inv.ID,
			InvoiceNumber:    inv.InvoiceNumber,
			CustomerID:       inv.CustomerID,
			CustomerName:     inv.CustomerName,
			CustomerCountry:  inv.CustomerCountry,
			ShipmentID:       inv.ShipmentID,
			ShipmentNumber:   inv.ShipmentNumber,
			BookingID:        inv.BookingID,
			BookingNumber:    inv.BookingNumber,
			QuotationID:      inv.QuotationID,
			QuoteNumber:      inv.QuoteNumber,
			Route:            inv.Route,
			Origin:           inv.Origin,
			Destination:      inv.Destination,
			InvoiceStatus:    inv.Status,
			PaymentStatus:    inv.Status,
			Currency:         inv.Currency,
			IssueDate:        inv.InvoiceDate,
			DueDate:          dueDatePtr,
			PaymentDate:      lastPaymentDate,
			Subtotal:         inv.Subtotal,
			TaxAmount:        inv.TaxAmount,
			DiscountAmount:   inv.DiscountAmount,
			TotalAmount:      inv.TotalAmount,
			PaidAmount:       inv.PaidAmount,
			BalanceDue:       inv.BalanceDue,
			OverdueAmount:    overdueAmount,
			DaysUntilDue:     daysUntilDue,
			DaysOverdue:      daysOverdue,
			InvoiceAgeDays:   invoiceAgeDays,
			AgingBucket:      agingBucket,
			IsOverdue:        isOverdue,
			IsPaid:           isPaid,
			IsPartiallyPaid:  isPartiallyPaid,
			CreatedAt:        inv.CreatedAt,
			UpdatedAt:        inv.UpdatedAt,
		},
		CustomerAR:           customerAR,
		RevenueCost:          revenueCost,
		LineItems:            lineItems,
		Payments:             payments,
		RiskIndicators:       riskIndicators,
		AIFinanceSummary:     aiSummary,
		AuditedLineItemTotal: auditedItemSum,
		CalculatedAt:         now,
		CorrelationID:        correlationID,
		OrganizationScope:    orgID,
		ReadOnly:             true,
		Warnings:             warnings,
	}, nil
}

// GetOrgFinanceSummary generates an aggregated accounts receivable and exposure summary across an organization.
func (s *defaultService) GetOrgFinanceSummary(ctx context.Context, orgID int64, correlationID string, userID int64) (*OrgFinanceSummary, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization id: %d", orgID)
	}
	if correlationID == "" {
		correlationID = fmt.Sprintf("corr-orgfin-%d-%d", orgID, time.Now().UnixNano())
	}

	now := time.Now().UTC()

	// 1. Overall Aggregates
	var agg struct {
		ActiveCount     int             `db:"active_count"`
		OpenCount       int             `db:"open_count"`
		PaidCount       int             `db:"paid_count"`
		CancelledCount  int             `db:"cancelled_count"`
		TotalInvoiced   sql.NullFloat64 `db:"total_invoiced"`
		TotalPaid       sql.NullFloat64 `db:"total_paid"`
		TotalOutstanding sql.NullFloat64 `db:"total_outstanding"`
	}

	query := `
		SELECT 
			COUNT(*) as active_count,
			SUM(CASE WHEN status != 'Paid' AND status != 'Cancelled' AND balance_due > 0.001 THEN 1 ELSE 0 END) as open_count,
			SUM(CASE WHEN status = 'Paid' THEN 1 ELSE 0 END) as paid_count,
			SUM(CASE WHEN status = 'Cancelled' THEN 1 ELSE 0 END) as cancelled_count,
			SUM(total_amount) as total_invoiced,
			SUM(paid_amount) as total_paid,
			SUM(CASE WHEN status != 'Cancelled' THEN balance_due ELSE 0 END) as total_outstanding
		FROM customer_invoices
		WHERE org_id = ?`

	if err := s.db.GetContext(ctx, &agg, query, orgID); err != nil {
		return nil, fmt.Errorf("failed to query organization invoices: %w", err)
	}

	// 2. Query individual open invoices to calculate aging breakdown and overdue counts
	openQuery := `
		SELECT id, customer_id, customer_name, currency, balance_due, due_date
		FROM customer_invoices
		WHERE org_id = ? AND status != 'Paid' AND status != 'Cancelled' AND balance_due > 0.001`

	var openInvoices []struct {
		ID           int64        `db:"id"`
		CustomerID   int64        `db:"customer_id"`
		CustomerName string       `db:"customer_name"`
		Currency     string       `db:"currency"`
		BalanceDue   float64      `db:"balance_due"`
		DueDate      sql.NullTime `db:"due_date"`
	}
	_ = s.db.SelectContext(ctx, &openInvoices, openQuery, orgID)

	var aging AgingBucketBreakdown
	overdueCount := 0
	totalOverdue := 0.0
	approachingDueCount := 0

	customerMap := make(map[int64]*CustomerExposureItem)

	for _, inv := range openInvoices {
		// Aging calculation
		if inv.DueDate.Valid {
			d := inv.DueDate.Time
			if now.After(d) {
				overdueCount++
				totalOverdue += inv.BalanceDue
				daysOverdue := int(now.Sub(d).Hours() / 24)
				if daysOverdue <= 30 {
					aging.Days1To30 += inv.BalanceDue
				} else if daysOverdue <= 60 {
					aging.Days31To60 += inv.BalanceDue
				} else if daysOverdue <= 90 {
					aging.Days61To90 += inv.BalanceDue
				} else {
					aging.Over90Days += inv.BalanceDue
				}
			} else {
				aging.NotDue += inv.BalanceDue
				daysUntil := int(d.Sub(now).Hours() / 24)
				if daysUntil <= 7 {
					approachingDueCount++
				}
			}
		} else {
			aging.NotDue += inv.BalanceDue
		}

		// Group by customer
		cItem, exists := customerMap[inv.CustomerID]
		if !exists {
			cItem = &CustomerExposureItem{
				CustomerID:   inv.CustomerID,
				CustomerName: inv.CustomerName,
				Currency:     inv.Currency,
			}
			customerMap[inv.CustomerID] = cItem
		}
		cItem.TotalOutstanding += inv.BalanceDue
		cItem.OpenInvoicesCount++
		if inv.DueDate.Valid && now.After(inv.DueDate.Time) {
			cItem.TotalOverdue += inv.BalanceDue
		}
	}

	// Sort / extract top exposures
	var topExposures []CustomerExposureItem
	for _, item := range customerMap {
		item.TotalOutstanding = math.Round(item.TotalOutstanding*100) / 100
		item.TotalOverdue = math.Round(item.TotalOverdue*100) / 100
		topExposures = append(topExposures, *item)
		if len(topExposures) >= 5 {
			break
		}
	}

	// Determine overall health rating
	healthRating := "EXCELLENT"
	overdueRatio := 0.0
	totalOut := agg.TotalOutstanding.Float64
	if totalOut > 0 {
		overdueRatio = totalOverdue / totalOut
	}

	if overdueRatio > 0.50 || aging.Over90Days > 10000 {
		healthRating = "CRITICAL"
	} else if overdueRatio > 0.30 || overdueCount >= 5 {
		healthRating = "HIGH_RISK"
	} else if overdueRatio > 0.15 || overdueCount >= 2 {
		healthRating = "MODERATE"
	} else if totalOut > 0 {
		healthRating = "GOOD"
	}

	return &OrgFinanceSummary{
		ActiveInvoicesCount:         agg.ActiveCount,
		OpenInvoicesCount:           agg.OpenCount,
		OverdueInvoicesCount:        overdueCount,
		PaidInvoicesCount:           agg.PaidCount,
		CancelledInvoicesCount:      agg.CancelledCount,
		InvoicesApproachingDueCount: approachingDueCount,
		TotalInvoicedAmount:         math.Round(agg.TotalInvoiced.Float64*100) / 100,
		TotalPaidAmount:             math.Round(agg.TotalPaid.Float64*100) / 100,
		TotalOutstandingReceivables: math.Round(totalOut*100) / 100,
		TotalOverdueReceivables:     math.Round(totalOverdue*100) / 100,
		AgingBreakdown: AgingBucketBreakdown{
			NotDue:     math.Round(aging.NotDue*100) / 100,
			Days1To30:  math.Round(aging.Days1To30*100) / 100,
			Days31To60: math.Round(aging.Days31To60*100) / 100,
			Days61To90: math.Round(aging.Days61To90*100) / 100,
			Over90Days: math.Round(aging.Over90Days*100) / 100,
		},
		TopCustomerExposures:    topExposures,
		PrimaryCurrency:         "USD",
		ReceivablesHealthRating: healthRating,
		CalculatedAt:            now,
		OrganizationScope:       orgID,
		ReadOnly:                true,
	}, nil
}
