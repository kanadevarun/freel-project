package approvals

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/rbac"
	"github.com/jmoiron/sqlx"
)

type DraftApprovalHandler func(ctx context.Context, orgID int64, draftID int64, actorName string, notes string) error
type DraftRejectHandler func(ctx context.Context, orgID int64, draftID int64, actorName string, reason string) error
type ActionExecutor func(ctx context.Context, orgID int64, actionName string, input []byte, actorName string, userID int64, approvalRef string) (map[string]interface{}, error)
type ResumeExecutor func(ctx context.Context, orgID int64, threadID string, action string, notes string) error

type Service interface {
	ListApprovals(ctx context.Context, orgID int64) ([]*ApprovalRequest, error)
	GetApprovalByID(ctx context.Context, orgID int64, id int64) (*ApprovalRequest, error)
	CreateApproval(ctx context.Context, orgID int64, input *CreateApprovalInput, actorName string) (*ApprovalRequest, error)
	ProposeAIApproval(ctx context.Context, orgID int64, input *ProposeAIApprovalInput) (*ApprovalRequest, error)
	ApproveRequest(ctx context.Context, orgID int64, id int64, actorName string, userID int64, notes string) (*ApprovalRequest, error)
	RejectRequest(ctx context.Context, orgID int64, id int64, actorName string, userID int64, reason string, notes string) (*ApprovalRequest, error)
	ReturnRequest(ctx context.Context, orgID int64, id int64, actorName string, userID int64, reason string, notes string) (*ApprovalRequest, error)
	CancelRequest(ctx context.Context, orgID int64, id int64, actorName string, userID int64, notes string) (*ApprovalRequest, error)
	GetActionPreview(ctx context.Context, orgID int64, id int64) (*ActionPreview, error)
	GetDecisionHistory(ctx context.Context, orgID int64, id int64) ([]*ApprovalDecision, error)
	GetExecutionStatus(ctx context.Context, orgID int64, id int64) (map[string]interface{}, error)
	RetryExecution(ctx context.Context, orgID int64, id int64, actorName string, userID int64) (*ApprovalRequest, error)
	GetRelatedRecommendation(ctx context.Context, orgID int64, id int64) (map[string]interface{}, error)
	GetRelatedSourceRecord(ctx context.Context, orgID int64, id int64) (map[string]interface{}, error)
	GetAuditHistory(ctx context.Context, orgID int64, id int64) ([]map[string]interface{}, error)
	GetApprovalRequirements(ctx context.Context, orgID int64, actionName string) (*ApprovalRequirements, error)
	GetStats(ctx context.Context, orgID int64) (*ApprovalStats, error)
	SetDraftHandlers(approve DraftApprovalHandler, reject DraftRejectHandler)
	SetActionExecutor(executor ActionExecutor)
	SetResumeExecutor(executor ResumeExecutor)
	SetRBACService(rbacSvc rbac.Service)
	ResolveUserName(ctx context.Context, userID int64) string
}

type service struct {
	repo                Repository
	rbacSvc             rbac.Service
	draftApproveHandler DraftApprovalHandler
	draftRejectHandler  DraftRejectHandler
	actionExecutor      ActionExecutor
	resumeExecutor      ResumeExecutor
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) getDB() *sqlx.DB {
	if r, ok := s.repo.(*repository); ok && r != nil {
		return r.db
	}
	return nil
}

func (s *service) SetDraftHandlers(approve DraftApprovalHandler, reject DraftRejectHandler) {
	s.draftApproveHandler = approve
	s.draftRejectHandler = reject
}

func (s *service) SetActionExecutor(executor ActionExecutor) {
	s.actionExecutor = executor
}

func (s *service) SetResumeExecutor(executor ResumeExecutor) {
	s.resumeExecutor = executor
}

func (s *service) SetRBACService(rbacSvc rbac.Service) {
	s.rbacSvc = rbacSvc
}

func (s *service) ListApprovals(ctx context.Context, orgID int64) ([]*ApprovalRequest, error) {
	_, _ = s.repo.ExpireStaleApprovals(ctx, orgID)
	return s.repo.GetApprovalsByOrg(ctx, orgID)
}

func (s *service) GetApprovalByID(ctx context.Context, orgID int64, id int64) (*ApprovalRequest, error) {
	_, _ = s.repo.ExpireStaleApprovals(ctx, orgID)
	return s.repo.GetApprovalByID(ctx, orgID, id)
}

