package ai

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

// Mock provider for testing gateway failover
type testProvider struct {
	response string
	err      error
	invoked  bool
}

func (p *testProvider) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	p.invoked = true
	if p.err != nil {
		return "", p.err
	}
	return p.response, nil
}

func TestRuntimeConfig_Validation(t *testing.T) {
	t.Run("Valid Development Config", func(t *testing.T) {
		cfg := &RuntimeConfig{
			Environment:       "development",
			PrimaryProvider:   "gemini",
			PrimaryModel:      "gemini-1.5-flash",
			FailoverProvider:  "openai",
			FailoverModel:     "gpt-4o-mini",
			AllowMockFallback: true,
		}
		if err := cfg.Validate(); err != nil {
			t.Fatalf("expected valid config, got: %v", err)
		}
	})

	t.Run("Production Rejection of Mock Mode", func(t *testing.T) {
		cfg := &RuntimeConfig{
			Environment:       "production",
			PrimaryProvider:   "gemini",
			PrimaryModel:      "gemini-1.5-flash",
			AllowMockFallback: true, // Forbidden!
		}
		err := cfg.Validate()
		if err == nil || !errors.Is(err, ErrProductionMockMode) {
			t.Fatalf("expected ErrProductionMockMode, got: %v", err)
		}
	})

	t.Run("Production Rejection of Missing Key", func(t *testing.T) {
		os.Unsetenv("GOOGLE_API_KEY")
		os.Unsetenv("GEMINI_API_KEY")
		cfg := &RuntimeConfig{
			Environment:       "production",
			PrimaryProvider:   "gemini",
			PrimaryModel:      "gemini-1.5-flash",
			AllowMockFallback: false,
		}
		err := cfg.Validate()
		if err == nil || !errors.Is(err, ErrMissingProviderKey) {
			t.Fatalf("expected ErrMissingProviderKey, got: %v", err)
		}
	})

	t.Run("Invalid Model Identifier Rejection", func(t *testing.T) {
		cfg := &RuntimeConfig{
			Environment:     "development",
			PrimaryProvider: "gemini",
			PrimaryModel:    "unsupported-hallucinated-model-v99",
		}
		err := cfg.Validate()
		if err == nil || !errors.Is(err, ErrUnsupportedModel) {
			t.Fatalf("expected ErrUnsupportedModel, got: %v", err)
		}
	})
}

func TestPromptManager_VersioningAndAliases(t *testing.T) {
	pm := NewPromptManager()

	t.Run("Resolve Canonical Key", func(t *testing.T) {
		text, err := pm.GetPrompt("pricing.analyst", map[string]interface{}{
			"Origin": "INNSA",
		})
		if err != nil {
			t.Fatalf("failed to get pricing.analyst prompt: %v", err)
		}
		if !strings.Contains(text, "Senior Pricing Analyst") {
			t.Errorf("prompt content did not match expected: %s", text[:100])
		}
	})

	t.Run("Resolve Legacy Alias score_lead", func(t *testing.T) {
		text, err := pm.GetPrompt("score_lead", map[string]interface{}{
			"CompanyName": "Acme Logistics",
		})
		if err != nil {
			t.Fatalf("failed to resolve legacy alias score_lead: %v", err)
		}
		if !strings.Contains(text, "Acme Logistics") {
			t.Errorf("expected Acme Logistics in prompt text, got: %s", text)
		}
	})

	t.Run("Missing Prompt Fails Loudly", func(t *testing.T) {
		_, err := pm.GetPrompt("nonexistent.prompt.key", nil)
		if err == nil {
			t.Fatal("expected error for nonexistent prompt, got nil")
		}
	})

	t.Run("List Prompts Covers All Workflows", func(t *testing.T) {
		prompts := pm.ListPrompts()
		if len(prompts) < 8 {
			t.Fatalf("expected at least 8 prompts, found: %d", len(prompts))
		}
		workflows := make(map[string]bool)
		for _, p := range prompts {
			workflows[p.WorkflowName] = true
		}
		requiredWorkflows := []string{"pricing", "sales", "operations", "contracts", "compliance", "finance", "leads", "outreach"}
		for _, rw := range requiredWorkflows {
			if !workflows[rw] {
				t.Errorf("missing prompt for workflow: %s", rw)
			}
		}
	})
}

