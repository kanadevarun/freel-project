package collections_automation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/freel/backend/internal/approvals"
	"github.com/freel/backend/internal/audit/domain"
	auditSvcPkg "github.com/freel/backend/internal/audit/service"
	"github.com/freel/backend/internal/invoices"
	"github.com/jmoiron/sqlx"
)

// Service defines the business interface for collections automation
type Service interface {
	GetCollectionsOverview(ctx context.Context, orgID, invoiceID int64) (*FinanceCollectionsOverview, error)
	AnalyzeReceivablesRisk(ctx context.Context, orgID, invoiceID int64) (*ReceivablesAnalysis, error)
	PrioritizeCollections(ctx context.Context, orgID, invoiceID int64) (json.RawMessage, error)
	AnalyzeCustomerPaymentBehavior(ctx context.Context, orgID, invoiceID int64) (json.RawMessage, error)
	GetOperationalRecommendations(ctx context.Context, orgID, invoiceID int64) (json.RawMessage, error)
	GenerateCollectionDraft(ctx context.Context, orgID, invoiceID, userID int64, input GenerateCollectionDraftInput) (*CollectionDraft, error)
	GetDraft(ctx context.Context, orgID, draftID int64) (*CollectionDraft, error)
	ListDrafts(ctx context.Context, orgID, invoiceID int64) ([]*CollectionDraft, error)
	UpdateDraft(ctx context.Context, orgID, draftID int64, input UpdateCollectionDraftInput) (*CollectionDraft, error)
	SubmitDraftForApproval(ctx context.Context, orgID, draftID, userID int64, input SubmitCollectionDraftApprovalInput) (*CollectionDraft, error)
}

type service struct {
	db             *sqlx.DB
	repo           Repository
	invoiceRepo    invoices.Repository
	approvalsSvc   approvals.Service
	auditSvc       auditSvcPkg.Service
	sidecarBaseURL string
	httpClient     *http.Client
}

// NewService instantiates a new Service instance
func NewService(
	db *sqlx.DB,
	repo Repository,
	invoiceRepo invoices.Repository,
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
		invoiceRepo:    invoiceRepo,
		approvalsSvc:   approvalsSvc,
		auditSvc:       auditSvc,
		sidecarBaseURL: sidecarURL,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}

func getInternalServiceToken() string {
	tok := os.Getenv("AI_SIDECAR_SERVICE_KEY")
	if tok == "" {
		tok = os.Getenv("INTERNAL_SERVICE_KEY")
	}
	if tok == "" {
		tok = "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456"
	}
	return tok
}

