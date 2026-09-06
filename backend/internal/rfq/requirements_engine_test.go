package rfq_test

// requirements_engine_test.go — Task 10: Requirements Engine Tests
// Tests the pure EvaluateRequirements function directly (no mocks, no DB).
// Uses only real spec.RFQ fields that exist in the actual database schema.

import (
	"context"
	"testing"
	"time"

	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/rfq/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// makeCompleteTestRFQ builds a fully-complete spec.RFQ using ONLY real schema fields:
//   rfqs: origin, destination, incoterms, target_date, stage, agent_status, lead_id, customer_*
//   rfq_items: description, weight_kg, volume_cbm, quantity
func makeCompleteTestRFQ() *spec.RFQ {
	origin := "Nhava Sheva (INNSA)"
	dest := "Hamburg (DEHAM)"
	incoterms := "FOB"
	now := time.Now().Add(30 * 24 * time.Hour)
	weight := 12500.0
	vol := 60.0
	email := "buyer@globalexports.com"
	phone := "+91-9876543210"
	contactName := "Rajesh Kumar"
	leadID := int64(171)
	return &spec.RFQ{
		ID:                  101,
		OrgID:               1,
		RFQNumber:           "RFQ-20260827-001",
		CustomerID:          42,
		CustomerName:        "Global Exports Ltd",
		CustomerEmail:       &email,
		CustomerPhone:       &phone,
		CustomerContactName: &contactName,
		Stage:               spec.StageRFQCreated,
		Origin:              &origin,
		Destination:         &dest,
		Incoterms:           &incoterms,
		TargetDate:          &now,
		AgentStatus:         "COMPLETED",
		LeadID:              &leadID,
		Items: []spec.RFQItem{
			{
				ID:          1,
				RFQID:       101,
				Description: "Industrial Machinery & Tooling",
				Quantity:    2,
				WeightKG:    &weight,
				VolumeCBM:   &vol,
			},
		},
	}
}

// Test 1: Fully complete RFQ → READY_FOR_QUOTATION or ATTENTION_REQUIRED (0 blockers, 0 missing required)
func TestRequirementsEngine_CompleteRFQ(t *testing.T) {
	rfqObj := makeCompleteTestRFQ()
	result := rfq.EvaluateRequirements(rfqObj)

	require.NotNil(t, result)
	// A fully complete RFQ should have NO blocking requirements and NO missing required items.
	// It may still have ATTENTION_REQUIRED due to conditional items (e.g. container type confirmation
	// when Incoterms = FOB). Both READY_FOR_QUOTATION and ATTENTION_REQUIRED are valid proceed-able states.
	assert.Contains(t,
		[]string{spec.ReadinessReadyForQuotation, spec.ReadinessAttentionRequired},
		result.OperationalReadiness.OverallStatus,
		"Fully complete RFQ should be READY_FOR_QUOTATION or ATTENTION_REQUIRED (no blocking items)")
	assert.Equal(t, 0, result.OperationalReadiness.BlockingCount,
		"Complete RFQ should have 0 blocking requirements")
	assert.Equal(t, 0, result.OperationalReadiness.MissingRequiredCount,
		"Complete RFQ should have 0 missing REQUIRED items")
	assert.NotEmpty(t, result.OperationalReadiness.NextBestAction)
	assert.NotEmpty(t, result.Groups)
	assert.Greater(t, result.OperationalReadiness.ReadinessScore, 50)
}

// Test 2: Missing origin/destination/cargo → INFORMATION_REQUIRED, blocking > 0
func TestRequirementsEngine_MissingCoreFields(t *testing.T) {
	rfqObj := &spec.RFQ{
		ID:           102,
		OrgID:        1,
		Stage:        spec.StageRFQCreated,
		CustomerName: "Partial Customer",
		// No origin, destination, incoterms, target_date, items
	}
	result := rfq.EvaluateRequirements(rfqObj)

	require.NotNil(t, result)
	assert.Equal(t, spec.ReadinessInformationRequired, result.OperationalReadiness.OverallStatus,
		"RFQ with missing core fields should be INFORMATION_REQUIRED")
	assert.Greater(t, result.OperationalReadiness.BlockingCount, 3,
		"Should have multiple blocking requirements (origin, destination, incoterms, weight, volume, etc.)")
}

// Test 3A: DG keywords in rfq_items.description → DG conditional activates
func TestRequirementsEngine_DGConditionalActivates(t *testing.T) {
	rfqObj := makeCompleteTestRFQ()
	weight := 5000.0
	vol := 20.0
	rfqObj.Items = []spec.RFQItem{
		{Description: "Lithium Battery Packs (dangerous goods class 9)", Quantity: 100, WeightKG: &weight, VolumeCBM: &vol},
	}
	result := rfq.EvaluateRequirements(rfqObj)

	found := false
	for _, g := range result.Groups {
		if g.Category == spec.CategoryConditionalCompliance {
			for _, r := range g.Requirements {
				if r.ID == "dg_declaration" {
					assert.Equal(t, spec.ReqStatusMissing, r.Status,
						"DG declaration should be MISSING when DG keywords detected in rfq_items.description")
					assert.Equal(t, spec.SeverityConditional, r.Severity)
					found = true
				}
			}
		}
	}
	assert.True(t, found, "DG conditional requirement should be present")
}

// Test 3B: No DG keywords → DG requirement NOT_APPLICABLE
func TestRequirementsEngine_DGConditionalNotApplicable(t *testing.T) {
	rfqObj := makeCompleteTestRFQ() // "Industrial Machinery & Tooling" — no DG keywords
	result := rfq.EvaluateRequirements(rfqObj)

	for _, g := range result.Groups {
		if g.Category == spec.CategoryConditionalCompliance {
			for _, r := range g.Requirements {
				if r.ID == "dg_declaration" {
					assert.Equal(t, spec.ReqStatusNotApplicable, r.Status,
						"DG declaration should be NOT_APPLICABLE for non-DG cargo")
				}
			}
		}
	}
}

// Test 4: HBL, MBL, Air Waybill are NOT_APPLICABLE at RFQ stage (not blockers)
func TestRequirementsEngine_FutureDocumentsNotBlocking(t *testing.T) {
	rfqObj := makeCompleteTestRFQ()
	result := rfq.EvaluateRequirements(rfqObj)

	for _, doc := range result.DocumentRequirements {
		switch doc.DocType {
		case "hbl", "mbl", "air_waybill":
			assert.Equal(t, spec.ReqStatusNotApplicable, doc.Status,
				"%s must be NOT_APPLICABLE at RFQ stage — must never appear as a blocker", doc.DocType)
			assert.False(t, doc.IsRequired,
				"%s must not be required at RFQ stage", doc.DocType)
			assert.NotEqual(t, spec.DocStageRFQ, doc.ApplicableStage,
				"%s must not be applicable at current RFQ stage", doc.DocType)
		case "commercial_invoice", "packing_list":
			assert.Equal(t, spec.DocStageRFQ, doc.ApplicableStage,
				"%s should be applicable at current RFQ stage", doc.DocType)
			assert.True(t, doc.IsRequired)
		}
	}
}

// Test 5: AI findings use real agent_status field; include confidence; don't override deterministic reqs
func TestRequirementsEngine_AIFindings(t *testing.T) {
	// COMPLETED agent_status → should produce findings
	rfqObj := makeCompleteTestRFQ()
	rfqObj.AgentStatus = "COMPLETED"
	result := rfq.EvaluateRequirements(rfqObj)

	assert.NotEmpty(t, result.AIFindings)
	for _, f := range result.AIFindings {
		assert.NotEmpty(t, f.ID)
		assert.Contains(t, []string{"HIGH", "MEDIUM", "LOW"}, f.Confidence,
			"All AI findings must have a valid confidence level")
	}

	// FAILED agent_status → human review required
	rfqObj2 := makeCompleteTestRFQ()
	rfqObj2.AgentStatus = "FAILED"
	result2 := rfq.EvaluateRequirements(rfqObj2)
	hasReviewRequired := false
	for _, f := range result2.AIFindings {
		if f.RequiresHumanReview {
			hasReviewRequired = true
		}
	}
	assert.True(t, hasReviewRequired)

	// IDLE agent_status → no AI processing findings
	rfqObj3 := makeCompleteTestRFQ()
	rfqObj3.AgentStatus = "IDLE"
	rfqObj3.LeadID = nil // Manual RFQ
	result3 := rfq.EvaluateRequirements(rfqObj3)
	for _, f := range result3.AIFindings {
		assert.NotEqual(t, "ai-shipment-extraction", f.ID, "IDLE status should not produce extraction finding")
		assert.NotEqual(t, "ai-processing", f.ID, "IDLE status should not produce processing finding")
	}
}

// Test 6: Lead-originated RFQ — lead_id preserved in response, source context set
func TestRequirementsEngine_LeadOriginatedRFQ(t *testing.T) {
	rfqObj := makeCompleteTestRFQ()
	leadID := int64(171)
	rfqObj.LeadID = &leadID // Real rfqs.lead_id column

	result := rfq.EvaluateRequirements(rfqObj)

	require.NotNil(t, result.LeadID)
	assert.Equal(t, int64(171), *result.LeadID)
}

// Test 7: Manual RFQ (no lead, IDLE agent_status) — works without lead/email/AI
func TestRequirementsEngine_ManualRFQ(t *testing.T) {
	origin := "Shanghai (CNSHA)"
	dest := "Los Angeles (USLAX)"
	incoterms := "CIF"
	now := time.Now().Add(45 * 24 * time.Hour)
	weight := 8000.0
	vol := 35.0
	email := "contact@pacific.com"
	contactName := "John Smith"

	rfqObj := &spec.RFQ{
		ID:                  200,
		OrgID:               2,
		RFQNumber:           "RFQ-20260827-MANUAL",
		CustomerID:          55,
		CustomerName:        "Pacific Trade Co",
		CustomerEmail:       &email,
		CustomerContactName: &contactName,
		Stage:               spec.StageRFQCreated,
		Origin:              &origin,
		Destination:         &dest,
		Incoterms:           &incoterms,
		TargetDate:          &now,
		AgentStatus:         "IDLE", // No AI involvement
		LeadID:              nil,    // Manual creation — no rfqs.lead_id
		Items: []spec.RFQItem{
			{Description: "Consumer Electronics", Quantity: 1, WeightKG: &weight, VolumeCBM: &vol},
		},
	}

	result := rfq.EvaluateRequirements(rfqObj)

	require.NotNil(t, result)
	assert.Nil(t, result.LeadID, "Manual RFQ should have nil LeadID in requirements response")

	// No AI processing findings for IDLE
	for _, f := range result.AIFindings {
		assert.NotEqual(t, "ai-shipment-extraction", f.ID)
		assert.NotEqual(t, "ai-processing", f.ID)
	}

	// Should still evaluate requirements correctly
	assert.NotEmpty(t, result.Groups)
}

// Test 8: Org isolation — GetRequirements via business logic returns error for cross-org access
func TestService_GetRequirements_OrgIsolation(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc, mockRepo, _ := newServiceWithMocks(t)
	_ = ctrl

	// Org 1 requests RFQ 999 which belongs to org 2 — GetRFQByID returns error
	mockRepo.EXPECT().
		GetRFQByID(gomock.Any(), int32(1), int32(999)).
		Return(nil, assert.AnError).
		Times(1)

	result, err := svc.GetRequirements(context.Background(), 1, 999)
	require.Error(t, err, "Cross-org requirements access should return an error")
	assert.Nil(t, result, "Cross-org requirements access should return nil response")
}
