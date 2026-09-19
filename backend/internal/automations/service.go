package automations

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/freel/backend/internal/audit/domain"
	auditService "github.com/freel/backend/internal/audit/service"
	"github.com/freel/backend/internal/recommendations"
	"github.com/jmoiron/sqlx"
)

type Service interface {
	ListAutomations(ctx context.Context, orgID int64, filter AutomationFilter) ([]*Automation, int64, error)
	GetAutomation(ctx context.Context, orgID int64, id int64) (*Automation, error)
	CreateAutomation(ctx context.Context, orgID int64, userID int64, input CreateAutomationInput) (*Automation, error)
	UpdateAutomation(ctx context.Context, orgID int64, id int64, userID int64, input UpdateAutomationInput) (*Automation, error)
	SetEnabled(ctx context.Context, orgID int64, id int64, isEnabled bool, userID int64) (*Automation, error)
	DeleteAutomation(ctx context.Context, orgID int64, id int64, userID int64) error
	PreviewNextRun(ctx context.Context, orgID int64, id int64) (*NextRunPreview, error)
	TriggerManualRun(ctx context.Context, orgID int64, id int64, userID int64) (*AutomationExecution, error)
	CancelExecution(ctx context.Context, orgID int64, execID int64, userID int64, reason string) (*AutomationExecution, error)
	RetryExecution(ctx context.Context, orgID int64, execID int64, userID int64) (*AutomationExecution, error)
	ListExecutions(ctx context.Context, orgID int64, filter ExecutionFilter) ([]*AutomationExecution, int64, error)
	GetExecution(ctx context.Context, orgID int64, execID int64) (*AutomationExecution, error)
	GetExecutionRecommendations(ctx context.Context, orgID int64, execID int64) ([]*recommendations.Recommendation, error)
	GetExecutionInsights(ctx context.Context, orgID int64, execID int64) ([]*OperationalInsight, error)
	GetStats(ctx context.Context, orgID int64) (*AutomationStats, error)
	GetSupportedTypes() []SupportedAutomationTypeInfo

	// Phase 3 Operational Intelligence & Event Ingestion
	EvaluateBusinessEvent(ctx context.Context, orgID int64, userID int64, input EvaluateEventInput) (*EvaluateEventResult, error)
	ListInsights(ctx context.Context, orgID int64, filter OperationalInsightFilter) ([]*OperationalInsight, int64, error)
	GetInsight(ctx context.Context, orgID int64, id int64) (*OperationalInsight, error)
	AcknowledgeInsight(ctx context.Context, orgID int64, id int64, userID int64) (*OperationalInsight, error)
	DismissInsight(ctx context.Context, orgID int64, id int64, userID int64, reason string) (*OperationalInsight, error)

	// Worker / Scheduler execution engine
	ExecuteJob(ctx context.Context, auto *Automation, exec *AutomationExecution) error
}

type service struct {
	repo      Repository
	generator recommendations.Generator
	recRepo   recommendations.Repository
	auditSvc  auditService.Service
	db        *sqlx.DB
	mu        sync.Mutex
}

func NewService(
	repo Repository,
	generator recommendations.Generator,
	recRepo recommendations.Repository,
	auditSvc auditService.Service,
) Service {
	return &service{
		repo:      repo,
		generator: generator,
		recRepo:   recRepo,
		auditSvc:  auditSvc,
	}
}

func (s *service) SetDB(db *sqlx.DB) {
	s.db = db
}

func (s *service) recordAudit(ctx context.Context, orgID int64, actorID *int64, action, resType, resID, resName, desc, result, errMsg string, meta map[string]interface{}) {
	if s.auditSvc == nil {
		return
	}
	actorType := "SYSTEM"
	if actorID != nil && *actorID > 0 {
		actorType = "USER"
	}
	if result == "" {
		result = domain.ResultSuccess
	}
	_, _ = s.auditSvc.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorID:      actorID,
		ActorType:    actorType,
		Action:       action,
		Module:       "AUTOMATION",
		ResourceType: resType,
		ResourceID:   resID,
		ResourceName: resName,
		Description:  desc,
		Result:       result,
		ErrorMessage: errMsg,
		Metadata:     meta,
	})
}

func (s *service) GetSupportedTypes() []SupportedAutomationTypeInfo {
	return GetSupportedAutomationTypes()
}

func (s *service) ListAutomations(ctx context.Context, orgID int64, filter AutomationFilter) ([]*Automation, int64, error) {
	return s.repo.List(ctx, orgID, filter)
}