func (s *service) calculateDeterministicFinanceSignals(
	ctx context.Context,
	orgID int64,
	inv *invoices.Invoice,
	custOutstanding float64,
	custOverdueCount int,
	custOverdueBal float64,
	creditLimit float64,
) DeterministicFinanceSignalsDTO {
	now := time.Now()

	// Days overdue calculation
	daysOverdue := 0
	isOverdue := false
	if !inv.DueDate.IsZero() && now.After(inv.DueDate) && inv.BalanceDue > 0 {
		isOverdue = true
		daysOverdue = int(now.Sub(inv.DueDate).Hours() / 24.0)
	}

	// Aging bucket
	agingBucket := "CURRENT"
	if isOverdue {
		if daysOverdue <= 30 {
			agingBucket = "1_30"
		} else if daysOverdue <= 60 {
			agingBucket = "31_60"
		} else if daysOverdue <= 90 {
			agingBucket = "61_90"
		} else {
			agingBucket = "90_PLUS"
		}
	}

	// Due-date proximity
	isApproaching := false
	daysUntilDue := 0
	if !isOverdue && !inv.DueDate.IsZero() && inv.BalanceDue > 0 {
		hoursUntilDue := inv.DueDate.Sub(now).Hours()
		if hoursUntilDue > 0 && hoursUntilDue <= 7*24.0 {
			isApproaching = true
			daysUntilDue = int(hoursUntilDue / 24.0)
		}
	}

	// High-value threshold ($10,000 USD equivalent or configured)
	isHighValue := inv.BalanceDue >= 10000.0

	// Partial payment
	isPartiallyPaid := inv.PaidAmount > 0 && inv.BalanceDue > 0

	// Customer multi-overdue
	hasMultipleOverdue := custOverdueCount > 1

	// Credit limit breach
	isCreditExceeded := creditLimit > 0 && custOutstanding > creditLimit

	// Dispute status
	isDisputed := strings.EqualFold(inv.Status, "Disputed")

	// Composite deterministic risk score (0 - 100)
	score := 5.0
	if isOverdue {
		score += math.Min(40.0, float64(daysOverdue)*0.8)
		if daysOverdue > 60 {
			score += 15.0
		}
	}
	if isHighValue {
		score += 15.0
	}
	if hasMultipleOverdue {
		score += 15.0
	}
	if isCreditExceeded {
		score += 15.0
	}
	if isDisputed {
		score += 20.0
	}
	if isPartiallyPaid && !isOverdue {
		score = math.Max(5.0, score-10.0) // partial payment within terms indicates willingness to pay
	}

	score = math.Min(100.0, math.Max(0.0, score))

	riskLevel := "LOW"
	if score >= 70.0 {
		riskLevel = "CRITICAL"
	} else if score >= 45.0 {
		riskLevel = "HIGH"
	} else if score >= 20.0 {
		riskLevel = "MEDIUM"
	}

	return DeterministicFinanceSignalsDTO{
		IsOverdue:                   isOverdue,
		DaysOverdue:                 daysOverdue,
		AgingBucket:                 agingBucket,
		IsApproachingDueDate:        isApproaching,
		DaysUntilDue:                daysUntilDue,
		IsHighValue:                 isHighValue,
		IsPartiallyPaid:             isPartiallyPaid,
		HasMultipleOverdue:          hasMultipleOverdue,
		CustomerOverdueCount:        custOverdueCount,
		CustomerTotalOverdueBalance: custOverdueBal,
		IsCreditLimitExceeded:       isCreditExceeded,
		IsDisputed:                  isDisputed,
		HasLinkedException:          false,
		RiskScore:                   math.Round(score*10) / 10,
		RiskLevel:                   riskLevel,
		RequiresApproval:            true,
	}
}

