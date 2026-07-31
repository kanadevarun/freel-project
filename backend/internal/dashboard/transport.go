package dashboard

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/freel/backend/internal/dashboard/spec"
	"github.com/freel/backend/internal/svcerror"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// AddDashboardHandlers adds the handlers to the rest methods for the dashboard module
func AddDashboardHandlers(
	router *mux.Router,
	endpoints Endpoints,
	authMiddleware func(http.Handler) http.Handler,
) {
	// Options for go-kit HTTP server
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	// Get Mission Control
	// We wrap the go-kit handler with our existing auth middleware
	missionControlHandler := authMiddleware(kitHttp.NewServer(
		endpoints.GetMissionControlEP,
		decodeGetMissionControlRequest,
		encodeAPIResponse,
		options...,
	))

	router.Methods(http.MethodGet).Path(spec.GetMissionControlURL).Handler(missionControlHandler)
}

func decodeGetMissionControlRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// The request doesn't have a body or query params for now.
	// The OrgID will be extracted from the context in the endpoint.
	return &spec.GetMissionControlRequest{}, nil
}

func encodeAPIResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func encodeErrorResponse(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if svcErr, ok := err.(*svcerror.ServiceError); ok {
		// Map some known errors to HTTP status codes
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
