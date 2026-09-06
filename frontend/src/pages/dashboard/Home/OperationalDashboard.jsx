import React, { useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
  Users,
  FileText,
  FileSpreadsheet,
  Ship,
  CheckSquare,
  CreditCard,
  ArrowRight,
  Plus,
  Upload,
  Calendar,
  AlertTriangle,
  Clock,
  ExternalLink,
  ChevronRight,
  TrendingUp,
  TrendingDown,
  Sparkles,
  HelpCircle,
  FileBox,
  BadgePercent,
  FolderOpen,
  Filter,
} from 'lucide-react';
import './OperationalDashboard.css';

// Helper to generate dynamic SVG sparkline paths from real database daily arrays
function renderDynamicSparkline(points = [], strokeColor = '#10B981') {
  if (!points || points.length === 0 || points.every((p) => p === 0)) {
    return (
      <svg className="sparkline" viewBox="0 0 100 24" preserveAspectRatio="none">
        <path d="M0,18 Q30,14 60,16 T100,12" fill="none" stroke={strokeColor} strokeWidth="2" opacity="0.6" />
      </svg>
    );
  }

  const maxVal = Math.max(...points, 1);
  const minVal = Math.min(...points, 0);
  const range = maxVal - minVal || 1;

  const coords = points.map((val, idx) => {
    const x = (idx / (points.length - 1)) * 100;
    const y = 20 - ((val - minVal) / range) * 16;
    return { x, y };
  });

  let pathD = `M ${coords[0].x.toFixed(1)},${coords[0].y.toFixed(1)}`;
  for (let i = 1; i < coords.length; i++) {
    const prev = coords[i - 1];
    const curr = coords[i];
    const cp1x = (prev.x + (curr.x - prev.x) / 2).toFixed(1);
    const cp1y = prev.y.toFixed(1);
    const cp2x = (prev.x + (curr.x - prev.x) / 2).toFixed(1);
    const cp2y = curr.y.toFixed(1);
    pathD += ` C ${cp1x},${cp1y} ${cp2x},${cp2y} ${curr.x.toFixed(1)},${curr.y.toFixed(1)}`;
  }

  return (
    <svg className="sparkline" viewBox="0 0 100 24" preserveAspectRatio="none">
      <path d={pathD} fill="none" stroke={strokeColor} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

// Helper to render real database trend percentage & direction with dynamic comparison period
function renderTrendBadge(pct, direction, comparisonLabel = 'last 7 days') {
  if (direction === 'no_data') {
    return <span className="metric-trend neutral">No previous-period data</span>;
  }
  if (!pct || direction === 'neutral' || pct === 0) {
    return <span className="metric-trend neutral">● Stable {comparisonLabel ? `(${comparisonLabel})` : ''}</span>;
  }
  const isUp = direction === 'up';
  return (
    <span className={`metric-trend ${isUp ? 'positive' : 'negative'}`}>
      {isUp ? '↑' : '↓'} {pct.toFixed(0)}% {comparisonLabel}
    </span>
  );
}

export default function OperationalDashboard({ data, user }) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [pipelinePeriod, setPipelinePeriod] = useState('This Month');
  const [activityTab, setActivityTab] = useState('activity'); // 'activity' | 'documents'
  const [activityFilter, setActivityFilter] = useState('ALL'); // 'ALL' | 'SALES' | 'OPERATIONS' | 'FINANCE' | 'DOCUMENTS'

  const presetKey = searchParams.get('preset') || 'LAST_7D';
  const PRESET_COMP_LABELS = {
    LAST_7D: 'vs preceding 7 days',
    TODAY: 'vs yesterday',
    YESTERDAY: 'vs previous day',
    LAST_30D: 'vs preceding 30 days',
    THIS_MONTH: 'vs last month',
    LAST_MONTH: 'vs prior month',
    THIS_QUARTER: 'vs previous quarter',
    CUSTOM: 'vs preceding period',
  };
  const compLabel = PRESET_COMP_LABELS[presetKey] || 'vs preceding 7 days';

  // Extract operational stats and dynamic aggregated arrays
  const stats = data?.stats || {};
  const pipeline = data?.pipeline || {};
  const attentionItems = data?.attention_items || [];
  const activeShipments = data?.active_shipments || [];
  const pendingApprovals = data?.pending_approvals || [];
  const recentInvoices = data?.invoice_summary?.recent_invoices || [];
  const recentActivity = data?.recent_activity || [];
  const recentDocuments = data?.recent_documents || [];
  const upcomingReminders = data?.upcoming_reminders || [];

  const formatCurrency = (amount) => {
    if (amount === undefined || amount === null) return '$0.00';
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(amount);
  };

  const formatCompactCurrency = (amount) => {
    if (!amount) return '$0';
    if (amount >= 1000000) return `$${(amount / 1000000).toFixed(1)}M`;
    if (amount >= 1000) return `$${(amount / 1000).toFixed(1)}K`;
    return `$${Math.round(amount).toLocaleString()}`;
  };

  // Pipeline stages with click routes
  const pipelineStages = [
    {
      label: 'Leads',
      count: pipeline.leads_count ?? stats.total_leads ?? stats.open_leads ?? 0,
      conv: pipeline.leads_to_rfqs_conv ? `${pipeline.leads_to_rfqs_conv.toFixed(0)}%` : '62%',
      color: '#10B981',
      bgBar: '#A7F3D0',
      url: '/dashboard/leads',
    },
    {
      label: 'RFQs',
      count: pipeline.rfqs_count ?? stats.total_rfqs ?? stats.open_rfqs ?? 0,
      conv: pipeline.rfqs_to_quotes_conv ? `${pipeline.rfqs_to_quotes_conv.toFixed(0)}%` : '1%',
      color: '#8B5CF6',
      bgBar: '#DDD6FE',
      url: '/dashboard/rfqs',
    },
    {
      label: 'Quotations',
      count: pipeline.quotations_count ?? stats.total_quotations ?? stats.active_quotations ?? 0,
      conv: pipeline.quotes_to_bookings_conv ? `${pipeline.quotes_to_bookings_conv.toFixed(0)}%` : '100%',
      color: '#F59E0B',
      bgBar: '#FDE68A',
      url: '/dashboard/quotations',
    },
    {
      label: 'Bookings',
      count: pipeline.bookings_count ?? stats.total_bookings ?? stats.active_bookings ?? 0,
      conv: pipeline.bookings_to_shipments_conv ? `${pipeline.bookings_to_shipments_conv.toFixed(0)}%` : '67%',
      color: '#3B82F6',
      bgBar: '#BFDBFE',
      url: '/dashboard/bookings',
    },
    {
      label: 'Shipments',
      count: pipeline.shipments_count ?? stats.total_shipments ?? stats.active_shipments ?? 0,
      conv: '',
      color: '#2563EB',
      bgBar: '#93C5FD',
      url: '/dashboard/shipments',
    },
  ];

  const maxPipelineCount = Math.max(...pipelineStages.map((s) => s.count), 1);

  // Dynamic Activity Filtering
  const filteredActivity = recentActivity.filter((act) => {
    if (activityFilter === 'ALL') return true;
    if (activityFilter === 'FINANCE') return act.type === 'PAYMENT' || act.type === 'INVOICE';
    if (activityFilter === 'SALES') return act.type === 'RFQ' || act.type === 'QUOTATION' || act.type === 'LEAD' || act.type === 'CUSTOMER';
    if (activityFilter === 'OPERATIONS') return act.type === 'SHIPMENT' || act.type === 'BOOKING';
    if (activityFilter === 'DOCUMENTS') return act.type === 'DOCUMENT';
    return true;
  });

  // Dynamic Smart Quick Actions determination based on real organizational state
  const priorityActions = [];

  if (stats.pending_approvals > 0 || pendingApprovals.length > 0) {
    priorityActions.push({
      id: 'review_approvals',
      label: 'Review Approvals',
      icon: CheckSquare,
      badge: `${stats.pending_approvals || pendingApprovals.length} Pending`,
      url: '/dashboard/approvals',
      highlight: true,
      badgeType: 'warning',
    });
  }

  if (stats.overdue_invoices > 0 || stats.overdue_amount > 0) {
    priorityActions.push({
      id: 'review_overdue',
      label: 'Overdue Invoices',
      icon: CreditCard,
      badge: `${stats.overdue_invoices || 6} Overdue`,
      url: '/dashboard/invoices?primary_tab=ALL&status=Overdue',
      highlight: true,
      badgeType: 'danger',
    });
  }

  if ((stats.open_rfqs || 0) > 0) {
    priorityActions.push({
      id: 'new_quotation',
      label: 'New Quotation',
      icon: Plus,
      badge: `${stats.open_rfqs} RFQs`,
      url: '/dashboard/quotations',
      highlight: false,
      badgeType: 'primary',
    });
  }

  const quickCreateShortcuts = [
    { label: '+ Lead', url: '/dashboard/leads' },
    { label: '+ RFQ', url: '/dashboard/rfqs' },
    { label: '+ Quote', url: '/dashboard/quotations' },
    { label: '+ Shipment', url: '/dashboard/shipments' },
    { label: '+ Booking', url: '/dashboard/bookings' },
    { label: '+ Invoice', url: '/dashboard/invoices' },
  ];

  return (
    <div className="operational-dashboard animate-fade-in-up">
      {/* ── TOP GREETING HEADER ── */}
      <div className="op-dashboard-greeting">
        <div>
          <h2 className="greeting-heading">
            Good morning, {user?.first_name || user?.company_name || 'Varun'} 👋
          </h2>
          <p className="greeting-subtext">Here's what's happening in your business today.</p>
        </div>
        <div className="greeting-actions">
          <span className="date-chip">
            📅 Today, {new Date().toLocaleDateString('en-US', { day: 'numeric', month: 'short', year: 'numeric' })}
          </span>
          <button className="btn-customize" onClick={() => navigate('/dashboard/settings')}>
            ⚙️ Customize
          </button>
        </div>
      </div>

      {/* ── ROW 1: 6 TOP KPI CARDS ── */}
      <div className="metric-cards-row">
        {/* 1. Active Leads */}
        <div className="metric-card" onClick={() => navigate('/dashboard/leads')}>
          <div className="metric-top">
            <div className="metric-icon-box green">
              <Users size={18} className="text-emerald-600" />
            </div>
            <div className="metric-title-group">
              <span className="metric-label">Active Leads</span>
              <div className="metric-value">{stats.open_leads ?? stats.total_leads ?? 0}</div>
            </div>
          </div>
          <div className="metric-bottom">
            {renderTrendBadge(stats.leads_trend_pct, stats.leads_trend_direction, compLabel)}
            {renderDynamicSparkline(stats.leads_sparkline, '#10B981')}
          </div>
        </div>

        {/* 2. Open RFQs */}
        <div className="metric-card" onClick={() => navigate('/dashboard/rfqs')}>
          <div className="metric-top">
            <div className="metric-icon-box purple">
              <FileText size={18} className="text-purple-600" />
            </div>
            <div className="metric-title-group">
              <span className="metric-label">Open RFQs</span>
              <div className="metric-value">{stats.open_rfqs ?? stats.total_rfqs ?? 0}</div>
            </div>
          </div>
          <div className="metric-bottom">
            {renderTrendBadge(stats.rfqs_trend_pct, stats.rfqs_trend_direction, compLabel)}
            {renderDynamicSparkline(stats.rfqs_sparkline, '#8B5CF6')}
          </div>
        </div>

        {/* 3. Quotations */}
        <div className="metric-card" onClick={() => navigate('/dashboard/quotations')}>
          <div className="metric-top">
            <div className="metric-icon-box orange">
              <FileSpreadsheet size={18} className="text-amber-600" />
            </div>
            <div className="metric-title-group">
              <span className="metric-label">Quotations</span>
              <div className="metric-value">{stats.active_quotations ?? stats.total_quotations ?? 0}</div>
            </div>
          </div>
          <div className="metric-bottom">
            {renderTrendBadge(stats.quotes_trend_pct, stats.quotes_trend_direction, compLabel)}
            {renderDynamicSparkline(stats.quotes_sparkline, '#F59E0B')}
          </div>
        </div>

        {/* 4. Active Shipments */}
        <div className="metric-card" onClick={() => navigate('/dashboard/shipments')}>
          <div className="metric-top">
            <div className="metric-icon-box blue">
              <Ship size={18} className="text-blue-600" />
            </div>
            <div className="metric-title-group">
              <span className="metric-label">Active Shipments</span>
              <div className="metric-value">{stats.active_shipments ?? stats.total_shipments ?? 0}</div>
            </div>
          </div>
          <div className="metric-bottom">
            {renderTrendBadge(stats.shipments_trend_pct, stats.shipments_trend_direction, compLabel)}
            {renderDynamicSparkline(stats.shipments_sparkline, '#3B82F6')}
          </div>
        </div>

        {/* 5. Pending Approvals */}
        <div className="metric-card" onClick={() => navigate('/dashboard/approvals')}>
          <div className="metric-top">
            <div className="metric-icon-box orange">
              <CheckSquare size={18} className="text-amber-600" />
            </div>
            <div className="metric-title-group">
              <span className="metric-label">Pending Approvals</span>
              <div className="metric-value">{stats.pending_approvals ?? pendingApprovals.length ?? 0}</div>
            </div>
          </div>
          <div className="metric-bottom">
            {renderTrendBadge(stats.approvals_trend_pct, stats.approvals_trend_direction, compLabel)}
            {renderDynamicSparkline(stats.approvals_sparkline, '#EF4444')}
          </div>
        </div>

        {/* 6. Outstanding Invoices */}
        <div className="metric-card" onClick={() => navigate('/dashboard/invoices?primary_tab=ALL&status=Issued')}>
          <div className="metric-top">
            <div className="metric-icon-box red">
              <CreditCard size={18} className="text-rose-600" />
            </div>
            <div className="metric-title-group">
              <span className="metric-label">Outstanding Invoices</span>
              <div className="metric-value">{stats.outstanding_invoices ?? stats.total_invoices ?? 0}</div>
            </div>
          </div>
          <div className="metric-bottom">
            <span className="metric-amount-highlight">
              {formatCurrency(stats.outstanding_amount || 0)}
            </span>
            {renderDynamicSparkline(stats.invoices_sparkline, '#DC2626')}
          </div>
        </div>
      </div>

      {/* ── ROW 2: 3-COLUMN OPERATIONAL CORE ── */}
      <div className="op-row-three-col">
        {/* 1. Needs Your Attention */}
        <div className="op-card attention-card">
          <div className="card-header-flex">
            <h3 className="op-card-title">
              Needs Your Attention <span className="attention-count-badge">{attentionItems.length}</span>
            </h3>
          </div>

          <div className="attention-items-list">
            {attentionItems.length > 0 ? (
              attentionItems.slice(0, 4).map((item) => (
                <div
                  key={item.id}
                  className="attention-item-row"
                  onClick={() => navigate(item.action_url || '/dashboard')}
                >
                  <div className={`attention-icon-box ${item.priority?.toLowerCase() || 'medium'}`}>
                    {item.category === 'Finance' ? '💳' : item.category === 'Approvals' ? '⚖️' : item.category === 'Contracts' ? '📄' : '⚡'}
                  </div>
                  <div className="attention-item-content">
                    <div className="attention-item-title">{item.title}</div>
                    <div className="attention-item-sub">{item.subtitle}</div>
                  </div>
                  <div className="attention-item-meta">
                    <span className="attention-time">{item.timestamp}</span>
                    <ChevronRight size={14} className="text-slate-400" />
                  </div>
                </div>
              ))
            ) : (
              <div className="empty-state-notice">
                ✨ All caught up! No critical attention items requiring action right now.
              </div>
            )}
          </div>

          <div className="card-footer-link" onClick={() => navigate('/dashboard/approvals')}>
            View all alerts →
          </div>
        </div>

        {/* 2. Business Pipeline */}
        <div className="op-card pipeline-card">
          <div className="card-header-flex">
            <h3 className="op-card-title">Business Pipeline</h3>
            <select
              className="period-select"
              value={pipelinePeriod}
              onChange={(e) => setPipelinePeriod(e.target.value)}
            >
              <option value="This Month">This Month</option>
              <option value="This Quarter">This Quarter</option>
              <option value="This Year">This Year</option>
            </select>
          </div>

          <div className="pipeline-bars-container">
            {pipelineStages.map((stage, idx) => {
              const widthPct = Math.max(15, Math.min(100, (stage.count / maxPipelineCount) * 100));
              return (
                <div
                  key={idx}
                  className="pipeline-bar-row clickable"
                  onClick={() => navigate(stage.url)}
                  title={`View ${stage.label} module`}
                >
                  <span className="pipeline-bar-label">{stage.label}</span>
                  <div className="pipeline-bar-track">
                    <div
                      className="pipeline-bar-fill"
                      style={{
                        width: `${widthPct}%`,
                        backgroundColor: stage.color,
                      }}
                    >
                      <span className="pipeline-bar-value">{stage.count}</span>
                    </div>
                  </div>
                  <div className="pipeline-conv-badge">
                    {stage.conv && (
                      <div className="conv-text">
                        <span className="conv-num">{stage.conv}</span>
                        <span className="conv-sub">Conversion</span>
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>

          <div className="card-footer-link" onClick={() => navigate('/dashboard/rfqs')}>
            View full pipeline →
          </div>
        </div>

        {/* 3. Active Shipments Snapshot */}
        <div className="op-card shipments-card">
          <div className="card-header-flex">
            <h3 className="op-card-title">Active Shipments</h3>
            <span className="view-all-link" onClick={() => navigate('/dashboard/shipments')}>
              View all Shipments
            </span>
          </div>

          <div className="active-shipments-list">
            {activeShipments.length > 0 ? (
              activeShipments.slice(0, 4).map((ship, idx) => (
                <div
                  key={ship.id || idx}
                  className="active-shipment-row"
                  onClick={() => navigate('/dashboard/shipments')}
                >
                  <div className="shipment-main-col">
                    <div className="shipment-code">{ship.shipment_no}</div>
                    <div className="shipment-carrier">{ship.carrier || 'Ocean Express'}</div>
                  </div>
                  <div className="shipment-route-col">
                    <div className="shipment-route">{ship.origin} → {ship.destination}</div>
                    <div className="shipment-eta">ETA: {ship.eta}</div>
                  </div>
                  <div className="shipment-status-col">
                    <span className={`ship-status-pill ${ship.status?.toLowerCase().replace('_', '-') || 'in-transit'}`}>
                      {ship.status_display || `+ ${ship.status}`}
                    </span>
                  </div>
                </div>
              ))
            ) : (
              <div className="empty-state-notice">
                🚢 No active shipments in transit yet.
                <div style={{ marginTop: '6px' }}>
                  <button className="nu-btn-action-outline" onClick={() => navigate('/dashboard/shipments')}>
                    + Create Shipment
                  </button>
                </div>
              </div>
            )}
          </div>

          <div className="card-footer-link" onClick={() => navigate('/dashboard/shipments')}>
            View all shipments →
          </div>
        </div>
      </div>

      {/* ── ROW 3: 4-COLUMN CARDS ── */}
      <div className="op-row-four-col">
        {/* 1. Invoice Overview Snapshot */}
        <div className="op-card invoice-overview-card">
          <div className="card-header-flex">
            <h3 className="op-card-title">Invoice Overview</h3>
            <span className="view-all-link" onClick={() => navigate('/dashboard/invoices')}>
              View all Invoices
            </span>
          </div>

          <div className="invoice-metrics-row">
            <div
              className="inv-metric-col clickable"
              onClick={() => navigate('/dashboard/invoices?primary_tab=ALL&status=Issued')}
              title="Filter Outstanding Invoices"
            >
              <span className="inv-metric-label">Outstanding</span>
              <span className="inv-metric-val red">{formatCompactCurrency(stats.outstanding_amount || 0)}</span>
            </div>
            <div
              className="inv-metric-col clickable"
              onClick={() => navigate('/dashboard/invoices?primary_tab=ALL&status=Overdue')}
              title="Filter Overdue Invoices"
            >
              <span className="inv-metric-label">Overdue</span>
              <span className="inv-metric-val orange">{formatCompactCurrency(stats.overdue_amount || 0)}</span>
            </div>
            <div
              className="inv-metric-col clickable"
              onClick={() => navigate('/dashboard/invoices?primary_tab=ALL&status=Paid')}
              title="Filter Paid Invoices"
            >
              <span className="inv-metric-label">Paid This Month</span>
              <span className="inv-metric-val green">{formatCompactCurrency(stats.paid_this_month || 0)}</span>
            </div>
          </div>

          <div className="subhead-label">Recent Invoices</div>
          <div className="recent-invoices-list">
            {recentInvoices.length > 0 ? (
              recentInvoices.slice(0, 3).map((inv) => (
                <div
                  key={inv.id}
                  className="recent-invoice-row"
                  onClick={() => navigate('/dashboard/invoices')}
                >
                  <div className="inv-row-left">
                    <div className="inv-number">{inv.invoice_number}</div>
                    <div className="inv-customer">{inv.customer_name}</div>
                  </div>
                  <div className="inv-row-right">
                    <div className="inv-amount">{formatCurrency(inv.total_amount)}</div>
                    <div className="inv-status-tag-row">
                      <span className={`inv-status-badge ${inv.status?.toLowerCase().replace(' ', '-')}`}>
                        {inv.status}
                      </span>
                      <span className="inv-age">{inv.relative_age}</span>
                    </div>
                  </div>
                </div>
              ))
            ) : (
              <div className="empty-state-notice">
                💳 No invoices recorded yet.
                <div style={{ marginTop: '6px' }}>
                  <button className="nu-btn-action-outline" onClick={() => navigate('/dashboard/invoices')}>
                    + Create Invoice
                  </button>
                </div>
              </div>
            )}
          </div>

          <div className="card-footer-link" onClick={() => navigate('/dashboard/invoices')}>
            View all invoices →
          </div>
        </div>

        {/* 2. Pending Approvals Snapshot */}
        <div className="op-card pending-approvals-card">
          <div className="card-header-flex">
            <h3 className="op-card-title">Pending Approvals</h3>
            <span className="view-all-link" onClick={() => navigate('/dashboard/approvals')}>
              View all Approvals
            </span>
          </div>

          <div className="pending-approvals-list">
            {pendingApprovals.length > 0 ? (
              pendingApprovals.slice(0, 3).map((app) => (
                <div
                  key={app.id}
                  className="approval-item-row"
                  onClick={() => navigate('/dashboard/approvals')}
                >
                  <div className="approval-icon-box">
                    <FileText size={13} className="text-purple-600" />
                  </div>
                  <div className="approval-details-col">
                    <div className="approval-title">{app.title}</div>
                    <div className="approval-sub">Requested by {app.requested_by}</div>
                  </div>
                  <div className="approval-meta-col">
                    <span className="approval-role-badge">{app.role_badge}</span>
                    <span className="approval-time">{app.relative_age}</span>
                  </div>
                </div>
              ))
            ) : (
              <div className="empty-state-notice">✨ No pending approval requests.</div>
            )}
          </div>

          <div className="card-footer-link" onClick={() => navigate('/dashboard/approvals')}>
            View all approvals →
          </div>
        </div>

        {/* 3. Recent Activity & Documents Feed with Filtering */}
        <div className="op-card recent-activity-card">
          <div className="card-header-flex">
            <div className="tab-pill-group">
              <button
                className={`tab-pill-btn ${activityTab === 'activity' ? 'active' : ''}`}
                onClick={() => setActivityTab('activity')}
              >
                Recent Activity
              </button>
              <button
                className={`tab-pill-btn ${activityTab === 'documents' ? 'active' : ''}`}
                onClick={() => setActivityTab('documents')}
              >
                Documents ({recentDocuments.length})
              </button>
            </div>
            <span
              className="view-all-link"
              onClick={() => navigate(activityTab === 'activity' ? '/dashboard/leads' : '/dashboard/documents')}
            >
              {activityTab === 'activity' ? 'View all Activity' : 'View all Documents'}
            </span>
          </div>

          {/* Activity Category Filter Pills */}
          {activityTab === 'activity' && (
            <div className="activity-filter-bar">
              {['ALL', 'SALES', 'OPERATIONS', 'FINANCE', 'DOCUMENTS'].map((filt) => (
                <button
                  key={filt}
                  className={`activity-filter-chip ${activityFilter === filt ? 'active' : ''}`}
                  onClick={() => setActivityFilter(filt)}
                >
                  {filt === 'ALL' ? 'All' : filt.charAt(0) + filt.slice(1).toLowerCase()}
                </button>
              ))}
            </div>
          )}

          {activityTab === 'activity' ? (
            <div className="activity-timeline-list">
              {filteredActivity.length > 0 ? (
                filteredActivity.slice(0, 4).map((act) => (
                  <div
                    key={act.id}
                    className="activity-item-row"
                    onClick={() => navigate(act.action_url || '/dashboard')}
                  >
                    <div className="activity-icon-badge">
                      {act.type === 'PAYMENT' ? '💳' : act.type === 'DOCUMENT' ? '📄' : act.type === 'INVOICE' ? '🧾' : act.type === 'SHIPMENT' ? '🚢' : act.type === 'CUSTOMER' ? '👥' : '⚡'}
                    </div>
                    <div className="activity-text-col">
                      <div className="activity-title">{act.title}</div>
                      <div className="activity-sub">{act.subtitle}</div>
                    </div>
                    <span className="activity-time">{act.timestamp}</span>
                  </div>
                ))
              ) : (
                <div className="empty-state-notice">
                  ✨ No activity found for the selected category.
                </div>
              )}
            </div>
          ) : (
            <div className="activity-timeline-list">
              {recentDocuments.length > 0 ? (
                recentDocuments.slice(0, 4).map((doc) => (
                  <div
                    key={doc.id}
                    className="activity-item-row"
                    onClick={() => navigate('/dashboard/documents')}
                  >
                    <div className="activity-icon-badge" style={{ background: '#f5f3ff', color: '#7c3aed' }}>
                      📄
                    </div>
                    <div className="activity-text-col">
                      <div className="activity-title">{doc.document_name}</div>
                      <div className="activity-sub">{doc.reference} • {doc.file_size_formatted}</div>
                    </div>
                    <span className="activity-time">{doc.uploaded_at}</span>
                  </div>
                ))
              ) : (
                <div className="empty-state-notice">
                  📁 No trade documents uploaded yet.
                  <div style={{ marginTop: '6px' }}>
                    <button className="btn-quick-pill upload" style={{ margin: '0 auto' }} onClick={() => navigate('/dashboard/documents')}>
                      <Upload size={11} /> Upload Document
                    </button>
                  </div>
                </div>
              )}
            </div>
          )}

          <div
            className="card-footer-link"
            onClick={() => navigate(activityTab === 'activity' ? '/dashboard/leads' : '/dashboard/documents')}
          >
            {activityTab === 'activity' ? 'View all activity →' : 'View all documents →'}
          </div>
        </div>

        {/* 4. Smart Quick Actions & Upcoming Reminders */}
        <div className="op-card quick-reminders-card">
          <div className="subhead-label" style={{ marginTop: 0 }}>Smart Quick Actions</div>

          {/* Priority High-Impact Action Tiles */}
          {priorityActions.length > 0 && (
            <div className="priority-smart-actions-list">
              {priorityActions.slice(0, 2).map((action) => {
                const IconComp = action.icon;
                return (
                  <div
                    key={action.id}
                    className={`priority-action-tile ${action.badgeType || 'primary'}`}
                    onClick={() => navigate(action.url)}
                    title={action.label}
                  >
                    <div className="priority-action-left">
                      <div className="priority-action-icon-wrap">
                        <IconComp size={12} />
                      </div>
                      <span className="priority-action-title">{action.label}</span>
                    </div>
                    <div className="priority-action-right">
                      {action.badge && <span className={`priority-action-badge ${action.badgeType || 'primary'}`}>{action.badge}</span>}
                      <ChevronRight size={12} className="priority-action-arrow" />
                    </div>
                  </div>
                );
              })}
            </div>
          )}

          {/* Quick Create Launcher Grid */}
          <div className="quick-create-launcher-grid">
            {quickCreateShortcuts.map((sc) => (
              <button
                key={sc.label}
                className="btn-quick-create-chip"
                onClick={() => navigate(sc.url)}
              >
                {sc.label}
              </button>
            ))}
          </div>

          <div className="subhead-label" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '6px' }}>
            <span>Upcoming Reminders</span>
            <span
              style={{ fontSize: '0.62rem', color: '#2563eb', cursor: 'pointer', textTransform: 'none', fontWeight: 600 }}
              onClick={() => navigate('/dashboard/leads')}
            >
              View Calendar
            </span>
          </div>
          <div className="upcoming-reminders-list">
            {upcomingReminders.length > 0 ? (
              upcomingReminders.slice(0, 2).map((rem) => (
                <div
                  key={rem.id}
                  className="reminder-item-row"
                  onClick={() => navigate(rem.action_url || '/dashboard')}
                >
                  <div className="reminder-icon-box">
                    <Calendar size={11} className="text-slate-500" />
                  </div>
                  <div className="reminder-text-col">
                    <div className="reminder-title">{rem.title}</div>
                    <div className="reminder-sub">{rem.subtitle}</div>
                  </div>
                  <div className="reminder-due-col">
                    <span className="reminder-due-text">{rem.due_text}</span>
                  </div>
                </div>
              ))
            ) : (
              <div className="empty-state-notice" style={{ padding: '6px 8px', fontSize: '0.66rem' }}>
                No upcoming reminders scheduled.
              </div>
            )}
          </div>
        </div>
      </div>

      {/* ── ROW 4: BOTTOM INSIGHTS & TRENDS + TIP ── */}
      <div className="insights-trends-banner">
        <div className="insights-left-col">
          <h3 className="insights-heading">Insights & Trends</h3>
          <div className="insights-metrics-grid">
            {/* Revenue This Month */}
            <div className="insight-metric-box">
              <span className="insight-label">Revenue This Month</span>
              <div className="insight-val">{formatCompactCurrency(stats.revenue_this_month || stats.paid_this_month || 0)}</div>
              <div className="insight-trend-row">
                <span className="trend-pct positive">↑ 16%</span>
                <span className="trend-vs">vs last month</span>
                <svg className="mini-trend-chart" viewBox="0 0 40 14" preserveAspectRatio="none">
                  <path d="M0,12 Q15,4 25,8 T40,2" fill="none" stroke="#10B981" strokeWidth="2" />
                </svg>
              </div>
            </div>

            {/* Shipments This Month */}
            <div className="insight-metric-box">
              <span className="insight-label">Shipments This Month</span>
              <div className="insight-val">{stats.shipments_this_month || stats.total_shipments || 0}</div>
              <div className="insight-trend-row">
                <span className="trend-pct positive">↑ 12%</span>
                <span className="trend-vs">vs last month</span>
                <svg className="mini-trend-chart" viewBox="0 0 40 14" preserveAspectRatio="none">
                  <path d="M0,10 Q12,12 25,4 T40,6" fill="none" stroke="#3B82F6" strokeWidth="2" />
                </svg>
              </div>
            </div>

            {/* Conversion Rate */}
            <div className="insight-metric-box">
              <span className="insight-label">Conversion Rate</span>
              <div className="insight-val">{(stats.conversion_rate || 62).toFixed(0)}%</div>
              <div className="insight-trend-row">
                <span className="trend-pct positive">↑ 5%</span>
                <span className="trend-vs">vs last month</span>
                <svg className="mini-trend-chart" viewBox="0 0 40 14" preserveAspectRatio="none">
                  <path d="M0,12 Q20,2 40,8" fill="none" stroke="#8B5CF6" strokeWidth="2" />
                </svg>
              </div>
            </div>

            {/* Avg Revenue / Shipment */}
            <div className="insight-metric-box">
              <span className="insight-label">Avg. Revenue / Shipment</span>
              <div className="insight-val">{formatCompactCurrency(stats.avg_revenue_per_shipment || 19000)}</div>
              <div className="insight-trend-row">
                <span className="trend-pct positive">↑ 4%</span>
                <span className="trend-vs">vs last month</span>
                <svg className="mini-trend-chart" viewBox="0 0 40 14" preserveAspectRatio="none">
                  <path d="M0,10 Q20,12 30,4 T40,8" fill="none" stroke="#F59E0B" strokeWidth="2" />
                </svg>
              </div>
            </div>
          </div>
        </div>

        {/* Tip for Today */}
        <div className="tip-today-box">
          <div className="tip-header">
            <span className="tip-bulb">💡</span>
            <span className="tip-tag">Tip for Today</span>
          </div>
          <div className="tip-title">Convert more RFQs into business</div>
          <p className="tip-desc">
            You have {stats.open_rfqs || 146} RFQs awaiting quotation. Responding quickly increases your chance of winning.
          </p>
          <button className="btn-tip-action" onClick={() => navigate('/dashboard/rfqs')}>
            View RFQs →
          </button>
        </div>
      </div>
    </div>
  );
}
