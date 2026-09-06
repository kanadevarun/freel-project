package rates

// Rate Management & Intelligence constants
const (
	// Rate statuses
	RateStatusDraft        = "DRAFT"
	RateStatusActive       = "ACTIVE"
	RateStatusExpiringSoon = "EXPIRING_SOON"
	RateStatusExpired      = "EXPIRED"
	RateStatusArchived     = "ARCHIVED"

	// Rate types
	RateTypeSpot     = "SPOT"
	RateTypeContract = "CONTRACT"
	RateTypeTariff   = "TARIFF"
	RateTypeCustom   = "CUSTOM"

	// Rate sources
	RateSourceSpotAPI     = "SPOT_API"
	RateSourceContractPDF = "CONTRACT_PDF"
	RateSourceManual      = "MANUAL"
	RateSourceEmail       = "EMAIL"

	// Extraction statuses
	ExtractionStatusConfirmed     = "CONFIRMED"
	ExtractionStatusPendingReview = "PENDING_REVIEW"
	ExtractionStatusFlagged       = "FLAGGED"
	ExtractionStatusRejected      = "REJECTED"

	// Surcharge units
	SurchargeUnitPerTEU       = "PER_TEU"
	SurchargeUnitPerContainer = "PER_CONTAINER"
	SurchargeUnitPerShipment  = "PER_SHIPMENT"
	SurchargeUnitPercent      = "PERCENT"

	// Task 19.2: Charge Categories
	ChargeCategoryFreight       = "FREIGHT"
	ChargeCategoryOrigin        = "ORIGIN"
	ChargeCategoryDestination   = "DESTINATION"
	ChargeCategorySurcharge     = "SURCHARGE"
	ChargeCategoryDocumentation = "DOCUMENTATION"
	ChargeCategoryCustoms       = "CUSTOMS"
	ChargeCategoryInsurance     = "INSURANCE"
	ChargeCategoryTax           = "TAX"
	ChargeCategoryOther         = "OTHER"

	// Task 19.2: Calculation Bases
	CalculationBasisFlat         = "FLAT"
	CalculationBasisPerContainer = "PER_CONTAINER"
	CalculationBasisPerShipment  = "PER_SHIPMENT"
	CalculationBasisPerWeight    = "PER_WEIGHT"
	CalculationBasisPerVolume    = "PER_VOLUME"
	CalculationBasisPerUnit      = "PER_UNIT"
	CalculationBasisPercentage   = "PERCENTAGE"

	// Task 19.3: Contract Statuses
	ContractStatusDraft        = "DRAFT"
	ContractStatusActive       = "ACTIVE"
	ContractStatusExpiringSoon = "EXPIRING_SOON"
	ContractStatusExpired      = "EXPIRED"
	ContractStatusArchived     = "ARCHIVED"

	// Task 19.3: Contract Renewal Statuses
	RenewalStatusNotStarted = "NOT_STARTED"
	RenewalStatusInProgress = "IN_PROGRESS"
	RenewalStatusRenewed    = "RENEWED"
	RenewalStatusNotRenewing = "NOT_RENEWING"

	// Task 19.3: Rate Version Statuses
	VersionStatusCurrent    = "CURRENT"
	VersionStatusSuperseded = "SUPERSEDED"
	VersionStatusHistorical = "HISTORICAL"
	VersionStatusDraft      = "DRAFT"

	// Task 19.3: Version History Actions
	ActionRateVersionCreated   = "RATE_VERSION_CREATED"
	ActionRateVersionUpdated   = "RATE_VERSION_UPDATED"
	ActionRateSuperseded       = "RATE_SUPERSEDED"
	ActionContractRenewed      = "CONTRACT_RENEWED"
	ActionContractExpired      = "CONTRACT_EXPIRED"

	// Task 19.4: Spot Rate Request Statuses
	SpotRequestDraft               = "DRAFT"
	SpotRequestSent                = "SENT"
	SpotRequestPartiallyResponded = "PARTIALLY_RESPONDED"
	SpotRequestResponded           = "RESPONDED"
	SpotRequestSelected            = "SELECTED"
	SpotRequestExpired             = "EXPIRED"
	SpotRequestCancelled           = "CANCELLED"

	// Task 19.4: Spot Rate Response Statuses
	SpotResponsePending  = "PENDING"
	SpotResponseReceived = "RECEIVED"
	SpotResponseDeclined = "DECLINED"
	SpotResponseExpired  = "EXPIRED"
	SpotResponseSelected = "SELECTED"

	// Task 19.4: Spot Rate Recommendation Tags
	RecommendationCheapest  = "CHEAPEST"
	RecommendationFastest   = "FASTEST"
	RecommendationBestValue = "BEST_VALUE"
	RecommendationPreferred = "PREFERRED"

	// Task 19.6: Rate Lifecycle Events
	EventRateExpiringSoon     = "RATE_EXPIRING_SOON"
	EventRateExpired          = "RATE_EXPIRED"
	EventRateSuperseded       = "RATE_SUPERSEDED"
	EventRateArchived         = "RATE_ARCHIVED"
	EventContractExpiringSoon = "CONTRACT_EXPIRING_SOON"
	EventContractExpired      = "CONTRACT_EXPIRED"
	EventContractRenewed      = "CONTRACT_RENEWED"

	// Task 19.6: Risk Severities
	SeverityInfo     = "INFO"
	SeverityWarning  = "WARNING"
	SeverityCritical = "CRITICAL"
)

