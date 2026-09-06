package quotations

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-kit/kit/endpoint"
)

// Endpoints groups all Go-Kit endpoint functions for the quotation domain.
type Endpoints struct {
	CreateQuotationEP         endpoint.Endpoint
	UpdateQuotationEP         endpoint.Endpoint
	GetQuotationEP            endpoint.Endpoint
	ListQuotationsEP          endpoint.Endpoint
	GetQuotationSummaryEP     endpoint.Endpoint

	// Pricing & Charges Endpoints (Task 18.2)
	GetQuotationPricingEP     endpoint.Endpoint
	AddQuotationChargeEP      endpoint.Endpoint
	UpdateQuotationChargeEP   endpoint.Endpoint
	DeleteQuotationChargeEP   endpoint.Endpoint
	ReorderQuotationChargesEP endpoint.Endpoint
	GetRateCandidatesEP       endpoint.Endpoint
	ImportRateChargesEP       endpoint.Endpoint

	// Templates & Commercial Terms Endpoints (Task 18.3)
	ListQuotationTemplatesEP        endpoint.Endpoint
	GetQuotationTemplateEP          endpoint.Endpoint
	CreateQuotationTemplateEP       endpoint.Endpoint
	UpdateQuotationTemplateEP       endpoint.Endpoint
	DeleteQuotationTemplateEP       endpoint.Endpoint
	CreateTemplateFromQuotationEP   endpoint.Endpoint
	ApplyTemplateToQuotationEP      endpoint.Endpoint
	UpdateQuotationCommercialTermsEP endpoint.Endpoint

	// Lifecycle & Approval Endpoints (Task 18.4)
	SubmitQuotationForReviewEP    endpoint.Endpoint
	ApproveQuotationEP            endpoint.Endpoint
	RequestQuotationChangesEP     endpoint.Endpoint
	SendQuotationEP               endpoint.Endpoint
	GetCustomerQuotationPreviewEP endpoint.Endpoint
	GetQuotationApprovalHistoryEP endpoint.Endpoint
	GetQuotationApprovalStatusEP  endpoint.Endpoint
	MarkQuotationViewedEP         endpoint.Endpoint
	AcceptQuotationEP             endpoint.Endpoint
	DeclineQuotationEP            endpoint.Endpoint
	CancelQuotationEP             endpoint.Endpoint

	// Documents & Public Sharing Endpoints (Task 18.5)
	GenerateQuotationDocumentEP endpoint.Endpoint
	ListQuotationDocumentsEP    endpoint.Endpoint
	GetQuotationDocumentEP      endpoint.Endpoint
	CreateQuotationPublicLinkEP endpoint.Endpoint
	ListQuotationPublicLinksEP  endpoint.Endpoint
	RevokeQuotationPublicLinkEP endpoint.Endpoint
	GetPublicQuotationByTokenEP endpoint.Endpoint
	PublicAcceptQuotationEP     endpoint.Endpoint
	PublicDeclineQuotationEP    endpoint.Endpoint

	// Conversion & Commercial Handover Endpoints (Task 18.6)
	GetQuotationConversionPreviewEP endpoint.Endpoint
	ConvertQuotationToBookingEP     endpoint.Endpoint
	GetQuotationConversionHistoryEP endpoint.Endpoint

	// Booking Confirmation & Handover Traceability Endpoints (Task 18.7)
	GetQuotationOperationalHandoverEP endpoint.Endpoint
	ConfirmQuotationBookingHandoverEP endpoint.Endpoint
	GetQuotationOperationalChangesEP  endpoint.Endpoint
	GetQuotationHandoverHistoryEP     endpoint.Endpoint

	// Quotation Analytics & Intelligence Endpoints (Task 18.8)
	GetQuotationAnalyticsOverviewEP   endpoint.Endpoint
	GetQuotationAnalyticsTrendsEP     endpoint.Endpoint
	GetCustomerQuotationPerformanceEP endpoint.Endpoint
	GetQuotationPerformanceByModeEP   endpoint.Endpoint
	GetQuotationExpiryRiskEP          endpoint.Endpoint

	// Rate-to-Quotation Integration Endpoints (Task 19.5)
	GetQuotationRateCandidatesEP       endpoint.Endpoint
	GetQuotationRateSelectionEP        endpoint.Endpoint
	SelectQuotationRateEP              endpoint.Endpoint
	ReplaceQuotationRateEP             endpoint.Endpoint
	RemoveQuotationRateEP              endpoint.Endpoint
	GetQuotationRateSnapshotEP         endpoint.Endpoint
	GetQuotationRateSelectionHistoryEP endpoint.Endpoint

	// Rate Lifecycle Intelligence & Commercial Risk Endpoints (Task 19.6)
	GetQuotationRateRisksEP          endpoint.Endpoint
	ResolveQuotationRateRiskEP       endpoint.Endpoint
	GetRateReplacementCandidatesEP   endpoint.Endpoint
	GetCommercialImpactAnalysisEP    endpoint.Endpoint
	EvaluateQuotationRateRisksEP     endpoint.Endpoint
}

