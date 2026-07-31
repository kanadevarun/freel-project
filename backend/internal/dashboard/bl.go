package dashboard

import (
	"context"

	"github.com/freel/backend/internal/dashboard/spec"
	"github.com/freel/backend/internal/svcerror"
)

type BusinessLogic interface {
	GetMissionControl(ctx context.Context, orgID int32) (*spec.GetMissionControlResponse, error)
}

type businessLogic struct {
	dl Datalayer
}

func NewBusinessLogic(dl Datalayer) BusinessLogic {
	return &businessLogic{dl: dl}
}

func (b *businessLogic) GetMissionControl(ctx context.Context, orgID int32) (*spec.GetMissionControlResponse, error) {
	if orgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	stats, err := b.dl.GetStats(ctx, orgID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	queue, err := b.dl.GetApprovalQueue(ctx, orgID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	if queue == nil {
		queue = []spec.PendingTask{}
	}

	aiStatus := spec.AIStatus{
		ActiveAgents:  1,
		TasksFinished: len(queue), // Mock
		HealthScore:   98,         // Mock
	}

	return &spec.GetMissionControlResponse{
		Stats:         stats,
		ApprovalQueue: queue,
		AIStatus:      aiStatus,
	}, nil
}
