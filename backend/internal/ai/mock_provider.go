package ai

import (
	"context"
	"time"
)

// mockProvider is a fake AI that returns hardcoded responses.
// Simple meaning: Since we don't want to spend real money on OpenAI tokens
// during testing, we use this fake AI. It waits a second to pretend it's thinking
// and then always gives a predefined answer.
type mockProvider struct{}

func NewMockProvider() Provider {
	return &mockProvider{}
}

func (m *mockProvider) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	// Pretend to think for 1 second...
	time.Sleep(1 * time.Second)

	// We are faking the response for the Lead Scoring prompt we just wrote.
	// We return exactly what the prompt asked for: a JSON object with score and research_report.
	fakeJSON := `{
		"score": 85,
		"research_report": "This company has high revenue and ships a significant volume of containers. They are a strong prospect for our logistics services."
	}`

	return fakeJSON, nil
}
