package recommendations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/freel/backend/internal/approvals"
	auditDomain "github.com/freel/backend/internal/audit/domain"
	auditService "github.com/freel/backend/internal/audit/service"
)

var (
	ErrNotFound          = errors.New("recommendation not found")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrReasonRequired    = errors.New("dismissal reason is required")
	ErrApprovalRequired  = errors.New("recommendation requires approval before it can be marked completed")
)

// Service defines recommendation center application operations
type Service interface {
	ListRecommendations(ctx context.Context, orgID int64, filter RecommendationFilter) (*RecommendationListResponse, error)
	GetRecommendation(ctx context.Context, orgID int64, id int64) (*Recommendation, error)
	ListBySource(ctx context.Context, orgID int64, sourceType string, sourceID int64) ([]*Recommendation, error)
	GenerateRecommendations(ctx context.Context, orgID int64, correlationID string, userID int64) (*GenerateResult, error)
	UpdateStatus(ctx context.Context, orgID int64, id int64, input UpdateStatusInput, userID int64, userName string, correlationID string) (*Recommendation, error)
	AssignRecommendation(ctx context.Context, orgID int64, id int64, input AssignInput, userID int64, userName string, correlationID string) (*Recommendation, error)
	DismissRecommendation(ctx context.Context, orgID int64, id int64, input DismissInput, userID int64, userName string, correlationID string) (*Recommendation, error)
	MarkReviewed(ctx context.Context, orgID int64, id int64, userID int64, userName string, correlationID string) (*Recommendation, error)
	GetStats(ctx context.Context, orgID int64) (*RecommendationStats, error)

	// Phase 2 Task 2.2: Customer Follow-Up Assistant operations
	GenerateDraft(ctx context.Context, orgID int64, id int64, userID int64, userName string, correlationID string) (*Recommendation, error)
	SaveDraft(ctx context.Context, orgID int64, id int64, input SaveDraftInput, userID int64, userName string, correlationID string) (*Recommendation, error)
	CreateFollowupTask(ctx context.Context, orgID int64, id int64, input CreateFollowupTaskInput, userID int64, userName string, correlationID string) (*CustomerFollowupTask, error)
	ListFollowupTasks(ctx context.Context, orgID int64, customerID *int64, status string) ([]*CustomerFollowupTask, error)
	GetFollowupStats(ctx context.Context, orgID int64) (*FollowupStats, error)

	// Phase 2 Task 2.3: RFQ & Quotation Workflow Assistant operations
	GetActionPreview(ctx context.Context, orgID int64, id int64, userID int64, userName string, correlationID string) (*ActionPreview, error)
	RequestApproval(ctx context.Context, orgID int64, id int64, input RequestApprovalInput, userID int64, userName string, correlationID string) (*Recommendation, error)
	SetApprovalsService(approvalsSvc approvals.Service)

	// Phase 2 Task 2.4: Shipment Exception & Operations Copilot operations
	GetEntityEvidence(ctx context.Context, orgID int64, entityType string, entityID int64) ([]EvidenceItem, error)

	// Phase 2 Task 2.5: Invoice and Collections Assistant operations
	GetInvoiceEvidence(ctx context.Context, orgID int64, invoiceID int64) (*InvoiceEvidencePayload, error)
	GetCustomerCollectionSummary(ctx context.Context, orgID int64, customerID int64) (*CustomerCollectionSummaryPayload, error)

	// Phase 2 Task 2.6: Contract, Document, and Compliance Assistant operations
	GetContractEvidence(ctx context.Context, orgID int64, contractID int64) (*ContractEvidencePayload, error)
	GetDocumentEvidence(ctx context.Context, orgID int64, docID string) (*DocumentEvidencePayload, error)
	GetComplianceEvidence(ctx context.Context, orgID int64, compID int64) (*ComplianceEvidencePayload, error)
	GetContractComplianceSummary(ctx context.Context, orgID int64) (*ContractComplianceSummaryPayload, error)
}

type service struct {
	repo         Repository
	generator    Generator
	auditSvc     auditService.Service
	approvalsSvc approvals.Service
}

// NewService creates a new recommendations service
func NewService(repo Repository, generator Generator, auditSvc auditService.Service) Service {
	return &service{
		repo:      repo,
		generator: generator,
		auditSvc:  auditSvc,
	}
}

func (s *service) SetApprovalsService(approvalsSvc approvals.Service) {
	s.approvalsSvc = approvalsSvc
}

func (s *service) ListRecommendations(ctx context.Context, orgID int64, filter RecommendationFilter) (*RecommendationListResponse, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization id: %d", orgID)
	}

	items, total, err := s.repo.List(ctx, orgID, filter)
	if err != nil {
		return nil, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	stats, _ := s.repo.GetStats(ctx, orgID)

	return &RecommendationListResponse{
		Recommendations: items,
		Pagination: PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
		Stats: stats,
	}, nil
}

func (s *service) GetRecommendation(ctx context.Context, orgID int64, id int64) (*Recommendation, error) {
	if orgID <= 0 || id <= 0 {
		return nil, ErrNotFound
	}

	rec, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, ErrNotFound
	}
	return rec, nil
}

func (s *service) ListBySource(ctx context.Context, orgID int64, sourceType string, sourceID int64) ([]*Recommendation, error) {
	if orgID <= 0 || sourceType == "" || sourceID <= 0 {
		return []*Recommendation{}, nil
	}
	return s.repo.ListBySource(ctx, orgID, sourceType, sourceID)
}

func (s *service) GenerateRecommendations(ctx context.Context, orgID int64, correlationID string, userID int64) (*GenerateResult, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization id: %d", orgID)
	}

	res, err := s.generator.Generate(ctx, orgID, correlationID, userID)
	if err != nil {
		s.recordAudit(ctx, orgID, userID, "SYSTEM", "RECOMMENDATION_GENERATE_FAILED", "0", "Generation Failed", fmt.Sprintf("Failed to generate recommendations: %v", err), correlationID, "FAILURE", err.Error())
		return nil, err
	}

	s.recordAudit(ctx, orgID, userID, "SYSTEM", "RECOMMENDATION_GENERATE", fmt.Sprintf("%d", res.TotalEvaluated), "Recommendations Generated", fmt.Sprintf("Evaluated %d records. Created: %d, Updated: %d", res.TotalEvaluated, res.CreatedCount, res.UpdatedCount), correlationID, "SUCCESS", "")

	return res, nil
}

func (s *service) UpdateStatus(ctx context.Context, orgID int64, id int64, input UpdateStatusInput, userID int64, userName string, correlationID string) (*Recommendation, error) {
	rec, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	targetStatus := strings.ToLower(input.Status)
	if !IsValidTransition(rec.Status, targetStatus) {
		s.recordAudit(ctx, orgID, userID, userName, "RECOMMENDATION_INVALID_TRANSITION", fmt.Sprintf("%d", id), rec.Title, fmt.Sprintf("Attempted invalid transition from %s to %s", rec.Status, targetStatus), correlationID, "FAILURE", "Invalid status transition")
		return nil, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidTransition, rec.Status, targetStatus)
	}

	// High risk / approval safety gate
	if targetStatus == StatusCompleted && rec.RequiresApproval && rec.Status != StatusApproved {
		return nil, ErrApprovalRequired
	}

	if targetStatus == StatusDismissed && strings.TrimSpace(input.Reason) == "" {
		return nil, ErrReasonRequired
	}

	var uid *int64
	if userID > 0 {
		uid = &userID
	}
	var rsn *string
	if input.Reason != "" {
		rsn = &input.Reason
	}

	if err := s.repo.UpdateStatus(ctx, id, orgID, targetStatus, uid, rsn); err != nil {
		return nil, err
	}

	updated, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, orgID, userID, userName, fmt.Sprintf("RECOMMENDATION_STATUS_%s", strings.ToUpper(targetStatus)), fmt.Sprintf("%d", id), rec.Title, fmt.Sprintf("Status changed from %s to %s: %s", rec.Status, targetStatus, input.Reason), correlationID, "SUCCESS", "")

	return updated, nil
}

