package predictions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// DeterministicShipmentData encapsulates verified operational facts for shipment ETA modeling
type DeterministicShipmentData struct {
	ShipmentID              int64
	OrgID                   int64
	Status                  string
	OriginPort              string
	DestinationPort         string
	CarrierSCAC             string
	VesselName              string
	VoyageNumber            string
	AuthoritativeETA        *time.Time
	PlannedETD              *time.Time
	ActualDeparture         *time.Time
	DepartureDelayHours     float64
	LatestMilestoneCode     string
	LatestMilestoneLocation string
	LatestMilestoneTime     *time.Time
	MilestoneDelayHours     float64
	OpenExceptions          []map[string]interface{}
	TrackingTelemetry       map[string]interface{}
	PortCongestionIndex     float64
	IsDelivered             bool
	MissingAuthoritativeETA bool
	ContextMap              map[string]interface{}
}

// ShipmentDataProvider prepares verified facts using deterministic calculations
type ShipmentDataProvider struct {
	db *sql.DB
}

// NewShipmentDataProvider initializes the deterministic shipment data provider
func NewShipmentDataProvider(db *sql.DB) *ShipmentDataProvider {
	return &ShipmentDataProvider{db: db}
}

// FetchDeterministicShipmentData gathers real, verified operational records without AI speculation
func (p *ShipmentDataProvider) FetchDeterministicShipmentData(ctx context.Context, orgID int64, shipmentID int64) (*DeterministicShipmentData, error) {
	// 1. Fetch Shipment Master Record
	queryShipment := `
		SELECT 
			id, org_id, status, origin_port, destination_port, carrier_scac, 
			COALESCE(vessel_name, ''), COALESCE(voyage_number, ''), etd, eta
		FROM shipments
		WHERE org_id = ? AND id = ?
	`
	data := &DeterministicShipmentData{
		ShipmentID: shipmentID,
		OrgID:      orgID,
	}

	var vesselName, voyageNum string
	var etd, eta sql.NullTime

	err := p.db.QueryRowContext(ctx, queryShipment, orgID, shipmentID).Scan(
		&data.ShipmentID, &data.OrgID, &data.Status, &data.OriginPort, &data.DestinationPort, &data.CarrierSCAC,
		&vesselName, &voyageNum, &etd, &eta,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("shipment %d not found for organization %d", shipmentID, orgID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shipment: %w", err)
	}

	data.VesselName = vesselName
	data.VoyageNumber = voyageNum
	if etd.Valid {
		data.PlannedETD = &etd.Time
	}
	if eta.Valid {
		data.AuthoritativeETA = &eta.Time
	} else {
		data.MissingAuthoritativeETA = true
	}

	statusUpper := strings.ToUpper(data.Status)
	if statusUpper == "ARRIVED" || statusUpper == "DELIVERED" || statusUpper == "COMPLETED" {
		data.IsDelivered = true
	}

	// 2. Fetch Milestones
	queryMilestones := `
		SELECT milestone_code, COALESCE(description, ''), planned_date, actual_date, status, COALESCE(location, ''), updated_at
		FROM shipment_milestones
		WHERE shipment_id = ?
		ORDER BY COALESCE(actual_date, planned_date, updated_at) ASC
	`
	rowsM, err := p.db.QueryContext(ctx, queryMilestones, shipmentID)
	if err == nil {
		defer rowsM.Close()
		for rowsM.Next() {
			var code, desc, loc string
			var pDate, aDate sql.NullTime
			var mStatus string
			var uAt time.Time
			if err := rowsM.Scan(&code, &desc, &pDate, &aDate, &mStatus, &loc, &uAt); err == nil {
				if aDate.Valid {
					data.LatestMilestoneCode = code
					data.LatestMilestoneLocation = loc
					data.LatestMilestoneTime = &aDate.Time
				}

				codeUpper := strings.ToUpper(code)
				if codeUpper == "DEPARTED" && aDate.Valid {
					data.ActualDeparture = &aDate.Time
					if data.PlannedETD != nil {
						diffHours := aDate.Time.Sub(*data.PlannedETD).Hours()
						if diffHours > 0 {
							data.DepartureDelayHours = diffHours
						}
					}
				}

				// Check for delay notices in milestone description
				descLower := strings.ToLower(desc)
				if strings.Contains(descLower, "delay") || strings.Contains(descLower, "typhoon") || strings.Contains(descLower, "congestion") {
					if strings.Contains(descLower, "10-day") || strings.Contains(descLower, "10 day") {
						data.MilestoneDelayHours = 240.0
					} else if strings.Contains(descLower, "5-day") || strings.Contains(descLower, "5 day") {
						data.MilestoneDelayHours = 120.0
					} else if strings.Contains(descLower, "2-day") || strings.Contains(descLower, "2 day") {
						data.MilestoneDelayHours = 48.0
					} else if data.MilestoneDelayHours == 0 {
						data.MilestoneDelayHours = 24.0
					}
				}

				if (codeUpper == "ARRIVED" || codeUpper == "ARRIVAL") && (mStatus == "COMPLETED" || aDate.Valid) {
					data.IsDelivered = true
				}
			}
		}
	}

	// 3. Fetch Open Exceptions
	queryExceptions := `
		SELECT id, exception_type, severity, COALESCE(title, ''), COALESCE(description, ''), created_at
		FROM shipment_exceptions
		WHERE org_id = ? AND shipment_id = ? AND status != 'RESOLVED'
		ORDER BY created_at DESC
	`
	rowsE, err := p.db.QueryContext(ctx, queryExceptions, orgID, shipmentID)
	data.OpenExceptions = make([]map[string]interface{}, 0)
	data.PortCongestionIndex = 1.0

	if err == nil {
		defer rowsE.Close()
		for rowsE.Next() {
			var eID int64
			var eType, eSev, eTitle, eDesc string
			var eCreated time.Time
			if err := rowsE.Scan(&eID, &eType, &eSev, &eTitle, &eDesc, &eCreated); err == nil {
				data.OpenExceptions = append(data.OpenExceptions, map[string]interface{}{
					"id":          eID,
					"type":        eType,
					"severity":    eSev,
					"title":       eTitle,
					"description": eDesc,
					"created_at":  eCreated.Format(time.RFC3339),
				})
				if strings.Contains(strings.ToUpper(eType), "PORT") || strings.Contains(strings.ToUpper(eTitle), "CONGESTION") {
					data.PortCongestionIndex = 1.25
				}
			}
		}
	}

	// 4. Fetch Latest Tracking Telemetry Position
	queryPos := `
		SELECT vessel_name, latitude, longitude, speed_knots, heading_degrees, location_name, data_freshness, recorded_at
		FROM shipment_tracking_positions
		WHERE org_id = ? AND shipment_id = ?
		ORDER BY recorded_at DESC
		LIMIT 1
	`
	var posVessel, posLoc, posFreshness string
	var lat, lon, speed, heading float64
	var recAt time.Time
	errPos := p.db.QueryRowContext(ctx, queryPos, orgID, shipmentID).Scan(
		&posVessel, &lat, &lon, &speed, &heading, &posLoc, &posFreshness, &recAt,
	)
	data.TrackingTelemetry = make(map[string]interface{})
	if errPos == nil {
		ageSeconds := int(time.Since(recAt).Seconds())
		if ageSeconds < 0 {
			ageSeconds = 0
		}
		data.TrackingTelemetry = map[string]interface{}{
			"vessel_name":       posVessel,
			"latitude":          lat,
			"longitude":         lon,
			"speed_knots":       speed,
			"heading_degrees":   heading,
			"location_name":     posLoc,
			"freshness_status":  posFreshness,
			"freshness_seconds": ageSeconds,
			"recorded_at":       recAt.Format(time.RFC3339),
		}
	}

	// 5. Construct Grounded Context Map for Python Engine
	nowStr := time.Now().UTC().Format(time.RFC3339)
	ctxMap := map[string]interface{}{
		"shipment_id":           data.ShipmentID,
		"status":                data.Status,
		"origin_port":           data.OriginPort,
		"destination_port":      data.DestinationPort,
		"carrier_scac":          data.CarrierSCAC,
		"vessel_name":           data.VesselName,
		"voyage_number":         data.VoyageNumber,
		"departure_delay_hours": data.DepartureDelayHours,
		"milestone_delay_hours": data.MilestoneDelayHours,
		"open_exceptions":       data.OpenExceptions,
		"port_congestion_index": data.PortCongestionIndex,
		"tracking_telemetry":    data.TrackingTelemetry,
		"latest_milestone_code": data.LatestMilestoneCode,
	}

	if data.AuthoritativeETA != nil {
		ctxMap["authoritative_eta"] = data.AuthoritativeETA.Format("2006-01-02 15:04:05")
		ctxMap["eta"] = data.AuthoritativeETA.Format("2006-01-02 15:04:05")
	}
	if data.PlannedETD != nil {
		ctxMap["planned_etd"] = data.PlannedETD.Format("2006-01-02 15:04:05")
	}
	if data.ActualDeparture != nil {
		ctxMap["actual_departure"] = data.ActualDeparture.Format(time.RFC3339)
	}
	if data.LatestMilestoneTime != nil {
		ctxMap["last_milestone_time"] = data.LatestMilestoneTime.Format(time.RFC3339)
	} else {
		ctxMap["last_milestone_time"] = nowStr
	}
	if data.IsDelivered {
		ctxMap["actual_arrival"] = true
	}

	data.ContextMap = ctxMap
	return data, nil
}

// DeterministicDisruptionData encapsulates verified operational facts for disruption forecasting
type DeterministicDisruptionData struct {
	ShipmentID                 int64                    `json:"shipment_id"`
	OrgID                      int64                    `json:"org_id"`
	Status                     string                   `json:"status"`
	OriginPort                 string                   `json:"origin_port"`
	DestinationPort            string                   `json:"destination_port"`
	CarrierSCAC                string                   `json:"carrier_scac"`
	VesselName                 string                   `json:"vessel_name"`
	VoyageNumber               string                   `json:"voyage_number"`
	AuthoritativeETA           *time.Time               `json:"authoritative_eta,omitempty"`
	PlannedETD                 *time.Time               `json:"planned_etd,omitempty"`
	ActualDeparture            *time.Time               `json:"actual_departure,omitempty"`
	IsDelivered                bool                     `json:"is_delivered"`
	MissingAuthoritativeETA    bool                     `json:"missing_authoritative_eta"`
	MilestonesCount            int                      `json:"milestones_count"`
	OverdueMilestonesCount     int                      `json:"overdue_milestones_count"`
	OverdueMilestones          []map[string]interface{} `json:"overdue_milestones"`
	Milestones                 []map[string]interface{} `json:"milestones"`
	MilestoneVarianceHours     float64                  `json:"milestone_variance_hours"`
	LatestTrackingPing         *time.Time               `json:"latest_tracking_ping,omitempty"`
	TrackingAgeHours           float64                  `json:"tracking_age_hours"`
	ExistingExceptions         []map[string]interface{} `json:"existing_exceptions"`
	ActiveExceptionsCount      int                      `json:"active_exceptions_count"`
	DemurrageFreeDaysRemaining *int                     `json:"demurrage_free_days_remaining,omitempty"`
	CustomerCommitmentDate     *time.Time               `json:"customer_commitment_date,omitempty"`
	ContextMap                 map[string]interface{}   `json:"context_map"`
}

// FetchExceptionDisruptionData gathers real, verified operational records for exception forecasting
func (p *ShipmentDataProvider) FetchExceptionDisruptionData(ctx context.Context, orgID int64, shipmentID int64) (*DeterministicDisruptionData, error) {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	// 1. Fetch Shipment Master Record
	queryShipment := `
		SELECT 
			id, org_id, status, origin_port, destination_port, carrier_scac, 
			COALESCE(vessel_name, ''), COALESCE(voyage_number, ''), etd, eta
		FROM shipments
		WHERE org_id = ? AND id = ?
	`
	data := &DeterministicDisruptionData{
		ShipmentID:         shipmentID,
		OrgID:              orgID,
		Milestones:         make([]map[string]interface{}, 0),
		OverdueMilestones:  make([]map[string]interface{}, 0),
		ExistingExceptions: make([]map[string]interface{}, 0),
	}

	var vesselName, voyageNum string
	var etd, eta sql.NullTime

	err := p.db.QueryRowContext(ctx, queryShipment, orgID, shipmentID).Scan(
		&data.ShipmentID, &data.OrgID, &data.Status, &data.OriginPort, &data.DestinationPort, &data.CarrierSCAC,
		&vesselName, &voyageNum, &etd, &eta,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("shipment %d not found for organization %d", shipmentID, orgID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shipment: %w", err)
	}

	data.VesselName = vesselName
	data.VoyageNumber = voyageNum
	if etd.Valid {
		data.PlannedETD = &etd.Time
	}
	if eta.Valid {
		data.AuthoritativeETA = &eta.Time
	} else {
		data.MissingAuthoritativeETA = true
	}

	statusUpper := strings.ToUpper(data.Status)
	if statusUpper == "ARRIVED" || statusUpper == "DELIVERED" || statusUpper == "COMPLETED" {
		data.IsDelivered = true
	}

	// 2. Fetch Milestones
	queryMilestones := `
		SELECT milestone_code, COALESCE(description, ''), planned_date, actual_date, status, COALESCE(location, ''), updated_at
		FROM shipment_milestones
		WHERE shipment_id = ?
		ORDER BY COALESCE(actual_date, planned_date, updated_at) ASC
	`
	rowsM, errM := p.db.QueryContext(ctx, queryMilestones, shipmentID)
	var maxDelayHours float64
	if errM == nil {
		defer rowsM.Close()
		for rowsM.Next() {
			var code, desc, status, loc string
			var pDate, aDate sql.NullTime
			var upAt time.Time
			if err := rowsM.Scan(&code, &desc, &pDate, &aDate, &status, &loc, &upAt); err == nil {
				mMap := map[string]interface{}{
					"code":        code,
					"description": desc,
					"status":      status,
					"location":    loc,
				}
				if pDate.Valid {
					mMap["planned_date"] = pDate.Time.Format(time.RFC3339)
				}
				if aDate.Valid {
					mMap["actual_date"] = aDate.Time.Format(time.RFC3339)
				}

				data.Milestones = append(data.Milestones, mMap)
				data.MilestonesCount++

				// Check overdue milestones
				if status == "PLANNED" && pDate.Valid && now.After(pDate.Time) {
					data.OverdueMilestones = append(data.OverdueMilestones, mMap)
					data.OverdueMilestonesCount++
				}

				// Check actual delay variance
				if aDate.Valid && pDate.Valid && aDate.Time.After(pDate.Time) {
					varHours := aDate.Time.Sub(pDate.Time).Hours()
					if varHours > maxDelayHours {
						maxDelayHours = varHours
					}
				}

				// Check delay notices
				if code == "DELAY_NOTICE" {
					if aDate.Valid && pDate.Valid {
						diffHours := aDate.Time.Sub(pDate.Time).Hours()
						if diffHours > maxDelayHours {
							maxDelayHours = diffHours
						}
					} else if maxDelayHours < 72.0 {
						maxDelayHours = 72.0 // Bulletin notice standard slip
					}
				}
			}
		}
	}
	data.MilestoneVarianceHours = maxDelayHours

	// 3. Fetch Existing Confirmed Exceptions
	queryExceptions := `
		SELECT id, exception_type, severity, title, COALESCE(description, ''), status, created_at
		FROM shipment_exceptions
		WHERE shipment_id = ? AND (resolved = 0 OR status = 'OPEN')
		ORDER BY created_at DESC
	`
	rowsE, errE := p.db.QueryContext(ctx, queryExceptions, shipmentID)
	if errE == nil {
		defer rowsE.Close()
		for rowsE.Next() {
			var excID int64
			var excType, excSev, excTitle, excDesc, excStatus string
			var excCreated time.Time
			if err := rowsE.Scan(&excID, &excType, &excSev, &excTitle, &excDesc, &excStatus, &excCreated); err == nil {
				data.ExistingExceptions = append(data.ExistingExceptions, map[string]interface{}{
					"id":             excID,
					"exception_type": excType,
					"severity":       excSev,
					"title":          excTitle,
					"description":    excDesc,
					"status":         excStatus,
					"created_at":     excCreated.Format(time.RFC3339),
				})
				data.ActiveExceptionsCount++
			}
		}
	}

	// 4. Fetch Latest Vessel Telemetry / Ping
	queryPos := `
		SELECT recorded_at
		FROM shipment_tracking_positions
		WHERE org_id = ? AND shipment_id = ?
		ORDER BY recorded_at DESC
		LIMIT 1
	`
	var recAt time.Time
	if errPos := p.db.QueryRowContext(ctx, queryPos, orgID, shipmentID).Scan(&recAt); errPos == nil {
		data.LatestTrackingPing = &recAt
		data.TrackingAgeHours = now.Sub(recAt).Hours()
		if data.TrackingAgeHours < 0 {
			data.TrackingAgeHours = 0
		}
	} else {
		// No tracking positions recorded
		data.TrackingAgeHours = 0
	}

	// 5. Construct Grounded Context Map for Python AI Engine
	isInsufficient := data.MilestonesCount == 0 && data.PlannedETD == nil && data.AuthoritativeETA == nil && len(data.ExistingExceptions) == 0

	ctxMap := map[string]interface{}{
		"shipment_id":              data.ShipmentID,
		"status":                   data.Status,
		"origin_port":              data.OriginPort,
		"destination_port":         data.DestinationPort,
		"carrier_scac":             data.CarrierSCAC,
		"vessel_name":              data.VesselName,
		"voyage_number":            data.VoyageNumber,
		"milestones":               data.Milestones,
		"overdue_milestones":       data.OverdueMilestones,
		"milestone_variance_hours": data.MilestoneVarianceHours,
		"existing_exceptions":      data.ExistingExceptions,
		"active_exceptions_count":  data.ActiveExceptionsCount,
		"tracking_age_hours":       data.TrackingAgeHours,
		"is_delivered":             data.IsDelivered,
		"insufficient_data":        isInsufficient,
		"last_milestone_time":      nowStr,
	}

	if data.LatestTrackingPing != nil {
		ctxMap["tracking_latest_ping"] = data.LatestTrackingPing.Format(time.RFC3339)
	}
	if data.AuthoritativeETA != nil {
		ctxMap["eta"] = data.AuthoritativeETA.Format("2006-01-02 15:04:05")
	}
	if data.PlannedETD != nil {
		ctxMap["etd"] = data.PlannedETD.Format("2006-01-02 15:04:05")
	}

	data.ContextMap = ctxMap
	return data, nil
}
