package jobs

import "context"

// Queue represents a system to enqueue background tasks.
type Queue interface {
	// Enqueue adds a new task to the list of things to do later.
	// Simple meaning: You tell the system to "do this heavy task later in the background so the user doesn't have to wait".
	// Example: err := queue.Enqueue(ctx, "enrich_lead", map[string]interface{}{"lead_id": 123})
	Enqueue(ctx context.Context, taskType string, payload interface{}) error
}
