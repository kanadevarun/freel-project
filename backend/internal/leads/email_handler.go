package leads

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/freel/backend/internal/leads/spec"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/rfq"
	rfqspec "github.com/freel/backend/internal/rfq/spec"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

// EmailHandler handles incoming email ingestion, sales callbacks, and lead interaction history.
type EmailHandler struct {
	leadsBL        BusinessLogic
	rfqBL          rfq.BusinessLogic
	backendBaseURL string // e.g. "http://backend:8080" — no trailing slash
}

// NewEmailHandler creates a new EmailHandler instance.
// backendBaseURL is the resolved base URL of the Go backend (e.g. GO_BACKEND_URL env var).
func NewEmailHandler(leadsBL BusinessLogic, rfqBL rfq.BusinessLogic, backendBaseURL string) *EmailHandler {
	return &EmailHandler{
		leadsBL:        leadsBL,
		rfqBL:          rfqBL,
		backendBaseURL: backendBaseURL,
	}
}

// InboundEmailRequest represents the request body schema from the mock email client.
type InboundEmailRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	MessageID string `json:"message_id"`
	ThreadID  string `json:"thread_id"`
}

// InboundEmailWebhook handles POST /api/v1/emails/inbound (unprotected or API-key protected webhook)
func (h *EmailHandler) InboundEmailWebhook(w http.ResponseWriter, r *http.Request) {
	var req InboundEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	if req.From == "" || req.Body == "" {
		utils.Error(w, http.StatusBadRequest, "From and Body are required", "MISSING_PARAMS")
		return
	}

	orgID := int32(1) // default organization ID for local testing (Freel Global Logistics)

	inboundEmail := InboundEmail{
		From:         req.From,
		To:           req.To,
		Subject:      req.Subject,
		Body:         req.Body,
		RawEmailID:   req.MessageID,
		RFCMessageID: req.MessageID,
		MessageID:    req.MessageID,
		ThreadID:     req.ThreadID,
		Sender:       req.From,
	}

	inter, err := h.leadsBL.ProcessInboundEmail(r.Context(), orgID, inboundEmail)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to process inbound email: "+err.Error(), "SERVER_ERROR")
		return
	}

	isReply := inter.ParentInteractionID != nil

	utils.Success(w, http.StatusOK, "Inbound email queued successfully for AI parsing", map[string]interface{}{
		"interaction_id": inter.ID,
		"thread_id":      inter.ThreadID,
		"lead_id":        inter.LeadID,
		"is_reply":       isReply,
		"idempotent":     inter.IsIdempotent,
	})
}

// CreateRFQFromEmailRequest represents the schema from Python tool invocation.
type CreateRFQFromEmailRequest struct {
	OrgID       int32             `json:"org_id"`
	CustomerID  int32             `json:"customer_id"`
	Origin      string            `json:"origin"`
	Destination string            `json:"destination"`
	Incoterms   string            `json:"incoterms"`
	TargetDate  string            `json:"target_date"` // YYYY-MM-DD
	Items       []rfqspec.RFQItem `json:"items"`
}

