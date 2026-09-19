package actions

import (
	"encoding/json"
	"fmt"
	"strings"

	bcontext "github.com/freel/backend/internal/context"
	"github.com/freel/backend/internal/rbac"
)

// ── Generic Context Read Action (context.get) ────────────────────────────────

type GetContextInput struct {
	EntityType string `json:"entity_type"`
	EntityID   int64  `json:"entity_id"`
}

type GetContextAction struct {
	svc bcontext.Service
}

func NewGetContextAction(svc bcontext.Service) *GetContextAction {
	return &GetContextAction{svc: svc}
}

func (a *GetContextAction) Name() string                         { return "context.get" }
func (a *GetContextAction) Module() string                       { return "intelligence" }
func (a *GetContextAction) Description() string                  { return "Retrieve cross-module unified business context for a record." }
func (a *GetContextAction) Category() ActionCategory             { return ActionCategoryRead }
func (a *GetContextAction) InputSchema() interface{}             { return &GetContextInput{} }
func (a *GetContextAction) RequiresConfirmation() bool           { return false }
func (a *GetContextAction) RequiredPermission() (string, string) { return rbac.ResourceCompanies, rbac.ActionRead }

func (a *GetContextAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in GetContextInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.EntityID <= 0 || in.EntityType == "" {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "entity_type and entity_id (>0) are required"},
		}, nil
	}

	bCtx, err := a.svc.GetBusinessContext(ctx.Context, bcontext.ContextRequest{
		OrgID:         ctx.OrganizationID,
		UserID:        ctx.ActingUserID,
		UserRole:      string(ctx.ActorType),
		PrimaryType:   bcontext.EntityType(strings.ToUpper(in.EntityType)),
		PrimaryID:     in.EntityID,
		CorrelationID: ctx.RequestID,
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
		ResourceType: in.EntityType,
		ResourceID:   fmt.Sprintf("%d", in.EntityID),
		Summary:      fmt.Sprintf("Unified business context retrieved for %s #%d", in.EntityType, in.EntityID),
		Data:         bCtx,
	}, nil
}

// ── Grounded Intelligence Insight Action (context.get_insight) ───────────────

type GetInsightInput struct {
	EntityType string `json:"entity_type"`
	EntityID   int64  `json:"entity_id"`
	Question   string `json:"question,omitempty"`
}

type GetInsightAction struct {
	svc bcontext.Service
}

func NewGetInsightAction(svc bcontext.Service) *GetInsightAction {
	return &GetInsightAction{svc: svc}
}

func (a *GetInsightAction) Name() string                         { return "context.get_insight" }
func (a *GetInsightAction) Module() string                       { return "intelligence" }
func (a *GetInsightAction) Description() string                  { return "Generate grounded, read-only AI business intelligence citing real records and fields." }
func (a *GetInsightAction) Category() ActionCategory             { return ActionCategoryRead }
func (a *GetInsightAction) InputSchema() interface{}             { return &GetInsightInput{} }
func (a *GetInsightAction) RequiresConfirmation() bool           { return false }
func (a *GetInsightAction) RequiredPermission() (string, string) { return rbac.ResourceCompanies, rbac.ActionRead }

func (a *GetInsightAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in GetInsightInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.EntityID <= 0 || in.EntityType == "" {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "entity_type and entity_id (>0) are required"},
		}, nil
	}

	insight, err := a.svc.GenerateInsight(ctx.Context, bcontext.InsightRequest{
		OrgID:         ctx.OrganizationID,
		UserID:        ctx.ActingUserID,
		UserRole:      string(ctx.ActorType),
		EntityType:    bcontext.EntityType(strings.ToUpper(in.EntityType)),
		EntityID:      in.EntityID,
		Question:      in.Question,
		CorrelationID: ctx.RequestID,
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
		ResourceType: in.EntityType,
		ResourceID:   fmt.Sprintf("%d", in.EntityID),
		Summary:      insight.Summary,
		Data:         insight,
	}, nil
}

// ── Customer 360° Intelligence Action (customer.get_intelligence) ───────────

type GetCustomerIntelligenceInput struct {
	CustomerID int64 `json:"customer_id"`
}

type GetCustomerIntelligenceAction struct {
	svc bcontext.Service
}

