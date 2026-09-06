package spec

// ──────────────────────────────────────────────────────────────────────────────
// Canonical Document Types
// ──────────────────────────────────────────────────────────────────────────────
const (
	// Commercial Documents
	DocTypeCommercialInvoice = "COMMERCIAL_INVOICE"
	DocTypePackingList       = "PACKING_LIST"
	DocTypeProformaInvoice   = "PROFORMA_INVOICE"
	DocTypePurchaseOrder     = "PURCHASE_ORDER"
	DocTypeSalesContract     = "SALES_CONTRACT"

	// Transport Documents
	DocTypeBillOfLading  = "BILL_OF_LADING"
	DocTypeHBL           = "HBL"
	DocTypeMBL           = "MBL"
	DocTypeAirWaybill    = "AIR_WAYBILL"
	DocTypeSeaWaybill    = "SEA_WAYBILL"
	DocTypeDeliveryOrder = "DELIVERY_ORDER"

	// Customs & Trade Documents
	DocTypeCertificateOfOrigin = "CERTIFICATE_OF_ORIGIN"
	DocTypeExportDeclaration   = "EXPORT_DECLARATION"
	DocTypeImportDeclaration   = "IMPORT_DECLARATION"
	DocTypeCustomsEntry        = "CUSTOMS_ENTRY"
	DocTypeCommercialLicense   = "COMMERCIAL_LICENSE"

	// Cargo & Compliance Documents
	DocTypeMSDS                      = "MSDS"
	DocTypeDangerousGoodsDeclaration = "DANGEROUS_GOODS_DECLARATION"
	DocTypeIMDGDeclaration           = "IMDG_DECLARATION"
	DocTypeCargoSpecification        = "CARGO_SPECIFICATION"
	DocTypeProductCertificate        = "PRODUCT_CERTIFICATE"
	DocTypeQualityCertificate        = "QUALITY_CERTIFICATE"
	DocTypeInspectionCertificate     = "INSPECTION_CERTIFICATE"

	// Insurance Documents
	DocTypeInsuranceCertificate = "INSURANCE_CERTIFICATE"
	DocTypeCargoInsurancePolicy = "CARGO_INSURANCE_POLICY"

	// Operational Documents
	DocTypeBookingConfirmation = "BOOKING_CONFIRMATION"
	DocTypeShippingInstruction = "SHIPPING_INSTRUCTION"
	DocTypeContainerInspection = "CONTAINER_INSPECTION"
	DocTypeVGMDeclaration      = "VGM_DECLARATION"

	// Other / Custom
	DocTypeOther = "OTHER"
)

// ──────────────────────────────────────────────────────────────────────────────
// Document Type Aliases & Normalized Inputs (Used for lookup / normalization)
// ──────────────────────────────────────────────────────────────────────────────
const (
	// Commercial Aliases
	DocAliasInvoice            = "INVOICE"
	DocAliasCI                 = "CI"
	DocAliasPL                 = "PL"
	DocAliasPackingListNoSpace = "PACKINGLIST"
	DocAliasPI                 = "PI"
	DocAliasProforma           = "PROFORMA"
	DocAliasPO                 = "PO"
	DocAliasSC                 = "SC"
	DocAliasContract           = "CONTRACT"

	// Transport Aliases
	DocAliasBL                 = "BL"
	DocAliasBOL                = "BOL"
	DocAliasHouseBL            = "HOUSE_BL"
	DocAliasHouseBillOfLading  = "HOUSE_BILL_OF_LADING"
	DocAliasMasterBL           = "MASTER_BL"
	DocAliasMasterBillOfLading = "MASTER_BILL_OF_LADING"
	DocAliasAWB                = "AWB"
	DocAliasHAWB               = "HAWB"
	DocAliasMAWB               = "MAWB"
	DocAliasSWB                = "SWB"
	DocAliasDO                 = "DO"

	// Customs Aliases
	DocAliasCOO              = "COO"
	DocAliasExportCustoms    = "EXPORT_CUSTOMS"
	DocAliasImportCustoms    = "IMPORT_CUSTOMS"
	DocAliasCustomsClearance = "CUSTOMS_CLEARANCE"

	// Cargo / Compliance Aliases
	DocAliasMaterialSafetyDataSheet = "MATERIAL_SAFETY_DATA_SHEET"
	DocAliasDGDeclaration           = "DG_DECLARATION"
	DocAliasDGD                     = "DGD"
	DocAliasIMDG                    = "IMDG"
	DocAliasSpecification           = "SPECIFICATION"

	// Insurance Aliases
	DocAliasInsurance = "INSURANCE"

	// Operational Aliases
	DocAliasBC  = "BC"
	DocAliasSI  = "SI"
	DocAliasVGM = "VGM"
)

