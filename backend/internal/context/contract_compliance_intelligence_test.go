package bcontext

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupContractIntelligenceMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, *defaultService) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	svc := &defaultService{db: sqlxDB}
	return sqlxDB, mock, svc
}

func TestGetContract360ComplianceIntelligence_Validation(t *testing.T) {
	sqlxDB, _, svc := setupContractIntelligenceMockDB(t)
	defer sqlxDB.Close()

	ctx := context.Background()

	// 1. Invalid Org ID
	_, err := svc.GetContract360ComplianceIntelligence(ctx, 0, 10, "corr-1", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid organization id")

	// 2. Invalid Contract ID
	_, err = svc.GetContract360ComplianceIntelligence(ctx, 1, 0, "corr-1", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid contract id")
}

func TestGetContract360ComplianceIntelligence_ActiveContract(t *testing.T) {
	sqlxDB, mock, svc := setupContractIntelligenceMockDB(t)
	defer sqlxDB.Close()

	ctx := context.Background()
	now := time.Now().UTC()
	effDate := now.Add(-30 * 24 * time.Hour)
	expDate := now.Add(90 * 24 * time.Hour)
	targetRenewal := expDate.Add(-15 * 24 * time.Hour)

	// 1. Contract header query
	cntRows := sqlmock.NewRows([]string{
		"id", "org_id", "contract_reference", "contract_name", "contract_type",
		"party_id", "party_name", "transport_mode", "status", "currency",
		"contract_value", "effective_date", "expiry_date", "owner",
		"created_at", "updated_at",
	}).AddRow(
		1, 1, "CNT-2026-001", "Global Logistics Master Agreement", "CUSTOMER_AGREEMENT",
		10, "Apex Global Corp", "Ocean FCL", "ACTIVE", "USD",
		500000.0, effDate, expDate, "John Doe",
		now.Add(-60*24*time.Hour), now,
	)
	mock.ExpectQuery("SELECT (.+) FROM contracts WHERE id = \\? AND org_id = \\?").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(cntRows)

	// 2. Renewal tracking
	renRows := sqlmock.NewRows([]string{"target_completion_date"}).AddRow(targetRenewal)
	mock.ExpectQuery("SELECT target_completion_date FROM contract_renewal_tracking WHERE contract_id = \\? AND org_id = \\?").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(renRows)

	// 3. Document count
	docRows := sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contract_documents WHERE contract_id = \\? AND org_id = \\?").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(docRows)

	// 4. Commercial terms
	termRows := sqlmock.NewRows([]string{
		"id", "term_category", "term_key", "term_title", "term_value", "value_type", "currency", "is_critical",
	}).
		AddRow(1, "COMMERCIAL", "BASE_RATE_40HC", "Base Rate 40ft HC", "2200", "CURRENCY", "USD", true).
		AddRow(2, "OPERATIONAL", "FREE_TIME_DEST", "Demurrage Free Time", "14 calendar days", "STRING", nil, false)
	mock.ExpectQuery("SELECT (.+) FROM contract_terms WHERE contract_id = \\? AND org_id = \\?").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(termRows)

	// 5. Obligations
	obRows := sqlmock.NewRows([]string{
		"id", "obligation_reference", "title", "obligation_type", "responsible_party", "owner",
		"priority", "status", "due_date", "is_recurring",
	}).AddRow(1, "OBL-001", "Transit SLA 28 Days", "SLA", "CARRIER", "Jane Smith", "HIGH", "ACTIVE", expDate, true)
	mock.ExpectQuery("SELECT (.+) FROM contract_obligations WHERE contract_id = \\? AND org_id = \\?").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(obRows)

	// 6. Compliance Requirements
	reqRows := sqlmock.NewRows([]string{
		"id", "requirement_type", "title", "responsible_party", "status", "risk_severity",
		"valid_from", "valid_until", "evidence_document_id",
	}).AddRow(1, "INSURANCE", "Cargo Liability $1M", "CARRIER", "VERIFIED", "HIGH", effDate, expDate, "DOC-123")
	mock.ExpectQuery("SELECT (.+) FROM contract_compliance_requirements WHERE contract_id = \\? AND org_id = \\?").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(reqRows)

	// 7. Compliance Events
	evRows := sqlmock.NewRows([]string{"id", "event_type", "severity", "status", "title", "description", "detected_at"})
	mock.ExpectQuery("SELECT (.+) FROM contract_compliance_events WHERE contract_id = \\? AND org_id = \\?").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(evRows)

	// 8. Connected Links (Quotes, Shipments, Invoices)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contract_links WHERE contract_id = \\? AND org_id = \\? AND linked_entity_type = 'QUOTATION'").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(3))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contract_links WHERE contract_id = \\? AND org_id = \\? AND linked_entity_type = 'SHIPMENT'").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(5))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contract_links WHERE contract_id = \\? AND org_id = \\? AND linked_entity_type = 'INVOICE'").
		WithArgs(int64(1), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(4))

	intel, err := svc.GetContract360ComplianceIntelligence(ctx, 1, 1, "corr-test-cnt-1", 10)
	require.NoError(t, err)
	require.NotNil(t, intel)

	// Assertions
	assert.Equal(t, int64(1), intel.ContractID)
	assert.Equal(t, "CNT-2026-001", intel.Identity.ContractReference)
	assert.Equal(t, "Apex Global Corp", intel.Identity.PartyName)
	assert.Equal(t, "CUSTOMER", intel.Identity.PartyType)
	assert.True(t, intel.Identity.IsActive)
	assert.False(t, intel.Identity.IsExpired)
	assert.False(t, intel.Identity.IsExpiringSoon)
	assert.Equal(t, 2, intel.Identity.DocumentCount)
	assert.Equal(t, 100, intel.Identity.CompletenessScore)

	// Commercial terms
	assert.Equal(t, 2, intel.Commercial.TotalTermsCount)
	assert.Equal(t, 1, intel.Commercial.CriticalTermsCount)
	assert.Equal(t, 1, len(intel.Commercial.FreeTimeDemurrageTerms))

	// Obligations
	assert.Equal(t, 1, intel.Obligations.TotalObligations)
	assert.Equal(t, 1, intel.Obligations.ActiveObligations)
	assert.False(t, intel.Obligations.HasUnownedObligation)

	// Compliance
	assert.Equal(t, 1, intel.Compliance.TotalRequirements)
	assert.Equal(t, 1, intel.Compliance.VerifiedRequirements)
	assert.Equal(t, "COMPLIANT", intel.Compliance.ComplianceStatus)

	// Risk rating
	assert.Equal(t, "LOW", intel.RiskIndicators.OverallRiskRating)
	assert.True(t, intel.ReadOnly)
	assert.NotEmpty(t, intel.AISummary.ExecutiveSummary)
	assert.Contains(t, intel.AISummary.Citations[0], "CNT-2026-001")
}

func TestGetContract360ComplianceIntelligence_ExpiredWithMissingDoc(t *testing.T) {
	sqlxDB, mock, svc := setupContractIntelligenceMockDB(t)
	defer sqlxDB.Close()

	ctx := context.Background()
	now := time.Now().UTC()
	effDate := now.Add(-180 * 24 * time.Hour)
	expDate := now.Add(-10 * 24 * time.Hour) // expired 10 days ago

	// 1. Contract query
	cntRows := sqlmock.NewRows([]string{
		"id", "org_id", "contract_reference", "contract_name", "contract_type",
		"party_id", "party_name", "transport_mode", "status", "currency",
		"contract_value", "effective_date", "expiry_date", "owner",
		"created_at", "updated_at",
	}).AddRow(
		2, 1, "CNT-EXPIRED-02", "Sea Freight Service Agreement", "CARRIER_SERVICE",
		25, "Maersk Ocean Line", "Ocean FCL", "ACTIVE", "USD",
		120000.0, effDate, expDate, "Risk Manager",
		now.Add(-200*24*time.Hour), now,
	)
	mock.ExpectQuery("SELECT (.+) FROM contracts WHERE id = \\? AND org_id = \\?").
		WithArgs(int64(2), int64(1)).
		WillReturnRows(cntRows)

	// 2. Renewal tracking - None
	mock.ExpectQuery("SELECT target_completion_date FROM contract_renewal_tracking WHERE contract_id = \\? AND org_id = \\?").
		WithArgs(int64(2), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"target_completion_date"}))

	// 3. Document count - 0 (missing doc!)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contract_documents WHERE contract_id = \\? AND org_id = \\?").
		WithArgs(int64(2), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))

	// 4. Terms - None
	mock.ExpectQuery("SELECT (.+) FROM contract_terms WHERE contract_id = \\? AND org_id = \\?").
		WithArgs(int64(2), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "term_category", "term_key", "term_title", "term_value", "value_type", "currency", "is_critical"}))

	// 5. Obligations - None
	mock.ExpectQuery("SELECT (.+) FROM contract_obligations WHERE contract_id = \\? AND org_id = \\?").
		WithArgs(int64(2), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "obligation_reference", "title", "obligation_type", "responsible_party", "owner", "priority", "status", "due_date", "is_recurring"}))

	// 6. Compliance Requirements - None
	mock.ExpectQuery("SELECT (.+) FROM contract_compliance_requirements WHERE contract_id = \\? AND org_id = \\?").
		WithArgs(int64(2), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "requirement_type", "title", "responsible_party", "status", "risk_severity", "valid_from", "valid_until", "evidence_document_id"}))

	// 7. Compliance Events - 1 High Risk Event
	evRows := sqlmock.NewRows([]string{
		"id", "event_type", "severity", "status", "title", "description", "detected_at",
	}).AddRow(101, "CERT_LAPSED", "CRITICAL", "OPEN", "Carrier Operating License Lapsed", "State licensing verification failed", now.Add(-5*24*time.Hour))
	mock.ExpectQuery("SELECT (.+) FROM contract_compliance_events WHERE contract_id = \\? AND org_id = \\?").
		WithArgs(int64(2), int64(1)).
		WillReturnRows(evRows)

	// 8. Links
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contract_links WHERE contract_id = \\? AND org_id = \\? AND linked_entity_type = 'QUOTATION'").
		WithArgs(int64(2), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contract_links WHERE contract_id = \\? AND org_id = \\? AND linked_entity_type = 'SHIPMENT'").
		WithArgs(int64(2), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contract_links WHERE contract_id = \\? AND org_id = \\? AND linked_entity_type = 'INVOICE'").
		WithArgs(int64(2), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))

	intel, err := svc.GetContract360ComplianceIntelligence(ctx, 1, 2, "corr-test-cnt-2", 10)
	require.NoError(t, err)
	require.NotNil(t, intel)

	// Assertions
	assert.True(t, intel.Identity.IsExpired)
	assert.False(t, intel.Identity.IsActive)
	assert.Equal(t, 10, intel.Identity.DaysExpired)
	assert.True(t, intel.Identity.HasMissingDocument)
	assert.Equal(t, "CARRIER", intel.Identity.PartyType)
	assert.Equal(t, "ACTION_REQUIRED", intel.Compliance.ComplianceStatus)
	assert.Equal(t, 1, intel.Compliance.HighSeverityEventsCount)

	// Risk rating should be CRITICAL (expired + missing doc + high severity event)
	assert.Equal(t, "CRITICAL", intel.RiskIndicators.OverallRiskRating)
	assert.GreaterOrEqual(t, intel.RiskIndicators.OverallRiskScore, 70)
	assert.True(t, len(intel.RiskIndicators.RiskFactors) >= 3)
}