func NewGetCustomerIntelligenceAction(svc bcontext.Service) *GetCustomerIntelligenceAction {
	return &GetCustomerIntelligenceAction{svc: svc}
}

func (a *GetCustomerIntelligenceAction) Name() string                         { return "customer.get_intelligence" }
func (a *GetCustomerIntelligenceAction) Module() string                       { return "customers" }
func (a *GetCustomerIntelligenceAction) Description() string                  { return "Assemble 360° customer context, deterministic commercial & operational metrics, and grounded intelligence." }
func (a *GetCustomerIntelligenceAction) Category() ActionCategory             { return ActionCategoryRead }
func (a *GetCustomerIntelligenceAction) InputSchema() interface{}             { return &GetCustomerIntelligenceInput{} }
func (a *GetCustomerIntelligenceAction) RequiresConfirmation() bool           { return false }
func (a *GetCustomerIntelligenceAction) RequiredPermission() (string, string) { return rbac.ResourceCompanies, rbac.ActionRead }

func (a *GetCustomerIntelligenceAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in GetCustomerIntelligenceInput
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
			Error:   &ActionError{Type: "Validation", Message: "customer_id (>0) is required"},
		}, nil
	}

	intel, err := a.svc.GetCustomer360Intelligence(ctx.Context, ctx.OrganizationID, in.CustomerID, ctx.RequestID, ctx.ActingUserID)
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
		ResourceType: "CUSTOMER",
		ResourceID:   fmt.Sprintf("%d", in.CustomerID),
		Summary:      fmt.Sprintf("Customer 360° Intelligence generated for customer #%d (%s)", in.CustomerID, intel.Customer.Name),
		Data:         intel,
	}, nil
}

// ── RFQ and Pricing Intelligence Action (rfq.get_intelligence) ───────────────

type GetRFQIntelligenceInput struct {
	RFQID int64 `json:"rfq_id"`
}

type GetRFQIntelligenceAction struct {
	svc bcontext.Service
}

func NewGetRFQIntelligenceAction(svc bcontext.Service) *GetRFQIntelligenceAction {
	return &GetRFQIntelligenceAction{svc: svc}
}

func (a *GetRFQIntelligenceAction) Name() string                         { return "rfq.get_intelligence" }
func (a *GetRFQIntelligenceAction) Module() string                       { return "rfq" }
func (a *GetRFQIntelligenceAction) Description() string                  { return "Assemble RFQ quality metrics, quotation comparison spread, commercial conversion, and grounded pricing intelligence." }
func (a *GetRFQIntelligenceAction) Category() ActionCategory             { return ActionCategoryRead }
func (a *GetRFQIntelligenceAction) InputSchema() interface{}             { return &GetRFQIntelligenceInput{} }
func (a *GetRFQIntelligenceAction) RequiresConfirmation() bool           { return false }
func (a *GetRFQIntelligenceAction) RequiredPermission() (string, string) { return rbac.ResourceRFQs, rbac.ActionRead }

func (a *GetRFQIntelligenceAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in GetRFQIntelligenceInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.RFQID <= 0 {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "rfq_id (>0) is required"},
		}, nil
	}

	intel, err := a.svc.GetRFQ360PricingIntelligence(ctx.Context, ctx.OrganizationID, in.RFQID, ctx.RequestID, ctx.ActingUserID)
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
		ResourceType: "RFQ",
		ResourceID:   fmt.Sprintf("%d", in.RFQID),
		Summary:      fmt.Sprintf("RFQ & Pricing Intelligence generated for %s (Quotes: %d, Margin Health: %s)", intel.Identity.RFQNumber, intel.QuotationComparison.TotalQuotationsReceived, intel.MarginAndRisk.MarginHealth),
		Data:         intel,
	}, nil
}

// ── Shipment Operations Intelligence Action (shipment.get_intelligence) ─────

type GetShipmentIntelligenceInput struct {
	ShipmentID int64 `json:"shipment_id"`
}

type GetShipmentIntelligenceAction struct {
	svc bcontext.Service
}

func NewGetShipmentIntelligenceAction(svc bcontext.Service) *GetShipmentIntelligenceAction {
	return &GetShipmentIntelligenceAction{svc: svc}
}

