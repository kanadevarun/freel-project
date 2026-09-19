package predictions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// ContractComplianceData represents authoritative facts for a contract's compliance, terms, and documentation
type ContractComplianceData struct {
	ContractID                 int64
	OrgID                      int64
	ContractReference          string
	ContractName               string
	ContractType               string
	PartyName                  string
	Status                     string
	Currency                   string
	ContractValue              float64
	EffectiveDate              string
	ExpiryDate                 string
	DaysUntilExpiry            int
	IsExpired                  bool
	IsExpiringSoon             bool
	ComplianceStatus           string
	RiskScore                  float64
	HasPendingDocument         bool
	PendingDocumentName        string
	MissingDocuments           []string
	ExtractedClauses           []map[string]interface{}
	Discrepancies              []map[string]interface{}
	StructuredDiscrepancyCount int
	HasUnresolvedDisputes      bool
	ContextMap                 map[string]interface{}
}

// ContractComplianceDataProvider queries real persistent MariaDB data with strict tenant isolation
type ContractComplianceDataProvider struct {
	db *sql.DB
}

// NewContractComplianceDataProvider instantiates the provider
func NewContractComplianceDataProvider(db *sql.DB) *ContractComplianceDataProvider {
	return &ContractComplianceDataProvider{db: db}
}

// FetchContractComplianceData queries contracts, contract_documents, and ai_contract_compliance_reviews
func (p *ContractComplianceDataProvider) FetchContractComplianceData(ctx context.Context, orgID int64, contractID int64) (*ContractComplianceData, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	query := `
		SELECT 
			id, org_id, COALESCE(contract_reference, ''), COALESCE(contract_name, ''),
			COALESCE(contract_type, ''), COALESCE(party_name, ''), COALESCE(status, 'DRAFT'),
			COALESCE(currency, 'USD'), COALESCE(contract_value, 0.0),
			effective_date, expiry_date
		FROM contracts
		WHERE id = ? AND org_id = ?
	`
	var (
		ref, name, cType, party, status, currency string
		val                                       float64
		effDate, expDate                          sql.NullTime
	)

	err := p.db.QueryRowContext(ctx, query, contractID, orgID).Scan(
		&contractID, &orgID, &ref, &name, &cType, &party, &status,
		&currency, &val, &effDate, &expDate,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("contract %d not found for organization %d", contractID, orgID)
	} else if err != nil {
		return nil, fmt.Errorf("failed querying contract: %w", err)
	}

	now := time.Now().UTC()
	daysUntilExpiry := 999
	isExpired := false
	isExpiringSoon := false
	effDateStr := ""
	expDateStr := ""

	if effDate.Valid {
		effDateStr = effDate.Time.Format("2006-01-02")
	}
	if expDate.Valid {
		expDateStr = expDate.Time.Format("2006-01-02")
		daysUntilExpiry = int(expDate.Time.Sub(now).Hours() / 24)
		if daysUntilExpiry < 0 {
			isExpired = true
		} else if daysUntilExpiry <= 30 {
			isExpiringSoon = true
		}
	}

	// 2. Query compliance review from ai_contract_compliance_reviews
	var (
		complianceStatus string = "PENDING"
		riskScore        float64 = 0.0
		extractedClausesRaw sql.NullString
		discrepanciesRaw    sql.NullString
		obligationsRaw      sql.NullString
		recommendationsRaw  sql.NullString
	)

	compQuery := `
		SELECT 
			COALESCE(compliance_status, 'PENDING'),
			COALESCE(risk_score, 0.0),
			extracted_clauses,
			structured_discrepancies,
			compliance_obligations,
			recommendations
		FROM ai_contract_compliance_reviews
		WHERE contract_id = ? AND org_id = ?
		ORDER BY id DESC LIMIT 1
	`
	_ = p.db.QueryRowContext(ctx, compQuery, contractID, orgID).Scan(
		&complianceStatus, &riskScore, &extractedClausesRaw,
		&discrepanciesRaw, &obligationsRaw, &recommendationsRaw,
	)

	var extractedClauses []map[string]interface{}
	if extractedClausesRaw.Valid && extractedClausesRaw.String != "" {
		_ = json.Unmarshal([]byte(extractedClausesRaw.String), &extractedClauses)
	}

	var discrepancies []map[string]interface{}
	if discrepanciesRaw.Valid && discrepanciesRaw.String != "" {
		_ = json.Unmarshal([]byte(discrepanciesRaw.String), &discrepancies)
	}

	// 3. Query contract_documents for pending or anomaly-flagged files
	var (
		hasPendingDocument  bool   = false
		pendingDocumentName string = ""
	)
	docQuery := `
		SELECT file_name
		FROM contract_documents
		WHERE org_id = ? AND status IN ('PENDING_REVIEW', 'FAILED', 'PENDING_EXTRACTION')
		ORDER BY created_at DESC LIMIT 1
	`
	_ = p.db.QueryRowContext(ctx, docQuery, orgID).Scan(&pendingDocumentName)
	if pendingDocumentName != "" {
		hasPendingDocument = true
	}

	// 4. Missing mandatory documents check
	missingDocs := []string{}
	if len(extractedClauses) > 0 && complianceStatus == "REVIEW_REQUIRED" {
		missingDocs = append(missingDocs, "Certificate of Cargo Liability Insurance (COI)")
	}

	contextMap := map[string]interface{}{
		"contract_id":                  contractID,
		"contract_reference":           ref,
		"contract_name":                name,
		"contract_type":                cType,
		"party_name":                   party,
		"status":                       status,
		"currency":                     currency,
		"contract_value":               val,
		"effective_date":               effDateStr,
		"expiry_date":                  expDateStr,
		"days_until_expiry":            daysUntilExpiry,
		"is_expired":                   isExpired,
		"is_expiring_soon":             isExpiringSoon,
		"compliance_status":            complianceStatus,
		"risk_score":                   riskScore,
		"has_pending_document":         hasPendingDocument,
		"pending_document_name":        pendingDocumentName,
		"missing_documents":            missingDocs,
		"extracted_clauses":            extractedClauses,
		"discrepancies":                discrepancies,
		"structured_discrepancy_count": len(discrepancies),
		"has_unresolved_disputes":      false,
	}

	return &ContractComplianceData{
		ContractID:                 contractID,
		OrgID:                      orgID,
		ContractReference:          ref,
		ContractName:               name,
		ContractType:               cType,
		PartyName:                  party,
		Status:                     status,
		Currency:                   currency,
		ContractValue:              val,
		EffectiveDate:              effDateStr,
		ExpiryDate:                 expDateStr,
		DaysUntilExpiry:            daysUntilExpiry,
		IsExpired:                  isExpired,
		IsExpiringSoon:             isExpiringSoon,
		ComplianceStatus:           complianceStatus,
		RiskScore:                  riskScore,
		HasPendingDocument:         hasPendingDocument,
		PendingDocumentName:        pendingDocumentName,
		MissingDocuments:           missingDocs,
		ExtractedClauses:           extractedClauses,
		Discrepancies:              discrepancies,
		StructuredDiscrepancyCount: len(discrepancies),
		HasUnresolvedDisputes:      false,
		ContextMap:                 contextMap,
	}, nil
}
