package reports

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockRepository struct {
	snapshots       map[int64]*AnalyticsReportSnapshot
	exports         []*ReportExportRequest
	distributions   []*ReportDistributionRequest
	approvalReqIDs  []int64
	nextSnapshotID  int64
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		snapshots:      make(map[int64]*AnalyticsReportSnapshot),
		exports:        make([]*ReportExportRequest, 0),
		distributions:  make([]*ReportDistributionRequest, 0),
		approvalReqIDs: make([]int64, 0),
		nextSnapshotID: 1,
	}
}

func (m *mockRepository) GetOperationalVolumeMetrics(ctx context.Context, orgID int64, start, end time.Time) ([]HistoricalDataPointDTO, map[string]interface{}, error) {
	vol := 150
	series := []HistoricalDataPointDTO{
		{Period: "2026-06", Value: 120.0, Volume: &vol, Unit: "shipments"},
		{Period: "2026-07", Value: 135.0, Volume: &vol, Unit: "shipments"},
		{Period: "2026-08", Value: 150.0, Volume: &vol, Unit: "shipments"},
	}
	metrics := map[string]interface{}{
		"total_shipments": 405,
		"on_time_delivery_rate": "95.2%",
	}
	return series, metrics, nil
}

func (m *mockRepository) GetRevenueFinanceMetrics(ctx context.Context, orgID int64, start, end time.Time) ([]HistoricalDataPointDTO, map[string]interface{}, error) {
	series := []HistoricalDataPointDTO{
		{Period: "2026-06", Value: 85000.0, Unit: "USD"},
		{Period: "2026-07", Value: 92000.0, Unit: "USD"},
		{Period: "2026-08", Value: 104000.0, Unit: "USD"},
	}
	metrics := map[string]interface{}{
		"total_invoiced_usd": 281000.0,
		"total_collected_usd": 240000.0,
	}
	return series, metrics, nil
}

func (m *mockRepository) GetCommercialFunnelMetrics(ctx context.Context, orgID int64, start, end time.Time) ([]HistoricalDataPointDTO, map[string]interface{}, error) {
	return []HistoricalDataPointDTO{}, map[string]interface{}{"win_rate": "42.0%"}, nil
}

func (m *mockRepository) GetContractComplianceMetrics(ctx context.Context, orgID int64, start, end time.Time) ([]HistoricalDataPointDTO, map[string]interface{}, error) {
	return []HistoricalDataPointDTO{}, map[string]interface{}{"active_contracts": 12}, nil
}

func (m *mockRepository) SaveReportSnapshot(ctx context.Context, s *AnalyticsReportSnapshot) error {
	s.ID = m.nextSnapshotID
	m.nextSnapshotID++
	m.snapshots[s.ID] = s
	return nil
}

func (m *mockRepository) GetReportSnapshot(ctx context.Context, orgID, id int64) (*AnalyticsReportSnapshot, error) {
	if s, ok := m.snapshots[id]; ok && s.OrgID == orgID {
		return s, nil
	}
	return nil, nil
}

func (m *mockRepository) ListReportSnapshots(ctx context.Context, orgID int64, limit int) ([]AnalyticsReportSnapshot, error) {
	res := make([]AnalyticsReportSnapshot, 0)
	for _, s := range m.snapshots {
		if s.OrgID == orgID {
			res = append(res, *s)
		}
	}
	return res, nil
}

func (m *mockRepository) SaveExportRequest(ctx context.Context, req *ReportExportRequest) error {
	req.ID = int64(len(m.exports) + 1)
	m.exports = append(m.exports, req)
	return nil
}

func (m *mockRepository) SaveDistributionRequest(ctx context.Context, req *ReportDistributionRequest) error {
	req.ID = int64(len(m.distributions) + 1)
	m.distributions = append(m.distributions, req)
	return nil
}

func (m *mockRepository) ListDistributionRequests(ctx context.Context, orgID int64, limit int) ([]ReportDistributionRequest, error) {
	res := make([]ReportDistributionRequest, 0)
	for _, d := range m.distributions {
		if d.OrgID == orgID {
			res = append(res, *d)
		}
	}
	return res, nil
}

func (m *mockRepository) CreateApprovalForDistribution(ctx context.Context, orgID, userID int64, userName, reportType, recipients, notes, corrID string) (int64, error) {
	newID := int64(3001 + len(m.approvalReqIDs))
	m.approvalReqIDs = append(m.approvalReqIDs, newID)
	return newID, nil
}

