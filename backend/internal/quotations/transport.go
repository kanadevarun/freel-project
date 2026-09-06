package quotations

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-chi/chi/v5"
	kitHttp "github.com/go-kit/kit/transport/http"
)

// AddQuotationHandlers registers all quotation REST routes on the given chi.Router.
func AddQuotationHandlers(
	router chi.Router,
	endpoints Endpoints,
	authMiddleware func(http.Handler) http.Handler,
) {
	opts := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeQuotationError),
	}

	// ── Reusable Templates Routes (Task 18.3) ──────────────────────────────────
	// GET  /api/v1/quotations/templates
	router.With(authMiddleware).Get("/templates", kitHttp.NewServer(
		endpoints.ListQuotationTemplatesEP,
		decodeListTemplatesRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/templates
	router.With(authMiddleware).Post("/templates", kitHttp.NewServer(
		endpoints.CreateQuotationTemplateEP,
		decodeCreateTemplateRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/templates/{templateId}
	router.With(authMiddleware).Get("/templates/{templateId:[0-9]+}", kitHttp.NewServer(
		endpoints.GetQuotationTemplateEP,
		decodeGetTemplateRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// PUT  /api/v1/quotations/templates/{templateId}
	router.With(authMiddleware).Put("/templates/{templateId:[0-9]+}", kitHttp.NewServer(
		endpoints.UpdateQuotationTemplateEP,
		decodeUpdateTemplateRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// DELETE /api/v1/quotations/templates/{templateId}
	router.With(authMiddleware).Delete("/templates/{templateId:[0-9]+}", kitHttp.NewServer(
		endpoints.DeleteQuotationTemplateEP,
		decodeDeleteTemplateRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// ── Summary & List Routes ──────────────────────────────────────────────────
	// GET  /api/v1/quotations/summary
	router.With(authMiddleware).Get("/summary", kitHttp.NewServer(
		endpoints.GetQuotationSummaryEP,
		decodeEmpty,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/
	router.With(authMiddleware).Get("/", kitHttp.NewServer(
		endpoints.ListQuotationsEP,
		decodeListQuotationsRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/
	router.With(authMiddleware).Post("/", kitHttp.NewServer(
		endpoints.CreateQuotationEP,
		decodeCreateQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}
	router.With(authMiddleware).Get("/{id:[0-9]+}", kitHttp.NewServer(
		endpoints.GetQuotationEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// PUT  /api/v1/quotations/{id}
	router.With(authMiddleware).Put("/{id:[0-9]+}", kitHttp.NewServer(
		endpoints.UpdateQuotationEP,
		decodeUpdateQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// ── Pricing & Charges Routes (Task 18.2) ───────────────────────────────────

	// GET  /api/v1/quotations/{id}/pricing
	router.With(authMiddleware).Get("/{id:[0-9]+}/pricing", kitHttp.NewServer(
		endpoints.GetQuotationPricingEP,
		decodeGetQuotationPricingRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/charges
	router.With(authMiddleware).Post("/{id:[0-9]+}/charges", kitHttp.NewServer(
		endpoints.AddQuotationChargeEP,
		decodeAddQuotationChargeRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// PUT  /api/v1/quotations/{id}/charges/{chargeId}
	router.With(authMiddleware).Put("/{id:[0-9]+}/charges/{chargeId:[0-9]+}", kitHttp.NewServer(
		endpoints.UpdateQuotationChargeEP,
		decodeUpdateQuotationChargeRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// DELETE /api/v1/quotations/{id}/charges/{chargeId}
	router.With(authMiddleware).Delete("/{id:[0-9]+}/charges/{chargeId:[0-9]+}", kitHttp.NewServer(
		endpoints.DeleteQuotationChargeEP,
		decodeDeleteQuotationChargeRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/charges/reorder
	router.With(authMiddleware).Post("/{id:[0-9]+}/charges/reorder", kitHttp.NewServer(
		endpoints.ReorderQuotationChargesEP,
		decodeReorderQuotationChargesRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/rate-candidates
	router.With(authMiddleware).Get("/{id:[0-9]+}/rate-candidates", kitHttp.NewServer(
		endpoints.GetRateCandidatesEP,
		decodeGetRateCandidatesRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/charges/import-rate
	router.With(authMiddleware).Post("/{id:[0-9]+}/charges/import-rate", kitHttp.NewServer(
		endpoints.ImportRateChargesEP,
		decodeImportRateChargesRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// ── Commercial Terms & Quotation Template Actions (Task 18.3) ──────────────

	// POST /api/v1/quotations/{id}/apply-template
	router.With(authMiddleware).Post("/{id:[0-9]+}/apply-template", kitHttp.NewServer(
		endpoints.ApplyTemplateToQuotationEP,
		decodeApplyTemplateRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/save-as-template
	router.With(authMiddleware).Post("/{id:[0-9]+}/save-as-template", kitHttp.NewServer(
		endpoints.CreateTemplateFromQuotationEP,
		decodeSaveAsTemplateRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// PUT  /api/v1/quotations/{id}/commercial-terms
	router.With(authMiddleware).Put("/{id:[0-9]+}/commercial-terms", kitHttp.NewServer(
		endpoints.UpdateQuotationCommercialTermsEP,
		decodeUpdateCommercialTermsRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// ── Lifecycle & Review Actions (Task 18.4) ─────────────────────────────────

	// POST /api/v1/quotations/{id}/submit-review
	router.With(authMiddleware).Post("/{id:[0-9]+}/submit-review", kitHttp.NewServer(
		endpoints.SubmitQuotationForReviewEP,
		decodeSubmitReviewRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/approve
	router.With(authMiddleware).Post("/{id:[0-9]+}/approve", kitHttp.NewServer(
		endpoints.ApproveQuotationEP,
		decodeApproveQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/request-changes
	router.With(authMiddleware).Post("/{id:[0-9]+}/request-changes", kitHttp.NewServer(
		endpoints.RequestQuotationChangesEP,
		decodeRequestChangesRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/send
	router.With(authMiddleware).Post("/{id:[0-9]+}/send", kitHttp.NewServer(
		endpoints.SendQuotationEP,
		decodeSendQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/customer-preview
	router.With(authMiddleware).Get("/{id:[0-9]+}/customer-preview", kitHttp.NewServer(
		endpoints.GetCustomerQuotationPreviewEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/approval-history
	router.With(authMiddleware).Get("/{id:[0-9]+}/approval-history", kitHttp.NewServer(
		endpoints.GetQuotationApprovalHistoryEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/approval-status
	router.With(authMiddleware).Get("/{id:[0-9]+}/approval-status", kitHttp.NewServer(
		endpoints.GetQuotationApprovalStatusEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/view
	router.With(authMiddleware).Post("/{id:[0-9]+}/view", kitHttp.NewServer(
		endpoints.MarkQuotationViewedEP,
		decodeMarkViewedRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/accept
	router.With(authMiddleware).Post("/{id:[0-9]+}/accept", kitHttp.NewServer(
		endpoints.AcceptQuotationEP,
		decodeAcceptQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/decline
	router.With(authMiddleware).Post("/{id:[0-9]+}/decline", kitHttp.NewServer(
		endpoints.DeclineQuotationEP,
		decodeDeclineQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/cancel
	router.With(authMiddleware).Post("/{id:[0-9]+}/cancel", kitHttp.NewServer(
		endpoints.CancelQuotationEP,
		decodeCancelQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// ── Documents & PDF Generation (Task 18.5) ─────────────────────────────────

	// POST /api/v1/quotations/{id}/documents/generate
	router.With(authMiddleware).Post("/{id:[0-9]+}/documents/generate", kitHttp.NewServer(
		endpoints.GenerateQuotationDocumentEP,
		decodeGenerateDocumentRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/documents
	router.With(authMiddleware).Get("/{id:[0-9]+}/documents", kitHttp.NewServer(
		endpoints.ListQuotationDocumentsEP,
		decodeListDocumentsRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/documents/{docId}
	router.With(authMiddleware).Get("/{id:[0-9]+}/documents/{docId:[0-9]+}", kitHttp.NewServer(
		endpoints.GetQuotationDocumentEP,
		decodeGetDocumentRequest,
		encodeQuotationDocumentResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/documents/{docId}/download
	router.With(authMiddleware).Get("/{id:[0-9]+}/documents/{docId:[0-9]+}/download", kitHttp.NewServer(
		endpoints.GetQuotationDocumentEP,
		decodeGetDocumentRequest,
		encodeQuotationDocumentDownload,
		opts...,
	).ServeHTTP)

	// ── Public Sharing Links (Task 18.5) ───────────────────────────────────────

	// POST /api/v1/quotations/{id}/public-links
	router.With(authMiddleware).Post("/{id:[0-9]+}/public-links", kitHttp.NewServer(
		endpoints.CreateQuotationPublicLinkEP,
		decodeCreatePublicLinkRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/public-links
	router.With(authMiddleware).Get("/{id:[0-9]+}/public-links", kitHttp.NewServer(
		endpoints.ListQuotationPublicLinksEP,
		decodeListPublicLinksRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/public-links/{linkId}/revoke
	router.With(authMiddleware).Post("/{id:[0-9]+}/public-links/{linkId:[0-9]+}/revoke", kitHttp.NewServer(
		endpoints.RevokeQuotationPublicLinkEP,
		decodeRevokePublicLinkRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// ── Quotation-to-Booking Operational Conversion Routes (Task 18.6) ──────────

	// GET  /api/v1/quotations/{id}/conversion-preview
	router.With(authMiddleware).Get("/{id:[0-9]+}/conversion-preview", kitHttp.NewServer(
		endpoints.GetQuotationConversionPreviewEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/convert-to-booking
	router.With(authMiddleware).Post("/{id:[0-9]+}/convert-to-booking", kitHttp.NewServer(
		endpoints.ConvertQuotationToBookingEP,
		decodeConvertQuotationToBookingRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/conversion-history
	router.With(authMiddleware).Get("/{id:[0-9]+}/conversion-history", kitHttp.NewServer(
		endpoints.GetQuotationConversionHistoryEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// ── Booking Confirmation & Handover Traceability Routes (Task 18.7) ─────────

	// GET  /api/v1/quotations/{id}/operational-handover
	router.With(authMiddleware).Get("/{id:[0-9]+}/operational-handover", kitHttp.NewServer(
		endpoints.GetQuotationOperationalHandoverEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/confirm-handover
	router.With(authMiddleware).Post("/{id:[0-9]+}/confirm-handover", kitHttp.NewServer(
		endpoints.ConfirmQuotationBookingHandoverEP,
		decodeConfirmHandoverRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/operational-changes
	router.With(authMiddleware).Get("/{id:[0-9]+}/operational-changes", kitHttp.NewServer(
		endpoints.GetQuotationOperationalChangesEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/handover-history
	router.With(authMiddleware).Get("/{id:[0-9]+}/handover-history", kitHttp.NewServer(
		endpoints.GetQuotationHandoverHistoryEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// ── Quotation Analytics & Intelligence Routes (Task 18.8) ──────────────────

	// GET  /api/v1/quotations/analytics/overview
	router.With(authMiddleware).Get("/analytics/overview", kitHttp.NewServer(
		endpoints.GetQuotationAnalyticsOverviewEP,
		decodeEmptyAnalyticsRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/analytics/trends
	router.With(authMiddleware).Get("/analytics/trends", kitHttp.NewServer(
		endpoints.GetQuotationAnalyticsTrendsEP,
		decodeAnalyticsTrendsRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/analytics/customers
	router.With(authMiddleware).Get("/analytics/customers", kitHttp.NewServer(
		endpoints.GetCustomerQuotationPerformanceEP,
		decodeEmptyAnalyticsRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/analytics/modes
	router.With(authMiddleware).Get("/analytics/modes", kitHttp.NewServer(
		endpoints.GetQuotationPerformanceByModeEP,
		decodeEmptyAnalyticsRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/analytics/expiry-risk
	router.With(authMiddleware).Get("/analytics/expiry-risk", kitHttp.NewServer(
		endpoints.GetQuotationExpiryRiskEP,
		decodeEmptyAnalyticsRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// ── Task 19.5: Rate-to-Quotation Integration Routes ────────────────────────

	// GET  /api/v1/quotations/{id}/rate-candidates
	router.With(authMiddleware).Get("/{id:[0-9]+}/rate-candidates", kitHttp.NewServer(
		endpoints.GetQuotationRateCandidatesEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/rate-selection
	router.With(authMiddleware).Get("/{id:[0-9]+}/rate-selection", kitHttp.NewServer(
		endpoints.GetQuotationRateSelectionEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/rate-selection
	router.With(authMiddleware).Post("/{id:[0-9]+}/rate-selection", kitHttp.NewServer(
		endpoints.SelectQuotationRateEP,
		decodeSelectQuotationRateRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// PUT  /api/v1/quotations/{id}/rate-selection
	router.With(authMiddleware).Put("/{id:[0-9]+}/rate-selection", kitHttp.NewServer(
		endpoints.ReplaceQuotationRateEP,
		decodeReplaceQuotationRateRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// DELETE /api/v1/quotations/{id}/rate-selection
	router.With(authMiddleware).Delete("/{id:[0-9]+}/rate-selection", kitHttp.NewServer(
		endpoints.RemoveQuotationRateEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/rate-snapshot
	router.With(authMiddleware).Get("/{id:[0-9]+}/rate-snapshot", kitHttp.NewServer(
		endpoints.GetQuotationRateSnapshotEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/rate-selection-history
	router.With(authMiddleware).Get("/{id:[0-9]+}/rate-selection-history", kitHttp.NewServer(
		endpoints.GetQuotationRateSelectionHistoryEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// ── Task 19.6: Rate Lifecycle Intelligence & Commercial Risk Routes ─────────

	// GET  /api/v1/quotations/{id}/rate-risks
	router.With(authMiddleware).Get("/{id:[0-9]+}/rate-risks", kitHttp.NewServer(
		endpoints.GetQuotationRateRisksEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/{id}/rate-risks/{riskId}/resolve
	router.With(authMiddleware).Post("/{id:[0-9]+}/rate-risks/{riskId:[0-9]+}/resolve", kitHttp.NewServer(
		endpoints.ResolveQuotationRateRiskEP,
		decodeResolveQuotationRateRiskRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/replacement-rate-candidates
	router.With(authMiddleware).Get("/{id:[0-9]+}/replacement-rate-candidates", kitHttp.NewServer(
		endpoints.GetRateReplacementCandidatesEP,
		decodeGetQuotationRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// GET  /api/v1/quotations/{id}/commercial-impact
	router.With(authMiddleware).Get("/{id:[0-9]+}/commercial-impact", kitHttp.NewServer(
		endpoints.GetCommercialImpactAnalysisEP,
		decodeCommercialImpactRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/quotations/rate-risks/evaluate
	router.With(authMiddleware).Post("/rate-risks/evaluate", kitHttp.NewServer(
		endpoints.EvaluateQuotationRateRisksEP,
		decodeEmptyAnalyticsRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)
}

// AddPublicQuotationHandlers registers public unauthenticated customer portal routes.
func AddPublicQuotationHandlers(
	router chi.Router,
	endpoints Endpoints,
) {
	opts := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeQuotationError),
	}

	// GET  /api/v1/public/quotations/{token}
	router.Get("/api/v1/public/quotations/{token}", kitHttp.NewServer(
		endpoints.GetPublicQuotationByTokenEP,
		decodePublicTokenRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/public/quotations/{token}/accept
	router.Post("/api/v1/public/quotations/{token}/accept", kitHttp.NewServer(
		endpoints.PublicAcceptQuotationEP,
		decodePublicAcceptRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)

	// POST /api/v1/public/quotations/{token}/decline
	router.Post("/api/v1/public/quotations/{token}/decline", kitHttp.NewServer(
		endpoints.PublicDeclineQuotationEP,
		decodePublicDeclineRequest,
		encodeQuotationResponse,
		opts...,
	).ServeHTTP)
}

// ── Decoders ─────────────────────────────────────────────────────────────────

func decodeEmpty(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeListQuotationsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	q := r.URL.Query()

	filters := &QuotationListFilters{
		Search:   q.Get("search"),
		Status:   q.Get("status"),
		Validity: q.Get("validity"),
	}

	if cidStr := q.Get("customer_id"); cidStr != "" {
		cid, err := strconv.ParseInt(cidStr, 10, 64)
		if err == nil {
			filters.CustomerID = &cid
		}
	}

	filters.Page, _ = strconv.Atoi(q.Get("page"))
	filters.Limit, _ = strconv.Atoi(q.Get("limit"))

	return filters, nil
}

func decodeCreateQuotationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req CreateQuotationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}
	return &req, nil
}

func decodeGetQuotationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return id, nil
}

func decodeUpdateQuotationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body UpdateQuotationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}
	return &updateQuotationEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeGetQuotationPricingRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return id, nil
}

func decodeAddQuotationChargeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body CreateQuotationChargeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}
	return &addChargeEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeUpdateQuotationChargeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	chargeIdStr := chi.URLParam(r, "chargeId")
	chargeId, err := strconv.ParseInt(chargeIdStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body UpdateQuotationChargeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}
	return &updateChargeEndpointRequest{QuotationID: id, ChargeID: chargeId, Body: &body}, nil
}

func decodeDeleteQuotationChargeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	chargeIdStr := chi.URLParam(r, "chargeId")
	chargeId, err := strconv.ParseInt(chargeIdStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &deleteChargeEndpointRequest{QuotationID: id, ChargeID: chargeId}, nil
}

func decodeReorderQuotationChargesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body ReorderQuotationChargesRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}
	return &reorderChargesEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeGetRateCandidatesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return id, nil
}

func decodeImportRateChargesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body ImportRateChargesRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}
	return &importRateEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeListTemplatesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	activeOnly := r.URL.Query().Get("active_only") != "false"
	return &listTemplatesEndpointRequest{ActiveOnly: activeOnly}, nil
}

func decodeGetTemplateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	templateIdStr := chi.URLParam(r, "templateId")
	id, err := strconv.ParseInt(templateIdStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return id, nil
}

func decodeCreateTemplateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req CreateQuotationTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}
	return &req, nil
}

func decodeUpdateTemplateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	templateIdStr := chi.URLParam(r, "templateId")
	id, err := strconv.ParseInt(templateIdStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body UpdateQuotationTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}
	return &updateTemplateEndpointRequest{TemplateID: id, Body: &body}, nil
}

func decodeDeleteTemplateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	templateIdStr := chi.URLParam(r, "templateId")
	id, err := strconv.ParseInt(templateIdStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return id, nil
}

func decodeApplyTemplateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body ApplyQuotationTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}
	return &applyTemplateEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeSaveAsTemplateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body CreateTemplateFromQuotationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}
	return &createTemplateFromQuoteEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeUpdateCommercialTermsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body UpdateQuotationCommercialTermsRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}
	return &updateCommercialTermsEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeSubmitReviewRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body SubmitQuotationForReviewRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	return &submitReviewEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeApproveQuotationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body ApproveQuotationRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	return &approveQuotationEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeRequestChangesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body RequestQuotationChangesRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}
	return &requestChangesEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeSendQuotationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body SendQuotationRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	return &sendQuotationEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeMarkViewedRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body MarkQuotationViewedRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	if body.IPAddress == "" {
		body.IPAddress = r.RemoteAddr
	}
	if body.UserAgent == "" {
		body.UserAgent = r.UserAgent()
	}
	return &markViewedEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeAcceptQuotationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body AcceptQuotationRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	return &acceptQuotationEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeDeclineQuotationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body DeclineQuotationRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	return &declineQuotationEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeCancelQuotationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body CancelQuotationRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	return &cancelQuotationEndpointRequest{QuotationID: id, Body: &body}, nil
}

// ── Document Decoders (Task 18.5) ───────────────────────────────────────────

func decodeGenerateDocumentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	docType := r.URL.Query().Get("document_type")
	if docType == "" {
		docType = "PDF"
	}
	return &generateDocEndpointRequest{QuotationID: id, DocumentType: docType}, nil
}

func decodeListDocumentsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &listDocsEndpointRequest{QuotationID: id}, nil
}

func decodeGetDocumentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	docIdStr := chi.URLParam(r, "docId")
	docId, err := strconv.ParseInt(docIdStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &getDocEndpointRequest{QuotationID: id, DocumentID: docId}, nil
}

// ── Public Link Decoders (Task 18.5) ───────────────────────────────────────

func decodeCreatePublicLinkRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body CreateQuotationPublicLinkRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	return &createPublicLinkEndpointRequest{QuotationID: id, Body: &body}, nil
}

func decodeListPublicLinksRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &listPublicLinksEndpointRequest{QuotationID: id}, nil
}

func decodeRevokePublicLinkRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	linkIdStr := chi.URLParam(r, "linkId")
	linkId, err := strconv.ParseInt(linkIdStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body RevokeQuotationPublicLinkRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	return &revokePublicLinkEndpointRequest{QuotationID: id, LinkID: linkId, Body: &body}, nil
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	return r.RemoteAddr
}

func decodePublicTokenRequest(_ context.Context, r *http.Request) (interface{}, error) {
	token := chi.URLParam(r, "token")
	if token == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &publicViewEndpointRequest{
		Token:     token,
		ClientIP:  getClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
	}, nil
}

func decodePublicAcceptRequest(_ context.Context, r *http.Request) (interface{}, error) {
	token := chi.URLParam(r, "token")
	if token == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body PublicAcceptQuotationRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		}
	}
	return &publicAcceptEndpointRequest{
		Token:     token,
		ClientIP:  getClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
		Body:      &body,
	}, nil
}

func decodePublicDeclineRequest(_ context.Context, r *http.Request) (interface{}, error) {
	token := chi.URLParam(r, "token")
	if token == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body PublicDeclineQuotationRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		}
	}
	return &publicDeclineEndpointRequest{
		Token:     token,
		ClientIP:  getClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
		Body:      &body,
	}, nil
}

// ── Encoders ─────────────────────────────────────────────────────────────────

func encodeQuotationResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func encodeQuotationDocumentResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	resp, ok := response.(*QuotationDocumentDownloadResponse)
	if !ok {
		return encodeQuotationResponse(context.Background(), w, response)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(&APIResponse{Success: true, Data: resp.Document})
}

func encodeQuotationDocumentDownload(_ context.Context, w http.ResponseWriter, response interface{}) error {
	resp, ok := response.(*QuotationDocumentDownloadResponse)
	if !ok {
		return encodeQuotationResponse(context.Background(), w, response)
	}
	contentType := "application/pdf"
	if strings.HasSuffix(resp.Document.FileName, ".html") {
		contentType = "text/html; charset=utf-8"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+resp.Document.FileName+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(resp.Content)))
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(resp.Content)
	return err
}

func encodeQuotationError(_ context.Context, err error, w http.ResponseWriter) {
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

func decodeConvertQuotationToBookingRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body ConvertQuotationToBookingRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}
	}
	return &convertQuotationEndpointRequest{
		QuotationID: id,
		Body:        &body,
	}, nil
}

func decodeConfirmHandoverRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var body ConfirmQuotationHandoverRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}
	}
	return &confirmHandoverEndpointRequest{
		QuotationID: id,
		Body:        &body,
	}, nil
}

func decodeEmptyAnalyticsRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeAnalyticsTrendsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if parsed, err := strconv.Atoi(daysStr); err == nil && parsed > 0 {
			days = parsed
		}
	}
	return days, nil
}

// ── Task 19.5: Rate-to-Quotation Integration Decoders ────────────────────────

func decodeSelectQuotationRateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req SelectQuotationRateRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}
	}
	req.QuotationID = id
	return &req, nil
}

func decodeReplaceQuotationRateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req ReplaceQuotationRateRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}
	}
	req.QuotationID = id
	return &req, nil
}

func decodeResolveQuotationRateRiskRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	qID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || qID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	riskIDStr := chi.URLParam(r, "riskId")
	riskID, err := strconv.ParseInt(riskIDStr, 10, 64)
	if err != nil || riskID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	user := "commercial_user"
	if uc, ok := middleware.GetUserContext(r.Context()); ok && uc.CognitoID != "" {
		user = uc.CognitoID
	}

	return &resolveRiskEndpointRequest{
		QuotationID: qID,
		RiskID:      riskID,
		User:        user,
	}, nil
}

func decodeCommercialImpactRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	qID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || qID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	req := &commercialImpactEndpointRequest{
		QuotationID: qID,
	}

	if rRate := r.URL.Query().Get("replacement_rate_id"); rRate != "" {
		if id, err := strconv.ParseInt(rRate, 10, 64); err == nil && id > 0 {
			req.ReplacementRateID = &id
		}
	}

	if rSpot := r.URL.Query().Get("replacement_spot_id"); rSpot != "" {
		if id, err := strconv.ParseInt(rSpot, 10, 64); err == nil && id > 0 {
			req.ReplacementSpotID = &id
		}
	}

	return req, nil
}






