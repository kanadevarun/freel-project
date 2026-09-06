package shipments

import (
	"fmt"
	"time"

	"github.com/freel/backend/internal/shipments/spec"
)

// BuildTrackingSchedule evaluates planned vs actual transit timestamps and calculates variances.
func BuildTrackingSchedule(
	sh *spec.Shipment,
	milestones []*spec.ShipmentMilestone,
) spec.TrackingSchedule {
	mMap := make(map[string]*spec.ShipmentMilestone)
	for _, m := range milestones {
		mMap[m.MilestoneCode] = m
	}

	schedule := spec.TrackingSchedule{
		DepartureState:       spec.VarianceAwaitingData,
		ArrivalState:         spec.VarianceAwaitingData,
		OverallVarianceState: spec.VarianceAwaitingData,
	}

	// 1. Planned & Actual ETD
	if depM, exists := mMap[spec.DEPARTED]; exists {
		schedule.PlannedETD = depM.PlannedDate
		schedule.ActualETD = depM.ActualDate
	}
	if schedule.PlannedETD == nil {
		schedule.PlannedETD = sh.ETD
	}

	if schedule.PlannedETD != nil {
		if schedule.ActualETD != nil {
			diffDays := schedule.ActualETD.Sub(*schedule.PlannedETD).Hours() / 24.0
			schedule.DepartureVarianceDays = diffDays
			if diffDays > 0.04 {
				schedule.DepartureState = spec.VarianceDelayed
			} else if diffDays < -0.04 {
				schedule.DepartureState = spec.VarianceEarly
			} else {
				schedule.DepartureState = spec.VarianceOnSchedule
			}
		} else {
			if time.Now().After(*schedule.PlannedETD) {
				diffDays := time.Since(*schedule.PlannedETD).Hours() / 24.0
				schedule.DepartureVarianceDays = diffDays
				if diffDays > 0.04 {
					schedule.DepartureState = spec.VarianceDelayed
				} else {
					schedule.DepartureState = spec.VarianceOnSchedule
				}
			} else {
				schedule.DepartureState = spec.VarianceOnSchedule
			}
		}
	}

	// 2. Planned & Estimated ETA
	if arrM, exists := mMap[spec.ARRIVED]; exists {
		schedule.PlannedETA = arrM.PlannedDate
		schedule.EstimatedArrival = arrM.ActualDate
	}
	if delM, exists := mMap[spec.DELIVERED]; exists {
		if schedule.PlannedETA == nil {
			schedule.PlannedETA = delM.PlannedDate
		}
		if schedule.EstimatedArrival == nil {
			schedule.EstimatedArrival = delM.ActualDate
		}
	}
	if schedule.PlannedETA == nil {
		schedule.PlannedETA = sh.ETA
	}

	if schedule.PlannedETA != nil {
		if schedule.EstimatedArrival != nil {
			diffDays := schedule.EstimatedArrival.Sub(*schedule.PlannedETA).Hours() / 24.0
			schedule.ArrivalVarianceDays = diffDays
			if diffDays > 0.04 {
				schedule.ArrivalState = spec.VarianceDelayed
			} else if diffDays < -0.04 {
				schedule.ArrivalState = spec.VarianceEarly
			} else {
				schedule.ArrivalState = spec.VarianceOnSchedule
			}
		} else {
			if time.Now().After(*schedule.PlannedETA) {
				diffDays := time.Since(*schedule.PlannedETA).Hours() / 24.0
				schedule.ArrivalVarianceDays = diffDays
				if diffDays > 0.04 {
					schedule.ArrivalState = "AT_RISK"
				} else {
					schedule.ArrivalState = spec.VarianceOnSchedule
				}
			} else {
				schedule.ArrivalState = spec.VarianceOnSchedule
			}
		}
	}

	if schedule.ArrivalState == spec.VarianceDelayed || schedule.ArrivalState == "AT_RISK" {
		schedule.OverallVarianceState = spec.VarianceDelayed
	} else if schedule.DepartureState == spec.VarianceDelayed {
		schedule.OverallVarianceState = spec.VarianceDelayed
	} else {
		schedule.OverallVarianceState = spec.VarianceOnSchedule
	}

	return schedule
}

