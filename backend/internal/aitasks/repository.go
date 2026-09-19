package aitasks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/freel/backend/internal/ai"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetTaskByID(ctx context.Context, orgID int64, taskID int64) (*Task, error)
	GetTaskByIDUnscoped(ctx context.Context, taskID int64) (*Task, error)
	ListTasks(ctx context.Context, orgID int64, filter TaskFilter) ([]*Task, int64, error)
	CreateTask(ctx context.Context, task *Task) error
	ClaimNextTask(ctx context.Context, workerID string, leaseDuration time.Duration, taskTypes []string) (*Task, error)
	HeartbeatTask(ctx context.Context, taskID int64, workerID string, leaseDuration time.Duration) error
	UpdateTaskStatus(ctx context.Context, taskID int64, input UpdateTaskStatusInput) (*Task, error)
	CancelTask(ctx context.Context, orgID int64, taskID int64, reason string) (*Task, error)
	RetryTask(ctx context.Context, orgID int64, taskID int64) (*Task, error)
	RecoverStaleTasks(ctx context.Context, defaultLeaseDuration time.Duration) (int64, int64, error)
	GetTaskStats(ctx context.Context, orgID int64) (*TaskStats, error)
	GetWorkforceSummary(ctx context.Context, orgID int64) (*WorkforceSummary, error)
	ListWorkforceTasks(ctx context.Context, orgID int64, filter WorkforceTaskFilter) ([]*WorkforceTaskItem, int64, error)
	CheckWorkforceHealth(ctx context.Context) WorkforceHealth
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

const selectTaskColumns = `
	id, org_id, document_id, entity_type, entity_id, task_type, resource_id,
	payload, status, error_message, last_error_code, retry_count, max_retries,
	worker_id, lease_expires_at, heartbeat_at, started_at, completed_at, available_at,
	thread_id, correlation_id, approval_id, acting_user_id, actor_type,
	created_at, updated_at
`

func (r *repository) GetTaskByID(ctx context.Context, orgID int64, taskID int64) (*Task, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_processing_tasks
		WHERE id = ? AND org_id = ?
		LIMIT 1
	`, selectTaskColumns)
	var task Task
	err := r.db.GetContext(ctx, &task, query, taskID, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get task #%d: %w", taskID, err)
	}
	return &task, nil
}

func (r *repository) GetTaskByIDUnscoped(ctx context.Context, taskID int64) (*Task, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_processing_tasks
		WHERE id = ?
		LIMIT 1
	`, selectTaskColumns)
	var task Task
	err := r.db.GetContext(ctx, &task, query, taskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get task #%d: %w", taskID, err)
	}
	return &task, nil
}

