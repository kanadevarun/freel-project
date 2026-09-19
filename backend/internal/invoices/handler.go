package invoices

import (
	"encoding/json"
	"errors"
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

// ListInvoices handles GET /api/v1/invoices
func (h *Handler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	params := ListInvoiceParams{
		PrimaryTab: r.URL.Query().Get("primary_tab"),
		Status:     r.URL.Query().Get("status"),
		Search:     r.URL.Query().Get("search"),
	}

	if id, err := strconv.ParseInt(r.URL.Query().Get("customer_id"), 10, 64); err == nil {
		params.CustomerID = id
	}
	if id, err := strconv.ParseInt(r.URL.Query().Get("shipment_id"), 10, 64); err == nil {
		params.ShipmentID = id
	}
	if id, err := strconv.ParseInt(r.URL.Query().Get("booking_id"), 10, 64); err == nil {
		params.BookingID = id
	}
	if id, err := strconv.ParseInt(r.URL.Query().Get("quotation_id"), 10, 64); err == nil {
		params.QuotationID = id
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			params.Page = p
		}
	}
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil {
			params.PageSize = ps
		}
	}

	list, total, err := h.svc.GetInvoices(r.Context(), userCtx.OrgID, params, userCtx.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch invoices: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	response := map[string]interface{}{
		"invoices": list,
		"total":    total,
		"page":     params.Page,
		"pageSize": params.PageSize,
	}

	utils.Success(w, http.StatusOK, "Invoices retrieved successfully", response)
}

// GetKPIStats handles GET /api/v1/invoices/kpi-stats
func (h *Handler) GetKPIStats(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	stats, err := h.svc.GetKPIStats(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch invoice KPI statistics: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Invoice KPI statistics retrieved successfully", stats)
}

// GetInvoiceByID handles GET /api/v1/invoices/{id}
func (h *Handler) GetInvoiceByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice ID parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	inv, err := h.svc.GetInvoiceByID(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Invoice not found", "NOT_FOUND")
		return
	}

	utils.Success(w, http.StatusOK, "Invoice details retrieved successfully", inv)
}

// CreateInvoice handles POST /api/v1/invoices
func (h *Handler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	var input CreateInvoiceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	userName := h.svc.ResolveUserName(r.Context(), userCtx.UserID)
	inv, err := h.svc.CreateInvoice(r.Context(), userCtx.OrgID, userCtx.UserID, userName, input)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusCreated, "Invoice created successfully", inv)
}