func (s *service) CreateApproval(ctx context.Context, orgID int64, input *CreateApprovalInput, actorName string) (*ApprovalRequest, error) {
	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	category := strings.ToUpper(input.Category)
	if category == "" {
		category = "DOCUMENTS"
	}

	appType := input.Type
	if appType == "" {
		switch category {
		case "DOCUMENTS":
			appType = "Document Approval"
		case "COMMERCIAL":
			appType = "Commercial Approval"
		case "OPERATIONS":
			appType = "Operations Approval"
		case "FINANCE":
			appType = "Finance Approval"
		default:
			appType = "General Approval"
		}
	}

	priority := strings.ToUpper(input.Priority)
	if priority == "" {
		priority = "MEDIUM"
	}

	reqCode := fmt.Sprintf("%s-APP-%d", category[:3], 1000+rand.Intn(9000))

	requester := input.RequestedByName
	if requester == "" {
		requester = actorName
	}
	if requester == "" {
		requester = "<IdentifiedUser>"
	}

	avatar := "AP"
	parts := strings.Fields(requester)
	if len(parts) >= 2 && len(parts[0]) > 0 && len(parts[1]) > 0 {
		avatar = fmt.Sprintf("%c%c", parts[0][0], parts[1][0])
	} else if len(parts) == 1 && len(parts[0]) > 0 {
		avatar = fmt.Sprintf("%c", parts[0][0])
	}

	var dueDatePtr *time.Time
	dueText := "7 days left"
	if input.DueDate != "" {
		if parsed, err := time.Parse("2006-01-02", input.DueDate); err == nil {
			dueDatePtr = &parsed
		} else if parsed, err := time.Parse(time.RFC3339, input.DueDate); err == nil {
			dueDatePtr = &parsed
		}
	}

	actorType := input.ActorType
	if actorType == "" {
		actorType = "USER"
	}
	riskLevel := input.RiskLevel
	if riskLevel == "" {
		riskLevel = "MEDIUM"
	}
	reqLevel := input.RequiredApprovalLevel
	if reqLevel == "" {
		reqLevel = "MANAGER"
	}

	req := &ApprovalRequest{
		OrgID:                 orgID,
		RequestCode:           reqCode,
		Title:                 input.Title,
		Category:              category,
		Type:                  appType,
		Status:                StatusPendingApproval,
		Priority:              priority,
		RequestedByName:       requester,
		Avatar:                &avatar,
		DueDate:               dueDatePtr,
		DueText:               &dueText,
		ActorType:             actorType,
		RiskLevel:             riskLevel,
		ExecutionStatus:       "NOT_STARTED",
		IsReversible:          input.IsReversible,
		ExternalCommunication: input.ExternalCommunication,
		RequiredApprovalLevel: reqLevel,
	}

	if input.RequestedByID > 0 {
		req.RequestedByID = &input.RequestedByID
	}
	if input.RelatedRef != "" {
		req.RelatedRef = &input.RelatedRef
	}
	if input.RelatedEntityType != "" {
		req.RelatedEntityType = &input.RelatedEntityType
	}
	if input.RelatedEntityID > 0 {
		req.RelatedEntityID = &input.RelatedEntityID
	}
	if input.CustomerName != "" {
		req.CustomerName = &input.CustomerName
	}
	if input.CustomerID > 0 {
		req.CustomerID = &input.CustomerID
	}
	if input.ShipmentID > 0 {
		req.ShipmentID = &input.ShipmentID
	}
	if input.DocumentID > 0 {
		req.DocumentID = &input.DocumentID
	}
	if input.BookingID > 0 {
		req.BookingID = &input.BookingID
	}
	if input.Department != "" {
		req.Department = &input.Department
	} else {
		dept := "Operations"
		req.Department = &dept
	}
	if input.Description != "" {
		req.Description = &input.Description
	}

	if input.Source != "" {
		req.Source = &input.Source
	}
	if input.ActionName != "" {
		req.ActionName = &input.ActionName
	}
	if input.RequiredPermission != "" {
		req.RequiredPermission = &input.RequiredPermission
	}
	if input.AITaskID > 0 {
		req.AITaskID = &input.AITaskID
	}
	if input.ThreadID != "" {
		req.ThreadID = &input.ThreadID
	}
	if input.CheckpointID != "" {
		req.CheckpointID = &input.CheckpointID
	}
	if input.ProposedPayload != "" {
		req.ProposedPayload = &input.ProposedPayload
	}
	if input.ApprovalReference != "" {
		req.ApprovalReference = &input.ApprovalReference
	}
	if input.CorrelationID != "" {
		req.CorrelationID = &input.CorrelationID
	}
	if input.IdempotencyKey != "" {
		req.IdempotencyKey = &input.IdempotencyKey
	}
	if input.SourceModule != "" {
		req.SourceModule = &input.SourceModule
	}
	if input.SourceRecordType != "" {
		req.SourceRecordType = &input.SourceRecordType
	}
	if input.SourceRecordID != "" {
		req.SourceRecordID = &input.SourceRecordID
	}
	if input.SourceRecordSnapshot != "" {
		req.SourceRecordSnapshot = &input.SourceRecordSnapshot
	}
	if input.Evidence != "" {
		req.Evidence = &input.Evidence
	}
	if input.ImpactSummary != "" {
		req.ImpactSummary = &input.ImpactSummary
	}

	err := s.repo.CreateApproval(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create approval request: %w", err)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorName:    requester,
		Action:       domain.ActionCreate,
		Module:       domain.ModuleApprovals,
		ResourceType: "APPROVAL",
		ResourceID:   fmt.Sprintf("%d", req.ID),
		ResourceName: req.RequestCode,
		Description:  fmt.Sprintf("Created approval request %s (%s)", req.RequestCode, req.Title),
		Result:       domain.ResultSuccess,
	})

	return req, nil
}

