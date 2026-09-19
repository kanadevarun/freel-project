package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	bcontext "github.com/freel/backend/internal/context"
	"github.com/freel/backend/internal/rbac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockContextService struct {
	getCtxFunc     func(ctx context.Context, req bcontext.ContextRequest) (*bcontext.BusinessContext, error)
	genInsightFunc func(ctx context.Context, req bcontext.InsightRequest) (*bcontext.IntelligenceInsight, error)
}

func (m *mockContextService) GetBusinessContext(ctx context.Context, req bcontext.ContextRequest) (*bcontext.BusinessContext, error) {
	if m.getCtxFunc != nil {
		return m.getCtxFunc(ctx, req)
	}
	return &bcontext.BusinessContext{
		OrgID:         req.OrgID,
		PrimaryType:   req.PrimaryType,
		PrimaryID:     req.PrimaryID,
		CorrelationID: req.CorrelationID,
		IsReadOnly:    true,
		DataFreshness: time.Now().UTC(),
	}, nil
}

func (m *mockContextService) GenerateInsight(ctx context.Context, req bcontext.InsightRequest) (*bcontext.IntelligenceInsight, error) {
	if m.genInsightFunc != nil {
		return m.genInsightFunc(ctx, req)
	}
	return &bcontext.IntelligenceInsight{
		Title:               "Test Insight",
		Summary:             "Test summary for verified record",
		ConfidenceLevel:     "HIGH",
		IsInformationalOnly: true,
		CorrelationID:       req.CorrelationID,
	}, nil
}

func (m *mockContextService) GetCustomer360Intelligence(ctx context.Context, orgID int64, customerID int64, correlationID string, requestingUserID int64) (*bcontext.Customer360Intelligence, error) {
	return &bcontext.Customer360Intelligence{
		OrgID:         orgID,
		CustomerID:    customerID,
		CorrelationID: correlationID,
		IsReadOnly:    true,
		Customer: bcontext.CustomerIdentitySummary{
			ID:           customerID,
			OrgID:        orgID,
			Name:         "Acme Freight Test",
			CustomerCode: fmt.Sprintf("CUST-%d", customerID),
			Status:       "ACTIVE",
		},
		AISummary: bcontext.CustomerAISummary{
			ExecutiveSummary:    "Acme Freight Test is an active account in good standing.",
			ConfidenceLevel:     "HIGH",
			IsInformationalOnly: true,
		},
	}, nil
}

func (m *mockContextService) GetRFQ360PricingIntelligence(ctx context.Context, orgID int64, rfqID int64, correlationID string, requestingUserID int64) (*bcontext.RFQ360PricingIntelligence, error) {
	return &bcontext.RFQ360PricingIntelligence{
		OrgID:         orgID,
		RFQID:         rfqID,
		CorrelationID: correlationID,
		IsReadOnly:    true,
		Identity: bcontext.RFQIdentitySummary{
			ID:           rfqID,
			OrgID:        orgID,
			RFQNumber:    fmt.Sprintf("RFQ-2026-%d", rfqID),
			CustomerName: "Test Customer",
			Stage:        "WON",
			Origin:       "INNSA",
			Destination:  "DEHAM",
		},
		QuotationComparison: bcontext.RFQQuotationComparison{
			TotalQuotationsReceived: 2,
			ValidQuotationsCount:    2,
		},
		AISummary: bcontext.RFQAISummary{
			ExecutiveSummary: "RFQ has 2 valid quotations with 16.7% profit margin.",
			ConfidenceLevel:  "HIGH",
			Classification:   "READ_ONLY_INFORMATIONAL",
		},
	}, nil
}

func (m *mockContextService) GetShipment360OperationsIntelligence(ctx context.Context, orgID int64, shipmentID int64, correlationID string, requestingUserID int64) (*bcontext.Shipment360OperationsIntelligence, error) {
	return &bcontext.Shipment360OperationsIntelligence{
		OrgID:         orgID,
		ShipmentID:    shipmentID,
		CorrelationID: correlationID,
		Identity: &bcontext.ShipmentIdentitySummary{
			ShipmentID:     shipmentID,
			OrgID:          orgID,
			ShipmentNumber: fmt.Sprintf("SH-%d", shipmentID),
			Status:         "IN_TRANSIT",
			OriginPort:     "INNSA",
			DestinationPort: "NLRTM",
			CarrierSCAC:    "MAEU",
		},
		Milestones: &bcontext.ShipmentMilestoneIntelligence{
			TotalMilestones:     4,
			CompletedMilestones: 2,
		},
		Exceptions: &bcontext.ShipmentExceptionIntelligence{
			OpenExceptions: 1,
		},
		RiskIndicators: &bcontext.ShipmentRiskIndicators{
			OverallRiskRating: "MODERATE",
		},
	}, nil
}

