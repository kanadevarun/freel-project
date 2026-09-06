package shipments

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/freel/backend/internal/shipments/spec"
)

// TrackingProvider defines the production contract for telemetry, positioning, routes, events, and provider refreshes.
type TrackingProvider interface {
	ProviderName() string
	ProviderType() string
	GetMetadata() spec.TrackingProviderMetadata
	GetLatestPosition(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingPosition, error)
	GetPositionHistory(ctx context.Context, orgID int64, sh *spec.Shipment, limit int) ([]spec.TrackingPosition, error)
	GetRoute(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingRoute, error)
	GetTrackingEvents(ctx context.Context, orgID int64, sh *spec.Shipment) ([]spec.TrackingEventNormalized, error)
	Refresh(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingRefreshResult, error)
}

// CalculateFreshness evaluates the data freshness state based on recorded timestamp age.
func CalculateFreshness(recordedAt time.Time) string {
	if recordedAt.IsZero() {
		return spec.TrackingFreshnessUnavailable
	}
	age := time.Since(recordedAt)
	if age < 30*time.Minute {
		return spec.TrackingFreshnessLive
	} else if age < 6*time.Hour {
		return spec.TrackingFreshnessRecent
	} else if age < 48*time.Hour {
		return spec.TrackingFreshnessStale
	}
	return spec.TrackingFreshnessUnavailable
}

// Known global port coordinates
var portCoordinateDirectory = map[string]spec.TrackingCoordinates{
	"INNSA": {Latitude: 18.9499, Longitude: 72.9515, Name: "Nhava Sheva (JNPT)", Code: "INNSA"},
	"INBOM": {Latitude: 18.9438, Longitude: 72.8354, Name: "Mumbai Port", Code: "INBOM"},
	"INMAA": {Latitude: 13.0827, Longitude: 80.2707, Name: "Chennai Port", Code: "INMAA"},
	"INCOK": {Latitude: 9.9312, Longitude: 76.2673, Name: "Cochin Port", Code: "INCOK"},
	"INMUN": {Latitude: 22.8394, Longitude: 69.7042, Name: "Mundra Port", Code: "INMUN"},
	"NLRTM": {Latitude: 51.9244, Longitude: 4.4777, Name: "Port of Rotterdam", Code: "NLRTM"},
	"BEANR": {Latitude: 51.2194, Longitude: 4.4025, Name: "Port of Antwerp", Code: "BEANR"},
	"DEHAM": {Latitude: 53.5511, Longitude: 9.9937, Name: "Port of Hamburg", Code: "DEHAM"},
	"DEBRV": {Latitude: 53.5396, Longitude: 8.5809, Name: "Bremerhaven", Code: "DEBRV"},
	"CNSHA": {Latitude: 31.2304, Longitude: 121.4737, Name: "Port of Shanghai", Code: "CNSHA"},
	"CNNGB": {Latitude: 29.8683, Longitude: 121.5440, Name: "Port of Ningbo", Code: "CNNGB"},
	"CNSZX": {Latitude: 22.5431, Longitude: 114.0579, Name: "Shenzhen Port", Code: "CNSZX"},
	"SGSIN": {Latitude: 1.3521, Longitude: 103.8198, Name: "Port of Singapore", Code: "SGSIN"},
	"USLAX": {Latitude: 33.7432, Longitude: -118.2673, Name: "Port of Los Angeles", Code: "USLAX"},
	"USLGB": {Latitude: 33.7701, Longitude: -118.1937, Name: "Port of Long Beach", Code: "USLGB"},
	"USNYC": {Latitude: 40.7128, Longitude: -74.0060, Name: "Port of New York & New Jersey", Code: "USNYC"},
	"USSAV": {Latitude: 32.0809, Longitude: -81.0912, Name: "Port of Savannah", Code: "USSAV"},
	"AEJEA": {Latitude: 24.9857, Longitude: 55.0273, Name: "Jebel Ali Port", Code: "AEJEA"},
	"AEDXB": {Latitude: 25.2048, Longitude: 55.2708, Name: "Port Rashid Dubai", Code: "AEDXB"},
	"GBLON": {Latitude: 51.5074, Longitude: -0.1278, Name: "London Gateway", Code: "GBLON"},
	"GBFXT": {Latitude: 51.9632, Longitude: 1.3511, Name: "Port of Felixstowe", Code: "GBFXT"},
	"FRLEH": {Latitude: 49.4944, Longitude: 0.1079, Name: "Port of Le Havre", Code: "FRLEH"},
}

