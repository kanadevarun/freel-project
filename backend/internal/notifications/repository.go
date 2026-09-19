package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	List(ctx context.Context, filter NotificationFilter) ([]Notification, int, error)
	GetByID(ctx context.Context, orgID int64, id int64) (*Notification, error)
	GetByDedupHash(ctx context.Context, orgID int64, dedupHash string) (*Notification, error)
	GetUnreadCount(ctx context.Context, orgID int64, userID int64, role string) (int, error)
	GetStats(ctx context.Context, orgID int64, userID int64, role string) (*NotificationStats, error)
	MarkRead(ctx context.Context, orgID int64, id int64, read bool) error
	MarkAllRead(ctx context.Context, orgID int64, userID int64, role string) error
	Dismiss(ctx context.Context, orgID int64, id int64) error
	Acknowledge(ctx context.Context, orgID int64, id int64, userID int64) error
	Snooze(ctx context.Context, orgID int64, id int64, snoozedUntil time.Time) error
	UpdateAIDetails(ctx context.Context, orgID int64, id int64, summary, reason, groupKey string, priorityScore float64) error
	Create(ctx context.Context, n *Notification) error
	UpdateStatus(ctx context.Context, orgID int64, id int64, status string) error
	Escalate(ctx context.Context, orgID int64, id int64, newSeverity string, reason string, triggerType string, thresholdHours int, correlationID string) error
	CreateEscalationEvent(ctx context.Context, e *EscalationEvent) (*EscalationEvent, error)
	ListEscalations(ctx context.Context, orgID int64, limit int) ([]EscalationEvent, error)
	GetPreferences(ctx context.Context, orgID int64, userID int64) (*UserNotificationPreferences, error)
	UpdatePreferences(ctx context.Context, prefs *UserNotificationPreferences) error
}

type sqlRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) List(ctx context.Context, filter NotificationFilter) ([]Notification, int, error) {
	where := []string{"org_id = ?"}
	args := []interface{}{filter.OrgID}

	// Recipient scoping: accessible if user_id is NULL or matches user, and role_target matches or user is SUPER_ADMIN
	if filter.UserID > 0 {
		if strings.EqualFold(filter.UserRole, "SUPER_ADMIN") || strings.EqualFold(filter.UserRole, "ADMIN") {
			where = append(where, "(user_id IS NULL OR user_id = ?)")
			args = append(args, filter.UserID)
		} else {
			where = append(where, "(user_id IS NULL OR user_id = ?) AND (role_target IS NULL OR role_target = ?)")
			args = append(args, filter.UserID, filter.UserRole)
		}
	}

	if filter.Severity != "" && filter.Severity != "ALL" {
		where = append(where, "severity = ?")
		args = append(args, filter.Severity)
	}

	if filter.SourceModule != "" && filter.SourceModule != "ALL" {
		where = append(where, "source_module = ?")
		args = append(args, filter.SourceModule)
	}

	if filter.IsRead != nil {
		where = append(where, "is_read = ?")
		args = append(args, *filter.IsRead)
	}

	if filter.IsDismissed != nil {
		where = append(where, "is_dismissed = ?")
		args = append(args, *filter.IsDismissed)
	} else {
		// By default, exclude dismissed notifications from general list
		where = append(where, "is_dismissed = 0")
	}

	if filter.ActionRequired != nil {
		where = append(where, "action_required = ?")
		args = append(args, *filter.ActionRequired)
	}

	if filter.IsEscalated != nil {
		where = append(where, "is_escalated = ?")
		args = append(args, *filter.IsEscalated)
	}

	if filter.DeliveryStatus != "" && filter.DeliveryStatus != "ALL" {
		where = append(where, "delivery_status = ?")
		args = append(args, filter.DeliveryStatus)
	}

	if filter.IsAcknowledged != nil {
		where = append(where, "is_acknowledged = ?")
		args = append(args, *filter.IsAcknowledged)
	}

	if filter.IsSnoozed != nil {
		where = append(where, "is_snoozed = ?")
		args = append(args, *filter.IsSnoozed)
	}

	if filter.GroupKey != "" {
		where = append(where, "group_key = ?")
		args = append(args, filter.GroupKey)
	}

	if filter.Search != "" {
		where = append(where, "(title LIKE ? OR message LIKE ? OR source_record_id LIKE ?)")
		pattern := "%" + filter.Search + "%"
		args = append(args, pattern, pattern, pattern)
	}

	whereClause := strings.Join(where, " AND ")

	// Count total matching
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notifications WHERE %s", whereClause)
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Fetch page
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	listQuery := fmt.Sprintf(`
		SELECT * FROM notifications 
		WHERE %s 
		ORDER BY is_escalated DESC, 
			CASE severity 
				WHEN 'CRITICAL' THEN 1 
				WHEN 'HIGH' THEN 2 
				WHEN 'MEDIUM' THEN 3 
				WHEN 'LOW' THEN 4 
				ELSE 5 
			END, 
			created_at DESC 
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)
	var items []Notification
	if err := r.db.SelectContext(ctx, &items, listQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}

	return items, total, nil
}

func (r *sqlRepository) GetByID(ctx context.Context, orgID int64, id int64) (*Notification, error) {
	query := "SELECT * FROM notifications WHERE org_id = ? AND id = ?"
	var notif Notification
	if err := r.db.GetContext(ctx, &notif, query, orgID, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get notification by id: %w", err)
	}
	return &notif, nil
}

func (r *sqlRepository) GetByDedupHash(ctx context.Context, orgID int64, dedupHash string) (*Notification, error) {
	query := "SELECT * FROM notifications WHERE org_id = ? AND dedup_hash = ?"
	var notif Notification
	if err := r.db.GetContext(ctx, &notif, query, orgID, dedupHash); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get notification by dedup hash: %w", err)
	}
	return &notif, nil
}

func (r *sqlRepository) GetUnreadCount(ctx context.Context, orgID int64, userID int64, role string) (int, error) {
	where := []string{"org_id = ?", "is_read = 0", "is_dismissed = 0"}
	args := []interface{}{orgID}

	if userID > 0 {
		if strings.EqualFold(role, "SUPER_ADMIN") || strings.EqualFold(role, "ADMIN") {
			where = append(where, "(user_id IS NULL OR user_id = ?)")
			args = append(args, userID)
		} else {
			where = append(where, "(user_id IS NULL OR user_id = ?) AND (role_target IS NULL OR role_target = ?)")
			args = append(args, userID, role)
		}
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM notifications WHERE %s", strings.Join(where, " AND "))
	var count int
	if err := r.db.GetContext(ctx, &count, query, args...); err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}

func (r *sqlRepository) GetStats(ctx context.Context, orgID int64, userID int64, role string) (*NotificationStats, error) {
	where := []string{"org_id = ?", "is_dismissed = 0"}
	args := []interface{}{orgID}

	if userID > 0 {
		if strings.EqualFold(role, "SUPER_ADMIN") || strings.EqualFold(role, "ADMIN") {
			where = append(where, "(user_id IS NULL OR user_id = ?)")
			args = append(args, userID)
		} else {
			where = append(where, "(user_id IS NULL OR user_id = ?) AND (role_target IS NULL OR role_target = ?)")
			args = append(args, userID, role)
		}
	}

	whereClause := strings.Join(where, " AND ")
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN is_read = 0 THEN 1 ELSE 0 END), 0) as unread,
			COALESCE(SUM(CASE WHEN action_required = 1 AND is_read = 0 THEN 1 ELSE 0 END), 0) as action_required,
			COALESCE(SUM(CASE WHEN is_escalated = 1 THEN 1 ELSE 0 END), 0) as escalated,
			COALESCE(SUM(CASE WHEN is_acknowledged = 1 THEN 1 ELSE 0 END), 0) as acknowledged,
			COALESCE(SUM(CASE WHEN is_snoozed = 1 AND (snoozed_until IS NULL OR snoozed_until > NOW()) THEN 1 ELSE 0 END), 0) as snoozed,
			COALESCE(SUM(CASE WHEN severity = 'CRITICAL' THEN 1 ELSE 0 END), 0) as critical_count,
			COALESCE(SUM(CASE WHEN severity = 'HIGH' THEN 1 ELSE 0 END), 0) as high_count
		FROM notifications 
		WHERE %s
	`, whereClause)

	type statsRow struct {
		Total          int `db:"total"`
		Unread         int `db:"unread"`
		ActionRequired int `db:"action_required"`
		Escalated      int `db:"escalated"`
		Acknowledged   int `db:"acknowledged"`
		Snoozed        int `db:"snoozed"`
		CriticalCount  int `db:"critical_count"`
		HighCount      int `db:"high_count"`
	}

	var row statsRow
	if err := r.db.GetContext(ctx, &row, query, args...); err != nil {
		return nil, fmt.Errorf("failed to get notification stats: %w", err)
	}

	return &NotificationStats{
		Total:          row.Total,
		Unread:         row.Unread,
		ActionRequired: row.ActionRequired,
		Escalated:      row.Escalated,
		Acknowledged:   row.Acknowledged,
		Snoozed:        row.Snoozed,
		CriticalCount:  row.CriticalCount,
		HighCount:      row.HighCount,
	}, nil
}

