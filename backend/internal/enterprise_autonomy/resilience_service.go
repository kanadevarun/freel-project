package enterprise_autonomy

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/approvals"
	"github.com/freel/backend/internal/audit/domain"
	auditSvc "github.com/freel/backend/internal/audit/service"
	"github.com/jmoiron/sqlx"
)

// EnterpriseResilienceService defines the operational resilience and health interface
type EnterpriseResilienceService interface {
	GetPlatformHealth(ctx context.Context, orgID int64) (*EnterprisePlatformHealthSummary, error)
	ClassifyFailure(err error, responsePayload map[string]interface{}) FailureClassification
	ShouldRetry(classification FailureClassification, currentAttempt int, maxAttempts int) (bool, time.Duration)
	SaveCheckpoint(ctx context.Context, orgID int64, checkpoint WorkflowCheckpoint) error
	GetLatestCheckpoint(ctx context.Context, orgID int64, workflowID string) (*WorkflowCheckpoint, error)
	DetectStuckWorkflows(ctx context.Context, orgID int64, inactivityThreshold time.Duration) ([]StuckWorkflowDescriptor, error)
	RecoverStuckWorkflow(ctx context.Context, orgID int64, workflowID string, actor string) error
	GetFailedWorkItems(ctx context.Context, orgID int64, limit int) ([]FailedWorkItem, error)
	ReplayFailedWorkItem(ctx context.Context, orgID int64, itemID string, actor string) error
	ReconcileMultiAgentExecution(ctx context.Context, workflowID string, originalConfidence float64, results []MultiAgentTaskResult) (*MultiAgentReconciliationResult, error)
	CheckEventStormProtection(ctx context.Context, orgID int64, entityType, entityID, eventType string) (bool, string)
	GetBackpressureMetrics(ctx context.Context, orgID int64) (*EventBackpressureMetrics, error)
	RecordHeartbeat(ctx context.Context, subsystem SubsystemIdentifier, state EnterpriseHealthState, latencyMs int64, msg string)
	CalculateDegradedAutonomyCeiling(overallState EnterpriseHealthState) AutonomyLevel
	ExecuteWithNotificationSeparation(ctx context.Context, orgID int64, workflowID string, businessAction func() error, notificationAction func() error) (bool, bool, error)
}

type defaultEnterpriseResilienceService struct {
	repo         Repository
	actionsSvc   actions.Service
	approvalsSvc approvals.Service
	auditSvc     auditSvc.Service
	db           *sqlx.DB

	mu                 sync.RWMutex
	subsystems         map[SubsystemIdentifier]*SubsystemHealth
	checkpoints        map[string]*WorkflowCheckpoint // Key: workflowID
	failedWorkItems    map[string]*FailedWorkItem     // Key: itemID
	eventArrivalWindow map[string][]time.Time         // Key: orgID:entityType:entityID -> timestamps
	coalescedCounts    map[int64]int64
	throttledCounts    map[int64]int64
	retryPolicy        RetryPolicy
}

// NewEnterpriseResilienceService creates the hardened operational resilience service
func NewEnterpriseResilienceService(
	repo Repository,
	actSvc actions.Service,
	apprSvc approvals.Service,
	audSvc auditSvc.Service,
	db *sqlx.DB,
) EnterpriseResilienceService {
	s := &defaultEnterpriseResilienceService{
		repo:               repo,
		actionsSvc:         actSvc,
		approvalsSvc:       apprSvc,
		auditSvc:           audSvc,
		db:                 db,
		subsystems:         make(map[SubsystemIdentifier]*SubsystemHealth),
		checkpoints:        make(map[string]*WorkflowCheckpoint),
		failedWorkItems:    make(map[string]*FailedWorkItem),
		eventArrivalWindow: make(map[string][]time.Time),
		coalescedCounts:    make(map[int64]int64),
		throttledCounts:    make(map[int64]int64),
		retryPolicy:        DefaultBoundedRetryPolicy(),
	}

	// Initialize baseline subsystem statuses
	now := time.Now().UTC()
	subsystems := []SubsystemIdentifier{
		SubsystemGoBackend,
		SubsystemPythonSidecar,
		SubsystemAIWorker,
		SubsystemEventMesh,
		SubsystemWorkflowEngine,
		SubsystemAIAgents,
		SubsystemActionSystem,
		SubsystemApprovals,
		SubsystemDatabase,
	}

	for _, sub := range subsystems {
		s.subsystems[sub] = &SubsystemHealth{
			Subsystem:       sub,
			State:           HealthStateHealthy,
			LatencyMs:       5,
			ActiveLoad:      0,
			CapacityLimit:   100,
			ErrorCount:      0,
			Message:         "Operational",
			LastHeartbeatAt: now,
			IsCritical:      sub == SubsystemGoBackend || sub == SubsystemDatabase || sub == SubsystemWorkflowEngine,
		}
	}

	return s
}

