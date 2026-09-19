package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	CreateProposal(ctx context.Context, p *ActionProposal) error
	GetProposalByID(ctx context.Context, orgID int64, id int64) (*ActionProposal, error)
	GetProposalByProposalID(ctx context.Context, orgID int64, proposalID string) (*ActionProposal, error)
	UpdateProposal(ctx context.Context, p *ActionProposal) error
	ListProposals(ctx context.Context, orgID int64, filter ProposalFilter) ([]*ActionProposal, int64, error)

	CreateExecution(ctx context.Context, exec *ActionExecution) error
	GetExecutionByID(ctx context.Context, orgID int64, id int64) (*ActionExecution, error)
	GetExecutionByIdempotencyKey(ctx context.Context, orgID int64, idempotencyKey string) (*ActionExecution, error)
	UpdateExecution(ctx context.Context, exec *ActionExecution) error
	ListExecutions(ctx context.Context, orgID int64, filter ExecutionFilter) ([]*ActionExecution, int64, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateProposal(ctx context.Context, p *ActionProposal) error {
	query := `INSERT INTO ai_action_proposals
		(org_id, proposal_id, source_module, source_record_type, source_record_id, trigger_event,
		 proposed_action_type, action_parameters, explanation, evidence, confidence, risk_level,
		 priority, requires_approval, approval_policy, expected_impact, reversibility, missing_information,
		 data_freshness, correlation_id, approval_id, execution_id, status, schema_version, created_by,
		 expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`

	res, err := r.db.ExecContext(ctx, query,
		p.OrgID, p.ProposalID, p.SourceModule, p.SourceRecordType, p.SourceRecordID, p.TriggerEvent,
		p.ProposedActionType, string(p.ActionParameters), p.Explanation, string(p.Evidence), p.Confidence, p.RiskLevel,
		p.Priority, p.RequiresApproval, p.ApprovalPolicy, p.ExpectedImpact, p.Reversibility, p.MissingInformation,
		p.DataFreshness, p.CorrelationID, p.ApprovalID, p.ExecutionID, p.Status, p.SchemaVersion, p.CreatedBy,
		p.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to insert action proposal: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		p.ID = id
	}
	return nil
}

func (r *repository) GetProposalByID(ctx context.Context, orgID int64, id int64) (*ActionProposal, error) {
	var p ActionProposal
	query := `SELECT id, org_id, proposal_id, source_module, source_record_type, source_record_id, trigger_event,
		proposed_action_type, action_parameters, explanation, evidence, confidence, risk_level,
		priority, requires_approval, approval_policy, expected_impact, reversibility, missing_information,
		data_freshness, correlation_id, approval_id, execution_id, status, schema_version, created_by,
		expires_at, created_at, updated_at
		FROM ai_action_proposals WHERE id = ? AND org_id = ?`
	err := r.db.GetContext(ctx, &p, query, id, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("action proposal #%d not found", id)
		}
		return nil, err
	}
	_ = json.Unmarshal(p.Evidence, &p.ParsedEvidence)
	return &p, nil
}

func (r *repository) GetProposalByProposalID(ctx context.Context, orgID int64, proposalID string) (*ActionProposal, error) {
	var p ActionProposal
	query := `SELECT id, org_id, proposal_id, source_module, source_record_type, source_record_id, trigger_event,
		proposed_action_type, action_parameters, explanation, evidence, confidence, risk_level,
		priority, requires_approval, approval_policy, expected_impact, reversibility, missing_information,
		data_freshness, correlation_id, approval_id, execution_id, status, schema_version, created_by,
		expires_at, created_at, updated_at
		FROM ai_action_proposals WHERE proposal_id = ? AND org_id = ?`
	err := r.db.GetContext(ctx, &p, query, proposalID, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("action proposal '%s' not found", proposalID)
		}
		return nil, err
	}
	_ = json.Unmarshal(p.Evidence, &p.ParsedEvidence)
	return &p, nil
}

func (r *repository) UpdateProposal(ctx context.Context, p *ActionProposal) error {
	query := `UPDATE ai_action_proposals SET
		approval_id = ?, execution_id = ?, status = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?`
	_, err := r.db.ExecContext(ctx, query, p.ApprovalID, p.ExecutionID, p.Status, p.ID, p.OrgID)
	return err
}

func (r *repository) ListProposals(ctx context.Context, orgID int64, filter ProposalFilter) ([]*ActionProposal, int64, error) {
	where := []string{"org_id = ?"}
	args := []interface{}{orgID}

	if filter.Status != nil && *filter.Status != "" {
		where = append(where, "status = ?")
		args = append(args, *filter.Status)
	}
	if filter.SourceModule != nil && *filter.SourceModule != "" {
		where = append(where, "source_module = ?")
		args = append(args, *filter.SourceModule)
	}
	if filter.RiskLevel != nil && *filter.RiskLevel != "" {
		where = append(where, "risk_level = ?")
		args = append(args, *filter.RiskLevel)
	}
	if filter.Search != nil && *filter.Search != "" {
		where = append(where, "(explanation LIKE ? OR proposal_id LIKE ? OR source_record_id LIKE ?)")
		pattern := "%" + *filter.Search + "%"
		args = append(args, pattern, pattern, pattern)
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ai_action_proposals WHERE %s", whereClause)
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`SELECT id, org_id, proposal_id, source_module, source_record_type, source_record_id, trigger_event,
		proposed_action_type, action_parameters, explanation, evidence, confidence, risk_level,
		priority, requires_approval, approval_policy, expected_impact, reversibility, missing_information,
		data_freshness, correlation_id, approval_id, execution_id, status, schema_version, created_by,
		expires_at, created_at, updated_at
		FROM ai_action_proposals WHERE %s ORDER BY created_at DESC LIMIT %d OFFSET %d`, whereClause, limit, offset)

	var list []*ActionProposal
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, 0, err
	}

	for _, p := range list {
		_ = json.Unmarshal(p.Evidence, &p.ParsedEvidence)
	}
	return list, total, nil
}

