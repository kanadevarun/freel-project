package integrations

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for external integrations.
type Handler struct {
	svc GatewayService
}

// NewHandler creates a new integrations HTTP handler.
func NewHandler(svc GatewayService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers integration endpoints.
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Protected management endpoints (authenticated with tenant context)
	r.Get("/status", h.GetIntegrationStatuses)
	r.Get("/configs", h.GetIntegrationConfigs)
	r.Put("/configs", h.SaveIntegrationConfig)
	r.Post("/configs", h.SaveIntegrationConfig)
	r.Get("/dead-letter", h.GetDeadLetterEvents)
	r.Get("/sms", h.ListSMSMessages)
	r.Get("/email", h.ListEmailMessages)

	// Explicit test dispatchers (to verify providers fail safely when not configured)
	r.Post("/test/sms", h.TestSendSMS)
	r.Post("/test/email", h.TestSendEmail)
	r.Get("/test/tracking", h.TestGetTracking)
	r.Post("/test/storage", h.TestStorage)
	r.Post("/test/textract", h.TestTextract)
}

// RegisterWebhookRoutes registers public webhook ingress routes (verified via HMAC).
func (h *Handler) RegisterWebhookRoutes(r chi.Router) {
	r.Post("/api/v1/integrations/webhooks/twilio", h.HandleTwilioWebhook)
	r.Post("/api/v1/integrations/webhooks/ses", h.HandleSESWebhook)
	r.Post("/api/v1/integrations/webhooks/{provider}", h.ProcessWebhook)
}

func (h *Handler) getOrgID(r *http.Request) (int64, error) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.OrgID <= 0 {
		return 0, errors.New("unauthorized: missing or invalid organization context")
	}
	return userCtx.OrgID, nil
}

