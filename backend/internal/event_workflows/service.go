package event_workflows

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/freel/backend/internal/approvals"
	auditSvc "github.com/freel/backend/internal/audit/service"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/orchestration"
	"github.com/jmoiron/sqlx"
)

const (
	DefaultSidecarTimeout = 12 * time.Second
	DefaultServiceKey     = "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456"
)

type Service interface {
	IngestAndProcessEvent(ctx context.Context, orgID int64, userID int64, input IngestEventInput) (*WorkflowInstance, error)
	GetOverview(ctx context.Context, orgID int64) (*EventWorkflowsOverview, error)
	ListEvents(ctx context.Context, orgID int64, limit, offset int) ([]*EventRecord, int, error)
	GetEventByID(ctx context.Context, orgID, id int64) (*EventRecord, error)
	ListWorkflowInstances(ctx context.Context, orgID int64, limit, offset int) ([]*WorkflowInstance, int, error)
	GetWorkflowInstanceByID(ctx context.Context, orgID, id int64) (*WorkflowInstance, error)
	RetryWorkflow(ctx context.Context, orgID, id int64, userID int64) (*WorkflowInstance, error)
	CancelWorkflow(ctx context.Context, orgID, id int64, reason string) (*WorkflowInstance, error)
}

type service struct {
	db             *sqlx.DB
	repo           Repository
	approvalsSvc   approvals.Service
	orchestrationR orchestration.Registry
	auditSvc       auditSvc.Service
	sidecarURL     string
	serviceKey     string
	httpClient     *http.Client
}

func NewService(
	db *sqlx.DB,
	repo Repository,
	approvalsSvc approvals.Service,
	orchestrationR orchestration.Registry,
	auditSvc auditSvc.Service,
) Service {
	sidecar := os.Getenv("AI_SIDECAR_URL")
	if sidecar == "" {
		sidecar = "http://localhost:8090"
	}
	key := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if key == "" {
		key = os.Getenv("INTERNAL_SERVICE_KEY")
	}
	if key == "" {
		key = os.Getenv("AI_SIDECAR_SERVICE_KEY")
	}
	if key == "" {
		rawEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
		if rawEnv == "" {
			rawEnv = strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
		}
		if rawEnv == "" {
			rawEnv = strings.ToLower(strings.TrimSpace(os.Getenv("ENVIRONMENT")))
		}
		if rawEnv == "development" || rawEnv == "test" {
			key = DefaultServiceKey
		}
	}

	return &service{
		db:             db,
		repo:           repo,
		approvalsSvc:   approvalsSvc,
		orchestrationR: orchestrationR,
		auditSvc:       auditSvc,
		sidecarURL:     sidecar,
		serviceKey:     key,
		httpClient:     &http.Client{Timeout: DefaultSidecarTimeout},
	}
}

