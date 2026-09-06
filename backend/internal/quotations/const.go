package quotations

// Quotation lifecycle status constants
const (
	QuotationStatusDraft            = "DRAFT"
	QuotationStatusReadyForReview   = "READY_FOR_REVIEW"
	QuotationStatusChangesRequested = "CHANGES_REQUESTED"
	QuotationStatusApproved         = "APPROVED"
	QuotationStatusSent             = "SENT"
	QuotationStatusViewed           = "VIEWED"
	QuotationStatusAccepted         = "ACCEPTED"
	QuotationStatusDeclined         = "DECLINED"
	QuotationStatusRejected         = "REJECTED" // Alias for backward compatibility
	QuotationStatusExpired          = "EXPIRED"
	QuotationStatusCancelled        = "CANCELLED"
)

// Quotation approval action constants
const (
	QuotationApprovalActionSubmitted        = "SUBMITTED_FOR_REVIEW"
	QuotationApprovalActionApproved         = "APPROVED"
	QuotationApprovalActionChangesRequested = "CHANGES_REQUESTED"
	QuotationApprovalActionSent             = "SENT"
	QuotationApprovalActionViewed           = "VIEWED"
	QuotationApprovalActionAccepted         = "ACCEPTED"
	QuotationApprovalActionDeclined         = "DECLINED"
	QuotationApprovalActionExpired          = "EXPIRED"
	QuotationApprovalActionCancelled        = "CANCELLED"
)

// Quotation activity type constants
const (
	QuotationCreated                = "QUOTATION_CREATED"
	QuotationUpdated                = "QUOTATION_UPDATED"
	QuotationSubmittedForReview     = "QUOTATION_SUBMITTED_FOR_REVIEW"
	QuotationApproved               = "QUOTATION_APPROVED"
	QuotationChangesRequested       = "QUOTATION_CHANGES_REQUESTED"
	QuotationSent                   = "QUOTATION_SENT"
	QuotationViewed                 = "QUOTATION_VIEWED"
	QuotationAccepted               = "QUOTATION_ACCEPTED"
	QuotationDeclined               = "QUOTATION_DECLINED"
	QuotationRejected               = "QUOTATION_REJECTED"
	QuotationExpired                = "QUOTATION_EXPIRED"
	QuotationCancelled               = "QUOTATION_CANCELLED"
	QuotationChargeAdded            = "QUOTATION_CHARGE_ADDED"
	QuotationChargeUpdated          = "QUOTATION_CHARGE_UPDATED"
	QuotationChargeRemoved          = "QUOTATION_CHARGE_REMOVED"
	QuotationChargesReordered        = "QUOTATION_CHARGES_REORDERED"
	QuotationPricingRecalculated    = "QUOTATION_PRICING_RECALCULATED"
	QuotationRateImported           = "QUOTATION_RATE_IMPORTED"
	QuotationTemplateCreated        = "QUOTATION_TEMPLATE_CREATED"
	QuotationTemplateApplied        = "QUOTATION_TEMPLATE_APPLIED"
	QuotationCommercialTermsUpdated = "QUOTATION_COMMERCIAL_TERMS_UPDATED"
	QuotationDocumentGenerated      = "QUOTATION_DOCUMENT_GENERATED"
	QuotationPublicLinkCreated      = "QUOTATION_PUBLIC_LINK_CREATED"
	QuotationPublicLinkAccessed     = "QUOTATION_PUBLIC_LINK_ACCESSED"
	QuotationActivityLinkRevoked    = "QUOTATION_PUBLIC_LINK_REVOKED"
	QuotationActivityLinkExpired    = "QUOTATION_PUBLIC_LINK_EXPIRED"

	// Conversion & Commercial Handover activity types (Task 18.6 & Task 18.7)
	QuotationReadyForConversion     = "QUOTATION_READY_FOR_CONVERSION"
	QuotationConversionStarted      = "QUOTATION_CONVERSION_STARTED"
	QuotationConvertedToBooking     = "QUOTATION_CONVERTED_TO_BOOKING"
	QuotationConversionFailed       = "QUOTATION_CONVERSION_FAILED"

	// Task 18.7 Traceability & Handover Event Types
	QuotationBookingCreated             = "QUOTATION_BOOKING_CREATED"
	BookingConfirmedFromQuotation       = "BOOKING_CONFIRMED_FROM_QUOTATION"
	OperationalHandoverUpdated          = "OPERATIONAL_HANDOVER_UPDATED"
	QuotationOperationalChangeDetected  = "QUOTATION_OPERATIONAL_CHANGE_DETECTED"
	CommercialHandoverCompletedEvent    = "COMMERCIAL_HANDOVER_COMPLETED"
)

