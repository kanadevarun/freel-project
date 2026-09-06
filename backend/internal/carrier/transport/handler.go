package transport

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/freel/backend/internal/carrier/domain"
	"github.com/freel/backend/internal/carrier/service"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

type CarrierHandler struct {
	svc service.CarrierService
}

func NewCarrierHandler(svc service.CarrierService) *CarrierHandler {
	return &CarrierHandler{svc: svc}
}

// HandleListProviders returns the global supported shipping lines.
// GET /api/v1/carrier-providers
func (h *CarrierHandler) HandleListProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.svc.GetProviders(r.Context())
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "FETCH_PROVIDERS_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Carrier providers retrieved successfully", providers)
}

// HandleListIntegrations returns all carrier integrations configured for the authenticated tenant.
// GET /api/v1/carrier-integrations
func (h *CarrierHandler) HandleListIntegrations(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	integrations, err := h.svc.GetIntegrations(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "FETCH_INTEGRATIONS_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Carrier integrations retrieved successfully", integrations)
}

// HandleGetIntegration returns a single integration by ID for the authenticated tenant.
// GET /api/v1/carrier-integrations/{id}
func (h *CarrierHandler) HandleGetIntegration(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid carrier integration ID", "INVALID_ID")
		return
	}

	integration, err := h.svc.GetIntegration(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error(), "INTEGRATION_NOT_FOUND")
		return
	}
	utils.Success(w, http.StatusOK, "Carrier integration retrieved successfully", integration)
}

// HandleConnectCarrier registers a new carrier integration for the tenant.
// POST /api/v1/carrier-integrations
func (h *CarrierHandler) HandleConnectCarrier(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	var req domain.ConnectCarrierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_PAYLOAD")
		return
	}

	created, err := h.svc.ConnectCarrier(r.Context(), userCtx.OrgID, userCtx.UserID, req)
	if err != nil {
		if err == service.ErrDuplicateIntegration {
			utils.Error(w, http.StatusConflict, "An active integration already exists for this carrier in this environment", "DUPLICATE_INTEGRATION")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "CONNECT_CARRIER_ERROR")
		return
	}
	utils.Success(w, http.StatusCreated, "Carrier connected successfully", created)
}

// HandleUpdateCarrier updates an existing integration's credentials, capabilities, or environment.
// PUT /api/v1/carrier-integrations/{id}
func (h *CarrierHandler) HandleUpdateCarrier(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid carrier integration ID", "INVALID_ID")
		return
	}

	var req domain.UpdateCarrierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_PAYLOAD")
		return
	}

	updated, err := h.svc.UpdateCarrier(r.Context(), userCtx.OrgID, userCtx.UserID, id, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "UPDATE_CARRIER_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Carrier integration updated successfully", updated)
}

// HandleToggleCarrier toggles active/disabled state of an integration.
// PATCH /api/v1/carrier-integrations/{id}/toggle
func (h *CarrierHandler) HandleToggleCarrier(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid carrier integration ID", "INVALID_ID")
		return
	}

	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_PAYLOAD")
		return
	}

	toggled, err := h.svc.ToggleCarrier(r.Context(), userCtx.OrgID, userCtx.UserID, id, body.IsActive)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "TOGGLE_CARRIER_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Carrier status updated successfully", toggled)
}

// HandleDisconnectCarrier removes the tenant's carrier integration.
// DELETE /api/v1/carrier-integrations/{id}
func (h *CarrierHandler) HandleDisconnectCarrier(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid carrier integration ID", "INVALID_ID")
		return
	}

	if err := h.svc.DisconnectCarrier(r.Context(), userCtx.OrgID, userCtx.UserID, id); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "DISCONNECT_CARRIER_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Carrier disconnected successfully", map[string]bool{"disconnected": true})
}

// HandleTestConnection tests connectivity and credentials for an existing integration.
// POST /api/v1/carrier-integrations/{id}/test
func (h *CarrierHandler) HandleTestConnection(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid carrier integration ID", "INVALID_ID")
		return
	}

	res, err := h.svc.TestConnection(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "TEST_CONNECTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, res.Message, res)
}

// HandleTestDirectConnection tests credentials before saving an integration record.
// POST /api/v1/carrier-integrations/test-direct
func (h *CarrierHandler) HandleTestDirectConnection(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	var req domain.TestDirectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "BAD_PAYLOAD")
		return
	}

	res, err := h.svc.TestDirectConnection(r.Context(), userCtx.OrgID, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "TEST_CONNECTION_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, res.Message, res)
}