func (r *repository) ListTasks(ctx context.Context, orgID int64, filter TaskFilter) ([]*Task, int64, error) {
	where := "WHERE org_id = ?"
	args := []interface{}{orgID}

	if filter.Status != "" {
		where += " AND status = ?"
		args = append(args, filter.Status)
	}
	if filter.TaskType != "" {
		where += " AND task_type = ?"
		args = append(args, filter.TaskType)
	}
	if filter.EntityType != "" {
		where += " AND entity_type = ?"
		args = append(args, filter.EntityType)
	}
	if filter.EntityID != "" {
		where += " AND entity_id = ?"
		args = append(args, filter.EntityID)
	}
	if filter.ThreadID != "" {
		where += " AND thread_id = ?"
		args = append(args, filter.ThreadID)
	}

	var count int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ai_processing_tasks %s", where)
	if err := r.db.GetContext(ctx, &count, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM ai_processing_tasks
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, selectTaskColumns, where)
	args = append(args, limit, offset)

	var tasks []*Task
	if err := r.db.SelectContext(ctx, &tasks, query, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, count, nil
}

func (r *repository) CreateTask(ctx context.Context, task *Task) error {
	if task.MaxRetries <= 0 {
		task.MaxRetries = 3
	}
	if task.Status == "" {
		task.Status = StatusQueued
	}
	if task.ActorType == "" {
		task.ActorType = "AI_AGENT"
	}

	query := `
		INSERT INTO ai_processing_tasks (
			org_id, document_id, entity_type, entity_id, task_type, resource_id,
			payload, status, error_message, last_error_code, retry_count, max_retries,
			worker_id, lease_expires_at, heartbeat_at, started_at, completed_at, available_at,
			thread_id, correlation_id, approval_id, acting_user_id, actor_type,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, 0, ?,
			?, ?, ?, ?, ?, NOW(),
			?, ?, ?, ?, ?,
			NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(ctx, query,
		task.OrgID, task.DocumentID, task.EntityType, task.EntityID, task.TaskType, task.ResourceID,
		task.Payload, task.Status, task.ErrorMessage, task.LastErrorCode, task.MaxRetries,
		task.WorkerID, task.LeaseExpiresAt, task.HeartbeatAt, task.StartedAt, task.CompletedAt,
		task.ThreadID, task.CorrelationID, task.ApprovalID, task.ActingUserID, task.ActorType,
	)
	if err != nil {
		return fmt.Errorf("failed to insert ai_processing_task: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		task.ID = id
	}
	return nil
}

func (r *repository) ClaimNextTask(ctx context.Context, workerID string, leaseDuration time.Duration, taskTypes []string) (*Task, error) {
	if leaseDuration <= 0 {
		leaseDuration = 5 * time.Minute
	}
	leaseSeconds := int(leaseDuration.Seconds())

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("failed to begin claim transaction: %w", err)
	}
	defer tx.Rollback()

	whereClause := `
		WHERE (status = 'QUEUED' OR (status = 'RETRYING' AND (available_at IS NULL OR available_at <= NOW())))
		  AND (available_at IS NULL OR available_at <= NOW())
		  AND retry_count < max_retries
	`
	var args []interface{}

	if len(taskTypes) > 0 {
		queryTypes, typeArgs, err := sqlx.In(" AND task_type IN (?)", taskTypes)
		if err == nil {
			whereClause += queryTypes
			args = append(args, typeArgs...)
		}
	}

	selectQuery := fmt.Sprintf(`
		SELECT %s
		FROM ai_processing_tasks
		%s
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`, selectTaskColumns, whereClause)

	var task Task
	err = tx.GetContext(ctx, &task, selectQuery, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No eligible tasks in queue
		}
		return nil, fmt.Errorf("failed to select task for claim: %w", err)
	}

	updateQuery := `
		UPDATE ai_processing_tasks
		SET status = 'PROCESSING',
			worker_id = ?,
			started_at = NOW(),
			heartbeat_at = NOW(),
			lease_expires_at = DATE_ADD(NOW(), INTERVAL ? SECOND),
			retry_count = retry_count + 1,
			updated_at = NOW()
		WHERE id = ? AND (status = 'QUEUED' OR status = 'RETRYING')
	`
	res, err := tx.ExecContext(ctx, updateQuery, workerID, leaseSeconds, task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to claim task #%d: %w", task.ID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return nil, nil // Concurrently claimed or updated
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit claim transaction: %w", err)
	}

	// Refresh task state
	return r.GetTaskByIDUnscoped(ctx, task.ID)
}

func (r *repository) HeartbeatTask(ctx context.Context, taskID int64, workerID string, leaseDuration time.Duration) error {
	if leaseDuration <= 0 {
		leaseDuration = 5 * time.Minute
	}
	leaseSeconds := int(leaseDuration.Seconds())

	query := `
		UPDATE ai_processing_tasks
		SET heartbeat_at = NOW(),
			lease_expires_at = DATE_ADD(NOW(), INTERVAL ? SECOND),
			updated_at = NOW()
		WHERE id = ? AND status = 'PROCESSING' AND worker_id = ?
	`
	res, err := r.db.ExecContext(ctx, query, leaseSeconds, taskID, workerID)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat for task #%d: %w", taskID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return fmt.Errorf("task #%d not in PROCESSING state with worker %s: %w", taskID, workerID, ErrWorkerMismatch)
	}

	return nil
}

func (r *repository) UpdateTaskStatus(ctx context.Context, taskID int64, input UpdateTaskStatusInput) (*Task, error) {
	current, err := r.GetTaskByIDUnscoped(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Validate state transition centrally
	if err := ValidateTransition(current.Status, input.Status, false); err != nil {
		return nil, err
	}

	var completedAt *time.Time
	if IsTerminalStatus(input.Status) {
		now := time.Now()
		completedAt = &now
	}

	var availableAtClause string
	var args []interface{}

	setClauses := "status = ?, updated_at = NOW()"
	args = append(args, input.Status)

	if input.ErrorMessage != "" {
		setClauses += ", error_message = ?"
		args = append(args, input.ErrorMessage)
	}
	if input.LastErrorCode != "" {
		setClauses += ", last_error_code = ?"
		args = append(args, input.LastErrorCode)
	}
	if input.ThreadID != "" {
		setClauses += ", thread_id = ?"
		args = append(args, input.ThreadID)
	}
	if input.CorrelationID != "" {
		setClauses += ", correlation_id = ?"
		args = append(args, input.CorrelationID)
	}
	if input.ApprovalID != nil && *input.ApprovalID > 0 {
		setClauses += ", approval_id = ?"
		args = append(args, input.ApprovalID)
	}
	if completedAt != nil {
		setClauses += ", completed_at = NOW()"
	}
	if input.Status == StatusRetrying && input.DelaySeconds > 0 {
		availableAtClause = ", available_at = DATE_ADD(NOW(), INTERVAL ? SECOND)"
		setClauses += availableAtClause
		args = append(args, input.DelaySeconds)
	} else if input.Status == StatusQueued {
		setClauses += ", available_at = NOW(), worker_id = NULL, lease_expires_at = NULL"
	} else if IsTerminalStatus(input.Status) {
		setClauses += ", lease_expires_at = NULL"
	}

	query := fmt.Sprintf("UPDATE ai_processing_tasks SET %s WHERE id = ?", setClauses)
	args = append(args, taskID)

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update status for task #%d: %w", taskID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return nil, errors.New("task update affected 0 rows")
	}

	return r.GetTaskByIDUnscoped(ctx, taskID)
}

func (r *repository) CancelTask(ctx context.Context, orgID int64, taskID int64, reason string) (*Task, error) {
	current, err := r.GetTaskByID(ctx, orgID, taskID)
	if err != nil {
		return nil, err
	}

	if err := ValidateTransition(current.Status, StatusCancelled, false); err != nil {
		return nil, err
	}

	if reason == "" {
		reason = "Cancelled by user"
	}

	query := `
		UPDATE ai_processing_tasks
		SET status = 'CANCELLED',
			last_error_code = 'USER_CANCELLED',
			error_message = ?,
			completed_at = NOW(),
			lease_expires_at = NULL,
			updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err = r.db.ExecContext(ctx, query, reason, taskID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel task #%d: %w", taskID, err)
	}

	return r.GetTaskByID(ctx, orgID, taskID)
}

func (r *repository) RetryTask(ctx context.Context, orgID int64, taskID int64) (*Task, error) {
	current, err := r.GetTaskByID(ctx, orgID, taskID)
	if err != nil {
		return nil, err
	}

	// Validate explicit retry transition from FAILED to QUEUED
	if err := ValidateTransition(current.Status, StatusQueued, true); err != nil {
		return nil, err
	}

	query := `
		UPDATE ai_processing_tasks
		SET status = 'QUEUED',
			retry_count = 0,
			available_at = NOW(),
			worker_id = NULL,
			lease_expires_at = NULL,
			started_at = NULL,
			completed_at = NULL,
			error_message = CONCAT(COALESCE(error_message, ''), '\n[Manual Retry] Task explicitly reset to QUEUED.'),
			last_error_code = NULL,
			updated_at = NOW()
		WHERE id = ? AND org_id = ?
	`
	_, err = r.db.ExecContext(ctx, query, taskID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to retry task #%d: %w", taskID, err)
	}

	return r.GetTaskByID(ctx, orgID, taskID)
}

func (r *repository) RecoverStaleTasks(ctx context.Context, defaultLeaseDuration time.Duration) (int64, int64, error) {
	// 1. Recover stale tasks that have not exceeded max_retries back to QUEUED
	recoverQuery := `
		UPDATE ai_processing_tasks
		SET status = 'QUEUED',
			worker_id = NULL,
			lease_expires_at = NULL,
			available_at = NOW(),
			error_message = CONCAT(COALESCE(error_message, ''), '\n[Recovery] Lease expired; reset to QUEUED for retry.'),
			updated_at = NOW()
		WHERE status = 'PROCESSING'
		  AND lease_expires_at IS NOT NULL
		  AND lease_expires_at < NOW()
		  AND retry_count < max_retries
	`
	resRecover, err := r.db.ExecContext(ctx, recoverQuery)
	var recoveredCount int64
	if err == nil {
		recoveredCount, _ = resRecover.RowsAffected()
	}

	// 2. Mark stale tasks that have exceeded max_retries as FAILED
	failQuery := `
		UPDATE ai_processing_tasks
		SET status = 'FAILED',
			last_error_code = 'LEASE_EXPIRED',
			completed_at = NOW(),
			lease_expires_at = NULL,
			error_message = CONCAT(COALESCE(error_message, ''), '\n[Recovery] Task failed: lease expired and max retries exceeded.'),
			updated_at = NOW()
		WHERE status = 'PROCESSING'
		  AND lease_expires_at IS NOT NULL
		  AND lease_expires_at < NOW()
		  AND retry_count >= max_retries
	`
	resFail, err := r.db.ExecContext(ctx, failQuery)
	var failedCount int64
	if err == nil {
		failedCount, _ = resFail.RowsAffected()
	}

	return recoveredCount, failedCount, nil
}

func (r *repository) GetTaskStats(ctx context.Context, orgID int64) (*TaskStats, error) {
	stats := &TaskStats{}
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN status = 'QUEUED' THEN 1 ELSE 0 END), 0) AS queued,
			COALESCE(SUM(CASE WHEN status = 'PROCESSING' THEN 1 ELSE 0 END), 0) AS processing,
			COALESCE(SUM(CASE WHEN status = 'WAITING_FOR_APPROVAL' THEN 1 ELSE 0 END), 0) AS waiting_for_approval,
			COALESCE(SUM(CASE WHEN status = 'RETRYING' THEN 1 ELSE 0 END), 0) AS retrying,
			COALESCE(SUM(CASE WHEN status = 'COMPLETED' THEN 1 ELSE 0 END), 0) AS completed,
			COALESCE(SUM(CASE WHEN status = 'FAILED' THEN 1 ELSE 0 END), 0) AS failed,
			COALESCE(SUM(CASE WHEN status = 'REJECTED' THEN 1 ELSE 0 END), 0) AS rejected,
			COALESCE(SUM(CASE WHEN status = 'CANCELLED' THEN 1 ELSE 0 END), 0) AS cancelled,
			COUNT(*) AS total
		FROM ai_processing_tasks
		WHERE org_id = ?
	`
	err := r.db.GetContext(ctx, stats, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch task stats: %w", err)
	}
	return stats, nil
}

func (r *repository) CheckWorkforceHealth(ctx context.Context) WorkforceHealth {
	health := WorkforceHealth{
		OverallStatus:         "healthy",
		BackendStatus:         "healthy",
		SidecarStatus:         "healthy",
		WorkerStatus:          "healthy",
		QueueStatus:           "healthy",
		CheckpointStatus:      "healthy",
		PrimaryProviderStatus: "healthy",
		FailoverStatus:        "ready",
	}

	// 1. Check DB connectivity
	if err := r.db.PingContext(ctx); err != nil {
		health.BackendStatus = "degraded"
		health.OverallStatus = "unavailable"
	}

	// 2. Check Checkpoint storage table
	var dummy int
	if err := r.db.GetContext(ctx, &dummy, "SELECT 1 FROM ai_checkpoints LIMIT 1"); err != nil && !errors.Is(err, sql.ErrNoRows) {
		health.CheckpointStatus = "degraded"
	}

	// 3. Check Sidecar HTTP ping (fail fast with 1.5s timeout)
	client := http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:8090/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		health.SidecarStatus = "unavailable"
		health.OverallStatus = "degraded"
	} else {
		resp.Body.Close()
	}

	// 4. Check Queue lag & stale tasks
	var staleCount int64
	_ = r.db.GetContext(ctx, &staleCount, "SELECT COUNT(*) FROM ai_processing_tasks WHERE status = 'PROCESSING' AND lease_expires_at IS NOT NULL AND lease_expires_at < NOW()")
	if staleCount > 0 {
		health.QueueStatus = "lagging"
		health.WorkerStatus = "degraded"
	}

	return health
}

func (r *repository) GetWorkforceSummary(ctx context.Context, orgID int64) (*WorkforceSummary, error) {
	// 1. Aggregated task status counts
	type taskCounts struct {
		Total              int64      `db:"total"`
		Queued             int64      `db:"queued"`
		Processing         int64      `db:"processing"`
		Stale              int64      `db:"stale"`
		WaitingForApproval int64      `db:"waiting_for_approval"`
		Retrying           int64      `db:"retrying"`
		Failed             int64      `db:"failed"`
		Completed24h       int64      `db:"completed_24h"`
		LastHeartbeat      *time.Time `db:"last_heartbeat"`
	}

	var counts taskCounts
	countQuery := `
		SELECT
			COUNT(*) AS total,
			COALESCE(SUM(CASE WHEN status = 'QUEUED' THEN 1 ELSE 0 END), 0) AS queued,
			COALESCE(SUM(CASE WHEN status = 'PROCESSING' AND (lease_expires_at IS NULL OR lease_expires_at >= NOW()) THEN 1 ELSE 0 END), 0) AS processing,
			COALESCE(SUM(CASE WHEN status = 'PROCESSING' AND lease_expires_at IS NOT NULL AND lease_expires_at < NOW() THEN 1 ELSE 0 END), 0) AS stale,
			COALESCE(SUM(CASE WHEN status = 'WAITING_FOR_APPROVAL' THEN 1 ELSE 0 END), 0) AS waiting_for_approval,
			COALESCE(SUM(CASE WHEN status = 'RETRYING' THEN 1 ELSE 0 END), 0) AS retrying,
			COALESCE(SUM(CASE WHEN status = 'FAILED' THEN 1 ELSE 0 END), 0) AS failed,
			COALESCE(SUM(CASE WHEN status = 'COMPLETED' AND completed_at >= NOW() - INTERVAL 24 HOUR THEN 1 ELSE 0 END), 0) AS completed_24h,
			MAX(heartbeat_at) AS last_heartbeat
		FROM ai_processing_tasks
		WHERE org_id = ?
	`
	if err := r.db.GetContext(ctx, &counts, countQuery, orgID); err != nil {
		return nil, fmt.Errorf("failed to fetch task counts for org %d: %w", orgID, err)
	}

	// 2. Query execution traces in last 24h
	type traceCounts struct {
		FailoverCount int64   `db:"failover_count"`
		MockCount     int64   `db:"mock_count"`
		AvgDuration   float64 `db:"avg_duration"`
	}
	var traces traceCounts
	traceQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN failover_occurred = 1 THEN 1 ELSE 0 END), 0) AS failover_count,
			COALESCE(SUM(CASE WHEN is_mock = 1 THEN 1 ELSE 0 END), 0) AS mock_count,
			COALESCE(AVG(CASE WHEN status = 'COMPLETED' AND duration_ms IS NOT NULL THEN duration_ms ELSE NULL END), 0) AS avg_duration
		FROM ai_execution_traces
		WHERE org_id = ? AND created_at >= NOW() - INTERVAL 24 HOUR
	`
	_ = r.db.GetContext(ctx, &traces, traceQuery, orgID)

	// 3. Query agent breakdown
	type agentRow struct {
		TaskType         string `db:"task_type"`
		ActiveCount      int64  `db:"active_count"`
		Completed24h     int64  `db:"completed_24h"`
		Failed24h        int64  `db:"failed_24h"`
		WaitingApprovals int64  `db:"waiting_approvals"`
	}
	var agentRows []agentRow
	agentQuery := `
		SELECT
			task_type,
			COALESCE(SUM(CASE WHEN status IN ('QUEUED', 'PROCESSING', 'RETRYING', 'WAITING_FOR_APPROVAL') THEN 1 ELSE 0 END), 0) AS active_count,
			COALESCE(SUM(CASE WHEN status = 'COMPLETED' AND completed_at >= NOW() - INTERVAL 24 HOUR THEN 1 ELSE 0 END), 0) AS completed_24h,
			COALESCE(SUM(CASE WHEN status = 'FAILED' AND updated_at >= NOW() - INTERVAL 24 HOUR THEN 1 ELSE 0 END), 0) AS failed_24h,
			COALESCE(SUM(CASE WHEN status = 'WAITING_FOR_APPROVAL' THEN 1 ELSE 0 END), 0) AS waiting_approvals
		FROM ai_processing_tasks
		WHERE org_id = ?
		GROUP BY task_type
	`
	_ = r.db.SelectContext(ctx, &agentRows, agentQuery, orgID)

	agentsMap := make(map[string]AgentWorkforceStat)
	canonicalAgents := []struct {
		Key         string
		DisplayName string
		Module      string
	}{
		{"pricing", "Pricing Analyst", "PRICING"},
		{"sales", "Sales Email Parser", "SALES"},
		{"operations", "Operations Carrier Tracker", "OPERATIONS"},
		{"contracts", "Contracts Intelligence", "CONTRACTS"},
		{"compliance", "Compliance Auditor", "COMPLIANCE"},
		{"finance", "Finance Invoice Auditor", "FINANCE"},
		{"leads", "Lead Scoring Specialist", "LEADS"},
		{"outreach", "Outreach Copywriter", "OUTREACH"},
	}

	for _, a := range canonicalAgents {
		agentsMap[a.Key] = AgentWorkforceStat{
			AgentKey:         a.Key,
			DisplayName:      a.DisplayName,
			Module:           a.Module,
			ActiveTasks:      0,
			Completed24h:     0,
			Failed24h:        0,
			WaitingApprovals: 0,
			Status:           "IDLE",
		}
	}

	for _, row := range agentRows {
		key, name, mod, _ := MapTaskTypeToAgent(row.TaskType)
		stat := agentsMap[key]
		stat.ActiveTasks += row.ActiveCount
		stat.Completed24h += row.Completed24h
		stat.Failed24h += row.Failed24h
		stat.WaitingApprovals += row.WaitingApprovals
		if stat.ActiveTasks > 0 {
			stat.Status = "ACTIVE"
		} else if stat.Failed24h > 0 {
			stat.Status = "ATTENTION"
		} else {
			stat.Status = "IDLE"
		}
		stat.DisplayName = name
		stat.Module = mod
		agentsMap[key] = stat
	}

	activeAgentsCount := 0
	for _, s := range agentsMap {
		if s.Status == "ACTIVE" {
			activeAgentsCount++
		}
	}

	// 4. Query recent failures by safe category
	type failRow struct {
		Category string `db:"category"`
		Cnt      int64  `db:"cnt"`
	}
	var failRows []failRow
	failQuery := `
		SELECT
			COALESCE(error_category, 'unknown_error') AS category,
			COUNT(*) AS cnt
		FROM ai_execution_traces
		WHERE org_id = ? AND status = 'FAILED' AND created_at >= NOW() - INTERVAL 24 HOUR
		GROUP BY category
		ORDER BY cnt DESC
		LIMIT 5
	`
	_ = r.db.SelectContext(ctx, &failRows, failQuery, orgID)
	var failureStats []FailureCategoryStat
	for _, fr := range failRows {
		failureStats = append(failureStats, FailureCategoryStat{
			Category: fr.Category,
			Count:    fr.Cnt,
		})
	}

	// 5. System health check
	health := r.CheckWorkforceHealth(ctx)
	health.LastWorkerHeartbeat = counts.LastHeartbeat

	totalActive := counts.Queued + counts.Processing + counts.Retrying + counts.WaitingForApproval

	return &WorkforceSummary{
		OrgID:                   orgID,
		TotalActiveTasks:        totalActive,
		QueuedTasks:             counts.Queued,
		ProcessingTasks:         counts.Processing,
		WaitingForApprovalTasks: counts.WaitingForApproval,
		RetryingTasks:           counts.Retrying,
		FailedTasks:             counts.Failed,
		StaleTasks:              counts.Stale,
		CompletedRecent24h:      counts.Completed24h,
		CompletedWithFailover:   traces.FailoverCount,
		CompletedInMockMode:     traces.MockCount,
		AvgDurationMs:           int64(traces.AvgDuration),
		ActiveAgentsCount:       activeAgentsCount,
		ByAgent:                 agentsMap,
		RecentFailures:          failureStats,
		Health:                  health,
		LastUpdated:             time.Now(),
	}, nil
}

