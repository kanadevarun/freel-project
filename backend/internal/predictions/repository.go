package predictions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type Repository interface {
	Create(ctx context.Context, p *Prediction) error
	GetByID(ctx context.Context, orgID int64, predictionID string) (*Prediction, error)
	GetByIdempotencyKey(ctx context.Context, orgID int64, key string) (*Prediction, error)
	List(ctx context.Context, orgID int64, params FilterParams) ([]*Prediction, int, error)
	UpdateStatus(ctx context.Context, orgID int64, predictionID string, newStatus PredictionStatus, reviewStatus string, userID *int64, notes *string) error
	RecordAudit(ctx context.Context, a *PredictionAuditHistory) error
	GetAuditHistory(ctx context.Context, orgID int64, predictionID string) ([]*PredictionAuditHistory, error)
	RecordOutcome(ctx context.Context, orgID int64, predictionID string, outcomeStatus OutcomeStatus, outcomeValue *string, feedbackNotes *string) error
	GetActivePrediction(ctx context.Context, orgID int64, module string, relatedRecordType string, relatedRecordID string, predictionType string) (*Prediction, error)
	SupersedeExisting(ctx context.Context, orgID int64, module string, relatedRecordType string, relatedRecordID string, predictionType string, exceptPredictionID string) error
}

type mysqlRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) Repository {
	return &mysqlRepository{db: db}
}

func (r *mysqlRepository) Create(ctx context.Context, p *Prediction) error {
	signalsJSON, _ := json.Marshal(p.SupportingSignals)
	sourcesJSON, _ := json.Marshal(p.SourceReferences)

	query := `
		INSERT INTO predictions (
			org_id, user_id, prediction_id, idempotency_key, module, prediction_type,
			status, severity, confidence_score, confidence_band, related_record_type, related_record_id,
			prediction_statement, predicted_value, time_horizon, target_date, explanation,
			supporting_signals, source_references, source_timestamp, recommended_action, action_type,
			is_action_required, requires_approval, review_status, actual_outcome_status, model_version,
			expires_at, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, NOW(), NOW()
		)
	`
	res, err := r.db.ExecContext(
		ctx, query,
		p.OrgID, p.UserID, p.PredictionID, p.IdempotencyKey, p.Module, p.PredictionType,
		string(p.Status), string(p.Severity), p.ConfidenceScore, string(p.ConfidenceBand), p.RelatedRecordType, p.RelatedRecordID,
		p.PredictionStatement, p.PredictedValue, p.TimeHorizon, p.TargetDate, p.Explanation,
		string(signalsJSON), string(sourcesJSON), p.SourceTimestamp, p.RecommendedAction, p.ActionType,
		p.IsActionRequired, p.RequiresApproval, p.ReviewStatus, string(p.ActualOutcomeStatus), p.ModelVersion,
		p.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert prediction: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		p.ID = id
	}
	return nil
}

func (r *mysqlRepository) GetByID(ctx context.Context, orgID int64, predictionID string) (*Prediction, error) {
	query := `
		SELECT 
			id, org_id, user_id, prediction_id, idempotency_key, module, prediction_type,
			status, severity, confidence_score, confidence_band, related_record_type, related_record_id,
			prediction_statement, predicted_value, time_horizon, target_date, explanation,
			COALESCE(supporting_signals, '[]'), COALESCE(source_references, '[]'), source_timestamp,
			recommended_action, action_type, is_action_required, requires_approval, action_proposal_id,
			review_status, reviewed_by, reviewed_at, review_notes, actual_outcome_status,
			actual_outcome_value, feedback_notes, model_version, expires_at, created_at, updated_at
		FROM predictions
		WHERE org_id = ? AND prediction_id = ?
	`
	row := r.db.QueryRowContext(ctx, query, orgID, predictionID)
	p := &Prediction{}
	var statusStr, severityStr, bandStr, outcomeStr string
	err := row.Scan(
		&p.ID, &p.OrgID, &p.UserID, &p.PredictionID, &p.IdempotencyKey, &p.Module, &p.PredictionType,
		&statusStr, &severityStr, &p.ConfidenceScore, &bandStr, &p.RelatedRecordType, &p.RelatedRecordID,
		&p.PredictionStatement, &p.PredictedValue, &p.TimeHorizon, &p.TargetDate, &p.Explanation,
		&p.SupportingSignalsRaw, &p.SourceReferencesRaw, &p.SourceTimestamp,
		&p.RecommendedAction, &p.ActionType, &p.IsActionRequired, &p.RequiresApproval, &p.ActionProposalID,
		&p.ReviewStatus, &p.ReviewedBy, &p.ReviewedAt, &p.ReviewNotes, &outcomeStr,
		&p.ActualOutcomeValue, &p.FeedbackNotes, &p.ModelVersion, &p.ExpiresAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query prediction: %w", err)
	}

	p.Status = PredictionStatus(statusStr)
	p.Severity = Severity(severityStr)
	p.ConfidenceBand = ConfidenceBand(bandStr)
	p.ActualOutcomeStatus = OutcomeStatus(outcomeStr)
	p.UnpackJSON()
	return p, nil
}

