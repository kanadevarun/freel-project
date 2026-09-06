package shipments

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/shipments/spec"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	ListShipmentsEP           endpoint.Endpoint
	GetShipmentEP             endpoint.Endpoint
	CarrierUpdateEP           endpoint.Endpoint
	ResolveExceptionEP        endpoint.Endpoint
	UpdateMilestoneEP         endpoint.Endpoint
	InboundCarrierEmailEP     endpoint.Endpoint
	InboundWebhookEP          endpoint.Endpoint
	GetShipmentInternalEP     endpoint.Endpoint
	UpdateMilestoneInternalEP endpoint.Endpoint
	CreateExceptionInternalEP endpoint.Endpoint
	CallbackInternalEP        endpoint.Endpoint

	// Exception endpoints Task 16.5
	GetShipmentExceptionsEP      endpoint.Endpoint
	CreateShipmentExceptionEP    endpoint.Endpoint
	UpdateShipmentExceptionEP    endpoint.Endpoint
	AcknowledgeShipmentExceptionEP endpoint.Endpoint
	ResolveShipmentExceptionEP   endpoint.Endpoint
	DismissShipmentExceptionEP   endpoint.Endpoint
	EvaluateShipmentExceptionsEP endpoint.Endpoint

	// Tracking, Intelligence & Monitoring endpoints (Task 16.6, 17.3, 17.4, 17.5)
	GetShipmentTrackingEP             endpoint.Endpoint
	GetLatestTrackingPositionEP       endpoint.Endpoint
	GetTrackingPositionsEP            endpoint.Endpoint
	GetTrackingRouteEP                endpoint.Endpoint
	GetTrackingEventsEP               endpoint.Endpoint
	GetShipmentTrackingIntelligenceEP endpoint.Endpoint
	GetTrackingAlertsEP               endpoint.Endpoint
	GetTrackingMonitoringSummaryEP    endpoint.Endpoint
	RefreshShipmentTrackingEP         endpoint.Endpoint
	GetTrackingRefreshHistoryEP       endpoint.Endpoint
	AcknowledgeTrackingAlertEP        endpoint.Endpoint
	ResolveTrackingAlertEP            endpoint.Endpoint
	SuppressTrackingAlertEP           endpoint.Endpoint

	// Tracking Analytics & Intelligence (Task 17.8)
	GetTrackingAnalyticsOverviewEP    endpoint.Endpoint
	GetTrackingAnalyticsTrendsEP      endpoint.Endpoint
	GetCarrierTrackingPerformanceEP   endpoint.Endpoint
	GetRouteTrackingPerformanceEP     endpoint.Endpoint

	GetShipmentClosureEP              endpoint.Endpoint
	EvaluateShipmentClosureEP  endpoint.Endpoint
	RequestShipmentClosureEP   endpoint.Endpoint
	CompleteShipmentEP         endpoint.Endpoint
	ReopenShipmentEP           endpoint.Endpoint

	// Document endpoints Task 16.7
	GetShipmentDocumentsEP    endpoint.Endpoint
	CreateShipmentDocumentEP  endpoint.Endpoint
	UpdateShipmentDocumentEP  endpoint.Endpoint
	ApproveShipmentDocumentEP endpoint.Endpoint
	RejectShipmentDocumentEP  endpoint.Endpoint
	DeleteShipmentDocumentEP  endpoint.Endpoint

	// Financial endpoints Task 16.8
	GetShipmentFinancialsEP         endpoint.Endpoint
	GetShipmentChargesEP            endpoint.Endpoint
	CreateShipmentChargeEP          endpoint.Endpoint
	UpdateShipmentChargeEP          endpoint.Endpoint
	DeleteShipmentChargeEP          endpoint.Endpoint
	RecalculateShipmentFinancialsEP endpoint.Endpoint
	ReviewShipmentFinancialsEP      endpoint.Endpoint
}

