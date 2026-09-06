package contracts

import (
	"context"
	"strconv"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-kit/kit/endpoint"
)

// Endpoints holds all Go kit endpoints for the Contracts service.
type Endpoints struct {
	CreateContractEP             endpoint.Endpoint
	GetContractEP                endpoint.Endpoint
	ListContractsEP              endpoint.Endpoint
	UpdateContractEP             endpoint.Endpoint
	UpdateContractLifecycleEP    endpoint.Endpoint
	ArchiveContractEP            endpoint.Endpoint
	GetContractLifecycleEventsEP endpoint.Endpoint
	GetContractOverviewEP        endpoint.Endpoint
	AddLinkEP                    endpoint.Endpoint
	RemoveLinkEP                 endpoint.Endpoint
	GetRelationshipSummaryEP     endpoint.Endpoint
	GetLinkHistoryEP             endpoint.Endpoint

	// Task 20.3: Lifecycle Intelligence, Renewal & Risk Endpoints
	EvaluateLifecycleEP                endpoint.Endpoint
	GetLifecycleSummaryEP              endpoint.Endpoint
	GetAttentionItemsEP                endpoint.Endpoint
	GetContractLifecycleIntelligenceEP endpoint.Endpoint
	GetLifecycleEventsEP               endpoint.Endpoint
	GetContractRisksEP                 endpoint.Endpoint
	ResolveRiskEP                      endpoint.Endpoint
	GetRenewalTrackingEP               endpoint.Endpoint
	StartRenewalEP                     endpoint.Endpoint
	UpdateRenewalTrackingEP            endpoint.Endpoint
	GetCommercialImpactEP              endpoint.Endpoint

	// Task 20.4: Versioning, Amendments & Approvals Endpoints
	CreateContractVersionEP        endpoint.Endpoint
	GetContractVersionsEP          endpoint.Endpoint
	GetContractVersionEP           endpoint.Endpoint
	GetContractVersionComparisonEP endpoint.Endpoint
	MakeVersionEffectiveEP         endpoint.Endpoint

	CreateContractAmendmentEP    endpoint.Endpoint
	GetContractAmendmentsEP      endpoint.Endpoint
	GetContractAmendmentEP       endpoint.Endpoint
	UpdateContractAmendmentEP    endpoint.Endpoint
	SubmitContractAmendmentEP    endpoint.Endpoint
	ImplementContractAmendmentEP endpoint.Endpoint
	CancelContractAmendmentEP    endpoint.Endpoint
	GetAmendmentChangesEP        endpoint.Endpoint

	GetContractApprovalsEP   endpoint.Endpoint
	ApproveContractChangeEP  endpoint.Endpoint
	RejectContractChangeEP   endpoint.Endpoint
	CancelContractApprovalEP endpoint.Endpoint

	// Tasks 20.5 & 20.6 Endpoints
	ListDocumentsEP      endpoint.Endpoint
	CreateDocumentEP     endpoint.Endpoint
	SupersedeDocumentEP  endpoint.Endpoint

	ListTermsEP   endpoint.Endpoint
	CreateTermEP  endpoint.Endpoint
	UpdateTermEP  endpoint.Endpoint
	DeleteTermEP  endpoint.Endpoint

	ListObligationsEP    endpoint.Endpoint
	CreateObligationEP   endpoint.Endpoint
	UpdateObligationEP   endpoint.Endpoint
	FulfillObligationEP  endpoint.Endpoint
	WaiveObligationEP    endpoint.Endpoint

	ListComplianceEventsEP          endpoint.Endpoint
	ResolveComplianceEventEP        endpoint.Endpoint
	ListComplianceRequirementsEP    endpoint.Endpoint
	CreateComplianceRequirementEP   endpoint.Endpoint
	VerifyComplianceRequirementEP   endpoint.Endpoint

	GetContractPerformanceEP              endpoint.Endpoint
	GetContractOperationalIntelligenceEP  endpoint.Endpoint
	GetComplianceSummaryEP                endpoint.Endpoint
	GetOpenComplianceAttentionEP          endpoint.Endpoint
	EvaluateComplianceEP                  endpoint.Endpoint

	// Contract Document Import & AI-Assisted Contract Creation
	ImportContractDocumentEP    endpoint.Endpoint
	GetExtractedContractDraftEP endpoint.Endpoint
	ConfirmContractImportEP     endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct wrapping the given BusinessLogic.
func MakeServerEndpoints(bl BusinessLogic) Endpoints {
	return Endpoints{
		CreateContractEP:             makeCreateContractEndpoint(bl),
		GetContractEP:                makeGetContractEndpoint(bl),
		ListContractsEP:              makeListContractsEndpoint(bl),
		UpdateContractEP:             makeUpdateContractEndpoint(bl),
		UpdateContractLifecycleEP:    makeUpdateContractLifecycleEndpoint(bl),
		ArchiveContractEP:            makeArchiveContractEndpoint(bl),
		GetContractLifecycleEventsEP: makeGetContractLifecycleEventsEndpoint(bl),
		GetContractOverviewEP:        makeGetContractOverviewEndpoint(bl),
		AddLinkEP:                    makeAddLinkEndpoint(bl),
		RemoveLinkEP:                 makeRemoveLinkEndpoint(bl),
		GetRelationshipSummaryEP:     makeGetRelationshipSummaryEndpoint(bl),
		GetLinkHistoryEP:             makeGetLinkHistoryEndpoint(bl),

		EvaluateLifecycleEP:                makeEvaluateLifecycleEndpoint(bl),
		GetLifecycleSummaryEP:              makeGetLifecycleSummaryEndpoint(bl),
		GetAttentionItemsEP:                makeGetAttentionItemsEndpoint(bl),
		GetContractLifecycleIntelligenceEP: makeGetContractLifecycleIntelligenceEndpoint(bl),
		GetLifecycleEventsEP:               makeGetLifecycleEventsEndpoint(bl),
		GetContractRisksEP:                 makeGetContractRisksEndpoint(bl),
		ResolveRiskEP:                      makeResolveRiskEndpoint(bl),
		GetRenewalTrackingEP:               makeGetRenewalTrackingEndpoint(bl),
		StartRenewalEP:                     makeStartRenewalEndpoint(bl),
		UpdateRenewalTrackingEP:            makeUpdateRenewalTrackingEndpoint(bl),
		GetCommercialImpactEP:              makeGetCommercialImpactEndpoint(bl),

		CreateContractVersionEP:        makeCreateContractVersionEndpoint(bl),
		GetContractVersionsEP:          makeGetContractVersionsEndpoint(bl),
		GetContractVersionEP:           makeGetContractVersionEndpoint(bl),
		GetContractVersionComparisonEP: makeGetContractVersionComparisonEndpoint(bl),
		MakeVersionEffectiveEP:         makeMakeVersionEffectiveEndpoint(bl),

		CreateContractAmendmentEP:    makeCreateContractAmendmentEndpoint(bl),
		GetContractAmendmentsEP:      makeGetContractAmendmentsEndpoint(bl),
		GetContractAmendmentEP:       makeGetContractAmendmentEndpoint(bl),
		UpdateContractAmendmentEP:    makeUpdateContractAmendmentEndpoint(bl),
		SubmitContractAmendmentEP:    makeSubmitContractAmendmentEndpoint(bl),
		ImplementContractAmendmentEP: makeImplementContractAmendmentEndpoint(bl),
		CancelContractAmendmentEP:    makeCancelContractAmendmentEndpoint(bl),
		GetAmendmentChangesEP:        makeGetAmendmentChangesEndpoint(bl),

		GetContractApprovalsEP:   makeGetContractApprovalsEndpoint(bl),
		ApproveContractChangeEP:  makeApproveContractChangeEndpoint(bl),
		RejectContractChangeEP:   makeRejectContractChangeEndpoint(bl),
		CancelContractApprovalEP: makeCancelContractApprovalEndpoint(bl),

		ListDocumentsEP:      makeListDocumentsEndpoint(bl),
		CreateDocumentEP:     makeCreateDocumentEndpoint(bl),
		SupersedeDocumentEP:  makeSupersedeDocumentEndpoint(bl),

		ListTermsEP:   makeListTermsEndpoint(bl),
		CreateTermEP:  makeCreateTermEndpoint(bl),
		UpdateTermEP:  makeUpdateTermEndpoint(bl),
		DeleteTermEP:  makeDeleteTermEndpoint(bl),

		ListObligationsEP:    makeListObligationsEndpoint(bl),
		CreateObligationEP:   makeCreateObligationEndpoint(bl),
		UpdateObligationEP:   makeUpdateObligationEndpoint(bl),
		FulfillObligationEP:  makeFulfillObligationEndpoint(bl),
		WaiveObligationEP:    makeWaiveObligationEndpoint(bl),

		ListComplianceEventsEP:          makeListComplianceEventsEndpoint(bl),
		ResolveComplianceEventEP:        makeResolveComplianceEventEndpoint(bl),
		ListComplianceRequirementsEP:    makeListComplianceRequirementsEndpoint(bl),
		CreateComplianceRequirementEP:   makeCreateComplianceRequirementEndpoint(bl),
		VerifyComplianceRequirementEP:   makeVerifyComplianceRequirementEndpoint(bl),

		GetContractPerformanceEP:              makeGetContractPerformanceEndpoint(bl),
		GetContractOperationalIntelligenceEP:  makeGetContractOperationalIntelligenceEndpoint(bl),
		GetComplianceSummaryEP:                makeGetComplianceSummaryEndpoint(bl),
		GetOpenComplianceAttentionEP:          makeGetOpenComplianceAttentionEndpoint(bl),
		EvaluateComplianceEP:                  makeEvaluateComplianceEndpoint(bl),

		ImportContractDocumentEP:    makeImportContractDocumentEndpoint(bl),
		GetExtractedContractDraftEP: makeGetExtractedContractDraftEndpoint(bl),
		ConfirmContractImportEP:     makeConfirmContractImportEndpoint(bl),
	}
}

func getUserContext(ctx context.Context) (middleware.UserContext, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		return middleware.UserContext{}, svcerror.NewServiceError(svcerror.ErrInsufficientResourceAccess)
	}
	return userCtx, nil
}

func makeCreateContractEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*CreateContractRequest)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		userIDStr := strconv.FormatInt(userCtx.UserID, 10)
		return bl.CreateContract(ctx, userCtx.OrgID, userIDStr, req)
	}
}

func makeGetContractEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		contractID := request.(int64)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		return bl.GetContract(ctx, userCtx.OrgID, contractID)
	}
}

func makeListContractsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*ListContractsRequest)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		return bl.ListContracts(ctx, userCtx.OrgID, req)
	}
}

type updateContractRequestWrapper struct {
	ContractID int64
	Req        *UpdateContractRequest
}

type updateLifecycleRequestWrapper struct {
	ContractID int64
	Req        *UpdateContractLifecycleRequest
}

func makeUpdateContractEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateContractRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		userIDStr := strconv.FormatInt(userCtx.UserID, 10)
		return bl.UpdateContract(ctx, userCtx.OrgID, req.ContractID, userIDStr, req.Req)
	}
}

func makeUpdateContractLifecycleEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateLifecycleRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		userIDStr := strconv.FormatInt(userCtx.UserID, 10)
		return bl.UpdateContractLifecycle(ctx, userCtx.OrgID, req.ContractID, userIDStr, req.Req)
	}
}

func makeArchiveContractEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		contractID := request.(int64)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		userIDStr := strconv.FormatInt(userCtx.UserID, 10)
		err = bl.ArchiveContract(ctx, userCtx.OrgID, contractID, userIDStr)
		if err != nil {
			return nil, err
		}
		return struct{}{}, nil
	}
}

func makeGetContractLifecycleEventsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		contractID := request.(int64)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		return bl.GetContractLifecycleEvents(ctx, userCtx.OrgID, contractID)
	}
}

func makeGetContractOverviewEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		return bl.GetContractOverview(ctx, userCtx.OrgID)
	}
}

type addLinkRequest struct {
	ContractID int64
	CreateContractLinkRequest
}

func makeAddLinkEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addLinkRequest)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		link, err := bl.AddContractLink(ctx, userCtx.OrgID, req.ContractID, strconv.FormatInt(userCtx.UserID, 10), &req.CreateContractLinkRequest)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": link}, nil
	}
}

type removeLinkRequest struct {
	ContractID int64
	LinkID     int64
}

func makeRemoveLinkEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(removeLinkRequest)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		err = bl.RemoveContractLink(ctx, userCtx.OrgID, req.ContractID, req.LinkID, strconv.FormatInt(userCtx.UserID, 10))
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"message": "link removed"}, nil
	}
}

func makeGetRelationshipSummaryEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		contractID := request.(int64)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		summary, err := bl.GetContractRelationshipSummary(ctx, userCtx.OrgID, contractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": summary}, nil
	}
}



func makeGetLinkHistoryEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		contractID := request.(int64)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		history, err := bl.GetContractLinkHistory(ctx, userCtx.OrgID, contractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": history}, nil
	}
}

