package recommendations

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for the Recommendation Center
type Handler struct {
	svc Service
}

// NewHandler creates a new recommendations HTTP handler
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// List handles GET /api/recommendations
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := RecommendationFilter{
		Page:       page,
		Limit:      limit,
		Category:   q.Get("category"),
		Priority:   q.Get("priority"),
		RiskLevel:  q.Get("risk_level"),
		Status:     q.Get("status"),
		SourceType: q.Get("source_type"),
		Search:     q.Get("search"),
		SortBy:     q.Get("sort_by"),
		SortDir:    q.Get("sort_dir"),
	}

	if reqApp := q.Get("requires_approval"); reqApp != "" {
		b := reqApp == "true" || reqApp == "1"
		filter.RequiresApproval = &b
	}

	if aIDStr := q.Get("assignee_id"); aIDStr != "" {
		if aID, err := strconv.ParseInt(aIDStr, 10, 64); err == nil {
			filter.AssigneeID = &aID
		}
	}

	if shipmentIDStr := q.Get("shipment_id"); shipmentIDStr != "" {
		if shipmentID, err := strconv.ParseInt(shipmentIDStr, 10, 64); err == nil && shipmentID > 0 {
			filter.ShipmentID = &shipmentID
		}
	}
	if milestoneIDStr := q.Get("milestone_id"); milestoneIDStr != "" {
		if milestoneID, err := strconv.ParseInt(milestoneIDStr, 10, 64); err == nil && milestoneID > 0 {
			filter.MilestoneID = &milestoneID
		}
	}
	if exceptionIDStr := q.Get("exception_id"); exceptionIDStr != "" {
		if exceptionID, err := strconv.ParseInt(exceptionIDStr, 10, 64); err == nil && exceptionID > 0 {
			filter.ExceptionID = &exceptionID
		}
	}
	if bookingIDStr := q.Get("booking_id"); bookingIDStr != "" {
		if bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64); err == nil && bookingID > 0 {
			filter.BookingID = &bookingID
		}
	}
	if q.Get("delayed_milestone") == "true" {
		t := true
		filter.DelayedMilestone = &t
	}
	if q.Get("active_exception") == "true" {
		t := true
		filter.ActiveException = &t
	}
	if q.Get("missing_ops_info") == "true" {
		t := true
		filter.MissingOpsInfo = &t
	}

	if rfqIDStr := q.Get("rfq_id"); rfqIDStr != "" {
		if rfqID, err := strconv.ParseInt(rfqIDStr, 10, 64); err == nil && rfqID > 0 {
			filter.RFQID = &rfqID
		}
	}

	if quoteIDStr := q.Get("quotation_id"); quoteIDStr != "" {
		if quoteID, err := strconv.ParseInt(quoteIDStr, 10, 64); err == nil && quoteID > 0 {
			filter.QuotationID = &quoteID
		}
	}

	if q.Get("missing_info") == "true" {
		t := true
		filter.MissingInfo = &t
	}
	if q.Get("expiring_soon") == "true" {
		t := true
		filter.ExpiringSoon = &t
	}
	if q.Get("pricing_concern") == "true" {
		t := true
		filter.PricingConcern = &t
	}

	if invoiceIDStr := q.Get("invoice_id"); invoiceIDStr != "" {
		if invoiceID, err := strconv.ParseInt(invoiceIDStr, 10, 64); err == nil && invoiceID > 0 {
			filter.InvoiceID = &invoiceID
		}
	}
	if q.Get("overdue_only") == "true" {
		t := true
		filter.OverdueOnly = &t
	}
	if q.Get("data_quality_only") == "true" {
		t := true
		filter.DataQualityOnly = &t
	}
	if q.Get("upcoming_only") == "true" {
		t := true
		filter.UpcomingOnly = &t
	}
	if agingBand := q.Get("aging_band"); agingBand != "" {
		filter.AgingBand = agingBand
	}

	// Phase 2 Task 2.6 Filters
	if contractIDStr := q.Get("contract_id"); contractIDStr != "" {
		if contractID, err := strconv.ParseInt(contractIDStr, 10, 64); err == nil && contractID > 0 {
			filter.ContractID = &contractID
		}
	}
	if docIDStr := q.Get("document_id"); docIDStr != "" {
		filter.DocumentID = &docIDStr
	}
	if compIDStr := q.Get("compliance_id"); compIDStr != "" {
		if compID, err := strconv.ParseInt(compIDStr, 10, 64); err == nil && compID > 0 {
			filter.ComplianceID = &compID
		}
	}
	if carrierIDStr := q.Get("carrier_id"); carrierIDStr != "" {
		if carrierID, err := strconv.ParseInt(carrierIDStr, 10, 64); err == nil && carrierID > 0 {
			filter.CarrierID = &carrierID
		}
	}
	if q.Get("expired_only") == "true" {
		t := true
		filter.ExpiredOnly = &t
	}

	// Phase 2 Task 2.7: Workflow Automation Filters
	if autoIDStr := q.Get("automation_id"); autoIDStr != "" {
		if autoID, err := strconv.ParseInt(autoIDStr, 10, 64); err == nil && autoID > 0 {
			filter.AutomationID = &autoID
		}
	}
	if execIDStr := q.Get("execution_id"); execIDStr != "" {
		if execID, err := strconv.ParseInt(execIDStr, 10, 64); err == nil && execID > 0 {
			filter.ExecutionID = &execID
		}
	}

	resp, err := h.svc.ListRecommendations(r.Context(), userCtx.OrgID, filter)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "RECOMMENDATIONS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"recommendations": resp.Recommendations,
		"pagination":      resp.Pagination,
		"stats":           resp.Stats,
	})
}

