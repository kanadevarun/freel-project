import React from 'react';
import './SharedAI.css';

export default function AIStatusBadge({ status, pulse = false, label }) {
  if (!status) return null;

  const normalized = String(status).toLowerCase();
  let badgeClass = 'sai-badge--neutral';
  let displayLabel = label || status;

  if (['critical', 'failed', 'breached'].includes(normalized)) {
    badgeClass = 'sai-badge--critical';
  } else if (['high', 'retrying', 'warning', 'elevated', 'waiting_for_approval'].includes(normalized)) {
    badgeClass = 'sai-badge--warning';
  } else if (['medium', 'moderate', 'attention'].includes(normalized)) {
    badgeClass = 'sai-badge--medium';
  } else if (['low', 'info', 'queued', 'claimed', 'processing'].includes(normalized)) {
    badgeClass = 'sai-badge--info';
  } else if (['completed', 'success', 'compliant', 'active', 'healthy', 'verified'].includes(normalized)) {
    badgeClass = 'sai-badge--success';
  }

  return (
    <span className={`sai-badge ${badgeClass}`}>
      {pulse && <span className="w-1.5 h-1.5 rounded-full bg-current animate-pulse" />}
      {displayLabel}
    </span>
  );
}
