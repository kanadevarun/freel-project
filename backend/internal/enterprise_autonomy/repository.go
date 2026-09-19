package enterprise_autonomy

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	CreateWorkflow(ctx context.Context, wf *EnterpriseWorkflow) error
	GetWorkflow(ctx context.Context, orgID int64, workflowID string) (*EnterpriseWorkflow, error)
	UpdateWorkflowState(ctx context.Context, orgID int64, workflowID string, state EnterpriseWorkflowState, errCode, errMsg *string) error
	UpdateWorkflowStop(ctx context.Context, orgID int64, workflowID string, stoppedByUserID *int64, stopReason string) error
	ListWorkflows(ctx context.Context, filter WorkflowFilter) ([]*EnterpriseWorkflow, int, error)
	SaveWorkflowSteps(ctx context.Context, orgID int64, workflowID string, steps []EnterpriseWorkflowStep) error
	GetWorkflowSteps(ctx context.Context, orgID int64, workflowID string) ([]EnterpriseWorkflowStep, error)
	UpdateWorkflowStep(ctx context.Context, orgID int64, step *EnterpriseWorkflowStep) error
	GetActiveOrInterruptedWorkflows(ctx context.Context) ([]*EnterpriseWorkflow, error)
	GetPolicyContext(ctx context.Context, orgID int64, module string) (*EnterprisePolicyContext, error)
	UpdatePolicyContext(ctx context.Context, orgID int64, policy *EnterprisePolicyContext) error
	CheckAndRecordEventDedup(ctx context.Context, orgID int64, dedupKey, eventType string) (bool, error)
	GetPlatformOverview(ctx context.Context, orgID int64) (*EnterprisePlatformOverview, error)
}

type mysqlRepository struct {
	db *sqlx.DB
}

func NewMySQLRepository(db *sqlx.DB) Repository {
	return &mysqlRepository{db: db}
}

