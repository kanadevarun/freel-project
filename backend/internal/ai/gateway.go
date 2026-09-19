package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Gateway routes AI requests to configured providers, managing failover, observability, and audit logging.
type Gateway interface {
	ExecutePrompt(ctx context.Context, prompt string) (string, error)
	ExecutePromptWithContext(ctx context.Context, execCtx *ExecutionContext, prompt string) (*ExecutionResult, error)
	GetConfig() *RuntimeConfig
	HasProvider(name string) bool
}

type gateway struct {
	providers map[string]Provider
	config    *RuntimeConfig
	db        *sqlx.DB
}

// NewGatewayWithConfig creates a new AI Gateway with authoritative runtime configuration and DB telemetry.
func NewGatewayWithConfig(providers map[string]Provider, cfg *RuntimeConfig, db *sqlx.DB) Gateway {
	if cfg == nil {
		var err error
		cfg, err = LoadRuntimeConfigFromEnv()
		if err != nil {
			log.Printf("[AI Gateway] Warning: Failed to load runtime config from env, using safe defaults: %v", err)
			cfg = &RuntimeConfig{
				Environment:            "development",
				PrimaryProvider:        "gemini",
				PrimaryModel:           "gemini-1.5-flash",
				FailoverProvider:       "openai",
				FailoverModel:          "gpt-4o-mini",
				AllowMockFallback:      true,
				MaxRetries:             2,
				ProviderTimeoutSeconds: 30,
			}
		}
	}
	return &gateway{
		providers: providers,
		config:    cfg,
		db:        db,
	}
}

// NewGateway creates a gateway with default configuration for backwards compatibility.
func NewGateway(providers map[string]Provider) Gateway {
	return NewGatewayWithConfig(providers, nil, nil)
}

func (g *gateway) GetConfig() *RuntimeConfig {
	return g.config
}

func (g *gateway) HasProvider(name string) bool {
	if g.providers == nil {
		return false
	}
	_, ok := g.providers[name]
	return ok
}

// ExecutePrompt runs a prompt with default development execution context.
func (g *gateway) ExecutePrompt(ctx context.Context, prompt string) (string, error) {
	execCtx := &ExecutionContext{
		OrgID:         1, // Default dev org
		ActorType:     "AI_AGENT",
		Source:        "GO_GATEWAY",
		RequestID:     uuid.New().String(),
		WorkflowName:  "generic",
		ExecutionMode: "COMPLETION",
		CreatedAt:     time.Now(),
	}
	res, err := g.ExecutePromptWithContext(ctx, execCtx, prompt)
	if err != nil {
		return "", err
	}
	return res.RawOutput, nil
}

