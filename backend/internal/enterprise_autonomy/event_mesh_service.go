package enterprise_autonomy

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/freel/backend/internal/actions"
	"github.com/freel/backend/internal/approvals"
	"github.com/freel/backend/internal/audit/domain"
	auditSvc "github.com/freel/backend/internal/audit/service"
	"github.com/freel/backend/internal/predictions"
	"github.com/freel/backend/internal/workforce"
	"github.com/jmoiron/sqlx"
)

// EnterpriseEventMeshService defines the unified event-to-workflow engine contract
type EnterpriseEventMeshService interface {
	IngestEvent(ctx context.Context, orgID int64, req IngestBusinessEventRequest) (*NormalizedBusinessEvent, *EnterpriseWorkflow, error)
	GetEvent(ctx context.Context, orgID int64, eventID string) (*NormalizedBusinessEvent, error)
	ListEvents(ctx context.Context, orgID int64, limit int, offset int) ([]*NormalizedBusinessEvent, int, error)
	GetDeadLetters(ctx context.Context, orgID int64, limit int) ([]*DeadLetterEvent, error)
	ReplayDeadLetter(ctx context.Context, orgID int64, eventID string) (*NormalizedBusinessEvent, *EnterpriseWorkflow, error)
	GetEventMeshOverview(ctx context.Context, orgID int64) (*EventMeshOverview, error)
	ListRoutingRules(ctx context.Context, orgID int64) ([]EventToWorkflowRule, error)
}

type defaultEnterpriseEventMeshService struct {
	repo                      Repository
	workforceSvc              workforce.Service
	actionsSvc                actions.Service
	approvalsSvc              approvals.Service
	auditSvc                  auditSvc.Service
	predictionsSvc            predictions.Service
	shipmentLifecycleSvc      ShipmentLifecycleService
	commercialLifecycleSvc    CommercialLifecycleService
	exceptionManagementSvc    ExceptionManagementService
	customerRelationshipSvc   CustomerRelationshipService
	revenueOptimizationSvc    RevenueOptimizationService
	contractComplianceRiskSvc ContractComplianceRiskService
	resilienceSvc             EnterpriseResilienceService
	db                        *sqlx.DB

	mu           sync.RWMutex
	eventsCache  map[string]*NormalizedBusinessEvent
	deadLetters  map[string]*DeadLetterEvent
	routingRules []EventToWorkflowRule
}

func NewEnterpriseEventMeshService(
	repo Repository,
	wfSvc workforce.Service,
	actSvc actions.Service,
	apprSvc approvals.Service,
	audSvc auditSvc.Service,
	predSvc predictions.Service,
	shipLifeSvc ShipmentLifecycleService,
	commLifeSvc CommercialLifecycleService,
	excMgmtSvc ExceptionManagementService,
	crmSvc CustomerRelationshipService,
	revOptSvc RevenueOptimizationService,
	riskSvc ContractComplianceRiskService,
	resilienceSvc EnterpriseResilienceService,
	db *sqlx.DB,
) EnterpriseEventMeshService {
	svc := &defaultEnterpriseEventMeshService{
		repo:                      repo,
		workforceSvc:              wfSvc,
		actionsSvc:                actSvc,
		approvalsSvc:              apprSvc,
		auditSvc:                  audSvc,
		predictionsSvc:            predSvc,
		shipmentLifecycleSvc:      shipLifeSvc,
		commercialLifecycleSvc:    commLifeSvc,
		exceptionManagementSvc:    excMgmtSvc,
		customerRelationshipSvc:   crmSvc,
		revenueOptimizationSvc:    revOptSvc,
		contractComplianceRiskSvc: riskSvc,
		resilienceSvc:             resilienceSvc,
		db:                        db,
		eventsCache:               make(map[string]*NormalizedBusinessEvent),
		deadLetters:               make(map[string]*DeadLetterEvent),
		routingRules:              defaultEventRoutingRules(),
	}
	return svc
}

