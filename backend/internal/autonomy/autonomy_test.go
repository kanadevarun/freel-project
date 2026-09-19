package autonomy_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/freel/backend/internal/autonomy"
	"github.com/stretchr/testify/assert"
)

// MockAutonomyRepo implements autonomy.Repository in-memory for testing
type MockAutonomyRepo struct {
	plans           map[string]*autonomy.AutonomousPlan
	steps           map[string][]autonomy.AutonomousPlanStep
	policies        map[string]*autonomy.AutonomyPolicy
	audits          map[string][]autonomy.AutonomousPlanAuditHistory
	preferences     map[string]*autonomy.CustomerCommunicationPreferences
	followupRecords map[string]*autonomy.CustomerFollowupRecord
	financePlans    map[string]*autonomy.FinanceCollectionPlan
	compliancePlans map[string]*autonomy.ContractComplianceMonitoringPlan
	exceptionPlans  map[string]*autonomy.ExceptionResolutionPlan
	decisions       map[string]*autonomy.HumanAIDecision
	govLimits       map[int64]*autonomy.TenantGovernanceLimits
	allowlist       map[string]*autonomy.ActionAllowlistItem
	featureFlags    map[string]*autonomy.GovernanceFeatureFlag
}

func NewMockAutonomyRepo() *MockAutonomyRepo {
	m := &MockAutonomyRepo{
		plans:           make(map[string]*autonomy.AutonomousPlan),
		steps:           make(map[string][]autonomy.AutonomousPlanStep),
		policies:        make(map[string]*autonomy.AutonomyPolicy),
		audits:          make(map[string][]autonomy.AutonomousPlanAuditHistory),
		preferences:     make(map[string]*autonomy.CustomerCommunicationPreferences),
		followupRecords: make(map[string]*autonomy.CustomerFollowupRecord),
		financePlans:    make(map[string]*autonomy.FinanceCollectionPlan),
		compliancePlans: make(map[string]*autonomy.ContractComplianceMonitoringPlan),
		exceptionPlans:  make(map[string]*autonomy.ExceptionResolutionPlan),
		decisions:       make(map[string]*autonomy.HumanAIDecision),
		govLimits:       make(map[int64]*autonomy.TenantGovernanceLimits),
		allowlist:       make(map[string]*autonomy.ActionAllowlistItem),
		featureFlags:    make(map[string]*autonomy.GovernanceFeatureFlag),
	}

	defActions := []autonomy.ActionAllowlistItem{
		{
			OrgID:                 1,
			ActionType:            "shipments.update_milestone",
			ActionName:            "Update Milestone",
			Module:                "shipments",
			RiskClass:             autonomy.RiskClassLow,
			AllowedAutonomyLevels: []byte("[1,2,3,4]"),
			ApprovalRequirement:   autonomy.ApprovalNever,
			Reversibility:         autonomy.ReversibilityReversible,
			IsEnabled:             true,
		},
		{
			OrgID:                 1,
			ActionType:            "finance.apply_invoice_discount",
			ActionName:            "Apply Invoice Discount",
			Module:                "finance",
			RiskClass:             autonomy.RiskClassHigh,
			AllowedAutonomyLevels: []byte("[2,3]"),
			ApprovalRequirement:   autonomy.ApprovalAlways,
			Reversibility:         autonomy.ReversibilityIrreversible,
			IsEnabled:             true,
		},
		{
			OrgID:                 1,
			ActionType:            "compliance.record_sanctions_check",
			ActionName:            "Record Sanctions Check",
			Module:                "compliance",
			RiskClass:             autonomy.RiskClassCritical,
			AllowedAutonomyLevels: []byte("[2,3]"),
			ApprovalRequirement:   autonomy.ApprovalAlways,
			Reversibility:         autonomy.ReversibilityIrreversible,
			IsEnabled:             true,
		},
		{
			OrgID:                 1,
			ActionType:            "shipments.reroute",
			ActionName:            "Execute Shipment Reroute",
			Module:                "shipments",
			RiskClass:             autonomy.RiskClassHigh,
			AllowedAutonomyLevels: []byte("[2,3,4]"),
			ApprovalRequirement:   autonomy.ApprovalAlways,
			Reversibility:         autonomy.ReversibilityIrreversible,
			IsEnabled:             true,
		},
	}
	for i := range defActions {
		key := fmt.Sprintf("%d:%s", defActions[i].OrgID, defActions[i].ActionType)
		m.allowlist[key] = &defActions[i]
	}

	defFlags := []autonomy.GovernanceFeatureFlag{
		{OrgID: 1, FlagKey: "shipment_automation", FlagName: "Shipment Automation", IsEnabled: true, MaxAutonomyLevel: 4, RequiresApproval: true},
		{OrgID: 1, FlagKey: "customer_automation", FlagName: "Customer Automation", IsEnabled: true, MaxAutonomyLevel: 3, RequiresApproval: true},
		{OrgID: 1, FlagKey: "finance_automation", FlagName: "Finance Automation", IsEnabled: true, MaxAutonomyLevel: 3, RequiresApproval: true},
		{OrgID: 1, FlagKey: "contract_compliance", FlagName: "Contract Compliance", IsEnabled: true, MaxAutonomyLevel: 3, RequiresApproval: true},
	}
	for i := range defFlags {
		key := fmt.Sprintf("%d:%s", defFlags[i].OrgID, defFlags[i].FlagKey)
		m.featureFlags[key] = &defFlags[i]
	}

	return m
}

func (m *MockAutonomyRepo) GetPolicy(ctx context.Context, orgID int64, module string) (*autonomy.AutonomyPolicy, error) {
	key := fmt.Sprintf("%d:%s", orgID, module)
	if p, ok := m.policies[key]; ok {
		return p, nil
	}
	defKey := fmt.Sprintf("%d:general", orgID)
	if p, ok := m.policies[defKey]; ok {
		return p, nil
	}
	sysKey := "0:general"
	if p, ok := m.policies[sysKey]; ok {
		return p, nil
	}
	return nil, autonomy.ErrPolicyNotFound
}

func (m *MockAutonomyRepo) SetPolicy(ctx context.Context, policy *autonomy.AutonomyPolicy) error {
	key := fmt.Sprintf("%d:%s", policy.OrgID, policy.Module)
	m.policies[key] = policy
	return nil
}

func (m *MockAutonomyRepo) ListPolicies(ctx context.Context, orgID int64) ([]autonomy.AutonomyPolicy, error) {
	var list []autonomy.AutonomyPolicy
	for _, p := range m.policies {
		if p.OrgID == orgID || p.OrgID == 0 {
			list = append(list, *p)
		}
	}
	return list, nil
}

func (m *MockAutonomyRepo) CreatePlan(ctx context.Context, plan *autonomy.AutonomousPlan, steps []autonomy.AutonomousPlanStep) error {
	key := fmt.Sprintf("%d:%s", plan.OrgID, plan.PlanID)
	m.plans[key] = plan
	m.steps[key] = steps
	return nil
}

func (m *MockAutonomyRepo) GetPlan(ctx context.Context, orgID int64, planID string) (*autonomy.AutonomousPlan, []autonomy.AutonomousPlanStep, error) {
	key := fmt.Sprintf("%d:%s", orgID, planID)
	p, ok := m.plans[key]
	if !ok {
		return nil, nil, autonomy.ErrPlanNotFound
	}
	return p, m.steps[key], nil
}

func (m *MockAutonomyRepo) ListPlans(ctx context.Context, orgID int64, module, status string, limit, offset int) ([]autonomy.AutonomousPlan, int, error) {
	var list []autonomy.AutonomousPlan
	for _, p := range m.plans {
		if p.OrgID == orgID {
			if module != "" && p.Module != module {
				continue
			}
			if status != "" && string(p.Status) != status {
				continue
			}
			list = append(list, *p)
		}
	}
	return list, len(list), nil
}

func (m *MockAutonomyRepo) GetPlanVersions(ctx context.Context, orgID int64, planID string) ([]autonomy.AutonomousPlan, error) {
	var vers []autonomy.AutonomousPlan
	for _, p := range m.plans {
		if p.OrgID == orgID && (p.PlanID == planID || (p.ParentPlanID.Valid && p.ParentPlanID.String == planID)) {
			vers = append(vers, *p)
		}
	}
	return vers, nil
}

func (m *MockAutonomyRepo) UpdatePlanStatus(ctx context.Context, orgID int64, planID string, status autonomy.PlanStatus, execStatus string, notes *string) error {
	key := fmt.Sprintf("%d:%s", orgID, planID)
	p, ok := m.plans[key]
	if !ok {
		return autonomy.ErrPlanNotFound
	}
	p.Status = status
	p.ExecutionStatus = execStatus
	p.UpdatedAt = time.Now()
	return nil
}

func (m *MockAutonomyRepo) UpdateStepStatus(ctx context.Context, orgID int64, planID, stepID string, status autonomy.StepStatus, resultJSON, errJSON *string) error {
	key := fmt.Sprintf("%d:%s", orgID, planID)
	steps, ok := m.steps[key]
	if !ok {
		return autonomy.ErrPlanNotFound
	}
	for i := range steps {
		if steps[i].StepID == stepID {
			steps[i].Status = status
			if resultJSON != nil {
				steps[i].ExecutionResult = json.RawMessage(*resultJSON)
			}
			if errJSON != nil {
				steps[i].ErrorMessage = sql.NullString{String: *errJSON, Valid: true}
			}
			steps[i].UpdatedAt = time.Now()
			m.steps[key] = steps
			return nil
		}
	}
	return autonomy.ErrStepNotFound
}

func (m *MockAutonomyRepo) RecordAudit(ctx context.Context, entry *autonomy.AutonomousPlanAuditHistory) error {
	key := fmt.Sprintf("%d:%s", entry.OrgID, entry.PlanID)
	m.audits[key] = append(m.audits[key], *entry)
	return nil
}

func (m *MockAutonomyRepo) GetAuditHistory(ctx context.Context, orgID int64, planID string) ([]autonomy.AutonomousPlanAuditHistory, error) {
	key := fmt.Sprintf("%d:%s", orgID, planID)
	return m.audits[key], nil
}

func (m *MockAutonomyRepo) CreateGoal(ctx context.Context, goal *autonomy.PlanningGoal) error {
	return nil
}

func (m *MockAutonomyRepo) GetGoal(ctx context.Context, orgID int64, goalID string) (*autonomy.PlanningGoal, error) {
	return &autonomy.PlanningGoal{
		OrgID:  orgID,
		GoalID: goalID,
		Status: "ACTIVE",
	}, nil
}

func (m *MockAutonomyRepo) ListGoals(ctx context.Context, orgID int64, module, status string, limit, offset int) ([]autonomy.PlanningGoal, int, error) {
	return []autonomy.PlanningGoal{}, 0, nil
}

func (m *MockAutonomyRepo) UpdateGoalStatus(ctx context.Context, orgID int64, goalID string, status string) error {
	return nil
}

func (m *MockAutonomyRepo) UpdatePlanSelectedCandidate(ctx context.Context, orgID int64, planID, candidateID string) error {
	key := fmt.Sprintf("%d:%s", orgID, planID)
	if p, ok := m.plans[key]; ok {
		p.SelectedCandidateID = sql.NullString{String: candidateID, Valid: true}
	}
	return nil
}

func (m *MockAutonomyRepo) UpdatePlanStaleness(ctx context.Context, orgID int64, planID, stalenessStatus string) error {
	key := fmt.Sprintf("%d:%s", orgID, planID)
	if p, ok := m.plans[key]; ok {
		p.StalenessStatus = stalenessStatus
	}
	return nil
}

func (m *MockAutonomyRepo) SaveMemory(ctx context.Context, mem *autonomy.OperationalMemory) error {
	return nil
}

func (m *MockAutonomyRepo) GetMemories(ctx context.Context, orgID int64, entityType, entityID string) ([]autonomy.OperationalMemory, error) {
	return nil, nil
}

func (m *MockAutonomyRepo) RecordShipmentEvent(ctx context.Context, event *autonomy.ShipmentAdaptiveEvent) error {
	return nil
}

func (m *MockAutonomyRepo) GetShipmentEventByDedupKey(ctx context.Context, orgID int64, dedupKey string) (*autonomy.ShipmentAdaptiveEvent, error) {
	return nil, nil
}

func (m *MockAutonomyRepo) ListShipmentEvents(ctx context.Context, orgID int64, shipmentID int64, limit int) ([]autonomy.ShipmentAdaptiveEvent, error) {
	return []autonomy.ShipmentAdaptiveEvent{}, nil
}

