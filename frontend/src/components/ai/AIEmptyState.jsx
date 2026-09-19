import React from 'react';
import { ShieldCheck } from 'lucide-react';
import './SharedAI.css';

export default function AIEmptyState({
  title = 'No Active Intelligence Anomalies',
  description = 'All connected operational, commercial, and financial records for this context are in healthy state.',
  icon = null,
  actionText = null,
  onAction = null,
}) {
  return (
    <div className="sai-empty-state" data-testid="ai-empty-state">
      <div className="sai-empty-icon">
        {icon || <ShieldCheck size={20} />}
      </div>
      <h4 className="sai-empty-title">{title}</h4>
      <p className="sai-empty-desc">{description}</p>
      {actionText && onAction && (
        <button className="sai-btn" onClick={onAction} style={{ marginTop: 4 }}>
          {actionText}
        </button>
      )}
    </div>
  );
}
