package pricing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/freel/backend/internal/rates"
	"github.com/freel/backend/internal/rfq"
	rfqspec "github.com/freel/backend/internal/rfq/spec"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

// Handler handles internal pricing API endpoints.
type Handler struct {
	rulesSvc Service
	rfqBL    rfq.BusinessLogic
	rateSvc  rates.Service
}

// NewHandler creates a new Handler instance.
func NewHandler(rulesSvc Service, rfqBL rfq.BusinessLogic, rateSvc rates.Service) *Handler {
	return &Handler{
		rulesSvc: rulesSvc,
		rfqBL:    rfqBL,
		rateSvc:  rateSvc,
	}
}

// GetRules handles GET /internal/pricing/rules (invoked by sidecar tool)
func (h *Handler) GetRules(w http.ResponseWriter, r *http.Request) {
	// Authentication
	if err := h.authenticate(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DB_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Rules retrieved", rules)
}

// GetRFQDetails handles GET /internal/rfqs/{id} (invoked by sidecar tool)
func (h *Handler) GetRFQDetails(w http.ResponseWriter, r *http.Request) {
	// Authentication
	if err := h.authenticate(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
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
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "DB_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "RFQ details retrieved", rfqObj)
}

// SearchRates handles GET /internal/rates/search (bypasses Cognito for AI sidecar)
func (h *Handler) SearchRates(w http.ResponseWriter, r *http.Request) {
	// Authentication
	if err := h.authenticate(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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
		utils.Error(w, http.StatusInternalServerError, err.Error(), "SEARCH_FAILED")
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
	// Authentication
	if err := h.authenticate(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	for _, q := range req.Quotes {
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
			utils.Error(w, http.StatusInternalServerError, err.Error(), "ADD_QUOTE_FAILED")
			return
		}
	}

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
	// Authentication
	if err := h.authenticate(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	// Map AgentStatus string to represent in RFQ entity UI
	agentStatus := "COLLECTING_INFORMATION"
	switch req.Status {
	case "COMPLETED":
		agentStatus = "DRAFT_READY"
	case "FAILED":
		agentStatus = "FAILED"
	case "NEEDS_REVIEW":
		agentStatus = "WAITING_FOR_HUMAN"
	}

	if err := h.rfqBL.UpdateAgentStatus(r.Context(), req.OrgID, req.RFQID, agentStatus); err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "UPDATE_STATUS_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Pricing callback processed successfully", nil)
}

func (h *Handler) authenticate(r *http.Request) error {
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