// defaultEventRoutingRules establishes the deterministic mapping across all 7 autonomous lifecycle domains
func defaultEventRoutingRules() []EventToWorkflowRule {
	return []EventToWorkflowRule{
		// 1. Shipment Domain
		{
			EventType:          "SHIPMENT_DELAY_DETECTED",
			SourceModule:       "shipments",
			TargetWorkflowType: WorkflowShipmentRecovery,
			DefaultPriority:    PriorityCritical,
			RequiredCapability: "shipment.telematics.evaluate",
			MinAutonomyLevel:   "LEVEL_2_PREPARE",
			RequiresApproval:   false,
			CooldownSeconds:    60,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Triggers autonomous shipment recovery and replanning when vessel/truck schedule slips.",
		},
		{
			EventType:          "MILESTONE_UPDATED",
			SourceModule:       "shipments",
			TargetWorkflowType: WorkflowShipmentRecovery,
			DefaultPriority:    PriorityNormal,
			RequiredCapability: "shipment.milestone.audit",
			MinAutonomyLevel:   "LEVEL_3_CONTROLLED_EXECUTION",
			RequiresApproval:   false,
			CooldownSeconds:    30,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Audits milestone completion against SLA commitments and updates transit predictions.",
		},
		{
			EventType:          "SHIPMENT_DELIVERED",
			SourceModule:       "shipments",
			TargetWorkflowType: WorkflowShipmentRecovery,
			DefaultPriority:    PriorityNormal,
			RequiredCapability: "shipment.pod.verify",
			MinAutonomyLevel:   "LEVEL_3_CONTROLLED_EXECUTION",
			RequiresApproval:   false,
			CooldownSeconds:    60,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Audits proof of delivery and initiates post-delivery commercial/demurrage reconciliations.",
		},

		// 2. Exception Domain
		{
			EventType:          "EXCEPTION_RAISED",
			SourceModule:       "exceptions",
			TargetWorkflowType: WorkflowOperationalRecovery,
			DefaultPriority:    PriorityHigh,
			RequiredCapability: "exceptions.investigate",
			MinAutonomyLevel:   "LEVEL_2_PREPARE",
			RequiresApproval:   true,
			CooldownSeconds:    45,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Triggers multi-agent investigation, root cause detection, and recovery options formulation.",
		},
		{
			EventType:          "TEMPERATURE_EXCURSION",
			SourceModule:       "iot_telematics",
			TargetWorkflowType: WorkflowOperationalRecovery,
			DefaultPriority:    PriorityCritical,
			RequiredCapability: "exceptions.iot.triage",
			MinAutonomyLevel:   "LEVEL_2_PREPARE",
			RequiresApproval:   true,
			CooldownSeconds:    30,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Coordinates cold-chain excursion protocols, GDP compliance audits, and re-icing dispatches.",
		},

		// 3. Commercial & Quote-to-Cash Domain
		{
			EventType:          "RFQ_CREATED",
			SourceModule:       "rfqs",
			TargetWorkflowType: WorkflowCommercialCycle,
			DefaultPriority:    PriorityHigh,
			RequiredCapability: "commercial.rfq.extract",
			MinAutonomyLevel:   "LEVEL_2_PREPARE",
			RequiresApproval:   false,
			CooldownSeconds:    30,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Executes multi-agent extraction, spot pricing optimization, margin validation, and quote preparation.",
		},
		{
			EventType:          "QUOTATION_ACCEPTED",
			SourceModule:       "quotations",
			TargetWorkflowType: WorkflowCommercialCycle,
			DefaultPriority:    PriorityHigh,
			RequiredCapability: "commercial.booking.handoff",
			MinAutonomyLevel:   "LEVEL_3_CONTROLLED_EXECUTION",
			RequiresApproval:   false,
			CooldownSeconds:    60,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Triggers autonomous carrier booking handoff and prepares downstream shipment records.",
		},
		{
			EventType:          "CUSTOMER_COUNTER_OFFER",
			SourceModule:       "quotations",
			TargetWorkflowType: WorkflowCommercialCycle,
			DefaultPriority:    PriorityHigh,
			RequiredCapability: "commercial.pricing.negotiate",
			MinAutonomyLevel:   "LEVEL_2_PREPARE",
			RequiresApproval:   true,
			CooldownSeconds:    45,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Evaluates counter-offer viability against target margin hurdles and formulates concession strategy.",
		},

		// 4. Finance & Receivables Domain
		{
			EventType:          "INVOICE_OVERDUE",
			SourceModule:       "finance",
			TargetWorkflowType: WorkflowFinancialCollection,
			DefaultPriority:    PriorityHigh,
			RequiredCapability: "finance.receivables.triage",
			MinAutonomyLevel:   "LEVEL_2_PREPARE",
			RequiresApproval:   false,
			CooldownSeconds:    120,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Analyzes aging debt, evaluates customer relationship health, and generates personalized dunning interventions.",
		},
		{
			EventType:          "PAYMENT_DISPUTE_RAISED",
			SourceModule:       "finance",
			TargetWorkflowType: WorkflowFinancialCollection,
			DefaultPriority:    PriorityCritical,
			RequiredCapability: "finance.dispute.reconcile",
			MinAutonomyLevel:   "LEVEL_2_PREPARE",
			RequiresApproval:   true,
			CooldownSeconds:    60,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Reconciles billed accessorials against carrier telematics and rate agreement terms.",
		},

		// 5. Customer Relationship Management Domain
		{
			EventType:          "CUSTOMER_CHURN_SIGNAL",
			SourceModule:       "crm",
			TargetWorkflowType: WorkflowCustomerRelationship,
			DefaultPriority:    PriorityHigh,
			RequiredCapability: "customer.health.audit",
			MinAutonomyLevel:   "LEVEL_2_PREPARE",
			RequiresApproval:   true,
			CooldownSeconds:    90,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Synthesizes volume drop-off and support sentiment to prepare proactive commercial retention plans.",
		},
		{
			EventType:          "CUSTOMER_COMPLAINT_LOGGED",
			SourceModule:       "crm",
			TargetWorkflowType: WorkflowCustomerRelationship,
			DefaultPriority:    PriorityHigh,
			RequiredCapability: "customer.complaint.remediate",
			MinAutonomyLevel:   "LEVEL_2_PREPARE",
			RequiresApproval:   true,
			CooldownSeconds:    60,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Investigates root cause across operational records and structures remedial outreach.",
		},

		// 6. Revenue & Margin Optimization Domain
		{
			EventType:          "MARGIN_EROSION_DETECTED",
			SourceModule:       "pricing",
			TargetWorkflowType: WorkflowRevenueOptimization,
			DefaultPriority:    PriorityHigh,
			RequiredCapability: "revenue.margin.protect",
			MinAutonomyLevel:   "LEVEL_2_PREPARE",
			RequiresApproval:   true,
			CooldownSeconds:    60,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Detects carrier surcharge drift and initiates lane repricing and carrier re-allocation.",
		},

		// 7. Contract, Compliance & Risk Governance Domain
		{
			EventType:          "CONTRACT_EXPIRATION_APPROACHING",
			SourceModule:       "contracts",
			TargetWorkflowType: WorkflowContractComplianceRisk,
			DefaultPriority:    PriorityHigh,
			RequiredCapability: "contracts.renewal.audit",
			MinAutonomyLevel:   "LEVEL_2_PREPARE",
			RequiresApproval:   true,
			CooldownSeconds:    120,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Assesses expiring rate cards, active shipment exposure, and initiates renewal negotiations.",
		},
		{
			EventType:          "CUSTOMS_REGULATORY_HOLD",
			SourceModule:       "compliance",
			TargetWorkflowType: WorkflowContractComplianceRisk,
			DefaultPriority:    PriorityCritical,
			RequiredCapability: "compliance.customs.escalate",
			MinAutonomyLevel:   "LEVEL_2_PREPARE",
			RequiresApproval:   true,
			CooldownSeconds:    30,
			MaxRetries:         3,
			IsEnabled:          true,
			Description:        "Triggers statutory compliance checks, assesses cascading delivery delay penalties, and expedites documentation.",
		},
	}
}

