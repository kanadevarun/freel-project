import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './OperationalDashboard.css';

export default function OperationalDashboard({ data }) {
  const navigate = useNavigate();
  const [chartPeriod, setChartPeriod] = useState('Last 7 days');
  const [aiQuestion, setAiQuestion] = useState('');
  const [showBanner, setShowBanner] = useState(true);

  // Extract operational stats
  const stats = data?.stats || {};
  const approvalQueue = data?.approval_queue || [];
  const openRFQs = stats.open_rfqs || 0;
  const activeShipments = stats.active_shipments || 0;
  const totalRevenue = stats.total_revenue || 0;
  const winRate = stats.win_rate ? `${stats.win_rate.toFixed(1)}%` : '97.2%';
  const draftQuotes = approvalQueue.length;

  const formatCurrency = (amount) => {
    if (!amount) return '$0';
    if (amount >= 1000000) return `$${(amount / 1000000).toFixed(2)}M`;
    if (amount >= 1000) return `$${(amount / 1000).toFixed(1)}K`;
    return `$${amount.toLocaleString()}`;
  };

  const quickActions = [
    { label: 'Create RFQ', icon: '📝', action: () => navigate('/dashboard/rfqs') },
    { label: 'Book Shipment', icon: '🚢', action: () => navigate('/dashboard/shipments') },
    { label: 'Track Shipment', icon: '📍', action: () => navigate('/dashboard/tracking') },
    { label: 'Create Invoice', icon: '💳', action: () => navigate('/dashboard/invoices') },
    { label: 'Add Customer', icon: '👤', action: () => navigate('/dashboard/companies') },
    { label: 'Upload Document', icon: '📤', action: () => navigate('/dashboard/documents') },
  ];

  const notifications = [
    { id: 1, title: 'New RFQ Received', subtitle: 'RFQ-2026-089 from Global Traders Inc.', time: '2m ago', icon: '📝', iconBg: '#EFF6FF' },
    { id: 2, title: 'Shipment Delayed', subtitle: 'SH-2026-045 is delayed at Singapore', time: '15m ago', icon: '⚠️', iconBg: '#FEF3C7' },
    { id: 3, title: 'Document Approved', subtitle: 'Commercial Invoice for SH-2026-044', time: '1h ago', icon: '📄', iconBg: '#DCFCE7' },
    { id: 4, title: 'Payment Received', subtitle: '$45,000 from Oceanic Imports', time: '2h ago', icon: '💳', iconBg: '#F3E8FF' },
  ];

  const starterPrompts = [
    'Which shipments are delayed?',
    'Show me top 5 customers by revenue',
    'What are the upcoming milestones?',
    'Suggest cost optimization for this route',
  ];

  return (
    <div className="operational-dashboard animate-fade-in-up">
      {/* ── 5 METRIC CARDS ROW ── */}
      <div className="metric-cards-row">
        {/* Total Shipments */}
        <div className="metric-card">
          <div className="metric-top">
            <div className="metric-icon-box blue">🚢</div>
            <div className="metric-title-group">
              <span className="metric-label">Total Shipments</span>
              <div className="metric-value">{activeShipments > 0 ? activeShipments.toLocaleString() : '1,247'}</div>
            </div>
          </div>
          <div className="metric-bottom">
            <span className="metric-trend positive">↑ 18.5% <span className="trend-period">vs last week</span></span>
            <svg className="sparkline" viewBox="0 0 100 24" preserveAspectRatio="none">
              <path d="M0,18 Q25,8 50,14 T100,6" fill="none" stroke="#3B82F6" strokeWidth="2" />
            </svg>
          </div>
        </div>

        {/* Active RFQs */}
        <div className="metric-card">
          <div className="metric-top">
            <div className="metric-icon-box green">📝</div>
            <div className="metric-title-group">
              <span className="metric-label">Active RFQs</span>
              <div className="metric-value">{openRFQs > 0 ? openRFQs : '89'}</div>
            </div>
          </div>
          <div className="metric-bottom">
            <span className="metric-trend positive">↑ 12.3% <span className="trend-period">vs last week</span></span>
            <svg className="sparkline" viewBox="0 0 100 24" preserveAspectRatio="none">
              <path d="M0,20 Q30,12 60,16 T100,8" fill="none" stroke="#10B981" strokeWidth="2" />
            </svg>
          </div>
        </div>

        {/* Draft Quotations */}
        <div className="metric-card">
          <div className="metric-top">
            <div className="metric-icon-box orange">📑</div>
            <div className="metric-title-group">
              <span className="metric-label">Draft Quotations</span>
              <div className="metric-value">{draftQuotes > 0 ? draftQuotes : '36'}</div>
            </div>
          </div>
          <div className="metric-bottom">
            <span className="metric-trend positive">↑ 8.7% <span className="trend-period">vs last week</span></span>
            <svg className="sparkline" viewBox="0 0 100 24" preserveAspectRatio="none">
              <path d="M0,16 Q20,20 50,10 T100,6" fill="none" stroke="#F59E0B" strokeWidth="2" />
            </svg>
          </div>
        </div>

        {/* Revenue (YTD) */}
        <div className="metric-card">
          <div className="metric-top">
            <div className="metric-icon-box purple">💰</div>
            <div className="metric-title-group">
              <span className="metric-label">Revenue (YTD)</span>
              <div className="metric-value">{totalRevenue > 0 ? formatCurrency(totalRevenue) : '$2.48M'}</div>
            </div>
          </div>
          <div className="metric-bottom">
            <span className="metric-trend positive">↑ 24.6% <span className="trend-period">vs last year</span></span>
            <svg className="sparkline" viewBox="0 0 100 24" preserveAspectRatio="none">
              <path d="M0,22 Q35,10 70,14 T100,4" fill="none" stroke="#8B5CF6" strokeWidth="2" />
            </svg>
          </div>
        </div>

        {/* On-Time Delivery */}
        <div className="metric-card">
          <div className="metric-top">
            <div className="metric-icon-box cyan">⏱️</div>
            <div className="metric-title-group">
              <span className="metric-label">On-Time Delivery</span>
              <div className="metric-value">{winRate}</div>
            </div>
          </div>
          <div className="metric-bottom">
            <span className="metric-trend positive">↑ 3.4% <span className="trend-period">vs last week</span></span>
            <svg className="sparkline" viewBox="0 0 100 24" preserveAspectRatio="none">
              <path d="M0,14 Q30,16 60,8 T100,4" fill="none" stroke="#06B6D4" strokeWidth="2" />
            </svg>
          </div>
        </div>
      </div>

      {/* ── 2-COLUMN MAIN CONTENT GRID ── */}
      <div className="operational-main-grid">
        {/* ── LEFT COLUMN (2/3) ── */}
        <div className="operational-left-col">
          {/* Charts Row */}
          <div className="charts-row">
            {/* Shipments Overview Chart */}
            <div className="op-card chart-card">
              <div className="card-header-flex">
                <div>
                  <h3 className="op-card-title">Shipments Overview</h3>
                  <div className="chart-legend-row">
                    <span className="legend-dot green"></span> Delivered
                    <span className="legend-dot blue"></span> In Transit
                    <span className="legend-dot orange"></span> Delayed
                    <span className="legend-dot red"></span> Cancelled
                  </div>
                </div>
                <div className="period-dropdown">
                  <span>{chartPeriod}</span> ⌵
                </div>
              </div>

              {/* Chart SVG Canvas */}
              <div className="chart-canvas-wrapper">
                <svg className="overview-chart-svg" viewBox="0 0 500 180" preserveAspectRatio="none">
                  {/* Grid Lines */}
                  <line x1="40" y1="30" x2="490" y2="30" stroke="#F1F5F9" strokeWidth="1" />
                  <line x1="40" y1="70" x2="490" y2="70" stroke="#F1F5F9" strokeWidth="1" />
                  <line x1="40" y1="110" x2="490" y2="110" stroke="#F1F5F9" strokeWidth="1" />
                  <line x1="40" y1="150" x2="490" y2="150" stroke="#F1F5F9" strokeWidth="1" />

                  {/* Y Axis Labels */}
                  <text x="10" y="34" fill="#94A3B8" fontSize="10">250</text>
                  <text x="10" y="74" fill="#94A3B8" fontSize="10">200</text>
                  <text x="10" y="114" fill="#94A3B8" fontSize="10">150</text>
                  <text x="10" y="154" fill="#94A3B8" fontSize="10">0</text>

                  {/* In Transit Line (Blue) */}
                  <path
                    d="M 50,110 Q 110,65 180,68 T 320,62 T 420,40 T 480,50"
                    fill="none"
                    stroke="#3B82F6"
                    strokeWidth="2.5"
                  />
                  {/* Delivered Line (Green) */}
                  <path
                    d="M 50,130 Q 110,120 180,135 T 320,105 T 420,102 T 480,110"
                    fill="none"
                    stroke="#10B981"
                    strokeWidth="2.5"
                  />
                  {/* Delayed Line (Orange) */}
                  <path
                    d="M 50,155 Q 110,150 180,148 T 320,138 T 420,142 T 480,132"
                    fill="none"
                    stroke="#F59E0B"
                    strokeWidth="2"
                  />
                  {/* Cancelled Line (Red) */}
                  <path
                    d="M 50,170 Q 110,170 180,168 T 320,166 T 420,165 T 480,168"
                    fill="none"
                    stroke="#EF4444"
                    strokeWidth="2"
                  />
                </svg>

                {/* X Axis Dates */}
                <div className="chart-x-labels">
                  <span>Aug 9</span>
                  <span>Aug 10</span>
                  <span>Aug 11</span>
                  <span>Aug 12</span>
                  <span>Aug 13</span>
                  <span>Aug 14</span>
                  <span>Aug 15</span>
                </div>
              </div>
            </div>

            {/* Shipment Status Donut */}
            <div className="op-card donut-card">
              <h3 className="op-card-title">Shipment Status</h3>
              
              <div className="donut-chart-container">
                <div className="donut-graphic">
                  <svg viewBox="0 0 100 100" className="donut-svg">
                    <circle cx="50" cy="50" r="38" stroke="#3B82F6" strokeWidth="14" fill="none" strokeDasharray="120 240" strokeDashoffset="0" />
                    <circle cx="50" cy="50" r="38" stroke="#10B981" strokeWidth="14" fill="none" strokeDasharray="88 240" strokeDashoffset="-120" />
                    <circle cx="50" cy="50" r="38" stroke="#F59E0B" strokeWidth="14" fill="none" strokeDasharray="25 240" strokeDashoffset="-208" />
                    <circle cx="50" cy="50" r="38" stroke="#EF4444" strokeWidth="14" fill="none" strokeDasharray="8 240" strokeDashoffset="-233" />
                  </svg>
                  <div className="donut-center-text">
                    <span className="donut-center-label">Total</span>
                    <span className="donut-center-value">1,247</span>
                  </div>
                </div>

                <div className="donut-legend-list">
                  <div className="donut-legend-item">
                    <span className="donut-dot blue"></span>
                    <span className="donut-text">In Transit</span>
                    <span className="donut-count">621 (49.8%)</span>
                  </div>
                  <div className="donut-legend-item">
                    <span className="donut-dot green"></span>
                    <span className="donut-text">Delivered</span>
                    <span className="donut-count">456 (36.6%)</span>
                  </div>
                  <div className="donut-legend-item">
                    <span className="donut-dot orange"></span>
                    <span className="donut-text">Delayed</span>
                    <span className="donut-count">128 (10.3%)</span>
                  </div>
                  <div className="donut-legend-item">
                    <span className="donut-dot red"></span>
                    <span className="donut-text">Cancelled</span>
                    <span className="donut-count">42 (3.3%)</span>
                  </div>
                </div>
              </div>

              <div className="donut-footer-link" onClick={() => navigate('/dashboard/shipments')}>
                View all shipments →
              </div>
            </div>
          </div>

          {/* Tables Row: Recent RFQs & Recent Shipments */}
          <div className="tables-row">
            {/* Recent RFQs */}
            <div className="op-card table-card">
              <div className="card-header-flex">
                <h3 className="op-card-title">Recent RFQs</h3>
                <span className="view-all-link" onClick={() => navigate('/dashboard/rfqs')}>View all</span>
              </div>

              <div className="table-responsive">
                <table className="op-table">
                  <thead>
                    <tr>
                      <th>RFQ ID</th>
                      <th>Customer</th>
                      <th>Origin → Destination</th>
                      <th>Cargo</th>
                      <th>Received</th>
                      <th>Status</th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td className="font-semibold text-blue-600">RFQ-2026-089</td>
                      <td>Global Traders Inc.</td>
                      <td>Shanghai → Los Angeles</td>
                      <td>Electronics</td>
                      <td className="text-slate-400">5m ago</td>
                      <td><span className="status-pill new">New</span></td>
                      <td className="text-slate-400 cursor-pointer">⋮</td>
                    </tr>
                    <tr>
                      <td className="font-semibold text-blue-600">RFQ-2026-088</td>
                      <td>Oceanic Imports</td>
                      <td>Mumbai → Dubai</td>
                      <td>Textiles</td>
                      <td className="text-slate-400">1h ago</td>
                      <td><span className="status-pill new">New</span></td>
                      <td className="text-slate-400 cursor-pointer">⋮</td>
                    </tr>
                    <tr>
                      <td className="font-semibold text-blue-600">RFQ-2026-087</td>
                      <td>Euro Exports Ltd.</td>
                      <td>Rotterdam → New York</td>
                      <td>Machinery</td>
                      <td className="text-slate-400">3h ago</td>
                      <td><span className="status-pill quoted">Quoted</span></td>
                      <td className="text-slate-400 cursor-pointer">⋮</td>
                    </tr>
                    <tr>
                      <td className="font-semibold text-blue-600">RFQ-2026-086</td>
                      <td>Pacific Solutions</td>
                      <td>Singapore → Sydney</td>
                      <td>Chemicals</td>
                      <td className="text-slate-400">5h ago</td>
                      <td><span className="status-pill in-progress">In Progress</span></td>
                      <td className="text-slate-400 cursor-pointer">⋮</td>
                    </tr>
                    <tr>
                      <td className="font-semibold text-blue-600">RFQ-2026-085</td>
                      <td>African Trading Co.</td>
                      <td>Durban → Hamburg</td>
                      <td>Automotive</td>
                      <td className="text-slate-400">1d ago</td>
                      <td><span className="status-pill quoted">Quoted</span></td>
                      <td className="text-slate-400 cursor-pointer">⋮</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            {/* Recent Shipments */}
            <div className="op-card table-card">
              <div className="card-header-flex">
                <h3 className="op-card-title">Recent Shipments</h3>
                <span className="view-all-link" onClick={() => navigate('/dashboard/shipments')}>View all</span>
              </div>

              <div className="table-responsive">
                <table className="op-table">
                  <thead>
                    <tr>
                      <th>Shipment ID</th>
                      <th>Route</th>
                      <th>Status</th>
                      <th>ETA</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td className="font-semibold">SH-2026-045</td>
                      <td>Shanghai → LA</td>
                      <td><span className="status-pill transit">In Transit</span></td>
                      <td>Aug 18</td>
                    </tr>
                    <tr>
                      <td className="font-semibold">SH-2026-044</td>
                      <td>Mumbai → Dubai</td>
                      <td><span className="status-pill transit">In Transit</span></td>
                      <td>Aug 16</td>
                    </tr>
                    <tr>
                      <td className="font-semibold">SH-2026-043</td>
                      <td>Rotterdam → NY</td>
                      <td><span className="status-pill delivered">Delivered</span></td>
                      <td>Aug 12</td>
                    </tr>
                    <tr>
                      <td className="font-semibold">SH-2026-042</td>
                      <td>Singapore → Sydney</td>
                      <td><span className="status-pill transit">In Transit</span></td>
                      <td>Aug 17</td>
                    </tr>
                    <tr>
                      <td className="font-semibold">SH-2026-041</td>
                      <td>Hamburg → Durban</td>
                      <td><span className="status-pill delayed">Delayed</span></td>
                      <td>Aug 20</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>

        {/* ── RIGHT COLUMN (1/3) ── */}
        <div className="operational-right-col">
          {/* Quick Actions Card */}
          <div className="op-card quick-actions-card">
            <h3 className="op-card-title">Quick Actions</h3>
            <div className="quick-actions-grid">
              {quickActions.map((qa, i) => (
                <div key={i} className="quick-action-item" onClick={qa.action}>
                  <div className="qa-icon">{qa.icon}</div>
                  <div className="qa-label">{qa.label}</div>
                </div>
              ))}
            </div>
          </div>

          {/* Notifications Card */}
          <div className="op-card notifications-card">
            <div className="card-header-flex">
              <h3 className="op-card-title">Notifications</h3>
              <span className="view-all-link">View all</span>
            </div>

            <div className="notifications-list">
              {notifications.map((n) => (
                <div key={n.id} className="notification-item">
                  <div className="notif-icon-box" style={{ backgroundColor: n.iconBg }}>
                    {n.icon}
                  </div>
                  <div className="notif-details">
                    <div className="notif-title-row">
                      <span className="notif-title">{n.title}</span>
                      <span className="notif-time">{n.time}</span>
                    </div>
                    <div className="notif-subtitle">{n.subtitle}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* AI Assistant Widget */}
          <div className="op-card ai-assistant-card">
            <div className="ai-widget-header">
              <div className="ai-title-badge">
                <span className="sparkle">✨</span> AI Assistant <span className="beta-chip">Beta</span>
              </div>
              <span className="ai-subtitle">Ask anything about your logistics</span>
            </div>

            <div className="ai-prompts-list">
              {starterPrompts.map((p, idx) => (
                <button key={idx} className="ai-prompt-chip" onClick={() => setAiQuestion(p)}>
                  {p}
                </button>
              ))}
            </div>

            <form onSubmit={(e) => e.preventDefault()} className="ai-input-wrapper">
              <input
                type="text"
                placeholder="Type your question..."
                value={aiQuestion}
                onChange={(e) => setAiQuestion(e.target.value)}
                className="ai-chat-input"
              />
              <button type="submit" className="ai-send-button">
                ↗
              </button>
            </form>
          </div>
        </div>
      </div>

      {/* ── BOTTOM AI OPTIMIZATION BANNER ── */}
      {showBanner && (
        <div className="optimization-banner animate-fade-in-up">
          <div className="banner-left">
            <div className="banner-icon-box">⚡</div>
            <div className="banner-text">
              <h4 className="banner-heading">Optimize Your Operations</h4>
              <p className="banner-desc">
                AI insights suggest consolidating 3 shipments on the Shanghai → Los Angeles route to save $2,450
              </p>
            </div>
          </div>
          <div className="banner-right">
            <button className="btn-view-rec" onClick={() => navigate('/dashboard/market-insights')}>
              View Recommendation
            </button>
            <button className="btn-close-banner" onClick={() => setShowBanner(false)}>
              ×
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