func (s *service) ProposeAIApproval(ctx context.Context, orgID int64, input *ProposeAIApprovalInput) (*ApprovalRequest, error) {
	if input == nil {
		return nil, errors.New("propose input cannot be nil")
	}

	// 1. Deduplication: check by ApprovalReference
	if input.ApprovalReference != "" {
		existing, err := s.repo.GetApprovalByReference(ctx, orgID, input.ApprovalReference)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	// 2. Deduplication: check by ActionName + ThreadID
	if input.ActionName != "" && input.ThreadID != "" {
		existing, err := s.repo.GetPendingApprovalByActionAndThread(ctx, orgID, input.ActionName, input.ThreadID)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	category := strings.ToUpper(input.Category)
	if category == "" {
		category = "COMMERCIAL"
	}
	appType := input.Type
	if appType == "" {
		appType = fmt.Sprintf("AI Action: %s", input.ActionName)
	}
	priority := strings.ToUpper(input.Priority)
	if priority == "" {
		priority = "HIGH"
	}
	riskLevel := strings.ToUpper(input.RiskLevel)
	if riskLevel == "" {
		riskLevel = "HIGH_RISK"
	}
	reqLevel := input.RequiredApprovalLevel
	if reqLevel == "" {
		if riskLevel == "HIGH_RISK" || riskLevel == "CRITICAL" {
			reqLevel = "ADMIN"
		} else {
			reqLevel = "MANAGER"
		}
	}

	reqCode := fmt.Sprintf("AI-APP-%d", 1000+rand.Intn(9000))
	requester := input.RequestedByName
	if requester == "" {
		requester = fmt.Sprintf("AI Agent (%s)", input.ActionName)
	}

	avatar := "AI"
	dept := input.Department
	if dept == "" {
		dept = "AI Intelligence"
	}

	hours := input.ExpiresInHours
	if hours <= 0 {
		hours = 48
	}
	expiresAt := time.Now().Add(time.Duration(hours) * time.Hour)
	dueText := fmt.Sprintf("%d hours left", hours)

	req := &ApprovalRequest{
		OrgID:                 orgID,
		RequestCode:           reqCode,
		Title:                 input.Title,
		Category:              category,
		Type:                  appType,
		Status:                StatusPendingApproval,
		Priority:              priority,
		RequestedByName:       requester,
		Department:            &dept,
		Avatar:                &avatar,
		DueDate:               &expiresAt,
		DueText:               &dueText,
		ExpiresAt:             &expiresAt,
		ActorType:             "AI_AGENT",
		RiskLevel:             riskLevel,
		Source:                &input.Source,
		ActionName:            &input.ActionName,
		RequiredPermission:    &input.RequiredPermission,
		ThreadID:              &input.ThreadID,
		CheckpointID:          &input.CheckpointID,
		ProposedPayload:       &input.ProposedPayload,
		ApprovalReference:     &input.ApprovalReference,
		CorrelationID:         &input.CorrelationID,
		IdempotencyKey:        &input.IdempotencyKey,
		ExecutionStatus:       "NOT_STARTED",
		IsReversible:          input.IsReversible,
		ExternalCommunication: input.ExternalCommunication,
		RequiredApprovalLevel: reqLevel,
	}

	if input.RelatedEntityType != "" {
		req.RelatedEntityType = &input.RelatedEntityType
	}
	if input.RelatedEntityID > 0 {
		req.RelatedEntityID = &input.RelatedEntityID
	}
	if input.RelatedRef != "" {
		req.RelatedRef = &input.RelatedRef
	}
	if input.CustomerName != "" {
		req.CustomerName = &input.CustomerName
	}
	if input.CustomerID > 0 {
		req.CustomerID = &input.CustomerID
	}
	if input.ShipmentID > 0 {
		req.ShipmentID = &input.ShipmentID
	}
	if input.Description != "" {
		req.Description = &input.Description
	}
	if input.SourceModule != "" {
		req.SourceModule = &input.SourceModule
	} else {
		req.SourceModule = &category
	}
	if input.SourceRecordType != "" {
		req.SourceRecordType = &input.SourceRecordType
	}
	if input.SourceRecordID != "" {
		req.SourceRecordID = &input.SourceRecordID
	}
	if input.SourceRecordSnapshot != "" {
		req.SourceRecordSnapshot = &input.SourceRecordSnapshot
	}
	if input.Evidence != "" {
		req.Evidence = &input.Evidence
	}
	if input.ImpactSummary != "" {
		req.ImpactSummary = &input.ImpactSummary
	}
	if input.AITaskID > 0 {
		req.AITaskID = &input.AITaskID
		// Mark ai_processing_task as WAITING_FOR_APPROVAL
		if db := s.getDB(); db != nil {
			_, _ = db.ExecContext(ctx, "UPDATE ai_processing_tasks SET status = 'WAITING_FOR_APPROVAL', updated_at = NOW() WHERE id = ? AND org_id = ?", input.AITaskID, orgID)
		}
	}

	err := s.repo.CreateApproval(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create proposed AI approval: %w", err)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorType:    domain.ActorTypeAIAgent,
		ActorName:    requester,
		Action:       "PROPOSE_ACTION",
		Module:       strings.ToUpper(category),
		ResourceType: "APPROVAL",
		ResourceID:   fmt.Sprintf("%d", req.ID),
		ResourceName: req.RequestCode,
		Description:  fmt.Sprintf("AI proposed high-risk action %s (%s)", input.ActionName, req.Title),
		Result:       "PROPOSED",
		Metadata: map[string]interface{}{
			"action_name":        input.ActionName,
			"approval_reference": input.ApprovalReference,
			"thread_id":          input.ThreadID,
			"risk_level":         riskLevel,
		},
	})

	return req, nil
}

func (s *service) ApproveRequest(ctx context.Context, orgID int64, id int64, actorName string, userID int64, notes string) (*ApprovalRequest, error) {
	if actorName == "" {
		actorName = "<IdentifiedUser>"
	}

	// Expire stale approvals first
	_, _ = s.repo.ExpireStaleApprovals(ctx, orgID)

	current, err := s.repo.GetApprovalByID(ctx, orgID, id)
	if err != nil || current == nil {
		return nil, fmt.Errorf("approval request %d not found or access denied", id)
	}

	// Check Expiration
	if strings.EqualFold(current.Status, StatusExpired) {
		return nil, fmt.Errorf("approval request %d has expired and cannot be approved", id)
	}
	if current.ExpiresAt != nil && current.ExpiresAt.Before(time.Now()) {
		_, _ = s.repo.UpdateApprovalStatus(ctx, orgID, id, StatusExpired, actorName, "Approval request expired", "")
		return nil, fmt.Errorf("approval request %d has expired and cannot be approved", id)
	}

	// Verify the approval is in an approvable state
	statusUpper := strings.ToUpper(current.Status)
	if statusUpper != "PENDING" && statusUpper != "PENDING_APPROVAL" && statusUpper != "IN REVIEW" && statusUpper != "RETURNED FOR CHANGES" {
		return nil, fmt.Errorf("conflict: approval request %d is in status '%s' and is no longer pending or reviewable", id, current.Status)
	}

	// Separation of Duties policy check: Requester cannot approve their own high-risk or financial actions
	isRequester := false
	if current.RequestedByID != nil && *current.RequestedByID == userID && userID > 0 {
		isRequester = true
	} else if actorName != "" && actorName != "<IdentifiedUser>" && strings.EqualFold(strings.TrimSpace(current.RequestedByName), strings.TrimSpace(actorName)) {
		isRequester = true
	}

	if isRequester {
		if strings.EqualFold(current.RiskLevel, "HIGH_RISK") || strings.EqualFold(current.RiskLevel, "CRITICAL") ||
			strings.EqualFold(current.Category, "FINANCE") || strings.EqualFold(current.Category, "COMMERCIAL") {
			return nil, fmt.Errorf("policy violation: separation of duties required. An operator cannot approve their own high-risk or commercial/financial request")
		}
	}

	// RBAC Permission Check at approval time
	if current.RequiredPermission != nil && *current.RequiredPermission != "" && s.rbacSvc != nil && userID > 0 {
		if db := s.getDB(); db != nil {
			var roleName string
			err := db.GetContext(ctx, &roleName, `
				SELECT r.name 
				FROM org_members om 
				JOIN roles r ON om.role_id = r.id 
				WHERE om.user_id = ? AND om.org_id = ? AND om.status = 'ACTIVE' 
				LIMIT 1
			`, userID, orgID)
			if err != nil || roleName == "" {
				return nil, fmt.Errorf("user #%d is not an active member of organization #%d", userID, orgID)
			}

			parts := strings.Split(*current.RequiredPermission, ":")
			if len(parts) == 2 {
				hasPerm, err := s.rbacSvc.HasPermission(ctx, roleName, orgID, parts[0], parts[1])
				if (!hasPerm || err != nil) && (roleName == rbac.RoleSuperAdmin || roleName == "ADMIN") {
					altRole := "ADMIN"
					if roleName == "ADMIN" {
						altRole = rbac.RoleSuperAdmin
					}
					hasPerm, err = s.rbacSvc.HasPermission(ctx, altRole, orgID, parts[0], parts[1])
				}
				if err != nil || !hasPerm {
					return nil, fmt.Errorf("forbidden: role '%s' lacks required permission '%s' to approve this request", roleName, *current.RequiredPermission)
				}
			}
		}
	}

	// Stale Action & Source Record Immutability Verification (Section 9)
	if current.SourceRecordType != nil && current.SourceRecordID != nil && current.SourceRecordSnapshot != nil && *current.SourceRecordSnapshot != "" {
		isStale, staleReason := s.checkSourceRecordStaleness(ctx, orgID, *current.SourceRecordType, *current.SourceRecordID, *current.SourceRecordSnapshot)
		if isStale {
			return nil, fmt.Errorf("conflict: action cannot be executed because target record changed materially: %s. Re-review or return for changes is required", staleReason)
		}
	}

	// Mark execution state as EXECUTING
	_ = s.repo.UpdateExecutionState(ctx, orgID, id, "EXECUTING", nil, nil)

	// Execute Action via Centralized Action System if action_name is set
	if current.ActionName != nil && *current.ActionName != "" && s.actionExecutor != nil {
		var payloadBytes []byte
		if current.ProposedPayload != nil && *current.ProposedPayload != "" {
			payloadBytes = []byte(*current.ProposedPayload)
		} else {
			payloadBytes = []byte("{}")
		}
		approvalRef := ""
		if current.ApprovalReference != nil {
			approvalRef = *current.ApprovalReference
		}

		execResult, execErr := s.actionExecutor(ctx, orgID, *current.ActionName, payloadBytes, actorName, userID, approvalRef)
		if execErr != nil {
			errStr := execErr.Error()
			_ = s.repo.UpdateExecutionState(ctx, orgID, id, "FAILED", nil, &errStr)
			_, _ = s.repo.UpdateApprovalStatus(ctx, orgID, id, StatusFailed, actorName, "Action execution failed: "+errStr, "")
			_ = s.repo.RecordDecision(ctx, orgID, id, current.ActionName, "FAILED", &userID, actorName, &errStr, &notes, current.CorrelationID)
			return nil, fmt.Errorf("failed to execute approved action '%s': %w", *current.ActionName, execErr)
		}

		resBytes, _ := json.Marshal(execResult)
		resStr := string(resBytes)
		_ = s.repo.UpdateExecutionState(ctx, orgID, id, "COMPLETED", &resStr, nil)
	} else {
		resStr := `{"executed": true, "type": "manual_signoff"}`
		_ = s.repo.UpdateExecutionState(ctx, orgID, id, "COMPLETED", &resStr, nil)
	}

	// Resume LangGraph workflow if thread_id is set
	if current.ThreadID != nil && *current.ThreadID != "" && s.resumeExecutor != nil {
		_ = s.resumeExecutor(ctx, orgID, *current.ThreadID, "APPROVE", notes)
	}

	app, err := s.repo.UpdateApprovalStatus(ctx, orgID, id, StatusApproved, actorName, notes, "")
	if err != nil {
		return nil, err
	}

	// Record immutable decision history
	_ = s.repo.RecordDecision(ctx, orgID, id, current.ActionName, "APPROVE", &userID, actorName, nil, &notes, current.CorrelationID)

	// Complete linked AI processing task if present
	if current.AITaskID != nil && *current.AITaskID > 0 {
		if db := s.getDB(); db != nil {
			_, _ = db.ExecContext(ctx, "UPDATE ai_processing_tasks SET status = 'COMPLETED', updated_at = NOW() WHERE id = ? AND org_id = ?", *current.AITaskID, orgID)
		}
	}

	// Cross-module sync with customer_invoices if this approval is for an Invoice
	if app != nil && app.RelatedEntityType != nil && strings.ToUpper(*app.RelatedEntityType) == "INVOICE" && app.RelatedEntityID != nil {
		invID := *app.RelatedEntityID
		if db := s.getDB(); db != nil {
			_, _ = db.ExecContext(ctx, "UPDATE customer_invoices SET status = 'Issued', updated_at = NOW() WHERE id = ? AND org_id = ?", invID, orgID)
			_, _ = db.ExecContext(ctx, "INSERT INTO customer_invoice_history (org_id, invoice_id, title, description, user_name, created_at) VALUES (?, ?, 'Invoice Approved', 'Manager approved request; status changed to Issued', ?, NOW())", orgID, invID, actorName)
		}
	}

	// Cross-module sync with lead_email_drafts if this approval is for a LEAD_EMAIL_DRAFT
	if app != nil && app.RelatedEntityType != nil && strings.ToUpper(*app.RelatedEntityType) == "LEAD_EMAIL_DRAFT" && app.RelatedEntityID != nil {
		draftID := *app.RelatedEntityID
		if s.draftApproveHandler != nil {
			_ = s.draftApproveHandler(ctx, orgID, draftID, actorName, notes)
		} else if db := s.getDB(); db != nil {
			_, _ = db.ExecContext(ctx, "UPDATE lead_email_drafts SET status = 'APPROVED', updated_at = NOW() WHERE id = ? AND org_id = ?", draftID, orgID)
		}
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorName:    actorName,
		Action:       domain.ActionApprove,
		Module:       domain.ModuleApprovals,
		ResourceType: "APPROVAL",
		ResourceID:   fmt.Sprintf("%d", id),
		ResourceName: app.RequestCode,
		Description:  fmt.Sprintf("Approved request %s (%s)", app.RequestCode, app.Title),
		Result:       domain.ResultSuccess,
	})

	return app, nil
}

