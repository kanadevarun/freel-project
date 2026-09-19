package sportal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
)

// Handler processes incoming HTTP requests for the internal SPortal API.
type Handler struct {
	service Service
}

// NewHandler initializes a new SPortal HTTP handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	parts := strings.Split(ip, ",")
	return strings.TrimSpace(parts[0])
}

// GetHealth returns basic SPortal health status.
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	utils.Success(w, http.StatusOK, "SPortal API is healthy and operational", map[string]interface{}{
		"status": "UP",
		"system": "SPortal Foundation",
	})
}

// GetMeta provides frontend module discovery and environment configuration.
func (h *Handler) GetMeta(w http.ResponseWriter, r *http.Request) {
	meta, err := h.service.GetMeta(r.Context())
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to load SPortal metadata", "INTERNAL_ERROR")
		return
	}
	utils.Success(w, http.StatusOK, "SPortal metadata loaded successfully", meta)
}

// Login handles internal staff authentication for SPortal.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req SPortalLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid login payload", "INVALID_PAYLOAD")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	resp, err := h.service.Login(r.Context(), req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "customer organization") || strings.Contains(errMsg, "access denied") {
			utils.Error(w, http.StatusForbidden, errMsg, "ACCESS_DENIED")
			return
		}
		utils.Error(w, http.StatusUnauthorized, errMsg, "LOGIN_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Login successful", resp)
}

// GetMe retrieves current authenticated internal user profile and SPortal permissions.
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.UserID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authentication context", "UNAUTHORIZED")
		return
	}

	data, err := h.service.GetMe(r.Context(), userCtx.UserID)
	if err != nil {
		utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
		return
	}

	utils.Success(w, http.StatusOK, "User profile retrieved successfully", data)
}

// Logout terminates the current SPortal session and logs audit event.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if ok && userCtx.UserID > 0 {
		actorID := userCtx.UserID
		_, _ = audit.Record(r.Context(), domain.CreateAuditLogParams{
			OrgID:        userCtx.OrgID,
			ActorID:      &actorID,
			ActorType:    domain.ActorTypeUser,
			ActorRole:    userCtx.Role,
			Action:       domain.ActionLogout,
			Module:       domain.ModuleAuthentication,
			ResourceType: "SPORTAL_SESSION",
			ResourceID:   fmt.Sprintf("%d", userCtx.UserID),
			Description:  "Internal user logged out of SPortal",
			Result:       domain.ResultSuccess,
			IPAddress:    getClientIP(r),
			UserAgent:    r.UserAgent(),
		})
	}
	utils.Success(w, http.StatusOK, "Logged out successfully", nil)
}

// GetPermissions returns the full permissions array for the authenticated user's role.
func (h *Handler) GetPermissions(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.UserID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	perms := h.service.GetPermissionsForRole(userCtx.Role)
	utils.Success(w, http.StatusOK, "Permissions retrieved", map[string]interface{}{
		"role":        userCtx.Role,
		"permissions": perms,
	})
}

// GetOverview returns platform-level metrics for authorized internal staff.
func (h *Handler) GetOverview(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	overview, err := h.service.GetPlatformOverview(r.Context(), userCtx.Role)
	if err != nil {
		utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
		return
	}

	utils.Success(w, http.StatusOK, "Platform overview retrieved successfully", overview)
}

// ListRecentOrganizations lists recent customer organizations for SPortal dashboard.
func (h *Handler) ListRecentOrganizations(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	limit := 10
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	orgs, err := h.service.ListRecentOrganizations(r.Context(), userCtx.Role, limit)
	if err != nil {
		utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
		return
	}

	utils.Success(w, http.StatusOK, "Recent organizations retrieved successfully", orgs)
}

// GetSensitiveFinancialData serves protected platform settlement details guarded by billing:sensitive_view.
func (h *Handler) GetSensitiveFinancialData(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	data, err := h.service.GetSensitiveFinancialData(r.Context(), userCtx)
	if err != nil {
		utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
		return
	}

	utils.Success(w, http.StatusOK, "Sensitive financial data retrieved", data)
}

// ListOrganizations retrieves paginated, searchable, filtered organizations.
func (h *Handler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	params := OrganizationListParams{
		Search:    strings.TrimSpace(r.URL.Query().Get("search")),
		Status:    strings.TrimSpace(r.URL.Query().Get("status")),
		Plan:      strings.TrimSpace(r.URL.Query().Get("plan")),
		Page:      page,
		Limit:     pageSize,
		SortBy:    strings.TrimSpace(r.URL.Query().Get("sort_by")),
		SortOrder: strings.TrimSpace(r.URL.Query().Get("sort_dir")),
	}

	result, err := h.service.ListOrganizations(r.Context(), userCtx.Role, params)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to list organizations: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Organizations retrieved successfully", result)
}

// GetOrganizationDetails retrieves complete Customer 360 foundation details for an organization.
func (h *Handler) GetOrganizationDetails(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	details, err := h.service.GetOrganizationDetails(r.Context(), userCtx.Role, orgID)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to fetch organization details: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Organization details retrieved successfully", details)
}

