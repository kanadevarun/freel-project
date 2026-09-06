package dashboard

import (
	"context"

	"github.com/freel/backend/internal/dashboard/spec"
	"github.com/freel/backend/internal/svcerror"
)

type BusinessLogic interface {
	GetMissionControl(ctx context.Context, orgID int64, startDate, endDate, preset string) (*spec.GetMissionControlResponse, error)
}

type businessLogic struct {
	dl Datalayer
}

func NewBusinessLogic(dl Datalayer) BusinessLogic {
	return &businessLogic{dl: dl}
}

func (b *businessLogic) GetMissionControl(ctx context.Context, orgID int64, startDate, endDate, preset string) (*spec.GetMissionControlResponse, error) {
	if orgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	stats, pipeline, shipmentStatus, invoiceSummary, moduleStatus, dateRangeInfo, err := b.dl.GetStats(ctx, orgID, startDate, endDate, preset)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	queue, pendingApprovals, err := b.dl.GetApprovalQueue(ctx, orgID)
	if err != nil {
		queue = []spec.PendingTask{}
		pendingApprovals = []spec.PendingApprovalItem{}
	}
	if queue == nil {
		queue = []spec.PendingTask{}
	}
	if pendingApprovals == nil {
		pendingApprovals = []spec.PendingApprovalItem{}
	}

	attentionItems, err := b.dl.GetAttentionItems(ctx, orgID)
	if err != nil || attentionItems == nil {
		attentionItems = []spec.AttentionItem{}
	}

	recentShipments, activeShipments, err := b.dl.GetActiveShipments(ctx, orgID)
	if err != nil || recentShipments == nil {
		recentShipments = []spec.ActiveShipment{}
		activeShipments = []spec.ActiveShipmentItem{}
	}

	recentDocuments, err := b.dl.GetRecentDocuments(ctx, orgID)
	if err != nil || recentDocuments == nil {
		recentDocuments = []spec.RecentDocument{}
	}

	recentActivity, err := b.dl.GetRecentActivity(ctx, orgID)
	if err != nil || recentActivity == nil {
		recentActivity = []spec.RecentActivity{}
	}

	upcomingReminders, err := b.dl.GetUpcomingReminders(ctx, orgID)
	if err != nil || upcomingReminders == nil {
		upcomingReminders = []spec.UpcomingReminder{}
	}

	orgInfo, err := b.dl.GetOrganizationInfo(ctx, orgID)
	if err != nil {
		orgInfo = spec.OrganizationInfo{ID: orgID, Name: "Freight Forwarder", DefaultCurrency: "USD", DefaultTimezone: "UTC"}
	}

	aiStatus := spec.AIStatus{
		ActiveAgents:  1,
		TasksFinished: len(queue) + stats.TotalRFQs,
		HealthScore:   98,
	}

	return &spec.GetMissionControlResponse{
		Stats:             stats,
		Pipeline:          pipeline,
		ShipmentStatus:    shipmentStatus,
		InvoiceSummary:    invoiceSummary,
		ApprovalQueue:     queue,
		PendingApprovals:  pendingApprovals,
		AttentionItems:    attentionItems,
		RecentShipments:   recentShipments,
		ActiveShipments:   activeShipments,
		RecentDocuments:   recentDocuments,
		RecentActivity:    recentActivity,
		UpcomingReminders: upcomingReminders,
		ModuleStatus:      moduleStatus,
		Organization:      orgInfo,
		AIStatus:          aiStatus,
		DateRange:         dateRangeInfo,
	}, nil
}