// BuildTrackingJourney builds the normalized operational journey milestones.
func BuildTrackingJourney(
	sh *spec.Shipment,
	milestones []*spec.ShipmentMilestone,
) spec.TrackingJourney {
	mMap := make(map[string]*spec.ShipmentMilestone)
	for _, m := range milestones {
		mMap[m.MilestoneCode] = m
	}

	milestoneDefinitions := []struct {
		Code  string
		Label string
	}{
		{Code: spec.BOOKED, Label: "Booking Confirmed"},
		{Code: "GATE_IN", Label: "Origin Gate In"},
		{Code: spec.DEPARTED, Label: "Vessel Departed"},
		{Code: spec.IN_TRANSIT, Label: "Ocean In Transit"},
		{Code: spec.ARRIVED, Label: "Arrived at Destination Port"},
		{Code: spec.DELIVERED, Label: "Cargo Delivered"},
	}

	journeyMilestones := make([]spec.TrackingJourneyMilestone, 0, len(milestoneDefinitions))
	currentFound := false
	completedCount := 0
	var currentCode, currentLabel string
	var criticalDelay float64

	now := time.Now()

	for _, def := range milestoneDefinitions {
		m, exists := mMap[def.Code]
		jm := spec.TrackingJourneyMilestone{
			Code:  def.Code,
			Label: def.Label,
			State: "UPCOMING",
		}

		if exists && m != nil {
			jm.PlannedDate = m.PlannedDate
			jm.ActualDate = m.ActualDate
			if m.Location != nil {
				jm.Location = *m.Location
			}
			if m.Notes != nil {
				jm.Notes = m.Notes
			}

			if m.Status == "COMPLETED" {
				jm.IsCompleted = true
				jm.State = "COMPLETED"
				completedCount++
			}
		}

		// Fallback locations
		if jm.Location == "" {
			switch def.Code {
			case spec.BOOKED:
				jm.Location = "Carrier Network"
			case "GATE_IN":
				jm.Location = sh.OriginPort
			case spec.DEPARTED:
				jm.Location = sh.OriginPort
			case spec.IN_TRANSIT:
				jm.Location = "Oceanic Corridor"
			case spec.ARRIVED:
				jm.Location = sh.DestinationPort
			case spec.DELIVERED:
				jm.Location = "Consignee Facility"
			}
		}

		// Evaluate delay
		if !jm.IsCompleted && jm.PlannedDate != nil && now.After(*jm.PlannedDate) {
			diffDays := now.Sub(*jm.PlannedDate).Hours() / 24.0
			if diffDays > 0.04 {
				jm.IsDelayed = true
				jm.DelayDays = diffDays
				if diffDays > criticalDelay {
					criticalDelay = diffDays
				}
			}
		}

		// Determine current milestone
		if !jm.IsCompleted && !currentFound {
			jm.IsCurrent = true
			if jm.IsDelayed {
				jm.State = "DELAYED"
			} else {
				jm.State = "CURRENT"
			}
			currentFound = true
			currentCode = def.Code
			currentLabel = def.Label
		}

		journeyMilestones = append(journeyMilestones, jm)
	}

	if !currentFound && completedCount == len(milestoneDefinitions) {
		currentCode = spec.DELIVERED
		currentLabel = "Cargo Delivered"
	}

	return spec.TrackingJourney{
		Milestones:            journeyMilestones,
		CurrentMilestoneCode:  currentCode,
		CurrentMilestoneLabel: currentLabel,
		TotalMilestones:       len(milestoneDefinitions),
		CompletedMilestones:   completedCount,
		CriticalPathDelayDays: criticalDelay,
	}
}