// ExecutePromptWithContext executes the prompt through the configured failover pipeline and records telemetry.
func (g *gateway) ExecutePromptWithContext(ctx context.Context, execCtx *ExecutionContext, prompt string) (*ExecutionResult, error) {
	startTime := time.Now()
	if execCtx == nil {
		execCtx = &ExecutionContext{
			OrgID:         1,
			ActorType:     "AI_AGENT",
			Source:        "GO_GATEWAY",
			RequestID:     uuid.New().String(),
			WorkflowName:  "generic",
			CreatedAt:     startTime,
		}
	}
	if execCtx.RequestID == "" {
		execCtx.RequestID = uuid.New().String()
	}

	result := &ExecutionResult{
		PrimaryProvider: g.config.PrimaryProvider,
		PrimaryModel:    g.config.PrimaryModel,
		CreatedAt:       startTime,
	}

	var primaryErr error
	var failoverErr error

	// ── STEP 1: PRIMARY PROVIDER (Default: Gemini) ───────────────────────────
	if primaryProv, exists := g.providers[g.config.PrimaryProvider]; exists && primaryProv != nil {
		log.Printf("[AI Gateway] [%s] Attempting primary provider '%s' (model: %s)...",
			execCtx.RequestID[:8], g.config.PrimaryProvider, g.config.PrimaryModel)

		res, err := primaryProv.GenerateCompletion(ctx, prompt)
		if err == nil {
			result.RawOutput = res
			result.FinalProvider = g.config.PrimaryProvider
			result.FinalModel = g.config.PrimaryModel
			result.CompletedAt = time.Now()
			result.DurationMs = result.CompletedAt.Sub(startTime).Milliseconds()
			g.recordTelemetry(ctx, execCtx, prompt, result, "COMPLETED", nil)
			return result, nil
		}

		primaryErr = err
		result.FailoverOccurred = true
		cat := ClassifyError(err)
		result.FailoverReason = fmt.Sprintf("Primary '%s' failed [%s]: %s", g.config.PrimaryProvider, cat, RedactSecrets(err.Error()))
		log.Printf("[AI Gateway] [%s] Primary failed: %s", execCtx.RequestID[:8], result.FailoverReason)
	}

	// ── STEP 2: FAILOVER PROVIDER (Default: OpenAI) ───────────────────────────
	if failoverProv, exists := g.providers[g.config.FailoverProvider]; exists && failoverProv != nil {
		log.Printf("[AI Gateway] [%s] Initiating failover to '%s' (model: %s)...",
			execCtx.RequestID[:8], g.config.FailoverProvider, g.config.FailoverModel)

		res, err := failoverProv.GenerateCompletion(ctx, prompt)
		if err == nil {
			result.RawOutput = res
			result.FinalProvider = g.config.FailoverProvider
			result.FinalModel = g.config.FailoverModel
			result.CompletedAt = time.Now()
			result.DurationMs = result.CompletedAt.Sub(startTime).Milliseconds()
			g.recordTelemetry(ctx, execCtx, prompt, result, "COMPLETED", nil)
			return result, nil
		}

		failoverErr = err
		log.Printf("[AI Gateway] [%s] Failover provider failed: %s", execCtx.RequestID[:8], RedactSecrets(err.Error()))
	}

	// ── STEP 3: MOCK FALLBACK (Development/Test Only) ─────────────────────────
	if g.config.AllowMockFallback && g.config.Environment != "production" {
		if mockProv, exists := g.providers["mock"]; exists && mockProv != nil {
			log.Printf("[AI Gateway] [%s] Both live providers failed. Using Local Mock Provider (dev mode allowed)...", execCtx.RequestID[:8])
			res, err := mockProv.GenerateCompletion(ctx, prompt)
			if err == nil {
				result.RawOutput = res
				result.FinalProvider = "mock"
				result.FinalModel = "mock-evaluator"
				result.IsMock = true
				result.CompletedAt = time.Now()
				result.DurationMs = result.CompletedAt.Sub(startTime).Milliseconds()
				g.recordTelemetry(ctx, execCtx, prompt, result, "COMPLETED", nil)
				return result, nil
			}
		}
	}

	// ── STEP 4: ALL PROVIDERS EXHAUSTED ──────────────────────────────────────
	result.FinalProvider = "none"
	result.FinalModel = "none"
	result.CompletedAt = time.Now()
	result.DurationMs = result.CompletedAt.Sub(startTime).Milliseconds()

	finalErrMsg := fmt.Sprintf("all configured providers failed. Primary: %v, Failover: %v",
		RedactSecrets(fmt.Sprintf("%v", primaryErr)), RedactSecrets(fmt.Sprintf("%v", failoverErr)))
	result.ErrorMessage = finalErrMsg
	result.ErrorCategory = ClassifyError(primaryErr)
	if result.ErrorCategory == ErrorCategoryUnknown && failoverErr != nil {
		result.ErrorCategory = ClassifyError(failoverErr)
	}

	g.recordTelemetry(ctx, execCtx, prompt, result, "FAILED", errors.New(finalErrMsg))
	return nil, fmt.Errorf("%w: %s", ErrProvidersUnavailable, finalErrMsg)
}

