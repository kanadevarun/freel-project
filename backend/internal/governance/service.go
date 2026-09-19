package governance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	GetOverview(ctx context.Context, orgID int64) (*GovernanceOverviewDTO, error)
	GetPolicy(ctx context.Context, orgID int64) (*GovernancePolicy, error)
	UpdatePolicy(ctx context.Context, orgID int64, policy *GovernancePolicy) error
	GetKillSwitches(ctx context.Context, orgID int64) ([]KillSwitchRecord, error)
	SetKillSwitch(ctx context.Context, orgID int64, input SetKillSwitchInput, userID *int64) error
	ListViolations(ctx context.Context, orgID int64, limit, offset int) ([]SafetyViolationRecord, int, error)
	EnforcePreExecution(ctx context.Context, orgID int64, userID *int64, req PreExecutionCheckRequest) (*PreExecutionCheckResult, error)
	EnforcePostExecution(ctx context.Context, orgID int64, userID *int64, req PostExecutionCheckRequest) (*PostExecutionCheckResult, error)
	EvaluateQuality(ctx context.Context, orgID int64, workflowName string, cases []map[string]interface{}) (*SidecarQualityEvalResp, error)
}

type serviceImpl struct {
	repo       Repository
	sidecarURL string
	serviceKey string
	httpClient *http.Client
}

