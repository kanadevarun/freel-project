package rfq

import (
	"context"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/rfq/spec"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	ListRFQsEP             endpoint.Endpoint
	GetRFQEP               endpoint.Endpoint
	GetTimelineEP          endpoint.Endpoint
	GetAgentStatusEP       endpoint.Endpoint
	CreateRFQEP            endpoint.Endpoint
	UpdateStageEP          endpoint.Endpoint
	ParseShipmentRequestEP endpoint.Endpoint
	AddQuoteEP             endpoint.Endpoint
	// GetCarrierRatesEP returns ranked carrier rates for an RFQ's trade lane.
	GetCarrierRatesEP      endpoint.Endpoint
	// ApproveQuoteEP approves a specific quote and advances the RFQ to QUOTE_SENT.
	ApproveQuoteEP         endpoint.Endpoint
	// GetRequirementsEP evaluates RFQ operational readiness against deterministic rules.
	GetRequirementsEP      endpoint.Endpoint
	// GetActivityEP aggregates and normalizes the full chronological operational timeline.
	GetActivityEP          endpoint.Endpoint
	// Document Management Endpoints (Task 12)
	GetDocumentsEP         endpoint.Endpoint
	CreateDocumentEP       endpoint.Endpoint
	UpdateDocumentStatusEP endpoint.Endpoint
	DeleteDocumentEP       endpoint.Endpoint

	// Quotes Management Endpoints (Task 13)
	GetQuotesEP         endpoint.Endpoint
	CreateQuoteEP       endpoint.Endpoint
	UpdateQuoteEP       endpoint.Endpoint
	UpdateQuoteStatusEP endpoint.Endpoint
	RecommendQuoteEP    endpoint.Endpoint
	ApproveRFQQuoteEP   endpoint.Endpoint
	SelectQuoteEP       endpoint.Endpoint
	DeleteQuoteEP       endpoint.Endpoint

	// Booking & Shipment Handoff Endpoints (Task 14)
	GetBookingHandoffEP   endpoint.Endpoint
	CreateBookingEP       endpoint.Endpoint
	UpdateBookingStatusEP endpoint.Endpoint
	GetShipmentHandoffEP  endpoint.Endpoint

	// Dedicated Booking Operations Workspace (Task 15)
	GetBookingsWorkspaceEP       endpoint.Endpoint
	GetBookingWorkspaceDetailEP  endpoint.Endpoint
	DirectUpdateBookingStatusEP  endpoint.Endpoint
	GetEligibleRFQsForBookingEP  endpoint.Endpoint
	CreateShipmentFromBookingEP  endpoint.Endpoint

	// Task 5: Live Carrier Booking Integration
	BookWithCarrierEP    endpoint.Endpoint
	SyncCarrierBookingEP endpoint.Endpoint
}

func NewAllRFQEndpoints(bl BusinessLogic) Endpoints {
	return Endpoints{
		ListRFQsEP:                  makeListRFQsEndpoint(bl),
		GetRFQEP:                    makeGetRFQEndpoint(bl),
		GetTimelineEP:               makeGetTimelineEndpoint(bl),
		GetAgentStatusEP:            makeGetAgentStatusEndpoint(bl),
		CreateRFQEP:                 makeCreateRFQEndpoint(bl),
		UpdateStageEP:               makeUpdateStageEndpoint(bl),
		ParseShipmentRequestEP:      makeParseShipmentRequestEndpoint(bl),
		AddQuoteEP:                  makeAddQuoteEndpoint(bl),
		GetCarrierRatesEP:           makeGetCarrierRatesEndpoint(bl),
		ApproveQuoteEP:              makeApproveQuoteEndpoint(bl),
		GetRequirementsEP:           makeGetRequirementsEndpoint(bl),
		GetActivityEP:               makeGetActivityEndpoint(bl),
		GetDocumentsEP:              makeGetDocumentsEndpoint(bl),
		CreateDocumentEP:            makeCreateDocumentEndpoint(bl),
		UpdateDocumentStatusEP:      makeUpdateDocumentStatusEndpoint(bl),
		DeleteDocumentEP:            makeDeleteDocumentEndpoint(bl),
		GetQuotesEP:                 makeGetQuotesEndpoint(bl),
		CreateQuoteEP:               makeCreateQuoteEndpoint(bl),
		UpdateQuoteEP:               makeUpdateQuoteEndpoint(bl),
		UpdateQuoteStatusEP:         makeUpdateQuoteStatusEndpoint(bl),
		RecommendQuoteEP:            makeRecommendQuoteEndpoint(bl),
		ApproveRFQQuoteEP:           makeApproveRFQQuoteEndpoint(bl),
		SelectQuoteEP:               makeSelectQuoteEndpoint(bl),
		DeleteQuoteEP:               makeDeleteQuoteEndpoint(bl),
		GetBookingHandoffEP:         makeGetBookingHandoffEndpoint(bl),
		CreateBookingEP:             makeCreateBookingEndpoint(bl),
		UpdateBookingStatusEP:       makeUpdateBookingStatusEndpoint(bl),
		GetShipmentHandoffEP:        makeGetShipmentHandoffEndpoint(bl),
		GetBookingsWorkspaceEP:      makeGetBookingsWorkspaceEndpoint(bl),
		GetBookingWorkspaceDetailEP: makeGetBookingWorkspaceDetailEndpoint(bl),
		DirectUpdateBookingStatusEP: makeDirectUpdateBookingStatusEndpoint(bl),
		GetEligibleRFQsForBookingEP: makeGetEligibleRFQsForBookingEndpoint(bl),
		CreateShipmentFromBookingEP: makeCreateShipmentFromBookingEndpoint(bl),
		BookWithCarrierEP:           makeBookWithCarrierEndpoint(bl),
		SyncCarrierBookingEP:        makeSyncCarrierBookingEndpoint(bl),
	}
}




