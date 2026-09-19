import React, { useState, useEffect, useCallback } from 'react';
import toast from 'react-hot-toast';
import {
  TrendingUp,
  Download,
  Send,
  History,
  AlertCircle,
  CheckCircle2,
  Clock,
  ShieldCheck,
  RefreshCw,
  FileSpreadsheet,
  FileCode,
  Sparkles,
  Info,
  Calendar,
  Layers,
  ArrowUpRight,
  ArrowDownRight,
  ChevronRight,
  X
} from 'lucide-react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  BarChart,
  Bar,
  AreaChart,
  Area
} from 'recharts';
import { reportingService } from '../../../services/reportingService';
import './ReportsPage.css';

const REPORT_TABS = [
  { id: 'OPERATIONAL_VOLUME', label: 'Operational Volume & Ops', icon: TrendingUp },
  { id: 'REVENUE_FINANCE', label: 'Revenue & Invoicing', icon: FileSpreadsheet },
  { id: 'COMMERCIAL_FUNNEL', label: 'Commercial Funnel', icon: Layers },
  { id: 'CONTRACT_COMPLIANCE', label: 'Contracts & Compliance', icon: ShieldCheck },
];

const DATE_RANGES = [
  { id: 'LAST_30D', label: 'Last 30 Days' },
  { id: 'LAST_90D', label: 'Last 90 Days' },
  { id: 'YTD', label: 'Year to Date' },
  { id: 'LAST_12M', label: 'Last 12 Months' },
];

