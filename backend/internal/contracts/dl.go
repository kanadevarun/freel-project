package contracts

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/freel/backend/internal/svcerror"
	"github.com/jmoiron/sqlx"
)

// DataLayer defines the contract repository methods
type DataLayer interface {
	CreateContract(ctx context.Context, contract *Contract) (int64, error)
	GetContractByID(ctx context.Context, orgID, contractID int64) (*Contract, error)
	ListContracts(ctx context.Context, orgID int64, req *ListContractsRequest) ([]*Contract, int, error)
	UpdateContract(ctx context.Context, contract *Contract) error
	UpdateContractStatus(ctx context.Context, orgID, contractID int64, status ContractStatus) error
	LogLifecycleEvent(ctx context.Context, event *ContractLifecycleEvent) error
	GetContractLifecycleEvents(ctx context.Context, orgID, contractID int64) ([]*ContractLifecycleEvent, error)
	GetContractOverview(ctx context.Context, orgID int64) (*ContractOverview, error)

	CreateContractLink(ctx context.Context, link *ContractLink) error
	GetContractLinkByID(ctx context.Context, orgID, linkID int64) (*ContractLink, error)
	DeleteContractLink(ctx context.Context, orgID, linkID int64) error
	CreateContractLinkHistory(ctx context.Context, history *ContractLinkHistory) error
	GetContractLinksHydrated(ctx context.Context, orgID, contractID int64) ([]ContractLinkedRecord, error)
	GetContractLinkHistory(ctx context.Context, orgID, contractID int64) ([]ContractLinkHistory, error)
	GetContractsForEntity(ctx context.Context, orgID int64, entityType LinkedEntityType, entityID int64) ([]Contract, error)
	
	// Task 20.3: Lifecycle Intelligence, Renewal & Risk
	CreateLifecycleIntelligenceEvent(ctx context.Context, e *ContractLifecycleIntelligenceEvent) error
	GetLifecycleIntelligenceEvents(ctx context.Context, orgID int64, contractID *int64, limit int) ([]ContractLifecycleIntelligenceEvent, error)
	CreateRiskEvent(ctx context.Context, r *ContractRiskEvent) error
	GetRiskEvents(ctx context.Context, orgID int64, contractID *int64, isResolved *bool) ([]ContractRiskEvent, error)
	ResolveRiskEvent(ctx context.Context, orgID, contractID, riskID int64, resolvedBy string, notes *string) error
	CreateOrUpdateRenewalTracking(ctx context.Context, r *ContractRenewalTracking) error
	GetRenewalTracking(ctx context.Context, orgID, contractID int64) (*ContractRenewalTracking, error)
	GetContractCommercialImpactSummary(ctx context.Context, orgID, contractID int64) (*ContractCommercialImpactSummary, error)
	GetAllContractsForOrg(ctx context.Context, orgID int64) ([]Contract, error)

	// Task 20.4: Contract Versioning, Amendments & Approval Workflow
	CreateContractVersion(ctx context.Context, v *ContractVersion) (string, error)
	GetContractVersionByID(ctx context.Context, orgID int64, contractID int64, versionID string) (*ContractVersion, error)
	GetContractVersions(ctx context.Context, orgID int64, contractID int64) ([]ContractVersion, error)
	GetCurrentEffectiveVersion(ctx context.Context, orgID int64, contractID int64) (*ContractVersion, error)
	GetLatestVersionNumber(ctx context.Context, orgID int64, contractID int64) (int, error)
	UpdateVersionStatus(ctx context.Context, orgID int64, versionID string, status VersionStatus, approvedBy *string) error
	SupersedeVersion(ctx context.Context, orgID int64, versionID string) error

	CreateContractAmendment(ctx context.Context, a *ContractAmendment) (string, error)
	GetContractAmendmentByID(ctx context.Context, orgID int64, contractID int64, amendmentID string) (*ContractAmendment, error)
	GetContractAmendments(ctx context.Context, orgID int64, contractID int64) ([]ContractAmendment, error)
	UpdateContractAmendment(ctx context.Context, a *ContractAmendment) error
	UpdateAmendmentStatus(ctx context.Context, orgID int64, amendmentID string, status AmendmentStatus) error
	CreateAmendmentChanges(ctx context.Context, orgID int64, amendmentID string, changes []ContractAmendmentChange) error
	GetAmendmentChanges(ctx context.Context, orgID int64, amendmentID string) ([]ContractAmendmentChange, error)

	CreateApprovalRequest(ctx context.Context, req *ContractApprovalRequest) (string, error)
	GetApprovalRequestByID(ctx context.Context, orgID int64, approvalID string) (*ContractApprovalRequest, error)
	GetContractApprovalRequests(ctx context.Context, orgID int64, contractID int64) ([]ContractApprovalRequest, error)
	UpdateApprovalDecision(ctx context.Context, orgID int64, approvalID string, status ApprovalStatus, decisionBy string, comment *string) error

	// Tasks 20.5 & 20.6: Agreement Documents, Terms, Obligations & Compliance
	ListContractAgreementDocuments(ctx context.Context, orgID int64, contractID int64) ([]ContractAgreementDocument, error)
	CreateContractAgreementDocument(ctx context.Context, doc *ContractAgreementDocument) (string, error)
	SupersedeContractAgreementDocument(ctx context.Context, orgID int64, docID string, newDocID string) error

	ListContractTerms(ctx context.Context, orgID int64, contractID int64) ([]ContractTerm, error)
	GetContractTermByID(ctx context.Context, orgID int64, termID int64) (*ContractTerm, error)
	CreateContractTerm(ctx context.Context, term *ContractTerm) (int64, error)
	UpdateContractTerm(ctx context.Context, term *ContractTerm) error
	DeleteContractTerm(ctx context.Context, orgID int64, termID int64) error

	ListContractObligations(ctx context.Context, orgID int64, contractID int64) ([]ContractObligation, error)
	GetContractObligationByID(ctx context.Context, orgID int64, obligationID int64) (*ContractObligation, error)
	CreateContractObligation(ctx context.Context, ob *ContractObligation) (int64, error)
	UpdateContractObligation(ctx context.Context, ob *ContractObligation) error
	UpdateContractObligationStatus(ctx context.Context, orgID int64, obligationID int64, status string) error
	FulfillContractObligation(ctx context.Context, orgID int64, obligationID int64, fulfilledBy int64, notes *string) error
	WaiveContractObligation(ctx context.Context, orgID int64, obligationID int64, waivedBy int64, notes *string) error

	ListContractComplianceEvents(ctx context.Context, orgID int64, contractID int64) ([]ContractComplianceEvent, error)
	CreateContractComplianceEventIfNotExists(ctx context.Context, ev *ContractComplianceEvent) error
	ResolveContractComplianceEvent(ctx context.Context, orgID int64, eventID int64, resolvedBy int64, notes string) error
	ListAllOpenComplianceEvents(ctx context.Context, orgID int64) ([]ContractComplianceEvent, error)

	ListContractComplianceRequirements(ctx context.Context, orgID int64, contractID int64) ([]ContractComplianceRequirement, error)
	GetComplianceRequirementByID(ctx context.Context, orgID int64, reqID int64) (*ContractComplianceRequirement, error)
	CreateContractComplianceRequirement(ctx context.Context, req *ContractComplianceRequirement) (int64, error)
	UpdateContractComplianceRequirement(ctx context.Context, req *ContractComplianceRequirement) error
	UpdateComplianceRequirementStatus(ctx context.Context, orgID int64, reqID int64, status string) error
	VerifyContractComplianceRequirement(ctx context.Context, orgID int64, reqID int64, status string, verifiedBy int64, docID *string, notes *string) error

	// Contract Document Import & AI-Assisted Contract Creation
	CheckContractDuplicates(ctx context.Context, orgID int64, ref string, partyName string) (*ContractDuplicateWarning, error)
	SaveExtractedContractData(ctx context.Context, orgID int64, docID string, data *ExtractedContractDraft, confidence int) error
	GetExtractedContractData(ctx context.Context, orgID int64, docID string) (*ExtractedContractDraft, error)
	GetCandidateParties(ctx context.Context, orgID int64, searchName string) ([]MatchedPartyCandidate, error)
	ConfirmContractImportTx(ctx context.Context, orgID int64, userID int64, req *ConfirmContractImportRequest) (*Contract, error)

	// Transaction support
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

type dataLayer struct {
	db *sqlx.DB
}

// NewDataLayer creates a new contracts data layer
func NewDataLayer(db *sqlx.DB) DataLayer {
	return &dataLayer{db: db}
}

func (d *dataLayer) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return d.db.BeginTxx(ctx, nil)
}

