package outreach

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/freel/backend/internal/outreach/spec"
	"github.com/freel/backend/internal/svcerror"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

func AddOutreachHandlers(
	router *mux.Router,
	endpoints Endpoints,
	authMiddleware func(http.Handler) http.Handler,
) {
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	router.Methods(http.MethodGet).Path(spec.ListCampaignsURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.ListCampaignsEP,
		decodeListCampaignsRequest,
		encodeAPIResponse,
		options...,
	)))

	router.Methods(http.MethodPost).Path(spec.CreateCampaignURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.CreateCampaignEP,
		decodeCreateCampaignRequest,
		encodeAPIResponse,
		options...,
	)))

	router.Methods(http.MethodGet).Path(spec.GetCampaignURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.GetCampaignEP,
		decodeGetCampaignRequest,
		encodeAPIResponse,
		options...,
	)))

	router.Methods(http.MethodPost).Path(spec.ActivateCampaignURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.ActivateCampaignEP,
		decodeActivateCampaignRequest,
		encodeAPIResponse,
		options...,
	)))

	router.Methods(http.MethodPost).Path(spec.PauseCampaignURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.PauseCampaignEP,
		decodePauseCampaignRequest,
		encodeAPIResponse,
		options...,
	)))

	router.Methods(http.MethodDelete).Path(spec.DeleteCampaignURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.DeleteCampaignEP,
		decodeDeleteCampaignRequest,
		encodeAPIResponse,
		options...,
	)))

	router.Methods(http.MethodPost).Path(spec.GenerateEmailURL).Handler(authMiddleware(kitHttp.NewServer(
		endpoints.GenerateEmailEP,
		decodeGenerateEmailRequest,
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

func decodeListCampaignsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	return &spec.ListCampaignsRequest{
		Limit:  limit,
		Offset: offset,
	}, nil
}

func decodeCreateCampaignRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

func decodeGetCampaignRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetCampaignRequest{ID: id}, nil
}

func decodeActivateCampaignRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.ActivateCampaignRequest{ID: id}, nil
}

func decodePauseCampaignRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.PauseCampaignRequest{ID: id}, nil
}

func decodeDeleteCampaignRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.DeleteCampaignRequest{ID: id}, nil
}

func decodeGenerateEmailRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.GenerateEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
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
