package operations_automation

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
	"github.com/freel/backend/internal/shipments"
	"github.com/freel/backend/internal/shipments/spec"
	"github.com/jmoiron/sqlx"
)

// Service defines the business interface for shipment operations automation
type Service interface {
	GetShipmentOperationsOverview(ctx context.Context, orgID, shipmentID int64) (*ShipmentOperationsOverview, error)
	AnalyzeShipmentRisks(ctx context.Context, orgID, shipmentID int64) (*ShipmentOperationsAnalysis, error)
	PrioritizeExceptions(ctx context.Context, orgID, shipmentID int64) (json.RawMessage, error)
	GetOperationalRecommendations(ctx context.Context, orgID, shipmentID int64) (json.RawMessage, error)
	GenerateCommunicationDraft(ctx context.Context, orgID, shipmentID, userID int64, input GenerateDraftInput) (*ShipmentCommunicationDraft, error)
	GetDraft(ctx context.Context, orgID, draftID int64) (*ShipmentCommunicationDraft, error)
	ListCommunicationDrafts(ctx context.Context, orgID, shipmentID int64) ([]*ShipmentCommunicationDraft, error)
	UpdateDraft(ctx context.Context, orgID, draftID int64, input UpdateDraftInput) (*ShipmentCommunicationDraft, error)
	SubmitDraftForApproval(ctx context.Context, orgID, draftID, userID int64, input SubmitDraftApprovalInput) (*ShipmentCommunicationDraft, error)
}

type service struct {
	db             *sqlx.DB
	repo           Repository
	shipmentRepo   shipments.Repository
	approvalsSvc   approvals.Service
	auditSvc       auditSvcPkg.Service
	sidecarBaseURL string
	httpClient     *http.Client
}

// NewService instantiates a new Service instance
func NewService(
	db *sqlx.DB,
	repo Repository,
	shipmentRepo shipments.Repository,
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
		shipmentRepo:   shipmentRepo,
		approvalsSvc:   approvalsSvc,
		auditSvc:       auditSvc,
		sidecarBaseURL: sidecarURL,
		httpClient: &http.Client{
			Timeout: 25 * time.Second,
		},
	}
}

func getInternalServiceToken() string {
	tok := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if tok == "" {
		tok = os.Getenv("INTERNAL_SERVICE_KEY")
	}
	if tok == "" {
		tok = "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456"
	}
	return tok
}

