package automations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, auto *Automation) (*Automation, error)
	GetByID(ctx context.Context, orgID int64, id int64) (*Automation, error)
	GetByIDUnscoped(ctx context.Context, id int64) (*Automation, error)
	List(ctx context.Context, orgID int64, filter AutomationFilter) ([]*Automation, int64, error)
	Update(ctx context.Context, auto *Automation) error
	SetEnabled(ctx context.Context, orgID int64, id int64, isEnabled bool, updatedBy int64) error
	Delete(ctx context.Context, orgID int64, id int64) error
	GetDueAutomations(ctx context.Context, limit int) ([]*Automation, error)
	FindMatchingAutomations(ctx context.Context, orgID int64, triggerType string, module string) ([]*Automation, error)
	UpdateExecutionTimestamps(ctx context.Context, id int64, lastExec time.Time, nextExec *time.Time, status string, lastError string) error

	CreateExecution(ctx context.Context, exec *AutomationExecution) (*AutomationExecution, error)
	GetExecutionByID(ctx context.Context, orgID int64, execID int64) (*AutomationExecution, error)
	GetExecutionByIDUnscoped(ctx context.Context, execID int64) (*AutomationExecution, error)
	GetExecutionByIdempotencyKey(ctx context.Context, orgID int64, key string) (*AutomationExecution, error)
	ListExecutions(ctx context.Context, orgID int64, filter ExecutionFilter) ([]*AutomationExecution, int64, error)
	UpdateExecution(ctx context.Context, exec *AutomationExecution) error
	GetActiveExecutionForAutomation(ctx context.Context, orgID int64, autoID int64) (*AutomationExecution, error)
	GetStats(ctx context.Context, orgID int64) (*AutomationStats, error)

	// Operational Insights
	CreateInsight(ctx context.Context, insight *OperationalInsight) (*OperationalInsight, error)
	GetInsightByID(ctx context.Context, orgID int64, id int64) (*OperationalInsight, error)
	ListInsights(ctx context.Context, orgID int64, filter OperationalInsightFilter) ([]*OperationalInsight, int64, error)
	UpdateInsightStatus(ctx context.Context, orgID int64, id int64, status string, actionID *int64, approvalID *int64) error
	GetActiveInsightForRecord(ctx context.Context, orgID int64, module string, recordID int64, rule string) (*OperationalInsight, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const autoSelectCols = `
	a.id, a.org_id, a.name, a.automation_type, a.description, a.is_enabled,
	a.trigger_type, a.trigger_config, a.scope, a.approval_policy, a.allowed_actions,
	a.owner_team, a.priority,
	a.schedule_type, a.schedule_time, a.schedule_days, a.timezone,
	a.execution_window_minutes, a.configuration, a.target_modules,
	a.last_execution_at, a.next_execution_at, a.last_execution_status,
	a.last_error, a.last_successful_run, a.last_failed_run,
	a.retry_count, a.max_retries, a.retry_policy, a.max_execution_duration_sec,
	a.correlation_id, a.created_by, a.updated_by, a.created_at, a.updated_at,
	COALESCE(CONCAT(u.first_name, ' ', u.last_name), 'System') AS creator_name
`

func (r *repository) Create(ctx context.Context, auto *Automation) (*Automation, error) {
	if auto.TriggerType == "" {
		auto.TriggerType = TriggerTypeScheduled
	}
	if auto.Scope == "" {
		auto.Scope = "ORGANIZATION"
	}
	if auto.ApprovalPolicy == "" {
		auto.ApprovalPolicy = ApprovalPolicyAlwaysRequire
	}
	if auto.OwnerTeam == "" {
		auto.OwnerTeam = "OPERATIONS"
	}
	if auto.Priority == "" {
		auto.Priority = "MEDIUM"
	}
	if auto.MaxExecutionDurationSec <= 0 {
		auto.MaxExecutionDurationSec = 300
	}

	query := `
		INSERT INTO ai_automations (
			org_id, name, automation_type, description, is_enabled,
			trigger_type, trigger_config, scope, approval_policy, allowed_actions,
			owner_team, priority,
			schedule_type, schedule_time, schedule_days, timezone,
			execution_window_minutes, configuration, target_modules,
			next_execution_at, retry_policy, max_execution_duration_sec,
			correlation_id, created_by, updated_by,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?,
			?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?,
			?, ?, ?,
			NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(
		ctx, query,
		auto.OrgID, auto.Name, auto.AutomationType, auto.Description, auto.IsEnabled,
		auto.TriggerType, auto.TriggerConfigJSON, auto.Scope, auto.ApprovalPolicy, auto.AllowedActionsJSON,
		auto.OwnerTeam, auto.Priority,
		auto.ScheduleType, auto.ScheduleTime, auto.ScheduleDays, auto.Timezone,
		auto.ExecutionWindowMinutes, auto.ConfigurationJSON, auto.TargetModulesJSON,
		auto.NextExecutionAt, auto.RetryPolicyJSON, auto.MaxExecutionDurationSec,
		auto.CorrelationID, auto.CreatedBy, auto.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create automation: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get insert id: %w", err)
	}

	return r.GetByID(ctx, auto.OrgID, id)
}

func (r *repository) GetByID(ctx context.Context, orgID int64, id int64) (*Automation, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_automations a
		LEFT JOIN users u ON a.created_by = u.id
		WHERE a.id = ? AND a.org_id = ?
		LIMIT 1
	`, autoSelectCols)

	var auto Automation
	if err := r.db.GetContext(ctx, &auto, query, id, orgID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAutomationNotFound
		}
		return nil, fmt.Errorf("failed to get automation #%d: %w", id, err)
	}

	auto.UnmarshalJSONFields()
	r.enrichActiveExecution(ctx, &auto)
	return &auto, nil
}

