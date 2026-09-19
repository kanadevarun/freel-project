package actions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/freel/backend/internal/contracts"
	"github.com/freel/backend/internal/rates"
	"github.com/freel/backend/internal/rbac"
)

// ── contracts.ingest_rates ───────────────────────────────────────────────────

type IngestRatesInput struct {
	DocumentID     string                      `json:"document_id"`
	Status         string                      `json:"status"`
	ConfirmedRates []rates.CanonicalRate       `json:"confirmed_rates"`
	FlaggedItems   []contracts.ReviewItemDraft `json:"flagged_items"`
	AISummary      string                      `json:"ai_summary"`
	CorrelationID  string                      `json:"correlation_id"`
}

type IngestRatesAction struct {
	svc contracts.Service
}

func NewIngestRatesAction(svc contracts.Service) *IngestRatesAction {
	return &IngestRatesAction{svc: svc}
}

func (a *IngestRatesAction) Name() string                         { return "contracts.ingest_rates" }
func (a *IngestRatesAction) Module() string                       { return "contracts" }
func (a *IngestRatesAction) Description() string                  { return "Ingest extracted contract rates and flagged anomalies into the review queue." }
func (a *IngestRatesAction) Category() ActionCategory             { return ActionCategoryWrite }
func (a *IngestRatesAction) InputSchema() interface{}             { return &IngestRatesInput{} }
func (a *IngestRatesAction) RequiresConfirmation() bool           { return false }
func (a *IngestRatesAction) RequiredPermission() (string, string) { return rbac.ResourceDocuments, rbac.ActionUpdate }

func (a *IngestRatesAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in IngestRatesInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if strings.TrimSpace(in.DocumentID) == "" {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "document_id is required"},
		}, nil
	}

	status := in.Status
	if status == "" {
		status = "COMPLETED"
	}

	callback := contracts.AIProcessingCallback{
		DocumentID:     in.DocumentID,
		OrgID:          ctx.OrganizationID,
		Status:         status,
		ConfirmedRates: in.ConfirmedRates,
		FlaggedItems:   in.FlaggedItems,
		AISummary:      in.AISummary,
		CorrelationID:  in.CorrelationID,
	}

	err := a.svc.HandleAICallback(ctx.Context, callback)
	if err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "BusinessRule", Message: err.Error()},
		}, nil
	}

	return &ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "ContractDocument",
		ResourceID:   in.DocumentID,
		Summary:      fmt.Sprintf("Contract %s rates ingested: %d confirmed, %d flagged.", in.DocumentID, len(in.ConfirmedRates), len(in.FlaggedItems)),
		Data: map[string]interface{}{
			"document_id":     in.DocumentID,
			"confirmed_count": len(in.ConfirmedRates),
			"flagged_count":   len(in.FlaggedItems),
		},
	}, nil
}

// ── contracts.review_extraction ───────────────────────────────────────────────

type ReviewExtractionInput struct {
	ReviewID      string          `json:"review_id"`
	Decision      string          `json:"decision"` // "APPROVE" or "REJECT"
	CorrectedData json.RawMessage `json:"corrected_data,omitempty"`
	Notes         string          `json:"notes"`
}

type ReviewExtractionAction struct {
	svc contracts.Service
}

func NewReviewExtractionAction(svc contracts.Service) *ReviewExtractionAction {
	return &ReviewExtractionAction{svc: svc}
}

func (a *ReviewExtractionAction) Name() string                         { return "contracts.review_extraction" }
func (a *ReviewExtractionAction) Module() string                       { return "contracts" }
func (a *ReviewExtractionAction) Description() string                  { return "Approve or reject flagged contract rate extraction (High-Risk action requiring human confirmation)." }
func (a *ReviewExtractionAction) Category() ActionCategory             { return ActionCategoryHighRisk }
func (a *ReviewExtractionAction) InputSchema() interface{}             { return &ReviewExtractionInput{} }
func (a *ReviewExtractionAction) RequiresConfirmation() bool           { return true }
func (a *ReviewExtractionAction) RequiredPermission() (string, string) { return rbac.ResourceDocuments, rbac.ActionUpdate }

func (a *ReviewExtractionAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in ReviewExtractionInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if strings.TrimSpace(in.ReviewID) == "" {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "review_id is required"},
		}, nil
	}

	reviewerID := ctx.ActingUserID
	if reviewerID <= 0 {
		reviewerID = 1
	}

	decision := strings.ToUpper(strings.TrimSpace(in.Decision))
	if decision == "APPROVE" {
		err := a.svc.ApproveReviewItem(ctx.Context, ctx.OrganizationID, in.ReviewID, reviewerID, in.CorrectedData, in.Notes)
		if err != nil {
			return &ActionResult{
				Success: false,
				Action:  a.Name(),
				Error:   &ActionError{Type: "BusinessRule", Message: err.Error()},
			}, nil
		}
	} else if decision == "REJECT" {
		err := a.svc.RejectReviewItem(ctx.Context, ctx.OrganizationID, in.ReviewID, reviewerID, in.Notes)
		if err != nil {
			return &ActionResult{
				Success: false,
				Action:  a.Name(),
				Error:   &ActionError{Type: "BusinessRule", Message: err.Error()},
			}, nil
		}
	} else {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "decision must be 'APPROVE' or 'REJECT'"},
		}, nil
	}

	return &ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "RateReviewItem",
		ResourceID:   in.ReviewID,
		Summary:      fmt.Sprintf("Contract review item %s marked as %s.", in.ReviewID, decision),
		Data: map[string]interface{}{
			"review_id": in.ReviewID,
			"decision":  decision,
		},
	}, nil
}
