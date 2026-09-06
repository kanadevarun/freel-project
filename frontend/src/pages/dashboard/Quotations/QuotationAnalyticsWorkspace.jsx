import React, { useState, useEffect, useCallback } from 'react';
import {
  TrendingUp,
  DollarSign,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Briefcase,
  Layers,
  Users,
  Percent,
  Calendar,
  ArrowUpRight,
  ShieldCheck,
  AlertCircle,
  RefreshCw,
  FileText,
  Sliders,
  ChevronRight,
  Info,
} from 'lucide-react';
import {
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import quotationService from '../../../services/quotationService';
import './QuotationAnalyticsWorkspace.css';

export default function QuotationAnalyticsWorkspace({ onSelectQuotation }) {
  const [timeframe, setTimeframe] = useState(30); // 7, 30, 90
  const [overview, setOverview] = useState(null);
  const [trends, setTrends] = useState([]);
  const [customers, setCustomers] = useState([]);
  const [modes, setModes] = useState([]);
  const [risks, setRisks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchAnalytics = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [overviewRes, trendsRes, custRes, modesRes, riskRes] = await Promise.all([
        quotationService.getQuotationAnalyticsOverview().catch(e => { console.warn('Overview err:', e); return null; }),
        quotationService.getQuotationAnalyticsTrends({ days: timeframe }).catch(e => { console.warn('Trends err:', e); return []; }),
        quotationService.getCustomerQuotationPerformance().catch(e => { console.warn('Customers err:', e); return []; }),
        quotationService.getQuotationPerformanceByMode().catch(e => { console.warn('Modes err:', e); return []; }),
        quotationService.getQuotationExpiryRisk().catch(e => { console.warn('Risk err:', e); return []; }),
      ]);

      if (overviewRes) setOverview(overviewRes?.data || overviewRes);
      setTrends(Array.isArray(trendsRes?.data || trendsRes) ? (trendsRes?.data || trendsRes) : []);
      setCustomers(Array.isArray(custRes?.data || custRes) ? (custRes?.data || custRes) : []);
      setModes(Array.isArray(modesRes?.data || modesRes) ? (modesRes?.data || modesRes) : []);
      setRisks(Array.isArray(riskRes?.data || riskRes) ? (riskRes?.data || riskRes) : []);
    } catch (err) {
      setError(err?.response?.data?.error?.message || err?.message || 'Failed to load analytics dashboard.');
    } finally {
      setLoading(false);
    }
  }, [timeframe]);

  useEffect(() => {
    fetchAnalytics();
  }, [fetchAnalytics]);

  const currency = overview?.currency || 'USD';

  const fmtCurrency = (val) => {
    if (val === undefined || val === null) return '—';
    return `${currency} ${Number(val).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  };

  const fmtPercent = (val) => {
    if (val === undefined || val === null) return '0.0%';
    return `${Number(val).toFixed(1)}%`;
  };

  return (
    <div className="qt-analytics-workspace">
      {/* ── Top Bar ── */}
      <div className="qt-analytics-topbar">
        <div>
          <h2 className="qt-analytics-title">Commercial Performance & Quotation Analytics</h2>
          <p className="qt-analytics-sub">
            Real-time management intelligence on sales proposals, conversion velocity, customer performance, and margin health.
          </p>
        </div>
        <div className="qt-analytics-controls">
          <div className="qt-timeframe-picker">
            <span className="qt-tf-label"><Calendar size={13} /> Timeframe:</span>
            {[
              { days: 7, label: '7 Days' },
              { days: 30, label: '30 Days' },
              { days: 90, label: '90 Days' },
            ].map(tf => (
              <button
                key={tf.days}
                className={`qt-tf-btn ${timeframe === tf.days ? 'active' : ''}`}
                onClick={() => setTimeframe(tf.days)}
              >
                {tf.label}
              </button>
            ))}
          </div>
          <button className="qt-btn qt-btn--ghost qt-btn--sm" onClick={fetchAnalytics} title="Refresh Analytics">
            <RefreshCw size={13} className={loading ? 'qt-spin' : ''} />
          </button>
        </div>
      </div>

      {error && (
        <div className="qt-error-banner">
          <AlertCircle size={14} /> {error}
        </div>
      )}

      {/* ── Executive KPI Strip ── */}
      <div className="qt-analytics-kpi-grid">
        <div className="qt-kpi-metric-card">
          <div className="qt-kpi-m-top">
            <div className="qt-kpi-m-icon blue"><FileText size={18} /></div>
            <span className="qt-kpi-m-tag">Total Pipeline</span>
          </div>
          <div className="qt-kpi-m-val">{overview?.total_quotes ?? 0}</div>
          <div className="qt-kpi-m-label">Total Quotes Created</div>
          <div className="qt-kpi-m-footer">
            <span>{fmtCurrency(overview?.pipeline_value)} active value</span>
          </div>
        </div>

        <div className="qt-kpi-metric-card">
          <div className="qt-kpi-m-top">
            <div className="qt-kpi-m-icon green"><CheckCircle2 size={18} /></div>
            <span className="qt-kpi-m-tag green">Win Rate</span>
          </div>
          <div className="qt-kpi-m-val green">{fmtPercent(overview?.acceptance_rate)}</div>
          <div className="qt-kpi-m-label">Customer Acceptance Rate</div>
          <div className="qt-kpi-m-footer">
            <span>{overview?.accepted_quotes ?? 0} accepted quotes</span>
          </div>
        </div>

        <div className="qt-kpi-metric-card">
          <div className="qt-kpi-m-top">
            <div className="qt-kpi-m-icon purple"><DollarSign size={18} /></div>
            <span className="qt-kpi-m-tag purple">Accepted</span>
          </div>
          <div className="qt-kpi-m-val">{fmtCurrency(overview?.accepted_value)}</div>
          <div className="qt-kpi-m-label">Accepted Commercial Revenue</div>
          <div className="qt-kpi-m-footer">
            <span>Avg {fmtCurrency(overview?.average_quote_value)} / quote</span>
          </div>
        </div>

        <div className="qt-kpi-metric-card">
          <div className="qt-kpi-m-top">
            <div className="qt-kpi-m-icon indigo"><Briefcase size={18} /></div>
            <span className="qt-kpi-m-tag indigo">Handover</span>
          </div>
          <div className="qt-kpi-m-val">{fmtPercent(overview?.quote_to_booking_conversion_rate)}</div>
          <div className="qt-kpi-m-label">Quote → Booking Conversion</div>
          <div className="qt-kpi-m-footer">
            <span>{fmtCurrency(overview?.converted_booking_value)} in bookings</span>
          </div>
        </div>

        <div className="qt-kpi-metric-card">
          <div className="qt-kpi-m-top">
            <div className="qt-kpi-m-icon emerald"><Percent size={18} /></div>
            <span className="qt-kpi-m-tag emerald">Profitability</span>
          </div>
          <div className="qt-kpi-m-val">{fmtPercent(overview?.average_gross_margin_pct)}</div>
          <div className="qt-kpi-m-label">Average Gross Margin</div>
          <div className="qt-kpi-m-footer">
            <span>{fmtCurrency(overview?.total_gross_profit)} gross profit</span>
          </div>
        </div>

        <div className="qt-kpi-metric-card">
          <div className="qt-kpi-m-top">
            <div className="qt-kpi-m-icon amber"><Clock size={18} /></div>
            <span className="qt-kpi-m-tag amber">At Risk</span>
          </div>
          <div className="qt-kpi-m-val amber">{overview?.expiring_soon_count ?? 0}</div>
          <div className="qt-kpi-m-label">Expiring Soon (&lt; 7 Days)</div>
          <div className="qt-kpi-m-footer">
            <span>{overview?.stuck_in_review_count ?? 0} pending review</span>
          </div>
        </div>
      </div>

      {/* ── Operational Intelligence & Deterministic Insights ── */}
      {overview?.insights?.length > 0 && (
        <div className="qt-insights-container">
          <div className="qt-section-title-row">
            <div className="qt-sec-title">
              <ShieldCheck size={16} color="#2563EB" /> Deterministic Operational Intelligence
            </div>
            <span className="qt-sec-badge">{overview.insights.length} Actionable Recommendations</span>
          </div>
          <div className="qt-insights-grid">
            {overview.insights.map((ins, idx) => (
              <div key={idx} className={`qt-insight-card ${ins.severity?.toLowerCase()}`}>
                <div className="qt-insight-top">
                  <span className={`qt-insight-badge ${ins.severity?.toLowerCase()}`}>{ins.category}</span>
                  <span className="qt-insight-metric">{ins.metric_value}</span>
                </div>
                <div className="qt-insight-headline">{ins.headline}</div>
                <div className="qt-insight-desc">{ins.description}</div>
                <div className="qt-insight-action">
                  <strong>Recommended Action:</strong> {ins.recommended_action}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ── Charts & Trend Visualizations ── */}
      <div className="qt-charts-grid">
        {/* Chart 1: Quote Volume Trends */}
        <div className="qt-chart-card">
          <div className="qt-chart-header">
            <div>
              <div className="qt-chart-title">Quotation Volume & Lifecycle Velocity</div>
              <div className="qt-chart-sub">Daily progression of quotes created, sent, and accepted</div>
            </div>
          </div>
          <div className="qt-chart-body">
            {trends.length === 0 ? (
              <div className="qt-chart-empty">
                <Info size={24} />
                <p>No quotation activity recorded in the last {timeframe} days.</p>
                <span>Trends will populate as quotes progress through the lifecycle.</span>
              </div>
            ) : (
              <ResponsiveContainer width="100%" height={260}>
                <AreaChart data={trends} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                  <defs>
                    <linearGradient id="colorCreated" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#3B82F6" stopOpacity={0.4} />
                      <stop offset="95%" stopColor="#3B82F6" stopOpacity={0.0} />
                    </linearGradient>
                    <linearGradient id="colorAccepted" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#10B981" stopOpacity={0.4} />
                      <stop offset="95%" stopColor="#10B981" stopOpacity={0.0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#F1F5F9" />
                  <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#64748B' }} />
                  <YAxis tick={{ fontSize: 11, fill: '#64748B' }} />
                  <Tooltip
                    contentStyle={{ background: '#0F172A', border: 'none', borderRadius: '8px', color: '#FFF', fontSize: '12px' }}
                  />
                  <Legend iconType="circle" wrapperStyle={{ fontSize: '12px', paddingTop: '8px' }} />
                  <Area type="monotone" dataKey="quotes_created" name="Created" stroke="#3B82F6" strokeWidth={2} fillOpacity={1} fill="url(#colorCreated)" />
                  <Area type="monotone" dataKey="quotes_accepted" name="Accepted" stroke="#10B981" strokeWidth={2} fillOpacity={1} fill="url(#colorAccepted)" />
                </AreaChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>

        {/* Chart 2: Commercial Value Trajectory */}
        <div className="qt-chart-card">
          <div className="qt-chart-header">
            <div>
              <div className="qt-chart-title">Commercial Revenue & Pipeline Trajectory</div>
              <div className="qt-chart-sub">Total pipeline value vs accepted quote revenue</div>
            </div>
          </div>
          <div className="qt-chart-body">
            {trends.length === 0 ? (
              <div className="qt-chart-empty">
                <Info size={24} />
                <p>No revenue data available for the selected timeframe.</p>
              </div>
            ) : (
              <ResponsiveContainer width="100%" height={260}>
                <BarChart data={trends} margin={{ top: 10, right: 10, left: 10, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#F1F5F9" />
                  <XAxis dataKey="date" tick={{ fontSize: 11, fill: '#64748B' }} />
                  <YAxis tick={{ fontSize: 11, fill: '#64748B' }} tickFormatter={val => `$${val > 1000 ? (val/1000).toFixed(0)+'k' : val}`} />
                  <Tooltip
                    formatter={(val) => fmtCurrency(val)}
                    contentStyle={{ background: '#0F172A', border: 'none', borderRadius: '8px', color: '#FFF', fontSize: '12px' }}
                  />
                  <Legend iconType="circle" wrapperStyle={{ fontSize: '12px', paddingTop: '8px' }} />
                  <Bar dataKey="pipeline_value" name="Pipeline Value" fill="#6366F1" radius={[4, 4, 0, 0]} />
                  <Bar dataKey="accepted_value" name="Accepted Value" fill="#10B981" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>
        </div>
      </div>

      {/* ── Two Column: Customer Leaderboard & Transport Mode Breakdown ── */}
      <div className="qt-analytics-split-grid">
        {/* Customer Performance */}
        <div className="qt-analytics-card">
          <div className="qt-card-hdr">
            <div className="qt-card-title"><Users size={15} /> Top Customer Quotation Performance</div>
            <span className="qt-badge-sub">{customers.length} Accounts</span>
          </div>
          {customers.length === 0 ? (
            <div className="qt-empty-card-body">
              <p>No customer quotation history available yet.</p>
            </div>
          ) : (
            <div className="qt-table-container">
              <table className="qt-perf-table">
                <thead>
                  <tr>
                    <th>CUSTOMER</th>
                    <th style={{ textAlign: 'center' }}>QUOTES</th>
                    <th style={{ textAlign: 'center' }}>WIN RATE</th>
                    <th style={{ textAlign: 'right' }}>ACCEPTED VALUE</th>
                    <th style={{ textAlign: 'right' }}>AVG QUOTE</th>
                  </tr>
                </thead>
                <tbody>
                  {customers.map((c, idx) => (
                    <tr key={c.customer_id || idx}>
                      <td><strong>{c.customer_name}</strong></td>
                      <td style={{ textAlign: 'center' }}>{c.quote_count}</td>
                      <td style={{ textAlign: 'center' }}>
                        <span className={`qt-rate-pill ${c.acceptance_rate >= 50 ? 'good' : 'neutral'}`}>
                          {fmtPercent(c.acceptance_rate)}
                        </span>
                      </td>
                      <td style={{ textAlign: 'right', fontWeight: 600, color: '#0F172A' }}>
                        {fmtCurrency(c.accepted_value)}
                      </td>
                      <td style={{ textAlign: 'right', color: '#64748B' }}>
                        {fmtCurrency(c.average_quote_value)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Transport Mode Performance */}
        <div className="qt-analytics-card">
          <div className="qt-card-hdr">
            <div className="qt-card-title"><Layers size={15} /> Performance by Transport Mode</div>
            <span className="qt-badge-sub">{modes.length} Categories</span>
          </div>
          {modes.length === 0 ? (
            <div className="qt-empty-card-body">
              <p>No mode performance data available yet.</p>
            </div>
          ) : (
            <div className="qt-table-container">
              <table className="qt-perf-table">
                <thead>
                  <tr>
                    <th>MODE & SERVICE</th>
                    <th style={{ textAlign: 'center' }}>VOLUME</th>
                    <th style={{ textAlign: 'center' }}>WIN RATE</th>
                    <th style={{ textAlign: 'right' }}>AVG MARGIN</th>
                    <th style={{ textAlign: 'right' }}>PIPELINE</th>
                  </tr>
                </thead>
                <tbody>
                  {modes.map((m, idx) => (
                    <tr key={idx}>
                      <td>
                        <strong>{m.transport_mode}</strong>
                        <span className="qt-sub-tag">{m.service_type}</span>
                      </td>
                      <td style={{ textAlign: 'center' }}>{m.quote_count}</td>
                      <td style={{ textAlign: 'center' }}>{fmtPercent(m.acceptance_rate)}</td>
                      <td style={{ textAlign: 'right', color: '#16A34A', fontWeight: 600 }}>
                        {fmtPercent(m.average_margin_pct)}
                      </td>
                      <td style={{ textAlign: 'right', fontWeight: 600, color: '#0F172A' }}>
                        {fmtCurrency(m.pipeline_value)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* ── Expiry Risk & Follow-Up Priority Queue ── */}
      <div className="qt-analytics-card">
        <div className="qt-card-hdr">
          <div className="qt-card-title"><AlertTriangle size={15} color="#D97706" /> Expiry Risk & Follow-Up Priority Queue</div>
          <span className="qt-badge-sub">{risks.length} Pending Attention</span>
        </div>
        {risks.length === 0 ? (
          <div className="qt-empty-card-body">
            <CheckCircle2 size={24} color="#16A34A" />
            <p>No quotations currently at risk of expiring or stalled in review.</p>
          </div>
        ) : (
          <div className="qt-table-container">
            <table className="qt-perf-table">
              <thead>
                <tr>
                  <th>QUOTATION</th>
                  <th>CUSTOMER</th>
                  <th>AMOUNT</th>
                  <th>STATUS</th>
                  <th>VALIDITY / DAYS</th>
                  <th>RECOMMENDED ACTION</th>
                  <th style={{ textAlign: 'right' }}>ACTION</th>
                </tr>
              </thead>
              <tbody>
                {risks.map((r) => (
                  <tr key={r.quotation_id}>
                    <td><strong>{r.quotation_number}</strong></td>
                    <td>{r.customer_name}</td>
                    <td style={{ fontWeight: 600 }}>{fmtCurrency(r.total_amount)}</td>
                    <td><span className="qt-status-tag-sm">{r.status}</span></td>
                    <td>
                      <span className={`qt-risk-tag ${r.risk_category?.toLowerCase()}`}>
                        {r.days_until_expiry <= 0 ? 'Expired' : `${r.days_until_expiry} days left`}
                      </span>
                    </td>
                    <td className="qt-rec-action">{r.recommended_action}</td>
                    <td style={{ textAlign: 'right' }}>
                      {onSelectQuotation && (
                        <button
                          className="qt-btn qt-btn--secondary qt-btn--sm"
                          onClick={() => onSelectQuotation(r.quotation_id)}
                        >
                          View Quote <ChevronRight size={12} />
                        </button>
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
  );
}