// ---------------------------------------------------------------------
// IngestEvent: Complete Ingestion Pipeline
// ---------------------------------------------------------------------

func (s *defaultEnterpriseEventMeshService) IngestEvent(ctx context.Context, orgID int64, req IngestBusinessEventRequest) (*NormalizedBusinessEvent, *EnterpriseWorkflow, error) {
	if orgID <= 0 {
		return nil, nil, ErrUnauthorizedTenant
	}

	// 1. Security & Prompt Injection Defense
	if containsSuspiciousPromptInjection(req.EventType) || containsSuspiciousPromptInjection(req.SourceModule) {
		s.recordDeadLetter(orgID, req, "Prompt injection pattern detected in event metadata")
		return nil, nil, ErrPromptInjectionDetected
	}
	for k, v := range req.Payload {
		if str, ok := v.(string); ok && containsSuspiciousPromptInjection(str) {
			s.recordDeadLetter(orgID, req, fmt.Sprintf("Prompt injection detected in payload key '%s'", k))
			return nil, nil, ErrPromptInjectionDetected
		}
	}

	// 2. Normalize and Validate Schema
	occAt := time.Now().UTC()
	if req.OccurredAt != nil {
		occAt = req.OccurredAt.UTC()
	}

	corrID := req.CorrelationID
	if corrID == "" {
		corrID = fmt.Sprintf("evt-corr-%d-%d", orgID, time.Now().UnixNano())
	}
	causID := req.CausationID
	if causID == "" {
		causID = corrID
	}
	eventID := req.EventID

	evt := &NormalizedBusinessEvent{
		EventID:          eventID,
		OrgID:            orgID,
		EventType:        strings.ToUpper(strings.TrimSpace(req.EventType)),
		SourceModule:     strings.ToLower(strings.TrimSpace(req.SourceModule)),
		EntityType:       strings.ToUpper(strings.TrimSpace(req.EntityType)),
		EntityID:         strings.TrimSpace(req.EntityID),
		ParentEntityID:   req.ParentEntityID,
		EventVersion:     req.EventVersion,
		Priority:         EventPriority(req.Priority),
		Payload:          req.Payload,
		CorrelationID:    corrID,
		CausationID:      causID,
		ActionID:         req.ActionID,
		OccurredAt:       occAt,
		ReceivedAt:       time.Now().UTC(),
		ProcessingStatus: EventStatusReceived,
		RetryCount:       0,
		MaxRetries:       3,
		DedupKey:         fmt.Sprintf("mesh-dedup:%d:%s:%s:%s", orgID, req.EventType, req.EntityType, req.EntityID),
	}

	if err := ValidateNormalizedEvent(evt); err != nil {
		s.recordDeadLetter(orgID, req, err.Error())
		return nil, nil, err
	}

	// 3. Stale Event / Out-of-Order Check
	// If event occurred over 7 days ago and has no explicit historical reprocessing flag, reject as stale
	if time.Since(evt.OccurredAt) > 7*24*time.Hour {
		evt.ProcessingStatus = EventStatusStaleIgnored
		reason := "Event occurred outside active processing window (>7 days old)"
		evt.FailureReason = &reason
		s.cacheEvent(evt)
		return evt, nil, ErrStaleEventIgnored
	}

	// 4. Loop & Feedback Cycle Suppression
	// If the event was caused directly by an action executed by this platform and has no material change
	if evt.ActionID != nil && strings.HasPrefix(*evt.ActionID, "risk-exec-") || (evt.CausationID != "" && strings.HasPrefix(evt.CausationID, "wf-")) {
		if hasMaterial, ok := req.Payload["material_change"].(bool); !ok || !hasMaterial {
			evt.ProcessingStatus = EventStatusLoopSuppressed
			reason := "Suppressed potential feedback loop: event originated from completed platform action without material change"
			evt.FailureReason = &reason
			s.cacheEvent(evt)
			return evt, nil, nil
		}
	}

	// 4.5. Event Storm & Rapid Ingestion Protection (Phase 7.11)
	if s.resilienceSvc != nil {
		allowed, reason := s.resilienceSvc.CheckEventStormProtection(ctx, orgID, evt.EntityType, evt.EntityID, evt.EventType)
		if !allowed {
			evt.ProcessingStatus = EventStatusLoopSuppressed
			evt.FailureReason = &reason
			s.cacheEvent(evt)
			return evt, nil, nil
		}
	}

	// 5. Deduplication & Replay Guard
	isNew, err := s.repo.CheckAndRecordEventDedup(ctx, orgID, evt.DedupKey, evt.EventType)
	if err == nil && !isNew {
		evt.ProcessingStatus = EventStatusDeduplicated
		s.cacheEvent(evt)
		// Return existing cached workflow if one exists for this entity
		existingWF := s.findActiveWorkflowForEntity(orgID, evt.EntityType, evt.EntityID)
		return evt, existingWF, nil
	}

	// 6. Deterministic Route Selection
	rule := s.findRoutingRule(evt.EventType, evt.SourceModule)
	if rule == nil || !rule.IsEnabled {
		evt.ProcessingStatus = EventStatusRouted
		s.cacheEvent(evt)
		return evt, nil, nil // Deterministically no workflow configured
	}

	evt.WorkflowType = &rule.TargetWorkflowType
	evt.Priority = rule.DefaultPriority

	// 7. Controlled Workflow Creation / Activation
	wf, err := s.dispatchWorkflow(ctx, orgID, evt, rule)
	if err != nil {
		reason := err.Error()
		evt.FailureReason = &reason
		evt.ProcessingStatus = EventStatusDeadLetter
		s.recordDeadLetter(orgID, req, reason)
		s.cacheEvent(evt)
		return evt, nil, err
	}

	if wf != nil {
		evt.WorkflowID = &wf.WorkflowID
		evt.ProcessingStatus = EventStatusWorkflowActive
	} else {
		evt.ProcessingStatus = EventStatusCompleted
	}

	s.cacheEvent(evt)

	// 8. Structured Audit Recording
	if s.auditSvc != nil {
		s.auditSvc.RecordAsync(ctx, domain.CreateAuditLogParams{
			OrgID:        orgID,
			ActorType:    domain.ActorTypeSystem,
			Action:       domain.ActionCreate,
			Module:       domain.ModuleSettings,
			ResourceType: "enterprise_event_mesh",
			ResourceID:   evt.EventID,
			Description:  fmt.Sprintf("Event Mesh routed %s on %s:%s to %s (Workflow: %v)", evt.EventType, evt.EntityType, evt.EntityID, rule.TargetWorkflowType, evt.WorkflowID),
		})
	}

	return evt, wf, nil
}