func (s *service) GetAutomation(ctx context.Context, orgID int64, id int64) (*Automation, error) {
	return s.repo.GetByID(ctx, orgID, id)
}

func (s *service) CreateAutomation(ctx context.Context, orgID int64, userID int64, input CreateAutomationInput) (*Automation, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, fmt.Errorf("automation name is required")
	}
	if !ValidateSupportedType(input.AutomationType) {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedType, input.AutomationType)
	}

	schedType, schedTime, schedDays, tz, err := ValidateSchedule(
		input.ScheduleType,
		input.ScheduleTime,
		input.ScheduleDays,
		input.Timezone,
	)
	if err != nil {
		return nil, err
	}
	input.ScheduleType = schedType
	input.ScheduleTime = schedTime
	input.ScheduleDays = schedDays
	input.Timezone = tz

	if input.ExecutionWindowMinutes <= 0 {
		input.ExecutionWindowMinutes = 60
	}
	if input.MaxExecutionDurationSec <= 0 {
		input.MaxExecutionDurationSec = 300
	}

	nextRun, err := CalculateNextRun(input.ScheduleType, input.ScheduleTime, input.ScheduleDays, input.Timezone, time.Now())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSchedule, err)
	}

	var configJSON *string
	if len(input.Configuration) > 0 && string(input.Configuration) != "null" {
		str := string(input.Configuration)
		configJSON = &str
	}

	var triggerConfigJSON *string
	if len(input.TriggerConfig) > 0 && string(input.TriggerConfig) != "null" {
		str := string(input.TriggerConfig)
		triggerConfigJSON = &str
	}

	var retryPolicyJSON *string
	if len(input.RetryPolicy) > 0 && string(input.RetryPolicy) != "null" {
		str := string(input.RetryPolicy)
		retryPolicyJSON = &str
	}

	var targetModsJSON *string
	if len(input.TargetModules) > 0 {
		b, _ := json.Marshal(input.TargetModules)
		str := string(b)
		targetModsJSON = &str
	}

	var allowedActsJSON *string
	if len(input.AllowedActions) > 0 {
		b, _ := json.Marshal(input.AllowedActions)
		str := string(b)
		allowedActsJSON = &str
	}

	isEnabled := true
	if input.IsEnabled != nil {
		isEnabled = *input.IsEnabled
	}

	corrID := fmt.Sprintf("auto-init-%d-%d", orgID, time.Now().UnixNano())
	auto := &Automation{
		OrgID:                  orgID,
		Name:                   strings.TrimSpace(input.Name),
		AutomationType:         input.AutomationType,
		Description:            input.Description,
		IsEnabled:              isEnabled,
		TriggerType:            ValidateTriggerType(input.TriggerType),
		TriggerConfigJSON:      triggerConfigJSON,
		Scope:                  input.Scope,
		ApprovalPolicy:         ValidateApprovalPolicy(input.ApprovalPolicy),
		AllowedActionsJSON:     allowedActsJSON,
		OwnerTeam:              input.OwnerTeam,
		Priority:               input.Priority,
		ScheduleType:           input.ScheduleType,
		ScheduleTime:           input.ScheduleTime,
		ScheduleDays:           input.ScheduleDays,
		Timezone:               input.Timezone,
		ExecutionWindowMinutes: input.ExecutionWindowMinutes,
		ConfigurationJSON:      configJSON,
		TargetModulesJSON:      targetModsJSON,
		NextExecutionAt:        &nextRun,
		RetryPolicyJSON:        retryPolicyJSON,
		MaxExecutionDurationSec: input.MaxExecutionDurationSec,
		CorrelationID:          &corrID,
		CreatedBy:              userID,
		UpdatedBy:              userID,
	}

	created, err := s.repo.Create(ctx, auto)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, orgID, &userID, domain.ActionCreate, "AUTOMATION", fmt.Sprintf("%d", created.ID), created.Name,
		fmt.Sprintf("Created automation %s (%s)", created.Name, created.AutomationType), domain.ResultSuccess, "", map[string]interface{}{
			"automation_id":   created.ID,
			"automation_type": created.AutomationType,
			"trigger_type":    created.TriggerType,
			"schedule_type":   created.ScheduleType,
			"schedule_time":   created.ScheduleTime,
			"next_execution":  nextRun,
		})

	return created, nil
}

