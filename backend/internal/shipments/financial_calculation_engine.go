package shipments

import (
	"math"
	"strings"

	"github.com/freel/backend/internal/shipments/spec"
)

// RFQQuoteCommercialSnapshot captures upstream quote commercial terms.
type RFQQuoteCommercialSnapshot struct {
	ID                 int64   `db:"id" json:"id"`
	RFQID              int64   `db:"rfq_id" json:"rfq_id"`
	CarrierName        string  `db:"carrier_name" json:"carrier_name"`
	BuyPrice           float64 `db:"buy_price" json:"buy_price"`
	SellPrice          float64 `db:"sell_price" json:"sell_price"`
	OceanFreight       float64 `db:"ocean_freight" json:"ocean_freight"`
	OriginCharges      float64 `db:"origin_charges" json:"origin_charges"`
	DestinationCharges float64 `db:"destination_charges" json:"destination_charges"`
	TotalBuyPrice      float64 `db:"total_buy_price" json:"total_buy_price"`
	Currency           string  `db:"currency" json:"currency"`
	Status             string  `db:"status" json:"status"`
}

// CarrierInvoiceSnapshot captures ingested carrier invoice context.
type CarrierInvoiceSnapshot struct {
	ID          int64   `db:"id" json:"id"`
	TotalAmount float64 `db:"total_amount" json:"total_amount"`
	Currency    string  `db:"currency" json:"currency"`
	Status      string  `db:"status" json:"status"`
}

// CustomerInvoiceSnapshot captures generated customer billing receivable context.
type CustomerInvoiceSnapshot struct {
	ID          int64   `db:"id" json:"id"`
	TotalAmount float64 `db:"total_amount" json:"total_amount"`
	Currency    string  `db:"currency" json:"currency"`
	Status      string  `db:"status" json:"status"`
}

// CalculateFinancialSummary synthesizes upstream quotes, bookings, carrier invoices,
// customer invoices, and direct operational charges into an authoritative financial summary.
func CalculateFinancialSummary(
	sh *spec.Shipment,
	quote *RFQQuoteCommercialSnapshot,
	charges []*spec.ShipmentFinancialCharge,
	carrierInvoices []*CarrierInvoiceSnapshot,
	customerInvoices []*CustomerInvoiceSnapshot,
) *spec.ShipmentFinancialSummary {
	currency := "USD"
	if quote != nil && quote.Currency != "" {
		currency = quote.Currency
	}

	var estRevenue, actRevenue float64
	var estCost, actCost float64

	// 1. Upstream Quote Commercial baseline
	if quote != nil {
		estRevenue += quote.SellPrice
		if quote.TotalBuyPrice > 0 {
			estCost += quote.TotalBuyPrice
		} else if quote.BuyPrice > 0 {
			estCost += quote.BuyPrice
		}
	}

	// 2. Customer Invoices (Actual Revenue)
	var hasApprovedCustInvoices bool
	for _, ci := range customerInvoices {
		st := strings.ToUpper(strings.TrimSpace(ci.Status))
		if st == "APPROVED" || st == "SENT" || st == "PAID" {
			actRevenue += ci.TotalAmount
			hasApprovedCustInvoices = true
		}
	}
	// Fallback to estimated quote sell price if no separate customer invoice generated yet
	if !hasApprovedCustInvoices && quote != nil && quote.SellPrice > 0 {
		actRevenue = quote.SellPrice
	}

	// 3. Carrier Invoices (Actual Carrier Cost)
	var hasApprovedCarrierInvoices bool
	var totalApprovedCarrierInvoices float64
	for _, inv := range carrierInvoices {
		st := strings.ToUpper(strings.TrimSpace(inv.Status))
		if st == spec.ChargeStatusApproved || st == "VERIFIED" || st == "PAID" {
			totalApprovedCarrierInvoices += inv.TotalAmount
			hasApprovedCarrierInvoices = true
		}
	}
	if hasApprovedCarrierInvoices {
		actCost += totalApprovedCarrierInvoices
	} else if quote != nil {
		// If no carrier invoices ingested yet, use quoted baseline for cost
		if quote.TotalBuyPrice > 0 {
			actCost = quote.TotalBuyPrice
		} else {
			actCost = quote.BuyPrice
		}
	}

	// 4. Line item operational charges
	var totalChargesCount, pendingChargesCount int
	var hasDisputedCharge bool

	for _, ch := range charges {
		totalChargesCount++
		st := strings.ToUpper(strings.TrimSpace(ch.Status))
		if st == spec.ChargeStatusDisputed {
			hasDisputedCharge = true
		}
		if st == spec.ChargeStatusEstimated || st == spec.ChargeStatusInvoiced {
			pendingChargesCount++
		}

		cType := strings.ToUpper(strings.TrimSpace(ch.ChargeType))
		if cType == spec.ChargeTypeRevenue {
			estRevenue += ch.EstimatedAmount
			if ch.ActualAmount > 0 {
				actRevenue += (ch.ActualAmount - ch.EstimatedAmount) // increment above baseline
			}
		} else { // COST
			estCost += ch.EstimatedAmount
			if ch.ActualAmount > 0 {
				actCost += ch.ActualAmount
			}
		}
	}

	// 5. Margin & Variance calculations
	estMargin := estRevenue - estCost
	actMargin := actRevenue - actCost

	var estMarginPct, actMarginPct float64
	if estRevenue > 0 {
		estMarginPct = math.Round((estMargin/estRevenue)*1000) / 10
	}
	if actRevenue > 0 {
		actMarginPct = math.Round((actMargin/actRevenue)*1000) / 10
	}

	costVariance := actCost - estCost
	var costVariancePct float64
	if estCost > 0 {
		costVariancePct = math.Round((costVariance/estCost)*1000) / 10
	}

	// 6. Determine Financial Health Status
	financialStatus := spec.FinancialStatusEstimated
	if hasDisputedCharge {
		financialStatus = spec.FinancialStatusPendingReview
	} else if actRevenue > 0 && actCost > 0 {
		if actMargin < 0 {
			financialStatus = spec.FinancialStatusLoss
		} else if actMarginPct < 8.0 {
			financialStatus = spec.FinancialStatusLowMargin
		} else {
			financialStatus = spec.FinancialStatusProfitable
		}
	} else if len(charges) > 0 || len(carrierInvoices) > 0 {
		financialStatus = spec.FinancialStatusInProgress
	}

	summary := &spec.ShipmentFinancialSummary{
		ShipmentID:             sh.ID,
		OrgID:                  sh.OrgID,
		Currency:               currency,
		EstimatedRevenue:       math.Round(estRevenue*100) / 100,
		ActualRevenue:          math.Round(actRevenue*100) / 100,
		EstimatedCost:          math.Round(estCost*100) / 100,
		ActualCost:             math.Round(actCost*100) / 100,
		EstimatedMargin:        math.Round(estMargin*100) / 100,
		ActualMargin:           math.Round(actMargin*100) / 100,
		EstimatedMarginPercent: estMarginPct,
		ActualMarginPercent:    actMarginPct,
		VarianceAmount:         math.Round(costVariance*100) / 100,
		VariancePercent:        costVariancePct,
		FinancialStatus:        financialStatus,
		TotalChargesCount:      totalChargesCount,
		PendingChargesCount:    pendingChargesCount,
		InvoicedCarrierCount:   len(carrierInvoices),
		InvoicedCustomerCount:  len(customerInvoices),
		RFQID:                  sh.RFQID,
		QuoteID:                sh.QuoteID,
		BookingID:              sh.BookingID,
	}

	return summary
}
