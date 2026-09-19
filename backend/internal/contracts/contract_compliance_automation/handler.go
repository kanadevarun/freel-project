package contract_compliance_automation

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

// Handler provides HTTP endpoints for contract and compliance automation
type Handler struct {
	svc Service
}

// NewHandler creates a new Handler instance
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts all contract compliance automation endpoints onto the given router
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/overview", h.HandleGetOverview)
	r.Post("/review", h.HandleReviewContract)
	r.Post("/extract-clauses", h.HandleExtractClauses)
	r.Post("/verify-terms", h.HandleVerifyStructuredTerms)
	r.Get("/compliance-checklist", h.HandleAssessCompliance)
	r.Get("/drafts", h.HandleListDrafts)
	r.Post("/drafts", h.HandleGenerateDraft)
	r.Get("/drafts/{draftId:[0-9]+}", h.HandleGetDraft)
	r.Put("/drafts/{draftId:[0-9]+}", h.HandleUpdateDraft)
	r.Post("/drafts/{draftId:[0-9]+}/submit-approval", h.HandleSubmitApproval)
}

func (h *Handler) getOrgAndContractID(r *http.Request) (int64, int64, int64, error) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		return 0, 0, 0, http.ErrNoCookie
	}
	orgID := userCtx.OrgID
	userID := userCtx.UserID

	contractIDStr := chi.URLParam(r, "id")
	contractID, err := strconv.ParseInt(contractIDStr, 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	return orgID, userID, contractID, nil
}

// HandleGetOverview GET /api/v1/contracts/{id}/compliance-automation/overview
func (h *Handler) HandleGetOverview(w http.ResponseWriter, r *http.Request) {
	orgID, _, contractID, err := h.getOrgAndContractID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	overview, err := h.svc.GetContractOverview(r.Context(), orgID, contractID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, "Contract not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    overview,
	})
}

// HandleReviewContract POST /api/v1/contracts/{id}/compliance-automation/review
func (h *Handler) HandleReviewContract(w http.ResponseWriter, r *http.Request) {
	orgID, _, contractID, err := h.getOrgAndContractID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	review, err := h.svc.ReviewContract(r.Context(), orgID, contractID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, "Contract not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "REVIEW_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    review,
	})
}

// HandleExtractClauses POST /api/v1/contracts/{id}/compliance-automation/extract-clauses
func (h *Handler) HandleExtractClauses(w http.ResponseWriter, r *http.Request) {
	orgID, _, contractID, err := h.getOrgAndContractID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	result, err := h.svc.ExtractClauses(r.Context(), orgID, contractID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXTRACTION_FAILED")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}

// HandleVerifyStructuredTerms POST /api/v1/contracts/{id}/compliance-automation/verify-terms
func (h *Handler) HandleVerifyStructuredTerms(w http.ResponseWriter, r *http.Request) {
	orgID, _, contractID, err := h.getOrgAndContractID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	result, err := h.svc.VerifyStructuredTerms(r.Context(), orgID, contractID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "VERIFICATION_FAILED")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}

// HandleAssessCompliance GET /api/v1/contracts/{id}/compliance-automation/compliance-checklist
func (h *Handler) HandleAssessCompliance(w http.ResponseWriter, r *http.Request) {
	orgID, _, contractID, err := h.getOrgAndContractID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	result, err := h.svc.AssessCompliance(r.Context(), orgID, contractID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "ASSESSMENT_FAILED")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}

// HandleListDrafts GET /api/v1/contracts/{id}/compliance-automation/drafts
func (h *Handler) HandleListDrafts(w http.ResponseWriter, r *http.Request) {
	orgID, _, contractID, err := h.getOrgAndContractID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	drafts, err := h.svc.ListDrafts(r.Context(), orgID, contractID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    drafts,
	})
}

// HandleGenerateDraft POST /api/v1/contracts/{id}/compliance-automation/drafts
func (h *Handler) HandleGenerateDraft(w http.ResponseWriter, r *http.Request) {
	orgID, userID, contractID, err := h.getOrgAndContractID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var input GenerateClarificationDraftInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	draft, err := h.svc.GenerateClarificationDraft(r.Context(), orgID, contractID, userID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DRAFT_GENERATION_FAILED")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    draft,
	})
}

// HandleGetDraft GET /api/v1/contracts/{id}/compliance-automation/drafts/{draftId}
func (h *Handler) HandleGetDraft(w http.ResponseWriter, r *http.Request) {
	orgID, _, _, err := h.getOrgAndContractID(r)
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

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    draft,
	})
}

// HandleUpdateDraft PUT /api/v1/contracts/{id}/compliance-automation/drafts/{draftId}
func (h *Handler) HandleUpdateDraft(w http.ResponseWriter, r *http.Request) {
	orgID, _, _, err := h.getOrgAndContractID(r)
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

	var input UpdateClarificationDraftInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	updated, err := h.svc.UpdateDraft(r.Context(), orgID, draftID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "UPDATE_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    updated,
	})
}

// HandleSubmitApproval POST /api/v1/contracts/{id}/compliance-automation/drafts/{draftId}/submit-approval
func (h *Handler) HandleSubmitApproval(w http.ResponseWriter, r *http.Request) {
	orgID, userID, _, err := h.getOrgAndContractID(r)
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

	var input SubmitDraftApprovalInput
	_ = json.NewDecoder(r.Body).Decode(&input)

	updated, err := h.svc.SubmitDraftForApproval(r.Context(), orgID, draftID, userID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "APPROVAL_SUBMIT_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    updated,
	})
}
