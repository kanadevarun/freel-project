package shipments

import (
	"encoding/json"
	"time"

	"github.com/freel/backend/internal/carrier"
	"github.com/freel/backend/internal/shipments/spec"
)

// Normalize converts a raw carrier.TrackingEvent into the canonical shipments.NormalizedTrackingEvent contract
func Normalize(raw carrier.TrackingEvent, carrierSCAC string, sourceType string) spec.NormalizedTrackingEvent {
	bookingNum := raw.BookingNumber
	containerNum := raw.ContainerNumber
	mblNum := raw.MBLNumber

	// Fallback to inspecting raw.RawPayload if references are not populated on struct
	if (bookingNum == "" || containerNum == "" || mblNum == "") && len(raw.RawPayload) > 0 {
		var m map[string]interface{}
		if err := json.Unmarshal(raw.RawPayload, &m); err == nil {
			if containerNum == "" {
				for _, k := range []string{"container", "container_number", "containerNumber", "equipmentNo", "equipmentReference"} {
					if v, ok := m[k].(string); ok && v != "" {
						containerNum = v
						break
					}
				}
			}
			if bookingNum == "" {
				for _, k := range []string{"booking", "booking_number", "bookingNumber", "carrierBookingReference", "bookingNum"} {
					if v, ok := m[k].(string); ok && v != "" {
						bookingNum = v
						break
					}
				}
			}
			if mblNum == "" {
				for _, k := range []string{"mbl", "mbl_number", "mblNumber", "transportDocumentReference", "bol"} {
					if v, ok := m[k].(string); ok && v != "" {
						mblNum = v
						break
					}
				}
			}
		}
	}

	return spec.NormalizedTrackingEvent{
		EventID:         raw.EventID,
		SourceType:      sourceType,
		CarrierSCAC:     carrierSCAC,
		BookingNumber:   bookingNum,
		ContainerNumber: containerNum,
		MBLNumber:       mblNum,
		MilestoneCode:   raw.MilestoneCode,
		EventTime:       raw.EventTime,
		Location:        raw.Location,
		Description:     raw.Description,
		VesselName:      raw.VesselName,
		VoyageNumber:    raw.VoyageNumber,
		RawPayload:      json.RawMessage(raw.RawPayload),
		ReceivedAt:      time.Now(),
	}
}
