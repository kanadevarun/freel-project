package rates

import (
	"context"
	"strconv"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/rates/spec"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-kit/kit/endpoint"
)

// Endpoints contains all Go-Kit endpoints for Rate Management
type Endpoints struct {
	ListRatesEP          endpoint.Endpoint
	GetRateSummaryEP     endpoint.Endpoint
	CreateRateEP         endpoint.Endpoint
	GetRateEP            endpoint.Endpoint
	UpdateRateEP         endpoint.Endpoint
	ArchiveRateEP        endpoint.Endpoint
	SearchRatesEP        endpoint.Endpoint
	RefreshSpotRatesEP   endpoint.Endpoint
	GetRatePricingEP     endpoint.Endpoint
	AddRateChargeEP      endpoint.Endpoint
	UpdateRateChargeEP   endpoint.Endpoint
	DeleteRateChargeEP   endpoint.Endpoint
	ReorderRateChargesEP endpoint.Endpoint

	// Task 19.3: Contracts & Versions
	ListRateContractsEP      endpoint.Endpoint
	GetRateContractSummaryEP endpoint.Endpoint
	CreateRateContractEP     endpoint.Endpoint
	GetRateContractEP        endpoint.Endpoint
	UpdateRateContractEP     endpoint.Endpoint
	ArchiveRateContractEP    endpoint.Endpoint
	RenewRateContractEP      endpoint.Endpoint
	CreateRateVersionEP      endpoint.Endpoint
	GetRateVersionsEP        endpoint.Endpoint
	GetRateVersionHistoryEP  endpoint.Endpoint

	// Task 19.4: Spot Rate Requests, Responses & Comparison
	ListSpotRateRequestsEP    endpoint.Endpoint
	GetSpotRateSummaryEP      endpoint.Endpoint
	CreateSpotRateRequestEP   endpoint.Endpoint
	GetSpotRateRequestEP      endpoint.Endpoint
	UpdateSpotRateRequestEP   endpoint.Endpoint
	SendSpotRateRequestEP     endpoint.Endpoint
	CancelSpotRateRequestEP   endpoint.Endpoint
	CreateSpotRateResponseEP  endpoint.Endpoint
	GetSpotRateResponseEP     endpoint.Endpoint
	GetSpotRateResponsesEP    endpoint.Endpoint
	UpdateSpotRateResponseEP  endpoint.Endpoint
	SelectPreferredSpotRateEP endpoint.Endpoint
	CompareSpotRatesEP        endpoint.Endpoint

	// Task 19.6: Rate Lifecycle Intelligence
	GetRateLifecycleSummaryEP        endpoint.Endpoint
	GetRateLifecycleEventsEP         endpoint.Endpoint
	GetRatesRequiringAttentionEP     endpoint.Endpoint
	GetContractsRequiringAttentionEP endpoint.Endpoint
	EvaluateRateLifecycleEP          endpoint.Endpoint

	// Task 19.7: Rate Analytics & Procurement Intelligence
	GetRateAnalyticsOverviewEP   endpoint.Endpoint
	GetRateAnalyticsTrendsEP     endpoint.Endpoint
	GetCarrierRatePerformanceEP  endpoint.Endpoint
	GetLaneRatePerformanceEP     endpoint.Endpoint
	GetRateLifecycleAnalyticsEP  endpoint.Endpoint
	GetSpotSourcingPerformanceEP endpoint.Endpoint
	GetRateCommercialInsightsEP  endpoint.Endpoint

	// Task 5: Live Carrier Integration Rates
	SearchCarrierLiveRatesEP endpoint.Endpoint
}