func (s *service) buildInvoiceContextPayload(
	ctx context.Context,
	orgID, invoiceID int64,
	inv *invoices.Invoice,
	corrID string,
) (map[string]interface{}, DeterministicFinanceSignalsDTO, error) {
	// Query aggregate customer statistics
	custOutstanding, _ := s.repo.GetCustomerTotalOutstanding(ctx, orgID, inv.CustomerID)
	custOverdueCount, custOverdueBal, _ := s.repo.GetCustomerOverdueInvoices(ctx, orgID, inv.CustomerID)
	creditLimit, _ := s.repo.GetCustomerCreditLimit(ctx, orgID, inv.CustomerID)

	signals := s.calculateDeterministicFinanceSignals(
		ctx, orgID, inv, custOutstanding, custOverdueCount, custOverdueBal, creditLimit,
	)

	// Fetch line items and payments
	lineItems, _ := s.invoiceRepo.GetInvoiceItems(ctx, orgID, invoiceID)
	payments, _ := s.invoiceRepo.GetInvoicePayments(ctx, orgID, invoiceID)

	var itemsList []map[string]interface{}
	for _, item := range lineItems {
		itemsList = append(itemsList, map[string]interface{}{
			"id":               item.ID,
			"description":      item.Description,
			"service_category": item.ServiceCategory,
			"quantity":         item.Quantity,
			"unit_price":       item.UnitPrice,
			"total_amount":     item.TotalAmount,
		})
	}
	if itemsList == nil {
		itemsList = []map[string]interface{}{}
	}

	var paymentsList []map[string]interface{}
	for _, p := range payments {
		pDateStr := ""
		if !p.PaymentDate.IsZero() {
			pDateStr = p.PaymentDate.Format("2006-01-02")
		}
		paymentsList = append(paymentsList, map[string]interface{}{
			"id":             p.ID,
			"payment_ref":    p.PaymentRef,
			"amount":         p.Amount,
			"payment_method": p.PaymentMethod,
			"status":         p.Status,
			"payment_date":   pDateStr,
		})
	}
	if paymentsList == nil {
		paymentsList = []map[string]interface{}{}
	}

	invDateStr := ""
	if !inv.InvoiceDate.IsZero() {
		invDateStr = inv.InvoiceDate.Format("2006-01-02")
	}
	dueDateStr := ""
	if !inv.DueDate.IsZero() {
		dueDateStr = inv.DueDate.Format("2006-01-02")
	}

	ctxMap := map[string]interface{}{
		"org_id":                       orgID,
		"invoice_id":                   invoiceID,
		"invoice_number":               inv.InvoiceNumber,
		"customer_id":                  inv.CustomerID,
		"customer_name":                inv.CustomerName,
		"customer_email":               "ap@" + strings.ToLower(strings.ReplaceAll(inv.CustomerName, " ", "")) + ".com",
		"customer_country":             inv.CustomerCountry,
		"shipment_id":                  inv.ShipmentID,
		"shipment_number":              inv.ShipmentNumber,
		"booking_id":                   inv.BookingID,
		"booking_number":               inv.BookingNumber,
		"invoice_date":                 invDateStr,
		"due_date":                     dueDateStr,
		"currency":                     inv.Currency,
		"total_amount":                 inv.TotalAmount,
		"paid_amount":                  inv.PaidAmount,
		"balance_due":                  inv.BalanceDue,
		"status":                       inv.Status,
		"line_items":                   itemsList,
		"payment_history":              paymentsList,
		"customer_credit_limit":        creditLimit,
		"customer_payment_terms":       "Net 30",
		"customer_open_invoices_count": custOverdueCount + 1,
		"customer_total_outstanding":   custOutstanding,
		"correlation_id":               corrID,
	}

	return ctxMap, signals, nil
}

