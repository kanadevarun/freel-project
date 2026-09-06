package contracts

// CanTransitionContractStatus returns true if a contract can transition from the current status to the new status.
func CanTransitionContractStatus(current, next ContractStatus) bool {
	// If the status is the same, no transition is needed, but we allow it as a no-op
	if current == next {
		return true
	}

	switch current {
	case ContractStatusDraft:
		return next == ContractStatusSubmitted || next == ContractStatusWithdrawn || next == ContractStatusArchived || next == ContractStatusApproved || next == ContractStatusActive
	case ContractStatusSubmitted:
		return next == ContractStatusApproved || next == ContractStatusWithdrawn || next == ContractStatusArchived || next == ContractStatusActive
	case ContractStatusApproved:
		return next == ContractStatusActive || next == ContractStatusTerminated || next == ContractStatusArchived
	case ContractStatusActive:
		return next == ContractStatusExpired || next == ContractStatusTerminated || next == ContractStatusArchived
	case ContractStatusExpired:
		// Expired contracts cannot become active again automatically. A new contract or renewal is needed.
		return next == ContractStatusArchived
	case ContractStatusWithdrawn, ContractStatusCancelled, ContractStatusTerminated:
		// Terminal states, can only be archived
		return next == ContractStatusArchived
	case ContractStatusArchived:
		// Cannot transition out of archived
		return false
	default:
		return false
	}
}

// IsContractEditable returns true if the contract is in a state where edits to its terms/parties are allowed.
func IsContractEditable(status ContractStatus) bool {
	switch status {
	case ContractStatusDraft, ContractStatusSubmitted:
		// Allow edits during the initial creation/review phases
		return true
	case ContractStatusApproved, ContractStatusActive, ContractStatusExpired, ContractStatusWithdrawn, ContractStatusCancelled, ContractStatusTerminated, ContractStatusArchived:
		// Once approved or active, contracts are locked. Edits should be done via amendments (Future task).
		// For now, we return false to prevent arbitrary updates to active contracts.
		return false
	default:
		return false
	}
}