func (s *service) UpdateAutomation(ctx context.Context, orgID int64, id int64, userID int64, input UpdateAutomationInput) (*Automation, error) {
	existing, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(input.Name) != "" {
		existing.Name = strings.TrimSpace(input.Name)
	}
	if input.Description != nil {
		existing.Description = input.Description
	}
	if input.TriggerType != "" {
		existing.TriggerType = ValidateTriggerType(input.TriggerType)
	}
	if input.Scope != "" {
		existing.Scope = input.Scope
	}
	if input.ApprovalPolicy != "" {
		existing.ApprovalPolicy = ValidateApprovalPolicy(input.ApprovalPolicy)
	}
	if input.OwnerTeam != "" {
		existing.OwnerTeam = input.OwnerTeam
	}
	if input.Priority != "" {
		existing.Priority = input.Priority
	}
	if input.MaxExecutionDurationSec > 0 {
		existing.MaxExecutionDurationSec = input.MaxExecutionDurationSec
	}

	if input.ScheduleType != "" || input.ScheduleTime != "" || input.ScheduleDays != nil || input.Timezone != "" {
		st := existing.ScheduleType
		if input.ScheduleType != "" {
			st = input.ScheduleType
		}
		stim := existing.ScheduleTime
		if input.ScheduleTime != "" {
			stim = input.ScheduleTime
		}
		sd := existing.ScheduleDays
		if input.ScheduleDays != nil {
			sd = input.ScheduleDays
		}
		tz := existing.Timezone
		if input.Timezone != "" {
			tz = input.Timezone
		}

		schedType, schedTime, schedDays, timezone, err := ValidateSchedule(st, stim, sd, tz)
		if err != nil {
			return nil, err
		}
		existing.ScheduleType = schedType
		existing.ScheduleTime = schedTime
		existing.ScheduleDays = schedDays
		existing.Timezone = timezone

		nextRun, err := CalculateNextRun(existing.ScheduleType, existing.ScheduleTime, existing.ScheduleDays, existing.Timezone, time.Now())
		if err == nil {
			existing.NextExecutionAt = &nextRun
		}
	}

	if input.ExecutionWindowMinutes > 0 {
		existing.ExecutionWindowMinutes = input.ExecutionWindowMinutes
	}

	if len(input.Configuration) > 0 && string(input.Configuration) != "null" {
		str := string(input.Configuration)
		existing.ConfigurationJSON = &str
	}
	if len(input.TriggerConfig) > 0 && string(input.TriggerConfig) != "null" {
		str := string(input.TriggerConfig)
		existing.TriggerConfigJSON = &str
	}
	if len(input.RetryPolicy) > 0 && string(input.RetryPolicy) != "null" {
		str := string(input.RetryPolicy)
		existing.RetryPolicyJSON = &str
	}

	if len(input.TargetModules) > 0 {
		b, _ := json.Marshal(input.TargetModules)
		str := string(b)
		existing.TargetModulesJSON = &str
	}
	if len(input.AllowedActions) > 0 {
		b, _ := json.Marshal(input.AllowedActions)
		str := string(b)
		existing.AllowedActionsJSON = &str
	}

	if input.IsEnabled != nil {
		existing.IsEnabled = *input.IsEnabled
	}

	existing.UpdatedBy = userID

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	s.recordAudit(ctx, orgID, &userID, domain.ActionUpdate, "AUTOMATION", fmt.Sprintf("%d", id), existing.Name,
		fmt.Sprintf("Updated automation %s", existing.Name), domain.ResultSuccess, "", map[string]interface{}{
			"automation_id": id,
			"name":          existing.Name,
		})

	return s.repo.GetByID(ctx, orgID, id)
}

func (s *service) SetEnabled(ctx context.Context, orgID int64, id int64, isEnabled bool, userID int64) (*Automation, error) {
	if err := s.repo.SetEnabled(ctx, orgID, id, isEnabled, userID); err != nil {
		return nil, err
	}

	if isEnabled {
		auto, err := s.repo.GetByID(ctx, orgID, id)
		if err == nil && auto != nil {
			nextRun, _ := CalculateNextRun(auto.ScheduleType, auto.ScheduleTime, auto.ScheduleDays, auto.Timezone, time.Now())
			_ = s.repo.UpdateExecutionTimestamps(ctx, id, time.Time{}, &nextRun, "", "")
		}
	}

	action := domain.ActionEnable
	desc := fmt.Sprintf("Enabled automation #%d", id)
	if !isEnabled {
		action = domain.ActionDisable
		desc = fmt.Sprintf("Disabled automation #%d", id)
	}
	s.recordAudit(ctx, orgID, &userID, action, "AUTOMATION", fmt.Sprintf("%d", id), "", desc, domain.ResultSuccess, "", nil)

	return s.repo.GetByID(ctx, orgID, id)
}