func (m *mockContextService) GetOrgOperationsSummary(ctx context.Context, orgID int64, correlationID string, requestingUserID int64) (*bcontext.OrgOperationsSummary, error) {
	return &bcontext.OrgOperationsSummary{
		OrgID:           orgID,
		CorrelationID:   correlationID,
		TotalShipments:  10,
		ActiveShipments: 8,
	}, nil
}

func (m *mockContextService) GetInvoice360FinanceIntelligence(ctx context.Context, orgID int64, invoiceID int64, correlationID string, requestingUserID int64) (*bcontext.Invoice360FinanceIntelligence, error) {
	return &bcontext.Invoice360FinanceIntelligence{
		InvoiceID:         invoiceID,
		OrganizationScope: orgID,
		CorrelationID:     correlationID,
		ReadOnly:          true,
		Identity: bcontext.InvoiceIdentitySummary{
			InvoiceID:     invoiceID,
			InvoiceNumber: fmt.Sprintf("INV-%d", invoiceID),
			CustomerID:    1,
			CustomerName:  "Acme Corp",
			TotalAmount:   15000.0,
			PaidAmount:    5000.0,
			BalanceDue:    10000.0,
			InvoiceStatus: "Partially Paid",
			Currency:      "USD",
			AgingBucket:   "1-30_DAYS",
			IsOverdue:     true,
		},
		RiskIndicators: bcontext.FinanceRiskIndicators{
			OverallRiskRating: "HIGH",
			OverallRiskScore:  45,
			IsOverdue:         true,
		},
	}, nil
}

func (m *mockContextService) GetOrgFinanceSummary(ctx context.Context, orgID int64, correlationID string, requestingUserID int64) (*bcontext.OrgFinanceSummary, error) {
	return &bcontext.OrgFinanceSummary{
		OrganizationScope:           orgID,
		ActiveInvoicesCount:         12,
		OpenInvoicesCount:           5,
		TotalOutstandingReceivables: 45000.0,
		ReceivablesHealthRating:     "GOOD",
		ReadOnly:                    true,
	}, nil
}

func (m *mockContextService) GetContract360ComplianceIntelligence(ctx context.Context, orgID int64, contractID int64, correlationID string, requestingUserID int64) (*bcontext.Contract360ComplianceIntelligence, error) {
	return &bcontext.Contract360ComplianceIntelligence{
		ContractID:    contractID,
		ReadOnly:      true,
		CorrelationID: correlationID,
		Identity: bcontext.ContractIdentitySummary{
			ContractID:        contractID,
			ContractReference: fmt.Sprintf("CNT-%d", contractID),
			PartyName:         "Apex Logistics",
			Status:            "ACTIVE",
			CompletenessScore: 90,
		},
		RiskIndicators: bcontext.ContractRiskIndicators{
			OverallRiskRating: "LOW",
			OverallRiskScore:  15,
		},
	}, nil
}

func (m *mockContextService) GetContractCoverageForEntity(ctx context.Context, orgID int64, entityType string, entityID int64, correlationID string, requestingUserID int64) (*bcontext.ContractCoverageEvaluation, error) {
	return &bcontext.ContractCoverageEvaluation{
		TargetEntityType:      entityType,
		TargetEntityID:        entityID,
		HasApplicableContract: true,
		CoverageStatus:        "FULLY_COVERED",
		ReadOnly:              true,
	}, nil
}

func (m *mockContextService) GetOrgContractComplianceSummary(ctx context.Context, orgID int64, correlationID string, requestingUserID int64) (*bcontext.OrgContractComplianceSummary, error) {
	return &bcontext.OrgContractComplianceSummary{
		TotalContracts:          5,
		ActiveContracts:         4,
		OverallComplianceHealth: "GOOD",
		ReadOnly:                true,
	}, nil
}

func (m *mockContextService) GetCrossModuleInsights(ctx context.Context, orgID int64, entityType string, entityID int64, correlationID string, requestingUserID int64) (*bcontext.CrossModuleInsightsResult, error) {
	return &bcontext.CrossModuleInsightsResult{
		OrganizationScope:  orgID,
		CorrelationID:      correlationID,
		ReadOnly:           true,
		TotalInsightsCount: 2,
		CriticalCount:      1,
		HighCount:          1,
		Insights: []bcontext.CrossModuleInsight{
			{
				InsightID:       "ins-1",
				InsightType:     bcontext.TypeOperationalFinance,
				Severity:        bcontext.SeverityCritical,
				PriorityScore:   90,
				Title:           "Delayed Shipment with Overdue Invoice",
				IsDeterministic: true,
				ReadOnly:        true,
			},
		},
		ConnectedModules: []string{"finance", "shipments"},
	}, nil
}