// ──────────────────────────────────────────────────────────────────────────────
// Document Lifecycle Statuses
// ──────────────────────────────────────────────────────────────────────────────
const (
	DocStatusMissing     = "MISSING"
	DocStatusRequested   = "REQUESTED"
	DocStatusUploaded    = "UPLOADED"
	DocStatusUnderReview = "UNDER_REVIEW"
	DocStatusApproved    = "APPROVED"
	DocStatusRejected    = "REJECTED"
	DocStatusExpired     = "EXPIRED"
	DocStatusNotRequired = "NOT_REQUIRED"
)

// ──────────────────────────────────────────────────────────────────────────────
// Document Applicable Stages
// ──────────────────────────────────────────────────────────────────────────────
const (
	DocStageRFQ               = "RFQ_STAGE"
	DocStageQuotation         = "QUOTATION_STAGE"
	DocStageBookingConfirmed  = "BOOKING_CONFIRMED"
	DocStageShipmentExecution = "SHIPMENT_EXECUTION"
	DocStageCustomsClearance  = "CUSTOMS_CLEARANCE"
	DocStagePostDelivery      = "POST_DELIVERY"
)

// ──────────────────────────────────────────────────────────────────────────────
// Requirements Engine — Statuses, Severities, Categories, and Readiness
// ──────────────────────────────────────────────────────────────────────────────
const (
	ReqStatusPending       = "PENDING"
	ReqStatusSatisfied     = "SATISFIED"
	ReqStatusMissing       = "MISSING"
	ReqStatusNotApplicable = "NOT_APPLICABLE"
	ReqStatusUnderReview   = "UNDER_REVIEW"

	SeverityBlocking      = "BLOCKING"
	SeverityRequired      = "REQUIRED"
	SeverityConditional   = "CONDITIONAL"
	SeverityOptional      = "OPTIONAL"
	SeverityInformational = "INFORMATIONAL"

	CategoryShipmentInfo          = "SHIPMENT_INFO"
	CategoryCustomerInfo          = "CUSTOMER_INFO"
	CategoryCargoOperational      = "CARGO_OPERATIONAL"
	CategoryConditionalCompliance = "CONDITIONAL_COMPLIANCE"
	CategoryDocumentRequirements  = "DOCUMENT_REQUIREMENTS"
	CategoryAIFindings            = "AI_FINDINGS"

	ReadinessInformationRequired   = "INFORMATION_REQUIRED"
	ReadinessRequirementsIncomplete = "REQUIREMENTS_INCOMPLETE"
	ReadinessReadyForQuotation     = "READY_FOR_QUOTATION"
	ReadinessAttentionRequired     = "ATTENTION_REQUIRED"
	ReadinessUnderReview           = "UNDER_REVIEW"
)

