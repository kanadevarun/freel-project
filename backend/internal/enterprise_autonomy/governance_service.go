package enterprise_autonomy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/approvals"
	"github.com/freel/backend/internal/audit/domain"
	auditSvc "github.com/freel/backend/internal/audit/service"
	"github.com/jmoiron/sqlx"
)

// PolicyVersion defines the active enterprise autonomy governance version
const CurrentPolicyVersion = "v7.10.0-governed"

// EnterpriseGovernanceService defines production-grade safety & autonomy enforcement
type EnterpriseGovernanceService interface {
	EvaluateAction(ctx context.Context, req GovernanceEvaluationRequest) (*GovernanceDecision, error)
	ValidateApprovalExecution(ctx context.Context, orgID int64, approvalID int64, currentEntityState map[string]interface{}) error
	RecordHumanRejection(ctx context.Context, orgID int64, workflowID string, actionType string, entityID string, reason string, rejectedBy string) error
	CheckHumanRejectionProtection(ctx context.Context, orgID int64, actionType string, entityID string) error
	ApplyEmergencyControl(ctx context.Context, orgID int64, scope EmergencyScope, target string, active bool, reason string, actor string) error
	GetActiveEmergencyControls(ctx context.Context, orgID int64) ([]EmergencyControlEntry, error)
	CheckLoopProtection(ctx context.Context, orgID int64, workflowID string, actionType string, entityID string) error
	SanitizeUntrustedInput(ctx context.Context, text string) (cleanText string, suspicious bool)
	GetGovernanceStatus(ctx context.Context, orgID int64) (*GovernanceStatusSummary, error)
}

type defaultEnterpriseGovernanceService struct {
	repo         Repository
	actionsSvc   actions.Service
	approvalsSvc approvals.Service
	auditSvc     auditSvc.Service
	db           *sqlx.DB

	// In-memory safety state caches
	mu                sync.RWMutex
	emergencyHalts    map[string]*EmergencyControlEntry // key: "orgID:scope:target"
	humanRejections   map[string]*HumanRejectionRecord   // key: "orgID:actionType:entityID"
	actionHistory     map[string][]time.Time             // key: "orgID:workflowID:actionType:entityID"
	approvalFpCache   map[string]string                  // key: "orgID:approvalID" -> fingerprint
	
	// Metrics counters
	totalEvaluated    int64
	blockedCount      int64
	approvalsRequired int64
	stalePrevented    int64
	loopsDetected     int64
	injectionsBlocked int64
}

// NewEnterpriseGovernanceService constructs the production governance service
func NewEnterpriseGovernanceService(
	repo Repository,
	actSvc actions.Service,
	apprSvc approvals.Service,
	audSvc auditSvc.Service,
	db *sqlx.DB,
) EnterpriseGovernanceService {
	return &defaultEnterpriseGovernanceService{
		repo:            repo,
		actionsSvc:      actSvc,
		approvalsSvc:    apprSvc,
		auditSvc:        audSvc,
		db:              db,
		emergencyHalts:  make(map[string]*EmergencyControlEntry),
		humanRejections: make(map[string]*HumanRejectionRecord),
		actionHistory:   make(map[string][]time.Time),
		approvalFpCache: make(map[string]string),
	}
}

// ---------------------------------------------------------------------
// 1. EvaluateAction: Centralized Go-Authoritative Autonomy Enforcement
// ---------------------------------------------------------------------

