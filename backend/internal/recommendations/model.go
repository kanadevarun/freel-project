package recommendations

import (
	"encoding/json"
	"time"
)

const (
	// Categories
	CategoryCustomer    = "customer"
	CategorySales       = "sales"
	CategoryRFQ         = "rfq"
	CategoryPricing     = "pricing"
	CategoryQuotation   = "quotation"
	CategoryShipment    = "shipment"
	CategoryOperations  = "operations"
	CategoryException   = "exception"
	CategoryInvoice     = "invoice"
	CategoryFinance     = "finance"
	CategoryContract    = "contract"
	CategoryCompliance  = "compliance"
	CategoryCollections = "collections"
	CategoryDataQuality = "data_quality"
	CategoryGeneral     = "general"

	// Priorities
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"

	// Risk Levels
	RiskLow      = "low"
	RiskMedium   = "medium"
	RiskHigh     = "high"
	RiskCritical = "critical"

	// Statuses
	StatusNew       = "new"
	StatusReviewed  = "reviewed"
	StatusAssigned  = "assigned"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusDismissed = "dismissed"
	StatusCompleted = "completed"
	StatusFailed    = "failed"

	// Source Types
	SourceCustomer  = "CUSTOMER"
	SourceShipment  = "SHIPMENT"
	SourceInvoice   = "INVOICE"
	SourceRFQ       = "RFQ"
	SourceQuotation = "QUOTATION"
	SourceContract  = "CONTRACT"
	SourceDocument  = "DOCUMENT"
	SourceCompliance = "COMPLIANCE"
	SourceCarrier   = "CARRIER"
	SourceLead      = "LEAD"

	// Follow-Up & Draft Types (Phase 2 Task 2.2 & 2.3)
	FollowupTypeGeneralCheckin      = "General check-in"
	FollowupTypeRFQFollowup         = "RFQ follow-up"
	FollowupTypeQuotationFollowup   = "Quotation follow-up"
	FollowupTypeShipmentUpdate      = "Shipment update"
	FollowupTypeExceptionResolution = "Exception resolution"
	FollowupTypeInvoiceReminder     = "Invoice reminder"
	FollowupTypeContractRenewal     = "Contract renewal"
	FollowupTypeServiceRecovery     = "Service recovery"
	FollowupTypeAccountReview       = "Account review"

	// Phase 2 Task 2.3: RFQ & Quotation Assistant Draft Types
	DraftTypeRFQClarification             = "RFQ clarification request"
	DraftTypeMissingInfoRequest           = "Missing-information request"
	DraftTypeQuotationApproval            = "Quotation approval request"
	DraftTypeCustomerPricingClarification = "Customer pricing clarification"
	DraftTypeInternalPricingReview        = "Internal pricing review note"
	DraftTypeCarrierInfoRequest           = "Carrier information request"
	DraftTypeQuotationFollowup            = "Quotation follow-up"

	// Phase 2 Task 2.4: Shipment Exception & Operations Copilot Draft Types
	DraftTypeInternalOperationsNote = "Internal operations note"
	DraftTypeCustomerShipmentUpdate = "Customer shipment update"
	DraftTypeExceptionEscalation    = "Exception escalation note"
	DraftTypeCarrierClarification   = "Carrier clarification request"
	DraftTypeMissingDocumentRequest = "Missing-document request"
	DraftTypeDeliveryClarification  = "Delivery-status clarification"
	DraftTypeInternalHandoffNote    = "Internal handoff note"

	// Phase 2 Task 2.5: Invoice and Collections Assistant Draft Types
	DraftTypeFirstPaymentReminder        = "First payment reminder"
	DraftTypeOverduePaymentNotice        = "Overdue payment notice"
	DraftTypeUrgentCollectionEscalation  = "Urgent collection escalation"
	DraftTypePaymentReconciliationQuery  = "Payment reconciliation query"
	DraftTypeFinanceInternalEscalation   = "Finance internal escalation note"
	DraftTypeCustomerStatementSummary    = "Customer statement summary"

	// Phase 2 Task 2.6: Contract, Document, and Compliance Assistant Draft Types
	DraftTypeContractRenewalReminder     = "Contract renewal reminder"
	DraftTypeComplianceFollowupNotice    = "Compliance follow-up notice"
	DraftTypeInternalContractReview      = "Internal contract review request"
	DraftTypeDocumentVerificationQuery   = "Document verification query"
	DraftTypeCarrierComplianceEscalation = "Carrier compliance escalation notice"

	// Action Types
	ActionTypeRequestRFQClarification       = "REQUEST_RFQ_CLARIFICATION"
	ActionTypeAssignRFQOwner                = "ASSIGN_RFQ_OWNER"
	ActionTypeSourceRates                   = "SOURCE_RATES"
	ActionTypeReviewQuotation               = "REVIEW_QUOTATION"
	ActionTypeRequestPricingApproval        = "REQUEST_PRICING_APPROVAL"
	ActionTypeExtendQuotationValidity       = "EXTEND_QUOTATION_VALIDITY"
	ActionTypeFollowupQuote                 = "FOLLOW_UP_QUOTE"
	ActionTypeCreateFollowupTask            = "CREATE_FOLLOWUP_TASK"
	ActionTypeCreditReview                  = "CREDIT_REVIEW"
	ActionTypeInvestigateException          = "INVESTIGATE_EXCEPTION"
	ActionTypeReviewInvoice                 = "REVIEW_INVOICE"
	ActionTypeReviewContract                = "REVIEW_CONTRACT"
	ActionTypeAssignLead                    = "ASSIGN_LEAD"
	ActionTypeScheduleCheckin               = "SCHEDULE_CHECKIN"
	ActionTypeRequestCarrierClarification   = "REQUEST_CARRIER_CLARIFICATION"
	ActionTypeSendCustomerShipmentUpdate    = "SEND_CUSTOMER_SHIPMENT_UPDATE"
	ActionTypeUpdateMilestone               = "UPDATE_MILESTONE"
	ActionTypeEscalateOperationalRisk       = "ESCALATE_OPERATIONAL_RISK"
	ActionTypeAssignOperationsOwner         = "ASSIGN_OPERATIONS_OWNER"
	ActionTypeRequestMissingDocuments       = "REQUEST_MISSING_DOCUMENTS"
	// Phase 2 Task 2.5: Invoice and Collections Assistant Action Types
	ActionTypeReviewOverdueInvoice          = "REVIEW_OVERDUE_INVOICE"
	ActionTypeContactCustomerCollections    = "CONTACT_CUSTOMER_COLLECTIONS"
	ActionTypeVerifyPaymentStatus           = "VERIFY_PAYMENT_STATUS"
	ActionTypeReviewInvoiceDispute          = "REVIEW_INVOICE_DISPUTE"
	ActionTypeAssignCollectionOwner         = "ASSIGN_COLLECTION_OWNER"
	ActionTypeEscalateCollectionRisk        = "ESCALATE_COLLECTION_RISK"
	ActionTypePreparePaymentReminder        = "PREPARE_PAYMENT_REMINDER"
	ActionTypeRequestFinanceReview          = "REQUEST_FINANCE_REVIEW"
	ActionTypeReconcilePaymentInfo          = "RECONCILE_PAYMENT_INFO"
	// Phase 2 Task 2.6: Contract, Document, and Compliance Assistant Action Types
	ActionTypeReviewExpiringContract        = "REVIEW_EXPIRING_CONTRACT"
	ActionTypeAssignContractOwner           = "ASSIGN_CONTRACT_OWNER"
	ActionTypeVerifyRenewalTerms            = "VERIFY_RENEWAL_TERMS"
	ActionTypeRequestMissingDocument        = "REQUEST_MISSING_DOCUMENT"
	ActionTypeReviewComplianceChecklist     = "REVIEW_COMPLIANCE_CHECKLIST"
	ActionTypeEscalateComplianceRisk        = "ESCALATE_COMPLIANCE_RISK"
	ActionTypeReviewContractObligation      = "REVIEW_CONTRACT_OBLIGATION"
	ActionTypePrepareRenewalReminder        = "PREPARE_RENEWAL_REMINDER"
	ActionTypeSubmitLegalApproval           = "SUBMIT_LEGAL_APPROVAL"
	ActionTypeVerifyDocumentMetadata        = "VERIFY_DOCUMENT_METADATA"

	// Draft Statuses
	DraftStatusNotGenerated = "NOT_GENERATED"
	DraftStatusDrafted      = "DRAFTED"
	DraftStatusEdited       = "EDITED"
)

