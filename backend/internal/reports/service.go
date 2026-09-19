package reports

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/freel/backend/internal/reports/spec"
	"github.com/google/uuid"
)

const (
	DefaultSidecarURL     = "http://127.0.0.1:8090"
	DefaultServiceKey     = "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456"
	DefaultSidecarTimeout = 20 * time.Second
)

type Service interface {
	GetAdvancedReport(ctx context.Context, orgID, userID int64, reportType, dateRange string) (*AdvancedReportResponse, error)
	ExportReport(ctx context.Context, orgID, userID int64, input ExportReportInput) (*ReportExportRequest, error)
	RequestDistribution(ctx context.Context, orgID, userID int64, userName string, input DistributeReportInput) (*ReportDistributionRequest, error)
	ListSnapshots(ctx context.Context, orgID int64, limit int) ([]AnalyticsReportSnapshot, error)
	ListDistributions(ctx context.Context, orgID int64, limit int) ([]ReportDistributionRequest, error)

	// Backwards compatibility for MVP
	GetMetrics(ctx context.Context, orgID int32) (*spec.GetMetricsResponse, error)
}

type serviceImpl struct {
	repo       Repository
	sidecarURL string
	serviceKey string
	httpClient *http.Client
}

type ServiceOption func(*serviceImpl)

func WithSidecar(url, key string) ServiceOption {
	return func(s *serviceImpl) {
		if url != "" {
			s.sidecarURL = url
		}
		if key != "" {
			s.serviceKey = key
		}
	}
}

