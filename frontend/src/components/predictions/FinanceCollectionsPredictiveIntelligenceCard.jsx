import React, { useState, useEffect, useCallback } from 'react';
import {
  AlertTriangle,
  AlertCircle,
  ShieldCheck,
  RefreshCw,
  ChevronDown,
  ChevronUp,
  CheckCircle2,
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
import './FinanceCollectionsPredictiveIntelligenceCard.css';

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
 * Resolve display label and color class for finance & collections category
 */
function getFinanceCategoryInfo(category = '', predictionType = '') {
  const cat = String(category || predictionType).toUpperCase();
  if (cat.includes('COLLECTION_PRIORITY')) {
    return { label: 'Collection Priority Escalation', colorClass: 'fc-cat-badge--priority' };
  }
  if (cat.includes('LATE_PAYMENT') || cat.includes('OVERDUE')) {
    return { label: 'Late-Payment Risk Alert', colorClass: 'fc-cat-badge--risk' };
  }
  if (cat.includes('CASH_INFLOW') || cat.includes('INFLOW')) {
    return { label: 'Cash Inflow Timing Window', colorClass: 'fc-cat-badge--inflow' };
  }
  if (cat.includes('DISPUTE')) {
    return { label: 'Dispute & Billing Exception', colorClass: 'fc-cat-badge--dispute' };
  }
  if (cat.includes('PAYMENT_BEHAVIOR') || cat.includes('BEHAVIOR')) {
    return { label: 'Debtor Settlement Behavior', colorClass: 'fc-cat-badge--behavior' };
  }
  if (cat.includes('CONCENTRATION')) {
    return { label: 'Receivables Exposure Risk', colorClass: 'fc-cat-badge--concentration' };
  }
  if (cat.includes('INSUFFICIENT')) {
    return { label: 'Insufficient Financial Data', colorClass: 'fc-cat-badge--neutral' };
  }
  return { label: 'Financial & Collections Intelligence', colorClass: 'fc-cat-badge--neutral' };
}

export default function FinanceCollectionsPredictiveIntelligenceCard({
  invoiceId,
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
    if (!invoiceId) return;
    if (forceRefresh) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }
    setError(null);

    try {
      let data;
      if (forceRefresh) {
        data = await predictionService.refreshInvoicePredictedCollection(invoiceId);
      } else {
        data = await predictionService.getInvoicePredictedCollection(invoiceId, false);
      }
      setPrediction(data?.data || data);
    } catch (err) {
      console.error('Failed to load predictive finance intelligence:', err);
      setError(err?.response?.data?.error || err.message || 'Failed to load collections prediction');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [invoiceId]);

  useEffect(() => {
    fetchIntelligence(false);
  }, [fetchIntelligence]);

  const handleAcknowledge = async () => {
    if (!prediction?.prediction_id) return;
    setActing(true);
    try {
      await predictionService.acknowledgePrediction(prediction.prediction_id);
      setPrediction(prev => ({
        ...prev,
        status: 'ACKNOWLEDGED',
        review_status: 'ACKNOWLEDGED'
      }));
      setActionSuccessMsg('Prediction acknowledged by finance team.');
      setTimeout(() => setActionSuccessMsg(null), 4000);
    } catch (err) {
      console.error('Failed to acknowledge prediction:', err);
      setError('Could not acknowledge prediction.');
    } finally {
      setActing(false);
    }
  };

  const handleRequestAction = async () => {
    if (!prediction?.prediction_id) return;
    setActing(true);
    try {
      const notes = `Operator authorized collections workflow: ${prediction.recommended_action || 'Review receivables'}`;
      await predictionService.requestAction(prediction.prediction_id, notes);
      setPrediction(prev => ({
        ...prev,
        status: prev.requires_approval ? 'AWAITING_APPROVAL' : 'ACTION_REQUESTED',
        review_status: 'ACTION_TAKEN'
      }));
      setActionSuccessMsg(
        prediction.requires_approval
          ? 'Action routed to Financial Controller for HITL approval.'
          : 'Action request registered in collections queue.'
      );
      setTimeout(() => setActionSuccessMsg(null), 4500);
    } catch (err) {
      console.error('Failed to request collections action:', err);
      setError('Could not request action.');
    } finally {
      setActing(false);
    }
  };

  // Helper to find supporting signal value
  const getSignalValue = (name) => {
    if (!prediction?.supporting_signals) return null;
    const sig = prediction.supporting_signals.find(s => s.signal_name === name);
    return sig ? sig.observed_value : null;
  };

  const totalAmountSignal = getSignalValue('total_amount');
  const paidAmountSignal = getSignalValue('paid_amount');
  const balanceDueSignal = getSignalValue('balance_due');
  const daysUntilDueSignal = getSignalValue('days_until_due');
  const agingBucketSignal = getSignalValue('aging_bucket');
  const isOverdueSignal = getSignalValue('is_overdue');

  if (loading) {
    return (
      <div className={`fc-pred-card fc-pred-card--loading ${className}`}>
        <div className="fc-pred-card__loading-content">
          <RefreshCw className="fc-pred-card__spinner animate-spin" size={20} />
          <span className="fc-pred-card__loading-text">
            Evaluating invoice aging, settlement history, and cash inflow timing...
          </span>
        </div>
      </div>
    );
  }

  if (error && !prediction) {
    return (
      <div className={`fc-pred-card fc-pred-card--error ${className}`}>
        <div className="fc-pred-card__error-content">
          <AlertCircle size={20} className="text-rose-500 shrink-0" />
          <div className="fc-pred-card__error-body">
            <h4 className="fc-pred-card__error-title">Finance Intelligence Unavailable</h4>
            <p className="fc-pred-card__error-desc">{error}</p>
          </div>
          <button
            type="button"
            className="fc-pred-card__btn-refresh-icon"
            onClick={() => fetchIntelligence(true)}
            title="Retry"
          >
            <RefreshCw size={15} />
          </button>
        </div>
      </div>
    );
  }

  if (!prediction) {
    return (
      <div className={`fc-pred-card fc-pred-card--empty ${className}`}>
        <div className="fc-pred-card__empty-content">
          <Info size={18} className="text-slate-400 shrink-0" />
          <span className="fc-pred-card__empty-text">
            No predictive collections forecast available for this invoice.
          </span>
        </div>
      </div>
    );
  }

  const categoryInfo = getFinanceCategoryInfo(prediction.prediction_category, prediction.prediction_type);
  const severity = (prediction.severity || 'LOW').toUpperCase();
  const severityClass = `fc-pred-card--severity-${severity.toLowerCase()}`;
  const isInsufficient = Boolean(prediction.insufficient_data);

  return (
    <div className={`fc-pred-card ${severityClass} ${className}`}>
      {/* ── CARD HEADER ── */}
      <div className="fc-pred-card__header">
        <div className="fc-pred-card__title-group">
          <div className="fc-pred-card__icon-wrap">
            <Sparkles size={18} className="fc-pred-card__icon text-indigo-600" />
          </div>
          <div>
            <div className="fc-pred-card__top-meta">
              <span className={`fc-cat-badge ${categoryInfo.colorClass}`}>
                {categoryInfo.label}
              </span>
              <SeverityBadge severity={prediction.severity} />
              <ConfidenceBadge
                score={prediction.confidence_score}
                band={prediction.confidence_band}
              />
              {prediction.status && (
                <span className={`fc-status-pill fc-status-pill--${prediction.status.toLowerCase()}`}>
                  {prediction.status.replace(/_/g, ' ')}
                </span>
              )}
            </div>
            <h3 className="fc-pred-card__title">
              Predictive Cash Flow & Collections Intelligence
            </h3>
          </div>
        </div>

        <div className="fc-pred-card__header-actions">
          <button
            type="button"
            className={`fc-pred-card__btn-refresh ${refreshing ? 'is-refreshing' : ''}`}
            onClick={() => fetchIntelligence(true)}
            disabled={refreshing}
            title="Recalculate with latest payment records"
          >
            <RefreshCw size={14} className={refreshing ? 'animate-spin' : ''} />
            <span>{refreshing ? 'Evaluating...' : 'Refresh Forecast'}</span>
          </button>
        </div>
      </div>

      {/* ── SUCCESS FEEDBACK BANNER ── */}
      {actionSuccessMsg && (
        <div className="fc-pred-card__toast">
          <CheckCircle2 size={16} className="text-emerald-600 shrink-0" />
          <span>{actionSuccessMsg}</span>
        </div>
      )}

      {/* ── DETERMINISTIC FINANCIAL METRICS STRIP (Authoritative Go facts) ── */}
      <div className="fc-metrics-grid">
        <div className="fc-metric-card">
          <span className="fc-metric-card__label">TOTAL INVOICE</span>
          <div className="fc-metric-card__value">
            <DollarSign size={15} className="text-slate-400 -mr-1" />
            <span>{totalAmountSignal != null ? String(totalAmountSignal) : '—'}</span>
          </div>
          <span className="fc-metric-card__hint">Authoritative Ledger</span>
        </div>

        <div className="fc-metric-card">
          <span className="fc-metric-card__label">AMOUNT PAID</span>
          <div className="fc-metric-card__value text-emerald-700">
            <Check size={15} className="text-emerald-500 mr-0.5" />
            <span>{paidAmountSignal != null ? String(paidAmountSignal) : '$0.00'}</span>
          </div>
          <span className="fc-metric-card__hint">Realized Cash Inflow</span>
        </div>

        <div className="fc-metric-card">
          <span className="fc-metric-card__label">BALANCE DUE</span>
          <div className={`fc-metric-card__value ${parseFloat(String(balanceDueSignal).replace(/[^0-9.-]+/g,"")) > 0 ? 'text-amber-700' : 'text-emerald-700'}`}>
            <span>{balanceDueSignal != null ? String(balanceDueSignal) : '—'}</span>
          </div>
          <span className="fc-metric-card__hint">Outstanding Receivables</span>
        </div>

        <div className="fc-metric-card">
          <span className="fc-metric-card__label">AGING / DUE STATUS</span>
          <div className="fc-metric-card__value">
            <Calendar size={14} className="text-slate-400 mr-1" />
            <span className="font-semibold text-slate-800">
              {isOverdueSignal ? 'OVERDUE' : (daysUntilDueSignal != null ? `${daysUntilDueSignal} Days Left` : (agingBucketSignal || 'Active'))}
            </span>
          </div>
          <span className="fc-metric-card__hint">
            {agingBucketSignal ? `Bucket: ${agingBucketSignal}` : 'Payment Terms Active'}
          </span>
        </div>
      </div>

      {/* ── PREDICTION STATEMENT & EXPLANATION ── */}
      <div className="fc-pred-card__statement-box">
        <div className="fc-pred-card__statement-header">
          <h4 className="fc-pred-card__statement-text">
            {prediction.prediction_statement}
          </h4>
          {prediction.predicted_value && (
            <span className="fc-pred-card__predicted-pill">
              {prediction.predicted_value}
            </span>
          )}
        </div>

        {prediction.explanation && (
          <p className="fc-pred-card__explanation">
            {prediction.explanation}
          </p>
        )}
      </div>

      {/* ── ACTION SYSTEM & HITL GOVERNANCE ── */}
      {(prediction.recommended_action || prediction.is_action_required) && (
        <div className="fc-action-box">
          <div className="fc-action-box__header">
            <div className="fc-action-box__badge">
              <ShieldCheck size={14} />
              <span>Recommended Financial Governance</span>
            </div>
            {prediction.requires_approval && (
              <span className="fc-action-box__approval-pill">
                Approval Required
              </span>
            )}
          </div>

          <p className="fc-action-box__recommendation">
            {prediction.recommended_action}
          </p>

          <div className="fc-action-box__footer">
            <div className="fc-action-box__meta">
              {prediction.action_type && (
                <span className="fc-action-box__action-type">
                  Workflow: <code>{prediction.action_type}</code>
                </span>
              )}
            </div>

            <div className="fc-action-box__buttons">
              {prediction.status === 'PUBLISHED' && (
                <button
                  type="button"
                  className="fc-btn fc-btn--secondary"
                  onClick={handleAcknowledge}
                  disabled={acting}
                >
                  <Check size={14} />
                  <span>Acknowledge</span>
                </button>
              )}

              {prediction.status !== 'ACTION_REQUESTED' && prediction.status !== 'AWAITING_APPROVAL' && (
                <button
                  type="button"
                  className="fc-btn fc-btn--primary"
                  onClick={handleRequestAction}
                  disabled={acting}
                >
                  <Send size={14} />
                  <span>
                    {prediction.requires_approval ? 'Queue for Approval' : 'Authorize Action'}
                  </span>
                </button>
              )}

              {(prediction.status === 'ACTION_REQUESTED' || prediction.status === 'AWAITING_APPROVAL') && (
                <span className="fc-action-box__applied-notice">
                  <CheckCircle2 size={14} className="text-emerald-600" />
                  <span>Action Queued in Action System</span>
                </span>
              )}
            </div>
          </div>
        </div>
      )}

      {/* ── GROUNDING & EVIDENCE DRAWER ── */}
      <div className="fc-evidence-section">
        <button
          className="fc-evidence-toggle"
          onClick={() => setShowEvidence(!showEvidence)}
        >
          <span className="fc-evidence-toggle-text">Evidence Grounding & Telemetry</span>
          {showEvidence ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
        </button>

        {showEvidence && (
          <div className="fc-evidence-content">
            {/* Supporting Signals Table */}
            {prediction.supporting_signals?.length > 0 && (
              <div className="fc-signals-table-wrap">
                <h5 className="fc-signals-heading">
                  <Layers size={13} />
                  <span>Quantitative Decision Signals</span>
                </h5>
                <table className="fc-signals-table">
                  <thead>
                    <tr>
                      <th>Signal</th>
                      <th>Observed Value</th>
                      <th>Weight</th>
                    </tr>
                  </thead>
                  <tbody>
                    {prediction.supporting_signals.map((sig, idx) => (
                      <tr key={idx}>
                        <td className="font-mono text-slate-600">{sig.signal_name}</td>
                        <td className="font-medium text-slate-900">{String(sig.observed_value)}</td>
                        <td>
                          <div className="flex items-center gap-1.5">
                            <div className="fc-weight-bar-bg">
                              <div
                                className="fc-weight-bar-fill"
                                style={{ width: `${Math.round((sig.importance_weight || 0.5) * 100)}%` }}
                              />
                            </div>
                            <span className="text-xs text-slate-500 font-mono">
                              {sig.importance_weight ? `${Math.round(sig.importance_weight * 100)}%` : '—'}
                            </span>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            {/* Source References */}
            {prediction.source_references?.length > 0 && (
              <div className="fc-sources-wrap">
                <h5 className="fc-signals-heading">
                  <Info size={13} />
                  <span>Grounding Database Records</span>
                </h5>
                <div className="fc-sources-grid">
                  {prediction.source_references.map((ref, idx) => (
                    <div key={idx} className="fc-source-chip">
                      <span className="fc-source-chip__module">{ref.source_module}</span>
                      <span className="fc-source-chip__field">{ref.source_field}</span>
                      <span className="fc-source-chip__id">Record #{ref.source_record_id}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Safety & Authoritative Guardrails Banner */}
            <div className="fc-guardrail-banner">
              <ShieldCheck size={15} className="text-slate-500 shrink-0 mt-0.5" />
              <div className="text-xs text-slate-600 leading-relaxed">
                <strong>Authoritative Ledger Integrity:</strong> Go remains the strict authority for all deterministic invoice calculations (total, balance, payments, due dates). Python agentic AI produces advisory insights only and cannot mutate accounting records or trigger unapproved customer communications.
              </div>
            </div>

            {/* Metadata Footer */}
            <div className="fc-evidence-footer">
              <span>Prediction ID: <code>{prediction.prediction_id}</code></span>
              <span>Generated: {formatDateTime(prediction.source_timestamp)}</span>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