func NewAllShipmentsEndpoints(svc Service) Endpoints {
	return Endpoints{
		ListShipmentsEP:           makeListShipmentsEP(svc),
		GetShipmentEP:             makeGetShipmentEP(svc),
		CarrierUpdateEP:             makeCarrierUpdateEP(svc),
		ResolveExceptionEP:        makeResolveExceptionEP(svc),
		UpdateMilestoneEP:         makeUpdateMilestoneEP(svc),
		InboundCarrierEmailEP:     makeInboundCarrierEmailEP(svc),
		InboundWebhookEP:          makeInboundWebhookEP(svc),
		GetShipmentInternalEP:     makeGetShipmentInternalEP(svc),
		UpdateMilestoneInternalEP: makeUpdateMilestoneInternalEP(svc),
		CreateExceptionInternalEP: makeCreateExceptionInternalEP(svc),
		CallbackInternalEP:        makeCallbackInternalEP(svc),

		GetShipmentExceptionsEP:      makeGetShipmentExceptionsEP(svc),
		CreateShipmentExceptionEP:    makeCreateShipmentExceptionEP(svc),
		UpdateShipmentExceptionEP:    makeUpdateShipmentExceptionEP(svc),
		AcknowledgeShipmentExceptionEP: makeAcknowledgeShipmentExceptionEP(svc),
		ResolveShipmentExceptionEP:   makeResolveShipmentExceptionEP(svc),
		DismissShipmentExceptionEP:   makeDismissShipmentExceptionEP(svc),
		EvaluateShipmentExceptionsEP: makeEvaluateShipmentExceptionsEP(svc),

		GetShipmentTrackingEP:             makeGetShipmentTrackingEP(svc),
		GetLatestTrackingPositionEP:        makeGetLatestTrackingPositionEP(svc),
		GetTrackingPositionsEP:            makeGetTrackingPositionsEP(svc),
		GetTrackingRouteEP:                makeGetTrackingRouteEP(svc),
		GetTrackingEventsEP:               makeGetTrackingEventsEP(svc),
		GetShipmentTrackingIntelligenceEP: makeGetShipmentTrackingIntelligenceEP(svc),
		GetTrackingAlertsEP:               makeGetTrackingAlertsEP(svc),
		GetTrackingMonitoringSummaryEP:    makeGetTrackingMonitoringSummaryEP(svc),
		RefreshShipmentTrackingEP:         makeRefreshShipmentTrackingEP(svc),
		GetTrackingRefreshHistoryEP:       makeGetTrackingRefreshHistoryEP(svc),
		AcknowledgeTrackingAlertEP:        makeAcknowledgeTrackingAlertEP(svc),
		ResolveTrackingAlertEP:            makeResolveTrackingAlertEP(svc),
		SuppressTrackingAlertEP:           makeSuppressTrackingAlertEP(svc),

		GetTrackingAnalyticsOverviewEP:    makeGetTrackingAnalyticsOverviewEP(svc),
		GetTrackingAnalyticsTrendsEP:      makeGetTrackingAnalyticsTrendsEP(svc),
		GetCarrierTrackingPerformanceEP:   makeGetCarrierTrackingPerformanceEP(svc),
		GetRouteTrackingPerformanceEP:     makeGetRouteTrackingPerformanceEP(svc),

		GetShipmentClosureEP:              makeGetShipmentClosureEP(svc),
		EvaluateShipmentClosureEP:  makeEvaluateShipmentClosureEP(svc),
		RequestShipmentClosureEP:   makeRequestShipmentClosureEP(svc),
		CompleteShipmentEP:         makeCompleteShipmentEP(svc),
		ReopenShipmentEP:           makeReopenShipmentEP(svc),

		GetShipmentDocumentsEP:    makeGetShipmentDocumentsEP(svc),
		CreateShipmentDocumentEP:  makeCreateShipmentDocumentEP(svc),
		UpdateShipmentDocumentEP:  makeUpdateShipmentDocumentEP(svc),
		ApproveShipmentDocumentEP: makeApproveShipmentDocumentEP(svc),
		RejectShipmentDocumentEP:  makeRejectShipmentDocumentEP(svc),
		DeleteShipmentDocumentEP:  makeDeleteShipmentDocumentEP(svc),

		GetShipmentFinancialsEP:         makeGetShipmentFinancialsEP(svc),
		GetShipmentChargesEP:            makeGetShipmentChargesEP(svc),
		CreateShipmentChargeEP:          makeCreateShipmentChargeEP(svc),
		UpdateShipmentChargeEP:          makeUpdateShipmentChargeEP(svc),
		DeleteShipmentChargeEP:          makeDeleteShipmentChargeEP(svc),
		RecalculateShipmentFinancialsEP: makeRecalculateShipmentFinancialsEP(svc),
		ReviewShipmentFinancialsEP:      makeReviewShipmentFinancialsEP(svc),
	}
}

