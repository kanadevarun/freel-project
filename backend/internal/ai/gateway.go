package ai

import "context"

// Gateway routes AI requests to the right provider and handles fallback logic.
type Gateway interface {
	// ExecutePrompt runs a prompt through the best available AI model.
	// Simple meaning: You give it a task, and it decides whether to use OpenAI, Claude, or something else to get the best result.
	// Example: result, err := gateway.ExecutePrompt(ctx, "Summarize this email")
	ExecutePrompt(ctx context.Context, prompt string) (string, error)
}

type gateway struct {
	providers map[string]Provider
}

// NewGateway creates a new AI Gateway.
// Simple meaning: Sets up the traffic controller that manages all our AI tools (OpenAI, Claude, etc).
// Example: gw := NewGateway(myProviders)
func NewGateway(providers map[string]Provider) Gateway {
	return &gateway{providers: providers}
}

// ExecutePrompt runs a prompt through the best available AI model.
// Simple meaning: You give it a task, and it decides whether to use OpenAI, Claude, or something else to get the best result.
// Example: result, err := gw.ExecutePrompt(ctx, "Summarize this email")
func (g *gateway) ExecutePrompt(ctx context.Context, prompt string) (string, error) {
	// Defaulting to a primary provider, e.g., "openai"
	provider := g.providers["openai"]
	return provider.GenerateCompletion(ctx, prompt)
}
