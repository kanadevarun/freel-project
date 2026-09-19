package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrDuplicateRequest   = errors.New("duplicate request detected")
	ErrIdempotencyConflict = errors.New("idempotency key conflict: key already used with different action")
)

// IdempotencyStore represents a persistent store for recording action execution state.
type IdempotencyStore interface {
	// Start checks if an idempotency key exists for the organization.
	// If it exists and is completed, it returns the stored result.
	// If it doesn't exist, it marks it as in-progress.
	Start(ctx context.Context, orgID int64, key string) (completed bool, storedResult *ActionResult, err error)

	// Complete stores the final result of an action against the key.
	Complete(ctx context.Context, orgID int64, key string, result *ActionResult, ttl time.Duration) error

	// CheckConflict verifies if an idempotency key has already been used for another action name.
	CheckConflict(ctx context.Context, orgID int64, key string, actionName string) (bool, error)
}

// DBIdempotencyStore persists idempotency records in MariaDB action_idempotency_keys.
type DBIdempotencyStore struct {
	db *sqlx.DB
}

// NewDBIdempotencyStore creates a MariaDB-backed idempotency store.
func NewDBIdempotencyStore(db *sqlx.DB) *DBIdempotencyStore {
	store := &DBIdempotencyStore{db: db}
	store.ensureTable()
	return store
}

func (s *DBIdempotencyStore) ensureTable() {
	query := `
	CREATE TABLE IF NOT EXISTS action_idempotency_keys (
		org_id BIGINT NOT NULL,
		idempotency_key VARCHAR(191) NOT NULL,
		action_name VARCHAR(100) NOT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'IN_PROGRESS',
		result_json LONGTEXT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NULL,
		PRIMARY KEY (org_id, idempotency_key),
		INDEX idx_action_idempotency_expiry (expires_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	_, _ = s.db.Exec(query)
}

func (s *DBIdempotencyStore) Start(ctx context.Context, orgID int64, key string) (bool, *ActionResult, error) {
	var row struct {
		ActionName string         `db:"action_name"`
		Status     string         `db:"status"`
		ResultJSON sql.NullString `db:"result_json"`
		ExpiresAt  sql.NullTime   `db:"expires_at"`
	}

	err := s.db.GetContext(ctx, &row, `
		SELECT action_name, status, result_json, expires_at 
		FROM action_idempotency_keys 
		WHERE org_id = ? AND idempotency_key = ?
	`, orgID, key)

	if err == nil {
		if row.ExpiresAt.Valid && time.Now().After(row.ExpiresAt.Time) {
			// Key expired, reset for fresh execution
			_, _ = s.db.ExecContext(ctx, `
				UPDATE action_idempotency_keys 
				SET status = 'IN_PROGRESS', result_json = NULL, created_at = NOW(), expires_at = NULL 
				WHERE org_id = ? AND idempotency_key = ?
			`, orgID, key)
			return false, nil, nil
		}

		if row.Status == "COMPLETED" && row.ResultJSON.Valid && row.ResultJSON.String != "" {
			var res ActionResult
			if err := json.Unmarshal([]byte(row.ResultJSON.String), &res); err == nil {
				return true, &res, nil
			}
		}
		return false, nil, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO action_idempotency_keys (org_id, idempotency_key, action_name, status, created_at)
			VALUES (?, ?, '', 'IN_PROGRESS', NOW())
			ON DUPLICATE KEY UPDATE status = 'IN_PROGRESS'
		`, orgID, key)
		if err != nil {
			return false, nil, err
		}
	}

	return false, nil, nil
}

func (s *DBIdempotencyStore) Complete(ctx context.Context, orgID int64, key string, result *ActionResult, ttl time.Duration) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(ttl)
	_, err = s.db.ExecContext(ctx, `
		UPDATE action_idempotency_keys 
		SET status = 'COMPLETED', result_json = ?, action_name = ?, expires_at = ?
		WHERE org_id = ? AND idempotency_key = ?
	`, string(b), result.Action, expiresAt, orgID, key)
	return err
}

func (s *DBIdempotencyStore) CheckConflict(ctx context.Context, orgID int64, key string, actionName string) (bool, error) {
	var row struct {
		ActionName string `db:"action_name"`
	}
	err := s.db.GetContext(ctx, &row, `
		SELECT action_name FROM action_idempotency_keys 
		WHERE org_id = ? AND idempotency_key = ?
	`, orgID, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if row.ActionName != "" && row.ActionName != actionName {
		return true, nil
	}
	return false, nil
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

func (w *IdempotentActionWrapper) Name() string                               { return w.action.Name() }
func (w *IdempotentActionWrapper) Module() string                             { return w.action.Module() }
func (w *IdempotentActionWrapper) Description() string                        { return w.action.Description() }
func (w *IdempotentActionWrapper) Category() ActionCategory                   { return w.action.Category() }
func (w *IdempotentActionWrapper) InputSchema() interface{}                   { return w.action.InputSchema() }
func (w *IdempotentActionWrapper) RequiresConfirmation() bool                 { return w.action.RequiresConfirmation() }
func (w *IdempotentActionWrapper) RequiredPermission() (string, string)       { return w.action.RequiredPermission() }

func (w *IdempotentActionWrapper) Execute(ctx *ActionContext, input []byte) (*ActionResult, error) {
	if ctx.IdempotencyKey == nil || *ctx.IdempotencyKey == "" {
		return w.action.Execute(ctx, input)
	}

	conflict, err := w.store.CheckConflict(ctx.Context, ctx.OrganizationID, *ctx.IdempotencyKey, w.action.Name())
	if err != nil {
		return nil, err
	}
	if conflict {
		return nil, ErrIdempotencyConflict
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
		return result, err
	}

	_ = w.store.Complete(ctx.Context, ctx.OrganizationID, *ctx.IdempotencyKey, result, w.ttl)

	return result, nil
}