func (r *mysqlRepository) GetByIdempotencyKey(ctx context.Context, orgID int64, key string) (*Prediction, error) {
	query := `
		SELECT 
			id, org_id, user_id, prediction_id, idempotency_key, module, prediction_type,
			status, severity, confidence_score, confidence_band, related_record_type, related_record_id,
			prediction_statement, predicted_value, time_horizon, target_date, explanation,
			COALESCE(supporting_signals, '[]'), COALESCE(source_references, '[]'), source_timestamp,
			recommended_action, action_type, is_action_required, requires_approval, action_proposal_id,
			review_status, reviewed_by, reviewed_at, review_notes, actual_outcome_status,
			actual_outcome_value, feedback_notes, model_version, expires_at, created_at, updated_at
		FROM predictions
		WHERE org_id = ? AND idempotency_key = ?
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, orgID, key)
	p := &Prediction{}
	var statusStr, severityStr, bandStr, outcomeStr string
	err := row.Scan(
		&p.ID, &p.OrgID, &p.UserID, &p.PredictionID, &p.IdempotencyKey, &p.Module, &p.PredictionType,
		&statusStr, &severityStr, &p.ConfidenceScore, &bandStr, &p.RelatedRecordType, &p.RelatedRecordID,
		&p.PredictionStatement, &p.PredictedValue, &p.TimeHorizon, &p.TargetDate, &p.Explanation,
		&p.SupportingSignalsRaw, &p.SourceReferencesRaw, &p.SourceTimestamp,
		&p.RecommendedAction, &p.ActionType, &p.IsActionRequired, &p.RequiresApproval, &p.ActionProposalID,
		&p.ReviewStatus, &p.ReviewedBy, &p.ReviewedAt, &p.ReviewNotes, &outcomeStr,
		&p.ActualOutcomeValue, &p.FeedbackNotes, &p.ModelVersion, &p.ExpiresAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query prediction by idempotency: %w", err)
	}

	p.Status = PredictionStatus(statusStr)
	p.Severity = Severity(severityStr)
	p.ConfidenceBand = ConfidenceBand(bandStr)
	p.ActualOutcomeStatus = OutcomeStatus(outcomeStr)
	p.UnpackJSON()
	return p, nil
}

func (r *mysqlRepository) List(ctx context.Context, orgID int64, params FilterParams) ([]*Prediction, int, error) {
	where := "WHERE org_id = ?"
	args := []interface{}{orgID}

	if params.Module != "" {
		where += " AND module = ?"
		args = append(args, params.Module)
	}
	if params.Severity != "" {
		where += " AND severity = ?"
		args = append(args, params.Severity)
	}
	if params.ConfidenceBand != "" {
		where += " AND confidence_band = ?"
		args = append(args, params.ConfidenceBand)
	}
	if params.Status != "" {
		where += " AND status = ?"
		args = append(args, params.Status)
	}
	if params.RelatedRecordType != "" {
		where += " AND related_record_type = ?"
		args = append(args, params.RelatedRecordType)
	}
	if params.RelatedRecordID != "" {
		where += " AND related_record_id = ?"
		args = append(args, params.RelatedRecordID)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM predictions %s", where)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count predictions: %w", err)
	}

	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT 
			id, org_id, user_id, prediction_id, idempotency_key, module, prediction_type,
			status, severity, confidence_score, confidence_band, related_record_type, related_record_id,
			prediction_statement, predicted_value, time_horizon, target_date, explanation,
			COALESCE(supporting_signals, '[]'), COALESCE(source_references, '[]'), source_timestamp,
			recommended_action, action_type, is_action_required, requires_approval, action_proposal_id,
			review_status, reviewed_by, reviewed_at, review_notes, actual_outcome_status,
			actual_outcome_value, feedback_notes, model_version, expires_at, created_at, updated_at
		FROM predictions
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, where)
	queryArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list predictions: %w", err)
	}
	defer rows.Close()

	var result []*Prediction
	for rows.Next() {
		p := &Prediction{}
		var statusStr, severityStr, bandStr, outcomeStr string
		if err := rows.Scan(
			&p.ID, &p.OrgID, &p.UserID, &p.PredictionID, &p.IdempotencyKey, &p.Module, &p.PredictionType,
			&statusStr, &severityStr, &p.ConfidenceScore, &bandStr, &p.RelatedRecordType, &p.RelatedRecordID,
			&p.PredictionStatement, &p.PredictedValue, &p.TimeHorizon, &p.TargetDate, &p.Explanation,
			&p.SupportingSignalsRaw, &p.SourceReferencesRaw, &p.SourceTimestamp,
			&p.RecommendedAction, &p.ActionType, &p.IsActionRequired, &p.RequiresApproval, &p.ActionProposalID,
			&p.ReviewStatus, &p.ReviewedBy, &p.ReviewedAt, &p.ReviewNotes, &outcomeStr,
			&p.ActualOutcomeValue, &p.FeedbackNotes, &p.ModelVersion, &p.ExpiresAt, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan prediction: %w", err)
		}
		p.Status = PredictionStatus(statusStr)
		p.Severity = Severity(severityStr)
		p.ConfidenceBand = ConfidenceBand(bandStr)
		p.ActualOutcomeStatus = OutcomeStatus(outcomeStr)
		p.UnpackJSON()
		result = append(result, p)
	}

	return result, total, nil
}

func (r *mysqlRepository) UpdateStatus(ctx context.Context, orgID int64, predictionID string, newStatus PredictionStatus, reviewStatus string, userID *int64, notes *string) error {
	query := `
		UPDATE predictions
		SET status = ?, review_status = ?, reviewed_by = ?, reviewed_at = NOW(), review_notes = ?, updated_at = NOW()
		WHERE org_id = ? AND prediction_id = ?
	`
	res, err := r.db.ExecContext(ctx, query, string(newStatus), reviewStatus, userID, notes, orgID, predictionID)
	if err != nil {
		return fmt.Errorf("failed to update prediction status: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("prediction %s not found for organization %d", predictionID, orgID)
	}
	return nil
}

func (r *mysqlRepository) RecordAudit(ctx context.Context, a *PredictionAuditHistory) error {
	query := `
		INSERT INTO prediction_audit_history (
			prediction_id, org_id, user_id, previous_status, new_status, action, notes, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, NOW()
		)
	`
	_, err := r.db.ExecContext(ctx, query, a.PredictionID, a.OrgID, a.UserID, a.PreviousStatus, a.NewStatus, a.Action, a.Notes)
	return err
}

func (r *mysqlRepository) GetAuditHistory(ctx context.Context, orgID int64, predictionID string) ([]*PredictionAuditHistory, error) {
	query := `
		SELECT id, prediction_id, org_id, user_id, previous_status, new_status, action, notes, created_at
		FROM prediction_audit_history
		WHERE org_id = ? AND prediction_id = ?
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, orgID, predictionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*PredictionAuditHistory
	for rows.Next() {
		a := &PredictionAuditHistory{}
		if err := rows.Scan(&a.ID, &a.PredictionID, &a.OrgID, &a.UserID, &a.PreviousStatus, &a.NewStatus, &a.Action, &a.Notes, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (r *mysqlRepository) RecordOutcome(ctx context.Context, orgID int64, predictionID string, outcomeStatus OutcomeStatus, outcomeValue *string, feedbackNotes *string) error {
	query := `
		UPDATE predictions
		SET actual_outcome_status = ?, actual_outcome_value = ?, feedback_notes = ?, updated_at = NOW()
		WHERE org_id = ? AND prediction_id = ?
	`
	res, err := r.db.ExecContext(ctx, query, string(outcomeStatus), outcomeValue, feedbackNotes, orgID, predictionID)
	if err != nil {
		return fmt.Errorf("failed to record outcome: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("prediction %s not found for organization %d", predictionID, orgID)
	}
	return nil
}

func (r *mysqlRepository) GetActivePrediction(ctx context.Context, orgID int64, module string, relatedRecordType string, relatedRecordID string, predictionType string) (*Prediction, error) {
	query := `
		SELECT 
			id, org_id, user_id, prediction_id, idempotency_key, module, prediction_type,
			status, severity, confidence_score, confidence_band, related_record_type, related_record_id,
			prediction_statement, predicted_value, time_horizon, target_date, explanation,
			COALESCE(supporting_signals, '[]'), COALESCE(source_references, '[]'), source_timestamp,
			recommended_action, action_type, is_action_required, requires_approval, action_proposal_id,
			review_status, reviewed_by, reviewed_at, review_notes, actual_outcome_status,
			actual_outcome_value, feedback_notes, model_version, expires_at, created_at, updated_at
		FROM predictions
		WHERE org_id = ? AND module = ? AND related_record_type = ? AND related_record_id = ? AND prediction_type = ?
		  AND status IN ('PUBLISHED', 'ACKNOWLEDGED', 'IN_REVIEW', 'ACCEPTED', 'ACTION_REQUESTED')
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, orgID, module, relatedRecordType, relatedRecordID, predictionType)
	p := &Prediction{}
	var statusStr, severityStr, bandStr, outcomeStr string
	err := row.Scan(
		&p.ID, &p.OrgID, &p.UserID, &p.PredictionID, &p.IdempotencyKey, &p.Module, &p.PredictionType,
		&statusStr, &severityStr, &p.ConfidenceScore, &bandStr, &p.RelatedRecordType, &p.RelatedRecordID,
		&p.PredictionStatement, &p.PredictedValue, &p.TimeHorizon, &p.TargetDate, &p.Explanation,
		&p.SupportingSignalsRaw, &p.SourceReferencesRaw, &p.SourceTimestamp,
		&p.RecommendedAction, &p.ActionType, &p.IsActionRequired, &p.RequiresApproval, &p.ActionProposalID,
		&p.ReviewStatus, &p.ReviewedBy, &p.ReviewedAt, &p.ReviewNotes, &outcomeStr,
		&p.ActualOutcomeValue, &p.FeedbackNotes, &p.ModelVersion, &p.ExpiresAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query active prediction: %w", err)
	}
	p.Status = PredictionStatus(statusStr)
	p.Severity = Severity(severityStr)
	p.ConfidenceBand = ConfidenceBand(bandStr)
	p.ActualOutcomeStatus = OutcomeStatus(outcomeStr)
	p.UnpackJSON()
	return p, nil
}

func (r *mysqlRepository) SupersedeExisting(ctx context.Context, orgID int64, module string, relatedRecordType string, relatedRecordID string, predictionType string, exceptPredictionID string) error {
	// 1. Fetch IDs of predictions to supersede
	queryFind := `
		SELECT prediction_id, status FROM predictions
		WHERE org_id = ? AND module = ? AND related_record_type = ? AND related_record_id = ? AND prediction_type = ?
		  AND prediction_id != ?
		  AND status IN ('PUBLISHED', 'ACKNOWLEDGED', 'IN_REVIEW')
	`
	rows, err := r.db.QueryContext(ctx, queryFind, orgID, module, relatedRecordType, relatedRecordID, predictionType, exceptPredictionID)
	if err != nil {
		return fmt.Errorf("failed to find predictions to supersede: %w", err)
	}
	defer rows.Close()

	type toSupersede struct {
		id  string
		old string
	}
	var items []toSupersede
	for rows.Next() {
		var it toSupersede
		if err := rows.Scan(&it.id, &it.old); err == nil {
			items = append(items, it)
		}
	}

	if len(items) == 0 {
		return nil
	}

	// 2. Mark them SUPERSEDED
	queryUpdate := `
		UPDATE predictions
		SET status = 'SUPERSEDED', updated_at = NOW()
		WHERE org_id = ? AND prediction_id = ?
	`
	for _, it := range items {
		_, _ = r.db.ExecContext(ctx, queryUpdate, orgID, it.id)
		oldStatus := it.old
		notes := fmt.Sprintf("Superseded by newer prediction %s", exceptPredictionID)
		_ = r.RecordAudit(ctx, &PredictionAuditHistory{
			PredictionID:   it.id,
			OrgID:          orgID,
			PreviousStatus: &oldStatus,
			NewStatus:      string(StatusSuperseded),
			Action:         "SUPERSEDED_BY_NEW_PREDICTION",
			Notes:          &notes,
		})
	}
	return nil
}
