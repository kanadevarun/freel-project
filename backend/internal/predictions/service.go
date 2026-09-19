package predictions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidOrgID          = errors.New("invalid or missing organization context")
	ErrInvalidModule         = errors.New("unsupported prediction module")
	ErrInvalidPredictionType = errors.New("unsupported prediction type")
	ErrInvalidRecordType     = errors.New("unsupported related record type")
	ErrMissingRecordID       = errors.New("missing related record identifier")
	ErrInvalidConfidence     = errors.New("confidence score must be between 0.0 and 1.0")
	ErrMissingSourceRef      = errors.New("prediction must contain at least one verifiable source reference")
	ErrStatementTooLong      = errors.New("prediction statement exceeds maximum length of 1000 characters")
	ErrExplanationTooLong    = errors.New("explanation exceeds maximum length of 5000 characters")
	ErrPredictionNotFound    = errors.New("prediction record not found")
	ErrInvalidTransition     = errors.New("illegal prediction lifecycle transition")
)

var allowedModules = map[string]bool{
	"shipments":    true,
	"rfqs":         true,
	"pricing":      true,
	"finance":      true,
	"contracts":    true,
	"customers":    true,
	"leads":        true,
	"cross_module": true,
	"carriers":     true,
	"network":      true,
	"workload":     true,
	"capacity":     true,
	"demand":       true,
	"resource":     true,
	"bottleneck":   true,
}

var allowedPredictionTypes = map[string]bool{
	"SHIPMENT_ETA_DELAY":               true,
	"SHIPMENT_TRANSSHIPMENT_EXCEPTION": true,
	"SHIPMENT_EXCEPTION_RISK":          true,
	"SHIPMENT_DISRUPTION_FORECAST":     true,
	"INVOICE_PAYMENT_DEFAULT":          true,
	"INVOICE_DISPUTE_PROBABILITY":      true,
	"CUSTOMER_CHURN_RISK":              true,
	"CUSTOMER_VOLUME_DROP":             true,
	"CUSTOMER_REPEAT_BUSINESS":         true,
	"CUSTOMER_ENGAGEMENT_RISK":         true,
	"LEAD_CONVERSION_LIKELIHOOD":       true,
	"LEAD_INACTIVITY_RISK":             true,
	"LEAD_ENGAGEMENT_RISK":             true,
	"RFQ_WIN_PROBABILITY":              true,
	"CONTRACT_DEMURRAGE_RISK":          true,
	"CONTRACT_COMPLIANCE_BREACH":       true,
	"CARRIER_RELIABILITY_DROP":         true,
	"RFQ_MARGIN_RISK":                  true,
	"QUOTATION_COMPETITIVENESS":        true,
	"PRICING_COST_VARIANCE_RISK":       true,
	"CONTRACT_RATE_PRESSURE":           true,
	"HISTORICAL_MARGIN_INTELLIGENCE":   true,
	"INVOICE_LATE_PAYMENT_RISK":        true,
	"COLLECTION_PRIORITY":              true,
	"CASH_INFLOW_FORECAST":             true,
	"DISPUTE_PAYMENT_DELAY_RISK":       true,
	"CUSTOMER_PAYMENT_BEHAVIOR":        true,
	"RECEIVABLES_CONCENTRATION_RISK":   true,
	"CONTRACT_EXPIRY_RENEWAL_RISK":     true,
	"CONTRACT_CLAUSE_COMMERCIAL_RISK":  true,
	"DOCUMENTATION_COMPLETENESS_RISK":  true,
	"COMPLIANCE_REVIEW_RISK":           true,
	"CROSS_MODULE_CONTRACT_RISK":       true,
	"HISTORICAL_DOCUMENTATION_RISK":    true,
	"SHIPMENT_READINESS_RISK":          true,
	"CUTOFF_MISS_RISK":                 true,
	"DOCUMENTATION_DELAY_RISK":         true,
	"CUSTOMS_PROCESSING_RISK":          true,
	"BILLING_READINESS_RISK":           true,
	"HISTORICAL_OPERATIONAL_RISK":      true,
	"CUSTOMER_SERVICE_RISK":            true,
	"CUSTOMER_RELATIONSHIP_RISK":       true,
	"CARRIER_PERFORMANCE_RISK":         true,
	"CARRIER_DELAY_RISK":               true,
	"LANE_PERFORMANCE_RISK":            true,
	"LANE_DISRUPTION_RISK":             true,
	"NETWORK_BOTTLENECK_RISK":          true,
	"SERVICE_LEVEL_RISK":               true,
	"COST_PRESSURE_RISK":               true,
	"APPROVAL_WORKLOAD_SPIKE":          true,
	"OPERATIONAL_WORKLOAD_SPIKE":       true,
	"DOCUMENTATION_WORKLOAD_SPIKE":     true,
	"DEMAND_CAPACITY_MISMATCH":         true,
	"LANE_CAPACITY_PRESSURE":           true,
	"CARRIER_CAPACITY_PRESSURE":        true,
	"QUOTE_PROCESSING_BOTTLENECK":      true,
	"SHIPMENT_PROCESSING_BOTTLENECK":   true,
	"DEMAND_VOLUME_FORECAST":           true,
	"CAPACITY_SHORTAGE_RISK":           true,
	"OPERATIONAL_BOTTLENECK":           true,
	"RESOURCE_ALLOCATION_IMBALANCE":    true,
	"APPROVAL_BOTTLENECK":              true,
	"DOCUMENTATION_BOTTLENECK":         true,
	"EXCEPTION_RESOLUTION_BOTTLENECK":   true,
	"CROSS_MODULE_BOTTLENECK":          true,
	"CUTOFF_CONCENTRATION_BOTTLENECK":   true,
	"OWNER_WORKLOAD_IMBALANCE":         true,
	"PORT_CONGESTION_RISK":             true,
	"CARRIER_UPDATE_GAP_RISK":          true,
	"FREE_TIME_EXPIRY_RISK":            true,
	"NETWORK_DISRUPTION_INTELLIGENCE":  true,
}

var allowedRecordTypes = map[string]bool{
	"SHIPMENT":   true,
	"RFQ":        true,
	"INVOICE":    true,
	"CONTRACT":   true,
	"CUSTOMER":   true,
	"LEAD":       true,
	"CARRIER":    true,
	"LANE":       true,
	"PORT":       true,
	"WORKLOAD":   true,
	"CAPACITY":   true,
	"DEMAND":     true,
	"OPERATIONS": true,
	"BOTTLENECK": true,
	"RESOURCE":   true,
}

type Service interface {
	GenerateAndPersist(ctx context.Context, orgID int64, userID *int64, req *GenerateRequest) (*Prediction, error)
	GetPrediction(ctx context.Context, orgID int64, predictionID string) (*Prediction, error)
	ListPredictions(ctx context.Context, orgID int64, params FilterParams) ([]*Prediction, int, error)
	GetOrPredictShipmentETA(ctx context.Context, orgID int64, userID *int64, shipmentID int64, forceRefresh bool) (*Prediction, error)
	GetOrForecastShipmentExceptions(ctx context.Context, orgID int64, userID *int64, shipmentID int64, forceRefresh bool) (*Prediction, error)
	GetOrPredictCustomerIntelligence(ctx context.Context, orgID int64, userID *int64, customerID int64, forceRefresh bool) (*Prediction, error)
	GetOrPredictLeadIntelligence(ctx context.Context, orgID int64, userID *int64, leadID int64, forceRefresh bool) (*Prediction, error)
	GetOrPredictRFQMarginIntelligence(ctx context.Context, orgID int64, userID *int64, rfqID int64, forceRefresh bool) (*Prediction, error)
	GetOrPredictContractRatePressure(ctx context.Context, orgID int64, userID *int64, contractID int64, forceRefresh bool) (*Prediction, error)
	GetOrPredictInvoiceCollections(ctx context.Context, orgID int64, userID *int64, invoiceID int64, forceRefresh bool) (*Prediction, error)
	GetOrPredictContractComplianceRisk(ctx context.Context, orgID int64, userID *int64, contractID int64, forceRefresh bool) (*Prediction, error)
	GetOrPredictShipmentReadiness(ctx context.Context, orgID int64, userID *int64, shipmentID int64, forceRefresh bool) (*Prediction, error)
	GetOrPredictCarrierPerformance(ctx context.Context, orgID int64, userID *int64, scac string, forceRefresh bool) (*Prediction, error)
	GetOrPredictLanePerformance(ctx context.Context, orgID int64, userID *int64, laneCode string, forceRefresh bool) (*Prediction, error)
	GetOrPredictCustomerServicePerformance(ctx context.Context, orgID int64, userID *int64, customerID int64, forceRefresh bool) (*Prediction, error)
	GetOrPredictWorkloadPlanning(ctx context.Context, orgID int64, userID *int64, workloadType string, forceRefresh bool) (*Prediction, error)
	GetOrPredictCapacityPlanning(ctx context.Context, orgID int64, userID *int64, dimension string, forceRefresh bool) (*Prediction, error)
	GetOrPredictDemandPlanning(ctx context.Context, orgID int64, userID *int64, segment string, forceRefresh bool) (*Prediction, error)
	GetOrPredictOperationalBottleneck(ctx context.Context, orgID int64, userID *int64, bottleneckType string, forceRefresh bool) (*Prediction, error)
	GetOrPredictResourceAllocation(ctx context.Context, orgID int64, userID *int64, resourceType string, forceRefresh bool) (*Prediction, error)
	GetResourceBottleneckSummary(ctx context.Context, orgID int64, userID *int64) (map[string]interface{}, error)
	AcknowledgePrediction(ctx context.Context, orgID int64, predictionID string, userID *int64) error
	DismissPrediction(ctx context.Context, orgID int64, predictionID string, userID *int64, reason string) error
	RequestAction(ctx context.Context, orgID int64, predictionID string, userID *int64, notes string) error
	RecordOutcome(ctx context.Context, orgID int64, predictionID string, outcomeStatus OutcomeStatus, outcomeValue *string, feedbackNotes *string) error
	GetAuditHistory(ctx context.Context, orgID int64, predictionID string) ([]*PredictionAuditHistory, error)
}