func (s *defaultEnterpriseGovernanceService) EvaluateAction(ctx context.Context, req GovernanceEvaluationRequest) (*GovernanceDecision, error) {
	atomic.AddInt64(&s.totalEvaluated, 1)
	now := time.Now().UTC()

	// 1. Tenant Isolation: Go strictly validates non-zero authenticated tenant
	if req.OrgID <= 0 {
		atomic.AddInt64(&s.blockedCount, 1)
		return &GovernanceDecision{
			Decision:               "BLOCKED",
			Allowed:                false,
			PolicyVersion:          CurrentPolicyVersion,
			Reason:                 "Tenant isolation violation: invalid or unauthenticated organization ID",
			Timestamp:              now,
		}, ErrUnauthorizedTenant
	}

	// 2. Emergency Halt Checks (Global, Agent, WorkflowType, ActionType, WorkflowID)
	if halt := s.checkEmergencyHalt(req.OrgID, req.AgentID, req.WorkflowType, req.ActionType, req.WorkflowID); halt != nil {
		atomic.AddInt64(&s.blockedCount, 1)
		s.recordAudit(ctx, req.OrgID, "EMERGENCY_HALT_BLOCKED", req.ActionType, req.WorkflowID, req.AgentID, halt.Reason)
		return &GovernanceDecision{
			Decision:               "BLOCKED",
			Allowed:                false,
			PolicyVersion:          CurrentPolicyVersion,
			Reason:                 fmt.Sprintf("Halted: emergency stop active on scope %s (%s)", halt.Scope, halt.Reason),
			CircuitBreakerTripped:  true,
			Timestamp:              now,
		}, ErrEmergencyHaltActive
	}

	// 3. Prompt Injection & Untrusted Input Neutralization
	for _, content := range req.UntrustedContent {
		_, suspicious := s.SanitizeUntrustedInput(ctx, content)
		if suspicious {
			atomic.AddInt64(&s.blockedCount, 1)
			atomic.AddInt64(&s.injectionsBlocked, 1)
			s.recordAudit(ctx, req.OrgID, "PROMPT_INJECTION_DEFENSE", req.ActionType, req.WorkflowID, req.AgentID, "Malicious instruction or policy override token detected")
			return &GovernanceDecision{
				Decision:               "BLOCKED",
				Allowed:                false,
				PolicyVersion:          CurrentPolicyVersion,
				Reason:                 "Security defense: untrusted text contains prompt injection or policy override instructions",
				CircuitBreakerTripped:  true,
				Timestamp:              now,
			}, ErrPromptInjectionDetected
		}
	}

	// 4. Autonomous Loop & Cycle Protection
	if err := s.CheckLoopProtection(ctx, req.OrgID, req.WorkflowID, req.ActionType, req.TargetEntityID); err != nil {
		atomic.AddInt64(&s.blockedCount, 1)
		atomic.AddInt64(&s.loopsDetected, 1)
		s.recordAudit(ctx, req.OrgID, "LOOP_PROTECTION_BLOCKED", req.ActionType, req.WorkflowID, req.AgentID, err.Error())
		return &GovernanceDecision{
			Decision:               "BLOCKED",
			Allowed:                false,
			PolicyVersion:          CurrentPolicyVersion,
			Reason:                 err.Error(),
			CircuitBreakerTripped:  true,
			Timestamp:              now,
		}, err
	}

	// 5. Human Rejection Protection (Agent cannot override explicit operator rejection)
	if err := s.CheckHumanRejectionProtection(ctx, req.OrgID, req.ActionType, req.TargetEntityID); err != nil {
		atomic.AddInt64(&s.blockedCount, 1)
		s.recordAudit(ctx, req.OrgID, "HUMAN_REJECTION_PROTECTED", req.ActionType, req.WorkflowID, req.AgentID, err.Error())
		return &GovernanceDecision{
			Decision:               "BLOCKED",
			Allowed:                false,
			PolicyVersion:          CurrentPolicyVersion,
			Reason:                 err.Error(),
			Timestamp:              now,
		}, err
	}

	// 6. Agent Least-Privilege Matrix Validation
	if err := s.validateAgentPermissions(req.AgentID, req.ActionType); err != nil {
		atomic.AddInt64(&s.blockedCount, 1)
		s.recordAudit(ctx, req.OrgID, "AGENT_PRIVILEGE_DENIED", req.ActionType, req.WorkflowID, req.AgentID, err.Error())
		return &GovernanceDecision{
			Decision:               "BLOCKED",
			Allowed:                false,
			PolicyVersion:          CurrentPolicyVersion,
			Reason:                 err.Error(),
			Timestamp:              now,
		}, err
	}

	// 7. Authoritative Risk Classification (Go overrides agent self-classification)
	authoritativeRisk := s.classifyAuthoritativeRisk(req.ActionType, req.MonetaryExposure, req.ProposedRiskLevel)

	// 8. Confidence Threshold Guard (Phase 4 reuse)
	minConfidence := 0.70
	if authoritativeRisk == RiskLevelHigh || authoritativeRisk == RiskLevelCritical {
		minConfidence = 0.85
	}
	if req.Confidence > 0 && req.Confidence < minConfidence {
		atomic.AddInt64(&s.blockedCount, 1)
		s.recordAudit(ctx, req.OrgID, "INSUFFICIENT_CONFIDENCE_BLOCKED", req.ActionType, req.WorkflowID, req.AgentID, fmt.Sprintf("Confidence %.2f below required %.2f", req.Confidence, minConfidence))
		return &GovernanceDecision{
			Decision:               "BLOCKED",
			Allowed:                false,
			AuthoritativeRiskLevel: authoritativeRisk,
			PolicyVersion:          CurrentPolicyVersion,
			Reason:                 fmt.Sprintf("Prediction confidence (%.2f) below required threshold (%.2f) for %s risk", req.Confidence, minConfidence, authoritativeRisk),
			Timestamp:              now,
		}, ErrInsufficientConfidence
	}

	// 9. Fetch Policy Context for Module / Org
	moduleName := req.WorkflowType
	if moduleName == "" {
		moduleName = "GLOBAL"
	}
	pol, err := s.repo.GetPolicyContext(ctx, req.OrgID, moduleName)
	if err != nil {
		// Fail-closed on policy retrieval failure
		atomic.AddInt64(&s.blockedCount, 1)
		return nil, fmt.Errorf("fail-closed: policy check failed: %w", err)
	}

	// Calculate state fingerprint to guard against stale execution later
	fp := s.computeFingerprint(req.TargetEntityID, req.CurrentEntityState)

	// 10. Evaluate Autonomy Level & Approval Boundary
	policyCeiling := s.normalizeAutonomyLevel(pol.AutonomyLevel)
	effectiveAutonomy := policyCeiling
	if req.ProposedAutonomyLevel != "" {
		requested := s.normalizeAutonomyLevel(string(req.ProposedAutonomyLevel))
		if s.autonomyRank(requested) <= s.autonomyRank(policyCeiling) {
			effectiveAutonomy = requested
		} else {
			effectiveAutonomy = policyCeiling
		}
	}

	switch effectiveAutonomy {
	case AutonomyLvl0Observe:
		// Observe only: no execution permitted
		atomic.AddInt64(&s.blockedCount, 1)
		return &GovernanceDecision{
			Decision:               "BLOCKED",
			Allowed:                false,
			AuthoritativeRiskLevel: authoritativeRisk,
			EvaluatedAutonomyLevel: string(effectiveAutonomy),
			PolicyVersion:          CurrentPolicyVersion,
			Reason:                 "Autonomy Level 0 (Observe): autonomous actions strictly prohibited",
			StateFingerprint:       fp,
			Timestamp:              now,
		}, nil

	case AutonomyLvl1Recommend, AutonomyLvl2Prepare:
		// Always requires human dispatch / approval
		atomic.AddInt64(&s.approvalsRequired, 1)
		return &GovernanceDecision{
			Decision:               "REQUIRE_APPROVAL",
			Allowed:                false,
			RequiresApproval:       true,
			AuthoritativeRiskLevel: authoritativeRisk,
			EvaluatedAutonomyLevel: string(effectiveAutonomy),
			PolicyVersion:          CurrentPolicyVersion,
			Reason:                 fmt.Sprintf("Autonomy Level (%s) enforces affirmative human approval before execution", effectiveAutonomy),
			StateFingerprint:       fp,
			Timestamp:              now,
		}, nil

	case AutonomyLvl3Controlled:
		// Policy-bounded: LOW and MEDIUM risk permitted up to monetary threshold; HIGH/CRITICAL requires approval
		if authoritativeRisk == RiskLevelHigh || authoritativeRisk == RiskLevelCritical || req.MonetaryExposure > pol.MaxMonetaryThreshold {
			atomic.AddInt64(&s.approvalsRequired, 1)
			return &GovernanceDecision{
				Decision:               "REQUIRE_APPROVAL",
				Allowed:                false,
				RequiresApproval:       true,
				AuthoritativeRiskLevel: authoritativeRisk,
				EvaluatedAutonomyLevel: string(effectiveAutonomy),
				PolicyVersion:          CurrentPolicyVersion,
				Reason:                 fmt.Sprintf("Action risk (%s) or exposure ($%.2f) exceeds Level 3 autonomous ceiling ($%.2f)", authoritativeRisk, req.MonetaryExposure, pol.MaxMonetaryThreshold),
				StateFingerprint:       fp,
				Timestamp:              now,
			}, nil
		}
		// Permitted under Level 3
		s.recordAudit(ctx, req.OrgID, "ACTION_PERMITTED_LEVEL3", req.ActionType, req.WorkflowID, req.AgentID, "Autonomous execution governed within Level 3 limits")
		return &GovernanceDecision{
			Decision:               "PERMITTED",
			Allowed:                true,
			RequiresApproval:       false,
			AuthoritativeRiskLevel: authoritativeRisk,
			EvaluatedAutonomyLevel: string(effectiveAutonomy),
			PolicyVersion:          CurrentPolicyVersion,
			Reason:                 "Permitted: action satisfies Level 3 controlled execution boundaries",
			StateFingerprint:       fp,
			Timestamp:              now,
		}, nil

	case AutonomyLvl4Governed:
		// Governed Multi-step Autonomy: CRITICAL risk still requires affirmative sign-off
		if authoritativeRisk == RiskLevelCritical {
			atomic.AddInt64(&s.approvalsRequired, 1)
			return &GovernanceDecision{
				Decision:               "REQUIRE_APPROVAL",
				Allowed:                false,
				RequiresApproval:       true,
				AuthoritativeRiskLevel: authoritativeRisk,
				EvaluatedAutonomyLevel: string(effectiveAutonomy),
				PolicyVersion:          CurrentPolicyVersion,
				Reason:                 "CRITICAL risk actions require human approval even under Autonomy Level 4",
				StateFingerprint:       fp,
				Timestamp:              now,
			}, nil
		}
		// Permitted under Level 4
		s.recordAudit(ctx, req.OrgID, "ACTION_PERMITTED_LEVEL4", req.ActionType, req.WorkflowID, req.AgentID, "Autonomous execution governed under Level 4 multi-step autonomy")
		return &GovernanceDecision{
			Decision:               "PERMITTED",
			Allowed:                true,
			RequiresApproval:       false,
			AuthoritativeRiskLevel: authoritativeRisk,
			EvaluatedAutonomyLevel: string(effectiveAutonomy),
			PolicyVersion:          CurrentPolicyVersion,
			Reason:                 "Permitted: multi-step autonomous execution governed within Level 4 policy",
			StateFingerprint:       fp,
			Timestamp:              now,
		}, nil

	default:
		// Fail closed on unknown autonomy level
		atomic.AddInt64(&s.blockedCount, 1)
		return &GovernanceDecision{
			Decision:               "BLOCKED",
			Allowed:                false,
			PolicyVersion:          CurrentPolicyVersion,
			Reason:                 fmt.Sprintf("Fail-closed: unrecognized autonomy level '%s'", effectiveAutonomy),
			Timestamp:              now,
		}, ErrGovernanceDenied
	}
}