func getOrgID(ctx context.Context) (int32, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		return 0, svcerror.NewServiceError(svcerror.ErrInsufficientResourceAccess)
	}
	return int32(userCtx.OrgID), nil
}

func makeListRFQsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ListRFQsRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		resp, err := bl.ListRFQs(ctx, *req)
		if err != nil {
			return nil, err
		}

		return resp, nil // Returns Data and TotalCount
	}
}

func makeGetRFQEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetRFQRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetRFQ(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		return &spec.GetRFQResponse{Data: *resp}, nil
	}
}

func makeGetTimelineEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetTimelineRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		timeline, err := bl.GetTimeline(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		return &spec.GetTimelineResponse{Data: timeline}, nil
	}
}

func makeGetAgentStatusEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetAgentStatusRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		rfq, err := bl.GetRFQ(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		return &spec.GetAgentStatusResponse{
			Data: map[string]interface{}{"agent_status": rfq.AgentStatus},
		}, nil
	}
}

func makeCreateRFQEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateRFQRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		resp, err := bl.CreateRFQ(ctx, *req)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeUpdateStageEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateStageRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.AdvanceStage(ctx, orgID, req.ID, req.Stage)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeParseShipmentRequestEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ParseShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		rawText := req.RawEmail
		if rawText == "" {
			rawText = req.RawText
		}
		return bl.ParseShipmentRequest(ctx, rawText)
	}
}

func makeAddQuoteEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.AddQuoteRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		req.Quote.RFQID = req.ID

		err = bl.AddQuote(ctx, orgID, &req.Quote)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": req.Quote}, nil
	}
}

// makeGetCarrierRatesEndpoint returns an endpoint that fetches and ranks
// carrier rate options for the given RFQ. It reads origin/destination from the
// RFQ record and calls the carrier service (FF partner API or mock).
func makeGetCarrierRatesEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetCarrierRatesRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetCarrierRates(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		return &spec.GetCarrierRatesResponse{Data: resp}, nil
	}
}

// makeApproveQuoteEndpoint returns an endpoint that approves a specific quote
// and advances the RFQ stage to QUOTE_SENT. This is the "Approve & Send" action.
func makeApproveQuoteEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ApproveQuoteRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		rfq, err := bl.ApproveQuote(ctx, orgID, req.ID, req.QuoteID)
		if err != nil {
			return nil, err
		}

		return &spec.ApproveQuoteResponse{Data: rfq}, nil
	}
}

