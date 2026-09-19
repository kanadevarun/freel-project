package autonomy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GovernanceEngine defines the interface for centralized policy evaluation
type GovernanceEngine interface {
	Evaluate(ctx context.Context, req GovernanceEvaluationRequest) (*GovernanceEvaluationResponse, error)
	PreviewPlan(ctx context.Context, req PlanPreviewRequest) (*PlanPreviewResponse, error)
}

type governanceEngine struct {
	repo          Repository
	sidecarClient SidecarClient
}

// NewGovernanceEngine creates a new centralized Go governance engine
func NewGovernanceEngine(repo Repository, sidecarClient SidecarClient) GovernanceEngine {
	return &governanceEngine{
		repo:          repo,
		sidecarClient: sidecarClient,
	}
}

func (e *governanceEngine) Evaluate(ctx context.Context, req GovernanceEvaluationRequest) (*GovernanceEvaluationResponse, error) {
	corrID := req.CorrelationID
	if corrID == "" {
		corrID = fmt.Sprintf("corr-gov-%d", time.Now().UnixNano())
	}

	var reasons []string
	decision := GovernanceDecisionAllow
	var riskClass ActionRiskClass = RiskClassMedium
	var sufficiency DataSufficiencyState = SufficiencySufficient
	var confidence ConfidenceClass = ConfidenceClassHigh
	var reversibility ActionReversibility = ReversibilityReversible
	approvalRequired := false
	fourEyesEnforced := false
	policyVersion := 1

	// 1. Fetch Tenant Limits & Emergency Stop
	limits, err := e.repo.GetTenantLimits(ctx, req.OrgID)
	if err != nil || limits == nil {
		limits = &TenantGovernanceLimits{
			OrgID:                           req.OrgID,
			MaxTenantAutonomy:               3,
			KillSwitchActive:                false,
			MaxActionsPerHour:               100,
			MaxFinancialExposurePerWorkflow: 5000.0,
			MaxRetriesPerStep:               3,
			MaxReplansPerPlan:               5,
			EnforceFourEyes:                 true,
		}
	}

	// 1a. Check Kill Switch / Emergency Stop
	if limits.KillSwitchActive {
		decision = GovernanceDecisionBlock
		reasons = append(reasons, "Tenant kill switch / emergency stop is currently ACTIVE. Autonomous execution blocked.")
		resp := &GovernanceEvaluationResponse{
			Decision:          decision,
			EffectiveAutonomy: 0,
			RiskClass:         RiskClassCritical,
			DataSufficiency:   SufficiencySufficient,
			ConfidenceClass:   ConfidenceClassHigh,
			RequiresApproval:  true,
			FourEyesEnforced:  false,
			Reversibility:     ReversibilityIrreversible,
			Reasons:           reasons,
			PolicyVersion:     policyVersion,
			CorrelationID:     corrID,
		}
		e.recordEvaluation(ctx, req, resp)
		return resp, nil
	}

	// 2. Fetch Capability Feature Flag
	flagKey := deriveFlagKey(req.Module)
	flags, _ := e.repo.GetFeatureFlags(ctx, req.OrgID)
	var flag *GovernanceFeatureFlag
	for i := range flags {
		if flags[i].FlagKey == flagKey {
			flag = &flags[i]
			break
		}
	}
	if flag == nil {
		flag = &GovernanceFeatureFlag{
			OrgID:            req.OrgID,
			FlagKey:          flagKey,
			FlagName:         flagKey,
			IsEnabled:        true,
			MaxAutonomyLevel: 3,
			RequiresApproval: true,
		}
	}

	if !flag.IsEnabled {
		decision = GovernanceDecisionBlock
		reasons = append(reasons, fmt.Sprintf("Autonomous capability '%s' is disabled by tenant feature flag.", flag.FlagName))
		resp := &GovernanceEvaluationResponse{
			Decision:          decision,
			EffectiveAutonomy: 0,
			RiskClass:         RiskClassHigh,
			DataSufficiency:   SufficiencySufficient,
			ConfidenceClass:   ConfidenceClassHigh,
			RequiresApproval:  true,
			FourEyesEnforced:  false,
			Reversibility:     ReversibilityReversible,
			Reasons:           reasons,
			PolicyVersion:     policyVersion,
			CorrelationID:     corrID,
		}
		e.recordEvaluation(ctx, req, resp)
		return resp, nil
	}

	// 3. Calculate Effective Autonomy: min(requested, tenantMax, flagMax)
	effectiveAutonomy := req.RequestedAutonomy
	if effectiveAutonomy <= 0 {
		effectiveAutonomy = 2 // default to prepare
	}
	if effectiveAutonomy > limits.MaxTenantAutonomy {
		effectiveAutonomy = limits.MaxTenantAutonomy
		reasons = append(reasons, fmt.Sprintf("Requested autonomy capped by tenant maximum (%d).", limits.MaxTenantAutonomy))
	}
	if effectiveAutonomy > flag.MaxAutonomyLevel {
		effectiveAutonomy = flag.MaxAutonomyLevel
		reasons = append(reasons, fmt.Sprintf("Requested autonomy capped by module feature flag maximum (%d).", flag.MaxAutonomyLevel))
	}

	// 4. Action Allowlist Lookup
	allowlistItem, err := e.repo.GetActionAllowlistItem(ctx, req.OrgID, req.ActionType)
	if err != nil || allowlistItem == nil || !allowlistItem.IsEnabled {
		decision = GovernanceDecisionBlock
		reasons = append(reasons, fmt.Sprintf("Action type '%s' is not registered or enabled in tenant allowlist.", req.ActionType))
		resp := &GovernanceEvaluationResponse{
			Decision:          decision,
			EffectiveAutonomy: effectiveAutonomy,
			RiskClass:         RiskClassHigh,
			DataSufficiency:   SufficiencyInsufficient,
			ConfidenceClass:   ConfidenceClassLow,
			RequiresApproval:  true,
			FourEyesEnforced:  false,
			Reversibility:     ReversibilityIrreversible,
			Reasons:           reasons,
			PolicyVersion:     policyVersion,
			CorrelationID:     corrID,
		}
		e.recordEvaluation(ctx, req, resp)
		return resp, nil
	}

	riskClass = allowlistItem.RiskClass
	reversibility = allowlistItem.Reversibility

	// Verify effective autonomy satisfies action's allowed tiers
	var allowedTiers []int
	_ = json.Unmarshal(allowlistItem.AllowedAutonomyLevels, &allowedTiers)
	tierAllowed := false
	for _, t := range allowedTiers {
		if effectiveAutonomy >= t {
			tierAllowed = true
			break
		}
	}
	if !tierAllowed {
		decision = GovernanceDecisionRequireReview
		approvalRequired = true
		reasons = append(reasons, fmt.Sprintf("Action '%s' requires higher autonomy tier than effective autonomy %d. Operator review required.", req.ActionType, effectiveAutonomy))
	}

	// 5. User Permissions & RBAC Enforcement
	if allowlistItem.RequiredPermission != "" {
		hasPerm := false
		for _, p := range req.UserPermissions {
			if p == "*" || p == allowlistItem.RequiredPermission {
				hasPerm = true
				break
			}
		}
		if !hasPerm {
			decision = GovernanceDecisionBlock
			reasons = append(reasons, fmt.Sprintf("User lacks required permission '%s' for action '%s'.", allowlistItem.RequiredPermission, req.ActionType))
		}
	}

	// 6. Role-Level Authority Segregation
	roleLower := strings.ToLower(req.UserRole)
	if (req.Module == "finance" || req.Module == "pricing") && !strings.Contains(roleLower, "admin") && !strings.Contains(roleLower, "finance") && !strings.Contains(roleLower, "super") {
		decision = GovernanceDecisionBlock
		reasons = append(reasons, fmt.Sprintf("Role '%s' is not authorized to approve financial or commercial actions.", req.UserRole))
	}
	if req.Module == "compliance" && !strings.Contains(roleLower, "admin") && !strings.Contains(roleLower, "compliance") && !strings.Contains(roleLower, "super") {
		decision = GovernanceDecisionBlock
		reasons = append(reasons, fmt.Sprintf("Role '%s' is not authorized to approve regulatory compliance actions.", req.UserRole))
	}

	// 7. High-Impact Action & Reversibility Checks
	if riskClass == RiskClassHigh || riskClass == RiskClassCritical || reversibility == ReversibilityIrreversible {
		if !req.IsApproved {
			if decision != GovernanceDecisionBlock {
				decision = GovernanceDecisionRequireReview
			}
			approvalRequired = true
			reasons = append(reasons, fmt.Sprintf("Action '%s' is %s risk / %s and strictly requires human approval.", req.ActionType, riskClass, reversibility))
		}
	}

	// 8. Four-Eyes Control Enforcement
	if limits.EnforceFourEyes && (riskClass == RiskClassHigh || riskClass == RiskClassCritical) {
		fourEyesEnforced = true
		if req.IsApproved && req.PreparerUserID > 0 && req.ApprovalApproverID > 0 && req.PreparerUserID == req.ApprovalApproverID {
			decision = GovernanceDecisionBlock
			reasons = append(reasons, "Four-Eyes Control Violation: The user who prepared the high-impact action cannot approve their own action.")
		}
	}

	// 9. Financial Exposure Bounds
	finAmount := extractFinancialAmount(req.Parameters)
	if finAmount > limits.MaxFinancialExposurePerWorkflow {
		if decision != GovernanceDecisionBlock {
			decision = GovernanceDecisionRequireReview
		}
		approvalRequired = true
		reasons = append(reasons, fmt.Sprintf("Financial exposure ($%.2f) exceeds tenant autonomous limit ($%.2f).", finAmount, limits.MaxFinancialExposurePerWorkflow))
	}
	if allowlistItem.MaxFinancialLimit > 0 && finAmount > allowlistItem.MaxFinancialLimit {
		if decision != GovernanceDecisionBlock {
			decision = GovernanceDecisionRequireReview
		}
		approvalRequired = true
		reasons = append(reasons, fmt.Sprintf("Financial exposure ($%.2f) exceeds action-specific limit ($%.2f).", finAmount, allowlistItem.MaxFinancialLimit))
	}

	// 10. Compliance Block Gate
	if strings.ToUpper(req.ComplianceStatus) == "BLOCKED" {
		decision = GovernanceDecisionBlock
		reasons = append(reasons, "Associated operational entity has a BLOCKED compliance status. Autonomous execution is strictly prohibited.")
	}

	// 11. Entity State Validation
	terminalStates := map[string]bool{
		"CANCELLED": true, "CANCELED": true, "DELETED": true, "TERMINATED": true, "SETTLED": true, "LOCKED": true,
	}
	if terminalStates[strings.ToUpper(req.EntityState)] {
		decision = GovernanceDecisionBlock
		reasons = append(reasons, fmt.Sprintf("Entity is in terminal state '%s'. Action '%s' cannot be applied.", req.EntityState, req.ActionType))
	}

	// 12. Sidecar Context Evaluation (Risk, Sufficiency, Prompt Injection)
	if e.sidecarClient != nil {
		sidecarResp, sErr := e.sidecarClient.EvaluateGovernanceContext(ctx, SidecarGovernanceEvalRequest{
			OrgID:              req.OrgID,
			Module:             req.Module,
			ActionType:         req.ActionType,
			EntityType:         req.EntityType,
			EntityID:           req.EntityID,
			Parameters:         req.Parameters,
			ContextText:        req.ContextText,
			RequestedAutonomy:  req.RequestedAutonomy,
			ConfiguredAutonomy: effectiveAutonomy,
			CorrelationID:      corrID,
		})
		if sErr == nil && sidecarResp != nil {
			sufficiency = DataSufficiencyState(sidecarResp.DataSufficiency)
			confidence = ConfidenceClass(sidecarResp.ConfidenceClass)

			if sufficiency == SufficiencyInsufficient {
				decision = GovernanceDecisionBlock
				reasons = append(reasons, fmt.Sprintf("Data sufficiency check failed: missing required information %v.", sidecarResp.MissingDataFields))
			} else if sufficiency == SufficiencyPartiallySufficient && decision == GovernanceDecisionAllow {
				decision = GovernanceDecisionRequireReview
				approvalRequired = true
				reasons = append(reasons, "Partial data sufficiency. Human verification required.")
			}

			if confidence == ConfidenceClassLow && riskClass != RiskClassLow && decision == GovernanceDecisionAllow {
				decision = GovernanceDecisionRequireReview
				approvalRequired = true
				reasons = append(reasons, "Low confidence assessment from sidecar. Review required.")
			}

			if sidecarResp.Explanation != "" {
				reasons = append(reasons, sidecarResp.Explanation)
			}
		}
	}

	// 13. Final approval requirement check
	if allowlistItem.ApprovalRequirement == ApprovalAlways && !req.IsApproved {
		if decision != GovernanceDecisionBlock {
			decision = GovernanceDecisionRequireReview
		}
		approvalRequired = true
		reasons = append(reasons, "Action policy requires explicit human approval under all conditions.")
	}

	// Fail-safe guarantee: if reasons are empty and decision is ALLOW, verify state is sound
	if decision == GovernanceDecisionAllow && len(reasons) == 0 {
		reasons = append(reasons, fmt.Sprintf("Action '%s' verified and authorized under Level %d autonomy.", req.ActionType, effectiveAutonomy))
	}

	resp := &GovernanceEvaluationResponse{
		Decision:          decision,
		EffectiveAutonomy: effectiveAutonomy,
		RiskClass:         riskClass,
		DataSufficiency:   sufficiency,
		ConfidenceClass:   confidence,
		RequiresApproval:  approvalRequired,
		FourEyesEnforced:  fourEyesEnforced,
		Reversibility:     reversibility,
		Reasons:           reasons,
		PolicyVersion:     policyVersion,
		CorrelationID:     corrID,
	}

	e.recordEvaluation(ctx, req, resp)
	return resp, nil
}

