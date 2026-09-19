import React, { useState, useEffect, useCallback } from 'react';
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
  ExternalLink,
  ShieldAlert,
  Send
} from 'lucide-react';
import predictionService from '../../services/predictionService';
import { SeverityBadge, ConfidenceBadge } from './PredictionBadge';
import './ShipmentDisruptionForecastCard.css';

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
 * Resolve visual styling class for disruption category
 */
function getCategoryClass(category = '') {
  const cat = String(category).toLowerCase();
  if (cat.includes('customs') || cat.includes('document')) return 'disruption-category-badge--customs';
  if (cat.includes('weather') || cat.includes('cyclone') || cat.includes('typhoon')) return 'disruption-category-badge--weather';
  if (cat.includes('tracking') || cat.includes('inactivity')) return 'disruption-category-badge--tracking';
  if (cat.includes('milestone') || cat.includes('schedule')) return 'disruption-category-badge--milestone';
  return '';
}

export default function ShipmentDisruptionForecastCard({
  shipmentId,
  onNavigateException = null,
  className = '',
}) {
  const [prediction, setPrediction] = useState(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState(null);
  const [showEvidence, setShowEvidence] = useState(false);
  const [actionSuccessMsg, setActionSuccessMsg] = useState(null);
  const [acting, setActing] = useState(false);

  const fetchForecast = useCallback(async (forceRefresh = false) => {
    if (!shipmentId) return;
    if (forceRefresh) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }
    setError(null);
    try {
      const data = await predictionService.getShipmentPredictedExceptions(shipmentId, forceRefresh);
      const pred = data?.data || data?.prediction || data;
      setPrediction(pred);
    } catch (err) {
      console.error('Failed to load disruption forecast:', err);
      setError(err.response?.data?.message || err.message || 'Failed to load predictive disruption forecast');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [shipmentId]);

  useEffect(() => {
    fetchForecast(false);
  }, [fetchForecast]);

  const handleAcknowledge = async () => {
    if (!prediction?.prediction_id) return;
    try {
      setActing(true);
      await predictionService.acknowledgePrediction(prediction.prediction_id);
      setPrediction(prev => ({ ...prev, status: 'ACKNOWLEDGED' }));
      setActionSuccessMsg('Early warning signal acknowledged by operations specialist.');
    } catch (err) {
      alert('Failed to acknowledge prediction: ' + (err.message || 'Unknown error'));
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
        `Proactive mitigation for ${prediction.disruption_category || 'disruption risk'}: ${prediction.recommended_action || 'Operational escalation'}`
      );
      setPrediction(prev => ({ ...prev, status: 'ACTION_REQUESTED', review_status: 'AWAITING_APPROVAL' }));
      setActionSuccessMsg('Mitigation action queued in Go Action System. Awaiting Human-in-the-Loop approval.');
    } catch (err) {
      alert('Failed to queue action: ' + (err.message || 'Unknown error'));
    } finally {
      setActing(false);
    }
  };

  if (loading && !refreshing) {
    return (
      <div className={`disruption-forecast-card ${className}`} data-testid="shipment-disruption-loading">
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px', color: '#64748b' }}>
          <RefreshCw className="animate-spin" size={16} />
          <span style={{ fontSize: '13px', fontWeight: 500 }}>
            Synthesizing early disruption warning signals & operational risk...
          </span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`disruption-forecast-card ${className}`} data-testid="shipment-disruption-error">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', color: '#dc2626' }}>
            <AlertCircle size={16} />
            <span style={{ fontSize: '13px', fontWeight: 600 }}>Disruption Intelligence Offline</span>
            <span style={{ fontSize: '12px', color: '#64748b' }}>({error})</span>
          </div>
          <button className="btn-disruption-refresh" onClick={() => fetchForecast(true)}>
            <RefreshCw size={12} /> Retry
          </button>
        </div>
      </div>
    );
  }

  if (!prediction) return null;

  const severity = String(prediction.severity || prediction.risk_level || 'LOW').toUpperCase();
  const isInsufficient = prediction.review_status === 'INSUFFICIENT_DATA' || prediction.insufficient_data || severity === 'INSUFFICIENT_DATA';
  const signals = Array.isArray(prediction.supporting_signals)
    ? prediction.supporting_signals.map(s => (typeof s === 'string' ? s : `${s.signal_name?.replace(/_/g, ' ')}: ${s.observed_value}`))
    : [];
  const category = prediction.disruption_category || prediction.predicted_value || 'OPERATIONAL_RISK';
  const linkedExceptionId = prediction.linked_exception_id ||
    (prediction.source_references?.find(s => s.source_module === 'shipment_exceptions')?.source_record_id);

  return (
    <div
      className={`disruption-forecast-card disruption-forecast-card--${severity.toLowerCase()} ${className}`}
      data-testid="shipment-disruption-forecast-card"
    >
      {/* Header */}
      <div className="disruption-forecast-header">
        <div className="disruption-forecast-title-wrap">
          <div className="disruption-forecast-icon-badge">
            <ShieldAlert size={18} />
          </div>
          <div>
            <h4 className="disruption-forecast-title">
              Predictive Exception & Disruption Forecasting
              <span className="disruption-forecast-badge-phase">Phase 4 Intelligence</span>
            </h4>
            <p className="disruption-forecast-subtitle">
              Early warning signals and operational risk forecasting before exceptions confirm
            </p>
          </div>
        </div>
        <div className="disruption-forecast-actions-top">
          <SeverityBadge severity={severity} />
          <ConfidenceBadge score={prediction.confidence_score} band={prediction.confidence_band} />
          <button
            className="btn-disruption-refresh"
            onClick={() => fetchForecast(true)}
            disabled={refreshing}
            title="Force refresh exception and disruption forecast"
          >
            <RefreshCw className={refreshing ? 'animate-spin' : ''} size={12} />
            {refreshing ? 'Refreshing...' : 'Refresh'}
          </button>
        </div>
      </div>

      {actionSuccessMsg && (
        <div style={{ background: '#f0fdf4', border: '1px solid #bbf7d0', color: '#166534', padding: '10px 14px', borderRadius: '8px', fontSize: '12px', marginBottom: '14px', display: 'flex', alignItems: 'center', gap: '8px' }}>
          <CheckCircle2 size={14} />
          <span>{actionSuccessMsg}</span>
        </div>
      )}

      {/* Insufficient Data State */}
      {isInsufficient ? (
        <div className="disruption-forecast-insufficient" data-testid="disruption-forecast-insufficient-banner">
          <AlertTriangle size={18} style={{ flexShrink: 0, marginTop: '2px' }} />
          <div>
            <strong>Insufficient Operational Telemetry:</strong>
            <p style={{ margin: '4px 0 0 0' }}>
              {prediction.explanation || 'Operational milestones or tracking telemetry missing. Update tracking positions to activate proactive disruption forecasting.'}
            </p>
          </div>
        </div>
      ) : (
        <>
          {/* Key Metric & Signal Grid */}
          <div className="disruption-forecast-metric-grid">
            {/* Early Warning Category */}
            <div className="disruption-forecast-metric-item">
              <span className="disruption-forecast-metric-label">
                <Radio size={12} /> Disruption Category
              </span>
              <div style={{ marginTop: '2px' }}>
                <span className={`disruption-category-badge ${getCategoryClass(category)}`}>
                  {category.replace(/_/g, ' ')}
                </span>
              </div>
              <span className="disruption-forecast-metric-sub">Identified threat domain</span>
            </div>

            {/* Risk Classification */}
            <div className="disruption-forecast-metric-item">
              <span className="disruption-forecast-metric-label">
                <ShieldCheck size={12} /> Risk Severity
              </span>
              <span className="disruption-forecast-metric-value">
                {severity} RISK
              </span>
              <span className="disruption-forecast-metric-sub">
                {severity === 'CRITICAL' ? 'Immediate escalation required' : (severity === 'HIGH' ? 'Preemptive mitigation advised' : 'Operational tolerance')}
              </span>
            </div>

            {/* Confirmed Exception Linkage */}
            <div className="disruption-forecast-metric-item">
              <span className="disruption-forecast-metric-label">
                <FileText size={12} /> Confirmed Exception Link
              </span>
              <div style={{ marginTop: '2px' }}>
                {linkedExceptionId ? (
                  <button
                    type="button"
                    className="disruption-linked-pill"
                    onClick={() => onNavigateException && onNavigateException(linkedExceptionId)}
                    title="Click to view existing confirmed exception"
                    style={{ border: 'none', cursor: onNavigateException ? 'pointer' : 'default' }}
                  >
                    Exception #{linkedExceptionId} <ExternalLink size={10} />
                  </button>
                ) : (
                  <span style={{ fontSize: '13px', fontWeight: 600, color: '#10b981' }}>
                    No Active Exception (Early Warning)
                  </span>
                )}
              </div>
              <span className="disruption-forecast-metric-sub">
                {linkedExceptionId ? 'Correlated with open operational issue' : 'Pre-confirmation detection'}
              </span>
            </div>

            {/* Telemetry Age & Grounding */}
            <div className="disruption-forecast-metric-item">
              <span className="disruption-forecast-metric-label">
                <Clock size={12} /> Evaluated Source
              </span>
              <span className="disruption-forecast-metric-value" style={{ fontSize: '12px' }}>
                {formatDateTime(prediction.created_at)}
              </span>
              <span className="disruption-forecast-metric-sub">Authoritative real MariaDB records</span>
            </div>
          </div>

          {/* Statement & Source-Grounded Explanation */}
          <div className="disruption-forecast-statement-box">
            <div className="disruption-forecast-statement-title">
              <Activity size={14} color="#0b192c" /> Early Warning Disruption Assessment
            </div>
            <p className="disruption-forecast-statement-text" data-testid="disruption-statement">
              {prediction.prediction_statement || 'Operational risk within standard freight variance tolerances.'}
            </p>
            <p className="disruption-forecast-explanation-text" data-testid="disruption-explanation">
              {prediction.explanation}
            </p>
          </div>

          {/* Grounded Evidence / Supporting Signals Collapsible */}
          <div className="disruption-forecast-evidence-section">
            <button
              className="disruption-forecast-evidence-toggle"
              onClick={() => setShowEvidence(!showEvidence)}
              type="button"
              aria-expanded={showEvidence}
            >
              <span style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                <ShieldCheck size={14} color="#2563eb" />
                Source-Grounded Signals & Audit Records ({signals.length} telemetry points)
              </span>
              {showEvidence ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
            </button>

            {showEvidence && (
              <div className="disruption-forecast-evidence-content" data-testid="disruption-evidence-content">
                <ul className="disruption-signals-list">
                  {signals.map((sig, idx) => (
                    <li key={idx} className="disruption-signal-item">
                      <span style={{ color: '#2563eb', fontWeight: 'bold' }}>•</span>
                      <span>{sig}</span>
                    </li>
                  ))}
                </ul>

                <div className="disruption-evidence-grid">
                  <div className="disruption-evidence-row">
                    <strong>Prediction ID:</strong>
                    <span>{prediction.prediction_id || 'N/A'}</span>
                  </div>
                  <div className="disruption-evidence-row">
                    <strong>Correlation ID:</strong>
                    <span>{prediction.correlation_id || 'N/A'}</span>
                  </div>
                  <div className="disruption-evidence-row">
                    <strong>Model Version:</strong>
                    <span>{prediction.prediction_version || 'v4.3.0'}</span>
                  </div>
                  <div className="disruption-evidence-row">
                    <strong>Status:</strong>
                    <span style={{ color: '#059669', fontWeight: 600 }}>{prediction.status || 'PUBLISHED'}</span>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Proactive Action System & Approval Footer */}
          <div className="disruption-forecast-action-panel">
            <div className="disruption-forecast-action-info">
              <span className="disruption-forecast-action-title">
                <Send size={13} /> Recommended Proactive Mitigation:
              </span>
              <p className="disruption-forecast-action-desc">
                {prediction.recommended_action || 'Continue standard voyage tracking monitoring.'}
              </p>
              <div className="disruption-hitl-note">
                <Lock size={11} />
                <span>Human-in-the-Loop Safeguard: Actions require approval via Go Action System. Authoritative shipment records remain immutable.</span>
              </div>
            </div>

            <div className="disruption-forecast-action-buttons">
              {prediction.status !== 'ACKNOWLEDGED' && prediction.status !== 'ACTION_REQUESTED' && (
                <button
                  className="btn-disruption-ack"
                  onClick={handleAcknowledge}
                  disabled={acting}
                  type="button"
                >
                  <CheckCircle2 size={13} /> Acknowledge Risk
                </button>
              )}

              {prediction.status !== 'ACTION_REQUESTED' ? (
                <button
                  className="btn-disruption-action"
                  onClick={handleRequestAction}
                  disabled={acting}
                  type="button"
                >
                  <ShieldAlert size={13} /> Queue Mitigation Action
                </button>
              ) : (
                <span style={{ fontSize: '12px', fontWeight: 600, color: '#ca8a04', display: 'flex', alignItems: 'center', gap: '4px' }}>
                  <Clock size={13} /> Awaiting HITL Approval
                </span>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
