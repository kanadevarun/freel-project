package contracts

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/freel/backend/internal/svcerror"
	"github.com/go-chi/chi/v5"
	kitHttp "github.com/go-kit/kit/transport/http"
)

// AddContractsHandlers mounts all Commercial Contracts endpoints onto the chi router
func AddContractsHandlers(
	router chi.Router,
	endpoints Endpoints,
	authMiddleware func(http.Handler) http.Handler,
) {
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	router.With(authMiddleware).Get("/overview", kitHttp.NewServer(
		endpoints.GetContractOverviewEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/", kitHttp.NewServer(
		endpoints.CreateContractEP,
		decodeCreateContractRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/", kitHttp.NewServer(
		endpoints.ListContractsEP,
		decodeListContractsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}", kitHttp.NewServer(
		endpoints.GetContractEP,
		decodeGetContractRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// ─── Contract Document Import & AI Creation Endpoints ──
	router.With(authMiddleware).Post("/import/upload", kitHttp.NewServer(
		endpoints.ImportContractDocumentEP,
		decodeImportContractDocumentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/import/{docId}/draft", kitHttp.NewServer(
		endpoints.GetExtractedContractDraftEP,
		decodeGetExtractedContractDraftRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/import/confirm", kitHttp.NewServer(
		endpoints.ConfirmContractImportEP,
		decodeConfirmContractImportRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Put("/{id}", kitHttp.NewServer(
		endpoints.UpdateContractEP,
		decodeUpdateContractRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/archive", kitHttp.NewServer(
		endpoints.ArchiveContractEP,
		decodeGetContractRequest, // Same parameter (just ID)
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Link Endpoints
	router.With(authMiddleware).Post("/{id}/links", kitHttp.NewServer(
		endpoints.AddLinkEP,
		decodeAddLinkRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Delete("/{id}/links/{linkId}", kitHttp.NewServer(
		endpoints.RemoveLinkEP,
		decodeRemoveLinkRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/relationship-summary", kitHttp.NewServer(
		endpoints.GetRelationshipSummaryEP,
		decodeGetContractRequest, // ID only
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/link-history", kitHttp.NewServer(
		endpoints.GetLinkHistoryEP,
		decodeGetContractRequest, // ID only
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/lifecycle", kitHttp.NewServer(
		endpoints.UpdateContractLifecycleEP,
		decodeUpdateContractLifecycleRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/lifecycle", kitHttp.NewServer(
		endpoints.GetContractLifecycleEventsEP,
		decodeGetContractRequest, // Same parameter (just ID)
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// ─── Task 20.3 & 20.6: Lifecycle Intelligence & Compliance Overview ──
	router.With(authMiddleware).Get("/lifecycle/summary", kitHttp.NewServer(
		endpoints.GetLifecycleSummaryEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/lifecycle/attention", kitHttp.NewServer(
		endpoints.GetAttentionItemsEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/lifecycle/events", kitHttp.NewServer(
		endpoints.GetLifecycleEventsEP,
		decodeGetLifecycleEventsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/lifecycle/evaluate", kitHttp.NewServer(
		endpoints.EvaluateLifecycleEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/compliance/summary", kitHttp.NewServer(
		endpoints.GetComplianceSummaryEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/compliance/attention", kitHttp.NewServer(
		endpoints.GetOpenComplianceAttentionEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/compliance/evaluate", kitHttp.NewServer(
		endpoints.EvaluateComplianceEP,
		decodeEmptyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/lifecycle-intelligence", kitHttp.NewServer(
		endpoints.GetContractLifecycleIntelligenceEP,
		decodeGetContractRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/risks", kitHttp.NewServer(
		endpoints.GetContractRisksEP,
		decodeGetContractRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/risks/{riskId}/resolve", kitHttp.NewServer(
		endpoints.ResolveRiskEP,
		decodeResolveRiskRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/renewal", kitHttp.NewServer(
		endpoints.GetRenewalTrackingEP,
		decodeGetContractRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/renewal/start", kitHttp.NewServer(
		endpoints.StartRenewalEP,
		decodeStartRenewalRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Put("/{id}/renewal", kitHttp.NewServer(
		endpoints.UpdateRenewalTrackingEP,
		decodeUpdateRenewalRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/commercial-impact", kitHttp.NewServer(
		endpoints.GetCommercialImpactEP,
		decodeGetContractRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// ─── Task 20.4: Versioning, Amendments & Approvals Routes ────────────────
	// Version Routes
	router.With(authMiddleware).Post("/{id}/versions", kitHttp.NewServer(
		endpoints.CreateContractVersionEP,
		decodeCreateVersionRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/versions", kitHttp.NewServer(
		endpoints.GetContractVersionsEP,
		decodeGetContractRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/versions/{versionId}", kitHttp.NewServer(
		endpoints.GetContractVersionEP,
		decodeGetVersionRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/versions/{versionId}/compare/{compareVersionId}", kitHttp.NewServer(
		endpoints.GetContractVersionComparisonEP,
		decodeCompareVersionsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/versions/{versionId}/make-effective", kitHttp.NewServer(
		endpoints.MakeVersionEffectiveEP,
		decodeGetVersionRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Amendment Routes
	router.With(authMiddleware).Post("/{id}/amendments", kitHttp.NewServer(
		endpoints.CreateContractAmendmentEP,
		decodeCreateAmendmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/amendments", kitHttp.NewServer(
		endpoints.GetContractAmendmentsEP,
		decodeGetContractRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/amendments/{amendmentId}", kitHttp.NewServer(
		endpoints.GetContractAmendmentEP,
		decodeGetAmendmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Put("/{id}/amendments/{amendmentId}", kitHttp.NewServer(
		endpoints.UpdateContractAmendmentEP,
		decodeUpdateAmendmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/amendments/{amendmentId}/submit", kitHttp.NewServer(
		endpoints.SubmitContractAmendmentEP,
		decodeSubmitAmendmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/amendments/{amendmentId}/implement", kitHttp.NewServer(
		endpoints.ImplementContractAmendmentEP,
		decodeGetAmendmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/amendments/{amendmentId}/cancel", kitHttp.NewServer(
		endpoints.CancelContractAmendmentEP,
		decodeGetAmendmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/amendments/{amendmentId}/changes", kitHttp.NewServer(
		endpoints.GetAmendmentChangesEP,
		decodeGetAmendmentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Approval Routes
	router.With(authMiddleware).Get("/{id}/approvals", kitHttp.NewServer(
		endpoints.GetContractApprovalsEP,
		decodeGetContractRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/approvals/{approvalId}/approve", kitHttp.NewServer(
		endpoints.ApproveContractChangeEP,
		decodeApproveContractRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/approvals/{approvalId}/reject", kitHttp.NewServer(
		endpoints.RejectContractChangeEP,
		decodeRejectContractRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/approvals/{approvalId}/cancel", kitHttp.NewServer(
		endpoints.CancelContractApprovalEP,
		decodeCancelApprovalRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// ─── Tasks 20.5 & 20.6: Documents, Terms, Obligations, Compliance, Performance ──
	// Documents
	router.With(authMiddleware).Get("/{id}/documents", kitHttp.NewServer(
		endpoints.ListDocumentsEP,
		decodeContractIDOnlyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/documents", kitHttp.NewServer(
		endpoints.CreateDocumentEP,
		decodeCreateDocumentRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/documents/{docId}/supersede", kitHttp.NewServer(
		endpoints.SupersedeDocumentEP,
		decodeSupersedeDocRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Terms
	router.With(authMiddleware).Get("/{id}/terms", kitHttp.NewServer(
		endpoints.ListTermsEP,
		decodeContractIDOnlyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/terms", kitHttp.NewServer(
		endpoints.CreateTermEP,
		decodeCreateTermRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Put("/{id}/terms/{termId}", kitHttp.NewServer(
		endpoints.UpdateTermEP,
		decodeUpdateTermRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Delete("/{id}/terms/{termId}", kitHttp.NewServer(
		endpoints.DeleteTermEP,
		decodeDeleteTermRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Obligations
	router.With(authMiddleware).Get("/{id}/obligations", kitHttp.NewServer(
		endpoints.ListObligationsEP,
		decodeContractIDOnlyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/obligations", kitHttp.NewServer(
		endpoints.CreateObligationEP,
		decodeCreateObligationRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Put("/{id}/obligations/{obligationId}", kitHttp.NewServer(
		endpoints.UpdateObligationEP,
		decodeUpdateObligationRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/obligations/{obligationId}/fulfill", kitHttp.NewServer(
		endpoints.FulfillObligationEP,
		decodeFulfillObligationRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/obligations/{obligationId}/waive", kitHttp.NewServer(
		endpoints.WaiveObligationEP,
		decodeWaiveObligationRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Compliance
	router.With(authMiddleware).Get("/{id}/compliance/events", kitHttp.NewServer(
		endpoints.ListComplianceEventsEP,
		decodeContractIDOnlyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/compliance/events/{eventId}/resolve", kitHttp.NewServer(
		endpoints.ResolveComplianceEventEP,
		decodeResolveComplianceEventRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/compliance/requirements", kitHttp.NewServer(
		endpoints.ListComplianceRequirementsEP,
		decodeContractIDOnlyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/compliance/requirements", kitHttp.NewServer(
		endpoints.CreateComplianceRequirementEP,
		decodeCreateComplianceReqRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/{id}/compliance/requirements/{reqId}/verify", kitHttp.NewServer(
		endpoints.VerifyComplianceRequirementEP,
		decodeVerifyComplianceReqRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	// Performance & Operational Intelligence
	router.With(authMiddleware).Get("/{id}/performance", kitHttp.NewServer(
		endpoints.GetContractPerformanceEP,
		decodeContractIDOnlyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id}/operational-intelligence", kitHttp.NewServer(
		endpoints.GetContractOperationalIntelligenceEP,
		decodeContractIDOnlyRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)
}

func decodeEmptyRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return struct{}{}, nil
}

func decodeCreateContractRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req CreateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

func decodeListContractsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	partyID, _ := strconv.ParseInt(q.Get("party_id"), 10, 64)

	return &ListContractsRequest{
		Page:         page,
		Limit:        limit,
		Search:       q.Get("search"),
		Status:       q.Get("status"),
		PartyID:      partyID,
		ContractType: q.Get("contract_type"),
	}, nil
}

func decodeGetContractRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return id, nil
}

func decodeAddLinkRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	var req addLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req.CreateContractLinkRequest); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ContractID = id
	return req, nil
}

func decodeRemoveLinkRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	linkID, err := strconv.ParseInt(chi.URLParam(r, "linkId"), 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return removeLinkRequest{ContractID: id, LinkID: linkID}, nil
}

func decodeUpdateContractRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req UpdateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	
	return updateContractRequestWrapper{
		ContractID: id,
		Req:        &req,
	}, nil
}

func decodeUpdateContractLifecycleRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req UpdateContractLifecycleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	
	return updateLifecycleRequestWrapper{
		ContractID: id,
		Req:        &req,
	}, nil
}

func decodeGetLifecycleEventsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var contractID *int64
	if cidStr := r.URL.Query().Get("contract_id"); cidStr != "" {
		if cid, err := strconv.ParseInt(cidStr, 10, 64); err == nil {
			contractID = &cid
		}
	}
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	return getLifecycleEventsRequest{
		ContractID: contractID,
		Limit:      limit,
	}, nil
}

func decodeResolveRiskRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	riskIdStr := chi.URLParam(r, "riskId")
	riskID, err := strconv.ParseInt(riskIdStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req ResolveRiskRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	return resolveRiskRequestWrapper{
		ContractID: id,
		RiskID:     riskID,
		Payload:    &req,
	}, nil
}

func decodeStartRenewalRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req StartRenewalRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	return startRenewalRequestWrapper{
		ContractID: id,
		Payload:    &req,
	}, nil
}

func decodeUpdateRenewalRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req UpdateRenewalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return updateRenewalRequestWrapper{
		ContractID: id,
		Payload:    &req,
	}, nil
}

// ── Task 20.4 Request Decoders ──────────────────────────────────────────────

func decodeCreateVersionRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req CreateContractVersionRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	return createVersionRequestWrapper{
		ContractID: id,
		Payload:    &req,
	}, nil
}

func decodeGetVersionRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	versionID := chi.URLParam(r, "versionId")
	if versionID == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return getVersionRequestWrapper{
		ContractID: id,
		VersionID:  versionID,
	}, nil
}

func decodeCompareVersionsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	baseVersionID := chi.URLParam(r, "versionId")
	targetVersionID := chi.URLParam(r, "compareVersionId")
	if baseVersionID == "" || targetVersionID == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return compareVersionsRequestWrapper{
		ContractID:      id,
		BaseVersionID:   baseVersionID,
		TargetVersionID: targetVersionID,
	}, nil
}

func decodeCreateAmendmentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req CreateContractAmendmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return createAmendmentRequestWrapper{
		ContractID: id,
		Payload:    &req,
	}, nil
}

func decodeGetAmendmentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	amendmentID := chi.URLParam(r, "amendmentId")
	if amendmentID == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return getAmendmentRequestWrapper{
		ContractID:  id,
		AmendmentID: amendmentID,
	}, nil
}

func decodeUpdateAmendmentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	amendmentID := chi.URLParam(r, "amendmentId")
	if amendmentID == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req UpdateContractAmendmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return updateAmendmentRequestWrapper{
		ContractID:  id,
		AmendmentID: amendmentID,
		Payload:     &req,
	}, nil
}

func decodeSubmitAmendmentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	amendmentID := chi.URLParam(r, "amendmentId")
	if amendmentID == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req SubmitContractAmendmentRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	return submitAmendmentRequestWrapper{
		ContractID:  id,
		AmendmentID: amendmentID,
		Payload:     &req,
	}, nil
}

func decodeApproveContractRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	approvalID := chi.URLParam(r, "approvalId")
	if approvalID == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req ApproveContractRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	return approveRequestWrapper{
		ContractID: id,
		ApprovalID: approvalID,
		Payload:    &req,
	}, nil
}

func decodeRejectContractRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	approvalID := chi.URLParam(r, "approvalId")
	if approvalID == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req RejectContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return rejectRequestWrapper{
		ContractID: id,
		ApprovalID: approvalID,
		Payload:    &req,
	}, nil
}

func decodeCancelApprovalRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	approvalID := chi.URLParam(r, "approvalId")
	if approvalID == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req CancelApprovalRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	return cancelApprovalWrapper{
		ContractID: id,
		ApprovalID: approvalID,
		Payload:    &req,
	}, nil
}

func decodeContractIDOnlyRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return contractIDOnlyWrapper{ContractID: id}, nil
}

func decodeCreateDocumentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req CreateAgreementDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return createDocumentWrapper{
		ContractID: id,
		Payload:    &req,
	}, nil
}

func decodeSupersedeDocRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	docID := chi.URLParam(r, "docId")
	var body struct {
		NewDocID string `json:"new_doc_id"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	return supersedeDocWrapper{
		ContractID: id,
		DocID:      docID,
		NewDocID:   body.NewDocID,
	}, nil
}

func decodeCreateTermRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req CreateContractTermRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return createTermWrapper{
		ContractID: id,
		Payload:    &req,
	}, nil
}

func decodeUpdateTermRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	termIDStr := chi.URLParam(r, "termId")
	termID, err := strconv.ParseInt(termIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req UpdateContractTermRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return updateTermWrapper{
		ContractID: id,
		TermID:     termID,
		Payload:    &req,
	}, nil
}

func decodeDeleteTermRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	termIDStr := chi.URLParam(r, "termId")
	termID, err := strconv.ParseInt(termIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return deleteTermWrapper{
		ContractID: id,
		TermID:     termID,
	}, nil
}

func decodeCreateObligationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req CreateContractObligationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return createObligationWrapper{
		ContractID: id,
		Payload:    &req,
	}, nil
}

func decodeUpdateObligationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	obIDStr := chi.URLParam(r, "obligationId")
	obID, err := strconv.ParseInt(obIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req UpdateContractObligationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return updateObligationWrapper{
		ContractID:   id,
		ObligationID: obID,
		Payload:      &req,
	}, nil
}

func decodeFulfillObligationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	obIDStr := chi.URLParam(r, "obligationId")
	obID, err := strconv.ParseInt(obIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req FulfillObligationRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	return fulfillObligationWrapper{
		ContractID:   id,
		ObligationID: obID,
		Payload:      &req,
	}, nil
}

func decodeWaiveObligationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	obIDStr := chi.URLParam(r, "obligationId")
	obID, err := strconv.ParseInt(obIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req WaiveObligationRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	return waiveObligationWrapper{
		ContractID:   id,
		ObligationID: obID,
		Payload:      &req,
	}, nil
}

func decodeResolveComplianceEventRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	eventIDStr := chi.URLParam(r, "eventId")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req ResolveComplianceEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return resolveComplianceEventWrapper{
		ContractID: id,
		EventID:    eventID,
		Payload:    &req,
	}, nil
}

func decodeCreateComplianceReqRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req CreateComplianceRequirementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return createComplianceReqWrapper{
		ContractID: id,
		Payload:    &req,
	}, nil
}

func decodeVerifyComplianceReqRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	reqIDStr := chi.URLParam(r, "reqId")
	reqID, err := strconv.ParseInt(reqIDStr, 10, 64)
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req VerifyComplianceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	return verifyComplianceReqWrapper{
		ContractID:    id,
		RequirementID: reqID,
		Payload:       &req,
	}, nil
}

type errorer interface {
	error() error
}

func encodeAPIResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeErrorResponse(ctx, e.error(), w)
		return nil
	}
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
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

	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "BAD_REQUEST",
			"message": err.Error(),
		},
	})
}

func decodeImportContractDocumentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		e := svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		e.Message = "File too large (max 25MB)"
		return nil, e
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		e := svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		e.Message = "Missing file payload in form"
		return nil, e
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var hint *string
	if h := r.FormValue("contract_type_hint"); h != "" {
		hint = &h
	}

	return importDocumentRequest{
		FileName:         header.Filename,
		FileBytes:        fileBytes,
		ContractTypeHint: hint,
	}, nil
}

func decodeGetExtractedContractDraftRequest(_ context.Context, r *http.Request) (interface{}, error) {
	docID := chi.URLParam(r, "docId")
	if docID == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return docID, nil
}

func decodeConfirmContractImportRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req ConfirmContractImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