func (s *service) DeleteAutomation(ctx context.Context, orgID int64, id int64, userID int64) error {
	auto, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, orgID, id); err != nil {
		return err
	}

	s.recordAudit(ctx, orgID, &userID, domain.ActionDelete, "AUTOMATION", fmt.Sprintf("%d", id), auto.Name,
		fmt.Sprintf("Deleted automation %s", auto.Name), domain.ResultSuccess, "", map[string]interface{}{
			"name":            auto.Name,
			"automation_type": auto.AutomationType,
		})

	return nil
}

func (s *service) PreviewNextRun(ctx context.Context, orgID int64, id int64) (*NextRunPreview, error) {
	auto, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	nextRun, err := CalculateNextRun(auto.ScheduleType, auto.ScheduleTime, auto.ScheduleDays, auto.Timezone, time.Now())
	if err != nil {
		return nil, err
	}

	return FormatNextRunPreview(nextRun, auto.Timezone), nil
}

func (s *service) TriggerManualRun(ctx context.Context, orgID int64, id int64, userID int64) (*AutomationExecution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	auto, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	if !auto.IsEnabled {
		return nil, ErrAutomationDisabled
	}

	active, err := s.repo.GetActiveExecutionForAutomation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, ErrExecutionAlreadyRunning
	}

	corrID := fmt.Sprintf("exec-manual-%d-%d", id, time.Now().UnixNano())
	triggerEvt := "MANUAL_DISPATCH"
	exec := &AutomationExecution{
		AutomationID:      id,
		OrgID:             orgID,
		CorrelationID:     corrID,
		TriggerType:       TriggerTypeManual,
		TriggerEvent:      &triggerEvt,
		CurrentStep:       StepInitializing,
		TriggeredByUserID: &userID,
		Status:            ExecutionStatusQueued,
		DetailsJSON:       nil,
	}

	createdExec, err := s.repo.CreateExecution(ctx, exec)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, orgID, &userID, "RUN", "AUTOMATION_EXECUTION", fmt.Sprintf("%d", createdExec.ID), auto.Name,
		fmt.Sprintf("Triggered manual run for automation %s", auto.Name), domain.ResultSuccess, "", map[string]interface{}{
			"automation_id":   id,
			"automation_name": auto.Name,
			"correlation_id":  corrID,
		})

	go func() {
		bgCtx := context.Background()
		_ = s.ExecuteJob(bgCtx, auto, createdExec)
	}()

	return s.repo.GetExecutionByID(ctx, orgID, createdExec.ID)
}

func (s *service) CancelExecution(ctx context.Context, orgID int64, execID int64, userID int64, reason string) (*AutomationExecution, error) {
	exec, err := s.repo.GetExecutionByID(ctx, orgID, execID)
	if err != nil {
		return nil, err
	}

	if exec.Status != ExecutionStatusQueued && exec.Status != ExecutionStatusRunning && exec.Status != ExecutionStatusWaitingForApproval {
		return nil, fmt.Errorf("cannot cancel execution with status %s", exec.Status)
	}

	now := time.Now()
	exec.Status = ExecutionStatusCancelled
	exec.CurrentStep = StepCancelled
	exec.CompletedAt = &now
	if reason == "" {
		reason = "Cancelled by operator"
	}
	exec.ErrorMessage = &reason

	if err := s.repo.UpdateExecution(ctx, exec); err != nil {
		return nil, err
	}

	s.recordAudit(ctx, orgID, &userID, "CANCEL", "AUTOMATION_EXECUTION", fmt.Sprintf("%d", execID), "",
		fmt.Sprintf("Cancelled execution #%d: %s", execID, reason), domain.ResultSuccess, "", map[string]interface{}{
			"reason": reason,
		})

	return s.repo.GetExecutionByID(ctx, orgID, execID)
}