func (r *sqlRepository) MarkRead(ctx context.Context, orgID int64, id int64, read bool) error {
	var query string
	if read {
		query = "UPDATE notifications SET is_read = 1, read_at = NOW(), updated_at = NOW() WHERE org_id = ? AND id = ?"
	} else {
		query = "UPDATE notifications SET is_read = 0, read_at = NULL, updated_at = NOW() WHERE org_id = ? AND id = ?"
	}
	res, err := r.db.ExecContext(ctx, query, orgID, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("notification %d not found in organization %d", id, orgID)
	}
	return nil
}

func (r *sqlRepository) MarkAllRead(ctx context.Context, orgID int64, userID int64, role string) error {
	where := []string{"org_id = ?", "is_read = 0"}
	args := []interface{}{orgID}

	if userID > 0 {
		if strings.EqualFold(role, "SUPER_ADMIN") || strings.EqualFold(role, "ADMIN") {
			where = append(where, "(user_id IS NULL OR user_id = ?)")
			args = append(args, userID)
		} else {
			where = append(where, "(user_id IS NULL OR user_id = ?) AND (role_target IS NULL OR role_target = ?)")
			args = append(args, userID, role)
		}
	}

	query := fmt.Sprintf("UPDATE notifications SET is_read = 1, read_at = NOW(), updated_at = NOW() WHERE %s", strings.Join(where, " AND "))
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *sqlRepository) Dismiss(ctx context.Context, orgID int64, id int64) error {
	query := "UPDATE notifications SET is_dismissed = 1, dismissed_at = NOW(), status = 'DISMISSED', delivery_status = 'DISMISSED', updated_at = NOW() WHERE org_id = ? AND id = ?"
	res, err := r.db.ExecContext(ctx, query, orgID, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("notification %d not found in organization %d", id, orgID)
	}
	return nil
}

func (r *sqlRepository) Acknowledge(ctx context.Context, orgID int64, id int64, userID int64) error {
	query := `UPDATE notifications 
		SET is_acknowledged = 1, acknowledged_at = NOW(), acknowledged_by = ?, delivery_status = 'ACKNOWLEDGED', is_read = 1, read_at = COALESCE(read_at, NOW()), updated_at = NOW() 
		WHERE org_id = ? AND id = ?`
	res, err := r.db.ExecContext(ctx, query, userID, orgID, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("notification %d not found in organization %d", id, orgID)
	}
	return nil
}

