package contract_compliance_automation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/freel/backend/internal/approvals"
	"github.com/freel/backend/internal/audit/domain"
	auditSvcPkg "github.com/freel/backend/internal/audit/service"
	"github.com/freel/backend/internal/contracts"
	"github.com/jmoiron/sqlx"
)

// Service defines the business interface for contract compliance automation
type Service interface {
	GetContractOverview(ctx context.Context, orgID, contractID int64) (*ContractOverviewDTO, error)
	ReviewContract(ctx context.Context, orgID, contractID int64) (*ContractComplianceReview, error)
	ExtractClauses(ctx context.Context, orgID, contractID int64) (json.RawMessage, error)
	VerifyStructuredTerms(ctx context.Context, orgID, contractID int64) (json.RawMessage, error)
	AssessCompliance(ctx context.Context, orgID, contractID int64) (json.RawMessage, error)
	GenerateClarificationDraft(ctx context.Context, orgID, contractID, userID int64, input GenerateClarificationDraftInput) (*ContractComplianceDraft, error)
	GetDraft(ctx context.Context, orgID, draftID int64) (*ContractComplianceDraft, error)
	ListDrafts(ctx context.Context, orgID, contractID int64) ([]*ContractComplianceDraft, error)
	UpdateDraft(ctx context.Context, orgID, draftID int64, input UpdateClarificationDraftInput) (*ContractComplianceDraft, error)
	SubmitDraftForApproval(ctx context.Context, orgID, draftID, userID int64, input SubmitDraftApprovalInput) (*ContractComplianceDraft, error)
}

type service struct {
	db             *sqlx.DB
	repo           Repository
	approvalsSvc   approvals.Service
	auditSvc       auditSvcPkg.Service
	sidecarBaseURL string
	httpClient     *http.Client
}

// NewService instantiates a new Service instance
func NewService(
	db *sqlx.DB,
	repo Repository,
	approvalsSvc approvals.Service,
	auditSvc auditSvcPkg.Service,
) Service {
	sidecarURL := os.Getenv("AI_SIDECAR_URL")
	if sidecarURL == "" {
		sidecarURL = "http://127.0.0.1:8090"
	}

	return &service{
		db:             db,
		repo:           repo,
		approvalsSvc:   approvalsSvc,
		auditSvc:       auditSvc,
		sidecarBaseURL: sidecarURL,
		httpClient:     &http.Client{Timeout: 45 * time.Second},
	}
}

