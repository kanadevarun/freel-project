import React from 'react';
import { AlertCircle, RotateCcw } from 'lucide-react';
import './SharedAI.css';

export default function AIErrorState({
  title = 'Intelligence Retrieval Fault',
  message = 'Unable to connect to intelligence runtime or database ledger.',
  onRetry = null,
  isRetrying = false,
}) {
  return (
    <div className="sai-error-state" data-testid="ai-error-state">
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <AlertCircle size={18} className="text-red-600 flex-shrink-0" />
        <div className="sai-error-text">
          <strong style={{ display: 'block', fontSize: '0.82rem' }}>{title}</strong>
          <span>{message}</span>
        </div>
      </div>
      {onRetry && (
        <button
          className="sai-btn sai-btn--retry flex-shrink-0"
          onClick={onRetry}
          disabled={isRetrying}
        >
          <RotateCcw size={12} className={isRetrying ? 'animate-spin' : ''} />
          <span>{isRetrying ? 'Retrying...' : 'Retry'}</span>
        </button>
      )}
    </div>
  );
}
