package shipments

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
)

type CarrierEmailRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	MessageID string `json:"message_id"`
}

// InboundCarrierEmailWebhook handles carrier email pushes to /api/v1/emails/carrier-inbound
func (h *Handler) InboundCarrierEmailWebhook(w http.ResponseWriter, r *http.Request) {
	var req CarrierEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}

	if req.From == "" || req.Body == "" {
		utils.Error(w, http.StatusBadRequest, "From and Body are required fields", "MISSING_PARAMS")
		return
	}

	// 3. Separate Ops Ingestion Path: Parse carrier email into NormalizedTrackingEvent
	normalized, err := ParseCarrierEmail(&req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "Failed to parse carrier email: "+err.Error(), "PARSE_ERROR")
		return
	}

	// Dynamic organization resolution based on integration configurations matching carrier email domains
	var orgID int64
	dbErr := h.svc.(*service).db.GetContext(r.Context(), &orgID, `
		SELECT org_id FROM carrier_integrations 
		WHERE carrier_scac = ? AND is_active = 1 LIMIT 1
	`, normalized.CarrierSCAC)
	if dbErr != nil {
		userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
		if ok && userCtx.OrgID > 0 {
			orgID = userCtx.OrgID
		} else if os.Getenv("APP_ENV") != "production" {
			_ = h.svc.(*service).db.GetContext(r.Context(), &orgID, "SELECT id FROM organizations LIMIT 1")
		}
	}

	if orgID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Unable to resolve org_id for carrier email", "RESOLVE_ERROR")
		return
	}

	err = h.svc.HandleInboundCarrierEvent(r.Context(), orgID, normalized)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to ingest carrier email event: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.Success(w, http.StatusOK, "Carrier email processed and enqueued", map[string]string{
		"event_id": normalized.EventID,
	})
}

// ParseCarrierEmail maps raw carrier email details to canonical NormalizedTrackingEvent contracts
func ParseCarrierEmail(req *CarrierEmailRequest) (*NormalizedTrackingEvent, error) {
	// Attempt to extract Booking, Container, HBL, MBL, and SCAC mappings via headers and regex
	scac := "MAEU" // Default to Maersk for testing if cannot resolve
	fromLower := strings.ToLower(req.From)
	if strings.Contains(fromLower, "msc.com") || strings.Contains(fromLower, "msc") {
		scac = "MSC"
	} else if strings.Contains(fromLower, "cma-cgm.com") || strings.Contains(fromLower, "cma") {
		scac = "CMA"
	} else if strings.Contains(fromLower, "hapag") {
		scac = "HAPAG"
	}

	body := req.Body
	bookingNum := extractRegex(body, `Booking\s*(?:Number|Ref|#)?\s*:?\s*([A-Za-z0-9\-]+)`)
	containerNum := extractRegex(body, `Container\s*(?:Number|#)?\s*:?\s*([A-Za-z0-9\-]{11})`)
	mblNum := extractRegex(body, `MBL\s*(?:Number|#)?\s*:?\s*([A-Za-z0-9\-]+)`)
	hblNum := extractRegex(body, `HBL\s*(?:Number|#)?\s*:?\s*([A-Za-z0-9\-]+)`)

	eventID := req.MessageID
	if eventID == "" {
		eventID = fmt.Sprintf("EMAIL-EVT-%d", time.Now().UnixNano())
	}

	rawPayload, _ := json.Marshal(req)

	return &NormalizedTrackingEvent{
		EventID:         eventID,
		SourceType:      "EMAIL",
		CarrierSCAC:     scac,
		BookingNumber:   bookingNum,
		ContainerNumber: containerNum,
		MBLNumber:       mblNum,
		HBLNumber:       hblNum,
		EventTime:       time.Now(),
		Description:     fmt.Sprintf("Email Subject: %s\n\n%s", req.Subject, req.Body),
		RawPayload:      rawPayload,
		ReceivedAt:      time.Now(),
	}, nil
}

func extractRegex(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}
