package shipments

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/shipments/spec"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-chi/chi/v5"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// WebhookError is a domain-specific HTTP error used in webhook decode paths.
type WebhookError struct {
	HTTPStatus int
	Code       string
	Message    string
}

func (e *WebhookError) Error() string {
	return e.Message
}

// AddShipmentHandlers registers all shipment routes onto the given chi.Router.
// Follows the same pattern as rfq.AddRFQHandlers / leads.AddLeadsHandlers.
func AddShipmentHandlers(
	router chi.Router,
	endpoints Endpoints,
	svc Service,
	authMiddleware func(http.Handler) http.Handler,
) {
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	// ── Public API routes (protected by authMiddleware) ────────────────────

	// GET /api/v1/shipments/
	router.With(authMiddleware).Get("/", kitHttp.NewServer(
		endpoints.ListShipmentsEP,
		decodeListShipmentsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/{id}
	router.With(authMiddleware).Get("/{id:[0-9]+}", kitHttp.NewServer(
		endpoints.GetShipmentEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id}/carrier-update
	router.With(authMiddleware).Post("/{id:[0-9]+}/carrier-update", kitHttp.NewServer(
		endpoints.CarrierUpdateEP,
		decodeCarrierUpdateRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// PUT /api/v1/shipments/{id:[0-9]+}/milestones
	router.With(authMiddleware).Put("/{id:[0-9]+}/milestones", kitHttp.NewServer(
		endpoints.UpdateMilestoneEP,
		decodeUpdateMilestoneRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/{id:[0-9]+}/exceptions
	router.With(authMiddleware).Get("/{id:[0-9]+}/exceptions", kitHttp.NewServer(
		endpoints.GetShipmentExceptionsEP,
		decodeGetShipmentExceptionsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/exceptions
	router.With(authMiddleware).Post("/{id:[0-9]+}/exceptions", kitHttp.NewServer(
		endpoints.CreateShipmentExceptionEP,
		decodeCreateShipmentExceptionRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// PATCH /api/v1/shipments/{id:[0-9]+}/exceptions/{exceptionId:[0-9]+}
	router.With(authMiddleware).Patch("/{id:[0-9]+}/exceptions/{exceptionId:[0-9]+}", kitHttp.NewServer(
		endpoints.UpdateShipmentExceptionEP,
		decodeUpdateShipmentExceptionRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/exceptions/{exceptionId:[0-9]+}/acknowledge
	router.With(authMiddleware).Post("/{id:[0-9]+}/exceptions/{exceptionId:[0-9]+}/acknowledge", kitHttp.NewServer(
		endpoints.AcknowledgeShipmentExceptionEP,
		decodeAcknowledgeShipmentExceptionRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/exceptions/{exceptionId:[0-9]+}/resolve
	router.With(authMiddleware).Post("/{id:[0-9]+}/exceptions/{exceptionId:[0-9]+}/resolve", kitHttp.NewServer(
		endpoints.ResolveShipmentExceptionEP,
		decodeResolveShipmentExceptionRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/exceptions/{exceptionId:[0-9]+}/dismiss
	router.With(authMiddleware).Post("/{id:[0-9]+}/exceptions/{exceptionId:[0-9]+}/dismiss", kitHttp.NewServer(
		endpoints.DismissShipmentExceptionEP,
		decodeDismissShipmentExceptionRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/exceptions/evaluate
	router.With(authMiddleware).Post("/{id:[0-9]+}/exceptions/evaluate", kitHttp.NewServer(
		endpoints.EvaluateShipmentExceptionsEP,
		decodeEvaluateShipmentExceptionsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/exceptions/{id}/resolve
	router.With(authMiddleware).Post("/exceptions/{id:[0-9]+}/resolve", kitHttp.NewServer(
		endpoints.ResolveExceptionEP,
		decodeResolveExceptionRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/{id:[0-9]+}/tracking
	router.With(authMiddleware).Get("/{id:[0-9]+}/tracking", kitHttp.NewServer(
		endpoints.GetShipmentTrackingEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/{id:[0-9]+}/tracking/position
	router.With(authMiddleware).Get("/{id:[0-9]+}/tracking/position", kitHttp.NewServer(
		endpoints.GetLatestTrackingPositionEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/{id:[0-9]+}/tracking/positions
	router.With(authMiddleware).Get("/{id:[0-9]+}/tracking/positions", kitHttp.NewServer(
		endpoints.GetTrackingPositionsEP,
		decodeGetTrackingPositionsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/{id:[0-9]+}/tracking/route
	router.With(authMiddleware).Get("/{id:[0-9]+}/tracking/route", kitHttp.NewServer(
		endpoints.GetTrackingRouteEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/{id:[0-9]+}/tracking/events
	router.With(authMiddleware).Get("/{id:[0-9]+}/tracking/events", kitHttp.NewServer(
		endpoints.GetTrackingEventsEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/{id:[0-9]+}/tracking/intelligence (Task 17.4)
	router.With(authMiddleware).Get("/{id:[0-9]+}/tracking/intelligence", kitHttp.NewServer(
		endpoints.GetShipmentTrackingIntelligenceEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/{id:[0-9]+}/tracking/alerts (Task 17.5)
	router.With(authMiddleware).Get("/{id:[0-9]+}/tracking/alerts", kitHttp.NewServer(
		endpoints.GetTrackingAlertsEP,
		decodeGetTrackingAlertsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/{id:[0-9]+}/tracking/monitoring (Task 17.5)
	router.With(authMiddleware).Get("/{id:[0-9]+}/tracking/monitoring", kitHttp.NewServer(
		endpoints.GetTrackingMonitoringSummaryEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/tracking/refresh (Task 17.5 & 17.6)
	router.With(authMiddleware).Post("/{id:[0-9]+}/tracking/refresh", kitHttp.NewServer(
		endpoints.RefreshShipmentTrackingEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/{id:[0-9]+}/tracking/refresh-history (Task 17.7)
	router.With(authMiddleware).Get("/{id:[0-9]+}/tracking/refresh-history", kitHttp.NewServer(
		endpoints.GetTrackingRefreshHistoryEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/tracking/alerts/{alertId:[0-9]+}/acknowledge (Task 17.5)
	router.With(authMiddleware).Post("/{id:[0-9]+}/tracking/alerts/{alertId:[0-9]+}/acknowledge", kitHttp.NewServer(
		endpoints.AcknowledgeTrackingAlertEP,
		decodeAcknowledgeTrackingAlertRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/tracking/alerts/{alertId:[0-9]+}/resolve (Task 17.5)
	router.With(authMiddleware).Post("/{id:[0-9]+}/tracking/alerts/{alertId:[0-9]+}/resolve", kitHttp.NewServer(
		endpoints.ResolveTrackingAlertEP,
		decodeResolveTrackingAlertRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/tracking/alerts/{alertId:[0-9]+}/suppress (Task 17.5)
	router.With(authMiddleware).Post("/{id:[0-9]+}/tracking/alerts/{alertId:[0-9]+}/suppress", kitHttp.NewServer(
		endpoints.SuppressTrackingAlertEP,
		decodeSuppressTrackingAlertRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// ── Tracking Analytics & Operational Insights (Task 17.8) ────────────
	// GET /api/v1/shipments/tracking/analytics/overview
	router.With(authMiddleware).Get("/tracking/analytics/overview", kitHttp.NewServer(
		endpoints.GetTrackingAnalyticsOverviewEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/tracking/analytics/trends
	router.With(authMiddleware).Get("/tracking/analytics/trends", kitHttp.NewServer(
		endpoints.GetTrackingAnalyticsTrendsEP,
		decodeGetTrackingAnalyticsTrendsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/tracking/analytics/carriers
	router.With(authMiddleware).Get("/tracking/analytics/carriers", kitHttp.NewServer(
		endpoints.GetCarrierTrackingPerformanceEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/tracking/analytics/routes
	router.With(authMiddleware).Get("/tracking/analytics/routes", kitHttp.NewServer(
		endpoints.GetRouteTrackingPerformanceEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/{id:[0-9]+}/closure
	router.With(authMiddleware).Get("/{id:[0-9]+}/closure", kitHttp.NewServer(
		endpoints.GetShipmentClosureEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/evaluate
	router.With(authMiddleware).Post("/{id:[0-9]+}/evaluate", kitHttp.NewServer(
		endpoints.EvaluateShipmentClosureEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/request-closure
	router.With(authMiddleware).Post("/{id:[0-9]+}/request-closure", kitHttp.NewServer(
		endpoints.RequestShipmentClosureEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/complete
	router.With(authMiddleware).Post("/{id:[0-9]+}/complete", kitHttp.NewServer(
		endpoints.CompleteShipmentEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/reopen
	router.With(authMiddleware).Post("/{id:[0-9]+}/reopen", kitHttp.NewServer(
		endpoints.ReopenShipmentEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// ─── Shipment Document Routes (Task 16.7) ──────────────────────────────────
	// GET /api/v1/shipments/{id:[0-9]+}/documents
	router.With(authMiddleware).Get("/{id:[0-9]+}/documents", kitHttp.NewServer(
		endpoints.GetShipmentDocumentsEP,
		decodeGetShipmentDocumentsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/documents
	router.With(authMiddleware).Post("/{id:[0-9]+}/documents", kitHttp.NewServer(
		endpoints.CreateShipmentDocumentEP,
		decodeCreateShipmentDocumentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// PATCH /api/v1/shipments/{id:[0-9]+}/documents/{docId:[0-9]+}
	router.With(authMiddleware).Patch("/{id:[0-9]+}/documents/{docId:[0-9]+}", kitHttp.NewServer(
		endpoints.UpdateShipmentDocumentEP,
		decodeUpdateShipmentDocumentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/documents/{docId:[0-9]+}/approve
	router.With(authMiddleware).Post("/{id:[0-9]+}/documents/{docId:[0-9]+}/approve", kitHttp.NewServer(
		endpoints.ApproveShipmentDocumentEP,
		decodeApproveShipmentDocumentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/documents/{docId:[0-9]+}/reject
	router.With(authMiddleware).Post("/{id:[0-9]+}/documents/{docId:[0-9]+}/reject", kitHttp.NewServer(
		endpoints.RejectShipmentDocumentEP,
		decodeRejectShipmentDocumentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// DELETE /api/v1/shipments/{id:[0-9]+}/documents/{docId:[0-9]+}
	router.With(authMiddleware).Delete("/{id:[0-9]+}/documents/{docId:[0-9]+}", kitHttp.NewServer(
		endpoints.DeleteShipmentDocumentEP,
		decodeDeleteShipmentDocumentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// ─── Shipment Financial Operations Routes (Task 16.8) ──────────────────────
	// GET /api/v1/shipments/{id:[0-9]+}/financials
	router.With(authMiddleware).Get("/{id:[0-9]+}/financials", kitHttp.NewServer(
		endpoints.GetShipmentFinancialsEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/shipments/{id:[0-9]+}/financials/charges
	router.With(authMiddleware).Get("/{id:[0-9]+}/financials/charges", kitHttp.NewServer(
		endpoints.GetShipmentChargesEP,
		decodeGetShipmentChargesRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/financials/charges
	router.With(authMiddleware).Post("/{id:[0-9]+}/financials/charges", kitHttp.NewServer(
		endpoints.CreateShipmentChargeEP,
		decodeCreateShipmentChargeRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// PATCH /api/v1/shipments/{id:[0-9]+}/financials/charges/{chargeId:[0-9]+}
	router.With(authMiddleware).Patch("/{id:[0-9]+}/financials/charges/{chargeId:[0-9]+}", kitHttp.NewServer(
		endpoints.UpdateShipmentChargeEP,
		decodeUpdateShipmentChargeRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// DELETE /api/v1/shipments/{id:[0-9]+}/financials/charges/{chargeId:[0-9]+}
	router.With(authMiddleware).Delete("/{id:[0-9]+}/financials/charges/{chargeId:[0-9]+}", kitHttp.NewServer(
		endpoints.DeleteShipmentChargeEP,
		decodeDeleteShipmentChargeRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/financials/recalculate
	router.With(authMiddleware).Post("/{id:[0-9]+}/financials/recalculate", kitHttp.NewServer(
		endpoints.RecalculateShipmentFinancialsEP,
		decodeGetShipmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /api/v1/shipments/{id:[0-9]+}/financials/review
	router.With(authMiddleware).Post("/{id:[0-9]+}/financials/review", kitHttp.NewServer(
		endpoints.ReviewShipmentFinancialsEP,
		decodeReviewShipmentFinancialsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)
}

// AddShipmentWebhookHandlers registers the carrier inbound email webhook route.
// Mounted directly on /api/v1 (no shipments prefix), protected by authMiddleware.
func AddShipmentEmailWebhookHandler(
	router chi.Router,
	endpoints Endpoints,
	svc Service,
	authMiddleware func(http.Handler) http.Handler,
) {
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	// POST /api/v1/emails/carrier-inbound
	router.With(authMiddleware).Post("/emails/carrier-inbound", kitHttp.NewServer(
		endpoints.InboundCarrierEmailEP,
		decodeInboundCarrierEmailRequest(svc),
		encodeAPIResponse,
		options...,
	).ServeHTTP)
}

// AddShipmentWebhookHandlers registers public (signature-verified) carrier webhook routes.
// Mounted at the root router level — no JWT auth, verified by HMAC signature.
func AddCarrierWebhookHandlers(
	router chi.Router,
	endpoints Endpoints,
	svc Service,
) {
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	// POST /webhooks/carriers/{carrier}
	router.Post("/webhooks/carriers/{carrier}", kitHttp.NewServer(
		endpoints.InboundWebhookEP,
		decodeInboundWebhookRequest(svc),
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /webhooks/carriers/{carrier}/{integration_id}
	router.Post("/webhooks/carriers/{carrier}/{integration_id}", kitHttp.NewServer(
		endpoints.InboundWebhookEP,
		decodeInboundWebhookRequest(svc),
		encodeAPIResponse,
		options...,
	).ServeHTTP)
}

// AddShipmentInternalHandlers registers internal (service-to-service auth) shipment routes.
// Mounted on /internal router, protected by InternalServiceAuthMiddleware.
func AddShipmentInternalHandlers(
	router chi.Router,
	endpoints Endpoints,
) {
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	// GET /internal/shipments/{id}
	router.Get("/shipments/{id:[0-9]+}", kitHttp.NewServer(
		endpoints.GetShipmentInternalEP,
		decodeGetShipmentInternalRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /internal/shipments/{id}/milestones
	router.Post("/shipments/{id:[0-9]+}/milestones", kitHttp.NewServer(
		endpoints.UpdateMilestoneInternalEP,
		decodeUpdateMilestoneInternalRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /internal/shipments/{id}/exceptions
	router.Post("/shipments/{id:[0-9]+}/exceptions", kitHttp.NewServer(
		endpoints.CreateExceptionInternalEP,
		decodeCreateExceptionInternalRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// POST /internal/operations/callback
	router.Post("/operations/callback", kitHttp.NewServer(
		endpoints.CallbackInternalEP,
		decodeCallbackInternalRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)
}

// ── Decoders ──────────────────────────────────────────────────────────────────

func getIDFromVars(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		vars := mux.Vars(r)
		idStr = vars["id"]
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

func decodeListShipmentsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	workspace := r.URL.Query().Get("workspace")

	req := &spec.ListShipmentsRequest{
		Workspace: workspace == "true",
	}
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			req.Page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = l
		}
	}
	if status != "" {
		req.Status = &status
	}
	if search != "" {
		req.Search = &search
	}

	return req, nil
}

func decodeGetShipmentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetShipmentRequest{ID: id}, nil
}

func decodeGetTrackingPositionsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	return &spec.GetTrackingPositionsRequest{ID: id, Limit: limit}, nil
}

func decodeGetTrackingAlertsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	status := r.URL.Query().Get("status")
	return &spec.GetTrackingAlertsRequest{ShipmentID: id, Status: status}, nil
}

func decodeAcknowledgeTrackingAlertRequest(_ context.Context, r *http.Request) (interface{}, error) {
	shipmentID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	alertIDStr := chi.URLParam(r, "alertId")
	alertID, err := strconv.ParseInt(alertIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.AcknowledgeTrackingAlertRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	req.ShipmentID = shipmentID
	req.AlertID = alertID
	return &req, nil
}

func decodeResolveTrackingAlertRequest(_ context.Context, r *http.Request) (interface{}, error) {
	shipmentID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	alertIDStr := chi.URLParam(r, "alertId")
	alertID, err := strconv.ParseInt(alertIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.ResolveTrackingAlertRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	req.ShipmentID = shipmentID
	req.AlertID = alertID
	return &req, nil
}

func decodeSuppressTrackingAlertRequest(_ context.Context, r *http.Request) (interface{}, error) {
	shipmentID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	alertIDStr := chi.URLParam(r, "alertId")
	alertID, err := strconv.ParseInt(alertIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.SuppressTrackingAlertRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	req.ShipmentID = shipmentID
	req.AlertID = alertID
	return &req, nil
}

func decodeCarrierUpdateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	var req spec.CarrierUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = id
	return &req, nil
}

func decodeResolveExceptionRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.ResolveExceptionRequest{ID: id}, nil
}

func decodeUpdateMilestoneRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	var req spec.UpdateMilestoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = id
	return &req, nil
}

func decodeGetShipmentExceptionsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.EvaluateExceptionsRequest{ShipmentID: id}, nil
}

func decodeCreateShipmentExceptionRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	var req spec.CreateExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ShipmentID = id
	return &req, nil
}

func decodeUpdateShipmentExceptionRequest(_ context.Context, r *http.Request) (interface{}, error) {
	shipmentID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	exceptionIDStr := chi.URLParam(r, "exceptionId")
	exceptionID, err := strconv.ParseInt(exceptionIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req spec.UpdateExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = exceptionID
	req.ShipmentID = shipmentID
	return &req, nil
}

func decodeAcknowledgeShipmentExceptionRequest(_ context.Context, r *http.Request) (interface{}, error) {
	shipmentID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	exceptionIDStr := chi.URLParam(r, "exceptionId")
	exceptionID, err := strconv.ParseInt(exceptionIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &spec.AcknowledgeExceptionRequest{
		ID:         exceptionID,
		ShipmentID: shipmentID,
	}, nil
}

func decodeResolveShipmentExceptionRequest(_ context.Context, r *http.Request) (interface{}, error) {
	shipmentID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	exceptionIDStr := chi.URLParam(r, "exceptionId")
	exceptionID, err := strconv.ParseInt(exceptionIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var req spec.ResolveShipmentExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = exceptionID
	req.ShipmentID = shipmentID
	return &req, nil
}

func decodeDismissShipmentExceptionRequest(_ context.Context, r *http.Request) (interface{}, error) {
	shipmentID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	exceptionIDStr := chi.URLParam(r, "exceptionId")
	exceptionID, err := strconv.ParseInt(exceptionIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &spec.DismissExceptionRequest{
		ID:         exceptionID,
		ShipmentID: shipmentID,
	}, nil
}

func decodeEvaluateShipmentExceptionsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.EvaluateExceptionsRequest{ShipmentID: id}, nil
}

// ─── Document Decoders (Task 16.7) ───────────────────────────────────────────

func decodeGetShipmentDocumentsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	category := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")
	return &spec.GetShipmentDocumentsRequest{
		ShipmentID: id,
		Category:   category,
		Status:     status,
	}, nil
}

func decodeCreateShipmentDocumentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	var req spec.CreateShipmentDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ShipmentID = id
	return &req, nil
}

func decodeUpdateShipmentDocumentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	shipmentID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	docIDStr := chi.URLParam(r, "docId")
	docID, err := strconv.ParseInt(docIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var req spec.UpdateShipmentDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = docID
	req.ShipmentID = shipmentID
	return &req, nil
}

func decodeApproveShipmentDocumentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	shipmentID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	docIDStr := chi.URLParam(r, "docId")
	docID, err := strconv.ParseInt(docIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &spec.ApproveShipmentDocumentRequest{
		ID:         docID,
		ShipmentID: shipmentID,
	}, nil
}

func decodeRejectShipmentDocumentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	shipmentID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	docIDStr := chi.URLParam(r, "docId")
	docID, err := strconv.ParseInt(docIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var req spec.RejectShipmentDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = docID
	req.ShipmentID = shipmentID
	return &req, nil
}

func decodeDeleteShipmentDocumentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	shipmentID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	docIDStr := chi.URLParam(r, "docId")
	docID, err := strconv.ParseInt(docIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &spec.DeleteShipmentDocumentRequest{
		ID:         docID,
		ShipmentID: shipmentID,
	}, nil
}

// ─── Financial Decoders (Task 16.8) ─────────────────────────────────────────

func decodeGetShipmentChargesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	category := r.URL.Query().Get("category")
	chargeType := r.URL.Query().Get("charge_type")
	return &spec.GetShipmentChargesRequest{
		ShipmentID: id,
		Category:   category,
		ChargeType: chargeType,
	}, nil
}

func decodeCreateShipmentChargeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	var req spec.CreateShipmentChargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ShipmentID = id
	return &req, nil
}

func decodeUpdateShipmentChargeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	shipmentID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	chargeIDStr := chi.URLParam(r, "chargeId")
	chargeID, err := strconv.ParseInt(chargeIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var req spec.UpdateShipmentChargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = chargeID
	req.ShipmentID = shipmentID
	return &req, nil
}

func decodeDeleteShipmentChargeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	shipmentID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	chargeIDStr := chi.URLParam(r, "chargeId")
	chargeID, err := strconv.ParseInt(chargeIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &spec.DeleteShipmentChargeRequest{
		ID:         chargeID,
		ShipmentID: shipmentID,
	}, nil
}

func decodeReviewShipmentFinancialsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	shipmentID, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	var req spec.ReviewShipmentFinancialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ShipmentID = shipmentID
	return &req, nil
}



func decodeInboundCarrierEmailRequest(svc Service) func(context.Context, *http.Request) (interface{}, error) {
	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		var req spec.InboundCarrierEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, &WebhookError{HTTPStatus: http.StatusBadRequest, Code: "INVALID_PAYLOAD", Message: "Invalid request body"}
		}

		if req.From == "" || req.Body == "" {
			return nil, &WebhookError{HTTPStatus: http.StatusBadRequest, Code: "MISSING_PARAMS", Message: "From and Body are required fields"}
		}

		normalized, err := ParseCarrierEmail(&spec.CarrierEmailRequest{
			From:      req.From,
			To:        req.To,
			Subject:   req.Subject,
			Body:      req.Body,
			MessageID: req.MessageID,
		})
		if err != nil {
			return nil, &WebhookError{HTTPStatus: http.StatusBadRequest, Code: "PARSE_ERROR", Message: "Failed to parse carrier email: " + err.Error()}
		}

		var orgID int64
		sImpl := svc.(*serviceImpl)
		dbErr := sImpl.db.GetContext(ctx, &orgID, `
			SELECT org_id FROM carrier_integrations 
			WHERE carrier_scac = ? AND is_active = 1 LIMIT 1
		`, normalized.CarrierSCAC)
		if dbErr != nil {
			userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
			if ok && userCtx.OrgID > 0 {
				orgID = userCtx.OrgID
			} else if os.Getenv("APP_ENV") != "production" {
				_ = sImpl.db.GetContext(ctx, &orgID, "SELECT id FROM organizations LIMIT 1")
			}
		}

		if orgID <= 0 {
			return nil, &WebhookError{HTTPStatus: http.StatusBadRequest, Code: "RESOLVE_ERROR", Message: "Unable to resolve org_id for carrier email"}
		}
		req.OrgID = orgID
		return &req, nil
	}
}

func decodeInboundWebhookRequest(svc Service) func(context.Context, *http.Request) (interface{}, error) {
	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		carrierParam := chi.URLParam(r, "carrier")
		if carrierParam == "" {
			return nil, &WebhookError{HTTPStatus: http.StatusBadRequest, Code: "MISSING_PARAM", Message: "carrier identifier is required"}
		}

		integrationIDParam := chi.URLParam(r, "integration_id")
		var orgID int64
		var dbCarrierSCAC string
		var isActive bool
		var err error
		var integrationID int64

		sImpl := svc.(*serviceImpl)

		if integrationIDParam != "" {
			integrationID, err = strconv.ParseInt(integrationIDParam, 10, 64)
			if err != nil {
				return nil, &WebhookError{HTTPStatus: http.StatusBadRequest, Code: "INVALID_PARAM", Message: "invalid integration_id format"}
			}

			err = sImpl.db.QueryRowContext(ctx, `
				SELECT org_id, carrier_scac, is_active FROM carrier_integrations 
				WHERE id = ? LIMIT 1
			`, integrationID).Scan(&orgID, &dbCarrierSCAC, &isActive)
			if err != nil {
				return nil, &WebhookError{HTTPStatus: http.StatusNotFound, Code: "NOT_FOUND", Message: "carrier integration not configured or inactive"}
			}

			if !isActive {
				return nil, &WebhookError{HTTPStatus: http.StatusBadRequest, Code: "INACTIVE_INTEGRATION", Message: "carrier integration is inactive"}
			}

			scacUpper := strings.ToUpper(carrierParam)
			dbScacUpper := strings.ToUpper(dbCarrierSCAC)
			isMaerskMatch := (scacUpper == "MAEU" || scacUpper == "MSK") && (dbScacUpper == "MAEU" || dbScacUpper == "MSK")
			if scacUpper != dbScacUpper && !isMaerskMatch {
				return nil, &WebhookError{HTTPStatus: http.StatusBadRequest, Code: "CARRIER_MISMATCH", Message: "carrier mismatch for specified integration ID"}
			}
		} else {
			if os.Getenv("APP_ENV") == "production" {
				return nil, &WebhookError{HTTPStatus: http.StatusBadRequest, Code: "MISSING_INTEGRATION_ID", Message: "production error: integration_id is required in webhook URL path"}
			}

			dbErr := sImpl.db.GetContext(ctx, &orgID, `
				SELECT org_id FROM carrier_integrations 
				WHERE carrier_scac = ? AND is_active = 1 LIMIT 1
			`, carrierParam)
			if dbErr != nil {
				userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
				if ok && userCtx.OrgID > 0 {
					orgID = userCtx.OrgID
				} else {
					_ = sImpl.db.GetContext(ctx, &orgID, "SELECT id FROM organizations LIMIT 1")
				}
			}
		}

		if orgID <= 0 {
			return nil, &WebhookError{HTTPStatus: http.StatusBadRequest, Code: "RESOLVE_ERROR", Message: "unable to resolve org_id for integration credentials"}
		}

		var body []byte
		if r.Body != nil {
			defer r.Body.Close()
			body, err = io.ReadAll(r.Body)
			if err != nil {
				return nil, &WebhookError{HTTPStatus: http.StatusBadRequest, Code: "READ_ERROR", Message: "failed to read request body"}
			}
		}

		headers := make(map[string]string)
		for k, v := range r.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		return &spec.InboundWebhookRequest{
			Carrier:       carrierParam,
			IntegrationID: integrationID,
			OrgID:         orgID,
			Body:          body,
			Headers:       headers,
		}, nil
	}
}

func decodeGetShipmentInternalRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}

	orgIDStr := r.URL.Query().Get("org_id")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil || orgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return &spec.GetShipmentInternalRequest{
		ID:    id,
		OrgID: orgID,
	}, nil
}

func decodeUpdateMilestoneInternalRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}

	var req spec.UpdateMilestoneInternalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	if req.OrgID == nil || *req.OrgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	req.ID = id
	return &req, nil
}

func decodeCreateExceptionInternalRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}

	var req spec.CreateExceptionInternalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	if req.OrgID == nil || *req.OrgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	req.ID = id
	return &req, nil
}

func decodeCallbackInternalRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.CallbackInternalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

// ── Encoders ──────────────────────────────────────────────────────────────────

func encodeAPIResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func encodeErrorResponse(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if wErr, ok := err.(*WebhookError); ok {
		w.WriteHeader(wErr.HTTPStatus)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    wErr.Code,
				"message": wErr.Message,
			},
		})
		return
	}

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
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    svcErr.Code,
				"message": svcErr.Message,
			},
		})
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    svcerror.ErrInternal,
			"message": err.Error(),
		},
	})
}

// AddTrackingAnalyticsHandlers mounts standalone tracking analytics endpoints onto /api/v1/tracking (Task 17.8)
func AddTrackingAnalyticsHandlers(
	router chi.Router,
	endpoints Endpoints,
	svc Service,
	authMiddleware func(http.Handler) http.Handler,
) {
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	// GET /api/v1/tracking/analytics/overview
	router.With(authMiddleware).Get("/analytics/overview", kitHttp.NewServer(
		endpoints.GetTrackingAnalyticsOverviewEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/tracking/analytics/trends
	router.With(authMiddleware).Get("/analytics/trends", kitHttp.NewServer(
		endpoints.GetTrackingAnalyticsTrendsEP,
		decodeGetTrackingAnalyticsTrendsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/tracking/analytics/carriers
	router.With(authMiddleware).Get("/analytics/carriers", kitHttp.NewServer(
		endpoints.GetCarrierTrackingPerformanceEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// GET /api/v1/tracking/analytics/routes
	router.With(authMiddleware).Get("/analytics/routes", kitHttp.NewServer(
		endpoints.GetRouteTrackingPerformanceEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)
}

func decodeEmptyRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeGetTrackingAnalyticsTrendsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	days := 14
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}
	return &GetTrackingAnalyticsTrendsRequest{Days: days}, nil
}