// CreateRFQFromEmail handles POST /internal/rfqs/from-email (invoked as an internal tool by the Python AI sidecar)
func (h *EmailHandler) CreateRFQFromEmail(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req CreateRFQFromEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	if req.CustomerID <= 0 || req.OrgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "customer_id and org_id are required", "MISSING_PARAMS")
		return
	}

	// 1. Resolve lead details (req.CustomerID is the lead_id).
	leadID := req.CustomerID
	lead, err := h.leadsBL.GetLead(r.Context(), req.OrgID, leadID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve lead context: "+err.Error(), "DB_ERROR")
		return
	}

	// 2. Ensure customer record exists by updating lead status to CONVERTED.
	leadStatus := "CONVERTED"
	updateReq := spec.UpdateLeadRequest{
		OrgID:  req.OrgID,
		ID:     leadID,
		Status: &leadStatus,
	}
	_, err = h.leadsBL.UpdateLead(r.Context(), updateReq)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to ensure customer for lead: "+err.Error(), "DB_ERROR")
		return
	}

	// 3. Resolve customer ID.
	customerID, err := h.leadsBL.GetCustomerIDByCompanyName(r.Context(), req.OrgID, lead.CompanyName)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to resolve customer ID: "+err.Error(), "DB_ERROR")
		return
	}

	var parsedDate time.Time
	if req.TargetDate != "" {
		parsedDate, _ = time.Parse("2006-01-02", req.TargetDate)
	} else {
		parsedDate = time.Now().AddDate(0, 1, 0) // default 1 month out
	}

	rfqReq := rfqspec.CreateRFQRequest{
		CustomerID:  customerID,
		Origin:      &req.Origin,
		Destination: &req.Destination,
		Incoterms:   &req.Incoterms,
		TargetDate:  &parsedDate,
		Items:       req.Items,
	}
	rfqReq.OrgID = req.OrgID
	leadID64 := int64(leadID)
	rfqReq.LeadID = &leadID64

	// If an RFQ already exists for this lead, return existing RFQ
	existingRFQ, err := h.rfqBL.GetRFQByLeadID(r.Context(), req.OrgID, leadID64)
	if err == nil && existingRFQ != nil {
		utils.Success(w, http.StatusOK, "RFQ already exists for lead", map[string]interface{}{
			"rfq_id":     existingRFQ.ID,
			"rfq_number": existingRFQ.RFQNumber,
		})
		return
	}

	newRFQ, err := h.rfqBL.CreateRFQ(r.Context(), rfqReq)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create RFQ: "+err.Error(), "DB_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "RFQ created from email", map[string]interface{}{
		"rfq_id":     newRFQ.ID,
		"rfq_number": newRFQ.RFQNumber,
	})
}

// SalesCallbackRequest represents the payload schema received from SalesAgent.
type SalesCallbackRequest struct {
	InteractionID int64  `json:"interaction_id"`
	OrgID         int64  `json:"org_id"`
	LeadID        int64  `json:"lead_id"`
	Sentiment     string `json:"sentiment"`
	Intent        string `json:"intent"`
	Confidence    int    `json:"confidence"`
	LinkedRFQID   *int64 `json:"linked_rfq_id,omitempty"`
	Summary       string `json:"summary"`
	DraftedReply  string `json:"drafted_reply,omitempty"`
	// Cumulative structured context from this conversation turn — persisted for next reply
	PartialRFQContext map[string]interface{} `json:"partial_rfq_context,omitempty"`
}

// SalesCallback handles POST /internal/sales/callback (invoked by Python AI sidecar upon graph execution finish)
func (h *EmailHandler) SalesCallback(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticate(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req SalesCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	if req.InteractionID <= 0 || req.OrgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "interaction_id and org_id are required", "MISSING_PARAMS")
		return
	}

	err := h.leadsBL.UpdateInteractionAI(r.Context(), req.OrgID, req.InteractionID, req.Intent, req.Sentiment, req.Confidence, req.LinkedRFQID, req.Summary, req.DraftedReply)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update interaction AI: "+err.Error(), "DB_ERROR")
		return
	}

	// Always persist cumulative partial RFQ context if provided by the callback
	if len(req.PartialRFQContext) > 0 {
		var prevContext map[string]interface{}
		interactions, listErr := h.leadsBL.ListInteractions(r.Context(), int32(req.OrgID), int32(req.LeadID))
		if listErr == nil {
			for i := len(interactions) - 1; i >= 0; i-- {
				inter := interactions[i]
				if inter.ID == req.InteractionID {
					continue
				}
				inter.UnmarshalPartialRFQContext()
				if len(inter.PartialRFQContext) > 0 {
					prevContext = inter.PartialRFQContext
					break
				}
			}
		}

		mergedContext := MergeRFQContext(prevContext, req.PartialRFQContext)

		if ctxErr := h.leadsBL.UpdateInteractionContext(r.Context(), req.OrgID, req.InteractionID, mergedContext); ctxErr != nil {
			log.Printf("[SalesCallback] Warning: failed to persist partial RFQ context for interaction %d: %v", req.InteractionID, ctxErr)
		}
		
		// Detailed structured logging for RFQ context merge validation
		log.Printf("[RFQ Context] Existing conversation found for Lead ID: %d", req.LeadID)
		log.Printf("[RFQ Context] Previous fields: %v", prevContext)
		log.Printf("[RFQ Context] New fields extracted: %v", req.PartialRFQContext)
		log.Printf("[RFQ Context] Context merged successfully.")
		
		missing := GetMissingRFQFields(mergedContext)
		if len(missing) > 0 {
			log.Printf("[RFQ Context] Missing fields: %s", strings.Join(missing, ", "))
		} else {
			log.Printf("[RFQ Context] Missing fields: none")
			log.Printf("[RFQ Conversion] Lead %d is ready for RFQ conversion.", req.LeadID)
			if req.LinkedRFQID != nil && *req.LinkedRFQID > 0 {
				log.Printf("[RFQ Conversion] RFQ created successfully: %d", *req.LinkedRFQID)
			}
		}
	}

	// Update Lead status based on intent or linked RFQ
	if req.LinkedRFQID != nil && *req.LinkedRFQID > 0 {
		updateReq := spec.UpdateLeadRequest{
			OrgID:  int32(req.OrgID),
			ID:     int32(req.LeadID),
			Status: strPtr("CONVERTED"),
		}
		_, _ = h.leadsBL.UpdateLead(r.Context(), updateReq)
	}

	// Trigger outbound sending if clarification was generated for incomplete RFQ
	if req.Intent == "NOT_LOGISTICS" || req.Intent == "PERSONAL" || req.Intent == "SPAM" || req.Intent == "OTHER" || req.Intent == "IRRELEVANT" {
		log.Printf("[SalesCallback] Intent is %s. Suppressing automated clarification reply.", req.Intent)
	} else if req.Intent == "RFQ_REQUEST_INCOMPLETE" && req.DraftedReply != "" {
		go func() {
			// Use context.Background() since the callback request context will terminate
			err := h.leadsBL.SendClarificationEmail(context.Background(), req.OrgID, req.InteractionID, req.DraftedReply, req.Summary)
			if err != nil {
				log.Printf("[SalesCallback] Failed to send outbound email: %v", err)
			}
		}()
	}

	utils.Success(w, http.StatusOK, "Sales callback processed successfully", nil)
}

