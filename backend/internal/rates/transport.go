package rates

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/freel/backend/internal/rates/spec"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-chi/chi/v5"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// AddRatesHandlers mounts all Rate Management & Intelligence endpoints onto the chi router
func AddRatesHandlers(
	router chi.Router,
	endpoints Endpoints,
	authMiddleware func(http.Handler) http.Handler,
) {
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	// GET /summary
	router.With(authMiddleware).Get("/summary", kitHttp.NewServer(
		endpoints.GetRateSummaryEP,
		decodeGetRateSummaryRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /
	router.With(authMiddleware).Get("/", kitHttp.NewServer(
		endpoints.ListRatesEP,
		decodeListRatesRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /
	router.With(authMiddleware).Post("/", kitHttp.NewServer(
		endpoints.CreateRateEP,
		decodeCreateRateRequest,
		encodeAPICreatedResponse,
		options...,
	).ServeHTTP)

	// GET /search
	router.With(authMiddleware).Get("/search", kitHttp.NewServer(
		endpoints.SearchRatesEP,
		decodeSearchRatesRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /carrier-search (Task 5: Live Connected Carrier Rates)
	router.With(authMiddleware).Post("/carrier-search", kitHttp.NewServer(
		endpoints.SearchCarrierLiveRatesEP,
		decodeSearchCarrierLiveRatesRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /carrier-rates (alias)
	router.With(authMiddleware).Post("/carrier-rates", kitHttp.NewServer(
		endpoints.SearchCarrierLiveRatesEP,
		decodeSearchCarrierLiveRatesRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /spot/refresh
	router.With(authMiddleware).Post("/spot/refresh", kitHttp.NewServer(
		endpoints.RefreshSpotRatesEP,
		decodeRefreshSpotRatesRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /{id}
	router.With(authMiddleware).Get("/{id}", kitHttp.NewServer(
		endpoints.GetRateEP,
		decodeGetRateRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// PUT /{id}
	router.With(authMiddleware).Put("/{id}", kitHttp.NewServer(
		endpoints.UpdateRateEP,
		decodeUpdateRateRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /{id}/archive
	router.With(authMiddleware).Post("/{id}/archive", kitHttp.NewServer(
		endpoints.ArchiveRateEP,
		decodeArchiveRateRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// ── Task 19.2: Rate Charges & Commercial Pricing Routes ───────────────

	// GET /{id}/pricing
	router.With(authMiddleware).Get("/{id}/pricing", kitHttp.NewServer(
		endpoints.GetRatePricingEP,
		decodeGetRateRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /{id}/charges
	router.With(authMiddleware).Post("/{id}/charges", kitHttp.NewServer(
		endpoints.AddRateChargeEP,
		decodeCreateRateChargeRequest,
		encodeAPICreatedResponse,
		options...,
	).ServeHTTP)

	// PUT /{id}/charges/{chargeId}
	router.With(authMiddleware).Put("/{id}/charges/{chargeId}", kitHttp.NewServer(
		endpoints.UpdateRateChargeEP,
		decodeUpdateRateChargeRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// DELETE /{id}/charges/{chargeId}
	router.With(authMiddleware).Delete("/{id}/charges/{chargeId}", kitHttp.NewServer(
		endpoints.DeleteRateChargeEP,
		decodeDeleteRateChargeRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /{id}/charges/reorder
	router.With(authMiddleware).Post("/{id}/charges/reorder", kitHttp.NewServer(
		endpoints.ReorderRateChargesEP,
		decodeReorderRateChargesRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// ── Task 19.3: Carrier Rate Contracts & Versions Routes ─────────────────

	// GET /contracts
	router.With(authMiddleware).Get("/contracts", kitHttp.NewServer(
		endpoints.ListRateContractsEP,
		decodeListRateContractsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /contracts/summary
	router.With(authMiddleware).Get("/contracts/summary", kitHttp.NewServer(
		endpoints.GetRateContractSummaryEP,
		decodeGetRateSummaryRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /contracts
	router.With(authMiddleware).Post("/contracts", kitHttp.NewServer(
		endpoints.CreateRateContractEP,
		decodeCreateRateContractRequest,
		encodeAPICreatedResponse,
		options...,
	).ServeHTTP)

	// GET /contracts/{id}
	router.With(authMiddleware).Get("/contracts/{id}", kitHttp.NewServer(
		endpoints.GetRateContractEP,
		decodeContractIDParam,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// PUT /contracts/{id}
	router.With(authMiddleware).Put("/contracts/{id}", kitHttp.NewServer(
		endpoints.UpdateRateContractEP,
		decodeUpdateRateContractRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /contracts/{id}/archive
	router.With(authMiddleware).Post("/contracts/{id}/archive", kitHttp.NewServer(
		endpoints.ArchiveRateContractEP,
		decodeContractIDParam,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /contracts/{id}/renew
	router.With(authMiddleware).Post("/contracts/{id}/renew", kitHttp.NewServer(
		endpoints.RenewRateContractEP,
		decodeRenewRateContractRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /{id}/versions
	router.With(authMiddleware).Post("/{id}/versions", kitHttp.NewServer(
		endpoints.CreateRateVersionEP,
		decodeCreateRateVersionRequest,
		encodeAPICreatedResponse,
		options...,
	).ServeHTTP)

	// GET /{id}/versions
	router.With(authMiddleware).Get("/{id}/versions", kitHttp.NewServer(
		endpoints.GetRateVersionsEP,
		decodeRateIDParam,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /{id}/version-history
	router.With(authMiddleware).Get("/{id}/version-history", kitHttp.NewServer(
		endpoints.GetRateVersionHistoryEP,
		decodeRateIDParam,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// ── Task 19.4: Spot Rate Requests, Responses & Comparison Routes ────────

	// GET /spot-requests
	router.With(authMiddleware).Get("/spot-requests", kitHttp.NewServer(
		endpoints.ListSpotRateRequestsEP,
		decodeListSpotRateRequestsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /spot-requests/summary
	router.With(authMiddleware).Get("/spot-requests/summary", kitHttp.NewServer(
		endpoints.GetSpotRateSummaryEP,
		decodeGetRateSummaryRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /spot-requests
	router.With(authMiddleware).Post("/spot-requests", kitHttp.NewServer(
		endpoints.CreateSpotRateRequestEP,
		decodeCreateSpotRateRequestRequest,
		encodeAPICreatedResponse,
		options...,
	).ServeHTTP)

	// GET /spot-requests/{id}
	router.With(authMiddleware).Get("/spot-requests/{id}", kitHttp.NewServer(
		endpoints.GetSpotRateRequestEP,
		decodeSpotRequestIDParam,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// PUT /spot-requests/{id}
	router.With(authMiddleware).Put("/spot-requests/{id}", kitHttp.NewServer(
		endpoints.UpdateSpotRateRequestEP,
		decodeUpdateSpotRateRequestRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /spot-requests/{id}/send
	router.With(authMiddleware).Post("/spot-requests/{id}/send", kitHttp.NewServer(
		endpoints.SendSpotRateRequestEP,
		decodeSpotRequestIDParam,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /spot-requests/{id}/cancel
	router.With(authMiddleware).Post("/spot-requests/{id}/cancel", kitHttp.NewServer(
		endpoints.CancelSpotRateRequestEP,
		decodeSpotRequestIDParam,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /spot-requests/{id}/responses
	router.With(authMiddleware).Get("/spot-requests/{id}/responses", kitHttp.NewServer(
		endpoints.GetSpotRateResponsesEP,
		decodeSpotRequestIDParam,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /spot-requests/{id}/responses
	router.With(authMiddleware).Post("/spot-requests/{id}/responses", kitHttp.NewServer(
		endpoints.CreateSpotRateResponseEP,
		decodeCreateSpotRateResponseRequest,
		encodeAPICreatedResponse,
		options...,
	).ServeHTTP)

	// GET /spot-requests/{id}/responses/{responseId}
	router.With(authMiddleware).Get("/spot-requests/{id}/responses/{responseId}", kitHttp.NewServer(
		endpoints.GetSpotRateResponseEP,
		decodeSpotResponseIDParam,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// PUT /spot-requests/{id}/responses/{responseId}
	router.With(authMiddleware).Put("/spot-requests/{id}/responses/{responseId}", kitHttp.NewServer(
		endpoints.UpdateSpotRateResponseEP,
		decodeUpdateSpotRateResponseRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /spot-requests/{id}/responses/{responseId}/select
	router.With(authMiddleware).Post("/spot-requests/{id}/responses/{responseId}/select", kitHttp.NewServer(
		endpoints.SelectPreferredSpotRateEP,
		decodeSelectPreferredSpotRateRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /spot-requests/{id}/comparison
	router.With(authMiddleware).Get("/spot-requests/{id}/comparison", kitHttp.NewServer(
		endpoints.CompareSpotRatesEP,
		decodeSpotRequestIDParam,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// ── Task 19.6: Rate Lifecycle Intelligence & Attention Routes ───────────

	// GET /lifecycle/summary
	router.With(authMiddleware).Get("/lifecycle/summary", kitHttp.NewServer(
		endpoints.GetRateLifecycleSummaryEP,
		decodeGetRateSummaryRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /lifecycle/events
	router.With(authMiddleware).Get("/lifecycle/events", kitHttp.NewServer(
		endpoints.GetRateLifecycleEventsEP,
		decodeGetRateLifecycleEventsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /lifecycle/attention
	router.With(authMiddleware).Get("/lifecycle/attention", kitHttp.NewServer(
		endpoints.GetRatesRequiringAttentionEP,
		decodeGetRateSummaryRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /contracts/attention
	router.With(authMiddleware).Get("/contracts/attention", kitHttp.NewServer(
		endpoints.GetContractsRequiringAttentionEP,
		decodeGetRateSummaryRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /lifecycle/evaluate
	router.With(authMiddleware).Post("/lifecycle/evaluate", kitHttp.NewServer(
		endpoints.EvaluateRateLifecycleEP,
		decodeGetRateSummaryRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// ── Task 19.7: Rate Analytics & Procurement Intelligence Routes ─────────────

	// GET /analytics/overview
	router.With(authMiddleware).Get("/analytics/overview", kitHttp.NewServer(
		endpoints.GetRateAnalyticsOverviewEP,
		decodeGetRateSummaryRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /analytics/trends?days=7|30|90
	router.With(authMiddleware).Get("/analytics/trends", kitHttp.NewServer(
		endpoints.GetRateAnalyticsTrendsEP,
		decodeAnalyticsTrendsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /analytics/carriers
	router.With(authMiddleware).Get("/analytics/carriers", kitHttp.NewServer(
		endpoints.GetCarrierRatePerformanceEP,
		decodeGetRateSummaryRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /analytics/lanes
	router.With(authMiddleware).Get("/analytics/lanes", kitHttp.NewServer(
		endpoints.GetLaneRatePerformanceEP,
		decodeGetRateSummaryRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /analytics/lifecycle
	router.With(authMiddleware).Get("/analytics/lifecycle", kitHttp.NewServer(
		endpoints.GetRateLifecycleAnalyticsEP,
		decodeGetRateSummaryRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /analytics/spot-performance
	router.With(authMiddleware).Get("/analytics/spot-performance", kitHttp.NewServer(
		endpoints.GetSpotSourcingPerformanceEP,
		decodeGetRateSummaryRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /analytics/insights
	router.With(authMiddleware).Get("/analytics/insights", kitHttp.NewServer(
		endpoints.GetRateCommercialInsightsEP,
		decodeGetRateSummaryRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)
}

// decodeAnalyticsTrendsRequest decodes the ?days= query parameter (7, 30, 90; default 30).
func decodeAnalyticsTrendsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			switch d {
			case 7, 30, 90:
				days = d
			}
		}
	}
	return map[string]int{"days": days}, nil
}

// AddPortInternalHandlers registers port normalization handlers for internal services
func AddPortInternalHandlers(router chi.Router) {
	router.Get("/ports/normalize", func(w http.ResponseWriter, r *http.Request) {
		port := r.URL.Query().Get("port")
		normalized := NormalizePort(port)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"input":      port,
			"normalized": normalized,
			"is_known":   strconv.FormatBool(IsKnownPort(port)),
		})
	})

	router.Get("/ports/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		results := SearchPorts(q)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"query":   q,
			"matches": results,
			"count":   len(results),
		})
	})
}

func getIDFromVars(r *http.Request) string {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		vars := mux.Vars(r)
		idStr = vars["id"]
	}
	if idStr == "" {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			idStr = parts[len(parts)-1]
		}
	}
	return idStr
}

func decodeListRatesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := &spec.ListRatesRequest{
		Search:        strings.TrimSpace(r.URL.Query().Get("search")),
		Status:        strings.TrimSpace(r.URL.Query().Get("status")),
		RateType:      strings.TrimSpace(r.URL.Query().Get("rate_type")),
		TransportMode: strings.TrimSpace(r.URL.Query().Get("transport_mode")),
		ServiceType:   strings.TrimSpace(r.URL.Query().Get("service_type")),
		EquipmentType: strings.TrimSpace(r.URL.Query().Get("equipment_type")),
		CarrierName:   strings.TrimSpace(r.URL.Query().Get("carrier_name")),
		Origin:        strings.TrimSpace(r.URL.Query().Get("origin")),
		Destination:   strings.TrimSpace(r.URL.Query().Get("destination")),
		SortBy:        strings.TrimSpace(r.URL.Query().Get("sort_by")),
		SortOrder:     strings.TrimSpace(r.URL.Query().Get("sort_order")),
	}

	if pStr := r.URL.Query().Get("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			req.Page = p
		}
	}
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			req.Limit = l
		}
	}
	if vDateStr := r.URL.Query().Get("valid_date"); vDateStr != "" {
		if vDate, err := time.Parse("2006-01-02", vDateStr); err == nil {
			req.ValidDate = &vDate
		}
	}

	return req, nil
}

func decodeGetRateSummaryRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return struct{}{}, nil
}

func decodeCreateRateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.CreateRateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

func decodeGetRateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id := getIDFromVars(r)
	if id == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &spec.GetRateRequest{ID: id}, nil
}

func decodeUpdateRateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := getIDFromVars(r)
	rateID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.UpdateRateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = rateID
	return &req, nil
}

func decodeArchiveRateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	idStr := ""
	if len(parts) >= 2 {
		idStr = parts[len(parts)-2]
	} else {
		idStr = getIDFromVars(r)
	}
	rateID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &spec.ArchiveRateRequest{ID: rateID}, nil
}

func decodeSearchRatesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	q := &spec.RateQuery{
		OriginPort:      strings.TrimSpace(r.URL.Query().Get("origin")),
		DestinationPort: strings.TrimSpace(r.URL.Query().Get("destination")),
		EquipmentType:   strings.TrimSpace(r.URL.Query().Get("equipment")),
		Incoterms:       strings.TrimSpace(r.URL.Query().Get("incoterms")),
	}

	if q.EquipmentType == "" {
		q.EquipmentType = strings.TrimSpace(r.URL.Query().Get("equipment_type"))
	}
	if dateStr := r.URL.Query().Get("target_date"); dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			q.TargetDate = &t
		}
	}
	if maxStr := r.URL.Query().Get("max_results"); maxStr != "" {
		if m, err := strconv.Atoi(maxStr); err == nil && m > 0 {
			q.MaxResults = m
		}
	}

	return q, nil
}

func decodeRefreshSpotRatesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.RefreshSpotRatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

// ── Task 19.2: Decoders for Rate Charges ─────────────────────────────────────

func decodeCreateRateChargeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	rateIDStr := chi.URLParam(r, "id")
	if rateIDStr == "" {
		rateIDStr = getIDFromVars(r)
	}
	rateID, err := strconv.ParseInt(rateIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.CreateRateChargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.RateID = rateID
	return &req, nil
}

func decodeUpdateRateChargeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	rateIDStr := chi.URLParam(r, "id")
	if rateIDStr == "" {
		rateIDStr = getIDFromVars(r)
	}
	rateID, err := strconv.ParseInt(rateIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	chargeIDStr := chi.URLParam(r, "chargeId")
	if chargeIDStr == "" {
		vars := mux.Vars(r)
		chargeIDStr = vars["chargeId"]
	}
	chargeID, err := strconv.ParseInt(chargeIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.UpdateRateChargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.RateID = rateID
	req.ChargeID = chargeID
	return &req, nil
}

func decodeDeleteRateChargeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	rateIDStr := chi.URLParam(r, "id")
	if rateIDStr == "" {
		rateIDStr = getIDFromVars(r)
	}
	rateID, err := strconv.ParseInt(rateIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	chargeIDStr := chi.URLParam(r, "chargeId")
	if chargeIDStr == "" {
		vars := mux.Vars(r)
		chargeIDStr = vars["chargeId"]
	}
	chargeID, err := strconv.ParseInt(chargeIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return &spec.DeleteRateChargeRequest{
		RateID:   rateID,
		ChargeID: chargeID,
	}, nil
}

func decodeReorderRateChargesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	rateIDStr := chi.URLParam(r, "id")
	if rateIDStr == "" {
		rateIDStr = getIDFromVars(r)
	}
	rateID, err := strconv.ParseInt(rateIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.ReorderRateChargesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.RateID = rateID
	return &req, nil
}

// ── Task 19.3 Decoders ────────────────────────────────────────────────────────

func decodeListRateContractsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	req := &spec.ListRateContractsRequest{
		Search:        q.Get("search"),
		CarrierName:   q.Get("carrier_name"),
		ContractType:  q.Get("contract_type"),
		Status:        q.Get("status"),
		RenewalStatus: q.Get("renewal_status"),
		TransportMode: q.Get("transport_mode"),
		Page:          page,
		Limit:         limit,
		SortBy:        q.Get("sort_by"),
		SortOrder:     q.Get("sort_order"),
	}
	return req, nil
}

func decodeCreateRateContractRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.CreateRateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

func decodeContractIDParam(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		idStr = getIDFromVars(r)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return id, nil
}

func decodeRateIDParam(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		idStr = getIDFromVars(r)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return id, nil
}

func decodeUpdateRateContractRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		idStr = getIDFromVars(r)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.UpdateRateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = id
	return &req, nil
}

func decodeRenewRateContractRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		idStr = getIDFromVars(r)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.RenewRateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = id
	return &req, nil
}

func decodeCreateRateVersionRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		idStr = getIDFromVars(r)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.CreateRateVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.RateID = id
	return &req, nil
}

func encodeAPIResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
		"message": "Success",
	})
}

func encodeAPICreatedResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
		"message": "Created",
	})
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

	if strings.Contains(err.Error(), "no rows") {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "NOT_FOUND",
				"message": "Rate not found",
			},
		})
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "INVALID_ARGUMENT",
			"message": err.Error(),
		},
	})
}

// ── Task 19.4: Spot Rate Decoders ────────────────────────────────────────────

func decodeListSpotRateRequestsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	return &spec.ListSpotRateRequestsRequest{
		Search:        q.Get("search"),
		Status:        q.Get("status"),
		TransportMode: q.Get("transport_mode"),
		Origin:        q.Get("origin"),
		Destination:   q.Get("destination"),
		Page:          page,
		Limit:         limit,
		SortBy:        q.Get("sort_by"),
		SortOrder:     q.Get("sort_order"),
	}, nil
}

func decodeCreateSpotRateRequestRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.CreateSpotRateRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

func decodeSpotRequestIDParam(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		idStr = getIDFromVars(r)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return id, nil
}

func decodeUpdateSpotRateRequestRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		idStr = getIDFromVars(r)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.UpdateSpotRateRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = id
	return &req, nil
}

func decodeCreateSpotRateResponseRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		idStr = getIDFromVars(r)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.CreateSpotRateResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.SpotRateRequestID = id
	return &req, nil
}

func decodeSpotResponseIDParam(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "responseId")
	if idStr == "" {
		idStr = getIDFromVars(r)
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return id, nil
}

func decodeUpdateSpotRateResponseRequest(_ context.Context, r *http.Request) (interface{}, error) {
	reqIDStr := chi.URLParam(r, "id")
	if reqIDStr == "" {
		reqIDStr = getIDFromVars(r)
	}
	reqID, _ := strconv.ParseInt(reqIDStr, 10, 64)

	respIDStr := chi.URLParam(r, "responseId")
	respID, err := strconv.ParseInt(respIDStr, 10, 64)
	if err != nil || respID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.UpdateSpotRateResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.SpotRateRequestID = reqID
	req.ResponseID = respID
	return &req, nil
}

func decodeSelectPreferredSpotRateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	reqIDStr := chi.URLParam(r, "id")
	if reqIDStr == "" {
		reqIDStr = getIDFromVars(r)
	}
	reqID, _ := strconv.ParseInt(reqIDStr, 10, 64)

	respIDStr := chi.URLParam(r, "responseId")
	respID, err := strconv.ParseInt(respIDStr, 10, 64)
	if err != nil || respID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.SelectPreferredSpotRateRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	req.SpotRateRequestID = reqID
	req.ResponseID = respID
	return &req, nil
}

func decodeGetRateLifecycleEventsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	return limit, nil
}

func decodeSearchCarrierLiveRatesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.CarrierRateSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	// Fall back to query params if body was empty
	if req.OriginPort == "" {
		req.OriginPort = r.URL.Query().Get("origin_port")
	}
	if req.OriginPort == "" {
		req.OriginPort = r.URL.Query().Get("origin")
	}
	if req.DestinationPort == "" {
		req.DestinationPort = r.URL.Query().Get("destination_port")
	}
	if req.DestinationPort == "" {
		req.DestinationPort = r.URL.Query().Get("destination")
	}
	if req.EquipmentType == "" {
		req.EquipmentType = r.URL.Query().Get("equipment_type")
	}
	if req.EquipmentType == "" {
		req.EquipmentType = r.URL.Query().Get("equipment")
	}
	if req.CarrierSCAC == "" {
		req.CarrierSCAC = r.URL.Query().Get("carrier_scac")
	}
	if req.CarrierSCAC == "" {
		req.CarrierSCAC = r.URL.Query().Get("carrier")
	}
	if req.RateType == "" {
		req.RateType = r.URL.Query().Get("rate_type")
	}

	return &req, nil
}
