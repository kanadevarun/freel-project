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
  Send,
  UserCheck,
  TrendingUp,
  Award,
  Users,
  Sparkles
} from 'lucide-react';
import predictionService from '../../services/predictionService';
import { SeverityBadge, ConfidenceBadge } from './PredictionBadge';
import './CustomerLeadPredictiveIntelligenceCard.css';

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
 * Resolve display label and color style for prediction category
 */
function getCategoryInfo(category = '', predictionType = '') {
  const cat = String(category || predictionType).toUpperCase();
  if (cat.includes('CONVERSION')) {
    return { label: 'Conversion Likelihood', colorClass: 'cl-category-badge--conversion' };
  }
  if (cat.includes('REPEAT')) {
    return { label: 'Repeat Business Momentum', colorClass: 'cl-category-badge--repeat' };
  }
  if (cat.includes('INACTIVITY')) {
    return { label: 'Lead Inactivity Risk', colorClass: 'cl-category-badge--inactivity' };
  }
  if (cat.includes('CHURN')) {
    return { label: 'Account Attrition Risk', colorClass: 'cl-category-badge--churn' };
  }
  if (cat.includes('ENGAGEMENT')) {
    return { label: 'Quotation Response Risk', colorClass: 'cl-category-badge--engagement' };
  }
  return { label: 'Customer Intelligence', colorClass: 'cl-category-badge--neutral' };
}