func (r *sqlRepository) Snooze(ctx context.Context, orgID int64, id int64, snoozedUntil time.Time) error {
	query := `UPDATE notifications 
		SET is_snoozed = 1, snoozed_until = ?, delivery_status = 'SNOOZED', updated_at = NOW() 
		WHERE org_id = ? AND id = ?`
	res, err := r.db.ExecContext(ctx, query, snoozedUntil, orgID, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("notification %d not found in organization %d", id, orgID)
	}
	return nil
}

func (r *sqlRepository) UpdateAIDetails(ctx context.Context, orgID int64, id int64, summary, reason, groupKey string, priorityScore float64) error {
	query := `UPDATE notifications 
		SET ai_summary = ?, ai_escalation_reason = ?, group_key = ?, ai_priority_score = ?, updated_at = NOW() 
		WHERE org_id = ? AND id = ?`
	res, err := r.db.ExecContext(ctx, query, summary, reason, groupKey, priorityScore, orgID, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("notification %d not found in organization %d", id, orgID)
	}
	return nil
}

func (r *sqlRepository) Create(ctx context.Context, n *Notification) error {
	if n.DeliveryStatus == "" {
		n.DeliveryStatus = "DELIVERED"
	}
	query := `
		INSERT INTO notifications (
			org_id, user_id, role_target, source_module, source_record_type, source_record_id,
			recommendation_id, approval_id, automation_execution_id, notification_type,
			title, message, severity, priority, status, delivery_status, is_read, is_dismissed,
			is_acknowledged, acknowledged_at, acknowledged_by, is_snoozed, snoozed_until,
			is_escalated, escalation_level, action_required, action_url,
			dedup_hash, group_key, correlation_id, ai_summary, ai_escalation_reason, ai_priority_score,
			expires_at, metadata, created_at, updated_at
		) VALUES (
			:org_id, :user_id, :role_target, :source_module, :source_record_type, :source_record_id,
			:recommendation_id, :approval_id, :automation_execution_id, :notification_type,
			:title, :message, :severity, :priority, :status, :delivery_status, :is_read, :is_dismissed,
			:is_acknowledged, :acknowledged_at, :acknowledged_by, :is_snoozed, :snoozed_until,
			:is_escalated, :escalation_level, :action_required, :action_url,
			:dedup_hash, :group_key, :correlation_id, :ai_summary, :ai_escalation_reason, :ai_priority_score,
			:expires_at, :metadata, NOW(), NOW()
		)
	`
	res, err := r.db.NamedExecContext(ctx, query, n)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		n.ID = id
	}
	return nil
}

func (r *sqlRepository) UpdateStatus(ctx context.Context, orgID int64, id int64, status string) error {
	query := "UPDATE notifications SET status = ?, updated_at = NOW() WHERE org_id = ? AND id = ?"
	_, err := r.db.ExecContext(ctx, query, status, orgID, id)
	return err
}

func (r *sqlRepository) Escalate(ctx context.Context, orgID int64, id int64, newSeverity string, reason string, triggerType string, thresholdHours int, correlationID string) error {
	// Fetch previous state
	notif, err := r.GetByID(ctx, orgID, id)
	if err != nil || notif == nil {
		return fmt.Errorf("notification not found for escalation: %w", err)
	}

	newLevel := notif.EscalationLevel + 1

	// Update notification table
	updateQuery := `
		UPDATE notifications 
		SET is_escalated = 1, 
			escalation_level = ?, 
			severity = ?, 
			delivery_status = 'ESCALATED',
			escalated_at = NOW(), 
			updated_at = NOW() 
		WHERE org_id = ? AND id = ?
	`
	if _, err := r.db.ExecContext(ctx, updateQuery, newLevel, newSeverity, orgID, id); err != nil {
		return fmt.Errorf("failed to update notification escalation state: %w", err)
	}

	// Insert audit escalation event
	eventQuery := `
		INSERT INTO notification_escalation_events (
			notification_id, org_id, escalation_level, escalation_reason,
			previous_severity, new_severity, trigger_type, threshold_hours,
			correlation_id, created_at
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, NOW()
		)
	`
	_, err = r.db.ExecContext(ctx, eventQuery,
		id, orgID, newLevel, reason,
		notif.Severity, newSeverity, triggerType, thresholdHours,
		correlationID,
	)
	if err != nil {
		return fmt.Errorf("failed to record escalation event: %w", err)
	}

	return nil
}