func (s *service) AssignRecommendation(ctx context.Context, orgID int64, id int64, input AssignInput, userID int64, userName string, correlationID string) (*Recommendation, error) {
	rec, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	if input.AssigneeID <= 0 || strings.TrimSpace(input.AssigneeName) == "" {
		return nil, fmt.Errorf("assignee_id and assignee_name are required")
	}

	if err := s.repo.Assign(ctx, id, orgID, input.AssigneeID, input.AssigneeName); err != nil {
		return nil, err
	}

	updated, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, orgID, userID, userName, "RECOMMENDATION_ASSIGNED", fmt.Sprintf("%d", id), rec.Title, fmt.Sprintf("Assigned recommendation to %s (ID: %d)", input.AssigneeName, input.AssigneeID), correlationID, "SUCCESS", "")

	return updated, nil
}

func (s *service) DismissRecommendation(ctx context.Context, orgID int64, id int64, input DismissInput, userID int64, userName string, correlationID string) (*Recommendation, error) {
	rec, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return nil, ErrReasonRequired
	}

	if err := s.repo.Dismiss(ctx, id, orgID, userID, reason); err != nil {
		return nil, err
	}

	updated, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, orgID, userID, userName, "RECOMMENDATION_DISMISSED", fmt.Sprintf("%d", id), rec.Title, fmt.Sprintf("Dismissed recommendation with reason: %s", reason), correlationID, "SUCCESS", "")

	return updated, nil
}

func (s *service) MarkReviewed(ctx context.Context, orgID int64, id int64, userID int64, userName string, correlationID string) (*Recommendation, error) {
	rec, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	if !IsValidTransition(rec.Status, StatusReviewed) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidTransition, rec.Status, StatusReviewed)
	}

	if err := s.repo.MarkReviewed(ctx, id, orgID, userID); err != nil {
		return nil, err
	}

	updated, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, orgID, userID, userName, "RECOMMENDATION_REVIEWED", fmt.Sprintf("%d", id), rec.Title, "Marked recommendation as reviewed", correlationID, "SUCCESS", "")

	return updated, nil
}

func (s *service) GetStats(ctx context.Context, orgID int64) (*RecommendationStats, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization id: %d", orgID)
	}
	return s.repo.GetStats(ctx, orgID)
}

type pythonDraftRequest struct {
	OrgID               int64                 `json:"org_id"`
	CustomerID          int64                 `json:"customer_id"`
	CustomerName        string                `json:"customer_name"`
	PrimaryContactName  *string               `json:"primary_contact_name,omitempty"`
	PrimaryContactEmail *string               `json:"primary_contact_email,omitempty"`
	OperatorName        string                `json:"operator_name"`
	DraftType           string                `json:"draft_type"`
	SourceModule        string                `json:"source_module"`
	SourceRecordID      int64                 `json:"source_record_id"`
	SourceReference     string                `json:"source_reference"`
	Description         string                `json:"description"`
	RecommendedAction   string                `json:"recommended_action"`
	Evidence            []pythonDraftEvidence `json:"evidence"`
	CorrelationID       string                `json:"correlation_id"`
}

type pythonDraftEvidence struct {
	SourceModule   string `json:"source_module"`
	SourceEntityID int64  `json:"source_entity_id"`
	SourceRef      string `json:"source_ref"`
	FieldName      string `json:"field_name"`
	ObservedValue  string `json:"observed_value"`
	Description    string `json:"description"`
}

type pythonDraftResponse struct {
	Status              string   `json:"status"`
	Subject             string   `json:"subject"`
	Body                string   `json:"body"`
	SuggestedRecipients []string `json:"suggested_recipients"`
	Confidence          float64  `json:"confidence"`
	RequiresApproval    bool     `json:"requires_approval"`
	CorrelationID       string   `json:"correlation_id"`
}

func (s *service) callPythonCustomerDraft(ctx context.Context, rec *Recommendation, custName string, operator string, draftType string, correlationID string) (*pythonDraftResponse, error) {
	sidecarURL := os.Getenv("AI_SIDECAR_URL")
	if sidecarURL == "" {
		sidecarURL = "http://localhost:8090"
	}
	apiKey := os.Getenv("LOGISTICSHQ_INTERNAL_SERVICE_KEY")
	if apiKey == "" {
		apiKey = "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456"
	}

	var evList []pythonDraftEvidence
	for _, ev := range rec.Evidence {
		evList = append(evList, pythonDraftEvidence{
			SourceModule:   ev.SourceModule,
			SourceEntityID: ev.SourceEntityID,
			SourceRef:      ev.SourceRef,
			FieldName:      ev.FieldName,
			ObservedValue:  fmt.Sprintf("%v", ev.ObservedValue),
			Description:    ev.Description,
		})
	}

	reqPayload := pythonDraftRequest{
		OrgID:             rec.OrgID,
		CustomerID:        0,
		CustomerName:      custName,
		OperatorName:      operator,
		DraftType:         draftType,
		SourceModule:      rec.SourceType,
		SourceRecordID:    rec.SourceID,
		SourceReference:   rec.SourceReference,
		Description:       rec.Description,
		RecommendedAction: rec.RecommendedAction,
		Evidence:          evList,
		CorrelationID:     correlationID,
	}
	if rec.CustomerID != nil {
		reqPayload.CustomerID = *rec.CustomerID
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/customer-relationship/draft", sidecarURL), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", apiKey)

	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar draft returned status: %d", resp.StatusCode)
	}

	var draftResp pythonDraftResponse
	if err := json.NewDecoder(resp.Body).Decode(&draftResp); err != nil {
		return nil, err
	}
	return &draftResp, nil
}