func (r *repository) GetByIDUnscoped(ctx context.Context, id int64) (*Automation, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_automations a
		LEFT JOIN users u ON a.created_by = u.id
		WHERE a.id = ?
		LIMIT 1
	`, autoSelectCols)

	var auto Automation
	if err := r.db.GetContext(ctx, &auto, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAutomationNotFound
		}
		return nil, fmt.Errorf("failed to get automation #%d: %w", id, err)
	}

	auto.UnmarshalJSONFields()
	r.enrichActiveExecution(ctx, &auto)
	return &auto, nil
}

func (r *repository) enrichActiveExecution(ctx context.Context, auto *Automation) {
	var activeID int64
	err := r.db.GetContext(ctx, &activeID, `
		SELECT id FROM ai_automation_executions
		WHERE automation_id = ? AND status IN ('QUEUED', 'RUNNING', 'WAITING_FOR_APPROVAL', 'EXECUTING_ACTION')
		ORDER BY id DESC LIMIT 1
	`, auto.ID)
	if err == nil && activeID > 0 {
		auto.IsRunning = true
		auto.ActiveExecutionID = &activeID
	}
}

func (r *repository) List(ctx context.Context, orgID int64, filter AutomationFilter) ([]*Automation, int64, error) {
	where := []string{"a.org_id = ?"}
	args := []interface{}{orgID}

	if filter.AutomationType != "" {
		where = append(where, "a.automation_type = ?")
		args = append(args, filter.AutomationType)
	}
	if filter.TriggerType != "" {
		where = append(where, "a.trigger_type = ?")
		args = append(args, filter.TriggerType)
	}
	if filter.Scope != "" {
		where = append(where, "a.scope = ?")
		args = append(args, filter.Scope)
	}
	if filter.OwnerTeam != "" {
		where = append(where, "a.owner_team = ?")
		args = append(args, filter.OwnerTeam)
	}
	if filter.IsEnabled != nil {
		where = append(where, "a.is_enabled = ?")
		args = append(args, *filter.IsEnabled)
	}
	if filter.Search != "" {
		searchPattern := "%" + strings.ToLower(filter.Search) + "%"
		where = append(where, "(LOWER(a.name) LIKE ? OR LOWER(COALESCE(a.description, '')) LIKE ?)")
		args = append(args, searchPattern, searchPattern)
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ai_automations a WHERE %s", whereClause)
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count automations: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	sortCol := "a.created_at"
	switch filter.SortBy {
	case "name":
		sortCol = "a.name"
	case "next_execution_at":
		sortCol = "a.next_execution_at"
	case "last_execution_at":
		sortCol = "a.last_execution_at"
	case "priority":
		sortCol = "a.priority"
	case "created_at":
		sortCol = "a.created_at"
	}

	sortDir := "DESC"
	if strings.ToUpper(filter.SortDir) == "ASC" {
		sortDir = "ASC"
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_automations a
		LEFT JOIN users u ON a.created_by = u.id
		WHERE %s
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, autoSelectCols, whereClause, sortCol, sortDir, limit, offset)

	var list []*Automation
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to list automations: %w", err)
	}

	for _, a := range list {
		a.UnmarshalJSONFields()
		r.enrichActiveExecution(ctx, a)
	}

	return list, total, nil
}

func (r *repository) Update(ctx context.Context, auto *Automation) error {
	query := `
		UPDATE ai_automations SET
			name = ?,
			description = ?,
			trigger_type = ?,
			trigger_config = ?,
			scope = ?,
			approval_policy = ?,
			allowed_actions = ?,
			owner_team = ?,
			priority = ?,
			schedule_type = ?,
			schedule_time = ?,
			schedule_days = ?,
			timezone = ?,
			execution_window_minutes = ?,
			configuration = ?,
			target_modules = ?,
			retry_policy = ?,
			max_execution_duration_sec = ?,
			updated_by = ?,
			updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(
		ctx, query,
		auto.Name, auto.Description,
		auto.TriggerType, auto.TriggerConfigJSON, auto.Scope, auto.ApprovalPolicy, auto.AllowedActionsJSON,
		auto.OwnerTeam, auto.Priority,
		auto.ScheduleType, auto.ScheduleTime, auto.ScheduleDays, auto.Timezone,
		auto.ExecutionWindowMinutes, auto.ConfigurationJSON, auto.TargetModulesJSON,
		auto.RetryPolicyJSON, auto.MaxExecutionDurationSec,
		auto.UpdatedBy, auto.ID, auto.OrgID,
	)
	if err != nil {
		return fmt.Errorf("failed to update automation #%d: %w", auto.ID, err)
	}
	return nil
}

