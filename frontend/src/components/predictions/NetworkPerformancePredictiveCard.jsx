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
  Compass,
  Anchor,
  MapPin,
  Sparkles
} from 'lucide-react';
import predictionService from '../../services/predictionService';
import { SeverityBadge, ConfidenceBadge } from './PredictionBadge';
import './NetworkPerformancePredictiveCard.css';

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
 * Resolve display label and icon for prediction category
 */
function getCategoryInfo(entityType = 'carrier', category = '', predictionType = '') {
  const cat = String(category || predictionType).toUpperCase();
  if (entityType === 'carrier') {
    if (cat.includes('HOLD') || cat.includes('CUSTOMS')) {
      return { label: 'Carrier Regulatory Hold Risk', colorClass: 'netperf-badge--danger' };
    }
    if (cat.includes('MANIFEST') || cat.includes('DISCREPANCY')) {
      return { label: 'Carrier Documentation Risk', colorClass: 'netperf-badge--warning' };
    }
    if (cat.includes('DELAY') || cat.includes('SCHEDULE')) {
      return { label: 'Carrier Transshipment Schedule Risk', colorClass: 'netperf-badge--warning' };
    }
    return { label: 'Carrier Performance Intelligence', colorClass: 'netperf-badge--info' };
  }
  if (entityType === 'lane') {
    if (cat.includes('CUSTOMS') || cat.includes('DWELL')) {
      return { label: 'Corridor Regulatory Inspection Risk', colorClass: 'netperf-badge--danger' };
    }
    if (cat.includes('MANIFEST')) {
      return { label: 'Corridor Cutoff Compliance Risk', colorClass: 'netperf-badge--warning' };
    }
    return { label: 'Trade Corridor Bottleneck Intelligence', colorClass: 'netperf-badge--info' };
  }
  // customer
  if (cat.includes('SERVICE')) {
    return { label: 'Customer Service Deterioration Risk', colorClass: 'netperf-badge--danger' };
  }
  if (cat.includes('RELATIONSHIP') || cat.includes('GROWTH')) {
    return { label: 'Customer Relationship Opportunity', colorClass: 'netperf-badge--success' };
  }
  return { label: 'Customer Performance Intelligence', colorClass: 'netperf-badge--info' };
}