// AllowedTransitions maps valid status transitions
var AllowedTransitions = map[string][]string{
	StatusNew:      {StatusReviewed, StatusAssigned, StatusDismissed},
	StatusReviewed: {StatusAssigned, StatusDismissed, StatusCompleted},
	StatusAssigned: {StatusReviewed, StatusApproved, StatusRejected, StatusCompleted, StatusDismissed},
	StatusApproved: {StatusCompleted, StatusFailed},
	StatusRejected: {StatusAssigned, StatusDismissed},
}

// IsValidTransition checks whether moving from oldStatus to newStatus is permitted
func IsValidTransition(oldStatus, newStatus string) bool {
	if oldStatus == newStatus {
		return true
	}
	allowed, ok := AllowedTransitions[oldStatus]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == newStatus {
			return true
		}
	}
	return false
}

// EvidenceItem provides factual evidence for a recommendation traced to real data
type EvidenceItem struct {
	SourceModule   string      `json:"source_module"`
	SourceEntityID int64       `json:"source_entity_id"`
	SourceRef      string      `json:"source_ref"`
	FieldName      string      `json:"field_name"`
	ObservedValue  interface{} `json:"observed_value"`
	Description    string      `json:"description"`
}

// Recommendation represents a persisted, deterministic-first AI action & recommendation
type Recommendation struct {
	ID                int64           `db:"id" json:"id"`
	OrgID             int64           `db:"org_id" json:"org_id"`
	SourceType        string          `db:"source_type" json:"source_type"`
	SourceID          int64           `db:"source_id" json:"source_id"`
	SourceReference   string          `db:"source_reference" json:"source_reference"`
	Title             string          `db:"title" json:"title"`
	Description       string          `db:"description" json:"description"`
	Category          string          `db:"category" json:"category"`
	Priority          string          `db:"priority" json:"priority"`
	RiskLevel         string          `db:"risk_level" json:"risk_level"`
	Confidence        string          `db:"confidence" json:"confidence"`
	ConfidenceScore   float64         `db:"confidence_score" json:"confidence_score"`
	EvidenceJSON      string          `db:"evidence" json:"-"`
	Evidence          []EvidenceItem  `db:"-" json:"evidence"`
	RecommendedAction string          `db:"recommended_action" json:"recommended_action"`
	ActionType        string          `db:"action_type" json:"action_type"`
	Status            string          `db:"status" json:"status"`
	AssigneeID        *int64          `db:"assignee_id" json:"assignee_id,omitempty"`
	AssigneeName      *string         `db:"assignee_name" json:"assignee_name,omitempty"`
	CreatedAt         time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time       `db:"updated_at" json:"updated_at"`
	ReviewedAt        *time.Time      `db:"reviewed_at" json:"reviewed_at,omitempty"`
	ReviewedByID      *int64          `db:"reviewed_by_id" json:"reviewed_by_id,omitempty"`
	CompletedAt       *time.Time      `db:"completed_at" json:"completed_at,omitempty"`
	CompletedByID     *int64          `db:"completed_by_id" json:"completed_by_id,omitempty"`
	DismissedAt       *time.Time      `db:"dismissed_at" json:"dismissed_at,omitempty"`
	DismissedByID     *int64          `db:"dismissed_by_id" json:"dismissed_by_id,omitempty"`
	DismissedReason   *string         `db:"dismissed_reason" json:"dismissed_reason,omitempty"`
	Freshness         time.Time       `db:"freshness" json:"freshness"`
	ExpiresAt         *time.Time      `db:"expires_at" json:"expires_at,omitempty"`
	RequiresApproval  bool            `db:"requires_approval" json:"requires_approval"`
	ApprovalID        *int64          `db:"approval_id" json:"approval_id,omitempty"`
	CorrelationID     string          `db:"correlation_id" json:"correlation_id"`
	CreatedBy         string          `db:"created_by" json:"created_by"`
	GeneratedBy       string          `db:"generated_by" json:"generated_by"`
	RuleApplied       string          `db:"rule_applied" json:"rule_applied"`
	DedupHash         string          `db:"dedup_hash" json:"dedup_hash"`
	MetadataJSON      *string         `db:"metadata" json:"-"`
	Metadata          json.RawMessage `db:"-" json:"metadata,omitempty"`

	// Phase 2 Task 2.2: Customer Follow-Up Assistant extensions
	CustomerID         *int64     `db:"customer_id" json:"customer_id,omitempty"`
	CustomerName       *string    `db:"customer_name" json:"customer_name,omitempty"`
	FollowupType       *string    `db:"followup_type" json:"followup_type,omitempty"`
	SuggestedOwnerID   *int64     `db:"suggested_owner_id" json:"suggested_owner_id,omitempty"`
	SuggestedOwnerName *string    `db:"suggested_owner_name" json:"suggested_owner_name,omitempty"`
	DraftSubject       *string    `db:"draft_subject" json:"draft_subject,omitempty"`
	DraftBody          *string    `db:"draft_body" json:"draft_body,omitempty"`
	DraftStatus        string     `db:"draft_status" json:"draft_status"`
	DraftGeneratedAt   *time.Time `db:"draft_generated_at" json:"draft_generated_at,omitempty"`
	FollowupTaskID     *int64     `db:"followup_task_id" json:"followup_task_id,omitempty"`
	// Phase 2 Task 2.3: RFQ & Quotation Assistant direct linkage
	RFQID              *int64     `db:"rfq_id" json:"rfq_id,omitempty"`
	QuotationID        *int64     `db:"quotation_id" json:"quotation_id,omitempty"`
	// Phase 2 Task 2.4: Shipment Exception & Operations Copilot direct linkage
	ShipmentID         *int64     `db:"shipment_id" json:"shipment_id,omitempty"`
	MilestoneID        *int64     `db:"milestone_id" json:"milestone_id,omitempty"`
	ExceptionID        *int64     `db:"exception_id" json:"exception_id,omitempty"`
	BookingID          *int64     `db:"booking_id" json:"booking_id,omitempty"`
	// Phase 2 Task 2.5: Invoice and Collections Assistant direct linkage
	InvoiceID          *int64     `db:"invoice_id" json:"invoice_id,omitempty"`
	// Phase 2 Task 2.6: Contract, Document, and Compliance Assistant direct linkage
	ContractID         *int64     `db:"contract_id" json:"contract_id,omitempty"`
	DocumentID         *string    `db:"document_id" json:"document_id,omitempty"`
	ComplianceID       *int64     `db:"compliance_id" json:"compliance_id,omitempty"`
	CarrierID          *int64     `db:"carrier_id" json:"carrier_id,omitempty"`
	// Phase 2 Task 2.7: Workflow Automation and Scheduled AI Jobs direct linkage
	AutomationID       *int64     `db:"automation_id" json:"automation_id,omitempty"`
	ExecutionID        *int64     `db:"execution_id" json:"execution_id,omitempty"`
}

