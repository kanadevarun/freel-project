package reports

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

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/advanced", h.HandleGetAdvancedReport)
	r.Post("/export", h.HandleExportReport)
	r.Post("/distribute", h.HandleDistributeReport)
	r.Get("/history", h.HandleListHistory)
	r.Get("/distributions", h.HandleListDistributions)
}

func (h *Handler) getUser(r *http.Request) (middleware.UserContext, bool) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		return middleware.UserContext{}, false
	}
	return userCtx, true
}

func (h *Handler) HandleGetAdvancedReport(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	reportType := r.URL.Query().Get("report_type")
	if reportType == "" {
		reportType = "OPERATIONAL_VOLUME"
	}
	dateRange := r.URL.Query().Get("date_range")
	if dateRange == "" {
		dateRange = "LAST_90D"
	}

	resp, err := h.svc.GetAdvancedReport(r.Context(), user.OrgID, user.UserID, reportType, dateRange)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleExportReport(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	var input ExportReportInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_ARGUMENT")
		return
	}

	exportReq, err := h.svc.ExportReport(r.Context(), user.OrgID, user.UserID, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, exportReq)
}

func (h *Handler) HandleDistributeReport(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	var input DistributeReportInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_ARGUMENT")
		return
	}

	if len(input.RecipientEmails) == 0 {
		utils.Error(w, http.StatusBadRequest, "Recipient emails required", "INVALID_ARGUMENT")
		return
	}

	userName := user.CognitoID
	if userName == "" {
		userName = "LogisticsHQ Operator"
	}

	distReq, err := h.svc.RequestDistribution(r.Context(), user.OrgID, user.UserID, userName, input)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, distReq)
}

func (h *Handler) HandleListHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	snapshots, err := h.svc.ListSnapshots(r.Context(), user.OrgID, limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	if snapshots == nil {
		snapshots = []AnalyticsReportSnapshot{}
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"snapshots": snapshots,
		"total":     len(snapshots),
	})
}

func (h *Handler) HandleListDistributions(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	dists, err := h.svc.ListDistributions(r.Context(), user.OrgID, limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}
	if dists == nil {
		dists = []ReportDistributionRequest{}
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"distributions": dists,
		"total":         len(dists),
	})
}