func (d *dataLayer) CreateContract(ctx context.Context, c *Contract) (int64, error) {
	query := `
		INSERT INTO contracts (
			org_id, contract_reference, contract_name, contract_type,
			party_id, party_name, transport_mode, status, currency,
			contract_value, effective_date, expiry_date, owner,
			description, notes, created_by, updated_by, source_document_id
		) VALUES (
			:org_id, :contract_reference, :contract_name, :contract_type,
			:party_id, :party_name, :transport_mode, :status, :currency,
			:contract_value, :effective_date, :expiry_date, :owner,
			:description, :notes, :created_by, :updated_by, :source_document_id
		)
	`
	res, err := d.db.NamedExecContext(ctx, query, c)
	if err != nil {
		return 0, fmt.Errorf("contracts.CreateContract: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (d *dataLayer) GetContractByID(ctx context.Context, orgID, contractID int64) (*Contract, error) {
	query := `
		SELECT 
			id, org_id, contract_reference, contract_name, contract_type,
			party_id, party_name, transport_mode, status, currency,
			contract_value, DATE_FORMAT(effective_date, '%Y-%m-%d') as effective_date, 
			DATE_FORMAT(expiry_date, '%Y-%m-%d') as expiry_date, owner, description, notes, 
			created_by, updated_by, created_at, updated_at, archived_at
		FROM contracts
		WHERE org_id = ? AND id = ? AND archived_at IS NULL
	`
	var c Contract
	if err := d.db.GetContext(ctx, &c, query, orgID, contractID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, fmt.Errorf("contracts.GetContractByID: %w", err)
	}
	return &c, nil
}

func (d *dataLayer) ListContracts(ctx context.Context, orgID int64, req *ListContractsRequest) ([]*Contract, int, error) {
	where := []string{"org_id = ?"}
	args := []interface{}{orgID}

	if req.Status != "" {
		where = append(where, "status = ?")
		args = append(args, req.Status)
	}
	if req.ContractType != "" {
		where = append(where, "contract_type = ?")
		args = append(args, req.ContractType)
	}
	if req.PartyID > 0 {
		where = append(where, "party_id = ?")
		args = append(args, req.PartyID)
	}
	if req.Search != "" {
		where = append(where, "(contract_name LIKE ? OR contract_reference LIKE ?)")
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	whereClause := strings.Join(where, " AND ")
	
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM contracts WHERE %s AND archived_at IS NULL", whereClause)
	var total int
	if err := d.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("contracts.ListContracts.Count: %w", err)
	}

	if total == 0 {
		return []*Contract{}, 0, nil
	}

	offset := (req.Page - 1) * req.Limit
	query := fmt.Sprintf(`
		SELECT 
			id, org_id, contract_reference, contract_name, contract_type,
			party_id, party_name, transport_mode, status, currency,
			contract_value, DATE_FORMAT(effective_date, '%%Y-%%m-%%d') as effective_date, 
			DATE_FORMAT(expiry_date, '%%Y-%%m-%%d') as expiry_date, owner, description, notes, 
			created_by, updated_by, created_at, updated_at, archived_at
		FROM contracts
		WHERE %s AND archived_at IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, req.Limit, offset)
	
	var list []*Contract
	if err := d.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, 0, fmt.Errorf("contracts.ListContracts.Select: %w", err)
	}

	return list, total, nil
}

func (d *dataLayer) UpdateContract(ctx context.Context, c *Contract) error {
	query := `
		UPDATE contracts SET
			contract_name = :contract_name,
			contract_type = :contract_type,
			party_id = :party_id,
			party_name = :party_name,
			transport_mode = :transport_mode,
			currency = :currency,
			contract_value = :contract_value,
			effective_date = :effective_date,
			expiry_date = :expiry_date,
			owner = :owner,
			description = :description,
			notes = :notes,
			updated_by = :updated_by,
			archived_at = :archived_at
		WHERE org_id = :org_id AND id = :id
	`
	_, err := d.db.NamedExecContext(ctx, query, c)
	if err != nil {
		return fmt.Errorf("contracts.UpdateContract: %w", err)
	}
	return nil
}

func (d *dataLayer) UpdateContractStatus(ctx context.Context, orgID, contractID int64, status ContractStatus) error {
	query := `UPDATE contracts SET status = ?, updated_at = NOW() WHERE org_id = ? AND id = ?`
	_, err := d.db.ExecContext(ctx, query, string(status), orgID, contractID)
	if err != nil {
		return fmt.Errorf("contracts.UpdateContractStatus: %w", err)
	}
	return nil
}

func (d *dataLayer) LogLifecycleEvent(ctx context.Context, event *ContractLifecycleEvent) error {
	query := `
		INSERT INTO contract_lifecycle_events (
			org_id, contract_id, previous_status, new_status, event_type, description, performed_by, metadata
		) VALUES (
			:org_id, :contract_id, :previous_status, :new_status, :event_type, :description, :performed_by, :metadata
		)
	`
	_, err := d.db.NamedExecContext(ctx, query, event)
	if err != nil {
		return fmt.Errorf("contracts.LogLifecycleEvent: %w", err)
	}
	return nil
}

func (d *dataLayer) GetContractLifecycleEvents(ctx context.Context, orgID, contractID int64) ([]*ContractLifecycleEvent, error) {
	query := `
		SELECT 
			id, org_id, contract_id, previous_status, new_status, event_type, 
			description, performed_by, metadata, created_at
		FROM contract_lifecycle_events
		WHERE org_id = ? AND contract_id = ?
		ORDER BY created_at ASC
	`
	var events []*ContractLifecycleEvent
	if err := d.db.SelectContext(ctx, &events, query, orgID, contractID); err != nil {
		return nil, fmt.Errorf("contracts.GetContractLifecycleEvents: %w", err)
	}
	if events == nil {
		events = []*ContractLifecycleEvent{}
	}
	return events, nil
}

func (d *dataLayer) GetContractOverview(ctx context.Context, orgID int64) (*ContractOverview, error) {
	query := `
		SELECT 
			COUNT(id) AS total_contracts,
			COALESCE(SUM(CASE WHEN status = 'ACTIVE' THEN 1 ELSE 0 END), 0) AS active_contracts,
			COALESCE(SUM(CASE WHEN status = 'EXPIRED' THEN 1 ELSE 0 END), 0) AS expired_contracts,
			COALESCE(SUM(CASE WHEN status = 'DRAFT' THEN 1 ELSE 0 END), 0) AS draft_contracts,
			COALESCE(SUM(CASE WHEN status = 'ACTIVE' AND expiry_date IS NOT NULL AND DATEDIFF(expiry_date, CURDATE()) <= 30 THEN 1 ELSE 0 END), 0) AS expiring_soon,
			COALESCE(SUM(CASE WHEN status = 'ACTIVE' THEN COALESCE(contract_value, 0) ELSE 0 END), 0) AS total_value
		FROM contracts
		WHERE org_id = ? AND archived_at IS NULL
	`
	var overview ContractOverview
	if err := d.db.GetContext(ctx, &overview, query, orgID); err != nil {
		return nil, fmt.Errorf("contracts.GetContractOverview: %w", err)
	}
	return &overview, nil
}

func (d *dataLayer) CreateContractLink(ctx context.Context, link *ContractLink) error {
	query := `
		INSERT INTO contract_links (
			org_id, contract_id, linked_entity_type, linked_entity_id, 
			link_type, is_primary, notes, created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := d.db.ExecContext(ctx, query,
		link.OrgID, link.ContractID, link.LinkedEntityType, link.LinkedEntityID,
		link.LinkType, link.IsPrimary, link.Notes, link.CreatedBy,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	link.ID = id
	return nil
}

func (d *dataLayer) GetContractLinkByID(ctx context.Context, orgID, linkID int64) (*ContractLink, error) {
	query := `SELECT * FROM contract_links WHERE org_id = ? AND id = ?`
	var link ContractLink
	err := d.db.GetContext(ctx, &link, query, orgID, linkID)
	if err != nil {
		if err == sql.ErrNoRows {
			e := svcerror.NewServiceError(svcerror.ErrResourceNotFound)
			e.Message = "Contract link not found"
			return nil, e
		}
		return nil, err
	}
	return &link, nil
}

func (d *dataLayer) DeleteContractLink(ctx context.Context, orgID, linkID int64) error {
	query := `DELETE FROM contract_links WHERE org_id = ? AND id = ?`
	_, err := d.db.ExecContext(ctx, query, orgID, linkID)
	return err
}

func (d *dataLayer) CreateContractLinkHistory(ctx context.Context, history *ContractLinkHistory) error {
	query := `
		INSERT INTO contract_link_history (
			org_id, contract_id, contract_link_id, linked_entity_type, 
			linked_entity_id, link_type, action, previous_metadata, performed_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := d.db.ExecContext(ctx, query,
		history.OrgID, history.ContractID, history.ContractLinkID, history.LinkedEntityType,
		history.LinkedEntityID, history.LinkType, history.Action, history.PreviousMetadata, history.PerformedBy,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	history.ID = id
	return nil
}

func (d *dataLayer) GetContractLinksHydrated(ctx context.Context, orgID, contractID int64) ([]ContractLinkedRecord, error) {
	query := `SELECT * FROM contract_links WHERE org_id = ? AND contract_id = ? ORDER BY created_at DESC`
	var links []ContractLink
	err := d.db.SelectContext(ctx, &links, query, orgID, contractID)
	if err != nil {
		return nil, err
	}
	
	var records []ContractLinkedRecord
	for _, link := range links {
		rec := ContractLinkedRecord{
			ContractLink: link,
			ReferenceName: fmt.Sprintf("%s #%d", link.LinkedEntityType, link.LinkedEntityID),
			EntityStatus: "ACTIVE",
		}

		switch link.LinkedEntityType {
		case EntityTypeQuotation:
			var ref, status string
			err := d.db.QueryRowContext(ctx, "SELECT quote_number, status FROM quotations WHERE org_id = ? AND id = ?", orgID, link.LinkedEntityID).Scan(&ref, &status)
			if err == nil {
				rec.ReferenceName = ref
				rec.EntityStatus = status
			}
		case EntityTypeSpotRateRequest:
			var ref, status string
			err := d.db.QueryRowContext(ctx, "SELECT request_reference, status FROM spot_rate_requests WHERE org_id = ? AND id = ?", orgID, link.LinkedEntityID).Scan(&ref, &status)
			if err == nil {
				rec.ReferenceName = ref
				rec.EntityStatus = status
			}
		case EntityTypeRate:
			var ref, status string
			err := d.db.QueryRowContext(ctx, "SELECT rate_reference, status FROM managed_rates WHERE org_id = ? AND id = ?", orgID, link.LinkedEntityID).Scan(&ref, &status)
			if err == nil {
				rec.ReferenceName = ref
				rec.EntityStatus = status
			}
		}

		records = append(records, rec)
	}
	return records, nil
}

func (d *dataLayer) GetContractLinkHistory(ctx context.Context, orgID, contractID int64) ([]ContractLinkHistory, error) {
	query := `SELECT * FROM contract_link_history WHERE org_id = ? AND contract_id = ? ORDER BY created_at DESC`
	var history []ContractLinkHistory
	err := d.db.SelectContext(ctx, &history, query, orgID, contractID)
	return history, err
}

func (d *dataLayer) GetContractsForEntity(ctx context.Context, orgID int64, entityType LinkedEntityType, entityID int64) ([]Contract, error) {
	query := `
		SELECT c.* FROM contracts c
		JOIN contract_links cl ON c.id = cl.contract_id
		WHERE cl.org_id = ? AND cl.linked_entity_type = ? AND cl.linked_entity_id = ?
	`
	var contracts []Contract
	err := d.db.SelectContext(ctx, &contracts, query, orgID, entityType, entityID)
	return contracts, err
}

func (d *dataLayer) GetAllContractsForOrg(ctx context.Context, orgID int64) ([]Contract, error) {
	query := `SELECT * FROM contracts WHERE org_id = ? ORDER BY created_at DESC`
	var list []Contract
	err := d.db.SelectContext(ctx, &list, query, orgID)
	return list, err
}

func (d *dataLayer) CreateLifecycleIntelligenceEvent(ctx context.Context, e *ContractLifecycleIntelligenceEvent) error {
	query := `
		INSERT INTO contract_lifecycle_intelligence_events (
			org_id, contract_id, event_type, previous_state,
			new_state, severity, description, metadata
		) VALUES (
			:org_id, :contract_id, :event_type, :previous_state,
			:new_state, :severity, :description, :metadata
		)
	`
	_, err := d.db.NamedExecContext(ctx, query, e)
	return err
}

func (d *dataLayer) GetLifecycleIntelligenceEvents(ctx context.Context, orgID int64, contractID *int64, limit int) ([]ContractLifecycleIntelligenceEvent, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var events []ContractLifecycleIntelligenceEvent
	if contractID != nil && *contractID > 0 {
		query := `SELECT * FROM contract_lifecycle_intelligence_events WHERE org_id = ? AND contract_id = ? ORDER BY created_at DESC LIMIT ?`
		err := d.db.SelectContext(ctx, &events, query, orgID, *contractID, limit)
		return events, err
	}
	query := `SELECT * FROM contract_lifecycle_intelligence_events WHERE org_id = ? ORDER BY created_at DESC LIMIT ?`
	err := d.db.SelectContext(ctx, &events, query, orgID, limit)
	return events, err
}

func (d *dataLayer) CreateRiskEvent(ctx context.Context, r *ContractRiskEvent) error {
	// Prevent duplicate unresolved risk events for the same contract & risk_type
	var existingID int64
	checkQuery := `SELECT id FROM contract_risk_events WHERE org_id = ? AND contract_id = ? AND risk_type = ? AND is_resolved = FALSE LIMIT 1`
	err := d.db.QueryRowContext(ctx, checkQuery, r.OrgID, r.ContractID, r.RiskType).Scan(&existingID)
	if err == nil && existingID > 0 {
		// Already active, update severity & description
		updateQuery := `UPDATE contract_risk_events SET severity = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND org_id = ?`
		_, updateErr := d.db.ExecContext(ctx, updateQuery, r.Severity, r.Description, existingID, r.OrgID)
		return updateErr
	}

	insertQuery := `
		INSERT INTO contract_risk_events (
			org_id, contract_id, risk_type, severity, description,
			is_resolved, metadata
		) VALUES (
			:org_id, :contract_id, :risk_type, :severity, :description,
			:is_resolved, :metadata
		)
	`
	_, err = d.db.NamedExecContext(ctx, insertQuery, r)
	return err
}

func (d *dataLayer) GetRiskEvents(ctx context.Context, orgID int64, contractID *int64, isResolved *bool) ([]ContractRiskEvent, error) {
	var risks []ContractRiskEvent
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "org_id = ?")
	args = append(args, orgID)

	if contractID != nil && *contractID > 0 {
		conditions = append(conditions, "contract_id = ?")
		args = append(args, *contractID)
	}

	if isResolved != nil {
		conditions = append(conditions, "is_resolved = ?")
		args = append(args, *isResolved)
	}

	query := fmt.Sprintf("SELECT * FROM contract_risk_events WHERE %s ORDER BY is_resolved ASC, CASE severity WHEN 'CRITICAL' THEN 1 WHEN 'WARNING' THEN 2 WHEN 'ATTENTION' THEN 3 ELSE 4 END, created_at DESC", strings.Join(conditions, " AND "))
	err := d.db.SelectContext(ctx, &risks, query, args...)
	return risks, err
}

func (d *dataLayer) ResolveRiskEvent(ctx context.Context, orgID, contractID, riskID int64, resolvedBy string, notes *string) error {
	query := `
		UPDATE contract_risk_events
		SET is_resolved = TRUE, resolved_by = ?, resolved_at = CURRENT_TIMESTAMP, resolution_notes = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND org_id = ? AND contract_id = ?
	`
	res, err := d.db.ExecContext(ctx, query, resolvedBy, notes, riskID, orgID, contractID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}
	return nil
}

func (d *dataLayer) CreateOrUpdateRenewalTracking(ctx context.Context, r *ContractRenewalTracking) error {
	query := `
		INSERT INTO contract_renewal_tracking (
			org_id, contract_id, renewal_status, renewal_start_date,
			target_completion_date, successor_contract_id, owner, notes, created_by
		) VALUES (
			:org_id, :contract_id, :renewal_status, :renewal_start_date,
			:target_completion_date, :successor_contract_id, :owner, :notes, :created_by
		)
		ON DUPLICATE KEY UPDATE
			renewal_status = VALUES(renewal_status),
			renewal_start_date = COALESCE(VALUES(renewal_start_date), renewal_start_date),
			target_completion_date = VALUES(target_completion_date),
			successor_contract_id = VALUES(successor_contract_id),
			owner = VALUES(owner),
			notes = VALUES(notes),
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := d.db.NamedExecContext(ctx, query, r)
	return err
}

func (d *dataLayer) GetRenewalTracking(ctx context.Context, orgID, contractID int64) (*ContractRenewalTracking, error) {
	var r ContractRenewalTracking
	query := `SELECT * FROM contract_renewal_tracking WHERE org_id = ? AND contract_id = ? LIMIT 1`
	err := d.db.GetContext(ctx, &r, query, orgID, contractID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// If successor is linked, hydrate successor info
	if r.SuccessorContractID != nil && *r.SuccessorContractID > 0 {
		var sName, sRef string
		_ = d.db.QueryRowContext(ctx, "SELECT contract_name, contract_reference FROM contracts WHERE org_id = ? AND id = ?", orgID, *r.SuccessorContractID).Scan(&sName, &sRef)
		r.SuccessorName = &sName
		r.SuccessorReference = &sRef
	}

	return &r, nil
}

func (d *dataLayer) GetContractCommercialImpactSummary(ctx context.Context, orgID, contractID int64) (*ContractCommercialImpactSummary, error) {
	summary := &ContractCommercialImpactSummary{ContractID: contractID}

	// 1. Linked Rates Count & Active Rates Count
	_ = d.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(cl.id),
			COALESCE(SUM(CASE WHEN r.status = 'ACTIVE' THEN 1 ELSE 0 END), 0)
		FROM contract_links cl
		LEFT JOIN managed_rates r ON cl.linked_entity_id = r.id AND cl.org_id = r.org_id
		WHERE cl.org_id = ? AND cl.contract_id = ? AND cl.linked_entity_type = 'RATE'
	`, orgID, contractID).Scan(&summary.LinkedRatesCount, &summary.ActiveRatesCount)

	// 2. Linked Quotations Count, Draft Count, Accepted Count
	_ = d.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(cl.id),
			COALESCE(SUM(CASE WHEN q.status = 'DRAFT' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN q.status = 'ACCEPTED' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(q.total_amount), 0)
		FROM contract_links cl
		LEFT JOIN quotations q ON cl.linked_entity_id = q.id AND cl.org_id = q.org_id
		WHERE cl.org_id = ? AND cl.contract_id = ? AND cl.linked_entity_type = 'QUOTATION'
	`, orgID, contractID).Scan(&summary.LinkedQuotationsCount, &summary.DraftQuotationsCount, &summary.AcceptedQuotationsCount, &summary.TotalCommercialExposure)

	// 3. Spot Requests & Responses
	_ = d.db.QueryRowContext(ctx, `
		SELECT COUNT(cl.id)
		FROM contract_links cl
		WHERE cl.org_id = ? AND cl.contract_id = ? AND cl.linked_entity_type = 'SPOT_RATE_REQUEST'
	`, orgID, contractID).Scan(&summary.SpotRequestsCount)

	_ = d.db.QueryRowContext(ctx, `
		SELECT COUNT(cl.id)
		FROM contract_links cl
		WHERE cl.org_id = ? AND cl.contract_id = ? AND cl.linked_entity_type = 'SPOT_RATE_RESPONSE'
	`, orgID, contractID).Scan(&summary.SpotResponsesCount)

	// 4. Affected Parties
	_ = d.db.QueryRowContext(ctx, `
		SELECT COUNT(cl.id)
		FROM contract_links cl
		WHERE cl.org_id = ? AND cl.contract_id = ? AND cl.linked_entity_type IN ('CUSTOMER', 'CARRIER', 'VENDOR')
	`, orgID, contractID).Scan(&summary.AffectedPartiesCount)

	return summary, nil
}

// ── Task 20.4 DataLayer Implementation ──────────────────────────────────────

func (d *dataLayer) CreateContractVersion(ctx context.Context, v *ContractVersion) (string, error) {
	query := `
		INSERT INTO contract_versions (
			org_id, contract_id, version_number, version_label, status,
			effective_date, expiry_date, contract_snapshot, change_summary,
			created_by, approved_by, approved_at, superseded_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?
		)
	`
	res, err := d.db.ExecContext(ctx, query,
		v.OrgID, v.ContractID, v.VersionNumber, v.VersionLabel, v.Status,
		v.EffectiveDate, v.ExpiryDate, v.ContractSnapshot, v.ChangeSummary,
		v.CreatedBy, v.ApprovedBy, v.ApprovedAt, v.SupersededAt,
	)
	if err != nil {
		return "", fmt.Errorf("contracts.CreateContractVersion: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return "", err
	}
	v.ID = fmt.Sprintf("%d", id)
	return v.ID, nil
}

func (d *dataLayer) GetContractVersionByID(ctx context.Context, orgID int64, contractID int64, versionID string) (*ContractVersion, error) {
	query := `
		SELECT 
			id, org_id, contract_id, version_number, version_label, status,
			DATE_FORMAT(effective_date, '%Y-%m-%d') as effective_date, 
			DATE_FORMAT(expiry_date, '%Y-%m-%d') as expiry_date, 
			contract_snapshot, change_summary,
			created_by, approved_by, approved_at, superseded_at, created_at, updated_at
		FROM contract_versions
		WHERE org_id = ? AND contract_id = ? AND id = ?
	`
	var v ContractVersion
	if err := d.db.GetContext(ctx, &v, query, orgID, contractID, versionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}
	return &v, nil
}

func (d *dataLayer) GetContractVersions(ctx context.Context, orgID int64, contractID int64) ([]ContractVersion, error) {
	query := `
		SELECT 
			id, org_id, contract_id, version_number, version_label, status,
			DATE_FORMAT(effective_date, '%Y-%m-%d') as effective_date, 
			DATE_FORMAT(expiry_date, '%Y-%m-%d') as expiry_date, 
			contract_snapshot, change_summary,
			created_by, approved_by, approved_at, superseded_at, created_at, updated_at
		FROM contract_versions
		WHERE org_id = ? AND contract_id = ?
		ORDER BY version_number DESC
	`
	var versions []ContractVersion
	if err := d.db.SelectContext(ctx, &versions, query, orgID, contractID); err != nil {
		return nil, err
	}
	return versions, nil
}

func (d *dataLayer) GetCurrentEffectiveVersion(ctx context.Context, orgID int64, contractID int64) (*ContractVersion, error) {
	query := `
		SELECT 
			id, org_id, contract_id, version_number, version_label, status,
			DATE_FORMAT(effective_date, '%Y-%m-%d') as effective_date, 
			DATE_FORMAT(expiry_date, '%Y-%m-%d') as expiry_date, 
			contract_snapshot, change_summary,
			created_by, approved_by, approved_at, superseded_at, created_at, updated_at
		FROM contract_versions
		WHERE org_id = ? AND contract_id = ? AND status = 'EFFECTIVE'
		ORDER BY version_number DESC
		LIMIT 1
	`
	var v ContractVersion
	if err := d.db.GetContext(ctx, &v, query, orgID, contractID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (d *dataLayer) GetLatestVersionNumber(ctx context.Context, orgID int64, contractID int64) (int, error) {
	query := `
		SELECT COALESCE(MAX(version_number), 0)
		FROM contract_versions
		WHERE org_id = ? AND contract_id = ?
	`
	var maxVer int
	if err := d.db.QueryRowContext(ctx, query, orgID, contractID).Scan(&maxVer); err != nil {
		return 0, err
	}
	return maxVer, nil
}

func (d *dataLayer) UpdateVersionStatus(ctx context.Context, orgID int64, versionID string, status VersionStatus, approvedBy *string) error {
	query := `
		UPDATE contract_versions
		SET status = ?, approved_by = COALESCE(?, approved_by), approved_at = CASE WHEN ? = 'APPROVED' OR ? = 'EFFECTIVE' THEN CURRENT_TIMESTAMP ELSE approved_at END, updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ?
	`
	_, err := d.db.ExecContext(ctx, query, status, approvedBy, status, status, orgID, versionID)
	return err
}

func (d *dataLayer) SupersedeVersion(ctx context.Context, orgID int64, versionID string) error {
	query := `
		UPDATE contract_versions
		SET status = 'SUPERSEDED', superseded_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ? AND status = 'EFFECTIVE'
	`
	_, err := d.db.ExecContext(ctx, query, orgID, versionID)
	return err
}

func (d *dataLayer) CreateContractAmendment(ctx context.Context, a *ContractAmendment) (string, error) {
	query := `
		INSERT INTO contract_amendments (
			org_id, contract_id, base_version_id, amendment_reference, amendment_type,
			title, description, change_summary, status, proposed_effective_date,
			created_by, submitted_at, approved_at, rejected_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?
		)
	`
	res, err := d.db.ExecContext(ctx, query,
		a.OrgID, a.ContractID, a.BaseVersionID, a.AmendmentReference, a.AmendmentType,
		a.Title, a.Description, a.ChangeSummary, a.Status, a.ProposedEffectiveDate,
		a.CreatedBy, a.SubmittedAt, a.ApprovedAt, a.RejectedAt,
	)
	if err != nil {
		return "", fmt.Errorf("contracts.CreateContractAmendment: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return "", err
	}
	a.ID = fmt.Sprintf("%d", id)
	return a.ID, nil
}

func (d *dataLayer) GetContractAmendmentByID(ctx context.Context, orgID int64, contractID int64, amendmentID string) (*ContractAmendment, error) {
	query := `
		SELECT 
			id, org_id, contract_id, base_version_id, amendment_reference, amendment_type,
			title, description, change_summary, status, 
			DATE_FORMAT(proposed_effective_date, '%Y-%m-%d') as proposed_effective_date,
			created_by, submitted_at, approved_at, rejected_at, created_at, updated_at
		FROM contract_amendments
		WHERE org_id = ? AND contract_id = ? AND id = ?
	`
	var a ContractAmendment
	if err := d.db.GetContext(ctx, &a, query, orgID, contractID, amendmentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	changes, err := d.GetAmendmentChanges(ctx, orgID, amendmentID)
	if err == nil {
		a.Changes = changes
	}

	return &a, nil
}

func (d *dataLayer) GetContractAmendments(ctx context.Context, orgID int64, contractID int64) ([]ContractAmendment, error) {
	query := `
		SELECT 
			id, org_id, contract_id, base_version_id, amendment_reference, amendment_type,
			title, description, change_summary, status, 
			DATE_FORMAT(proposed_effective_date, '%Y-%m-%d') as proposed_effective_date,
			created_by, submitted_at, approved_at, rejected_at, created_at, updated_at
		FROM contract_amendments
		WHERE org_id = ? AND contract_id = ?
		ORDER BY created_at DESC
	`
	var amendments []ContractAmendment
	if err := d.db.SelectContext(ctx, &amendments, query, orgID, contractID); err != nil {
		return nil, err
	}

	for i := range amendments {
		changes, err := d.GetAmendmentChanges(ctx, orgID, amendments[i].ID)
		if err == nil {
			amendments[i].Changes = changes
		}
	}

	return amendments, nil
}

func (d *dataLayer) UpdateContractAmendment(ctx context.Context, a *ContractAmendment) error {
	query := `
		UPDATE contract_amendments
		SET title = ?, description = ?, change_summary = ?, proposed_effective_date = ?,
		    status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ? AND status IN ('DRAFT', 'SUBMITTED')
	`
	_, err := d.db.ExecContext(ctx, query, a.Title, a.Description, a.ChangeSummary, a.ProposedEffectiveDate, a.Status, a.OrgID, a.ID)
	return err
}

func (d *dataLayer) UpdateAmendmentStatus(ctx context.Context, orgID int64, amendmentID string, status AmendmentStatus) error {
	query := `
		UPDATE contract_amendments
		SET status = ?, 
		    submitted_at = CASE WHEN ? = 'SUBMITTED' THEN CURRENT_TIMESTAMP ELSE submitted_at END,
		    approved_at = CASE WHEN ? = 'APPROVED' THEN CURRENT_TIMESTAMP ELSE approved_at END,
		    rejected_at = CASE WHEN ? = 'REJECTED' THEN CURRENT_TIMESTAMP ELSE rejected_at END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ?
	`
	_, err := d.db.ExecContext(ctx, query, status, status, status, status, orgID, amendmentID)
	return err
}

func (d *dataLayer) CreateAmendmentChanges(ctx context.Context, orgID int64, amendmentID string, changes []ContractAmendmentChange) error {
	_, _ = d.db.ExecContext(ctx, `DELETE FROM contract_amendment_changes WHERE org_id = ? AND amendment_id = ?`, orgID, amendmentID)

	for _, c := range changes {
		query := `
			INSERT INTO contract_amendment_changes (
				org_id, amendment_id, field_name, previous_value, proposed_value, change_type
			) VALUES (
				?, ?, ?, ?, ?, ?
			)
		`
		_, err := d.db.ExecContext(ctx, query, orgID, amendmentID, c.FieldName, c.PreviousValue, c.ProposedValue, c.ChangeType)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *dataLayer) GetAmendmentChanges(ctx context.Context, orgID int64, amendmentID string) ([]ContractAmendmentChange, error) {
	query := `
		SELECT id, org_id, amendment_id, field_name, previous_value, proposed_value, change_type, created_at
		FROM contract_amendment_changes
		WHERE org_id = ? AND amendment_id = ?
		ORDER BY created_at ASC
	`
	var changes []ContractAmendmentChange
	if err := d.db.SelectContext(ctx, &changes, query, orgID, amendmentID); err != nil {
		return nil, err
	}
	return changes, nil
}

func (d *dataLayer) CreateApprovalRequest(ctx context.Context, req *ContractApprovalRequest) (string, error) {
	query := `
		INSERT INTO contract_approval_requests (
			org_id, contract_id, version_id, amendment_id, approval_type,
			status, requested_by, assigned_to
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?
		)
	`
	res, err := d.db.ExecContext(ctx, query,
		req.OrgID, req.ContractID, req.VersionID, req.AmendmentID, req.ApprovalType,
		req.Status, req.RequestedBy, req.AssignedTo,
	)
	if err != nil {
		return "", fmt.Errorf("contracts.CreateApprovalRequest: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return "", err
	}
	req.ID = fmt.Sprintf("%d", id)
	return req.ID, nil
}

func (d *dataLayer) GetApprovalRequestByID(ctx context.Context, orgID int64, approvalID string) (*ContractApprovalRequest, error) {
	query := `
		SELECT 
			id, org_id, contract_id, version_id, amendment_id, approval_type,
			status, requested_by, assigned_to, decision_by, decision_comment,
			requested_at, decided_at, created_at, updated_at
		FROM contract_approval_requests
		WHERE org_id = ? AND id = ?
	`
	var r ContractApprovalRequest
	if err := d.db.GetContext(ctx, &r, query, orgID, approvalID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}
	return &r, nil
}

func (d *dataLayer) GetContractApprovalRequests(ctx context.Context, orgID int64, contractID int64) ([]ContractApprovalRequest, error) {
	query := `
		SELECT 
			id, org_id, contract_id, version_id, amendment_id, approval_type,
			status, requested_by, assigned_to, decision_by, decision_comment,
			requested_at, decided_at, created_at, updated_at
		FROM contract_approval_requests
		WHERE org_id = ? AND contract_id = ?
		ORDER BY requested_at DESC
	`
	var requests []ContractApprovalRequest
	if err := d.db.SelectContext(ctx, &requests, query, orgID, contractID); err != nil {
		return nil, err
	}
	return requests, nil
}

func (d *dataLayer) UpdateApprovalDecision(ctx context.Context, orgID int64, approvalID string, status ApprovalStatus, decisionBy string, comment *string) error {
	query := `
		UPDATE contract_approval_requests
		SET status = ?, decision_by = ?, decision_comment = ?, decided_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ? AND status = 'PENDING'
	`
	_, err := d.db.ExecContext(ctx, query, status, decisionBy, comment, orgID, approvalID)
	return err
}

// ═══════════════════════════════════════════════════════════════════════════
// Tasks 20.5 & 20.6: Agreement Documents, Terms, Obligations & Compliance
// ═══════════════════════════════════════════════════════════════════════════

func (d *dataLayer) ListContractAgreementDocuments(ctx context.Context, orgID int64, contractID int64) ([]ContractAgreementDocument, error) {
	query := `
		SELECT 
			id, org_id, contract_id, contract_version_id, document_type, document_name,
			file_name, s3_key, file_type, file_size_bytes, status, document_status,
			is_current, supersedes_document_id, 
			DATE_FORMAT(effective_date, '%Y-%m-%d') as effective_date,
			DATE_FORMAT(expiry_date, '%Y-%m-%d') as expiry_date,
			description, created_by, created_at, updated_at
		FROM contract_documents
		WHERE org_id = ? AND contract_id = ?
		ORDER BY created_at DESC
	`
	var docs []ContractAgreementDocument
	if err := d.db.SelectContext(ctx, &docs, query, orgID, contractID); err != nil {
		return nil, err
	}
	return docs, nil
}

func (d *dataLayer) CreateContractAgreementDocument(ctx context.Context, doc *ContractAgreementDocument) (string, error) {
	query := `
		INSERT INTO contract_documents (
			id, org_id, contract_id, contract_version_id, document_type, document_name,
			file_name, s3_key, file_type, file_size_bytes, status, document_status,
			is_current, supersedes_document_id, effective_date, expiry_date, description, created_by
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?
		)
	`
	_, err := d.db.ExecContext(ctx, query,
		doc.ID, doc.OrgID, doc.ContractID, doc.ContractVersionID, doc.DocumentType, doc.DocumentName,
		doc.FileName, doc.S3Key, doc.FileType, doc.FileSizeBytes, doc.Status, doc.DocumentStatus,
		doc.IsCurrent, doc.SupersedesDocumentID, doc.EffectiveDate, doc.ExpiryDate, doc.Description, doc.CreatedBy,
	)
	if err != nil {
		return "", fmt.Errorf("contracts.CreateContractAgreementDocument: %w", err)
	}
	return doc.ID, nil
}

func (d *dataLayer) SupersedeContractAgreementDocument(ctx context.Context, orgID int64, docID string, newDocID string) error {
	query := `
		UPDATE contract_documents
		SET is_current = FALSE, document_status = 'SUPERSEDED', updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ?
	`
	_, err := d.db.ExecContext(ctx, query, orgID, docID)
	return err
}

func (d *dataLayer) ListContractTerms(ctx context.Context, orgID int64, contractID int64) ([]ContractTerm, error) {
	query := `
		SELECT 
			id, org_id, contract_id, contract_version_id, term_category, term_key,
			term_title, term_value, value_type, currency,
			DATE_FORMAT(effective_date, '%Y-%m-%d') as effective_date,
			DATE_FORMAT(expiry_date, '%Y-%m-%d') as expiry_date,
			display_order, is_critical, created_by, updated_by, created_at, updated_at
		FROM contract_terms
		WHERE org_id = ? AND contract_id = ?
		ORDER BY display_order ASC, created_at ASC
	`
	var terms []ContractTerm
	if err := d.db.SelectContext(ctx, &terms, query, orgID, contractID); err != nil {
		return nil, err
	}
	return terms, nil
}

func (d *dataLayer) GetContractTermByID(ctx context.Context, orgID int64, termID int64) (*ContractTerm, error) {
	query := `
		SELECT 
			id, org_id, contract_id, contract_version_id, term_category, term_key,
			term_title, term_value, value_type, currency,
			DATE_FORMAT(effective_date, '%Y-%m-%d') as effective_date,
			DATE_FORMAT(expiry_date, '%Y-%m-%d') as expiry_date,
			display_order, is_critical, created_by, updated_by, created_at, updated_at
		FROM contract_terms
		WHERE org_id = ? AND id = ?
	`
	var t ContractTerm
	if err := d.db.GetContext(ctx, &t, query, orgID, termID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}
	return &t, nil
}

func (d *dataLayer) CreateContractTerm(ctx context.Context, term *ContractTerm) (int64, error) {
	query := `
		INSERT INTO contract_terms (
			org_id, contract_id, contract_version_id, term_category, term_key,
			term_title, term_value, value_type, currency, effective_date, expiry_date,
			display_order, is_critical, created_by
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?
		)
	`
	res, err := d.db.ExecContext(ctx, query,
		term.OrgID, term.ContractID, term.ContractVersionID, term.TermCategory, term.TermKey,
		term.TermTitle, term.TermValue, term.ValueType, term.Currency, term.EffectiveDate, term.ExpiryDate,
		term.DisplayOrder, term.IsCritical, term.CreatedBy,
	)
	if err != nil {
		return 0, fmt.Errorf("contracts.CreateContractTerm: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	term.ID = id
	return id, nil
}

func (d *dataLayer) UpdateContractTerm(ctx context.Context, term *ContractTerm) error {
	query := `
		UPDATE contract_terms
		SET term_category = ?, term_title = ?, term_value = ?, value_type = ?,
		    currency = ?, effective_date = ?, expiry_date = ?, display_order = ?,
		    is_critical = ?, updated_by = ?, updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ?
	`
	_, err := d.db.ExecContext(ctx, query,
		term.TermCategory, term.TermTitle, term.TermValue, term.ValueType,
		term.Currency, term.EffectiveDate, term.ExpiryDate, term.DisplayOrder,
		term.IsCritical, term.UpdatedBy, term.OrgID, term.ID,
	)
	return err
}

func (d *dataLayer) DeleteContractTerm(ctx context.Context, orgID int64, termID int64) error {
	query := `DELETE FROM contract_terms WHERE org_id = ? AND id = ?`
	_, err := d.db.ExecContext(ctx, query, orgID, termID)
	return err
}

func (d *dataLayer) ListContractObligations(ctx context.Context, orgID int64, contractID int64) ([]ContractObligation, error) {
	query := `
		SELECT 
			id, org_id, contract_id, contract_version_id, obligation_reference,
			title, description, obligation_type, category, responsible_party,
			owner, priority, status,
			DATE_FORMAT(effective_date, '%Y-%m-%d') as effective_date,
			DATE_FORMAT(due_date, '%Y-%m-%d') as due_date,
			completion_date, is_recurring, recurrence_type,
			target_value, target_unit, current_value, warning_threshold, critical_threshold,
			source_document_id, source_term_id, notes, created_by, fulfilled_by,
			created_at, updated_at
		FROM contract_obligations
		WHERE org_id = ? AND contract_id = ?
		ORDER BY due_date ASC, priority DESC
	`
	var obs []ContractObligation
	if err := d.db.SelectContext(ctx, &obs, query, orgID, contractID); err != nil {
		return nil, err
	}
	return obs, nil
}

func (d *dataLayer) GetContractObligationByID(ctx context.Context, orgID int64, obligationID int64) (*ContractObligation, error) {
	query := `
		SELECT 
			id, org_id, contract_id, contract_version_id, obligation_reference,
			title, description, obligation_type, category, responsible_party,
			owner, priority, status,
			DATE_FORMAT(effective_date, '%Y-%m-%d') as effective_date,
			DATE_FORMAT(due_date, '%Y-%m-%d') as due_date,
			completion_date, is_recurring, recurrence_type,
			target_value, target_unit, current_value, warning_threshold, critical_threshold,
			source_document_id, source_term_id, notes, created_by, fulfilled_by,
			created_at, updated_at
		FROM contract_obligations
		WHERE org_id = ? AND id = ?
	`
	var ob ContractObligation
	if err := d.db.GetContext(ctx, &ob, query, orgID, obligationID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}
	return &ob, nil
}

func (d *dataLayer) CreateContractObligation(ctx context.Context, ob *ContractObligation) (int64, error) {
	query := `
		INSERT INTO contract_obligations (
			org_id, contract_id, contract_version_id, obligation_reference,
			title, description, obligation_type, category, responsible_party,
			owner, priority, status, effective_date, due_date, is_recurring,
			recurrence_type, target_value, target_unit, current_value,
			warning_threshold, critical_threshold, source_document_id, source_term_id,
			notes, created_by
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?
		)
	`
	res, err := d.db.ExecContext(ctx, query,
		ob.OrgID, ob.ContractID, ob.ContractVersionID, ob.ObligationReference,
		ob.Title, ob.Description, ob.ObligationType, ob.Category, ob.ResponsibleParty,
		ob.Owner, ob.Priority, ob.Status, ob.EffectiveDate, ob.DueDate, ob.IsRecurring,
		ob.RecurrenceType, ob.TargetValue, ob.TargetUnit, ob.CurrentValue,
		ob.WarningThreshold, ob.CriticalThreshold, ob.SourceDocumentID, ob.SourceTermID,
		ob.Notes, ob.CreatedBy,
	)
	if err != nil {
		return 0, fmt.Errorf("contracts.CreateContractObligation: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	ob.ID = id
	return id, nil
}

func (d *dataLayer) UpdateContractObligation(ctx context.Context, ob *ContractObligation) error {
	query := `
		UPDATE contract_obligations
		SET title = ?, description = ?, owner = ?, priority = ?, status = ?,
		    due_date = ?, current_value = ?, target_value = ?, target_unit = ?,
		    warning_threshold = ?, critical_threshold = ?, notes = ?, updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ?
	`
	_, err := d.db.ExecContext(ctx, query,
		ob.Title, ob.Description, ob.Owner, ob.Priority, ob.Status,
		ob.DueDate, ob.CurrentValue, ob.TargetValue, ob.TargetUnit,
		ob.WarningThreshold, ob.CriticalThreshold, ob.Notes, ob.OrgID, ob.ID,
	)
	return err
}

func (d *dataLayer) UpdateContractObligationStatus(ctx context.Context, orgID int64, obligationID int64, status string) error {
	query := `
		UPDATE contract_obligations
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ?
	`
	_, err := d.db.ExecContext(ctx, query, status, orgID, obligationID)
	return err
}

func (d *dataLayer) FulfillContractObligation(ctx context.Context, orgID int64, obligationID int64, fulfilledBy int64, notes *string) error {
	query := `
		UPDATE contract_obligations
		SET status = 'FULFILLED', completion_date = CURRENT_TIMESTAMP, fulfilled_by = ?,
		    notes = COALESCE(?, notes), updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ?
	`
	_, err := d.db.ExecContext(ctx, query, fulfilledBy, notes, orgID, obligationID)
	return err
}

func (d *dataLayer) WaiveContractObligation(ctx context.Context, orgID int64, obligationID int64, waivedBy int64, notes *string) error {
	query := `
		UPDATE contract_obligations
		SET status = 'WAIVED', completion_date = CURRENT_TIMESTAMP, fulfilled_by = ?,
		    notes = COALESCE(?, notes), updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ?
	`
	_, err := d.db.ExecContext(ctx, query, waivedBy, notes, orgID, obligationID)
	return err
}

func (d *dataLayer) ListContractComplianceEvents(ctx context.Context, orgID int64, contractID int64) ([]ContractComplianceEvent, error) {
	query := `
		SELECT 
			id, org_id, contract_id, contract_obligation_id, related_entity_type,
			related_entity_id, event_type, severity, status, title, description,
			detected_at, resolved_at, resolved_by, resolution_notes, created_at, updated_at
		FROM contract_compliance_events
		WHERE org_id = ? AND contract_id = ?
		ORDER BY detected_at DESC
	`
	var events []ContractComplianceEvent
	if err := d.db.SelectContext(ctx, &events, query, orgID, contractID); err != nil {
		return nil, err
	}
	return events, nil
}

func (d *dataLayer) CreateContractComplianceEventIfNotExists(ctx context.Context, ev *ContractComplianceEvent) error {
	// Check if already open
	var count int
	checkQuery := `
		SELECT COUNT(*) FROM contract_compliance_events
		WHERE org_id = ? AND contract_id = ? AND event_type = ? AND status IN ('OPEN', 'ACKNOWLEDGED', 'IN_PROGRESS')
	`
	if err := d.db.QueryRowContext(ctx, checkQuery, ev.OrgID, ev.ContractID, ev.EventType).Scan(&count); err == nil && count > 0 {
		return nil // Avoid spamming duplicate open events
	}

	query := `
		INSERT INTO contract_compliance_events (
			org_id, contract_id, contract_obligation_id, related_entity_type,
			related_entity_id, event_type, severity, status, title, description,
			detected_at
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			CURRENT_TIMESTAMP
		)
	`
	_, err := d.db.ExecContext(ctx, query,
		ev.OrgID, ev.ContractID, ev.ContractObligationID, ev.RelatedEntityType,
		ev.RelatedEntityID, ev.EventType, ev.Severity, ev.Status, ev.Title, ev.Description,
	)
	return err
}

func (d *dataLayer) ResolveContractComplianceEvent(ctx context.Context, orgID int64, eventID int64, resolvedBy int64, notes string) error {
	query := `
		UPDATE contract_compliance_events
		SET status = 'RESOLVED', resolved_at = CURRENT_TIMESTAMP, resolved_by = ?,
		    resolution_notes = ?, updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ?
	`
	_, err := d.db.ExecContext(ctx, query, resolvedBy, notes, orgID, eventID)
	return err
}

func (d *dataLayer) ListAllOpenComplianceEvents(ctx context.Context, orgID int64) ([]ContractComplianceEvent, error) {
	query := `
		SELECT 
			id, org_id, contract_id, contract_obligation_id, related_entity_type,
			related_entity_id, event_type, severity, status, title, description,
			detected_at, resolved_at, resolved_by, resolution_notes, created_at, updated_at
		FROM contract_compliance_events
		WHERE org_id = ? AND status IN ('OPEN', 'ACKNOWLEDGED', 'IN_PROGRESS')
		ORDER BY CASE severity WHEN 'CRITICAL' THEN 1 WHEN 'WARNING' THEN 2 WHEN 'ATTENTION' THEN 3 ELSE 4 END, detected_at DESC
	`
	var events []ContractComplianceEvent
	if err := d.db.SelectContext(ctx, &events, query, orgID); err != nil {
		return nil, err
	}
	return events, nil
}

func (d *dataLayer) ListContractComplianceRequirements(ctx context.Context, orgID int64, contractID int64) ([]ContractComplianceRequirement, error) {
	query := `
		SELECT 
			id, org_id, contract_id, requirement_type, title, description,
			responsible_party, 
			DATE_FORMAT(valid_from, '%Y-%m-%d') as valid_from,
			DATE_FORMAT(valid_until, '%Y-%m-%d') as valid_until,
			status, evidence_document_id, verification_date, verified_by,
			risk_severity, created_at, updated_at
		FROM contract_compliance_requirements
		WHERE org_id = ? AND contract_id = ?
		ORDER BY created_at ASC
	`
	var reqs []ContractComplianceRequirement
	if err := d.db.SelectContext(ctx, &reqs, query, orgID, contractID); err != nil {
		return nil, err
	}
	return reqs, nil
}

func (d *dataLayer) GetComplianceRequirementByID(ctx context.Context, orgID int64, reqID int64) (*ContractComplianceRequirement, error) {
	query := `
		SELECT 
			id, org_id, contract_id, requirement_type, title, description,
			responsible_party, 
			DATE_FORMAT(valid_from, '%Y-%m-%d') as valid_from,
			DATE_FORMAT(valid_until, '%Y-%m-%d') as valid_until,
			status, evidence_document_id, verification_date, verified_by,
			risk_severity, created_at, updated_at
		FROM contract_compliance_requirements
		WHERE org_id = ? AND id = ?
	`
	var r ContractComplianceRequirement
	if err := d.db.GetContext(ctx, &r, query, orgID, reqID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}
	return &r, nil
}

func (d *dataLayer) CreateContractComplianceRequirement(ctx context.Context, req *ContractComplianceRequirement) (int64, error) {
	query := `
		INSERT INTO contract_compliance_requirements (
			org_id, contract_id, requirement_type, title, description,
			responsible_party, valid_from, valid_until, status, evidence_document_id,
			risk_severity
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?
		)
	`
	res, err := d.db.ExecContext(ctx, query,
		req.OrgID, req.ContractID, req.RequirementType, req.Title, req.Description,
		req.ResponsibleParty, req.ValidFrom, req.ValidUntil, req.Status, req.EvidenceDocumentID,
		req.RiskSeverity,
	)
	if err != nil {
		return 0, fmt.Errorf("contracts.CreateContractComplianceRequirement: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	req.ID = id
	return id, nil
}

func (d *dataLayer) UpdateContractComplianceRequirement(ctx context.Context, req *ContractComplianceRequirement) error {
	query := `
		UPDATE contract_compliance_requirements
		SET requirement_type = ?, title = ?, description = ?, responsible_party = ?,
		    valid_from = ?, valid_until = ?, risk_severity = ?, updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ?
	`
	_, err := d.db.ExecContext(ctx, query,
		req.RequirementType, req.Title, req.Description, req.ResponsibleParty,
		req.ValidFrom, req.ValidUntil, req.RiskSeverity, req.OrgID, req.ID,
	)
	return err
}

func (d *dataLayer) UpdateComplianceRequirementStatus(ctx context.Context, orgID int64, reqID int64, status string) error {
	query := `
		UPDATE contract_compliance_requirements
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ?
	`
	_, err := d.db.ExecContext(ctx, query, status, orgID, reqID)
	return err
}

func (d *dataLayer) VerifyContractComplianceRequirement(ctx context.Context, orgID int64, reqID int64, status string, verifiedBy int64, docID *string, notes *string) error {
	query := `
		UPDATE contract_compliance_requirements
		SET status = ?, verified_by = ?, verification_date = CURRENT_TIMESTAMP,
		    evidence_document_id = COALESCE(?, evidence_document_id),
		    description = COALESCE(?, description), updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ?
	`
	_, err := d.db.ExecContext(ctx, query, status, verifiedBy, docID, notes, orgID, reqID)
	return err
}

// ── Contract Document Import & AI-Assisted Contract Creation Methods ─────────

func (d *dataLayer) CheckContractDuplicates(ctx context.Context, orgID int64, ref string, partyName string) (*ContractDuplicateWarning, error) {
	if ref == "" && partyName == "" {
		return &ContractDuplicateWarning{IsDuplicate: false}, nil
	}

	query := `
		SELECT id, contract_reference, contract_name
		FROM contracts
		WHERE org_id = ? AND status != 'ARCHIVED' AND (
			(? != '' AND LOWER(contract_reference) = LOWER(?))
		)
		LIMIT 1
	`
	var existing struct {
		ID       int64  `db:"id"`
		Ref      string `db:"contract_reference"`
		Name     string `db:"contract_name"`
	}
	err := d.db.GetContext(ctx, &existing, query, orgID, ref, ref)
	if err == nil {
		return &ContractDuplicateWarning{
			IsDuplicate:        true,
			ExistingContractID: &existing.ID,
			ExistingReference:  existing.Ref,
			ExistingName:       existing.Name,
			Message:            fmt.Sprintf("Contract with reference '%s' already exists (Contract #%d: %s).", existing.Ref, existing.ID, existing.Name),
		}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return &ContractDuplicateWarning{IsDuplicate: false}, nil
}

func (d *dataLayer) SaveExtractedContractData(ctx context.Context, orgID int64, docID string, data *ExtractedContractDraft, confidence int) error {
	draftJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal extracted draft: %w", err)
	}

	query := `
		UPDATE contract_documents
		SET extracted_contract_data = ?,
		    extraction_confidence = ?,
		    status = 'AI_EXTRACTED',
		    ai_document_summary = COALESCE(?, ai_document_summary),
		    carrier_name = COALESCE(?, carrier_name),
		    carrier_scac = COALESCE(?, carrier_scac),
		    updated_at = CURRENT_TIMESTAMP
		WHERE org_id = ? AND id = ?
	`
	var scac *string
	if data.CarrierSCAC != nil && *data.CarrierSCAC != "" {
		scac = data.CarrierSCAC
	}
	_, err = d.db.ExecContext(ctx, query, draftJSON, confidence, data.AISummary, data.PartyName, scac, orgID, docID)
	return err
}

func (d *dataLayer) GetExtractedContractData(ctx context.Context, orgID int64, docID string) (*ExtractedContractDraft, error) {
	query := `
		SELECT extracted_contract_data
		FROM contract_documents
		WHERE org_id = ? AND id = ?
	`
	var draftRaw sql.NullString
	if err := d.db.GetContext(ctx, &draftRaw, query, orgID, docID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	if !draftRaw.Valid || draftRaw.String == "" {
		return nil, nil
	}

	var draft ExtractedContractDraft
	if err := json.Unmarshal([]byte(draftRaw.String), &draft); err != nil {
		return nil, fmt.Errorf("unmarshal extracted contract draft: %w", err)
	}
	return &draft, nil
}

func (d *dataLayer) GetCandidateParties(ctx context.Context, orgID int64, searchName string) ([]MatchedPartyCandidate, error) {
	candidates := make([]MatchedPartyCandidate, 0)

	// 1. Fetch carriers
	carrierQuery := `
		SELECT id, name, scac
		FROM carriers
		ORDER BY name ASC
		LIMIT 20
	`
	var carriers []struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
		SCAC string `db:"scac"`
	}
	if err := d.db.SelectContext(ctx, &carriers, carrierQuery); err == nil {
		for _, c := range carriers {
			candidates = append(candidates, MatchedPartyCandidate{
				ID:        c.ID,
				Name:      c.Name,
				PartyType: "CARRIER",
				SCAC:      c.SCAC,
			})
		}
	}

	// 2. Fetch customers for this org
	custQuery := `
		SELECT id, name, code
		FROM customers
		WHERE org_id = ?
		ORDER BY name ASC
		LIMIT 20
	`
	var customers []struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
		Code string `db:"code"`
	}
	if err := d.db.SelectContext(ctx, &customers, custQuery, orgID); err == nil {
		for _, c := range customers {
			candidates = append(candidates, MatchedPartyCandidate{
				ID:        c.ID,
				Name:      c.Name,
				PartyType: "CUSTOMER",
				Code:      c.Code,
			})
		}
	}

	return candidates, nil
}

func (d *dataLayer) ConfirmContractImportTx(ctx context.Context, orgID int64, userID int64, req *ConfirmContractImportRequest) (*Contract, error) {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Insert into contracts
	status := req.Status
	if status == "" {
		status = "DRAFT"
	}
	currency := req.Currency
	if currency == nil || *currency == "" {
		defCurr := "USD"
		currency = &defCurr
	}
	userStr := fmt.Sprintf("%d", userID)

	insertContractQuery := `
		INSERT INTO contracts (
			org_id, contract_reference, contract_name, contract_type,
			party_id, party_name, transport_mode, status, currency,
			contract_value, effective_date, expiry_date, owner,
			description, notes, created_by, updated_by, source_document_id
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?
		)
	`
	res, err := tx.ExecContext(ctx, insertContractQuery,
		orgID, req.ContractReference, req.ContractName, req.ContractType,
		req.PartyID, req.PartyName, req.TransportMode, status, currency,
		req.ContractValue, req.EffectiveDate, req.ExpiryDate, req.Owner,
		req.Description, req.Notes, userStr, userStr, req.DocumentID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert contract: %w", err)
	}

	contractID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get contract id: %w", err)
	}

	// 2. Insert initial Version baseline (v1.0)
	snapshotData := map[string]interface{}{
		"contract_id":        contractID,
		"contract_reference": req.ContractReference,
		"contract_name":      req.ContractName,
		"contract_type":      req.ContractType,
		"party_id":           req.PartyID,
		"party_name":         req.PartyName,
		"transport_mode":     req.TransportMode,
		"status":             status,
		"currency":           currency,
		"contract_value":     req.ContractValue,
		"effective_date":     req.EffectiveDate,
		"expiry_date":        req.ExpiryDate,
		"source_document_id": req.DocumentID,
		"included_terms":     req.IncludedTerms,
		"included_obligations": req.IncludedObligations,
	}
	snapshotJSON, _ := json.Marshal(snapshotData)

	insertVersionQuery := `
		INSERT INTO contract_versions (
			org_id, contract_id, version_number, version_label,
			status, change_summary, contract_snapshot, created_by
		) VALUES (
			?, ?, 1, 'v1.0',
			'DRAFT', 'Initial contract version baseline created via AI Document Import', ?, ?
		)
	`
	if _, err := tx.ExecContext(ctx, insertVersionQuery, orgID, contractID, snapshotJSON, userID); err != nil {
		return nil, fmt.Errorf("insert initial version: %w", err)
	}

	// 3. Log lifecycle event
	lifecycleMeta, _ := json.Marshal(map[string]interface{}{
		"source":             "DOCUMENT_IMPORT",
		"source_document_id": req.DocumentID,
		"contract_reference": req.ContractReference,
	})
	insertEventQuery := `
		INSERT INTO contract_lifecycle_events (
			org_id, contract_id, previous_status, new_status,
			event_type, description, performed_by, metadata
		) VALUES (
			?, ?, NULL, ?,
			'CONTRACT_CREATED', 'Contract created via AI Document Import', ?, ?
		)
	`
	if _, err := tx.ExecContext(ctx, insertEventQuery, orgID, contractID, status, userStr, lifecycleMeta); err != nil {
		return nil, fmt.Errorf("log lifecycle event: %w", err)
	}

	// 4. Update source contract_documents record to associate with newly created contract
	if req.DocumentID != "" {
		updateDocQuery := `
			UPDATE contract_documents
			SET contract_id = ?,
			    document_type = 'MAIN_AGREEMENT',
			    is_current = TRUE,
			    document_status = 'CURRENT',
			    updated_at = CURRENT_TIMESTAMP
			WHERE org_id = ? AND id = ?
		`
		if _, err := tx.ExecContext(ctx, updateDocQuery, contractID, orgID, req.DocumentID); err != nil {
			return nil, fmt.Errorf("update contract document lineage: %w", err)
		}
	}

	// 5. Insert structured terms if provided
	for i, term := range req.IncludedTerms {
		insertTermQuery := `
			INSERT INTO contract_terms (
				org_id, contract_id, term_category, term_key,
				term_title, term_value, value_type, currency,
				display_order, is_critical, created_by
			) VALUES (
				?, ?, ?, ?,
				?, ?, ?, ?,
				?, ?, ?
			)
		`
		valType := term.ValueType
		if valType == "" {
			valType = "STRING"
		}
		if _, err := tx.ExecContext(ctx, insertTermQuery,
			orgID, contractID, term.TermCategory, term.TermKey,
			term.TermTitle, term.TermValue, valType, term.Currency,
			i+1, term.IsCritical, userID,
		); err != nil {
			return nil, fmt.Errorf("insert contract term: %w", err)
		}
	}

	// 6. Insert structured obligations if provided
	for i, ob := range req.IncludedObligations {
		ref := fmt.Sprintf("OBL-%d-%03d", contractID, i+1)
		insertObligationQuery := `
			INSERT INTO contract_obligations (
				org_id, contract_id, obligation_reference, title,
				description, obligation_type, category, responsible_party,
				priority, status, target_value, target_unit,
				notes, source_document_id, created_by
			) VALUES (
				?, ?, ?, ?,
				?, ?, 'OPERATIONAL', ?,
				'HIGH', 'ACTIVE', ?, ?,
				?, ?, ?
			)
		`
		if _, err := tx.ExecContext(ctx, insertObligationQuery,
			orgID, contractID, ref, ob.ObligationTitle,
			ob.PenaltyTerms, ob.ObligationType, ob.PartyResponsible,
			ob.TargetValue, ob.MetricUnit,
			ob.PenaltyTerms, req.DocumentID, userID,
		); err != nil {
			return nil, fmt.Errorf("insert contract obligation: %w", err)
		}
	}

	// 7. If matched party is provided, create link in contract_links
	if req.PartyID > 0 {
		entityType := EntityTypeCarrier
		if req.ContractType == "CUSTOMER_SLA" || req.ContractType == "CUSTOMER_AGREEMENT" {
			entityType = EntityTypeCustomer
		}
		insertLinkQuery := `
			INSERT INTO contract_links (
				org_id, contract_id, entity_type, entity_id,
				relationship_type, notes, created_by
			) VALUES (
				?, ?, ?, ?,
				'PRIMARY_PARTY', 'Linked upon AI Document Import creation', ?
			)
		`
		_, _ = tx.ExecContext(ctx, insertLinkQuery, orgID, contractID, entityType, req.PartyID, userStr)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit contract import transaction: %w", err)
	}

	// Fetch created contract to return
	var created Contract
	fetchQuery := `
		SELECT id, org_id, contract_reference, contract_name, contract_type,
		       party_id, party_name, transport_mode, status, currency,
		       contract_value, effective_date, expiry_date, owner,
		       description, notes, created_by, updated_by, created_at, updated_at,
		       source_document_id
		FROM contracts
		WHERE org_id = ? AND id = ?
	`
	if err := d.db.GetContext(ctx, &created, fetchQuery, orgID, contractID); err != nil {
		return nil, err
	}

	return &created, nil
}