// ──────────────────────────────────────────────────────────────────────────────
// Requirements Engine — Requirement IDs & Document Types (Lowercase IDs)
// ──────────────────────────────────────────────────────────────────────────────
const (
	ReqIDOrigin          = "origin"
	ReqIDDestination     = "destination"
	ReqIDIncoterms       = "incoterms"
	ReqIDTargetDate      = "target_date"
	ReqIDCustomerName    = "customer_name"
	ReqIDCustomerContact = "customer_contact"
	ReqIDCargoItems      = "cargo_items"
	ReqIDWeightConfirmed = "weight_confirmed"
	ReqIDVolumeConfirmed = "volume_confirmed"
	ReqIDTransportMode   = "transport_mode"
	ReqIDDGDeclaration   = "dg_declaration"
	ReqIDReeferTemp      = "reefer_temp"
	ReqIDContainerType   = "container_type"

	ReqDocCommercialInvoice   = "commercial_invoice"
	ReqDocPackingList         = "packing_list"
	ReqDocProformaInvoice     = "proforma_invoice"
	ReqDocBillOfLading        = "bill_of_lading"
	ReqDocHBL                 = "hbl"
	ReqDocMBL                 = "mbl"
	ReqDocCertificateOfOrigin = "certificate_of_origin"
	ReqDocAirWaybill          = "air_waybill"
	ReqDocDGDeclaration       = "dg_declaration"
)

// ──────────────────────────────────────────────────────────────────────────────
// Activity Timeline & Audit Trail Engine — Spec Constants
// ──────────────────────────────────────────────────────────────────────────────
const (
	ActivityCatCustomer     = "CUSTOMER"
	ActivityCatOperations   = "OPERATIONS"
	ActivityCatAI           = "AI"
	ActivityCatRequirements = "REQUIREMENTS"
	ActivityCatDocuments    = "DOCUMENTS"
	ActivityCatQuotes       = "QUOTES"
	ActivityCatSystem       = "SYSTEM"
)

// ActivityEventType identifies the specific action performed.
type ActivityEventType string

const (
	ActivityLeadCreated            ActivityEventType = "LEAD_CREATED"
	ActivityCustomerInquiry        ActivityEventType = "CUSTOMER_INQUIRY"
	ActivityEmailReceived          ActivityEventType = "EMAIL_RECEIVED"
	ActivityEmailSent              ActivityEventType = "EMAIL_SENT"
	ActivityAIProcessing           ActivityEventType = "AI_PROCESSING"
	ActivityAIExtraction           ActivityEventType = "AI_EXTRACTION"
	ActivityInformationMissing     ActivityEventType = "INFORMATION_MISSING"
	ActivityClarificationRequested ActivityEventType = "CLARIFICATION_REQUESTED"
	ActivityCustomerReply          ActivityEventType = "CUSTOMER_REPLY"
	ActivityLeadConverted          ActivityEventType = "LEAD_CONVERTED"
	ActivityRFQCreated             ActivityEventType = "RFQ_CREATED"
	ActivityRFQUpdated             ActivityEventType = "RFQ_UPDATED"
	ActivityRequirementsEvaluated  ActivityEventType = "REQUIREMENTS_EVALUATED"
	ActivityDocumentRequired       ActivityEventType = "DOCUMENT_REQUIRED"
	ActivityDocumentRequested      ActivityEventType = "DOCUMENT_REQUESTED"
	ActivityDocumentUploaded       ActivityEventType = "DOCUMENT_UPLOADED"
	ActivityDocumentReviewStarted  ActivityEventType = "DOCUMENT_REVIEW_STARTED"
	ActivityDocumentApproved       ActivityEventType = "DOCUMENT_APPROVED"
	ActivityDocumentRejected       ActivityEventType = "DOCUMENT_REJECTED"
	ActivityDocumentExpired        ActivityEventType = "DOCUMENT_EXPIRED"
	ActivityDocumentDeleted        ActivityEventType = "DOCUMENT_DELETED"
	ActivityQuoteGenerated         ActivityEventType = "QUOTE_GENERATED"
	ActivityQuoteRecommended       ActivityEventType = "QUOTE_RECOMMENDED"
	ActivityQuoteApproved          ActivityEventType = "QUOTE_APPROVED"
	ActivityQuoteSelected          ActivityEventType = "QUOTE_SELECTED"
	ActivityQuoteRejected          ActivityEventType = "QUOTE_REJECTED"
	ActivityQuoteWithdrawn         ActivityEventType = "QUOTE_WITHDRAWN"
	ActivityStatusChanged          ActivityEventType = "STATUS_CHANGED"
	ActivityUserAction             ActivityEventType = "USER_ACTION"
)

