package bcontext

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

// GetShipment360OperationsIntelligence assembles a comprehensive, deterministic operational intelligence view for a shipment.
func (s *defaultService) GetShipment360OperationsIntelligence(ctx context.Context, orgID, shipmentID int64, correlationID string, userID int64) (*Shipment360OperationsIntelligence, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization ID: %d", orgID)
	}
	if shipmentID <= 0 {
		return nil, fmt.Errorf("invalid shipment ID: %d", shipmentID)
	}
	if correlationID == "" {
		correlationID = fmt.Sprintf("sh-intel-%d-%d", shipmentID, time.Now().UnixNano())
	}

	now := time.Now().UTC()

	// 1. Fetch Shipment Identity and Joined Metadata
	shQuery := `
		SELECT 
			s.id, s.org_id, s.status, s.origin_port, s.destination_port, s.carrier_scac,
			COALESCE(c.name, s.carrier_scac) as carrier_name,
			s.vessel_name, s.voyage_number, s.booking_number, s.booking_id, s.mbl_number, s.hbl_number,
			s.container_numbers, s.rfq_id, s.quote_id, s.etd, s.eta, s.closure_status,
			s.created_at, s.updated_at,
			r.rfq_number,
			r.customer_id,
			cust.name as customer_name
		FROM shipments s
		LEFT JOIN carriers c ON s.carrier_scac = c.scac
		LEFT JOIN rfqs r ON s.rfq_id = r.id AND r.org_id = s.org_id
		LEFT JOIN customers cust ON r.customer_id = cust.id AND cust.org_id = s.org_id
		WHERE s.id = ? AND s.org_id = ?
		LIMIT 1
	`

	var (
		shID, shOrgID                                     int64
		status, originPort, destPort, carrierSCAC         string
		carrierName                                       string
		vesselName, voyageNumber                          sql.NullString
		bookingNumber, mblNumber, hblNumber               sql.NullString
		bookingID, rfqID, quoteID, customerID             sql.NullInt64
		containerNumbersJSON                              sql.NullString
		closureStatus                                     string
		etd, eta                                          sql.NullTime
		createdAt, updatedAt                              time.Time
		rfqNumber, customerName                           sql.NullString
	)

	err := s.db.QueryRowContext(ctx, shQuery, shipmentID, orgID).Scan(
		&shID, &shOrgID, &status, &originPort, &destPort, &carrierSCAC,
		&carrierName,
		&vesselName, &voyageNumber, &bookingNumber, &bookingID, &mblNumber, &hblNumber,
		&containerNumbersJSON, &rfqID, &quoteID, &etd, &eta, &closureStatus,
		&createdAt, &updatedAt,
		&rfqNumber,
		&customerID,
		&customerName,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("shipment %d not found in organization %d", shipmentID, orgID)
	} else if err != nil {
		return nil, fmt.Errorf("failed to query shipment %d: %w", shipmentID, err)
	}

	// Parse containers
	var containerList []string
	if containerNumbersJSON.Valid && containerNumbersJSON.String != "" {
		_ = json.Unmarshal([]byte(containerNumbersJSON.String), &containerList)
	}

	identity := &ShipmentIdentitySummary{
		ShipmentID:       shID,
		OrgID:            shOrgID,
		ShipmentNumber:   fmt.Sprintf("SH-%d", shID),
		OriginPort:       originPort,
		DestinationPort:  destPort,
		TransportMode:    "OCEAN", // Default ocean for SCAC-based container freight
		CarrierSCAC:      carrierSCAC,
		CarrierName:      carrierName,
		Status:           status,
		ClosureStatus:    closureStatus,
		ContainerNumbers: containerList,
		ContainerCount:   len(containerList),
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}

	if vesselName.Valid {
		identity.VesselName = vesselName.String
	}
	if voyageNumber.Valid {
		identity.VoyageNumber = voyageNumber.String
	}
	if bookingNumber.Valid {
		identity.BookingNumber = bookingNumber.String
	}
	if bookingID.Valid {
		bID := bookingID.Int64
		identity.BookingID = &bID
	}
	if mblNumber.Valid {
		identity.MBLNumber = mblNumber.String
	}
	if hblNumber.Valid {
		identity.HBLNumber = hblNumber.String
	}
	if rfqID.Valid {
		rID := rfqID.Int64
		identity.RFQID = &rID
	}
	if rfqNumber.Valid {
		identity.RFQNumber = rfqNumber.String
	}
	if quoteID.Valid {
		qID := quoteID.Int64
		identity.QuoteID = &qID
	}
	if customerID.Valid {
		cID := customerID.Int64
		identity.CustomerID = &cID
	}
	if customerName.Valid {
		identity.CustomerName = customerName.String
	}
	if etd.Valid {
		t := etd.Time
		identity.ETD = &t
	}
	if eta.Valid {
		t := eta.Time
		identity.ETA = &t
	}

	// 2. Fetch Milestones
	mQuery := `
		SELECT id, milestone_code, description, planned_date, actual_date, status, location, notes
		FROM shipment_milestones
		WHERE shipment_id = ?
		ORDER BY id ASC
	`
	mRows, err := s.db.QueryContext(ctx, mQuery, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query milestones for shipment %d: %w", shipmentID, err)
	}
	defer mRows.Close()

	var (
		milestoneItems             []*ShipmentMilestoneItem
		completedCount             int
		overdueCount               int
		pendingCount               int
		delayedMilestonesCount     int
		missingDateCount           int
		lastCompletedCode          string
		lastCompletedDate          *time.Time
		nextExpectedCode           string
		nextExpectedDate           *time.Time
		latestMilestoneActivity    time.Time = updatedAt
		criticalMilestoneRisks     []string
		actualDepartureDate        *time.Time
		actualArrivalDate          *time.Time
	)

	for mRows.Next() {
		var (
			mID                                    int64
			mCode, mStatus                         string
			mDesc, mLoc, mNotes                    sql.NullString
			mPlanned, mActual                      sql.NullTime
		)

		if err := mRows.Scan(&mID, &mCode, &mDesc, &mPlanned, &mActual, &mStatus, &mLoc, &mNotes); err != nil {
			continue
		}

		item := &ShipmentMilestoneItem{
			ID:            mID,
			MilestoneCode: mCode,
			Status:        mStatus,
		}
		if mDesc.Valid {
			item.Description = mDesc.String
		}
		if mLoc.Valid {
			item.Location = mLoc.String
		}
		if mNotes.Valid {
			item.Notes = mNotes.String
		}
		if mPlanned.Valid {
			t := mPlanned.Time
			item.PlannedDate = &t
		}
		if mActual.Valid {
			t := mActual.Time
			item.ActualDate = &t
			if t.After(latestMilestoneActivity) {
				latestMilestoneActivity = t
			}
		}

		// Track specific actual events
		upperCode := strings.ToUpper(mCode)
		if (upperCode == "DEPARTED" || upperCode == "DEPARTURE") && item.ActualDate != nil {
			actualDepartureDate = item.ActualDate
		}
		if (upperCode == "ARRIVAL" || upperCode == "ARRIVED" || upperCode == "DELIVERED") && item.ActualDate != nil {
			actualArrivalDate = item.ActualDate
		}

		// Milestone Evaluation
		if item.Status == "COMPLETED" {
			completedCount++
			lastCompletedCode = item.MilestoneCode
			lastCompletedDate = item.ActualDate

			// Check if completed late
			if item.PlannedDate != nil && item.ActualDate != nil && item.ActualDate.After(*item.PlannedDate) {
				diff := item.ActualDate.Sub(*item.PlannedDate).Hours()
				if diff > 1.0 { // tolerance 1 hr
					item.IsDelayed = true
					item.DelayHours = math.Round(diff*10) / 10
					delayedMilestonesCount++
				}
			}
		} else {
			pendingCount++
			// If not completed and planned date is in the past, it's overdue
			if item.PlannedDate != nil && item.PlannedDate.Before(now) {
				overdueCount++
				item.IsDelayed = true
				diff := now.Sub(*item.PlannedDate).Hours()
				item.DelayHours = math.Round(diff*10) / 10
				delayedMilestonesCount++

				if upperCode == "DEPARTED" || upperCode == "ARRIVAL" || upperCode == "CUSTOMS_CLEARANCE" {
					criticalMilestoneRisks = append(criticalMilestoneRisks, fmt.Sprintf("Milestone %s is overdue by %.1f hours", item.MilestoneCode, item.DelayHours))
				}
			}

			// Track next expected milestone
			if nextExpectedCode == "" {
				nextExpectedCode = item.MilestoneCode
				nextExpectedDate = item.PlannedDate
			}
		}

		if item.PlannedDate == nil && item.ActualDate == nil {
			missingDateCount++
		}

		milestoneItems = append(milestoneItems, item)
	}

	identity.ActualDeparture = actualDepartureDate
	identity.ActualArrival = actualArrivalDate

	totalMilestones := len(milestoneItems)
	var completionRate float64
	if totalMilestones > 0 {
		completionRate = math.Round((float64(completedCount)/float64(totalMilestones))*1000) / 10
	}

	// Time since last update
	hoursSinceUpdate := math.Round(now.Sub(latestMilestoneActivity).Hours()*10) / 10
	staleTrackingFlag := false
	// For active shipments, if no tracking update in > 48 hours, flag as stale
	if (status == "IN_TRANSIT" || status == "DEPARTED" || status == "BOOKED") && hoursSinceUpdate > 48.0 {
		staleTrackingFlag = true
	}

	milestoneIntel := &ShipmentMilestoneIntelligence{
		TotalMilestones:             totalMilestones,
		CompletedMilestones:         completedCount,
		PendingMilestones:           pendingCount,
		OverdueMilestones:           overdueCount,
		UpcomingMilestones:          pendingCount - overdueCount,
		MilestoneCompletionRate:     completionRate,
		LastCompletedMilestoneCode:  lastCompletedCode,
		LastCompletedMilestoneDate:  lastCompletedDate,
		NextExpectedMilestoneCode:   nextExpectedCode,
		NextExpectedMilestoneDate:   nextExpectedDate,
		TimeSinceLastUpdateHours:    &hoursSinceUpdate,
		NumberOfDelayedMilestones:   delayedMilestonesCount,
		MissingDateMilestonesCount:  missingDateCount,
		StaleTrackingFlag:           staleTrackingFlag,
		CriticalMilestoneRisks:      criticalMilestoneRisks,
		MilestonesList:              milestoneItems,
	}

	// 3. Fetch Exceptions
	eQuery := `
		SELECT id, exception_type, severity, status, title, description, resolved, resolved_at, resolution_notes, created_at
		FROM shipment_exceptions
		WHERE shipment_id = ? AND org_id = ?
		ORDER BY created_at DESC
	`
	eRows, err := s.db.QueryContext(ctx, eQuery, shipmentID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query exceptions for shipment %d: %w", shipmentID, err)
	}
	defer eRows.Close()

	var (
		exceptionItems         []*ShipmentExceptionItem
		openExceptions         int
		resolvedExceptions     int
		criticalExceptions     int
		highExceptions         int
		oldestUnresolvedTitle  string
		oldestUnresolvedHours  *float64
		oldestCreated          time.Time = now
		categoriesMap          = make(map[string]int)
	)

	for eRows.Next() {
		var (
			eID                                         int64
			eType, eSeverity, eStatus, eTitle           string
			eDesc, eResNotes                            sql.NullString
			eResolved                                   bool
			eResolvedAt                                 sql.NullTime
			eCreatedAt                                  time.Time
		)

		if err := eRows.Scan(&eID, &eType, &eSeverity, &eStatus, &eTitle, &eDesc, &eResolved, &eResolvedAt, &eResNotes, &eCreatedAt); err != nil {
			continue
		}

		hoursOpen := math.Round(now.Sub(eCreatedAt).Hours()*10) / 10
		item := &ShipmentExceptionItem{
			ID:            eID,
			ExceptionType: eType,
			Severity:      eSeverity,
			Status:        eStatus,
			Title:         eTitle,
			Resolved:      eResolved,
			HoursOpen:     hoursOpen,
			CreatedAt:     eCreatedAt,
		}
		if eDesc.Valid {
			item.Description = eDesc.String
		}
		if eResNotes.Valid {
			item.ResolutionNotes = eResNotes.String
		}
		if eResolvedAt.Valid {
			t := eResolvedAt.Time
			item.ResolvedAt = &t
		}

		categoriesMap[eType]++

		if !eResolved && (eStatus == "OPEN" || eStatus == "ACKNOWLEDGED" || eStatus == "") {
			openExceptions++
			sevUpper := strings.ToUpper(eSeverity)
			if sevUpper == "CRITICAL" {
				criticalExceptions++
			} else if sevUpper == "HIGH" {
				highExceptions++
			}

			if eCreatedAt.Before(oldestCreated) {
				oldestCreated = eCreatedAt
				oldestUnresolvedTitle = eTitle
				hrs := math.Round(now.Sub(eCreatedAt).Hours()*10) / 10
				oldestUnresolvedHours = &hrs
			}
		} else {
			resolvedExceptions++
		}

		exceptionItems = append(exceptionItems, item)
	}

	exceptionIntel := &ShipmentExceptionIntelligence{
		TotalExceptions:           len(exceptionItems),
		OpenExceptions:            openExceptions,
		ResolvedExceptions:        resolvedExceptions,
		CriticalExceptions:        criticalExceptions,
		HighSeverityExceptions:    highExceptions,
		OldestUnresolvedTitle:     oldestUnresolvedTitle,
		OldestUnresolvedHoursOpen: oldestUnresolvedHours,
		ExceptionsByCategory:      categoriesMap,
		ExceptionsList:            exceptionItems,
	}

	// 4. Operational Performance Analytics
	perf := &ShipmentOperationalPerformance{
		NumberOfDelays: delayedMilestonesCount,
	}

	// Departure variance
	if identity.ETD != nil {
		if actualDepartureDate != nil {
			diff := actualDepartureDate.Sub(*identity.ETD).Hours()
			diffRound := math.Round(diff*10) / 10
			perf.DepartureVarianceHours = &diffRound
			onTime := diff <= 2.0 // 2 hours grace
			perf.OnTimeDeparture = &onTime
		} else if now.After(*identity.ETD) && (status == "BOOKED" || status == "BOOKING_PENDING") {
			diff := now.Sub(*identity.ETD).Hours()
			diffRound := math.Round(diff*10) / 10
			perf.DepartureVarianceHours = &diffRound
			onTime := false
			perf.OnTimeDeparture = &onTime
		}
	}

	// Arrival variance
	if identity.ETA != nil {
		if actualArrivalDate != nil {
			diff := actualArrivalDate.Sub(*identity.ETA).Hours()
			diffRound := math.Round(diff*10) / 10
			perf.ArrivalVarianceHours = &diffRound
			onTime := diff <= 4.0 // 4 hours grace
			perf.OnTimeArrival = &onTime
		} else if now.After(*identity.ETA) && status != "DELIVERED" && status != "ARRIVED" {
			diff := now.Sub(*identity.ETA).Hours()
			diffRound := math.Round(diff*10) / 10
			perf.ArrivalVarianceHours = &diffRound
			onTime := false
			perf.OnTimeArrival = &onTime
		}
	}

	// Transit duration
	if identity.ETD != nil && identity.ETA != nil {
		plannedDays := math.Round(identity.ETA.Sub(*identity.ETD).Hours()/24.0*10) / 10
		perf.PlannedTransitDays = &plannedDays
	}
	if actualDepartureDate != nil && actualArrivalDate != nil {
		actualDays := math.Round(actualArrivalDate.Sub(*actualDepartureDate).Hours()/24.0*10) / 10
		perf.ActualTransitDays = &actualDays
	}

	// Delay duration
	if perf.ArrivalVarianceHours != nil && *perf.ArrivalVarianceHours > 0 {
		perf.DelayDurationHours = perf.ArrivalVarianceHours
	} else if perf.DepartureVarianceHours != nil && *perf.DepartureVarianceHours > 0 {
		perf.DelayDurationHours = perf.DepartureVarianceHours
	}

	// Cycle time
	cycleDays := math.Round(now.Sub(createdAt).Hours()/24.0*10) / 10
	perf.CycleTimeDays = &cycleDays

	// Tracking freshness rating
	if hoursSinceUpdate <= 6.0 {
		perf.CarrierUpdateFreshness = "LIVE"
	} else if hoursSinceUpdate <= 24.0 {
		perf.CarrierUpdateFreshness = "RECENT"
	} else if hoursSinceUpdate <= 72.0 {
		perf.CarrierUpdateFreshness = "STALE"
	} else {
		perf.CarrierUpdateFreshness = "UNAVAILABLE"
	}

	// 5. Risk Indicators & Scoring
	risk := &ShipmentRiskIndicators{
		ActiveRiskFactors:  make([]string, 0),
		MissingDataReasons: make([]string, 0),
	}

	riskScore := 0

	// Delayed?
	if (perf.DepartureVarianceHours != nil && *perf.DepartureVarianceHours > 12.0) ||
		(perf.ArrivalVarianceHours != nil && *perf.ArrivalVarianceHours > 12.0) ||
		delayedMilestonesCount > 0 {
		risk.ShipmentDelayed = true
		riskScore += 25
		risk.ActiveRiskFactors = append(risk.ActiveRiskFactors, fmt.Sprintf("Shipment delayed (%d delayed milestone(s))", delayedMilestonesCount))
	}

	// Critical milestone overdue?
	if len(criticalMilestoneRisks) > 0 {
		risk.CriticalMilestoneOverdue = true
		riskScore += 30
		risk.ActiveRiskFactors = append(risk.ActiveRiskFactors, criticalMilestoneRisks...)
	}

	// Unresolved high/critical exceptions?
	if criticalExceptions > 0 {
		risk.UnresolvedHighException = true
		riskScore += 40
		risk.ActiveRiskFactors = append(risk.ActiveRiskFactors, fmt.Sprintf("%d unresolved CRITICAL operational exception(s)", criticalExceptions))
	} else if highExceptions > 0 {
		risk.UnresolvedHighException = true
		riskScore += 25
		risk.ActiveRiskFactors = append(risk.ActiveRiskFactors, fmt.Sprintf("%d unresolved HIGH severity exception(s)", highExceptions))
	} else if openExceptions > 0 {
		riskScore += 10
		risk.ActiveRiskFactors = append(risk.ActiveRiskFactors, fmt.Sprintf("%d open exception(s) pending resolution", openExceptions))
	}

	// Stale tracking?
	if staleTrackingFlag {
		risk.StaleTrackingData = true
		riskScore += 20
		risk.ActiveRiskFactors = append(risk.ActiveRiskFactors, fmt.Sprintf("No carrier telemetry update for %.1f hours", hoursSinceUpdate))
	}

	// Missing data checks
	if identity.ETA == nil {
		risk.ETAMissingOrStale = true
		risk.MissingScheduleData = true
		riskScore += 15
		risk.MissingDataReasons = append(risk.MissingDataReasons, "Planned ETA date not recorded by carrier or operator")
	}
	if identity.ETD == nil {
		risk.MissingScheduleData = true
		risk.MissingDataReasons = append(risk.MissingDataReasons, "Planned ETD date not recorded")
	}
	if identity.CarrierSCAC == "" {
		risk.MissingCarrierInfo = true
		riskScore += 20
		risk.MissingDataReasons = append(risk.MissingDataReasons, "Carrier SCAC unassigned")
	}
	if totalMilestones == 0 {
		risk.DataIncomplete = true
		riskScore += 10
		risk.MissingDataReasons = append(risk.MissingDataReasons, "No milestone tracking schedule recorded")
	}

	if riskScore > 100 {
		riskScore = 100
	}
	risk.RiskScore = riskScore

	// Rating
	if riskScore >= 70 || criticalExceptions > 0 {
		risk.OverallRiskRating = "CRITICAL"
	} else if riskScore >= 40 || highExceptions > 0 || risk.CriticalMilestoneOverdue {
		risk.OverallRiskRating = "HIGH"
	} else if riskScore >= 20 || risk.ShipmentDelayed || risk.StaleTrackingData {
		risk.OverallRiskRating = "MODERATE"
	} else {
		risk.OverallRiskRating = "LOW"
	}

	// 6. Grounded AI Operational Summary
	aiSummary := s.synthesizeShipmentOperationsAISummary(identity, milestoneIntel, exceptionIntel, perf, risk)

	// 7. Grounded Traceability Observations
	var observations []*ShipmentObservation

	observations = append(observations, &ShipmentObservation{
		Category: "STATUS",
		Severity: "INFO",
		Message:  fmt.Sprintf("Shipment SH-%d is currently in status %s with milestone progress of %.1f%%.", shID, status, completionRate),
		Evidence: fmt.Sprintf("Shipment #%d status=%s, %d/%d milestones completed", shID, status, completedCount, totalMilestones),
	})

	if openExceptions > 0 {
		sev := "WARNING"
		if criticalExceptions > 0 {
			sev = "CRITICAL"
		}
		observations = append(observations, &ShipmentObservation{
			Category: "EXCEPTION",
			Severity: sev,
			Message:  fmt.Sprintf("%d open exception(s) require intervention. Oldest: %s.", openExceptions, oldestUnresolvedTitle),
			Evidence: fmt.Sprintf("Exceptions table: %d open, %d critical, %d high", openExceptions, criticalExceptions, highExceptions),
		})
	}

	if perf.DepartureVarianceHours != nil && *perf.DepartureVarianceHours > 0 {
		observations = append(observations, &ShipmentObservation{
			Category: "SCHEDULE",
			Severity: "WARNING",
			Message:  fmt.Sprintf("Departure schedule experienced %.1f hours variance from planned ETD.", *perf.DepartureVarianceHours),
			Evidence: fmt.Sprintf("ETD: %v, Actual/Now: %.1f hrs diff", identity.ETD, *perf.DepartureVarianceHours),
		})
	}

	if risk.OverallRiskRating == "CRITICAL" || risk.OverallRiskRating == "HIGH" {
		observations = append(observations, &ShipmentObservation{
			Category: "RISK",
			Severity: "CRITICAL",
			Message:  fmt.Sprintf("Operational health flagged as %s (Risk Score: %d/100). Primary causes: %s", risk.OverallRiskRating, risk.RiskScore, strings.Join(risk.ActiveRiskFactors, "; ")),
			Evidence: fmt.Sprintf("Risk factors evaluated deterministically against live milestone & exception records"),
		})
	}

	return &Shipment360OperationsIntelligence{
		ShipmentID:             shID,
		OrgID:                  shOrgID,
		CorrelationID:          correlationID,
		DataFreshnessTimestamp: now,
		Identity:               identity,
		Milestones:             milestoneIntel,
		Exceptions:             exceptionIntel,
		Performance:            perf,
		RiskIndicators:         risk,
		AISummary:              aiSummary,
		GroundedObservations:   observations,
	}, nil
}

