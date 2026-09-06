package rfq_test

// activity_engine_test.go — Task 11: RFQ Activity Timeline & Audit Trail Tests
// Tests the pure BuildRFQActivity aggregation engine and the GetActivity service method.

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/rfq/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func makeTestRFQWithLead() *spec.RFQ {
	origin := "Nhava Sheva (INNSA)"
	dest := "Hamburg (DEHAM)"
	incoterms := "FOB"
	now := time.Now().Add(30 * 24 * time.Hour)
	createdAt := time.Now().Add(-2 * time.Hour)
	updatedAt := time.Now().Add(-30 * time.Minute)
	weight := 12500.0
	vol := 60.0
	email := "buyer@globalexports.com"
	contactName := "Rajesh Kumar"
	leadID := int64(1268)

	return &spec.RFQ{
		ID:                  101,
		OrgID:               1,
		RFQNumber:           "RFQ-20260827-001",
		CustomerID:          42,
		CustomerName:        "Global Exports Ltd",
		CustomerEmail:       &email,
		CustomerContactName: &contactName,
		Stage:               spec.StageRFQCreated,
		Origin:              &origin,
		Destination:         &dest,
		Incoterms:           &incoterms,
		TargetDate:          &now,
		AgentStatus:         "COMPLETED",
		LeadID:              &leadID,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
		Items: []spec.RFQItem{
			{ID: 1, RFQID: 101, Description: "Industrial Machinery", Quantity: 2, WeightKG: &weight, VolumeCBM: &vol},
		},
		Quotes: []spec.Quote{
			{
				ID:          201,
				RFQID:       101,
				CarrierName: "Maersk Line",
				BuyPrice:    4200.00,
				SellPrice:   4850.00,
				Status:      "GENERATED",
				CreatedAt:   time.Now().Add(-15 * time.Minute),
				UpdatedAt:   time.Now().Add(-15 * time.Minute),
			},
		},
	}
}

// Test 1: RFQ Activity Retrieval — returns complete activity response and valid summary counts
func TestActivityEngine_Retrieval(t *testing.T) {
	rfqObj := makeTestRFQWithLead()
	rawTimeline := []spec.TimelineEvent{
		{
			ID:          "act-1",
			EntityType:  "LEAD",
			EntityID:    1268,
			Category:    "LEAD",
			Action:      "CREATED",
			Description: "Lead was manually added to the system.",
			Actor:       "Varun Kanade",
			Timestamp:   time.Now().Add(-3 * time.Hour),
		},
		{
			ID:          "inter-1",
			EntityType:  "LEAD",
			EntityID:    1268,
			Category:    "EMAIL",
			Action:      "EMAIL_INBOUND",
			Description: "INBOUND: RFQ 2x40ft HC Containers",
			Actor:       "buyer@globalexports.com",
			Timestamp:   time.Now().Add(-2*time.Hour - 30*time.Minute),
		},
		{
			ID:          "act-2",
			EntityType:  "LEAD",
			EntityID:    1268,
			Category:    "LEAD",
			Action:      "LEAD_CONVERTED",
			Description: "Lead #1268 was successfully converted into RFQ-20260827-001.",
			Actor:       "Operations Team",
			Timestamp:   time.Now().Add(-2 * time.Hour),
		},
	}

	reqEval := rfq.EvaluateRequirements(rfqObj)
	activityResp := rfq.BuildRFQActivity(rfqObj, rawTimeline, reqEval)

	require.NotNil(t, activityResp)
	assert.NotEmpty(t, activityResp.Events, "Should return normalized activity events")
	assert.GreaterOrEqual(t, activityResp.Summary.TotalEvents, 4, "Should have at least 4 events")
	assert.GreaterOrEqual(t, activityResp.Summary.CustomerEvents, 1, "Should have customer events")
	assert.GreaterOrEqual(t, activityResp.Summary.OperationalEvents, 2, "Should have operational events")
	assert.GreaterOrEqual(t, activityResp.Summary.RequirementsEvents, 1, "Should have requirements events")
	assert.GreaterOrEqual(t, activityResp.Summary.QuoteEvents, 1, "Should have quote events")
	assert.NotNil(t, activityResp.Summary.LatestActivityAt)
}

// Test 2: Organization Isolation — Cross-org RFQ access fails safely with error
func TestService_GetActivity_OrgIsolation(t *testing.T) {
	svc, mockRepo, _ := newServiceWithMocks(t)

	// Org 1 requests RFQ 999 belonging to Org 2 — GetRFQByID must reject
	mockRepo.EXPECT().
		GetRFQByID(gomock.Any(), int32(1), int32(999)).
		Return(nil, assert.AnError).
		Times(1)

	result, err := svc.GetActivity(context.Background(), 1, 999)
	require.Error(t, err, "Cross-org access must return an error")
	assert.Nil(t, result, "Cross-org result must be nil")
}

