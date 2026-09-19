package bcontext

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for the Unified Business Context and Intelligence layer.
type Handler struct {
	svc Service
}

// NewHandler creates a new context and intelligence HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// GetContext handles GET /api/v1/context/{type}/{id}.
func (h *Handler) GetContext(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	entityTypeStr := strings.ToUpper(chi.URLParam(r, "type"))
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid entity ID", "INVALID_ID")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")

	bCtx, err := h.svc.GetBusinessContext(r.Context(), ContextRequest{
		OrgID:         userCtx.OrgID,
		UserID:        userCtx.UserID,
		UserRole:      userCtx.Role,
		PrimaryType:   EntityType(entityTypeStr),
		PrimaryID:     id,
		CorrelationID: correlationID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "CONTEXT_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    bCtx,
	})
}

// GetInsight handles POST /api/v1/intelligence/insight.
func (h *Handler) GetInsight(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	var req struct {
		EntityType string `json:"entity_type"`
		EntityID   int64  `json:"entity_id"`
		Question   string `json:"question,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_JSON")
		return
	}

	if req.EntityID <= 0 || req.EntityType == "" {
		utils.Error(w, http.StatusBadRequest, "entity_type and entity_id are required", "VALIDATION_FAILED")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")

	insight, err := h.svc.GenerateInsight(r.Context(), InsightRequest{
		OrgID:         userCtx.OrgID,
		UserID:        userCtx.UserID,
		UserRole:      userCtx.Role,
		EntityType:    EntityType(strings.ToUpper(req.EntityType)),
		EntityID:      req.EntityID,
		Question:      req.Question,
		CorrelationID: correlationID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INSIGHT_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    insight,
	})
}

// InternalGetContext handles POST /internal/context/retrieve for the Python AI Sidecar.
func (h *Handler) InternalGetContext(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized internal service request", "UNAUTHORIZED")
		return
	}

	var req ContextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.OrgID <= 0 || req.PrimaryID <= 0 || req.PrimaryType == "" {
		utils.Error(w, http.StatusBadRequest, "org_id, primary_type, and primary_id are required", "VALIDATION_ERROR")
		return
	}

	bCtx, err := h.svc.GetBusinessContext(r.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    bCtx,
	})
}

// InternalGetInsight handles POST /internal/intelligence/insight for the Python AI Sidecar.
func (h *Handler) InternalGetInsight(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized internal service request", "UNAUTHORIZED")
		return
	}

	var req InsightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.OrgID <= 0 || req.EntityID <= 0 || req.EntityType == "" {
		utils.Error(w, http.StatusBadRequest, "org_id, entity_type, and entity_id are required", "VALIDATION_ERROR")
		return
	}

	insight, err := h.svc.GenerateInsight(r.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    insight,
	})
}

// GetCustomerIntelligence handles GET /api/v1/customers/{id}/intelligence.
func (h *Handler) GetCustomerIntelligence(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid customer ID", "INVALID_ID")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")

	intel, err := h.svc.GetCustomer360Intelligence(r.Context(), userCtx.OrgID, id, correlationID, userCtx.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTELLIGENCE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    intel,
	})
}

// InternalGetCustomerIntelligence handles POST /internal/customers/intelligence for the Python AI Sidecar.
func (h *Handler) InternalGetCustomerIntelligence(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized internal service request", "UNAUTHORIZED")
		return
	}

	var req struct {
		OrgID         int64  `json:"org_id"`
		CustomerID    int64  `json:"customer_id"`
		CorrelationID string `json:"correlation_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.OrgID <= 0 || req.CustomerID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id and customer_id are required", "VALIDATION_ERROR")
		return
	}

	intel, err := h.svc.GetCustomer360Intelligence(r.Context(), req.OrgID, req.CustomerID, req.CorrelationID, 0)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    intel,
	})
}

// GetRFQIntelligence handles GET /api/v1/rfqs/{id}/intelligence.
func (h *Handler) GetRFQIntelligence(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid RFQ ID", "INVALID_ID")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")

	intel, err := h.svc.GetRFQ360PricingIntelligence(r.Context(), userCtx.OrgID, id, correlationID, userCtx.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTELLIGENCE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    intel,
	})
}