func (s *service) calculateDeterministicSignals(
	ctx context.Context,
	orgID, shipmentID int64,
	sh *spec.Shipment,
	milestones []*spec.ShipmentMilestone,
	exceptions []*spec.ShipmentException,
	docs []*spec.ShipmentDocument,
	trackingEvents []*spec.CarrierTrackingEvent,
) DeterministicSignalsDTO {
	now := time.Now()

	overdueMilestones := []string{}
	approachingMilestones := []string{}
	for _, m := range milestones {
		if m.Status != "COMPLETED" && m.PlannedDate != nil {
			if m.PlannedDate.Before(now) {
				overdueMilestones = append(overdueMilestones, m.MilestoneCode)
			} else if m.PlannedDate.Before(now.Add(48 * time.Hour)) {
				approachingMilestones = append(approachingMilestones, m.MilestoneCode)
			}
		}
	}

	activeExCount := 0
	critExCount := 0
	for _, ex := range exceptions {
		if !ex.Resolved && ex.Status != "DISMISSED" {
			activeExCount++
			if strings.EqualFold(ex.Severity, "CRITICAL") {
				critExCount++
			}
		}
	}

	// Calculate tracking freshness
	hoursSinceTracking := 0.0
	isTrackingStale := false
	var lastEventTime time.Time
	if len(trackingEvents) > 0 {
		lastEventTime = trackingEvents[0].EventTime
		for _, te := range trackingEvents {
			if te.EventTime.After(lastEventTime) {
				lastEventTime = te.EventTime
			}
		}
	} else if !sh.UpdatedAt.IsZero() {
		lastEventTime = sh.UpdatedAt
	} else {
		lastEventTime = sh.CreatedAt
	}

	if !lastEventTime.IsZero() {
		hoursSinceTracking = math.Max(0.0, now.Sub(lastEventTime).Hours())
		if hoursSinceTracking > 24.0 && (sh.Status == "IN_TRANSIT" || sh.Status == "BOOKED") {
			isTrackingStale = true
		}
	}

	// Schedule variance
	isDelayed := false
	delayDays := 0.0
	if sh.ETA != nil && now.After(*sh.ETA) && sh.Status != "DELIVERED" {
		isDelayed = true
		delayDays = math.Max(0.0, now.Sub(*sh.ETA).Hours()/24.0)
	}

	// Missing document detection
	mandatoryDocs := []string{"BILL_OF_LADING", "COMMERCIAL_INVOICE", "PACKING_LIST"}
	missingDocs := []string{}
	existingDocTypes := make(map[string]bool)
	for _, d := range docs {
		existingDocTypes[strings.ToUpper(d.DocType)] = true
	}
	for _, mDoc := range mandatoryDocs {
		if !existingDocTypes[mDoc] {
			missingDocs = append(missingDocs, mDoc)
		}
	}

	carrierGapHours := 0.0
	isCarrierOverdue := false
	if isTrackingStale {
		carrierGapHours = hoursSinceTracking
		if carrierGapHours > 24.0 {
			isCarrierOverdue = true
		}
	}

	// Deterministic risk score calculation (0 - 100)
	riskScore := 10.0
	if len(overdueMilestones) > 0 {
		riskScore += 20.0 + float64(len(overdueMilestones)-1)*5.0
	}
	if critExCount > 0 {
		riskScore += 35.0 + float64(critExCount-1)*15.0
	} else if activeExCount > 0 {
		riskScore += 15.0 + float64(activeExCount-1)*5.0
	}
	if isTrackingStale {
		riskScore += 15.0
	}
	if len(missingDocs) > 0 {
		riskScore += 15.0
	}
	if isDelayed {
		riskScore += math.Min(20.0, delayDays*5.0)
	}
	if riskScore > 100.0 {
		riskScore = 100.0
	}

	riskLevel := "LOW"
	if riskScore >= 70.0 || critExCount > 0 {
		riskLevel = "CRITICAL"
	} else if riskScore >= 45.0 {
		riskLevel = "HIGH"
	} else if riskScore >= 25.0 {
		riskLevel = "MEDIUM"
	}

	requiresApproval := riskLevel == "HIGH" || riskLevel == "CRITICAL" || activeExCount > 0

	return DeterministicSignalsDTO{
		HasOverdueMilestone:      len(overdueMilestones) > 0,
		OverdueMilestones:        overdueMilestones,
		ApproachingMilestones:    approachingMilestones,
		HasActiveException:       activeExCount > 0,
		ActiveExceptionsCount:    activeExCount,
		CriticalExceptionsCount:  critExCount,
		IsTrackingStale:          isTrackingStale,
		HoursSinceLastTracking:   math.Round(hoursSinceTracking*10) / 10,
		IsDelayed:                isDelayed,
		DelayDays:                math.Round(delayDays*10) / 10,
		MissingDocumentsCount:    len(missingDocs),
		MissingDocuments:         missingDocs,
		CarrierResponseGapHours:  math.Round(carrierGapHours*10) / 10,
		IsCarrierResponseOverdue: isCarrierOverdue,
		RiskScore:                math.Round(riskScore*10) / 10,
		RiskLevel:                riskLevel,
		RequiresApproval:         requiresApproval,
	}
}