// Test 3: Lead-Originated RFQ — Verifies Lead creation and conversion events appear with SourceType = LEAD
func TestActivityEngine_LeadOriginatedRFQ(t *testing.T) {
	rfqObj := makeTestRFQWithLead()
	rawTimeline := []spec.TimelineEvent{
		{
			ID:          "act-lead-create",
			EntityType:  "LEAD",
			EntityID:    1268,
			Category:    "LEAD",
			Action:      "CREATED",
			Description: "Lead was manually added to the system.",
			Actor:       "System",
			Timestamp:   time.Now().Add(-4 * time.Hour),
		},
		{
			ID:          "act-lead-conv",
			EntityType:  "LEAD",
			EntityID:    1268,
			Category:    "LEAD",
			Action:      "LEAD_CONVERTED",
			Description: "Lead #1268 was converted to RFQ-20260827-001.",
			Actor:       "Operations Team",
			Timestamp:   time.Now().Add(-2 * time.Hour),
		},
	}

	reqEval := rfq.EvaluateRequirements(rfqObj)
	activityResp := rfq.BuildRFQActivity(rfqObj, rawTimeline, reqEval)

	require.NotNil(t, activityResp.LeadID)
	assert.Equal(t, int64(1268), *activityResp.LeadID)

	var foundLeadConverted bool
	for _, ev := range activityResp.Events {
		if ev.Type == spec.ActivityLeadConverted {
			foundLeadConverted = true
			assert.Equal(t, "LEAD", ev.SourceType)
			assert.Equal(t, "1268", ev.SourceID)
			assert.True(t, ev.IsImportant)
		}
	}
	assert.True(t, foundLeadConverted, "Lead converted event must be present in activity")
}

// Test 4: Manual RFQ — Works seamlessly without Lead history
func TestActivityEngine_ManualRFQ(t *testing.T) {
	origin := "Shanghai (CNSHA)"
	dest := "Los Angeles (USLAX)"
	incoterms := "CIF"
	now := time.Now().Add(45 * 24 * time.Hour)
	createdAt := time.Now().Add(-1 * time.Hour)
	weight := 8000.0
	vol := 35.0
	email := "contact@pacific.com"
	contactName := "John Smith"

	manualRFQ := &spec.RFQ{
		ID:                  200,
		OrgID:               2,
		RFQNumber:           "RFQ-MANUAL-001",
		CustomerID:          55,
		CustomerName:        "Pacific Trade Co",
		CustomerEmail:       &email,
		CustomerContactName: &contactName,
		Stage:               spec.StageRFQCreated,
		Origin:              &origin,
		Destination:         &dest,
		Incoterms:           &incoterms,
		TargetDate:          &now,
		AgentStatus:         "IDLE",
		LeadID:              nil, // No lead
		CreatedAt:           createdAt,
		UpdatedAt:           createdAt,
		Items: []spec.RFQItem{
			{ID: 10, RFQID: 200, Description: "Consumer Electronics", Quantity: 1, WeightKG: &weight, VolumeCBM: &vol},
		},
	}

	rawTimeline := []spec.TimelineEvent{} // Empty timeline for manual creation
	reqEval := rfq.EvaluateRequirements(manualRFQ)
	activityResp := rfq.BuildRFQActivity(manualRFQ, rawTimeline, reqEval)

	require.NotNil(t, activityResp)
	assert.Nil(t, activityResp.LeadID, "Manual RFQ must have nil LeadID")
	assert.NotEmpty(t, activityResp.Events, "Manual RFQ must still have RFQ created and requirements events")

	var foundRFQCreated bool
	for _, ev := range activityResp.Events {
		if ev.Type == spec.ActivityRFQCreated {
			foundRFQCreated = true
			assert.Equal(t, "RFQ", ev.SourceType)
			assert.Equal(t, "200", ev.SourceID)
			assert.Contains(t, ev.Description, "RFQ-MANUAL-001")
		}
	}
	assert.True(t, foundRFQCreated, "RFQ Created event must be synthesized for manual RFQs")
}