// InternalGetRFQIntelligence handles POST /internal/rfqs/intelligence for the Python AI Sidecar.
func (h *Handler) InternalGetRFQIntelligence(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized internal service request", "UNAUTHORIZED")
		return
	}

	var req struct {
		OrgID         int64  `json:"org_id"`
		RFQID         int64  `json:"rfq_id"`
		CorrelationID string `json:"correlation_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.OrgID <= 0 || req.RFQID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id and rfq_id are required", "VALIDATION_ERROR")
		return
	}

	intel, err := h.svc.GetRFQ360PricingIntelligence(r.Context(), req.OrgID, req.RFQID, req.CorrelationID, 0)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    intel,
	})
}

// GetShipmentIntelligence handles GET /api/v1/shipments/{id}/intelligence.
func (h *Handler) GetShipmentIntelligence(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment ID", "INVALID_ID")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")

	intel, err := h.svc.GetShipment360OperationsIntelligence(r.Context(), userCtx.OrgID, id, correlationID, userCtx.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTELLIGENCE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    intel,
	})
}

// InternalGetShipmentIntelligence handles POST /internal/shipments/intelligence for the Python AI Sidecar.
func (h *Handler) InternalGetShipmentIntelligence(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized internal service request", "UNAUTHORIZED")
		return
	}

	var req struct {
		OrgID         int64  `json:"org_id"`
		ShipmentID    int64  `json:"shipment_id"`
		CorrelationID string `json:"correlation_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.OrgID <= 0 || req.ShipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id and shipment_id are required", "VALIDATION_ERROR")
		return
	}

	intel, err := h.svc.GetShipment360OperationsIntelligence(r.Context(), req.OrgID, req.ShipmentID, req.CorrelationID, 0)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    intel,
	})
}

// GetOrgOperationsSummary handles GET /api/v1/shipments/operations-summary.
func (h *Handler) GetOrgOperationsSummary(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")

	summary, err := h.svc.GetOrgOperationsSummary(r.Context(), userCtx.OrgID, correlationID, userCtx.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "OPERATIONS_SUMMARY_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    summary,
	})
}

// InternalGetOrgOperationsSummary handles POST /internal/shipments/operations-summary for the Python AI Sidecar.
func (h *Handler) InternalGetOrgOperationsSummary(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized internal service request", "UNAUTHORIZED")
		return
	}

	var req struct {
		OrgID         int64  `json:"org_id"`
		CorrelationID string `json:"correlation_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.OrgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id is required", "VALIDATION_ERROR")
		return
	}

	summary, err := h.svc.GetOrgOperationsSummary(r.Context(), req.OrgID, req.CorrelationID, 0)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    summary,
	})
}

// GetInvoiceIntelligence handles GET /api/v1/invoices/{id}/intelligence.
func (h *Handler) GetInvoiceIntelligence(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice ID", "INVALID_ID")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")

	intel, err := h.svc.GetInvoice360FinanceIntelligence(r.Context(), userCtx.OrgID, id, correlationID, userCtx.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "FINANCE_INTELLIGENCE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    intel,
	})
}

// InternalGetInvoiceIntelligence handles POST /internal/invoices/intelligence for the Python AI Sidecar.
func (h *Handler) InternalGetInvoiceIntelligence(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized internal service request", "UNAUTHORIZED")
		return
	}

	var req struct {
		OrgID         int64  `json:"org_id"`
		InvoiceID     int64  `json:"invoice_id"`
		CorrelationID string `json:"correlation_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.OrgID <= 0 || req.InvoiceID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id and invoice_id are required", "VALIDATION_ERROR")
		return
	}

	intel, err := h.svc.GetInvoice360FinanceIntelligence(r.Context(), req.OrgID, req.InvoiceID, req.CorrelationID, 0)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    intel,
	})
}

// GetOrgFinanceSummary handles GET /api/v1/invoices/finance-summary.
func (h *Handler) GetOrgFinanceSummary(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")

	summary, err := h.svc.GetOrgFinanceSummary(r.Context(), userCtx.OrgID, correlationID, userCtx.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "FINANCE_SUMMARY_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    summary,
	})
}

// InternalGetOrgFinanceSummary handles POST /internal/invoices/finance-summary for the Python AI Sidecar.
func (h *Handler) InternalGetOrgFinanceSummary(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized internal service request", "UNAUTHORIZED")
		return
	}

	var req struct {
		OrgID         int64  `json:"org_id"`
		CorrelationID string `json:"correlation_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.OrgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id is required", "VALIDATION_ERROR")
		return
	}

	summary, err := h.svc.GetOrgFinanceSummary(r.Context(), req.OrgID, req.CorrelationID, 0)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    summary,
	})
}

// GetContractIntelligence handles GET /api/v1/contracts/{id}/intelligence.
func (h *Handler) GetContractIntelligence(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid contract ID", "INVALID_ID")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")

	intel, err := h.svc.GetContract360ComplianceIntelligence(r.Context(), userCtx.OrgID, id, correlationID, userCtx.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "CONTRACT_INTELLIGENCE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    intel,
	})
}

// InternalGetContractIntelligence handles POST /internal/contracts/intelligence for the Python AI Sidecar.
func (h *Handler) InternalGetContractIntelligence(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized internal service request", "UNAUTHORIZED")
		return
	}

	var req struct {
		OrgID         int64  `json:"org_id"`
		ContractID    int64  `json:"contract_id"`
		CorrelationID string `json:"correlation_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.OrgID <= 0 || req.ContractID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id and contract_id are required", "VALIDATION_ERROR")
		return
	}

	intel, err := h.svc.GetContract360ComplianceIntelligence(r.Context(), req.OrgID, req.ContractID, req.CorrelationID, 0)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    intel,
	})
}