func (r *repository) SetEnabled(ctx context.Context, orgID int64, id int64, isEnabled bool, updatedBy int64) error {
	query := `
		UPDATE ai_automations SET
			is_enabled = ?,
			updated_by = ?,
			updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, isEnabled, updatedBy, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to update is_enabled on automation #%d: %w", id, err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, orgID int64, id int64) error {
	query := `DELETE FROM ai_automations WHERE id = ? AND org_id = ?`
	_, err := r.db.ExecContext(ctx, query, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete automation #%d: %w", id, err)
	}
	return nil
}

func (r *repository) GetDueAutomations(ctx context.Context, limit int) ([]*Automation, error) {
	if limit <= 0 {
		limit = 10
	}
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_automations a
		LEFT JOIN users u ON a.created_by = u.id
		WHERE a.is_enabled = 1
		  AND a.trigger_type = 'SCHEDULED'
		  AND a.next_execution_at IS NOT NULL
		  AND a.next_execution_at <= NOW()
		ORDER BY a.next_execution_at ASC
		LIMIT %d
	`, autoSelectCols, limit)

	var list []*Automation
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("failed to get due automations: %w", err)
	}

	for _, a := range list {
		a.UnmarshalJSONFields()
	}
	return list, nil
}

func (r *repository) FindMatchingAutomations(ctx context.Context, orgID int64, triggerType string, module string) ([]*Automation, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_automations a
		LEFT JOIN users u ON a.created_by = u.id
		WHERE a.org_id = ?
		  AND a.is_enabled = 1
		  AND (a.trigger_type = ? OR a.trigger_type = 'MANUAL')
		ORDER BY a.priority DESC, a.id ASC
	`, autoSelectCols)

	var list []*Automation
	if err := r.db.SelectContext(ctx, &list, query, orgID, triggerType); err != nil {
		return nil, fmt.Errorf("failed to find matching automations: %w", err)
	}

	var matched []*Automation
	for _, a := range list {
		a.UnmarshalJSONFields()
		// If specific target modules are configured, filter by module
		if len(a.TargetModules) > 0 && module != "" {
			hasModule := false
			for _, m := range a.TargetModules {
				if strings.EqualFold(m, module) {
					hasModule = true
					break
				}
			}
			if hasModule {
				matched = append(matched, a)
			}
		} else {
			matched = append(matched, a)
		}
	}
	return matched, nil
}

func (r *repository) UpdateExecutionTimestamps(ctx context.Context, id int64, lastExec time.Time, nextExec *time.Time, status string, lastError string) error {
	var successUpdate, failureUpdate string
	if status == ExecutionStatusCompleted {
		successUpdate = ", last_successful_run = NOW()"
	} else if status == ExecutionStatusFailed {
		failureUpdate = ", last_failed_run = NOW()"
	}

	query := fmt.Sprintf(`
		UPDATE ai_automations SET
			last_execution_at = ?,
			next_execution_at = ?,
			last_execution_status = ?,
			last_error = ?%s%s,
			updated_at = NOW()
		WHERE id = ?
	`, successUpdate, failureUpdate)

	_, err := r.db.ExecContext(ctx, query, lastExec, nextExec, status, lastError, id)
	return err
}

// ── Execution Repository Methods ─────────────────────────────────────────────

const execSelectCols = `
	e.id, e.automation_id, e.org_id, e.correlation_id, e.trigger_type,
	e.trigger_event, e.input_record_ref, e.current_step, e.step_results,
	e.triggered_by_user_id, e.status, e.queued_at, e.started_at, e.completed_at,
	e.duration_ms, e.records_reviewed, e.recommendations_created, e.recommendations_updated,
	e.summary_text, e.error_message, e.details, e.approval_id, e.action_id, e.idempotency_key,
	e.retry_count, e.created_at, e.updated_at,
	a.name AS automation_name, a.automation_type,
	COALESCE(CONCAT(u.first_name, ' ', u.last_name), 'System') AS triggered_by_name
`

func (r *repository) CreateExecution(ctx context.Context, exec *AutomationExecution) (*AutomationExecution, error) {
	if exec.CurrentStep == "" {
		exec.CurrentStep = StepInitializing
	}
	query := `
		INSERT INTO ai_automation_executions (
			automation_id, org_id, correlation_id, trigger_type,
			trigger_event, input_record_ref, current_step, step_results,
			triggered_by_user_id, status, queued_at, duration_ms,
			records_reviewed, recommendations_created, recommendations_updated,
			summary_text, error_message, details, approval_id, action_id,
			idempotency_key, retry_count,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, NOW(), 0,
			0, 0, 0,
			?, ?, ?, ?, ?,
			?, ?,
			NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(
		ctx, query,
		exec.AutomationID, exec.OrgID, exec.CorrelationID, exec.TriggerType,
		exec.TriggerEvent, exec.InputRecordRef, exec.CurrentStep, exec.StepResultsJSON,
		exec.TriggeredByUserID, exec.Status,
		exec.SummaryText, exec.ErrorMessage, exec.DetailsJSON, exec.ApprovalID, exec.ActionID,
		exec.IdempotencyKey, exec.RetryCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get execution id: %w", err)
	}

	return r.GetExecutionByID(ctx, exec.OrgID, id)
}

func (r *repository) GetExecutionByID(ctx context.Context, orgID int64, execID int64) (*AutomationExecution, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_automation_executions e
		JOIN ai_automations a ON e.automation_id = a.id
		LEFT JOIN users u ON e.triggered_by_user_id = u.id
		WHERE e.id = ? AND e.org_id = ?
		LIMIT 1
	`, execSelectCols)

	var exec AutomationExecution
	if err := r.db.GetContext(ctx, &exec, query, execID, orgID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrExecutionNotFound
		}
		return nil, fmt.Errorf("failed to get execution #%d: %w", execID, err)
	}

	exec.UnmarshalJSONFields()
	return &exec, nil
}

func (r *repository) GetExecutionByIDUnscoped(ctx context.Context, execID int64) (*AutomationExecution, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_automation_executions e
		JOIN ai_automations a ON e.automation_id = a.id
		LEFT JOIN users u ON e.triggered_by_user_id = u.id
		WHERE e.id = ?
		LIMIT 1
	`, execSelectCols)

	var exec AutomationExecution
	if err := r.db.GetContext(ctx, &exec, query, execID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrExecutionNotFound
		}
		return nil, fmt.Errorf("failed to get execution #%d: %w", execID, err)
	}

	exec.UnmarshalJSONFields()
	return &exec, nil
}