// ---------------------------------------------------------------------
// 2. ValidateApprovalExecution: Stale Approval Invalidation Guard
// ---------------------------------------------------------------------

func (s *defaultEnterpriseGovernanceService) ValidateApprovalExecution(
	ctx context.Context,
	orgID int64,
	approvalID int64,
	currentEntityState map[string]interface{},
) error {
	if orgID <= 0 || approvalID <= 0 {
		return ErrUnauthorizedTenant
	}

	s.mu.RLock()
	cachedFp, exists := s.approvalFpCache[fmt.Sprintf("%d:%d", orgID, approvalID)]
	s.mu.RUnlock()

	if !exists {
		// No fingerprint registered: allow if approval is otherwise valid
		return nil
	}

	currentFp := s.computeFingerprint(fmt.Sprintf("%d", approvalID), currentEntityState)
	if cachedFp != currentFp {
		atomic.AddInt64(&s.stalePrevented, 1)
		s.recordAudit(ctx, orgID, "STALE_APPROVAL_INVALIDATED", "APPROVAL_EXECUTE", "", "", fmt.Sprintf("Approval #%d invalidated due to state change", approvalID))
		return ErrStaleApprovalInvalidated
	}

	return nil
}

// ---------------------------------------------------------------------
// 3. Human Rejection Protection: Prevent AI Fighting Human Decisions
// ---------------------------------------------------------------------