// UpdateDraftInvoice handles PUT /api/v1/invoices/{id}
func (h *Handler) UpdateDraftInvoice(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice ID parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	var input CreateInvoiceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	userName := h.svc.ResolveUserName(r.Context(), userCtx.UserID)
	inv, err := h.svc.UpdateDraftInvoice(r.Context(), userCtx.OrgID, id, input, userName)
	if err != nil {
		if errors.Is(err, ErrInvoiceNotFound) {
			utils.Error(w, http.StatusNotFound, "Invoice not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Draft invoice updated successfully", inv)
}

// IssueInvoice handles POST /api/v1/invoices/{id}/issue
func (h *Handler) IssueInvoice(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice ID parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	userName := h.svc.ResolveUserName(r.Context(), userCtx.UserID)
	inv, err := h.svc.IssueInvoice(r.Context(), userCtx.OrgID, id, userName)
	if err != nil {
		if errors.Is(err, ErrInvoiceNotFound) {
			utils.Error(w, http.StatusNotFound, "Invoice not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Invoice issued successfully", inv)
}

// SubmitForApproval handles POST /api/v1/invoices/{id}/submit-approval
func (h *Handler) SubmitForApproval(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice ID parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	userName := h.svc.ResolveUserName(r.Context(), userCtx.UserID)
	inv, err := h.svc.SubmitForApproval(r.Context(), userCtx.OrgID, id, userName)
	if err != nil {
		if errors.Is(err, ErrInvoiceNotFound) {
			utils.Error(w, http.StatusNotFound, "Invoice not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Invoice submitted for approval", inv)
}

// UpdateInvoiceStatus handles PUT /api/v1/invoices/{id}/status
func (h *Handler) UpdateInvoiceStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice ID parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Status == "" {
		utils.Error(w, http.StatusBadRequest, "status is required", "INVALID_PAYLOAD")
		return
	}

	userName := h.svc.ResolveUserName(r.Context(), userCtx.UserID)
	if err := h.svc.UpdateInvoiceStatus(r.Context(), userCtx.OrgID, id, req.Status, userName); err != nil {
		if errors.Is(err, ErrInvoiceNotFound) {
			utils.Error(w, http.StatusNotFound, "Invoice not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Invoice status updated successfully", nil)
}

// ToggleBookmark handles POST /api/v1/invoices/{id}/bookmark
func (h *Handler) ToggleBookmark(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice ID parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	isBookmarked, err := h.svc.ToggleBookmark(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to toggle bookmark: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Invoice bookmark updated successfully", map[string]bool{"bookmarked": isBookmarked})
}

// CancelInvoice handles POST /api/v1/invoices/{id}/cancel
func (h *Handler) CancelInvoice(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice ID parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	userName := h.svc.ResolveUserName(r.Context(), userCtx.UserID)
	if err := h.svc.CancelInvoice(r.Context(), userCtx.OrgID, id, req.Reason, userName); err != nil {
		if errors.Is(err, ErrInvoiceNotFound) {
			utils.Error(w, http.StatusNotFound, "Invoice not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Invoice cancelled successfully", nil)
}

// RecordPayment handles POST /api/v1/invoices/{id}/payments
func (h *Handler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice ID parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	var input RecordPaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid payment request payload", "INVALID_PAYLOAD")
		return
	}

	userName := h.svc.ResolveUserName(r.Context(), userCtx.UserID)
	inv, err := h.svc.RecordPayment(r.Context(), userCtx.OrgID, id, input, userName)
	if err != nil {
		if errors.Is(err, ErrInvoiceNotFound) {
			utils.Error(w, http.StatusNotFound, "Invoice not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusCreated, "Payment recorded successfully", inv)
}

// GetInvoicePayments handles GET /api/v1/invoices/{id}/payments
func (h *Handler) GetInvoicePayments(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice ID parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	inv, err := h.svc.GetInvoiceByID(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Invoice not found", "NOT_FOUND")
		return
	}

	utils.Success(w, http.StatusOK, "Invoice payments retrieved successfully", inv.Payments)
}

// ListAllPayments handles GET /api/v1/invoices/payments
func (h *Handler) ListAllPayments(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	payments, err := h.svc.GetAllPayments(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve payments: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Payments retrieved successfully", payments)
}

// UploadDocument handles POST /api/v1/invoices/{id}/documents
func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice ID parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	var req struct {
		DocumentName string `json:"document_name"`
		FileSize     string `json:"file_size"`
		FileType     string `json:"file_type"`
		S3Key        string `json:"s3_key"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	doc, err := h.svc.AddDocument(r.Context(), userCtx.OrgID, id, req.DocumentName, req.FileSize, req.FileType, req.S3Key)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusCreated, "Document attached to invoice successfully", doc)
}

// ──────────────────────────────────────────────────────────────────
// Debit Note HTTP Handlers
// ──────────────────────────────────────────────────────────────────

// CreateDebitNote handles POST /api/v1/debit-notes
func (h *Handler) CreateDebitNote(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	var input CreateDebitNoteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	dn, err := h.svc.CreateDebitNote(r.Context(), userCtx.OrgID, userCtx.UserID, input)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusCreated, "Debit note created successfully", dn)
}

// ListDebitNotes handles GET /api/v1/debit-notes
func (h *Handler) ListDebitNotes(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	params := ListDebitNoteParams{
		Status: r.URL.Query().Get("status"),
		Search: r.URL.Query().Get("search"),
	}
	if id, err := strconv.ParseInt(r.URL.Query().Get("customer_id"), 10, 64); err == nil {
		params.CustomerID = id
	}
	if id, err := strconv.ParseInt(r.URL.Query().Get("shipment_id"), 10, 64); err == nil {
		params.ShipmentID = id
	}
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil {
		params.Page = p
	}
	if ps, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil {
		params.PageSize = ps
	}

	list, total, err := h.svc.ListDebitNotes(r.Context(), userCtx.OrgID, params)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch debit notes: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Debit notes retrieved successfully", map[string]interface{}{
		"debit_notes": list,
		"total":       total,
		"page":        params.Page,
		"page_size":   params.PageSize,
	})
}

// GetDebitNoteKPIStats handles GET /api/v1/debit-notes/kpi-stats
func (h *Handler) GetDebitNoteKPIStats(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	stats, err := h.svc.GetDebitNoteKPIStats(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch debit note KPI statistics: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Debit note KPI statistics retrieved successfully", stats)
}

// GetDebitNote handles GET /api/v1/debit-notes/{id}
func (h *Handler) GetDebitNote(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid debit note ID parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	dn, err := h.svc.GetDebitNote(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
		return
	}

	utils.Success(w, http.StatusOK, "Debit note retrieved successfully", dn)
}

// IssueDebitNote handles POST /api/v1/debit-notes/{id}/issue
func (h *Handler) IssueDebitNote(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid debit note ID parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	if err := h.svc.IssueDebitNote(r.Context(), userCtx.OrgID, id); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Debit note issued successfully", nil)
}

// VoidDebitNote handles POST /api/v1/debit-notes/{id}/void
func (h *Handler) VoidDebitNote(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid debit note ID parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization context", "UNAUTHORIZED")
		return
	}

	if err := h.svc.VoidDebitNote(r.Context(), userCtx.OrgID, id); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "BAD_REQUEST")
		return
	}

	utils.Success(w, http.StatusOK, "Debit note voided successfully", nil)
}

