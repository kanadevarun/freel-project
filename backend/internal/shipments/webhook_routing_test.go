package shipments_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/shipments"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookRoutingTenantSafety(t *testing.T) {
	// Connect to local test DB
	dbURL := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true"
	db, err := database.Connect(dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Clean up any test pollution
	_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id IN (8881, 8882)")
	_, _ = db.Exec("DELETE FROM organizations WHERE id IN (8881, 8882)")
	_, _ = db.Exec("DELETE FROM carrier_tracking_events WHERE org_id IN (8881, 8882)")

	// Seed test organizations and carriers
	_, err = db.Exec(`
		INSERT INTO organizations (id, name, created_at, updated_at) 
		VALUES (8881, 'Org A', NOW(), NOW()), (8882, 'Org B', NOW(), NOW())
		ON DUPLICATE KEY UPDATE name = VALUES(name)
	`)
	require.NoError(t, err)

	_, _ = db.Exec("INSERT INTO carriers (scac, name) VALUES ('MAEU', 'Maersk Line'), ('MSC', 'MSC') ON DUPLICATE KEY UPDATE name = VALUES(name)")

	// Seed carrier integrations (active MAEU for Org A & B, inactive MSC for Org A)
	resA, err := db.Exec(`
		INSERT INTO carrier_integrations (org_id, carrier_scac, is_active, created_at, updated_at)
		VALUES (8881, 'MAEU', 1, NOW(), NOW())
	`)
	require.NoError(t, err)
	integrationID_A, _ := resA.LastInsertId()

	resB, err := db.Exec(`
		INSERT INTO carrier_integrations (org_id, carrier_scac, is_active, created_at, updated_at)
		VALUES (8882, 'MAEU', 1, NOW(), NOW())
	`)
	require.NoError(t, err)
	integrationID_B, _ := resB.LastInsertId()

	resC, err := db.Exec(`
		INSERT INTO carrier_integrations (org_id, carrier_scac, is_active, created_at, updated_at)
		VALUES (8881, 'MSC', 0, NOW(), NOW())
	`)
	require.NoError(t, err)
	integrationID_C, _ := resC.LastInsertId()

	// Clean up DB records at the end
	defer func() {
		_, _ = db.Exec("DELETE FROM carrier_integrations WHERE org_id IN (8881, 8882)")
		_, _ = db.Exec("DELETE FROM organizations WHERE id IN (8881, 8882)")
		_, _ = db.Exec("DELETE FROM carrier_tracking_events WHERE org_id IN (8881, 8882)")
	}()

	// Setup environment variables for organization-scoped credentials and capabilities
	t.Setenv("CARRIER_MAEU_API_KEY_8881", "org-a-key-secure")
	t.Setenv("CARRIER_MAEU_BASE_URL_8881", "https://api.maersk.com/org-a")
	t.Setenv("CARRIER_MAEU_CAPABILITIES_8881", "TRACKING,WEBHOOK")

	t.Setenv("CARRIER_MAEU_API_KEY_8882", "org-b-key-secure")
	t.Setenv("CARRIER_MAEU_BASE_URL_8882", "https://api.maersk.com/org-b")
	t.Setenv("CARRIER_MAEU_CAPABILITIES_8882", "TRACKING,WEBHOOK")

	t.Setenv("CARRIER_MSC_API_KEY_8881", "org-a-msc-key")
	t.Setenv("CARRIER_MSC_BASE_URL_8881", "https://api.msc.com/org-a")
	t.Setenv("CARRIER_MSC_CAPABILITIES_8881", "TRACKING") // MSC does not have WEBHOOK capability

	// Instantiate the handler
	repo := shipments.NewRepository(db)
	svc := shipments.NewService(repo, db, nil, "http://localhost:8080")
	handler := shipments.NewHandler(svc)

	// Setup helper to perform request
	performWebhookRequest := func(carrier string, integrationID string, payload []byte, signature string) *httptest.ResponseRecorder {
		req, _ := http.NewRequest("POST", fmt.Sprintf("/webhooks/carriers/%s/%s", carrier, integrationID), bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		if signature != "" {
			req.Header.Set("X-Mock-Signature", signature)
		}

		// Inject chi routing params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("carrier", carrier)
		rctx.URLParams.Add("integration_id", integrationID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		handler.InboundWebhook(rr, req)
		return rr
	}

	// 1. Correct Webhook Signature & Routing for Org A
	payloadA := []byte(`{
		"event_id": "EVT-ORG-A-1",
		"booking_number": "BK-A123",
		"milestone": "DEPARTED",
		"description": "Vessel departed Nhava Sheva",
		"location": "INNSA"
	}`)
	sigA := fmt.Sprintf("%x", sha256.Sum256(payloadA))

	rrA := performWebhookRequest("MAEU", fmt.Sprintf("%d", integrationID_A), payloadA, sigA)
	assert.Equal(t, http.StatusOK, rrA.Code)

	// 2. Correct Webhook Signature & Routing for Org B (Simultaneously using MAEU)
	payloadB := []byte(`{
		"event_id": "EVT-ORG-B-1",
		"booking_number": "BK-B456",
		"milestone": "IN_TRANSIT",
		"description": "Vessel departed SGSIN",
		"location": "SGSIN"
	}`)
	sigB := fmt.Sprintf("%x", sha256.Sum256(payloadB))

	rrB := performWebhookRequest("MAEU", fmt.Sprintf("%d", integrationID_B), payloadB, sigB)
	assert.Equal(t, http.StatusOK, rrB.Code)

	// Verify Tenant Isolation: Check that events were written to the correct org_id
	var orgAEventCount int
	err = db.Get(&orgAEventCount, "SELECT count(*) FROM carrier_tracking_events WHERE org_id = 8881 AND event_id = 'EVT-ORG-A-1'")
	require.NoError(t, err)
	assert.Equal(t, 1, orgAEventCount, "Org A event should be successfully routed to Org A (8881)")

	var orgBEventCount int
	err = db.Get(&orgBEventCount, "SELECT count(*) FROM carrier_tracking_events WHERE org_id = 8882 AND event_id = 'EVT-ORG-B-1'")
	require.NoError(t, err)
	assert.Equal(t, 1, orgBEventCount, "Org B event should be successfully routed to Org B (8882)")

	// 3. Ensuring Org A webhook cannot route to Org B (using Org A's integration ID, event should not hit Org B)
	var crossOrgBEventCount int
	err = db.Get(&crossOrgBEventCount, "SELECT count(*) FROM carrier_tracking_events WHERE org_id = 8882 AND event_id = 'EVT-ORG-A-1'")
	require.NoError(t, err)
	assert.Equal(t, 0, crossOrgBEventCount, "Org A event must not be mapped to Org B (8882)")

	// 4. Invalid Integration ID
	rrInvalidID := performWebhookRequest("MAEU", "9999999", payloadA, sigA)
	assert.Equal(t, http.StatusNotFound, rrInvalidID.Code)

	// 5. Inactive Integration (CMA CGM or inactive MSC for Org A)
	rrInactive := performWebhookRequest("MSC", fmt.Sprintf("%d", integrationID_C), []byte(`{"event_id":"evt-3"}`), "sig")
	assert.Equal(t, http.StatusBadRequest, rrInactive.Code)

	// 6. Wrong Signature
	rrWrongSig := performWebhookRequest("MAEU", fmt.Sprintf("%d", integrationID_A), payloadA, "wrong-signature-value")
	assert.Equal(t, http.StatusUnauthorized, rrWrongSig.Code)

	// 7. Unsupported Capability (WEBHOOK is not enabled in CARRIER_MSC_CAPABILITIES_8881)
	// Enable MSC integration temporarily to test capability restriction
	_, _ = db.Exec("UPDATE carrier_integrations SET is_active = 1 WHERE id = ?", integrationID_C)
	payloadMSC := []byte(`{"event_id": "evt-msc-1"}`)
	sigMSC := fmt.Sprintf("%x", sha256.Sum256(payloadMSC))
	rrMSC := performWebhookRequest("MSC", fmt.Sprintf("%d", integrationID_C), payloadMSC, sigMSC)
	assert.Equal(t, http.StatusUnauthorized, rrMSC.Code, "Should return StatusUnauthorized (unsupported webhook capability)")
	
	var mscEventCount int
	_ = db.Get(&mscEventCount, "SELECT count(*) FROM carrier_tracking_events WHERE org_id = 8881 AND event_id = 'evt-msc-1'")
	assert.Equal(t, 0, mscEventCount, "Event with unsupported capability must not be processed")

	// 8. Strict Production Check (Integration ID is required in production)
	t.Setenv("APP_ENV", "production")
	reqProd, _ := http.NewRequest("POST", "/webhooks/carriers/MAEU", bytes.NewReader(payloadA))
	reqProd.Header.Set("Content-Type", "application/json")
	reqProd.Header.Set("X-Mock-Signature", sigA)

	rctxProd := chi.NewRouteContext()
	rctxProd.URLParams.Add("carrier", "MAEU")
	reqProd = reqProd.WithContext(context.WithValue(reqProd.Context(), chi.RouteCtxKey, rctxProd))

	rrProd := httptest.NewRecorder()
	handler.InboundWebhook(rrProd, reqProd)
	assert.Equal(t, http.StatusBadRequest, rrProd.Code)
	
	var prodBody map[string]interface{}
	_ = json.Unmarshal(rrProd.Body.Bytes(), &prodBody)
	assert.Equal(t, "MISSING_INTEGRATION_ID", prodBody["error"].(map[string]interface{})["code"])
}