// makeGetRequirementsEndpoint returns an endpoint that evaluates the operational
// readiness of an RFQ using deterministic business rules.
// Org isolation is enforced: orgID comes from the authenticated JWT context.
func makeGetRequirementsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetRequirementsRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetRequirements(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

// makeGetActivityEndpoint returns an endpoint that aggregates and normalizes
// the complete chronological operational timeline and audit trail for an RFQ.
// Org isolation is enforced: orgID comes from the authenticated JWT context.
func makeGetActivityEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetActivityRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetActivity(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Document Management Endpoints (Task 12)
// ──────────────────────────────────────────────────────────────────────────────

func makeGetDocumentsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetDocumentsRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetDocuments(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeCreateDocumentEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateDocumentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		uploader := "Operations Team"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.Role != "" {
			uploader = userCtx.Role
		}

		resp, err := bl.CreateDocument(ctx, orgID, req.RFQID, *req, uploader)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeUpdateDocumentStatusEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateDocumentStatusRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		reviewer := "Operations Team"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.Role != "" {
			reviewer = userCtx.Role
		}

		resp, err := bl.UpdateDocumentStatus(ctx, orgID, req.RFQID, req.DocumentID, *req, reviewer)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeDeleteDocumentEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.DeleteDocumentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		err = bl.DeleteDocument(ctx, orgID, req.RFQID, req.DocumentID)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"success": true}, nil
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Quote Management Endpoint Handlers (Task 13)
// ──────────────────────────────────────────────────────────────────────────────

func makeGetQuotesEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetQuotesRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetQuotes(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeCreateQuoteEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateQuoteRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		creator := "Operations Team"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.Role != "" {
			creator = userCtx.Role
		}

		resp, err := bl.CreateRFQQuote(ctx, orgID, req.RFQID, *req, creator)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeUpdateQuoteEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateQuoteRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		updater := "Pricing Operations"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.Role != "" {
			updater = userCtx.Role
		}

		resp, err := bl.UpdateRFQQuote(ctx, orgID, req.RFQID, req.QuoteID, *req, updater)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeUpdateQuoteStatusEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateQuoteStatusRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		updater := "Pricing Operations"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.Role != "" {
			updater = userCtx.Role
		}

		resp, err := bl.UpdateRFQQuoteStatus(ctx, orgID, req.RFQID, req.QuoteID, *req, updater)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeRecommendQuoteEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.RecommendQuoteRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		recommender := "Pricing Team"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.Role != "" {
			recommender = userCtx.Role
		}

		resp, err := bl.RecommendRFQQuote(ctx, orgID, req.RFQID, req.QuoteID, recommender)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeApproveRFQQuoteEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ApproveRFQQuoteRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		approver := "Operations Manager"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.Role != "" {
			approver = userCtx.Role
		}

		resp, err := bl.ApproveRFQQuote(ctx, orgID, req.RFQID, req.QuoteID, *req, approver)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeSelectQuoteEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.SelectQuoteRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		selector := "Commercial Operations"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.Role != "" {
			selector = userCtx.Role
		}

		resp, err := bl.SelectRFQQuoteForCustomer(ctx, orgID, req.RFQID, req.QuoteID, selector)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeDeleteQuoteEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.DeleteQuoteRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		err = bl.DeleteRFQQuote(ctx, orgID, req.RFQID, req.QuoteID)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"success": true}, nil
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Task 14: Booking & Shipment Handoff Endpoint Implementations
// ──────────────────────────────────────────────────────────────────────────────

func makeGetBookingHandoffEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetBookingHandoffRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetBookingHandoff(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeCreateBookingEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateBookingRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		creator := "Operations Team"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.Role != "" {
			creator = userCtx.Role
		}

		resp, err := bl.CreateBookingFromRFQ(ctx, orgID, req.RFQID, *req, creator)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeUpdateBookingStatusEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateBookingStatusRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		updater := "Operations Team"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.Role != "" {
			updater = userCtx.Role
		}

		resp, err := bl.UpdateBookingStatus(ctx, orgID, req.RFQID, req.BookingID, *req, updater)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeGetShipmentHandoffEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentHandoffRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetShipmentHandoff(ctx, orgID, req.ID)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Task 15: Dedicated Booking Workspace Endpoints
// ──────────────────────────────────────────────────────────────────────────────

func makeGetBookingsWorkspaceEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.BookingListFilter)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetBookingsWorkspace(ctx, orgID, *req)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"data":        resp.Bookings,
			"items":       resp.Bookings,
			"kpis":        resp.KPIs,
			"pagination":  resp.Pagination,
			"carriers":    resp.Carriers,
			"total":       resp.Pagination.TotalItems,
			"total_count": resp.Pagination.TotalItems,
			"page":        resp.Pagination.CurrentPage,
			"limit":       resp.Pagination.PageSize,
		}, nil
	}
}

func makeGetBookingWorkspaceDetailEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetBookingWorkspaceDetailRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetBookingWorkspaceDetail(ctx, orgID, req.BookingID)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeDirectUpdateBookingStatusEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.DirectUpdateBookingStatusRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		updater := "Carrier Operations"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.Role != "" {
			updater = userCtx.Role
		}

		resp, err := bl.DirectUpdateBookingStatus(ctx, orgID, req.BookingID, *req, updater)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeGetEligibleRFQsForBookingEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetEligibleRFQsForBooking(ctx, orgID)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeCreateShipmentFromBookingEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateShipmentFromBookingRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		creator := "Logistics Operations"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.Role != "" {
			creator = userCtx.Role
		}

		resp, err := bl.CreateShipmentFromBooking(ctx, orgID, req.BookingID, *req, creator)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"data": resp}, nil
	}
}

func makeBookWithCarrierEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.BookWithCarrierRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		user := "Logistics Operator"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.Role != "" {
			user = userCtx.Role
		}

		resp, err := bl.BookWithCarrier(ctx, orgID, req.BookingID, *req, user)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"message": "Carrier booking confirmed successfully",
			"data":    resp,
		}, nil
	}
}

func makeSyncCarrierBookingEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.SyncCarrierBookingRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		user := "Logistics Operator"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.Role != "" {
			user = userCtx.Role
		}

		resp, err := bl.SyncCarrierBooking(ctx, orgID, req.BookingID, user)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"message": "Carrier booking synchronized successfully",
			"data":    resp,
		}, nil
	}
}





