package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Repository defines database operations for AI memory, preferences, and personalization
type Repository interface {
	// Memory Items
	CreateMemoryItem(ctx context.Context, item *MemoryItem) (*MemoryItem, error)
	GetMemoryItemByID(ctx context.Context, orgID int64, id int64) (*MemoryItem, error)
	ListMemoryItems(ctx context.Context, orgID int64, userID int64, filter MemoryFilter) ([]*MemoryItem, int64, error)
	UpdateMemoryItem(ctx context.Context, item *MemoryItem) (*MemoryItem, error)
	DeleteMemoryItem(ctx context.Context, orgID int64, id int64) error
	ClearUserMemories(ctx context.Context, orgID int64, userID int64) (int64, error)
	SetMemoryStatus(ctx context.Context, orgID int64, id int64, status string, updatedBy string) error
	GetActiveMemoriesForRuntime(ctx context.Context, orgID int64, userID int64) ([]*MemoryItem, error)

	// Preferences
	SetPreference(ctx context.Context, pref *Preference) (*Preference, error)
	GetPreference(ctx context.Context, orgID int64, userID int64, scope string, key string) (*Preference, error)
	ListPreferences(ctx context.Context, orgID int64, userID int64, scope string) ([]*Preference, error)
	DeletePreference(ctx context.Context, orgID int64, userID int64, scope string, key string) error

	// User Personalization Settings
	GetUserSettings(ctx context.Context, orgID int64, userID int64) (*UserPersonalizationSettings, error)
	UpsertUserSettings(ctx context.Context, settings *UserPersonalizationSettings) (*UserPersonalizationSettings, error)

	// Audit Events
	RecordAuditEvent(ctx context.Context, event *MemoryAuditEvent) error
	ListAuditEvents(ctx context.Context, orgID int64, userID int64, limit int, offset int) ([]*MemoryAuditEvent, error)

	// Stats
	GetMemoryStats(ctx context.Context, orgID int64, userID int64) (*MemoryStats, error)
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new memory repository
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

// CreateMemoryItem persists a new memory item
func (r *repository) CreateMemoryItem(ctx context.Context, item *MemoryItem) (*MemoryItem, error) {
	query := `
		INSERT INTO ai_memory_items (
			org_id, user_id, scope, memory_type, title, content, structured_value,
			source_type, source_reference, evidence, confidence, explicitly_confirmed,
			status, review_at, expires_at, created_by, updated_by, correlation_id
		) VALUES (
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?
		)
	`
	var structJSON *string
	if len(item.StructuredValue) > 0 {
		str := string(item.StructuredValue)
		structJSON = &str
	}

	res, err := r.db.ExecContext(ctx, query,
		item.OrgID, item.UserID, item.Scope, item.MemoryType, item.Title, item.Content, structJSON,
		item.SourceType, item.SourceReference, item.Evidence, item.Confidence, item.ExplicitlyConfirmed,
		item.Status, item.ReviewAt, item.ExpiresAt, item.CreatedBy, item.UpdatedBy, item.CorrelationID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert memory item: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetMemoryItemByID(ctx, item.OrgID, id)
}

// GetMemoryItemByID fetches a memory item by ID ensuring tenant isolation
func (r *repository) GetMemoryItemByID(ctx context.Context, orgID int64, id int64) (*MemoryItem, error) {
	query := `
		SELECT id, org_id, user_id, scope, memory_type, title, content, structured_value,
		       source_type, source_reference, evidence, confidence, explicitly_confirmed,
		       status, review_at, expires_at, last_used_at, created_by, updated_by,
		       correlation_id, created_at, updated_at
		FROM ai_memory_items
		WHERE org_id = ? AND id = ?
	`
	var item MemoryItem
	var structVal sql.NullString
	err := r.db.QueryRowContext(ctx, query, orgID, id).Scan(
		&item.ID, &item.OrgID, &item.UserID, &item.Scope, &item.MemoryType, &item.Title, &item.Content, &structVal,
		&item.SourceType, &item.SourceReference, &item.Evidence, &item.Confidence, &item.ExplicitlyConfirmed,
		&item.Status, &item.ReviewAt, &item.ExpiresAt, &item.LastUsedAt, &item.CreatedBy, &item.UpdatedBy,
		&item.CorrelationID, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get memory item: %w", err)
	}

	if structVal.Valid && structVal.String != "" {
		item.StructuredValue = json.RawMessage(structVal.String)
	}

	return &item, nil
}

// ListMemoryItems returns paginated memory items according to scope and user ownership
func (r *repository) ListMemoryItems(ctx context.Context, orgID int64, userID int64, filter MemoryFilter) ([]*MemoryItem, int64, error) {
	var whereClauses []string
	var args []interface{}

	whereClauses = append(whereClauses, "org_id = ?")
	args = append(args, orgID)

	// Scope and ownership filtering
	if filter.Scope == ScopeUser {
		whereClauses = append(whereClauses, "scope = 'USER' AND user_id = ?")
		args = append(args, userID)
	} else if filter.Scope == ScopeOrganization {
		whereClauses = append(whereClauses, "scope = 'ORGANIZATION'")
	} else {
		// Both scopes available to this user: their personal memories + org memories
		whereClauses = append(whereClauses, "((scope = 'USER' AND user_id = ?) OR scope = 'ORGANIZATION')")
		args = append(args, userID)
	}

	if filter.Status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, filter.Status)
	} else {
		// By default exclude soft-deleted
		whereClauses = append(whereClauses, "status != 'DELETED'")
	}

	if filter.MemoryType != "" {
		whereClauses = append(whereClauses, "memory_type = ?")
		args = append(args, filter.MemoryType)
	}

	if filter.Search != "" {
		whereClauses = append(whereClauses, "(title LIKE ? OR content LIKE ?)")
		pattern := "%" + filter.Search + "%"
		args = append(args, pattern, pattern)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Count query
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ai_memory_items WHERE %s", whereSQL)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count memory items: %w", err)
	}

	// Fetch query
	limit := 50
	if filter.Limit > 0 && filter.Limit <= 100 {
		limit = filter.Limit
	}
	offset := 0
	if filter.Offset > 0 {
		offset = filter.Offset
	}

	fetchQuery := fmt.Sprintf(`
		SELECT id, org_id, user_id, scope, memory_type, title, content, structured_value,
		       source_type, source_reference, evidence, confidence, explicitly_confirmed,
		       status, review_at, expires_at, last_used_at, created_by, updated_by,
		       correlation_id, created_at, updated_at
		FROM ai_memory_items
		WHERE %s
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`, whereSQL)

	fetchArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, fetchQuery, fetchArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list memory items: %w", err)
	}
	defer rows.Close()

	var items []*MemoryItem
	for rows.Next() {
		var item MemoryItem
		var structVal sql.NullString
		if err := rows.Scan(
			&item.ID, &item.OrgID, &item.UserID, &item.Scope, &item.MemoryType, &item.Title, &item.Content, &structVal,
			&item.SourceType, &item.SourceReference, &item.Evidence, &item.Confidence, &item.ExplicitlyConfirmed,
			&item.Status, &item.ReviewAt, &item.ExpiresAt, &item.LastUsedAt, &item.CreatedBy, &item.UpdatedBy,
			&item.CorrelationID, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if structVal.Valid && structVal.String != "" {
			item.StructuredValue = json.RawMessage(structVal.String)
		}
		items = append(items, &item)
	}

	return items, total, nil
}