func (r *mysqlRepository) CreateWorkflow(ctx context.Context, wf *EnterpriseWorkflow) error {
	if wf.OrgID <= 0 {
		return ErrUnauthorizedTenant
	}

	query := `
		INSERT INTO autonomous_plans (
			org_id, plan_id, version, parent_plan_id, correlation_id,
			goal, module, related_entity_type, related_entity_id,
			current_state_summary, status, autonomy_level, policy_decision, policy_reason,
			triggering_event, execution_status, current_step_id, confidence_score,
			created_at, updated_at
		) VALUES (
			?, ?, 1, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			NOW(), NOW()
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		wf.OrgID, wf.WorkflowID, wf.ParentWorkflowID, wf.CorrelationID,
		wf.Objective, string(wf.WorkflowType), wf.RelatedEntityType, wf.RelatedEntityID,
		wf.Objective, string(wf.CurrentState), wf.AutonomyLevel, wf.PolicyDecision, wf.PolicyReason,
		wf.InitiatingEvent, string(wf.CurrentState), wf.CurrentStep, wf.Confidence,
	)
	if err != nil {
		return fmt.Errorf("failed to create autonomous plan record: %w", err)
	}

	if len(wf.Steps) > 0 {
		if err := r.SaveWorkflowSteps(ctx, wf.OrgID, wf.WorkflowID, wf.Steps); err != nil {
			return fmt.Errorf("failed to save workflow steps: %w", err)
		}
	}

	return nil
}

func (r *mysqlRepository) GetWorkflow(ctx context.Context, orgID int64, workflowID string) (*EnterpriseWorkflow, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	query := `
		SELECT id, org_id, plan_id, version, parent_plan_id, correlation_id,
		       goal, module, related_entity_type, related_entity_id,
		       status, autonomy_level, policy_decision, policy_reason,
		       triggering_event, execution_status, current_step_id,
		       confidence_score, escalation_reason,
		       stopped_by_user_id, stopped_at, stop_reason,
		       created_at, updated_at
		FROM autonomous_plans
		WHERE org_id = ? AND plan_id = ?
		LIMIT 1
	`
	var row struct {
		ID                int64          `db:"id"`
		OrgID             int64          `db:"org_id"`
		PlanID            string         `db:"plan_id"`
		Version           int            `db:"version"`
		ParentPlanID      sql.NullString `db:"parent_plan_id"`
		CorrelationID     string         `db:"correlation_id"`
		Goal              string         `db:"goal"`
		Module            string         `db:"module"`
		RelatedEntityType string         `db:"related_entity_type"`
		RelatedEntityID   string         `db:"related_entity_id"`
		Status            string         `db:"status"`
		AutonomyLevel     string         `db:"autonomy_level"`
		PolicyDecision    string         `db:"policy_decision"`
		PolicyReason      sql.NullString `db:"policy_reason"`
		TriggeringEvent   sql.NullString `db:"triggering_event"`
		ExecutionStatus   string         `db:"execution_status"`
		CurrentStepID     sql.NullString `db:"current_step_id"`
		ConfidenceScore   float64        `db:"confidence_score"`
		EscalationReason  sql.NullString `db:"escalation_reason"`
		StoppedByUserID   sql.NullInt64  `db:"stopped_by_user_id"`
		StoppedAt         *time.Time     `db:"stopped_at"`
		StopReason        sql.NullString `db:"stop_reason"`
		CreatedAt         time.Time      `db:"created_at"`
		UpdatedAt         time.Time      `db:"updated_at"`
	}

	if err := r.db.GetContext(ctx, &row, query, orgID, workflowID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWorkflowNotFound
		}
		return nil, fmt.Errorf("failed getting workflow %s: %w", workflowID, err)
	}

	wf := &EnterpriseWorkflow{
		ID:                row.ID,
		WorkflowID:        row.PlanID,
		OrgID:             row.OrgID,
		WorkflowType:      EnterpriseWorkflowType(row.Module),
		Objective:         row.Goal,
		CorrelationID:     row.CorrelationID,
		CurrentState:      EnterpriseWorkflowState(row.Status),
		AutonomyLevel:     row.AutonomyLevel,
		PolicyDecision:    row.PolicyDecision,
		RelatedEntityType: row.RelatedEntityType,
		RelatedEntityID:   row.RelatedEntityID,
		Confidence:        row.ConfidenceScore,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}

	if row.ParentPlanID.Valid {
		wf.ParentWorkflowID = &row.ParentPlanID.String
	}
	if row.PolicyReason.Valid {
		wf.PolicyReason = &row.PolicyReason.String
	}
	if row.TriggeringEvent.Valid {
		wf.InitiatingEvent = row.TriggeringEvent.String
	}
	if row.CurrentStepID.Valid {
		wf.CurrentStep = row.CurrentStepID.String
	}
	if row.EscalationReason.Valid {
		wf.ErrorMessage = &row.EscalationReason.String
	}
	if row.StoppedByUserID.Valid {
		wf.StoppedByUserID = &row.StoppedByUserID.Int64
	}
	wf.StoppedAt = row.StoppedAt
	if row.StopReason.Valid {
		wf.StopReason = &row.StopReason.String
	}

	// Fetch Steps
	steps, err := r.GetWorkflowSteps(ctx, orgID, workflowID)
	if err == nil {
		wf.Steps = steps
		var agents []string
		for _, s := range steps {
			if s.AgentID != "" {
				found := false
				for _, a := range agents {
					if a == s.AgentID {
						found = true
						break
					}
				}
				if !found {
					agents = append(agents, s.AgentID)
				}
			}
		}
		wf.AssignedAgents = agents
	}

	return wf, nil
}

func (r *mysqlRepository) UpdateWorkflowState(ctx context.Context, orgID int64, workflowID string, state EnterpriseWorkflowState, errCode, errMsg *string) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}

	query := `
		UPDATE autonomous_plans
		SET status = ?,
		    execution_status = ?,
		    escalation_reason = COALESCE(?, escalation_reason),
		    updated_at = NOW()
		WHERE org_id = ? AND plan_id = ?
	`
	res, err := r.db.ExecContext(ctx, query, string(state), string(state), errMsg, orgID, workflowID)
	if err != nil {
		return fmt.Errorf("failed to update workflow state: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrWorkflowNotFound
	}
	return nil
}

func (r *mysqlRepository) UpdateWorkflowStop(ctx context.Context, orgID int64, workflowID string, stoppedByUserID *int64, stopReason string) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}

	query := `
		UPDATE autonomous_plans
		SET status = 'PAUSED',
		    execution_status = 'STOPPED',
		    stopped_by_user_id = ?,
		    stopped_at = NOW(),
		    stop_reason = ?,
		    updated_at = NOW()
		WHERE org_id = ? AND plan_id = ?
	`
	res, err := r.db.ExecContext(ctx, query, stoppedByUserID, stopReason, orgID, workflowID)
	if err != nil {
		return fmt.Errorf("failed to update workflow stop: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrWorkflowNotFound
	}
	return nil
}

func (r *mysqlRepository) ListWorkflows(ctx context.Context, filter WorkflowFilter) ([]*EnterpriseWorkflow, int, error) {
	if filter.OrgID <= 0 {
		return nil, 0, ErrUnauthorizedTenant
	}

	whereClauses := []string{"org_id = ?"}
	args := []interface{}{filter.OrgID}

	if filter.WorkflowType != "" {
		whereClauses = append(whereClauses, "module = ?")
		args = append(args, filter.WorkflowType)
	}
	if filter.State != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, filter.State)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM autonomous_plans WHERE %s", whereSQL)
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed counting workflows: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, org_id, plan_id, version, parent_plan_id, correlation_id,
		       goal, module, related_entity_type, related_entity_id,
		       status, autonomy_level, policy_decision, policy_reason,
		       triggering_event, execution_status, current_step_id,
		       confidence_score, escalation_reason,
		       stopped_by_user_id, stopped_at, stop_reason,
		       created_at, updated_at
		FROM autonomous_plans
		WHERE %s
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, whereSQL)

	argsWithPaging := append(args, limit, offset)

	var rows []struct {
		ID                int64          `db:"id"`
		OrgID             int64          `db:"org_id"`
		PlanID            string         `db:"plan_id"`
		Version           int            `db:"version"`
		ParentPlanID      sql.NullString `db:"parent_plan_id"`
		CorrelationID     string         `db:"correlation_id"`
		Goal              string         `db:"goal"`
		Module            string         `db:"module"`
		RelatedEntityType string         `db:"related_entity_type"`
		RelatedEntityID   string         `db:"related_entity_id"`
		Status            string         `db:"status"`
		AutonomyLevel     string         `db:"autonomy_level"`
		PolicyDecision    string         `db:"policy_decision"`
		PolicyReason      sql.NullString `db:"policy_reason"`
		TriggeringEvent   sql.NullString `db:"triggering_event"`
		ExecutionStatus   string         `db:"execution_status"`
		CurrentStepID     sql.NullString `db:"current_step_id"`
		ConfidenceScore   float64        `db:"confidence_score"`
		EscalationReason  sql.NullString `db:"escalation_reason"`
		StoppedByUserID   sql.NullInt64  `db:"stopped_by_user_id"`
		StoppedAt         *time.Time     `db:"stopped_at"`
		StopReason        sql.NullString `db:"stop_reason"`
		CreatedAt         time.Time      `db:"created_at"`
		UpdatedAt         time.Time      `db:"updated_at"`
	}

	if err := r.db.SelectContext(ctx, &rows, dataQuery, argsWithPaging...); err != nil {
		return nil, 0, fmt.Errorf("failed selecting workflows: %w", err)
	}

	workflows := make([]*EnterpriseWorkflow, 0, len(rows))
	for _, rRow := range rows {
		wf := &EnterpriseWorkflow{
			ID:                rRow.ID,
			WorkflowID:        rRow.PlanID,
			OrgID:             rRow.OrgID,
			WorkflowType:      EnterpriseWorkflowType(rRow.Module),
			Objective:         rRow.Goal,
			CorrelationID:     rRow.CorrelationID,
			CurrentState:      EnterpriseWorkflowState(rRow.Status),
			AutonomyLevel:     rRow.AutonomyLevel,
			PolicyDecision:    rRow.PolicyDecision,
			RelatedEntityType: rRow.RelatedEntityType,
			RelatedEntityID:   rRow.RelatedEntityID,
			Confidence:        rRow.ConfidenceScore,
			CreatedAt:         rRow.CreatedAt,
			UpdatedAt:         rRow.UpdatedAt,
		}
		if rRow.ParentPlanID.Valid {
			wf.ParentWorkflowID = &rRow.ParentPlanID.String
		}
		if rRow.TriggeringEvent.Valid {
			wf.InitiatingEvent = rRow.TriggeringEvent.String
		}
		if rRow.CurrentStepID.Valid {
			wf.CurrentStep = rRow.CurrentStepID.String
		}
		workflows = append(workflows, wf)
	}

	return workflows, total, nil
}

func (r *mysqlRepository) SaveWorkflowSteps(ctx context.Context, orgID int64, workflowID string, steps []EnterpriseWorkflowStep) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}

	for i, step := range steps {
		stepNum := i + 1
		if step.StepNumber > 0 {
			stepNum = step.StepNumber
		}
		stepID := step.StepID
		if stepID == "" {
			stepID = fmt.Sprintf("step-%d", stepNum)
		}

		paramsJSON, _ := json.Marshal(step.Parameters)
		depsJSON, _ := json.Marshal(step.Dependencies)
		if step.IdempotencyKey == "" {
			step.IdempotencyKey = fmt.Sprintf("idemp-%s-%s", workflowID, stepID)
		}
		if step.Status == "" {
			step.Status = "PENDING"
		}
		if step.RiskLevel == "" {
			step.RiskLevel = "LOW"
		}

		query := `
			INSERT INTO autonomous_plan_steps (
				plan_id, org_id, step_number, step_id, action_type,
				title, description, parameters, dependencies, expected_outcome,
				risk_level, requires_approval, action_system_action_id, idempotency_key,
				status, execution_attempt, max_attempts, created_at, updated_at
			) VALUES (
				?, ?, ?, ?, ?,
				?, ?, ?, ?, ?,
				?, ?, ?, ?,
				?, 0, 3, NOW(), NOW()
			)
			ON DUPLICATE KEY UPDATE
				action_type = VALUES(action_type),
				title = VALUES(title),
				description = VALUES(description),
				parameters = VALUES(parameters),
				dependencies = VALUES(dependencies),
				expected_outcome = VALUES(expected_outcome),
				risk_level = VALUES(risk_level),
				requires_approval = VALUES(requires_approval),
				status = VALUES(status),
				updated_at = NOW()
		`
		_, err := r.db.ExecContext(ctx, query,
			workflowID, orgID, stepNum, stepID, step.ActionType,
			step.Title, step.Description, string(paramsJSON), string(depsJSON), step.ExpectedOutcome,
			step.RiskLevel, step.RequiresApproval, step.ActionSystemActionID, step.IdempotencyKey,
			step.Status,
		)
		if err != nil {
			return fmt.Errorf("failed saving step %s for plan %s: %w", stepID, workflowID, err)
		}
	}
	return nil
}

func (r *mysqlRepository) GetWorkflowSteps(ctx context.Context, orgID int64, workflowID string) ([]EnterpriseWorkflowStep, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	query := `
		SELECT id, plan_id, org_id, step_number, step_id, action_type,
		       title, description, parameters, dependencies, expected_outcome,
		       risk_level, requires_approval, action_system_action_id, idempotency_key,
		       status, execution_attempt, max_attempts, executed_at, execution_result,
		       error_message, created_at, updated_at
		FROM autonomous_plan_steps
		WHERE org_id = ? AND plan_id = ?
		ORDER BY step_number ASC
	`
	var rows []struct {
		ID                   int64          `db:"id"`
		PlanID               string         `db:"plan_id"`
		OrgID                int64          `db:"org_id"`
		StepNumber           int            `db:"step_number"`
		StepID               string         `db:"step_id"`
		ActionType           string         `db:"action_type"`
		Title                string         `db:"title"`
		Description          string         `db:"description"`
		Parameters           string         `db:"parameters"`
		Dependencies         sql.NullString `db:"dependencies"`
		ExpectedOutcome      string         `db:"expected_outcome"`
		RiskLevel            string         `db:"risk_level"`
		RequiresApproval     bool           `db:"requires_approval"`
		ActionSystemActionID sql.NullString `db:"action_system_action_id"`
		IdempotencyKey       string         `db:"idempotency_key"`
		Status               string         `db:"status"`
		ExecutionAttempt     int            `db:"execution_attempt"`
		MaxAttempts          int            `db:"max_attempts"`
		ExecutedAt           *time.Time     `db:"executed_at"`
		ExecutionResult      sql.NullString `db:"execution_result"`
		ErrorMessage         sql.NullString `db:"error_message"`
		CreatedAt            time.Time      `db:"created_at"`
		UpdatedAt            time.Time      `db:"updated_at"`
	}

	if err := r.db.SelectContext(ctx, &rows, query, orgID, workflowID); err != nil {
		return nil, fmt.Errorf("failed querying steps for workflow %s: %w", workflowID, err)
	}

	steps := make([]EnterpriseWorkflowStep, 0, len(rows))
	for _, rRow := range rows {
		st := EnterpriseWorkflowStep{
			ID:               rRow.ID,
			StepID:           rRow.StepID,
			WorkflowID:       rRow.PlanID,
			OrgID:            rRow.OrgID,
			StepNumber:       rRow.StepNumber,
			ActionType:       rRow.ActionType,
			Title:            rRow.Title,
			Description:      rRow.Description,
			ExpectedOutcome:  rRow.ExpectedOutcome,
			RiskLevel:        rRow.RiskLevel,
			RequiresApproval: rRow.RequiresApproval,
			IdempotencyKey:   rRow.IdempotencyKey,
			Status:           rRow.Status,
			ExecutionAttempt: rRow.ExecutionAttempt,
			MaxAttempts:      rRow.MaxAttempts,
			ExecutedAt:       rRow.ExecutedAt,
			CreatedAt:        rRow.CreatedAt,
			UpdatedAt:        rRow.UpdatedAt,
		}
		if rRow.ActionSystemActionID.Valid {
			st.ActionSystemActionID = &rRow.ActionSystemActionID.String
		}
		if rRow.ErrorMessage.Valid {
			st.ErrorMessage = &rRow.ErrorMessage.String
		}
		if rRow.Parameters != "" {
			_ = json.Unmarshal([]byte(rRow.Parameters), &st.Parameters)
		}
		if rRow.Dependencies.Valid && rRow.Dependencies.String != "" {
			_ = json.Unmarshal([]byte(rRow.Dependencies.String), &st.Dependencies)
		}
		if rRow.ExecutionResult.Valid && rRow.ExecutionResult.String != "" {
			_ = json.Unmarshal([]byte(rRow.ExecutionResult.String), &st.ExecutionResult)
		}
		steps = append(steps, st)
	}

	return steps, nil
}

func (r *mysqlRepository) UpdateWorkflowStep(ctx context.Context, orgID int64, step *EnterpriseWorkflowStep) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}

	resJSON, _ := json.Marshal(step.ExecutionResult)

	query := `
		UPDATE autonomous_plan_steps
		SET status = ?,
		    execution_attempt = ?,
		    executed_at = ?,
		    execution_result = ?,
		    error_message = ?,
		    action_system_action_id = ?,
		    updated_at = NOW()
		WHERE org_id = ? AND plan_id = ? AND step_id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		step.Status, step.ExecutionAttempt, step.ExecutedAt, string(resJSON),
		step.ErrorMessage, step.ActionSystemActionID,
		orgID, step.WorkflowID, step.StepID,
	)
	if err != nil {
		return fmt.Errorf("failed updating step %s: %w", step.StepID, err)
	}
	return nil
}

