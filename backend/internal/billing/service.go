package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	GenerateInvoiceFromShipment(ctx context.Context, orgID int64, shipmentID int64) (*CustomerInvoice, error)
	GetInvoicesByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*CustomerInvoice, error)
	GetInvoiceByID(ctx context.Context, orgID int64, id string) (*CustomerInvoice, []*CustomerInvoiceItem, error)
	ApproveInvoice(ctx context.Context, orgID int64, id string) error
	PayInvoice(ctx context.Context, orgID int64, id string) error
	RecalculateProfitability(ctx context.Context, orgID int64, shipmentID int64) (*Profitability, error)
	GetProfitability(ctx context.Context, orgID int64, shipmentID int64) (*Profitability, error)
	AuditClosure(ctx context.Context, orgID int64, shipmentID int64) (*ShipmentClosureAudit, error)
	CloseShipment(ctx context.Context, orgID int64, shipmentID int64) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GenerateInvoiceFromShipment(ctx context.Context, orgID int64, shipmentID int64) (*CustomerInvoice, error) {
	// 1. Fetch shipment's accepted quote pricing context
	buyPrice, sellPrice, surchargesJSON, oceanBuyPrice, err := s.repo.GetShipmentQuoteAndRateEntry(ctx, orgID, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quote for invoice generation: %w", err)
	}

	// 2. Compute overall sell-to-buy markup ratio (default to 1.20 if buyPrice is zero)
	markup := 1.20
	if buyPrice > 0 {
		markup = sellPrice / buyPrice
	}

	invoiceUUID := uuid.New().String()
	invoiceNumber := fmt.Sprintf("CINV-%d-%d", shipmentID, time.Now().Unix())

	var items []*CustomerInvoiceItem

	// Try parsing surcharges breakdown from contract entries
	type SurchargeItem struct {
		Code   string  `json:"code"`
		Amount float64 `json:"amount"`
	}
	var surcharges []SurchargeItem
	err = json.Unmarshal([]byte(surchargesJSON), &surcharges)

	// If breakdown exists, markup each component proportionally
	if err == nil && len(surcharges) > 0 {
		// Ocean freight base cost
		baseSell := oceanBuyPrice * markup
		if baseSell <= 0 {
			// If base ocean freight buy rate is missing in entry, calculate by subtracting surcharges total from total buy
			surchTotal := 0.0
			for _, surch := range surcharges {
				surchTotal += surch.Amount
			}
			baseSell = (buyPrice - surchTotal) * markup
		}
		if baseSell > 0 {
			items = append(items, &CustomerInvoiceItem{
				OrgID:       orgID,
				InvoiceID:   invoiceUUID,
				ChargeCode:  "OCEAN_FREIGHT",
				Description: "Ocean Base Freight Charge",
				Quantity:    1.0,
				UnitPrice:   baseSell,
				TotalAmount: baseSell,
				Currency:    "USD",
			})
		}

		// Surcharges
		for _, surch := range surcharges {
			surchSell := surch.Amount * markup
			items = append(items, &CustomerInvoiceItem{
				OrgID:       orgID,
				InvoiceID:   invoiceUUID,
				ChargeCode:  surch.Code,
				Description: fmt.Sprintf("Surcharge: %s", surch.Code),
				Quantity:    1.0,
				UnitPrice:   surchSell,
				TotalAmount: surchSell,
				Currency:    "USD",
			})
		}
	} else {
		// Fallback to single ocean freight item if no breakdown is available
		items = append(items, &CustomerInvoiceItem{
			OrgID:       orgID,
			InvoiceID:   invoiceUUID,
			ChargeCode:  "OCEAN_FREIGHT",
			Description: "Ocean Freight Charges (Flat Quoted)",
			Quantity:    1.0,
			UnitPrice:   sellPrice,
			TotalAmount: sellPrice,
			Currency:    "USD",
		})
	}

	// Calculate total amount
	total := 0.0
	for _, item := range items {
		total += item.TotalAmount
	}

	dueDate := time.Now().AddDate(0, 0, 30) // Net 30 terms

	invoice := &CustomerInvoice{
		ID:            invoiceUUID,
		OrgID:         orgID,
		ShipmentID:    shipmentID,
		InvoiceNumber: invoiceNumber,
		Status:        "DRAFT",
		DueDate:       &dueDate,
		Currency:      "USD",
		TotalAmount:   total,
	}

	err = s.repo.CreateInvoice(ctx, invoice, items)
	if err != nil {
		return nil, fmt.Errorf("failed to save generated invoice: %w", err)
	}

	// 3. Recalculate profitability
	_, _ = s.RecalculateProfitability(ctx, orgID, shipmentID)

	return invoice, nil
}

