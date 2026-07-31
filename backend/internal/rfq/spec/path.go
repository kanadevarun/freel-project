package spec

// RFQ paths
const (
	// swagger:operation GET /rfqs RFQ ListRFQs
	// ---
	// summary: List RFQs
	ListURL = "/api/v1/rfqs"

	// swagger:operation GET /rfqs/{id} RFQ GetRFQ
	// ---
	// summary: Get RFQ
	GetURL = "/api/v1/rfqs/{id:[0-9]+}"

	// swagger:operation GET /rfqs/{id}/timeline RFQ GetTimeline
	// ---
	// summary: Get RFQ Timeline
	GetTimelineURL = "/api/v1/rfqs/{id:[0-9]+}/timeline"

	// swagger:operation GET /rfqs/{id}/agent-status RFQ GetAgentStatus
	// ---
	// summary: Get Agent Status
	GetAgentStatusURL = "/api/v1/rfqs/{id:[0-9]+}/agent-status"

	// swagger:operation POST /rfqs RFQ CreateRFQ
	// ---
	// summary: Create RFQ
	CreateURL = "/api/v1/rfqs"

	// swagger:operation PUT /rfqs/{id}/stage RFQ UpdateStage
	// ---
	// summary: Update RFQ Stage
	UpdateStageURL = "/api/v1/rfqs/{id:[0-9]+}/stage"

	// swagger:operation POST /rfqs/parse-shipment-request RFQ ParseShipmentRequest
	// ---
	// summary: Parse Shipment Request via AI
	ParseShipmentRequestURL = "/api/v1/rfqs/parse-shipment-request"

	// swagger:operation POST /rfqs/{id}/quotes RFQ AddQuote
	// ---
	// summary: Add Quote to RFQ
	AddQuoteURL = "/api/v1/rfqs/{id:[0-9]+}/quotes"
)
