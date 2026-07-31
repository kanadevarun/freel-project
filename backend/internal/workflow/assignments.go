package workflow

import (
	"context"
	"errors"
)

// Assigner handles determining who should take ownership of a task.
type Assigner interface {
	// DetermineAssignee looks at the rules and the data to figure out who gets the job.
	// Simple meaning: It reads the item (like a new Quote) and decides which team or person needs to handle it next.
	// Example: assignee, err := assigner.DetermineAssignee(ctx, myRfqConfig, rfqData)
	DetermineAssignee(ctx context.Context, config *WorkflowConfig, entityData map[string]interface{}) (string, error)
}

type assigner struct{}

// NewAssigner creates a new Assigner tool.
// Simple meaning: It builds the machine that figures out who gets assigned to what.
// Example: myAssigner := NewAssigner()
func NewAssigner() Assigner {
	return &assigner{}
}

// DetermineAssignee looks at the rules and the data to figure out who gets the job.
// Simple meaning: It reads the item (like a new Quote) and decides which team or person needs to handle it next.
// Example: assignee, err := assigner.DetermineAssignee(ctx, myRfqConfig, rfqData)
func (a *assigner) DetermineAssignee(ctx context.Context, config *WorkflowConfig, entityData map[string]interface{}) (string, error) {
	if config == nil || len(config.Rules) == 0 {
		return "", errors.New("no rules configured")
	}

	// Simple mock evaluation loop
	for _, rule := range config.Rules {
		if val, ok := entityData[rule.Field]; ok {
			// Extremely simple exact match for MVP
			if val == rule.Value { 
				return rule.AssignTo, nil
			}
		}
	}

	// Fallback assignment if no rules match
	return "ROLE:ADMIN", nil
}