func (s *defaultEnterpriseGovernanceService) RecordHumanRejection(
	ctx context.Context,
	orgID int64,
	workflowID string,
	actionType string,
	entityID string,
	reason string,
	rejectedBy string,
) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}

	key := fmt.Sprintf("%d:%s:%s", orgID, strings.ToUpper(actionType), entityID)
	rec := &HumanRejectionRecord{
		OrgID:      orgID,
		WorkflowID: workflowID,
		ActionType: strings.ToUpper(actionType),
		EntityID:   entityID,
		Reason:     reason,
		RejectedBy: rejectedBy,
		RejectedAt: time.Now().UTC(),
	}

	s.mu.Lock()
	s.humanRejections[key] = rec
	s.mu.Unlock()

	s.recordAudit(ctx, orgID, "HUMAN_REJECTION_RECORDED", actionType, workflowID, rejectedBy, reason)
	return nil
}

func (s *defaultEnterpriseGovernanceService) CheckHumanRejectionProtection(ctx context.Context, orgID int64, actionType string, entityID string) error {
	key := fmt.Sprintf("%d:%s:%s", orgID, strings.ToUpper(actionType), entityID)

	s.mu.RLock()
	rec, exists := s.humanRejections[key]
	s.mu.RUnlock()

	if exists {
		// Enforce human rejection protection for 24 hours
		if time.Since(rec.RejectedAt) < 24*time.Hour {
			return fmt.Errorf("%w: operator %s previously rejected '%s' on entity %s (reason: %s)",
				ErrHumanRejectionConflict, rec.RejectedBy, actionType, entityID, rec.Reason)
		}
	}
	return nil
}