// GetStats handles GET /api/recommendations/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	stats, err := h.svc.GetStats(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "STATS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

// GetByID handles GET /api/recommendations/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid recommendation ID", "INVALID_ID")
		return
	}

	rec, err := h.svc.GetRecommendation(r.Context(), userCtx.OrgID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(w, http.StatusNotFound, "Recommendation not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "RECOMMENDATION_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"recommendation": rec,
	})
}

// Generate handles POST /api/recommendations/generate
func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")

	res, err := h.svc.GenerateRecommendations(r.Context(), userCtx.OrgID, correlationID, userCtx.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "GENERATE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"result":  res,
	})
}

// UpdateStatus handles PATCH /api/recommendations/{id}/status
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid recommendation ID", "INVALID_ID")
		return
	}

	var input UpdateStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")
	userLabel := userCtx.Role
	if userLabel == "" {
		userLabel = strconv.FormatInt(userCtx.UserID, 10)
	}

	updated, err := h.svc.UpdateStatus(r.Context(), userCtx.OrgID, id, input, userCtx.UserID, userLabel, correlationID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(w, http.StatusNotFound, "Recommendation not found", "NOT_FOUND")
			return
		}
		if errors.Is(err, ErrInvalidTransition) || errors.Is(err, ErrReasonRequired) || errors.Is(err, ErrApprovalRequired) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "VALIDATION_FAILED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "UPDATE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"recommendation": updated,
	})
}

// Assign handles POST /api/recommendations/{id}/assign
func (h *Handler) Assign(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid recommendation ID", "INVALID_ID")
		return
	}

	var input AssignInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")
	userLabel := userCtx.Role
	if userLabel == "" {
		userLabel = strconv.FormatInt(userCtx.UserID, 10)
	}

	updated, err := h.svc.AssignRecommendation(r.Context(), userCtx.OrgID, id, input, userCtx.UserID, userLabel, correlationID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(w, http.StatusNotFound, "Recommendation not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "ASSIGN_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"recommendation": updated,
	})
}

// Dismiss handles POST /api/recommendations/{id}/dismiss
func (h *Handler) Dismiss(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid recommendation ID", "INVALID_ID")
		return
	}

	var input DismissInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")
	userLabel := userCtx.Role
	if userLabel == "" {
		userLabel = strconv.FormatInt(userCtx.UserID, 10)
	}

	updated, err := h.svc.DismissRecommendation(r.Context(), userCtx.OrgID, id, input, userCtx.UserID, userLabel, correlationID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(w, http.StatusNotFound, "Recommendation not found", "NOT_FOUND")
			return
		}
		if errors.Is(err, ErrReasonRequired) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "REASON_REQUIRED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DISMISS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"recommendation": updated,
	})
}

// MarkReviewed handles POST /api/recommendations/{id}/review
func (h *Handler) MarkReviewed(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid recommendation ID", "INVALID_ID")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")
	userLabel := userCtx.Role
	if userLabel == "" {
		userLabel = strconv.FormatInt(userCtx.UserID, 10)
	}

	updated, err := h.svc.MarkReviewed(r.Context(), userCtx.OrgID, id, userCtx.UserID, userLabel, correlationID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(w, http.StatusNotFound, "Recommendation not found", "NOT_FOUND")
			return
		}
		if errors.Is(err, ErrInvalidTransition) {
			utils.Error(w, http.StatusBadRequest, err.Error(), "INVALID_TRANSITION")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "REVIEW_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"recommendation": updated,
	})
}