// NewAllQuotationEndpoints wires up all endpoints against the service.
func NewAllQuotationEndpoints(svc Service) Endpoints {
	return Endpoints{
		CreateQuotationEP:         makeCreateQuotationEP(svc),
		UpdateQuotationEP:         makeUpdateQuotationEP(svc),
		GetQuotationEP:            makeGetQuotationEP(svc),
		ListQuotationsEP:          makeListQuotationsEP(svc),
		GetQuotationSummaryEP:     makeGetQuotationSummaryEP(svc),

		GetQuotationPricingEP:     makeGetQuotationPricingEP(svc),
		AddQuotationChargeEP:      makeAddQuotationChargeEP(svc),
		UpdateQuotationChargeEP:   makeUpdateQuotationChargeEP(svc),
		DeleteQuotationChargeEP:   makeDeleteQuotationChargeEP(svc),
		ReorderQuotationChargesEP: makeReorderQuotationChargesEP(svc),
		GetRateCandidatesEP:       makeGetRateCandidatesEP(svc),
		ImportRateChargesEP:       makeImportRateChargesEP(svc),

		ListQuotationTemplatesEP:        makeListQuotationTemplatesEP(svc),
		GetQuotationTemplateEP:          makeGetQuotationTemplateEP(svc),
		CreateQuotationTemplateEP:       makeCreateQuotationTemplateEP(svc),
		UpdateQuotationTemplateEP:       makeUpdateQuotationTemplateEP(svc),
		DeleteQuotationTemplateEP:       makeDeleteQuotationTemplateEP(svc),
		CreateTemplateFromQuotationEP:   makeCreateTemplateFromQuotationEP(svc),
		ApplyTemplateToQuotationEP:      makeApplyTemplateToQuotationEP(svc),
		UpdateQuotationCommercialTermsEP: makeUpdateQuotationCommercialTermsEP(svc),

		SubmitQuotationForReviewEP:    makeSubmitQuotationForReviewEP(svc),
		ApproveQuotationEP:            makeApproveQuotationEP(svc),
		RequestQuotationChangesEP:     makeRequestQuotationChangesEP(svc),
		SendQuotationEP:               makeSendQuotationEP(svc),
		GetCustomerQuotationPreviewEP: makeGetCustomerQuotationPreviewEP(svc),
		GetQuotationApprovalHistoryEP: makeGetQuotationApprovalHistoryEP(svc),
		GetQuotationApprovalStatusEP:  makeGetQuotationApprovalStatusEP(svc),
		MarkQuotationViewedEP:         makeMarkQuotationViewedEP(svc),
		AcceptQuotationEP:             makeAcceptQuotationEP(svc),
		DeclineQuotationEP:            makeDeclineQuotationEP(svc),
		CancelQuotationEP:             makeCancelQuotationEP(svc),

		GenerateQuotationDocumentEP: makeGenerateQuotationDocumentEP(svc),
		ListQuotationDocumentsEP:    makeListQuotationDocumentsEP(svc),
		GetQuotationDocumentEP:      makeGetQuotationDocumentEP(svc),
		CreateQuotationPublicLinkEP: makeCreateQuotationPublicLinkEP(svc),
		ListQuotationPublicLinksEP:  makeListQuotationPublicLinksEP(svc),
		RevokeQuotationPublicLinkEP: makeRevokeQuotationPublicLinkEP(svc),
		GetPublicQuotationByTokenEP: makeGetPublicQuotationByTokenEP(svc),
		PublicAcceptQuotationEP:     makePublicAcceptQuotationEP(svc),
		PublicDeclineQuotationEP:    makePublicDeclineQuotationEP(svc),

		GetQuotationConversionPreviewEP: makeGetQuotationConversionPreviewEP(svc),
		ConvertQuotationToBookingEP:     makeConvertQuotationToBookingEP(svc),
		GetQuotationConversionHistoryEP: makeGetQuotationConversionHistoryEP(svc),

		GetQuotationOperationalHandoverEP: makeGetQuotationOperationalHandoverEP(svc),
		ConfirmQuotationBookingHandoverEP: makeConfirmQuotationBookingHandoverEP(svc),
		GetQuotationOperationalChangesEP:  makeGetQuotationOperationalChangesEP(svc),
		GetQuotationHandoverHistoryEP:     makeGetQuotationHandoverHistoryEP(svc),

		GetQuotationAnalyticsOverviewEP:   makeGetQuotationAnalyticsOverviewEP(svc),
		GetQuotationAnalyticsTrendsEP:     makeGetQuotationAnalyticsTrendsEP(svc),
		GetCustomerQuotationPerformanceEP: makeGetCustomerQuotationPerformanceEP(svc),
		GetQuotationPerformanceByModeEP:   makeGetQuotationPerformanceByModeEP(svc),
		GetQuotationExpiryRiskEP:          makeGetQuotationExpiryRiskEP(svc),

		GetQuotationRateCandidatesEP:       makeGetQuotationRateCandidatesEP(svc),
		GetQuotationRateSelectionEP:        makeGetQuotationRateSelectionEP(svc),
		SelectQuotationRateEP:              makeSelectQuotationRateEP(svc),
		ReplaceQuotationRateEP:             makeReplaceQuotationRateEP(svc),
		RemoveQuotationRateEP:              makeRemoveQuotationRateEP(svc),
		GetQuotationRateSnapshotEP:         makeGetQuotationRateSnapshotEP(svc),
		GetQuotationRateSelectionHistoryEP: makeGetQuotationRateSelectionHistoryEP(svc),

		// Task 19.6
		GetQuotationRateRisksEP:          makeGetQuotationRateRisksEP(svc),
		ResolveQuotationRateRiskEP:       makeResolveQuotationRateRiskEP(svc),
		GetRateReplacementCandidatesEP:   makeGetRateReplacementCandidatesEP(svc),
		GetCommercialImpactAnalysisEP:    makeGetCommercialImpactAnalysisEP(svc),
		EvaluateQuotationRateRisksEP:     makeEvaluateQuotationRateRisksEP(svc),
	}
}