// Raw Timeline Action Constants (Used in database activities / timeline normalization)
const (
	ActionCreated                = "CREATED"
	ActionConverted              = "CONVERTED"
	ActionLeadConverted          = "LEAD_CONVERTED"
	ActionEmailInbound           = "EMAIL_INBOUND"
	ActionEmailOutbound          = "EMAIL_OUTBOUND"
	ActionAIParsed               = "AI_PARSED"
	ActionAIExtracted            = "AI_EXTRACTED"
	ActionAIEnriched             = "AI_ENRICHED"
	ActionRFQCreated             = "RFQ_CREATED"
	ActionQuoteGenerated         = "QUOTE_GENERATED"
	ActionDocumentUploaded       = "DOCUMENT_UPLOADED"
	ActionDocumentApproved       = "DOCUMENT_APPROVED"
	ActionDocumentRejected       = "DOCUMENT_REJECTED"
	ActionDocumentRequested      = "DOCUMENT_REQUESTED"
	ActionDocumentReviewStarted  = "DOCUMENT_REVIEW_STARTED"
	ActionDocumentExpired        = "DOCUMENT_EXPIRED"
	ActionDocumentDeleted        = "DOCUMENT_DELETED"
	ActionQuoteCreated           = "QUOTE_CREATED"
	ActionQuoteRequested         = "QUOTE_REQUESTED"
	ActionQuoteReceived          = "QUOTE_RECEIVED"
	ActionQuoteUpdated           = "QUOTE_UPDATED"
	ActionQuoteUnderReview       = "QUOTE_UNDER_REVIEW"
	ActionQuoteRecommended       = "QUOTE_RECOMMENDED"
	ActionQuoteApproved          = "QUOTE_APPROVED"
	ActionQuoteRejected          = "QUOTE_REJECTED"
	ActionQuoteSelected          = "QUOTE_SELECTED"
	ActionQuoteExpired           = "QUOTE_EXPIRED"
	ActionQuoteWithdrawn         = "QUOTE_WITHDRAWN"
	ActionQuoteDeleted           = "QUOTE_DELETED"
)

// Raw Timeline Entity / Category Constants
const (
	EntityLead      = "LEAD"
	EntityRFQ       = "RFQ"
	EntityEmail     = "EMAIL"
	EntityAI        = "AI"
	EntityQuote     = "QUOTE"
	EntityQuotes    = "QUOTES"
	EntityDocument  = "DOCUMENT"
	EntityDocuments = "DOCUMENTS"
	EntityTimeline  = "TIMELINE"
)

// ActorType identifies the category of actor that performed the event.
const (
	ActorCustomer   = "CUSTOMER"
	ActorUser       = "USER"
	ActorSystem     = "SYSTEM"
	ActorAI         = "AI"
	ActorOperations = "OPERATIONS"
)

// ──────────────────────────────────────────────────────────────────────────────
// Quote Management & Commercial Decision Intelligence — Spec Constants (Task 13)
// ──────────────────────────────────────────────────────────────────────────────
const (
	QuoteStatusDraft               = "DRAFT"
	QuoteStatusRequested           = "REQUESTED"
	QuoteStatusReceived            = "RECEIVED"
	QuoteStatusUnderReview         = "UNDER_REVIEW"
	QuoteStatusRecommended         = "RECOMMENDED"
	QuoteStatusApproved            = "APPROVED"
	QuoteStatusSelectedForCustomer = "SELECTED_FOR_CUSTOMER"
	QuoteStatusRejected            = "REJECTED"
	QuoteStatusExpired             = "EXPIRED"
	QuoteStatusWithdrawn           = "WITHDRAWN"
)