// ---------------------------------------------------------------------
// Dispatching to Authorized Enterprise Workflows
// ---------------------------------------------------------------------

func (s *defaultEnterpriseEventMeshService) dispatchWorkflow(
	ctx context.Context,
	orgID int64,
	evt *NormalizedBusinessEvent,
	rule *EventToWorkflowRule,
) (*EnterpriseWorkflow, error) {
	// Enforce Policy Checks
	policy, err := s.repo.GetPolicyContext(ctx, orgID, string(rule.TargetWorkflowType))
	if err == nil && policy != nil && policy.EmergencyStopActive {
		return nil, ErrEmergencyStopActive
	}

	// Check for active workflow to prevent duplication of ongoing execution
	existingWF := s.findActiveWorkflowForEntity(orgID, evt.EntityType, evt.EntityID)
	if existingWF != nil && existingWF.CurrentState != StateCompleted && existingWF.CurrentState != StateCancelled {
		return existingWF, nil
	}

	// Construct and launch target EnterpriseWorkflow
	objective := fmt.Sprintf("Event-driven autonomous workflow for %s on %s #%s (%s)", evt.EventType, evt.EntityType, evt.EntityID, rule.Description)
	startReq := StartWorkflowRequest{
		WorkflowType:      rule.TargetWorkflowType,
		Objective:         objective,
		InitiatingEvent:   evt.EventType,
		RelatedEntityType: evt.EntityType,
		RelatedEntityID:   evt.EntityID,
		AutonomyLevel:     rule.MinAutonomyLevel,
		InitialContext:    evt.Payload,
		CorrelationID:     evt.CorrelationID,
		IdempotencyKey:    evt.DedupKey,
	}

	// Use root enterprise service via default workflow creator
	wf, err := s.createAndPersistWorkflow(ctx, orgID, startReq)
	if err != nil {
		return nil, err
	}

	return wf, nil
}

