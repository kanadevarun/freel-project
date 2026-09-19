from typing import List, Optional, Dict, Any
from pydantic import BaseModel, Field

class HistoricalDataPoint(BaseModel):
    period: str
    value: float
    volume: Optional[int] = None
    unit: Optional[str] = "USD"
    metadata: Optional[Dict[str, Any]] = None

class ForecastDataPoint(BaseModel):
    period: str
    projected_value: float
    lower_bound: float
    upper_bound: float
    is_forecast: bool = True
    label: str = "Model-Generated Forecast"
    confidence_score: float = Field(ge=0.0, le=1.0)

class AnomalyPoint(BaseModel):
    period: str
    metric_name: str
    actual_value: float
    expected_value: float
    deviation_pct: float
    severity: str = "MEDIUM"  # "LOW", "MEDIUM", "HIGH"
    suspected_driver: str
    explanation: str

class TrendInsight(BaseModel):
    metric_name: str
    direction: str  # "INCREASING", "DECREASING", "STABLE"
    change_pct: float
    significance: str  # "HIGH", "MODERATE", "LOW"
    explanation: str

class ReportingForecastRequest(BaseModel):
    org_id: int
    report_type: str  # "OPERATIONAL_VOLUME", "REVENUE_FINANCE", "COMMERCIAL_FUNNEL", "CONTRACT_COMPLIANCE"
    date_range: str   # "LAST_30D", "LAST_90D", "YTD", "LAST_12M", "CUSTOM"
    historical_series: List[HistoricalDataPoint]
    authoritative_metrics: Dict[str, Any] = {}
    correlation_id: str

class ReportingForecastResponse(BaseModel):
    report_type: str
    date_range: str
    is_forecast_available: bool
    forecast_series: List[ForecastDataPoint] = []
    forecast_horizon: str
    forecast_assumptions: List[str] = []
    forecast_limitations: List[str] = []
    confidence_score: float = Field(ge=0.0, le=1.0)
    trend_insights: List[TrendInsight] = []
    anomalies: List[AnomalyPoint] = []
    executive_narrative: str
    action_recommendations: List[str] = []
    correlation_id: str