func (s *service) GenerateDraft(ctx context.Context, orgID int64, id int64, userID int64, userName string, correlationID string) (*Recommendation, error) {
	rec, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	custName := "Customer"
	if rec.CustomerName != nil && *rec.CustomerName != "" {
		custName = *rec.CustomerName
	}

	operator := userName
	if operator == "" {
		operator = "Operations Representative"
	}

	ft := ""
	if rec.FollowupType != nil {
		ft = *rec.FollowupType
	}
	if ft == "" {
		if rec.Category == CategoryContract || rec.SourceType == SourceContract || rec.ActionType == ActionTypePrepareRenewalReminder || rec.ActionType == ActionTypeReviewExpiringContract {
			ft = DraftTypeContractRenewalReminder
		} else if rec.Category == CategoryCompliance || rec.SourceType == SourceCompliance || rec.ActionType == ActionTypeReviewComplianceChecklist {
			ft = DraftTypeComplianceFollowupNotice
		} else if rec.SourceType == SourceDocument || rec.ActionType == ActionTypeRequestMissingDocument {
			ft = DraftTypeMissingDocumentRequest
		}
	}

	var subject, body string
	if sidecarResp, sErr := s.callPythonCustomerDraft(ctx, rec, custName, operator, ft, correlationID); sErr == nil && sidecarResp != nil && sidecarResp.Subject != "" && sidecarResp.Body != "" {
		subject = sidecarResp.Subject
		body = sidecarResp.Body
	} else {
		switch ft {
		case DraftTypeContractRenewalReminder:
		var evLines []string
		for _, ev := range rec.Evidence {
			evLines = append(evLines, fmt.Sprintf("• %s: %s (%s)", ev.FieldName, ev.ObservedValue, ev.Description))
		}
		evText := strings.Join(evLines, "\n")
		subject = fmt.Sprintf("Contract Term & Renewal Notice: Agreement %s - %s", rec.SourceReference, custName)
		body = fmt.Sprintf("Dear %s Commercial Team,\n\nRegarding master agreement %s:\n\nOur contract review system has flagged that this agreement is approaching its conclusion date or renewal notification window.\n\nContract Overview:\n- Reference: %s\n- Party: %s\n- Details: %s\n\nVerified Record Details:\n%s\n\nRecommended Next Step:\n%s\n\nPlease let us know your preferred timeline to review term commitments, volume performance, and renewal options.\n\nSincerely,\n%s\nLogisticsHQ Commercial & Legal Desk", custName, rec.SourceReference, rec.SourceReference, custName, rec.Description, evText, rec.RecommendedAction, operator)

	case DraftTypeComplianceFollowupNotice:
		var evLines []string
		for _, ev := range rec.Evidence {
			evLines = append(evLines, fmt.Sprintf("• %s: %s (%s)", ev.FieldName, ev.ObservedValue, ev.Description))
		}
		evText := strings.Join(evLines, "\n")
		subject = fmt.Sprintf("Compliance Verification Follow-up: Requirement for Contract %s", rec.SourceReference)
		body = fmt.Sprintf("Dear %s Compliance Desk,\n\nWe are conducting our periodic trade compliance review regarding master agreement %s.\n\nRequirement Summary:\n%s\n\nRecorded Observations:\n%s\n\nRecommended Action:\n%s\n\nPlease furnish the updated documentation or verification certificates so our compliance team can update your account status.\n\nSincerely,\n%s\nLogisticsHQ Trade Compliance Team", custName, rec.SourceReference, rec.Description, evText, rec.RecommendedAction, operator)

	case DraftTypeInternalContractReview:
		var evLines []string
		for _, ev := range rec.Evidence {
			evLines = append(evLines, fmt.Sprintf("• %s: %s (%s)", ev.FieldName, ev.ObservedValue, ev.Description))
		}
		evText := strings.Join(evLines, "\n")
		subject = fmt.Sprintf("Internal Review Note: Contract %s Expiry & Terms Verification", rec.SourceReference)
		body = fmt.Sprintf("Internal Contract & Risk Memorandum\n\nContract Reference: %s\nCounterparty: %s\n\nCondition:\n%s\n\nSupporting Verified Evidence:\n%s\n\nRecommended Next Step:\n%s\n\nPlease evaluate renewal terms, legal liability, and rate card margins before confirming renewal.\n\nLogged by: %s\nLogisticsHQ Legal & Contract Intelligence", rec.SourceReference, custName, rec.Description, evText, rec.RecommendedAction, operator)

	case DraftTypeDocumentVerificationQuery:
		var evLines []string
		for _, ev := range rec.Evidence {
			evLines = append(evLines, fmt.Sprintf("• %s: %s (%s)", ev.FieldName, ev.ObservedValue, ev.Description))
		}
		evText := strings.Join(evLines, "\n")
		subject = fmt.Sprintf("Document Verification Inquiry: %s", rec.SourceReference)
		body = fmt.Sprintf("Dear %s Operations Team,\n\nWe are reviewing transport and compliance documentation on file for reference %s.\n\nObservations:\n%s\n\nDocument Evidence:\n%s\n\nAction Required:\n%s\n\nPlease review and advise if an amended document or formal explanation can be provided.\n\nSincerely,\n%s\nLogisticsHQ Document Workspace", custName, rec.SourceReference, rec.Description, evText, rec.RecommendedAction, operator)

	case DraftTypeCarrierComplianceEscalation:
		subject = fmt.Sprintf("Carrier Compliance Notice: Regulatory Document Expiry for %s", rec.SourceReference)
		body = fmt.Sprintf("Carrier Operations & Compliance Management,\n\nThis notice concerns pending compliance documentation for carrier reference %s.\n\nSummary:\n%s\n\nNext Step Required:\n%s\n\nPlease submit updated insurance certificate or regulatory documentation to avoid operational hold on upcoming bookings.\n\nSincerely,\n%s\nLogisticsHQ Carrier Compliance", rec.SourceReference, rec.Description, rec.RecommendedAction, operator)

	case DraftTypeRFQClarification, DraftTypeMissingInfoRequest:
		var missingDetails []string
		for _, ev := range rec.Evidence {
			missingDetails = append(missingDetails, fmt.Sprintf("• %s: %s (%s)", ev.FieldName, ev.ObservedValue, ev.Description))
		}
		missingList := strings.Join(missingDetails, "\n")
		if missingList == "" {
			missingList = "• Commercial lane routing, cargo dimensions, and trade incoterms"
		}
		subject = fmt.Sprintf("LogisticsHQ Clarification Request: RFQ %s Missing Commercial Specifications", rec.SourceReference)
		body = fmt.Sprintf("Dear %s Commercial Team,\n\nRegarding freight inquiry %s:\nOur pricing desk has reviewed your submission and requires clarification on the following items before we can finalize competitive carrier rates:\n\n%s\n\nRecommended next step: %s\n\nPlease provide these details at your earliest convenience so we can complete your commercial quotation.\n\nSincerely,\n%s\nLogisticsHQ Commercial Team", custName, rec.SourceReference, missingList, rec.RecommendedAction, operator)

	case DraftTypeQuotationApproval:
		subject = fmt.Sprintf("Internal Review: Quotation %s Commercial Approval Request", rec.SourceReference)
		body = fmt.Sprintf("Commercial & Pricing Management,\n\nQuotation %s for %s requires review and approval prior to customer release.\n\nSummary:\n- Reference: %s\n- Client: %s\n- Assessment: %s\n\nRecommended Action:\n%s\n\nPlease inspect rate components and record approval in LogisticsHQ.\n\nSubmitted by: %s\nLogisticsHQ Pricing Desk", rec.SourceReference, custName, rec.SourceReference, custName, rec.Description, rec.RecommendedAction, operator)

	case DraftTypeInternalPricingReview:
		subject = fmt.Sprintf("Internal Note: Quotation %s Pricing & Margin Verification", rec.SourceReference)
		body = fmt.Sprintf("Pricing Review Desk,\n\nPlease evaluate quotation %s for %s.\n\nIdentified pricing concern:\n%s\n\nRecommended verification:\n%s\n\nEnsure carrier buy rates, accessorials, and required margin percentage meet commercial guidelines before releasing.\n\nLogged by: %s\nCommercial Intelligence System", rec.SourceReference, custName, rec.Description, rec.RecommendedAction, operator)

	case DraftTypeCustomerPricingClarification:
		subject = fmt.Sprintf("LogisticsHQ Commercial Notice: Quotation %s Rate Specifications", rec.SourceReference)
		body = fmt.Sprintf("Dear %s Procurement Team,\n\nFollowing up regarding quotation %s:\n%s\n\nRecommended next step: %s\n\nPlease let us know if you require adjustments to equipment types, cargo ready dates, or lane options.\n\nSincerely,\n%s\nLogisticsHQ Commercial Desk", custName, rec.SourceReference, rec.Description, rec.RecommendedAction, operator)

	case DraftTypeCarrierInfoRequest:
		subject = fmt.Sprintf("Rate & Space Request: RFQ %s Carrier Inquiry", rec.SourceReference)
		body = fmt.Sprintf("Carrier Operations Desk,\n\nRegarding upcoming shipment requirement for inquiry %s:\n%s\n\nPlease transmit valid spot ocean/air freight rates, vessel departure schedules, and equipment availability free-time terms.\n\nRecommended action: %s\n\nSincerely,\n%s\nLogisticsHQ Carrier Procurement", rec.SourceReference, rec.Description, rec.RecommendedAction, operator)

	case DraftTypeFirstPaymentReminder:
		subject = fmt.Sprintf("Courtesy Payment Reminder: Invoice %s - %s", rec.SourceReference, custName)
		body = fmt.Sprintf("Dear %s Accounts Payable,\n\nThis is a courtesy reminder regarding upcoming invoice %s.\n\nInvoice Details:\n- Reference: %s\n- Account: %s\n- Status: %s\n\n%s\n\nIf payment has already been scheduled or remitted, please reply with your remittance advice or transaction reference so we can ensure proper allocation.\n\nThank you for your business,\n%s\nLogisticsHQ Credit & Receivables Desk", custName, rec.SourceReference, rec.SourceReference, custName, rec.Status, rec.Description, operator)

	case DraftTypeOverduePaymentNotice:
		var evLines []string
		for _, ev := range rec.Evidence {
			evLines = append(evLines, fmt.Sprintf("• %s: %s", ev.FieldName, ev.ObservedValue))
		}
		evText := strings.Join(evLines, "\n")
		subject = fmt.Sprintf("Payment Overdue Notice: Invoice %s - %s", rec.SourceReference, custName)
		body = fmt.Sprintf("Dear %s Accounts Payable Desk,\n\nWe are contacting you regarding outstanding invoice %s, which is currently overdue according to our records.\n\nSummary of Account:\n- Invoice Number: %s\n- Customer: %s\n%s\n\nDetails:\n%s\n\nPlease confirm when payment will be remitted, or forward transaction confirmation if payment was recently processed.\n\nSincerely,\n%s\nLogisticsHQ Accounts Receivable", custName, rec.SourceReference, rec.SourceReference, custName, evText, rec.Description, operator)

	case DraftTypeUrgentCollectionEscalation:
		subject = fmt.Sprintf("Urgent Payment Escalation: Delinquent Balance for %s (Invoice %s)", custName, rec.SourceReference)
		body = fmt.Sprintf("Attention: %s Finance Leadership,\n\nThis is an urgent notice regarding severely delinquent accounts receivable balance associated with invoice %s.\n\nAccount Summary:\n- Client: %s\n- Reference: %s\n- Overview: %s\n\nNext Action Required:\n%s\n\nPlease provide immediate remittance confirmation to prevent credit hold status or disruption to active freight operations.\n\nRespectfully,\n%s\nLogisticsHQ Credit Control & Treasury", custName, rec.SourceReference, custName, rec.SourceReference, rec.Description, rec.RecommendedAction, operator)

	case DraftTypePaymentReconciliationQuery:
		subject = fmt.Sprintf("Payment Reconciliation Inquiry: Invoice %s - %s", rec.SourceReference, custName)
		body = fmt.Sprintf("Dear %s Accounts Team,\n\nWe are reviewing recent settlement records for invoice %s and would like to reconcile the outstanding ledger balance.\n\nLedger Observations:\n%s\n\nCould you please share your latest remittance advice or transaction receipt detailing bank transfer date and amount so our finance team can accurately clear this item?\n\nSincerely,\n%s\nLogisticsHQ Billing & Settlement Desk", custName, rec.SourceReference, rec.Description, operator)

	case DraftTypeFinanceInternalEscalation:
		var evLines []string
		for _, ev := range rec.Evidence {
			evLines = append(evLines, fmt.Sprintf("• %s: %s (%s)", ev.FieldName, ev.ObservedValue, ev.Description))
		}
		evSummary := strings.Join(evLines, "\n")
		subject = fmt.Sprintf("Internal Finance Escalation: Invoice %s Audit Review", rec.SourceReference)
		body = fmt.Sprintf("Internal Finance Escalation Memo\n\nReference: %s\nDebtor: %s\n\nIssue Description:\n%s\n\nEvidence & Records Checked:\n%s\n\nRecommended Finance Action:\n%s\n\nPlease review in the Finance Ledger / Approvals Console.\n\nLogged by: %s\nLogisticsHQ Financial Intelligence", rec.SourceReference, custName, rec.Description, evSummary, rec.RecommendedAction, operator)

	case DraftTypeCustomerStatementSummary:
		subject = fmt.Sprintf("Statement of Account & Delinquent Balance Summary: %s", custName)
		body = fmt.Sprintf("Dear %s Finance & Accounts Payable,\n\nPlease find an updated summary of outstanding and overdue invoices for your account.\n\nOverview:\n%s\n\nRecommended Action:\n%s\n\nPlease review your accounts and advise on payment schedule for the open balance.\n\nSincerely,\n%s\nLogisticsHQ Credit Management", custName, rec.Description, rec.RecommendedAction, operator)

	case FollowupTypeInvoiceReminder:
		subject = fmt.Sprintf("LogisticsHQ Payment Reminder: %s (%s)", rec.SourceReference, custName)
		body = fmt.Sprintf("Dear %s Finance Desk,\n\nI am writing from LogisticsHQ to follow up regarding outstanding invoice %s for %s.\n\nOur accounts records indicate this invoice remains open with an outstanding balance awaiting settlement.\n\nSummary:\n- Reference: %s\n- Details: %s\n\nRecommended next step: %s\n\nPlease let us know if remittance has already been processed or if you require an updated statement of account.\n\nSincerely,\n%s\nLogisticsHQ Finance & Operations", custName, rec.SourceReference, custName, rec.SourceReference, rec.Description, rec.RecommendedAction, operator)

	case FollowupTypeExceptionResolution:
		subject = fmt.Sprintf("LogisticsHQ Operational Update: Shipment %s", rec.SourceReference)
		body = fmt.Sprintf("Dear %s Logistics Team,\n\nI am contacting you to provide an operational status update regarding your shipment %s.\n\n%s\n\nOur freight operations desk is actively addressing this matter. Recommended action: %s\n\nPlease let us know if you require any specific coordination or transit adjustments.\n\nSincerely,\n%s\nLogisticsHQ Operations Desk", custName, rec.SourceReference, rec.Description, rec.RecommendedAction, operator)

	case FollowupTypeRFQFollowup:
		subject = fmt.Sprintf("LogisticsHQ Rate Inquiry Follow-Up: %s", rec.SourceReference)
		body = fmt.Sprintf("Dear %s Shipping Team,\n\nFollowing up on freight inquiry %s.\n\nOur pricing team has reviewed the lane requirements. %s\n\nNext step: %s\n\nPlease advise if you have updated cargo ready dates or volume requirements.\n\nSincerely,\n%s\nLogisticsHQ Commercial Team", custName, rec.SourceReference, rec.Description, rec.RecommendedAction, operator)

	case FollowupTypeQuotationFollowup:
		subject = fmt.Sprintf("LogisticsHQ Follow-Up: Quotation %s Commercial Notice", rec.SourceReference)
		body = fmt.Sprintf("Dear %s Procurement Lead,\n\nI am following up regarding quotation %s.\n\n%s\n\nTo ensure current rate validity and secure vessel space allocation, please let us know if you would like to proceed with booking.\n\nNext step: %s\n\nSincerely,\n%s\nLogisticsHQ Sales & Commercial Desk", custName, rec.SourceReference, rec.Description, rec.RecommendedAction, operator)

	case FollowupTypeContractRenewal:
		subject = fmt.Sprintf("LogisticsHQ Agreement Review: %s Renewal", rec.SourceReference)
		body = fmt.Sprintf("Dear %s Commercial Management,\n\nAs part of our periodic freight service review, we noted that agreement %s is approaching its renewal window.\n\n%s\n\nWe would appreciate scheduling an account discussion to review trade lane volume commitments and rate structures for the upcoming period.\n\nNext step: %s\n\nSincerely,\n%s\nLogisticsHQ Commercial Management", custName, rec.SourceReference, rec.Description, rec.RecommendedAction, operator)

	case DraftTypeInternalOperationsNote:
		var evidenceLines []string
		for _, ev := range rec.Evidence {
			evidenceLines = append(evidenceLines, fmt.Sprintf("• %s: %s (%s)", ev.FieldName, ev.ObservedValue, ev.Description))
		}
		evSummary := strings.Join(evidenceLines, "\n")
		if evSummary == "" {
			evSummary = "• Operational record flags verified."
		}
		subject = fmt.Sprintf("Internal Note: Operational Assessment for Shipment %s", rec.SourceReference)
		body = fmt.Sprintf("Internal Operations Assessment - Shipment %s\n\nIdentified Operational Condition:\n%s\n\nSupporting Verified Evidence:\n%s\n\nRecommended Operational Action:\n%s\n\nNext Steps:\n1. Verify carrier tracking and milestone status with designated logistics coordinator.\n2. Confirm all document requirements before cargo movement.\n3. Keep consignee advised via official updates once status is physically confirmed.\n\nLogged by: %s\nLogisticsHQ Operations Intelligence", rec.SourceReference, rec.Description, evSummary, rec.RecommendedAction, operator)

	case DraftTypeCustomerShipmentUpdate:
		// Safe customer-facing update. MUST NOT expose internal margin, buy costs, or internal hold notes.
		// Avoid unverified delivery commitments.
		subject = fmt.Sprintf("LogisticsHQ Shipment Status Update: %s", rec.SourceReference)
		body = fmt.Sprintf("Dear %s Logistics Desk,\n\nWe are contacting you with an operational update regarding shipment %s.\n\nStatus Summary:\nOur operations team is actively tracking this consignment. Based on carrier records currently verified in our system:\n- Reference: %s\n- Current Overview: %s\n\nNext Step:\n%s\n\nPlease rest assured our logistics operations desk is monitoring this movement closely. We will provide further verified transit updates as milestones progress.\n\nSincerely,\n%s\nLogisticsHQ Freight Operations Desk", custName, rec.SourceReference, rec.SourceReference, rec.Description, rec.RecommendedAction, operator)

	case DraftTypeExceptionEscalation:
		var evidenceLines []string
		for _, ev := range rec.Evidence {
			evidenceLines = append(evidenceLines, fmt.Sprintf("• %s: %s (%s)", ev.FieldName, ev.ObservedValue, ev.Description))
		}
		evSummary := strings.Join(evidenceLines, "\n")
		subject = fmt.Sprintf("Operational Escalation: Exception Flagged on Shipment %s", rec.SourceReference)
		body = fmt.Sprintf("Urgent Operational Escalation\n\nShipment Reference: %s\nClient: %s\n\nException Summary:\n%s\n\nVerified Evidence:\n%s\n\nRecommended Intervention:\n%s\n\nPlease route through operations management for expedited carrier escalation and resource allocation.\n\nEscalated by: %s\nLogisticsHQ Operations Desk", rec.SourceReference, custName, rec.Description, evSummary, rec.RecommendedAction, operator)

	case DraftTypeCarrierClarification:
		subject = fmt.Sprintf("Urgent Status Clarification: Shipment %s Milestone Verification", rec.SourceReference)
		body = fmt.Sprintf("Carrier Operations & Vessel Desk,\n\nRegarding shipment booking %s:\n\nOur tracking system indicates the following milestone status requires clarification:\n%s\n\nAction Requested:\n%s\n\nPlease confirm container status, latest vessel position, and estimated schedule at your earliest convenience.\n\nSincerely,\n%s\nLogisticsHQ Ocean & Air Operations", rec.SourceReference, rec.Description, rec.RecommendedAction, operator)

	case DraftTypeMissingDocumentRequest:
		var missingLines []string
		for _, ev := range rec.Evidence {
			missingLines = append(missingLines, fmt.Sprintf("• %s: %s (%s)", ev.FieldName, ev.ObservedValue, ev.Description))
		}
		mList := strings.Join(missingLines, "\n")
		if mList == "" {
			mList = "• Commercial invoice, packing list, or transport reference details"
		}
		subject = fmt.Sprintf("Information Required: Operational Details for Shipment %s", rec.SourceReference)
		body = fmt.Sprintf("Dear %s Shipping Team,\n\nRegarding active shipment %s:\nTo ensure smooth customs clearance and carrier processing, our operations team requires the following information:\n\n%s\n\nRecommended action: %s\n\nPlease forward these details to our operations coordinator so transport documentation can be finalized.\n\nSincerely,\n%s\nLogisticsHQ Operations Coordinator", custName, rec.SourceReference, mList, rec.RecommendedAction, operator)

	case DraftTypeDeliveryClarification:
		subject = fmt.Sprintf("Delivery Status Inquiry: Shipment %s Past Planned ETA", rec.SourceReference)
		body = fmt.Sprintf("Destination Port / Terminal Operations,\n\nRegarding shipment %s:\n\nOur system indicates the scheduled estimated time of arrival has elapsed without confirmed gate-out / delivery confirmation.\n\nCondition:\n%s\n\nRequested Next Step:\n%s\n\nPlease advise current container discharge status and cargo availability.\n\nSincerely,\n%s\nLogisticsHQ Destination Operations Desk", rec.SourceReference, rec.Description, rec.RecommendedAction, operator)

	case DraftTypeInternalHandoffNote:
		subject = fmt.Sprintf("Operations Handoff: Shipment %s Active Status", rec.SourceReference)
		body = fmt.Sprintf("Operations Shift Handoff - Shipment %s\n\nAccount: %s\nStatus Summary: %s\n\nOperational Next Step: %s\n\nHandoff logged by: %s\nLogisticsHQ Operations", rec.SourceReference, custName, rec.Description, rec.RecommendedAction, operator)

	case FollowupTypeAccountReview, FollowupTypeServiceRecovery:
		subject = fmt.Sprintf("LogisticsHQ Account Review: %s", custName)
		body = fmt.Sprintf("Dear %s Executive Team,\n\nI am reaching out on behalf of LogisticsHQ account management regarding your logistics account status.\n\n%s\n\nWe are committed to delivering seamless freight operations and would like to coordinate a brief review to ensure all operational matters are resolved.\n\nNext step: %s\n\nSincerely,\n%s\nLogisticsHQ Customer Success", custName, rec.Description, rec.RecommendedAction, operator)

	default: // General check-in
		subject = fmt.Sprintf("LogisticsHQ Freight Account Check-In: %s", custName)
		body = fmt.Sprintf("Dear %s Shipping Team,\n\nI am checking in on behalf of LogisticsHQ regarding your ongoing freight forwarding requirements.\n\n%s\n\nRecommended next step: %s\n\nPlease let us know if you have upcoming shipping inquiries or lane reviews where our team can assist.\n\nSincerely,\n%s\nLogisticsHQ Customer Success", custName, rec.Description, rec.RecommendedAction, operator)
	}
	}

	if err := s.repo.SaveDraft(ctx, id, orgID, subject, body, "DRAFTED"); err != nil {
		return nil, fmt.Errorf("failed to save generated draft: %w", err)
	}

	updated, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, orgID, userID, userName, "CUSTOMER_FOLLOWUP_DRAFT_GENERATED", fmt.Sprintf("%d", id), rec.Title, fmt.Sprintf("Generated %s communication draft for %s", ft, custName), correlationID, "SUCCESS", "")
	return updated, nil
}