// ---------------------------------------------------------------------
// 1. Enterprise Health Model & Adaptive Degradation
// ---------------------------------------------------------------------

func (s *defaultEnterpriseResilienceService) RecordHeartbeat(
	ctx context.Context,
	subsystem SubsystemIdentifier,
	state EnterpriseHealthState,
	latencyMs int64,
	msg string,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.subsystems[subsystem]
	if !exists {
		item = &SubsystemHealth{
			Subsystem:  subsystem,
			IsCritical: subsystem == SubsystemGoBackend || subsystem == SubsystemDatabase || subsystem == SubsystemWorkflowEngine,
		}
		s.subsystems[subsystem] = item
	}

	item.State = state
	item.LatencyMs = latencyMs
	item.Message = msg
	item.LastHeartbeatAt = time.Now().UTC()
	if state == HealthStateFailed || state == HealthStateDegraded {
		item.ErrorCount++
	}
}

func (s *defaultEnterpriseResilienceService) GetPlatformHealth(ctx context.Context, orgID int64) (*EnterprisePlatformHealthSummary, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()

	// Live database ping check
	if s.db != nil {
		dbStart := time.Now()
		if err := s.db.PingContext(ctx); err != nil {
			s.subsystems[SubsystemDatabase].State = HealthStateFailed
			s.subsystems[SubsystemDatabase].Message = fmt.Sprintf("DB ping failed: %v", err)
		} else {
			s.subsystems[SubsystemDatabase].State = HealthStateHealthy
			s.subsystems[SubsystemDatabase].LatencyMs = time.Since(dbStart).Milliseconds()
			s.subsystems[SubsystemDatabase].Message = "Connected to MariaDB"
			s.subsystems[SubsystemDatabase].LastHeartbeatAt = now
		}
	}

	// Scan stuck workflows for this org
	stuckCount := 0
	if s.db != nil && orgID > 0 {
		var cnt int
		err := s.db.GetContext(ctx, &cnt, `
			SELECT COUNT(*) 
			FROM autonomous_plans 
			WHERE org_id = ? 
			  AND status IN ('RUNNING', 'PENDING', 'WAITING')
			  AND updated_at < DATE_SUB(NOW(), INTERVAL 15 MINUTE)
		`, orgID)
		if err == nil {
			stuckCount = cnt
		}
	}

	// Count failed work items
	deadLetterCount := 0
	for _, it := range s.failedWorkItems {
		if it.OrgID == orgID {
			deadLetterCount++
		}
	}

	// Evaluate overall platform state deterministically
	overallState := HealthStateHealthy
	degradationReason := ""

	// Check if any critical subsystem is FAILED
	for _, sub := range s.subsystems {
		if sub.IsCritical && sub.State == HealthStateFailed {
			overallState = HealthStateFailed
			degradationReason = fmt.Sprintf("Critical subsystem %s is in FAILED state", sub.Subsystem)
			break
		}
	}

	// If not failed, check if any subsystem is BLOCKED or PAUSED
	if overallState == HealthStateHealthy {
		for _, sub := range s.subsystems {
			if sub.State == HealthStateBlocked {
				overallState = HealthStateBlocked
				degradationReason = fmt.Sprintf("Subsystem %s is BLOCKED", sub.Subsystem)
				break
			}
			if sub.State == HealthStatePaused {
				overallState = HealthStatePaused
				degradationReason = fmt.Sprintf("Subsystem %s is PAUSED", sub.Subsystem)
				break
			}
		}
	}

	// If not blocked or paused, check if any subsystem is DEGRADED or stuck workflows exist
	if overallState == HealthStateHealthy {
		for _, sub := range s.subsystems {
			if sub.State == HealthStateDegraded {
				overallState = HealthStateDegraded
				degradationReason = fmt.Sprintf("Subsystem %s is DEGRADED: %s", sub.Subsystem, sub.Message)
				break
			}
		}
		if stuckCount > 5 {
			overallState = HealthStateDegraded
			degradationReason = fmt.Sprintf("%d workflows currently stuck without progress", stuckCount)
		}
	}

	// Calculate degraded autonomy ceiling governed by Go
	effectiveCeil := s.calculateDegradedAutonomyCeilingUnsafe(overallState)

	// Copy subsystem map
	subsystemsCopy := make(map[SubsystemIdentifier]SubsystemHealth)
	for k, v := range s.subsystems {
		subsystemsCopy[k] = *v
	}

	summary := &EnterprisePlatformHealthSummary{
		OrgID:                 orgID,
		OverallState:          overallState,
		EffectiveAutonomyCeil: effectiveCeil,
		Subsystems:            subsystemsCopy,
		StuckWorkflowCount:    stuckCount,
		DeadLetterCount:       deadLetterCount,
		EventBacklogDepth:     len(s.failedWorkItems),
		BackpressureActive:    overallState == HealthStateDegraded && stuckCount > 10,
		DegradationReason:     degradationReason,
		CheckedAt:             now,
	}

	return summary, nil
}