// GetInteractions handles GET /api/v1/leads/{id}/interactions
func (h *EmailHandler) GetInteractions(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	leadID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid lead id", "INVALID_PARAM")
		return
	}

	// Fetch org_id from Cognito user context first, with query param fallback for local tests/CLI
	orgID := int32(1)
	if userCtx, ok := middleware.GetUserContext(r.Context()); ok {
		orgID = int32(userCtx.OrgID)
	} else {
		orgIDStr := r.URL.Query().Get("org_id")
		if orgIDStr != "" {
			if parsed, err := strconv.Atoi(orgIDStr); err == nil {
				orgID = int32(parsed)
			}
		}
	}

	list, err := h.leadsBL.ListInteractions(r.Context(), orgID, int32(leadID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list interactions: "+err.Error(), "DB_ERROR")
		return
	}

	// Support optional chronological ASC sorting (default is DESC from database)
	order := r.URL.Query().Get("order")
	if strings.ToLower(order) == "asc" {
		sort.Slice(list, func(i, j int) bool {
			return list[i].CreatedAt.Before(list[j].CreatedAt)
		})
	}

	utils.Success(w, http.StatusOK, "Interactions retrieved", list)
}

// CreateInteraction handles POST /api/v1/leads/{id}/interactions (manual logging of call notes, etc.)
func (h *EmailHandler) CreateInteraction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	leadID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid lead id", "INVALID_PARAM")
		return
	}

	var req LeadInteraction
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	orgIDStr := r.URL.Query().Get("org_id")
	if orgIDStr == "" {
		orgIDStr = "1"
	}
	orgID, _ := strconv.Atoi(orgIDStr)

	req.LeadID = leadID
	req.Direction = "OUTBOUND" // manual entries from FFs are typically outbound notes
	req.CreatedAt = time.Now()

	err = h.leadsBL.LogInteraction(r.Context(), int32(orgID), &req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to log interaction: "+err.Error(), "DB_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Lead interaction logged", req)
}

func (h *EmailHandler) authenticate(r *http.Request) error {
	token := r.Header.Get("X-LogisticsHQ-Service-Key")
	if token == "" {
		token = r.URL.Query().Get("service_key")
	}

	expectedToken := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if expectedToken == "" {
		if os.Getenv("APP_ENV") == "production" {
			return fmt.Errorf("Configuration error: INTERNAL_SERVICE_TOKEN must be specified in production environments")
		}
		expectedToken = "internal-service-key-logisticshq"
	}

	if token != expectedToken {
		return fmt.Errorf("Unauthorized access: Invalid service key token")
	}
	return nil
}

