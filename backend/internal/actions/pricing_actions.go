package actions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/freel/backend/internal/rbac"
	"github.com/freel/backend/internal/rfq"
	rfqspec "github.com/freel/backend/internal/rfq/spec"
)

// ── pricing.save_draft_quotes ─────────────────────────────────────────────────

type QuoteDraftItem struct {
	CarrierName           string  `json:"carrier_name"`
	TransitTimeDays       int     `json:"transit_time_days"`
	BuyPrice              float64 `json:"buy_price"`
	SellPrice             float64 `json:"sell_price"`
	IsRecommended         bool    `json:"is_recommended"`
	ReliabilityScore      int     `json:"reliability_score"`
	HistoricalSuccessRate float64 `json:"historical_success_rate"`
	AiReasoning           string  `json:"ai_reasoning"`
}

type SaveDraftQuotesInput struct {
	RFQID  int64            `json:"rfq_id"`
	Quotes []QuoteDraftItem `json:"quotes"`
}

type SaveDraftQuotesAction struct {
	rfqBL rfq.BusinessLogic
}

func NewSaveDraftQuotesAction(rfqBL rfq.BusinessLogic) *SaveDraftQuotesAction {
	return &SaveDraftQuotesAction{rfqBL: rfqBL}
}

func (a *SaveDraftQuotesAction) Name() string                         { return "pricing.save_draft_quotes" }
func (a *SaveDraftQuotesAction) Module() string                       { return "pricing" }
func (a *SaveDraftQuotesAction) Description() string                  { return "Save draft quotation options recommended for an RFQ." }
func (a *SaveDraftQuotesAction) Category() ActionCategory             { return ActionCategoryHighRisk }
func (a *SaveDraftQuotesAction) InputSchema() interface{}             { return &SaveDraftQuotesInput{} }
func (a *SaveDraftQuotesAction) RequiresConfirmation() bool           { return true }
func (a *SaveDraftQuotesAction) RequiredPermission() (string, string) { return rbac.ResourceRFQs, rbac.ActionUpdate }

func (a *SaveDraftQuotesAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in SaveDraftQuotesInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.RFQID <= 0 || len(in.Quotes) == 0 {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "rfq_id and non-empty quotes are required"},
		}, nil
	}

	orgID32 := int32(ctx.OrganizationID)
	rfqID32 := int32(in.RFQID)

	// Verify RFQ belongs to actor's organization
	rfqObj, err := a.rfqBL.GetRFQ(ctx.Context, orgID32, rfqID32)
	if err != nil || rfqObj == nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "NotFound", Message: "RFQ not found or organization mismatch"},
		}, nil
	}

	savedCount := 0
	for _, q := range in.Quotes {
		if strings.TrimSpace(q.CarrierName) == "" || q.BuyPrice < 0 || q.SellPrice < 0 {
			return &ActionResult{
				Success: false,
				Action:  a.Name(),
				Error:   &ActionError{Type: "Validation", Message: "Invalid quote pricing or missing carrier name"},
			}, nil
		}
		transitDays := q.TransitTimeDays
		reasoning := q.AiReasoning
		quote := &rfqspec.Quote{
			RFQID:                 int32(in.RFQID),
			CarrierName:           q.CarrierName,
			TransitTimeDays:       &transitDays,
			BuyPrice:              q.BuyPrice,
			SellPrice:             q.SellPrice,
			IsRecommended:         q.IsRecommended,
			ReliabilityScore:      q.ReliabilityScore,
			HistoricalSuccessRate: q.HistoricalSuccessRate,
			AiReasoning:           &reasoning,
			Status:                "DRAFT",
		}
		if err := a.rfqBL.AddQuote(ctx.Context, orgID32, quote); err != nil {
			return &ActionResult{
				Success: false,
				Action:  a.Name(),
				Error:   &ActionError{Type: "BusinessRule", Message: "Failed to add draft quote: " + err.Error()},
			}, nil
		}
		savedCount++
	}

	return &ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "RFQQuote",
		ResourceID:   fmt.Sprintf("%d", in.RFQID),
		Summary:      fmt.Sprintf("Saved %d draft quote(s) for RFQ #%d.", savedCount, in.RFQID),
		Data: map[string]interface{}{
			"rfq_id":      in.RFQID,
			"saved_count": savedCount,
		},
	}, nil
}

// ── pricing.apply_selected_rate ───────────────────────────────────────────────

type ApplySelectedRateInput struct {
	RFQID   int64 `json:"rfq_id"`
	QuoteID int64 `json:"quote_id"`
}

type ApplySelectedRateAction struct {
	rfqBL rfq.BusinessLogic
}

func NewApplySelectedRateAction(rfqBL rfq.BusinessLogic) *ApplySelectedRateAction {
	return &ApplySelectedRateAction{rfqBL: rfqBL}
}

func (a *ApplySelectedRateAction) Name() string                         { return "pricing.apply_selected_rate" }
func (a *ApplySelectedRateAction) Module() string                       { return "pricing" }
func (a *ApplySelectedRateAction) Description() string                  { return "Approve and bind selected quote rate to RFQ (High-Risk action requiring human confirmation)." }
func (a *ApplySelectedRateAction) Category() ActionCategory             { return ActionCategoryHighRisk }
func (a *ApplySelectedRateAction) InputSchema() interface{}             { return &ApplySelectedRateInput{} }
func (a *ApplySelectedRateAction) RequiresConfirmation() bool           { return true }
func (a *ApplySelectedRateAction) RequiredPermission() (string, string) { return rbac.ResourceRFQs, rbac.ActionUpdate }

func (a *ApplySelectedRateAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in ApplySelectedRateInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.RFQID <= 0 || in.QuoteID <= 0 {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "rfq_id and quote_id are required"},
		}, nil
	}

	orgID32 := int32(ctx.OrganizationID)
	rfqID32 := int32(in.RFQID)
	quoteID32 := int32(in.QuoteID)

	updatedRFQ, err := a.rfqBL.ApproveQuote(ctx.Context, orgID32, rfqID32, quoteID32)
	if err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "BusinessRule", Message: "Failed to apply rate: " + err.Error()},
		}, nil
	}

	return &ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "RFQ",
		ResourceID:   fmt.Sprintf("%d", in.RFQID),
		Summary:      fmt.Sprintf("Rate quote #%d applied and approved for RFQ #%d.", in.QuoteID, in.RFQID),
		Data: map[string]interface{}{
			"rfq_id":   in.RFQID,
			"quote_id": in.QuoteID,
			"stage":    updatedRFQ.Stage,
		},
	}, nil
}