func (r *mysqlRepository) GetActiveOrInterruptedWorkflows(ctx context.Context) ([]*EnterpriseWorkflow, error) {
	query := `
		SELECT id, org_id, plan_id, version, parent_plan_id, correlation_id,
		       goal, module, related_entity_type, related_entity_id,
		       status, autonomy_level, policy_decision, policy_reason,
		       triggering_event, execution_status, current_step_id,
		       confidence_score, escalation_reason,
		       stopped_by_user_id, stopped_at, stop_reason,
		       created_at, updated_at
		FROM autonomous_plans
		WHERE status IN ('PENDING', 'RUNNING', 'WAITING', 'WAITING_FOR_APPROVAL')
		ORDER BY id ASC
	`
	var rows []struct {
		ID                int64          `db:"id"`
		OrgID             int64          `db:"org_id"`
		PlanID            string         `db:"plan_id"`
		Version           int            `db:"version"`
		ParentPlanID      sql.NullString `db:"parent_plan_id"`
		CorrelationID     string         `db:"correlation_id"`
		Goal              string         `db:"goal"`
		Module            string         `db:"module"`
		RelatedEntityType string         `db:"related_entity_type"`
		RelatedEntityID   string         `db:"related_entity_id"`
		Status            string         `db:"status"`
		AutonomyLevel     string         `db:"autonomy_level"`
		PolicyDecision    string         `db:"policy_decision"`
		PolicyReason      sql.NullString `db:"policy_reason"`
		TriggeringEvent   sql.NullString `db:"triggering_event"`
		ExecutionStatus   string         `db:"execution_status"`
		CurrentStepID     sql.NullString `db:"current_step_id"`
		ConfidenceScore   float64        `db:"confidence_score"`
		EscalationReason  sql.NullString `db:"escalation_reason"`
		StoppedByUserID   sql.NullInt64  `db:"stopped_by_user_id"`
		StoppedAt         *time.Time     `db:"stopped_at"`
		StopReason        sql.NullString `db:"stop_reason"`
		CreatedAt         time.Time      `db:"created_at"`
		UpdatedAt         time.Time      `db:"updated_at"`
	}

	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("failed selecting interrupted workflows: %w", err)
	}

	workflows := make([]*EnterpriseWorkflow, 0, len(rows))
	for _, rRow := range rows {
		wf := &EnterpriseWorkflow{
			ID:                rRow.ID,
			WorkflowID:        rRow.PlanID,
			OrgID:             rRow.OrgID,
			WorkflowType:      EnterpriseWorkflowType(rRow.Module),
			Objective:         rRow.Goal,
			CorrelationID:     rRow.CorrelationID,
			CurrentState:      EnterpriseWorkflowState(rRow.Status),
			AutonomyLevel:     rRow.AutonomyLevel,
			PolicyDecision:    rRow.PolicyDecision,
			RelatedEntityType: rRow.RelatedEntityType,
			RelatedEntityID:   rRow.RelatedEntityID,
			Confidence:        rRow.ConfidenceScore,
			CreatedAt:         rRow.CreatedAt,
			UpdatedAt:         rRow.UpdatedAt,
		}
		if rRow.ParentPlanID.Valid {
			wf.ParentWorkflowID = &rRow.ParentPlanID.String
		}
		if rRow.CurrentStepID.Valid {
			wf.CurrentStep = rRow.CurrentStepID.String
		}
		workflows = append(workflows, wf)
	}

	return workflows, nil
}