func (s *service) SaveDraft(ctx context.Context, orgID int64, id int64, input SaveDraftInput, userID int64, userName string, correlationID string) (*Recommendation, error) {
	rec, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	subj := strings.TrimSpace(input.Subject)
	if subj == "" {
		return nil, errors.New("draft subject is required")
	}
	body := strings.TrimSpace(input.Body)
	if body == "" {
		return nil, errors.New("draft body is required")
	}

	if err := s.repo.SaveDraft(ctx, id, orgID, subj, body, "SAVED"); err != nil {
		return nil, fmt.Errorf("failed to save draft: %w", err)
	}

	updated, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, orgID, userID, userName, "CUSTOMER_FOLLOWUP_DRAFT_SAVED", fmt.Sprintf("%d", id), rec.Title, "Saved edited communication draft", correlationID, "SUCCESS", "")
	return updated, nil
}

func (s *service) CreateFollowupTask(ctx context.Context, orgID int64, id int64, input CreateFollowupTaskInput, userID int64, userName string, correlationID string) (*CustomerFollowupTask, error) {
	rec, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	priority := input.Priority
	if priority == "" {
		priority = rec.Priority
	}

	custID := int64(0)
	if rec.CustomerID != nil {
		custID = *rec.CustomerID
	}
	custName := rec.SourceReference
	if rec.CustomerName != nil && *rec.CustomerName != "" {
		custName = *rec.CustomerName
	}
	ft := rec.Category
	if rec.FollowupType != nil && *rec.FollowupType != "" {
		ft = *rec.FollowupType
	}

	var assigneeID *int64
	var assigneeName *string
	if input.AssigneeID != nil && *input.AssigneeID > 0 {
		assigneeID = input.AssigneeID
		assigneeName = input.AssigneeName
	} else if rec.SuggestedOwnerID != nil && *rec.SuggestedOwnerID > 0 {
		assigneeID = rec.SuggestedOwnerID
		assigneeName = rec.SuggestedOwnerName
	} else if userID > 0 {
		assigneeID = &userID
		assigneeName = &userName
	}

	var notes *string
	if strings.TrimSpace(input.Notes) != "" {
		n := strings.TrimSpace(input.Notes)
		notes = &n
	}

	task := &CustomerFollowupTask{
		OrgID:            orgID,
		RecommendationID: id,
		CustomerID:       custID,
		CustomerName:     custName,
		SourceType:       rec.SourceType,
		SourceID:         rec.SourceID,
		SourceReference:  rec.SourceReference,
		FollowupType:     ft,
		Title:            fmt.Sprintf("Follow-up: %s", rec.Title),
		Reason:           rec.Description,
		SuggestedAction:  rec.RecommendedAction,
		Priority:         priority,
		AssigneeID:       assigneeID,
		AssigneeName:     assigneeName,
		DueDate:          input.DueDate,
		Status:           "pending",
		CreatedByID:      &userID,
		CreatedByName:    userName,
		CorrelationID:    correlationID,
		Notes:            notes,
	}

	created, err := s.repo.CreateFollowupTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to create followup task: %w", err)
	}

	s.recordAudit(ctx, orgID, userID, userName, "CUSTOMER_FOLLOWUP_TASK_CREATED", fmt.Sprintf("%d", id), rec.Title, fmt.Sprintf("Created internal follow-up task #%d for %s", created.ID, custName), correlationID, "SUCCESS", "")
	return created, nil
}

