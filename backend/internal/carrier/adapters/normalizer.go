package adapters

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/freel/backend/internal/carrier/domain"
)

// NormalizeDCSATrackingEvents transforms a slice of DCSA Track & Trace events into a unified NormalizedTrackingResult.
func NormalizeDCSATrackingEvents(carrierSCAC, requestedRef string, events []domain.DCSAEvent) *domain.NormalizedTrackingResult {
	result := &domain.NormalizedTrackingResult{
		CarrierSCAC:     carrierSCAC,
		ContainerNumber: requestedRef,
		FetchedAt:       time.Now().UTC(),
		Events:          make([]domain.NormalizedTrackingEvent, 0, len(events)),
	}

	var latestEventTime time.Time
	var latestLocation string
	var latestStatus string
	isDelivered := false
	var estimatedArrival *time.Time
	var actualDeparture *time.Time

	for _, ev := range events {
		eventTime := domain.ParseDCSATime(ev.EventDateTime)
		if eventTime.IsZero() {
			continue
		}

		// Milestone resolution
		milestoneRaw := ev.EquipmentEventTypeCode
		if milestoneRaw == "" {
			milestoneRaw = ev.TransportEventTypeCode
		}
		if milestoneRaw == "" {
			milestoneRaw = ev.ShipmentEventTypeCode
		}
		canonicalMilestone := domain.NormalizeDCSAMilestone(ev.EventType, milestoneRaw)

		// Location resolution
		locStr := ""
		if ev.EventLocation != nil {
			locStr = ev.EventLocation.LocationName
			if locStr == "" {
				locStr = ev.EventLocation.UNLocationCode
			}
		} else if ev.TransportCall != nil {
			if ev.TransportCall.Location != nil {
				locStr = ev.TransportCall.Location.LocationName
			}
			if locStr == "" {
				locStr = ev.TransportCall.UNLocationCode
			}
			if locStr == "" {
				locStr = ev.TransportCall.FacilityCode
			}
		}

		// Vessel details
		vesselName := ""
		voyageNum := ""
		if ev.TransportCall != nil {
			voyageNum = ev.TransportCall.CarrierVoyageNumber
			if ev.TransportCall.Vessel != nil {
				vesselName = ev.TransportCall.Vessel.VesselName
			}
		}

		containerNum := ev.EquipmentReference
		if containerNum == "" {
			containerNum = requestedRef
		}
		if result.ContainerNumber == "" {
			result.ContainerNumber = containerNum
		}
		if ev.CarrierBookingRef != "" && result.BookingNumber == "" {
			result.BookingNumber = ev.CarrierBookingRef
		}

		desc := fmt.Sprintf("%s (%s)", canonicalMilestone, ev.EventType)
		if locStr != "" {
			desc += " at " + locStr
		}

		normEv := domain.NormalizedTrackingEvent{
			EventID:         ev.EventID,
			MilestoneCode:   canonicalMilestone,
			EventTime:       eventTime,
			Location:        locStr,
			VesselName:      vesselName,
			VoyageNumber:    voyageNum,
			ContainerNumber: containerNum,
			Description:     desc,
		}
		result.Events = append(result.Events, normEv)

		// Track latest event status & milestones
		if ev.EventClassifierCode == domain.DCSAClassifierActual {
			if eventTime.After(latestEventTime) {
				latestEventTime = eventTime
				latestLocation = locStr
				latestStatus = canonicalMilestone
			}
			if canonicalMilestone == "DEPARTED" && actualDeparture == nil {
				actualDeparture = &eventTime
			}
			if canonicalMilestone == "DELIVERED" || canonicalMilestone == "GATE_OUT" {
				isDelivered = true
			}
		} else if ev.EventClassifierCode == domain.DCSAClassifierEstimated {
			if canonicalMilestone == "ARRIVED" || canonicalMilestone == "DISCHARGED" {
				if estimatedArrival == nil || eventTime.Before(*estimatedArrival) {
					estimatedArrival = &eventTime
				}
			}
		}
	}

	// Sort events chronologically (oldest to newest)
	sort.Slice(result.Events, func(i, j int) bool {
		return result.Events[i].EventTime.Before(result.Events[j].EventTime)
	})

	if latestStatus == "" {
		latestStatus = "IN_TRANSIT"
	}
	result.CurrentStatus = latestStatus
	result.LatestLocation = latestLocation
	result.IsDelivered = isDelivered
	result.EstimatedArrival = estimatedArrival
	result.ActualDeparture = actualDeparture

	return result
}

// RawMaerskRatePayload structure for tariff parsing.
type RawMaerskRatePayload struct {
	Rates []struct {
		PriceID         string  `json:"priceId"`
		Origin          string  `json:"origin"`
		Destination     string  `json:"destination"`
		Commodity       string  `json:"commodity"`
		EquipmentType   string  `json:"equipmentType"`
		Currency        string  `json:"currency"`
		BaseOceanPrice  float64 `json:"baseOceanPrice"`
		OriginSurcharge float64 `json:"originSurcharge"`
		DestSurcharge   float64 `json:"destSurcharge"`
		TotalPrice      float64 `json:"totalPrice"`
		TransitTimeDays int     `json:"transitTimeDays"`
		FreeDays        int     `json:"freeDays"`
		ValidityStart   string  `json:"validityStart"`
		ValidityEnd     string  `json:"validityEnd"`
		IsContract      bool    `json:"isContract"`
	} `json:"rates"`
}

// NormalizeMaerskRates converts raw Maersk pricing JSON into LogisticsHQ canonical rates.
func NormalizeMaerskRates(carrierSCAC string, rawJSON []byte) ([]domain.NormalizedCarrierRate, error) {
	var payload RawMaerskRatePayload
	if err := json.Unmarshal(rawJSON, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse carrier rates JSON: %w", err)
	}

	normalized := make([]domain.NormalizedCarrierRate, 0, len(payload.Rates))
	for _, r := range payload.Rates {
		validFrom := domain.ParseDCSATime(r.ValidityStart)
		if validFrom.IsZero() {
			validFrom = time.Now().UTC()
		}
		validUntil := domain.ParseDCSATime(r.ValidityEnd)
		if validUntil.IsZero() {
			validUntil = time.Now().UTC().AddDate(0, 1, 0)
		}

		total := r.TotalPrice
		if total == 0 {
			total = r.BaseOceanPrice + r.OriginSurcharge + r.DestSurcharge
		}

		rate := domain.NormalizedCarrierRate{
			RateID:             r.PriceID,
			CarrierSCAC:        carrierSCAC,
			CarrierName:        "A.P. Moller – Maersk",
			OriginPort:         r.Origin,
			DestinationPort:    r.Destination,
			EquipmentType:      r.EquipmentType,
			Currency:           r.Currency,
			OceanFreight:       r.BaseOceanPrice,
			OriginCharges:      r.OriginSurcharge,
			DestinationCharges: r.DestSurcharge,
			TotalBuyPrice:      total,
			TransitDays:        r.TransitTimeDays,
			FreeDays:           r.FreeDays,
			ValidFrom:          validFrom,
			ValidUntil:         validUntil,
			IsContractRate:     r.IsContract,
		}
		normalized = append(normalized, rate)
	}
	return normalized, nil
}