// CreateOrganization creates a new freight forwarder customer organization.
func (h *Handler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	profile, err := h.service.CreateOrganization(r.Context(), userCtx, req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "duplicate organization detected") {
			utils.Error(w, http.StatusConflict, errMsg, "DUPLICATE_ORGANIZATION")
			return
		}
		if strings.Contains(errMsg, "required") || strings.Contains(errMsg, "invalid") {
			utils.Error(w, http.StatusBadRequest, errMsg, "VALIDATION_FAILED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to create organization: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusCreated, "Organization created successfully", profile)
}

// UpdateOrganization modifies an existing freight forwarder customer organization profile.
func (h *Handler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	profile, err := h.service.UpdateOrganization(r.Context(), userCtx, orgID, req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "cannot update organization") || strings.Contains(errMsg, "duplicate") {
			utils.Error(w, http.StatusConflict, errMsg, "DUPLICATE_ORGANIZATION")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		if strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "empty") {
			utils.Error(w, http.StatusBadRequest, errMsg, "VALIDATION_FAILED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to update organization: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Organization updated successfully", profile)
}

// UploadOrganizationLogo handles uploading an organization brand logo via multipart/form-data.
func (h *Handler) UploadOrganizationLogo(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	// Max 5MB file upload limit
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		utils.Error(w, http.StatusBadRequest, "Failed to parse multipart form or file exceeds 5MB limit", "INVALID_FILE")
		return
	}

	file, header, err := r.FormFile("logo")
	if err != nil {
		file, header, err = r.FormFile("file")
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "No logo file found in request under 'logo' or 'file'", "MISSING_FILE")
			return
		}
	}
	defer file.Close()

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	details, logoURL, err := h.service.UploadOrganizationLogo(r.Context(), userCtx, orgID, header.Filename, file, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "invalid") {
			utils.Error(w, http.StatusBadRequest, errMsg, "INVALID_FILE_TYPE")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to upload logo: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Brand logo uploaded successfully", map[string]interface{}{
		"logo_url":     logoURL,
		"organization": details.Organization,
	})
}

// --- Task S5: Subscription & Plan Management Handlers ---

// ListPlans returns all available commercial plans and their customer adoption metrics.
func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	plans, err := h.service.ListPlans(r.Context(), userCtx)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to list subscription plans", "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Subscription plans retrieved successfully", plans)
}

// GetPlanByID retrieves a single commercial plan by its identifier.
func (h *Handler) GetPlanByID(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	planID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || planID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid plan ID", "INVALID_ID")
		return
	}

	plan, err := h.service.GetPlanByID(r.Context(), userCtx, planID)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, "Subscription plan not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve subscription plan", "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Subscription plan retrieved successfully", plan)
}

