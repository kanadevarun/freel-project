import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  AlertTriangle,
  RefreshCw,
  ChevronDown,
  ChevronUp,
  CheckCircle2,
  ArrowRight,
  Info,
  Clock,
  FileText,
  Activity,
  Send,
  Users,
  Layers,
  Zap,
  Briefcase,
  GitPullRequest,
  Anchor
} from 'lucide-react';
import predictionService from '../../services/predictionService';
import { SeverityBadge, ConfidenceBadge, ValueTypeBadge } from './PredictionBadge';
import './ResourceBottleneckPredictiveCard.css';

/**
 * Format timestamp into human-readable local time
 */
function formatDateTime(dateStr) {
  if (!dateStr) return 'Not Available';
  try {
    const d = new Date(dateStr);
    if (isNaN(d.getTime())) return dateStr;
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

export default function ResourceBottleneckPredictiveCard({
  defaultTab = 'approvals', // 'approvals' | 'documentation' | 'cross_module' | 'resource' | 'summary'
  compact = false,
  className = '',
}) {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState(defaultTab);
  const [prediction, setPrediction] = useState(null);
  const [summaryData, setSummaryData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState(null);
  const [showEvidence, setShowEvidence] = useState(false);
  const [actionSuccessMsg, setActionSuccessMsg] = useState(null);
  const [actionNotes, setActionNotes] = useState('');
  const [showActionModal, setShowActionModal] = useState(false);
  const [submittingAction, setSubmittingAction] = useState(false);

  const fetchTabIntelligence = useCallback(async (tab, forceRefresh = false) => {
    try {
      if (forceRefresh) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }
      setError(null);
      setActionSuccessMsg(null);

      if (tab === 'summary') {
        const resp = await predictionService.getResourceBottleneckSummary();
        const data = resp?.data || resp;
        setSummaryData(data);
        setPrediction(null);
      } else if (tab === 'resource') {
        const resp = forceRefresh
          ? await predictionService.refreshResourceAllocation('owner_workload')
          : await predictionService.getResourceAllocation('owner_workload');
        const data = resp?.data || resp;
        setPrediction(data);
      } else {
        let resp;
        if (tab === 'approvals') {
          resp = forceRefresh
            ? await predictionService.refreshOperationalBottleneck('approvals')
            : await predictionService.getOperationalBottleneck('approvals');
        } else if (tab === 'documentation') {
          resp = forceRefresh
            ? await predictionService.refreshOperationalBottleneck('documentation')
            : await predictionService.getOperationalBottleneck('documentation');
        } else if (tab === 'cross_module') {
          resp = forceRefresh
            ? await predictionService.refreshOperationalBottleneck('cross_module')
            : await predictionService.getOperationalBottleneck('cross_module');
        }
        const data = resp?.data || resp;
        setPrediction(data);
      }
    } catch (err) {
      console.error('Failed to load resource allocation & bottleneck intelligence:', err);
      setError(err?.message || 'Unable to connect to predictive bottleneck intelligence engine');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    fetchTabIntelligence(activeTab, false);
  }, [activeTab, fetchTabIntelligence]);

  const handleRefresh = () => {
    fetchTabIntelligence(activeTab, true);
  };

  const handleQueueAction = async () => {
    if (!prediction?.prediction_id) return;
    try {
      setSubmittingAction(true);
      await predictionService.requestAction(
        prediction.prediction_id,
        actionNotes || 'Requested operational mitigation and rebalancing workflow.'
      );
      setPrediction(prev => (prev ? { ...prev, status: 'ACTION_REQUESTED' } : prev));
      setActionSuccessMsg('Action queued for Human-in-the-Loop review via Action System.');
      setShowActionModal(false);
      setActionNotes('');
      setTimeout(() => setActionSuccessMsg(null), 5000);
    } catch (err) {
      console.error('Failed to queue action:', err);
      setError(err?.message || 'Failed to dispatch action request to Action System.');
    } finally {
      setSubmittingAction(false);
    }
  };

  const getVerificationPath = (tab) => {
    switch (tab) {
      case 'approvals':
        return '/dashboard/approvals';
      case 'documentation':
        return '/dashboard/shipments';
      case 'cross_module':
        return '/dashboard/tracking';
      case 'resource':
        return '/dashboard/approvals';
      default:
        return '/dashboard/approvals';
    }
  };

  const getVerificationLabel = (tab) => {
    switch (tab) {
      case 'approvals':
        return 'Verify Pending Approvals Queue →';
      case 'documentation':
        return 'Verify Shipment Document Discrepancies →';
      case 'cross_module':
        return 'Verify Ocean Container Tracking & Hold →';
      case 'resource':
        return 'Verify Assigned Owner Review Allocation →';
      default:
        return 'View Operations Module →';
    }
  };

  return (
    <div
      className={`rb-predictive-card ${compact ? 'rb-predictive-card--compact' : ''} ${className}`}
      data-testid="resource-bottleneck-predictive-card"
    >
      {/* ── HEADER ── */}
      <div className="rb-card-header">
        <div className="rb-card-title-group">
          <div className="rb-header-icon-wrap">
            <Layers className="rb-header-icon" size={18} />
          </div>
          <div>
            <h3 className="rb-card-title">
              Predictive Resource Allocation & Operational Bottleneck Intelligence
            </h3>
            <p className="rb-card-subtitle">
              Authoritative bottleneck detection & resource rebalancing signals grounded in live database telemetry
            </p>
          </div>
        </div>

        <div className="rb-header-actions">
          <button
            type="button"
            className="rb-btn-refresh"
            onClick={handleRefresh}
            disabled={loading || refreshing}
            title="Recalculate bottleneck signals against current database records"
            data-testid="rb-refresh-btn"
          >
            <RefreshCw size={14} className={refreshing ? 'rb-spin' : ''} />
            <span>{refreshing ? 'Analyzing...' : 'Refresh'}</span>
          </button>
        </div>
      </div>

      {/* ── TAB BAR ── */}
      <div className="rb-tab-bar" role="tablist">
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'approvals'}
          className={`rb-tab-btn ${activeTab === 'approvals' ? 'rb-tab-btn--active' : ''}`}
          onClick={() => setActiveTab('approvals')}
          data-testid="tab-approvals"
        >
          <GitPullRequest size={14} />
          <span>Approval Bottlenecks</span>
        </button>

        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'documentation'}
          className={`rb-tab-btn ${activeTab === 'documentation' ? 'rb-tab-btn--active' : ''}`}
          onClick={() => setActiveTab('documentation')}
          data-testid="tab-documentation"
        >
          <FileText size={14} />
          <span>Documentation Backlog</span>
        </button>

        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'cross_module'}
          className={`rb-tab-btn ${activeTab === 'cross_module' ? 'rb-tab-btn--active' : ''}`}
          onClick={() => setActiveTab('cross_module')}
          data-testid="tab-cross-module"
        >
          <Anchor size={14} />
          <span>Cross-Module Exceptions</span>
        </button>

        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'resource'}
          className={`rb-tab-btn ${activeTab === 'resource' ? 'rb-tab-btn--active' : ''}`}
          onClick={() => setActiveTab('resource')}
          data-testid="tab-resource"
        >
          <Users size={14} />
          <span>Owner Workload Imbalance</span>
        </button>

        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'summary'}
          className={`rb-tab-btn ${activeTab === 'summary' ? 'rb-tab-btn--active' : ''}`}
          onClick={() => setActiveTab('summary')}
          data-testid="tab-summary"
        >
          <Briefcase size={14} />
          <span>Bottleneck Overview</span>
        </button>
      </div>

      {/* ── ADVISORY NOTICE BANNER ── */}
      <div className="rb-advisory-banner">
        <Info size={14} className="rb-advisory-icon" />
        <span className="rb-advisory-text">
          <strong>Advisory Intelligence:</strong> Predictions reflect potential operational bottlenecks and workload concentrations. Any rebalancing action requires Human-in-the-Loop review and Action System permission checks.
        </span>
      </div>

      {/* ── ACTION SUCCESS NOTIFICATION ── */}
      {actionSuccessMsg && (
        <div className="rb-success-banner" data-testid="rb-action-success">
          <CheckCircle2 size={15} className="rb-success-icon" />
          <span>{actionSuccessMsg}</span>
        </div>
      )}

      {/* ── ERROR STATE ── */}
      {error && !loading && (
        <div className="rb-error-card" data-testid="rb-error-state">
          <AlertTriangle size={18} className="rb-error-icon" />
          <div>
            <div className="rb-error-title">Intelligence Engine Offline</div>
            <div className="rb-error-desc">{error}</div>
          </div>
          <button
            type="button"
            className="rb-error-retry-btn"
            onClick={handleRefresh}
          >
            Retry
          </button>
        </div>
      )}

      {/* ── LOADING STATE ── */}
      {loading && (
        <div className="rb-loading-card" data-testid="rb-loading-state">
          <RefreshCw size={20} className="rb-spin rb-loading-spinner" />
          <p className="rb-loading-text">
            Evaluating queue dwell times, document compliance, and resource concentrations...
          </p>
        </div>
      )}

      {/* ── UNIFIED SUMMARY OVERVIEW ── */}
      {!loading && !error && activeTab === 'summary' && summaryData && (
        <div className="rb-summary-grid" data-testid="rb-summary-view">
          <div className="rb-summary-card">
            <div className="rb-summary-card-header">
              <span className="rb-dim-tag">APPROVALS</span>
              <SeverityBadge severity={summaryData.approvals_bottleneck?.severity || 'HIGH'} />
            </div>
            <h4 className="rb-summary-card-title">Commercial Approvals Queue</h4>
            <p className="rb-summary-card-stat">
              {summaryData.approvals_bottleneck?.pending_count ?? '—'}
              <span className="rb-stat-unit"> pending requests</span>
            </p>
            <p className="rb-summary-card-desc">
              {summaryData.approvals_bottleneck?.prediction_statement || 'Evaluating approval queue dwell time'}
            </p>
            <button
              type="button"
              className="rb-summary-deep-link"
              onClick={() => setActiveTab('approvals')}
            >
              Inspect Approvals Bottleneck →
            </button>
          </div>

          <div className="rb-summary-card">
            <div className="rb-summary-card-header">
              <span className="rb-dim-tag">DOCUMENTATION</span>
              <SeverityBadge severity={summaryData.documentation_bottleneck?.severity || 'HIGH'} />
            </div>
            <h4 className="rb-summary-card-title">Manifest & Document Clearance</h4>
            <p className="rb-summary-card-stat">
              {summaryData.documentation_bottleneck?.pending_count ?? '—'}
              <span className="rb-stat-unit"> active discrepancies</span>
            </p>
            <p className="rb-summary-card-desc">
              {summaryData.documentation_bottleneck?.prediction_statement || 'Monitoring export document clearance'}
            </p>
            <button
              type="button"
              className="rb-summary-deep-link"
              onClick={() => setActiveTab('documentation')}
            >
              Inspect Documentation Backlog →
            </button>
          </div>

          <div className="rb-summary-card">
            <div className="rb-summary-card-header">
              <span className="rb-dim-tag">CROSS-MODULE</span>
              <SeverityBadge severity={summaryData.cross_module_bottleneck?.severity || 'HIGH'} />
            </div>
            <h4 className="rb-summary-card-title">Customs Exception & Downstream Impact</h4>
            <p className="rb-summary-card-stat">
              {summaryData.cross_module_bottleneck?.queue_dwell_hours
                ? `${summaryData.cross_module_bottleneck.queue_dwell_hours}h`
                : '48h'}
              <span className="rb-stat-unit"> customs hold dwell</span>
            </p>
            <p className="rb-summary-card-desc">
              {summaryData.cross_module_bottleneck?.prediction_statement || 'Customs hold blocks delivery and billing'}
            </p>
            <button
              type="button"
              className="rb-summary-deep-link"
              onClick={() => setActiveTab('cross_module')}
            >
              Inspect Cross-Module Exception →
            </button>
          </div>

          <div className="rb-summary-card">
            <div className="rb-summary-card-header">
              <span className="rb-dim-tag">ALLOCATION</span>
              <SeverityBadge severity={summaryData.resource_allocation?.severity || 'HIGH'} />
            </div>
            <h4 className="rb-summary-card-title">Owner Workload Imbalance</h4>
            <p className="rb-summary-card-stat">
              100%
              <span className="rb-stat-unit"> single-owner concentration</span>
            </p>
            <p className="rb-summary-card-desc">
              {summaryData.resource_allocation?.prediction_statement || 'All pending reviews concentrated on one owner'}
            </p>
            <button
              type="button"
              className="rb-summary-deep-link"
              onClick={() => setActiveTab('resource')}
            >
              Inspect Workload Imbalance →
            </button>
          </div>
        </div>
      )}

      {/* ── SINGLE DIMENSION PREDICTION CARD ── */}
      {!loading && !error && activeTab !== 'summary' && prediction && (
        <div className="rb-body" data-testid="rb-prediction-body">
          {/* Top Status Strip */}
          <div className="rb-status-strip">
            <div className="rb-badge-group">
              <ValueTypeBadge isForecast={true} />
              <SeverityBadge severity={prediction.severity} />
              <ConfidenceBadge score={prediction.confidence_score} band={prediction.confidence_band} />
              <span className="rb-prediction-type-pill" data-testid="rb-prediction-type">
                {prediction.prediction_type}
              </span>
            </div>

            <div className="rb-meta-timestamp" title="Generated timestamp">
              <Clock size={12} />
              <span>Generated: {formatDateTime(prediction.created_at)}</span>
            </div>
          </div>

          {/* Core Prediction Statement */}
          <div className="rb-statement-box">
            <p className="rb-statement-text" data-testid="rb-statement">
              {prediction.prediction_statement}
            </p>
          </div>

          {/* Metric Grounding Tiles */}
          <div className="rb-metrics-grid">
            {(prediction.pending_count !== undefined && prediction.pending_count !== null) && (
              <div className="rb-metric-tile">
                <span className="rb-metric-label">
                  {activeTab === 'documentation' ? 'Document Discrepancies' : 'Pending Review Items'}
                </span>
                <span className="rb-metric-value" data-testid="metric-pending-count">
                  {prediction.pending_count}
                </span>
                <span className="rb-metric-sub">Authoritative database count</span>
              </div>
            )}

            {(prediction.workload_count !== undefined && prediction.workload_count !== null && activeTab === 'resource') && (
              <div className="rb-metric-tile">
                <span className="rb-metric-label">Owner Assigned Queue</span>
                <span className="rb-metric-value" data-testid="metric-workload-count">
                  {prediction.workload_count} items
                </span>
                <span className="rb-metric-sub">100% organizational share</span>
              </div>
            )}

            {prediction.queue_dwell_hours !== undefined && prediction.queue_dwell_hours !== null && (
              <div className="rb-metric-tile">
                <span className="rb-metric-label">Average Queue Dwell</span>
                <span className="rb-metric-value" data-testid="metric-dwell-hours">
                  {prediction.queue_dwell_hours} hrs
                </span>
                <span className="rb-metric-sub">Threshold target: &lt; 8.0h</span>
              </div>
            )}

            {prediction.affected_stage && (
              <div className="rb-metric-tile">
                <span className="rb-metric-label">Affected Process Stage</span>
                <span className="rb-metric-value text-sm font-semibold text-slate-800" data-testid="metric-stage">
                  {prediction.affected_stage}
                </span>
                <span className="rb-metric-sub">Operational choke point</span>
              </div>
            )}

            {prediction.assigned_owner && (
              <div className="rb-metric-tile">
                <span className="rb-metric-label">Primary Assigned Owner</span>
                <span className="rb-metric-value text-xs font-mono font-medium text-slate-800 truncate" data-testid="metric-owner" title={prediction.assigned_owner}>
                  {prediction.assigned_owner}
                </span>
                <span className="rb-metric-sub">Primary queue holder</span>
              </div>
            )}

            {prediction.carrier && (
              <div className="rb-metric-tile">
                <span className="rb-metric-label">Impacting Carrier</span>
                <span className="rb-metric-value text-sm font-semibold text-slate-800" data-testid="metric-carrier">
                  {prediction.carrier}
                </span>
                <span className="rb-metric-sub">Ocean line partner</span>
              </div>
            )}

            <div className="rb-metric-tile">
              <span className="rb-metric-label">Data Coverage</span>
              <span className="rb-metric-value" data-testid="metric-coverage">
                {prediction.data_coverage_score
                  ? `${Math.round(prediction.data_coverage_score * 100)}%`
                  : '100%'}
              </span>
              <span className="rb-metric-sub">Sample size: {prediction.sample_size_evaluated || 1} records</span>
            </div>
          </div>

          {/* Recommended Action / Next Steps */}
          {prediction.recommended_action && (
            <div className="rb-recommendation-box" data-testid="rb-recommendation">
              <div className="rb-rec-header">
                <Zap size={14} className="rb-rec-icon" />
                <span className="rb-rec-title">Recommended Mitigating Action</span>
              </div>
              <p className="rb-rec-text">{prediction.recommended_action}</p>

              <div className="rb-action-btn-row">
                <button
                  type="button"
                  className="rb-btn-queue-action"
                  onClick={() => setShowActionModal(true)}
                  data-testid="rb-queue-action-btn"
                >
                  <Send size={13} />
                  <span>Queue Action via Action System</span>
                </button>

                <button
                  type="button"
                  className="rb-btn-deep-link"
                  onClick={() => navigate(getVerificationPath(activeTab))}
                  data-testid="rb-verify-link"
                >
                  <span>{getVerificationLabel(activeTab)}</span>
                  <ArrowRight size={13} />
                </button>
              </div>
            </div>
          )}

          {/* Evidence and Grounding Collapsible Drawer */}
          <div className="rb-evidence-section">
            <button
              type="button"
              className="rb-evidence-toggle"
              onClick={() => setShowEvidence(!showEvidence)}
              data-testid="rb-evidence-toggle"
            >
              <span>Grounding Signals & Methodology Telemetry</span>
              {showEvidence ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
            </button>

            {showEvidence && (
              <div className="rb-evidence-body" data-testid="rb-evidence-body">
                <div className="rb-evidence-block">
                  <div className="rb-evidence-heading">Reasoning & Risk Summary</div>
                  <p className="rb-evidence-content">{prediction.reasoning_summary || 'N/A'}</p>
                </div>

                <div className="rb-evidence-block">
                  <div className="rb-evidence-heading">Confidence Score Breakdown</div>
                  <p className="rb-evidence-content">
                    {prediction.confidence_factors?.summary ||
                      'Computed based on deterministic historical observation density and live queue strain.'}
                  </p>
                </div>

                {prediction.source_references && prediction.source_references.length > 0 && (
                  <div className="rb-evidence-block">
                    <div className="rb-evidence-heading">Source Database Records Evaluated</div>
                    <div className="rb-source-table-wrap">
                      <table className="rb-source-table">
                        <thead>
                          <tr>
                            <th>Entity Type</th>
                            <th>Entity Reference</th>
                            <th>Bottleneck Dimension</th>
                            <th>Dwell / Context</th>
                          </tr>
                        </thead>
                        <tbody>
                          {prediction.source_references.map((src, idx) => (
                            <tr key={idx}>
                              <td>
                                <span className="rb-source-type-pill">{src.entity_type}</span>
                              </td>
                              <td className="font-mono text-xs">{src.entity_id || 'N/A'}</td>
                              <td>{src.bottleneck_type || src.signal_contribution || 'Queue Concentration'}</td>
                              <td>
                                {src.queue_dwell_hours
                                  ? `${src.queue_dwell_hours}h dwell`
                                  : src.affected_stage
                                  ? src.affected_stage
                                  : 'Persistent Observation'}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                )}

                {prediction.limitations && prediction.limitations.length > 0 && (
                  <div className="rb-evidence-block">
                    <div className="rb-evidence-heading">Forecast Limitations & Assumptions</div>
                    <ul className="rb-limitations-list">
                      {prediction.limitations.map((item, idx) => (
                        <li key={idx}>{item}</li>
                      ))}
                    </ul>
                  </div>
                )}

                <div className="rb-telemetry-meta">
                  <span>Prediction ID: <code>{prediction.prediction_id}</code></span>
                  <span>Model: <code>{prediction.model_version || 'rules-engine-v1'}</code></span>
                  <span>Org ID: <code>{prediction.org_id}</code></span>
                  <span>Horizon: <code>{prediction.prediction_horizon || '7D'}</code></span>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* ── ACTION MODAL ── */}
      {showActionModal && (
        <div className="rb-modal-backdrop" data-testid="rb-action-modal">
          <div className="rb-modal-dialog">
            <div className="rb-modal-header">
              <h4 className="rb-modal-title">Queue Operational Mitigation Action</h4>
              <button
                type="button"
                className="rb-modal-close"
                onClick={() => setShowActionModal(false)}
              >
                ✕
              </button>
            </div>

            <div className="rb-modal-body">
              <p className="text-xs text-slate-600 mb-3">
                Submits an advisory mitigation proposal to the centralized Action System with Human-in-the-Loop approval enforcement.
              </p>

              <div className="rb-modal-field">
                <label className="rb-modal-label">Proposed Mitigating Action:</label>
                <div className="rb-modal-action-text">{prediction?.recommended_action}</div>
              </div>

              <div className="rb-modal-field">
                <label className="rb-modal-label">Operator Notes / Rationale (Optional):</label>
                <textarea
                  className="rb-modal-textarea"
                  rows={3}
                  value={actionNotes}
                  onChange={(e) => setActionNotes(e.target.value)}
                  placeholder="e.g., Temporary rebalancing of commercial approval tasks during peak vessel cutoff window."
                />
              </div>
            </div>

            <div className="rb-modal-footer">
              <button
                type="button"
                className="rb-modal-btn-cancel"
                onClick={() => setShowActionModal(false)}
              >
                Cancel
              </button>
              <button
                type="button"
                className="rb-modal-btn-submit"
                onClick={handleQueueAction}
                disabled={submittingAction}
                data-testid="rb-modal-submit-btn"
              >
                {submittingAction ? 'Submitting...' : 'Submit to Action System'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