func (m *MockAutonomyRepo) GetActivePlanForShipment(ctx context.Context, orgID int64, shipmentID int64) (*autonomy.AutonomousPlan, error) {
	for _, p := range m.plans {
		if p.OrgID == orgID && p.RelatedEntityType == "shipment" && p.RelatedEntityID == fmt.Sprintf("%d", shipmentID) {
			return p, nil
		}
	}
	return nil, nil
}

func (m *MockAutonomyRepo) UpdatePlanWaitingState(ctx context.Context, orgID int64, planID string, waitingState string, waitingUntil *time.Time) error {
	return nil
}

func (m *MockAutonomyRepo) UpdateShipmentAdaptiveMetrics(ctx context.Context, orgID int64, shipmentID int64, riskLevel string, adaptiveStatus string, commitmentDate *time.Time) error {
	return nil
}

func (m *MockAutonomyRepo) GetShipmentAdaptiveContext(ctx context.Context, orgID int64, shipmentID int64) (map[string]interface{}, []map[string]interface{}, []map[string]interface{}, error) {
	return map[string]interface{}{
		"id":         shipmentID,
		"org_id":     orgID,
		"status":     "IN_TRANSIT",
		"risk_level": "LOW",
	}, nil, nil, nil
}

func (m *MockAutonomyRepo) GetCustomerPreferences(ctx context.Context, orgID, customerID int64) (*autonomy.CustomerCommunicationPreferences, error) {
	key := fmt.Sprintf("%d:%d", orgID, customerID)
	if p, ok := m.preferences[key]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *MockAutonomyRepo) SaveCustomerPreferences(ctx context.Context, pref *autonomy.CustomerCommunicationPreferences) error {
	key := fmt.Sprintf("%d:%d", pref.OrgID, pref.CustomerID)
	m.preferences[key] = pref
	return nil
}

func (m *MockAutonomyRepo) CreateFollowupRecord(ctx context.Context, rec *autonomy.CustomerFollowupRecord) (*autonomy.CustomerFollowupRecord, error) {
	if rec.ID == 0 {
		rec.ID = int64(len(m.followupRecords) + 1001)
	}
	rec.CreatedAt = time.Now()
	rec.UpdatedAt = time.Now()
	idKey := fmt.Sprintf("%d:%d", rec.OrgID, rec.ID)
	m.followupRecords[idKey] = rec
	if rec.IdempotencyKey != "" {
		idempKey := fmt.Sprintf("%d:idemp:%s", rec.OrgID, rec.IdempotencyKey)
		m.followupRecords[idempKey] = rec
	}
	return rec, nil
}

func (m *MockAutonomyRepo) GetFollowupRecord(ctx context.Context, orgID, id int64) (*autonomy.CustomerFollowupRecord, error) {
	idKey := fmt.Sprintf("%d:%d", orgID, id)
	if r, ok := m.followupRecords[idKey]; ok {
		return r, nil
	}
	return nil, autonomy.ErrPlanNotFound
}

func (m *MockAutonomyRepo) GetFollowupRecordByIdempotency(ctx context.Context, orgID int64, idempKey string) (*autonomy.CustomerFollowupRecord, error) {
	key := fmt.Sprintf("%d:idemp:%s", orgID, idempKey)
	if r, ok := m.followupRecords[key]; ok {
		return r, nil
	}
	return nil, nil
}

func (m *MockAutonomyRepo) ListCustomerFollowupRecords(ctx context.Context, orgID, customerID int64, limit int) ([]autonomy.CustomerFollowupRecord, error) {
	return []autonomy.CustomerFollowupRecord{}, nil
}

func (m *MockAutonomyRepo) UpdateFollowupRecordStatus(ctx context.Context, orgID, id int64, status string, approvalID *string, sentAt *time.Time) error {
	return nil
}

func (m *MockAutonomyRepo) RecordCustomerResponse(ctx context.Context, orgID, id int64, response string, classification string, stopReason *string) error {
	return nil
}

func (m *MockAutonomyRepo) GetCustomerAuthoritativeContact(ctx context.Context, orgID, customerID int64, contactID *int64) (*autonomy.VerifiedContact, []autonomy.VerifiedContact, error) {
	c := &autonomy.VerifiedContact{
		ContactID: 10,
		FirstName: "Vikram",
		LastName:  "Malhotra",
		Email:     "v.malhotra@apexlogistics.com",
		IsPrimary: true,
	}
	return c, []autonomy.VerifiedContact{*c}, nil
}

func (m *MockAutonomyRepo) GetCustomerFollowupContext(ctx context.Context, orgID, customerID int64) (string, string, string, error) {
	return "Apex Logistics Inc", "ACTIVE", "ENTERPRISE", nil
}

func (m *MockAutonomyRepo) CountRecentFollowups(ctx context.Context, orgID, customerID int64, hours int) (int, *float64, error) {
	return 0, nil, nil
}

func (m *MockAutonomyRepo) GetActivePlanForCustomer(ctx context.Context, orgID, customerID int64) (*autonomy.AutonomousPlan, error) {
	return nil, nil
}

func (m *MockAutonomyRepo) UpdateCustomerFollowupStatus(ctx context.Context, orgID, customerID int64, status string, planID *string) error {
	return nil
}

func (m *MockAutonomyRepo) GetRfqPricingContext(ctx context.Context, orgID, rfqID int64) (*autonomy.RfqPricingContextDTO, error) {
	return &autonomy.RfqPricingContextDTO{
		OrgID:         orgID,
		RfqID:         rfqID,
		RfqNumber:     fmt.Sprintf("RFQ-2026-%d", rfqID),
		CustomerID:    101,
		CustomerName:  "Apex Global Logistics Corp",
		AccountTier:   "ENTERPRISE",
		Origin:        "Nhava Sheva (INNSA)",
		Destination:   "Rotterdam (NLRTM)",
		TransportMode: "OCEAN",
		Incoterms:     "FOB",
		RateBasis: map[string]interface{}{
			"carrier_name":  "Maersk Line",
			"base_cost":     2400.0,
			"surcharges":    300.0,
			"currency":      "USD",
			"rate_age_days": 5,
			"is_stale":      false,
		},
		PricingPolicy: map[string]interface{}{
			"min_margin_pct":         8.0,
			"target_margin_pct":      16.0,
			"max_monetary_threshold": 25000.0,
		},
	}, nil
}

func (m *MockAutonomyRepo) SaveRfqPricingOptimization(ctx context.Context, opt *autonomy.RfqPricingOptimization) error {
	if opt.ID == 0 {
		opt.ID = 1001
	}
	return nil
}

func (m *MockAutonomyRepo) GetRfqPricingOptimization(ctx context.Context, orgID, rfqID int64) (*autonomy.RfqPricingOptimization, error) {
	return nil, nil
}

func (m *MockAutonomyRepo) GetRfqPricingOptimizationByID(ctx context.Context, orgID, id int64) (*autonomy.RfqPricingOptimization, error) {
	return nil, nil
}

func (m *MockAutonomyRepo) SaveRfqPricingVersion(ctx context.Context, ver *autonomy.RfqPricingVersion) error {
	return nil
}

func (m *MockAutonomyRepo) ListRfqPricingVersions(ctx context.Context, orgID, optimizationID int64) ([]autonomy.RfqPricingVersion, error) {
	return []autonomy.RfqPricingVersion{}, nil
}

func (m *MockAutonomyRepo) UpdateRfqPricingOptimizationStrategy(ctx context.Context, orgID, rfqID int64, strategyID string, price, marginPct float64, requiresApproval bool, approvalReason *string) error {
	return nil
}

func (m *MockAutonomyRepo) UpdateRfqPricingOptimizationQuotation(ctx context.Context, orgID, rfqID int64, quotationID int64, status string) error {
	return nil
}

func (m *MockAutonomyRepo) CreateQuotationRecord(ctx context.Context, orgID, rfqID, customerID int64, rfqNumber, quotationNumber, customerName, origin, destination, transportMode, currency string, totalAmount, totalCost, grossMarginPct float64, status string) (int64, error) {
	return 2001, nil
}

// Phase 5 Task 5.6: Adaptive Finance and Collections Mock Repo
func (m *MockAutonomyRepo) GetFinanceInvoiceContext(ctx context.Context, orgID, invoiceID int64) (*autonomy.FinanceInvoiceContextDTO, error) {
	dueDate := "2026-08-15"
	bal := 4500.0
	status := "Overdue"
	isDisputed := false
	daysOverdue := 27
	aging := "16-30_DAYS"
	if invoiceID == 101 {
		bal = 0.0
		status = "Paid"
		daysOverdue = 0
		aging = "CURRENT"
	} else if invoiceID == 104 {
		isDisputed = true
		status = "Disputed"
	}
	return &autonomy.FinanceInvoiceContextDTO{
		OrgID:         orgID,
		InvoiceID:     invoiceID,
		InvoiceNumber: fmt.Sprintf("INV-2026-DEV-%03d", invoiceID),
		CustomerID:    102,
		CustomerName:  "Nordic Freight Dynamics AB",
		AccountTier:   "STANDARD",
		Currency:      "USD",
		TotalAmount:   4500.0,
		PaidAmount:    4500.0 - bal,
		BalanceDue:    bal,
		DueDate:       &dueDate,
		DaysOverdue:   daysOverdue,
		AgingBucket:   aging,
		InvoiceStatus: status,
		IsDisputed:    isDisputed,
	}, nil
}

func (m *MockAutonomyRepo) SaveFinanceCollectionPlan(ctx context.Context, plan *autonomy.FinanceCollectionPlan) error {
	if plan.ID == 0 {
		plan.ID = 5001
	}
	key := fmt.Sprintf("%d:%d", plan.OrgID, plan.InvoiceID)
	m.financePlans[key] = plan
	return nil
}

