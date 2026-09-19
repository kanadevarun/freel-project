package operations_automation

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for Shipment Operations Automation
type Handler struct {
	svc Service
}

// NewHandler creates a new Handler instance
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts all shipment operations automation routes onto the provided chi Router
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/overview", h.HandleGetOverview)
	r.Post("/analyze-risks", h.HandleAnalyzeRisks)
	r.Post("/prioritize-exceptions", h.HandlePrioritizeExceptions)
	r.Get("/recommendations", h.HandleGetRecommendations)
	r.Get("/drafts", h.HandleListDrafts)
	r.Post("/drafts", h.HandleGenerateDraft)
	r.Get("/drafts/{draftId:[0-9]+}", h.HandleGetDraft)
	r.Put("/drafts/{draftId:[0-9]+}", h.HandleUpdateDraft)
	r.Post("/drafts/{draftId:[0-9]+}/submit-approval", h.HandleSubmitApproval)
}

func (h *Handler) getOrgAndShipmentID(r *http.Request) (int64, int64, int64, error) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		return 0, 0, 0, http.ErrNoCookie
	}
	orgID := userCtx.OrgID
	userID := userCtx.UserID

	shipmentIDStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(shipmentIDStr, 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	return orgID, userID, shipmentID, nil
}

// HandleGetOverview GET /api/v1/shipments/{id}/operations-automation/overview
func (h *Handler) HandleGetOverview(w http.ResponseWriter, r *http.Request) {
	orgID, _, shipmentID, err := h.getOrgAndShipmentID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	overview, err := h.svc.GetShipmentOperationsOverview(r.Context(), orgID, shipmentID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "Shipment not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "OPERATIONS_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Shipment operations overview retrieved", overview)
}

// HandleAnalyzeRisks POST /api/v1/shipments/{id}/operations-automation/analyze-risks
func (h *Handler) HandleAnalyzeRisks(w http.ResponseWriter, r *http.Request) {
	orgID, _, shipmentID, err := h.getOrgAndShipmentID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	analysis, err := h.svc.AnalyzeShipmentRisks(r.Context(), orgID, shipmentID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "Shipment not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "RISK_ANALYSIS_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Shipment operational risks analyzed", analysis)
}

// HandlePrioritizeExceptions POST /api/v1/shipments/{id}/operations-automation/prioritize-exceptions
func (h *Handler) HandlePrioritizeExceptions(w http.ResponseWriter, r *http.Request) {
	orgID, _, shipmentID, err := h.getOrgAndShipmentID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	result, err := h.svc.PrioritizeExceptions(r.Context(), orgID, shipmentID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "Shipment not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EXCEPTION_PRIORITIZATION_ERROR")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}

// HandleGetRecommendations GET /api/v1/shipments/{id}/operations-automation/recommendations
func (h *Handler) HandleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	orgID, _, shipmentID, err := h.getOrgAndShipmentID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	result, err := h.svc.GetOperationalRecommendations(r.Context(), orgID, shipmentID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "Shipment not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "RECOMMENDATIONS_ERROR")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result)
}

// HandleListDrafts GET /api/v1/shipments/{id}/operations-automation/drafts
func (h *Handler) HandleListDrafts(w http.ResponseWriter, r *http.Request) {
	orgID, _, shipmentID, err := h.getOrgAndShipmentID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	drafts, err := h.svc.ListCommunicationDrafts(r.Context(), orgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "FETCH_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Shipment communication drafts retrieved", drafts)
}

// HandleGenerateDraft POST /api/v1/shipments/{id}/operations-automation/drafts
func (h *Handler) HandleGenerateDraft(w http.ResponseWriter, r *http.Request) {
	orgID, userID, shipmentID, err := h.getOrgAndShipmentID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var input GenerateDraftInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON input", "BAD_REQUEST")
		return
	}

	draft, err := h.svc.GenerateCommunicationDraft(r.Context(), orgID, shipmentID, userID, input)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, "Shipment not found or access denied", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DRAFT_ERROR")
		return
	}

	utils.Success(w, http.StatusCreated, "Shipment communication draft generated", draft)
}

// HandleGetDraft GET /api/v1/shipments/{id}/operations-automation/drafts/{draftId}
func (h *Handler) HandleGetDraft(w http.ResponseWriter, r *http.Request) {
	orgID, _, _, err := h.getOrgAndShipmentID(r)
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

	utils.Success(w, http.StatusOK, "Shipment communication draft retrieved", draft)
}

// HandleUpdateDraft PUT /api/v1/shipments/{id}/operations-automation/drafts/{draftId}
func (h *Handler) HandleUpdateDraft(w http.ResponseWriter, r *http.Request) {
	orgID, _, _, err := h.getOrgAndShipmentID(r)
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

	utils.Success(w, http.StatusOK, "Shipment communication draft updated", draft)
}

// HandleSubmitApproval POST /api/v1/shipments/{id}/operations-automation/drafts/{draftId}/submit-approval
func (h *Handler) HandleSubmitApproval(w http.ResponseWriter, r *http.Request) {
	orgID, userID, _, err := h.getOrgAndShipmentID(r)
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
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&input)
	}

	draft, err := h.svc.SubmitDraftForApproval(r.Context(), orgID, draftID, userID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "SUBMIT_APPROVAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Shipment draft submitted for managerial approval", draft)
}