// calculateDeterministicSignals performs authoritative compliance calculations in Go
func (s *service) calculateDeterministicSignals(
	c *contracts.Contract,
	terms []map[string]interface{},
	obligations []map[string]interface{},
	docs []map[string]interface{},
) DeterministicComplianceSignalsDTO {
	now := time.Now()
	dto := DeterministicComplianceSignalsDTO{
		DaysUntilExpiry: 999,
	}

	// 1. Expiry & Term Evaluation
	if c.ExpiryDate == nil || *c.ExpiryDate == "" {
		dto.MissingExpiryDate = true
	} else {
		expiryTime, err := time.Parse("2006-01-02", *c.ExpiryDate)
		if err == nil {
			if now.After(expiryTime) {
				dto.IsExpired = true
				dto.DaysUntilExpiry = 0
			} else {
				days := int(expiryTime.Sub(now).Hours() / 24)
				dto.DaysUntilExpiry = days
				if days <= 60 {
					dto.IsNearingExpiry = true
				}
			}
		}
	}

	if c.EffectiveDate == nil || *c.EffectiveDate == "" {
		dto.MissingEffectiveDate = true
	}

	// 2. Required Document Checking
	var missingDocs []string
	hasInsuranceCert := false
	for _, t := range terms {
		tKey, _ := t["term_key"].(string)
		if strings.Contains(strings.ToUpper(tKey), "INSURANCE") {
			hasInsuranceCert = true
		}
	}
	if !hasInsuranceCert {
		dto.MissingInsuranceTerms = true
		missingDocs = append(missingDocs, "Certificate of Cargo Liability Insurance (COI)")
	}

	if len(docs) == 0 {
		missingDocs = append(missingDocs, "Signed Master Agreement Document (PDF)")
	}
	dto.RequiredDocumentsMissing = missingDocs
	dto.MissingRequiredDocuments = len(missingDocs) > 0

	// 3. Document Review Status
	for _, d := range docs {
		st, _ := d["status"].(string)
		if st == "PENDING_REVIEW" {
			dto.HasDocumentAwaitingReview = true
		} else if st == "FAILED" || st == "REJECTED" {
			dto.HasRejectedDocument = true
		}
	}

	// 4. Rate & SLA Information
	hasRateTerm := false
	hasSLATerm := false
	for _, t := range terms {
		cat, _ := t["term_category"].(string)
		if strings.ToUpper(cat) == "RATE" || strings.ToUpper(cat) == "PRICING" {
			hasRateTerm = true
		}
		if strings.ToUpper(cat) == "SLA" || strings.ToUpper(cat) == "SERVICE_LEVEL" {
			hasSLATerm = true
		}
	}
	dto.MissingRateInformation = !hasRateTerm
	dto.MissingServiceLevelTerms = !hasSLATerm

	// 5. Inactive agreement link check
	if c.Status == "EXPIRED" || c.Status == "TERMINATED" {
		dto.IsLinkedToInactiveAgreement = true
	}

	// 6. Manual review requirement threshold
	if dto.IsExpired || dto.IsNearingExpiry || dto.MissingRequiredDocuments || dto.HasRejectedDocument {
		dto.RequiresManualComplianceReview = true
	}

	return dto
}

func (s *service) buildContractContextPayload(
	ctx context.Context,
	orgID, contractID int64,
	c *contracts.Contract,
	corrID string,
) (map[string]interface{}, DeterministicComplianceSignalsDTO, error) {
	terms, _ := s.repo.GetContractTerms(ctx, orgID, contractID)
	obligations, _ := s.repo.GetContractObligations(ctx, orgID, contractID)
	reqs, _ := s.repo.GetContractComplianceRequirements(ctx, orgID, contractID)
	docs, _ := s.repo.GetContractDocuments(ctx, orgID, contractID)

	signals := s.calculateDeterministicSignals(c, terms, obligations, docs)

	cVal := 0.0
	if c.ContractValue != nil {
		cVal = *c.ContractValue
	}
	curr := "USD"
	if c.Currency != nil {
		curr = *c.Currency
	}
	mode := "MULTIMODAL"
	if c.TransportMode != nil {
		mode = *c.TransportMode
	}
	owner := "Commercial Operations"
	if c.Owner != nil {
		owner = *c.Owner
	}

	var primaryDocID *string
	var primaryDocName *string
	if len(docs) > 0 {
		if idStr, ok := docs[0]["id"].(string); ok {
			primaryDocID = &idStr
		}
		if fnStr, ok := docs[0]["file_name"].(string); ok {
			primaryDocName = &fnStr
		}
	}

	ctxMap := map[string]interface{}{
		"org_id":                  orgID,
		"contract_id":             contractID,
		"contract_reference":      c.ContractReference,
		"contract_name":           c.ContractName,
		"contract_type":           c.ContractType,
		"party_id":                c.PartyID,
		"party_name":              c.PartyName,
		"status":                  string(c.Status),
		"effective_date":          c.EffectiveDate,
		"expiry_date":             c.ExpiryDate,
		"contract_value":          cVal,
		"currency":                curr,
		"transport_mode":          mode,
		"owner":                   owner,
		"document_id":             primaryDocID,
		"document_name":           primaryDocName,
		"document_text":           "Master Service Agreement terms governing ocean freight, detention schedules, and cargo liability.",
		"structured_terms":        terms,
		"structured_obligations":   obligations,
		"compliance_requirements": reqs,
		"linked_documents":        docs,
		"correlation_id":          corrID,
	}

	return ctxMap, signals, nil
}

