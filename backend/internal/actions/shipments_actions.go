package actions

import (
	"encoding/json"
	"fmt"

	"github.com/freel/backend/internal/shipments"
)

type GetShipmentInput struct {
	ShipmentID int64 `json:"shipment_id"`
}

type GetShipmentAction struct {
	svc shipments.Service
}

func NewGetShipmentAction(svc shipments.Service) *GetShipmentAction {
	return &GetShipmentAction{svc: svc}
}

func (a *GetShipmentAction) Name() string { return "shipments.get" }
func (a *GetShipmentAction) Module() string { return "shipments" }
func (a *GetShipmentAction) Description() string { return "Retrieve a specific shipment by ID." }
func (a *GetShipmentAction) Category() ActionCategory { return ActionCategoryRead }
func (a *GetShipmentAction) InputSchema() interface{} { return &GetShipmentInput{} }
func (a *GetShipmentAction) RequiresConfirmation() bool { return false }

func (a *GetShipmentAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in GetShipmentInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.ShipmentID == 0 {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "shipment_id is required"},
		}, nil
	}

	// The shipments service requires the standard context to extract OrganizationID
	// but currently accepts orgID explicitly in GetShipmentByID.
	shipment, err := a.svc.GetShipmentByID(ctx.Context, ctx.OrganizationID, in.ShipmentID)
	if err != nil {
		// In a real implementation we would map svcerror to ActionError.
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "BusinessRule", Message: err.Error()},
		}, nil
	}

	return &ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "Shipment",
		ResourceID:   fmt.Sprintf("%d", shipment.ID),
		Summary:      fmt.Sprintf("Shipment %d retrieved successfully.", shipment.ID),
		Data:         shipment,
	}, nil
}