func getOrgAndUserID(ctx context.Context) (orgID int64, userID int64, err error) {
	uc, ok := middleware.GetUserContext(ctx)
	if !ok || uc.OrgID <= 0 {
		return 0, 0, svcerror.NewServiceError(svcerror.ErrInsufficientResourceAccess)
	}
	return uc.OrgID, uc.UserID, nil
}

func makeCreateQuotationEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*CreateQuotationRequest)
		q, err := svc.CreateQuotation(ctx, orgID, userID, req)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotation created successfully", Data: q}, nil
	}
}

func makeUpdateQuotationEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*updateQuotationEndpointRequest)
		q, err := svc.UpdateQuotation(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotation updated successfully", Data: q}, nil
	}
}

type updateQuotationEndpointRequest struct {
	QuotationID int64
	Body        *UpdateQuotationRequest
}

func makeGetQuotationEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		quotationID := request.(int64)
		detail, err := svc.GetQuotation(ctx, orgID, quotationID)
		if err != nil {
			log.Printf("[ERROR] GetQuotation failed for orgID=%d, id=%d: %v", orgID, quotationID, err)
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotation retrieved successfully", Data: detail}, nil
	}
}

func makeListQuotationsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		filters := request.(*QuotationListFilters)
		filters.OrgID = orgID

		resp, err := svc.ListQuotations(ctx, filters)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotations retrieved successfully", Data: resp}, nil
	}
}

func makeGetQuotationSummaryEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		summary, err := svc.GetQuotationSummary(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotation summary retrieved successfully", Data: summary}, nil
	}
}

// ── Pricing & Charges Endpoints (Task 18.2) ──────────────────────────────────

func makeGetQuotationPricingEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		quotationID := request.(int64)
		pricing, err := svc.GetQuotationPricing(ctx, orgID, quotationID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotation pricing retrieved successfully", Data: pricing}, nil
	}
}

type addChargeEndpointRequest struct {
	QuotationID int64
	Body        *CreateQuotationChargeRequest
}

func makeAddQuotationChargeEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*addChargeEndpointRequest)
		pricing, err := svc.AddQuotationCharge(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Charge added successfully", Data: pricing}, nil
	}
}

type updateChargeEndpointRequest struct {
	QuotationID int64
	ChargeID    int64
	Body        *UpdateQuotationChargeRequest
}

func makeUpdateQuotationChargeEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*updateChargeEndpointRequest)
		pricing, err := svc.UpdateQuotationCharge(ctx, orgID, req.QuotationID, req.ChargeID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Charge updated successfully", Data: pricing}, nil
	}
}

type deleteChargeEndpointRequest struct {
	QuotationID int64
	ChargeID    int64
}

func makeDeleteQuotationChargeEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*deleteChargeEndpointRequest)
		pricing, err := svc.DeleteQuotationCharge(ctx, orgID, req.QuotationID, req.ChargeID, userID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Charge deleted successfully", Data: pricing}, nil
	}
}

type reorderChargesEndpointRequest struct {
	QuotationID int64
	Body        *ReorderQuotationChargesRequest
}

func makeReorderQuotationChargesEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*reorderChargesEndpointRequest)
		pricing, err := svc.ReorderQuotationCharges(ctx, orgID, req.QuotationID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Charges reordered successfully", Data: pricing}, nil
	}
}

func makeGetRateCandidatesEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		quotationID := request.(int64)
		candidates, err := svc.GetRateCandidates(ctx, orgID, quotationID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Rate candidates retrieved successfully", Data: candidates}, nil
	}
}

type importRateEndpointRequest struct {
	QuotationID int64
	Body        *ImportRateChargesRequest
}

func makeImportRateChargesEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*importRateEndpointRequest)
		pricing, err := svc.ImportRateCharges(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Rate imported successfully", Data: pricing}, nil
	}
}

// ── Templates & Commercial Terms Endpoints (Task 18.3) ────────────────────────

type listTemplatesEndpointRequest struct {
	ActiveOnly bool
}

func makeListQuotationTemplatesEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*listTemplatesEndpointRequest)
		templates, err := svc.ListQuotationTemplates(ctx, orgID, req.ActiveOnly)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Templates retrieved successfully", Data: templates}, nil
	}
}

func makeGetQuotationTemplateEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		templateID := request.(int64)
		detail, err := svc.GetQuotationTemplate(ctx, orgID, templateID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Template retrieved successfully", Data: detail}, nil
	}
}

func makeCreateQuotationTemplateEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*CreateQuotationTemplateRequest)
		tmpl, err := svc.CreateQuotationTemplate(ctx, orgID, userID, req)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Template created successfully", Data: tmpl}, nil
	}
}

type updateTemplateEndpointRequest struct {
	TemplateID int64
	Body       *UpdateQuotationTemplateRequest
}

func makeUpdateQuotationTemplateEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*updateTemplateEndpointRequest)
		tmpl, err := svc.UpdateQuotationTemplate(ctx, orgID, req.TemplateID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Template updated successfully", Data: tmpl}, nil
	}
}

func makeDeleteQuotationTemplateEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		templateID := request.(int64)
		if err := svc.DeleteQuotationTemplate(ctx, orgID, templateID); err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Template archived successfully"}, nil
	}
}

