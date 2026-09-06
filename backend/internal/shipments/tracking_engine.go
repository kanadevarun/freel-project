package shipments

import (
	"time"

	"github.com/freel/backend/internal/shipments/spec"
)

// CalculateTrackingSummary evaluates operational progress, schedule health, ETA variance,
// and closure eligibility from authoritative persisted database records.
func CalculateTrackingSummary(
	sh *spec.Shipment,
	milestones []*spec.ShipmentMilestone,
	exceptions []*spec.ShipmentException,
) spec.ShipmentTrackingSummary {
	summary := spec.ShipmentTrackingSummary{
		ShipmentStatus: sh.Status,
		ClosureStatus:  sh.ClosureStatus,
	}

	// 1. Identify milestone maps for lookup
	mMap := make(map[string]*spec.ShipmentMilestone)
	for _, m := range milestones {
		mMap[m.MilestoneCode] = m
	}

	// 2. Determine highest completed milestone rank
	milestoneOrder := []string{
		spec.BOOKING_PENDING,
		spec.BOOKED,
		spec.DEPARTED,
		spec.IN_TRANSIT,
		spec.ARRIVED,
		spec.DELIVERED,
	}

	highestCompleted := spec.BOOKING_PENDING
	highestCompletedRank := 0

	// Check milestones in database
	for rank, code := range milestoneOrder {
		if m, exists := mMap[code]; exists && m.Status == "COMPLETED" {
			if rank > highestCompletedRank {
				highestCompleted = code
				highestCompletedRank = rank
			}
		}
	}

	// Make sure progress matches the actual highest completed milestone
	progressMap := map[string]int{
		spec.BOOKING_PENDING: 10,
		spec.BOOKED:          30,
		spec.DEPARTED:        50,
		spec.IN_TRANSIT:      70,
		spec.ARRIVED:         90,
		spec.DELIVERED:       100,
	}
	summary.HighestCompletedMilestone = highestCompleted
	summary.ProgressPercentage = progressMap[highestCompleted]

	// 3. Planned vs Actual dates
	// ETD (DEPARTED milestone)
	if depM, exists := mMap[spec.BOOKED]; exists && depM.PlannedDate != nil {
		// Use Booked milestone planned date as ETD fallback if Departed is missing
		summary.PlannedETD = depM.PlannedDate
	}
	if depM, exists := mMap[spec.DEPARTED]; exists {
		if depM.PlannedDate != nil {
			summary.PlannedETD = depM.PlannedDate
		}
		if depM.ActualDate != nil {
			summary.ActualETD = depM.ActualDate
		}
	}
	if summary.PlannedETD == nil {
		summary.PlannedETD = sh.ETD
	}

	// ETA (ARRIVED milestone or DELIVERED milestone)
	var plannedETA *time.Time
	var actualArrival *time.Time

	if arrM, exists := mMap[spec.ARRIVED]; exists {
		plannedETA = arrM.PlannedDate
		actualArrival = arrM.ActualDate
	}
	if delM, exists := mMap[spec.DELIVERED]; exists {
		if plannedETA == nil {
			plannedETA = delM.PlannedDate
		}
		if actualArrival == nil {
			actualArrival = delM.ActualDate
		}
	}
	if plannedETA == nil {
		plannedETA = sh.ETA
	}

	summary.PlannedETA = plannedETA
	summary.ActualArrival = actualArrival

	// 4. Schedule Variance
	summary.ScheduleVarianceState = spec.VarianceAwaitingData
	if actualArrival != nil && plannedETA != nil {
		diff := actualArrival.Sub(*plannedETA)
		diffDays := diff.Hours() / 24.0
		summary.ScheduleVariance = &diffDays

		if diffDays > 0.04 { // Late by > 1 hour
			summary.ScheduleVarianceState = spec.VarianceDelayed
		} else if diffDays < -0.04 { // Early by > 1 hour
			summary.ScheduleVarianceState = spec.VarianceEarly
		} else {
			summary.ScheduleVarianceState = spec.VarianceOnSchedule
		}
	} else if plannedETA != nil && actualArrival == nil {
		now := time.Now()
		if now.After(*plannedETA) {
			diff := now.Sub(*plannedETA)
			diffDays := diff.Hours() / 24.0
			summary.ScheduleVariance = &diffDays
			if diffDays > 0.04 {
				summary.ScheduleVarianceState = spec.VarianceDelayed
			} else {
				summary.ScheduleVarianceState = spec.VarianceOnSchedule
			}
		} else {
			summary.ScheduleVarianceState = spec.VarianceOnSchedule
		}
	}

	// 5. Unresolved Exceptions & counts
	var unresolvedExceptionsCount int64
	var hasCriticalOrHigh bool
	var hasMediumOrLow bool

	for _, ex := range exceptions {
		if !ex.Resolved && ex.Status != spec.ExceptionStatusDismissed {
			unresolvedExceptionsCount++
			if ex.Severity == spec.ExceptionSeverityCritical || ex.Severity == spec.ExceptionSeverityHigh {
				hasCriticalOrHigh = true
			} else if ex.Severity == spec.ExceptionSeverityMedium || ex.Severity == spec.ExceptionSeverityLow {
				hasMediumOrLow = true
			}
		}
	}
	summary.ActiveExceptionsCount = unresolvedExceptionsCount

	// 6. Schedule Health (Tracking State)
	now := time.Now()
	if sh.Status == spec.DELIVERED || sh.ClosureStatus == spec.ClosureStatusClosed {
		summary.TrackingState = spec.TrackingStateCompleted
	} else if hasCriticalOrHigh {
		summary.TrackingState = spec.TrackingStateException
	} else {
		// Overdue check
		isDelayed := false
		
		// If DEPARTED is overdue > 24 hours
		if depM, exists := mMap[spec.DEPARTED]; exists && depM.Status == "PLANNED" && depM.PlannedDate != nil {
			if now.Sub(*depM.PlannedDate) > 24*time.Hour {
				isDelayed = true
			}
		}
		// If ARRIVED/DELIVERED is overdue > 24 hours
		if plannedETA != nil && now.Sub(*plannedETA) > 24*time.Hour {
			isDelayed = true
		}

		if isDelayed {
			summary.TrackingState = spec.TrackingStateDelayed
		} else if hasMediumOrLow {
			summary.TrackingState = spec.TrackingStateAtRisk
		} else {
			summary.TrackingState = spec.TrackingStateOnTrack
		}
	}

	return summary
}

