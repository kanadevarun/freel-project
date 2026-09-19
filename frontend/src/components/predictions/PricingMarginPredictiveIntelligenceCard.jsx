import React, { useState, useEffect, useCallback } from 'react';
import {
  AlertTriangle,
  AlertCircle,
  ShieldCheck,
  RefreshCw,
  ChevronDown,
  ChevronUp,
  CheckCircle2,
  TrendingDown,
  TrendingUp,
  DollarSign,
  Calendar,
  Layers,
  Sparkles,
  Info,
  Clock,
  ArrowRight,
  Send,
  SlidersHorizontal,
  Check
} from 'lucide-react';
import predictionService from '../../services/predictionService';
import { SeverityBadge, ConfidenceBadge } from './PredictionBadge';
import './PricingMarginPredictiveIntelligenceCard.css';

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
 * Resolve display label and color class for pricing prediction category
 */
function getPricingCategoryInfo(category = '', predictionType = '') {
  const cat = String(category || predictionType).toUpperCase();
  if (cat.includes('COMPETITIVENESS')) {
    return { label: 'Quotation Competitiveness', colorClass: 'pm-cat-badge--competitive' };
  }
  if (cat.includes('MARGIN_RISK')) {
    return { label: 'RFQ Margin Risk Alert', colorClass: 'pm-cat-badge--risk' };
  }
  if (cat.includes('COST_VARIANCE')) {
    return { label: 'Cost Variance & Leakage', colorClass: 'pm-cat-badge--variance' };
  }
  if (cat.includes('CONTRACT') || cat.includes('RATE_PRESSURE')) {
    return { label: 'Contract Rate Pressure', colorClass: 'pm-cat-badge--contract' };
  }
  if (cat.includes('INSUFFICIENT')) {
    return { label: 'Insufficient Data', colorClass: 'pm-cat-badge--neutral' };
  }
  return { label: 'Pricing Intelligence', colorClass: 'pm-cat-badge--neutral' };
}