func (s *service) RetryExecution(ctx context.Context, orgID int64, execID int64, userID int64) (*AutomationExecution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	exec, err := s.repo.GetExecutionByID(ctx, orgID, execID)
	if err != nil {
		return nil, err
	}

	if exec.Status != ExecutionStatusFailed {
		return nil, fmt.Errorf("can only retry failed executions (current status: %s)", exec.Status)
	}

	auto, err := s.repo.GetByID(ctx, orgID, exec.AutomationID)
	if err != nil {
		return nil, err
	}

	if exec.RetryCount >= auto.MaxRetries && auto.MaxRetries > 0 {
		return nil, fmt.Errorf("retry limit exceeded (%d/%d)", exec.RetryCount, auto.MaxRetries)
	}

	exec.RetryCount++
	exec.Status = ExecutionStatusRetrying
	exec.CurrentStep = StepInitializing
	exec.ErrorMessage = nil

	if err := s.repo.UpdateExecution(ctx, exec); err != nil {
		return nil, err
	}

	s.recordAudit(ctx, orgID, &userID, "RETRY", "AUTOMATION_EXECUTION", fmt.Sprintf("%d", execID), auto.Name,
		fmt.Sprintf("Retrying failed execution #%d (attempt %d)", execID, exec.RetryCount), domain.ResultSuccess, "", map[string]interface{}{
			"automation_id": auto.ID,
			"retry_count":   exec.RetryCount,
		})

	go func() {
		bgCtx := context.Background()
		_ = s.ExecuteJob(bgCtx, auto, exec)
	}()

	return s.repo.GetExecutionByID(ctx, orgID, execID)
}

func (s *service) ListExecutions(ctx context.Context, orgID int64, filter ExecutionFilter) ([]*AutomationExecution, int64, error) {
	return s.repo.ListExecutions(ctx, orgID, filter)
}

func (s *service) GetExecution(ctx context.Context, orgID int64, execID int64) (*AutomationExecution, error) {
	return s.repo.GetExecutionByID(ctx, orgID, execID)
}

func (s *service) GetExecutionRecommendations(ctx context.Context, orgID int64, execID int64) ([]*recommendations.Recommendation, error) {
	recs, _, err := s.recRepo.List(ctx, orgID, recommendations.RecommendationFilter{
		ExecutionID: &execID,
		Limit:       100,
	})
	return recs, err
}

func (s *service) GetExecutionInsights(ctx context.Context, orgID int64, execID int64) ([]*OperationalInsight, error) {
	insights, _, err := s.repo.ListInsights(ctx, orgID, OperationalInsightFilter{
		ExecutionID: &execID,
		Limit:       100,
	})
	return insights, err
}

func (s *service) GetStats(ctx context.Context, orgID int64) (*AutomationStats, error) {
	return s.repo.GetStats(ctx, orgID)
}

