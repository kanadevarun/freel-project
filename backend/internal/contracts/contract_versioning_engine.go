package contracts

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/freel/backend/internal/svcerror"
)

// ContractVersioningEngine enforces deterministic version creation, immutability, and promotion.
type ContractVersioningEngine struct {
	dl DataLayer
}

func NewContractVersioningEngine(dl DataLayer) *ContractVersioningEngine {
	return &ContractVersioningEngine{dl: dl}
}

// CreateNewVersion creates an immutable draft version snapshot of a contract.
func (e *ContractVersioningEngine) CreateNewVersion(
	ctx context.Context,
	orgID int64,
	contractID int64,
	label string,
	changeSummary string,
	userID string,
) (*ContractVersion, error) {
	contract, err := e.dl.GetContractByID(ctx, orgID, contractID)
	if err != nil {
		return nil, err
	}

	latestNum, err := e.dl.GetLatestVersionNumber(ctx, orgID, contractID)
	if err != nil {
		return nil, err
	}

	nextVersionNum := latestNum + 1
	if label == "" {
		label = fmt.Sprintf("v%d.0", nextVersionNum)
	}

	// Capture contract snapshot
	snapshotBytes, err := json.Marshal(contract)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize contract snapshot: %w", err)
	}

	version := &ContractVersion{
		OrgID:            fmt.Sprintf("%d", orgID),
		ContractID:       contractID,
		VersionNumber:    nextVersionNum,
		VersionLabel:     label,
		Status:           VersionStatusDraft,
		EffectiveDate:    contract.EffectiveDate,
		ExpiryDate:       contract.ExpiryDate,
		ContractSnapshot: snapshotBytes,
		ChangeSummary:    &changeSummary,
		CreatedBy:        &userID,
	}

	id, err := e.dl.CreateContractVersion(ctx, version)
	if err != nil {
		return nil, err
	}
	version.ID = id
	return version, nil
}

// PromoteVersionToEffective transitions a version to EFFECTIVE and supersedes the existing effective version.
func (e *ContractVersioningEngine) PromoteVersionToEffective(
	ctx context.Context,
	orgID int64,
	contractID int64,
	versionID string,
	approvedBy string,
) (*ContractVersion, error) {
	targetVersion, err := e.dl.GetContractVersionByID(ctx, orgID, contractID, versionID)
	if err != nil {
		return nil, err
	}

	if targetVersion.Status == VersionStatusEffective {
		return targetVersion, nil
	}
	if targetVersion.Status == VersionStatusSuperseded {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	// 1. Supersede currently effective version if exists
	currentEffective, err := e.dl.GetCurrentEffectiveVersion(ctx, orgID, contractID)
	if err == nil && currentEffective != nil && currentEffective.ID != versionID {
		_ = e.dl.SupersedeVersion(ctx, orgID, currentEffective.ID)
	}

	// 2. Mark target version as EFFECTIVE
	err = e.dl.UpdateVersionStatus(ctx, orgID, versionID, VersionStatusEffective, &approvedBy)
	if err != nil {
		return nil, err
	}

	// 3. Update active contract parameters if snapshot has updated terms
	if len(targetVersion.ContractSnapshot) > 0 {
		var snapshotContract Contract
		if jsonErr := json.Unmarshal(targetVersion.ContractSnapshot, &snapshotContract); jsonErr == nil {
			// Apply non-destructive updates to base contract record
			contract, getErr := e.dl.GetContractByID(ctx, orgID, contractID)
			if getErr == nil {
				if snapshotContract.ContractValue != nil {
					contract.ContractValue = snapshotContract.ContractValue
				}
				if snapshotContract.EffectiveDate != nil {
					contract.EffectiveDate = snapshotContract.EffectiveDate
				}
				if snapshotContract.ExpiryDate != nil {
					contract.ExpiryDate = snapshotContract.ExpiryDate
				}
				_ = e.dl.UpdateContract(ctx, contract)
			}
		}
	}

	// 4. Log lifecycle audit event
	now := time.Now()
	_ = e.dl.LogLifecycleEvent(ctx, &ContractLifecycleEvent{
		OrgID:       orgID,
		ContractID:  contractID,
		NewStatus:   string(ContractStatusActive),
		EventType:   EventType(EventContractUpdated),
		Description: ptrString(fmt.Sprintf("Contract promoted to Version %s (%d) as EFFECTIVE by %s", targetVersion.VersionLabel, targetVersion.VersionNumber, approvedBy)),
		PerformedBy: &approvedBy,
		CreatedAt:   now,
	})

	targetVersion.Status = VersionStatusEffective
	return targetVersion, nil
}

// CompareVersions generates field-level differences between two versions.
func (e *ContractVersioningEngine) CompareVersions(
	ctx context.Context,
	orgID int64,
	contractID int64,
	baseVersionID string,
	targetVersionID string,
) (*ContractVersionComparison, error) {
	baseVer, err := e.dl.GetContractVersionByID(ctx, orgID, contractID, baseVersionID)
	if err != nil {
		return nil, err
	}
	targetVer, err := e.dl.GetContractVersionByID(ctx, orgID, contractID, targetVersionID)
	if err != nil {
		return nil, err
	}

	var baseContract, targetContract map[string]interface{}
	_ = json.Unmarshal(baseVer.ContractSnapshot, &baseContract)
	_ = json.Unmarshal(targetVer.ContractSnapshot, &targetContract)

	changes := make([]ContractFieldChangeDiff, 0)
	comparisonKeys := []string{
		"contract_name", "contract_reference", "contract_type", "transport_mode",
		"party_name", "currency", "contract_value", "effective_date", "expiry_date",
		"owner", "description", "status",
	}

	for _, key := range comparisonKeys {
		baseVal := formatInterfaceVal(baseContract[key])
		targetVal := formatInterfaceVal(targetContract[key])

		if baseVal != targetVal {
			changeType := "MODIFY"
			if baseVal == "" && targetVal != "" {
				changeType = "ADD"
			} else if baseVal != "" && targetVal == "" {
				changeType = "REMOVE"
			}

			changes = append(changes, ContractFieldChangeDiff{
				FieldName:     key,
				PreviousValue: &baseVal,
				ProposedValue: &targetVal,
				ChangeType:    changeType,
			})
		}
	}

	return &ContractVersionComparison{
		BaseVersionID:    baseVersionID,
		BaseVersionNum:   baseVer.VersionNumber,
		TargetVersionID:  targetVersionID,
		TargetVersionNum: targetVer.VersionNumber,
		Changes:          changes,
	}, nil
}

func formatInterfaceVal(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.2f", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case int:
		return fmt.Sprintf("%d", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func ptrString(s string) *string {
	return &s
}

