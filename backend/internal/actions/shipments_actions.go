package actions

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/freel/backend/internal/rbac"
	"github.com/freel/backend/internal/shipments"
)

// ── shipments.get ─────────────────────────────────────────────────────────────

type GetShipmentInput struct {
	ShipmentID int64 `json:"shipment_id"`
}

type GetShipmentAction struct {
	svc shipments.Service
}

func NewGetShipmentAction(svc shipments.Service) *GetShipmentAction {
	return &GetShipmentAction{svc: svc}
}

func (a *GetShipmentAction) Name() string                               { return "shipments.get" }
func (a *GetShipmentAction) Module() string                             { return "shipments" }
func (a *GetShipmentAction) Description() string                        { return "Retrieve a specific shipment by ID." }
func (a *GetShipmentAction) Category() ActionCategory                   { return ActionCategoryRead }
func (a *GetShipmentAction) InputSchema() interface{}                   { return &GetShipmentInput{} }
func (a *GetShipmentAction) RequiresConfirmation() bool                 { return false }
func (a *GetShipmentAction) RequiredPermission() (string, string)       { return rbac.ResourceShipments, rbac.ActionRead }

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

	shipment, err := a.svc.GetShipmentByID(ctx.Context, ctx.OrganizationID, in.ShipmentID)
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
		ResourceType: "Shipment",
		ResourceID:   fmt.Sprintf("%d", shipment.ID),
		Summary:      fmt.Sprintf("Shipment %d retrieved successfully.", shipment.ID),
		Data:         shipment,
	}, nil
}

// ── shipments.update_milestone ────────────────────────────────────────────────

type UpdateMilestoneInput struct {
	ShipmentID    int64   `json:"shipment_id"`
	MilestoneCode string  `json:"milestone_code"`
	ActualDate    string  `json:"actual_date"`
	Location      *string `json:"location,omitempty"`
	Notes         *string `json:"notes,omitempty"`
}

func (in *UpdateMilestoneInput) UnmarshalJSON(data []byte) error {
	type Alias UpdateMilestoneInput
	var aux struct {
		RawShipmentID interface{} `json:"shipment_id"`
		Alias
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*in = UpdateMilestoneInput(aux.Alias)
	switch v := aux.RawShipmentID.(type) {
	case float64:
		in.ShipmentID = int64(v)
	case int64:
		in.ShipmentID = v
	case int:
		in.ShipmentID = int64(v)
	case string:
		var id int64
		fmt.Sscanf(v, "%d", &id)
		in.ShipmentID = id
	}
	return nil
}

type UpdateMilestoneAction struct {
	svc shipments.Service
}

func NewUpdateMilestoneAction(svc shipments.Service) *UpdateMilestoneAction {
	return &UpdateMilestoneAction{svc: svc}
}

func (a *UpdateMilestoneAction) Name() string                               { return "shipments.update_milestone" }
func (a *UpdateMilestoneAction) Module() string                             { return "shipments" }
func (a *UpdateMilestoneAction) Description() string                        { return "Update shipment tracking milestone code and completion timestamp." }
func (a *UpdateMilestoneAction) Category() ActionCategory                   { return ActionCategoryWrite }
func (a *UpdateMilestoneAction) InputSchema() interface{}                   { return &UpdateMilestoneInput{} }
func (a *UpdateMilestoneAction) RequiresConfirmation() bool                 { return false }
func (a *UpdateMilestoneAction) RequiredPermission() (string, string)       { return rbac.ResourceShipments, rbac.ActionUpdate }

func (a *UpdateMilestoneAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in UpdateMilestoneInput
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
	if strings.TrimSpace(in.MilestoneCode) == "" {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "milestone_code is required"},
		}, nil
	}

	var actualTime time.Time
	if strings.TrimSpace(in.ActualDate) != "" {
		t, err := time.Parse(time.RFC3339, in.ActualDate)
		if err != nil {
			t, err = time.Parse("2006-01-02 15:04:05", in.ActualDate)
		}
		if err != nil {
			t, err = time.Parse("2006-01-02", in.ActualDate)
		}
		if err == nil {
			actualTime = t
		} else {
			actualTime = time.Now()
		}
	} else {
		actualTime = time.Now()
	}

	err := a.svc.UpdateMilestone(ctx.Context, ctx.OrganizationID, in.ShipmentID, in.MilestoneCode, &actualTime, in.Location, in.Notes)
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
		ResourceType: "ShipmentMilestone",
		ResourceID:   fmt.Sprintf("%d-%s", in.ShipmentID, in.MilestoneCode),
		Summary:      fmt.Sprintf("Milestone %s for shipment #%d updated to COMPLETED.", in.MilestoneCode, in.ShipmentID),
		Data: map[string]interface{}{
			"shipment_id":    in.ShipmentID,
			"milestone_code": in.MilestoneCode,
			"status":         "COMPLETED",
			"actual_date":    actualTime.Format(time.RFC3339),
		},
	}, nil
}

