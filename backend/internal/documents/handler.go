package documents

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
	audit "github.com/freel/backend/internal/audit"
	auditDomain "github.com/freel/backend/internal/audit/domain"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// UploadGeneralDocument handles POST /api/v1/documents/upload
func (h *Handler) UploadGeneralDocument(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization user context", "UNAUTHORIZED")
		return
	}

	if err := r.ParseMultipartForm(25 << 20); err != nil {
		utils.Error(w, http.StatusBadRequest, "File too large (max 25MB)", "FILE_TOO_LARGE")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Missing 'file' parameter", "MISSING_FILE")
		return
	}
	defer file.Close()

	docType := strings.TrimSpace(r.FormValue("doc_type"))
	if docType == "" {
		docType = "OTHER"
	}

	var shipmentIDPtr, customerIDPtr, leadIDPtr, bookingIDPtr *int64

	if sIDStr := r.FormValue("shipment_id"); sIDStr != "" {
		if id, err := strconv.ParseInt(sIDStr, 10, 64); err == nil && id > 0 {
			shipmentIDPtr = &id
		}
	}
	if cIDStr := r.FormValue("customer_id"); cIDStr != "" {
		if id, err := strconv.ParseInt(cIDStr, 10, 64); err == nil && id > 0 {
			customerIDPtr = &id
		}
	}
	if lIDStr := r.FormValue("lead_id"); lIDStr != "" {
		if id, err := strconv.ParseInt(lIDStr, 10, 64); err == nil && id > 0 {
			leadIDPtr = &id
		}
	}
	if bIDStr := r.FormValue("booking_id"); bIDStr != "" {
		if id, err := strconv.ParseInt(bIDStr, 10, 64); err == nil && id > 0 {
			bookingIDPtr = &id
		}
	}

	ext := filepath.Ext(header.Filename)
	fileType := strings.ToUpper(strings.TrimPrefix(ext, "."))
	if fileType == "" {
		fileType = "UNKNOWN"
	}
	mimeType := header.Header.Get("Content-Type")

	origName := header.Filename
	doc := &ShipmentDocument{
		OrgID:            userCtx.OrgID,
		ShipmentID:       shipmentIDPtr,
		CustomerID:       customerIDPtr,
		LeadID:           leadIDPtr,
		BookingID:        bookingIDPtr,
		DocType:          strings.ToUpper(docType),
		FileName:         header.Filename,
		OriginalFileName: &origName,
		FileType:         fileType,
		MIMEType:         &mimeType,
		FileSize:         header.Size,
		Status:           "VERIFIED",
	}

	savedDoc, err := h.svc.UploadGeneralDocument(r.Context(), userCtx.OrgID, doc, file)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to upload document: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	actorID := userCtx.UserID
	_, _ = audit.Record(r.Context(), auditDomain.CreateAuditLogParams{
		OrgID:        userCtx.OrgID,
		ActorID:      &actorID,
		ActorRole:    userCtx.Role,
		Action:       auditDomain.ActionCreate,
		Module:       auditDomain.ModuleDocuments,
		ResourceType: "DOCUMENT",
		ResourceID:   savedDoc.ID,
		ResourceName: header.Filename,
		Description:  fmt.Sprintf("Uploaded document %s (%s)", header.Filename, strings.ToUpper(docType)),
		Result:       auditDomain.ResultSuccess,
		Metadata: map[string]interface{}{
			"doc_type":  docType,
			"file_name": header.Filename,
			"file_size": header.Size,
		},
	})

	utils.Success(w, http.StatusCreated, "Document uploaded successfully", savedDoc)
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

	customerIDStr := r.URL.Query().Get("customer_id")
	shipmentIDStr := r.URL.Query().Get("shipment_id")
	leadIDStr := r.URL.Query().Get("lead_id")
	bookingIDStr := r.URL.Query().Get("booking_id")

	if customerIDStr != "" || shipmentIDStr != "" || leadIDStr != "" || bookingIDStr != "" {
		filtered := make([]*ShipmentDocument, 0)
		for _, d := range docs {
			if customerIDStr != "" && (d.CustomerID == nil || fmt.Sprintf("%d", *d.CustomerID) != customerIDStr) {
				continue
			}
			if shipmentIDStr != "" && (d.ShipmentID == nil || fmt.Sprintf("%d", *d.ShipmentID) != shipmentIDStr) {
				continue
			}
			if leadIDStr != "" && (d.LeadID == nil || fmt.Sprintf("%d", *d.LeadID) != leadIDStr) {
				continue
			}
			if bookingIDStr != "" && (d.BookingID == nil || fmt.Sprintf("%d", *d.BookingID) != bookingIDStr) {
				continue
			}
			filtered = append(filtered, d)
		}
		docs = filtered
	}

	utils.Success(w, http.StatusOK, "Retrieved organization documents successfully", docs)
}

// DeleteDocument handles DELETE /api/v1/documents/{id}
func (h *Handler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.Error(w, http.StatusBadRequest, "Missing document id", "INVALID_PARAM")
		return
	}

	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		utils.Error(w, http.StatusUnauthorized, "Missing or invalid authorization user context", "UNAUTHORIZED")
		return
	}

	err := h.svc.DeleteDocument(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete document: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	actorID := userCtx.UserID
	_, _ = audit.Record(r.Context(), auditDomain.CreateAuditLogParams{
		OrgID:        userCtx.OrgID,
		ActorID:      &actorID,
		ActorRole:    userCtx.Role,
		Action:       auditDomain.ActionDelete,
		Module:       auditDomain.ModuleDocuments,
		ResourceType: "DOCUMENT",
		ResourceID:   id,
		Description:  fmt.Sprintf("Deleted document #%s", id),
		Result:       auditDomain.ResultSuccess,
	})

	utils.Success(w, http.StatusOK, "Document deleted successfully", nil)
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

	actorID := userCtx.UserID
	_, _ = audit.Record(r.Context(), auditDomain.CreateAuditLogParams{
		OrgID:        userCtx.OrgID,
		ActorID:      &actorID,
		ActorRole:    userCtx.Role,
		Action:       auditDomain.ActionCreate,
		Module:       auditDomain.ModuleDocuments,
		ResourceType: "DOCUMENT",
		ResourceID:   doc.ID,
		ResourceName: req.FileName,
		Description:  fmt.Sprintf("Uploaded %s for shipment #%d", req.FileName, shipmentID),
		Result:       auditDomain.ResultSuccess,
		Metadata: map[string]interface{}{
			"shipment_id": shipmentID,
			"doc_type":    req.DocType,
			"file_name":   req.FileName,
		},
	})

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