func (s *defaultEnterpriseResilienceService) CalculateDegradedAutonomyCeiling(overallState EnterpriseHealthState) AutonomyLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.calculateDegradedAutonomyCeilingUnsafe(overallState)
}

func (s *defaultEnterpriseResilienceService) calculateDegradedAutonomyCeilingUnsafe(overallState EnterpriseHealthState) AutonomyLevel {
	switch overallState {
	case HealthStateHealthy:
		return AutonomyLvl4Governed
	case HealthStateDegraded:
		return AutonomyLvl3Controlled
	case HealthStateBlocked:
		return AutonomyLvl2Prepare
	case HealthStateFailed:
		return AutonomyLvl1Recommend
	case HealthStatePaused:
		return AutonomyLvl0Observe
	default:
		return AutonomyLvl1Recommend
	}
}

// ---------------------------------------------------------------------
// 2. Retry Governance & Failure Classification
// ---------------------------------------------------------------------

func (s *defaultEnterpriseResilienceService) ClassifyFailure(err error, responsePayload map[string]interface{}) FailureClassification {
	if err == nil {
		return ""
	}

	msg := strings.ToLower(err.Error())

	// 1. Policy & Governance Blocked (FAIL-CLOSED, NEVER RETRY)
	if strings.Contains(msg, "governance") || strings.Contains(msg, "policy") ||
		strings.Contains(msg, "denied") || strings.Contains(msg, "restricted") ||
		strings.Contains(msg, "halt") || strings.Contains(msg, "circuit breaker") ||
		strings.Contains(msg, "loop detected") {
		return FailureTypePolicyBlocked
	}

	// 2. Authorization & Tenant Isolation (NEVER RETRY)
	if strings.Contains(msg, "unauthorized") || strings.Contains(msg, "forbidden") ||
		strings.Contains(msg, "privilege") || strings.Contains(msg, "401") ||
		strings.Contains(msg, "403") || strings.Contains(msg, "tenant") {
		return FailureTypeAuthorization
	}

	// 3. Stale Approvals (NEVER RETRY)
	if strings.Contains(msg, "stale approval") || strings.Contains(msg, "approval invalidated") ||
		strings.Contains(msg, "fingerprint mismatch") {
		return FailureTypeStaleApproval
	}

	// 4. Business Conflicts (NEVER RETRY)
	if strings.Contains(msg, "already paid") || strings.Contains(msg, "already delivered") ||
		strings.Contains(msg, "conflict") || strings.Contains(msg, "precondition failed") ||
		strings.Contains(msg, "rejected by human") {
		return FailureTypeBusinessConflict
	}

	// 5. Data Quality Failures (NEVER RETRY)
	if strings.Contains(msg, "missing required") || strings.Contains(msg, "invalid format") ||
		strings.Contains(msg, "malformed") || strings.Contains(msg, "validation failed") ||
		strings.Contains(msg, "cannot unmarshal") {
		return FailureTypeDataQuality
	}

	// 6. Timeouts (RETRYABLE TRANSIENT)
	if strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "context deadline") {
		return FailureTypeTimeout
	}

	// 7. External Dependency (RETRYABLE)
	if strings.Contains(msg, "carrier") || strings.Contains(msg, "third party") ||
		strings.Contains(msg, "external service") {
		return FailureTypeDependency
	}

	// 8. AI Model / LLM Provider (RETRYABLE)
	if strings.Contains(msg, "llm") || strings.Contains(msg, "model provider") ||
		strings.Contains(msg, "token limit") || strings.Contains(msg, "hallucination") {
		return FailureTypeAIModel
	}

	// 9. Network / Transient (RETRYABLE)
	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "network reset") ||
		strings.Contains(msg, "503") || strings.Contains(msg, "429") ||
		strings.Contains(msg, "rate limit") || strings.Contains(msg, "deadlock") ||
		strings.Contains(msg, "lock wait timeout") {
		return FailureTypeTransient
	}

	// Default to Permanent if unclassified
	return FailureTypePermanent
}

