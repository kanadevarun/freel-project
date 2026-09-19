package aitasks

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/freel/backend/internal/audit"
	auditDomain "github.com/freel/backend/internal/audit/domain"
)

type ApprovalsDelegate func(ctx context.Context, orgID int64, approvalID int64, actorName string, reason string) error

type Service interface {
	SetApprovalsDelegate(delegate ApprovalsDelegate)
	GetTaskByID(ctx context.Context, orgID int64, taskID int64) (*Task, error)
	GetTaskByIDUnscoped(ctx context.Context, taskID int64) (*Task, error)
	ListTasks(ctx context.Context, orgID int64, filter TaskFilter) ([]*Task, int64, error)
	CreateTask(ctx context.Context, input *CreateTaskInput) (*Task, error)
	ClaimNextTask(ctx context.Context, workerID string, leaseDuration time.Duration, taskTypes []string) (*Task, error)
	HeartbeatTask(ctx context.Context, taskID int64, workerID string, leaseDuration time.Duration) error
	UpdateTaskStatus(ctx context.Context, orgID int64, taskID int64, input UpdateTaskStatusInput) (*Task, error)
	CancelTask(ctx context.Context, orgID int64, taskID int64, actorName string, reason string) (*Task, error)
	RetryTask(ctx context.Context, orgID int64, taskID int64, actorName string) (*Task, error)
	RecoverStaleTasks(ctx context.Context, leaseDuration time.Duration) (int64, int64, error)
	GetTaskStats(ctx context.Context, orgID int64) (*TaskStats, error)
	GetWorkforceSummary(ctx context.Context, orgID int64) (*WorkforceSummary, error)
	ListWorkforceTasks(ctx context.Context, orgID int64, filter WorkforceTaskFilter) ([]*WorkforceTaskItem, int64, error)
	GetWorkforceHealth(ctx context.Context) (*WorkforceHealth, error)
}

type service struct {
	repo              Repository
	approvalsDelegate ApprovalsDelegate
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) SetApprovalsDelegate(delegate ApprovalsDelegate) {
	s.approvalsDelegate = delegate
}

func (s *service) GetTaskByID(ctx context.Context, orgID int64, taskID int64) (*Task, error) {
	return s.repo.GetTaskByID(ctx, orgID, taskID)
}

func (s *service) GetTaskByIDUnscoped(ctx context.Context, taskID int64) (*Task, error) {
	return s.repo.GetTaskByIDUnscoped(ctx, taskID)
}

func (s *service) ListTasks(ctx context.Context, orgID int64, filter TaskFilter) ([]*Task, int64, error) {
	return s.repo.ListTasks(ctx, orgID, filter)
}