func (s *service) ListFollowupTasks(ctx context.Context, orgID int64, customerID *int64, status string) ([]*CustomerFollowupTask, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization id: %d", orgID)
	}
	return s.repo.ListFollowupTasks(ctx, orgID, customerID, status)
}

func (s *service) GetFollowupStats(ctx context.Context, orgID int64) (*FollowupStats, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization id: %d", orgID)
	}
	return s.repo.GetFollowupStats(ctx, orgID)
}

func (s *service) GetActionPreview(ctx context.Context, orgID int64, id int64, userID int64, userName string, correlationID string) (*ActionPreview, error) {
	rec, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	userLabel := userName
	if userLabel == "" {
		userLabel = fmt.Sprintf("user-%d", userID)
	}

	proposedAction := rec.ActionType
	expectedEffect := rec.RecommendedAction
	switch rec.ActionType {
	case ActionTypeRequestRFQClarification:
		proposedAction = "Request Missing RFQ Information"
		expectedEffect = "Prepare clarification inquiry for shipper specifying all missing commercial routing and cargo dimensions."
	case ActionTypeAssignRFQOwner:
		proposedAction = "Assign RFQ Commercial Owner"
		expectedEffect = "Assign active business development representative to expedite rate sourcing and quote generation."
	case ActionTypeSourceRates:
		proposedAction = "Source Carrier Rates & Prepare Quotation"
		expectedEffect = "Initiate carrier lane lookup and spot rate procurement to satisfy customer response target date."
	case ActionTypeReviewQuotation:
		proposedAction = "Review Quotation Commercial Terms"
		expectedEffect = "Review itemized charges, equipment specifications, and validity window before customer dispatch."
	case ActionTypeRequestPricingApproval:
		proposedAction = "Request Pricing & Margin Approval"
		expectedEffect = "Submit quotation valuation, cost structure, and gross margin percentage to commercial management via HITL approval system."
	case ActionTypeExtendQuotationValidity:
		proposedAction = "Extend Quotation Validity"
		expectedEffect = "Re-evaluate spot rate validity and extend quotation expiration window."
	case ActionTypeFollowupQuote:
		proposedAction = "Quotation Commercial Follow-Up"
		expectedEffect = "Transmit proactive customer communication to confirm offer receipt and assess booking conversion."
	case ActionTypeInvestigateException:
		proposedAction = "Investigate Active Shipment Exception"
		expectedEffect = "Initiate carrier / freight terminal investigation and coordinate resolution plan for flagged exception."
	case ActionTypeRequestCarrierClarification:
		proposedAction = "Request Carrier Milestone / Tracking Clarification"
		expectedEffect = "Query carrier operational desk for milestone verification and current schedule status."
	case ActionTypeSendCustomerShipmentUpdate:
		proposedAction = "Send Controlled Customer Shipment Update"
		expectedEffect = "Prepare verified status update for customer highlighting milestone progress without unverified delivery commitments."
	case ActionTypeUpdateMilestone:
		proposedAction = "Review & Update Milestone Schedule"
		expectedEffect = "Update milestone actual or planned timestamp based on verified carrier transport documents."
	case ActionTypeEscalateOperationalRisk:
		proposedAction = "Escalate High-Risk Operational Disruption"
		expectedEffect = "Route critical operational exception to operations management for expedited carrier intervention."
	case ActionTypeAssignOperationsOwner:
		proposedAction = "Assign Operations Shipment Owner"
		expectedEffect = "Assign authorized logistics operations coordinator to resolve missing operational details."
	case ActionTypeRequestMissingDocuments:
		proposedAction = "Request Missing Shipping Documents & Information"
		expectedEffect = "Request missing transport reference, container details, or documentation from shipper or carrier."
	case ActionTypeReviewOverdueInvoice:
		proposedAction = "Review Overdue Receivable Invoice"
		expectedEffect = "Examine invoice maturity, debtor aging band, and prior payment history to determine appropriate collections action."
	case ActionTypeContactCustomerCollections:
		proposedAction = "Contact Customer Accounts Payable"
		expectedEffect = "Initiate controlled receivables inquiry with debtor finance contact referencing verified outstanding balance and maturity date."
	case ActionTypeVerifyPaymentStatus:
		proposedAction = "Verify Remittance & Payment Status"
		expectedEffect = "Check bank statements and remittance advice to confirm if customer transaction is in transit or unallocated."
	case ActionTypeReviewInvoiceDispute:
		proposedAction = "Review Commercial Invoice Dispute"
		expectedEffect = "Examine reason for invoice dispute, coordinate with operations and sales teams, and determine billing adjustment."
	case ActionTypeAssignCollectionOwner:
		proposedAction = "Assign Accounts Receivable Collection Owner"
		expectedEffect = "Assign dedicated credit controller or account representative to pursue overdue balance settlement."
	case ActionTypeEscalateCollectionRisk:
		proposedAction = "Escalate Customer Credit Default Risk"
		expectedEffect = "Route high-exposure overdue account to finance management for credit limit hold and executive escalation."
	case ActionTypePreparePaymentReminder:
		proposedAction = "Prepare Controlled Payment Reminder"
		expectedEffect = "Generate editable reminder statement detailing approaching or overdue maturity dates without threatening language."
	case ActionTypeRequestFinanceReview:
		proposedAction = "Request Internal Finance & Ledger Review"
		expectedEffect = "Submit invoice data quality discrepancy, impossible balance, or missing due date to finance controller."
	case ActionTypeReconcilePaymentInfo:
		proposedAction = "Reconcile Invoice Settlement & Ledger Status"
		expectedEffect = "Audit invoice status versus payment records to resolve status/balance ledger contradiction."
	// Phase 2 Task 2.6 Action Previews
	case ActionTypeReviewExpiringContract:
		proposedAction = "Review Expired / Expiring Commercial Contract"
		expectedEffect = "Inspect contract terms, rate sheet validity, and commercial commitments before renewing or terminating agreement."
	case ActionTypeAssignContractOwner:
		proposedAction = "Assign Legal / Commercial Contract Owner"
		expectedEffect = "Designate responsible team member to manage contract renewal timelines and compliance obligations."
	case ActionTypeVerifyRenewalTerms:
		proposedAction = "Verify Renewal Terms & Notice Deadlines"
		expectedEffect = "Review counterparty notice period requirements and prepare commercial extension documentation."
	case ActionTypeRequestMissingDocument:
		proposedAction = "Request Missing Compliance Document"
		expectedEffect = "Prepare missing-document request to shipper or carrier specifying required regulatory credentials."
	case ActionTypeReviewComplianceChecklist:
		proposedAction = "Review Compliance Verification Checklist"
		expectedEffect = "Audit pending compliance items and verify certificate validity dates."
	case ActionTypeEscalateComplianceRisk:
		proposedAction = "Escalate Compliance Non-Conformance Risk"
		expectedEffect = "Notify compliance officer regarding unresolved exceptions or expired certificates."
	case ActionTypeReviewContractObligation:
		proposedAction = "Review Operational Contract Obligation"
		expectedEffect = "Designate operational lead and verify fulfillment of specific contractual commitments."
	case ActionTypePrepareRenewalReminder:
		proposedAction = "Prepare Contract Renewal Reminder"
		expectedEffect = "Generate editable renewal reminder notice detailing upcoming expiration date and next steps."
	case ActionTypeSubmitLegalApproval:
		proposedAction = "Submit Contract for Legal Approval"
		expectedEffect = "Route contract terms to legal team through central approval workflow."
	case ActionTypeVerifyDocumentMetadata:
		proposedAction = "Verify Document Extraction & Metadata"
		expectedEffect = "Review OCR extraction accuracy, resolve discrepancies, and confirm metadata in Document Workspace."
	}

	preview := &ActionPreview{
		RecommendationID:  rec.ID,
		ActionType:        rec.ActionType,
		ProposedAction:    proposedAction,
		SourceType:        rec.SourceType,
		SourceID:          rec.SourceID,
		SourceReference:   rec.SourceReference,
		ExpectedEffect:    expectedEffect,
		RiskLevel:         rec.RiskLevel,
		RequiresApproval:  rec.RequiresApproval,
		Evidence:          rec.Evidence,
		InitiatedByUser:   userLabel,
		CorrelationID:     correlationID,
		SuggestedNextStep: rec.RecommendedAction,
	}

	s.recordAudit(ctx, orgID, userID, userName, "ACTION_PREVIEW_GENERATED", fmt.Sprintf("%d", id), rec.Title, fmt.Sprintf("Generated controlled action preview for %s on %s #%d", rec.ActionType, rec.SourceType, rec.SourceID), correlationID, "SUCCESS", "")

	return preview, nil
}