func (e *governanceEngine) PreviewPlan(ctx context.Context, req PlanPreviewRequest) (*PlanPreviewResponse, error) {
	corrID := req.CorrelationID
	if corrID == "" {
		corrID = fmt.Sprintf("corr-prev-%d", time.Now().UnixNano())
	}

	if e.sidecarClient != nil {
		sResp, err := e.sidecarClient.PreviewGovernancePlan(ctx, SidecarPlanPreviewRequest{
			OrgID:         req.OrgID,
			PlanID:        req.PlanID,
			Goal:          req.Goal,
			Module:        req.Module,
			Steps:         req.Steps,
			CorrelationID: corrID,
		})
		if err == nil && sResp != nil {
			return &PlanPreviewResponse{
				PlanID:                        sResp.PlanID,
				TotalSteps:                    sResp.TotalSteps,
				ExecutableSteps:               sResp.ExecutableSteps,
				ApprovalRequiredSteps:         sResp.ApprovalRequiredSteps,
				MaxRiskClass:                  sResp.MaxRiskClass,
				EstimatedFinancialExposureUSD: sResp.EstimatedFinancialExposureUSD,
				AffectedEntities:              sResp.AffectedEntities,
				SafetySummary:                 sResp.SafetySummary,
				CorrelationID:                 corrID,
			}, nil
		}
	}

	// Local fallback preview
	totalSteps := len(req.Steps)
	execSteps := 0
	appSteps := 0
	exposure := 0.0
	for _, step := range req.Steps {
		cost := extractFinancialAmount(step)
		exposure += cost
		if reqVal, ok := step["requires_approval"].(bool); ok && reqVal {
			appSteps++
		} else {
			execSteps++
		}
	}

	return &PlanPreviewResponse{
		PlanID:                        req.PlanID,
		TotalSteps:                    totalSteps,
		ExecutableSteps:               execSteps,
		ApprovalRequiredSteps:         appSteps,
		MaxRiskClass:                  "MEDIUM",
		EstimatedFinancialExposureUSD: exposure,
		AffectedEntities:              []map[string]string{{"entity_type": "plan", "entity_id": req.PlanID}},
		SafetySummary:                 fmt.Sprintf("Plan preview: %d steps, %d approval gates, $%.2f exposure.", totalSteps, appSteps, exposure),
		CorrelationID:                 corrID,
	}, nil
}

