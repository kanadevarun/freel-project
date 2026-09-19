package pricing_workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/freel/backend/internal/approvals"
	"github.com/freel/backend/internal/audit/domain"
	auditSvcPkg "github.com/freel/backend/internal/audit/service"
	"github.com/freel/backend/internal/orchestration"
	"github.com/freel/backend/internal/quotations"
	"github.com/freel/backend/internal/rfq"
	rfqspec "github.com/freel/backend/internal/rfq/spec"
	"github.com/jmoiron/sqlx"
)

// Service defines high-level operations for RFQ pricing and quotation workflow
type Service interface {
	GetWorkflowOverview(ctx context.Context, orgID, rfqID int64) (*WorkflowOverviewResponse, error)
	ExtractRequirements(ctx context.Context, orgID, rfqID int64) (*RFQRequirementsExtraction, error)
	CalculatePricingPreview(ctx context.Context, orgID, rfqID int64, req PricingPreviewRequest) (*PricingPreviewResponse, error)
	GenerateDraft(ctx context.Context, orgID, rfqID int64, userID int64, input GenerateDraftInput) (*QuotationDraft, error)
	GetDraft(ctx context.Context, orgID, draftID int64) (*QuotationDraft, error)
	UpdateDraft(ctx context.Context, orgID, draftID int64, input UpdateDraftInput) (*QuotationDraft, error)
	SubmitDraftForApproval(ctx context.Context, orgID, draftID int64, userID int64, input SubmitApprovalInput) (*QuotationDraft, error)
}

type service struct {
	db             *sqlx.DB
	repo           Repository
	rfqBL          rfq.BusinessLogic
	approvalsSvc   approvals.Service
	orchestration  orchestration.Service
	auditSvc       auditSvcPkg.Service
	sidecarBaseURL string
	httpClient     *http.Client
}

// NewService creates a new Service instance
func NewService(
	db *sqlx.DB,
	repo Repository,
	rfqBL rfq.BusinessLogic,
	approvalsSvc approvals.Service,
	orchSvc orchestration.Service,
	auditSvc auditSvcPkg.Service,
) Service {
	baseURL := os.Getenv("AI_SIDECAR_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8090"
	}
	return &service{
		db:             db,
		repo:           repo,
		rfqBL:          rfqBL,
		approvalsSvc:   approvalsSvc,
		orchestration:  orchSvc,
		auditSvc:       auditSvc,
		sidecarBaseURL: strings.TrimRight(baseURL, "/"),
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}
}

// round2 rounds float64 to 2 decimal places
func round2(val float64) float64 {
	return math.Round(val*100.0) / 100.0
}

