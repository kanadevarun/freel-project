package spec

// RFQ paths (relative to /api/v1/rfqs mount)
const (
	ListURL                 = "/"
	GetURL                  = "/{id:[0-9]+}"
	GetTimelineURL          = "/{id:[0-9]+}/timeline"
	GetAgentStatusURL       = "/{id:[0-9]+}/agent-status"
	CreateURL               = "/"
	UpdateStageURL          = "/{id:[0-9]+}/stage"
	ParseShipmentRequestURL = "/parse-shipment-request"
	AddQuoteURL             = "/{id:[0-9]+}/quotes"
	GetCarrierRatesURL      = "/{id:[0-9]+}/carrier-rates"
	ApproveQuoteURL         = "/{id:[0-9]+}/approve-quote"
)