func TestGetContractCoverageForEntity_Shipment(t *testing.T) {
	sqlxDB, mock, svc := setupContractIntelligenceMockDB(t)
	defer sqlxDB.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// 1. Shipment query
	shRows := sqlmock.NewRows([]string{
		"id", "carrier_scac", "carrier_name", "customer_id", "customer_name", "created_at",
	}).AddRow(int64(100), "MAEU", "Maersk", int64(50), "Beta Freight Corp", now.Add(-48*time.Hour))
	mock.ExpectQuery("SELECT (.+) FROM shipments s").
		WithArgs(int64(100), int64(1)).
		WillReturnRows(shRows)

	// 2. Matching active contract query
	eff := now.Add(-60 * 24 * time.Hour)
	exp := now.Add(60 * 24 * time.Hour)
	cntRows := sqlmock.NewRows([]string{
		"id", "contract_reference", "contract_name", "currency", "transport_mode", "status", "effective_date", "expiry_date",
	}).AddRow(10, "CNT-BETA-01", "Beta Ocean Master Agreement", "USD", "Ocean FCL", "ACTIVE", eff, exp)
	mock.ExpectQuery("SELECT (.+) FROM contracts WHERE org_id = \\? AND party_id = \\? AND status = 'ACTIVE'").
		WithArgs(int64(1), int64(50)).
		WillReturnRows(cntRows)

	eval, err := svc.GetContractCoverageForEntity(ctx, 1, "SHIPMENT", 100, "corr-test-cov-1", 10)
	require.NoError(t, err)
	require.NotNil(t, eval)

	assert.Equal(t, "SHIPMENT", eval.TargetEntityType)
	assert.Equal(t, int64(100), eval.TargetEntityID)
	assert.Equal(t, "SH-100", eval.TargetReference)
	assert.True(t, eval.HasApplicableContract)
	assert.Equal(t, int64(10), *eval.MatchedContractID)
	assert.Equal(t, "FULLY_COVERED", eval.CoverageStatus)
	assert.True(t, eval.PartyMatches)
	assert.True(t, eval.ModeMatches)
	assert.True(t, eval.DateWithinValidity)
	assert.Empty(t, eval.CoverageGaps)
}

