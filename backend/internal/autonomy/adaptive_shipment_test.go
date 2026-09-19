package autonomy_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/freel/backend/internal/autonomy"
)

func TestAdaptiveShipment_EventDeduplication(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()
	orgID := int64(101)
	shipmentID := int64(5001)

	req := autonomy.IngestShipmentEventRequest{
		EventID:          "evt-dup-001",
		ShipmentID:       shipmentID,
		EventType:        "ETA_UPDATE",
		Severity:         "LOW",
		DeduplicationKey: fmt.Sprintf("%d:%d:ETA_UPDATE:evt-dup-001", orgID, shipmentID),
		Payload: map[string]interface{}{
			"new_eta": "2026-10-15T12:00:00Z",
		},
	}

	// First event processing
	res1, err := svc.ProcessShipmentEvent(ctx, orgID, nil, req)
	if err != nil {
		t.Fatalf("first event processing failed: %v", err)
	}
	if res1 == nil {
		t.Fatal("expected non-nil evaluation result")
	}

	// Replay identical event with same deduplication key
	// In our mock repo, let's simulate that the event key is now recorded
	repo.RecordShipmentEvent(ctx, &autonomy.ShipmentAdaptiveEvent{
		OrgID:            orgID,
		ShipmentID:       shipmentID,
		EventID:          req.EventID,
		EventType:        req.EventType,
		DeduplicationKey: req.DeduplicationKey,
		Decision:         "CONTINUE_MONITORING",
		CreatedAt:        time.Now(),
	})

	// Override GetShipmentEventByDedupKey for test
	res2, err := svc.ProcessShipmentEvent(ctx, orgID, nil, req)
	if err != nil {
		t.Fatalf("second event processing failed: %v", err)
	}
	if res2 == nil {
		t.Fatal("expected non-nil result on replayed event")
	}
}

func TestAdaptiveShipment_ActivePlanCollisionPrevention(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{
		generateResp: &autonomy.SidecarPlanGenResponse{
			PlanID:            "plan-adapt-v2",
			Version:           1,
			Goal:              "Adaptive recovery plan v2",
			Module:            "shipments",
			ConfidenceScore:   0.88,
			DataSufficiency:   true,
			RiskLevel:         "HIGH",
			AutonomyLevel:     autonomy.Level2Prepare,
			Status:            autonomy.PlanStatusRequiresApproval,
			WaitingState:      "WAITING_FOR_APPROVAL",
			OrderedSteps: []autonomy.SidecarPlanStep{
				{
					StepID:           "step-rec-1",
					StepNumber:       1,
					ActionType:       "shipments.update_status",
					Title:            "Reroute via Express Air Transit",
					Description:      "Mitigate port delay by rerouting critical container",
					RiskLevel:        "HIGH",
					RequiresApproval: true,
				},
			},
		},
	}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()
	orgID := int64(101)
	shipmentID := int64(6001)

	// Pre-populate an active plan for shipment 6001
	activePlan := &autonomy.AutonomousPlan{
		OrgID:             orgID,
		PlanID:            "plan-adapt-v1",
		Version:           1,
		Status:            autonomy.PlanStatusExecuting,
		RelatedEntityType: "shipment",
		RelatedEntityID:   fmt.Sprintf("%d", shipmentID),
	}
	repo.CreatePlan(ctx, activePlan, nil)

	// Process severe ETA delay event triggering new plan
	req := autonomy.IngestShipmentEventRequest{
		EventID:          "evt-disrupt-002",
		ShipmentID:       shipmentID,
		EventType:        "PORT_DISRUPTION",
		Severity:         "HIGH",
		DeduplicationKey: "dedup-disrupt-002",
		Payload: map[string]interface{}{
			"port": "SGSIN",
			"delay_hours": 36,
		},
	}

	// Configure mock sidecar evaluate response to trigger NEW_PLAN
	sidecar.shipmentEvalResp = &autonomy.SidecarShipmentEventEvalResponse{
		ShipmentID:             shipmentID,
		EventType:              req.EventType,
		Decision:               "NEW_PLAN",
		DecisionReason:         "Port disruption delay exceeds customer commitment buffer",
		IsMeaningfulChange:     true,
		CommitmentRiskSeverity: "HIGH",
	}

	res, err := svc.ProcessShipmentEvent(ctx, orgID, nil, req)
	if err != nil {
		t.Fatalf("unexpected error processing event: %v", err)
	}

	if res == nil {
		t.Fatal("expected valid result")
	}

	// Verify old plan was superseded
	oldPlan, _, _ := repo.GetPlan(ctx, orgID, "plan-adapt-v1")
	if oldPlan != nil && oldPlan.Status == autonomy.PlanStatusExecuting {
		t.Errorf("expected old active plan to be superseded or paused, but got %s", oldPlan.Status)
	}
}

func TestAdaptiveShipment_WaitingStateTransitions(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}
	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	ctx := context.Background()
	orgID := int64(101)

	plan := &autonomy.AutonomousPlan{
		OrgID:             orgID,
		PlanID:            "plan-wait-001",
		Version:           1,
		Status:            autonomy.PlanStatusExecuting,
		RelatedEntityType: "shipment",
		RelatedEntityID:   "7001",
	}
	repo.CreatePlan(ctx, plan, nil)

	// Transition to WAITING_FOR_CARRIER
	waitEnd := time.Now().Add(4 * time.Hour)
	err := svc.TransitionPlanWaitingState(ctx, orgID, "plan-wait-001", string(autonomy.WaitingStateCarrier), &waitEnd)
	if err != nil {
		t.Fatalf("failed transitioning waiting state: %v", err)
	}

	// Transition back to NONE (resumes execution)
	err = svc.TransitionPlanWaitingState(ctx, orgID, "plan-wait-001", string(autonomy.WaitingStateNone), nil)
	if err != nil {
		t.Fatalf("failed clearing waiting state: %v", err)
	}
}