type createTemplateFromQuoteEndpointRequest struct {
	QuotationID int64
	Body        *CreateTemplateFromQuotationRequest
}

func makeCreateTemplateFromQuotationEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*createTemplateFromQuoteEndpointRequest)
		tmpl, err := svc.CreateTemplateFromQuotation(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Saved as reusable template successfully", Data: tmpl}, nil
	}
}

type applyTemplateEndpointRequest struct {
	QuotationID int64
	Body        *ApplyQuotationTemplateRequest
}

func makeApplyTemplateToQuotationEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*applyTemplateEndpointRequest)
		pricing, err := svc.ApplyTemplateToQuotation(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Template applied successfully", Data: pricing}, nil
	}
}

type updateCommercialTermsEndpointRequest struct {
	QuotationID int64
	Body        *UpdateQuotationCommercialTermsRequest
}

func makeUpdateQuotationCommercialTermsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*updateCommercialTermsEndpointRequest)
		detail, err := svc.UpdateQuotationCommercialTerms(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Commercial terms updated successfully", Data: detail}, nil
	}
}

// ── Lifecycle & Approval Endpoints (Task 18.4) ───────────────────────────────

type submitReviewEndpointRequest struct {
	QuotationID int64
	Body        *SubmitQuotationForReviewRequest
}

func makeSubmitQuotationForReviewEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*submitReviewEndpointRequest)
		detail, err := svc.SubmitQuotationForReview(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotation submitted for review", Data: detail}, nil
	}
}

type approveQuotationEndpointRequest struct {
	QuotationID int64
	Body        *ApproveQuotationRequest
}

func makeApproveQuotationEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*approveQuotationEndpointRequest)
		detail, err := svc.ApproveQuotation(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotation approved successfully", Data: detail}, nil
	}
}

type requestChangesEndpointRequest struct {
	QuotationID int64
	Body        *RequestQuotationChangesRequest
}

func makeRequestQuotationChangesEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*requestChangesEndpointRequest)
		detail, err := svc.RequestQuotationChanges(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Changes requested on quotation", Data: detail}, nil
	}
}

type sendQuotationEndpointRequest struct {
	QuotationID int64
	Body        *SendQuotationRequest
}

func makeSendQuotationEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*sendQuotationEndpointRequest)
		detail, err := svc.SendQuotation(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotation marked as sent", Data: detail}, nil
	}
}

func makeGetCustomerQuotationPreviewEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		quotationID := request.(int64)
		preview, err := svc.GetCustomerQuotationPreview(ctx, orgID, quotationID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Customer quotation preview generated", Data: preview}, nil
	}
}

func makeGetQuotationApprovalHistoryEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		quotationID := request.(int64)
		history, err := svc.GetQuotationApprovalHistory(ctx, orgID, quotationID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Approval history retrieved successfully", Data: history}, nil
	}
}

func makeGetQuotationApprovalStatusEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		quotationID := request.(int64)
		status, err := svc.GetQuotationApprovalStatus(ctx, orgID, quotationID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Approval status retrieved successfully", Data: status}, nil
	}
}

type markViewedEndpointRequest struct {
	QuotationID int64
	Body        *MarkQuotationViewedRequest
}

func makeMarkQuotationViewedEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*markViewedEndpointRequest)
		if err := svc.MarkQuotationViewed(ctx, orgID, req.QuotationID, req.Body); err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "View recorded"}, nil
	}
}

type acceptQuotationEndpointRequest struct {
	QuotationID int64
	Body        *AcceptQuotationRequest
}

func makeAcceptQuotationEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*acceptQuotationEndpointRequest)
		detail, err := svc.AcceptQuotation(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotation accepted", Data: detail}, nil
	}
}

type declineQuotationEndpointRequest struct {
	QuotationID int64
	Body        *DeclineQuotationRequest
}

func makeDeclineQuotationEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*declineQuotationEndpointRequest)
		detail, err := svc.DeclineQuotation(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotation declined", Data: detail}, nil
	}
}