func (a *GetShipmentIntelligenceAction) Name() string                         { return "shipment.get_intelligence" }
func (a *GetShipmentIntelligenceAction) Module() string                       { return "shipments" }
func (a *GetShipmentIntelligenceAction) Description() string                  { return "Assemble operational milestone timeline, exception prioritization, schedule variance, and grounded shipment intelligence." }
func (a *GetShipmentIntelligenceAction) Category() ActionCategory             { return ActionCategoryRead }
func (a *GetShipmentIntelligenceAction) InputSchema() interface{}             { return &GetShipmentIntelligenceInput{} }
func (a *GetShipmentIntelligenceAction) RequiresConfirmation() bool           { return false }
func (a *GetShipmentIntelligenceAction) RequiredPermission() (string, string) { return rbac.ResourceShipments, rbac.ActionRead }

func (a *GetShipmentIntelligenceAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in GetShipmentIntelligenceInput
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
			Error:   &ActionError{Type: "Validation", Message: "shipment_id (>0) is required"},
		}, nil
	}

	intel, err := a.svc.GetShipment360OperationsIntelligence(ctx.Context, ctx.OrganizationID, in.ShipmentID, ctx.RequestID, ctx.ActingUserID)
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
		ResourceType: "SHIPMENT",
		ResourceID:   fmt.Sprintf("%d", in.ShipmentID),
		Summary:      fmt.Sprintf("Shipment & Operations Intelligence generated for %s (Status: %s, Risk: %s, Exceptions: %d)", intel.Identity.ShipmentNumber, intel.Identity.Status, intel.RiskIndicators.OverallRiskRating, intel.Exceptions.OpenExceptions),
		Data:         intel,
	}, nil
}

// ── Invoice Finance Intelligence Action (invoice.get_intelligence) ───────────

type GetInvoiceIntelligenceInput struct {
	InvoiceID int64 `json:"invoice_id"`
}

type GetInvoiceIntelligenceAction struct {
	svc bcontext.Service
}

func NewGetInvoiceIntelligenceAction(svc bcontext.Service) *GetInvoiceIntelligenceAction {
	return &GetInvoiceIntelligenceAction{svc: svc}
}

func (a *GetInvoiceIntelligenceAction) Name() string                         { return "invoice.get_intelligence" }
func (a *GetInvoiceIntelligenceAction) Module() string                       { return "finance" }
func (a *GetInvoiceIntelligenceAction) Description() string                  { return "Assemble receivables aging, payment exposure, margin audit, and grounded invoice finance intelligence." }
func (a *GetInvoiceIntelligenceAction) Category() ActionCategory             { return ActionCategoryRead }
func (a *GetInvoiceIntelligenceAction) InputSchema() interface{}             { return &GetInvoiceIntelligenceInput{} }
func (a *GetInvoiceIntelligenceAction) RequiresConfirmation() bool           { return false }
func (a *GetInvoiceIntelligenceAction) RequiredPermission() (string, string) { return rbac.ResourceFinance, rbac.ActionRead }

func (a *GetInvoiceIntelligenceAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in GetInvoiceIntelligenceInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.InvoiceID <= 0 {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "invoice_id (>0) is required"},
		}, nil
	}

	intel, err := a.svc.GetInvoice360FinanceIntelligence(ctx.Context, ctx.OrganizationID, in.InvoiceID, ctx.RequestID, ctx.ActingUserID)
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
		ResourceType: "INVOICE",
		ResourceID:   fmt.Sprintf("%d", in.InvoiceID),
		Summary:      fmt.Sprintf("Invoice & Finance Intelligence generated for %s (Status: %s, Total: %.2f %s, Risk: %s, Aging: %s)", intel.Identity.InvoiceNumber, intel.Identity.InvoiceStatus, intel.Identity.TotalAmount, intel.Identity.Currency, intel.RiskIndicators.OverallRiskRating, intel.Identity.AgingBucket),
		Data:         intel,
	}, nil
}

// ── Contract & Compliance Intelligence Action (contract.get_intelligence) ────

type GetContractIntelligenceInput struct {
	ContractID int64 `json:"contract_id"`
}

type GetContractIntelligenceAction struct {
	svc bcontext.Service
}

