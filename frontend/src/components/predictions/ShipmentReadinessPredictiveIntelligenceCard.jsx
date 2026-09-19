import React, { useState, useEffect, useCallback } from 'react';
import {
  FileCheck,
  FileText,
  AlertTriangle,
  AlertCircle,
  CheckCircle2,
  ShieldCheck,
  Clock,
  RefreshCw,
  ChevronDown,
  ChevronUp,
  Sparkles,
  Info,
  ExternalLink,
  FileSearch,
  Layers,
  Calendar,
  Check,
  ArrowRight,
  ShieldAlert,
  Ship,
  Anchor,
  FileWarning
} from 'lucide-react';
import predictionService from '../../services/predictionService';
import { SeverityBadge, ConfidenceBadge } from './PredictionBadge';
import './ShipmentReadinessPredictiveIntelligenceCard.css';

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

/**
 * Resolve display label and color class for shipment readiness prediction category
 */
function getReadinessCategoryInfo(predictionType = '', category = '') {
  const typeUpper = String(predictionType || category).toUpperCase();
  if (typeUpper.includes('DOCUMENTATION') || typeUpper.includes('DISCREPANCY')) {
    return { label: 'Documentation Discrepancy & Cutoff Risk', colorClass: 'sr-cat-badge--discrepancy' };
  }
  if (typeUpper.includes('CUSTOMS') || typeUpper.includes('HOLD') || typeUpper.includes('REGULATORY')) {
    return { label: 'Customs & Regulatory Hold Risk', colorClass: 'sr-cat-badge--customs' };
  }
  if (typeUpper.includes('CUTOFF') || typeUpper.includes('DEADLINE')) {
    return { label: 'Cutoff Miss & Manifest Deadline', colorClass: 'sr-cat-badge--cutoff' };
  }
  if (typeUpper.includes('BILLING') || typeUpper.includes('POD')) {
    return { label: 'Billing & Collection Readiness', colorClass: 'sr-cat-badge--billing' };
  }
  if (typeUpper.includes('INSUFFICIENT')) {
    return { label: 'Insufficient Data', colorClass: 'sr-cat-badge--neutral' };
  }
  return { label: 'Shipment Readiness Intelligence', colorClass: 'sr-cat-badge--readiness' };
}