// Test 5: Customer Interaction — Inbound/outbound emails appear with accurate actor/direction
func TestActivityEngine_CustomerInteractions(t *testing.T) {
	rfqObj := makeTestRFQWithLead()
	rawTimeline := []spec.TimelineEvent{
		{
			ID:          "inter-inbound",
			EntityType:  "LEAD",
			EntityID:    1268,
			Category:    "EMAIL",
			Action:      "EMAIL_INBOUND",
			Description: "INBOUND: RFQ Inquiry",
			Actor:       "buyer@globalexports.com",
			Timestamp:   time.Now().Add(-2 * time.Hour),
		},
		{
			ID:          "inter-outbound",
			EntityType:  "LEAD",
			EntityID:    1268,
			Category:    "EMAIL",
			Action:      "EMAIL_OUTBOUND",
			Description: "OUTBOUND: Clarification requested on cargo dimensions",
			Actor:       "Operations Team",
			Timestamp:   time.Now().Add(-1 * time.Hour),
		},
	}

	reqEval := rfq.EvaluateRequirements(rfqObj)
	activityResp := rfq.BuildRFQActivity(rfqObj, rawTimeline, reqEval)

	var foundInbound, foundOutbound bool
	for _, ev := range activityResp.Events {
		if ev.ID == "inter-inbound" {
			foundInbound = true
			assert.Equal(t, spec.ActivityCatCustomer, ev.Category)
			assert.Equal(t, spec.ActorCustomer, ev.ActorType)
			assert.Equal(t, "buyer@globalexports.com", ev.ActorName)
		}
		if ev.ID == "inter-outbound" {
			foundOutbound = true
			assert.Equal(t, spec.ActivityCatOperations, ev.Category)
			assert.Equal(t, spec.ActorOperations, ev.ActorType)
		}
	}
	assert.True(t, foundInbound, "Inbound customer email must be categorized as CUSTOMER")
	assert.True(t, foundOutbound, "Outbound email must be categorized as OPERATIONS")
}

// Test 6: Requirements Integration — Task 10 requirements evaluation appears as operational milestone
func TestActivityEngine_RequirementsIntegration(t *testing.T) {
	rfqObj := makeTestRFQWithLead()
	rawTimeline := []spec.TimelineEvent{}
	reqEval := rfq.EvaluateRequirements(rfqObj)

	activityResp := rfq.BuildRFQActivity(rfqObj, rawTimeline, reqEval)

	var foundReqEvent bool
	for _, ev := range activityResp.Events {
		if ev.Type == spec.ActivityRequirementsEvaluated {
			foundReqEvent = true
			assert.Equal(t, spec.ActivityCatRequirements, ev.Category)
			assert.Contains(t, ev.Description, fmt.Sprintf("%d%%", reqEval.OperationalReadiness.ReadinessScore))
			assert.NotNil(t, ev.Metadata)
			assert.Equal(t, reqEval.OperationalReadiness.OverallStatus, ev.Metadata["overall_status"])
		}
	}
	assert.True(t, foundReqEvent, "Requirements evaluated event must be present")
}

// Test 7: Chronological Ordering — Events are strictly ordered descending (newest first)
func TestActivityEngine_ChronologicalOrdering(t *testing.T) {
	rfqObj := makeTestRFQWithLead()
	now := time.Now()
	rawTimeline := []spec.TimelineEvent{
		{ID: "ev-old", EntityType: "LEAD", EntityID: 1268, Action: "CREATED", Timestamp: now.Add(-5 * time.Hour), Actor: "System"},
		{ID: "ev-mid", EntityType: "LEAD", EntityID: 1268, Action: "EMAIL_INBOUND", Timestamp: now.Add(-3 * time.Hour), Actor: "Customer"},
		{ID: "ev-new", EntityType: "RFQ", EntityID: 101, Action: "QUOTE_GENERATED", Timestamp: now.Add(-10 * time.Minute), Actor: "Pricing"},
	}

	reqEval := rfq.EvaluateRequirements(rfqObj)
	activityResp := rfq.BuildRFQActivity(rfqObj, rawTimeline, reqEval)

	events := activityResp.Events
	require.GreaterOrEqual(t, len(events), 3)

	for i := 0; i < len(events)-1; i++ {
		assert.True(t, events[i].Timestamp.After(events[i+1].Timestamp) || events[i].Timestamp.Equal(events[i+1].Timestamp),
			"Event %d (%s) must be after or equal to event %d (%s)", i, events[i].Timestamp, i+1, events[i+1].Timestamp)
	}
}

// Test 8: Source Integrity — Verifies no activity references another org's data
func TestActivityEngine_SourceIntegrity(t *testing.T) {
	svc, mockRepo, _ := newServiceWithMocks(t)

	rfqObj := makeTestRFQWithLead()
	rfqObj.OrgID = 1 // Org 1

	// Service loads RFQ with Org 1
	mockRepo.EXPECT().
		GetRFQByID(gomock.Any(), int32(1), int32(101)).
		Return(rfqObj, nil).
		Times(1)

	// Timeline fetched with Org 1
	mockRepo.EXPECT().
		GetRFQTimeline(gomock.Any(), int32(1), int32(101), gomock.Any()).
		Return([]spec.TimelineEvent{
			{ID: "act-1", EntityType: "RFQ", EntityID: 101, Action: "CREATED", Actor: "System", Timestamp: time.Now()},
		}, nil).
		Times(1)

	resp, err := svc.GetActivity(context.Background(), 1, 101)
	require.NoError(t, err)
	require.NotNil(t, resp)

	for _, ev := range resp.Events {
		assert.Equal(t, "101", ev.RelatedEntityID, "All events must be strictly scoped to RFQ 101")
	}
}