// RetryClarificationEmail handles POST /api/v1/leads/{id}/interactions/{interaction_id}/retry
func (h *EmailHandler) RetryClarificationEmail(w http.ResponseWriter, r *http.Request) {
	// Let's use internal service token or session token?
	// Wait, since this is invoked from the frontend (FF user), it's authenticated using session token!
	// So we don't need INTERNAL_SERVICE_TOKEN auth, we should use standard session or skip internal authenticate call!
	// Let's double check other user endpoints like GetInteractions. They do NOT call h.authenticate(r), because session verification is done by standard router middleware!
	// Yes! Router middleware handles user authentication for /api/v1 prefix routes automatically!
	// So RetryClarificationEmail does NOT need h.authenticate(r)!
	idStr := chi.URLParam(r, "id")
	_, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid lead ID", "INVALID_PARAMS")
		return
	}

	interIDStr := chi.URLParam(r, "interaction_id")
	interactionID, err := strconv.ParseInt(interIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid interaction ID", "INVALID_PARAMS")
		return
	}

	orgIDStr := r.URL.Query().Get("org_id")
	if orgIDStr == "" {
		orgIDStr = "1"
	}
	orgIDVal, _ := strconv.Atoi(orgIDStr)
	orgID := int32(orgIDVal)

	// Fetch parent interaction to retrieve drafted reply and summary
	parentInter, err := h.leadsBL.GetInteractionByID(r.Context(), orgID, interactionID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "Interaction not found: "+err.Error(), "NOT_FOUND")
		return
	}

	if parentInter.DraftedReply == "" {
		utils.Error(w, http.StatusBadRequest, "No drafted clarification email available for this interaction", "INVALID_ACTION")
		return
	}

	// Trigger sending synchronously (so we can return immediate success/failure back to the retry UI!)
	err = h.leadsBL.SendClarificationEmail(r.Context(), int64(orgID), interactionID, parentInter.DraftedReply, parentInter.AISummary)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to send email: "+err.Error(), "SEND_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Clarification email resent successfully", nil)
}