// UnmarshalDetails unpacks JSON fields for client responses
func (r *Recommendation) UnmarshalDetails() {
	if r.EvidenceJSON != "" {
		var ev []EvidenceItem
		if err := json.Unmarshal([]byte(r.EvidenceJSON), &ev); err == nil {
			r.Evidence = ev
		} else {
			r.Evidence = []EvidenceItem{}
		}
	} else {
		r.Evidence = []EvidenceItem{}
	}

	if r.MetadataJSON != nil && *r.MetadataJSON != "" {
		r.Metadata = json.RawMessage(*r.MetadataJSON)
	}

	// Canonical resolution of RFQ and Quotation linkages
	if r.RFQID == nil && r.SourceType == SourceRFQ && r.SourceID > 0 {
		id := r.SourceID
		r.RFQID = &id
	}
	if r.QuotationID == nil && r.SourceType == SourceQuotation && r.SourceID > 0 {
		id := r.SourceID
		r.QuotationID = &id
	}

	// Canonical resolution of Shipment, Milestone, and Exception linkages
	if r.ShipmentID == nil && r.SourceType == SourceShipment && r.SourceID > 0 {
		id := r.SourceID
		r.ShipmentID = &id
	}

	// Canonical resolution of Invoice linkage
	if r.InvoiceID == nil && r.SourceType == SourceInvoice && r.SourceID > 0 {
		id := r.SourceID
		r.InvoiceID = &id
	}

	// Canonical resolution of Contract & Compliance linkage (Phase 2 Task 2.6)
	if r.ContractID == nil && r.SourceType == SourceContract && r.SourceID > 0 {
		id := r.SourceID
		r.ContractID = &id
	}
	if r.ComplianceID == nil && r.SourceType == SourceCompliance && r.SourceID > 0 {
		id := r.SourceID
		r.ComplianceID = &id
	}
}