func (s *defaultEnterpriseResilienceService) ShouldRetry(
	classification FailureClassification,
	currentAttempt int,
	maxAttempts int,
) (bool, time.Duration) {
	if maxAttempts <= 0 {
		maxAttempts = s.retryPolicy.MaxAttempts
	}

	// Never retry permanent or policy-governed failures
	if currentAttempt >= maxAttempts ||
		classification == FailureTypePermanent ||
		classification == FailureTypePolicyBlocked ||
		classification == FailureTypeAuthorization ||
		classification == FailureTypeStaleApproval ||
		classification == FailureTypeDataQuality ||
		classification == FailureTypeBusinessConflict {
		return false, 0
	}

	// Only retry transient, timeout, or external dependency failures
	if classification != FailureTypeTransient &&
		classification != FailureTypeTimeout &&
		classification != FailureTypeDependency &&
		classification != FailureTypeAIModel {
		return false, 0
	}

	// Bounded exponential backoff with jitter
	backoffFactor := math.Pow(s.retryPolicy.BackoffFactor, float64(currentAttempt-1))
	baseDelay := float64(s.retryPolicy.InitialBackoff) * backoffFactor

	if baseDelay > float64(s.retryPolicy.MaxBackoff) {
		baseDelay = float64(s.retryPolicy.MaxBackoff)
	}

	// Add random jitter fraction
	jitter := (rand.Float64()*2 - 1) * s.retryPolicy.JitterFraction * baseDelay
	finalDelay := time.Duration(baseDelay + jitter)
	if finalDelay < s.retryPolicy.InitialBackoff {
		finalDelay = s.retryPolicy.InitialBackoff
	}

	return true, finalDelay
}

// ---------------------------------------------------------------------
// 3. Durable Checkpoint Management
// ---------------------------------------------------------------------

func (s *defaultEnterpriseResilienceService) SaveCheckpoint(
	ctx context.Context,
	orgID int64,
	checkpoint WorkflowCheckpoint,
) error {
	if orgID <= 0 || checkpoint.WorkflowID == "" {
		return ErrUnauthorizedTenant
	}

	checkpoint.OrgID = orgID
	if checkpoint.CreatedAt.IsZero() {
		checkpoint.CreatedAt = time.Now().UTC()
	}

	s.mu.Lock()
	s.checkpoints[checkpoint.WorkflowID] = &checkpoint
	s.mu.Unlock()

	// Update autonomous_plans current checkpoint milestone in DB if available
	if s.db != nil {
		_, _ = s.db.ExecContext(ctx, `
			UPDATE autonomous_plans 
			SET current_step_id = ?,
			    updated_at = NOW()
			WHERE org_id = ? AND plan_id = ?
		`, string(checkpoint.Milestone), orgID, checkpoint.WorkflowID)
	}

	return nil
}