func (r *repository) GetExecutionByIdempotencyKey(ctx context.Context, orgID int64, key string) (*AutomationExecution, error) {
	if key == "" {
		return nil, nil
	}
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_automation_executions e
		JOIN ai_automations a ON e.automation_id = a.id
		LEFT JOIN users u ON e.triggered_by_user_id = u.id
		WHERE e.org_id = ? AND e.idempotency_key = ?
		ORDER BY e.id DESC
		LIMIT 1
	`, execSelectCols)

	var exec AutomationExecution
	if err := r.db.GetContext(ctx, &exec, query, orgID, key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	exec.UnmarshalJSONFields()
	return &exec, nil
}

func (r *repository) ListExecutions(ctx context.Context, orgID int64, filter ExecutionFilter) ([]*AutomationExecution, int64, error) {
	where := []string{"e.org_id = ?"}
	args := []interface{}{orgID}

	if filter.AutomationID != nil && *filter.AutomationID > 0 {
		where = append(where, "e.automation_id = ?")
		args = append(args, *filter.AutomationID)
	}
	if filter.Status != "" {
		where = append(where, "e.status = ?")
		args = append(args, filter.Status)
	}
	if filter.TriggerType != "" {
		where = append(where, "e.trigger_type = ?")
		args = append(args, filter.TriggerType)
	}
	if filter.CurrentStep != "" {
		where = append(where, "e.current_step = ?")
		args = append(args, filter.CurrentStep)
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ai_automation_executions e WHERE %s", whereClause)
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count executions: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_automation_executions e
		JOIN ai_automations a ON e.automation_id = a.id
		LEFT JOIN users u ON e.triggered_by_user_id = u.id
		WHERE %s
		ORDER BY e.created_at DESC
		LIMIT %d OFFSET %d
	`, execSelectCols, whereClause, limit, offset)

	var list []*AutomationExecution
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to list executions: %w", err)
	}

	for _, e := range list {
		e.UnmarshalJSONFields()
	}

	return list, total, nil
}

