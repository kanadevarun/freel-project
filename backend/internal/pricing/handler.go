package pricing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/approvals"
	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/rates"
	"github.com/freel/backend/internal/rfq"
	rfqspec "github.com/freel/backend/internal/rfq/spec"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

// Handler handles internal pricing API endpoints.
type Handler struct {
	rulesSvc     Service
	rfqBL        rfq.BusinessLogic
	rateSvc      rates.Service
	approvalsSvc approvals.Service
}

// NewHandler creates a new Handler instance.
func NewHandler(rulesSvc Service, rfqBL rfq.BusinessLogic, rateSvc rates.Service) *Handler {
	return &Handler{
		rulesSvc: rulesSvc,
		rfqBL:    rfqBL,
		rateSvc:  rateSvc,
	}
}

func (h *Handler) SetApprovalsService(svc approvals.Service) {
	h.approvalsSvc = svc
}

// GetRules handles GET /internal/pricing/rules (invoked by sidecar tool)
func (h *Handler) GetRules(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized access: Invalid service key token", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	orgID, _ := strconv.ParseInt(q.Get("org_id"), 10, 64)
	if orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id query param is required", "MISSING_PARAM")
		return
	}

	origin := q.Get("origin")
	destination := q.Get("destination")
	tier := q.Get("tier")
	equipment := q.Get("equipment")

	rules, err := h.rulesSvc.GetApplicableRules(r.Context(), orgID, origin, destination, tier, equipment)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to retrieve pricing rules", "DB_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Rules retrieved", rules)
}

// GetRFQDetails handles GET /internal/rfqs/{id} (invoked by sidecar tool)
func (h *Handler) GetRFQDetails(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized access: Invalid service key token", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid rfq id", "INVALID_PARAM")
		return
	}

	q := r.URL.Query()
	orgID, _ := strconv.ParseInt(q.Get("org_id"), 10, 64)
	if orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id query param is required", "MISSING_PARAM")
		return
	}

	rfqObj, err := h.rfqBL.GetRFQ(r.Context(), int32(orgID), int32(id))
	if err != nil || rfqObj == nil {
		utils.Error(w, http.StatusNotFound, "RFQ not found or access denied", "NOT_FOUND")
		return
	}

	utils.Success(w, http.StatusOK, "RFQ details retrieved", rfqObj)
}

// SearchRates handles GET /internal/rates/search (bypasses Cognito for AI sidecar)
func (h *Handler) SearchRates(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized access: Invalid service key token", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	orgID, _ := strconv.ParseInt(q.Get("org_id"), 10, 64)
	if orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "org_id query param is required", "MISSING_PARAM")
		return
	}

	origin := q.Get("origin")
	destination := q.Get("destination")
	equipment := q.Get("equipment")
	if equipment == "" {
		equipment = "40GP"
	}
	incoterms := q.Get("incoterms")

	rateQuery := rates.RateQuery{
		OrgID:           orgID,
		OriginPort:      origin,
		DestinationPort: destination,
		EquipmentType:   equipment,
		MaxResults:      20,
		Incoterms:       incoterms,
	}

	result, err := h.rateSvc.SearchRates(r.Context(), rateQuery)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to search rates", "SEARCH_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Rates retrieved successfully", result)
}

type DraftQuoteInput struct {
	CarrierName           string   `json:"carrier_name"`
	TransitTimeDays       *int     `json:"transit_time_days"`
	BuyPrice              float64  `json:"buy_price"`
	SellPrice             float64  `json:"sell_price"`
	IsRecommended         bool     `json:"is_recommended"`
	ReliabilityScore      int      `json:"reliability_score"`
	HistoricalSuccessRate float64  `json:"historical_success_rate"`
	AiReasoning           *string  `json:"ai_reasoning"`
}

type CreateDraftQuotesRequest struct {
	RFQID  int32             `json:"rfq_id"`
	OrgID  int32             `json:"org_id"`
	Quotes []DraftQuoteInput `json:"quotes"`
}

// CreateDraftQuotes handles POST /internal/pricing/quotes/draft (invoked by sidecar validation/save node)
func (h *Handler) CreateDraftQuotes(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized access: Invalid service key token", "UNAUTHORIZED")
		return
	}

	var req CreateDraftQuotesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	if req.RFQID <= 0 || req.OrgID <= 0 || len(req.Quotes) == 0 {
		utils.Error(w, http.StatusBadRequest, "rfq_id, org_id, and quotes are required", "MISSING_PARAMS")
		return
	}

	// Tenant ownership verification: Ensure RFQ exists and belongs to the specified org_id
	rfqObj, err := h.rfqBL.GetRFQ(r.Context(), req.OrgID, req.RFQID)
	if err != nil || rfqObj == nil {
		utils.Error(w, http.StatusNotFound, "RFQ not found or organization mismatch", "NOT_FOUND")
		return
	}

	for _, q := range req.Quotes {
		if strings.TrimSpace(q.CarrierName) == "" || q.BuyPrice < 0 || q.SellPrice < 0 {
			utils.Error(w, http.StatusBadRequest, "Invalid quote pricing or missing carrier name", "INVALID_PAYLOAD")
			return
		}
		quote := &rfqspec.Quote{
			RFQID:                 req.RFQID,
			CarrierName:           q.CarrierName,
			TransitTimeDays:       q.TransitTimeDays,
			BuyPrice:              q.BuyPrice,
			SellPrice:             q.SellPrice,
			IsRecommended:         q.IsRecommended,
			ReliabilityScore:      q.ReliabilityScore,
			HistoricalSuccessRate: q.HistoricalSuccessRate,
			AiReasoning:           q.AiReasoning,
			Status:                "DRAFT",
		}
		if err := h.rfqBL.AddQuote(r.Context(), req.OrgID, quote); err != nil {
			utils.Error(w, http.StatusInternalServerError, "Failed to create draft quote", "ADD_QUOTE_FAILED")
			return
		}
	}

	// Preserve actor context: AI_AGENT universal audit log
	_, _ = audit.Record(r.Context(), domain.CreateAuditLogParams{
		OrgID:        int64(req.OrgID),
		ActorType:    domain.ActorTypeAIAgent,
		ActorName:    "AI Agent: PricingAgent",
		ActorRole:    "AI_AGENT",
		Action:       domain.ActionCreate,
		Module:       domain.ModuleQuotations,
		ResourceType: "QUOTE",
		ResourceID:   strconv.Itoa(int(req.RFQID)),
		Description:  fmt.Sprintf("AI Pricing Agent created %d draft quote(s) for RFQ #%d", len(req.Quotes), req.RFQID),
		Result:       domain.ResultSuccess,
		Metadata: map[string]interface{}{
			"rfq_id": req.RFQID,
			"source": "AI_AGENT",
		},
	})

	utils.Success(w, http.StatusOK, "Draft quotes created successfully", nil)
}