// ExecuteJob runs an automation execution instance through deterministic assistants with durable step lifecycle
func (s *service) ExecuteJob(ctx context.Context, auto *Automation, exec *AutomationExecution) error {
	startTime := time.Now()
	exec.Status = ExecutionStatusRunning
	exec.CurrentStep = StepEvaluatingRules
	exec.StartedAt = &startTime
	_ = s.repo.UpdateExecution(ctx, exec)

	log.Printf("🤖 [Phase 3 Automation Worker] Starting %s (#%d) for Org %d [Exec #%d, Corr: %s]",
		auto.AutomationType, auto.ID, auto.OrgID, exec.ID, exec.CorrelationID)

	// Step 1: Evaluate deterministic rules through recommendation generator
	genRes, err := s.generator.GenerateForAutomation(
		ctx,
		auto.OrgID,
		auto.AutomationType,
		exec.CorrelationID,
		auto.CreatedBy,
		auto.ID,
		exec.ID,
	)

	completedTime := time.Now()
	durationMs := completedTime.Sub(startTime).Milliseconds()
	exec.CompletedAt = &completedTime
	exec.DurationMs = durationMs

	if err != nil {
		log.Printf("❌ [Phase 3 Automation Worker] Execution #%d failed: %v", exec.ID, err)
		errMsg := err.Error()
		exec.Status = ExecutionStatusFailed
		exec.CurrentStep = StepFailed
		exec.ErrorMessage = &errMsg
		_ = s.repo.UpdateExecution(ctx, exec)

		_ = s.repo.UpdateExecutionTimestamps(ctx, auto.ID, startTime, auto.NextExecutionAt, ExecutionStatusFailed, errMsg)

		s.recordAudit(ctx, auto.OrgID, nil, "EXECUTE", "AUTOMATION_EXECUTION", fmt.Sprintf("%d", exec.ID), auto.Name,
			fmt.Sprintf("Automation execution #%d failed: %s", exec.ID, errMsg), domain.ResultFailed, errMsg, map[string]interface{}{
				"automation_id":   auto.ID,
				"automation_type": auto.AutomationType,
				"error":           errMsg,
				"duration_ms":     durationMs,
			})
		return err
	}

	// Step 2: Convert and persist durable Operational Insights from generator results
	exec.CurrentStep = StepGeneratingInsights
	_ = s.repo.UpdateExecution(ctx, exec)

	var insightsCreatedCount int
	if len(genRes.TopPriorityItems) > 0 {
		for _, item := range genRes.TopPriorityItems {
			// Check if active insight already exists for this record & rule
			existingInsight, _ := s.repo.GetActiveInsightForRecord(ctx, auto.OrgID, item.SourceType, item.SourceID, item.RuleApplied)
			if existingInsight == nil {
				evidenceBytes, _ := json.Marshal(item.Evidence)
				evidenceStr := string(evidenceBytes)

				sourceMod := strings.ToLower(item.SourceType)
				recAction := item.RecommendedAction
				actionType := item.ActionType
				sourceRef := item.SourceReference

				insight := &OperationalInsight{
					OrgID:                 auto.OrgID,
					AutomationID:          &auto.ID,
					ExecutionID:           &exec.ID,
					SourceModule:          sourceMod,
					SourceRecordID:        item.SourceID,
					SourceRecordRef:       &sourceRef,
					InsightType:           item.RuleApplied,
					Severity:              strings.ToUpper(item.Priority),
					Title:                 item.Title,
					Description:           item.Description,
					EvidenceJSON:          &evidenceStr,
					DetectionRule:         item.RuleApplied,
					Confidence:            item.ConfidenceScore,
					Priority:              strings.ToUpper(item.Priority),
					RiskLevel:             strings.ToUpper(item.RiskLevel),
					DataFreshness:         "REAL_TIME",
					RecommendedNextStep:   &recAction,
					IsApprovalRequired:    item.RequiresApproval,
					RecommendedActionType: &actionType,
					Status:                InsightStatusActive,
					CorrelationID:         exec.CorrelationID,
				}

				createdInsight, err := s.repo.CreateInsight(ctx, insight)
				if err == nil && createdInsight != nil {
					insightsCreatedCount++
				}
			}
		}
	}

	// Step 3: Execution succeeded
	exec.Status = ExecutionStatusCompleted
	exec.CurrentStep = StepCompleted
	exec.RecordsReviewed = genRes.TotalEvaluated
	exec.RecommendationsCreated = genRes.CreatedCount
	exec.RecommendationsUpdated = genRes.UpdatedCount

	var summary string
	if genRes.CreatedCount == 0 && genRes.UpdatedCount == 0 && insightsCreatedCount == 0 {
		summary = "Analysis completed successfully. No new operational signals or risks detected. All target records verified."
	} else {
		summary = fmt.Sprintf("Analysis completed successfully. Evaluated %d records: created %d new recommendations, refreshed %d items, recorded %d operational insights.",
			genRes.TotalEvaluated, genRes.CreatedCount, genRes.UpdatedCount, insightsCreatedCount)
	}
	exec.SummaryText = &summary

	stepResultsMap := map[string]interface{}{
		"evaluating_rules": map[string]interface{}{
			"status":         "SUCCESS",
			"rules_executed": genRes.RulesExecuted,
			"evaluated":      genRes.TotalEvaluated,
		},
		"generating_insights": map[string]interface{}{
			"status":           "SUCCESS",
			"insights_created": insightsCreatedCount,
			"recs_created":     genRes.CreatedCount,
			"recs_updated":     genRes.UpdatedCount,
		},
		"approval_gating": map[string]interface{}{
			"policy":               auto.ApprovalPolicy,
			"is_approval_required": true,
			"read_only_guarantee":  true,
		},
	}
	stepResultsBytes, _ := json.Marshal(stepResultsMap)
	stepResultsStr := string(stepResultsBytes)
	exec.StepResultsJSON = &stepResultsStr

	detailsMap := map[string]interface{}{
		"rules_executed":            genRes.RulesExecuted,
		"top_priority_items_count": len(genRes.TopPriorityItems),
		"insights_created_count":    insightsCreatedCount,
		"correlation_id":            genRes.CorrelationID,
		"total_evaluated":           genRes.TotalEvaluated,
		"read_only":                 true,
		"automated_mutations":       0,
		"automated_messages":        0,
	}
	detailsBytes, _ := json.Marshal(detailsMap)
	detailsStr := string(detailsBytes)
	exec.DetailsJSON = &detailsStr

	_ = s.repo.UpdateExecution(ctx, exec)

	// Calculate and update next scheduled execution
	nextRun, _ := CalculateNextRun(auto.ScheduleType, auto.ScheduleTime, auto.ScheduleDays, auto.Timezone, completedTime)
	_ = s.repo.UpdateExecutionTimestamps(ctx, auto.ID, startTime, &nextRun, ExecutionStatusCompleted, "")

	log.Printf("✅ [Phase 3 Automation Worker] Execution #%d completed: %s [Next: %s]",
		exec.ID, summary, nextRun.Format(time.RFC3339))

	s.recordAudit(ctx, auto.OrgID, nil, "EXECUTE", "AUTOMATION_EXECUTION", fmt.Sprintf("%d", exec.ID), auto.Name,
		fmt.Sprintf("Automation execution #%d completed successfully", exec.ID), domain.ResultSuccess, "", map[string]interface{}{
			"automation_id":           auto.ID,
			"automation_name":         auto.Name,
			"automation_type":         auto.AutomationType,
			"records_reviewed":        exec.RecordsReviewed,
			"recommendations_created": exec.RecommendationsCreated,
			"recommendations_updated": exec.RecommendationsUpdated,
			"insights_created":        insightsCreatedCount,
			"duration_ms":             durationMs,
			"next_run":                nextRun,
		})

	return nil
}