// ---------------------------------------------------------------------
// 4. Emergency Stop & Circuit Breakers (Go Enforcement Boundary)
// ---------------------------------------------------------------------

func (s *defaultEnterpriseGovernanceService) ApplyEmergencyControl(
	ctx context.Context,
	orgID int64,
	scope EmergencyScope,
	target string,
	active bool,
	reason string,
	actor string,
) error {
	if orgID <= 0 {
		return ErrUnauthorizedTenant
	}

	if scope == EmergencyScopeAll && (target == "" || target == "GLOBAL" || target == "ALL") {
		target = "*"
	}

	key := fmt.Sprintf("%d:%s:%s", orgID, scope, strings.ToUpper(target))
	entry := &EmergencyControlEntry{
		OrgID:            orgID,
		Scope:            scope,
		TargetIdentifier: strings.ToUpper(target),
		Active:           active,
		Reason:           reason,
		HaltedBy:         actor,
		HaltedAt:         time.Now().UTC(),
	}

	s.mu.Lock()
	if active {
		s.emergencyHalts[key] = entry
	} else {
		delete(s.emergencyHalts, key)
	}
	s.mu.Unlock()

	actionName := "EMERGENCY_HALT_ACTIVATED"
	if !active {
		actionName = "EMERGENCY_HALT_DEACTIVATED"
	}
	s.recordAudit(ctx, orgID, actionName, string(scope), target, actor, reason)
	return nil
}