// synthesizeShipmentOperationsAISummary builds a strictly read-only, grounded operational assessment.
func (s *defaultService) synthesizeShipmentOperationsAISummary(
	id *ShipmentIdentitySummary,
	m *ShipmentMilestoneIntelligence,
	e *ShipmentExceptionIntelligence,
	p *ShipmentOperationalPerformance,
	r *ShipmentRiskIndicators,
) *ShipmentAIOperationsSummary {
	var citations []string
	citations = append(citations, fmt.Sprintf("[Shipment: #%d]", id.ShipmentID))
	if id.CarrierSCAC != "" {
		citations = append(citations, fmt.Sprintf("[Carrier: %s]", id.CarrierSCAC))
	}
	if id.BookingNumber != "" {
		citations = append(citations, fmt.Sprintf("[Booking: %s]", id.BookingNumber))
	}

	// 1. Executive Summary
	exec := fmt.Sprintf(
		"Shipment SH-%d (%s to %s) via %s is %s with %.1f%% milestone completion (%d of %d milestones executed).",
		id.ShipmentID,
		id.OriginPort,
		id.DestinationPort,
		id.CarrierName,
		strings.ReplaceAll(id.Status, "_", " "),
		m.MilestoneCompletionRate,
		m.CompletedMilestones,
		m.TotalMilestones,
	)

	// 2. Status & Milestone Explanation
	var statusExpl string
	if m.NextExpectedMilestoneCode != "" {
		statusExpl = fmt.Sprintf("Next scheduled milestone is %s.", m.NextExpectedMilestoneCode)
		if m.NextExpectedMilestoneDate != nil {
			statusExpl += fmt.Sprintf(" Planned date: %s.", m.NextExpectedMilestoneDate.Format("Jan 02 15:04 UTC"))
		}
	} else if m.CompletedMilestones == m.TotalMilestones && m.TotalMilestones > 0 {
		statusExpl = "All recorded milestones are completed. Shipment is ready for final delivery or archiving."
	} else {
		statusExpl = "No upcoming milestone scheduled in system tracking ledger."
	}

	var mileExpl string
	if m.NumberOfDelayedMilestones > 0 {
		mileExpl = fmt.Sprintf("%d milestone(s) completed late or currently overdue.", m.NumberOfDelayedMilestones)
	} else {
		mileExpl = "All completed milestones were recorded within scheduled window."
	}

	// 3. Operational Risks
	var likelyRisks string
	if len(r.ActiveRiskFactors) > 0 {
		likelyRisks = strings.Join(r.ActiveRiskFactors, ". ") + "."
	} else {
		likelyRisks = "No significant operational blockers or schedule delays detected."
	}

	// 4. Delay Causes Evidence
	var delayEvidence string
	if p.DelayDurationHours != nil && *p.DelayDurationHours > 0 {
		delayEvidence = fmt.Sprintf("Schedule variance of %.1f hours observed against carrier schedule.", *p.DelayDurationHours)
	}

	// 5. Exception Prioritization
	var excPrior string
	if e.OpenExceptions > 0 {
		excPrior = fmt.Sprintf("%d open exception(s) require operations review. Priority: %d Critical, %d High. Oldest issue: %s.",
			e.OpenExceptions, e.CriticalExceptions, e.HighSeverityExceptions, e.OldestUnresolvedTitle)
		for _, ex := range e.ExceptionsList {
			if !ex.Resolved {
				citations = append(citations, fmt.Sprintf("[Exception: #%d (%s)]", ex.ID, ex.Severity))
			}
		}
	} else {
		excPrior = "Zero open operational exceptions recorded."
	}

	// 6. Actionable Attention Items
	var attentionItems []string
	if e.CriticalExceptions > 0 {
		attentionItems = append(attentionItems, "Escalate critical exception to carrier line operations manager.")
	}
	if r.StaleTrackingData {
		attentionItems = append(attentionItems, "Request fresh telemetry/EDI tracking update from carrier.")
	}
	if r.CriticalMilestoneOverdue {
		attentionItems = append(attentionItems, "Confirm actual departure/arrival status with local port agent.")
	}
	if len(attentionItems) == 0 {
		attentionItems = append(attentionItems, "Routine monitoring — shipment is progressing as planned.")
	}

	// 7. Operator Inquiries
	var inquiries []string
	if r.ShipmentDelayed {
		inquiries = append(inquiries, "Inquire with carrier for updated vessel schedule and revised destination ETA.")
	}
	if e.OpenExceptions > 0 {
		inquiries = append(inquiries, "Confirm whether cargo documentation or customs inspection is required to release hold.")
	}
	if id.ETA == nil {
		inquiries = append(inquiries, "Contact ocean carrier to obtain firm vessel ETA.")
	}

	confidence := "HIGH"
	if m.TotalMilestones == 0 || r.DataIncomplete {
		confidence = "LOW"
	} else if r.MissingScheduleData || r.StaleTrackingData {
		confidence = "MEDIUM"
	}

	return &ShipmentAIOperationsSummary{
		ExecutiveSummary:             exec,
		CurrentStatusExplanation:     statusExpl,
		MilestoneProgressExplanation: mileExpl,
		LikelyOperationalRisks:       likelyRisks,
		DelayCausesEvidence:          delayEvidence,
		ExceptionPrioritization:      excPrior,
		ActionableAttentionItems:     attentionItems,
		SuggestedOperatorInquiries:   inquiries,
		ConfidenceScore:              confidence,
		VerifiableSourceCitations:    citations,
	}
}

