package actions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/freel/backend/internal/leads"
	leadspec "github.com/freel/backend/internal/leads/spec"
	"github.com/freel/backend/internal/rbac"
	"github.com/freel/backend/internal/rfq"
	rfqspec "github.com/freel/backend/internal/rfq/spec"
)

// ── sales.create_rfq_from_email ───────────────────────────────────────────────

type CreateRFQFromEmailInput struct {
	CustomerID  int64             `json:"customer_id"`
	Origin      string            `json:"origin"`
	Destination string            `json:"destination"`
	Incoterms   string            `json:"incoterms"`
	TargetDate  string            `json:"target_date"`
	Items       []rfqspec.RFQItem `json:"items"`
}

type CreateRFQFromEmailAction struct {
	leadsBL leads.BusinessLogic
	rfqBL   rfq.BusinessLogic
}

func NewCreateRFQFromEmailAction(leadsBL leads.BusinessLogic, rfqBL rfq.BusinessLogic) *CreateRFQFromEmailAction {
	return &CreateRFQFromEmailAction{
		leadsBL: leadsBL,
		rfqBL:   rfqBL,
	}
}

func (a *CreateRFQFromEmailAction) Name() string                         { return "sales.create_rfq_from_email" }
func (a *CreateRFQFromEmailAction) Module() string                       { return "sales" }
func (a *CreateRFQFromEmailAction) Description() string                  { return "Create an RFQ parsed from inbound customer email." }
func (a *CreateRFQFromEmailAction) Category() ActionCategory             { return ActionCategoryWrite }
func (a *CreateRFQFromEmailAction) InputSchema() interface{}             { return &CreateRFQFromEmailInput{} }
func (a *CreateRFQFromEmailAction) RequiresConfirmation() bool           { return false }
func (a *CreateRFQFromEmailAction) RequiredPermission() (string, string) { return rbac.ResourceRFQs, rbac.ActionCreate }

func (a *CreateRFQFromEmailAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in CreateRFQFromEmailInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.CustomerID <= 0 {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "customer_id is required"},
		}, nil
	}

	orgID32 := int32(ctx.OrganizationID)
	leadID32 := int32(in.CustomerID)

	// 1. Resolve lead and ensure organization isolation
	lead, err := a.leadsBL.GetLead(ctx.Context, orgID32, leadID32)
	if err != nil || lead == nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "NotFound", Message: "Lead not found or organization mismatch"},
		}, nil
	}

	// 2. Mark lead converted if not already
	convertedStatus := "CONVERTED"
	_, _ = a.leadsBL.UpdateLead(ctx.Context, leadspec.UpdateLeadRequest{
		OrgID:  orgID32,
		ID:     leadID32,
		Status: &convertedStatus,
	})

	// 3. Resolve customer ID for company
	customerID, err := a.leadsBL.GetCustomerIDByCompanyName(ctx.Context, orgID32, lead.CompanyName)
	if err != nil {
		customerID = leadID32
	}

	// 4. Check if RFQ already drafted for this lead (Idempotency check)
	existingRFQ, err := a.rfqBL.GetRFQByLeadID(ctx.Context, orgID32, in.CustomerID)
	if err == nil && existingRFQ != nil {
		return &ActionResult{
			Success:      true,
			Action:       a.Name(),
			ResourceType: "RFQ",
			ResourceID:   fmt.Sprintf("%d", existingRFQ.ID),
			Summary:      fmt.Sprintf("RFQ #%s already exists for lead #%d.", existingRFQ.RFQNumber, in.CustomerID),
			Data: map[string]interface{}{
				"rfq_id":     existingRFQ.ID,
				"rfq_number": existingRFQ.RFQNumber,
				"is_new":     false,
			},
		}, nil
	}

	var parsedDate time.Time
	if in.TargetDate != "" {
		parsedDate, _ = time.Parse("2006-01-02", in.TargetDate)
	} else {
		parsedDate = time.Now().AddDate(0, 1, 0)
	}

	rfqReq := rfqspec.CreateRFQRequest{
		CustomerID:  customerID,
		Origin:      &in.Origin,
		Destination: &in.Destination,
		Incoterms:   &in.Incoterms,
		TargetDate:  &parsedDate,
		Items:       in.Items,
	}
	rfqReq.OrgID = orgID32
	leadID64 := in.CustomerID
	rfqReq.LeadID = &leadID64

	newRFQ, err := a.rfqBL.CreateRFQ(ctx.Context, rfqReq)
	if err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "BusinessRule", Message: "Failed to create RFQ: " + err.Error()},
		}, nil
	}

	return &ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "RFQ",
		ResourceID:   fmt.Sprintf("%d", newRFQ.ID),
		Summary:      fmt.Sprintf("Draft RFQ #%s created from email for lead #%d.", newRFQ.RFQNumber, in.CustomerID),
		Data: map[string]interface{}{
			"rfq_id":     newRFQ.ID,
			"rfq_number": newRFQ.RFQNumber,
			"is_new":     true,
		},
	}, nil
}