func (e *governanceEngine) recordEvaluation(ctx context.Context, req GovernanceEvaluationRequest, resp *GovernanceEvaluationResponse) {
	reasonsJSON, _ := json.Marshal(resp.Reasons)
	var uID *int64
	if req.UserID > 0 {
		uID = &req.UserID
	}
	rec := &PolicyEvaluationRecord{
		OrgID:             req.OrgID,
		UserID:            uID,
		EntityType:        req.EntityType,
		EntityID:          req.EntityID,
		ActionType:        req.ActionType,
		Module:            req.Module,
		RequestedAutonomy: req.RequestedAutonomy,
		EffectiveAutonomy: resp.EffectiveAutonomy,
		Decision:          resp.Decision,
		Reasons:           reasonsJSON,
		RiskLevel:         resp.RiskClass,
		DataSufficiency:   resp.DataSufficiency,
		Confidence:        resp.ConfidenceClass,
		ApprovalRequired:  resp.RequiresApproval,
		FourEyesRequired:  resp.FourEyesEnforced,
		PolicyVersion:     resp.PolicyVersion,
		CorrelationID:     resp.CorrelationID,
	}
	_ = e.repo.RecordPolicyEvaluation(ctx, rec)
}

func deriveFlagKey(module string) string {
	switch strings.ToLower(module) {
	case "shipments", "shipment":
		return "shipment_automation"
	case "customers", "customer", "customer_followup":
		return "customer_automation"
	case "finance", "finance_collections":
		return "finance_automation"
	case "pricing", "rfq":
		return "pricing_optimization"
	case "compliance", "contracts", "contract_compliance":
		return "contract_compliance"
	case "exceptions", "exception":
		return "exception_resolution"
	case "multistep", "multistep_workflows":
		return "multistep_workflows"
	default:
		return "continuous_monitoring"
	}
}