// GenerateTrackingAlerts generates read-only operational alerts from existing shipment conditions.
func GenerateTrackingAlerts(
	sh *spec.Shipment,
	milestones []*spec.ShipmentMilestone,
	exceptions []*spec.ShipmentException,
	pos *spec.TrackingPosition,
	schedule *spec.TrackingSchedule,
) []spec.TrackingAlert {
	alerts := make([]spec.TrackingAlert, 0)
	now := time.Now()

	// 1. Critical & High Exceptions
	for _, ex := range exceptions {
		if !ex.Resolved && ex.Status != spec.ExceptionStatusDismissed {
			actionURL := fmt.Sprintf("/dashboard/shipments/%d", sh.ID)
			severity := "WARNING"
			if ex.Severity == spec.ExceptionSeverityCritical || ex.Severity == spec.ExceptionSeverityHigh {
				severity = "CRITICAL"
			}
			desc := ""
			if ex.Description != nil {
				desc = *ex.Description
			}
			alerts = append(alerts, spec.TrackingAlert{
				ID:          fmt.Sprintf("alert-exc-%d", ex.ID),
				Type:        "CRITICAL_EXCEPTION",
				Severity:    severity,
				Title:       ex.ExceptionType,
				Description: desc,
				CreatedAt:   ex.CreatedAt,
				ActionURL:   &actionURL,
			})
		}
	}

	// 2. Schedule Delays
	if schedule != nil {
		if schedule.DepartureVarianceDays > 0.04 {
			alerts = append(alerts, spec.TrackingAlert{
				ID:          "alert-sched-dep",
				Type:        "DELAYED_DEPARTURE",
				Severity:    "WARNING",
				Title:       "Delayed Origin Departure",
				Description: fmt.Sprintf("Shipment departure was delayed by %.1f days vs planned schedule.", schedule.DepartureVarianceDays),
				CreatedAt:   now,
			})
		}

		if schedule.ArrivalState == "AT_RISK" || schedule.ArrivalVarianceDays > 0.04 {
			alerts = append(alerts, spec.TrackingAlert{
				ID:          "alert-sched-arr",
				Type:        "ETA_AT_RISK",
				Severity:    "WARNING",
				Title:       "Destination ETA at Risk",
				Description: "Vessel arrival projection is currently running behind the planned delivery schedule.",
				CreatedAt:   now,
			})
		}
	}

	// 3. Telemetry Freshness Alerts
	if pos != nil {
		if pos.DataFreshness == spec.TrackingFreshnessLive || pos.DataFreshness == spec.TrackingFreshnessRecent {
			alerts = append(alerts, spec.TrackingAlert{
				ID:          "alert-telemetry-live",
				Type:        "LIVE_TELEMETRY",
				Severity:    "INFO",
				Title:       "Live Telemetry Synchronized",
				Description: fmt.Sprintf("Vessel position updated via %s (speed: %.1f kts, heading: %.0f°).", pos.TrackingSource, pos.SpeedKnots, pos.HeadingDegrees),
				CreatedAt:   pos.RecordedAt,
			})
		} else if pos.DataFreshness == spec.TrackingFreshnessStale {
			alerts = append(alerts, spec.TrackingAlert{
				ID:          "alert-telemetry-stale",
				Type:        "STALE_TRACKING",
				Severity:    "INFO",
				Title:       "Tracking Telemetry Stale",
				Description: "Last satellite AIS update was received > 6 hours ago. Milestone tracking remains active.",
				CreatedAt:   pos.RecordedAt,
			})
		}
	} else {
		alerts = append(alerts, spec.TrackingAlert{
			ID:          "alert-telemetry-unavail",
			Type:        "MISSING_TRACKING_DATA",
			Severity:    "INFO",
			Title:       "Live Telemetry Unavailable",
			Description: "Operational visibility is currently tracking against schedule milestones and carrier status.",
			CreatedAt:   now,
		})
	}

	return alerts
}