func (s *service) RejectRequest(ctx context.Context, orgID int64, id int64, actorName string, userID int64, reason string, notes string) (*ApprovalRequest, error) {
	if actorName == "" {
		actorName = "<IdentifiedUser>"
	}
	if strings.TrimSpace(reason) == "" {
		return nil, fmt.Errorf("rejection reason is required")
	}

	current, err := s.repo.GetApprovalByID(ctx, orgID, id)
	if err != nil || current == nil {
		return nil, fmt.Errorf("approval request %d not found or access denied", id)
	}
	if strings.EqualFold(current.Status, StatusRejected) {
		return current, nil
	}
	if strings.EqualFold(current.Status, StatusApproved) || strings.EqualFold(current.Status, StatusCompleted) {
		return nil, fmt.Errorf("approval request %d is already approved and cannot be rejected", id)
	}
	if strings.EqualFold(current.Status, StatusCancelled) {
		return nil, fmt.Errorf("approval request %d is cancelled and cannot be rejected", id)
	}
	if strings.EqualFold(current.Status, StatusExpired) {
		return nil, fmt.Errorf("approval request %d has expired and cannot be rejected", id)
	}

	// RBAC Permission Check at rejection time
	if current.RequiredPermission != nil && *current.RequiredPermission != "" && s.rbacSvc != nil && userID > 0 {
		if db := s.getDB(); db != nil {
			var roleName string
			err := db.GetContext(ctx, &roleName, `
				SELECT r.name 
				FROM org_members om 
				JOIN roles r ON om.role_id = r.id 
				WHERE om.user_id = ? AND om.org_id = ? AND om.status = 'ACTIVE' 
				LIMIT 1
			`, userID, orgID)
			if err != nil || roleName == "" {
				return nil, fmt.Errorf("user #%d is not an active member of organization #%d", userID, orgID)
			}

			parts := strings.Split(*current.RequiredPermission, ":")
			if len(parts) == 2 {
				hasPerm, err := s.rbacSvc.HasPermission(ctx, roleName, orgID, parts[0], parts[1])
				if (!hasPerm || err != nil) && (roleName == rbac.RoleSuperAdmin || roleName == "ADMIN") {
					altRole := "ADMIN"
					if roleName == "ADMIN" {
						altRole = rbac.RoleSuperAdmin
					}
					hasPerm, err = s.rbacSvc.HasPermission(ctx, altRole, orgID, parts[0], parts[1])
				}
				if err != nil || !hasPerm {
					return nil, fmt.Errorf("forbidden: role '%s' lacks required permission '%s' to reject this request", roleName, *current.RequiredPermission)
				}
			}
		}
	}

	// Resume LangGraph workflow with REJECT if thread_id is set
	if current.ThreadID != nil && *current.ThreadID != "" && s.resumeExecutor != nil {
		_ = s.resumeExecutor(ctx, orgID, *current.ThreadID, "REJECT", reason)
	}

	// Cancel linked AI processing task if present
	if current.AITaskID != nil && *current.AITaskID > 0 {
		if db := s.getDB(); db != nil {
			_, _ = db.ExecContext(ctx, "UPDATE ai_processing_tasks SET status = 'CANCELLED', error_message = ?, updated_at = NOW() WHERE id = ? AND org_id = ?", reason, *current.AITaskID, orgID)
		}
	}

	app, err := s.repo.UpdateApprovalStatus(ctx, orgID, id, StatusRejected, actorName, notes, reason)
	if err != nil {
		return nil, err
	}

	// Record immutable decision history
	_ = s.repo.RecordDecision(ctx, orgID, id, current.ActionName, "REJECT", &userID, actorName, &reason, &notes, current.CorrelationID)

	// Cross-module sync with customer_invoices if this approval is for an Invoice
	if app != nil && app.RelatedEntityType != nil && strings.ToUpper(*app.RelatedEntityType) == "INVOICE" && app.RelatedEntityID != nil {
		invID := *app.RelatedEntityID
		if db := s.getDB(); db != nil {
			_, _ = db.ExecContext(ctx, "UPDATE customer_invoices SET status = 'Draft', updated_at = NOW() WHERE id = ? AND org_id = ?", invID, orgID)
			desc := "Approval request rejected by manager; status reverted to Draft"
			if reason != "" {
				desc += ". Reason: " + reason
			}
			_, _ = db.ExecContext(ctx, "INSERT INTO customer_invoice_history (org_id, invoice_id, title, description, user_name, created_at) VALUES (?, ?, 'Approval Rejected', ?, ?, NOW())", orgID, invID, desc, actorName)
		}
	}

	// Cross-module sync with lead_email_drafts if this approval is for a LEAD_EMAIL_DRAFT
	if app != nil && app.RelatedEntityType != nil && strings.ToUpper(*app.RelatedEntityType) == "LEAD_EMAIL_DRAFT" && app.RelatedEntityID != nil {
		draftID := *app.RelatedEntityID
		if s.draftRejectHandler != nil {
			_ = s.draftRejectHandler(ctx, orgID, draftID, actorName, reason)
		} else if db := s.getDB(); db != nil {
			_, _ = db.ExecContext(ctx, "UPDATE lead_email_drafts SET status = 'REJECTED', error_message = ?, updated_at = NOW() WHERE id = ? AND org_id = ?", reason, draftID, orgID)
		}
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorName:    actorName,
		Action:       domain.ActionReject,
		Module:       domain.ModuleApprovals,
		ResourceType: "APPROVAL",
		ResourceID:   fmt.Sprintf("%d", id),
		ResourceName: app.RequestCode,
		Description:  fmt.Sprintf("Rejected request %s. Reason: %s", app.RequestCode, reason),
		Result:       domain.ResultSuccess,
	})

	return app, nil
}