// EvaluateShipmentClosure checks the shipment eligibility for closure.
func EvaluateShipmentClosure(
	sh *spec.Shipment,
	milestones []*spec.ShipmentMilestone,
	exceptions []*spec.ShipmentException,
	docSummary *spec.ShipmentDocumentComplianceSummary,
) string {
	if sh.ClosureStatus == spec.ClosureStatusClosed {
		return spec.ClosureStatusClosed
	}

	// 1. Check if any open exceptions are CRITICAL
	hasOpenCriticalException := false
	for _, ex := range exceptions {
		if !ex.Resolved && ex.Status != spec.ExceptionStatusDismissed && ex.Severity == spec.ExceptionSeverityCritical {
			hasOpenCriticalException = true
		}
	}

	if hasOpenCriticalException {
		return spec.ClosureStatusBlocked
	}

	// 2. Check Document Compliance blockers (Task 16.7)
	if docSummary != nil && docSummary.ComplianceState == spec.ComplianceStateBlocked {
		return spec.ClosureStatusBlocked
	}

	// 3. Check if DELIVERED milestone is completed and other milestones are completed
	mMap := make(map[string]*spec.ShipmentMilestone)
	for _, m := range milestones {
		mMap[m.MilestoneCode] = m
	}

	// Required milestones to close are BOOKED, DEPARTED, IN_TRANSIT, ARRIVED, DELIVERED
	requiredCodes := []string{
		spec.BOOKED,
		spec.DEPARTED,
		spec.IN_TRANSIT,
		spec.ARRIVED,
		spec.DELIVERED,
	}

	allRequiredCompleted := true
	for _, code := range requiredCodes {
		m, exists := mMap[code]
		if !exists || m.Status != "COMPLETED" {
			allRequiredCompleted = false
			break
		}
	}

	if allRequiredCompleted {
		return spec.ClosureStatusReady
	}

	return spec.ClosureStatusActive
}