// RecommendationFilter represents search, filter, and pagination options
type RecommendationFilter struct {
	Page             int    `json:"page"`
	Limit            int    `json:"limit"`
	Category         string `json:"category"`
	Priority         string `json:"priority"`
	RiskLevel        string `json:"risk_level"`
	Status           string `json:"status"`
	SourceType       string `json:"source_type"`
	RequiresApproval *bool  `json:"requires_approval"`
	AssigneeID       *int64 `json:"assignee_id"`
	CustomerID       *int64 `json:"customer_id"`
	FollowupType     string `json:"followup_type"`
	Search           string `json:"search"`
	SortBy           string `json:"sort_by"`
	SortDir          string `json:"sort_dir"`

	// Phase 2 Task 2.3 Filters
	RFQID          *int64 `json:"rfq_id"`
	QuotationID    *int64 `json:"quotation_id"`
	MissingInfo    *bool  `json:"missing_info"`
	ExpiringSoon   *bool  `json:"expiring_soon"`
	PricingConcern *bool  `json:"pricing_concern"`

	// Phase 2 Task 2.4: Shipment Exception & Operations Copilot Filters
	ShipmentID       *int64 `json:"shipment_id"`
	MilestoneID      *int64 `json:"milestone_id"`
	ExceptionID      *int64 `json:"exception_id"`
	BookingID        *int64 `json:"booking_id"`
	DelayedMilestone *bool  `json:"delayed_milestone"`
	ActiveException  *bool  `json:"active_exception"`
	MissingOpsInfo   *bool  `json:"missing_ops_info"`

	// Phase 2 Task 2.5: Invoice and Collections Assistant Filters
	InvoiceID       *int64 `json:"invoice_id"`
	OverdueOnly     *bool  `json:"overdue_only"`
	DataQualityOnly *bool  `json:"data_quality_only"`
	UpcomingOnly    *bool  `json:"upcoming_only"`
	AgingBand       string `json:"aging_band"`

	// Phase 2 Task 2.6: Contract, Document, and Compliance Assistant Filters
	ContractID   *int64  `json:"contract_id"`
	DocumentID   *string `json:"document_id"`
	ComplianceID *int64  `json:"compliance_id"`
	CarrierID    *int64  `json:"carrier_id"`
	ExpiredOnly  *bool   `json:"expired_only"`

	// Phase 2 Task 2.7: Workflow Automation Filters
	AutomationID *int64 `json:"automation_id"`
	ExecutionID  *int64 `json:"execution_id"`
}