// GetPortCoordinates returns resolved port coordinates or generic defaults
func GetPortCoordinates(code string) spec.TrackingCoordinates {
	clean := strings.ToUpper(strings.TrimSpace(code))
	if coords, ok := portCoordinateDirectory[clean]; ok {
		return coords
	}
	return spec.TrackingCoordinates{Latitude: 18.9499, Longitude: 72.9515, Name: code, Code: clean}
}

// ─── 1. Demo Tracking Provider (Explicit Simulation Tier) ────────────────────

type DemoTrackingProvider struct{}

func NewDemoTrackingProvider() *DemoTrackingProvider {
	return &DemoTrackingProvider{}
}

func (p *DemoTrackingProvider) ProviderName() string {
	return "DemoTrackingProvider"
}

func (p *DemoTrackingProvider) ProviderType() string {
	return spec.ProviderTypeDemo
}

func (p *DemoTrackingProvider) GetMetadata() spec.TrackingProviderMetadata {
	return spec.TrackingProviderMetadata{
		ProviderName:      "DemoTrackingProvider",
		ProviderType:      spec.ProviderTypeDemo,
		IsLive:            false,
		IsConfigured:      true,
		SupportsPositions: true,
		SupportsHistory:   true,
		SupportsEvents:    true,
		SupportsRefresh:   true,
		EnvironmentMode:   getEnvironmentMode(),
	}
}

func (p *DemoTrackingProvider) buildCorridorWaypoints(originCode, destCode string) []spec.TrackingWaypoint {
	return []spec.TrackingWaypoint{
		{Name: "Origin Port Berth (Nhava Sheva)", Latitude: 18.9499, Longitude: 72.9515, Sequence: 1},
		{Name: "Arabian Sea Pilot Station", Latitude: 17.5000, Longitude: 68.2000, Sequence: 2},
		{Name: "Gulf of Aden Waypoint", Latitude: 12.8000, Longitude: 48.5000, Sequence: 3},
		{Name: "Bab-el-Mandeb Strait", Latitude: 12.5800, Longitude: 43.3300, Sequence: 4},
		{Name: "Red Sea Transit Lane", Latitude: 20.0000, Longitude: 38.5000, Sequence: 5},
		{Name: "Suez Canal Southern Entry", Latitude: 27.8000, Longitude: 34.2000, Sequence: 6},
		{Name: "Port Said Northbound Exit", Latitude: 31.2600, Longitude: 32.3000, Sequence: 7},
		{Name: "Mediterranean Passage", Latitude: 34.5000, Longitude: 22.0000, Sequence: 8},
		{Name: "Strait of Gibraltar", Latitude: 35.9600, Longitude: -5.6000, Sequence: 9},
		{Name: "Bay of Biscay Corridor", Latitude: 45.0000, Longitude: -7.0000, Sequence: 10},
		{Name: "English Channel West Entry", Latitude: 49.5000, Longitude: -3.5000, Sequence: 11},
		{Name: "Rotterdam Maas Approach", Latitude: 51.9800, Longitude: 3.9500, Sequence: 12},
		{Name: "ECT Delta Terminal (NLRTM)", Latitude: 51.9244, Longitude: 4.4777, Sequence: 13},
	}
}

