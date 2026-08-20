package leads

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/freel/backend/internal/leads/spec"
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

	// 1. Resolve lead from 'From' email address.
	lead, err := h.leadsBL.GetLeadByEmail(r.Context(), orgID, req.From)
	if err != nil {
		// Lead does not exist; create a new lead automatically.
		contactName := req.From
		companyName := "Inbound Lead (" + req.From + ")"
		createReq := spec.CreateLeadRequest{
			OrgID:       orgID,
			CompanyName: companyName,
			ContactName: strPtr(contactName),
			Email:       strPtr(req.From),
			Source:      strPtr("EMAIL"),
		}
		lead, err = h.leadsBL.CreateLead(r.Context(), createReq)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to auto-create lead: "+err.Error(), "DB_ERROR")
			return
		}
	}

	// 2. Resolve/Generate ThreadID for conversation tracking.
	threadID := req.ThreadID
	if threadID == "" {
		if req.MessageID != "" {
			threadID = req.MessageID
		} else {
			threadID = fmt.Sprintf("thread_%d_%d", lead.ID, time.Now().UnixNano())
		}
	}

	// 3. Detect if this email is a reply to a prior INCOMPLETE RFQ conversation.
	//    We check for prior interactions on this thread. If any are RFQ_REQUEST_INCOMPLETE,
	//    the customer is providing the missing information — pass the prior extracted context
	//    as structured JSON so the AI agent can merge it without re-reading raw emails.
	var parentInteractionID *int64
	var priorRFQContext map[string]interface{}
	isReply := false

	if existingInteractions, err := h.leadsBL.FindByThreadID(r.Context(), orgID, threadID); err == nil {
		// Walk backwards to find the most recent incomplete interaction
		for i := len(existingInteractions) - 1; i >= 0; i-- {
			prev := existingInteractions[i]
			if prev.Intent == "RFQ_REQUEST_INCOMPLETE" && len(prev.PartialRFQContextRaw) > 0 {
				isReply = true
				parentID := prev.ID
				parentInteractionID = &parentID
				prev.UnmarshalPartialRFQContext()
				priorRFQContext = prev.PartialRFQContext
				break
			}
		}
	}

	// 4. Log the interaction as INBOUND.
	inter := &LeadInteraction{
		LeadID:              int64(lead.ID),
		Channel:             "EMAIL",
		Direction:           "INBOUND",
		Subject:             req.Subject,
		Content:             req.Body,
		RawEmailID:          req.MessageID,
		ThreadID:            threadID,
		Sentiment:           "NEUTRAL",
		Intent:              "RFQ_REQUEST",
		AIConfidence:        0,
		ParentInteractionID: parentInteractionID,
	}
	err = h.leadsBL.LogInteraction(r.Context(), orgID, inter)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to log lead interaction: "+err.Error(), "DB_ERROR")
		return
	}

	// 5. Enqueue the EMAIL_PARSE task in PostgreSQL task queue for the Python AI sidecar.
	taskPayload := map[string]interface{}{
		"from":           req.From,
		"subject":        req.Subject,
		"body":           req.Body,
		"message_id":     req.MessageID,
		"thread_id":      threadID,
		"interaction_id": inter.ID,
		"lead_id":        lead.ID,
		"callback_url":   h.backendBaseURL + "/internal/sales/callback",
		// Thread-awareness fields: agent uses structured context, NOT raw emails
		"is_reply":              isReply,
		"parent_interaction_id": parentInteractionID,
		"prior_rfq_context":     priorRFQContext,
	}

	err = h.leadsBL.CreateAITask(
		r.Context(),
		int64(orgID),
		"LEAD",
		strconv.FormatInt(inter.ID, 10),
		"EMAIL_PARSE",
		taskPayload,
	)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create parsing task in queue: "+err.Error(), "DB_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Inbound email queued successfully for AI parsing", map[string]interface{}{
		"interaction_id": inter.ID,
		"thread_id":      threadID,
		"lead_id":        lead.ID,
		"is_reply":       isReply,
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

	var parsedDate time.Time
	if req.TargetDate != "" {
		parsedDate, _ = time.Parse("2006-01-02", req.TargetDate)
	} else {
		parsedDate = time.Now().AddDate(0, 1, 0) // default 1 month out
	}

	rfqReq := rfqspec.CreateRFQRequest{
		CustomerID:  req.CustomerID,
		Origin:      &req.Origin,
		Destination: &req.Destination,
		Incoterms:   &req.Incoterms,
		TargetDate:  &parsedDate,
		Items:       req.Items,
	}
	rfqReq.OrgID = req.OrgID

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

	// If intent is incomplete and partial context is provided, persist it so the next reply can restore state.
	if req.Intent == "RFQ_REQUEST_INCOMPLETE" && len(req.PartialRFQContext) > 0 {
		if ctxErr := h.leadsBL.UpdateInteractionContext(r.Context(), req.OrgID, req.InteractionID, req.PartialRFQContext); ctxErr != nil {
			// Non-fatal: log but don't fail the callback
			fmt.Printf("[SalesCallback] Warning: failed to persist partial RFQ context for interaction %d: %v\n", req.InteractionID, ctxErr)
		}
	}

	// Update Lead status based on intent
	if req.LinkedRFQID != nil && *req.LinkedRFQID > 0 {
		// Update lead status to ACTIVE since they have active RFQ request
		updateReq := spec.UpdateLeadRequest{
			OrgID:  int32(req.OrgID),
			ID:     int32(req.LeadID),
			Status: strPtr("ACTIVE"),
		}
		_, _ = h.leadsBL.UpdateLead(r.Context(), updateReq)
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

	// Fetch org_id from query (cognito auth simulation for simplicity, or hardcoded for now)
	orgIDStr := r.URL.Query().Get("org_id")
	if orgIDStr == "" {
		orgIDStr = "1"
	}
	orgID, _ := strconv.Atoi(orgIDStr)

	list, err := h.leadsBL.ListInteractions(r.Context(), int32(orgID), int32(leadID))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list interactions: "+err.Error(), "DB_ERROR")
		return
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

func strPtr(s string) *string {
	return &s
}
