package actions_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/freel/backend/internal/actions"
)

type MockActionInput struct {
	Value string `json:"value"`
}

type MockWriteAction struct{}

func (m *MockWriteAction) Name() string { return "mock.write" }
func (m *MockWriteAction) Module() string { return "mock" }
func (m *MockWriteAction) Description() string { return "A mock write action." }
func (m *MockWriteAction) Category() actions.ActionCategory { return actions.ActionCategoryWrite }
func (m *MockWriteAction) InputSchema() interface{} { return &MockActionInput{} }
func (m *MockWriteAction) RequiresConfirmation() bool { return false }

func (m *MockWriteAction) Execute(ctx *actions.ActionContext, input []byte) (*actions.ActionResult, error) {
	if ctx.ActorType == "" {
		return &actions.ActionResult{Success: false, Error: &actions.ActionError{Type: "Unauthorized", Message: "Missing actor type"}}, nil
	}
	
	if ctx.OrganizationID == 0 {
		return &actions.ActionResult{Success: false, Error: &actions.ActionError{Type: "Unauthorized", Message: "Missing organization context"}}, nil
	}

	var in MockActionInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &actions.ActionResult{Success: false, Error: &actions.ActionError{Type: "Validation", Message: "Invalid input"}}, nil
	}

	if in.Value == "" {
		return &actions.ActionResult{Success: false, Error: &actions.ActionError{Type: "Validation", Message: "value is required"}}, nil
	}

	return &actions.ActionResult{
		Success:      true,
		Action:       m.Name(),
		ResourceType: "Mock",
		Summary:      "Mock executed by " + string(ctx.ActorType),
		Data:         map[string]string{"received": in.Value},
	}, nil
}

type MockIdempotencyStore struct {
	store map[string]*actions.ActionResult
}

func (m *MockIdempotencyStore) Start(ctx context.Context, orgID int64, key string) (bool, *actions.ActionResult, error) {
	if res, exists := m.store[key]; exists {
		return true, res, nil
	}
	return false, nil, nil
}

func (m *MockIdempotencyStore) Complete(ctx context.Context, orgID int64, key string, result *actions.ActionResult, ttl time.Duration) error {
	m.store[key] = result
	return nil
}

func TestActionRegistryAndExecution(t *testing.T) {
	registry := actions.NewRegistry()
	mockAction := &MockWriteAction{}
	
	if err := registry.Register(mockAction); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}

	action, err := registry.GetAction("mock.write")
	if err != nil {
		t.Fatalf("Failed to retrieve action: %v", err)
	}

	ctx := actions.NewActionContext(context.Background(), 1, 100, actions.ActorTypeAIAgent, "req-123")
	
	// Test Validation Failure
	invalidInput := []byte(`{}`)
	res, err := action.Execute(ctx, invalidInput)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if res.Success {
		t.Fatalf("Expected validation failure")
	}
	if res.Error.Type != "Validation" {
		t.Fatalf("Expected Validation error, got %s", res.Error.Type)
	}

	// Test Success
	validInput := []byte(`{"value": "test-data"}`)
	res, err = action.Execute(ctx, validInput)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatalf("Expected success, got error: %v", res.Error)
	}
	if res.Summary != "Mock executed by AI_AGENT" {
		t.Fatalf("Unexpected summary: %s", res.Summary)
	}
}

func TestIdempotency(t *testing.T) {
	store := &MockIdempotencyStore{store: make(map[string]*actions.ActionResult)}
	baseAction := &MockWriteAction{}
	
	idempotentAction := actions.NewIdempotentAction(store, baseAction, time.Minute)
	
	ctx := actions.NewActionContext(context.Background(), 1, 100, actions.ActorTypeAIAgent, "req-123").WithIdempotencyKey("idem-1")
	validInput := []byte(`{"value": "first-call"}`)
	
	// First call
	res1, err := idempotentAction.Execute(ctx, validInput)
	if err != nil || !res1.Success {
		t.Fatalf("First call failed")
	}
	
	// Second call with same idempotency key but different input
	validInput2 := []byte(`{"value": "second-call"}`)
	res2, err := idempotentAction.Execute(ctx, validInput2)
	if err != nil || !res2.Success {
		t.Fatalf("Second call failed")
	}
	
	// Result should be from the first call
	data1 := res1.Data.(map[string]string)
	data2 := res2.Data.(map[string]string)
	
	if data1["received"] != data2["received"] {
		t.Fatalf("Idempotency failed: expected %s, got %s", data1["received"], data2["received"])
	}
}