// HandleSyncCarrier triggers manual synchronization for an integration.
// POST /api/v1/carrier-integrations/{id}/sync
func (h *CarrierHandler) HandleSyncCarrier(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid carrier integration ID", "INVALID_ID")
		return
	}

	var req domain.SyncNowRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	jobView, err := h.svc.SyncNow(r.Context(), userCtx.OrgID, id, req)
	if err != nil {
		if errors.Is(err, service.ErrSyncInProgress) {
			utils.Error(w, http.StatusConflict, "Sync already in progress.", "SYNC_IN_PROGRESS")
			return
		}
		if errors.Is(err, service.ErrIntegrationNotFound) {
			utils.Error(w, http.StatusNotFound, "Carrier integration not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "SYNC_CARRIER_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "Carrier synchronization completed", jobView)
}

// HandleGetSyncHistory returns paginated sync logs for an integration.
// GET /api/v1/carrier-integrations/{id}/sync-history
func (h *CarrierHandler) HandleGetSyncHistory(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid carrier integration ID", "INVALID_ID")
		return
	}

	limit := 20
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	offset := 0
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	items, total, err := h.svc.GetSyncHistory(r.Context(), userCtx.OrgID, id, limit, offset)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "FETCH_SYNC_HISTORY_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Sync history retrieved successfully", map[string]interface{}{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// HandleGetSyncJob returns a single sync job log with technical diagnostics.
// GET /api/v1/carrier-integrations/{id}/sync-history/{syncId}
func (h *CarrierHandler) HandleGetSyncJob(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	syncIDStr := chi.URLParam(r, "syncId")
	syncID, err := strconv.ParseInt(syncIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid sync job ID", "INVALID_ID")
		return
	}

	job, err := h.svc.GetSyncJob(r.Context(), userCtx.OrgID, syncID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error(), "SYNC_JOB_NOT_FOUND")
		return
	}

	utils.Success(w, http.StatusOK, "Sync job details retrieved successfully", job)
}

// HandleGetIntegrationHealth returns operational health metrics for an integration.
// GET /api/v1/carrier-integrations/{id}/health
func (h *CarrierHandler) HandleGetIntegrationHealth(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid carrier integration ID", "INVALID_ID")
		return
	}

	health, err := h.svc.GetIntegrationHealth(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error(), "INTEGRATION_NOT_FOUND")
		return
	}

	utils.Success(w, http.StatusOK, "Carrier integration health retrieved successfully", health)
}

// HandleInboundWebhook handles external unauthenticated carrier webhooks securely.
// POST /api/v1/carrier-integrations/webhooks/{providerCode}
func (h *CarrierHandler) HandleInboundWebhook(w http.ResponseWriter, r *http.Request) {
	providerCode := chi.URLParam(r, "providerCode")
	if providerCode == "" {
		utils.Error(w, http.StatusBadRequest, "Carrier provider code is required", "MISSING_PROVIDER")
		return
	}

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Failed to read request body", "BAD_REQUEST_BODY")
		return
	}
	defer r.Body.Close()

	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	evt, err := h.svc.ProcessWebhook(r.Context(), providerCode, rawBody, headers)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "WEBHOOK_PROCESSING_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Carrier webhook accepted and queued for ingestion", map[string]interface{}{
		"received":       true,
		"event_id":       evt.ID,
		"status":         evt.Status,
		"correlation_id": evt.CorrelationID,
	})
}

// HandleGetTracking fetches real-time normalized tracking milestones from the carrier adapter.
// POST /api/v1/carrier-integrations/{id}/tracking
func (h *CarrierHandler) HandleGetTracking(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid carrier integration ID", "INVALID_ID")
		return
	}

	var req domain.TrackingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid tracking request payload", "BAD_PAYLOAD")
		return
	}

	result, err := h.svc.GetTracking(r.Context(), userCtx.OrgID, id, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "CARRIER_TRACKING_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Tracking events retrieved successfully", result)
}

// HandleGetRates queries live or contracted rates from the carrier adapter.
// POST /api/v1/carrier-integrations/{id}/rates
func (h *CarrierHandler) HandleGetRates(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid carrier integration ID", "INVALID_ID")
		return
	}

	var req domain.RateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid rate request payload", "BAD_PAYLOAD")
		return
	}

	result, err := h.svc.GetRates(r.Context(), userCtx.OrgID, id, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "CARRIER_RATES_FAILED")
		return
	}
	utils.Success(w, http.StatusOK, "Carrier rates retrieved successfully", result)
}

// HandleCreateBooking initiates a container space booking with the carrier adapter.
// POST /api/v1/carrier-integrations/{id}/booking
func (h *CarrierHandler) HandleCreateBooking(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid carrier integration ID", "INVALID_ID")
		return
	}

	var req domain.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid booking request payload", "BAD_PAYLOAD")
		return
	}

	result, err := h.svc.CreateBooking(r.Context(), userCtx.OrgID, id, req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "CARRIER_BOOKING_FAILED")
		return
	}
	utils.Success(w, http.StatusCreated, "Carrier booking initiated successfully", result)
}