type service struct {
	repo                   Repository
	sidecar                SidecarClient
	db                     *sql.DB
	dataProvider           *ShipmentDataProvider
	clDataProvider         *CustomerLeadDataProvider
	pricingDataProvider    *PricingMarginDataProvider
	financeDataProvider    *FinanceCollectionsDataProvider
	contractComplianceData *ContractComplianceDataProvider
	shipmentReadinessData  *ShipmentReadinessDataProvider
	networkPerfData        *NetworkPerformanceDataProvider
	workloadCapacityData   *WorkloadCapacityDataProvider
	resourceBottleneckData *ResourceBottleneckDataProvider
}

func NewService(repo Repository, sidecar SidecarClient, db ...*sql.DB) Service {
	s := &service{
		repo:    repo,
		sidecar: sidecar,
	}
	if len(db) > 0 && db[0] != nil {
		s.db = db[0]
		s.dataProvider = NewShipmentDataProvider(db[0])
		s.clDataProvider = NewCustomerLeadDataProvider(db[0])
		s.pricingDataProvider = NewPricingMarginDataProvider(db[0])
		s.financeDataProvider = NewFinanceCollectionsDataProvider(db[0])
		s.contractComplianceData = NewContractComplianceDataProvider(db[0])
		s.shipmentReadinessData = NewShipmentReadinessDataProvider(db[0])
		s.networkPerfData = NewNetworkPerformanceDataProvider(db[0])
		s.workloadCapacityData = NewWorkloadCapacityDataProvider(db[0])
		s.resourceBottleneckData = NewResourceBottleneckDataProvider(db[0])
	}
	return s
}

func (s *service) isKillSwitchActive(ctx context.Context, orgID int64, module, predType string) (bool, string) {
	if s.db == nil {
		return false, ""
	}
	query := `
		SELECT reason FROM ai_governance_kill_switches
		WHERE is_killed = 1 AND (org_id = ? OR org_id = 0) AND (
			(scope = 'GLOBAL' AND target_identifier = 'GLOBAL_AI') OR
			(scope = 'WORKFLOW' AND target_identifier IN ('WORKFLOW:predictions', 'WORKFLOW:predictive_intelligence', ?)) OR
			(scope = 'PREDICTION' AND target_identifier IN ('PREDICTION:*', ?))
		)
		LIMIT 1
	`
	wfTarget := "WORKFLOW:" + strings.ToLower(module)
	predTarget := "PREDICTION:" + strings.ToUpper(predType)
	var reason string
	err := s.db.QueryRowContext(ctx, query, orgID, wfTarget, predTarget).Scan(&reason)
	if err == nil {
		return true, reason
	}
	return false, ""
}

func (s *service) GenerateAndPersist(ctx context.Context, orgID int64, userID *int64, req *GenerateRequest) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	req.OrgID = orgID

	// 1. Validate request parameters
	req.Module = strings.ToLower(strings.TrimSpace(req.Module))
	if !allowedModules[req.Module] {
		return nil, fmt.Errorf("%w: %s", ErrInvalidModule, req.Module)
	}

	req.PredictionType = strings.ToUpper(strings.TrimSpace(req.PredictionType))
	if !allowedPredictionTypes[req.PredictionType] {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPredictionType, req.PredictionType)
	}

	req.RelatedRecordType = strings.ToUpper(strings.TrimSpace(req.RelatedRecordType))
	if !allowedRecordTypes[req.RelatedRecordType] {
		return nil, fmt.Errorf("%w: %s", ErrInvalidRecordType, req.RelatedRecordType)
	}

	req.RelatedRecordID = strings.TrimSpace(req.RelatedRecordID)
	if req.RelatedRecordID == "" {
		return nil, ErrMissingRecordID
	}

	// 2. Check Administrative AI Kill Switch
	if killed, reason := s.isKillSwitchActive(ctx, orgID, req.Module, req.PredictionType); killed {
		return nil, fmt.Errorf("predictive intelligence execution blocked by administrative kill switch: %s", reason)
	}

	// 3. Deterministic Idempotency Deduplication
	idempotencyKey := fmt.Sprintf("pred-%d-%s-%s-%s", orgID, req.Module, req.PredictionType, req.RelatedRecordID)
	existing, err := s.repo.GetByIdempotencyKey(ctx, orgID, idempotencyKey)
	if err == nil && existing != nil {
		// If existing prediction is active (not dismissed or expired), return it
		if existing.Status != StatusDismissed && existing.Status != StatusExpired {
			return existing, nil
		}
	}

	// 4. Invoke Python AI Sidecar for grounded prediction inference
	sidecarResp, err := s.sidecar.GeneratePrediction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prediction via AI sidecar: %w", err)
	}

	// 4. Validate Python sidecar output schema and boundaries
	if sidecarResp.OrgID != orgID {
		return nil, fmt.Errorf("tenant context violation: sidecar returned org %d, expected %d", sidecarResp.OrgID, orgID)
	}

	if sidecarResp.ConfidenceScore < 0.0 || sidecarResp.ConfidenceScore > 1.0 {
		return nil, ErrInvalidConfidence
	}

	if len(sidecarResp.PredictionStatement) > 1000 {
		return nil, ErrStatementTooLong
	}

	if len(sidecarResp.Explanation) > 5000 {
		return nil, ErrExplanationTooLong
	}

	if len(sidecarResp.SourceReferences) == 0 {
		return nil, ErrMissingSourceRef
	}

	for _, ref := range sidecarResp.SourceReferences {
		if strings.TrimSpace(ref.SourceRecordID) == "" || strings.TrimSpace(ref.SourceField) == "" {
			return nil, fmt.Errorf("%w: invalid source citation: %+v", ErrMissingSourceRef, ref)
		}
	}

	// 5. Build persistent record
	now := time.Now().UTC()
	var targetDate *time.Time
	if sidecarResp.TargetDate != nil && *sidecarResp.TargetDate != "" {
		if t, err := time.Parse(time.RFC3339, *sidecarResp.TargetDate); err == nil {
			targetDate = &t
		}
	}

	var sourceTS time.Time
	if t, err := time.Parse(time.RFC3339, sidecarResp.SourceTimestamp); err == nil {
		sourceTS = t
	} else {
		sourceTS = now
	}

	expiresAt := now.Add(14 * 24 * time.Hour) // Predictions expire in 14 days by default

	predID := sidecarResp.PredictionID
	if predID == "" {
		predID = fmt.Sprintf("pred-%s-%d", strings.ToLower(req.RelatedRecordType), time.Now().UnixNano())
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        predID,
		IdempotencyKey:      idempotencyKey,
		Module:              req.Module,
		PredictionType:      req.PredictionType,
		Status:              StatusPublished,
		Severity:            Severity(strings.ToUpper(sidecarResp.Severity)),
		ConfidenceScore:     sidecarResp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(strings.ToUpper(sidecarResp.ConfidenceBand)),
		RelatedRecordType:   req.RelatedRecordType,
		RelatedRecordID:     req.RelatedRecordID,
		PredictionStatement: sidecarResp.PredictionStatement,
		PredictedValue:      sidecarResp.PredictedValue,
		TimeHorizon:         sidecarResp.TimeHorizon,
		TargetDate:          targetDate,
		Explanation:         sidecarResp.Explanation,
		SupportingSignals:   sidecarResp.SupportingSignals,
		SourceReferences:    sidecarResp.SourceReferences,
		SourceTimestamp:     sourceTS,
		RecommendedAction:   sidecarResp.RecommendedAction,
		ActionType:          sidecarResp.ActionType,
		IsActionRequired:    sidecarResp.IsActionRequired,
		RequiresApproval:    sidecarResp.RequiresApproval,
		DocumentReference:   sidecarResp.DocumentReference,
		ClauseReference:     sidecarResp.ClauseReference,
		PageNumber:          sidecarResp.PageNumber,
		SectionHeading:      sidecarResp.SectionHeading,
		MilestoneReference:  sidecarResp.MilestoneReference,
		CutoffReference:     sidecarResp.CutoffReference,
		ReviewStatus:        "UNREVIEWED",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        sidecarResp.ModelVersion,
		ExpiresAt:           &expiresAt,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	// 6. Save in MariaDB
	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed to persist prediction: %w", err)
	}

	// 7. Write audit log entry
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID: pred.PredictionID,
		OrgID:        orgID,
		UserID:       userID,
		NewStatus:    string(StatusPublished),
		Action:       "PREDICTION_PUBLISHED",
		Notes:        &pred.PredictionStatement,
	})

	return pred, nil
}

func (s *service) GetPrediction(ctx context.Context, orgID int64, predictionID string) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	pred, err := s.repo.GetByID(ctx, orgID, predictionID)
	if err != nil {
		return nil, err
	}
	if pred == nil {
		return nil, ErrPredictionNotFound
	}
	return pred, nil
}

func (s *service) ListPredictions(ctx context.Context, orgID int64, params FilterParams) ([]*Prediction, int, error) {
	if orgID <= 0 {
		return nil, 0, ErrInvalidOrgID
	}
	return s.repo.List(ctx, orgID, params)
}

