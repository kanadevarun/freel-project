package documents

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

// UploadDocument handles POST /api/v1/shipments/{id}/documents/upload
func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	shipmentIDStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(shipmentIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization user context", "UNAUTHORIZED")
		return
	}

	var req struct {
		DocType  string `json:"doc_type"`
		S3Key    string `json:"s3_key"`
		FileName string `json:"file_name"`
		FileType string `json:"file_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	if req.DocType == "" || req.S3Key == "" || req.FileName == "" {
		utils.Error(w, http.StatusBadRequest, "doc_type, s3_key, and file_name are required", "MISSING_PARAM")
		return
	}

	doc, err := h.svc.UploadDocument(r.Context(), userCtx.OrgID, shipmentID, req.DocType, req.S3Key, req.FileName, req.FileType)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to upload document: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Document uploaded and enqueued for compliance validation", doc)
}

// ListDocuments handles GET /api/v1/shipments/{id}/documents
func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	shipmentIDStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(shipmentIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization user context", "UNAUTHORIZED")
		return
	}

	docs, err := h.svc.GetDocumentsByShipment(r.Context(), userCtx.OrgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve documents: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	discrepancies, err := h.svc.GetDiscrepancies(r.Context(), userCtx.OrgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve discrepancies: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Retrieved shipment compliance documents details", map[string]interface{}{
		"documents":     docs,
		"discrepancies": discrepancies,
	})
}

// ListAllDocuments handles GET /api/v1/documents
func (h *Handler) ListAllDocuments(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization user context", "UNAUTHORIZED")
		return
	}

	docs, err := h.svc.GetDocumentsByOrg(r.Context(), userCtx.OrgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve organization documents: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Retrieved organization documents successfully", docs)
}

// ResolveDiscrepancy handles POST /api/v1/shipments/discrepancies/{id}/resolve
func (h *Handler) ResolveDiscrepancy(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid discrepancy id parameter", "INVALID_PARAM")
		return
	}

	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization user context", "UNAUTHORIZED")
		return
	}

	err = h.svc.ResolveDiscrepancy(r.Context(), userCtx.OrgID, id, userCtx.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to resolve discrepancy: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Discrepancy marked as resolved successfully", nil)
}

// CallbackInternal handles POST /internal/compliance/callback
func (h *Handler) CallbackInternal(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgID         int64                          `json:"org_id"`
		ShipmentID    int64                          `json:"shipment_id"`
		DocStatusList map[string]string              `json:"doc_status_list"` // docID -> JSON payload
		Discrepancies []*ShipmentDocumentDiscrepancy `json:"discrepancies"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	if req.OrgID <= 0 || req.ShipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id and shipment_id are required", "MISSING_PARAM")
		return
	}

	// Idempotency: Complete compliance audit records transactional saving
	err := h.svc.CompleteVerification(r.Context(), req.OrgID, req.ShipmentID, req.DocStatusList, req.Discrepancies)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to save verification callback: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Compliance validation callback complete", nil)
}

// ListDocumentsInternal handles GET /internal/shipments/{id}/documents
func (h *Handler) ListDocumentsInternal(w http.ResponseWriter, r *http.Request) {
	shipmentIDStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(shipmentIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id parameter", "INVALID_PARAM")
		return
	}

	orgIDStr := r.URL.Query().Get("org_id")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil || orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Missing or invalid org_id query parameter", "MISSING_PARAM")
		return
	}

	docs, err := h.svc.GetDocumentsByShipment(r.Context(), orgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve documents: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	discrepancies, err := h.svc.GetDiscrepancies(r.Context(), orgID, shipmentID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve discrepancies: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Retrieved shipment compliance documents details", map[string]interface{}{
		"documents":     docs,
		"discrepancies": discrepancies,
	})
}