func (s *service) RequestApproval(ctx context.Context, orgID int64, id int64, input RequestApprovalInput, userID int64, userName string, correlationID string) (*Recommendation, error) {
	rec, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	userLabel := userName
	if userLabel == "" {
		userLabel = fmt.Sprintf("user-%d", userID)
	}

	var approvalID int64
	if s.approvalsSvc != nil {
		desc := rec.Description
		if input.Notes != "" {
			desc = fmt.Sprintf("%s\n\nOperator Notes: %s", rec.Description, input.Notes)
		}

		appCategory := "OPERATIONS"
		appType := "Operations Approval"
		dept := "Logistics Operations"
		if rec.Category == CategoryPricing {
			appCategory = "COMMERCIAL"
			appType = "Commercial Approval"
			dept = "Commercial Pricing"
		} else if rec.Category == CategoryRFQ {
			appCategory = "COMMERCIAL"
			appType = "Commercial Approval"
			dept = "Commercial Pricing"
		} else if rec.Category == CategoryOperations || rec.Category == CategoryShipment || rec.Category == CategoryException {
			appCategory = "OPERATIONS"
			appType = "Operational Risk Approval"
			dept = "Logistics Operations"
		} else if rec.Category == CategoryFinance || rec.ActionType == ActionTypeEscalateCollectionRisk || rec.ActionType == ActionTypeReviewInvoiceDispute {
			appCategory = "FINANCE"
			appType = "Finance & Credit Risk Approval"
			dept = "Finance & Credit Control"
		} else if rec.Category == CategoryContract || rec.SourceType == SourceContract || rec.ActionType == ActionTypeSubmitLegalApproval || rec.ActionType == ActionTypeReviewExpiringContract {
			appCategory = "LEGAL"
			appType = "Contract & Renewal Approval"
			dept = "Legal & Contracts Desk"
		} else if rec.Category == CategoryCompliance || rec.SourceType == SourceCompliance || rec.SourceType == SourceDocument || rec.ActionType == ActionTypeEscalateComplianceRisk {
			appCategory = "COMPLIANCE"
			appType = "Trade Compliance Approval"
			dept = "Trade Compliance & Governance"
		}

		priority := "HIGH"
		if rec.Priority == PriorityCritical {
			priority = "URGENT"
		}

		cName := "Customer"
		if rec.CustomerName != nil && *rec.CustomerName != "" {
			cName = *rec.CustomerName
		}
		var cID int64
		if rec.CustomerID != nil {
			cID = *rec.CustomerID
		}

		approvalReq, err := s.approvalsSvc.CreateApproval(ctx, orgID, &approvals.CreateApprovalInput{
			Title:             fmt.Sprintf("Approval: %s", rec.Title),
			Category:          appCategory,
			Type:              appType,
			Priority:          priority,
			RelatedRef:        rec.SourceReference,
			RelatedEntityType: rec.SourceType,
			RelatedEntityID:   rec.SourceID,
			CustomerName:      cName,
			CustomerID:        cID,
			RequestedByName:   userLabel,
			Department:        dept,
			Description:       desc,
		}, userLabel)
		if err != nil {
			return nil, fmt.Errorf("failed to create approval request: %w", err)
		}
		if approvalReq != nil {
			approvalID = approvalReq.ID
			_ = s.repo.LinkApproval(ctx, id, orgID, approvalID)
		}
	}

	// Update status to assigned if currently new
	if rec.Status == StatusNew {
		_ = s.repo.UpdateStatus(ctx, id, orgID, StatusAssigned, &userID, nil)
	}

	updated, err := s.GetRecommendation(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, orgID, userID, userName, "APPROVAL_REQUESTED", fmt.Sprintf("%d", id), rec.Title, fmt.Sprintf("Submitted recommendation for operational approval (Approval Request #%d)", approvalID), correlationID, "SUCCESS", "")

	return updated, nil
}

