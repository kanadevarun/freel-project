package rfq

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/freel/backend/internal/rfq/spec"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-chi/chi/v5"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// AddRFQHandlers adds the handlers to the rest methods for the rfq module
func AddRFQHandlers(
	router chi.Router,
	endpoints Endpoints,
	authMiddleware func(http.Handler) http.Handler,
) {
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	// List RFQs
	router.With(authMiddleware).Get("/", kitHttp.NewServer(
		endpoints.ListRFQsEP,
		decodeListRFQsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Create RFQ
	router.With(authMiddleware).Post("/", kitHttp.NewServer(
		endpoints.CreateRFQEP,
		decodeCreateRFQRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Parse Shipment Request
	router.With(authMiddleware).Post("/parse-shipment-request", kitHttp.NewServer(
		endpoints.ParseShipmentRequestEP,
		decodeParseShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Get RFQ
	router.With(authMiddleware).Get("/{id:[0-9]+}", kitHttp.NewServer(
		endpoints.GetRFQEP,
		decodeGetRFQRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Get Timeline
	router.With(authMiddleware).Get("/{id:[0-9]+}/timeline", kitHttp.NewServer(
		endpoints.GetTimelineEP,
		decodeGetTimelineRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Get Agent Status
	router.With(authMiddleware).Get("/{id:[0-9]+}/agent-status", kitHttp.NewServer(
		endpoints.GetAgentStatusEP,
		decodeGetAgentStatusRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Update Stage
	router.With(authMiddleware).Put("/{id:[0-9]+}/stage", kitHttp.NewServer(
		endpoints.UpdateStageEP,
		decodeUpdateStageRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Add Quote
	router.With(authMiddleware).Post("/{id:[0-9]+}/quotes", kitHttp.NewServer(
		endpoints.AddQuoteEP,
		decodeAddQuoteRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Get Carrier Rates — returns ranked carrier options for an RFQ trade lane.
	router.With(authMiddleware).Get("/{id:[0-9]+}/carrier-rates", kitHttp.NewServer(
		endpoints.GetCarrierRatesEP,
		decodeGetCarrierRatesRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Approve Quote — Pricing team action: approve a quote and advance RFQ to QUOTE_SENT.
	router.With(authMiddleware).Post("/{id:[0-9]+}/approve-quote", kitHttp.NewServer(
		endpoints.ApproveQuoteEP,
		decodeApproveQuoteRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Get Requirements — returns deterministic operational readiness evaluation.
	router.With(authMiddleware).Get("/{id:[0-9]+}/requirements", kitHttp.NewServer(
		endpoints.GetRequirementsEP,
		decodeGetRequirementsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Get Activity — returns complete aggregated operational timeline and audit trail.
	router.With(authMiddleware).Get("/{id:[0-9]+}/activity", kitHttp.NewServer(
		endpoints.GetActivityEP,
		decodeGetActivityRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Get Documents — returns all resolved document requirements and active records for an RFQ.
	router.With(authMiddleware).Get("/{id:[0-9]+}/documents", kitHttp.NewServer(
		endpoints.GetDocumentsEP,
		decodeGetDocumentsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Create Document — attaches or uploads a document for an RFQ.
	router.With(authMiddleware).Post("/{id:[0-9]+}/documents", kitHttp.NewServer(
		endpoints.CreateDocumentEP,
		decodeCreateDocumentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Update Document Status — transitions a document lifecycle state.
	router.With(authMiddleware).Patch("/{id:[0-9]+}/documents/{documentId:[0-9]+}/status", kitHttp.NewServer(
		endpoints.UpdateDocumentStatusEP,
		decodeUpdateDocumentStatusRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Delete Document — deletes a document record for an RFQ.
	router.With(authMiddleware).Delete("/{id:[0-9]+}/documents/{documentId:[0-9]+}", kitHttp.NewServer(
		endpoints.DeleteDocumentEP,
		decodeDeleteDocumentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// ── Quote Management Endpoints (Task 13) ──────────────────────────────────
	// Get Quotes — retrieves all quotes and comparison intelligence for an RFQ.
	router.With(authMiddleware).Get("/{id:[0-9]+}/quotes", kitHttp.NewServer(
		endpoints.GetQuotesEP,
		decodeGetQuotesRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Create Quote — creates a persistent carrier quotation.
	router.With(authMiddleware).Post("/{id:[0-9]+}/quotes", kitHttp.NewServer(
		endpoints.CreateQuoteEP,
		decodeCreateRFQQuoteRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Update Quote — updates commercial details of an existing quote.
	router.With(authMiddleware).Patch("/{id:[0-9]+}/quotes/{quoteId:[0-9]+}", kitHttp.NewServer(
		endpoints.UpdateQuoteEP,
		decodeUpdateQuoteRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Update Quote Status — transitions lifecycle status of a quote.
	router.With(authMiddleware).Patch("/{id:[0-9]+}/quotes/{quoteId:[0-9]+}/status", kitHttp.NewServer(
		endpoints.UpdateQuoteStatusEP,
		decodeUpdateQuoteStatusRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Recommend Quote — marks a quote as the recommended option.
	router.With(authMiddleware).Post("/{id:[0-9]+}/quotes/{quoteId:[0-9]+}/recommend", kitHttp.NewServer(
		endpoints.RecommendQuoteEP,
		decodeRecommendQuoteRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Approve Quote — approves quote for operations and advances RFQ stage.
	router.With(authMiddleware).Post("/{id:[0-9]+}/quotes/{quoteId:[0-9]+}/approve", kitHttp.NewServer(
		endpoints.ApproveRFQQuoteEP,
		decodeApproveRFQQuoteRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Select Quote — selects an approved quote for customer presentation.
	router.With(authMiddleware).Post("/{id:[0-9]+}/quotes/{quoteId:[0-9]+}/select", kitHttp.NewServer(
		endpoints.SelectQuoteEP,
		decodeSelectQuoteRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Delete Quote — withdraws or deletes a carrier quote.
	router.With(authMiddleware).Delete("/{id:[0-9]+}/quotes/{quoteId:[0-9]+}", kitHttp.NewServer(
		endpoints.DeleteQuoteEP,
		decodeDeleteQuoteRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// ── Booking & Shipment Handoff Endpoints (Task 14) ─────────────────────────
	// Get Booking Handoff — returns eligibility, summary, and linked bookings.
	router.With(authMiddleware).Get("/{id:[0-9]+}/bookings", kitHttp.NewServer(
		endpoints.GetBookingHandoffEP,
		decodeGetBookingHandoffRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Create Booking — creates a booking from the approved RFQ quotation.
	router.With(authMiddleware).Post("/{id:[0-9]+}/bookings", kitHttp.NewServer(
		endpoints.CreateBookingEP,
		decodeCreateBookingRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Update Booking Status — transitions booking lifecycle state.
	router.With(authMiddleware).Patch("/{id:[0-9]+}/bookings/{bookingId:[0-9]+}/status", kitHttp.NewServer(
		endpoints.UpdateBookingStatusEP,
		decodeUpdateBookingStatusRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Get Shipment Handoff — returns active shipment execution status and containers.
	router.With(authMiddleware).Get("/{id:[0-9]+}/shipments", kitHttp.NewServer(
		endpoints.GetShipmentHandoffEP,
		decodeGetShipmentHandoffRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)
}

// AddBookingsWorkspaceHandlers mounts the dedicated /api/v1/bookings workspace routes.
func AddBookingsWorkspaceHandlers(
	router chi.Router,
	endpoints Endpoints,
	authMiddleware func(http.Handler) http.Handler,
) {
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	// GET /api/v1/bookings — List bookings with search, filter, pagination, and live KPIs
	router.With(authMiddleware).Get("/", kitHttp.NewServer(
		endpoints.GetBookingsWorkspaceEP,
		decodeGetBookingsWorkspaceRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/bookings/eligible-rfqs — Eligible RFQs for + Create Booking modal
	router.With(authMiddleware).Get("/eligible-rfqs", kitHttp.NewServer(
		endpoints.GetEligibleRFQsForBookingEP,
		decodeGetEligibleRFQsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/bookings/{bookingId} — Full Booking Detail Workspace
	router.With(authMiddleware).Get("/{bookingId:[0-9]+}", kitHttp.NewServer(
		endpoints.GetBookingWorkspaceDetailEP,
		decodeGetBookingWorkspaceDetailRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// PATCH /api/v1/bookings/{bookingId}/status — Direct Booking Status Transition
	router.With(authMiddleware).Patch("/{bookingId:[0-9]+}/status", kitHttp.NewServer(
		endpoints.DirectUpdateBookingStatusEP,
		decodeDirectUpdateBookingStatusRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/bookings/{bookingId}/shipments — Create Shipment Handoff
	router.With(authMiddleware).Post("/{bookingId:[0-9]+}/shipments", kitHttp.NewServer(
		endpoints.CreateShipmentFromBookingEP,
		decodeCreateShipmentFromBookingRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/bookings/{bookingId}/carrier-book — Live Carrier Booking Submission (Task 5)
	router.With(authMiddleware).Post("/{bookingId:[0-9]+}/carrier-book", kitHttp.NewServer(
		endpoints.BookWithCarrierEP,
		decodeBookWithCarrierRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/bookings/{bookingId}/carrier-sync — Sync Carrier Booking Status (Task 5)
	router.With(authMiddleware).Post("/{bookingId:[0-9]+}/carrier-sync", kitHttp.NewServer(
		endpoints.SyncCarrierBookingEP,
		decodeSyncCarrierBookingRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)
}





func getIDFromVars(r *http.Request) (int32, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		vars := mux.Vars(r)
		idStr = vars["id"]
	}
	if idStr == "" {
		return 0, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return 0, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return int32(id), nil
}

func decodeListRFQsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	return &spec.ListRFQsRequest{
		Limit:  limit,
		Offset: offset,
	}, nil
}

func decodeGetRFQRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetRFQRequest{ID: id}, nil
}

func decodeGetTimelineRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetTimelineRequest{ID: id}, nil
}

func decodeGetAgentStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetAgentStatusRequest{ID: id}, nil
}

func decodeCreateRFQRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.CreateRFQRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

func decodeUpdateStageRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	var req spec.UpdateStageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = id
	return &req, nil
}

func decodeParseShipmentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.ParseShipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

func decodeAddQuoteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	var req spec.AddQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req.Quote); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = id
	return &req, nil
}

// decodeGetCarrierRatesRequest reads the RFQ ID from the URL path.
// No request body is needed — the carrier service reads origin/destination from the DB.
func decodeGetCarrierRatesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetCarrierRatesRequest{ID: id}, nil
}

// decodeApproveQuoteRequest reads the RFQ ID from the path and quote_id from the JSON body.
func decodeApproveQuoteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	var req spec.ApproveQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = id
	return &req, nil
}

// decodeGetRequirementsRequest reads the RFQ ID from the URL path.
// No request body — requirements are evaluated purely from the RFQ data in DB.
func decodeGetRequirementsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetRequirementsRequest{ID: id}, nil
}

// decodeGetActivityRequest reads the RFQ ID from the URL path.
// No request body — activity is aggregated from RFQ, Lead, interactions, requirements, and quotes.
func decodeGetActivityRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetActivityRequest{ID: id}, nil
}

func getDocumentIDFromVars(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "documentId")
	if idStr == "" {
		vars := mux.Vars(r)
		idStr = vars["documentId"]
	}
	if idStr == "" {
		return 0, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return id, nil
}

func decodeGetDocumentsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetDocumentsRequest{ID: id}, nil
}

func decodeCreateDocumentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}

	var req spec.CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.RFQID = id
	return &req, nil
}

func decodeUpdateDocumentStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	rfqID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}

	docID, err := getDocumentIDFromVars(r)
	if err != nil {
		return nil, err
	}

	var req spec.UpdateDocumentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.RFQID = rfqID
	req.DocumentID = docID
	return &req, nil
}

func decodeDeleteDocumentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	rfqID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}

	docID, err := getDocumentIDFromVars(r)
	if err != nil {
		return nil, err
	}

	return &spec.DeleteDocumentRequest{
		RFQID:      rfqID,
		DocumentID: docID,
	}, nil
}

func getQuoteIDFromVars(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "quoteId")
	if idStr == "" {
		vars := mux.Vars(r)
		idStr = vars["quoteId"]
	}
	if idStr == "" {
		return 0, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	val, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return val, nil
}

func decodeGetQuotesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetQuotesRequest{ID: id}, nil
}

func decodeCreateRFQQuoteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	rfqID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}

	var req spec.CreateQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.RFQID = rfqID
	return &req, nil
}

func decodeUpdateQuoteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	rfqID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	quoteID, err := getQuoteIDFromVars(r)
	if err != nil {
		return nil, err
	}

	var req spec.UpdateQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.RFQID = rfqID
	req.QuoteID = quoteID
	return &req, nil
}

func decodeUpdateQuoteStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	rfqID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	quoteID, err := getQuoteIDFromVars(r)
	if err != nil {
		return nil, err
	}

	var req spec.UpdateQuoteStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.RFQID = rfqID
	req.QuoteID = quoteID
	return &req, nil
}

func decodeRecommendQuoteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	rfqID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	quoteID, err := getQuoteIDFromVars(r)
	if err != nil {
		return nil, err
	}

	return &spec.RecommendQuoteRequest{
		RFQID:   rfqID,
		QuoteID: quoteID,
	}, nil
}

func decodeApproveRFQQuoteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	rfqID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	quoteID, err := getQuoteIDFromVars(r)
	if err != nil {
		return nil, err
	}

	var req spec.ApproveRFQQuoteRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	req.RFQID = rfqID
	req.QuoteID = quoteID
	return &req, nil
}

func decodeSelectQuoteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	rfqID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	quoteID, err := getQuoteIDFromVars(r)
	if err != nil {
		return nil, err
	}

	return &spec.SelectQuoteRequest{
		RFQID:   rfqID,
		QuoteID: quoteID,
	}, nil
}

func decodeDeleteQuoteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	rfqID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	quoteID, err := getQuoteIDFromVars(r)
	if err != nil {
		return nil, err
	}

	return &spec.DeleteQuoteRequest{
		RFQID:   rfqID,
		QuoteID: quoteID,
	}, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Task 14: Booking & Shipment Decoders
// ──────────────────────────────────────────────────────────────────────────────

func getBookingIDFromVars(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "bookingId")
	if idStr == "" {
		vars := mux.Vars(r)
		idStr = vars["bookingId"]
	}
	if idStr == "" {
		return 0, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return id, nil
}

func decodeGetBookingHandoffRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetBookingHandoffRequest{ID: id}, nil
}

func decodeCreateBookingRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}

	var req spec.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.RFQID = id
	return &req, nil
}

func decodeUpdateBookingStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	rfqID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	bookingID, err := getBookingIDFromVars(r)
	if err != nil {
		return nil, err
	}

	var req spec.UpdateBookingStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.RFQID = rfqID
	req.BookingID = bookingID
	return &req, nil
}

func decodeGetShipmentHandoffRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetShipmentHandoffRequest{ID: id}, nil
}

func decodeGetBookingsWorkspaceRequest(_ context.Context, r *http.Request) (interface{}, error) {
	q := r.URL.Query()
	req := &spec.BookingListFilter{
		Page:    1,
		Limit:   10,
		SortBy:  q.Get("sort_by"),
		SortDir: q.Get("sort_dir"),
	}

	if pageStr := q.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			req.Page = p
		}
	}
	if limitStr := q.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			req.Limit = l
		}
	}

	if status := q.Get("status"); status != "" && status != "ALL" {
		req.Status = &status
	}
	if carrier := q.Get("carrier"); carrier != "" {
		req.Carrier = &carrier
	}
	if origin := q.Get("origin_port"); origin != "" {
		req.OriginPort = &origin
	}
	if dest := q.Get("destination_port"); dest != "" {
		req.DestinationPort = &dest
	}
	if search := q.Get("search"); search != "" {
		req.Search = &search
	}
	if etdFromStr := q.Get("etd_from"); etdFromStr != "" {
		if t, err := time.Parse(time.RFC3339, etdFromStr); err == nil {
			req.ETDFrom = &t
		} else if t2, err2 := time.Parse("2006-01-02", etdFromStr); err2 == nil {
			req.ETDFrom = &t2
		}
	}
	if etdToStr := q.Get("etd_to"); etdToStr != "" {
		if t, err := time.Parse(time.RFC3339, etdToStr); err == nil {
			req.ETDTo = &t
		} else if t2, err2 := time.Parse("2006-01-02", etdToStr); err2 == nil {
			req.ETDTo = &t2
		}
	}

	return req, nil
}

func decodeGetEligibleRFQsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return &spec.GetEligibleRFQsForBookingRequest{}, nil
}

func decodeGetBookingWorkspaceDetailRequest(_ context.Context, r *http.Request) (interface{}, error) {
	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil || bookingID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &spec.GetBookingWorkspaceDetailRequest{BookingID: bookingID}, nil
}

func decodeDirectUpdateBookingStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil || bookingID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.DirectUpdateBookingStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}
	req.BookingID = bookingID
	return &req, nil
}

func decodeCreateShipmentFromBookingRequest(_ context.Context, r *http.Request) (interface{}, error) {
	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil || bookingID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.CreateShipmentFromBookingRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	req.BookingID = bookingID
	return &req, nil
}

func decodeBookWithCarrierRequest(_ context.Context, r *http.Request) (interface{}, error) {
	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil || bookingID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.BookWithCarrierRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}
	}
	req.BookingID = bookingID
	return &req, nil
}

func decodeSyncCarrierBookingRequest(_ context.Context, r *http.Request) (interface{}, error) {
	bookingIDStr := chi.URLParam(r, "bookingId")
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil || bookingID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return &spec.SyncCarrierBookingRequest{
		BookingID: bookingID,
	}, nil
}






func encodeAPIResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func encodeErrorResponse(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if svcErr, ok := err.(*svcerror.ServiceError); ok {
		switch svcErr.Code {
		case svcerror.ErrInvalidArgument:
			w.WriteHeader(http.StatusBadRequest)
		case svcerror.ErrInsufficientResourceAccess:
			w.WriteHeader(http.StatusUnauthorized)
		case svcerror.ErrResourceNotFound:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    svcErr.Code,
				"message": svcErr.Message,
			},
		})
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    svcerror.ErrInternal,
			"message": err.Error(),
		},
	})
}
