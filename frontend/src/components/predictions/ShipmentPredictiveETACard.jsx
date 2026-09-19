import React, { useState, useEffect, useCallback } from 'react';
import {
  Calendar,
  Clock,
  ShieldCheck,
  AlertTriangle,
  AlertCircle,
  RefreshCw,
  Sparkles,
  ChevronDown,
  ChevronUp,
  ExternalLink,
  CheckCircle2,
  Lock,
  ArrowRight,
  Info,
  Ship,
  Navigation
} from 'lucide-react';
import predictionService from '../../services/predictionService';
import { SeverityBadge, ConfidenceBadge, ValueTypeBadge } from './PredictionBadge';
import './ShipmentPredictiveETACard.css';

/**
 * Format timestamp nicely into local string
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

export default function ShipmentPredictiveETACard({
  shipmentId,
  authoritativeEta = null,
  onActionTriggered = null,
  className = '',
}) {
  const [prediction, setPrediction] = useState(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState(null);
  const [showEvidence, setShowEvidence] = useState(false);
  const [actionSuccessMsg, setActionSuccessMsg] = useState(null);
  const [acting, setActing] = useState(false);

  const fetchPrediction = useCallback(async (forceRefresh = false) => {
    if (!shipmentId) return;
    if (forceRefresh) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }
    setError(null);
    try {
      const data = await predictionService.getShipmentPredictedETA(shipmentId, forceRefresh);
      const pred = data?.data || data?.prediction || data;
      setPrediction(pred);
    } catch (err) {
      console.error('Failed to load predictive ETA:', err);
      setError(err.response?.data?.message || err.message || 'Failed to load predictive ETA telemetry');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [shipmentId]);

  useEffect(() => {
    fetchPrediction(false);
  }, [fetchPrediction]);

  const handleAcknowledge = async () => {
    if (!prediction?.prediction_id) return;
    try {
      setActing(true);
      await predictionService.acknowledgePrediction(prediction.prediction_id);
      setPrediction(prev => ({ ...prev, status: 'ACKNOWLEDGED' }));
      setActionSuccessMsg('Forecast acknowledged by operations coordinator.');
      if (onActionTriggered) onActionTriggered('ACKNOWLEDGE');
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
      await predictionService.requestAction(prediction.prediction_id, 'Requested consignee notification and carrier schedule verification');
      setPrediction(prev => ({ ...prev, status: 'ACTION_REQUESTED', review_status: 'AWAITING_APPROVAL' }));
      setActionSuccessMsg('Consignee notification queued in Action System. Awaiting Human-in-the-Loop review.');
      if (onActionTriggered) onActionTriggered('REQUEST_ACTION');
    } catch (err) {
      alert('Failed to queue action: ' + (err.message || 'Unknown error'));
    } finally {
      setActing(false);
    }
  };

  if (loading && !refreshing) {
    return (
      <div className={`predictive-eta-card ${className}`} data-testid="shipment-predictive-eta-loading">
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px', color: '#64748b' }}>
          <RefreshCw className="animate-spin" size={16} />
          <span style={{ fontSize: '13px', fontWeight: 500 }}>Synthesizing real-time vessel telemetry and predictive ETA...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`predictive-eta-card ${className}`} data-testid="shipment-predictive-eta-error">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', color: '#dc2626' }}>
            <AlertCircle size={16} />
            <span style={{ fontSize: '13px', fontWeight: 600 }}>Predictive ETA Offline</span>
            <span style={{ fontSize: '12px', color: '#64748b' }}>({error})</span>
          </div>
          <button className="btn-predictive-refresh" onClick={() => fetchPrediction(true)}>
            <RefreshCw size={12} /> Retry
          </button>
        </div>
      </div>
    );
  }

  if (!prediction) return null;

  const severity = String(prediction.severity || prediction.risk_level || 'LOW').toUpperCase();
  const confidenceScore = typeof prediction.confidence_score === 'number' ? Math.round(prediction.confidence_score * 100) : 85;
  const isInsufficient = prediction.review_status === 'INSUFFICIENT_DATA' || prediction.insufficient_data || severity === 'INSUFFICIENT_DATA';
  const isDelivered = prediction.predicted_value === 'ARRIVED';

  return (
    <div
      className={`predictive-eta-card predictive-eta-card--${severity.toLowerCase()} ${className}`}
      data-testid="shipment-predictive-eta-card"
    >
      {/* Header */}
      <div className="predictive-eta-header">
        <div className="predictive-eta-title-wrap">
          <div className="predictive-eta-icon-badge">
            <Sparkles size={18} />
          </div>
          <div>
            <h4 className="predictive-eta-title">
              Predictive ETA & Schedule Variance
              <span className="predictive-eta-badge-phase">Phase 4 Intelligence</span>
            </h4>
            <p className="predictive-eta-subtitle">
              Source-grounded voyage modeling based on real carrier telemetry, port dwell, and active exceptions
            </p>
          </div>
        </div>
        <div className="predictive-eta-actions-top">
          <SeverityBadge severity={severity} />
          <ConfidenceBadge score={prediction.confidence_score} band={prediction.confidence_band} />
          <button
            className="btn-predictive-refresh"
            onClick={() => fetchPrediction(true)}
            disabled={refreshing}
            title="Force refresh predictive model"
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

      {/* Insufficient Data Banner */}
      {isInsufficient ? (
        <div className="predictive-eta-insufficient" data-testid="predictive-eta-insufficient-banner">
          <AlertTriangle size={18} style={{ flexShrink: 0, marginTop: '2px' }} />
          <div>
            <strong>Insufficient Schedule Telemetry:</strong>
            <p style={{ margin: '4px 0 0 0' }}>
              {prediction.explanation || 'Authoritative ETA or departure milestone is missing. Update the master shipment ETA to activate forward voyage modeling.'}
            </p>
          </div>
        </div>
      ) : (
        <>
          {/* Comparison Metric Grid: Authoritative vs Predicted */}
          <div className="predictive-eta-metric-grid">
            {/* Authoritative ETA */}
            <div className="predictive-eta-metric-item">
              <span className="predictive-eta-metric-label">
                <Lock size={12} /> Authoritative Master ETA
              </span>
              <span className="predictive-eta-metric-value">
                {formatDateTime(authoritativeEta || prediction.target_date)}
              </span>
              <span className="predictive-eta-metric-sub">Official booking schedule</span>
            </div>

            {/* Predicted Arrival Window */}
            <div className="predictive-eta-metric-item">
              <span className="predictive-eta-metric-label" style={{ color: '#2563eb' }}>
                <Sparkles size={12} /> Predicted Arrival Window
              </span>
              <span className="predictive-eta-metric-value predictive-eta-metric-value--predicted">
                {isDelivered ? 'Concluded (Arrived)' : (prediction.predicted_arrival_window || prediction.predicted_value || 'On Track')}
              </span>
              <span className="predictive-eta-metric-sub">Grounded ML transit forecast</span>
            </div>

            {/* Delay Risk Level */}
            <div className="predictive-eta-metric-item">
              <span className="predictive-eta-metric-label">Delay Risk Classification</span>
              <span className="predictive-eta-metric-value">
                {severity} RISK
              </span>
              <span className="predictive-eta-metric-sub">
                {severity === 'CRITICAL' ? 'Immediate escalation required' : (severity === 'HIGH' ? 'Consignee notice recommended' : 'Within normal operational tolerance')}
              </span>
            </div>

            {/* Projected Variance */}
            <div className="predictive-eta-metric-item">
              <span className="predictive-eta-metric-label">Schedule Variance</span>
              <span className="predictive-eta-metric-value">
                {prediction.predicted_value || '0h (On Schedule)'}
              </span>
              <span className="predictive-eta-metric-sub">Observed drift + buffer</span>
            </div>
          </div>

          {/* Statement & Explanation */}
          <div className="predictive-eta-statement-box">
            <p className="predictive-eta-statement">{prediction.prediction_statement}</p>
            <p className="predictive-eta-explanation">{prediction.explanation}</p>
          </div>

          {/* Confidence Meter */}
          <div className="predictive-confidence-wrapper">
            <div className="predictive-confidence-header">
              <span>Forecast Confidence</span>
              <span style={{ fontWeight: 600 }}>{confidenceScore}% Certainty</span>
            </div>
            <div className="predictive-confidence-bar-bg">
              <div
                className="predictive-confidence-bar-fill"
                style={{
                  width: `${confidenceScore}%`,
                  background: confidenceScore >= 80 ? '#10b981' : (confidenceScore >= 60 ? '#f59e0b' : '#ef4444')
                }}
              />
            </div>
          </div>

          {/* Collapsible Evidence Section */}
          <div className="predictive-eta-evidence">
            <button
              className="btn-toggle-evidence"
              onClick={() => setShowEvidence(!showEvidence)}
              type="button"
            >
              {showEvidence ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
              {showEvidence ? 'Hide Source Citations & Signals' : 'View Supporting Signals & Source Citations'}
            </button>

            {showEvidence && (
              <div style={{ marginTop: '10px' }}>
                {prediction.supporting_signals?.length > 0 && (
                  <>
                    <span style={{ fontSize: '11px', fontWeight: 700, color: '#475569', textTransform: 'uppercase' }}>
                      Supporting Analytical Signals:
                    </span>
                    <div className="predictive-signals-list">
                      {prediction.supporting_signals.map((sig, idx) => (
                        <div key={idx} className="predictive-signal-badge">
                          <span className="predictive-signal-name">{sig.signal_name?.replace(/_/g, ' ')}</span>
                          <span className="predictive-signal-values">
                            Observed: {String(sig.observed_value)}
                            {sig.baseline_value && <span style={{ fontWeight: 400, color: '#64748b' }}> (Norm: {String(sig.baseline_value)})</span>}
                          </span>
                        </div>
                      ))}
                    </div>
                  </>
                )}

                {prediction.source_references?.length > 0 && (
                  <div className="predictive-sources-list">
                    <span style={{ fontWeight: 600, display: 'block', marginBottom: '4px' }}>
                      Verifiable Source Records (No Fabricated Data):
                    </span>
                    {prediction.source_references.map((src, idx) => (
                      <div key={idx} className="predictive-source-item">
                        <span style={{ background: '#e2e8f0', padding: '1px 5px', borderRadius: '3px', fontSize: '10px', fontWeight: 600 }}>
                          {src.source_module}
                        </span>
                        <span>Field: <code>{src.source_field}</code></span>
                        <span>Record #{src.source_record_id}</span>
                        {src.source_timestamp && (
                          <span style={{ color: '#94a3b8' }}>• Timestamp: {formatDateTime(src.source_timestamp)}</span>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Recommended Action System Intervention */}
          {prediction.recommended_action && (
            <div className="predictive-action-box">
              <div className="predictive-action-text-wrap">
                <Navigation className="predictive-action-icon" size={18} />
                <div>
                  <h5 className="predictive-action-title">Recommended Proactive Intervention</h5>
                  <p className="predictive-action-desc">{prediction.recommended_action}</p>
                </div>
              </div>
              <div className="predictive-action-btns">
                {prediction.status !== 'ACKNOWLEDGED' && prediction.status !== 'ACTION_REQUESTED' && (
                  <button
                    className="btn-pred-action-secondary"
                    onClick={handleAcknowledge}
                    disabled={acting}
                  >
                    <CheckCircle2 size={13} /> Acknowledge
                  </button>
                )}
                {prediction.is_action_required && prediction.status !== 'ACTION_REQUESTED' && (
                  <button
                    className="btn-pred-action-primary"
                    onClick={handleRequestAction}
                    disabled={acting}
                  >
                    <ArrowRight size={13} /> Request Action (HITL)
                  </button>
                )}
                {prediction.status === 'ACTION_REQUESTED' && (
                  <span style={{ fontSize: '12px', fontWeight: 600, color: '#d97706', display: 'inline-flex', alignItems: 'center', gap: '5px' }}>
                    <Clock size={13} /> Awaiting Approval
                  </span>
                )}
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
