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
  ShieldAlert
} from 'lucide-react';
import predictionService from '../../services/predictionService';
import { SeverityBadge, ConfidenceBadge } from './PredictionBadge';
import './ContractCompliancePredictiveIntelligenceCard.css';

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
 * Resolve display label and color class for contract compliance prediction category
 */
function getComplianceCategoryInfo(category = '', predictionType = '') {
  const cat = String(category || predictionType).toUpperCase();
  if (cat.includes('CLAUSE') || cat.includes('COMMERCIAL_RISK')) {
    return { label: 'Clause & Commercial Risk', colorClass: 'cc-cat-badge--clause' };
  }
  if (cat.includes('EXPIRY') || cat.includes('RENEWAL')) {
    return { label: 'Expiry & Renewal Window', colorClass: 'cc-cat-badge--expiry' };
  }
  if (cat.includes('DOCUMENTATION') || cat.includes('COMPLETENESS')) {
    return { label: 'Documentation Completeness', colorClass: 'cc-cat-badge--doc' };
  }
  if (cat.includes('COMPLIANCE_REVIEW') || cat.includes('COMPLIANCE')) {
    return { label: 'Compliance Review Task', colorClass: 'cc-cat-badge--compliance' };
  }
  if (cat.includes('CROSS_MODULE')) {
    return { label: 'Cross-Module Risk', colorClass: 'cc-cat-badge--cross' };
  }
  if (cat.includes('INSUFFICIENT')) {
    return { label: 'Insufficient Data', colorClass: 'cc-cat-badge--neutral' };
  }
  return { label: 'Contract Compliance Intelligence', colorClass: 'cc-cat-badge--neutral' };
}