func (s *defaultEnterpriseResilienceService) GetLatestCheckpoint(
	ctx context.Context,
	orgID int64,
	workflowID string,
) (*WorkflowCheckpoint, error) {
	if orgID <= 0 || workflowID == "" {
		return nil, ErrUnauthorizedTenant
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	cp, exists := s.checkpoints[workflowID]
	if !exists {
		return &WorkflowCheckpoint{
			WorkflowID: workflowID,
			OrgID:      orgID,
			Milestone:  CheckpointPlanningComplete,
			CreatedAt:  time.Now().UTC(),
		}, nil
	}

	return cp, nil
}

// ---------------------------------------------------------------------
// 4. Stuck Workflow Detection & Recovery
// ---------------------------------------------------------------------

func (s *defaultEnterpriseResilienceService) DetectStuckWorkflows(
	ctx context.Context,
	orgID int64,
	inactivityThreshold time.Duration,
) ([]StuckWorkflowDescriptor, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	if inactivityThreshold <= 0 {
		inactivityThreshold = 10 * time.Minute
	}

	descriptors := make([]StuckWorkflowDescriptor, 0)
	now := time.Now().UTC()

	if s.db == nil {
		return descriptors, nil
	}

	query := `
		SELECT plan_id, org_id, module, status, current_step_id, updated_at
		FROM autonomous_plans
		WHERE org_id = ?
		  AND status IN ('RUNNING', 'PENDING', 'WAITING')
		  AND updated_at < DATE_SUB(NOW(), INTERVAL ? SECOND)
		ORDER BY updated_at ASC
		LIMIT 50
	`

	var rows []struct {
		PlanID        string         `db:"plan_id"`
		OrgID         int64          `db:"org_id"`
		Module        string         `db:"module"`
		Status        string         `db:"status"`
		CurrentStepID sql.NullString `db:"current_step_id"`
		UpdatedAt     time.Time      `db:"updated_at"`
	}

	thresholdSec := int(inactivityThreshold.Seconds())
	if err := s.db.SelectContext(ctx, &rows, query, orgID, thresholdSec); err != nil {
		return nil, fmt.Errorf("failed detecting stuck workflows: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, r := range rows {
		duration := now.Sub(r.UpdatedAt)
		desc := StuckWorkflowDescriptor{
			WorkflowID:         r.PlanID,
			OrgID:              r.OrgID,
			WorkflowType:       EnterpriseWorkflowType(r.Module),
			CurrentState:       EnterpriseWorkflowState(r.Status),
			CurrentStepID:      r.CurrentStepID.String,
			StuckSince:         r.UpdatedAt,
			InactivityDuration: duration,
			SuspectedCause:     "No state transition or step heartbeat within expected timeout window",
			CanAutoRecover:     true,
			RecommendedAction:  "Resume from last verified checkpoint or escalate to human dispatcher",
		}

		if cp, ok := s.checkpoints[r.PlanID]; ok {
			desc.LastMilestone = cp.Milestone
		} else {
			desc.LastMilestone = CheckpointPlanningComplete
		}

		descriptors = append(descriptors, desc)
	}

	return descriptors, nil
}

func (s *defaultEnterpriseResilienceService) RecoverStuckWorkflow(
	ctx context.Context,
	orgID int64,
	workflowID string,
	actor string,
) error {
	if orgID <= 0 || workflowID == "" {
		return ErrUnauthorizedTenant
	}

	cp, _ := s.GetLatestCheckpoint(ctx, orgID, workflowID)

	reason := fmt.Sprintf("Stuck workflow recovered by operator %s from checkpoint milestone %s", actor, cp.Milestone)
	if s.db != nil {
		_, err := s.db.ExecContext(ctx, `
			UPDATE autonomous_plans
			SET status = 'RUNNING',
			    execution_status = 'RECOVERED_IN_PROGRESS',
			    escalation_reason = ?,
			    updated_at = NOW()
			WHERE org_id = ? AND plan_id = ?
		`, reason, orgID, workflowID)
		if err != nil {
			return fmt.Errorf("failed updating stuck workflow recovery: %w", err)
		}
	}

	if s.auditSvc != nil {
		s.auditSvc.RecordAsync(ctx, domain.CreateAuditLogParams{
			OrgID:        orgID,
			ActorType:    domain.ActorTypeUser,
			Action:       domain.ActionUpdate,
			Module:       domain.ModuleSettings,
			ResourceType: "enterprise_workflow_recovery",
			ResourceID:   workflowID,
			Description:  reason,
		})
	}

	return nil
}

// ---------------------------------------------------------------------
// 5. Dead-Letter & Failed Work Visibility
// ---------------------------------------------------------------------

func (s *defaultEnterpriseResilienceService) GetFailedWorkItems(ctx context.Context, orgID int64, limit int) ([]FailedWorkItem, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]FailedWorkItem, 0)
	for _, it := range s.failedWorkItems {
		if it.OrgID == orgID {
			items = append(items, *it)
			if len(items) >= limit {
				break
			}
		}
	}

	// Also pull FAILED plans from DB if fewer than limit
	if s.db != nil && len(items) < limit {
		var rows []struct {
			PlanID           string         `db:"plan_id"`
			OrgID            int64          `db:"org_id"`
			Module           string         `db:"module"`
			RelatedType      string         `db:"related_entity_type"`
			RelatedID        string         `db:"related_entity_id"`
			EscalationReason sql.NullString `db:"escalation_reason"`
			UpdatedAt        time.Time      `db:"updated_at"`
		}
		_ = s.db.SelectContext(ctx, &rows, `
			SELECT plan_id, org_id, module, related_entity_type, related_entity_id, escalation_reason, updated_at
			FROM autonomous_plans
			WHERE org_id = ? AND status = 'FAILED'
			ORDER BY updated_at DESC
			LIMIT ?
		`, orgID, limit-len(items))

		for _, r := range rows {
			reason := "Workflow execution encountered terminal error"
			if r.EscalationReason.Valid && r.EscalationReason.String != "" {
				reason = r.EscalationReason.String
			}
			cls := s.ClassifyFailure(fmt.Errorf("%s", reason), nil)
			items = append(items, FailedWorkItem{
				ID:                    fmt.Sprintf("fail-wf-%s", r.PlanID),
				OrgID:                 r.OrgID,
				ItemType:              "WORKFLOW",
				WorkflowID:            r.PlanID,
				EntityID:              r.RelatedID,
				EntityType:            r.RelatedType,
				FailureClassification: cls,
				FailureReason:         reason,
				RetryCount:            3,
				MaxRetries:            3,
				IsRetryable:           cls == FailureTypeTransient || cls == FailureTypeTimeout,
				LastAttemptAt:         r.UpdatedAt,
				NextAction:            "Human intervention required to resolve root cause",
				EscalationState:       "ESCALATED_OPERATOR",
			})
		}
	}

	return items, nil
}