export default function NetworkPerformancePredictiveCard({
  entityType = 'carrier', // 'carrier' | 'lane' | 'customer'
  entityId,
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
    if (!entityId) return;
    if (forceRefresh) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }
    setError(null);
    try {
      let data;
      if (entityType === 'carrier') {
        data = await predictionService.getCarrierPredictedPerformance(entityId, forceRefresh);
      } else if (entityType === 'lane') {
        data = await predictionService.getLanePredictedPerformance(entityId, forceRefresh);
      } else {
        data = await predictionService.getCustomerPredictedServicePerformance(entityId, forceRefresh);
      }
      const pred = data?.data || data?.prediction || data;
      setPrediction(pred);
    } catch (err) {
      console.error('[NetworkPerformanceCard] Fetch failed:', err);
      setError(err?.response?.data?.message || err?.message || 'Failed to retrieve performance intelligence');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [entityType, entityId]);

  useEffect(() => {
    fetchIntelligence(false);
  }, [fetchIntelligence]);

  const handleRequestAction = async () => {
    if (!prediction?.prediction_id || acting) return;
    setActing(true);
    try {
      await predictionService.requestAction(
        prediction.prediction_id,
        `Operational escalation initiated for ${entityType} ${entityId} via predictive intelligence card`
      );
      setActionSuccessMsg('Action request routed to Action System for human approval.');
      setTimeout(() => setActionSuccessMsg(null), 5000);
      fetchIntelligence(false);
    } catch (err) {
      setError(err?.response?.data?.message || err?.message || 'Failed to request action');
    } finally {
      setActing(false);
    }
  };

  if (loading) {
    return (
      <div className={`netperf-card netperf-card--loading ${className}`}>
        <div className="netperf-loading-spinner" />
        <span>Evaluating authoritative network and operational records...</span>
      </div>
    );
  }

  if (error && !prediction) {
    return (
      <div className={`netperf-card netperf-card--error ${className}`}>
        <div className="netperf-error-header">
          <AlertCircle className="w-5 h-5 text-rose-500" />
          <span className="font-semibold text-slate-800">Performance Intelligence Unavailable</span>
        </div>
        <p className="netperf-error-msg">{error}</p>
        <button
          type="button"
          onClick={() => fetchIntelligence(true)}
          className="netperf-retry-btn"
        >
          <RefreshCw className="w-4 h-4 mr-1.5" />
          Retry Analysis
        </button>
      </div>
    );
  }

  if (!prediction) return null;

  const categoryInfo = getCategoryInfo(entityType, prediction.prediction_category, prediction.prediction_type);
  const isInsufficient = prediction.insufficient_data || prediction.predicted_value === 'INSUFFICIENT_DATA';
  const comparisonPeriod = prediction.comparison_period || 'LAST_90_DAYS';
  const sampleSize = prediction.sample_size ?? 0;

  return (
    <div className={`netperf-card ${className}`}>
      {/* Header */}
      <div className="netperf-header">
        <div className="netperf-header__left">
          <div className="netperf-title-row">
            {entityType === 'carrier' && <Anchor className="w-5 h-5 text-indigo-600" />}
            {entityType === 'lane' && <Compass className="w-5 h-5 text-sky-600" />}
            {entityType === 'customer' && <Users className="w-5 h-5 text-emerald-600" />}
            <h3 className="netperf-title">
              {entityType === 'carrier' && `Carrier Performance Intelligence: ${prediction.carrier_reference || entityId}`}
              {entityType === 'lane' && `Trade Corridor Risk: ${prediction.lane_reference || entityId}`}
              {entityType === 'customer' && `Customer Service & Relationship Intelligence: ${prediction.customer_reference || entityId}`}
            </h3>
            <span className={`netperf-badge ${categoryInfo.colorClass}`}>
              {categoryInfo.label}
            </span>
          </div>

          <div className="netperf-meta-row">
            <span className="netperf-meta-item">
              <Clock className="w-3.5 h-3.5 text-slate-400" />
              Source Fact Time: {formatDateTime(prediction.source_timestamp)}
            </span>
            <span className="netperf-meta-item">
              <Activity className="w-3.5 h-3.5 text-slate-400" />
              Benchmark Window: {comparisonPeriod}
            </span>
            <span className="netperf-meta-item">
              <FileText className="w-3.5 h-3.5 text-slate-400" />
              Verified Sample: {sampleSize} authoritative record{sampleSize !== 1 ? 's' : ''}
            </span>
          </div>
        </div>

        <div className="netperf-header__right">
          <button
            type="button"
            onClick={() => fetchIntelligence(true)}
            disabled={refreshing}
            className="netperf-refresh-btn"
            title="Re-run predictive model against active persistent database records"
          >
            <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
            <span>{refreshing ? 'Refreshing...' : 'Refresh'}</span>
          </button>
        </div>
      </div>

      {/* Main Statement Banner */}
      <div className={`netperf-statement-banner ${isInsufficient ? 'netperf-statement-banner--insufficient' : ''}`}>
        <div className="netperf-statement-icon">
          {isInsufficient ? (
            <Info className="w-5 h-5 text-slate-500" />
          ) : prediction.severity === 'HIGH' || prediction.severity === 'CRITICAL' ? (
            <AlertTriangle className="w-5 h-5 text-amber-500" />
          ) : (
            <ShieldCheck className="w-5 h-5 text-emerald-500" />
          )}
        </div>
        <div className="netperf-statement-text">
          <p className="netperf-statement-title">{prediction.prediction_statement}</p>
          <p className="netperf-statement-explanation">{prediction.explanation}</p>
        </div>
      </div>

      {/* Metrics Row */}
      <div className="netperf-metrics-grid">
        <div className="netperf-metric-card">
          <span className="netperf-metric-label">Predicted Risk Posture</span>
          <div className="netperf-metric-value-row">
            <SeverityBadge severity={prediction.severity} />
            <ConfidenceBadge confidenceBand={prediction.confidence_band} confidenceScore={prediction.confidence_score} />
          </div>
        </div>

        <div className="netperf-metric-card">
          <span className="netperf-metric-label">Forecast Outcome / Value</span>
          <span className="netperf-metric-value font-mono">
            {prediction.predicted_value || (isInsufficient ? 'INSUFFICIENT_DATA' : 'STABLE')}
          </span>
        </div>

        <div className="netperf-metric-card">
          <span className="netperf-metric-label">Trade Corridor / Focus</span>
          <span className="netperf-metric-value text-slate-800">
            {prediction.lane_reference || (entityType === 'lane' ? entityId : 'INNSA-USNYC')}
          </span>
        </div>

        <div className="netperf-metric-card">
          <span className="netperf-metric-label">Forecast Time Horizon</span>
          <span className="netperf-metric-value text-slate-700 font-medium">
            {prediction.time_horizon ? prediction.time_horizon.replace('_', ' ') : '14 DAYS'}
          </span>
        </div>
      </div>

      {/* Action Recommendation Bar */}
      {prediction.recommended_action && (
        <div className="netperf-action-box">
          <div className="netperf-action-content">
            <div className="netperf-action-title">
              <Sparkles className="w-4 h-4 text-indigo-600" />
              <span>Recommended Advisory Action</span>
              {prediction.requires_approval && (
                <span className="netperf-approval-tag">
                  <Lock className="w-3 h-3 mr-1" />
                  Requires Operator Approval
                </span>
              )}
            </div>
            <p className="netperf-action-text">{prediction.recommended_action}</p>
          </div>
          {prediction.is_action_required && (
            <button
              type="button"
              onClick={handleRequestAction}
              disabled={acting || actionSuccessMsg}
              className="netperf-execute-btn"
            >
              <Send className="w-4 h-4 mr-1.5" />
              {acting ? 'Dispatching...' : actionSuccessMsg ? 'Action Queued' : 'Request Action'}
            </button>
          )}
        </div>
      )}

      {actionSuccessMsg && (
        <div className="netperf-success-alert">
          <CheckCircle2 className="w-4 h-4 text-emerald-600 mr-2" />
          <span>{actionSuccessMsg}</span>
        </div>
      )}

      {/* Collapsible Evidence & Source Grounding */}
      <div className="netperf-evidence-section">
        <button
          type="button"
          onClick={() => setShowEvidence(!showEvidence)}
          className="netperf-evidence-toggle"
        >
          <div className="flex items-center gap-2">
            <Radio className="w-4 h-4 text-slate-500" />
            <span className="font-medium text-slate-700">Authoritative Source Grounding & Signals</span>
            <span className="netperf-signal-count">
              ({(prediction.supporting_signals || []).length} signals, {(prediction.source_references || []).length} sources)
            </span>
          </div>
          {showEvidence ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
        </button>

        {showEvidence && (
          <div className="netperf-evidence-body">
            {/* Supporting Signals */}
            {prediction.supporting_signals && prediction.supporting_signals.length > 0 && (
              <div className="netperf-evidence-group">
                <h4 className="netperf-evidence-subheading">Analytical Telemetry & Observed Signals</h4>
                <div className="netperf-signals-table">
                  <div className="netperf-signals-header">
                    <span>Signal Name</span>
                    <span>Observed Fact</span>
                    <span>Baseline / SLA</span>
                    <span>Weight</span>
                  </div>
                  {prediction.supporting_signals.map((sig, idx) => (
                    <div key={idx} className="netperf-signals-row">
                      <span className="font-mono text-xs text-slate-700">{sig.signal_name}</span>
                      <span className="font-semibold text-slate-900">{String(sig.observed_value)}</span>
                      <span className="text-slate-500">{sig.baseline_value ? String(sig.baseline_value) : 'N/A'}</span>
                      <span className="text-indigo-600 font-medium">{Math.round((sig.importance_weight || 1.0) * 100)}%</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Source References */}
            {prediction.source_references && prediction.source_references.length > 0 && (
              <div className="netperf-evidence-group">
                <h4 className="netperf-evidence-subheading">Verifiable Database Records Cited</h4>
                <ul className="netperf-sources-list">
                  {prediction.source_references.map((src, idx) => (
                    <li key={idx} className="netperf-source-item">
                      <FileText className="w-4 h-4 text-slate-400 mt-0.5" />
                      <div>
                        <span className="font-mono text-xs text-indigo-700 font-semibold">
                          {src.source_module}.{src.source_field}
                        </span>
                        <span className="text-slate-500 text-xs ml-2">
                          (Record ID: #{src.source_record_id})
                        </span>
                        {src.carrier_reference && (
                          <span className="netperf-evidence-pill ml-2">Carrier: {src.carrier_reference}</span>
                        )}
                        {src.lane_reference && (
                          <span className="netperf-evidence-pill ml-2">Corridor: {src.lane_reference}</span>
                        )}
                        {src.comparison_period && (
                          <span className="netperf-evidence-pill ml-2">Window: {src.comparison_period}</span>
                        )}
                      </div>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {/* Safety & Compliance Notice */}
            <div className="netperf-safety-notice">
              <ShieldCheck className="w-4 h-4 text-slate-400 mr-2 shrink-0" />
              <span>
                <strong>LogisticsHQ Governance Safeguard:</strong> This predictive assessment is strictly advisory. No contracts, customer accounts, carrier allocations, or rate agreements are altered automatically.
              </span>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