// NewAllRatesEndpoints initializes all endpoints wired to business logic
func NewAllRatesEndpoints(bl BusinessLogic) Endpoints {
	return Endpoints{
		ListRatesEP:          makeListRatesEndpoint(bl),
		GetRateSummaryEP:     makeGetRateSummaryEndpoint(bl),
		CreateRateEP:         makeCreateRateEndpoint(bl),
		GetRateEP:            makeGetRateEndpoint(bl),
		UpdateRateEP:         makeUpdateRateEndpoint(bl),
		ArchiveRateEP:        makeArchiveRateEndpoint(bl),
		SearchRatesEP:        makeSearchRatesEndpoint(bl),
		RefreshSpotRatesEP:   makeRefreshSpotRatesEndpoint(bl),
		GetRatePricingEP:     makeGetRatePricingEndpoint(bl),
		AddRateChargeEP:      makeAddRateChargeEndpoint(bl),
		UpdateRateChargeEP:   makeUpdateRateChargeEndpoint(bl),
		DeleteRateChargeEP:   makeDeleteRateChargeEndpoint(bl),
		ReorderRateChargesEP: makeReorderRateChargesEndpoint(bl),

		// Task 5: Live Carrier Integration Rates
		SearchCarrierLiveRatesEP: makeSearchCarrierLiveRatesEndpoint(bl),

		// Task 19.3
		ListRateContractsEP:      makeListRateContractsEndpoint(bl),
		GetRateContractSummaryEP: makeGetRateContractSummaryEndpoint(bl),
		CreateRateContractEP:     makeCreateRateContractEndpoint(bl),
		GetRateContractEP:        makeGetRateContractEndpoint(bl),
		UpdateRateContractEP:     makeUpdateRateContractEndpoint(bl),
		ArchiveRateContractEP:    makeArchiveRateContractEndpoint(bl),
		RenewRateContractEP:      makeRenewRateContractEndpoint(bl),
		CreateRateVersionEP:      makeCreateRateVersionEndpoint(bl),
		GetRateVersionsEP:        makeGetRateVersionsEndpoint(bl),
		GetRateVersionHistoryEP:  makeGetRateVersionHistoryEndpoint(bl),

		// Task 19.4
		ListSpotRateRequestsEP:    makeListSpotRateRequestsEndpoint(bl),
		GetSpotRateSummaryEP:      makeGetSpotRateSummaryEndpoint(bl),
		CreateSpotRateRequestEP:   makeCreateSpotRateRequestEndpoint(bl),
		GetSpotRateRequestEP:      makeGetSpotRateRequestEndpoint(bl),
		UpdateSpotRateRequestEP:   makeUpdateSpotRateRequestEndpoint(bl),
		SendSpotRateRequestEP:     makeSendSpotRateRequestEndpoint(bl),
		CancelSpotRateRequestEP:   makeCancelSpotRateRequestEndpoint(bl),
		CreateSpotRateResponseEP:  makeCreateSpotRateResponseEndpoint(bl),
		GetSpotRateResponseEP:     makeGetSpotRateResponseEndpoint(bl),
		GetSpotRateResponsesEP:    makeGetSpotRateResponsesEndpoint(bl),
		UpdateSpotRateResponseEP:  makeUpdateSpotRateResponseEndpoint(bl),
		SelectPreferredSpotRateEP: makeSelectPreferredSpotRateEndpoint(bl),
		CompareSpotRatesEP:        makeCompareSpotRatesEndpoint(bl),

		// Task 19.6
		GetRateLifecycleSummaryEP:        makeGetRateLifecycleSummaryEndpoint(bl),
		GetRateLifecycleEventsEP:         makeGetRateLifecycleEventsEndpoint(bl),
		GetRatesRequiringAttentionEP:     makeGetRatesRequiringAttentionEndpoint(bl),
		GetContractsRequiringAttentionEP: makeGetContractsRequiringAttentionEndpoint(bl),
		EvaluateRateLifecycleEP:          makeEvaluateRateLifecycleEndpoint(bl),

		// Task 19.7: Rate Analytics & Procurement Intelligence
		GetRateAnalyticsOverviewEP:   makeGetRateAnalyticsOverviewEndpoint(bl),
		GetRateAnalyticsTrendsEP:     makeGetRateAnalyticsTrendsEndpoint(bl),
		GetCarrierRatePerformanceEP:  makeGetCarrierRatePerformanceEndpoint(bl),
		GetLaneRatePerformanceEP:     makeGetLaneRatePerformanceEndpoint(bl),
		GetRateLifecycleAnalyticsEP:  makeGetRateLifecycleAnalyticsEndpoint(bl),
		GetSpotSourcingPerformanceEP: makeGetSpotSourcingPerformanceEndpoint(bl),
		GetRateCommercialInsightsEP:  makeGetRateCommercialInsightsEndpoint(bl),
	}
}

func getOrgIDFromContext(ctx context.Context) (int64, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		return 0, svcerror.NewServiceError(svcerror.ErrInsufficientResourceAccess)
	}
	return userCtx.OrgID, nil
}