// GetEvidence handles GET /api/recommendations/{id}/evidence
func (h *Handler) GetEvidence(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid recommendation ID", "INVALID_ID")
		return
	}

	rec, err := h.svc.GetRecommendation(r.Context(), userCtx.OrgID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(w, http.StatusNotFound, "Recommendation not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "RECOMMENDATION_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"recommendation_id": rec.ID,
		"source_type":     rec.SourceType,
		"source_id":       rec.SourceID,
		"source_ref":      rec.SourceReference,
		"evidence":        rec.Evidence,
		"rule_applied":    rec.RuleApplied,
		"correlation_id":  rec.CorrelationID,
	})
}

// ListBySource handles GET /api/recommendations/source/{sourceType}/{sourceId}
func (h *Handler) ListBySource(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	sourceType := strings.ToUpper(chi.URLParam(r, "sourceType"))
	sourceIDStr := chi.URLParam(r, "sourceId")
	sourceID, err := strconv.ParseInt(sourceIDStr, 10, 64)
	if err != nil || sourceID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid source ID", "INVALID_ID")
		return
	}

	items, err := h.svc.ListBySource(r.Context(), userCtx.OrgID, sourceType, sourceID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "SOURCE_RECOMMENDATIONS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"source_type":     sourceType,
		"source_id":       sourceID,
		"recommendations": items,
	})
}

// GenerateDraft handles POST /api/recommendations/{id}/draft
// Explicit user action to generate a draft message
func (h *Handler) GenerateDraft(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid recommendation ID", "INVALID_ID")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")
	userLabel := userCtx.Role
	if userLabel == "" {
		userLabel = strconv.FormatInt(userCtx.UserID, 10)
	}

	rec, err := h.svc.GenerateDraft(r.Context(), userCtx.OrgID, id, userCtx.UserID, userLabel, correlationID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(w, http.StatusNotFound, "Recommendation not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DRAFT_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"recommendation": rec,
	})
}

// SaveDraft handles PATCH /api/recommendations/{id}/draft
// Allows user to edit and save the draft message
func (h *Handler) SaveDraft(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid recommendation ID", "INVALID_ID")
		return
	}

	var input SaveDraftInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")
	userLabel := userCtx.Role
	if userLabel == "" {
		userLabel = strconv.FormatInt(userCtx.UserID, 10)
	}

	rec, err := h.svc.SaveDraft(r.Context(), userCtx.OrgID, id, input, userCtx.UserID, userLabel, correlationID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(w, http.StatusNotFound, "Recommendation not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "SAVE_DRAFT_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"recommendation": rec,
	})
}

// CreateFollowupTask handles POST /api/recommendations/{id}/task
// Idempotent creation of a controlled internal follow-up task
func (h *Handler) CreateFollowupTask(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid recommendation ID", "INVALID_ID")
		return
	}

	var input CreateFollowupTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")
	userLabel := userCtx.Role
	if userLabel == "" {
		userLabel = strconv.FormatInt(userCtx.UserID, 10)
	}

	task, err := h.svc.CreateFollowupTask(r.Context(), userCtx.OrgID, id, input, userCtx.UserID, userLabel, correlationID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(w, http.StatusNotFound, "Recommendation not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "CREATE_TASK_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"task":    task,
	})
}

// ListFollowupTasks handles GET /api/recommendations/tasks
func (h *Handler) ListFollowupTasks(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	var customerIDPtr *int64
	if cIDStr := q.Get("customer_id"); cIDStr != "" {
		if cID, err := strconv.ParseInt(cIDStr, 10, 64); err == nil && cID > 0 {
			customerIDPtr = &cID
		}
	}

	tasks, err := h.svc.ListFollowupTasks(r.Context(), userCtx.OrgID, customerIDPtr, q.Get("status"))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "TASKS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"tasks":   tasks,
		"total":   len(tasks),
	})
}

