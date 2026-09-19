package collections_automation

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

// Handler provides HTTP endpoints for finance collections automation
type Handler struct {
	svc Service
}

// NewHandler creates a new Handler instance
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts all collections automation endpoints onto the given router
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/overview", h.HandleGetOverview)
	r.Post("/analyze-receivables", h.HandleAnalyzeReceivables)
	r.Post("/prioritize", h.HandlePrioritizeCollections)
	r.Get("/customer-behavior", h.HandleCustomerBehavior)
	r.Get("/recommendations", h.HandleGetRecommendations)
	r.Get("/drafts", h.HandleListDrafts)
	r.Post("/drafts", h.HandleGenerateDraft)
	r.Get("/drafts/{draftId:[0-9]+}", h.HandleGetDraft)
	r.Put("/drafts/{draftId:[0-9]+}", h.HandleUpdateDraft)
	r.Post("/drafts/{draftId:[0-9]+}/submit-approval", h.HandleSubmitApproval)
}

func (h *Handler) getOrgAndInvoiceID(r *http.Request) (int64, int64, int64, error) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		return 0, 0, 0, http.ErrNoCookie
	}
	orgID := userCtx.OrgID
	userID := userCtx.UserID

	invoiceIDStr := chi.URLParam(r, "id")
	invoiceID, err := strconv.ParseInt(invoiceIDStr, 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	return orgID, userID, invoiceID, nil
}

// HandleGetOverview GET /api/v1/invoices/{id}/collections-automation/overview
func (h *Handler) HandleGetOverview(w http.ResponseWriter, r *http.Request) {
	orgID, _, invoiceID, err := h.getOrgAndInvoiceID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	overview, err := h.svc.GetCollectionsOverview(r.Context(), orgID, invoiceID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "Invoice not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "OVERVIEW_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Finance collections overview retrieved", overview)
}

// HandleAnalyzeReceivables POST /api/v1/invoices/{id}/collections-automation/analyze-receivables
func (h *Handler) HandleAnalyzeReceivables(w http.ResponseWriter, r *http.Request) {
	orgID, _, invoiceID, err := h.getOrgAndInvoiceID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	analysis, err := h.svc.AnalyzeReceivablesRisk(r.Context(), orgID, invoiceID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "Invoice not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "ANALYSIS_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Receivables risk analysis generated", analysis)
}

// HandlePrioritizeCollections POST /api/v1/invoices/{id}/collections-automation/prioritize
func (h *Handler) HandlePrioritizeCollections(w http.ResponseWriter, r *http.Request) {
	orgID, _, invoiceID, err := h.getOrgAndInvoiceID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	result, err := h.svc.PrioritizeCollections(r.Context(), orgID, invoiceID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "Invoice not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PRIORITIZATION_ERROR")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}

// HandleCustomerBehavior GET /api/v1/invoices/{id}/collections-automation/customer-behavior
func (h *Handler) HandleCustomerBehavior(w http.ResponseWriter, r *http.Request) {
	orgID, _, invoiceID, err := h.getOrgAndInvoiceID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	result, err := h.svc.AnalyzeCustomerPaymentBehavior(r.Context(), orgID, invoiceID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "Invoice not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "BEHAVIOR_ERROR")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}

// HandleGetRecommendations GET /api/v1/invoices/{id}/collections-automation/recommendations
func (h *Handler) HandleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	orgID, _, invoiceID, err := h.getOrgAndInvoiceID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	result, err := h.svc.GetOperationalRecommendations(r.Context(), orgID, invoiceID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "Invoice not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "RECOMMENDATIONS_ERROR")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}

// HandleListDrafts GET /api/v1/invoices/{id}/collections-automation/drafts
func (h *Handler) HandleListDrafts(w http.ResponseWriter, r *http.Request) {
	orgID, _, invoiceID, err := h.getOrgAndInvoiceID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	drafts, err := h.svc.ListDrafts(r.Context(), orgID, invoiceID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "FETCH_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Collection drafts retrieved", drafts)
}

// HandleGenerateDraft POST /api/v1/invoices/{id}/collections-automation/drafts
func (h *Handler) HandleGenerateDraft(w http.ResponseWriter, r *http.Request) {
	orgID, userID, invoiceID, err := h.getOrgAndInvoiceID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var input GenerateCollectionDraftInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input", "BAD_REQUEST")
		return
	}

	draft, err := h.svc.GenerateCollectionDraft(r.Context(), orgID, invoiceID, userID, input)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "Invoice not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DRAFT_ERROR")
		return
	}

	utils.Success(w, http.StatusCreated, "Collection draft generated", draft)
}

// HandleGetDraft GET /api/v1/invoices/{id}/collections-automation/drafts/{draftId}
func (h *Handler) HandleGetDraft(w http.ResponseWriter, r *http.Request) {
	orgID, _, _, err := h.getOrgAndInvoiceID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	draftIDStr := chi.URLParam(r, "draftId")
	draftID, err := strconv.ParseInt(draftIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid draft ID", "BAD_REQUEST")
		return
	}

	draft, err := h.svc.GetDraft(r.Context(), orgID, draftID)
	if err != nil || draft == nil {
		utils.Error(w, http.StatusNotFound, "Draft not found", "NOT_FOUND")
		return
	}

	utils.Success(w, http.StatusOK, "Collection draft retrieved", draft)
}

// HandleUpdateDraft PUT /api/v1/invoices/{id}/collections-automation/drafts/{draftId}
func (h *Handler) HandleUpdateDraft(w http.ResponseWriter, r *http.Request) {
	orgID, _, _, err := h.getOrgAndInvoiceID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	draftIDStr := chi.URLParam(r, "draftId")
	draftID, err := strconv.ParseInt(draftIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid draft ID", "BAD_REQUEST")
		return
	}

	var input UpdateCollectionDraftInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input", "BAD_REQUEST")
		return
	}

	draft, err := h.svc.UpdateDraft(r.Context(), orgID, draftID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "UPDATE_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Collection draft updated", draft)
}

// HandleSubmitApproval POST /api/v1/invoices/{id}/collections-automation/drafts/{draftId}/submit-approval
func (h *Handler) HandleSubmitApproval(w http.ResponseWriter, r *http.Request) {
	orgID, userID, _, err := h.getOrgAndInvoiceID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	draftIDStr := chi.URLParam(r, "draftId")
	draftID, err := strconv.ParseInt(draftIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid draft ID", "BAD_REQUEST")
		return
	}

	var input SubmitCollectionDraftApprovalInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input", "BAD_REQUEST")
		return
	}

	draft, err := h.svc.SubmitDraftForApproval(r.Context(), orgID, draftID, userID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "APPROVAL_SUBMIT_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Collection draft submitted for managerial approval", draft)
}