func (r *repository) CreateExecution(ctx context.Context, exec *ActionExecution) error {
	query := `INSERT INTO ai_action_executions
		(org_id, proposal_id, action_name, idempotency_key, status, current_step, step_results,
		 input_payload, output_payload, verification_status, verification_details, retry_count,
		 max_retries, error_message, executed_by, correlation_id, started_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), NOW())`

	res, err := r.db.ExecContext(ctx, query,
		exec.OrgID, exec.ProposalID, exec.ActionName, exec.IdempotencyKey, exec.Status, exec.CurrentStep,
		exec.StepResults, string(exec.InputPayload), exec.OutputPayload, exec.VerificationStatus,
		exec.VerificationDetails, exec.RetryCount, exec.MaxRetries, exec.ErrorMessage, exec.ExecutedBy,
		exec.CorrelationID)
	if err != nil {
		return fmt.Errorf("failed to insert action execution: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		exec.ID = id
	}
	return nil
}

func (r *repository) GetExecutionByID(ctx context.Context, orgID int64, id int64) (*ActionExecution, error) {
	var exec ActionExecution
	query := `SELECT id, org_id, proposal_id, action_name, idempotency_key, status, current_step,
		step_results, input_payload, output_payload, verification_status, verification_details,
		retry_count, max_retries, error_message, executed_by, correlation_id, started_at, completed_at,
		created_at, updated_at
		FROM ai_action_executions WHERE id = ? AND org_id = ?`
	err := r.db.GetContext(ctx, &exec, query, id, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("action execution #%d not found", id)
		}
		return nil, err
	}
	return &exec, nil
}

func (r *repository) GetExecutionByIdempotencyKey(ctx context.Context, orgID int64, idempotencyKey string) (*ActionExecution, error) {
	var exec ActionExecution
	query := `SELECT id, org_id, proposal_id, action_name, idempotency_key, status, current_step,
		step_results, input_payload, output_payload, verification_status, verification_details,
		retry_count, max_retries, error_message, executed_by, correlation_id, started_at, completed_at,
		created_at, updated_at
		FROM ai_action_executions WHERE idempotency_key = ? AND org_id = ?`
	err := r.db.GetContext(ctx, &exec, query, idempotencyKey, orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &exec, nil
}

func (r *repository) UpdateExecution(ctx context.Context, exec *ActionExecution) error {
	query := `UPDATE ai_action_executions SET
		status = ?, current_step = ?, step_results = ?, output_payload = ?,
		verification_status = ?, verification_details = ?, retry_count = ?,
		error_message = ?, completed_at = ?, updated_at = NOW()
		WHERE id = ? AND org_id = ?`

	var compTime *time.Time
	if exec.Status == StatusCompleted || exec.Status == StatusFailed || exec.Status == StatusCancelled {
		now := time.Now()
		compTime = &now
		exec.CompletedAt = compTime
	}

	_, err := r.db.ExecContext(ctx, query,
		exec.Status, exec.CurrentStep, exec.StepResults, exec.OutputPayload,
		exec.VerificationStatus, exec.VerificationDetails, exec.RetryCount,
		exec.ErrorMessage, compTime, exec.ID, exec.OrgID)
	return err
}

func (r *repository) ListExecutions(ctx context.Context, orgID int64, filter ExecutionFilter) ([]*ActionExecution, int64, error) {
	where := []string{"org_id = ?"}
	args := []interface{}{orgID}

	if filter.Status != nil && *filter.Status != "" {
		where = append(where, "status = ?")
		args = append(args, *filter.Status)
	}
	if filter.ActionName != nil && *filter.ActionName != "" {
		where = append(where, "action_name = ?")
		args = append(args, *filter.ActionName)
	}
	if filter.ProposalID != nil && *filter.ProposalID != "" {
		where = append(where, "proposal_id = ?")
		args = append(args, *filter.ProposalID)
	}

	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ai_action_executions WHERE %s", whereClause)
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`SELECT id, org_id, proposal_id, action_name, idempotency_key, status, current_step,
		step_results, input_payload, output_payload, verification_status, verification_details,
		retry_count, max_retries, error_message, executed_by, correlation_id, started_at, completed_at,
		created_at, updated_at
		FROM ai_action_executions WHERE %s ORDER BY created_at DESC LIMIT %d OFFSET %d`, whereClause, limit, offset)

	var list []*ActionExecution
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
