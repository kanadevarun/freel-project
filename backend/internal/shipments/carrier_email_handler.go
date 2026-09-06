package shipments

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/freel/backend/internal/shipments/spec"
)

// ParseCarrierEmail maps raw carrier email details to canonical NormalizedTrackingEvent contracts
func ParseCarrierEmail(req *spec.CarrierEmailRequest) (*spec.NormalizedTrackingEvent, error) {
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

	return &spec.NormalizedTrackingEvent{
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