func extractFinancialAmount(params map[string]interface{}) float64 {
	if params == nil {
		return 0.0
	}
	for _, key := range []string{"amount", "discount_amount", "cost_impact", "total_amount", "balance_due"} {
		if val, exists := params[key]; exists {
			switch v := val.(type) {
			case float64:
				return v
			case float32:
				return float64(v)
			case int:
				return float64(v)
			case int64:
				return float64(v)
			}
		}
	}
	return 0.0
}

// PlanPreviewRequest DTO
type PlanPreviewRequest struct {
	OrgID         int64                    `json:"org_id"`
	PlanID        string                   `json:"plan_id"`
	Goal          string                   `json:"goal"`
	Module        string                   `json:"module"`
	Steps         []map[string]interface{} `json:"steps"`
	CorrelationID string                   `json:"correlation_id"`
}

// PlanPreviewResponse DTO
type PlanPreviewResponse struct {
	PlanID                        string              `json:"plan_id"`
	TotalSteps                    int                 `json:"total_steps"`
	ExecutableSteps               int                 `json:"executable_steps"`
	ApprovalRequiredSteps         int                 `json:"approval_required_steps"`
	MaxRiskClass                  string              `json:"max_risk_class"`
	EstimatedFinancialExposureUSD float64             `json:"estimated_financial_exposure_usd"`
	AffectedEntities              []map[string]string `json:"affected_entities"`
	SafetySummary                 string              `json:"safety_summary"`
	CorrelationID                 string              `json:"correlation_id"`
}
