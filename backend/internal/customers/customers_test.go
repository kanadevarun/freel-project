package customers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDuplicateEngine(t *testing.T) {
	req := CheckDuplicateReq{
		Name:   "Oceanic Exports Pvt Ltd",
		Domain: strPtr("oceanicexports.com"),
		Email:  strPtr("rajesh@oceanicexports.com"),
		TaxID:  strPtr("27AAACO1234B1Z5"),
	}

	candExactTax := Customer{
		ID:           101,
		CustomerCode: "CUST-001",
		Name:         "Oceanic Exports",
		TaxID:        strPtr("27AAACO1234B1Z5"),
		Status:       "ACTIVE",
	}

	scoreTax, reasonTax := EvaluateDuplicateScore(req, candExactTax)
	assert.Equal(t, 100, scoreTax)
	assert.Contains(t, reasonTax, "Tax ID")

	candDomain := Customer{
		ID:           102,
		CustomerCode: "CUST-002",
		Name:         "Oceanic Logistics",
		Domain:       strPtr("oceanicexports.com"),
		Status:       "ACTIVE",
	}

	scoreDomain, reasonDomain := EvaluateDuplicateScore(req, candDomain)
	assert.GreaterOrEqual(t, scoreDomain, 80)
	assert.Contains(t, reasonDomain, "Domain")

	candNoMatch := Customer{
		ID:           103,
		CustomerCode: "CUST-003",
		Name:         "Unrelated Acme Logistics",
		Domain:       strPtr("acmelogistics.com"),
		Status:       "ACTIVE",
	}

	scoreNoMatch, _ := EvaluateDuplicateScore(req, candNoMatch)
	assert.Equal(t, 0, scoreNoMatch)
}

func strPtr(s string) *string {
	return &s
}
