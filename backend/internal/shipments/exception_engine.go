package shipments

import (
	"fmt"
	"time"

	"github.com/freel/backend/internal/shipments/spec"
)

// EvaluateDeterministicExceptions evaluates milestones and shipment metadata for the shipment ID,
// returning a list of ShipmentExceptions to be created/upserted.
func EvaluateDeterministicExceptions(sh *spec.Shipment, milestones []*spec.ShipmentMilestone) []*spec.ShipmentException {
	var exceptions []*spec.ShipmentException
	now := time.Now()

	for _, m := range milestones {
		// 1. Schedule Delay check
		// Completed milestones: actual_date > planned_date
		if m.Status == "COMPLETED" && m.ActualDate != nil && m.PlannedDate != nil {
			if m.ActualDate.After(*m.PlannedDate) {
				diff := m.ActualDate.Sub(*m.PlannedDate)
				// If delayed by more than 1 hour (tolerance)
				if diff > 1*time.Hour {
					sourceID := fmt.Sprintf("DELAY-%s", m.MilestoneCode)
					desc := fmt.Sprintf("Milestone %s was completed late by %s (Planned: %s, Actual: %s).",
						m.MilestoneCode, diff.Round(time.Minute), m.PlannedDate.Format(time.RFC3339), m.ActualDate.Format(time.RFC3339))
					
					exceptions = append(exceptions, &spec.ShipmentException{
						OrgID:         sh.OrgID,
						ShipmentID:    sh.ID,
						ExceptionType: spec.ExceptionCategoryScheduleDelay,
						Severity:      spec.ExceptionSeverityMedium,
						Status:        spec.ExceptionStatusOpen,
						Title:         fmt.Sprintf("%s Delayed Completion", m.MilestoneCode),
						Description:   &desc,
						SourceEventID: &sourceID,
					})
				}
			}
		}

		// 2. Overdue Milestone check
		// Planned milestones: planned_date has passed but milestone is still PLANNED
		if m.Status == "PLANNED" && m.PlannedDate != nil {
			if now.After(*m.PlannedDate) {
				diff := now.Sub(*m.PlannedDate)
				// If overdue by more than 24 hours
				if diff > 24*time.Hour {
					sourceID := fmt.Sprintf("OVERDUE-%s", m.MilestoneCode)
					desc := fmt.Sprintf("Milestone %s is overdue by %s. Planned date was %s but is still not completed.",
						m.MilestoneCode, diff.Round(time.Hour), m.PlannedDate.Format(time.RFC3339))
					
					exceptions = append(exceptions, &spec.ShipmentException{
						OrgID:         sh.OrgID,
						ShipmentID:    sh.ID,
						ExceptionType: spec.ExceptionCategoryScheduleDelay,
						Severity:      spec.ExceptionSeverityHigh,
						Status:        spec.ExceptionStatusOpen,
						Title:         fmt.Sprintf("%s Overdue", m.MilestoneCode),
						Description:   &desc,
						SourceEventID: &sourceID,
					})
				}
			}
		}
	}

	// 3. Overall ETA Risk / ETA Delay check
	// If current time is past ETA but the final status is not DELIVERED or ARRIVED
	if sh.ETA != nil && now.After(*sh.ETA) {
		isFinished := false
		for _, m := range milestones {
			if (m.MilestoneCode == spec.DELIVERED || m.MilestoneCode == spec.ARRIVED) && m.Status == "COMPLETED" {
				isFinished = true
			}
		}
		// Also check main shipment status
		if sh.Status == spec.DELIVERED || sh.Status == spec.ARRIVED {
			isFinished = true
		}

		if !isFinished {
			diff := now.Sub(*sh.ETA)
			sourceID := "ETA-DELAY"
			desc := fmt.Sprintf("Shipment has missed its planned ETA date %s by %s.", sh.ETA.Format(time.RFC3339), diff.Round(time.Hour))
			
			exceptions = append(exceptions, &spec.ShipmentException{
				OrgID:         sh.OrgID,
				ShipmentID:    sh.ID,
				ExceptionType: spec.ExceptionCategoryETADelay,
				Severity:      spec.ExceptionSeverityCritical,
				Status:        spec.ExceptionStatusOpen,
				Title:         "ETA Overdue Risk",
				Description:   &desc,
				SourceEventID: &sourceID,
			})
		}
	}

	// 4. Overall ETD Delay Risk
	// If ETD is in the past and shipment is still BOOKING_PENDING or BOOKED
	if sh.ETD != nil && now.After(*sh.ETD) && (sh.Status == spec.BOOKING_PENDING || sh.Status == spec.BOOKED) {
		diff := now.Sub(*sh.ETD)
		sourceID := "ETD-DELAY"
		desc := fmt.Sprintf("Shipment has missed its scheduled departure date %s by %s and is still not in transit.", sh.ETD.Format(time.RFC3339), diff.Round(time.Hour))
		
		exceptions = append(exceptions, &spec.ShipmentException{
			OrgID:         sh.OrgID,
			ShipmentID:    sh.ID,
			ExceptionType: spec.ExceptionCategoryETDDelay,
			Severity:      spec.ExceptionSeverityHigh,
			Status:        spec.ExceptionStatusOpen,
			Title:         "ETD Departure Delay",
			Description:   &desc,
			SourceEventID: &sourceID,
		})
	}

	return exceptions
}
