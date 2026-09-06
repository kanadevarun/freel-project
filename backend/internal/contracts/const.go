package contracts

// ContractStatus represents the lifecycle state of a commercial contract.
type ContractStatus string

const (
	ContractStatusDraft      ContractStatus = "DRAFT"
	ContractStatusSubmitted  ContractStatus = "SUBMITTED"
	ContractStatusApproved   ContractStatus = "APPROVED"
	ContractStatusActive     ContractStatus = "ACTIVE"
	ContractStatusExpired    ContractStatus = "EXPIRED"
	ContractStatusWithdrawn  ContractStatus = "WITHDRAWN"
	ContractStatusCancelled  ContractStatus = "CANCELLED"
	ContractStatusTerminated ContractStatus = "TERMINATED"
	ContractStatusArchived   ContractStatus = "ARCHIVED"
)

// PartyType defines the type of a contract party.
type PartyType string

const (
	PartyTypeCustomer PartyType = "CUSTOMER"
	PartyTypeCarrier  PartyType = "CARRIER"
	PartyTypeVendor   PartyType = "VENDOR"
	PartyTypeOther    PartyType = "OTHER"
)

// LinkedEntityType defines the target entity being linked
type LinkedEntityType string

const (
	EntityTypeRate             LinkedEntityType = "RATE"
	EntityTypeRateContract     LinkedEntityType = "RATE_CONTRACT"
	EntityTypeQuotation        LinkedEntityType = "QUOTATION"
	EntityTypeSpotRateRequest  LinkedEntityType = "SPOT_RATE_REQUEST"
	EntityTypeSpotRateResponse LinkedEntityType = "SPOT_RATE_RESPONSE"
	EntityTypeCustomer         LinkedEntityType = "CUSTOMER"
	EntityTypeCarrier          LinkedEntityType = "CARRIER"
	EntityTypeVendor           LinkedEntityType = "VENDOR"
)

// LinkType defines the nature of the relationship
type LinkType string

const (
	LinkTypePrimary        LinkType = "PRIMARY"
	LinkTypeCommercialRate LinkType = "COMMERCIAL_RATE"
	LinkTypeContractRate   LinkType = "CONTRACT_RATE"
	LinkTypeQuotation      LinkType = "QUOTATION"
	LinkTypeSpotSourcing   LinkType = "SPOT_SOURCING"
	LinkTypeCustomer       LinkType = "CUSTOMER"
	LinkTypeCarrier        LinkType = "CARRIER"
	LinkTypeVendor         LinkType = "VENDOR"
	LinkTypeRelated        LinkType = "RELATED"
)

// EventType defines the type of lifecycle event.
type EventType string

const (
	EventContractCreated      EventType = "CONTRACT_CREATED"
	EventContractUpdated      EventType = "CONTRACT_UPDATED"
	EventContractStatusChange EventType = "STATUS_CHANGE"
	EventContractArchived     EventType = "CONTRACT_ARCHIVED"
)

// LifecycleCondition represents deterministic contract health state
type LifecycleCondition string

const (
	LifecycleConditionActive            LifecycleCondition = "ACTIVE"
	LifecycleConditionExpiringSoon      LifecycleCondition = "EXPIRING_SOON"
	LifecycleConditionExpired           LifecycleCondition = "EXPIRED"
	LifecycleConditionRenewalRequired   LifecycleCondition = "RENEWAL_REQUIRED"
	LifecycleConditionRenewalInProgress LifecycleCondition = "RENEWAL_IN_PROGRESS"
	LifecycleConditionRenewed           LifecycleCondition = "RENEWED"
	LifecycleConditionSuperseded        LifecycleCondition = "SUPERSEDED"
	LifecycleConditionArchived          LifecycleCondition = "ARCHIVED"
	LifecycleConditionDraft             LifecycleCondition = "DRAFT"
)

// RiskSeverity defines the severity levels of lifecycle risk events.
type RiskSeverity string

const (
	SeverityCritical  RiskSeverity = "CRITICAL"
	SeverityWarning   RiskSeverity = "WARNING"
	SeverityAttention RiskSeverity = "ATTENTION"
	SeverityInfo      RiskSeverity = "INFO"
)

// RiskType defines standard commercial risk categories.
type RiskType string

const (
	RiskExpiringActiveRates RiskType = "EXPIRING_WITH_ACTIVE_RATES"
	RiskExpiredActiveRates  RiskType = "EXPIRED_WITH_ACTIVE_RATES"
	RiskExpiredDraftQuotes  RiskType = "EXPIRED_WITH_DRAFT_QUOTES"
	RiskRenewalOverdue      RiskType = "RENEWAL_OVERDUE"
	RiskOrphanedCarrierSLA  RiskType = "ORPHANED_CARRIER_SLA"
	RiskUnlinkedCounterpart RiskType = "UNLINKED_COUNTERPARTY"
)

