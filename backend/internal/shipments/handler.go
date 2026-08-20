package shipments

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// ListShipments handles GET /api/v1/shipments
func (h *Handler) ListShipments(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	list, err := h.svc.ListShipments(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list shipments: "+err.Error(), "DB_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Shipments retrieved", list)
}

// GetShipment handles GET /api/v1/shipments/{id}
func (h *Handler) GetShipment(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id", "INVALID_PARAM")
		return
	}

	sh, err := h.svc.GetShipmentByID(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get shipment: "+err.Error(), "DB_ERROR")
		return
	}
	if sh == nil {
		utils.Error(w, http.StatusNotFound, "Shipment not found", "NOT_FOUND")
		return
	}

	milestones, err := h.svc.GetMilestones(r.Context(), id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get milestones: "+err.Error(), "DB_ERROR")
		return
	}

	exceptions, err := h.svc.GetExceptions(r.Context(), id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get exceptions: "+err.Error(), "DB_ERROR")
		return
	}

	response := map[string]interface{}{
		"shipment":   sh,
		"milestones": milestones,
		"exceptions": exceptions,
	}

	utils.Success(w, http.StatusOK, "Shipment details retrieved", response)
}

// CarrierUpdate handles POST /api/v1/shipments/{id}/carrier-update
func (h *Handler) CarrierUpdate(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id", "INVALID_PARAM")
		return
	}

	var req struct {
		EventID     string `json:"event_id"`
		Description string `json:"description"` // Unstructured update text
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	if req.EventID == "" || req.Description == "" {
		utils.Error(w, http.StatusBadRequest, "event_id and description are required", "MISSING_PARAMS")
		return
	}

	sh, err := h.svc.GetShipmentByID(r.Context(), userCtx.OrgID, id)
	if err != nil || sh == nil {
		utils.Error(w, http.StatusNotFound, "Shipment not found", "NOT_FOUND")
		return
	}

	// Construct normalized event representing the update
	bookingNum := ""
	if sh.BookingNumber != nil {
		bookingNum = *sh.BookingNumber
	}
	event := &NormalizedTrackingEvent{
		EventID:       req.EventID,
		SourceType:    "MANUAL",
		CarrierSCAC:   sh.CarrierSCAC,
		BookingNumber: bookingNum,
		EventTime:     time.Now(),
		Description:   req.Description,
		RawPayload:    json.RawMessage([]byte(fmt.Sprintf(`{"manual_description": %q}`, req.Description))),
	}

	err = h.svc.HandleInboundCarrierEvent(r.Context(), userCtx.OrgID, event)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to enqueue carrier event: "+err.Error(), "QUEUE_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Carrier event enqueued for processing", map[string]interface{}{"event_id": req.EventID})
}

// ResolveException handles POST /api/v1/shipments/exceptions/{id}/resolve
func (h *Handler) ResolveException(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid exception id", "INVALID_PARAM")
		return
	}

	err = h.svc.ResolveException(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to resolve exception: "+err.Error(), "DB_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Exception resolved successfully", nil)
}

// ── INTERNAL ENDPOINTS (Python Sidecar Protected) ──────────────────────────

// GetShipmentInternal handles GET /internal/shipments/{id}
func (h *Handler) GetShipmentInternal(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id", "INVALID_PARAM")
		return
	}

	orgIDStr := r.URL.Query().Get("org_id")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Missing or invalid org_id query parameter", "MISSING_PARAM")
		return
	}

	sh, err := h.svc.GetShipmentByID(r.Context(), orgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get shipment: "+err.Error(), "DB_ERROR")
		return
	}
	if sh == nil {
		utils.Error(w, http.StatusNotFound, "Shipment not found", "NOT_FOUND")
		return
	}

	milestones, err := h.svc.GetMilestones(r.Context(), id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get milestones: "+err.Error(), "DB_ERROR")
		return
	}

	response := map[string]interface{}{
		"shipment":   sh,
		"milestones": milestones,
	}

	utils.Success(w, http.StatusOK, "Internal shipment details retrieved", response)
}

// UpdateMilestoneInternal handles POST /internal/shipments/{id}/milestones
func (h *Handler) UpdateMilestoneInternal(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id", "INVALID_PARAM")
		return
	}

	var req struct {
		MilestoneCode string     `json:"milestone_code"`
		ActualDate    time.Time  `json:"actual_date"`
		Location      *string    `json:"location"`
		Notes         *string    `json:"notes"`
		OrgID         *int64     `json:"org_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	if req.OrgID == nil || *req.OrgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id is required", "MISSING_PARAM")
		return
	}

	err = h.svc.UpdateMilestone(r.Context(), *req.OrgID, id, req.MilestoneCode, &req.ActualDate, req.Location, req.Notes)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update milestone: "+err.Error(), "DB_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Milestone updated", nil)
}

// CreateExceptionInternal handles POST /internal/shipments/{id}/exceptions
func (h *Handler) CreateExceptionInternal(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id", "INVALID_PARAM")
		return
	}

	var req struct {
		ExceptionType string  `json:"exception_type"`
		Severity      string  `json:"severity"`
		Title         string  `json:"title"`
		Description   string  `json:"description"`
		OrgID         *int64  `json:"org_id"`
		SourceEventID *string `json:"source_event_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	if req.OrgID == nil || *req.OrgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id is required", "MISSING_PARAM")
		return
	}

	err = h.svc.CreateException(r.Context(), *req.OrgID, id, req.ExceptionType, req.Severity, req.Title, req.Description, req.SourceEventID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create exception: "+err.Error(), "DB_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Exception created", nil)
}

// CallbackInternal handles POST /internal/operations/callback
func (h *Handler) CallbackInternal(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ShipmentID           int64  `json:"shipment_id"`
		OrgID                int64  `json:"org_id"`
		HasCriticalException bool   `json:"has_critical_exception"`
		AISummary            string `json:"ai_summary"`
		EventID              string `json:"event_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	// 17. Clean handler direct repo access code smell by using CompleteCarrierEvent
	if req.EventID != "" {
		_ = h.svc.CompleteCarrierEvent(r.Context(), req.EventID, req.OrgID, req.ShipmentID, req.HasCriticalException, req.AISummary)
	}

	log.Printf("[Shipment Callback] Processing agent results for Shipment #%d. Critical Exception: %t, Summary: %s",
		req.ShipmentID, req.HasCriticalException, req.AISummary)

	utils.Success(w, http.StatusOK, "Operations callback processed successfully", nil)
}