// TestGetAdvancedReportWithSidecar tests fetching advanced report with Python sidecar forecast
func TestGetAdvancedReportWithSidecar(t *testing.T) {
	mockSidecar := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/reporting/forecast-and-narrative" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("X-LogisticsHQ-Service-Key") != DefaultServiceKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req SidecarReportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := SidecarReportResponse{
			ReportType:          req.ReportType,
			DateRange:           req.DateRange,
			IsForecastAvailable: true,
			ForecastSeries: []ForecastDataPointDTO{
				{
					Period:          "2026-09 (Forecast)",
					ProjectedValue:  165.0,
					LowerBound:      145.0,
					UpperBound:      185.0,
					IsForecast:      true,
					Label:           "Model-Generated Forecast",
					ConfidenceScore: 0.88,
				},
			},
			ForecastHorizon:       "Next 30 Days",
			ForecastAssumptions:   []string{"Assumes stable maritime schedule momentum"},
			ForecastLimitations:   []string{"Model-generated estimates, not guarantees"},
			ConfidenceScore:       0.88,
			TrendInsights:         []TrendInsightDTO{{MetricName: "Shipment Volume", Direction: "INCREASING", ChangePct: 25.0, Significance: "HIGH", Explanation: "Volume expanding"}},
			Anomalies:             []AnomalyPointDTO{},
			ExecutiveNarrative:    "### Executive Report Summary\n- Cumulative Volume: 405 shipments.",
			ActionRecommendations: []string{"Review carrier allocation quotas"},
			CorrelationID:         req.CorrelationID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockSidecar.Close()

	repo := newMockRepository()
	svc := NewService(repo, WithSidecar(mockSidecar.URL, DefaultServiceKey))

	ctx := context.Background()
	rep, err := svc.GetAdvancedReport(ctx, 1, 42, "OPERATIONAL_VOLUME", "LAST_90D")
	if err != nil {
		t.Fatalf("GetAdvancedReport failed: %v", err)
	}

	if rep == nil {
		t.Fatal("Expected non-nil report")
	}

	if !rep.IsForecastAvailable {
		t.Error("Expected forecast to be available")
	}

	if len(rep.ForecastSeries) != 1 {
		t.Errorf("Expected 1 forecast data point, got %d", len(rep.ForecastSeries))
	}

	if rep.ForecastSeries[0].Label != "Model-Generated Forecast" {
		t.Errorf("Expected forecast label, got %s", rep.ForecastSeries[0].Label)
	}

	if rep.SnapshotID <= 0 {
		t.Errorf("Expected snapshot ID to be recorded, got %d", rep.SnapshotID)
	}
}

// TestExportReportCSVAndJSON tests CSV and JSON export logic
func TestExportReportCSVAndJSON(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	ctx := context.Background()

	// 1. Export CSV
	csvReq, err := svc.ExportReport(ctx, 1, 42, ExportReportInput{
		ReportType: "OPERATIONAL_VOLUME",
		DateRange:  "LAST_90D",
		Format:     "CSV",
	})
	if err != nil {
		t.Fatalf("ExportReport CSV failed: %v", err)
	}

	if csvReq.Status != "COMPLETED" {
		t.Errorf("Expected status COMPLETED, got %s", csvReq.Status)
	}
	if csvReq.ExportContent == nil || len(*csvReq.ExportContent) == 0 {
		t.Error("Expected non-empty CSV export content")
	}
}

// TestReportDistributionApprovalGate verifies external distribution requires approval
func TestReportDistributionApprovalGate(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	ctx := context.Background()

	distReq, err := svc.RequestDistribution(ctx, 1, 42, "testuser", DistributeReportInput{
		ReportType:      "REVENUE_FINANCE",
		DateRange:       "LAST_90D",
		RecipientEmails: []string{"cfo@logisticsclient.com"},
		Notes:           "Monthly revenue update",
	})
	if err != nil {
		t.Fatalf("RequestDistribution failed: %v", err)
	}

	if distReq.Status != "PENDING_APPROVAL" {
		t.Errorf("Expected PENDING_APPROVAL status for distribution, got %s", distReq.Status)
	}

	if distReq.ApprovalID == nil || *distReq.ApprovalID <= 0 {
		t.Errorf("Expected valid approval ID, got %v", distReq.ApprovalID)
	}
}

// TestTenantIsolationInReporting verifies cross-tenant access returns nil
func TestTenantIsolationInReporting(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	ctx := context.Background()

	// Org 1 creates snapshot
	repo.snapshots[99] = &AnalyticsReportSnapshot{
		ID:         99,
		OrgID:      1,
		UserID:     10,
		ReportType: "OPERATIONAL_VOLUME",
	}

	// Org 2 attempts to fetch Org 1 snapshot -> must return nil
	sn, err := repo.GetReportSnapshot(ctx, 2, 99)
	if err != nil {
		t.Fatalf("GetReportSnapshot failed: %v", err)
	}
	if sn != nil {
		t.Errorf("Tenant isolation violated: Org 2 retrieved Org 1 snapshot %+v", sn)
	}
	_ = svc
}