// RenewalStatus defines renewal pipeline tracking state.
type RenewalStatus string

const (
	RenewalStatusNone        RenewalStatus = "NONE"
	RenewalStatusNotStarted  RenewalStatus = "NOT_STARTED"
	RenewalStatusRequired    RenewalStatus = "REQUIRED"
	RenewalStatusInProgress  RenewalStatus = "IN_PROGRESS"
	RenewalStatusNegotiating RenewalStatus = "NEGOTIATING"
	RenewalStatusApproved    RenewalStatus = "APPROVED"
	RenewalStatusRejected    RenewalStatus = "REJECTED"
	RenewalStatusCompleted   RenewalStatus = "COMPLETED"
	RenewalStatusRenewed     RenewalStatus = "RENEWED"
	RenewalStatusAbandoned   RenewalStatus = "ABANDONED"
	RenewalStatusCancelled   RenewalStatus = "CANCELLED"
)

// VersionStatus represents immutable contract version status.
type VersionStatus string

const (
	VersionStatusDraft           VersionStatus = "DRAFT"
	VersionStatusPendingApproval VersionStatus = "PENDING_APPROVAL"
	VersionStatusApproved        VersionStatus = "APPROVED"
	VersionStatusEffective       VersionStatus = "EFFECTIVE"
	VersionStatusSuperseded      VersionStatus = "SUPERSEDED"
	VersionStatusRejected        VersionStatus = "REJECTED"
)

// AmendmentStatus represents the approval and implementation state of a contract amendment.
type AmendmentStatus string

const (
	AmendmentStatusDraft       AmendmentStatus = "DRAFT"
	AmendmentStatusSubmitted   AmendmentStatus = "SUBMITTED"
	AmendmentStatusUnderReview AmendmentStatus = "UNDER_REVIEW"
	AmendmentStatusApproved    AmendmentStatus = "APPROVED"
	AmendmentStatusRejected    AmendmentStatus = "REJECTED"
	AmendmentStatusCancelled   AmendmentStatus = "CANCELLED"
	AmendmentStatusImplemented AmendmentStatus = "IMPLEMENTED"
)

// AmendmentType defines the commercial domain of an amendment.
type AmendmentType string

const (
	AmendmentTypeCommercialTerms   AmendmentType = "COMMERCIAL_TERMS"
	AmendmentTypeRateRevision      AmendmentType = "RATE_REVISION"
	AmendmentTypeScopeExtension    AmendmentType = "SCOPE_EXTENSION"
	AmendmentTypeClauseUpdate      AmendmentType = "CLAUSE_UPDATE"
	AmendmentTypeValidityExtension AmendmentType = "VALIDITY_EXTENSION"
)

// ApprovalStatus represents state of an approval request.
type ApprovalStatus string

const (
	ApprovalStatusPending   ApprovalStatus = "PENDING"
	ApprovalStatusApproved  ApprovalStatus = "APPROVED"
	ApprovalStatusRejected  ApprovalStatus = "REJECTED"
	ApprovalStatusCancelled ApprovalStatus = "CANCELLED"
)

// ApprovalType defines whether the request targets a Version or an Amendment.
type ApprovalType string

const (
	ApprovalTypeVersion   ApprovalType = "VERSION"
	ApprovalTypeAmendment ApprovalType = "AMENDMENT"
)

// IntelligenceEventType defines immutable lifecycle intelligence audit events.
type IntelligenceEventType string

const (
	IntelEventExpiryApproaching     IntelligenceEventType = "EXPIRY_APPROACHING"
	IntelEventCriticalExpiryRisk    IntelligenceEventType = "CRITICAL_EXPIRY_RISK"
	IntelEventExpiredDetected       IntelligenceEventType = "EXPIRED_DETECTED"
	IntelEventRenewalRecommended    IntelligenceEventType = "RENEWAL_RECOMMENDED"
	IntelEventRenewalStarted        IntelligenceEventType = "RENEWAL_STARTED"
	IntelEventRenewalStatusUpdated  IntelligenceEventType = "RENEWAL_STATUS_UPDATED"
	IntelEventRenewalCompleted      IntelligenceEventType = "RENEWAL_COMPLETED"
	IntelEventSupersededDetected    IntelligenceEventType = "SUPERSEDED_DETECTED"
	IntelEventRiskResolved          IntelligenceEventType = "RISK_RESOLVED"
)