type PricingCallbackRequest struct {
	RFQID         int32  `json:"rfq_id"`
	OrgID         int32  `json:"org_id"`
	Status        string `json:"status"` // COMPLETED | FAILED | NEEDS_REVIEW
	CorrelationID string `json:"correlation_id"`
	AiReasoning   string `json:"ai_reasoning"`
}

// Callback handles POST /internal/pricing/callback (invoked by sidecar agent)
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	if err := middleware.ValidateInternalServiceToken(r); err != nil {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized access: Invalid service key token", "UNAUTHORIZED")
		return
	}

	var req PricingCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	if req.RFQID <= 0 || req.OrgID <= 0 || req.Status == "" {
		utils.Error(w, http.StatusBadRequest, "rfq_id, org_id, and status are required", "MISSING_PARAMS")
		return
	}

	// Tenant ownership verification: Ensure RFQ exists and belongs to the specified org_id
	rfqObj, err := h.rfqBL.GetRFQ(r.Context(), req.OrgID, req.RFQID)
	if err != nil || rfqObj == nil {
		utils.Error(w, http.StatusNotFound, "RFQ not found or organization mismatch", "NOT_FOUND")
		return
	}

	// Map AgentStatus string to represent in RFQ entity UI
	agentStatus := "COLLECTING_INFORMATION"
	switch req.Status {
	case "COMPLETED":
		agentStatus = "DRAFT_READY"
	case "FAILED":
		agentStatus = "FAILED"
	case "NEEDS_REVIEW", "WAITING_FOR_HUMAN":
		agentStatus = "WAITING_FOR_HUMAN"
	}

	if err := h.rfqBL.UpdateAgentStatus(r.Context(), req.OrgID, req.RFQID, agentStatus); err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update RFQ agent status", "UPDATE_STATUS_FAILED")
		return
	}

	// If human approval is required, create or link canonical approval request
	if agentStatus == "WAITING_FOR_HUMAN" && h.approvalsSvc != nil {
		threadID := fmt.Sprintf("rfq-%d", req.RFQID)
		approvalRef := fmt.Sprintf("approval-pricing.anomaly-rfq%d", req.RFQID)
		_, _ = h.approvalsSvc.ProposeAIApproval(r.Context(), int64(req.OrgID), &approvals.ProposeAIApprovalInput{
			Title:              fmt.Sprintf("Pricing Anomaly Approval for RFQ #%d", req.RFQID),
			Category:           "COMMERCIAL",
			Type:               "Pricing Anomaly Approval",
			Priority:           "HIGH",
			RelatedEntityType:  "RFQ",
			RelatedEntityID:    int64(req.RFQID),
			RelatedRef:         fmt.Sprintf("RFQ #%d", req.RFQID),
			Description:        req.AiReasoning,
			ActorType:          "AI_AGENT",
			Source:             "langgraph.pricing",
			ActionName:         "pricing.save_draft_quotes",
			RiskLevel:          "HIGH_RISK",
			RequiredPermission: "rfqs:approve",
			ThreadID:           threadID,
			ApprovalReference:  approvalRef,
			CorrelationID:      req.CorrelationID,
			ExpiresInHours:     48,
		})
	}

	// Preserve actor context: AI_AGENT universal audit log
	_, _ = audit.Record(r.Context(), domain.CreateAuditLogParams{
		OrgID:        int64(req.OrgID),
		ActorType:    domain.ActorTypeAIAgent,
		ActorName:    "AI Agent: PricingAgent",
		ActorRole:    "AI_AGENT",
		Action:       domain.ActionUpdate,
		Module:       domain.ModuleRFQs,
		ResourceType: "RFQ",
		ResourceID:   strconv.Itoa(int(req.RFQID)),
		Description:  fmt.Sprintf("AI Pricing Agent callback processed with status %s: %s", req.Status, agentStatus),
		Result:       domain.ResultSuccess,
		Metadata: map[string]interface{}{
			"rfq_id":         req.RFQID,
			"source":         "AI_AGENT",
			"status":         req.Status,
			"correlation_id": req.CorrelationID,
		},
	})

	utils.Success(w, http.StatusOK, "Pricing callback processed successfully", nil)
}