func (s *service) GetContractOverview(ctx context.Context, orgID, contractID int64) (*ContractOverviewDTO, error) {
	c, err := s.repo.GetContractByID(ctx, orgID, contractID)
	if err != nil || c == nil {
		return nil, fmt.Errorf("contract not found or access denied: %w", err)
	}

	terms, _ := s.repo.GetContractTerms(ctx, orgID, contractID)
	obligations, _ := s.repo.GetContractObligations(ctx, orgID, contractID)
	docs, _ := s.repo.GetContractDocuments(ctx, orgID, contractID)
	signals := s.calculateDeterministicSignals(c, terms, obligations, docs)

	latestReview, _ := s.repo.GetLatestReview(ctx, orgID, contractID)
	drafts, _ := s.repo.ListDrafts(ctx, orgID, contractID)

	cVal := 0.0
	if c.ContractValue != nil {
		cVal = *c.ContractValue
	}
	curr := "USD"
	if c.Currency != nil {
		curr = *c.Currency
	}
	mode := "MULTIMODAL"
	if c.TransportMode != nil {
		mode = *c.TransportMode
	}
	owner := "Commercial Operations"
	if c.Owner != nil {
		owner = *c.Owner
	}

	dto := &ContractOverviewDTO{
		ContractID:           contractID,
		OrgID:                orgID,
		ContractReference:    c.ContractReference,
		ContractName:         c.ContractName,
		ContractType:         c.ContractType,
		PartyName:            c.PartyName,
		Status:               string(c.Status),
		EffectiveDate:        c.EffectiveDate,
		ExpiryDate:           c.ExpiryDate,
		ContractValue:        cVal,
		Currency:             curr,
		TransportMode:        mode,
		Owner:                owner,
		DeterministicSignals: signals,
		LatestReview:         latestReview,
		ExistingDrafts:       drafts,
		LinkedDocumentsCount: len(docs),
		StructuredTermsCount: len(terms),
	}

	return dto, nil
}

