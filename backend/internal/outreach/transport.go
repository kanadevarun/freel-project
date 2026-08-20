package outreach

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/freel/backend/internal/outreach/spec"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-chi/chi/v5"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

func AddOutreachHandlers(
	router chi.Router,
	endpoints Endpoints,
	authMiddleware func(http.Handler) http.Handler,
) {
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	router.With(authMiddleware).Get("/campaigns", kitHttp.NewServer(
		endpoints.ListCampaignsEP,
		decodeListCampaignsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/campaigns", kitHttp.NewServer(
		endpoints.CreateCampaignEP,
		decodeCreateCampaignRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/campaigns/{id:[0-9]+}", kitHttp.NewServer(
		endpoints.GetCampaignEP,
		decodeGetCampaignRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/campaigns/{id:[0-9]+}/activate", kitHttp.NewServer(
		endpoints.ActivateCampaignEP,
		decodeActivateCampaignRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/campaigns/{id:[0-9]+}/pause", kitHttp.NewServer(
		endpoints.PauseCampaignEP,
		decodePauseCampaignRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Delete("/campaigns/{id:[0-9]+}", kitHttp.NewServer(
		endpoints.DeleteCampaignEP,
		decodeDeleteCampaignRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/generate-email", kitHttp.NewServer(
		endpoints.GenerateEmailEP,
		decodeGenerateEmailRequest,
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
