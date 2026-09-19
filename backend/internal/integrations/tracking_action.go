package integrations

import (
	"encoding/json"
	"fmt"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/rbac"
)

// FetchTrackingActionInput defines the parameters for the tracking.fetch action.
type FetchTrackingActionInput struct {
	CarrierSCAC    string `json:"carrier_scac"`
	TrackingNumber string `json:"tracking_number"`
}

// FetchTrackingAction integrates carrier telemetry lookups into the LogisticsHQ Action System.
type FetchTrackingAction struct {
	service GatewayService
}

// NewFetchTrackingAction creates an action instance for carrier tracking query.
func NewFetchTrackingAction(service GatewayService) actions.Action {
	return &FetchTrackingAction{service: service}
}

func (a *FetchTrackingAction) Name() string {
	return "tracking.fetch"
}

func (a *FetchTrackingAction) Module() string {
	return "carrier_tracking"
}

func (a *FetchTrackingAction) Description() string {
	return "Fetch real-time normalized carrier telemetry and milestones via external carrier gateway."
}

func (a *FetchTrackingAction) Category() actions.ActionCategory {
	return actions.ActionCategoryRead
}

func (a *FetchTrackingAction) InputSchema() interface{} {
	return &FetchTrackingActionInput{}
}

func (a *FetchTrackingAction) RequiresConfirmation() bool {
	return false
}

func (a *FetchTrackingAction) RequiredPermission() (string, string) {
	return rbac.ResourceShipments, rbac.ActionRead
}

func (a *FetchTrackingAction) Execute(ctx *actions.ActionContext, input []byte) (*actions.ActionResult, error) {
	var in FetchTrackingActionInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &actions.ActionResult{
			Success: false,
			Action:  a.Name(),
			Error: &actions.ActionError{
				Type:    "Validation",
				Message: fmt.Sprintf("invalid tracking action payload: %v", err),
			},
		}, nil
	}

	if in.CarrierSCAC == "" || in.TrackingNumber == "" {
		return &actions.ActionResult{
			Success: false,
			Action:  a.Name(),
			Error: &actions.ActionError{
				Type:    "Validation",
				Message: "carrier_scac and tracking_number are required",
			},
		}, nil
	}

	var orgID int64
	if ctx != nil {
		orgID = ctx.OrganizationID
	}

	resp, err := a.service.GetTracking(ctx.Context, orgID, in.CarrierSCAC, in.TrackingNumber)
	if err != nil {
		return &actions.ActionResult{
			Success: false,
			Action:  a.Name(),
			Error: &actions.ActionError{
				Type:    "ExecutionFailed",
				Message: err.Error(),
			},
		}, nil
	}

	return &actions.ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "CarrierTracking",
		ResourceID:   fmt.Sprintf("%s:%s", in.CarrierSCAC, in.TrackingNumber),
		Summary:      fmt.Sprintf("Tracking telemetry retrieved for %s #%s (Status: %s)", in.CarrierSCAC, in.TrackingNumber, resp.Status),
		Data:         resp,
	}, nil
}