func TestGetOrgContractComplianceSummary(t *testing.T) {
	sqlxDB, mock, svc := setupContractIntelligenceMockDB(t)
	defer sqlxDB.Close()

	ctx := context.Background()

	// 1. Total counts
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contracts WHERE org_id = \\?").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(12))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contracts WHERE org_id = \\? AND status = 'ACTIVE'").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(8))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contracts WHERE org_id = \\? AND status = 'DRAFT'").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contracts WHERE org_id = \\? AND status = 'EXPIRED'").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(2))

	// 2. Missing docs & expiry
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contracts c WHERE c.org_id = \\? AND NOT EXISTS").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contracts WHERE org_id = \\? AND expiry_date IS NULL").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))

	// 3. Expiring contracts within 30 days
	expRows := sqlmock.NewRows([]string{
		"id", "contract_reference", "contract_name", "contract_type", "party_id", "party_name",
		"transport_mode", "status", "currency", "contract_value", "effective_date", "expiry_date", "owner",
		"created_at", "updated_at",
	}).AddRow(
		5, "CNT-EXP-05", "Contract Near Expiry", "CUSTOMER_AGREEMENT", 12, "Delta Corp",
		"Air", "ACTIVE", "USD", 50000.0, time.Now().Add(-100*time.Hour), time.Now().Add(5*24*time.Hour), "Owner",
		time.Now().Add(-200*time.Hour), time.Now(),
	)
	mock.ExpectQuery("SELECT (.+) FROM contracts WHERE org_id = \\? AND status = 'ACTIVE' AND expiry_date IS NOT NULL").
		WithArgs(int64(1)).
		WillReturnRows(expRows)

	// 4. Compliance requirements counts
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contract_compliance_requirements WHERE org_id = \\?").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(15))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contract_compliance_requirements WHERE org_id = \\? AND status = 'VERIFIED'").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(12))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contract_compliance_requirements WHERE org_id = \\? AND status = 'PENDING'").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(3))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contract_compliance_requirements WHERE org_id = \\? AND valid_until IS NOT NULL AND valid_until < CURDATE\\(\\)").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))

	// 5. Compliance events counts
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contract_compliance_events WHERE org_id = \\? AND status != 'RESOLVED'").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM contract_compliance_events WHERE org_id = \\? AND status != 'RESOLVED' AND severity IN \\('HIGH', 'CRITICAL'\\)").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))

	// 6. High issues query
	mock.ExpectQuery("SELECT (.+) FROM contract_compliance_events WHERE org_id = \\? AND status != 'RESOLVED' AND severity IN \\('HIGH', 'CRITICAL'\\)").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "event_type", "severity", "status", "title", "description", "detected_at"}))

	summary, err := svc.GetOrgContractComplianceSummary(ctx, 1, "corr-org-summary-1", 10)
	require.NoError(t, err)
	require.NotNil(t, summary)

	assert.Equal(t, 12, summary.TotalContracts)
	assert.Equal(t, 8, summary.ActiveContracts)
	assert.Equal(t, 1, summary.ExpiringContracts30d)
	assert.Equal(t, 1, summary.CriticalExpiring7d)
	assert.Equal(t, 15, summary.TotalComplianceRequirements)
	assert.Equal(t, 12, summary.VerifiedRequirements)
	assert.True(t, summary.ReadOnly)
}