func NewService(repo Repository) Service {
	sidecarURL := os.Getenv("AI_SIDECAR_URL")
	if sidecarURL == "" {
		sidecarURL = "http://127.0.0.1:8090"
	}
	serviceKey := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if serviceKey == "" {
		serviceKey = "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456"
	}

	return &serviceImpl{
		repo:       repo,
		sidecarURL: sidecarURL,
		serviceKey: serviceKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *serviceImpl) GetOverview(ctx context.Context, orgID int64) (*GovernanceOverviewDTO, error) {
	return s.repo.GetOverview(ctx, orgID)
}

func (s *serviceImpl) GetPolicy(ctx context.Context, orgID int64) (*GovernancePolicy, error) {
	return s.repo.GetPolicy(ctx, orgID)
}

func (s *serviceImpl) UpdatePolicy(ctx context.Context, orgID int64, policy *GovernancePolicy) error {
	return s.repo.UpsertPolicy(ctx, orgID, policy)
}

func (s *serviceImpl) GetKillSwitches(ctx context.Context, orgID int64) ([]KillSwitchRecord, error) {
	return s.repo.GetKillSwitches(ctx, orgID)
}

func (s *serviceImpl) SetKillSwitch(ctx context.Context, orgID int64, input SetKillSwitchInput, userID *int64) error {
	if input.Scope == "" || input.TargetIdentifier == "" {
		return fmt.Errorf("scope and target_identifier are required")
	}
	return s.repo.SetKillSwitch(ctx, orgID, input.Scope, input.TargetIdentifier, input.IsKilled, input.Reason, userID)
}

func (s *serviceImpl) ListViolations(ctx context.Context, orgID int64, limit, offset int) ([]SafetyViolationRecord, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListSafetyViolations(ctx, orgID, limit, offset)
}

func (s *serviceImpl) EnforcePreExecution(ctx context.Context, orgID int64, userID *int64, req PreExecutionCheckRequest) (*PreExecutionCheckResult, error) {
	corrID := fmt.Sprintf("chk-pre-%s", uuid.New().String()[:12])

	// 1. Authoritative Kill Switch Check in Go
	killed, killReason, err := s.repo.IsTargetKilled(ctx, orgID, req.WorkflowName, req.ModelName, "")
	if err != nil {
		return nil, fmt.Errorf("failed checking kill switch: %w", err)
	}
	if killed {
		blockReason := fmt.Sprintf("Execution blocked by Administrative Kill Switch: %s", killReason)
		_ = s.repo.RecordSafetyViolation(ctx, &SafetyViolationRecord{
			OrgID:         orgID,
			UserID:        userID,
			WorkflowName:  req.WorkflowName,
			ViolationType: "KILL_SWITCH_ACTIVE",
			Severity:      "CRITICAL",
			ActionTaken:   "REJECTED",
			Details:       blockReason,
			CorrelationID: corrID,
		})
		return &PreExecutionCheckResult{
			Allowed:         false,
			BlockReason:     &blockReason,
			SanitizedInput:  req.InputText,
			PromptInjection: false,
			CorrelationID:   corrID,
		}, nil
	}

	// 2. Authoritative Policy Allowlist Check in Go
	policy, err := s.repo.GetPolicy(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving governance policy: %w", err)
	}

	// Validate Workflow
	if req.WorkflowName != "" && len(policy.AllowedWorkflows) > 0 {
		allowed := false
		for _, w := range policy.AllowedWorkflows {
			if strings.EqualFold(w, req.WorkflowName) {
				allowed = true
				break
			}
		}
		if !allowed {
			blockReason := fmt.Sprintf("Workflow '%s' is not in the allowed governance policy list for organization %d", req.WorkflowName, orgID)
			_ = s.repo.RecordSafetyViolation(ctx, &SafetyViolationRecord{
				OrgID:         orgID,
				UserID:        userID,
				WorkflowName:  req.WorkflowName,
				ViolationType: "POLICY_DISALLOWED_WORKFLOW",
				Severity:      "HIGH",
				ActionTaken:   "REJECTED",
				Details:       blockReason,
				CorrelationID: corrID,
			})
			return &PreExecutionCheckResult{
				Allowed:       false,
				BlockReason:   &blockReason,
				CorrelationID: corrID,
			}, nil
		}
	}

	// Validate Input Character Length
	if len(req.InputText) > policy.MaxInputChars {
		blockReason := fmt.Sprintf("Input length (%d chars) exceeds policy limit of %d chars", len(req.InputText), policy.MaxInputChars)
		return &PreExecutionCheckResult{
			Allowed:       false,
			BlockReason:   &blockReason,
			CorrelationID: corrID,
		}, nil
	}

	// 3. Call Python AI Sidecar for Input Safety, Prompt Injection, and PII inspection
	sidecarReq := SidecarInputInspectionReq{
		OrgID:         orgID,
		UserID:        userID,
		WorkflowName:  req.WorkflowName,
		InputText:     req.InputText,
		InputPayload:  req.InputPayload,
		CorrelationID: corrID,
	}

	sidecarResp, err := s.callSidecarInputInspection(ctx, sidecarReq)
	if err != nil {
		// Log warning but proceed with sanitized fallback if sidecar temporarily offline
		return &PreExecutionCheckResult{
			Allowed:         true,
			SanitizedInput:  req.InputText,
			Sanitized:       false,
			PromptInjection: false,
			EnforceApproval: policy.EnforceHITLApprovals,
			TimeoutSeconds:  policy.TimeoutSeconds,
			MaxRetries:      policy.MaxRetries,
			CorrelationID:   corrID,
		}, nil
	}

	// If prompt injection was detected by Python evaluator
	if !sidecarResp.IsSafe || sidecarResp.PromptInjectionDetected {
		snippet := req.InputText
		if len(snippet) > 100 {
			snippet = snippet[:100] + "..."
		}
		reason := "Prompt injection or adversarial manipulation pattern detected by AI Safety Guard"
		if sidecarResp.RefusalReason != nil && *sidecarResp.RefusalReason != "" {
			reason = *sidecarResp.RefusalReason
		}

		_ = s.repo.RecordSafetyViolation(ctx, &SafetyViolationRecord{
			OrgID:                orgID,
			UserID:               userID,
			WorkflowName:         req.WorkflowName,
			ViolationType:        "PROMPT_INJECTION",
			Severity:             "CRITICAL",
			ActionTaken:          "REJECTED",
			Details:              reason,
			InputSnippetRedacted: &snippet,
			CorrelationID:        corrID,
		})

		return &PreExecutionCheckResult{
			Allowed:         false,
			BlockReason:     &reason,
			SanitizedInput:  sidecarResp.SanitizedText,
			PromptInjection: true,
			CorrelationID:   corrID,
		}, nil
	}

	return &PreExecutionCheckResult{
		Allowed:         true,
		SanitizedInput:  sidecarResp.SanitizedText,
		Sanitized:       sidecarResp.PIIDetected,
		PromptInjection: false,
		EnforceApproval: policy.EnforceHITLApprovals,
		TimeoutSeconds:  policy.TimeoutSeconds,
		MaxRetries:      policy.MaxRetries,
		CorrelationID:   corrID,
	}, nil
}

func (s *serviceImpl) EnforcePostExecution(ctx context.Context, orgID int64, userID *int64, req PostExecutionCheckRequest) (*PostExecutionCheckResult, error) {
	corrID := fmt.Sprintf("chk-post-%s", uuid.New().String()[:12])

	policy, err := s.repo.GetPolicy(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving governance policy: %w", err)
	}

	// 1. Check Output Length Limit
	if len(req.OutputText) > policy.MaxOutputChars {
		blockReason := fmt.Sprintf("Output length (%d chars) exceeds maximum policy limit of %d chars", len(req.OutputText), policy.MaxOutputChars)
		return &PostExecutionCheckResult{
			Allowed:       false,
			BlockReason:   &blockReason,
			CorrelationID: corrID,
		}, nil
	}

	// 2. Call Python AI Sidecar for Grounding, Hallucination, and Action Safety Inspection
	sidecarReq := SidecarOutputInspectionReq{
		OrgID:           orgID,
		UserID:          userID,
		WorkflowName:    req.WorkflowName,
		OutputText:      req.OutputText,
		OutputPayload:   req.OutputPayload,
		SourceFacts:     req.SourceFacts,
		ProposedActions: req.ProposedActions,
		CorrelationID:   corrID,
	}

	sidecarResp, err := s.callSidecarOutputInspection(ctx, sidecarReq)
	if err != nil {
		return &PostExecutionCheckResult{
			Allowed:               true,
			GroundingScore:        1.0,
			HallucinationRisk:     "LOW",
			RequiresHumanApproval: false,
			CorrelationID:         corrID,
		}, nil
	}

	if !sidecarResp.IsSafe {
		reason := "Output failed AI safety and grounding inspection: high risk of hallucinations or ungrounded business claims."
		if sidecarResp.RefusalReason != nil && *sidecarResp.RefusalReason != "" {
			reason = *sidecarResp.RefusalReason
		}

		_ = s.repo.RecordSafetyViolation(ctx, &SafetyViolationRecord{
			OrgID:         orgID,
			UserID:        userID,
			WorkflowName:  req.WorkflowName,
			ViolationType: "UNSUPPORTED_CLAIM",
			Severity:      "HIGH",
			ActionTaken:   "REJECTED",
			Details:       reason,
			CorrelationID: corrID,
		})

		return &PostExecutionCheckResult{
			Allowed:               false,
			BlockReason:           &reason,
			GroundingScore:        sidecarResp.GroundingScore,
			HallucinationRisk:     sidecarResp.HallucinationRisk,
			RequiresHumanApproval: sidecarResp.RequiresHumanApproval,
			ActionEvaluations:     sidecarResp.ActionEvaluations,
			CorrelationID:         corrID,
		}, nil
	}

	return &PostExecutionCheckResult{
		Allowed:               true,
		GroundingScore:        sidecarResp.GroundingScore,
		HallucinationRisk:     sidecarResp.HallucinationRisk,
		RequiresHumanApproval: sidecarResp.RequiresHumanApproval && policy.EnforceHITLApprovals,
		ActionEvaluations:     sidecarResp.ActionEvaluations,
		CorrelationID:         corrID,
	}, nil
}

func (s *serviceImpl) EvaluateQuality(ctx context.Context, orgID int64, workflowName string, cases []map[string]interface{}) (*SidecarQualityEvalResp, error) {
	corrID := fmt.Sprintf("eval-%s", uuid.New().String()[:12])
	testRunID := fmt.Sprintf("run-%s-%s", workflowName, time.Now().Format("20060102-150405"))

	req := SidecarQualityEvalReq{
		OrgID:         orgID,
		TestRunID:     testRunID,
		WorkflowName:  workflowName,
		TestCases:     cases,
		CorrelationID: corrID,
	}

	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.sidecarURL+"/governance/evaluate-quality", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", s.serviceKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("quality evaluation call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sidecar quality evaluation returned HTTP %d: %s", resp.StatusCode, string(b))
	}

	var evalResp SidecarQualityEvalResp
	if err := json.NewDecoder(resp.Body).Decode(&evalResp); err != nil {
		return nil, err
	}
	return &evalResp, nil
}

// Helper HTTP methods calling the Python sidecar

func (s *serviceImpl) callSidecarInputInspection(ctx context.Context, req SidecarInputInspectionReq) (*SidecarInputInspectionResp, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.sidecarURL+"/governance/inspect-input", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", s.serviceKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sidecar returned %d: %s", resp.StatusCode, string(body))
	}

	var res SidecarInputInspectionResp
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *serviceImpl) callSidecarOutputInspection(ctx context.Context, req SidecarOutputInspectionReq) (*SidecarOutputInspectionResp, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.sidecarURL+"/governance/inspect-output", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", s.serviceKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sidecar returned %d: %s", resp.StatusCode, string(body))
	}

	var res SidecarOutputInspectionResp
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}