type cancelQuotationEndpointRequest struct {
	QuotationID int64
	Body        *CancelQuotationRequest
}

func makeCancelQuotationEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*cancelQuotationEndpointRequest)
		detail, err := svc.CancelQuotation(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotation cancelled", Data: detail}, nil
	}
}

// parseID is a helper to extract route param as int64.
func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// ── Document Endpoints (Task 18.5) ──────────────────────────────────────────

type generateDocEndpointRequest struct {
	QuotationID  int64
	DocumentType string
}

func makeGenerateQuotationDocumentEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*generateDocEndpointRequest)
		doc, err := svc.GenerateQuotationDocument(ctx, orgID, req.QuotationID, userID, req.DocumentType)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotation document generated successfully", Data: doc}, nil
	}
}

type listDocsEndpointRequest struct {
	QuotationID int64
}

func makeListQuotationDocumentsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*listDocsEndpointRequest)
		docs, err := svc.ListQuotationDocuments(ctx, orgID, req.QuotationID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Data: docs}, nil
	}
}

type getDocEndpointRequest struct {
	QuotationID int64
	DocumentID  int64
}

type QuotationDocumentDownloadResponse struct {
	Document *QuotationDocument
	Content  []byte
}

func makeGetQuotationDocumentEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*getDocEndpointRequest)
		doc, content, err := svc.GetQuotationDocument(ctx, orgID, req.QuotationID, req.DocumentID)
		if err != nil {
			return nil, err
		}
		return &QuotationDocumentDownloadResponse{
			Document: doc,
			Content:  content,
		}, nil
	}
}

// ── Public Sharing Endpoints (Task 18.5) ────────────────────────────────────

type createPublicLinkEndpointRequest struct {
	QuotationID int64
	Body        *CreateQuotationPublicLinkRequest
}

func makeCreateQuotationPublicLinkEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*createPublicLinkEndpointRequest)
		link, err := svc.CreateQuotationPublicLink(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Public sharing link created", Data: link}, nil
	}
}

type listPublicLinksEndpointRequest struct {
	QuotationID int64
}

func makeListQuotationPublicLinksEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*listPublicLinksEndpointRequest)
		links, err := svc.ListQuotationPublicLinks(ctx, orgID, req.QuotationID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Data: links}, nil
	}
}

type revokePublicLinkEndpointRequest struct {
	QuotationID int64
	LinkID      int64
	Body        *RevokeQuotationPublicLinkRequest
}

func makeRevokeQuotationPublicLinkEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req := request.(*revokePublicLinkEndpointRequest)
		if err := svc.RevokeQuotationPublicLink(ctx, orgID, req.QuotationID, req.LinkID, userID, req.Body); err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Public sharing link revoked successfully"}, nil
	}
}

type publicViewEndpointRequest struct {
	Token     string
	ClientIP  string
	UserAgent string
}

func makeGetPublicQuotationByTokenEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*publicViewEndpointRequest)
		resp, err := svc.GetPublicQuotationByToken(ctx, req.Token, req.ClientIP, req.UserAgent)
		if err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Data: resp}, nil
	}
}

type publicAcceptEndpointRequest struct {
	Token     string
	ClientIP  string
	UserAgent string
	Body      *PublicAcceptQuotationRequest
}

func makePublicAcceptQuotationEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*publicAcceptEndpointRequest)
		if err := svc.PublicAcceptQuotation(ctx, req.Token, req.ClientIP, req.UserAgent, req.Body); err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotation accepted successfully"}, nil
	}
}

type publicDeclineEndpointRequest struct {
	Token     string
	ClientIP  string
	UserAgent string
	Body      *PublicDeclineQuotationRequest
}

func makePublicDeclineQuotationEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*publicDeclineEndpointRequest)
		if err := svc.PublicDeclineQuotation(ctx, req.Token, req.ClientIP, req.UserAgent, req.Body); err != nil {
			return nil, err
		}
		return &APIResponse{Success: true, Message: "Quotation declined successfully"}, nil
	}
}

// ─── Quotation-to-Booking Conversion Endpoint Handlers (Task 18.6) ────────────

func makeGetQuotationConversionPreviewEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		quotationID, ok := request.(int64)
		if !ok {
			return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		}
		preview, err := svc.GetQuotationConversionPreview(ctx, orgID, quotationID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Quotation conversion preview retrieved successfully",
			Data:    preview,
		}, nil
	}
}

type convertQuotationEndpointRequest struct {
	QuotationID int64
	Body        *ConvertQuotationToBookingRequest
}

func makeConvertQuotationToBookingEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req, ok := request.(*convertQuotationEndpointRequest)
		if !ok {
			return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		}
		result, err := svc.ConvertQuotationToBooking(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: result.Message,
			Data:    result,
		}, nil
	}
}

func makeGetQuotationConversionHistoryEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		quotationID, ok := request.(int64)
		if !ok {
			return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		}
		history, err := svc.GetQuotationConversionHistory(ctx, orgID, quotationID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Quotation conversion history retrieved successfully",
			Data:    history,
		}, nil
	}
}

// ── Booking Confirmation & Handover Traceability Endpoints (Task 18.7) ────────

func makeGetQuotationOperationalHandoverEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		quotationID, ok := request.(int64)
		if !ok {
			return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		}
		handover, err := svc.GetQuotationOperationalHandover(ctx, orgID, quotationID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Quotation operational handover details retrieved successfully",
			Data:    handover,
		}, nil
	}
}

type confirmHandoverEndpointRequest struct {
	QuotationID int64
	Body        *ConfirmQuotationHandoverRequest
}

func makeConfirmQuotationBookingHandoverEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, userID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req, ok := request.(*confirmHandoverEndpointRequest)
		if !ok {
			return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		}
		handover, err := svc.ConfirmQuotationBookingHandover(ctx, orgID, req.QuotationID, userID, req.Body)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Commercial handover confirmed successfully",
			Data:    handover,
		}, nil
	}
}

func makeGetQuotationOperationalChangesEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		quotationID, ok := request.(int64)
		if !ok {
			return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		}
		changes, err := svc.GetQuotationOperationalChanges(ctx, orgID, quotationID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Operational changes retrieved successfully",
			Data:    changes,
		}, nil
	}
}

func makeGetQuotationHandoverHistoryEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		quotationID, ok := request.(int64)
		if !ok {
			return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		}
		history, err := svc.GetQuotationHandoverHistory(ctx, orgID, quotationID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Operational handover history retrieved successfully",
			Data:    history,
		}, nil
	}
}

// ── Quotation Analytics & Performance Endpoints (Task 18.8) ──────────────────

func makeGetQuotationAnalyticsOverviewEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		overview, err := svc.GetQuotationAnalyticsOverview(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Quotation analytics overview retrieved successfully",
			Data:    overview,
		}, nil
	}
}

func makeGetQuotationAnalyticsTrendsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		days, ok := request.(int)
		if !ok || days <= 0 {
			days = 30
		}
		trends, err := svc.GetQuotationAnalyticsTrends(ctx, orgID, days)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Quotation analytics trends retrieved successfully",
			Data:    trends,
		}, nil
	}
}

func makeGetCustomerQuotationPerformanceEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		custPerf, err := svc.GetCustomerQuotationPerformance(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Customer quotation performance retrieved successfully",
			Data:    custPerf,
		}, nil
	}
}

func makeGetQuotationPerformanceByModeEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		modes, err := svc.GetQuotationPerformanceByMode(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Quotation performance by mode retrieved successfully",
			Data:    modes,
		}, nil
	}
}

func makeGetQuotationExpiryRiskEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		risks, err := svc.GetQuotationExpiryRisk(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Quotation expiry risk items retrieved successfully",
			Data:    risks,
		}, nil
	}
}

// ── Task 19.5: Rate-to-Quotation Integration Endpoint Makers ─────────────────

func makeGetQuotationRateCandidatesEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		qID := request.(int64)
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := svc.GetQuotationRateCandidates(ctx, orgID, qID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Rate candidates retrieved successfully",
			Data:    resp,
		}, nil
	}
}

func makeGetQuotationRateSelectionEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		qID := request.(int64)
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}

		sel, err := svc.GetQuotationRateSelection(ctx, orgID, qID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Active quotation rate selection retrieved successfully",
			Data:    sel,
		}, nil
	}
}

func makeSelectQuotationRateEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*SelectQuotationRateRequest)
		orgID, uID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID
		if req.User == "" {
			uc, _ := middleware.GetUserContext(ctx)
			req.User = uc.CognitoID
			if req.User == "" {
				req.User = strconv.FormatInt(uID, 10)
			}
		}

		snap, err := svc.SelectQuotationRate(ctx, req)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Commercial rate selected and snapshot created successfully",
			Data:    snap,
		}, nil
	}
}

func makeReplaceQuotationRateEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*ReplaceQuotationRateRequest)
		orgID, uID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID
		if req.User == "" {
			uc, _ := middleware.GetUserContext(ctx)
			req.User = uc.CognitoID
			if req.User == "" {
				req.User = strconv.FormatInt(uID, 10)
			}
		}

		snap, err := svc.ReplaceQuotationRate(ctx, req)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Quotation rate replaced and new snapshot created successfully",
			Data:    snap,
		}, nil
	}
}

func makeRemoveQuotationRateEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		qID := request.(int64)
		orgID, uID, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}
		uc, _ := middleware.GetUserContext(ctx)
		user := uc.CognitoID
		if user == "" {
			user = strconv.FormatInt(uID, 10)
		}

		if err := svc.RemoveQuotationRate(ctx, orgID, qID, user); err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Quotation rate selection removed successfully",
		}, nil
	}
}

func makeGetQuotationRateSnapshotEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		qID := request.(int64)
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}

		snap, err := svc.GetQuotationRateSnapshot(ctx, orgID, qID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Quotation rate snapshot retrieved successfully",
			Data:    snap,
		}, nil
	}
}

func makeGetQuotationRateSelectionHistoryEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		qID := request.(int64)
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}

		history, err := svc.GetQuotationRateSelectionHistory(ctx, orgID, qID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Quotation rate selection history retrieved successfully",
			Data:    history,
		}, nil
	}
}

// ── Task 19.6: Quotation Rate Risk & Commercial Impact Endpoints ───────────────

func makeGetQuotationRateRisksEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		qID := request.(int64)
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}

		summary, err := svc.GetQuotationRateRisks(ctx, orgID, qID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Quotation rate risks retrieved successfully",
			Data:    summary,
		}, nil
	}
}

type resolveRiskEndpointRequest struct {
	QuotationID int64
	RiskID      int64
	User        string
}

func makeResolveQuotationRateRiskEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*resolveRiskEndpointRequest)
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}

		if err := svc.ResolveQuotationRateRisk(ctx, orgID, req.QuotationID, req.RiskID, req.User); err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Quotation rate risk marked as resolved",
		}, nil
	}
}

func makeGetRateReplacementCandidatesEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		qID := request.(int64)
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}

		candidates, err := svc.GetRateReplacementCandidates(ctx, orgID, qID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Rate replacement candidates retrieved successfully",
			Data:    candidates,
		}, nil
	}
}

type commercialImpactEndpointRequest struct {
	QuotationID       int64
	ReplacementRateID *int64
	ReplacementSpotID *int64
}

func makeGetCommercialImpactAnalysisEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*commercialImpactEndpointRequest)
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}

		analysis, err := svc.GetCommercialImpactAnalysis(ctx, orgID, req.QuotationID, req.ReplacementRateID, req.ReplacementSpotID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: "Commercial impact analysis calculated successfully",
			Data:    analysis,
		}, nil
	}
}

func makeEvaluateQuotationRateRisksEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, _, err := getOrgAndUserID(ctx)
		if err != nil {
			return nil, err
		}

		count, err := svc.EvaluateQuotationRateRisksForOrg(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return &APIResponse{
			Success: true,
			Message: fmt.Sprintf("Quotation rate risk evaluation completed, %d risks evaluated/updated", count),
			Data: map[string]interface{}{
				"risks_evaluated": count,
			},
		}, nil
	}
}







