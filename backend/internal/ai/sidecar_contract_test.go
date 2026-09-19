package ai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/freel/backend/internal/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSidecarClient_LeadScoringContract verifies the contract for lead scoring between Go and Python sidecar.
func TestSidecarClient_LeadScoringContract(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/leads/score-lead", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.NotEmpty(t, r.Header.Get("X-LogisticsHQ-Service-Key"))

		var req ai.ScoreLeadRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, int32(42), req.LeadID)
		assert.Equal(t, int32(1), req.OrgID)
		assert.Equal(t, "Apex Global Logistics", req.CompanyName)
		assert.Equal(t, "corr-lead-42", req.CorrelationID)

		resp := ai.ScoreLeadResponse{
			Score:           88,
			ResearchReport:  "Strong enterprise fit with consistent ocean freight volumes.",
			Confidence:      0.92,
			RecommendedTier: "HOT",
			KeyStrengths:    []string{"High volume", "Verified exporter"},
			RiskFactors:     []string{},
			CorrelationID:   req.CorrelationID,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := ai.NewSidecarClient(ts.URL, "test-secret-key")
	req := &ai.ScoreLeadRequest{
		LeadID:        42,
		OrgID:         1,
		CompanyName:   "Apex Global Logistics",
		CorrelationID: "corr-lead-42",
	}

	res, err := client.ScoreLead(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, int32(88), res.Score)
	assert.Equal(t, "HOT", res.RecommendedTier)
	assert.Equal(t, 0.92, res.Confidence)
	assert.Equal(t, "corr-lead-42", res.CorrelationID)
}

// TestSidecarClient_EmailClassificationContract verifies email classification contract.
func TestSidecarClient_EmailClassificationContract(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/leads/classify-email", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req ai.ClassifyEmailRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "shipper@cargo.com", req.FromEmail)

		resp := ai.ClassifyEmailResponse{
			IsLogisticsRelated: true,
			Intent:             "RFQ_REQUEST",
			Sentiment:          "POSITIVE",
			Reasoning:          "Explicit request for 40ft container freight rates.",
			Confidence:         0.98,
			CorrelationID:      req.CorrelationID,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := ai.NewSidecarClient(ts.URL, "test-secret-key")
	req := &ai.ClassifyEmailRequest{
		OrgID:         1,
		FromEmail:     "shipper@cargo.com",
		Subject:       "Quote: Nhava Sheva to Rotterdam",
		Body:          "Please provide container rates.",
		CorrelationID: "corr-email-1",
	}

	res, err := client.ClassifyEmail(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, res.IsLogisticsRelated)
	assert.Equal(t, "RFQ_REQUEST", res.Intent)
	assert.Equal(t, "POSITIVE", res.Sentiment)
}

// TestSidecarClient_RFQParsingContract verifies RFQ parsing contract.
func TestSidecarClient_RFQParsingContract(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rfq/parse-shipment-request", r.URL.Path)

		origin := "Shanghai"
		dest := "Los Angeles"
		incoterms := "FOB"
		weight := "12000 kg"
		vol := "24 CBM"

		resp := ai.ParseShipmentResponseContract{
			Data: ai.ExtractedShipmentDataContract{
				Origin:      &origin,
				Destination: &dest,
				Incoterms:   &incoterms,
				Weight:      &weight,
				Volume:      &vol,
			},
			ConfidenceScore: 90,
			MissingFields:   []string{},
			CorrelationID:   "corr-rfq-1",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := ai.NewSidecarClient(ts.URL, "test-secret-key")
	req := &ai.ParseShipmentRequestContract{
		OrgID:         1,
		Text:          "Need freight quote from Shanghai to LA FOB 12000kg 24cbm.",
		CorrelationID: "corr-rfq-1",
	}

	res, err := client.ParseShipmentRequest(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 90, res.ConfidenceScore)
	assert.Equal(t, "Shanghai", *res.Data.Origin)
	assert.Equal(t, "Los Angeles", *res.Data.Destination)
}

// TestSidecarClient_MalformedJSONHandling verifies safe error handling when sidecar returns invalid JSON.
func TestSidecarClient_MalformedJSONHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html>502 Bad Gateway</html>"))
	}))
	defer ts.Close()

	client := ai.NewSidecarClient(ts.URL, "test-secret-key")
	req := &ai.ScoreLeadRequest{LeadID: 1, OrgID: 1, CompanyName: "Test"}

	_, err := client.ScoreLead(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode sidecar response")
}

// TestSidecarClient_TimeoutHandling verifies client timeout behavior.
func TestSidecarClient_TimeoutHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := ai.NewSidecarClient(ts.URL, "test-secret-key")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	req := &ai.ScoreLeadRequest{LeadID: 1, OrgID: 1, CompanyName: "Test"}
	_, err := client.ScoreLead(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// TestSidecarClient_TenantIsolationContext verifies org_id scoping on every outgoing contract.
func TestSidecarClient_TenantIsolationContext(t *testing.T) {
	recordedOrgID := int32(0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ai.ScoreLeadRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		recordedOrgID = req.OrgID
		_ = json.NewEncoder(w).Encode(ai.ScoreLeadResponse{Score: 75})
	}))
	defer ts.Close()

	client := ai.NewSidecarClient(ts.URL, "test-secret-key")
	targetOrgID := int32(9876)
	_, err := client.ScoreLead(context.Background(), &ai.ScoreLeadRequest{
		LeadID:      10,
		OrgID:       targetOrgID,
		CompanyName: "Tenant A Corp",
	})
	require.NoError(t, err)
	assert.Equal(t, targetOrgID, recordedOrgID, "OrgID must be strictly propagated to Python sidecar")
}