// ActionPreview provides structured details of a proposed controlled next step before execution
type ActionPreview struct {
	RecommendationID  int64          `json:"recommendation_id"`
	ActionType        string         `json:"action_type"`
	ProposedAction    string         `json:"proposed_action"`
	SourceType        string         `json:"source_type"`
	SourceID          int64          `json:"source_id"`
	SourceReference   string         `json:"source_reference"`
	ExpectedEffect    string         `json:"expected_effect"`
	RiskLevel         string         `json:"risk_level"`
	RequiresApproval  bool           `json:"requires_approval"`
	Evidence          []EvidenceItem `json:"evidence"`
	InitiatedByUser   string         `json:"initiated_by_user"`
	CorrelationID     string         `json:"correlation_id"`
	SuggestedNextStep string         `json:"suggested_next_step"`
}

// RequestApprovalInput carries user notes when submitting a high-risk recommendation for approval
type RequestApprovalInput struct {
	Notes string `json:"notes"`
}

// PaginationMeta carries standard pagination details
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// RecommendationStats provides summary metrics across priorities and workflow states
type RecommendationStats struct {
	TotalActive      int64 `db:"total_active" json:"total_active"`
	CriticalCount    int64 `db:"critical_count" json:"critical_count"`
	HighCount        int64 `db:"high_count" json:"high_count"`
	MediumCount      int64 `db:"medium_count" json:"medium_count"`
	LowCount         int64 `db:"low_count" json:"low_count"`
	RequiresReview   int64 `db:"requires_review" json:"requires_review"`
	RequiresApproval int64 `db:"requires_approval" json:"requires_approval"`
	AssignedCount    int64 `db:"assigned_count" json:"assigned_count"`
	CompletedCount   int64 `db:"completed_count" json:"completed_count"`
	DismissedCount   int64 `db:"dismissed_count" json:"dismissed_count"`
}