func (s *service) GetOrPredictShipmentETA(ctx context.Context, orgID int64, userID *int64, shipmentID int64, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	if shipmentID <= 0 {
		return nil, fmt.Errorf("invalid shipment id: %d", shipmentID)
	}

	shipmentIDStr := strconv.FormatInt(shipmentID, 10)

	// 1. If not forcing refresh, check for an existing active prediction
	if !forceRefresh {
		active, err := s.repo.GetActivePrediction(ctx, orgID, "shipments", "SHIPMENT", shipmentIDStr, "SHIPMENT_ETA_DELAY")
		if err == nil && active != nil {
			// If not expired and created within last 24h, return cached active prediction
			if active.ExpiresAt == nil || active.ExpiresAt.After(time.Now()) {
				return active, nil
			}
		}
	}

	// 2. Fetch deterministic shipment facts
	if s.dataProvider == nil {
		return nil, fmt.Errorf("shipment data provider not configured (database connection missing)")
	}

	shData, err := s.dataProvider.FetchDeterministicShipmentData(ctx, orgID, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare deterministic shipment facts: %w", err)
	}

	// 3. Construct GenerateRequest for Python sidecar
	genReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "shipments",
		PredictionType:    "SHIPMENT_ETA_DELAY",
		RelatedRecordType: "SHIPMENT",
		RelatedRecordID:   shipmentIDStr,
		RecordContext:     shData.ContextMap,
		TimeHorizon:       "7_DAYS",
	}

	// 4. Call Python sidecar
	resp, err := s.sidecar.GeneratePrediction(ctx, genReq)
	if err != nil {
		return nil, fmt.Errorf("predictive AI sidecar error: %w", err)
	}

	// 5. Validate response
	if err := s.validateResponse(genReq, resp); err != nil {
		return nil, fmt.Errorf("prediction validation error: %w", err)
	}

	// 6. Compute deterministic idempotency key
	dayBucket := time.Now().UTC().Format("2006-01-02-15")
	milestoneTimeStr := "none"
	if shData.LatestMilestoneTime != nil {
		milestoneTimeStr = shData.LatestMilestoneTime.Format("200601021504")
	}
	idempotencyKey := fmt.Sprintf("%d:shipment:%d:eta:%s:%s", orgID, shipmentID, milestoneTimeStr, dayBucket)
	if forceRefresh {
		idempotencyKey = fmt.Sprintf("%s:refresh:%d", idempotencyKey, time.Now().UnixNano())
	} else {
		// Check if duplicate already exists
		existing, _ := s.repo.GetByIdempotencyKey(ctx, orgID, idempotencyKey)
		if existing != nil {
			return existing, nil
		}
	}

	// 7. Parse dates & create Prediction object
	sourceTime := time.Now()
	if shData.LatestMilestoneTime != nil {
		sourceTime = *shData.LatestMilestoneTime
	}

	var targetDate *time.Time
	if resp.TargetDate != nil && *resp.TargetDate != "" {
		for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339, "2006-01-02"} {
			if t, err := time.Parse(layout, *resp.TargetDate); err == nil {
				targetDate = &t
				break
			}
		}
	}
	if targetDate == nil && shData.AuthoritativeETA != nil {
		targetDate = shData.AuthoritativeETA
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "shipments",
		PredictionType:      "SHIPMENT_ETA_DELAY",
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "SHIPMENT",
		RelatedRecordID:     shipmentIDStr,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		TargetDate:          targetDate,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	// 8. Supersede any existing active predictions for this shipment
	_ = s.repo.SupersedeExisting(ctx, orgID, "shipments", "SHIPMENT", shipmentIDStr, "SHIPMENT_ETA_DELAY", resp.PredictionID)

	// 9. Persist the new prediction
	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed to persist shipment ETA prediction: %w", err)
	}

	// 10. Audit logging
	prevStatus := string(StatusGenerated)
	auditNotes := fmt.Sprintf("Published %s predictive ETA forecast with confidence %.2f", pred.Severity, pred.ConfidenceScore)
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

func (s *service) GetOrForecastShipmentExceptions(ctx context.Context, orgID int64, userID *int64, shipmentID int64, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	if shipmentID <= 0 {
		return nil, errors.New("invalid shipment ID")
	}

	shipmentIDStr := fmt.Sprintf("%d", shipmentID)
	now := time.Now().UTC()
	hourBucket := now.Format("2006010215")
	idempotencyKey := fmt.Sprintf("org:%d:shipment_exception_forecast:%d:cycle_%s", orgID, shipmentID, hourBucket)

	// 1. Check if active prediction already exists for this hour cycle (unless forceRefresh requested)
	if !forceRefresh {
		existing, err := s.repo.GetActivePrediction(ctx, orgID, "shipments", "SHIPMENT", shipmentIDStr, "SHIPMENT_EXCEPTION_RISK")
		if err == nil && existing != nil {
			return existing, nil
		}
	} else {
		idempotencyKey = fmt.Sprintf("%s:refresh:%d", idempotencyKey, time.Now().UnixNano())
	}

	// 2. Fetch deterministic operational records from MariaDB
	disruptData, err := s.dataProvider.FetchExceptionDisruptionData(ctx, orgID, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch operational records for shipment %d: %w", shipmentID, err)
	}

	// 3. Build Sidecar Prediction Request
	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "shipments",
		PredictionType:    "SHIPMENT_EXCEPTION_RISK",
		RelatedRecordType: "SHIPMENT",
		RelatedRecordID:   shipmentIDStr,
		RecordContext:     disruptData.ContextMap,
		TimeHorizon:       "7_DAYS",
	}

	// 4. Call Python AI Sidecar
	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar predictive exception engine failed: %w", err)
	}

	// 5. Validate Sidecar Response
	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar prediction validation failed: %w", err)
	}

	// 6. Build persistent Prediction record
	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "shipments",
		PredictionType:      "SHIPMENT_EXCEPTION_RISK",
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "SHIPMENT",
		RelatedRecordID:     shipmentIDStr,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	// 7. Supersede any older active disruption forecasts for this shipment
	_ = s.repo.SupersedeExisting(ctx, orgID, "shipments", "SHIPMENT", shipmentIDStr, "SHIPMENT_EXCEPTION_RISK", resp.PredictionID)

	// 8. Persist the new prediction
	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed to persist shipment exception disruption prediction: %w", err)
	}

	// 9. Audit logging
	prevStatus := string(StatusGenerated)
	auditNotes := fmt.Sprintf("Published %s disruption forecast for shipment %d with confidence %.2f", pred.Severity, shipmentID, pred.ConfidenceScore)
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