func (r *repository) UpdateExecution(ctx context.Context, exec *AutomationExecution) error {
	query := `
		UPDATE ai_automation_executions SET
			status = ?,
			current_step = ?,
			step_results = ?,
			started_at = ?,
			completed_at = ?,
			duration_ms = ?,
			records_reviewed = ?,
			recommendations_created = ?,
			recommendations_updated = ?,
			summary_text = ?,
			error_message = ?,
			details = ?,
			approval_id = ?,
			action_id = ?,
			retry_count = ?,
			updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(
		ctx, query,
		exec.Status, exec.CurrentStep, exec.StepResultsJSON,
		exec.StartedAt, exec.CompletedAt, exec.DurationMs,
		exec.RecordsReviewed, exec.RecommendationsCreated, exec.RecommendationsUpdated,
		exec.SummaryText, exec.ErrorMessage, exec.DetailsJSON,
		exec.ApprovalID, exec.ActionID, exec.RetryCount,
		exec.ID, exec.OrgID,
	)
	if err != nil {
		return fmt.Errorf("failed to update execution #%d: %w", exec.ID, err)
	}
	return nil
}

func (r *repository) GetActiveExecutionForAutomation(ctx context.Context, orgID int64, autoID int64) (*AutomationExecution, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_automation_executions e
		JOIN ai_automations a ON e.automation_id = a.id
		LEFT JOIN users u ON e.triggered_by_user_id = u.id
		WHERE e.automation_id = ? AND e.org_id = ? AND e.status IN ('QUEUED', 'RUNNING', 'WAITING_FOR_APPROVAL', 'EXECUTING_ACTION')
		ORDER BY e.id DESC
		LIMIT 1
	`, execSelectCols)

	var exec AutomationExecution
	if err := r.db.GetContext(ctx, &exec, query, autoID, orgID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	exec.UnmarshalJSONFields()
	return &exec, nil
}

