package contracts

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
)

// Handler represents the HTTP handler for contract-related requests.
// It wraps the Service interface which encapsulates the core business logic.
type Handler struct {
	svc Service
}

// NewHandler acts as a constructor, returning an initialized Handler pointer.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Upload handles POST /api/v1/contracts/upload
//
// What it does:
//   Allows a freight forwarder operator to upload a new carrier contract document (PDF or Excel).
//   It stores the file locally or in S3, creates a queued record in the database,
//   and triggers the AI Sidecar parser in the background.
//
// How to use it (Example):
//   Method: POST
//   Headers:
//     Authorization: Bearer <JWT_TOKEN>
//     Content-Type: multipart/form-data
//   Form Fields:
//     file: [binary payload of maersk_contract.pdf]
//     carrier_scac: "MAEU" (Optional SCAC code)
//
// Responses:
//   - 202 Accepted: Upload successful. The file is being processed asynchronously.
//     Payload: { "status": "success", "message": "...", "data": { "id": "uuid-123", "status": "QUEUED" } }
//   - 400 Bad Request: Missing file or file exceeds 20MB.
//   - 401 Unauthorized: Cognito token is invalid or missing.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	// Retrieve the authenticated user context injected by the middleware.
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		// Respond with a 401 Unauthorized status if no user context is present.
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	// Limit the request body reading buffer to 20MB to prevent denial-of-service file sizes.
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		utils.Error(w, http.StatusBadRequest, "File too large (max 20MB)", "FILE_TOO_LARGE")
		return
	}

	// Retrieve the file stream and multipart headers using the "file" key.
	file, header, err := r.FormFile("file")
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Missing 'file' parameter", "MISSING_FILE")
		return
	}
	// Ensure the file is closed after this handler returns to prevent descriptor leaks.
	defer file.Close()

	// Extract the optional carrier SCAC code from the form values.
	carrierSCAC := r.FormValue("carrier_scac")
	var scacPtr *string
	// Sanitize and check if a valid carrier SCAC value was provided.
	if scac := strings.TrimSpace(carrierSCAC); scac != "" {
		upper := strings.ToUpper(scac)
		scacPtr = &upper
	}

	// Delegate file storage and AI bridge trigger orchestration to the service layer.
	doc, err := h.svc.UploadContract(r.Context(), userCtx.OrgID, userCtx.UserID, header.Filename, file, header.Size, scacPtr)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "UPLOAD_FAILED")
		return
	}

	// Respond with a 202 Accepted status indicating processing has started asynchronously.
	utils.Success(w, http.StatusAccepted, "Contract upload accepted and queued for processing", doc)
}

// List handles GET /api/v1/contracts
//
// What it does:
//   Fetches a history list of all uploaded contract documents belonging to the user's organization.
//   You can filter the list by status using query parameters.
//
// How to use it (Example):
//   Method: GET
//   Request URL: /api/v1/contracts?status=PENDING_REVIEW
//   Headers:
//     Authorization: Bearer <JWT_TOKEN>
//
// Responses:
//   - 200 OK: Returns a JSON array of contract records.
//     Payload: { "status": "success", "data": [ { "id": "uuid-123", "file_name": "...", "status": "PENDING_REVIEW" } ] }
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	// Authenticate the caller using the user context.
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	// Read optional query parameter to filter documents by their processing status.
	var statusPtr *ProcessingStatus
	if statusParam := r.URL.Query().Get("status"); statusParam != "" {
		status := ProcessingStatus(strings.ToUpper(strings.TrimSpace(statusParam)))
		statusPtr = &status
	}

	// Query the service for the filtered list of contract documents.
	docs, err := h.svc.ListDocuments(r.Context(), userCtx.OrgID, statusPtr)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "LIST_FAILED")
		return
	}

	// Return the resulting document list with a 200 OK status.
	utils.Success(w, http.StatusOK, "Contracts retrieved successfully", docs)
}