// buildRFQContext constructs the sanitized context for the Python AI sidecar
func (s *service) buildRFQContext(ctx context.Context, orgID, rfqID int64, rfqObj *rfqspec.RFQ, corrID string) (map[string]interface{}, error) {
	var custName string = "Commercial Cargo Customer"
	var custTier string = "STANDARD"
	var payTerms string = "NET_30"
	var outBal float64 = 0.0

	var custRow struct {
		Name           string  `db:"name"`
		Tier           *string `db:"tier"`
		PaymentTerms   *string `db:"payment_terms"`
		CurrentBalance float64 `db:"current_balance"`
	}
	err := s.db.GetContext(ctx, &custRow, `
		SELECT name, tier, payment_terms, COALESCE(current_balance, 0) as current_balance
		FROM customers
		WHERE org_id = ? AND id = ?
	`, orgID, rfqObj.CustomerID)
	if err == nil {
		custName = custRow.Name
		if custRow.Tier != nil {
			custTier = *custRow.Tier
		}
		if custRow.PaymentTerms != nil {
			payTerms = *custRow.PaymentTerms
		}
		outBal = custRow.CurrentBalance
	}

	// Items list
	items := make([]map[string]interface{}, 0)
	for _, itm := range rfqObj.Items {
		var w float64
		if itm.WeightKG != nil {
			w = *itm.WeightKG
		}
		var v float64
		if itm.VolumeCBM != nil {
			v = *itm.VolumeCBM
		}

		items = append(items, map[string]interface{}{
			"item_id":      itm.ID,
			"description":  itm.Description,
			"quantity":     itm.Quantity,
			"weight_kg":    w,
			"volume_cbm":   v,
		})
	}

	// Carrier Quotes
	carrierQuotes := make([]map[string]interface{}, 0)
	for _, q := range rfqObj.Quotes {
		cName := q.CarrierName
		if cName == "" {
			cName = "Carrier Bid"
		}
		carrierQuotes = append(carrierQuotes, map[string]interface{}{
			"quote_id":     q.ID,
			"carrier_name": cName,
			"total_cost":   q.BuyPrice,
			"currency":     "USD",
			"expired":      q.Status == "EXPIRED",
		})
	}

	// Available Rates from rates table
	availableRates := make([]map[string]interface{}, 0)
	originVal := ""
	if rfqObj.Origin != nil {
		originVal = *rfqObj.Origin
	}
	destVal := ""
	if rfqObj.Destination != nil {
		destVal = *rfqObj.Destination
	}

	var rateRows []struct {
		ID          int64   `db:"id"`
		CarrierName string  `db:"carrier_name"`
		OriginPort  string  `db:"origin_port"`
		DestPort    string  `db:"destination_port"`
		Mode        string  `db:"mode"`
		BaseRate    float64 `db:"base_rate"`
		BunkerFuel  float64 `db:"bunker_fuel"`
		DocFee      float64 `db:"doc_fee"`
		TotalCost   float64 `db:"total_cost"`
		Currency    string  `db:"currency"`
		ExpiryDate  *string `db:"expiry_date"`
	}
	_ = s.db.SelectContext(ctx, &rateRows, `
		SELECT r.id, COALESCE(r.carrier_name, 'Carrier Tariff') as carrier_name,
		       COALESCE(r.origin_port, '') as origin_port, COALESCE(r.destination_port, '') as destination_port,
		       COALESCE(r.mode, 'OCEAN_FCL') as mode,
		       COALESCE(r.base_ocean_rate, 0) as base_rate,
		       COALESCE(r.bunker_fuel_surcharge, 0) as bunker_fuel,
		       COALESCE(r.documentation_fee, 0) as doc_fee,
		       COALESCE(r.total_calculated_rate, r.base_ocean_rate, 0) as total_cost,
		       COALESCE(r.currency, 'USD') as currency,
		       DATE_FORMAT(r.validity_end, '%Y-%m-%d') as expiry_date
		FROM rates r
		WHERE r.org_id = ? AND (r.origin_port = ? OR r.destination_port = ? OR r.origin_port = '')
		ORDER BY r.id DESC LIMIT 10
	`, orgID, originVal, destVal)

	for _, rr := range rateRows {
		availableRates = append(availableRates, map[string]interface{}{
			"rate_id":          rr.ID,
			"carrier_name":     rr.CarrierName,
			"origin_port":      rr.OriginPort,
			"destination_port": rr.DestPort,
			"mode":             rr.Mode,
			"base_ocean_rate":  rr.BaseRate,
			"total_cost":       rr.TotalCost,
			"currency":         rr.Currency,
			"expiry_date":      rr.ExpiryDate,
		})
	}

	incotermsVal := ""
	if rfqObj.Incoterms != nil {
		incotermsVal = *rfqObj.Incoterms
	}
	targetDateVal := ""
	if rfqObj.TargetDate != nil {
		targetDateVal = rfqObj.TargetDate.Format("2006-01-02")
	}

	payload := map[string]interface{}{
		"org_id":              orgID,
		"rfq_id":              rfqID,
		"rfq_number":          rfqObj.RFQNumber,
		"customer_id":         rfqObj.CustomerID,
		"customer_name":       custName,
		"customer_tier":       custTier,
		"payment_terms":       payTerms,
		"credit_limit":        50000.00,
		"outstanding_balance": outBal,
		"origin":              originVal,
		"destination":         destVal,
		"shipment_mode":       "OCEAN_FCL",
		"incoterms":           incotermsVal,
		"target_date":         targetDateVal,
		"items":               items,
		"carrier_quotes":      carrierQuotes,
		"available_rates":     availableRates,
		"correlation_id":      corrID,
	}
	return payload, nil
}