func (s *service) buildShipmentContextPayload(
	ctx context.Context,
	orgID, shipmentID int64,
	sh *spec.Shipment,
	corrID string,
) (map[string]interface{}, DeterministicSignalsDTO, error) {
	milestones, _ := s.shipmentRepo.GetMilestones(ctx, shipmentID)
	exceptions, _ := s.shipmentRepo.GetExceptions(ctx, orgID, shipmentID)
	docs, _ := s.shipmentRepo.GetShipmentDocuments(ctx, orgID, shipmentID)
	trackingEvents, _ := s.shipmentRepo.GetCarrierEventsForShipment(ctx, orgID, shipmentID)

	signals := s.calculateDeterministicSignals(ctx, orgID, shipmentID, sh, milestones, exceptions, docs, trackingEvents)

	// Build milestones list
	msList := []map[string]interface{}{}
	for _, m := range milestones {
		var pDateStr, aDateStr *string
		if m.PlannedDate != nil {
			str := m.PlannedDate.Format(time.RFC3339)
			pDateStr = &str
		}
		if m.ActualDate != nil {
			str := m.ActualDate.Format(time.RFC3339)
			aDateStr = &str
		}
		msList = append(msList, map[string]interface{}{
			"milestone_id":   m.ID,
			"milestone_code": m.MilestoneCode,
			"description":    m.Description,
			"planned_date":   pDateStr,
			"actual_date":    aDateStr,
			"status":         m.Status,
			"location":       m.Location,
		})
	}

	// Build exceptions list
	exList := []map[string]interface{}{}
	for _, ex := range exceptions {
		cAtStr := ex.CreatedAt.Format(time.RFC3339)
		exList = append(exList, map[string]interface{}{
			"exception_id":     ex.ID,
			"exception_type":   ex.ExceptionType,
			"severity":         ex.Severity,
			"status":           ex.Status,
			"title":            ex.Title,
			"description":      ex.Description,
			"resolved":         ex.Resolved,
			"resolution_notes": ex.ResolutionNotes,
			"created_at":       cAtStr,
		})
	}

	// Build documents list
	docList := []map[string]interface{}{}
	for _, d := range docs {
		docList = append(docList, map[string]interface{}{
			"document_id": d.ID,
			"doc_type":    d.DocType,
			"file_name":   d.FileName,
			"status":      d.Status,
			"is_verified": d.Status == "VERIFIED",
		})
	}

	// Customer name lookup
	custName := "Commercial Cargo Customer"
	if sh.CustomerName != nil && *sh.CustomerName != "" {
		custName = *sh.CustomerName
	}

	carrierName := sh.CarrierSCAC
	if sh.CarrierName != nil && *sh.CarrierName != "" {
		carrierName = *sh.CarrierName
	}

	var etdStr, etaStr *string
	if sh.ETD != nil {
		str := sh.ETD.Format(time.RFC3339)
		etdStr = &str
	}
	if sh.ETA != nil {
		str := sh.ETA.Format(time.RFC3339)
		etaStr = &str
	}

	ctxMap := map[string]interface{}{
		"org_id":            orgID,
		"shipment_id":       shipmentID,
		"rfq_id":            sh.RFQID,
		"booking_id":        sh.BookingID,
		"carrier_scac":      sh.CarrierSCAC,
		"carrier_name":      carrierName,
		"booking_number":    sh.BookingNumber,
		"mbl_number":        sh.MBLNumber,
		"hbl_number":        sh.HBLNumber,
		"container_numbers": sh.ContainerNumbers,
		"status":            sh.Status,
		"origin_port":       sh.OriginPort,
		"destination_port":  sh.DestinationPort,
		"vessel_name":       sh.VesselName,
		"voyage_number":     sh.VoyageNumber,
		"etd":               etdStr,
		"eta":               etaStr,
		"customer_id":       sh.CustomerID,
		"customer_name":     custName,
		"customer_tier":     "ENTERPRISE",
		"milestones":        msList,
		"exceptions":        exList,
		"documents":         docList,
		"correlation_id":    corrID,
	}

	return ctxMap, signals, nil
}

