import React, { useState, useEffect, useCallback } from 'react';
import {
  Ship,
  Clock,
  ShieldCheck,
  AlertTriangle,
  RefreshCw,
  Sparkles,
  CheckCircle2,
  HelpCircle,
  FileText,
  AlertCircle,
  Calendar,
  Layers,
  Activity,
  ArrowUpRight,
  Info
} from 'lucide-react';
import { shipmentService } from '../../../../services/shipmentService';
import './ShipmentOperationsIntelligenceSection.css';

/**
 * Format date nicely
 */
function formatDateTime(dateStr) {
  if (!dateStr) return '—';
  try {
    const d = new Date(dateStr);
    return d.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return dateStr;
  }
}

/**
 * Format duration hours into readable format
 */
function formatHours(hours) {
  if (hours === null || hours === undefined) return '—';
  if (Math.abs(hours) < 1) {
    const mins = Math.round(hours * 60);
    return `${mins}m`;
  }
  if (Math.abs(hours) < 24) {
    const rounded = Math.round(hours * 10) / 10;
    return `${rounded}h`;
  }
  const days = Math.round((hours / 24) * 10) / 10;
  return `${days}d`;
}

export default function ShipmentOperationsIntelligenceSection({ shipmentId }) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchIntelligence = useCallback(async () => {
    if (!shipmentId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await shipmentService.getShipment360OperationsIntelligence(shipmentId);
      const payload = res?.data || res;
      setData(payload);
    } catch (err) {
      console.error('Failed to load shipment operations intelligence:', err);
      setError(err?.response?.data?.error || err?.message || 'Failed to load operational intelligence');
    } finally {
      setLoading(false);
    }
  }, [shipmentId]);

  useEffect(() => {
    fetchIntelligence();
  }, [fetchIntelligence]);

  if (loading) {
    return (
      <div className="sh-intel-loading" data-testid="sh-intel-loading">
        <div className="sh-intel-spinner" />
        <p className="sh-intel-loading-text">Computing deterministic shipment milestone progress & operational risks...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="sh-intel-error" data-testid="sh-intel-error">
        <AlertTriangle size={28} className="text-amber-500 mb-2" />
        <h4 className="font-semibold text-slate-800">Operational Intelligence Unavailable</h4>
        <p className="text-sm text-slate-500 mb-4">{error}</p>
        <button className="sh-intel-btn sh-intel-btn--secondary" onClick={fetchIntelligence}>
          <RefreshCw size={14} className="mr-1.5" />
          Retry Calculation
        </button>
      </div>
    );
  }

  if (!data) return null;

  const {
    identity = {},
    milestones = {},
    exceptions = {},
    performance = {},
    risk_indicators = {},
    ai_summary = {},
    data_freshness_timestamp,
    correlation_id,
  } = data;

  const riskRating = risk_indicators.overall_risk_rating || 'LOW';

  return (
    <div className="sh-intel-container" data-testid="shipment-operations-intelligence-section">
      {/* Top Header Bar */}
      <div className="sh-intel-header">
        <div className="sh-intel-header-left">
          <div className="sh-intel-badge sh-intel-badge--readonly">
            <ShieldCheck size={14} className="sh-intel-badge-icon" />
            <span>READ-ONLY OPERATIONS INTELLIGENCE</span>
          </div>
          <div className="sh-intel-title-row">
            <h3 className="sh-intel-title">Shipment & Operational Risk Intelligence</h3>
            <span className={`sh-intel-risk-pill sh-intel-risk-pill--${riskRating.toLowerCase()}`} data-testid="sh-intel-risk-rating">
              Risk: {riskRating} ({risk_indicators.risk_score || 0}/100)
            </span>
            {performance.carrier_update_freshness && (
              <span className={`sh-intel-freshness-pill sh-intel-freshness-pill--${performance.carrier_update_freshness.toLowerCase()}`}>
                Tracking: {performance.carrier_update_freshness}
              </span>
            )}
          </div>
          <p className="sh-intel-subtitle">
            Deterministic milestone progression, schedule variance, exception tracking, and grounded risk indicators for {identity.shipment_number || `Shipment #${shipmentId}`}.
          </p>
        </div>

        <div className="sh-intel-header-right">
          <div className="sh-intel-meta">
            <div className="sh-intel-meta-item">
              <Clock size={12} />
              <span>Calculated: {data_freshness_timestamp ? new Date(data_freshness_timestamp).toLocaleTimeString() : 'Just now'}</span>
            </div>
            {correlation_id && (
              <div className="sh-intel-meta-item" title={correlation_id}>
                <span>Corr: {correlation_id.substring(0, 14)}...</span>
              </div>
            )}
          </div>
          <button
            className="sh-intel-btn sh-intel-btn--secondary"
            onClick={fetchIntelligence}
            title="Re-run deterministic operational calculations"
            data-testid="sh-intel-refresh-btn"
          >
            <RefreshCw size={14} className="sh-intel-btn-icon" />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* Grounded AI Operational Summary Card */}
      {ai_summary && ai_summary.executive_summary && (
        <div className="sh-intel-ai-card" data-testid="sh-intel-ai-card">
          <div className="sh-intel-ai-header">
            <div className="sh-intel-ai-header-left">
              <div className="sh-intel-ai-sparkle">
                <Sparkles size={16} />
              </div>
              <span className="sh-intel-ai-title">Grounded Operational Analysis</span>
              <span className="sh-intel-ai-confidence">
                Confidence: {ai_summary.confidence_score || 'HIGH'}
              </span>
            </div>
            <span className="sh-intel-ai-disclaimer">Read-Only Operational Guidance</span>
          </div>

          <p className="sh-intel-ai-text">{ai_summary.executive_summary}</p>

          <div className="sh-intel-ai-grid">
            {ai_summary.current_status_explanation && (
              <div className="sh-intel-ai-subitem">
                <div className="sh-intel-ai-subitem-title">
                  <Activity size={14} className="text-blue-600" />
                  <span>Status & Next Milestone</span>
                </div>
                <p>{ai_summary.current_status_explanation}</p>
              </div>
            )}

            {ai_summary.likely_operational_risks && (
              <div className="sh-intel-ai-subitem">
                <div className="sh-intel-ai-subitem-title">
                  <AlertTriangle size={14} className="text-amber-600" />
                  <span>Identified Operational Risks</span>
                </div>
                <p>{ai_summary.likely_operational_risks}</p>
              </div>
            )}

            {ai_summary.exception_prioritization && (
              <div className="sh-intel-ai-subitem">
                <div className="sh-intel-ai-subitem-title">
                  <AlertCircle size={14} className="text-rose-600" />
                  <span>Exception Prioritization</span>
                </div>
                <p>{ai_summary.exception_prioritization}</p>
              </div>
            )}
          </div>

          {/* Actionable Attention Items & Suggested Inquiries */}
          {(ai_summary.actionable_attention_items?.length > 0 || ai_summary.suggested_operator_inquiries?.length > 0) && (
            <div className="sh-intel-ai-actions-row">
              {ai_summary.actionable_attention_items?.length > 0 && (
                <div className="sh-intel-ai-action-col">
                  <span className="sh-intel-ai-action-heading">Actionable Attention Items:</span>
                  <ul className="sh-intel-ai-list">
                    {ai_summary.actionable_attention_items.map((item, idx) => (
                      <li key={idx} className="sh-intel-ai-list-item">
                        <CheckCircle2 size={13} className="text-indigo-500 mr-1.5 flex-shrink-0 mt-0.5" />
                        <span>{item}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {ai_summary.suggested_operator_inquiries?.length > 0 && (
                <div className="sh-intel-ai-action-col">
                  <span className="sh-intel-ai-action-heading">Suggested Follow-Up Inquiries:</span>
                  <ul className="sh-intel-ai-list">
                    {ai_summary.suggested_operator_inquiries.map((inq, idx) => (
                      <li key={idx} className="sh-intel-ai-list-item">
                        <HelpCircle size={13} className="text-slate-500 mr-1.5 flex-shrink-0 mt-0.5" />
                        <span>{inq}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}

          {/* Verifiable Source Citations */}
          {ai_summary.verifiable_source_citations?.length > 0 && (
            <div className="sh-intel-ai-citations">
              <span className="sh-intel-ai-citations-label">Grounded Sources:</span>
              {ai_summary.verifiable_source_citations.map((cite, idx) => (
                <span key={idx} className="sh-intel-citation-chip">
                  <FileText size={11} className="mr-1" />
                  {cite}
                </span>
              ))}
            </div>
          )}
        </div>
      )}

      {/* 4-Quadrant Operational KPI Grid */}
      <div className="sh-intel-kpi-grid">
        {/* Quadrant 1: Milestone Progress & Schedule */}
        <div className="sh-intel-card">
          <div className="sh-intel-card-header">
            <div className="sh-intel-card-icon-box bg-blue-50 text-blue-600">
              <Clock size={18} />
            </div>
            <div>
              <h4 className="sh-intel-card-title">Milestone Progress</h4>
              <p className="sh-intel-card-desc">Deterministic execution tracking</p>
            </div>
          </div>

          <div className="sh-intel-kpi-row">
            <div className="sh-intel-kpi-cell">
              <span className="sh-intel-kpi-label">Completed</span>
              <span className="sh-intel-kpi-val" data-testid="sh-intel-milestones-completed">
                {milestones.completed_milestones || 0}
                <span className="sh-intel-kpi-subval">
                  {' '}/ {milestones.total_milestones || 0} total
                </span>
              </span>
            </div>

            <div className="sh-intel-kpi-cell">
              <span className="sh-intel-kpi-label">Completion Rate</span>
              <span className="sh-intel-kpi-val" data-testid="sh-intel-completion-rate">
                {milestones.milestone_completion_rate || 0}%
              </span>
            </div>
          </div>

          <div className="sh-intel-stat-list">
            <div className="sh-intel-stat-row">
              <span className="sh-intel-stat-name">Next Milestone</span>
              <span className="sh-intel-stat-val font-semibold text-slate-800" data-testid="sh-intel-next-milestone">
                {milestones.next_expected_milestone_code || 'None (Completed)'}
              </span>
            </div>
            {milestones.next_expected_milestone_date && (
              <div className="sh-intel-stat-row">
                <span className="sh-intel-stat-name">Next Scheduled</span>
                <span className="sh-intel-stat-val text-slate-600 text-xs">
                  {formatDateTime(milestones.next_expected_milestone_date)}
                </span>
              </div>
            )}
            <div className="sh-intel-stat-row">
              <span className="sh-intel-stat-name">Delayed Milestones</span>
              <span className={`sh-intel-stat-val font-medium ${milestones.number_of_delayed_milestones > 0 ? 'text-rose-600' : 'text-slate-600'}`}>
                {milestones.number_of_delayed_milestones || 0}
              </span>
            </div>
          </div>
        </div>

        {/* Quadrant 2: Exceptions & Operational Blockers */}
        <div className="sh-intel-card">
          <div className="sh-intel-card-header">
            <div className="sh-intel-card-icon-box bg-rose-50 text-rose-600">
              <AlertCircle size={18} />
            </div>
            <div>
              <h4 className="sh-intel-card-title">Exceptions & Holds</h4>
              <p className="sh-intel-card-desc">Active operational blockers</p>
            </div>
          </div>

          <div className="sh-intel-kpi-row">
            <div className="sh-intel-kpi-cell">
              <span className="sh-intel-kpi-label">Open Exceptions</span>
              <span
                className={`sh-intel-kpi-val ${exceptions.open_exceptions > 0 ? 'text-rose-600' : 'text-slate-800'}`}
                data-testid="sh-intel-open-exceptions"
              >
                {exceptions.open_exceptions || 0}
                <span className="sh-intel-kpi-subval">
                  {' '}/ {exceptions.total_exceptions || 0} total
                </span>
              </span>
            </div>

            <div className="sh-intel-kpi-cell">
              <span className="sh-intel-kpi-label">Critical Severity</span>
              <span className="sh-intel-kpi-val text-rose-700" data-testid="sh-intel-critical-exceptions">
                {exceptions.critical_exceptions || 0}
              </span>
            </div>
          </div>

          <div className="sh-intel-stat-list">
            <div className="sh-intel-stat-row">
              <span className="sh-intel-stat-name">High Severity</span>
              <span className="sh-intel-stat-val font-medium text-amber-700">
                {exceptions.high_severity_exceptions || 0}
              </span>
            </div>
            {exceptions.oldest_unresolved_title && (
              <div className="sh-intel-stat-row">
                <span className="sh-intel-stat-name">Oldest Open</span>
                <span className="sh-intel-stat-val font-medium text-slate-800 text-xs truncate max-w-[160px]" title={exceptions.oldest_unresolved_title}>
                  {exceptions.oldest_unresolved_title}
                </span>
              </div>
            )}
            {exceptions.oldest_unresolved_hours_open != null && (
              <div className="sh-intel-stat-row">
                <span className="sh-intel-stat-name">Open Duration</span>
                <span className="sh-intel-stat-val font-medium text-slate-700">
                  {formatHours(exceptions.oldest_unresolved_hours_open)}
                </span>
              </div>
            )}
          </div>
        </div>

        {/* Quadrant 3: Schedule & Transit Variance */}
        <div className="sh-intel-card">
          <div className="sh-intel-card-header">
            <div className="sh-intel-card-icon-box bg-indigo-50 text-indigo-600">
              <Activity size={18} />
            </div>
            <div>
              <h4 className="sh-intel-card-title">Schedule Adherence</h4>
              <p className="sh-intel-card-desc">Transit variance & delay tracking</p>
            </div>
          </div>

          <div className="sh-intel-kpi-row">
            <div className="sh-intel-kpi-cell">
              <span className="sh-intel-kpi-label">Departure Variance</span>
              <span
                className={`sh-intel-kpi-val ${
                  performance.departure_variance_hours > 0 ? 'text-amber-700' : 'text-slate-800'
                }`}
                data-testid="sh-intel-departure-variance"
              >
                {performance.departure_variance_hours != null
                  ? `${performance.departure_variance_hours > 0 ? '+' : ''}${formatHours(performance.departure_variance_hours)}`
                  : 'On Time'}
              </span>
            </div>

            <div className="sh-intel-kpi-cell">
              <span className="sh-intel-kpi-label">Arrival Variance</span>
              <span
                className={`sh-intel-kpi-val ${
                  performance.arrival_variance_hours > 0 ? 'text-rose-600' : 'text-slate-800'
                }`}
                data-testid="sh-intel-arrival-variance"
              >
                {performance.arrival_variance_hours != null
                  ? `${performance.arrival_variance_hours > 0 ? '+' : ''}${formatHours(performance.arrival_variance_hours)}`
                  : 'On Schedule'}
              </span>
            </div>
          </div>

          <div className="sh-intel-stat-list">
            <div className="sh-intel-stat-row">
              <span className="sh-intel-stat-name">Planned Transit</span>
              <span className="sh-intel-stat-val font-medium text-slate-700">
                {performance.planned_transit_days != null ? `${performance.planned_transit_days} days` : '—'}
              </span>
            </div>
            {performance.actual_transit_days != null && (
              <div className="sh-intel-stat-row">
                <span className="sh-intel-stat-name">Actual Transit</span>
                <span className="sh-intel-stat-val font-medium text-slate-800">
                  {performance.actual_transit_days} days
                </span>
              </div>
            )}
            <div className="sh-intel-stat-row">
              <span className="sh-intel-stat-name">Total Delay Recorded</span>
              <span className="sh-intel-stat-val font-semibold text-slate-700">
                {performance.delay_duration_hours != null ? formatHours(performance.delay_duration_hours) : 'None'}
              </span>
            </div>
          </div>
        </div>

        {/* Quadrant 4: Shipment Risk Indicators */}
        <div className="sh-intel-card">
          <div className="sh-intel-card-header">
            <div className="sh-intel-card-icon-box bg-amber-50 text-amber-600">
              <ShieldCheck size={18} />
            </div>
            <div>
              <h4 className="sh-intel-card-title">Risk Health Profile</h4>
              <p className="sh-intel-card-desc">Proactive disruption assessment</p>
            </div>
          </div>

          <div className="sh-intel-kpi-row">
            <div className="sh-intel-kpi-cell">
              <span className="sh-intel-kpi-label">Risk Rating</span>
              <span className={`sh-intel-kpi-val sh-intel-risk-text--${riskRating.toLowerCase()}`}>
                {riskRating}
              </span>
            </div>

            <div className="sh-intel-kpi-cell">
              <span className="sh-intel-kpi-label">Risk Score</span>
              <span className="sh-intel-kpi-val text-slate-800">
                {risk_indicators.risk_score || 0} / 100
              </span>
            </div>
          </div>

          <div className="sh-intel-stat-list">
            <div className="sh-intel-stat-row">
              <span className="sh-intel-stat-name">Active Risk Factors</span>
              <span className="sh-intel-stat-val font-medium text-slate-800">
                {risk_indicators.active_risk_factors?.length || 0} factors
              </span>
            </div>
            <div className="sh-intel-stat-row">
              <span className="sh-intel-stat-name">Data Gaps</span>
              <span className="sh-intel-stat-val text-slate-600">
                {risk_indicators.missing_data_reasons?.length || 0} detected
              </span>
            </div>
          </div>

          {/* Missing data notes */}
          {risk_indicators.missing_data_reasons?.length > 0 && (
            <div className="sh-intel-notice-box">
              <Info size={13} className="text-slate-500 mr-1.5 flex-shrink-0 mt-0.5" />
              <div className="text-xs text-slate-600">
                {risk_indicators.missing_data_reasons.join('; ')}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Milestone Execution Timeline Table */}
      <div className="sh-intel-table-card" data-testid="sh-intel-milestones-table-card">
        <div className="sh-intel-table-header">
          <div>
            <h4 className="sh-intel-table-title">Milestone Execution & Schedule Adherence</h4>
            <p className="sh-intel-table-desc">
              Authoritative timeline for {identity.shipment_number || `Shipment #${shipmentId}`}. Planned vs. actual milestones.
            </p>
          </div>
          <span className="sh-intel-count-badge">
            {milestones.milestones_list?.length || 0} Milestones
          </span>
        </div>

        {milestones.milestones_list && milestones.milestones_list.length > 0 ? (
          <div className="sh-intel-table-wrapper">
            <table className="sh-intel-table">
              <thead>
                <tr>
                  <th>Milestone</th>
                  <th>Status</th>
                  <th>Location</th>
                  <th>Planned Date</th>
                  <th>Actual Date</th>
                  <th>Schedule Variance</th>
                </tr>
              </thead>
              <tbody>
                {milestones.milestones_list.map((m) => (
                  <tr key={m.id} className={m.is_delayed ? 'sh-intel-tr--delayed' : ''}>
                    <td className="font-semibold text-slate-800">
                      <span>{m.milestone_code}</span>
                      {m.description && (
                        <span className="text-xs text-slate-400 block font-normal">{m.description}</span>
                      )}
                    </td>
                    <td>
                      <span className={`sh-intel-mstatus-badge sh-intel-mstatus-badge--${m.status.toLowerCase()}`}>
                        {m.status}
                      </span>
                    </td>
                    <td className="text-slate-600">{m.location || '—'}</td>
                    <td className="text-slate-600 text-xs">{formatDateTime(m.planned_date)}</td>
                    <td className="text-slate-800 text-xs font-medium">{formatDateTime(m.actual_date)}</td>
                    <td>
                      {m.is_delayed ? (
                        <span className="sh-intel-delay-tag">
                          +{formatHours(m.delay_hours)} late
                        </span>
                      ) : (
                        <span className="text-xs text-slate-400 font-medium">On Schedule</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="sh-intel-empty-placeholder" data-testid="sh-intel-empty-milestones">
            <p className="text-xs text-slate-500">No operational milestones have been recorded for this shipment yet.</p>
          </div>
        )}
      </div>

      {/* Traceable Grounded Observations Ledger */}
      {data.grounded_observations && data.grounded_observations.length > 0 && (
        <div className="sh-intel-observations-card">
          <div className="sh-intel-obs-header">
            <h5 className="sh-intel-obs-title">Operational Audit Ledger & Evidence Lineage</h5>
            <span className="text-xs text-slate-400">Strictly Verified Records</span>
          </div>
          <div className="sh-intel-obs-list">
            {data.grounded_observations.map((obs, idx) => (
              <div key={idx} className="sh-intel-obs-item">
                <div className="sh-intel-obs-item-head">
                  <span className={`sh-intel-obs-severity sh-intel-obs-severity--${obs.severity.toLowerCase()}`}>
                    {obs.severity}
                  </span>
                  <span className="sh-intel-obs-category">{obs.category}</span>
                  <span className="sh-intel-obs-evidence">Evidence: {obs.evidence}</span>
                </div>
                <p className="sh-intel-obs-msg">{obs.message}</p>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