func (s *service) GetCollectionsOverview(ctx context.Context, orgID, invoiceID int64) (*FinanceCollectionsOverview, error) {
	inv, err := s.invoiceRepo.GetInvoiceByID(ctx, orgID, invoiceID)
	if err != nil || inv == nil {
		return nil, fmt.Errorf("invoice not found or access denied: %w", err)
	}

	corrID := fmt.Sprintf("corr-fin-overview-%d-%d", invoiceID, time.Now().Unix())
	ctxMap, signals, err := s.buildInvoiceContextPayload(ctx, orgID, invoiceID, inv, corrID)
	if err != nil {
		return nil, err
	}

	analysis, _ := s.repo.GetLatestAnalysis(ctx, orgID, invoiceID)
	latestDraft, _ := s.repo.GetLatestDraft(ctx, orgID, invoiceID)

	var prioritizedItems json.RawMessage
	var recsRaw json.RawMessage

	reqPayload := map[string]interface{}{
		"context": ctxMap,
		"signals": signals,
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	// Call /finance-ops/prioritize-collections
	urlPrio := fmt.Sprintf("%s/finance-ops/prioritize-collections", s.sidecarBaseURL)
	if hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, urlPrio, bytes.NewReader(bodyBytes)); err == nil {
		hReq.Header.Set("Content-Type", "application/json")
		hReq.Header.Set("X-LogisticsHQ-Service-Key", getInternalServiceToken())
		if hResp, err := s.httpClient.Do(hReq); err == nil && hResp.StatusCode == http.StatusOK {
			defer hResp.Body.Close()
			var parsed map[string]interface{}
			if err := json.NewDecoder(hResp.Body).Decode(&parsed); err == nil {
				if items, ok := parsed["prioritized_items"]; ok {
					prioritizedItems, _ = json.Marshal(items)
				}
			}
		}
	}

	// Call /finance-ops/recommend-actions
	urlRec := fmt.Sprintf("%s/finance-ops/recommend-actions", s.sidecarBaseURL)
	if hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, urlRec, bytes.NewReader(bodyBytes)); err == nil {
		hReq.Header.Set("Content-Type", "application/json")
		hReq.Header.Set("X-LogisticsHQ-Service-Key", getInternalServiceToken())
		if hResp, err := s.httpClient.Do(hReq); err == nil && hResp.StatusCode == http.StatusOK {
			defer hResp.Body.Close()
			var parsed map[string]interface{}
			if err := json.NewDecoder(hResp.Body).Decode(&parsed); err == nil {
				if items, ok := parsed["recommendations"]; ok {
					recsRaw, _ = json.Marshal(items)
				}
			}
		}
	}

	pendingApproval := latestDraft != nil && latestDraft.Status == "PENDING_APPROVAL"
	var approvalID *int64
	if latestDraft != nil {
		approvalID = latestDraft.ApprovalID
	}

	invDateStr := ""
	if !inv.InvoiceDate.IsZero() {
		invDateStr = inv.InvoiceDate.Format("2006-01-02")
	}
	dueDateStr := ""
	if !inv.DueDate.IsZero() {
		dueDateStr = inv.DueDate.Format("2006-01-02")
	}

	custEmail := "ap@" + strings.ToLower(strings.ReplaceAll(inv.CustomerName, " ", "")) + ".com"

	return &FinanceCollectionsOverview{
		InvoiceID:        invoiceID,
		OrgID:            orgID,
		InvoiceNumber:    inv.InvoiceNumber,
		CustomerID:       inv.CustomerID,
		CustomerName:     inv.CustomerName,
		CustomerEmail:    custEmail,
		InvoiceDate:      invDateStr,
		DueDate:          dueDateStr,
		Currency:         inv.Currency,
		TotalAmount:      inv.TotalAmount,
		PaidAmount:       inv.PaidAmount,
		BalanceDue:       inv.BalanceDue,
		Status:           inv.Status,
		ShipmentID:       inv.ShipmentID,
		ShipmentNumber:   inv.ShipmentNumber,
		BookingNumber:    inv.BookingNumber,
		Signals:          signals,
		LatestAnalysis:   analysis,
		LatestDraft:      latestDraft,
		PrioritizedItems: prioritizedItems,
		Recommendations:  recsRaw,
		PendingApproval:  pendingApproval,
		ApprovalID:       approvalID,
		CorrelationID:    corrID,
	}, nil
}