func (s *service) ReturnRequest(ctx context.Context, orgID int64, id int64, actorName string, userID int64, reason string, notes string) (*ApprovalRequest, error) {
	if actorName == "" {
		actorName = "<IdentifiedUser>"
	}
	if strings.TrimSpace(reason) == "" {
		return nil, fmt.Errorf("return reason is required")
	}

	current, err := s.repo.GetApprovalByID(ctx, orgID, id)
	if err != nil || current == nil {
		return nil, fmt.Errorf("approval request %d not found or access denied", id)
	}
	if strings.EqualFold(current.Status, StatusApproved) || strings.EqualFold(current.Status, StatusCompleted) {
		return nil, fmt.Errorf("cannot return an already approved approval request")
	}
	if strings.EqualFold(current.Status, StatusCancelled) {
		return nil, fmt.Errorf("cannot return a cancelled approval request")
	}
	if strings.EqualFold(current.Status, StatusExpired) {
		return nil, fmt.Errorf("cannot return an expired approval request")
	}

	app, err := s.repo.ReturnForChanges(ctx, orgID, id, actorName, reason, notes)
	if err != nil {
		return nil, err
	}

	// Record immutable decision history
	_ = s.repo.RecordDecision(ctx, orgID, id, current.ActionName, "RETURN_FOR_CHANGES", &userID, actorName, &reason, &notes, current.CorrelationID)

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorName:    actorName,
		Action:       "RETURN_FOR_CHANGES",
		Module:       domain.ModuleApprovals,
		ResourceType: "APPROVAL",
		ResourceID:   fmt.Sprintf("%d", id),
		ResourceName: app.RequestCode,
		Description:  fmt.Sprintf("Returned approval request %s for changes. Reason: %s", app.RequestCode, reason),
		Result:       domain.ResultSuccess,
	})

	return app, nil
}

func (s *service) CancelRequest(ctx context.Context, orgID int64, id int64, actorName string, userID int64, notes string) (*ApprovalRequest, error) {
	if actorName == "" {
		actorName = "<IdentifiedUser>"
	}

	current, err := s.repo.GetApprovalByID(ctx, orgID, id)
	if err != nil || current == nil {
		return nil, fmt.Errorf("approval request %d not found or access denied", id)
	}
	if strings.EqualFold(current.Status, StatusCancelled) {
		return current, nil
	}
	if strings.EqualFold(current.Status, StatusApproved) || strings.EqualFold(current.Status, StatusCompleted) {
		return nil, fmt.Errorf("cannot cancel an already approved approval request")
	}
	if strings.EqualFold(current.Status, StatusRejected) {
		return nil, fmt.Errorf("cannot cancel an already rejected approval request")
	}

	// Cancel linked AI processing task if present
	if current.AITaskID != nil && *current.AITaskID > 0 {
		if db := s.getDB(); db != nil {
			_, _ = db.ExecContext(ctx, "UPDATE ai_processing_tasks SET status = 'CANCELLED', error_message = 'Cancelled by user', updated_at = NOW() WHERE id = ? AND org_id = ?", *current.AITaskID, orgID)
		}
	}

	app, err := s.repo.CancelApproval(ctx, orgID, id, actorName, notes)
	if err != nil {
		return nil, err
	}

	// Record immutable decision history
	_ = s.repo.RecordDecision(ctx, orgID, id, current.ActionName, "CANCEL", &userID, actorName, nil, &notes, current.CorrelationID)

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorName:    actorName,
		Action:       "CANCEL",
		Module:       domain.ModuleApprovals,
		ResourceType: "APPROVAL",
		ResourceID:   fmt.Sprintf("%d", id),
		ResourceName: app.RequestCode,
		Description:  fmt.Sprintf("Cancelled approval request %s", app.RequestCode),
		Result:       domain.ResultSuccess,
	})

	return app, nil
}