func (s *service) IngestAndProcessEvent(ctx context.Context, orgID int64, userID int64, input IngestEventInput) (*WorkflowInstance, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization id")
	}
	if strings.TrimSpace(input.EventType) == "" {
		return nil, fmt.Errorf("event_type is required")
	}
	if strings.TrimSpace(input.SourceRecordType) == "" || strings.TrimSpace(input.SourceRecordID) == "" {
		return nil, fmt.Errorf("source_record_type and source_record_id are required")
	}

	corrID := input.CorrelationID
	if corrID == "" {
		corrID = fmt.Sprintf("evt-corr-%d-%d", orgID, time.Now().UnixNano())
	}
	eventVersion := input.EventVersion
	if eventVersion == "" {
		eventVersion = "1.0"
	}
	actorType := input.ActorType
	if actorType == "" {
		actorType = "SYSTEM"
	}
	eventTimestamp := time.Now()
	if input.Timestamp != nil {
		eventTimestamp = *input.Timestamp
	}

	// 1. Stable Deduplication Key for Event
	eventDedupKey := fmt.Sprintf("evt:%d:%s:%s:%s:%s", orgID, input.EventType, input.SourceRecordType, input.SourceRecordID, corrID)

	existingEvent, err := s.repo.GetEventByDedupKey(ctx, orgID, eventDedupKey)
	if err == nil && existingEvent != nil {
		// If event already recorded and has a workflow, retrieve existing workflow
		existingWf, _ := s.repo.GetWorkflowInstanceByDedupKey(ctx, orgID, eventDedupKey)
		if existingWf != nil {
			return existingWf, nil
		}
	}

	payloadBytes, _ := json.Marshal(input.Payload)

	eventRecord := &EventRecord{
		OrgID:            orgID,
		EventType:        input.EventType,
		SourceModule:     input.SourceModule,
		SourceRecordType: input.SourceRecordType,
		SourceRecordID:   input.SourceRecordID,
		EventVersion:     eventVersion,
		ActorType:        actorType,
		ActorID:          input.ActorID,
		Payload:          RawJSON(payloadBytes),
		DedupKey:         eventDedupKey,
		Status:           "RECEIVED",
		RetryCount:       0,
		CorrelationID:    corrID,
		CausationID:      input.CausationID,
		EventTimestamp:   eventTimestamp,
	}

	savedEvent, err := s.repo.SaveEvent(ctx, eventRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to persist event: %w", err)
	}

	// 2. Deterministic Eligibility Check in Go
	isEligible, workflowType := s.evaluateEligibility(input.EventType, input.SourceRecordType)
	if !isEligible {
		_ = s.repo.UpdateEventStatus(ctx, orgID, savedEvent.ID, "IGNORED", nil)
		return nil, nil
	}

	// 3. Aggregate Authorized Cross-Module Context
	crossCtx := s.gatherCrossModuleContext(ctx, orgID, input.SourceRecordType, input.SourceRecordID)

	// 4. Create Initial Workflow Instance in MariaDB
	wfInstance := &WorkflowInstance{
		OrgID:            orgID,
		EventID:          savedEvent.ID,
		WorkflowType:     workflowType,
		Status:           "RUNNING",
		TriggerEventType: input.EventType,
		SourceModule:     input.SourceModule,
		SourceRecordType: input.SourceRecordType,
		SourceRecordID:   input.SourceRecordID,
		Urgency:          "MEDIUM",
		AISummary:        fmt.Sprintf("Evaluating cross-module workflow for event %s on %s #%s", input.EventType, input.SourceRecordType, input.SourceRecordID),
		AIAnalysis:       RawJSON("{}"),
		Recommendations:  RawJSON("[]"),
		ActionStatus:     "NOT_STARTED",
		DedupKey:         eventDedupKey,
		CorrelationID:    corrID,
		StartedAt:        time.Now(),
	}

	savedWf, err := s.repo.SaveWorkflowInstance(ctx, wfInstance)
	if err != nil {
		_ = s.repo.UpdateEventStatus(ctx, orgID, savedEvent.ID, "FAILED", stringPtr(err.Error()))
		return nil, fmt.Errorf("failed to initialize workflow instance: %w", err)
	}

	// 5. Call Python AI Sidecar
	aiResp, aiErr := s.callSidecarAnalyzeEvent(ctx, savedEvent.ID, orgID, input.EventType, input.SourceModule, input.SourceRecordType, input.SourceRecordID, input.Payload, crossCtx, corrID)
	if aiErr != nil {
		errMsg := aiErr.Error()
		_ = s.repo.UpdateWorkflowStatus(ctx, orgID, savedWf.ID, "FAILED", "FAILED", nil, &errMsg)
		_ = s.repo.UpdateEventStatus(ctx, orgID, savedEvent.ID, "FAILED", &errMsg)
		return savedWf, fmt.Errorf("AI event analysis failed: %w", aiErr)
	}

	// 6. Process AI Response & Recommendations
	analysisBytes, _ := json.Marshal(aiResp)
	recsBytes, _ := json.Marshal(aiResp.RecommendedActions)

	savedWf.AISummary = aiResp.Summary
	savedWf.Urgency = aiResp.Urgency
	savedWf.AIAnalysis = RawJSON(analysisBytes)
	savedWf.Recommendations = RawJSON(recsBytes)

	// Check if any recommended action requires human approval
	var highestRiskRec *WorkflowRecommendation
	for _, rec := range aiResp.RecommendedActions {
		if rec.RequiresApproval || rec.RiskLevel == "HIGH" || rec.RiskLevel == "CRITICAL" {
			highestRiskRec = &rec
			break
		}
	}

	if highestRiskRec != nil && s.approvalsSvc != nil {
		actionIntentBytes, _ := json.Marshal(highestRiskRec)
		savedWf.ActionIntent = RawJSON(actionIntentBytes)
		savedWf.ActionName = &highestRiskRec.ActionName
		savedWf.Status = "AWAITING_APPROVAL"
		savedWf.ActionStatus = "AWAITING_APPROVAL"

		entityID, _ := strconv.ParseInt(input.SourceRecordID, 10, 64)
		custName := "Commercial Counterparty"
		if cn, ok := crossCtx["customer_name"].(string); ok && cn != "" {
			custName = cn
		} else if pn, ok := crossCtx["party_name"].(string); ok && pn != "" {
			custName = pn
		}

		payloadJsonBytes, _ := json.Marshal(highestRiskRec.ActionPayload)

		appReq, appErr := s.approvalsSvc.CreateApproval(ctx, orgID, &approvals.CreateApprovalInput{
			Title:              fmt.Sprintf("Cross-Module Event Approval: %s (%s)", highestRiskRec.Title, input.EventType),
			Category:           "CROSS_MODULE_WORKFLOW",
			Type:               "OPERATIONAL_ACTION",
			Priority:           highestRiskRec.RiskLevel,
			RelatedEntityType:  input.SourceRecordType,
			RelatedEntityID:    entityID,
			RelatedRef:         fmt.Sprintf("%s-%s", input.SourceRecordType, input.SourceRecordID),
			CustomerName:       custName,
			RequestedByID:     userID,
			Description:        fmt.Sprintf("Event %s triggered high-risk action %s: %s", input.EventType, highestRiskRec.ActionName, highestRiskRec.Description),
			ActionName:         highestRiskRec.ActionName,
			RiskLevel:          highestRiskRec.RiskLevel,
			RequiredPermission: "operations:write",
			ProposedPayload:    string(payloadJsonBytes),
			CorrelationID:      corrID,
			ActorType:          "AI",
		}, "Event Workflow Engine")

		if appErr == nil && appReq != nil {
			savedWf.ApprovalID = &appReq.ID
		}
	} else if len(aiResp.RecommendedActions) > 0 {
		// Safe internal action (e.g. notifications.create)
		firstRec := aiResp.RecommendedActions[0]
		actionIntentBytes, _ := json.Marshal(firstRec)
		savedWf.ActionIntent = actionIntentBytes
		savedWf.ActionName = &firstRec.ActionName
		savedWf.ActionStatus = "EXECUTED"
		savedWf.Status = "COMPLETED"
	} else {
		savedWf.Status = "COMPLETED"
		savedWf.ActionStatus = "NOT_STARTED"
	}

	_ = s.repo.UpdateWorkflowStatus(ctx, orgID, savedWf.ID, savedWf.Status, savedWf.ActionStatus, savedWf.ApprovalID, nil)
	_ = s.repo.UpdateEventStatus(ctx, orgID, savedEvent.ID, "PROCESSED", nil)

	// Audit Logging
	if s.auditSvc != nil {
		s.auditSvc.RecordAsync(ctx, domain.CreateAuditLogParams{
			OrgID:        orgID,
			ActorType:    domain.ActorTypeSystem,
			Action:       domain.ActionCreate,
			Module:       domain.ModuleSettings,
			ResourceType: "ai_cross_module_workflows",
			ResourceID:   fmt.Sprintf("%d", savedWf.ID),
			Description:  fmt.Sprintf("Event %s processed workflow #%d (Urgency: %s, Status: %s)", input.EventType, savedWf.ID, savedWf.Urgency, savedWf.Status),
		})
	}

	return savedWf, nil
}

