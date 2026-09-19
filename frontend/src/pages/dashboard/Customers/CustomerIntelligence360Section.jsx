import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
  Sparkles,
  TrendingUp,
  ShieldAlert,
  AlertTriangle,
  CheckCircle2,
  Clock,
  DollarSign,
  Package,
  FileText,
  FileCheck,
  RefreshCw,
  Info,
  ExternalLink,
  Layers,
  ArrowUpRight,
  HelpCircle,
  Calendar,
  CreditCard,
  UserCheck,
  Briefcase
} from 'lucide-react';
import { getCustomer360Intelligence } from '../../../services/customerService';
import { recommendationService } from '../../../services/recommendationService';
import './CustomerIntelligence360Section.css';

export default function CustomerIntelligence360Section({ customerId }) {
  const [data, setData] = useState(null);
  const [collectionSummary, setCollectionSummary] = useState(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState(null);

  const fetchIntelligence = useCallback(async (isManualRefresh = false) => {
    if (!customerId) return;
    if (isManualRefresh) setRefreshing(true);
    else setLoading(true);
    setError(null);

    try {
      const result = await getCustomer360Intelligence(customerId);
      setData(result);

      // Task 2.5: Customer Collection Summary & Delinquency Signals
      try {
        if (recommendationService?.getCustomerCollectionSummary) {
          const collRes = await recommendationService.getCustomerCollectionSummary(customerId);
          if (collRes && collRes.data) {
            setCollectionSummary(collRes.data);
          }
        }
      } catch (collErr) {
        console.warn('Could not fetch customer collection summary:', collErr);
      }
    } catch (err) {
      console.error('Failed to load customer 360 intelligence:', err);
      setError(err?.response?.data?.error || err.message || 'Failed to load customer intelligence');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [customerId]);

  useEffect(() => {
    fetchIntelligence(false);
  }, [fetchIntelligence]);

  const formatCurrency = (val) => {
    if (val === null || val === undefined || isNaN(val)) return '$0.00';
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      maximumFractionDigits: 0
    }).format(val);
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '—';
    try {
      const d = new Date(dateStr);
      if (isNaN(d.getTime())) return '—';
      return d.toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric'
      });
    } catch {
      return '—';
    }
  };

  if (loading) {
    return (
      <div className="cust-intel-loading-card">
        <div className="cust-intel-spinner" />
        <div className="cust-intel-loading-text">
          <strong>Assembling Customer 360° Intelligence</strong>
          <span>Aggregating real-time RFQs, quotations, shipments, invoices, contracts, and audit trails...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="cust-intel-error-card">
        <AlertTriangle size={24} className="text-rose-600" />
        <div className="cust-intel-error-text">
          <strong>Unable to load Customer 360° Intelligence</strong>
          <p>{error}</p>
        </div>
        <button
          type="button"
          className="cust-intel-btn-retry"
          onClick={() => fetchIntelligence(true)}
        >
          <RefreshCw size={14} /> Retry
        </button>
      </div>
    );
  }

  if (!data) return null;

  const {
    identity = {},
    commercial_metrics: comm = {},
    operations_metrics: ops = {},
    financial_metrics: fin = {},
    governance_metrics: gov = {},
    attention_items = [],
    ai_summary: ai = {},
    correlation_id: corrId = ''
  } = data;

  const engagementTrend = (comm.engagement_trend || 'MODERATE').toUpperCase();
  const getTrendBadgeClass = () => {
    switch (engagementTrend) {
      case 'HIGH': return 'badge-trend-high';
      case 'MODERATE': return 'badge-trend-mod';
      case 'LOW': return 'badge-trend-low';
      default: return 'badge-trend-inactive';
    }
  };

  return (
    <div className="cust-intel-container" id="customer-intelligence-360">
      
      {/* ── TOP HEADER / TOOLBAR ── */}
      <div className="cust-intel-header">
        <div className="cust-intel-header-info">
          <div className="cust-intel-title-row">
            <span className="cust-intel-badge-readonly">READ-ONLY 360° CONTEXT</span>
            <span className={`cust-intel-trend-badge ${getTrendBadgeClass()}`}>
              <TrendingUp size={13} /> {engagementTrend} ENGAGEMENT
            </span>
            {ai.confidence_level && (
              <span className="cust-intel-confidence-badge">
                <CheckCircle2 size={13} className="text-emerald-600" />
                {ai.confidence_level} CONFIDENCE
              </span>
            )}
          </div>
          <p className="cust-intel-header-sub">
            Deterministic commercial, operational, and financial intelligence grounded across active LogisticsHQ records.
          </p>
        </div>

        <div className="cust-intel-header-actions">
          {ai.data_freshness && (
            <span className="cust-intel-freshness" title="Data freshness">
              <Clock size={12} /> {ai.data_freshness}
            </span>
          )}
          <button
            type="button"
            className="cust-intel-refresh-btn"
            disabled={refreshing}
            onClick={() => fetchIntelligence(true)}
            id="btn-refresh-customer-intel"
          >
            <RefreshCw size={13} className={refreshing ? 'cust-intel-spin' : ''} />
            {refreshing ? 'Evaluating...' : 'Refresh 360° Intelligence'}
          </button>
        </div>
      </div>

      {/* ── GROUNDED READ-ONLY AI EXECUTIVE SUMMARY ── */}
      <div className="cust-intel-card cust-intel-ai-summary-card">
        <div className="cust-intel-ai-header">
          <div className="cust-intel-ai-header-left">
            <div className="cust-intel-ai-icon-box">
              <Sparkles size={18} className="text-blue-600" />
            </div>
            <div>
              <h3 className="cust-intel-ai-title">Customer Intelligence Summary</h3>
              <span className="cust-intel-ai-classification">
                Automated synthesized account view &bull; Strictly informational &bull; No autonomous writes
              </span>
            </div>
          </div>
          {corrId && (
            <span className="cust-intel-corrid" title={`Correlation ID: ${corrId}`}>
              ID: {corrId.slice(0, 18)}...
            </span>
          )}
        </div>

        <div className="cust-intel-ai-body">
          <p className="cust-intel-ai-lead">{ai.executive_summary || 'No summary available.'}</p>

          <div className="cust-intel-ai-columns">
            <div className="cust-intel-ai-column">
              <span className="cust-intel-column-label">
                <Briefcase size={13} className="text-blue-600" /> Commercial Position
              </span>
              <p>{ai.commercial_position || 'Commercial status stable.'}</p>
            </div>

            <div className="cust-intel-ai-column">
              <span className="cust-intel-column-label">
                <Package size={13} className="text-emerald-600" /> Operational Position
              </span>
              <p>{ai.operational_position || 'Operations running as scheduled.'}</p>
            </div>

            <div className="cust-intel-ai-column">
              <span className="cust-intel-column-label">
                <DollarSign size={13} className="text-amber-600" /> Financial & Approvals
              </span>
              <p>{ai.financial_and_approval_concerns || 'Financial standing good.'}</p>
            </div>
          </div>

          {/* Attention Items Alert Box */}
          {attention_items.length > 0 ? (
            <div className="cust-intel-attention-box">
              <div className="cust-intel-attention-header">
                <AlertTriangle size={15} className="text-amber-600" />
                <strong>Key Attention & Risk Items ({attention_items.length})</strong>
              </div>
              <ul className="cust-intel-attention-list">
                {attention_items.map((item, idx) => (
                  <li key={idx}>
                    <span className="cust-intel-bullet" />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </div>
          ) : (
            <div className="cust-intel-good-standing-box">
              <CheckCircle2 size={15} className="text-emerald-600" />
              <span>No critical attention items detected. Account is operational with zero outstanding blockers.</span>
            </div>
          )}

          {/* Recommended Human Actions */}
          {Array.isArray(ai.recommended_follow_up_actions) && ai.recommended_follow_up_actions.length > 0 && (
            <div className="cust-intel-recommendations-section">
              <span className="cust-intel-rec-title">Recommended Human Actions</span>
              <div className="cust-intel-rec-grid">
                {ai.recommended_follow_up_actions.map((act, idx) => (
                  <div className="cust-intel-rec-item" key={idx}>
                    <span className="cust-intel-rec-num">{idx + 1}</span>
                    <span className="cust-intel-rec-text">{act}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Warnings & Data Caveats */}
          {Array.isArray(ai.data_warnings) && ai.data_warnings.length > 0 && (
            <div className="cust-intel-warnings-box">
              <Info size={14} className="text-slate-500" />
              <div className="cust-intel-warnings-content">
                {ai.data_warnings.map((warn, idx) => (
                  <span key={idx} className="cust-intel-warn-line">{warn}</span>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* ── 4-QUADRANT DETERMINISTIC METRICS BREAKDOWN ── */}
      <div className="cust-intel-quad-grid">
        
        {/* QUADRANT 1: Commercial Summary */}
        <div className="cust-intel-metric-card">
          <div className="cust-intel-card-header">
            <div className="cust-intel-card-title">
              <Briefcase size={16} className="text-blue-600" />
              <span>Commercial Summary</span>
            </div>
            <span className="cust-intel-card-tag">{comm.recent_activity_count || 0} recent</span>
          </div>

          <div className="cust-intel-metric-rows">
            <div className="cust-intel-metric-row">
              <span className="metric-label">RFQs Registered</span>
              <strong className="metric-value">{comm.total_rfqs ?? 0}</strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Quotations Generated</span>
              <strong className="metric-value">{comm.total_quotations ?? 0}</strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Bookings Confirmed</span>
              <strong className="metric-value">{comm.total_bookings ?? 0}</strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Quote-to-Booking Conversion</span>
              <strong className="metric-value">
                {comm.quote_to_booking_conversion_rate !== null && comm.quote_to_booking_conversion_rate !== undefined
                  ? `${comm.quote_to_booking_conversion_rate.toFixed(1)}%`
                  : '—'}
              </strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Avg Quotation Value</span>
              <strong className="metric-value">
                {comm.average_quotation_value ? formatCurrency(comm.average_quotation_value) : '—'}
              </strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Last Commercial Interaction</span>
              <span className="metric-sub">{formatDate(comm.last_interaction_date)}</span>
            </div>
          </div>
        </div>

        {/* QUADRANT 2: Operations Summary */}
        <div className="cust-intel-metric-card">
          <div className="cust-intel-card-header">
            <div className="cust-intel-card-title">
              <Package size={16} className="text-emerald-600" />
              <span>Operations Summary</span>
            </div>
            <span className="cust-intel-card-tag">{ops.total_shipments ?? 0} Total</span>
          </div>

          <div className="cust-intel-metric-rows">
            <div className="cust-intel-metric-row">
              <span className="metric-label">Active In-Transit Shipments</span>
              <strong className="metric-value text-emerald-700">{ops.active_shipment_count ?? 0}</strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Delayed Shipments</span>
              <strong className={`metric-value ${ops.delayed_shipment_count > 0 ? 'text-amber-600 font-bold' : ''}`}>
                {ops.delayed_shipment_count ?? 0}
              </strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Open Operational Exceptions</span>
              <strong className={`metric-value ${ops.open_exception_count > 0 ? 'text-rose-600 font-bold' : ''}`}>
                {ops.open_exception_count ?? 0}
              </strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Avg Transit Cycle</span>
              <strong className="metric-value">
                {ops.average_shipment_cycle_days ? `${ops.average_shipment_cycle_days.toFixed(1)} days` : '—'}
              </strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Last Shipment Date</span>
              <span className="metric-sub">{formatDate(ops.last_shipment_date)}</span>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Performance Status</span>
              <span className="metric-sub">{ops.delivery_performance_summary || 'Normal'}</span>
            </div>
          </div>
        </div>

        {/* QUADRANT 3: Financial Summary */}
        <div className="cust-intel-metric-card">
          <div className="cust-intel-card-header">
            <div className="cust-intel-card-title">
              <DollarSign size={16} className="text-violet-600" />
              <span>Financial Summary</span>
            </div>
            <span className="cust-intel-card-tag">{fin.payment_terms || 'NET30'}</span>
          </div>

          <div className="cust-intel-metric-rows">
            <div className="cust-intel-metric-row">
              <span className="metric-label">Total Invoiced (All-time)</span>
              <strong className="metric-value">{formatCurrency(fin.total_invoiced_amount)}</strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Total Paid Amount</span>
              <strong className="metric-value text-emerald-700">{formatCurrency(fin.total_paid_amount)}</strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Outstanding Balance</span>
              <strong className={`metric-value ${fin.outstanding_balance > 0 ? 'text-rose-700' : ''}`}>
                {formatCurrency(fin.outstanding_balance)}
              </strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Overdue Invoices</span>
              <strong className={`metric-value ${fin.overdue_invoice_count > 0 ? 'text-rose-600 font-bold' : ''}`}>
                {fin.overdue_invoice_count ?? 0}
              </strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Approved Credit Limit</span>
              <strong className="metric-value">{formatCurrency(fin.credit_limit)}</strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Credit Status</span>
              <span className="cust-intel-credit-tag">{fin.credit_status || 'GOOD_STANDING'}</span>
            </div>
            {collectionSummary && (
              <>
                <div className="cust-intel-metric-row">
                  <span className="metric-label">Collection Completion</span>
                  <strong className="metric-value text-emerald-700">
                    {collectionSummary.payment_completion_rate ? `${collectionSummary.payment_completion_rate.toFixed(1)}%` : '100%'}
                  </strong>
                </div>
                <div className="cust-intel-metric-row">
                  <span className="metric-label">Avg Payment Delay</span>
                  <span className="metric-sub">
                    {collectionSummary.average_payment_delay_days ? `${collectionSummary.average_payment_delay_days.toFixed(1)} days` : '0 days (on-time)'}
                  </span>
                </div>
                {collectionSummary.risk_signals && collectionSummary.risk_signals.length > 0 && (
                  <div className="cust-intel-metric-row" style={{ flexDirection: 'column', alignItems: 'flex-start', gap: '4px' }}>
                    <span className="metric-label">Collection Risk Signals:</span>
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px' }}>
                      {collectionSummary.risk_signals.map((sig, idx) => (
                        <span key={idx} style={{ fontSize: '10px', padding: '1px 6px', borderRadius: '3px', backgroundColor: '#fef2f2', color: '#b91c1c', border: '1px solid #fecaca' }}>
                          {sig}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
                <div className="cust-intel-metric-row" style={{ paddingTop: '6px', borderTop: '1px dashed #e2e8f0' }}>
                  <Link
                    to={`/dashboard/recommendations?customer_id=${customerId}&category=finance`}
                    style={{ fontSize: '11px', color: '#2563eb', fontWeight: 600, display: 'inline-flex', alignItems: 'center', gap: '4px', textDecoration: 'none' }}
                  >
                    <span>Collections Assistant Actions</span>
                    <ArrowUpRight size={12} />
                  </Link>
                </div>
              </>
            )}
          </div>
        </div>

        {/* QUADRANT 4: Governance & Activity */}
        <div className="cust-intel-metric-card">
          <div className="cust-intel-card-header">
            <div className="cust-intel-card-title">
              <FileCheck size={16} className="text-slate-700" />
              <span>Governance & Controls</span>
            </div>
            <span className="cust-intel-card-tag">Strict Scoping</span>
          </div>

          <div className="cust-intel-metric-rows">
            <div className="cust-intel-metric-row">
              <span className="metric-label">Pending Approval Requests</span>
              <strong className={`metric-value ${gov.open_approval_count > 0 ? 'text-amber-600 font-bold' : ''}`}>
                {gov.open_approval_count ?? 0}
              </strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Recent AI Tasks Executed</span>
              <strong className="metric-value">{gov.recent_ai_task_count ?? 0}</strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Audit Trail Events</span>
              <strong className="metric-value">{gov.relevant_audit_count ?? 0}</strong>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Account Manager</span>
              <span className="metric-sub">{identity.account_manager_name || 'Unassigned'}</span>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Customer Tier / Stage</span>
              <span className="metric-sub">{identity.tier || 'STANDARD'} &bull; {identity.lifecycle_stage || 'ACTIVE'}</span>
            </div>
            <div className="cust-intel-metric-row">
              <span className="metric-label">Last Audit Action</span>
              <span className="metric-sub">{formatDate(gov.last_audit_timestamp)}</span>
            </div>
          </div>
        </div>

      </div>

      {/* ── GROUNDED TRACEABILITY: OBSERVATIONS & SUPPORTING RECORDS ── */}
      {Array.isArray(ai.supporting_observations) && ai.supporting_observations.length > 0 && (
        <div className="cust-intel-card cust-intel-traceability-card">
          <div className="cust-intel-card-header">
            <div className="cust-intel-card-title">
              <Layers size={16} className="text-blue-600" />
              <span>Supporting Record Traceability</span>
            </div>
            <span className="cust-intel-verified-tag">
              <CheckCircle2 size={13} className="text-emerald-600" /> Verified Record References
            </span>
          </div>

          <div className="cust-intel-observations-table-wrapper">
            <table className="cust-intel-table">
              <thead>
                <tr>
                  <th>Module</th>
                  <th>Record Type</th>
                  <th>Record ID</th>
                  <th>Date</th>
                  <th>Grounded Observation & Evidence</th>
                </tr>
              </thead>
              <tbody>
                {ai.supporting_observations.map((obs, idx) => (
                  <tr key={idx}>
                    <td>
                      <span className="cust-intel-module-badge">{obs.source_module || 'SYSTEM'}</span>
                    </td>
                    <td className="cust-intel-type-cell">
                      {obs.record_type || 'RECORD'}
                    </td>
                    <td>
                      <span className="cust-intel-record-id-chip">
                        {obs.record_id || '—'}
                      </span>
                    </td>
                    <td className="cust-intel-date-cell">
                      {formatDate(obs.timestamp)}
                    </td>
                    <td className="cust-intel-explanation-cell">
                      {obs.explanation || 'Referenced in customer operational context.'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Authorized Contacts Summary */}
      {Array.isArray(identity.authorized_contacts) && identity.authorized_contacts.length > 0 && (
        <div className="cust-intel-card cust-intel-contacts-card">
          <div className="cust-intel-card-header">
            <div className="cust-intel-card-title">
              <UserCheck size={16} className="text-slate-700" />
              <span>Authorized Key Contacts ({identity.authorized_contacts.length})</span>
            </div>
            <span className="cust-intel-card-tag">RBAC Verified</span>
          </div>
          <div className="cust-intel-contacts-grid">
            {identity.authorized_contacts.map((contact) => (
              <div className="cust-intel-contact-pill" key={contact.id}>
                <div className="cust-intel-contact-avatar">
                  {contact.first_name?.[0] || 'C'}{contact.last_name?.[0] || ''}
                </div>
                <div className="cust-intel-contact-info">
                  <div className="cust-intel-contact-name">
                    <strong>{contact.first_name} {contact.last_name}</strong>
                    {contact.is_primary && <span className="badge-primary-contact">Primary</span>}
                  </div>
                  <span className="cust-intel-contact-role">{contact.job_title || 'Contact'}</span>
                  <span className="cust-intel-contact-email">{contact.email || '—'}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

    </div>
  );
}