func (m *mockContextService) GetOrgCrossModuleSummary(ctx context.Context, orgID int64, correlationID string, requestingUserID int64) (*bcontext.OrgCrossModuleSummary, error) {
	return &bcontext.OrgCrossModuleSummary{
		OrganizationScope: orgID,
		CorrelationID:     correlationID,
		TotalInsights:     2,
		CriticalInsights:  1,
		HighInsights:      1,
		ReadOnly:          true,
	}, nil
}

func TestContextActions_Metadata(t *testing.T) {
	mockSvc := &mockContextService{}
	getAction := NewGetContextAction(mockSvc)
	insightAction := NewGetInsightAction(mockSvc)
	custAction := NewGetCustomerIntelligenceAction(mockSvc)
	rfqAction := NewGetRFQIntelligenceAction(mockSvc)
	shAction := NewGetShipmentIntelligenceAction(mockSvc)
	invAction := NewGetInvoiceIntelligenceAction(mockSvc)
	cntAction := NewGetContractIntelligenceAction(mockSvc)

	// Verify read-only classification
	assert.Equal(t, "context.get", getAction.Name())
	assert.Equal(t, ActionCategoryRead, getAction.Category())
	assert.False(t, getAction.RequiresConfirmation())

	assert.Equal(t, "context.get_insight", insightAction.Name())
	assert.Equal(t, ActionCategoryRead, insightAction.Category())
	assert.False(t, insightAction.RequiresConfirmation())

	assert.Equal(t, "customer.get_intelligence", custAction.Name())
	assert.Equal(t, ActionCategoryRead, custAction.Category())
	assert.False(t, custAction.RequiresConfirmation())

	assert.Equal(t, "rfq.get_intelligence", rfqAction.Name())
	assert.Equal(t, ActionCategoryRead, rfqAction.Category())
	assert.False(t, rfqAction.RequiresConfirmation())

	assert.Equal(t, "shipment.get_intelligence", shAction.Name())
	assert.Equal(t, ActionCategoryRead, shAction.Category())
	assert.False(t, shAction.RequiresConfirmation())

	assert.Equal(t, "invoice.get_intelligence", invAction.Name())
	assert.Equal(t, ActionCategoryRead, invAction.Category())
	assert.False(t, invAction.RequiresConfirmation())

	assert.Equal(t, "contract.get_intelligence", cntAction.Name())
	assert.Equal(t, ActionCategoryRead, cntAction.Category())
	assert.False(t, cntAction.RequiresConfirmation())
}

func TestGetContextAction_Execute(t *testing.T) {
	mockSvc := &mockContextService{}
	action := NewGetContextAction(mockSvc)
	ctx := NewActionContext(context.Background(), 10, 5, ActorTypeAIAgent, "req-123")

	// 1. Invalid input (empty entity type)
	inBad, _ := json.Marshal(GetContextInput{EntityType: "", EntityID: 1})
	resBad, err := action.Execute(ctx, inBad)
	require.NoError(t, err)
	assert.False(t, resBad.Success)
	assert.Equal(t, "Validation", resBad.Error.Type)

	// 2. Successful execution
	inGood, _ := json.Marshal(GetContextInput{EntityType: "CUSTOMER", EntityID: 101})
	resGood, err := action.Execute(ctx, inGood)
	require.NoError(t, err)
	assert.True(t, resGood.Success)
	assert.Equal(t, "context.get", resGood.Action)
	assert.NotNil(t, resGood.Data)

	bCtx, ok := resGood.Data.(*bcontext.BusinessContext)
	require.True(t, ok)
	assert.Equal(t, int64(10), bCtx.OrgID)
	assert.Equal(t, bcontext.EntityTypeCustomer, bCtx.PrimaryType)
	assert.Equal(t, int64(101), bCtx.PrimaryID)
	assert.True(t, bCtx.IsReadOnly)
}

func TestGetInsightAction_Execute(t *testing.T) {
	mockSvc := &mockContextService{}
	action := NewGetInsightAction(mockSvc)
	ctx := NewActionContext(context.Background(), 10, 5, ActorTypeAIAgent, "req-123")

	inGood, _ := json.Marshal(GetInsightInput{EntityType: "RFQ", EntityID: 42, Question: "Summarize status"})
	res, err := action.Execute(ctx, inGood)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "context.get_insight", res.Action)

	insight, ok := res.Data.(*bcontext.IntelligenceInsight)
	require.True(t, ok)
	assert.True(t, insight.IsInformationalOnly)
	assert.Equal(t, "HIGH", insight.ConfidenceLevel)
	assert.Equal(t, "Test Insight", insight.Title)
}

