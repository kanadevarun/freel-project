package spec

const (
	// Shipment statuses
	BOOKING_PENDING = "BOOKING_PENDING"
	BOOKED          = "BOOKED"
	DEPARTED        = "DEPARTED"
	IN_TRANSIT      = "IN_TRANSIT"
	ARRIVED         = "ARRIVED"
	DELIVERED       = "DELIVERED"
	EXCEPTION       = "EXCEPTION"

	// Activity timeline actions
	SHIPMENT_CREATED        = "SHIPMENT_CREATED"
	SHIPMENT_STATUS_UPDATED = "SHIPMENT_STATUS_UPDATED"

	// Activity timeline exception actions
	SHIPMENT_EXCEPTION_CREATED      = "SHIPMENT_EXCEPTION_CREATED"
	SHIPMENT_EXCEPTION_ACKNOWLEDGED = "SHIPMENT_EXCEPTION_ACKNOWLEDGED"
	SHIPMENT_EXCEPTION_RESOLVED     = "SHIPMENT_EXCEPTION_RESOLVED"
	SHIPMENT_EXCEPTION_DISMISSED    = "SHIPMENT_EXCEPTION_DISMISSED"

	// Exception Statuses
	ExceptionStatusOpen         = "OPEN"
	ExceptionStatusAcknowledged = "ACKNOWLEDGED"
	ExceptionStatusInProgress   = "IN_PROGRESS"
	ExceptionStatusResolved     = "RESOLVED"
	ExceptionStatusDismissed    = "DISMISSED"

	// Exception Severities
	ExceptionSeverityLow      = "LOW"
	ExceptionSeverityMedium   = "MEDIUM"
	ExceptionSeverityHigh     = "HIGH"
	ExceptionSeverityCritical = "CRITICAL"

	// Exception Categories
	ExceptionCategoryScheduleDelay  = "SCHEDULE_DELAY"
	ExceptionCategoryETDDelay       = "ETD_DELAY"
	ExceptionCategoryETADelay       = "ETA_DELAY"
	ExceptionCategoryVesselRollover = "VESSEL_ROLLOVER"
	ExceptionCategoryPortCongestion = "PORT_CONGESTION"
	ExceptionCategoryCustomsHold    = "CUSTOMS_HOLD"
	ExceptionCategoryDocumentIssue  = "DOCUMENT_ISSUE"
	ExceptionCategoryCarrierDelay   = "CARRIER_DELAY"
	ExceptionCategoryRouteDeviation = "ROUTE_DEVIATION"
	ExceptionCategoryContainerIssue = "CONTAINER_ISSUE"
	ExceptionCategoryOther          = "OTHER"

	// Closure statuses (Task 16.6)
	ClosureStatusActive  = "ACTIVE"
	ClosureStatusReady   = "READY_FOR_CLOSURE"
	ClosureStatusClosed  = "CLOSED"
	ClosureStatusBlocked = "BLOCKED_BY_EXCEPTION"

	// Tracking states (Task 16.6)
	TrackingStateOnTrack   = "ON_TRACK"
	TrackingStateAtRisk    = "AT_RISK"
	TrackingStateDelayed   = "DELAYED"
	TrackingStateException = "EXCEPTION"
	TrackingStateCompleted = "COMPLETED"

	// Data Freshness States (Task 17.3)
	TrackingFreshnessLive        = "LIVE"
	TrackingFreshnessRecent      = "RECENT"
	TrackingFreshnessStale       = "STALE"
	TrackingFreshnessUnavailable = "UNAVAILABLE"

	// Provider Types & Modes (Task 17.6)
	ProviderTypeDemo        = "DEMO"
	ProviderTypeCarrier     = "CARRIER"
	ProviderTypeAIS         = "AIS"
	ProviderTypeDatabase    = "DATABASE"
	ProviderTypeUnavailable = "UNAVAILABLE"

	// Alert Lifecycle Statuses (Task 17.5)
	TrackingAlertStatusOpen         = "OPEN"
	TrackingAlertStatusAcknowledged = "ACKNOWLEDGED"
	TrackingAlertStatusResolved     = "RESOLVED"
	TrackingAlertStatusSuppressed   = "SUPPRESSED"

	// Tracking Refresh Triggers & Statuses (Task 17.7)
	TrackingTriggerManual    = "MANUAL"
	TrackingTriggerScheduled = "SCHEDULED"
	TrackingTriggerRetry     = "SYSTEM_RETRY"

	TrackingRunStatusStarted = "STARTED"
	TrackingRunStatusSuccess = "SUCCESS"
	TrackingRunStatusPartial = "PARTIAL"
	TrackingRunStatusFailed  = "FAILED"
	TrackingRunStatusSkipped = "SKIPPED"

	// Schedule variance states (Task 16.6)
	VarianceOnSchedule   = "ON_SCHEDULE"
	VarianceEarly        = "EARLY"
	VarianceDelayed      = "DELAYED"
	VarianceAwaitingData = "AWAITING_DATA"

	// Activity actions (Task 16.6 & 17.5)
	SHIPMENT_TRACKING_STATE_UPDATED      = "SHIPMENT_TRACKING_STATE_UPDATED"
	SHIPMENT_CLOSURE_REQUESTED           = "SHIPMENT_CLOSURE_REQUESTED"
	SHIPMENT_COMPLETED                   = "SHIPMENT_COMPLETED"
	SHIPMENT_REOPENED                    = "SHIPMENT_REOPENED"
	SHIPMENT_TRACKING_ALERT_DETECTED     = "SHIPMENT_TRACKING_ALERT_DETECTED"
	SHIPMENT_TRACKING_ALERT_ACKNOWLEDGED = "SHIPMENT_TRACKING_ALERT_ACKNOWLEDGED"
	SHIPMENT_TRACKING_ALERT_RESOLVED     = "SHIPMENT_TRACKING_ALERT_RESOLVED"
	SHIPMENT_TRACKING_ALERT_SUPPRESSED   = "SHIPMENT_TRACKING_ALERT_SUPPRESSED"
	SHIPMENT_TRACKING_REFRESHED          = "SHIPMENT_TRACKING_REFRESHED"
	SHIPMENT_TRACKING_NOTIFICATION_SENT  = "SHIPMENT_TRACKING_NOTIFICATION_SENT"

	// Document Categories (Task 16.9)
	DocCategoryTransport   = "TRANSPORT"
	DocCategoryCommercial  = "COMMERCIAL"
	DocCategoryCustoms     = "CUSTOMS"
	DocCategoryCargo       = "CARGO"
	DocCategoryInsurance   = "INSURANCE"
	DocCategoryOperational = "OPERATIONAL"
	DocCategoryOther       = "OTHER"

	// Document Types (Task 16.9)
	DocTypeBillOfLading          = "BILL_OF_LADING"
	DocTypeMasterBillOfLading    = "MBL"
	DocTypeHouseBillOfLading     = "HBL"
	DocTypeSeaWaybill            = "SEA_WAYBILL"
	DocTypeAirWaybill            = "AIR_WAYBILL"
	DocTypeArrivalNotice         = "ARRIVAL_NOTICE"
	DocTypeDeliveryOrder         = "DELIVERY_ORDER"
	DocTypeBookingConfirmation   = "BOOKING_CONFIRMATION"
	DocTypeCarrierRelease        = "CARRIER_RELEASE"
	DocTypeCommercialInvoice     = "COMMERCIAL_INVOICE"
	DocTypeProformaInvoice       = "PROFORMA_INVOICE"
	DocTypePackingList           = "PACKING_LIST"
	DocTypePurchaseOrder         = "PURCHASE_ORDER"
	DocTypeCertificateOfOrigin   = "CERTIFICATE_OF_ORIGIN"
	DocTypeCustomsDeclaration    = "CUSTOMS_DECLARATION"
	DocTypeImportPermit          = "IMPORT_PERMIT"
	DocTypeExportPermit          = "EXPORT_PERMIT"
	DocTypeDutyDocumentation     = "DUTY_DOCUMENTATION"
	DocTypeRegulatoryCertificate = "REGULATORY_CERTIFICATE"
	DocTypeCustomsClearance      = "CUSTOMS_CLEARANCE"
	DocTypeImportDocumentation   = "IMPORT_DOCUMENTATION"
	DocTypeExportDocumentation   = "EXPORT_DOCUMENTATION"
	DocTypeInspectionCert        = "INSPECTION_CERTIFICATE"
	DocTypeInsuranceCert         = "INSURANCE_CERTIFICATE"
	DocTypeInsurancePolicy       = "INSURANCE_POLICY"
	DocTypeDangerousGoods        = "DANGEROUS_GOODS_DECLARATION"
	DocTypeCargoCertificate      = "CARGO_CERTIFICATE"
	DocTypeShippingInstructions  = "SHIPPING_INSTRUCTIONS"
	DocTypeCargoManifest         = "CARGO_MANIFEST"
	DocTypeWeightCertificate     = "WEIGHT_CERTIFICATE"
	DocTypeDeliveryReceipt       = "DELIVERY_RECEIPT"
	DocTypeCarrierDocument       = "CARRIER_DOCUMENT"
	DocTypeCustomerDocument      = "CUSTOMER_DOCUMENT"
	DocTypeOther                 = "OTHER"

	// Document Lifecycle Statuses (Task 16.9)
	DocStatusMissing     = "MISSING"
	DocStatusRequested   = "REQUESTED"
	DocStatusUploaded    = "UPLOADED"
	DocStatusUnderReview = "UNDER_REVIEW"
	DocStatusApproved    = "APPROVED"
	DocStatusRejected    = "REJECTED"
	DocStatusExpired     = "EXPIRED"
	DocStatusSuperseded  = "SUPERSEDED"

	// Document Requirement Levels (Task 16.9)
	RequirementLevelCritical = "CRITICAL"
	RequirementLevelRequired = "REQUIRED"
	RequirementLevelOptional = "OPTIONAL"

	// Document Validity / Risk States (Task 16.9)
	DocValidityValid        = "VALID"
	DocValidityExpiringSoon = "EXPIRING_SOON"
	DocValidityExpired      = "EXPIRED"

	// Document Compliance States (Task 16.9)
	ComplianceStateCompliant         = "COMPLIANT"
	ComplianceStateReady             = "READY" // alias
	ComplianceStateAttentionRequired = "ATTENTION_REQUIRED"
	ComplianceStateActionRequired    = "ACTION_REQUIRED" // alias
	ComplianceStateAtRisk            = "AT_RISK"
	ComplianceStateNonCompliant      = "NON_COMPLIANT"
	ComplianceStateBlocked           = "BLOCKED"

	// Document Activity Timeline Actions (Task 16.9)
	SHIPMENT_DOCUMENT_UPLOADED  = "SHIPMENT_DOCUMENT_UPLOADED"
	SHIPMENT_DOCUMENT_ATTACHED  = "SHIPMENT_DOCUMENT_ATTACHED"
	SHIPMENT_DOCUMENT_UPDATED   = "SHIPMENT_DOCUMENT_UPDATED"
	SHIPMENT_DOCUMENT_REVIEWED  = "SHIPMENT_DOCUMENT_REVIEWED"
	SHIPMENT_DOCUMENT_APPROVED  = "SHIPMENT_DOCUMENT_APPROVED"
	SHIPMENT_DOCUMENT_REJECTED  = "SHIPMENT_DOCUMENT_REJECTED"
	SHIPMENT_DOCUMENT_REPLACED  = "SHIPMENT_DOCUMENT_REPLACED"
	SHIPMENT_DOCUMENT_DELETED   = "SHIPMENT_DOCUMENT_DELETED"
	SHIPMENT_DOCUMENT_EXPIRED   = "SHIPMENT_DOCUMENT_EXPIRED"
	SHIPMENT_COMPLIANCE_CHANGED = "SHIPMENT_COMPLIANCE_CHANGED"

	// Document Source Constants (Task 16.9)
	DocSourceShipment = "SHIPMENT"
	DocSourceRFQ      = "RFQ"
	DocSourceBooking  = "BOOKING"

	// Financial Statuses (Task 16.8)
	FinancialStatusEstimated         = "ESTIMATED"
	FinancialStatusInProgress        = "IN_PROGRESS"
	FinancialStatusPendingReview     = "PENDING_REVIEW"
	FinancialStatusProfitable        = "PROFITABLE"
	FinancialStatusLowMargin         = "LOW_MARGIN"
	FinancialStatusLoss              = "LOSS"
	FinancialStatusFinanciallyClosed = "FINANCIALLY_CLOSED"

	// Cost Categories (Task 16.8)
	CostCategoryOceanFreight       = "OCEAN_FREIGHT"
	CostCategoryAirFreight         = "AIR_FREIGHT"
	CostCategoryOriginCharges      = "ORIGIN_CHARGES"
	CostCategoryDestinationCharges = "DESTINATION_CHARGES"
	CostCategoryDocumentation      = "DOCUMENTATION"
	CostCategoryCustoms            = "CUSTOMS"
	CostCategoryInsurance          = "INSURANCE"
	CostCategoryDetention          = "DETENTION"
	CostCategoryDemurrage          = "DEMURRAGE"
	CostCategoryTrucking           = "TRUCKING"
	CostCategoryOther              = "OTHER"

	// Financial Charge Types & Statuses (Task 16.8)
	ChargeTypeCost        = "COST"
	ChargeTypeRevenue     = "REVENUE"
	ChargeStatusEstimated = "ESTIMATED"
	ChargeStatusInvoiced  = "INVOICED"
	ChargeStatusApproved  = "APPROVED"
	ChargeStatusDisputed  = "DISPUTED"
	ChargeStatusPaid      = "PAID"

	// Financial Activity Timeline Actions (Task 16.8)
	SHIPMENT_FINANCIAL_CREATED      = "SHIPMENT_FINANCIAL_CREATED"
	SHIPMENT_COST_ADDED             = "SHIPMENT_COST_ADDED"
	SHIPMENT_COST_UPDATED           = "SHIPMENT_COST_UPDATED"
	SHIPMENT_COST_REMOVED           = "SHIPMENT_COST_REMOVED"
	SHIPMENT_FINANCIAL_RECALCULATED = "SHIPMENT_FINANCIAL_RECALCULATED"
	SHIPMENT_MARGIN_CHANGED         = "SHIPMENT_MARGIN_CHANGED"
	SHIPMENT_FINANCIAL_REVIEWED     = "SHIPMENT_FINANCIAL_REVIEWED"
	SHIPMENT_FINANCIALLY_CLOSED     = "SHIPMENT_FINANCIALLY_CLOSED"
)