// ─── Task 20.3: Lifecycle Intelligence Endpoint Handlers ───────────────────

func makeEvaluateLifecycleEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		summary, err := bl.EvaluateLifecycleForOrg(ctx, userCtx.OrgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": summary}, nil
	}
}

func makeGetLifecycleSummaryEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		summary, err := bl.GetLifecycleSummary(ctx, userCtx.OrgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": summary}, nil
	}
}

func makeGetAttentionItemsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		items, err := bl.GetContractsRequiringAttention(ctx, userCtx.OrgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": items}, nil
	}
}

func makeGetContractLifecycleIntelligenceEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		contractID := request.(int64)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		detail, err := bl.GetContractLifecycleIntelligence(ctx, userCtx.OrgID, contractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": detail}, nil
	}
}

type getLifecycleEventsRequest struct {
	ContractID *int64
	Limit      int
}

func makeGetLifecycleEventsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getLifecycleEventsRequest)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		events, err := bl.GetLifecycleEvents(ctx, userCtx.OrgID, req.ContractID, req.Limit)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": events}, nil
	}
}

func makeGetContractRisksEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		contractID := request.(int64)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		risks, err := bl.GetContractRisks(ctx, userCtx.OrgID, contractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": risks}, nil
	}
}

type resolveRiskRequestWrapper struct {
	ContractID int64
	RiskID     int64
	Payload    *ResolveRiskRequest
}

func makeResolveRiskEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(resolveRiskRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		err = bl.ResolveRisk(ctx, userCtx.OrgID, req.ContractID, req.RiskID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"message": "risk resolved"}, nil
	}
}

func makeGetRenewalTrackingEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		contractID := request.(int64)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		tracking, err := bl.GetRenewalTracking(ctx, userCtx.OrgID, contractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": tracking}, nil
	}
}

type startRenewalRequestWrapper struct {
	ContractID int64
	Payload    *StartRenewalRequest
}

func makeStartRenewalEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(startRenewalRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		tracking, err := bl.StartRenewal(ctx, userCtx.OrgID, req.ContractID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": tracking}, nil
	}
}

type updateRenewalRequestWrapper struct {
	ContractID int64
	Payload    *UpdateRenewalRequest
}

func makeUpdateRenewalTrackingEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateRenewalRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		tracking, err := bl.UpdateRenewalTracking(ctx, userCtx.OrgID, req.ContractID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": tracking}, nil
	}
}

func makeGetCommercialImpactEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		contractID := request.(int64)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		impact, err := bl.GetCommercialImpact(ctx, userCtx.OrgID, contractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": impact}, nil
	}
}

// ── Task 20.4 Endpoint Handlers ─────────────────────────────────────────────

type createVersionRequestWrapper struct {
	ContractID int64
	Payload    *CreateContractVersionRequest
}

func makeCreateContractVersionEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createVersionRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		ver, err := bl.CreateContractVersion(ctx, userCtx.OrgID, req.ContractID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": ver}, nil
	}
}

func makeGetContractVersionsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		contractID := request.(int64)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		versions, err := bl.GetContractVersions(ctx, userCtx.OrgID, contractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": versions}, nil
	}
}

type getVersionRequestWrapper struct {
	ContractID int64
	VersionID  string
}

func makeGetContractVersionEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getVersionRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		ver, err := bl.GetContractVersion(ctx, userCtx.OrgID, req.ContractID, req.VersionID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": ver}, nil
	}
}

type compareVersionsRequestWrapper struct {
	ContractID      int64
	BaseVersionID   string
	TargetVersionID string
}

func makeGetContractVersionComparisonEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(compareVersionsRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		comparison, err := bl.GetContractVersionComparison(ctx, userCtx.OrgID, req.ContractID, req.BaseVersionID, req.TargetVersionID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": comparison}, nil
	}
}

func makeMakeVersionEffectiveEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getVersionRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		ver, err := bl.MakeVersionEffective(ctx, userCtx.OrgID, req.ContractID, req.VersionID, strconv.FormatInt(userCtx.UserID, 10))
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": ver}, nil
	}
}

type createAmendmentRequestWrapper struct {
	ContractID int64
	Payload    *CreateContractAmendmentRequest
}

func makeCreateContractAmendmentEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createAmendmentRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		amendment, err := bl.CreateContractAmendment(ctx, userCtx.OrgID, req.ContractID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": amendment}, nil
	}
}

func makeGetContractAmendmentsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		contractID := request.(int64)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		amendments, err := bl.GetContractAmendments(ctx, userCtx.OrgID, contractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": amendments}, nil
	}
}

type getAmendmentRequestWrapper struct {
	ContractID  int64
	AmendmentID string
}

func makeGetContractAmendmentEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getAmendmentRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		amendment, err := bl.GetContractAmendment(ctx, userCtx.OrgID, req.ContractID, req.AmendmentID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": amendment}, nil
	}
}

type updateAmendmentRequestWrapper struct {
	ContractID  int64
	AmendmentID string
	Payload     *UpdateContractAmendmentRequest
}

func makeUpdateContractAmendmentEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateAmendmentRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		amendment, err := bl.UpdateContractAmendment(ctx, userCtx.OrgID, req.ContractID, req.AmendmentID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": amendment}, nil
	}
}

type submitAmendmentRequestWrapper struct {
	ContractID  int64
	AmendmentID string
	Payload     *SubmitContractAmendmentRequest
}

func makeSubmitContractAmendmentEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(submitAmendmentRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		amendment, err := bl.SubmitContractAmendment(ctx, userCtx.OrgID, req.ContractID, req.AmendmentID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": amendment}, nil
	}
}

func makeImplementContractAmendmentEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getAmendmentRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		ver, err := bl.ImplementContractAmendment(ctx, userCtx.OrgID, req.ContractID, req.AmendmentID, strconv.FormatInt(userCtx.UserID, 10))
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": ver}, nil
	}
}

func makeCancelContractAmendmentEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getAmendmentRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		err = bl.CancelContractAmendment(ctx, userCtx.OrgID, req.ContractID, req.AmendmentID, strconv.FormatInt(userCtx.UserID, 10))
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": map[string]string{"status": "CANCELLED"}}, nil
	}
}

func makeGetAmendmentChangesEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getAmendmentRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		changes, err := bl.GetAmendmentChanges(ctx, userCtx.OrgID, req.ContractID, req.AmendmentID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": changes}, nil
	}
}

func makeGetContractApprovalsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		contractID := request.(int64)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		approvals, err := bl.GetContractApprovals(ctx, userCtx.OrgID, contractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": approvals}, nil
	}
}

type approveRequestWrapper struct {
	ContractID int64
	ApprovalID string
	Payload    *ApproveContractRequest
}

func makeApproveContractChangeEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(approveRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		res, err := bl.ApproveContractChange(ctx, userCtx.OrgID, req.ContractID, req.ApprovalID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": res}, nil
	}
}

type rejectRequestWrapper struct {
	ContractID int64
	ApprovalID string
	Payload    *RejectContractRequest
}

func makeRejectContractChangeEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(rejectRequestWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		res, err := bl.RejectContractChange(ctx, userCtx.OrgID, req.ContractID, req.ApprovalID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": res}, nil
	}
}

type cancelApprovalWrapper struct {
	ContractID int64
	ApprovalID string
	Payload    *CancelApprovalRequest
}

func makeCancelContractApprovalEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(cancelApprovalWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		res, err := bl.CancelContractApproval(ctx, userCtx.OrgID, req.ContractID, req.ApprovalID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": res}, nil
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// Tasks 20.5 & 20.6: Endpoint Handlers & Request Wrappers
// ═══════════════════════════════════════════════════════════════════════════

type contractIDOnlyWrapper struct {
	ContractID int64
}

type createDocumentWrapper struct {
	ContractID int64
	Payload    *CreateAgreementDocumentRequest
}

type supersedeDocWrapper struct {
	ContractID int64
	DocID      string
	NewDocID   string
}

func makeListDocumentsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(contractIDOnlyWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		docs, err := bl.ListContractDocuments(ctx, userCtx.OrgID, req.ContractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": docs}, nil
	}
}

func makeCreateDocumentEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createDocumentWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		doc, err := bl.CreateContractDocument(ctx, userCtx.OrgID, req.ContractID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": doc}, nil
	}
}

func makeSupersedeDocumentEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(supersedeDocWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		err = bl.SupersedeContractDocument(ctx, userCtx.OrgID, req.ContractID, req.DocID, req.NewDocID, strconv.FormatInt(userCtx.UserID, 10))
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

type createTermWrapper struct {
	ContractID int64
	Payload    *CreateContractTermRequest
}

type updateTermWrapper struct {
	ContractID int64
	TermID     int64
	Payload    *UpdateContractTermRequest
}

type deleteTermWrapper struct {
	ContractID int64
	TermID     int64
}

func makeListTermsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(contractIDOnlyWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		terms, err := bl.ListContractTerms(ctx, userCtx.OrgID, req.ContractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": terms}, nil
	}
}

func makeCreateTermEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createTermWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		term, err := bl.CreateContractTerm(ctx, userCtx.OrgID, req.ContractID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": term}, nil
	}
}

func makeUpdateTermEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateTermWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		term, err := bl.UpdateContractTerm(ctx, userCtx.OrgID, req.ContractID, req.TermID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": term}, nil
	}
}

func makeDeleteTermEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteTermWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		err = bl.DeleteContractTerm(ctx, userCtx.OrgID, req.ContractID, req.TermID, strconv.FormatInt(userCtx.UserID, 10))
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

type createObligationWrapper struct {
	ContractID int64
	Payload    *CreateContractObligationRequest
}

type updateObligationWrapper struct {
	ContractID   int64
	ObligationID int64
	Payload      *UpdateContractObligationRequest
}

type fulfillObligationWrapper struct {
	ContractID   int64
	ObligationID int64
	Payload      *FulfillObligationRequest
}

type waiveObligationWrapper struct {
	ContractID   int64
	ObligationID int64
	Payload      *WaiveObligationRequest
}

func makeListObligationsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(contractIDOnlyWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		obs, err := bl.ListContractObligations(ctx, userCtx.OrgID, req.ContractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": obs}, nil
	}
}

func makeCreateObligationEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createObligationWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		ob, err := bl.CreateContractObligation(ctx, userCtx.OrgID, req.ContractID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": ob}, nil
	}
}

func makeUpdateObligationEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateObligationWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		ob, err := bl.UpdateContractObligation(ctx, userCtx.OrgID, req.ContractID, req.ObligationID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": ob}, nil
	}
}

func makeFulfillObligationEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(fulfillObligationWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		ob, err := bl.FulfillContractObligation(ctx, userCtx.OrgID, req.ContractID, req.ObligationID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": ob}, nil
	}
}

func makeWaiveObligationEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(waiveObligationWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		ob, err := bl.WaiveContractObligation(ctx, userCtx.OrgID, req.ContractID, req.ObligationID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": ob}, nil
	}
}

type resolveComplianceEventWrapper struct {
	ContractID int64
	EventID    int64
	Payload    *ResolveComplianceEventRequest
}

func makeListComplianceEventsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(contractIDOnlyWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		events, err := bl.ListContractComplianceEvents(ctx, userCtx.OrgID, req.ContractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": events}, nil
	}
}

func makeResolveComplianceEventEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(resolveComplianceEventWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		err = bl.ResolveComplianceEvent(ctx, userCtx.OrgID, req.ContractID, req.EventID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	}
}

type createComplianceReqWrapper struct {
	ContractID int64
	Payload    *CreateComplianceRequirementRequest
}

type verifyComplianceReqWrapper struct {
	ContractID    int64
	RequirementID int64
	Payload       *VerifyComplianceRequest
}

func makeListComplianceRequirementsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(contractIDOnlyWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		reqs, err := bl.ListContractComplianceRequirements(ctx, userCtx.OrgID, req.ContractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": reqs}, nil
	}
}

func makeCreateComplianceRequirementEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createComplianceReqWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		cr, err := bl.CreateContractComplianceRequirement(ctx, userCtx.OrgID, req.ContractID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": cr}, nil
	}
}

func makeVerifyComplianceRequirementEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(verifyComplianceReqWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		cr, err := bl.VerifyContractComplianceRequirement(ctx, userCtx.OrgID, req.ContractID, req.RequirementID, strconv.FormatInt(userCtx.UserID, 10), req.Payload)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": cr}, nil
	}
}

func makeGetContractPerformanceEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(contractIDOnlyWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		perf, err := bl.GetContractPerformance(ctx, userCtx.OrgID, req.ContractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": perf}, nil
	}
}

func makeGetContractOperationalIntelligenceEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(contractIDOnlyWrapper)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		data, err := bl.GetContractOperationalIntelligence(ctx, userCtx.OrgID, req.ContractID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": data}, nil
	}
}

func makeGetComplianceSummaryEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		summary, err := bl.GetComplianceSummary(ctx, userCtx.OrgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": summary}, nil
	}
}

func makeGetOpenComplianceAttentionEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		events, err := bl.GetOpenComplianceAttention(ctx, userCtx.OrgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": events}, nil
	}
}

func makeEvaluateComplianceEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		summary, err := bl.EvaluateContractComplianceForOrg(ctx, userCtx.OrgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"data": summary}, nil
	}
}

// ── Contract Document Import & AI-Assisted Contract Creation Endpoints ───────

type importDocumentRequest struct {
	FileName         string
	FileBytes        []byte
	ContractTypeHint *string
}

func makeImportContractDocumentEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		req, ok := request.(importDocumentRequest)
		if !ok {
			return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		}
		res, err := bl.ImportContractDocument(ctx, userCtx.OrgID, strconv.FormatInt(userCtx.UserID, 10), req.FileName, req.FileBytes, req.ContractTypeHint)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":  "success",
			"message": "Contract document uploaded and analyzed successfully",
			"data":    res,
		}, nil
	}
}

func makeGetExtractedContractDraftEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		docID, ok := request.(string)
		if !ok || docID == "" {
			return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		}
		res, err := bl.GetExtractedContractDraft(ctx, userCtx.OrgID, docID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status": "success",
			"data":   res,
		}, nil
	}
}

func makeConfirmContractImportEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		req, ok := request.(*ConfirmContractImportRequest)
		if !ok || req == nil {
			return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		}
		contract, err := bl.ConfirmContractImport(ctx, userCtx.OrgID, strconv.FormatInt(userCtx.UserID, 10), req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":  "success",
			"message": "Contract created successfully from document import",
			"data":    contract,
		}, nil
	}
}