func (s *service) GetEntityEvidence(ctx context.Context, orgID int64, entityType string, entityID int64) ([]EvidenceItem, error) {
	if orgID <= 0 || entityID <= 0 {
		return nil, errors.New("invalid org_id or entity_id")
	}

	entityType = strings.ToUpper(strings.TrimSpace(entityType))
	var filter RecommendationFilter
	filter.Limit = 20
	switch entityType {
	case "SHIPMENT":
		filter.ShipmentID = &entityID
	case "MILESTONE":
		filter.MilestoneID = &entityID
	case "EXCEPTION":
		filter.ExceptionID = &entityID
	case "RFQ":
		filter.RFQID = &entityID
	case "QUOTATION":
		filter.QuotationID = &entityID
	default:
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}

	recs, _, err := s.repo.List(ctx, orgID, filter)
	if err != nil {
		return nil, err
	}

	var evidenceList []EvidenceItem
	seen := make(map[string]bool)
	for _, rec := range recs {
		for _, ev := range rec.Evidence {
			key := fmt.Sprintf("%s:%s:%s", ev.SourceModule, ev.FieldName, ev.ObservedValue)
			if !seen[key] {
				seen[key] = true
				evidenceList = append(evidenceList, ev)
			}
		}
	}

	return evidenceList, nil
}