// CreatePlan defines a new subscription tier.
func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	var req CreatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	plan, err := h.service.CreatePlan(r.Context(), userCtx, req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "cannot be empty") || strings.Contains(errMsg, "non-negative") {
			utils.Error(w, http.StatusBadRequest, errMsg, "VALIDATION_FAILED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to create subscription plan: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusCreated, "Subscription plan created successfully", plan)
}

// UpdatePlan modifies an existing commercial subscription tier.
func (h *Handler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	planID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || planID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid plan ID", "INVALID_ID")
		return
	}

	var req UpdatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	plan, err := h.service.UpdatePlan(r.Context(), userCtx, planID, req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Subscription plan not found", "NOT_FOUND")
			return
		}
		if strings.Contains(errMsg, "cannot be empty") || strings.Contains(errMsg, "negative") {
			utils.Error(w, http.StatusBadRequest, errMsg, "VALIDATION_FAILED")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to update subscription plan: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Subscription plan updated successfully", plan)
}

// ListSubscriptions returns paginated customer subscription records and commercial metrics.
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	params := CustomerSubscriptionListParams{
		Search:    r.URL.Query().Get("search"),
		Status:    r.URL.Query().Get("status"),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}

	if planStr := r.URL.Query().Get("plan_id"); planStr != "" {
		if pid, err := strconv.ParseInt(planStr, 10, 64); err == nil && pid > 0 {
			params.PlanID = pid
		}
	}

	if autoRenewStr := r.URL.Query().Get("auto_renew"); autoRenewStr != "" {
		val := strings.ToLower(autoRenewStr) == "true" || autoRenewStr == "1"
		params.AutoRenew = &val
	}

	page := 1
	if pStr := r.URL.Query().Get("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	}
	params.Page = page

	limit := 15
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	params.Limit = limit

	result, err := h.service.ListCustomerSubscriptions(r.Context(), userCtx, params)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to list customer subscriptions: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer subscriptions retrieved successfully", result)
}

// GetOrganizationSubscription returns the full subscription detail view for an organization.
func (h *Handler) GetOrganizationSubscription(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	detail, err := h.service.GetCustomerSubscription(r.Context(), userCtx, orgID)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve subscription details: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Subscription details retrieved successfully", detail)
}

// AssignOrganizationSubscription sets up initial commercial subscription parameters for a tenant.
func (h *Handler) AssignOrganizationSubscription(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	var req AssignSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	err = h.service.AssignSubscription(r.Context(), userCtx, orgID, req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to assign subscription: "+errMsg, "INTERNAL_ERROR")
		return
	}

	detail, _ := h.service.GetCustomerSubscription(r.Context(), userCtx, orgID)
	utils.Success(w, http.StatusCreated, "Subscription assigned successfully", detail)
}

// ChangeOrganizationPlan changes an organization's plan tier or billing frequency.
func (h *Handler) ChangeOrganizationPlan(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	var req ChangeCustomerPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	err = h.service.ChangeCustomerPlan(r.Context(), userCtx, orgID, req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to change plan: "+errMsg, "INTERNAL_ERROR")
		return
	}

	detail, _ := h.service.GetCustomerSubscription(r.Context(), userCtx, orgID)
	utils.Success(w, http.StatusOK, "Subscription plan changed successfully", detail)
}

// ToggleOrganizationAutoRenew toggles the auto-renew status for a customer subscription.
func (h *Handler) ToggleOrganizationAutoRenew(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	var req ToggleAutoRenewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	err = h.service.ToggleAutoRenew(r.Context(), userCtx, orgID, req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to toggle auto-renew: "+errMsg, "INTERNAL_ERROR")
		return
	}

	detail, _ := h.service.GetCustomerSubscription(r.Context(), userCtx, orgID)
	utils.Success(w, http.StatusOK, "Subscription auto-renew setting updated successfully", detail)
}

// RenewOrganizationSubscription extends an active subscription period.
func (h *Handler) RenewOrganizationSubscription(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	var req RenewSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body for default 1 month renewal
		req = RenewSubscriptionRequest{ExtendMonths: 1}
	}
	if req.ExtendMonths <= 0 {
		req.ExtendMonths = 1
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	err = h.service.RenewSubscription(r.Context(), userCtx, orgID, req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to renew subscription: "+errMsg, "INTERNAL_ERROR")
		return
	}

	detail, _ := h.service.GetCustomerSubscription(r.Context(), userCtx, orgID)
	utils.Success(w, http.StatusOK, "Subscription renewed successfully", detail)
}

// CancelOrganizationSubscription terminates or schedules cancellation of a subscription.
func (h *Handler) CancelOrganizationSubscription(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	var req CancelCustomerSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	err = h.service.CancelCustomerSubscription(r.Context(), userCtx, orgID, req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to cancel subscription: "+errMsg, "INTERNAL_ERROR")
		return
	}

	detail, _ := h.service.GetCustomerSubscription(r.Context(), userCtx, orgID)
	utils.Success(w, http.StatusOK, "Subscription canceled successfully", detail)
}

// -----------------------------------------------------------------------------
// TASK S6: Customer Organization Users, Roles, Invitations & Access Lifecycle
// -----------------------------------------------------------------------------

// ListCustomerUsers handles GET /api/v1/sportal/users and GET /api/v1/sportal/organizations/{id}/users
func (h *Handler) ListCustomerUsers(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	var orgID int64
	if orgIDStr := chi.URLParam(r, "id"); orgIDStr != "" {
		orgID, _ = strconv.ParseInt(orgIDStr, 10, 64)
	} else if orgIDStr := q.Get("org_id"); orgIDStr != "" {
		orgID, _ = strconv.ParseInt(orgIDStr, 10, 64)
	}

	params := CustomerUserListParams{
		Search:           q.Get("search"),
		OrgID:            orgID,
		RoleName:         q.Get("role"),
		Status:           q.Get("status"),
		InvitationStatus: q.Get("invitation_status"),
		Page:             page,
		Limit:            limit,
		SortBy:           q.Get("sort_by"),
		SortOrder:        q.Get("sort_order"),
	}

	result, err := h.service.ListCustomerUsers(r.Context(), userCtx, params)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to list customer users: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer users retrieved successfully", result)
}

// GetCustomerUserDetail handles GET /api/v1/sportal/organizations/{id}/users/{userId}
func (h *Handler) GetCustomerUserDetail(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	userID, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil || userID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID", "INVALID_ID")
		return
	}

	detail, err := h.service.GetCustomerUserDetail(r.Context(), userCtx, orgID, userID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to get customer user detail: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer user details retrieved successfully", detail)
}

// GetOrgUserSummary handles GET /api/v1/sportal/organizations/{id}/users/summary
func (h *Handler) GetOrgUserSummary(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	summary, err := h.service.GetOrgUserSummary(r.Context(), userCtx, orgID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to get organization user summary: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Organization user summary retrieved successfully", summary)
}

// ListCustomerRoles handles GET /api/v1/sportal/users/roles
func (h *Handler) ListCustomerRoles(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	var orgID int64
	if orgIDStr := r.URL.Query().Get("org_id"); orgIDStr != "" {
		orgID, _ = strconv.ParseInt(orgIDStr, 10, 64)
	}

	roles, err := h.service.ListCustomerRoles(r.Context(), userCtx, orgID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to list customer roles: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer roles retrieved successfully", roles)
}

// InviteCustomerUser handles POST /api/v1/sportal/organizations/{id}/users/invite
func (h *Handler) InviteCustomerUser(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	var req InviteCustomerUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	invite, err := h.service.InviteCustomerUser(r.Context(), userCtx, orgID, req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "already an active member") || strings.Contains(errMsg, "required") {
			utils.Error(w, http.StatusBadRequest, errMsg, "BAD_REQUEST")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to invite customer user: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusCreated, "Customer user invitation created successfully", invite)
}

// ResendCustomerInvitation handles POST /api/v1/sportal/users/invitations/{invitationId}/resend
func (h *Handler) ResendCustomerInvitation(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	invitationID, err := strconv.ParseInt(chi.URLParam(r, "invitationId"), 10, 64)
	if err != nil || invitationID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid invitation ID", "INVALID_ID")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	invite, err := h.service.ResendCustomerInvitation(r.Context(), userCtx, invitationID, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to resend invitation: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Invitation resent successfully", invite)
}

// RevokeCustomerInvitation handles DELETE /api/v1/sportal/users/invitations/{invitationId}
func (h *Handler) RevokeCustomerInvitation(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	invitationID, err := strconv.ParseInt(chi.URLParam(r, "invitationId"), 10, 64)
	if err != nil || invitationID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid invitation ID", "INVALID_ID")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	err = h.service.RevokeCustomerInvitation(r.Context(), userCtx, invitationID, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to revoke invitation: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Invitation revoked successfully", nil)
}

// UpdateCustomerUserStatus handles PATCH /api/v1/sportal/organizations/{id}/users/{userId}/status
func (h *Handler) UpdateCustomerUserStatus(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	userID, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil || userID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID", "INVALID_ID")
		return
	}

	var req UpdateCustomerUserStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	err = h.service.UpdateCustomerUserStatus(r.Context(), userCtx, orgID, userID, req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		if strings.Contains(errMsg, "sole active Super Admin") || strings.Contains(errMsg, "invalid") {
			utils.Error(w, http.StatusBadRequest, errMsg, "BAD_REQUEST")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to update user status: "+errMsg, "INTERNAL_ERROR")
		return
	}

	detail, _ := h.service.GetCustomerUserDetail(r.Context(), userCtx, orgID, userID)
	utils.Success(w, http.StatusOK, "Customer user status updated successfully", detail)
}

// -----------------------------------------------------------------------------
// TASK S7: Customer Roles, Permissions Matrix & Access Administration
// -----------------------------------------------------------------------------

// GetPermissionMatrix handles GET /api/v1/sportal/roles/matrix
func (h *Handler) GetPermissionMatrix(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	var orgID int64
	if orgIDStr := r.URL.Query().Get("org_id"); orgIDStr != "" {
		orgID, _ = strconv.ParseInt(orgIDStr, 10, 64)
	}

	matrix, err := h.service.GetPermissionMatrix(r.Context(), userCtx, orgID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to get permission matrix: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Permission matrix retrieved successfully", matrix)
}

// UpdateCustomerUserRole handles PATCH /api/v1/sportal/organizations/{id}/users/{userId}/role
func (h *Handler) UpdateCustomerUserRole(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	userID, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil || userID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid user ID", "INVALID_ID")
		return
	}

	var req UpdateCustomerUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	updatedUser, err := h.service.UpdateCustomerUserRole(r.Context(), userCtx, orgID, userID, req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		if strings.Contains(errMsg, "must have at least one active Super Admin") || strings.Contains(errMsg, "required") || strings.Contains(errMsg, "invalid") {
			utils.Error(w, http.StatusBadRequest, errMsg, "BAD_REQUEST")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to update user role: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer user role updated successfully", updatedUser)
}

// --- Task S9: Customer 360 Cross-Module Handlers ---

// GetCustomerShipments handles GET /api/v1/sportal/organizations/{id}/shipments
func (h *Handler) GetCustomerShipments(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.service.GetCustomerShipments(r.Context(), userCtx, orgID, limit)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve customer shipments: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer shipments retrieved successfully", items)
}

// GetCustomerInvoices handles GET /api/v1/sportal/organizations/{id}/invoices
func (h *Handler) GetCustomerInvoices(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.service.GetCustomerInvoices(r.Context(), userCtx, orgID, limit)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve customer invoices: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer invoices retrieved successfully", items)
}

// GetCustomerContracts handles GET /api/v1/sportal/organizations/{id}/contracts
func (h *Handler) GetCustomerContracts(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.service.GetCustomerContracts(r.Context(), userCtx, orgID, limit)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve customer contracts: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer contracts retrieved successfully", items)
}

// GetCustomerExceptions handles GET /api/v1/sportal/organizations/{id}/exceptions
func (h *Handler) GetCustomerExceptions(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.service.GetCustomerExceptions(r.Context(), userCtx, orgID, limit)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve customer exceptions: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer exceptions retrieved successfully", items)
}

// GetCustomerIntegrations handles GET /api/v1/sportal/organizations/{id}/integrations
func (h *Handler) GetCustomerIntegrations(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	if r.URL.Query().Get("format") == "legacy" {
		items, err := h.service.GetCustomerIntegrations(r.Context(), userCtx, orgID)
		if err != nil {
			if strings.Contains(err.Error(), "forbidden") {
				utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
				return
			}
			if strings.Contains(err.Error(), "not found") {
				utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
				return
			}
			utils.Error(w, http.StatusInternalServerError, "Failed to retrieve customer integrations: "+err.Error(), "INTERNAL_ERROR")
			return
		}
		utils.Success(w, http.StatusOK, "Customer integrations retrieved successfully", items)
		return
	}

	overview, err := h.service.GetCustomerIntegrationsOverview(r.Context(), userCtx, orgID)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve customer integrations: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer integrations overview retrieved successfully", overview)
}

// GetCustomerDocuments handles GET /api/v1/sportal/organizations/{id}/documents
func (h *Handler) GetCustomerDocuments(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	if r.URL.Query().Get("page") != "" || r.URL.Query().Get("paginated") == "true" || r.URL.Query().Get("search") != "" || r.URL.Query().Get("doc_type") != "" || r.URL.Query().Get("expiry_filter") != "" {
		h.GetCustomerDocumentsPaginated(w, r)
		return
	}

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.service.GetCustomerDocuments(r.Context(), userCtx, orgID, limit)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve customer documents: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer documents retrieved successfully", items)
}

// GetCustomerAiSummary handles GET /api/v1/sportal/organizations/{id}/ai-summary
func (h *Handler) GetCustomerAiSummary(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	summary, err := h.service.GetCustomerAiSummary(r.Context(), userCtx, orgID)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve customer AI summary: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer AI summary retrieved successfully", summary)
}

// ============================================================================
// TASK S10: CUSTOMER USAGE & PLATFORM ANALYTICS HANDLERS
// ============================================================================

// GetCustomerUsageAnalytics handles GET /api/v1/sportal/organizations/{id}/usage
func (h *Handler) GetCustomerUsageAnalytics(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	period := r.URL.Query().Get("period")
	analytics, err := h.service.GetCustomerUsageAnalytics(r.Context(), userCtx, orgID, period)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve customer usage analytics: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer usage analytics retrieved successfully", analytics)
}

// GetPlatformUsageAnalytics handles GET /api/v1/sportal/usage
func (h *Handler) GetPlatformUsageAnalytics(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	// If orgId query parameter is provided, scope to that specific organization
	if orgParam := r.URL.Query().Get("orgId"); orgParam != "" {
		if parsedID, err := strconv.ParseInt(orgParam, 10, 64); err == nil && parsedID > 0 {
			analytics, err := h.service.GetCustomerUsageAnalytics(r.Context(), userCtx, parsedID, r.URL.Query().Get("period"))
			if err != nil {
				if strings.Contains(err.Error(), "forbidden") {
					utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
					return
				}
				if strings.Contains(err.Error(), "not found") {
					utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
					return
				}
				utils.Error(w, http.StatusInternalServerError, "Failed to retrieve organization usage analytics: "+err.Error(), "INTERNAL_ERROR")
				return
			}
			utils.Success(w, http.StatusOK, "Organization usage analytics retrieved successfully", analytics)
			return
		}
	}

	period := r.URL.Query().Get("period")
	analytics, err := h.service.GetPlatformUsageAnalytics(r.Context(), userCtx, period)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve platform usage analytics: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Platform usage analytics retrieved successfully", analytics)
}

// ============================================================================
// TASK S11: CUSTOMER HEALTH, CUSTOMER SUCCESS INTELLIGENCE & RISK SIGNALS
// ============================================================================

// GetCustomerHealth handles GET /api/v1/sportal/organizations/{id}/health
func (h *Handler) GetCustomerHealth(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	health, err := h.service.GetCustomerHealth(r.Context(), userCtx, orgID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve customer health: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer health retrieved successfully", health)
}

// GetPlatformHealth handles GET /api/v1/sportal/customer-health
func (h *Handler) GetPlatformHealth(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	// If orgId query parameter is provided, scope to that specific organization
	if orgParam := r.URL.Query().Get("orgId"); orgParam != "" {
		if parsedID, err := strconv.ParseInt(orgParam, 10, 64); err == nil && parsedID > 0 {
			health, err := h.service.GetCustomerHealth(r.Context(), userCtx, parsedID)
			if err != nil {
				if strings.Contains(err.Error(), "forbidden") {
					utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
					return
				}
				if strings.Contains(err.Error(), "not found") {
					utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
					return
				}
				utils.Error(w, http.StatusInternalServerError, "Failed to retrieve customer health: "+err.Error(), "INTERNAL_ERROR")
				return
			}
			utils.Success(w, http.StatusOK, "Customer health retrieved successfully", health)
			return
		}
	}

	health, err := h.service.GetPlatformHealth(r.Context(), userCtx)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve platform customer health: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Platform customer health retrieved successfully", health)
}

// CreateCustomerNote handles POST /api/v1/sportal/organizations/{id}/health/notes
func (h *Handler) CreateCustomerNote(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	var req CreateCustomerNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	note, err := h.service.CreateCustomerNote(r.Context(), userCtx, orgID, req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		if strings.Contains(errMsg, "validation") || strings.Contains(errMsg, "empty") {
			utils.Error(w, http.StatusBadRequest, errMsg, "VALIDATION_ERROR")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to create customer note: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusCreated, "Customer note created successfully", note)
}

// GetCustomerNotes handles GET /api/v1/sportal/organizations/{id}/health/notes
func (h *Handler) GetCustomerNotes(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	notes, err := h.service.GetCustomerNotes(r.Context(), userCtx, orgID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve customer notes: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer notes retrieved successfully", notes)
}

// ==============================================================================
// TASK S12: SPORTAL CUSTOMER INTEGRATIONS & CONNECTIVITY MANAGEMENT
// ==============================================================================

// GetPlatformIntegrations handles GET /api/v1/sportal/integrations
func (h *Handler) GetPlatformIntegrations(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	overview, err := h.service.GetPlatformIntegrationsOverview(r.Context(), userCtx)
	if err != nil {
		if strings.Contains(err.Error(), "forbidden") {
			utils.Error(w, http.StatusForbidden, err.Error(), "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve platform integrations: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Platform integrations overview retrieved successfully", overview)
}

// ToggleCustomerIntegration handles POST /api/v1/sportal/organizations/{id}/integrations/toggle
func (h *Handler) ToggleCustomerIntegration(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	var req IntegrationActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload: "+err.Error(), "BAD_REQUEST")
		return
	}

	if req.IntegrationType == "" || req.ProviderName == "" {
		utils.Error(w, http.StatusBadRequest, "integration_type and provider_name are required", "VALIDATION_FAILED")
		return
	}

	result, err := h.service.ToggleCustomerIntegration(r.Context(), userCtx, orgID, req)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to update integration state: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Integration status updated successfully", result)
}

// TestCustomerIntegrationConnection handles POST /api/v1/sportal/organizations/{id}/integrations/test
func (h *Handler) TestCustomerIntegrationConnection(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	var req IntegrationActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload: "+err.Error(), "BAD_REQUEST")
		return
	}

	if req.IntegrationType == "" || req.ProviderName == "" {
		utils.Error(w, http.StatusBadRequest, "integration_type and provider_name are required", "VALIDATION_FAILED")
		return
	}

	result, err := h.service.TestCustomerIntegrationConnection(r.Context(), userCtx, orgID, req)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to test integration connectivity: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Integration connectivity test completed", result)
}

// GetCustomerWebhooks handles GET /api/v1/sportal/organizations/{id}/integrations/webhooks
func (h *Handler) GetCustomerWebhooks(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.service.GetCustomerWebhooks(r.Context(), userCtx, orgID, limit)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve webhook events: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer webhook events retrieved successfully", items)
}

// GetCustomerSyncJobs handles GET /api/v1/sportal/organizations/{id}/integrations/sync-jobs
func (h *Handler) GetCustomerSyncJobs(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.service.GetCustomerSyncJobs(r.Context(), userCtx, orgID, limit)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve sync jobs: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer sync jobs retrieved successfully", items)
}

// ============================================================================
// Task S13: Customer Documents, Compliance, Contracts & Customer Records
// ============================================================================

// GetCustomerDocumentsPaginated handles GET /api/v1/sportal/organizations/{id}/documents
func (h *Handler) GetCustomerDocumentsPaginated(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	params := DocumentListParams{
		Page:         1,
		Limit:        20,
		Search:       r.URL.Query().Get("search"),
		DocType:      r.URL.Query().Get("doc_type"),
		Status:       r.URL.Query().Get("status"),
		ExpiryFilter: r.URL.Query().Get("expiry_filter"),
	}

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			params.Page = parsed
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			params.Limit = parsed
		}
	}

	res, err := h.service.GetCustomerDocumentsPaginated(r.Context(), userCtx, orgID, params)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve customer documents: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer documents retrieved successfully", res)
}

// GetCustomerDocumentDetail handles GET /api/v1/sportal/organizations/{id}/documents/{docId}
func (h *Handler) GetCustomerDocumentDetail(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	docID, err := strconv.ParseInt(chi.URLParam(r, "docId"), 10, 64)
	if err != nil || docID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid document ID", "INVALID_ID")
		return
	}

	detail, err := h.service.GetCustomerDocumentDetail(r.Context(), userCtx, orgID, docID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Document not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve document detail: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer document detail retrieved successfully", detail)
}

// DownloadCustomerDocument handles GET /api/v1/sportal/organizations/{id}/documents/{docId}/download
func (h *Handler) DownloadCustomerDocument(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	docID, err := strconv.ParseInt(chi.URLParam(r, "docId"), 10, 64)
	if err != nil || docID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid document ID", "INVALID_ID")
		return
	}

	data, mimeType, fileName, err := h.service.GetCustomerDocumentFile(r.Context(), userCtx, orgID, docID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Document not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to download document: "+errMsg, "INTERNAL_ERROR")
		return
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// GetCustomerComplianceOverview handles GET /api/v1/sportal/organizations/{id}/compliance
func (h *Handler) GetCustomerComplianceOverview(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	compliance, err := h.service.GetCustomerComplianceOverview(r.Context(), userCtx, orgID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve compliance overview: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer compliance overview retrieved successfully", compliance)
}

// GetCustomerContractsOverview handles GET /api/v1/sportal/organizations/{id}/contracts
func (h *Handler) GetCustomerContractsOverview(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	contracts, err := h.service.GetCustomerContractsOverview(r.Context(), userCtx, orgID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve contracts overview: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer contracts overview retrieved successfully", contracts)
}

// GetPlatformDocumentsOverview handles GET /api/v1/sportal/documents/overview
func (h *Handler) GetPlatformDocumentsOverview(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	overview, err := h.service.GetPlatformDocumentsOverview(r.Context(), userCtx)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve platform documents overview: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Platform documents overview retrieved successfully", overview)
}

// UpdateCustomerDocumentStatus handles PATCH /api/v1/sportal/organizations/{id}/documents/{docId}/status
func (h *Handler) UpdateCustomerDocumentStatus(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID", "INVALID_ID")
		return
	}

	docID, err := strconv.ParseInt(chi.URLParam(r, "docId"), 10, 64)
	if err != nil || docID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid document ID", "INVALID_ID")
		return
	}

	var payload struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	updated, err := h.service.UpdateCustomerDocumentStatus(r.Context(), userCtx, orgID, docID, payload.Status, payload.Reason)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusBadRequest, errMsg, "UPDATE_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Document status updated successfully", updated)
}

// ============================================================================
// TASK S16: SPORTAL AI, INTERNAL INTELLIGENCE & GOVERNED AI OPERATIONS HANDLERS
// ============================================================================

// QueryAi handles POST /api/v1/sportal/ai/query
func (h *Handler) QueryAi(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	var req SPortalAiQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid AI query request payload", "INVALID_PAYLOAD")
		return
	}

	resp, err := h.service.QueryAi(r.Context(), userCtx, req)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to process SPortal AI query: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "AI intelligence query processed successfully", resp)
}

// ExecuteAiAction handles POST /api/v1/sportal/ai/action
func (h *Handler) ExecuteAiAction(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	var req SPortalAiActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid AI action request payload", "INVALID_PAYLOAD")
		return
	}

	resp, err := h.service.ExecuteAiAction(r.Context(), userCtx, req)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to execute AI action: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "AI action executed successfully", resp)
}

