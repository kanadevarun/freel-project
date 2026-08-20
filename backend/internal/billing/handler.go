package billing

import (
	"net/http"
	"strconv"

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

// GenerateInvoice handles POST /api/v1/shipments/{id}/billing/invoices/generate
func (h *Handler) GenerateInvoice(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "unauthorized context", "UNAUTHORIZED")
		return
	}

	shipmentIDStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(shipmentIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid shipment id parameter", "INVALID_PARAM")
		return
	}

	invoice, err := h.svc.GenerateInvoiceFromShipment(r.Context(), userCtx.OrgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "SERVICE_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer invoice generated successfully", invoice)
}

// GetBillingWorkspace handles GET /api/v1/shipments/{id}/billing
func (h *Handler) GetBillingWorkspace(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "unauthorized context", "UNAUTHORIZED")
		return
	}

	shipmentIDStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(shipmentIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid shipment id parameter", "INVALID_PARAM")
		return
	}

	invoices, err := h.svc.GetInvoicesByShipment(r.Context(), userCtx.OrgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DB_ERROR")
		return
	}

	profitability, err := h.svc.GetProfitability(r.Context(), userCtx.OrgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PROFITABILITY_ERROR")
		return
	}

	audit, err := h.svc.AuditClosure(r.Context(), userCtx.OrgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "CLOSURE_AUDIT_ERROR")
		return
	}

	response := map[string]interface{}{
		"invoices":      invoices,
		"profitability": profitability,
		"closure_audit": audit,
	}

	utils.Success(w, http.StatusOK, "Billing workspace details retrieved successfully", response)
}

// ApproveInvoice handles POST /api/v1/billing/invoices/{id}/approve
func (h *Handler) ApproveInvoice(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "unauthorized context", "UNAUTHORIZED")
		return
	}

	invoiceID := chi.URLParam(r, "id")
	if invoiceID == "" {
		utils.Error(w, http.StatusBadRequest, "invoice id parameter is required", "MISSING_PARAM")
		return
	}

	err := h.svc.ApproveInvoice(r.Context(), userCtx.OrgID, invoiceID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "SERVICE_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer invoice approved successfully", nil)
}

// PayInvoice handles POST /api/v1/billing/invoices/{id}/pay
func (h *Handler) PayInvoice(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "unauthorized context", "UNAUTHORIZED")
		return
	}

	invoiceID := chi.URLParam(r, "id")
	if invoiceID == "" {
		utils.Error(w, http.StatusBadRequest, "invoice id parameter is required", "MISSING_PARAM")
		return
	}

	err := h.svc.PayInvoice(r.Context(), userCtx.OrgID, invoiceID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "SERVICE_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer invoice paid successfully", nil)
}

// CloseShipment handles POST /api/v1/shipments/{id}/close
func (h *Handler) CloseShipment(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "unauthorized context", "UNAUTHORIZED")
		return
	}

	shipmentIDStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(shipmentIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid shipment id parameter", "INVALID_PARAM")
		return
	}

	err = h.svc.CloseShipment(r.Context(), userCtx.OrgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "CLOSURE_BLOCKED")
		return
	}

	utils.Success(w, http.StatusOK, "Shipment closed successfully", nil)
}