func (s *service) recordAudit(ctx context.Context, orgID int64, userID int64, userName string, action string, resID string, resName string, description string, correlationID string, result string, errMsg string) {
	if s.auditSvc == nil {
		return
	}

	var uid *int64
	if userID > 0 {
		uid = &userID
	}
	uName := userName
	if uName == "" {
		uName = "SYSTEM"
	}

	metadata := map[string]interface{}{
		"correlation_id": correlationID,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}

	params := auditDomain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorID:      uid,
		Action:       action,
		Module:       "RECOMMENDATIONS",
		ResourceType: "RECOMMENDATION",
		ResourceID:   resID,
		ResourceName: resName,
		Description:  description,
		ActorType:    "USER",
		ActorName:    uName,
		Result:       result,
		ErrorMessage: errMsg,
		Metadata:     metadata,
	}

	s.auditSvc.RecordAsync(ctx, params)
}

func (s *service) GetInvoiceEvidence(ctx context.Context, orgID int64, invoiceID int64) (*InvoiceEvidencePayload, error) {
	if orgID <= 0 || invoiceID <= 0 {
		return nil, fmt.Errorf("invalid organization or invoice id")
	}
	return s.repo.GetInvoiceEvidence(ctx, orgID, invoiceID)
}

func (s *service) GetCustomerCollectionSummary(ctx context.Context, orgID int64, customerID int64) (*CustomerCollectionSummaryPayload, error) {
	if orgID <= 0 || customerID <= 0 {
		return nil, fmt.Errorf("invalid organization or customer id")
	}
	return s.repo.GetCustomerCollectionSummary(ctx, orgID, customerID)
}

func (s *service) GetContractEvidence(ctx context.Context, orgID int64, contractID int64) (*ContractEvidencePayload, error) {
	if orgID <= 0 || contractID <= 0 {
		return nil, fmt.Errorf("invalid organization or contract id")
	}
	return s.repo.GetContractEvidence(ctx, orgID, contractID)
}

func (s *service) GetDocumentEvidence(ctx context.Context, orgID int64, docID string) (*DocumentEvidencePayload, error) {
	if orgID <= 0 || strings.TrimSpace(docID) == "" {
		return nil, fmt.Errorf("invalid organization or document id")
	}
	return s.repo.GetDocumentEvidence(ctx, orgID, docID)
}

func (s *service) GetComplianceEvidence(ctx context.Context, orgID int64, compID int64) (*ComplianceEvidencePayload, error) {
	if orgID <= 0 || compID <= 0 {
		return nil, fmt.Errorf("invalid organization or compliance id")
	}
	return s.repo.GetComplianceEvidence(ctx, orgID, compID)
}

func (s *service) GetContractComplianceSummary(ctx context.Context, orgID int64) (*ContractComplianceSummaryPayload, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization id")
	}
	return s.repo.GetContractComplianceSummary(ctx, orgID)
}