func (s *service) GetActionPreview(ctx context.Context, orgID int64, id int64) (*ActionPreview, error) {
	current, err := s.repo.GetApprovalByID(ctx, orgID, id)
	if err != nil || current == nil {
		return nil, fmt.Errorf("approval request %d not found", id)
	}

	actionName := "general.approval"
	if current.ActionName != nil && *current.ActionName != "" {
		actionName = *current.ActionName
	}

	targetType := "RECORD"
	if current.RelatedEntityType != nil && *current.RelatedEntityType != "" {
		targetType = *current.RelatedEntityType
	}
	targetID := ""
	if current.RelatedEntityID != nil {
		targetID = fmt.Sprintf("%d", *current.RelatedEntityID)
	} else if current.RelatedRef != nil {
		targetID = *current.RelatedRef
	}

	targetTitle := current.Title
	if current.RelatedRef != nil && *current.RelatedRef != "" {
		targetTitle = fmt.Sprintf("%s (%s)", current.Title, *current.RelatedRef)
	}

	proposedMap := make(map[string]interface{})
	if current.ProposedPayload != nil && *current.ProposedPayload != "" {
		_ = json.Unmarshal([]byte(*current.ProposedPayload), &proposedMap)
	}

	// Fetch live target record for Current Value and Stale Check
	currentMap := make(map[string]interface{})
	isStale := false
	staleReason := ""

	db := s.getDB()
	if db != nil && strings.EqualFold(targetType, "INVOICE") && current.RelatedEntityID != nil {
		var inv struct {
			InvoiceNumber string  `db:"invoice_number"`
			Status        string  `db:"status"`
			TotalAmount   float64 `db:"total_amount"`
		}
		err := db.GetContext(ctx, &inv, "SELECT invoice_number, status, total_amount FROM customer_invoices WHERE id = ? AND org_id = ?", *current.RelatedEntityID, orgID)
		if err == nil {
			currentMap["invoice_number"] = inv.InvoiceNumber
			currentMap["status"] = inv.Status
			currentMap["total_amount"] = inv.TotalAmount
		} else {
			isStale = true
			staleReason = "Associated invoice record not found or inaccessible"
		}
	} else if db != nil && strings.EqualFold(targetType, "SHIPMENT") && current.RelatedEntityID != nil {
		var ship struct {
			ShipmentNo string `db:"shipment_no"`
			Status     string `db:"status"`
			Carrier    string `db:"carrier"`
		}
		err := db.GetContext(ctx, &ship, "SELECT shipment_no, status, carrier FROM shipments WHERE id = ? AND org_id = ?", *current.RelatedEntityID, orgID)
		if err == nil {
			currentMap["shipment_no"] = ship.ShipmentNo
			currentMap["status"] = ship.Status
			currentMap["carrier"] = ship.Carrier
		} else {
			isStale = true
			staleReason = "Associated shipment record not found"
		}
	}

	// Build domains affected
	domains := []string{}
	switch strings.ToUpper(current.Category) {
	case "FINANCE":
		domains = append(domains, "FINANCIAL")
	case "COMMERCIAL":
		domains = append(domains, "COMMERCIAL", "CONTRACTUAL")
	case "OPERATIONS":
		domains = append(domains, "OPERATIONAL")
	case "DOCUMENTS":
		domains = append(domains, "COMPLIANCE", "DOCUMENTATION")
	default:
		domains = append(domains, "GENERAL")
	}

	// Check if external communication is involved
	hasExtComm := current.ExternalCommunication
	var msgPreview *MessagePreviewDetails
	if strings.Contains(strings.ToLower(actionName), "email") || strings.Contains(strings.ToLower(actionName), "clarification") || strings.Contains(strings.ToLower(current.Type), "email") {
		hasExtComm = true
		recipient := ""
		subject := ""
		body := ""

		if r, ok := proposedMap["recipient"].(string); ok {
			recipient = r
		} else if r, ok := proposedMap["to_email"].(string); ok {
			recipient = r
		} else if r, ok := proposedMap["lead_email"].(string); ok {
			recipient = r
		}
		if s, ok := proposedMap["subject"].(string); ok {
			subject = s
		}
		if b, ok := proposedMap["body"].(string); ok {
			body = b
		} else if b, ok := proposedMap["content"].(string); ok {
			body = b
		}

		if recipient == "" && current.CustomerName != nil {
			recipient = *current.CustomerName
		}
		if body == "" && current.Description != nil {
			body = *current.Description
		}

		msgPreview = &MessagePreviewDetails{
			Recipient: recipient,
			Channel:   "EMAIL",
			Subject:   subject,
			Body:      body,
		}
	}

	reasonStr := "Business operation requiring human sign-off"
	if current.Description != nil && *current.Description != "" {
		reasonStr = *current.Description
	}

	evidenceStr := "Autonomous intelligence detected optimization or required action."
	if current.Evidence != nil && *current.Evidence != "" {
		evidenceStr = *current.Evidence
	}

	impactStr := fmt.Sprintf("Execution will mutate %s record #%s in module %s.", targetType, targetID, current.Category)
	if current.ImpactSummary != nil && *current.ImpactSummary != "" {
		impactStr = *current.ImpactSummary
	}

	preview := &ActionPreview{
		ApprovalID:            current.ID,
		ActionName:            actionName,
		ActionDescription:     current.Title,
		TargetRecordType:      targetType,
		TargetRecordID:        targetID,
		TargetRecordTitle:     targetTitle,
		CurrentValue:          currentMap,
		ProposedValue:         proposedMap,
		RequestedBy:           current.RequestedByName,
		Reason:                reasonStr,
		Evidence:              evidenceStr,
		ExpectedImpact:        impactStr,
		RiskLevel:             current.RiskLevel,
		ExternalCommunication: hasExtComm,
		DataDomainsAffected:   domains,
		IsReversible:          current.IsReversible,
		RequiredApprovalLevel: current.RequiredApprovalLevel,
		IsStale:               isStale,
		StaleReason:           staleReason,
		MessagePreview:        msgPreview,
	}

	return preview, nil
}

func (s *service) GetDecisionHistory(ctx context.Context, orgID int64, id int64) ([]*ApprovalDecision, error) {
	return s.repo.GetDecisionHistory(ctx, orgID, id)
}

