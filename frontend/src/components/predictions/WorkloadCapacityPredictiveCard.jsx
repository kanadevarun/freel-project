import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  AlertTriangle,
  AlertCircle,
  ShieldCheck,
  RefreshCw,
  ChevronDown,
  ChevronUp,
  CheckCircle2,
  Lock,
  ArrowRight,
  Info,
  Clock,
  Radio,
  FileText,
  Activity,
  Send,
  Users,
  Compass,
  TrendingUp,
  BarChart3,
  Layers,
  Zap,
  Sliders
} from 'lucide-react';
import predictionService from '../../services/predictionService';
import { SeverityBadge, ConfidenceBadge, ValueTypeBadge } from './PredictionBadge';
import './WorkloadCapacityPredictiveCard.css';

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

export default function WorkloadCapacityPredictiveCard({
  defaultTab = 'approvals', // 'approvals' | 'documentation' | 'corridor' | 'quotes' | 'summary'
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
        const resp = await predictionService.getWorkloadCapacitySummary();
        const data = resp?.data || resp;
        setSummaryData(data);
        setPrediction(null);
      } else {
        let resp;
        if (tab === 'approvals') {
          resp = forceRefresh
            ? await predictionService.refreshWorkloadPrediction('approvals')
            : await predictionService.getWorkloadPrediction('approvals');
        } else if (tab === 'documentation') {
          resp = forceRefresh
            ? await predictionService.refreshWorkloadPrediction('documentation')
            : await predictionService.getWorkloadPrediction('documentation');
        } else if (tab === 'corridor') {
          resp = forceRefresh
            ? await predictionService.refreshCapacityPrediction('corridor')
            : await predictionService.getCapacityPrediction('corridor');
        } else if (tab === 'quotes') {
          resp = forceRefresh
            ? await predictionService.refreshDemandPrediction('commercial')
            : await predictionService.getDemandPrediction('commercial');
        }
        const data = resp?.data || resp;
        setPrediction(data);
      }
    } catch (err) {
      console.error('Failed to load predictive planning intelligence:', err);
      setError(err?.message || 'Unable to connect to predictive intelligence engine');
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
        actionNotes || 'Requested automated planning review workflow.'
      );
      setPrediction(prev => (prev ? { ...prev, status: 'ACTION_REQUESTED' } : prev));
      setActionSuccessMsg('Action queued for Human-in-the-Loop review.');
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
      case 'corridor':
        return '/dashboard/tracking';
      case 'quotes':
        return '/dashboard/rfqs';
      default:
        return '/dashboard/approvals';
    }
  };

  const getVerificationLabel = (tab) => {
    switch (tab) {
      case 'approvals':
        return 'Verify Pending Approvals Queue →';
      case 'documentation':
        return 'Verify Shipment Documentation Holds →';
      case 'corridor':
        return 'Verify Ocean Tracking Corridors →';
      case 'quotes':
        return 'Verify RFQ Pipeline Queue →';
      default:
        return 'View Operations Module →';
    }
  };

  return (
    <div
      className={`wc-predictive-card ${compact ? 'wc-predictive-card--compact' : ''} ${className}`}
      data-testid="workload-capacity-predictive-card"
    >
      {/* ── HEADER ── */}
      <div className="wc-card-header">
        <div className="wc-card-title-group">
          <div className="wc-header-icon-wrap">
            <Activity className="wc-header-icon" size={18} />
          </div>
          <div>
            <h3 className="wc-card-title">
              Predictive Demand, Capacity & Workload Intelligence
            </h3>
            <p className="wc-card-subtitle">
              Authoritative real-time predictive forecasting powered by isolated business telemetry
            </p>
          </div>
        </div>

        <div className="wc-header-actions">
          <button
            type="button"
            className="wc-btn-refresh"
            onClick={handleRefresh}
            disabled={loading || refreshing}
            title="Recalculate AI forecast against current persistent database state"
            data-testid="wc-refresh-btn"
          >
            <RefreshCw size={14} className={refreshing ? 'wc-spin' : ''} />
            <span>{refreshing ? 'Recalculating...' : 'Refresh'}</span>
          </button>
        </div>
      </div>

      {/* ── TAB BAR ── */}
      <div className="wc-tab-bar" role="tablist">
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'approvals'}
          className={`wc-tab-btn ${activeTab === 'approvals' ? 'wc-tab-btn--active' : ''}`}
          onClick={() => setActiveTab('approvals')}
          data-testid="tab-approvals"
        >
          <Users size={14} />
          <span>Approvals Workload</span>
        </button>

        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'documentation'}
          className={`wc-tab-btn ${activeTab === 'documentation' ? 'wc-tab-btn--active' : ''}`}
          onClick={() => setActiveTab('documentation')}
          data-testid="tab-documentation"
        >
          <FileText size={14} />
          <span>Documentation Backlog</span>
        </button>

        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'corridor'}
          className={`wc-tab-btn ${activeTab === 'corridor' ? 'wc-tab-btn--active' : ''}`}
          onClick={() => setActiveTab('corridor')}
          data-testid="tab-corridor"
        >
          <Compass size={14} />
          <span>Corridor Capacity</span>
        </button>

        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'quotes'}
          className={`wc-tab-btn ${activeTab === 'quotes' ? 'wc-tab-btn--active' : ''}`}
          onClick={() => setActiveTab('quotes')}
          data-testid="tab-quotes"
        >
          <TrendingUp size={14} />
          <span>Commercial Quote Demand</span>
        </button>

        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'summary'}
          className={`wc-tab-btn ${activeTab === 'summary' ? 'wc-tab-btn--active' : ''}`}
          onClick={() => setActiveTab('summary')}
          data-testid="tab-summary"
        >
          <BarChart3 size={14} />
          <span>Portfolio Summary</span>
        </button>
      </div>

      {/* ── ADVISORY NOTICE BANNER ── */}
      <div className="wc-advisory-banner">
        <Info size={14} className="wc-advisory-icon" />
        <span className="wc-advisory-text">
          <strong>Advisory Intelligence:</strong> Predictions are non-binding operational forecasts. Action dispatch enforces human-in-the-loop review and Action System permission checks.
        </span>
      </div>

      {/* ── ACTION SUCCESS NOTIFICATION ── */}
      {actionSuccessMsg && (
        <div className="wc-success-banner" data-testid="wc-action-success">
          <CheckCircle2 size={15} className="wc-success-icon" />
          <span>{actionSuccessMsg}</span>
        </div>
      )}

      {/* ── ERROR STATE ── */}
      {error && !loading && (
        <div className="wc-error-card" data-testid="wc-error-state">
          <AlertTriangle size={18} className="wc-error-icon" />
          <div>
            <div className="wc-error-title">Intelligence Feed Offline</div>
            <div className="wc-error-desc">{error}</div>
          </div>
          <button
            type="button"
            className="wc-error-retry-btn"
            onClick={handleRefresh}
          >
            Retry
          </button>
        </div>
      )}

      {/* ── LOADING STATE ── */}
      {loading && (
        <div className="wc-loading-card" data-testid="wc-loading-state">
          <RefreshCw size={20} className="wc-spin wc-loading-spinner" />
          <p className="wc-loading-text">
            Synthesizing deterministic database records & AI predictive signals...
          </p>
        </div>
      )}

      {/* ── UNIFIED SUMMARY VIEW ── */}
      {!loading && !error && activeTab === 'summary' && summaryData && (
        <div className="wc-summary-grid" data-testid="wc-summary-view">
          <div className="wc-summary-card">
            <div className="wc-summary-card-header">
              <span className="wc-dim-tag">WORKLOAD</span>
              <SeverityBadge severity={summaryData.approvals_workload?.severity || 'HIGH'} />
            </div>
            <h4 className="wc-summary-card-title">Approval Queue Pressure</h4>
            <p className="wc-summary-card-stat">
              {summaryData.approvals_workload?.pending_count ?? '—'}
              <span className="wc-stat-unit"> pending requests</span>
            </p>
            <p className="wc-summary-card-desc">
              {summaryData.approvals_workload?.prediction_statement || 'Evaluating queue dwell time'}
            </p>
            <button
              type="button"
              className="wc-summary-deep-link"
              onClick={() => setActiveTab('approvals')}
            >
              Inspect Workload Signals →
            </button>
          </div>

          <div className="wc-summary-card">
            <div className="wc-summary-card-header">
              <span className="wc-dim-tag">OPERATIONS</span>
              <SeverityBadge severity={summaryData.documentation_workload?.severity || 'HIGH'} />
            </div>
            <h4 className="wc-summary-card-title">Documentation Compliance</h4>
            <p className="wc-summary-card-stat">
              {summaryData.documentation_workload?.pending_count ?? '—'}
              <span className="wc-stat-unit"> active discrepancies</span>
            </p>
            <p className="wc-summary-card-desc">
              {summaryData.documentation_workload?.prediction_statement || 'Monitoring vessel manifests'}
            </p>
            <button
              type="button"
              className="wc-summary-deep-link"
              onClick={() => setActiveTab('documentation')}
            >
              Inspect Documentation Backlog →
            </button>
          </div>

          <div className="wc-summary-card">
            <div className="wc-summary-card-header">
              <span className="wc-dim-tag">CAPACITY</span>
              <SeverityBadge severity={summaryData.corridor_capacity?.severity || 'MEDIUM'} />
            </div>
            <h4 className="wc-summary-card-title">Trade Corridor Capacity</h4>
            <p className="wc-summary-card-stat">
              {summaryData.corridor_capacity?.utilization_rate
                ? `${(summaryData.corridor_capacity.utilization_rate > 1
                    ? summaryData.corridor_capacity.utilization_rate
                    : summaryData.corridor_capacity.utilization_rate * 100
                  ).toFixed(1)}%`
                : '87.5%'}
              <span className="wc-stat-unit"> corridor load</span>
            </p>
            <p className="wc-summary-card-desc">
              {summaryData.corridor_capacity?.prediction_statement || 'Evaluating terminal dwell'}
            </p>
            <button
              type="button"
              className="wc-summary-deep-link"
              onClick={() => setActiveTab('corridor')}
            >
              Inspect Corridor Signals →
            </button>
          </div>

          <div className="wc-summary-card">
            <div className="wc-summary-card-header">
              <span className="wc-dim-tag">DEMAND</span>
              <SeverityBadge severity={summaryData.quote_demand?.severity || 'MEDIUM'} />
            </div>
            <h4 className="wc-summary-card-title">Commercial Quote Demand</h4>
            <p className="wc-summary-card-stat">
              {summaryData.quote_demand?.pending_count ?? '—'}
              <span className="wc-stat-unit"> active RFQs</span>
            </p>
            <p className="wc-summary-card-desc">
              {summaryData.quote_demand?.prediction_statement || 'Analyzing RFQ turnaround'}
            </p>
            <button
              type="button"
              className="wc-summary-deep-link"
              onClick={() => setActiveTab('quotes')}
            >
              Inspect Commercial Demand →
            </button>
          </div>
        </div>
      )}

      {/* ── SINGLE DIMENSION PREDICTION CARD ── */}
      {!loading && !error && activeTab !== 'summary' && prediction && (
        <div className="wc-body" data-testid="wc-prediction-body">
          {/* Top Status Strip */}
          <div className="wc-status-strip">
            <div className="wc-badge-group">
              <ValueTypeBadge isForecast={true} />
              <SeverityBadge severity={prediction.severity} />
              <ConfidenceBadge score={prediction.confidence_score} band={prediction.confidence_band} />
              <span className="wc-prediction-type-pill" data-testid="wc-prediction-type">
                {prediction.prediction_type}
              </span>
            </div>

            <div className="wc-meta-timestamp" title="Generated timestamp">
              <Clock size={12} />
              <span>Forecast Generated: {formatDateTime(prediction.created_at)}</span>
            </div>
          </div>

          {/* Core Prediction Statement */}
          <div className="wc-statement-box">
            <p className="wc-statement-text" data-testid="wc-statement">
              {prediction.prediction_statement}
            </p>
          </div>

          {/* Metric Grounding Tiles */}
          <div className="wc-metrics-grid">
            {prediction.pending_count !== undefined && prediction.pending_count !== null && (
              <div className="wc-metric-tile">
                <span className="wc-metric-label">
                  {activeTab === 'quotes' ? 'Active Pipeline RFQs' : 'Pending Work Items'}
                </span>
                <span className="wc-metric-value" data-testid="metric-pending-count">
                  {prediction.pending_count}
                </span>
                <span className="wc-metric-sub">Authoritative DB count</span>
              </div>
            )}

            {prediction.utilization_rate !== undefined && prediction.utilization_rate !== null && (
              <div className="wc-metric-tile">
                <span className="wc-metric-label">Estimated Utilization</span>
                <span className="wc-metric-value" data-testid="metric-utilization">
                  {(prediction.utilization_rate > 1
                    ? prediction.utilization_rate
                    : prediction.utilization_rate * 100
                  ).toFixed(1)}%
                </span>
                <span className="wc-metric-sub">Capacity threshold: 85%</span>
              </div>
            )}

            {prediction.lane_corridor && (
              <div className="wc-metric-tile">
                <span className="wc-metric-label">Focus Trade Corridor</span>
                <span className="wc-metric-value text-base font-semibold" data-testid="metric-corridor">
                  {prediction.lane_corridor}
                </span>
                <span className="wc-metric-sub">Ocean trade lane</span>
              </div>
            )}

            {prediction.customer_reference && (
              <div className="wc-metric-tile">
                <span className="wc-metric-label">Key Account Concentration</span>
                <span className="wc-metric-value text-base font-semibold" data-testid="metric-customer">
                  {prediction.customer_reference}
                </span>
                <span className="wc-metric-sub">Pipeline driver</span>
              </div>
            )}

            <div className="wc-metric-tile">
              <span className="wc-metric-label">Data Coverage</span>
              <span className="wc-metric-value" data-testid="metric-coverage">
                {prediction.data_coverage_score
                  ? `${Math.round(prediction.data_coverage_score * 100)}%`
                  : '100%'}
              </span>
              <span className="wc-metric-sub">Sample size: {prediction.sample_size_evaluated || 1} records</span>
            </div>
          </div>

          {/* Recommended Action / Next Steps */}
          {prediction.recommended_action && (
            <div className="wc-recommendation-box" data-testid="wc-recommendation">
              <div className="wc-rec-header">
                <Zap size={14} className="wc-rec-icon" />
                <span className="wc-rec-title">Recommended Mitigating Action</span>
              </div>
              <p className="wc-rec-text">{prediction.recommended_action}</p>

              <div className="wc-action-btn-row">
                <button
                  type="button"
                  className="wc-btn-queue-action"
                  onClick={() => setShowActionModal(true)}
                  data-testid="wc-queue-action-btn"
                >
                  <Send size={13} />
                  <span>Queue Action via Action System</span>
                </button>

                <button
                  type="button"
                  className="wc-btn-deep-link"
                  onClick={() => navigate(getVerificationPath(activeTab))}
                  data-testid="wc-verify-link"
                >
                  <span>{getVerificationLabel(activeTab)}</span>
                  <ArrowRight size={13} />
                </button>
              </div>
            </div>
          )}

          {/* Evidence and Grounding Collapsible Drawer */}
          <div className="wc-evidence-section">
            <button
              type="button"
              className="wc-evidence-toggle"
              onClick={() => setShowEvidence(!showEvidence)}
              data-testid="wc-evidence-toggle"
            >
              <span>Grounding Signals & Methodology Telemetry</span>
              {showEvidence ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
            </button>

            {showEvidence && (
              <div className="wc-evidence-body" data-testid="wc-evidence-body">
                <div className="wc-evidence-block">
                  <div className="wc-evidence-heading">Reasoning & Risk Summary</div>
                  <p className="wc-evidence-content">{prediction.reasoning_summary || 'N/A'}</p>
                </div>

                <div className="wc-evidence-block">
                  <div className="wc-evidence-heading">Confidence Score Breakdown</div>
                  <p className="wc-evidence-content">
                    {prediction.confidence_factors?.summary ||
                      'Computed based on deterministic historical observation density and live queue strain.'}
                  </p>
                </div>

                {prediction.source_references && prediction.source_references.length > 0 && (
                  <div className="wc-evidence-block">
                    <div className="wc-evidence-heading">Source Database Records Evaluated</div>
                    <div className="wc-source-table-wrap">
                      <table className="wc-source-table">
                        <thead>
                          <tr>
                            <th>Entity Type</th>
                            <th>Entity Reference</th>
                            <th>Metric Signal</th>
                            <th>Status / Context</th>
                          </tr>
                        </thead>
                        <tbody>
                          {prediction.source_references.map((src, idx) => (
                            <tr key={idx}>
                              <td>
                                <span className="wc-source-type-pill">{src.entity_type}</span>
                              </td>
                              <td className="font-mono text-xs">{src.entity_id || 'N/A'}</td>
                              <td>{src.signal_contribution || 'Primary Queue Strain'}</td>
                              <td>
                                {src.workload_type
                                  ? `Type: ${src.workload_type}`
                                  : src.utilization_rate
                                  ? `Util: ${(src.utilization_rate * 100).toFixed(0)}%`
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
                  <div className="wc-evidence-block">
                    <div className="wc-evidence-heading">Forecast Limitations & Assumptions</div>
                    <ul className="wc-limitations-list">
                      {prediction.limitations.map((item, idx) => (
                        <li key={idx}>{item}</li>
                      ))}
                    </ul>
                  </div>
                )}

                <div className="wc-telemetry-meta">
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
        <div className="wc-modal-backdrop" data-testid="wc-action-modal">
          <div className="wc-modal-dialog">
            <div className="wc-modal-header">
              <h4 className="wc-modal-title">Queue Action for Operational Review</h4>
              <button
                type="button"
                className="wc-modal-close"
                onClick={() => setShowActionModal(false)}
              >
                ✕
              </button>
            </div>

            <div className="wc-modal-body">
              <p className="text-xs text-slate-600 mb-3">
                This dispatch submits an advisory recommendation to the centralized Action System with HITL human approval requirement.
              </p>

              <div className="wc-modal-field">
                <label className="wc-modal-label">Proposed Action:</label>
                <div className="wc-modal-action-text">{prediction?.recommended_action}</div>
              </div>

              <div className="wc-modal-field">
                <label className="wc-modal-label">Operator Notes / Rationale (Optional):</label>
                <textarea
                  className="wc-modal-textarea"
                  rows={3}
                  value={actionNotes}
                  onChange={(e) => setActionNotes(e.target.value)}
                  placeholder="e.g., Routing approval prioritization to commercial escalation team due to impending quarter-end cutoff."
                />
              </div>
            </div>

            <div className="wc-modal-footer">
              <button
                type="button"
                className="wc-modal-btn-cancel"
                onClick={() => setShowActionModal(false)}
              >
                Cancel
              </button>
              <button
                type="button"
                className="wc-modal-btn-submit"
                onClick={handleQueueAction}
                disabled={submittingAction}
                data-testid="wc-modal-submit-btn"
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
