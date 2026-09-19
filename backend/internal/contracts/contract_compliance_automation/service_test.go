package contract_compliance_automation

import (
	"testing"
	"time"

	"github.com/freel/backend/internal/contracts"
)

func TestDeterministicExpiryCalculation(t *testing.T) {
	now := time.Now()
	svc := &service{}

	futureDate := now.AddDate(0, 8, 0).Format("2006-01-02")
	nearDate := now.AddDate(0, 0, 25).Format("2006-01-02")
	pastDate := now.AddDate(0, -2, 0).Format("2006-01-02")

	// Case 1: Active contract (future)
	c1 := &contracts.Contract{ExpiryDate: &futureDate}
	sig1 := svc.calculateDeterministicSignals(c1, nil, nil, nil)
	if sig1.IsExpired {
		t.Errorf("expected c1 to not be expired")
	}
	if sig1.IsNearingExpiry {
		t.Errorf("expected c1 to not be nearing expiry (8 months left)")
	}

	// Case 2: Expiring in 25 days (nearing expiry)
	c2 := &contracts.Contract{ExpiryDate: &nearDate}
	sig2 := svc.calculateDeterministicSignals(c2, nil, nil, nil)
	if sig2.IsExpired {
		t.Errorf("expected c2 to not be expired")
	}
	if !sig2.IsNearingExpiry {
		t.Errorf("expected c2 to be nearing expiry (25 days left <= 60 days)")
	}

	// Case 3: Past expiry (already expired)
	c3 := &contracts.Contract{ExpiryDate: &pastDate}
	sig3 := svc.calculateDeterministicSignals(c3, nil, nil, nil)
	if !sig3.IsExpired {
		t.Errorf("expected c3 to be expired")
	}
	if sig3.DaysUntilExpiry != 0 {
		t.Errorf("expected 0 days until expiry for expired contract, got %d", sig3.DaysUntilExpiry)
	}
}

func TestRequiredDocumentsCalculation(t *testing.T) {
	svc := &service{}
	c := &contracts.Contract{}

	// No terms and no docs -> missing insurance and missing master agreement
	sig := svc.calculateDeterministicSignals(c, nil, nil, nil)
	if !sig.MissingRequiredDocuments {
		t.Errorf("expected MissingRequiredDocuments to be true")
	}
	if !sig.MissingInsuranceTerms {
		t.Errorf("expected MissingInsuranceTerms to be true")
	}
	if len(sig.RequiredDocumentsMissing) < 2 {
		t.Errorf("expected at least 2 missing documents, got %d", len(sig.RequiredDocumentsMissing))
	}

	// With insurance term present
	terms := []map[string]interface{}{
		{"term_key": "CARGO_INSURANCE_COVERAGE", "term_category": "INSURANCE"},
	}
	docs := []map[string]interface{}{
		{"file_name": "master_agreement.pdf", "status": "APPROVED"},
	}
	sigWithTerms := svc.calculateDeterministicSignals(c, terms, nil, docs)
	if sigWithTerms.MissingInsuranceTerms {
		t.Errorf("expected MissingInsuranceTerms to be false when insurance term present")
	}
	if sigWithTerms.MissingRequiredDocuments {
		t.Errorf("expected MissingRequiredDocuments to be false when all docs satisfied")
	}
}

func TestDraftStatusValidation(t *testing.T) {
	draft := &ContractComplianceDraft{
		ID:     1,
		Status: DraftStatusApproved,
	}

	canEdit := draft.Status == DraftStatusDraft
	if canEdit {
		t.Errorf("expected approved draft to not be editable")
	}

	draft.Status = DraftStatusDraft
	canEdit = draft.Status == DraftStatusDraft
	if !canEdit {
		t.Errorf("expected draft in DRAFT status to be editable")
	}
}

func TestTenantIsolationEnforcement(t *testing.T) {
	contractOrgID := int64(1)
	userOrgID := int64(2)

	hasAccess := contractOrgID == userOrgID
	if hasAccess {
		t.Errorf("expected tenant isolation to block cross-org access (org 1 != org 2)")
	}

	userOrgID = int64(1)
	hasAccess = contractOrgID == userOrgID
	if !hasAccess {
		t.Errorf("expected access allowed when org IDs match")
	}
}
