/**
 * Canonical AI Workforce Status Contract (Frontend Mirror)
 * Defined in accordance with Phase 0 — Task 0.9.
 * Perfectly mirrors backend/internal/aitasks/workforce_status.go
 */

export const WORKFORCE_STATUS = {
  QUEUED: 'queued',
  CLAIMED: 'claimed',
  PROCESSING: 'processing',
  WAITING_FOR_APPROVAL: 'waiting_for_approval',
  PAUSED: 'paused',
  RETRYING: 'retrying',
  COMPLETED: 'completed',
  COMPLETED_WITH_FAILOVER: 'completed_with_failover',
  COMPLETED_IN_MOCK_MODE: 'completed_in_mock_mode',
  FAILED: 'failed',
  CANCELLED: 'cancelled',
  STALE: 'stale',
  UNKNOWN: 'unknown',
};

export const WORKFORCE_STATUS_METADATA = {
  [WORKFORCE_STATUS.QUEUED]: {
    value: WORKFORCE_STATUS.QUEUED,
    label: 'Queued',
    description: 'Task is waiting in the operational queue for a worker to claim it.',
    severity: 'neutral',
    color: '#475569',
    bg: '#f1f5f9',
    border: '#cbd5e1',
    isTerminal: false,
    isActionable: true,
    retryPermitted: false,
    approvalRequired: false,
    pulse: true,
  },
  [WORKFORCE_STATUS.CLAIMED]: {
    value: WORKFORCE_STATUS.CLAIMED,
    label: 'Claimed',
    description: 'Task claimed by a worker process and preparing execution.',
    severity: 'info',
    color: '#0284c7',
    bg: '#e0f2fe',
    border: '#bae6fd',
    isTerminal: false,
    isActionable: true,
    retryPermitted: false,
    approvalRequired: false,
    pulse: true,
  },
  [WORKFORCE_STATUS.PROCESSING]: {
    value: WORKFORCE_STATUS.PROCESSING,
    label: 'Processing',
    description: 'Agent workflow is actively executing nodes and tools.',
    severity: 'info',
    color: '#2563eb',
    bg: '#eff6ff',
    border: '#bfdbfe',
    isTerminal: false,
    isActionable: true,
    retryPermitted: false,
    approvalRequired: false,
    pulse: true,
  },
  [WORKFORCE_STATUS.WAITING_FOR_APPROVAL]: {
    value: WORKFORCE_STATUS.WAITING_FOR_APPROVAL,
    label: 'Awaiting Sign-Off',
    description: 'Workflow paused at a HITL interrupt awaiting human sign-off.',
    severity: 'warning',
    color: '#b45309',
    bg: '#fef3c7',
    border: '#fde68a',
    isTerminal: false,
    isActionable: true,
    retryPermitted: false,
    approvalRequired: true,
    pulse: false,
  },
  [WORKFORCE_STATUS.PAUSED]: {
    value: WORKFORCE_STATUS.PAUSED,
    label: 'Paused',
    description: 'Workflow execution is paused pending external dependencies.',
    severity: 'neutral',
    color: '#475569',
    bg: '#f1f5f9',
    border: '#cbd5e1',
    isTerminal: false,
    isActionable: true,
    retryPermitted: false,
    approvalRequired: false,
    pulse: false,
  },
  [WORKFORCE_STATUS.RETRYING]: {
    value: WORKFORCE_STATUS.RETRYING,
    label: 'Retrying',
    description: 'Encountered transient fault and is scheduled for backoff retry.',
    severity: 'warning',
    color: '#c2410c',
    bg: '#fff7ed',
    border: '#fed7aa',
    isTerminal: false,
    isActionable: true,
    retryPermitted: false,
    approvalRequired: false,
    pulse: true,
  },
  [WORKFORCE_STATUS.COMPLETED]: {
    value: WORKFORCE_STATUS.COMPLETED,
    label: 'Completed',
    description: 'Task executed to completion with business output committed.',
    severity: 'success',
    color: '#047857',
    bg: '#ecfdf5',
    border: '#a7f3d0',
    isTerminal: true,
    isActionable: false,
    retryPermitted: false,
    approvalRequired: false,
    pulse: false,
  },
  [WORKFORCE_STATUS.COMPLETED_WITH_FAILOVER]: {
    value: WORKFORCE_STATUS.COMPLETED_WITH_FAILOVER,
    label: 'Completed (Failover)',
    description: 'Task completed successfully using the secondary failover provider.',
    severity: 'success',
    color: '#0f766e',
    bg: '#f0fdfa',
    border: '#99f6e4',
    isTerminal: true,
    isActionable: false,
    retryPermitted: false,
    approvalRequired: false,
    pulse: false,
  },
  [WORKFORCE_STATUS.COMPLETED_IN_MOCK_MODE]: {
    value: WORKFORCE_STATUS.COMPLETED_IN_MOCK_MODE,
    label: 'Completed (Mock)',
    description: 'Task simulated via deterministic mock engine in development mode.',
    severity: 'neutral',
    color: '#6d28d9',
    bg: '#f5f3ff',
    border: '#ddd6fe',
    isTerminal: true,
    isActionable: false,
    retryPermitted: false,
    approvalRequired: false,
    pulse: false,
  },
  [WORKFORCE_STATUS.FAILED]: {
    value: WORKFORCE_STATUS.FAILED,
    label: 'Failed',
    description: 'Task failed and exhausted retries, or suffered fatal error.',
    severity: 'danger',
    color: '#b91c1c',
    bg: '#fef2f2',
    border: '#fecaca',
    isTerminal: true,
    isActionable: true,
    retryPermitted: true,
    approvalRequired: false,
    pulse: false,
  },
  [WORKFORCE_STATUS.CANCELLED]: {
    value: WORKFORCE_STATUS.CANCELLED,
    label: 'Cancelled',
    description: 'Task was cancelled by an operator before final execution.',
    severity: 'neutral',
    color: '#64748b',
    bg: '#f8fafc',
    border: '#e2e8f0',
    isTerminal: true,
    isActionable: false,
    retryPermitted: false,
    approvalRequired: false,
    pulse: false,
  },
  [WORKFORCE_STATUS.STALE]: {
    value: WORKFORCE_STATUS.STALE,
    label: 'Stale / Interrupted',
    description: 'Worker heartbeat expired or task lease exceeded without progress.',
    severity: 'danger',
    color: '#be123c',
    bg: '#fff1f2',
    border: '#fecdd3',
    isTerminal: false,
    isActionable: true,
    retryPermitted: true,
    approvalRequired: false,
    pulse: true,
  },
  [WORKFORCE_STATUS.UNKNOWN]: {
    value: WORKFORCE_STATUS.UNKNOWN,
    label: 'Unknown',
    description: 'Source execution data is genuinely unavailable or unmapped.',
    severity: 'neutral',
    color: '#64748b',
    bg: '#f8fafc',
    border: '#e2e8f0',
    isTerminal: false,
    isActionable: false,
    retryPermitted: false,
    approvalRequired: false,
    pulse: false,
  },
};

