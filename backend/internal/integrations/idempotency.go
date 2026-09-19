package integrations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// IdempotencyManager ensures external side effects (SMS, Email, carrier dispatches) execute at most once.
type IdempotencyManager interface {
	GenerateKey(orgID int64, actionType string, target string, payload []byte) string
	Acquire(ctx context.Context, orgID int64, key string, actionName string, ttl time.Duration) (acquired bool, existingResult []byte, err error)
	Release(ctx context.Context, orgID int64, key string, result []byte) error
}

type sqlIdempotencyManager struct {
	db *sqlx.DB
}

// NewIdempotencyManager creates a MariaDB-backed idempotency manager reusing the canonical action_idempotency_keys table.
func NewIdempotencyManager(db *sqlx.DB) IdempotencyManager {
	return &sqlIdempotencyManager{db: db}
}

func (m *sqlIdempotencyManager) GenerateKey(orgID int64, actionType string, target string, payload []byte) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d:%s:%s:", orgID, actionType, target)))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

func (m *sqlIdempotencyManager) Acquire(ctx context.Context, orgID int64, key string, actionName string, ttl time.Duration) (bool, []byte, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(ttl)

	var row struct {
		ActionName string         `db:"action_name"`
		Status     string         `db:"status"`
		ResultJSON sql.NullString `db:"result_json"`
		ExpiresAt  sql.NullTime   `db:"expires_at"`
	}

	err := m.db.GetContext(ctx, &row, `
		SELECT action_name, status, result_json, expires_at
		FROM action_idempotency_keys
		WHERE org_id = ? AND idempotency_key = ?
	`, orgID, key)

	if err == nil {
		// Key already exists
		if row.ExpiresAt.Valid && now.After(row.ExpiresAt.Time) {
			// Expired: reset and reacquire
			_, err = m.db.ExecContext(ctx, `
				UPDATE action_idempotency_keys
				SET status = 'IN_PROGRESS', result_json = NULL, action_name = ?, created_at = NOW(), expires_at = ?
				WHERE org_id = ? AND idempotency_key = ?
			`, actionName, expiresAt, orgID, key)
			if err != nil {
				return false, nil, err
			}
			return true, nil, nil
		}

		if row.Status == "COMPLETED" && row.ResultJSON.Valid {
			// Already completed: return stored response
			return false, []byte(row.ResultJSON.String), nil
		}

		// In progress or not completed: conflict / duplicate
		return false, nil, NewDuplicateRequestError(key)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return false, nil, err
	}

	// Insert fresh in-progress lock
	query := `
		INSERT INTO action_idempotency_keys (org_id, idempotency_key, action_name, status, created_at, expires_at)
		VALUES (?, ?, ?, 'IN_PROGRESS', NOW(), ?)
	`
	_, err = m.db.ExecContext(ctx, query, orgID, key, actionName, expiresAt)
	if err != nil {
		return false, nil, err
	}

	return true, nil, nil
}

func (m *sqlIdempotencyManager) Release(ctx context.Context, orgID int64, key string, result []byte) error {
	query := `
		UPDATE action_idempotency_keys
		SET status = 'COMPLETED', result_json = ?
		WHERE org_id = ? AND idempotency_key = ?
	`
	_, err := m.db.ExecContext(ctx, query, string(result), orgID, key)
	return err
}