// GetAiWorkforceOverview handles GET /api/v1/sportal/ai/workforce
func (h *Handler) GetAiWorkforceOverview(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	overview, err := h.service.GetAiWorkforceOverview(r.Context(), userCtx)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve AI workforce telemetry: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "AI workforce telemetry retrieved successfully", overview)
}

// ListAiRecommendations handles GET /api/v1/sportal/ai/recommendations
func (h *Handler) ListAiRecommendations(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	var orgIDPtr *int64
	if orgIDStr := r.URL.Query().Get("org_id"); orgIDStr != "" {
		if id, err := strconv.ParseInt(orgIDStr, 10, 64); err == nil && id > 0 {
			orgIDPtr = &id
		}
	}

	recs, err := h.service.ListAiRecommendations(r.Context(), userCtx, orgIDPtr)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve AI recommendations: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "AI recommendations retrieved successfully", recs)
}

// GetCustomerAiContext handles GET /api/v1/sportal/organizations/{id}/ai-context
func (h *Handler) GetCustomerAiContext(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	orgID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid organization ID parameter", "BAD_REQUEST")
		return
	}

	data, err := h.service.GetCustomerAiContext(r.Context(), userCtx, orgID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve customer AI context: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Customer AI context retrieved successfully", data)
}

// --- Task S17: SPortal Settings Handlers ---