// GetFollowupStats handles GET /api/recommendations/followups/stats
func (h *Handler) GetFollowupStats(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	stats, err := h.svc.GetFollowupStats(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "STATS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

// GetActionPreview handles GET /api/recommendations/{id}/action-preview
func (h *Handler) GetActionPreview(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid recommendation ID", "INVALID_ID")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")
	userLabel := userCtx.Role
	if userLabel == "" {
		userLabel = strconv.FormatInt(userCtx.UserID, 10)
	}

	preview, err := h.svc.GetActionPreview(r.Context(), userCtx.OrgID, id, userCtx.UserID, userLabel, correlationID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(w, http.StatusNotFound, "Recommendation not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREVIEW_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"action_preview": preview,
		"preview":        preview,
	})
}

// RequestApproval handles POST /api/recommendations/{id}/request-approval
func (h *Handler) RequestApproval(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid recommendation ID", "INVALID_ID")
		return
	}

	var input RequestApprovalInput
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&input)
	}

	correlationID := r.Header.Get("X-Correlation-Id")
	userLabel := userCtx.Role
	if userLabel == "" {
		userLabel = strconv.FormatInt(userCtx.UserID, 10)
	}

	rec, err := h.svc.RequestApproval(r.Context(), userCtx.OrgID, id, input, userCtx.UserID, userLabel, correlationID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			utils.Error(w, http.StatusNotFound, "Recommendation not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "APPROVAL_REQUEST_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"recommendation": rec,
	})
}

// GetShipmentEvidence handles GET /api/recommendations/shipments/{shipmentId}/evidence
func (h *Handler) GetShipmentEvidence(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "shipmentId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment ID", "INVALID_ID")
		return
	}

	items, err := h.svc.GetEntityEvidence(r.Context(), userCtx.OrgID, "SHIPMENT", id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EVIDENCE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"entity_type": "SHIPMENT",
		"entity_id":   id,
		"evidence":    items,
	})
}

// GetMilestoneEvidence handles GET /api/recommendations/milestones/{milestoneId}/evidence
func (h *Handler) GetMilestoneEvidence(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "milestoneId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid milestone ID", "INVALID_ID")
		return
	}

	items, err := h.svc.GetEntityEvidence(r.Context(), userCtx.OrgID, "MILESTONE", id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EVIDENCE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"entity_type": "MILESTONE",
		"entity_id":   id,
		"evidence":    items,
	})
}

// GetExceptionEvidence handles GET /api/recommendations/exceptions/{exceptionId}/evidence
func (h *Handler) GetExceptionEvidence(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "exceptionId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid exception ID", "INVALID_ID")
		return
	}

	items, err := h.svc.GetEntityEvidence(r.Context(), userCtx.OrgID, "EXCEPTION", id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EVIDENCE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"entity_type": "EXCEPTION",
		"entity_id":   id,
		"evidence":    items,
	})
}

// GetInvoiceEvidence handles GET /api/recommendations/invoices/{invoiceId}/evidence
func (h *Handler) GetInvoiceEvidence(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "invoiceId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice ID", "INVALID_ID")
		return
	}

	payload, err := h.svc.GetInvoiceEvidence(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EVIDENCE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"evidence": payload,
	})
}

// GetCustomerCollectionSummary handles GET /api/recommendations/customers/{customerId}/collection-summary
func (h *Handler) GetCustomerCollectionSummary(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "customerId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid customer ID", "INVALID_ID")
		return
	}

	summary, err := h.svc.GetCustomerCollectionSummary(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "SUMMARY_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"summary": summary,
	})
}

// GetContractEvidence handles GET /api/recommendations/contracts/{contractId}/evidence
func (h *Handler) GetContractEvidence(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "contractId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid contract ID", "INVALID_ID")
		return
	}

	payload, err := h.svc.GetContractEvidence(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EVIDENCE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"evidence": payload,
	})
}

// GetDocumentEvidence handles GET /api/recommendations/documents/{documentId}/evidence
func (h *Handler) GetDocumentEvidence(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	docID := chi.URLParam(r, "documentId")
	if strings.TrimSpace(docID) == "" {
		utils.Error(w, http.StatusBadRequest, "Invalid document ID", "INVALID_ID")
		return
	}

	payload, err := h.svc.GetDocumentEvidence(r.Context(), userCtx.OrgID, docID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EVIDENCE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"evidence": payload,
	})
}

// GetComplianceEvidence handles GET /api/recommendations/compliance/{complianceId}/evidence
func (h *Handler) GetComplianceEvidence(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "complianceId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid compliance ID", "INVALID_ID")
		return
	}

	payload, err := h.svc.GetComplianceEvidence(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "EVIDENCE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"evidence": payload,
	})
}

// GetContractComplianceSummary handles GET /api/recommendations/contracts/compliance-summary
func (h *Handler) GetContractComplianceSummary(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	summary, err := h.svc.GetContractComplianceSummary(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "SUMMARY_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"summary": summary,
	})
}