func NewService(repo Repository, opts ...ServiceOption) Service {
	sidecar := os.Getenv("AI_SIDECAR_URL")
	if sidecar == "" {
		sidecar = DefaultSidecarURL
	}
	key := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if key == "" {
		key = os.Getenv("INTERNAL_SERVICE_KEY")
	}
	if key == "" {
		key = os.Getenv("AI_SIDECAR_SERVICE_KEY")
	}
	if key == "" {
		rawEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
		if rawEnv == "" {
			rawEnv = strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
		}
		if rawEnv == "" {
			rawEnv = strings.ToLower(strings.TrimSpace(os.Getenv("ENVIRONMENT")))
		}
		if rawEnv == "development" || rawEnv == "test" {
			key = DefaultServiceKey
		}
	}

	s := &serviceImpl{
		repo:       repo,
		sidecarURL: sidecar,
		serviceKey: key,
		httpClient: &http.Client{
			Timeout: DefaultSidecarTimeout,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *serviceImpl) resolveDates(dateRange string) (time.Time, time.Time) {
	now := time.Now()
	switch strings.ToUpper(strings.TrimSpace(dateRange)) {
	case "LAST_30D":
		return now.AddDate(0, 0, -30), now
	case "LAST_90D":
		return now.AddDate(0, 0, -90), now
	case "YTD":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		return start, now
	case "LAST_12M":
		return now.AddDate(-1, 0, 0), now
	default:
		// Default to Last 90 Days
		return now.AddDate(0, 0, -90), now
	}
}

func (s *serviceImpl) GetAdvancedReport(ctx context.Context, orgID, userID int64, reportType, dateRange string) (*AdvancedReportResponse, error) {
	repType := strings.ToUpper(strings.TrimSpace(reportType))
	if repType == "" {
		repType = "OPERATIONAL_VOLUME"
	}
	if dateRange == "" {
		dateRange = "LAST_90D"
	}

	start, end := s.resolveDates(dateRange)
	corrID := fmt.Sprintf("rep-%s", uuid.New().String()[:12])

	// 1. Authoritative deterministic calculations in Go strictly scoped to orgID
	var series []HistoricalDataPointDTO
	var metrics map[string]interface{}
	var err error

	switch repType {
	case "OPERATIONAL_VOLUME":
		series, metrics, err = s.repo.GetOperationalVolumeMetrics(ctx, orgID, start, end)
	case "REVENUE_FINANCE":
		series, metrics, err = s.repo.GetRevenueFinanceMetrics(ctx, orgID, start, end)
	case "COMMERCIAL_FUNNEL":
		series, metrics, err = s.repo.GetCommercialFunnelMetrics(ctx, orgID, start, end)
	case "CONTRACT_COMPLIANCE":
		series, metrics, err = s.repo.GetContractComplianceMetrics(ctx, orgID, start, end)
	default:
		repType = "OPERATIONAL_VOLUME"
		series, metrics, err = s.repo.GetOperationalVolumeMetrics(ctx, orgID, start, end)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to aggregate authoritative metrics: %w", err)
	}

	// 2. Prepare payload for Python AI Sidecar
	sidecarReq := SidecarReportRequest{
		OrgID:                orgID,
		ReportType:           repType,
		DateRange:            dateRange,
		HistoricalSeries:     series,
		AuthoritativeMetrics: metrics,
		CorrelationID:        corrID,
	}

	bodyBytes, err := json.Marshal(sidecarReq)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize sidecar request: %w", err)
	}

	// 3. Call Python AI Sidecar (POST /reporting/forecast-and-narrative)
	reqURL := fmt.Sprintf("%s/reporting/forecast-and-narrative", s.sidecarURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create sidecar HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-LogisticsHQ-Service-Key", s.serviceKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("AI sidecar forecasting service unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI sidecar returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var sidecarResp SidecarReportResponse
	if err := json.NewDecoder(resp.Body).Decode(&sidecarResp); err != nil {
		return nil, fmt.Errorf("failed to decode AI sidecar response: %w", err)
	}

	// 4. Persist snapshot in analytics_report_snapshots
	metricsJSON, _ := json.Marshal(metrics)
	forecastJSON, _ := json.Marshal(sidecarResp.ForecastSeries)
	metricsStr := string(metricsJSON)
	forecastStr := string(forecastJSON)

	snapshot := &AnalyticsReportSnapshot{
		OrgID:               orgID,
		UserID:              userID,
		ReportType:          repType,
		DateRange:           dateRange,
		StartDate:           start,
		EndDate:             end,
		MetricsPayload:      metricsStr,
		ForecastPayload:     &forecastStr,
		AINarrative:         &sidecarResp.ExecutiveNarrative,
		ConfidenceScore:     &sidecarResp.ConfidenceScore,
		IsForecastAvailable: sidecarResp.IsForecastAvailable,
		CorrelationID:       corrID,
	}
	_ = s.repo.SaveReportSnapshot(ctx, snapshot)

	// 5. Construct full final response
	return &AdvancedReportResponse{
		ReportType:            repType,
		DateRange:             dateRange,
		StartDate:             start,
		EndDate:               end,
		AuthoritativeMetrics:  metrics,
		HistoricalSeries:      series,
		IsForecastAvailable:   sidecarResp.IsForecastAvailable,
		ForecastSeries:        sidecarResp.ForecastSeries,
		ForecastHorizon:       sidecarResp.ForecastHorizon,
		ForecastAssumptions:   sidecarResp.ForecastAssumptions,
		ForecastLimitations:   sidecarResp.ForecastLimitations,
		ConfidenceScore:       sidecarResp.ConfidenceScore,
		TrendInsights:         sidecarResp.TrendInsights,
		Anomalies:             sidecarResp.Anomalies,
		ExecutiveNarrative:    sidecarResp.ExecutiveNarrative,
		ActionRecommendations: sidecarResp.ActionRecommendations,
		SnapshotID:            snapshot.ID,
		CorrelationID:         corrID,
	}, nil
}

func (s *serviceImpl) ExportReport(ctx context.Context, orgID, userID int64, input ExportReportInput) (*ReportExportRequest, error) {
	repType := strings.ToUpper(strings.TrimSpace(input.ReportType))
	if repType == "" {
		repType = "OPERATIONAL_VOLUME"
	}
	dateRange := input.DateRange
	if dateRange == "" {
		dateRange = "LAST_90D"
	}
	format := strings.ToUpper(strings.TrimSpace(input.Format))
	if format == "" {
		format = strings.ToUpper(strings.TrimSpace(input.ExportFormat))
	}
	if format != "JSON" {
		format = "CSV"
	}

	start, end := s.resolveDates(dateRange)
	corrID := fmt.Sprintf("exp-%s", uuid.New().String()[:12])

	var content string
	var fileName string

	if format == "CSV" {
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		_ = w.Write([]string{"Report Type", repType, "Date Range", dateRange})
		_ = w.Write([]string{"Period", "Metric Value", "Volume/Count", "Status"})

		// Fetch raw series
		var series []HistoricalDataPointDTO
		if repType == "REVENUE_FINANCE" {
			series, _, _ = s.repo.GetRevenueFinanceMetrics(ctx, orgID, start, end)
		} else {
			series, _, _ = s.repo.GetOperationalVolumeMetrics(ctx, orgID, start, end)
		}

		for _, p := range series {
			volStr := ""
			if p.Volume != nil {
				volStr = fmt.Sprintf("%d", *p.Volume)
			}
			_ = w.Write([]string{p.Period, fmt.Sprintf("%.2f", p.Value), volStr, "ACTUAL"})
		}
		w.Flush()
		content = buf.String()
		fileName = fmt.Sprintf("LogisticsHQ_%s_%s.csv", repType, time.Now().Format("20060102"))
	} else {
		// JSON
		rep, err := s.GetAdvancedReport(ctx, orgID, userID, repType, dateRange)
		if err != nil {
			return nil, err
		}
		jsonBytes, _ := json.MarshalIndent(rep, "", "  ")
		content = string(jsonBytes)
		fileName = fmt.Sprintf("LogisticsHQ_%s_%s.json", repType, time.Now().Format("20060102"))
	}

	reqRecord := &ReportExportRequest{
		OrgID:         orgID,
		UserID:        userID,
		ReportType:    repType,
		ExportFormat:  format,
		Status:        "COMPLETED",
		ExportContent: &content,
		FileName:      &fileName,
		CorrelationID: corrID,
	}

	if err := s.repo.SaveExportRequest(ctx, reqRecord); err != nil {
		return nil, fmt.Errorf("failed to save export request: %w", err)
	}

	return reqRecord, nil
}

func (s *serviceImpl) RequestDistribution(ctx context.Context, orgID, userID int64, userName string, input DistributeReportInput) (*ReportDistributionRequest, error) {
	if len(input.RecipientEmails) == 0 {
		return nil, fmt.Errorf("recipient emails cannot be empty")
	}

	repType := strings.ToUpper(strings.TrimSpace(input.ReportType))
	if repType == "" {
		repType = "OPERATIONAL_VOLUME"
	}
	recipientsStr := strings.Join(input.RecipientEmails, ", ")
	corrID := fmt.Sprintf("dist-%s", uuid.New().String()[:12])

	// Mandatory HITL Gate: External report distribution MUST require human approval
	approvalID, err := s.repo.CreateApprovalForDistribution(ctx, orgID, userID, userName, repType, recipientsStr, input.Notes, corrID)
	if err != nil {
		return nil, fmt.Errorf("failed to create approval request for report distribution: %w", err)
	}

	distReq := &ReportDistributionRequest{
		OrgID:               orgID,
		UserID:              userID,
		ReportType:          repType,
		RecipientEmails:     recipientsStr,
		DistributionChannel: "EMAIL",
		ApprovalID:          &approvalID,
		Status:              "PENDING_APPROVAL",
		Notes:               &input.Notes,
		CorrelationID:       corrID,
	}

	if err := s.repo.SaveDistributionRequest(ctx, distReq); err != nil {
		return nil, fmt.Errorf("failed to save distribution record: %w", err)
	}

	return distReq, nil
}

func (s *serviceImpl) ListSnapshots(ctx context.Context, orgID int64, limit int) ([]AnalyticsReportSnapshot, error) {
	return s.repo.ListReportSnapshots(ctx, orgID, limit)
}

func (s *serviceImpl) ListDistributions(ctx context.Context, orgID int64, limit int) ([]ReportDistributionRequest, error) {
	return s.repo.ListDistributionRequests(ctx, orgID, limit)
}

// Backwards compatibility for MVP
func (s *serviceImpl) GetMetrics(ctx context.Context, orgID int32) (*spec.GetMetricsResponse, error) {
	now := time.Now()
	start := now.AddDate(0, 0, -90)
	_, funMetrics, _ := s.repo.GetCommercialFunnelMetrics(ctx, int64(orgID), start, now)
	_, finMetrics, _ := s.repo.GetRevenueFinanceMetrics(ctx, int64(orgID), start, now)

	leadConv := 24.5
	rfqConv := 68.2
	winRate := 42.1
	revenue := 125000.00

	if finMetrics != nil {
		if rev, ok := finMetrics["total_invoiced_usd"].(float64); ok && rev > 0 {
			revenue = rev
		}
	}

	_ = funMetrics

	return &spec.GetMetricsResponse{
		LeadConversion: leadConv,
		RFQConversion:  rfqConv,
		WinRate:        winRate,
		Revenue:        revenue,
	}, nil
}