func (s *service) AnalyzeReceivablesRisk(ctx context.Context, orgID, invoiceID int64) (*ReceivablesAnalysis, error) {
	inv, err := s.invoiceRepo.GetInvoiceByID(ctx, orgID, invoiceID)
	if err != nil || inv == nil {
		return nil, fmt.Errorf("invoice not found or access denied: %w", err)
	}

	corrID := fmt.Sprintf("corr-fin-risk-%d-%d", invoiceID, time.Now().Unix())
	ctxMap, signals, err := s.buildInvoiceContextPayload(ctx, orgID, invoiceID, inv, corrID)
	if err != nil {
		return nil, err
	}

	reqPayload := map[string]interface{}{
		"context": ctxMap,
		"signals": signals,
	}
	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode risk analysis payload: %w", err)
	}

	url := fmt.Sprintf("%s/finance-ops/analyze-receivables", s.sidecarBaseURL)
	hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create sidecar risk request: %w", err)
	}
	hReq.Header.Set("Content-Type", "application/json")
	hReq.Header.Set("X-LogisticsHQ-Service-Key", getInternalServiceToken())

	hResp, err := s.httpClient.Do(hReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar receivables analysis failed: %w", err)
	}
	defer hResp.Body.Close()

	if hResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar receivables analysis returned status %d", hResp.StatusCode)
	}

	var sidecarRes struct {
		RiskLevel            string        `json:"risk_level"`
		RiskScore            float64       `json:"risk_score"`
		ReceivablesSummary   string        `json:"receivables_summary"`
		KeyRisks             []string      `json:"key_risks"`
		RecommendedNextSteps []string      `json:"recommended_next_steps"`
		Evidence             []interface{} `json:"evidence"`
		ConfidenceScore      float64       `json:"confidence_score"`
		CorrelationID        string        `json:"correlation_id"`
	}
	if err := json.NewDecoder(hResp.Body).Decode(&sidecarRes); err != nil {
		return nil, fmt.Errorf("failed to decode sidecar risk response: %w", err)
	}

	sigBytes, _ := json.Marshal(signals)
	risksBytes, _ := json.Marshal(sidecarRes.KeyRisks)
	stepsBytes, _ := json.Marshal(sidecarRes.RecommendedNextSteps)
	evBytes, _ := json.Marshal(sidecarRes.Evidence)

	analysis := &ReceivablesAnalysis{
		OrgID:                orgID,
		InvoiceID:            invoiceID,
		CustomerID:           inv.CustomerID,
		RiskLevel:            sidecarRes.RiskLevel,
		RiskScore:            sidecarRes.RiskScore,
		DaysOverdue:          signals.DaysOverdue,
		AgingBucket:          signals.AgingBucket,
		OutstandingAmount:    inv.BalanceDue,
		Currency:             inv.Currency,
		ReceivablesSummary:   sidecarRes.ReceivablesSummary,
		DeterministicSignals: sigBytes,
		KeyRisks:             risksBytes,
		RecommendedNextSteps: stepsBytes,
		Evidence:             evBytes,
		ConfidenceScore:      sidecarRes.ConfidenceScore,
		CorrelationID:        corrID,
	}

	if err := s.repo.SaveAnalysis(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to persist receivables analysis: %w", err)
	}

	if s.auditSvc != nil {
		s.auditSvc.RecordAsync(ctx, domain.CreateAuditLogParams{
			OrgID:        orgID,
			ActorType:    domain.ActorTypeSystem,
			Action:       domain.ActionCreate,
			Module:       domain.ModuleInvoices,
			ResourceType: "ai_finance_receivables_analyses",
			ResourceID:   fmt.Sprintf("%d", analysis.ID),
			Description:  fmt.Sprintf("Completed AI receivables risk analysis for invoice %s (%s)", inv.InvoiceNumber, sidecarRes.RiskLevel),
		})
	}

	return analysis, nil
}