// GetContractCoverage handles GET /api/v1/contracts/coverage-check?entity_type=...&entity_id=...
func (h *Handler) GetContractCoverage(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	entityType := r.URL.Query().Get("entity_type")
	entityIDStr := r.URL.Query().Get("entity_id")
	entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
	if err != nil || entityID <= 0 || entityType == "" {
		utils.Error(w, http.StatusBadRequest, "Valid entity_type and entity_id (>0) required", "VALIDATION_ERROR")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")

	eval, err := h.svc.GetContractCoverageForEntity(r.Context(), userCtx.OrgID, entityType, entityID, correlationID, userCtx.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "COVERAGE_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    eval,
	})
}

// InternalGetContractCoverage handles POST /internal/contracts/coverage for the Python AI Sidecar.
func (h *Handler) InternalGetContractCoverage(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized internal service request", "UNAUTHORIZED")
		return
	}

	var req struct {
		OrgID         int64  `json:"org_id"`
		EntityType    string `json:"entity_type"`
		EntityID      int64  `json:"entity_id"`
		CorrelationID string `json:"correlation_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.OrgID <= 0 || req.EntityID <= 0 || req.EntityType == "" {
		utils.Error(w, http.StatusBadRequest, "org_id, entity_type, and entity_id (>0) are required", "VALIDATION_ERROR")
		return
	}

	eval, err := h.svc.GetContractCoverageForEntity(r.Context(), req.OrgID, req.EntityType, req.EntityID, req.CorrelationID, 0)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    eval,
	})
}

// GetOrgContractComplianceSummary handles GET /api/v1/contracts/compliance-summary.
func (h *Handler) GetOrgContractComplianceSummary(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")

	summary, err := h.svc.GetOrgContractComplianceSummary(r.Context(), userCtx.OrgID, correlationID, userCtx.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "COMPLIANCE_SUMMARY_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    summary,
	})
}

// InternalGetOrgContractComplianceSummary handles POST /internal/contracts/compliance-summary for the Python AI Sidecar.
func (h *Handler) InternalGetOrgContractComplianceSummary(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized internal service request", "UNAUTHORIZED")
		return
	}

	var req struct {
		OrgID         int64  `json:"org_id"`
		CorrelationID string `json:"correlation_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.OrgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id is required", "VALIDATION_ERROR")
		return
	}

	summary, err := h.svc.GetOrgContractComplianceSummary(r.Context(), req.OrgID, req.CorrelationID, 0)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    summary,
	})
}

// GetCrossModuleInsights handles GET /api/v1/insights/cross-module.
func (h *Handler) GetCrossModuleInsights(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	entityType := r.URL.Query().Get("entity_type")
	entityIDStr := r.URL.Query().Get("entity_id")
	var entityID int64
	if entityIDStr != "" {
		parsed, err := strconv.ParseInt(entityIDStr, 10, 64)
		if err == nil {
			entityID = parsed
		}
	}

	correlationID := r.Header.Get("X-Correlation-Id")

	result, err := h.svc.GetCrossModuleInsights(r.Context(), userCtx.OrgID, entityType, entityID, correlationID, userCtx.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INSIGHTS_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// InternalGetCrossModuleInsights handles POST /internal/insights/cross-module for the Python AI Sidecar.
func (h *Handler) InternalGetCrossModuleInsights(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized internal service request", "UNAUTHORIZED")
		return
	}

	var req struct {
		OrgID         int64  `json:"org_id"`
		EntityType    string `json:"entity_type,omitempty"`
		EntityID      int64  `json:"entity_id,omitempty"`
		CorrelationID string `json:"correlation_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.OrgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id is required", "VALIDATION_ERROR")
		return
	}

	result, err := h.svc.GetCrossModuleInsights(r.Context(), req.OrgID, req.EntityType, req.EntityID, req.CorrelationID, 0)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// GetOrgCrossModuleSummary handles GET /api/v1/insights/summary.
func (h *Handler) GetOrgCrossModuleSummary(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "User context required", "UNAUTHORIZED")
		return
	}

	correlationID := r.Header.Get("X-Correlation-Id")

	summary, err := h.svc.GetOrgCrossModuleSummary(r.Context(), userCtx.OrgID, correlationID, userCtx.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INSIGHTS_SUMMARY_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    summary,
	})
}

// InternalGetOrgCrossModuleSummary handles POST /internal/insights/summary for the Python AI Sidecar.
func (h *Handler) InternalGetOrgCrossModuleSummary(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized internal service request", "UNAUTHORIZED")
		return
	}

	var req struct {
		OrgID         int64  `json:"org_id"`
		CorrelationID string `json:"correlation_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.OrgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id is required", "VALIDATION_ERROR")
		return
	}

	summary, err := h.svc.GetOrgCrossModuleSummary(r.Context(), req.OrgID, req.CorrelationID, 0)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    summary,
	})
}