export default function ReportsPage() {
  const [activeTab, setActiveTab] = useState('OPERATIONAL_VOLUME');
  const [dateRange, setDateRange] = useState('LAST_90D');
  const [loading, setLoading] = useState(true);
  const [reportData, setReportData] = useState(null);

  // Modals
  const [showDistributeModal, setShowDistributeModal] = useState(false);
  const [showHistoryModal, setShowHistoryModal] = useState(false);
  const [distributionForm, setDistributionForm] = useState({
    recipientEmails: '',
    channel: 'EMAIL',
    notes: '',
  });
  const [isDistributing, setIsDistributing] = useState(false);
  const [snapshotsHistory, setSnapshotsHistory] = useState([]);
  const [distributionsHistory, setDistributionsHistory] = useState([]);
  const [historyLoading, setHistoryLoading] = useState(false);

  const fetchReport = useCallback(async () => {
    setLoading(true);
    try {
      const data = await reportingService.getAdvancedReport(activeTab, dateRange);
      setReportData(data);
    } catch (err) {
      console.error('Failed to load advanced report:', err);
      toast.error('Failed to load authoritative report data.');
    } finally {
      setLoading(false);
    }
  }, [activeTab, dateRange]);

  useEffect(() => {
    fetchReport();
  }, [fetchReport]);

  const handleExport = async (format) => {
    try {
      toast.loading(`Generating authoritative ${format} export...`, { id: 'export-toast' });
      const res = await reportingService.exportReport({
        report_type: activeTab,
        export_format: format,
        date_range: dateRange,
      });

      // Trigger client-side download
      const mimeType = format === 'CSV' ? 'text/csv' : 'application/json';
      const blob = new Blob([res.export_content], { type: mimeType });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = res.file_name || `LogisticsHQ_${activeTab}.${format.toLowerCase()}`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      toast.success(`Export downloaded: ${res.file_name}`, { id: 'export-toast' });
    } catch (err) {
      console.error('Export failed:', err);
      toast.error('Failed to generate export', { id: 'export-toast' });
    }
  };

  const handleDistributionSubmit = async (e) => {
    e.preventDefault();
    if (!distributionForm.recipientEmails.trim()) {
      toast.error('Please specify at least one recipient email.');
      return;
    }

    const emails = distributionForm.recipientEmails
      .split(',')
      .map((e) => e.trim())
      .filter(Boolean);

    setIsDistributing(true);
    try {
      const res = await reportingService.requestDistribution({
        report_type: activeTab,
        recipient_emails: emails,
        distribution_channel: distributionForm.channel,
        notes: distributionForm.notes,
      });

      toast.success(
        `Distribution request submitted (Approval Request #${res.approval_id || res.id} created - Status: PENDING_APPROVAL). External delivery requires Human Approval.`,
        { duration: 6000 }
      );
      setShowDistributeModal(false);
      setDistributionForm({ recipientEmails: '', channel: 'EMAIL', notes: '' });
    } catch (err) {
      console.error('Distribution request failed:', err);
      toast.error('Failed to submit distribution request');
    } finally {
      setIsDistributing(false);
    }
  };

  const openHistoryModal = async () => {
    setShowHistoryModal(true);
    setHistoryLoading(true);
    try {
      const [snaps, dists] = await Promise.all([
        reportingService.getReportHistory(10),
        reportingService.getDistributionHistory(10),
      ]);
      setSnapshotsHistory(snaps || []);
      setDistributionsHistory(dists || []);
    } catch (err) {
      console.error('Failed to load history:', err);
      toast.error('Could not load historical records');
    } finally {
      setHistoryLoading(false);
    }
  };

  // Combine historical and forecast series for seamless unified Recharts visualization
  const buildChartData = () => {
    if (!reportData) return [];
    const chartMap = new Map();

    (reportData.historical_series || []).forEach((pt) => {
      chartMap.set(pt.period, {
        period: pt.period,
        actual: pt.value,
        forecast: null,
        lower_bound: null,
        upper_bound: null,
        status: pt.status || 'MEASURED',
      });
    });

    if (reportData.is_forecast_available && reportData.forecast_series) {
      reportData.forecast_series.forEach((fc) => {
        const existing = chartMap.get(fc.period) || {
          period: fc.period,
          actual: null,
          status: 'PROJECTED',
        };
        existing.forecast = fc.forecast_value;
        existing.lower_bound = fc.lower_bound;
        existing.upper_bound = fc.upper_bound;
        chartMap.set(fc.period, existing);
      });
    }

    return Array.from(chartMap.values());
  };

  const chartData = buildChartData();

  // Render Metric Cards depending on the active report type
  const renderAuthoritativeMetrics = () => {
    if (!reportData || !reportData.authoritative_metrics) return null;
    const m = reportData.authoritative_metrics;

    if (activeTab === 'OPERATIONAL_VOLUME') {
      return (
        <div className="metrics-strip">
          <div className="metric-card">
            <span className="metric-label">Total Shipments</span>
            <span className="metric-value">{m.total_shipments ?? 0}</span>
            <span className="fact-badge">Measured Fact • MariaDB</span>
          </div>
          <div className="metric-card">
            <span className="metric-label">Delivered Shipments</span>
            <span className="metric-value text-emerald">{m.delivered ?? 0}</span>
            <span className="fact-badge">Measured Fact • MariaDB</span>
          </div>
          <div className="metric-card">
            <span className="metric-label">In-Transit Freight</span>
            <span className="metric-value text-blue">{m.in_transit ?? 0}</span>
            <span className="fact-badge">Measured Fact • MariaDB</span>
          </div>
          <div className="metric-card">
            <span className="metric-label">On-Time Delivery Rate</span>
            <span className="metric-value">{m.on_time_delivery_rate || '0.0%'}</span>
            <span className="fact-badge">Deterministic Calculation</span>
          </div>
        </div>
      );
    }

    if (activeTab === 'REVENUE_FINANCE') {
      return (
        <div className="metrics-strip">
          <div className="metric-card">
            <span className="metric-label">Total Invoiced</span>
            <span className="metric-value text-blue">
              ${(m.total_invoiced_usd || 0).toLocaleString(undefined, { minimumFractionDigits: 2 })}
            </span>
            <span className="fact-badge">Measured Fact • MariaDB</span>
          </div>
          <div className="metric-card">
            <span className="metric-label">Total Collected</span>
            <span className="metric-value text-emerald">
              ${(m.total_collected_usd || 0).toLocaleString(undefined, { minimumFractionDigits: 2 })}
            </span>
            <span className="fact-badge">Measured Fact • MariaDB</span>
          </div>
          <div className="metric-card">
            <span className="metric-label">Outstanding Receivables</span>
            <span className="metric-value text-amber">
              ${(m.outstanding_receivables_usd || 0).toLocaleString(undefined, { minimumFractionDigits: 2 })}
            </span>
            <span className="fact-badge">Measured Fact • MariaDB</span>
          </div>
          <div className="metric-card">
            <span className="metric-label">Invoices Issued</span>
            <span className="metric-value">{m.total_invoices_issued ?? 0}</span>
            <span className="fact-badge">Deterministic Count</span>
          </div>
        </div>
      );
    }

    if (activeTab === 'COMMERCIAL_FUNNEL') {
      return (
        <div className="metrics-strip">
          <div className="metric-card">
            <span className="metric-label">Total Leads</span>
            <span className="metric-value">{m.total_leads ?? 0}</span>
            <span className="fact-badge">Measured Fact • MariaDB</span>
          </div>
          <div className="metric-card">
            <span className="metric-label">RFQs Generated</span>
            <span className="metric-value">{m.total_rfqs ?? 0}</span>
            <span className="fact-badge">Measured Fact • MariaDB</span>
          </div>
          <div className="metric-card">
            <span className="metric-label">Quotations Won</span>
            <span className="metric-value text-emerald">{m.total_won_quotes ?? 0}</span>
            <span className="fact-badge">Measured Fact • MariaDB</span>
          </div>
          <div className="metric-card">
            <span className="metric-label">Overall Win Rate</span>
            <span className="metric-value text-blue">{m.overall_win_rate || '0.0%'}</span>
            <span className="fact-badge">Deterministic Conversion</span>
          </div>
        </div>
      );
    }

    if (activeTab === 'CONTRACT_COMPLIANCE') {
      return (
        <div className="metrics-strip">
          <div className="metric-card">
            <span className="metric-label">Active Master Contracts</span>
            <span className="metric-value text-emerald">{m.active_contracts ?? 0}</span>
            <span className="fact-badge">Measured Fact • MariaDB</span>
          </div>
          <div className="metric-card">
            <span className="metric-label">Expiring within 30 Days</span>
            <span className="metric-value text-amber">{m.expiring_within_30d ?? 0}</span>
            <span className="fact-badge">Deterministic Calculation</span>
          </div>
          <div className="metric-card">
            <span className="metric-label">Expiring within 60 Days</span>
            <span className="metric-value">{m.expiring_within_60d ?? 0}</span>
            <span className="fact-badge">Deterministic Calculation</span>
          </div>
          <div className="metric-card">
            <span className="metric-label">Compliance Audits Done</span>
            <span className="metric-value">{m.compliance_reviews_completed ?? 0}</span>
            <span className="fact-badge">Audit Verification</span>
          </div>
        </div>
      );
    }

    return null;
  };

  return (
    <div className="reports-page-container">
      {/* Header & Main Controls */}
      <div className="reports-top-header">
        <div className="reports-title-area">
          <div className="title-row">
            <h1>Advanced Reporting & Forecasting</h1>
            <span className="system-pill">Authoritative Analytics</span>
          </div>
          <p>
            Deterministic business metrics from MariaDB combined with calibrated statistical forecasting from the Python AI Sidecar.
          </p>
        </div>

        <div className="reports-header-actions">
          <button className="btn-secondary" onClick={() => handleExport('CSV')} title="Download Authoritative CSV">
            <FileSpreadsheet size={16} />
            <span>Export CSV</span>
          </button>
          <button className="btn-secondary" onClick={() => handleExport('JSON')} title="Download Authoritative JSON">
            <FileCode size={16} />
            <span>Export JSON</span>
          </button>
          <button
            className="btn-primary"
            onClick={() => setShowDistributeModal(true)}
            title="External distribution requires Human Approval"
          >
            <Send size={16} />
            <span>Distribute Report (HITL Gate)</span>
          </button>
          <button className="btn-icon" onClick={openHistoryModal} title="View Snapshot & Distribution History">
            <History size={18} />
          </button>
          <button className="btn-icon" onClick={fetchReport} title="Refresh Live Data">
            <RefreshCw size={18} className={loading ? 'spinning' : ''} />
          </button>
        </div>
      </div>

      {/* Module Tabs & Date Range Filter Bar */}
      <div className="reports-controls-bar">
        <div className="report-tabs">
          {REPORT_TABS.map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                className={`tab-button ${isActive ? 'active' : ''}`}
                onClick={() => setActiveTab(tab.id)}
              >
                <Icon size={16} />
                <span>{tab.label}</span>
              </button>
            );
          })}
        </div>

        <div className="date-range-selector">
          <Calendar size={16} className="calendar-icon" />
          <select
            value={dateRange}
            onChange={(e) => setDateRange(e.target.value)}
            className="date-select"
          >
            {DATE_RANGES.map((r) => (
              <option key={r.id} value={r.id}>
                {r.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Main Body */}
      {loading ? (
        <div className="reports-loading-state">
          <div className="spinner"></div>
          <p>Compiling authoritative business metrics & querying forecasting sidecar...</p>
        </div>
      ) : (
        <div className="reports-grid">
          {/* Authoritative Metric Strip */}
          {renderAuthoritativeMetrics()}

          {/* Unified Chart Card: Actual Historical vs Model Forecast */}
          <div className="report-card chart-card">
            <div className="card-header-row">
              <div>
                <h3>
                  {REPORT_TABS.find((t) => t.id === activeTab)?.label} Trend & Calibrated Projection
                </h3>
                <p className="card-subtitle">
                  Historical actuals from MariaDB plotted alongside Python AI statistical projections with confidence intervals.
                </p>
              </div>

              <div className="chart-legend-indicators">
                <span className="legend-item">
                  <span className="legend-dot actual-dot"></span> Actual Measured (Fact)
                </span>
                <span className="legend-item">
                  <span className="legend-dot forecast-dot"></span> Projected Forecast (Model-generated)
                </span>
              </div>
            </div>

            {chartData.length === 0 ? (
              <div className="chart-empty-state">
                <Info size={32} />
                <p>No activity records found in MariaDB for {dateRange}.</p>
                <span className="empty-sub">Once records are processed, authoritative trend lines will render here.</span>
              </div>
            ) : (
              <div className="chart-wrapper">
                <ResponsiveContainer width="100%" height={320}>
                  <LineChart data={chartData} margin={{ top: 20, right: 30, left: 10, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#f1f5f9" />
                    <XAxis dataKey="period" stroke="#64748b" tick={{ fontSize: 12 }} />
                    <YAxis stroke="#64748b" tick={{ fontSize: 12 }} />
                    <Tooltip
                      formatter={(val, name, item) => {
                        if (name === 'actual') return [val ?? 'N/A', 'Actual (Fact)'];
                        if (name === 'forecast') {
                          const lb = item.payload.lower_bound ?? 'N/A';
                          const ub = item.payload.upper_bound ?? 'N/A';
                          return [`${val} (Range: ${lb} - ${ub})`, 'Model Forecast [Estimate]'];
                        }
                        return [val, name];
                      }}
                      contentStyle={{
                        backgroundColor: '#ffffff',
                        border: '1px solid #e2e8f0',
                        borderRadius: '8px',
                        boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
                      }}
                    />
                    {/* Actual Measured Line */}
                    <Line
                      type="monotone"
                      dataKey="actual"
                      stroke="#2563eb"
                      strokeWidth={2.5}
                      dot={{ r: 4, fill: '#2563eb' }}
                      activeDot={{ r: 6 }}
                      name="actual"
                      connectNulls={false}
                    />
                    {/* Forecast Model Line (Dashed) */}
                    <Line
                      type="monotone"
                      dataKey="forecast"
                      stroke="#d97706"
                      strokeWidth={2}
                      strokeDasharray="5 5"
                      dot={{ r: 4, fill: '#d97706' }}
                      name="forecast"
                      connectNulls={false}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            )}

            {/* Safeguard Notice when Forecast is Unavailable */}
            {reportData && !reportData.is_forecast_available && (
              <div className="forecast-safeguard-banner">
                <AlertCircle size={18} className="safeguard-icon" />
                <div className="safeguard-text">
                  <strong>Statistical Forecasting Unavailable: Insufficient Historical Series</strong>
                  <p>
                    LogisticsHQ safety guidelines require at least 3 historical periods to calculate statistical projections without fabrication.
                    As more activity is logged over time, projected lines will automatically calibrate.
                  </p>
                </div>
              </div>
            )}

            {/* Forecast Assumptions & Horizon when available */}
            {reportData && reportData.is_forecast_available && (
              <div className="forecast-active-banner">
                <Sparkles size={18} className="sparkle-icon" />
                <div className="forecast-banner-details">
                  <div className="banner-top">
                    <strong>
                      Model Projection Active (Horizon: {reportData.forecast_horizon || 'Next 3 Periods'})
                    </strong>
                    <span className="confidence-pill">
                      Confidence: {Math.round((reportData.forecast_confidence || 0) * 100)}%
                    </span>
                  </div>
                  {reportData.forecast_assumptions && reportData.forecast_assumptions.length > 0 && (
                    <div className="assumptions-list">
                      <em>Assumptions:</em> {reportData.forecast_assumptions.join(' • ')}
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>

          {/* AI Executive Narrative & Trend Insights Card */}
          <div className="report-card narrative-card">
            <div className="narrative-header">
              <div className="narrative-badge-group">
                <span className="ai-source-badge">
                  <Sparkles size={14} /> AI Narrative & Trend Insights
                </span>
                <span className="ai-runtime-pill">Python AI Sidecar</span>
                <span className="fact-check-badge">
                  <ShieldCheck size={14} /> Grounded in Real Data
                </span>
              </div>

              <div className="model-notice">
                <Info size={14} />
                <span>Model-generated interpretation. Forecasts and recommendations are estimates, not facts.</span>
              </div>
            </div>

            {/* Executive Narrative */}
            <div className="executive-narrative-box">
              <h4>Executive Summary</h4>
              <p>{reportData?.executive_narrative || 'No analytical commentary available for this reporting window.'}</p>
            </div>

            {/* Two-column insights: Observed Trends & Detected Anomalies */}
            <div className="insights-columns">
              <div className="insight-column">
                <h4>
                  <TrendingUp size={16} /> Observed Statistical Trends
                </h4>
                {reportData?.trend_insights && reportData.trend_insights.length > 0 ? (
                  <div className="trend-items-list">
                    {reportData.trend_insights.map((trend, idx) => (
                      <div key={idx} className="trend-item">
                        <div className="trend-item-top">
                          <span className="trend-metric">{trend.metric}</span>
                          <span
                            className={`trend-direction-badge ${
                              trend.direction === 'UP' ? 'up' : trend.direction === 'DOWN' ? 'down' : 'stable'
                            }`}
                          >
                            {trend.direction === 'UP' ? (
                              <ArrowUpRight size={14} />
                            ) : trend.direction === 'DOWN' ? (
                              <ArrowDownRight size={14} />
                            ) : (
                              '•'
                            )}
                            {trend.change_pct ? `${trend.change_pct}%` : trend.direction}
                          </span>
                        </div>
                        <p className="trend-driver">{trend.driver}</p>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="empty-sub">No distinct trend movements detected in current sample.</p>
                )}
              </div>

              <div className="insight-column">
                <h4>
                  <AlertCircle size={16} /> Detected Statistical Anomalies
                </h4>
                {reportData?.anomalies && reportData.anomalies.length > 0 ? (
                  <div className="anomaly-items-list">
                    {reportData.anomalies.map((anom, idx) => (
                      <div key={idx} className="anomaly-item">
                        <div className="anomaly-item-top">
                          <span className="anomaly-period">{anom.period}</span>
                          <span className="anomaly-deviation">
                            {anom.deviation_pct > 0 ? `+${anom.deviation_pct}%` : `${anom.deviation_pct}%`}
                          </span>
                        </div>
                        <p className="anomaly-explanation">{anom.explanation}</p>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="no-anomalies-box">
                    <CheckCircle2 size={16} className="text-emerald" />
                    <span>All observed data points fall within standard operational tolerances (&lt;28% variance).</span>
                  </div>
                )}
              </div>
            </div>

            {/* Controlled Action System Recommendations */}
            {reportData?.recommended_actions && reportData.recommended_actions.length > 0 && (
              <div className="recommended-actions-section">
                <h4>Action System Recommendations</h4>
                <div className="recommended-actions-grid">
                  {reportData.recommended_actions.map((act, idx) => (
                    <div key={idx} className="action-recommendation-card">
                      <div className="action-card-header">
                        <span className={`priority-tag ${act.priority?.toLowerCase() || 'medium'}`}>
                          {act.priority || 'MEDIUM'}
                        </span>
                        <span className="action-type-label">{act.action_type}</span>
                      </div>
                      <h5>{act.title}</h5>
                      <p className="action-rationale">{act.rationale}</p>
                      {act.potential_impact && (
                        <div className="action-impact">
                          <strong>Impact:</strong> {act.potential_impact}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* External Report Distribution Modal (Human-in-the-Loop Gate) */}
      {showDistributeModal && (
        <div className="modal-backdrop">
          <div className="modal-dialog">
            <div className="modal-header">
              <div className="modal-title-wrap">
                <Send size={20} className="modal-icon" />
                <h3>Distribute Report Externally</h3>
              </div>
              <button className="btn-close" onClick={() => setShowDistributeModal(false)}>
                <X size={18} />
              </button>
            </div>

            <div className="approval-gate-banner">
              <ShieldCheck size={18} className="shield-icon" />
              <div>
                <strong>Human-in-the-Loop Governance Enforced</strong>
                <p>
                  To prevent unauthorized data leakage, all external report transmissions require review and approval
                  in the LogisticsHQ Human Approval Center prior to delivery.
                </p>
              </div>
            </div>

            <form onSubmit={handleDistributionSubmit} className="modal-form">
              <div className="form-group">
                <label>Report Type</label>
                <input type="text" value={activeTab} disabled className="form-input disabled" />
              </div>

              <div className="form-group">
                <label>Recipient Emails (comma-separated)</label>
                <input
                  type="text"
                  placeholder="e.g. board@acme.com, vp.finance@acme.com"
                  value={distributionForm.recipientEmails}
                  onChange={(e) =>
                    setDistributionForm({ ...distributionForm, recipientEmails: e.target.value })
                  }
                  required
                  className="form-input"
                />
              </div>

              <div className="form-group">
                <label>Distribution Channel</label>
                <select
                  value={distributionForm.channel}
                  onChange={(e) =>
                    setDistributionForm({ ...distributionForm, channel: e.target.value })
                  }
                  className="form-input"
                >
                  <option value="EMAIL">Email (Encrypted PDF/CSV Attachment)</option>
                  <option value="SECURE_LINK">Time-limited Secure Portal Link</option>
                </select>
              </div>

              <div className="form-group">
                <label>Justification & Distribution Notes</label>
                <textarea
                  rows={3}
                  placeholder="State the business rationale for transmitting this report to external recipients..."
                  value={distributionForm.notes}
                  onChange={(e) => setDistributionForm({ ...distributionForm, notes: e.target.value })}
                  className="form-input"
                />
              </div>

              <div className="modal-footer">
                <button
                  type="button"
                  className="btn-secondary"
                  onClick={() => setShowDistributeModal(false)}
                >
                  Cancel
                </button>
                <button type="submit" className="btn-primary" disabled={isDistributing}>
                  {isDistributing ? 'Submitting to Approval Center...' : 'Submit for Human Approval'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Snapshot & Distribution History Modal */}
      {showHistoryModal && (
        <div className="modal-backdrop">
          <div className="modal-dialog modal-wide">
            <div className="modal-header">
              <div className="modal-title-wrap">
                <History size={20} className="modal-icon" />
                <h3>Snapshot & Distribution Audit History</h3>
              </div>
              <button className="btn-close" onClick={() => setShowHistoryModal(false)}>
                <X size={18} />
              </button>
            </div>

            <div className="history-modal-content">
              {historyLoading ? (
                <div className="reports-loading-state">
                  <div className="spinner"></div>
                  <p>Loading historical records...</p>
                </div>
              ) : (
                <div className="history-sections">
                  {/* Snapshots Table */}
                  <div className="history-subcard">
                    <h4>Authoritative Report Snapshots</h4>
                    {snapshotsHistory.length === 0 ? (
                      <p className="empty-sub">No snapshots generated yet.</p>
                    ) : (
                      <div className="table-responsive">
                        <table className="audit-table">
                          <thead>
                            <tr>
                              <th>ID</th>
                              <th>Report Type</th>
                              <th>Date Range</th>
                              <th>Forecast Available</th>
                              <th>Correlation ID</th>
                            </tr>
                          </thead>
                          <tbody>
                            {snapshotsHistory.map((snap) => (
                              <tr key={snap.id}>
                                <td>#{snap.id}</td>
                                <td>
                                  <span className="type-pill">{snap.report_type}</span>
                                </td>
                                <td>{snap.date_range}</td>
                                <td>
                                  {snap.is_forecast_available ? (
                                    <span className="status-pill green">Yes</span>
                                  ) : (
                                    <span className="status-pill gray">No</span>
                                  )}
                                </td>
                                <td className="font-mono">{snap.correlation_id}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </div>

                  {/* Distribution Requests Table */}
                  <div className="history-subcard">
                    <h4>External Distribution Requests & HITL Status</h4>
                    {distributionsHistory.length === 0 ? (
                      <p className="empty-sub">No external distribution requests recorded.</p>
                    ) : (
                      <div className="table-responsive">
                        <table className="audit-table">
                          <thead>
                            <tr>
                              <th>ID</th>
                              <th>Report</th>
                              <th>Recipients</th>
                              <th>Approval Status</th>
                              <th>Approval ID</th>
                            </tr>
                          </thead>
                          <tbody>
                            {distributionsHistory.map((dist) => (
                              <tr key={dist.id}>
                                <td>#{dist.id}</td>
                                <td>{dist.report_type}</td>
                                <td>{dist.recipient_emails}</td>
                                <td>
                                  <span
                                    className={`status-pill ${
                                      dist.status === 'APPROVED'
                                        ? 'green'
                                        : dist.status === 'PENDING_APPROVAL'
                                        ? 'amber'
                                        : 'gray'
                                    }`}
                                  >
                                    {dist.status}
                                  </span>
                                </td>
                                <td>
                                  {dist.approval_id ? (
                                    <span className="approval-link">Request #{dist.approval_id}</span>
                                  ) : (
                                    'N/A'
                                  )}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>

            <div className="modal-footer">
              <button className="btn-secondary" onClick={() => setShowHistoryModal(false)}>
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