func (s *service) evaluateEligibility(eventType, sourceRecordType string) (bool, string) {
	et := strings.ToLower(eventType)
	if strings.Contains(et, "shipment") || strings.Contains(et, "milestone") || strings.Contains(et, "exception") {
		return true, "SHIPMENT_EXCEPTION_RESPONSE"
	}
	if strings.Contains(et, "invoice") || strings.Contains(et, "payment") || strings.Contains(et, "receivable") {
		return true, "INVOICE_COLLECTION_ESCALATION"
	}
	if strings.Contains(et, "contract") || strings.Contains(et, "compliance") || strings.Contains(et, "document") {
		return true, "CONTRACT_COMPLIANCE_RENEWAL"
	}
	if strings.Contains(et, "rfq") || strings.Contains(et, "quote") || strings.Contains(et, "pricing") {
		return true, "RFQ_QUOTATION_DISPATCH"
	}
	if strings.Contains(et, "lead") {
		return true, "LEAD_FOLLOWUP"
	}
	return false, ""
}

func (s *service) gatherCrossModuleContext(ctx context.Context, orgID int64, recordType, recordID string) map[string]interface{} {
	id, err := strconv.ParseInt(recordID, 10, 64)
	if err != nil || id <= 0 {
		return map[string]interface{}{}
	}

	switch strings.ToUpper(recordType) {
	case "SHIPMENT":
		sCtx, _ := s.repo.GetShipmentContext(ctx, orgID, id)
		return sCtx
	case "INVOICE", "CUSTOMER_INVOICE":
		iCtx, _ := s.repo.GetInvoiceContext(ctx, orgID, id)
		return iCtx
	case "CONTRACT":
		cCtx, _ := s.repo.GetContractContext(ctx, orgID, id)
		return cCtx
	case "CUSTOMER":
		custCtx, _ := s.repo.GetCustomerContext(ctx, orgID, id)
		return custCtx
	default:
		return map[string]interface{}{}
	}
}