// GetSettingsOverview handles GET /api/v1/sportal/settings/overview
func (h *Handler) GetSettingsOverview(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	overview, err := h.service.GetSettingsOverview(r.Context(), userCtx)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve settings overview: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Settings overview retrieved successfully", overview)
}

// GetInternalUserProfile handles GET /api/v1/sportal/settings/profile
func (h *Handler) GetInternalUserProfile(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	user, role, prefs, err := h.service.GetInternalUserProfile(r.Context(), userCtx)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve user profile: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "User profile retrieved successfully", map[string]interface{}{
		"user":        user,
		"role":        role,
		"preferences": prefs,
	})
}

// UpdateInternalUserProfile handles PATCH /api/v1/sportal/settings/profile
func (h *Handler) UpdateInternalUserProfile(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	var req SPortalProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body: "+err.Error(), "BAD_REQUEST")
		return
	}

	user, prefs, err := h.service.UpdateInternalUserProfile(r.Context(), userCtx, req, r.RemoteAddr, r.UserAgent())
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to update profile: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "User profile and preferences updated successfully", map[string]interface{}{
		"user":        user,
		"preferences": prefs,
	})
}

// GetPlatformSettings handles GET /api/v1/sportal/settings/platform
func (h *Handler) GetPlatformSettings(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	settings, err := h.service.GetPlatformSettings(r.Context(), userCtx)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve platform settings: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Platform settings retrieved successfully", settings)
}

