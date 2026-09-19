package dashboard

import (
	"context"
	"fmt"
	"strings"

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

	resp := &spec.GetMissionControlResponse{
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
	}

	// Authoritative Deduplication and Information Density Optimization
	deduplicateMissionControl(resp)

	return resp, nil
}

// deduplicateMissionControl ensures every section provides unique, high-density business value
// without repeating the exact same underlying entities across Priority Actions, Reminders, and Approvals.
func deduplicateMissionControl(resp *spec.GetMissionControlResponse) {
	if resp == nil {
		return
	}

	// 1. Build set of claimed entity IDs from Priority Actions
	claimedEntities := make(map[string]bool)
	for _, item := range resp.AttentionItems {
		if item.SourceEntityID > 0 {
			claimedEntities[fmt.Sprintf("%s:%d", strings.ToLower(item.Category), item.SourceEntityID)] = true
		}
		if item.ApprovalID > 0 {
			claimedEntities[fmt.Sprintf("approvals:%d", item.ApprovalID)] = true
		}
	}

	// 2. Deduplicate Upcoming Reminders:
	// Filter out reminders whose underlying record is already featured as an urgent Priority Action.
	if len(resp.UpcomingReminders) > 0 {
		var dedupedReminders []spec.UpcomingReminder
		seenReminderTypes := make(map[string]bool)

		for _, rem := range resp.UpcomingReminders {
			var entityKey string
			if strings.HasPrefix(rem.ID, "lead_rem_") {
				entityKey = "leads:" + strings.TrimPrefix(rem.ID, "lead_rem_")
			} else if strings.HasPrefix(rem.ID, "contract_rem_") {
				entityKey = "contracts:" + strings.TrimPrefix(rem.ID, "contract_rem_")
			} else if strings.HasPrefix(rem.ID, "inv_rem_") {
				entityKey = "finance:" + strings.TrimPrefix(rem.ID, "inv_rem_")
			}

			// If the underlying entity is already featured as an urgent Priority Action, omit it
			if entityKey != "" && claimedEntities[entityKey] {
				continue
			}

			// Keep at most 1 distinct reminder per type, up to 3 total
			if !seenReminderTypes[rem.Type] && len(dedupedReminders) < 3 {
				seenReminderTypes[rem.Type] = true
				dedupedReminders = append(dedupedReminders, rem)
			}
		}
		resp.UpcomingReminders = dedupedReminders
	}

	// 3. Deduplicate Pending Approvals in Finance column:
	// If the top pending approval is already featured in Priority Actions,
	// prioritize other pending approvals in the Finance & Approvals section so users see distinct tasks.
	if len(resp.PendingApprovals) > 1 {
		var otherApprovals []spec.PendingApprovalItem
		var priorityMatchedApprovals []spec.PendingApprovalItem

		for _, app := range resp.PendingApprovals {
			key := fmt.Sprintf("approvals:%d", app.ID)
			if claimedEntities[key] {
				priorityMatchedApprovals = append(priorityMatchedApprovals, app)
			} else {
				otherApprovals = append(otherApprovals, app)
			}
		}

		if len(otherApprovals) > 0 {
			resp.PendingApprovals = append(otherApprovals, priorityMatchedApprovals...)
		}
	}

	// 4. Deduplicate Recent Activity by event ID
	if len(resp.RecentActivity) > 1 {
		seenActivity := make(map[string]bool)
		var dedupedActivity []spec.RecentActivity
		for _, act := range resp.RecentActivity {
			if !seenActivity[act.ID] {
				seenActivity[act.ID] = true
				dedupedActivity = append(dedupedActivity, act)
			}
		}
		resp.RecentActivity = dedupedActivity
	}
}