export default function PricingMarginPredictiveIntelligenceCard({
  recordType = 'rfq', // 'rfq' | 'contract'
  recordId,
  className = '',
}) {
  const [prediction, setPrediction] = useState(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState(null);
  const [showEvidence, setShowEvidence] = useState(false);
  const [actionSuccessMsg, setActionSuccessMsg] = useState(null);
  const [acting, setActing] = useState(false);

  const isRFQ = recordType === 'rfq';

  const fetchIntelligence = useCallback(async (forceRefresh = false) => {
    if (!recordId) return;
    if (forceRefresh) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }
    setError(null);

    try {
      let data;
      if (isRFQ) {
        data = await predictionService.getRFQPredictedMargin(recordId, forceRefresh);
      } else {
        data = await predictionService.getContractPredictedRatePressure(recordId, forceRefresh);
      }
      const pred = data?.data || data?.prediction || data;
      setPrediction(pred);
    } catch (err) {
      console.error(`Failed to load predictive pricing intelligence for ${recordType}:`, err);
      setError(err.response?.data?.message || err.message || `Failed to load predictive pricing intelligence for ${recordType}`);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [recordType, recordId, isRFQ]);

  useEffect(() => {
    fetchIntelligence(false);
  }, [fetchIntelligence]);

  const handleAcknowledge = async () => {
    if (!prediction?.prediction_id) return;
    try {
      setActing(true);
      await predictionService.acknowledgePrediction(prediction.prediction_id);
      setPrediction(prev => ({ ...prev, status: 'ACKNOWLEDGED' }));
      setActionSuccessMsg('Pricing risk insight acknowledged.');
      setTimeout(() => setActionSuccessMsg(null), 4000);
    } catch (err) {
      console.error('Failed to acknowledge prediction:', err);
      setError(err.response?.data?.message || 'Failed to acknowledge');
    } finally {
      setActing(false);
    }
  };

  const handleRequestAction = async () => {
    if (!prediction?.prediction_id) return;
    try {
      setActing(true);
      await predictionService.requestAction(
        prediction.prediction_id,
        `Pricing desk initiated mitigation for ${prediction.prediction_type} on ${recordType.toUpperCase()} #${recordId}`
      );
      setPrediction(prev => ({ ...prev, status: 'ACTION_REQUESTED' }));
      setActionSuccessMsg('Commercial action queued for Human-In-The-Loop approval.');
      setTimeout(() => setActionSuccessMsg(null), 4500);
    } catch (err) {
      console.error('Failed to request mitigation action:', err);
      setError(err.response?.data?.message || 'Failed to request action');
    } finally {
      setActing(false);
    }
  };

  const categoryInfo = getPricingCategoryInfo(prediction?.prediction_category, prediction?.prediction_type);

  if (loading) {
    return (
      <div className={`pm-pred-card pm-pred-card--loading ${className}`}>
        <div className="pm-pred-card__loading-content">
          <RefreshCw className="pm-pred-card__spinner animate-spin text-indigo-600" size={24} />
          <span className="pm-pred-card__loading-text">
            Evaluating real-time {isRFQ ? 'margin risk & quotation competitiveness' : 'tariff expiry & rate pressure'}...
          </span>
        </div>
      </div>
    );
  }

  if (error && !prediction) {
    return (
      <div className={`pm-pred-card pm-pred-card--error ${className}`}>
        <div className="pm-pred-card__error-content">
          <AlertCircle className="text-red-500 flex-shrink-0" size={20} />
          <div className="pm-pred-card__error-body">
            <h4 className="pm-pred-card__error-title">Pricing Intelligence Unavailable</h4>
            <p className="pm-pred-card__error-desc">{error}</p>
          </div>
          <button
            onClick={() => fetchIntelligence(true)}
            className="pm-pred-card__btn pm-pred-card__btn--secondary"
          >
            Retry Analysis
          </button>
        </div>
      </div>
    );
  }

  if (!prediction) {
    return (
      <div className={`pm-pred-card pm-pred-card--empty ${className}`}>
        <div className="pm-pred-card__empty-content">
          <Info className="text-slate-400 flex-shrink-0" size={18} />
          <p className="pm-pred-card__empty-text">
            No predictive signals currently detected for this {recordType.toUpperCase()}.
          </p>
        </div>
      </div>
    );
  }

  const isAcknowledged = prediction.status === 'ACKNOWLEDGED';
  const isActionRequested = prediction.status === 'ACTION_REQUESTED';
  const severity = prediction.severity || 'LOW';
  const confidenceScore = prediction.confidence_score !== undefined ? Math.round(prediction.confidence_score * 100) : null;
  const supportingSignals = prediction.supporting_signals || [];
  const sourceRefs = prediction.source_references || [];
  const isInsufficient = Boolean(prediction.insufficient_data);

  return (
    <div className={`pm-pred-card pm-pred-card--severity-${severity.toLowerCase()} ${className}`}>
      {/* Card Header */}
      <div className="pm-pred-card__header">
        <div className="pm-pred-card__title-group">
          <div className="pm-pred-card__icon-wrapper">
            {isRFQ ? <DollarSign size={18} /> : <Calendar size={18} />}
          </div>
          <div>
            <div className="flex items-center gap-2 flex-wrap">
              <h3 className="pm-pred-card__title">
                {isRFQ ? 'Predictive Pricing & Margin Intelligence' : 'Predictive Contract & Tariff Intelligence'}
              </h3>
              <span className={`pm-cat-badge ${categoryInfo.colorClass}`}>
                {categoryInfo.label}
              </span>
              {prediction.status === 'SUPERSEDED' && (
                <span className="pm-pred-card__status-pill pm-pred-card__status-pill--superseded">
                  Superseded
                </span>
              )}
            </div>
            <p className="pm-pred-card__meta">
              Authoritative financial base: MariaDB &middot; Python AI Evaluation as of {formatDateTime(prediction.source_timestamp)}
            </p>
          </div>
        </div>

        <div className="pm-pred-card__actions-header">
          <button
            onClick={() => fetchIntelligence(true)}
            disabled={refreshing}
            className="pm-pred-card__refresh-btn"
            title="Recalculate predictive signals from real database telemetry"
          >
            <RefreshCw size={14} className={refreshing ? 'animate-spin text-indigo-600' : 'text-slate-500'} />
            <span>{refreshing ? 'Recalculating...' : 'Recalculate'}</span>
          </button>
        </div>
      </div>

      {/* Success Notification */}
      {actionSuccessMsg && (
        <div className="pm-pred-card__alert pm-pred-card__alert--success">
          <CheckCircle2 size={16} className="text-emerald-600 flex-shrink-0" />
          <span>{actionSuccessMsg}</span>
        </div>
      )}

      {/* Main Statement & Badges */}
      <div className="pm-pred-card__body">
        <div className="pm-pred-card__statement-row">
          <div className="pm-pred-card__statement-text">
            {prediction.prediction_statement}
          </div>
          <div className="pm-pred-card__badges-group">
            <SeverityBadge severity={severity} />
            {confidenceScore !== null && (
              <ConfidenceBadge score={prediction.confidence_score} band={prediction.confidence_band} />
            )}
          </div>
        </div>

        {/* Insufficient Data Callout */}
        {isInsufficient && (
          <div className="pm-pred-card__insufficient-callout">
            <Info size={16} className="text-amber-600 flex-shrink-0" />
            <div>
              <strong className="text-amber-900 block text-xs">Insufficient Pricing Telemetry</strong>
              <p className="text-amber-800 text-xs mt-0.5">
                {prediction.explanation || 'No carrier quotes or contracted buying rates exist yet in the database. Safe baseline maintained without fabricating predictive margins.'}
              </p>
            </div>
          </div>
        )}

        {/* Authoritative vs Predicted Key Metric Pillars */}
        {!isInsufficient && (
          <div className="pm-pred-card__metrics-grid">
            {supportingSignals.slice(0, 4).map((sig, idx) => (
              <div key={idx} className="pm-metric-tile">
                <span className="pm-metric-tile__label">
                  {sig.signal_name.replace(/_/g, ' ').toUpperCase()}
                </span>
                <span className="pm-metric-tile__val">
                  {String(sig.observed_value)}
                </span>
                {sig.baseline_value && (
                  <span className="pm-metric-tile__sub">
                    Benchmark: {String(sig.baseline_value)}
                  </span>
                )}
              </div>
            ))}
          </div>
        )}

        {/* Detailed Explanation */}
        {!isInsufficient && (
          <div className="pm-pred-card__explanation-box">
            <div className="pm-pred-card__explanation-label">Grounded Intelligence Assessment:</div>
            <p className="pm-pred-card__explanation-text">{prediction.explanation}</p>
          </div>
        )}

        {/* Recommended HITL Action & Governance */}
        {prediction.recommended_action && (
          <div className="pm-pred-card__action-box">
            <div className="pm-pred-card__action-header">
              <div className="flex items-center gap-2">
                <Sparkles size={16} className="text-indigo-600" />
                <span className="pm-pred-card__action-title">Recommended Commercial Action:</span>
              </div>
              {prediction.requires_approval && (
                <span className="pm-pred-card__approval-badge">Requires Human Approval</span>
              )}
            </div>
            <p className="pm-pred-card__action-desc">{prediction.recommended_action}</p>

            <div className="pm-pred-card__action-buttons">
              {!isAcknowledged && !isActionRequested && (
                <button
                  onClick={handleAcknowledge}
                  disabled={acting}
                  className="pm-pred-card__btn pm-pred-card__btn--secondary"
                >
                  <Check size={14} />
                  <span>Acknowledge Insight</span>
                </button>
              )}

              {prediction.is_action_required && !isActionRequested && (
                <button
                  onClick={handleRequestAction}
                  disabled={acting}
                  className="pm-pred-card__btn pm-pred-card__btn--primary"
                >
                  <Send size={14} />
                  <span>{acting ? 'Queuing...' : 'Queue Commercial Review'}</span>
                </button>
              )}

              {isAcknowledged && (
                <span className="pm-pred-card__status-tag pm-pred-card__status-tag--ack">
                  <CheckCircle2 size={13} /> Acknowledged by Commercial Team
                </span>
              )}

              {isActionRequested && (
                <span className="pm-pred-card__status-tag pm-pred-card__status-tag--queued">
                  <Clock size={13} /> Action Queued in HITL Approvals
                </span>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Expandable Evidence & Telemetry Drawer */}
      <div className="pm-pred-card__footer">
        <button
          onClick={() => setShowEvidence(!showEvidence)}
          className="pm-pred-card__toggle-evidence-btn"
        >
          <span>Evidence Grounding & Telemetry ({sourceRefs.length} Sources, {supportingSignals.length} Signals)</span>
          {showEvidence ? <ChevronUp size={15} /> : <ChevronDown size={15} />}
        </button>

        {showEvidence && (
          <div className="pm-pred-card__evidence-panel">
            {/* Supporting Quantitative Signals */}
            {supportingSignals.length > 0 && (
              <div className="pm-evidence-section">
                <h5 className="pm-evidence-title">Quantitative Decision Drivers:</h5>
                <div className="pm-signals-table">
                  <div className="pm-signals-head">
                    <span>Driver Signal</span>
                    <span>Observed Value</span>
                    <span>Baseline / SLA</span>
                    <span>Model Weight</span>
                  </div>
                  {supportingSignals.map((sig, idx) => (
                    <div key={idx} className="pm-signals-row">
                      <span className="font-mono text-slate-700">{sig.signal_name}</span>
                      <span className="font-semibold text-slate-900">{String(sig.observed_value)}</span>
                      <span className="text-slate-500">{String(sig.baseline_value || '—')}</span>
                      <span className="text-slate-500">{(sig.importance_weight * 100).toFixed(0)}%</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Authoritative Database Source References */}
            {sourceRefs.length > 0 && (
              <div className="pm-evidence-section mt-3">
                <h5 className="pm-evidence-title">Persistent Database Lineage:</h5>
                <div className="pm-sources-grid">
                  {sourceRefs.map((src, idx) => (
                    <div key={idx} className="pm-source-item">
                      <span className="pm-source-field">{src.source_field}</span>
                      <span className="pm-source-meta">
                        Table: {src.source_module} &middot; Record ID: {src.source_record_id}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Verification Metadata */}
            <div className="pm-evidence-meta mt-3">
              <span>Prediction ID: <code className="font-mono text-xs">{prediction.prediction_id}</code></span>
              <span>Model: <code className="font-mono text-xs">{prediction.model_version || 'v4.5-pricing-rules-llm'}</code></span>
              <span>Time Horizon: <code className="font-mono text-xs">{prediction.time_horizon || '7_DAYS'}</code></span>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