func (s *service) callSidecarAnalyzeEvent(
	ctx context.Context,
	eventID, orgID int64,
	eventType, sourceModule, sourceRecordType, sourceRecordID string,
	facts, crossCtx map[string]interface{},
	correlationID string,
) (*EventAnalysisResponse, error) {
	url := fmt.Sprintf("%s/event-workflows/analyze-event", s.sidecarURL)

	payload := map[string]interface{}{
		"event_id":             eventID,
		"org_id":               orgID,
		"event_type":           eventType,
		"source_module":        sourceModule,
		"source_record_type":   sourceRecordType,
		"source_record_id":     sourceRecordID,
		"event_facts":          facts,
		"cross_module_context": crossCtx,
		"correlation_id":       correlationID,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sidecar payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to build sidecar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LogisticsHQ-Service-Key", s.serviceKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sidecar call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar returned non-200 status code: %d", resp.StatusCode)
	}

	var aiResp EventAnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return nil, fmt.Errorf("failed to parse sidecar response: %w", err)
	}
	return &aiResp, nil
}

func (s *service) GetOverview(ctx context.Context, orgID int64) (*EventWorkflowsOverview, error) {
	events, totalEvents, err := s.repo.ListEvents(ctx, orgID, 15, 0)
	if err != nil {
		return nil, err
	}
	workflows, _, err := s.repo.ListWorkflowInstances(ctx, orgID, 15, 0)
	if err != nil {
		return nil, err
	}

	var activeCount, pendingApprovals int
	for _, wf := range workflows {
		if wf.Status == "RUNNING" || wf.Status == "AWAITING_APPROVAL" {
			activeCount++
		}
		if wf.Status == "AWAITING_APPROVAL" {
			pendingApprovals++
		}
	}

	supported := []string{
		"SHIPMENT_EXCEPTION_RESPONSE",
		"INVOICE_COLLECTION_ESCALATION",
		"CONTRACT_COMPLIANCE_RENEWAL",
		"RFQ_QUOTATION_DISPATCH",
		"LEAD_FOLLOWUP",
	}

	return &EventWorkflowsOverview{
		TotalEventsIngested:    totalEvents,
		ActiveWorkflowsCount:   activeCount,
		PendingApprovalsCount:  pendingApprovals,
		RecentEvents:           events,
		RecentWorkflows:        workflows,
		SupportedWorkflowsList: supported,
	}, nil
}