// ReplyToInteractionRequest represents the payload for sending a manual reply.
type ReplyToInteractionRequest struct {
	From    string `json:"from"`
	To      string `json:"to"`
	CC      string `json:"cc"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type EmailDraftResponse struct {
	ID                  int64  `json:"id"`
	OrgID               int64  `json:"org_id"`
	LeadID              int64  `json:"lead_id"`
	ParentInteractionID int64  `json:"parent_interaction_id"`
	MailboxID           *int64 `json:"mailbox_id"`
	From                string `json:"from"`
	To                  string `json:"to"`
	CC                  string `json:"cc"`
	Subject             string `json:"subject"`
	Body                string `json:"body"`
}

type SaveEmailDraftRequest struct {
	MailboxID *int64 `json:"mailbox_id"`
	From      string `json:"from"`
	To        string `json:"to"`
	CC        string `json:"cc"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

// ReplyToInteraction handles POST /api/v1/leads/{id}/interactions/{interaction_id}/reply
// It manages manual email replies initiated by sales operators within LogisticsHQ.
//
// URL Parameters:
//   - id: The numeric Lead ID (e.g. "182").
//   - interaction_id: The ID of the inbound email interaction being replied to (e.g. "701").
//
// Payload Example:
//
//	{
//	  "to": "customer@example.com",
//	  "cc": "manager@logistics.com",
//	  "subject": "Re: Freight quote request",
//	  "body": "Hi Jane, we can ship your cargo next week."
//	}
func (h *EmailHandler) ReplyToInteraction(w http.ResponseWriter, r *http.Request) {
	// Parse the lead ID from the URL path.
	// Example: In /api/v1/leads/182/interactions/701/reply, this extracts "182" as the leadID.
	idStr := chi.URLParam(r, "id")
	leadID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid lead ID", "INVALID_PARAMS")
		return
	}

	// Parse the parent interaction ID from the URL path.
	// Example: In /api/v1/leads/182/interactions/701/reply, this extracts "701" as parent interaction ID.
	interIDStr := chi.URLParam(r, "interaction_id")
	interactionID, err := strconv.ParseInt(interIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid interaction ID", "INVALID_PARAMS")
		return
	}

	// Decode the JSON request body containing recipient, subject, and message body.
	var req ReplyToInteractionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	// Validate that mandatory parameters are supplied.
	if req.To == "" || req.Subject == "" || req.Body == "" {
		utils.Error(w, http.StatusBadRequest, "To, Subject, and Body are required", "MISSING_PARAMS")
		return
	}

	// Resolve the organization context to ensure organization isolation.
	// 1. First, attempt to retrieve the logged-in user context populated by the Cognito auth middleware.
	// 2. Fall back to reading the "org_id" query parameter (useful for local CLI testing or mock scripts).
	orgID := int32(1)
	if userCtx, ok := middleware.GetUserContext(r.Context()); ok {
		orgID = int32(userCtx.OrgID)
	} else {
		orgIDStr := r.URL.Query().Get("org_id")
		if orgIDStr != "" {
			if parsed, err := strconv.Atoi(orgIDStr); err == nil {
				orgID = int32(parsed)
			}
		}
	}

	// Delegate the core logic of selecting a mailbox, threading headers, logging PENDING state,
	// and calling the Gmail API to the business logic layer.
	outboundInter, err := h.leadsBL.ReplyToInteraction(
		r.Context(),
		int64(orgID),
		leadID,
		interactionID,
		req.From,
		req.To,
		req.CC,
		req.Subject,
		req.Body,
	)

	if err != nil {
		// If Gmail sending fails but we successfully logged a PENDING database record,
		// we updated its status to FAILED. We return this failed record to the frontend (with HTTP 200 OK)
		// so that the UI can render a "Failed to send" card and offer the operator a retry action
		// without losing their typed text.
		if outboundInter != nil {
			utils.Success(w, http.StatusOK, "Manual reply attempted but failed: "+err.Error(), outboundInter)
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to reply: "+err.Error(), "SEND_FAILED")
		return
	}

	// Return the successfully sent outbound interaction record to the frontend client.
	utils.Success(w, http.StatusOK, "Manual reply sent successfully", outboundInter)
}

// RetryEmailInteraction handles POST /api/v1/leads/{id}/interactions/{interaction_id}/retry
func (h *EmailHandler) RetryEmailInteraction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	leadID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid lead ID", "INVALID_PARAMS")
		return
	}

	interIDStr := chi.URLParam(r, "interaction_id")
	interactionID, err := strconv.ParseInt(interIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid interaction ID", "INVALID_PARAMS")
		return
	}

	orgID := int32(1)
	if userCtx, ok := middleware.GetUserContext(r.Context()); ok {
		orgID = int32(userCtx.OrgID)
	} else {
		orgIDStr := r.URL.Query().Get("org_id")
		if orgIDStr != "" {
			if parsed, err := strconv.Atoi(orgIDStr); err == nil {
				orgID = int32(parsed)
			}
		}
	}

	updatedInter, err := h.leadsBL.RetryEmailInteraction(
		r.Context(),
		int64(orgID),
		leadID,
		interactionID,
	)

	if err != nil {
		if updatedInter != nil {
			utils.Success(w, http.StatusOK, "Retry failed: "+err.Error(), updatedInter)
			return
		}
		utils.Error(w, http.StatusInternalServerError, "Failed to retry email interaction: "+err.Error(), "RETRY_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Manual reply sent successfully", updatedInter)
}

// GetDraft handles GET /api/v1/leads/{id}/interactions/{interaction_id}/draft
func (h *EmailHandler) GetDraft(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	leadID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid lead ID", "INVALID_PARAMS")
		return
	}

	interIDStr := chi.URLParam(r, "interaction_id")
	interactionID, err := strconv.ParseInt(interIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid interaction ID", "INVALID_PARAMS")
		return
	}

	orgID := int32(1)
	if userCtx, ok := middleware.GetUserContext(r.Context()); ok {
		orgID = int32(userCtx.OrgID)
	} else {
		orgIDStr := r.URL.Query().Get("org_id")
		if orgIDStr != "" {
			if parsed, err := strconv.Atoi(orgIDStr); err == nil {
				orgID = int32(parsed)
			}
		}
	}

	draft, err := h.leadsBL.GetDraft(r.Context(), int64(orgID), leadID, interactionID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get draft: "+err.Error(), "DRAFT_GET_FAILED")
		return
	}

	if draft == nil {
		utils.Success(w, http.StatusOK, "No draft found", nil)
		return
	}

	resp := EmailDraftResponse{
		ID:                  draft.ID,
		OrgID:               draft.OrgID,
		LeadID:              draft.LeadID,
		ParentInteractionID: draft.ParentInteractionID,
		MailboxID:           draft.MailboxID,
		From:                draft.From,
		To:                  draft.Recipients,
		CC:                  draft.CCRecipients,
		Subject:             draft.Subject,
		Body:                draft.Content,
	}
	utils.Success(w, http.StatusOK, "Draft retrieved successfully", resp)
}