func (s *service) GetExecutionStatus(ctx context.Context, orgID int64, id int64) (map[string]interface{}, error) {
	current, err := s.repo.GetApprovalByID(ctx, orgID, id)
	if err != nil || current == nil {
		return nil, fmt.Errorf("approval request %d not found", id)
	}

	return map[string]interface{}{
		"approval_id":       current.ID,
		"status":            current.Status,
		"execution_status":  current.ExecutionStatus,
		"execution_result":  current.ExecutionResult,
		"execution_error":   current.ExecutionError,
		"execution_retries": current.ExecutionRetries,
		"action_name":       current.ActionName,
		"updated_at":        current.UpdatedAt,
	}, nil
}

func (s *service) RetryExecution(ctx context.Context, orgID int64, id int64, actorName string, userID int64) (*ApprovalRequest, error) {
	current, err := s.repo.GetApprovalByID(ctx, orgID, id)
	if err != nil || current == nil {
		return nil, fmt.Errorf("approval request %d not found", id)
	}

	if current.ExecutionStatus != "FAILED" && !strings.EqualFold(current.Status, StatusFailed) {
		return nil, fmt.Errorf("only failed actions can be retried; current execution status is '%s'", current.ExecutionStatus)
	}

	if current.ActionName == nil || *current.ActionName == "" || s.actionExecutor == nil {
		return nil, errors.New("no executable action registered for this approval request")
	}

	_ = s.repo.UpdateExecutionState(ctx, orgID, id, "EXECUTING", nil, nil)

	var payloadBytes []byte
	if current.ProposedPayload != nil && *current.ProposedPayload != "" {
		payloadBytes = []byte(*current.ProposedPayload)
	} else {
		payloadBytes = []byte("{}")
	}
	approvalRef := ""
	if current.ApprovalReference != nil {
		approvalRef = *current.ApprovalReference
	}

	execResult, execErr := s.actionExecutor(ctx, orgID, *current.ActionName, payloadBytes, actorName, userID, approvalRef)
	if execErr != nil {
		errStr := execErr.Error()
		_ = s.repo.UpdateExecutionState(ctx, orgID, id, "FAILED", nil, &errStr)
		return nil, fmt.Errorf("retry execution failed: %w", execErr)
	}

	resBytes, _ := json.Marshal(execResult)
	resStr := string(resBytes)
	_ = s.repo.UpdateExecutionState(ctx, orgID, id, "COMPLETED", &resStr, nil)
	return s.repo.UpdateApprovalStatus(ctx, orgID, id, StatusApproved, actorName, "Execution succeeded upon retry", "")
}

func (s *service) GetRelatedRecommendation(ctx context.Context, orgID int64, id int64) (map[string]interface{}, error) {
	current, err := s.repo.GetApprovalByID(ctx, orgID, id)
	if err != nil || current == nil {
		return nil, fmt.Errorf("approval request %d not found", id)
	}

	db := s.getDB()
	if db == nil {
		return map[string]interface{}{"linked": false, "message": "Database not available"}, nil
	}
	var rec struct {
		ID                int64   `db:"id" json:"id"`
		Title             string  `db:"title" json:"title"`
		Category          string  `db:"category" json:"category"`
		Priority          string  `db:"priority" json:"priority"`
		RiskLevel         string  `db:"risk_level" json:"risk_level"`
		Status            string  `db:"status" json:"status"`
		RecommendedAction string  `db:"recommended_action" json:"recommended_action"`
		ActionType        string  `db:"action_type" json:"action_type"`
		Evidence          string  `db:"evidence" json:"evidence"`
		CorrelationID     string  `db:"correlation_id" json:"correlation_id"`
		CustomerName      *string `db:"customer_name" json:"customer_name,omitempty"`
	}

	ref := ""
	if current.ApprovalReference != nil {
		ref = *current.ApprovalReference
	}
	corr := ""
	if current.CorrelationID != nil {
		corr = *current.CorrelationID
	}

	query := `
		SELECT id, title, category, priority, risk_level, status, recommended_action, action_type, evidence, correlation_id, customer_name
		FROM ai_recommendations
		WHERE org_id = ? AND (approval_id = ? OR correlation_id = ? OR (? != '' AND correlation_id = ?))
		LIMIT 1
	`
	err = db.GetContext(ctx, &rec, query, orgID, id, ref, corr, corr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return map[string]interface{}{"linked": false, "message": "No linked recommendation found for this approval"}, nil
		}
		return nil, err
	}

	return map[string]interface{}{
		"linked":             true,
		"recommendation_id":  rec.ID,
		"title":              rec.Title,
		"category":           rec.Category,
		"priority":           rec.Priority,
		"risk_level":         rec.RiskLevel,
		"status":             rec.Status,
		"recommended_action": rec.RecommendedAction,
		"action_type":        rec.ActionType,
		"evidence":           rec.Evidence,
		"correlation_id":     rec.CorrelationID,
		"customer_name":      rec.CustomerName,
	}, nil
}

func (s *service) GetRelatedSourceRecord(ctx context.Context, orgID int64, id int64) (map[string]interface{}, error) {
	current, err := s.repo.GetApprovalByID(ctx, orgID, id)
	if err != nil || current == nil {
		return nil, fmt.Errorf("approval request %d not found", id)
	}

	if current.RelatedEntityType == nil || current.RelatedEntityID == nil {
		return map[string]interface{}{"resolved": false, "message": "No linked entity specified on approval request"}, nil
	}

	db := s.getDB()
	if db == nil {
		return map[string]interface{}{"resolved": false, "message": "Database not available"}, nil
	}

	entityType := strings.ToUpper(*current.RelatedEntityType)
	entityID := *current.RelatedEntityID

	switch entityType {
	case "INVOICE":
		var inv struct {
			ID            int64   `db:"id" json:"id"`
			InvoiceNumber string  `db:"invoice_number" json:"invoice_number"`
			CustomerID    int64   `db:"customer_id" json:"customer_id"`
			CustomerName  string  `db:"customer_name" json:"customer_name"`
			TotalAmount   float64 `db:"total_amount" json:"total_amount"`
			Status        string  `db:"status" json:"status"`
			CreatedAt     string  `db:"created_at" json:"created_at"`
		}
		err := db.GetContext(ctx, &inv, `
			SELECT ci.id, ci.invoice_number, ci.customer_id, c.name AS customer_name, ci.total_amount, ci.status, ci.created_at
			FROM customer_invoices ci
			LEFT JOIN customers c ON ci.customer_id = c.id
			WHERE ci.id = ? AND ci.org_id = ?
		`, entityID, orgID)
		if err == nil {
			return map[string]interface{}{"resolved": true, "type": "INVOICE", "data": inv}, nil
		}
	case "SHIPMENT":
		var ship struct {
			ID          int64   `db:"id" json:"id"`
			ShipmentNo  string  `db:"shipment_no" json:"shipment_no"`
			Status      string  `db:"status" json:"status"`
			Carrier     string  `db:"carrier" json:"carrier"`
			Origin      string  `db:"origin" json:"origin"`
			Destination string  `db:"destination" json:"destination"`
		}
		err := db.GetContext(ctx, &ship, `
			SELECT id, shipment_no, status, carrier, origin, destination
			FROM shipments
			WHERE id = ? AND org_id = ?
		`, entityID, orgID)
		if err == nil {
			return map[string]interface{}{"resolved": true, "type": "SHIPMENT", "data": ship}, nil
		}
	case "RFQ":
		var rfq struct {
			ID        int64  `db:"id" json:"id"`
			RFQNumber string `db:"rfq_number" json:"rfq_number"`
			Status    string `db:"status" json:"status"`
			Shipper   string `db:"shipper" json:"shipper"`
		}
		err := db.GetContext(ctx, &rfq, `
			SELECT id, rfq_number, status, shipper
			FROM rfqs
			WHERE id = ? AND org_id = ?
		`, entityID, orgID)
		if err == nil {
			return map[string]interface{}{"resolved": true, "type": "RFQ", "data": rfq}, nil
		}
	case "CONTRACT":
		var contract struct {
			ID             int64  `db:"id" json:"id"`
			ContractNumber string `db:"contract_number" json:"contract_number"`
			CarrierName    string `db:"carrier_name" json:"carrier_name"`
			Status         string `db:"status" json:"status"`
		}
		err := db.GetContext(ctx, &contract, `
			SELECT id, contract_number, carrier_name, status
			FROM rate_contracts
			WHERE id = ? AND org_id = ?
		`, entityID, orgID)
		if err == nil {
			return map[string]interface{}{"resolved": true, "type": "CONTRACT", "data": contract}, nil
		}
	}

	return map[string]interface{}{
		"resolved":    false,
		"entity_type": entityType,
		"entity_id":   entityID,
		"message":     "Target record not found or module adapter unavailable",
	}, nil
}