func (s *service) ReviewContract(ctx context.Context, orgID, contractID int64) (*ContractComplianceReview, error) {
	c, err := s.repo.GetContractByID(ctx, orgID, contractID)
	if err != nil || c == nil {
		return nil, fmt.Errorf("contract not found or access denied: %w", err)
	}

	corrID := fmt.Sprintf("corr-contract-rev-%d-%d", contractID, time.Now().UnixNano())
	ctxMap, signals, err := s.buildContractContextPayload(ctx, orgID, contractID, c, corrID)
	if err != nil {
		return nil, err
	}

	reqBody := map[string]interface{}{
		"context":        ctxMap,
		"signals":        signals,
		"correlation_id": corrID,
	}
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sidecar review request: %w", err)
	}

	url := s.sidecarBaseURL + "/contract-compliance/review-contract"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LogisticsHQ-Service-Key", "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sidecar review request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar returned non-200 status: %d", resp.StatusCode)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode sidecar review response: %w", err)
	}

	// Map to ContractComplianceReview struct
	riskLevel, _ := raw["risk_level"].(string)
	if riskLevel == "" {
		riskLevel = "MEDIUM"
	}
	riskScore := 25.0
	if rs, ok := raw["risk_score"].(float64); ok {
		riskScore = rs
	}
	compStatus, _ := raw["compliance_status"].(string)
	if compStatus == "" {
		compStatus = ComplianceStatusReviewRequired
	}
	execSummary, _ := raw["executive_summary"].(string)
	confScore := 0.90
	if cs, ok := raw["confidence_score"].(float64); ok {
		confScore = cs
	}

	signalsJSON, _ := json.Marshal(signals)
	clausesJSON, _ := json.Marshal(raw["extracted_clauses"])
	discrepanciesJSON, _ := json.Marshal(raw["structured_discrepancies"])
	obligationsJSON, _ := json.Marshal(raw["compliance_obligations"])
	missingJSON, _ := json.Marshal(raw["missing_information"])
	recommendationsJSON, _ := json.Marshal(raw["recommendations"])
	evidenceJSON, _ := json.Marshal(raw["evidence"])

	var docIDPtr *string
	if docIDStr, ok := raw["document_id"].(string); ok && docIDStr != "" {
		docIDPtr = &docIDStr
	}

	review := &ContractComplianceReview{
		OrgID:                  orgID,
		ContractID:             contractID,
		DocumentID:             docIDPtr,
		RiskLevel:              riskLevel,
		RiskScore:              riskScore,
		ComplianceStatus:       compStatus,
		ExecutiveSummary:       execSummary,
		DeterministicSignals:   signalsJSON,
		ExtractedClauses:       clausesJSON,
		StructuredDiscrepancies: discrepanciesJSON,
		ComplianceObligations:  obligationsJSON,
		MissingInformation:     missingJSON,
		Recommendations:        recommendationsJSON,
		Evidence:               evidenceJSON,
		ConfidenceScore:        confScore,
		CorrelationID:          corrID,
	}

	saved, err := s.repo.SaveReview(ctx, review)
	if err != nil {
		return nil, fmt.Errorf("failed to persist compliance review: %w", err)
	}

	// Audit log
	if s.auditSvc != nil {
		s.auditSvc.RecordAsync(ctx, domain.CreateAuditLogParams{
			OrgID:        orgID,
			ActorType:    domain.ActorTypeSystem,
			Action:       domain.ActionCreate,
			Module:       domain.ModuleContracts,
			ResourceType: "ai_contract_compliance_reviews",
			ResourceID:   fmt.Sprintf("%d", saved.ID),
			Description:  fmt.Sprintf("Completed compliance review for contract %s. Risk Level: %s, Score: %.1f", c.ContractReference, riskLevel, riskScore),
		})
	}

	return saved, nil
}

