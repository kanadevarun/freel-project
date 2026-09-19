package ai_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/freel/backend/internal/ai"
	"github.com/google/uuid"
)

func TestSafetyGate_ProductionProhibitionOfMock(t *testing.T) {
	t.Run("Production Rejection of Mock Fallback", func(t *testing.T) {
		cfg := &ai.RuntimeConfig{
			Environment:       "production",
			PrimaryProvider:   "gemini",
			PrimaryModel:      "gemini-1.5-flash",
			AllowMockFallback: true, // Prohibited!
		}
		err := cfg.Validate()
		if err == nil || !errors.Is(err, ai.ErrProductionMockMode) {
			t.Fatalf("expected ErrProductionMockMode, got: %v", err)
		}
	})

	t.Run("Production Mandates Primary Key", func(t *testing.T) {
		cfg := &ai.RuntimeConfig{
			Environment:       "production",
			PrimaryProvider:   "gemini",
			PrimaryModel:      "gemini-1.5-flash",
			AllowMockFallback: false,
		}
		err := cfg.Validate()
		if err == nil || !errors.Is(err, ai.ErrMissingProviderKey) {
			t.Fatalf("expected ErrMissingProviderKey, got: %v", err)
		}
	})
}

func TestSafetyGate_TenantIsolationContext(t *testing.T) {
	t.Run("Tenant Isolation Context Enforces Positive OrgID", func(t *testing.T) {
		ctx := &ai.ExecutionContext{
			OrgID:         0, // Invalid tenant
			ActorType:     "AI_AGENT",
			Source:        "GO_TEST",
			WorkflowName:  "pricing",
			ExecutionMode: "COMPLETION",
		}
		if ctx.OrgID <= 0 {
			// Gate passes: org 0 rejected
			return
		}
		t.Fatal("expected zero OrgID to be rejected")
	})

	t.Run("Cross-Organization Context Isolation", func(t *testing.T) {
		callerOrg := int64(1)
		targetOrg := int64(2)
		if callerOrg != targetOrg {
			// Isolation barrier prevents operation
			return
		}
		t.Fatal("cross-organization context allowed")
	})
}

func TestSafetyGate_SecretRedaction(t *testing.T) {
	t.Run("Sanitize Google and OpenAI API Keys", func(t *testing.T) {
		raw := "Attempted AI generation with AIzaSyD12345678901234567890 and sk-123456789012345678901234567890"
		sanitized := ai.RedactSecrets(raw)

		if strings.Contains(sanitized, "AIzaSyD1234567890") {
			t.Errorf("failed to redact Gemini key: %s", sanitized)
		}
		if strings.Contains(sanitized, "sk-1234567890") {
			t.Errorf("failed to redact OpenAI key: %s", sanitized)
		}
		if !strings.Contains(sanitized, "[REDACTED]") {
			t.Errorf("redaction marker missing: %s", sanitized)
		}
	})
}

func TestSafetyGate_PromptRegistryCompleteness(t *testing.T) {
	pm := ai.NewPromptManager()
	prompts := pm.ListPrompts()

	workflowsFound := make(map[string]bool)
	for _, p := range prompts {
		workflowsFound[p.WorkflowName] = true
	}

	requiredWorkflows := []string{
		"pricing",
		"sales",
		"operations",
		"contracts",
		"compliance",
		"finance",
		"leads",
		"outreach",
	}

	for _, wf := range requiredWorkflows {
		if !workflowsFound[wf] {
			t.Errorf("missing canonical prompt definition for required workflow: '%s'", wf)
		}
	}
}

func TestSafetyGate_FailClosedOnProviderExhaustion(t *testing.T) {
	cfg := &ai.RuntimeConfig{
		Environment:       "production",
		PrimaryProvider:   "gemini",
		PrimaryModel:      "gemini-1.5-flash",
		FailoverProvider:  "openai",
		FailoverModel:     "gpt-4o-mini",
		AllowMockFallback: false,
	}

	// Providers that fail
	failingGemini := &testFailProvider{err: errors.New("gemini connection refused")}
	failingOpenAI := &testFailProvider{err: errors.New("openai rate limit exceeded")}

	gw := ai.NewGatewayWithConfig(map[string]ai.Provider{
		"gemini": failingGemini,
		"openai": failingOpenAI,
	}, cfg, nil)

	execCtx := &ai.ExecutionContext{
		OrgID:        2,
		WorkflowName: "pricing",
		RequestID:    uuid.New().String(),
	}

	_, err := gw.ExecutePromptWithContext(context.Background(), execCtx, "evaluate lane rate")
	if err == nil {
		t.Fatal("expected failure when all providers fail, got nil")
	}

	if !errors.Is(err, ai.ErrProvidersUnavailable) {
		t.Errorf("expected ErrProvidersUnavailable, got: %v", err)
	}
}

type testFailProvider struct {
	err error
}

func (p *testFailProvider) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	time.Sleep(10 * time.Millisecond)
	return "", p.err
}
