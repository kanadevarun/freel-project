package contracts

import (
	"context"
	"fmt"
	"time"
)

// TermsAndObligationsEngine manages evaluation of contract terms, obligation timelines, and compliance rules.
type TermsAndObligationsEngine interface {
	EvaluateContractObligations(ctx context.Context, orgID int64, contractID int64) error
	EvaluateComplianceRequirements(ctx context.Context, orgID int64, contractID int64) error
}

type termsEngine struct {
	dl DataLayer
}

// NewTermsEngine creates an instance of TermsAndObligationsEngine.
func NewTermsEngine(dl DataLayer) TermsAndObligationsEngine {
	return &termsEngine{dl: dl}
}

// EvaluateContractObligations checks due dates and progress thresholds to flag DUE_SOON or OVERDUE obligations without mutating commercial terms.
func (e *termsEngine) EvaluateContractObligations(ctx context.Context, orgID int64, contractID int64) error {
	obligations, err := e.dl.ListContractObligations(ctx, orgID, contractID)
	if err != nil {
		return fmt.Errorf("failed to list obligations: %w", err)
	}

	today := time.Now().UTC().Format("2006-01-02")
	soonThreshold := time.Now().UTC().AddDate(0, 0, 7).Format("2006-01-02")

	for _, ob := range obligations {
		if ob.Status == "COMPLETED" || ob.Status == "WAIVED" || ob.Status == "CANCELLED" {
			continue
		}

		if ob.DueDate != nil && *ob.DueDate != "" {
			if *ob.DueDate < today && ob.Status != "OVERDUE" && ob.Status != "BREACHED" {
				// Mark as OVERDUE
				newStatus := "OVERDUE"
				_ = e.dl.UpdateContractObligationStatus(ctx, orgID, ob.ID, newStatus)

				// Create compliance breach event if not already open
				event := &ContractComplianceEvent{
					OrgID:                orgID,
					ContractID:           contractID,
					ContractObligationID: &ob.ID,
					EventType:            "OBLIGATION_OVERDUE",
					Severity:             "WARNING",
					Status:               "OPEN",
					Title:                fmt.Sprintf("Contract Obligation Overdue: %s", ob.Title),
					Description:          stringPtr(fmt.Sprintf("Obligation '%s' (Ref: %s) passed its target deadline of %s without recorded completion.", ob.Title, ob.ObligationReference, *ob.DueDate)),
				}
				if ob.Priority == "CRITICAL" {
					event.Severity = "CRITICAL"
				}
				_ = e.dl.CreateContractComplianceEventIfNotExists(ctx, event)
			} else if *ob.DueDate >= today && *ob.DueDate <= soonThreshold && ob.Status == "ACTIVE" {
				// Mark as DUE_SOON
				_ = e.dl.UpdateContractObligationStatus(ctx, orgID, ob.ID, "DUE_SOON")

				event := &ContractComplianceEvent{
					OrgID:                orgID,
					ContractID:           contractID,
					ContractObligationID: &ob.ID,
					EventType:            "OBLIGATION_DUE_SOON",
					Severity:             "ATTENTION",
					Status:               "OPEN",
					Title:                fmt.Sprintf("Obligation Due Soon: %s", ob.Title),
					Description:          stringPtr(fmt.Sprintf("Obligation '%s' is due on %s (within 7 days).", ob.Title, *ob.DueDate)),
				}
				_ = e.dl.CreateContractComplianceEventIfNotExists(ctx, event)
			}
		}
	}

	return nil
}

// EvaluateComplianceRequirements checks expiry of certificates and mandatory compliance evidence.
func (e *termsEngine) EvaluateComplianceRequirements(ctx context.Context, orgID int64, contractID int64) error {
	reqs, err := e.dl.ListContractComplianceRequirements(ctx, orgID, contractID)
	if err != nil {
		return fmt.Errorf("failed to list compliance requirements: %w", err)
	}

	today := time.Now().UTC().Format("2006-01-02")
	soonThreshold := time.Now().UTC().AddDate(0, 0, 15).Format("2006-01-02")

	for _, req := range reqs {
		if req.Status == "WAIVED" {
			continue
		}

		if req.ValidUntil != nil && *req.ValidUntil != "" {
			if *req.ValidUntil < today && req.Status != "EXPIRED" {
				_ = e.dl.UpdateComplianceRequirementStatus(ctx, orgID, req.ID, "EXPIRED")

				event := &ContractComplianceEvent{
					OrgID:             orgID,
					ContractID:        contractID,
					RelatedEntityType: stringPtr("COMPLIANCE_REQUIREMENT"),
					RelatedEntityID:   &req.ID,
					EventType:         "COMPLIANCE_EXPIRED",
					Severity:          "CRITICAL",
					Status:            "OPEN",
					Title:             fmt.Sprintf("Mandatory Compliance Expired: %s", req.Title),
					Description:       stringPtr(fmt.Sprintf("Requirement '%s' (%s) expired on %s.", req.Title, req.RequirementType, *req.ValidUntil)),
				}
				_ = e.dl.CreateContractComplianceEventIfNotExists(ctx, event)
			} else if *req.ValidUntil >= today && *req.ValidUntil <= soonThreshold && req.Status == "COMPLIANT" {
				event := &ContractComplianceEvent{
					OrgID:             orgID,
					ContractID:        contractID,
					RelatedEntityType: stringPtr("COMPLIANCE_REQUIREMENT"),
					RelatedEntityID:   &req.ID,
					EventType:         "COMPLIANCE_EXPIRING_SOON",
					Severity:          "WARNING",
					Status:            "OPEN",
					Title:             fmt.Sprintf("Compliance Expiring Soon: %s", req.Title),
					Description:       stringPtr(fmt.Sprintf("Requirement '%s' validity expires on %s.", req.Title, *req.ValidUntil)),
				}
				_ = e.dl.CreateContractComplianceEventIfNotExists(ctx, event)
			}
		}
	}

	return nil
}

func stringPtr(s string) *string {
	return &s
}