func TestErrorClassificationAndRedaction(t *testing.T) {
	t.Run("Redact Sensitive Secrets", func(t *testing.T) {
		input := "Error calling api with key AIzaSyA1234567890abcdef1234567890abcdef and sk-1234567890abcdef1234567890abcdef"
		sanitized := RedactSecrets(input)
		if strings.Contains(sanitized, "AIzaSyA1234567890") || strings.Contains(sanitized, "sk-1234567890") {
			t.Errorf("failed to redact secrets: %s", sanitized)
		}
		if !strings.Contains(sanitized, "AIza[REDACTED]") || !strings.Contains(sanitized, "sk-[REDACTED]") {
			t.Errorf("redaction marker not found: %s", sanitized)
		}
	})

	t.Run("Classify Rate Limit 429", func(t *testing.T) {
		err := errors.New("google api returned status 429: quota exhausted")
		cat := ClassifyError(err)
		if cat != ErrorCategoryRateLimit {
			t.Errorf("expected ErrorCategoryRateLimit, got: %s", cat)
		}
	})

	t.Run("Classify Gateway Timeout", func(t *testing.T) {
		err := errors.New("context deadline exceeded (timeout)")
		cat := ClassifyError(err)
		if cat != ErrorCategoryTimeout {
			t.Errorf("expected ErrorCategoryTimeout, got: %s", cat)
		}
	})

	t.Run("Classify Unavailable 503", func(t *testing.T) {
		err := errors.New("http 503 service unavailable")
		cat := ClassifyError(err)
		if cat != ErrorCategoryUnavailable {
			t.Errorf("expected ErrorCategoryUnavailable, got: %s", cat)
		}
	})
}

func TestGateway_FailoverBehavior(t *testing.T) {
	t.Run("Primary Provider Success", func(t *testing.T) {
		primary := &testProvider{response: "primary success", err: nil}
		failover := &testProvider{response: "failover success", err: nil}

		cfg := &RuntimeConfig{
			Environment:       "development",
			PrimaryProvider:   "gemini",
			PrimaryModel:      "gemini-1.5-flash",
			FailoverProvider:  "openai",
			FailoverModel:     "gpt-4o-mini",
			AllowMockFallback: false,
		}

		gw := NewGatewayWithConfig(map[string]Provider{
			"gemini": primary,
			"openai": failover,
		}, cfg, nil)

		execCtx := &ExecutionContext{
			OrgID:        2,
			WorkflowName: "pricing",
			RequestID:    "req-test-1",
		}
		res, err := gw.ExecutePromptWithContext(context.Background(), execCtx, "test prompt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.RawOutput != "primary success" || res.FinalProvider != "gemini" {
			t.Errorf("expected primary success, got: %+v", res)
		}
		if res.FailoverOccurred {
			t.Error("failover should not have occurred")
		}
		if failover.invoked {
			t.Error("failover provider should not have been invoked")
		}
	})

	t.Run("Primary Failure Triggers Failover", func(t *testing.T) {
		primary := &testProvider{response: "", err: errors.New("rate limit 429 exceeded")}
		failover := &testProvider{response: "fallback response", err: nil}

		cfg := &RuntimeConfig{
			Environment:       "development",
			PrimaryProvider:   "gemini",
			PrimaryModel:      "gemini-1.5-flash",
			FailoverProvider:  "openai",
			FailoverModel:     "gpt-4o-mini",
			AllowMockFallback: false,
		}

		gw := NewGatewayWithConfig(map[string]Provider{
			"gemini": primary,
			"openai": failover,
		}, cfg, nil)

		execCtx := &ExecutionContext{
			OrgID:        2,
			WorkflowName: "pricing",
			RequestID:    "req-test-2",
		}
		res, err := gw.ExecutePromptWithContext(context.Background(), execCtx, "test prompt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.RawOutput != "fallback response" || res.FinalProvider != "openai" {
			t.Errorf("expected fallback response from openai, got: %+v", res)
		}
		if !res.FailoverOccurred {
			t.Error("failover flag should be true")
		}
	})

	t.Run("Both Providers Fail Without Mock Fallback Returns Error", func(t *testing.T) {
		primary := &testProvider{response: "", err: errors.New("primary down")}
		failover := &testProvider{response: "", err: errors.New("failover down")}

		cfg := &RuntimeConfig{
			Environment:       "production", // Production prohibits mock fallback!
			PrimaryProvider:   "gemini",
			PrimaryModel:      "gemini-1.5-flash",
			FailoverProvider:  "openai",
			FailoverModel:     "gpt-4o-mini",
			AllowMockFallback: false,
		}

		gw := NewGatewayWithConfig(map[string]Provider{
			"gemini": primary,
			"openai": failover,
		}, cfg, nil)

		execCtx := &ExecutionContext{
			OrgID:        2,
			WorkflowName: "contracts",
			RequestID:    "req-test-3",
		}
		_, err := gw.ExecutePromptWithContext(context.Background(), execCtx, "test prompt")
		if err == nil || !errors.Is(err, ErrProvidersUnavailable) {
			t.Fatalf("expected ErrProvidersUnavailable, got: %v", err)
		}
	})
}