func (s *defaultEnterpriseResilienceService) ReplayFailedWorkItem(ctx context.Context, orgID int64, itemID string, actor string) error {
	if orgID <= 0 || itemID == "" {
		return ErrUnauthorizedTenant
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.failedWorkItems[itemID]
	if !exists {
		return fmt.Errorf("failed work item %s not found", itemID)
	}

	if !item.IsRetryable {
		return ErrNonRetryableFailure
	}

	item.RetryCount++
	item.LastAttemptAt = time.Now().UTC()
	item.EscalationState = "REPLAYED"

	return nil
}

// ---------------------------------------------------------------------
// 6. Event Storm Protection & Backpressure
// ---------------------------------------------------------------------

func (s *defaultEnterpriseResilienceService) CheckEventStormProtection(
	ctx context.Context,
	orgID int64,
	entityType, entityID, eventType string,
) (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	key := fmt.Sprintf("%d:%s:%s", orgID, strings.ToUpper(entityType), strings.TrimSpace(entityID))

	// Clean up old events outside 10-second sliding window
	windowCutoff := now.Add(-10 * time.Second)
	existing := s.eventArrivalWindow[key]
	valid := make([]time.Time, 0, len(existing))
	for _, t := range existing {
		if t.After(windowCutoff) {
			valid = append(valid, t)
		}
	}

	// Storm threshold: > 8 events for same entity within 10 seconds triggers coalescing
	if len(valid) >= 8 {
		s.coalescedCounts[orgID]++
		valid = append(valid, now)
		s.eventArrivalWindow[key] = valid
		return false, fmt.Sprintf("Event storm detected: %d events for %s:%s within 10s. Event coalesced.", len(valid), entityType, entityID)
	}

	// Throttling threshold: > 20 total active events in backlog
	if len(s.failedWorkItems) > 200 {
		s.throttledCounts[orgID]++
		return false, "Mesh backpressure active: backlog exceeded threshold (200). Ingestion throttled."
	}

	valid = append(valid, now)
	s.eventArrivalWindow[key] = valid
	return true, ""
}

func (s *defaultEnterpriseResilienceService) GetBackpressureMetrics(ctx context.Context, orgID int64) (*EventBackpressureMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	coalesced := s.coalescedCounts[orgID]
	throttled := s.throttledCounts[orgID]

	backlog := len(s.failedWorkItems)
	isActive := backlog > 50

	return &EventBackpressureMetrics{
		OrgID:                orgID,
		CurrentBacklog:       backlog,
		BacklogThreshold:     50,
		BackpressureActive:   isActive,
		CoalescedEventCount:  coalesced,
		ThrottledEventCount:  throttled,
		DeadLetterCount:      backlog,
		ActiveWorkers:        8,
		WorkerConcurrencyCap: 20,
		LastEvaluatedAt:      time.Now().UTC(),
	}, nil
}

// ---------------------------------------------------------------------
// 7. Partial Multi-Agent Failure Reconciliation
// ---------------------------------------------------------------------

func (s *defaultEnterpriseResilienceService) ReconcileMultiAgentExecution(
	ctx context.Context,
	workflowID string,
	originalConfidence float64,
	results []MultiAgentTaskResult,
) (*MultiAgentReconciliationResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no agent execution results provided for reconciliation")
	}

	successful := make([]string, 0)
	failed := make([]string, 0)
	consolidated := make(map[string]interface{})
	hasCriticalPathFailure := false

	for _, res := range results {
		if res.Success {
			successful = append(successful, res.AgentID)
			for k, v := range res.Output {
				consolidated[fmt.Sprintf("%s:%s", res.AgentID, k)] = v
			}
		} else {
			failed = append(failed, res.AgentID)
			if res.IsCriticalPath {
				hasCriticalPathFailure = true
			}
		}
	}

	// Calculate adjusted confidence score
	successRatio := float64(len(successful)) / float64(len(results))
	confidenceImpact := (1.0 - successRatio) * 0.4
	adjustedConfidence := math.Max(0.1, originalConfidence-confidenceImpact)

	recon := &MultiAgentReconciliationResult{
		WorkflowID:         workflowID,
		SuccessfulAgents:   successful,
		FailedAgents:       failed,
		OriginalConfidence: originalConfidence,
		AdjustedConfidence: adjustedConfidence,
		ConfidenceImpact:   confidenceImpact,
		ConsolidatedOutput: consolidated,
	}

	if len(failed) == 0 {
		recon.OverallStatus = "SUCCEEDED"
		recon.CanProceedWithPlan = true
		recon.RequiresHumanReview = false
		recon.FallbackStrategy = "NONE"
	} else if hasCriticalPathFailure {
		recon.OverallStatus = "FAILED"
		recon.CanProceedWithPlan = false
		recon.RequiresHumanReview = true
		recon.FallbackStrategy = "ESCALATE_TO_DISPATCHER"
	} else {
		// Non-critical specialist failed: Degrade gracefully with fallback
		recon.OverallStatus = "PARTIAL_SUCCESS"
		recon.CanProceedWithPlan = adjustedConfidence >= 0.60
		recon.RequiresHumanReview = adjustedConfidence < 0.75
		recon.FallbackStrategy = "RULE_BASED_SPECIALIST_FALLBACK"
	}

	return recon, nil
}

