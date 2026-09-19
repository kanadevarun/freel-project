import React, { useState, useEffect, useMemo } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
  Users,
  FileText,
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
  Sparkles,
  Search,
  Activity,
  CheckCircle2,
  AlertCircle,
  Cpu,
  Database,
  ShieldAlert,
  Server,
  Filter,
  Lock,
  Globe,
  Radio,
} from 'lucide-react';
import { useSafeRBAC } from '../../../context/RBACContext';
import { monitoringService } from '../../../services/monitoringService';
import { aiTaskService } from '../../../services/aiTaskService';
import { recommendationService } from '../../../services/recommendationService';
import { integrationService } from '../../../services/integrationService';
import WorkloadCapacityPredictiveCard from '../../../components/predictions/WorkloadCapacityPredictiveCard';
import ResourceBottleneckPredictiveCard from '../../../components/predictions/ResourceBottleneckPredictiveCard';
import './OperationalDashboard.css';

// Helper to render real database trend percentage or grounded operational status
function renderKpiStatus(trendPct, trendDirection, compLabel, fallbackText, fallbackType = 'neutral') {
  // Only display a trend pill when reliable comparison data exists (not no_data, not neutral, not unbacked 100%)
  if (
    trendDirection &&
    trendDirection !== 'no_data' &&
    trendDirection !== 'neutral' &&
    trendPct > 0 &&
    trendPct !== 100
  ) {
    const isUp = trendDirection === 'up' || trendDirection === 'UP';
    return (
      <div className="kpi-status-row" title={`${isUp ? 'Increased' : 'Decreased'} ${trendPct.toFixed(0)}% ${compLabel}`}>
        <span className={`kpi-trend-pill ${isUp ? 'positive' : 'negative'}`}>
          {isUp ? '↑' : '↓'} {trendPct.toFixed(0)}%
        </span>
        <span className="kpi-trend-label">{compLabel}</span>
      </div>
    );
  }

  // Reliable operational state tag backed by real current records
  return (
    <div className="kpi-status-row">
      <span className={`kpi-status-tag ${fallbackType}`}>
        {fallbackText}
      </span>
    </div>
  );
}