func getUserContext(ctx context.Context) (middleware.UserContext, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		return middleware.UserContext{}, svcerror.NewServiceError(svcerror.ErrInsufficientResourceAccess)
	}
	return userCtx, nil
}

func makeListRatesEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ListRatesRequest)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		resp, err := bl.ListRates(ctx, *req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

func makeGetRateSummaryEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		resp, err := bl.GetRateSummary(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

func makeCreateRateEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateRateRequest)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = userCtx.OrgID
		author := userCtx.CognitoID
		if author == "" {
			author = strconv.FormatInt(userCtx.UserID, 10)
		}
		req.Author = author

		resp, err := bl.CreateRate(ctx, *req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

func makeGetRateEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetRateRequest)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		if rateID, err := strconv.ParseInt(req.ID, 10, 64); err == nil {
			return bl.GetRate(ctx, orgID, rateID)
		}
		return bl.GetRateByID(ctx, orgID, req.ID)
	}
}

func makeUpdateRateEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateRateRequest)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = userCtx.OrgID
		updater := userCtx.CognitoID
		if updater == "" {
			updater = strconv.FormatInt(userCtx.UserID, 10)
		}
		req.Updater = updater

		resp, err := bl.UpdateRate(ctx, *req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

func makeArchiveRateEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ArchiveRateRequest)
		userCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = userCtx.OrgID
		user := userCtx.CognitoID
		if user == "" {
			user = strconv.FormatInt(userCtx.UserID, 10)
		}
		req.User = user

		if err := bl.ArchiveRate(ctx, *req); err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"id":     req.ID,
			"status": "ARCHIVED",
		}, nil
	}
}

func makeSearchRatesEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.RateQuery)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		resp, err := bl.SearchRates(ctx, *req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

func makeRefreshSpotRatesEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.RefreshSpotRatesRequest)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		q := spec.RateQuery{
			OrgID:           orgID,
			OriginPort:      req.Origin,
			DestinationPort: req.Destination,
			EquipmentType:   req.EquipmentType,
		}

		resp, err := bl.RefreshSpotRates(ctx, orgID, q, nil)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

// ── Task 19.2: Rate Charges Endpoints ────────────────────────────────────────

func makeGetRatePricingEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetRateRequest)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		rateID, err := strconv.ParseInt(req.ID, 10, 64)
		if err != nil {
			return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
		}

		pricing, err := bl.GetRatePricing(ctx, orgID, rateID)
		if err != nil {
			return nil, err
		}
		return spec.RatePricingResponse{Pricing: *pricing}, nil
	}
}

func makeAddRateChargeEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateRateChargeRequest)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		charge, pricing, err := bl.AddRateCharge(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"charge":  charge,
			"pricing": pricing,
		}, nil
	}
}

func makeUpdateRateChargeEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateRateChargeRequest)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		charge, pricing, err := bl.UpdateRateCharge(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"charge":  charge,
			"pricing": pricing,
		}, nil
	}
}

func makeDeleteRateChargeEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.DeleteRateChargeRequest)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		pricing, err := bl.DeleteRateCharge(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message": "Rate charge deleted successfully",
			"pricing": pricing,
		}, nil
	}
}

func makeReorderRateChargesEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ReorderRateChargesRequest)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		pricing, err := bl.ReorderRateCharges(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message": "Rate charges reordered successfully",
			"pricing": pricing,
		}, nil
	}
}

// ── Task 19.3: Contracts & Versions Endpoints ─────────────────────────────────

func makeListRateContractsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ListRateContractsRequest)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		res, err := bl.ListRateContracts(ctx, *req)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func makeGetRateContractSummaryEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		res, err := bl.GetRateContractSummary(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func makeCreateRateContractEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateRateContractRequest)
		uCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = uCtx.OrgID
		author := uCtx.CognitoID
		if author == "" {
			author = strconv.FormatInt(uCtx.UserID, 10)
		}
		req.Author = author

		contract, err := bl.CreateRateContract(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message":  "Rate contract created successfully",
			"contract": contract,
		}, nil
	}
}

func makeGetRateContractEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		contractID := request.(int64)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		contract, err := bl.GetRateContract(ctx, orgID, contractID)
		if err != nil {
			return nil, err
		}
		rates, _ := bl.GetRatesByContract(ctx, orgID, contractID)
		return map[string]interface{}{
			"contract":     contract,
			"linked_rates": rates,
		}, nil
	}
}

func makeUpdateRateContractEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateRateContractRequest)
		uCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = uCtx.OrgID
		updater := uCtx.CognitoID
		if updater == "" {
			updater = strconv.FormatInt(uCtx.UserID, 10)
		}
		req.Updater = updater

		contract, err := bl.UpdateRateContract(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message":  "Rate contract updated successfully",
			"contract": contract,
		}, nil
	}
}

func makeArchiveRateContractEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		contractID := request.(int64)
		uCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		user := uCtx.CognitoID
		if user == "" {
			user = strconv.FormatInt(uCtx.UserID, 10)
		}

		if err := bl.ArchiveRateContract(ctx, uCtx.OrgID, contractID, user); err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message": "Rate contract archived successfully",
		}, nil
	}
}

func makeRenewRateContractEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.RenewRateContractRequest)
		uCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = uCtx.OrgID
		user := uCtx.CognitoID
		if user == "" {
			user = strconv.FormatInt(uCtx.UserID, 10)
		}
		req.User = user

		contract, err := bl.RenewRateContract(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message":  "Rate contract renewed successfully",
			"contract": contract,
		}, nil
	}
}

func makeCreateRateVersionEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateRateVersionRequest)
		uCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = uCtx.OrgID
		user := uCtx.CognitoID
		if user == "" {
			user = strconv.FormatInt(uCtx.UserID, 10)
		}
		req.User = user

		rate, err := bl.CreateNewRateVersion(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message": "New rate version created successfully",
			"rate":    rate,
		}, nil
	}
}

func makeGetRateVersionsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		rateID := request.(int64)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		chain, err := bl.GetRateVersionChain(ctx, orgID, rateID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"versions": chain,
		}, nil
	}
}

func makeGetRateVersionHistoryEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		rateID := request.(int64)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		history, err := bl.GetRateVersionHistory(ctx, orgID, rateID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"history": history,
		}, nil
	}
}

// ── Task 19.4: Spot Rate Requests, Responses & Comparison Endpoints ──────────

func makeListSpotRateRequestsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ListSpotRateRequestsRequest)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		resp, err := bl.ListSpotRateRequests(ctx, *req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

func makeGetSpotRateSummaryEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		summary, err := bl.GetSpotRateSummary(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return summary, nil
	}
}

func makeCreateSpotRateRequestEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateSpotRateRequestRequest)
		uCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = uCtx.OrgID
		user := uCtx.CognitoID
		if user == "" {
			user = strconv.FormatInt(uCtx.UserID, 10)
		}
		req.User = user

		created, err := bl.CreateSpotRateRequest(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message": "Spot rate request created successfully",
			"request": created,
		}, nil
	}
}

func makeGetSpotRateRequestEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID := request.(int64)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		spotReq, err := bl.GetSpotRateRequest(ctx, orgID, reqID)
		if err != nil {
			return nil, err
		}
		return spotReq, nil
	}
}

func makeUpdateSpotRateRequestEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateSpotRateRequestRequest)
		uCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = uCtx.OrgID
		user := uCtx.CognitoID
		if user == "" {
			user = strconv.FormatInt(uCtx.UserID, 10)
		}
		req.User = user

		updated, err := bl.UpdateSpotRateRequest(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message": "Spot rate request updated successfully",
			"request": updated,
		}, nil
	}
}

func makeSendSpotRateRequestEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID := request.(int64)
		uCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		user := uCtx.CognitoID
		if user == "" {
			user = strconv.FormatInt(uCtx.UserID, 10)
		}

		sent, err := bl.SendSpotRateRequest(ctx, uCtx.OrgID, reqID, user)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message": "Spot rate request marked as sent to carriers",
			"request": sent,
		}, nil
	}
}

func makeCancelSpotRateRequestEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID := request.(int64)
		uCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		user := uCtx.CognitoID
		if user == "" {
			user = strconv.FormatInt(uCtx.UserID, 10)
		}

		if err := bl.CancelSpotRateRequest(ctx, uCtx.OrgID, reqID, user); err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message": "Spot rate request cancelled successfully",
		}, nil
	}
}

func makeCreateSpotRateResponseEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateSpotRateResponseRequest)
		uCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = uCtx.OrgID
		user := uCtx.CognitoID
		if user == "" {
			user = strconv.FormatInt(uCtx.UserID, 10)
		}
		req.User = user

		resp, err := bl.CreateSpotRateResponse(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message":  "Carrier spot rate response logged successfully",
			"response": resp,
		}, nil
	}
}

func makeGetSpotRateResponseEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		respID := request.(int64)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := bl.GetSpotRateResponse(ctx, orgID, respID)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
}

func makeGetSpotRateResponsesEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID := request.(int64)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		responses, err := bl.GetSpotRateResponses(ctx, orgID, reqID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"responses": responses,
		}, nil
	}
}

func makeUpdateSpotRateResponseEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateSpotRateResponseRequest)
		uCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = uCtx.OrgID
		user := uCtx.CognitoID
		if user == "" {
			user = strconv.FormatInt(uCtx.UserID, 10)
		}
		req.User = user

		resp, err := bl.UpdateSpotRateResponse(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message":  "Carrier response updated successfully",
			"response": resp,
		}, nil
	}
}

func makeSelectPreferredSpotRateEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.SelectPreferredSpotRateRequest)
		uCtx, err := getUserContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = uCtx.OrgID
		user := uCtx.CognitoID
		if user == "" {
			user = strconv.FormatInt(uCtx.UserID, 10)
		}
		req.User = user

		resp, err := bl.SelectPreferredSpotRate(ctx, *req)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message":  "Preferred spot rate selected successfully",
			"response": resp,
		}, nil
	}
}

func makeCompareSpotRatesEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID := request.(int64)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		comp, err := bl.CompareSpotRates(ctx, orgID, reqID)
		if err != nil {
			return nil, err
		}
		return comp, nil
	}
}

// ── Task 19.6: Rate Lifecycle Intelligence Endpoints ──────────────────────────

func makeGetRateLifecycleSummaryEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		summary, err := bl.GetRateLifecycleDashboard(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return summary, nil
	}
}

func makeGetRateLifecycleEventsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		limit := 50
		if l, ok := request.(int); ok && l > 0 {
			limit = l
		}
		events, err := bl.GetRateLifecycleEvents(ctx, orgID, limit)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"events": events,
		}, nil
	}
}

func makeGetRatesRequiringAttentionEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		ratesList, err := bl.GetRatesRequiringAttention(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"rates": ratesList,
		}, nil
	}
}

func makeGetContractsRequiringAttentionEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		contracts, err := bl.GetContractsRequiringAttention(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"contracts": contracts,
		}, nil
	}
}

func makeEvaluateRateLifecycleEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		summary, err := bl.EvaluateRateLifecycleForOrg(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"message": "Rate lifecycle evaluation completed successfully",
			"summary": summary,
		}, nil
	}
}

// ── Task 19.7: Rate Analytics Endpoint Builders ─────────────────────────────

func makeGetRateAnalyticsOverviewEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		overview, err := bl.GetRateAnalyticsOverview(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"overview": overview}, nil
	}
}

func makeGetRateAnalyticsTrendsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		days := 30
		if req, ok := request.(map[string]int); ok {
			if d, exists := req["days"]; exists {
				days = d
			}
		}
		trends, err := bl.GetRateAnalyticsTrends(ctx, orgID, days)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"trends": trends, "days": days}, nil
	}
}

func makeGetCarrierRatePerformanceEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		carriers, err := bl.GetCarrierRatePerformance(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"carriers": carriers, "count": len(carriers)}, nil
	}
}

func makeGetLaneRatePerformanceEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		lanes, err := bl.GetLaneRatePerformance(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"lanes": lanes, "count": len(lanes)}, nil
	}
}

func makeGetRateLifecycleAnalyticsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		analytics, err := bl.GetRateLifecycleAnalytics(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"lifecycle": analytics}, nil
	}
}

func makeGetSpotSourcingPerformanceEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		perf, err := bl.GetSpotSourcingPerformance(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"performance": perf}, nil
	}
}

func makeGetRateCommercialInsightsEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		insights, err := bl.GetRateCommercialInsights(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"insights": insights, "count": len(insights)}, nil
	}
}

func makeSearchCarrierLiveRatesEndpoint(bl BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CarrierRateSearchRequest)
		orgID, err := getOrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID
		return bl.SearchCarrierLiveRates(ctx, orgID, *req)
	}
}