// ---------------------------------------------------------------------
// 8. Notification Failure vs Business Action Separation
// ---------------------------------------------------------------------

func (s *defaultEnterpriseResilienceService) ExecuteWithNotificationSeparation(
	ctx context.Context,
	orgID int64,
	workflowID string,
	businessAction func() error,
	notificationAction func() error,
) (bool, bool, error) {
	// 1. Authoritative business action execution first
	if err := businessAction(); err != nil {
		// Business action failed: entire operation halts
		return false, false, err
	}

	// 2. Notification executed independently
	notifErr := notificationAction()
	if notifErr != nil {
		// Business effect succeeded! Record notification failure separately in failedWorkItems
		s.mu.Lock()
		itemID := fmt.Sprintf("notif-fail-%s-%d", workflowID, time.Now().UnixNano())
		cls := s.ClassifyFailure(notifErr, nil)
		s.failedWorkItems[itemID] = &FailedWorkItem{
			ID:                    itemID,
			OrgID:                 orgID,
			ItemType:              "NOTIFICATION",
			WorkflowID:            workflowID,
			FailureClassification: cls,
			FailureReason:         fmt.Sprintf("Business action completed, but notification delivery failed: %v", notifErr),
			RetryCount:            0,
			MaxRetries:            5,
			IsRetryable:           true,
			LastAttemptAt:         time.Now().UTC(),
			NextAction:            "Scheduled for async retry without re-executing business action",
			EscalationState:       "PENDING_RETRY",
		}
		s.mu.Unlock()

		return true, false, nil
	}

	return true, true, nil
}