func (s *service) GetAuditHistory(ctx context.Context, orgID int64, id int64) ([]map[string]interface{}, error) {
	current, err := s.repo.GetApprovalByID(ctx, orgID, id)
	if err != nil || current == nil {
		return nil, fmt.Errorf("approval request %d not found", id)
	}

	db := s.getDB()
	if db == nil {
		return []map[string]interface{}{}, nil
	}
	type auditRow struct {
		ID          int64     `db:"id" json:"id"`
		Action      string    `db:"action" json:"action"`
		ActorType   string    `db:"actor_type" json:"actor_type"`
		ActorName   *string   `db:"actor_name" json:"actor_name"`
		Description *string   `db:"description" json:"description"`
		Result      string    `db:"result" json:"result"`
		CreatedAt   time.Time `db:"created_at" json:"created_at"`
	}

	var rows []auditRow
	idStr := fmt.Sprintf("%d", id)
	err = db.SelectContext(ctx, &rows, `
		SELECT id, action, actor_type, actor_name, description, result, created_at
		FROM audit_logs
		WHERE org_id = ? AND resource_type = 'APPROVAL' AND (resource_id = ? OR resource_name = ?)
		ORDER BY created_at DESC
		LIMIT 50
	`, orgID, idStr, current.RequestCode)

	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"id":          r.ID,
			"action":      r.Action,
			"actor_type":  r.ActorType,
			"actor_name":  r.ActorName,
			"description": r.Description,
			"result":      r.Result,
			"created_at":  r.CreatedAt,
		})
	}

	return result, nil
}

func (s *service) GetApprovalRequirements(ctx context.Context, orgID int64, actionName string) (*ApprovalRequirements, error) {
	if actionName == "" {
		return nil, errors.New("action_name is required")
	}

	category := "HIGH_RISK"
	riskLevel := "HIGH_RISK"
	perm := "approvals:approve"
	reqLevel := "MANAGER"
	sepDuties := true

	switch actionName {
	case "shipments.update_milestone", "shipments.create_exception", "shipments.operations_callback":
		perm = "operations:update"
		reqLevel = "OPERATIONS_LEAD"
	case "pricing.save_draft_quotes", "pricing.apply_selected_rate":
		perm = "pricing:update"
		reqLevel = "PRICING_MANAGER"
	case "sales.send_clarification_email", "sales.create_rfq_from_email", "sales.convert_lead":
		perm = "sales:create"
		reqLevel = "SALES_DIRECTOR"
	case "finance.reconcile_invoice":
		perm = "billing:update"
		reqLevel = "FINANCE_ADMIN"
	case "contracts.ingest_rates", "contracts.review_extraction":
		perm = "contracts:create"
		reqLevel = "LEGAL_COUNSEL"
	case "compliance.record_compliance_discrepancies":
		perm = "compliance:update"
		reqLevel = "COMPLIANCE_OFFICER"
	}

	return &ApprovalRequirements{
		ActionName:            actionName,
		Category:              category,
		RiskLevel:             riskLevel,
		RequiresConfirmation:  true,
		RequiredPermission:    perm,
		RequiredApprovalLevel: reqLevel,
		SeparationOfDuties:    sepDuties,
		ApprovalPolicySummary: fmt.Sprintf("High-risk action '%s' requires explicit human confirmation by an authorized %s with permission '%s'.", actionName, reqLevel, perm),
	}, nil
}

func (s *service) GetStats(ctx context.Context, orgID int64) (*ApprovalStats, error) {
	_, _ = s.repo.ExpireStaleApprovals(ctx, orgID)
	return s.repo.GetApprovalStats(ctx, orgID)
}

func (s *service) ResolveUserName(ctx context.Context, userID int64) string {
	if userID <= 0 {
		return "<IdentifiedUser>"
	}
	if name, err := s.repo.ResolveUserName(ctx, userID); err == nil && name != "" {
		return name
	}
	return fmt.Sprintf("User #%d", userID)
}

func (s *service) checkSourceRecordStaleness(ctx context.Context, orgID int64, recordType string, recordID string, snapshot string) (bool, string) {
	db := s.getDB()
	if db == nil {
		return false, ""
	}
	switch strings.ToUpper(recordType) {
	case "INVOICE":
		var currentStatus string
		err := db.GetContext(ctx, &currentStatus, "SELECT status FROM customer_invoices WHERE id = ? AND org_id = ?", recordID, orgID)
		if err != nil {
			return true, "Invoice record no longer exists or access was revoked"
		}
		if snapshot != "" && !strings.Contains(snapshot, currentStatus) {
			return true, fmt.Sprintf("Invoice status changed to '%s' since approval was requested", currentStatus)
		}
	case "SHIPMENT":
		var currentStatus string
		err := db.GetContext(ctx, &currentStatus, "SELECT status FROM shipments WHERE id = ? AND org_id = ?", recordID, orgID)
		if err != nil {
			return true, "Shipment record no longer exists"
		}
		if snapshot != "" && !strings.Contains(snapshot, currentStatus) {
			return true, fmt.Sprintf("Shipment status updated to '%s' since approval was requested", currentStatus)
		}
	}
	return false, ""
}