func (s *service) ListEvents(ctx context.Context, orgID int64, limit, offset int) ([]*EventRecord, int, error) {
	return s.repo.ListEvents(ctx, orgID, limit, offset)
}

func (s *service) GetEventByID(ctx context.Context, orgID, id int64) (*EventRecord, error) {
	return s.repo.GetEventByID(ctx, orgID, id)
}

func (s *service) ListWorkflowInstances(ctx context.Context, orgID int64, limit, offset int) ([]*WorkflowInstance, int, error) {
	return s.repo.ListWorkflowInstances(ctx, orgID, limit, offset)
}

func (s *service) GetWorkflowInstanceByID(ctx context.Context, orgID, id int64) (*WorkflowInstance, error) {
	return s.repo.GetWorkflowInstanceByID(ctx, orgID, id)
}

func (s *service) RetryWorkflow(ctx context.Context, orgID, id int64, userID int64) (*WorkflowInstance, error) {
	wf, err := s.repo.GetWorkflowInstanceByID(ctx, orgID, id)
	if err != nil || wf == nil {
		return nil, fmt.Errorf("workflow instance not found")
	}

	evt, err := s.repo.GetEventByID(ctx, orgID, wf.EventID)
	if err != nil || evt == nil {
		return nil, fmt.Errorf("underlying event not found")
	}

	var payloadMap map[string]interface{}
	_ = json.Unmarshal(evt.Payload, &payloadMap)

	crossCtx := s.gatherCrossModuleContext(ctx, orgID, wf.SourceRecordType, wf.SourceRecordID)

	aiResp, aiErr := s.callSidecarAnalyzeEvent(ctx, evt.ID, orgID, evt.EventType, evt.SourceModule, evt.SourceRecordType, evt.SourceRecordID, payloadMap, crossCtx, wf.CorrelationID)
	if aiErr != nil {
		return nil, fmt.Errorf("retry AI analysis failed: %w", aiErr)
	}

	analysisBytes, _ := json.Marshal(aiResp)
	recsBytes, _ := json.Marshal(aiResp.RecommendedActions)

	wf.AISummary = aiResp.Summary
	wf.AIAnalysis = analysisBytes
	wf.Recommendations = recsBytes
	wf.RetryCount++
	wf.Status = "COMPLETED"
	wf.ActionStatus = "EXECUTED"

	_ = s.repo.UpdateWorkflowStatus(ctx, orgID, wf.ID, wf.Status, wf.ActionStatus, wf.ApprovalID, nil)
	return wf, nil
}

func (s *service) CancelWorkflow(ctx context.Context, orgID, id int64, reason string) (*WorkflowInstance, error) {
	wf, err := s.repo.GetWorkflowInstanceByID(ctx, orgID, id)
	if err != nil || wf == nil {
		return nil, fmt.Errorf("workflow instance not found")
	}

	msg := fmt.Sprintf("Workflow cancelled by operator: %s", reason)
	err = s.repo.UpdateWorkflowStatus(ctx, orgID, id, "CANCELLED", "SKIPPED", nil, &msg)
	if err != nil {
		return nil, err
	}
	wf.Status = "CANCELLED"
	wf.ActionStatus = "SKIPPED"
	wf.ErrorMessage = &msg
	return wf, nil
}

func stringPtr(s string) *string {
	return &s
}
