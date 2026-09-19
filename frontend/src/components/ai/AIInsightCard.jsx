import React from 'react';
import { ShieldCheck, AlertTriangle, AlertCircle, Info, Clock } from 'lucide-react';
import AIStatusBadge from './AIStatusBadge';
import AIConfidenceIndicator from './AIConfidenceIndicator';
import AIEvidenceList from './AIEvidenceList';
import './SharedAI.css';

export default function AIInsightCard({
  title,
  explanation,
  severity = 'INFO',
  confidence = 'HIGH',
  category,
  evidence = [],
  suggestedAction,
  ruleApplied,
  readOnly = true,
  freshness,
  correlationId,
  limitations = [],
  actionButton = null,
}) {
  const normSev = String(severity).toLowerCase();
  const alertClass = `sai-card--alert-${normSev}`;

  const renderIcon = () => {
    switch (normSev) {
      case 'critical': return <AlertCircle size={16} className="text-red-600" />;
      case 'high':
      case 'medium': return <AlertTriangle size={16} className="text-amber-600" />;
      case 'success': return <ShieldCheck size={16} className="text-emerald-600" />;
      default: return <Info size={16} className="text-blue-600" />;
    }
  };

  return (
    <div className={`sai-card ${alertClass}`}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 10, marginBottom: 8 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
          {renderIcon()}
          <h4 style={{ margin: 0, fontSize: '0.86rem', fontWeight: 750, color: '#0f172a' }}>{title}</h4>
          {category && (
            <span style={{ fontSize: '0.68rem', background: '#f1f5f9', color: '#475569', padding: '1px 6px', borderRadius: 4, fontWeight: 600 }}>
              {category}
            </span>
          )}
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexShrink: 0 }}>
          <AIStatusBadge status={severity} />
          <AIConfidenceIndicator confidence={confidence} />
        </div>
      </div>

      <p style={{ margin: '0 0 10px 0', fontSize: '0.78rem', color: '#334155', lineHeight: 1.5 }}>
        {explanation}
      </p>

      {/* Verifiable Evidence Table */}
      {evidence && evidence.length > 0 && (
        <AIEvidenceList evidence={evidence} defaultExpanded={false} />
      )}

      {/* Suggested Action Bar if available */}
      {suggestedAction && (
        <div style={{ marginTop: 10, background: '#f8fafc', border: '1px solid #e2e8f0', borderRadius: 6, padding: '7px 10px', display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 8 }}>
          <div style={{ fontSize: '0.74rem', color: '#475569' }}>
            <strong style={{ color: '#0f172a' }}>Recommended Operator Step:</strong> {suggestedAction}
          </div>
          {actionButton}
        </div>
      )}

      {/* Card Footer: Metadata & Safe Read-Only Indicators */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 10, paddingTop: 8, borderTop: '1px solid #f1f5f9', fontSize: '0.68rem', color: '#64748b' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          {readOnly && (
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 3, color: '#059669', fontWeight: 600 }}>
              <ShieldCheck size={11} /> READ-ONLY
            </span>
          )}
          {ruleApplied && <span>Rule: <code style={{ fontFamily: 'monospace' }}>{ruleApplied}</code></span>}
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          {freshness && (
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 3 }}>
              <Clock size={11} /> {new Date(freshness).toLocaleTimeString()}
            </span>
          )}
          {correlationId && (
            <span className="sai-corr-pill" title={correlationId}>
              {correlationId.slice(0, 16)}...
            </span>
          )}
        </div>
      </div>
    </div>
  );
}