func (s *service) GetShipmentOperationsOverview(ctx context.Context, orgID, shipmentID int64) (*ShipmentOperationsOverview, error) {
	sh, err := s.shipmentRepo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil || sh == nil {
		return nil, fmt.Errorf("shipment not found or access denied: %w", err)
	}

	corrID := fmt.Sprintf("corr-ops-overview-%d-%d", shipmentID, time.Now().Unix())
	ctxMap, signals, err := s.buildShipmentContextPayload(ctx, orgID, shipmentID, sh, corrID)
	if err != nil {
		return nil, err
	}

	analysis, _ := s.repo.GetLatestAnalysis(ctx, orgID, shipmentID)
	latestDraft, _ := s.repo.GetLatestDraft(ctx, orgID, shipmentID)

	var prioritizedEx json.RawMessage
	var recsRaw json.RawMessage

	// Fetch recommendations & prioritized exceptions from sidecar for instant operational overview
	reqPayload := map[string]interface{}{
		"context": ctxMap,
		"signals": signals,
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	// Call /shipment-ops/prioritize-exceptions
	urlEx := fmt.Sprintf("%s/shipment-ops/prioritize-exceptions", s.sidecarBaseURL)
	if hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, urlEx, bytes.NewReader(bodyBytes)); err == nil {
		hReq.Header.Set("Content-Type", "application/json")
		hReq.Header.Set("X-LogisticsHQ-Service-Key", getInternalServiceToken())
		if hResp, err := s.httpClient.Do(hReq); err == nil && hResp.StatusCode == http.StatusOK {
			defer hResp.Body.Close()
			var parsed map[string]interface{}
			if err := json.NewDecoder(hResp.Body).Decode(&parsed); err == nil {
				if items, ok := parsed["prioritized_exceptions"]; ok {
					prioritizedEx, _ = json.Marshal(items)
				}
			}
		}
	}

	// Call /shipment-ops/recommend-actions
	urlRec := fmt.Sprintf("%s/shipment-ops/recommend-actions", s.sidecarBaseURL)
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

	custName := "Commercial Cargo Customer"
	if sh.CustomerName != nil {
		custName = *sh.CustomerName
	}
	carrierName := sh.CarrierSCAC
	if sh.CarrierName != nil {
		carrierName = *sh.CarrierName
	}
	mblNum := ""
	if sh.MBLNumber != nil {
		mblNum = *sh.MBLNumber
	}
	bkgNum := ""
	if sh.BookingNumber != nil {
		bkgNum = *sh.BookingNumber
	}
	vesselName := ""
	if sh.VesselName != nil {
		vesselName = *sh.VesselName
	}
	voyageNum := ""
	if sh.VoyageNumber != nil {
		voyageNum = *sh.VoyageNumber
	}
	var etdStr, etaStr *string
	if sh.ETD != nil {
		str := sh.ETD.Format(time.RFC3339)
		etdStr = &str
	}
	if sh.ETA != nil {
		str := sh.ETA.Format(time.RFC3339)
		etaStr = &str
	}

	var approvalID *int64
	pendingApproval := false
	if latestDraft != nil {
		approvalID = latestDraft.ApprovalID
		pendingApproval = latestDraft.Status == "PENDING_APPROVAL"
	}

	return &ShipmentOperationsOverview{
		ShipmentID:            shipmentID,
		CarrierSCAC:           sh.CarrierSCAC,
		CarrierName:           carrierName,
		Status:                sh.Status,
		MBLNumber:             mblNum,
		BookingNumber:         bkgNum,
		OriginPort:            sh.OriginPort,
		DestinationPort:       sh.DestinationPort,
		VesselName:            vesselName,
		VoyageNumber:          voyageNum,
		ETD:                   etdStr,
		ETA:                   etaStr,
		CustomerName:          custName,
		CustomerTier:          "ENTERPRISE",
		Signals:               signals,
		LatestAnalysis:        analysis,
		LatestDraft:           latestDraft,
		PrioritizedExceptions: prioritizedEx,
		Recommendations:       recsRaw,
		PendingApproval:       pendingApproval,
		ApprovalID:            approvalID,
		CorrelationID:         corrID,
	}, nil
}

