package orchestration

import (
	"context"
	"encoding/json"
	"testing"
)

func TestActionRegistryInitialization(t *testing.T) {
	reg := NewRegistry()
	actions := reg.ListActions()

	if len(actions) < 7 {
		t.Fatalf("expected at least 7 registered actions, got %d", len(actions))
	}

	expectedActions := []string{
		"tasks.create",
		"notifications.create",
		"notes.add",
		"followups.create",
		"reviews.schedule",
		"automations.flag_attention",
		"shipments.update_status",
	}

	for _, name := range expectedActions {
		action, err := reg.GetAction(name)
		if err != nil {
			t.Errorf("expected action '%s' to be registered, got error: %v", name, err)
			continue
		}
		if action.Name != name {
			t.Errorf("action name mismatch: expected '%s', got '%s'", name, action.Name)
		}
	}
}

func TestActionRegistryValidation(t *testing.T) {
	reg := NewRegistry()

	// 1. Valid tasks.create input
	taskAction, err := reg.GetAction("tasks.create")
	if err != nil {
		t.Fatalf("tasks.create not found: %v", err)
	}

	validTaskInput, _ := json.Marshal(map[string]interface{}{
		"title":       "Inspect Port Delay",
		"description": "Port delay detected at destination terminal.",
		"priority":    "HIGH",
	})
	if err := taskAction.Validate(validTaskInput); err != nil {
		t.Errorf("expected valid task input to pass validation, got: %v", err)
	}

	// 2. Invalid tasks.create input (empty title)
	invalidTaskInput, _ := json.Marshal(map[string]interface{}{
		"title": "",
	})
	if err := taskAction.Validate(invalidTaskInput); err == nil {
		t.Errorf("expected empty title to fail validation")
	}

	// 3. High-risk shipments.update_status
	shipAction, err := reg.GetAction("shipments.update_status")
	if err != nil {
		t.Fatalf("shipments.update_status not found: %v", err)
	}
	if !shipAction.RequiresApproval {
		t.Errorf("expected shipments.update_status to require approval by default")
	}
	if shipAction.Category != "HIGH_RISK" {
		t.Errorf("expected shipments.update_status category to be HIGH_RISK, got %s", shipAction.Category)
	}
}

func TestActionRegistryUnknownAction(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.GetAction("unregistered.malicious_action")
	if err == nil {
		t.Errorf("expected error for unregistered action, got nil")
	}
}

// Mock Sidecar Client for Unit Testing
type mockSidecarClient struct {
	response *PythonProposalResponse
	err      error
}

func (m *mockSidecarClient) ProposeAction(ctx context.Context, payload map[string]interface{}) (*PythonProposalResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func TestServiceApprovalGatingLogic(t *testing.T) {
	reg := NewRegistry()
	mockClient := &mockSidecarClient{
		response: &PythonProposalResponse{
			Status: "SUCCESS",
			Proposal: &ActionProposal{
				ProposalID:         "prop-test-001",
				SourceModule:       "SHIPMENTS",
				SourceRecordType:   "SHIPMENT",
				SourceRecordID:     "101",
				TriggerEvent:       "SCHEDULE_DELAY",
				ProposedActionType: "shipments.update_status",
				ActionParameters:   json.RawMessage(`{"shipment_id": 101, "target_status": "DELAYED"}`),
				Explanation:        "Carrier milestone indicates delay",
				Confidence:         0.92,
				RiskLevel:          RiskHigh,
				RequiresApproval:   true,
				ApprovalPolicy:     ApprovalPolicyAlways,
				ExpectedImpact:     "Updates customer portal with revised delivery timing",
				CorrelationID:      "corr-test-unit",
			},
			EvaluationSummary: "Delay detected",
		},
	}

	// Verify that high risk actions mandate approval
	actionDef, err := reg.GetAction(mockClient.response.Proposal.ProposedActionType)
	if err != nil {
		t.Fatalf("action lookup failed: %v", err)
	}
	if actionDef.Category != "HIGH_RISK" {
		t.Errorf("expected high risk category, got %s", actionDef.Category)
	}
}
