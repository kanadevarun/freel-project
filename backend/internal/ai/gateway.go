package ai

import (
	"context"
	"fmt"
)


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
	var errors []error

	// ── STEP 1: TRY GOOGLE GEMINI (PRIMARY) ───────────────────────────────────
	// We check if the Gemini provider is registered (meaning a GEMINI_API_KEY was
	// provided in the .env file). If it exists, we run the prompt through Gemini.
	// If it succeeds, we return the answer immediately.
	if geminiProv, exists := g.providers["gemini"]; exists && geminiProv != nil {
		res, err := geminiProv.GenerateCompletion(ctx, prompt)
		if err == nil {
			return res, nil
		}
		// If Gemini fails, save the error and try the next backup model
		errors = append(errors, fmt.Errorf("gemini provider failed: %w", err))
	}

	// ── STEP 2: TRY OPENAI CHATGPT (FAILOVER / BACKUP) ───────────────────────
	// If Gemini was down or returned an error, we fall back to ChatGPT (OpenAI)
	// if an OPENAI_API_KEY is configured in the .env file.
	// This ensures our AI features keep working even if one service goes down.
	if openaiProv, exists := g.providers["openai"]; exists && openaiProv != nil {
		res, err := openaiProv.GenerateCompletion(ctx, prompt)
		if err == nil {
			return res, nil
		}
		// Save the error in case the backup model also fails
		errors = append(errors, fmt.Errorf("openai provider failed: %w", err))
	}

	// ── STEP 3: TRY LOCAL MOCK PROVIDER (FALLBACK FOR DEVELOPMENT) ────────────
	// If both live APIs failed (or if you did not provide any keys in your .env file),
	// we fall back to a local mock provider. This returns predefined mock data
	// immediately, letting you test the website workflow locally without API charges.
	if mockProv, exists := g.providers["mock"]; exists && mockProv != nil {
		return mockProv.GenerateCompletion(ctx, prompt)
	}

	// If everything failed and even the mock provider is missing, return a final error
	return "", fmt.Errorf("no AI providers succeeded. errors: %v", errors)
}