// RecommendationListResponse wraps recommendation records with pagination and stats
type RecommendationListResponse struct {
	Recommendations []*Recommendation    `json:"recommendations"`
	Pagination      PaginationMeta       `json:"pagination"`
	Stats           *RecommendationStats `json:"stats,omitempty"`
}

// UpdateStatusInput carries status change requests
type UpdateStatusInput struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// AssignInput carries assignment requests
type AssignInput struct {
	AssigneeID   int64  `json:"assignee_id"`
	AssigneeName string `json:"assignee_name"`
}

// DismissInput carries dismissal requests with mandatory reason
type DismissInput struct {
	Reason string `json:"reason"`
}

// CustomerFollowupTask represents a controlled internal task created from a follow-up recommendation
type CustomerFollowupTask struct {
	ID               int64      `db:"id" json:"id"`
	OrgID            int64      `db:"org_id" json:"org_id"`
	RecommendationID int64      `db:"recommendation_id" json:"recommendation_id"`
	CustomerID       int64      `db:"customer_id" json:"customer_id"`
	CustomerName     string     `db:"customer_name" json:"customer_name"`
	SourceType       string     `db:"source_type" json:"source_type"`
	SourceID         int64      `db:"source_id" json:"source_id"`
	SourceReference  string     `db:"source_reference" json:"source_reference"`
	FollowupType     string     `db:"followup_type" json:"followup_type"`
	Title            string     `db:"title" json:"title"`
	Reason           string     `db:"reason" json:"reason"`
	SuggestedAction  string     `db:"suggested_action" json:"suggested_action"`
	Priority         string     `db:"priority" json:"priority"`
	AssigneeID       *int64     `db:"assignee_id" json:"assignee_id,omitempty"`
	AssigneeName     *string    `db:"assignee_name" json:"assignee_name,omitempty"`
	DueDate          *time.Time `db:"due_date" json:"due_date,omitempty"`
	Status           string     `db:"status" json:"status"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
	CompletedAt      *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	CreatedByID      *int64     `db:"created_by_id" json:"created_by_id,omitempty"`
	CreatedByName    string     `db:"created_by_name" json:"created_by_name"`
	CorrelationID    string     `db:"correlation_id" json:"correlation_id"`
	Notes            *string    `db:"notes" json:"notes,omitempty"`
}

// CreateFollowupTaskInput defines input for creating a follow-up task
type CreateFollowupTaskInput struct {
	AssigneeID   *int64     `json:"assignee_id,omitempty"`
	AssigneeName *string    `json:"assignee_name,omitempty"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	Priority     string     `json:"priority,omitempty"`
	Notes        string     `json:"notes,omitempty"`
}