// UpdateMemoryItem updates an existing memory item
func (r *repository) UpdateMemoryItem(ctx context.Context, item *MemoryItem) (*MemoryItem, error) {
	query := `
		UPDATE ai_memory_items
		SET title = ?, content = ?, structured_value = ?, status = ?,
		    expires_at = ?, review_at = ?, updated_by = ?, updated_at = NOW()
		WHERE org_id = ? AND id = ?
	`
	var structJSON *string
	if len(item.StructuredValue) > 0 {
		str := string(item.StructuredValue)
		structJSON = &str
	}

	_, err := r.db.ExecContext(ctx, query,
		item.Title, item.Content, structJSON, item.Status,
		item.ExpiresAt, item.ReviewAt, item.UpdatedBy,
		item.OrgID, item.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update memory item: %w", err)
	}

	return r.GetMemoryItemByID(ctx, item.OrgID, item.ID)
}

// DeleteMemoryItem marks a memory item as soft-deleted or removes it
func (r *repository) DeleteMemoryItem(ctx context.Context, orgID int64, id int64) error {
	query := `UPDATE ai_memory_items SET status = 'DELETED', updated_at = NOW() WHERE org_id = ? AND id = ?`
	_, err := r.db.ExecContext(ctx, query, orgID, id)
	return err
}