// Quote Charge Types
const (
	ChargeTypeFreight       = "FREIGHT"
	ChargeTypeOrigin        = "ORIGIN_CHARGE"
	ChargeTypeDestination   = "DESTINATION_CHARGE"
	ChargeTypeFuel          = "FUEL_SURCHARGE"
	ChargeTypeSecurity      = "SECURITY_SURCHARGE"
	ChargeTypeDocumentation = "DOCUMENTATION"
	ChargeTypeCustoms       = "CUSTOMS"
	ChargeTypeInsurance     = "INSURANCE"
	ChargeTypeHandling      = "HANDLING"
	ChargeTypeOther         = "OTHER"
)

// Quote Validity Statuses
const (
	ValidityValid        = "VALID"
	ValidityExpiringSoon = "EXPIRING_SOON"
	ValidityExpired      = "EXPIRED"
)

// ──────────────────────────────────────────────────────────────────────────────
// RFQ Commercial Closure, Booking & Shipment Handoff — Spec Constants (Task 14)
// ──────────────────────────────────────────────────────────────────────────────

// Booking Statuses
const (
	BookingStatusDraft               = "DRAFT"
	BookingStatusRequested           = "REQUESTED"
	BookingStatusPendingConfirmation = "PENDING_CONFIRMATION"
	BookingStatusConfirmed           = "CONFIRMED"
	BookingStatusCancelled           = "CANCELLED"
	BookingStatusCompleted           = "COMPLETED"
)

// Shipment Statuses
const (
	ShipmentStatusBookingPending = "BOOKING_PENDING"
	ShipmentStatusBooked         = "BOOKED"
	ShipmentStatusDeparted       = "DEPARTED"
	ShipmentStatusInTransit      = "IN_TRANSIT"
	ShipmentStatusArrived        = "ARRIVED"
	ShipmentStatusDelivered      = "DELIVERED"
	ShipmentStatusException      = "EXCEPTION"
	ShipmentStatusCompleted      = "COMPLETED"
)

// Raw Timeline Booking & Shipment Action Constants
const (
	ActionBookingCreated   = "BOOKING_CREATED"
	ActionBookingRequested = "BOOKING_REQUESTED"
	ActionBookingConfirmed = "BOOKING_CONFIRMED"
	ActionBookingCancelled = "BOOKING_CANCELLED"
	ActionBookingUpdated   = "BOOKING_UPDATED"
	ActionShipmentCreated   = "SHIPMENT_CREATED"
	ActionShipmentStarted   = "SHIPMENT_STARTED"
	ActionShipmentDeparted  = "SHIPMENT_DEPARTED"
	ActionShipmentArrived   = "SHIPMENT_ARRIVED"
	ActionShipmentDelivered = "SHIPMENT_DELIVERED"
	ActionShipmentCompleted = "SHIPMENT_COMPLETED"
)

// Raw Timeline Entity Constants for Booking & Shipment
const (
	EntityBooking   = "BOOKING"
	EntityShipment  = "SHIPMENT"
)

// ActivityEventType constants for Booking & Shipment
const (
	ActivityBookingCreated   ActivityEventType = "BOOKING_CREATED"
	ActivityBookingRequested ActivityEventType = "BOOKING_REQUESTED"
	ActivityBookingConfirmed ActivityEventType = "BOOKING_CONFIRMED"
	ActivityBookingCancelled ActivityEventType = "BOOKING_CANCELLED"
	ActivityShipmentCreated   ActivityEventType = "SHIPMENT_CREATED"
	ActivityShipmentDeparted  ActivityEventType = "SHIPMENT_DEPARTED"
	ActivityShipmentArrived   ActivityEventType = "SHIPMENT_ARRIVED"
	ActivityShipmentDelivered ActivityEventType = "SHIPMENT_DELIVERED"
	ActivityShipmentCompleted ActivityEventType = "SHIPMENT_COMPLETED"
)