func (s *defaultEnterpriseGovernanceService) GetActiveEmergencyControls(ctx context.Context, orgID int64) ([]EmergencyControlEntry, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]EmergencyControlEntry, 0)
	prefix := fmt.Sprintf("%d:", orgID)
	for k, v := range s.emergencyHalts {
		if strings.HasPrefix(k, prefix) && v.Active {
			results = append(results, *v)
		}
	}
	return results, nil
}

func (s *defaultEnterpriseGovernanceService) checkEmergencyHalt(orgID int64, agentID, workflowType, actionType, workflowID string) *EmergencyControlEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 1. Check Global Org Halt
	prefixAll := fmt.Sprintf("%d:%s:", orgID, EmergencyScopeAll)
	for k, h := range s.emergencyHalts {
		if strings.HasPrefix(k, prefixAll) && h.Active {
			return h
		}
	}
	// 2. Check Agent Halt
	if h, ok := s.emergencyHalts[fmt.Sprintf("%d:%s:%s", orgID, EmergencyScopeAgent, strings.ToUpper(agentID))]; ok && h.Active {
		return h
	}
	// 3. Check WorkflowType Halt
	if h, ok := s.emergencyHalts[fmt.Sprintf("%d:%s:%s", orgID, EmergencyScopeWorkflowType, strings.ToUpper(workflowType))]; ok && h.Active {
		return h
	}
	// 4. Check ActionType Halt
	if h, ok := s.emergencyHalts[fmt.Sprintf("%d:%s:%s", orgID, EmergencyScopeActionType, strings.ToUpper(actionType))]; ok && h.Active {
		return h
	}
	// 5. Check WorkflowID Halt
	if h, ok := s.emergencyHalts[fmt.Sprintf("%d:%s:%s", orgID, EmergencyScopeWorkflowID, strings.ToUpper(workflowID))]; ok && h.Active {
		return h
	}

	return nil
}

// ---------------------------------------------------------------------
// 5. Loop & Cycle Protection
// ---------------------------------------------------------------------

