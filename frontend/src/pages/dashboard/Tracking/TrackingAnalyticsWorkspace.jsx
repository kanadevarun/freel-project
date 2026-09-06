import React, { useState, useEffect } from 'react';
import {
  Activity,
  TrendingUp,
  AlertTriangle,
  Clock,
  CheckCircle2,
  XCircle,
  Radio,
  RefreshCw,
  Layers,
  ArrowUpRight,
  ArrowDownRight,
  ShieldAlert,
  Ship,
  Navigation,
  Compass,
  Zap,
  Info,
  ChevronRight,
  Sparkles,
} from 'lucide-react';
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  BarChart,
  Bar,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
} from 'recharts';
import shipmentService from '../../../services/shipmentService';
import './TrackingAnalyticsWorkspace.css';

export default function TrackingAnalyticsWorkspace({ onNavigateShipment }) {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [timeRange, setTimeRange] = useState(14); // 7, 14, 30 days

  // Analytics State
  const [overview, setOverview] = useState(null);
  const [trends, setTrends] = useState([]);
  const [carriers, setCarriers] = useState([]);
  const [routes, setRoutes] = useState([]);
  const [carrierSearch, setCarrierSearch] = useState('');
  const [carrierSort, setCarrierSort] = useState('score'); // 'score', 'ontime', 'volume'

  useEffect(() => {
    loadAnalyticsData(timeRange);
  }, [timeRange]);

  const loadAnalyticsData = async (days) => {
    try {
      setLoading(true);
      setError(null);

      const [overviewRes, trendsRes, carriersRes, routesRes] = await Promise.allSettled([
        shipmentService.getTrackingAnalyticsOverview(),
        shipmentService.getTrackingAnalyticsTrends({ days }),
        shipmentService.getCarrierTrackingPerformance(),
        shipmentService.getRouteTrackingPerformance(),
      ]);

      if (overviewRes.status === 'fulfilled' && overviewRes.value?.data) {
        setOverview(overviewRes.value.data);
      }
      if (trendsRes.status === 'fulfilled' && Array.isArray(trendsRes.value?.data)) {
        setTrends(trendsRes.value.data);
      }
      if (carriersRes.status === 'fulfilled' && Array.isArray(carriersRes.value?.data)) {
        setCarriers(carriersRes.value.data);
      }
      if (routesRes.status === 'fulfilled' && Array.isArray(routesRes.value?.data)) {
        setRoutes(routesRes.value.data);
      }
    } catch (err) {
      console.error('Failed to load tracking analytics:', err);
      setError('Unable to aggregate tracking intelligence. Please retry.');
    } finally {
      setLoading(false);
    }
  };

  // Filter & sort carriers
  const filteredCarriers = carriers
    .filter(
      (c) =>
        c.carrier_name.toLowerCase().includes(carrierSearch.toLowerCase()) ||
        c.carrier_scac.toLowerCase().includes(carrierSearch.toLowerCase())
    )
    .sort((a, b) => {
      if (carrierSort === 'ontime') return b.on_time_rate - a.on_time_rate;
      if (carrierSort === 'volume') return b.shipments_tracked - a.shipments_tracked;
      return b.reliability_score - a.reliability_score;
    });

  if (loading && !overview) {
    return (
      <div className="trka-loading-container">
        <RefreshCw size={28} className="trka-spin" />
        <span className="trka-loading-text">Computing fleet performance intelligence & telemetry metrics...</span>
      </div>
    );
  }

  if (error && !overview) {
    return (
      <div className="trka-error-container">
        <AlertTriangle size={32} className="trka-err-icon" />
        <p className="trka-err-msg">{error}</p>
        <button className="trka-btn-primary" onClick={() => loadAnalyticsData(timeRange)}>
          <RefreshCw size={14} /> Retry Aggregation
        </button>
      </div>
    );
  }

  return (
    <div className="trka-workspace">
      {/* ── Top Header Controls ──────────────────────────────────────────── */}
      <div className="trka-header-bar">
        <div className="trka-hb-left">
          <div className="trka-title-group">
            <h2 className="trka-heading">Tracking Performance & Operational Intelligence</h2>
            <span className="trka-subheading">
              Deterministic fleet reliability, carrier scorecards, route corridors, and telemetry health
            </span>
          </div>
        </div>
        <div className="trka-hb-right">
          <div className="trka-range-pills">
            <button
              className={`trka-range-btn ${timeRange === 7 ? 'active' : ''}`}
              onClick={() => setTimeRange(7)}
            >
              7 Days
            </button>
            <button
              className={`trka-range-btn ${timeRange === 14 ? 'active' : ''}`}
              onClick={() => setTimeRange(14)}
            >
              14 Days
            </button>
            <button
              className={`trka-range-btn ${timeRange === 30 ? 'active' : ''}`}
              onClick={() => setTimeRange(30)}
            >
              30 Days
            </button>
          </div>
          <button className="trka-btn-refresh" onClick={() => loadAnalyticsData(timeRange)} title="Refresh Intelligence">
            <RefreshCw size={13} className={loading ? 'trka-spin' : ''} />
            <span>Sync</span>
          </button>
        </div>
      </div>

      {/* ── 1. Executive KPI Strip ────────────────────────────────────────── */}
      {overview && (
        <div className="trka-kpi-grid">
          {/* Active Shipments */}
          <div className="trka-kpi-card">
            <div className="trka-kpi-header">
              <span className="trka-kpi-title">ACTIVE TRACKED</span>
              <div className="trka-kpi-icon-wrap blue">
                <Ship size={16} />
              </div>
            </div>
            <div className="trka-kpi-body">
              <div className="trka-kpi-val">{overview.active_shipments}</div>
              <div className="trka-kpi-sub">
                <span>{overview.total_tracked_shipments} lifetime shipments recorded</span>
              </div>
            </div>
          </div>

          {/* On-Time Reliability */}
          <div className="trka-kpi-card">
            <div className="trka-kpi-header">
              <span className="trka-kpi-title">ON-TIME RELIABILITY</span>
              <div className="trka-kpi-icon-wrap green">
                <CheckCircle2 size={16} />
              </div>
            </div>
            <div className="trka-kpi-body">
              <div className="trka-kpi-val green">
                {overview.on_time_percentage.toFixed(1)}%
              </div>
              <div className="trka-kpi-sub">
                <span className="trka-pill-sub green">{overview.on_schedule_shipments} On Schedule</span>
                {overview.early_shipments > 0 && (
                  <span className="trka-pill-sub blue">{overview.early_shipments} Early</span>
                )}
              </div>
            </div>
          </div>

          {/* Delayed Shipments */}
          <div className="trka-kpi-card">
            <div className="trka-kpi-header">
              <span className="trka-kpi-title">DELAYED SHIPMENTS</span>
              <div className="trka-kpi-icon-wrap amber">
                <Clock size={16} />
              </div>
            </div>
            <div className="trka-kpi-body">
              <div className={`trka-kpi-val ${overview.delayed_shipments > 0 ? 'amber' : 'neutral'}`}>
                {overview.delayed_shipments}
              </div>
              <div className="trka-kpi-sub">
                <span>
                  {overview.average_delay_hours > 0
                    ? `Avg delay ~${overview.average_delay_hours.toFixed(1)}h`
                    : 'Zero critical delays'}
                </span>
              </div>
            </div>
          </div>

          {/* Schedule ETA Variance */}
          <div className="trka-kpi-card">
            <div className="trka-kpi-header">
              <span className="trka-kpi-title">AVG SCHEDULE VARIANCE</span>
              <div className="trka-kpi-icon-wrap purple">
                <Compass size={16} />
              </div>
            </div>
            <div className="trka-kpi-body">
              <div className="trka-kpi-val purple">
                ±{(overview.average_eta_variance_hours || 0).toFixed(1)}d
              </div>
              <div className="trka-kpi-sub">
                <span>Transit milestone accuracy</span>
              </div>
            </div>
          </div>

          {/* Open Critical Alerts */}
          <div className="trka-kpi-card">
            <div className="trka-kpi-header">
              <span className="trka-kpi-title">CRITICAL ALERTS</span>
              <div className={`trka-kpi-icon-wrap ${overview.open_critical_alerts > 0 ? 'red' : 'green'}`}>
                <ShieldAlert size={16} />
              </div>
            </div>
            <div className="trka-kpi-body">
              <div className={`trka-kpi-val ${overview.open_critical_alerts > 0 ? 'red' : 'green'}`}>
                {overview.open_critical_alerts}
              </div>
              <div className="trka-kpi-sub">
                <span>{overview.total_open_alerts} total open operational alerts</span>
              </div>
            </div>
          </div>

          {/* Refresh Success & Data Freshness */}
          <div className="trka-kpi-card">
            <div className="trka-kpi-header">
              <span className="trka-kpi-title">SYNC SUCCESS (30D)</span>
              <div className="trka-kpi-icon-wrap teal">
                <Radio size={16} />
              </div>
            </div>
            <div className="trka-kpi-body">
              <div className="trka-kpi-val teal">
                {overview.refresh_success_rate.toFixed(1)}%
              </div>
              <div className="trka-kpi-sub">
                <span className="trka-fresh-dot live" title="Live (<6h)" /> {overview.data_freshness_live} Live
                <span className="trka-fresh-dot recent" title="Recent (<24h)" /> {overview.data_freshness_recent} Recent
                <span className="trka-fresh-dot stale" title="Stale (>24h)" /> {overview.data_freshness_stale} Stale
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ── 2. Operational Insights Panel (Deterministic AI Engine) ────────── */}
      {overview?.insights && overview.insights.length > 0 && (
        <div className="trka-insights-section">
          <div className="trka-section-header">
            <div className="trka-sh-title-box">
              <Sparkles size={16} className="trka-sparkle-icon" />
              <h3 className="trka-section-title">Operational Intelligence & Risk Recommendations</h3>
              <span className="trka-badge-ai">Deterministic Calculation</span>
            </div>
            <span className="trka-insights-count">{overview.insights.length} Actionable Observations</span>
          </div>

          <div className="trka-insights-grid">
            {overview.insights.map((ins) => (
              <div key={ins.id} className={`trka-insight-card sev-${ins.severity.toLowerCase()}`}>
                <div className="trka-ic-top">
                  <div className="trka-ic-badge-wrap">
                    <span className={`trka-sev-pill sev-${ins.severity.toLowerCase()}`}>
                      {ins.severity}
                    </span>
                    <span className="trka-cat-pill">{ins.category.replace('_', ' ')}</span>
                  </div>
                  <span className="trka-ic-metric">{ins.metric}</span>
                </div>

                <h4 className="trka-ic-title">{ins.title}</h4>
                <p className="trka-ic-desc">{ins.description}</p>

                <div className="trka-ic-rec">
                  <span className="trka-ic-rec-label">RECOMMENDED ACTION:</span>
                  <p className="trka-ic-rec-text">{ins.recommendation}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ── 3. Visual Performance Trend Charts (Recharts) ─────────────────── */}
      <div className="trka-charts-section">
        <div className="trka-section-header">
          <div className="trka-sh-title-box">
            <TrendingUp size={16} className="trka-chart-icon" />
            <h3 className="trka-section-title">Fleet Telemetry & Reliability Trends ({timeRange} Days)</h3>
          </div>
        </div>

        <div className="trka-charts-grid">
          {/* Chart 1: On-Time Delivery Rate & Delayed Volume */}
          <div className="trka-chart-card">
            <div className="trka-cc-header">
              <span className="trka-cc-title">On-Time Performance & Delays</span>
              <span className="trka-cc-sub">Daily reliability vs delayed shipment count</span>
            </div>
            <div className="trka-cc-body">
              {trends.length === 0 ? (
                <div className="trka-empty-chart">No historical trend data in this period.</div>
              ) : (
                <ResponsiveContainer width="100%" height={240}>
                  <AreaChart data={trends} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                    <defs>
                      <linearGradient id="ontimeGrad" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#10b981" stopOpacity={0.4} />
                        <stop offset="95%" stopColor="#10b981" stopOpacity={0.0} />
                      </linearGradient>
                      <linearGradient id="delayGrad" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#f59e0b" stopOpacity={0.4} />
                        <stop offset="95%" stopColor="#f59e0b" stopOpacity={0.0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#f1f5f9" />
                    <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#64748b' }} tickFormatter={(d) => d.slice(5)} />
                    <YAxis yAxisId="left" domain={[0, 100]} tick={{ fontSize: 11, fill: '#64748b' }} unit="%" />
                    <YAxis yAxisId="right" orientation="right" tick={{ fontSize: 11, fill: '#64748b' }} allowDecimals={false} />
                    <Tooltip
                      contentStyle={{ background: '#0f172a', border: 'none', borderRadius: '6px', color: '#fff', fontSize: '0.75rem' }}
                      formatter={(val, name) => [name === 'On-Time Rate' ? `${Number(val).toFixed(1)}%` : val, name]}
                    />
                    <Legend wrapperStyle={{ fontSize: '0.75rem', paddingTop: '10px' }} />
                    <Area
                      yAxisId="left"
                      type="monotone"
                      dataKey="on_time_rate"
                      name="On-Time Rate"
                      stroke="#10b981"
                      strokeWidth={2}
                      fillOpacity={1}
                      fill="url(#ontimeGrad)"
                    />
                    <Line
                      yAxisId="right"
                      type="monotone"
                      dataKey="delayed_count"
                      name="Delayed Shipments"
                      stroke="#f59e0b"
                      strokeWidth={2}
                      dot={{ r: 3 }}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              )}
            </div>
          </div>

          {/* Chart 2: Daily Alert Volume & Refresh Reliability */}
          <div className="trka-chart-card">
            <div className="trka-cc-header">
              <span className="trka-cc-title">Tracking Alerts & Sync Success</span>
              <span className="trka-cc-sub">Operational exceptions vs telemetry sync rate</span>
            </div>
            <div className="trka-cc-body">
              {trends.length === 0 ? (
                <div className="trka-empty-chart">No historical trend data in this period.</div>
              ) : (
                <ResponsiveContainer width="100%" height={240}>
                  <BarChart data={trends} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#f1f5f9" />
                    <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#64748b' }} tickFormatter={(d) => d.slice(5)} />
                    <YAxis yAxisId="left" tick={{ fontSize: 11, fill: '#64748b' }} allowDecimals={false} />
                    <YAxis yAxisId="right" orientation="right" domain={[0, 100]} tick={{ fontSize: 11, fill: '#64748b' }} unit="%" />
                    <Tooltip
                      contentStyle={{ background: '#0f172a', border: 'none', borderRadius: '6px', color: '#fff', fontSize: '0.75rem' }}
                      formatter={(val, name) => [name === 'Sync Success Rate' ? `${Number(val).toFixed(1)}%` : val, name]}
                    />
                    <Legend wrapperStyle={{ fontSize: '0.75rem', paddingTop: '10px' }} />
                    <Bar yAxisId="left" dataKey="alert_count" name="Alert Volume" fill="#ef4444" radius={[4, 4, 0, 0]} />
                    <Line
                      yAxisId="right"
                      type="monotone"
                      dataKey="refresh_success_rate"
                      name="Sync Success Rate"
                      stroke="#2563eb"
                      strokeWidth={2}
                      dot={{ r: 3 }}
                    />
                  </BarChart>
                </ResponsiveContainer>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* ── 4. Carrier Performance & Route Intelligence Dual Layout ──────── */}
      <div className="trka-dual-grid">
        {/* Left: Carrier Reliability Leaderboard */}
        <div className="trka-dual-card">
          <div className="trka-dc-header">
            <div className="trka-dch-title-box">
              <Ship size={16} className="trka-blue-icon" />
              <h3 className="trka-dc-title">Ocean & Air Carrier Performance Scorecard</h3>
            </div>
            <div className="trka-dch-filters">
              <input
                type="text"
                className="trka-input-search"
                placeholder="Filter carrier..."
                value={carrierSearch}
                onChange={(e) => setCarrierSearch(e.target.value)}
              />
              <select
                className="trka-select-sort"
                value={carrierSort}
                onChange={(e) => setCarrierSort(e.target.value)}
              >
                <option value="score">Sort: Reliability Score</option>
                <option value="ontime">Sort: On-Time %</option>
                <option value="volume">Sort: Volume Tracked</option>
              </select>
            </div>
          </div>

          <div className="trka-table-wrap">
            {filteredCarriers.length === 0 ? (
              <div className="trka-empty-state">No carrier performance data recorded.</div>
            ) : (
              <table className="trka-table">
                <thead>
                  <tr>
                    <th>Carrier</th>
                    <th>Volume</th>
                    <th>On-Time Rate</th>
                    <th>Avg Delay</th>
                    <th>Alerts</th>
                    <th>Reliability</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredCarriers.map((c) => (
                    <tr key={c.carrier_scac}>
                      <td>
                        <div className="trka-carrier-cell">
                          <span className="trka-carrier-scac">{c.carrier_scac}</span>
                          <span className="trka-carrier-name">{c.carrier_name}</span>
                        </div>
                      </td>
                      <td className="trka-num bold">{c.shipments_tracked}</td>
                      <td>
                        <div className="trka-progress-cell">
                          <div className="trka-progress-bar">
                            <div
                              className="trka-progress-fill green"
                              style={{ width: `${Math.min(c.on_time_rate, 100)}%` }}
                            />
                          </div>
                          <span className="trka-progress-pct">{c.on_time_rate.toFixed(1)}%</span>
                        </div>
                      </td>
                      <td>
                        {c.average_delay_hours > 0 ? (
                          <span className="trka-delay-pill">+{c.average_delay_hours.toFixed(1)}h</span>
                        ) : (
                          <span className="trka-ok-pill">0h</span>
                        )}
                      </td>
                      <td>
                        <span className={`trka-alert-count ${c.critical_alert_count > 0 ? 'red' : 'neutral'}`}>
                          {c.alert_count}
                          {c.critical_alert_count > 0 && ` (${c.critical_alert_count} crit)`}
                        </span>
                      </td>
                      <td>
                        <span className={`trka-tier-badge tier-${c.reliability_tier.toLowerCase()}`}>
                          {c.reliability_tier} ({c.reliability_score.toFixed(0)})
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>

        {/* Right: Route Corridor Intelligence */}
        <div className="trka-dual-card">
          <div className="trka-dc-header">
            <div className="trka-dch-title-box">
              <Navigation size={16} className="trka-purple-icon" />
              <h3 className="trka-dc-title">Route Corridor Schedule Variance</h3>
            </div>
            <span className="trka-dc-badge">{routes.length} Active Corridors</span>
          </div>

          <div className="trka-routes-list">
            {routes.length === 0 ? (
              <div className="trka-empty-state">No route corridor analytics recorded yet.</div>
            ) : (
              routes.map((r) => (
                <div key={r.route_key} className={`trka-route-item risk-${r.risk_level.toLowerCase()}`}>
                  <div className="trka-ri-top">
                    <div className="trka-ri-corridor">
                      <span className="trka-ri-port origin">{r.origin_port}</span>
                      <span className="trka-ri-arrow">➔</span>
                      <span className="trka-ri-port dest">{r.destination_port}</span>
                    </div>
                    <span className={`trka-risk-badge risk-${r.risk_level.toLowerCase()}`}>
                      {r.risk_level} RISK
                    </span>
                  </div>

                  <div className="trka-ri-metrics">
                    <div className="trka-rim-col">
                      <span className="trka-rim-lbl">VOLUME</span>
                      <span className="trka-rim-val">{r.shipments_count} Shipments</span>
                    </div>
                    <div className="trka-rim-col">
                      <span className="trka-rim-lbl">ON-TIME %</span>
                      <span className={`trka-rim-val ${r.on_time_rate >= 80 ? 'green' : 'amber'}`}>
                        {r.on_time_rate.toFixed(1)}%
                      </span>
                    </div>
                    <div className="trka-rim-col">
                      <span className="trka-rim-lbl">AVG TRANSIT</span>
                      <span className="trka-rim-val">{(r.avg_transit_hours / 24).toFixed(1)} days</span>
                    </div>
                    <div className="trka-rim-col">
                      <span className="trka-rim-lbl">AVG VARIANCE</span>
                      <span className={`trka-rim-val ${r.avg_transit_variance_hours > 24 ? 'red' : 'purple'}`}>
                        ±{(r.avg_transit_variance_hours / 24).toFixed(1)}d
                      </span>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
