package shipments

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/carrier/adapters"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

// InboundWebhookHandler handles inbound push updates from carrier webhooks
func (h *Handler) InboundWebhook(w http.ResponseWriter, r *http.Request) {
	carrierParam := chi.URLParam(r, "carrier")
	if carrierParam == "" {
		utils.Error(w, http.StatusBadRequest, "carrier identifier is required", "MISSING_PARAM")
		return
	}

	integrationIDParam := chi.URLParam(r, "integration_id")
	var orgID int64
	var dbCarrierSCAC string
	var isActive bool
	var err error

	if integrationIDParam != "" {
		integrationID, parseErr := strconv.ParseInt(integrationIDParam, 10, 64)
		if parseErr != nil {
			utils.Error(w, http.StatusBadRequest, "invalid integration_id format", "INVALID_PARAM")
			return
		}

		err = h.svc.(*service).db.QueryRowContext(r.Context(), `
			SELECT org_id, carrier_scac, is_active FROM carrier_integrations 
			WHERE id = ? LIMIT 1
		`, integrationID).Scan(&orgID, &dbCarrierSCAC, &isActive)
		if err != nil {
			utils.Error(w, http.StatusNotFound, "carrier integration not configured or inactive", "NOT_FOUND")
			return
		}

		if !isActive {
			utils.Error(w, http.StatusBadRequest, "carrier integration is inactive", "INACTIVE_INTEGRATION")
			return
		}

		// Verify carrier SCAC matches the request parameter
		scacUpper := strings.ToUpper(carrierParam)
		dbScacUpper := strings.ToUpper(dbCarrierSCAC)
		isMaerskMatch := (scacUpper == "MAEU" || scacUpper == "MSK") && (dbScacUpper == "MAEU" || dbScacUpper == "MSK")
		if scacUpper != dbScacUpper && !isMaerskMatch {
			utils.Error(w, http.StatusBadRequest, "carrier mismatch for specified integration ID", "CARRIER_MISMATCH")
			return
		}
	} else {
		// Production strictness: resolve without integration_id is not allowed
		if os.Getenv("APP_ENV") == "production" {
			utils.Error(w, http.StatusBadRequest, "production error: integration_id is required in webhook URL path", "MISSING_INTEGRATION_ID")
			return
		}

		// Fallback lookup in local development/mock verification context
		dbErr := h.svc.(*service).db.GetContext(r.Context(), &orgID, `
			SELECT org_id FROM carrier_integrations 
			WHERE carrier_scac = ? AND is_active = 1 LIMIT 1
		`, carrierParam)
		if dbErr != nil {
			userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
			if ok && userCtx.OrgID > 0 {
				orgID = userCtx.OrgID
			} else {
				_ = h.svc.(*service).db.GetContext(r.Context(), &orgID, "SELECT id FROM organizations LIMIT 1")
			}
		}
	}

	if orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "unable to resolve org_id for integration credentials", "RESOLVE_ERROR")
		return
	}

	// Read body bytes for signature verification and parsing
	var body []byte
	if r.Body != nil {
		defer r.Body.Close()
		body, err = io.ReadAll(r.Body)
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "failed to read request body", "READ_ERROR")
			return
		}
	}

	// 2. Resolve WebhookProvider from CarrierAdapterFactory (Group 1 / 19 fix)
	adapter, err := adapters.GetWebhookProvider(h.svc.(*service).db, orgID, carrierParam)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "unsupported carrier adapter: "+err.Error(), "NOT_FOUND")
		return
	}

	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// Verify webhook authenticity
	if err := adapter.VerifyWebhookSignature(body, headers); err != nil {
		utils.Error(w, http.StatusUnauthorized, "invalid webhook signature: "+err.Error(), "UNAUTHORIZED")
		return
	}

	// Parse payload into generic TrackingEvent struct
	rawEv, err := adapter.ParseWebhookPayload(body)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "failed to parse webhook payload: "+err.Error(), "INVALID_PAYLOAD")
		return
	}

	// Convert raw TrackingEvent to system NormalizedTrackingEvent contract
	normalized := Normalize(*rawEv, carrierParam, "WEBHOOK")

	err = h.svc.HandleInboundCarrierEvent(r.Context(), orgID, &normalized)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to handle carrier event: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "carrier event received", map[string]string{
		"event_id": normalized.EventID,
	})
}
