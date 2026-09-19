import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
  Sparkles,
  AlertTriangle,
  CheckCircle2,
  ExternalLink,
  RefreshCw,
  Info,
  ShieldCheck,
  ChevronRight,
  Database,
} from 'lucide-react';
import api from '../../services/api';
import './BusinessIntelligenceCard.css';

/**
 * BusinessIntelligenceCard
 *
 * Compact, light-themed, grounded intelligence component for LogisticsHQ.
 * Strictly read-only, informational, and cites verified source records with deep links.
 */
export default function BusinessIntelligenceCard({
  entityType,
  entityId,
  title: customTitle,
  className = '',
  compact = false,
}) {
  const [insight, setInsight] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchInsight = useCallback(async () => {
    if (!entityType || !entityId) return;
    setLoading(true);
    setError(null);
    try {
      const res = await api.post('/api/v1/intelligence/insight', {
        entity_type: entityType,
        entity_id: Number(entityId),
      });
      const data = res?.data !== undefined ? res.data : res;
      setInsight(data && typeof data === 'object' && data.summary ? data : null);
    } catch (err) {
      setError(err?.message || 'Unable to retrieve intelligence insights for this record.');
      setInsight(null);
    } finally {
      setLoading(false);
    }
  }, [entityType, entityId]);

  useEffect(() => {
    fetchInsight();
  }, [fetchInsight]);

  if (loading) {
    return (
      <div className={`bic-card bic-card--loading ${className}`}>
        <div className="bic-header">
          <div className="bic-title-group">
            <div className="bic-icon-wrapper">
              <Sparkles size={16} className="bic-sparkle-icon" />
            </div>
            <div>
              <h4 className="bic-title">{customTitle || 'Business Context & Intelligence'}</h4>
              <p className="bic-subtitle">Reading related LogisticsHQ records…</p>
            </div>
          </div>
          <span className="bic-badge bic-badge--read-only">Read-Only</span>
        </div>
        <div className="bic-loading-body">
          <div className="bic-spinner" />
          <span>Synthesizing cross-module records…</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`bic-card bic-card--error ${className}`}>
        <div className="bic-header">
          <div className="bic-title-group">
            <div className="bic-icon-wrapper bic-icon-wrapper--error">
              <AlertTriangle size={16} />
            </div>
            <div>
              <h4 className="bic-title">{customTitle || 'Business Context Intelligence'}</h4>
              <p className="bic-subtitle text-danger">Service unavailable</p>
            </div>
          </div>
          <button
            type="button"
            className="bic-retry-btn"
            onClick={fetchInsight}
            title="Retry retrieval"
          >
            <RefreshCw size={13} />
            <span>Retry</span>
          </button>
        </div>
        <p className="bic-error-message">{error}</p>
      </div>
    );
  }

  if (!insight) {
    return (
      <div className={`bic-card bic-card--empty ${className}`}>
        <div className="bic-header">
          <div className="bic-title-group">
            <div className="bic-icon-wrapper">
              <Info size={16} />
            </div>
            <h4 className="bic-title">{customTitle || 'Business Context & Intelligence'}</h4>
          </div>
          <span className="bic-badge bic-badge--read-only">Informational</span>
        </div>
        <p className="bic-empty-text">No cross-module context available for this record.</p>
      </div>
    );
  }

  const confidenceClass =
    insight.confidence_level === 'HIGH'
      ? 'bic-confidence--high'
      : insight.confidence_level === 'MEDIUM'
      ? 'bic-confidence--medium'
      : 'bic-confidence--low';

  return (
    <div className={`bic-card ${className}`}>
      {/* Card Header */}
      <div className="bic-header">
        <div className="bic-title-group">
          <div className="bic-icon-wrapper">
            <Sparkles size={16} className="bic-sparkle-icon" />
          </div>
          <div>
            <div className="bic-header-line">
              <h4 className="bic-title">{insight.title || customTitle || 'Business Context & Intelligence'}</h4>
              <span className={`bic-confidence-pill ${confidenceClass}`}>
                {insight.confidence_level || 'HIGH'} Confidence
              </span>
            </div>
            <p className="bic-subtitle">Cross-module verification • AI-Assisted (Read-Only)</p>
          </div>
        </div>

        <div className="bic-actions">
          <button
            type="button"
            className="bic-icon-btn"
            onClick={fetchInsight}
            title="Refresh business insights"
          >
            <RefreshCw size={13} />
          </button>
        </div>
      </div>

      {/* Summary Section */}
      <div className="bic-body">
        <p className="bic-summary-text">{insight.summary}</p>

        {/* Operational Warnings Banner */}
        {insight.warnings && insight.warnings.length > 0 && (
          <div className="bic-warnings-box">
            <div className="bic-warnings-header">
              <AlertTriangle size={14} className="bic-warning-icon" />
              <span>Attention Items & Discrepancies ({insight.warnings.length})</span>
            </div>
            <ul className="bic-warnings-list">
              {insight.warnings.map((w, i) => (
                <li key={i}>{w}</li>
              ))}
            </ul>
          </div>
        )}

        {/* Key Highlights */}
        {insight.key_highlights && insight.key_highlights.length > 0 && (
          <div className="bic-highlights-box">
            {insight.key_highlights.map((h, i) => (
              <div key={i} className="bic-highlight-item">
                <CheckCircle2 size={13} className="bic-highlight-icon" />
                <span>{h}</span>
              </div>
            ))}
          </div>
        )}

        {/* Supporting Records Traceability (Deep Links) */}
        {insight.supporting_records && insight.supporting_records.length > 0 && (
          <div className="bic-sources-section">
            <div className="bic-sources-title">
              <Database size={13} />
              <span>Grounded Source Records ({insight.supporting_records.length})</span>
            </div>
            <div className="bic-sources-grid">
              {insight.supporting_records.slice(0, 6).map((src, i) => {
                const targetPath = src.path || '#';
                return (
                  <Link
                    key={i}
                    to={targetPath}
                    className="bic-source-chip"
                    title={`Open ${src.entity_type} ${src.reference_number}`}
                  >
                    <span className="bic-source-type">{src.entity_type}</span>
                    <span className="bic-source-ref">{src.reference_number || `#${src.entity_id}`}</span>
                    {src.status && (
                      <span className="bic-source-status">{src.status}</span>
                    )}
                    <ExternalLink size={11} className="bic-source-ext" />
                  </Link>
                );
              })}
            </div>
          </div>
        )}

        {/* Supporting Field Citations */}
        {!compact && insight.supporting_field_references && insight.supporting_field_references.length > 0 && (
          <div className="bic-fields-section">
            <div className="bic-fields-title">Verified Field Values:</div>
            <div className="bic-fields-chips">
              {insight.supporting_field_references.map((f, i) => (
                <div key={i} className="bic-field-chip">
                  <span className="bic-field-name">{f.field_name}:</span>
                  <span className="bic-field-val">{f.field_value}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Recommended Follow-Up */}
        {insight.recommended_follow_up && insight.recommended_follow_up.length > 0 && (
          <div className="bic-followup-box">
            <div className="bic-followup-title">Recommended Follow-Up (Informational):</div>
            <ul className="bic-followup-list">
              {insight.recommended_follow_up.map((item, i) => (
                <li key={i}>{item}</li>
              ))}
            </ul>
          </div>
        )}
      </div>

      {/* Card Footer */}
      <div className="bic-footer">
        <div className="bic-footer-meta">
          <ShieldCheck size={12} className="bic-shield-icon" />
          <span>All insights are organization-isolated, read-only, and grounded in real records.</span>
        </div>
        <div className="bic-footer-freshness">
          {insight.data_freshness ? `Updated ${new Date(insight.data_freshness).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}` : 'Live'}
        </div>
      </div>
    </div>
  );
}