func (r *mysqlRepository) GetPolicyContext(ctx context.Context, orgID int64, module string) (*EnterprisePolicyContext, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	query := `
		SELECT org_id, module, autonomy_level, allowed_action_types, prohibited_action_types,
		       requires_approval, max_monetary_threshold, min_confidence_threshold, emergency_stop
		FROM autonomy_policies
		WHERE org_id = ? AND (module = ? OR module = 'CROSS_MODULE')
		ORDER BY (module = ?) DESC
		LIMIT 1
	`
	var row struct {
		OrgID                  int64          `db:"org_id"`
		Module                 string         `db:"module"`
		AutonomyLevel          string         `db:"autonomy_level"`
		AllowedActionTypes     sql.NullString `db:"allowed_action_types"`
		ProhibitedActionTypes  sql.NullString `db:"prohibited_action_types"`
		RequiresApproval       bool           `db:"requires_approval"`
		MaxMonetaryThreshold   float64        `db:"max_monetary_threshold"`
		MinConfidenceThreshold float64        `db:"min_confidence_threshold"`
		EmergencyStop          bool           `db:"emergency_stop"`
	}

	err := r.db.GetContext(ctx, &row, query, orgID, module, module)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Baseline safe defaults
			return &EnterprisePolicyContext{
				OrgID:                  orgID,
				AutonomyLevel:          "LEVEL_3_CONTROLLED_EXECUTION",
				AllowedWorkflowTypes:   []string{"SHIPMENT_RECOVERY", "COMMERCIAL_CYCLE", "FINANCIAL_COLLECTION", "CROSS_MODULE_RISK", "ADAPTIVE_REPLANNING", "COMPLIANCE_AUDIT", "OPERATIONAL_RECOVERY", "CUSTOMER_RELATIONSHIP", "REVENUE_OPTIMIZATION", "CONTRACT_COMPLIANCE_RISK", "SHIPMENT_MONITORING", "*"},
				AllowedActionTypes:     []string{"TAG_INTERNAL_STATE", "LOG_OBSERVATION", "CARRIER_TELEMETRY_REFRESH", "SEND_SHIPMENT_STATUS_UPDATE", "SEND_ALERT"},
				ProhibitedActionTypes:  []string{"OVERRIDE_SECURITY_POLICY", "BYPASS_APPROVAL_GATE"},
				RequiresApproval:       false,
				MaxMonetaryThreshold:   2500.0,
				MinConfidenceThreshold: 0.75,
				EmergencyStopActive:    false,
			}, nil
		}
		return nil, fmt.Errorf("failed fetching policy context: %w", err)
	}

	pol := &EnterprisePolicyContext{
		OrgID:                  row.OrgID,
		AutonomyLevel:          row.AutonomyLevel,
		RequiresApproval:       row.RequiresApproval,
		MaxMonetaryThreshold:   row.MaxMonetaryThreshold,
		MinConfidenceThreshold: row.MinConfidenceThreshold,
		EmergencyStopActive:    row.EmergencyStop,
	}
	if row.AllowedActionTypes.Valid && row.AllowedActionTypes.String != "" {
		_ = json.Unmarshal([]byte(row.AllowedActionTypes.String), &pol.AllowedActionTypes)
	}
	if row.ProhibitedActionTypes.Valid && row.ProhibitedActionTypes.String != "" {
		_ = json.Unmarshal([]byte(row.ProhibitedActionTypes.String), &pol.ProhibitedActionTypes)
	}

	return pol, nil
}