// recordTelemetry asynchronously logs execution trace to MariaDB ai_execution_traces table.
func (g *gateway) recordTelemetry(ctx context.Context, execCtx *ExecutionContext, prompt string, result *ExecutionResult, status string, execErr error) {
	if g.db == nil {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[AI Gateway Telemetry] Panic recovered: %v", r)
			}
		}()

		feature := execCtx.WorkflowName
		if feature == "" {
			feature = "generic"
		}
		assistant := "general_assistant"
		module := "SYSTEM"
		if execCtx.Metadata != nil {
			if a, ok := execCtx.Metadata["assistant"].(string); ok && a != "" {
				assistant = a
			}
			if m, ok := execCtx.Metadata["module"].(string); ok && m != "" {
				module = m
			}
		}
		reqType := execCtx.ExecutionMode
		if reqType == "" {
			reqType = "COMPLETION"
		}

		inTokens := len(prompt) / 4
		if inTokens == 0 && len(prompt) > 0 {
			inTokens = 1
		}
		outTokens := len(result.RawOutput) / 4
		if outTokens == 0 && len(result.RawOutput) > 0 {
			outTokens = 1
		}
		totTokens := inTokens + outTokens
		result.TotalTokens = totTokens

		var estimatedCost float64
		if !result.IsMock && result.FinalProvider != "none" {
			// Reference rate: $0.000150 / 1k in, $0.000600 / 1k out
			estimatedCost = (float64(inTokens)/1000.0)*0.000150 + (float64(outTokens)/1000.0)*0.000600
		}

		metaJSON, _ := json.Marshal(execCtx.Metadata)
		query := `
			INSERT INTO ai_execution_traces
			(org_id, user_id, task_id, thread_id, request_id, correlation_id,
			 workflow_name, feature, assistant, module, request_type, runtime_route,
			 prompt_key, prompt_version, primary_provider, primary_model,
			 final_provider, final_model, failover_occurred, failover_reason,
			 is_mock, status, duration_ms, input_tokens, output_tokens, total_tokens,
			 estimated_cost, cost_currency, retry_count, safety_status, grounding_status,
			 is_memory_assisted, error_category, error_message, execution_metadata,
			 created_at, completed_at)
			VALUES (?, ?, ?, ?, ?, ?,
			        ?, ?, ?, ?, ?, ?,
			        ?, ?, ?, ?,
			        ?, ?, ?, ?,
			        ?, ?, ?, ?, ?, ?,
			        ?, ?, ?, ?, ?,
			        ?, ?, ?, ?,
			        ?, ?)
		`
		var completedAt sql.NullTime
		if !result.CompletedAt.IsZero() {
			completedAt = sql.NullTime{Time: result.CompletedAt, Valid: true}
		}

		var taskID sql.NullInt64
		if execCtx.TaskID != nil {
			taskID = sql.NullInt64{Int64: *execCtx.TaskID, Valid: true}
		}

		var userID sql.NullInt64
		if execCtx.ActingUserID != nil {
			userID = sql.NullInt64{Int64: *execCtx.ActingUserID, Valid: true}
		}

		var corrID sql.NullString
		if execCtx.CorrelationID != "" {
			corrID = sql.NullString{String: execCtx.CorrelationID, Valid: true}
		}

		_, err := g.db.Exec(query,
			execCtx.OrgID,
			userID,
			taskID,
			execCtx.ThreadID,
			execCtx.RequestID,
			corrID,
			execCtx.WorkflowName,
			feature,
			assistant,
			module,
			reqType,
			"/api/v1/ai/completion",
			execCtx.PromptKey,
			execCtx.PromptVersion,
			result.PrimaryProvider,
			result.PrimaryModel,
			result.FinalProvider,
			result.FinalModel,
			result.FailoverOccurred,
			result.FailoverReason,
			result.IsMock,
			status,
			result.DurationMs,
			inTokens,
			outTokens,
			totTokens,
			estimatedCost,
			"USD",
			0, // retry_count
			"PASSED",
			"GROUNDED",
			false, // is_memory_assisted
			result.ErrorCategory,
			RedactSecrets(result.ErrorMessage),
			metaJSON,
			result.CreatedAt,
			completedAt,
		)
		if err != nil {
			log.Printf("[AI Gateway Telemetry] Failed to insert execution trace: %v", err)
		}
	}()
}