func (s *service) ExtractRequirements(ctx context.Context, orgID, rfqID int64) (*RFQRequirementsExtraction, error) {
	rfqObj, err := s.rfqBL.GetRFQ(ctx, int32(orgID), int32(rfqID))
	if err != nil || rfqObj == nil {
		return nil, fmt.Errorf("rfq not found or access denied: %w", err)
	}

	corrID := fmt.Sprintf("corr-req-extract-%d-%d", rfqID, time.Now().Unix())
	rfqCtx, err := s.buildRFQContext(ctx, orgID, rfqID, rfqObj, corrID)
	if err != nil {
		return nil, err
	}

	bodyBytes, _ := json.Marshal(rfqCtx)
	url := fmt.Sprintf("%s/rfq-pricing/extract-requirements", s.sidecarBaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create sidecar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LogisticsHQ-Service-Key", getInternalServiceToken())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sidecar extraction call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar extraction returned status %d", resp.StatusCode)
	}

	var sidecarRes struct {
		RFQID                       int64                  `json:"rfq_id"`
		Status                      string                 `json:"status"`
		Fields                      map[string]interface{} `json:"fields"`
		MissingMandatory            []string               `json:"missing_mandatory"`
		MissingOptional             []string               `json:"missing_optional"`
		ClarificationRecommendation *string                `json:"clarification_recommendation"`
		ConfidenceScore             float64                `json:"confidence_score"`
		CorrelationID               string                 `json:"correlation_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&sidecarRes); err != nil {
		return nil, fmt.Errorf("failed to decode sidecar extraction output: %w", err)
	}

	extData, _ := json.Marshal(sidecarRes.Fields)
	missData, _ := json.Marshal(map[string]interface{}{
		"mandatory":      sidecarRes.MissingMandatory,
		"optional":       sidecarRes.MissingOptional,
		"recommendation": sidecarRes.ClarificationRecommendation,
	})

	extraction := &RFQRequirementsExtraction{
		OrgID:               orgID,
		RFQID:               rfqID,
		Status:              sidecarRes.Status,
		ExtractedData:       extData,
		MissingFields:       missData,
		ClarificationNeeded: len(sidecarRes.MissingMandatory) > 0,
		ConfidenceScore:     sidecarRes.ConfidenceScore,
		CorrelationID:       corrID,
	}

	if err := s.repo.SaveExtraction(ctx, extraction); err != nil {
		return nil, err
	}

	return extraction, nil
}

func (s *service) CalculatePricingPreview(ctx context.Context, orgID, rfqID int64, req PricingPreviewRequest) (*PricingPreviewResponse, error) {
	rfqObj, err := s.rfqBL.GetRFQ(ctx, int32(orgID), int32(rfqID))
	if err != nil || rfqObj == nil {
		return nil, fmt.Errorf("rfq not found: %w", err)
	}

	corrID := fmt.Sprintf("corr-pricing-calc-%d-%d", rfqID, time.Now().Unix())

	var baseCost float64 = 2200.00
	var surcharges float64 = 350.00
	var carrierName string = "Standard Ocean Carrier"
	var appliedRateID *int64
	var rateExpired bool = false

	if req.RateID != nil && *req.RateID > 0 {
		var rRow struct {
			ID          int64   `db:"id"`
			CarrierName string  `db:"carrier_name"`
			BaseRate    float64 `db:"base_rate"`
			BunkerFuel  float64 `db:"bunker_fuel"`
			DocFee      float64 `db:"doc_fee"`
			Expired     bool    `db:"expired"`
		}
		err := s.db.GetContext(ctx, &rRow, `
			SELECT id, COALESCE(carrier_name, 'Carrier Tariff') as carrier_name,
			       COALESCE(base_ocean_rate, 2000) as base_rate,
			       COALESCE(bunker_fuel_surcharge, 250) as bunker_fuel,
			       COALESCE(documentation_fee, 100) as doc_fee,
			       CASE WHEN validity_end IS NOT NULL AND validity_end < CURDATE() THEN 1 ELSE 0 END as expired
			FROM rates WHERE org_id = ? AND id = ?
		`, orgID, *req.RateID)
		if err == nil {
			appliedRateID = &rRow.ID
			carrierName = rRow.CarrierName
			baseCost = rRow.BaseRate
			surcharges = rRow.BunkerFuel + rRow.DocFee
			rateExpired = rRow.Expired
		}
	} else if len(rfqObj.Quotes) > 0 {
		q := rfqObj.Quotes[0]
		if q.BuyPrice > 0 {
			baseCost = round2(q.BuyPrice * 0.85)
			surcharges = round2(q.BuyPrice * 0.15)
		}
		if q.CarrierName != "" {
			carrierName = q.CarrierName
		}
		rateExpired = q.Status == "EXPIRED"
	}

	totalCost := round2(baseCost + surcharges)

	targetMargin := 0.20
	if req.TargetMarginPct != nil && *req.TargetMarginPct >= -50.0 && *req.TargetMarginPct <= 100.0 {
		targetMargin = *req.TargetMarginPct / 100.0
	}

	baseSell := round2(baseCost * (1.0 + targetMargin))
	discount := 0.0
	if req.DiscountAmount != nil && *req.DiscountAmount > 0 {
		discount = round2(*req.DiscountAmount)
	}

	totalSellingPrice := round2(baseSell + surcharges - discount)
	if totalSellingPrice < 0 {
		totalSellingPrice = 0.0
	}

	grossProfit := round2(totalSellingPrice - totalCost)
	var grossMarginPct float64 = 0.0
	if totalSellingPrice > 0 {
		grossMarginPct = round2((grossProfit / totalSellingPrice) * 100.0)
	}

	marginHealth := quotations.MarginHealthHealthy
	if grossProfit < 0 {
		marginHealth = quotations.MarginHealthNegative
	} else if grossMarginPct < 15.0 {
		marginHealth = quotations.MarginHealthLow
	}

	costBreakdown, _ := json.Marshal(map[string]interface{}{
		"base_ocean_freight": baseCost,
		"surcharges":         surcharges,
		"total_carrier_cost": totalCost,
		"carrier_name":       carrierName,
	})

	sellBreakdown, _ := json.Marshal(map[string]interface{}{
		"base_sell":            baseSell,
		"ancillary_surcharges": surcharges,
		"discounts":            discount,
		"total_selling_price":  totalSellingPrice,
		"gross_margin_pct":     grossMarginPct,
	})

	return &PricingPreviewResponse{
		Currency:           "USD",
		BaseCost:           baseCost,
		Surcharges:         surcharges,
		TotalCost:          totalCost,
		BaseSell:           baseSell,
		Discounts:          discount,
		TaxAmount:          0.00,
		TotalSellingPrice:  totalSellingPrice,
		GrossProfit:        grossProfit,
		GrossMarginPct:     grossMarginPct,
		MarginHealth:       marginHealth,
		AppliedRateID:      appliedRateID,
		AppliedCarrierName: carrierName,
		RateIsExpired:      rateExpired,
		CostBreakdown:      costBreakdown,
		SellBreakdown:      sellBreakdown,
		CorrelationID:      corrID,
	}, nil
}

func (s *service) GenerateDraft(ctx context.Context, orgID, rfqID int64, userID int64, input GenerateDraftInput) (*QuotationDraft, error) {
	rfqObj, err := s.rfqBL.GetRFQ(ctx, int32(orgID), int32(rfqID))
	if err != nil || rfqObj == nil {
		return nil, fmt.Errorf("rfq not found: %w", err)
	}

	prevReq := PricingPreviewRequest{
		RateID:         input.RateID,
		DiscountAmount: input.DiscountAmount,
	}
	pricing, err := s.CalculatePricingPreview(ctx, orgID, rfqID, prevReq)
	if err != nil {
		return nil, err
	}

	if input.BaseSellOverride != nil && *input.BaseSellOverride > 0 {
		pricing.BaseSell = round2(*input.BaseSellOverride)
		pricing.TotalSellingPrice = round2(pricing.BaseSell + pricing.Surcharges - pricing.Discounts)
		pricing.GrossProfit = round2(pricing.TotalSellingPrice - pricing.TotalCost)
		if pricing.TotalSellingPrice > 0 {
			pricing.GrossMarginPct = round2((pricing.GrossProfit / pricing.TotalSellingPrice) * 100.0)
		}
		if pricing.GrossProfit < 0 {
			pricing.MarginHealth = quotations.MarginHealthNegative
		} else if pricing.GrossMarginPct < 15.0 {
			pricing.MarginHealth = quotations.MarginHealthLow
		} else {
			pricing.MarginHealth = quotations.MarginHealthHealthy
		}
	}

	corrID := fmt.Sprintf("corr-draft-%d-%d", rfqID, time.Now().Unix())

	custName := "Valued Customer"
	contactEmail := input.RecipientEmail
	contactName := input.RecipientName
	if contactEmail == "" || contactName == "" {
		var cRow struct {
			Name         string  `db:"name"`
			ContactName  *string `db:"primary_contact_name"`
			ContactEmail *string `db:"primary_contact_email"`
		}
		if err := s.db.GetContext(ctx, &cRow, `SELECT name, primary_contact_name, primary_contact_email FROM customers WHERE org_id = ? AND id = ?`, orgID, rfqObj.CustomerID); err == nil {
			custName = cRow.Name
			if contactName == "" && cRow.ContactName != nil {
				contactName = *cRow.ContactName
			}
			if contactEmail == "" && cRow.ContactEmail != nil {
				contactEmail = *cRow.ContactEmail
			}
		}
	}

	draftReqPayload := map[string]interface{}{
		"rfq_id":            rfqID,
		"rfq_number":        rfqObj.RFQNumber,
		"customer_id":       rfqObj.CustomerID,
		"customer_name":     custName,
		"contact_name":      contactName,
		"contact_email":     contactEmail,
		"origin":            "POL",
		"destination":       "POD",
		"shipment_mode":     "OCEAN_FCL",
		"incoterms":         "FOB",
		"items_summary":     "Commercial Cargo",
		"validity_days":     input.ValidityDays,
		"user_prompt_notes": input.UserPromptNotes,
		"pricing": map[string]interface{}{
			"currency":             pricing.Currency,
			"base_cost":            pricing.BaseCost,
			"total_cost":           pricing.TotalCost,
			"base_sell":            pricing.BaseSell,
			"surcharges":           pricing.Surcharges,
			"discounts":            pricing.Discounts,
			"tax_amount":           pricing.TaxAmount,
			"total_selling_price":  pricing.TotalSellingPrice,
			"gross_profit":         pricing.GrossProfit,
			"gross_margin_pct":     pricing.GrossMarginPct,
			"margin_health":        pricing.MarginHealth,
			"applied_carrier_name": pricing.AppliedCarrierName,
			"rate_is_expired":      pricing.RateIsExpired,
		},
		"correlation_id": corrID,
	}
	if rfqObj.Origin != nil {
		draftReqPayload["origin"] = *rfqObj.Origin
	}
	if rfqObj.Destination != nil {
		draftReqPayload["destination"] = *rfqObj.Destination
	}
	if rfqObj.Incoterms != nil {
		draftReqPayload["incoterms"] = *rfqObj.Incoterms
	}

	bodyBytes, _ := json.Marshal(draftReqPayload)
	url := fmt.Sprintf("%s/rfq-pricing/generate-draft", s.sidecarBaseURL)
	hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create sidecar draft request: %w", err)
	}
	hReq.Header.Set("Content-Type", "application/json")
	hReq.Header.Set("X-LogisticsHQ-Service-Key", getInternalServiceToken())

	hResp, err := s.httpClient.Do(hReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar draft call failed: %w", err)
	}
	defer hResp.Body.Close()

	var sidecarDraft struct {
		InternalSummary    string `json:"internal_summary"`
		CustomerWording    string `json:"customer_wording"`
		PricingExplanation string `json:"pricing_explanation"`
		TermsAndConditions string `json:"terms_and_conditions"`
		RequiresApproval   bool   `json:"requires_approval"`
	}
	if err := json.NewDecoder(hResp.Body).Decode(&sidecarDraft); err != nil {
		return nil, fmt.Errorf("failed to decode sidecar draft response: %w", err)
	}

	now := time.Now()
	valEnd := now.AddDate(0, 0, input.ValidityDays)
	if input.ValidityDays <= 0 {
		valEnd = now.AddDate(0, 0, 14)
	}

	requiresApproval := pricing.MarginHealth != quotations.MarginHealthHealthy || pricing.RateIsExpired || pricing.Discounts > 0

	rateRefs, _ := json.Marshal(map[string]interface{}{
		"rate_id":      pricing.AppliedRateID,
		"carrier_name": pricing.AppliedCarrierName,
		"expired":      pricing.RateIsExpired,
	})

	draft := &QuotationDraft{
		OrgID:              orgID,
		RFQID:              rfqID,
		Status:             "DRAFT",
		Currency:           pricing.Currency,
		BaseCost:           pricing.BaseCost,
		TotalCost:          pricing.TotalCost,
		BaseSell:           pricing.BaseSell,
		Surcharges:         pricing.Surcharges,
		Discounts:          pricing.Discounts,
		TaxAmount:          pricing.TaxAmount,
		TotalSellingPrice:  pricing.TotalSellingPrice,
		GrossMarginAmount:  pricing.GrossProfit,
		GrossMarginPct:     pricing.GrossMarginPct,
		MarginHealth:       pricing.MarginHealth,
		RateReferences:     rateRefs,
		CostComponents:     pricing.CostBreakdown,
		SellingComponents:  pricing.SellBreakdown,
		TermsAndConditions: sidecarDraft.TermsAndConditions,
		InternalSummary:    sidecarDraft.InternalSummary,
		CustomerWording:    sidecarDraft.CustomerWording,
		PricingExplanation: sidecarDraft.PricingExplanation,
		RecipientEmail:     contactEmail,
		RecipientName:      contactName,
		ValidityStart:      &now,
		ValidityEnd:        &valEnd,
		RequiresApproval:   requiresApproval,
		CreatedByUserID:    &userID,
		CorrelationID:      corrID,
	}

	if err := s.repo.SaveDraft(ctx, draft); err != nil {
		return nil, err
	}

	return draft, nil
}

func (s *service) GetDraft(ctx context.Context, orgID, draftID int64) (*QuotationDraft, error) {
	return s.repo.GetDraftByID(ctx, orgID, draftID)
}

func (s *service) UpdateDraft(ctx context.Context, orgID, draftID int64, input UpdateDraftInput) (*QuotationDraft, error) {
	draft, err := s.repo.GetDraftByID(ctx, orgID, draftID)
	if err != nil || draft == nil {
		return nil, fmt.Errorf("draft not found: %w", err)
	}

	if draft.Status == "APPROVED" || draft.Status == "EXECUTED" {
		return nil, fmt.Errorf("cannot modify draft in '%s' status", draft.Status)
	}

	draft.CustomerWording = input.CustomerWording
	draft.InternalSummary = input.InternalSummary
	draft.TermsAndConditions = input.TermsAndConditions
	draft.RecipientName = input.RecipientName
	draft.RecipientEmail = input.RecipientEmail

	if input.BaseSell != nil && *input.BaseSell > 0 {
		draft.BaseSell = round2(*input.BaseSell)
	}
	if input.Discounts != nil {
		draft.Discounts = round2(*input.Discounts)
	}
	draft.TotalSellingPrice = round2(draft.BaseSell + draft.Surcharges - draft.Discounts)
	draft.GrossMarginAmount = round2(draft.TotalSellingPrice - draft.TotalCost)
	if draft.TotalSellingPrice > 0 {
		draft.GrossMarginPct = round2((draft.GrossMarginAmount / draft.TotalSellingPrice) * 100.0)
	}
	if draft.GrossMarginAmount < 0 {
		draft.MarginHealth = quotations.MarginHealthNegative
	} else if draft.GrossMarginPct < 15.0 {
		draft.MarginHealth = quotations.MarginHealthLow
	} else {
		draft.MarginHealth = quotations.MarginHealthHealthy
	}

	draft.RequiresApproval = draft.MarginHealth != quotations.MarginHealthHealthy || draft.Discounts > 0

	if err := s.repo.UpdateDraft(ctx, draft); err != nil {
		return nil, err
	}

	return draft, nil
}

func (s *service) SubmitDraftForApproval(ctx context.Context, orgID, draftID int64, userID int64, input SubmitApprovalInput) (*QuotationDraft, error) {
	draft, err := s.repo.GetDraftByID(ctx, orgID, draftID)
	if err != nil || draft == nil {
		return nil, fmt.Errorf("draft not found: %w", err)
	}

	if draft.Status == "PENDING_APPROVAL" {
		return draft, nil
	}

	title := fmt.Sprintf("Quotation Draft Approval for RFQ #%d ($%.2f, Margin: %.2f%%)", draft.RFQID, draft.TotalSellingPrice, draft.GrossMarginPct)
	desc := fmt.Sprintf("Commercial proposal submitted for review. Total: %s %.2f. Cost: %s %.2f. Gross Margin: %.2f%% (%s). Reason: %s",
		draft.Currency, draft.TotalSellingPrice, draft.Currency, draft.TotalCost, draft.GrossMarginPct, draft.MarginHealth, input.Reason)

	riskLevel := "MEDIUM"
	if draft.MarginHealth == quotations.MarginHealthNegative {
		riskLevel = "CRITICAL"
	} else if draft.MarginHealth == quotations.MarginHealthLow {
		riskLevel = "HIGH"
	}

	actionName := "quotations.send_draft"
	payloadMap := map[string]interface{}{
		"draft_id":            draft.ID,
		"rfq_id":              draft.RFQID,
		"recipient_name":      draft.RecipientName,
		"recipient_email":     draft.RecipientEmail,
		"total_selling_price": draft.TotalSellingPrice,
		"currency":            draft.Currency,
		"margin_health":       draft.MarginHealth,
		"gross_margin_pct":    draft.GrossMarginPct,
	}
	payloadBytes, _ := json.Marshal(payloadMap)

	var approvalID *int64
	if s.approvalsSvc != nil {
		approvalReq, err := s.approvalsSvc.CreateApproval(ctx, orgID, &approvals.CreateApprovalInput{
			Title:             title,
			Category:          "COMMERCIAL",
			Type:              "Quotation Approval",
			Priority:          riskLevel,
			RelatedRef:        fmt.Sprintf("DRAFT-%d", draft.ID),
			RelatedEntityType: "QUOTATION_DRAFT",
			RelatedEntityID:   draft.ID,
			CustomerName:      draft.RecipientName,
			RequestedByID:     userID,
			Description:       desc,
			RiskLevel:         riskLevel,
			ActionName:        actionName,
			ProposedPayload:   string(payloadBytes),
			CorrelationID:     draft.CorrelationID,
		}, "Operations Specialist")
		if err == nil && approvalReq != nil {
			approvalID = &approvalReq.ID
		}
	}

	proposalID := fmt.Sprintf("prop-draft-%d-%d", draft.ID, time.Now().Unix())
	newStatus := "PENDING_APPROVAL"

	if err := s.repo.UpdateDraftStatus(ctx, orgID, draftID, newStatus, approvalID, &proposalID); err != nil {
		return nil, err
	}

	draft.Status = newStatus
	draft.ApprovalID = approvalID
	draft.ActionProposalID = &proposalID

	if s.auditSvc != nil {
		s.auditSvc.RecordAsync(ctx, domain.CreateAuditLogParams{
			OrgID:        orgID,
			ActorID:      &userID,
			ActorType:    domain.ActorTypeUser,
			Action:       domain.ActionUpdate,
			Module:       domain.ModuleQuotations,
			ResourceType: "quotation_drafts",
			ResourceID:   fmt.Sprintf("%d", draftID),
			Description:  fmt.Sprintf("Submitted quotation draft #%d for managerial approval. Margin: %.2f%%", draftID, draft.GrossMarginPct),
		})
	}

	return draft, nil
}

func (s *service) GetWorkflowOverview(ctx context.Context, orgID, rfqID int64) (*WorkflowOverviewResponse, error) {
	rfqObj, err := s.rfqBL.GetRFQ(ctx, int32(orgID), int32(rfqID))
	if err != nil || rfqObj == nil {
		return nil, fmt.Errorf("rfq not found: %w", err)
	}

	corrID := fmt.Sprintf("corr-overview-%d-%d", rfqID, time.Now().Unix())

	extraction, _ := s.repo.GetLatestExtractionByRFQ(ctx, orgID, rfqID)
	draft, _ := s.repo.GetLatestDraftByRFQ(ctx, orgID, rfqID)

	prevReq := PricingPreviewRequest{}
	if draft != nil && draft.RateReferences != nil {
		var rr struct {
			RateID *int64 `json:"rate_id"`
		}
		if err := json.Unmarshal(draft.RateReferences, &rr); err == nil && rr.RateID != nil {
			prevReq.RateID = rr.RateID
		}
	}
	pricingPrev, _ := s.CalculatePricingPreview(ctx, orgID, rfqID, prevReq)

	rfqCtx, _ := s.buildRFQContext(ctx, orgID, rfqID, rfqObj, corrID)

	var aiExplRaw json.RawMessage
	var riskRaw json.RawMessage

	if pricingPrev != nil {
		explPayload, _ := json.Marshal(map[string]interface{}{
			"context": rfqCtx,
			"pricing": map[string]interface{}{
				"currency":             pricingPrev.Currency,
				"base_cost":            pricingPrev.BaseCost,
				"total_cost":           pricingPrev.TotalCost,
				"base_sell":            pricingPrev.BaseSell,
				"surcharges":           pricingPrev.Surcharges,
				"discounts":            pricingPrev.Discounts,
				"total_selling_price":  pricingPrev.TotalSellingPrice,
				"gross_profit":         pricingPrev.GrossProfit,
				"gross_margin_pct":     pricingPrev.GrossMarginPct,
				"margin_health":        pricingPrev.MarginHealth,
				"applied_carrier_name": pricingPrev.AppliedCarrierName,
				"rate_is_expired":      pricingPrev.RateIsExpired,
			},
		})

		urlExpl := fmt.Sprintf("%s/rfq-pricing/explain-pricing", s.sidecarBaseURL)
		if hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, urlExpl, bytes.NewReader(explPayload)); err == nil {
			hReq.Header.Set("Content-Type", "application/json")
			hReq.Header.Set("X-LogisticsHQ-Service-Key", getInternalServiceToken())
			if hResp, err := s.httpClient.Do(hReq); err == nil && hResp.StatusCode == http.StatusOK {
				defer hResp.Body.Close()
				var parsedExpl map[string]interface{}
				if err := json.NewDecoder(hResp.Body).Decode(&parsedExpl); err == nil {
					aiExplRaw, _ = json.Marshal(parsedExpl)
				}
			}
		}

		urlRisk := fmt.Sprintf("%s/rfq-pricing/analyze-risks", s.sidecarBaseURL)
		if hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, urlRisk, bytes.NewReader(explPayload)); err == nil {
			hReq.Header.Set("Content-Type", "application/json")
			hReq.Header.Set("X-LogisticsHQ-Service-Key", getInternalServiceToken())
			if hResp, err := s.httpClient.Do(hReq); err == nil && hResp.StatusCode == http.StatusOK {
				defer hResp.Body.Close()
				var parsedRisk map[string]interface{}
				if err := json.NewDecoder(hResp.Body).Decode(&parsedRisk); err == nil {
					riskRaw, _ = json.Marshal(parsedRisk)
				}
			}
		}
	}

	custName := "Commercial Cargo Customer"
	var cName struct {
		Name string `db:"name"`
	}
	if err := s.db.GetContext(ctx, &cName, `SELECT name FROM customers WHERE org_id = ? AND id = ?`, orgID, rfqObj.CustomerID); err == nil {
		custName = cName.Name
	}

	originVal := ""
	if rfqObj.Origin != nil {
		originVal = *rfqObj.Origin
	}
	destVal := ""
	if rfqObj.Destination != nil {
		destVal = *rfqObj.Destination
	}
	incotermsVal := ""
	if rfqObj.Incoterms != nil {
		incotermsVal = *rfqObj.Incoterms
	}

	var approvalID *int64
	pendingApproval := false
	if draft != nil {
		approvalID = draft.ApprovalID
		pendingApproval = draft.Status == "PENDING_APPROVAL"
	}

	return &WorkflowOverviewResponse{
		RFQID:               rfqID,
		RFQNumber:           rfqObj.RFQNumber,
		CustomerName:        custName,
		Origin:              originVal,
		Destination:         destVal,
		ShipmentMode:        "OCEAN_FCL",
		Incoterms:           incotermsVal,
		Extraction:          extraction,
		LatestDraft:         draft,
		PricingPreview:      pricingPrev,
		AIExplanation:       aiExplRaw,
		RiskAnalysis:        riskRaw,
		PendingApproval:     pendingApproval,
		ApprovalID:          approvalID,
		AvailableRatesCount: 5,
		CorrelationID:       corrID,
	}, nil
}

func getInternalServiceToken() string {
	tok := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if tok == "" {
		tok = os.Getenv("INTERNAL_SERVICE_KEY")
	}
	if tok == "" {
		tok = "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456"
	}
	return tok
}