func (s *service) AnalyzeShipmentRisks(ctx context.Context, orgID, shipmentID int64) (*ShipmentOperationsAnalysis, error) {
	sh, err := s.shipmentRepo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil || sh == nil {
		return nil, fmt.Errorf("shipment not found or access denied: %w", err)
	}

	corrID := fmt.Sprintf("corr-risk-analysis-%d-%d", shipmentID, time.Now().Unix())
	ctxMap, signals, err := s.buildShipmentContextPayload(ctx, orgID, shipmentID, sh, corrID)
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

	url := fmt.Sprintf("%s/shipment-ops/analyze-risks", s.sidecarBaseURL)
	hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create sidecar risk request: %w", err)
	}
	hReq.Header.Set("Content-Type", "application/json")
	hReq.Header.Set("X-LogisticsHQ-Service-Key", getInternalServiceToken())

	hResp, err := s.httpClient.Do(hReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar risk analysis call failed: %w", err)
	}
	defer hResp.Body.Close()

	if hResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sidecar risk analysis returned status %d", hResp.StatusCode)
	}

	var sidecarRes struct {
		RiskLevel            string                 `json:"risk_level"`
		RiskScore            float64                `json:"risk_score"`
		OperationalSummary   string                 `json:"operational_summary"`
		KeySignals           []string               `json:"key_signals"`
		RecommendedNextSteps []string               `json:"recommended_next_steps"`
		Evidence             []interface{}          `json:"evidence"`
		ConfidenceScore      float64                `json:"confidence_score"`
		CorrelationID        string                 `json:"correlation_id"`
	}
	if err := json.NewDecoder(hResp.Body).Decode(&sidecarRes); err != nil {
		return nil, fmt.Errorf("failed to decode sidecar risk response: %w", err)
	}

	sigBytes, _ := json.Marshal(signals)
	risksBytes, _ := json.Marshal(sidecarRes.KeySignals)
	stepsBytes, _ := json.Marshal(sidecarRes.RecommendedNextSteps)
	evBytes, _ := json.Marshal(sidecarRes.Evidence)

	analysis := &ShipmentOperationsAnalysis{
		OrgID:                orgID,
		ShipmentID:           shipmentID,
		RiskLevel:            signals.RiskLevel, // Authoritative Go signal preserves risk level
		RiskScore:            signals.RiskScore, // Authoritative Go signal preserves risk score
		OperationalSummary:   sidecarRes.OperationalSummary,
		DeterministicSignals: sigBytes,
		KeyRisks:             risksBytes,
		RecommendedNextSteps: stepsBytes,
		Evidence:             evBytes,
		ConfidenceScore:      sidecarRes.ConfidenceScore,
		CorrelationID:        corrID,
	}

	if err := s.repo.SaveAnalysis(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to persist shipment operations analysis: %w", err)
	}

	return analysis, nil
}