func (m *MockAutonomyRepo) GetFinanceCollectionPlan(ctx context.Context, orgID, invoiceID int64) (*autonomy.FinanceCollectionPlan, error) {
	key := fmt.Sprintf("%d:%d", orgID, invoiceID)
	if p, ok := m.financePlans[key]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *MockAutonomyRepo) GetFinanceCollectionPlanByID(ctx context.Context, orgID, id int64) (*autonomy.FinanceCollectionPlan, error) {
	for _, p := range m.financePlans {
		if p.OrgID == orgID && p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}

func (m *MockAutonomyRepo) SaveFinanceCollectionVersion(ctx context.Context, ver *autonomy.FinanceCollectionVersion) error {
	return nil
}

func (m *MockAutonomyRepo) ListFinanceCollectionVersions(ctx context.Context, orgID, planID int64) ([]autonomy.FinanceCollectionVersion, error) {
	return []autonomy.FinanceCollectionVersion{}, nil
}

func (m *MockAutonomyRepo) UpdateFinanceCollectionPlanStrategy(ctx context.Context, orgID, invoiceID int64, strategyID, draftSubject, draftMsg string, requiresApproval bool, approvalReason *string, priorityLevel string, priorityScore float64) error {
	key := fmt.Sprintf("%d:%d", orgID, invoiceID)
	if p, ok := m.financePlans[key]; ok {
		p.SelectedStrategyID = strategyID
		p.DraftMessage = &draftMsg
		p.RequiresApproval = requiresApproval
		p.ApprovalReason = approvalReason
		p.PriorityLevel = priorityLevel
		p.PriorityScore = priorityScore
	}
	return nil
}

func (m *MockAutonomyRepo) UpdateInvoiceCollectionStatus(ctx context.Context, orgID, invoiceID int64, newStatus string) error {
	return nil
}

// Phase 5 Task 5.7: Contract and Compliance Monitoring Mock Repo
func (m *MockAutonomyRepo) GetContractComplianceContext(ctx context.Context, orgID, contractID int64) (*autonomy.ContractComplianceContextDTO, error) {
	status := "Active"
	effDate := "2026-01-01"
	expDate := "2027-06-30"
	daysUntilExp := 291

	if contractID == 202 {
		expDate = "2026-09-28"
		daysUntilExp = 16
	} else if contractID == 204 {
		status = "Expired"
		expDate = "2025-08-01"
		daysUntilExp = -407
	}

	terms := []map[string]interface{}{
		{"term_type": "SERVICE_LEVEL", "term_description": "Standard Ocean Freight SLA 99.5%"},
		{"term_type": "PRICING_RULE", "term_description": "Base rate locked with BAF indexed monthly"},
	}
	reqs := []map[string]interface{}{
		{"requirement_type": "FMC_FILING", "is_mandatory": true, "compliance_status": "COMPLIANT"},
	}
	docs := []map[string]interface{}{
		{"document_type": "CARRIER_INSURANCE", "verification_status": "VERIFIED"},
	}

	if contractID == 203 {
		reqs = append(reqs, map[string]interface{}{
			"requirement_type": "GDP_PHARMA_CERT", "is_mandatory": true, "compliance_status": "NON_COMPLIANT", "description": "Mandatory Good Distribution Practice Certification for Cold Chain",
		})
	}

	return &autonomy.ContractComplianceContextDTO{
		OrgID:                  orgID,
		ContractID:             contractID,
		ContractReference:      fmt.Sprintf("CTR-2026-DEV-%03d", contractID),
		ContractName:           "Annual Master Service Agreement",
		ContractType:           "CUSTOMER_CARRIER_AGREEMENT",
		PartyName:              "Nordic Freight Dynamics AB",
		Status:                 status,
		Currency:               "USD",
		EffectiveDate:          &effDate,
		ExpiryDate:             &expDate,
		DaysUntilExpiration:    daysUntilExp,
		Terms:                  terms,
		ComplianceRequirements: reqs,
		Documents:              docs,
	}, nil
}

func (m *MockAutonomyRepo) SaveContractComplianceMonitoringPlan(ctx context.Context, plan *autonomy.ContractComplianceMonitoringPlan) error {
	if plan.ID == 0 {
		plan.ID = 7001
	}
	key := fmt.Sprintf("%d:%d", plan.OrgID, plan.ContractID)
	m.compliancePlans[key] = plan
	return nil
}

func (m *MockAutonomyRepo) GetContractComplianceMonitoringPlan(ctx context.Context, orgID, contractID int64) (*autonomy.ContractComplianceMonitoringPlan, error) {
	key := fmt.Sprintf("%d:%d", orgID, contractID)
	if p, ok := m.compliancePlans[key]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *MockAutonomyRepo) GetContractComplianceMonitoringPlanByID(ctx context.Context, orgID, id int64) (*autonomy.ContractComplianceMonitoringPlan, error) {
	for _, p := range m.compliancePlans {
		if p.OrgID == orgID && p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}

func (m *MockAutonomyRepo) SaveContractComplianceMonitoringVersion(ctx context.Context, ver *autonomy.ContractComplianceMonitoringVersion) error {
	return nil
}

func (m *MockAutonomyRepo) ListContractComplianceMonitoringVersions(ctx context.Context, orgID, planID int64) ([]autonomy.ContractComplianceMonitoringVersion, error) {
	return []autonomy.ContractComplianceMonitoringVersion{}, nil
}

func (m *MockAutonomyRepo) UpdateContractCompliancePlanStrategy(ctx context.Context, orgID, contractID int64, strategyID, action, draftSubject, draftMsg string, requiresApproval bool, approvalReason *string) error {
	key := fmt.Sprintf("%d:%d", orgID, contractID)
	if p, ok := m.compliancePlans[key]; ok {
		p.SelectedRemediationStrategyID = strategyID
		p.RequiresApproval = requiresApproval
		p.ApprovalReason = approvalReason
	}
	return nil
}

func (m *MockAutonomyRepo) UpdateContractComplianceStatus(ctx context.Context, orgID, contractID int64, newStatus string) error {
	return nil
}

// Phase 5 Task 5.8: Autonomous Exception Resolution Mock Repo
func (m *MockAutonomyRepo) GetExceptionResolutionContext(ctx context.Context, orgID, exceptionID int64) (*autonomy.ExceptionResolutionContextDTO, error) {
	excType := "CUSTOMS_HOLD"
	sev := "CRITICAL"
	title := "Customs Hold: Discrepancy in HS Code declarations"
	desc := "Customs authority held container due to discrepancy in HS Code declarations."
	if exceptionID == 102 {
		excType = "ETA_DELAY"
		sev = "HIGH"
		title = "Vessel ETA Delayed by 10 Days"
		desc = "Vessel delayed by typhoon."
	}
	return &autonomy.ExceptionResolutionContextDTO{
		OrgID:         orgID,
		ExceptionID:   exceptionID,
		ShipmentID:    103,
		ExceptionType: excType,
		Severity:      sev,
		Title:         title,
		Description:   &desc,
		Status:        "OPEN",
		ShipmentDetails: map[string]interface{}{
			"booking_number":   "BK-2026-DEV-003",
			"carrier_scac":     "CMDU",
			"origin_port":      "INNSA",
			"destination_port": "USNYC",
		},
		CustomerDetails: map[string]interface{}{
			"customer_name": "Apex Global Logistics",
			"account_tier":  "ENTERPRISE",
		},
		ActiveMilestones:  []map[string]interface{}{},
		RelatedExceptions: []map[string]interface{}{},
		Documents:         []map[string]interface{}{},
	}, nil
}

func (m *MockAutonomyRepo) SaveExceptionResolutionPlan(ctx context.Context, plan *autonomy.ExceptionResolutionPlan) error {
	if plan.ID == 0 {
		plan.ID = 8001
	}
	key := fmt.Sprintf("%d:%d", plan.OrgID, plan.ExceptionID)
	m.exceptionPlans[key] = plan
	return nil
}

func (m *MockAutonomyRepo) GetExceptionResolutionPlan(ctx context.Context, orgID, exceptionID int64) (*autonomy.ExceptionResolutionPlan, error) {
	key := fmt.Sprintf("%d:%d", orgID, exceptionID)
	if p, ok := m.exceptionPlans[key]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *MockAutonomyRepo) GetExceptionResolutionPlanByID(ctx context.Context, orgID, id int64) (*autonomy.ExceptionResolutionPlan, error) {
	for _, p := range m.exceptionPlans {
		if p.OrgID == orgID && p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}

func (m *MockAutonomyRepo) SaveExceptionResolutionVersion(ctx context.Context, ver *autonomy.ExceptionResolutionVersion) error {
	return nil
}

func (m *MockAutonomyRepo) ListExceptionResolutionVersions(ctx context.Context, orgID, planID int64) ([]autonomy.ExceptionResolutionVersion, error) {
	return []autonomy.ExceptionResolutionVersion{}, nil
}

func (m *MockAutonomyRepo) UpdateExceptionResolutionPlanStrategy(ctx context.Context, orgID, exceptionID int64, strategyID string, requiresApproval bool, approvalReason *string) error {
	key := fmt.Sprintf("%d:%d", orgID, exceptionID)
	if p, ok := m.exceptionPlans[key]; ok {
		p.SelectedStrategyID = strategyID
		p.RequiresApproval = requiresApproval
		p.ApprovalReason = approvalReason
	}
	return nil
}

func (m *MockAutonomyRepo) UpdateExceptionResolutionStatus(ctx context.Context, orgID, exceptionID int64, lifecycleStatus, waitingState, notes string, resolved bool) error {
	key := fmt.Sprintf("%d:%d", orgID, exceptionID)
	if p, ok := m.exceptionPlans[key]; ok {
		p.LifecycleStatus = lifecycleStatus
		p.WaitingState = &waitingState
		p.ResolutionNotes = &notes
	}
	return nil
}

func (m *MockAutonomyRepo) UpdatePlanCurrentStep(ctx context.Context, orgID int64, planID, stepID string) error {
	key := fmt.Sprintf("%d:%s", orgID, planID)
	if p, ok := m.plans[key]; ok {
		p.CurrentStepID = sql.NullString{String: stepID, Valid: true}
	}
	return nil
}

func (m *MockAutonomyRepo) ApproveStep(ctx context.Context, orgID int64, planID, stepID string) error {
	key := fmt.Sprintf("%d:%s", orgID, planID)
	if steps, ok := m.steps[key]; ok {
		for i := range steps {
			if steps[i].StepID == stepID {
				steps[i].Status = autonomy.StepStatusApproved
				steps[i].RequiresApproval = false
			}
		}
	}
	return nil
}

func (m *MockAutonomyRepo) ResetStepForRetry(ctx context.Context, orgID int64, planID, stepID string, attempt int) error {
	key := fmt.Sprintf("%d:%s", orgID, planID)
	if steps, ok := m.steps[key]; ok {
		for i := range steps {
			if steps[i].StepID == stepID {
				steps[i].Status = autonomy.StepStatusReady
				steps[i].ExecutionAttempt = attempt
			}
		}
	}
	return nil
}

func (m *MockAutonomyRepo) CompensateStep(ctx context.Context, orgID int64, planID, stepID string, compensationDetails string) error {
	key := fmt.Sprintf("%d:%s", orgID, planID)
	if steps, ok := m.steps[key]; ok {
		for i := range steps {
			if steps[i].StepID == stepID {
				steps[i].Status = autonomy.StepStatusCancelled
			}
		}
	}
	return nil
}

func (m *MockAutonomyRepo) GetActivePlansForEntity(ctx context.Context, orgID int64, entityType, entityID string) ([]autonomy.AutonomousPlan, error) {
	var res []autonomy.AutonomousPlan
	for _, p := range m.plans {
		if p.OrgID == orgID && p.RelatedEntityType == entityType && p.RelatedEntityID == entityID {
			res = append(res, *p)
		}
	}
	return res, nil
}

func (m *MockAutonomyRepo) GetPlanningMetrics(ctx context.Context, orgID int64) (*autonomy.PlanningMetricsResponse, error) {
	return &autonomy.PlanningMetricsResponse{
		OrgID:                   orgID,
		TotalPlans:              int64(len(m.plans)),
		CompletedPlans:          1,
		ActivePlans:             1,
		PlanSuccessRate:         1.0,
		TotalSteps:              5,
		CompletedSteps:          5,
		AutonomousExecutionRate: 0.8,
	}, nil
}

func (m *MockAutonomyRepo) IngestMonitoringEvent(ctx context.Context, evt *autonomy.AIMonitoringEvent) error {
	return nil
}

func (m *MockAutonomyRepo) CheckEventDeduplication(ctx context.Context, orgID int64, eventID string) (bool, error) {
	return false, nil
}

func (m *MockAutonomyRepo) UpdatePlanHealth(ctx context.Context, orgID int64, planID string, health autonomy.PlanHealthState, reason string, changedAssumptions []string) error {
	key := fmt.Sprintf("%d:%s", orgID, planID)
	if p, ok := m.plans[key]; ok {
		p.PlanHealth = health
		p.HealthReason = sql.NullString{String: reason, Valid: reason != ""}
	}
	return nil
}

func (m *MockAutonomyRepo) IncrementPlanReplanCount(ctx context.Context, orgID int64, planID string) (int, error) {
	key := fmt.Sprintf("%d:%s", orgID, planID)
	if p, ok := m.plans[key]; ok {
		p.ReplanCount++
		return p.ReplanCount, nil
	}
	return 1, nil
}

func (m *MockAutonomyRepo) ListMonitoringEvents(ctx context.Context, orgID int64, limit int) ([]autonomy.AIMonitoringEvent, error) {
	return nil, nil
}

func (m *MockAutonomyRepo) GetMonitoringMetrics(ctx context.Context, orgID int64) (*autonomy.ContinuousMonitoringMetrics, error) {
	return &autonomy.ContinuousMonitoringMetrics{
		EventsProcessed:    10,
		EventsFiltered:     5,
		AIEvaluatedCount:   5,
		MaterialChangeRate: 0.5,
	}, nil
}

func (m *MockAutonomyRepo) CreateHumanAIDecision(ctx context.Context, dec *autonomy.HumanAIDecision) error {
	key := fmt.Sprintf("%d:%s", dec.OrgID, dec.DecisionID)
	m.decisions[key] = dec
	return nil
}

func (m *MockAutonomyRepo) GetHumanAIDecision(ctx context.Context, orgID int64, decisionID string) (*autonomy.HumanAIDecision, error) {
	key := fmt.Sprintf("%d:%s", orgID, decisionID)
	if d, ok := m.decisions[key]; ok {
		return d, nil
	}
	return &autonomy.HumanAIDecision{
		OrgID:            orgID,
		DecisionID:       decisionID,
		Title:            "Mock Decision",
		AIRecommendation: "Approve priority discharge",
		DecisionStatus:   autonomy.DecisionStatusPending,
		Confidence:       "HIGH",
	}, nil
}

func (m *MockAutonomyRepo) ListHumanAIDecisions(ctx context.Context, orgID int64, status string, module string, limit, offset int) ([]autonomy.HumanAIDecision, int, error) {
	var list []autonomy.HumanAIDecision
	for _, d := range m.decisions {
		if d.OrgID == orgID {
			list = append(list, *d)
		}
	}
	return list, len(list), nil
}

func (m *MockAutonomyRepo) UpdateHumanDecision(ctx context.Context, orgID int64, decisionID string, status autonomy.HumanAIDecisionStatus, decision, reason, humanEditedPayload string, decidedByID int64, decidedByName string, feedbackType string) error {
	key := fmt.Sprintf("%d:%s", orgID, decisionID)
	if d, ok := m.decisions[key]; ok {
		d.DecisionStatus = status
		d.HumanDecision = sql.NullString{String: decision, Valid: true}
		d.DecisionReason = sql.NullString{String: reason, Valid: true}
		if humanEditedPayload != "" {
			d.HumanEditedPayload = json.RawMessage(humanEditedPayload)
		}
		d.DecidedByID = sql.NullInt64{Int64: decidedByID, Valid: true}
		d.DecidedByName = sql.NullString{String: decidedByName, Valid: true}
		d.FeedbackType = sql.NullString{String: feedbackType, Valid: true}
	}
	return nil
}


func (m *MockAutonomyRepo) StopAutonomousPlan(ctx context.Context, orgID int64, planID string, stoppedByUserID int64, reason string) error {
	return nil
}

func (m *MockAutonomyRepo) InvalidatePendingDecisionsForEntity(ctx context.Context, orgID int64, module, entityType, entityID, reason string) (int64, error) {
	return 1, nil
}

func (m *MockAutonomyRepo) UpdateStepHumanContent(ctx context.Context, orgID int64, planID, stepID string, humanContent string) error {
	return nil
}

func (m *MockAutonomyRepo) GetHumanDecisionSummary(ctx context.Context, orgID int64, userID int64) (*autonomy.HumanDecisionCenterSummary, error) {
	return &autonomy.HumanDecisionCenterSummary{}, nil
}

func (m *MockAutonomyRepo) GetCommandCenterOverview(ctx context.Context, orgID int64) (*autonomy.CommandCenterOverviewDTO, error) {
	return &autonomy.CommandCenterOverviewDTO{
		ActiveShipments:       5,
		ShipmentsAtRisk:       2,
		ActiveExceptions:      3,
		CriticalExceptions:    1,
		ActiveWorkflows:       4,
		WorkflowsWaitingHuman: 1,
		PendingApprovals:      2,
		EscalationsCount:      1,
		FailedActionsCount:    0,
		StalledPlansCount:     1,
		ActualSummary:         "5 active shipments, 3 exceptions",
		PredictedSummary:      "2 shipments with predicted delay",
		AIAnalysisSummary:     "4 active workflows running",
		AutonomyDistribution: map[string]int{
			"LEVEL_2_PREPARE":        3,
			"LEVEL_3_POLICY_EXECUTE": 1,
		},
		SystemHealthStatus:    "HEALTHY",
		LastUpdated:           time.Now(),
		IsStale:               false,
	}, nil
}

func (m *MockAutonomyRepo) GetCommandCenterCriticalAttention(ctx context.Context, orgID int64, limit int) ([]autonomy.CriticalAttentionItemDTO, error) {
	return []autonomy.CriticalAttentionItemDTO{
		{
			ID:                "item-crit-1",
			PriorityScore:     95.0,
			PriorityTier:      "CRITICAL_SAFETY_COMPLIANCE",
			Severity:          "CRITICAL",
			EntityType:        "SHIPMENT",
			EntityID:          "101",
			EntityReference:   "SHP-101",
			Title:             "Customs Documentation Hold",
			IssueSummary:      "Missing certificate of origin",
			WhyFlagged:        "Flagged under critical safety compliance",
			ActualFacts:       "Customs hold placed at port",
			PredictedImpact:   "Port detention fee accumulation",
			RecommendedAction: "Review and submit amended documentation",
			Impact:            "Potential port detention",
			RequiredAction:    "Authorize document submission",
			Owner:             "Compliance Officer",
			Urgency:           "IMMEDIATE",
			Source:            "AUTONOMY_ENGINE",
			RequiresHuman:     true,
			CreatedAt:         time.Now(),
		},
	}, nil
}

func (m *MockAutonomyRepo) GetCommandCenterWorkflows(ctx context.Context, orgID int64, module, status, autonomyLevel, search string, limit, offset int) ([]autonomy.AutonomousPlan, int, error) {
	var list []autonomy.AutonomousPlan
	for _, p := range m.plans {
		if p.OrgID == orgID {
			list = append(list, *p)
		}
	}
	return list, len(list), nil
}

func (m *MockAutonomyRepo) GetCommandCenterDecisions(ctx context.Context, orgID int64, limit, offset int) ([]autonomy.HumanAIDecision, int, error) {
	return m.ListHumanAIDecisions(ctx, orgID, "PENDING", "", limit, offset)
}

func (m *MockAutonomyRepo) GetCommandCenterDomainRisks(ctx context.Context, orgID int64) ([]autonomy.DomainRiskSummaryDTO, error) {
	return []autonomy.DomainRiskSummaryDTO{
		{
			Domain:                  "SHIPMENT",
			TotalAtRisk:             1,
			CriticalCount:           0,
			HighCount:               1,
			AuthoritativeState:      "Tracking 1 shipment",
			PredictedRisk:           "1 shipment with delayed ETA",
			ActiveRecoveryWorkflows: 1,
			Items: []autonomy.DomainRiskItemDTO{
				{
					EntityID:          "101",
					EntityReference:   "SHP-101",
					Issue:             "Weather delay on route",
					Severity:          "HIGH",
					ActualFact:        "Shipment in transit",
					PredictedRisk:     "Predicted delay +18h",
					RecommendedAction: "Monitor berth availability",
					Status:            "IN_TRANSIT",
					HasActivePlan:     true,
					LastEvent:         "Milestone recorded",
					UpdatedAt:         time.Now(),
				},
			},
		},
	}, nil
}

func (m *MockAutonomyRepo) GetCommandCenterActivity(ctx context.Context, orgID int64, limit int) (*autonomy.CommandCenterActivityDTO, error) {
	return &autonomy.CommandCenterActivityDTO{
		RecentActions:    []autonomy.CommandCenterActionDTO{},
		ReplanningEvents: []autonomy.CommandCenterReplanDTO{},
		Escalations:      []autonomy.CommandCenterEscalationDTO{},
	}, nil
}

func (m *MockAutonomyRepo) GetSystemHealth(ctx context.Context) (*autonomy.SystemHealthStatusDTO, error) {
	now := time.Now()
	return &autonomy.SystemHealthStatusDTO{
		OverallStatus: "HEALTHY",
		Subsystems: map[string]autonomy.SubsystemHealthDTO{
			"go_backend": {Name: "Go Backend", Status: "HEALTHY", LatencyMs: 1, Message: "Online", LastChecked: now},
			"database":   {Name: "Database", Status: "HEALTHY", LatencyMs: 1, Message: "Connected", LastChecked: now},
		},
		LastCheckedAt: now,
	}, nil
}

// Phase 5 Task 5.13 Mock Methods
func (m *MockAutonomyRepo) RecordOutcome(ctx context.Context, outcome *autonomy.AgentOutcome) error {
	if outcome.OutcomeID == "" {
		outcome.OutcomeID = fmt.Sprintf("out_%d_%d", outcome.OrgID, time.Now().UnixNano())
	}
	return nil
}

func (m *MockAutonomyRepo) GetOutcome(ctx context.Context, orgID int64, outcomeID string) (*autonomy.AgentOutcome, error) {
	return &autonomy.AgentOutcome{
		OrgID:            orgID,
		OutcomeID:        outcomeID,
		SourceEntityType: "shipment",
		SourceEntityID:   "101",
		OutcomeType:      "ACTION_RESULT",
		Status:           "SUCCESS",
		IsVerified:       true,
		ConfidenceScore:  0.95,
	}, nil
}

func (m *MockAutonomyRepo) ListOutcomes(ctx context.Context, orgID int64, entityType, entityID, outcomeType, status string, limit, offset int) ([]autonomy.AgentOutcome, int, error) {
	return []autonomy.AgentOutcome{
		{
			OrgID:            orgID,
			OutcomeID:        "out-mock-1",
			SourceEntityType: "shipment",
			SourceEntityID:   "101",
			OutcomeType:      "ACTION_RESULT",
			Status:           "SUCCESS",
			IsVerified:       true,
		},
	}, 1, nil
}

func (m *MockAutonomyRepo) VerifyOutcome(ctx context.Context, orgID int64, outcomeID string, status string, actualResult, verificationMethod string, verifiedByID *int64, failureCategory *string) error {
	return nil
}

func (m *MockAutonomyRepo) SaveLearnedMemory(ctx context.Context, mem *autonomy.ExtendedMemoryItem) error {
	if mem.ID == 0 {
		mem.ID = time.Now().UnixNano()
	}
	return nil
}

func (m *MockAutonomyRepo) GetLearnedMemoryByID(ctx context.Context, orgID int64, memoryID int64) (*autonomy.ExtendedMemoryItem, error) {
	return &autonomy.ExtendedMemoryItem{
		ID:             memoryID,
		OrgID:          orgID,
		Category:       "OPERATIONAL",
		Title:          "Mock Learned Memory",
		Content:        "Carrier responds rapidly to priority escalation on Port of Oakland lanes.",
		Confidence:     0.90,
		RecencyWeight:  1.0,
		Status:         "ACTIVE",
		ProvenanceType: "SYSTEM_DERIVED",
	}, nil
}

func (m *MockAutonomyRepo) ListLearnedMemories(ctx context.Context, orgID int64, category, entityType, entityID, scope string, includeStale bool, limit, offset int) ([]autonomy.ExtendedMemoryItem, int, error) {
	return []autonomy.ExtendedMemoryItem{
		{
			ID:             9001,
			OrgID:          orgID,
			Category:       "OPERATIONAL",
			Title:          "Carrier Escalation Pattern",
			Content:        "Carrier responds rapidly to priority escalation on Port of Oakland lanes.",
			Confidence:     0.90,
			RecencyWeight:  1.0,
			Status:         "ACTIVE",
			ProvenanceType: "AI_DERIVED",
		},
	}, 1, nil
}

func (m *MockAutonomyRepo) UpdateMemoryCorrection(ctx context.Context, orgID int64, memoryID int64, newContent, reason string, correctedByID int64) error {
	return nil
}

func (m *MockAutonomyRepo) InvalidateMemory(ctx context.Context, orgID int64, memoryID int64, reason string, invalidatedByID int64) error {
	return nil
}

func (m *MockAutonomyRepo) FlagMemoryUnreliable(ctx context.Context, orgID int64, memoryID int64, reason string, flaggedByID int64) error {
	return nil
}

func (m *MockAutonomyRepo) SaveLearnedPattern(ctx context.Context, pat *autonomy.LearnedPattern) error {
	if pat.ID == 0 {
		pat.ID = time.Now().UnixNano()
	}
	return nil
}

func (m *MockAutonomyRepo) ListLearnedPatterns(ctx context.Context, orgID int64, patternType, entityType string, limit, offset int) ([]autonomy.LearnedPattern, int, error) {
	strat := "Escalate carrier operations lead after 4h delay"
	return []autonomy.LearnedPattern{
		{
			ID:                     8001,
			OrgID:                  orgID,
			PatternID:              "pat-mock-1",
			PatternType:            "CARRIER_BEHAVIOR",
			EntityType:             "carrier",
			EntityIdentifier:       "MAEU",
			Title:                  "Maersk Pacific Route Escalation",
			Description:            "Recurring recovery success when escalating early.",
			RecommendedStrategy:    &strat,
			SupportingObservations: 4,
			SuccessRate:            0.80,
			Confidence:             "HIGH",
			IsActive:               true,
		},
	}, 1, nil
}

func (m *MockAutonomyRepo) GetMemoryLearningSummary(ctx context.Context, orgID int64) (*autonomy.MemoryLearningSummaryDTO, error) {
	return &autonomy.MemoryLearningSummaryDTO{
		TotalMemories:           15,
		ActiveMemories:          12,
		StaleMemories:           2,
		InvalidatedMemories:     1,
		TotalOutcomes:           20,
		VerifiedOutcomes:        18,
		SuccessfulOutcomes:      16,
		FailedOutcomes:          2,
		DetectedPatterns:        4,
		OverallSuccessRate:      0.88,
		RecommendationAcceptPct: 0.92,
		MemoryCategoryCounts:    map[string]int{"OPERATIONAL": 8, "CUSTOMER": 4},
		TopPatterns:             []autonomy.LearnedPattern{},
		RecentOutcomes:          []autonomy.AgentOutcome{},
	}, nil
}

func (m *MockAutonomyRepo) GetTenantLimits(ctx context.Context, orgID int64) (*autonomy.TenantGovernanceLimits, error) {
	if l, ok := m.govLimits[orgID]; ok {
		return l, nil
	}
	return &autonomy.TenantGovernanceLimits{
		OrgID:                           orgID,
		MaxTenantAutonomy:               4,
		KillSwitchActive:                false,
		MaxActionsPerHour:               100,
		MaxFinancialExposurePerWorkflow: 5000.0,
		MaxRetriesPerStep:               3,
		MaxReplansPerPlan:               5,
		EnforceFourEyes:                 true,
	}, nil
}

func (m *MockAutonomyRepo) UpdateTenantLimits(ctx context.Context, limits *autonomy.TenantGovernanceLimits) error {
	m.govLimits[limits.OrgID] = limits
	return nil
}

func (m *MockAutonomyRepo) ToggleKillSwitch(ctx context.Context, orgID int64, active bool, reason string, userID int64) error {
	l, err := m.GetTenantLimits(ctx, orgID)
	if err != nil {
		return err
	}
	l.KillSwitchActive = active
	l.KillSwitchReason = &reason
	m.govLimits[orgID] = l
	return nil
}

func (m *MockAutonomyRepo) GetActionAllowlist(ctx context.Context, orgID int64, module string) ([]autonomy.ActionAllowlistItem, error) {
	var list []autonomy.ActionAllowlistItem
	for _, item := range m.allowlist {
		if item.OrgID == orgID {
			if module == "" || item.Module == module {
				list = append(list, *item)
			}
		}
	}
	return list, nil
}

func (m *MockAutonomyRepo) GetActionAllowlistItem(ctx context.Context, orgID int64, actionType string) (*autonomy.ActionAllowlistItem, error) {
	key := fmt.Sprintf("%d:%s", orgID, actionType)
	if it, ok := m.allowlist[key]; ok {
		return it, nil
	}
	return nil, nil
}

func (m *MockAutonomyRepo) SaveActionAllowlistItem(ctx context.Context, item *autonomy.ActionAllowlistItem) error {
	key := fmt.Sprintf("%d:%s", item.OrgID, item.ActionType)
	m.allowlist[key] = item
	return nil
}

func (m *MockAutonomyRepo) GetFeatureFlags(ctx context.Context, orgID int64) ([]autonomy.GovernanceFeatureFlag, error) {
	var flags []autonomy.GovernanceFeatureFlag
	for _, f := range m.featureFlags {
		if f.OrgID == orgID {
			flags = append(flags, *f)
		}
	}
	return flags, nil
}

func (m *MockAutonomyRepo) UpdateFeatureFlag(ctx context.Context, orgID int64, flagKey string, enabled bool, maxAutonomy int, reqApproval bool, userID int64) error {
	key := fmt.Sprintf("%d:%s", orgID, flagKey)
	if f, ok := m.featureFlags[key]; ok {
		f.IsEnabled = enabled
		if maxAutonomy > 0 {
			f.MaxAutonomyLevel = maxAutonomy
		}
		f.RequiresApproval = reqApproval
		return nil
	}
	m.featureFlags[key] = &autonomy.GovernanceFeatureFlag{
		OrgID:            orgID,
		FlagKey:          flagKey,
		FlagName:         flagKey,
		IsEnabled:        enabled,
		MaxAutonomyLevel: maxAutonomy,
		RequiresApproval: reqApproval,
	}
	return nil
}

func (m *MockAutonomyRepo) RecordPolicyEvaluation(ctx context.Context, record *autonomy.PolicyEvaluationRecord) error {
	return nil
}

func (m *MockAutonomyRepo) ListPolicyEvaluations(ctx context.Context, orgID int64, limit, offset int) ([]autonomy.PolicyEvaluationRecord, int, error) {
	return []autonomy.PolicyEvaluationRecord{}, 0, nil
}

func (m *MockAutonomyRepo) RecordPolicyAuditLog(ctx context.Context, log *autonomy.PolicyAuditLog) error {
	return nil
}

func (m *MockAutonomyRepo) ListPolicyAuditLogs(ctx context.Context, orgID int64, limit, offset int) ([]autonomy.PolicyAuditLog, int, error) {
	return []autonomy.PolicyAuditLog{}, 0, nil
}

func (m *MockAutonomyRepo) GetGovernanceTelemetry(ctx context.Context, orgID int64) (*autonomy.GovernanceTelemetrySummary, error) {
	return &autonomy.GovernanceTelemetrySummary{
		TotalEvaluations:    10,
		AllowedCount:        8,
		BlockedCount:        1,
		ReviewRequiredCount: 1,
		KillSwitchActive:    false,
		MaxTenantAutonomy:   4,
		ActiveFlagsCount:    3,
		AllowlistCount:      3,
		RecentEvaluations:   []autonomy.PolicyEvaluationRecord{},
	}, nil
}

// MockSidecarClient satisfies autonomy.SidecarClient



type MockSidecarClient struct {
	generateResp      *autonomy.SidecarPlanGenResponse
	evalResp          *autonomy.SidecarPlanEvalResponse
	replanResp        *autonomy.SidecarReplanResponse
	shipmentEvalResp  *autonomy.SidecarShipmentEventEvalResponse
	custEvalResp      *autonomy.SidecarCustomerFollowupEvalResponse
	custClassResp     *autonomy.SidecarClassifyCustomerResponseResponse
	pricingEvalResp   *autonomy.SidecarEvaluateRfqPricingResponse
	pricingReplanResp *autonomy.SidecarEvaluateRfqPricingResponse
	err               error
}

func (m *MockSidecarClient) GeneratePlan(ctx context.Context, req *autonomy.SidecarPlanGenRequest) (*autonomy.SidecarPlanGenResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.generateResp, nil
}

func (m *MockSidecarClient) EvaluatePlan(ctx context.Context, req *autonomy.SidecarPlanEvalRequest) (*autonomy.SidecarPlanEvalResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.evalResp, nil
}

func (m *MockSidecarClient) Replan(ctx context.Context, req *autonomy.SidecarReplanRequest) (*autonomy.SidecarReplanResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.replanResp, nil
}

func (m *MockSidecarClient) EvaluateShipmentEvent(ctx context.Context, req *autonomy.SidecarShipmentEventEvalRequest) (*autonomy.SidecarShipmentEventEvalResponse, error) {
	if m.shipmentEvalResp != nil {
		return m.shipmentEvalResp, nil
	}
	return &autonomy.SidecarShipmentEventEvalResponse{
		ShipmentID:             req.ShipmentID,
		EventType:              req.EventType,
		Decision:               "CONTINUE_MONITORING",
		DecisionReason:         "Mock evaluation",
		IsMeaningfulChange:     false,
		CommitmentRiskSeverity: "NONE",
	}, nil
}

func (m *MockSidecarClient) GenerateAdaptiveShipmentPlan(ctx context.Context, req *autonomy.SidecarPlanGenRequest) (*autonomy.SidecarPlanGenResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.generateResp, nil
}

func (m *MockSidecarClient) EvaluateCustomerFollowup(ctx context.Context, req *autonomy.SidecarCustomerFollowupEvalRequest) (*autonomy.SidecarCustomerFollowupEvalResponse, error) {
	if m.custEvalResp != nil {
		return m.custEvalResp, nil
	}
	return &autonomy.SidecarCustomerFollowupEvalResponse{
		CustomerID:       req.CustomerID,
		EventType:        req.EventType,
		Decision:         "PREPARE_FOLLOW_UP",
		DecisionReason:   "Mock customer followup evaluation",
		Urgency:          "MEDIUM",
		Channel:          "EMAIL",
		RequiresApproval: false,
		Confidence:       0.90,
	}, nil
}

func (m *MockSidecarClient) ClassifyCustomerResponse(ctx context.Context, req *autonomy.SidecarClassifyCustomerResponseRequest) (*autonomy.SidecarClassifyCustomerResponseResponse, error) {
	if m.custClassResp != nil {
		return m.custClassResp, nil
	}
	return &autonomy.SidecarClassifyCustomerResponseResponse{
		CustomerID:          req.CustomerID,
		Classification:      "CONFIRMATION_APPROVAL",
		Sentiment:           "POSITIVE",
		RecommendedNextStep: "RESOLVE_AND_STOP",
		Confidence:          0.95,
	}, nil
}

func (m *MockSidecarClient) EvaluateRfqPricing(ctx context.Context, req *autonomy.SidecarEvaluateRfqPricingRequest) (*autonomy.SidecarEvaluateRfqPricingResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.pricingEvalResp != nil {
		return m.pricingEvalResp, nil
	}
	return &autonomy.SidecarEvaluateRfqPricingResponse{
		RfqID:                 req.Context.RfqID,
		Currency:              "USD",
		BaseCost:              2700.0,
		PredictedCost:         2781.0,
		ActualFacts:           []string{"Verified Route: INNSA -> NLRTM", "Carrier Base Cost: $2,700.00 USD"},
		Predictions:           []string{"Predicted Cost: $2,781.00 USD"},
		Assumptions:           []string{"14 free days demurrage"},
		RecommendedStrategyID: "strat-competitive-std",
		RecommendedPrice:      3214.29,
		RecommendedMarginPct:  16.0,
		MarginRiskLevel:       "LOW",
		OperationalRiskLevel:  "LOW",
		ConfidenceScore:       0.90,
		DataSufficiency:       "COMPLETE",
		RateFreshnessStatus:   "FRESH",
		RequiresApproval:      false,
		ReasoningSummary:      "Mock pricing evaluation summary",
		CandidateStrategies: []autonomy.PricingStrategyCandidateDTO{
			{
				StrategyID:            "strat-competitive-std",
				StrategyType:          "COMPETITIVE_STANDARD",
				Title:                 "Competitive Market Baseline",
				Price:                 3214.29,
				BaseCost:              2700.0,
				PredictedCost:         2781.0,
				MarginPct:             16.0,
				MarginAmount:          514.29,
				OperationalRiskLevel:  "LOW",
				MarginRiskLevel:       "LOW",
				AcceptanceProbability: 0.82,
				ConfidenceScore:       0.90,
				IsFeasible:            true,
			},
			{
				StrategyID:            "strat-below-floor-infeasible",
				StrategyType:          "CONSERVATIVE_HOLD",
				Title:                 "Aggressive Sub-Floor Bid (Infeasible)",
				Price:                 2800.0,
				BaseCost:              2700.0,
				PredictedCost:         2781.0,
				MarginPct:             3.57,
				MarginAmount:          100.0,
				OperationalRiskLevel:  "HIGH",
				MarginRiskLevel:       "CRITICAL",
				AcceptanceProbability: 0.99,
				ConfidenceScore:       0.50,
				IsFeasible:            false,
				InfeasibilityReason:   stringPtr("Hard Constraint Violation: Margin below minimum floor"),
			},
		},
	}, nil
}

func (m *MockSidecarClient) ReplanRfqPricing(ctx context.Context, req *autonomy.SidecarReplanPricingRequest) (*autonomy.SidecarEvaluateRfqPricingResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.pricingReplanResp != nil {
		return m.pricingReplanResp, nil
	}
	return m.EvaluateRfqPricing(ctx, &autonomy.SidecarEvaluateRfqPricingRequest{
		Context:       req.Context,
		CorrelationID: req.CorrelationID,
	})
}

func (m *MockSidecarClient) EvaluateFinanceCollection(ctx context.Context, req *autonomy.FinanceCollectionEvaluationRequestDTO) (*autonomy.FinanceCollectionEvaluationResponseDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	draftMsg := "Friendly reminder regarding outstanding invoice."
	draftSubj := "Payment Reminder"
	var stopReason *string
	if req.Context.BalanceDue <= 0.0 {
		s := "Invoice has been fully settled. No collection action required."
		stopReason = &s
	}
	return &autonomy.FinanceCollectionEvaluationResponseDTO{
		InvoiceID:             req.Context.InvoiceID,
		InvoiceNumber:         req.Context.InvoiceNumber,
		CustomerName:          req.Context.CustomerName,
		Currency:              "USD",
		TotalAmount:           req.Context.TotalAmount,
		BalanceDue:            req.Context.BalanceDue,
		DaysOverdue:           req.Context.DaysOverdue,
		AgingBucket:           req.Context.AgingBucket,
		PriorityLevel:         "MEDIUM",
		PriorityScore:         55.0,
		RiskLevel:             "LOW",
		RiskScore:             0.25,
		ActualFacts:           []string{"Authoritative balance: $4500.00 USD", "Days overdue: 27 days"},
		Predictions:           []string{"Collection probability: 85%", "Cash-flow risk: Low-Medium"},
		Assumptions:           []string{"Standard bank clearing period", "Primary contact valid"},
		RecommendedStrategyID: "strat-friendly-reminder",
		RecommendedAction:     "SEND_FRIENDLY_REMINDER",
		DraftSubject:          draftSubj,
		DraftMessage:          draftMsg,
		RequiresApproval:      false,
		StopReason:            stopReason,
		ConfidenceScore:       0.90,
		DataSufficiency:       "COMPLETE",
		PlanSteps: []map[string]interface{}{
			{"step_id": "step-fin-1", "action_type": "VERIFY_INVOICE_BALANCE"},
			{"step_id": "step-fin-2", "action_type": "CHECK_PAYMENT_STATUS"},
			{"step_id": "step-fin-3", "action_type": "EVALUATE_AGING"},
			{"step_id": "step-fin-4", "action_type": "SELECT_STRATEGY"},
			{"step_id": "step-fin-5", "action_type": "DRAFT_COMMUNICATION"},
			{"step_id": "step-fin-6", "action_type": "OBTAIN_APPROVAL"},
			{"step_id": "step-fin-7", "action_type": "DISPATCH_ACTION"},
		},
		CandidateStrategies: []autonomy.CollectionStrategyCandidateDTO{
			{
				StrategyID:        "strat-friendly-reminder",
				StrategyName:      "Friendly Payment Reminder",
				Description:       "Standard reminder",
				RecommendedAction: "SEND_FRIENDLY_REMINDER",
				PriorityLevel:     "MEDIUM",
				PriorityScore:     55.0,
				IsFeasible:        true,
			},
			{
				StrategyID:        "strat-finance-escalation",
				StrategyName:      "Finance Escalation",
				Description:       "Senior credit review",
				RecommendedAction: "ESCALATE_TO_FINANCE",
				PriorityLevel:     "HIGH",
				PriorityScore:     85.0,
				IsFeasible:        true,
				RequiresApproval:  true,
			},
			{
				StrategyID:        "strat-status-confirmation",
				StrategyName:      "Status Confirmation",
				Description:       "Request wire or processing schedule",
				RecommendedAction: "REQUEST_PAYMENT_STATUS",
				PriorityLevel:     "HIGH",
				PriorityScore:     70.0,
				IsFeasible:        true,
			},
		},
		CorrelationID: req.CorrelationID,
	}, nil
}

func (m *MockSidecarClient) ReplanFinanceCollection(ctx context.Context, req *autonomy.FinanceCollectionReplanningRequestDTO) (*autonomy.FinanceCollectionEvaluationResponseDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	ctxCopy := req.Context
	if req.TriggerEvent == "PARTIAL_PAYMENT" {
		if paid, ok := req.EventPayload["amount_paid"].(float64); ok {
			ctxCopy.BalanceDue = req.Context.BalanceDue - paid
		}
	}
	return m.EvaluateFinanceCollection(ctx, &autonomy.FinanceCollectionEvaluationRequestDTO{
		Context:       ctxCopy,
		CorrelationID: req.CorrelationID,
	})
}

func (m *MockSidecarClient) EvaluateContractCompliance(ctx context.Context, req *autonomy.ContractComplianceEvaluationRequestDTO) (*autonomy.ContractComplianceEvaluationResponseDTO, error) {
	if m.err != nil {
		return nil, m.err
	}

	compStatus := "COMPLIANT"
	expStatus := "CURRENT"
	if req.Context.DaysUntilExpiration < 0 {
		expStatus = "EXPIRED"
	} else if req.Context.DaysUntilExpiration <= 30 {
		expStatus = "EXPIRING_SOON"
	}

	for _, reqItem := range req.Context.ComplianceRequirements {
		if s, ok := reqItem["compliance_status"].(string); ok && s == "NON_COMPLIANT" {
			compStatus = "NON_COMPLIANT"
		}
	}

	stratID := "strat-compliance-ok-monitor"
	recAction := "CONTINUE_MONITORING"
	requiresApproval := false
	var approvalReason *string
	var stopReason *string
	confScore := 0.95

	draftSubj := fmt.Sprintf("LogisticsHQ Compliance Status - %s", req.Context.ContractReference)
	draftMsg := fmt.Sprintf("Contract %s compliance evaluation completed successfully.", req.Context.ContractReference)

	candidates := []autonomy.ComplianceRemediationStrategyCandidateDTO{
		{
			StrategyID:        "strat-compliance-ok-monitor",
			StrategyName:      "Routine Compliance Monitoring",
			StrategyType:      "MONITOR_ONLY",
			Description:       "Continuous telemetry monitoring with periodic expiration alerts",
			RecommendedAction: "CONTINUE_MONITORING",
			SeverityLevel:     "LOW",
			IsFeasible:        true,
			RequiresApproval:  false,
		},
	}

	if expStatus == "EXPIRED" {
		stratID = "strat-hard-compliance-escalation"
		recAction = "ESCALATE_TO_LEGAL"
		requiresApproval = true
		reason := "Contract has expired. All commercial actions blocked until formal amendment or renewal."
		approvalReason = &reason
		stopReason = &reason
		candidates = append(candidates, autonomy.ComplianceRemediationStrategyCandidateDTO{
			StrategyID:        "strat-hard-compliance-escalation",
			StrategyName:      "Statutory Legal & Compliance Escalation",
			StrategyType:      "HARD_COMPLIANCE_ESCALATION",
			Description:       "Cease autonomous dispatch. Immediate human review required.",
			RecommendedAction: "ESCALATE_TO_LEGAL",
			SeverityLevel:     "CRITICAL",
			IsFeasible:        true,
			RequiresApproval:  true,
		})
	} else if expStatus == "EXPIRING_SOON" {
		stratID = "strat-expiring-renewal-prep"
		recAction = "PREPARE_RENEWAL_NOTICE"
		requiresApproval = true
		reason := "Contract approaches expiration threshold (< 30 days remaining)."
		approvalReason = &reason
		candidates = append(candidates, autonomy.ComplianceRemediationStrategyCandidateDTO{
			StrategyID:        "strat-expiring-renewal-prep",
			StrategyName:      "Proactive Renewal Notice & Commercial Extension",
			StrategyType:      "RENEWAL_PREP",
			Description:       "Draft commercial renewal notice with existing rate schedule baseline",
			RecommendedAction: "PREPARE_RENEWAL_NOTICE",
			SeverityLevel:     "MEDIUM",
			IsFeasible:        true,
			RequiresApproval:  true,
		})
	} else if compStatus == "NON_COMPLIANT" {
		stratID = "strat-missing-document-remediation"
		recAction = "REQUEST_STATUTORY_DOCUMENT"
		requiresApproval = false
		candidates = append(candidates, autonomy.ComplianceRemediationStrategyCandidateDTO{
			StrategyID:        "strat-missing-document-remediation",
			StrategyName:      "Mandatory Document Expedited Request",
			StrategyType:      "DOCUMENT_REMEDIATION",
			Description:       "Automated request to counterparty for missing statutory compliance documentation",
			RecommendedAction: "REQUEST_STATUTORY_DOCUMENT",
			SeverityLevel:     "HIGH",
			IsFeasible:        true,
			RequiresApproval:  false,
		})
	}

	return &autonomy.ContractComplianceEvaluationResponseDTO{
		ContractID:            req.Context.ContractID,
		ContractReference:     req.Context.ContractReference,
		ContractName:          req.Context.ContractName,
		PartyName:             req.Context.PartyName,
		ContractType:          req.Context.ContractType,
		Status:                req.Context.Status,
		EffectiveDate:         req.Context.EffectiveDate,
		ExpiryDate:            req.Context.ExpiryDate,
		DaysUntilExpiration:   req.Context.DaysUntilExpiration,
		ExpirationStatus:      expStatus,
		ComplianceStatus:      compStatus,
		HardViolationsCount:   0,
		SoftDeviationsCount:   0,
		RiskLevel:             "LOW",
		RiskScore:             0.1,
		ConfidenceScore:       confScore,
		DataSufficiency:       "COMPLETE",
		AuthoritativeFacts:    []string{"Contract active in registry", "Effective date: 2026-01-01"},
		ExtractedTerms:        []string{"FMC Filing required", "SLA 99.5%"},
		Predictions:           []string{"Estimated renewal friction: LOW"},
		Assumptions:           []string{"Counterparty operational contact is current"},
		Deviations:            []map[string]interface{}{},
		CandidateStrategies:   candidates,
		RecommendedStrategyID: stratID,
		RecommendedAction:     recAction,
		DraftSubject:          draftSubj,
		DraftMessage:          draftMsg,
		RequiresApproval:      requiresApproval,
		ApprovalReason:        approvalReason,
		StopReason:            stopReason,
		RemediationPlanSteps: []map[string]interface{}{
			{"step_id": "step-ccm-1", "action_type": "AUDIT_REGULATORY_REQUIREMENTS"},
			{"step_id": "step-ccm-2", "action_type": "INSPECT_DOCUMENT_EXPIRATIONS"},
			{"step_id": "step-ccm-3", "action_type": "VERIFY_COMMERCIAL_RATES"},
			{"step_id": "step-ccm-4", "action_type": "ASSESS_SEVERITY_AND_IMPACT"},
			{"step_id": "step-ccm-5", "action_type": "FORMULATE_REMEDIATION_STRATEGY"},
			{"step_id": "step-ccm-6", "action_type": "GOVERN_HUMAN_APPROVAL"},
			{"step_id": "step-ccm-7", "action_type": "DISPATCH_AUTHORIZED_ACTION"},
		},
		CorrelationID: req.CorrelationID,
	}, nil
}

func (m *MockSidecarClient) ReplanContractCompliance(ctx context.Context, req *autonomy.ContractComplianceReplanningRequestDTO) (*autonomy.ContractComplianceEvaluationResponseDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	ctxCopy := req.Context
	if req.TriggerEvent == "DOCUMENT_VERIFIED" {
		// clear non-compliant requirement
		ctxCopy.ComplianceRequirements = []map[string]interface{}{
			{"requirement_type": "FMC_FILING", "is_mandatory": true, "compliance_status": "COMPLIANT"},
		}
	}
	return m.EvaluateContractCompliance(ctx, &autonomy.ContractComplianceEvaluationRequestDTO{
		Context:       ctxCopy,
		CorrelationID: req.CorrelationID,
	})
}

func (m *MockSidecarClient) EvaluateExceptionResolution(ctx context.Context, req *autonomy.ExceptionEvaluationRequestDTO) (*autonomy.ExceptionEvaluationResponseDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	excType := req.Context.ExceptionType
	sev := req.Context.Severity
	stratID := "strat-carrier-escalation"
	stratName := "Carrier Priority Escalation & Reroute Request"
	recAction := "ESCALATE_CARRIER"
	_ = recAction
	requiresApproval := false
	var approvalReason *string
	waitingState := "WAITING_FOR_CARRIER"

	if excType == "CUSTOMS_HOLD" {
		stratID = "strat-customs-document-remedy"
		stratName = "Broker Expedited Customs Document Remediation"
		recAction = "SUBMIT_CUSTOMS_CORRECTION"
		_ = recAction
		waitingState = "WAITING_FOR_APPROVAL"
		requiresApproval = true
		r := "Critical statutory document remediation requires compliance supervisor validation."
		approvalReason = &r
	}

	candidates := []autonomy.ExceptionCandidateRecoveryStrategyDTO{
		{
			StrategyID:             "strat-carrier-escalation",
			StrategyName:           "Carrier Priority Escalation & Reroute Request",
			StrategyType:           "CARRIER_ESCALATION",
			Description:            "Issue carrier priority escalation",
			RecommendedAction:      "ESCALATE_CARRIER",
			ExpectedResolutionProb: 0.88,
			TimeToResolution:       "6-12 hours",
			Score:                  0.90,
			IsFeasible:             true,
		},
		{
			StrategyID:             "strat-customs-document-remedy",
			StrategyName:           "Broker Expedited Customs Document Remediation",
			StrategyType:           "DOCUMENT_REMEDY",
			Description:            "Dispatch automated amendment packet with certified HS code",
			RecommendedAction:      "SUBMIT_CUSTOMS_CORRECTION",
			ExpectedResolutionProb: 0.94,
			TimeToResolution:       "4-8 hours",
			RequiresApproval:       sev == "CRITICAL",
			Score:                  0.95,
			IsFeasible:             true,
		},
		{
			StrategyID:             "strat-customer-proactive-notice",
			StrategyName:           "Proactive Shipper Delivery Schedule Adjustment Advisory",
			StrategyType:           "CUSTOMER_ADVISORY",
			Description:            "Send transparent proactive delivery advisory",
			RecommendedAction:      "ISSUE_CUSTOMER_ADVISORY",
			ExpectedResolutionProb: 0.85,
			TimeToResolution:       "Immediate (< 1 hour)",
			Score:                  0.82,
			IsFeasible:             true,
		},
		{
			StrategyID:             "strat-re-route-alternate-corridor",
			StrategyName:           "Intermodal Feeder / Alternate Inland Corridor Diversion",
			StrategyType:           "ALTERNATE_CORRIDOR",
			Description:            "Divert cargo to adjacent feeder terminal",
			RecommendedAction:      "DISPATCH_FEEDER_REROUTE",
			ExpectedResolutionProb: 0.70,
			TimeToResolution:       "12-24 hours",
			RequiresApproval:       true,
			Score:                  0.68,
			IsFeasible:             excType != "CUSTOMS_HOLD",
		},
		{
			StrategyID:             "strat-executive-ops-escalation",
			StrategyName:           "Operations Director Manual Incident Escalation",
			StrategyType:           "OPERATIONS_ESCALATION",
			Description:            "Escalate exception directly to Senior Operations Leadership",
			RecommendedAction:      "ESCALATE_OPS_DIRECTOR",
			ExpectedResolutionProb: 0.70,
			TimeToResolution:       "2-4 hours",
			RequiresApproval:       true,
			Score:                  0.72,
			IsFeasible:             true,
		},
	}

	return &autonomy.ExceptionEvaluationResponseDTO{
		ExceptionID:          req.Context.ExceptionID,
		ShipmentID:           req.Context.ShipmentID,
		ExceptionType:        excType,
		Severity:             sev,
		LifecycleStatus:      "PLAN_READY",
		WaitingState:         &waitingState,
		Symptom:              fmt.Sprintf("Exception detected on shipment %d: %s", req.Context.ShipmentID, req.Context.Title),
		LikelyRootCause:      "Discrepancy in HS Code declarations with customs authority.",
		ContributingFactors:  []string{"Declared HS Code does not match bill of lading description."},
		Evidence:             []string{"Authoritative exception record flagged: " + req.Context.Title},
		ConfidenceScore:      0.92,
		UnknownFactors:       []string{"Customs broker inspection queue duration."},
		ImpactAssessment:     map[string]interface{}{"operational_impact": "HIGH", "financial_impact": 450.0},
		HardConstraints:      []string{"Compliance prohibition: Autonomous cargo release without verified clearance is prohibited."},
		CandidateStrategies:  candidates,
		SelectedStrategyID:   stratID,
		SelectedStrategyName: stratName,
		RecoveryPlanSteps: []map[string]interface{}{
			{"step_id": "step-1-triage-lock", "name": "Triage & Verification Lock", "status": "COMPLETED"},
			{"step_id": "step-2-evidence-retrieval", "name": "Root-Cause Evidence Retrieval", "status": "COMPLETED"},
			{"step_id": "step-3-impact-assessment", "name": "Operational Impact & Constraint Assessment", "status": "COMPLETED"},
			{"step_id": "step-4-remediation-dispatch", "name": "Remediation Dispatch", "status": "PENDING"},
			{"step_id": "step-5-asynchronous-wait", "name": "Asynchronous Wait", "status": "PENDING"},
			{"step_id": "step-6-outcome-verification", "name": "Authoritative Outcome Verification", "status": "PENDING"},
			{"step_id": "step-7-closure-finalization", "name": "Closure & Audit Finalization", "status": "PENDING"},
		},
		VerificationCriteria: map[string]interface{}{"required_business_state": "RESOLVED"},
		RequiresApproval:     requiresApproval,
		ApprovalReason:       approvalReason,
		DataSufficiency:      "COMPLETE",
		CorrelationID:        req.CorrelationID,
	}, nil
}

func (m *MockSidecarClient) ReplanExceptionResolution(ctx context.Context, req *autonomy.ExceptionReplanningRequestDTO) (*autonomy.ExceptionEvaluationResponseDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	resp, err := m.EvaluateExceptionResolution(ctx, &autonomy.ExceptionEvaluationRequestDTO{
		Context:       req.Context,
		CorrelationID: req.CorrelationID,
	})
	if err != nil {
		return nil, err
	}
	resp.LifecycleStatus = "RESOLVING"
	w := "WAITING_FOR_VERIFICATION"
	resp.WaitingState = &w
	return resp, nil
}

func (m *MockSidecarClient) ValidatePlanGraph(ctx context.Context, req *autonomy.SidecarPlanValidationRequest) (*autonomy.SidecarPlanValidationResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	var order []string
	for _, s := range req.Steps {
		order = append(order, s.StepID)
	}
	return &autonomy.SidecarPlanValidationResponse{
		IsValid:        true,
		ExecutionOrder: order,
		ParallelGroups: [][]string{order},
	}, nil
}

func (m *MockSidecarClient) GenerateCrossModulePlan(ctx context.Context, req *autonomy.SidecarCrossModulePlanRequest) (*autonomy.SidecarCrossModulePlanResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &autonomy.SidecarCrossModulePlanResponse{
		PlanID:          "cm-plan-001",
		Goal:            req.Goal,
		GoalType:        "CROSS_MODULE",
		PrimaryModule:   req.PrimaryModule,
		PrimaryEntityID: req.PrimaryEntityID,
		InvolvedModules: req.InvolvedModules,
		ConfidenceScore: 0.95,
		DataSufficiency: true,
		OrderedSteps: []map[string]interface{}{
			{
				"step_id":           "cm-step-1",
				"step_number":       1,
				"action_type":       "shipments.update_milestone",
				"title":             "Assess Shipment State",
				"risk_level":        "LOW",
				"requires_approval": false,
				"idempotency_key":   "idemp-cm-1",
			},
		},
		ParallelGroups: [][]string{{"cm-step-1"}},
	}, nil
}

func (m *MockSidecarClient) EvaluateStateChange(ctx context.Context, req *autonomy.SidecarStateChangeEvaluationRequest) (*autonomy.SidecarStateChangeEvaluationResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &autonomy.SidecarStateChangeEvaluationResponse{
		PlanID:              req.PlanID,
		PlanHealth:          autonomy.PlanHealthHealthy,
		IsPlanValid:         true,
		MaterialityAnalysis: "Mock state change: non-material",
		RecommendedAction:   autonomy.ActionContinue,
		ConfidenceScore:     0.90,
		CorrelationID:       req.CorrelationID,
	}, nil
}

func (m *MockSidecarClient) GenerateAdaptiveReplan(ctx context.Context, req *autonomy.SidecarContinuousReplanningRequest) (*autonomy.SidecarContinuousReplanningResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &autonomy.SidecarContinuousReplanningResponse{
		NewPlanID:               "plan-adapted-002",
		Version:                 2,
		ParentPlanID:            req.PlanID,
		OrderedSteps:            []map[string]interface{}{{"step_id": "step-rev-1", "step_number": 1, "title": "Adapted Step", "action_type": "shipments.update_milestone"}},
		ProtectedCompletedSteps: []string{},
		ReplanRationale:         "Adapted via mock",
		ConfidenceScore:         0.92,
		CorrelationID:           req.CorrelationID,
	}, nil
}

func (m *MockSidecarClient) AnalyzeDecisionPoint(ctx context.Context, req *autonomy.SidecarDecisionAnalysisRequest) (*autonomy.SidecarDecisionAnalysisResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &autonomy.SidecarDecisionAnalysisResponse{
		DecisionID:            "dec-mock-001",
		OperatingMode:         "HUMAN_APPROVAL",
		Title:                 "Mock Decision Point",
		ContextSummary:        "Mock context summary",
		Facts:                 map[string]interface{}{"status": "AT_PORT"},
		Predictions:           map[string]interface{}{"eta_delay_h": 2},
		AIRecommendation:      "Approve priority discharge",
		PreparedPayload:       map[string]interface{}{"action": "dispatch"},
		Alternatives:          []map[string]interface{}{{"id": "alt-1", "title": "Wait"}},
		Confidence:            "HIGH",
		DataSufficiency:       "SUFFICIENT",
		RiskLevel:             "MEDIUM",
		RequiresHumanApproval: true,
		IsReversible:          true,
		ProvenanceSummary:     "Facts and predictions verified",
		CorrelationID:         req.CorrelationID,
	}, nil
}

func (m *MockSidecarClient) AnalyzeHumanFeedback(ctx context.Context, req *autonomy.SidecarFeedbackAnalysisRequest) (*autonomy.SidecarFeedbackAnalysisResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &autonomy.SidecarFeedbackAnalysisResponse{
		DecisionID:        req.DecisionID,
		LearnedPreference: "Operator preference recorded",
		ReplanSuggested:   false,
		Notes:             "Human feedback accepted",
		CorrelationID:     "corr-fb-mock",
	}, nil
}

func (m *MockSidecarClient) PrioritizeCommandCenterItems(ctx context.Context, req *autonomy.SidecarPrioritizeRequest) (*autonomy.SidecarPrioritizeResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	items := make([]autonomy.SidecarPrioritizedItemDTO, 0, len(req.Items))
	for idx, itm := range req.Items {
		id, _ := itm["id"].(string)
		title, _ := itm["title"].(string)
		sev, _ := itm["severity"].(string)
		entType, _ := itm["entity_type"].(string)
		entID, _ := itm["entity_id"].(string)
		cat, _ := itm["category"].(string)
		why := fmt.Sprintf("Flagged item %s: %s", id, title)
		score := 50.0
		tier := "NORMAL_OPERATIONAL"
		if cat == "CRITICAL_SAFETY_COMPLIANCE" || sev == "CRITICAL" {
			score = 95.0
			tier = "CRITICAL_SAFETY_COMPLIANCE"
		}
		items = append(items, autonomy.SidecarPrioritizedItemDTO{
			ID:                id,
			PriorityRank:      idx + 1,
			PriorityScore:     score,
			PriorityTier:      tier,
			Severity:          sev,
			EntityType:        entType,
			EntityID:          entID,
			Title:             title,
			IssueSummary:      title,
			WhyFlagged:        why,
			ActualFacts:       fmt.Sprintf("Fact for %s", id),
			PredictedImpact:   "Predicted disruption",
			RecommendedAction: "Review and approve",
			Urgency:           "HIGH",
			Source:            "AUTONOMY_ENGINE",
			RequiresHuman:     true,
		})
	}
	return &autonomy.SidecarPrioritizeResponse{
		Items:            items,
		CriticalCount:    1,
		HighCount:        0,
		ExecutiveSummary: "Mock prioritized summary",
		CorrelationID:    "cc-corr-mock",
	}, nil
}

func (m *MockSidecarClient) EvaluateOutcome(ctx context.Context, req *autonomy.SidecarOutcomeEvaluationRequest) (*autonomy.SidecarOutcomeEvaluationResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	stat := "SUCCESS"
	if req.Status != "" {
		stat = req.Status
	}
	isVer := stat == "SUCCESS"
	failCat := ""
	if stat == "FAILED" {
		failCat = "CARRIER_NON_RESPONSIVE"
		isVer = false
	}
	return &autonomy.SidecarOutcomeEvaluationResponse{
		OutcomeStatus:      stat,
		IsVerified:         isVer,
		FailureCategory:    failCat,
		EvaluationSummary:  "Actual result evaluated",
		Confidence:         "HIGH",
		ConfidenceScore:    0.92,
		ShouldCreateMemory: true,
		MemoryCandidate: &autonomy.SidecarMemoryCandidateDTO{
			OrgID:           req.OrgID,
			Scope:           "TENANT",
			Category:        "OPERATIONAL",
			MemoryType:      "RECOVERY_STRATEGY",
			Title:           "Carrier Escalation Recovery Success",
			Content:         "Carrier responds within 6h when escalated with tier-1 priority.",
			Confidence:      "HIGH",
			ConfidenceScore: 0.92,
			RecencyWeight:   1.0,
			ProvenanceType:  "AI_DERIVED",
			Status:          "ACTIVE",
		},
		CorrelationID: "corr-eval-mock",
	}, nil
}

func (m *MockSidecarClient) RetrieveRelevantMemory(ctx context.Context, req *autonomy.SidecarMemoryRetrievalRequest) (*autonomy.SidecarMemoryRetrievalResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	memories := make([]map[string]interface{}, 0)
	for _, cand := range req.CandidateMemories {
		memories = append(memories, cand)
	}
	return &autonomy.SidecarMemoryRetrievalResponse{
		RetrievedMemories:   memories,
		TotalFound:          len(memories),
		ContextSummary:      fmt.Sprintf("Found %d relevant memories for context", len(memories)),
		ProvenanceBreakdown: map[string]int{"AI_DERIVED": len(memories)},
		HasConflicts:        false,
		ConflictWarnings:    []string{},
		CorrelationID:       "corr-ret-mock",
	}, nil
}

func (m *MockSidecarClient) DetectLearnedPatterns(ctx context.Context, req *autonomy.SidecarPatternDetectionRequest) (*autonomy.SidecarPatternDetectionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &autonomy.SidecarPatternDetectionResponse{
		DetectedPatterns: []map[string]interface{}{
			{
				"pattern_id":              "pat-det-1",
				"pattern_type":            "CARRIER_BEHAVIOR",
				"entity_type":             "carrier",
				"entity_identifier":       "MAEU",
				"title":                   "Maersk Pacific Escalation",
				"description":             "Carrier responds effectively to direct operations escalation.",
				"recommended_strategy":    "Direct escalation on 4h+ delays",
				"supporting_observations": 3,
				"success_rate":            0.85,
				"confidence":              "HIGH",
			},
		},
		TotalPatterns: 1,
		Summary:       "1 pattern detected",
		CorrelationID: "corr-pat-mock",
	}, nil
}

func (m *MockSidecarClient) ResolveMemoryConflicts(ctx context.Context, req *autonomy.SidecarMemoryConflictRequest) (*autonomy.SidecarMemoryConflictResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &autonomy.SidecarMemoryConflictResponse{
		HasConflict:             false,
		ConflictExplanation:     "No contradiction with current authoritative observation",
		AuthoritativeResolution: req.NewObservation,
		RecommendedAction:       "KEEP_CURRENT",
		CorrelationID:           "corr-conf-mock",
	}, nil
}

func (m *MockSidecarClient) EvaluateGovernanceContext(ctx context.Context, req autonomy.SidecarGovernanceEvalRequest) (*autonomy.SidecarGovernanceEvalResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &autonomy.SidecarGovernanceEvalResponse{
		RiskScore:               0.15,
		RiskClass:               "LOW",
		DataSufficiency:         "SUFFICIENT",
		ConfidenceClass:         "HIGH",
		ConfidenceScore:         0.95,
		RecommendedAutonomyTier: 3,
		BlastRadiusAssessment:   "LOW",
		Explanation:             "Mock governance evaluation passed",
		SanitizedContext:        req.ContextText,
		CorrelationID:           req.CorrelationID,
	}, nil
}

func (m *MockSidecarClient) PreviewGovernancePlan(ctx context.Context, req autonomy.SidecarPlanPreviewRequest) (*autonomy.SidecarPlanPreviewResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &autonomy.SidecarPlanPreviewResponse{
		PlanID:                       req.PlanID,
		TotalSteps:                   len(req.Steps),
		ExecutableSteps:              len(req.Steps),
		ApprovalRequiredSteps:        0,
		MaxRiskClass:                 "LOW",
		EstimatedFinancialExposureUSD: 0.0,
		AffectedEntities:             []map[string]string{},
		SafetySummary:                "All steps permitted under policy",
		CorrelationID:                req.CorrelationID,
	}, nil
}



// Tests

func TestPolicyEvaluation_EmergencyStop(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{
		generateResp: &autonomy.SidecarPlanGenResponse{
			PlanID:            "plan-em-stop-1",
			Version:           1,
			Goal:              "Expedite urgent shipment",
			Module:            "shipments",
			ConfidenceScore:   0.95,
			DataSufficiency:   true,
			RiskLevel:         "LOW",
			AutonomyLevel:     autonomy.Level3ControlledExecution,
			Status:            autonomy.PlanStatusGenerated,
			OrderedSteps: []autonomy.SidecarPlanStep{
				{
					StepID:         "step-1",
					StepNumber:     1,
					ActionType:     "shipments.update_milestone",
					Title:          "Update milestone",
					IdempotencyKey: "idem-stop-1",
				},
			},
		},
		evalResp: &autonomy.SidecarPlanEvalResponse{
			IsPermitted:      false,
			PolicyDecision:   string(autonomy.DecisionBlockedEmergencyStop),
			RequiresApproval: true,
			PolicyReason:     "Emergency stop is activated",
		},
	}

	// Policy has EmergencyStop active
	_ = repo.SetPolicy(context.Background(), &autonomy.AutonomyPolicy{
		OrgID:                  1,
		Module:                 "shipments",
		AutonomyLevel:          autonomy.Level3ControlledExecution,
		AllowedActionTypes:     json.RawMessage(`["shipments.update_milestone"]`),
		MinConfidenceThreshold: 0.80,
		EmergencyStop:          true, // ACTIVE
		IsActive:               true,
	})

	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	userID := int64(10)

	_, _, err := svc.GeneratePlan(context.Background(), 1, &userID, autonomy.GeneratePlanRequest{
		Goal:          "Expedite urgent shipment",
		Module:        "shipments",
		AutonomyLevel: autonomy.Level3ControlledExecution,
	})

	assert.Error(t, err)
	assert.Equal(t, autonomy.ErrEmergencyStopActive, err)
}

func TestPolicyEvaluation_LevelExceeded(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{}

	// Policy limits to LEVEL 1 (RECOMMEND)
	_ = repo.SetPolicy(context.Background(), &autonomy.AutonomyPolicy{
		OrgID:         1,
		Module:        "bookings",
		AutonomyLevel: autonomy.Level1Recommend,
		IsActive:      true,
	})

	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	userID := int64(10)

	// User requests Level 3 (CONTROLLED_EXECUTION)
	_, _, err := svc.GeneratePlan(context.Background(), 1, &userID, autonomy.GeneratePlanRequest{
		Goal:          "Auto route booking",
		Module:        "bookings",
		AutonomyLevel: autonomy.Level3ControlledExecution,
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, autonomy.ErrAutonomyLevelExceeded)
}

func TestPlanGeneration_And_StepExecution_Controlled(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{
		generateResp: &autonomy.SidecarPlanGenResponse{
			PlanID:          "plan-ctrl-001",
			Version:         1,
			Goal:            "Update delayed shipment milestone",
			Module:          "shipments",
			ConfidenceScore: 0.92,
			DataSufficiency: true,
			RiskLevel:       "LOW",
			AutonomyLevel:   autonomy.Level3ControlledExecution,
			Status:          autonomy.PlanStatusGenerated,
			OrderedSteps: []autonomy.SidecarPlanStep{
				{
					StepID:           "step-001",
					StepNumber:       1,
					ActionType:       "shipments.update_milestone",
					Title:            "Update Milestone",
					IdempotencyKey:   "idem-step-milestone-001",
					RequiresApproval: false,
					RiskLevel:        "LOW",
					Parameters: map[string]interface{}{
						"milestone": "IN_TRANSIT",
					},
				},
			},
		},
		evalResp: &autonomy.SidecarPlanEvalResponse{
			IsPermitted:       true,
			PolicyDecision:    string(autonomy.DecisionPermitted),
			RequiresApproval:  false,
			ConfidenceAccept:  true,
			DataSufficiencyOk: true,
		},
	}

	_ = repo.SetPolicy(context.Background(), &autonomy.AutonomyPolicy{
		OrgID:                  1,
		Module:                 "shipments",
		AutonomyLevel:          autonomy.Level3ControlledExecution,
		AllowedActionTypes:     json.RawMessage(`["shipments.update_milestone"]`),
		MinConfidenceThreshold: 0.80,
		RequiresApproval:       false,
		IsActive:               true,
	})

	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	userID := int64(10)

	// 1. Generate plan at Level 3
	plan, steps, err := svc.GeneratePlan(context.Background(), 1, &userID, autonomy.GeneratePlanRequest{
		Goal:          "Update delayed shipment milestone",
		Module:        "shipments",
		AutonomyLevel: autonomy.Level3ControlledExecution,
	})

	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, autonomy.PlanStatusApproved, plan.Status) // Policy permitted directly
	assert.Len(t, steps, 1)

	// 2. Execute Step
	result, err := svc.ExecuteStep(context.Background(), 1, 10, plan.PlanID, steps[0].StepID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, autonomy.StepStatusCompleted, result.StepStatus)

	// Check plan transitioned to COMPLETED
	updatedPlan, _, err := svc.GetPlan(context.Background(), 1, plan.PlanID)
	assert.NoError(t, err)
	assert.Equal(t, autonomy.PlanStatusCompleted, updatedPlan.Status)

	// Check audit history
	audits, err := svc.GetAuditHistory(context.Background(), 1, plan.PlanID)
	assert.NoError(t, err)
	assert.NotEmpty(t, audits)
}

func TestPlan_TenantIsolation(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{
		generateResp: &autonomy.SidecarPlanGenResponse{
			PlanID:          "plan-tenant-org1",
			Version:         1,
			Goal:            "Private finance review",
			Module:          "finance",
			ConfidenceScore: 0.95,
			DataSufficiency: true,
			RiskLevel:       "LOW",
			AutonomyLevel:   autonomy.Level2Prepare,
			Status:          autonomy.PlanStatusGenerated,
			OrderedSteps:    []autonomy.SidecarPlanStep{},
		},
		evalResp: &autonomy.SidecarPlanEvalResponse{
			IsPermitted:       true,
			PolicyDecision:    string(autonomy.DecisionRequiresApproval),
			RequiresApproval:  true,
			ConfidenceAccept:  true,
			DataSufficiencyOk: true,
		},
	}

	_ = repo.SetPolicy(context.Background(), &autonomy.AutonomyPolicy{
		OrgID:                  1,
		Module:                 "finance",
		AutonomyLevel:          autonomy.Level2Prepare,
		MinConfidenceThreshold: 0.80,
		IsActive:               true,
	})

	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	userID := int64(10)

	plan, _, err := svc.GeneratePlan(context.Background(), 1, &userID, autonomy.GeneratePlanRequest{
		Goal:          "Private finance review",
		Module:        "finance",
		AutonomyLevel: autonomy.Level2Prepare,
	})
	assert.NoError(t, err)

	// Org 2 tries to access Org 1's plan
	_, _, err = svc.GetPlan(context.Background(), 2, plan.PlanID)
	assert.Error(t, err)
	assert.Equal(t, autonomy.ErrPlanNotFound, err)
}

func TestPlan_ReplanningVersionLineage(t *testing.T) {
	repo := NewMockAutonomyRepo()
	sidecar := &MockSidecarClient{
		generateResp: &autonomy.SidecarPlanGenResponse{
			PlanID:          "plan-v1",
			Version:         1,
			Goal:            "Re-route cargo",
			Module:          "shipments",
			ConfidenceScore: 0.88,
			DataSufficiency: true,
			RiskLevel:       "LOW",
			AutonomyLevel:   autonomy.Level2Prepare,
			Status:          autonomy.PlanStatusGenerated,
			OrderedSteps: []autonomy.SidecarPlanStep{
				{
					StepID:         "s-1",
					StepNumber:     1,
					ActionType:     "shipments.update_milestone",
					Title:          "Step 1",
					IdempotencyKey: "idem-v1",
				},
			},
		},
		evalResp: &autonomy.SidecarPlanEvalResponse{
			IsPermitted:       true,
			PolicyDecision:    string(autonomy.DecisionRequiresApproval),
			RequiresApproval:  true,
			ConfidenceAccept:  true,
			DataSufficiencyOk: true,
		},
		replanResp: &autonomy.SidecarReplanResponse{
			RevisedPlanID:   "plan-v1-rev-1",
			NewVersion:      2,
			ParentPlanID:    "plan-v1",
			ReplanReason:    "Severe storm on primary route",
			ChangesSummary:  "Added alternative carrier step",
			ConfidenceScore: 0.90,
			RiskLevel:       "MEDIUM",
			Status:          autonomy.PlanStatusRequiresApproval,
			UpdatedSteps: []autonomy.SidecarPlanStep{
				{
					StepID:         "s-1-done",
					StepNumber:     1,
					ActionType:     "shipments.update_milestone",
					Title:          "Step 1 Executed",
					IdempotencyKey: "idem-v1",
				},
				{
					StepID:         "s-2-alt",
					StepNumber:     2,
					ActionType:     "shipments.create_exception",
					Title:          "Step 2 Log Exception",
					IdempotencyKey: "idem-v2",
				},
			},
		},
	}

	_ = repo.SetPolicy(context.Background(), &autonomy.AutonomyPolicy{
		OrgID:                  1,
		Module:                 "shipments",
		AutonomyLevel:          autonomy.Level2Prepare,
		MinConfidenceThreshold: 0.80,
		IsActive:               true,
	})

	svc := autonomy.NewService(repo, sidecar, nil, nil, nil)
	userID := int64(10)

	// V1
	p1, _, err := svc.GeneratePlan(context.Background(), 1, &userID, autonomy.GeneratePlanRequest{
		Goal:          "Re-route cargo",
		Module:        "shipments",
		AutonomyLevel: autonomy.Level2Prepare,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, p1.Version)

	// Replan to create V2
	p2, steps2, err := svc.Replan(context.Background(), 1, &userID, p1.PlanID, "Severe storm on primary route", map[string]interface{}{"weather_delay": 12})
	assert.NoError(t, err)
	assert.Equal(t, 2, p2.Version)
	assert.True(t, p2.ParentPlanID.Valid)
	assert.Equal(t, p1.PlanID, p2.ParentPlanID.String)
	assert.True(t, p2.ReplanReason.Valid)
	assert.Equal(t, "Severe storm on primary route", p2.ReplanReason.String)
	assert.Len(t, steps2, 2)
}
