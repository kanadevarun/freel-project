package contracts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/freel/backend/internal/svcerror"
)

// ContractAmendmentEngine manages amendment creation, change calculations, review, and implementation.
type ContractAmendmentEngine struct {
	dl DataLayer
	ve *ContractVersioningEngine
}

func NewContractAmendmentEngine(dl DataLayer, ve *ContractVersioningEngine) *ContractAmendmentEngine {
	return &ContractAmendmentEngine{
		dl: dl,
		ve: ve,
	}
}

// CreateAmendment initializes a new structured amendment for a contract.
func (e *ContractAmendmentEngine) CreateAmendment(
	ctx context.Context,
	orgID int64,
	contractID int64,
	req *CreateContractAmendmentRequest,
	userID string,
) (*ContractAmendment, error) {
	contract, err := e.dl.GetContractByID(ctx, orgID, contractID)
	if err != nil {
		return nil, err
	}

	// 1. Get base version
	currentEffective, _ := e.dl.GetCurrentEffectiveVersion(ctx, orgID, contractID)
	var baseVersionID *string
	if currentEffective != nil {
		baseVersionID = &currentEffective.ID
	}

	// Generate deterministic amendment reference
	amendments, _ := e.dl.GetContractAmendments(ctx, orgID, contractID)
	amendmentRef := fmt.Sprintf("AMD-%s-%03d", contract.ContractReference, len(amendments)+1)

	amendment := &ContractAmendment{
		OrgID:                 fmt.Sprintf("%d", orgID),
		ContractID:            contractID,
		BaseVersionID:         baseVersionID,
		AmendmentReference:    amendmentRef,
		AmendmentType:         req.AmendmentType,
		Title:                 req.Title,
		Description:           req.Description,
		ChangeSummary:         req.ChangeSummary,
		Status:                AmendmentStatusDraft,
		ProposedEffectiveDate: req.ProposedEffectiveDate,
		CreatedBy:             &userID,
	}

	id, err := e.dl.CreateContractAmendment(ctx, amendment)
	if err != nil {
		return nil, err
	}
	amendment.ID = id

	// 2. Insert structured field changes
	if len(req.Changes) > 0 {
		changes := make([]ContractAmendmentChange, len(req.Changes))
		for i, c := range req.Changes {
			chgType := c.ChangeType
			if chgType == "" {
				chgType = "MODIFY"
			}
			changes[i] = ContractAmendmentChange{
				OrgID:         fmt.Sprintf("%d", orgID),
				AmendmentID:   id,
				FieldName:     c.FieldName,
				PreviousValue: c.PreviousValue,
				ProposedValue: c.ProposedValue,
				ChangeType:    chgType,
			}
		}
		_ = e.dl.CreateAmendmentChanges(ctx, orgID, id, changes)
		amendment.Changes = changes
	}

	return amendment, nil
}

// SubmitAmendment submits a draft amendment into review/approval.
func (e *ContractAmendmentEngine) SubmitAmendment(
	ctx context.Context,
	orgID int64,
	contractID int64,
	amendmentID string,
	req *SubmitContractAmendmentRequest,
	userID string,
) (*ContractAmendment, error) {
	amendment, err := e.dl.GetContractAmendmentByID(ctx, orgID, contractID, amendmentID)
	if err != nil {
		return nil, err
	}

	if amendment.Status != AmendmentStatusDraft {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	// Update status to SUBMITTED
	err = e.dl.UpdateAmendmentStatus(ctx, orgID, amendmentID, AmendmentStatusSubmitted)
	if err != nil {
		return nil, err
	}

	// Automatically create an approval request
	var assignedTo *string
	if req != nil && req.AssignedTo != nil {
		assignedTo = req.AssignedTo
	}

	approvalReq := &ContractApprovalRequest{
		OrgID:        fmt.Sprintf("%d", orgID),
		ContractID:   contractID,
		AmendmentID:  &amendmentID,
		ApprovalType: ApprovalTypeAmendment,
		Status:       ApprovalStatusPending,
		RequestedBy:  &userID,
		AssignedTo:   assignedTo,
	}
	_, _ = e.dl.CreateApprovalRequest(ctx, approvalReq)

	amendment.Status = AmendmentStatusSubmitted
	return amendment, nil
}

// ImplementApprovedAmendment converts an approved amendment into a new EFFECTIVE version of the contract.
func (e *ContractAmendmentEngine) ImplementApprovedAmendment(
	ctx context.Context,
	orgID int64,
	contractID int64,
	amendmentID string,
	userID string,
) (*ContractVersion, error) {
	amendment, err := e.dl.GetContractAmendmentByID(ctx, orgID, contractID, amendmentID)
	if err != nil {
		return nil, err
	}

	if amendment.Status != AmendmentStatusApproved {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	contract, err := e.dl.GetContractByID(ctx, orgID, contractID)
	if err != nil {
		return nil, err
	}

	// 1. Build updated snapshot applying amendment changes
	snapshotMap := make(map[string]interface{})
	if contractBytes, marshalErr := json.Marshal(contract); marshalErr == nil {
		_ = json.Unmarshal(contractBytes, &snapshotMap)
	}

	for _, change := range amendment.Changes {
		if change.ProposedValue != nil {
			snapshotMap[change.FieldName] = *change.ProposedValue
		}
	}
	snapshotMap["implemented_from_amendment"] = amendment.AmendmentReference

	updatedSnapshotBytes, _ := json.Marshal(snapshotMap)

	// 2. Create new version
	latestNum, err := e.dl.GetLatestVersionNumber(ctx, orgID, contractID)
	if err != nil {
		return nil, err
	}

	nextVersionNum := latestNum + 1
	versionLabel := fmt.Sprintf("v%d.0 (via %s)", nextVersionNum, amendment.AmendmentReference)
	changeSummary := fmt.Sprintf("Implemented Amendment %s: %s", amendment.AmendmentReference, amendment.Title)

	effectiveDate := contract.EffectiveDate
	if amendment.ProposedEffectiveDate != nil && *amendment.ProposedEffectiveDate != "" {
		effectiveDate = amendment.ProposedEffectiveDate
	}

	newVersion := &ContractVersion{
		OrgID:            fmt.Sprintf("%d", orgID),
		ContractID:       contractID,
		VersionNumber:    nextVersionNum,
		VersionLabel:     versionLabel,
		Status:           VersionStatusApproved,
		EffectiveDate:    effectiveDate,
		ExpiryDate:       contract.ExpiryDate,
		ContractSnapshot: updatedSnapshotBytes,
		ChangeSummary:    &changeSummary,
		CreatedBy:        &userID,
	}

	verID, err := e.dl.CreateContractVersion(ctx, newVersion)
	if err != nil {
		return nil, err
	}
	newVersion.ID = verID

	// 3. Promote version to EFFECTIVE (supersedes previous effective version)
	promotedVer, err := e.ve.PromoteVersionToEffective(ctx, orgID, contractID, verID, userID)
	if err != nil {
		return nil, err
	}

	// 4. Update amendment status to IMPLEMENTED
	_ = e.dl.UpdateAmendmentStatus(ctx, orgID, amendmentID, AmendmentStatusImplemented)

	return promotedVer, nil
}
