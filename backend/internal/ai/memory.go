package ai

// Memory handles storing and retrieving past conversation context for the AI.
type Memory interface {
	// SaveContext saves what was just said so the AI remembers it later.
	// Simple meaning: It writes down the current conversation turn into the AI's notebook.
	// Example: err := memory.SaveContext("session-123", "User asked about pricing.")
	SaveContext(sessionID string, contextData string) error

	// LoadContext gets the past conversation history.
	// Simple meaning: It reads the AI's notebook so it knows what was talked about before.
	// Example: history, err := memory.LoadContext("session-123")
	LoadContext(sessionID string) (string, error)
}