// ── sales.convert_lead ────────────────────────────────────────────────────────

type ConvertLeadInput struct {
	LeadID int64  `json:"lead_id"`
	Status string `json:"status"`
}

type ConvertLeadAction struct {
	leadsBL leads.BusinessLogic
}

func NewConvertLeadAction(leadsBL leads.BusinessLogic) *ConvertLeadAction {
	return &ConvertLeadAction{leadsBL: leadsBL}
}

func (a *ConvertLeadAction) Name() string                         { return "sales.convert_lead" }
func (a *ConvertLeadAction) Module() string                       { return "sales" }
func (a *ConvertLeadAction) Description() string                  { return "Convert sales lead to customer." }
func (a *ConvertLeadAction) Category() ActionCategory             { return ActionCategoryWrite }
func (a *ConvertLeadAction) InputSchema() interface{}             { return &ConvertLeadInput{} }
func (a *ConvertLeadAction) RequiresConfirmation() bool           { return false }
func (a *ConvertLeadAction) RequiredPermission() (string, string) { return rbac.ResourceLeads, rbac.ActionUpdate }

func (a *ConvertLeadAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in ConvertLeadInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.LeadID <= 0 {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "lead_id is required"},
		}, nil
	}

	status := in.Status
	if status == "" {
		status = "CONVERTED"
	}

	lead, err := a.leadsBL.UpdateLead(ctx.Context, leadspec.UpdateLeadRequest{
		OrgID:  int32(ctx.OrganizationID),
		ID:     int32(in.LeadID),
		Status: &status,
	})
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
		ResourceType: "Lead",
		ResourceID:   fmt.Sprintf("%d", in.LeadID),
		Summary:      fmt.Sprintf("Lead #%d status updated to %s.", in.LeadID, status),
		Data:         lead,
	}, nil
}

// ── sales.send_clarification_email ────────────────────────────────────────────

type SendClarificationEmailInput struct {
	LeadID        int64  `json:"lead_id"`
	InteractionID int64  `json:"interaction_id"`
	Subject       string `json:"subject"`
	Content       string `json:"content"`
}

type SendClarificationEmailAction struct {
	leadsBL leads.BusinessLogic
}

func NewSendClarificationEmailAction(leadsBL leads.BusinessLogic) *SendClarificationEmailAction {
	return &SendClarificationEmailAction{leadsBL: leadsBL}
}

func (a *SendClarificationEmailAction) Name() string                         { return "sales.send_clarification_email" }
func (a *SendClarificationEmailAction) Module() string                       { return "sales" }
func (a *SendClarificationEmailAction) Description() string                  { return "Send outbound clarification email to lead (High-Risk action requiring human confirmation)." }
func (a *SendClarificationEmailAction) Category() ActionCategory             { return ActionCategoryHighRisk }
func (a *SendClarificationEmailAction) InputSchema() interface{}             { return &SendClarificationEmailInput{} }
func (a *SendClarificationEmailAction) RequiresConfirmation() bool           { return true }
func (a *SendClarificationEmailAction) RequiredPermission() (string, string) { return rbac.ResourceOutreach, rbac.ActionCreate }

func (a *SendClarificationEmailAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in SendClarificationEmailInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.LeadID <= 0 {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "lead_id is required"},
		}, nil
	}

	actorName := "Centralized Action Bridge"
	if ctx.ActingUserID > 0 {
		actorName = fmt.Sprintf("User #%d", ctx.ActingUserID)
	}

	// High risk action: Must be executed only when confirmed
	interaction, err := a.leadsBL.ApproveClarificationDraft(ctx.Context, ctx.OrganizationID, in.LeadID, in.InteractionID, actorName, "Approved via Centralized Action System")
	if err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "BusinessRule", Message: "Failed to send outbound email: " + err.Error()},
		}, nil
	}

	return &ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "LeadInteraction",
		ResourceID:   fmt.Sprintf("%d", in.LeadID),
		Summary:      fmt.Sprintf("Outbound clarification email approved and sent for lead #%d.", in.LeadID),
		Data:         interaction,
	}, nil
}