// ── shipments.create_exception ────────────────────────────────────────────────

type CreateExceptionInput struct {
	ShipmentID    int64   `json:"shipment_id"`
	ExceptionType string  `json:"exception_type"`
	Severity      string  `json:"severity"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	SourceEventID *string `json:"source_event_id,omitempty"`
}

type CreateExceptionAction struct {
	svc shipments.Service
}

func NewCreateExceptionAction(svc shipments.Service) *CreateExceptionAction {
	return &CreateExceptionAction{svc: svc}
}

func (a *CreateExceptionAction) Name() string                               { return "shipments.create_exception" }
func (a *CreateExceptionAction) Module() string                             { return "shipments" }
func (a *CreateExceptionAction) Description() string                        { return "Create an operational exception or alert on a shipment." }
func (a *CreateExceptionAction) Category() ActionCategory                   { return ActionCategoryWrite }
func (a *CreateExceptionAction) InputSchema() interface{}                   { return &CreateExceptionInput{} }
func (a *CreateExceptionAction) RequiresConfirmation() bool                 { return false }
func (a *CreateExceptionAction) RequiredPermission() (string, string)       { return rbac.ResourceShipments, rbac.ActionCreate }

func (a *CreateExceptionAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in CreateExceptionInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.ShipmentID <= 0 || strings.TrimSpace(in.Title) == "" {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "shipment_id and title are required"},
		}, nil
	}

	sev := strings.ToUpper(strings.TrimSpace(in.Severity))
	switch sev {
	case "CRITICAL":
		in.Severity = "CRITICAL"
	case "HIGH", "WARNING":
		in.Severity = "HIGH"
	case "LOW", "INFO":
		in.Severity = "LOW"
	default:
		in.Severity = "MEDIUM"
	}

	exType := strings.ToUpper(strings.TrimSpace(in.ExceptionType))
	switch exType {
	case "DELAY":
		in.ExceptionType = "SCHEDULE_DELAY"
	case "ROLLOVER":
		in.ExceptionType = "VESSEL_ROLLOVER"
	case "CONGESTION":
		in.ExceptionType = "PORT_CONGESTION"
	case "HOLD":
		in.ExceptionType = "CUSTOMS_HOLD"
	case "DOCUMENT":
		in.ExceptionType = "DOCUMENT_ISSUE"
	}

	err := a.svc.CreateShipmentException(ctx.Context, ctx.OrganizationID, in.ShipmentID, in.ExceptionType, in.Severity, in.Title, in.Description, in.SourceEventID)
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
		ResourceType: "ShipmentException",
		ResourceID:   fmt.Sprintf("%d-%s", in.ShipmentID, in.ExceptionType),
		Summary:      fmt.Sprintf("Exception %s (%s) created for shipment #%d.", in.Title, in.Severity, in.ShipmentID),
		Data: map[string]interface{}{
			"shipment_id":    in.ShipmentID,
			"exception_type": in.ExceptionType,
			"severity":       in.Severity,
			"title":          in.Title,
		},
	}, nil
}

// ── operations.send_callback ──────────────────────────────────────────────────

type OperationsCallbackInput struct {
	ShipmentID           int64  `json:"shipment_id"`
	EventID              string `json:"event_id,omitempty"`
	HasCriticalException bool   `json:"has_critical_exception"`
	AISummary            string `json:"ai_summary"`
}

type OperationsCallbackAction struct {
	svc shipments.Service
}

func NewOperationsCallbackAction(svc shipments.Service) *OperationsCallbackAction {
	return &OperationsCallbackAction{svc: svc}
}

func (a *OperationsCallbackAction) Name() string                               { return "operations.send_callback" }
func (a *OperationsCallbackAction) Module() string                             { return "operations" }
func (a *OperationsCallbackAction) Description() string                        { return "Process OperationsAgent tracking evaluation callback." }
func (a *OperationsCallbackAction) Category() ActionCategory                   { return ActionCategoryWrite }
func (a *OperationsCallbackAction) InputSchema() interface{}                   { return &OperationsCallbackInput{} }
func (a *OperationsCallbackAction) RequiresConfirmation() bool                 { return false }
func (a *OperationsCallbackAction) RequiredPermission() (string, string)       { return rbac.ResourceShipments, rbac.ActionUpdate }

func (a *OperationsCallbackAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in OperationsCallbackInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.ShipmentID > 0 {
		sh, err := a.svc.GetShipmentByID(ctx.Context, ctx.OrganizationID, in.ShipmentID)
		if err != nil || sh == nil {
			return &ActionResult{
				Success: false,
				Action:  a.Name(),
				Error:   &ActionError{Type: "NotFound", Message: "Shipment not found or organization mismatch"},
			}, nil
		}
	}

	if in.EventID != "" {
		err := a.svc.CompleteCarrierEvent(ctx.Context, in.EventID, ctx.OrganizationID, in.ShipmentID, in.HasCriticalException, in.AISummary)
		if err != nil {
			return &ActionResult{
				Success: false,
				Action:  a.Name(),
				Error:   &ActionError{Type: "BusinessRule", Message: err.Error()},
			}, nil
		}
	}

	return &ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "OperationsCallback",
		ResourceID:   fmt.Sprintf("%d", in.ShipmentID),
		Summary:      fmt.Sprintf("Operations callback processed for shipment #%d.", in.ShipmentID),
		Data: map[string]interface{}{
			"shipment_id":            in.ShipmentID,
			"has_critical_exception": in.HasCriticalException,
		},
	}, nil
}

// ── shipments.update_eta ──────────────────────────────────────────────────────

type UpdateETAInput struct {
	ShipmentID       int64  `json:"shipment_id"`
	EstimatedArrival string `json:"estimated_arrival"`
	Notes            string `json:"notes,omitempty"`
}

type UpdateETAAction struct {
	svc shipments.Service
}

func NewUpdateETAAction(svc shipments.Service) *UpdateETAAction {
	return &UpdateETAAction{svc: svc}
}

func (a *UpdateETAAction) Name() string                               { return "shipments.update_eta" }
func (a *UpdateETAAction) Module() string                             { return "shipments" }
func (a *UpdateETAAction) Description() string                        { return "Update shipment estimated time of arrival (ETA)." }
func (a *UpdateETAAction) Category() ActionCategory                   { return ActionCategoryWrite }
func (a *UpdateETAAction) InputSchema() interface{}                   { return &UpdateETAInput{} }
func (a *UpdateETAAction) RequiresConfirmation() bool                 { return false }
func (a *UpdateETAAction) RequiredPermission() (string, string)       { return rbac.ResourceShipments, rbac.ActionUpdate }

func (a *UpdateETAAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in UpdateETAInput
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

	return &ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "ShipmentETA",
		ResourceID:   fmt.Sprintf("%d", in.ShipmentID),
		Summary:      fmt.Sprintf("Shipment #%d ETA updated successfully.", in.ShipmentID),
		Data: map[string]interface{}{
			"shipment_id": in.ShipmentID,
			"status":      "ETA_UPDATED",
		},
	}, nil
}

// ── shipments.refresh_tracking ────────────────────────────────────────────────

type RefreshTrackingInput struct {
	ShipmentID int64 `json:"shipment_id"`
}

func (in *RefreshTrackingInput) UnmarshalJSON(data []byte) error {
	type Alias RefreshTrackingInput
	var aux struct {
		RawShipmentID interface{} `json:"shipment_id"`
		Alias
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*in = RefreshTrackingInput(aux.Alias)
	switch v := aux.RawShipmentID.(type) {
	case float64:
		in.ShipmentID = int64(v)
	case int64:
		in.ShipmentID = v
	case int:
		in.ShipmentID = int64(v)
	case string:
		var id int64
		if _, err := fmt.Sscanf(v, "%d", &id); err == nil {
			in.ShipmentID = id
		}
	}
	return nil
}

type RefreshShipmentTrackingAction struct {
	svc shipments.Service
}

func NewRefreshShipmentTrackingAction(svc shipments.Service) *RefreshShipmentTrackingAction {
	return &RefreshShipmentTrackingAction{svc: svc}
}

func (a *RefreshShipmentTrackingAction) Name() string                         { return "shipments.refresh_tracking" }
func (a *RefreshShipmentTrackingAction) Module() string                       { return "shipments" }
func (a *RefreshShipmentTrackingAction) Description() string                  { return "Synchronize and refresh carrier telemetry and milestones for a shipment." }
func (a *RefreshShipmentTrackingAction) Category() ActionCategory             { return ActionCategoryWrite }
func (a *RefreshShipmentTrackingAction) InputSchema() interface{}             { return &RefreshTrackingInput{} }
func (a *RefreshShipmentTrackingAction) RequiresConfirmation() bool           { return false }
func (a *RefreshShipmentTrackingAction) RequiredPermission() (string, string) { return rbac.ResourceShipments, rbac.ActionUpdate }

func (a *RefreshShipmentTrackingAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in RefreshTrackingInput
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

	var userID *int64
	actor := "ACTION_SYSTEM"
	if ctx != nil {
		if ctx.ActingUserID > 0 {
			uid := ctx.ActingUserID
			userID = &uid
		}
		if ctx.ActorType != "" {
			actor = string(ctx.ActorType)
		}
	}

	res, err := a.svc.RefreshShipmentTracking(ctx.Context, ctx.OrganizationID, in.ShipmentID, userID, actor)
	if err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "ExecutionFailed", Message: err.Error()},
		}, nil
	}

	return &ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "ShipmentTracking",
		ResourceID:   fmt.Sprintf("%d", in.ShipmentID),
		Summary:      fmt.Sprintf("Tracking refresh executed for shipment #%d: %s", in.ShipmentID, res.Message),
		Data:         res,
	}, nil
}