func (r *repository) GetStats(ctx context.Context, orgID int64) (*AutomationStats, error) {
	var totalAutos, activeAutos int
	_ = r.db.GetContext(ctx, &totalAutos, "SELECT COUNT(*) FROM ai_automations WHERE org_id = ?", orgID)
	_ = r.db.GetContext(ctx, &activeAutos, "SELECT COUNT(*) FROM ai_automations WHERE org_id = ? AND is_enabled = 1", orgID)

	var totalExecs, successExecs, failedExecs, waitingApprovalExecs int
	_ = r.db.GetContext(ctx, &totalExecs, "SELECT COUNT(*) FROM ai_automation_executions WHERE org_id = ?", orgID)
	_ = r.db.GetContext(ctx, &successExecs, "SELECT COUNT(*) FROM ai_automation_executions WHERE org_id = ? AND status = 'COMPLETED'", orgID)
	_ = r.db.GetContext(ctx, &failedExecs, "SELECT COUNT(*) FROM ai_automation_executions WHERE org_id = ? AND status = 'FAILED'", orgID)
	_ = r.db.GetContext(ctx, &waitingApprovalExecs, "SELECT COUNT(*) FROM ai_automation_executions WHERE org_id = ? AND status = 'WAITING_FOR_APPROVAL'", orgID)

	var activeInsights int
	_ = r.db.GetContext(ctx, &activeInsights, "SELECT COUNT(*) FROM ai_operational_insights WHERE org_id = ? AND status = 'ACTIVE'", orgID)

	var recsCreated, recsUpdated int
	_ = r.db.GetContext(ctx, &recsCreated, "SELECT COALESCE(SUM(recommendations_created), 0) FROM ai_automation_executions WHERE org_id = ?", orgID)
	_ = r.db.GetContext(ctx, &recsUpdated, "SELECT COALESCE(SUM(recommendations_updated), 0) FROM ai_automation_executions WHERE org_id = ?", orgID)

	successRate := 100.0
	if totalExecs > 0 {
		successRate = float64(successExecs) / float64(totalExecs) * 100.0
	}

	return &AutomationStats{
		TotalAutomations:         totalAutos,
		ActiveAutomations:        activeAutos,
		TotalExecutions:          totalExecs,
		SuccessfulExecutions:     successExecs,
		FailedExecutions:         failedExecs,
		WaitingApprovalCount:     waitingApprovalExecs,
		ActiveInsightsCount:      activeInsights,
		RecommendationsGenerated: recsCreated,
		RecommendationsUpdated:   recsUpdated,
		RecentSuccessRate:        successRate,
	}, nil
}

// ── Operational Insights Repository Methods ──────────────────────────────────

const insightSelectCols = `
	o.id, o.org_id, o.automation_id, o.execution_id, o.source_module,
	o.source_record_id, o.source_record_ref, o.insight_type, o.severity,
	o.title, o.description, o.evidence, o.detection_rule, o.confidence,
	o.priority, o.risk_level, o.data_freshness, o.recommended_next_step,
	o.is_approval_required, o.recommended_action_type, o.action_payload,
	o.status, o.action_id, o.approval_id, o.correlation_id,
	o.created_at, o.updated_at,
	COALESCE(a.name, 'Real-time Signal Engine') AS automation_name
`

func (r *repository) CreateInsight(ctx context.Context, insight *OperationalInsight) (*OperationalInsight, error) {
	if insight.Status == "" {
		insight.Status = InsightStatusActive
	}
	if insight.Severity == "" {
		insight.Severity = InsightSeverityMedium
	}
	if insight.Priority == "" {
		insight.Priority = "MEDIUM"
	}
	if insight.RiskLevel == "" {
		insight.RiskLevel = "LOW"
	}
	if insight.DataFreshness == "" {
		insight.DataFreshness = "REAL_TIME"
	}
	if insight.Confidence <= 0 {
		insight.Confidence = 1.0
	}

	query := `
		INSERT INTO ai_operational_insights (
			org_id, automation_id, execution_id, source_module,
			source_record_id, source_record_ref, insight_type, severity,
			title, description, evidence, detection_rule, confidence,
			priority, risk_level, data_freshness, recommended_next_step,
			is_approval_required, recommended_action_type, action_payload,
			status, action_id, approval_id, correlation_id,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?, ?,
			NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(
		ctx, query,
		insight.OrgID, insight.AutomationID, insight.ExecutionID, insight.SourceModule,
		insight.SourceRecordID, insight.SourceRecordRef, insight.InsightType, insight.Severity,
		insight.Title, insight.Description, insight.EvidenceJSON, insight.DetectionRule, insight.Confidence,
		insight.Priority, insight.RiskLevel, insight.DataFreshness, insight.RecommendedNextStep,
		insight.IsApprovalRequired, insight.RecommendedActionType, insight.ActionPayloadJSON,
		insight.Status, insight.ActionID, insight.ApprovalID, insight.CorrelationID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create operational insight: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get operational insight id: %w", err)
	}

	return r.GetInsightByID(ctx, insight.OrgID, id)
}

func (r *repository) GetInsightByID(ctx context.Context, orgID int64, id int64) (*OperationalInsight, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_operational_insights o
		LEFT JOIN ai_automations a ON o.automation_id = a.id
		WHERE o.id = ? AND o.org_id = ?
		LIMIT 1
	`, insightSelectCols)

	var insight OperationalInsight
	if err := r.db.GetContext(ctx, &insight, query, id, orgID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInsightNotFound
		}
		return nil, fmt.Errorf("failed to get operational insight #%d: %w", id, err)
	}

	insight.UnmarshalJSONFields()
	return &insight, nil
}

