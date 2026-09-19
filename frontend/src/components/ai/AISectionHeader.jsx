import React from 'react';
import { ShieldCheck, Clock, RefreshCw } from 'lucide-react';
import './SharedAI.css';

export default function AISectionHeader({
  title,
  subtitle,
  badgeText,
  badgeType = 'neutral',
  correlationId,
  freshness,
  onRefresh,
  isRefreshing = false,
  readOnlyNotice = 'READ-ONLY INTELLIGENCE',
}) {
  return (
    <div className="sai-section-header">
      <div className="sai-header-left">
        {readOnlyNotice && (
          <div className="sai-meta-tag">
            <ShieldCheck size={12} className="text-emerald-600" />
            <span>{readOnlyNotice}</span>
          </div>
        )}
        <div className="sai-title-row">
          <h3 className="sai-title">{title}</h3>
          {badgeText && (
            <span className={`sai-badge sai-badge--${badgeType}`}>
              {badgeText}
            </span>
          )}
        </div>
        {subtitle && <p className="sai-subtitle">{subtitle}</p>}
      </div>

      <div className="sai-header-right">
        {freshness && (
          <div className="sai-meta-text">
            <Clock size={11} />
            <span>{new Date(freshness).toLocaleTimeString()}</span>
          </div>
        )}
        {correlationId && (
          <div className="sai-corr-pill" title={correlationId}>
            Corr: {correlationId.slice(0, 10)}...
          </div>
        )}
        {onRefresh && (
          <button
            className="sai-btn sai-btn--icon"
            onClick={onRefresh}
            disabled={isRefreshing}
            title="Refresh intelligence from persistent records"
            aria-label="Refresh intelligence"
          >
            <RefreshCw size={13} className={isRefreshing ? 'animate-spin text-teal-600' : ''} />
          </button>
        )}
      </div>
    </div>
  );
}
