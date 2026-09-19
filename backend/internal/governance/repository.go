package governance

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetPolicy(ctx context.Context, orgID int64) (*GovernancePolicy, error)
	UpsertPolicy(ctx context.Context, orgID int64, policy *GovernancePolicy) error
	GetKillSwitches(ctx context.Context, orgID int64) ([]KillSwitchRecord, error)
	SetKillSwitch(ctx context.Context, orgID int64, scope, target string, isKilled bool, reason string, userID *int64) error
	IsTargetKilled(ctx context.Context, orgID int64, workflowName, modelName, actionType string) (bool, string, error)
	RecordSafetyViolation(ctx context.Context, v *SafetyViolationRecord) error
	ListSafetyViolations(ctx context.Context, orgID int64, limit, offset int) ([]SafetyViolationRecord, int, error)
	GetOverview(ctx context.Context, orgID int64) (*GovernanceOverviewDTO, error)
}

type repositoryImpl struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) GetPolicy(ctx context.Context, orgID int64) (*GovernancePolicy, error) {
	// First check tenant-specific policy
	query := `
		SELECT id, org_id, allowed_workflows, allowed_models, allowed_providers, allowed_actions,
		       max_input_chars, max_output_chars, timeout_seconds, max_retries, rate_limit_rpm,
		       enforce_prompt_injection_check, enforce_sensitive_data_redaction, enforce_hitl_approvals,
		       is_active, created_at, updated_at
		FROM ai_governance_policies
		WHERE org_id = ? AND is_active = 1
		LIMIT 1
	`
	var p GovernancePolicy
	err := r.db.GetContext(ctx, &p, query, orgID)
	if err == nil {
		p.UnpackJSON()
		return &p, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query tenant governance policy: %w", err)
	}

	// Fallback to system-wide default policy (org_id = 0)
	err = r.db.GetContext(ctx, &p, query, 0)
	if err == nil {
		p.UnpackJSON()
		return &p, nil
	}

	// In-memory safe default fallback if DB has no row
	defaultWorkflows := []string{"pricing", "sales", "operations", "contracts", "compliance", "finance", "reporting", "copilot", "notifications", "outreach"}
	defaultModels := []string{"gemini-1.5-pro", "gemini-1.5-flash", "gpt-4o", "gpt-4o-mini", "claude-3-5-sonnet", "deterministic_test"}
	defaultProviders := []string{"gemini", "openai", "anthropic", "deterministic_test"}
	defaultActions := []string{"CREATE_RECOMMENDATION", "CREATE_INTERNAL_TASK", "REQUEST_HUMAN_APPROVAL", "CALCULATE_FORECAST", "GENERATE_DRAFT", "NOTIFY_OPERATOR"}

	return &GovernancePolicy{
		OrgID:                         orgID,
		AllowedWorkflows:              defaultWorkflows,
		AllowedModels:                 defaultModels,
		AllowedProviders:              defaultProviders,
		AllowedActions:                defaultActions,
		MaxInputChars:                 100000,
		MaxOutputChars:                50000,
		TimeoutSeconds:                30,
		MaxRetries:                    3,
		RateLimitRPM:                  120,
		EnforcePromptInjectionCheck:   true,
		EnforceSensitiveDataRedaction: true,
		EnforceHITLApprovals:          true,
		IsActive:                      true,
		CreatedAt:                     time.Now(),
		UpdatedAt:                     time.Now(),
	}, nil
}

func (r *repositoryImpl) UpsertPolicy(ctx context.Context, orgID int64, policy *GovernancePolicy) error {
	wfJSON, _ := json.Marshal(policy.AllowedWorkflows)
	modelsJSON, _ := json.Marshal(policy.AllowedModels)
	provJSON, _ := json.Marshal(policy.AllowedProviders)
	actJSON, _ := json.Marshal(policy.AllowedActions)

	query := `
		INSERT INTO ai_governance_policies (
			org_id, allowed_workflows, allowed_models, allowed_providers, allowed_actions,
			max_input_chars, max_output_chars, timeout_seconds, max_retries, rate_limit_rpm,
			enforce_prompt_injection_check, enforce_sensitive_data_redaction, enforce_hitl_approvals,
			is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			allowed_workflows = VALUES(allowed_workflows),
			allowed_models = VALUES(allowed_models),
			allowed_providers = VALUES(allowed_providers),
			allowed_actions = VALUES(allowed_actions),
			max_input_chars = VALUES(max_input_chars),
			max_output_chars = VALUES(max_output_chars),
			timeout_seconds = VALUES(timeout_seconds),
			max_retries = VALUES(max_retries),
			rate_limit_rpm = VALUES(rate_limit_rpm),
			enforce_prompt_injection_check = VALUES(enforce_prompt_injection_check),
			enforce_sensitive_data_redaction = VALUES(enforce_sensitive_data_redaction),
			enforce_hitl_approvals = VALUES(enforce_hitl_approvals),
			is_active = VALUES(is_active),
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(ctx, query,
		orgID, string(wfJSON), string(modelsJSON), string(provJSON), string(actJSON),
		policy.MaxInputChars, policy.MaxOutputChars, policy.TimeoutSeconds, policy.MaxRetries, policy.RateLimitRPM,
		policy.EnforcePromptInjectionCheck, policy.EnforceSensitiveDataRedaction, policy.EnforceHITLApprovals,
		policy.IsActive,
	)
	return err
}

func (r *repositoryImpl) GetKillSwitches(ctx context.Context, orgID int64) ([]KillSwitchRecord, error) {
	query := `
		SELECT id, org_id, scope, target_identifier, is_killed, reason, killed_by_user_id, created_at, updated_at
		FROM ai_governance_kill_switches
		WHERE org_id = ? OR org_id = 0
		ORDER BY is_killed DESC, updated_at DESC
	`
	var switches []KillSwitchRecord
	err := r.db.SelectContext(ctx, &switches, query, orgID)
	return switches, err
}

func (r *repositoryImpl) SetKillSwitch(ctx context.Context, orgID int64, scope, target string, isKilled bool, reason string, userID *int64) error {
	query := `
		INSERT INTO ai_governance_kill_switches (org_id, scope, target_identifier, is_killed, reason, killed_by_user_id)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			is_killed = VALUES(is_killed),
			reason = VALUES(reason),
			killed_by_user_id = VALUES(killed_by_user_id),
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(ctx, query, orgID, scope, target, isKilled, reason, userID)
	return err
}