func (s *service) PrioritizeExceptions(ctx context.Context, orgID, shipmentID int64) (json.RawMessage, error) {
	sh, err := s.shipmentRepo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil || sh == nil {
		return nil, fmt.Errorf("shipment not found: %w", err)
	}

	corrID := fmt.Sprintf("corr-ex-prioritize-%d-%d", shipmentID, time.Now().Unix())
	ctxMap, signals, err := s.buildShipmentContextPayload(ctx, orgID, shipmentID, sh, corrID)
	if err != nil {
		return nil, err
	}

	reqPayload := map[string]interface{}{
		"context": ctxMap,
		"signals": signals,
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	url := fmt.Sprintf("%s/shipment-ops/prioritize-exceptions", s.sidecarBaseURL)
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

	var result json.RawMessage
	if err := json.NewDecoder(hResp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *service) GetOperationalRecommendations(ctx context.Context, orgID, shipmentID int64) (json.RawMessage, error) {
	sh, err := s.shipmentRepo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil || sh == nil {
		return nil, fmt.Errorf("shipment not found: %w", err)
	}

	corrID := fmt.Sprintf("corr-ops-recs-%d-%d", shipmentID, time.Now().Unix())
	ctxMap, signals, err := s.buildShipmentContextPayload(ctx, orgID, shipmentID, sh, corrID)
	if err != nil {
		return nil, err
	}

	reqPayload := map[string]interface{}{
		"context": ctxMap,
		"signals": signals,
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	url := fmt.Sprintf("%s/shipment-ops/recommend-actions", s.sidecarBaseURL)
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

	var result json.RawMessage
	if err := json.NewDecoder(hResp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *service) GenerateCommunicationDraft(
	ctx context.Context,
	orgID, shipmentID, userID int64,
	input GenerateDraftInput,
) (*ShipmentCommunicationDraft, error) {
	sh, err := s.shipmentRepo.GetShipmentByID(ctx, orgID, shipmentID)
	if err != nil || sh == nil {
		return nil, fmt.Errorf("shipment not found: %w", err)
	}

	corrID := fmt.Sprintf("corr-ship-draft-%d-%d", shipmentID, time.Now().Unix())
	ctxMap, signals, err := s.buildShipmentContextPayload(ctx, orgID, shipmentID, sh, corrID)
	if err != nil {
		return nil, err
	}

	draftType := input.DraftType
	if draftType == "" {
		draftType = "CARRIER_FOLLOWUP"
	}

	draftReq := map[string]interface{}{
		"shipment_id":       shipmentID,
		"draft_type":        draftType,
		"recipient_name":    input.RecipientName,
		"recipient_email":   input.RecipientEmail,
		"context":           ctxMap,
		"signals":           signals,
		"tone":              input.Tone,
		"user_instructions": input.UserInstructions,
		"correlation_id":    corrID,
	}
	bodyBytes, err := json.Marshal(draftReq)
	if err != nil {
		return nil, fmt.Errorf("failed to encode draft request: %w", err)
	}

	url := fmt.Sprintf("%s/shipment-ops/generate-draft", s.sidecarBaseURL)
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
		CustomerWording  string            `json:"customer_wording"`
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

	draft := &ShipmentCommunicationDraft{
		OrgID:            orgID,
		ShipmentID:       shipmentID,
		DraftType:        draftType,
		Subject:          sidecarDraft.Subject,
		CustomerWording:  sidecarDraft.CustomerWording,
		InternalNotes:    &internalNotes,
		RecipientName:    recipientName,
		RecipientEmail:   recipientEmail,
		Status:           "DRAFT",
		RequiresApproval: true, // Consequential communication always requires approval
		CreatedByUserID:  &userID,
		CorrelationID:    corrID,
	}

	if err := s.repo.SaveDraft(ctx, draft); err != nil {
		return nil, fmt.Errorf("failed to persist shipment communication draft: %w", err)
	}

	if s.auditSvc != nil {
		s.auditSvc.RecordAsync(ctx, domain.CreateAuditLogParams{
			OrgID:        orgID,
			ActorID:      &userID,
			ActorType:    domain.ActorTypeUser,
			Action:       domain.ActionCreate,
			Module:       domain.ModuleShipments,
			ResourceType: "shipment_communication_drafts",
			ResourceID:   fmt.Sprintf("%d", draft.ID),
			Description:  fmt.Sprintf("Generated AI %s communication draft for shipment #%d", draftType, shipmentID),
		})
	}

	return draft, nil
}

func (s *service) GetDraft(ctx context.Context, orgID, draftID int64) (*ShipmentCommunicationDraft, error) {
	return s.repo.GetDraftByID(ctx, orgID, draftID)
}

func (s *service) ListCommunicationDrafts(ctx context.Context, orgID, shipmentID int64) ([]*ShipmentCommunicationDraft, error) {
	return s.repo.ListDraftsByShipment(ctx, orgID, shipmentID)
}

func (s *service) UpdateDraft(ctx context.Context, orgID, draftID int64, input UpdateDraftInput) (*ShipmentCommunicationDraft, error) {
	draft, err := s.repo.GetDraftByID(ctx, orgID, draftID)
	if err != nil || draft == nil {
		return nil, fmt.Errorf("draft not found: %w", err)
	}

	if draft.Status == "APPROVED" || draft.Status == "EXECUTED" {
		return nil, fmt.Errorf("cannot modify draft in '%s' status", draft.Status)
	}

	draft.Subject = input.Subject
	draft.CustomerWording = input.CustomerWording
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
	input SubmitDraftApprovalInput,
) (*ShipmentCommunicationDraft, error) {
	draft, err := s.repo.GetDraftByID(ctx, orgID, draftID)
	if err != nil || draft == nil {
		return nil, fmt.Errorf("draft not found: %w", err)
	}

	if draft.Status == "PENDING_APPROVAL" {
		return draft, nil
	}

	title := fmt.Sprintf("Shipment Communication Approval: %s (Shipment #%d)", draft.DraftType, draft.ShipmentID)
	desc := fmt.Sprintf("Operational message proposed for transmission to %s (%s). Subject: %s. Reason: %s",
		draft.RecipientName, draft.RecipientEmail, draft.Subject, input.Reason)

	riskLevel := "HIGH"
	actionName := "shipments.send_customer_update"
	if draft.DraftType == "CARRIER_FOLLOWUP" {
		actionName = "shipments.request_carrier_followup"
	} else if draft.DraftType == "INTERNAL_ESCALATION" {
		actionName = "shipments.escalate_exception"
	}

	payloadMap := map[string]interface{}{
		"draft_id":        draft.ID,
		"shipment_id":     draft.ShipmentID,
		"draft_type":      draft.DraftType,
		"subject":         draft.Subject,
		"recipient_name":  draft.RecipientName,
		"recipient_email": draft.RecipientEmail,
		"reason":          input.Reason,
	}
	payloadBytes, _ := json.Marshal(payloadMap)

	var approvalID *int64
	if s.approvalsSvc != nil {
		approvalReq, err := s.approvalsSvc.CreateApproval(ctx, orgID, &approvals.CreateApprovalInput{
			Title:             title,
			Category:          "OPERATIONS",
			Type:              "Shipment Communication Approval",
			Priority:          riskLevel,
			RelatedRef:        fmt.Sprintf("DRAFT-%d", draft.ID),
			RelatedEntityType: "SHIPMENT_DRAFT",
			RelatedEntityID:   draft.ID,
			CustomerName:      draft.RecipientName,
			RequestedByID:     userID,
			Description:       desc,
			RiskLevel:         riskLevel,
			ActionName:        actionName,
			ProposedPayload:   string(payloadBytes),
			CorrelationID:     draft.CorrelationID,
		}, "Operations Specialist")
		if err == nil && approvalReq != nil {
			approvalID = &approvalReq.ID
		}
	}

	proposalID := fmt.Sprintf("prop-ship-draft-%d-%d", draft.ID, time.Now().Unix())
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
			Module:       domain.ModuleShipments,
			ResourceType: "shipment_communication_drafts",
			ResourceID:   fmt.Sprintf("%d", draftID),
			Description:  fmt.Sprintf("Submitted %s draft #%d for managerial approval", draft.DraftType, draftID),
		})
	}

	return draft, nil
}

// Helpers for unmarshaling sidecar arrays
func (s *service) getPrioritizedExceptionsArray(m map[string]interface{}) interface{} {
	if p, ok := m["prioritized_exceptions"]; ok {
		return p
	}
	return []interface{}{}
}

func (s *service) getRecommendationsArray(m map[string]interface{}) interface{} {
	if r, ok := m["recommendations"]; ok {
		return r
	}
	return []interface{}{}
}
