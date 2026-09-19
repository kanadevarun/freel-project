package monitoring

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
)

// Handler exposes AI monitoring and observability HTTP endpoints.
type Handler struct {
	service Service
}

// NewHandler creates a new HTTP handler for AI monitoring.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes mounts monitoring routes under the provided Chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.GetHealthSummary)
	r.Get("/executions", h.ListExecutionTraces)
	r.Get("/executions/{id}", h.GetExecutionTrace)
	r.Get("/performance", h.GetPerformanceMetrics)
	r.Get("/cost", h.GetCostMetrics)
	r.Get("/quality", h.GetQualitySummary)
	r.Get("/queue", h.GetQueueWorkerMetrics)
	r.Get("/recommendations", h.GetRecommendationMetrics)
	r.Get("/approvals", h.GetApprovalMetrics)
	r.Get("/memory", h.GetMemorySafetyMetrics)
	r.Get("/security", h.GetSecurityMetrics)

	r.Get("/pricing", h.GetModelPricingCatalog)
	r.Put("/pricing", h.UpdateModelPricing)

	r.Get("/thresholds", h.GetHealthThresholds)
	r.Put("/thresholds", h.UpdateHealthThreshold)

	r.Post("/evaluate", h.EvaluateTraceQuality)
}

func (h *Handler) GetHealthSummary(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized: valid organization context required",
		})
		return
	}

	summary, err := h.service.GetHealthSummary(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "failed to compute AI health summary: " + err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    summary,
	})
}

func (h *Handler) ListExecutionTraces(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized: valid organization context required",
		})
		return
	}

	q := r.URL.Query()
	days, _ := strconv.Atoi(q.Get("days"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := ListExecutionsFilter{
		Feature:       q.Get("feature"),
		Assistant:     q.Get("assistant"),
		ModelProvider: q.Get("provider"),
		Status:        q.Get("status"),
		CorrelationID: q.Get("correlation_id"),
		Days:          days,
		Limit:         limit,
		Offset:        offset,
	}

	traces, total, err := h.service.ListExecutionTraces(r.Context(), userCtx.OrgID, filter)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "failed to list execution traces: " + err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"items":  traces,
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	})
}

func (h *Handler) GetExecutionTrace(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized: valid organization context required",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "invalid execution trace ID",
		})
		return
	}

	trace, err := h.service.GetExecutionTrace(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "failed to get execution trace: " + err.Error(),
		})
		return
	}
	if trace == nil {
		utils.JSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   "execution trace not found",
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    trace,
	})
}

func (h *Handler) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 7
	}

	perf, err := h.service.GetPerformanceMetrics(r.Context(), userCtx.OrgID, days)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "failed to get performance metrics: " + err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    perf,
	})
}

func (h *Handler) GetCostMetrics(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30
	}

	cost, err := h.service.GetCostMetrics(r.Context(), userCtx.OrgID, days)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "failed to get cost metrics: " + err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    cost,
	})
}

func (h *Handler) GetQualitySummary(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	quality, err := h.service.GetQualitySummary(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "failed to get quality summary: " + err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    quality,
	})
}

func (h *Handler) GetQueueWorkerMetrics(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	queue, err := h.service.GetQueueWorkerMetrics(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "failed to get queue metrics: " + err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    queue,
	})
}

func (h *Handler) GetRecommendationMetrics(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	recs, err := h.service.GetRecommendationMetrics(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "failed to get recommendation metrics: " + err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    recs,
	})
}

func (h *Handler) GetApprovalMetrics(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	apps, err := h.service.GetApprovalMetrics(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "failed to get approval metrics: " + err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    apps,
	})
}

func (h *Handler) GetMemorySafetyMetrics(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	mem, err := h.service.GetMemorySafetyMetrics(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "failed to get memory safety metrics: " + err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    mem,
	})
}

func (h *Handler) GetSecurityMetrics(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	sec, err := h.service.GetSecurityMetrics(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "failed to get security metrics: " + err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    sec,
	})
}

func (h *Handler) GetModelPricingCatalog(w http.ResponseWriter, r *http.Request) {
	catalog, err := h.service.GetModelPricingCatalog(r.Context())
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "failed to get pricing catalog: " + err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    catalog,
	})
}

func (h *Handler) UpdateModelPricing(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	var req ModelPricingRecord
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "invalid request body",
		})
		return
	}

	if err := h.service.UpdateModelPricing(r.Context(), userCtx.Role, &req); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "insufficient permissions") {
			status = http.StatusForbidden
		} else if strings.Contains(err.Error(), "required") {
			status = http.StatusBadRequest
		}
		utils.JSON(w, status, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Model pricing updated successfully",
	})
}

func (h *Handler) GetHealthThresholds(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	thresholds, err := h.service.GetHealthThresholds(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "failed to get health thresholds: " + err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    thresholds,
	})
}

func (h *Handler) UpdateHealthThreshold(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	var req HealthThresholdRecord
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "invalid request body",
		})
		return
	}

	if err := h.service.UpdateHealthThreshold(r.Context(), userCtx.OrgID, userCtx.Role, &req); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "insufficient permissions") {
			status = http.StatusForbidden
		} else if strings.Contains(err.Error(), "required") {
			status = http.StatusBadRequest
		}
		utils.JSON(w, status, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Health threshold updated successfully",
	})
}

func (h *Handler) EvaluateTraceQuality(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	var req QualityEvaluationRecord
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "invalid request body",
		})
		return
	}

	eval, err := h.service.EvaluateTraceQuality(r.Context(), userCtx.OrgID, &req)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "failed to record quality evaluation: " + err.Error(),
		})
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    eval,
	})
}