func (r *repository) ListInsights(ctx context.Context, orgID int64, filter OperationalInsightFilter) ([]*OperationalInsight, int64, error) {
	where := []string{"o.org_id = ?"}
	args := []interface{}{orgID}

	if filter.SourceModule != "" {
		where = append(where, "o.source_module = ?")
		args = append(args, filter.SourceModule)
	}
	if filter.Status != "" {
		where = append(where, "o.status = ?")
		args = append(args, filter.Status)
	}
	if filter.Severity != "" {
		where = append(where, "o.severity = ?")
		args = append(args, filter.Severity)
	}
	if filter.Priority != "" {
		where = append(where, "o.priority = ?")
		args = append(args, filter.Priority)
	}
	if filter.AutomationID != nil && *filter.AutomationID > 0 {
		where = append(where, "o.automation_id = ?")
		args = append(args, *filter.AutomationID)
	}
	if filter.ExecutionID != nil && *filter.ExecutionID > 0 {
		where = append(where, "o.execution_id = ?")
		args = append(args, *filter.ExecutionID)
	}
	if filter.SourceRecordID != nil && *filter.SourceRecordID > 0 {
		where = append(where, "o.source_record_id = ?")
		args = append(args, *filter.SourceRecordID)
	}
	if filter.Search != "" {
		searchPattern := "%" + strings.ToLower(filter.Search) + "%"
		where = append(where, "(LOWER(o.title) LIKE ? OR LOWER(o.description) LIKE ? OR LOWER(COALESCE(o.source_record_ref, '')) LIKE ?)")
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ai_operational_insights o WHERE %s", whereClause)
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count operational insights: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	sortCol := "o.created_at"
	switch filter.SortBy {
	case "severity":
		sortCol = "FIELD(o.severity, 'CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'INFO')"
	case "priority":
		sortCol = "FIELD(o.priority, 'CRITICAL', 'HIGH', 'MEDIUM', 'LOW')"
	case "created_at":
		sortCol = "o.created_at"
	}

	sortDir := "DESC"
	if strings.ToUpper(filter.SortDir) == "ASC" {
		sortDir = "ASC"
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_operational_insights o
		LEFT JOIN ai_automations a ON o.automation_id = a.id
		WHERE %s
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, insightSelectCols, whereClause, sortCol, sortDir, limit, offset)

	var list []*OperationalInsight
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to list operational insights: %w", err)
	}

	for _, o := range list {
		o.UnmarshalJSONFields()
	}

	return list, total, nil
}

func (r *repository) UpdateInsightStatus(ctx context.Context, orgID int64, id int64, status string, actionID *int64, approvalID *int64) error {
	query := `
		UPDATE ai_operational_insights SET
			status = ?,
			action_id = COALESCE(?, action_id),
			approval_id = COALESCE(?, approval_id),
			updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, status, actionID, approvalID, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to update operational insight #%d status: %w", id, err)
	}
	return nil
}

func (r *repository) GetActiveInsightForRecord(ctx context.Context, orgID int64, module string, recordID int64, rule string) (*OperationalInsight, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_operational_insights o
		LEFT JOIN ai_automations a ON o.automation_id = a.id
		WHERE o.org_id = ? AND o.source_module = ? AND o.source_record_id = ? AND o.detection_rule = ? AND o.status = 'ACTIVE'
		ORDER BY o.id DESC
		LIMIT 1
	`, insightSelectCols)

	var insight OperationalInsight
	if err := r.db.GetContext(ctx, &insight, query, orgID, module, recordID, rule); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	insight.UnmarshalJSONFields()
	return &insight, nil
}
