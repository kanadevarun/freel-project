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

	// Sequence routes
	router.With(authMiddleware).Get("/campaigns/{id:[0-9]+}/sequence", kitHttp.NewServer(
		endpoints.GetCampaignSequenceEP,
		decodeGetCampaignRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/campaigns/{id:[0-9]+}/sequence", kitHttp.NewServer(
		endpoints.AddSequenceStepEP,
		decodeAddSequenceStepRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Put("/campaigns/{id:[0-9]+}/sequence/{step_id:[0-9]+}", kitHttp.NewServer(
		endpoints.UpdateSequenceStepEP,
		decodeUpdateSequenceStepRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Delete("/campaigns/{id:[0-9]+}/sequence/{step_id:[0-9]+}", kitHttp.NewServer(
		endpoints.DeleteSequenceStepEP,
		decodeDeleteSequenceStepRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Put("/campaigns/{id:[0-9]+}/sequence/reorder", kitHttp.NewServer(
		endpoints.ReorderSequenceEP,
		decodeReorderSequenceRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Audience routes
	router.With(authMiddleware).Get("/campaigns/{id:[0-9]+}/audience", kitHttp.NewServer(
		endpoints.GetCampaignAudienceEP,
		decodeGetCampaignAudienceRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/campaigns/{id:[0-9]+}/audience", kitHttp.NewServer(
		endpoints.AddCampaignAudienceEP,
		decodeAddCampaignAudienceRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Delete("/campaigns/{id:[0-9]+}/audience/{lead_id:[0-9]+}", kitHttp.NewServer(
		endpoints.RemoveCampaignAudienceEP,
		decodeRemoveCampaignAudienceRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Analytics & Insights routes
	router.With(authMiddleware).Get("/analytics", kitHttp.NewServer(
		endpoints.GetOutreachAnalyticsEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/campaigns/{id:[0-9]+}/analytics", kitHttp.NewServer(
		endpoints.GetCampaignAnalyticsEP,
		decodeGetCampaignRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/campaigns/{id:[0-9]+}/leads", kitHttp.NewServer(
		endpoints.GetCampaignLeadsEP,
		decodeGetCampaignRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/campaigns/{id:[0-9]+}/insights", kitHttp.NewServer(
		endpoints.GetCampaignInsightsEP,
		decodeGetCampaignRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/conversion-funnel", kitHttp.NewServer(
		endpoints.GetConversionFunnelEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Activity CRUD routes
	router.Route("/activities", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/", kitHttp.NewServer(
			endpoints.CreateActivityEP,
			decodeCreateActivityRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Get("/{id:[0-9]+}", kitHttp.NewServer(
			endpoints.GetActivityEP,
			decodeGetActivityRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Put("/{id:[0-9]+}", kitHttp.NewServer(
			endpoints.UpdateActivityEP,
			decodeUpdateActivityRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Put("/{id:[0-9]+}/complete", kitHttp.NewServer(
			endpoints.CompleteActivityEP,
			decodeGetActivityRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Delete("/{id:[0-9]+}", kitHttp.NewServer(
			endpoints.DeleteActivityEP,
			decodeGetActivityRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)
	})

	// Campaign Recipients and Activities Timeline routes
	router.With(authMiddleware).Get("/campaigns/{id:[0-9]+}/recipients", kitHttp.NewServer(
		endpoints.GetCampaignRecipientsEP,
		decodeGetCampaignRecipientsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/campaigns/{id:[0-9]+}/activity", kitHttp.NewServer(
		endpoints.GetCampaignActivityEP,
		decodeGetCampaignRecipientsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Prospects parent routes mapped directly to support slash-less requests
	router.With(authMiddleware).Get("/prospects", kitHttp.NewServer(
		endpoints.GetProspectsEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)
	router.With(authMiddleware).Post("/prospects", kitHttp.NewServer(
		endpoints.EnrollProspectEP,
		decodeEnrollProspectRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Prospects & Engagement routes
	router.Route("/prospects", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/", kitHttp.NewServer(
			endpoints.GetProspectsEP,
			decodeEmptyRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Get("/{id:[0-9]+}/engagement", kitHttp.NewServer(
			endpoints.GetProspectEngagementEP,
			decodeGetProspectEngagementRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Get("/{id:[0-9]+}/activity", kitHttp.NewServer(
			endpoints.GetLeadOutreachActivityEP,
			decodeGetProspectEngagementRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Post("/", kitHttp.NewServer(
			endpoints.EnrollProspectEP,
			decodeEnrollProspectRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Get("/{id:[0-9]+}", kitHttp.NewServer(
			endpoints.GetProspectDetailEP,
			decodeGetProspectEngagementRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Put("/{id:[0-9]+}", kitHttp.NewServer(
			endpoints.UpdateProspectEP,
			decodeUpdateProspectRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Post("/{id:[0-9]+}/pause", kitHttp.NewServer(
			endpoints.PauseProspectEP,
			decodeProspectControlRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Post("/{id:[0-9]+}/resume", kitHttp.NewServer(
			endpoints.ResumeProspectEP,
			decodeProspectControlRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Post("/{id:[0-9]+}/stop", kitHttp.NewServer(
			endpoints.StopProspectEP,
			decodeProspectControlRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)
	})

	// Follow-ups parent routes mapped directly to support slash-less requests
	router.With(authMiddleware).Get("/follow-ups", kitHttp.NewServer(
		endpoints.GetFollowUpsEP,
		decodeGetFollowUpsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)
	router.With(authMiddleware).Post("/follow-ups", kitHttp.NewServer(
		endpoints.CreateActivityEP,
		decodeCreateActivityRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Follow-ups Queue routes
	router.Route("/follow-ups", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/", kitHttp.NewServer(
			endpoints.GetFollowUpsEP,
			decodeGetFollowUpsRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Post("/", kitHttp.NewServer(
			endpoints.CreateActivityEP,
			decodeCreateActivityRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Get("/{id:[0-9]+}", kitHttp.NewServer(
			endpoints.GetActivityEP,
			decodeGetActivityRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Put("/{id:[0-9]+}", kitHttp.NewServer(
			endpoints.UpdateActivityEP,
			decodeUpdateActivityRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Post("/{id:[0-9]+}/complete", kitHttp.NewServer(
			endpoints.CompleteActivityEP,
			decodeGetActivityRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Post("/{id:[0-9]+}/cancel", kitHttp.NewServer(
			endpoints.CancelFollowUpEP,
			decodeGetFollowUpByIDRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)

		r.Post("/{id:[0-9]+}/reschedule", kitHttp.NewServer(
			endpoints.RescheduleFollowUpEP,
			decodeRescheduleFollowUpRequest,
			encodeAPIResponse,
			options...,
		).ServeHTTP)
	})
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

func decodeAddSequenceStepRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	var req spec.AddSequenceStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.CampaignID = id
	return &req, nil
}

func decodeUpdateSequenceStepRequest(_ context.Context, r *http.Request) (interface{}, error) {
	campaignID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	stepID, err := getParamFromVars(r, "step_id")
	if err != nil {
		return nil, err
	}
	var req spec.UpdateSequenceStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.CampaignID = campaignID
	req.StepID = stepID
	return &req, nil
}

func decodeDeleteSequenceStepRequest(_ context.Context, r *http.Request) (interface{}, error) {
	campaignID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	stepID, err := getParamFromVars(r, "step_id")
	if err != nil {
		return nil, err
	}
	return &spec.DeleteSequenceStepRequest{
		CampaignID: campaignID,
		StepID:     stepID,
	}, nil
}

func decodeReorderSequenceRequest(_ context.Context, r *http.Request) (interface{}, error) {
	campaignID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	var req spec.ReorderSequenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.CampaignID = campaignID
	return &req, nil
}

func decodeGetCampaignAudienceRequest(_ context.Context, r *http.Request) (interface{}, error) {
	campaignID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetCampaignAudienceRequest{CampaignID: campaignID}, nil
}

func decodeAddCampaignAudienceRequest(_ context.Context, r *http.Request) (interface{}, error) {
	campaignID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	var req spec.AddCampaignAudienceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.CampaignID = campaignID
	return &req, nil
}

func decodeRemoveCampaignAudienceRequest(_ context.Context, r *http.Request) (interface{}, error) {
	campaignID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	leadID, err := getParamFromVars(r, "lead_id")
	if err != nil {
		return nil, err
	}
	return &spec.RemoveCampaignAudienceRequest{
		CampaignID: campaignID,
		LeadID:     leadID,
	}, nil
}

func decodeEmptyRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func getParamFromVars(r *http.Request, name string) (int32, error) {
	valStr := chi.URLParam(r, name)
	if valStr == "" {
		vars := mux.Vars(r)
		valStr = vars[name]
	}
	if valStr == "" {
		return 0, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	val, err := strconv.ParseInt(valStr, 10, 32)
	if err != nil {
		return 0, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return int32(val), nil
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

func decodeCreateActivityRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.CreateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

func decodeGetActivityRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &spec.GetActivityRequest{ID: id}, nil
}

func decodeUpdateActivityRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var req spec.UpdateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = id
	return &req, nil
}

func decodeGetProspectEngagementRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return id, nil
}

func decodeGetCampaignRecipientsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return int32(id), nil
}

func decodeEnrollProspectRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.EnrollProspectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

func decodeUpdateProspectRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var req spec.UpdateProspectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.LeadID = id
	return &req, nil
}

func decodeProspectControlRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var req spec.UpdateProspectRequest
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	req.LeadID = id
	return &req, nil
}

func decodeGetFollowUpsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	filter := r.URL.Query().Get("filter")
	return filter, nil
}

func decodeRescheduleFollowUpRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var req spec.RescheduleFollowUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = id
	return &req, nil
}

func decodeGetFollowUpByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return id, nil
}
