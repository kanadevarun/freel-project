package notifications_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/freel/backend/internal/notifications"
)

// ============================================================================
// SECTION 5 & 8: ENVIRONMENT MATRIX & SERVICE KEY HARDENING TESTS
// ============================================================================

func TestResolveServiceKey_FullEnvironmentMatrix(t *testing.T) {
	// Save existing environment to restore after tests
	origToken := os.Getenv("INTERNAL_SERVICE_TOKEN")
	origKey := os.Getenv("INTERNAL_SERVICE_KEY")
	origSidecarKey := os.Getenv("AI_SIDECAR_SERVICE_KEY")
	origAppEnv := os.Getenv("APP_ENV")
	origEnv := os.Getenv("ENV")
	origEnvironment := os.Getenv("ENVIRONMENT")

	defer func() {
		os.Setenv("INTERNAL_SERVICE_TOKEN", origToken)
		os.Setenv("INTERNAL_SERVICE_KEY", origKey)
		os.Setenv("AI_SIDECAR_SERVICE_KEY", origSidecarKey)
		os.Setenv("APP_ENV", origAppEnv)
		os.Setenv("ENV", origEnv)
		os.Setenv("ENVIRONMENT", origEnvironment)
	}()

	const testCustomKey = "custom-configured-production-key-999"

	tests := []struct {
		name         string
		explicitKey  string
		envVarKey    string
		appEnv       string
		expectedKey  string
		allowDefault bool
	}{
		// 1. Development + key present -> custom key
		{
			name:        "development + key present -> works with configured key",
			explicitKey: "",
			envVarKey:   testCustomKey,
			appEnv:      "development",
			expectedKey: testCustomKey,
		},
		// 2. Development + key absent -> DefaultServiceKey (existing dev behavior)
		{
			name:         "development + key absent -> follows existing safe development behavior",
			explicitKey:  "",
			envVarKey:    "",
			appEnv:       "development",
			expectedKey:  notifications.DefaultServiceKey,
			allowDefault: true,
		},
		// 3. Test + key present -> custom key
		{
			name:        "test + key present -> works with configured key",
			explicitKey: "",
			envVarKey:   testCustomKey,
			appEnv:      "test",
			expectedKey: testCustomKey,
		},
		// 4. Test + key absent -> DefaultServiceKey (existing test behavior)
		{
			name:         "test + key absent -> follows existing safe test behavior",
			explicitKey:  "",
			envVarKey:    "",
			appEnv:       "test",
			expectedKey:  notifications.DefaultServiceKey,
			allowDefault: true,
		},
		// 5. Staging + key present -> custom key
		{
			name:        "staging + key present -> works with configured key",
			explicitKey: "",
			envVarKey:   testCustomKey,
			appEnv:      "staging",
			expectedKey: testCustomKey,
		},
		// 6. Staging + key absent -> fails closed (empty string, NO default)
		{
			name:        "staging + key absent -> fails safely; no fallback",
			explicitKey: "",
			envVarKey:   "",
			appEnv:      "staging",
			expectedKey: "",
		},
		// 7. Production + key present -> custom key
		{
			name:        "production + key present -> works with configured key",
			explicitKey: "",
			envVarKey:   testCustomKey,
			appEnv:      "production",
			expectedKey: testCustomKey,
		},
		// 8. Production + key absent -> fails closed (empty string, NO default)
		{
			name:        "production + key absent -> fails safely; no fallback",
			explicitKey: "",
			envVarKey:   "",
			appEnv:      "production",
			expectedKey: "",
		},
		// 9. Unknown environment + key absent -> no development fallback
		{
			name:        "unknown environment (sandbox) + key absent -> no development fallback",
			explicitKey: "",
			envVarKey:   "",
			appEnv:      "sandbox",
			expectedKey: "",
		},
		// 10. Unknown environment (preview) + key absent -> no development fallback
		{
			name:        "unknown environment (preview) + key absent -> no development fallback",
			explicitKey: "",
			envVarKey:   "",
			appEnv:      "preview",
			expectedKey: "",
		},
		// 11. Missing environment + key absent -> no development fallback
		{
			name:        "missing environment + key absent -> no development fallback",
			explicitKey: "",
			envVarKey:   "",
			appEnv:      "",
			expectedKey: "",
		},
		// 12. Explicit option key always respected across any environment
		{
			name:        "production + explicit option key -> uses explicit key",
			explicitKey: "explicit-key-via-option",
			envVarKey:   "",
			appEnv:      "production",
			expectedKey: "explicit-key-via-option",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars
			os.Unsetenv("INTERNAL_SERVICE_TOKEN")
			os.Unsetenv("INTERNAL_SERVICE_KEY")
			os.Unsetenv("AI_SIDECAR_SERVICE_KEY")
			os.Unsetenv("ENV")
			os.Unsetenv("ENVIRONMENT")

			if tt.envVarKey != "" {
				os.Setenv("INTERNAL_SERVICE_TOKEN", tt.envVarKey)
			}
			if tt.appEnv != "" {
				os.Setenv("APP_ENV", tt.appEnv)
			} else {
				os.Unsetenv("APP_ENV")
			}

			resolved := notifications.ResolveServiceKey(tt.explicitKey, tt.appEnv)
			if resolved != tt.expectedKey {
				t.Fatalf("ResolveServiceKey mismatch: got %q, expected %q", resolved, tt.expectedKey)
			}

			// Critical assertion: verify DefaultServiceKey is NEVER returned outside dev/test
			if !tt.allowDefault && resolved == notifications.DefaultServiceKey {
				t.Fatalf("SECURITY VIOLATION: DefaultServiceKey was returned in non-dev/non-test environment (%s)", tt.appEnv)
			}
		})
	}
}

func TestNotifications_SidecarCallsFailClosedWithoutServiceKey(t *testing.T) {
	// Setup isolated service in production environment with no service key
	svc := notifications.NewService(
		nil,
		nil,
		nil,
		nil,
		notifications.WithEnvironment("production"),
		notifications.WithServiceKey(""), // explicitly unconfigured
	)

	ctx := context.Background()

	// 1. Test AnalyzeWithAI fails safely without panic or network call
	_, errAnalyze := svc.AnalyzeWithAI(ctx, 1, 100)
	if errAnalyze == nil {
		t.Fatal("Expected error for AnalyzeWithAI without configured service key in production, got nil")
	}
	if !strings.Contains(errAnalyze.Error(), "internal service key is not configured") {
		t.Errorf("Unexpected error message for AnalyzeWithAI: %v", errAnalyze)
	}

	// 2. Test GenerateDraftWithAI fails safely without panic or network call
	_, errDraft := svc.GenerateDraftWithAI(ctx, 1, 100, "EMAIL")
	if errDraft == nil {
		t.Fatal("Expected error for GenerateDraftWithAI without configured service key in production, got nil")
	}
	if !strings.Contains(errDraft.Error(), "internal service key is not configured") {
		t.Errorf("Unexpected error message for GenerateDraftWithAI: %v", errDraft)
	}

	// 3. Verify error messages do not leak secrets or internal tokens
	forbiddenStrings := []string{
		notifications.DefaultServiceKey,
		"lhq_sec",
		"dev-local-only",
	}
	for _, f := range forbiddenStrings {
		if strings.Contains(errAnalyze.Error(), f) {
			t.Errorf("Security violation: Analyze error contains secret: %s", errAnalyze.Error())
		}
		if strings.Contains(errDraft.Error(), f) {
			t.Errorf("Security violation: Draft error contains secret: %s", errDraft.Error())
		}
	}
}
