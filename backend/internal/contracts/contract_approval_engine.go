package contracts

import (
	"context"
	"fmt"
	"time"

	"github.com/freel/backend/internal/svcerror"
)

// ContractApprovalEngine coordinates approvals for Versions and Amendments.
type ContractApprovalEngine struct {
	dl DataLayer
	ae *ContractAmendmentEngine
	ve *ContractVersioningEngine
}

func NewContractApprovalEngine(dl DataLayer, ae *ContractAmendmentEngine, ve *ContractVersioningEngine) *ContractApprovalEngine {
	return &ContractApprovalEngine{
		dl: dl,
		ae: ae,
		ve: ve,
	}
}

// ApproveRequest records an approval decision and advances the target amendment or version.
func (e *ContractApprovalEngine) ApproveRequest(
	ctx context.Context,
	orgID int64,
	contractID int64,
	approvalID string,
	req *ApproveContractRequest,
	userID string,
) (*ContractApprovalRequest, error) {
	appReq, err := e.dl.GetApprovalRequestByID(ctx, orgID, approvalID)
	if err != nil {
		return nil, err
	}

	if appReq.Status != ApprovalStatusPending {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var comment *string
	if req != nil && req.Comment != "" {
		comment = &req.Comment
	}

	// 1. Update approval request record
	err = e.dl.UpdateApprovalDecision(ctx, orgID, approvalID, ApprovalStatusApproved, userID, comment)
	if err != nil {
		return nil, err
	}

	// 2. Advance target entity
	if appReq.ApprovalType == ApprovalTypeAmendment && appReq.AmendmentID != nil {
		_ = e.dl.UpdateAmendmentStatus(ctx, orgID, *appReq.AmendmentID, AmendmentStatusApproved)
	} else if appReq.ApprovalType == ApprovalTypeVersion && appReq.VersionID != nil {
		_ = e.dl.UpdateVersionStatus(ctx, orgID, *appReq.VersionID, VersionStatusApproved, &userID)
	}

	// 3. Log audit event
	now := time.Now()
	_ = e.dl.LogLifecycleEvent(ctx, &ContractLifecycleEvent{
		OrgID:       orgID,
		ContractID:  contractID,
		NewStatus:   string(ContractStatusActive),
		EventType:   EventType(EventContractUpdated),
		Description: ptrString(fmt.Sprintf("Approval request %s approved by %s", approvalID, userID)),
		PerformedBy: &userID,
		CreatedAt:   now,
	})

	appReq.Status = ApprovalStatusApproved
	appReq.DecisionBy = &userID
	appReq.DecisionComment = comment
	return appReq, nil
}

// RejectRequest marks the approval as rejected and moves the target entity to REJECTED.
func (e *ContractApprovalEngine) RejectRequest(
	ctx context.Context,
	orgID int64,
	contractID int64,
	approvalID string,
	req *RejectContractRequest,
	userID string,
) (*ContractApprovalRequest, error) {
	appReq, err := e.dl.GetApprovalRequestByID(ctx, orgID, approvalID)
	if err != nil {
		return nil, err
	}

	if appReq.Status != ApprovalStatusPending {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var reason *string
	if req != nil && req.Reason != "" {
		reason = &req.Reason
	}

	err = e.dl.UpdateApprovalDecision(ctx, orgID, approvalID, ApprovalStatusRejected, userID, reason)
	if err != nil {
		return nil, err
	}

	if appReq.ApprovalType == ApprovalTypeAmendment && appReq.AmendmentID != nil {
		_ = e.dl.UpdateAmendmentStatus(ctx, orgID, *appReq.AmendmentID, AmendmentStatusRejected)
	} else if appReq.ApprovalType == ApprovalTypeVersion && appReq.VersionID != nil {
		_ = e.dl.UpdateVersionStatus(ctx, orgID, *appReq.VersionID, VersionStatusRejected, &userID)
	}

	now := time.Now()
	_ = e.dl.LogLifecycleEvent(ctx, &ContractLifecycleEvent{
		OrgID:       orgID,
		ContractID:  contractID,
		NewStatus:   string(ContractStatusActive),
		EventType:   EventType(EventContractUpdated),
		Description: ptrString(fmt.Sprintf("Approval request %s rejected by %s: %s", approvalID, userID, req.Reason)),
		PerformedBy: &userID,
		CreatedAt:   now,
	})

	appReq.Status = ApprovalStatusRejected
	appReq.DecisionBy = &userID
	appReq.DecisionComment = reason
	return appReq, nil
}

// CancelRequest cancels an open approval request.
func (e *ContractApprovalEngine) CancelRequest(
	ctx context.Context,
	orgID int64,
	contractID int64,
	approvalID string,
	req *CancelApprovalRequest,
	userID string,
) (*ContractApprovalRequest, error) {
	appReq, err := e.dl.GetApprovalRequestByID(ctx, orgID, approvalID)
	if err != nil {
		return nil, err
	}

	if appReq.Status != ApprovalStatusPending {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var reason *string
	if req != nil && req.Reason != "" {
		reason = &req.Reason
	}

	err = e.dl.UpdateApprovalDecision(ctx, orgID, approvalID, ApprovalStatusCancelled, userID, reason)
	if err != nil {
		return nil, err
	}

	if appReq.ApprovalType == ApprovalTypeAmendment && appReq.AmendmentID != nil {
		_ = e.dl.UpdateAmendmentStatus(ctx, orgID, *appReq.AmendmentID, AmendmentStatusCancelled)
	}

	appReq.Status = ApprovalStatusCancelled
	appReq.DecisionBy = &userID
	appReq.DecisionComment = reason
	return appReq, nil
}