func NewGetContractIntelligenceAction(svc bcontext.Service) *GetContractIntelligenceAction {
	return &GetContractIntelligenceAction{svc: svc}
}

func (a *GetContractIntelligenceAction) Name() string                         { return "contract.get_intelligence" }
func (a *GetContractIntelligenceAction) Module() string                       { return "contracts" }
func (a *GetContractIntelligenceAction) Description() string                  { return "Assemble contract overview, commercial terms, obligations, compliance requirements, and risk intelligence." }
func (a *GetContractIntelligenceAction) Category() ActionCategory             { return ActionCategoryRead }
func (a *GetContractIntelligenceAction) InputSchema() interface{}             { return &GetContractIntelligenceInput{} }
func (a *GetContractIntelligenceAction) RequiresConfirmation() bool           { return false }
func (a *GetContractIntelligenceAction) RequiredPermission() (string, string) { return rbac.ResourceDocuments, rbac.ActionRead }

func (a *GetContractIntelligenceAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in GetContractIntelligenceInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
		}, nil
	}

	if in.ContractID <= 0 {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "Validation", Message: "contract_id (>0) is required"},
		}, nil
	}

	intel, err := a.svc.GetContract360ComplianceIntelligence(ctx.Context, ctx.OrganizationID, in.ContractID, ctx.RequestID, ctx.ActingUserID)
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
		ResourceType: "CONTRACT",
		ResourceID:   fmt.Sprintf("%d", in.ContractID),
		Summary:      fmt.Sprintf("Contract & Compliance Intelligence generated for %s (Status: %s, Party: %s, Risk: %s, Completeness: %d%%)", intel.Identity.ContractReference, intel.Identity.Status, intel.Identity.PartyName, intel.RiskIndicators.OverallRiskRating, intel.Identity.CompletenessScore),
		Data:         intel,
	}, nil
}

// ── Cross-Module Business Insights Action (insights.get_cross_module) ───────

type GetCrossModuleInsightsInput struct {
	EntityType string `json:"entity_type,omitempty"`
	EntityID   int64  `json:"entity_id,omitempty"`
}

type GetCrossModuleInsightsAction struct {
	svc bcontext.Service
}

func NewGetCrossModuleInsightsAction(svc bcontext.Service) *GetCrossModuleInsightsAction {
	return &GetCrossModuleInsightsAction{svc: svc}
}

func (a *GetCrossModuleInsightsAction) Name() string                         { return "insights.get_cross_module" }
func (a *GetCrossModuleInsightsAction) Module() string                       { return "insights" }
func (a *GetCrossModuleInsightsAction) Description() string                  { return "Assemble cross-module connected business signals, relationship evidence, and grounded risk insights." }
func (a *GetCrossModuleInsightsAction) Category() ActionCategory             { return ActionCategoryRead }
func (a *GetCrossModuleInsightsAction) InputSchema() interface{}             { return &GetCrossModuleInsightsInput{} }
func (a *GetCrossModuleInsightsAction) RequiresConfirmation() bool           { return false }
func (a *GetCrossModuleInsightsAction) RequiredPermission() (string, string) { return rbac.ResourceDashboard, rbac.ActionRead }

func (a *GetCrossModuleInsightsAction) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	var in GetCrossModuleInsightsInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return &ActionResult{
				Success: false,
				Action:  a.Name(),
				Error:   &ActionError{Type: "Validation", Message: "Invalid input schema"},
			}, nil
		}
	}

	result, err := a.svc.GetCrossModuleInsights(ctx.Context, ctx.OrganizationID, in.EntityType, in.EntityID, ctx.RequestID, ctx.ActingUserID)
	if err != nil {
		return &ActionResult{
			Success: false,
			Action:  a.Name(),
			Error:   &ActionError{Type: "BusinessRule", Message: err.Error()},
		}, nil
	}

	summary := fmt.Sprintf("Cross-Module Insights generated: %d signals (%d critical, %d high)", result.TotalInsightsCount, result.CriticalCount, result.HighCount)
	return &ActionResult{
		Success:      true,
		Action:       a.Name(),
		ResourceType: "INSIGHTS",
		ResourceID:   fmt.Sprintf("%s:%d", in.EntityType, in.EntityID),
		Summary:      summary,
		Data:         result,
	}, nil
}





