import React from 'react';
import { getWorkforceStatusMetadata, WORKFORCE_STATUS } from '../../utils/workforceStatus';

/**
 * Visualizes the current state of an AI Task or Agent Workflow.
 * Supports all 13 authoritative workforce statuses:
 * queued, claimed, processing, waiting_for_approval, paused, retrying,
 * completed, completed_with_failover, completed_in_mock_mode, failed,
 * cancelled, stale, unknown.
 */
export default function AgentStatusBadge({
  status,
  error = null,
  onRetry = null,
  isRetrying = false,
  onCancel = null,
  isCancelling = false,
  mockMode = false,
  providerFailover = false,
}) {
  if (!status || status === 'IDLE' || status === 'idle') return null;

  const meta = getWorkforceStatusMetadata(status, { mock_mode: mockMode, provider_failover: providerFailover });
  const label = meta.label;
  const pulse = meta.pulse;
  const badgeStyle = {
    background: meta.bg,
    color: meta.color,
    borderColor: meta.border,
  };

  let icon = null;
  switch (meta.value) {
    case WORKFORCE_STATUS.QUEUED:
    case WORKFORCE_STATUS.CLAIMED:
      icon = (
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="10" />
          <polyline points="12 6 12 12 16 14" />
        </svg>
      );
      break;

    case WORKFORCE_STATUS.PROCESSING:
      icon = (
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
        </svg>
      );
      break;

    case WORKFORCE_STATUS.WAITING_FOR_APPROVAL:
      icon = (
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      );
      break;

    case WORKFORCE_STATUS.RETRYING:
      icon = (
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M21.5 2v6h-6M21.34 15.57a10 10 0 1 1-.57-8.38l5.67-5.67" />
        </svg>
      );
      break;

    case WORKFORCE_STATUS.COMPLETED:
    case WORKFORCE_STATUS.COMPLETED_WITH_FAILOVER:
    case WORKFORCE_STATUS.COMPLETED_IN_MOCK_MODE:
      icon = (
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <polyline points="20 6 9 17 4 12" />
        </svg>
      );
      break;

    case WORKFORCE_STATUS.FAILED:
    case WORKFORCE_STATUS.STALE:
      icon = (
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="10" />
          <line x1="15" y1="9" x2="9" y2="15" />
          <line x1="9" y1="9" x2="15" y2="15" />
        </svg>
      );
      break;

    case WORKFORCE_STATUS.CANCELLED:
      icon = (
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <line x1="18" y1="6" x2="6" y2="18" />
          <line x1="6" y1="6" x2="18" y2="18" />
        </svg>
      );
      break;

    case WORKFORCE_STATUS.PAUSED:
    default:
      icon = (
        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <circle cx="12" cy="12" r="10" />
        </svg>
      );
      break;
  }

  return (
    <div style={{ display: 'inline-flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
      <div
        className="agent-status-badge"
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: '6px',
          padding: '4px 10px',
          borderRadius: '9999px',
          fontSize: '0.75rem',
          fontWeight: 600,
          letterSpacing: '0.025em',
          border: '1px solid',
          lineHeight: 1.2,
          ...badgeStyle,
        }}
        title={error ? `Error: ${error}` : undefined}
      >
        {pulse && (
          <span
            style={{
              display: 'inline-block',
              width: '6px',
              height: '6px',
              borderRadius: '50%',
              backgroundColor: 'currentColor',
              animation: 'pulse 1.5s cubic-bezier(0.4, 0, 0.6, 1) infinite',
            }}
          />
        )}
        {icon}
        <span>{label}</span>
      </div>

      {/* Clear error message if failed */}
      {error && (meta.value === WORKFORCE_STATUS.FAILED || String(status).toUpperCase() === 'FAILED') && (
        <span
          style={{
            fontSize: '0.72rem',
            color: '#f87171',
            maxWidth: '240px',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
          title={error}
        >
          {error}
        </span>
      )}

      {/* Safe retry action */}
      {onRetry && meta.retryPermitted && (
        <button
          type="button"
          onClick={onRetry}
          disabled={isRetrying}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '4px',
            padding: '3px 8px',
            fontSize: '0.72rem',
            fontWeight: 600,
            borderRadius: '6px',
            background: 'rgba(239, 68, 68, 0.15)',
            border: '1px solid rgba(239, 68, 68, 0.3)',
            color: '#fca5a5',
            cursor: isRetrying ? 'not-allowed' : 'pointer',
            opacity: isRetrying ? 0.6 : 1,
          }}
        >
          <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21.5 2v6h-6M21.34 15.57a10 10 0 1 1-.57-8.38l5.67-5.67" />
          </svg>
          {isRetrying ? 'Retrying...' : 'Retry'}
        </button>
      )}

      {/* Safe cancel action for non-terminal tasks */}
      {onCancel && !meta.isTerminal && meta.isActionable && (
        <button
          type="button"
          onClick={onCancel}
          disabled={isCancelling}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '4px',
            padding: '3px 8px',
            fontSize: '0.72rem',
            fontWeight: 600,
            borderRadius: '6px',
            background: 'rgba(100, 116, 139, 0.15)',
            border: '1px solid rgba(100, 116, 139, 0.3)',
            color: '#94a3b8',
            cursor: isCancelling ? 'not-allowed' : 'pointer',
            opacity: isCancelling ? 0.6 : 1,
          }}
        >
          {isCancelling ? 'Cancelling...' : 'Cancel'}
        </button>
      )}
    </div>
  );
}