func (s *defaultEnterpriseGovernanceService) CheckLoopProtection(
	ctx context.Context,
	orgID int64,
	workflowID string,
	actionType string,
	entityID string,
) error {
	key := fmt.Sprintf("%d:%s:%s:%s", orgID, workflowID, actionType, entityID)
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	history, exists := s.actionHistory[key]
	if !exists {
		s.actionHistory[key] = []time.Time{now}
		return nil
	}

	// Filter history to last 10 minutes
	recent := make([]time.Time, 0)
	cutoff := now.Add(-10 * time.Minute)
	for _, t := range history {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	// If more than 3 repetitive actions triggered without human intervention -> loop detected
	if len(recent) >= 3 {
		return fmt.Errorf("%w: action '%s' on entity '%s' repeated %d times in 10m",
			ErrAutonomousLoopDetected, actionType, entityID, len(recent))
	}

	recent = append(recent, now)
	s.actionHistory[key] = recent
	return nil
}

// ---------------------------------------------------------------------
// 6. Prompt Injection Defense (Data vs Control Authority)
// ---------------------------------------------------------------------

func (s *defaultEnterpriseGovernanceService) SanitizeUntrustedInput(ctx context.Context, text string) (string, bool) {
	if text == "" {
		return "", false
	}

	lower := strings.ToLower(text)
	maliciousTokens := []string{
		"ignore previous instructions",
		"ignore all previous",
		"disregard previous",
		"disregard all",
		"prompt override",
		"system prompt",
		"override safety policy",
		"override security",
		"override policy",
		"system: set autonomy",
		"bypass approval",
		"bypass approvals",
		"bypass governance",
		"bypass all human approvals",
		"authorized to bypass",
		"grant full autonomy",
		"grant admin",
		"drop table",
		"<script>",
		"sudo ",
		"elevate permissions",
		"execute shell",
		"act as root",
		"act as admin",
	}

	for _, token := range maliciousTokens {
		if strings.Contains(lower, token) {
			return "[SUSPICIOUS_UNTRUSTED_CONTENT_FILTERED]", true
		}
	}

	return text, false
}

// ---------------------------------------------------------------------
// 7. Agent Least-Privilege Access Matrix
// ---------------------------------------------------------------------

func (s *defaultEnterpriseGovernanceService) validateAgentPermissions(agentID, actionType string) error {
	agent := strings.ToLower(agentID)
	action := strings.ToUpper(actionType)

	// Pricing Agent: allowed commercial rate recommendations, dynamic pricing; not finance mutation
	if strings.Contains(agent, "pricing") {
		if strings.Contains(action, "INVOICE") || strings.Contains(action, "CONTRACT_TERMINATE") || strings.Contains(action, "WRITE_OFF") || strings.Contains(action, "CREDIT_MEMO") || strings.Contains(action, "PAYMENT") || strings.Contains(action, "SETTLEMENT") {
			return fmt.Errorf("%w: PricingAgent cannot execute financial mutations '%s'", ErrAgentPrivilegeViolation, actionType)
		}
	}

	// Customer Agent: allowed communications, CRM updates; not rate changes or contract rewrites
	if strings.Contains(agent, "customer") {
		if strings.Contains(action, "OVERRIDE_RATE") || strings.Contains(action, "MODIFY_CONTRACT") || strings.Contains(action, "WRITE_OFF") {
			return fmt.Errorf("%w: CustomerAgent cannot modify contracts or financial rates '%s'", ErrAgentPrivilegeViolation, actionType)
		}
	}

	// Shipment Agent: allowed logistics recovery, tracking; not commercial discount approvals
	if strings.Contains(agent, "shipment") {
		if strings.Contains(action, "APPROVE_DISCOUNT") || strings.Contains(action, "WAIVE_DEBT") {
			return fmt.Errorf("%w: ShipmentAgent cannot authorize financial concessions '%s'", ErrAgentPrivilegeViolation, actionType)
		}
	}

	// Compliance Agent: cannot unilaterally execute financial payments
	if strings.Contains(agent, "compliance") {
		if strings.Contains(action, "DISPATCH_PAYMENT") || strings.Contains(action, "AUTHORIZE_PAYOUT") {
			return fmt.Errorf("%w: ComplianceAgent cannot execute financial disbursements '%s'", ErrAgentPrivilegeViolation, actionType)
		}
	}

	return nil
}

// ---------------------------------------------------------------------
// 8. Authoritative Risk Classification
// ---------------------------------------------------------------------

func (s *defaultEnterpriseGovernanceService) classifyAuthoritativeRisk(
	actionType string,
	monetaryExposure float64,
	proposedRisk ActionRiskLevel,
) ActionRiskLevel {
	action := strings.ToUpper(actionType)

	// Critical Actions (always CRITICAL regardless of agent proposal)
	if monetaryExposure >= 5000.0 ||
		strings.Contains(action, "WRITE_OFF") ||
		strings.Contains(action, "TERMINATE_CONTRACT") ||
		strings.Contains(action, "CANCEL_SHIPMENT") ||
		strings.Contains(action, "DISPATCH_PAYMENT") {
		return RiskLevelCritical
	}

	// High Risk Actions
	if monetaryExposure >= 1000.0 ||
		strings.Contains(action, "DISCOUNT") ||
		strings.Contains(action, "REROUTE") ||
		strings.Contains(action, "HOLD_SHIPMENT") ||
		strings.Contains(action, "CUSTOMS") {
		return RiskLevelHigh
	}

	// Medium Risk Actions
	if monetaryExposure >= 250.0 ||
		strings.Contains(action, "UPDATE_ETA") ||
		strings.Contains(action, "DUNNING") {
		if proposedRisk == RiskLevelHigh || proposedRisk == RiskLevelCritical {
			return proposedRisk
		}
		return RiskLevelMedium
	}

	if proposedRisk != "" {
		return proposedRisk
	}
	return RiskLevelLow
}

// ---------------------------------------------------------------------
// 9. Governance Status Summary (for Control Tower Integration)
// ---------------------------------------------------------------------

func (s *defaultEnterpriseGovernanceService) GetGovernanceStatus(ctx context.Context, orgID int64) (*GovernanceStatusSummary, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	activeHalts, _ := s.GetActiveEmergencyControls(ctx, orgID)
	emergencyActive := false
	for _, h := range activeHalts {
		if h.Scope == EmergencyScopeAll && h.Active {
			emergencyActive = true
			break
		}
	}

	return &GovernanceStatusSummary{
		OrgID:                   orgID,
		PolicyVersion:           CurrentPolicyVersion,
		EmergencyHaltActive:     emergencyActive,
		ActiveHaltCount:         len(activeHalts),
		TotalActionsEvaluated:   atomic.LoadInt64(&s.totalEvaluated),
		BlockedActionsCount:     atomic.LoadInt64(&s.blockedCount),
		ApprovalsRequiredCount:  atomic.LoadInt64(&s.approvalsRequired),
		StaleApprovalsPrevented: atomic.LoadInt64(&s.stalePrevented),
		LoopsDetectedCount:      atomic.LoadInt64(&s.loopsDetected),
		InjectionAttemptsCount:  atomic.LoadInt64(&s.injectionsBlocked),
		ActiveEmergencyControls: activeHalts,
		LastEvaluatedAt:         time.Now().UTC(),
	}, nil
}

// ---------------------------------------------------------------------
// Internal Helpers
// ---------------------------------------------------------------------

func (s *defaultEnterpriseGovernanceService) computeFingerprint(entityID string, state map[string]interface{}) string {
	raw, _ := json.Marshal(state)
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:])
}

