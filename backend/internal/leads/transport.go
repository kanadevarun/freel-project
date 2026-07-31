package leads

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/freel/backend/internal/leads/spec"
	"github.com/freel/backend/internal/svcerror"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

func AddLeadsHandlers(
	router *mux.Router,
	endpoints Endpoints,
	authMiddleware func(http.Handler) http.Handler,
) {
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	router.Methods(http.MethodGet).Path(spec.ListURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.ListLeadsEP,
		decodeListLeadsRequest,
		encodeAPIResponse,
		options...,
	)))

	router.Methods(http.MethodPost).Path(spec.CreateURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.CreateLeadEP,
		decodeCreateLeadRequest,
		encodeAPIResponse,
		options...,
	)))

	router.Methods(http.MethodPost).Path(spec.ImportURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.ImportLeadsEP,
		decodeImportLeadsRequest,
		encodeAPIResponse,
		options...,
	)))

	router.Methods(http.MethodGet).Path(spec.GetURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.GetLeadEP,
		decodeGetLeadRequest,
		encodeAPIResponse,
		options...,
	)))

	router.Methods(http.MethodPut).Path(spec.UpdateURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.UpdateLeadEP,
		decodeUpdateLeadRequest,
		encodeAPIResponse,
		options...,
	)))

	router.Methods(http.MethodDelete).Path(spec.DeleteURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.DeleteLeadEP,
		decodeDeleteLeadRequest,
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

func decodeListLeadsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	statusStr := r.URL.Query().Get("status")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	var status *string
	if statusStr != "" {
		status = &statusStr
	}

	return &spec.ListLeadsRequest{
		Limit:  limit,
		Offset: offset,
		Status: status,
	}, nil
}

func decodeCreateLeadRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.CreateLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

func decodeImportLeadsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req []*spec.CreateLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &spec.ImportLeadsRequest{Leads: req}, nil
}

func decodeGetLeadRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetLeadRequest{ID: id}, nil
}

func decodeUpdateLeadRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	var req spec.UpdateLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = id
	return &req, nil
}

func decodeDeleteLeadRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.DeleteLeadRequest{ID: id}, nil
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