// GetOrgOperationsSummary computes an aggregate operational health overview across an entire organization.
func (s *defaultService) GetOrgOperationsSummary(ctx context.Context, orgID int64, correlationID string, userID int64) (*OrgOperationsSummary, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization ID: %d", orgID)
	}
	if correlationID == "" {
		correlationID = fmt.Sprintf("org-ops-%d-%d", orgID, time.Now().UnixNano())
	}

	now := time.Now().UTC()

	// Shipments summary
	q := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status NOT IN ('DELIVERED', 'CANCELLED') THEN 1 END) as active,
			COUNT(CASE WHEN status = 'DELIVERED' THEN 1 END) as completed
		FROM shipments
		WHERE org_id = ?
	`
	var total, active, completed int
	if err := s.db.QueryRowContext(ctx, q, orgID).Scan(&total, &active, &completed); err != nil {
		return nil, fmt.Errorf("failed to query organization shipments count: %w", err)
	}

	// Exceptions summary
	eq := `
		SELECT 
			COUNT(CASE WHEN resolved = 0 AND status != 'RESOLVED' THEN 1 END) as open_exc,
			COUNT(CASE WHEN resolved = 0 AND status != 'RESOLVED' AND UPPER(severity) = 'CRITICAL' THEN 1 END) as crit_exc
		FROM shipment_exceptions
		WHERE org_id = ?
	`
	var openExc, critExc int
	if err := s.db.QueryRowContext(ctx, eq, orgID).Scan(&openExc, &critExc); err != nil {
		return nil, fmt.Errorf("failed to query organization exceptions count: %w", err)
	}

	// Delayed shipments
	dq := `
		SELECT COUNT(DISTINCT s.id)
		FROM shipments s
		LEFT JOIN shipment_milestones m ON s.id = m.shipment_id
		WHERE s.org_id = ?
		  AND s.status NOT IN ('DELIVERED', 'CANCELLED')
		  AND ((s.eta IS NOT NULL AND s.eta < ?) OR (m.status = 'PLANNED' AND m.planned_date < ?))
	`
	var delayedCount int
	_ = s.db.QueryRowContext(ctx, dq, orgID, now, now).Scan(&delayedCount)

	// Stale tracking shipments (> 48h since updated_at on active shipments)
	staleQ := `
		SELECT COUNT(*)
		FROM shipments
		WHERE org_id = ?
		  AND status IN ('IN_TRANSIT', 'DEPARTED', 'BOOKED')
		  AND updated_at < DATE_SUB(?, INTERVAL 48 HOUR)
	`
	var staleCount int
	_ = s.db.QueryRowContext(ctx, staleQ, orgID, now).Scan(&staleCount)

	// On-time percentage where data exists
	var onTimePct *float64
	if total > 0 {
		rate := math.Max(0, math.Min(100.0, 100.0-float64(delayedCount*100)/float64(math.Max(1, float64(active)))))
		rateRound := math.Round(rate*10) / 10
		onTimePct = &rateRound
	}

	riskDist := map[string]int{
		"CRITICAL": critExc,
		"HIGH":     delayedCount,
		"MODERATE": staleCount,
		"LOW":      int(math.Max(0, float64(active-critExc-delayedCount-staleCount))),
	}

	return &OrgOperationsSummary{
		OrgID:                  orgID,
		CorrelationID:          correlationID,
		DataFreshnessTimestamp: now,
		TotalShipments:         total,
		ActiveShipments:        active,
		CompletedShipments:     completed,
		DelayedShipments:       delayedCount,
		TotalOpenExceptions:    openExc,
		CriticalOpenExceptions: critExc,
		StaleTrackingShipments: staleCount,
		OnTimeRatePercentage:   onTimePct,
		RiskDistribution:       riskDist,
	}, nil
}
