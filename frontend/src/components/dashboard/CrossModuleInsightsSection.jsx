import React, { useState, useEffect, useCallback } from 'react';
import {
  Link2,
  AlertTriangle,
  ShieldCheck,
  RefreshCw,
  Clock,
  Sparkles,
  ArrowRight,
  HelpCircle,
  FileText,
  DollarSign,
  Truck,
  Building2,
  Layers,
  ChevronDown,
  ChevronUp,
  Info
} from 'lucide-react';
import insightsService from '../../services/insightsService';
import './CrossModuleInsightsSection.css';

export default function CrossModuleInsightsSection({
  entityType = '',
  entityId = 0,
  title = 'Cross-Module Connected Intelligence',
  subtitle = 'Connected business situations across operational freight, financial receivables, and commercial contracts.',
  compact = false,
}) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState(null);
  const [expandedInsights, setExpandedInsights] = useState({});

  const fetchInsights = useCallback(async (isManual = false) => {
    if (isManual) setRefreshing(true);
    else setLoading(true);
    setError(null);

    try {
      const response = await insightsService.getCrossModuleInsights(entityType, entityId);
      if (response && response.data) {
        setData(response.data);
      } else if (response && (response.insights !== undefined || response.total_insights_count !== undefined)) {
        setData(response);
      } else {
        setData(null);
      }
    } catch (err) {
      console.error('Failed to load cross-module insights', err);
      setError(err.message || 'Failed to load cross-module insights');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [entityType, entityId]);

  useEffect(() => {
    fetchInsights();
  }, [fetchInsights]);

  const toggleInsight = (id) => {
    setExpandedInsights((prev) => ({
      ...prev,
      [id]: !prev[id],
    }));
  };

  const getSeverityBadgeClass = (severity) => {
    switch ((severity || '').toUpperCase()) {
      case 'CRITICAL':
        return 'cmi-badge--critical';
      case 'HIGH':
        return 'cmi-badge--high';
      case 'MODERATE':
      case 'MEDIUM':
        return 'cmi-badge--moderate';
      case 'LOW':
      case 'INFO':
      default:
        return 'cmi-badge--low';
    }
  };

  const getModuleIcon = (module) => {
    switch ((module || '').toLowerCase()) {
      case 'finance':
        return <DollarSign size={13} className="cmi-mod-icon text-emerald" />;
      case 'shipments':
        return <Truck size={13} className="cmi-mod-icon text-blue" />;
      case 'contracts':
        return <FileText size={13} className="cmi-mod-icon text-purple" />;
      case 'customers':
        return <Building2 size={13} className="cmi-mod-icon text-indigo" />;
      case 'rfq':
      case 'quotations':
        return <Layers size={13} className="cmi-mod-icon text-amber" />;
      default:
        return <Link2 size={13} className="cmi-mod-icon text-muted" />;
    }
  };

  if (loading) {
    return (
      <div className="cmi-card cmi-loading-state" data-testid="cmi-loading">
        <div className="cmi-spinner"></div>
        <span className="cmi-loading-text">
          Evaluating cross-module signals & relationships...
        </span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="cmi-card cmi-error-state" data-testid="cmi-error">
        <AlertTriangle size={20} className="cmi-error-icon" />
        <div className="cmi-error-body">
          <h4 className="cmi-error-title">Insights Service Temporarily Unavailable</h4>
          <p className="cmi-error-msg">{error}</p>
          <button
            className="cmi-btn cmi-btn--secondary"
            onClick={() => fetchInsights(true)}
            data-testid="cmi-retry-btn"
          >
            <RefreshCw size={13} />
            <span>Retry Connection</span>
          </button>
        </div>
      </div>
    );
  }

  const insights = data?.insights || [];
  const aiSynthesis = data?.ai_synthesis || null;
  const criticalCount = data?.critical_count || 0;
  const highCount = data?.high_count || 0;
  const totalCount = data?.total_insights_count || insights.length;
  const correlationId = data?.correlation_id || '';
  const calculatedAt = data?.calculated_at || '';

  return (
    <div className={`cmi-container ${compact ? 'cmi-container--compact' : ''}`} data-testid="cross-module-insights-section">
      {/* ── HEADER ── */}
      <div className="cmi-header">
        <div className="cmi-header-left">
          <div className="cmi-meta-tag">
            <ShieldCheck size={13} className="cmi-shield-icon" />
            <span>READ-ONLY CROSS-MODULE INTELLIGENCE</span>
          </div>
          <div className="cmi-title-row">
            <h3 className="cmi-title">{title}</h3>
            {totalCount > 0 ? (
              <span className={`cmi-count-pill ${criticalCount > 0 ? 'cmi-count-pill--alert' : 'cmi-count-pill--active'}`}>
                {totalCount} {totalCount === 1 ? 'Signal' : 'Signals'} ({criticalCount} Critical, {highCount} High)
              </span>
            ) : (
              <span className="cmi-count-pill cmi-count-pill--clean">
                All Modules Synchronized
              </span>
            )}
          </div>
          <p className="cmi-subtitle">{subtitle}</p>
        </div>

        <div className="cmi-header-right">
          <div className="cmi-timing">
            <Clock size={12} />
            <span>{calculatedAt ? new Date(calculatedAt).toLocaleTimeString() : 'Live'}</span>
          </div>
          {correlationId && (
            <div className="cmi-corr" title={correlationId}>
              Corr: {correlationId.slice(0, 12)}...
            </div>
          )}
          <button
            className="cmi-btn cmi-btn--icon"
            onClick={() => fetchInsights(true)}
            disabled={refreshing}
            data-testid="cmi-refresh-btn"
            title="Re-run cross-module evaluation rules"
          >
            <RefreshCw size={14} className={refreshing ? 'cmi-spin' : ''} />
          </button>
        </div>
      </div>

      {/* ── GROUNDED AI SYNTHESIS ── */}
      {aiSynthesis && (
        <div className="cmi-ai-card" data-testid="cmi-ai-synthesis">
          <div className="cmi-ai-header">
            <div className="cmi-ai-header-left">
              <Sparkles size={14} className="cmi-ai-icon" />
              <span className="cmi-ai-title">Connected Cross-Module Synthesis</span>
            </div>
            <span className="cmi-confidence-badge">
              Confidence: {aiSynthesis.confidence || 'HIGH'}
            </span>
          </div>
          <div className="cmi-ai-body">
            <p className="cmi-ai-summary">{aiSynthesis.executive_summary}</p>
            {aiSynthesis.connected_situation_analysis && (
              <p className="cmi-ai-analysis">{aiSynthesis.connected_situation_analysis}</p>
            )}

            {aiSynthesis.suggested_investigation_questions?.length > 0 && (
              <div className="cmi-ai-questions">
                <div className="cmi-ai-q-title">Recommended Operator Investigation Areas:</div>
                <ul className="cmi-ai-q-list">
                  {aiSynthesis.suggested_investigation_questions.map((q, idx) => (
                    <li key={idx}>{q}</li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        </div>
      )}

      {/* ── INSIGHTS LIST ── */}
      {insights.length === 0 ? (
        <div className="cmi-card cmi-empty-state" data-testid="cmi-empty">
          <ShieldCheck size={28} className="cmi-empty-icon" />
          <h4 className="cmi-empty-title">Zero Cross-Module Risks Detected</h4>
          <p className="cmi-empty-desc">
            Operational timelines, invoices, quotations, and commercial contracts are operating without detected discrepancies or unhedged exposure.
          </p>
        </div>
      ) : (
        <div className="cmi-list" data-testid="cmi-list">
          {insights.map((item) => {
            const isExpanded = expandedInsights[item.insight_id] !== false; // default open
            return (
              <div
                key={item.insight_id}
                className={`cmi-item-card cmi-item-card--${(item.severity || 'low').toLowerCase()}`}
                data-testid={`cmi-item-${item.insight_id}`}
              >
                <div className="cmi-item-header" onClick={() => toggleInsight(item.insight_id)}>
                  <div className="cmi-item-header-left">
                    <span className={`cmi-badge ${getSeverityBadgeClass(item.severity)}`}>
                      {item.severity}
                    </span>
                    <span className="cmi-category-tag">{item.category}</span>
                    <h4 className="cmi-item-title">{item.title}</h4>
                  </div>

                  <div className="cmi-item-header-right">
                    <span className="cmi-score-pill">Score: {item.priorityScore || item.priority_score || 0}/100</span>
                    <button className="cmi-expand-btn">
                      {isExpanded ? <ChevronUp size={15} /> : <ChevronDown size={15} />}
                    </button>
                  </div>
                </div>

                {isExpanded && (
                  <div className="cmi-item-body">
                    <p className="cmi-item-desc">{item.explanation}</p>

                    {/* Connected Relationship Pills */}
                    {item.related_entities?.length > 0 && (
                      <div className="cmi-relationships">
                        <span className="cmi-subheading">Connected Records:</span>
                        <div className="cmi-rel-chips">
                          {item.related_entities.map((rel, idx) => (
                            <div key={idx} className="cmi-rel-chip">
                              {getModuleIcon(rel.module)}
                              <span><strong>{rel.entity_type}:</strong> {rel.reference || `#${rel.entity_id}`}</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Factual Evidence Grid */}
                    {item.evidence?.length > 0 && (
                      <div className="cmi-evidence-block">
                        <span className="cmi-subheading">Verified Evidence Fields:</span>
                        <div className="cmi-evidence-grid">
                          {item.evidence.map((ev, idx) => (
                            <div key={idx} className="cmi-evidence-item">
                              <span className="cmi-ev-label">{ev.field_name.replace(/_/g, ' ')}:</span>
                              <span className="cmi-ev-val font-mono">
                                {typeof ev.observed_value === 'number'
                                  ? ev.observed_value.toLocaleString()
                                  : String(ev.observed_value ?? '—')}
                              </span>
                              <span className="cmi-ev-desc text-muted">({ev.description})</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {/* Recommended Action */}
                    {item.suggested_human_action && (
                      <div className="cmi-action-callout">
                        <ArrowRight size={14} className="cmi-action-icon" />
                        <div>
                          <strong>Recommended Human Review:</strong> {item.suggested_human_action}
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* ── FOOTER DISCLAIMER ── */}
      <div className="cmi-footer-note">
        <Info size={12} />
        <span>
          Cross-module deterministic audit based on active MariaDB records. AI synthesis is strictly informational and does not trigger automated mutations.
        </span>
      </div>
    </div>
  );
}