func TestGetCustomerIntelligenceAction_Execute(t *testing.T) {
	mockSvc := &mockContextService{}
	action := NewGetCustomerIntelligenceAction(mockSvc)
	ctx := NewActionContext(context.Background(), 2, 10, ActorTypeAIAgent, "req-cust-789")

	// 1. Validation failure: missing customer_id
	inBad, _ := json.Marshal(GetCustomerIntelligenceInput{CustomerID: 0})
	resBad, err := action.Execute(ctx, inBad)
	require.NoError(t, err)
	assert.False(t, resBad.Success)
	assert.Equal(t, "Validation", resBad.Error.Type)

	// 2. Success
	inGood, _ := json.Marshal(GetCustomerIntelligenceInput{CustomerID: 101})
	resGood, err := action.Execute(ctx, inGood)
	require.NoError(t, err)
	assert.True(t, resGood.Success)
	assert.Equal(t, "customer.get_intelligence", resGood.Action)

	intel, ok := resGood.Data.(*bcontext.Customer360Intelligence)
	require.True(t, ok)
	assert.Equal(t, int64(2), intel.OrgID)
	assert.Equal(t, int64(101), intel.CustomerID)
	assert.Equal(t, "Acme Freight Test", intel.Customer.Name)
	assert.True(t, intel.IsReadOnly)
	assert.True(t, intel.AISummary.IsInformationalOnly)
}

func TestGetRFQIntelligenceAction_Execute(t *testing.T) {
	mockSvc := &mockContextService{}
	action := NewGetRFQIntelligenceAction(mockSvc)
	ctx := NewActionContext(context.Background(), 1, 10, ActorTypeAIAgent, "req-rfq-111")

	// 1. Validation failure: missing rfq_id
	inBad, _ := json.Marshal(GetRFQIntelligenceInput{RFQID: 0})
	resBad, err := action.Execute(ctx, inBad)
	require.NoError(t, err)
	assert.False(t, resBad.Success)
	assert.Equal(t, "Validation", resBad.Error.Type)

	// 2. Success
	inGood, _ := json.Marshal(GetRFQIntelligenceInput{RFQID: 1})
	resGood, err := action.Execute(ctx, inGood)
	require.NoError(t, err)
	assert.True(t, resGood.Success)
	assert.Equal(t, "rfq.get_intelligence", resGood.Action)

	intel, ok := resGood.Data.(*bcontext.RFQ360PricingIntelligence)
	require.True(t, ok)
	assert.Equal(t, int64(1), intel.OrgID)
	assert.Equal(t, int64(1), intel.RFQID)
	assert.Equal(t, "RFQ-2026-1", intel.Identity.RFQNumber)
	assert.True(t, intel.IsReadOnly)
	assert.Equal(t, "READ_ONLY_INFORMATIONAL", intel.AISummary.Classification)
}

func TestGetShipmentIntelligenceAction_Execute(t *testing.T) {
	mockSvc := &mockContextService{}
	action := NewGetShipmentIntelligenceAction(mockSvc)
	ctx := NewActionContext(context.Background(), 2, 10, ActorTypeAIAgent, "req-sh-101")

	// 1. Validation failure: missing shipment_id
	inBad, _ := json.Marshal(GetShipmentIntelligenceInput{ShipmentID: 0})
	resBad, err := action.Execute(ctx, inBad)
	require.NoError(t, err)
	assert.False(t, resBad.Success)
	assert.Equal(t, "Validation", resBad.Error.Type)

	// 2. Success
	inGood, _ := json.Marshal(GetShipmentIntelligenceInput{ShipmentID: 101})
	resGood, err := action.Execute(ctx, inGood)
	require.NoError(t, err)
	assert.True(t, resGood.Success)
	assert.Equal(t, "shipment.get_intelligence", resGood.Action)

	intel, ok := resGood.Data.(*bcontext.Shipment360OperationsIntelligence)
	require.True(t, ok)
	assert.Equal(t, int64(2), intel.OrgID)
	assert.Equal(t, int64(101), intel.ShipmentID)
	assert.Equal(t, "SH-101", intel.Identity.ShipmentNumber)
	assert.Equal(t, "IN_TRANSIT", intel.Identity.Status)
	assert.Equal(t, "MODERATE", intel.RiskIndicators.OverallRiskRating)
}