func (s *service) GetOrPredictLeadIntelligence(ctx context.Context, orgID int64, userID *int64, leadID int64, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	if leadID <= 0 {
		return nil, errors.New("invalid lead ID")
	}

	leadIDStr := fmt.Sprintf("%d", leadID)
	now := time.Now().UTC()
	hourBucket := now.Format("2006010215")
	idempotencyKey := fmt.Sprintf("org:%d:lead_intelligence:%d:cycle_%s", orgID, leadID, hourBucket)

	// Determine prediction type dynamically or check cache
	targetPredType := "LEAD_CONVERSION_LIKELIHOOD"

	if !forceRefresh {
		// Check cache for active prediction of any lead prediction type
		for _, pt := range []string{"LEAD_CONVERSION_LIKELIHOOD", "LEAD_INACTIVITY_RISK", "LEAD_ENGAGEMENT_RISK"} {
			existing, err := s.repo.GetActivePrediction(ctx, orgID, "leads", "LEAD", leadIDStr, pt)
			if err == nil && existing != nil {
				return existing, nil
			}
		}
	} else {
		idempotencyKey = fmt.Sprintf("%s:refresh:%d", idempotencyKey, time.Now().UnixNano())
	}

	// Fetch deterministic telemetry from real MariaDB
	if s.clDataProvider == nil {
		return nil, errors.New("customer lead data provider not initialized")
	}
	leadData, err := s.clDataProvider.FetchLeadData(ctx, orgID, leadID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching lead data for lead %d: %w", leadID, err)
	}

	// Classify appropriate prediction type based on deterministic facts
	if leadData.UnansweredInquiries > 0 || leadData.InactivityHours >= 48.0 {
		targetPredType = "LEAD_INACTIVITY_RISK"
	} else if len(leadData.MissingFields) > 0 {
		targetPredType = "LEAD_ENGAGEMENT_RISK"
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "leads",
		PredictionType:    targetPredType,
		RelatedRecordType: "LEAD",
		RelatedRecordID:   leadIDStr,
		RecordContext:     leadData.ContextMap,
		TimeHorizon:       "30_DAYS",
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar lead intelligence engine failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "leads",
		PredictionType:      targetPredType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "LEAD",
		RelatedRecordID:     leadIDStr,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
	}

	if forceRefresh {
		for _, pt := range []string{"LEAD_CONVERSION_LIKELIHOOD", "LEAD_INACTIVITY_RISK", "LEAD_ENGAGEMENT_RISK"} {
			_ = s.repo.SupersedeExisting(ctx, orgID, "leads", "LEAD", leadIDStr, pt, pred.PredictionID)
		}
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting lead intelligence prediction: %w", err)
	}

	prevStatus := string(StatusGenerated)
	auditNotes := fmt.Sprintf("Published %s %s prediction for lead %d (confidence: %.2f)", pred.Severity, targetPredType, leadID, pred.ConfidenceScore)
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

func (s *service) GetOrPredictCustomerIntelligence(ctx context.Context, orgID int64, userID *int64, customerID int64, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	if customerID <= 0 {
		return nil, errors.New("invalid customer ID")
	}

	customerIDStr := fmt.Sprintf("%d", customerID)
	now := time.Now().UTC()
	hourBucket := now.Format("2006010215")
	idempotencyKey := fmt.Sprintf("org:%d:customer_intelligence:%d:cycle_%s", orgID, customerID, hourBucket)

	targetPredType := "CUSTOMER_REPEAT_BUSINESS"

	if !forceRefresh {
		for _, pt := range []string{"CUSTOMER_CHURN_RISK", "CUSTOMER_REPEAT_BUSINESS", "CUSTOMER_ENGAGEMENT_RISK"} {
			existing, err := s.repo.GetActivePrediction(ctx, orgID, "customers", "CUSTOMER", customerIDStr, pt)
			if err == nil && existing != nil {
				return existing, nil
			}
		}
	} else {
		idempotencyKey = fmt.Sprintf("%s:refresh:%d", idempotencyKey, time.Now().UnixNano())
	}

	if s.clDataProvider == nil {
		return nil, errors.New("customer lead data provider not initialized")
	}
	custData, err := s.clDataProvider.FetchCustomerData(ctx, orgID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching customer data for customer %d: %w", customerID, err)
	}

	// Classify appropriate prediction type based on deterministic operational facts
	if custData.InactivityDays >= 45 || custData.HealthScore < 70 || custData.CreditStatus == "WARNING" || custData.CreditStatus == "HOLD" {
		targetPredType = "CUSTOMER_CHURN_RISK"
	} else if custData.HasUnresolvedQuotes {
		targetPredType = "CUSTOMER_ENGAGEMENT_RISK"
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "customers",
		PredictionType:    targetPredType,
		RelatedRecordType: "CUSTOMER",
		RelatedRecordID:   customerIDStr,
		RecordContext:     custData.ContextMap,
		TimeHorizon:       "60_DAYS",
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar customer intelligence engine failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "customers",
		PredictionType:      targetPredType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "CUSTOMER",
		RelatedRecordID:     customerIDStr,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
	}

	if forceRefresh {
		for _, pt := range []string{"CUSTOMER_CHURN_RISK", "CUSTOMER_REPEAT_BUSINESS", "CUSTOMER_ENGAGEMENT_RISK"} {
			_ = s.repo.SupersedeExisting(ctx, orgID, "customers", "CUSTOMER", customerIDStr, pt, pred.PredictionID)
		}
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting customer intelligence prediction: %w", err)
	}

	prevStatus := string(StatusGenerated)
	auditNotes := fmt.Sprintf("Published %s %s prediction for customer %d (confidence: %.2f)", pred.Severity, targetPredType, customerID, pred.ConfidenceScore)
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

func (s *service) validateResponse(req *GenerateRequest, resp *SidecarPredictionResponse) error {
	if resp.OrgID != req.OrgID {
		return fmt.Errorf("tenant context violation: sidecar returned org %d, expected %d", resp.OrgID, req.OrgID)
	}
	if resp.ConfidenceScore < 0.0 || resp.ConfidenceScore > 1.0 {
		return ErrInvalidConfidence
	}
	if len(resp.PredictionStatement) > 1000 {
		return ErrStatementTooLong
	}
	if len(resp.Explanation) > 5000 {
		return ErrExplanationTooLong
	}
	if len(resp.SourceReferences) == 0 {
		return ErrMissingSourceRef
	}
	for _, ref := range resp.SourceReferences {
		if strings.TrimSpace(ref.SourceRecordID) == "" || strings.TrimSpace(ref.SourceField) == "" {
			return fmt.Errorf("%w: invalid source citation: %+v", ErrMissingSourceRef, ref)
		}
	}
	return nil
}

func (s *service) AcknowledgePrediction(ctx context.Context, orgID int64, predictionID string, userID *int64) error {
	pred, err := s.GetPrediction(ctx, orgID, predictionID)
	if err != nil {
		return err
	}

	if pred.Status == StatusDismissed || pred.Status == StatusExpired {
		return fmt.Errorf("%w: cannot acknowledge a %s prediction", ErrInvalidTransition, pred.Status)
	}

	prevStatus := string(pred.Status)
	notes := "Acknowledged by operator"
	if err := s.repo.UpdateStatus(ctx, orgID, predictionID, StatusAcknowledged, "ACKNOWLEDGED", userID, &notes); err != nil {
		return err
	}

	return s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   predictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(StatusAcknowledged),
		Action:         "ACKNOWLEDGED",
		Notes:          &notes,
	})
}

func (s *service) DismissPrediction(ctx context.Context, orgID int64, predictionID string, userID *int64, reason string) error {
	pred, err := s.GetPrediction(ctx, orgID, predictionID)
	if err != nil {
		return err
	}

	prevStatus := string(pred.Status)
	notes := strings.TrimSpace(reason)
	if notes == "" {
		notes = "Dismissed by operator without notes"
	}

	if err := s.repo.UpdateStatus(ctx, orgID, predictionID, StatusDismissed, "DISMISSED", userID, &notes); err != nil {
		return err
	}

	return s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   predictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(StatusDismissed),
		Action:         "DISMISSED",
		Notes:          &notes,
	})
}

func (s *service) RequestAction(ctx context.Context, orgID int64, predictionID string, userID *int64, notes string) error {
	pred, err := s.GetPrediction(ctx, orgID, predictionID)
	if err != nil {
		return err
	}

	if pred.ActionType == nil || *pred.ActionType == "" {
		return fmt.Errorf("prediction %s has no actionable recommendation configured", predictionID)
	}

	prevStatus := string(pred.Status)
	nextStatus := StatusActionRequested
	if pred.RequiresApproval {
		nextStatus = StatusAwaitingApproval
	}

	auditNotes := fmt.Sprintf("Action requested: %s. Operator notes: %s", *pred.ActionType, notes)
	if err := s.repo.UpdateStatus(ctx, orgID, predictionID, nextStatus, "ACTION_TAKEN", userID, &auditNotes); err != nil {
		return err
	}

	return s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   predictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(nextStatus),
		Action:         "ACTION_REQUESTED",
		Notes:          &auditNotes,
	})
}

func (s *service) RecordOutcome(ctx context.Context, orgID int64, predictionID string, outcomeStatus OutcomeStatus, outcomeValue *string, feedbackNotes *string) error {
	pred, err := s.GetPrediction(ctx, orgID, predictionID)
	if err != nil {
		return err
	}

	if err := s.repo.RecordOutcome(ctx, orgID, predictionID, outcomeStatus, outcomeValue, feedbackNotes); err != nil {
		return err
	}

	notes := fmt.Sprintf("Recorded actual outcome: %s (Observed: %v)", outcomeStatus, outcomeValue)
	currentStatus := string(pred.Status)
	return s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   predictionID,
		OrgID:          orgID,
		PreviousStatus: &currentStatus,
		NewStatus:      currentStatus,
		Action:         "OUTCOME_RECORDED",
		Notes:          &notes,
	})
}

func (s *service) GetAuditHistory(ctx context.Context, orgID int64, predictionID string) ([]*PredictionAuditHistory, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	return s.repo.GetAuditHistory(ctx, orgID, predictionID)
}