// ClearUserMemories soft-deletes all personal memories for a user
func (r *repository) ClearUserMemories(ctx context.Context, orgID int64, userID int64) (int64, error) {
	query := `UPDATE ai_memory_items SET status = 'DELETED', updated_at = NOW() WHERE org_id = ? AND user_id = ? AND scope = 'USER' AND status != 'DELETED'`
	res, err := r.db.ExecContext(ctx, query, orgID, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// SetMemoryStatus updates status (e.g. DISABLED, ACTIVE)
func (r *repository) SetMemoryStatus(ctx context.Context, orgID int64, id int64, status string, updatedBy string) error {
	query := `UPDATE ai_memory_items SET status = ?, updated_by = ?, updated_at = NOW() WHERE org_id = ? AND id = ?`
	_, err := r.db.ExecContext(ctx, query, status, updatedBy, orgID, id)
	return err
}

// GetActiveMemoriesForRuntime retrieves non-expired, active memories for AI injection
func (r *repository) GetActiveMemoriesForRuntime(ctx context.Context, orgID int64, userID int64) ([]*MemoryItem, error) {
	query := `
		SELECT id, org_id, user_id, scope, memory_type, title, content, structured_value,
		       source_type, source_reference, evidence, confidence, explicitly_confirmed,
		       status, review_at, expires_at, last_used_at, created_by, updated_by,
		       correlation_id, created_at, updated_at
		FROM ai_memory_items
		WHERE org_id = ?
		  AND status = 'ACTIVE'
		  AND (expires_at IS NULL OR expires_at > NOW())
		  AND ((scope = 'USER' AND user_id = ?) OR scope = 'ORGANIZATION')
		ORDER BY scope DESC, updated_at DESC
		LIMIT 20
	`
	rows, err := r.db.QueryContext(ctx, query, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query runtime memories: %w", err)
	}
	defer rows.Close()

	var items []*MemoryItem
	for rows.Next() {
		var item MemoryItem
		var structVal sql.NullString
		if err := rows.Scan(
			&item.ID, &item.OrgID, &item.UserID, &item.Scope, &item.MemoryType, &item.Title, &item.Content, &structVal,
			&item.SourceType, &item.SourceReference, &item.Evidence, &item.Confidence, &item.ExplicitlyConfirmed,
			&item.Status, &item.ReviewAt, &item.ExpiresAt, &item.LastUsedAt, &item.CreatedBy, &item.UpdatedBy,
			&item.CorrelationID, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if structVal.Valid && structVal.String != "" {
			item.StructuredValue = json.RawMessage(structVal.String)
		}
		items = append(items, &item)
	}

	return items, nil
}

// SetPreference creates or updates a preference
func (r *repository) SetPreference(ctx context.Context, pref *Preference) (*Preference, error) {
	query := `
		INSERT INTO ai_preferences (
			org_id, user_id, scope, preference_key, preference_value, value_type,
			description, source, explicitly_confirmed, is_disabled, created_by, updated_by, correlation_id
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?
		)
		ON DUPLICATE KEY UPDATE
			preference_value = VALUES(preference_value),
			value_type = VALUES(value_type),
			description = VALUES(description),
			is_disabled = VALUES(is_disabled),
			updated_by = VALUES(updated_by),
			correlation_id = VALUES(correlation_id),
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		pref.OrgID, pref.UserID, pref.Scope, pref.PreferenceKey, pref.PreferenceValue, pref.ValueType,
		pref.Description, pref.Source, pref.ExplicitlyConfirmed, pref.IsDisabled, pref.CreatedBy, pref.UpdatedBy, pref.CorrelationID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set preference: %w", err)
	}

	return r.GetPreference(ctx, pref.OrgID, pref.UserID, pref.Scope, pref.PreferenceKey)
}

// GetPreference fetches a preference by key and scope
func (r *repository) GetPreference(ctx context.Context, orgID int64, userID int64, scope string, key string) (*Preference, error) {
	effectiveUserID := userID
	if scope == ScopeOrganization {
		effectiveUserID = 0
	}

	query := `
		SELECT id, org_id, user_id, scope, preference_key, preference_value, value_type,
		       description, source, explicitly_confirmed, is_disabled, disabled_at, last_used_at,
		       created_by, updated_by, correlation_id, created_at, updated_at
		FROM ai_preferences
		WHERE org_id = ? AND user_id = ? AND preference_key = ?
	`
	var p Preference
	err := r.db.QueryRowContext(ctx, query, orgID, effectiveUserID, key).Scan(
		&p.ID, &p.OrgID, &p.UserID, &p.Scope, &p.PreferenceKey, &p.PreferenceValue, &p.ValueType,
		&p.Description, &p.Source, &p.ExplicitlyConfirmed, &p.IsDisabled, &p.DisabledAt, &p.LastUsedAt,
		&p.CreatedBy, &p.UpdatedBy, &p.CorrelationID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get preference: %w", err)
	}

	return &p, nil
}

// ListPreferences lists preferences for scope
func (r *repository) ListPreferences(ctx context.Context, orgID int64, userID int64, scope string) ([]*Preference, error) {
	var query string
	var args []interface{}

	if scope == ScopeOrganization {
		query = `
			SELECT id, org_id, user_id, scope, preference_key, preference_value, value_type,
			       description, source, explicitly_confirmed, is_disabled, disabled_at, last_used_at,
			       created_by, updated_by, correlation_id, created_at, updated_at
			FROM ai_preferences
			WHERE org_id = ? AND user_id = 0 AND scope = 'ORGANIZATION'
			ORDER BY preference_key ASC
		`
		args = append(args, orgID)
	} else {
		query = `
			SELECT id, org_id, user_id, scope, preference_key, preference_value, value_type,
			       description, source, explicitly_confirmed, is_disabled, disabled_at, last_used_at,
			       created_by, updated_by, correlation_id, created_at, updated_at
			FROM ai_preferences
			WHERE org_id = ? AND user_id = ? AND scope = 'USER'
			ORDER BY preference_key ASC
		`
		args = append(args, orgID, userID)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list preferences: %w", err)
	}
	defer rows.Close()

	var list []*Preference
	for rows.Next() {
		var p Preference
		if err := rows.Scan(
			&p.ID, &p.OrgID, &p.UserID, &p.Scope, &p.PreferenceKey, &p.PreferenceValue, &p.ValueType,
			&p.Description, &p.Source, &p.ExplicitlyConfirmed, &p.IsDisabled, &p.DisabledAt, &p.LastUsedAt,
			&p.CreatedBy, &p.UpdatedBy, &p.CorrelationID, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}

	return list, nil
}

// DeletePreference removes a preference
func (r *repository) DeletePreference(ctx context.Context, orgID int64, userID int64, scope string, key string) error {
	effectiveUserID := userID
	if scope == ScopeOrganization {
		effectiveUserID = 0
	}
	query := `DELETE FROM ai_preferences WHERE org_id = ? AND user_id = ? AND preference_key = ?`
	_, err := r.db.ExecContext(ctx, query, orgID, effectiveUserID, key)
	return err
}

// GetUserSettings gets personalization settings for a user, returning default if not yet stored
func (r *repository) GetUserSettings(ctx context.Context, orgID int64, userID int64) (*UserPersonalizationSettings, error) {
	query := `
		SELECT id, org_id, user_id, personalization_enabled, preferred_response_style,
		       preferred_summary_depth, preferred_currency, preferred_timezone,
		       preferred_date_format, preferred_default_module, explanation_level,
		       created_at, updated_at
		FROM ai_user_personalization_settings
		WHERE org_id = ? AND user_id = ?
	`
	var s UserPersonalizationSettings
	err := r.db.QueryRowContext(ctx, query, orgID, userID).Scan(
		&s.ID, &s.OrgID, &s.UserID, &s.PersonalizationEnabled, &s.PreferredResponseStyle,
		&s.PreferredSummaryDepth, &s.PreferredCurrency, &s.PreferredTimezone,
		&s.PreferredDateFormat, &s.PreferredDefaultModule, &s.ExplanationLevel,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return default configuration
			return &UserPersonalizationSettings{
				OrgID:                  orgID,
				UserID:                 userID,
				PersonalizationEnabled: true,
				PreferredResponseStyle: "CONCISE",
				PreferredSummaryDepth:  "STANDARD",
				PreferredCurrency:      "USD",
				PreferredTimezone:      "UTC",
				PreferredDateFormat:    "YYYY-MM-DD",
				PreferredDefaultModule: "DASHBOARD",
				ExplanationLevel:       "STANDARD",
			}, nil
		}
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	return &s, nil
}

// UpsertUserSettings updates user personalization settings
func (r *repository) UpsertUserSettings(ctx context.Context, settings *UserPersonalizationSettings) (*UserPersonalizationSettings, error) {
	query := `
		INSERT INTO ai_user_personalization_settings (
			org_id, user_id, personalization_enabled, preferred_response_style,
			preferred_summary_depth, preferred_currency, preferred_timezone,
			preferred_date_format, preferred_default_module, explanation_level
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?
		)
		ON DUPLICATE KEY UPDATE
			personalization_enabled = VALUES(personalization_enabled),
			preferred_response_style = VALUES(preferred_response_style),
			preferred_summary_depth = VALUES(preferred_summary_depth),
			preferred_currency = VALUES(preferred_currency),
			preferred_timezone = VALUES(preferred_timezone),
			preferred_date_format = VALUES(preferred_date_format),
			preferred_default_module = VALUES(preferred_default_module),
			explanation_level = VALUES(explanation_level),
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		settings.OrgID, settings.UserID, settings.PersonalizationEnabled, settings.PreferredResponseStyle,
		settings.PreferredSummaryDepth, settings.PreferredCurrency, settings.PreferredTimezone,
		settings.PreferredDateFormat, settings.PreferredDefaultModule, settings.ExplanationLevel,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert user settings: %w", err)
	}

	return r.GetUserSettings(ctx, settings.OrgID, settings.UserID)
}

// RecordAuditEvent logs a memory lifecycle event
func (r *repository) RecordAuditEvent(ctx context.Context, event *MemoryAuditEvent) error {
	query := `
		INSERT INTO ai_memory_audit_events (
			org_id, user_id, memory_item_id, event_type, scope,
			actor_name, actor_id, details, correlation_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		event.OrgID, event.UserID, event.MemoryItemID, event.EventType, event.Scope,
		event.ActorName, event.ActorID, event.Details, event.CorrelationID,
	)
	return err
}

// ListAuditEvents retrieves recent audit events for user or organization
func (r *repository) ListAuditEvents(ctx context.Context, orgID int64, userID int64, limit int, offset int) ([]*MemoryAuditEvent, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, org_id, user_id, memory_item_id, event_type, scope,
		       actor_name, actor_id, details, correlation_id, created_at
		FROM ai_memory_audit_events
		WHERE org_id = ? AND (user_id = ? OR user_id = 0)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, orgID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit events: %w", err)
	}
	defer rows.Close()

	var events []*MemoryAuditEvent
	for rows.Next() {
		var e MemoryAuditEvent
		if err := rows.Scan(
			&e.ID, &e.OrgID, &e.UserID, &e.MemoryItemID, &e.EventType, &e.Scope,
			&e.ActorName, &e.ActorID, &e.Details, &e.CorrelationID, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, &e)
	}

	return events, nil
}

// GetMemoryStats calculates active count metrics
func (r *repository) GetMemoryStats(ctx context.Context, orgID int64, userID int64) (*MemoryStats, error) {
	userSettings, err := r.GetUserSettings(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}

	stats := &MemoryStats{
		PersonalizationEnabled: userSettings.PersonalizationEnabled,
	}

	// Active personal count
	_ = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM ai_memory_items
		WHERE org_id = ? AND user_id = ? AND scope = 'USER' AND status = 'ACTIVE'
		  AND (expires_at IS NULL OR expires_at > NOW())
	`, orgID, userID).Scan(&stats.ActivePersonalCount)

	// Active org count
	_ = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM ai_memory_items
		WHERE org_id = ? AND scope = 'ORGANIZATION' AND status = 'ACTIVE'
		  AND (expires_at IS NULL OR expires_at > NOW())
	`, orgID).Scan(&stats.ActiveOrgCount)

	// Pending review count
	_ = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM ai_memory_items
		WHERE org_id = ? AND status = 'PENDING_REVIEW'
		  AND ((scope = 'USER' AND user_id = ?) OR scope = 'ORGANIZATION')
	`, orgID, userID).Scan(&stats.PendingReviewCount)

	// Expired count
	_ = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM ai_memory_items
		WHERE org_id = ? AND status = 'ACTIVE' AND expires_at <= NOW()
		  AND ((scope = 'USER' AND user_id = ?) OR scope = 'ORGANIZATION')
	`, orgID, userID).Scan(&stats.ExpiredCount)

	// Total audit events
	_ = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM ai_memory_audit_events
		WHERE org_id = ? AND (user_id = ? OR user_id = 0)
	`, orgID, userID).Scan(&stats.TotalAuditEvents)

	return stats, nil
}
