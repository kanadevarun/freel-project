package actions

import (
	"encoding/json"
	"fmt"

	"github.com/freel/backend/internal/documents"
	"github.com/freel/backend/internal/rbac"
)

// ── compliance.record_discrepancies ───────────────────────────────────────────

type RecordComplianceDiscrepanciesInput struct {
	ShipmentID    int64                                     `json:"shipment_id"`
	DocStatusList map[string]string                         `json:"doc_status_list"`
	Discrepancies []*documents.ShipmentDocumentDiscrepancy `json:"discrepancies"`
}

type RecordComplianceDiscrepanciesAction struct {
	svc documents.Service
}

func NewRecordComplianceDiscrepanciesAction(svc documents.Service) *RecordComplianceDiscrepanciesAction {
	return &RecordComplianceDiscrepanciesAction{svc: svc}
}

func (a *RecordComplianceDiscrepanciesAction) Name() string                         { return "compliance.record_discrepancies" }
func (a *RecordComplianceDiscrepanciesAction) Module() string                       { return "compliance" }
func (a *RecordComplianceDiscrepanciesAction) Description() string                  { return "Record compliance verification results and discrepancies for shipment documents." }
func (a *RecordComplianceDiscrepanciesAction) Category() ActionCategory             { return ActionCategoryWrite }
func (a *RecordComplianceDiscrepanciesAction) InputSchema() interface{}             { return &RecordComplianceDiscrepanciesInput{} }
func (a *RecordComplianceDiscrepanciesAction) RequiresConfirmation() bool           { return false }
func (a *RecordComplianceDiscrepanciesAction) RequiredPermission() (string, string) { return rbac.ResourceDocuments, rbac.ActionUpdate }

func (a *RecordComplianceDiscrepanciesAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in RecordComplianceDiscrepanciesInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.ShipmentID <= 0 {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "shipment_id is required"},
		}, nil
	}

	err := a.svc.CompleteVerification(ctx.Context, ctx.OrganizationID, in.ShipmentID, in.DocStatusList, in.Discrepancies)
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
		ResourceType: "ShipmentCompliance",
		ResourceID:   fmt.Sprintf("%d", in.ShipmentID),
		Summary:      fmt.Sprintf("Recorded %d compliance discrepancy records for shipment #%d.", len(in.Discrepancies), in.ShipmentID),
		Data: map[string]interface{}{
			"shipment_id":         in.ShipmentID,
			"discrepancies_count": len(in.Discrepancies),
		},
	}, nil
}