// ── Phase 3 Operational Intelligence & Event Ingestion ───────────────────────

func (s *service) EvaluateBusinessEvent(ctx context.Context, orgID int64, userID int64, input EvaluateEventInput) (*EvaluateEventResult, error) {
	if strings.TrimSpace(input.EventType) == "" {
		return nil, fmt.Errorf("event_type is required")
	}

	corrID := fmt.Sprintf("event-%s-%d-%d", input.EventType, input.SourceRecordID, time.Now().UnixNano())

	// 1. Check idempotency
	if input.IdempotencyKey != "" {
		existing, err := s.repo.GetExecutionByIdempotencyKey(ctx, orgID, input.IdempotencyKey)
		if err == nil && existing != nil {
			return &EvaluateEventResult{
				EventProcessed:      true,
				SuppressedDuplicate: true,
				MatchedAutomations:  0,
				CorrelationID:       existing.CorrelationID,
				ExecutionStatus:     existing.Status,
				Summary:             fmt.Sprintf("Duplicate event suppressed by idempotency key '%s'. Existing execution #%d is %s.", input.IdempotencyKey, existing.ID, existing.Status),
			}, nil
		}
	}

	// 2. Find matching active automations for this trigger type and module
	matchedAutos, err := s.repo.FindMatchingAutomations(ctx, orgID, input.EventType, input.SourceModule)
	if err != nil {
		return nil, err
	}

	result := &EvaluateEventResult{
		EventProcessed:      true,
		SuppressedDuplicate: false,
		MatchedAutomations:  len(matchedAutos),
		CorrelationID:       corrID,
		ExecutionStatus:     "COMPLETED",
		ExecutionsTriggered: []*AutomationExecution{},
		InsightsCreated:     []*OperationalInsight{},
	}

	recordRef := fmt.Sprintf("%s:%d", input.SourceModule, input.SourceRecordID)
	if input.SourceRecordRef != "" {
		recordRef = input.SourceRecordRef
	}

	// 3. If automations match, trigger executions with durable lifecycle
	for _, auto := range matchedAutos {
		exec := &AutomationExecution{
			AutomationID:      auto.ID,
			OrgID:             orgID,
			CorrelationID:     corrID,
			TriggerType:       input.EventType,
			TriggerEvent:      &input.EventType,
			InputRecordRef:    &recordRef,
			CurrentStep:       StepInitializing,
			TriggeredByUserID: &userID,
			Status:            ExecutionStatusQueued,
			IdempotencyKey:    &input.IdempotencyKey,
		}

		createdExec, err := s.repo.CreateExecution(ctx, exec)
		if err == nil && createdExec != nil {
			result.ExecutionsTriggered = append(result.ExecutionsTriggered, createdExec)

			// Execute asynchronously
			go func(a *Automation, e *AutomationExecution) {
				bgCtx := context.Background()
				_ = s.ExecuteJob(bgCtx, a, e)
			}(auto, createdExec)
		}
	}

	// 4. Create an operational insight for this specific event
	insightTitle := fmt.Sprintf("%s detected on %s #%d", strings.ReplaceAll(input.EventType, "_", " "), strings.ToUpper(input.SourceModule), input.SourceRecordID)
	insightDesc := fmt.Sprintf("Real operational signal %s triggered on %s record #%d. Evaluated against business rules.", input.EventType, input.SourceModule, input.SourceRecordID)
	
	rule := input.EventType
	actionType := "REVIEW_OPERATIONAL_SIGNAL"
	severity := InsightSeverityMedium
	risk := "MEDIUM"

	if strings.Contains(input.EventType, "EXCEPTION") || strings.Contains(input.EventType, "OVERDUE") {
		severity = InsightSeverityHigh
		risk = "HIGH"
	}

	var payloadJSON *string
	if len(input.Payload) > 0 && string(input.Payload) != "null" {
		str := string(input.Payload)
		payloadJSON = &str
	}

	recNextStep := fmt.Sprintf("Review %s #%d and coordinate resolution.", input.SourceModule, input.SourceRecordID)
	insight := &OperationalInsight{
		OrgID:                 orgID,
		SourceModule:          input.SourceModule,
		SourceRecordID:        input.SourceRecordID,
		SourceRecordRef:       &recordRef,
		InsightType:           input.EventType,
		Severity:              severity,
		Title:                 insightTitle,
		Description:           insightDesc,
		EvidenceJSON:          payloadJSON,
		DetectionRule:         rule,
		Confidence:            1.0,
		Priority:              severity,
		RiskLevel:             risk,
		DataFreshness:         "REAL_TIME",
		RecommendedNextStep:   &recNextStep,
		IsApprovalRequired:    true,
		RecommendedActionType: &actionType,
		Status:                InsightStatusActive,
		CorrelationID:         corrID,
	}

	createdInsight, err := s.repo.CreateInsight(ctx, insight)
	if err == nil && createdInsight != nil {
		result.InsightsCreated = append(result.InsightsCreated, createdInsight)
	}

	result.Summary = fmt.Sprintf("Processed event '%s' for %s #%d. Triggered %d automations, created operational insight.",
		input.EventType, input.SourceModule, input.SourceRecordID, len(result.ExecutionsTriggered))

	s.recordAudit(ctx, orgID, &userID, "EVALUATE_EVENT", "BUSINESS_EVENT", recordRef, input.EventType,
		result.Summary, domain.ResultSuccess, "", map[string]interface{}{
			"event_type":     input.EventType,
			"source_module":  input.SourceModule,
			"record_id":      input.SourceRecordID,
			"correlation_id": corrID,
			"triggered":      len(result.ExecutionsTriggered),
		})

	return result, nil
}