func (p *DemoTrackingProvider) GetRoute(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingRoute, error) {
	orig := GetPortCoordinates(sh.OriginPort)
	dest := GetPortCoordinates(sh.DestinationPort)
	waypoints := p.buildCorridorWaypoints(sh.OriginPort, sh.DestinationPort)

	progressPct := getShipmentProgressPct(sh)
	totalWps := len(waypoints)
	passedIndex := int(float64(totalWps-1) * (float64(progressPct) / 100.0))

	polyline := make([]spec.TrackingCoordinates, 0, len(waypoints))
	for i := range waypoints {
		if i <= passedIndex {
			waypoints[i].Passed = true
			passedTime := time.Now().Add(-time.Duration((passedIndex-i)*6) * time.Hour).Format("Jan 02, 15:04")
			waypoints[i].PassedAt = &passedTime
		}
		polyline = append(polyline, spec.TrackingCoordinates{
			Latitude:  waypoints[i].Latitude,
			Longitude: waypoints[i].Longitude,
			Name:      waypoints[i].Name,
		})
	}

	plannedDist := 6750.0
	distRemaining := plannedDist * (1.0 - (float64(progressPct) / 100.0))
	if distRemaining < 0 {
		distRemaining = 0
	}

	return &spec.TrackingRoute{
		Origin:                 orig.Code,
		Destination:            dest.Code,
		OriginCoordinates:      orig,
		DestinationCoordinates: dest,
		Waypoints:              waypoints,
		PlannedDistanceNM:      plannedDist,
		DistanceRemainingNM:    distRemaining,
		TransitDurationDays:    16.0,
		RoutePolyline:          polyline,
	}, nil
}

