package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// EvaluationResultRecord represents a single scenario execution outcome in MariaDB.
type EvaluationResultRecord struct {
	ID                  int64     `db:"id" json:"id"`
	OrgID               int64     `db:"org_id" json:"org_id"`
	TestRunID           string    `db:"test_run_id" json:"test_run_id"`
	ScenarioID          string    `db:"scenario_id" json:"scenario_id"`
	ScenarioName        string    `db:"scenario_name" json:"scenario_name"`
	ScenarioVersion     string    `db:"scenario_version" json:"scenario_version"`
	AgentKey            string    `db:"agent_key" json:"agent_key"`
	Category            string    `db:"category" json:"category"`
	Status              string    `db:"status" json:"status"`
	ExpectedResult      string    `db:"expected_result" json:"expected_result"`
	ActualResultSummary string    `db:"actual_result_summary" json:"actual_result_summary"`
	ErrorCategory       *string   `db:"error_category" json:"error_category"`
	ProviderMode        string    `db:"provider_mode" json:"provider_mode"`
	PromptVersion       *string   `db:"prompt_version" json:"prompt_version"`
	ActionNamesInvoked  *string   `db:"action_names_invoked" json:"action_names_invoked"`
	ApprovalState       *string   `db:"approval_state" json:"approval_state"`
	TaskState           *string   `db:"task_state" json:"task_state"`
	CheckpointState     *string   `db:"checkpoint_state" json:"checkpoint_state"`
	CorrelationID       *string   `db:"correlation_id" json:"correlation_id"`
	DurationMs          int64     `db:"duration_ms" json:"duration_ms"`
	StartedAt           time.Time `db:"started_at" json:"started_at"`
	CompletedAt         time.Time `db:"completed_at" json:"completed_at"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
}

// EvaluationStore manages persistence and querying of AI safety gate evaluation results.
type EvaluationStore interface {
	SaveResult(ctx context.Context, record *EvaluationResultRecord) (int64, error)
	ListResultsByRun(ctx context.Context, orgID int64, testRunID string) ([]*EvaluationResultRecord, error)
	GetRunSummary(ctx context.Context, orgID int64, testRunID string) (map[string]interface{}, error)
}

type evalStore struct {
	db *sqlx.DB
}

// NewEvaluationStore creates a new evaluation results persistence manager.
func NewEvaluationStore(db *sqlx.DB) EvaluationStore {
	return &evalStore{db: db}
}

func (s *evalStore) SaveResult(ctx context.Context, record *EvaluationResultRecord) (int64, error) {
	if s.db == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	query := `
		INSERT INTO ai_evaluation_results (
			org_id, test_run_id, scenario_id, scenario_name, scenario_version,
			agent_key, category, status, expected_result, actual_result_summary,
			error_category, provider_mode, prompt_version, action_names_invoked,
			approval_state, task_state, checkpoint_state, correlation_id,
			duration_ms, started_at, completed_at
		) VALUES (
			:org_id, :test_run_id, :scenario_id, :scenario_name, :scenario_version,
			:agent_key, :category, :status, :expected_result, :actual_result_summary,
			:error_category, :provider_mode, :prompt_version, :action_names_invoked,
			:approval_state, :task_state, :checkpoint_state, :correlation_id,
			:duration_ms, :started_at, :completed_at
		)
	`
	if record.StartedAt.IsZero() {
		record.StartedAt = time.Now()
	}
	if record.CompletedAt.IsZero() {
		record.CompletedAt = time.Now()
	}

	res, err := s.db.NamedExecContext(ctx, query, record)
	if err != nil {
		return 0, fmt.Errorf("failed to insert evaluation result: %w", err)
	}
	return res.LastInsertId()
}

func (s *evalStore) ListResultsByRun(ctx context.Context, orgID int64, testRunID string) ([]*EvaluationResultRecord, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `
		SELECT id, org_id, test_run_id, scenario_id, scenario_name, scenario_version,
		       agent_key, category, status, expected_result, actual_result_summary,
		       error_category, provider_mode, prompt_version, action_names_invoked,
		       approval_state, task_state, checkpoint_state, correlation_id,
		       duration_ms, started_at, completed_at, created_at
		FROM ai_evaluation_results
		WHERE org_id = ? AND test_run_id = ?
		ORDER BY id ASC
	`
	var records []*EvaluationResultRecord
	err := s.db.SelectContext(ctx, &records, query, orgID, testRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to query evaluation results: %w", err)
	}
	return records, nil
}

func (s *evalStore) GetRunSummary(ctx context.Context, orgID int64, testRunID string) (map[string]interface{}, error) {
	records, err := s.ListResultsByRun(ctx, orgID, testRunID)
	if err != nil {
		return nil, err
	}

	total := len(records)
	passed := 0
	failed := 0
	var totalDuration int64 = 0

	catStats := make(map[string]map[string]int)

	for _, r := range records {
		totalDuration += r.DurationMs
		if _, exists := catStats[r.Category]; !exists {
			catStats[r.Category] = map[string]int{"total": 0, "passed": 0, "failed": 0}
		}
		catStats[r.Category]["total"]++

		if r.Status == "PASS" {
			passed++
			catStats[r.Category]["passed"]++
		} else {
			failed++
			catStats[r.Category]["failed"]++
		}
	}

	summary := map[string]interface{}{
		"test_run_id":      testRunID,
		"org_id":           orgID,
		"total_scenarios":  total,
		"passed":           passed,
		"failed":           failed,
		"total_duration":   totalDuration,
		"is_release_ready": (failed == 0 && total > 0),
		"categories":       catStats,
	}
	return summary, nil
}