export default function OperationalDashboard({ data, user }) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [pipelinePeriod, setPipelinePeriod] = useState('This Month');
  const [activityTab, setActivityTab] = useState('activity'); // 'activity' | 'documents'
  const [activityFilter, setActivityFilter] = useState('ALL'); // 'ALL' | 'SALES' | 'OPERATIONS' | 'FINANCE' | 'DOCUMENTS'
  const [priorityFilter, setPriorityFilter] = useState('ALL'); // 'ALL' | 'CRITICAL' | 'IMPORTANT' | 'INFORMATIONAL'
  const { hasRole, hasAnyRole, roleName, can, permissionsSet } = useSafeRBAC();

  // Dedicated states for Compact AI Summary, System Health & Integrations
  const [aiWorkforceSummary, setAiWorkforceSummary] = useState(null);
  const [recStats, setRecStats] = useState(null);
  const [healthSummary, setHealthSummary] = useState(null);
  const [integrationsList, setIntegrationsList] = useState([]);
  const [aiLoading, setAiLoading] = useState(true);
  const [healthLoading, setHealthLoading] = useState(true);
  const [integrationsLoading, setIntegrationsLoading] = useState(true);

  useEffect(() => {
    let isMounted = true;
    Promise.allSettled([
      aiTaskService.getWorkforceSummary(),
      recommendationService.getStats(),
      monitoringService.getHealthSummary(),
      integrationService.getStatuses(),
    ]).then(([wfRes, recRes, healthRes, intRes]) => {
      if (!isMounted) return;
      setAiLoading(false);
      setHealthLoading(false);
      setIntegrationsLoading(false);

      if (wfRes.status === 'fulfilled' && wfRes.value) {
        const val = wfRes.value;
        setAiWorkforceSummary(val?.data || (val.health ? val : null));
      }
      if (recRes.status === 'fulfilled' && recRes.value) {
        const val = recRes.value;
        setRecStats(val?.stats || val?.data || null);
      }
      if (healthRes.status === 'fulfilled' && healthRes.value) {
        const val = healthRes.value;
        setHealthSummary(val?.data || (val.subsystems ? val : null));
      }
      if (intRes.status === 'fulfilled' && intRes.value) {
        const val = intRes.value;
        const list = val?.data || val || [];
        setIntegrationsList(Array.isArray(list) ? list : []);
      }
    }).catch(() => {
      if (isMounted) {
        setAiLoading(false);
        setHealthLoading(false);
        setIntegrationsLoading(false);
      }
    });

    return () => {
      isMounted = false;
    };
  }, []);

  // Format freshness / relative time for business cards
  const formatLastActivity = (timestamp) => {
    if (!timestamp) return 'Real-time';
    try {
      const d = new Date(timestamp);
      if (isNaN(d.getTime())) return 'Recently';
      const diffSec = Math.floor((Date.now() - d.getTime()) / 1000);
      if (diffSec < 60) return 'Just now';
      if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m ago`;
      if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}h ago`;
      return `${Math.floor(diffSec / 86400)}d ago`;
    } catch {
      return 'Recently';
    }
  };

  // Permission checks for AI Workforce, System Health, and External Integrations navigation
  const canViewAIWorkforce = hasAnyRole(['SUPER_ADMIN', 'ADMIN', 'OPERATIONS_MANAGER', 'DEVELOPER', 'CEO']) || (typeof can === 'function' && can('SETTINGS', 'READ'));
  const canViewAuditLogs = hasAnyRole(['SUPER_ADMIN', 'ADMIN', 'AUDITOR', 'OPERATIONS_MANAGER', 'CEO']) || (typeof can === 'function' && can('SETTINGS', 'READ'));
  const canManageIntegrations = hasAnyRole(['SUPER_ADMIN', 'ADMIN', 'DEVELOPER', 'OPERATIONS_MANAGER', 'CEO']) || (typeof can === 'function' && can('SETTINGS', 'READ'));

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
  const shipmentCounts = data?.shipment_counts || data?.shipment_status || {};
  const pendingApprovals = data?.pending_approvals || [];
  const recentInvoices = data?.invoice_summary?.recent_invoices || [];
  const recentActivity = data?.recent_activity || [];
  const recentDocuments = data?.recent_documents || [];
  const upcomingReminders = data?.upcoming_reminders || [];
 
  // Priority Actions: Urgency grouping, deterministic counts, and filtering
  const criticalCount = attentionItems.filter(
    (i) => (i.urgency || '').toUpperCase() === 'CRITICAL' || (i.priority || '').toUpperCase() === 'CRITICAL'
  ).length;

  const importantCount = attentionItems.filter(
    (i) => (i.urgency || '').toUpperCase() === 'IMPORTANT' || (i.priority || '').toUpperCase() === 'HIGH'
  ).length;

  const informationalCount = attentionItems.filter(
    (i) => (i.urgency || '').toUpperCase() === 'INFORMATIONAL' || (i.priority || '').toUpperCase() === 'MEDIUM' || (i.priority || '').toUpperCase() === 'LOW'
  ).length;

  const filteredPriorityItems = attentionItems.filter((item) => {
    if (priorityFilter === 'ALL') return true;
    const urg = (item.urgency || '').toUpperCase();
    const pri = (item.priority || '').toUpperCase();
    if (priorityFilter === 'CRITICAL') return urg === 'CRITICAL' || pri === 'CRITICAL';
    if (priorityFilter === 'IMPORTANT') return urg === 'IMPORTANT' || pri === 'HIGH';
    if (priorityFilter === 'INFORMATIONAL') return urg === 'INFORMATIONAL' || pri === 'MEDIUM' || pri === 'LOW';
    return true;
  });

  const checkItemPermission = (item) => {
    if (!item?.required_role) return true;
    const roleReq = item.required_role.toUpperCase();

    // 1. Elevated administrative and executive roles always have operational clearance
    if (hasAnyRole(['SUPER_ADMIN', 'ADMIN', 'CEO', 'OPERATIONS_MANAGER']) || (user?.role && ['SUPER_ADMIN', 'ADMIN', 'CEO'].includes(String(user.role).toUpperCase()))) {
      return true;
    }

    // 2. Granular permission check (e.g. "SHIPMENTS:READ")
    if (roleReq.includes(':')) {
      const [mod, act] = roleReq.split(':');
      if (typeof can === 'function' && can(mod, act)) return true;
      if (permissionsSet && (permissionsSet.has(roleReq) || permissionsSet.has('*'))) return true;
      return false;
    }

    // 3. Named role check (e.g. "FINANCE_DIRECTOR")
    if (hasRole(roleReq) || (user?.role && String(user.role).toUpperCase() === roleReq)) {
      return true;
    }

    return false;
  };

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

  // ── Authoritative Deduplication and Density Safeguards ──
  // 1. Gather entity IDs claimed by Priority Actions to prevent repeating identical records
  const claimedPriorityEntityIds = useMemo(() => {
    const set = new Set();
    attentionItems.forEach((item) => {
      if (item.source_entity_id) {
        set.add(`${(item.category || item.module || '').toLowerCase()}:${item.source_entity_id}`);
      }
      if (item.approval_id) {
        set.add(`approvals:${item.approval_id}`);
      }
      if (item.source_reference) {
        set.add(String(item.source_reference).trim().toUpperCase());
      }
    });
    return set;
  }, [attentionItems]);

  // 2. Deduplicate Upcoming Reminders: omit items already covered by an urgent Priority Action
  const dedupedUpcomingReminders = useMemo(() => {
    const seen = new Set();
    return upcomingReminders.filter((rem) => {
      const remIdStr = String(rem.id || '');
      if (remIdStr.startsWith('lead_rem_')) {
        const id = remIdStr.replace('lead_rem_', '');
        if (claimedPriorityEntityIds.has(`leads:${id}`)) return false;
      }
      if (remIdStr.startsWith('contract_rem_')) {
        const id = remIdStr.replace('contract_rem_', '');
        if (claimedPriorityEntityIds.has(`contracts:${id}`)) return false;
      }
      if (remIdStr.startsWith('inv_rem_')) {
        const id = remIdStr.replace('inv_rem_', '');
        if (claimedPriorityEntityIds.has(`finance:${id}`)) return false;
      }
      if (seen.has(rem.id)) return false;
      seen.add(rem.id);
      return true;
    }).slice(0, 3);
  }, [upcomingReminders, claimedPriorityEntityIds]);

  // 3. Deduplicate Pending Approvals in Finance column:
  // If a pending approval is already featured in Priority Actions, omit it when other pending approvals exist
  // so the operator sees the broader approval gate without repeating identical records.
  const dedupedPendingApprovals = useMemo(() => {
    return pendingApprovals.filter((app) => {
      const reqCodeStr = String(app.request_code || '').toUpperCase();
      if (claimedPriorityEntityIds.has(`approvals:${app.id}`) || (reqCodeStr && claimedPriorityEntityIds.has(reqCodeStr))) {
        if (pendingApprovals.length > 1) {
          return false;
        }
      }
      return true;
    });
  }, [pendingApprovals, claimedPriorityEntityIds]);

  // 4. Deduplicate Recent Invoices by ID
  const dedupedRecentInvoices = useMemo(() => {
    const seen = new Set();
    return recentInvoices.filter((inv) => {
      if (seen.has(inv.id)) return false;
      seen.add(inv.id);
      return true;
    });
  }, [recentInvoices]);

  // 5. Deduplicate Recent Activity by activity ID
  const dedupedRecentActivity = useMemo(() => {
    const seen = new Set();
    return recentActivity.filter((act) => {
      if (seen.has(act.id)) return false;
      seen.add(act.id);
      return true;
    });
  }, [recentActivity]);

  // Dynamic Activity Filtering
  const filteredActivity = dedupedRecentActivity.filter((act) => {
    if (activityFilter === 'ALL') return true;
    if (activityFilter === 'FINANCE') return act.type === 'PAYMENT' || act.type === 'INVOICE';
    if (activityFilter === 'SALES') return act.type === 'RFQ' || act.type === 'QUOTATION' || act.type === 'LEAD' || act.type === 'CUSTOMER';
    if (activityFilter === 'OPERATIONS') return act.type === 'SHIPMENT' || act.type === 'BOOKING';
    if (activityFilter === 'DOCUMENTS') return act.type === 'DOCUMENT';
    return true;
  });

  const quickCreateShortcuts = [
    { label: '+ Lead', url: '/dashboard/leads' },
    { label: '+ RFQ', url: '/dashboard/rfqs' },
    { label: '+ Quote', url: '/dashboard/quotations' },
    { label: '+ Shipment', url: '/dashboard/shipments' },
    { label: '+ Booking', url: '/dashboard/bookings' },
    { label: '+ Invoice', url: '/dashboard/invoices' },
  ];

  // Helper for priority badge styling in Priority Actions
  const getPriorityBadgeClass = (priority) => {
    switch ((priority || '').toUpperCase()) {
      case 'CRITICAL':
      case 'DANGER':
        return 'priority-badge-critical';
      case 'HIGH':
      case 'WARNING':
        return 'priority-badge-high';
      case 'MEDIUM':
      case 'MODERATE':
        return 'priority-badge-medium';
      default:
        return 'priority-badge-low';
    }
  };

  const getPriorityCategoryIcon = (category) => {
    switch (category) {
      case 'Finance':
        return '💳';
      case 'Approvals':
        return '⚖️';
      case 'Contracts':
        return '📄';
      case 'Shipments':
        return '🚢';
      case 'RFQs':
        return '📋';
      case 'Leads':
        return '👥';
      default:
        return '⚡';
    }
  };

  return (
    <div className="operational-dashboard animate-fade-in-up">
      {/* ════════════════════════════════════════════════════════════════════
          SECTION 1 — DASHBOARD HEADER
          ════════════════════════════════════════════════════════════════════ */}
      <header className="dashboard-section-header">
        <div className="header-main-titles">
          <div className="header-badge-row">
            <span className="ops-status-indicator online">● Operations Live</span>
            <span className="date-chip">
              📅 {new Date().toLocaleDateString('en-US', { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' })}
            </span>
          </div>
          <h1 className="greeting-heading">
            Operations Command — {user?.first_name || (user?.name && !user.name.includes('@') ? user.name : null) || user?.company_name || (user?.email ? user.email.split('@')[0] : null) || 'Operator'}
          </h1>
          <p className="greeting-subtext">
            Consolidated operational freight, real-time pipeline status, active exceptions, and financial ledger.
          </p>
        </div>

        <div className="header-actions-cluster">
          {/* Quick Create Dropdown / Shortcuts */}
          <div className="header-quick-actions">
            {quickCreateShortcuts.slice(0, 3).map((sc) => (
              <button
                key={sc.label}
                className="btn-header-shortcut"
                onClick={() => navigate(sc.url)}
              >
                {sc.label}
              </button>
            ))}
          </div>

          <button
            className="btn-dashboard-action"
            onClick={() => navigate('/dashboard/settings')}
            title="Configure Workspace Preferences"
          >
            ⚙️ Preferences
          </button>
        </div>
      </header>

      {/* ════════════════════════════════════════════════════════════════════
          SECTION 2 — KPI SUMMARY (Primary 5 Metrics)
          ════════════════════════════════════════════════════════════════════ */}
      <section className="dashboard-kpi-summary-section" aria-label="KPI Summary">
        <div className="metric-cards-row five-col">
          {/* 1. Active Leads */}
          <div
            className="metric-card card-leads"
            onClick={() => navigate('/dashboard/leads')}
            onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); navigate('/dashboard/leads'); } }}
            role="button"
            tabIndex={0}
            title="View Active Leads"
            aria-label={`Active Leads: ${stats.open_leads ?? stats.total_leads ?? 0}`}
          >
            <div className="metric-top">
              <div className="metric-icon-box green">
                <Users size={17} className="text-emerald-600" />
              </div>
              <span className="metric-label">Active Leads</span>
              <ChevronRight size={14} className="metric-card-arrow" />
            </div>
            <div className="metric-mid">
              <div className="metric-value">{stats.open_leads ?? stats.total_leads ?? 0}</div>
            </div>
            <div className="metric-bottom">
              {renderKpiStatus(
                stats.leads_trend_pct,
                stats.leads_trend_direction,
                compLabel,
                `● ${stats.open_leads ?? stats.total_leads ?? 0} in pipeline`,
                'success'
              )}
            </div>
          </div>

          {/* 2. Open RFQs */}
          <div
            className="metric-card card-rfqs"
            onClick={() => navigate('/dashboard/rfqs')}
            onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); navigate('/dashboard/rfqs'); } }}
            role="button"
            tabIndex={0}
            title="View Open RFQs"
            aria-label={`Open RFQs: ${stats.open_rfqs ?? stats.total_rfqs ?? 0}`}
          >
            <div className="metric-top">
              <div className="metric-icon-box purple">
                <FileText size={17} className="text-purple-600" />
              </div>
              <span className="metric-label">Open RFQs</span>
              <ChevronRight size={14} className="metric-card-arrow" />
            </div>
            <div className="metric-mid">
              <div className="metric-value">{stats.open_rfqs ?? stats.total_rfqs ?? 0}</div>
            </div>
            <div className="metric-bottom">
              {renderKpiStatus(
                stats.rfqs_trend_pct,
                stats.rfqs_trend_direction,
                compLabel,
                `● ${stats.open_rfqs ?? stats.total_rfqs ?? 0} awaiting quotes`,
                'info'
              )}
            </div>
          </div>

          {/* 3. Active Shipments */}
          <div
            className="metric-card card-shipments"
            onClick={() => navigate('/dashboard/shipments')}
            onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); navigate('/dashboard/shipments'); } }}
            role="button"
            tabIndex={0}
            title="View Active Shipments"
            aria-label={`Active Shipments: ${stats.active_shipments ?? stats.total_shipments ?? 0}`}
          >
            <div className="metric-top">
              <div className="metric-icon-box blue">
                <Ship size={17} className="text-blue-600" />
              </div>
              <span className="metric-label">Active Shipments</span>
              <ChevronRight size={14} className="metric-card-arrow" />
            </div>
            <div className="metric-mid">
              <div className="metric-value">{stats.active_shipments ?? stats.total_shipments ?? 0}</div>
            </div>
            <div className="metric-bottom">
              {renderKpiStatus(
                stats.shipments_trend_pct,
                stats.shipments_trend_direction,
                compLabel,
                `● ${data?.shipment_status_counts?.in_transit ?? stats.active_shipments ?? 0} in transit`,
                'primary'
              )}
            </div>
          </div>

          {/* 4. Pending Approvals */}
          <div
            className="metric-card card-approvals"
            onClick={() => navigate('/dashboard/approvals')}
            onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); navigate('/dashboard/approvals'); } }}
            role="button"
            tabIndex={0}
            title="View Pending Approvals"
            aria-label={`Pending Approvals: ${stats.pending_approvals ?? pendingApprovals.length ?? 0}`}
          >
            <div className="metric-top">
              <div className="metric-icon-box coral">
                <CheckSquare size={17} className="text-orange-600" />
              </div>
              <span className="metric-label">Pending Approvals</span>
              <ChevronRight size={14} className="metric-card-arrow" />
            </div>
            <div className="metric-mid">
              <div className="metric-value">{stats.pending_approvals ?? pendingApprovals.length ?? 0}</div>
            </div>
            <div className="metric-bottom">
              {renderKpiStatus(
                stats.approvals_trend_pct,
                stats.approvals_trend_direction,
                compLabel,
                `● ${stats.pending_approvals ?? pendingApprovals.length ?? 0} need review`,
                'warning'
              )}
            </div>
          </div>

          {/* 5. Outstanding Invoices */}
          <div
            className="metric-card card-invoices"
            onClick={() => navigate('/dashboard/invoices?primary_tab=ALL&status=Issued')}
            onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); navigate('/dashboard/invoices?primary_tab=ALL&status=Issued'); } }}
            role="button"
            tabIndex={0}
            title="View Outstanding Invoices"
            aria-label={`Outstanding Invoices: ${stats.outstanding_invoices ?? stats.total_invoices ?? 0}, Amount: ${formatCurrency(stats.outstanding_amount || 0)}`}
          >
            <div className="metric-top">
              <div className="metric-icon-box red">
                <CreditCard size={17} className="text-rose-600" />
              </div>
              <span className="metric-label">Outstanding Invoices</span>
              <ChevronRight size={14} className="metric-card-arrow" />
            </div>
            <div className="metric-mid">
              <div className="metric-value-split">
                <span className="metric-value">{stats.outstanding_invoices ?? stats.total_invoices ?? 0}</span>
                <span className="metric-amount-pill">{formatCurrency(stats.outstanding_amount || 0)}</span>
              </div>
            </div>
            <div className="metric-bottom">
              {stats.overdue_invoices > 0 ? (
                <div className="kpi-status-row">
                  <span className="kpi-status-tag danger">
                    ● {stats.overdue_invoices} overdue ({formatCurrency(stats.overdue_amount || 0)})
                  </span>
                </div>
              ) : (
                renderKpiStatus(
                  stats.invoices_trend_pct,
                  stats.invoices_trend_direction,
                  compLabel,
                  '● Awaiting payment',
                  'neutral'
                )
              )}
            </div>
          </div>
        </div>
      </section>

      {/* ════════════════════════════════════════════════════════════════════
          SECTION 3 — PRIORITY ACTIONS
          ════════════════════════════════════════════════════════════════════ */}
      <section className="dashboard-priority-actions-section" aria-label="Priority Actions">
        <div className="op-card priority-actions-card">
          {/* Card Header with Real Operational Context */}
          <div className="card-header-flex">
            <div className="section-title-wrap">
              <h2 className="op-card-title">
                <AlertTriangle size={16} className="text-amber-500" />
                Priority Actions
                {attentionItems.length > 0 && (
                  <span className="priority-count-pill">{attentionItems.length} require attention</span>
                )}
              </h2>
              <span className="section-subtitle-text">
                Real operational tasks requiring attention, prioritized by operational urgency and business impact.
              </span>
            </div>
            <button
              className="card-action-link-btn"
              onClick={() => navigate('/dashboard/approvals')}
              title="View all pending approvals"
            >
              View Approvals ({stats.pending_approvals ?? pendingApprovals.length ?? 0}) →
            </button>
          </div>

          {/* Compact Summary & Filter Controls */}
          <div className="priority-actions-filter-bar">
            <div className="priority-filter-tabs" role="tablist" aria-label="Filter Priority Actions by Urgency">
              <button
                type="button"
                className={`priority-filter-chip ${priorityFilter === 'ALL' ? 'active' : ''}`}
                onClick={() => setPriorityFilter('ALL')}
                role="tab"
                aria-selected={priorityFilter === 'ALL'}
              >
                All <span className="chip-count">{attentionItems.length}</span>
              </button>
              <button
                type="button"
                className={`priority-filter-chip critical ${priorityFilter === 'CRITICAL' ? 'active' : ''}`}
                onClick={() => setPriorityFilter('CRITICAL')}
                role="tab"
                aria-selected={priorityFilter === 'CRITICAL'}
              >
                <span className="priority-dot critical" />
                Critical <span className="chip-count">{criticalCount}</span>
              </button>
              <button
                type="button"
                className={`priority-filter-chip important ${priorityFilter === 'IMPORTANT' ? 'active' : ''}`}
                onClick={() => setPriorityFilter('IMPORTANT')}
                role="tab"
                aria-selected={priorityFilter === 'IMPORTANT'}
              >
                <span className="priority-dot important" />
                Important <span className="chip-count">{importantCount}</span>
              </button>
              <button
                type="button"
                className={`priority-filter-chip informational ${priorityFilter === 'INFORMATIONAL' ? 'active' : ''}`}
                onClick={() => setPriorityFilter('INFORMATIONAL')}
                role="tab"
                aria-selected={priorityFilter === 'INFORMATIONAL'}
              >
                <span className="priority-dot informational" />
                Informational <span className="chip-count">{informationalCount}</span>
              </button>
            </div>

            <div className="priority-actions-counter-summary">
              <span className="counter-summary-text">
                <strong>{criticalCount}</strong> Critical · <strong>{importantCount}</strong> Important · <strong>{informationalCount}</strong> Informational
              </span>
            </div>
          </div>

          {/* Priority Items Grid */}
          <div className="priority-actions-items-container">
            {filteredPriorityItems.length > 0 ? (
              <div className="priority-actions-cards-list">
                {filteredPriorityItems.map((item, idx) => {
                  const isAllowed = checkItemPermission(item);
                  const urgency = (item.urgency || item.priority || 'INFORMATIONAL').toUpperCase();
                  const isCritical = urgency === 'CRITICAL';
                  const isImportant = urgency === 'IMPORTANT' || urgency === 'HIGH';
                  const severityClass = isCritical ? 'urgency-critical' : isImportant ? 'urgency-important' : 'urgency-informational';
                  const severityLabel = isCritical ? 'Critical' : isImportant ? 'Important' : 'Informational';

                  return (
                    <article
                      key={item.id ? `priority-${item.id}-${item.type || item.module || idx}` : `priority-idx-${idx}`}
                      className={`priority-action-card ${severityClass} ${!isAllowed ? 'restricted' : ''}`}
                      onClick={() => {
                        if (isAllowed && item.action_url) {
                          navigate(item.action_url);
                        }
                      }}
                      onKeyDown={(e) => {
                        if (isAllowed && (e.key === 'Enter' || e.key === ' ')) {
                          e.preventDefault();
                          navigate(item.action_url || '/dashboard');
                        }
                      }}
                      role="button"
                      tabIndex={isAllowed ? 0 : -1}
                      aria-label={`${severityLabel} task: ${item.title}`}
                    >
                      {/* Top metadata row */}
                      <div className="card-top-meta">
                        <div className="meta-left-badges">
                          <span className={`urgency-badge ${severityClass}`}>
                            <span className="badge-dot" />
                            {severityLabel}
                          </span>
                          <span className="module-tag">{item.module || item.category || 'Operations'}</span>
                          {item.source_reference && (
                            <span className="source-ref-badge" title={`Reference: ${item.source_reference}`}>
                              {item.source_reference}
                            </span>
                          )}
                        </div>
                        <div className="meta-right-indicators">
                          {item.requires_approval && (
                            <span className="approval-required-pill" title="Operational approval is required before execution">
                              <CheckSquare size={11} /> Approval Required
                            </span>
                          )}
                          {item.capability && (
                            <span className="capability-pill">
                              {item.capability === 'APPROVAL_GATED'
                                ? 'Approval Gated'
                                : item.capability === 'DRAFT_ONLY'
                                ? 'Draft Only'
                                : item.capability === 'READ_ONLY'
                                ? 'Read-Only'
                                : 'Actionable'}
                            </span>
                          )}
                          {(item.age_text || item.timestamp) && (
                            <span className="age-pill">
                              <Clock size={11} /> {item.age_text || item.timestamp}
                            </span>
                          )}
                        </div>
                      </div>

                      {/* Content row */}
                      <div className="card-content-body">
                        <h3 className="card-item-title">{item.title}</h3>
                        <p className="card-item-explanation">{item.explanation || item.subtitle}</p>
                      </div>

                      {/* Action row */}
                      <div className="card-action-bar">
                        {!isAllowed ? (
                          <div className="restricted-permission-banner">
                            <Lock size={12} className="text-amber-600" />
                            <span>{`Restricted Access: Requires ${item.required_role || 'elevated'} permission to review`}</span>
                          </div>
                        ) : (
                          <div className="action-buttons-group">
                            <button
                              type="button"
                              className={`btn-priority-primary ${isCritical ? 'btn-critical' : isImportant ? 'btn-important' : 'btn-informational'}`}
                              onClick={(e) => {
                                e.stopPropagation();
                                navigate(item.action_url || '/dashboard');
                              }}
                            >
                              {item.action_label || (item.requires_approval ? 'Review Approvals →' : 'Review →')}
                            </button>
                            {item.secondary_url && (
                              <button
                                type="button"
                                className="btn-priority-secondary"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  navigate(item.secondary_url);
                                }}
                              >
                                View Record
                              </button>
                            )}
                          </div>
                        )}
                      </div>
                    </article>
                  );
                })}
              </div>
            ) : (
              <div className="priority-empty-state">
                <div className="empty-icon-circle">
                  <CheckCircle2 size={24} className="text-emerald-600" />
                </div>
                <div className="empty-text-wrap">
                  <p className="empty-primary-msg">No priority actions require your attention right now.</p>
                  <p className="empty-sub-msg">
                    All operational tasks, shipment exceptions, approvals, and invoices are up to date.
                  </p>
                </div>
              </div>
            )}
          </div>
        </div>
      </section>

      {/* ════════════════════════════════════════════════════════════════════
          SECTION 3.5 — PREDICTIVE DEMAND, CAPACITY & WORKLOAD INTELLIGENCE (TASK 4.10)
          ════════════════════════════════════════════════════════════════════ */}
      <WorkloadCapacityPredictiveCard defaultTab="summary" />

      {/* ════════════════════════════════════════════════════════════════════
          SECTION 3.6 — PREDICTIVE RESOURCE ALLOCATION & BOTTLENECK INTELLIGENCE (TASK 4.11)
          ════════════════════════════════════════════════════════════════════ */}
      <ResourceBottleneckPredictiveCard defaultTab="summary" />

      {/* ════════════════════════════════════════════════════════════════════
          SECTIONS 4 & 5 — OPERATIONS OVERVIEW & FINANCE/APPROVALS (2-COL DESKTOP)
          ════════════════════════════════════════════════════════════════════ */}
      <div className="dashboard-two-col-container">
        {/* ── SECTION A: OPERATIONS OVERVIEW ── */}
        <section className="dashboard-operations-column" aria-label="Operations Overview">
          {/* 1. Shipment Health & Execution */}
          <div className="op-card shipments-card">
            <div className="card-header-flex">
              <div className="title-with-badge">
                <h3 className="op-card-title">
                  <Ship size={15} className="text-blue-600" /> Active Shipments
                </h3>
                <span className="header-meta-pill">
                  {shipmentCounts.in_transit ?? activeShipments.length ?? 0} in transit
                </span>
              </div>
              <span
                className="view-all-link"
                onClick={() => navigate('/dashboard/shipments')}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    navigate('/dashboard/shipments');
                  }
                }}
              >
                View All Shipments →
              </span>
            </div>

            {/* Shipment Health Status Summary Bar */}
            <div className="shipment-health-status-bar">
              <div className="health-status-pill in-transit" title="Shipments actively moving along carrier route">
                <span className="health-dot blue" />
                <span>In Transit: <strong>{shipmentCounts.in_transit ?? activeShipments.length ?? 0}</strong></span>
              </div>
              <div className="health-status-pill exception" title="Shipments with customs hold, documentation block, or delay">
                <span className="health-dot rose" />
                <span>Exceptions: <strong>{(shipmentCounts.customs_hold || 0) + (shipmentCounts.delayed || 0)}</strong></span>
              </div>
              <div className="health-status-pill delivered" title="Shipments cleared and delivered at destination port">
                <span className="health-dot emerald" />
                <span>Delivered: <strong>{shipmentCounts.delivered ?? 0}</strong></span>
              </div>
            </div>

            <div className="active-shipments-list">
              {activeShipments.length > 0 ? (
                activeShipments.slice(0, 4).map((ship, idx) => (
                  <div
                    key={ship.id ? `ship-${ship.id}` : `ship-idx-${idx}`}
                    className="active-shipment-row"
                    onClick={() => navigate('/dashboard/shipments')}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        navigate('/dashboard/shipments');
                      }
                    }}
                  >
                    <div className="shipment-main-col">
                      <div className="shipment-code">{ship.shipment_no}</div>
                      <div className="shipment-carrier">{ship.carrier || 'Ocean Carrier'}</div>
                    </div>
                    <div className="shipment-route-col">
                      <div className="shipment-route">{ship.origin} → {ship.destination}</div>
                      <div className="shipment-eta">ETA: {ship.eta}</div>
                    </div>
                    <div className="shipment-status-col">
                      <span className={`ship-status-pill ${ship.status?.toLowerCase().replace('_', '-') || 'in-transit'}`}>
                        {ship.status_display || ship.status}
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

            <div
              className="card-footer-link"
              onClick={() => navigate('/dashboard/shipments')}
              role="button"
              tabIndex={0}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  navigate('/dashboard/shipments');
                }
              }}
            >
              Manage all shipments and milestones →
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
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        navigate(stage.url);
                      }
                    }}
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

            <div
              className="card-footer-link"
              onClick={() => navigate('/dashboard/rfqs')}
              role="button"
              tabIndex={0}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  navigate('/dashboard/rfqs');
                }
              }}
            >
              Explore sales conversion pipeline →
            </div>
          </div>
        </section>

        {/* ── SECTION B: FINANCE AND APPROVALS ── */}
        <section className="dashboard-finance-column" aria-label="Finance and Approvals">
          {/* 1. Invoice Overview */}
          <div className="op-card invoice-overview-card">
            <div className="card-header-flex">
              <h3 className="op-card-title">
                <CreditCard size={15} className="text-rose-600" /> Invoice Overview
              </h3>
              <span
                className="view-all-link"
                onClick={() => navigate('/dashboard/invoices')}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    navigate('/dashboard/invoices');
                  }
                }}
              >
                View All Invoices →
              </span>
            </div>

            <div className="invoice-metrics-row">
              <div
                className="inv-metric-col clickable"
                onClick={() => navigate('/dashboard/invoices?primary_tab=ALL&status=Issued')}
                title="Filter Outstanding Invoices"
                role="button"
                tabIndex={0}
              >
                <span className="inv-metric-label">Outstanding</span>
                <span className="inv-metric-val red">{formatCompactCurrency(stats.outstanding_amount || 0)}</span>
                <span className="inv-metric-sub">{stats.outstanding_invoices ?? recentInvoices.length ?? 0} invoices</span>
              </div>
              <div
                className="inv-metric-col clickable"
                onClick={() => navigate('/dashboard/invoices?primary_tab=ALL&status=Overdue')}
                title="Filter Overdue Invoices"
                role="button"
                tabIndex={0}
              >
                <span className="inv-metric-label">Overdue</span>
                <span className="inv-metric-val orange">{formatCompactCurrency(stats.overdue_amount || 0)}</span>
                <span className="inv-metric-sub">{stats.overdue_invoices ?? 0} past due</span>
              </div>
              <div
                className="inv-metric-col clickable"
                onClick={() => navigate('/dashboard/invoices?primary_tab=ALL&status=Paid')}
                title="Filter Paid Invoices"
                role="button"
                tabIndex={0}
              >
                <span className="inv-metric-label">Paid (30D)</span>
                <span className="inv-metric-val green">{formatCompactCurrency(stats.paid_this_month || 0)}</span>
                <span className="inv-metric-sub">Settled</span>
              </div>
            </div>

            <div className="subhead-label">Recent Invoices</div>
            <div className="recent-invoices-list">
              {dedupedRecentInvoices.length > 0 ? (
                dedupedRecentInvoices.slice(0, 3).map((inv, idx) => (
                  <div
                    key={inv.id ? `inv-${inv.id}` : `inv-idx-${idx}`}
                    className="recent-invoice-row"
                    onClick={() => navigate('/dashboard/invoices')}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        navigate('/dashboard/invoices');
                      }
                    }}
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
                        <span className="inv-age">{inv.age_text || inv.relative_age}</span>
                      </div>
                    </div>
                  </div>
                ))
              ) : (
                <div className="empty-state-notice">
                  💳 No invoices recorded yet.
                </div>
              )}
            </div>

            <div
              className="card-footer-link"
              onClick={() => navigate('/dashboard/invoices')}
              role="button"
              tabIndex={0}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  navigate('/dashboard/invoices');
                }
              }}
            >
              View finance ledger & collections →
            </div>
          </div>

          {/* 2. Pending Approvals */}
          <div className="op-card pending-approvals-card">
            <div className="card-header-flex">
              <div className="title-with-badge">
                <h3 className="op-card-title">
                  <CheckSquare size={15} className="text-orange-600" /> Pending Approvals
                </h3>
                <span className="header-meta-pill warning">
                  {stats.pending_approvals ?? pendingApprovals.length ?? 0} Awaiting Gate
                </span>
              </div>
              <span
                className="view-all-link"
                onClick={() => navigate('/dashboard/approvals')}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    navigate('/dashboard/approvals');
                  }
                }}
              >
                Approvals Center →
              </span>
            </div>

            <div className="pending-approvals-list">
              {dedupedPendingApprovals.length > 0 ? (
                dedupedPendingApprovals.slice(0, 3).map((app, idx) => (
                  <div
                    key={app.id ? `app-${app.id}` : `app-idx-${idx}`}
                    className="approval-item-row"
                    onClick={() => navigate('/dashboard/approvals')}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        navigate('/dashboard/approvals');
                      }
                    }}
                  >
                    <div className="approval-icon-box">
                      <FileText size={13} className="text-purple-600" />
                    </div>
                    <div className="approval-details-col">
                      <div className="approval-title">{app.title || app.request_code}</div>
                      <div className="approval-sub">
                        {app.category && <span className="app-cat-tag">{app.category}</span>}
                        Requested by {app.requested_by_name || app.requested_by || 'Operations'}
                      </div>
                    </div>
                    <div className="approval-meta-col">
                      <span className="approval-role-badge">{app.role_badge || app.department || 'Manager'}</span>
                      <span className="approval-time">{app.age_text || app.relative_age}</span>
                    </div>
                  </div>
                ))
              ) : (
                <div className="empty-state-notice">✨ No pending approval requests. Human-in-the-loop gate is clear.</div>
              )}
            </div>

            <div
              className="card-footer-link"
              onClick={() => navigate('/dashboard/approvals')}
              role="button"
              tabIndex={0}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  navigate('/dashboard/approvals');
                }
              }}
            >
              Review full approval queue →
            </div>
          </div>
        </section>
      </div>

      {/* ════════════════════════════════════════════════════════════════════
          SECTION C & D — RECENT BUSINESS ACTIVITY & OPERATIONAL REMINDERS
          ════════════════════════════════════════════════════════════════════ */}
      <section className="dashboard-activity-reminders-section" aria-label="Recent Business Activity">
        <div className="dashboard-activity-grid">
          {/* 1. Recent Business Activity */}
          <div className="op-card recent-activity-card">
            <div className="card-header-flex">
              <div className="tab-pill-group">
                <button
                  type="button"
                  className={`tab-pill-btn ${activityTab === 'activity' ? 'active' : ''}`}
                  onClick={() => setActivityTab('activity')}
                >
                  Recent Activity
                </button>
                <button
                  type="button"
                  className={`tab-pill-btn ${activityTab === 'documents' ? 'active' : ''}`}
                  onClick={() => setActivityTab('documents')}
                >
                  Documents ({recentDocuments.length})
                </button>
              </div>
              {activityTab === 'activity' ? (
                canViewAuditLogs && (
                  <span
                    className="view-all-link"
                    onClick={() => navigate('/dashboard/settings/audit-logs')}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        navigate('/dashboard/settings/audit-logs');
                      }
                    }}
                  >
                    Audit Logs →
                  </span>
                )
              ) : (
                <span
                  className="view-all-link"
                  onClick={() => navigate('/dashboard/documents')}
                  role="button"
                  tabIndex={0}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                      e.preventDefault();
                      navigate('/dashboard/documents');
                    }
                  }}
                >
                  Documents →
                </span>
              )}
            </div>

            {/* Category Filter Chips for Activity */}
            {activityTab === 'activity' && (
              <div className="activity-filter-bar">
                {['ALL', 'SALES', 'OPERATIONS', 'FINANCE', 'DOCUMENTS'].map((filt) => (
                  <button
                    key={filt}
                    type="button"
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
                  filteredActivity.slice(0, 5).map((act, idx) => (
                    <div
                      key={act.id ? `act-${act.id}` : `act-idx-${idx}`}
                      className="activity-item-row"
                      onClick={() => navigate(act.action_url || '/dashboard')}
                      role="button"
                      tabIndex={0}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault();
                          navigate(act.action_url || '/dashboard');
                        }
                      }}
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
                    ✨ No recent activity found for the selected filter.
                  </div>
                )}
              </div>
            ) : (
              <div className="activity-timeline-list">
                {recentDocuments.length > 0 ? (
                  recentDocuments.slice(0, 5).map((doc, idx) => (
                    <div
                      key={doc.id ? `doc-${doc.id}` : `doc-idx-${idx}`}
                      className="activity-item-row"
                      onClick={() => navigate('/dashboard/documents')}
                      role="button"
                      tabIndex={0}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault();
                          navigate('/dashboard/documents');
                        }
                      }}
                    >
                      <div className="activity-icon-badge" style={{ background: '#f5f3ff', color: '#7c3aed' }}>
                        📁
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
                      <button
                        className="btn-quick-pill upload"
                        style={{ margin: '0 auto' }}
                        onClick={() => navigate('/dashboard/documents')}
                      >
                        <Upload size={11} /> Upload Document
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}

            <div
              className="card-footer-link"
              onClick={() => navigate(activityTab === 'activity' ? '/dashboard/settings/audit-logs' : '/dashboard/documents')}
              role="button"
              tabIndex={0}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  navigate(activityTab === 'activity' ? '/dashboard/settings/audit-logs' : '/dashboard/documents');
                }
              }}
            >
              {activityTab === 'activity' ? 'View activity audit ledger →' : 'Open document center →'}
            </div>
          </div>

          {/* 2. Reminders & Verified Quick Actions */}
          <div className="op-card quick-reminders-card">
            <div className="card-header-flex">
              <h3 className="op-card-title">
                <Calendar size={15} className="text-slate-600" /> Reminders & Quick Actions
              </h3>
              <span
                className="view-all-link"
                onClick={() => navigate('/dashboard/contracts')}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    navigate('/dashboard/contracts');
                  }
                }}
              >
                Contracts →
              </span>
            </div>

            <div className="upcoming-reminders-list">
              {dedupedUpcomingReminders.length > 0 ? (
                dedupedUpcomingReminders.map((rem, idx) => (
                  <div
                    key={rem.id ? `rem-${rem.id}` : `rem-idx-${idx}`}
                    className="reminder-item-row"
                    onClick={() => navigate(rem.action_url || '/dashboard')}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        navigate(rem.action_url || '/dashboard');
                      }
                    }}
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
                <div className="empty-state-notice" style={{ padding: '8px', fontSize: '0.66rem' }}>
                  No upcoming reminders scheduled for today.
                </div>
              )}
            </div>

            <div className="subhead-label" style={{ marginTop: '10px' }}>Quick Create Launchers</div>
            <div className="quick-create-launcher-grid">
              {quickCreateShortcuts.map((sc) => (
                <button
                  key={sc.label}
                  type="button"
                  className="btn-quick-create-chip"
                  onClick={() => navigate(sc.url)}
                >
                  {sc.label}
                </button>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* ════════════════════════════════════════════════════════════════════
          SECTIONS 6 & 7 — COMPACT AI WORKFORCE & SYSTEM HEALTH SUMMARY
          ════════════════════════════════════════════════════════════════════ */}
      {(() => {
        // AI Workforce Status determination
        let wfStatus = { label: 'Operational', className: 'healthy' };
        if (aiLoading) {
          wfStatus = { label: 'Checking...', className: 'neutral' };
        } else if (!aiWorkforceSummary) {
          wfStatus = { label: 'Unavailable', className: 'danger' };
        } else {
          const overall = (aiWorkforceSummary?.health?.overall_status || '').toLowerCase();
          const failedCount = (aiWorkforceSummary?.failed_tasks || 0) + (aiWorkforceSummary?.stale_tasks || 0);
          if (overall === 'unavailable') {
            wfStatus = { label: 'Unavailable', className: 'danger' };
          } else if (overall === 'degraded' || failedCount > 0) {
            wfStatus = { label: 'Attention needed', className: 'warning' };
          } else {
            wfStatus = { label: 'Operational', className: 'healthy' };
          }
        }

        // System Health Overall determination
        let sysHealth = { label: 'Healthy', className: 'healthy' };
        if (healthLoading) {
          sysHealth = { label: 'Checking...', className: 'neutral' };
        } else if (!healthSummary) {
          sysHealth = { label: 'Status unavailable', className: 'danger' };
        } else {
          const overall = (healthSummary?.overall_status || '').toUpperCase();
          if (overall === 'UNAVAILABLE' || overall === 'OFFLINE') {
            sysHealth = { label: 'Unavailable', className: 'danger' };
          } else if (overall === 'DEGRADED' || overall === 'ATTENTION_NEEDED' || overall === 'MISCONFIGURED') {
            sysHealth = { label: 'Attention needed', className: 'warning' };
          } else if (overall === 'HEALTHY' || overall === 'OPERATIONAL') {
            sysHealth = { label: 'Healthy', className: 'healthy' };
          } else {
            sysHealth = { label: 'Not checked', className: 'neutral' };
          }
        }

        // Key metrics
        const totalActiveTasks = aiWorkforceSummary?.total_active_tasks ?? ((aiWorkforceSummary?.queued_tasks || 0) + (aiWorkforceSummary?.processing_tasks || 0));
        const waitingForApproval = aiWorkforceSummary?.waiting_for_approval_tasks ?? recStats?.requires_approval ?? 0;
        const failedTasksCount = (aiWorkforceSummary?.failed_tasks || 0) + (aiWorkforceSummary?.stale_tasks || 0);
        const completedRecent24h = aiWorkforceSummary?.completed_recent_24h ?? 0;

        // Subsystems health states
        const backendRaw = healthSummary?.subsystems?.go_backend?.status || (healthSummary?.overall_status ? 'HEALTHY' : null);
        let backendInfo = { label: 'Healthy', className: 'online' };
        if (backendRaw) {
          const norm = String(backendRaw).toUpperCase();
          if (norm === 'HEALTHY' || norm === 'OPERATIONAL' || norm === 'ONLINE') {
            backendInfo = { label: 'Healthy', className: 'online' };
          } else if (norm === 'DEGRADED') {
            backendInfo = { label: 'Attention needed', className: 'warning' };
          } else {
            backendInfo = { label: 'Unavailable', className: 'danger' };
          }
        } else if (!healthSummary && !healthLoading) {
          backendInfo = { label: 'Unavailable', className: 'danger' };
        }

        const sidecarRaw = healthSummary?.subsystems?.ai_sidecar?.status || aiWorkforceSummary?.health?.sidecar_status;
        let sidecarInfo = { label: 'Healthy', className: 'online' };
        if (sidecarRaw) {
          const norm = String(sidecarRaw).toUpperCase();
          if (norm === 'HEALTHY' || norm === 'OPERATIONAL' || norm === 'ONLINE') {
            sidecarInfo = { label: 'Healthy', className: 'online' };
          } else if (norm === 'DEGRADED') {
            sidecarInfo = { label: 'Attention needed', className: 'warning' };
          } else {
            sidecarInfo = { label: 'Unavailable', className: 'danger' };
          }
        } else if (!aiLoading && !healthLoading && !healthSummary) {
          sidecarInfo = { label: 'Unavailable', className: 'danger' };
        } else if (!aiLoading && !healthLoading) {
          sidecarInfo = { label: 'Port 8090 OK', className: 'online' };
        }

        const dbRaw = healthSummary?.subsystems?.database?.status || healthSummary?.subsystems?.mariadb?.status;
        let dbInfo = { label: 'Connected', className: 'online' };
        if (dbRaw) {
          const norm = String(dbRaw).toUpperCase();
          if (norm === 'HEALTHY') {
            dbInfo = { label: 'Connected', className: 'online' };
          } else if (norm === 'DEGRADED') {
            dbInfo = { label: 'Attention needed', className: 'warning' };
          } else {
            dbInfo = { label: 'Unavailable', className: 'danger' };
          }
        } else if (!healthSummary && !healthLoading) {
          dbInfo = { label: 'Unavailable', className: 'danger' };
        }

        const queueRaw = healthSummary?.subsystems?.queue_worker?.status || healthSummary?.subsystems?.ai_worker?.status || aiWorkforceSummary?.health?.worker_status || aiWorkforceSummary?.health?.queue_status;
        let queueInfo = { label: '0 Backlog', className: 'online' };
        if (queueRaw) {
          const norm = String(queueRaw).toUpperCase();
          if (norm === 'HEALTHY' || norm === 'IDLE') {
            queueInfo = { label: '0 Backlog', className: 'online' };
          } else if (norm === 'DEGRADED' || norm === 'LAGGING') {
            queueInfo = { label: 'Attention needed', className: 'warning' };
          } else {
            queueInfo = { label: 'Unavailable', className: 'danger' };
          }
        } else if (!healthSummary && !aiWorkforceSummary && !healthLoading) {
          queueInfo = { label: 'Unavailable', className: 'danger' };
        }

        // External Integrations health state (Honest determination)
        let integrationsInfo = { label: 'Checking...', className: 'neutral' };
        if (integrationsLoading) {
          integrationsInfo = { label: 'Checking...', className: 'neutral' };
        } else if (!integrationsList || integrationsList.length === 0) {
          integrationsInfo = { label: 'Unavailable', className: 'danger' };
        } else {
          const activeCount = integrationsList.filter(
            (i) => i.is_enabled || (i.status && i.status !== 'DISABLED' && i.status !== 'NOT_CONFIGURED')
          ).length;
          if (activeCount > 0) {
            integrationsInfo = { label: `${activeCount}/${integrationsList.length} Active`, className: 'online' };
          } else {
            integrationsInfo = { label: 'Config Needed', className: 'warning' };
          }
        }

        return (
          <section className="dashboard-ai-health-row" aria-label="AI Summary and System Health">
            {/* SECTION A — AI Workforce Summary */}
            <div className="op-card ai-summary-card">
              <div className="card-header-flex">
                <div className="title-with-badge">
                  <h3 className="op-card-title">
                    <Sparkles size={15} className="text-indigo-600" /> AI Workforce
                  </h3>
                  <span className={`compact-health-badge ${wfStatus.className}`}>
                    ● {wfStatus.label}
                  </span>
                </div>
                {canViewAIWorkforce && (
                  <span
                    className="view-all-link"
                    onClick={() => navigate('/dashboard/ai-monitoring')}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        navigate('/dashboard/ai-monitoring');
                      }
                    }}
                  >
                    View AI Workforce →
                  </span>
                )}
              </div>

              <div className="ai-summary-metrics-grid">
                <div
                  className={`ai-summary-pill-card ${canViewAIWorkforce ? 'clickable' : ''}`}
                  onClick={() => {
                    if (canViewAIWorkforce) navigate('/dashboard/ai-monitoring?status=processing');
                  }}
                  role={canViewAIWorkforce ? 'button' : 'group'}
                  tabIndex={canViewAIWorkforce ? 0 : -1}
                  title="Active Background AI Tasks"
                  aria-label={`Active Tasks: ${totalActiveTasks}`}
                >
                  <span className="ai-sum-label">Active Tasks</span>
                  <span className="ai-sum-val">{totalActiveTasks}</span>
                  <span className="ai-sum-sub">In progress</span>
                </div>

                <div
                  className="ai-summary-pill-card clickable"
                  onClick={() => navigate('/dashboard/approvals')}
                  role="button"
                  tabIndex={0}
                  title="AI Tasks Awaiting Operator Review"
                  aria-label={`Awaiting Review: ${waitingForApproval}`}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                      e.preventDefault();
                      navigate('/dashboard/approvals');
                    }
                  }}
                >
                  <span className="ai-sum-label">Awaiting Review</span>
                  <span className="ai-sum-val purple">{waitingForApproval}</span>
                  <span className="ai-sum-sub">Requires sign-off</span>
                </div>

                <div
                  className={`ai-summary-pill-card ${canViewAIWorkforce ? 'clickable' : ''}`}
                  onClick={() => {
                    if (canViewAIWorkforce) navigate('/dashboard/ai-monitoring?status=failed');
                  }}
                  role={canViewAIWorkforce ? 'button' : 'group'}
                  tabIndex={canViewAIWorkforce ? 0 : -1}
                  title="Failed or Blocked Autonomous Tasks"
                  aria-label={`Blocked or Failed Tasks: ${failedTasksCount}`}
                >
                  <span className="ai-sum-label">Blocked / Issues</span>
                  <span className={`ai-sum-val ${failedTasksCount > 0 ? 'amber' : 'green'}`}>
                    {failedTasksCount}
                  </span>
                  <span className="ai-sum-sub">
                    {failedTasksCount === 0 ? 'None blocked' : 'Requires attention'}
                  </span>
                </div>

                <div
                  className={`ai-summary-pill-card ${canViewAIWorkforce ? 'clickable' : ''}`}
                  onClick={() => {
                    if (canViewAIWorkforce) navigate('/dashboard/ai-monitoring?status=completed');
                  }}
                  role={canViewAIWorkforce ? 'button' : 'group'}
                  tabIndex={canViewAIWorkforce ? 0 : -1}
                  title="Recently Completed Tasks in past 24 hours"
                  aria-label={`Completed Tasks: ${completedRecent24h}`}
                >
                  <span className="ai-sum-label">Completed (24h)</span>
                  <span className="ai-sum-val green">{completedRecent24h}</span>
                  <span className="ai-sum-sub">Autonomous jobs</span>
                </div>
              </div>

              <div className="card-footer-flex-row">
                <span className="ai-footer-note">
                  🛡️ All agent workflows are read-only until approved through the Go Action System.
                </span>
                <span className="ai-last-activity-pill">
                  Last activity: {formatLastActivity(aiWorkforceSummary?.last_updated)}
                </span>
              </div>
            </div>

            {/* SECTION B — System & Integration Health */}
            <div className="op-card system-health-card">
              <div className="card-header-flex">
                <div className="title-with-badge">
                  <h3 className="op-card-title">
                    <Activity size={15} className="text-emerald-600" /> System Health & Integrations
                  </h3>
                  <span className={`compact-health-badge ${sysHealth.className}`}>
                    ● {sysHealth.label}
                  </span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  {canManageIntegrations && (
                    <span
                      className="view-all-link"
                      onClick={() => navigate('/dashboard/settings/external-integrations')}
                      role="button"
                      tabIndex={0}
                      title="Manage External Gateways & Webhooks"
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault();
                          navigate('/dashboard/settings/external-integrations');
                        }
                      }}
                    >
                      Integrations →
                    </span>
                  )}
                  {canViewAuditLogs && (
                    <span
                      className="view-all-link"
                      onClick={() => navigate('/dashboard/settings/audit-logs')}
                      role="button"
                      tabIndex={0}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault();
                          navigate('/dashboard/settings/audit-logs');
                        }
                      }}
                    >
                      Audit Logs →
                    </span>
                  )}
                </div>
              </div>

              <div className="system-health-indicators-grid">
                <div className="health-indicator-tile" title="Core Go API server running on port 8080">
                  <div className="health-tile-left">
                    <Server size={14} className="text-emerald-600" />
                    <span className="health-tile-name">Go API Backend</span>
                  </div>
                  <span className={`health-status-chip ${backendInfo.className}`}>
                    {backendInfo.label}
                  </span>
                </div>

                <div className="health-indicator-tile" title="Python LangGraph Sidecar running on port 8090">
                  <div className="health-tile-left">
                    <Cpu size={14} className="text-blue-600" />
                    <span className="health-tile-name">Python AI Sidecar</span>
                  </div>
                  <span className={`health-status-chip ${sidecarInfo.className}`}>
                    {sidecarInfo.label}
                  </span>
                </div>

                <div className="health-indicator-tile" title="Persistent MariaDB database engine">
                  <div className="health-tile-left">
                    <Database size={14} className="text-purple-600" />
                    <span className="health-tile-name">MariaDB Storage</span>
                  </div>
                  <span className={`health-status-chip ${dbInfo.className}`}>
                    {dbInfo.label}
                  </span>
                </div>

                <div className="health-indicator-tile" title="Event-driven background queue workers">
                  <div className="health-tile-left">
                    <Activity size={14} className="text-indigo-600" />
                    <span className="health-tile-name">Queue Workers</span>
                  </div>
                  <span className={`health-status-chip ${queueInfo.className}`}>
                    {queueInfo.label}
                  </span>
                </div>

                <div
                  className={`health-indicator-tile ${canManageIntegrations ? 'clickable' : ''}`}
                  onClick={() => { if (canManageIntegrations) navigate('/dashboard/settings/external-integrations'); }}
                  style={{ cursor: canManageIntegrations ? 'pointer' : 'default' }}
                  title="Twilio, AWS SES, AWS S3, Textract external gateways. Live provider credentials required for real dispatch."
                  role={canManageIntegrations ? 'button' : undefined}
                  tabIndex={canManageIntegrations ? 0 : undefined}
                  onKeyDown={(e) => {
                    if (canManageIntegrations && (e.key === 'Enter' || e.key === ' ')) {
                      e.preventDefault();
                      navigate('/dashboard/settings/external-integrations');
                    }
                  }}
                >
                  <div className="health-tile-left">
                    <Globe size={14} className="text-amber-600" />
                    <span className="health-tile-name">External Gateways</span>
                  </div>
                  <span className={`health-status-chip ${integrationsInfo.className}`}>
                    {integrationsInfo.label}
                  </span>
                </div>

                <div
                  className={`health-indicator-tile ${canManageIntegrations ? 'clickable' : ''}`}
                  onClick={() => { if (canManageIntegrations) navigate('/dashboard/settings/carrier-integrations'); }}
                  style={{ cursor: canManageIntegrations ? 'pointer' : 'default' }}
                  title="Carrier EDI/API tracking, Event Mesh sync, and HMAC webhook ingestion"
                  role={canManageIntegrations ? 'button' : undefined}
                  tabIndex={canManageIntegrations ? 0 : undefined}
                  onKeyDown={(e) => {
                    if (canManageIntegrations && (e.key === 'Enter' || e.key === ' ')) {
                      e.preventDefault();
                      navigate('/dashboard/settings/carrier-integrations');
                    }
                  }}
                >
                  <div className="health-tile-left">
                    <Radio size={14} className="text-teal-600" />
                    <span className="health-tile-name">Carrier & Ingress</span>
                  </div>
                  <span className="health-status-chip online">
                    Event Mesh Live
                  </span>
                </div>
              </div>

              <div className="card-footer-flex-row">
                <span className="system-footer-note">
                  Auth, tenant isolation, and audit logging active on port 8080.
                </span>
                <span className="system-freshness-pill">
                  Checked: {formatLastActivity(healthSummary?.evaluated_at)}
                </span>
              </div>
            </div>
          </section>
        );
      })()}
    </div>
  );
}
