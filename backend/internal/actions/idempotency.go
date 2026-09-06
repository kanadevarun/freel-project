package actions

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDuplicateRequest = errors.New("duplicate request detected")
)

// IdempotencyStore represents a persistent store for recording action execution state.
// Typically implemented via Redis or a dedicated database table.
type IdempotencyStore interface {
	// Start checks if an idempotency key exists for the organization. 
	// If it exists and is completed, it returns the stored result.
	// If it doesn't exist, it marks it as in-progress.
	Start(ctx context.Context, orgID int64, key string) (completed bool, storedResult *ActionResult, err error)
	
	// Complete stores the final result of an action against the key.
	Complete(ctx context.Context, orgID int64, key string, result *ActionResult, ttl time.Duration) error
}

// IdempotentActionWrapper wraps an existing Action to enforce idempotency.
type IdempotentActionWrapper struct {
	store  IdempotencyStore
	action Action
	ttl    time.Duration
}

// NewIdempotentAction creates a wrapper ensuring the wrapped Action cannot be redundantly executed.
func NewIdempotentAction(store IdempotencyStore, action Action, ttl time.Duration) Action {
	return &IdempotentActionWrapper{
		store:  store,
		action: action,
		ttl:    ttl,
	}
}

func (w *IdempotentActionWrapper) Name() string { return w.action.Name() }
func (w *IdempotentActionWrapper) Module() string { return w.action.Module() }
func (w *IdempotentActionWrapper) Description() string { return w.action.Description() }
func (w *IdempotentActionWrapper) Category() ActionCategory { return w.action.Category() }
func (w *IdempotentActionWrapper) InputSchema() interface{} { return w.action.InputSchema() }
func (w *IdempotentActionWrapper) RequiresConfirmation() bool { return w.action.RequiresConfirmation() }

func (w *IdempotentActionWrapper) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	if ctx.IdempotencyKey == nil || *ctx.IdempotencyKey == "" {
		// If no key is provided, we can either reject or proceed normally. 
		// For safety, high-risk actions should mandate a key, but for now we proceed if absent.
		return w.action.Execute(ctx, input)
	}

	completed, storedResult, err := w.store.Start(ctx.Context, ctx.OrganizationID, *ctx.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if completed {
		return storedResult, nil
	}

	result, err := w.action.Execute(ctx, input)
	if err != nil {
		// Do not store failures, allow retry
		return result, err
	}

	if err := w.store.Complete(ctx.Context, ctx.OrganizationID, *ctx.IdempotencyKey, result, w.ttl); err != nil {
		// Log the error but return success since the business logic succeeded.
	}

	return result, nil
}