// SaveDraftInput defines input for saving an edited communication draft
type SaveDraftInput struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// FollowupStats summarizes follow-up pipeline metrics
type FollowupStats struct {
	TotalFollowups    int64 `db:"total_followups" json:"total_followups"`
	CriticalCount     int64 `db:"critical_count" json:"critical_count"`
	HighCount         int64 `db:"high_count" json:"high_count"`
	RequiresReview    int64 `db:"requires_review" json:"requires_review"`
	TasksCreatedCount int64 `db:"tasks_created_count" json:"tasks_created_count"`
	DraftsReadyCount  int64 `db:"drafts_ready_count" json:"drafts_ready_count"`
}

// GenerateResult summarizes recommendations produced by generation cycle
type GenerateResult struct {
	CreatedCount     int               `json:"created_count"`
	UpdatedCount     int               `json:"updated_count"`
	UnchangedCount   int               `json:"unchanged_count"`
	TotalEvaluated   int               `json:"total_evaluated"`
	RulesExecuted    []string          `json:"rules_executed"`
	CorrelationID    string            `json:"correlation_id"`
	CalculatedAt     time.Time         `json:"calculated_at"`
	TopPriorityItems []*Recommendation `json:"top_priority_items,omitempty"`
}

// InvoiceEvidencePayload represents complete evidence for an invoice (Phase 2 Task 2.5)
type InvoiceEvidencePayload struct {
	InvoiceID         int64                    `json:"invoice_id"`
	InvoiceNumber     string                   `json:"invoice_number"`
	CustomerID        int64                    `json:"customer_id"`
	CustomerName      string                   `json:"customer_name"`
	TotalAmount       float64                  `json:"total_amount"`
	PaidAmount        float64                  `json:"paid_amount"`
	BalanceDue        float64                  `json:"balance_due"`
	Currency          string                   `json:"currency"`
	Status            string                   `json:"status"`
	InvoiceDate       string                   `json:"invoice_date"`
	DueDate           string                   `json:"due_date"`
	DaysOverdue       int                      `json:"days_overdue"`
	AgingBand         string                   `json:"aging_band"` // NOT_DUE, 1-15_DAYS, 16-30_DAYS, 31-60_DAYS, OVER_60_DAYS
	IsHighValue       bool                     `json:"is_high_value"`
	PaymentsCount     int                      `json:"payments_count"`
	Payments          []map[string]interface{} `json:"payments"`
	CustomerOpenCount int                      `json:"customer_open_count"`
	CustomerTotalAR   float64                  `json:"customer_total_ar"`
	RiskFactors       []string                 `json:"risk_factors"`
	DataQualityIssues []string                 `json:"data_quality_issues"`
}

// CustomerCollectionSummaryPayload represents collections overview for a customer (Phase 2 Task 2.5)
type CustomerCollectionSummaryPayload struct {
	CustomerID              int64                    `json:"customer_id"`
	CustomerName            string                   `json:"customer_name"`
	TotalInvoiced           float64                  `json:"total_invoiced"`
	TotalPaid               float64                  `json:"total_paid"`
	TotalOutstanding        float64                  `json:"total_outstanding"`
	TotalOverdue            float64                  `json:"total_overdue"`
	OpenInvoicesCount       int                      `json:"open_invoices_count"`
	OverdueInvoicesCount    int                      `json:"overdue_invoices_count"`
	UpcomingDueCount        int                      `json:"upcoming_due_count"`
	PaymentCompletionRate   float64                  `json:"payment_completion_rate"`
	AveragePaymentDelayDays float64                  `json:"average_payment_delay_days"`
	RiskSignals             []string                 `json:"risk_signals"`
	Invoices                []map[string]interface{} `json:"invoices"`
}