// UpdatePlatformSetting handles PATCH /api/v1/sportal/settings/platform/{key}
func (h *Handler) UpdatePlatformSetting(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	key := chi.URLParam(r, "key")
	if key == "" {
		utils.Error(w, http.StatusBadRequest, "Missing setting key parameter", "BAD_REQUEST")
		return
	}

	var req SPortalPlatformSettingUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body: "+err.Error(), "BAD_REQUEST")
		return
	}

	if err := h.service.UpdatePlatformSetting(r.Context(), userCtx, key, req, r.RemoteAddr, r.UserAgent()); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to update platform setting: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, fmt.Sprintf("Platform setting '%s' updated successfully", key), nil)
}

// GetFeatureFlags handles GET /api/v1/sportal/settings/feature-flags
func (h *Handler) GetFeatureFlags(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	flags, err := h.service.GetFeatureFlags(r.Context(), userCtx)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve feature flags: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Feature flags retrieved successfully", flags)
}

// UpdateFeatureFlag handles PATCH /api/v1/sportal/settings/feature-flags/{key}
func (h *Handler) UpdateFeatureFlag(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	key := chi.URLParam(r, "key")
	if key == "" {
		utils.Error(w, http.StatusBadRequest, "Missing feature flag key parameter", "BAD_REQUEST")
		return
	}

	var req SPortalFeatureFlagUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body: "+err.Error(), "BAD_REQUEST")
		return
	}

	if err := h.service.UpdateFeatureFlag(r.Context(), userCtx, key, req, r.RemoteAddr, r.UserAgent()); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to update feature flag: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, fmt.Sprintf("Feature flag '%s' updated successfully", key), nil)
}