export default function ShipmentReadinessPredictiveIntelligenceCard({
  shipmentId,
  shipment = null,
  className = '',
}) {
  const [prediction, setPrediction] = useState(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState(null);
  const [showEvidence, setShowEvidence] = useState(false);
  const [actionSuccessMsg, setActionSuccessMsg] = useState(null);
  const [acting, setActing] = useState(false);

  const fetchIntelligence = useCallback(async (forceRefresh = false) => {
    if (!shipmentId) return;
    if (forceRefresh) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }
    setError(null);

    try {
      let data;
      if (forceRefresh) {
        data = await predictionService.refreshShipmentPredictedReadiness(shipmentId);
      } else {
        data = await predictionService.getShipmentPredictedReadiness(shipmentId, false);
      }
      setPrediction(data?.data || data?.prediction || data);
    } catch (err) {
      console.error('Failed to load shipment readiness intelligence:', err);
      setError(err?.response?.data?.error || err.message || 'Failed to load readiness prediction');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [shipmentId]);

  useEffect(() => {
    fetchIntelligence(false);
  }, [fetchIntelligence]);

  const handleAcknowledge = async () => {
    if (!prediction?.prediction_id) return;
    setActing(true);
    try {
      await predictionService.acknowledgePrediction(prediction.prediction_id);
      setPrediction(prev => prev ? { ...prev, status: 'ACKNOWLEDGED', review_status: 'ACKNOWLEDGED' } : prev);
      setActionSuccessMsg('Prediction acknowledged successfully.');
      setTimeout(() => setActionSuccessMsg(null), 4000);
    } catch (err) {
      console.error('Failed to acknowledge prediction:', err);
      alert('Failed to acknowledge prediction: ' + (err.message || 'Server error'));
    } finally {
      setActing(false);
    }
  };

  const handleRequestAction = async () => {
    if (!prediction?.prediction_id) return;
    setActing(true);
    try {
      await predictionService.requestAction(
        prediction.prediction_id,
        `Action queued for operational review from Shipment Readiness Card: ${prediction.recommended_action || prediction.prediction_statement}`
      );
      setPrediction(prev => prev ? { ...prev, status: 'AWAITING_APPROVAL', review_status: 'ACTION_REQUESTED' } : prev);
      setActionSuccessMsg('Action queued for Human-in-the-Loop review and Action System dispatch.');
      setTimeout(() => setActionSuccessMsg(null), 5000);
    } catch (err) {
      console.error('Failed to request action:', err);
      alert('Failed to request action: ' + (err.message || 'Server error'));
    } finally {
      setActing(false);
    }
  };

  if (loading && !refreshing) {
    return (
      <div className={`shipment-readiness-card ${className}`}>
        <div className="sr-header">
          <div className="sr-header__left">
            <div className="sr-icon-pill">
              <FileCheck size={18} className="text-navy" />
            </div>
            <div>
              <div className="sr-title-row">
                <h4 className="sr-title">Predictive Documentation & Readiness Intelligence</h4>
              </div>
              <p className="sr-subtitle">Analyzing verified shipment records, cutoffs, and document package completeness...</p>
            </div>
          </div>
        </div>
        <div className="sr-loading-skeleton" style={{ padding: '24px', textAlign: 'center', color: '#64748b' }}>
          <RefreshCw size={20} className="sr-spin" style={{ margin: '0 auto 8px auto', display: 'block' }} />
          <span>Evaluating operational compliance and cutoff milestones...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`shipment-readiness-card ${className}`}>
        <div className="sr-header">
          <div className="sr-header__left">
            <div className="sr-icon-pill">
              <FileWarning size={18} className="text-amber" />
            </div>
            <div>
              <h4 className="sr-title">Predictive Documentation & Readiness Intelligence</h4>
              <p className="sr-subtitle">Error retrieving predictive readiness records.</p>
            </div>
          </div>
          <button
            type="button"
            className="sr-refresh-btn"
            onClick={() => fetchIntelligence(true)}
            disabled={refreshing}
          >
            <RefreshCw size={14} className={refreshing ? 'sr-spin' : ''} />
            <span>Retry</span>
          </button>
        </div>
        <div className="sr-alert-banner" style={{ backgroundColor: '#fef2f2', color: '#b91c1c', border: '1px solid #fecaca' }}>
          <AlertCircle size={15} />
          <span>{String(error)}</span>
        </div>
      </div>
    );
  }

  if (!prediction) {
    return null;
  }

  const catInfo = getReadinessCategoryInfo(prediction.prediction_type, prediction.prediction_category);
  const isInsufficient = prediction.insufficient_data || prediction.predicted_value === 'INSUFFICIENT_DATA';
  const signals = prediction.supporting_signals || [];
  const sources = prediction.source_references || [];

  const getSignal = (name) => {
    const s = signals.find(item => item.signal_name === name || item.signal_name?.includes(name));
    return s ? s.observed_value : null;
  };

  // Authoritative metrics
  const authoritativeStatus = shipment?.status || getSignal('shipment_status') || 'DEPARTED';
  const milestoneRef = prediction.milestone_reference || getSignal('milestone_reference') || (sources.find(s => s.milestone_reference)?.milestone_reference) || 'ARRIVAL';
  const cutoffRef = prediction.cutoff_reference || getSignal('cutoff_reference') || (sources.find(s => s.cutoff_reference)?.cutoff_reference) || 'IMPORT_MANIFEST_DEADLINE';
  const docRef = prediction.document_reference || getSignal('document_reference') || (sources.find(s => s.document_reference)?.document_reference);
  const openDiscrepancyCount = getSignal('open_discrepancy_count') || (prediction.prediction_type === 'DOCUMENTATION_DELAY_RISK' ? '1 Flag' : '0');
  const missingDocsCount = getSignal('missing_document_count') || '0';

  return (
    <div className={`shipment-readiness-card ${className}`}>
      {/* 1. Header Bar */}
      <div className="sr-header">
        <div className="sr-header__left">
          <div className="sr-icon-pill">
            <FileCheck size={18} className="text-navy" />
          </div>
          <div>
            <div className="sr-title-row">
              <h4 className="sr-title">Predictive Documentation & Readiness Intelligence</h4>
              <span className={`sr-category-badge ${catInfo.colorClass}`}>
                {catInfo.label}
              </span>
            </div>
            <p className="sr-subtitle">
              Predictive risk forecasting grounded in authoritative shipment records, documents, milestones, and cutoffs.
            </p>
          </div>
        </div>
        <div className="sr-header__right">
          <button
            type="button"
            className="sr-refresh-btn"
            onClick={() => fetchIntelligence(true)}
            disabled={refreshing}
            title="Refresh readiness forecast"
          >
            <RefreshCw size={14} className={refreshing ? 'sr-spin' : ''} />
            <span>{refreshing ? 'Refreshing...' : 'Refresh'}</span>
          </button>
        </div>
      </div>

      {/* Action confirmation toast */}
      {actionSuccessMsg && (
        <div className="sr-alert-banner sr-alert-banner--success">
          <CheckCircle2 size={15} />
          <span>{actionSuccessMsg}</span>
        </div>
      )}

      {/* Insufficient Data State */}
      {isInsufficient ? (
        <div className="sr-insufficient-state">
          <Info size={20} className="text-slate" />
          <div>
            <h5 className="sr-insufficient-title">Insufficient Shipment or Documentation Records</h5>
            <p className="sr-insufficient-desc">
              {prediction.insufficient_data_reason || prediction.explanation || 'No active milestones, vessel schedule, or document records found to evaluate operational readiness.'}
            </p>
          </div>
        </div>
      ) : (
        <>
          {/* 2. Authoritative 4-Metric Grid */}
          <div className="sr-metric-grid">
            <div className="sr-metric-tile">
              <span className="sr-metric-tile__label">AUTHORITATIVE STATUS</span>
              <div className="sr-metric-tile__val-row">
                <span className={`sr-status-pill sr-status-pill--${authoritativeStatus.toLowerCase()}`}>
                  {authoritativeStatus}
                </span>
              </div>
              <span className="sr-metric-tile__sub">Operational state</span>
            </div>

            <div className="sr-metric-tile">
              <span className="sr-metric-tile__label">NEXT MILESTONE & CUTOFF</span>
              <div className="sr-metric-tile__val-row">
                <Clock size={16} className="text-slate" />
                <span className="sr-metric-tile__val text-truncate" title={milestoneRef}>
                  {milestoneRef}
                </span>
              </div>
              <span className="sr-metric-tile__sub text-truncate" title={cutoffRef}>
                Cutoff: {cutoffRef}
              </span>
            </div>

            <div className="sr-metric-tile">
              <span className="sr-metric-tile__label">DOCUMENT READINESS</span>
              <div className="sr-metric-tile__val-row">
                {String(openDiscrepancyCount) !== '0' && String(openDiscrepancyCount) !== '0 Flag' ? (
                  <FileWarning size={16} className="text-amber" />
                ) : (
                  <CheckCircle2 size={16} className="text-emerald" />
                )}
                <span className="sr-metric-tile__val">
                  {String(openDiscrepancyCount) !== '0' && String(openDiscrepancyCount) !== '0 Flag' ? `${openDiscrepancyCount}` : 'Documents Matched'}
                </span>
              </div>
              <span className="sr-metric-tile__sub">
                {String(missingDocsCount) !== '0' ? `${missingDocsCount} Missing Doc(s)` : 'Required docs on file'}
              </span>
            </div>

            <div className="sr-metric-tile">
              <span className="sr-metric-tile__label">COMPLIANCE POSTURE</span>
              <div className="sr-metric-tile__val-row">
                {authoritativeStatus === 'CUSTOMS_HOLD' ? (
                  <ShieldAlert size={16} className="text-rose" />
                ) : (
                  <ShieldCheck size={16} className="text-emerald" />
                )}
                <span className="sr-metric-tile__val text-truncate">
                  {authoritativeStatus === 'CUSTOMS_HOLD' ? 'Hold Encountered' : 'Standard Compliant'}
                </span>
              </div>
              <span className="sr-metric-tile__sub">Regulatory clearance</span>
            </div>
          </div>

          {/* 3. Prediction Statement & Impact Callout */}
          <div className="sr-prediction-callout">
            <div className="sr-callout-top">
              <div className="sr-callout-badges">
                <SeverityBadge severity={prediction.severity} />
                <ConfidenceBadge
                  score={prediction.confidence_score}
                  band={prediction.confidence_band}
                />
              </div>
              <span className="sr-advisory-tag">
                <Info size={13} /> Advisory Forecast
              </span>
            </div>

            <p className="sr-statement">
              {prediction.prediction_statement}
            </p>

            {prediction.explanation && (
              <div className="sr-explanation-box">
                <p className="sr-explanation-text">
                  <strong>Supporting Operational Context: </strong>
                  {prediction.explanation}
                </p>
              </div>
            )}

            {/* Grounding Signals Strip */}
            <div className="sr-grounding-strip">
              <span className="sr-grounding-label">SOURCE GROUNDING:</span>
              {milestoneRef && (
                <span className="sr-grounding-badge sr-grounding-badge--milestone">
                  <Anchor size={12} /> Milestone: {milestoneRef}
                </span>
              )}
              {cutoffRef && (
                <span className="sr-grounding-badge sr-grounding-badge--cutoff">
                  <Clock size={12} /> Cutoff: {cutoffRef}
                </span>
              )}
              {docRef && (
                <span className="sr-grounding-badge sr-grounding-badge--discrepancy">
                  <FileText size={12} /> Doc: {docRef}
                </span>
              )}
            </div>
          </div>

          {/* 4. Action Governance & Human Review Bar */}
          <div className="sr-action-bar">
            <div className="sr-action-left">
              <div className="sr-governance-note">
                <ShieldCheck size={14} className="text-emerald" />
                <span>
                  <strong>Advisory Only:</strong> Operational changes, milestone adjustments, and customs resubmissions require authorized human review.
                </span>
              </div>
            </div>

            <div className="sr-action-right">
              {prediction.review_status !== 'ACKNOWLEDGED' && prediction.review_status !== 'ACTION_REQUESTED' && (
                <button
                  type="button"
                  className="sr-btn sr-btn--secondary"
                  onClick={handleAcknowledge}
                  disabled={acting}
                >
                  <Check size={14} /> Acknowledge
                </button>
              )}

              {prediction.is_action_required && prediction.review_status !== 'ACTION_REQUESTED' && (
                <button
                  type="button"
                  className="sr-btn sr-btn--primary"
                  onClick={handleRequestAction}
                  disabled={acting}
                >
                  <ArrowRight size={14} />
                  <span>{prediction.recommended_action ? 'Queue Recommended Action' : 'Request Operational Review'}</span>
                </button>
              )}

              {prediction.review_status === 'ACTION_REQUESTED' && (
                <span className="sr-status-pill sr-status-pill--action-requested">
                  Action In Review
                </span>
              )}
            </div>
          </div>

          {/* 5. Collapsible Telemetry & Grounding Evidence */}
          <div className="sr-evidence-accordion">
            <button
              type="button"
              className="sr-evidence-toggle"
              onClick={() => setShowEvidence(prev => !prev)}
            >
              <span>Operational Grounding & Audit Signals ({signals.length + sources.length})</span>
              {showEvidence ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
            </button>

            {showEvidence && (
              <div className="sr-evidence-body">
                {signals.length > 0 && (
                  <div className="sr-evidence-section">
                    <h6 className="sr-evidence-title">Quantitative Operational Signals</h6>
                    <div className="sr-signals-table">
                      {signals.map((sig, idx) => (
                        <div key={idx} className="sr-signal-row">
                          <span className="sr-signal-name">{sig.signal_name}</span>
                          <span className="sr-signal-val">{String(sig.observed_value)}</span>
                          <span className="sr-signal-baseline">
                            Baseline: {sig.baseline_value ? String(sig.baseline_value) : 'N/A'}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {sources.length > 0 && (
                  <div className="sr-evidence-section">
                    <h6 className="sr-evidence-title">Authoritative Source Audit Trails</h6>
                    <ul className="sr-sources-list">
                      {sources.map((src, idx) => (
                        <li key={idx} className="sr-source-item">
                          <span className="sr-source-field">{src.source_field || src.source_module}</span>
                          <span className="sr-source-id">Record #{src.source_record_id}</span>
                          <span className="sr-source-time">{formatDateTime(src.source_timestamp)}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                <div className="sr-metadata-footer">
                  <span>Prediction ID: <code>{prediction.prediction_id}</code></span>
                  <span>Model: <code>{prediction.model_version || 'ai-sidecar-shipments-v1'}</code></span>
                  <span>Forecast Horizon: <code>{prediction.time_horizon || '14_DAYS'}</code></span>
                </div>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}
