package actions

// ActionExecutionRequest standardizes the bridge invocation payload sent by Python agents.
type ActionExecutionRequest struct {
	ActionName     string                 `json:"action_name"`
	OrgID          int64                  `json:"org_id"`
	ActingUserID   int64                  `json:"acting_user_id,omitempty"`
	ActorType      ActorType              `json:"actor_type,omitempty"`
	Source         string                 `json:"source,omitempty"`
	TaskID         string                 `json:"task_id,omitempty"`
	ThreadID       string                 `json:"thread_id,omitempty"`
	IdempotencyKey string                 `json:"idempotency_key,omitempty"`
	IsConfirmed    bool                   `json:"is_confirmed,omitempty"`
	Input          map[string]interface{} `json:"input"`
}

// ActionExecutionResponse standardizes the result returned to Python agents.
type ActionExecutionResponse struct {
	Success              bool         `json:"success"`
	ActionName           string       `json:"action_name"`
	CorrelationID        string       `json:"correlation_id"`
	ConfirmationRequired bool         `json:"confirmation_required,omitempty"`
	ApprovalReference    string       `json:"approval_reference,omitempty"`
	IdempotentReplay     bool         `json:"idempotent_replay,omitempty"`
	Data                 interface{}  `json:"data,omitempty"`
	Error                *ActionError `json:"error,omitempty"`
}

// ActionDescriptor provides metadata about an action for introspection and dynamic discovery.
type ActionDescriptor struct {
	Name                 string         `json:"name"`
	Module               string         `json:"module"`
	Description          string         `json:"description"`
	Category             ActionCategory `json:"category"`
	RequiredPermission   string         `json:"required_permission"`
	RequiresConfirmation bool           `json:"requires_confirmation"`
	InputSchema          interface{}    `json:"input_schema,omitempty"`
}