// GetAutonomyPolicies handles GET /api/v1/sportal/settings/autonomy
func (h *Handler) GetAutonomyPolicies(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	policies, globalHalt, err := h.service.GetAutonomyPolicies(r.Context(), userCtx)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve autonomy policies: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Autonomy policies retrieved successfully", map[string]interface{}{
		"policies":             policies,
		"global_emergency_halt": globalHalt,
	})
}

// TriggerEmergencyHalt handles POST /api/v1/sportal/settings/autonomy/emergency-halt
func (h *Handler) TriggerEmergencyHalt(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	var req SPortalEmergencyHaltRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body: "+err.Error(), "BAD_REQUEST")
		return
	}

	if err := h.service.TriggerEmergencyHalt(r.Context(), userCtx, req, r.RemoteAddr, r.UserAgent()); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to set emergency halt: "+errMsg, "INTERNAL_ERROR")
		return
	}

	statusMsg := "Emergency halt engaged across autonomous modules"
	if !req.HaltActive {
		statusMsg = "Emergency halt disengaged. Autonomous operations resumed"
	}

	utils.Success(w, http.StatusOK, statusMsg, map[string]interface{}{
		"halt_active": req.HaltActive,
		"module":      req.Module,
	})
}

// GetIntegrationSettings handles GET /api/v1/sportal/settings/integrations
func (h *Handler) GetIntegrationSettings(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	list, err := h.service.GetIntegrationSettings(r.Context(), userCtx)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve integration settings: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Integration settings retrieved successfully", list)
}

// ToggleIntegrationSetting handles PATCH /api/v1/sportal/settings/integrations/{type}/toggle
func (h *Handler) ToggleIntegrationSetting(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	integType := chi.URLParam(r, "type")
	if integType == "" {
		utils.Error(w, http.StatusBadRequest, "Missing integration type parameter", "BAD_REQUEST")
		return
	}

	var req SPortalIntegrationToggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body: "+err.Error(), "BAD_REQUEST")
		return
	}

	if err := h.service.ToggleIntegrationSetting(r.Context(), userCtx, integType, req, r.RemoteAddr, r.UserAgent()); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to toggle integration: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, fmt.Sprintf("Integration '%s' toggled successfully", integType), nil)
}

// GetRecentAdministrativeAudits handles GET /api/v1/sportal/settings/audit
func (h *Handler) GetRecentAdministrativeAudits(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	audits, err := h.service.GetRecentAdministrativeAudits(r.Context(), userCtx, limit)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve audit logs: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Administrative audits retrieved successfully", audits)
}

// GetOperationsHealthSummary handles GET /api/v1/sportal/settings/operations
func (h *Handler) GetOperationsHealthSummary(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	summary, err := h.service.GetOperationsHealthSummary(r.Context(), userCtx)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve operations health: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Operations health summary retrieved successfully", summary)
}

// -----------------------------------------------------------------------------
// P12: Support Center, Activity Timeline, Notifications & Forensic Audit
// -----------------------------------------------------------------------------

// GetSupportCases handles GET /api/v1/sportal/support/cases
func (h *Handler) GetSupportCases(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	var orgID int64
	if orgStr := q.Get("org_id"); orgStr != "" {
		orgID, _ = strconv.ParseInt(orgStr, 10, 64)
	}
	status := q.Get("status")
	severity := q.Get("severity")
	search := q.Get("search")
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	res, err := h.service.GetSupportCases(r.Context(), userCtx, orgID, status, severity, search, page, limit)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve support cases: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Support cases retrieved successfully", res)
}

// GetSupportCaseDetail handles GET /api/v1/sportal/support/cases/{id}
func (h *Handler) GetSupportCaseDetail(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	caseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid case ID", "INVALID_REQUEST")
		return
	}

	item, err := h.service.GetSupportCaseDetail(r.Context(), userCtx, caseID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve support case detail: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Support case detail retrieved successfully", item)
}

// UpdateSupportCaseStatus handles PATCH /api/v1/sportal/support/cases/{id}/status
func (h *Handler) UpdateSupportCaseStatus(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	caseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid case ID", "INVALID_REQUEST")
		return
	}

	var req SPortalUpdateCaseStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_JSON")
		return
	}

	if err := h.service.UpdateSupportCaseStatus(r.Context(), userCtx, caseID, req); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "invalid status") {
			utils.Error(w, http.StatusBadRequest, errMsg, "INVALID_STATUS")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to update support case: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Support case updated successfully", map[string]interface{}{
		"case_id": caseID,
		"status":  req.Status,
	})
}

// AddSupportCaseNote handles POST /api/v1/sportal/support/cases/{id}/notes
func (h *Handler) AddSupportCaseNote(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	caseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid case ID", "INVALID_REQUEST")
		return
	}

	var req SPortalAddCaseNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_JSON")
		return
	}

	if err := h.service.AddSupportCaseNote(r.Context(), userCtx, caseID, req); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, errMsg, "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to add support note: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusCreated, "Internal support note recorded successfully", map[string]interface{}{
		"case_id":          caseID,
		"is_internal_only": true,
	})
}