export default function CustomerLeadPredictiveIntelligenceCard({
  recordType = 'lead', // 'lead' | 'customer'
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
      if (recordType === 'customer') {
        data = await predictionService.getCustomerPredictedIntelligence(recordId, forceRefresh);
      } else {
        data = await predictionService.getLeadPredictedIntelligence(recordId, forceRefresh);
      }
      const pred = data?.data || data?.prediction || data;
      setPrediction(pred);
    } catch (err) {
      console.error(`Failed to load predictive intelligence for ${recordType}:`, err);
      setError(err.response?.data?.message || err.message || `Failed to load predictive intelligence for ${recordType}`);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [recordType, recordId]);

  useEffect(() => {
    fetchIntelligence(false);
  }, [fetchIntelligence]);

  const handleAcknowledge = async () => {
    if (!prediction?.prediction_id) return;
    try {
      setActing(true);
      await predictionService.acknowledgePrediction(prediction.prediction_id);
      setPrediction(prev => ({ ...prev, status: 'ACKNOWLEDGED' }));
      setActionSuccessMsg('Intelligence insight acknowledged.');
      setTimeout(() => setActionSuccessMsg(null), 4000);
    } catch (err) {
      console.error('Failed to acknowledge intelligence:', err);
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
        `Sales representative initiated mitigation for ${prediction.prediction_type} on ${recordType} #${recordId}`
      );
      setPrediction(prev => ({ ...prev, status: 'ACTION_REQUESTED' }));
      setActionSuccessMsg('Mitigation action queued for Human-In-The-Loop approval.');
      setTimeout(() => setActionSuccessMsg(null), 4500);
    } catch (err) {
      console.error('Failed to request mitigation action:', err);
      setError(err.response?.data?.message || 'Failed to request action');
    } finally {
      setActing(false);
    }
  };

  const isCustomer = recordType === 'customer';
  const categoryInfo = getCategoryInfo(prediction?.prediction_category, prediction?.prediction_type);

  if (loading) {
    return (
      <div className={`cl-pred-card cl-pred-card--loading ${className}`}>
        <div className="cl-pred-card__loading-content">
          <RefreshCw className="cl-pred-card__spinner animate-spin text-indigo-600" size={24} />
          <span className="cl-pred-card__loading-text">
            Evaluating real-time {isCustomer ? 'customer retention' : 'lead conversion'} signals...
          </span>
        </div>
      </div>
    );
  }

  if (error && !prediction) {
    return (
      <div className={`cl-pred-card cl-pred-card--error ${className}`}>
        <div className="cl-pred-card__error-content">
          <AlertCircle className="text-red-500 flex-shrink-0" size={20} />
          <div className="cl-pred-card__error-body">
            <h4 className="cl-pred-card__error-title">Predictive Intelligence Unavailable</h4>
            <p className="cl-pred-card__error-desc">{error}</p>
          </div>
          <button
            onClick={() => fetchIntelligence(true)}
            className="cl-pred-card__btn cl-pred-card__btn--secondary"
          >
            Retry Analysis
          </button>
        </div>
      </div>
    );
  }

  if (!prediction) {
    return (
      <div className={`cl-pred-card cl-pred-card--empty ${className}`}>
        <div className="cl-pred-card__empty-content">
          <Info className="text-slate-400 flex-shrink-0" size={18} />
          <p className="cl-pred-card__empty-text">
            No predictive signals currently detected for this {isCustomer ? 'customer' : 'lead'}.
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

  return (
    <div className={`cl-pred-card cl-pred-card--severity-${severity.toLowerCase()} ${className}`}>
      {/* Card Header */}
      <div className="cl-pred-card__header">
        <div className="cl-pred-card__title-group">
          <div className="cl-pred-card__icon-wrapper">
            {isCustomer ? <TrendingUp size={18} /> : <UserCheck size={18} />}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="cl-pred-card__title">
                {isCustomer ? 'Predictive Customer Intelligence' : 'Predictive Lead Intelligence'}
              </h3>
              <span className={`cl-category-badge ${categoryInfo.colorClass}`}>
                {categoryInfo.label}
              </span>
            </div>
            <p className="cl-pred-card__meta">
              Ground-truth source: MariaDB authoritative CRM records &middot; Telemetry as of {formatDateTime(prediction.source_timestamp)}
            </p>
          </div>
        </div>

        <div className="cl-pred-card__actions-header">
          <button
            onClick={() => fetchIntelligence(true)}
            disabled={refreshing}
            className="cl-pred-card__refresh-btn"
            title="Recalculate predictive signals from real database telemetry"
          >
            <RefreshCw size={14} className={refreshing ? 'animate-spin text-indigo-600' : 'text-slate-500'} />
            <span>{refreshing ? 'Recalculating...' : 'Recalculate'}</span>
          </button>
        </div>
      </div>

      {/* Success Notification Banner */}
      {actionSuccessMsg && (
        <div className="cl-pred-card__banner cl-pred-card__banner--success">
          <CheckCircle2 size={16} className="text-emerald-600 flex-shrink-0" />
          <span>{actionSuccessMsg}</span>
        </div>
      )}

      {/* Badges Ribbon */}
      <div className="cl-pred-card__ribbon">
        <div className="flex items-center gap-2 flex-wrap">
          <SeverityBadge severity={severity} />
          <ConfidenceBadge confidenceScore={prediction.confidence_score} confidenceBand={prediction.confidence_band} />

          {prediction.status && (
            <span className={`cl-status-pill cl-status-pill--${prediction.status.toLowerCase()}`}>
              {prediction.status}
            </span>
          )}

          {prediction.time_horizon && (
            <span className="cl-pred-card__horizon-pill">
              <Clock size={12} />
              Horizon: {prediction.time_horizon.replace('_', ' ')}
            </span>
          )}
        </div>
      </div>

      {/* Prediction Primary Statement */}
      <div className="cl-pred-card__statement-box">
        <div className="cl-pred-card__statement-icon">
          {severity === 'CRITICAL' || severity === 'HIGH' ? (
            <AlertTriangle size={18} className="text-rose-600" />
          ) : (
            <Sparkles size={18} className="text-indigo-600" />
          )}
        </div>
        <div className="cl-pred-card__statement-content">
          <p className="cl-pred-card__statement-text">
            {prediction.prediction_statement}
          </p>
          {prediction.explanation && (
            <p className="cl-pred-card__explanation-text">
              {prediction.explanation}
            </p>
          )}
        </div>
      </div>

      {/* Key Drivers & Grounded Signals Grid */}
      {supportingSignals.length > 0 && (
        <div className="cl-pred-card__signals-section">
          <h4 className="cl-pred-card__section-subtitle">
            <Activity size={14} className="text-indigo-600" />
            Key Deterministic Drivers & Observed Signals
          </h4>
          <div className="cl-pred-card__signals-grid">
            {supportingSignals.map((sig, idx) => (
              <div key={idx} className="cl-signal-tile">
                <div className="cl-signal-tile__name">{sig.signal_name?.replace(/_/g, ' ')}</div>
                <div className="cl-signal-tile__values">
                  <span className="cl-signal-tile__observed">{sig.observed_value}</span>
                  {sig.baseline_value && (
                    <span className="cl-signal-tile__baseline">
                      baseline: {sig.baseline_value}
                    </span>
                  )}
                </div>
                {sig.importance_weight !== undefined && (
                  <div className="cl-signal-tile__weight">
                    Influence: {Math.round(sig.importance_weight * 100)}%
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Recommended Proactive Action (Action System) */}
      {prediction.recommended_action && (
        <div className="cl-pred-card__recommendation-box">
          <div className="cl-pred-card__rec-header">
            <div className="flex items-center gap-2">
              <Send size={15} className="text-indigo-600" />
              <span className="font-semibold text-slate-800 text-sm">Recommended Commercial Next Step</span>
            </div>
            {prediction.requires_approval ? (
              <span className="cl-pred-card__hitl-badge" title="Automated mutation prevented: Human-in-the-loop review required">
                <Lock size={12} />
                Requires Sales/Manager Approval
              </span>
            ) : (
              <span className="cl-pred-card__internal-badge">
                <ShieldCheck size={12} />
                Internal Advisory
              </span>
            )}
          </div>
          <p className="cl-pred-card__rec-text">{prediction.recommended_action}</p>

          <div className="cl-pred-card__rec-actions">
            {!isAcknowledged && !isActionRequested && (
              <button
                onClick={handleAcknowledge}
                disabled={acting}
                className="cl-pred-card__btn cl-pred-card__btn--secondary"
              >
                <CheckCircle2 size={14} />
                Acknowledge Insight
              </button>
            )}

            {prediction.is_action_required && !isActionRequested && (
              <button
                onClick={handleRequestAction}
                disabled={acting}
                className="cl-pred-card__btn cl-pred-card__btn--primary"
              >
                <ArrowRight size={14} />
                Queue Mitigation Action
              </button>
            )}

            {isActionRequested && (
              <div className="cl-pred-card__action-pending-badge">
                <Clock size={14} />
                Action Queued for HITL Approval &middot; Go Action System
              </div>
            )}
          </div>
        </div>
      )}

      {/* Expandable Grounded Source References & Telemetry */}
      <div className="cl-pred-card__evidence-wrapper">
        <button
          onClick={() => setShowEvidence(!showEvidence)}
          className="cl-pred-card__evidence-toggle"
        >
          <div className="flex items-center gap-2">
            <Radio size={14} className="text-slate-500" />
            <span>Telemetry Audit &amp; Data Grounding ({sourceRefs.length} DB References)</span>
          </div>
          {showEvidence ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
        </button>

        {showEvidence && (
          <div className="cl-pred-card__evidence-body">
            <div className="cl-pred-card__evidence-table-container">
              <table className="cl-pred-card__evidence-table">
                <thead>
                  <tr>
                    <th>Source Module</th>
                    <th>Record ID</th>
                    <th>Database Field</th>
                    <th>Recorded Timestamp</th>
                  </tr>
                </thead>
                <tbody>
                  {sourceRefs.map((ref, idx) => (
                    <tr key={idx}>
                      <td className="font-medium text-slate-700">{ref.source_module}</td>
                      <td>
                        <span className="cl-ref-pill">#{ref.source_record_id}</span>
                      </td>
                      <td className="text-slate-600 font-mono text-xs">{ref.source_field}</td>
                      <td className="text-slate-500 text-xs">{formatDateTime(ref.source_timestamp)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="cl-pred-card__evidence-footer">
              <span>Model: {prediction.model_version || 'CustomerLeadPredictor_v1.0'}</span>
              <span>ID: {prediction.prediction_id}</span>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