// GetIntegrationStatuses returns masked integration operational status.
func (h *Handler) GetIntegrationStatuses(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	statuses, err := h.svc.GetIntegrationStatuses(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Integration statuses retrieved", statuses)
}

// GetIntegrationConfigs returns all tenant integration configs with secrets fully masked.
func (h *Handler) GetIntegrationConfigs(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	configs, err := h.svc.GetIntegrationStatuses(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Integration configurations retrieved", configs)
}

// SaveIntegrationConfig updates tenant non-secret settings or sets protected secrets.
func (h *Handler) SaveIntegrationConfig(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req SaveConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON payload", "INVALID_PAYLOAD")
		return
	}

	if err := h.svc.SaveIntegrationConfig(r.Context(), orgID, req); err != nil {
		var intErr *IntegrationError
		if errors.As(err, &intErr) {
			utils.Error(w, intErr.HTTPStatus(), intErr.Message(), intErr.Code())
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Integration configuration saved successfully", nil)
}

// TestSendSMS triggers an SMS dispatch, confirming unconfigured providers fail cleanly.
func (h *Handler) TestSendSMS(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req SMSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	resp, err := h.svc.SendSMS(r.Context(), orgID, req)
	if err != nil {
		var intErr *IntegrationError
		if errors.As(err, &intErr) {
			utils.Error(w, intErr.HTTPStatus(), intErr.Message(), intErr.Code())
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "DISPATCH_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "SMS processed", resp)
}

// TestSendEmail triggers an Email dispatch, confirming unconfigured providers fail cleanly.
func (h *Handler) TestSendEmail(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	resp, err := h.svc.SendEmail(r.Context(), orgID, req)
	if err != nil {
		var intErr *IntegrationError
		if errors.As(err, &intErr) {
			utils.Error(w, intErr.HTTPStatus(), intErr.Message(), intErr.Code())
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "DISPATCH_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Email processed", resp)
}

// TestGetTracking tests carrier tracking query.
func (h *Handler) TestGetTracking(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	carrierSCAC := r.URL.Query().Get("carrier_scac")
	trackingNumber := r.URL.Query().Get("tracking_number")
	if carrierSCAC == "" || trackingNumber == "" {
		utils.Error(w, http.StatusBadRequest, "carrier_scac and tracking_number are required", "INVALID_REQUEST")
		return
	}

	resp, err := h.svc.GetTracking(r.Context(), orgID, carrierSCAC, trackingNumber)
	if err != nil {
		var intErr *IntegrationError
		if errors.As(err, &intErr) {
			utils.Error(w, intErr.HTTPStatus(), intErr.Message(), intErr.Code())
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "TRACKING_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Tracking retrieved", resp)
}

// GetDeadLetterEvents returns dead-letter events for the tenant.
func (h *Handler) GetDeadLetterEvents(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	entries, err := h.svc.GetDeadLetterEvents(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	if entries == nil {
		entries = []DeadLetterEntry{}
	}
	utils.Success(w, http.StatusOK, "Dead-letter events retrieved", entries)
}

// ProcessWebhook handles incoming webhooks with signature verification and replay defense.
func (h *Handler) ProcessWebhook(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		utils.Error(w, http.StatusBadRequest, "Provider parameter missing", "INVALID_PROVIDER")
		return
	}

	// Resolve orgID from header or query param if provided, default to 1 or 2
	var orgID int64 = 2
	if orgStr := r.Header.Get("X-Org-ID"); orgStr != "" {
		if id, err := strconv.ParseInt(orgStr, 10, 64); err == nil && id > 0 {
			orgID = id
		}
	} else if orgStr := r.URL.Query().Get("org_id"); orgStr != "" {
		if id, err := strconv.ParseInt(orgStr, 10, 64); err == nil && id > 0 {
			orgID = id
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Failed to read webhook payload", "INVALID_BODY")
		return
	}
	defer r.Body.Close()

	res, err := h.svc.ProcessWebhook(r.Context(), strings.ToUpper(provider), orgID, r.Header, body)
	if err != nil {
		var intErr *IntegrationError
		if errors.As(err, &intErr) {
			utils.Error(w, intErr.HTTPStatus(), intErr.Message(), intErr.Code())
			return
		}
		utils.Error(w, http.StatusUnauthorized, err.Error(), "WEBHOOK_REJECTED")
		return
	}

	utils.Success(w, http.StatusOK, "Webhook verified and processed", res)
}

// ListSMSMessages returns dispatch history for outbound SMS messages in the tenant.
func (h *Handler) ListSMSMessages(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	messages, err := h.svc.ListSMSMessages(r.Context(), orgID, limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "SMS messages retrieved", messages)
}

// HandleTwilioWebhook processes Twilio delivery status callbacks.
func (h *Handler) HandleTwilioWebhook(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.HandleTwilioWebhook(r.Context(), r); err != nil {
		var intErr *IntegrationError
		if errors.As(err, &intErr) {
			utils.Error(w, intErr.HTTPStatus(), intErr.Message(), intErr.Code())
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "WEBHOOK_FAILED")
		return
	}

	// Twilio expects a 200 OK response with TwiML or empty body
	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("<Response></Response>"))
}

// ListEmailMessages returns dispatch history for outbound Email messages in the tenant.
func (h *Handler) ListEmailMessages(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	messages, err := h.svc.ListEmailMessages(r.Context(), orgID, limit)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Email messages retrieved", messages)
}

// HandleSESWebhook processes AWS SNS / SES delivery, bounce, and complaint callbacks.
func (h *Handler) HandleSESWebhook(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.HandleSESWebhook(r.Context(), r); err != nil {
		var intErr *IntegrationError
		if errors.As(err, &intErr) {
			utils.Error(w, intErr.HTTPStatus(), intErr.Message(), intErr.Code())
			return
		}
		utils.Error(w, http.StatusBadRequest, err.Error(), "WEBHOOK_FAILED")
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// TestStorage validates the active S3 configuration by attempting a test operation.
func (h *Handler) TestStorage(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req struct {
		Key string `json:"key"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Key == "" {
		req.Key = "test_connectivity.txt"
	}

	testData := []byte("LogisticsHQ storage provider validation ping")
	uploadResp, err := h.svc.UploadDocument(r.Context(), orgID, req.Key, testData, "text/plain")
	if err != nil {
		var intErr *IntegrationError
		if errors.As(err, &intErr) {
			utils.Error(w, intErr.HTTPStatus(), intErr.Message(), intErr.Code())
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "STORAGE_FAILED")
		return
	}

	// Clean up after test
	_ = h.svc.DeleteDocument(r.Context(), orgID, req.Key)

	utils.Success(w, http.StatusOK, "Storage provider test succeeded", uploadResp)
}

// TestTextract validates active Textract OCR configuration.
func (h *Handler) TestTextract(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.getOrgID(r)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
		return
	}

	var req struct {
		S3Key string `json:"s3_key"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	resp, err := h.svc.ExtractDocumentText(r.Context(), orgID, TextractRequest{
		S3Key:        req.S3Key,
		DocumentType: "PDF",
		MIMEType:     "application/pdf",
	})
	if err != nil {
		var intErr *IntegrationError
		if errors.As(err, &intErr) {
			utils.Error(w, intErr.HTTPStatus(), intErr.Message(), intErr.Code())
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "TEXTRACT_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Textract OCR test succeeded", resp)
}