// CreateSupportCase handles POST /api/v1/sportal/support/cases
func (h *Handler) CreateSupportCase(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	var req SPortalCreateCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_JSON")
		return
	}

	caseID, err := h.service.CreateSupportCase(r.Context(), userCtx, req)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to create support case: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusCreated, "Support case created successfully", map[string]interface{}{
		"case_id":   caseID,
		"case_code": fmt.Sprintf("CAS-%04d", caseID),
	})
}

// GetNotificationsList handles GET /api/v1/sportal/notifications
func (h *Handler) GetNotificationsList(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	var orgID int64
	if orgStr := q.Get("org_id"); orgStr != "" {
		orgID, _ = strconv.ParseInt(orgStr, 10, 64)
	}

	var isRead *bool
	if readStr := q.Get("is_read"); readStr != "" {
		b := readStr == "true" || readStr == "1"
		isRead = &b
	}

	severity := q.Get("severity")
	deliveryStatus := q.Get("delivery_status")
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	res, err := h.service.GetNotificationsList(r.Context(), userCtx, orgID, isRead, severity, deliveryStatus, page, limit)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve notifications: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Notifications retrieved successfully", res)
}

// MarkNotificationRead handles PATCH /api/v1/sportal/notifications/{id}/read
func (h *Handler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	notifID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid notification ID", "INVALID_REQUEST")
		return
	}

	if err := h.service.MarkNotificationRead(r.Context(), userCtx, notifID); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to mark notification read: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Notification marked as read", map[string]interface{}{
		"id":      notifID,
		"is_read": true,
	})
}

// MarkAllNotificationsRead handles POST /api/v1/sportal/notifications/mark-all-read
func (h *Handler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	var orgID int64
	if orgStr := r.URL.Query().Get("org_id"); orgStr != "" {
		orgID, _ = strconv.ParseInt(orgStr, 10, 64)
	}

	if err := h.service.MarkAllNotificationsRead(r.Context(), userCtx, orgID); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to mark all read: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "All notifications marked as read", map[string]interface{}{
		"org_id": orgID,
	})
}

// AcknowledgeNotification handles PATCH /api/v1/sportal/notifications/{id}/acknowledge
func (h *Handler) AcknowledgeNotification(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	notifID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid notification ID", "INVALID_REQUEST")
		return
	}

	if err := h.service.AcknowledgeNotification(r.Context(), userCtx, notifID); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to acknowledge notification: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Notification acknowledged successfully", map[string]interface{}{
		"id":              notifID,
		"is_acknowledged": true,
	})
}

// GetUnifiedActivityTimeline handles GET /api/v1/sportal/activity/timeline
func (h *Handler) GetUnifiedActivityTimeline(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	var orgID int64
	if orgStr := q.Get("org_id"); orgStr != "" {
		orgID, _ = strconv.ParseInt(orgStr, 10, 64)
	}
	category := q.Get("category")
	limit, _ := strconv.Atoi(q.Get("limit"))

	items, err := h.service.GetUnifiedActivityTimeline(r.Context(), userCtx, orgID, category, limit)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve activity timeline: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Unified activity timeline retrieved successfully", items)
}

// SearchAuditLogs handles GET /api/v1/sportal/audit/search
func (h *Handler) SearchAuditLogs(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Missing authentication context", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	var orgID int64
	if orgStr := q.Get("org_id"); orgStr != "" {
		orgID, _ = strconv.ParseInt(orgStr, 10, 64)
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := SPortalAuditSearchFilter{
		OrgID:     orgID,
		Module:    q.Get("module"),
		Action:    q.Get("action"),
		Actor:     q.Get("actor"),
		Result:    q.Get("result"),
		Search:    q.Get("search"),
		StartDate: q.Get("start_date"),
		EndDate:   q.Get("end_date"),
		Limit:     limit,
		Offset:    offset,
	}

	logs, err := h.service.SearchAuditLogs(r.Context(), userCtx, filter)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to search audit logs: "+errMsg, "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Audit logs retrieved successfully", logs)
}

// ── Demo Requests Handlers ───────────────────────────────────────────────────

// SubmitDemoRequest handles public demo request submissions from the website.
func (h *Handler) SubmitDemoRequest(w http.ResponseWriter, r *http.Request) {
	var req CreateDemoRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	clientIP := getClientIP(r)
	userAgent := r.UserAgent()

	demo, err := h.service.SubmitDemoRequest(r.Context(), req, clientIP, userAgent)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "is required") {
			utils.Error(w, http.StatusBadRequest, errMsg, "VALIDATION_ERROR")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to submit demo request", "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusCreated, "Demo request submitted successfully", demo)
}

// ListDemoRequests returns paginated demo requests for SPortal staff.
func (h *Handler) ListDemoRequests(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	params := DemoRequestListParams{
		Search: q.Get("search"),
		Status: q.Get("status"),
		Page:   page,
		Limit:  limit,
	}

	result, err := h.service.ListDemoRequests(r.Context(), userCtx, params)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to list demo requests", "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Demo requests retrieved successfully", result)
}

// GetDemoRequest returns a single demo request by ID.
func (h *Handler) GetDemoRequest(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid demo request ID", "INVALID_ID")
		return
	}

	demo, err := h.service.GetDemoRequestByID(r.Context(), userCtx, id)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Demo request not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to get demo request", "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Demo request retrieved successfully", demo)
}

// UpdateDemoRequest updates demo request status/notes/assignment.
func (h *Handler) UpdateDemoRequest(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid demo request ID", "INVALID_ID")
		return
	}

	var req UpdateDemoRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	demo, err := h.service.UpdateDemoRequest(r.Context(), userCtx, id, req)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "forbidden") {
			utils.Error(w, http.StatusForbidden, errMsg, "FORBIDDEN")
			return
		}
		if strings.Contains(errMsg, "not found") {
			utils.Error(w, http.StatusNotFound, "Demo request not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to update demo request", "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Demo request updated successfully", demo)
}
