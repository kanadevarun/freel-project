package actions

import (
	"context"
)

// ActorType represents the type of client/actor invoking the action.
type ActorType string

const (
	ActorTypeUI      ActorType = "UI"
	ActorTypeAPI     ActorType = "API"
	ActorTypeSystem  ActorType = "SYSTEM"
	ActorTypeAIAgent ActorType = "AI_AGENT"
)

// ActionContext encapsulates the execution environment for a Business Action.
// It carries authorization information and the actor type to ensure 
// operations are safely bounded (e.g. AI tools cannot elevate privileges).
type ActionContext struct {
	context.Context
	OrganizationID int64
	ActingUserID   int64
	ActorType      ActorType
	RequestID      string
	IdempotencyKey *string
}

// NewActionContext creates a new ActionContext.
func NewActionContext(ctx context.Context, orgID int64, userID int64, actorType ActorType, reqID string) *ActionContext {
	return &ActionContext{
		Context:        ctx,
		OrganizationID: orgID,
		ActingUserID:   userID,
		ActorType:      actorType,
		RequestID:      reqID,
	}
}

// WithIdempotencyKey attaches an idempotency key to the context for safe retries.
func (ac *ActionContext) WithIdempotencyKey(key string) *ActionContext {
	ac.IdempotencyKey = &key
	return ac
}