// ContractEvidencePayload represents complete verifiable evidence for a contract (Phase 2 Task 2.6)
type ContractEvidencePayload struct {
	ContractID             int64                    `json:"contract_id"`
	ContractReference      string                   `json:"contract_reference"`
	ContractName           string                   `json:"contract_name"`
	ContractType           string                   `json:"contract_type"`
	PartyID                *int64                   `json:"party_id,omitempty"`
	PartyName              string                   `json:"party_name"`
	TransportMode          string                   `json:"transport_mode"`
	Status                 string                   `json:"status"`
	Currency               string                   `json:"currency"`
	ContractValue          float64                  `json:"contract_value"`
	EffectiveDate          string                   `json:"effective_date"`
	ExpiryDate             string                   `json:"expiry_date"`
	DaysUntilExpiry        int                      `json:"days_until_expiry"`
	IsExpired              bool                     `json:"is_expired"`
	IsExpiringSoon         bool                     `json:"is_expiring_soon"`
	Owner                  string                   `json:"owner"`
	HasMissingOwner        bool                     `json:"has_missing_owner"`
	DocumentsCount         int                      `json:"documents_count"`
	ObligationsCount       int                      `json:"obligations_count"`
	ComplianceReqCount     int                      `json:"compliance_req_count"`
	PendingComplianceCount int                      `json:"pending_compliance_count"`
	Documents              []map[string]interface{} `json:"documents"`
	Obligations            []map[string]interface{} `json:"obligations"`
	ComplianceReqs         []map[string]interface{} `json:"compliance_reqs"`
	RiskSignals            []string                 `json:"risk_signals"`
	DataFreshness          string                   `json:"data_freshness"`
}

// DocumentEvidencePayload represents complete verifiable evidence for a document (Phase 2 Task 2.6)
type DocumentEvidencePayload struct {
	DocumentID             string                   `json:"document_id"`
	DocumentType           string                   `json:"document_type"`
	FileName               string                   `json:"file_name"`
	Status                 string                   `json:"status"`
	ShipmentID             *int64                   `json:"shipment_id,omitempty"`
	ContractID             *int64                   `json:"contract_id,omitempty"`
	EffectiveDate          *string                  `json:"effective_date,omitempty"`
	ExpiryDate             *string                  `json:"expiry_date,omitempty"`
	DaysUntilExpiry        *int                     `json:"days_until_expiry,omitempty"`
	IsExpired              bool                     `json:"is_expired"`
	ExtractionStatus       string                   `json:"extraction_status"`
	ExtractionConfidence   float64                  `json:"extraction_confidence"`
	DiscrepancyNotes       string                   `json:"discrepancy_notes"`
	RequiresHumanReview    bool                     `json:"requires_human_review"`
	RiskSignals            []string                 `json:"risk_signals"`
	DataFreshness          string                   `json:"data_freshness"`
}

// ComplianceEvidencePayload represents complete verifiable evidence for a compliance requirement (Phase 2 Task 2.6)
type ComplianceEvidencePayload struct {
	ComplianceID           int64                    `json:"compliance_id"`
	RequirementType        string                   `json:"requirement_type"`
	Description            string                   `json:"description"`
	ContractID             int64                    `json:"contract_id"`
	ContractReference      string                   `json:"contract_reference"`
	Mandatory              bool                     `json:"mandatory"`
	Status                 string                   `json:"status"`
	DueDate                *string                  `json:"due_date,omitempty"`
	DaysUntilDue           *int                     `json:"days_until_due,omitempty"`
	IsOverdue              bool                     `json:"is_overdue"`
	RiskSignals            []string                 `json:"risk_signals"`
	DataFreshness          string                   `json:"data_freshness"`
}

// ContractComplianceSummaryPayload represents high-level metrics for contracts, documents & compliance
type ContractComplianceSummaryPayload struct {
	TotalContracts         int                      `json:"total_contracts"`
	ActiveContracts        int                      `json:"active_contracts"`
	ExpiredContracts       int                      `json:"expired_contracts"`
	ExpiringSoonContracts  int                      `json:"expiring_soon_contracts"`
	MissingOwnerContracts  int                      `json:"missing_owner_contracts"`
	TotalDocuments         int                      `json:"total_documents"`
	ExpiredDocuments       int                      `json:"expired_documents"`
	ExpiringSoonDocuments  int                      `json:"expiring_soon_documents"`
	DiscrepancyDocuments   int                      `json:"discrepancy_documents"`
	PendingComplianceReqs  int                      `json:"pending_compliance_reqs"`
	TotalActiveRisks       int                      `json:"total_active_risks"`
	ActiveRecommendations  []*Recommendation        `json:"active_recommendations,omitempty"`
}