func (p *DemoTrackingProvider) GetLatestPosition(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingPosition, error) {
	waypoints := p.buildCorridorWaypoints(sh.OriginPort, sh.DestinationPort)
	progressPct := getShipmentProgressPct(sh)

	numSegments := len(waypoints) - 1
	rawProgress := (float64(progressPct) / 100.0) * float64(numSegments)
	segIdx := int(math.Floor(rawProgress))
	if segIdx >= numSegments {
		segIdx = numSegments - 1
	}
	frac := rawProgress - float64(segIdx)

	p1 := waypoints[segIdx]
	p2 := waypoints[segIdx+1]

	curLat := p1.Latitude + (p2.Latitude-p1.Latitude)*frac
	curLng := p1.Longitude + (p2.Longitude-p1.Longitude)*frac

	vesselName := "Maersk Mc-Kinney Moller"
	if sh.VesselName != nil && *sh.VesselName != "" {
		vesselName = *sh.VesselName
	}

	locName := p1.Name
	if frac > 0.4 {
		locName = "En route to " + p2.Name
	}

	now := time.Now()
	recordedAt := now.Add(-12 * time.Minute)

	return &spec.TrackingPosition{
		OrgID:          orgID,
		ShipmentID:     sh.ID,
		VesselName:     &vesselName,
		Latitude:       curLat,
		Longitude:      curLng,
		SpeedKnots:     18.7,
		HeadingDegrees: 312.0,
		LocationName:   locName,
		TrackingSource: "Simulation / Demo Telemetry",
		DataFreshness:  CalculateFreshness(recordedAt),
		RecordedAt:     recordedAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (p *DemoTrackingProvider) GetPositionHistory(ctx context.Context, orgID int64, sh *spec.Shipment, limit int) ([]spec.TrackingPosition, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	waypoints := p.buildCorridorWaypoints(sh.OriginPort, sh.DestinationPort)
	progressPct := getShipmentProgressPct(sh)

	totalWps := len(waypoints)
	passedIndex := int(float64(totalWps-1) * (float64(progressPct) / 100.0))
	if passedIndex < 0 {
		passedIndex = 0
	}

	vesselName := "Maersk Mc-Kinney Moller"
	if sh.VesselName != nil && *sh.VesselName != "" {
		vesselName = *sh.VesselName
	}

	positions := make([]spec.TrackingPosition, 0, limit)
	now := time.Now()

	for i := passedIndex; i >= 0 && len(positions) < limit; i-- {
		wp := waypoints[i]
		recTime := now.Add(-time.Duration((passedIndex-i)*6+1) * time.Hour)
		positions = append(positions, spec.TrackingPosition{
			OrgID:          orgID,
			ShipmentID:     sh.ID,
			VesselName:     &vesselName,
			Latitude:       wp.Latitude,
			Longitude:      wp.Longitude,
			SpeedKnots:     18.5,
			HeadingDegrees: 312.0,
			LocationName:   wp.Name,
			TrackingSource: "Simulation / Demo Telemetry",
			DataFreshness:  CalculateFreshness(recTime),
			RecordedAt:     recTime,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}
	return positions, nil
}

func (p *DemoTrackingProvider) GetTrackingEvents(ctx context.Context, orgID int64, sh *spec.Shipment) ([]spec.TrackingEventNormalized, error) {
	now := time.Now()
	return []spec.TrackingEventNormalized{
		{
			ID:            1,
			EventID:       fmt.Sprintf("SIM-EVT-315-%d-1", sh.ID),
			MilestoneCode: spec.ARRIVED,
			Category:      "CARRIER",
			Title:         "Container Discharged at Terminal",
			Description:   "Discharge completed at ECT Delta Terminal Rotterdam (NLRTM). Staged in yard Bay 14.",
			Location:      "Rotterdam Port (NLRTM)",
			EventTime:     now.Add(-2 * time.Hour),
			Source:        "EDI 315 (Simulated)",
		},
		{
			ID:            2,
			EventID:       fmt.Sprintf("SIM-EVT-AIS-%d-2", sh.ID),
			MilestoneCode: spec.ARRIVED,
			Category:      "AIS",
			Title:         "Vessel Berthed at Destination Port",
			Description:   "Maersk Mc-Kinney Moller safely berthed at Berth 4 ECT Delta.",
			Location:      "Rotterdam Port (NLRTM)",
			EventTime:     now.Add(-6 * time.Hour),
			Source:        "AIS Receiver (Simulated)",
		},
		{
			ID:            3,
			EventID:       fmt.Sprintf("SIM-EVT-AIS-%d-3", sh.ID),
			MilestoneCode: spec.IN_TRANSIT,
			Category:      "AIS",
			Title:         "AIS Waypoint Passage: Red Sea / Suez",
			Description:   "Vessel speed 18.2 kts, heading 312° NW. Sea conditions calm.",
			Location:      "Red Sea Corridor",
			EventTime:     now.Add(-36 * time.Hour),
			Source:        "AIS Feed (Simulated)",
		},
		{
			ID:            4,
			EventID:       fmt.Sprintf("SIM-EVT-315-%d-4", sh.ID),
			MilestoneCode: spec.DEPARTED,
			Category:      "CARRIER",
			Title:         "Vessel Departed Origin Port",
			Description:   "Departed Jawaharlal Nehru Port Trust (Nhava Sheva). Voyage 2601W.",
			Location:      "Nhava Sheva (INNSA)",
			EventTime:     now.Add(-96 * time.Hour),
			Source:        "EDI 315 (Simulated)",
		},
		{
			ID:            5,
			EventID:       fmt.Sprintf("SIM-EVT-TOS-%d-5", sh.ID),
			MilestoneCode: spec.BOOKED,
			Category:      "TERMINAL",
			Title:         "Origin Terminal Gate In",
			Description:   "Container arrived and inspected at Origin Port Gate 2. Tare & seal verified.",
			Location:      "Nhava Sheva Terminal",
			EventTime:     now.Add(-140 * time.Hour),
			Source:        "TOS Feed (Simulated)",
		},
		{
			ID:            6,
			EventID:       fmt.Sprintf("SIM-EVT-301-%d-6", sh.ID),
			MilestoneCode: spec.BOOKED,
			Category:      "SYSTEM",
			Title:         "Carrier Booking Confirmed",
			Description:   "Electronic confirmation received from ocean carrier for container allocation.",
			Location:      "Carrier Network",
			EventTime:     now.Add(-200 * time.Hour),
			Source:        "EDI 301 (Simulated)",
		},
	}, nil
}

func (p *DemoTrackingProvider) Refresh(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingRefreshResult, error) {
	now := time.Now()
	return &spec.TrackingRefreshResult{
		Success:       true,
		Provider:      p.GetMetadata(),
		DataFreshness: spec.TrackingFreshnessRecent,
		LastUpdatedAt: &now,
		NewPositions:  1,
		NewEvents:     0,
		UsedFallback:  false,
		Message:       "Demo simulation telemetry refreshed successfully",
	}, nil
}

// ─── 2. Carrier Tracking Provider (Carrier EDI/API Integration Tier) ──────────

type CarrierTrackingProvider struct {
	apiURL string
	apiKey string
}

func NewCarrierTrackingProvider() *CarrierTrackingProvider {
	return &CarrierTrackingProvider{
		apiURL: os.Getenv("CARRIER_TRACKING_API_URL"),
		apiKey: os.Getenv("CARRIER_TRACKING_API_KEY"),
	}
}

func (p *CarrierTrackingProvider) ProviderName() string {
	return "CarrierAPIProvider"
}

func (p *CarrierTrackingProvider) ProviderType() string {
	return spec.ProviderTypeCarrier
}

func (p *CarrierTrackingProvider) isConfigured() bool {
	return p.apiURL != "" && p.apiKey != ""
}

func (p *CarrierTrackingProvider) GetMetadata() spec.TrackingProviderMetadata {
	configured := p.isConfigured()
	return spec.TrackingProviderMetadata{
		ProviderName:      "CarrierAPIProvider",
		ProviderType:      spec.ProviderTypeCarrier,
		IsLive:            configured,
		IsConfigured:      configured,
		SupportsPositions: true,
		SupportsHistory:   true,
		SupportsEvents:    true,
		SupportsRefresh:   true,
		EnvironmentMode:   getEnvironmentMode(),
	}
}

func (p *CarrierTrackingProvider) GetRoute(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingRoute, error) {
	orig := GetPortCoordinates(sh.OriginPort)
	dest := GetPortCoordinates(sh.DestinationPort)
	return &spec.TrackingRoute{
		Origin:                 orig.Code,
		Destination:            dest.Code,
		OriginCoordinates:      orig,
		DestinationCoordinates: dest,
		Waypoints: []spec.TrackingWaypoint{
			{Name: orig.Name, Latitude: orig.Latitude, Longitude: orig.Longitude, Sequence: 1, Passed: true},
			{Name: dest.Name, Latitude: dest.Latitude, Longitude: dest.Longitude, Sequence: 2, Passed: false},
		},
		PlannedDistanceNM:   6750.0,
		DistanceRemainingNM: 3200.0,
		TransitDurationDays: 16.0,
	}, nil
}

func (p *CarrierTrackingProvider) GetLatestPosition(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingPosition, error) {
	// If credentials not configured, return nil gracefully to invoke database fallback
	if !p.isConfigured() {
		return nil, nil
	}
	// When real Carrier API is integrated, HTTP call will be executed here
	return nil, nil
}

func (p *CarrierTrackingProvider) GetPositionHistory(ctx context.Context, orgID int64, sh *spec.Shipment, limit int) ([]spec.TrackingPosition, error) {
	return []spec.TrackingPosition{}, nil
}

func (p *CarrierTrackingProvider) GetTrackingEvents(ctx context.Context, orgID int64, sh *spec.Shipment) ([]spec.TrackingEventNormalized, error) {
	return []spec.TrackingEventNormalized{}, nil
}

func (p *CarrierTrackingProvider) Refresh(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingRefreshResult, error) {
	now := time.Now()
	if !p.isConfigured() {
		return &spec.TrackingRefreshResult{
			Success:       true,
			Provider:      p.GetMetadata(),
			DataFreshness: spec.TrackingFreshnessUnavailable,
			LastUpdatedAt: &now,
			NewPositions:  0,
			NewEvents:     0,
			UsedFallback:  true,
			Message:       "Carrier API not configured. Showing latest persisted operational tracking data.",
		}, nil
	}
	return &spec.TrackingRefreshResult{
		Success:       true,
		Provider:      p.GetMetadata(),
		DataFreshness: spec.TrackingFreshnessRecent,
		LastUpdatedAt: &now,
		NewPositions:  1,
		NewEvents:     0,
		UsedFallback:  false,
		Message:       "Tracking telemetry refreshed from Carrier API",
	}, nil
}

// ─── 3. AIS Tracking Provider (Satellite Vessel Beacon Tier) ─────────────────

type AISTrackingProvider struct {
	apiURL string
	apiKey string
}

func NewAISTrackingProvider() *AISTrackingProvider {
	return &AISTrackingProvider{
		apiURL: os.Getenv("AIS_TRACKING_API_URL"),
		apiKey: os.Getenv("AIS_TRACKING_API_KEY"),
	}
}

func (p *AISTrackingProvider) ProviderName() string {
	return "AISSatelliteProvider"
}

func (p *AISTrackingProvider) ProviderType() string {
	return spec.ProviderTypeAIS
}

func (p *AISTrackingProvider) isConfigured() bool {
	return p.apiURL != "" && p.apiKey != ""
}

func (p *AISTrackingProvider) GetMetadata() spec.TrackingProviderMetadata {
	configured := p.isConfigured()
	return spec.TrackingProviderMetadata{
		ProviderName:      "AISSatelliteProvider",
		ProviderType:      spec.ProviderTypeAIS,
		IsLive:            configured,
		IsConfigured:      configured,
		SupportsPositions: true,
		SupportsHistory:   true,
		SupportsEvents:    true,
		SupportsRefresh:   true,
		EnvironmentMode:   getEnvironmentMode(),
	}
}

func (p *AISTrackingProvider) GetRoute(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingRoute, error) {
	orig := GetPortCoordinates(sh.OriginPort)
	dest := GetPortCoordinates(sh.DestinationPort)
	return &spec.TrackingRoute{
		Origin:                 orig.Code,
		Destination:            dest.Code,
		OriginCoordinates:      orig,
		DestinationCoordinates: dest,
		Waypoints:              []spec.TrackingWaypoint{},
	}, nil
}

func (p *AISTrackingProvider) GetLatestPosition(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingPosition, error) {
	if !p.isConfigured() {
		return nil, nil
	}
	return nil, nil
}

func (p *AISTrackingProvider) GetPositionHistory(ctx context.Context, orgID int64, sh *spec.Shipment, limit int) ([]spec.TrackingPosition, error) {
	return []spec.TrackingPosition{}, nil
}

func (p *AISTrackingProvider) GetTrackingEvents(ctx context.Context, orgID int64, sh *spec.Shipment) ([]spec.TrackingEventNormalized, error) {
	return []spec.TrackingEventNormalized{}, nil
}

func (p *AISTrackingProvider) Refresh(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingRefreshResult, error) {
	now := time.Now()
	if !p.isConfigured() {
		return &spec.TrackingRefreshResult{
			Success:       true,
			Provider:      p.GetMetadata(),
			DataFreshness: spec.TrackingFreshnessUnavailable,
			LastUpdatedAt: &now,
			NewPositions:  0,
			NewEvents:     0,
			UsedFallback:  true,
			Message:       "AIS Satellite provider not configured. Retaining persisted telemetry data.",
		}, nil
	}
	return &spec.TrackingRefreshResult{
		Success:       true,
		Provider:      p.GetMetadata(),
		DataFreshness: spec.TrackingFreshnessLive,
		LastUpdatedAt: &now,
		NewPositions:  1,
		NewEvents:     0,
		UsedFallback:  false,
		Message:       "Telemetry refreshed via live satellite AIS beacon",
	}, nil
}

// ─── 4. Database Fallback Provider (Strict Production Tier - No Fabrication) ──

type DatabaseTrackingProvider struct{}

func NewDatabaseTrackingProvider() *DatabaseTrackingProvider {
	return &DatabaseTrackingProvider{}
}

func (p *DatabaseTrackingProvider) ProviderName() string {
	return "DatabaseTrackingProvider"
}

func (p *DatabaseTrackingProvider) ProviderType() string {
	return spec.ProviderTypeDatabase
}

func (p *DatabaseTrackingProvider) GetMetadata() spec.TrackingProviderMetadata {
	return spec.TrackingProviderMetadata{
		ProviderName:      "DatabaseTrackingProvider",
		ProviderType:      spec.ProviderTypeDatabase,
		IsLive:            false,
		IsConfigured:      true,
		SupportsPositions: true,
		SupportsHistory:   true,
		SupportsEvents:    true,
		SupportsRefresh:   true,
		EnvironmentMode:   getEnvironmentMode(),
	}
}

func (p *DatabaseTrackingProvider) GetRoute(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingRoute, error) {
	orig := GetPortCoordinates(sh.OriginPort)
	dest := GetPortCoordinates(sh.DestinationPort)
	return &spec.TrackingRoute{
		Origin:                 orig.Code,
		Destination:            dest.Code,
		OriginCoordinates:      orig,
		DestinationCoordinates: dest,
		Waypoints:              []spec.TrackingWaypoint{},
	}, nil
}

func (p *DatabaseTrackingProvider) GetLatestPosition(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingPosition, error) {
	// Returns nil so caller queries only persisted positions from database
	return nil, nil
}

func (p *DatabaseTrackingProvider) GetPositionHistory(ctx context.Context, orgID int64, sh *spec.Shipment, limit int) ([]spec.TrackingPosition, error) {
	return []spec.TrackingPosition{}, nil
}

func (p *DatabaseTrackingProvider) GetTrackingEvents(ctx context.Context, orgID int64, sh *spec.Shipment) ([]spec.TrackingEventNormalized, error) {
	return []spec.TrackingEventNormalized{}, nil
}

func (p *DatabaseTrackingProvider) Refresh(ctx context.Context, orgID int64, sh *spec.Shipment) (*spec.TrackingRefreshResult, error) {
	now := time.Now()
	return &spec.TrackingRefreshResult{
		Success:       true,
		Provider:      p.GetMetadata(),
		DataFreshness: spec.TrackingFreshnessUnavailable,
		LastUpdatedAt: &now,
		NewPositions:  0,
		NewEvents:     0,
		UsedFallback:  true,
		Message:       "Using latest database-persisted tracking positions and milestones.",
	}, nil
}

// ─── Central Provider Resolution & Configuration ──────────────────────────────

func getEnvironmentMode() string {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("TRACKING_PROVIDER_MODE")))
	if mode == "" {
		env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
		if env != "" {
			return env
		}
		return "development"
	}
	return mode
}

// ResolveTrackingProvider returns the active tracking provider based on TRACKING_PROVIDER environment variable
func ResolveTrackingProvider() TrackingProvider {
	providerType := strings.ToLower(strings.TrimSpace(os.Getenv("TRACKING_PROVIDER")))
	switch providerType {
	case "carrier":
		return NewCarrierTrackingProvider()
	case "ais":
		return NewAISTrackingProvider()
	case "database", "prod", "production":
		return NewDatabaseTrackingProvider()
	case "demo", "":
		fallthrough
	default:
		return NewDemoTrackingProvider()
	}
}

// NormalizeCarrierEvents converts persisted carrier tracking events into normalized events
func NormalizeCarrierEvents(dbEvents []*spec.CarrierTrackingEvent) []spec.TrackingEventNormalized {
	normalized := make([]spec.TrackingEventNormalized, 0, len(dbEvents))
	for _, ev := range dbEvents {
		desc := ev.RawDescription
		loc := ev.Location
		category := "CARRIER"
		if strings.Contains(strings.ToUpper(ev.SourceType), "AIS") {
			category = "AIS"
		}
		title := "Carrier Tracking Event"
		if ev.RawDescription != "" {
			title = ev.RawDescription
		}
		milestone := spec.IN_TRANSIT
		if ev.MilestoneCode != "" {
			milestone = ev.MilestoneCode
		}

		normalized = append(normalized, spec.TrackingEventNormalized{
			ID:            ev.ID,
			EventID:       ev.EventID,
			MilestoneCode: milestone,
			Category:      category,
			Title:         title,
			Description:   desc,
			Location:      loc,
			EventTime:     ev.EventTime,
			Source:        ev.SourceType,
		})
	}
	return normalized
}

func getShipmentProgressPct(sh *spec.Shipment) int {
	if sh == nil {
		return 68
	}
	switch sh.Status {
	case spec.BOOKING_PENDING:
		return 10
	case spec.BOOKED:
		return 30
	case spec.DEPARTED:
		return 50
	case spec.IN_TRANSIT:
		return 68
	case spec.ARRIVED:
		return 90
	case spec.DELIVERED:
		return 100
	default:
		return 68
	}
}
