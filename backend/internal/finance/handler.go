package finance

import (
	"encoding/json"
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

// IngestInvoice handles POST /api/v1/shipments/{id}/finance/invoices/upload
func (h *Handler) IngestInvoice(w http.ResponseWriter, r *http.Request) {
	shipmentIDStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(shipmentIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	var req struct {
		InvoiceNumber string `json:"invoice_number"`
		VendorName    string `json:"vendor_name"`
		S3Key         string `json:"s3_key"`
		FileName      string `json:"file_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}
	if req.InvoiceNumber == "" || req.S3Key == "" {
		utils.Error(w, http.StatusBadRequest, "invoice_number and s3_key are required", "MISSING_PARAM")
		return
	}

	inv, err := h.svc.IngestInvoice(r.Context(), userCtx.OrgID, shipmentID, req.InvoiceNumber, req.VendorName, req.S3Key, req.FileName)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to ingest invoice: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Invoice ingested and audit queued", inv)
}

// GetFinanceWorkspace handles GET /api/v1/shipments/{id}/finance
func (h *Handler) GetFinanceWorkspace(w http.ResponseWriter, r *http.Request) {
	shipmentIDStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(shipmentIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	data, err := h.svc.GetFinanceWorkspaceInternal(r.Context(), userCtx.OrgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve finance workspace: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Finance workspace data retrieved", data)
}

// ListAllInvoices handles GET /api/v1/invoices
func (h *Handler) ListAllInvoices(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	invoices, err := h.svc.GetInvoicesByOrg(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve organization invoices: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Retrieved organization invoices successfully", invoices)
}

// ResolveDiscrepancy handles POST /api/v1/finance/discrepancies/{id}/resolve
func (h *Handler) ResolveDiscrepancy(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid discrepancy id parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	if err := h.svc.ResolveDiscrepancy(r.Context(), userCtx.OrgID, id, userCtx.UserID); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to resolve discrepancy: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Finance discrepancy resolved successfully", nil)
}

// ApproveInvoice handles POST /api/v1/finance/invoices/{id}/approve
func (h *Handler) ApproveInvoice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice id parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	if err := h.svc.ApproveInvoice(r.Context(), userCtx.OrgID, id); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to approve invoice: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Invoice approved successfully", nil)
}

// CallbackInternal handles POST /internal/finance/callback from the Python sidecar.
func (h *Handler) CallbackInternal(w http.ResponseWriter, r *http.Request) {
	var req FinanceCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	if req.OrgID <= 0 || req.ShipmentID <= 0 || req.InvoiceID == "" {
		utils.Error(w, http.StatusBadRequest, "org_id, shipment_id and invoice_id are required", "MISSING_PARAM")
		return
	}

	if err := h.svc.CompleteReconciliation(r.Context(), &req); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to complete reconciliation: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Finance reconciliation callback complete", nil)
}

// GetFinanceWorkspaceInternal handles GET /internal/shipments/{id}/finance (for the sidecar to fetch quote data)
func (h *Handler) GetFinanceWorkspaceInternal(w http.ResponseWriter, r *http.Request) {
	shipmentIDStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(shipmentIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id parameter", "INVALID_PARAM")
		return
	}

	orgIDStr := r.URL.Query().Get("org_id")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Missing or invalid org_id query parameter", "MISSING_PARAM")
		return
	}

	data, err := h.svc.GetFinanceWorkspaceInternal(r.Context(), orgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve internal finance workspace: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Finance workspace data retrieved", data)
}
