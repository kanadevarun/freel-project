package workflow

// RuleConfig defines a routing or assignment rule.
type RuleConfig struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // e.g., "EQUALS", "CONTAINS"
	Value    string `json:"value"`
	AssignTo string `json:"assign_to"` // e.g., "ROLE:PRICING", "USER:123"
}

// WorkflowConfig defines a set of rules for a specific entity type (like RFQ or Lead).
type WorkflowConfig struct {
	EntityType string       `json:"entity_type"`
	Rules      []RuleConfig `json:"rules"`
}

// LoadConfig loads the workflow configuration for an entity.
// Simple meaning: It grabs the rulebook for a specific type of item, like an RFQ or a Lead, so we know how to route it.
// Example: config, err := LoadConfig("RFQ")
func LoadConfig(entityType string) (*WorkflowConfig, error) {
	// Placeholder: load from DB or JSON file in the future
	return &WorkflowConfig{
		EntityType: entityType,
		Rules:      []RuleConfig{},
	}, nil
}