func TestGetInvoiceIntelligenceAction_Execute(t *testing.T) {
	mockSvc := &mockContextService{}
	action := NewGetInvoiceIntelligenceAction(mockSvc)
	ctx := NewActionContext(context.Background(), 1, 10, ActorTypeAIAgent, "req-inv-1")

	// 1. Validation failure: missing invoice_id
	inBad, _ := json.Marshal(GetInvoiceIntelligenceInput{InvoiceID: 0})
	resBad, err := action.Execute(ctx, inBad)
	require.NoError(t, err)
	assert.False(t, resBad.Success)
	assert.Equal(t, "Validation", resBad.Error.Type)

	// 2. Success
	inGood, _ := json.Marshal(GetInvoiceIntelligenceInput{InvoiceID: 1})
	resGood, err := action.Execute(ctx, inGood)
	require.NoError(t, err)
	assert.True(t, resGood.Success)
	assert.Equal(t, "invoice.get_intelligence", resGood.Action)

	intel, ok := resGood.Data.(*bcontext.Invoice360FinanceIntelligence)
	require.True(t, ok)
	assert.Equal(t, int64(1), intel.OrganizationScope)
	assert.Equal(t, int64(1), intel.InvoiceID)
	assert.Equal(t, "INV-1", intel.Identity.InvoiceNumber)
	assert.Equal(t, "Partially Paid", intel.Identity.InvoiceStatus)
	assert.Equal(t, "HIGH", intel.RiskIndicators.OverallRiskRating)
	assert.Equal(t, "1-30_DAYS", intel.Identity.AgingBucket)
	assert.True(t, intel.ReadOnly)
}

func TestGetContractIntelligenceAction_Execute(t *testing.T) {
	mockSvc := &mockContextService{}
	action := NewGetContractIntelligenceAction(mockSvc)
	ctx := NewActionContext(context.Background(), 1, 10, ActorTypeAIAgent, "req-cnt-1")

	// 1. Validation failure: missing contract_id
	inBad, _ := json.Marshal(GetContractIntelligenceInput{ContractID: 0})
	resBad, err := action.Execute(ctx, inBad)
	require.NoError(t, err)
	assert.False(t, resBad.Success)
	assert.Equal(t, "Validation", resBad.Error.Type)

	// 2. Success
	inGood, _ := json.Marshal(GetContractIntelligenceInput{ContractID: 5})
	resGood, err := action.Execute(ctx, inGood)
	require.NoError(t, err)
	assert.True(t, resGood.Success)
	assert.Equal(t, "contract.get_intelligence", resGood.Action)

	intel, ok := resGood.Data.(*bcontext.Contract360ComplianceIntelligence)
	require.True(t, ok)
	assert.Equal(t, int64(5), intel.ContractID)
	assert.Equal(t, "CNT-5", intel.Identity.ContractReference)
	assert.Equal(t, "Apex Logistics", intel.Identity.PartyName)
	assert.Equal(t, "ACTIVE", intel.Identity.Status)
	assert.Equal(t, "LOW", intel.RiskIndicators.OverallRiskRating)
	assert.True(t, intel.ReadOnly)
}

func TestGetCrossModuleInsightsAction_Metadata(t *testing.T) {
	mockSvc := &mockContextService{}
	action := NewGetCrossModuleInsightsAction(mockSvc)

	assert.Equal(t, "insights.get_cross_module", action.Name())
	assert.Equal(t, "insights", action.Module())
	assert.Equal(t, ActionCategoryRead, action.Category())
	assert.False(t, action.RequiresConfirmation())

	res, act := action.RequiredPermission()
	assert.Equal(t, rbac.ResourceDashboard, res)
	assert.Equal(t, rbac.ActionRead, act)
}

func TestGetCrossModuleInsightsAction_Execute(t *testing.T) {
	mockSvc := &mockContextService{}
	action := NewGetCrossModuleInsightsAction(mockSvc)

	ctx := NewActionContext(context.Background(), 1, 42, ActorTypeAIAgent, "req-cross-test-1")

	inGood, _ := json.Marshal(GetCrossModuleInsightsInput{EntityType: "CUSTOMER", EntityID: 10})
	resGood, err := action.Execute(ctx, inGood)
	require.NoError(t, err)
	assert.True(t, resGood.Success)
	assert.Equal(t, "insights.get_cross_module", resGood.Action)

	result, ok := resGood.Data.(*bcontext.CrossModuleInsightsResult)
	require.True(t, ok)
	assert.Equal(t, 2, result.TotalInsightsCount)
	assert.Equal(t, 1, result.CriticalCount)
	assert.Equal(t, 1, result.HighCount)
	assert.True(t, result.ReadOnly)
}





