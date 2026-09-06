package shipments

import (
	"encoding/json"
	"time"

	"github.com/freel/backend/internal/carrier"
	"github.com/freel/backend/internal/shipments/spec"
)

// Normalize converts a raw carrier.TrackingEvent into the canonical shipments.NormalizedTrackingEvent contract
func Normalize(raw carrier.TrackingEvent, carrierSCAC string, sourceType string) spec.NormalizedTrackingEvent {
	return spec.NormalizedTrackingEvent{
		EventID:         raw.EventID,
		SourceType:      sourceType,
		CarrierSCAC:     carrierSCAC,
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