export default function ContractCompliancePredictiveIntelligenceCard({
  contractId,
  contract = null,
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
    if (!contractId) return;
    if (forceRefresh) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }
    setError(null);

    try {
      let data;
      if (forceRefresh) {
        data = await predictionService.refreshContractPredictedComplianceRisk(contractId);
      } else {
        data = await predictionService.getContractPredictedComplianceRisk(contractId, false);
      }
      setPrediction(data?.data || data?.prediction || data);
    } catch (err) {
      console.error('Failed to load contract compliance intelligence:', err);
      setError(err?.response?.data?.error || err.message || 'Failed to load compliance prediction');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [contractId]);

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
      setActionSuccessMsg('Prediction acknowledged by compliance review team.');
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
      const notes = `Operator authorized compliance review: ${prediction.recommended_action || 'Review contract terms and clauses'}`;
      await predictionService.requestAction(prediction.prediction_id, {
        action_type: prediction.action_type || 'contracts.request_compliance_review',
        notes
      });
      setPrediction(prev => ({
        ...prev,
        status: 'ACTION_REQUESTED',
        review_status: 'ACTION_REQUESTED'
      }));
      setActionSuccessMsg('Compliance review task queued in Action System.');
      setTimeout(() => setActionSuccessMsg(null), 4000);
    } catch (err) {
      console.error('Failed to request compliance action:', err);
      setError('Failed to queue action in Action System.');
    } finally {
      setActing(false);
    }
  };

  if (loading && !prediction) {
    return (
      <div className={`contract-compliance-card contract-compliance-card--loading ${className}`}>
        <div className="cc-loading-row">
          <RefreshCw size={18} className="cc-spin text-slate" />
          <span className="cc-loading-text">Synthesizing predictive contract & compliance intelligence...</span>
        </div>
      </div>
    );
  }

  if (error && !prediction) {
    return (
      <div className={`contract-compliance-card contract-compliance-card--error ${className}`}>
        <div className="cc-error-header">
          <AlertCircle size={18} className="text-rose" />
          <span className="cc-error-title">Compliance Intelligence Offline</span>
        </div>
        <p className="cc-error-desc">{error}</p>
        <button
          type="button"
          className="cc-btn cc-btn--secondary cc-btn--sm"
          onClick={() => fetchIntelligence(true)}
        >
          <RefreshCw size={13} /> Retry Analysis
        </button>
      </div>
    );
  }

  if (!prediction) return null;

  // Category & UI Badging
  const catInfo = getComplianceCategoryInfo(prediction.prediction_category, prediction.prediction_type);
  const isInsufficient = prediction.insufficient_data || prediction.review_status === 'INSUFFICIENT_DATA';
  const signals = prediction.supporting_signals || [];
  const sources = prediction.source_references || [];

  // Helper signal extractors
  const getSignal = (name) => {
    const s = signals.find(item => item.signal_name === name || item.signal_name?.includes(name));
    return s ? s.observed_value : null;
  };

  // Authoritative metrics
  const authoritativeStatus = getSignal('contract_status') || contract?.status || 'ACTIVE';
  const daysUntilExpiry = getSignal('days_until_expiry') || getSignal('days_past_expiry');
  const discrepancyCount = getSignal('structured_discrepancies') || getSignal('discrepancy_count') || '0';
  const missingDocsCount = getSignal('missing_document_count') || (prediction.prediction_type === 'DOCUMENTATION_COMPLETENESS_RISK' ? '1+' : '0');

  // Grounding clause and document citations
  const clauseRef = prediction.clause_reference || getSignal('clause_reference') || (sources.find(s => s.clause_reference)?.clause_reference);
  const docRef = prediction.document_reference || getSignal('document_file') || (sources.find(s => s.document_reference)?.document_reference);
  const pageNo = prediction.page_number || (sources.find(s => s.page_number)?.page_number);
  const sectionHeading = prediction.section_heading || (sources.find(s => s.section_heading)?.section_heading);

  return (
    <div className={`contract-compliance-card ${className}`}>
      {/* 1. Header Bar */}
      <div className="cc-header">
        <div className="cc-header__left">
          <div className="cc-icon-pill">
            <FileCheck size={18} className="text-navy" />
          </div>
          <div>
            <div className="cc-title-row">
              <h4 className="cc-title">Predictive Contract & Compliance Intelligence</h4>
              <span className={`cc-category-badge ${catInfo.colorClass}`}>
                {catInfo.label}
              </span>
            </div>
            <p className="cc-subtitle">
              Predictive risk forecasting grounded in authoritative contract terms, document OCR, and compliance reviews.
            </p>
          </div>
        </div>
        <div className="cc-header__right">
          <button
            type="button"
            className="cc-refresh-btn"
            onClick={() => fetchIntelligence(true)}
            disabled={refreshing}
            title="Refresh compliance forecast"
          >
            <RefreshCw size={14} className={refreshing ? 'cc-spin' : ''} />
            <span>{refreshing ? 'Refreshing...' : 'Refresh'}</span>
          </button>
        </div>
      </div>

      {/* Action confirmation toast */}
      {actionSuccessMsg && (
        <div className="cc-alert-banner cc-alert-banner--success">
          <CheckCircle2 size={15} />
          <span>{actionSuccessMsg}</span>
        </div>
      )}

      {/* Insufficient Data State */}
      {isInsufficient ? (
        <div className="cc-insufficient-state">
          <Info size={20} className="text-slate" />
          <div>
            <h5 className="cc-insufficient-title">Insufficient Historical or Document Records</h5>
            <p className="cc-insufficient-desc">
              {prediction.insufficient_data_reason || prediction.explanation || 'No active documents, structured rate schedules, or compliance review logs found for this contract.'}
            </p>
          </div>
        </div>
      ) : (
        <>
          {/* 2. Authoritative 4-Metric Grid */}
          <div className="cc-metric-grid">
            <div className="cc-metric-tile">
              <span className="cc-metric-tile__label">AUTHORITATIVE STATUS</span>
              <div className="cc-metric-tile__val-row">
                <span className={`cc-status-pill cc-status-pill--${authoritativeStatus.toLowerCase()}`}>
                  {authoritativeStatus}
                </span>
              </div>
              <span className="cc-metric-tile__sub">Deterministic record</span>
            </div>

            <div className="cc-metric-tile">
              <span className="cc-metric-tile__label">VALIDITY TIMELINE</span>
              <div className="cc-metric-tile__val-row">
                <Clock size={16} className="text-slate" />
                <span className="cc-metric-tile__val">
                  {authoritativeStatus === 'EXPIRED'
                    ? `Expired (${daysUntilExpiry || 'Past'}d ago)`
                    : daysUntilExpiry
                      ? `${daysUntilExpiry} Days Left`
                      : 'Valid Schedule'}
                </span>
              </div>
              <span className="cc-metric-tile__sub">Term horizon</span>
            </div>

            <div className="cc-metric-tile">
              <span className="cc-metric-tile__label">COMPLIANCE POSTURE</span>
              <div className="cc-metric-tile__val-row">
                {String(discrepancyCount) !== '0' || authoritativeStatus === 'EXPIRED' ? (
                  <ShieldAlert size={16} className="text-amber" />
                ) : (
                  <ShieldCheck size={16} className="text-emerald" />
                )}
                <span className="cc-metric-tile__val">
                  {String(discrepancyCount) !== '0' ? `${discrepancyCount} Discrepancy Flag` : 'Standard Posture'}
                </span>
              </div>
              <span className="cc-metric-tile__sub">
                {String(missingDocsCount) !== '0' ? `${missingDocsCount} Missing Document(s)` : 'Documents complete'}
              </span>
            </div>

            <div className="cc-metric-tile">
              <span className="cc-metric-tile__label">DOCUMENT & CLAUSE CITATION</span>
              <div className="cc-metric-tile__val-row">
                <FileSearch size={16} className="text-indigo" />
                <span className="cc-metric-tile__val text-truncate" title={clauseRef || docRef || 'None flagged'}>
                  {clauseRef ? `${clauseRef}${pageNo ? ` (p.${pageNo})` : ''}` : docRef ? docRef : 'Standard Terms'}
                </span>
              </div>
              <span className="cc-metric-tile__sub text-truncate" title={sectionHeading || 'No flagged clauses'}>
                {sectionHeading ? sectionHeading : 'No discrepancies detected'}
              </span>
            </div>
          </div>

          {/* 3. Prediction Statement & Impact Callout */}
          <div className="cc-prediction-callout">
            <div className="cc-callout-top">
              <div className="cc-callout-badges">
                <SeverityBadge severity={prediction.severity} />
                <ConfidenceBadge
                  score={prediction.confidence_score}
                  band={prediction.confidence_band}
                />
              </div>
              <span className="cc-advisory-tag">
                <Info size={13} /> Advisory Forecast
              </span>
            </div>

            <p className="cc-statement">
              {prediction.prediction_statement}
            </p>

            {prediction.explanation && (
              <div className="cc-explanation-box">
                <p className="cc-explanation-text">
                  <strong>Supporting Context: </strong>
                  {prediction.explanation}
                </p>
              </div>
            )}

            {/* Document Grounding Pill */}
            {(docRef || clauseRef || pageNo) && (
              <div className="cc-grounding-strip">
                <span className="cc-grounding-label">SOURCE GROUNDING:</span>
                {docRef && (
                  <span className="cc-grounding-badge">
                    <FileText size={12} /> {docRef}
                  </span>
                )}
                {clauseRef && (
                  <span className="cc-grounding-badge cc-grounding-badge--clause">
                    <FileSearch size={12} /> Clause: {clauseRef}
                  </span>
                )}
                {pageNo && (
                  <span className="cc-grounding-badge">
                    Page {pageNo}
                  </span>
                )}
                {sectionHeading && (
                  <span className="cc-grounding-badge cc-grounding-badge--section">
                    Section: {sectionHeading}
                  </span>
                )}
              </div>
            )}
          </div>

          {/* 4. Action Governance & Human Review Bar */}
          <div className="cc-action-bar">
            <div className="cc-action-left">
              <div className="cc-governance-note">
                <ShieldCheck size={14} className="text-emerald" />
                <span>
                  <strong>Advisory Only:</strong> Final contract activation, termination, or compliance determination requires authorized human review.
                </span>
              </div>
            </div>

            <div className="cc-action-right">
              {prediction.review_status !== 'ACKNOWLEDGED' && prediction.review_status !== 'ACTION_REQUESTED' && (
                <button
                  type="button"
                  className="cc-btn cc-btn--secondary"
                  onClick={handleAcknowledge}
                  disabled={acting}
                >
                  <Check size={14} /> Acknowledge
                </button>
              )}

              {prediction.is_action_required && prediction.review_status !== 'ACTION_REQUESTED' && (
                <button
                  type="button"
                  className="cc-btn cc-btn--primary"
                  onClick={handleRequestAction}
                  disabled={acting}
                >
                  <ArrowRight size={14} />
                  <span>{prediction.recommended_action ? 'Queue Recommended Action' : 'Request Compliance Review'}</span>
                </button>
              )}

              {prediction.review_status === 'ACTION_REQUESTED' && (
                <span className="cc-status-pill cc-status-pill--action-requested">
                  Action In Review
                </span>
              )}
            </div>
          </div>

          {/* 5. Collapsible Telemetry & Grounding Evidence */}
          <div className="cc-evidence-accordion">
            <button
              type="button"
              className="cc-evidence-toggle"
              onClick={() => setShowEvidence(prev => !prev)}
            >
              <span>Audit Grounding & Telemetry Signals ({signals.length + sources.length})</span>
              {showEvidence ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
            </button>

            {showEvidence && (
              <div className="cc-evidence-body">
                {signals.length > 0 && (
                  <div className="cc-evidence-section">
                    <h6 className="cc-evidence-title">Quantitative Signals</h6>
                    <div className="cc-signals-table">
                      {signals.map((sig, idx) => (
                        <div key={idx} className="cc-signal-row">
                          <span className="cc-signal-name">{sig.signal_name}</span>
                          <span className="cc-signal-val">{String(sig.observed_value)}</span>
                          <span className="cc-signal-baseline">
                            Baseline: {sig.baseline_value ? String(sig.baseline_value) : 'N/A'}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {sources.length > 0 && (
                  <div className="cc-evidence-section">
                    <h6 className="cc-evidence-title">Authoritative Source Audit Trails</h6>
                    <ul className="cc-sources-list">
                      {sources.map((src, idx) => (
                        <li key={idx} className="cc-source-item">
                          <span className="cc-source-field">{src.source_field || src.source_module}</span>
                          <span className="cc-source-id">Record #{src.source_record_id}</span>
                          {src.clause_reference && (
                            <span className="cc-source-clause">Clause: {src.clause_reference} (p.{src.page_number || 'N/A'})</span>
                          )}
                          <span className="cc-source-time">{formatDateTime(src.source_timestamp)}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                <div className="cc-metadata-footer">
                  <span>Prediction ID: <code>{prediction.prediction_id}</code></span>
                  <span>Model: <code>{prediction.model_version || 'ai-sidecar-contracts-v1'}</code></span>
                  <span>Forecast Horizon: <code>{prediction.time_horizon || '30_DAYS'}</code></span>
                </div>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}