func (r *repositoryImpl) IsTargetKilled(ctx context.Context, orgID int64, workflowName, modelName, actionType string) (bool, string, error) {
	// Checks if global kill switch is on, or specific workflow / model / action is killed
	query := `
		SELECT reason FROM ai_governance_kill_switches
		WHERE is_killed = 1 AND (org_id = ? OR org_id = 0) AND (
			(scope = 'GLOBAL' AND target_identifier = 'GLOBAL_AI') OR
			(scope = 'WORKFLOW' AND target_identifier = ?) OR
			(scope = 'MODEL' AND target_identifier = ?) OR
			(scope = 'ACTION' AND target_identifier = ?)
		)
		LIMIT 1
	`
	var reason string
	err := r.db.GetContext(ctx, &reason, query, orgID, "WORKFLOW:"+workflowName, "MODEL:"+modelName, "ACTION:"+actionType)
	if err == nil {
		return true, reason, nil
	}
	if err == sql.ErrNoRows {
		return false, "", nil
	}
	return false, "", err
}

func (r *repositoryImpl) RecordSafetyViolation(ctx context.Context, v *SafetyViolationRecord) error {
	query := `
		INSERT INTO ai_safety_violations (
			org_id, user_id, workflow_name, violation_type, severity, action_taken, details, input_snippet_redacted, correlation_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		v.OrgID, v.UserID, v.WorkflowName, v.ViolationType, v.Severity, v.ActionTaken, v.Details, v.InputSnippetRedacted, v.CorrelationID,
	)
	return err
}

func (r *repositoryImpl) ListSafetyViolations(ctx context.Context, orgID int64, limit, offset int) ([]SafetyViolationRecord, int, error) {
	countQuery := `SELECT COUNT(*) FROM ai_safety_violations WHERE org_id = ?`
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, orgID); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, org_id, user_id, workflow_name, violation_type, severity, action_taken, details, input_snippet_redacted, correlation_id, created_at
		FROM ai_safety_violations
		WHERE org_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	var items []SafetyViolationRecord
	err := r.db.SelectContext(ctx, &items, query, orgID, limit, offset)
	return items, total, err
}

func (r *repositoryImpl) GetOverview(ctx context.Context, orgID int64) (*GovernanceOverviewDTO, error) {
	policy, err := r.GetPolicy(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Active kill switches
	var activeKills int
	_ = r.db.GetContext(ctx, &activeKills, `SELECT COUNT(*) FROM ai_governance_kill_switches WHERE is_killed = 1 AND (org_id = ? OR org_id = 0)`, orgID)

	// Global kill switch status
	var globalKillCount int
	_ = r.db.GetContext(ctx, &globalKillCount, `SELECT COUNT(*) FROM ai_governance_kill_switches WHERE is_killed = 1 AND scope = 'GLOBAL' AND (org_id = ? OR org_id = 0)`, orgID)
	isGlobalKilled := globalKillCount > 0

	// 24h violations count
	var violations24h int
	_ = r.db.GetContext(ctx, &violations24h, `SELECT COUNT(*) FROM ai_safety_violations WHERE org_id = ? AND created_at >= NOW() - INTERVAL 24 HOUR`, orgID)

	// Recent violations
	recent, _, _ := r.ListSafetyViolations(ctx, orgID, 5, 0)

	sysStatus := "HEALTHY"
	readiness := "PRODUCTION_READY"
	if isGlobalKilled {
		sysStatus = "BLOCKED"
		readiness = "BLOCKED"
	} else if activeKills > 0 || violations24h > 10 {
		sysStatus = "DEGRADED"
		readiness = "REVIEW_REQUIRED"
	}

	return &GovernanceOverviewDTO{
		SystemStatus:        sysStatus,
		GlobalKillSwitch:    isGlobalKilled,
		ActiveKillSwitches:  activeKills,
		AllowedWorkflows:    policy.AllowedWorkflows,
		AllowedProviders:    policy.AllowedProviders,
		AllowedModels:       policy.AllowedModels,
		Violations24h:       violations24h,
		RecentViolations:    recent,
		HITLApprovalsActive: policy.EnforceHITLApprovals,
		ProductionReadiness: readiness,
	}, nil
}
