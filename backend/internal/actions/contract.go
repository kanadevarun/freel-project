package actions

// ActionCategory represents the overarching type of business operation.
type ActionCategory string

const (
	ActionCategoryRead      ActionCategory = "READ"
	ActionCategoryWrite     ActionCategory = "WRITE"
	ActionCategoryHighRisk  ActionCategory = "HIGH_RISK"
)

// ActionResult standardizes output across all AI and UI business actions.
type ActionResult struct {
	Success      bool        `json:"success"`
	Action       string      `json:"action"`
	ResourceType string      `json:"resource_type,omitempty"`
	ResourceID   string      `json:"resource_id,omitempty"`
	Summary      string      `json:"summary,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	Error        *ActionError `json:"error,omitempty"`
}

// ActionError standardizes business errors to prevent leaking stack traces.
type ActionError struct {
	Type    string `json:"type"` // e.g. "Validation", "Unauthorized", "Conflict", "BusinessRule"
	Message string `json:"message"`
}

// Action defines the contract for an executable business operation in LogisticsHQ.
// AI agents or system integrations must adhere to this interface.
type Action interface {
	// Name returns a unique, stable identifier (e.g., "shipments.update").
	Name() string
	
	// Module returns the business domain (e.g., "shipments", "billing").
	Module() string
	
	// Description provides a clear intent, useful for the AI Tool Registry.
	Description() string
	
	// Category classifies the risk and access type.
	Category() ActionCategory
	
	// InputSchema returns a pointer to the expected input struct (for JSON validation/discovery).
	InputSchema() interface{}
	
	// RequiresConfirmation indicates if a human must explicitly approve execution.
	RequiresConfirmation() bool
	
	// Execute performs the operation. It must validate inputs and authorize based on ctx.
	Execute(ctx *ActionContext, input []byte) (*ActionResult, error)
}