func (s *service) GetOrPredictRFQMarginIntelligence(ctx context.Context, orgID int64, userID *int64, rfqID int64, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	if rfqID <= 0 {
		return nil, errors.New("invalid rfq ID")
	}

	rfqIDStr := fmt.Sprintf("%d", rfqID)
	now := time.Now().UTC()
	hourBucket := now.Format("2006010215")
	idempotencyKey := fmt.Sprintf("org:%d:rfq_margin:%d:cycle_%s", orgID, rfqID, hourBucket)

	targetPredType := "RFQ_MARGIN_RISK"

	if !forceRefresh {
		for _, pt := range []string{"RFQ_MARGIN_RISK", "QUOTATION_COMPETITIVENESS", "PRICING_COST_VARIANCE_RISK", "HISTORICAL_MARGIN_INTELLIGENCE"} {
			existing, err := s.repo.GetActivePrediction(ctx, orgID, "pricing", "RFQ", rfqIDStr, pt)
			if err == nil && existing != nil {
				return existing, nil
			}
		}
	} else {
		idempotencyKey = fmt.Sprintf("%s:refresh:%d", idempotencyKey, time.Now().UnixNano())
	}

	if s.pricingDataProvider == nil {
		return nil, errors.New("pricing data provider not initialized")
	}
	rfqData, err := s.pricingDataProvider.FetchRFQData(ctx, orgID, rfqID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching rfq pricing data for rfq %d: %w", rfqID, err)
	}

	// Classify appropriate prediction type based on deterministic operational facts
	if rfqData.HasMissingCosts || rfqData.BuyPrice <= 0 || rfqData.SellPrice <= 0 {
		targetPredType = "PRICING_COST_VARIANCE_RISK"
	} else if rfqData.MarginPercentage >= 10.0 {
		targetPredType = "QUOTATION_COMPETITIVENESS"
	} else {
		targetPredType = "RFQ_MARGIN_RISK"
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "pricing",
		PredictionType:    targetPredType,
		RelatedRecordType: "RFQ",
		RelatedRecordID:   rfqIDStr,
		RecordContext:     rfqData.ContextMap,
		TimeHorizon:       "14_DAYS",
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar rfq pricing engine failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "pricing",
		PredictionType:      targetPredType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "RFQ",
		RelatedRecordID:     rfqIDStr,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
		InsufficientData:    resp.InsufficientData,
		InsufficientReason:  resp.InsufficientReason,
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	if forceRefresh {
		for _, pt := range []string{"RFQ_MARGIN_RISK", "QUOTATION_COMPETITIVENESS", "PRICING_COST_VARIANCE_RISK", "HISTORICAL_MARGIN_INTELLIGENCE"} {
			_ = s.repo.SupersedeExisting(ctx, orgID, "pricing", "RFQ", rfqIDStr, pt, pred.PredictionID)
		}
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting rfq pricing prediction: %w", err)
	}

	prevStatus := "DRAFT"
	auditNotes := "Initial rfq pricing margin prediction published"
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

func (s *service) GetOrPredictContractRatePressure(ctx context.Context, orgID int64, userID *int64, contractID int64, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	if contractID <= 0 {
		return nil, errors.New("invalid contract ID")
	}

	contractIDStr := fmt.Sprintf("%d", contractID)
	now := time.Now().UTC()
	hourBucket := now.Format("2006010215")
	idempotencyKey := fmt.Sprintf("org:%d:contract_rate:%d:cycle_%s", orgID, contractID, hourBucket)

	targetPredType := "CONTRACT_RATE_PRESSURE"

	if !forceRefresh {
		existing, err := s.repo.GetActivePrediction(ctx, orgID, "contracts", "CONTRACT", contractIDStr, targetPredType)
		if err == nil && existing != nil {
			return existing, nil
		}
	} else {
		idempotencyKey = fmt.Sprintf("%s:refresh:%d", idempotencyKey, time.Now().UnixNano())
	}

	if s.pricingDataProvider == nil {
		return nil, errors.New("pricing data provider not initialized")
	}
	contractData, err := s.pricingDataProvider.FetchContractData(ctx, orgID, contractID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching contract rate data for contract %d: %w", contractID, err)
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "contracts",
		PredictionType:    targetPredType,
		RelatedRecordType: "CONTRACT",
		RelatedRecordID:   contractIDStr,
		RecordContext:     contractData.ContextMap,
		TimeHorizon:       "30_DAYS",
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar contract rate pressure engine failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "contracts",
		PredictionType:      targetPredType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "CONTRACT",
		RelatedRecordID:     contractIDStr,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
		InsufficientData:    resp.InsufficientData,
		InsufficientReason:  resp.InsufficientReason,
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	if forceRefresh {
		_ = s.repo.SupersedeExisting(ctx, orgID, "contracts", "CONTRACT", contractIDStr, targetPredType, pred.PredictionID)
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting contract rate pressure prediction: %w", err)
	}

	prevStatus := "DRAFT"
	auditNotes := "Initial contract rate pressure prediction published"
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

func (s *service) GetOrPredictInvoiceCollections(ctx context.Context, orgID int64, userID *int64, invoiceID int64, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	if invoiceID <= 0 {
		return nil, errors.New("invalid invoice ID")
	}

	invoiceIDStr := fmt.Sprintf("%d", invoiceID)
	now := time.Now().UTC()
	hourBucket := now.Format("2006010215")
	idempotencyKey := fmt.Sprintf("org:%d:invoice_collections:%d:cycle_%s", orgID, invoiceID, hourBucket)

	targetPredType := "INVOICE_LATE_PAYMENT_RISK"

	if !forceRefresh {
		existing, err := s.repo.GetByIdempotencyKey(ctx, orgID, idempotencyKey)
		if err == nil && existing != nil && existing.Status != StatusSuperseded {
			return existing, nil
		}
		for _, pt := range []string{"COLLECTION_PRIORITY", "INVOICE_LATE_PAYMENT_RISK", "CASH_INFLOW_FORECAST", "DISPUTE_PAYMENT_DELAY_RISK", "CUSTOMER_PAYMENT_BEHAVIOR", "RECEIVABLES_CONCENTRATION_RISK"} {
			active, err := s.repo.GetActivePrediction(ctx, orgID, "finance", "INVOICE", invoiceIDStr, pt)
			if err == nil && active != nil && (active.ExpiresAt == nil || active.ExpiresAt.After(now)) {
				return active, nil
			}
		}
	} else {
		idempotencyKey = fmt.Sprintf("%s:refresh:%d", idempotencyKey, time.Now().UnixNano())
	}

	if s.financeDataProvider == nil {
		return nil, fmt.Errorf("finance data provider not initialized")
	}

	invData, err := s.financeDataProvider.FetchInvoiceData(ctx, orgID, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching invoice facts: %w", err)
	}

	if invData.BalanceDue <= 0.0 || invData.Status == "Paid" || invData.Status == "PAID" {
		targetPredType = "COLLECTION_PRIORITY"
	} else if !invData.IsOverdue && invData.DaysLeft > 0 {
		targetPredType = "CASH_INFLOW_FORECAST"
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "finance",
		PredictionType:    targetPredType,
		RelatedRecordType: "INVOICE",
		RelatedRecordID:   invoiceIDStr,
		RecordContext:     invData.ContextMap,
		TimeHorizon:       "30_DAYS",
		CorrelationID:     idempotencyKey,
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar finance collections engine failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "finance",
		PredictionType:      targetPredType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "INVOICE",
		RelatedRecordID:     invoiceIDStr,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
		InsufficientData:    resp.InsufficientData,
		InsufficientReason:  resp.InsufficientReason,
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	if forceRefresh {
		for _, pt := range []string{
			"INVOICE_LATE_PAYMENT_RISK",
			"COLLECTION_PRIORITY",
			"CASH_INFLOW_FORECAST",
			"DISPUTE_PAYMENT_DELAY_RISK",
			"CUSTOMER_PAYMENT_BEHAVIOR",
			"RECEIVABLES_CONCENTRATION_RISK",
			"INVOICE_PAYMENT_DEFAULT",
		} {
			_ = s.repo.SupersedeExisting(ctx, orgID, "finance", "INVOICE", invoiceIDStr, pt, pred.PredictionID)
		}
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting invoice collections prediction: %w", err)
	}

	prevStatus := "DRAFT"
	auditNotes := "Initial invoice predictive collection forecast published"
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

func (s *service) GetOrPredictContractComplianceRisk(ctx context.Context, orgID int64, userID *int64, contractID int64, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	if contractID <= 0 {
		return nil, ErrMissingRecordID
	}

	contractIDStr := strconv.FormatInt(contractID, 10)
	now := time.Now().UTC()

	// Default target prediction type
	targetPredType := "CONTRACT_EXPIRY_RENEWAL_RISK"
	idempotencyKey := fmt.Sprintf("pred:contracts:comp:%d:%d:%s", orgID, contractID, now.Format("2006-01-02-15"))

	if !forceRefresh {
		existing, err := s.repo.GetActivePrediction(ctx, orgID, "contracts", "CONTRACT", contractIDStr, targetPredType)
		if err == nil && existing != nil {
			return existing, nil
		}
		// Also check clause discrepancy or documentation risk
		for _, altType := range []string{"CONTRACT_CLAUSE_COMMERCIAL_RISK", "DOCUMENTATION_COMPLETENESS_RISK", "COMPLIANCE_REVIEW_RISK"} {
			if altExisting, err := s.repo.GetActivePrediction(ctx, orgID, "contracts", "CONTRACT", contractIDStr, altType); err == nil && altExisting != nil {
				return altExisting, nil
			}
		}
	} else {
		idempotencyKey = fmt.Sprintf("%s:refresh:%d", idempotencyKey, time.Now().UnixNano())
	}

	if s.contractComplianceData == nil {
		return nil, errors.New("contract compliance data provider not initialized")
	}

	compData, err := s.contractComplianceData.FetchContractComplianceData(ctx, orgID, contractID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching contract compliance data for contract %d: %w", contractID, err)
	}

	// Determine specific prediction type based on facts
	if compData.HasPendingDocument || compData.StructuredDiscrepancyCount > 0 {
		targetPredType = "CONTRACT_CLAUSE_COMMERCIAL_RISK"
	} else if len(compData.MissingDocuments) > 0 {
		targetPredType = "DOCUMENTATION_COMPLETENESS_RISK"
	} else if compData.IsExpired || compData.IsExpiringSoon {
		targetPredType = "CONTRACT_EXPIRY_RENEWAL_RISK"
	} else {
		targetPredType = "COMPLIANCE_REVIEW_RISK"
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "contracts",
		PredictionType:    targetPredType,
		RelatedRecordType: "CONTRACT",
		RelatedRecordID:   contractIDStr,
		RecordContext:     compData.ContextMap,
		TimeHorizon:       "30_DAYS",
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar contract compliance prediction failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "contracts",
		PredictionType:      resp.PredictionType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "CONTRACT",
		RelatedRecordID:     contractIDStr,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		DocumentReference:   resp.DocumentReference,
		ClauseReference:     resp.ClauseReference,
		PageNumber:          resp.PageNumber,
		SectionHeading:      resp.SectionHeading,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
		InsufficientData:    resp.InsufficientData,
		InsufficientReason:  resp.InsufficientReason,
	}

	if resp.ClauseReference != nil && len(pred.SourceReferences) > 0 {
		for i := range pred.SourceReferences {
			if pred.SourceReferences[i].ClauseReference == "" {
				pred.SourceReferences[i].ClauseReference = *resp.ClauseReference
				if resp.PageNumber != nil {
					pred.SourceReferences[i].PageNumber = *resp.PageNumber
				}
				if resp.SectionHeading != nil {
					pred.SourceReferences[i].SectionHeading = *resp.SectionHeading
				}
				if resp.DocumentReference != nil && pred.SourceReferences[i].DocumentReference == "" {
					pred.SourceReferences[i].DocumentReference = *resp.DocumentReference
				}
				break
			}
		}
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	if forceRefresh {
		for _, pt := range []string{
			"CONTRACT_EXPIRY_RENEWAL_RISK",
			"CONTRACT_CLAUSE_COMMERCIAL_RISK",
			"DOCUMENTATION_COMPLETENESS_RISK",
			"COMPLIANCE_REVIEW_RISK",
			"CROSS_MODULE_CONTRACT_RISK",
			"HISTORICAL_DOCUMENTATION_RISK",
			"CONTRACT_COMPLIANCE_BREACH",
		} {
			_ = s.repo.SupersedeExisting(ctx, orgID, "contracts", "CONTRACT", contractIDStr, pt, pred.PredictionID)
		}
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting contract compliance prediction: %w", err)
	}

	prevStatus := "DRAFT"
	auditNotes := "Initial contract compliance & documentation risk forecast published"
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

func (s *service) GetOrPredictShipmentReadiness(ctx context.Context, orgID int64, userID *int64, shipmentID int64, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	if shipmentID <= 0 {
		return nil, ErrMissingRecordID
	}

	shipmentIDStr := strconv.FormatInt(shipmentID, 10)
	now := time.Now().UTC()

	targetPredType := "SHIPMENT_READINESS_RISK"
	idempotencyKey := fmt.Sprintf("pred:shipments:readiness:%d:%d:%s", orgID, shipmentID, now.Format("2006-01-02-15"))

	readinessPredictionTypes := []string{
		"SHIPMENT_READINESS_RISK",
		"DOCUMENTATION_DELAY_RISK",
		"CUSTOMS_PROCESSING_RISK",
		"BILLING_READINESS_RISK",
		"CUTOFF_MISS_RISK",
		"HISTORICAL_OPERATIONAL_RISK",
	}

	if !forceRefresh {
		for _, pt := range readinessPredictionTypes {
			if existing, err := s.repo.GetActivePrediction(ctx, orgID, "shipments", "SHIPMENT", shipmentIDStr, pt); err == nil && existing != nil {
				return existing, nil
			}
		}
	} else {
		idempotencyKey = fmt.Sprintf("%s:refresh:%d", idempotencyKey, time.Now().UnixNano())
	}

	if s.shipmentReadinessData == nil {
		return nil, errors.New("shipment readiness data provider not initialized")
	}

	rData, err := s.shipmentReadinessData.FetchShipmentReadinessData(ctx, orgID, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching shipment readiness data for shipment %d: %w", shipmentID, err)
	}

	// Determine specific prediction type based on verified operational facts
	statusUpper := strings.ToUpper(rData.Status)
	if statusUpper == "CUSTOMS_HOLD" {
		targetPredType = "CUSTOMS_PROCESSING_RISK"
	} else if len(rData.OpenDiscrepancies) > 0 {
		targetPredType = "DOCUMENTATION_DELAY_RISK"
	} else if (statusUpper == "IN_TRANSIT" || statusUpper == "DEPARTED") && len(rData.MissingDocuments) > 0 {
		targetPredType = "BILLING_READINESS_RISK"
	} else {
		targetPredType = "SHIPMENT_READINESS_RISK"
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "shipments",
		PredictionType:    targetPredType,
		RelatedRecordType: "SHIPMENT",
		RelatedRecordID:   shipmentIDStr,
		RecordContext:     rData.ContextMap,
		TimeHorizon:       "14_DAYS",
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar shipment readiness prediction failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "shipments",
		PredictionType:      resp.PredictionType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "SHIPMENT",
		RelatedRecordID:     shipmentIDStr,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		DocumentReference:   resp.DocumentReference,
		ClauseReference:     resp.ClauseReference,
		PageNumber:          resp.PageNumber,
		SectionHeading:      resp.SectionHeading,
		MilestoneReference:  resp.MilestoneReference,
		CutoffReference:     resp.CutoffReference,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
		InsufficientData:    resp.InsufficientData,
		InsufficientReason:  resp.InsufficientReason,
	}

	// Ensure milestone and cutoff references are attached to source references
	if resp.MilestoneReference != nil && len(pred.SourceReferences) > 0 {
		for i := range pred.SourceReferences {
			if pred.SourceReferences[i].MilestoneReference == "" {
				pred.SourceReferences[i].MilestoneReference = *resp.MilestoneReference
			}
			if resp.CutoffReference != nil && pred.SourceReferences[i].CutoffReference == "" {
				pred.SourceReferences[i].CutoffReference = *resp.CutoffReference
			}
		}
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	if forceRefresh {
		for _, pt := range readinessPredictionTypes {
			_ = s.repo.SupersedeExisting(ctx, orgID, "shipments", "SHIPMENT", shipmentIDStr, pt, pred.PredictionID)
		}
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting shipment readiness prediction: %w", err)
	}

	prevStatus := "DRAFT"
	auditNotes := "Initial predictive documentation & shipment readiness intelligence published"
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

// GetOrPredictCarrierPerformance evaluates predictive service, delay, and reliability risk for ocean/air carriers
func (s *service) GetOrPredictCarrierPerformance(ctx context.Context, orgID int64, userID *int64, scac string, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	cleanSCAC := strings.ToUpper(strings.TrimSpace(scac))
	if cleanSCAC == "" {
		return nil, ErrMissingRecordID
	}

	if s.networkPerfData == nil {
		return nil, fmt.Errorf("network performance data provider not configured")
	}

	cData, err := s.networkPerfData.FetchCarrierData(ctx, orgID, cleanSCAC)
	if err != nil {
		return nil, fmt.Errorf("failed fetching carrier metrics for %s: %w", cleanSCAC, err)
	}

	targetPredType := "CARRIER_PERFORMANCE_RISK"
	if cData.DelayedShipments > 0 {
		targetPredType = "CARRIER_DELAY_RISK"
	}

	idempotencyKey := fmt.Sprintf("pred-carrier-%d-%s-%s", orgID, cleanSCAC, targetPredType)
	now := time.Now().UTC()

	if !forceRefresh {
		existing, err := s.repo.GetByIdempotencyKey(ctx, orgID, idempotencyKey)
		if err == nil && existing != nil {
			if existing.Status != StatusDismissed && existing.Status != StatusExpired && (existing.ExpiresAt == nil || existing.ExpiresAt.After(now)) {
				existing.UnpackJSON()
				return existing, nil
			}
		}
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "carriers",
		PredictionType:    targetPredType,
		RelatedRecordType: "CARRIER",
		RelatedRecordID:   cleanSCAC,
		RecordContext:     cData.ContextMap,
		TimeHorizon:       "14_DAYS",
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar carrier performance prediction failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "carriers",
		PredictionType:      resp.PredictionType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "CARRIER",
		RelatedRecordID:     cleanSCAC,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		ComparisonPeriod:    resp.ComparisonPeriod,
		SampleSize:          resp.SampleSize,
		CarrierReference:    resp.CarrierReference,
		LaneReference:       resp.LaneReference,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
		InsufficientData:    resp.InsufficientData,
		InsufficientReason:  resp.InsufficientReason,
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	if forceRefresh {
		carrierPredTypes := []string{"CARRIER_PERFORMANCE_RISK", "CARRIER_DELAY_RISK", "CARRIER_RELIABILITY_DROP"}
		for _, pt := range carrierPredTypes {
			_ = s.repo.SupersedeExisting(ctx, orgID, "carriers", "CARRIER", cleanSCAC, pt, pred.PredictionID)
		}
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting carrier performance prediction: %w", err)
	}

	prevStatus := "DRAFT"
	auditNotes := "Initial carrier performance intelligence published"
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

// GetOrPredictLanePerformance evaluates predictive corridor risk, customs dwell, and bottleneck pressure
func (s *service) GetOrPredictLanePerformance(ctx context.Context, orgID int64, userID *int64, laneCode string, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	cleanLane := strings.ToUpper(strings.TrimSpace(laneCode))
	if cleanLane == "" {
		return nil, ErrMissingRecordID
	}

	if s.networkPerfData == nil {
		return nil, fmt.Errorf("network performance data provider not configured")
	}

	lData, err := s.networkPerfData.FetchLaneData(ctx, orgID, cleanLane)
	if err != nil {
		return nil, fmt.Errorf("failed fetching lane metrics for %s: %w", cleanLane, err)
	}

	targetPredType := "LANE_PERFORMANCE_RISK"
	idempotencyKey := fmt.Sprintf("pred-lane-%d-%s-%s", orgID, cleanLane, targetPredType)
	now := time.Now().UTC()

	if !forceRefresh {
		existing, err := s.repo.GetByIdempotencyKey(ctx, orgID, idempotencyKey)
		if err == nil && existing != nil {
			if existing.Status != StatusDismissed && existing.Status != StatusExpired && (existing.ExpiresAt == nil || existing.ExpiresAt.After(now)) {
				existing.UnpackJSON()
				return existing, nil
			}
		}
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "network",
		PredictionType:    targetPredType,
		RelatedRecordType: "LANE",
		RelatedRecordID:   cleanLane,
		RecordContext:     lData.ContextMap,
		TimeHorizon:       "14_DAYS",
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar lane performance prediction failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "network",
		PredictionType:      resp.PredictionType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "LANE",
		RelatedRecordID:     cleanLane,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		ComparisonPeriod:    resp.ComparisonPeriod,
		SampleSize:          resp.SampleSize,
		LaneReference:       resp.LaneReference,
		CarrierReference:    resp.CarrierReference,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
		InsufficientData:    resp.InsufficientData,
		InsufficientReason:  resp.InsufficientReason,
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	if forceRefresh {
		lanePredTypes := []string{"LANE_PERFORMANCE_RISK", "LANE_DISRUPTION_RISK", "NETWORK_BOTTLENECK_RISK"}
		for _, pt := range lanePredTypes {
			_ = s.repo.SupersedeExisting(ctx, orgID, "network", "LANE", cleanLane, pt, pred.PredictionID)
		}
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting lane performance prediction: %w", err)
	}

	prevStatus := "DRAFT"
	auditNotes := "Initial network lane performance intelligence published"
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

// GetOrPredictCustomerServicePerformance evaluates predictive customer service risk, SLA health, and relationship trajectory
func (s *service) GetOrPredictCustomerServicePerformance(ctx context.Context, orgID int64, userID *int64, customerID int64, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	if customerID <= 0 {
		return nil, ErrMissingRecordID
	}

	if s.networkPerfData == nil {
		return nil, fmt.Errorf("network performance data provider not configured")
	}

	// Enforce strict tenant isolation via MariaDB query in provider
	csData, err := s.networkPerfData.FetchCustomerServiceData(ctx, orgID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching customer service metrics for %d: %w", customerID, err)
	}

	customerIDStr := strconv.FormatInt(customerID, 10)
	targetPredType := "CUSTOMER_RELATIONSHIP_RISK"
	if csData.ActiveExceptions > 0 || csData.CustomsHoldsCount > 0 || csData.HealthScore < 75 {
		targetPredType = "CUSTOMER_SERVICE_RISK"
	}

	idempotencyKey := fmt.Sprintf("pred-custserv-%d-%s-%s", orgID, customerIDStr, targetPredType)
	now := time.Now().UTC()

	if !forceRefresh {
		existing, err := s.repo.GetByIdempotencyKey(ctx, orgID, idempotencyKey)
		if err == nil && existing != nil {
			if existing.Status != StatusDismissed && existing.Status != StatusExpired && (existing.ExpiresAt == nil || existing.ExpiresAt.After(now)) {
				existing.UnpackJSON()
				return existing, nil
			}
		}
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "customers",
		PredictionType:    targetPredType,
		RelatedRecordType: "CUSTOMER",
		RelatedRecordID:   customerIDStr,
		RecordContext:     csData.ContextMap,
		TimeHorizon:       "30_DAYS",
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar customer service prediction failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "customers",
		PredictionType:      resp.PredictionType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "CUSTOMER",
		RelatedRecordID:     customerIDStr,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		ComparisonPeriod:    resp.ComparisonPeriod,
		SampleSize:          resp.SampleSize,
		CustomerReference:   resp.CustomerReference,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
		InsufficientData:    resp.InsufficientData,
		InsufficientReason:  resp.InsufficientReason,
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	if forceRefresh {
		custPredTypes := []string{"CUSTOMER_SERVICE_RISK", "CUSTOMER_RELATIONSHIP_RISK", "CUSTOMER_CHURN_RISK"}
		for _, pt := range custPredTypes {
			_ = s.repo.SupersedeExisting(ctx, orgID, "customers", "CUSTOMER", customerIDStr, pt, pred.PredictionID)
		}
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting customer service prediction: %w", err)
	}

	prevStatus := "DRAFT"
	auditNotes := "Initial customer service & relationship intelligence published"
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

func (s *service) GetOrPredictWorkloadPlanning(ctx context.Context, orgID int64, userID *int64, workloadType string, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	if workloadType == "" {
		workloadType = "approvals"
	}
	normType := strings.ToLower(strings.TrimSpace(workloadType))

	targetPredType := "APPROVAL_WORKLOAD_SPIKE"
	var ctxMap map[string]interface{}
	var err error

	if normType == "documentation" || normType == "docs" || normType == "shipments" {
		targetPredType = "DOCUMENTATION_WORKLOAD_SPIKE"
		if s.workloadCapacityData != nil {
			ctxMap, err = s.workloadCapacityData.FetchDocumentationWorkloadData(ctx, orgID)
		} else {
			p := NewWorkloadCapacityDataProvider(nil)
			ctxMap, err = p.FetchDocumentationWorkloadData(ctx, orgID)
		}
	} else {
		if s.workloadCapacityData != nil {
			ctxMap, err = s.workloadCapacityData.FetchApprovalWorkloadData(ctx, orgID)
		} else {
			p := NewWorkloadCapacityDataProvider(nil)
			ctxMap, err = p.FetchApprovalWorkloadData(ctx, orgID)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed fetching workload telemetry: %w", err)
	}

	idempotencyKey := fmt.Sprintf("pred-workload-%d-%s-%s", orgID, normType, targetPredType)
	now := time.Now().UTC()

	if !forceRefresh {
		existing, err := s.repo.GetByIdempotencyKey(ctx, orgID, idempotencyKey)
		if err == nil && existing != nil {
			if existing.Status != StatusDismissed && existing.Status != StatusExpired && (existing.ExpiresAt == nil || existing.ExpiresAt.After(now)) {
				existing.UnpackJSON()
				return existing, nil
			}
		}
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "workload",
		PredictionType:    targetPredType,
		RelatedRecordType: "WORKLOAD",
		RelatedRecordID:   normType,
		RecordContext:     ctxMap,
		TimeHorizon:       "7_DAYS",
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar workload prediction failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "workload",
		PredictionType:      resp.PredictionType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "WORKLOAD",
		RelatedRecordID:     normType,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		WorkloadType:        resp.WorkloadType,
		PendingCount:        resp.PendingCount,
		CapacityLimit:       resp.CapacityLimit,
		UtilizationRate:     resp.UtilizationRate,
		ComparisonPeriod:    resp.ComparisonPeriod,
		SampleSize:          resp.SampleSize,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
		InsufficientData:    resp.InsufficientData,
		InsufficientReason:  resp.InsufficientReason,
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	if forceRefresh {
		_ = s.repo.SupersedeExisting(ctx, orgID, "workload", "WORKLOAD", normType, targetPredType, pred.PredictionID)
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting workload prediction: %w", err)
	}

	prevStatus := "DRAFT"
	auditNotes := "Initial workload planning intelligence published"
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

func (s *service) GetOrPredictCapacityPlanning(ctx context.Context, orgID int64, userID *int64, dimension string, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	if dimension == "" {
		dimension = "corridors"
	}
	normDim := strings.ToLower(strings.TrimSpace(dimension))

	var ctxMap map[string]interface{}
	var err error
	if s.workloadCapacityData != nil {
		ctxMap, err = s.workloadCapacityData.FetchCorridorCapacityData(ctx, orgID, normDim)
	} else {
		p := NewWorkloadCapacityDataProvider(nil)
		ctxMap, err = p.FetchCorridorCapacityData(ctx, orgID, normDim)
	}
	if err != nil {
		return nil, fmt.Errorf("failed fetching capacity telemetry: %w", err)
	}

	targetPredType := "DEMAND_CAPACITY_MISMATCH"
	idempotencyKey := fmt.Sprintf("pred-capacity-%d-%s-%s", orgID, normDim, targetPredType)
	now := time.Now().UTC()

	if !forceRefresh {
		existing, err := s.repo.GetByIdempotencyKey(ctx, orgID, idempotencyKey)
		if err == nil && existing != nil {
			if existing.Status != StatusDismissed && existing.Status != StatusExpired && (existing.ExpiresAt == nil || existing.ExpiresAt.After(now)) {
				existing.UnpackJSON()
				return existing, nil
			}
		}
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "capacity",
		PredictionType:    targetPredType,
		RelatedRecordType: "CAPACITY",
		RelatedRecordID:   normDim,
		RecordContext:     ctxMap,
		TimeHorizon:       "14_DAYS",
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar capacity prediction failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "capacity",
		PredictionType:      resp.PredictionType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "CAPACITY",
		RelatedRecordID:     normDim,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		LaneReference:       resp.LaneReference,
		WorkloadType:        resp.WorkloadType,
		UtilizationRate:     resp.UtilizationRate,
		ComparisonPeriod:    resp.ComparisonPeriod,
		SampleSize:          resp.SampleSize,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
		InsufficientData:    resp.InsufficientData,
		InsufficientReason:  resp.InsufficientReason,
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	if forceRefresh {
		_ = s.repo.SupersedeExisting(ctx, orgID, "capacity", "CAPACITY", normDim, targetPredType, pred.PredictionID)
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting capacity prediction: %w", err)
	}

	prevStatus := "DRAFT"
	auditNotes := "Initial capacity planning intelligence published"
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

func (s *service) GetOrPredictDemandPlanning(ctx context.Context, orgID int64, userID *int64, segment string, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}
	if segment == "" {
		segment = "commercial"
	}
	normSeg := strings.ToLower(strings.TrimSpace(segment))

	var ctxMap map[string]interface{}
	var err error
	if s.workloadCapacityData != nil {
		ctxMap, err = s.workloadCapacityData.FetchQuoteDemandData(ctx, orgID)
	} else {
		p := NewWorkloadCapacityDataProvider(nil)
		ctxMap, err = p.FetchQuoteDemandData(ctx, orgID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed fetching demand telemetry: %w", err)
	}

	targetPredType := "QUOTE_PROCESSING_BOTTLENECK"
	idempotencyKey := fmt.Sprintf("pred-demand-%d-%s-%s", orgID, normSeg, targetPredType)
	now := time.Now().UTC()

	if !forceRefresh {
		existing, err := s.repo.GetByIdempotencyKey(ctx, orgID, idempotencyKey)
		if err == nil && existing != nil {
			if existing.Status != StatusDismissed && existing.Status != StatusExpired && (existing.ExpiresAt == nil || existing.ExpiresAt.After(now)) {
				existing.UnpackJSON()
				return existing, nil
			}
		}
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "demand",
		PredictionType:    targetPredType,
		RelatedRecordType: "DEMAND",
		RelatedRecordID:   normSeg,
		RecordContext:     ctxMap,
		TimeHorizon:       "14_DAYS",
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar demand prediction failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "demand",
		PredictionType:      resp.PredictionType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "DEMAND",
		RelatedRecordID:     normSeg,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		CustomerReference:   resp.CustomerReference,
		WorkloadType:        resp.WorkloadType,
		PendingCount:        resp.PendingCount,
		ComparisonPeriod:    resp.ComparisonPeriod,
		SampleSize:          resp.SampleSize,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
		InsufficientData:    resp.InsufficientData,
		InsufficientReason:  resp.InsufficientReason,
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	if forceRefresh {
		_ = s.repo.SupersedeExisting(ctx, orgID, "demand", "DEMAND", normSeg, targetPredType, pred.PredictionID)
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting demand prediction: %w", err)
	}

	prevStatus := "DRAFT"
	auditNotes := "Initial demand planning intelligence published"
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

func (s *service) GetOrPredictOperationalBottleneck(ctx context.Context, orgID int64, userID *int64, bottleneckType string, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}

	normType := strings.ToLower(strings.TrimSpace(bottleneckType))
	if normType == "" {
		normType = "approvals"
	}

	p := s.resourceBottleneckData
	if p == nil {
		p = NewResourceBottleneckDataProvider(nil)
	}

	var ctxMap map[string]interface{}
	var err error
	var targetPredType string

	switch normType {
	case "documentation", "docs":
		normType = "documentation"
		targetPredType = "DOCUMENTATION_BOTTLENECK"
		ctxMap, err = p.FetchDocumentationBottleneckData(ctx, orgID)
	case "exceptions", "cross_module", "cross":
		normType = "exceptions"
		targetPredType = "CROSS_MODULE_BOTTLENECK"
		ctxMap, err = p.FetchExceptionCrossModuleBottleneckData(ctx, orgID)
	default:
		normType = "approvals"
		targetPredType = "APPROVAL_BOTTLENECK"
		ctxMap, err = p.FetchApprovalBottleneckData(ctx, orgID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed fetching operational bottleneck telemetry: %w", err)
	}

	idempotencyKey := fmt.Sprintf("pred-btln-%d-%s-%s", orgID, normType, targetPredType)
	now := time.Now().UTC()

	if !forceRefresh {
		existing, err := s.repo.GetByIdempotencyKey(ctx, orgID, idempotencyKey)
		if err == nil && existing != nil {
			if existing.Status != StatusDismissed && existing.Status != StatusExpired && (existing.ExpiresAt == nil || existing.ExpiresAt.After(now)) {
				existing.UnpackJSON()
				return existing, nil
			}
		}
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "bottleneck",
		PredictionType:    targetPredType,
		RelatedRecordType: "BOTTLENECK",
		RelatedRecordID:   normType,
		RecordContext:     ctxMap,
		TimeHorizon:       "7_DAYS",
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar operational bottleneck prediction failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "bottleneck",
		PredictionType:      resp.PredictionType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "BOTTLENECK",
		RelatedRecordID:     normType,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		BottleneckType:      resp.BottleneckType,
		AffectedStage:       resp.AffectedStage,
		AssignedOwner:       resp.AssignedOwner,
		QueueDwellHours:     resp.QueueDwellHours,
		PendingCount:        resp.PendingCount,
		CapacityLimit:       resp.CapacityLimit,
		ComparisonPeriod:    resp.ComparisonPeriod,
		SampleSize:          resp.SampleSize,
		CarrierReference:    resp.CarrierReference,
		LinkedExceptionID:   resp.LinkedExceptionID,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
		InsufficientData:    resp.InsufficientData,
		InsufficientReason:  resp.InsufficientReason,
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	if forceRefresh {
		_ = s.repo.SupersedeExisting(ctx, orgID, "bottleneck", "BOTTLENECK", normType, targetPredType, pred.PredictionID)
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting operational bottleneck prediction: %w", err)
	}

	prevStatus := "DRAFT"
	auditNotes := "Initial operational bottleneck intelligence published"
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

func (s *service) GetOrPredictResourceAllocation(ctx context.Context, orgID int64, userID *int64, resourceType string, forceRefresh bool) (*Prediction, error) {
	if orgID <= 0 {
		return nil, ErrInvalidOrgID
	}

	normType := strings.ToLower(strings.TrimSpace(resourceType))
	if normType == "" {
		normType = "allocation"
	}

	p := s.resourceBottleneckData
	if p == nil {
		p = NewResourceBottleneckDataProvider(nil)
	}

	ctxMap, err := p.FetchResourceAllocationData(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching resource allocation telemetry: %w", err)
	}

	targetPredType := "RESOURCE_ALLOCATION_IMBALANCE"
	idempotencyKey := fmt.Sprintf("pred-rsrc-%d-%s-%s", orgID, normType, targetPredType)
	now := time.Now().UTC()

	if !forceRefresh {
		existing, err := s.repo.GetByIdempotencyKey(ctx, orgID, idempotencyKey)
		if err == nil && existing != nil {
			if existing.Status != StatusDismissed && existing.Status != StatusExpired && (existing.ExpiresAt == nil || existing.ExpiresAt.After(now)) {
				existing.UnpackJSON()
				return existing, nil
			}
		}
	}

	sidecarReq := &GenerateRequest{
		OrgID:             orgID,
		Module:            "resource",
		PredictionType:    targetPredType,
		RelatedRecordType: "RESOURCE",
		RelatedRecordID:   normType,
		RecordContext:     ctxMap,
		TimeHorizon:       "14_DAYS",
	}

	resp, err := s.sidecar.GeneratePrediction(ctx, sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("sidecar resource allocation prediction failed: %w", err)
	}

	if err := s.validateResponse(sidecarReq, resp); err != nil {
		return nil, fmt.Errorf("sidecar response validation failed: %w", err)
	}

	expiresAt := now.Add(24 * time.Hour)
	sourceTime := now
	if t, err := time.Parse(time.RFC3339, resp.SourceTimestamp); err == nil {
		sourceTime = t
	}

	pred := &Prediction{
		OrgID:               orgID,
		UserID:              userID,
		PredictionID:        resp.PredictionID,
		IdempotencyKey:      idempotencyKey,
		Module:              "resource",
		PredictionType:      resp.PredictionType,
		Status:              StatusPublished,
		Severity:            Severity(resp.Severity),
		ConfidenceScore:     resp.ConfidenceScore,
		ConfidenceBand:      ConfidenceBand(resp.ConfidenceBand),
		RelatedRecordType:   "RESOURCE",
		RelatedRecordID:     normType,
		PredictionStatement: resp.PredictionStatement,
		PredictedValue:      resp.PredictedValue,
		TimeHorizon:         resp.TimeHorizon,
		Explanation:         resp.Explanation,
		SupportingSignals:   resp.SupportingSignals,
		SourceReferences:    resp.SourceReferences,
		SourceTimestamp:     sourceTime,
		RecommendedAction:   resp.RecommendedAction,
		ActionType:          resp.ActionType,
		IsActionRequired:    resp.IsActionRequired,
		RequiresApproval:    resp.RequiresApproval,
		BottleneckType:      resp.BottleneckType,
		AffectedStage:       resp.AffectedStage,
		AssignedOwner:       resp.AssignedOwner,
		PendingCount:        resp.PendingCount,
		UtilizationRate:     resp.UtilizationRate,
		ComparisonPeriod:    resp.ComparisonPeriod,
		SampleSize:          resp.SampleSize,
		ReviewStatus:        "PENDING",
		ActualOutcomeStatus: OutcomePending,
		ModelVersion:        resp.ModelVersion,
		ExpiresAt:           &expiresAt,
		InsufficientData:    resp.InsufficientData,
		InsufficientReason:  resp.InsufficientReason,
	}

	if resp.InsufficientData {
		pred.ReviewStatus = "INSUFFICIENT_DATA"
	}

	if forceRefresh {
		_ = s.repo.SupersedeExisting(ctx, orgID, "resource", "RESOURCE", normType, targetPredType, pred.PredictionID)
	}

	if err := s.repo.Create(ctx, pred); err != nil {
		return nil, fmt.Errorf("failed persisting resource allocation prediction: %w", err)
	}

	prevStatus := "DRAFT"
	auditNotes := "Initial resource allocation intelligence published"
	_ = s.repo.RecordAudit(ctx, &PredictionAuditHistory{
		PredictionID:   pred.PredictionID,
		OrgID:          orgID,
		UserID:         userID,
		PreviousStatus: &prevStatus,
		NewStatus:      string(pred.Status),
		Action:         "PREDICTION_PUBLISHED",
		Notes:          &auditNotes,
	})

	return pred, nil
}

func (s *service) GetResourceBottleneckSummary(ctx context.Context, orgID int64, userID *int64) (map[string]interface{}, error) {
	apprPred, errAppr := s.GetOrPredictOperationalBottleneck(ctx, orgID, userID, "approvals", false)
	docsPred, errDocs := s.GetOrPredictOperationalBottleneck(ctx, orgID, userID, "documentation", false)
	crossPred, errCross := s.GetOrPredictOperationalBottleneck(ctx, orgID, userID, "exceptions", false)
	rsrcPred, errRsrc := s.GetOrPredictResourceAllocation(ctx, orgID, userID, "allocation", false)

	res := map[string]interface{}{
		"org_id":       orgID,
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	}

	if errAppr == nil && apprPred != nil {
		res["approvals_bottleneck"] = apprPred
	}
	if errDocs == nil && docsPred != nil {
		res["documentation_bottleneck"] = docsPred
	}
	if errCross == nil && crossPred != nil {
		res["exception_bottleneck"] = crossPred
	}
	if errRsrc == nil && rsrcPred != nil {
		res["resource_allocation"] = rsrcPred
	}

	return res, nil
}