/**
 * Normalizes any backend, database, or graph status string into the canonical contract.
 */
export function normalizeWorkforceStatus(rawStatus, task = {}) {
  if (!rawStatus) return WORKFORCE_STATUS.UNKNOWN;

  const s = String(rawStatus).toLowerCase().trim();

  // Handle direct matches
  if (WORKFORCE_STATUS_METADATA[s]) {
    return s;
  }

  // Handle mock mode
  if (task.mock_mode || task.is_mock) {
    if (s === 'completed' || s === 'succeeded' || s === 'success') {
      return WORKFORCE_STATUS.COMPLETED_IN_MOCK_MODE;
    }
  }

  // Handle failover
  if (task.provider_failover || task.used_failover) {
    if (s === 'completed' || s === 'succeeded' || s === 'success') {
      return WORKFORCE_STATUS.COMPLETED_WITH_FAILOVER;
    }
  }

  // Map legacy / graph states
  switch (s) {
    case 'queued':
    case 'pending':
    case 'submitted':
      return WORKFORCE_STATUS.QUEUED;
    case 'claimed':
      return WORKFORCE_STATUS.CLAIMED;
    case 'processing':
    case 'running':
    case 'in_progress':
    case 'collecting_information':
    case 'analyzing_data':
    case 'waiting_for_llm':
    case 'generating_draft':
      return WORKFORCE_STATUS.PROCESSING;
    case 'waiting_for_approval':
    case 'waiting_for_human_review':
    case 'approval_required':
      return WORKFORCE_STATUS.WAITING_FOR_APPROVAL;
    case 'paused':
      return WORKFORCE_STATUS.PAUSED;
    case 'retrying':
      return WORKFORCE_STATUS.RETRYING;
    case 'completed':
    case 'succeeded':
    case 'success':
      return WORKFORCE_STATUS.COMPLETED;
    case 'failed':
    case 'error':
    case 'rejected':
      return WORKFORCE_STATUS.FAILED;
    case 'cancelled':
    case 'aborted':
      return WORKFORCE_STATUS.CANCELLED;
    case 'stale':
    case 'interrupted':
    case 'timed_out':
      return WORKFORCE_STATUS.STALE;
    default:
      return WORKFORCE_STATUS.UNKNOWN;
  }
}

/**
 * Helper to fetch metadata with fallback to UNKNOWN.
 */
export function getWorkforceStatusMetadata(rawStatus, task = {}) {
  const norm = normalizeWorkforceStatus(rawStatus, task);
  return WORKFORCE_STATUS_METADATA[norm] || WORKFORCE_STATUS_METADATA[WORKFORCE_STATUS.UNKNOWN];
}