func (s *defaultEnterpriseEventMeshService) createAndPersistWorkflow(ctx context.Context, orgID int64, req StartWorkflowRequest) (*EnterpriseWorkflow, error) {
	wfID := fmt.Sprintf("wf-%s-%d", strings.ToLower(string(req.WorkflowType)), time.Now().UnixNano())
	now := time.Now().UTC()

	wf := &EnterpriseWorkflow{
		WorkflowID:        wfID,
		OrgID:             orgID,
		WorkflowType:      req.WorkflowType,
		Objective:         req.Objective,
		InitiatingEvent:   req.InitiatingEvent,
		CurrentState:      StateRunning,
		AutonomyLevel:     req.AutonomyLevel,
		PolicyDecision:    "PERMITTED",
		RelatedEntityType: req.RelatedEntityType,
		RelatedEntityID:   req.RelatedEntityID,
		Confidence:        0.95,
		CorrelationID:     req.CorrelationID,
		IdempotencyKey:    req.IdempotencyKey,
		Steps:             []EnterpriseWorkflowStep{},
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err := s.repo.CreateWorkflow(ctx, wf)
	if err != nil {
		return nil, err
	}

	// Append initial stage step
	step := EnterpriseWorkflowStep{
		StepID:           fmt.Sprintf("step-event-ingest-%d", now.Unix()),
		WorkflowID:       wfID,
		OrgID:            orgID,
		StepNumber:       1,
		AgentID:          "orchestration_agent",
		ActionType:       "EVENT_MESH_TRIGGER",
		Title:            fmt.Sprintf("Ingested Event: %s", req.InitiatingEvent),
		Description:      req.Objective,
		ExpectedOutcome:  "Workflow initiated from event mesh",
		RiskLevel:        "LOW",
		RequiresApproval: false,
		Status:           "COMPLETED",
		ExecutionAttempt: 1,
		MaxAttempts:      3,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	_ = s.repo.SaveWorkflowSteps(ctx, orgID, wfID, []EnterpriseWorkflowStep{step})
	wf.Steps = append(wf.Steps, step)

	return wf, nil
}

// ---------------------------------------------------------------------
// Helper Functions & Lookups
// ---------------------------------------------------------------------

func (s *defaultEnterpriseEventMeshService) findRoutingRule(eventType, sourceModule string) *EventToWorkflowRule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	et := strings.ToUpper(strings.TrimSpace(eventType))
	for _, rule := range s.routingRules {
		if rule.EventType == et {
			return &rule
		}
	}
	// Fallback loose match
	for _, rule := range s.routingRules {
		if strings.Contains(et, strings.ToUpper(rule.SourceModule)) {
			return &rule
		}
	}
	return nil
}

func (s *defaultEnterpriseEventMeshService) findActiveWorkflowForEntity(orgID int64, entityType, entityID string) *EnterpriseWorkflow {
	// Query repository for active workflow
	wfs, _, err := s.repo.ListWorkflows(context.Background(), WorkflowFilter{
		OrgID: orgID,
		Limit: 10,
	})
	if err == nil {
		for _, w := range wfs {
			if w.RelatedEntityType == entityType && w.RelatedEntityID == entityID && (w.CurrentState == StateRunning || w.CurrentState == StatePending || w.CurrentState == StateWaitingForApproval) {
				return w
			}
		}
	}
	return nil
}

func (s *defaultEnterpriseEventMeshService) cacheEvent(evt *NormalizedBusinessEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := fmt.Sprintf("%d:%s", evt.OrgID, evt.EventID)
	s.eventsCache[key] = evt
}

func (s *defaultEnterpriseEventMeshService) recordDeadLetter(orgID int64, req IngestBusinessEventRequest, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dl := &DeadLetterEvent{
		EventID:        req.EventID,
		OrgID:          orgID,
		EventType:      req.EventType,
		SourceModule:   req.SourceModule,
		EntityType:     req.EntityType,
		EntityID:       req.EntityID,
		CorrelationID:  req.CorrelationID,
		CausationID:    req.CausationID,
		Payload:        req.Payload,
		FailureReason:  reason,
		AttemptCount:   1,
		DeadLetteredAt: time.Now().UTC(),
		Resolved:       false,
	}
	key := fmt.Sprintf("%d:%s", orgID, req.EventID)
	s.deadLetters[key] = dl
}

// ---------------------------------------------------------------------
// Inspection, Overview & Dead-Letter Replay
// ---------------------------------------------------------------------

func (s *defaultEnterpriseEventMeshService) GetEvent(ctx context.Context, orgID int64, eventID string) (*NormalizedBusinessEvent, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	key := fmt.Sprintf("%d:%s", orgID, eventID)
	if evt, ok := s.eventsCache[key]; ok {
		return evt, nil
	}
	return nil, fmt.Errorf("event %s not found for organization %d", eventID, orgID)
}

func (s *defaultEnterpriseEventMeshService) ListEvents(ctx context.Context, orgID int64, limit int, offset int) ([]*NormalizedBusinessEvent, int, error) {
	if orgID <= 0 {
		return nil, 0, ErrUnauthorizedTenant
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []*NormalizedBusinessEvent
	for _, evt := range s.eventsCache {
		if evt.OrgID == orgID {
			list = append(list, evt)
		}
	}
	total := len(list)
	if offset > total {
		return []*NormalizedBusinessEvent{}, total, nil
	}
	end := offset + limit
	if end > total || limit <= 0 {
		end = total
	}
	return list[offset:end], total, nil
}

func (s *defaultEnterpriseEventMeshService) GetDeadLetters(ctx context.Context, orgID int64, limit int) ([]*DeadLetterEvent, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []*DeadLetterEvent
	for _, dl := range s.deadLetters {
		if dl.OrgID == orgID {
			list = append(list, dl)
		}
	}
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (s *defaultEnterpriseEventMeshService) ReplayDeadLetter(ctx context.Context, orgID int64, eventID string) (*NormalizedBusinessEvent, *EnterpriseWorkflow, error) {
	if orgID <= 0 {
		return nil, nil, ErrUnauthorizedTenant
	}

	s.mu.Lock()
	key := fmt.Sprintf("%d:%s", orgID, eventID)
	dl, ok := s.deadLetters[key]
	if !ok {
		s.mu.Unlock()
		return nil, nil, fmt.Errorf("dead-letter event %s not found", eventID)
	}
	delete(s.deadLetters, key)
	s.mu.Unlock()

	// Re-ingest with clean retry parameters
	return s.IngestEvent(ctx, orgID, IngestBusinessEventRequest{
		EventID:       dl.EventID,
		EventType:     dl.EventType,
		SourceModule:  dl.SourceModule,
		EntityType:    dl.EntityType,
		EntityID:      dl.EntityID,
		Payload:       dl.Payload,
		CorrelationID: dl.CorrelationID,
		CausationID:   dl.CausationID,
	})
}

func (s *defaultEnterpriseEventMeshService) ListRoutingRules(ctx context.Context, orgID int64) ([]EventToWorkflowRule, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.routingRules, nil
}

func (s *defaultEnterpriseEventMeshService) GetEventMeshOverview(ctx context.Context, orgID int64) (*EventMeshOverview, error) {
	if orgID <= 0 {
		return nil, ErrUnauthorizedTenant
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	overview := &EventMeshOverview{
		OrgID:              orgID,
		RecentEvents:       make([]*NormalizedBusinessEvent, 0),
		RecentDeadLetters:  make([]*DeadLetterEvent, 0),
		ActiveRoutingRules: s.routingRules,
	}

	for _, evt := range s.eventsCache {
		if evt.OrgID == orgID {
			overview.TotalEventsReceived++
			if evt.ProcessingStatus == EventStatusDeduplicated {
				overview.DeduplicatedCount++
			} else if evt.ProcessingStatus == EventStatusLoopSuppressed {
				overview.LoopsSuppressedCount++
			} else if evt.ProcessingStatus == EventStatusWorkflowActive {
				overview.WorkflowsTriggered++
				overview.EventsRoutedCount++
			} else if evt.ProcessingStatus == EventStatusRouted {
				overview.EventsRoutedCount++
			}
			overview.RecentEvents = append(overview.RecentEvents, evt)
		}
	}

	for _, dl := range s.deadLetters {
		if dl.OrgID == orgID {
			overview.DeadLetterCount++
			overview.RecentDeadLetters = append(overview.RecentDeadLetters, dl)
		}
	}

	overview.ActiveWorkflowsCount = overview.WorkflowsTriggered
	return overview, nil
}