func (s *service) PrioritizeCollections(ctx context.Context, orgID, invoiceID int64) (json.RawMessage, error) {
	inv, err := s.invoiceRepo.GetInvoiceByID(ctx, orgID, invoiceID)
	if err != nil || inv == nil {
		return nil, fmt.Errorf("invoice not found or access denied: %w", err)
	}

	corrID := fmt.Sprintf("corr-fin-prio-%d-%d", invoiceID, time.Now().Unix())
	ctxMap, signals, err := s.buildInvoiceContextPayload(ctx, orgID, invoiceID, inv, corrID)
	if err != nil {
		return nil, err
	}

	reqPayload := map[string]interface{}{
		"context": ctxMap,
		"signals": signals,
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	url := fmt.Sprintf("%s/finance-ops/prioritize-collections", s.sidecarBaseURL)
	hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	hReq.Header.Set("Content-Type", "application/json")
	hReq.Header.Set("X-LogisticsHQ-Service-Key", getInternalServiceToken())

	hResp, err := s.httpClient.Do(hReq)
	if err != nil {
		return nil, err
	}
	defer hResp.Body.Close()

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(hResp.Body)
	return json.RawMessage(buf.Bytes()), nil
}

func (s *service) AnalyzeCustomerPaymentBehavior(ctx context.Context, orgID, invoiceID int64) (json.RawMessage, error) {
	inv, err := s.invoiceRepo.GetInvoiceByID(ctx, orgID, invoiceID)
	if err != nil || inv == nil {
		return nil, fmt.Errorf("invoice not found or access denied: %w", err)
	}

	corrID := fmt.Sprintf("corr-fin-cust-behavior-%d-%d", invoiceID, time.Now().Unix())
	ctxMap, signals, err := s.buildInvoiceContextPayload(ctx, orgID, invoiceID, inv, corrID)
	if err != nil {
		return nil, err
	}

	reqPayload := map[string]interface{}{
		"context": ctxMap,
		"signals": signals,
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	url := fmt.Sprintf("%s/finance-ops/analyze-customer-behavior", s.sidecarBaseURL)
	hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	hReq.Header.Set("Content-Type", "application/json")
	hReq.Header.Set("X-LogisticsHQ-Service-Key", getInternalServiceToken())

	hResp, err := s.httpClient.Do(hReq)
	if err != nil {
		return nil, err
	}
	defer hResp.Body.Close()

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(hResp.Body)
	return json.RawMessage(buf.Bytes()), nil
}

func (s *service) GetOperationalRecommendations(ctx context.Context, orgID, invoiceID int64) (json.RawMessage, error) {
	inv, err := s.invoiceRepo.GetInvoiceByID(ctx, orgID, invoiceID)
	if err != nil || inv == nil {
		return nil, fmt.Errorf("invoice not found or access denied: %w", err)
	}

	corrID := fmt.Sprintf("corr-fin-recs-%d-%d", invoiceID, time.Now().Unix())
	ctxMap, signals, err := s.buildInvoiceContextPayload(ctx, orgID, invoiceID, inv, corrID)
	if err != nil {
		return nil, err
	}

	reqPayload := map[string]interface{}{
		"context": ctxMap,
		"signals": signals,
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	url := fmt.Sprintf("%s/finance-ops/recommend-actions", s.sidecarBaseURL)
	hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	hReq.Header.Set("Content-Type", "application/json")
	hReq.Header.Set("X-LogisticsHQ-Service-Key", getInternalServiceToken())

	hResp, err := s.httpClient.Do(hReq)
	if err != nil {
		return nil, err
	}
	defer hResp.Body.Close()

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(hResp.Body)
	return json.RawMessage(buf.Bytes()), nil
}

func (s *service) GenerateCollectionDraft(
	ctx context.Context,
	orgID, invoiceID, userID int64,
	input GenerateCollectionDraftInput,
) (*CollectionDraft, error) {
	inv, err := s.invoiceRepo.GetInvoiceByID(ctx, orgID, invoiceID)
	if err != nil || inv == nil {
		return nil, fmt.Errorf("invoice not found or access denied: %w", err)
	}

	corrID := fmt.Sprintf("corr-fin-draft-%d-%d", invoiceID, time.Now().Unix())
	ctxMap, signals, err := s.buildInvoiceContextPayload(ctx, orgID, invoiceID, inv, corrID)
	if err != nil {
		return nil, err
	}

	draftType := "OVERDUE_NOTICE"
	if input.DraftType != "" {
		draftType = input.DraftType
	}
	tone := "ASSERTIVE"
	if input.Tone != "" {
		tone = input.Tone
	}

	draftReq := map[string]interface{}{
		"context":           ctxMap,
		"signals":           signals,
		"draft_type":        draftType,
		"tone":              tone,
		"user_instructions": input.UserInstructions,
		"correlation_id":    corrID,
	}
	bodyBytes, err := json.Marshal(draftReq)
	if err != nil {
		return nil, fmt.Errorf("failed to encode draft request: %w", err)
	}

	url := fmt.Sprintf("%s/finance-ops/generate-collection-draft", s.sidecarBaseURL)
	hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create sidecar draft request: %w", err)
	}
	hReq.Header.Set("Content-Type", "application/json")
	hReq.Header.Set("X-LogisticsHQ-Service-Key", getInternalServiceToken())

	hResp, err := s.httpClient.Do(hReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar draft call failed: %w", err)
	}
	defer hResp.Body.Close()

	if hResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar draft returned status %d", hResp.StatusCode)
	}

	var sidecarDraft struct {
		Subject          string            `json:"subject"`
		MessageBody      string            `json:"message_body"`
		InternalNotes    string            `json:"internal_notes"`
		RecipientPreview map[string]string `json:"recipient_preview"`
		RequiresApproval bool              `json:"requires_approval"`
	}
	if err := json.NewDecoder(hResp.Body).Decode(&sidecarDraft); err != nil {
		return nil, fmt.Errorf("failed to decode sidecar draft response: %w", err)
	}

	recipientName := sidecarDraft.RecipientPreview["name"]
	recipientEmail := sidecarDraft.RecipientPreview["email"]
	if input.RecipientName != "" {
		recipientName = input.RecipientName
	}
	if input.RecipientEmail != "" {
		recipientEmail = input.RecipientEmail
	}

	internalNotes := sidecarDraft.InternalNotes

	draft := &CollectionDraft{
		OrgID:             orgID,
		InvoiceID:         invoiceID,
		CustomerID:        inv.CustomerID,
		DraftType:         draftType,
		Subject:           sidecarDraft.Subject,
		MessageBody:       sidecarDraft.MessageBody,
		InternalNotes:     &internalNotes,
		RecipientName:     recipientName,
		RecipientEmail:    recipientEmail,
		OutstandingAmount: inv.BalanceDue,
		Currency:          inv.Currency,
		Status:            "DRAFT",
		RequiresApproval:  true, // Consequential communication always requires approval
		CreatedByUserID:   &userID,
		CorrelationID:     corrID,
	}

	if err := s.repo.SaveDraft(ctx, draft); err != nil {
		return nil, fmt.Errorf("failed to persist collection draft: %w", err)
	}

	if s.auditSvc != nil {
		s.auditSvc.RecordAsync(ctx, domain.CreateAuditLogParams{
			OrgID:        orgID,
			ActorID:      &userID,
			ActorType:    domain.ActorTypeUser,
			Action:       domain.ActionCreate,
			Module:       domain.ModuleInvoices,
			ResourceType: "ai_finance_collection_drafts",
			ResourceID:   fmt.Sprintf("%d", draft.ID),
			Description:  fmt.Sprintf("Generated AI %s draft for invoice %s (%s %.2f)", draftType, inv.InvoiceNumber, inv.Currency, inv.BalanceDue),
		})
	}

	return draft, nil
}

func (s *service) GetDraft(ctx context.Context, orgID, draftID int64) (*CollectionDraft, error) {
	return s.repo.GetDraftByID(ctx, orgID, draftID)
}

func (s *service) ListDrafts(ctx context.Context, orgID, invoiceID int64) ([]*CollectionDraft, error) {
	return s.repo.ListDraftsByInvoice(ctx, orgID, invoiceID)
}

func (s *service) UpdateDraft(ctx context.Context, orgID, draftID int64, input UpdateCollectionDraftInput) (*CollectionDraft, error) {
	draft, err := s.repo.GetDraftByID(ctx, orgID, draftID)
	if err != nil || draft == nil {
		return nil, fmt.Errorf("draft not found: %w", err)
	}

	if draft.Status == "APPROVED" || draft.Status == "SENT" {
		return nil, fmt.Errorf("cannot modify draft in '%s' status", draft.Status)
	}

	draft.Subject = input.Subject
	draft.MessageBody = input.MessageBody
	if input.InternalNotes != "" {
		draft.InternalNotes = &input.InternalNotes
	}
	if input.RecipientName != "" {
		draft.RecipientName = input.RecipientName
	}
	if input.RecipientEmail != "" {
		draft.RecipientEmail = input.RecipientEmail
	}

	if err := s.repo.UpdateDraft(ctx, draft); err != nil {
		return nil, err
	}
	return draft, nil
}

func (s *service) SubmitDraftForApproval(
	ctx context.Context,
	orgID, draftID, userID int64,
	input SubmitCollectionDraftApprovalInput,
) (*CollectionDraft, error) {
	draft, err := s.repo.GetDraftByID(ctx, orgID, draftID)
	if err != nil || draft == nil {
		return nil, fmt.Errorf("draft not found: %w", err)
	}

	if draft.Status == "PENDING_APPROVAL" {
		return draft, nil
	}

	title := fmt.Sprintf("Collection Message Approval: %s (Invoice #%d)", draft.DraftType, draft.InvoiceID)
	desc := fmt.Sprintf("Proposed collection communication to %s (%s) for %s %.2f. Subject: %s. Justification: %s",
		draft.RecipientName, draft.RecipientEmail, draft.Currency, draft.OutstandingAmount, draft.Subject, input.Reason)

	riskLevel := "HIGH"
	actionName := "finance.send_collection_reminder"
	if draft.DraftType == "FINAL_DEMAND" {
		actionName = "finance.send_formal_demand"
	} else if draft.DraftType == "INTERNAL_ESCALATION" {
		actionName = "finance.escalate_overdue_receivable"
	}

	payloadMap := map[string]interface{}{
		"draft_id":           draft.ID,
		"invoice_id":         draft.InvoiceID,
		"customer_id":        draft.CustomerID,
		"draft_type":         draft.DraftType,
		"subject":            draft.Subject,
		"recipient_name":     draft.RecipientName,
		"recipient_email":    draft.RecipientEmail,
		"outstanding_amount": draft.OutstandingAmount,
		"currency":           draft.Currency,
		"reason":             input.Reason,
	}
	payloadBytes, _ := json.Marshal(payloadMap)

	var approvalID *int64
	if s.approvalsSvc != nil {
		approvalReq, err := s.approvalsSvc.CreateApproval(ctx, orgID, &approvals.CreateApprovalInput{
			Title:             title,
			Category:          "FINANCE",
			Type:              "Collection Message Approval",
			Priority:          riskLevel,
			RelatedRef:        fmt.Sprintf("COL-DRAFT-%d", draft.ID),
			RelatedEntityType: "COLLECTION_DRAFT",
			RelatedEntityID:   draft.ID,
			CustomerName:      draft.RecipientName,
			RequestedByID:     userID,
			Description:       desc,
			RiskLevel:         riskLevel,
			ActionName:        actionName,
			ProposedPayload:   string(payloadBytes),
			CorrelationID:     draft.CorrelationID,
		}, "Finance Specialist")
		if err == nil && approvalReq != nil {
			approvalID = &approvalReq.ID
		}
	}

	proposalID := fmt.Sprintf("prop-fin-draft-%d-%d", draft.ID, time.Now().Unix())
	newStatus := "PENDING_APPROVAL"

	if err := s.repo.UpdateDraftStatus(ctx, orgID, draftID, newStatus, approvalID, &proposalID); err != nil {
		return nil, err
	}

	draft.Status = newStatus
	draft.ApprovalID = approvalID
	draft.ActionProposalID = &proposalID

	if s.auditSvc != nil {
		s.auditSvc.RecordAsync(ctx, domain.CreateAuditLogParams{
			OrgID:        orgID,
			ActorID:      &userID,
			ActorType:    domain.ActorTypeUser,
			Action:       domain.ActionUpdate,
			Module:       domain.ModuleInvoices,
			ResourceType: "ai_finance_collection_drafts",
			ResourceID:   fmt.Sprintf("%d", draftID),
			Description:  fmt.Sprintf("Submitted %s draft #%d for managerial approval", draft.DraftType, draftID),
		})
	}

	return draft, nil
}