// Get handles GET /api/v1/contracts/{id}
//
// What it does:
//   Retrieves specific details, metadata, and full processing log entries for a single contract document.
//
// How to use it (Example):
//   Method: GET
//   Request URL: /api/v1/contracts/3ae5c3ab-51a2-4a0b-9cc3-1a224fbc11e3
//   Headers:
//     Authorization: Bearer <JWT_TOKEN>
//
// Responses:
//   - 200 OK: Returns document metadata and logs showing OCR_PROCESSING, CLASSIFICATION, etc.
//     Payload: { "status": "success", "data": { "id": "...", "file_name": "...", "processing_log": [...] } }
//   - 404 Not Found: The contract document does not exist or belongs to another organization.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	// Authenticate the caller.
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	// Extract the document ID from the URL path suffix.
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	id := ""
	if len(parts) > 0 {
		id = parts[len(parts)-1]
	}
	if id == "" {
		utils.Error(w, http.StatusBadRequest, "contract id is required", "MISSING_ID")
		return
	}

	// Retrieve the document details from the database via the service layer.
	doc, err := h.svc.GetDocument(r.Context(), userCtx.OrgID, id)
	if err != nil {
		// Return 404 if the row does not exist.
		if strings.Contains(err.Error(), "no rows") {
			utils.Error(w, http.StatusNotFound, "Contract not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "GET_FAILED")
		return
	}

	// Return the retrieved document details.
	utils.Success(w, http.StatusOK, "Contract retrieved successfully", doc)
}

// Reprocess handles POST /api/v1/contracts/{id}/reprocess
//
// What it does:
//   Forces the system to re-trigger the Python AI extraction sidecar for an existing contract.
//   Useful if the AI sidecar was temporarily offline or if prompt templates were updated.
//
// How to use it (Example):
//   Method: POST
//   Request URL: /api/v1/contracts/3ae5c3ab-51a2-4a0b-9cc3-1a224fbc11e3/reprocess
//   Headers:
//     Authorization: Bearer <JWT_TOKEN>
//
// Responses:
//   - 200 OK: Reprocessing triggered.
//     Payload: { "status": "success", "message": "Reprocessing triggered successfully" }
func (h *Handler) Reprocess(w http.ResponseWriter, r *http.Request) {
	// Verify user authentication.
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	// Parse the contract ID from the URL path, removing the trailing '/reprocess'.
	path := strings.TrimSuffix(r.URL.Path, "/reprocess")
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	id := ""
	if len(parts) > 0 {
		id = parts[len(parts)-1]
	}
	if id == "" {
		utils.Error(w, http.StatusBadRequest, "contract id is required", "MISSING_ID")
		return
	}

	// Call the service to update status and trigger the Python sidecar.
	err := h.svc.TriggerReprocessing(r.Context(), userCtx.OrgID, id)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "REPROCESS_FAILED")
		return
	}

	// Respond with success indication.
	utils.Success(w, http.StatusOK, "Reprocessing triggered successfully", nil)
}

// Callback handles POST /internal/contracts/callback (AI sidecar endpoint)
//
// What it does:
//   Internal webhook called by the Python AI sidecar upon finishing its multi-agent extraction pipeline.
//   It ingests the clean confirmed rates directly into the rates table, and flags anomalies
//   into the human rate review queue.
//
// How to use it (Example):
//   Method: POST
//   Request Body:
//     {
//       "document_id": "3ae5c3ab-51a2-4a0b-9cc3-1a224fbc11e3",
//       "org_id": 5,
//       "status": "COMPLETED",
//       "confirmed_rates": [ { "origin_port": "INNSA", "destination_port": "DEHAM", "ocean_freight": 2800 } ],
//       "flagged_items": [ { "confidence_score": 60, "review_flags": ["PRICE_ANOMALY"], "extracted_data": { ... } } ],
//       "processing_log": [ { "step": "CLASSIFICATION", "message": "Matched Maersk" } ],
//       "ai_summary": "Summary of Maersk Contract SC-99201"
//     }
//
// Responses:
//   - 200 OK: Callback processed successfully.
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate the caller. Verify the shared service key token in the request header.
	//
	// Simple meaning:
	//   We verify that the incoming POST request actually comes from our secure
	//   AI sidecar service. We look for a token inside the 'X-LogisticsHQ-Service-Key' header.
	//
	// Example:
	//   Header 'X-LogisticsHQ-Service-Key' must match env INTERNAL_SERVICE_TOKEN.
	token := r.Header.Get("X-LogisticsHQ-Service-Key")
	if token == "" {
		token = r.URL.Query().Get("service_key")
	}

	expectedToken := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if expectedToken == "" {
		if os.Getenv("APP_ENV") == "production" {
			http.Error(w, "Configuration error: INTERNAL_SERVICE_TOKEN must be specified in production environments", http.StatusInternalServerError)
			return
		}
		expectedToken = "internal-service-key-logisticshq"
	}

	if token != expectedToken {
		http.Error(w, "Unauthorized access: Invalid service key token", http.StatusUnauthorized)
		return
	}

	var callback AIProcessingCallback
	// Parse the JSON callback payload.
	if err := json.NewDecoder(r.Body).Decode(&callback); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	// Validate required identifiers in the callback.
	if callback.DocumentID == "" || callback.OrgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "document_id and org_id are required", "MISSING_PARAMS")
		return
	}

	// Process the extracted rates, queue reviews, and persist results.
	err := h.svc.HandleAICallback(r.Context(), callback)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "CALLBACK_FAILED")
		return
	}

	// Confirm successful processing to the sidecar.
	utils.Success(w, http.StatusOK, "Callback processed successfully", nil)
}