func (s *service) GetInvoicesByShipment(ctx context.Context, orgID int64, shipmentID int64) ([]*CustomerInvoice, error) {
	return s.repo.GetInvoicesByShipment(ctx, orgID, shipmentID)
}

func (s *service) GetInvoiceByID(ctx context.Context, orgID int64, id string) (*CustomerInvoice, []*CustomerInvoiceItem, error) {
	return s.repo.GetInvoiceByID(ctx, orgID, id)
}

func (s *service) ApproveInvoice(ctx context.Context, orgID int64, id string) error {
	err := s.repo.UpdateInvoiceStatus(ctx, orgID, id, "APPROVED")
	if err != nil {
		return err
	}
	invoice, _, err := s.repo.GetInvoiceByID(ctx, orgID, id)
	if err == nil && invoice != nil {
		_, _ = s.RecalculateProfitability(ctx, orgID, invoice.ShipmentID)
	}
	return nil
}

func (s *service) PayInvoice(ctx context.Context, orgID int64, id string) error {
	err := s.repo.UpdateInvoiceStatus(ctx, orgID, id, "PAID")
	if err != nil {
		return err
	}
	invoice, _, err := s.repo.GetInvoiceByID(ctx, orgID, id)
	if err == nil && invoice != nil {
		_, _ = s.RecalculateProfitability(ctx, orgID, invoice.ShipmentID)
	}
	return nil
}

func (s *service) RecalculateProfitability(ctx context.Context, orgID int64, shipmentID int64) (*Profitability, error) {
	// Fetch expected buy and sell totals from quote record
	quotedBuy, quotedSell, _, _, err := s.repo.GetShipmentQuoteAndRateEntry(ctx, orgID, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("recalculate profitability: %w", err)
	}

	// Fetch actual costs and revenues
	actualCarrierCost, err := s.repo.GetApprovedCarrierInvoiceTotal(ctx, orgID, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("recalculate actual carrier cost: %w", err)
	}

	actualRevenue, err := s.repo.GetApprovedCustomerInvoiceTotal(ctx, orgID, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("recalculate actual revenue: %w", err)
	}

	expectedProfit := quotedSell - quotedBuy
	expectedMargin := 0.0
	if quotedSell > 0 {
		expectedMargin = (expectedProfit / quotedSell) * 100.0
	}

	actualProfit := actualRevenue - actualCarrierCost
	actualMargin := 0.0
	if actualRevenue > 0 {
		actualMargin = (actualProfit / actualRevenue) * 100.0
	}

	variance := actualProfit - expectedProfit

	status := "PENDING"
	if actualRevenue > 0 {
		if actualProfit < 0 {
			status = "NEGATIVE_MARGIN"
		} else if actualProfit >= expectedProfit {
			status = "ON_TARGET"
		} else {
			status = "UNDER_TARGET"
		}
	}

	profitability := &Profitability{
		OrgID:               orgID,
		ShipmentID:          shipmentID,
		QuotedSellPrice:     quotedSell,
		QuotedBuyPrice:      quotedBuy,
		ActualCarrierCost:   actualCarrierCost,
		AdditionalCharges:   0.0, // Extensions can add extra costs here
		ActualRevenue:       actualRevenue,
		ExpectedProfit:      expectedProfit,
		ActualProfit:        actualProfit,
		ExpectedMarginPct:   expectedMargin,
		ActualMarginPct:     actualMargin,
		Variance:            variance,
		ProfitabilityStatus: status,
	}

	err = s.repo.SaveProfitability(ctx, profitability)
	if err != nil {
		return nil, fmt.Errorf("failed to save profitability calculations: %w", err)
	}

	return profitability, nil
}

