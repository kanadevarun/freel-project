package domain

import (
	"testing"
)

func TestSanitizeMap(t *testing.T) {
	input := map[string]interface{}{
		"username":        "testuser@logisticshq.in",
		"password":        "SuperSecret123!",
		"api_key":         "carrier_live_key_998811",
		"access_token":    "eyJh...token...",
		"refresh_token":   "refresh...secret...",
		"client_secret":   "secret_oauth_abc",
		"carrier_code":    "MAEU",
		"cvv":             "123",
		"credit_card":     "4111222233334444",
		"safe_notes":      "Cargo cleared customs at Port of Rotterdam",
		"nested_payload": map[string]interface{}{
			"bearer_token": "token123",
			"status":       "IN_TRANSIT",
		},
	}

	sanitized := SanitizeMap(input)

	// Check sensitive fields are redacted
	sensitiveKeys := []string{"password", "api_key", "access_token", "refresh_token", "client_secret", "cvv", "credit_card"}
	for _, key := range sensitiveKeys {
		val, ok := sanitized[key]
		if !ok {
			t.Errorf("expected key %s to exist", key)
		}
		if val != "[REDACTED]" {
			t.Errorf("expected key %s to be redacted, got %v", key, val)
		}
	}

	// Check non-sensitive fields are preserved
	if sanitized["username"] != "testuser@logisticshq.in" {
		t.Errorf("expected username preserved, got %v", sanitized["username"])
	}
	if sanitized["carrier_code"] != "MAEU" {
		t.Errorf("expected carrier_code preserved, got %v", sanitized["carrier_code"])
	}
	if sanitized["safe_notes"] != "Cargo cleared customs at Port of Rotterdam" {
		t.Errorf("expected safe_notes preserved, got %v", sanitized["safe_notes"])
	}

	// Check nested map
	nested, ok := sanitized["nested_payload"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected nested_payload to be map")
	}
	if nested["bearer_token"] != "[REDACTED]" {
		t.Errorf("expected nested bearer_token to be redacted, got %v", nested["bearer_token"])
	}
	if nested["status"] != "IN_TRANSIT" {
		t.Errorf("expected nested status preserved, got %v", nested["status"])
	}
}

func TestComputeChanges(t *testing.T) {
	before := map[string]interface{}{
		"status":   "In Transit",
		"eta":      "2026-09-10",
		"vessel":   "MSC Anna",
		"password": "old_password_123",
	}

	after := map[string]interface{}{
		"status":   "Delayed",
		"eta":      "2026-09-14",
		"vessel":   "MSC Anna", // Unchanged
		"password": "new_password_456",
		"new_note": "Port congestion",
	}

	changes := ComputeChanges(before, after)

	if len(changes) != 4 {
		t.Fatalf("expected 4 changes (status, eta, password, new_note), got %d: %+v", len(changes), changes)
	}

	changesMap := make(map[string]FieldChange)
	for _, c := range changes {
		changesMap[c.Field] = c
	}

	// status changed
	if ch, ok := changesMap["status"]; !ok || ch.Before != "In Transit" || ch.After != "Delayed" {
		t.Errorf("unexpected status change: %+v", ch)
	}

	// eta changed
	if ch, ok := changesMap["eta"]; !ok || ch.Before != "2026-09-10" || ch.After != "2026-09-14" {
		t.Errorf("unexpected eta change: %+v", ch)
	}

	// password changed (must be redacted)
	if ch, ok := changesMap["password"]; !ok || ch.Before != "[REDACTED]" || ch.After != "[REDACTED]" {
		t.Errorf("expected password change to be redacted, got: %+v", ch)
	}

	// new_note added
	if ch, ok := changesMap["new_note"]; !ok || ch.Before != nil || ch.After != "Port congestion" {
		t.Errorf("unexpected new_note change: %+v", ch)
	}
}
