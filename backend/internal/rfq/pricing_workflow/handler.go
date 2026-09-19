package pricing_workflow

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for RFQ Pricing Workflow
type Handler struct {
	svc Service
}

// NewHandler creates a new Handler instance
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts all pricing workflow routes onto the provided chi Router
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/overview", h.HandleGetOverview)
	r.Post("/extract-requirements", h.HandleExtractRequirements)
	r.Post("/pricing-preview", h.HandlePricingPreview)
	r.Post("/drafts", h.HandleGenerateDraft)
	r.Get("/drafts/{draftId:[0-9]+}", h.HandleGetDraft)
	r.Put("/drafts/{draftId:[0-9]+}", h.HandleUpdateDraft)
	r.Post("/drafts/{draftId:[0-9]+}/submit-approval", h.HandleSubmitApproval)
}

func (h *Handler) getOrgAndRFQID(r *http.Request) (int64, int64, int64, error) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		return 0, 0, 0, http.ErrNoCookie
	}
	orgID := userCtx.OrgID
	userID := userCtx.UserID

	rfqIDStr := chi.URLParam(r, "id")
	rfqID, err := strconv.ParseInt(rfqIDStr, 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	return orgID, userID, rfqID, nil
}

// HandleGetOverview GET /api/v1/rfqs/{id}/pricing-workflow/overview
func (h *Handler) HandleGetOverview(w http.ResponseWriter, r *http.Request) {
	orgID, _, rfqID, err := h.getOrgAndRFQID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	overview, err := h.svc.GetWorkflowOverview(r.Context(), orgID, rfqID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "RFQ not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "WORKFLOW_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Workflow overview retrieved", overview)
}

// HandleExtractRequirements POST /api/v1/rfqs/{id}/pricing-workflow/extract-requirements
func (h *Handler) HandleExtractRequirements(w http.ResponseWriter, r *http.Request) {
	orgID, _, rfqID, err := h.getOrgAndRFQID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	extraction, err := h.svc.ExtractRequirements(r.Context(), orgID, rfqID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "RFQ not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXTRACTION_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Requirements extracted successfully", extraction)
}

// HandlePricingPreview POST /api/v1/rfqs/{id}/pricing-workflow/pricing-preview
func (h *Handler) HandlePricingPreview(w http.ResponseWriter, r *http.Request) {
	orgID, _, rfqID, err := h.getOrgAndRFQID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var req PricingPreviewRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	preview, err := h.svc.CalculatePricingPreview(r.Context(), orgID, rfqID, req)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "RFQ not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PRICING_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Pricing preview calculated", preview)
}

// HandleGenerateDraft POST /api/v1/rfqs/{id}/pricing-workflow/drafts
func (h *Handler) HandleGenerateDraft(w http.ResponseWriter, r *http.Request) {
	orgID, userID, rfqID, err := h.getOrgAndRFQID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var input GenerateDraftInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input", "BAD_REQUEST")
		return
	}

	draft, err := h.svc.GenerateDraft(r.Context(), orgID, rfqID, userID, input)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "RFQ not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DRAFT_ERROR")
		return
	}

	utils.Success(w, http.StatusCreated, "Quotation draft generated successfully", draft)
}

// HandleGetDraft GET /api/v1/rfqs/{id}/pricing-workflow/drafts/{draftId}
func (h *Handler) HandleGetDraft(w http.ResponseWriter, r *http.Request) {
	orgID, _, _, err := h.getOrgAndRFQID(r)
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

	utils.Success(w, http.StatusOK, "Quotation draft retrieved", draft)
}

// HandleUpdateDraft PUT /api/v1/rfqs/{id}/pricing-workflow/drafts/{draftId}
func (h *Handler) HandleUpdateDraft(w http.ResponseWriter, r *http.Request) {
	orgID, _, _, err := h.getOrgAndRFQID(r)
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

	var input UpdateDraftInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input", "BAD_REQUEST")
		return
	}

	draft, err := h.svc.UpdateDraft(r.Context(), orgID, draftID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "UPDATE_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Draft updated successfully", draft)
}

// HandleSubmitApproval POST /api/v1/rfqs/{id}/pricing-workflow/drafts/{draftId}/submit-approval
func (h *Handler) HandleSubmitApproval(w http.ResponseWriter, r *http.Request) {
	orgID, userID, _, err := h.getOrgAndRFQID(r)
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

	var input SubmitApprovalInput
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&input)
	}

	draft, err := h.svc.SubmitDraftForApproval(r.Context(), orgID, draftID, userID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "SUBMIT_APPROVAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Draft submitted for approval successfully", draft)
}
