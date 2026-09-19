package actions

import (
	"encoding/json"
	"fmt"

	"github.com/freel/backend/internal/finance"
	"github.com/freel/backend/internal/rbac"
)

// ── finance.reconcile_invoice ─────────────────────────────────────────────────

type ReconcileInvoiceInput struct {
	ShipmentID    int64                        `json:"shipment_id"`
	InvoiceID     string                       `json:"invoice_id"`
	Status        string                       `json:"status"`
	Items         []finance.InvoiceItem         `json:"items"`
	Discrepancies []*finance.FinanceDiscrepancy `json:"discrepancies"`
	AISummary     string                       `json:"ai_summary"`
}

type ReconcileInvoiceAction struct {
	svc finance.Service
}

func NewReconcileInvoiceAction(svc finance.Service) *ReconcileInvoiceAction {
	return &ReconcileInvoiceAction{svc: svc}
}

func (a *ReconcileInvoiceAction) Name() string                         { return "finance.reconcile_invoice" }
func (a *ReconcileInvoiceAction) Module() string                       { return "finance" }
func (a *ReconcileInvoiceAction) Description() string                  { return "Record 3-way invoice match reconciliation items and discrepancies." }
func (a *ReconcileInvoiceAction) Category() ActionCategory             { return ActionCategoryWrite }
func (a *ReconcileInvoiceAction) InputSchema() interface{}             { return &ReconcileInvoiceInput{} }
func (a *ReconcileInvoiceAction) RequiresConfirmation() bool           { return false }
func (a *ReconcileInvoiceAction) RequiredPermission() (string, string) { return rbac.ResourceFinance, rbac.ActionUpdate }

func (a *ReconcileInvoiceAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in ReconcileInvoiceInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.ShipmentID <= 0 || in.InvoiceID == "" {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "shipment_id and invoice_id are required"},
		}, nil
	}

	status := in.Status
	if status == "" {
		status = "PROCESSED"
	}

	req := &finance.FinanceCallbackRequest{
		OrgID:         ctx.OrganizationID,
		ShipmentID:    in.ShipmentID,
		InvoiceID:     in.InvoiceID,
		Status:        status,
		Items:         in.Items,
		Discrepancies: in.Discrepancies,
		AISummary:     in.AISummary,
	}

	err := a.svc.CompleteReconciliation(ctx.Context, req)
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
		ResourceType: "InvoiceReconciliation",
		ResourceID:   in.InvoiceID,
		Summary:      fmt.Sprintf("Invoice %s reconciliation recorded (%d items, %d discrepancies).", in.InvoiceID, len(in.Items), len(in.Discrepancies)),
		Data: map[string]interface{}{
			"invoice_id":          in.InvoiceID,
			"shipment_id":         in.ShipmentID,
			"status":              status,
			"discrepancies_count": len(in.Discrepancies),
		},
	}, nil
}