func (s *service) ListInsights(ctx context.Context, orgID int64, filter OperationalInsightFilter) ([]*OperationalInsight, int64, error) {
	return s.repo.ListInsights(ctx, orgID, filter)
}

func (s *service) GetInsight(ctx context.Context, orgID int64, id int64) (*OperationalInsight, error) {
	return s.repo.GetInsightByID(ctx, orgID, id)
}

func (s *service) AcknowledgeInsight(ctx context.Context, orgID int64, id int64, userID int64) (*OperationalInsight, error) {
	insight, err := s.repo.GetInsightByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateInsightStatus(ctx, orgID, id, InsightStatusAcknowledged, nil, nil); err != nil {
		return nil, err
	}

	s.recordAudit(ctx, orgID, &userID, "ACKNOWLEDGE", "OPERATIONAL_INSIGHT", fmt.Sprintf("%d", id), insight.Title,
		fmt.Sprintf("Acknowledged operational insight #%d: %s", id, insight.Title), domain.ResultSuccess, "", nil)

	return s.repo.GetInsightByID(ctx, orgID, id)
}

func (s *service) DismissInsight(ctx context.Context, orgID int64, id int64, userID int64, reason string) (*OperationalInsight, error) {
	insight, err := s.repo.GetInsightByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateInsightStatus(ctx, orgID, id, InsightStatusDismissed, nil, nil); err != nil {
		return nil, err
	}

	s.recordAudit(ctx, orgID, &userID, "DISMISS", "OPERATIONAL_INSIGHT", fmt.Sprintf("%d", id), insight.Title,
		fmt.Sprintf("Dismissed operational insight #%d: %s (reason: %s)", id, insight.Title, reason), domain.ResultSuccess, "", map[string]interface{}{
			"reason": reason,
		})

	return s.repo.GetInsightByID(ctx, orgID, id)
}