func (s *service) ExtractClauses(ctx context.Context, orgID, contractID int64) (json.RawMessage, error) {
	c, err := s.repo.GetContractByID(ctx, orgID, contractID)
	if err != nil || c == nil {
		return nil, fmt.Errorf("contract not found: %w", err)
	}

	corrID := fmt.Sprintf("corr-clauses-%d-%d", contractID, time.Now().UnixNano())
	ctxMap, signals, err := s.buildContractContextPayload(ctx, orgID, contractID, c, corrID)
	if err != nil {
		return nil, err
	}

	reqBody := map[string]interface{}{
		"context":        ctxMap,
		"signals":        signals,
		"correlation_id": corrID,
	}
	jsonBytes, _ := json.Marshal(reqBody)

	url := s.sidecarBaseURL + "/contract-compliance/extract-clauses"
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LogisticsHQ-Service-Key", "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *service) VerifyStructuredTerms(ctx context.Context, orgID, contractID int64) (json.RawMessage, error) {
	c, err := s.repo.GetContractByID(ctx, orgID, contractID)
	if err != nil || c == nil {
		return nil, fmt.Errorf("contract not found: %w", err)
	}

	corrID := fmt.Sprintf("corr-verify-terms-%d-%d", contractID, time.Now().UnixNano())
	ctxMap, signals, err := s.buildContractContextPayload(ctx, orgID, contractID, c, corrID)
	if err != nil {
		return nil, err
	}

	reqBody := map[string]interface{}{
		"context":        ctxMap,
		"signals":        signals,
		"correlation_id": corrID,
	}
	jsonBytes, _ := json.Marshal(reqBody)

	url := s.sidecarBaseURL + "/contract-compliance/verify-structured-terms"
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LogisticsHQ-Service-Key", "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *service) AssessCompliance(ctx context.Context, orgID, contractID int64) (json.RawMessage, error) {
	c, err := s.repo.GetContractByID(ctx, orgID, contractID)
	if err != nil || c == nil {
		return nil, fmt.Errorf("contract not found: %w", err)
	}

	corrID := fmt.Sprintf("corr-compliance-%d-%d", contractID, time.Now().UnixNano())
	ctxMap, signals, err := s.buildContractContextPayload(ctx, orgID, contractID, c, corrID)
	if err != nil {
		return nil, err
	}

	reqBody := map[string]interface{}{
		"context":        ctxMap,
		"signals":        signals,
		"correlation_id": corrID,
	}
	jsonBytes, _ := json.Marshal(reqBody)

	url := s.sidecarBaseURL + "/contract-compliance/assess-compliance"
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LogisticsHQ-Service-Key", "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (s *service) GenerateClarificationDraft(
	ctx context.Context,
	orgID, contractID, userID int64,
	input GenerateClarificationDraftInput,
) (*ContractComplianceDraft, error) {
	c, err := s.repo.GetContractByID(ctx, orgID, contractID)
	if err != nil || c == nil {
		return nil, fmt.Errorf("contract not found: %w", err)
	}

	corrID := fmt.Sprintf("corr-draft-%d-%d", contractID, time.Now().UnixNano())
	ctxMap, signals, err := s.buildContractContextPayload(ctx, orgID, contractID, c, corrID)
	if err != nil {
		return nil, err
	}

	draftType := input.DraftType
	if draftType == "" {
		draftType = "CLAUSE_CLARIFICATION"
	}
	tone := input.Tone
	if tone == "" {
		tone = "PROFESSIONAL"
	}

	reqBody := map[string]interface{}{
		"context":             ctxMap,
		"signals":             signals,
		"draft_type":          draftType,
		"tone":                tone,
		"custom_instructions": input.CustomInstructions,
		"target_clause_id":    input.TargetClauseID,
		"correlation_id":      corrID,
	}
	jsonBytes, _ := json.Marshal(reqBody)

	url := s.sidecarBaseURL + "/contract-compliance/generate-clarification-draft"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LogisticsHQ-Service-Key", "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sidecar draft request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar draft returned non-200 status: %d", resp.StatusCode)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	subject, _ := raw["subject"].(string)
	messageBody, _ := raw["message_body"].(string)
	internalNotes, _ := raw["internal_review_notes"].(string)

	recipientName := input.RecipientName
	if recipientName == "" {
		recipientName, _ = raw["suggested_recipient_name"].(string)
		if recipientName == "" {
			recipientName = fmt.Sprintf("%s Contracts Lead", c.PartyName)
		}
	}

	recipientEmail := input.RecipientEmail
	if recipientEmail == "" {
		recipientEmail, _ = raw["suggested_recipient_email"].(string)
		if recipientEmail == "" {
			recipientEmail = fmt.Sprintf("contracts@%s.com", strings.ToLower(strings.ReplaceAll(c.PartyName, " ", "")))
		}
	}

	draft := &ContractComplianceDraft{
		OrgID:            orgID,
		ContractID:       contractID,
		DraftType:        draftType,
		Subject:          subject,
		MessageBody:      messageBody,
		InternalNotes:    &internalNotes,
		RecipientName:    recipientName,
		RecipientEmail:   recipientEmail,
		Status:           DraftStatusDraft,
		RequiresApproval: true,
		CreatedByUserID:  &userID,
		CorrelationID:    corrID,
	}

	saved, err := s.repo.CreateDraft(ctx, draft)
	if err != nil {
		return nil, fmt.Errorf("failed to persist compliance draft: %w", err)
	}

	return saved, nil
}

func (s *service) GetDraft(ctx context.Context, orgID, draftID int64) (*ContractComplianceDraft, error) {
	return s.repo.GetDraft(ctx, orgID, draftID)
}

func (s *service) ListDrafts(ctx context.Context, orgID, contractID int64) ([]*ContractComplianceDraft, error) {
	return s.repo.ListDrafts(ctx, orgID, contractID)
}

func (s *service) UpdateDraft(ctx context.Context, orgID, draftID int64, input UpdateClarificationDraftInput) (*ContractComplianceDraft, error) {
	draft, err := s.repo.GetDraft(ctx, orgID, draftID)
	if err != nil || draft == nil {
		return nil, fmt.Errorf("draft not found: %w", err)
	}

	if draft.Status == DraftStatusApproved || draft.Status == DraftStatusDispatched {
		return nil, fmt.Errorf("cannot modify draft in %s status", draft.Status)
	}

	if input.Subject != "" {
		draft.Subject = input.Subject
	}
	if input.MessageBody != "" {
		draft.MessageBody = input.MessageBody
	}
	if input.RecipientName != "" {
		draft.RecipientName = input.RecipientName
	}
	if input.RecipientEmail != "" {
		draft.RecipientEmail = input.RecipientEmail
	}
	if input.InternalNotes != nil {
		draft.InternalNotes = input.InternalNotes
	}

	return s.repo.UpdateDraft(ctx, draft)
}

func (s *service) SubmitDraftForApproval(
	ctx context.Context,
	orgID, draftID, userID int64,
	input SubmitDraftApprovalInput,
) (*ContractComplianceDraft, error) {
	draft, err := s.repo.GetDraft(ctx, orgID, draftID)
	if err != nil || draft == nil {
		return nil, fmt.Errorf("draft not found: %w", err)
	}

	if draft.Status != DraftStatusDraft {
		return nil, fmt.Errorf("draft cannot be submitted for approval in %s status", draft.Status)
	}

	c, _ := s.repo.GetContractByID(ctx, orgID, draft.ContractID)
	cRef := fmt.Sprintf("CTR-%d", draft.ContractID)
	if c != nil {
		cRef = c.ContractReference
	}

	// Create Managerial Approval Request via Centralized Approvals Center
	var approvalID int64 = 0
	if s.approvalsSvc != nil {
		payloadMap := map[string]interface{}{
			"contract_id":     draft.ContractID,
			"draft_id":        draft.ID,
			"draft_type":      draft.DraftType,
			"recipient_email": draft.RecipientEmail,
			"recipient_name":  draft.RecipientName,
			"subject":         draft.Subject,
			"message_body":    draft.MessageBody,
		}
		payloadBytes, _ := json.Marshal(payloadMap)

		actionName := "contracts.request_term_clarification"
		if draft.DraftType == "MISSING_DOCUMENT_REQUEST" {
			actionName = "contracts.request_missing_document"
		}

		appReq, appErr := s.approvalsSvc.CreateApproval(ctx, orgID, &approvals.CreateApprovalInput{
			Title:              fmt.Sprintf("Compliance Communication Approval: %s (%s)", draft.Subject, cRef),
			Category:           "CONTRACTS",
			Type:               "EXTERNAL_COMMUNICATION",
			Priority:           "HIGH",
			RelatedEntityType:  "CONTRACT",
			RelatedEntityID:    draft.ContractID,
			RelatedRef:         cRef,
			CustomerName:       draft.RecipientName,
			RequestedByID:     userID,
			Description:        input.Notes,
			ActionName:         actionName,
			RiskLevel:          "HIGH",
			RequiredPermission: "contracts:write",
			ProposedPayload:    string(payloadBytes),
			CorrelationID:      draft.CorrelationID,
			ActorType:          "HUMAN",
		}, "Compliance Specialist")

		if appErr == nil && appReq != nil {
			approvalID = appReq.ID
		}
	}

	if err := s.repo.UpdateDraftApproval(ctx, orgID, draftID, approvalID, DraftStatusPendingApproval); err != nil {
		return nil, fmt.Errorf("failed to update draft approval state: %w", err)
	}

	draft.ApprovalID = &approvalID
	draft.Status = DraftStatusPendingApproval
	return draft, nil
}