func getOrgID(ctx context.Context) (int64, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		return 0, svcerror.NewServiceError(svcerror.ErrInsufficientResourceAccess)
	}
	return userCtx.OrgID, nil
}

func makeListShipmentsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ListShipmentsRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		// Legacy mode check: if no workspace params or search/status filter, fall back to legacy raw list
		if req.Page <= 0 && req.Limit <= 0 && !req.Workspace && (req.Status == nil || *req.Status == "") && (req.Search == nil || *req.Search == "") {
			list, err := svc.ListShipments(ctx, req.OrgID)
			if err != nil {
				return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
			}
			return &spec.APIResponse{
				Success: true,
				Message: "Shipments retrieved",
				Data:    list,
			}, nil
		}

		limit := req.Limit
		if limit <= 0 {
			limit = 10
		}
		page := req.Page
		if page <= 0 {
			page = 1
		}

		filter := spec.ShipmentListFilter{
			Page:  page,
			Limit: limit,
		}
		if req.Status != nil {
			filter.Status = req.Status
		}
		if req.Search != nil {
			filter.Search = req.Search
		}

		list, kpis, totalItems, err := svc.GetShipmentsWorkspace(ctx, req.OrgID, filter)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		totalPages := (totalItems + limit - 1) / limit
		if totalPages < 1 {
			totalPages = 1
		}

		workspaceData := spec.ShipmentWorkspaceData{
			Shipments: list,
			KPIs:      kpis,
			Pagination: spec.Pagination{
				CurrentPage: page,
				PageSize:    limit,
				TotalItems:  totalItems,
				TotalPages:  totalPages,
			},
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Shipments workspace retrieved",
			Data:    workspaceData,
		}, nil
	}
}

func makeGetShipmentEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		sh, err := svc.GetShipmentByID(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}
		if sh == nil {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}

		milestones, err := svc.GetMilestones(ctx, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		exceptions, err := svc.GetShipmentExceptions(ctx, orgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Shipment details retrieved",
			Data: spec.ShipmentDetailData{
				Shipment:   sh,
				Milestones: milestones,
				Exceptions: exceptions,
			},
		}, nil
	}
}

func makeCarrierUpdateEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CarrierUpdateRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		sh, err := svc.GetShipmentByID(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}
		if sh == nil {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}

		bookingNum := ""
		if sh.BookingNumber != nil {
			bookingNum = *sh.BookingNumber
		}
		event := &spec.NormalizedTrackingEvent{
			EventID:       req.EventID,
			SourceType:    "MANUAL",
			CarrierSCAC:   sh.CarrierSCAC,
			BookingNumber: bookingNum,
			EventTime:     time.Now(),
			Description:   req.Description,
			RawPayload:    json.RawMessage([]byte(fmt.Sprintf(`{"manual_description": %q}`, req.Description))),
		}

		err = svc.HandleInboundCarrierEvent(ctx, req.OrgID, event)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Carrier event enqueued for processing",
			Data:    map[string]interface{}{"event_id": req.EventID},
		}, nil
	}
}

func makeResolveExceptionEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ResolveExceptionRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		err = svc.ResolveException(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Exception resolved successfully",
		}, nil
	}
}

func makeInboundCarrierEmailEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.InboundCarrierEmailRequest)

		normalized, err := ParseCarrierEmail(&spec.CarrierEmailRequest{
			From:      req.From,
			To:        req.To,
			Subject:   req.Subject,
			Body:      req.Body,
			MessageID: req.MessageID,
		})
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		err = svc.HandleInboundCarrierEvent(ctx, req.OrgID, normalized)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Carrier email processed and enqueued",
			Data:    map[string]string{"event_id": normalized.EventID},
		}, nil
	}
}

func makeInboundWebhookEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.InboundWebhookRequest)

		// 1. Resolve WebhookProvider from CarrierAdapterFactory
		sImpl := svc.(*serviceImpl)
		adapter, err := adapters.GetWebhookProvider(sImpl.db, req.OrgID, req.Carrier)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		// 2. Verify webhook authenticity
		if err := adapter.VerifyWebhookSignature(req.Body, req.Headers); err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInsufficientResourceAccess, err)
		}

		// 3. Parse payload into generic TrackingEvent struct
		rawEv, err := adapter.ParseWebhookPayload(req.Body)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		// 4. Convert raw TrackingEvent to system NormalizedTrackingEvent contract
		normalized := Normalize(*rawEv, req.Carrier, "WEBHOOK")

		err = svc.HandleInboundCarrierEvent(ctx, req.OrgID, &normalized)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "carrier event received",
			Data:    map[string]string{"event_id": normalized.EventID},
		}, nil
	}
}

func makeGetShipmentInternalEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentInternalRequest)

		sh, err := svc.GetShipmentByID(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}
		if sh == nil {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}

		milestones, err := svc.GetMilestones(ctx, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Internal shipment details retrieved",
			Data: spec.InternalShipmentDetailData{
				Shipment:   sh,
				Milestones: milestones,
			},
		}, nil
	}
}

func makeUpdateMilestoneInternalEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateMilestoneInternalRequest)

		err := svc.UpdateMilestone(ctx, *req.OrgID, req.ID, req.MilestoneCode, &req.ActualDate, req.Location, req.Notes)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Milestone updated",
		}, nil
	}
}

func makeCreateExceptionInternalEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateExceptionInternalRequest)

		err := svc.CreateShipmentException(ctx, *req.OrgID, req.ID, req.ExceptionType, req.Severity, req.Title, req.Description, req.SourceEventID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Exception created",
		}, nil
	}
}

func makeCallbackInternalEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CallbackInternalRequest)

		if req.EventID != "" {
			_ = svc.CompleteCarrierEvent(ctx, req.EventID, req.OrgID, req.ShipmentID, req.HasCriticalException, req.AISummary)
		}

		log.Printf("[Shipment Callback] Processing agent results for Shipment #%d. Critical Exception: %t, Summary: %s",
			req.ShipmentID, req.HasCriticalException, req.AISummary)

		return &spec.APIResponse{
			Success: true,
			Message: "Operations callback processed successfully",
		}, nil
	}
}

func makeUpdateMilestoneEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateMilestoneRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		err = svc.UpdateMilestone(ctx, req.OrgID, req.ID, req.MilestoneCode, &req.ActualDate, req.Location, req.Notes)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Milestone updated successfully",
		}, nil
	}
}

func makeGetShipmentExceptionsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.EvaluateExceptionsRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		exceptions, err := svc.GetShipmentExceptions(ctx, orgID, req.ShipmentID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Data:    exceptions,
		}, nil
	}
}

func makeCreateShipmentExceptionEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateExceptionRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		err = svc.CreateShipmentException(ctx, orgID, req.ShipmentID, req.ExceptionType, req.Severity, req.Title, req.Description, req.SourceEventID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Exception created successfully",
		}, nil
	}
}

func makeUpdateShipmentExceptionEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateExceptionRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		err = svc.UpdateShipmentException(ctx, orgID, req.ShipmentID, req.ID, req.Status, req.Severity, req.Notes)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Exception updated successfully",
		}, nil
	}
}

func makeAcknowledgeShipmentExceptionEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.AcknowledgeExceptionRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		err = svc.AcknowledgeShipmentException(ctx, orgID, req.ShipmentID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Exception acknowledged successfully",
		}, nil
	}
}

func makeResolveShipmentExceptionEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ResolveShipmentExceptionRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		userCtx, _ := middleware.GetUserContext(ctx)
		req.ResolvedBy = userCtx.UserID

		err = svc.ResolveShipmentException(ctx, orgID, req.ShipmentID, req.ID, req.ResolutionNotes, req.ResolvedBy)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Exception resolved successfully",
		}, nil
	}
}

func makeDismissShipmentExceptionEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.DismissExceptionRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		err = svc.DismissShipmentException(ctx, orgID, req.ShipmentID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Exception dismissed successfully",
		}, nil
	}
}

func makeEvaluateShipmentExceptionsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.EvaluateExceptionsRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		err = svc.EvaluateShipmentExceptions(ctx, orgID, req.ShipmentID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Exceptions evaluated successfully",
		}, nil
	}
}

func makeGetShipmentTrackingEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		summary, err := svc.GetShipmentTracking(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Tracking summary retrieved",
			Data:    summary,
		}, nil
	}
}

func makeGetLatestTrackingPositionEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		pos, err := svc.GetLatestTrackingPosition(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Latest tracking position retrieved",
			Data:    pos,
		}, nil
	}
}

func makeGetTrackingPositionsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetTrackingPositionsRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		positions, err := svc.GetTrackingPositionHistory(ctx, req.OrgID, req.ID, req.Limit)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Tracking position history retrieved",
			Data:    positions,
		}, nil
	}
}

func makeGetTrackingRouteEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		route, err := svc.GetTrackingRoute(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Tracking route retrieved",
			Data:    route,
		}, nil
	}
}

func makeGetTrackingEventsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		eventsList, err := svc.GetTrackingEventsList(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Tracking events retrieved",
			Data:    eventsList,
		}, nil
	}
}

func makeGetShipmentTrackingIntelligenceEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		intelligence, err := svc.GetShipmentTrackingIntelligence(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Shipment tracking intelligence retrieved",
			Data:    intelligence,
		}, nil
	}
}

func makeGetTrackingAlertsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetTrackingAlertsRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		alerts, err := svc.GetTrackingAlerts(ctx, req.OrgID, req.ShipmentID, req.Status)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Tracking alerts retrieved",
			Data:    alerts,
		}, nil
	}
}

func makeGetTrackingMonitoringSummaryEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		summary, err := svc.GetTrackingMonitoringSummary(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Tracking monitoring summary retrieved",
			Data:    summary,
		}, nil
	}
}

func makeRefreshShipmentTrackingEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		actor := "Operations Team"
		var userID *int64
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.UserID > 0 {
			userID = &userCtx.UserID
			if userCtx.Role != "" {
				actor = fmt.Sprintf("User #%d (%s)", userCtx.UserID, userCtx.Role)
			} else {
				actor = fmt.Sprintf("User #%d", userCtx.UserID)
			}
		}

		refreshRes, err := svc.RefreshShipmentTracking(ctx, req.OrgID, req.ID, userID, actor)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: refreshRes.Success,
			Message: refreshRes.Message,
			Data:    refreshRes,
		}, nil
	}
}

func makeGetTrackingRefreshHistoryEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		limit := 20
		history, err := svc.GetTrackingRefreshHistory(ctx, req.OrgID, req.ID, limit)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Tracking refresh history retrieved successfully",
			Data:    history,
		}, nil
	}
}

func makeAcknowledgeTrackingAlertEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.AcknowledgeTrackingAlertRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		actor := "Operations Team"
		var userID *int64
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.UserID > 0 {
			userID = &userCtx.UserID
			if userCtx.Role != "" {
				actor = fmt.Sprintf("User #%d (%s)", userCtx.UserID, userCtx.Role)
			} else {
				actor = fmt.Sprintf("User #%d", userCtx.UserID)
			}
		}

		if err := svc.AcknowledgeTrackingAlert(ctx, orgID, req.ShipmentID, req.AlertID, userID, actor); err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Tracking alert acknowledged",
		}, nil
	}
}

func makeResolveTrackingAlertEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ResolveTrackingAlertRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		actor := "Operations Team"
		var userID *int64
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.UserID > 0 {
			userID = &userCtx.UserID
			if userCtx.Role != "" {
				actor = fmt.Sprintf("User #%d (%s)", userCtx.UserID, userCtx.Role)
			} else {
				actor = fmt.Sprintf("User #%d", userCtx.UserID)
			}
		}

		if err := svc.ResolveTrackingAlert(ctx, orgID, req.ShipmentID, req.AlertID, userID, req.ResolutionNotes, actor); err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Tracking alert resolved",
		}, nil
	}
}

func makeSuppressTrackingAlertEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.SuppressTrackingAlertRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		actor := "Operations Team"
		var userID *int64
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.UserID > 0 {
			userID = &userCtx.UserID
			if userCtx.Role != "" {
				actor = fmt.Sprintf("User #%d (%s)", userCtx.UserID, userCtx.Role)
			} else {
				actor = fmt.Sprintf("User #%d", userCtx.UserID)
			}
		}

		if err := svc.SuppressTrackingAlert(ctx, orgID, req.ShipmentID, req.AlertID, userID, req.Reason, actor); err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Tracking alert suppressed",
		}, nil
	}
}

func makeGetShipmentClosureEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		closureStatus, err := svc.EvaluateClosure(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Closure status retrieved",
			Data:    map[string]string{"closure_status": closureStatus},
		}, nil
	}
}

func makeEvaluateShipmentClosureEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		closureStatus, err := svc.EvaluateClosure(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Closure evaluated successfully",
			Data:    map[string]string{"closure_status": closureStatus},
		}, nil
	}
}

func makeRequestShipmentClosureEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		err = svc.RequestClosure(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Closure requested successfully",
		}, nil
	}
}

func makeCompleteShipmentEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		err = svc.CompleteShipment(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Shipment completed successfully",
		}, nil
	}
}

func makeReopenShipmentEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		err = svc.ReopenShipment(ctx, req.OrgID, req.ID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Shipment reopened successfully",
		}, nil
	}
}

// ─── Document Endpoints (Task 16.7) ──────────────────────────────────────────

func makeGetShipmentDocumentsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentDocumentsRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		docs, compliance, discrepancies, err := svc.GetShipmentDocuments(ctx, req.OrgID, req.ShipmentID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Shipment documents retrieved successfully",
			Data: map[string]interface{}{
				"documents":          docs,
				"compliance_summary": compliance,
				"discrepancies":      discrepancies,
			},
		}, nil
	}
}

func makeCreateShipmentDocumentEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateShipmentDocumentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		uploader := "Operations Team"
		var userID *int64
		if userCtx, ok := middleware.GetUserContext(ctx); ok {
			if userCtx.UserID > 0 {
				uid := userCtx.UserID
				userID = &uid
				if userCtx.Role != "" {
					uploader = fmt.Sprintf("User #%d (%s)", userCtx.UserID, userCtx.Role)
				} else {
					uploader = fmt.Sprintf("User #%d", userCtx.UserID)
				}
			}
		}

		doc, err := svc.CreateShipmentDocument(ctx, req.OrgID, req.ShipmentID, *req, uploader, userID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Document uploaded and recorded successfully",
			Data:    doc,
		}, nil
	}
}

func makeUpdateShipmentDocumentEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateShipmentDocumentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		reviewer := "Operations Team"
		var userID *int64
		if userCtx, ok := middleware.GetUserContext(ctx); ok {
			if userCtx.UserID > 0 {
				uid := userCtx.UserID
				userID = &uid
				if userCtx.Role != "" {
					reviewer = fmt.Sprintf("User #%d (%s)", userCtx.UserID, userCtx.Role)
				} else {
					reviewer = fmt.Sprintf("User #%d", userCtx.UserID)
				}
			}
		}

		doc, err := svc.UpdateShipmentDocument(ctx, req.OrgID, req.ShipmentID, req.ID, *req, reviewer, userID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Document updated successfully",
			Data:    doc,
		}, nil
	}
}

func makeApproveShipmentDocumentEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ApproveShipmentDocumentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		reviewer := "Compliance Reviewer"
		var userID *int64
		if userCtx, ok := middleware.GetUserContext(ctx); ok {
			if userCtx.UserID > 0 {
				uid := userCtx.UserID
				userID = &uid
				if userCtx.Role != "" {
					reviewer = fmt.Sprintf("User #%d (%s)", userCtx.UserID, userCtx.Role)
				} else {
					reviewer = fmt.Sprintf("User #%d", userCtx.UserID)
				}
			}
		}

		doc, err := svc.ApproveShipmentDocument(ctx, req.OrgID, req.ShipmentID, req.ID, reviewer, userID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Document approved successfully",
			Data:    doc,
		}, nil
	}
}

func makeRejectShipmentDocumentEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.RejectShipmentDocumentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		reviewer := "Compliance Reviewer"
		var userID *int64
		if userCtx, ok := middleware.GetUserContext(ctx); ok {
			if userCtx.UserID > 0 {
				uid := userCtx.UserID
				userID = &uid
				if userCtx.Role != "" {
					reviewer = fmt.Sprintf("User #%d (%s)", userCtx.UserID, userCtx.Role)
				} else {
					reviewer = fmt.Sprintf("User #%d", userCtx.UserID)
				}
			}
		}

		doc, err := svc.RejectShipmentDocument(ctx, req.OrgID, req.ShipmentID, req.ID, req.RejectionReason, reviewer, userID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Document rejected",
			Data:    doc,
		}, nil
	}
}

func makeDeleteShipmentDocumentEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.DeleteShipmentDocumentRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		var userID *int64
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.UserID > 0 {
			uid := userCtx.UserID
			userID = &uid
		}

		err = svc.DeleteShipmentDocument(ctx, req.OrgID, req.ShipmentID, req.ID, userID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Document deleted successfully",
		}, nil
	}
}

// ─── Financial Operations Endpoints (Task 16.8) ─────────────────────────────

func makeGetShipmentFinancialsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		shipmentID := request.(int64)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		summary, err := svc.GetShipmentFinancials(ctx, orgID, shipmentID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Shipment financials retrieved successfully",
			Data:    summary,
		}, nil
	}
}

func makeGetShipmentChargesEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.GetShipmentChargesRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		charges, err := svc.GetShipmentCharges(ctx, req.OrgID, req.ShipmentID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Shipment charges retrieved successfully",
			Data:    charges,
		}, nil
	}
}

func makeCreateShipmentChargeEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.CreateShipmentChargeRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		actor := "Operations Team"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.UserID > 0 {
			if userCtx.Role != "" {
				actor = fmt.Sprintf("User #%d (%s)", userCtx.UserID, userCtx.Role)
			} else {
				actor = fmt.Sprintf("User #%d", userCtx.UserID)
			}
		}

		charge, summary, err := svc.CreateShipmentCharge(ctx, req.OrgID, req.ShipmentID, req, actor)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Shipment charge line item added successfully",
			Data: map[string]interface{}{
				"charge":  charge,
				"summary": summary,
			},
		}, nil
	}
}

func makeUpdateShipmentChargeEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.UpdateShipmentChargeRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		actor := "Operations Team"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.UserID > 0 {
			if userCtx.Role != "" {
				actor = fmt.Sprintf("User #%d (%s)", userCtx.UserID, userCtx.Role)
			} else {
				actor = fmt.Sprintf("User #%d", userCtx.UserID)
			}
		}

		charge, summary, err := svc.UpdateShipmentCharge(ctx, req.OrgID, req.ShipmentID, req.ID, req, actor)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Shipment charge line item updated successfully",
			Data: map[string]interface{}{
				"charge":  charge,
				"summary": summary,
			},
		}, nil
	}
}

func makeDeleteShipmentChargeEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.DeleteShipmentChargeRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		actor := "Operations Team"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.UserID > 0 {
			if userCtx.Role != "" {
				actor = fmt.Sprintf("User #%d (%s)", userCtx.UserID, userCtx.Role)
			} else {
				actor = fmt.Sprintf("User #%d", userCtx.UserID)
			}
		}

		summary, err := svc.DeleteShipmentCharge(ctx, req.OrgID, req.ShipmentID, req.ID, actor)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Shipment charge line item deleted successfully",
			Data:    summary,
		}, nil
	}
}

func makeRecalculateShipmentFinancialsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		shipmentID := request.(int64)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		actor := "Operations Team"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.UserID > 0 {
			if userCtx.Role != "" {
				actor = fmt.Sprintf("User #%d (%s)", userCtx.UserID, userCtx.Role)
			} else {
				actor = fmt.Sprintf("User #%d", userCtx.UserID)
			}
		}

		summary, err := svc.RecalculateShipmentFinancials(ctx, orgID, shipmentID, actor)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Shipment financials recalculated successfully",
			Data:    summary,
		}, nil
	}
}

func makeReviewShipmentFinancialsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*spec.ReviewShipmentFinancialsRequest)
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}
		req.OrgID = orgID

		actor := "Finance Manager"
		if userCtx, ok := middleware.GetUserContext(ctx); ok && userCtx.UserID > 0 {
			if userCtx.Role != "" {
				actor = fmt.Sprintf("User #%d (%s)", userCtx.UserID, userCtx.Role)
			} else {
				actor = fmt.Sprintf("User #%d", userCtx.UserID)
			}
		}

		summary, err := svc.ReviewShipmentFinancials(ctx, req.OrgID, req.ShipmentID, req.FinancialStatus, req.Notes, actor)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Shipment financial review recorded successfully",
			Data:    summary,
		}, nil
	}
}

// ─── Tracking Analytics & Performance Intelligence Endpoint Makers (Task 17.8) ─

type GetTrackingAnalyticsTrendsRequest struct {
	Days int `json:"days"`
}

func makeGetTrackingAnalyticsOverviewEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		overview, err := svc.GetTrackingAnalyticsOverview(ctx, orgID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Tracking analytics overview retrieved successfully",
			Data:    overview,
		}, nil
	}
}

func makeGetTrackingAnalyticsTrendsEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		days := 14
		if req, ok := request.(*GetTrackingAnalyticsTrendsRequest); ok && req.Days > 0 {
			days = req.Days
		}

		trends, err := svc.GetTrackingAnalyticsTrends(ctx, orgID, days)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Tracking analytics trends retrieved successfully",
			Data:    trends,
		}, nil
	}
}

func makeGetCarrierTrackingPerformanceEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		carriers, err := svc.GetCarrierTrackingPerformance(ctx, orgID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Carrier tracking performance retrieved successfully",
			Data:    carriers,
		}, nil
	}
}

func makeGetRouteTrackingPerformanceEP(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orgID, err := getOrgID(ctx)
		if err != nil {
			return nil, err
		}

		routes, err := svc.GetRouteTrackingPerformance(ctx, orgID)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}

		return &spec.APIResponse{
			Success: true,
			Message: "Route tracking performance retrieved successfully",
			Data:    routes,
		}, nil
	}
}