func (r *repository) ListWorkforceTasks(ctx context.Context, orgID int64, filter WorkforceTaskFilter) ([]*WorkforceTaskItem, int64, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	var conditions []string
	var args []interface{}

	conditions = append(conditions, "org_id = ?")
	args = append(args, orgID)

	if filter.Status != "" {
		st := strings.ToLower(strings.TrimSpace(filter.Status))
		switch st {
		case "stale":
			conditions = append(conditions, "status = 'PROCESSING' AND lease_expires_at IS NOT NULL AND lease_expires_at < NOW()")
		case "processing":
			conditions = append(conditions, "status = 'PROCESSING' AND (lease_expires_at IS NULL OR lease_expires_at >= NOW())")
		case "claimed":
			conditions = append(conditions, "status = 'QUEUED' AND worker_id IS NOT NULL AND worker_id != ''")
		case "queued":
			conditions = append(conditions, "status = 'QUEUED' AND (worker_id IS NULL OR worker_id = '')")
		default:
			conditions = append(conditions, "status = ?")
			args = append(args, strings.ToUpper(filter.Status))
		}
	}

	if filter.RequiresApproval != nil && *filter.RequiresApproval {
		conditions = append(conditions, "status = 'WAITING_FOR_APPROVAL'")
	}

	if filter.FailedOrStale != nil && *filter.FailedOrStale {
		conditions = append(conditions, "(status = 'FAILED' OR (status = 'PROCESSING' AND lease_expires_at IS NOT NULL AND lease_expires_at < NOW()))")
	}

	if filter.AgentKey != "" {
		switch strings.ToLower(filter.AgentKey) {
		case "pricing":
			conditions = append(conditions, "task_type LIKE '%PRICING%'")
		case "sales":
			conditions = append(conditions, "(task_type LIKE '%EMAIL%' OR task_type LIKE '%SALES%')")
		case "operations":
			conditions = append(conditions, "(task_type LIKE '%CARRIER%' OR task_type LIKE '%OPERATIONS%' OR task_type LIKE '%TRACKING%')")
		case "contracts":
			conditions = append(conditions, "(task_type LIKE '%CONTRACT%' OR task_type LIKE '%DOC_PROCESS%' OR task_type LIKE '%RATE_EXTRACT%')")
		case "compliance":
			conditions = append(conditions, "(task_type LIKE '%COMPLIANCE%' OR task_type LIKE '%DOC_VERIFY%')")
		case "finance":
			conditions = append(conditions, "(task_type LIKE '%FINANCE%' OR task_type LIKE '%BILL%' OR task_type LIKE '%INVOICE%')")
		case "leads":
			conditions = append(conditions, "task_type LIKE '%LEAD%'")
		case "outreach":
			conditions = append(conditions, "task_type LIKE '%OUTREACH%'")
		}
	}

	if filter.Module != "" {
		switch strings.ToUpper(strings.TrimSpace(filter.Module)) {
		case "PRICING", "RFQ", "RFQS":
			conditions = append(conditions, "task_type LIKE '%PRICING%'")
		case "SALES":
			conditions = append(conditions, "(task_type LIKE '%EMAIL%' OR task_type LIKE '%SALES%')")
		case "OPERATIONS", "SHIPMENTS", "TRACKING":
			conditions = append(conditions, "(task_type LIKE '%CARRIER%' OR task_type LIKE '%OPERATIONS%' OR task_type LIKE '%TRACKING%')")
		case "CONTRACTS":
			conditions = append(conditions, "(task_type LIKE '%CONTRACT%' OR task_type LIKE '%DOC_PROCESS%' OR task_type LIKE '%RATE_EXTRACT%')")
		case "COMPLIANCE", "DOCUMENTS":
			conditions = append(conditions, "(task_type LIKE '%COMPLIANCE%' OR task_type LIKE '%DOC_VERIFY%')")
		case "FINANCE", "INVOICES", "BILLING":
			conditions = append(conditions, "(task_type LIKE '%FINANCE%' OR task_type LIKE '%BILL%' OR task_type LIKE '%INVOICE%')")
		case "LEADS":
			conditions = append(conditions, "task_type LIKE '%LEAD%'")
		case "OUTREACH":
			conditions = append(conditions, "task_type LIKE '%OUTREACH%'")
		default:
			conditions = append(conditions, "(task_type LIKE ? OR entity_type = ?)")
			mVal := "%" + filter.Module + "%"
			args = append(args, mVal, strings.ToUpper(filter.Module))
		}
	}

	if filter.Search != "" {
		term := "%" + filter.Search + "%"
		conditions = append(conditions, "(entity_id LIKE ? OR task_type LIKE ? OR thread_id LIKE ? OR correlation_id LIKE ?)")
		args = append(args, term, term, term, term)
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ai_processing_tasks %s", where)
	var count int64
	if err := r.db.GetContext(ctx, &count, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count workforce tasks: %w", err)
	}

	// Select query
	selectQuery := fmt.Sprintf(`
		SELECT %s
		FROM ai_processing_tasks
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, selectTaskColumns, where)
	args = append(args, limit, offset)

	var tasks []*Task
	if err := r.db.SelectContext(ctx, &tasks, selectQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to fetch workforce tasks: %w", err)
	}

	// Convert to WorkforceTaskItem
	var items []*WorkforceTaskItem
	now := time.Now()
	for _, t := range tasks {
		agentKey, agentName, module, _ := MapTaskTypeToAgent(t.TaskType)
		relLabel, relID, navURL := ResolveRelatedRef(t)

		isStale := t.Status == StatusProcessing && t.LeaseExpiresAt != nil && t.LeaseExpiresAt.Before(now)
		stVal, stMeta := ResolveTaskWorkforceStatus(t, false, false)

		var durMs *int64
		if t.StartedAt != nil && t.CompletedAt != nil {
			ms := t.CompletedAt.Sub(*t.StartedAt).Milliseconds()
			durMs = &ms
		}

		var sanitizedErr *string
		if t.ErrorMessage != nil && *t.ErrorMessage != "" {
			redacted := ai.RedactSecrets(*t.ErrorMessage)
			sanitizedErr = &redacted
		}

		canRetry := (t.Status == StatusFailed || isStale)
		canCancel := (t.Status == StatusQueued || t.Status == StatusProcessing || t.Status == StatusRetrying)

		items = append(items, &WorkforceTaskItem{
			ID:            t.ID,
			OrgID:         t.OrgID,
			TaskType:      t.TaskType,
			AgentKey:      agentKey,
			AgentName:     agentName,
			Module:        module,
			RelatedRef:    relLabel,
			RelatedID:     relID,
			Status:        stVal,
			StatusMeta:    stMeta,
			CreatedAt:     t.CreatedAt,
			UpdatedAt:     t.UpdatedAt,
			StartedAt:     t.StartedAt,
			CompletedAt:   t.CompletedAt,
			DurationMs:    durMs,
			RetryCount:    t.RetryCount,
			MaxRetries:    t.MaxRetries,
			IsStale:       isStale,
			ApprovalID:    t.ApprovalID,
			ErrorCategory: t.LastErrorCode,
			ErrorMessage:  sanitizedErr,
			CanRetry:      canRetry,
			CanCancel:     canCancel,
			NavigationURL: navURL,
			ThreadID:      t.ThreadID,
			CorrelationID: t.CorrelationID,
		})
	}

	return items, count, nil
}
