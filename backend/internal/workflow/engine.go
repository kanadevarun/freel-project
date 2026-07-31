package workflow

import "context"

// Engine is the central brain that runs items through the configured workflows.
type Engine interface {
	// ProcessEvent evaluates a system event and triggers the right workflow.
	// Simple meaning: When something happens (like a Lead is created), this function checks the rulebook and passes it to the right person.
	// Example: err := engine.ProcessEvent(ctx, "LEAD_CREATED", leadDataMap)
	ProcessEvent(ctx context.Context, eventType string, entityData map[string]interface{}) error
}

type engine struct {
	assigner Assigner
}

// NewEngine creates a new Workflow Engine.
// Simple meaning: It starts up the main control center for routing tasks across the company.
// Example: myEngine := NewEngine(myAssigner)
func NewEngine(assigner Assigner) Engine {
	return &engine{
		assigner: assigner,
	}
}

// ProcessEvent evaluates a system event and triggers the right workflow.
// Simple meaning: When something happens (like a Lead is created), this function checks the rulebook and passes it to the right person.
// Example: err := engine.ProcessEvent(ctx, "LEAD_CREATED", leadDataMap)
func (e *engine) ProcessEvent(ctx context.Context, eventType string, entityData map[string]interface{}) error {
	// 1. Load config based on event type
	config, err := LoadConfig(eventType)
	if err != nil {
		return err
	}

	// 2. Determine who should get it
	assignee, err := e.assigner.DetermineAssignee(ctx, config, entityData)
	if err != nil {
		return err
	}

	// 3. (Future) Trigger notification or DB update to actually assign the item to `assignee`
	_ = assignee // ignoring for now

	return nil
}