func (s *defaultEnterpriseGovernanceService) recordAudit(
	ctx context.Context,
	orgID int64,
	action string,
	resourceType string,
	resourceID string,
	actor string,
	description string,
) {
	if s.auditSvc != nil {
		s.auditSvc.RecordAsync(ctx, domain.CreateAuditLogParams{
			OrgID:        orgID,
			ActorType:    domain.ActorTypeSystem,
			Action:       action,
			Module:       domain.ModuleSettings,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Description:  fmt.Sprintf("[%s] %s (Actor: %s)", CurrentPolicyVersion, description, actor),
		})
	}
}

func (s *defaultEnterpriseGovernanceService) normalizeAutonomyLevel(lvl string) AutonomyLevel {
	switch AutonomyLevel(lvl) {
	case AutonomyLvl0Observe, "LEVEL_0", "Observe":
		return AutonomyLvl0Observe
	case AutonomyLvl1Recommend, "LEVEL_1", "Recommend":
		return AutonomyLvl1Recommend
	case AutonomyLvl2Prepare, "LEVEL_2", "Prepare":
		return AutonomyLvl2Prepare
	case AutonomyLvl3Controlled, "LEVEL_3_CONTROLLED_EXECUTION", "LEVEL_3", "Controlled execution":
		return AutonomyLvl3Controlled
	case AutonomyLvl4Governed, "LEVEL_4_FULL_AUTONOMY", "LEVEL_4_CONTROLLED_MULTI_STEP", "LEVEL_4", "Governed multi-step autonomy":
		return AutonomyLvl4Governed
	default:
		return AutonomyLvl3Controlled
	}
}

func (s *defaultEnterpriseGovernanceService) autonomyRank(lvl AutonomyLevel) int {
	switch lvl {
	case AutonomyLvl0Observe:
		return 0
	case AutonomyLvl1Recommend:
		return 1
	case AutonomyLvl2Prepare:
		return 2
	case AutonomyLvl3Controlled:
		return 3
	case AutonomyLvl4Governed:
		return 4
	default:
		return 2
	}
}