func (s *service) CreateTask(ctx context.Context, input *CreateTaskInput) (*Task, error) {
	if input.OrgID <= 0 {
		return nil, fmt.Errorf("org_id is required")
	}
	if input.TaskType == "" {
		return nil, fmt.Errorf("task_type is required")
	}

	payloadJSON := "{}"
	if input.Payload != nil {
		bytes, err := json.Marshal(input.Payload)
		if err == nil {
			payloadJSON = string(bytes)
		}
	}

	maxRetries := input.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	task := &Task{
		OrgID:      input.OrgID,
		TaskType:   input.TaskType,
		Payload:    &payloadJSON,
		Status:     StatusQueued,
		MaxRetries: maxRetries,
		ActorType:  input.ActorType,
	}

	if input.EntityType != "" {
		task.EntityType = &input.EntityType
	}
	if input.EntityID != "" {
		task.EntityID = &input.EntityID
	}
	if input.DocumentID != "" {
		task.DocumentID = &input.DocumentID
	}
	if input.ResourceID != nil && *input.ResourceID > 0 {
		task.ResourceID = input.ResourceID
	}
	if input.ThreadID != "" {
		task.ThreadID = &input.ThreadID
	}
	if input.CorrelationID != "" {
		task.CorrelationID = &input.CorrelationID
	}
	if input.ActingUserID != nil && *input.ActingUserID > 0 {
		task.ActingUserID = input.ActingUserID
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	// Record audit event
	_, _ = audit.Record(ctx, auditDomain.CreateAuditLogParams{
		OrgID:        input.OrgID,
		ActorType:    "AI_AGENT",
		ActorName:    "QueueDispatcher",
		Action:       "ENQUEUE_TASK",
		Module:       "AI_TASKS",
		ResourceType: "ai_processing_task",
		ResourceID:   strconv.FormatInt(task.ID, 10),
		Result:       "QUEUED",
		Description:  fmt.Sprintf("Enqueued AI task #%d (type: %s) for entity %s:%s", task.ID, input.TaskType, input.EntityType, input.EntityID),
	})

	return task, nil
}

func (s *service) ClaimNextTask(ctx context.Context, workerID string, leaseDuration time.Duration, taskTypes []string) (*Task, error) {
	if workerID == "" {
		return nil, fmt.Errorf("worker_id is required to claim a task")
	}
	return s.repo.ClaimNextTask(ctx, workerID, leaseDuration, taskTypes)
}

func (s *service) HeartbeatTask(ctx context.Context, taskID int64, workerID string, leaseDuration time.Duration) error {
	if workerID == "" {
		return fmt.Errorf("worker_id is required for heartbeat")
	}
	return s.repo.HeartbeatTask(ctx, taskID, workerID, leaseDuration)
}

func (s *service) UpdateTaskStatus(ctx context.Context, orgID int64, taskID int64, input UpdateTaskStatusInput) (*Task, error) {
	current, err := s.repo.GetTaskByIDUnscoped(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Organization boundary check: ensure task belongs to caller org if orgID specified
	if orgID > 0 && current.OrgID != orgID {
		return nil, fmt.Errorf("%w: task #%d belongs to organization #%d", ErrTenantMismatch, taskID, current.OrgID)
	}

	// Prevent overwriting terminal state with non-terminal status
	if IsTerminalStatus(current.Status) && !IsTerminalStatus(input.Status) {
		return nil, fmt.Errorf("%w: cannot transition from terminal status '%s' to '%s'", ErrTerminalState, current.Status, input.Status)
	}

	updated, err := s.repo.UpdateTaskStatus(ctx, taskID, input)
	if err != nil {
		return nil, err
	}

	// Log audit event for significant transitions
	if input.Status == StatusCompleted || input.Status == StatusFailed || input.Status == StatusWaitingForApproval {
		_, _ = audit.Record(ctx, auditDomain.CreateAuditLogParams{
			OrgID:        current.OrgID,
			ActorType:    "AI_AGENT",
			ActorName:    input.WorkerID,
			Action:       fmt.Sprintf("TASK_%s", input.Status),
			Module:       "AI_TASKS",
			ResourceType: "ai_processing_task",
			ResourceID:   strconv.FormatInt(taskID, 10),
			Result:       input.Status,
			Description:  fmt.Sprintf("AI task #%d transitioned to %s", taskID, input.Status),
		})
	}

	return updated, nil
}

func (s *service) CancelTask(ctx context.Context, orgID int64, taskID int64, actorName string, reason string) (*Task, error) {
	current, err := s.repo.GetTaskByID(ctx, orgID, taskID)
	if err != nil {
		return nil, err
	}

	if IsTerminalStatus(current.Status) {
		return nil, fmt.Errorf("%w: task #%d is already in terminal state '%s'", ErrTerminalState, taskID, current.Status)
	}

	updated, err := s.repo.CancelTask(ctx, orgID, taskID, reason)
	if err != nil {
		return nil, err
	}

	// If linked to an active approval request, cancel the approval too
	if current.ApprovalID != nil && *current.ApprovalID > 0 && s.approvalsDelegate != nil {
		_ = s.approvalsDelegate(ctx, orgID, *current.ApprovalID, actorName, fmt.Sprintf("AI task #%d cancelled: %s", taskID, reason))
	}

	_, _ = audit.Record(ctx, auditDomain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorType:    "USER",
		ActorName:    actorName,
		Action:       "CANCEL_TASK",
		Module:       "AI_TASKS",
		ResourceType: "ai_processing_task",
		ResourceID:   strconv.FormatInt(taskID, 10),
		Result:       "CANCELLED",
		Description:  fmt.Sprintf("User cancelled AI task #%d: %s", taskID, reason),
	})

	return updated, nil
}

func (s *service) RetryTask(ctx context.Context, orgID int64, taskID int64, actorName string) (*Task, error) {
	current, err := s.repo.GetTaskByID(ctx, orgID, taskID)
	if err != nil {
		return nil, err
	}

	if current.Status != StatusFailed {
		return nil, fmt.Errorf("only FAILED tasks can be retried (current status: %s)", current.Status)
	}

	updated, err := s.repo.RetryTask(ctx, orgID, taskID)
	if err != nil {
		return nil, err
	}

	_, _ = audit.Record(ctx, auditDomain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorType:    "USER",
		ActorName:    actorName,
		Action:       "RETRY_TASK",
		Module:       "AI_TASKS",
		ResourceType: "ai_processing_task",
		ResourceID:   strconv.FormatInt(taskID, 10),
		Result:       "QUEUED",
		Description:  fmt.Sprintf("User explicitly retried failed AI task #%d", taskID),
	})

	return updated, nil
}

func (s *service) RecoverStaleTasks(ctx context.Context, leaseDuration time.Duration) (int64, int64, error) {
	rec, fail, err := s.repo.RecoverStaleTasks(ctx, leaseDuration)
	if err == nil && (rec > 0 || fail > 0) {
		_, _ = audit.Record(ctx, auditDomain.CreateAuditLogParams{
			OrgID:        1, // System-level event
			ActorType:    "SYSTEM",
			ActorName:    "StaleTaskRecovery",
			Action:       "RECOVER_STALE_TASKS",
			Module:       "AI_TASKS",
			Result:       "RECOVERED",
			Description:  fmt.Sprintf("Recovered %d stale tasks to QUEUED, marked %d expired tasks as FAILED", rec, fail),
		})
	}
	return rec, fail, err
}

func (s *service) GetTaskStats(ctx context.Context, orgID int64) (*TaskStats, error) {
	return s.repo.GetTaskStats(ctx, orgID)
}

func (s *service) GetWorkforceSummary(ctx context.Context, orgID int64) (*WorkforceSummary, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization context")
	}
	return s.repo.GetWorkforceSummary(ctx, orgID)
}

func (s *service) ListWorkforceTasks(ctx context.Context, orgID int64, filter WorkforceTaskFilter) ([]*WorkforceTaskItem, int64, error) {
	if orgID <= 0 {
		return nil, 0, fmt.Errorf("invalid organization context")
	}
	return s.repo.ListWorkforceTasks(ctx, orgID, filter)
}

func (s *service) GetWorkforceHealth(ctx context.Context) (*WorkforceHealth, error) {
	health := s.repo.CheckWorkforceHealth(ctx)
	return &health, nil
}