// ListReview handles GET /api/v1/contracts/review
//
// What it does:
//   Fetches items from the human verification review queue (e.g. flagged anomalies)
//   for operators to examine, edit, and approve/reject.
//
// How to use it (Example):
//   Method: GET
//   Request URL: /api/v1/contracts/review?status=PENDING
//   Headers:
//     Authorization: Bearer <JWT_TOKEN>
//
// Responses:
//   - 200 OK: Returns JSON array of flagged items.
//     Payload: { "status": "success", "data": [ { "id": "review-item-uuid", "confidence_score": 62, "extracted_data": { ... } } ] }
func (h *Handler) ListReview(w http.ResponseWriter, r *http.Request) {
	// Authenticate user context.
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	// Read optional query parameter to filter review items by status.
	var statusPtr *ReviewStatus
	if statusParam := r.URL.Query().Get("status"); statusParam != "" {
		status := ReviewStatus(strings.ToUpper(strings.TrimSpace(statusParam)))
		statusPtr = &status
	}

	// Get matching review items from the repository.
	items, err := h.svc.ListReviewItems(r.Context(), userCtx.OrgID, statusPtr)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "LIST_REVIEW_FAILED")
		return
	}

	// Return retrieved queue items.
	utils.Success(w, http.StatusOK, "Review items retrieved successfully", items)
}

// ApproveReview handles PUT /api/v1/contracts/review/{id}/approve
//
// What it does:
//   Approves a flagged rate item, optionally saving manual corrections submitted by the human operator.
//   It normalizes the ports and saves the rate into the live rates catalog (`rate_entries`).
//
// How to use it (Example):
//   Method: PUT
//   Request URL: /api/v1/contracts/review/9ba8f7ea-1234-4567-890a-bcdef1234567/approve
//   Headers:
//     Authorization: Bearer <JWT_TOKEN>
//     Content-Type: application/json
//   Request Body (Optional, if operator corrected pricing fields):
//     {
//       "corrected_data": {
//         "origin_port": "INNSA",
//         "destination_port": "DEHAM",
//         "carrier_scac": "MAEU",
//         "carrier_name": "Maersk",
//         "ocean_freight": 2800.0,
//         "total_buy_price": 3200.0,
//         "valid_from": "2026-09-01T00:00:00Z",
//         "valid_until": "2026-12-31T23:59:59Z"
//       },
//       "notes": "Corrected ocean freight based on surcharge table."
//     }
//
// Responses:
//   - 200 OK: Approved and ingested.
//     Payload: { "status": "success", "message": "Review item approved and rate ingested successfully" }
func (h *Handler) ApproveReview(w http.ResponseWriter, r *http.Request) {
	// Authenticate the user context.
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	// Extract the review item ID from the path, removing '/approve'.
	path := strings.TrimSuffix(r.URL.Path, "/approve")
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	id := ""
	if len(parts) > 0 {
		id = parts[len(parts)-1]
	}
	if id == "" {
		utils.Error(w, http.StatusBadRequest, "review item id is required", "MISSING_ID")
		return
	}

	// Parse optional correction data and reviewer notes.
	var body struct {
		CorrectedData interface{} `json:"corrected_data"`
		Notes         string      `json:"notes"`
	}
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	// Serialize corrected rate payload back to raw JSON bytes.
	var correctedBytes []byte
	if body.CorrectedData != nil {
		var err error
		correctedBytes, err = json.Marshal(body.CorrectedData)
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "Invalid corrected data JSON", "INVALID_PAYLOAD")
			return
		}
	}

	// Process approval, update review entry status, and ingest finalized rate.
	err := h.svc.ApproveReviewItem(r.Context(), userCtx.OrgID, id, userCtx.UserID, correctedBytes, body.Notes)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "APPROVE_FAILED")
		return
	}

	// Return success response.
	utils.Success(w, http.StatusOK, "Review item approved and rate ingested successfully", nil)
}

// RejectReview handles PUT /api/v1/contracts/review/{id}/reject
//
// What it does:
//   Rejects a flagged rate extraction draft, marking it as REJECTED.
//   The rate will NOT be ingested into the live catalog.
//
// How to use it (Example):
//   Method: PUT
//   Request URL: /api/v1/contracts/review/9ba8f7ea-1234-4567-890a-bcdef1234567/reject
//   Headers:
//     Authorization: Bearer <JWT_TOKEN>
//     Content-Type: application/json
//   Request Body (Optional):
//     {
//       "notes": "This table is from an expired contract ref, do not ingest."
//     }
//
// Responses:
//   - 200 OK: Rejection logged successfully.
//     Payload: { "status": "success", "message": "Review item rejected successfully" }
func (h *Handler) RejectReview(w http.ResponseWriter, r *http.Request) {
	// Authenticate user context.
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	// Extract the review item ID from the path, removing '/reject'.
	path := strings.TrimSuffix(r.URL.Path, "/reject")
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	id := ""
	if len(parts) > 0 {
		id = parts[len(parts)-1]
	}
	if id == "" {
		utils.Error(w, http.StatusBadRequest, "review item id is required", "MISSING_ID")
		return
	}

	// Parse optional reviewer rejection notes.
	var body struct {
		Notes string `json:"notes"`
	}
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	// Update the review entry status to 'REJECTED' in the database.
	err := h.svc.RejectReviewItem(r.Context(), userCtx.OrgID, id, userCtx.UserID, body.Notes)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "REJECT_FAILED")
		return
	}

	// Return success response.
	utils.Success(w, http.StatusOK, "Review item rejected successfully", nil)
}