func (r *mysqlRepository) UpdatePolicyContext(ctx context.Context, orgID int64, policy *EnterprisePolicyContext) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}

	allowedJSON, _ := json.Marshal(policy.AllowedActionTypes)
	prohibitedJSON, _ := json.Marshal(policy.ProhibitedActionTypes)

	query := `
		INSERT INTO autonomy_policies (
			org_id, module, autonomy_level, allowed_action_types, prohibited_action_types,
			requires_approval, max_monetary_threshold, min_confidence_threshold, emergency_stop,
			is_active, policy_version, created_at, updated_at
		) VALUES (
			?, 'CROSS_MODULE', ?, ?, ?,
			?, ?, ?, ?,
			1, 1, NOW(), NOW()
		)
		ON DUPLICATE KEY UPDATE
			autonomy_level = VALUES(autonomy_level),
			allowed_action_types = VALUES(allowed_action_types),
			prohibited_action_types = VALUES(prohibited_action_types),
			requires_approval = VALUES(requires_approval),
			max_monetary_threshold = VALUES(max_monetary_threshold),
			min_confidence_threshold = VALUES(min_confidence_threshold),
			emergency_stop = VALUES(emergency_stop),
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		orgID, policy.AutonomyLevel, string(allowedJSON), string(prohibitedJSON),
		policy.RequiresApproval, policy.MaxMonetaryThreshold, policy.MinConfidenceThreshold,
		policy.EmergencyStopActive,
	)
	if err != nil {
		return fmt.Errorf("failed updating policy context: %w", err)
	}
	return nil
}

func (r *mysqlRepository) CheckAndRecordEventDedup(ctx context.Context, orgID int64, dedupKey, eventType string) (bool, error) {
	if orgID <= 0 {
		return false, ErrUnauthorizedTenant
	}
	if dedupKey == "" {
		return true, nil
	}

	query := `
		INSERT INTO action_idempotency_keys (
			org_id, idempotency_key, action_name, status, created_at, expires_at
		) VALUES (
			?, ?, ?, 'IN_PROGRESS', NOW(), DATE_ADD(NOW(), INTERVAL 24 HOUR)
		)
	`
	_, err := r.db.ExecContext(ctx, query, orgID, dedupKey, eventType)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return false, nil // Duplicate detected!
		}
		return false, fmt.Errorf("failed recording event deduplication key: %w", err)
	}
	return true, nil
}

func (r *mysqlRepository) GetPlatformOverview(ctx context.Context, orgID int64) (*EnterprisePlatformOverview, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	overview := &EnterprisePlatformOverview{
		OrgID:          orgID,
		PlatformHealth: "HEALTHY",
	}

	// 1. Query status counts
	queryCounts := `
		SELECT status, COUNT(*) as count
		FROM autonomous_plans
		WHERE org_id = ?
		GROUP BY status
	`
	var rows []struct {
		Status string `db:"status"`
		Count  int    `db:"count"`
	}
	if err := r.db.SelectContext(ctx, &rows, queryCounts, orgID); err == nil {
		for _, rRow := range rows {
			switch rRow.Status {
			case "RUNNING", "IN_PROGRESS":
				overview.ActiveWorkflows += rRow.Count
			case "PAUSED":
				overview.PausedWorkflows += rRow.Count
			case "WAITING_FOR_APPROVAL", "REQUIRES_APPROVAL":
				overview.WaitingApprovals += rRow.Count
			case "ESCALATED":
				overview.EscalatedWorkflows += rRow.Count
			case "FAILED":
				overview.FailedWorkflows += rRow.Count
			case "COMPLETED":
				overview.CompletedWorkflows += rRow.Count
			}
		}
	}

	// 2. Query autonomy policy
	policy, err := r.GetPolicyContext(ctx, orgID, "CROSS_MODULE")
	if err == nil && policy != nil {
		overview.AutonomyLevel = policy.AutonomyLevel
		overview.EmergencyStopActive = policy.EmergencyStopActive
		if policy.EmergencyStopActive {
			overview.PlatformHealth = "EMERGENCY_HALT"
		} else if overview.FailedWorkflows > 5 || overview.EscalatedWorkflows > 5 {
			overview.PlatformHealth = "DEGRADED"
		}
	}

	// 3. Query recent workflows
	recent, _, err := r.ListWorkflows(ctx, WorkflowFilter{OrgID: orgID, Limit: 5})
	if err == nil {
		overview.RecentWorkflows = recent
	}

	return overview, nil
}
