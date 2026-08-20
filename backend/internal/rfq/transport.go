package rfq

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

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