// SaveDraft handles PUT /api/v1/leads/{id}/interactions/{interaction_id}/draft
func (h *EmailHandler) SaveDraft(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	leadID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid lead ID", "INVALID_PARAMS")
		return
	}

	interIDStr := chi.URLParam(r, "interaction_id")
	interactionID, err := strconv.ParseInt(interIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid interaction ID", "INVALID_PARAMS")
		return
	}

	var req SaveEmailDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	orgID := int32(1)
	if userCtx, ok := middleware.GetUserContext(r.Context()); ok {
		orgID = int32(userCtx.OrgID)
	} else {
		orgIDStr := r.URL.Query().Get("org_id")
		if orgIDStr != "" {
			if parsed, err := strconv.Atoi(orgIDStr); err == nil {
				orgID = int32(parsed)
			}
		}
	}

	draft := &LeadEmailDraft{
		OrgID:               int64(orgID),
		LeadID:              leadID,
		ParentInteractionID: interactionID,
		MailboxID:           req.MailboxID,
		From:                req.From,
		Recipients:          req.To,
		CCRecipients:        req.CC,
		Subject:             req.Subject,
		Content:             req.Body,
	}

	err = h.leadsBL.SaveDraft(r.Context(), draft)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to save draft: "+err.Error(), "DRAFT_SAVE_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Draft saved successfully", nil)
}

// DeleteDraft handles DELETE /api/v1/leads/{id}/interactions/{interaction_id}/draft
func (h *EmailHandler) DeleteDraft(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	leadID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid lead ID", "INVALID_PARAMS")
		return
	}

	interIDStr := chi.URLParam(r, "interaction_id")
	interactionID, err := strconv.ParseInt(interIDStr, 10, 64)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid interaction ID", "INVALID_PARAMS")
		return
	}

	orgID := int32(1)
	if userCtx, ok := middleware.GetUserContext(r.Context()); ok {
		orgID = int32(userCtx.OrgID)
	} else {
		orgIDStr := r.URL.Query().Get("org_id")
		if orgIDStr != "" {
			if parsed, err := strconv.Atoi(orgIDStr); err == nil {
				orgID = int32(parsed)
			}
		}
	}

	err = h.leadsBL.DeleteDraft(r.Context(), int64(orgID), leadID, interactionID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete draft: "+err.Error(), "DRAFT_DELETE_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Draft deleted successfully", nil)
}

func GetMissingRFQFields(ctx map[string]interface{}) []string {
	var missing []string
	required := map[string]string{
		"origin_port":       "origin",
		"destination_port":  "destination",
		"incoterms":         "incoterms",
		"cargo_description": "cargo_description",
		"cargo_weight":      "cargo_weight",
		"cargo_volume":      "cargo_volume",
		"target_date":       "target_date",
	}
	for key, label := range required {
		val, exists := ctx[key]
		if !exists || val == nil {
			missing = append(missing, label)
		} else {
			switch v := val.(type) {
			case string:
				if v == "" {
					missing = append(missing, label)
				}
			case float64:
				if v <= 0 {
					missing = append(missing, label)
				}
			case int:
				if v <= 0 {
					missing = append(missing, label)
				}
			}
		}
	}
	return missing
}

func MergeRFQContext(prev, new map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	
	// Copy previous values
	for k, v := range prev {
		if v != nil {
			merged[k] = v
		}
	}
	
	// Overwrite/merge new values
	for k, v := range new {
		if v == nil {
			continue
		}
		
		// Skip empty strings
		if str, ok := v.(string); ok && str == "" {
			continue
		}
		
		// Skip non-positive float/int values
		if f, ok := v.(float64); ok && f <= 0 {
			continue
		}
		if i, ok := v.(int); ok && i <= 0 {
			continue
		}
		
		merged[k] = v
	}
	
	return merged
}

