package actions

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/freel/backend/internal/rbac"
	"github.com/freel/backend/internal/recommendations"
)

// ── List Recommendations Action (recommendations.list) ──────────────────────

type ListRecommendationsInput struct {
	Category string `json:"category,omitempty"`
	Priority string `json:"priority,omitempty"`
	Status   string `json:"status,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

type ListRecommendationsAction struct {
	svc recommendations.Service
}

func NewListRecommendationsAction(svc recommendations.Service) *ListRecommendationsAction {
	return &ListRecommendationsAction{svc: svc}
}

func (a *ListRecommendationsAction) Name() string                         { return "recommendations.list" }
func (a *ListRecommendationsAction) Module() string                       { return "recommendations" }
func (a *ListRecommendationsAction) Description() string                  { return "List grounded AI recommendations with verified evidence and priority ratings." }
func (a *ListRecommendationsAction) Category() ActionCategory             { return ActionCategoryRead }
func (a *ListRecommendationsAction) InputSchema() interface{}             { return &ListRecommendationsInput{} }
func (a *ListRecommendationsAction) RequiresConfirmation() bool           { return false }
func (a *ListRecommendationsAction) RequiredPermission() (string, string) { return rbac.ResourceDashboard, rbac.ActionRead }

func (a *ListRecommendationsAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in ListRecommendationsInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return &ActionResult{
				Success: false,
				Action:  a.Name(),
				Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
			}, nil
		}
	}

	filter := recommendations.RecommendationFilter{
		Category: in.Category,
		Priority: in.Priority,
		Status:   in.Status,
		Limit:    in.Limit,
	}

	resp, err := a.svc.ListRecommendations(ctx.Context, ctx.OrganizationID, filter)
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
		ResourceType: "RECOMMENDATIONS",
		Summary:      fmt.Sprintf("Found %d recommendations (Total: %d)", len(resp.Recommendations), resp.Pagination.Total),
		Data:         resp,
	}, nil
}

// ── Get Recommendation Action (recommendations.get) ─────────────────────────

type GetRecommendationInput struct {
	ID int64 `json:"id"`
}

type GetRecommendationAction struct {
	svc recommendations.Service
}

func NewGetRecommendationAction(svc recommendations.Service) *GetRecommendationAction {
	return &GetRecommendationAction{svc: svc}
}

func (a *GetRecommendationAction) Name() string                         { return "recommendations.get" }
func (a *GetRecommendationAction) Module() string                       { return "recommendations" }
func (a *GetRecommendationAction) Description() string                  { return "Retrieve details, rationale, and verified evidence for a recommendation." }
func (a *GetRecommendationAction) Category() ActionCategory             { return ActionCategoryRead }
func (a *GetRecommendationAction) InputSchema() interface{}             { return &GetRecommendationInput{} }
func (a *GetRecommendationAction) RequiresConfirmation() bool           { return false }
func (a *GetRecommendationAction) RequiredPermission() (string, string) { return rbac.ResourceDashboard, rbac.ActionRead }

func (a *GetRecommendationAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in GetRecommendationInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.ID <= 0 {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "id (>0) is required"},
		}, nil
	}

	rec, err := a.svc.GetRecommendation(ctx.Context, ctx.OrganizationID, in.ID)
	if err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "NotFound", Message: err.Error()},
		}, nil
	}

	return &ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "RECOMMENDATION",
		ResourceID:   fmt.Sprintf("%d", rec.ID),
		Summary:      fmt.Sprintf("Recommendation %d: %s (%s, Priority: %s)", rec.ID, rec.Title, rec.Category, rec.Priority),
		Data:         rec,
	}, nil
}

// ── Generate Follow-Up Draft Action (followups.generate_draft) ───────────────

type GenerateFollowupDraftInput struct {
	RecommendationID int64 `json:"recommendation_id"`
}

type GenerateFollowupDraftAction struct {
	svc recommendations.Service
}

func NewGenerateFollowupDraftAction(svc recommendations.Service) *GenerateFollowupDraftAction {
	return &GenerateFollowupDraftAction{svc: svc}
}

func (a *GenerateFollowupDraftAction) Name() string                         { return "followups.generate_draft" }
func (a *GenerateFollowupDraftAction) Module() string                       { return "recommendations" }
func (a *GenerateFollowupDraftAction) Description() string                  { return "Generate an editable communication draft for a customer follow-up." }
func (a *GenerateFollowupDraftAction) Category() ActionCategory             { return ActionCategoryRead }
func (a *GenerateFollowupDraftAction) InputSchema() interface{}             { return &GenerateFollowupDraftInput{} }
func (a *GenerateFollowupDraftAction) RequiresConfirmation() bool           { return false }
func (a *GenerateFollowupDraftAction) RequiredPermission() (string, string) { return rbac.ResourceCompanies, rbac.ActionRead }

func (a *GenerateFollowupDraftAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in GenerateFollowupDraftInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.RecommendationID <= 0 {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "recommendation_id is required"},
		}, nil
	}

	userLabel := fmt.Sprintf("user-%d", ctx.ActingUserID)

	rec, err := a.svc.GenerateDraft(ctx.Context, ctx.OrganizationID, in.RecommendationID, ctx.ActingUserID, userLabel, ctx.RequestID)
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
		ResourceType: "FOLLOWUP_DRAFT",
		ResourceID:   fmt.Sprintf("%d", rec.ID),
		Summary:      fmt.Sprintf("Generated draft for customer %s (Rec #%d)", *rec.CustomerName, rec.ID),
		Data:         rec,
	}, nil
}

// ── Create Follow-Up Task Action (followups.create_task) ─────────────────────

type CreateFollowupTaskActionInput struct {
	RecommendationID int64   `json:"recommendation_id"`
	Notes            string  `json:"notes,omitempty"`
	AssigneeID       *int64  `json:"assignee_id,omitempty"`
	AssigneeName     *string `json:"assignee_name,omitempty"`
	DueDate          *string `json:"due_date,omitempty"`
	Priority         string  `json:"priority,omitempty"`
}

type CreateFollowupTaskAction struct {
	svc recommendations.Service
}

func NewCreateFollowupTaskAction(svc recommendations.Service) *CreateFollowupTaskAction {
	return &CreateFollowupTaskAction{svc: svc}
}

func (a *CreateFollowupTaskAction) Name() string                         { return "followups.create_task" }
func (a *CreateFollowupTaskAction) Module() string                       { return "recommendations" }
func (a *CreateFollowupTaskAction) Description() string                  { return "Create an internal follow-up task from a validated recommendation." }
func (a *CreateFollowupTaskAction) Category() ActionCategory             { return ActionCategoryWrite }
func (a *CreateFollowupTaskAction) InputSchema() interface{}             { return &CreateFollowupTaskActionInput{} }
func (a *CreateFollowupTaskAction) RequiresConfirmation() bool           { return false }
func (a *CreateFollowupTaskAction) RequiredPermission() (string, string) { return rbac.ResourceCompanies, rbac.ActionUpdate }

func (a *CreateFollowupTaskAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in CreateFollowupTaskActionInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.RecommendationID <= 0 {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "recommendation_id is required"},
		}, nil
	}

	var dueTime *time.Time
	if in.DueDate != nil && strings.TrimSpace(*in.DueDate) != "" {
		if t, err := time.Parse(time.RFC3339, *in.DueDate); err == nil {
			dueTime = &t
		} else if t, err := time.Parse("2006-01-02", *in.DueDate); err == nil {
			dueTime = &t
		}
	}

	taskInput := recommendations.CreateFollowupTaskInput{
		Notes:        in.Notes,
		AssigneeID:   in.AssigneeID,
		AssigneeName: in.AssigneeName,
		DueDate:      dueTime,
		Priority:     in.Priority,
	}

	userLabel := fmt.Sprintf("user-%d", ctx.ActingUserID)

	task, err := a.svc.CreateFollowupTask(ctx.Context, ctx.OrganizationID, in.RecommendationID, taskInput, ctx.ActingUserID, userLabel, ctx.RequestID)
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
		ResourceType: "FOLLOWUP_TASK",
		ResourceID:   fmt.Sprintf("%d", task.ID),
		Summary:      fmt.Sprintf("Created follow-up task #%d for %s", task.ID, task.CustomerName),
		Data: map[string]interface{}{
			"task": task,
		},
	}, nil
}

