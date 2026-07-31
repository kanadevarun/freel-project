package rfq

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/freel/backend/internal/rfq/spec"
	"github.com/freel/backend/internal/svcerror"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// AddRFQHandlers adds the handlers to the rest methods for the rfq module
func AddRFQHandlers(
	router *mux.Router,
	endpoints Endpoints,
	authMiddleware func(http.Handler) http.Handler,
) {
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	// List RFQs
	router.Methods(http.MethodGet).Path(spec.ListURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.ListRFQsEP,
		decodeListRFQsRequest,
		encodeAPIResponse,
		options...,
	)))

	// Get RFQ
	router.Methods(http.MethodGet).Path(spec.GetURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.GetRFQEP,
		decodeGetRFQRequest,
		encodeAPIResponse,
		options...,
	)))

	// Get Timeline
	router.Methods(http.MethodGet).Path(spec.GetTimelineURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.GetTimelineEP,
		decodeGetTimelineRequest,
		encodeAPIResponse,
		options...,
	)))

	// Get Agent Status
	router.Methods(http.MethodGet).Path(spec.GetAgentStatusURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.GetAgentStatusEP,
		decodeGetAgentStatusRequest,
		encodeAPIResponse,
		options...,
	)))

	// Create RFQ
	router.Methods(http.MethodPost).Path(spec.CreateURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.CreateRFQEP,
		decodeCreateRFQRequest,
		encodeAPIResponse,
		options...,
	)))

	// Update Stage
	router.Methods(http.MethodPut).Path(spec.UpdateStageURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.UpdateStageEP,
		decodeUpdateStageRequest,
		encodeAPIResponse,
		options...,
	)))

	// Parse Shipment Request
	router.Methods(http.MethodPost).Path(spec.ParseShipmentRequestURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.ParseShipmentRequestEP,
		decodeParseShipmentRequest,
		encodeAPIResponse,
		options...,
	)))

	// Add Quote
	router.Methods(http.MethodPost).Path(spec.AddQuoteURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.AddQuoteEP,
		decodeAddQuoteRequest,
		encodeAPIResponse,
		options...,
	)))
}

func getIDFromVars(r *http.Request) (int32, error) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
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