// Quotation conversion status constants (Task 18.6)
const (
	QuotationConversionStatusNotConverted = "NOT_CONVERTED"
	QuotationConversionStatusReady        = "READY_FOR_CONVERSION"
	QuotationConversionStatusConverted    = "CONVERTED"
	QuotationConversionStatusFailed       = "FAILED"
)

// Commercial Handover Statuses (Task 18.7)
const (
	CommercialHandoverPending            = "PENDING"
	CommercialHandoverConverted          = "CONVERTED"
	CommercialHandoverBookingConfirmed    = "BOOKING_CONFIRMED"
	CommercialHandoverOperationalChanges = "OPERATIONAL_CHANGES"
	CommercialHandoverCompleted          = "COMPLETED"
)

// Document constants (Task 18.5)
const (
	QuotationDocumentTypePDF          = "PDF"
	QuotationDocumentTypeCustomerCopy = "CUSTOMER_COPY"
)

// Public link status constants (Task 18.5)
const (
	QuotationPublicLinkActive  = "ACTIVE"
	QuotationPublicLinkRevoked = "REVOKED"
	QuotationPublicLinkExpired = "EXPIRED"
)


// Charge category constants
const (
	QuotationChargeCategoryFreight       = "FREIGHT"
	QuotationChargeCategoryOrigin        = "ORIGIN"
	QuotationChargeCategoryDestination   = "DESTINATION"
	QuotationChargeCategorySurcharge     = "SURCHARGE"
	QuotationChargeCategoryDocumentation = "DOCUMENTATION"
	QuotationChargeCategoryCustoms       = "CUSTOMS"
	QuotationChargeCategoryInsurance     = "INSURANCE"
	QuotationChargeCategoryTax           = "TAX"
	QuotationChargeCategoryOther         = "OTHER"
)

// Charge calculation basis constants
const (
	QuotationChargeBasisFlat         = "FLAT"
	QuotationChargeBasisPerContainer = "PER_CONTAINER"
	QuotationChargeBasisPerShipment  = "PER_SHIPMENT"
	QuotationChargeBasisPerWeight    = "PER_WEIGHT"
	QuotationChargeBasisPerVolume    = "PER_VOLUME"
	QuotationChargeBasisPerUnit      = "PER_UNIT"
	QuotationChargeBasisPercentage   = "PERCENTAGE"
)

// Charge type constants
const (
	QuotationChargeTypeSell = "SELL"
	QuotationChargeTypeCost = "COST"
)

// Discount type constants
const (
	QuotationDiscountTypeNone       = "NONE"
	QuotationDiscountTypeFixed      = "FIXED"
	QuotationDiscountTypePercentage = "PERCENTAGE"
)

// Margin health status constants
const (
	MarginHealthHealthy  = "HEALTHY"
	MarginHealthLow      = "LOW"
	MarginHealthNegative = "NEGATIVE"
)

// Payment terms constants
const (
	PaymentTermsPrepaid      = "PREPAID"
	PaymentTermsCollect      = "COLLECT"
	PaymentTermsNet15        = "NET_15"
	PaymentTermsNet30        = "NET_30"
	PaymentTermsNet60        = "NET_60"
	PaymentTermsDueOnReceipt = "DUE_ON_RECEIPT"
	PaymentTermsCustom       = "CUSTOM"
)

// Quotation commercial validity statuses
const (
	QuotationValidityActive       = "ACTIVE"
	QuotationValidityExpiringSoon = "EXPIRING_SOON"
	QuotationValidityExpired      = "EXPIRED"
	QuotationValidityNotSet       = "NOT_SET"
)

// Editable statuses — only these may be mutated by commercial/charge update calls
var EditableStatuses = map[string]bool{
	QuotationStatusDraft:            true,
	QuotationStatusChangesRequested: true,
}

// ── Task 19.5: Rate Integration Constants ─────────────────────────────────────

const (
	RateSourceManaged = "MANAGED_RATE"
	RateSourceSpot    = "SPOT_RATE"
	RateSourceCustom  = "CUSTOM_RATE"
)

const (
	RateEventSelected        = "RATE_SELECTED"
	RateEventReplaced        = "RATE_REPLACED"
	RateEventSnapshotCreated = "RATE_SNAPSHOT_CREATED"
	RateEventSpotSelected    = "SPOT_RATE_SELECTED"
	RateEventRemoved         = "RATE_REMOVED"
)


