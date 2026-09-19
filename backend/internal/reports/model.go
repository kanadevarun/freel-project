package reports

import (
	"encoding/json"
	"time"
)

// AnalyticsReportSnapshot represents a stored report snapshot with metrics, forecast, and narrative.
type AnalyticsReportSnapshot struct {
	ID                  int64      `db:"id" json:"id"`
	OrgID               int64      `db:"org_id" json:"org_id"`
	UserID              int64      `db:"user_id" json:"user_id"`
	ReportType          string     `db:"report_type" json:"report_type"`
	DateRange           string     `db:"date_range" json:"date_range"`
	StartDate           time.Time  `db:"start_date" json:"start_date"`
	EndDate             time.Time  `db:"end_date" json:"end_date"`
	MetricsPayload      string     `db:"metrics_payload" json:"metrics_payload"`
	ForecastPayload     *string    `db:"forecast_payload" json:"forecast_payload,omitempty"`
	AINarrative         *string    `db:"ai_narrative" json:"ai_narrative,omitempty"`
	ConfidenceScore     *float64   `db:"confidence_score" json:"confidence_score,omitempty"`
	IsForecastAvailable bool       `db:"is_forecast_available" json:"is_forecast_available"`
	CorrelationID       string     `db:"correlation_id" json:"correlation_id"`
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

// ReportExportRequest tracks CSV/JSON export requests.
type ReportExportRequest struct {
	ID            int64     `db:"id" json:"id"`
	OrgID         int64     `db:"org_id" json:"org_id"`
	UserID        int64     `db:"user_id" json:"user_id"`
	ReportType    string    `db:"report_type" json:"report_type"`
	ExportFormat  string    `db:"export_format" json:"export_format"` // CSV, JSON
	Status        string    `db:"status" json:"status"`               // COMPLETED, PENDING, FAILED
	ExportContent *string   `db:"export_content" json:"export_content,omitempty"`
	FileName      *string   `db:"file_name" json:"file_name,omitempty"`
	CorrelationID string    `db:"correlation_id" json:"correlation_id"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// ReportDistributionRequest tracks external distribution with mandatory approval gating.
type ReportDistributionRequest struct {
	ID                  int64     `db:"id" json:"id"`
	OrgID               int64     `db:"org_id" json:"org_id"`
	UserID              int64     `db:"user_id" json:"user_id"`
	ReportType          string    `db:"report_type" json:"report_type"`
	RecipientEmails     string    `db:"recipient_emails" json:"recipient_emails"`
	DistributionChannel string    `db:"distribution_channel" json:"distribution_channel"`
	ApprovalID          *int64    `db:"approval_id" json:"approval_id,omitempty"`
	Status              string    `db:"status" json:"status"` // PENDING_APPROVAL, APPROVED, DISTRIBUTED, REJECTED
	Notes               *string   `db:"notes" json:"notes,omitempty"`
	CorrelationID       string    `db:"correlation_id" json:"correlation_id"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}

// DTOs for Python AI Sidecar Communication

type HistoricalDataPointDTO struct {
	Period   string                 `json:"period"`
	Value    float64                `json:"value"`
	Volume   *int                   `json:"volume,omitempty"`
	Unit     string                 `json:"unit,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ForecastDataPointDTO struct {
	Period          string  `json:"period"`
	ProjectedValue  float64 `json:"projected_value"`
	LowerBound      float64 `json:"lower_bound"`
	UpperBound      float64 `json:"upper_bound"`
	IsForecast      bool    `json:"is_forecast"`
	Label           string  `json:"label"`
	ConfidenceScore float64 `json:"confidence_score"`
}

type TrendInsightDTO struct {
	MetricName   string  `json:"metric_name"`
	Direction    string  `json:"direction"`
	ChangePct    float64 `json:"change_pct"`
	Significance string  `json:"significance"`
	Explanation  string  `json:"explanation"`
}

type AnomalyPointDTO struct {
	Period          string  `json:"period"`
	MetricName      string  `json:"metric_name"`
	ActualValue     float64 `json:"actual_value"`
	ExpectedValue   float64 `json:"expected_value"`
	DeviationPct    float64 `json:"deviation_pct"`
	Severity        string  `json:"severity"`
	SuspectedDriver string  `json:"suspected_driver"`
	Explanation     string  `json:"explanation"`
}

type SidecarReportRequest struct {
	OrgID                int64                    `json:"org_id"`
	ReportType           string                   `json:"report_type"`
	DateRange            string                   `json:"date_range"`
	HistoricalSeries     []HistoricalDataPointDTO `json:"historical_series"`
	AuthoritativeMetrics map[string]interface{}   `json:"authoritative_metrics"`
	CorrelationID        string                   `json:"correlation_id"`
}

type SidecarReportResponse struct {
	ReportType            string                 `json:"report_type"`
	DateRange             string                 `json:"date_range"`
	IsForecastAvailable   bool                   `json:"is_forecast_available"`
	ForecastSeries        []ForecastDataPointDTO `json:"forecast_series"`
	ForecastHorizon       string                 `json:"forecast_horizon"`
	ForecastAssumptions   []string               `json:"forecast_assumptions"`
	ForecastLimitations   []string               `json:"forecast_limitations"`
	ConfidenceScore       float64                `json:"confidence_score"`
	TrendInsights         []TrendInsightDTO      `json:"trend_insights"`
	Anomalies             []AnomalyPointDTO      `json:"anomalies"`
	ExecutiveNarrative    string                 `json:"executive_narrative"`
	ActionRecommendations []string               `json:"action_recommendations"`
	CorrelationID         string                 `json:"correlation_id"`
}

// User Request / Response DTOs

type AdvancedReportResponse struct {
	ReportType            string                   `json:"report_type"`
	DateRange             string                   `json:"date_range"`
	StartDate             time.Time                `json:"start_date"`
	EndDate               time.Time                `json:"end_date"`
	AuthoritativeMetrics  map[string]interface{}   `json:"authoritative_metrics"`
	HistoricalSeries      []HistoricalDataPointDTO `json:"historical_series"`
	IsForecastAvailable   bool                     `json:"is_forecast_available"`
	ForecastSeries        []ForecastDataPointDTO   `json:"forecast_series"`
	ForecastHorizon       string                   `json:"forecast_horizon"`
	ForecastAssumptions   []string                 `json:"forecast_assumptions"`
	ForecastLimitations   []string                 `json:"forecast_limitations"`
	ConfidenceScore       float64                  `json:"confidence_score"`
	TrendInsights         []TrendInsightDTO        `json:"trend_insights"`
	Anomalies             []AnomalyPointDTO        `json:"anomalies"`
	ExecutiveNarrative    string                   `json:"executive_narrative"`
	ActionRecommendations []string                 `json:"action_recommendations"`
	SnapshotID            int64                    `json:"snapshot_id"`
	CorrelationID         string                   `json:"correlation_id"`
}

type ExportReportInput struct {
	ReportType   string `json:"report_type"`
	DateRange    string `json:"date_range"`
	Format       string `json:"format"` // CSV or JSON
	ExportFormat string `json:"export_format"`
}

type DistributeReportInput struct {
	ReportType      string   `json:"report_type"`
	DateRange       string   `json:"date_range"`
	RecipientEmails []string `json:"recipient_emails"`
	Notes           string   `json:"notes,omitempty"`
}

func ToJSONString(v interface{}) *string {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}