func (r *sqlRepository) CreateEscalationEvent(ctx context.Context, e *EscalationEvent) (*EscalationEvent, error) {
	query := `
		INSERT INTO notification_escalation_events (
			notification_id, org_id, escalation_level, escalation_reason,
			previous_severity, new_severity, trigger_type, threshold_hours,
			ai_escalation_summary, recommended_action, action_proposal_id, approval_id,
			acknowledged_at, acknowledged_by, resolved_at, resolved_by,
			correlation_id, created_at
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		e.NotificationID, e.OrgID, e.EscalationLevel, e.EscalationReason,
		e.PreviousSeverity, e.NewSeverity, e.TriggerType, e.ThresholdHours,
		e.AIEscalationSummary, e.RecommendedAction, e.ActionProposalID, e.ApprovalID,
		e.AcknowledgedAt, e.AcknowledgedBy, e.ResolvedAt, e.ResolvedBy,
		e.CorrelationID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert escalation event: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		e.ID = id
	}
	return e, nil
}

func (r *sqlRepository) ListEscalations(ctx context.Context, orgID int64, limit int) ([]EscalationEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT * FROM notification_escalation_events 
		WHERE org_id = ? 
		ORDER BY created_at DESC 
		LIMIT ?
	`
	var events []EscalationEvent
	if err := r.db.SelectContext(ctx, &events, query, orgID, limit); err != nil {
		return nil, fmt.Errorf("failed to list escalation events: %w", err)
	}
	return events, nil
}

func (r *sqlRepository) GetPreferences(ctx context.Context, orgID int64, userID int64) (*UserNotificationPreferences, error) {
	query := "SELECT * FROM user_notification_preferences WHERE org_id = ? AND user_id = ?"
	var prefs UserNotificationPreferences
	if err := r.db.GetContext(ctx, &prefs, query, orgID, userID); err != nil {
		if err == sql.ErrNoRows {
			// Return default preferences
			return &UserNotificationPreferences{
				UserID:                 userID,
				OrgID:                  orgID,
				MinSeverity:            SeverityInformational,
				InAppEnabled:           true,
				AssignedOnly:           false,
				ApprovalsEnabled:       true,
				AutomationsEnabled:     true,
				RecommendationsEnabled: true,
				FinanceEnabled:         true,
				OperationsEnabled:      true,
				ComplianceEnabled:      true,
			}, nil
		}
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}
	return &prefs, nil
}

func (r *sqlRepository) UpdatePreferences(ctx context.Context, prefs *UserNotificationPreferences) error {
	query := `
		INSERT INTO user_notification_preferences (
			user_id, org_id, min_severity, in_app_enabled, assigned_only,
			approvals_enabled, automations_enabled, recommendations_enabled,
			finance_enabled, operations_enabled, compliance_enabled,
			created_at, updated_at
		) VALUES (
			:user_id, :org_id, :min_severity, :in_app_enabled, :assigned_only,
			:approvals_enabled, :automations_enabled, :recommendations_enabled,
			:finance_enabled, :operations_enabled, :compliance_enabled,
			NOW(), NOW()
		)
		ON DUPLICATE KEY UPDATE
			min_severity = VALUES(min_severity),
			in_app_enabled = VALUES(in_app_enabled),
			assigned_only = VALUES(assigned_only),
			approvals_enabled = VALUES(approvals_enabled),
			automations_enabled = VALUES(automations_enabled),
			recommendations_enabled = VALUES(recommendations_enabled),
			finance_enabled = VALUES(finance_enabled),
			operations_enabled = VALUES(operations_enabled),
			compliance_enabled = VALUES(compliance_enabled),
			updated_at = NOW()
	`
	_, err := r.db.NamedExecContext(ctx, query, prefs)
	return err
}
