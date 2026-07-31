package ai

import "context"

// Provider defines the interface for an AI model provider (like OpenAI or Claude).
type Provider interface {
	// GenerateCompletion sends a prompt to the AI and gets a response back.
	// Simple meaning: You ask the AI a question, and this function gets the AI's answer.
	// Example: answer, err := provider.GenerateCompletion(ctx, "Write an email to a client.")
	GenerateCompletion(ctx context.Context, prompt string) (string, error)
}