func (s *service) GetProfitability(ctx context.Context, orgID int64, shipmentID int64) (*Profitability, error) {
	p, err := s.repo.GetProfitability(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		// Fallback to run a default recalculate if record doesn't exist yet
		return s.RecalculateProfitability(ctx, orgID, shipmentID)
	}
	return p, nil
}

func (s *service) AuditClosure(ctx context.Context, orgID int64, shipmentID int64) (*ShipmentClosureAudit, error) {
	status, docCount, unverifiedDocs, carrierInvCount, unapprovedCarrierInvs, err := s.repo.GetClosureChecksData(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}

	// Load customer invoices to check if generated and approved
	custInvoices, err := s.repo.GetInvoicesByShipment(ctx, orgID, shipmentID)
	if err != nil {
		return nil, err
	}

	var checks []ClosureCheckResult
	ready := true

	// Rule 1: Shipment status must be DELIVERED
	r1Passed := status == "DELIVERED"
	r1Desc := "Shipment cargo must be operationally delivered (status DELIVERED)."
	if !r1Passed {
		ready = false
		r1Desc += fmt.Sprintf(" Current status is '%s'.", status)
	}
	checks = append(checks, ClosureCheckResult{
		RuleName:    "Operational Delivery Status Check",
		Passed:      r1Passed,
		Description: r1Desc,
	})

	// Rule 2: At least one document must exist, and all compliance documents must be VERIFIED
	r2Passed := docCount > 0 && unverifiedDocs == 0
	r2Desc := "All uploaded shipping documents must be successfully verified."
	if docCount == 0 {
		ready = false
		r2Desc += " Gaps: No compliance documents have been uploaded."
	} else if unverifiedDocs > 0 {
		ready = false
		r2Desc += fmt.Sprintf(" Gaps: %d document(s) still pending verification or contain discrepancy issues.", unverifiedDocs)
	}
	checks = append(checks, ClosureCheckResult{
		RuleName:    "Document Compliance Verification Check",
		Passed:      r2Passed,
		Description: r2Desc,
	})

	// Rule 3: All carrier invoices must be APPROVED
	r3Passed := carrierInvCount > 0 && unapprovedCarrierInvs == 0
	r3Desc := "All received carrier invoices must be audited and approved."
	if carrierInvCount == 0 {
		ready = false
		r3Desc += " Gaps: No carrier invoices have been ingested."
	} else if unapprovedCarrierInvs > 0 {
		ready = false
		r3Desc += fmt.Sprintf(" Gaps: %d carrier invoice(s) pending reconciliation or contain variance disputes.", unapprovedCarrierInvs)
	}
	checks = append(checks, ClosureCheckResult{
		RuleName:    "Carrier Invoice Reconciliation Check",
		Passed:      r3Passed,
		Description: r3Desc,
	})

	// Rule 4: At least one customer invoice must be APPROVED or PAID
	r4Passed := false
	approvedCustInvs := 0
	for _, ci := range custInvoices {
		if ci.Status == "APPROVED" || ci.Status == "SENT" || ci.Status == "PAID" {
			r4Passed = true
			approvedCustInvs++
		}
	}
	r4Desc := "Customer invoices must be generated and approved for billing."
	if len(custInvoices) == 0 {
		ready = false
		r4Desc += " Gaps: No customer invoice generated."
	} else if !r4Passed {
		ready = false
		r4Desc += " Gaps: Customer invoice exists but is still in DRAFT status."
	}
	checks = append(checks, ClosureCheckResult{
		RuleName:    "Customer Billing Approval Check",
		Passed:      r4Passed,
		Description: r4Desc,
	})

	return &ShipmentClosureAudit{
		ShipmentID: shipmentID,
		Ready:      ready,
		Checks:     checks,
	}, nil
}

func (s *service) CloseShipment(ctx context.Context, orgID int64, shipmentID int64) error {
	audit, err := s.AuditClosure(ctx, orgID, shipmentID)
	if err != nil {
		return err
	}
	if !audit.Ready {
		return errors.New("cannot close shipment: mandatory operational or financial closure checks failed")
	}

	return s.repo.UpdateShipmentStatus(ctx, orgID, shipmentID, "CLOSED")
}
